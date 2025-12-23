package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

// RetryConfig holds retry configuration.
type RetryConfig struct {
	MaxAttempts       int           // Maximum number of retry attempts
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffMultiplier float64       // Backoff multiplier (e.g., 2.0 for exponential)
}

// DefaultRetryConfig returns a default retry configuration.
//
// Default values:
// - MaxAttempts: 3 (initial attempt + 2 retries)
// - InitialBackoff: 100ms
// - MaxBackoff: 5s
// - BackoffMultiplier: 2.0 (exponential backoff)
//
// This results in backoff durations: 100ms, 200ms, 400ms, ...
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// isRetryableError checks if an error is retryable.
//
// Retryable errors:
// - Network timeout errors
// - Temporary network errors
//
// Non-retryable errors:
// - Context cancellation
// - Context deadline exceeded
// - Validation errors
// - 4xx HTTP errors (client errors)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation (not retryable)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Retry on timeout or temporary network errors
		return netErr.Timeout() || netErr.Temporary()
	}

	return false
}

// isRetryableStatusCode checks if an HTTP status code is retryable.
//
// Retryable status codes:
// - 5xx (server errors)
// - 429 (Too Many Requests)
//
// Non-retryable status codes:
// - 4xx (client errors, except 429)
// - 2xx (success)
// - 3xx (redirects)
func isRetryableStatusCode(code int) bool {
	// Retry on 5xx errors and 429 (Too Many Requests)
	return code >= 500 || code == http.StatusTooManyRequests
}

// calculateBackoff calculates the backoff duration for attempt n.
//
// Formula: initial * (multiplier ^ attempt)
// Example with initial=100ms, multiplier=2.0:
// - Attempt 0: 100ms * (2.0 ^ 0) = 100ms
// - Attempt 1: 100ms * (2.0 ^ 1) = 200ms
// - Attempt 2: 100ms * (2.0 ^ 2) = 400ms
//
// The backoff is capped at MaxBackoff to prevent excessive delays.
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
//
// This function:
// - Attempts to fetch the rate up to MaxAttempts times
// - Uses exponential backoff between retries
// - Only retries retryable errors (network timeouts, temporary errors)
// - Respects context cancellation
// - Returns the first successful result
//
// Returns an error if:
// - All retry attempts are exhausted
// - The error is not retryable
// - The context is cancelled or times out
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

// RetryableFetchAllRates executes FetchAllRates with retry logic.
//
// This function:
// - Attempts to fetch all rates up to MaxAttempts times
// - Uses exponential backoff between retries
// - Only retries retryable errors (network timeouts, temporary errors)
// - Respects context cancellation
// - Returns the first successful result
//
// Returns an error if:
// - All retry attempts are exhausted
// - The error is not retryable
// - The context is cancelled or times out
func RetryableFetchAllRates(
	ctx context.Context,
	provider provider.ExchangeRateProvider,
	base entity.CurrencyCode,
	config RetryConfig,
) ([]*entity.ExchangeRate, error) {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check context before retry
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Execute request
		rates, err := provider.FetchAllRates(ctx, base)
		if err == nil {
			return rates, nil
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
