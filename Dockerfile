FROM python:3.12-slim

COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

WORKDIR /app

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev

COPY models_data.py registry.py ./

ENV MCP_TRANSPORT=sse
ENV PORT=8000

CMD ["uv", "run", "registry.py"]
