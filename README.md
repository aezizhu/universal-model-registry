# Model ID Cheatsheet (MCP Server)

Stop your AI coding agent from hallucinating outdated model names. This MCP server gives any AI assistant instant access to accurate, up-to-date API model IDs, pricing, and specs for **46 models across 7 providers**.

Built in Go. Single 10MB binary. Zero external calls. Sub-millisecond responses. Auto-updated weekly.

## Why?

When you ask an AI coding agent to "use the latest OpenAI model", it might generate `gpt-4-turbo` (deprecated) instead of `gpt-5.2` (current). This MCP server solves that by giving your agent a tool it can call to look up the correct model ID before writing code.

**Example:** Your agent needs to write an API call. Instead of guessing, it calls `get_model_info("gpt-5.2")` and gets back the exact API model ID, pricing, context window, and capabilities — verified against official docs.

## Quick Install (Pick Your IDE)

### Claude Code (one command)

```bash
claude mcp add --transport sse --scope user model-id-cheatsheet \
  https://universal-model-registry-production.up.railway.app/sse
```

That's it. Claude Code can now call `list_models`, `get_model_info`, `recommend_model`, etc.

### Cursor

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

### Windsurf

Add to your Windsurf MCP config (Settings > MCP Servers):

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "serverUrl": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

### Codex CLI

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.model-id-cheatsheet]
command = "uvx"
args = ["mcp-proxy", "--transport", "sse", "https://universal-model-registry-production.up.railway.app/sse"]
```

### OpenCode

Add to `~/.config/opencode/opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "model-id-cheatsheet": {
      "type": "remote",
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

### Any MCP Client (Generic)

Connect to the SSE endpoint:

```
https://universal-model-registry-production.up.railway.app/sse
```

No API key needed. No auth. Just point your MCP client at that URL.

## Self-Hosting

If you prefer to run your own instance instead of using the hosted one:

### Option 1: Docker (Recommended)

```bash
git clone https://github.com/aezizhu/universal-model-registry.git
cd universal-model-registry
docker build -t model-id-cheatsheet .
docker run -p 8000:8000 model-id-cheatsheet
```

Your SSE endpoint will be at `http://localhost:8000/sse`.

### Option 2: Build from Source

Requires Go 1.23+.

```bash
git clone https://github.com/aezizhu/universal-model-registry.git
cd universal-model-registry/go-server
go build -o server ./cmd/server

# stdio mode (for local MCP clients like Claude Code local)
./server

# SSE mode (for HTTP-based clients)
MCP_TRANSPORT=sse PORT=8000 ./server
```

### Option 3: Deploy to Railway

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/template)

Or manually:

```bash
railway login
railway init
railway up
```

### Local MCP Config (self-hosted)

For Claude Code connecting to a local instance:

```bash
claude mcp add --transport sse --scope user model-id-cheatsheet \
  http://localhost:8000/sse
```

For Cursor (local):

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "url": "http://localhost:8000/sse"
    }
  }
}
```

## Available Tools

Your AI agent gets these 6 tools:

| Tool | What It Does | Example Use |
|------|-------------|-------------|
| `get_model_info(model_id)` | Get exact API ID, pricing, context window, capabilities | "What's the model ID for Claude Sonnet?" |
| `list_models(provider?, status?, capability?)` | Browse models with filters | "Show me all current Google models" |
| `recommend_model(task, budget?)` | Get ranked recommendations for a task | "Best model for coding, cheap budget" |
| `check_model_status(model_id)` | Is this model current, legacy, or deprecated? | "Is gpt-4o still available?" |
| `compare_models(model_ids)` | Side-by-side comparison table | "Compare gpt-5.2 vs claude-opus-4-6" |
| `search_models(query)` | Free-text search across all fields | "Search for reasoning models" |

## Resources

| URI | Description |
|-----|-------------|
| `model://registry/all` | Full JSON dump of all 46 models |
| `model://registry/current` | Only current (non-deprecated) models as JSON |
| `model://registry/pricing` | Pricing table sorted cheapest-first (markdown) |

## Covered Models (46 total)

### Current Models (32)

| Provider | Models | API IDs |
|----------|--------|---------|
| **OpenAI** (13) | GPT-5.2, GPT-5.2 Codex, GPT-5.2 Pro, GPT-5.1, GPT-5, GPT-5 Mini, GPT-5 Nano, GPT-4.1 Mini, GPT-4.1 Nano, o3, o3 Pro, o4-mini, o3-mini | `gpt-5.2`, `gpt-5.2-codex`, `gpt-5.2-pro`, `gpt-5.1`, `gpt-5`, `gpt-5-mini`, `gpt-5-nano`, `gpt-4.1-mini`, `gpt-4.1-nano`, `o3`, `o3-pro`, `o4-mini`, `o3-mini` |
| **Anthropic** (3) | Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 | `claude-opus-4-6`, `claude-sonnet-4-5-20250929`, `claude-haiku-4-5-20251001` |
| **Google** (5) | Gemini 3 Pro, Gemini 3 Flash, Gemini 2.5 Pro, 2.5 Flash, 2.5 Flash Lite | `gemini-3-pro-preview`, `gemini-3-flash-preview`, `gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.5-flash-lite` |
| **xAI** (4) | Grok 4, Grok 4.1 Fast, Grok 4 Fast, Grok Code Fast 1 | `grok-4`, `grok-4.1-fast`, `grok-4-fast`, `grok-code-fast-1` |
| **Meta** (2) | Llama 4 Maverick, Llama 4 Scout | `llama-4-maverick`, `llama-4-scout` |
| **Mistral** (3) | Mistral Large 3, Mistral Small 3.2, Devstral 2 | `mistral-large-2512`, `mistral-small-2506`, `devstral-2512` |
| **DeepSeek** (2) | DeepSeek Reasoner, DeepSeek Chat | `deepseek-reasoner`, `deepseek-chat` |

### Legacy & Deprecated Models (14)

Also tracked so your agent can detect outdated model IDs and suggest replacements:

- OpenAI: `gpt-4.1` (legacy), `gpt-4o` (deprecated), `gpt-4o-mini` (deprecated)
- Anthropic: `claude-opus-4-5` (legacy), `claude-opus-4-1` (legacy), `claude-opus-4-0` (legacy), `claude-sonnet-4-0` (legacy), `claude-3-7-sonnet-20250219` (deprecated)
- Google: `gemini-2.0-flash` (deprecated)
- xAI: `grok-3` (legacy), `grok-3-mini` (legacy)
- Meta: `llama-3.3-70b` (legacy)
- Mistral: `codestral-2508` (legacy)
- DeepSeek: `deepseek-v3` (deprecated)

## How It Helps Your Coding Agent

**Scenario 1: Writing an API call**
```
You: "Call the OpenAI API with their best coding model"
Agent calls: get_model_info("gpt-5.2-codex")
Agent writes: model="gpt-5.2-codex"  # Correct!
```

**Scenario 2: Checking if a model is still valid**
```
You: "Use gpt-4o for this task"
Agent calls: check_model_status("gpt-4o")
Agent responds: "gpt-4o is deprecated (retiring Feb 13, 2026). I'll use gpt-5 instead."
```

**Scenario 3: Finding the cheapest option**
```
You: "Use the cheapest model that supports vision"
Agent calls: list_models(capability="vision", status="current")
Agent picks: gpt-5-nano at $0.05/$0.40 per 1M tokens
```

**Scenario 4: Comparing options**
```
You: "Should I use Claude or GPT for this?"
Agent calls: compare_models(["claude-opus-4-6", "gpt-5.2"])
Agent gets: Side-by-side table of pricing, context, capabilities
```

## Security

### Built-in Protection

- **Rate limiting**: 60 requests/minute per IP
- **Connection limits**: Max 5 SSE connections per IP, 100 total
- **Request body limit**: 64KB max
- **Input sanitization**: All string inputs truncated to safe lengths
- **HTTP hardening**: ReadTimeout 15s, ReadHeaderTimeout 5s, IdleTimeout 120s, 64KB max headers
- **Non-root Docker**: Containers run as unprivileged user
- **Graceful shutdown**: Clean connection draining on SIGINT/SIGTERM

## Staying Up to Date

Model data is automatically checked every Monday via a GitHub Actions workflow. The updater queries each provider's API to detect new models, deprecations, or pricing changes, and opens a GitHub issue if updates are needed. You can also trigger it manually via `workflow_dispatch`.

## Tech Stack

- **Language**: Go 1.23
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` v1.3.0 (official)
- **Transports**: stdio, SSE, Streamable HTTP
- **Binary size**: ~10MB
- **Tests**: 83 unit tests
- **Security**: Per-IP rate limiting, connection limits, input sanitization
- **Deploy**: Docker (alpine), Railway

## Contributing

To add or update a model:

1. Edit `go-server/internal/models/data.go`
2. Update test counts in `go-server/internal/models/data_test.go`
3. Run `cd go-server && go test ./... -v`
4. Submit a PR

## License

MIT
