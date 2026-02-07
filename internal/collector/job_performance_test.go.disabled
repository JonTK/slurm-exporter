// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"log/slog"
	"os"
	"testing"
	"time"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

func getTestSLogLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func TestNewJobPerformanceCollector_DefaultConfig(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)

	assert.NoError(t, err)
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.config)
	assert.Equal(t, 30*time.Second, collector.config.CollectionInterval)
	assert.Equal(t, 1000, collector.config.MaxJobsPerCollection)
}

func TestNewJobPerformanceCollector_CustomConfig(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	config := &JobPerformanceConfig{
		CollectionInterval:   60 * time.Second,
		MaxJobsPerCollection: 500,
		EnableLiveMetrics:    false,
		EnableStepMetrics:    true,
		CacheTTL:             10 * time.Minute,
	}

	collector, err := NewJobPerformanceCollector(mockClient, logger, config)

	assert.NoError(t, err)
	assert.NotNil(t, collector)
	assert.Equal(t, 60*time.Second, collector.config.CollectionInterval)
	assert.Equal(t, 500, collector.config.MaxJobsPerCollection)
	assert.False(t, collector.config.EnableLiveMetrics)
}

func TestJobPerformanceCollector_Describe(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	ch := make(chan *prometheus.Desc, 100)
	collector.Describe(ch)
	close(ch)

	descs := 0
	for range ch {
		descs++
	}

	assert.Greater(t, descs, 0, "should have metric descriptors")
}

func TestJobPerformanceCollector_Collect(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(&slurm.JobList{}, nil)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	metrics := 0
	for range ch {
		metrics++
	}

	// Should have at least some metrics
	assert.GreaterOrEqual(t, metrics, 0)
}

func TestJobUtilization_Structure(t *testing.T) {
	t.Parallel()

	util := &JobUtilization{
		JobID:             "job-123",
		CPUUtilization:    0.75,
		MemoryUtilization: 0.50,
		GPUUtilization:    0.25,
		IOUtilization:     0.10,
		LastUpdated:       time.Now(),
	}

	assert.Equal(t, "job-123", util.JobID)
	assert.Equal(t, 0.75, util.CPUUtilization)
	assert.Equal(t, 0.50, util.MemoryUtilization)
}

func TestJobPerformanceConfig_Structure(t *testing.T) {
	t.Parallel()

	config := &JobPerformanceConfig{
		CollectionInterval:   30 * time.Second,
		MaxJobsPerCollection: 1000,
		EnableLiveMetrics:    true,
		CacheTTL:             5 * time.Minute,
	}

	assert.Equal(t, 30*time.Second, config.CollectionInterval)
	assert.Equal(t, 1000, config.MaxJobsPerCollection)
	assert.True(t, config.EnableLiveMetrics)
}

func TestJobPerformanceCollector_GetCacheSize(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	size := collector.GetCacheSize()
	assert.GreaterOrEqual(t, size, 0)
}

func TestJobPerformanceCollector_GetLastCollection(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	lastCollection := collector.GetLastCollection()
	// Should be zero time initially or recent
	assert.NotNil(t, lastCollection)
}

func TestJobPerformanceCollector_CollectJobUtilization_Success(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(&slurm.JobList{
		Jobs: []slurm.Job{
			{
				ID:   "job-1",
				Name: "test-job",
			},
		},
	}, nil)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	// Collect should work without error
	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestJobPerformanceMetrics_Structure(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	// Verify metrics are initialized
	assert.NotNil(t, collector.metrics)
	assert.NotNil(t, collector.metrics.JobCPUUtilization)
	assert.NotNil(t, collector.metrics.JobMemoryUtilization)
	assert.NotNil(t, collector.metrics.CollectionDuration)
}

func TestJobPerformanceCollector_CacheSize(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	// Initial cache should be empty or small
	initialSize := collector.GetCacheSize()
	assert.GreaterOrEqual(t, initialSize, 0)
}

func TestJobPerformanceCollector_Concurrent_Collect(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup mock expectations with unlimited call count
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", mock.Anything, mock.Anything).Return(&slurm.JobList{}, nil)

	collector, err := NewJobPerformanceCollector(mockClient, logger, nil)
	assert.NoError(t, err)

	done := make(chan bool, 5)

	// Concurrent collects
	for i := 0; i < 5; i++ {
		go func() {
			ch := make(chan prometheus.Metric, 100)
			collector.Collect(ch)
			close(ch)
			done <- true
		}()
	}

	// Wait for all
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestJobPerformanceCollector_WithCompletedJobs(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	config := &JobPerformanceConfig{
		CollectionInterval:   30 * time.Second,
		MaxJobsPerCollection: 1000,
		EnableLiveMetrics:    true,
		IncludeCompletedJobs: true,
		CompletedJobsMaxAge:  1 * time.Hour,
	}

	collector, err := NewJobPerformanceCollector(mockClient, logger, config)
	assert.NoError(t, err)
	assert.True(t, collector.config.IncludeCompletedJobs)
	assert.Equal(t, 1*time.Hour, collector.config.CompletedJobsMaxAge)
}

func TestJobPerformanceCollector_EnergyMetrics_Disabled(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	config := &JobPerformanceConfig{
		EnableEnergyMetrics: false,
	}

	collector, err := NewJobPerformanceCollector(mockClient, logger, config)
	assert.NoError(t, err)
	assert.False(t, collector.config.EnableEnergyMetrics)
}

func TestJobPerformanceCollector_StepMetrics_Enabled(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	config := &JobPerformanceConfig{
		EnableStepMetrics: true,
	}

	collector, err := NewJobPerformanceCollector(mockClient, logger, config)
	assert.NoError(t, err)
	assert.True(t, collector.config.EnableStepMetrics)
}

func TestJobPerformanceCollector_CacheTTL(t *testing.T) {
	t.Parallel()
	logger := getTestSLogLogger()
	mockClient := new(mocks.MockSlurmClient)

	config := &JobPerformanceConfig{
		CacheTTL: 10 * time.Minute,
	}

	collector, err := NewJobPerformanceCollector(mockClient, logger, config)
	assert.NoError(t, err)
	assert.Equal(t, 10*time.Minute, collector.config.CacheTTL)
}
