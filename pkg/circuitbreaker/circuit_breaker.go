package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in Open state
// and requests are not allowed.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State int

const (
	// StateClosed represents the normal operating state.
	// All requests pass through. Failures are counted.
	StateClosed State = iota

	// StateOpen represents the failing state.
	// All requests fail immediately without calling the external service.
	// After cooldown period, transitions to HalfOpen.
	StateOpen

	// StateHalfOpen represents the testing state.
	// Allows one test request to check if the service has recovered.
	// If test succeeds, transitions to Closed. If fails, transitions back to Open.
	StateHalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "Closed"
	case StateOpen:
		return "Open"
	case StateHalfOpen:
		return "HalfOpen"
	default:
		return "Unknown"
	}
}

// Config holds circuit breaker configuration.
type Config struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	// Default: 5
	FailureThreshold int

	// CooldownDuration is the time to wait in Open state before transitioning to HalfOpen.
	// Default: 30 seconds
	CooldownDuration time.Duration

	// SuccessThreshold is the number of consecutive successes in HalfOpen state needed to close the circuit.
	// Typically 1 (single successful test call).
	// Default: 1
	SuccessThreshold int
}

// DefaultConfig returns a default circuit breaker configuration.
//
// Default values:
// - FailureThreshold: 5
// - CooldownDuration: 30 seconds
// - SuccessThreshold: 1
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		CooldownDuration: 30 * time.Second,
		SuccessThreshold: 1,
	}
}

// Validate validates the configuration.
// Returns an error if any value is invalid.
func (c Config) Validate() error {
	if c.FailureThreshold <= 0 {
		return errors.New("failure threshold must be greater than 0")
	}
	if c.CooldownDuration <= 0 {
		return errors.New("cooldown duration must be greater than 0")
	}
	if c.SuccessThreshold <= 0 {
		return errors.New("success threshold must be greater than 0")
	}
	return nil
}

// CircuitBreaker implements the circuit breaker pattern for resilience.
//
// The circuit breaker has three states:
// - Closed: Normal operation, all requests pass through
// - Open: Failing fast, all requests are rejected immediately
// - HalfOpen: Testing recovery, allows one test request
//
// State transitions:
// - Closed → Open: When failure count reaches threshold
// - Open → HalfOpen: After cooldown period expires
// - HalfOpen → Closed: When test request succeeds
// - HalfOpen → Open: When test request fails
//
// The circuit breaker is thread-safe and can be used concurrently.
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	config          Config
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	lastStateChange time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration.
//
// The circuit breaker starts in Closed state.
//
// Parameters:
//   - config: Circuit breaker configuration (use DefaultConfig() for defaults)
//
// Returns an error if the configuration is invalid.
func NewCircuitBreaker(config Config) (*CircuitBreaker, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	return &CircuitBreaker{
		state:           StateClosed,
		config:          config,
		failureCount:    0,
		successCount:    0,
		lastFailureTime: time.Time{},
		lastStateChange: now,
	}, nil
}

// State returns the current state of the circuit breaker.
// This method is thread-safe.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Allow checks if a request is allowed based on the current state.
//
// Returns:
//   - true if the request is allowed
//   - false if the circuit is open (request should be rejected)
//
// This method also handles automatic state transitions:
// - Open → HalfOpen when cooldown expires
//
// This method is thread-safe.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Handle automatic state transitions
	cb.updateState()

	switch cb.state {
	case StateClosed:
		// Allow all requests in Closed state
		return true

	case StateOpen:
		// Reject all requests in Open state
		return false

	case StateHalfOpen:
		// Allow one test request in HalfOpen state
		// After this, the state will change based on success/failure
		return true

	default:
		// Unknown state - be safe and reject
		return false
	}
}

// RecordSuccess records a successful call.
//
// This method:
// - Resets failure count in Closed state
// - Increments success count in HalfOpen state
// - Transitions HalfOpen → Closed if threshold reached
//
// This method is thread-safe.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// Reset failure count on success (consecutive failures are what matter)
		cb.failureCount = 0

	case StateHalfOpen:
		// Increment success count
		cb.successCount++

		// Check if we've reached the success threshold
		if cb.successCount >= cb.config.SuccessThreshold {
			// Transition to Closed
			cb.transitionToClosed()
		}
	}
}

// RecordFailure records a failed call.
//
// This method:
// - Increments failure count in Closed state
// - Transitions Closed → Open if threshold reached
// - Transitions HalfOpen → Open immediately
//
// This method is thread-safe.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	cb.lastFailureTime = now

	switch cb.state {
	case StateClosed:
		// Increment failure count
		cb.failureCount++

		// Check if we've reached the failure threshold
		if cb.failureCount >= cb.config.FailureThreshold {
			// Transition to Open
			cb.transitionToOpen(now)
		}

	case StateHalfOpen:
		// Test request failed - immediately transition back to Open
		cb.transitionToOpen(now)
	}
}

// updateState handles automatic state transitions based on time.
// Must be called with lock held.
func (cb *CircuitBreaker) updateState() {
	if cb.state == StateOpen {
		// Check if cooldown period has elapsed
		cooldownExpired := time.Since(cb.lastStateChange) >= cb.config.CooldownDuration
		if cooldownExpired {
			// Transition to HalfOpen
			cb.transitionToHalfOpen()
		}
	}
}

// transitionToOpen transitions the circuit breaker to Open state.
// Must be called with lock held.
func (cb *CircuitBreaker) transitionToOpen(now time.Time) {
	cb.state = StateOpen
	cb.lastStateChange = now
	cb.failureCount = 0 // Reset for next cycle
	cb.successCount = 0
}

// transitionToHalfOpen transitions the circuit breaker to HalfOpen state.
// Must be called with lock held.
func (cb *CircuitBreaker) transitionToHalfOpen() {
	cb.state = StateHalfOpen
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
}

// transitionToClosed transitions the circuit breaker to Closed state.
// Must be called with lock held.
func (cb *CircuitBreaker) transitionToClosed() {
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	cb.failureCount = 0
	cb.successCount = 0
}
