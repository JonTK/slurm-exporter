// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

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
	t.Helper()
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
		for _, labelPair := range dto.GetLabel() {
			expectedValue, exists := labels[labelPair.GetName()]
			if exists && labelPair.GetValue() != expectedValue {
				labelMatch = false
				break
			}
		}

		if labelMatch {
			if dto.GetGauge() != nil {
				return dto.GetGauge().GetValue(), nil
			}
			if dto.GetCounter() != nil {
				return dto.GetCounter().GetValue(), nil
			}
		}
	}

	return 0, fmt.Errorf("metric %s with labels %v not found", metricName, labels)
}

// AssertMetricExists checks if a metric exists with the given labels
func AssertMetricExists(t *testing.T, collector prometheus.Collector, metricName string, labels prometheus.Labels) {
	t.Helper()
	_, err := GetMetricValue(collector, metricName, labels)
	assert.NoError(t, err, "metric %s with labels %v should exist", metricName, labels)
}

// AssertMetricValue checks if a metric has the expected value
func AssertMetricValue(t *testing.T, collector prometheus.Collector, metricName string, labels prometheus.Labels, expected float64) {
	t.Helper()
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
	t.Helper()
	err := registry.Register(collector)
	assert.NoError(t, err, "failed to register collector")
}

// DrainMetrics drains a metric channel and returns all collected metrics
func DrainMetrics(ch <-chan prometheus.Metric) []prometheus.Metric {
	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}
	return metrics
}

// AssertMetricWithLabels asserts that a specific metric exists with the given labels
// labelsMatch checks if the metric labels match expected labels
func labelsMatch(metricLabels []*io_prometheus_client.LabelPair, expectedLabels map[string]string) bool {
	for labelName, expectedLabelValue := range expectedLabels {
		found := false
		for _, labelPair := range metricLabels {
			if labelPair.GetName() == labelName && labelPair.GetValue() == expectedLabelValue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// getMetricValue extracts the numeric value from a metric
func getMetricValue(dto *io_prometheus_client.Metric) float64 {
	if dto.GetGauge() != nil {
		return dto.GetGauge().GetValue()
	}
	if dto.GetCounter() != nil {
		return dto.GetCounter().GetValue()
	}
	return 0
}

func AssertMetricWithLabels(t *testing.T, collector prometheus.Collector, metricName string, labels map[string]string, expectedValue float64) {
	t.Helper()
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	found := false
	for metric := range ch {
		dto := &io_prometheus_client.Metric{}
		_ = metric.Write(dto)

		// Check metric name
		desc := metric.Desc()
		if !strings.Contains(desc.String(), metricName) {
			continue
		}

		// Check labels
		if labelsMatch(dto.GetLabel(), labels) {
			found = true
			actualValue := getMetricValue(dto)
			assert.Equal(t, expectedValue, actualValue, "metric %s with labels %v", metricName, labels)
			break
		}
	}

	assert.True(t, found, "metric %s with labels %v not found", metricName, labels)
}

// AssertMetricCount asserts the count of metrics matching a pattern
func AssertMetricCount(t *testing.T, collector prometheus.Collector, namePattern string, expectedCount int) {
	t.Helper()
	ch := make(chan prometheus.Metric, 1000)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	count := 0
	for metric := range ch {
		desc := metric.Desc()
		if strings.Contains(desc.String(), namePattern) {
			count++
		}
	}

	assert.Equal(t, expectedCount, count, "expected %d metrics matching pattern %s, got %d", expectedCount, namePattern, count)
}
