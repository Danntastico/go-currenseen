package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// getStatusCode maps domain errors to HTTP status codes.
//
// This function:
// - Maps domain errors to appropriate HTTP status codes
// - Maps validation errors (path parameter, method) to 400
// - Returns 500 for unknown errors (internal server error)
// - Preserves context cancellation errors
//
// Security: Returns generic error messages to clients, not internal details.
func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout
	}

	// Map domain errors
	if errors.Is(err, entity.ErrInvalidCurrencyCode) {
		return http.StatusBadRequest
	}
	if errors.Is(err, entity.ErrCurrencyCodeMismatch) {
		return http.StatusBadRequest
	}
	if errors.Is(err, entity.ErrRateNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		return http.StatusServiceUnavailable
	}

	// Check for validation errors (path parameter, method validation)
	errMsg := err.Error()
	if contains(errMsg, "path parameter") || contains(errMsg, "method") || contains(errMsg, "not allowed") {
		return http.StatusBadRequest
	}

	// Default to internal server error for unknown errors
	return http.StatusInternalServerError
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// getErrorCode maps domain errors to error codes for client handling.
//
// Returns a string code that clients can use to handle errors programmatically.
func getErrorCode(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, entity.ErrInvalidCurrencyCode) {
		return "INVALID_CURRENCY_CODE"
	}
	if errors.Is(err, entity.ErrCurrencyCodeMismatch) {
		return "CURRENCY_CODE_MISMATCH"
	}
	if errors.Is(err, entity.ErrRateNotFound) {
		return "RATE_NOT_FOUND"
	}
	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		return "CIRCUIT_BREAKER_OPEN"
	}

	return "INTERNAL_ERROR"
}

// getClientMessage maps internal errors to safe client-facing messages.
//
// Security: Never exposes internal error details, stack traces, or system information.
// Returns generic, user-friendly error messages.
func getClientMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for context cancellation
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "Request timeout"
	}

	// Map domain errors to client messages
	if errors.Is(err, entity.ErrInvalidCurrencyCode) {
		return "Invalid currency code provided"
	}
	if errors.Is(err, entity.ErrCurrencyCodeMismatch) {
		return "Base and target currencies cannot be the same"
	}
	if errors.Is(err, entity.ErrRateNotFound) {
		return "Exchange rate not found"
	}
	if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
		return "Service temporarily unavailable"
	}

	// Generic message for unknown errors (security: don't leak internal details)
	return "An error occurred processing your request"
}

// ErrorResponse creates an error response for API Gateway.
//
// This function:
// - Maps errors to appropriate HTTP status codes
// - Returns safe client messages (not internal details)
// - Includes error codes for programmatic handling
// - Sets proper headers
//
// Security: Never exposes internal error details to clients.
func ErrorResponse(err error) events.APIGatewayProxyResponse {
	statusCode := getStatusCode(err)
	errorCode := getErrorCode(err)
	clientMessage := getClientMessage(err)

	errorResp := dto.ErrorResponse{
		Error:     clientMessage,
		Code:      errorCode,
		Timestamp: time.Now(),
	}

	body, marshalErr := json.Marshal(errorResp)
	if marshalErr != nil {
		// Fallback if JSON marshaling fails
		body = []byte(fmt.Sprintf(`{"error":"%s","timestamp":"%s"}`, clientMessage, time.Now().Format(time.RFC3339)))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// SuccessResponse creates a success response for API Gateway.
//
// This function:
// - Marshals the response body to JSON
// - Sets proper headers
// - Returns 200 status code
func SuccessResponse(statusCode int, body interface{}) events.APIGatewayProxyResponse {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		// If marshaling fails, return error response
		return ErrorResponse(fmt.Errorf("failed to marshal response: %w", err))
	}

	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonBody),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}
