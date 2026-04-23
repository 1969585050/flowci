.PHONY: help setup build run test lint clean install dev go-build go-run

# Variables
GO_CMD := go
GO_FILES := $(shell find . -name '*.go' -type f | grep -v vendor)
VERSION := 0.1.0
API_PORT := 3847

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Setup development environment
	@echo "Setting up FlowCI development environment..."
	@./scripts/setup.sh

install: ## Install dependencies
	@echo "Installing dependencies..."
	cd src && npm install
	cd go && $(GO_CMD) mod download

dev: ## Run development mode
	@echo "Starting FlowCI in development mode..."
	@echo "Note: Run 'make run-api' and 'make run-frontend' in separate terminals"

run-api: ## Run Go API server
	@echo "Starting Go API server on port $(API_PORT)..."
	cd go && $(GO_CMD) run cmd/server/main.go

run-frontend: ## Run frontend dev server
	@echo "Starting Vue frontend..."
	cd src && npm run dev

run-tauri: ## Run Tauri desktop app
	@echo "Starting Tauri desktop app..."
	cd src-tauri && cargo tauri dev

build: ## Build all components
	@echo "Building FlowCI..."
	$(MAKE) build-frontend
	$(MAKE) build-tauri

build-frontend: ## Build frontend
	@echo "Building frontend..."
	cd src && npm run build

build-tauri: ## Build Tauri app
	@echo "Building Tauri app..."
	cd src-tauri && cargo tauri build

go-build: ## Build Go API server
	@echo "Building Go API server..."
	cd go && $(GO_CMD) build -o bin/flowci-api cmd/server/main.go

test: ## Run all tests
	@echo "Running tests..."
	$(MAKE) test-go
	$(MAKE) test-frontend

test-go: ## Run Go tests
	@echo "Running Go tests..."
	cd go && $(GO_CMD) test -v -cover ./...

test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	cd src && npm run test:unit

test-e2e: ## Run e2e tests
	@echo "Running e2e tests..."
	cd src && npm run test:e2e

lint: ## Run all linters
	@echo "Running linters..."
	$(MAKE) lint-go
	$(MAKE) lint-frontend

lint-go: ## Run Go linter
	@echo "Running Go linter..."
	cd go && $(GO_CMD) fmt ./...
	cd go && golangci-lint run || true

lint-frontend: ## Run frontend linter
	@echo "Running frontend linter..."
	cd src && npm run lint
	cd src && npm run typecheck

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf src/dist
	rm -rf src-tauri/target
	rm -rf go/bin
	rm -rf *.log
	cd src && rm -rf node_modules/.vite

fmt: ## Format code
	@echo "Formatting code..."
	cd go && $(GO_CMD) fmt ./...
	cd src && npm run format

ci: ## Run CI pipeline
	@echo "Running CI pipeline..."
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) build

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t flowci/flowci:$(VERSION) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p $(API_PORT):$(API_PORT) flowci/flowci:$(VERSION)

# Release
release: ## Create release
	@echo "Creating release v$(VERSION)..."
	$(MAKE) test
	$(MAKE) build
	$(MAKE) docker-build
