.PHONY: help build test lint fmt vet clean run deps tidy install-tools

# Variables
BINARY_NAME=currenseen
LAMBDA_BINARY=lambda-handler
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*' -not -path './.aws-sam/*')
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed successfully!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying go modules..."
	@go mod tidy

fmt: ## Format code with gofmt and goimports
	@echo "Formatting code..."
	@goimports -w $(GO_FILES)
	@gofmt -s -w $(GO_FILES)
	@echo "Code formatted!"

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=$(COVERAGE_FILE) ./...

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@go tool cover -func=$(COVERAGE_FILE)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race ./tests/unit/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/...

build: ## Build the Lambda binary
	@echo "Building Lambda binary..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(LAMBDA_BINARY) ./cmd/lambda
	@echo "Build complete: bin/$(LAMBDA_BINARY)"

build-local: ## Build for local development
	@echo "Building local binary..."
	@go build -o bin/$(BINARY_NAME) ./cmd/lambda
	@echo "Build complete: bin/$(BINARY_NAME)"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -rf .aws-sam/
	@echo "Clean complete!"

run: build-local ## Build and run locally
	@echo "Running locally..."
	@./bin/$(BINARY_NAME)

validate: fmt lint vet test ## Run all validation checks

sam-build: build ## Build for SAM deployment
	@echo "Building for SAM..."
	@sam build

sam-local: sam-build ## Run SAM locally
	@echo "Running SAM locally..."
	@sam local start-api

sam-deploy: sam-build ## Deploy to AWS
	@echo "Deploying to AWS..."
	@sam deploy --guided

check: validate ## Alias for validate

all: clean deps validate build ## Run everything: clean, deps, validate, and build

