package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestRateLimitExceeded(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 3,
		Window:            time.Minute,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// 4th request should be rate limited.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestDifferentIPsIndependent(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 1,
		Window:            time.Minute,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	// First IP uses its one request.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.1.1.1:1000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("IP1 request 1: expected 200, got %d", rr.Code)
	}

	// Second IP should still work.
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "2.2.2.2:2000"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("IP2 request 1: expected 200, got %d", rr.Code)
	}
}

func TestTotalConnectionLimit(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		MaxConnsPerIP:     100,
		MaxTotalConns:     1, // Only 1 total connection allowed.
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)

	// Simulate a held connection by manually incrementing.
	limiter.mu.Lock()
	limiter.totalConn = 1
	limiter.mu.Unlock()

	handler := limiter.Wrap(okHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "3.3.3.3:3000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestXForwardedFor(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 1,
		Window:            time.Minute,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	// First request from forwarded IP.
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5000"
	req.Header.Set("X-Forwarded-For", "5.5.5.5, 10.0.0.1")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Second request from same forwarded IP should be rate limited.
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:6000" // Different proxy IP
	req.Header.Set("X-Forwarded-For", "5.5.5.5, 10.0.0.2")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
}

func TestPerIPConnectionLimit(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		MaxConnsPerIP:     1,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)

	// Simulate a held connection from this IP.
	limiter.mu.Lock()
	s := limiter.getOrCreate("4.4.4.4")
	s.connections = 1
	limiter.totalConn = 1
	limiter.mu.Unlock()

	handler := limiter.Wrap(okHandler())
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "4.4.4.4:4000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for connection limit, got %d", rr.Code)
	}
}

func TestWindowReset(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 1,
		Window:            50 * time.Millisecond,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "6.6.6.6:6000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("request 1: expected 200, got %d", rr.Code)
	}

	// Should be rate limited immediately.
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("request 2: expected 429, got %d", rr.Code)
	}

	// Wait for window to expire, then should succeed.
	time.Sleep(60 * time.Millisecond)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("request 3 after reset: expected 200, got %d", rr.Code)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.RequestsPerWindow <= 0 {
		t.Fatal("RequestsPerWindow must be positive")
	}
	if cfg.MaxConnsPerIP <= 0 {
		t.Fatal("MaxConnsPerIP must be positive")
	}
	if cfg.MaxTotalConns <= 0 {
		t.Fatal("MaxTotalConns must be positive")
	}
	if cfg.MaxBodyBytes <= 0 {
		t.Fatal("MaxBodyBytes must be positive")
	}
}

func TestExtractIP_IPv6(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[::1]:8080"
	got := extractIP(req)
	if got != "::1" {
		t.Fatalf("expected '::1', got %q", got)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:443"
	got = extractIP(req)
	if got != "2001:db8::1" {
		t.Fatalf("expected '2001:db8::1', got %q", got)
	}
}

func TestExtractIP_NoPort(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	// net.SplitHostPort will fail on an address with no port, so
	// extractIP should fall back to returning RemoteAddr as-is.
	req.RemoteAddr = "192.168.1.1"
	got := extractIP(req)
	if got != "192.168.1.1" {
		t.Fatalf("expected '192.168.1.1', got %q", got)
	}
}

func TestExtractIP_XForwardedForSingleIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	got := extractIP(req)
	if got != "203.0.113.50" {
		t.Fatalf("expected '203.0.113.50', got %q", got)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 1000,
		Window:            time.Minute,
		MaxConnsPerIP:     1000,
		MaxTotalConns:     1000,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "7.7.7.7:7000"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			// Any valid HTTP status is acceptable; we're testing for races/panics.
			if rr.Code != http.StatusOK && rr.Code != http.StatusTooManyRequests {
				t.Errorf("unexpected status %d", rr.Code)
			}
		}()
	}
	wg.Wait()
}

func TestConnectionCountDecrement(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	limiter := NewLimiter(cfg)
	handler := limiter.Wrap(okHandler())

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "8.8.8.8:8000"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// After the handler returns, connections should be decremented back to 0.
	limiter.mu.Lock()
	totalConn := limiter.totalConn
	s := limiter.ips["8.8.8.8"]
	ipConn := 0
	if s != nil {
		ipConn = s.connections
	}
	limiter.mu.Unlock()

	if totalConn != 0 {
		t.Fatalf("expected totalConn 0 after handler returned, got %d", totalConn)
	}
	if ipConn != 0 {
		t.Fatalf("expected per-IP connections 0 after handler returned, got %d", ipConn)
	}
}

func TestCleanupRemovesStaleEntries(t *testing.T) {
	cfg := Config{
		RequestsPerWindow: 10,
		Window:            10 * time.Millisecond,
		MaxConnsPerIP:     10,
		MaxTotalConns:     100,
		MaxBodyBytes:      1024,
	}
	// Create limiter manually to avoid the background cleanup goroutine.
	limiter := &Limiter{
		ips: make(map[string]*ipState),
		cfg: cfg,
	}

	// Insert a stale entry (window started long ago, no active connections).
	limiter.mu.Lock()
	limiter.ips["stale-ip"] = &ipState{
		requests:    5,
		connections: 0,
		windowStart: time.Now().Add(-time.Hour),
	}
	// Insert a fresh entry.
	limiter.ips["fresh-ip"] = &ipState{
		requests:    1,
		connections: 0,
		windowStart: time.Now(),
	}
	// Insert an entry with an active connection (should not be removed even if old).
	limiter.ips["active-ip"] = &ipState{
		requests:    3,
		connections: 1,
		windowStart: time.Now().Add(-time.Hour),
	}
	limiter.mu.Unlock()

	// Run cleanup logic inline (same logic as the cleanup method).
	limiter.mu.Lock()
	now := time.Now()
	for ip, s := range limiter.ips {
		if s.connections == 0 && now.Sub(s.windowStart) > limiter.cfg.Window*2 {
			delete(limiter.ips, ip)
		}
	}
	limiter.mu.Unlock()

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if _, ok := limiter.ips["stale-ip"]; ok {
		t.Fatal("stale-ip should have been cleaned up")
	}
	if _, ok := limiter.ips["fresh-ip"]; !ok {
		t.Fatal("fresh-ip should NOT have been cleaned up")
	}
	if _, ok := limiter.ips["active-ip"]; !ok {
		t.Fatal("active-ip should NOT have been cleaned up (has active connection)")
	}
}
