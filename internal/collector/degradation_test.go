// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("NormalOperation", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)

		// Successful calls should work
		err := cb.Call(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected successful call, got error: %v", err)
		}

		// State should remain closed
		if cb.GetState() != StateClosed {
			t.Errorf("Expected closed state, got %v", cb.GetState())
		}
	})

	t.Run("CircuitOpensAfterFailures", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 3, 1*time.Second)

		// Generate failures
		for i := 0; i < 3; i++ {
			err := cb.Call(func() error {
				return errors.New("test failure")
			})
			if err == nil {
				t.Error("Expected error from failing call")
			}
		}

		// Circuit should be open
		if cb.GetState() != StateOpen {
			t.Errorf("Expected open state after %d failures, got %v", 3, cb.GetState())
		}

		// Calls should be rejected
		err := cb.Call(func() error {
			return nil
		})
		if err == nil {
			t.Error("Expected error when circuit is open")
		}
	})

	t.Run("CircuitTransitionsToHalfOpen", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

		// Open the circuit
		for i := 0; i < 2; i++ {
			_ = cb.Call(func() error {
				return errors.New("test failure")
			})
		}

		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// Next call should go through (half-open)
		called := false
		err := cb.Call(func() error {
			called = true
			return nil
		})

		if !called {
			t.Error("Call should have been attempted in half-open state")
		}
		if err != nil {
			t.Errorf("Expected successful call in half-open state, got: %v", err)
		}

		// Circuit should be closed after success
		if cb.GetState() != StateClosed {
			t.Errorf("Expected closed state after successful test, got %v", cb.GetState())
		}
	})

	t.Run("FailureInHalfOpenReturnsToOpen", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 100*time.Millisecond)

		// Open the circuit
		for i := 0; i < 2; i++ {
			_ = cb.Call(func() error {
				return errors.New("test failure")
			})
		}

		// Wait for reset timeout
		time.Sleep(150 * time.Millisecond)

		// Fail in half-open state
		_ = cb.Call(func() error {
			return errors.New("test failure in half-open")
		})

		// Circuit should be open again
		if cb.GetState() != StateOpen {
			t.Errorf("Expected open state after failure in half-open, got %v", cb.GetState())
		}
	})

	t.Run("ManualReset", func(t *testing.T) {
		cb := NewCircuitBreaker("test", 2, 10*time.Second)

		// Open the circuit
		for i := 0; i < 2; i++ {
			_ = cb.Call(func() error {
				return errors.New("test failure")
			})
		}

		// Manually reset
		cb.Reset()

		// Should be closed
		if cb.GetState() != StateClosed {
			t.Errorf("Expected closed state after reset, got %v", cb.GetState())
		}

		// Calls should work
		err := cb.Call(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected successful call after reset, got: %v", err)
		}
	})
}

func TestDegradationManager(t *testing.T) {
	cfg := &config.DegradationConfig{
		Enabled:          true,
		MaxFailures:      2,
		ResetTimeout:     100 * time.Millisecond,
		UseCachedMetrics: true,
		CacheTTL:         1 * time.Second,
	}

	promRegistry := prometheus.NewRegistry()
	dm, err := NewDegradationManager(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create degradation manager: %v", err)
	}

	t.Run("SuccessfulCollection", func(t *testing.T) {
		ctx := context.Background()
		testMetrics := []prometheus.Metric{
			prometheus.MustNewConstMetric(
				prometheus.NewDesc("test_metric", "Test metric", nil, nil),
				prometheus.GaugeValue,
				42.0,
			),
		}

		metrics, err := dm.ExecuteWithDegradation(ctx, "test_collector", func(context.Context) ([]prometheus.Metric, error) {
			return testMetrics, nil
		})

		if err != nil {
			t.Errorf("Expected successful collection, got error: %v", err)
		}
		if len(metrics) != 1 {
			t.Errorf("Expected 1 metric, got %d", len(metrics))
		}

		// Check circuit breaker state metric
		metricVal := testutil.ToFloat64(dm.degradationMetrics.CircuitBreakerState.WithLabelValues("test_collector"))
		if metricVal != float64(StateClosed) {
			t.Errorf("Expected circuit breaker state %d, got %f", StateClosed, metricVal)
		}
	})

	t.Run("CachedMetricsAfterFailure", func(t *testing.T) {
		ctx := context.Background()

		// First successful collection to cache metrics
		testMetrics := []prometheus.Metric{
			prometheus.MustNewConstMetric(
				prometheus.NewDesc("cached_metric", "Cached metric", nil, nil),
				prometheus.GaugeValue,
				100.0,
			),
		}

		_, _ = dm.ExecuteWithDegradation(ctx, "cache_test", func(context.Context) ([]prometheus.Metric, error) {
			return testMetrics, nil
		})

		// Generate failures to open circuit
		for i := 0; i < cfg.MaxFailures; i++ {
			_, _ = dm.ExecuteWithDegradation(ctx, "cache_test", func(context.Context) ([]prometheus.Metric, error) {
				return nil, errors.New("test failure")
			})
		}

		// Should get cached metrics
		metrics, err := dm.ExecuteWithDegradation(ctx, "cache_test", func(context.Context) ([]prometheus.Metric, error) {
			t.Error("Should not call collect function when circuit is open")
			return nil, errors.New("should not be called")
		})

		if err != nil {
			t.Errorf("Expected cached metrics, got error: %v", err)
		}
		if len(metrics) != 1 {
			t.Errorf("Expected 1 cached metric, got %d", len(metrics))
		}

		// Check cached metrics served counter - should be at least 1
		cachedServed := testutil.ToFloat64(dm.degradationMetrics.CachedMetricsServed.WithLabelValues("cache_test"))
		if cachedServed < 1 {
			t.Errorf("Expected at least 1 cached metric served, got %f", cachedServed)
		}
	})

	t.Run("CacheExpiry", func(t *testing.T) {
		ctx := context.Background()

		// Create short-lived cache
		shortCfg := &config.DegradationConfig{
			Enabled:          true,
			MaxFailures:      2,
			ResetTimeout:     100 * time.Millisecond,
			UseCachedMetrics: true,
			CacheTTL:         100 * time.Millisecond,
		}

		promRegistry2 := prometheus.NewRegistry()
		dm2, _ := NewDegradationManager(shortCfg, promRegistry2)

		// Cache metrics
		testMetrics := []prometheus.Metric{
			prometheus.MustNewConstMetric(
				prometheus.NewDesc("expiring_metric", "Expiring metric", nil, nil),
				prometheus.GaugeValue,
				200.0,
			),
		}

		_, _ = dm2.ExecuteWithDegradation(ctx, "expiry_test", func(context.Context) ([]prometheus.Metric, error) {
			return testMetrics, nil
		})

		// Open circuit
		for i := 0; i < shortCfg.MaxFailures; i++ {
			_, _ = dm2.ExecuteWithDegradation(ctx, "expiry_test", func(context.Context) ([]prometheus.Metric, error) {
				return nil, errors.New("test failure")
			})
		}

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		// Should not get cached metrics
		metrics, err := dm2.ExecuteWithDegradation(ctx, "expiry_test", func(context.Context) ([]prometheus.Metric, error) {
			return nil, errors.New("no metrics available")
		})

		if err == nil {
			t.Error("Expected error when cache expired and circuit open")
		}
		if len(metrics) > 0 {
			t.Errorf("Expected no metrics after cache expiry, got %d", len(metrics))
		}
	})

	t.Run("DegradationModeTracking", func(t *testing.T) {
		// Update degradation mode
		dm.UpdateDegradationMode()

		// Check degradation mode metric
		modeVal := testutil.ToFloat64(dm.degradationMetrics.DegradationMode)
		// Could be Normal or Partial depending on previous tests
		if modeVal < 0 || modeVal > 2 {
			t.Errorf("Invalid degradation mode value: %f", modeVal)
		}
	})

	t.Run("GetDegradationStats", func(t *testing.T) {
		stats := dm.GetDegradationStats()

		// Should have stats for collectors we've used
		if _, exists := stats["test_collector"]; !exists {
			t.Error("Expected stats for test_collector")
		}
		if _, exists := stats["cache_test"]; !exists {
			t.Error("Expected stats for cache_test")
		}

		// Check stats structure
		for name, stat := range stats {
			if stat.CollectorName != name {
				t.Errorf("Mismatched collector name: expected %s, got %s", name, stat.CollectorName)
			}
			// State should be valid
			if stat.State < StateClosed || stat.State > StateHalfOpen {
				t.Errorf("Invalid state for %s: %v", name, stat.State)
			}
		}
	})

	t.Run("ResetAllBreakers", func(t *testing.T) {
		// Reset all breakers
		dm.ResetAllBreakers()

		// All breakers should be closed
		stats := dm.GetDegradationStats()
		for name, stat := range stats {
			if stat.State != StateClosed {
				t.Errorf("Expected closed state for %s after reset, got %v", name, stat.State)
			}
			if stat.Failures != 0 {
				t.Errorf("Expected 0 failures for %s after reset, got %d", name, stat.Failures)
			}
		}

		// Mode should be normal
		dm.UpdateDegradationMode()
		modeVal := testutil.ToFloat64(dm.degradationMetrics.DegradationMode)
		if modeVal != float64(ModeNormal) {
			t.Errorf("Expected normal mode after reset, got %f", modeVal)
		}
	})
}

func TestDegradationMetrics(t *testing.T) {
	metrics := NewDegradationMetrics("test", "degradation")
	registry := prometheus.NewRegistry()

	err := metrics.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Update some metrics
	metrics.CircuitBreakerState.WithLabelValues("test").Set(float64(StateOpen))
	metrics.DegradationMode.Set(float64(ModePartial))
	metrics.CachedMetricsServed.WithLabelValues("test").Inc()
	metrics.FailuresBeforeDegradation.WithLabelValues("test").Observe(3)

	// Verify metrics were registered
	metricCount, err := testutil.GatherAndCount(registry)
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
	if metricCount < 4 {
		t.Errorf("Expected at least 4 metrics, got %d", metricCount)
	}
}

func TestDegradationConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		cfg := &config.DegradationConfig{
			Enabled:          true,
			MaxFailures:      3,
			ResetTimeout:     5 * time.Minute,
			UseCachedMetrics: true,
			CacheTTL:         10 * time.Minute,
		}

		if err := cfg.Validate(); err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("InvalidMaxFailures", func(t *testing.T) {
		cfg := &config.DegradationConfig{
			Enabled:      true,
			MaxFailures:  0,
			ResetTimeout: 5 * time.Minute,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid max failures")
		}
	})

	t.Run("InvalidResetTimeout", func(t *testing.T) {
		cfg := &config.DegradationConfig{
			Enabled:      true,
			MaxFailures:  3,
			ResetTimeout: 0,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid reset timeout")
		}
	})

	t.Run("InvalidCacheTTL", func(t *testing.T) {
		cfg := &config.DegradationConfig{
			Enabled:          true,
			MaxFailures:      3,
			ResetTimeout:     5 * time.Minute,
			UseCachedMetrics: true,
			CacheTTL:         0,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("Expected error for invalid cache TTL")
		}
	})

	t.Run("DisabledConfig", func(t *testing.T) {
		cfg := &config.DegradationConfig{
			Enabled: false,
			// Other fields can be invalid when disabled
			MaxFailures:  0,
			ResetTimeout: 0,
		}

		if err := cfg.Validate(); err != nil {
			t.Errorf("Disabled config should be valid, got error: %v", err)
		}
	})
}
