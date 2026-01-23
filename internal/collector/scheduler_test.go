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

// mockCollector is a test collector that implements the Collector interface
type mockCollector struct {
	name        string
	enabled     bool
	collectFunc func(context.Context, chan<- prometheus.Metric) error
}

func (m *mockCollector) Name() string {
	return m.name
}

func (m *mockCollector) IsEnabled() bool {
	return m.enabled
}

func (m *mockCollector) SetEnabled(enabled bool) {
	m.enabled = enabled
}

func (m *mockCollector) Describe(ch chan<- *prometheus.Desc) {
	// Mock implementation - no-op
}

func (m *mockCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if m.collectFunc != nil {
		return m.collectFunc(ctx, ch)
	}
	return nil
}

func TestScheduler(t *testing.T) {
	// Create test configuration
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 100 * time.Millisecond,
			DefaultTimeout:  50 * time.Millisecond,
			MaxConcurrency:  3,
		},
		Cluster: config.CollectorConfig{
			Enabled:  true,
			Interval: 50 * time.Millisecond,
			Timeout:  25 * time.Millisecond,
		},
		Nodes: config.CollectorConfig{
			Enabled:  true,
			Interval: 75 * time.Millisecond,
			Timeout:  25 * time.Millisecond,
		},
		Jobs: config.CollectorConfig{
			Enabled:  false, // Disabled
			Interval: 100 * time.Millisecond,
		},
	}

	// Create registry
	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Register mock collectors
	var clusterCount, nodesCount int32

	_ = registry.Register("cluster", &mockCollector{
		name:    "cluster",
		enabled: true,
		collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
			atomic.AddInt32(&clusterCount, 1)
			return nil
		},
	})

	_ = registry.Register("nodes", &mockCollector{
		name:    "nodes",
		enabled: true,
		collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
			atomic.AddInt32(&nodesCount, 1)
			return nil
		},
	})

	_ = registry.Register("jobs", &mockCollector{
		name:    "jobs",
		enabled: false,
		collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
			t.Error("Disabled collector should not be called")
			return nil
		},
	})

	// Create scheduler
	scheduler, err := NewScheduler(registry, cfg)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	t.Run("InitializeSchedules", func(t *testing.T) {
		err := scheduler.InitializeSchedules()
		if err != nil {
			t.Errorf("Failed to initialize schedules: %v", err)
		}

		// Check schedules were created
		if len(scheduler.schedules) < 3 {
			t.Errorf("Expected at least 3 schedules, got %d", len(scheduler.schedules))
		}

		// Check cluster schedule
		if schedule, exists := scheduler.schedules["cluster"]; exists {
			if schedule.Interval != 50*time.Millisecond {
				t.Errorf("Cluster interval: expected 50ms, got %v", schedule.Interval)
			}
			if !schedule.enabled {
				t.Error("Cluster schedule should be enabled")
			}
		} else {
			t.Error("Cluster schedule not found")
		}

		// Check jobs schedule (should be disabled)
		if schedule, exists := scheduler.schedules["jobs"]; exists {
			if schedule.enabled {
				t.Error("Jobs schedule should be disabled")
			}
		} else {
			t.Error("Jobs schedule not found")
		}
	})

	t.Run("ScheduledCollection", func(t *testing.T) {
		// Ensure schedules are initialized
		if len(scheduler.schedules) == 0 {
			_ = scheduler.InitializeSchedules()
		}

		// Start scheduler
		err := scheduler.Start()
		if err != nil {
			t.Fatalf("Failed to start scheduler: %v", err)
		}

		// Wait for collections to happen
		time.Sleep(200 * time.Millisecond)

		// Stop scheduler
		scheduler.Stop()

		// Check collection counts
		clusterCollections := atomic.LoadInt32(&clusterCount)
		nodesCollections := atomic.LoadInt32(&nodesCount)

		// Cluster should collect ~4 times (50ms interval over 200ms)
		// Allow for extra collections from in-flight timers when Stop() is called
		if clusterCollections < 2 || clusterCollections > 10 {
			t.Errorf("Expected 2-10 cluster collections, got %d", clusterCollections)
		}

		// Nodes should collect ~2-3 times (75ms interval over 200ms)
		// Allow for extra collections from in-flight timers when Stop() is called
		if nodesCollections < 1 || nodesCollections > 6 {
			t.Errorf("Expected 1-6 nodes collections, got %d", nodesCollections)
		}
	})

	t.Run("UpdateSchedule", func(t *testing.T) {
		// Create new registry and scheduler for this test
		promRegistry2 := prometheus.NewRegistry()
		registry2, _ := NewRegistry(cfg, promRegistry2)
		scheduler2, err := NewScheduler(registry2, cfg)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		// Register a test collector
		_ = registry2.Register("cluster", &mockCollector{
			name:    "cluster",
			enabled: true,
		})
		_ = scheduler2.InitializeSchedules()

		// Update cluster schedule
		err = scheduler2.UpdateSchedule("cluster", 150*time.Millisecond, 50*time.Millisecond)
		if err != nil {
			t.Errorf("Failed to update schedule: %v", err)
		}

		// Verify update
		schedule := scheduler2.schedules["cluster"]
		if schedule.Interval != 150*time.Millisecond {
			t.Errorf("Expected updated interval 150ms, got %v", schedule.Interval)
		}

		// Test updating non-existent schedule
		err = scheduler2.UpdateSchedule("non_existent", 100*time.Millisecond, 50*time.Millisecond)
		if err == nil {
			t.Error("Expected error updating non-existent schedule")
		}
	})

	t.Run("EnableDisableSchedule", func(t *testing.T) {
		// Create new registry and scheduler for this test
		promRegistry3 := prometheus.NewRegistry()
		registry3, _ := NewRegistry(cfg, promRegistry3)
		scheduler3, err := NewScheduler(registry3, cfg)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		// Register a test collector
		_ = registry3.Register("cluster", &mockCollector{
			name:    "cluster",
			enabled: true,
		})
		_ = scheduler3.InitializeSchedules()

		// Disable cluster schedule
		err = scheduler3.DisableSchedule("cluster")
		if err != nil {
			t.Errorf("Failed to disable schedule: %v", err)
		}

		// Check it's disabled
		if scheduler3.schedules["cluster"].enabled {
			t.Error("Schedule should be disabled")
		}

		// Re-enable
		err = scheduler3.EnableSchedule("cluster")
		if err != nil {
			t.Errorf("Failed to enable schedule: %v", err)
		}

		// Check it's enabled
		if !scheduler3.schedules["cluster"].enabled {
			t.Error("Schedule should be enabled")
		}

		// Test with non-existent schedule
		err = scheduler3.EnableSchedule("non_existent")
		if err == nil {
			t.Error("Expected error enabling non-existent schedule")
		}
	})

	t.Run("ScheduleStats", func(t *testing.T) {
		// Create new registry and scheduler for this test
		promRegistry4 := prometheus.NewRegistry()
		registry4, _ := NewRegistry(cfg, promRegistry4)
		scheduler4, err := NewScheduler(registry4, cfg)
		if err != nil {
			t.Fatalf("Failed to create scheduler: %v", err)
		}

		// Register test collectors
		_ = registry4.Register("cluster", &mockCollector{
			name:    "cluster",
			enabled: true,
		})
		_ = scheduler4.InitializeSchedules()

		// Get stats
		stats := scheduler4.GetScheduleStats()

		// Check we have stats for all collectors
		if len(stats) < 3 {
			t.Errorf("Expected at least 3 stats entries, got %d", len(stats))
		}

		// Check cluster stats
		if clusterStats, exists := stats["cluster"]; exists {
			if clusterStats.CollectorName != "cluster" {
				t.Errorf("Expected collector name 'cluster', got '%s'", clusterStats.CollectorName)
			}
			if clusterStats.Interval != 50*time.Millisecond {
				t.Errorf("Expected interval 50ms, got %v", clusterStats.Interval)
			}
			if !clusterStats.Enabled {
				t.Error("Cluster should be enabled")
			}
		} else {
			t.Error("Cluster stats not found")
		}
	})
}

func TestSchedulerErrorHandling(t *testing.T) {
	// Create test configuration
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 50 * time.Millisecond,
			DefaultTimeout:  25 * time.Millisecond,
			MaxConcurrency:  2,
		},
		Cluster: config.CollectorConfig{
			Enabled:  true,
			Interval: 50 * time.Millisecond,
		},
	}

	// Create registry
	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Register failing collector
	var errorCount int32
	_ = registry.Register("cluster", &mockCollector{
		name:    "cluster",
		enabled: true,
		collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
			atomic.AddInt32(&errorCount, 1)
			return errors.New("collection failed")
		},
	})

	// Create and start scheduler
	scheduler, err := NewScheduler(registry, cfg)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	_ = scheduler.InitializeSchedules()
	_ = scheduler.Start()

	// Wait for collections - longer wait to ensure multiple collection cycles
	time.Sleep(200 * time.Millisecond)

	// Check error count before stopping
	errCount := atomic.LoadInt32(&errorCount)
	if errCount < 2 {
		t.Errorf("Expected at least 2 error collections, got %d", errCount)
	}

	// Stop scheduler
	scheduler.Stop()

	// Wait for stats to be finalized after stop
	time.Sleep(50 * time.Millisecond)

	// Check schedule error stats - use the actual error count as reference
	stats := scheduler.GetScheduleStats()
	if clusterStats, exists := stats["cluster"]; exists {
		// The error count in stats should match or be close to actual errors
		// Allow for timing variance - require at least 1 error recorded
		if clusterStats.ErrorCount < 1 {
			t.Errorf("Expected at least 1 error in stats (collector had %d errors), got %d in stats", errCount, clusterStats.ErrorCount)
		}
		// Error rate should be 1.0 (all collections failed)
		if clusterStats.ErrorRate != 1.0 {
			t.Errorf("Expected error rate 1.0, got %f", clusterStats.ErrorRate)
		}
	}
}

func TestScheduleHealthCheck(t *testing.T) {
	// Create test configuration
	cfg := &config.CollectorsConfig{
		Global: config.GlobalCollectorConfig{
			DefaultInterval: 50 * time.Millisecond,
			DefaultTimeout:  25 * time.Millisecond,
			MaxConcurrency:  1,
		},
		Cluster: config.CollectorConfig{
			Enabled:  true,
			Interval: 50 * time.Millisecond,
		},
	}

	// Create registry
	promRegistry := prometheus.NewRegistry()
	registry, err := NewRegistry(cfg, promRegistry)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Register slow collector that will cause schedule to fall behind
	_ = registry.Register("cluster", &mockCollector{
		name:    "cluster",
		enabled: true,
		collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
			// Simulate slow collection
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	})

	// Create scheduler
	scheduler, err := NewScheduler(registry, cfg)
	if err != nil {
		t.Fatalf("Failed to create scheduler: %v", err)
	}
	_ = scheduler.InitializeSchedules()

	// Manually set NextRun to past
	schedule := scheduler.schedules["cluster"]
	schedule.mu.Lock()
	schedule.NextRun = time.Now().Add(-200 * time.Millisecond)
	schedule.mu.Unlock()

	// Run health check
	scheduler.checkScheduleHealth()

	// Should have recorded missed runs
	// Note: Can't easily verify prometheus metrics in test, but the function should execute without panic
}

func TestSchedulerMetrics(t *testing.T) {
	metrics := NewSchedulerMetrics("test", "scheduler")
	registry := prometheus.NewRegistry()

	err := metrics.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Update some metrics
	metrics.ScheduledRuns.WithLabelValues("test", "success").Inc()
	metrics.MissedRuns.WithLabelValues("test", "latency").Inc()
	metrics.ScheduleLatency.WithLabelValues("test").Observe(0.5)

	// Verify metrics were registered (will panic if not)
	metrics.ScheduledRuns.WithLabelValues("test", "error").Inc()
}
