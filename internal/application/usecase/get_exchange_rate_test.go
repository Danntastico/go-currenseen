package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// mockRepository is a mock implementation of ExchangeRateRepository for testing.
type mockRepository struct {
	getFunc       func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
	saveFunc      func(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error
	getByBaseFunc func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)
	deleteFunc    func(ctx context.Context, base, target entity.CurrencyCode) error
	getStaleFunc  func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
}

func (m *mockRepository) Get(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, base, target)
	}
	return nil, entity.ErrRateNotFound
}

func (m *mockRepository) Save(ctx context.Context, rate *entity.ExchangeRate, ttl time.Duration) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, rate, ttl)
	}
	return nil
}

func (m *mockRepository) GetByBase(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	if m.getByBaseFunc != nil {
		return m.getByBaseFunc(ctx, base)
	}
	return []*entity.ExchangeRate{}, nil
}

func (m *mockRepository) Delete(ctx context.Context, base, target entity.CurrencyCode) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, base, target)
	}
	return nil
}

func (m *mockRepository) GetStale(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	if m.getStaleFunc != nil {
		return m.getStaleFunc(ctx, base, target)
	}
	return nil, entity.ErrRateNotFound
}

// mockProvider is a mock implementation of ExchangeRateProvider for testing.
type mockProvider struct {
	fetchRateFunc     func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
	fetchAllRatesFunc func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)
}

func (m *mockProvider) FetchRate(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
	if m.fetchRateFunc != nil {
		return m.fetchRateFunc(ctx, base, target)
	}
	return nil, errors.New("not implemented")
}

func (m *mockProvider) FetchAllRates(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
	if m.fetchAllRatesFunc != nil {
		return m.fetchAllRatesFunc(ctx, base)
	}
	return nil, errors.New("not implemented")
}

func TestGetExchangeRateUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	cacheTTL := 1 * time.Hour

	validTimestamp := time.Now().Add(-30 * time.Minute) // 30 minutes ago, within TTL
	expiredTimestamp := time.Now().Add(-2 * time.Hour)  // 2 hours ago, expired

	tests := []struct {
		name             string
		request          dto.GetRateRequest
		repoGetFunc      func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
		repoGetStaleFunc func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
		providerFunc     func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error)
		wantErr          bool
		wantStale        bool
		validateResult   func(t *testing.T, resp dto.RateResponse)
	}{
		{
			name:    "cache hit - valid rate",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.85, validTimestamp, false)
				return rate, nil
			},
			repoGetStaleFunc: nil,
			wantErr:          false,
			wantStale:        false,
			validateResult: func(t *testing.T, resp dto.RateResponse) {
				if resp.Rate != 0.85 {
					t.Errorf("expected rate 0.85, got %f", resp.Rate)
				}
				if resp.Stale {
					t.Error("expected rate not to be stale")
				}
			},
		},
		{
			name:    "cache miss - fetch from provider",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, entity.ErrRateNotFound
			},
			repoGetStaleFunc: nil,
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.86, time.Now(), false)
				return rate, nil
			},
			wantErr:   false,
			wantStale: false,
			validateResult: func(t *testing.T, resp dto.RateResponse) {
				if resp.Rate != 0.86 {
					t.Errorf("expected rate 0.86, got %f", resp.Rate)
				}
			},
		},
		{
			name:    "cache expired - fetch from provider",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.85, expiredTimestamp, false)
				return rate, nil
			},
			repoGetStaleFunc: nil,
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.87, time.Now(), false)
				return rate, nil
			},
			wantErr:   false,
			wantStale: false,
			validateResult: func(t *testing.T, resp dto.RateResponse) {
				if resp.Rate != 0.87 {
					t.Errorf("expected rate 0.87, got %f", resp.Rate)
				}
			},
		},
		{
			name:    "provider fails - fallback to stale cache",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.85, expiredTimestamp, false)
				return rate, nil
			},
			repoGetStaleFunc: nil,
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, errors.New("provider unavailable")
			},
			wantErr:   false,
			wantStale: true,
			validateResult: func(t *testing.T, resp dto.RateResponse) {
				if resp.Rate != 0.85 {
					t.Errorf("expected rate 0.85, got %f", resp.Rate)
				}
				if !resp.Stale {
					t.Error("expected rate to be stale")
				}
			},
		},
		{
			name:             "invalid base currency",
			request:          dto.GetRateRequest{Base: "XX", Target: "EUR"},
			repoGetFunc:      nil,
			repoGetStaleFunc: nil,
			providerFunc:     nil,
			wantErr:          true,
		},
		{
			name:             "invalid target currency",
			request:          dto.GetRateRequest{Base: "USD", Target: "YY"},
			repoGetFunc:      nil,
			repoGetStaleFunc: nil,
			providerFunc:     nil,
			wantErr:          true,
		},
		{
			name:             "same base and target",
			request:          dto.GetRateRequest{Base: "USD", Target: "USD"},
			repoGetFunc:      nil,
			repoGetStaleFunc: nil,
			providerFunc:     nil,
			wantErr:          true,
		},
		{
			name:    "circuit open - fallback to stale cache",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.85, expiredTimestamp, false)
				return rate, nil
			},
			repoGetStaleFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				rate, _ := entity.NewExchangeRate(base, target, 0.85, expiredTimestamp, false)
				return rate, nil
			},
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
			},
			wantErr:   false,
			wantStale: true,
			validateResult: func(t *testing.T, resp dto.RateResponse) {
				if resp.Rate != 0.85 {
					t.Errorf("expected rate 0.85, got %f", resp.Rate)
				}
				if !resp.Stale {
					t.Error("expected rate to be stale")
				}
			},
		},
		{
			name:    "circuit open - no stale cache",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, entity.ErrRateNotFound
			},
			repoGetStaleFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, entity.ErrRateNotFound
			},
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
			},
			wantErr: true,
		},
		{
			name:    "both cache and provider fail",
			request: dto.GetRateRequest{Base: "USD", Target: "EUR"},
			repoGetFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, entity.ErrRateNotFound
			},
			repoGetStaleFunc: nil,
			providerFunc: func(ctx context.Context, base, target entity.CurrencyCode) (*entity.ExchangeRate, error) {
				return nil, errors.New("provider unavailable")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				getFunc:      tt.repoGetFunc,
				getStaleFunc: tt.repoGetStaleFunc,
			}
			prov := &mockProvider{
				fetchRateFunc: tt.providerFunc,
			}

			uc := NewGetExchangeRateUseCase(repo, prov, cacheTTL)
			resp, err := uc.Execute(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.validateResult != nil {
					tt.validateResult(t, resp)
				}
				if resp.Stale != tt.wantStale {
					t.Errorf("Execute() Stale = %v, want %v", resp.Stale, tt.wantStale)
				}
			}
		})
	}
}
