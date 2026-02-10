package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go-server/internal/models"
)

func TestNewServerReturnsDistinctInstances(t *testing.T) {
	s1 := newServer()
	s2 := newServer()
	s3 := newServer()

	if s1 == s2 {
		t.Fatal("s1 and s2 are the same instance")
	}
	if s1 == s3 {
		t.Fatal("s1 and s3 are the same instance")
	}
	if s2 == s3 {
		t.Fatal("s2 and s3 are the same instance")
	}
}

// TestConcurrentSSESessions verifies that multiple concurrent SSE connections
// each get their own server instance and can complete the full MCP lifecycle
// (initialize → notifications/initialized → tools/call) without interfering
// with each other.
//
// Regression test: a shared *mcp.Server caused Docker health-check connections
// to corrupt state for real clients ("method tools/call is invalid during
// session initialization").
func TestConcurrentSSESessions(t *testing.T) {
	// SSE handler with per-connection server factory (same pattern as main).
	handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return newServer()
	}, nil)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	const numClients = 3

	var wg sync.WaitGroup
	errs := make(chan error, numClients)

	for i := range numClients {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			transport := &mcp.SSEClientTransport{Endpoint: ts.URL}
			client := mcp.NewClient(
				&mcp.Implementation{
					Name:    fmt.Sprintf("test-client-%d", id),
					Version: "1.0.1",
				},
				nil,
			)

			session, err := client.Connect(ctx, transport, nil)
			if err != nil {
				errs <- fmt.Errorf("client %d: connect: %w", id, err)
				return
			}
			defer session.Close()

			// Call list_models — every session should succeed.
			listRes, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name: "list_models",
			})
			if err != nil {
				errs <- fmt.Errorf("client %d: list_models: %w", id, err)
				return
			}
			if len(listRes.Content) == 0 {
				errs <- fmt.Errorf("client %d: list_models returned empty content", id)
				return
			}

			// Call get_model_info — verifies tool dispatch works per-session.
			infoRes, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      "get_model_info",
				Arguments: map[string]any{"model_id": "gpt-5.2"},
			})
			if err != nil {
				errs <- fmt.Errorf("client %d: get_model_info: %w", id, err)
				return
			}
			if len(infoRes.Content) == 0 {
				errs <- fmt.Errorf("client %d: get_model_info returned empty content", id)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

// newTestMux builds the same mux as serveHTTP: /health (JSON) + SSE handler.
func newTestMux() *http.ServeMux {
	sseHandler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
		return newServer()
	}, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"models":  len(models.Models),
			"version": "1.0.1",
		})
	})
	mux.Handle("/", sseHandler)
	return mux
}

func TestHealthEndpoint(t *testing.T) {
	srv := httptest.NewServer(newTestMux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	var health map[string]any
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("health response is not valid JSON: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", health["status"])
	}
	if health["version"] != "1.0.1" {
		t.Errorf("expected version '1.0.1', got %v", health["version"])
	}
	if int(health["models"].(float64)) != len(models.Models) {
		t.Errorf("expected models %d, got %v", len(models.Models), health["models"])
	}
}

func TestHealthDoesNotAffectSSE(t *testing.T) {
	srv := httptest.NewServer(newTestMux())
	defer srv.Close()

	// Hit /health many times — none of these should create MCP sessions.
	for i := 0; i < 20; i++ {
		resp, err := http.Get(srv.URL + "/health")
		if err != nil {
			t.Fatalf("health request %d failed: %v", i, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("health request %d returned %d", i, resp.StatusCode)
		}
	}

	// Now verify that SSE still works: GET /sse should return 200 with
	// text/event-stream content type (the SSE handshake).
	resp, err := http.Get(srv.URL + "/sse")
	if err != nil {
		t.Fatalf("GET /sse failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected SSE status 200, got %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/event-stream") {
		t.Errorf("expected Content-Type text/event-stream, got %q", ct)
	}
}
