# Currency Exchange Rate Service - Architecture Definition

## Table of Contents
1. [System Overview](#system-overview)
2. [Architecture Pattern](#architecture-pattern)
3. [High-Level Architecture](#high-level-architecture)
4. [Hexagonal Architecture Layers](#hexagonal-architecture-layers)
5. [Component Interaction Flow](#component-interaction-flow)
6. [Data Flow Diagrams](#data-flow-diagrams)
7. [AWS Infrastructure Architecture](#aws-infrastructure-architecture)
8. [Design Patterns](#design-patterns)
9. [Technology Stack](#technology-stack)

---

## System Overview

The Currency Exchange Rate Service is a serverless microservice that provides real-time and cached exchange rates through a REST API. It implements resilience patterns to ensure high availability even when external dependencies fail.

### Key Characteristics
- **Serverless**: AWS Lambda-based, no server management
- **Resilient**: Circuit breaker, cache fallback, retry logic
- **Scalable**: Auto-scales with API Gateway and Lambda
- **Cost-Effective**: Pay-per-use model, efficient caching
- **Observable**: Structured logging and CloudWatch metrics

---

## Architecture Pattern

**Hexagonal Architecture (Ports & Adapters)**

The system follows Hexagonal Architecture principles, ensuring:
- **Dependency Inversion**: Inner layers don't depend on outer layers
- **Testability**: Easy to mock external dependencies
- **Flexibility**: Easy to swap implementations (e.g., different providers, databases)
- **Clear Boundaries**: Well-defined interfaces between layers

```mermaid
graph TB
    subgraph "External World"
        Client[API Client]
        ExtAPI[External Exchange Rate API]
    end
    
    subgraph "Infrastructure Layer (Adapters)"
        Lambda[Lambda Handler]
        DynamoDB[DynamoDB Adapter]
        HTTPClient[HTTP Client Adapter]
    end
    
    subgraph "Application Layer"
        UC1[Get Rate Use Case]
        UC2[Get All Rates Use Case]
        UC3[Health Check Use Case]
    end
    
    subgraph "Domain Layer (Core)"
        Entity[ExchangeRate Entity]
        RepoPort[Repository Port]
        ProviderPort[Provider Port]
    end
    
    Client -->|HTTP Request| Lambda
    Lambda -->|Calls| UC1
    Lambda -->|Calls| UC2
    Lambda -->|Calls| UC3
    
    UC1 -->|Uses| RepoPort
    UC1 -->|Uses| ProviderPort
    
    RepoPort -.->|Implemented by| DynamoDB
    ProviderPort -.->|Implemented by| HTTPClient
    
    DynamoDB -->|Read/Write| ExtAPI
    HTTPClient -->|Fetch| ExtAPI
    
    style Domain fill:#e1f5ff
    style Application fill:#fff4e1
    style Infrastructure fill:#ffe1f5
```

---

## High-Level Architecture

```mermaid
graph LR
    subgraph "Client Layer"
        Browser[Web Browser]
        Mobile[Mobile App]
        API_Client[API Client]
    end
    
    subgraph "AWS API Gateway"
        APIGW[API Gateway<br/>REST API]
        Auth[API Key Auth]
        Throttle[Rate Limiting]
    end
    
    subgraph "AWS Lambda"
        Handler[Lambda Handler]
        Middleware[Middleware Stack]
    end
    
    subgraph "Application Core"
        UseCases[Use Cases]
        CircuitBreaker[Circuit Breaker]
    end
    
    subgraph "Data Layer"
        DynamoDB[(DynamoDB<br/>Cache)]
        Secrets[Secrets Manager<br/>API Keys]
    end
    
    subgraph "External Services"
        Provider1[Exchange Rate API<br/>Primary]
        Provider2[Exchange Rate API<br/>Fallback]
    end
    
    subgraph "Observability"
        CloudWatch[CloudWatch<br/>Logs & Metrics]
    end
    
    Browser --> APIGW
    Mobile --> APIGW
    API_Client --> APIGW
    
    APIGW --> Auth
    Auth --> Throttle
    Throttle --> Handler
    
    Handler --> Middleware
    Middleware --> UseCases
    
    UseCases --> CircuitBreaker
    CircuitBreaker --> Provider1
    CircuitBreaker --> Provider2
    
    UseCases --> DynamoDB
    UseCases --> Secrets
    
    Handler --> CloudWatch
    DynamoDB --> CloudWatch
    
    style DynamoDB fill:#ff9999
    style Provider1 fill:#99ff99
    style Provider2 fill:#99ff99
    style CloudWatch fill:#ffcc99
```

---

## Hexagonal Architecture Layers

### Layer Structure

```mermaid
graph TB
    subgraph "Domain Layer (Inner Core)"
        Entities[Entities<br/>ExchangeRate, CurrencyCode]
        DomainErrors[Domain Errors]
        DomainServices[Domain Services<br/>Validation, Calculation]
        Ports[Ports/Interfaces<br/>Repository, Provider]
    end
    
    subgraph "Application Layer"
        UseCases[Use Cases<br/>GetRate, GetAllRates, HealthCheck]
        DTOs[DTOs<br/>Request, Response]
        Mappers[Entity ↔ DTO Mappers]
    end
    
    subgraph "Infrastructure Layer (Adapters)"
        LambdaAdapter[Lambda Adapter<br/>Handlers]
        DynamoDBAdapter[DynamoDB Adapter<br/>Repository Implementation]
        HTTPAdapter[HTTP Adapter<br/>Provider Implementation]
        ConfigAdapter[Config Adapter<br/>Environment, Secrets]
        LoggerAdapter[Logger Adapter<br/>CloudWatch]
    end
    
    subgraph "External Systems"
        APIGateway[API Gateway]
        DynamoDB[(DynamoDB)]
        ExternalAPI[External API]
        SecretsMgr[Secrets Manager]
    end
    
    Entities --> UseCases
    Ports --> UseCases
    DomainServices --> UseCases
    
    UseCases --> DTOs
    UseCases --> Mappers
    
    LambdaAdapter --> UseCases
    DynamoDBAdapter -.->|implements| Ports
    HTTPAdapter -.->|implements| Ports
    
    LambdaAdapter --> APIGateway
    DynamoDBAdapter --> DynamoDB
    HTTPAdapter --> ExternalAPI
    ConfigAdapter --> SecretsMgr
    LoggerAdapter --> CloudWatch
    
    style Domain fill:#e1f5ff
    style Application fill:#fff4e1
    style Infrastructure fill:#ffe1f5
```

### Dependency Flow

**Rule**: Dependencies point INWARD (toward Domain)
- Infrastructure → Application → Domain ✅
- Domain → Application → Infrastructure ❌

```mermaid
graph LR
    I[Infrastructure] -->|depends on| A[Application]
    A -->|depends on| D[Domain]
    
    style D fill:#e1f5ff
    style A fill:#fff4e1
    style I fill:#ffe1f5
```

---

## Component Interaction Flow

### Get Exchange Rate Flow (UC1)

```mermaid
sequenceDiagram
    participant Client
    participant APIGateway
    participant LambdaHandler
    participant GetRateUseCase
    participant Repository
    participant CircuitBreaker
    participant ExternalAPI
    participant DynamoDB
    
    Client->>APIGateway: GET /rates/USD/EUR
    APIGateway->>LambdaHandler: API Gateway Event
    LambdaHandler->>LambdaHandler: Validate Input
    LambdaHandler->>GetRateUseCase: Execute(base, target)
    
    GetRateUseCase->>Repository: Get(base, target)
    Repository->>DynamoDB: GetItem(PK=RATE#USD#EUR)
    
    alt Cache Hit & Valid
        DynamoDB-->>Repository: ExchangeRate (cached)
        Repository-->>GetRateUseCase: ExchangeRate
        GetRateUseCase-->>LambdaHandler: RateResponse
        LambdaHandler-->>APIGateway: 200 OK
        APIGateway-->>Client: JSON Response
    else Cache Miss or Expired
        DynamoDB-->>Repository: Not Found / Expired
        Repository-->>GetRateUseCase: Cache Miss
        
        GetRateUseCase->>CircuitBreaker: Call Provider
        alt Circuit Closed
            CircuitBreaker->>ExternalAPI: FetchRate(USD, EUR)
            ExternalAPI-->>CircuitBreaker: ExchangeRate
            CircuitBreaker-->>GetRateUseCase: ExchangeRate
            
            GetRateUseCase->>Repository: Save(rate)
            Repository->>DynamoDB: PutItem(rate with TTL)
            GetRateUseCase-->>LambdaHandler: RateResponse
        else Circuit Open
            CircuitBreaker-->>GetRateUseCase: Error (Circuit Open)
            GetRateUseCase->>Repository: GetStale(base, target)
            Repository->>DynamoDB: GetItem (ignore TTL)
            DynamoDB-->>Repository: Stale Rate
            Repository-->>GetRateUseCase: Stale ExchangeRate
            GetRateUseCase-->>LambdaHandler: RateResponse (stale, with warning)
        end
        
        LambdaHandler-->>APIGateway: 200 OK
        APIGateway-->>Client: JSON Response
    end
```

### Get All Rates Flow (UC2)

```mermaid
sequenceDiagram
    participant Client
    participant LambdaHandler
    participant GetAllRatesUseCase
    participant Repository
    participant ExternalAPI
    participant DynamoDB
    
    Client->>LambdaHandler: GET /rates/USD
    LambdaHandler->>GetAllRatesUseCase: Execute(base)
    
    GetAllRatesUseCase->>Repository: GetByBase(USD)
    Repository->>DynamoDB: Query(GSI: base=USD)
    
    alt All Rates Cached
        DynamoDB-->>Repository: List of ExchangeRates
        Repository-->>GetAllRatesUseCase: Rates
        GetAllRatesUseCase-->>LambdaHandler: RatesResponse
    else Partial or No Cache
        GetAllRatesUseCase->>ExternalAPI: FetchAllRates(USD)
        ExternalAPI-->>GetAllRatesUseCase: All Rates
        
        loop For each rate
            GetAllRatesUseCase->>Repository: Save(rate)
            Repository->>DynamoDB: PutItem(rate)
        end
        
        GetAllRatesUseCase-->>LambdaHandler: RatesResponse
    end
    
    LambdaHandler-->>Client: JSON Response
```

### Health Check Flow (UC3)

```mermaid
sequenceDiagram
    participant Client
    participant LambdaHandler
    participant HealthCheckUseCase
    participant Repository
    participant ExternalAPI
    participant DynamoDB
    
    Client->>LambdaHandler: GET /health
    LambdaHandler->>HealthCheckUseCase: Execute()
    
    par Check Lambda Status
        HealthCheckUseCase->>HealthCheckUseCase: Check Lambda Runtime
    and Check DynamoDB
        HealthCheckUseCase->>Repository: Ping()
        Repository->>DynamoDB: DescribeTable()
        DynamoDB-->>Repository: Table Status
        Repository-->>HealthCheckUseCase: OK/Error
    and Check External API (Optional)
        HealthCheckUseCase->>ExternalAPI: Health Check
        ExternalAPI-->>HealthCheckUseCase: OK/Error
    end
    
    HealthCheckUseCase->>HealthCheckUseCase: Aggregate Status
    HealthCheckUseCase-->>LambdaHandler: HealthResponse
    LambdaHandler-->>Client: 200 OK / 503 Service Unavailable
```

---

## Data Flow Diagrams

### Request Flow with Resilience

```mermaid
flowchart TD
    Start([Client Request]) --> Validate{Validate Input}
    Validate -->|Invalid| Error1[400 Bad Request]
    Validate -->|Valid| Auth{Authenticate API Key}
    
    Auth -->|Invalid| Error2[401 Unauthorized]
    Auth -->|Valid| RateLimit{Rate Limit Check}
    
    RateLimit -->|Exceeded| Error3[429 Too Many Requests]
    RateLimit -->|OK| CheckCache{Check Cache}
    
    CheckCache -->|Hit & Valid| ReturnCached[Return Cached Rate]
    CheckCache -->|Miss/Expired| CheckCircuit{Circuit Breaker State}
    
    CheckCircuit -->|Closed| FetchAPI[Fetch from External API]
    CheckCircuit -->|Open| CheckStaleCache{Check Stale Cache}
    
    FetchAPI -->|Success| UpdateCache[Update Cache]
    FetchAPI -->|Failure| RecordFailure[Record Failure]
    
    RecordFailure --> CheckThreshold{Threshold Reached?}
    CheckThreshold -->|Yes| OpenCircuit[Open Circuit]
    CheckThreshold -->|No| CheckStaleCache
    
    OpenCircuit --> CheckStaleCache
    CheckStaleCache -->|Found| ReturnStale[Return Stale Rate with Warning]
    CheckStaleCache -->|Not Found| Error4[503 Service Unavailable]
    
    UpdateCache --> ReturnFresh[Return Fresh Rate]
    
    ReturnCached --> Log[Log Request]
    ReturnFresh --> Log
    ReturnStale --> Log
    Error1 --> Log
    Error2 --> Log
    Error3 --> Log
    Error4 --> Log
    
    Log --> End([Response])
    
    style CheckCircuit fill:#ffeb3b
    style CheckStaleCache fill:#ff9800
    style ReturnStale fill:#ff9800
```

### Cache Strategy Flow

```mermaid
stateDiagram-v2
    [*] --> CheckCache
    
    CheckCache --> CacheHit: Rate exists & TTL valid
    CheckCache --> CacheMiss: Rate not found
    CheckCache --> CacheExpired: Rate exists but TTL expired
    
    CacheHit --> ReturnCached: Return immediately
    ReturnCached --> [*]
    
    CacheMiss --> FetchExternal: Fetch from API
    CacheExpired --> FetchExternal: Fetch from API
    
    FetchExternal --> Success: API responds
    FetchExternal --> Failure: API fails
    
    Success --> UpdateCache: Save to DynamoDB
    UpdateCache --> ReturnFresh: Return fresh rate
    ReturnFresh --> [*]
    
    Failure --> CheckStale: Check for stale cache
    CheckStale --> ReturnStale: Stale exists
    CheckStale --> ReturnError: No stale cache
    ReturnStale --> [*]
    ReturnError --> [*]
```

---

## AWS Infrastructure Architecture

### Infrastructure Components

```mermaid
graph TB
    subgraph "Internet"
        Users[API Clients]
    end
    
    subgraph "AWS Cloud"
        subgraph "API Gateway"
            REST[REST API]
            Auth[API Key Auth]
            Throttle[Throttling<br/>10,000 req/sec]
        end
        
        subgraph "Lambda Function"
            Handler[Handler Function]
            Env[Environment Variables]
            Role[IAM Role]
        end
        
        subgraph "DynamoDB"
            Table[ExchangeRates Table]
            GSI[GSI: BaseCurrency]
            TTL[TTL Attribute]
        end
        
        subgraph "Secrets Manager"
            Secrets[API Keys<br/>External Provider Keys]
        end
        
        subgraph "CloudWatch"
            Logs[Log Groups]
            Metrics[Metrics]
            Alarms[Alarms]
        end
        
        subgraph "External"
            ExtAPI[Exchange Rate APIs]
        end
    end
    
    Users -->|HTTPS| REST
    REST --> Auth
    Auth --> Throttle
    Throttle -->|Invoke| Handler
    
    Handler -->|Read/Write| Table
    Handler -->|Query| GSI
    Handler -->|Get Secrets| Secrets
    Handler -->|Fetch Rates| ExtAPI
    
    Handler -->|Logs| Logs
    Handler -->|Metrics| Metrics
    Table -->|Metrics| Metrics
    
    Metrics --> Alarms
    
    style Handler fill:#ff9999
    style Table fill:#99ff99
    style ExtAPI fill:#ffcc99
```

### DynamoDB Table Design

```mermaid
classDiagram
    class ExchangeRates {
        +string PartitionKey
        +string SortKey
        +string BaseCurrency
        +string TargetCurrency
        +float Rate
        +string UpdatedAt
        +int TTL
        +string Metadata
    }
```

**Table Schema Details:**
- **Table Name**: `ExchangeRates`
- **Partition Key**: `PartitionKey` (stores values like `RATE#USD#EUR`)
- **Sort Key**: `SortKey` (optional, reserved for future use)
- **GSI**: `BaseCurrencyIndex` for querying all rates by base currency
  - GSI Partition Key: `BaseCurrency`
  - GSI Sort Key: `PartitionKey` (original partition key)
- **TTL Attribute**: `TTL` (Unix epoch timestamp)
- **Attributes**: `BaseCurrency`, `TargetCurrency`, `Rate`, `UpdatedAt`, `Metadata`

**Domain Entity Mapping:**
The DynamoDB table stores data that maps to the `ExchangeRate` domain entity:
- `PartitionKey` → Identifies the currency pair (format: `RATE#BASE#TARGET`)
- `BaseCurrency` → Domain entity `Base` field
- `TargetCurrency` → Domain entity `Target` field
- `Rate` → Domain entity `Rate` field
- `UpdatedAt` → Domain entity `Timestamp` field

### IAM Permissions

```mermaid
graph LR
    LambdaRole[Lambda Execution Role] --> DynamoDBPolicy[DynamoDB Policy]
    LambdaRole --> SecretsPolicy[Secrets Manager Policy]
    LambdaRole --> CloudWatchPolicy[CloudWatch Policy]
    
    DynamoDBPolicy --> DynamoDBOps[GetItem<br/>PutItem<br/>Query<br/>DescribeTable]
    
    SecretsPolicy --> SecretsOps[GetSecretValue]
    
    CloudWatchPolicy --> LogsOps[CreateLogGroup<br/>CreateLogStream<br/>PutLogEvents]
    CloudWatchPolicy --> MetricsOps[PutMetricData]
    
    style LambdaRole fill:#ff9999
    style DynamoDBPolicy fill:#99ff99
    style SecretsPolicy fill:#99ff99
    style CloudWatchPolicy fill:#99ff99
```

---

## Design Patterns

### Pattern Overview

```mermaid
graph TB
    subgraph "Creational Patterns"
        Factory[Factory Pattern<br/>Provider Creation]
    end
    
    subgraph "Structural Patterns"
        Adapter[Adapter Pattern<br/>External API Adapters]
        Repository[Repository Pattern<br/>Data Access Abstraction]
    end
    
    subgraph "Behavioral Patterns"
        Strategy[Strategy Pattern<br/>Provider Selection]
        CircuitBreaker[Circuit Breaker<br/>Resilience Pattern]
    end
    
    Factory --> Adapter
    Adapter --> Strategy
    Strategy --> CircuitBreaker
    Repository --> Adapter
    
    style Factory fill:#e1f5ff
    style Adapter fill:#fff4e1
    style Repository fill:#ffe1f5
    style Strategy fill:#e1f5ff
    style CircuitBreaker fill:#fff4e1
```

### Circuit Breaker State Machine

```mermaid
stateDiagram-v2
    [*] --> Closed: Initial State
    
    Closed --> Open: Failure Threshold Reached
    Closed --> Closed: Success
    
    Open --> HalfOpen: Cooldown Period Expired
    
    HalfOpen --> Closed: Test Call Succeeds
    HalfOpen --> Open: Test Call Fails
    
    note right of Closed
        Normal operation
        All requests pass through
    end note
    
    note right of Open
        External API failing
        Requests fail fast
        Use cache fallback
    end note
    
    note right of HalfOpen
        Testing recovery
        Single request allowed
        Monitor result
    end note
```

---

## Technology Stack

### Core Technologies

```mermaid
mindmap
    root((Currency Exchange<br/>Rate Service))
        Language
            Go 1.21+
            Go Modules
            Standard Library
        Architecture
            Hexagonal Architecture
            Clean Architecture Principles
            SOLID Principles
        AWS Services
            Lambda
            API Gateway
            DynamoDB
            CloudWatch
            Secrets Manager
        Patterns
            Repository Pattern
            Adapter Pattern
            Circuit Breaker
            Strategy Pattern
            Factory Pattern
        Libraries
            AWS SDK v2
            Structured Logging
            HTTP Client
            Testing Framework
```

---

## Key Architectural Decisions

### 1. Hexagonal Architecture
**Decision**: Use Hexagonal Architecture (Ports & Adapters)  
**Rationale**: 
- Enables easy testing through interfaces
- Allows swapping implementations (e.g., different providers, databases)
- Clear separation of concerns
- Dependency inversion principle

### 2. Serverless Architecture
**Decision**: AWS Lambda + API Gateway  
**Rationale**:
- No server management
- Auto-scaling
- Cost-effective (pay-per-use)
- Fast deployment

### 3. DynamoDB for Caching
**Decision**: Use DynamoDB instead of Redis/ElastiCache  
**Rationale**:
- Serverless (no infrastructure management)
- Built-in TTL support
- Integrated with Lambda
- Cost-effective for this use case

### 4. Circuit Breaker Pattern
**Decision**: Implement circuit breaker for external API calls  
**Rationale**:
- Prevents cascading failures
- Fast failure when external API is down
- Enables graceful degradation with cache fallback

### 5. Cache-First Strategy
**Decision**: Always check cache before external API  
**Rationale**:
- Reduces external API calls (>80% reduction)
- Faster response times (<200ms for cached)
- Lower costs
- Better resilience

---

## Component Responsibilities

### Domain Layer
- **Entities**: Core business objects (ExchangeRate, CurrencyCode)
- **Ports**: Interfaces defining contracts (Repository, Provider)
- **Domain Services**: Business logic (validation, calculation)
- **Domain Errors**: Business-specific error types

### Application Layer
- **Use Cases**: Orchestrate business workflows
- **DTOs**: Data transfer objects for API boundaries
- **Mappers**: Convert between domain entities and DTOs

### Infrastructure Layer
- **Adapters**: Implement ports (DynamoDB, HTTP clients)
- **Handlers**: Lambda function handlers
- **Middleware**: Cross-cutting concerns (logging, validation)
- **Configuration**: Environment and secrets management

---

## Data Flow Summary

1. **Request** → API Gateway → Lambda Handler
2. **Handler** → Validates input, authenticates, rate limits
3. **Use Case** → Orchestrates business logic
4. **Repository** → Checks cache (DynamoDB)
5. **Provider** → Fetches from external API (if needed)
6. **Circuit Breaker** → Monitors provider health
7. **Cache** → Stores/retrieves rates with TTL
8. **Response** → Returns rate or error
9. **Logging** → Records all operations to CloudWatch

---

## Security Architecture

```mermaid
graph TB
    subgraph "Security Layers"
        APIKey[API Key Authentication]
        InputValidation[Input Validation]
        RateLimiting[Rate Limiting]
        Secrets[Secrets Management]
        HTTPS[HTTPS/TLS]
    end
    
    subgraph "Security Measures"
        ValidateCurrency[Currency Code Validation]
        SanitizeInput[Input Sanitization]
        MaskSecrets[Secret Masking in Logs]
        IAMRoles[IAM Least Privilege]
    end
    
    APIKey --> Secrets
    InputValidation --> ValidateCurrency
    InputValidation --> SanitizeInput
    Secrets --> MaskSecrets
    RateLimiting --> IAMRoles
    HTTPS --> APIKey
    
    style APIKey fill:#ff9999
    style Secrets fill:#99ff99
    style HTTPS fill:#ffcc99
```

---

## Component Structure

### Detailed Codebase Organization

```mermaid
graph TB
    subgraph "cmd/"
        Main[lambda/main.go<br/>Entry Point]
    end
    
    subgraph "internal/domain/"
        Entity[entity/<br/>ExchangeRate<br/>CurrencyCode]
        RepoPort[repository/<br/>ExchangeRateRepository<br/>Interface]
        ProviderPort[provider/<br/>ExchangeRateProvider<br/>Interface]
        DomainService[service/<br/>ValidationService<br/>RateCalculator]
        DomainErrors[errors.go<br/>Domain Errors]
    end
    
    subgraph "internal/application/"
        UseCases[usecase/<br/>GetRateUseCase<br/>GetAllRatesUseCase<br/>HealthCheckUseCase]
        DTOs[dto/<br/>Request DTOs<br/>Response DTOs]
        Mappers[mapper.go<br/>Entity ↔ DTO]
    end
    
    subgraph "internal/infrastructure/adapter/"
        LambdaHandler[lambda/<br/>GetRateHandler<br/>GetAllRatesHandler<br/>HealthHandler]
        DynamoDBAdapter[dynamodb/<br/>ExchangeRateRepository<br/>Implementation]
        HTTPAdapter[api/<br/>ExchangeRateProvider<br/>Implementation]
        ConfigAdapter[config/<br/>Config Loader<br/>Secrets Manager]
    end
    
    subgraph "internal/infrastructure/middleware/"
        Validation[validation_middleware.go]
        Error[error_middleware.go]
        Logging[logging_middleware.go]
        Auth[auth_middleware.go]
    end
    
    subgraph "pkg/"
        CircuitBreaker[circuitbreaker/<br/>CircuitBreaker<br/>State Machine]
        Logger[pkg/logger/<br/>Structured Logger]
        CacheUtils[cache/<br/>Cache Utilities]
    end
    
    subgraph "tests/"
        UnitTests[unit/<br/>Unit Tests]
        IntegrationTests[integration/<br/>Integration Tests]
    end
    
    Main --> LambdaHandler
    LambdaHandler --> Middleware
    Middleware --> UseCases
    
    UseCases --> RepoPort
    UseCases --> ProviderPort
    UseCases --> DTOs
    UseCases --> Mappers
    
    DynamoDBAdapter -.->|implements| RepoPort
    HTTPAdapter -.->|implements| ProviderPort
    
    UseCases --> DomainService
    UseCases --> Entity
    UseCases --> DomainErrors
    
    HTTPAdapter --> CircuitBreaker
    LambdaHandler --> ConfigAdapter
    LambdaHandler --> Logger
    
    style Domain fill:#e1f5ff
    style Application fill:#fff4e1
    style Infrastructure fill:#ffe1f5
```

### Package Dependencies

```mermaid
graph LR
    subgraph "Domain (No Dependencies)"
        D[Domain Layer]
    end
    
    subgraph "Application (Depends on Domain)"
        A[Application Layer]
    end
    
    subgraph "Infrastructure (Depends on Application & Domain)"
        I[Infrastructure Layer]
    end
    
    subgraph "Packages (Shared Utilities)"
        P[Pkg Layer]
    end
    
    A --> D
    I --> A
    I --> D
    I --> P
    
    style D fill:#e1f5ff
    style A fill:#fff4e1
    style I fill:#ffe1f5
    style P fill:#f0f0f0
```

### File Structure Detail

```
go-currenseen/
├── cmd/
│   └── lambda/
│       └── main.go                    # Lambda entry point, dependency wiring
│
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── exchange_rate.go       # ExchangeRate domain entity
│   │   │   └── currency_code.go       # CurrencyCode value object
│   │   ├── repository/
│   │   │   └── exchange_rate_repository.go  # Repository port (interface)
│   │   ├── provider/
│   │   │   └── exchange_rate_provider.go    # Provider port (interface)
│   │   ├── service/
│   │   │   ├── validation_service.go  # Domain validation logic
│   │   │   └── rate_calculator.go     # Rate calculation utilities
│   │   └── errors.go                  # Domain error types
│   │
│   ├── application/
│   │   ├── usecase/
│   │   │   ├── get_exchange_rate.go   # UC1: Get rate for pair
│   │   │   ├── get_all_rates.go       # UC2: Get all rates for base
│   │   │   └── health_check.go        # UC3: Health check
│   │   ├── dto/
│   │   │   ├── request.go             # Request DTOs
│   │   │   ├── response.go            # Response DTOs
│   │   │   └── mapper.go              # Entity ↔ DTO conversion
│   │
│   └── infrastructure/
│       ├── adapter/
│       │   ├── lambda/
│       │   │   ├── get_rate_handler.go
│       │   │   ├── get_rates_handler.go
│       │   │   └── health_handler.go
│       │   ├── dynamodb/
│       │   │   └── exchange_rate_repository.go  # Repository implementation
│       │   ├── api/
│       │   │   ├── exchange_rate_provider.go    # Provider implementation
│       │   │   ├── http_client.go                # HTTP client wrapper
│       │   │   └── provider_factory.go          # Factory for providers
│       │   └── config/
│       │       ├── config.go                    # Configuration loader
│       │       └── secrets.go                   # Secrets Manager integration
│       └── middleware/
│           ├── validation_middleware.go
│           ├── error_middleware.go
│           ├── logging_middleware.go
│           └── auth_middleware.go
│
├── pkg/
│   ├── circuitbreaker/
│   │   └── circuit_breaker.go         # Circuit breaker implementation
│   ├── logger/
│   │   └── logger.go                  # Structured logger
│   └── cache/
│       └── cache.go                   # Cache utilities
│
├── tests/
│   ├── unit/                          # Unit tests
│   └── integration/                   # Integration tests
│
├── infrastructure/
│   └── sam.yaml                       # AWS SAM template
│
├── docs/
│   ├── ARCHITECTURE.md                # This file
│   ├── INITIAL_SPEC.md                # Project specification
│   └── IMPLEMENTATION_PLAN.md         # Implementation phases
│
├── go.mod
├── go.sum
└── README.md
```

---

## Next Steps

After reviewing this architecture:

1. **Phase 0**: Set up project structure matching this architecture
2. **Phase 1**: Implement domain layer (entities, ports)
3. **Phase 2**: Implement application layer (use cases, DTOs)
4. **Phase 3+**: Implement infrastructure adapters incrementally

---

**Note**: This architecture document should evolve as the implementation progresses. Update it when making significant architectural decisions or changes.

