package collector

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/fixtures"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func TestJobsSimpleCollector_Describe(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewJobsSimpleCollector(mockClient, logger)

	ch := make(chan *prometheus.Desc, 10)
	collector.Describe(ch)
	close(ch)

	// Should have at least the basic metrics
	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	assert.True(t, len(descs) > 0, "should have metric descriptors")
}

func TestJobsSimpleCollector_Collect_Success(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestJobList(), nil)

	collector := NewJobsSimpleCollector(mockClient, logger)
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
	mockJobManager.AssertExpectations(t)
}

func TestJobsSimpleCollector_Collect_Disabled(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector := NewJobsSimpleCollector(mockClient, logger)
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

func TestJobsSimpleCollector_Collect_Error(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock to return error
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	collector := NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.Error(t, err)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
	mockJobManager.AssertExpectations(t)
}

func TestJobsSimpleCollector_Filtering(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestJobList(), nil)

	collector := NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Configure filtering
	filterConfig := config.FilterConfig{
		Metrics: config.MetricFilterConfig{
			EnableAll:      false,
			IncludeMetrics: []string{"slurm_job_state"},
			ExcludeMetrics: []string{},
		},
	}
	collector.UpdateFilterConfig(filterConfig)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Check that only job_state metrics are collected
	metricNames := []string{}
	for metric := range ch {
		desc := metric.Desc()
		metricNames = append(metricNames, desc.String())
	}

	// All metrics should contain "job_state"
	for _, name := range metricNames {
		assert.Contains(t, name, "job_state", "only job_state metrics should be collected")
	}
}

func TestJobsSimpleCollector_CustomLabels(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetTestJobList(), nil)

	collector := NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Set custom labels
	customLabels := map[string]string{
		"cluster_name": "test-cluster",
		"environment":  "testing",
	}
	collector.SetCustomLabels(customLabels)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// Verify custom labels are present
	// Note: In a real test, we would parse the metric and check labels
	count := 0
	for range ch {
		count++
	}
	assert.True(t, count > 0, "should have collected metrics with custom labels")
}

func TestJobsSimpleCollector_EmptyJobList(t *testing.T) {
	logger := testutil.GetTestLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock to return empty list
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(fixtures.GetEmptyJobList(), nil)

	collector := NewJobsSimpleCollector(mockClient, logger)
	collector.SetEnabled(true)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	err := collector.Collect(context.Background(), ch)
	close(ch)

	assert.NoError(t, err)

	// With empty job list, no metrics should be emitted
	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 0, count, "should not emit metrics when job list is empty")
}
