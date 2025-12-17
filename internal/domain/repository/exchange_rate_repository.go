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
// - TTL management (storage-level, not validation)
// - Error mapping (infrastructure errors → domain errors)
//
// Thread Safety:
// Implementations should be safe for concurrent use by multiple goroutines.
// However, concurrent modifications to the same rate may result in race conditions.
//
// Context Behavior:
// All methods respect context cancellation. If ctx is cancelled, implementations
// should return immediately with an appropriate error.
type ExchangeRateRepository interface {
	// Get retrieves an exchange rate for a specific currency pair.
	//
	// Returns entity.ErrRateNotFound if the rate doesn't exist.
	//
	// Important: This method returns rates regardless of TTL expiration.
	// TTL checking should be performed by the caller (use cases) using
	// entity.ExchangeRate.IsExpired() or entity.ExchangeRate.IsValid().
	//
	// The repository is responsible only for storage/retrieval, not validation
	// of expiration. This allows use cases to decide whether to use expired
	// rates for fallback scenarios.
	//
	// Context cancellation: Returns error if ctx is cancelled.
	Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)

	// Save stores an exchange rate with TTL.
	//
	// If the rate already exists, it will be updated (upsert behavior).
	//
	// The ttl parameter is used for storage-level TTL management (e.g., DynamoDB TTL).
	// The repository implementation should store the TTL but not validate expiration
	// on retrieval. Expiration validation is the responsibility of use cases.
	//
	// Context cancellation: Returns error if ctx is cancelled.
	Save(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error

	// GetByBase retrieves all exchange rates for a base currency.
	//
	// Returns an empty slice (not nil) if no rates are found. This is not an error.
	//
	// Like Get(), this method returns rates regardless of TTL expiration.
	// The caller should check expiration if needed.
	//
	// Context cancellation: Returns error if ctx is cancelled.
	GetByBase(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)

	// Delete removes an exchange rate for a specific currency pair.
	//
	// Returns entity.ErrRateNotFound if the rate doesn't exist.
	//
	// Context cancellation: Returns error if ctx is cancelled.
	Delete(ctx context.Context, base, target entity.CurrencyCode) error

	// GetStale retrieves a stale (expired) exchange rate for fallback scenarios.
	//
	// This method is specifically designed for fallback scenarios when the external
	// API is unavailable. It retrieves cached data even if it's expired.
	//
	// Returns entity.ErrRateNotFound if no rate exists (even if expired).
	//
	// Note: This method may return the same data as Get() if the implementation
	// doesn't filter by TTL. The distinction is semantic - this method explicitly
	// indicates the caller wants stale data for fallback purposes.
	//
	// Context cancellation: Returns error if ctx is cancelled.
	GetStale(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
}
