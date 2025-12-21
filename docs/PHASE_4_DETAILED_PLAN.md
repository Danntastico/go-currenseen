# Phase 4: Infrastructure Layer - External API Adapter

## Overview

**Goal**: Implement external exchange rate API adapter following Hexagonal Architecture principles.

**What We're Building**:
- HTTP client adapter for external exchange rate APIs
- Implementation of `ExchangeRateProvider` domain interface
- Retry logic with exponential backoff
- Provider factory for creating adapters
- Comprehensive unit tests

**Why This Phase**:
- Connects domain layer to external exchange rate APIs
- Implements resilience patterns (retries, error handling)
- Demonstrates adapter pattern in Hexagonal Architecture
- Provides foundation for use cases (Phase 5)

**Dependencies**: Phase 1 (Domain Layer) ✅ Complete

**Estimated Time**: 6-8 hours

---

## Step 1: Choose Exchange Rate Provider

**Objective**: Select a free exchange rate API for development

**Why**: Need a real API to integrate with. Free tier APIs are perfect for learning.

**What to Do**:

### 1.1 Research Options

**Recommended Options**:

1. **Currency-api** (Recommended for Phase 4)
   - ✅ Completely free, **no API key required**
   - ✅ No rate limits
   - ✅ Simple REST API
   - ✅ Supports 150+ currencies
   - ✅ Supports single pair and base currency queries
   - Endpoint: `https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1/currencies/usd.json`
   - Alternative endpoint: `https://api.fawazahmed0.currency-api.com/v1/latest?base=USD`
   - Documentation: https://github.com/fawazahmed0/currency-api

2. **HexaRate** (Alternative)
   - ✅ Completely free, **no API key required**
   - ✅ Unlimited requests
   - ✅ 170+ currencies
   - ✅ Daily updates
   - Endpoint: `https://api.hexarate.paikama.co/latest?base=USD`
   - Documentation: https://hexarate.paikama.co/

3. **ExchangeRate-API** (Alternative - requires API key)
   - ✅ Free tier available (1,500 requests/month)
   - ⚠️ Requires API key (free signup)
   - ✅ More features (historical data, etc.)
   - Endpoint: `https://v6.exchangerate-api.com/v6/{API_KEY}/pair/USD/EUR`
   - Documentation: https://www.exchangerate-api.com/docs/overview

4. **exchangerate.host** (Alternative - requires API key)
   - ⚠️ Free tier (100 requests/month) - **requires API key**
   - ⚠️ Requires signup
   - Endpoint: `https://api.exchangerate.host/latest?base=USD&symbols=EUR`
   - Documentation: https://exchangerate.host/

5. **Fixer.io** (Alternative - requires API key)
   - ⚠️ Free tier (1,000 requests/month) - **requires API key**
   - ✅ Professional API
   - Endpoint: `https://api.fixer.io/latest?base=USD&symbols=EUR`
   - Documentation: https://fixer.io/documentation

**Decision for Phase 4**: Use **Currency-api** because:
- ✅ No API key needed (simpler setup for learning)
- ✅ No rate limits (good for testing)
- ✅ Simple response format
- ✅ Reliable and actively maintained
- ✅ Good for learning HTTP client patterns

### 1.2 Understand API Response Format

**Currency-api Response Format**:

**Single Rate Query**: `GET https://api.fawazahmed0.currency-api.com/v1/latest?base=USD`
```json
{
  "date": "2024-01-15",
  "usd": {
    "eur": 0.85,
    "gbp": 0.75,
    "jpy": 110.50
  }
}
```

**Note**: Currency-api returns rates in lowercase. The base currency is the key (e.g., `"usd"`), and target currencies are nested under it.

**Alternative Endpoint Format** (using query parameters):
`GET https://api.fawazahmed0.currency-api.com/v1/latest?base=USD`

Response:
```json
{
  "date": "2024-01-15",
  "base": "USD",
  "rates": {
    "EUR": 0.85,
    "GBP": 0.75,
    "JPY": 110.50
  }
}
```

**Error Response** (if base currency is invalid):
```json
{
  "error": "Invalid base currency"
}
```

**Note**: For Phase 4, we'll use the query parameter format (`?base=USD`) as it provides a cleaner response structure with uppercase currency codes.

**Deliverable**: ✅ API provider selected and response format understood

**Time**: 15 minutes

---

## Step 2: Create Directory Structure

**Objective**: Set up infrastructure layer directory structure for API adapter

**Why**: Clear organization follows hexagonal architecture principles

**What to Do**:

1. Create directory structure:
   ```
   internal/
     infrastructure/
       adapter/
         api/
           exchange_rate_provider.go
           exchange_rate_provider_test.go
           http_client.go
           http_client_test.go
           retry.go
           retry_test.go
         dynamodb/
           ... (existing)
       config/
         api.go (new)
   ```

2. Create empty files with package declarations:
   - `internal/infrastructure/adapter/api/exchange_rate_provider.go`
   - `internal/infrastructure/adapter/api/exchange_rate_provider_test.go`
   - `internal/infrastructure/adapter/api/http_client.go`
   - `internal/infrastructure/adapter/api/http_client_test.go`
   - `internal/infrastructure/adapter/api/retry.go`
   - `internal/infrastructure/adapter/api/retry_test.go`
   - `internal/infrastructure/config/api.go`

**Deliverable**: ✅ Directory structure created

**Time**: 5 minutes

---

## Step 3: Implement HTTP Client

**Objective**: Create a secure HTTP client with timeout and proper configuration

**Why**: 
- Centralizes HTTP client configuration
- Ensures security (TLS, timeouts)
- Makes testing easier (can inject mock client)
- Follows Go best practices

**What to Do**:

### 3.1 Create HTTP Client Function

In `internal/infrastructure/adapter/api/http_client.go`:

```go
package api

import (
    "crypto/tls"
    "net/http"
    "time"
)

// NewHTTPClient creates a new HTTP client with secure defaults.
//
// Configuration:
// - Timeout: 10 seconds (prevents hanging requests)
// - TLS: Minimum TLS 1.2 (security)
// - Transport: HTTP/1.1 (compatibility)
//
// The client is safe for concurrent use by multiple goroutines.
func NewHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 10 * time.Second,
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                MinVersion: tls.VersionTLS12,
            },
            // Disable HTTP/2 if needed for compatibility
            ForceAttemptHTTP2: false,
        },
    }
}
```

**Key Points**:
- 10 second timeout prevents hanging requests
- TLS 1.2 minimum for security
- Thread-safe (can be used concurrently)

### 3.2 Write Tests

In `internal/infrastructure/adapter/api/http_client_test.go`:

```go
package api

import (
    "net/http"
    "testing"
    "time"
)

func TestNewHTTPClient(t *testing.T) {
    client := NewHTTPClient()
    
    // Verify timeout
    if client.Timeout != 10*time.Second {
        t.Errorf("Timeout = %v, want 10s", client.Timeout)
    }
    
    // Verify transport exists
    if client.Transport == nil {
        t.Error("Transport is nil")
    }
    
    // Verify TLS config
    transport, ok := client.Transport.(*http.Transport)
    if !ok {
        t.Fatal("Transport is not *http.Transport")
    }
    
    if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
        t.Errorf("TLS MinVersion = %v, want TLS 1.2", transport.TLSClientConfig.MinVersion)
    }
}
```

**Deliverable**: ✅ HTTP client implemented and tested

**Time**: 20 minutes

---

## Step 4: Implement Response Parsing

**Objective**: Create response parsing and validation functions

**Why**: 
- Separates parsing logic from HTTP logic
- Makes testing easier
- Handles API-specific response format

**What to Do**:

### 4.1 Define Response Structures

In `internal/infrastructure/adapter/api/exchange_rate_provider.go`:

```go
package api

// exchangerateHostResponse represents the API response structure.
type exchangerateHostResponse struct {
    Success bool              `json:"success"`
    Base    string            `json:"base"`
    Date    string            `json:"date"`
    Rates   map[string]float64 `json:"rates"`
    Error   *apiError         `json:"error,omitempty"`
}

// apiError represents an API error response.
type apiError struct {
    Code int    `json:"code"`
    Type string `json:"type"`
    Info string `json:"info"`
}
```

### 4.2 Implement Parsing Functions

```go
// parseRateResponse parses a single rate response.
func parseRateResponse(resp *currencyAPIResponse, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // Validate response
    if resp.Error != "" {
        return nil, fmt.Errorf("api returned error: %s", resp.Error)
    }
    
    // Validate base currency matches
    if resp.Base != base.String() {
        return nil, fmt.Errorf("base currency mismatch: expected %s, got %s", base, resp.Base)
    }
    
    // Get rate for target currency
    rate, ok := resp.Rates[target.String()]
    if !ok {
        return nil, fmt.Errorf("target currency %s not found in response", target)
    }
    
    // Validate rate is positive
    if rate <= 0 {
        return nil, fmt.Errorf("invalid rate: %f (must be positive)", rate)
    }
    
    // Create domain entity
    return entity.NewExchangeRate(base, target, rate, time.Now(), false)
}

// parseAllRatesResponse parses an all-rates response.
func parseAllRatesResponse(resp *exchangerateHostResponse, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
    // Validate response
    if !resp.Success {
        return nil, fmt.Errorf("api returned error: %v", resp.Error)
    }
    
    // Validate base currency matches
    if resp.Base != base.String() {
        return nil, fmt.Errorf("base currency mismatch: expected %s, got %s", base, resp.Base)
    }
    
    // Convert rates map to entity slice
    rates := make([]*entity.ExchangeRate, 0, len(resp.Rates))
    for targetStr, rate := range resp.Rates {
        // Validate rate is positive
        if rate <= 0 {
            continue // Skip invalid rates
        }
        
        // Create currency code
        target, err := entity.NewCurrencyCode(targetStr)
        if err != nil {
            continue // Skip invalid currency codes
        }
        
        // Create entity
        rateEntity, err := entity.NewExchangeRate(base, target, rate, time.Now(), false)
        if err != nil {
            continue // Skip if entity creation fails
        }
        
        rates = append(rates, rateEntity)
    }
    
    return rates, nil
}
```

**Deliverable**: ✅ Response parsing functions implemented

**Time**: 30 minutes

---

## Step 5: Implement ExchangeRateProvider Adapter

**Objective**: Implement the `ExchangeRateProvider` interface

**Why**: This is the core adapter that connects domain to external API

**What to Do**:

### 5.1 Create Provider Struct

In `internal/infrastructure/adapter/api/exchange_rate_provider.go`:

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/misterfancybg/go-currenseen/internal/domain/entity"
    "github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

// CurrencyAPIProvider implements ExchangeRateProvider using Currency-api.
type CurrencyAPIProvider struct {
    client  *http.Client
    baseURL string
}

// NewCurrencyAPIProvider creates a new CurrencyAPIProvider.
//
// Parameters:
//   - client: HTTP client (can be real or mock for testing)
//   - baseURL: Base URL for the API (default: "https://api.fawazahmed0.currency-api.com/v1")
func NewCurrencyAPIProvider(client *http.Client, baseURL string) *CurrencyAPIProvider {
    if baseURL == "" {
        baseURL = "https://api.fawazahmed0.currency-api.com/v1"
    }
    return &CurrencyAPIProvider{
        client:  client,
        baseURL: baseURL,
    }
}

// FetchRate implements provider.ExchangeRateProvider.
func (p *CurrencyAPIProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // Build URL - Currency-api doesn't support symbols filter, so we fetch all and filter
    url := fmt.Sprintf("%s/latest?base=%s", p.baseURL, base.String())
    
    // Create request with context
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Execute request
    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("http request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Parse JSON
    var apiResp exchangerateHostResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    // Convert to domain entity
    return parseRateResponse(&apiResp, base, target)
}

// FetchAllRates implements provider.ExchangeRateProvider.
func (p *CurrencyAPIProvider) FetchAllRates(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
    // Build URL
    url := fmt.Sprintf("%s/latest?base=%s", p.baseURL, base.String())
    
    // Create request with context
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    // Execute request
    resp, err := p.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("http request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Check status code
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }
    
    // Read response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response: %w", err)
    }
    
    // Parse JSON
    var apiResp exchangerateHostResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    // Convert to domain entities
    rates, err := parseAllRatesResponse(&apiResp, base)
    if err != nil {
        return nil, err
    }
    
    // Return empty slice (not nil) if no rates
    if rates == nil {
        return []*entity.ExchangeRate{}, nil
    }
    
    return rates, nil
}

// Ensure CurrencyAPIProvider implements ExchangeRateProvider interface.
var _ provider.ExchangeRateProvider = (*CurrencyAPIProvider)(nil)
```

**Key Points**:
- Uses context for cancellation/timeout
- Proper error wrapping
- Validates HTTP status codes
- Returns domain entities (not API DTOs)
- Interface compliance check

**Deliverable**: ✅ ExchangeRateProvider adapter implemented

**Time**: 45 minutes

---

## Step 6: Implement Retry Logic

**Objective**: Add retry logic with exponential backoff

**Why**: 
- Handles transient network errors
- Improves reliability
- Demonstrates resilience patterns

**What to Do**:

### 6.1 Create Retry Configuration

In `internal/infrastructure/adapter/api/retry.go`:

```go
package api

import (
    "context"
    "errors"
    "fmt"
    "math"
    "net"
    "net/http"
    "time"
)

// RetryConfig holds retry configuration.
type RetryConfig struct {
    MaxAttempts      int           // Maximum number of retry attempts
    InitialBackoff   time.Duration // Initial backoff duration
    MaxBackoff       time.Duration // Maximum backoff duration
    BackoffMultiplier float64      // Backoff multiplier (e.g., 2.0 for exponential)
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts:      3,
        InitialBackoff:   100 * time.Millisecond,
        MaxBackoff:       5 * time.Second,
        BackoffMultiplier: 2.0,
    }
}

// isRetryableError checks if an error is retryable.
func isRetryableError(err error) bool {
    if err == nil {
        return false
    }
    
    // Check for network errors
    var netErr net.Error
    if errors.As(err, &netErr) {
        return netErr.Timeout() || netErr.Temporary()
    }
    
    // Check for context cancellation (not retryable)
    if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
        return false
    }
    
    return false
}

// isRetryableStatusCode checks if an HTTP status code is retryable.
func isRetryableStatusCode(code int) bool {
    // Retry on 5xx errors and 429 (Too Many Requests)
    return code >= 500 || code == http.StatusTooManyRequests
}

// calculateBackoff calculates the backoff duration for attempt n.
func calculateBackoff(config RetryConfig, attempt int) time.Duration {
    // Calculate exponential backoff: initial * (multiplier ^ attempt)
    backoff := float64(config.InitialBackoff) * math.Pow(config.BackoffMultiplier, float64(attempt))
    
    // Cap at max backoff
    if backoff > float64(config.MaxBackoff) {
        backoff = float64(config.MaxBackoff)
    }
    
    return time.Duration(backoff)
}

// RetryableFetchRate executes FetchRate with retry logic.
func RetryableFetchRate(
    ctx context.Context,
    provider provider.ExchangeRateProvider,
    base, target entity.CurrencyCode,
    config RetryConfig,
) (*entity.ExchangeRate, error) {
    var lastErr error
    
    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        // Check context before retry
        if ctx.Err() != nil {
            return nil, ctx.Err()
        }
        
        // Execute request
        rate, err := provider.FetchRate(ctx, base, target)
        if err == nil {
            return rate, nil
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryableError(err) {
            return nil, err // Don't retry non-retryable errors
        }
        
        // Don't sleep after last attempt
        if attempt < config.MaxAttempts-1 {
            backoff := calculateBackoff(config, attempt)
            time.Sleep(backoff)
        }
    }
    
    return nil, fmt.Errorf("max retry attempts (%d) exceeded: %w", config.MaxAttempts, lastErr)
}
```

**Key Points**:
- Exponential backoff (100ms, 200ms, 400ms, ...)
- Only retries retryable errors (network timeouts, 5xx errors)
- Respects context cancellation
- Configurable retry attempts and backoff

**Deliverable**: ✅ Retry logic implemented

**Time**: 45 minutes

---

## Step 7: Integrate Retry into Provider

**Objective**: Add retry logic to the provider implementation

**Why**: Makes the provider resilient to transient failures

**What to Do**:

### 7.1 Update Provider to Use Retry

Modify `ExchangeRateHostProvider` to include retry config and use retry logic:

```go
type ExchangeRateHostProvider struct {
    client     *http.Client
    baseURL    string
    retryConfig RetryConfig
}

func NewCurrencyAPIProvider(client *http.Client, baseURL string) *CurrencyAPIProvider {
    if baseURL == "" {
        baseURL = "https://api.fawazahmed0.currency-api.com/v1"
    }
    return &CurrencyAPIProvider{
        client:      client,
        baseURL:     baseURL,
        retryConfig: DefaultRetryConfig(),
    }
}

// FetchRate now uses retry logic internally
func (p *ExchangeRateHostProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
    // ... existing implementation with retry wrapper
}
```

**Alternative Approach**: Keep provider simple, add retry wrapper externally (better separation of concerns).

**Deliverable**: ✅ Retry integrated into provider

**Time**: 30 minutes

---

## Step 8: Create Provider Factory

**Objective**: Create factory for creating provider instances

**Why**: 
- Centralizes provider creation
- Enables multiple provider support (primary/fallback)
- Follows factory pattern

**What to Do**:

### 8.1 Create Factory Function

In `internal/infrastructure/adapter/api/provider_factory.go`:

```go
package api

import (
    "github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

// ProviderType represents the type of provider.
type ProviderType string

const (
    ProviderTypeCurrencyAPI ProviderType = "currency_api"
)

// ProviderConfig holds provider configuration.
type ProviderConfig struct {
    Type   ProviderType
    BaseURL string
    APIKey  string // For future use with APIs that require keys
}

// NewProvider creates a new ExchangeRateProvider based on configuration.
func NewProvider(config ProviderConfig) (provider.ExchangeRateProvider, error) {
    client := NewHTTPClient()
    
    switch config.Type {
    case ProviderTypeExchangeRateHost:
        return NewExchangeRateHostProvider(client, config.BaseURL), nil
    default:
        return nil, fmt.Errorf("unknown provider type: %s", config.Type)
    }
}

// NewDefaultProvider creates a provider with default configuration.
func NewDefaultProvider() provider.ExchangeRateProvider {
    config := ProviderConfig{
        Type:   ProviderTypeCurrencyAPI,
        BaseURL: "https://api.fawazahmed0.currency-api.com/v1",
    }
    provider, _ := NewProvider(config)
    return provider
}
```

**Deliverable**: ✅ Provider factory implemented

**Time**: 20 minutes

---

## Step 9: Write Unit Tests

**Objective**: Create comprehensive unit tests

**Why**: Ensures correctness and demonstrates testability

**What to Do**:

### 9.1 Test HTTP Client

Test timeout, TLS config, etc.

### 9.2 Test Response Parsing

Test valid responses, error responses, invalid data.

### 9.3 Test Provider with Mock HTTP Client

Use `httptest` package to mock HTTP responses.

### 9.4 Test Retry Logic

Test retry attempts, backoff calculation, non-retryable errors.

**Deliverable**: ✅ Unit tests written and passing

**Time**: 2 hours

---

## Step 10: Create Configuration

**Objective**: Create API configuration helper

**Why**: Centralizes API configuration (URLs, timeouts, etc.)

**What to Do**:

In `internal/infrastructure/config/api.go`:

```go
package config

import (
    "os"
    "time"
)

// APIConfig holds API configuration.
type APIConfig struct {
    BaseURL    string
    Timeout    time.Duration
    RetryAttempts int
}

// LoadAPIConfig loads API configuration from environment variables.
func LoadAPIConfig() APIConfig {
    baseURL := os.Getenv("EXCHANGE_RATE_API_URL")
    if baseURL == "" {
        baseURL = "https://api.fawazahmed0.currency-api.com/v1"
    }
    
    return APIConfig{
        BaseURL:      baseURL,
        Timeout:      10 * time.Second,
        RetryAttempts: 3,
    }
}
```

**Deliverable**: ✅ API configuration implemented

**Time**: 15 minutes

---

## Summary Checklist

Before considering Phase 4 complete:

- [ ] Exchange rate provider selected
- [ ] Directory structure created
- [ ] HTTP client implemented and tested
- [ ] Response parsing implemented
- [ ] ExchangeRateProvider adapter implemented
- [ ] Retry logic implemented
- [ ] Retry integrated into provider
- [ ] Provider factory implemented
- [ ] Unit tests written and passing
- [ ] Configuration implemented
- [ ] Code documented
- [ ] Code reviewed

---

## Estimated Total Time

- Step 1: 15 minutes
- Step 2: 5 minutes
- Step 3: 20 minutes
- Step 4: 30 minutes
- Step 5: 45 minutes
- Step 6: 45 minutes
- Step 7: 30 minutes
- Step 8: 20 minutes
- Step 9: 2 hours
- Step 10: 15 minutes

**Total**: ~6-7 hours

---

## Next Steps

After Phase 4 completion:
- Phase 5: Application Layer - Use Cases
- Phase 6: Circuit Breaker Implementation
- Phase 7: Lambda Handlers
