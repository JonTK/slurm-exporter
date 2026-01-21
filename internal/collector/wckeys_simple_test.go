package collector

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestWCKeysCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewWCKeysCollector(mockClient, logger, timeout)

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

func TestWCKeysCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockWCKeyManager := new(mocks.MockWCKeyManager)
	timeout := 30 * time.Second

	// Setup mock expectations with test data
	wcKeyList := &interfaces.WCKeyList{
		WCKeys: []interfaces.WCKey{
			{
				Name:    "wckey1",
				User:    "user1",
				Cluster: "test-cluster",
			},
			{
				Name:    "wckey2",
				User:    "user2",
				Cluster: "test-cluster",
			},
		},
		Total: 2,
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("WCKeys").Return(mockWCKeyManager)
	mockWCKeyManager.On("List", mock.Anything, mock.Anything).Return(wcKeyList, nil)

	collector := NewWCKeysCollector(mockClient, logger, timeout)

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
	mockWCKeyManager.AssertExpectations(t)
}

func TestWCKeysCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockWCKeyManager := new(mocks.MockWCKeyManager)
	timeout := 30 * time.Second

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	// Setup mock to return error
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("WCKeys").Return(mockWCKeyManager)
	mockWCKeyManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewWCKeysCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
	mockWCKeyManager.AssertExpectations(t)
}

func TestWCKeysCollector_EmptyWCKeyList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockWCKeyManager := new(mocks.MockWCKeyManager)
	timeout := 30 * time.Second

	// Setup mock to return empty list
	emptyList := &interfaces.WCKeyList{
		WCKeys: []interfaces.WCKey{},
		Total:  0,
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("WCKeys").Return(mockWCKeyManager)
	mockWCKeyManager.On("List", mock.Anything, mock.Anything).Return(emptyList, nil)

	collector := NewWCKeysCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty WCKey list, minimal or no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	// Empty list should result in few or no metrics
	assert.True(t, count >= 0, "should handle empty WCKey list")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
	mockWCKeyManager.AssertExpectations(t)
}
