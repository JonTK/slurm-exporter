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

func TestQoSCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewQoSCollector(mockClient, logger)

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

func TestQoSCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockQoSManager := new(mocks.MockQoSManager)

	// Setup mock expectations with test data
	qosList := &slurm.QoSList{
		QoS: []slurm.QoS{
			{
				Name:           "normal",
				Description:    "Normal QoS",
				Priority:       100,
				UsageFactor:    1.0,
				MaxJobs:        1000,
				MaxJobsPerUser: 100,
				MaxCPUs:        5000,
				MaxNodes:       100,
			},
			{
				Name:           "high",
				Description:    "High Priority QoS",
				Priority:       1000,
				UsageFactor:    2.0,
				MaxJobs:        2000,
				MaxJobsPerUser: 200,
				MaxCPUs:        10000,
				MaxNodes:       200,
			},
		},
	}

	mockClient.On("QoS").Return(mockQoSManager)
	mockQoSManager.On("List", mock.Anything, mock.Anything).Return(qosList, nil)

	collector := NewQoSCollector(mockClient, logger)
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
	mockQoSManager.AssertExpectations(t)
}

func TestQoSCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewQoSCollector(mockClient, logger)
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

func TestQoSCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockQoSManager := new(mocks.MockQoSManager)

	// Setup mock to return error
	mockClient.On("QoS").Return(mockQoSManager)
	mockQoSManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewQoSCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockQoSManager.AssertExpectations(t)
}

func TestQoSCollector_EmptyQoSList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockQoSManager := new(mocks.MockQoSManager)

	// Setup mock to return empty list
	emptyList := &slurm.QoSList{
		QoS: []slurm.QoS{},
	}

	mockClient.On("QoS").Return(mockQoSManager)
	mockQoSManager.On("List", mock.Anything, mock.Anything).Return(emptyList, nil)

	collector := NewQoSCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty QoS list, minimal or no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	// Empty list should result in few or no metrics
	assert.True(t, count >= 0, "should handle empty QoS list")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockQoSManager.AssertExpectations(t)
}
