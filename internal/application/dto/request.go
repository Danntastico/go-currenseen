package dto

// GetRateRequest represents a request to get an exchange rate for a currency pair.
type GetRateRequest struct {
	Base   string `json:"base"`   // Base currency code (e.g., "USD")
	Target string `json:"target"` // Target currency code (e.g., "EUR")
}

// GetRatesRequest represents a request to get all exchange rates for a base currency.
type GetRatesRequest struct {
	Base string `json:"base"` // Base currency code (e.g., "USD")
}

// HealthCheckRequest represents a request for a health check.
// This is typically an empty request, but we define it for consistency.
type HealthCheckRequest struct{}
