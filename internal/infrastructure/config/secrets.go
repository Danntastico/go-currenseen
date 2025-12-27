package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManager is an interface for retrieving secrets.
// This interface allows for easy testing by providing a mock implementation.
type SecretsManager interface {
	// GetAPIKey retrieves the API key from the secret.
	// Returns the API key or an error if retrieval fails.
	GetAPIKey(ctx context.Context) (string, error)
}

// cachedSecret holds a cached secret value with expiration time.
type cachedSecret struct {
	value     string
	expiresAt time.Time
	mu        sync.RWMutex
}

// isExpired checks if the cached secret has expired.
func (c *cachedSecret) isExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Now().After(c.expiresAt)
}

// get returns the cached secret value if not expired.
func (c *cachedSecret) get() (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().After(c.expiresAt) {
		return "", false
	}
	return c.value, true
}

// set updates the cached secret value and expiration time.
func (c *cachedSecret) set(value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = value
	c.expiresAt = time.Now().Add(ttl)
}

// AWSSecretsManager implements SecretsManager using AWS Secrets Manager.
type AWSSecretsManager struct {
	client     *secretsmanager.Client
	secretName string
	cacheTTL   time.Duration
	cache      *cachedSecret
	mu         sync.RWMutex
}

// NewAWSSecretsManager creates a new AWS Secrets Manager client.
//
// Parameters:
// - ctx: Context for AWS SDK operations
// - secretName: Name or ARN of the secret in Secrets Manager
// - cacheTTL: Time-to-live for cached secrets (default: 5 minutes)
//
// Returns an error if the AWS client cannot be created.
//
// Example usage:
//
//	sm, err := NewAWSSecretsManager(ctx, "my-secret", 5*time.Minute)
//	if err != nil {
//	    log.Fatalf("failed to create secrets manager: %v", err)
//	}
func NewAWSSecretsManager(ctx context.Context, secretName string, cacheTTL time.Duration) (*AWSSecretsManager, error) {
	if secretName == "" {
		return nil, fmt.Errorf("secret name is required")
	}
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute // default
	}

	// Load default AWS configuration
	// This automatically handles credentials from:
	// 1. Environment variables
	// 2. AWS credentials file
	// 3. IAM role (when running on AWS)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &AWSSecretsManager{
		client:     client,
		secretName: secretName,
		cacheTTL:   cacheTTL,
		cache:      &cachedSecret{},
	}, nil
}

// NewAWSSecretsManagerWithClient creates a new AWS Secrets Manager with a provided client.
// This is useful for testing with mock clients.
//
// Parameters:
// - client: Secrets Manager client (can be a mock for testing)
// - secretName: Name or ARN of the secret in Secrets Manager
// - cacheTTL: Time-to-live for cached secrets (default: 5 minutes)
//
// Returns an error if secretName is empty.
func NewAWSSecretsManagerWithClient(client *secretsmanager.Client, secretName string, cacheTTL time.Duration) (*AWSSecretsManager, error) {
	if secretName == "" {
		return nil, fmt.Errorf("secret name is required")
	}
	if cacheTTL <= 0 {
		cacheTTL = 5 * time.Minute // default
	}

	return &AWSSecretsManager{
		client:     client,
		secretName: secretName,
		cacheTTL:   cacheTTL,
		cache:      &cachedSecret{},
	}, nil
}

// GetAPIKey retrieves the API key from AWS Secrets Manager.
//
// The secret is expected to be a JSON object with an "api-key" field:
// {"api-key": "your-api-key-here"}
//
// The secret is cached for the configured TTL to reduce API calls.
// If the cache is expired or missing, the secret is fetched from Secrets Manager.
//
// Security: This method never logs the API key value.
func (s *AWSSecretsManager) GetAPIKey(ctx context.Context) (string, error) {
	// Check cache first
	if value, ok := s.cache.get(); ok {
		return value, nil
	}

	// Fetch from Secrets Manager
	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(s.secretName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret from Secrets Manager: %w", err)
	}

	// Parse JSON secret
	var secretData map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		return "", fmt.Errorf("failed to parse secret JSON: %w", err)
	}

	// Extract API key
	apiKey, ok := secretData["api-key"]
	if !ok {
		return "", fmt.Errorf("secret does not contain 'api-key' field")
	}
	if apiKey == "" {
		return "", fmt.Errorf("secret 'api-key' field is empty")
	}

	// Cache the secret
	s.cache.set(apiKey, s.cacheTTL)

	return apiKey, nil
}

// InvalidateCache clears the cached secret, forcing a fresh fetch on next call.
// This is useful when secrets are rotated.
func (s *AWSSecretsManager) InvalidateCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = &cachedSecret{}
}


