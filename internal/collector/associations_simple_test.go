// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	"github.com/jontk/slurm-client/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestAssociationsSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewAssociationsSimpleCollector(mockClient, logger)

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

func TestAssociationsSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAssociationManager := new(mocks.MockAssociationManager)

	// Setup mock expectations with test data
	maxJobs := 100
	associationList := &interfaces.AssociationList{
		Associations: []*interfaces.Association{
			{
				ID:        1,
				User:      "user1",
				Account:   "account1",
				Cluster:   "test-cluster",
				IsDefault: true,
				SharesRaw: 100,
				Priority:  50,
				MaxJobs:   &maxJobs,
			},
			{
				ID:        2,
				User:      "user2",
				Account:   "account2",
				Cluster:   "test-cluster",
				IsDefault: false,
				SharesRaw: 200,
				Priority:  75,
			},
		},
		Total: 2,
	}

	mockClient.On("Associations").Return(mockAssociationManager)
	mockAssociationManager.On("List", mock.Anything, mock.Anything).Return(associationList, nil)

	collector := NewAssociationsSimpleCollector(mockClient, logger)
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
	mockAssociationManager.AssertExpectations(t)
}

func TestAssociationsSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewAssociationsSimpleCollector(mockClient, logger)
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

func TestAssociationsSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAssociationManager := new(mocks.MockAssociationManager)

	// Setup mock to return error
	mockClient.On("Associations").Return(mockAssociationManager)
	mockAssociationManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewAssociationsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockAssociationManager.AssertExpectations(t)
}

func TestAssociationsSimpleCollector_EmptyAssociationList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockAssociationManager := new(mocks.MockAssociationManager)

	// Setup mock to return empty list
	emptyList := &interfaces.AssociationList{
		Associations: []*interfaces.Association{},
		Total:        0,
	}

	mockClient.On("Associations").Return(mockAssociationManager)
	mockAssociationManager.On("List", mock.Anything, mock.Anything).Return(emptyList, nil)

	collector := NewAssociationsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty association list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when association list is empty")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockAssociationManager.AssertExpectations(t)
}
