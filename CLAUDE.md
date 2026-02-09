# Universal Model Registry

MCP server exposing a curated, static registry of current AI models. No API keys, no database — just a Python dict.

## Architecture

Flat-file FastMCP server. Model data lives in a static Python dict (`MODELS` in `models_data.py`). The server (`registry.py`) exposes tools and resources over MCP (stdio or SSE transport). No external calls at runtime.

## Key Files

| File | Purpose |
|------|---------|
| `registry.py` | MCP server entry point — tools, resources, helpers |
| `models_data.py` | `MODELS` dict with all model entries |
| `tests/test_registry.py` | Pytest tests for all tools and data integrity |
| `pyproject.toml` | Project config, Python >=3.12, deps: `fastmcp` |
| `Dockerfile` | Production container (SSE transport on port 8000) |

## Running

```bash
# Install deps and run tests
uv sync && uv run pytest tests/ -v

# Local server (stdio)
uv run python registry.py

# Local server (SSE)
MCP_TRANSPORT=sse uv run python registry.py

# Docker
docker build -t model-registry . && docker run -p 8000:8000 model-registry
```

## Adding a New Model

1. Add an entry to the `MODELS` dict in `models_data.py` following the existing schema (all 13 keys required: `id`, `display_name`, `provider`, `context_window`, `max_output_tokens`, `vision`, `reasoning`, `pricing_input`, `pricing_output`, `knowledge_cutoff`, `release_date`, `status`, `notes`)
2. Ensure `id` matches the dict key
3. Run tests: `uv run pytest tests/ -v`

## Adding a New Tool

1. Add a `@mcp.tool()` function in `registry.py`
2. Add corresponding tests in `tests/test_registry.py`
3. Unwrap with `.fn` for testing — FastMCP wraps tools in `FunctionTool` objects, so use `tool_name.fn` to get the raw callable

## Testing Gotcha

FastMCP `@mcp.tool()` wraps functions in `FunctionTool` objects. In tests, import the tool and access `.fn` to get the underlying function:

```python
from registry import my_tool as _my_tool
my_tool = _my_tool.fn
```

## Coding Conventions

- Python 3.12+
- Type hints on function signatures
- Docstrings on public functions
- `ruff` for formatting/linting
- Valid `status` values: `current`, `legacy`, `deprecated`
