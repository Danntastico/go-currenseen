package circuitbreaker

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.FailureThreshold != 5 {
		t.Errorf("FailureThreshold = %d, want 5", config.FailureThreshold)
	}

	if config.CooldownDuration != 30*time.Second {
		t.Errorf("CooldownDuration = %v, want 30s", config.CooldownDuration)
	}

	if config.SuccessThreshold != 1 {
		t.Errorf("SuccessThreshold = %d, want 1", config.SuccessThreshold)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				FailureThreshold: 5,
				CooldownDuration: 30 * time.Second,
				SuccessThreshold: 1,
			},
			wantErr: false,
		},
		{
			name: "zero failure threshold",
			config: Config{
				FailureThreshold: 0,
				CooldownDuration: 30 * time.Second,
				SuccessThreshold: 1,
			},
			wantErr: true,
		},
		{
			name: "zero cooldown",
			config: Config{
				FailureThreshold: 5,
				CooldownDuration: 0,
				SuccessThreshold: 1,
			},
			wantErr: true,
		},
		{
			name: "zero success threshold",
			config: Config{
				FailureThreshold: 5,
				CooldownDuration: 30 * time.Second,
				SuccessThreshold: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "Closed"},
		{StateOpen, "Open"},
		{StateHalfOpen, "HalfOpen"},
		{State(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("State.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultConfig()
	cb, err := NewCircuitBreaker(config)

	if err != nil {
		t.Fatalf("NewCircuitBreaker() error = %v, want nil", err)
	}

	if cb == nil {
		t.Fatal("NewCircuitBreaker() returned nil")
	}

	if cb.State() != StateClosed {
		t.Errorf("Initial state = %v, want Closed", cb.State())
	}
}

func TestNewCircuitBreaker_InvalidConfig(t *testing.T) {
	config := Config{
		FailureThreshold: 0, // Invalid
		CooldownDuration: 30 * time.Second,
		SuccessThreshold: 1,
	}

	_, err := NewCircuitBreaker(config)
	if err == nil {
		t.Fatal("NewCircuitBreaker() error = nil, want error")
	}
}

func TestCircuitBreaker_Allow_ClosedState(t *testing.T) {
	config := DefaultConfig()
	cb, _ := NewCircuitBreaker(config)

	// In Closed state, all requests should be allowed
	if !cb.Allow() {
		t.Error("Allow() = false, want true (Closed state should allow requests)")
	}
}

func TestCircuitBreaker_Allow_OpenState(t *testing.T) {
	config := Config{
		FailureThreshold: 2,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Record failures to open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Circuit should now be Open
	if cb.State() != StateOpen {
		t.Fatalf("State = %v, want Open", cb.State())
	}

	// In Open state, requests should be rejected
	if cb.Allow() {
		t.Error("Allow() = true, want false (Open state should reject requests)")
	}
}

func TestCircuitBreaker_RecordFailure_ClosedToOpen(t *testing.T) {
	config := Config{
		FailureThreshold: 3,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Record failures
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Errorf("State after 2 failures = %v, want Closed", cb.State())
	}

	// Third failure should open the circuit
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("State after 3 failures = %v, want Open", cb.State())
	}
}

func TestCircuitBreaker_RecordSuccess_ResetsFailureCount(t *testing.T) {
	config := DefaultConfig()
	cb, _ := NewCircuitBreaker(config)

	// Record some failures
	cb.RecordFailure()
	cb.RecordFailure()

	// Record success - should reset failure count
	cb.RecordSuccess()

	// Record more failures - should not open yet (count was reset)
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	// Should still be Closed (only 3 failures since reset)
	if cb.State() != StateClosed {
		t.Errorf("State = %v, want Closed (failure count should have been reset)", cb.State())
	}
}

func TestCircuitBreaker_OpenToHalfOpen_AfterCooldown(t *testing.T) {
	config := Config{
		FailureThreshold: 2,
		CooldownDuration: 50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Fatalf("State = %v, want Open", cb.State())
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	// Allow() should trigger transition to HalfOpen
	if !cb.Allow() {
		t.Error("Allow() = false, want true (should transition to HalfOpen after cooldown)")
	}

	if cb.State() != StateHalfOpen {
		t.Errorf("State = %v, want HalfOpen", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToClosed_OnSuccess(t *testing.T) {
	config := Config{
		FailureThreshold: 2,
		CooldownDuration: 50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for cooldown and transition to HalfOpen
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // Triggers transition to HalfOpen

	if cb.State() != StateHalfOpen {
		t.Fatalf("State = %v, want HalfOpen", cb.State())
	}

	// Record success - should close the circuit
	cb.RecordSuccess()

	if cb.State() != StateClosed {
		t.Errorf("State = %v, want Closed", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenToOpen_OnFailure(t *testing.T) {
	config := Config{
		FailureThreshold: 2,
		CooldownDuration: 50 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	// Wait for cooldown and transition to HalfOpen
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // Triggers transition to HalfOpen

	if cb.State() != StateHalfOpen {
		t.Fatalf("State = %v, want HalfOpen", cb.State())
	}

	// Record failure - should immediately open again
	cb.RecordFailure()

	if cb.State() != StateOpen {
		t.Errorf("State = %v, want Open", cb.State())
	}
}

func TestCircuitBreaker_ThreadSafety(t *testing.T) {
	config := DefaultConfig()
	cb, _ := NewCircuitBreaker(config)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent calls to Allow()
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = cb.Allow()
		}()
	}

	// Concurrent calls to RecordSuccess()
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			cb.RecordSuccess()
		}()
	}

	// Concurrent calls to RecordFailure()
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			cb.RecordFailure()
		}()
	}

	// Concurrent calls to State()
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = cb.State()
		}()
	}

	wg.Wait()

	// If we get here without race condition, test passes
	// The circuit breaker should still be in a valid state
	state := cb.State()
	if state != StateClosed && state != StateOpen && state != StateHalfOpen {
		t.Errorf("Invalid state after concurrent operations: %v", state)
	}
}

func TestCircuitBreaker_ConsecutiveFailures(t *testing.T) {
	config := Config{
		FailureThreshold: 3,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Record failures one by one
	for i := 0; i < 2; i++ {
		cb.RecordFailure()
		if cb.State() != StateClosed {
			t.Errorf("State after %d failures = %v, want Closed", i+1, cb.State())
		}
	}

	// Third failure should open
	cb.RecordFailure()
	if cb.State() != StateOpen {
		t.Errorf("State after 3 failures = %v, want Open", cb.State())
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	config := Config{
		FailureThreshold: 3,
		CooldownDuration: 100 * time.Millisecond,
		SuccessThreshold: 1,
	}
	cb, _ := NewCircuitBreaker(config)

	// Record 2 failures
	cb.RecordFailure()
	cb.RecordFailure()

	// Record success - should reset failure count
	cb.RecordSuccess()

	// Now record 2 more failures - should still be Closed
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.State() != StateClosed {
		t.Errorf("State = %v, want Closed (success should have reset failure count)", cb.State())
	}
}
