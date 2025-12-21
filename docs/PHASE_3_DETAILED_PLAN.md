# Phase 3: Infrastructure Layer - DynamoDB Adapter
## Detailed Step-by-Step Implementation Plan

**Goal**: Implement DynamoDB repository adapter that implements the domain repository interface, connecting our domain layer to AWS DynamoDB.

**Why This Phase Matters**: 
- This is where we connect our **pure domain logic** (Phase 1) to **real infrastructure** (AWS DynamoDB)
- Demonstrates **Hexagonal Architecture** in action: domain defines the interface (port), infrastructure implements it (adapter)
- Enables **persistent storage** for exchange rates, making caching work in production

---

## Prerequisites Checklist

Before starting, ensure:
- [ ] Phase 1 (Domain Layer) is complete ✅
- [ ] Phase 2 (Application Layer) is complete ✅
- [ ] AWS account access (for DynamoDB)
- [ ] Go 1.21+ installed
- [ ] Understanding of DynamoDB basics (partition keys, TTL, GetItem, PutItem, Query)

---

## Step 1: Understand the Repository Interface

**Objective**: Review what we need to implement

**What to Do**:
1. Read `internal/domain/repository/exchange_rate_repository.go`
2. Understand each method signature:
   - `Get(ctx, base, target)` → Returns single rate or `ErrRateNotFound`
   - `Save(ctx, rate, ttl)` → Stores rate with TTL
   - `GetByBase(ctx, base)` → Returns all rates for a base currency
   - `Delete(ctx, base, target)` → Removes a rate
   - `GetStale(ctx, base, target)` → Returns expired rate for fallback

**Key Understanding**:
- Repository is a **port** (interface) defined in domain layer
- DynamoDB adapter is an **adapter** (implementation) in infrastructure layer
- Repository **doesn't validate TTL** - it just stores/retrieves. Use cases handle expiration.

**Deliverable**: ✅ Understanding of interface requirements

**Time**: 10 minutes

---

## Step 2: Set Up AWS SDK v2 Dependencies

**Objective**: Add AWS SDK v2 for Go to project dependencies

**Why**: AWS SDK v2 is the modern, recommended SDK for Go. It provides:
- Better performance
- Context support
- Modular design
- Better error handling

**What to Do**:
1. Add dependencies to `go.mod`:
   ```bash
   go get github.com/aws/aws-sdk-go-v2/config
   go get github.com/aws/aws-sdk-go-v2/service/dynamodb
   go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
   ```

2. Verify dependencies:
   ```bash
   go mod tidy
   go mod verify
   ```

**Key Packages**:
- `config`: Loads AWS configuration (credentials, region)
- `service/dynamodb`: DynamoDB client
- `attributevalue`: Converts Go structs ↔ DynamoDB attributes

**Deliverable**: ✅ AWS SDK v2 dependencies added to `go.mod`

**Time**: 5 minutes

---

## Step 3: Design DynamoDB Table Schema

**Objective**: Design the table structure for storing exchange rates

**Why**: Good schema design is crucial for:
- Efficient queries
- Cost optimization
- Scalability

**What to Do**:

### 3.1 Understand Access Patterns

**Primary Access Patterns**:
1. Get rate by currency pair: `Get(base, target)` → Need partition key
2. Get all rates for base: `GetByBase(base)` → Need GSI
3. Save/Update rate: `Save(rate, ttl)` → Need partition key
4. Delete rate: `Delete(base, target)` → Need partition key

### 3.2 Design Schema

**Table Name**: `ExchangeRates` (configurable)

**Primary Key**:
- **Partition Key (PK)**: `RATE#USD#EUR` (format: `RATE#{BASE}#{TARGET}`)
  - Why: Allows direct lookup for `Get()` and `Delete()`
  - Example: `RATE#USD#EUR`, `RATE#GBP#JPY`

**Sort Key (SK)**: Not needed for Phase 3 (can add later for versioning)

**Attributes**:
- `PK` (String): Partition key
- `Base` (String): Base currency code (e.g., "USD")
- `Target` (String): Target currency code (e.g., "EUR")
- `Rate` (Number): Exchange rate (e.g., 0.85)
- `Timestamp` (Number): Unix timestamp in seconds
- `Stale` (Boolean): Whether rate is marked as stale
- `ttl` (Number): TTL timestamp (Unix epoch in seconds) - DynamoDB TTL attribute

**Global Secondary Index (GSI)**:
- **Index Name**: `BaseCurrencyIndex`
- **Partition Key**: `Base` (e.g., "USD")
- **Purpose**: Enable `GetByBase()` query efficiently

### 3.3 Document Schema

Create a schema document or diagram showing:
- Table structure
- Attribute types
- GSI structure
- Example items

**Deliverable**: ✅ Schema design documented (can be in code comments or separate doc)

**Time**: 20 minutes

---

## Step 4: Create Infrastructure Directory Structure

**Objective**: Set up the infrastructure layer directory structure

**Why**: Clear organization follows hexagonal architecture principles

**What to Do**:
1. Create directory structure:
   ```
   internal/
     infrastructure/
       adapter/
         dynamodb/
           exchange_rate_repository.go
           exchange_rate_repository_test.go
       config/
         dynamodb.go
   ```

2. Create empty files with package declarations:
   - `internal/infrastructure/adapter/dynamodb/exchange_rate_repository.go`
   - `internal/infrastructure/adapter/dynamodb/exchange_rate_repository_test.go`
   - `internal/infrastructure/config/dynamodb.go`

**Deliverable**: ✅ Directory structure created

**Time**: 5 minutes

---

## Step 5: Implement DynamoDB Configuration

**Objective**: Create configuration helper for DynamoDB client

**Why**: 
- Centralizes AWS configuration
- Makes testing easier (can inject mock client)
- Follows dependency injection pattern

**What to Do**:

### 5.1 Create Configuration Function

In `internal/infrastructure/config/dynamodb.go`:

```go
package config

import (
    "context"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// NewDynamoDBClient creates a new DynamoDB client with default configuration.
// In production, this will use AWS credentials from environment/IAM role.
// For local development, use AWS credentials file or environment variables.
func NewDynamoDBClient(ctx context.Context) (*dynamodb.Client, error) {
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }
    
    return dynamodb.NewFromConfig(cfg), nil
}

// DynamoDBConfig holds configuration for DynamoDB repository.
type DynamoDBConfig struct {
    TableName string
    Client    *dynamodb.Client
}
```

**Key Points**:
- `LoadDefaultConfig()` loads credentials from:
  - Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
  - AWS credentials file (`~/.aws/credentials`)
  - IAM role (when running on AWS)
- Returns configured client ready to use

**Deliverable**: ✅ Configuration function implemented

**Time**: 15 minutes

---

## Step 6: Implement Entity ↔ DynamoDB Mapping

**Objective**: Create helper functions to convert between domain entities and DynamoDB items

**Why**: 
- DynamoDB uses AttributeValue format (not Go structs directly)
- Need to map domain entities to DynamoDB attributes
- Need to map DynamoDB attributes back to domain entities

**What to Do**:

### 6.1 Create Mapping Functions

In `internal/infrastructure/adapter/dynamodb/exchange_rate_repository.go`:

Add helper functions:

```go
// dynamoItem represents a DynamoDB item structure
type dynamoItem struct {
    PK        string  `dynamodbav:"PK"`
    Base      string  `dynamodbav:"Base"`
    Target    string  `dynamodbav:"Target"`
    Rate      float64 `dynamodbav:"Rate"`
    Timestamp int64   `dynamodbav:"Timestamp"`
    Stale     bool    `dynamodbav:"Stale"`
    TTL       *int64  `dynamodbav:"ttl,omitempty"` // Optional, for TTL
}

// entityToDynamoItem converts a domain entity to DynamoDB item
func entityToDynamoItem(rate *entity.ExchangeRate, ttl time.Duration) (*dynamoItem, error) {
    // Calculate TTL timestamp (Unix epoch in seconds)
    var ttlTimestamp *int64
    if ttl > 0 {
        ttlSec := time.Now().Add(ttl).Unix()
        ttlTimestamp = &ttlSec
    }
    
    return &dynamoItem{
        PK:        buildPartitionKey(rate.Base, rate.Target),
        Base:      rate.Base.String(),
        Target:    rate.Target.String(),
        Rate:      rate.Rate,
        Timestamp: rate.Timestamp.Unix(),
        Stale:     rate.Stale,
        TTL:       ttlTimestamp,
    }, nil
}

// dynamoItemToEntity converts a DynamoDB item to domain entity
func dynamoItemToEntity(item *dynamoItem) (*entity.ExchangeRate, error) {
    base, err := entity.NewCurrencyCode(item.Base)
    if err != nil {
        return nil, fmt.Errorf("invalid base currency in stored data: %w", err)
    }
    
    target, err := entity.NewCurrencyCode(item.Target)
    if err != nil {
        return nil, fmt.Errorf("invalid target currency in stored data: %w", err)
    }
    
    timestamp := time.Unix(item.Timestamp, 0)
    
    return entity.NewExchangeRate(base, target, item.Rate, timestamp, item.Stale)
}

// buildPartitionKey creates a partition key from currency codes
func buildPartitionKey(base, target entity.CurrencyCode) string {
    return fmt.Sprintf("RATE#%s#%s", base.String(), target.String())
}
```

**Key Understanding**:
- `dynamodbav` tags tell AWS SDK how to map struct fields to DynamoDB attributes
- TTL is stored as Unix timestamp (seconds)
- Partition key format: `RATE#{BASE}#{TARGET}`

**Deliverable**: ✅ Mapping functions implemented

**Time**: 30 minutes

---

## Step 7: Implement Get() Method

**Objective**: Implement the `Get()` method to retrieve a single exchange rate

**Why**: This is the most common operation - fetching a cached rate

**What to Do**:

### 7.1 Implement Get() Method

```go
func (r *DynamoDBRepository) Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // Build partition key
    pk := buildPartitionKey(base, target)
    
    // Prepare GetItem input
    input := &dynamodb.GetItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "PK": &types.AttributeValueMemberS{Value: pk},
        },
    }
    
    // Execute GetItem
    result, err := r.client.GetItem(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("dynamodb get item failed: %w", err)
    }
    
    // Check if item exists
    if result.Item == nil {
        return nil, entity.ErrRateNotFound
    }
    
    // Convert DynamoDB item to domain entity
    var item dynamoItem
    err = attributevalue.UnmarshalMap(result.Item, &item)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal dynamodb item: %w", err)
    }
    
    // Convert to domain entity
    return dynamoItemToEntity(&item)
}
```

**Key Points**:
- Uses `GetItem` for direct lookup by partition key
- Returns `entity.ErrRateNotFound` if item doesn't exist (domain error)
- Maps infrastructure errors to domain errors

**Deliverable**: ✅ `Get()` method implemented

**Time**: 30 minutes

---

## Step 8: Implement Save() Method

**Objective**: Implement the `Save()` method to store/update an exchange rate

**Why**: Needed to cache rates after fetching from external API

**What to Do**:

### 8.1 Implement Save() Method

```go
func (r *DynamoDBRepository) Save(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error {
    // Convert entity to DynamoDB item
    item, err := entityToDynamoItem(rate, ttl)
    if err != nil {
        return fmt.Errorf("failed to convert entity to dynamo item: %w", err)
    }
    
    // Marshal to DynamoDB AttributeValue map
    av, err := attributevalue.MarshalMap(item)
    if err != nil {
        return fmt.Errorf("failed to marshal dynamo item: %w", err)
    }
    
    // Prepare PutItem input (upsert behavior)
    input := &dynamodb.PutItemInput{
        TableName: aws.String(r.tableName),
        Item:      av,
    }
    
    // Execute PutItem
    _, err = r.client.PutItem(ctx, input)
    if err != nil {
        return fmt.Errorf("dynamodb put item failed: %w", err)
    }
    
    return nil
}
```

**Key Points**:
- `PutItem` creates or updates (upsert behavior)
- TTL is calculated and stored as Unix timestamp
- All errors are wrapped with context

**Deliverable**: ✅ `Save()` method implemented

**Time**: 25 minutes

---

## Step 9: Implement GetByBase() Method

**Objective**: Implement query to get all rates for a base currency

**Why**: Needed for UC2 (Get All Rates use case)

**What to Do**:

### 9.1 Implement GetByBase() Method

```go
func (r *DynamoDBRepository) GetByBase(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
    // Query GSI by base currency
    input := &dynamodb.QueryInput{
        TableName:              aws.String(r.tableName),
        IndexName:              aws.String("BaseCurrencyIndex"),
        KeyConditionExpression: aws.String("Base = :base"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":base": &types.AttributeValueMemberS{Value: base.String()},
        },
    }
    
    // Execute query
    result, err := r.client.Query(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("dynamodb query failed: %w", err)
    }
    
    // Convert items to entities
    rates := make([]*entity.ExchangeRate, 0, len(result.Items))
    for _, item := range result.Items {
        var dItem dynamoItem
        err := attributevalue.UnmarshalMap(item, &dItem)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal item: %w", err)
        }
        
        entity, err := dynamoItemToEntity(&dItem)
        if err != nil {
            return nil, fmt.Errorf("failed to convert item to entity: %w", err)
        }
        
        rates = append(rates, entity)
    }
    
    // Return empty slice (not nil) if no results
    return rates, nil
}
```

**Key Points**:
- Uses `Query` operation (not Scan) for efficiency
- Queries GSI `BaseCurrencyIndex`
- Returns empty slice (not nil) per interface contract

**Deliverable**: ✅ `GetByBase()` method implemented

**Time**: 30 minutes

---

## Step 10: Implement Delete() and GetStale() Methods

**Objective**: Implement remaining repository methods

**Why**: Complete the interface implementation

**What to Do**:

### 10.1 Implement Delete() Method

```go
func (r *DynamoDBRepository) Delete(ctx context.Context, base, target entity.CurrencyCode) error {
    pk := buildPartitionKey(base, target)
    
    input := &dynamodb.DeleteItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "PK": &types.AttributeValueMemberS{Value: pk},
        },
        ReturnValues: types.ReturnValueAllOld,
    }
    
    result, err := r.client.DeleteItem(ctx, input)
    if err != nil {
        return fmt.Errorf("dynamodb delete item failed: %w", err)
    }
    
    // Check if item existed
    if result.Attributes == nil {
        return entity.ErrRateNotFound
    }
    
    return nil
}
```

### 10.2 Implement GetStale() Method

For Phase 3, `GetStale()` can be the same as `Get()` (repository doesn't filter by TTL):

```go
func (r *DynamoDBRepository) GetStale(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // For now, GetStale is the same as Get
    // Repository doesn't filter by TTL - use cases handle expiration
    return r.Get(ctx, base, target)
}
```

**Key Points**:
- `Delete()` returns `ErrRateNotFound` if item doesn't exist
- `GetStale()` is semantically the same as `Get()` (repository doesn't validate TTL)

**Deliverable**: ✅ All repository methods implemented

**Time**: 20 minutes

---

## Step 11: Create Repository Constructor

**Objective**: Create a constructor function for the repository

**Why**: Follows Go best practices and enables dependency injection

**What to Do**:

### 11.1 Define Repository Struct

```go
// DynamoDBRepository implements ExchangeRateRepository using DynamoDB
type DynamoDBRepository struct {
    client    *dynamodb.Client
    tableName string
}
```

### 11.2 Create Constructor

```go
// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(client *dynamodb.Client, tableName string) *DynamoDBRepository {
    return &DynamoDBRepository{
        client:    client,
        tableName: tableName,
    }
}
```

**Key Points**:
- Takes client and table name as parameters (dependency injection)
- Client can be real DynamoDB client or mock for testing

**Deliverable**: ✅ Repository struct and constructor implemented

**Time**: 10 minutes

---

## Step 12: Error Mapping and Context Handling

**Objective**: Ensure proper error mapping and context cancellation

**Why**: 
- Domain layer expects domain errors (`entity.ErrRateNotFound`)
- Infrastructure errors should be mapped appropriately
- Context cancellation must be respected

**What to Do**:

### 12.1 Add Error Mapping Helper

```go
// mapDynamoDBError maps DynamoDB errors to domain errors
func mapDynamoDBError(err error) error {
    if err == nil {
        return nil
    }
    
    // Check for context cancellation
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return err // Return context errors as-is
    }
    
    // Check for "item not found" scenarios
    var notFoundErr *types.ResourceNotFoundException
    if errors.As(err, &notFoundErr) {
        return entity.ErrRateNotFound
    }
    
    // For other errors, wrap with context
    return fmt.Errorf("dynamodb operation failed: %w", err)
}
```

### 12.2 Update Methods to Use Error Mapping

Update all methods to check context and use error mapping:

```go
func (r *DynamoDBRepository) Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // Check context before operation
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    
    // ... existing Get() implementation ...
    
    // Map errors
    if err != nil {
        return nil, mapDynamoDBError(err)
    }
    
    // ... rest of implementation ...
}
```

**Deliverable**: ✅ Error mapping and context handling implemented

**Time**: 20 minutes

---

## Step 13: Write Unit Tests (Mock-Based)

**Objective**: Write unit tests using mocks for DynamoDB client

**Why**: 
- Test repository logic without real DynamoDB
- Fast execution
- Test error handling

**What to Do**:

### 13.1 Create Mock DynamoDB Client

Use a mocking library or create a simple mock:

```go
// mockDynamoDBClient is a mock implementation for testing
type mockDynamoDBClient struct {
    getItemFunc    func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
    putItemFunc    func(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
    queryFunc      func(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
    deleteItemFunc func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}
```

### 13.2 Write Test Cases

Test each method:
- `Get()`: success, not found, error
- `Save()`: success, error
- `GetByBase()`: success, empty results, error
- `Delete()`: success, not found, error
- `GetStale()`: same as Get

**Deliverable**: ✅ Unit tests with mocks implemented

**Time**: 60 minutes

---

## Step 14: Write Integration Tests (Optional but Recommended)

**Objective**: Test against real DynamoDB (local or test table)

**Why**: 
- Catches integration issues
- Tests actual DynamoDB behavior
- Validates schema and GSI

**What to Do**:

### 14.1 Set Up DynamoDB Local or Test Table

**Option A: DynamoDB Local**
```bash
# Download DynamoDB Local
# Run: java -Djava.library.path=./DynamoDBLocal_lib -jar DynamoDBLocal.jar
```

**Option B: Test Table in AWS**
- Create a test table in AWS
- Use separate table for testing

### 14.2 Write Integration Tests

Test scenarios:
- Create table schema
- Save rate → Get rate
- Save rate → GetByBase
- Save rate → Delete rate
- TTL expiration (if testing with real TTL)

**Deliverable**: ✅ Integration tests implemented

**Time**: 45 minutes (if doing this step)

---

## Step 15: Documentation and Code Review

**Objective**: Document the implementation and review code quality

**What to Do**:

1. **Add Code Comments**:
   - Document each method
   - Explain DynamoDB-specific logic
   - Document error handling

2. **Update Architecture Docs**:
   - Document table schema
   - Document GSI usage
   - Add diagrams if helpful

3. **Code Review Checklist**:
   - [x] All methods implemented
   - [x] Error mapping correct
   - [x] Context handling correct
   - [x] Tests passing
   - [x] No hardcoded values (design constants acceptable)
   - [x] Follows Go conventions

**Code Review Document**: See `docs/PHASE_3_CODE_REVIEW.md` for detailed review.

**Deliverable**: ✅ Code documented and reviewed

**Time**: 30 minutes

---

## Summary Checklist

Before considering Phase 3 complete:

- [x] AWS SDK v2 dependencies added
- [x] Table schema designed and documented
- [x] Infrastructure directory structure created
- [x] DynamoDB configuration implemented
- [x] Entity ↔ DynamoDB mapping functions implemented
- [x] `Get()` method implemented and tested
- [x] `Save()` method implemented and tested
- [x] `GetByBase()` method implemented and tested
- [x] `Delete()` method implemented and tested
- [x] `GetStale()` method implemented and tested
- [x] Repository constructor implemented
- [x] Error mapping implemented
- [x] Context handling implemented
- [x] Unit tests written and passing
- [x] Integration tests written (optional)
- [x] Code documented
- [x] Code reviewed

**Status**: ✅ **Phase 3 Complete** - All checklist items verified and completed.

---

## Estimated Total Time

- **Minimum (without integration tests)**: 5-6 hours
- **With integration tests**: 6-8 hours

---

## Next Steps After Phase 3

Once Phase 3 is complete:
1. Phase 4: External API Adapter (implement provider interface)
2. Phase 5: Lambda Handler (wire everything together)
3. Phase 6: AWS SAM Template (infrastructure as code)

---

## Learning Resources

- [AWS SDK for Go v2 Documentation](https://aws.github.io/aws-sdk-go-v2/docs/)
- [DynamoDB Best Practices](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/best-practices.html)
- [DynamoDB TTL](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html)
- [Go Context Package](https://pkg.go.dev/context)

---

**Note**: This plan is designed for step-by-step execution. Complete each step, test it, and get feedback before moving to the next step. This ensures understanding and catches issues early.
