package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadCircuitBreakerConfig_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD")
	os.Unsetenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS")
	os.Unsetenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")

	cfg := LoadCircuitBreakerConfig()

	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", cfg.FailureThreshold)
	}

	if cfg.CooldownDuration != 30*time.Second {
		t.Errorf("CooldownDuration = %v, want 30s", cfg.CooldownDuration)
	}

	if cfg.SuccessThreshold != 1 {
		t.Errorf("SuccessThreshold = %d, want 1", cfg.SuccessThreshold)
	}
}

func TestLoadCircuitBreakerConfig_CustomValues(t *testing.T) {
	// Set custom values
	os.Setenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD", "10")
	os.Setenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS", "60")
	os.Setenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD", "2")
	defer func() {
		os.Unsetenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD")
		os.Unsetenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS")
		os.Unsetenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")
	}()

	cfg := LoadCircuitBreakerConfig()

	if cfg.FailureThreshold != 10 {
		t.Errorf("FailureThreshold = %d, want 10", cfg.FailureThreshold)
	}

	if cfg.CooldownDuration != 60*time.Second {
		t.Errorf("CooldownDuration = %v, want 60s", cfg.CooldownDuration)
	}

	if cfg.SuccessThreshold != 2 {
		t.Errorf("SuccessThreshold = %d, want 2", cfg.SuccessThreshold)
	}
}

func TestLoadCircuitBreakerConfig_InvalidValues(t *testing.T) {
	// Set invalid values (should use defaults)
	os.Setenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD", "invalid")
	os.Setenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS", "invalid")
	os.Setenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD", "invalid")
	defer func() {
		os.Unsetenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD")
		os.Unsetenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS")
		os.Unsetenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")
	}()

	cfg := LoadCircuitBreakerConfig()

	// Should use defaults
	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5 (default)", cfg.FailureThreshold)
	}

	if cfg.CooldownDuration != 30*time.Second {
		t.Errorf("CooldownDuration = %v, want 30s (default)", cfg.CooldownDuration)
	}

	if cfg.SuccessThreshold != 1 {
		t.Errorf("SuccessThreshold = %d, want 1 (default)", cfg.SuccessThreshold)
	}
}

func TestLoadCircuitBreakerConfig_ZeroValues(t *testing.T) {
	// Set zero values (should use defaults)
	os.Setenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD", "0")
	os.Setenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS", "0")
	os.Setenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD", "0")
	defer func() {
		os.Unsetenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD")
		os.Unsetenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS")
		os.Unsetenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD")
	}()

	cfg := LoadCircuitBreakerConfig()

	// Should use defaults (zero is invalid)
	if cfg.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5 (default)", cfg.FailureThreshold)
	}

	if cfg.CooldownDuration != 30*time.Second {
		t.Errorf("CooldownDuration = %v, want 30s (default)", cfg.CooldownDuration)
	}

	if cfg.SuccessThreshold != 1 {
		t.Errorf("SuccessThreshold = %d, want 1 (default)", cfg.SuccessThreshold)
	}
}
