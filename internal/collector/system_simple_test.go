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

func TestSystemSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewSystemSimpleCollector(mockClient, logger)

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

func TestSystemSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	mockAccountManager := new(mocks.MockAccountManager)

	// Setup mock expectations
	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
		Version:     "23.02.1",
	}

	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)
	mockInfoManager.On("Ping", mock.Anything).Return(nil)
	mockClient.On("Accounts").Return(mockAccountManager)
	mockAccountManager.On("List", mock.Anything, mock.Anything).Return(&slurm.AccountList{}, nil)

	collector := NewSystemSimpleCollector(mockClient, logger)
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
	mockAccountManager.AssertExpectations(t)
}

func TestSystemSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewSystemSimpleCollector(mockClient, logger)
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
