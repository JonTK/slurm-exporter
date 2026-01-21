package collector

import (
	"context"
	"testing"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestPerformanceSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewPerformanceSimpleCollector(mockClient, logger)

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

func TestPerformanceSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockJobManager := new(mocks.MockJobManager)
	mockNodeManager := new(mocks.MockNodeManager)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup mock expectations
	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
		Version:     "23.02.1",
	}

	jobList := &interfaces.JobList{
		Jobs: []interfaces.Job{
			{
				ID:    "12345",
				Name:  "test-job",
				State: "RUNNING",
			},
		},
		Total: 1,
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

	partitionList := &interfaces.PartitionList{
		Partitions: []interfaces.Partition{
			{
				Name:  "compute",
				State: "UP",
			},
		},
		Total: 1,
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(jobList, nil)
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", mock.Anything, mock.Anything).Return(nodeList, nil)
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", mock.Anything, mock.Anything).Return(partitionList, nil)

	collector := NewPerformanceSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

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
	mockJobManager.AssertExpectations(t)
	mockNodeManager.AssertExpectations(t)
	mockPartitionManager.AssertExpectations(t)
}

func TestPerformanceSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewPerformanceSimpleCollector(mockClient, logger)
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
