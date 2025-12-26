# Phase 7: Lambda Handlers & API Gateway Integration

## Overview

**Goal**: Create Lambda handlers and API Gateway integration for the currency exchange rate service.

**What We're Building**:
- Lambda handlers for all endpoints (GET /rates/{base}/{target}, GET /rates/{base}, GET /health)
- Middleware for validation, error handling, and logging
- Lambda entry point with dependency injection
- API Gateway configuration
- Comprehensive handler tests

**Why This Phase**:
- Connects the application layer to AWS Lambda
- Enables REST API access via API Gateway
- Implements request/response handling
- Demonstrates serverless architecture patterns

**Dependencies**: 
- Phase 2 (Repository) ✅ Complete
- Phase 3 (DynamoDB Adapter) ✅ Complete
- Phase 4 (External API Adapter) ✅ Complete
- Phase 6 (Cache Strategy) ✅ Complete

**Estimated Time**: 8-10 hours

---

## Step 1: Add AWS Lambda Dependencies

**Objective**: Add required AWS Lambda Go SDK dependencies

**Why**: Need AWS Lambda Go SDK for handler types

**What to Do**:

1. Add dependency:
   ```bash
   go get github.com/aws/aws-lambda-go/events
   go get github.com/aws/aws-lambda-go/lambda
   ```

2. Verify in `go.mod`

**Deliverable**: ✅ Dependencies added

**Time**: 5 minutes

---

## Step 2: Create Handler Package Structure

**Objective**: Set up directory structure for Lambda handlers

**Why**: Clear organization follows Go package conventions

**What to Do**:

1. Create directory structure:
   ```
   internal/infrastructure/adapter/lambda/
     handlers.go
     handlers_test.go
   internal/infrastructure/middleware/
     error_handler.go
     error_handler_test.go
     validation.go
     validation_test.go
     logging.go
     logging_test.go
   ```

2. Create empty files with package declarations

**Deliverable**: ✅ Package structure created

**Time**: 5 minutes

---

## Step 3: Implement Error Handling Middleware

**Objective**: Create error handling utilities for Lambda responses

**Why**: Consistent error responses across handlers

**What to Do**:

### 3.1 Error Response Helpers

- `errorResponse()` - Create error response with status code
- `successResponse()` - Create success response
- `getStatusCode()` - Map domain errors to HTTP status codes

**Key Points**:
- Map domain errors to appropriate HTTP status codes
- Return generic error messages (security)
- Include error codes for client handling
- Set proper Content-Type headers

**Deliverable**: ✅ Error handling middleware implemented

**Time**: 1 hour

---

## Step 4: Implement Validation Middleware

**Objective**: Create request validation utilities

**Why**: Validate API Gateway requests before processing

**What to Do**:

### 4.1 Path Parameter Extraction

- Extract `{base}` and `{target}` from path parameters
- Validate currency code format
- Return validation errors

### 4.2 Request Validation

- Validate HTTP method (only GET allowed)
- Validate path parameters exist
- Validate currency code format

**Deliverable**: ✅ Validation middleware implemented

**Time**: 1 hour

---

## Step 5: Implement Logging Middleware

**Objective**: Create logging utilities for Lambda handlers

**Why**: Observability and debugging

**What to Do**:

### 5.1 Request Logging

- Log incoming requests (sanitized)
- Log request ID from API Gateway
- Log response status codes

### 5.2 Error Logging

- Log errors with context
- Don't log sensitive data
- Use structured logging

**Deliverable**: ✅ Logging middleware implemented

**Time**: 30 minutes

---

## Step 6: Implement GetRate Handler

**Objective**: Create handler for GET /rates/{base}/{target}

**Why**: Handle single exchange rate requests

**What to Do**:

### 6.1 Handler Function

```go
func GetRateHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // 1. Extract and validate path parameters
    // 2. Create request DTO
    // 3. Call use case
    // 4. Format response
    // 5. Return API Gateway response
}
```

### 6.2 Key Points

- Extract `base` and `target` from `event.PathParameters`
- Validate currency codes
- Call `GetExchangeRateUseCase.Execute()`
- Handle errors appropriately
- Return JSON response

**Deliverable**: ✅ GetRate handler implemented

**Time**: 1.5 hours

---

## Step 7: Implement GetAllRates Handler

**Objective**: Create handler for GET /rates/{base}

**Why**: Handle all rates for base currency requests

**What to Do**:

Similar to Step 6, but for `GetAllRatesUseCase`:

1. Extract `base` from path parameters
2. Validate currency code
3. Call `GetAllRatesUseCase.Execute()`
4. Format response
5. Return API Gateway response

**Deliverable**: ✅ GetAllRates handler implemented

**Time**: 1 hour

---

## Step 8: Implement Health Handler

**Objective**: Create handler for GET /health

**Why**: Health check endpoint

**What to Do**:

1. Create empty request (no parameters)
2. Call `HealthCheckUseCase.Execute()`
3. Format response
4. Return appropriate status code (200 if healthy, 503 if unhealthy)

**Deliverable**: ✅ Health handler implemented

**Time**: 30 minutes

---

## Step 9: Create Lambda Entry Point

**Objective**: Wire everything together in main.go

**Why**: Dependency injection and routing

**What to Do**:

### 9.1 Dependency Initialization

- Load configuration
- Create DynamoDB client
- Create repository
- Create HTTP client
- Create API provider
- Create circuit breaker
- Create circuit breaker provider
- Create use cases

### 9.2 Handler Routing

- Route based on `event.Path` and `event.HTTPMethod`
- Call appropriate handler
- Return response

### 9.3 Lambda Start

- Use `lambda.Start()` for Lambda runtime

**Deliverable**: ✅ Lambda entry point implemented

**Time**: 2 hours

---

## Step 10: Write Handler Tests

**Objective**: Test all handlers

**Why**: Ensure correctness

**What to Do**:

### 10.1 GetRate Handler Tests

- Test successful request
- Test invalid currency codes
- Test missing path parameters
- Test use case errors
- Test circuit breaker open scenario

### 10.2 GetAllRates Handler Tests

- Test successful request
- Test invalid base currency
- Test use case errors

### 10.3 Health Handler Tests

- Test healthy response
- Test unhealthy response

**Deliverable**: ✅ Comprehensive tests written

**Time**: 2 hours

---

## Step 11: Update API Gateway Configuration

**Objective**: Verify and update SAM template

**Why**: Ensure API Gateway routes are correct

**What to Do**:

1. Review `infrastructure/template.yaml`
2. Verify routes match handlers
3. Verify CORS configuration
4. Verify API key requirement

**Deliverable**: ✅ API Gateway configuration verified

**Time**: 30 minutes

---

## Step 12: Documentation and Code Review

**Objective**: Document and review implementation

**What to Do**:

1. Document handler patterns
2. Document middleware usage
3. Code review checklist

**Deliverable**: ✅ Code documented and reviewed

**Time**: 30 minutes

---

## Summary Checklist

Before considering Phase 7 complete:

- [x] AWS Lambda dependencies added
- [x] Handler package structure created
- [x] Error handling middleware implemented
- [x] Validation middleware implemented
- [x] GetRate handler implemented
- [x] GetAllRates handler implemented
- [x] Health handler implemented
- [x] Lambda entry point implemented
- [x] Handler tests written and passing
- [x] Middleware tests written and passing
- [x] API Gateway configuration verified
- [x] Code documented
- [x] Code reviewed

**Status**: ✅ **Phase 7 Complete** - All checklist items verified and completed.

**Code Review Document**: See `docs/PHASE_7_CODE_REVIEW.md` for detailed review.

**Note**: Logging middleware was deferred to a future phase as it's not critical for Phase 7 functionality.

---

## Estimated Total Time

- Step 1: 5 minutes
- Step 2: 5 minutes
- Step 3: 1 hour
- Step 4: 1 hour
- Step 5: 30 minutes
- Step 6: 1.5 hours
- Step 7: 1 hour
- Step 8: 30 minutes
- Step 9: 2 hours
- Step 10: 2 hours
- Step 11: 30 minutes
- Step 12: 30 minutes

**Total**: ~9-10 hours

---

## Next Steps

After Phase 7 completion:
- Phase 8: Configuration & Environment Management
