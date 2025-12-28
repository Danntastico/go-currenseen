package config

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"TABLE_NAME",
		"AWS_REGION",
		"CACHE_TTL",
		"EXCHANGE_RATE_API_URL",
		"EXCHANGE_RATE_API_TIMEOUT",
		"EXCHANGE_RATE_API_RETRY_ATTEMPTS",
		"CIRCUIT_BREAKER_FAILURE_THRESHOLD",
		"CIRCUIT_BREAKER_COOLDOWN_SECONDS",
		"CIRCUIT_BREAKER_SUCCESS_THRESHOLD",
		"SECRETS_MANAGER_SECRET_NAME",
		"SECRETS_MANAGER_CACHE_TTL",
		"SECRETS_MANAGER_ENABLED",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		// Restore original environment
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	tests := []struct {
		name       string
		envVars    map[string]string
		wantErr    bool
		validateFn func(*testing.T, *Config)
	}{
		{
			name: "minimal valid config",
			envVars: map[string]string{
				"TABLE_NAME": "TestTable",
			},
			wantErr: false,
			validateFn: func(t *testing.T, cfg *Config) {
				if cfg.DynamoDB.TableName != "TestTable" {
					t.Errorf("expected TableName = 'TestTable', got %q", cfg.DynamoDB.TableName)
				}
				if cfg.Cache.TTL != 1*time.Hour {
					t.Errorf("expected default Cache.TTL = 1h, got %v", cfg.Cache.TTL)
				}
				if cfg.API.BaseURL == "" {
					t.Error("expected default API.BaseURL to be set")
				}
			},
		},
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"TABLE_NAME":                        "MyTable",
				"AWS_REGION":                        "us-east-1",
				"CACHE_TTL":                         "2h",
				"EXCHANGE_RATE_API_URL":             "https://api.example.com",
				"EXCHANGE_RATE_API_TIMEOUT":         "15",
				"EXCHANGE_RATE_API_RETRY_ATTEMPTS":  "5",
				"CIRCUIT_BREAKER_FAILURE_THRESHOLD": "10",
				"CIRCUIT_BREAKER_COOLDOWN_SECONDS":  "60",
				"CIRCUIT_BREAKER_SUCCESS_THRESHOLD": "2",
				"SECRETS_MANAGER_SECRET_NAME":       "my-secret",
				"SECRETS_MANAGER_CACHE_TTL":         "10m",
				"SECRETS_MANAGER_ENABLED":           "true",
			},
			wantErr: false,
			validateFn: func(t *testing.T, cfg *Config) {
				if cfg.DynamoDB.TableName != "MyTable" {
					t.Errorf("expected TableName = 'MyTable', got %q", cfg.DynamoDB.TableName)
				}
				if cfg.DynamoDB.Region != "us-east-1" {
					t.Errorf("expected Region = 'us-east-1', got %q", cfg.DynamoDB.Region)
				}
				if cfg.Cache.TTL != 2*time.Hour {
					t.Errorf("expected Cache.TTL = 2h, got %v", cfg.Cache.TTL)
				}
				if cfg.API.BaseURL != "https://api.example.com" {
					t.Errorf("expected API.BaseURL = 'https://api.example.com', got %q", cfg.API.BaseURL)
				}
				if cfg.API.Timeout != 15*time.Second {
					t.Errorf("expected API.Timeout = 15s, got %v", cfg.API.Timeout)
				}
				if cfg.API.RetryAttempts != 5 {
					t.Errorf("expected API.RetryAttempts = 5, got %d", cfg.API.RetryAttempts)
				}
				if cfg.CircuitBreaker.FailureThreshold != 10 {
					t.Errorf("expected CircuitBreaker.FailureThreshold = 10, got %d", cfg.CircuitBreaker.FailureThreshold)
				}
				if cfg.CircuitBreaker.CooldownDuration != 60*time.Second {
					t.Errorf("expected CircuitBreaker.CooldownDuration = 60s, got %v", cfg.CircuitBreaker.CooldownDuration)
				}
				if cfg.CircuitBreaker.SuccessThreshold != 2 {
					t.Errorf("expected CircuitBreaker.SuccessThreshold = 2, got %d", cfg.CircuitBreaker.SuccessThreshold)
				}
				if cfg.SecretsManager.SecretName != "my-secret" {
					t.Errorf("expected SecretsManager.SecretName = 'my-secret', got %q", cfg.SecretsManager.SecretName)
				}
				if cfg.SecretsManager.CacheTTL != 10*time.Minute {
					t.Errorf("expected SecretsManager.CacheTTL = 10m, got %v", cfg.SecretsManager.CacheTTL)
				}
				if !cfg.SecretsManager.Enabled {
					t.Error("expected SecretsManager.Enabled = true")
				}
			},
		},
		{
			name:    "missing required TABLE_NAME",
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "invalid CACHE_TTL",
			envVars: map[string]string{
				"TABLE_NAME": "TestTable",
				"CACHE_TTL":  "invalid",
			},
			wantErr: false, // Invalid duration is ignored, uses default
			validateFn: func(t *testing.T, cfg *Config) {
				if cfg.Cache.TTL != 1*time.Hour {
					t.Errorf("expected default Cache.TTL = 1h for invalid duration, got %v", cfg.Cache.TTL)
				}
			},
		},
		{
			name: "Secrets Manager enabled without secret name",
			envVars: map[string]string{
				"TABLE_NAME":              "TestTable",
				"SECRETS_MANAGER_ENABLED": "true",
			},
			wantErr: true,
		},
		{
			name: "Secrets Manager enabled with secret name",
			envVars: map[string]string{
				"TABLE_NAME":                  "TestTable",
				"SECRETS_MANAGER_SECRET_NAME": "my-secret",
				"SECRETS_MANAGER_ENABLED":     "true",
			},
			wantErr: false,
			validateFn: func(t *testing.T, cfg *Config) {
				if !cfg.SecretsManager.Enabled {
					t.Error("expected SecretsManager.Enabled = true")
				}
				if cfg.SecretsManager.SecretName != "my-secret" {
					t.Errorf("expected SecretsManager.SecretName = 'my-secret', got %q", cfg.SecretsManager.SecretName)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			for _, key := range envVars {
				os.Unsetenv(key)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Load configuration
			cfg, err := LoadConfig()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate configuration if no error expected
			if !tt.wantErr && tt.validateFn != nil {
				tt.validateFn(t, cfg)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "TestTable",
				},
				Cache: CacheConfig{
					TTL: 1 * time.Hour,
				},
			},
			wantErr: false,
		},
		{
			name: "missing table name",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "",
				},
				Cache: CacheConfig{
					TTL: 1 * time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid cache TTL",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "TestTable",
				},
				Cache: CacheConfig{
					TTL: 0,
				},
			},
			wantErr: true,
		},
		{
			name: "Secrets Manager enabled without secret name",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "TestTable",
				},
				Cache: CacheConfig{
					TTL: 1 * time.Hour,
				},
				SecretsManager: SecretsManagerConfig{
					Enabled:    true,
					SecretName: "",
				},
			},
			wantErr: true,
		},
		{
			name: "Secrets Manager enabled with invalid cache TTL",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "TestTable",
				},
				Cache: CacheConfig{
					TTL: 1 * time.Hour,
				},
				SecretsManager: SecretsManagerConfig{
					Enabled:    true,
					SecretName: "my-secret",
					CacheTTL:   0,
				},
			},
			wantErr: true,
		},
		{
			name: "Secrets Manager enabled with valid config",
			config: &Config{
				DynamoDB: DynamoDBConfig{
					TableName: "TestTable",
				},
				Cache: CacheConfig{
					TTL: 1 * time.Hour,
				},
				SecretsManager: SecretsManagerConfig{
					Enabled:    true,
					SecretName: "my-secret",
					CacheTTL:   5 * time.Minute,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetAPIKey(t *testing.T) {
	// Save original environment
	originalAPIKey := os.Getenv("EXCHANGE_RATE_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("EXCHANGE_RATE_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("EXCHANGE_RATE_API_KEY")
		}
	}()

	tests := []struct {
		name           string
		config         *Config
		secretsManager SecretsManager
		envAPIKey      string
		want           string
		wantErr        bool
	}{
		{
			name: "Secrets Manager enabled and working",
			config: &Config{
				SecretsManager: SecretsManagerConfig{
					Enabled: true,
				},
			},
			secretsManager: &mockSecretsManager{apiKey: "secret-api-key"},
			want:           "secret-api-key",
			wantErr:        false,
		},
		{
			name: "Secrets Manager enabled but fails, fallback to env",
			config: &Config{
				SecretsManager: SecretsManagerConfig{
					Enabled: true,
				},
			},
			secretsManager: &mockSecretsManager{err: true},
			envAPIKey:      "env-api-key",
			want:           "env-api-key",
			wantErr:        false,
		},
		{
			name: "Secrets Manager disabled, use env",
			config: &Config{
				SecretsManager: SecretsManagerConfig{
					Enabled: false,
				},
			},
			secretsManager: nil,
			envAPIKey:      "env-api-key",
			want:           "env-api-key",
			wantErr:        false,
		},
		{
			name: "No Secrets Manager, no env, returns empty",
			config: &Config{
				SecretsManager: SecretsManagerConfig{
					Enabled: false,
				},
			},
			secretsManager: nil,
			want:           "",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envAPIKey != "" {
				os.Setenv("EXCHANGE_RATE_API_KEY", tt.envAPIKey)
			} else {
				os.Unsetenv("EXCHANGE_RATE_API_KEY")
			}

			got, err := tt.config.GetAPIKey(context.Background(), tt.secretsManager)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Config.GetAPIKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

// mockSecretsManager is a mock implementation of SecretsManager for testing.
type mockSecretsManager struct {
	apiKey string
	err    bool
}

func (m *mockSecretsManager) GetAPIKey(ctx context.Context) (string, error) {
	if m.err {
		return "", fmt.Errorf("mock error")
	}
	return m.apiKey, nil
}
