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
	"github.com/misterfancybg/go-currenseen/pkg/logger"
)

// GetExchangeRateUseCase handles the use case for getting an exchange rate for a currency pair.
// This implements UC1 from the specification.
type GetExchangeRateUseCase struct {
	repository repository.ExchangeRateRepository
	provider   provider.ExchangeRateProvider
	cacheTTL   time.Duration // TTL for cached rates
	logger     *logger.Logger
}

// NewGetExchangeRateUseCase creates a new GetExchangeRateUseCase with dependency injection.
func NewGetExchangeRateUseCase(
	repo repository.ExchangeRateRepository,
	prov provider.ExchangeRateProvider,
	cacheTTL time.Duration,
	log *logger.Logger,
) *GetExchangeRateUseCase {
	if log == nil {
		log = logger.NewFromEnv()
	}
	return &GetExchangeRateUseCase{
		repository: repo,
		provider:   prov,
		cacheTTL:   cacheTTL,
		logger:     log,
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
	startTime := time.Now()
	log := uc.logger.WithContext(ctx)

	// Add currency codes to context for logging
	ctx = logger.WithCurrencyCodes(ctx, req.Base, req.Target)
	log = uc.logger.WithContext(ctx)

	log.Debug("executing get exchange rate use case",
		"base", req.Base,
		"target", req.Target,
	)

	// Validate currency codes
	base, err := entity.NewCurrencyCode(req.Base)
	if err != nil {
		log.LogError(ctx, err, "invalid base currency code")
		return dto.RateResponse{}, fmt.Errorf("invalid base currency: %w", err)
	}

	target, err := entity.NewCurrencyCode(req.Target)
	if err != nil {
		log.LogError(ctx, err, "invalid target currency code")
		return dto.RateResponse{}, fmt.Errorf("invalid target currency: %w", err)
	}

	// Check if base and target are the same
	if base.Equal(target) {
		log.Warn("base and target currencies are the same",
			"base", base.String(),
			"target", target.String(),
		)
		return dto.RateResponse{}, fmt.Errorf("currency code validation: %w", entity.ErrCurrencyCodeMismatch)
	}

	// Step 1: Check cache
	log.Debug("checking cache for exchange rate")
	cachedRate, err := uc.repository.Get(ctx, base, target)
	if err != nil {
		log.Debug("cache check error", "error", err.Error())
	} else if cachedRate != nil {
		isValid := cachedRate.IsValid(uc.cacheTTL)
		log.Debug("cache check result",
			"cache_hit", true,
			"rate", cachedRate.Rate,
			"valid", isValid,
			"timestamp", cachedRate.Timestamp,
		)
	}
	if err == nil && cachedRate != nil {
		// Cache hit - check if still valid
		if cachedRate.IsValid(uc.cacheTTL) {
			// Cache is valid, return it
			duration := time.Since(startTime)
			log.Info("cache hit, returning cached rate",
				"rate", cachedRate.Rate,
				"duration_ms", duration.Milliseconds(),
			)
			return dto.ToRateResponse(cachedRate), nil
		}
		log.Debug("cache expired, fetching fresh rate")
		// Cache exists but expired - will fetch fresh rate below
	}

	// Step 2: Fetch from external API
	log.Debug("fetching rate from external API")
	freshRate, err := uc.provider.FetchRate(ctx, base, target)
	if err == nil && freshRate != nil {
		// Successfully fetched - save to cache
		if saveErr := uc.repository.Save(ctx, freshRate, uc.cacheTTL); saveErr != nil {
			log.Warn("failed to save rate to cache",
				"error", saveErr.Error(),
			)
		} else {
			log.Debug("rate saved to cache successfully")
		}
		duration := time.Since(startTime)
		log.Info("successfully fetched rate from API",
			"rate", freshRate.Rate,
			"duration_ms", duration.Milliseconds(),
		)
		return dto.ToRateResponse(freshRate), nil
	}

	// Step 3: Fallback to stale cache if external API failed
	// Check if circuit breaker is open (specific handling)
	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		log.Warn("circuit breaker is open, attempting stale cache fallback")
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
				log.Info("returning stale cache due to circuit breaker open",
					"rate", staleEntity.Rate,
					"stale", true,
				)
				return dto.ToRateResponse(staleEntity), nil
			}
		}
		// No stale cache available - return circuit open error
		log.Error("circuit breaker open and no stale cache available",
			"error", err.Error(),
		)
		return dto.RateResponse{}, fmt.Errorf("circuit breaker is open and no stale cache available: %w", err)
	}

	// Step 4: Fallback to stale cache for other provider errors
	if cachedRate != nil {
		log.Warn("provider error, falling back to stale cache",
			"error", err.Error(),
		)
		// Return stale cache as fallback
		staleRate, err := entity.NewExchangeRate(
			cachedRate.Base,
			cachedRate.Target,
			cachedRate.Rate,
			cachedRate.Timestamp,
			true, // Mark as stale
		)
		if err == nil {
			log.Info("returning stale cache as fallback",
				"rate", staleRate.Rate,
				"stale", true,
			)
			return dto.ToRateResponse(staleRate), nil
		}
	}

	// Both cache and external API failed
	log.Error("both cache and API failed",
		"error", err.Error(),
	)
	if errors.Is(err, entity.ErrRateNotFound) {
		return dto.RateResponse{}, fmt.Errorf("exchange rate not found for %s/%s: %w", base, target, err)
	}

	return dto.RateResponse{}, fmt.Errorf("failed to fetch exchange rate: %w", err)
}
