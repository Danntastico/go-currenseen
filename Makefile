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

test-integration: ## Run integration tests only (requires INTEGRATION_TESTS=true and -tags=integration)
	@echo "Running integration tests..."
	@echo "Note: Set INTEGRATION_TESTS=true environment variable"
	@go test -tags=integration -v -race ./tests/integration/...

build: ## Build the Lambda binary
	@echo "Building Lambda binary..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(LAMBDA_BINARY) ./cmd/lambda
	@echo "Build complete: bin/$(LAMBDA_BINARY)"

build-local: ## Build for local development
	@echo "Building local binary..."
	@go build -o bin/$(BINARY_NAME) ./cmd/lambda
	@echo "Build complete: bin/$(BINARY_NAME)"

build-local-server: ## Build local HTTP server for testing
	@echo "Building local HTTP server..."
	@go build -o bin/local-server ./cmd/local
	@echo "Build complete: bin/local-server"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -rf .aws-sam/
	@echo "Clean complete!"

run: build-local ## Build and run locally (Lambda mode - requires SAM or Lambda runtime)
	@echo "Running locally..."
	@echo "Note: This requires Lambda runtime. Use 'make run-local-server' for HTTP server."
	@./bin/$(BINARY_NAME)

run-local-server: build-local-server ## Build and run local HTTP server
	@echo "Running local HTTP server..."
	@echo "Server will start on http://localhost:8080 (or PORT env var)"
	@echo "Set required environment variables (TABLE_NAME, etc.) before running"
	@./bin/local-server

validate: fmt lint vet test ## Run all validation checks

sam-build: ## Build for SAM deployment
	@echo "Building for SAM..."
	@echo "Building Go Lambda function..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/lambda
	@chmod +x bootstrap
	@echo "Running SAM build..."
	@sam build
	@echo "SAM build complete!"

sam-validate: ## Validate SAM template
	@echo "Validating SAM template..."
	@sam validate
	@echo "Template validation complete!"

sam-local: sam-build ## Run SAM locally
	@echo "Running SAM locally..."
	@echo "API will be available at http://localhost:3000"
	@sam local start-api --port 3000

sam-deploy-dev: sam-build sam-validate ## Deploy to dev environment
	@echo "Deploying to dev environment..."
	@sam deploy \
		--stack-name currenseen-dev \
		--parameter-overrides Environment=dev \
		--capabilities CAPABILITY_IAM \
		--region us-east-1 \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset

sam-deploy-staging: sam-build sam-validate ## Deploy to staging environment
	@echo "Deploying to staging environment..."
	@sam deploy \
		--stack-name currenseen-staging \
		--parameter-overrides Environment=staging \
		--capabilities CAPABILITY_IAM \
		--region us-east-1 \
		--confirm-changeset

sam-deploy-prod: sam-build sam-validate ## Deploy to production environment
	@echo "Deploying to production environment..."
	@echo "WARNING: This will deploy to PRODUCTION!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		sam deploy \
			--stack-name currenseen-prod \
			--parameter-overrides Environment=prod \
			--capabilities CAPABILITY_IAM \
			--region us-east-1 \
			--confirm-changeset; \
	fi

sam-deploy: sam-build sam-validate ## Deploy to AWS (guided mode)
	@echo "Deploying to AWS (guided mode)..."
	@sam deploy --guided

sam-logs: ## Tail Lambda function logs
	@echo "Tailing Lambda function logs..."
	@sam logs -n ExchangeRateFunction --stack-name currenseen-dev --tail

sam-delete: ## Delete SAM stack
	@echo "Deleting SAM stack..."
	@read -p "Enter stack name (default: currenseen-dev): " stack_name; \
	stack_name=$${stack_name:-currenseen-dev}; \
	sam delete --stack-name $$stack_name

check: validate ## Alias for validate

all: clean deps validate build ## Run everything: clean, deps, validate, and build

