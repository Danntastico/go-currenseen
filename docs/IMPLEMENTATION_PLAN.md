# Currency Exchange Rate Service - Implementation Plan

## Overview

This document outlines the step-by-step implementation plan for the Currency Exchange Rate Service following **Spec-Driven Development (SDD)** methodology. Each phase builds upon the previous one, ensuring clean architecture boundaries and comprehensive learning.

## Implementation Phases

### Phase 0: Project Setup & Foundation
**Goal**: Establish project structure and development environment

**Tasks**:
1. Initialize Go module (`go mod init`)
2. Create project directory structure according to spec
3. Set up development tools:
   - `golangci-lint` configuration
   - `goimports` setup
   - Git hooks for pre-commit checks
4. Create `README.md` with setup instructions
5. Set up `.gitignore` for Go projects
6. Initialize AWS SAM template structure

**Deliverables**:
- ✅ Project structure created
- ✅ Go module initialized
- ✅ Linting and formatting configured
- ✅ Basic README with setup instructions

**Learning Objectives**:
- Go module management
- Project structure best practices
- Development tooling setup

**Estimated Time**: 1-2 hours

---

### Phase 1: Domain Layer - Core Entities & Interfaces
**Goal**: Define domain entities and port interfaces (no external dependencies)

**Tasks**:
1. **Create domain entities** (`internal/domain/entity/`):
   - `exchange_rate.go`: ExchangeRate struct with validation
   - `currency_code.go`: CurrencyCode type with validation
   - Domain error types (`errors.go`)

2. **Define repository interfaces** (`internal/domain/repository/`):
   - `exchange_rate_repository.go`: Interface for data access
   - Methods: `Get`, `Save`, `GetByBase`, `Delete`

3. **Define provider interfaces** (`internal/domain/provider/`):
   - `exchange_rate_provider.go`: Interface for external API
   - Methods: `FetchRate`, `FetchAllRates`

4. **Create domain services** (`internal/domain/service/`):
   - `validation_service.go`: Currency code validation
   - `rate_calculator.go`: Rate calculation utilities

5. **Write unit tests** for all domain entities and services

**Deliverables**:
- ✅ Domain entities with validation
- ✅ Repository interface (port)
- ✅ Exchange rate provider interface (port)
- ✅ Domain error types
- ✅ Unit tests with >90% coverage

**Learning Objectives**:
- Domain-driven design
- Interface design in Go
- Type safety and validation
- Error handling patterns

**Dependencies**: None

**Estimated Time**: 4-6 hours

---

### Phase 2: Application Layer - Use Cases & DTOs
**Goal**: Implement use case orchestration and data transfer objects

**Tasks**:
1. **Create DTOs** (`internal/application/dto/`):
   - `request.go`: Request DTOs (GetRateRequest, GetRatesRequest)
   - `response.go`: Response DTOs (RateResponse, RatesResponse)
   - `mapper.go`: Domain entity ↔ DTO conversion

2. **Implement Use Cases** (`internal/application/usecase/`):
   - `get_exchange_rate.go`: UC1 - Get rate for currency pair
   - `get_all_rates.go`: UC2 - Get all rates for base currency
   - `health_check.go`: UC3 - Health check use case

3. **Create use case constructors** with dependency injection

4. **Write unit tests** for use cases (mock dependencies)

**Deliverables**:
- ✅ All use cases implemented
- ✅ DTOs and mappers
- ✅ Unit tests with mocked dependencies
- ✅ Dependency injection pattern

**Learning Objectives**:
- Use case pattern
- DTO pattern
- Dependency injection
- Mocking in tests

**Dependencies**: Phase 1

**Estimated Time**: 6-8 hours

---

### Phase 3: Infrastructure Layer - DynamoDB Adapter
**Goal**: Implement DynamoDB repository adapter

**Tasks**:
1. **Set up AWS SDK v2** dependencies
2. **Create DynamoDB repository** (`internal/infrastructure/adapter/dynamodb/`):
   - `exchange_rate_repository.go`: Implement domain repository interface
   - Handle DynamoDB operations (GetItem, PutItem, Query)
   - Implement TTL logic
   - Error mapping (DynamoDB errors → domain errors)

3. **Create DynamoDB table schema**:
   - Partition key: `PK` (e.g., `RATE#USD#EUR`)
   - Sort key: `SK` (optional, for future expansion)
   - TTL attribute: `ttl`
   - GSI for querying by base currency

4. **Write integration tests**:
   - Use DynamoDB Local or test table
   - Test CRUD operations
   - Test TTL expiration

5. **Create configuration** (`internal/infrastructure/config/`):
   - `dynamodb.go`: DynamoDB client configuration

**Deliverables**:
- ✅ DynamoDB repository implementation
- ✅ Table schema design
- ✅ Integration tests
- ✅ Configuration management

**Learning Objectives**:
- AWS SDK v2 usage
- DynamoDB operations
- Repository pattern implementation
- Integration testing
- TTL management

**Dependencies**: Phase 1

**Estimated Time**: 8-10 hours

---

### Phase 4: Infrastructure Layer - External API Adapter
**Goal**: Implement external exchange rate API adapter

**Tasks**:
1. **Choose exchange rate provider** (e.g., ExchangeRate-API, Fixer.io, or free tier)
2. **Create HTTP client** (`internal/infrastructure/adapter/api/`):
   - `http_client.go`: Secure HTTP client with timeout
   - `exchange_rate_provider.go`: Implement domain provider interface
   - Response parsing and validation
   - Error handling

3. **Implement retry logic**:
   - Exponential backoff
   - Maximum retry attempts
   - Context-based cancellation

4. **Create provider factory**:
   - `provider_factory.go`: Create provider based on config
   - Support multiple providers (primary/fallback)

5. **Write unit tests**:
   - Mock HTTP responses
   - Test retry logic
   - Test error handling

**Deliverables**:
- ✅ External API adapter
- ✅ Retry logic implementation
- ✅ Provider factory
- ✅ Unit tests with mocked HTTP

**Learning Objectives**:
- HTTP client patterns
- Retry strategies
- External API integration
- Factory pattern
- Response validation

**Dependencies**: Phase 1

**Estimated Time**: 6-8 hours

---

### Phase 5: Circuit Breaker Implementation
**Goal**: Implement circuit breaker pattern for resilience

**Tasks**:
1. **Create circuit breaker package** (`pkg/circuitbreaker/`):
   - `circuit_breaker.go`: Core circuit breaker logic
   - States: Closed, Open, HalfOpen
   - Thread-safe implementation (sync.RWMutex)
   - Failure counting and cooldown logic

2. **Integrate with API provider**:
   - Wrap external API calls with circuit breaker
   - Record successes/failures
   - Handle circuit states

3. **Add configuration**:
   - Failure threshold
   - Cooldown duration
   - Half-open test interval

4. **Write comprehensive tests**:
   - Test state transitions
   - Test failure counting
   - Test cooldown logic
   - Test thread safety

**Deliverables**:
- ✅ Circuit breaker implementation
- ✅ Integration with API provider
- ✅ Configuration options
- ✅ Comprehensive tests

**Learning Objectives**:
- Circuit breaker pattern
- Concurrency patterns (mutexes)
- State machine implementation
- Resilience patterns

**Dependencies**: Phase 4

**Estimated Time**: 6-8 hours

---

### Phase 6: Cache Strategy & Fallback Logic
**Goal**: Implement caching strategy with fallback

**Tasks**:
1. **Enhance use cases** with cache-first logic:
   - Check cache before external API
   - Return stale cache if API unavailable
   - Update cache after successful fetch

2. **Implement cache expiration logic**:
   - Check TTL before returning cached data
   - Handle expired cache gracefully

3. **Add fallback strategy**:
   - Return stale cache when circuit breaker is open
   - Include cache timestamp in response
   - Log fallback usage

4. **Update use case tests**:
   - Test cache hit scenarios
   - Test cache miss scenarios
   - Test fallback scenarios

**Deliverables**:
- ✅ Cache-first logic in use cases
- ✅ Fallback strategy
- ✅ Updated tests

**Learning Objectives**:
- Caching strategies
- Fallback patterns
- Graceful degradation

**Dependencies**: Phase 2, Phase 3, Phase 5

**Estimated Time**: 4-6 hours

---

### Phase 7: Lambda Handlers & API Gateway Integration
**Goal**: Create Lambda handlers and API Gateway integration

**Tasks**:
1. **Create Lambda handlers** (`internal/infrastructure/adapter/lambda/`):
   - `get_rate_handler.go`: Handle GET /rates/{base}/{target}
   - `get_rates_handler.go`: Handle GET /rates/{base}
   - `health_handler.go`: Handle GET /health
   - Request parsing and validation
   - Response formatting

2. **Create middleware** (`internal/infrastructure/middleware/`):
   - `validation_middleware.go`: Input validation
   - `error_middleware.go`: Error handling
   - `logging_middleware.go`: Request logging

3. **Create Lambda entry point** (`cmd/lambda/main.go`):
   - Handler routing
   - Dependency initialization
   - Configuration loading

4. **Set up API Gateway**:
   - Define API routes
   - Configure request/response transformation
   - Set up CORS (if needed)

5. **Write integration tests**:
   - Test handlers with mocked dependencies
   - Test error scenarios
   - Test validation

**Deliverables**:
- ✅ Lambda handlers for all endpoints
- ✅ Middleware implementation
- ✅ API Gateway configuration
- ✅ Handler tests

**Learning Objectives**:
- AWS Lambda development
- API Gateway integration
- Middleware patterns
- Request/response handling

**Dependencies**: Phase 2, Phase 3, Phase 4

**Estimated Time**: 8-10 hours

---

### Phase 8: Configuration & Environment Management
**Goal**: Implement configuration management

**Tasks**:
1. **Create configuration package** (`internal/infrastructure/config/`):
   - `config.go`: Main configuration struct
   - Environment variable loading
   - Configuration validation
   - Default values

2. **Add configuration for**:
   - DynamoDB table name
   - External API keys/URLs
   - Circuit breaker settings
   - Cache TTL
   - Lambda settings

3. **Integrate AWS Secrets Manager**:
   - Load API keys from Secrets Manager
   - Cache secrets (with TTL)
   - Handle secret rotation

4. **Write configuration tests**

**Deliverables**:
- ✅ Configuration management
- ✅ Secrets Manager integration
- ✅ Environment variable support

**Learning Objectives**:
- Configuration patterns
- AWS Secrets Manager
- Environment management

**Dependencies**: Phase 3, Phase 4

**Estimated Time**: 4-6 hours

---

### Phase 9: Logging & Observability
**Goal**: Implement structured logging and monitoring

**Tasks**:
1. **Create logger package** (`pkg/logger/`):
   - Structured logging setup
   - Log levels (DEBUG, INFO, WARN, ERROR)
   - Context propagation
   - Log sanitization (mask sensitive data)

2. **Add logging throughout**:
   - Request/response logging
   - Error logging with context
   - Performance logging
   - Circuit breaker state changes

3. **Set up CloudWatch integration**:
   - Log groups configuration
   - Custom metrics
   - Alarms setup

4. **Add request ID tracking**:
   - Generate request IDs
   - Propagate through context
   - Include in all logs

**Deliverables**:
- ✅ Structured logging
- ✅ CloudWatch integration
- ✅ Request tracking

**Learning Objectives**:
- Structured logging
- Observability patterns
- CloudWatch integration
- Context propagation

**Dependencies**: Phase 7

**Estimated Time**: 4-6 hours

---

### Phase 10: Security Implementation
**Goal**: Implement security measures

**Tasks**:
1. **Input validation**:
   - Currency code validation middleware
   - Request size limits
   - Path parameter sanitization

2. **API Key authentication**:
   - API key validation middleware
   - Secrets Manager integration
   - Constant-time comparison

3. **Rate limiting**:
   - Per-API-key rate limiting
   - Integration with API Gateway throttling
   - Rate limit headers

4. **Security headers**:
   - HSTS
   - X-Content-Type-Options
   - X-Frame-Options
   - CSP

5. **Security testing**:
   - Test input validation
   - Test authentication
   - Test rate limiting
   - Test injection attempts

**Deliverables**:
- ✅ Input validation
- ✅ API key authentication
- ✅ Rate limiting
- ✅ Security headers
- ✅ Security tests

**Learning Objectives**:
- Web security best practices
- Authentication patterns
- Rate limiting
- Security testing

**Dependencies**: Phase 7, Phase 8

**Estimated Time**: 6-8 hours

---

### Phase 11: AWS Infrastructure as Code
**Goal**: Define AWS infrastructure using SAM/CDK

**Tasks**:
1. **Create SAM template** (`infrastructure/sam.yaml`):
   - Lambda function definition
   - API Gateway configuration
   - DynamoDB table
   - IAM roles and policies
   - Environment variables
   - CloudWatch log groups

2. **Configure IAM permissions**:
   - Least privilege principle
   - DynamoDB access
   - Secrets Manager access
   - CloudWatch logging

3. **Set up deployment scripts**:
   - Build script
   - Deploy script
   - Environment-specific configs

4. **Document deployment process**

**Deliverables**:
- ✅ SAM template
- ✅ IAM roles configured
- ✅ Deployment scripts
- ✅ Deployment documentation

**Learning Objectives**:
- Infrastructure as Code
- AWS SAM
- IAM best practices
- Deployment automation

**Dependencies**: All previous phases

**Estimated Time**: 6-8 hours

---

### Phase 12: Testing & Quality Assurance
**Goal**: Comprehensive testing and quality checks

**Tasks**:
1. **Unit test coverage**:
   - Achieve >80% coverage
   - Test all error paths
   - Test edge cases

2. **Integration tests**:
   - End-to-end handler tests
   - DynamoDB integration tests
   - External API integration tests (mocked)

3. **Performance testing**:
   - Load testing
   - Cold start measurement
   - Cache performance

4. **Security testing**:
   - Dependency scanning
   - Security vulnerability scanning
   - Penetration testing basics

5. **Code quality**:
   - Run linters
   - Fix all warnings
   - Code review checklist

**Deliverables**:
- ✅ >80% test coverage
- ✅ Integration tests
- ✅ Performance benchmarks
- ✅ Security scan results
- ✅ Quality metrics

**Learning Objectives**:
- Testing strategies
- Performance optimization
- Security testing
- Code quality tools

**Dependencies**: All previous phases

**Estimated Time**: 8-10 hours

---

### Phase 13: Documentation & Deployment
**Goal**: Complete documentation and deploy to AWS

**Tasks**:
1. **Update README.md**:
   - Project overview
   - Architecture diagram
   - Setup instructions
   - API documentation
   - Deployment guide

2. **Create API documentation**:
   - Endpoint descriptions
   - Request/response examples
   - Error codes
   - Rate limits

3. **Deploy to AWS**:
   - Deploy to dev environment
   - Test in AWS environment
   - Deploy to production (if applicable)

4. **Create runbook**:
   - Troubleshooting guide
   - Common issues
   - Monitoring setup

**Deliverables**:
- ✅ Complete README
- ✅ API documentation
- ✅ Deployed service
- ✅ Runbook

**Learning Objectives**:
- Documentation best practices
- AWS deployment
- Operations knowledge

**Dependencies**: All previous phases

**Estimated Time**: 4-6 hours

---

## Implementation Timeline

### Week 1: Foundation & Domain Layer
- Phase 0: Project Setup
- Phase 1: Domain Layer

### Week 2: Application & Infrastructure Core
- Phase 2: Application Layer
- Phase 3: DynamoDB Adapter

### Week 3: External Integration & Resilience
- Phase 4: External API Adapter
- Phase 5: Circuit Breaker

### Week 4: Handlers & Configuration
- Phase 6: Cache Strategy
- Phase 7: Lambda Handlers
- Phase 8: Configuration

### Week 5: Observability & Security
- Phase 9: Logging & Observability
- Phase 10: Security Implementation

### Week 6: Infrastructure & Deployment
- Phase 11: AWS Infrastructure
- Phase 12: Testing & QA
- Phase 13: Documentation & Deployment

**Total Estimated Time**: 80-100 hours

## Success Metrics

After each phase, verify:
- ✅ Code compiles without errors
- ✅ Tests pass
- ✅ No linter warnings
- ✅ Architecture boundaries respected
- ✅ Learning objectives achieved

## Risk Mitigation

### Common Challenges & Solutions

1. **External API Rate Limits**
   - Solution: Implement aggressive caching, use multiple providers

2. **DynamoDB Costs**
   - Solution: Use on-demand pricing, optimize table design

3. **Lambda Cold Starts**
   - Solution: Optimize initialization, use provisioned concurrency if needed

4. **Complexity Overload**
   - Solution: Implement incrementally, test each phase thoroughly

## Next Steps After Completion

1. **Enhancements**:
   - Add more exchange rate providers
   - Implement WebSocket support for real-time rates
   - Add rate history tracking
   - Implement rate alerts

2. **Advanced Learning**:
   - Event-driven architecture
   - GraphQL API
   - Multi-region deployment
   - Advanced monitoring (X-Ray)

3. **Production Hardening**:
   - Load testing
   - Disaster recovery planning
   - Security audit
   - Performance optimization

---

**Note**: This plan is flexible. Adjust phases based on learning pace and specific interests. The key is maintaining clean architecture boundaries and comprehensive testing throughout.

