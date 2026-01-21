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

func TestSelfMonitoringCollector(t *testing.T) {
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
		Subsystem: "selfmon",
		Timeout:   30 * time.Second,
		Logger:    logEntry,
	}

	// Create metric definitions
	metricDefs := metrics.NewMetricDefinitions("slurm", "selfmon", nil)

	// Create self-monitoring collector
	var client interface{} = nil // Mock client
	clusterName := "test-cluster"
	collector := NewSelfMonitoringCollector(cfg, opts, client, metricDefs, clusterName)

	t.Run("Name", func(t *testing.T) {
		if collector.Name() != "selfmon" {
			t.Errorf("Expected name 'selfmon', got '%s'", collector.Name())
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

		// Should have self-monitoring metrics
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

		// Should collect self-monitoring metrics (mainly gauges)
		if count < 15 {
			t.Errorf("Expected at least 15 metrics, got %d", count)
		}
	})

	t.Run("CollectRuntimeMetrics", func(t *testing.T) {
		hook.Reset()

		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectRuntimeMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Runtime metrics collection failed: %v", err)
		}

		// Should emit runtime metrics
		count := 0
		collectorUpFound := false
		lastCollectionFound := false
		metricsExportedFound := false

		for metric := range metricChan {
			count++
			desc := metric.Desc()
			if desc != nil {
				fqName := desc.String()
				if contains(fqName, "collector_up") {
					collectorUpFound = true
				}
				if contains(fqName, "last_collection") {
					lastCollectionFound = true
				}
				if contains(fqName, "metrics_exported") {
					metricsExportedFound = true
				}
			}
		}

		if count < 15 {
			t.Errorf("Expected at least 15 runtime metrics, got %d", count)
		}

		// Verify we got the expected metric types
		if !collectorUpFound {
			t.Error("Expected to find collector_up metrics")
		}
		if !lastCollectionFound {
			t.Error("Expected to find last_collection metrics")
		}
		if !metricsExportedFound {
			t.Error("Expected to find metrics_exported metrics")
		}
	})

	t.Run("CollectPerformanceMetrics", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectPerformanceMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Performance metrics collection failed: %v", err)
		}

		// Performance metrics are histograms and counters that don't send to the channel
		// in the same way as gauges, so we mainly check for no errors
		count := 0
		for range metricChan {
			count++
		}

		// The performance metrics primarily update existing metric instances
		// rather than sending new ones, so count may be low or zero
		t.Logf("Performance metrics collection completed, channel received %d metrics", count)
	})

	t.Run("CollectAPIMetrics", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectAPIMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("API metrics collection failed: %v", err)
		}

		// API metrics are histograms and counters that don't send to the channel
		// in the same way as gauges, so we mainly check for no errors
		count := 0
		for range metricChan {
			count++
		}

		t.Logf("API metrics collection completed, channel received %d metrics", count)
	})

	t.Run("CollectCacheMetrics", func(t *testing.T) {
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 50)

		err := collector.collectCacheMetrics(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Cache metrics collection failed: %v", err)
		}

		// Cache metrics are counters that don't send to the channel
		// in the same way as gauges, so we mainly check for no errors
		count := 0
		for range metricChan {
			count++
		}

		t.Logf("Cache metrics collection completed, channel received %d metrics", count)
	})
}

func TestSelfMonitoringCollectorUtilities(t *testing.T) {
	// Create test configuration for collector
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm",
		Subsystem: "selfmon",
		Timeout:   30 * time.Second,
	}

	metricDefs := metrics.NewMetricDefinitions("slurm", "selfmon", nil)
	var client interface{} = nil
	collector := NewSelfMonitoringCollector(cfg, opts, client, metricDefs, "test-cluster")

	t.Run("CalculateCacheHitRatio", func(t *testing.T) {
		// Test cache hit ratio calculation
		testCases := []struct {
			hits     int
			misses   int
			expected float64
		}{
			{100, 0, 1.0},
			{75, 25, 0.75},
			{50, 50, 0.5},
			{0, 100, 0.0},
			{0, 0, 0.0},
		}

		for _, tc := range testCases {
			ratio := collector.calculateCacheHitRatio(tc.hits, tc.misses)
			if ratio != tc.expected {
				t.Errorf("calculateCacheHitRatio(%d, %d) = %.3f, expected %.3f",
					tc.hits, tc.misses, ratio, tc.expected)
			}
		}
	})

	t.Run("GetCollectorHealth", func(t *testing.T) {
		// Test collector health check
		health := collector.getCollectorHealth()
		expectedCollectors := []string{
			"cluster", "node", "job", "user", "partition", "performance", "selfmon",
		}

		for _, collectorName := range expectedCollectors {
			if isUp, exists := health[collectorName]; !exists {
				t.Errorf("Expected collector '%s' in health status", collectorName)
			} else if !isUp {
				t.Errorf("Expected collector '%s' to be up", collectorName)
			}
		}
	})

	t.Run("ParseSelfMonitoringData", func(t *testing.T) {
		// Test parsing self-monitoring data
		selfMonData, err := collector.parseSelfMonitoringData(nil)
		if err != nil {
			t.Errorf("parseSelfMonitoringData failed: %v", err)
		}
		if selfMonData == nil {
			t.Error("Expected self-monitoring data, got nil")
		}
		if selfMonData.RuntimeMetrics.GoroutineCount != 25 {
			t.Errorf("Expected goroutine count 25, got %d", selfMonData.RuntimeMetrics.GoroutineCount)
		}
		if len(selfMonData.PerformanceMetrics.CollectionDurations) == 0 {
			t.Error("Expected collection durations, got empty map")
		}
		if len(selfMonData.APIMetrics.EndpointDurations) == 0 {
			t.Error("Expected API endpoint durations, got empty map")
		}
		if len(selfMonData.CacheMetrics.CacheHits) == 0 {
			t.Error("Expected cache hits, got empty map")
		}
	})
}

func TestSelfMonitoringCollectorDataTypes(t *testing.T) {
	t.Run("SelfMonitoringMetrics", func(t *testing.T) {
		metrics := &SelfMonitoringMetrics{
			RuntimeMetrics: RuntimeMetrics{
				Uptime:         2 * time.Hour,
				MemoryUsage:    100 * 1024 * 1024, // 100MB
				GoroutineCount: 30,
				CPUUsage:       5.5,
				CollectorsUp: map[string]bool{
					"cluster": true,
					"node":    true,
					"job":     false,
				},
				LastCollection: map[string]time.Time{
					"cluster": time.Now().Add(-1 * time.Minute),
					"node":    time.Now().Add(-2 * time.Minute),
				},
			},
			PerformanceMetrics: CollectorPerformanceMetrics{
				CollectionDurations: map[string]time.Duration{
					"cluster": 2 * time.Second,
					"node":    5 * time.Second,
				},
				CollectionErrors: map[string]map[string]int{
					"node": {"timeout": 2},
				},
				CollectionSuccess: map[string]int{
					"cluster": 100,
					"node":    98,
				},
				MetricsExported: map[string]int{
					"cluster": 25,
					"node":    150,
				},
			},
		}

		if metrics.RuntimeMetrics.Uptime != 2*time.Hour {
			t.Errorf("Expected uptime 2h, got %v", metrics.RuntimeMetrics.Uptime)
		}
		if metrics.RuntimeMetrics.GoroutineCount != 30 {
			t.Errorf("Expected goroutine count 30, got %d", metrics.RuntimeMetrics.GoroutineCount)
		}
		if !metrics.RuntimeMetrics.CollectorsUp["cluster"] {
			t.Error("Expected cluster collector to be up")
		}
		if metrics.RuntimeMetrics.CollectorsUp["job"] {
			t.Error("Expected job collector to be down")
		}
	})

	t.Run("RuntimeMetrics", func(t *testing.T) {
		runtime := &RuntimeMetrics{
			Uptime:         3 * time.Hour,
			MemoryUsage:    75 * 1024 * 1024, // 75MB
			GoroutineCount: 22,
			CPUUsage:       3.2,
			CollectorsUp: map[string]bool{
				"cluster":     true,
				"node":        true,
				"job":         true,
				"partition":   false,
				"performance": true,
			},
		}

		if runtime.Uptime != 3*time.Hour {
			t.Errorf("Expected uptime 3h, got %v", runtime.Uptime)
		}
		if runtime.MemoryUsage != 75*1024*1024 {
			t.Errorf("Expected memory usage 75MB, got %d", runtime.MemoryUsage)
		}
		if runtime.CPUUsage != 3.2 {
			t.Errorf("Expected CPU usage 3.2, got %.1f", runtime.CPUUsage)
		}
	})

	t.Run("CollectorPerformanceMetrics", func(t *testing.T) {
		performance := &CollectorPerformanceMetrics{
			CollectionDurations: map[string]time.Duration{
				"cluster": 1500 * time.Millisecond,
				"node":    4200 * time.Millisecond,
				"job":     6800 * time.Millisecond,
			},
			CollectionErrors: map[string]map[string]int{
				"node": {"timeout": 3, "connection_error": 1},
				"job":  {"api_error": 2},
			},
			CollectionSuccess: map[string]int{
				"cluster": 150,
				"node":    145,
				"job":     148,
			},
			MetricsExported: map[string]int{
				"cluster": 25,
				"node":    175,
				"job":     380,
			},
		}

		if performance.CollectionDurations["cluster"] != 1500*time.Millisecond {
			t.Errorf("Expected cluster duration 1500ms, got %v", performance.CollectionDurations["cluster"])
		}
		if performance.CollectionErrors["node"]["timeout"] != 3 {
			t.Errorf("Expected node timeout errors 3, got %d", performance.CollectionErrors["node"]["timeout"])
		}
		if performance.MetricsExported["job"] != 380 {
			t.Errorf("Expected job metrics exported 380, got %d", performance.MetricsExported["job"])
		}
	})

	t.Run("APIInteractionMetrics", func(t *testing.T) {
		api := &APIInteractionMetrics{
			EndpointDurations: map[string]time.Duration{
				"/slurm/v1/jobs":  120 * time.Millisecond,
				"/slurm/v1/nodes": 200 * time.Millisecond,
			},
			EndpointErrors: map[string]map[string]int{
				"/slurm/v1/nodes": {"timeout": 2},
				"/slurm/v1/jobs":  {"rate_limit": 1},
			},
			EndpointCalls: map[string]int{
				"/slurm/v1/jobs":  8,
				"/slurm/v1/nodes": 6,
			},
		}

		if api.EndpointDurations["/slurm/v1/jobs"] != 120*time.Millisecond {
			t.Errorf("Expected jobs endpoint duration 120ms, got %v", api.EndpointDurations["/slurm/v1/jobs"])
		}
		if api.EndpointErrors["/slurm/v1/nodes"]["timeout"] != 2 {
			t.Errorf("Expected nodes timeout errors 2, got %d", api.EndpointErrors["/slurm/v1/nodes"]["timeout"])
		}
		if api.EndpointCalls["/slurm/v1/jobs"] != 8 {
			t.Errorf("Expected jobs endpoint calls 8, got %d", api.EndpointCalls["/slurm/v1/jobs"])
		}
	})

	t.Run("CachePerformanceMetrics", func(t *testing.T) {
		cache := &CachePerformanceMetrics{
			CacheHits: map[string]int{
				"job_cache":  500,
				"node_cache": 300,
			},
			CacheMisses: map[string]int{
				"job_cache":  50,
				"node_cache": 20,
			},
			CacheSize: map[string]int{
				"job_cache":  1000,
				"node_cache": 500,
			},
			HitRatio: map[string]float64{
				"job_cache":  0.909, // 500/(500+50)
				"node_cache": 0.938, // 300/(300+20)
			},
		}

		if cache.CacheHits["job_cache"] != 500 {
			t.Errorf("Expected job cache hits 500, got %d", cache.CacheHits["job_cache"])
		}
		if cache.CacheMisses["node_cache"] != 20 {
			t.Errorf("Expected node cache misses 20, got %d", cache.CacheMisses["node_cache"])
		}
		expectedJobRatio := 500.0 / (500.0 + 50.0)
		if cache.HitRatio["job_cache"] < expectedJobRatio-0.01 || cache.HitRatio["job_cache"] > expectedJobRatio+0.01 {
			t.Errorf("Expected job cache hit ratio ~%.3f, got %.3f", expectedJobRatio, cache.HitRatio["job_cache"])
		}
	})
}

func TestSelfMonitoringCollectorIntegration(t *testing.T) {
	// Create a full integration test
	registry := prometheus.NewRegistry()

	// Create test configuration
	cfg := &config.CollectorConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
		Timeout:  10 * time.Second,
	}

	opts := &CollectorOptions{
		Namespace: "slurm_selfmon_integration",
		Subsystem: "selfmon",
		Timeout:   30 * time.Second,
	}

	// Create metric definitions and register them
	metricDefs := metrics.NewMetricDefinitions("slurm_selfmon_integration", "selfmon", nil)
	err := metricDefs.Register(registry)
	if err != nil {
		t.Fatalf("Failed to register metrics: %v", err)
	}

	// Create self-monitoring collector
	var client interface{} = nil
	collector := NewSelfMonitoringCollector(cfg, opts, client, metricDefs, "test-cluster")

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
			if contains(fqName, "collector_up") {
				metricTypes["up"]++
			} else if contains(fqName, "last_collection") {
				metricTypes["collection"]++
			} else if contains(fqName, "metrics_exported") {
				metricTypes["exported"]++
			}
		}
	}

	if metricCount == 0 {
		t.Error("Expected to collect metrics, got none")
	}

	t.Logf("Successfully collected %d metrics", metricCount)
	t.Logf("Metric breakdown: %+v", metricTypes)

	// Verify we have the major self-monitoring categories
	expectedCategories := []string{"up", "collection", "exported"}
	for _, category := range expectedCategories {
		if metricTypes[category] == 0 {
			t.Errorf("Expected to find %s metrics", category)
		}
	}

	// Test metrics by checking if they're registered in the registry
	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	selfMonMetricsFound := false
	for _, mf := range metricFamilies {
		name := mf.GetName()
		if contains(name, "collection_") || contains(name, "collector_") || contains(name, "exporter_") {
			selfMonMetricsFound = true
			break
		}
	}

	if !selfMonMetricsFound {
		t.Error("Expected to find self-monitoring metrics in registry")
	}
}
