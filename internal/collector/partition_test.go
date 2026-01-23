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

func TestPartitionCollector(t *testing.T) {
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
		Subsystem: "partition",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "partition", nil)

	// Create partition collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewPartitionCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "partition" {
			t.Errorf("Expected name 'partition', got '%s'", collector.Name())
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

		// Should have partition-level metrics
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

		// Should collect many partition metrics
		if count < 40 {
			t.Errorf("Expected at least 40 metrics, got %d", count)
		}
	})

	t.Run("CollectPartitionInfo", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

		err := collector.collectPartitionInfo(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Partition info collection failed: %v", err)
		}

		// Should emit partition info metrics for each partition
		count := 0
		partitionInfoFound := false
		partitionNodesFound := false
		partitionCPUsFound := false
		partitionMemoryFound := false
		partitionStateFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "partition_info") {
					partitionInfoFound = true
				}
				if contains(fqName, "partition_nodes") {
					partitionNodesFound = true
				}
				if contains(fqName, "partition_cpus") {
					partitionCPUsFound = true
				}
				if contains(fqName, "partition_memory") {
					partitionMemoryFound = true
				}
				if contains(fqName, "partition_state") {
					partitionStateFound = true
				}
			}
		}

		if count < 25 {
			t.Errorf("Expected at least 25 partition info metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !partitionInfoFound {
			t.Error("Expected to find partition_info metrics")
		}
		if !partitionNodesFound {
			t.Error("Expected to find partition_nodes metrics")
		}
		if !partitionCPUsFound {
			t.Error("Expected to find partition_cpus metrics")
		}
		if !partitionMemoryFound {
			t.Error("Expected to find partition_memory metrics")
		}
		if !partitionStateFound {
			t.Error("Expected to find partition_state metrics")
		}
	})

	t.Run("CollectPartitionUtilization", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectPartitionUtilization(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Partition utilization collection failed: %v", err)
		}

		// Should emit utilization metrics
		count := 0
		jobCountFound := false
		allocatedCPUsFound := false
		allocatedMemoryFound := false
		utilizationFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "job_count") {
					jobCountFound = true
				}
				if contains(fqName, "allocated_cpus") {
					allocatedCPUsFound = true
				}
				if contains(fqName, "allocated_memory") {
					allocatedMemoryFound = true
				}
				if contains(fqName, "utilization") {
					utilizationFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 utilization metrics, got %d", count)
		}
		if !jobCountFound {
			t.Error("Expected to find job_count metrics")
		}
		if !allocatedCPUsFound {
			t.Error("Expected to find allocated_cpus metrics")
		}
		if !allocatedMemoryFound {
			t.Error("Expected to find allocated_memory metrics")
		}
		if !utilizationFound {
			t.Error("Expected to find utilization metrics")
		}
	})

	t.Run("CollectPartitionPolicies", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 20)

		err := collector.collectPartitionPolicies(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Partition policies collection failed: %v", err)
		}

		// Should emit policy metrics
		count := 0
		maxTimeFound := false
		defaultTimeFound := false
		priorityFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "max_time") {
					maxTimeFound = true
				}
				if contains(fqName, "default_time") {
					defaultTimeFound = true
				}
				if contains(fqName, "priority") {
					priorityFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 policy metrics, got %d", count)
		}
		if !maxTimeFound {
			t.Error("Expected to find max_time metrics")
		}
		if !defaultTimeFound {
			t.Error("Expected to find default_time metrics")
		}
		if !priorityFound {
			t.Error("Expected to find priority metrics")
		}
	})
}

func TestPartitionCollectorDataTypes(t *testing.T) {
	t.Run("PartitionInfo", func(t *testing.T) {
		partition := &PartitionInfo{
			Name:        "test_partition",
			State:       "UP",
			Nodes:       []string{"node01", "node02"},
			TotalNodes:  2,
			TotalCPUs:   96,
			TotalMemory: 128 * 1024 * 1024 * 1024, // 128GB
			AllowGroups: []string{"all"},
			AllowUsers:  []string{"all"},
			Default:     true,
			Hidden:      false,
			RootOnly:    false,
			Priority:    1,
			Limits: PartitionLimits{
				MaxTime:         24 * time.Hour,
				DefaultTime:     8 * time.Hour,
				MaxNodesPerJob:  2,
				MaxCPUsPerJob:   96,
				MaxMemoryPerJob: 128 * 1024 * 1024 * 1024, // 128GB
			},
		}

		if partition.Name != "test_partition" {
			t.Errorf("Expected name 'test_partition', got '%s'", partition.Name)
		}
		if partition.State != "UP" {
			t.Errorf("Expected state 'UP', got '%s'", partition.State)
		}
		if !partition.Default {
			t.Error("Expected partition to be default")
		}
		if len(partition.Nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(partition.Nodes))
		}
	})

	t.Run("PartitionLimits", func(t *testing.T) {
		limits := &PartitionLimits{
			MaxTime:         48 * time.Hour,
			DefaultTime:     8 * time.Hour,
			MaxNodes:        10,
			MaxNodesPerJob:  5,
			MaxCPUs:         480,
			MaxCPUsPerJob:   240,
			MaxMemory:       512 * 1024 * 1024 * 1024, // 512GB
			MaxMemoryPerJob: 256 * 1024 * 1024 * 1024, // 256GB
			MaxJobsPerUser:  100,
			MaxJobsTotal:    1000,
		}

		if limits.MaxTime != 48*time.Hour {
			t.Errorf("Expected max time 48h, got %v", limits.MaxTime)
		}
		if limits.MaxNodesPerJob != 5 {
			t.Errorf("Expected max nodes per job 5, got %d", limits.MaxNodesPerJob)
		}
	})

	t.Run("PartitionPolicies", func(t *testing.T) {
		policies := &PartitionPolicies{
			PreemptMode:       "REQUEUE",
			OversubscribeMode: "NO",
			Priority:          10,
			GraceTime:         30 * time.Second,
			PreemptType:       "partition_prio",
			SelectType:        "select/cons_tres",
			DefMemPerCPU:      2 * 1024 * 1024 * 1024, // 2GB
			MaxMemPerCPU:      8 * 1024 * 1024 * 1024, // 8GB
		}

		if policies.PreemptMode != "REQUEUE" {
			t.Errorf("Expected preempt mode 'REQUEUE', got '%s'", policies.PreemptMode)
		}
		if policies.Priority != 10 {
			t.Errorf("Expected priority 10, got %d", policies.Priority)
		}
	})
}

func TestPartitionCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_partition_integration",
		Subsystem: "partition",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_partition_integration", "partition", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create partition collector
	var client interface{} = nil
	collector := NewPartitionCollector(cfg, opts, client, metricDefs, "test-cluster")

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
			if contains(fqName, "partition_info") {
				metricTypes["info"]++
			} else if contains(fqName, "partition_state") {
				metricTypes["state"]++
			} else if contains(fqName, "nodes") || contains(fqName, "cpus") || contains(fqName, "memory") {
				metricTypes["resources"]++
			} else if contains(fqName, "utilization") || contains(fqName, "allocated") || contains(fqName, "job_count") {
				metricTypes["utilization"]++
			} else if contains(fqName, "time") || contains(fqName, "priority") {
				metricTypes["policies"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have all major categories
	expectedCategories := []string{"info", "state", "resources", "utilization", "policies"}
	for _, category := range expectedCategories {
		if metricTypes[category] == 0 {
			t.Errorf("Expected to find %s metrics", category)
		}
	}
}
