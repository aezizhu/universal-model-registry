# Installation Guide

Connect the **model-id-cheatsheet** MCP server to your AI coding tool. Choose your client below.

**Hosted endpoint (no setup required):**
```
https://universal-model-registry-production.up.railway.app/sse
```

---

## Claude Code

One command:

```bash
claude mcp add model-id-cheatsheet --transport sse https://universal-model-registry-production.up.railway.app/sse
```

Verify it's connected:

```bash
claude mcp list
```

To remove:

```bash
claude mcp remove model-id-cheatsheet
```

---

## Cursor

Add to your MCP config file at `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

Restart Cursor to pick up the change.

---

## Windsurf

Open Windsurf settings and navigate to the MCP section. Add a new server with:

- **Name:** `model-id-cheatsheet`
- **Type:** SSE
- **URL:** `https://universal-model-registry-production.up.railway.app/sse`

Or edit `~/.codeium/windsurf/mcp_config.json` directly:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "serverUrl": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

---

## Codex CLI

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.model-id-cheatsheet]
type = "sse"
url = "https://universal-model-registry-production.up.railway.app/sse"
```

---

## OpenCode

Add to `~/.config/opencode/opencode.json`:

```json
{
  "mcpServers": {
    "model-id-cheatsheet": {
      "type": "sse",
      "url": "https://universal-model-registry-production.up.railway.app/sse"
    }
  }
}
```

---

## Self-Hosting

If you prefer to run the server yourself instead of using the hosted endpoint.

### Docker

```bash
docker pull ghcr.io/aezizhu/universal-model-registry:latest
docker run -p 8000:8000 ghcr.io/aezizhu/universal-model-registry:latest
```

Then point your client to `http://localhost:8000/sse`.

### Build from Source

```bash
git clone https://github.com/aezizhu/universal-model-registry.git
cd universal-model-registry/go-server
go build -o bin/server ./cmd/server
MCP_TRANSPORT=sse ./bin/server
```

The server starts on port 8000 by default. Point your client to `http://localhost:8000/sse`.

### Deploy to Railway

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/template)

1. Fork [the repository](https://github.com/aezizhu/universal-model-registry)
2. Create a new project on [Railway](https://railway.com)
3. Connect your forked repo
4. Railway auto-detects the Dockerfile and deploys
5. Use the generated Railway URL with `/sse` appended as your server endpoint

---

## Verify Connection

Once connected, try asking your AI assistant:

- "What's the correct model ID for Claude Opus 4.6?"
- "Compare GPT-5.2 vs Claude Sonnet 4.5"
- "Which models support vision under $5/M input tokens?"

The assistant should use the model-id-cheatsheet tools to answer with accurate, up-to-date information.
