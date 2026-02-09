# Universal Model Registry (MCP Server)

A lightweight MCP server that exposes a curated, static registry of current AI models. Prevents coding agents from hallucinating outdated model names.

**No API keys required.** Data is a static Python dict — no external calls.

## Setup

```bash
cd ~/Desktop/universal-model-registry
uv sync
```

## Run

```bash
uv run registry.py
```

Or use the FastMCP inspector for interactive testing:

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

### Claude Code

Copy `configs/claude_mcp.json` to your project's `.mcp.json`, or add the server entry to `~/.claude/claude_code_config.json`.

### Cursor

Merge `configs/cursor_mcp.json` into `~/.cursor/mcp.json`.

### Windsurf

Merge `configs/windsurf_mcp.json` into your Windsurf MCP config.

**Important:** Update the `--directory` path in the config if you move the project.

## Covered Providers & Models

- **OpenAI:** GPT-5.2, GPT-5, GPT-5 Mini, GPT-4.1 series, o3, o4-mini, o3-mini + legacy GPT-4o
- **Anthropic:** Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 + legacy Opus 4.5/4.0, Sonnet 4.0, 3.7
- **Google:** Gemini 3 Pro/Flash (preview), 2.5 Pro/Flash/Flash Lite + deprecated 2.0 Flash
