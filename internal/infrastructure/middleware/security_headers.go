package middleware

import (
	"github.com/aws/aws-lambda-go/events"
)

// SecurityHeaders returns a map of security headers to add to responses.
//
// Headers included:
// - Strict-Transport-Security (HSTS): Enforces HTTPS
// - X-Content-Type-Options: Prevents MIME type sniffing
// - X-Frame-Options: Prevents clickjacking
// - X-XSS-Protection: Enables XSS filter (legacy, but still useful)
// - Content-Security-Policy: Restricts resource loading
// - Referrer-Policy: Controls referrer information
//
// Security: These headers help protect against common web vulnerabilities.
func SecurityHeaders() map[string]string {
	return map[string]string{
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Content-Security-Policy":   "default-src 'self'",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
	}
}

// AddSecurityHeaders adds security headers to an API Gateway response.
//
// This function merges security headers with existing headers in the response.
// If a header already exists, it will be overwritten with the security header value.
func AddSecurityHeaders(resp events.APIGatewayProxyResponse) events.APIGatewayProxyResponse {
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}

	// Add all security headers
	securityHeaders := SecurityHeaders()
	for key, value := range securityHeaders {
		resp.Headers[key] = value
	}

	return resp
}
