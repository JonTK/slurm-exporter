package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestPerformanceMonitor_NewPerformanceMonitor(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.logger)
}

func TestPerformanceMonitor_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	ch := make(chan *prometheus.Desc, 20)
	monitor.Describe(ch)
	close(ch)

	// Should have multiple metric descriptors
	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	assert.True(t, len(descs) > 5, "should have multiple metric descriptors")

	// Check for key metric descriptors
	found := make(map[string]bool)
	for _, desc := range descs {
		if contains(desc, "collection_duration") {
			found["duration"] = true
		}
		if contains(desc, "collection_errors") {
			found["errors"] = true
		}
		if contains(desc, "collection_success") {
			found["success"] = true
		}
		if contains(desc, "metrics_collected") {
			found["metrics"] = true
		}
	}

	assert.True(t, found["duration"], "should have duration metrics")
	assert.True(t, found["errors"], "should have error metrics")
	assert.True(t, found["success"], "should have success metrics")
	assert.True(t, found["metrics"], "should have metrics count")
}

func TestPerformanceMonitor_Collect(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should collect performance metrics
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should collect performance metrics")
}

func TestPerformanceMonitor_RecordCollectionDuration(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record some durations
	monitor.RecordCollectionDuration("jobs", 1.5)
	monitor.RecordCollectionDuration("nodes", 2.3)
	monitor.RecordCollectionDuration("partitions", 0.8)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have duration metrics
	hasDurationMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "collection_duration") {
			hasDurationMetrics = true
			break
		}
	}

	assert.True(t, hasDurationMetrics, "should have duration metrics")
}

func TestPerformanceMonitor_RecordCollectionError(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record some errors
	monitor.RecordCollectionError("jobs", "connection_timeout")
	monitor.RecordCollectionError("nodes", "api_error")
	monitor.RecordCollectionError("jobs", "connection_timeout") // Same error again

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have error metrics
	hasErrorMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "collection_errors") {
			hasErrorMetrics = true
			break
		}
	}

	assert.True(t, hasErrorMetrics, "should have error metrics")
}

func TestPerformanceMonitor_RecordCollectionSuccess(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record some successes
	monitor.RecordCollectionSuccess("jobs")
	monitor.RecordCollectionSuccess("nodes")
	monitor.RecordCollectionSuccess("partitions")

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have success metrics
	hasSuccessMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "collection_success") {
			hasSuccessMetrics = true
			break
		}
	}

	assert.True(t, hasSuccessMetrics, "should have success metrics")
}

func TestPerformanceMonitor_RecordMetricsCollected(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record metrics counts
	monitor.RecordMetricsCollected("jobs", 150)
	monitor.RecordMetricsCollected("nodes", 75)
	monitor.RecordMetricsCollected("partitions", 25)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have metrics count
	hasMetricsCount := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "metrics_collected") {
			hasMetricsCount = true
			break
		}
	}

	assert.True(t, hasMetricsCount, "should have metrics count")
}

func TestPerformanceMonitor_RecordSLAViolation(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record SLA violations
	monitor.RecordSLAViolation("jobs", "response_time")
	monitor.RecordSLAViolation("nodes", "error_rate")

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have SLA violation metrics
	hasSLAMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "sla_violations") {
			hasSLAMetrics = true
			break
		}
	}

	assert.True(t, hasSLAMetrics, "should have SLA violation metrics")
}

func TestPerformanceMonitor_RecordCacheHit(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record cache hits and misses
	monitor.RecordCacheHit("jobs")
	monitor.RecordCacheHit("nodes")
	monitor.RecordCacheMiss("partitions")

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have cache metrics
	hasCacheMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "cache") {
			hasCacheMetrics = true
			break
		}
	}

	assert.True(t, hasCacheMetrics, "should have cache metrics")
}

func TestPerformanceMonitor_UpdateActiveConnections(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Update active connections
	monitor.UpdateActiveConnections(5)
	monitor.UpdateActiveConnections(8)
	monitor.UpdateActiveConnections(3)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have connection metrics
	hasConnectionMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "active_connections") {
			hasConnectionMetrics = true
			break
		}
	}

	assert.True(t, hasConnectionMetrics, "should have connection metrics")
}

func TestPerformanceMonitor_RecordQueueSize(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record queue sizes
	monitor.RecordQueueSize("collection", 10)
	monitor.RecordQueueSize("processing", 5)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have queue metrics
	hasQueueMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "queue_size") {
			hasQueueMetrics = true
			break
		}
	}

	assert.True(t, hasQueueMetrics, "should have queue metrics")
}

func TestPerformanceMonitor_WithCollector_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	mockCollector := new(mocks.MockCollectorInterface)
	mockCollector.On("Collect", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	ch := make(chan prometheus.Metric, 100)

	err := monitor.WithCollector("test", mockCollector, func() error {
		return mockCollector.Collect(ctx, ch)
	})

	assert.NoError(t, err)
	mockCollector.AssertExpectations(t)

	// Should have recorded success metrics
	metricCh := make(chan prometheus.Metric, 100)
	monitor.Collect(metricCh)
	close(metricCh)

	hasSuccessMetrics := false
	for metric := range metricCh {
		desc := metric.Desc()
		if contains(desc.String(), "collection_success") {
			hasSuccessMetrics = true
			break
		}
	}

	assert.True(t, hasSuccessMetrics, "should record success metrics")
}

func TestPerformanceMonitor_WithCollector_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	mockCollector := new(mocks.MockCollectorInterface)
	mockCollector.On("Collect", mock.Anything, mock.Anything).Return(assert.AnError)

	ctx := context.Background()
	ch := make(chan prometheus.Metric, 100)

	err := monitor.WithCollector("test", mockCollector, func() error {
		return mockCollector.Collect(ctx, ch)
	})

	assert.Error(t, err)
	mockCollector.AssertExpectations(t)

	// Should have recorded error metrics
	metricCh := make(chan prometheus.Metric, 100)
	monitor.Collect(metricCh)
	close(metricCh)

	hasErrorMetrics := false
	for metric := range metricCh {
		desc := metric.Desc()
		if contains(desc.String(), "collection_errors") {
			hasErrorMetrics = true
			break
		}
	}

	assert.True(t, hasErrorMetrics, "should record error metrics")
}

func TestPerformanceMonitor_Timeout(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record a timeout
	monitor.RecordTimeout("jobs")

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have timeout metrics
	hasTimeoutMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "timeouts") {
			hasTimeoutMetrics = true
			break
		}
	}

	assert.True(t, hasTimeoutMetrics, "should have timeout metrics")
}

func TestPerformanceMonitor_StartupTime(t *testing.T) {
	logger := testutil.GetTestLogger()
	monitor := NewPerformanceMonitor(logger)

	// Record startup time
	startupTime := time.Now().Add(-1 * time.Minute)
	monitor.RecordStartupTime(startupTime)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	monitor.Collect(ch)
	close(ch)

	// Should have startup time metrics
	hasStartupMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "startup_time") {
			hasStartupMetrics = true
			break
		}
	}

	assert.True(t, hasStartupMetrics, "should have startup time metrics")
}