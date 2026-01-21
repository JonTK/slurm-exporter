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

func TestJobCollector(t *testing.T) {
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
		Subsystem: "job",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "job", nil)

	// Create job collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewJobCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "job" {
			t.Errorf("Expected name 'job', got '%s'", collector.Name())
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

		// Should have job-level metrics
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

		// Should collect many job metrics
		if count < 30 {
			t.Errorf("Expected at least 30 metrics, got %d", count)
		}
	})

	t.Run("CollectActiveJobs", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 100)

		err := collector.collectActiveJobs(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Active jobs collection failed: %v", err)
		}

		// Should emit job metrics for each active job
		count := 0
		jobInfoFound := false
		jobCPURequestedFound := false
		jobMemoryRequestedFound := false
		jobPriorityFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "job_info") {
					jobInfoFound = true
				}
				if contains(fqName, "cpus_requested") {
					jobCPURequestedFound = true
				}
				if contains(fqName, "memory_requested") {
					jobMemoryRequestedFound = true
				}
				if contains(fqName, "job_priority") {
					jobPriorityFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 active job metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !jobInfoFound {
			t.Error("Expected to find job_info metrics")
		}
		if !jobCPURequestedFound {
			t.Error("Expected to find cpus_requested metrics")
		}
		if !jobMemoryRequestedFound {
			t.Error("Expected to find memory_requested metrics")
		}
		if !jobPriorityFound {
			t.Error("Expected to find job_priority metrics")
		}
	})

	t.Run("CollectJobQueueStats", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectJobQueueStats(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Job queue stats collection failed: %v", err)
		}

		// Queue stats are histogram observations, so we don't get direct metrics back
		// The histograms will be collected when the registry gathers metrics
		// For now, just verify no error occurred
	})

	t.Run("CollectJobStatesSummary", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 20)

		err := collector.collectJobStatesSummary(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Job states summary collection failed: %v", err)
		}

		// Should emit job state metrics
		count := 0
		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if !contains(fqName, "job_states") {
					t.Errorf("Expected job_states metric, got %s", fqName)
				}
			}
		}

		if count < 6 {
			t.Errorf("Expected at least 6 job state metrics, got %d", count)
		}
	})
}

func TestJobCollectorUtilities(t *testing.T) {
	// Create test configuration for collector
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm",
		Subsystem: "job",
		Timeout:   30 * time.Second,
	}

	metricDefs := metrics.NewMetricDefinitions("slurm", "job", nil)
	var client interface{} = nil
	collector := NewJobCollector(cfg, opts, client, metricDefs, "test-cluster")

	t.Run("ParseJobState", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"PENDING", "pending"},
			{"PD", "pending"},
			{"RUNNING", "running"},
			{"R", "running"},
			{"COMPLETED", "completed"},
			{"CD", "completed"},
			{"CANCELLED", "cancelled"},
			{"CA", "cancelled"},
			{"FAILED", "failed"},
			{"F", "failed"},
			{"TIMEOUT", "timeout"},
			{"TO", "timeout"},
			{"NODE_FAIL", "node_fail"},
			{"NF", "node_fail"},
			{"PREEMPTED", "preempted"},
			{"PR", "preempted"},
			{"BOOT_FAIL", "boot_fail"},
			{"BF", "boot_fail"},
			{"DEADLINE", "deadline"},
			{"DL", "deadline"},
			{"OUT_OF_MEMORY", "out_of_memory"},
			{"OOM", "out_of_memory"},
			{"UNKNOWN_STATE", "unknown"},
		}

		for _, tc := range testCases {
			result := collector.parseJobState(tc.input)
			if result != tc.expected {
				t.Errorf("parseJobState(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		}
	})

}

func TestJobCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_job_integration",
		Subsystem: "job",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_job_integration", "job", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create job collector
	var client interface{} = nil
	collector := NewJobCollector(cfg, opts, client, metricDefs, "test-cluster")

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
			if contains(fqName, "job_info") {
				metricTypes["info"]++
			} else if contains(fqName, "job_states") {
				metricTypes["states"]++
			} else if contains(fqName, "requested") {
				metricTypes["requests"]++
			} else if contains(fqName, "allocated") {
				metricTypes["allocations"]++
			} else if contains(fqName, "priority") || contains(fqName, "start") || contains(fqName, "end") || contains(fqName, "exit") {
				metricTypes["scheduling"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have all major categories
	expectedCategories := []string{"info", "states", "requests", "allocations", "scheduling"}
	for _, category := range expectedCategories {
		if metricTypes[category] == 0 {
			t.Errorf("Expected to find %s metrics", category)
		}
	}

	// Test histogram metrics by checking if they're registered
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	histogramFound := false
	for _, mf := range metricFamilies {
		name := mf.GetName()
		if contains(name, "queue_time") || contains(name, "run_time") || contains(name, "wait_time") {
			histogramFound = true
			break
		}
	}

	if !histogramFound {
		t.Error("Expected to find histogram metrics in registry")
	}
}
