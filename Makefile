.PHONY: help install dev build build-all clean test

# Detect Wails binary
GOPATH ?= $(shell go env GOPATH)
WAILS := $(shell command -v wails 2>/dev/null || echo $(GOPATH)/bin/wails)

help: ## Show this help message
	@echo "MailStone Verifier - Makefile"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install Wails CLI (required for building)
	@echo "Installing Wails v2..."
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "Wails installed successfully!"

deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies downloaded!"

dev: ## Run application in development mode with hot reload
	$(WAILS) dev

build: ## Build application for current platform
	@echo "Building for current platform..."
	$(WAILS) build
	@echo "Build complete! Binary is in build/bin/"

build-all: ## Build for all platforms (Mac Intel, Mac ARM, Linux, Windows)
	@echo "Building for all platforms..."
	@echo "Building macOS Intel (amd64)..."
	$(WAILS) build -platform darwin/amd64 -o mailstone-verifier-darwin-amd64
	@echo "Building macOS Apple Silicon (arm64)..."
	$(WAILS) build -platform darwin/arm64 -o mailstone-verifier-darwin-arm64
	@echo "Building Linux (amd64)..."
	$(WAILS) build -platform linux/amd64 -o mailstone-verifier-linux-amd64
	@echo "Building Windows (amd64)..."
	$(WAILS) build -platform windows/amd64 -o mailstone-verifier-windows-amd64.exe
	@echo ""
	@echo "✅ All builds complete! Binaries are in build/bin/"
	@ls -lh build/bin/

test: ## Run Go tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf build/
	rm -f coverage.out coverage.html
	@echo "Clean complete!"

run: build ## Build and run the application
	@echo "Running application..."
	./build/bin/mailstone-verifier

.DEFAULT_GOAL := help
