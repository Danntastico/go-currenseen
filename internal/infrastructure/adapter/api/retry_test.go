package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", config.MaxAttempts)
	}

	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", config.InitialBackoff)
	}

	if config.MaxBackoff != 5*time.Second {
		t.Errorf("MaxBackoff = %v, want 5s", config.MaxBackoff)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("BackoffMultiplier = %f, want 2.0", config.BackoffMultiplier)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
			want: false,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: false,
		},
		{
			name: "network timeout error",
			err:  &net.DNSError{Err: "timeout", IsTimeout: true},
			want: true,
		},
		{
			name: "temporary network error",
			err:  &net.DNSError{Err: "temporary", IsTemporary: true},
			want: true,
		},
		{
			name: "non-timeout network error",
			err:  &net.DNSError{Err: "other", IsTimeout: false, IsTemporary: false},
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRetryableStatusCode(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"200 OK", http.StatusOK, false},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"404 Not Found", http.StatusNotFound, false},
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"504 Gateway Timeout", http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableStatusCode(tt.code)
			if got != tt.want {
				t.Errorf("isRetryableStatusCode(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{"attempt 0", 0, 100 * time.Millisecond},
		{"attempt 1", 1, 200 * time.Millisecond},
		{"attempt 2", 2, 400 * time.Millisecond},
		{"attempt 3", 3, 800 * time.Millisecond},
		{"attempt 4", 4, 1600 * time.Millisecond},
		{"attempt 10 (capped)", 10, 5 * time.Second}, // Should be capped at MaxBackoff
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateBackoff(config, tt.attempt)
			if got != tt.want {
				t.Errorf("calculateBackoff(config, %d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

// mockProvider is defined in test_helpers.go

func TestRetryableFetchRate_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	rate, _ := entity.NewExchangeRate(base, target, 0.85, time.Now(), false)

	mock := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			return rate, nil
		},
	}

	config := DefaultRetryConfig()
	ctx := context.Background()

	result, err := RetryableFetchRate(ctx, mock, base, target, config)

	if err != nil {
		t.Fatalf("RetryableFetchRate() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("RetryableFetchRate() returned nil rate")
	}

	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should succeed on first attempt)", mock.callCount)
	}
}

func TestRetryableFetchRate_RetryableError_SucceedsOnRetry(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	rate, _ := entity.NewExchangeRate(base, target, 0.85, time.Now(), false)

	attempt := 0
	mock := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			attempt++
			if attempt == 1 {
				// First attempt fails with retryable error
				return nil, &net.DNSError{Err: "timeout", IsTimeout: true}
			}
			// Second attempt succeeds
			return rate, nil
		},
	}

	config := DefaultRetryConfig()
	ctx := context.Background()

	result, err := RetryableFetchRate(ctx, mock, base, target, config)

	if err != nil {
		t.Fatalf("RetryableFetchRate() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("RetryableFetchRate() returned nil rate")
	}

	if mock.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (should retry once)", mock.callCount)
	}
}

func TestRetryableFetchRate_NonRetryableError(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	mock := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			// Non-retryable error (context canceled)
			return nil, context.Canceled
		},
	}

	config := DefaultRetryConfig()
	ctx := context.Background()

	_, err := RetryableFetchRate(ctx, mock, base, target, config)

	if err == nil {
		t.Fatal("RetryableFetchRate() error = nil, want error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Error = %v, want context.Canceled", err)
	}

	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should not retry non-retryable errors)", mock.callCount)
	}
}

func TestRetryableFetchRate_MaxAttemptsExceeded(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	mock := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			// Always return retryable error
			return nil, &net.DNSError{Err: "timeout", IsTimeout: true}
		},
	}

	config := RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond, // Short backoff for testing
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	_, err := RetryableFetchRate(ctx, mock, base, target, config)

	if err == nil {
		t.Fatal("RetryableFetchRate() error = nil, want error")
	}

	if mock.callCount != 3 {
		t.Errorf("callCount = %d, want 3 (should exhaust all attempts)", mock.callCount)
	}
}

func TestRetryableFetchRate_ContextCancellation(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	mock := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			return nil, &net.DNSError{Err: "timeout", IsTimeout: true}
		},
	}

	config := RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    50 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := RetryableFetchRate(ctx, mock, base, target, config)

	if err == nil {
		t.Fatal("RetryableFetchRate() error = nil, want error")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Error = %v, want context.Canceled", err)
	}
}

func TestRetryableFetchAllRates_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	rates := []*entity.ExchangeRate{}

	mock := &mockProvider{
		fetchAllRatesFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
			return rates, nil
		},
	}

	config := DefaultRetryConfig()
	ctx := context.Background()

	result, err := RetryableFetchAllRates(ctx, mock, base, config)

	if err != nil {
		t.Fatalf("RetryableFetchAllRates() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("RetryableFetchAllRates() returned nil rates")
	}

	if mock.callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should succeed on first attempt)", mock.callCount)
	}
}

func TestRetryableFetchAllRates_RetryableError_SucceedsOnRetry(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	rates := []*entity.ExchangeRate{}

	attempt := 0
	mock := &mockProvider{
		fetchAllRatesFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
			attempt++
			if attempt == 1 {
				// First attempt fails with retryable error
				return nil, &net.DNSError{Err: "timeout", IsTimeout: true}
			}
			// Second attempt succeeds
			return rates, nil
		},
	}

	config := DefaultRetryConfig()
	ctx := context.Background()

	result, err := RetryableFetchAllRates(ctx, mock, base, config)

	if err != nil {
		t.Fatalf("RetryableFetchAllRates() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("RetryableFetchAllRates() returned nil rates")
	}

	if mock.callCount != 2 {
		t.Errorf("callCount = %d, want 2 (should retry once)", mock.callCount)
	}
}

func TestRetryableFetchAllRates_MaxAttemptsExceeded(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	mock := &mockProvider{
		fetchAllRatesFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
			// Always return retryable error
			return nil, &net.DNSError{Err: "timeout", IsTimeout: true}
		},
	}

	config := RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond, // Short backoff for testing
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	_, err := RetryableFetchAllRates(ctx, mock, base, config)

	if err == nil {
		t.Fatal("RetryableFetchAllRates() error = nil, want error")
	}

	if mock.callCount != 3 {
		t.Errorf("callCount = %d, want 3 (should exhaust all attempts)", mock.callCount)
	}
}
