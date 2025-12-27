package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// Config holds all configuration for the application.
type Config struct {
	// DynamoDB configuration
	DynamoDB DynamoDBConfig

	// API configuration
	API APIConfig

	// Circuit breaker configuration
	CircuitBreaker circuitbreaker.Config

	// Cache configuration
	Cache CacheConfig

	// Secrets Manager configuration
	SecretsManager SecretsManagerConfig
}

// DynamoDBConfig holds DynamoDB-specific configuration.
type DynamoDBConfig struct {
	TableName string // DynamoDB table name (required)
	Region    string // AWS region (optional, uses default if not set)
}

// CacheConfig holds cache-specific configuration.
type CacheConfig struct {
	TTL time.Duration // Cache TTL (default: 1 hour)
}

// SecretsManagerConfig holds Secrets Manager configuration.
type SecretsManagerConfig struct {
	SecretName string        // Secret name or ARN (optional)
	CacheTTL   time.Duration // Secret cache TTL (default: 5 minutes)
	Enabled    bool          // Whether to use Secrets Manager (default: false)
}

// LoadConfig loads all configuration from environment variables.
//
// Environment variables:
// - TABLE_NAME: DynamoDB table name (required)
// - AWS_REGION: AWS region (optional)
// - CACHE_TTL: Cache TTL as duration string (default: "1h")
// - EXCHANGE_RATE_API_URL: Base URL for the API (default: "https://api.fawazahmed0.currency-api.com/v1")
// - EXCHANGE_RATE_API_TIMEOUT: HTTP client timeout in seconds (default: 10)
// - EXCHANGE_RATE_API_RETRY_ATTEMPTS: Maximum retry attempts (default: 3)
// - CIRCUIT_BREAKER_FAILURE_THRESHOLD: Number of failures before opening (default: 5)
// - CIRCUIT_BREAKER_COOLDOWN_SECONDS: Cooldown duration in seconds (default: 30)
// - CIRCUIT_BREAKER_SUCCESS_THRESHOLD: Successes needed in HalfOpen to close (default: 1)
// - SECRETS_MANAGER_SECRET_NAME: Secret name or ARN (optional)
// - SECRETS_MANAGER_CACHE_TTL: Secret cache TTL as duration string (default: "5m")
// - SECRETS_MANAGER_ENABLED: Enable Secrets Manager (default: "false")
//
// Returns an error if required configuration is missing or invalid.
//
// Example usage:
//
//	cfg, err := LoadConfig()
//	if err != nil {
//	    log.Fatalf("failed to load config: %v", err)
//	}
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Load DynamoDB configuration
	cfg.DynamoDB.TableName = os.Getenv("TABLE_NAME")
	cfg.DynamoDB.Region = os.Getenv("AWS_REGION")

	// Load API configuration (reuse existing function)
	cfg.API = LoadAPIConfig()

	// Load circuit breaker configuration (reuse existing function)
	cfg.CircuitBreaker = LoadCircuitBreakerConfig()

	// Load cache configuration
	cacheTTL := 1 * time.Hour // default
	if ttlStr := os.Getenv("CACHE_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil && parsed > 0 {
			cacheTTL = parsed
		}
	}
	cfg.Cache.TTL = cacheTTL

	// Load Secrets Manager configuration
	cfg.SecretsManager.SecretName = os.Getenv("SECRETS_MANAGER_SECRET_NAME")
	cfg.SecretsManager.Enabled = os.Getenv("SECRETS_MANAGER_ENABLED") == "true"
	secretCacheTTL := 5 * time.Minute // default
	if ttlStr := os.Getenv("SECRETS_MANAGER_CACHE_TTL"); ttlStr != "" {
		if parsed, err := time.ParseDuration(ttlStr); err == nil && parsed > 0 {
			secretCacheTTL = parsed
		}
	}
	cfg.SecretsManager.CacheTTL = secretCacheTTL

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration and returns an error if invalid.
//
// Required fields:
// - DynamoDB.TableName
//
// Optional validations:
// - Cache TTL must be positive
// - Secrets Manager secret name must be set if enabled
func (c *Config) Validate() error {
	// Validate required fields
	if c.DynamoDB.TableName == "" {
		return fmt.Errorf("TABLE_NAME is required")
	}

	// Validate cache TTL
	if c.Cache.TTL <= 0 {
		return fmt.Errorf("CACHE_TTL must be positive")
	}

	// Validate Secrets Manager configuration
	if c.SecretsManager.Enabled {
		if c.SecretsManager.SecretName == "" {
			return fmt.Errorf("SECRETS_MANAGER_SECRET_NAME is required when SECRETS_MANAGER_ENABLED is true")
		}
		if c.SecretsManager.CacheTTL <= 0 {
			return fmt.Errorf("SECRETS_MANAGER_CACHE_TTL must be positive")
		}
	}

	return nil
}

// GetAPIKey retrieves the API key from Secrets Manager or environment variable.
//
// Priority:
// 1. Secrets Manager (if enabled)
// 2. EXCHANGE_RATE_API_KEY environment variable
// 3. Empty string (if neither is set)
//
// This method requires a SecretsManager instance. If Secrets Manager is not enabled,
// it falls back to the environment variable.
func (c *Config) GetAPIKey(ctx context.Context, sm SecretsManager) (string, error) {
	// Try Secrets Manager first if enabled
	if c.SecretsManager.Enabled && sm != nil {
		apiKey, err := sm.GetAPIKey(ctx)
		if err == nil && apiKey != "" {
			return apiKey, nil
		}
		// If Secrets Manager fails, fall through to environment variable
	}

	// Fallback to environment variable
	return os.Getenv("EXCHANGE_RATE_API_KEY"), nil
}
