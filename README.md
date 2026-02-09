# Universal Model Registry (MCP Server)

A lightweight MCP server that exposes a curated, static registry of current AI models. Prevents coding agents from hallucinating outdated model names.

**42 models across 7 providers** | No API keys required | Static Python dict — no external calls | Auto-update detection via provider APIs

## Go Server (Recommended)

A high-performance Go rewrite is available in [`go-server/`](go-server/). Single binary, sub-millisecond responses, no runtime dependencies. See [go-server/README.md](go-server/README.md) for details.

## Deployed Instance

SSE endpoint: `https://universal-model-registry-production.up.railway.app/sse`

## Local Development

```bash
cd ~/Desktop/universal-model-registry
uv sync
uv run registry.py              # stdio transport (default)
MCP_TRANSPORT=sse uv run registry.py  # SSE transport (HTTP server)
```

Interactive testing with FastMCP inspector:

```bash
uv run fastmcp dev registry.py
```

## Available Tools

| Tool | Description |
|------|-------------|
| `list_models(provider?, status?, capability?)` | Filtered markdown table of models |
| `get_model_info(model_id)` | Full specs for a specific model |
| `recommend_model(task, budget?)` | Best model for a task |
| `check_model_status(model_id)` | Is this model current, legacy, or deprecated? |
| `compare_models(model_ids)` | Side-by-side comparison of 2-5 models |
| `search_models(query)` | Free-text search across names, IDs, providers, notes |

## Resources

| URI | Description |
|-----|-------------|
| `model://registry/all` | Full JSON dump |
| `model://registry/current` | Only current models |
| `model://registry/pricing` | Pricing table sorted by cost |

## Connect to Your IDE

All configs use the deployed Railway SSE endpoint. See `configs/` for copy-paste examples.

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

## Covered Providers & Models

- **OpenAI:** GPT-5.2, GPT-5.2 Codex, GPT-5.1, GPT-5, GPT-5 Mini, GPT-5 Nano, GPT-4.1 Mini/Nano, o3, o4-mini, o3-mini + legacy GPT-4.1, deprecated GPT-4o/4o-mini
- **Anthropic:** Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 + legacy Opus 4.5/4.1/4.0, Sonnet 4.0 + deprecated 3.7 Sonnet
- **Google:** Gemini 3 Pro/Flash (preview), 2.5 Pro/Flash/Flash Lite + deprecated 2.0 Flash
- **xAI:** Grok 4, Grok 4.1 Fast + legacy Grok 3/3 Mini
- **Meta:** Llama 4 Maverick, Llama 4 Scout + legacy Llama 3.3 70B
- **Mistral:** Mistral Large 3, Mistral Small, Devstral 2 + legacy Codestral
- **DeepSeek:** DeepSeek Reasoner, DeepSeek Chat + deprecated DeepSeek V3

## Quick Stats

| Provider | Current | Legacy | Deprecated | Total |
|----------|---------|--------|------------|-------|
| OpenAI | 11 | 1 | 2 | 14 |
| Anthropic | 3 | 4 | 1 | 8 |
| Google | 5 | 0 | 1 | 6 |
| xAI | 2 | 2 | 0 | 4 |
| Meta | 2 | 1 | 0 | 3 |
| Mistral | 3 | 1 | 0 | 4 |
| DeepSeek | 2 | 0 | 1 | 3 |
| **Total** | **28** | **9** | **5** | **42** |

## Auto-Update Detection

The registry includes an auto-update script that queries provider APIs to detect new or retired models:

```bash
make auto-update                         # Run locally (needs API keys in env)
uv run python scripts/auto_update.py     # Direct invocation
```

CI runs this weekly (Sunday midnight UTC) and creates a GitHub issue if changes are detected.
