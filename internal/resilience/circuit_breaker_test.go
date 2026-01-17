package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_Basic(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in tests

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 3,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 2,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	// Initially closed
	assert.Equal(t, StateClosed, cb.GetState())

	// Successful calls should work
	err := cb.Call(func() error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailures())
}

func TestCircuitBreaker_FailureThreshold(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 3,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 2,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// First two failures should keep circuit closed
	err := cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailures())

	err = cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 2, cb.GetFailures())

	// Third failure should open the circuit
	err = cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 3, cb.GetFailures())

	// Further calls should be blocked
	err = cb.Call(func() error { return nil })
	assert.Equal(t, ErrCircuitOpen, err)
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_Recovery(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// Trigger circuit opening
	_ = cb.Call(func() error { return testErr })
	_ = cb.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open
	err := cb.Call(func() error { return nil })
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState()) // Success in half-open closes circuit
	assert.Equal(t, 0, cb.GetFailures())
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenRequests: 2,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// Trigger circuit opening
	_ = cb.Call(func() error { return testErr })
	_ = cb.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// Failure in half-open should reopen circuit
	err := cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateOpen, cb.GetState())
}

func TestCircuitBreaker_HalfOpenCapacity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		ResetTimeout:     50 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// Trigger circuit opening
	_ = cb.Call(func() error { return testErr })
	_ = cb.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for reset timeout
	time.Sleep(60 * time.Millisecond)

	// First call transitions to half-open
	err := cb.Call(func() error {
		// While processing, circuit should be half-open
		assert.Equal(t, StateHalfOpen, cb.GetState())
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_CallWithContext(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	ctx := context.Background()

	// Test successful call with context
	err := cb.CallWithContext(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)

	// Test failed call with context
	testErr := errors.New("test error")
	err = cb.CallWithContext(ctx, func(ctx context.Context) error {
		return testErr
	})
	assert.Equal(t, testErr, err)
}

func TestCircuitBreaker_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          false, // Disabled
		FailureThreshold: 1,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// Even with failures, calls should go through when disabled
	err := cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateClosed, cb.GetState()) // State doesn't change when disabled

	err = cb.Call(func() error { return testErr })
	assert.Equal(t, testErr, err)
	assert.Equal(t, StateClosed, cb.GetState())
}

func TestCircuitBreaker_Reset(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 2,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	testErr := errors.New("test error")

	// Trigger circuit opening
	_ = cb.Call(func() error { return testErr })
	_ = cb.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 2, cb.GetFailures())

	// Manual reset
	cb.Reset()
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailures())

	// Should work normally after reset
	err := cb.Call(func() error { return nil })
	assert.NoError(t, err)
}

func TestCircuitBreaker_Status(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 3,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 2,
	}

	cb := NewCircuitBreaker("test", cfg, perfMetrics, logger)

	status := cb.GetStatus()
	assert.Equal(t, "test", status.Name)
	assert.Equal(t, StateClosed, status.State)
	assert.Equal(t, 0, status.Failures)
	assert.Equal(t, 3, status.FailureThreshold)
	assert.Equal(t, 100*time.Millisecond, status.ResetTimeout)
	assert.True(t, status.IsHealthy())
	assert.Equal(t, time.Duration(0), status.TimeUntilRetry())
}

func TestCircuitBreakerManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 3,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 2,
	}

	manager := NewCircuitBreakerManager(cfg, perfMetrics, logger)

	// Create circuit breakers
	cb1 := manager.GetOrCreate("cb1")
	cb2 := manager.GetOrCreate("cb2")

	assert.NotNil(t, cb1)
	assert.NotNil(t, cb2)
	assert.NotEqual(t, cb1, cb2)

	// Get existing circuit breaker
	cb1Again := manager.GetOrCreate("cb1")
	assert.Equal(t, cb1, cb1Again)

	// Test Get method
	retrieved, exists := manager.Get("cb1")
	assert.True(t, exists)
	assert.Equal(t, cb1, retrieved)

	_, exists = manager.Get("nonexistent")
	assert.False(t, exists)

	// Test List
	breakers := manager.List()
	assert.Len(t, breakers, 2)
	assert.Contains(t, breakers, "cb1")
	assert.Contains(t, breakers, "cb2")

	// Test GetStatuses
	statuses := manager.GetStatuses()
	assert.Len(t, statuses, 2)
	assert.Contains(t, statuses, "cb1")
	assert.Contains(t, statuses, "cb2")

	// Test HealthCheck when all are healthy
	err := manager.HealthCheck()
	assert.NoError(t, err)

	// Test ResetAll
	manager.ResetAll()
	assert.Equal(t, StateClosed, cb1.GetState())
	assert.Equal(t, StateClosed, cb2.GetState())
}

func TestCircuitBreakerManager_HealthCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	perfMetrics := metrics.NewPerformanceMetrics("test")

	cfg := config.CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 1,
		ResetTimeout:     100 * time.Millisecond,
		HalfOpenRequests: 1,
	}

	manager := NewCircuitBreakerManager(cfg, perfMetrics, logger)

	cb1 := manager.GetOrCreate("cb1")
	cb2 := manager.GetOrCreate("cb2")

	// Initially healthy
	err := manager.HealthCheck()
	assert.NoError(t, err)

	// Break one circuit
	testErr := errors.New("test error")
	_ = cb1.Call(func() error { return testErr })
	assert.Equal(t, StateOpen, cb1.GetState())

	// Should report unhealthy
	err = manager.HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cb1")
	assert.Contains(t, err.Error(), "open")

	// cb2 should still be healthy
	assert.Equal(t, StateClosed, cb2.GetState())
}

func TestState_String(t *testing.T) {
	assert.Equal(t, "closed", StateClosed.String())
	assert.Equal(t, "open", StateOpen.String())
	assert.Equal(t, "half-open", StateHalfOpen.String())
	assert.Equal(t, "unknown", State(999).String())
}

func TestStatus_TimeUntilRetry(t *testing.T) {
	now := time.Now()

	// Closed state - no wait time
	status := Status{
		State:        StateClosed,
		ResetTimeout: 100 * time.Millisecond,
		LastFailure:  now,
	}
	assert.Equal(t, time.Duration(0), status.TimeUntilRetry())

	// Open state - should calculate wait time
	status = Status{
		State:        StateOpen,
		ResetTimeout: 100 * time.Millisecond,
		LastFailure:  now.Add(-50 * time.Millisecond),
	}
	remaining := status.TimeUntilRetry()
	assert.Greater(t, remaining, 40*time.Millisecond)
	assert.Less(t, remaining, 60*time.Millisecond)

	// Open state - timeout exceeded
	status = Status{
		State:        StateOpen,
		ResetTimeout: 100 * time.Millisecond,
		LastFailure:  now.Add(-200 * time.Millisecond),
	}
	assert.Equal(t, time.Duration(0), status.TimeUntilRetry())
}
