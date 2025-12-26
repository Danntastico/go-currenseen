package middleware

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestValidateMethod(t *testing.T) {
	tests := []struct {
		name           string
		event          events.APIGatewayProxyRequest
		expectedMethod string
		wantErr        bool
	}{
		{
			name:           "valid GET method",
			event:          events.APIGatewayProxyRequest{HTTPMethod: "GET"},
			expectedMethod: "GET",
			wantErr:        false,
		},
		{
			name:           "invalid POST method",
			event:          events.APIGatewayProxyRequest{HTTPMethod: "POST"},
			expectedMethod: "GET",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMethod(tt.event, tt.expectedMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractPathParameter(t *testing.T) {
	tests := []struct {
		name      string
		event     events.APIGatewayProxyRequest
		paramName string
		want      string
		wantErr   bool
	}{
		{
			name: "valid parameter",
			event: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"base": "USD"},
			},
			paramName: "base",
			want:      "USD",
			wantErr:   false,
		},
		{
			name: "missing parameter",
			event: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{},
			},
			paramName: "base",
			wantErr:   true,
		},
		{
			name: "nil path parameters",
			event: events.APIGatewayProxyRequest{
				PathParameters: nil,
			},
			paramName: "base",
			wantErr:   true,
		},
		{
			name: "empty parameter value",
			event: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"base": ""},
			},
			paramName: "base",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractPathParameter(tt.event, tt.paramName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPathParameter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExtractPathParameter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateCurrencyCode(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid code", "USD", false},
		{"valid code lowercase", "usd", false}, // Should be converted to uppercase
		{"invalid code - too short", "US", true},
		{"invalid code - too long", "USDD", true},
		{"empty code", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateCurrencyCode(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurrencyCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGetRateRequest(t *testing.T) {
	tests := []struct {
		name    string
		event   events.APIGatewayProxyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				PathParameters: map[string]string{
					"base":   "USD",
					"target": "EUR",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid method",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				PathParameters: map[string]string{
					"base":   "USD",
					"target": "EUR",
				},
			},
			wantErr: true,
		},
		{
			name: "missing base parameter",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "GET",
				PathParameters: map[string]string{"target": "EUR"},
			},
			wantErr: true,
		},
		{
			name: "missing target parameter",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "GET",
				PathParameters: map[string]string{"base": "USD"},
			},
			wantErr: true,
		},
		{
			name: "invalid base currency",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				PathParameters: map[string]string{
					"base":   "XX",
					"target": "EUR",
				},
			},
			wantErr: true,
		},
		{
			name: "same base and target",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				PathParameters: map[string]string{
					"base":   "USD",
					"target": "USD",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ValidateGetRateRequest(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGetRateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGetRatesRequest(t *testing.T) {
	tests := []struct {
		name    string
		event   events.APIGatewayProxyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "GET",
				PathParameters: map[string]string{"base": "USD"},
			},
			wantErr: false,
		},
		{
			name: "invalid method",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "POST",
				PathParameters: map[string]string{"base": "USD"},
			},
			wantErr: true,
		},
		{
			name: "missing base parameter",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "GET",
				PathParameters: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "invalid base currency",
			event: events.APIGatewayProxyRequest{
				HTTPMethod:     "GET",
				PathParameters: map[string]string{"base": "XX"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateGetRatesRequest(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGetRatesRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHealthRequest(t *testing.T) {
	tests := []struct {
		name    string
		event   events.APIGatewayProxyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
			},
			wantErr: false,
		},
		{
			name: "invalid method",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHealthRequest(tt.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHealthRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name           string
		event          events.APIGatewayProxyRequest
		expectedMethod string
		wantErr        bool
	}{
		{
			name: "valid request",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Body:       "small body",
			},
			expectedMethod: "GET",
			wantErr:        false,
		},
		{
			name: "invalid method",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
			},
			expectedMethod: "GET",
			wantErr:        true,
		},
		{
			name: "oversized body",
			event: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Body:       string(make([]byte, 11*1024*1024)), // 11MB
			},
			expectedMethod: "GET",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.event, tt.expectedMethod)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
