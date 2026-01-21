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

func TestClusterCollector(t *testing.T) {
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
		Subsystem: "cluster",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "cluster", nil)

	// Create cluster collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewClusterCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "cluster" {
			t.Errorf("Expected name 'cluster', got '%s'", collector.Name())
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

		// Should have cluster overview metrics
		if count < 10 {
			t.Errorf("Expected at least 10 metric descriptions, got %d", count)
		}
	})

	t.Run("Collect", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

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

		// Should collect multiple metrics
		if count < 10 {
			t.Errorf("Expected at least 10 metrics, got %d", count)
		}

		// Metrics collection successful - logging is tested elsewhere
	})

	t.Run("CollectClusterInfo", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := collector.collectClusterInfo(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Cluster info collection failed: %v", err)
		}

		// Should emit cluster info metric
		count := 0
		for metric := range metricChan {
			count++
			// Verify metric family name contains cluster_info
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if !contains(fqName, "cluster_info") {
					t.Errorf("Expected cluster_info metric, got %s", fqName)
				}
			}
		}

		if count != 1 {
			t.Errorf("Expected 1 cluster info metric, got %d", count)
		}
	})

	t.Run("CollectNodeSummary", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 20)

		err := collector.collectNodeSummary(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Node summary collection failed: %v", err)
		}

		// Should emit multiple node-related metrics
		count := 0
		metricTypes := make(map[string]bool)
		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "capacity") {
					metricTypes["capacity"] = true
				}
				if contains(fqName, "allocated") {
					metricTypes["allocated"] = true
				}
				if contains(fqName, "utilization") {
					metricTypes["utilization"] = true
				}
				if contains(fqName, "node_states") {
					metricTypes["node_states"] = true
				}
			}
		}

		if count < 10 {
			t.Errorf("Expected at least 10 node metrics, got %d", count)
		}

		// Verify we got the expected metric types
		expectedTypes := []string{"capacity", "allocated", "utilization", "node_states"}
		for _, expectedType := range expectedTypes {
			if !metricTypes[expectedType] {
				t.Errorf("Expected to find %s metrics", expectedType)
			}
		}
	})

	t.Run("CollectJobSummary", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := collector.collectJobSummary(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Job summary collection failed: %v", err)
		}

		// Should emit job and user metrics
		count := 0
		jobMetricsFound := false
		userMetricsFound := false
		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "jobs_total") {
					jobMetricsFound = true
				}
				if contains(fqName, "users_total") {
					userMetricsFound = true
				}
			}
		}

		if count < 2 {
			t.Errorf("Expected at least 2 job metrics, got %d", count)
		}
		if !jobMetricsFound {
			t.Error("Expected to find job metrics")
		}
		if !userMetricsFound {
			t.Error("Expected to find user metrics")
		}
	})

	t.Run("CollectPartitionSummary", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 5)

		err := collector.collectPartitionSummary(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Partition summary collection failed: %v", err)
		}

		// Should emit partition count metric
		count := 0
		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if !contains(fqName, "partitions_total") {
					t.Errorf("Expected partitions_total metric, got %s", fqName)
				}
			}
		}

		if count != 1 {
			t.Errorf("Expected 1 partition metric, got %d", count)
		}
	})
}

func TestClusterCollectorIntegration(t *testing.T) {
	// Create a full integration test with registry
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_integration",
		Subsystem: "cluster",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_integration", "cluster", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create cluster collector
	var client interface{} = nil
	collector := NewClusterCollector(cfg, opts, client, metricDefs, "test-cluster")

	// Test direct collection without registry to avoid conflicts
	ctx := context.Background()
	metricChan := make(chan prometheus.Metric, 100)

	err = collector.Collect(ctx, metricChan)
	close(metricChan)

	if err != nil {
		t.Fatalf("Failed to collect metrics: %v", err)
	}

	// Count collected metrics
	metricCount := 0
	for range metricChan {
		metricCount++
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
}

// RegistryCollector adapts our Collector interface to prometheus.Collector
type RegistryCollector struct {
	collector Collector
}

func (rc *RegistryCollector) Describe(ch chan<- *prometheus.Desc) {
	rc.collector.Describe(ch)
}

func (rc *RegistryCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	// Ignore errors for this test
	_ = rc.collector.Collect(ctx, ch)
}
