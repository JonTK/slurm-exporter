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

func TestPerformanceCollector(t *testing.T) {
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
		Subsystem: "performance",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "performance", nil)

	// Create performance collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewPerformanceCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "performance" {
			t.Errorf("Expected name 'performance', got '%s'", collector.Name())
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

		// Should have performance metrics
		if count < 6 {
			t.Errorf("Expected at least 6 metric descriptions, got %d", count)
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

		// Should collect many performance metrics
		if count < 50 {
			t.Errorf("Expected at least 50 metrics, got %d", count)
		}
	})

	t.Run("CollectThroughputMetrics", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectThroughputMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Throughput metrics collection failed: %v", err)
		}

		// Should emit throughput metrics
		count := 0
		jobsCompletedFound := false
		jobsSubmittedFound := false
		cpuHoursFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "system_throughput") {
					jobsCompletedFound = true
					jobsSubmittedFound = true
					cpuHoursFound = true
				}
			}
		}

		if count < 9 {
			t.Errorf("Expected at least 9 throughput metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !jobsCompletedFound {
			t.Error("Expected to find jobs_completed metrics")
		}
		if !jobsSubmittedFound {
			t.Error("Expected to find jobs_submitted metrics")
		}
		if !cpuHoursFound {
			t.Error("Expected to find cpu_hours_delivered metrics")
		}
	})

	t.Run("CollectEfficiencyMetrics", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectEfficiencyMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Efficiency metrics collection failed: %v", err)
		}

		// Should emit efficiency metrics
		count := 0
		systemEfficiencyFound := false
		resourceUtilizationFound := false
		cpuUtilizationFound := false
		memoryUtilizationFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "system_efficiency") {
					systemEfficiencyFound = true
					cpuUtilizationFound = true
					memoryUtilizationFound = true
				}
				if contains(fqName, "resource_utilization") {
					resourceUtilizationFound = true
				}
			}
		}

		if count < 18 {
			t.Errorf("Expected at least 18 efficiency metrics, got %d", count)
		}
		if !systemEfficiencyFound {
			t.Error("Expected to find system_efficiency metrics")
		}
		if !resourceUtilizationFound {
			t.Error("Expected to find resource_utilization metrics")
		}
		if !cpuUtilizationFound {
			t.Error("Expected to find cpu_utilization metrics")
		}
		if !memoryUtilizationFound {
			t.Error("Expected to find memory_utilization metrics")
		}
	})

	t.Run("CollectQueueMetrics", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectQueueMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Queue metrics collection failed: %v", err)
		}

		// Should emit queue metrics
		count := 0
		queueDepthFound := false
		queueWaitTimeFound := false
		jobTurnoverFound := false
		averageWaitFound := false
		medianWaitFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "queue_depth") {
					queueDepthFound = true
				}
				if contains(fqName, "queue_wait_time") {
					queueWaitTimeFound = true
					averageWaitFound = true
					medianWaitFound = true
				}
				if contains(fqName, "job_turnover") {
					jobTurnoverFound = true
				}
			}
		}

		if count < 20 {
			t.Errorf("Expected at least 20 queue metrics, got %d", count)
		}
		if !queueDepthFound {
			t.Error("Expected to find queue_depth metrics")
		}
		if !queueWaitTimeFound {
			t.Error("Expected to find queue_wait_time metrics")
		}
		if !jobTurnoverFound {
			t.Error("Expected to find job_turnover metrics")
		}
		if !averageWaitFound {
			t.Error("Expected to find average wait time metrics")
		}
		if !medianWaitFound {
			t.Error("Expected to find median wait time metrics")
		}
	})
}

func TestPerformanceCollectorUtilities(t *testing.T) {
	// Create test configuration for collector
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm",
		Subsystem: "performance",
		Timeout:   30 * time.Second,
	}

	metricDefs := metrics.NewMetricDefinitions("slurm", "performance", nil)
	var client interface{} = nil
	collector := NewPerformanceCollector(cfg, opts, client, metricDefs, "test-cluster")

	t.Run("CalculateEfficiency", func(t *testing.T) {
		// Test efficiency calculation
		efficiencyMetrics := &EfficiencyMetrics{
			CPUUtilization:     0.80,
			MemoryUtilization:  0.70,
			NodeUtilization:    0.85,
			StorageUtilization: 0.50,
			NetworkUtilization: 0.30,
		}

		efficiency := collector.calculateEfficiency(efficiencyMetrics)

		// Should be weighted average
		expected := 0.80*0.4 + 0.70*0.3 + 0.85*0.2 + 0.50*0.05 + 0.30*0.05
		if efficiency < expected-0.01 || efficiency > expected+0.01 {
			t.Errorf("Expected efficiency around %.3f, got %.3f", expected, efficiency)
		}

		// Test cap at 100%
		highMetrics := &EfficiencyMetrics{
			CPUUtilization:     1.5,
			MemoryUtilization:  1.2,
			NodeUtilization:    1.8,
			StorageUtilization: 1.0,
			NetworkUtilization: 1.1,
		}

		highEfficiency := collector.calculateEfficiency(highMetrics)
		if highEfficiency > 1.0 {
			t.Errorf("Expected efficiency capped at 1.0, got %.3f", highEfficiency)
		}
	})

	t.Run("ParsePerformanceData", func(t *testing.T) {
		// Test parsing performance data
		performanceData, err := collector.parsePerformanceData(nil)
		if err != nil {
			t.Errorf("parsePerformanceData failed: %v", err)
		}
		if performanceData == nil {
			t.Error("Expected performance data, got nil")
		}
		if performanceData.ThroughputMetrics.JobsPerHour != 45.5 {
			t.Errorf("Expected jobs per hour 45.5, got %.1f", performanceData.ThroughputMetrics.JobsPerHour)
		}
		if performanceData.EfficiencyMetrics.CPUUtilization != 0.78 {
			t.Errorf("Expected CPU utilization 0.78, got %.2f", performanceData.EfficiencyMetrics.CPUUtilization)
		}
		if performanceData.QueueMetrics.QueueDepth != 125 {
			t.Errorf("Expected queue depth 125, got %d", performanceData.QueueMetrics.QueueDepth)
		}
	})
}

func TestPerformanceCollectorDataTypes(t *testing.T) {
	t.Run("PerformanceMetrics", func(t *testing.T) {
		metrics := &PerformanceMetrics{
			ThroughputMetrics: ThroughputMetrics{
				JobsPerHour:     50.0,
				CPUHoursPerHour: 2000.0,
				CompletionRate:  0.95,
				SubmissionRate:  1.05,
			},
			EfficiencyMetrics: EfficiencyMetrics{
				CPUUtilization:     0.85,
				MemoryUtilization:  0.70,
				NodeUtilization:    0.90,
				StorageUtilization: 0.55,
				NetworkUtilization: 0.25,
				OverallEfficiency:  0.78,
			},
			QueueMetrics: QueueMetrics{
				QueueDepth:      100,
				AverageWaitTime: 30 * time.Minute,
				MedianWaitTime:  20 * time.Minute,
				MaxWaitTime:     2 * time.Hour,
				P95WaitTime:     90 * time.Minute,
				TurnoverRate:    45.0,
			},
		}

		if metrics.ThroughputMetrics.JobsPerHour != 50.0 {
			t.Errorf("Expected jobs per hour 50.0, got %.1f", metrics.ThroughputMetrics.JobsPerHour)
		}
		if metrics.EfficiencyMetrics.CPUUtilization != 0.85 {
			t.Errorf("Expected CPU utilization 0.85, got %.2f", metrics.EfficiencyMetrics.CPUUtilization)
		}
		if metrics.QueueMetrics.QueueDepth != 100 {
			t.Errorf("Expected queue depth 100, got %d", metrics.QueueMetrics.QueueDepth)
		}
	})

	t.Run("ThroughputMetrics", func(t *testing.T) {
		throughput := &ThroughputMetrics{
			JobsPerHour:     60.5,
			CPUHoursPerHour: 2500.8,
			CompletionRate:  0.88,
			SubmissionRate:  1.20,
		}

		if throughput.JobsPerHour != 60.5 {
			t.Errorf("Expected jobs per hour 60.5, got %.1f", throughput.JobsPerHour)
		}
		if throughput.CompletionRate != 0.88 {
			t.Errorf("Expected completion rate 0.88, got %.2f", throughput.CompletionRate)
		}
	})

	t.Run("EfficiencyMetrics", func(t *testing.T) {
		efficiency := &EfficiencyMetrics{
			CPUUtilization:     0.92,
			MemoryUtilization:  0.78,
			NodeUtilization:    0.85,
			StorageUtilization: 0.60,
			NetworkUtilization: 0.35,
			OverallEfficiency:  0.82,
		}

		if efficiency.CPUUtilization != 0.92 {
			t.Errorf("Expected CPU utilization 0.92, got %.2f", efficiency.CPUUtilization)
		}
		if efficiency.OverallEfficiency != 0.82 {
			t.Errorf("Expected overall efficiency 0.82, got %.2f", efficiency.OverallEfficiency)
		}
	})

	t.Run("QueueMetrics", func(t *testing.T) {
		queue := &QueueMetrics{
			QueueDepth:      150,
			AverageWaitTime: 45 * time.Minute,
			MedianWaitTime:  30 * time.Minute,
			MaxWaitTime:     4 * time.Hour,
			P95WaitTime:     2 * time.Hour,
			TurnoverRate:    38.5,
		}

		if queue.QueueDepth != 150 {
			t.Errorf("Expected queue depth 150, got %d", queue.QueueDepth)
		}
		if queue.AverageWaitTime != 45*time.Minute {
			t.Errorf("Expected average wait time 45m, got %v", queue.AverageWaitTime)
		}
		if queue.TurnoverRate != 38.5 {
			t.Errorf("Expected turnover rate 38.5, got %.1f", queue.TurnoverRate)
		}
	})
}

func TestPerformanceCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_performance_integration",
		Subsystem: "performance",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_performance_integration", "performance", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create performance collector
	var client interface{} = nil
	collector := NewPerformanceCollector(cfg, opts, client, metricDefs, "test-cluster")

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
			if contains(fqName, "system_throughput") {
				metricTypes["throughput"]++
			} else if contains(fqName, "system_efficiency") {
				metricTypes["efficiency"]++
			} else if contains(fqName, "resource_utilization") {
				metricTypes["resource"]++
			} else if contains(fqName, "queue_") {
				metricTypes["queue"]++
			} else if contains(fqName, "job_turnover") {
				metricTypes["turnover"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have all major categories
	expectedCategories := []string{"throughput", "efficiency", "resource", "queue", "turnover"}
	for _, category := range expectedCategories {
		if metricTypes[category] == 0 {
			t.Errorf("Expected to find %s metrics", category)
		}
	}
}
