package entity

import (
	"testing"
	"time"
)

func TestNewExchangeRate(t *testing.T) {
	base, _ := NewCurrencyCode("USD")
	target, _ := NewCurrencyCode("EUR")
	validTimestamp := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name      string
		base      CurrencyCode
		target    CurrencyCode
		rate      float64
		timestamp time.Time
		wantErr   bool
		wantStale bool
	}{
		{
			name:      "valid exchange rate",
			base:      base,
			target:    target,
			rate:      0.85,
			timestamp: validTimestamp,
			wantErr:   false,
			wantStale: false,
		},
		{
			name:      "same base and target",
			base:      base,
			target:    base,
			rate:      1.0,
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "zero rate",
			base:      base,
			target:    target,
			rate:      0.0,
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "negative rate",
			base:      base,
			target:    target,
			rate:      -0.85,
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "zero timestamp",
			base:      base,
			target:    target,
			rate:      0.85,
			timestamp: time.Time{},
			wantErr:   true,
		},
		{
			name:      "future timestamp",
			base:      base,
			target:    target,
			rate:      0.85,
			timestamp: time.Now().Add(10 * time.Minute),
			wantErr:   true,
		},
		{
			name:      "invalid base currency",
			base:      CurrencyCode("XX"),
			target:    target,
			rate:      0.85,
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "invalid target currency",
			base:      base,
			target:    CurrencyCode("YY"),
			rate:      0.85,
			timestamp: validTimestamp,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewExchangeRate(tt.base, tt.target, tt.rate, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExchangeRate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got == nil {
					t.Fatal("NewExchangeRate() returned nil")
				}
				if got.Stale != tt.wantStale {
					t.Errorf("NewExchangeRate() Stale = %v, want %v", got.Stale, tt.wantStale)
				}
				if got.Base != tt.base {
					t.Errorf("NewExchangeRate() Base = %v, want %v", got.Base, tt.base)
				}
				if got.Target != tt.target {
					t.Errorf("NewExchangeRate() Target = %v, want %v", got.Target, tt.target)
				}
				if got.Rate != tt.rate {
					t.Errorf("NewExchangeRate() Rate = %v, want %v", got.Rate, tt.rate)
				}
			}
		})
	}
}

func TestNewStaleExchangeRate(t *testing.T) {
	base, _ := NewCurrencyCode("USD")
	target, _ := NewCurrencyCode("EUR")
	timestamp := time.Now().Add(-2 * time.Hour)

	rate, err := NewStaleExchangeRate(base, target, 0.85, timestamp)
	if err != nil {
		t.Fatalf("NewStaleExchangeRate() error = %v", err)
	}

	if !rate.Stale {
		t.Error("NewStaleExchangeRate() Stale = false, want true")
	}
}

func TestExchangeRate_IsExpired(t *testing.T) {
	base, _ := NewCurrencyCode("USD")
	target, _ := NewCurrencyCode("EUR")

	tests := []struct {
		name      string
		timestamp time.Time
		ttl       time.Duration
		want      bool
	}{
		{
			name:      "not expired - within TTL",
			timestamp: time.Now().Add(-30 * time.Minute),
			ttl:       1 * time.Hour,
			want:      false,
		},
		{
			name:      "expired - past TTL",
			timestamp: time.Now().Add(-2 * time.Hour),
			ttl:       1 * time.Hour,
			want:      true,
		},
		{
			name:      "just expired",
			timestamp: time.Now().Add(-1 * time.Hour),
			ttl:       1 * time.Hour,
			want:      true, // Should be expired if exactly at TTL boundary
		},
		{
			name:      "zero TTL - never expires",
			timestamp: time.Now().Add(-100 * time.Hour),
			ttl:       0,
			want:      false,
		},
		{
			name:      "negative TTL - never expires",
			timestamp: time.Now().Add(-100 * time.Hour),
			ttl:       -1 * time.Hour,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := NewExchangeRate(base, target, 0.85, tt.timestamp)
			if err != nil {
				t.Fatalf("NewExchangeRate() error = %v", err)
			}

			if got := rate.IsExpired(tt.ttl); got != tt.want {
				t.Errorf("ExchangeRate.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExchangeRate_Age(t *testing.T) {
	base, _ := NewCurrencyCode("USD")
	target, _ := NewCurrencyCode("EUR")
	expectedAge := 2 * time.Hour
	timestamp := time.Now().Add(-expectedAge)

	rate, err := NewExchangeRate(base, target, 0.85, timestamp)
	if err != nil {
		t.Fatalf("NewExchangeRate() error = %v", err)
	}

	age := rate.Age()

	// Allow tolerance for test execution time (tests should run quickly)
	tolerance := 5 * time.Second
	if age < expectedAge-tolerance || age > expectedAge+tolerance {
		t.Errorf("ExchangeRate.Age() = %v, want approximately %v (within %v)", age, expectedAge, tolerance)
	}
}

func TestExchangeRate_IsValid(t *testing.T) {
	base, _ := NewCurrencyCode("USD")
	target, _ := NewCurrencyCode("EUR")

	tests := []struct {
		name      string
		timestamp time.Time
		ttl       time.Duration
		want      bool
	}{
		{
			name:      "valid - within TTL",
			timestamp: time.Now().Add(-30 * time.Minute),
			ttl:       1 * time.Hour,
			want:      true,
		},
		{
			name:      "invalid - expired",
			timestamp: time.Now().Add(-2 * time.Hour),
			ttl:       1 * time.Hour,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := NewExchangeRate(base, target, 0.85, tt.timestamp)
			if err != nil {
				t.Fatalf("NewExchangeRate() error = %v", err)
			}

			if got := rate.IsValid(tt.ttl); got != tt.want {
				t.Errorf("ExchangeRate.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

