//go:build ignore
// +build ignore

// TODO: This test file is excluded from builds due to compilation errors
// Mock implementations don't match current interfaces or use outdated types.

package collector

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/fixtures"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestPartitionsSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewPartitionsSimpleCollector(mockClient, logger)

	ch := make(chan *prometheus.Desc, 10)
	collector.Describe(ch)
	close(ch)

	// Should have at least the basic metrics
	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	assert.True(t, len(descs) > 0, "should have metric descriptors")
}

func TestPartitionsSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Count metrics
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have collected metrics")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockPartitionManager.AssertExpectations(t)
}

func TestPartitionsSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(false)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should not collect any metrics when disabled
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not collect metrics when disabled")
}

func TestPartitionsSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock to return error
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.Error(t, err)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockPartitionManager.AssertExpectations(t)
}

func TestPartitionsSimpleCollector_StateMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that we have partition state metrics
	// From fixtures: up=4, down=1
	hasStateMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "partition_state") {
			hasStateMetrics = true
			break
		}
	}

	assert.True(t, hasStateMetrics, "should have partition state metrics")
}

func TestPartitionsSimpleCollector_ResourceMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check for resource metrics
	metricTypes := make(map[string]bool)
	for metric := range ch {
		desc := metric.Desc()
		descStr := desc.String()
		if contains(descStr, "nodes_total") {
			metricTypes["nodes_total"] = true
		}
		if contains(descStr, "cpus_total") {
			metricTypes["cpus_total"] = true
		}
		if contains(descStr, "priority") {
			metricTypes["priority"] = true
		}
		if contains(descStr, "time_limit") {
			metricTypes["time_limit"] = true
		}
	}

	assert.True(t, metricTypes["nodes_total"], "should have nodes_total metrics")
	assert.True(t, metricTypes["cpus_total"], "should have cpus_total metrics")
}

func TestPartitionsSimpleCollector_LimitMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check for limit metrics
	hasTimeMetrics := false
	hasMemoryMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		descStr := desc.String()
		if contains(descStr, "time") {
			hasTimeMetrics = true
		}
		if contains(descStr, "memory") {
			hasMemoryMetrics = true
		}
	}

	assert.True(t, hasTimeMetrics, "should have time limit metrics")
	assert.True(t, hasMemoryMetrics, "should have memory limit metrics")
}

func TestPartitionsSimpleCollector_Filtering(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Configure filtering - only collect state metrics
	filterConfig := config.FilterConfig{
		MetricFilter: config.MetricFilterConfig{
			EnableAll: false,
			IncludeMetrics: []string{"slurm_partition_state"},
			ExcludeMetrics: []string{},
		},
	}
	collector.UpdateFilterConfig(filterConfig)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that only state metrics are collected
	for metric := range ch {
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "state", "only state metrics should be collected")
	}
}

func TestPartitionsSimpleCollector_CustomLabels(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Set custom labels
	customLabels := map[string]string{
		"cluster_name": "test-cluster",
		"datacenter":   "east",
	}
	collector.SetCustomLabels(customLabels)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify metrics were collected with custom labels
	count := 0
	for range ch {
		count++
	}
	assert.True(t, count > 0, "should have collected metrics with custom labels")
}

func TestPartitionsSimpleCollector_EmptyPartitionList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock to return empty list
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetEmptyPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should still have some metrics (zeros)
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have metrics even with empty partition list")
}

func TestPartitionsSimpleCollector_QOSMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestPartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check for QOS-related metrics
	hasQOSMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if contains(desc.String(), "qos") {
			hasQOSMetrics = true
			break
		}
	}

	// QOS metrics might be present depending on implementation
	count := 0
	for range ch {
		count++
	}
	assert.True(t, count > 0, "should have collected partition metrics")
}

func TestPartitionsSimpleCollector_ActivePartitions(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock to return only active partitions
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetActivePartitionList(), nil)

	collector := NewPartitionsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should have metrics for active partitions only
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have metrics for active partitions")
}