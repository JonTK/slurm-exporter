// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

func TestNodeCollector(t *testing.T) {
	// Create test logger
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	logEntry := logrus.NewEntry(logger)

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
		Namespace: "slurm",
		Subsystem: "node",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "node", nil)

	// Create node collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewNodeCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "node" {
			t.Errorf("Expected name 'node', got '%s'", collector.Name())
		}
	})

	t.Run("Enabled", func(t *testing.T) {
		if !collector.IsEnabled() {
			t.Error("Collector should be enabled")
		}
	})

	t.Run("Describe", func(t *testing.T) {
		descChan := make(chan *prometheus.Desc, 100)

		collector.Describe(descChan)
		close(descChan)

		// Count descriptions
		count := 0
		for range descChan {
			count++
		}

		// Should have node-level metrics
		if count < 10 {
			t.Errorf("Expected at least 10 metric descriptions, got %d", count)
		}
	})

	t.Run("Collect", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 200)

		err := collector.Collect(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Collection failed: %v", err)
		}

		// Count collected metrics
		count := 0
		for range metricChan {
			count++
		}

		// Should collect many node metrics
		if count < 20 {
			t.Errorf("Expected at least 20 metrics, got %d", count)
		}
	})

	t.Run("CollectNodeInfo", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectNodeInfo(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Node info collection failed: %v", err)
		}

		// Should emit node info metrics for each node
		count := 0
		nodeInfoFound := false
		nodeCPUsFound := false
		nodeMemoryFound := false
		nodeWeightFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "node_info") {
					nodeInfoFound = true
				}
				if contains(fqName, "node_cpus") {
					nodeCPUsFound = true
				}
				if contains(fqName, "node_memory") {
					nodeMemoryFound = true
				}
				if contains(fqName, "node_weight") {
					nodeWeightFound = true
				}
			}
		}

		if count < 10 {
			t.Errorf("Expected at least 10 node info metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !nodeInfoFound {
			t.Error("Expected to find node_info metrics")
		}
		if !nodeCPUsFound {
			t.Error("Expected to find node_cpus metrics")
		}
		if !nodeMemoryFound {
			t.Error("Expected to find node_memory metrics")
		}
		if !nodeWeightFound {
			t.Error("Expected to find node_weight metrics")
		}
	})

	t.Run("CollectNodeStates", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectNodeStates(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Node states collection failed: %v", err)
		}

		// Should emit state metrics for each node and state combination
		count := 0
		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if !contains(fqName, "node_state") {
					t.Errorf("Expected node_state metric, got %s", fqName)
				}
			}
		}

		// 3 nodes * 8 states = 24 metrics
		if count < 20 {
			t.Errorf("Expected at least 20 node state metrics, got %d", count)
		}
	})

	t.Run("CollectNodeUtilization", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 20)

		err := collector.collectNodeUtilization(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Node utilization collection failed: %v", err)
		}

		// Should emit utilization metrics
		count := 0
		allocatedCPUsFound := false
		allocatedMemoryFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "allocated_cpus") {
					allocatedCPUsFound = true
				}
				if contains(fqName, "allocated_memory") {
					allocatedMemoryFound = true
				}
			}
		}

		if count < 6 {
			t.Errorf("Expected at least 6 utilization metrics, got %d", count)
		}
		if !allocatedCPUsFound {
			t.Error("Expected to find allocated_cpus metrics")
		}
		if !allocatedMemoryFound {
			t.Error("Expected to find allocated_memory metrics")
		}
	})

	t.Run("CollectNodeHealth", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

		err := collector.collectNodeHealth(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Node health collection failed: %v", err)
		}

		// Should emit health metrics
		count := 0
		loadFound := false
		uptimeFound := false
		heartbeatFound := false
		featuresFound := false
		gpusFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "load_average") {
					loadFound = true
				}
				if contains(fqName, "uptime") {
					uptimeFound = true
				}
				if contains(fqName, "heartbeat") {
					heartbeatFound = true
				}
				if contains(fqName, "features") {
					featuresFound = true
				}
				if contains(fqName, "gpus") {
					gpusFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 health metrics, got %d", count)
		}
		if !loadFound {
			t.Error("Expected to find load_average metrics")
		}
		if !uptimeFound {
			t.Error("Expected to find uptime metrics")
		}
		if !heartbeatFound {
			t.Error("Expected to find heartbeat metrics")
		}
		if !featuresFound {
			t.Error("Expected to find features metrics")
		}
		if !gpusFound {
			t.Error("Expected to find gpus metrics")
		}
	})
}

func TestNodeCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_node_integration",
		Subsystem: "node",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_node_integration", "node", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create node collector
	var client interface{} = nil
	collector := NewNodeCollector(cfg, opts, client, metricDefs, "test-cluster")

	// Test direct collection without registry to avoid conflicts
	ctx := context.Background()
	metricChan := make(chan prometheus.Metric, 200)

	err = collector.Collect(ctx, metricChan)
	close(metricChan)

	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Count collected metrics
	metricCount := 0
	metricTypes := make(map[string]int)
	for metric := range metricChan {
		metricCount++
		desc := metric.Desc()
		if desc != nil {
			fqName := desc.String()
			if contains(fqName, "node_info") {
				metricTypes["info"]++
			} else if contains(fqName, "node_state") {
				metricTypes["state"]++
			} else if contains(fqName, "allocated") {
				metricTypes["utilization"]++
			} else if contains(fqName, "load") || contains(fqName, "uptime") || contains(fqName, "heartbeat") {
				metricTypes["health"]++
			} else if contains(fqName, "features") || contains(fqName, "gpus") {
				metricTypes["hardware"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have all major categories
	expectedCategories := []string{"info", "state", "utilization", "health", "hardware"}
	for _, category := range expectedCategories {
		if metricTypes[category] == 0 {
			t.Errorf("Expected to find %s metrics", category)
		}
	}
}
