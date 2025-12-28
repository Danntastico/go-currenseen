package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
	"github.com/misterfancybg/go-currenseen/internal/domain/repository"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// GetExchangeRateUseCase handles the use case for getting an exchange rate for a currency pair.
// This implements UC1 from the specification.
type GetExchangeRateUseCase struct {
	repository repository.ExchangeRateRepository
	provider   provider.ExchangeRateProvider
	cacheTTL   time.Duration // TTL for cached rates
}

// NewGetExchangeRateUseCase creates a new GetExchangeRateUseCase with dependency injection.
func NewGetExchangeRateUseCase(
	repo repository.ExchangeRateRepository,
	prov provider.ExchangeRateProvider,
	cacheTTL time.Duration,
) *GetExchangeRateUseCase {
	return &GetExchangeRateUseCase{
		repository: repo,
		provider:   prov,
		cacheTTL:   cacheTTL,
	}
}

// Execute executes the use case to get an exchange rate for a currency pair.
//
// Flow:
// 1. Validate currency codes
// 2. Check cache (repository.Get)
// 3. If cache hit and valid (not expired) → return cached rate
// 4. If cache miss or expired → fetch from external API
// 5. Update cache with new rate
// 6. Return rate to client
//
// Fallback Strategy:
// - If circuit breaker is open (ErrCircuitOpen) → use GetStale() for fallback
// - If other provider error → fallback to stale cache (if available)
// - If both unavailable → return error
//
// Cache-First Strategy:
// - Always check cache before external API
// - Reduces external API calls (>80% reduction)
// - Faster response times (<200ms for cached)
func (uc *GetExchangeRateUseCase) Execute(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error) {
	fmt.Printf("[GetExchangeRateUseCase] Execute called with Base=%s, Target=%s\n", req.Base, req.Target)

	// Validate currency codes
	base, err := entity.NewCurrencyCode(req.Base)
	if err != nil {
		fmt.Printf("[GetExchangeRateUseCase] Invalid base currency: %v\n", err)
		return dto.RateResponse{}, fmt.Errorf("invalid base currency: %w", err)
	}

	target, err := entity.NewCurrencyCode(req.Target)
	if err != nil {
		fmt.Printf("[GetExchangeRateUseCase] Invalid target currency: %v\n", err)
		return dto.RateResponse{}, fmt.Errorf("invalid target currency: %w", err)
	}

	// Check if base and target are the same
	if base.Equal(target) {
		fmt.Printf("[GetExchangeRateUseCase] Base and target are the same\n")
		return dto.RateResponse{}, fmt.Errorf("currency code validation: %w", entity.ErrCurrencyCodeMismatch)
	}

	// Step 1: Check cache
	fmt.Printf("[GetExchangeRateUseCase] Checking cache for %s/%s\n", base, target)
	cachedRate, err := uc.repository.Get(ctx, base, target)
	if err != nil {
		fmt.Printf("[GetExchangeRateUseCase] Cache check error: %v\n", err)
	} else if cachedRate != nil {
		fmt.Printf("[GetExchangeRateUseCase] Cache hit: rate=%.4f, valid=%v\n", cachedRate.Rate, cachedRate.IsValid(uc.cacheTTL))
	}
	if err == nil && cachedRate != nil {
		// Cache hit - check if still valid
		if cachedRate.IsValid(uc.cacheTTL) {
			// Cache is valid, return it
			return dto.ToRateResponse(cachedRate), nil
		}
		// Cache exists but expired - will fetch fresh rate below
	}

	// Step 2: Fetch from external API
	freshRate, err := uc.provider.FetchRate(ctx, base, target)
	if err == nil && freshRate != nil {
		// Successfully fetched - save to cache
		if saveErr := uc.repository.Save(ctx, freshRate, uc.cacheTTL); saveErr != nil {
			// Log error but don't fail the request - cache save failure shouldn't break the flow
			// In production, you'd log this error
		}
		return dto.ToRateResponse(freshRate), nil
	}

	// Step 3: Fallback to stale cache if external API failed
	// Check if circuit breaker is open (specific handling)
	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		// Circuit is open - explicitly use GetStale() for fallback
		staleRate, staleErr := uc.repository.GetStale(ctx, base, target)
		if staleErr == nil && staleRate != nil {
			// Create stale rate entity (mark as stale)
			staleEntity, entityErr := entity.NewExchangeRate(
				staleRate.Base,
				staleRate.Target,
				staleRate.Rate,
				staleRate.Timestamp,
				true, // Mark as stale
			)
			if entityErr == nil {
				return dto.ToRateResponse(staleEntity), nil
			}
		}
		// No stale cache available - return circuit open error
		return dto.RateResponse{}, fmt.Errorf("circuit breaker is open and no stale cache available: %w", err)
	}

	// Step 4: Fallback to stale cache for other provider errors
	if cachedRate != nil {
		// Return stale cache as fallback
		staleRate, err := entity.NewExchangeRate(
			cachedRate.Base,
			cachedRate.Target,
			cachedRate.Rate,
			cachedRate.Timestamp,
			true, // Mark as stale
		)
		if err == nil {
			return dto.ToRateResponse(staleRate), nil
		}
	}

	// Both cache and external API failed
	fmt.Printf("[GetExchangeRateUseCase] Both cache and API failed. Error: %v\n", err)
	if errors.Is(err, entity.ErrRateNotFound) {
		return dto.RateResponse{}, fmt.Errorf("exchange rate not found for %s/%s: %w", base, target, err)
	}

	return dto.RateResponse{}, fmt.Errorf("failed to fetch exchange rate: %w", err)
}
