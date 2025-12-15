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

**Phase**: Architecture Definition ✅

- [x] Project specification
- [x] Implementation plan
- [x] Architecture definition
- [ ] Phase 0: Project setup
- [ ] Phase 1: Domain layer
- [ ] Phase 2: Application layer
- [ ] Phase 3+: Infrastructure adapters

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