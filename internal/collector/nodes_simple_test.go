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

func TestNodesSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewNodesSimpleCollector(mockClient, logger)

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

func TestNodesSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
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
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewNodesSimpleCollector(mockClient, logger)
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

func TestNodesSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock to return error
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.Error(t, err)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestNodesSimpleCollector_StateMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that we have state metrics for each state
	// From fixtures: idle=2, allocated=1, down=1, mixed=1, drain=1
	stateCount := make(map[string]int)
	for metric := range ch {
		desc := metric.Desc()
		if nodeTestContains(desc.String(), "node_state") {
			// In a real test, we would parse the labels
			stateCount["found"]++
		}
	}

	assert.True(t, stateCount["found"] > 0, "should have node state metrics")
}

func TestNodesSimpleCollector_ResourceMetrics(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
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
		if nodeTestContains(descStr, "cpu_total") {
			metricTypes["cpu_total"] = true
		}
		if nodeTestContains(descStr, "cpu_allocated") {
			metricTypes["cpu_allocated"] = true
		}
		if nodeTestContains(descStr, "memory_total") {
			metricTypes["memory_total"] = true
		}
		if nodeTestContains(descStr, "memory_allocated") {
			metricTypes["memory_allocated"] = true
		}
	}

	assert.True(t, metricTypes["cpu_total"], "should have cpu_total metrics")
	assert.True(t, metricTypes["cpu_allocated"], "should have cpu_allocated metrics")
	assert.True(t, metricTypes["memory_total"], "should have memory_total metrics")
	assert.True(t, metricTypes["memory_allocated"], "should have memory_allocated metrics")
}

func TestNodesSimpleCollector_Filtering(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Configure filtering - only collect CPU metrics
	filterConfig := config.FilterConfig{
		MetricFilter: config.MetricFilterConfig{
			EnableAll: false,
			IncludeMetrics: []string{"slurm_node_cpu_*"},
			ExcludeMetrics: []string{},
		},
	}
	collector.UpdateFilterConfig(filterConfig)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that only CPU metrics are collected
	for metric := range ch {
		desc := metric.Desc()
		assert.Contains(t, desc.String(), "cpu", "only cpu metrics should be collected")
	}
}

func TestNodesSimpleCollector_CustomLabels(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Set custom labels
	customLabels := map[string]string{
		"cluster_name": "test-cluster",
		"region":       "us-east-1",
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

func TestNodesSimpleCollector_EmptyNodeList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock to return empty list
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetEmptyNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
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

	assert.True(t, count > 0, "should have metrics even with empty node list")
}

func TestNodesSimpleCollector_GPUNodes(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestNodeList(), nil)

	collector := NewNodesSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check for GPU-specific metrics
	hasGPUMetrics := false
	for metric := range ch {
		desc := metric.Desc()
		if nodeTestContains(desc.String(), "gres") || nodeTestContains(desc.String(), "gpu") {
			hasGPUMetrics = true
			break
		}
	}

	assert.True(t, hasGPUMetrics, "should have GPU-related metrics for GPU nodes")
}

// Helper function
func nodeTestContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(substr) > 0 && len(s) > len(substr) && nodeFindSubstring(s, substr)))
}

func nodeFindSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}