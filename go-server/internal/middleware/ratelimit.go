package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// Config holds rate limiting and connection security settings.
type Config struct {
	// Max requests per IP per window.
	RequestsPerWindow int
	// Time window for rate limiting.
	Window time.Duration
	// Max concurrent SSE/streaming connections per IP.
	MaxConnsPerIP int
	// Max total concurrent connections across all IPs.
	MaxTotalConns int
	// Max request body size in bytes.
	MaxBodyBytes int64
}

// DefaultConfig returns production-safe defaults.
func DefaultConfig() Config {
	return Config{
		RequestsPerWindow: 60,
		Window:            time.Minute,
		MaxConnsPerIP:     5,
		MaxTotalConns:     100,
		MaxBodyBytes:      64 * 1024, // 64KB
	}
}

type ipState struct {
	requests    int
	connections int
	windowStart time.Time
}

// Limiter is an in-memory per-IP rate limiter and connection tracker.
type Limiter struct {
	mu        sync.Mutex
	ips       map[string]*ipState
	totalConn int
	cfg       Config
}

// NewLimiter creates a new rate limiter with the given config.
func NewLimiter(cfg Config) *Limiter {
	l := &Limiter{
		ips: make(map[string]*ipState),
		cfg: cfg,
	}
	// Periodically clean up stale entries.
	go l.cleanup()
	return l
}

func (l *Limiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for ip, s := range l.ips {
			if s.connections == 0 && now.Sub(s.windowStart) > l.cfg.Window*2 {
				delete(l.ips, ip)
			}
		}
		l.mu.Unlock()
	}
}

func extractIP(r *http.Request) string {
	// Trust X-Forwarded-For from Railway's reverse proxy.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// First IP in the chain is the client.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (l *Limiter) getOrCreate(ip string) *ipState {
	s, ok := l.ips[ip]
	if !ok {
		s = &ipState{windowStart: time.Now()}
		l.ips[ip] = s
	}
	return s
}

// Wrap wraps an http.Handler with rate limiting, connection limits, and body size limits.
func (l *Limiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		now := time.Now()

		l.mu.Lock()

		// Check total connection limit.
		if l.totalConn >= l.cfg.MaxTotalConns {
			l.mu.Unlock()
			http.Error(w, "server busy", http.StatusServiceUnavailable)
			return
		}

		s := l.getOrCreate(ip)

		// Reset window if expired.
		if now.Sub(s.windowStart) > l.cfg.Window {
			s.requests = 0
			s.windowStart = now
		}

		// Check rate limit.
		if s.requests >= l.cfg.RequestsPerWindow {
			retryAfter := l.cfg.Window - now.Sub(s.windowStart)
			l.mu.Unlock()
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Check per-IP connection limit.
		if s.connections >= l.cfg.MaxConnsPerIP {
			l.mu.Unlock()
			http.Error(w, "too many connections", http.StatusTooManyRequests)
			return
		}

		s.requests++
		s.connections++
		l.totalConn++
		l.mu.Unlock()

		// Track connection close.
		defer func() {
			l.mu.Lock()
			s.connections--
			l.totalConn--
			l.mu.Unlock()
		}()

		// Limit request body size.
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, l.cfg.MaxBodyBytes)
		}

		next.ServeHTTP(w, r)
	})
}
