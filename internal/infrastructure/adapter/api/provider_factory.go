package api

import (
	"fmt"

	"github.com/misterfancybg/go-currenseen/internal/domain/provider"
)

// ProviderType represents the type of exchange rate provider.
type ProviderType string

const (
	// ProviderTypeCurrencyAPI represents the Currency-api provider.
	ProviderTypeCurrencyAPI ProviderType = "currency_api"
)

// ProviderConfig holds configuration for creating an exchange rate provider.
type ProviderConfig struct {
	Type    ProviderType // Type of provider to create
	BaseURL string       // Base URL for the API (optional, uses default if empty)
	APIKey  string       // API key (optional, for future use with APIs that require keys)
}

// NewProvider creates a new ExchangeRateProvider based on configuration.
//
// This factory function:
// - Creates a secure HTTP client
// - Instantiates the appropriate provider based on Type
// - Returns an error if the provider type is unknown
//
// Supported provider types:
// - ProviderTypeCurrencyAPI: Currency-api (free, no API key required)
//
// Example usage:
//
//	config := ProviderConfig{
//	    Type:    ProviderTypeCurrencyAPI,
//	    BaseURL: "https://api.example.com/v1",
//	}
//	provider, err := NewProvider(config)
//	if err != nil {
//	    // Handle error
//	}
func NewProvider(config ProviderConfig) (provider.ExchangeRateProvider, error) {
	client := NewHTTPClient()

	switch config.Type {
	case ProviderTypeCurrencyAPI:
		return NewCurrencyAPIProvider(client, config.BaseURL), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s", config.Type)
	}
}

// NewDefaultProvider creates a provider with default configuration.
//
// Default configuration:
// - Type: ProviderTypeCurrencyAPI
// - BaseURL: "https://api.fawazahmed0.currency-api.com/v1"
//
// This is a convenience function for quick setup. For production use,
// consider using NewProvider with explicit configuration.
func NewDefaultProvider() provider.ExchangeRateProvider {
	config := ProviderConfig{
		Type:    ProviderTypeCurrencyAPI,
		BaseURL: "https://api.fawazahmed0.currency-api.com/v1",
	}
	// Ignore error - ProviderTypeCurrencyAPI is always valid
	provider, _ := NewProvider(config)
	return provider
}
