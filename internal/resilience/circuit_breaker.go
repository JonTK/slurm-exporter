// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
	"github.com/sirupsen/logrus"
)

var (
	// ErrCircuitOpen is returned when the circuit breaker is open
	ErrCircuitOpen = errors.New("circuit breaker is open")

	// ErrCircuitHalfOpen is returned when too many requests are made in half-open state
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open and at capacity")
)

// State represents the circuit breaker state
type State int

const (
	// StateClosed means requests are allowed through
	StateClosed State = iota

	// StateOpen means requests are blocked
	StateOpen

	// StateHalfOpen means limited requests are allowed for testing
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for protecting SLURM API calls
type CircuitBreaker struct {
	name             string
	config           config.CircuitBreakerConfig
	metrics          *metrics.PerformanceMetrics
	logger           *logrus.Logger
	state            State
	failures         int
	halfOpenRequests int
	lastFailure      time.Time
	lastStateChange  time.Time
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, cfg config.CircuitBreakerConfig, perfMetrics *metrics.PerformanceMetrics, logger *logrus.Logger) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:            name,
		config:          cfg,
		metrics:         perfMetrics,
		logger:          logger,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}

	// Update initial state metric
	if cb.metrics != nil {
		cb.metrics.UpdateCircuitBreakerState(name, float64(StateClosed))
	}

	return cb
}

// Call executes the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.config.Enabled {
		return fn()
	}

	// Check if we can make the call
	if err := cb.allowRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.recordResult(err)

	return err
}

// CallWithContext executes the given function with context if the circuit breaker allows it
func (cb *CircuitBreaker) CallWithContext(ctx context.Context, fn func(context.Context) error) error {
	if !cb.config.Enabled {
		return fn(ctx)
	}

	// Check if we can make the call
	if err := cb.allowRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn(ctx)

	// Record the result
	cb.recordResult(err)

	return err
}

// allowRequest checks if a request is allowed based on current state
func (cb *CircuitBreaker) allowRequest() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
			cb.setState(StateHalfOpen)
			cb.halfOpenRequests = 0
			return nil
		}
		return ErrCircuitOpen

	case StateHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests >= cb.config.HalfOpenRequests {
			return ErrCircuitHalfOpen
		}
		cb.halfOpenRequests++
		return nil

	default:
		return ErrCircuitOpen
	}
}

// recordResult records the result of a function call and updates state accordingly
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		// Log the failure
		cb.logger.WithFields(logrus.Fields{
			"circuit_breaker": cb.name,
			"state":           cb.state.String(),
			"failures":        cb.failures,
			"error":           err,
		}).Warn("Circuit breaker recorded failure")

		// Check if we should open the circuit
		if cb.state == StateClosed && cb.failures >= cb.config.FailureThreshold {
			cb.setState(StateOpen)
		} else if cb.state == StateHalfOpen {
			// Any failure in half-open state reopens the circuit
			cb.setState(StateOpen)
		}
	} else {
		// Success
		switch cb.state {
		case StateHalfOpen:
			// Successful request in half-open state closes the circuit
			cb.setState(StateClosed)
			cb.failures = 0
		case StateClosed:
			// Reset failure count on success in closed state
			cb.failures = 0
		}
	}
}

// setState changes the circuit breaker state and updates metrics
func (cb *CircuitBreaker) setState(newState State) {
	oldState := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()

	// Update metrics
	if cb.metrics != nil {
		cb.metrics.UpdateCircuitBreakerState(cb.name, float64(newState))
	}

	// Log state change
	cb.logger.WithFields(logrus.Fields{
		"circuit_breaker": cb.name,
		"old_state":       oldState.String(),
		"new_state":       newState.String(),
		"failures":        cb.failures,
	}).Info("Circuit breaker state changed")
}

// GetState returns the current state of the circuit breaker (thread-safe)
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetFailures returns the current number of failures (thread-safe)
func (cb *CircuitBreaker) GetFailures() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.failures
}

// GetStatus returns detailed status information
func (cb *CircuitBreaker) GetStatus() Status {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return Status{
		Name:             cb.name,
		State:            cb.state,
		Failures:         cb.failures,
		HalfOpenRequests: cb.halfOpenRequests,
		LastFailure:      cb.lastFailure,
		LastStateChange:  cb.lastStateChange,
		FailureThreshold: cb.config.FailureThreshold,
		ResetTimeout:     cb.config.ResetTimeout,
		HalfOpenCapacity: cb.config.HalfOpenRequests,
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.setState(StateClosed)
	cb.failures = 0
	cb.halfOpenRequests = 0

	cb.logger.WithField("circuit_breaker", cb.name).Info("Circuit breaker manually reset")
}

// Status represents the current status of a circuit breaker
type Status struct {
	Name             string        `json:"name"`
	State            State         `json:"state"`
	Failures         int           `json:"failures"`
	HalfOpenRequests int           `json:"half_open_requests"`
	LastFailure      time.Time     `json:"last_failure"`
	LastStateChange  time.Time     `json:"last_state_change"`
	FailureThreshold int           `json:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
	HalfOpenCapacity int           `json:"half_open_capacity"`
}

// IsHealthy returns true if the circuit breaker is in a healthy state
func (s Status) IsHealthy() bool {
	return s.State == StateClosed || s.State == StateHalfOpen
}

// TimeUntilRetry returns the time until the circuit breaker will allow requests again
func (s Status) TimeUntilRetry() time.Duration {
	if s.State != StateOpen {
		return 0
	}

	elapsed := time.Since(s.LastFailure)
	if elapsed >= s.ResetTimeout {
		return 0
	}

	return s.ResetTimeout - elapsed
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	config   config.CircuitBreakerConfig
	metrics  *metrics.PerformanceMetrics
	logger   *logrus.Logger
	mu       sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager(cfg config.CircuitBreakerConfig, perfMetrics *metrics.PerformanceMetrics, logger *logrus.Logger) *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
		config:   cfg,
		metrics:  perfMetrics,
		logger:   logger,
	}
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string) *CircuitBreaker {
	cbm.mu.RLock()
	if cb, exists := cbm.breakers[name]; exists {
		cbm.mu.RUnlock()
		return cb
	}
	cbm.mu.RUnlock()

	cbm.mu.Lock()
	defer cbm.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	// Create new circuit breaker
	cb := NewCircuitBreaker(name, cbm.config, cbm.metrics, cbm.logger)
	cbm.breakers[name] = cb

	cbm.logger.WithField("circuit_breaker", name).Info("Created new circuit breaker")

	return cb
}

// Get returns an existing circuit breaker
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	cb, exists := cbm.breakers[name]
	return cb, exists
}

// List returns all circuit breakers
func (cbm *CircuitBreakerManager) List() map[string]*CircuitBreaker {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, cb := range cbm.breakers {
		result[name] = cb
	}
	return result
}

// GetStatuses returns the status of all circuit breakers
func (cbm *CircuitBreakerManager) GetStatuses() map[string]Status {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()

	statuses := make(map[string]Status)
	for name, cb := range cbm.breakers {
		statuses[name] = cb.GetStatus()
	}
	return statuses
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mu.RLock()
	breakers := make([]*CircuitBreaker, 0, len(cbm.breakers))
	for _, cb := range cbm.breakers {
		breakers = append(breakers, cb)
	}
	cbm.mu.RUnlock()

	for _, cb := range breakers {
		cb.Reset()
	}

	cbm.logger.Info("Reset all circuit breakers")
}

// HealthCheck returns the overall health status of all circuit breakers
func (cbm *CircuitBreakerManager) HealthCheck() error {
	statuses := cbm.GetStatuses()

	var unhealthy []string
	for name, status := range statuses {
		if !status.IsHealthy() {
			unhealthy = append(unhealthy, fmt.Sprintf("%s (%s)", name, status.State.String()))
		}
	}

	if len(unhealthy) > 0 {
		return fmt.Errorf("unhealthy circuit breakers: %v", unhealthy)
	}

	return nil
}
