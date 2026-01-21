package collector

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockJobManager is a mock implementation of the JobManager interface
type MockJobManager struct {
	mock.Mock
}

func (m *MockJobManager) List(ctx context.Context, opts *slurm.ListJobsOptions) (*slurm.JobList, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(*slurm.JobList), args.Error(1)
}

func (m *MockJobManager) Get(ctx context.Context, jobID string) (*slurm.Job, error) {
	args := m.Called(ctx, jobID)
	return args.Get(0).(*slurm.Job), args.Error(1)
}

// MockSlurmClient is a mock implementation of the SLURM client
type MockSlurmClient struct {
	mock.Mock
	jobManager *MockJobManager
}

func (m *MockSlurmClient) Jobs() slurm.JobManager {
	return m.jobManager
}

func (m *MockSlurmClient) Nodes() slurm.NodeManager {
	args := m.Called()
	return args.Get(0).(slurm.NodeManager)
}

func (m *MockSlurmClient) Partitions() slurm.PartitionManager {
	args := m.Called()
	return args.Get(0).(slurm.PartitionManager)
}

func (m *MockSlurmClient) Info() slurm.InfoManager {
	args := m.Called()
	return args.Get(0).(slurm.InfoManager)
}

func (m *MockSlurmClient) Accounts() slurm.AccountManager {
	args := m.Called()
	return args.Get(0).(slurm.AccountManager)
}

func (m *MockSlurmClient) Users() slurm.UserManager {
	args := m.Called()
	return args.Get(0).(slurm.UserManager)
}

func (m *MockSlurmClient) QoS() slurm.QoSManager {
	args := m.Called()
	return args.Get(0).(slurm.QoSManager)
}

func (m *MockSlurmClient) Reservations() slurm.ReservationManager {
	args := m.Called()
	return args.Get(0).(slurm.ReservationManager)
}

func TestNewSimplifiedJobPerformanceCollector(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *JobPerformanceConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &JobPerformanceConfig{
				CollectionInterval:   60 * time.Second,
				MaxJobsPerCollection: 500,
				CacheTTL:             10 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, collector)
			assert.NotNil(t, collector.metrics)
			assert.NotNil(t, collector.jobCache)

			if tt.config == nil {
				assert.Equal(t, 30*time.Second, collector.config.CollectionInterval)
				assert.Equal(t, 1000, collector.config.MaxJobsPerCollection)
			} else {
				assert.Equal(t, tt.config.CollectionInterval, collector.config.CollectionInterval)
				assert.Equal(t, tt.config.MaxJobsPerCollection, collector.config.MaxJobsPerCollection)
			}
		})
	}
}

func TestSimplifiedJobPerformanceCollector_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 20)
	collector.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 10)
}

func TestSimplifiedJobPerformanceCollector_CollectJobMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, &JobPerformanceConfig{
		CollectionInterval:   30 * time.Second,
		MaxJobsPerCollection: 100,
		CacheTTL:             5 * time.Minute,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	submitTime := now.Add(-2 * time.Hour)
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "12345",
		Name:       "test-job",
		UserName:   "testuser",
		Account:    "testaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192, // 8GB in MB
		Nodes:      1,
		SubmitTime: &submitTime,
		StartTime:  &startTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 2 && opts.MaxCount == 100
	})).Return(jobList, nil)

	// Test collection
	ctx := context.Background()
	err = collector.collectJobMetrics(ctx)
	require.NoError(t, err)

	// Verify cache
	assert.Equal(t, 1, collector.GetCacheSize())

	// Verify metrics were updated
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find CPU allocated metric
	var foundCPUMetric bool
	for _, mf := range metricFamilies {
		if *mf.Name == "slurm_job_cpus_allocated_total" {
			foundCPUMetric = true
			assert.Equal(t, 1, len(mf.Metric))
			assert.Equal(t, 4.0, *mf.Metric[0].Gauge.Value)
		}
	}
	assert.True(t, foundCPUMetric, "CPU allocated metric should be present")

	mockJobManager.AssertExpectations(t)
}

func TestSimplifiedJobPerformanceCollector_CacheExpiration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, &JobPerformanceConfig{
		CacheTTL: 100 * time.Millisecond, // Very short TTL for testing
	})
	require.NoError(t, err)

	// Add an entry to cache
	expiredSubmitTime := time.Now().Add(-200 * time.Millisecond) // Already expired
	expiredJob := &slurm.Job{
		JobID:      "expired-job",
		SubmitTime: &expiredSubmitTime,
	}
	collector.jobCache["expired-job"] = expiredJob

	// Add a fresh entry
	freshSubmitTime := time.Now()
	freshJob := &slurm.Job{
		JobID:      "fresh-job",
		SubmitTime: &freshSubmitTime,
	}
	collector.jobCache["fresh-job"] = freshJob

	assert.Equal(t, 2, collector.GetCacheSize())

	// Clean expired cache
	collector.cleanExpiredCache()

	// Only fresh entry should remain
	assert.Equal(t, 1, collector.GetCacheSize())
	_, exists := collector.jobCache["fresh-job"]
	assert.True(t, exists)
	_, exists = collector.jobCache["expired-job"]
	assert.False(t, exists)
}

func TestSimplifiedJobPerformanceCollector_UpdateMetricsFromJob(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now
	submitTime := now.Add(-2 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "test-123",
		Name:       "test-job",
		UserName:   "user1",
		Account:    "account1",
		Partition:  "partition1",
		JobState:   "COMPLETED",
		CPUs:       4,
		Memory:     8192, // 8GB in MB
		Nodes:      1,
		TimeLimit:  120, // 2 hours in minutes
		StartTime:  &startTime,
		EndTime:    &endTime,
		SubmitTime: &submitTime,
	}

	// Update metrics
	collector.updateMetricsFromJob(testJob)

	// Verify CPU allocated metric was set
	cpuAllocatedMetric := testutil.ToFloat64(collector.metrics.JobCPUAllocated.WithLabelValues(
		"test-123", "test-job", "user1", "account1", "partition1", "COMPLETED",
	))
	assert.Equal(t, 4.0, cpuAllocatedMetric)

	// Verify memory allocated metric was set (8192 MB = 8192 * 1024 * 1024 bytes)
	memoryAllocatedMetric := testutil.ToFloat64(collector.metrics.JobMemoryAllocated.WithLabelValues(
		"test-123", "test-job", "user1", "account1", "partition1", "COMPLETED",
	))
	expectedMemoryBytes := 8192.0 * 1024 * 1024
	assert.Equal(t, expectedMemoryBytes, memoryAllocatedMetric)

	// Verify duration metric was set (1 hour = 3600 seconds)
	durationMetric := testutil.ToFloat64(collector.metrics.JobDuration.WithLabelValues(
		"test-123", "test-job", "user1", "account1", "partition1", "COMPLETED",
	))
	assert.Equal(t, 3600.0, durationMetric)

	// Verify queue time metric was set (1 hour = 3600 seconds)
	queueTimeMetric := testutil.ToFloat64(collector.metrics.JobQueueTime.WithLabelValues(
		"test-123", "test-job", "user1", "account1", "partition1",
	))
	assert.Equal(t, 3600.0, queueTimeMetric)
}

func TestSimplifiedJobPerformanceCollector_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	collector, err := NewSimplifiedJobPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = collector.collectJobMetrics(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs")

	mockJobManager.AssertExpectations(t)
}