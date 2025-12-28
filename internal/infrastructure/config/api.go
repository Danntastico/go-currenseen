package config

import (
	"os"
	"strconv"
	"time"
)

// APIConfig holds API configuration for external exchange rate providers.
type APIConfig struct {
	BaseURL       string        // Base URL for the exchange rate API
	Timeout       time.Duration // HTTP client timeout
	RetryAttempts int           // Maximum number of retry attempts
}

// LoadAPIConfig loads API configuration from environment variables.
//
// Environment variables:
// - EXCHANGE_RATE_API_URL: Base URL for the API (default: "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1")
// - EXCHANGE_RATE_API_TIMEOUT: HTTP client timeout in seconds (default: 10)
// - EXCHANGE_RATE_API_RETRY_ATTEMPTS: Maximum retry attempts (default: 3)
//
// Returns a configuration with defaults if environment variables are not set.
//
// Note: The API has been migrated from currency-api to exchange-api.
// The new API uses jsDelivr CDN and has a different URL structure.
//
// Example usage:
//
//	cfg := LoadAPIConfig()
//	// Use cfg.BaseURL, cfg.Timeout, cfg.RetryAttempts
func LoadAPIConfig() APIConfig {
	// Load base URL from environment
	baseURL := os.Getenv("EXCHANGE_RATE_API_URL")
	if baseURL == "" {
		// New API URL: uses jsDelivr CDN (migrated from old currency-api)
		baseURL = "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1"
	}

	// Load timeout from environment (in seconds)
	timeoutSeconds := 10 // default
	if timeoutStr := os.Getenv("EXCHANGE_RATE_API_TIMEOUT"); timeoutStr != "" {
		if parsed, err := strconv.Atoi(timeoutStr); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	// Load retry attempts from environment
	retryAttempts := 3 // default
	if retryStr := os.Getenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS"); retryStr != "" {
		if parsed, err := strconv.Atoi(retryStr); err == nil && parsed > 0 {
			retryAttempts = parsed
		}
	}

	return APIConfig{
		BaseURL:       baseURL,
		Timeout:       time.Duration(timeoutSeconds) * time.Second,
		RetryAttempts: retryAttempts,
	}
}
