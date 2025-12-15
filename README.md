# go-currenseen

A Currency Exchange Rate Service built with Go, implementing Hexagonal Architecture and deployed on AWS Lambda.

## 📚 Documentation

- **[Architecture Definition](docs/ARCHITECTURE.md)** - Complete architecture documentation with diagrams
- **[Project Specification](docs/INITIAL_SPEC.md)** - Detailed project requirements and use cases
- **[Implementation Plan](docs/IMPLEMENTATION_PLAN.md)** - Step-by-step implementation phases

## 🎯 Project Overview

A serverless microservice that provides real-time and cached exchange rates through a REST API. The service implements resilience patterns (circuit breaker, cache fallback) to ensure high availability.

## 🏗️ Architecture

**Hexagonal Architecture (Ports & Adapters)**
- **Domain Layer**: Core business logic, entities, interfaces (ports)
- **Application Layer**: Use cases, DTOs, orchestration
- **Infrastructure Layer**: Adapters (DynamoDB, HTTP clients, Lambda handlers)

See [ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture diagrams and component interactions.

## 🚀 Key Features

- **Exchange Rate API**: Fetch rates for currency pairs
- **Caching**: DynamoDB-based caching with TTL
- **Resilience**: Circuit breaker pattern for external APIs
- **Cache Fallback**: Graceful degradation when external APIs fail
- **Security**: API key authentication, input validation, rate limiting

## 📖 Learning Objectives

### Golang Concepts
- Interfaces and polymorphism
- Error handling patterns
- Context usage
- Concurrency (goroutines, channels)
- Testing (unit, integration, table-driven)

### Backend Development
- RESTful API design
- Middleware patterns
- Configuration management
- Observability

### AWS & Cloud
- Serverless architecture (Lambda)
- API Gateway
- DynamoDB (operations, TTL, GSI)
- IAM roles & policies
- CloudWatch
- Infrastructure as Code (SAM)

### Software Engineering
- Clean/Hexagonal Architecture
- SOLID principles
- Design patterns (Repository, Adapter, Strategy, Circuit Breaker)
- TDD approach

## 📋 Current Status

**Phase**: Phase 0 - Project Setup ✅

- [x] Project specification
- [x] Implementation plan
- [x] Architecture definition
- [x] Phase 0: Project setup
- [ ] Phase 1: Domain layer
- [ ] Phase 2: Application layer
- [ ] Phase 3+: Infrastructure adapters

## 🚀 Getting Started

### Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **AWS CLI**: [Install AWS CLI](https://aws.amazon.com/cli/)
- **AWS SAM CLI**: [Install SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
- **Make**: Usually pre-installed on Linux/macOS
- **Docker**: Required for local SAM testing

### Setup Instructions

1. **Clone the repository** (if applicable):
   ```bash
   git clone <repository-url>
   cd go-currenseen
   ```

2. **Install development tools**:
   ```bash
   make install-tools
   ```
   This installs:
   - `golangci-lint` - Go linter
   - `goimports` - Code formatter

3. **Download dependencies**:
   ```bash
   make deps
   # or
   go mod download
   ```

4. **Verify setup**:
   ```bash
   make validate
   ```
   This runs formatting, linting, vetting, and tests.

### Development Workflow

#### Common Make Commands

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Run tests with coverage
make test-coverage

# Build Lambda binary
make build

# Run all validation checks
make validate

# Clean build artifacts
make clean

# Run everything (clean, deps, validate, build)
make all
```

#### Project Structure

```
go-currenseen/
├── cmd/
│   └── lambda/              # Lambda entry point (main.go)
├── internal/
│   ├── domain/              # Domain layer (no external dependencies)
│   │   ├── entity/         # Domain entities
│   │   ├── repository/     # Repository interfaces (ports)
│   │   ├── provider/       # Provider interfaces (ports)
│   │   ├── service/        # Domain services
│   │   └── errors.go       # Domain error types
│   ├── application/         # Application layer
│   │   ├── usecase/        # Use cases
│   │   └── dto/            # Data Transfer Objects
│   └── infrastructure/     # Infrastructure layer (adapters)
│       ├── adapter/        # External adapters (DynamoDB, HTTP, Lambda)
│       ├── middleware/     # HTTP middleware
│       └── config/         # Configuration management
├── pkg/                     # Shared packages
│   ├── circuitbreaker/     # Circuit breaker implementation
│   └── logger/             # Structured logging
├── tests/                   # Test files
│   ├── unit/               # Unit tests
│   └── integration/        # Integration tests
├── infrastructure/          # AWS SAM templates
│   ├── template.yaml       # SAM template
│   └── samconfig.toml      # SAM configuration
├── docs/                    # Documentation
├── Makefile                 # Development tasks
├── .golangci.yml           # Linter configuration
├── .gitignore              # Git ignore rules
└── go.mod                  # Go module definition
```

### Local Development

#### Running Locally

```bash
# Build for local development
make build-local

# Run locally
make run
```

#### AWS SAM Local Testing

```bash
# Build for SAM
make sam-build

# Run SAM locally (starts API Gateway locally)
make sam-local
```

### Code Quality

The project uses:
- **golangci-lint**: Comprehensive Go linter (see `.golangci.yml`)
- **goimports**: Automatic import organization
- **go vet**: Static analysis
- **go test**: Testing framework

Run all quality checks:
```bash
make validate
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Generate coverage report
make test-coverage
```

### AWS Deployment

```bash
# Build and deploy
make sam-deploy

# Or manually
sam build
sam deploy --guided
```

### Environment Variables

Create a `.env` file for local development (not committed to git):

```bash
TABLE_NAME=ExchangeRates
LOG_LEVEL=DEBUG
AWS_REGION=us-east-1
```

### Troubleshooting

**Issue**: `golangci-lint: command not found`
- Solution: Run `make install-tools`

**Issue**: SAM build fails
- Solution: Ensure Docker is running and AWS credentials are configured

**Issue**: Go module errors
- Solution: Run `go mod tidy` or `make tidy`

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **Architecture**: Hexagonal Architecture
- **AWS Services**: Lambda, API Gateway, DynamoDB, CloudWatch, Secrets Manager
- **Infrastructure**: AWS SAM (Infrastructure as Code)
- **Patterns**: Repository, Adapter, Strategy, Circuit Breaker, Factory

## 📁 Project Structure

```
go-currenseen/
├── cmd/lambda/              # Lambda entry point
├── internal/
│   ├── domain/              # Domain layer (entities, ports)
│   ├── application/         # Application layer (use cases, DTOs)
│   └── infrastructure/      # Infrastructure layer (adapters)
├── pkg/                     # Shared packages (circuit breaker, logger)
├── tests/                   # Unit and integration tests
├── infrastructure/          # AWS SAM template
└── docs/                    # Documentation
```

## 🎓 Learning Tasks

1. **Hexagonal Architecture** - Implementing ports and adapters
2. **Circuit Breaker** - Resilience pattern implementation
3. **AWS Lambda** - Serverless function development
4. **AWS SAM** - Infrastructure as Code
5. **DynamoDB** - NoSQL database operations
6. **API Gateway** - REST API configuration
7. **CloudWatch** - Logging and monitoring
8. **IAM** - Security and permissions

## 📝 Next Steps

1. Review [ARCHITECTURE.md](docs/ARCHITECTURE.md) for system design
2. Follow [IMPLEMENTATION_PLAN.md](docs/IMPLEMENTATION_PLAN.md) for step-by-step implementation
3. Start with Phase 0: Project Setup

---

**Note**: This project is designed for learning. Each component focuses on understanding concepts rather than just making it work.