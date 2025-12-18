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
// Fallback:
// - If external API unavailable → return stale cache (if available)
// - If both unavailable → return error
func (uc *GetExchangeRateUseCase) Execute(ctx context.Context, req dto.GetRateRequest) (dto.RateResponse, error) {
	// Validate currency codes
	base, err := entity.NewCurrencyCode(req.Base)
	if err != nil {
		return dto.RateResponse{}, fmt.Errorf("invalid base currency: %w", err)
	}

	target, err := entity.NewCurrencyCode(req.Target)
	if err != nil {
		return dto.RateResponse{}, fmt.Errorf("invalid target currency: %w", err)
	}

	// Check if base and target are the same
	if base.Equal(target) {
		return dto.RateResponse{}, entity.ErrCurrencyCodeMismatch
	}

	// Step 1: Check cache
	cachedRate, err := uc.repository.Get(ctx, base, target)
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
	if errors.Is(err, entity.ErrRateNotFound) {
		return dto.RateResponse{}, fmt.Errorf("exchange rate not found for %s/%s: %w", base, target, err)
	}

	return dto.RateResponse{}, fmt.Errorf("failed to fetch exchange rate: %w", err)
}
