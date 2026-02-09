.PHONY: test lint format serve docker-build docker-run clean help validate-models

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run test suite
	uv run pytest tests/ -v

lint: ## Run ruff linter
	uv run ruff check .

format: ## Format code with ruff
	uv run ruff format .

check: lint test ## Run linter and tests

serve: ## Run MCP server locally (stdio)
	uv run python registry.py

serve-sse: ## Run MCP server with SSE transport
	MCP_TRANSPORT=sse uv run python registry.py

docker-build: ## Build Docker image
	docker build -t universal-model-registry .

docker-run: ## Run Docker container
	docker run -p 8000:8000 universal-model-registry

clean: ## Remove caches and build artifacts
	rm -rf __pycache__ .pytest_cache .ruff_cache
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true

validate-models: ## Validate model IDs against live APIs
	uv run python scripts/validate_models.py

sync: ## Install/update dependencies
	uv sync
