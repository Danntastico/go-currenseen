package entity

import (
	"math"
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
			name:      "infinity rate",
			base:      base,
			target:    target,
			rate:      math.Inf(1),
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "negative infinity rate",
			base:      base,
			target:    target,
			rate:      math.Inf(-1),
			timestamp: validTimestamp,
			wantErr:   true,
		},
		{
			name:      "NaN rate",
			base:      base,
			target:    target,
			rate:      math.NaN(),
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
			got, err := NewExchangeRate(tt.base, tt.target, tt.rate, tt.timestamp, tt.wantStale)
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
		{
			name:      "very large TTL - should not expire",
			timestamp: time.Now().Add(-1000 * time.Hour),
			ttl:       10000 * time.Hour, // Very large TTL
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := NewExchangeRate(base, target, 0.85, tt.timestamp, false)
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

	tests := []struct {
		name        string
		timestamp   time.Time
		expectedAge time.Duration
		tolerance   time.Duration
	}{
		{
			name:        "normal age calculation",
			timestamp:   time.Now().Add(-2 * time.Hour),
			expectedAge: 2 * time.Hour,
			tolerance:   5 * time.Second,
		},
		{
			name:        "very old rate",
			timestamp:   time.Now().Add(-100 * time.Hour),
			expectedAge: 100 * time.Hour,
			tolerance:   5 * time.Second,
		},
		{
			name:        "recent rate",
			timestamp:   time.Now().Add(-5 * time.Minute),
			expectedAge: 5 * time.Minute,
			tolerance:   5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := NewExchangeRate(base, target, 0.85, tt.timestamp, false)
			if err != nil {
				t.Fatalf("NewExchangeRate() error = %v", err)
			}

			age := rate.Age()

			// Allow tolerance for test execution time
			if age < tt.expectedAge-tt.tolerance || age > tt.expectedAge+tt.tolerance {
				t.Errorf("ExchangeRate.Age() = %v, want approximately %v (within %v)", age, tt.expectedAge, tt.tolerance)
			}
		})
	}

	// Edge case: future timestamp (shouldn't happen, but defensive)
	// Note: This would fail validation in NewExchangeRate, so we test Age() directly
	// by creating a rate with a future timestamp (if we bypass validation)
	t.Run("future timestamp edge case", func(t *testing.T) {
		// Create a rate struct directly to test Age() with future timestamp
		// This tests defensive behavior even though NewExchangeRate would reject it
		futureTimestamp := time.Now().Add(1 * time.Hour)
		rate := &ExchangeRate{
			Base:      base,
			Target:    target,
			Rate:      0.85,
			Timestamp: futureTimestamp,
			Stale:     false,
		}

		age := rate.Age()
		// Age should be negative for future timestamps
		if age >= 0 {
			t.Errorf("ExchangeRate.Age() for future timestamp = %v, want negative value", age)
		}
	})
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
			rate, err := NewExchangeRate(base, target, 0.85, tt.timestamp, false)
			if err != nil {
				t.Fatalf("NewExchangeRate() error = %v", err)
			}

			if got := rate.IsValid(tt.ttl); got != tt.want {
				t.Errorf("ExchangeRate.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
