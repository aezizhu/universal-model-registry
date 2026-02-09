# Universal Model Registry (MCP Server)

A lightweight MCP server that exposes a curated, static registry of current AI models. Prevents coding agents from hallucinating outdated model names.

**No API keys required.** Data is a static Python dict — no external calls.

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

## Resources

| URI | Description |
|-----|-------------|
| `model://registry/all` | Full JSON dump |
| `model://registry/current` | Only current models |

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

- **OpenAI:** GPT-5.2, GPT-5, GPT-5 Mini, GPT-4.1 series, o3, o4-mini, o3-mini + legacy GPT-4o
- **Anthropic:** Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 + legacy Opus 4.5/4.0, Sonnet 4.0, 3.7
- **Google:** Gemini 3 Pro/Flash (preview), 2.5 Pro/Flash/Flash Lite + deprecated 2.0 Flash
