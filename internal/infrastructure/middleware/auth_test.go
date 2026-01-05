package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/config"
)

// mockSecretsManager is a mock implementation of SecretsManager for testing.
type mockSecretsManager struct {
	apiKey string
	err    error
}

func (m *mockSecretsManager) GetAPIKey(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.apiKey, nil
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name        string
		event       events.APIGatewayProxyRequest
		expectedKey string
		expectedErr error
	}{
		{
			name: "X-API-Key header present",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"X-API-Key": "test-key-123",
				},
			},
			expectedKey: "test-key-123",
			expectedErr: nil,
		},
		{
			name: "x-api-key header present (lowercase)",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"x-api-key": "test-key-456",
				},
			},
			expectedKey: "test-key-456",
			expectedErr: nil,
		},
		{
			name: "Authorization Bearer header present",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"Authorization": "Bearer test-key-789",
				},
			},
			expectedKey: "test-key-789",
			expectedErr: nil,
		},
		{
			name: "authorization Bearer header present (lowercase)",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Bearer test-key-abc",
				},
			},
			expectedKey: "test-key-abc",
			expectedErr: nil,
		},
		{
			name: "X-API-Key takes priority over Authorization",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"X-API-Key":     "priority-key",
					"Authorization": "Bearer bearer-key",
				},
			},
			expectedKey: "priority-key",
			expectedErr: nil,
		},
		{
			name: "no API key present",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{},
			},
			expectedKey: "",
			expectedErr: ErrAPIKeyMissing,
		},
		{
			name: "Authorization header without Bearer",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"Authorization": "Basic dGVzdDp0ZXN0",
				},
			},
			expectedKey: "",
			expectedErr: ErrAPIKeyMissing,
		},
		{
			name: "API key with whitespace",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"X-API-Key": "  test-key  ",
				},
			},
			expectedKey: "test-key",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ExtractAPIKey(tt.event)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if key != tt.expectedKey {
					t.Errorf("expected API key %q, got %q", tt.expectedKey, key)
				}
			}
		})
	}
}

func TestAPIKeyAuthenticator_ValidateAPIKey(t *testing.T) {
	tests := []struct {
		name           string
		enabled        bool
		providedKey    string
		validKey       string
		secretsErr     error
		secretsEnabled bool
		expectedErr    error
		description    string
	}{
		{
			name:        "authentication disabled",
			enabled:     false,
			providedKey: "any-key",
			validKey:    "valid-key",
			expectedErr: nil,
			description: "should allow any key when disabled",
		},
		{
			name:           "valid API key",
			enabled:        true,
			providedKey:    "valid-key",
			validKey:       "valid-key",
			secretsEnabled: true,
			expectedErr:    nil,
			description:    "should accept valid key",
		},
		{
			name:           "invalid API key",
			enabled:        true,
			providedKey:    "invalid-key",
			validKey:       "valid-key",
			secretsEnabled: true,
			expectedErr:    ErrUnauthorized,
			description:    "should reject invalid key",
		},
		{
			name:           "empty provided key",
			enabled:        true,
			providedKey:    "",
			validKey:       "valid-key",
			secretsEnabled: true,
			expectedErr:    ErrAPIKeyMissing,
			description:    "should reject empty key",
		},
		{
			name:           "no valid key configured",
			enabled:        true,
			providedKey:    "any-key",
			validKey:       "",
			secretsEnabled: false,
			expectedErr:    nil,
			description:    "should allow when no key configured (dev mode)",
		},
		{
			name:           "secrets manager error",
			enabled:        true,
			providedKey:    "any-key",
			validKey:       "",
			secretsErr:     errors.New("secrets error"),
			secretsEnabled: true,
			expectedErr:    nil, // Config.GetAPIKey falls back to env var, so no error
			description:    "should fallback to env var when secrets manager fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &mockSecretsManager{
				apiKey: tt.validKey,
				err:    tt.secretsErr,
			}

			cfg := &config.Config{
				SecretsManager: config.SecretsManagerConfig{
					Enabled: tt.secretsEnabled,
				},
			}

			authenticator := NewAPIKeyAuthenticator(sm, cfg, tt.enabled)

			err := authenticator.ValidateAPIKey(context.Background(), tt.providedKey)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if !errors.Is(err, tt.expectedErr) && !errors.Is(err, ErrUnauthorized) && !errors.Is(err, ErrAPIKeyMissing) {
					// Check if error message contains expected text
					if tt.secretsErr != nil && !containsString(err.Error(), "failed to retrieve API key") {
						t.Errorf("expected error containing 'failed to retrieve API key', got %v", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAPIKeyAuthenticator_AuthenticateRequest(t *testing.T) {
	sm := &mockSecretsManager{
		apiKey: "valid-key",
	}

	cfg := &config.Config{
		SecretsManager: config.SecretsManagerConfig{
			Enabled: true,
		},
	}

	authenticator := NewAPIKeyAuthenticator(sm, cfg, true)

	tests := []struct {
		name        string
		event       events.APIGatewayProxyRequest
		expectedErr error
	}{
		{
			name: "valid API key in X-API-Key header",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"X-API-Key": "valid-key",
				},
			},
			expectedErr: nil,
		},
		{
			name: "invalid API key",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"X-API-Key": "invalid-key",
				},
			},
			expectedErr: ErrUnauthorized,
		},
		{
			name: "missing API key",
			event: events.APIGatewayProxyRequest{
				Headers: map[string]string{},
			},
			expectedErr: ErrAPIKeyMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authenticator.AuthenticateRequest(context.Background(), tt.event)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedErr)
				} else if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
