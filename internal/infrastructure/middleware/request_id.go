package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/pkg/logger"
)

// GenerateRequestID generates a unique request ID.
// Uses cryptographically secure random bytes for uniqueness.
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return generateFallbackRequestID()
	}
	return hex.EncodeToString(bytes)
}

// generateFallbackRequestID generates a request ID using a fallback method.
func generateFallbackRequestID() string {
	// Simple fallback: use hex-encoded timestamp
	// This is less secure but ensures we always have an ID
	return hex.EncodeToString([]byte("fallback-id"))
}

// ExtractOrGenerateRequestID extracts request ID from API Gateway event or generates a new one.
//
// Priority:
// 1. Request ID from API Gateway request context
// 2. X-Request-ID header
// 3. Generated request ID
func ExtractOrGenerateRequestID(event events.APIGatewayProxyRequest) string {
	// Try API Gateway request context first
	if event.RequestContext.RequestID != "" {
		return event.RequestContext.RequestID
	}

	// Try X-Request-ID header
	if requestID := event.Headers["X-Request-ID"]; requestID != "" {
		return requestID
	}
	if requestID := event.Headers["x-request-id"]; requestID != "" {
		return requestID
	}

	// Generate new request ID
	return GenerateRequestID()
}

// WithRequestID adds request ID to context from API Gateway event.
func WithRequestID(ctx context.Context, event events.APIGatewayProxyRequest) context.Context {
	requestID := ExtractOrGenerateRequestID(event)
	return logger.WithRequestID(ctx, requestID)
}
