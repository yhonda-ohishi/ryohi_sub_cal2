package models

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled         bool          `json:"enabled" yaml:"enabled"`
	MaxRequests     uint32        `json:"max_requests" yaml:"max_requests"`
	Interval        time.Duration `json:"interval" yaml:"interval"`
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	FailureRatio    float64       `json:"failure_ratio" yaml:"failure_ratio"`
	MinimumRequests uint32        `json:"minimum_requests" yaml:"minimum_requests"`
}

// Validate validates the circuit breaker configuration
func (c *CircuitBreakerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	
	if c.MaxRequests == 0 {
		c.MaxRequests = 3 // Default max requests in half-open state
	}
	
	if c.Interval == 0 {
		c.Interval = 60 * time.Second // Default interval
	}
	
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second // Default timeout
	}
	
	if c.FailureRatio < 0 || c.FailureRatio > 1 {
		return fmt.Errorf("failure ratio must be between 0 and 1")
	}
	
	if c.FailureRatio == 0 {
		c.FailureRatio = 0.6 // Default failure ratio
	}
	
	if c.MinimumRequests == 0 {
		c.MinimumRequests = 3 // Default minimum requests
	}
	
	return nil
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config *CircuitBreakerConfig
	state  CircuitBreakerState
	mutex  sync.RWMutex
	
	// Counters for closed state
	consecutiveSuccesses uint32
	consecutiveFailures  uint32
	
	// Counters for current interval
	requests uint32
	failures uint32
	
	// Timestamps
	lastFailureTime time.Time
	nextAttemptTime time.Time
	intervalStart   time.Time
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:        config,
		state:         StateClosed,
		intervalStart: time.Now(),
	}
}

// Call executes the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.config.Enabled {
		return fn()
	}
	
	if !cb.CanExecute() {
		return fmt.Errorf("circuit breaker is open")
	}
	
	err := fn()
	cb.RecordResult(err == nil)
	return err
}

// CanExecute checks if a request can be executed
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	now := time.Now()
	
	switch cb.state {
	case StateClosed:
		return true
		
	case StateOpen:
		// Check if timeout has passed
		if now.After(cb.nextAttemptTime) {
			// Transition to half-open
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = StateHalfOpen
			cb.consecutiveSuccesses = 0
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
		
	case StateHalfOpen:
		// Allow limited requests
		return cb.consecutiveSuccesses < cb.config.MaxRequests
		
	default:
		return false
	}
}

// RecordResult records the result of a request
func (cb *CircuitBreaker) RecordResult(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	
	now := time.Now()
	
	// Reset counters if interval has passed
	if now.Sub(cb.intervalStart) > cb.config.Interval {
		cb.requests = 0
		cb.failures = 0
		cb.intervalStart = now
	}
	
	cb.requests++
	if !success {
		cb.failures++
		cb.lastFailureTime = now
	}
	
	switch cb.state {
	case StateClosed:
		if !success {
			cb.consecutiveFailures++
			cb.consecutiveSuccesses = 0
			
			// Check if we should open the circuit
			if cb.requests >= cb.config.MinimumRequests {
				failureRatio := float64(cb.failures) / float64(cb.requests)
				if failureRatio >= cb.config.FailureRatio {
					cb.openCircuit()
				}
			}
		} else {
			cb.consecutiveSuccesses++
			cb.consecutiveFailures = 0
		}
		
	case StateHalfOpen:
		if success {
			cb.consecutiveSuccesses++
			if cb.consecutiveSuccesses >= cb.config.MaxRequests {
				// Enough successes, close the circuit
				cb.closeCircuit()
			}
		} else {
			// Failure in half-open state, open the circuit again
			cb.openCircuit()
		}
		
	case StateOpen:
		// Should not happen as requests are blocked in open state
	}
}

// openCircuit transitions the circuit to open state
func (cb *CircuitBreaker) openCircuit() {
	cb.state = StateOpen
	cb.nextAttemptTime = time.Now().Add(cb.config.Timeout)
	cb.consecutiveSuccesses = 0
}

// closeCircuit transitions the circuit to closed state
func (cb *CircuitBreaker) closeCircuit() {
	cb.state = StateClosed
	cb.consecutiveFailures = 0
	cb.requests = 0
	cb.failures = 0
	cb.intervalStart = time.Now()
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return CircuitBreakerStats{
		State:                string(cb.state),
		Requests:             cb.requests,
		Failures:             cb.failures,
		ConsecutiveSuccesses: cb.consecutiveSuccesses,
		ConsecutiveFailures:  cb.consecutiveFailures,
		LastFailureTime:      cb.lastFailureTime,
		NextAttemptTime:      cb.nextAttemptTime,
	}
}

// CircuitBreakerStats represents circuit breaker statistics
type CircuitBreakerStats struct {
	State                string    `json:"state"`
	Requests             uint32    `json:"requests"`
	Failures             uint32    `json:"failures"`
	ConsecutiveSuccesses uint32    `json:"consecutive_successes"`
	ConsecutiveFailures  uint32    `json:"consecutive_failures"`
	LastFailureTime      time.Time `json:"last_failure_time"`
	NextAttemptTime      time.Time `json:"next_attempt_time"`
}