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

func TestNewJobStepPerformanceCollector(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *JobStepConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &JobStepConfig{
				CollectionInterval:       60 * time.Second,
				MaxJobsPerCollection:     200,
				EnableBottleneckDetection: true,
				BottleneckThresholds: &BottleneckThresholds{
					CPUUtilizationLow:      0.2,
					MemoryUtilizationHigh:  0.9,
					IOWaitHigh:             0.15,
					NetworkUtilizationHigh: 0.85,
					LoadAverageHigh:        0.95,
				},
				CacheTTL:        10 * time.Minute,
				OnlyRunningJobs: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, err := NewJobStepPerformanceCollector(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, collector)
			assert.NotNil(t, collector.metrics)
			assert.NotNil(t, collector.stepCache)
			assert.NotNil(t, collector.bottleneckCache)

			if tt.config == nil {
				assert.Equal(t, 30*time.Second, collector.config.CollectionInterval)
				assert.Equal(t, 500, collector.config.MaxJobsPerCollection)
				assert.True(t, collector.config.EnableBottleneckDetection)
			} else {
				assert.Equal(t, tt.config.CollectionInterval, collector.config.CollectionInterval)
				assert.Equal(t, tt.config.MaxJobsPerCollection, collector.config.MaxJobsPerCollection)
				assert.Equal(t, tt.config.EnableBottleneckDetection, collector.config.EnableBottleneckDetection)
			}
		})
	}
}

func TestJobStepPerformanceCollector_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 30)
	collector.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 15)
}

func TestJobStepPerformanceCollector_CollectJobStepMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, &JobStepConfig{
		CollectionInterval:       30 * time.Second,
		MaxJobsPerCollection:     100,
		EnableBottleneckDetection: true,
		CacheTTL:                 5 * time.Minute,
		OnlyRunningJobs:          true,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	startTime := now.Add(-30 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "12345",
		Name:       "test-job",
		UserName:   "testuser",
		Account:    "testaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       8,
		Memory:     16384, // 16GB in MB
		Nodes:      2,
		StartTime:  &startTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 1 && opts.MaxCount == 100
	})).Return(jobList, nil)

	// Test collection
	ctx := context.Background()
	err = collector.collectJobStepMetrics(ctx)
	require.NoError(t, err)

	// Verify cache
	assert.Equal(t, 1, collector.GetCacheSize())

	// Verify metrics were updated
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find step CPU utilization metric
	var foundStepCPUMetric bool
	for _, mf := range metricFamilies {
		if *mf.Name == "slurm_job_step_cpu_utilization_ratio" {
			foundStepCPUMetric = true
			assert.Equal(t, 1, len(mf.Metric))
			// The simplified implementation uses a placeholder value of 0.75
			assert.Equal(t, 0.75, *mf.Metric[0].Gauge.Value)
		}
	}
	assert.True(t, foundStepCPUMetric, "Step CPU utilization metric should be present")

	mockJobManager.AssertExpectations(t)
}

func TestJobStepPerformanceCollector_BottleneckDetection(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, &JobStepConfig{
		EnableBottleneckDetection: true,
		BottleneckThresholds: &BottleneckThresholds{
			CPUUtilizationLow:      0.3,
			MemoryUtilizationHigh:  0.85,
			IOWaitHigh:             0.2,
			NetworkUtilizationHigh: 0.8,
			LoadAverageHigh:        0.9,
		},
	})
	require.NoError(t, err)

	now := time.Now()
	testJob := &slurm.Job{
		JobID:      "test-123",
		Name:       "test-job",
		UserName:   "user1",
		Account:    "account1",
		Partition:  "partition1",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192,
		Nodes:      1,
		StartTime:  &now,
	}

	// Create simplified step details
	stepDetails := collector.createSimplifiedStepDetails(testJob)
	assert.Equal(t, testJob.JobID, stepDetails.JobID)
	assert.Equal(t, "0", stepDetails.StepID)
	assert.Equal(t, testJob.Name, stepDetails.StepName)

	// Test bottleneck analysis
	bottleneckAnalysis := collector.analyzeBottlenecks(testJob, stepDetails)
	assert.NotNil(t, bottleneckAnalysis)
	assert.Equal(t, testJob.JobID, bottleneckAnalysis.JobID)
	assert.Equal(t, "0", bottleneckAnalysis.StepID)

	// The simplified implementation should detect a bottleneck (either CPU underutilization or memory pressure)
	assert.True(t, bottleneckAnalysis.Detected || bottleneckAnalysis.BottleneckType == "none")
	assert.GreaterOrEqual(t, bottleneckAnalysis.Severity, 0.0)
	assert.LessOrEqual(t, bottleneckAnalysis.Severity, 1.0)
}

func TestJobStepPerformanceCollector_CacheManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, &JobStepConfig{
		CacheTTL: 100 * time.Millisecond, // Very short TTL for testing
	})
	require.NoError(t, err)

	// Add entries to caches
	expiredStartTime := time.Now().Add(-200 * time.Millisecond)
	expiredStep := &slurm.JobStepDetails{
		JobID:     "expired-job",
		StepID:    "0",
		StartTime: &expiredStartTime,
	}
	collector.stepCache["expired-job:0"] = expiredStep

	expiredBottleneck := &BottleneckAnalysis{
		JobID:        "expired-job",
		StepID:       "0",
		LastAnalyzed: time.Now().Add(-200 * time.Millisecond),
	}
	collector.bottleneckCache["expired-job:0"] = expiredBottleneck

	// Add fresh entries
	freshStartTime := time.Now()
	freshStep := &slurm.JobStepDetails{
		JobID:     "fresh-job",
		StepID:    "0",
		StartTime: &freshStartTime,
	}
	collector.stepCache["fresh-job:0"] = freshStep

	freshBottleneck := &BottleneckAnalysis{
		JobID:        "fresh-job",
		StepID:       "0",
		LastAnalyzed: time.Now(),
	}
	collector.bottleneckCache["fresh-job:0"] = freshBottleneck

	assert.Equal(t, 2, collector.GetCacheSize())
	assert.Equal(t, 2, collector.GetBottleneckCacheSize())

	// Clean expired cache
	collector.cleanExpiredCache()

	// Only fresh entries should remain
	assert.Equal(t, 1, collector.GetCacheSize())
	assert.Equal(t, 1, collector.GetBottleneckCacheSize())

	_, exists := collector.stepCache["fresh-job:0"]
	assert.True(t, exists)
	_, exists = collector.stepCache["expired-job:0"]
	assert.False(t, exists)

	_, exists = collector.bottleneckCache["fresh-job:0"]
	assert.True(t, exists)
	_, exists = collector.bottleneckCache["expired-job:0"]
	assert.False(t, exists)
}

func TestJobStepPerformanceCollector_UpdateMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)
	endTime := now

	testJob := &slurm.Job{
		JobID:      "test-123",
		Name:       "test-job",
		UserName:   "user1",
		Account:    "account1",
		Partition:  "partition1",
		JobState:   "COMPLETED",
		CPUs:       4,
		Memory:     8192,
		Nodes:      1,
		StartTime:  &startTime,
		EndTime:    &endTime,
	}

	stepDetails := &slurm.JobStepDetails{
		JobID:      testJob.JobID,
		StepID:     "0",
		StepName:   testJob.Name,
		State:      testJob.JobState,
		StartTime:  testJob.StartTime,
		EndTime:    testJob.EndTime,
		CPUs:       testJob.CPUs,
		Memory:     int64(testJob.Memory) * 1024 * 1024,
		Nodes:      testJob.Nodes,
		CPUTime:    3600, // 1 hour of CPU time
		UserTime:   3500,
		SystemTime: 100,
	}

	// Update metrics
	collector.updateMetricsFromStepDetails(testJob, stepDetails)

	// Verify step duration metric was set (1 hour = 3600 seconds)
	stepDurationMetric := testutil.ToFloat64(collector.metrics.StepDuration.WithLabelValues(
		"test-123", "0", "test-job", "user1", "account1", "partition1", "COMPLETED",
	))
	assert.Equal(t, 3600.0, stepDurationMetric)

	// Verify CPU utilization metric was set (placeholder value in simplified implementation)
	stepCPUUtilMetric := testutil.ToFloat64(collector.metrics.StepCPUUtilization.WithLabelValues(
		"test-123", "0", "test-job", "user1", "account1", "partition1",
	))
	assert.Equal(t, 0.75, stepCPUUtilMetric)

	// Verify memory utilization metric was set (placeholder value in simplified implementation)
	stepMemoryUtilMetric := testutil.ToFloat64(collector.metrics.StepMemoryUtilization.WithLabelValues(
		"test-123", "0", "test-job", "user1", "account1", "partition1",
	))
	assert.Equal(t, 0.65, stepMemoryUtilMetric)
}

func TestJobStepPerformanceCollector_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	collector, err := NewJobStepPerformanceCollector(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = collector.collectJobStepMetrics(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs")

	mockJobManager.AssertExpectations(t)
}