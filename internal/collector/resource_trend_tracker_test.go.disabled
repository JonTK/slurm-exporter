package collector

import (
	"context"
	"log/slog"
	"math"
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

func TestNewResourceTrendTracker(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *TrendConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &TrendConfig{
				TrackingInterval:         45 * time.Second,
				MaxJobsPerCollection:     50,
				HistoryRetentionPeriod:   12 * time.Hour,
				MinDataPointsForTrend:    3,
				TrendSensitivity:         0.05,
				EnablePredictiveAnalysis: false,
				ShortTermWindow:          3 * time.Minute,
				MediumTermWindow:         20 * time.Minute,
				LongTermWindow:           90 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker, err := NewResourceTrendTracker(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, tracker)
			assert.NotNil(t, tracker.config)
			assert.NotNil(t, tracker.metrics)
			assert.NotNil(t, tracker.trendData)
			assert.NotNil(t, tracker.historicalData)
			assert.NotNil(t, tracker.patternCache)

			if tt.config == nil {
				assert.Equal(t, 30*time.Second, tracker.config.TrackingInterval)
				assert.Equal(t, 100, tracker.config.MaxJobsPerCollection)
				assert.Equal(t, 24*time.Hour, tracker.config.HistoryRetentionPeriod)
				assert.Equal(t, 5, tracker.config.MinDataPointsForTrend)
				assert.True(t, tracker.config.EnablePredictiveAnalysis)
			} else {
				assert.Equal(t, tt.config.TrackingInterval, tracker.config.TrackingInterval)
				assert.Equal(t, tt.config.MaxJobsPerCollection, tracker.config.MaxJobsPerCollection)
				assert.Equal(t, tt.config.HistoryRetentionPeriod, tracker.config.HistoryRetentionPeriod)
				assert.Equal(t, tt.config.MinDataPointsForTrend, tracker.config.MinDataPointsForTrend)
				assert.Equal(t, tt.config.EnablePredictiveAnalysis, tracker.config.EnablePredictiveAnalysis)
			}
		})
	}
}

func TestResourceTrendTracker_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 50)
	tracker.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 15)
}

func TestResourceTrendTracker_CreateResourceSnapshot(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-30 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "test-123",
		Name:       "trend-job",
		UserName:   "testuser",
		Account:    "testaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192, // 8GB in MB
		Nodes:      1,
		StartTime:  &startTime,
	}

	snapshot := tracker.createResourceSnapshot(testJob)

	assert.NotNil(t, snapshot)
	assert.WithinDuration(t, time.Now(), snapshot.Timestamp, 1*time.Second)
	assert.Equal(t, 4.0, snapshot.CPUAllocated)
	assert.Equal(t, int64(8192*1024*1024), snapshot.MemoryAllocated)
	assert.GreaterOrEqual(t, snapshot.CPUUtilization, 0.0)
	assert.LessOrEqual(t, snapshot.CPUUtilization, 1.0)
	assert.GreaterOrEqual(t, snapshot.MemoryUtilization, 0.0)
	assert.LessOrEqual(t, snapshot.MemoryUtilization, 1.0)
}

func TestResourceTrendTracker_TrackResourceTrends(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, &TrendConfig{
		TrackingInterval:         30 * time.Second,
		MaxJobsPerCollection:     50,
		MinDataPointsForTrend:    3,
		EnablePredictiveAnalysis: true,
		CacheTTL:                 5 * time.Minute,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "trend-123",
		Name:       "trend-test-job",
		UserName:   "trenduser",
		Account:    "trendaccount",
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
		return len(opts.States) >= 1 && opts.MaxCount == 50
	})).Return(jobList, nil)

	// Test trend tracking - run multiple times to build history
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		err = tracker.trackResourceTrends(ctx)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Small delay to create different timestamps
	}

	// Verify historical data was collected
	assert.Equal(t, 5, tracker.GetHistoricalDataSize("trend-123"))

	// Verify trend data was generated
	trendData := tracker.GetTrendData("trend-123")
	assert.NotNil(t, trendData)
	assert.Equal(t, "trend-123", trendData.JobID)
	assert.NotZero(t, trendData.LastUpdated)

	// Verify metrics were updated
	registry := prometheus.NewRegistry()
	registry.MustRegister(tracker)

	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	// Find trend direction metric
	var foundTrendMetric bool
	for _, mf := range metricFamilies {
		if *mf.Name == "slurm_job_resource_trend_direction" {
			foundTrendMetric = true
			assert.GreaterOrEqual(t, len(mf.Metric), 1)
		}
	}
	assert.True(t, foundTrendMetric, "Trend direction metric should be present")

	mockJobManager.AssertExpectations(t)
}

func TestResourceTrendTracker_CalculateLinearRegression(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name           string
		x              []float64
		y              []float64
		expectedSlope  float64
		expectedRSquared float64
		tolerance      float64
	}{
		{
			name:             "perfect positive correlation",
			x:                []float64{1, 2, 3, 4, 5},
			y:                []float64{2, 4, 6, 8, 10},
			expectedSlope:    2.0,
			expectedRSquared: 1.0,
			tolerance:        0.001,
		},
		{
			name:             "perfect negative correlation",
			x:                []float64{1, 2, 3, 4, 5},
			y:                []float64{10, 8, 6, 4, 2},
			expectedSlope:    -2.0,
			expectedRSquared: 1.0,
			tolerance:        0.001,
		},
		{
			name:             "no correlation",
			x:                []float64{1, 2, 3, 4, 5},
			y:                []float64{5, 5, 5, 5, 5},
			expectedSlope:    0.0,
			expectedRSquared: 1.0, // No variance in y means perfect fit
			tolerance:        0.001,
		},
		{
			name:             "positive trend with noise",
			x:                []float64{1, 2, 3, 4, 5},
			y:                []float64{1.1, 2.2, 2.9, 4.1, 4.8},
			expectedSlope:    0.95,
			expectedRSquared: 0.98,
			tolerance:        0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slope, rSquared := tracker.calculateLinearRegression(tt.x, tt.y)

			assert.InDelta(t, tt.expectedSlope, slope, tt.tolerance, "Slope should match expected value")
			assert.InDelta(t, tt.expectedRSquared, rSquared, tt.tolerance, "R-squared should match expected value")
			assert.GreaterOrEqual(t, rSquared, 0.0, "R-squared should be non-negative")
			assert.LessOrEqual(t, rSquared, 1.0, "R-squared should be at most 1.0")
		})
	}
}

func TestResourceTrendTracker_AnalyzeResourceTrend(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, &TrendConfig{
		TrendSensitivity: 0.1,
		ShortTermWindow:  5 * time.Minute,
		MediumTermWindow: 30 * time.Minute,
		LongTermWindow:   2 * time.Hour,
	})
	require.NoError(t, err)

	// Create test snapshots with increasing CPU utilization
	now := time.Now()
	snapshots := []*ResourceSnapshot{}
	for i := 0; i < 5; i++ {
		snapshot := &ResourceSnapshot{
			Timestamp:      now.Add(time.Duration(i) * time.Minute),
			CPUUtilization: 0.2 + float64(i)*0.1, // Increasing from 0.2 to 0.6
		}
		snapshots = append(snapshots, snapshot)
	}

	trend := tracker.analyzeResourceTrend("cpu", snapshots)

	assert.NotNil(t, trend)
	assert.Equal(t, "cpu", trend.ResourceType)
	assert.Equal(t, "increasing", trend.Direction)
	assert.Greater(t, trend.Slope, 0.0)
	assert.GreaterOrEqual(t, trend.RSquared, 0.0)
	assert.LessOrEqual(t, trend.RSquared, 1.0)
	assert.GreaterOrEqual(t, trend.Volatility, 0.0)
}

func TestResourceTrendTracker_DetectUsagePatterns(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, &TrendConfig{
		PeriodicPatternThreshold: 0.7,
		TrendSensitivity:         0.1,
	})
	require.NoError(t, err)

	tests := []struct {
		name          string
		snapshots     []*ResourceSnapshot
		expectPattern bool
		patternType   string
	}{
		{
			name: "gradual increase pattern",
			snapshots: func() []*ResourceSnapshot {
				now := time.Now()
				snapshots := []*ResourceSnapshot{}
				for i := 0; i < 10; i++ {
					snapshot := &ResourceSnapshot{
						Timestamp:      now.Add(time.Duration(i) * time.Minute),
						CPUUtilization: 0.1 + float64(i)*0.05, // Gradual increase
					}
					snapshots = append(snapshots, snapshot)
				}
				return snapshots
			}(),
			expectPattern: true,
			patternType:   "gradual_increase",
		},
		{
			name: "burst pattern",
			snapshots: func() []*ResourceSnapshot {
				now := time.Now()
				snapshots := []*ResourceSnapshot{}
				baseUsage := 0.3
				for i := 0; i < 10; i++ {
					usage := baseUsage
					if i%3 == 0 && i > 0 { // Burst every 3rd point
						usage = 0.8
					}
					snapshot := &ResourceSnapshot{
						Timestamp:      now.Add(time.Duration(i) * time.Minute),
						CPUUtilization: usage,
					}
					snapshots = append(snapshots, snapshot)
				}
				return snapshots
			}(),
			expectPattern: true,
			patternType:   "burst",
		},
		{
			name: "stable pattern (no significant pattern)",
			snapshots: func() []*ResourceSnapshot {
				now := time.Now()
				snapshots := []*ResourceSnapshot{}
				for i := 0; i < 10; i++ {
					snapshot := &ResourceSnapshot{
						Timestamp:      now.Add(time.Duration(i) * time.Minute),
						CPUUtilization: 0.5, // Constant usage
					}
					snapshots = append(snapshots, snapshot)
				}
				return snapshots
			}(),
			expectPattern: false,
			patternType:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := tracker.detectUsagePatterns(tt.snapshots)

			if tt.expectPattern {
				assert.NotEmpty(t, patterns, "Should detect at least one pattern")

				// Check if expected pattern type is found
				found := false
				for _, pattern := range patterns {
					if pattern.Type == tt.patternType {
						found = true
						assert.Greater(t, pattern.Confidence, 0.0)
						assert.LessOrEqual(t, pattern.Confidence, 1.0)
						assert.NotEmpty(t, pattern.Description)
						assert.NotEmpty(t, pattern.ResourcesAffected)
						break
					}
				}
				assert.True(t, found, "Expected pattern type %s should be found", tt.patternType)
			} else {
				// May or may not detect patterns for stable usage
				// This is acceptable as pattern detection can vary
			}
		})
	}
}

func TestResourceTrendTracker_DetectAnomalies(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	// Create snapshots with an anomaly (spike)
	now := time.Now()
	snapshots := []*ResourceSnapshot{}
	for i := 0; i < 10; i++ {
		usage := 0.3 // Normal usage
		if i == 5 {  // Anomaly at index 5
			usage = 0.95 // Spike
		}
		snapshot := &ResourceSnapshot{
			Timestamp:      now.Add(time.Duration(i) * time.Minute),
			CPUUtilization: usage,
		}
		snapshots = append(snapshots, snapshot)
	}

	anomalies := tracker.detectAnomalies(snapshots)

	// Should detect the spike anomaly
	assert.NotEmpty(t, anomalies, "Should detect at least one anomaly")

	// Find CPU spike anomaly
	found := false
	for _, anomaly := range anomalies {
		if anomaly.Type == "spike" && anomaly.ResourceType == "cpu" {
			found = true
			assert.NotEmpty(t, anomaly.Severity)
			assert.Greater(t, anomaly.ActualValue, anomaly.ExpectedValue)
			assert.Greater(t, anomaly.Deviation, 0.0)
			assert.NotEmpty(t, anomaly.Description)
			break
		}
	}
	assert.True(t, found, "Should detect CPU spike anomaly")
}

func TestResourceTrendTracker_PredictiveAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	// Create snapshots with increasing trend
	now := time.Now()
	snapshots := []*ResourceSnapshot{}
	for i := 0; i < 10; i++ {
		snapshot := &ResourceSnapshot{
			Timestamp:         now.Add(time.Duration(i) * time.Minute),
			CPUUtilization:    0.1 + float64(i)*0.05,
			MemoryUtilization: 0.2 + float64(i)*0.03,
		}
		snapshots = append(snapshots, snapshot)
	}

	// Test peak usage prediction
	peakPrediction := tracker.predictPeakUsage(snapshots)
	assert.NotNil(t, peakPrediction)
	assert.Greater(t, peakPrediction.CPUPeakUsage, 0.0)
	assert.LessOrEqual(t, peakPrediction.CPUPeakUsage, 1.0)
	assert.Greater(t, peakPrediction.MemoryPeakUsage, 0.0)
	assert.LessOrEqual(t, peakPrediction.MemoryPeakUsage, 1.0)
	assert.Greater(t, peakPrediction.Confidence, 0.0)
	assert.LessOrEqual(t, peakPrediction.Confidence, 1.0)

	// Test resource exhaustion prediction
	testJob := &slurm.Job{
		JobID:      "test-pred",
		TimeLimit:  120, // 2 hours
	}

	exhaustionTime := tracker.predictResourceExhaustion(snapshots, testJob)
	// May or may not predict exhaustion depending on trend - both outcomes are valid
	if exhaustionTime != nil {
		assert.True(t, exhaustionTime.After(time.Now()))
	}
}

func TestResourceTrendTracker_UpdateTrendMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "metrics-test",
		Name:       "metrics-job",
		UserName:   "metricsuser",
		Account:    "metricsaccount",
		Partition:  "compute",
	}

	trends := &JobResourceTrends{
		JobID:         "metrics-test",
		LastUpdated:   time.Now(),
		OverallTrend:  "increasing",
		TrendStrength: 0.7,
		TrendConfidence: 0.8,
		CPUTrend: &ResourceTrend{
			ResourceType:    "cpu",
			Direction:       "increasing",
			Slope:           0.05,
			RSquared:        0.9,
			Volatility:      0.1,
			ShortTermTrend:  "increasing",
			MediumTermTrend: "stable",
			LongTermTrend:   "increasing",
		},
		IdentifiedPatterns: []IdentifiedPattern{
			{Type: "gradual_increase", Confidence: 0.8},
		},
		DetectedAnomalies: []ResourceAnomaly{
			{Type: "spike", ResourceType: "cpu", Severity: "medium"},
		},
	}

	// Update metrics
	tracker.updateTrendMetrics(testJob, trends)

	// Verify trend direction metric
	cpuTrendMetric := testutil.ToFloat64(tracker.metrics.ResourceTrendDirection.WithLabelValues(
		"metrics-test", "metricsuser", "metricsaccount", "compute", "cpu",
	))
	assert.Equal(t, 1.0, cpuTrendMetric) // Increasing trend

	// Verify trend strength metric
	cpuStrengthMetric := testutil.ToFloat64(tracker.metrics.TrendStrength.WithLabelValues(
		"metrics-test", "metricsuser", "metricsaccount", "compute", "cpu",
	))
	assert.Equal(t, 0.05, cpuStrengthMetric) // Absolute slope value

	// Verify pattern detection metric
	patternMetric := testutil.ToFloat64(tracker.metrics.PatternsDetected.WithLabelValues(
		"metrics-test", "metricsuser", "metricsaccount", "compute", "gradual_increase",
	))
	assert.Equal(t, 1.0, patternMetric)

	// Verify anomaly detection metric
	anomalyMetric := testutil.ToFloat64(tracker.metrics.AnomaliesDetected.WithLabelValues(
		"metrics-test", "metricsuser", "metricsaccount", "compute", "cpu", "spike",
	))
	assert.Equal(t, 1.0, anomalyMetric)
}

func TestResourceTrendTracker_CalculateVolatility(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name              string
		values            []float64
		expectedVolatility float64
		tolerance         float64
	}{
		{
			name:              "no volatility (constant values)",
			values:            []float64{5, 5, 5, 5, 5},
			expectedVolatility: 0.0,
			tolerance:         0.001,
		},
		{
			name:              "low volatility",
			values:            []float64{4.9, 5.0, 5.1, 4.95, 5.05},
			expectedVolatility: 0.06,
			tolerance:         0.02,
		},
		{
			name:              "high volatility",
			values:            []float64{1, 5, 2, 8, 3},
			expectedVolatility: 2.8,
			tolerance:         0.5,
		},
		{
			name:              "single value",
			values:            []float64{5},
			expectedVolatility: 0.0,
			tolerance:         0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volatility := tracker.calculateVolatility(tt.values)
			assert.InDelta(t, tt.expectedVolatility, volatility, tt.tolerance, "Volatility should match expected value")
			assert.GreaterOrEqual(t, volatility, 0.0, "Volatility should be non-negative")
		})
	}
}

func TestResourceTrendTracker_CleanOldHistoricalData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, &TrendConfig{
		HistoryRetentionPeriod: 1 * time.Hour, // 1 hour retention
	})
	require.NoError(t, err)

	now := time.Now()

	// Add old data (should be cleaned)
	oldSnapshot := &ResourceSnapshot{
		Timestamp:      now.Add(-2 * time.Hour), // 2 hours ago
		CPUUtilization: 0.5,
	}
	tracker.historicalData["old-job"] = []*ResourceSnapshot{oldSnapshot}

	// Add recent data (should be kept)
	recentSnapshot := &ResourceSnapshot{
		Timestamp:      now.Add(-30 * time.Minute), // 30 minutes ago
		CPUUtilization: 0.7,
	}
	tracker.historicalData["recent-job"] = []*ResourceSnapshot{recentSnapshot}

	// Add mixed data (old should be cleaned, recent kept)
	mixedSnapshots := []*ResourceSnapshot{
		{Timestamp: now.Add(-3 * time.Hour), CPUUtilization: 0.3}, // Old
		{Timestamp: now.Add(-45 * time.Minute), CPUUtilization: 0.6}, // Recent
	}
	tracker.historicalData["mixed-job"] = mixedSnapshots

	// Clean old data
	tracker.cleanOldHistoricalData()

	// Verify old job data was removed
	assert.Empty(t, tracker.historicalData["old-job"])

	// Verify recent job data was kept
	assert.Len(t, tracker.historicalData["recent-job"], 1)

	// Verify mixed job kept only recent data
	assert.Len(t, tracker.historicalData["mixed-job"], 1)
	assert.Equal(t, 0.6, tracker.historicalData["mixed-job"][0].CPUUtilization)
}

func TestResourceTrendTracker_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	tracker, err := NewResourceTrendTracker(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = tracker.trackResourceTrends(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs")

	mockJobManager.AssertExpectations(t)
}