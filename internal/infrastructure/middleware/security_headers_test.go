package middleware

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestSecurityHeaders(t *testing.T) {
	headers := SecurityHeaders()

	// Verify all expected headers are present
	expectedHeaders := []string{
		"Strict-Transport-Security",
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Content-Security-Policy",
		"Referrer-Policy",
	}

	for _, header := range expectedHeaders {
		if _, ok := headers[header]; !ok {
			t.Errorf("expected security header %s not found", header)
		}
		if headers[header] == "" {
			t.Errorf("security header %s has empty value", header)
		}
	}

	// Verify specific header values
	if headers["X-Frame-Options"] != "DENY" {
		t.Errorf("expected X-Frame-Options to be 'DENY', got %q", headers["X-Frame-Options"])
	}
	if headers["X-Content-Type-Options"] != "nosniff" {
		t.Errorf("expected X-Content-Type-Options to be 'nosniff', got %q", headers["X-Content-Type-Options"])
	}
}

func TestAddSecurityHeaders(t *testing.T) {
	tests := []struct {
		name           string
		response       events.APIGatewayProxyResponse
		expectedHeader string
		expectedValue  string
	}{
		{
			name: "response with nil headers",
			response: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers:    nil,
			},
			expectedHeader: "X-Frame-Options",
			expectedValue:  "DENY",
		},
		{
			name: "response with existing headers",
			response: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			expectedHeader: "X-Frame-Options",
			expectedValue:  "DENY",
		},
		{
			name: "response with conflicting header",
			response: events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"X-Frame-Options": "SAMEORIGIN", // Should be overwritten
				},
			},
			expectedHeader: "X-Frame-Options",
			expectedValue:  "DENY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddSecurityHeaders(tt.response)

			if result.Headers == nil {
				t.Fatal("expected headers map to be initialized")
			}

			value, ok := result.Headers[tt.expectedHeader]
			if !ok {
				t.Errorf("expected header %s not found in response", tt.expectedHeader)
			}
			if value != tt.expectedValue {
				t.Errorf("expected header %s to be %q, got %q", tt.expectedHeader, tt.expectedValue, value)
			}

			// Verify all security headers are present
			securityHeaders := SecurityHeaders()
			for key, expectedValue := range securityHeaders {
				if actualValue, ok := result.Headers[key]; !ok {
					t.Errorf("expected security header %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("expected security header %s to be %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}
