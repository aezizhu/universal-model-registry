package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"go-server/internal/models"
	"go-server/internal/resources"
	"go-server/internal/tools"
)

// Tool input types matching the SDK's ToolHandlerFor generic pattern.

type GetModelInfoInput struct {
	ModelID string `json:"model_id" jsonschema:"The API model ID string"`
}

type SearchModelsInput struct {
	Query string `json:"query" jsonschema:"Search term to match against model names and notes"`
}

func main() {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "universal-model-registry",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: "Query this server to get accurate, up-to-date information about AI models. " +
				"Use list_models to browse, get_model_info for details, recommend_model for " +
				"task-based suggestions, and check_model_status to verify if a model ID is " +
				"current, legacy, or deprecated.",
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
		result := tools.GetModelInfo(input.ModelID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_models",
		Description: "Search for models by keyword across names, providers, and notes.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input SearchModelsInput) (*mcp.CallToolResult, any, error) {
		result := tools.SearchModels(input.Query)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "recommend_model",
		Description: "Recommend the best model for a given task and budget.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.RecommendModelInput) (*mcp.CallToolResult, any, error) {
		result := tools.RecommendModel(input.Task, input.Budget)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_model_status",
		Description: "Check whether a model ID is current, legacy, or deprecated.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.CheckModelStatusInput) (*mcp.CallToolResult, any, error) {
		result := tools.CheckModelStatus(input.ModelID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: result}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "compare_models",
		Description: "Compare 2-5 models side by side in a markdown table.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, input tools.CompareModelsInput) (*mcp.CallToolResult, any, error) {
		result := tools.CompareModels(input.ModelIDs)
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

	// ── Start Server ────────────────────────────────────────────────────

	fmt.Fprintf(os.Stderr, "Universal Model Registry — %d models loaded\n", len(models.Models))

	transport := os.Getenv("MCP_TRANSPORT")
	switch transport {
	case "sse":
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000"
		}
		addr := ":" + port
		fmt.Fprintf(os.Stderr, "Starting SSE server on %s\n", addr)
		handler := mcp.NewSSEHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)
		log.Fatal(http.ListenAndServe(addr, handler))

	case "streamable-http":
		port := os.Getenv("PORT")
		if port == "" {
			port = "8000"
		}
		addr := ":" + port
		fmt.Fprintf(os.Stderr, "Starting Streamable HTTP server on %s\n", addr)
		handler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
			return server
		}, nil)
		log.Fatal(http.ListenAndServe(addr, handler))

	default:
		// stdio transport (default)
		fmt.Fprintln(os.Stderr, "Starting stdio transport")
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
