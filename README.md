[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Models](https://img.shields.io/badge/Models-97-blueviolet)](https://github.com/aezizhu/universal-model-registry)
[![Providers](https://img.shields.io/badge/Providers-19-orange)](https://github.com/aezizhu/universal-model-registry)
[![Tests](https://img.shields.io/badge/Tests-123%20passing-brightgreen)](https://github.com/aezizhu/universal-model-registry)

# Model ID Cheatsheet

**Stop your AI coding agent from hallucinating outdated model names.** This plugin gives any AI assistant instant access to accurate, up-to-date API model IDs, pricing, and specs for **97 models across 19 providers**.

Built in Go. Single 10MB binary. Zero external calls. Sub-millisecond responses. Auto-updated daily.

```diff
- model = "gpt-4-turbo"           # Hallucinated — doesn't exist anymore
+ model = "gpt-5.2"               # Correct — verified against official docs
```

```diff
- model = "claude-3-opus-20240229" # Deprecated
+ model = "claude-opus-4-6"        # Current — latest Anthropic flagship
```

---

## Quick Start

> **Claude Code Plugin** (recommended — one command, done)
>
> ```bash
> claude plugin add -- https://github.com/aezizhu/universal-model-registry.git
> ```
>
> Claude Code gets **6 tools** + a smart lookup skill that **automatically verifies model IDs** before writing code or answering any question about AI models.

### Alternative: Direct MCP Connection

Connect any MCP-compatible client to the hosted server — no API key, no auth:

**Claude Code:**

```bash
claude mcp add --transport sse --scope user model-id-cheatsheet \
  https://universal-model-registry-production.up.railway.app/sse
```

**Cursor** — add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

**Windsurf** — add to Settings > MCP Servers:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "serverUrl": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

**Codex CLI** — add to `~/.codex/config.toml`:

```toml
[mcp_servers.model-id-cheatsheet]
command = "uvx"
args = ["mcp-proxy", "--transport", "sse", "https://universal-model-registry-production.up.railway.app/sse"]
```

**OpenCode** — add to `~/.config/opencode/opencode.json`:

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

**Any MCP Client** — connect to the SSE endpoint directly:

```
https://universal-model-registry-production.up.railway.app/sse
```

---

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

### Resources

| URI | Description |
|-----|-------------|
| `model://registry/all` | Full JSON dump of all 97 models |
| `model://registry/current` | Only current (non-deprecated) models as JSON |
| `model://registry/pricing` | Pricing table sorted cheapest-first (markdown) |

---

## How It Helps Your AI Agent

**Scenario 1: Writing an API call**

```python
# You: "Call the OpenAI API with their best coding model"
# Agent calls: get_model_info("gpt-5.2-codex")
# Agent writes:
response = client.chat.completions.create(
    model="gpt-5.2-codex",  # Correct! Verified via model registry
    messages=[...]
)
```

**Scenario 2: Catching deprecated models**

```python
# You: "Use gpt-4o for this task"
# Agent calls: check_model_status("gpt-4o")
# Agent responds: "gpt-4o is deprecated (retiring Feb 13, 2026). I'll use gpt-5 instead."
response = client.chat.completions.create(
    model="gpt-5",  # Updated automatically
    messages=[...]
)
```

**Scenario 3: Finding the cheapest option**

```python
# You: "Use the cheapest model that supports vision"
# Agent calls: list_models(capability="vision", status="current")
# Agent picks: gpt-5-nano at $0.05/$0.40 per 1M tokens
response = client.chat.completions.create(
    model="gpt-5-nano",  # Cheapest vision model
    messages=[...]
)
```

**Scenario 4: Answering model questions (not just code)**

```
# You: "What's the latest Gemini model?"
# Agent calls: list_models(provider="Google", status="current")
# Agent responds with verified info: "The latest is Gemini 3 Pro (gemini-3-pro-preview),
#   released January 2026. There's also Gemini 3 Flash for faster/cheaper use."
```

**Scenario 5: Comparing options**

```python
# You: "Should I use Claude or GPT for this?"
# Agent calls: compare_models(["claude-opus-4-6", "gpt-5.2"])
# Agent gets: Side-by-side table of pricing, context, capabilities
```

---

## Covered Models (97 total)

### Current Models (71)

| Provider | Models | API IDs |
|----------|--------|---------|
| **OpenAI** (12) | GPT-5.2, GPT-5.2 Pro, GPT-5.1, GPT-5.1 Codex, GPT-5.1 Mini, GPT-5, GPT-5 Mini, GPT-5 Nano, GPT-4.1 Mini, GPT-4.1 Nano, o3, o4-mini | `gpt-5.2`, `gpt-5.2-pro`, `gpt-5.1`, `gpt-5.1-codex`, `gpt-5.1-mini`, `gpt-5`, `gpt-5-mini`, `gpt-5-nano`, `gpt-4.1-mini`, `gpt-4.1-nano`, `o3`, `o4-mini` |
| **Anthropic** (4) | Claude Sonnet 4.6, Claude Opus 4.6, Sonnet 4.5, Haiku 4.5 | `claude-sonnet-4-6`, `claude-opus-4-6`, `claude-sonnet-4-5-20250929`, `claude-haiku-4-5-20251001` |
| **Google** (4) | Gemini 3 Pro, Gemini 3 Flash, Gemini 2.5 Pro, 2.5 Flash | `gemini-3-pro-preview`, `gemini-3-flash-preview`, `gemini-2.5-pro`, `gemini-2.5-flash` |
| **xAI** (4) | Grok 4, Grok 4.1 Fast, Grok 4 Fast, Grok Code Fast 1 | `grok-4`, `grok-4.1-fast`, `grok-4-fast`, `grok-code-fast-1` |
| **Meta** (2) | Llama 4 Maverick, Llama 4 Scout | `llama-4-maverick`, `llama-4-scout` |
| **Mistral** (10) | Mistral Large 3, Mistral Medium 3, Mistral Small 3.2, Ministral 3B, Ministral 8B, Ministral 14B, Magistral Small 1.2, Magistral Medium 1.2, Devstral 2, Devstral Small 2 | `mistral-large-2512`, `mistral-medium-2505`, `mistral-small-2506`, `ministral-3b-2512`, `ministral-8b-2512`, `ministral-14b-2512`, `magistral-small-2509`, `magistral-medium-2509`, `devstral-2512`, `devstral-small-2512` |
| **DeepSeek** (2) | DeepSeek Reasoner, DeepSeek Chat | `deepseek-reasoner`, `deepseek-chat` |
| **Amazon** (6) | Nova Micro, Nova Lite, Nova Pro, Nova Premier, Nova 2 Lite, Nova 2 Pro | `amazon-nova-micro`, `amazon-nova-lite`, `amazon-nova-pro`, `amazon-nova-premier`, `amazon-nova-2-lite`, `amazon-nova-2-pro` |
| **Cohere** (4) | Command A, Command A Reasoning, Command A Vision, Command R7B | `command-a-03-2025`, `command-a-reasoning-08-2025`, `command-a-vision-07-2025`, `command-r7b-12-2024` |
| **Perplexity** (4) | Sonar, Sonar Pro, Sonar Reasoning Pro, Sonar Deep Research | `sonar`, `sonar-pro`, `sonar-reasoning-pro`, `sonar-deep-research` |
| **AI21** (2) | Jamba Large 1.7, Jamba Mini 1.7 | `jamba-large-1.7`, `jamba-mini-1.7` |
| **Moonshot** (3) | Kimi K2.5, Kimi K2 Thinking, Kimi K2 (0905) | `kimi-k2.5`, `kimi-k2-thinking`, `kimi-k2-0905-preview` |
| **Zhipu** (3) | GLM-4.7, GLM-4.7 FlashX, GLM-4.6V | `glm-4.7`, `glm-4.7-flashx`, `glm-4.6v` |
| **NVIDIA** (2) | Nemotron 3 Nano 30B, Nemotron Ultra 253B | `nvidia/nemotron-3-nano-30b-a3b`, `nvidia/llama-3.1-nemotron-ultra-253b-v1` |
| **Tencent** (3) | Hunyuan TurboS, Hunyuan T1, Hunyuan A13B | `hunyuan-turbos`, `hunyuan-t1`, `hunyuan-a13b` |
| **Microsoft** (3) | Phi-4, Phi-4 Multimodal, Phi-4 Reasoning Plus | `phi-4`, `phi-4-multimodal-instruct`, `phi-4-reasoning-plus` |
| **MiniMax** (1) | MiniMax M2.1 | `minimax-m2.1` |
| **Xiaomi** (1) | MiMo V2 Flash | `mimo-v2-flash` |
| **Kuaishou** (1) | KAT-Coder Pro | `kat-coder-pro` |

### Legacy & Deprecated Models (26)

Also tracked so your agent can detect outdated model IDs and suggest replacements:

- OpenAI: `o3-mini` (legacy), `gpt-5.2-codex` (deprecated), `gpt-5.1-codex-mini` (deprecated), `o3-pro` (deprecated), `o3-deep-research` (deprecated), `gpt-4.1` (deprecated), `gpt-4o` (deprecated), `gpt-4o-mini` (deprecated)
- Anthropic: `claude-opus-4-5` (legacy), `claude-opus-4-1` (legacy), `claude-opus-4-0` (legacy), `claude-sonnet-4-0` (legacy), `claude-3-7-sonnet-20250219` (deprecated)
- Google: `gemini-3-pro-image-preview` (deprecated), `gemini-2.5-flash-lite` (deprecated), `gemini-2.0-flash-lite` (deprecated), `gemini-2.0-flash` (deprecated)
- xAI: `grok-4.1` (deprecated), `grok-3` (legacy), `grok-3-mini` (legacy)
- Meta: `llama-3.3-70b` (legacy)
- Mistral: `codestral-2508` (legacy)
- DeepSeek: `deepseek-r1` (legacy), `deepseek-v3` (deprecated)
- Zhipu: `glm-4.6v` (deprecated)
- MiniMax: `minimax-01` (deprecated)

---

<details>
<summary><strong>Self-Hosting</strong></summary>

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

</details>

<details>
<summary><strong>Local MCP Config (self-hosted)</strong></summary>

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

</details>

---

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

Model data is automatically checked and updated **daily at 7 PM Pacific Time** — no human intervention needed.

**How it works:**
1. Railway cron runs the updater daily, scraping 6 providers' public documentation pages (no API keys needed)
2. **Models removed from docs** → auto-deprecated via PR (status changed to `"deprecated"` in code)
3. **New models detected** → GitHub issue created for review
4. CI runs on the auto-generated PR → if tests pass → **auto-merged** into main
5. Railway auto-deploys from main

**No provider API keys required.** The updater reads publicly available documentation pages to detect model changes. Only `GITHUB_TOKEN` and `GITHUB_REPO` are needed for creating PRs and issues.

<details>
<summary><strong>Auto-Update Pipeline Details</strong></summary>

**Railway Cron (primary)** — The hosted instance uses a Railway cron service that runs the updater daily. See `configs/railway-updater.toml` for the configuration.

Required env vars (set in Railway dashboard):
- `GITHUB_TOKEN` — GitHub personal access token with repo scope
- `GITHUB_REPO` — Repository in `"owner/repo"` format (e.g. `"aezizhu/universal-model-registry"`)

**Providers checked (via public docs):**
- OpenAI (via GitHub SDK source), Anthropic, Google, Mistral, xAI, DeepSeek

**CI/CD Workflows:**
- `.github/workflows/ci.yml` — runs tests on every PR
- `.github/workflows/auto-merge.yml` — auto-merges bot PRs (labeled `auto-update`) after CI passes

**GitHub Actions (alternative)** — A GitHub Actions workflow is also included at `.github/workflows/auto-update.yml` for users who self-host without Railway. No API keys needed — only `GITHUB_TOKEN` (automatically provided by GitHub Actions).

</details>

## Tech Stack

- **Language**: Go 1.23
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` v1.3.0 (official)
- **Transports**: stdio, SSE, Streamable HTTP
- **Binary size**: ~10MB
- **Tests**: 123 unit tests
- **Security**: Per-IP rate limiting, connection limits, input sanitization
- **Deploy**: Docker (alpine), Railway

## Contributing

Contributions are welcome! Whether it's adding a new model, fixing data, or improving the server — here's how to get started:

1. Fork the repo and clone it locally
2. Edit model data in `go-server/internal/models/data.go`
3. Update test counts in `go-server/internal/models/data_test.go`
4. Run the tests:
   ```bash
   cd go-server && go test ./... -v
   ```
5. Submit a PR — we'll review it quickly

If you spot an outdated model or incorrect pricing, opening an issue is just as helpful.

## License

MIT
