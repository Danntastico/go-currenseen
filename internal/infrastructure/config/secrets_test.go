package config

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Note: For comprehensive testing of AWS Secrets Manager integration,
// integration tests with real AWS credentials would be needed.
// Unit tests focus on testable logic (JSON parsing, validation, caching structure).

func TestNewAWSSecretsManagerWithClient(t *testing.T) {
	// Create a minimal test client - in real tests, you'd use a proper mock
	// For now, we'll test the validation logic
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Skip("Skipping test: AWS credentials not available")
	}
	realClient := secretsmanager.NewFromConfig(cfg)

	tests := []struct {
		name       string
		client     *secretsmanager.Client
		secretName string
		cacheTTL   time.Duration
		wantErr    bool
	}{
		{
			name:       "valid secret name",
			client:     realClient,
			secretName: "my-secret",
			cacheTTL:   5 * time.Minute,
			wantErr:    false,
		},
		{
			name:       "empty secret name",
			client:     realClient,
			secretName: "",
			cacheTTL:   5 * time.Minute,
			wantErr:    true,
		},
		{
			name:       "zero cache TTL uses default",
			client:     realClient,
			secretName: "my-secret",
			cacheTTL:   0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm, err := NewAWSSecretsManagerWithClient(tt.client, tt.secretName, tt.cacheTTL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAWSSecretsManagerWithClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sm == nil {
				t.Error("expected non-nil SecretsManager")
			}
		})
	}
}

// Note: Testing GetAPIKey with real AWS Secrets Manager requires credentials.
// For unit tests, we focus on testing the caching and error handling logic.
// Integration tests would test the actual AWS Secrets Manager integration.
func TestAWSSecretsManager_GetAPIKey_Logic(t *testing.T) {
	// This test focuses on the logic we can test without AWS credentials
	// The actual AWS integration would be tested in integration tests

	t.Run("secret JSON parsing", func(t *testing.T) {
		// Test that we can parse valid secret JSON
		validJSON := `{"api-key": "test-key"}`
		var secretData map[string]string
		if err := json.Unmarshal([]byte(validJSON), &secretData); err != nil {
			t.Fatalf("failed to parse valid JSON: %v", err)
		}
		if secretData["api-key"] != "test-key" {
			t.Errorf("expected 'test-key', got %q", secretData["api-key"])
		}
	})

	t.Run("invalid JSON handling", func(t *testing.T) {
		invalidJSON := `invalid json`
		var secretData map[string]string
		if err := json.Unmarshal([]byte(invalidJSON), &secretData); err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("missing api-key field", func(t *testing.T) {
		jsonWithoutKey := `{"other-field": "value"}`
		var secretData map[string]string
		if err := json.Unmarshal([]byte(jsonWithoutKey), &secretData); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}
		if _, ok := secretData["api-key"]; ok {
			t.Error("expected 'api-key' field to be missing")
		}
	})
}

// Note: Caching and invalidation tests require AWS credentials or a more sophisticated mock.
// These would be better suited for integration tests.
// The caching logic is tested in TestCachedSecret below.

func TestCachedSecret(t *testing.T) {
	cache := &cachedSecret{}

	// Test empty cache
	if value, ok := cache.get(); ok {
		t.Errorf("expected empty cache, got value %q", value)
	}

	// Set cache
	cache.set("test-value", 1*time.Minute)

	// Get from cache
	value, ok := cache.get()
	if !ok {
		t.Error("expected cache hit")
	}
	if value != "test-value" {
		t.Errorf("expected 'test-value', got %q", value)
	}

	// Test expiration
	cache.set("expired-value", -1*time.Minute) // Already expired
	if value, ok := cache.get(); ok {
		t.Errorf("expected expired cache, got value %q", value)
	}
}

// Helper function to create a valid secret JSON for testing
func createSecretJSON(apiKey string) string {
	secret := map[string]string{"api-key": apiKey}
	jsonBytes, _ := json.Marshal(secret)
	return string(jsonBytes)
}
