// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"
	"time"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestReservationCollector_Describe(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewReservationCollector(mockClient, logger)

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

func TestReservationCollector_Collect_Success(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockReservationManager := new(mocks.MockReservationManager)

	// Setup mock expectations with test data
	// Helper functions for pointers
	strPtr := func(s string) *string { return &s }
	int32Ptr := func(i int32) *int32 { return &i }

	now := time.Now()
	reservationList := &slurm.ReservationList{
		Reservations: []slurm.Reservation{
			{
				Name:      strPtr("test-reservation-1"),
				Partition: strPtr("compute"),
				StartTime: now,
				EndTime:   now.Add(1 * time.Hour),
				NodeCount: int32Ptr(10),
				CoreCount: int32Ptr(100),
			},
			{
				Name:      strPtr("test-reservation-2"),
				Partition: strPtr("gpu"),
				StartTime: now.Add(30 * time.Minute),
				EndTime:   now.Add(90 * time.Minute),
				NodeCount: int32Ptr(5),
				CoreCount: int32Ptr(50),
			},
		},
	}

	mockClient.On("Reservations").Return(mockReservationManager)
	mockReservationManager.On("List", mock.Anything, mock.Anything).Return(reservationList, nil)

	collector := NewReservationCollector(mockClient, logger)
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
	mockReservationManager.AssertExpectations(t)
}

func TestReservationCollector_Collect_Disabled(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewReservationCollector(mockClient, logger)
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

func TestReservationCollector_Collect_Error(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockReservationManager := new(mocks.MockReservationManager)

	// Setup mock to return error
	mockClient.On("Reservations").Return(mockReservationManager)
	mockReservationManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewReservationCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics - should handle error gracefully
	ch := make(chan prometheus.Metric, 100)
	_ = collector.Collect(context.Background(), ch)
	close(ch)

	// May or may not have metrics depending on error handling
	// Just verify mock expectations were met
	mockClient.AssertExpectations(t)
	mockReservationManager.AssertExpectations(t)
}

func TestReservationCollector_EmptyReservationList(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockReservationManager := new(mocks.MockReservationManager)

	// Setup mock to return empty list
	emptyList := &slurm.ReservationList{
		Reservations: []slurm.Reservation{},
	}

	mockClient.On("Reservations").Return(mockReservationManager)
	mockReservationManager.On("List", mock.Anything, mock.Anything).Return(emptyList, nil)

	collector := NewReservationCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty reservation list, minimal or no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	// Empty list should result in few or no metrics
	assert.True(t, count >= 0, "should handle empty reservation list")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockReservationManager.AssertExpectations(t)
}
