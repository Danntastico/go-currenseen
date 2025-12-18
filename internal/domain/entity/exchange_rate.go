package entity

import (
	"fmt"
	"math"
	"time"
)

// ExchangeRate represents an exchange rate between two currencies.
// It is the core domain entity for the currency exchange rate service.
type ExchangeRate struct {
	Base      CurrencyCode
	Target    CurrencyCode
	Rate      float64
	Timestamp time.Time
	Stale     bool // Indicates if the rate is stale (from cache fallback)
}

// This is a constructor function, using the Constructor/Factory pattern
// NewExchangeRate creates a new ExchangeRate with validation.
// Returns an error if any field is invalid.
// The stale parameter indicates if the rate is stale (from cache fallback).
func NewExchangeRate(base, target CurrencyCode, rate float64, timestamp time.Time, stale bool) (*ExchangeRate, error) {
	if err := validateExchangeRate(base, target, rate, timestamp); err != nil {
		return nil, err
	}
	// Address-of operator, returns the memory address of the struct
	return &ExchangeRate{
		Base:      base,
		Target:    target,
		Rate:      rate,
		Timestamp: timestamp,
		Stale:     stale,
	}, nil
}

// NewStaleExchangeRate creates a new ExchangeRate marked as stale.
// This is used when returning cached data as a fallback.
// Deprecated: Use NewExchangeRate with stale=true instead.
func NewStaleExchangeRate(base, target CurrencyCode, rate float64, timestamp time.Time) (*ExchangeRate, error) {
	return NewExchangeRate(base, target, rate, timestamp, true)
}

// validateExchangeRate validates all fields of an ExchangeRate.
func validateExchangeRate(base, target CurrencyCode, rate float64, timestamp time.Time) error {
	if !base.IsValid() {
		return fmt.Errorf("%w: base currency %q", ErrInvalidCurrencyCode, base)
	}

	if !target.IsValid() {
		return fmt.Errorf("%w: target currency %q", ErrInvalidCurrencyCode, target)
	}

	if base.Equal(target) {
		return fmt.Errorf("%w: base=%q, target=%q", ErrCurrencyCodeMismatch, base, target)
	}

	// Validate rate: must be positive, finite, and not NaN
	if rate <= 0 || math.IsInf(rate, 0) || math.IsNaN(rate) {
		return fmt.Errorf("%w: rate must be positive and finite, got %f", ErrInvalidExchangeRate, rate)
	}

	if timestamp.IsZero() {
		return fmt.Errorf("%w: timestamp cannot be zero", ErrInvalidTimestamp)
	}

	// Timestamp should not be in the future (with small tolerance for clock skew)
	maxFutureTime := time.Now().Add(5 * time.Minute)
	if timestamp.After(maxFutureTime) {
		return fmt.Errorf("%w: timestamp cannot be in the future, got %v", ErrInvalidTimestamp, timestamp)
	}

	return nil
}

// IsExpired checks if the exchange rate is expired based on the given TTL duration.
// Returns true if the current time is at or after the expiration time (timestamp + TTL).
// Returns false if TTL is zero or negative (no expiration).
func (e *ExchangeRate) IsExpired(ttl time.Duration) bool {
	if ttl <= 0 {
		return false // No expiration if TTL is zero or negative
	}

	expirationTime := e.Timestamp.Add(ttl)
	// Use !Before() to include boundary: "not before" = "after or equal"
	return !time.Now().Before(expirationTime)
}

// Age returns the age of the exchange rate.
func (e *ExchangeRate) Age() time.Duration {
	return time.Since(e.Timestamp)
}

// IsValid checks if the exchange rate is still valid (not expired) for the given TTL.
func (e *ExchangeRate) IsValid(ttl time.Duration) bool {
	return !e.IsExpired(ttl)
}
