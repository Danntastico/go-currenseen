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

// currencyAPIResponse represents the Currency-api response structure.
// Currency-api returns rates in the format:
//
//	{
//	  "date": "2024-01-15",
//	  "base": "USD",
//	  "rates": {
//	    "EUR": 0.85,
//	    "GBP": 0.75
//	  }
//	}
//
// Or on error:
//
//	{
//	  "error": "Invalid base currency"
//	}
type currencyAPIResponse struct {
	Date  string             `json:"date"`
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
	Error string             `json:"error,omitempty"`
}

// parseRateResponse parses a single rate response from Currency-api.
//
// This function:
// - Validates the response doesn't contain an error
// - Validates the base currency matches the expected base
// - Extracts the rate for the target currency
// - Validates the rate is positive
// - Creates a domain entity with the current timestamp and stale=false
//
// Returns an error if:
// - The API returned an error
// - Base currency mismatch
// - Target currency not found in response
// - Rate is invalid (non-positive)
// - Entity creation fails
func parseRateResponse(resp *currencyAPIResponse, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// Validate response doesn't contain an error
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

	// Validate rate is positive (entity validation will also check this, but fail fast here)
	if rate <= 0 {
		return nil, fmt.Errorf("invalid rate: %f (must be positive)", rate)
	}

	// Create domain entity
	// Note: Currency-api doesn't provide timestamp in response, so we use current time
	// Stale is false because rates from external APIs are always fresh
	return entity.NewExchangeRate(base, target, rate, time.Now(), false)
}

// parseAllRatesResponse parses an all-rates response from Currency-api.
//
// This function:
// - Validates the response doesn't contain an error
// - Validates the base currency matches the expected base
// - Converts the rates map to a slice of domain entities
// - Skips invalid rates or currency codes (graceful degradation)
// - Returns an empty slice if no valid rates are found (not an error)
//
// Returns an error if:
// - The API returned an error
// - Base currency mismatch
//
// Note: Invalid rates or currency codes are skipped (not returned as errors)
// to allow partial success when some rates are valid.
func parseAllRatesResponse(resp *currencyAPIResponse, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	// Validate response doesn't contain an error
	if resp.Error != "" {
		return nil, fmt.Errorf("api returned error: %s", resp.Error)
	}

	// Validate base currency matches
	if resp.Base != base.String() {
		return nil, fmt.Errorf("base currency mismatch: expected %s, got %s", base, resp.Base)
	}

	// Convert rates map to entity slice
	// Pre-allocate with capacity for better performance
	rates := make([]*entity.ExchangeRate, 0, len(resp.Rates))

	for targetStr, rate := range resp.Rates {
		// Skip invalid rates (non-positive)
		if rate <= 0 {
			continue
		}

		// Create currency code (validates format)
		target, err := entity.NewCurrencyCode(targetStr)
		if err != nil {
			// Skip invalid currency codes (graceful degradation)
			continue
		}

		// Skip if target equals base (entity validation would reject this)
		if target.Equal(base) {
			continue
		}

		// Create entity (includes full validation)
		rateEntity, err := entity.NewExchangeRate(base, target, rate, time.Now(), false)
		if err != nil {
			// Skip if entity creation fails (graceful degradation)
			continue
		}

		rates = append(rates, rateEntity)
	}

	// Return empty slice (not nil) if no rates found
	// This is consistent with repository.GetByBase() behavior
	return rates, nil
}

// CurrencyAPIProvider implements ExchangeRateProvider using Currency-api.
// This is an adapter in the Hexagonal Architecture pattern, connecting the domain layer
// to the external Currency-api infrastructure.
type CurrencyAPIProvider struct {
	client  *http.Client
	baseURL string
}

// NewCurrencyAPIProvider creates a new CurrencyAPIProvider.
//
// Parameters:
//   - client: HTTP client (can be real or mock for testing)
//   - baseURL: Base URL for the API (default: "https://api.fawazahmed0.currency-api.com/v1")
//
// Returns a new CurrencyAPIProvider instance.
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
//
// This method:
// - Builds the API URL for fetching all rates for the base currency
// - Makes an HTTP GET request with context support
// - Validates the HTTP response status code
// - Parses the JSON response
// - Extracts and returns the rate for the target currency
//
// Context cancellation: Returns error if ctx is cancelled or times out.
// The HTTP client respects the context deadline for request timeout.
func (p *CurrencyAPIProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// Check context before starting operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build URL - Currency-api doesn't support symbols filter, so we fetch all and filter
	url := fmt.Sprintf("%s/latest?base=%s", p.baseURL, base.String())

	// Create request with context (enables cancellation and timeout)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	// This defer cannot be called before the if statement because if the resp fails, the defer will panic because the resp.Body is nil
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
	var apiResp currencyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to domain entity using parsing function
	return parseRateResponse(&apiResp, base, target)
}

// FetchAllRates implements provider.ExchangeRateProvider.
//
// This method:
// - Builds the API URL for fetching all rates for the base currency
// - Makes an HTTP GET request with context support
// - Validates the HTTP response status code
// - Parses the JSON response
// - Converts all rates to domain entities
// - Returns empty slice (not nil) if no rates are found
//
// Context cancellation: Returns error if ctx is cancelled or times out.
// The HTTP client respects the context deadline for request timeout.
func (p *CurrencyAPIProvider) FetchAllRates(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	// Check context before starting operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build URL
	url := fmt.Sprintf("%s/latest?base=%s", p.baseURL, base.String())

	// Create request with context (enables cancellation and timeout)
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
	var apiResp currencyAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to domain entities using parsing function
	rates, err := parseAllRatesResponse(&apiResp, base)
	if err != nil {
		return nil, err
	}

	// Return empty slice (not nil) if no rates
	// This is consistent with repository.GetByBase() behavior
	if rates == nil {
		return []*entity.ExchangeRate{}, nil
	}

	return rates, nil
}

// Ensure CurrencyAPIProvider implements ExchangeRateProvider interface.
// This compile-time check ensures we've implemented all required methods.
var _ provider.ExchangeRateProvider = (*CurrencyAPIProvider)(nil)
