package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/aws/aws-lambda-go/events"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

// ValidateMethod validates that the HTTP method matches the expected method.
//
// Returns an error if the method doesn't match.
func ValidateMethod(event events.APIGatewayProxyRequest, expectedMethod string) error {
	if event.HTTPMethod != expectedMethod {
		return fmt.Errorf("method %s not allowed, expected %s", event.HTTPMethod, expectedMethod)
	}
	return nil
}

// ExtractPathParameter extracts a path parameter from the API Gateway event.
//
// Returns an error if the parameter is missing or empty.
func ExtractPathParameter(event events.APIGatewayProxyRequest, paramName string) (string, error) {
	if event.PathParameters == nil {
		return "", fmt.Errorf("path parameter %s not found", paramName)
	}

	value, ok := event.PathParameters[paramName]
	if !ok || value == "" {
		return "", fmt.Errorf("path parameter %s not found or empty", paramName)
	}

	return value, nil
}

// ValidateCurrencyCode validates a currency code string.
//
// This function:
// - Validates the currency code format using domain validation
// - Returns a domain error if invalid
//
// Security: Validates input before processing to prevent injection attacks.
func ValidateCurrencyCode(code string) (entity.CurrencyCode, error) {
	if code == "" {
		var zero entity.CurrencyCode
		return zero, entity.ErrInvalidCurrencyCode
	}

	currencyCode, err := entity.NewCurrencyCode(code)
	if err != nil {
		var zero entity.CurrencyCode
		return zero, fmt.Errorf("invalid currency code %s: %w", code, err)
	}

	return currencyCode, nil
}

// ValidateGetRateRequest validates a GET /rates/{base}/{target} request.
//
// This function:
// - Validates HTTP method is GET
// - Extracts and validates base currency code
// - Extracts and validates target currency code
// - Validates base and target are different
//
// Returns domain errors for invalid input.
func ValidateGetRateRequest(event events.APIGatewayProxyRequest) (base, target entity.CurrencyCode, err error) {
	var zero entity.CurrencyCode

	// Validate HTTP method
	if err := ValidateMethod(event, http.MethodGet); err != nil {
		return zero, zero, err
	}

	// Extract base currency
	baseStr, err := ExtractPathParameter(event, "base")
	if err != nil {
		return zero, zero, fmt.Errorf("base currency: %w", err)
	}

	base, err = ValidateCurrencyCode(baseStr)
	if err != nil {
		return zero, zero, err
	}

	// Extract target currency
	targetStr, err := ExtractPathParameter(event, "target")
	if err != nil {
		return zero, zero, fmt.Errorf("target currency: %w", err)
	}

	target, err = ValidateCurrencyCode(targetStr)
	if err != nil {
		return zero, zero, err
	}

	// Validate base and target are different
	if base.Equal(target) {
		return zero, zero, entity.ErrCurrencyCodeMismatch
	}

	return base, target, nil
}

// ValidateGetRatesRequest validates a GET /rates/{base} request.
//
// This function:
// - Validates HTTP method is GET
// - Extracts and validates base currency code
//
// Returns domain errors for invalid input.
func ValidateGetRatesRequest(event events.APIGatewayProxyRequest) (base entity.CurrencyCode, err error) {
	var zero entity.CurrencyCode

	// Validate HTTP method
	if err := ValidateMethod(event, http.MethodGet); err != nil {
		return zero, err
	}

	// Extract base currency
	baseStr, err := ExtractPathParameter(event, "base")
	if err != nil {
		return zero, fmt.Errorf("base currency: %w", err)
	}

	base, err = ValidateCurrencyCode(baseStr)
	if err != nil {
		return zero, err
	}

	return base, nil
}

// ValidateHealthRequest validates a GET /health request.
//
// This function:
// - Validates HTTP method is GET
//
// Returns an error if validation fails.
func ValidateHealthRequest(event events.APIGatewayProxyRequest) error {
	// Validate HTTP method
	if err := ValidateMethod(event, http.MethodGet); err != nil {
		return err
	}

	// Health endpoint has no parameters, so no further validation needed
	return nil
}

// SanitizePathParameter sanitizes a path parameter by removing control characters
// and normalizing Unicode characters to prevent injection attacks.
//
// Security: This function helps prevent path traversal and injection attacks
// by removing potentially dangerous characters.
func SanitizePathParameter(param string) string {
	// Remove control characters and normalize
	var builder strings.Builder
	for _, r := range param {
		// Allow alphanumeric and common safe characters
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// ValidateRequestSize validates that the request body size is within limits.
//
// Security: Prevents oversized requests that could cause DoS attacks.
func ValidateRequestSize(event events.APIGatewayProxyRequest) error {
	const maxRequestSize = 10 * 1024 * 1024 // 10MB (API Gateway default)
	if len(event.Body) > maxRequestSize {
		return errors.New("request body too large")
	}
	return nil
}

// ValidateRequest is a generic request validator that checks basic request properties.
//
// This function:
// - Validates HTTP method
// - Validates request size (if needed)
// - Returns errors for invalid requests
func ValidateRequest(event events.APIGatewayProxyRequest, expectedMethod string) error {
	// Validate HTTP method
	if err := ValidateMethod(event, expectedMethod); err != nil {
		return err
	}

	// Validate request size (security: prevent oversized requests)
	if err := ValidateRequestSize(event); err != nil {
		return err
	}

	return nil
}
