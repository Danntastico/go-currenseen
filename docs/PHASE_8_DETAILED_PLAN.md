# Phase 8: Configuration & Environment Management - Detailed Plan

## Overview
This phase consolidates configuration management into a unified system, integrates AWS Secrets Manager for secure API key storage, and provides comprehensive configuration validation and testing.

## Goals
1. Create a unified configuration structure
2. Integrate AWS Secrets Manager for API keys
3. Consolidate all configuration loading
4. Add configuration validation
5. Write comprehensive tests

## Implementation Steps

### Step 1: Create Main Configuration Struct
**File**: `internal/infrastructure/config/config.go`

**Tasks**:
- Define `Config` struct containing all configuration:
  - DynamoDB settings (table name, region)
  - API settings (base URL, timeout, retry attempts)
  - Circuit breaker settings (failure threshold, cooldown, success threshold)
  - Cache settings (TTL)
  - Lambda settings (if needed)
  - Secrets Manager settings (secret name, cache TTL)
- Define `LoadConfig()` function that:
  - Loads all configuration from environment variables
  - Uses existing `LoadAPIConfig()`, `LoadCircuitBreakerConfig()` functions
  - Validates required fields
  - Returns error if validation fails
- Define `Validate()` method on `Config` struct
- Provide sensible defaults where appropriate

**Key Considerations**:
- Keep backward compatibility with existing config functions
- Use environment variable names that match AWS SAM template
- Validate required fields (e.g., TABLE_NAME)

### Step 2: Create Secrets Manager Integration
**File**: `internal/infrastructure/config/secrets.go`

**Tasks**:
- Create `SecretsManager` interface for testability
- Implement `AWSSecretsManager` struct:
  - Wraps AWS Secrets Manager client
  - Implements secret caching with TTL
  - Handles secret rotation
- Create `GetAPIKey()` method:
  - Fetches API key from Secrets Manager
  - Caches result with configurable TTL
  - Handles errors gracefully
- Create `NewSecretsManager()` constructor
- Add configuration for:
  - Secret name/ARN
  - Cache TTL (default: 5 minutes)
  - Region

**Key Considerations**:
- Cache secrets to reduce API calls (but respect TTL)
- Handle secret rotation (invalidate cache on rotation)
- Use IAM roles for authentication (not hardcoded credentials)
- Return errors that can be handled by callers

### Step 3: Update Main Configuration to Include Secrets
**File**: `internal/infrastructure/config/config.go`

**Tasks**:
- Add `SecretsManagerConfig` to `Config` struct
- Update `LoadConfig()` to load secrets manager configuration
- Add `GetAPIKey()` method to `Config` that uses Secrets Manager
- Support fallback to environment variable if Secrets Manager is not configured

**Key Considerations**:
- Make Secrets Manager optional (for local development)
- Support both Secrets Manager and environment variable for API keys
- Validate secret name format if provided

### Step 4: Update Lambda Main to Use Unified Config
**File**: `cmd/lambda/main.go`

**Tasks**:
- Replace scattered `os.Getenv()` calls with `config.LoadConfig()`
- Use unified `Config` struct throughout
- Initialize Secrets Manager if configured
- Pass `Config` to dependency initialization functions

**Key Considerations**:
- Maintain backward compatibility
- Handle missing configuration gracefully
- Log configuration loading (without sensitive data)

### Step 5: Write Configuration Tests
**Files**: 
- `internal/infrastructure/config/config_test.go`
- `internal/infrastructure/config/secrets_test.go`

**Tasks**:
- Test `LoadConfig()` with:
  - All environment variables set
  - Missing required variables
  - Invalid values
  - Default values
- Test `Config.Validate()`:
  - Valid configuration
  - Missing required fields
  - Invalid values
- Test Secrets Manager:
  - Successful secret retrieval
  - Caching behavior
  - Error handling (secret not found, access denied)
  - Cache expiration
- Use mocks for AWS Secrets Manager client

**Key Considerations**:
- Test all error paths
- Test default values
- Mock AWS services for unit tests
- Test cache behavior

### Step 6: Update Existing Config Files (Optional Refactor)
**Files**: 
- `internal/infrastructure/config/api.go`
- `internal/infrastructure/config/circuitbreaker.go`

**Tasks**:
- Keep existing functions for backward compatibility
- Optionally add comments referencing unified `Config` struct
- Ensure consistency with unified config

**Key Considerations**:
- Don't break existing code
- Maintain backward compatibility

## Deliverables Checklist

- [x] `internal/infrastructure/config/config.go` - Main configuration struct and loader
- [x] `internal/infrastructure/config/config_test.go` - Configuration tests
- [x] `internal/infrastructure/config/secrets.go` - Secrets Manager integration
- [x] `internal/infrastructure/config/secrets_test.go` - Secrets Manager tests
- [x] Updated `cmd/lambda/main.go` - Use unified configuration
- [x] All tests passing
- [x] Code compiles without errors
- [x] Documentation updated

## Learning Objectives

1. **Configuration Patterns**:
   - Unified configuration structure
   - Environment variable loading
   - Configuration validation
   - Default values

2. **AWS Secrets Manager**:
   - Secret storage and retrieval
   - Secret caching
   - Secret rotation handling
   - IAM role-based access

3. **Environment Management**:
   - Development vs production configuration
   - Secure credential handling
   - Configuration testing

## Dependencies

- Phase 3: DynamoDB Adapter (for table name config)
- Phase 4: External API Adapter (for API config)
- Phase 5: Circuit Breaker (for circuit breaker config)
- Phase 7: Lambda Handlers (for integration)

## Testing Strategy

1. **Unit Tests**:
   - Configuration loading with various environment variable combinations
   - Configuration validation
   - Secrets Manager with mocked AWS client
   - Cache behavior

2. **Integration Tests** (optional):
   - Test with real AWS Secrets Manager (if credentials available)
   - Test configuration loading in Lambda-like environment

## Security Considerations

- **Never log secrets**: Configuration loading should never log API keys or secrets
- **Use IAM roles**: Secrets Manager access should use IAM roles, not hardcoded credentials
- **Cache securely**: Cached secrets should be stored in memory only, not persisted
- **Validate secret names**: Prevent secret name injection attacks
- **Handle rotation**: Invalidate cache when secrets are rotated

## Next Steps After Phase 8

- Phase 9: Logging & Observability (will use configuration for log levels)
- Phase 10: Security Implementation (will use configuration for API key validation)


