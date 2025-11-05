package resilience

import (
	"errors"
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
)

func TestNewCircuitBreaker(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
	}

	cb := NewCircuitBreaker(cfg)

	if cb == nil {
		t.Fatal("Expected non-nil circuit breaker")
	}

	if cb.state != StateClosed {
		t.Error("Expected initial state to be CLOSED")
	}

	if cb.GetState() != "CLOSED" {
		t.Error("Expected GetState to return CLOSED")
	}
}

func TestCircuitBreaker_SuccessfulExecution(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
	}

	cb := NewCircuitBreaker(cfg)

	// Execute successful function
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != "CLOSED" {
		t.Error("Expected circuit to remain CLOSED after success")
	}
}

func TestCircuitBreaker_FailedExecution(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 2, // Low threshold for testing
	}

	cb := NewCircuitBreaker(cfg)

	testErr := errors.New("test error")

	// First failure
	err := cb.Execute(func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	if cb.GetState() != "CLOSED" {
		t.Error("Expected circuit to remain CLOSED after first failure")
	}

	// Second failure should open the circuit
	err = cb.Execute(func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	if cb.GetState() != "OPEN" {
		t.Error("Expected circuit to be OPEN after threshold failures")
	}
}

func TestCircuitBreaker_OpenCircuit(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          100 * time.Millisecond, // Short timeout for testing
		FailureThreshold: 1,
	}

	cb := NewCircuitBreaker(cfg)

	// Trigger circuit open
	cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != "OPEN" {
		t.Fatal("Expected circuit to be OPEN")
	}

	// Execute should fail fast
	err := cb.Execute(func() error {
		t.Error("Function should not be called when circuit is open")
		return nil
	})

	if err == nil {
		t.Error("Expected error when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpenTransition(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          100 * time.Millisecond, // Short timeout for testing
		FailureThreshold: 1,
	}

	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != "OPEN" {
		t.Fatal("Expected circuit to be OPEN")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Next execution should transition to HALF_OPEN
	executed := false
	cb.Execute(func() error {
		executed = true
		return nil
	})

	if !executed {
		t.Error("Expected function to execute in HALF_OPEN state")
	}

	// After successful execution, should transition to CLOSED
	if cb.GetState() != "CLOSED" {
		t.Error("Expected circuit to be CLOSED after successful HALF_OPEN execution")
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          100 * time.Millisecond,
		FailureThreshold: 1,
	}

	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	cb.Execute(func() error {
		return errors.New("failure")
	})

	// Wait for timeout to enter HALF_OPEN
	time.Sleep(150 * time.Millisecond)

	// Fail in HALF_OPEN state
	cb.Execute(func() error {
		return errors.New("still failing")
	})

	// Should go back to OPEN
	if cb.GetState() != "OPEN" {
		t.Error("Expected circuit to return to OPEN after HALF_OPEN failure")
	}
}

func TestCircuitBreaker_GetMetrics(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
	}

	cb := NewCircuitBreaker(cfg)

	// Execute some operations
	cb.Execute(func() error { return nil })
	cb.Execute(func() error { return errors.New("error") })

	metrics := cb.GetMetrics()

	if metrics["state"] != "CLOSED" {
		t.Error("Expected state in metrics")
	}

	if _, ok := metrics["failure_count"]; !ok {
		t.Error("Expected failure_count in metrics")
	}

	if _, ok := metrics["success_count"]; !ok {
		t.Error("Expected success_count in metrics")
	}

	if _, ok := metrics["last_state_change"]; !ok {
		t.Error("Expected last_state_change in metrics")
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 1,
	}

	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	cb.Execute(func() error {
		return errors.New("failure")
	})

	if cb.GetState() != "OPEN" {
		t.Fatal("Expected circuit to be OPEN")
	}

	// Reset
	cb.Reset()

	if cb.GetState() != "CLOSED" {
		t.Error("Expected circuit to be CLOSED after reset")
	}

	// Should be able to execute again
	executed := false
	cb.Execute(func() error {
		executed = true
		return nil
	})

	if !executed {
		t.Error("Expected function to execute after reset")
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      3,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 100,
	}

	cb := NewCircuitBreaker(cfg)

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cb.Execute(func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// No panics means thread-safe
	if cb.GetState() != "CLOSED" {
		t.Error("Expected circuit to remain CLOSED")
	}
}

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cfg := &config.CircuitBreakerConfig{
		Enabled:          true,
		MaxRequests:      2,
		Interval:         60 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 2,
	}

	cb := NewCircuitBreaker(cfg)

	// Start in CLOSED
	if cb.GetState() != "CLOSED" {
		t.Error("Expected initial state CLOSED")
	}

	// Fail twice to open
	cb.Execute(func() error { return errors.New("error") })
	cb.Execute(func() error { return errors.New("error") })

	if cb.GetState() != "OPEN" {
		t.Error("Expected state OPEN after failures")
	}

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	// Should transition to HALF_OPEN
	cb.Execute(func() error { return nil })

	// After max_requests successes, should close
	cb.Execute(func() error { return nil })

	if cb.GetState() != "CLOSED" {
		t.Error("Expected state CLOSED after successful recovery")
	}
}
