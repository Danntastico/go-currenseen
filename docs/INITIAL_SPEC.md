# Currency Exchange Rate Service - Project Specification

## Project Overview

I am learning Golang, Backend development, and Cloud Computing with AWS. My goal is to learn by making Golang projects that implement many concepts in software development, backend development, and cloud computing.

These concepts are focused on the **Fintech industry**.

## Project Goal

Create a **Currency Exchange Rate Service** - a small service that fetches and caches exchange rates with high availability and resilience.

## Technical Requirements

### Core Requirements

1. **Exchange Rate Fetching**
   - Fetch real-time exchange rates from external API(s)
   - Support multiple currency pairs
   - Handle base currency conversion (e.g., USD, EUR, GBP)

2. **Caching Strategy**
   - Cache exchange rates in DynamoDB
   - Implement TTL (Time To Live) for cached data
   - Cache invalidation and refresh strategies

3. **API Endpoints**
   - `GET /rates/{base}/{target}` - Get exchange rate for specific pair
   - `GET /rates/{base}` - Get all rates for a base currency
   - `GET /health` - Health check endpoint
   - `GET /metrics` - Service metrics (optional, for learning)

4. **Resilience & Reliability**
   - Circuit breaker pattern for external API calls
   - Cache fallback when external API is unavailable
   - Graceful degradation
   - Retry logic with exponential backoff
   - Rate limiting awareness

5. **Error Handling**
   - Structured error responses
   - Error logging and monitoring
   - User-friendly error messages

### Architecture Requirements

**Architecture Pattern**: Hexagonal Architecture (Ports & Adapters)

**Layers**:
- **Domain Layer**: Core business logic, entities, use cases
- **Application Layer**: Use case orchestration, DTOs
- **Infrastructure Layer**: External adapters (API clients, DynamoDB, Lambda handlers)
- **Ports**: Interfaces defining contracts (Repository, ExchangeRateProvider, Cache)

**Key Principles**:
- Dependency inversion (depend on abstractions)
- Clear boundaries between layers
- Testability through interfaces
- Loose coupling

### AWS Infrastructure Requirements

1. **AWS Lambda**
   - Serverless function handlers
   - Environment variable configuration
   - Cold start optimization
   - Proper timeout and memory configuration

2. **API Gateway**
   - REST API configuration
   - Request/response transformation
   - CORS configuration
   - API versioning (for learning)

3. **DynamoDB**
   - Table design for caching exchange rates
   - Partition key strategy
   - TTL attribute configuration
   - Read/write capacity considerations

4. **Additional AWS Services** (Learning Opportunities)
   - CloudWatch Logs (logging)
   - CloudWatch Metrics (monitoring)
   - IAM Roles & Policies (security)
   - AWS SAM or CDK (Infrastructure as Code)

### Design Patterns

1. **Adapter Pattern**
   - Adapt external exchange rate APIs to internal interface
   - Multiple provider support (e.g., ExchangeRate-API, Fixer.io, etc.)

2. **Strategy Pattern**
   - Different strategies for fetching rates (primary provider, fallback provider)
   - Different caching strategies

3. **Repository Pattern**
   - Abstract data access layer
   - DynamoDB implementation behind interface

4. **Factory Pattern**
   - Create appropriate adapters based on configuration

5. **Circuit Breaker Pattern**
   - Prevent cascading failures
   - Open/Closed/Half-Open states
   - Automatic recovery

### Use Cases

#### UC1: Get Exchange Rate for Currency Pair
**Actor**: API Client  
**Preconditions**: Service is running, cache may or may not have data  
**Main Flow**:
1. Client requests rate for `{base}/{target}` (e.g., USD/EUR)
2. System checks cache (DynamoDB)
3. If cache hit and not expired → return cached rate
4. If cache miss or expired → fetch from external API
5. Update cache with new rate
6. Return rate to client

**Alternative Flow**:
- External API unavailable → return cached rate (even if expired)
- Cache unavailable → fetch from external API directly
- Both unavailable → return error with last known rate timestamp

#### UC2: Get All Rates for Base Currency
**Actor**: API Client  
**Preconditions**: Service is running  
**Main Flow**:
1. Client requests all rates for base currency (e.g., USD)
2. System checks cache for base currency
3. If cache hit → return all cached rates
4. If cache miss → fetch from external API
5. Cache all rates
6. Return rates to client

#### UC3: Health Check
**Actor**: Monitoring System / API Client  
**Main Flow**:
1. Client requests `/health`
2. System checks:
   - Lambda function status
   - DynamoDB connectivity
   - External API connectivity (optional)
3. Return health status

#### UC4: Cache Refresh (Background/On-Demand)
**Actor**: System / Admin  
**Main Flow**:
1. System detects cache expiration approaching
2. Pre-fetch rates from external API
3. Update cache before expiration
4. Ensure zero downtime for clients

### Learning Targets

#### Golang Concepts
- ✅ Interfaces and polymorphism
- ✅ Error handling best practices
- ✅ Context usage for cancellation/timeouts
- ✅ Goroutines and channels (for concurrent cache refresh)
- ✅ Struct embedding and composition
- ✅ Package organization and module management
- ✅ Testing (unit, integration, table-driven tests)
- ✅ JSON marshaling/unmarshaling
- ✅ HTTP client/server patterns
- ✅ Dependency injection

#### Backend Development Concepts
- ✅ RESTful API design
- ✅ Request validation
- ✅ Response formatting
- ✅ Middleware patterns
- ✅ Configuration management
- ✅ Logging and observability
- ✅ Graceful shutdown
- ✅ Concurrency patterns

#### AWS & Cloud Concepts
- ✅ Serverless architecture
- ✅ Lambda function development
- ✅ API Gateway integration
- ✅ DynamoDB operations (GetItem, PutItem, Query)
- ✅ IAM roles and permissions
- ✅ CloudWatch integration
- ✅ Infrastructure as Code (SAM/CDK)
- ✅ Environment configuration
- ✅ Cold start optimization

#### Software Engineering Concepts
- ✅ Clean Architecture / Hexagonal Architecture
- ✅ SOLID principles
- ✅ Design patterns implementation
- ✅ Test-driven development (TDD)
- ✅ Dependency injection
- ✅ Configuration management
- ✅ Error handling strategies
- ✅ Resilience patterns

### Resilience Question & Solutions

**Question**: "If the external API is slow or down?"

**Solutions**:

1. **Circuit Breaker**
   - Monitor external API health
   - Open circuit after threshold failures
   - Prevent unnecessary calls when API is down
   - Half-open state for recovery testing
   - Automatic circuit closure after cooldown

2. **Cache Fallback**
   - Always check cache first
   - Return stale cache if external API unavailable
   - Include cache timestamp in response
   - Graceful degradation messaging

3. **Retry Logic**
   - Exponential backoff for transient failures
   - Maximum retry attempts
   - Context-based timeout

4. **Multiple Providers**
   - Primary and fallback exchange rate providers
   - Automatic failover
   - Provider health tracking

5. **Timeout Management**
   - Short timeouts for external calls
   - Prevent Lambda timeout
   - Fast failure for better UX

### Project Structure

```
go-currenseen/
├── cmd/
│   └── lambda/
│       └── main.go              # Lambda entry point
├── internal/
│   ├── domain/
│   │   ├── entity/              # Domain entities
│   │   ├── repository/          # Repository interfaces (ports)
│   │   └── service/             # Domain services
│   ├── application/
│   │   ├── usecase/             # Use case implementations
│   │   └── dto/                 # Data transfer objects
│   └── infrastructure/
│       ├── adapter/             # External adapters
│       │   ├── api/             # Exchange rate API clients
│       │   ├── dynamodb/        # DynamoDB repository implementation
│       │   └── lambda/          # Lambda handlers
│       ├── config/              # Configuration
│       └── middleware/          # HTTP middleware
├── pkg/
│   ├── circuitbreaker/          # Circuit breaker implementation
│   ├── cache/                   # Cache utilities
│   └── logger/                  # Logging utilities
├── tests/
│   ├── unit/                    # Unit tests
│   └── integration/             # Integration tests
├── docs/
│   └── INITIAL_SPEC.md          # This file
├── infrastructure/
│   └── sam.yaml                 # AWS SAM template
├── go.mod
├── go.sum
└── README.md
```

### Development Methodology

**Spec-Driven Development (SDD)**:
1. Define use cases and requirements
2. Design architecture and interfaces
3. Write tests first (TDD)
4. Implement features incrementally
5. Refactor and improve
6. Document learnings

### Success Criteria

- ✅ Service handles external API failures gracefully
- ✅ Cache reduces external API calls by >80%
- ✅ Response time <200ms for cached requests
- ✅ 99%+ uptime (with cache fallback)
- ✅ All use cases implemented and tested
- ✅ Clean architecture boundaries respected
- ✅ Comprehensive test coverage (>80%)
- ✅ Deployed and accessible via API Gateway
- ✅ Proper error handling and logging
- ✅ Documentation complete

### Next Steps

1. Set up Go module and project structure
2. Define domain entities and interfaces
3. Implement repository interfaces (ports)
4. Create DynamoDB adapter (infrastructure)
5. Implement exchange rate API adapters
6. Build use cases
7. Create Lambda handlers
8. Implement circuit breaker
9. Add tests
10. Deploy to AWS

---

**Note**: This project is designed for learning. Each component should be implemented with educational value in mind, focusing on understanding concepts rather than just making it work.
