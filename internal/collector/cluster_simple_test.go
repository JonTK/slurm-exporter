// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestClusterSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewClusterSimpleCollector(mockClient, logger)

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

func TestClusterSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)

	// Setup mock expectations with test data
	clusterInfo := &slurm.ClusterInfo{
		Version:     "23.02.1",
		Release:     "23.02",
		ClusterName: "test-cluster",
		APIVersion:  "v0.0.40",
		Uptime:      864000, // 10 days
	}

	clusterStats := &slurm.ClusterStats{
		TotalNodes:     100,
		IdleNodes:      20,
		AllocatedNodes: 80,
		TotalCPUs:      1000,
		IdleCPUs:       200,
		AllocatedCPUs:  800,
		TotalJobs:      50,
		RunningJobs:    30,
		PendingJobs:    15,
		CompletedJobs:  5,
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockInfoManager.On("Stats", mock.Anything).Return(clusterStats, nil)

	collector := NewClusterSimpleCollector(mockClient, logger)
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
	mockInfoManager.AssertExpectations(t)
}

func TestClusterSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewClusterSimpleCollector(mockClient, logger)
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

func TestClusterSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)

	// Setup mock to return error
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(nil, assert.AnError)

	collector := NewClusterSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}

func TestClusterSimpleCollector_Collect_StatsError(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)

	// Setup mock: Get succeeds, Stats fails
	clusterInfo := &slurm.ClusterInfo{
		Version:     "23.02.1",
		Release:     "23.02",
		ClusterName: "test-cluster",
		APIVersion:  "v0.0.40",
		Uptime:      864000,
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockInfoManager.On("Stats", mock.Anything).Return(nil, assert.AnError)

	collector := NewClusterSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should continue with info only
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should still have metrics from info
	count := 0
	for range ch {
		count++
	}

	assert.True(t, count > 0, "should have collected metrics from info even if stats failed")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}
