.PHONY: all build test clean run lint fmt deps

# Variables
BINARY_NAME=router
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/router/main.go
GO=go
GOLINT=golangci-lint
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOFMT=gofmt

# Build the binary
all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@$(GOBUILD) -o $(BINARY_PATH) -v $(MAIN_PATH)

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html

# Run contract tests only
test-contract:
	@echo "Running contract tests..."
	@$(GOTEST) -v ./tests/contract/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -v ./tests/integration/...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@$(GOTEST) -v ./tests/unit/...

# Run performance tests
test-performance:
	@echo "Running performance tests..."
	@$(GOTEST) -v -bench=. ./tests/performance/...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf bin/ coverage.out coverage.html

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Run with specific config
run-dev:
	@$(GO) run $(MAIN_PATH) --config configs/dev.yaml

# Format code
fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .

# Run linter
lint:
	@echo "Running linter..."
	@$(GOLINT) run ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t ryohi-router:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -v $$(pwd)/configs:/app/configs ryohi-router:latest

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make test           - Run all tests"
	@echo "  make test-contract  - Run contract tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-unit      - Run unit tests"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make run            - Build and run the application"
	@echo "  make run-dev        - Run with dev config"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Install dependencies"
	@echo "  make install-tools  - Install dev tools"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run Docker container"