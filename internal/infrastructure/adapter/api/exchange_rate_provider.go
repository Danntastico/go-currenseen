package api

import (
	"fmt"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
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
