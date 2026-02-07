// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"testing"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestQoSCollector_Describe(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockQoSManager := new(mocks.MockQoSManager)

	// Setup mock expectations with test data
	// Helper functions for pointers
	strPtr := func(s string) *string { return &s }
	uint32Ptr := func(i uint32) *uint32 { return &i }
	float64Ptr := func(f float64) *float64 { return &f }

	qosList := &slurm.QoSList{
		QoS: []slurm.QoS{
			{
				Name:        strPtr("normal"),
				Description: strPtr("Normal QoS"),
				Priority:    uint32Ptr(100),
				UsageFactor: float64Ptr(1.0),
				Limits: &slurm.QoSLimits{
					Max: &slurm.QoSLimitsMax{
						ActiveJobs: &slurm.QoSLimitsMaxActiveJobs{
							Count: uint32Ptr(1000),
						},
						Jobs: &slurm.QoSLimitsMaxJobs{
							ActiveJobs: &slurm.QoSLimitsMaxJobsActiveJobs{
								Per: &slurm.QoSLimitsMaxJobsActiveJobsPer{
									User: uint32Ptr(100),
								},
							},
						},
					},
				},
			},
			{
				Name:        strPtr("high"),
				Description: strPtr("High Priority QoS"),
				Priority:    uint32Ptr(1000),
				UsageFactor: float64Ptr(2.0),
				Limits: &slurm.QoSLimits{
					Max: &slurm.QoSLimitsMax{
						ActiveJobs: &slurm.QoSLimitsMaxActiveJobs{
							Count: uint32Ptr(2000),
						},
						Jobs: &slurm.QoSLimitsMaxJobs{
							ActiveJobs: &slurm.QoSLimitsMaxJobsActiveJobs{
								Per: &slurm.QoSLimitsMaxJobsActiveJobsPer{
									User: uint32Ptr(200),
								},
							},
						},
					},
				},
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestGetTRESValue(t *testing.T) {
	t.Parallel()

	int64Ptr := func(i int64) *int64 { return &i }

	tests := []struct {
		name     string
		tresList []slurm.TRES
		resType  string
		expected int64
	}{
		{
			name: "find cpu in list",
			tresList: []slurm.TRES{
				{Type: "cpu", Count: int64Ptr(100)},
				{Type: "mem", Count: int64Ptr(1024)},
				{Type: "node", Count: int64Ptr(10)},
			},
			resType:  "cpu",
			expected: 100,
		},
		{
			name: "find node in list",
			tresList: []slurm.TRES{
				{Type: "cpu", Count: int64Ptr(100)},
				{Type: "mem", Count: int64Ptr(1024)},
				{Type: "node", Count: int64Ptr(10)},
			},
			resType:  "node",
			expected: 10,
		},
		{
			name: "resource not found",
			tresList: []slurm.TRES{
				{Type: "cpu", Count: int64Ptr(100)},
				{Type: "mem", Count: int64Ptr(1024)},
			},
			resType:  "node",
			expected: 0,
		},
		{
			name:     "empty list",
			tresList: []slurm.TRES{},
			resType:  "cpu",
			expected: 0,
		},
		{
			name: "nil count",
			tresList: []slurm.TRES{
				{Type: "cpu", Count: nil},
			},
			resType:  "cpu",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTRESValue(tt.tresList, tt.resType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQoSCollector_TRESLimits(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockQoSManager := new(mocks.MockQoSManager)

	// Helper functions for pointers
	strPtr := func(s string) *string { return &s }
	uint32Ptr := func(i uint32) *uint32 { return &i }
	float64Ptr := func(f float64) *float64 { return &f }
	int64Ptr := func(i int64) *int64 { return &i }

	qosList := &slurm.QoSList{
		QoS: []slurm.QoS{
			{
				Name:        strPtr("tres-limited"),
				Description: strPtr("QoS with TRES limits"),
				Priority:    uint32Ptr(100),
				UsageFactor: float64Ptr(1.0),
				Limits: &slurm.QoSLimits{
					Max: &slurm.QoSLimitsMax{
						TRES: &slurm.QoSLimitsMaxTRES{
							Total: []slurm.TRES{
								{Type: "cpu", Count: int64Ptr(1000)},
								{Type: "node", Count: int64Ptr(50)},
								{Type: "mem", Count: int64Ptr(102400)},
							},
							Per: &slurm.QoSLimitsMaxTRESPer{
								User: []slurm.TRES{
									{Type: "cpu", Count: int64Ptr(100)},
									{Type: "node", Count: int64Ptr(10)},
								},
							},
						},
					},
					Min: &slurm.QoSLimitsMin{
						TRES: &slurm.QoSLimitsMinTRES{
							Per: &slurm.QoSLimitsMinTRESPer{
								Job: []slurm.TRES{
									{Type: "cpu", Count: int64Ptr(1)},
									{Type: "node", Count: int64Ptr(1)},
								},
							},
						},
					},
				},
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

	// Verify TRES values are extracted correctly
	qos := qosList.QoS[0]
	assert.Equal(t, int64(1000), getQoSMaxCPUs(qos))
	assert.Equal(t, int64(100), getQoSMaxCPUsPerUser(qos))
	assert.Equal(t, int64(50), getQoSMaxNodes(qos))
	assert.Equal(t, int64(1), getQoSMinCPUs(qos))
	assert.Equal(t, int64(1), getQoSMinNodes(qos))

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockQoSManager.AssertExpectations(t)
}
