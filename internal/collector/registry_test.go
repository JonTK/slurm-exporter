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

// mockRegistryCollector implements the Collector interface for testing
type mockRegistryCollector struct {
	name        string
	enabled     bool
	collectFunc func(context.Context, chan<- prometheus.Metric) error
	descs       []*prometheus.Desc
}

func (m *mockRegistryCollector) Name() string {
	return m.name
}

func (m *mockRegistryCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range m.descs {
		ch <- desc
	}
}

func (m *mockRegistryCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if m.collectFunc != nil {
		return m.collectFunc(ctx, ch)
	}
	return nil
}

func (m *mockRegistryCollector) IsEnabled() bool {
	return m.enabled
}

func (m *mockRegistryCollector) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func TestRegistry(t *testing.T) {
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 30 * time.Second,
			DefaultTimeout:  10 * time.Second,
			MaxConcurrency:  5,
		},
	}

	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	t.Run("RegisterAndGet", func(t *testing.T) {
		collector := &mockRegistryCollector{
			name:    "test_collector",
			enabled: true,
		}

		// Register collector
		err := registry.Register("test", collector)
		if err != nil {
			t.Errorf("Failed to register collector: %v", err)
		}

		// Get collector
		retrieved, exists := registry.Get("test")
		if !exists {
			t.Error("Collector should exist")
		}
		if retrieved.Name() != "test_collector" {
			t.Errorf("Expected collector name 'test_collector', got '%s'", retrieved.Name())
		}

		// Try to register duplicate
		err = registry.Register("test", collector)
		if err == nil {
			t.Error("Expected error when registering duplicate collector")
		}

		// Get non-existent collector
		_, exists = registry.Get("non_existent")
		if exists {
			t.Error("Non-existent collector should not exist")
		}
	})

	t.Run("List", func(t *testing.T) {
		// Clear registry first
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Register multiple collectors
		for i := 0; i < 3; i++ {
			collector := &mockRegistryCollector{
				name:    string(rune('a' + i)),
				enabled: true,
			}
			err := registry.Register(string(rune('a'+i)), collector)
			if err != nil {
				t.Errorf("Failed to register collector %d: %v", i, err)
			}
		}

		// List collectors
		names := registry.List()
		if len(names) != 3 {
			t.Errorf("Expected 3 collectors, got %d", len(names))
		}
	})

	t.Run("EnableDisable", func(t *testing.T) {
		collector := &mockRegistryCollector{
			name:    "toggle_collector",
			enabled: true,
		}

		err := registry.Register("toggle", collector)
		if err != nil {
			t.Fatalf("Failed to register collector: %v", err)
		}

		// Disable collector
		err = registry.DisableCollector("toggle")
		if err != nil {
			t.Errorf("Failed to disable collector: %v", err)
		}
		if collector.IsEnabled() {
			t.Error("Collector should be disabled")
		}

		// Enable collector
		err = registry.EnableCollector("toggle")
		if err != nil {
			t.Errorf("Failed to enable collector: %v", err)
		}
		if !collector.IsEnabled() {
			t.Error("Collector should be enabled")
		}

		// Try to enable non-existent collector
		err = registry.EnableCollector("non_existent")
		if err == nil {
			t.Error("Expected error when enabling non-existent collector")
		}
	})

	t.Run("Unregister", func(t *testing.T) {
		collector := &mockRegistryCollector{
			name:    "temp_collector",
			enabled: true,
		}

		// Register and then unregister
		err := registry.Register("temp", collector)
		if err != nil {
			t.Fatalf("Failed to register collector: %v", err)
		}

		err = registry.Unregister("temp")
		if err != nil {
			t.Errorf("Failed to unregister collector: %v", err)
		}

		// Verify it's gone
		_, exists := registry.Get("temp")
		if exists {
			t.Error("Collector should not exist after unregistering")
		}

		// Try to unregister non-existent collector
		err = registry.Unregister("non_existent")
		if err == nil {
			t.Error("Expected error when unregistering non-existent collector")
		}
	})

	t.Run("CollectAll", func(t *testing.T) {
		// Clear registry
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Register collectors with different behaviors
		successCollector := &mockRegistryCollector{
			name:    "success",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				// Simulate successful collection
				return nil
			},
		}

		failCollector := &mockRegistryCollector{
			name:    "fail",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return errors.New("collection failed")
			},
		}

		disabledCollector := &mockRegistryCollector{
			name:    "disabled",
			enabled: false,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				t.Error("Disabled collector should not be called")
				return nil
			},
		}

		_ = registry.Register("success", successCollector)
		_ = registry.Register("fail", failCollector)
		_ = registry.Register("disabled", disabledCollector)

		// Collect all
		ctx := context.Background()
		err := registry.CollectAll(ctx)
		if err == nil {
			t.Error("Expected error from failing collector")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		// Clear registry
		for _, name := range registry.List() {
			_ = registry.Unregister(name)
		}

		// Register a mix of collectors
		mockCol := &mockRegistryCollector{
			name:    "mock",
			enabled: true,
		}

		baseCol := NewBaseCollector(
			"base",
			&config.CollectorConfig{Enabled: true},
			&CollectorOptions{},
			nil,
			NewCollectorMetrics("test", "collector"),
			nil,
		)

		_ = registry.Register("mock", mockCol)
		_ = registry.Register("base", baseCol)

		// Get stats
		stats := registry.GetStats()
		if len(stats) != 2 {
			t.Errorf("Expected 2 stats entries, got %d", len(stats))
		}

		// Check mock collector stats
		if mockStats, ok := stats["mock"]; ok {
			if mockStats.Name != "mock" {
				t.Errorf("Expected name 'mock', got '%s'", mockStats.Name)
			}
			if !mockStats.Enabled {
				t.Error("Mock collector should be enabled")
			}
		} else {
			t.Error("Mock collector stats not found")
		}

		// Check base collector stats
		if baseStats, ok := stats["base"]; ok {
			if baseStats.Name != "base" {
				t.Errorf("Expected name 'base', got '%s'", baseStats.Name)
			}
			if !baseStats.Enabled {
				t.Error("Base collector should be enabled")
			}
		} else {
			t.Error("Base collector stats not found")
		}
	})
}

func TestCollectorAdapter(t *testing.T) {
	// Create a mock collector
	mockCol := &mockRegistryCollector{
		name:    "adapter_test",
		enabled: true,
		descs: []*prometheus.Desc{
			prometheus.NewDesc("test_metric", "Test metric", nil, nil),
		},
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

	adapter := &collectorAdapter{collector: mockCol}

	t.Run("Describe", func(t *testing.T) {
		ch := make(chan *prometheus.Desc, 10)
		adapter.Describe(ch)
		close(ch)

		// Should receive one descriptor
		count := 0
		for range ch {
			count++
		}
		if count != 1 {
			t.Errorf("Expected 1 descriptor, got %d", count)
		}
	})

	t.Run("Collect", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 10)
		adapter.Collect(ch)
		close(ch)

		// Should receive one metric
		count := 0
		for range ch {
			count++
		}
		if count != 1 {
			t.Errorf("Expected 1 metric, got %d", count)
		}
	})

	t.Run("CollectWithError", func(t *testing.T) {
		errorCol := &mockRegistryCollector{
			name:    "error_test",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return errors.New("collection error")
			},
		}

		errorAdapter := &collectorAdapter{collector: errorCol}
		ch := make(chan prometheus.Metric, 10)

		// Should not panic on error
		errorAdapter.Collect(ch)
		close(ch)

		// Should receive no metrics
		count := 0
		for range ch {
			count++
		}
		if count != 0 {
			t.Errorf("Expected 0 metrics on error, got %d", count)
		}
	})
}

func TestCollectorFactory(t *testing.T) {
	// Test factory registration
	factoryCalled := false
	testFactory := func(cfg *config.CollectorConfig) (Collector, error) {
		factoryCalled = true
		return &mockRegistryCollector{
			name:    "factory_test",
			enabled: cfg.Enabled,
		}, nil
	}

	RegisterCollectorFactory("test", testFactory)

	// Verify factory is registered
	if _, exists := RegisteredCollectorFactories["test"]; !exists {
		t.Error("Factory should be registered")
	}

	// Test factory usage
	cfg := &config.CollectorConfig{Enabled: true}
	collector, err := RegisteredCollectorFactories["test"](cfg)
	if err != nil {
		t.Errorf("Factory returned error: %v", err)
	}
	if !factoryCalled {
		t.Error("Factory was not called")
	}
	if collector.Name() != "factory_test" {
		t.Errorf("Expected collector name 'factory_test', got '%s'", collector.Name())
	}
}
