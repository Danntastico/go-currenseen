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
type ExchangeRateProvider interface {
	// FetchRate retrieves the exchange rate for a specific currency pair.
	// Returns an error if the rate cannot be fetched.
	FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)

	// FetchAllRates retrieves all exchange rates for a base currency.
	// Returns a map where keys are target currency codes and values are exchange rates.
	// Returns an error if the rates cannot be fetched.
	FetchAllRates(ctx context.Context, base entity.CurrencyCode) (map[entity.CurrencyCode]*entity.ExchangeRate, error)
}


