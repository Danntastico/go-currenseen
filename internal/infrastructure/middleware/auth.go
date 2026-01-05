package middleware

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/infrastructure/config"
)

// ErrUnauthorized is returned when API key authentication fails.
var ErrUnauthorized = errors.New("unauthorized")

// ErrAPIKeyMissing is returned when no API key is provided in the request.
var ErrAPIKeyMissing = errors.New("api key missing")

// APIKeyAuthenticator handles API key authentication for requests to our service.
//
// This authenticator protects our Lambda endpoints (e.g., /rates/{base}/{target}).
// It does NOT authenticate with the external currency API (which is free and public).
//
// Clients calling our service must provide a valid API key in the X-API-Key header
// or Authorization: Bearer header. The key is validated against the value stored
// in AWS Secrets Manager or the EXCHANGE_RATE_API_KEY environment variable.
type APIKeyAuthenticator struct {
	secretsManager config.SecretsManager
	config         *config.Config
	enabled        bool
}

// NewAPIKeyAuthenticator creates a new API key authenticator.
//
// Parameters:
// - secretsManager: Secrets Manager instance for retrieving API keys
// - cfg: Configuration instance
// - enabled: Whether authentication is enabled (can be disabled for local dev)
//
// If enabled is false, authentication will be skipped.
func NewAPIKeyAuthenticator(secretsManager config.SecretsManager, cfg *config.Config, enabled bool) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{
		secretsManager: secretsManager,
		config:         cfg,
		enabled:        enabled,
	}
}

// ExtractAPIKey extracts the API key from the request headers.
//
// Priority:
// 1. X-API-Key header
// 2. Authorization header (Bearer token format)
//
// Returns the API key or an error if not found.
func ExtractAPIKey(event events.APIGatewayProxyRequest) (string, error) {
	// Try X-API-Key header first
	if apiKey := event.Headers["X-API-Key"]; apiKey != "" {
		return strings.TrimSpace(apiKey), nil
	}
	if apiKey := event.Headers["x-api-key"]; apiKey != "" {
		return strings.TrimSpace(apiKey), nil
	}

	// Try Authorization header (Bearer token)
	if authHeader := event.Headers["Authorization"]; authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1]), nil
		}
	}
	if authHeader := event.Headers["authorization"]; authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1]), nil
		}
	}

	return "", ErrAPIKeyMissing
}

// ValidateAPIKey validates an API key against the stored secret.
//
// Security: Uses constant-time comparison to prevent timing attacks.
func (a *APIKeyAuthenticator) ValidateAPIKey(ctx context.Context, providedKey string) error {
	if !a.enabled {
		// Authentication disabled (e.g., for local development)
		return nil
	}

	if providedKey == "" {
		return ErrAPIKeyMissing
	}

	// Get valid API key from Secrets Manager or environment
	validKey, err := a.config.GetAPIKey(ctx, a.secretsManager)
	if err != nil {
		return fmt.Errorf("failed to retrieve API key: %w", err)
	}

	if validKey == "" {
		// No API key configured - allow request (for development)
		return nil
	}

	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(providedKey), []byte(validKey)) != 1 {
		return ErrUnauthorized
	}

	return nil
}

// AuthenticateRequest authenticates an API Gateway request using API key.
//
// This function:
// 1. Extracts the API key from request headers
// 2. Validates it against the stored secret
// 3. Returns an error if authentication fails
//
// Security: Uses constant-time comparison and never leaks secret information.
func (a *APIKeyAuthenticator) AuthenticateRequest(ctx context.Context, event events.APIGatewayProxyRequest) error {
	// Extract API key from request
	apiKey, err := ExtractAPIKey(event)
	if err != nil {
		return err
	}

	// Validate API key
	return a.ValidateAPIKey(ctx, apiKey)
}
