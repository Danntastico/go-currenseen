package api

import (
	"strings"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

func TestParseRateResponse_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// New API format: rates nested under base currency (lowercase)
	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"usd": {
				"eur": 0.85,
			},
		},
	}

	rate, err := parseRateResponse(resp, base, target)
	if err != nil {
		t.Fatalf("parseRateResponse() error = %v, want nil", err)
	}

	if rate == nil {
		t.Fatal("parseRateResponse() returned nil rate")
	}

	if !rate.Base.Equal(base) {
		t.Errorf("Base = %v, want %v", rate.Base, base)
	}

	if !rate.Target.Equal(target) {
		t.Errorf("Target = %v, want %v", rate.Target, target)
	}

	if rate.Rate != 0.85 {
		t.Errorf("Rate = %f, want 0.85", rate.Rate)
	}

	if rate.Stale {
		t.Error("Stale = true, want false (rates from external APIs are always fresh)")
	}

	// Verify timestamp is recent (within last minute)
	if time.Since(rate.Timestamp) > time.Minute {
		t.Errorf("Timestamp is too old: %v", rate.Timestamp)
	}
}

func TestParseRateResponse_APIError(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// New API: base currency not found (empty rates)
	resp := &currencyAPIResponse{
		Date:  "2024-01-15",
		Rates: map[string]map[string]float64{},
	}

	_, err := parseRateResponse(resp, base, target)
	if err == nil {
		t.Fatal("parseRateResponse() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "base currency") {
		t.Errorf("Error message = %q, want to contain 'base currency'", err.Error())
	}
}

func TestParseRateResponse_BaseCurrencyMismatch(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// New API: base currency not found (different base in response)
	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"eur": { // Wrong base
				"usd": 1.18,
			},
		},
	}

	_, err := parseRateResponse(resp, base, target)
	if err == nil {
		t.Fatal("parseRateResponse() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "base currency") {
		t.Errorf("Error message = %q, want to contain 'base currency'", err.Error())
	}
}

func TestParseRateResponse_TargetNotFound(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"usd": {
				"gbp": 0.75, // EUR not present
			},
		},
	}

	_, err := parseRateResponse(resp, base, target)
	if err == nil {
		t.Fatal("parseRateResponse() error = nil, want error")
	}

	expectedErr := "target currency EUR not found in response"
	if err.Error() != expectedErr {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedErr)
	}
}

func TestParseRateResponse_InvalidRate(t *testing.T) {
	tests := []struct {
		name string
		rate float64
	}{
		{"zero", 0.0},
		{"negative", -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, _ := entity.NewCurrencyCode("USD")
			target, _ := entity.NewCurrencyCode("EUR")

			resp := &currencyAPIResponse{
				Date: "2024-01-15",
				Rates: map[string]map[string]float64{
					"usd": {
						"eur": tt.rate,
					},
				},
			}

			_, err := parseRateResponse(resp, base, target)
			if err == nil {
				t.Fatal("parseRateResponse() error = nil, want error")
			}
		})
	}
}

func TestParseAllRatesResponse_Success(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"usd": {
				"eur": 0.85,
				"gbp": 0.75,
				"jpy": 110.50,
			},
		},
	}

	rates, err := parseAllRatesResponse(resp, base)
	if err != nil {
		t.Fatalf("parseAllRatesResponse() error = %v, want nil", err)
	}

	if len(rates) != 3 {
		t.Errorf("len(rates) = %d, want 3", len(rates))
	}

	// Verify all rates have correct base
	for _, rate := range rates {
		if !rate.Base.Equal(base) {
			t.Errorf("Rate base = %v, want %v", rate.Base, base)
		}

		if rate.Stale {
			t.Error("Stale = true, want false")
		}
	}
}

func TestParseAllRatesResponse_APIError(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	// New API: base currency not found
	resp := &currencyAPIResponse{
		Date:  "2024-01-15",
		Rates: map[string]map[string]float64{},
	}

	_, err := parseAllRatesResponse(resp, base)
	if err == nil {
		t.Fatal("parseAllRatesResponse() error = nil, want error")
	}
}

func TestParseAllRatesResponse_BaseCurrencyMismatch(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"eur": { // Wrong base
				"usd": 1.18,
			},
		},
	}

	_, err := parseAllRatesResponse(resp, base)
	if err == nil {
		t.Fatal("parseAllRatesResponse() error = nil, want error")
	}
}

func TestParseAllRatesResponse_SkipsInvalidRates(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"usd": {
				"eur": 0.85, // Valid
				"gbp": 0.0,  // Invalid (zero)
				"jpy": -1.0, // Invalid (negative)
				"xx":  1.0,  // Invalid (wrong length - will fail NewCurrencyCode)
				"usd": 1.0,  // Invalid (same as base)
			},
		},
	}

	rates, err := parseAllRatesResponse(resp, base)
	if err != nil {
		t.Fatalf("parseAllRatesResponse() error = %v, want nil", err)
	}

	// Should only return EUR (the valid one)
	if len(rates) != 1 {
		t.Errorf("len(rates) = %d, want 1", len(rates))
	}

	eur, _ := entity.NewCurrencyCode("EUR")
	if !rates[0].Target.Equal(eur) {
		t.Errorf("Rate target = %v, want EUR", rates[0].Target)
	}
}

func TestParseAllRatesResponse_EmptyRates(t *testing.T) {
	base, _ := entity.NewCurrencyCode("USD")

	resp := &currencyAPIResponse{
		Date: "2024-01-15",
		Rates: map[string]map[string]float64{
			"usd": {}, // Empty
		},
	}

	rates, err := parseAllRatesResponse(resp, base)
	if err != nil {
		t.Fatalf("parseAllRatesResponse() error = %v, want nil", err)
	}

	// Should return empty slice (not nil)
	if rates == nil {
		t.Error("rates is nil, want empty slice")
	}

	if len(rates) != 0 {
		t.Errorf("len(rates) = %d, want 0", len(rates))
	}
}
