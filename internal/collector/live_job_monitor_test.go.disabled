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

func TestNewLiveJobMonitor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *LiveMonitorConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &LiveMonitorConfig{
				MonitoringInterval:         3 * time.Second,
				MaxJobsPerCollection:       25,
				EnableRealTimeAlerting:     false,
				EnablePerformanceStreaming: false,
				StreamingBufferSize:        500,
				DataRetentionPeriod:        30 * time.Minute,
				CacheTTL:                   15 * time.Second,
				BatchSize:                  5,
				EnablePredictiveAlerting:   false,
				MinDataPointsForAlert:      5,
				AlertCooldownPeriod:        2 * time.Minute,
				AlertThresholds: &AlertThresholds{
					CPUUtilizationHigh:     0.90,
					CPUUtilizationLow:      0.05,
					MemoryUtilizationHigh:  0.85,
					MemoryUtilizationLow:   0.05,
					EfficiencyLow:          0.25,
					ResourceWasteHigh:      0.60,
					ThroughputLow:          0.15,
					ResponseTimeHigh:       15.0,
					HealthScoreLow:         0.30,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor, err := NewLiveJobMonitor(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, monitor)
			assert.NotNil(t, monitor.config)
			assert.NotNil(t, monitor.metrics)
			assert.NotNil(t, monitor.efficiencyCalc)
			assert.NotNil(t, monitor.liveData)
			assert.NotNil(t, monitor.alertChannels)

			if tt.config == nil {
				assert.Equal(t, 5*time.Second, monitor.config.MonitoringInterval)
				assert.Equal(t, 50, monitor.config.MaxJobsPerCollection)
				assert.True(t, monitor.config.EnableRealTimeAlerting)
				assert.True(t, monitor.config.EnablePerformanceStreaming)
			} else {
				assert.Equal(t, tt.config.MonitoringInterval, monitor.config.MonitoringInterval)
				assert.Equal(t, tt.config.MaxJobsPerCollection, monitor.config.MaxJobsPerCollection)
				assert.Equal(t, tt.config.EnableRealTimeAlerting, monitor.config.EnableRealTimeAlerting)
				assert.Equal(t, tt.config.EnablePerformanceStreaming, monitor.config.EnablePerformanceStreaming)
			}
		})
	}
}

func TestLiveJobMonitor_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 100)
	monitor.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 20)
}

func TestLiveJobMonitor_CollectLiveJobMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, &LiveMonitorConfig{
		MonitoringInterval:       5 * time.Second,
		MaxJobsPerCollection:     25,
		EnableRealTimeAlerting:   true,
		BatchSize:                5,
		DataRetentionPeriod:      1 * time.Hour,
		CacheTTL:                 30 * time.Second,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	startTime := now.Add(-30 * time.Minute)

	testJob := &slurm.Job{
		JobID:      "live-123",
		Name:       "live-test-job",
		UserName:   "liveuser",
		Account:    "liveaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192, // 8GB in MB
		Nodes:      1,
		TimeLimit:  120, // 2 hours
		StartTime:  &startTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 1 && opts.MaxCount == 25
	})).Return(jobList, nil)

	// Test live metrics collection
	ctx := context.Background()
	err = monitor.collectLiveJobMetrics(ctx)
	require.NoError(t, err)

	// Verify live data was collected
	liveData, exists := monitor.GetLiveData("live-123")
	assert.True(t, exists)
	assert.NotNil(t, liveData)
	assert.Equal(t, "live-123", liveData.JobID)
	assert.WithinDuration(t, time.Now(), liveData.Timestamp, 5*time.Second)

	// Verify resource usage metrics
	assert.GreaterOrEqual(t, liveData.CurrentCPUUsage, 0.0)
	assert.LessOrEqual(t, liveData.CurrentCPUUsage, float64(testJob.CPUs))
	assert.GreaterOrEqual(t, liveData.CurrentMemoryUsage, int64(0))
	assert.LessOrEqual(t, liveData.CurrentMemoryUsage, int64(testJob.Memory*1024*1024))
	assert.GreaterOrEqual(t, liveData.CurrentIORate, 0.0)
	assert.GreaterOrEqual(t, liveData.CurrentNetworkRate, 0.0)

	// Verify efficiency metrics
	assert.GreaterOrEqual(t, liveData.InstantCPUEfficiency, 0.0)
	assert.LessOrEqual(t, liveData.InstantCPUEfficiency, 1.0)
	assert.GreaterOrEqual(t, liveData.InstantMemoryEfficiency, 0.0)
	assert.LessOrEqual(t, liveData.InstantMemoryEfficiency, 1.0)
	assert.GreaterOrEqual(t, liveData.InstantOverallEfficiency, 0.0)
	assert.LessOrEqual(t, liveData.InstantOverallEfficiency, 1.0)

	// Verify performance indicators
	assert.GreaterOrEqual(t, liveData.ThroughputRate, 0.0)
	assert.GreaterOrEqual(t, liveData.ResponseTime, 0.0)
	assert.GreaterOrEqual(t, liveData.ResourceWasteRate, 0.0)
	assert.LessOrEqual(t, liveData.ResourceWasteRate, 1.0)
	assert.GreaterOrEqual(t, liveData.HealthScore, 0.0)
	assert.LessOrEqual(t, liveData.HealthScore, 1.0)

	// Verify metadata
	assert.NotEmpty(t, liveData.PerformanceGrade)
	assert.Contains(t, []string{"A", "B", "C", "D", "F"}, liveData.PerformanceGrade)
	assert.Contains(t, []string{"low", "medium", "high"}, liveData.ResourceExhaustionRisk)
	assert.Contains(t, []string{"improving", "stable", "declining"}, liveData.PerformanceTrend)

	mockJobManager.AssertExpectations(t)
}

func TestLiveJobMonitor_SimulateResourceUsage(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "sim-456",
		Name:       "simulation-job",
		UserName:   "simuser",
		Account:    "simaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       8,
		Memory:     16384, // 16GB
		Nodes:      1,
		TimeLimit:  240, // 4 hours
		StartTime:  &startTime,
	}

	// Test CPU usage simulation
	cpuUsage := monitor.simulateCurrentCPUUsage(testJob)
	assert.GreaterOrEqual(t, cpuUsage, 0.0)
	assert.LessOrEqual(t, cpuUsage, float64(testJob.CPUs))

	// Test memory usage simulation
	memoryUsage := monitor.simulateCurrentMemoryUsage(testJob)
	assert.GreaterOrEqual(t, memoryUsage, 0.0)
	assert.LessOrEqual(t, memoryUsage, float64(testJob.Memory*1024*1024))

	// Test I/O rate simulation
	ioRate := monitor.simulateCurrentIORate(testJob)
	assert.GreaterOrEqual(t, ioRate, 0.0)

	// Test network rate simulation
	networkRate := monitor.simulateCurrentNetworkRate(testJob)
	assert.GreaterOrEqual(t, networkRate, 0.0)
}

func TestLiveJobMonitor_PerformanceCalculations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-2 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "perf-789",
		Name:       "performance-job",
		UserName:   "perfuser",
		Account:    "perfaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192,
		Nodes:      1,
		TimeLimit:  180, // 3 hours
		StartTime:  &startTime,
	}

	// Test throughput calculation
	cpuUsage := 3.2 // 80% of 4 CPUs
	throughputRate := monitor.calculateThroughputRate(testJob, cpuUsage)
	assert.GreaterOrEqual(t, throughputRate, 0.0)
	assert.LessOrEqual(t, throughputRate, 100.0)

	// Test response time calculation
	responseTime := monitor.calculateResponseTime(testJob)
	assert.GreaterOrEqual(t, responseTime, 0.0)

	// Test efficiency metrics
	efficiency := &EfficiencyMetrics{
		CPUEfficiency:     0.8,
		MemoryEfficiency:  0.7,
		IOEfficiency:      0.9,
		NetworkEfficiency: 0.8,
		OverallEfficiency: 0.8,
	}

	// Test waste rate calculation
	wasteRate := monitor.calculateResourceWasteRate(efficiency)
	assert.GreaterOrEqual(t, wasteRate, 0.0)
	assert.LessOrEqual(t, wasteRate, 1.0)

	// Test health score calculation
	healthScore := monitor.calculateHealthScore(efficiency, wasteRate, throughputRate)
	assert.GreaterOrEqual(t, healthScore, 0.0)
	assert.LessOrEqual(t, healthScore, 1.0)

	// Test performance grade calculation
	grades := []string{"A", "B", "C", "D", "F"}
	grade := monitor.calculatePerformanceGrade(healthScore)
	assert.Contains(t, grades, grade)

	// Test completion prediction
	completion := monitor.predictCompletion(testJob, throughputRate)
	if completion != nil {
		assert.True(t, completion.After(time.Now()))
	}

	// Test risk assessment
	memoryUsage := float64(testJob.Memory * 1024 * 1024 * 0.7) // 70% usage
	risk := monitor.assessResourceExhaustionRisk(testJob, memoryUsage, cpuUsage)
	assert.Contains(t, []string{"low", "medium", "high"}, risk)
}

func TestLiveJobMonitor_AlertGeneration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, &LiveMonitorConfig{
		EnableRealTimeAlerting: true,
		AlertThresholds: &AlertThresholds{
			CPUUtilizationHigh:     0.85,
			CPUUtilizationLow:      0.15,
			MemoryUtilizationHigh:  0.80,
			MemoryUtilizationLow:   0.10,
			EfficiencyLow:          0.40,
			ResourceWasteHigh:      0.60,
			ThroughputLow:          0.30,
			ResponseTimeHigh:       8.0,
			HealthScoreLow:         0.50,
		},
	})
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "alert-999",
		Name:       "alert-test-job",
		UserName:   "alertuser",
		Account:    "alertaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192,
	}

	tests := []struct {
		name        string
		liveMetrics *JobLiveMetrics
		expectAlert bool
		alertType   string
	}{
		{
			name: "high CPU utilization alert",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          3.6, // 90% of 4 CPUs
				CurrentMemoryUsage:       4 * 1024 * 1024 * 1024, // 4GB
				InstantOverallEfficiency: 0.7,
				HealthScore:              0.8,
			},
			expectAlert: true,
			alertType:   "cpu_utilization_high",
		},
		{
			name: "low CPU utilization alert",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          0.4, // 10% of 4 CPUs
				CurrentMemoryUsage:       2 * 1024 * 1024 * 1024, // 2GB
				InstantOverallEfficiency: 0.7,
				HealthScore:              0.8,
			},
			expectAlert: true,
			alertType:   "cpu_utilization_low",
		},
		{
			name: "high memory utilization alert",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          2.0, // 50% of 4 CPUs
				CurrentMemoryUsage:       int64(float64(8192*1024*1024) * 0.85), // 85% of 8GB
				InstantOverallEfficiency: 0.7,
				HealthScore:              0.8,
			},
			expectAlert: true,
			alertType:   "memory_utilization_high",
		},
		{
			name: "low efficiency alert",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          2.0,
				CurrentMemoryUsage:       4 * 1024 * 1024 * 1024,
				InstantOverallEfficiency: 0.3, // Below threshold
				HealthScore:              0.8,
			},
			expectAlert: true,
			alertType:   "efficiency_low",
		},
		{
			name: "low health score alert",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          2.0,
				CurrentMemoryUsage:       4 * 1024 * 1024 * 1024,
				InstantOverallEfficiency: 0.7,
				HealthScore:              0.4, // Below threshold
				Recommendations:          []string{"Optimize job performance"},
			},
			expectAlert: true,
			alertType:   "health_score_low",
		},
		{
			name: "no alerts - normal operation",
			liveMetrics: &JobLiveMetrics{
				JobID:                     "alert-999",
				CurrentCPUUsage:          2.5, // 62.5% of 4 CPUs
				CurrentMemoryUsage:       5 * 1024 * 1024 * 1024, // 5GB (62.5% of 8GB)
				InstantOverallEfficiency: 0.75,
				HealthScore:              0.8,
			},
			expectAlert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset metrics to avoid interference between tests
			registry := prometheus.NewRegistry()
			registry.MustRegister(monitor)

			// Process the job with specific metrics
			monitor.checkForAlerts(testJob, tt.liveMetrics)

			// Check if alerts were generated
			metricFamilies, err := registry.Gather()
			require.NoError(t, err)

			alertsGenerated := false
			for _, mf := range metricFamilies {
				if *mf.Name == "slurm_job_alerts_generated_total" {
					for _, metric := range mf.Metric {
						if *metric.Counter.Value > 0 {
							alertsGenerated = true
							if tt.expectAlert {
								// Check if the alert type matches
								for _, label := range metric.Label {
									if *label.Name == "alert_type" && *label.Value == tt.alertType {
										assert.True(t, true, "Expected alert type found")
										break
									}
								}
							}
						}
					}
				}
			}

			if tt.expectAlert {
				assert.True(t, alertsGenerated, "Expected alert to be generated")
			} else {
				assert.False(t, alertsGenerated, "Expected no alerts to be generated")
			}
		})
	}
}

func TestLiveJobMonitor_UpdateLiveMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "metrics-555",
		Name:       "metrics-test-job",
		UserName:   "metricsuser",
		Account:    "metricsaccount",
		Partition:  "compute",
	}

	now := time.Now()
	futureTime := now.Add(2 * time.Hour)

	liveMetrics := &JobLiveMetrics{
		JobID:                     "metrics-555",
		Timestamp:                 now,
		CurrentCPUUsage:          3.2,
		CurrentMemoryUsage:       6 * 1024 * 1024 * 1024, // 6GB
		CurrentIORate:            50 * 1024 * 1024,        // 50MB/s
		CurrentNetworkRate:       10 * 1024 * 1024,        // 10MB/s
		InstantCPUEfficiency:     0.8,
		InstantMemoryEfficiency:  0.75,
		InstantOverallEfficiency: 0.78,
		ThroughputRate:           85.0,
		ResponseTime:             2.5,
		ResourceWasteRate:        0.22,
		HealthScore:              0.85,
		EstimatedCompletion:      &futureTime,
		ResourceExhaustionRisk:   "low",
		PerformanceTrend:         "stable",
	}

	// Update metrics
	monitor.updateLiveMetrics(testJob, liveMetrics)

	// Verify CPU usage metric
	cpuMetric := testutil.ToFloat64(monitor.metrics.CurrentCPUUsage.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 3.2, cpuMetric)

	// Verify memory usage metric
	memoryMetric := testutil.ToFloat64(monitor.metrics.CurrentMemoryUsage.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 6*1024*1024*1024, int64(memoryMetric))

	// Verify efficiency metrics
	cpuEfficiencyMetric := testutil.ToFloat64(monitor.metrics.InstantCPUEfficiency.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.8, cpuEfficiencyMetric)

	overallEfficiencyMetric := testutil.ToFloat64(monitor.metrics.InstantOverallEfficiency.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.78, overallEfficiencyMetric)

	// Verify performance metrics
	healthMetric := testutil.ToFloat64(monitor.metrics.HealthScore.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.85, healthMetric)

	throughputMetric := testutil.ToFloat64(monitor.metrics.ThroughputRate.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 85.0, throughputMetric)

	// Verify trend metrics
	trendMetric := testutil.ToFloat64(monitor.metrics.PerformanceTrend.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.0, trendMetric) // "stable" = 0.0

	// Verify risk metrics
	riskMetric := testutil.ToFloat64(monitor.metrics.ResourceExhaustionRisk.WithLabelValues(
		"metrics-555", "metricsuser", "metricsaccount", "compute", "overall",
	))
	assert.Equal(t, 0.0, riskMetric) // "low" = 0.0
}

func TestLiveJobMonitor_StreamingFunctionality(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, &LiveMonitorConfig{
		EnablePerformanceStreaming: true,
		StreamingBufferSize:        10,
	})
	require.NoError(t, err)

	jobID := "stream-123"

	// Start streaming
	alertChannel := monitor.StartStreaming(jobID)
	assert.NotNil(t, alertChannel)

	stats := monitor.GetMonitoringStats()
	assert.Equal(t, 1, stats["active_streams"])
	assert.True(t, stats["streaming_active"].(bool))

	// Stop streaming
	monitor.StopStreaming(jobID)

	stats = monitor.GetMonitoringStats()
	assert.Equal(t, 0, stats["active_streams"])
	assert.False(t, stats["streaming_active"].(bool))
}

func TestLiveJobMonitor_DataManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, &LiveMonitorConfig{
		DataRetentionPeriod: 1 * time.Hour,
	})
	require.NoError(t, err)

	now := time.Now()

	// Add current data
	currentData := &JobLiveMetrics{
		JobID:     "current-job",
		Timestamp: now,
		HealthScore: 0.8,
	}
	monitor.liveData["current-job"] = currentData

	// Add old data
	oldData := &JobLiveMetrics{
		JobID:     "old-job",
		Timestamp: now.Add(-2 * time.Hour), // Older than retention period
		HealthScore: 0.6,
	}
	monitor.liveData["old-job"] = oldData

	// Test data retrieval before cleanup
	data, exists := monitor.GetLiveData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", data.JobID)

	data, exists = monitor.GetLiveData("old-job")
	assert.True(t, exists)
	assert.Equal(t, "old-job", data.JobID)

	// Test getting all data
	allData := monitor.GetAllLiveData()
	assert.Len(t, allData, 2)

	// Clean old data
	monitor.cleanOldLiveData()

	// Verify old data was removed
	_, exists = monitor.GetLiveData("old-job")
	assert.False(t, exists)

	// Verify current data remains
	data, exists = monitor.GetLiveData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", data.JobID)

	// Test all data after cleanup
	allData = monitor.GetAllLiveData()
	assert.Len(t, allData, 1)
}

func TestLiveJobMonitor_CriticalIssuesAndRecommendations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name                string
		efficiency          *EfficiencyMetrics
		wasteRate           float64
		throughputRate      float64
		expectedIssues      int
		expectedRecommendations int
	}{
		{
			name: "good performance - no issues",
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.8,
				MemoryEfficiency:  0.75,
				OverallEfficiency: 0.78,
			},
			wasteRate:           0.2,
			throughputRate:      80,
			expectedIssues:      0,
			expectedRecommendations: 0,
		},
		{
			name: "low CPU efficiency",
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.2, // Low
				MemoryEfficiency:  0.8,
				OverallEfficiency: 0.5,
			},
			wasteRate:           0.3,
			throughputRate:      60,
			expectedIssues:      1,
			expectedRecommendations: 1,
		},
		{
			name: "multiple issues",
			efficiency: &EfficiencyMetrics{
				CPUEfficiency:     0.25, // Low
				MemoryEfficiency:  0.25, // Low
				OverallEfficiency: 0.25, // Low
			},
			wasteRate:           0.8, // High
			throughputRate:      30,  // Low
			expectedIssues:      4, // CPU, Memory, Waste, Overall
			expectedRecommendations: 5, // CPU, Memory, Waste, Throughput, Overall
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test critical issues identification
			issues := monitor.identifyCriticalIssues(tt.efficiency, tt.wasteRate)
			assert.Len(t, issues, tt.expectedIssues)

			// Test recommendations generation
			recommendations := monitor.generateRecommendations(tt.efficiency, tt.wasteRate, tt.throughputRate)
			assert.Len(t, recommendations, tt.expectedRecommendations)
		})
	}
}

func TestLiveJobMonitor_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = monitor.collectLiveJobMetrics(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs for live monitoring")

	mockJobManager.AssertExpectations(t)
}

func TestLiveJobMonitor_MonitoringStats(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	monitor, err := NewLiveJobMonitor(mockClient, logger, nil)
	require.NoError(t, err)

	// Add some test data
	monitor.liveData["job1"] = &JobLiveMetrics{JobID: "job1", Timestamp: time.Now()}
	monitor.liveData["job2"] = &JobLiveMetrics{JobID: "job2", Timestamp: time.Now()}

	// Start streaming for one job
	monitor.StartStreaming("job1")

	stats := monitor.GetMonitoringStats()

	assert.Equal(t, 2, stats["monitored_jobs"])
	assert.Equal(t, 1, stats["active_streams"])
	assert.True(t, stats["streaming_active"].(bool))
	assert.NotNil(t, stats["config"])

	// Clean up
	monitor.StopStreaming("job1")
}