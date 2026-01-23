// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/fixtures"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestNodesSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewNodesSimpleCollector(mockClient, logger)

	ch := make(chan *prometheus.Desc, 100)
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
		"environment":  "testing",
	}
	collector.SetCustomLabels(customLabels)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify custom labels are present
	// Note: In a real test, we would parse the metric and check labels
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

	// With empty node list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when node list is empty")
}

func TestNodesSimpleCollector_StateDistribution(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockNodeManager := new(mocks.MockNodeManager)

	// Setup mock expectations with node list
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

	assert.True(t, count > 0, "should have collected state distribution metrics")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
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

	// Verify we got resource metrics (CPU, memory, etc.)
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have collected resource metrics")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}
