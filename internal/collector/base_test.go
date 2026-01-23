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
	// "github.com/jontk/slurm-exporter/internal/slurm"
)

func TestBaseCollector(t *testing.T) {
	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:        true,
		Interval:       30 * time.Second,
		Timeout:        10 * time.Second,
		MaxConcurrency: 2,
		ErrorHandling: config.ErrorHandlingConfig{
			MaxRetries:    3,
			RetryDelay:    1 * time.Second,
			BackoffFactor: 2.0,
			MaxRetryDelay: 30 * time.Second,
			FailFast:      false,
		},
	}

	opts := &CollectorOptions{
		Namespace: "test",
		Subsystem: "collector",
		Timeout:   30 * time.Second,
	}

	metrics := NewCollectorMetrics("test", "collector")

	// Create a nil client for now due to slurm-client build issues
	var client interface{} = nil

	// Create base collector
	base := NewBaseCollector("test_collector", cfg, opts, client, metrics, nil)

	t.Run("Name", func(t *testing.T) {
		if base.Name() != "test_collector" {
			t.Errorf("Expected name 'test_collector', got '%s'", base.Name())
		}
	})

	t.Run("EnabledState", func(t *testing.T) {
		// Initially enabled
		if !base.IsEnabled() {
			t.Error("Collector should be initially enabled")
		}

		// Disable
		base.SetEnabled(false)
		if base.IsEnabled() {
			t.Error("Collector should be disabled")
		}

		// Re-enable
		base.SetEnabled(true)
		if !base.IsEnabled() {
			t.Error("Collector should be enabled")
		}
	})

	t.Run("State", func(t *testing.T) {
		state := base.GetState()
		if state.Name != "test_collector" {
			t.Errorf("State name mismatch: %s", state.Name)
		}
		if state.TotalCollections != 0 {
			t.Error("Initial total collections should be 0")
		}

		// Update state
		base.UpdateState(func(s *CollectorState) {
			s.TotalCollections = 5
			s.TotalErrors = 1
		})

		state = base.GetState()
		if state.TotalCollections != 5 {
			t.Errorf("Expected 5 total collections, got %d", state.TotalCollections)
		}
		if state.TotalErrors != 1 {
			t.Errorf("Expected 1 total error, got %d", state.TotalErrors)
		}
	})

	t.Run("CollectWithMetrics", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 100)
		ctx := context.Background()

		// Successful collection
		err := base.CollectWithMetrics(ctx, ch, func(ctx context.Context, ch chan<- prometheus.Metric) error {
			// Simulate successful collection
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		state := base.GetState()
		if state.TotalCollections != 6 { // 5 from previous test + 1
			t.Errorf("Expected 6 total collections, got %d", state.TotalCollections)
		}
		if state.ConsecutiveErrors != 0 {
			t.Error("Consecutive errors should be 0 after success")
		}

		// Failed collection
		err = base.CollectWithMetrics(ctx, ch, func(ctx context.Context, ch chan<- prometheus.Metric) error {
			return errors.New("collection failed")
		})

		if err == nil {
			t.Error("Expected error from failed collection")
		}

		state = base.GetState()
		if state.ConsecutiveErrors != 1 {
			t.Errorf("Expected 1 consecutive error, got %d", state.ConsecutiveErrors)
		}
		if state.TotalErrors != 2 { // 1 from previous test + 1
			t.Errorf("Expected 2 total errors, got %d", state.TotalErrors)
		}
	})

	t.Run("RetryLogic", func(t *testing.T) {
		// Test retry conditions
		if !base.ShouldRetry(0) {
			t.Error("Should retry on first attempt")
		}
		if !base.ShouldRetry(2) {
			t.Error("Should retry on third attempt")
		}
		if base.ShouldRetry(3) {
			t.Error("Should not retry after max retries")
		}

		// Test retry delay calculation
		delay := base.GetRetryDelay(0)
		if delay != 1*time.Second {
			t.Errorf("Expected 1s delay for first retry, got %v", delay)
		}

		delay = base.GetRetryDelay(1)
		if delay != 2*time.Second {
			t.Errorf("Expected 2s delay for second retry, got %v", delay)
		}

		delay = base.GetRetryDelay(10)
		if delay != 30*time.Second {
			t.Errorf("Expected max delay of 30s, got %v", delay)
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testErr := errors.New("test error")

		// Test error wrapping
		wrapped := base.WrapError(testErr, "operation failed")
		if wrapped == nil {
			t.Error("Expected wrapped error")
		}
		expectedMsg := "collector test_collector: operation failed: test error"
		if wrapped.Error() != expectedMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedMsg, wrapped.Error())
		}

		// Test nil error
		wrapped = base.WrapError(nil, "operation succeeded")
		if wrapped != nil {
			t.Error("Expected nil for nil input error")
		}
	})

	t.Run("MetricBuilding", func(t *testing.T) {
		desc := prometheus.NewDesc(
			"test_metric",
			"Test metric",
			[]string{"label1"},
			nil,
		)

		metric := base.BuildMetric(desc, prometheus.GaugeValue, 42.0, "value1")
		if metric == nil {
			t.Error("Expected metric to be created")
		}

		// Test metric value by collecting it
		ch := make(chan prometheus.Metric, 1)
		ch <- metric
		close(ch)

		collected := <-ch
		if collected == nil {
			t.Error("Expected to collect the metric")
		}
	})
}

func TestCollectorMetrics(t *testing.T) {
	metrics := NewCollectorMetrics("test", "exporter")
	registry := prometheus.NewRegistry()

	// Register metrics
	err := metrics.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Update some metrics
	labels := prometheus.Labels{"collector": "test"}
	metrics.Total.With(labels).Inc()
	metrics.Errors.With(labels).Inc()
	metrics.Duration.With(labels).Observe(1.5)
	metrics.Up.With(labels).Set(1)

	// Verify metric values
	metricCount, err := testutil.GatherAndCount(
		registry,
		"test_exporter_collections_total",
		"test_exporter_collection_errors_total",
		"test_exporter_collection_duration_seconds",
		"test_exporter_collector_up",
	)
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Should have 4 metrics
	if metricCount != 4 {
		t.Errorf("Expected 4 metrics, got %d", metricCount)
	}
}

// Note: TestCollectionError moved to errors_test.go with enhanced testing
