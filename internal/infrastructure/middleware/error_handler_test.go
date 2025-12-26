package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/application/dto"
	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"nil error", nil, http.StatusOK},
		{"context canceled", context.Canceled, http.StatusRequestTimeout},
		{"context deadline exceeded", context.DeadlineExceeded, http.StatusRequestTimeout},
		{"invalid currency code", entity.ErrInvalidCurrencyCode, http.StatusBadRequest},
		{"currency code mismatch", entity.ErrCurrencyCodeMismatch, http.StatusBadRequest},
		{"rate not found", entity.ErrRateNotFound, http.StatusNotFound},
		{"circuit open", circuitbreaker.ErrCircuitOpen, http.StatusServiceUnavailable},
		{"path parameter error", errors.New("path parameter base not found"), http.StatusBadRequest},
		{"method error", errors.New("method POST not allowed"), http.StatusBadRequest},
		{"unknown error", errors.New("unknown error"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStatusCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("getStatusCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode string
	}{
		{"nil error", nil, ""},
		{"invalid currency code", entity.ErrInvalidCurrencyCode, "INVALID_CURRENCY_CODE"},
		{"currency code mismatch", entity.ErrCurrencyCodeMismatch, "CURRENCY_CODE_MISMATCH"},
		{"rate not found", entity.ErrRateNotFound, "RATE_NOT_FOUND"},
		{"circuit open", circuitbreaker.ErrCircuitOpen, "CIRCUIT_BREAKER_OPEN"},
		{"unknown error", errors.New("unknown"), "INTERNAL_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getErrorCode(tt.err)
			if got != tt.wantCode {
				t.Errorf("getErrorCode() = %q, want %q", got, tt.wantCode)
			}
		})
	}
}

func TestGetClientMessage(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{"nil error", nil, ""},
		{"context canceled", context.Canceled, "Request timeout"},
		{"invalid currency code", entity.ErrInvalidCurrencyCode, "Invalid currency code provided"},
		{"currency code mismatch", entity.ErrCurrencyCodeMismatch, "Base and target currencies cannot be the same"},
		{"rate not found", entity.ErrRateNotFound, "Exchange rate not found"},
		{"circuit open", circuitbreaker.ErrCircuitOpen, "Service temporarily unavailable"},
		{"unknown error", errors.New("internal error"), "An error occurred processing your request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getClientMessage(tt.err)
			if got != tt.wantMsg {
				t.Errorf("getClientMessage() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	err := entity.ErrInvalidCurrencyCode
	resp := ErrorResponse(err)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", resp.Headers["Content-Type"])
	}

	// Parse response body
	var errorResp dto.ErrorResponse
	if err := json.Unmarshal([]byte(resp.Body), &errorResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if errorResp.Error == "" {
		t.Error("Error message is empty")
	}

	if errorResp.Code == "" {
		t.Error("Error code is empty")
	}

	if errorResp.Timestamp.IsZero() {
		t.Error("Timestamp is zero")
	}
}

func TestSuccessResponse(t *testing.T) {
	body := dto.RateResponse{
		Base:      "USD",
		Target:    "EUR",
		Rate:      0.85,
		Timestamp: time.Now(),
		Stale:     false,
	}

	resp := SuccessResponse(http.StatusOK, body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if resp.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", resp.Headers["Content-Type"])
	}

	// Parse response body
	var rateResp dto.RateResponse
	if err := json.Unmarshal([]byte(resp.Body), &rateResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if rateResp.Base != "USD" {
		t.Errorf("Base = %q, want USD", rateResp.Base)
	}

	if rateResp.Rate != 0.85 {
		t.Errorf("Rate = %f, want 0.85", rateResp.Rate)
	}
}

func TestSuccessResponse_DefaultStatusCode(t *testing.T) {
	body := map[string]string{"message": "success"}

	resp := SuccessResponse(0, body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestSuccessResponse_MarshalError(t *testing.T) {
	// Create a body that cannot be marshaled
	body := make(chan int)

	resp := SuccessResponse(http.StatusOK, body)

	// Should return error response
	if resp.StatusCode == http.StatusOK {
		t.Error("Expected error response, got success response")
	}
}
