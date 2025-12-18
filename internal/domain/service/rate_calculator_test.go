package service

import (
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

func TestRateCalculator_Convert(t *testing.T) {
	calculator := NewRateCalculator()
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	timestamp := time.Now()

	rate, err := entity.NewExchangeRate(base, target, 0.85, timestamp, false)
	if err != nil {
		t.Fatalf("Failed to create exchange rate: %v", err)
	}

	tests := []struct {
		name    string
		amount  float64
		rate    *entity.ExchangeRate
		want    float64
		wantErr bool
	}{
		{
			name:    "valid conversion",
			amount:  100.0,
			rate:    rate,
			want:    85.0,
			wantErr: false,
		},
		{
			name:    "zero amount",
			amount:  0.0,
			rate:    rate,
			want:    0.0,
			wantErr: false,
		},
		{
			name:    "negative amount",
			amount:  -100.0,
			rate:    rate,
			want:    0,
			wantErr: true,
		},
		{
			name:    "nil rate",
			amount:  100.0,
			rate:    nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "very large amount",
			amount:  1e15, // 1 quadrillion
			rate:    rate,
			want:    8.5e14, // Should handle large numbers
			wantErr: false,
		},
		{
			name:    "very small amount",
			amount:  0.0001,
			rate:    rate,
			want:    0.000085,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.Convert(tt.amount, tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRateCalculator_InverseRate(t *testing.T) {
	calculator := NewRateCalculator()
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")
	timestamp := time.Now()

	rate, err := entity.NewExchangeRate(base, target, 0.85, timestamp, false)
	if err != nil {
		t.Fatalf("Failed to create exchange rate: %v", err)
	}

	tests := []struct {
		name    string
		rate    *entity.ExchangeRate
		wantErr bool
	}{
		{
			name:    "valid inverse",
			rate:    rate,
			wantErr: false,
		},
		{
			name:    "nil rate",
			rate:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.InverseRate(tt.rate)
			if (err != nil) != tt.wantErr {
				t.Errorf("InverseRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Check that inverse rate is correct
				expectedRate := 1.0 / tt.rate.Rate
				if got.Rate != expectedRate {
					t.Errorf("InverseRate() Rate = %v, want %v", got.Rate, expectedRate)
				}
				// Check that base and target are swapped
				if !got.Base.Equal(tt.rate.Target) {
					t.Errorf("InverseRate() Base = %v, want %v", got.Base, tt.rate.Target)
				}
				if !got.Target.Equal(tt.rate.Base) {
					t.Errorf("InverseRate() Target = %v, want %v", got.Target, tt.rate.Base)
				}
			}
		})
	}
}

func TestRateCalculator_CrossRate(t *testing.T) {
	calculator := NewRateCalculator()
	timestamp := time.Now()

	usdEur, _ := entity.NewExchangeRate(
		entity.CurrencyCode("USD"),
		entity.CurrencyCode("EUR"),
		0.85,
		timestamp,
		false,
	)

	usdGbp, _ := entity.NewExchangeRate(
		entity.CurrencyCode("USD"),
		entity.CurrencyCode("GBP"),
		0.75,
		timestamp,
		false,
	)

	eurGbp, _ := entity.NewExchangeRate(
		entity.CurrencyCode("EUR"),
		entity.CurrencyCode("GBP"),
		0.90,
		timestamp,
		false,
	)

	tests := []struct {
		name    string
		rate1   *entity.ExchangeRate
		rate2   *entity.ExchangeRate
		wantErr bool
	}{
		{
			name:    "valid cross rate - USD/EUR and USD/GBP",
			rate1:   usdEur,
			rate2:   usdGbp,
			wantErr: false,
		},
		{
			name:    "different base currencies",
			rate1:   usdEur,
			rate2:   eurGbp,
			wantErr: true,
		},
		{
			name:    "same target currencies",
			rate1:   usdEur,
			rate2:   usdEur,
			wantErr: true,
		},
		{
			name:    "nil rate1",
			rate1:   nil,
			rate2:   usdGbp,
			wantErr: true,
		},
		{
			name:    "nil rate2",
			rate1:   usdEur,
			rate2:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculator.CrossRate(tt.rate1, tt.rate2)
			if (err != nil) != tt.wantErr {
				t.Errorf("CrossRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify cross rate calculation
				expectedRate := tt.rate2.Rate / tt.rate1.Rate
				if got.Rate != expectedRate {
					t.Errorf("CrossRate() Rate = %v, want %v", got.Rate, expectedRate)
				}
				// Verify currencies
				if !got.Base.Equal(tt.rate1.Target) {
					t.Errorf("CrossRate() Base = %v, want %v", got.Base, tt.rate1.Target)
				}
				if !got.Target.Equal(tt.rate2.Target) {
					t.Errorf("CrossRate() Target = %v, want %v", got.Target, tt.rate2.Target)
				}
			}
		})
	}
}
