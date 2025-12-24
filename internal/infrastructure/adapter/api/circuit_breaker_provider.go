package api

import (
	"context"
	"fmt"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// CircuitBreakerProvider wraps an ExchangeRateProvider with circuit breaker protection.
//
// This wrapper:
// - Checks circuit breaker state before calling the underlying provider
// - Records success/failure based on provider call results
// - Returns ErrCircuitOpen when circuit is open
//
// This enables graceful degradation: when the circuit is open, use cases can
// fall back to cached (stale) data instead of failing completely.
type CircuitBreakerProvider struct {
	provider       provider.ExchangeRateProvider
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// NewCircuitBreakerProvider creates a new CircuitBreakerProvider.
//
// Parameters:
//   - provider: The underlying ExchangeRateProvider to wrap
//   - circuitBreaker: The circuit breaker instance
//
// Returns a new CircuitBreakerProvider that wraps the given provider.
func NewCircuitBreakerProvider(provider provider.ExchangeRateProvider, circuitBreaker *circuitbreaker.CircuitBreaker) *CircuitBreakerProvider {
	return &CircuitBreakerProvider{
		provider:       provider,
		circuitBreaker: circuitBreaker,
	}
}

// FetchRate implements provider.ExchangeRateProvider.
//
// This method:
// - Checks if the circuit breaker allows the request
// - Calls the underlying provider if allowed
// - Records success/failure based on the result
// - Returns ErrCircuitOpen if circuit is open
//
// Context cancellation: Returns error if ctx is cancelled or times out.
func (p *CircuitBreakerProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	// Check if circuit breaker allows the request
	if !p.circuitBreaker.Allow() {
		return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
	}

	// Call underlying provider
	rate, err := p.provider.FetchRate(ctx, base, target)

	// Record result in circuit breaker
	if err != nil {
		p.circuitBreaker.RecordFailure()
		return nil, err
	}

	p.circuitBreaker.RecordSuccess()
	return rate, nil
}

// FetchAllRates implements provider.ExchangeRateProvider.
//
// This method:
// - Checks if the circuit breaker allows the request
// - Calls the underlying provider if allowed
// - Records success/failure based on the result
// - Returns ErrCircuitOpen if circuit is open
//
// Context cancellation: Returns error if ctx is cancelled or times out.
func (p *CircuitBreakerProvider) FetchAllRates(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	// Check if circuit breaker allows the request
	if !p.circuitBreaker.Allow() {
		return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
	}

	// Call underlying provider
	rates, err := p.provider.FetchAllRates(ctx, base)

	// Record result in circuit breaker
	if err != nil {
		p.circuitBreaker.RecordFailure()
		return nil, err
	}

	p.circuitBreaker.RecordSuccess()
	return rates, nil
}

// Ensure CircuitBreakerProvider implements ExchangeRateProvider interface.
var _ provider.ExchangeRateProvider = (*CircuitBreakerProvider)(nil)
