# Model ID Cheatsheet

MCP server exposing a curated, static registry of 64 AI models across 11 providers. Built in Go with the official MCP SDK.

## Architecture

Go server using `github.com/modelcontextprotocol/go-sdk` v1.3.0. Model data lives in a static Go map (`models.Models` in `internal/models/data.go`). The server exposes 6 tools and 3 resources over MCP (stdio, SSE, or streamable-http transport). HTTP transports (SSE, streamable-http) are protected by rate limiting and connection limits via `internal/middleware`. No external calls at runtime. Single ~10MB binary.

## Key Files

| File | Purpose |
|------|---------|
| `go-server/cmd/server/main.go` | Entry point — registers tools, resources, starts transport |
| `go-server/internal/models/data.go` | `Models` map with all 64 model entries |
| `go-server/internal/models/models.go` | `Model` struct definition |
| `go-server/internal/tools/*.go` | 6 tool handlers + shared helpers |
| `go-server/internal/resources/resources.go` | 3 resource handlers |
| `go-server/internal/models/data_test.go` | Data integrity tests |
| `go-server/internal/tools/tools_test.go` | Tool unit tests |
| `go-server/internal/middleware/` | Rate limiting and connection limit middleware |
| `Dockerfile` | Production container (Go multi-stage, SSE on port 8000) |

## Running

```bash
cd go-server

# Run tests
go test ./... -v

# Build
go build -o bin/server ./cmd/server

# Local server (stdio)
./bin/server

# Local server (SSE)
MCP_TRANSPORT=sse ./bin/server

# Docker
docker build -t model-registry -f ../Dockerfile .. && docker run -p 8000:8000 model-registry
```

## Adding a New Model

1. Add an entry to the `Models` map in `go-server/internal/models/data.go` following the existing schema (all 13 fields: ID, DisplayName, Provider, ContextWindow, MaxOutputTokens, Vision, Reasoning, PricingInput, PricingOutput, KnowledgeCutoff, ReleaseDate, Status, Notes)
2. Ensure `ID` matches the map key
3. Update `TestTotalModelCount` and `TestProviderCounts` in `data_test.go`
4. Run tests: `go test ./... -v`

## Adding a New Tool

1. Create a new file in `go-server/internal/tools/` with input struct + handler function
2. Register in `cmd/server/main.go` via `mcp.AddTool()`
3. Add tests in `tools_test.go`

## Coding Conventions

- Go 1.23+
- `golangci-lint` for linting
- `go fmt` for formatting
- Valid `status` values: `current`, `legacy`, `deprecated`
- Exported functions and types for cross-package use
