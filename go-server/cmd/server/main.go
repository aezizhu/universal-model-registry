package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go-server/internal/middleware"
	"go-server/internal/models"
	"go-server/internal/resources"
	"go-server/internal/tools"
)

var startTime = time.Now()

// Tool input types matching the SDK's ToolHandlerFor generic pattern.

type GetModelInfoInput struct {
	ModelID string `json:"model_id" jsonschema:"The API model ID string"`
}

type SearchModelsInput struct {
	Query string `json:"query" jsonschema:"Search term to match against model names and notes"`
}

// newServer creates a fresh MCP server with all tools and resources registered.
// Each SSE/HTTP session needs its own server instance to avoid shared state issues.
func newServer() *mcp.Server {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "model-id-cheatsheet",
			Version: "1.0.1",
		},
		&mcp.ServerOptions{
			Instructions: "Query this server to get accurate, up-to-date information about AI models. " +
				"Use list_models to browse, get_model_info for details, recommend_model for " +
				"task-based suggestions, and check_model_status to verify if a model ID is " +
				"current, legacy, or deprecated. " +
				"Always prefer the newest model (by release date) when recommending or writing code. " +
				"When a user specifies a model ID, use check_model_status to verify it's current. " +
				"If it's legacy or deprecated, suggest the newest replacement from the same provider. " +
				"When listing models, the newest model per provider is marked with ★.",
		},
	)

	// ── Register Tools ──────────────────────────────────────────────────

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_models",
		Description: "List AI models with optional filters for provider, status, and capability.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.ListModelsInput) (*mcp.CallToolResult, any, error) {
		result := tools.ListModels(input.Provider, input.Status, input.Capability)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_model_info",
		Description: "Get full specifications for a specific model by its API model ID.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input GetModelInfoInput) (*mcp.CallToolResult, any, error) {
		result := tools.GetModelInfo(truncate(input.ModelID, 256))
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_models",
		Description: "Search for models by keyword across names, providers, and notes.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input SearchModelsInput) (*mcp.CallToolResult, any, error) {
		result := tools.SearchModels(truncate(input.Query, 512))
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "recommend_model",
		Description: "Recommend the best model for a given task and budget.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.RecommendModelInput) (*mcp.CallToolResult, any, error) {
		result := tools.RecommendModel(truncate(input.Task, 1024), truncate(input.Budget, 64))
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_model_status",
		Description: "Check whether a model ID is current, legacy, or deprecated.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.CheckModelStatusInput) (*mcp.CallToolResult, any, error) {
		result := tools.CheckModelStatus(truncate(input.ModelID, 256))
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_models",
		Description: "Compare 2-5 models side by side in a markdown table.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.CompareModelsInput) (*mcp.CallToolResult, any, error) {
		ids := input.ModelIDs
		for i := range ids {
			ids[i] = truncate(ids[i], 256)
		}
		result := tools.CompareModels(ids)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	// ── Register Resources ──────────────────────────────────────────────

	server.AddResource(
		&mcp.Resource{
			URI:         "model://registry/all",
			Name:        "all-models",
			Description: "Full JSON dump of the entire model registry.",
			MIMEType:    "application/json",
		},
		func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     resources.AllModels(),
				}},
			}, nil
		},
	)

	server.AddResource(
		&mcp.Resource{
			URI:         "model://registry/current",
			Name:        "current-models",
			Description: "JSON dump of only current (non-legacy, non-deprecated) models.",
			MIMEType:    "application/json",
		},
		func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "application/json",
					Text:     resources.CurrentModels(),
				}},
			}, nil
		},
	)

	server.AddResource(
		&mcp.Resource{
			URI:         "model://registry/pricing",
			Name:        "pricing-summary",
			Description: "Markdown table of all current models sorted by input pricing (cheapest first).",
			MIMEType:    "text/markdown",
		},
		func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      req.Params.URI,
					MIMEType: "text/markdown",
					Text:     resources.PricingSummary(),
				}},
			}, nil
		},
	)

	return server
}

func main() {
	fmt.Fprintf(os.Stderr, "Model ID Cheatsheet — %d models loaded\n", len(models.Models))

	transport := os.Getenv("MCP_TRANSPORT")
	switch transport {
	case "sse":
		serveHTTP("SSE", mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
			return newServer()
		}, nil))

	case "streamable-http":
		serveHTTP("Streamable HTTP", mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
			return newServer()
		}, nil))

	default:
		// stdio transport (default) — single session, one server is fine.
		fmt.Fprintln(os.Stderr, "Starting stdio transport")
		if err := newServer().Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// serveHTTP starts an HTTP server with timeouts, rate limiting, and graceful shutdown.
func serveHTTP(label string, handler http.Handler) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	addr := ":" + port

	// Health endpoint that doesn't create MCP sessions.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":      "ok",
			"models":      len(models.Models),
			"version":     "1.0.1",
			"uptime_secs": int(time.Since(startTime).Seconds()),
			"transport":   label,
		})
	})
	mux.Handle("/", handler)

	limiter := middleware.NewLimiter(middleware.DefaultConfig())
	protected := limiter.Wrap(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           protected,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      0, // SSE requires no write timeout (long-lived streams).
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16, // 64KB max headers.
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nShutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		}
		close(done)
	}()

	fmt.Fprintf(os.Stderr, "Starting %s server on %s (rate limit: %d req/min, max %d conns)\n",
		label, addr, middleware.DefaultConfig().RequestsPerWindow, middleware.DefaultConfig().MaxTotalConns)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	<-done
}

// truncate limits string length to prevent abuse from oversized inputs.
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
