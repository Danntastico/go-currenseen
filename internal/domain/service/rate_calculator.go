package service

import (
	"fmt"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// RateCalculator provides exchange rate calculation utilities.
// This is a domain service that encapsulates rate calculation logic.
type RateCalculator struct{}

// NewRateCalculator creates a new RateCalculator.
func NewRateCalculator() *RateCalculator {
	return &RateCalculator{}
}

// Convert converts an amount from base currency to target currency using the exchange rate.
// Returns an error if the amount is negative or if the rate is invalid.
func (c *RateCalculator) Convert(amount float64, rate *entity.ExchangeRate) (float64, error) {
	if amount < 0 {
		return 0, fmt.Errorf("amount cannot be negative: %f", amount)
	}

	if rate == nil {
		return 0, fmt.Errorf("exchange rate cannot be nil")
	}

	if rate.Rate <= 0 {
		return 0, fmt.Errorf("invalid exchange rate: %f", rate.Rate)
	}

	return amount * rate.Rate, nil
}

// InverseRate calculates the inverse exchange rate (1/rate).
// Useful for converting in the opposite direction (target to base).
func (c *RateCalculator) InverseRate(rate *entity.ExchangeRate) (*entity.ExchangeRate, error) {
	if rate == nil {
		return nil, fmt.Errorf("exchange rate cannot be nil")
	}

	if rate.Rate <= 0 {
		return nil, fmt.Errorf("invalid exchange rate: %f", rate.Rate)
	}

	inverseRate := 1.0 / rate.Rate

	// Swap base and target for inverse rate
	inverse, err := entity.NewExchangeRate(
		rate.Target,
		rate.Base,
		inverseRate,
		rate.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create inverse rate: %w", err)
	}

	// Preserve stale flag
	inverse.Stale = rate.Stale

	return inverse, nil
}

// CrossRate calculates a cross rate between two currencies using a common base.
// For example, to get EUR/GBP, you can use USD/EUR and USD/GBP.
//
// Formula: EUR/GBP = (USD/GBP) / (USD/EUR)
func (c *RateCalculator) CrossRate(rate1, rate2 *entity.ExchangeRate) (*entity.ExchangeRate, error) {
	if rate1 == nil || rate2 == nil {
		return nil, fmt.Errorf("exchange rates cannot be nil")
	}

	// Both rates must have the same base currency
	if !rate1.Base.Equal(rate2.Base) {
		return nil, fmt.Errorf("cross rate calculation requires same base currency: %s != %s", rate1.Base, rate2.Base)
	}

	// Cannot calculate cross rate if target currencies are the same
	if rate1.Target.Equal(rate2.Target) {
		return nil, fmt.Errorf("cross rate calculation requires different target currencies")
	}

	if rate1.Rate <= 0 || rate2.Rate <= 0 {
		return nil, fmt.Errorf("invalid exchange rates for cross rate calculation")
	}

	// Cross rate = rate2 / rate1
	// Example: USD/EUR = 0.85, USD/GBP = 0.75
	// EUR/GBP = (USD/GBP) / (USD/EUR) = 0.75 / 0.85 = 0.882
	crossRateValue := rate2.Rate / rate1.Rate

	// Use the earlier timestamp (more conservative)
	timestamp := rate1.Timestamp
	if rate2.Timestamp.Before(timestamp) {
		timestamp = rate2.Timestamp
	}

	crossRate, err := entity.NewExchangeRate(
		rate1.Target,
		rate2.Target,
		crossRateValue,
		timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create cross rate: %w", err)
	}

	// Mark as stale if either rate is stale
	crossRate.Stale = rate1.Stale || rate2.Stale

	return crossRate, nil
}


