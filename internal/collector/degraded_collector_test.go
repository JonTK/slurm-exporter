// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestDegradedCollector(t *testing.T) {
	// Create degradation config
	degradationConfig := &config.DegradationConfig{
		Enabled:          true,
		MaxFailures:      2,
		ResetTimeout:     100 * time.Millisecond,
		UseCachedMetrics: true,
		CacheTTL:         1 * time.Second,
	}

	promRegistry := prometheus.NewRegistry()
	degradationManager, err := NewDegradationManager(degradationConfig, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create degradation manager: %v", err)
	}

	t.Run("SuccessfulCollection", func(t *testing.T) {
		// Create mock collector
		mockCollector := &mockCollector{
			name:    "test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				metric := prometheus.MustNewConstMetric(
					prometheus.NewDesc("test_metric", "Test metric", nil, nil),
					prometheus.GaugeValue,
					42.0,
				)
				ch <- metric
				return nil
			},
		}

		// Create degraded collector
		degradedCollector := NewDegradedCollector(mockCollector, degradationManager)

		// Test collection
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := degradedCollector.Collect(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Expected successful collection, got error: %v", err)
		}

		// Count metrics
		metricCount := 0
		for range metricChan {
			metricCount++
		}

		if metricCount != 1 {
			t.Errorf("Expected 1 metric, got %d", metricCount)
		}
	})

	t.Run("FailureWithCaching", func(t *testing.T) {
		// Create mock collector that caches then fails
		var callCount int
		mockCollector := &mockCollector{
			name:    "failing_test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				callCount++
				if callCount == 1 {
					// First call succeeds and caches
					metric := prometheus.MustNewConstMetric(
						prometheus.NewDesc("cached_metric", "Cached metric", nil, nil),
						prometheus.GaugeValue,
						100.0,
					)
					ch <- metric
					return nil
				}
				// Subsequent calls fail
				return errors.New("collection failed")
			},
		}

		degradedCollector := NewDegradedCollector(mockCollector, degradationManager)
		ctx := context.Background()

		// First collection - should succeed and cache
		metricChan1 := make(chan prometheus.Metric, 10)
		err := degradedCollector.Collect(ctx, metricChan1)
		close(metricChan1)

		if err != nil {
			t.Errorf("First collection should succeed, got error: %v", err)
		}

		// Generate failures to open circuit
		for i := 0; i < degradationConfig.MaxFailures; i++ {
			metricChan := make(chan prometheus.Metric, 10)
			_ = degradedCollector.Collect(ctx, metricChan)
			close(metricChan)
		}

		// Should now get cached metrics
		metricChan2 := make(chan prometheus.Metric, 10)
		err = degradedCollector.Collect(ctx, metricChan2)
		close(metricChan2)

		if err != nil {
			t.Errorf("Expected cached metrics, got error: %v", err)
		}

		// Should have received cached metric
		metricCount := 0
		for range metricChan2 {
			metricCount++
		}

		if metricCount != 1 {
			t.Errorf("Expected 1 cached metric, got %d", metricCount)
		}
	})

	t.Run("InterfaceMethods", func(t *testing.T) {
		mockCollector := &mockCollector{
			name:    "interface_test",
			enabled: true,
		}

		degradedCollector := NewDegradedCollector(mockCollector, degradationManager)

		// Test Name
		if degradedCollector.Name() != "interface_test" {
			t.Errorf("Expected name 'interface_test', got '%s'", degradedCollector.Name())
		}

		// Test IsEnabled
		if !degradedCollector.IsEnabled() {
			t.Error("Expected collector to be enabled")
		}

		// Test SetEnabled
		degradedCollector.SetEnabled(false)
		if degradedCollector.IsEnabled() {
			t.Error("Expected collector to be disabled")
		}

		// Test Describe
		descChan := make(chan *prometheus.Desc, 10)
		degradedCollector.Describe(descChan)
		close(descChan)

		// mockCollector doesn't provide descriptions, so should be empty
		descCount := 0
		for range descChan {
			descCount++
		}

		if descCount != 0 {
			t.Errorf("Expected 0 descriptions from mock collector, got %d", descCount)
		}
	})
}

func TestDegradedRegistry(t *testing.T) {
	// Create test configuration
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval:     30 * time.Second,
			DefaultTimeout:      10 * time.Second,
			MaxConcurrency:      3,
			ErrorThreshold:      5,
			RecoveryDelay:       60 * time.Second,
			GracefulDegradation: true,
		},
		Cluster: config.CollectorConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  10 * time.Second,
		},
		Degradation: config.DegradationConfig{
			Enabled:          true,
			MaxFailures:      3,
			ResetTimeout:     5 * time.Minute,
			UseCachedMetrics: true,
			CacheTTL:         10 * time.Minute,
		},
	}

	promRegistry := prometheus.NewRegistry()
	registry, err := NewDegradedRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create degraded registry: %v", err)
	}

	t.Run("RegisterAndGet", func(t *testing.T) {
		mockCollector := &mockCollector{
			name:    "test_collector",
			enabled: true,
		}

		// Register collector
		err := registry.Register("test_collector", mockCollector)
		if err != nil {
			t.Errorf("Failed to register collector: %v", err)
		}

		// Get collector - should return degraded wrapper
		collector, exists := registry.Get("test_collector")
		if !exists {
			t.Error("Expected collector to exist")
		}

		if collector.Name() != "test_collector" {
			t.Errorf("Expected collector name 'test_collector', got '%s'", collector.Name())
		}

		// Should be wrapped in DegradedCollector
		if _, ok := collector.(*DegradedCollector); !ok {
			t.Error("Expected collector to be wrapped in DegradedCollector")
		}
	})

	t.Run("List", func(t *testing.T) {
		collectors := registry.List()
		found := false
		for _, name := range collectors {
			if name == "test_collector" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'test_collector' in list")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		err := registry.Unregister("test_collector")
		if err != nil {
			t.Errorf("Failed to unregister collector: %v", err)
		}

		_, exists := registry.Get("test_collector")
		if exists {
			t.Error("Expected collector to be unregistered")
		}
	})

	t.Run("DegradationManagement", func(t *testing.T) {
		// Test degradation manager access
		dm := registry.GetDegradationManager()
		if dm == nil {
			t.Error("Expected degradation manager to be available")
		}

		// Test degradation mode update
		registry.UpdateDegradationMode()

		// Test stats
		stats := registry.GetDegradationStats()
		if stats == nil {
			t.Error("Expected degradation stats to be available")
		}

		// Test reset
		registry.ResetAllBreakers()
	})

	t.Run("PrometheusCollector", func(t *testing.T) {
		collector := registry.PrometheusCollector()
		if collector == nil {
			t.Error("Expected Prometheus collector to be available")
		}
	})

	t.Run("CollectAll", func(t *testing.T) {
		// Register a test collector
		mockCollector := &mockCollector{
			name:    "collect_all_test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				metric := prometheus.MustNewConstMetric(
					prometheus.NewDesc("collect_all_metric", "Collect all metric", nil, nil),
					prometheus.GaugeValue,
					123.0,
				)
				ch <- metric
				return nil
			},
		}

		_ = registry.Register("collect_all_test", mockCollector)

		// Test collect all
		ctx := context.Background()
		results, err := registry.CollectAll(ctx)

		if err != nil {
			t.Errorf("CollectAll failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one collection result")
		}

		// Find our test collector result
		found := false
		for _, result := range results {
			if result.CollectorName == "collect_all_test" {
				found = true
				if !result.Success {
					t.Errorf("Expected successful collection, got error: %v", result.Error)
				}
				if result.MetricCount != 1 {
					t.Errorf("Expected 1 metric, got %d", result.MetricCount)
				}
				break
			}
		}

		if !found {
			t.Error("Expected to find collect_all_test in results")
		}
	})
}

func TestDegradedRegistryWithDisabledDegradation(t *testing.T) {
	// Create config with degradation disabled
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval:     30 * time.Second,
			DefaultTimeout:      10 * time.Second,
			MaxConcurrency:      3,
			ErrorThreshold:      5,
			RecoveryDelay:       60 * time.Second,
			GracefulDegradation: false,
		},
		Degradation: config.DegradationConfig{
			Enabled: false,
		},
	}

	promRegistry := prometheus.NewRegistry()
	registry, err := NewDegradedRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create degraded registry: %v", err)
	}

	// Register collector
	mockCollector := &mockCollector{
		name:    "no_degradation_test",
		enabled: true,
	}

	err = registry.Register("no_degradation_test", mockCollector)
	if err != nil {
		t.Errorf("Failed to register collector: %v", err)
	}

	// Get collector - should NOT be wrapped when degradation is disabled
	collector, exists := registry.Get("no_degradation_test")
	if !exists {
		t.Error("Expected collector to exist")
	}

	// Should NOT be wrapped in DegradedCollector when degradation is disabled
	if _, ok := collector.(*DegradedCollector); ok {
		t.Error("Expected collector NOT to be wrapped when degradation is disabled")
	}
}
