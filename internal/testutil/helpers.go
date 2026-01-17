package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// GetTestLogger returns a test logger
func GetTestLogger() *logrus.Entry {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return logger.WithField("test", true)
}

// CollectAndCount collects metrics from a collector and returns the count
func CollectAndCount(collector prometheus.Collector) int {
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}
	return count
}

// CollectAndCompare collects metrics and compares with expected output
func CollectAndCompare(t *testing.T, collector prometheus.Collector, expected string, metricNames ...string) {
	err := testutil.CollectAndCompare(collector, strings.NewReader(expected), metricNames...)
	assert.NoError(t, err)
}

// GetMetricValue extracts a single metric value for testing
func GetMetricValue(collector prometheus.Collector, metricName string, labels prometheus.Labels) (float64, error) {
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	for metric := range ch {
		dto := &io_prometheus_client.Metric{}
		_ = metric.Write(dto)

		// Check if this is the metric we're looking for
		desc := metric.Desc()
		if !strings.Contains(desc.String(), metricName) {
			continue
		}

		// Check labels
		labelMatch := true
		for _, labelPair := range dto.Label {
			expectedValue, exists := labels[*labelPair.Name]
			if exists && *labelPair.Value != expectedValue {
				labelMatch = false
				break
			}
		}

		if labelMatch {
			if dto.Gauge != nil {
				return *dto.Gauge.Value, nil
			}
			if dto.Counter != nil {
				return *dto.Counter.Value, nil
			}
		}
	}

	return 0, fmt.Errorf("metric %s with labels %v not found", metricName, labels)
}

// AssertMetricExists checks if a metric exists with the given labels
func AssertMetricExists(t *testing.T, collector prometheus.Collector, metricName string, labels prometheus.Labels) {
	_, err := GetMetricValue(collector, metricName, labels)
	assert.NoError(t, err, "metric %s with labels %v should exist", metricName, labels)
}

// AssertMetricValue checks if a metric has the expected value
func AssertMetricValue(t *testing.T, collector prometheus.Collector, metricName string, labels prometheus.Labels, expected float64) {
	value, err := GetMetricValue(collector, metricName, labels)
	assert.NoError(t, err)
	assert.Equal(t, expected, value, "metric %s with labels %v should have value %f", metricName, labels, expected)
}

// CreateTestRegistry creates a test prometheus registry
func CreateTestRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

// MustRegister registers a collector and panics on error (for tests)
func MustRegister(t *testing.T, registry *prometheus.Registry, collector prometheus.Collector) {
	err := registry.Register(collector)
	assert.NoError(t, err, "failed to register collector")
}
