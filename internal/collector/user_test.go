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

func TestUserCollector(t *testing.T) {
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
		Subsystem: "user",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "user", nil)

	// Create user collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewUserCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "user" {
			t.Errorf("Expected name 'user', got '%s'", collector.Name())
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

		// Should have user and account metrics
		if count < 8 {
			t.Errorf("Expected at least 8 metric descriptions, got %d", count)
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

		// Should collect many user and account metrics
		if count < 30 {
			t.Errorf("Expected at least 30 metrics, got %d", count)
		}
	})

	t.Run("CollectUserStats", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

		err := collector.collectUserStats(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("User stats collection failed: %v", err)
		}

		// Should emit user metrics for each user
		count := 0
		userJobCountFound := false
		userCPUAllocatedFound := false
		userMemoryAllocatedFound := false
		userJobStatesFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "user_job_count") {
					userJobCountFound = true
				}
				if contains(fqName, "user_cpu_allocated") {
					userCPUAllocatedFound = true
				}
				if contains(fqName, "user_memory_allocated") {
					userMemoryAllocatedFound = true
				}
				if contains(fqName, "user_job_states") {
					userJobStatesFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 user metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !userJobCountFound {
			t.Error("Expected to find user_job_count metrics")
		}
		if !userCPUAllocatedFound {
			t.Error("Expected to find user_cpu_allocated metrics")
		}
		if !userMemoryAllocatedFound {
			t.Error("Expected to find user_memory_allocated metrics")
		}
		if !userJobStatesFound {
			t.Error("Expected to find user_job_states metrics")
		}
	})

	t.Run("CollectAccountStats", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

		err := collector.collectAccountStats(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Account stats collection failed: %v", err)
		}

		// Should emit account metrics
		count := 0
		accountInfoFound := false
		accountJobCountFound := false
		accountCPUAllocatedFound := false
		accountMemoryAllocatedFound := false
		accountFairShareFound := false
		accountUsageFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "account_info") {
					accountInfoFound = true
				}
				if contains(fqName, "account_job_count") {
					accountJobCountFound = true
				}
				if contains(fqName, "account_cpu_allocated") {
					accountCPUAllocatedFound = true
				}
				if contains(fqName, "account_memory_allocated") {
					accountMemoryAllocatedFound = true
				}
				if contains(fqName, "account_fair_share") {
					accountFairShareFound = true
				}
				if contains(fqName, "account_usage") {
					accountUsageFound = true
				}
			}
		}

		if count < 20 {
			t.Errorf("Expected at least 20 account metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !accountInfoFound {
			t.Error("Expected to find account_info metrics")
		}
		if !accountJobCountFound {
			t.Error("Expected to find account_job_count metrics")
		}
		if !accountCPUAllocatedFound {
			t.Error("Expected to find account_cpu_allocated metrics")
		}
		if !accountMemoryAllocatedFound {
			t.Error("Expected to find account_memory_allocated metrics")
		}
		if !accountFairShareFound {
			t.Error("Expected to find account_fair_share metrics")
		}
		if !accountUsageFound {
			t.Error("Expected to find account_usage metrics")
		}
	})
}

func TestUserCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_user_integration",
		Subsystem: "user",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_user_integration", "user", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create user collector
	var client interface{} = nil
	collector := NewUserCollector(cfg, opts, client, metricDefs, "test-cluster")

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
			if contains(fqName, "user_") {
				metricTypes["user"]++
			}
			if contains(fqName, "account_") {
				metricTypes["account"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have both user and account metrics
	if metricTypes["user"] == 0 {
		t.Error("Expected to find user metrics")
	}
	if metricTypes["account"] == 0 {
		t.Error("Expected to find account metrics")
	}

	// Test metrics by checking if they're registered
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	userMetricsFound := false
	accountMetricsFound := false
	for _, mf := range metricFamilies {
		name := mf.GetName()
		if contains(name, "user_") {
			userMetricsFound = true
		}
		if contains(name, "account_") {
			accountMetricsFound = true
		}
	}

	if !userMetricsFound {
		t.Error("Expected to find user metrics in registry")
	}
	if !accountMetricsFound {
		t.Error("Expected to find account metrics in registry")
	}
}
