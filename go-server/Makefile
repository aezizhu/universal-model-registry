.PHONY: build test lint run run-sse clean docker-build docker-run check help

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	go build -o bin/server ./cmd/server

test: ## Run tests
	go test ./... -v

lint: ## Run golangci-lint
	golangci-lint run ./...

run: ## Run server (stdio)
	go run ./cmd/server

run-sse: ## Run server (SSE transport)
	MCP_TRANSPORT=sse go run ./cmd/server

clean: ## Remove build artifacts
	rm -rf bin/

docker-build: ## Build Docker image
	docker build -t model-id-cheatsheet .

docker-run: ## Run Docker container
	docker run -p 8000:8000 model-id-cheatsheet

check: lint test ## Run lint and tests
