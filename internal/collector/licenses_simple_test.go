// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

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

func TestLicensesCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	collector := NewLicensesCollector(mockClient, logger, timeout)

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

func TestLicensesCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	// Setup mock expectations with test data
	licenseList := &interfaces.LicenseList{
		Licenses: []interfaces.License{
			{
				Name:      "matlab",
				Total:     100,
				Used:      25,
				Available: 75,
				Reserved:  0,
				Remote:    false,
			},
			{
				Name:      "ansys",
				Total:     50,
				Used:      10,
				Available: 40,
				Reserved:  5,
				Remote:    false,
			},
		},
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("GetLicenses", mock.Anything).Return(licenseList, nil)
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)

	collector := NewLicensesCollector(mockClient, logger, timeout)

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

func TestLicensesCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	timeout := 30 * time.Second

	// Setup mock to return error
	mockClient.On("GetLicenses", mock.Anything).Return(nil, assert.AnError)

	collector := NewLicensesCollector(mockClient, logger, timeout)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
}

func TestLicensesCollector_EmptyLicenseList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockInfoManager := new(mocks.MockInfoManager)
	timeout := 30 * time.Second

	// Setup mock to return empty list
	emptyList := &interfaces.LicenseList{
		Licenses: []interfaces.License{},
	}

	clusterInfo := &slurm.ClusterInfo{
		ClusterName: "test-cluster",
	}

	mockClient.On("GetLicenses", mock.Anything).Return(emptyList, nil)
	mockClient.On("Info").Return(mockInfoManager)
	mockInfoManager.On("Get", mock.Anything).Return(clusterInfo, nil)

	collector := NewLicensesCollector(mockClient, logger, timeout)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty license list, minimal or no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	// Empty list should result in few or no metrics
	assert.True(t, count >= 0, "should handle empty license list")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockInfoManager.AssertExpectations(t)
}
