package api

import (
	"testing"

	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

func TestNewProvider_CurrencyAPI(t *testing.T) {
	config := ProviderConfig{
		Type:    ProviderTypeCurrencyAPI,
		BaseURL: "https://api.example.com/v1",
	}

	prov, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v, want nil", err)
	}

	if prov == nil {
		t.Fatal("NewProvider() returned nil provider")
	}

	// Verify it implements the interface
	_, ok := prov.(provider.ExchangeRateProvider)
	if !ok {
		t.Error("Provider does not implement ExchangeRateProvider interface")
	}

	// Verify it's the correct type
	_, ok = prov.(*CurrencyAPIProvider)
	if !ok {
		t.Error("Provider is not *CurrencyAPIProvider")
	}
}

func TestNewProvider_CurrencyAPI_DefaultBaseURL(t *testing.T) {
	config := ProviderConfig{
		Type: ProviderTypeCurrencyAPI,
		// BaseURL not set - should use default
	}

	prov, err := NewProvider(config)
	if err != nil {
		t.Fatalf("NewProvider() error = %v, want nil", err)
	}

	if prov == nil {
		t.Fatal("NewProvider() returned nil provider")
	}

	// Verify it's a CurrencyAPIProvider
	currencyProvider, ok := prov.(*CurrencyAPIProvider)
	if !ok {
		t.Fatal("Provider is not *CurrencyAPIProvider")
	}

	// Verify default base URL is set
	expectedURL := "https://api.fawazahmed0.currency-api.com/v1"
	if currencyProvider.baseURL != expectedURL {
		t.Errorf("baseURL = %q, want %q", currencyProvider.baseURL, expectedURL)
	}
}

func TestNewProvider_UnknownType(t *testing.T) {
	config := ProviderConfig{
		Type: ProviderType("unknown_type"),
	}

	_, err := NewProvider(config)
	if err == nil {
		t.Fatal("NewProvider() error = nil, want error")
	}

	expectedErr := "unknown provider type: unknown_type"
	if err.Error() != expectedErr {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedErr)
	}
}

func TestNewDefaultProvider(t *testing.T) {
	prov := NewDefaultProvider()

	if prov == nil {
		t.Fatal("NewDefaultProvider() returned nil provider")
	}

	// Verify it implements the interface
	_, ok := prov.(provider.ExchangeRateProvider)
	if !ok {
		t.Error("Provider does not implement ExchangeRateProvider interface")
	}

	// Verify it's the correct type
	currencyProvider, ok := prov.(*CurrencyAPIProvider)
	if !ok {
		t.Error("Provider is not *CurrencyAPIProvider")
	}

	// Verify default base URL is set
	expectedURL := "https://api.fawazahmed0.currency-api.com/v1"
	if currencyProvider.baseURL != expectedURL {
		t.Errorf("baseURL = %q, want %q", currencyProvider.baseURL, expectedURL)
	}
}

func TestProviderType_String(t *testing.T) {
	tests := []struct {
		name string
		pt   ProviderType
		want string
	}{
		{"CurrencyAPI", ProviderTypeCurrencyAPI, "currency_api"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.pt)
			if got != tt.want {
				t.Errorf("ProviderType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
