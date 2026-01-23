// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestClustersCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewClustersCollector(mockClient, logger, timeout)

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

func TestClustersCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockClusterManager := new(mocks.MockClusterManager)
	timeout := 30 * time.Second

	// Setup mock expectations with test data
	now := time.Now()
	clusterList := &interfaces.ClusterList{
		Clusters: []*interfaces.Cluster{
			{
				Name:               "cluster1",
				ControlHost:        "controller1.example.com",
				ControlPort:        6817,
				RPCVersion:         4002,
				PluginIDSelect:     101,
				PluginIDAuth:       102,
				PluginIDAcct:       103,
				TRESList:           []string{"cpu", "mem", "node", "gres/gpu"},
				Features:           []string{"avx", "avx2", "gpu"},
				FederationFeatures: []string{"batch", "interactive"},
				FederationState:    "ACTIVE",
				Created:            now,
				Modified:           now,
			},
			{
				Name:            "cluster2",
				ControlHost:     "controller2.example.com",
				ControlPort:     6817,
				RPCVersion:      4001,
				TRESList:        []string{"cpu", "mem", "node"},
				Features:        []string{"avx"},
				FederationState: "INACTIVE",
				Created:         now,
				Modified:        now,
			},
		},
		Total: 2,
	}

	mockClient.On("Clusters").Return(mockClusterManager)
	mockClusterManager.On("List", mock.Anything, mock.Anything).Return(clusterList, nil)

	collector := NewClustersCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
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
	mockClusterManager.AssertExpectations(t)
}

func TestClustersCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockClusterManager := new(mocks.MockClusterManager)
	timeout := 30 * time.Second

	// Setup mock to return error
	mockClient.On("Clusters").Return(mockClusterManager)
	mockClusterManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewClustersCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockClusterManager.AssertExpectations(t)
}

func TestClustersCollector_EmptyClusterList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockClusterManager := new(mocks.MockClusterManager)
	timeout := 30 * time.Second

	// Setup mock to return empty list
	emptyList := &interfaces.ClusterList{
		Clusters: []*interfaces.Cluster{},
		Total:    0,
	}

	mockClient.On("Clusters").Return(mockClusterManager)
	mockClusterManager.On("List", mock.Anything, mock.Anything).Return(emptyList, nil)

	collector := NewClustersCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty cluster list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when cluster list is empty")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockClusterManager.AssertExpectations(t)
}

func TestClustersCollector_NullClustersManager(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	// Setup mock to return nil manager
	mockClient.On("Clusters").Return(nil)

	collector := NewClustersCollector(mockClient, logger, timeout)

	// Collect metrics - should handle nil manager gracefully
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Should not collect any metrics
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not collect metrics when manager is nil")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
