// Package resilience provides circuit breaker and fault tolerance mechanisms
package resilience

import (
	"fmt"
	"sync"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed - circuit is closed, requests go through
	StateClosed State = iota
	// StateOpen - circuit is open, requests fail fast
	StateOpen
	// StateHalfOpen - testing if service recovered
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     uint32
	successCount     uint32
	lastFailureTime  time.Time
	lastStateChange  time.Time
	maxRequests      uint32
	interval         time.Duration
	timeout          time.Duration
	failureThreshold uint32
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(cfg *config.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		maxRequests:      cfg.MaxRequests,
		interval:         cfg.Interval,
		timeout:          cfg.Timeout,
		failureThreshold: cfg.FailureThreshold,
		lastStateChange:  time.Now(),
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	if !cb.canExecute() {
		return errors.NewCircuitBreakerError(
			fmt.Sprintf("Circuit breaker is %s", cb.GetState()),
		)
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// canExecute checks if the request can be executed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if timeout has passed
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited requests in half-open state
		return cb.successCount < cb.maxRequests
	default:
		return false
	}
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
}

// onFailure handles a failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		// Check if we should open the circuit
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(StateOpen)
		}
	case StateHalfOpen:
		// Go back to open state on any failure
		cb.setState(StateOpen)
	}
}

// onSuccess handles a successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success in closed state
		if time.Since(cb.lastFailureTime) > cb.interval {
			cb.failureCount = 0
		}
	case StateHalfOpen:
		cb.successCount++
		// If we've had enough successful requests, close the circuit
		if cb.successCount >= cb.maxRequests {
			cb.setState(StateClosed)
		}
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(newState State) {
	if cb.state == newState {
		return
	}

	cb.state = newState
	cb.lastStateChange = time.Now()

	// Reset counters on state change
	cb.failureCount = 0
	cb.successCount = 0
}

// GetState returns the current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) GetState() string {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state.String()
}

// GetMetrics returns current circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":             cb.state.String(),
		"failure_count":     cb.failureCount,
		"success_count":     cb.successCount,
		"last_state_change": cb.lastStateChange,
		"last_failure_time": cb.lastFailureTime,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
}
