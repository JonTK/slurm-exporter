package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformanceMetrics(t *testing.T) {
	// Create a test registry
	registry := prometheus.NewRegistry()

	// Create performance metrics
	metrics := NewPerformanceMetrics("test")

	// Register with test registry
	err := metrics.Register(registry)
	require.NoError(t, err)

	// Test collection timer
	timer := metrics.NewCollectionTimer("test_collector", "collect")
	assert.NotNil(t, timer)

	// Sleep briefly and then stop timer
	time.Sleep(10 * time.Millisecond)
	duration := timer.Stop()
	assert.Greater(t, duration, time.Duration(0))

	// Test API timer
	apiTimer := metrics.NewAPITimer("test_endpoint", "GET")
	assert.NotNil(t, apiTimer)

	// Sleep longer to ensure measurable duration on Windows (15.6ms timer resolution)
	time.Sleep(50 * time.Millisecond)
	duration = apiTimer.StopWithStatus("success")
	assert.Greater(t, duration, time.Duration(0))

	// Test error recording
	metrics.RecordCollectionError("test_collector", "timeout")
	metrics.RecordAPIError("test_endpoint", "500", "server_error")

	// Test cardinality updates
	metrics.UpdateCardinality("test_metric", "test_collector", 100.0)

	// Test cache metrics
	metrics.RecordCacheHit("config", "test_collector")
	metrics.RecordCacheMiss("config", "test_collector")
	metrics.UpdateCacheSize("config", 1024.0)

	// Test system metrics
	metrics.UpdateMemoryUsage("heap", 50000000.0) // 50MB
	metrics.UpdateCPUUsage("total", 25.5)         // 25.5%

	// Test queue and adaptive metrics
	metrics.UpdateQueueDepth("test_collector", 5.0)
	metrics.UpdateAdaptiveInterval("test_collector", 30.0) // 30 seconds

	// Test circuit breaker state
	metrics.UpdateCircuitBreakerState("test_cb", 0.0) // closed

	// Test thread pool utilization
	metrics.UpdateThreadPoolUtilization("worker_pool", 75.0) // 75%

	// Verify metrics can be gathered
	families, err := registry.Gather()
	require.NoError(t, err)
	assert.Greater(t, len(families), 0)

	// Check that we have some expected metrics
	metricNames := make(map[string]bool)
	for _, family := range families {
		metricNames[*family.Name] = true
	}

	// Verify key metrics exist
	expectedMetrics := []string{
		"test_exporter_collection_duration_seconds",
		"test_exporter_api_duration_seconds",
		"test_exporter_collection_errors_total",
		"test_exporter_metric_cardinality",
		"test_exporter_cache_hits_total",
		"test_exporter_memory_usage_bytes",
		"test_exporter_cpu_usage_percent",
	}

	for _, expected := range expectedMetrics {
		assert.True(t, metricNames[expected], "Expected metric %s not found", expected)
	}
}

func TestCacheMetricsRecorder(t *testing.T) {
	metrics := NewPerformanceMetrics("test")
	recorder := NewCacheMetricsRecorder(metrics)

	// Test cache operations
	recorder.RecordHit("config", "test_collector")
	recorder.RecordMiss("config", "test_collector")
	recorder.UpdateSize("config", 2048.0)

	// Should not panic or error
	assert.NotNil(t, recorder)
}

func TestCardinalityTracker(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPerformanceMetrics("test")

	// Register some test metrics
	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "A test counter",
	})
	registry.MustRegister(testCounter)
	testCounter.Inc()

	tracker := NewCardinalityTracker(metrics, registry, 100*time.Millisecond)
	assert.NotNil(t, tracker)

	// Update should work
	tracker.Update()

	// Wait briefly and update again
	time.Sleep(150 * time.Millisecond)
	tracker.Update()

	// Should not panic
}

func TestTimerZeroDuration(t *testing.T) {
	metrics := NewPerformanceMetrics("test")

	// Create timer and stop immediately
	timer := metrics.NewCollectionTimer("test", "phase")
	duration := timer.Stop()

	// Duration should be very small but >= 0
	assert.GreaterOrEqual(t, duration, time.Duration(0))
}

func TestMetricsDoubleRegistration(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewPerformanceMetrics("test")

	// First registration should succeed
	err := metrics.Register(registry)
	require.NoError(t, err)

	// Second registration should fail (metric already registered)
	err = metrics.Register(registry)
	assert.Error(t, err)
}

func TestUpdateFunctions(t *testing.T) {
	metrics := NewPerformanceMetrics("test")

	// Test all update functions don't panic
	metrics.UpdateCardinality("metric1", "collector1", 50.0)
	metrics.UpdateCacheSize("cache1", 1024.0)
	metrics.UpdateMemoryUsage("heap", 100000.0)
	metrics.UpdateCPUUsage("user", 15.5)
	metrics.UpdateQueueDepth("collector1", 3.0)
	metrics.UpdateAdaptiveInterval("collector1", 60.0)
	metrics.UpdateCircuitBreakerState("breaker1", 1.0)
	metrics.UpdateThreadPoolUtilization("pool1", 80.0)

	// Should not panic
	assert.NotNil(t, metrics)
}
