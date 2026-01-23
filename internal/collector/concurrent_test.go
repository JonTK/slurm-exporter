// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestConcurrentCollector(t *testing.T) {
	// Create registry
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 30 * time.Second,
			DefaultTimeout:  10 * time.Second,
			MaxConcurrency:  3,
		},
	}

	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Create concurrent collector
	cc := NewConcurrentCollector(registry, 3)

	t.Run("EmptyRegistry", func(t *testing.T) {
		ctx := context.Background()
		results, err := cc.CollectAll(ctx)
		if err != nil {
			t.Errorf("Expected no error for empty registry, got: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results for empty registry, got %d", len(results))
		}
	})

	t.Run("ConcurrentCollection", func(t *testing.T) {
		// Register multiple collectors
		var collectCount int32

		for i := 0; i < 5; i++ {
			name := string(rune('a' + i))
			collector := &mockCollector{
				name:    name,
				enabled: true,
				collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
					// Simulate some work
					atomic.AddInt32(&collectCount, 1)
					time.Sleep(10 * time.Millisecond)

					// Send a metric
					metric := prometheus.MustNewConstMetric(
						prometheus.NewDesc("test_metric", "Test metric", nil, nil),
						prometheus.GaugeValue,
						float64(atomic.LoadInt32(&collectCount)),
					)
					ch <- metric
					return nil
				},
			}
			_ = registry.Register(name, collector)
		}

		// Collect all
		ctx := context.Background()
		results, err := cc.CollectAll(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify results
		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		// Verify all collectors were called
		if atomic.LoadInt32(&collectCount) != 5 {
			t.Errorf("Expected 5 collections, got %d", atomic.LoadInt32(&collectCount))
		}

		// Verify metrics
		for _, result := range results {
			if !result.Success {
				t.Errorf("Collection for %s failed: %v", result.CollectorName, result.Error)
			}
			if result.MetricCount != 1 {
				t.Errorf("Expected 1 metric for %s, got %d", result.CollectorName, result.MetricCount)
			}
			if result.Duration <= 0 {
				t.Errorf("Expected positive duration for %s", result.CollectorName)
			}
		}
	})

	t.Run("MaxConcurrency", func(t *testing.T) {
		// Clear registry
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Create collector with max concurrency of 2
		cc2 := NewConcurrentCollector(registry, 2)

		var maxConcurrent int32
		var currentConcurrent int32

		// Register collectors that track concurrency
		for i := 0; i < 5; i++ {
			name := string(rune('a' + i))
			collector := &mockCollector{
				name:    name,
				enabled: true,
				collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
					current := atomic.AddInt32(&currentConcurrent, 1)

					// Update max if needed
					for {
						maxVal := atomic.LoadInt32(&maxConcurrent)
						if current <= maxVal || atomic.CompareAndSwapInt32(&maxConcurrent, maxVal, current) {
							break
						}
					}

					// Simulate work
					time.Sleep(20 * time.Millisecond)

					atomic.AddInt32(&currentConcurrent, -1)
					return nil
				},
			}
			_ = registry.Register(name, collector)
		}

		// Collect all
		ctx := context.Background()
		_, err := cc2.CollectAll(ctx)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify max concurrency was respected
		if atomic.LoadInt32(&maxConcurrent) > 2 {
			t.Errorf("Expected max concurrency of 2, but saw %d", atomic.LoadInt32(&maxConcurrent))
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Clear registry
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Register collectors with different behaviors
		_ = registry.Register("success", &mockCollector{
			name:    "success",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return nil
			},
		})

		_ = registry.Register("fail", &mockCollector{
			name:    "fail",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return errors.New("collection failed")
			},
		})

		// Collect all
		ctx := context.Background()
		results, err := cc.CollectAll(ctx)

		// Should have error
		if err == nil {
			t.Error("Expected error from failing collector")
		}

		// Should have results for both
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// Verify individual results
		for _, result := range results {
			if result.CollectorName == "success" && !result.Success {
				t.Error("Success collector should have succeeded")
			}
			if result.CollectorName == "fail" && result.Success {
				t.Error("Fail collector should have failed")
			}
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		// Clear registry
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Register slow collector
		_ = registry.Register("slow", &mockCollector{
			name:    "slow",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				select {
				case <-time.After(1 * time.Second):
					return errors.New("should have been cancelled")
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		})

		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Collect should handle cancellation gracefully
		results, err := cc.CollectAll(ctx)
		if err != nil {
			// Error is expected but should be handled gracefully
			if len(results) == 0 {
				t.Error("Expected at least one result even with cancellation")
			}
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		// Clear registry first
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Wait a bit for any ongoing collections to complete
		time.Sleep(100 * time.Millisecond)

		// Verify metrics tracking
		totalBefore := cc.GetTotalCollections()
		failedBefore := cc.GetFailedCollections()

		// Register and collect
		_ = registry.Register("metric_test", &mockCollector{
			name:    "metric_test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return nil
			},
		})

		ctx := context.Background()
		_, _ = cc.CollectAll(ctx)

		// Check metrics increased
		if cc.GetTotalCollections() <= totalBefore {
			t.Error("Total collections should have increased")
		}

		// Failed should not increase for successful collection
		if cc.GetFailedCollections() != failedBefore {
			t.Error("Failed collections should not have increased")
		}

		// Check active collections (should be 0 after completion)
		if cc.GetActiveCollections() != 0 {
			t.Errorf("Expected 0 active collections, got %d", cc.GetActiveCollections())
		}
	})
}

func TestCollectionOrchestrator(t *testing.T) {
	// Create registry
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 30 * time.Second,
			DefaultTimeout:  10 * time.Second,
			MaxConcurrency:  3,
		},
	}

	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Create orchestrator
	orchestrator := NewCollectionOrchestrator(registry, 3)

	t.Run("SetInterval", func(t *testing.T) {
		orchestrator.SetCollectorInterval("test", 100*time.Millisecond)

		// Verify interval was set
		orchestrator.mu.RLock()
		interval, exists := orchestrator.intervals["test"]
		orchestrator.mu.RUnlock()

		if !exists {
			t.Error("Interval should exist")
		}
		if interval != 100*time.Millisecond {
			t.Errorf("Expected 100ms interval, got %v", interval)
		}
	})

	t.Run("CollectNow", func(t *testing.T) {
		// Register collector
		collected := false
		collector := &mockCollector{
			name:    "immediate",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				collected = true
				return nil
			},
		}
		_ = registry.Register("immediate", collector)

		// Collect immediately
		result, err := orchestrator.CollectNow("immediate")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !result.Success {
			t.Error("Collection should have succeeded")
		}
		if !collected {
			t.Error("Collector should have been called")
		}

		// Try non-existent collector
		_, err = orchestrator.CollectNow("non_existent")
		if err == nil {
			t.Error("Expected error for non-existent collector")
		}

		// Try disabled collector
		collector.SetEnabled(false)
		_, err = orchestrator.CollectNow("immediate")
		if err == nil {
			t.Error("Expected error for disabled collector")
		}
	})

	t.Run("ScheduledCollection", func(t *testing.T) {
		// Register collector with short interval
		var collectCount int32
		collector := &mockCollector{
			name:    "scheduled",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				atomic.AddInt32(&collectCount, 1)
				return nil
			},
		}
		_ = registry.Register("scheduled", collector)

		// Set short interval
		orchestrator.SetCollectorInterval("scheduled", 50*time.Millisecond)

		// Start orchestrator
		orchestrator.Start()

		// Wait for multiple collections
		time.Sleep(200 * time.Millisecond)

		// Verify collections happened before stopping
		count := atomic.LoadInt32(&collectCount)
		if count < 2 {
			t.Errorf("Expected at least 2 collections, got %d", count)
		}

		// Stop orchestrator
		orchestrator.Stop()

		// Wait for any in-flight collection to complete (interval is 50ms, so wait 100ms)
		time.Sleep(100 * time.Millisecond)

		// Record count after in-flight collections complete
		countAfterStop := atomic.LoadInt32(&collectCount)

		// Wait longer than the interval to verify no new collections start
		time.Sleep(150 * time.Millisecond)

		// Verify no NEW collections started after stop
		finalCount := atomic.LoadInt32(&collectCount)
		if finalCount != countAfterStop {
			t.Errorf("New collections started after Stop() (had %d after stop wait, now have %d)", countAfterStop, finalCount)
		}
	})
}

func TestCollectionResult(t *testing.T) {
	result := &CollectionResult{
		CollectorName: "test",
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(100 * time.Millisecond),
		Duration:      100 * time.Millisecond,
		MetricCount:   10,
		Error:         nil,
		Success:       true,
	}

	if result.CollectorName != "test" {
		t.Errorf("Expected collector name 'test', got '%s'", result.CollectorName)
	}
	if result.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", result.Duration)
	}
	if result.MetricCount != 10 {
		t.Errorf("Expected 10 metrics, got %d", result.MetricCount)
	}
	if !result.Success {
		t.Error("Result should be successful")
	}
}
