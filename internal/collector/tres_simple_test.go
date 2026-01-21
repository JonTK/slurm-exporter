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

func TestTRESCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewTRESCollector(mockClient, logger, timeout)

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

func TestTRESCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockNodeManager := new(mocks.MockNodeManager)
	timeout := 30 * time.Second

	// Setup mock expectations with test data
	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	now := time.Now()
	tresList := &interfaces.TRESList{
		TRES: []interfaces.TRES{
			{
				ID:       1,
				Type:     "cpu",
				Name:     "cpu",
				Count:    1000,
				Created:  now,
				Modified: now,
			},
			{
				ID:       2,
				Type:     "mem",
				Name:     "mem",
				Count:    102400, // 100GB
				Created:  now,
				Modified: now,
			},
			{
				ID:       3,
				Type:     "node",
				Name:     "node",
				Count:    10,
				Created:  now,
				Modified: now,
			},
			{
				ID:          4,
				Type:        "gres",
				Name:        "gpu",
				Count:       8,
				Created:     now,
				Modified:    now,
				Description: "GPU resources",
			},
		},
	}

	nodeList := &interfaces.NodeList{
		Nodes: []interfaces.Node{
			{
				Name:  "node1",
				State: "IDLE",
			},
		},
		Total: 1,
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("GetTRES", mock.Anything).Return(tresList, nil)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nodeList, nil)

	collector := NewTRESCollector(mockClient, logger, timeout)

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
	mockInfoManager.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}

func TestTRESCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	// Setup mock to return error
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("GetTRES", mock.Anything).Return(nil, assert.AnError)

	collector := NewTRESCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}

func TestTRESCollector_EmptyTRESList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockNodeManager := new(mocks.MockNodeManager)
	timeout := 30 * time.Second

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	// Setup mock to return empty list
	emptyList := &interfaces.TRESList{
		TRES: []interfaces.TRES{},
	}

	nodeList := &interfaces.NodeList{
		Nodes: []interfaces.Node{},
		Total: 0,
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("GetTRES", mock.Anything).Return(emptyList, nil)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nodeList, nil)

	collector := NewTRESCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty TRES list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when TRES list is empty")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
}
