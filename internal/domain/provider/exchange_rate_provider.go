package provider

import (
	"context"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ExchangeRateProvider defines the interface for fetching exchange rates from external APIs.
// This is a port in the Hexagonal Architecture pattern.
//
// Implementations of this interface should handle:
// - HTTP requests to external exchange rate APIs
// - Response parsing and validation
// - Error handling and retries (handled by adapter)
// - Rate limiting awareness
//
// Thread Safety:
// Implementations should be safe for concurrent use by multiple goroutines.
// However, external API rate limits may apply.
//
// Context Behavior:
// All methods respect context cancellation and timeouts. If ctx is cancelled or
// times out, implementations should return immediately with an appropriate error.
// Implementations should also respect context deadlines for HTTP requests.
type ExchangeRateProvider interface {
	// FetchRate retrieves the exchange rate for a specific currency pair from an external API.
	//
	// Returns a fresh ExchangeRate with:
	// - Stale flag set to false (rates from external APIs are always fresh)
	// - Timestamp set to the current time or API-provided timestamp
	// - Validated currency codes and rate value
	//
	// Returns an error if:
	// - The rate cannot be fetched (network error, API error, etc.)
	// - The response cannot be parsed
	// - The response fails validation
	// - The context is cancelled or times out
	//
	// Error types: Implementations should return domain-appropriate errors.
	// Infrastructure errors (network, HTTP) should be wrapped with context.
	//
	// Context cancellation: Returns error if ctx is cancelled or times out.
	// Implementations should use ctx for HTTP request timeouts.
	FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)

	// FetchAllRates retrieves all exchange rates for a base currency from an external API.
	//
	// Returns a map where:
	// - Keys are target currency codes (entity.CurrencyCode)
	// - Values are exchange rates (*entity.ExchangeRate)
	//
	// All returned rates will have:
	// - Stale flag set to false
	// - Timestamp set to the current time or API-provided timestamp
	// - Validated currency codes and rate values
	//
	// Returns an error if:
	// - The rates cannot be fetched (network error, API error, etc.)
	// - The response cannot be parsed
	// - The response fails validation
	// - The context is cancelled or times out
	//
	// Note: The map return type allows efficient lookups by target currency.
	// If no rates are available, returns an empty map (not an error).
	//
	// Context cancellation: Returns error if ctx is cancelled or times out.
	FetchAllRates(ctx context.Context, base entity.CurrencyCode) (map[entity.CurrencyCode]*entity.ExchangeRate, error)
}
