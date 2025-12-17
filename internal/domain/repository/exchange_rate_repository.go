package repository

import (
	"context"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ExchangeRateRepository defines the interface for exchange rate data access.
// This is a port in the Hexagonal Architecture pattern.
//
// Implementations of this interface should handle:
// - DynamoDB operations (GetItem, PutItem, Query)
// - TTL management
// - Error mapping (infrastructure errors → domain errors)
type ExchangeRateRepository interface {
	// Get retrieves an exchange rate for a specific currency pair.
	// Returns entity.ErrRateNotFound if the rate doesn't exist.
	Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)

	// Save stores an exchange rate with TTL.
	// If the rate already exists, it will be updated.
	Save(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error

	// GetByBase retrieves all exchange rates for a base currency.
	// Returns an empty slice if no rates are found (not an error).
	GetByBase(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)

	// Delete removes an exchange rate for a specific currency pair.
	// Returns entity.ErrRateNotFound if the rate doesn't exist.
	Delete(ctx context.Context, base, target entity.CurrencyCode) error

	// GetStale retrieves a stale (expired) exchange rate for fallback scenarios.
	// This is used when the external API is unavailable and we need to return
	// cached data even if it's expired.
	GetStale(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
}

