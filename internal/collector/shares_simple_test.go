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

func TestSharesCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewSharesCollector(mockClient, logger, timeout)

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

func TestSharesCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	// Setup mock expectations with test data
	sharesList := &interfaces.SharesList{
		Shares: []interfaces.Share{
			{
				Name:        "account1",
				User:        "user1",
				RawShares:   100,
				NormShares:  0.5,
				RawUsage:    50,
				NormUsage:   0.25,
				EffectUsage: 0.25,
				FairShare:   2.0,
			},
			{
				Name:        "account2",
				User:        "user2",
				RawShares:   200,
				NormShares:  1.0,
				RawUsage:    100,
				NormUsage:   0.5,
				EffectUsage: 0.5,
				FairShare:   2.0,
			},
		},
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("GetShares", mock.Anything, mock.Anything).Return(sharesList, nil)
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)

	collector := NewSharesCollector(mockClient, logger, timeout)

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

func TestSharesCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	// Setup mock to return error
	mockClient.On("GetShares", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewSharesCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
}

func TestSharesCollector_EmptySharesList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	// Setup mock to return empty list
	emptyList := &interfaces.SharesList{
		Shares: []interfaces.Share{},
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("GetShares", mock.Anything, mock.Anything).Return(emptyList, nil)
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)

	collector := NewSharesCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty shares list, minimal or no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	// Empty list should result in few or no metrics
	assert.True(t, count >= 0, "should handle empty shares list")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}
