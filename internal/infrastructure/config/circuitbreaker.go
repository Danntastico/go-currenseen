package config

import (
	"os"
	"strconv"
	"time"

	"github.com/misterfancybg/go-currenseen/pkg/circuitbreaker"
)

// LoadCircuitBreakerConfig loads circuit breaker configuration from environment variables.
//
// Environment variables:
// - CIRCUIT_BREAKER_FAILURE_THRESHOLD: Number of failures before opening (default: 5)
// - CIRCUIT_BREAKER_COOLDOWN_SECONDS: Cooldown duration in seconds (default: 30)
// - CIRCUIT_BREAKER_SUCCESS_THRESHOLD: Successes needed in HalfOpen to close (default: 1)
//
// Returns a circuitbreaker.Config with defaults if environment variables are not set.
//
// Example usage:
//
//	cfg := LoadCircuitBreakerConfig()
//	cb, err := circuitbreaker.NewCircuitBreaker(cfg)
func LoadCircuitBreakerConfig() circuitbreaker.Config {
	// Load failure threshold from environment
	failureThreshold := 5 // default
	if thresholdStr := os.Getenv("CIRCUIT_BREAKER_FAILURE_THRESHOLD"); thresholdStr != "" {
		if parsed, err := strconv.Atoi(thresholdStr); err == nil && parsed > 0 {
			failureThreshold = parsed
		}
	}

	// Load cooldown duration from environment (in seconds)
	cooldownSeconds := 30 // default
	if cooldownStr := os.Getenv("CIRCUIT_BREAKER_COOLDOWN_SECONDS"); cooldownStr != "" {
		if parsed, err := strconv.Atoi(cooldownStr); err == nil && parsed > 0 {
			cooldownSeconds = parsed
		}
	}

	// Load success threshold from environment
	successThreshold := 1 // default
	if successStr := os.Getenv("CIRCUIT_BREAKER_SUCCESS_THRESHOLD"); successStr != "" {
		if parsed, err := strconv.Atoi(successStr); err == nil && parsed > 0 {
			successThreshold = parsed
		}
	}

	return circuitbreaker.Config{
		FailureThreshold: failureThreshold,
		CooldownDuration: time.Duration(cooldownSeconds) * time.Second,
		SuccessThreshold: successThreshold,
	}
}
