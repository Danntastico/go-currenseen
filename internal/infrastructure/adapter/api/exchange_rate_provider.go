package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

// currencyAPIResponse represents the new Exchange-api response structure.
// The new API (migrated from currency-api) returns rates in the format:
//
//	{
//	  "date": "2024-01-15",
//	  "usd": {
//	    "eur": 0.85,
//	    "gbp": 0.75
//	  }
//	}
//
// Note: Currency codes in the response are lowercase, and rates are nested
// under the base currency code (also lowercase).
type currencyAPIResponse struct {
	Date  string                        `json:"date"`
	Rates map[string]map[string]float64 `json:"-"` // Dynamic: key is base currency (lowercase)
	// We'll use a custom unmarshaler or access rates dynamically
}

// UnmarshalJSON implements custom JSON unmarshaling for the new API format.
// The new API nests rates under the base currency code (lowercase).
func (r *currencyAPIResponse) UnmarshalJSON(data []byte) error {
	// First, unmarshal into a map to handle dynamic base currency key
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Extract date
	if dateVal, ok := raw["date"].(string); ok {
		r.Date = dateVal
	}

	// Initialize rates map
	r.Rates = make(map[string]map[string]float64)

	// Find the currency code key (everything except "date")
	for key, value := range raw {
		if key == "date" {
			continue
		}
		// This should be the base currency code (lowercase)
		if ratesMap, ok := value.(map[string]interface{}); ok {
			convertedRates := make(map[string]float64)
			for targetKey, rateVal := range ratesMap {
				if rate, ok := rateVal.(float64); ok {
					convertedRates[targetKey] = rate
				}
			}
			r.Rates[key] = convertedRates
		}
	}

	return nil
}

// parseRateResponse parses a single rate response from the new Exchange-api.
//
// This function:
// - Validates the response structure
// - Validates the base currency matches (case-insensitive)
// - Extracts the rate for the target currency (case-insensitive)
// - Validates the rate is positive
// - Creates a domain entity with the current timestamp and stale=false
//
// Returns an error if:
// - The API returned an error
// - Base currency not found in response
// - Target currency not found in response
// - Rate is invalid (non-positive)
// - Entity creation fails
func parseRateResponse(resp *currencyAPIResponse, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// Get base currency code in lowercase (API uses lowercase)
	baseLower := strings.ToLower(base.String())

	// Find rates for the base currency
	baseRates, ok := resp.Rates[baseLower]
	if !ok {
		return nil, fmt.Errorf("base currency %s not found in response", base)
	}

	// Get target currency code in lowercase (API uses lowercase)
	targetLower := strings.ToLower(target.String())

	// Get rate for target currency
	rate, ok := baseRates[targetLower]
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

// parseAllRatesResponse parses an all-rates response from the new Exchange-api.
//
// This function:
// - Validates the response structure
// - Validates the base currency matches (case-insensitive)
// - Converts the rates map to a slice of domain entities
// - Skips invalid rates or currency codes (graceful degradation)
// - Returns an empty slice if no valid rates are found (not an error)
//
// Returns an error if:
// - The API returned an error
// - Base currency not found in response
//
// Note: Invalid rates or currency codes are skipped (not returned as errors)
// to allow partial success when some rates are valid.
func parseAllRatesResponse(resp *currencyAPIResponse, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	// Get base currency code in lowercase (API uses lowercase)
	baseLower := strings.ToLower(base.String())

	// Find rates for the base currency
	baseRates, ok := resp.Rates[baseLower]
	if !ok {
		return nil, fmt.Errorf("base currency %s not found in response", base)
	}

	// Convert rates map to entity slice
	// Pre-allocate with capacity for better performance
	rates := make([]*entity.ExchangeRate, 0, len(baseRates))

	for targetStr, rate := range baseRates {
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
	client      *http.Client
	baseURL     string
	fallbackURL string // Fallback URL for high availability
}

// NewCurrencyAPIProvider creates a new CurrencyAPIProvider.
//
// Parameters:
//   - client: HTTP client (can be real or mock for testing)
//   - baseURL: Base URL for the API (default: "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1")
//
// Returns a new CurrencyAPIProvider instance.
//
// Note: The API has been migrated from currency-api to exchange-api.
// The new API uses a different URL structure and response format.
func NewCurrencyAPIProvider(client *http.Client, baseURL string) *CurrencyAPIProvider {
	if baseURL == "" {
		// New API URL: uses jsDelivr CDN (primary)
		baseURL = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1"
	}
	// Fallback URL: Cloudflare Pages (as recommended by API docs)
	fallbackURL := "https://latest.currency-api.pages.dev/v1"
	return &CurrencyAPIProvider{
		client:      client,
		baseURL:     baseURL,
		fallbackURL: fallbackURL,
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
	fmt.Println("FetchRate called with base: ", base, " and target: ", target)
	// Check context before starting operation
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Build URL - New API format: /currencies/{baseCurrency}.json
	// Currency codes must be lowercase in the URL
	baseLower := strings.ToLower(base.String())
	path := fmt.Sprintf("/currencies/%s.json", baseLower)

	// Try primary URL first, then fallback
	urls := []string{
		fmt.Sprintf("%s%s", p.baseURL, path),
		fmt.Sprintf("%s%s", p.fallbackURL, path),
	}

	var lastErr error
	for i, url := range urls {
		fmt.Printf("[CurrencyAPIProvider] Attempting request %d/%d to: %s\n", i+1, len(urls), url)

		// Create request with context (enables cancellation and timeout)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			// Log error but try fallback
			fmt.Printf("[CurrencyAPIProvider] Request failed: %v\n", err)
			lastErr = fmt.Errorf("http request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		// Parse JSON
		var apiResp currencyAPIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		// Success! Convert to domain entity
		fmt.Printf("[CurrencyAPIProvider] Successfully fetched from: %s\n", url)
		return parseRateResponse(&apiResp, base, target)
	}

	// All URLs failed
	return nil, fmt.Errorf("all API endpoints failed, last error: %w", lastErr)
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

	// Build URL - New API format: /currencies/{baseCurrency}.json
	// Currency codes must be lowercase in the URL
	baseLower := strings.ToLower(base.String())
	path := fmt.Sprintf("/currencies/%s.json", baseLower)

	// Try primary URL first, then fallback
	urls := []string{
		fmt.Sprintf("%s%s", p.baseURL, path),
		fmt.Sprintf("%s%s", p.fallbackURL, path),
	}

	var lastErr error
	for i, url := range urls {
		fmt.Printf("[CurrencyAPIProvider] Attempting request %d/%d to: %s\n", i+1, len(urls), url)

		// Create request with context (enables cancellation and timeout)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Execute request
		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("http request failed: %w", err)
			continue
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		// Parse JSON
		var apiResp currencyAPIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			lastErr = fmt.Errorf("failed to parse response: %w", err)
			continue
		}

		// Success! Convert to domain entities
		fmt.Printf("[CurrencyAPIProvider] Successfully fetched from: %s\n", url)
		rates, err := parseAllRatesResponse(&apiResp, base)
		if err != nil {
			lastErr = err
			continue
		}

		// Return empty slice (not nil) if no rates
		// This is consistent with repository.GetByBase() behavior
		if rates == nil {
			return []*entity.ExchangeRate{}, nil
		}

		return rates, nil
	}

	// All URLs failed
	return nil, fmt.Errorf("all API endpoints failed, last error: %w", lastErr)
}

// Ensure CurrencyAPIProvider implements ExchangeRateProvider interface.
// This compile-time check ensures we've implemented all required methods.
var _ provider.ExchangeRateProvider = (*CurrencyAPIProvider)(nil)
