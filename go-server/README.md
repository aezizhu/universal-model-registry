# Universal Model Registry — Go Server

High-performance MCP server in Go exposing a curated registry of 42 AI models across 7 providers. No API keys, no database — just a compiled binary.

## Why Go?

- Single binary deployment (~10MB)
- Sub-millisecond response times
- No runtime dependencies
- Built on the official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk`)

## Quick Start

### Build and run locally

```bash
go build -o bin/server ./cmd/server
./bin/server                        # stdio transport (default)
MCP_TRANSPORT=sse ./bin/server      # SSE transport on :8000
```

### Using Docker

```bash
docker build -t model-registry .
docker run -p 8000:8000 model-registry
```

### Using Make

```bash
make build       # Build binary to bin/server
make run         # Run server (stdio)
make run-sse     # Run server (SSE transport)
make test        # Run all tests
make lint        # Run golangci-lint
make check       # Run lint + tests
make clean       # Remove build artifacts
```

## Available Tools (6)

| Tool | Parameters | Description |
|------|-----------|-------------|
| `list_models` | `provider?`, `status?`, `capability?` | Filtered markdown table of models |
| `get_model_info` | `model_id` | Full specs for a specific model |
| `recommend_model` | `task`, `budget?` | Best model for a task (top 3 recommendations) |
| `check_model_status` | `model_id` | Is this model current, legacy, or deprecated? |
| `compare_models` | `model_ids` (2-5) | Side-by-side comparison table |
| `search_models` | `query` | Free-text search across names, IDs, providers, notes |

## Resources (3)

| URI | Description |
|-----|-------------|
| `model://registry/all` | Full JSON dump of all models |
| `model://registry/current` | Only current models |
| `model://registry/pricing` | Pricing table sorted by cost |

## Connect to Your IDE

All configs use the deployed Railway SSE endpoint. Replace with `http://localhost:8000/sse` for local development.

### Claude Code

```bash
claude mcp add --transport sse --scope user universal-model-registry \
  https://universal-model-registry-production.up.railway.app/sse
```

### Codex

Add to `~/.codex/config.toml` (uses `mcp-proxy` to bridge stdio to SSE):

```toml
[mcp_servers.universal-model-registry]
command = "uvx"
args = ["mcp-proxy", "--transport", "sse", "https://universal-model-registry-production.up.railway.app/sse"]
```

### OpenCode

Add to `~/.config/opencode/opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "universal-model-registry": {
      "type": "remote",
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

### Cursor

Merge into `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "universal-model-registry": {
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

### Windsurf

Merge into your Windsurf MCP config:

```json
{
  "mcpServers": {
    "universal-model-registry": {
      "serverUrl": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

## Project Structure

```
go-server/
├── cmd/server/main.go          # Entry point, MCP server setup
├── internal/
│   ├── models/
│   │   ├── models.go           # Model struct definition
│   │   └── data.go             # Static MODELS map (42 entries)
│   └── tools/
│       ├── helpers.go          # Shared formatting and filtering
│       ├── list.go             # list_models tool
│       ├── info.go             # get_model_info tool
│       ├── recommend.go        # recommend_model tool
│       ├── status.go           # check_model_status tool
│       ├── compare.go          # compare_models tool
│       └── search.go           # search_models tool
├── Dockerfile                  # Multi-stage build (golang → alpine)
├── Makefile                    # Build, test, lint, run targets
├── go.mod
└── go.sum
```

## Covered Providers (7)

| Provider | Models | Highlights |
|----------|--------|-----------|
| **OpenAI** | 14 | GPT-5.2, GPT-5 series, o3/o4-mini reasoning |
| **Anthropic** | 8 | Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 |
| **Google** | 6 | Gemini 3 Pro/Flash, 2.5 series (1M context) |
| **xAI** | 4 | Grok 4, Grok 4.1 Fast (2M context) |
| **Meta** | 3 | Llama 4 Maverick/Scout (open-weight) |
| **Mistral** | 4 | Mistral Large 3, Devstral 2 (open-weight) |
| **DeepSeek** | 3 | DeepSeek Reasoner/Chat (open-weight MoE) |
