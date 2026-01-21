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

func TestNewBottleneckAnalyzer(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *BottleneckConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &BottleneckConfig{
				AnalysisInterval:                45 * time.Second,
				MaxJobsPerAnalysis:              50,
				EnablePredictiveAnalysis:        false,
				EnableRootCauseAnalysis:         false,
				CPUBottleneckThreshold:          0.95,
				MemoryBottleneckThreshold:       0.8,
				IOBottleneckThreshold:           0.15,
				SensitivityLevel:                "high",
				AnalysisConfidenceLevel:         0.8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewBottleneckAnalyzer(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, analyzer)
			assert.NotNil(t, analyzer.config)
			assert.NotNil(t, analyzer.metrics)
			assert.NotNil(t, analyzer.efficiencyCalc)
			assert.NotNil(t, analyzer.analysisCache)

			if tt.config == nil {
				assert.Equal(t, 30*time.Second, analyzer.config.AnalysisInterval)
				assert.Equal(t, 100, analyzer.config.MaxJobsPerAnalysis)
				assert.True(t, analyzer.config.EnablePredictiveAnalysis)
				assert.True(t, analyzer.config.EnableRootCauseAnalysis)
			} else {
				assert.Equal(t, tt.config.AnalysisInterval, analyzer.config.AnalysisInterval)
				assert.Equal(t, tt.config.MaxJobsPerAnalysis, analyzer.config.MaxJobsPerAnalysis)
				assert.Equal(t, tt.config.EnablePredictiveAnalysis, analyzer.config.EnablePredictiveAnalysis)
				assert.Equal(t, tt.config.EnableRootCauseAnalysis, analyzer.config.EnableRootCauseAnalysis)
			}
		})
	}
}

func TestBottleneckAnalyzer_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 50)
	analyzer.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 15)
}

func TestBottleneckAnalyzer_AnalyzeBottlenecks(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, &BottleneckConfig{
		AnalysisInterval:       30 * time.Second,
		MaxJobsPerAnalysis:     50,
		EnablePredictiveAnalysis: true,
		EnableRootCauseAnalysis: true,
		CPUBottleneckThreshold: 0.9,
		MemoryBottleneckThreshold: 0.85,
		CacheTTL:               5 * time.Minute,
	})
	require.NoError(t, err)

	// Setup mock expectations with a high-CPU usage job
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "12345",
		Name:       "cpu-intensive-job",
		UserName:   "testuser",
		Account:    "testaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       8,
		Memory:     16384, // 16GB in MB
		Nodes:      1,
		TimeLimit:  120, // 2 hours
		StartTime:  &startTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 1 && opts.MaxCount == 50
	})).Return(jobList, nil)

	// Test bottleneck analysis
	ctx := context.Background()
	err = analyzer.analyzeBottlenecks(ctx)
	require.NoError(t, err)

	// Verify cache
	assert.Equal(t, 1, analyzer.GetCacheSize())

	// Verify metrics were updated
	registry := prometheus.NewRegistry()
	registry.MustRegister(analyzer)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find bottleneck detection metric
	var foundBottleneckMetric bool
	for _, mf := range metricFamilies {
		if *mf.Name == "slurm_job_bottleneck_detected" {
			foundBottleneckMetric = true
			assert.Equal(t, 1, len(mf.Metric))
			// Should detect some kind of bottleneck or no bottleneck
			value := *mf.Metric[0].Gauge.Value
			assert.True(t, value == 0.0 || value == 1.0)
		}
	}
	assert.True(t, foundBottleneckMetric, "Bottleneck detection metric should be present")

	mockJobManager.AssertExpectations(t)
}

func TestBottleneckAnalyzer_IdentifyPrimaryBottleneck(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name           string
		data           *ResourceUtilizationData
		efficiency     *EfficiencyMetrics
		expectedType   string
		expectBottleneck bool
	}{
		{
			name: "CPU bottleneck scenario",
			data: &ResourceUtilizationData{
				CPUAllocated: 8.0,
				CPUUsed:      7.5, // 93.75% utilization - high
				WallTime:     3600.0,
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.5, // Low efficiency despite high usage
				MemoryEfficiency:  0.8,
				IOEfficiency:      0.9,
				NetworkEfficiency: 0.9,
			},
			expectedType:     "cpu",
			expectBottleneck: true,
		},
		{
			name: "Memory bottleneck scenario",
			data: &ResourceUtilizationData{
				CPUAllocated:    4.0,
				CPUUsed:         3.0,
				MemoryAllocated: 8 * 1024 * 1024 * 1024, // 8GB
				MemoryUsed:      7.5 * 1024 * 1024 * 1024, // 7.5GB (93.75% usage)
				WallTime:        3600.0,
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.8,
				MemoryEfficiency:  0.4, // Low memory efficiency
				IOEfficiency:      0.9,
				NetworkEfficiency: 0.9,
			},
			expectedType:     "memory",
			expectBottleneck: true,
		},
		{
			name: "I/O bottleneck scenario",
			data: &ResourceUtilizationData{
				CPUAllocated: 4.0,
				CPUUsed:      3.0,
				MemoryAllocated: 8 * 1024 * 1024 * 1024,
				MemoryUsed:      4 * 1024 * 1024 * 1024,
				WallTime:        3600.0,
				IOWaitTime:      900.0, // 25% I/O wait - high
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.8,
				MemoryEfficiency:  0.8,
				IOEfficiency:      0.3, // Low I/O efficiency
				NetworkEfficiency: 0.9,
			},
			expectedType:     "io",
			expectBottleneck: true,
		},
		{
			name: "No bottleneck scenario",
			data: &ResourceUtilizationData{
				CPUAllocated:    4.0,
				CPUUsed:         3.0, // 75% utilization - reasonable
				MemoryAllocated: 8 * 1024 * 1024 * 1024,
				MemoryUsed:      5 * 1024 * 1024 * 1024, // 62.5% usage - reasonable
				WallTime:        3600.0,
				IOWaitTime:      36.0, // 1% I/O wait - low
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.8,
				MemoryEfficiency:  0.8,
				IOEfficiency:      0.9,
				NetworkEfficiency: 0.9,
			},
			expectedType:     "none",
			expectBottleneck: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bottleneckType, severity, confidence := analyzer.identifyPrimaryBottleneck(tt.data, tt.efficiency)

			assert.Equal(t, tt.expectedType, bottleneckType)

			if tt.expectBottleneck {
				assert.Greater(t, severity, 0.0)
				assert.Greater(t, confidence, 0.0)
				assert.LessOrEqual(t, severity, 1.0)
				assert.LessOrEqual(t, confidence, 1.0)
			} else {
				assert.Equal(t, 0.0, severity)
			}
		})
	}
}

func TestBottleneckAnalyzer_PerformStepPerformanceAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, &BottleneckConfig{
		EnablePredictiveAnalysis: true,
		EnableRootCauseAnalysis:  true,
	})
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-30 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "test-123",
		Name:       "analysis-job",
		UserName:   "user1",
		Account:    "account1",
		Partition:  "partition1",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192, // 8GB
		Nodes:      1,
		TimeLimit:  120, // 2 hours
		StartTime:  &startTime,
	}

	analysis := analyzer.performStepPerformanceAnalysis(testJob)

	// Verify basic analysis structure
	assert.Equal(t, testJob.JobID, analysis.JobID)
	assert.Equal(t, "0", analysis.StepID)
	assert.NotZero(t, analysis.AnalysisTimestamp)

	// Verify efficiency metrics are set
	assert.GreaterOrEqual(t, analysis.CPUEfficiency, 0.0)
	assert.LessOrEqual(t, analysis.CPUEfficiency, 1.0)
	assert.GreaterOrEqual(t, analysis.MemoryEfficiency, 0.0)
	assert.LessOrEqual(t, analysis.MemoryEfficiency, 1.0)
	assert.GreaterOrEqual(t, analysis.OverallEfficiency, 0.0)
	assert.LessOrEqual(t, analysis.OverallEfficiency, 1.0)

	// Verify bottleneck analysis
	assert.Contains(t, []string{"cpu", "memory", "io", "network", "none"}, analysis.PrimaryBottleneck)
	assert.GreaterOrEqual(t, analysis.BottleneckSeverity, 0.0)
	assert.LessOrEqual(t, analysis.BottleneckSeverity, 1.0)
	assert.GreaterOrEqual(t, analysis.BottleneckConfidence, 0.0)
	assert.LessOrEqual(t, analysis.BottleneckConfidence, 1.0)

	// Verify patterns are analyzed
	assert.NotEmpty(t, analysis.CPUUtilizationPattern)
	assert.NotEmpty(t, analysis.MemoryUtilizationPattern)

	// Verify predictive analysis was performed
	assert.Greater(t, analysis.EstimatedTimeRemaining, 0.0)
	assert.NotEmpty(t, analysis.PerformanceTrend)

	// Verify recommendations are provided
	assert.NotEmpty(t, analysis.RecommendedActions)
}

func TestBottleneckAnalyzer_DetectPerformanceIssues(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name       string
		data       *ResourceUtilizationData
		efficiency *EfficiencyMetrics
		expectIssues bool
		expectedIssueTypes []string
	}{
		{
			name: "CPU underutilization issue",
			data: &ResourceUtilizationData{
				CPUAllocated: 8.0,
				CPUUsed:      1.0, // Very low utilization
				WallTime:     3600.0,
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency: 0.2, // Very low efficiency
				MemoryEfficiency: 0.8,
			},
			expectIssues: true,
			expectedIssueTypes: []string{"cpu_underutilization"},
		},
		{
			name: "Memory pressure issue",
			data: &ResourceUtilizationData{
				CPUAllocated:    4.0,
				CPUUsed:         3.0,
				MemoryAllocated: 4 * 1024 * 1024 * 1024, // 4GB
				MemoryUsed:      3.8 * 1024 * 1024 * 1024, // 95% usage
				WallTime:        3600.0,
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:    0.8,
				MemoryEfficiency: 0.8,
			},
			expectIssues: true,
			expectedIssueTypes: []string{"memory_pressure"},
		},
		{
			name: "No performance issues",
			data: &ResourceUtilizationData{
				CPUAllocated:    4.0,
				CPUUsed:         3.0,
				MemoryAllocated: 8 * 1024 * 1024 * 1024,
				MemoryUsed:      5 * 1024 * 1024 * 1024, // 62.5% usage
				WallTime:        3600.0,
			},
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:    0.8,
				MemoryEfficiency: 0.8,
			},
			expectIssues: false,
			expectedIssueTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := analyzer.detectPerformanceIssues(tt.data, tt.efficiency)

			if tt.expectIssues {
				assert.NotEmpty(t, issues)

				// Check for expected issue types
				foundTypes := make(map[string]bool)
				for _, issue := range issues {
					foundTypes[issue.Type] = true

					// Verify issue structure
					assert.NotEmpty(t, issue.Type)
					assert.NotEmpty(t, issue.Severity)
					assert.NotEmpty(t, issue.Description)
					assert.NotEmpty(t, issue.Impact)
					assert.GreaterOrEqual(t, issue.Frequency, 0.0)
					assert.LessOrEqual(t, issue.Frequency, 1.0)
				}

				for _, expectedType := range tt.expectedIssueTypes {
					assert.True(t, foundTypes[expectedType], "Expected issue type %s not found", expectedType)
				}
			} else {
				assert.Empty(t, issues)
			}
		})
	}
}

func TestBottleneckAnalyzer_UpdateAnalysisMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "test-456",
		Name:       "metrics-job",
		UserName:   "user2",
		Account:    "account2",
		Partition:  "partition2",
		JobState:   "RUNNING",
	}

	now := time.Now()
	predictedCompletion := now.Add(1 * time.Hour)

	analysis := &StepPerformanceAnalysis{
		JobID:                  "test-456",
		StepID:                 "0",
		AnalysisTimestamp:      now,
		PrimaryBottleneck:      "cpu",
		BottleneckSeverity:     0.7,
		BottleneckConfidence:   0.8,
		OverallEfficiency:      0.6,
		EstimatedTimeRemaining: 3600.0, // 1 hour
		PredictedCompletion:    &predictedCompletion,
		PerformanceTrend:       "stable",
		PerformanceIssues: []PerformanceIssue{
			{Type: "cpu_underutilization", Severity: "medium"},
		},
		OptimizationOpportunities: []OptimizationOpportunity{
			{Type: "resource_reduction"},
		},
		RootCauses: []RootCause{
			{Cause: "Over-allocation", Confidence: 0.9},
		},
	}

	// Update metrics
	analyzer.updateAnalysisMetrics(testJob, analysis)

	// Verify bottleneck detection metric
	bottleneckMetric := testutil.ToFloat64(analyzer.metrics.BottleneckDetected.WithLabelValues(
		"test-456", "0", "user2", "account2", "partition2",
	))
	assert.Equal(t, 1.0, bottleneckMetric)

	// Verify bottleneck severity metric
	severityMetric := testutil.ToFloat64(analyzer.metrics.BottleneckSeverity.WithLabelValues(
		"test-456", "0", "user2", "account2", "partition2", "cpu",
	))
	assert.Equal(t, 0.7, severityMetric)

	// Verify efficiency score metric
	efficiencyMetric := testutil.ToFloat64(analyzer.metrics.PerformanceEfficiencyScore.WithLabelValues(
		"test-456", "0", "user2", "account2", "partition2",
	))
	assert.Equal(t, 0.6, efficiencyMetric)

	// Verify estimated time remaining
	timeMetric := testutil.ToFloat64(analyzer.metrics.EstimatedTimeRemaining.WithLabelValues(
		"test-456", "0", "user2", "account2", "partition2",
	))
	assert.Equal(t, 3600.0, timeMetric)
}

func TestBottleneckAnalyzer_CacheManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, &BottleneckConfig{
		CacheTTL: 100 * time.Millisecond, // Very short TTL for testing
	})
	require.NoError(t, err)

	// Add expired analysis to cache
	expiredAnalysis := &StepPerformanceAnalysis{
		JobID:             "expired-job",
		StepID:            "0",
		AnalysisTimestamp: time.Now().Add(-200 * time.Millisecond),
	}
	analyzer.analysisCache["expired-job:0"] = expiredAnalysis

	// Add fresh analysis to cache
	freshAnalysis := &StepPerformanceAnalysis{
		JobID:             "fresh-job",
		StepID:            "0",
		AnalysisTimestamp: time.Now(),
	}
	analyzer.analysisCache["fresh-job:0"] = freshAnalysis

	assert.Equal(t, 2, analyzer.GetCacheSize())

	// Clean expired cache
	analyzer.cleanExpiredCache()

	// Only fresh entry should remain
	assert.Equal(t, 1, analyzer.GetCacheSize())

	_, exists := analyzer.analysisCache["fresh-job:0"]
	assert.True(t, exists)
	_, exists = analyzer.analysisCache["expired-job:0"]
	assert.False(t, exists)
}

func TestBottleneckAnalyzer_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	analyzer, err := NewBottleneckAnalyzer(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = analyzer.analyzeBottlenecks(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs")

	mockJobManager.AssertExpectations(t)
}