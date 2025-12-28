package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/misterfancybg/go-currenseen/internal/domain/entity"
)

func TestNewCurrencyAPIProvider(t *testing.T) {
	client := NewHTTPClient()
	baseURL := "https://api.example.com/v1"

	provider := NewCurrencyAPIProvider(client, baseURL)

	if provider == nil {
		t.Fatal("NewCurrencyAPIProvider() returned nil")
	}

	if provider.client != client {
		t.Error("client not set correctly")
	}

	if provider.baseURL != baseURL {
		t.Errorf("baseURL = %q, want %q", provider.baseURL, baseURL)
	}
}

func TestNewCurrencyAPIProvider_DefaultBaseURL(t *testing.T) {
	client := NewHTTPClient()

	provider := NewCurrencyAPIProvider(client, "")

	expectedURL := "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1"
	if provider.baseURL != expectedURL {
		t.Errorf("baseURL = %q, want %q", provider.baseURL, expectedURL)
	}
}

func TestCurrencyAPIProvider_FetchRate_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// New API format: /currencies/{base}.json
		if r.URL.Path != "/currencies/usd.json" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// New API response format: rates nested under base currency (lowercase)
		response := map[string]interface{}{
			"date": "2024-01-15",
			"usd": map[string]float64{
				"eur": 0.85,
				"gbp": 0.75,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create provider with test server URL
	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	ctx := context.Background()
	rate, err := provider.FetchRate(ctx, base, target)

	if err != nil {
		t.Fatalf("FetchRate() error = %v, want nil", err)
	}

	if rate == nil {
		t.Fatal("FetchRate() returned nil rate")
	}

	if !rate.Base.Equal(base) {
		t.Errorf("Base = %v, want %v", rate.Base, base)
	}

	if !rate.Target.Equal(target) {
		t.Errorf("Target = %v, want %v", rate.Target, target)
	}

	if rate.Rate != 0.85 {
		t.Errorf("Rate = %f, want 0.85", rate.Rate)
	}

	if rate.Stale {
		t.Error("Stale = true, want false")
	}
}

func TestCurrencyAPIProvider_FetchRate_HTTPError(t *testing.T) {
	// Create mock server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	ctx := context.Background()
	_, err := provider.FetchRate(ctx, base, target)

	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}
}

func TestCurrencyAPIProvider_FetchRate_InvalidJSON(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	ctx := context.Background()
	_, err := provider.FetchRate(ctx, base, target)

	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}
}

func TestCurrencyAPIProvider_FetchRate_ContextCancellation(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.FetchRate(ctx, base, target)

	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}

	if err != context.Canceled {
		t.Errorf("Error = %v, want context.Canceled", err)
	}
}

func TestCurrencyAPIProvider_FetchRate_ContextTimeout(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")
	target, _ := entity.NewCurrencyCode("EUR")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := provider.FetchRate(ctx, base, target)

	if err == nil {
		t.Fatal("FetchRate() error = nil, want error")
	}

	// HTTP client wraps context.DeadlineExceeded, so check with errors.Is
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Error = %v, want context.DeadlineExceeded (or wrapped)", err)
	}
}

func TestCurrencyAPIProvider_FetchAllRates_Success(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// New API format: /currencies/{base}.json
		if r.URL.Path != "/currencies/usd.json" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		// New API response format
		response := map[string]interface{}{
			"date": "2024-01-15",
			"usd": map[string]float64{
				"eur": 0.85,
				"gbp": 0.75,
				"jpy": 110.50,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")

	ctx := context.Background()
	rates, err := provider.FetchAllRates(ctx, base)

	if err != nil {
		t.Fatalf("FetchAllRates() error = %v, want nil", err)
	}

	if len(rates) != 3 {
		t.Errorf("len(rates) = %d, want 3", len(rates))
	}

	// Verify all rates have correct base
	for _, rate := range rates {
		if !rate.Base.Equal(base) {
			t.Errorf("Rate base = %v, want %v", rate.Base, base)
		}

		if rate.Stale {
			t.Error("Stale = true, want false")
		}
	}
}

func TestCurrencyAPIProvider_FetchAllRates_EmptyRates(t *testing.T) {
	// Create mock server that returns empty rates
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// New API response format
		response := map[string]interface{}{
			"date": "2024-01-15",
			"usd":  map[string]float64{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")

	ctx := context.Background()
	rates, err := provider.FetchAllRates(ctx, base)

	if err != nil {
		t.Fatalf("FetchAllRates() error = %v, want nil", err)
	}

	// Should return empty slice (not nil)
	if rates == nil {
		t.Error("rates is nil, want empty slice")
	}

	if len(rates) != 0 {
		t.Errorf("len(rates) = %d, want 0", len(rates))
	}
}

func TestCurrencyAPIProvider_FetchAllRates_APIError(t *testing.T) {
	// Create mock server that returns API error
	// Note: New API doesn't have an "error" field in the same way
	// We'll return an empty response or invalid structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty response (no base currency found)
		response := map[string]interface{}{
			"date": "2024-01-15",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")

	ctx := context.Background()
	_, err := provider.FetchAllRates(ctx, base)

	if err == nil {
		t.Fatal("FetchAllRates() error = nil, want error")
	}
}

func TestCurrencyAPIProvider_FetchAllRates_HTTPError(t *testing.T) {
	// Create mock server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")

	ctx := context.Background()
	_, err := provider.FetchAllRates(ctx, base)

	if err == nil {
		t.Fatal("FetchAllRates() error = nil, want error")
	}
}

func TestCurrencyAPIProvider_FetchAllRates_ContextCancellation(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient()
	provider := NewCurrencyAPIProvider(client, server.URL)

	base, _ := entity.NewCurrencyCode("USD")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := provider.FetchAllRates(ctx, base)

	if err == nil {
		t.Fatal("FetchAllRates() error = nil, want error")
	}

	if err != context.Canceled {
		t.Errorf("Error = %v, want context.Canceled", err)
	}
}
