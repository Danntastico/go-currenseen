package middleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrRateLimitExceeded is returned when the rate limit is exceeded.
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimiterConfig holds configuration for rate limiting.
type RateLimiterConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per minute per API key.
	RequestsPerMinute int
	// BurstSize is the maximum burst size (defaults to RequestsPerMinute if 0).
	BurstSize int
	// Enabled controls whether rate limiting is active.
	Enabled bool
}

// DefaultRateLimiterConfig returns a default rate limiter configuration.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 100,
		BurstSize:         10,
		Enabled:           true,
	}
}

// tokenBucket represents a token bucket for rate limiting.
type tokenBucket struct {
	capacity   int       // Maximum tokens
	tokens     int       // Current tokens
	lastRefill time.Time // Last time tokens were refilled
	refillRate float64   // Tokens per second
	mu         sync.Mutex
}

// newTokenBucket creates a new token bucket.
func newTokenBucket(capacity int, refillRate float64) *tokenBucket {
	return &tokenBucket{
		capacity:   capacity,
		tokens:     capacity, // Start with full bucket
		lastRefill: time.Now(),
		refillRate: refillRate,
	}
}

// take attempts to take a token from the bucket.
// Returns true if a token was available, false otherwise.
func (tb *tokenBucket) take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed * tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens = min(tb.capacity, tb.tokens+tokensToAdd)
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimiter implements rate limiting using token bucket algorithm.
type RateLimiter struct {
	buckets map[string]*tokenBucket
	config  RateLimiterConfig
	mu      sync.RWMutex
	cleanup *time.Ticker
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.BurstSize == 0 {
		config.BurstSize = config.RequestsPerMinute
	}

	rl := &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		config:  config,
	}

	// Start cleanup goroutine to remove old buckets (every 5 minutes)
	rl.cleanup = time.NewTicker(5 * time.Minute)
	go rl.cleanupBuckets()

	return rl
}

// cleanupBuckets periodically removes old buckets to prevent memory leaks.
func (rl *RateLimiter) cleanupBuckets() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		// In a production system, you might want to track last access time
		// and remove buckets that haven't been accessed in a while.
		// For simplicity, we'll keep all buckets here.
		rl.mu.Unlock()
	}
}

// Allow checks if a request is allowed for the given key.
//
// Returns:
// - true if the request is allowed
// - false if the rate limit is exceeded
// - error if rate limiting is disabled or key is empty
func (rl *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	if !rl.config.Enabled {
		return true, nil
	}

	if key == "" {
		return false, fmt.Errorf("rate limiter key cannot be empty")
	}

	// Get or create bucket for this key
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		// Calculate refill rate (tokens per second)
		refillRate := float64(rl.config.RequestsPerMinute) / 60.0
		bucket = newTokenBucket(rl.config.BurstSize, refillRate)
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	// Try to take a token
	if bucket.take() {
		return true, nil
	}

	return false, ErrRateLimitExceeded
}

// GetRemainingRequests returns the estimated number of remaining requests for a key.
// This is approximate and may not be exact due to concurrent access.
func (rl *RateLimiter) GetRemainingRequests(key string) int {
	if !rl.config.Enabled || key == "" {
		return -1 // Unknown
	}

	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.config.BurstSize
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Refill tokens to get accurate count
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	tokensToAdd := int(elapsed * bucket.refillRate)
	if tokensToAdd > 0 {
		bucket.tokens = min(bucket.capacity, bucket.tokens+tokensToAdd)
		bucket.lastRefill = now
	}

	return bucket.tokens
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


