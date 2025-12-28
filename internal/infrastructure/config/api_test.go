package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadAPIConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("EXCHANGE_RATE_API_URL")
	os.Unsetenv("EXCHANGE_RATE_API_TIMEOUT")
	os.Unsetenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS")

	cfg := LoadAPIConfig()

	expectedURL := "https://cdn.jsdelivr.net/npm/@fawazahmed0/currency-api@latest/v1"
	if cfg.BaseURL != expectedURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, expectedURL)
	}

	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", cfg.Timeout)
	}

	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3", cfg.RetryAttempts)
	}
}

func TestLoadAPIConfig_CustomBaseURL(t *testing.T) {
	// Set custom base URL
	os.Setenv("EXCHANGE_RATE_API_URL", "https://api.example.com/v1")
	defer os.Unsetenv("EXCHANGE_RATE_API_URL")

	cfg := LoadAPIConfig()

	expectedURL := "https://api.example.com/v1"
	if cfg.BaseURL != expectedURL {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, expectedURL)
	}
}

func TestLoadAPIConfig_CustomTimeout(t *testing.T) {
	// Set custom timeout
	os.Setenv("EXCHANGE_RATE_API_TIMEOUT", "30")
	defer os.Unsetenv("EXCHANGE_RATE_API_TIMEOUT")

	cfg := LoadAPIConfig()

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", cfg.Timeout)
	}
}

func TestLoadAPIConfig_CustomRetryAttempts(t *testing.T) {
	// Set custom retry attempts
	os.Setenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS", "5")
	defer os.Unsetenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS")

	cfg := LoadAPIConfig()

	if cfg.RetryAttempts != 5 {
		t.Errorf("RetryAttempts = %d, want 5", cfg.RetryAttempts)
	}
}

func TestLoadAPIConfig_InvalidTimeout(t *testing.T) {
	// Set invalid timeout (should use default)
	os.Setenv("EXCHANGE_RATE_API_TIMEOUT", "invalid")
	defer os.Unsetenv("EXCHANGE_RATE_API_TIMEOUT")

	cfg := LoadAPIConfig()

	// Should use default
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s (default)", cfg.Timeout)
	}
}

func TestLoadAPIConfig_InvalidRetryAttempts(t *testing.T) {
	// Set invalid retry attempts (should use default)
	os.Setenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS", "invalid")
	defer os.Unsetenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS")

	cfg := LoadAPIConfig()

	// Should use default
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3 (default)", cfg.RetryAttempts)
	}
}

func TestLoadAPIConfig_ZeroTimeout(t *testing.T) {
	// Set zero timeout (should use default)
	os.Setenv("EXCHANGE_RATE_API_TIMEOUT", "0")
	defer os.Unsetenv("EXCHANGE_RATE_API_TIMEOUT")

	cfg := LoadAPIConfig()

	// Should use default (zero is invalid)
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s (default)", cfg.Timeout)
	}
}

func TestLoadAPIConfig_ZeroRetryAttempts(t *testing.T) {
	// Set zero retry attempts (should use default)
	os.Setenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS", "0")
	defer os.Unsetenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS")

	cfg := LoadAPIConfig()

	// Should use default (zero is invalid)
	if cfg.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3 (default)", cfg.RetryAttempts)
	}
}

func TestLoadAPIConfig_AllCustom(t *testing.T) {
	// Set all custom values
	os.Setenv("EXCHANGE_RATE_API_URL", "https://api.custom.com/v1")
	os.Setenv("EXCHANGE_RATE_API_TIMEOUT", "20")
	os.Setenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS", "4")
	defer func() {
		os.Unsetenv("EXCHANGE_RATE_API_URL")
		os.Unsetenv("EXCHANGE_RATE_API_TIMEOUT")
		os.Unsetenv("EXCHANGE_RATE_API_RETRY_ATTEMPTS")
	}()

	cfg := LoadAPIConfig()

	if cfg.BaseURL != "https://api.custom.com/v1" {
		t.Errorf("BaseURL = %q, want https://api.custom.com/v1", cfg.BaseURL)
	}

	if cfg.Timeout != 20*time.Second {
		t.Errorf("Timeout = %v, want 20s", cfg.Timeout)
	}

	if cfg.RetryAttempts != 4 {
		t.Errorf("RetryAttempts = %d, want 4", cfg.RetryAttempts)
	}
}
