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

func TestGetAllRatesUseCase_Execute(t *testing.T) {
	ctx := context.Background()
	cacheTTL := 1 * time.Hour

	eur, _ := entity.NewCurrencyCode("EUR")
	gbp, _ := entity.NewCurrencyCode("GBP")
	validTimestamp := time.Now().Add(-30 * time.Minute)

	tests := []struct {
		name              string
		request           dto.GetRatesRequest
		repoGetByBaseFunc func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)
		providerFunc      func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error)
		wantErr           bool
		wantStale         bool
		validateResult    func(t *testing.T, resp dto.RatesResponse)
	}{
		{
			name:    "cache hit - valid rates",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				rate1, _ := entity.NewExchangeRate(base, eur, 0.85, validTimestamp, false)
				rate2, _ := entity.NewExchangeRate(base, gbp, 0.75, validTimestamp, false)
				return []*entity.ExchangeRate{rate1, rate2}, nil
			},
			wantErr:   false,
			wantStale: false,
			validateResult: func(t *testing.T, resp dto.RatesResponse) {
				if len(resp.Rates) != 2 {
					t.Errorf("expected 2 rates, got %d", len(resp.Rates))
				}
				if resp.Rates["EUR"].Rate != 0.85 {
					t.Errorf("expected EUR rate 0.85, got %f", resp.Rates["EUR"].Rate)
				}
			},
		},
		{
			name:    "cache miss - fetch from provider",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return []*entity.ExchangeRate{}, nil
			},
			providerFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				rate1, _ := entity.NewExchangeRate(base, eur, 0.86, time.Now(), false)
				rate2, _ := entity.NewExchangeRate(base, gbp, 0.76, time.Now(), false)
				return []*entity.ExchangeRate{rate1, rate2}, nil
			},
			wantErr:   false,
			wantStale: false,
			validateResult: func(t *testing.T, resp dto.RatesResponse) {
				if len(resp.Rates) != 2 {
					t.Errorf("expected 2 rates, got %d", len(resp.Rates))
				}
			},
		},
		{
			name:    "provider fails - fallback to stale cache",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				rate1, _ := entity.NewExchangeRate(base, eur, 0.85, time.Now().Add(-2*time.Hour), false)
				return []*entity.ExchangeRate{rate1}, nil
			},
			providerFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return nil, errors.New("provider unavailable")
			},
			wantErr:   false,
			wantStale: true,
			validateResult: func(t *testing.T, resp dto.RatesResponse) {
				if len(resp.Rates) == 0 {
					t.Error("expected at least one stale rate")
				}
				if !resp.Stale {
					t.Error("expected response to be stale")
				}
			},
		},
		{
			name:    "invalid base currency",
			request: dto.GetRatesRequest{Base: "XX"},
			wantErr: true,
		},
		{
			name:    "circuit open - fallback to stale cache",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				rate1, _ := entity.NewExchangeRate(base, eur, 0.85, time.Now().Add(-2*time.Hour), false)
				return []*entity.ExchangeRate{rate1}, nil
			},
			providerFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
			},
			wantErr:   false,
			wantStale: true,
			validateResult: func(t *testing.T, resp dto.RatesResponse) {
				if len(resp.Rates) == 0 {
					t.Error("expected at least one stale rate")
				}
				if !resp.Stale {
					t.Error("expected response to be stale")
				}
			},
		},
		{
			name:    "circuit open - no stale cache",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return []*entity.ExchangeRate{}, nil
			},
			providerFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return nil, fmt.Errorf("%w: external API unavailable", circuitbreaker.ErrCircuitOpen)
			},
			wantErr: true,
		},
		{
			name:    "provider fails and no cache",
			request: dto.GetRatesRequest{Base: "USD"},
			repoGetByBaseFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return []*entity.ExchangeRate{}, nil
			},
			providerFunc: func(ctx context.Context, base entity.CurrencyCode) ([]*entity.ExchangeRate, error) {
				return nil, errors.New("provider unavailable")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				getByBaseFunc: tt.repoGetByBaseFunc,
			}
			prov := &mockProvider{
				fetchAllRatesFunc: tt.providerFunc,
			}

			uc := NewGetAllRatesUseCase(repo, prov, cacheTTL)
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
