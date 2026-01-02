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

// GetAllRatesUseCase handles the use case for getting all exchange rates for a base currency.
// This implements UC2 from the specification.
type GetAllRatesUseCase struct {
	repository repository.ExchangeRateRepository
	provider   provider.ExchangeRateProvider
	cacheTTL   time.Duration // TTL for cached rates
	logger     *logger.Logger
}

// NewGetAllRatesUseCase creates a new GetAllRatesUseCase with dependency injection.
func NewGetAllRatesUseCase(
	repo repository.ExchangeRateRepository,
	prov provider.ExchangeRateProvider,
	cacheTTL time.Duration,
	log *logger.Logger,
) *GetAllRatesUseCase {
	if log == nil {
		log = logger.NewFromEnv()
	}
	return &GetAllRatesUseCase{
		repository: repo,
		provider:   prov,
		cacheTTL:   cacheTTL,
		logger:     log,
	}
}

// Execute executes the use case to get all exchange rates for a base currency.
//
// Flow:
// 1. Validate base currency code
// 2. Check cache (repository.GetByBase)
// 3. If cache hit and all valid → return all cached rates
// 4. If cache miss or some expired → fetch from external API
// 5. Cache all rates
// 6. Return rates to client
//
// Fallback Strategy:
// - If circuit breaker is open (ErrCircuitOpen) → return stale cached rates
// - If other provider error → fallback to stale cached rates (if available)
// - If both unavailable → return error
//
// Cache-First Strategy:
// - Always check cache before external API
// - Reduces external API calls (>80% reduction)
// - Faster response times (<200ms for cached)
//
// Note: This implementation fetches all rates from the provider if cache miss.
// In a production system, you might want to check which rates are missing/expired
// and only fetch those, but for simplicity, we fetch all rates.
func (uc *GetAllRatesUseCase) Execute(ctx context.Context, req dto.GetRatesRequest) (dto.RatesResponse, error) {
	startTime := time.Now()
	log := uc.logger.WithContext(ctx)

	// Add base currency to context for logging
	ctx = logger.WithCurrencyCodes(ctx, req.Base, "")
	log = uc.logger.WithContext(ctx)

	log.Debug("executing get all rates use case",
		"base", req.Base,
	)

	// Validate base currency code
	base, err := entity.NewCurrencyCode(req.Base)
	if err != nil {
		log.LogError(ctx, err, "invalid base currency code")
		return dto.RatesResponse{}, fmt.Errorf("invalid base currency: %w", err)
	}

	// Step 1: Check cache
	log.Debug("checking cache for exchange rates")
	cachedRates, err := uc.repository.GetByBase(ctx, base)
	if err == nil && len(cachedRates) > 0 {
		// Check if all cached rates are still valid
		allValid := true
		for _, rate := range cachedRates {
			if rate != nil && !rate.IsValid(uc.cacheTTL) {
				allValid = false
				break
			}
		}

		if allValid {
			// All cached rates are valid, return them
			duration := time.Since(startTime)
			log.Info("cache hit, returning cached rates",
				"rates_count", len(cachedRates),
				"duration_ms", duration.Milliseconds(),
			)
			return dto.ToRatesResponse(cachedRates), nil
		}
		log.Debug("some cached rates expired, fetching fresh rates")
		// Some rates expired - will fetch fresh rates below
	}

	// Step 2: Fetch from external API
	log.Debug("fetching rates from external API")
	freshRates, err := uc.provider.FetchAllRates(ctx, base)
	if err != nil {
		// Check if circuit breaker is open (specific handling)
		if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
			log.Warn("circuit breaker is open, attempting stale cache fallback")
			// Circuit is open - return stale cached rates (GetByBase already returns stale data)
			if len(cachedRates) > 0 {
				// Mark all as stale since they're expired
				staleRates := make([]*entity.ExchangeRate, 0, len(cachedRates))
				for _, rate := range cachedRates {
					if rate != nil {
						staleRate, staleErr := entity.NewExchangeRate(
							rate.Base,
							rate.Target,
							rate.Rate,
							rate.Timestamp,
							true, // Mark as stale
						)
						if staleErr == nil {
							staleRates = append(staleRates, staleRate)
						}
					}
				}
				if len(staleRates) > 0 {
					log.Info("returning stale cache due to circuit breaker open",
						"rates_count", len(staleRates),
						"stale", true,
					)
					return dto.ToRatesResponse(staleRates), nil
				}
			}
			// No stale cache available - return circuit open error
			log.Error("circuit breaker open and no stale cache available",
				"error", err.Error(),
			)
			return dto.RatesResponse{}, fmt.Errorf("circuit breaker is open and no stale cache available: %w", err)
		}

		// Step 3: Fallback to stale cache for other provider errors
		log.Warn("provider error, falling back to stale cache",
			"error", err.Error(),
		)
		if len(cachedRates) > 0 {
			// Mark all as stale since they're expired
			staleRates := make([]*entity.ExchangeRate, 0, len(cachedRates))
			for _, rate := range cachedRates {
				if rate != nil {
					staleRate, staleErr := entity.NewExchangeRate(
						rate.Base,
						rate.Target,
						rate.Rate,
						rate.Timestamp,
						true, // Mark as stale
					)
					if staleErr == nil {
						staleRates = append(staleRates, staleRate)
					}
				}
			}
			if len(staleRates) > 0 {
				log.Info("returning stale cache as fallback",
					"rates_count", len(staleRates),
					"stale", true,
				)
				return dto.ToRatesResponse(staleRates), nil
			}
		}
		log.Error("failed to fetch exchange rates",
			"error", err.Error(),
		)
		return dto.RatesResponse{}, fmt.Errorf("failed to fetch exchange rates: %w", err)
	}

	// Step 3: Save all rates to cache (or Step 2 if no error)
	for _, rate := range freshRates {
		if rate != nil {
			if saveErr := uc.repository.Save(ctx, rate, uc.cacheTTL); saveErr != nil {
				log.Warn("failed to save rate to cache",
					"error", saveErr.Error(),
					"base", rate.Base.String(),
					"target", rate.Target.String(),
				)
			}
		}
	}

	duration := time.Since(startTime)
	log.Info("successfully fetched rates from API",
		"rates_count", len(freshRates),
		"duration_ms", duration.Milliseconds(),
	)
	return dto.ToRatesResponse(freshRates), nil
}
