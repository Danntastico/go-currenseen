package api

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

func TestNewCircuitBreakerProvider(t *testing.T) {
	mockProv := &mockProvider{}
	cb, _ := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())

	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	if wrapper == nil {
		t.Fatal("NewCircuitBreakerProvider() returned nil")
	}

	if wrapper.provider != mockProv {
		t.Error("provider not set correctly")
	}

	if wrapper.circuitBreaker != cb {
		t.Error("circuitBreaker not set correctly")
	}
}

func TestCircuitBreakerProvider_FetchRate_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	rate, _ := entity.NewExchangeRate(base, target, 0.85, time.Now(), false)

	mockProv := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			return rate, nil
		},
	}

	cb, _ := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()
	result, err := wrapper.FetchRate(ctx, base, target)

	if err != nil {
		t.Fatalf("FetchRate() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("FetchRate() returned nil rate")
	}

	// Verify circuit breaker recorded success
	if cb.State() != circuitbreaker.StateClosed {
		t.Errorf("Circuit breaker state = %v, want Closed", cb.State())
	}
}

func TestCircuitBreakerProvider_FetchRate_CircuitOpen(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	mockProv := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			return nil, errors.New("provider error")
		},
	}

	config := circuitbreaker.Config{
		FailureThreshold: 2,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := circuitbreaker.NewCircuitBreaker(config)
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()

	// Record failures to open the circuit
	_, _ = wrapper.FetchRate(ctx, base, target) // First failure
	_, _ = wrapper.FetchRate(ctx, base, target) // Second failure (opens circuit)

	// Circuit should now be open
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("Circuit breaker state = %v, want Open", cb.State())
	}

	// Next call should fail immediately with ErrCircuitOpen
	_, err := wrapper.FetchRate(ctx, base, target)
	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}

	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("Error = %v, want ErrCircuitOpen", err)
	}
}

func TestCircuitBreakerProvider_FetchRate_ProviderError(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	providerErr := errors.New("provider error")
	mockProv := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			return nil, providerErr
		},
	}

	cb, _ := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()
	_, err := wrapper.FetchRate(ctx, base, target)

	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}

	if err != providerErr {
		t.Errorf("Error = %v, want %v", err, providerErr)
	}

	// Circuit breaker should have recorded failure
	if cb.State() == circuitbreaker.StateOpen {
		// This is expected if failure threshold was reached
		// (depends on config, but with default config of 5, one failure shouldn't open)
	}
}

func TestCircuitBreakerProvider_FetchAllRates_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	rates := []*entity.ExchangeRate{}

	mockProv := &mockProvider{
		fetchAllRatesFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
			return rates, nil
		},
	}

	cb, _ := circuitbreaker.NewCircuitBreaker(circuitbreaker.DefaultConfig())
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()
	result, err := wrapper.FetchAllRates(ctx, base)

	if err != nil {
		t.Fatalf("FetchAllRates() error = %v, want nil", err)
	}

	if result == nil {
		t.Fatal("FetchAllRates() returned nil rates")
	}

	// Verify circuit breaker recorded success
	if cb.State() != circuitbreaker.StateClosed {
		t.Errorf("Circuit breaker state = %v, want Closed", cb.State())
	}
}

func TestCircuitBreakerProvider_FetchAllRates_CircuitOpen(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	mockProv := &mockProvider{
		fetchAllRatesFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
			return nil, errors.New("provider error")
		},
	}

	config := circuitbreaker.Config{
		FailureThreshold: 2,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := circuitbreaker.NewCircuitBreaker(config)
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()

	// Record failures to open the circuit
	_, _ = wrapper.FetchAllRates(ctx, base) // First failure
	_, _ = wrapper.FetchAllRates(ctx, base) // Second failure (opens circuit)

	// Circuit should now be open
	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("Circuit breaker state = %v, want Open", cb.State())
	}

	// Next call should fail immediately with ErrCircuitOpen
	_, err := wrapper.FetchAllRates(ctx, base)
	if err == nil {
		t.Fatal("FetchAllRates() error = nil, want error")
	}

	if !errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		t.Errorf("Error = %v, want ErrCircuitOpen", err)
	}
}

func TestCircuitBreakerProvider_HalfOpen_Recovery(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	rate, _ := entity.NewExchangeRate(base, target, 0.85, time.Now(), false)

	callCount := 0
	mockProv := &mockProvider{
		fetchRateFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
			callCount++
			if callCount <= 2 {
				// First two calls fail
				return nil, errors.New("provider error")
			}
			// Third call succeeds (after circuit opens and cooldown)
			return rate, nil
		},
	}

	config := circuitbreaker.Config{
		FailureThreshold: 2,
		CooldownDuration: 50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := circuitbreaker.NewCircuitBreaker(config)
	wrapper := NewCircuitBreakerProvider(mockProv, cb)

	ctx := context.Background()

	// Open the circuit
	_, _ = wrapper.FetchRate(ctx, base, target) // First failure
	_, _ = wrapper.FetchRate(ctx, base, target) // Second failure (opens circuit)

	if cb.State() != circuitbreaker.StateOpen {
		t.Fatalf("Circuit breaker state = %v, want Open", cb.State())
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to HalfOpen and succeed
	result, err := wrapper.FetchRate(ctx, base, target)
	if err != nil {
		t.Fatalf("FetchRate() error = %v, want nil (should succeed in HalfOpen)", err)
	}

	if result == nil {
		t.Fatal("FetchRate() returned nil rate")
	}

	// Circuit should now be Closed (recovered)
	if cb.State() != circuitbreaker.StateClosed {
		t.Errorf("Circuit breaker state = %v, want Closed (recovered)", cb.State())
	}
}
