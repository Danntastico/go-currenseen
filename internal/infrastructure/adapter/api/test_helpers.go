package api

import (
	"context"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// mockProvider is a mock implementation of ExchangeRateProvider for testing.
// This is shared across multiple test files in the api package.
type mockProvider struct {
	fetchRateFunc     func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
	fetchAllRatesFunc func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)
	callCount         int
}

func (m *mockProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	m.callCount++
	if m.fetchRateFunc != nil {
		return m.fetchRateFunc(ctx, base, target)
	}
	return nil, nil
}

func (m *mockProvider) FetchAllRates(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	m.callCount++
	if m.fetchAllRatesFunc != nil {
		return m.fetchAllRatesFunc(ctx, base)
	}
	return nil, nil
}
