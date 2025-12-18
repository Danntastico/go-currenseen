package dto

import "time"

// RateResponse represents a single exchange rate response.
type RateResponse struct {
	Base      string    `json:"base"`            // Base currency code
	Target    string    `json:"target"`          // Target currency code
	Rate      float64   `json:"rate"`            // Exchange rate
	Timestamp time.Time `json:"timestamp"`       // When the rate was last updated
	Stale     bool      `json:"stale,omitempty"` // Indicates if the rate is stale (from cache fallback)
}

// RatesResponse represents a response containing multiple exchange rates.
type RatesResponse struct {
	Base      string                  `json:"base"`            // Base currency code
	Rates     map[string]RateResponse `json:"rates"`           // Map of target currency to rate
	Timestamp time.Time               `json:"timestamp"`       // When the rates were last updated
	Stale     bool                    `json:"stale,omitempty"` // Indicates if any rate is stale
}

// HealthCheckResponse represents the health status of the service.
type HealthCheckResponse struct {
	Status    string            `json:"status"`           // Overall status: "healthy" or "unhealthy"
	Checks    map[string]string `json:"checks,omitempty"` // Individual component checks
	Timestamp time.Time         `json:"timestamp"`        // When the health check was performed
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error     string    `json:"error"`          // Error message
	Code      string    `json:"code,omitempty"` // Error code (e.g., "RATE_NOT_FOUND")
	Timestamp time.Time `json:"timestamp"`      // When the error occurred
}
