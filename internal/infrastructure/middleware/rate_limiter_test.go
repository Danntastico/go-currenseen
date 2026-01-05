package middleware

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Take(t *testing.T) {
	// Create a bucket with capacity 10 and refill rate of 1 token per second
	bucket := newTokenBucket(10, 1.0)

	// Should be able to take 10 tokens immediately
	for i := 0; i < 10; i++ {
		if !bucket.take() {
			t.Errorf("expected to be able to take token %d", i+1)
		}
	}

	// Should not be able to take more tokens immediately
	if bucket.take() {
		t.Error("expected to not be able to take token after bucket is empty")
	}

	// Wait for tokens to refill
	time.Sleep(1100 * time.Millisecond)

	// Should be able to take at least 1 token after refill
	if !bucket.take() {
		t.Error("expected to be able to take token after refill")
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	tests := []struct {
		name        string
		config      RateLimiterConfig
		key         string
		requests    int
		expectedErr error
	}{
		{
			name: "rate limiting disabled",
			config: RateLimiterConfig{
				Enabled: false,
			},
			key:         "test-key",
			requests:    1000,
			expectedErr: nil,
		},
		{
			name: "allow requests within limit",
			config: RateLimiterConfig{
				Enabled:           true,
				RequestsPerMinute: 10,
				BurstSize:         10,
			},
			key:         "test-key",
			requests:    10,
			expectedErr: nil,
		},
		{
			name: "reject requests over limit",
			config: RateLimiterConfig{
				Enabled:           true,
				RequestsPerMinute: 5,
				BurstSize:         5,
			},
			key:         "test-key",
			requests:    10,
			expectedErr: ErrRateLimitExceeded,
		},
		{
			name: "empty key",
			config: RateLimiterConfig{
				Enabled: true,
			},
			key:         "",
			requests:    1,
			expectedErr: nil, // Will return error from Allow, not ErrRateLimitExceeded
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.config)
			defer limiter.cleanup.Stop()

			allowedCount := 0
			rejectedCount := 0

			for i := 0; i < tt.requests; i++ {
				allowed, err := limiter.Allow(context.Background(), tt.key)
				if err != nil {
					if tt.key == "" {
						// Empty key should return error
						if err == nil {
							t.Error("expected error for empty key")
						}
						return
					}
					if err == ErrRateLimitExceeded {
						rejectedCount++
					} else {
						t.Errorf("unexpected error: %v", err)
					}
				} else if allowed {
					allowedCount++
				} else {
					rejectedCount++
				}
			}

			if tt.config.Enabled && tt.key != "" {
				if allowedCount > tt.config.BurstSize {
					t.Errorf("expected at most %d allowed requests, got %d", tt.config.BurstSize, allowedCount)
				}
				if tt.requests > tt.config.BurstSize && rejectedCount == 0 {
					t.Error("expected some requests to be rejected when over limit")
				}
			}
		})
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	config := RateLimiterConfig{
		Enabled:           true,
		RequestsPerMinute: 100,
		BurstSize:         100,
	}

	limiter := NewRateLimiter(config)
	defer limiter.cleanup.Stop()

	key := "concurrent-key"
	numGoroutines := 10
	requestsPerGoroutine := 20

	var wg sync.WaitGroup
	allowedCount := 0
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				allowed, err := limiter.Allow(context.Background(), key)
				if err == nil && allowed {
					mu.Lock()
					allowedCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Should not exceed burst size
	if allowedCount > config.BurstSize {
		t.Errorf("expected at most %d allowed requests, got %d", config.BurstSize, allowedCount)
	}

	// Should allow at least some requests
	if allowedCount == 0 {
		t.Error("expected at least some requests to be allowed")
	}
}

func TestRateLimiter_GetRemainingRequests(t *testing.T) {
	config := RateLimiterConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		BurstSize:         10,
	}

	limiter := NewRateLimiter(config)
	defer limiter.cleanup.Stop()

	key := "test-key"

	// Initially should have full bucket
	remaining := limiter.GetRemainingRequests(key)
	if remaining != config.BurstSize {
		t.Errorf("expected %d remaining requests initially, got %d", config.BurstSize, remaining)
	}

	// Make some requests
	for i := 0; i < 5; i++ {
		_, _ = limiter.Allow(context.Background(), key)
	}

	// Should have fewer remaining
	remaining = limiter.GetRemainingRequests(key)
	if remaining >= config.BurstSize {
		t.Errorf("expected fewer than %d remaining requests after 5 requests, got %d", config.BurstSize, remaining)
	}

	// Disabled limiter should return -1
	limiter.config.Enabled = false
	remaining = limiter.GetRemainingRequests(key)
	if remaining != -1 {
		t.Errorf("expected -1 for disabled limiter, got %d", remaining)
	}
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	if config.RequestsPerMinute != 100 {
		t.Errorf("expected RequestsPerMinute to be 100, got %d", config.RequestsPerMinute)
	}
	if config.BurstSize != 10 {
		t.Errorf("expected BurstSize to be 10, got %d", config.BurstSize)
	}
	if !config.Enabled {
		t.Error("expected Enabled to be true")
	}
}
