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

func TestNewJobAnalyticsEngine(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	tests := []struct {
		name   string
		config *AnalyticsConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &AnalyticsConfig{
				AnalysisInterval:           60 * time.Second,
				MaxJobsPerAnalysis:         50,
				EnableWasteDetection:       false,
				EnablePerformanceComparison: false,
				EnableTrendAnalysis:        false,
				EnableBenchmarking:         false,
				ComparisonTimeWindow:       3 * 24 * time.Hour,
				MinSampleSizeForComparison: 5,
				HistoricalDataRetention:    7 * 24 * time.Hour,
				CacheTTL:                   30 * time.Minute,
				BatchSize:                  10,
				EnableParallelProcessing:   false,
				MaxConcurrentAnalyses:      2,
				WasteThresholds: &WasteThresholds{
					CPUWasteThreshold:       0.40,
					MemoryWasteThreshold:    0.40,
					TimeWasteThreshold:      0.30,
					ResourceOverallocation:  0.50,
					IdleTimeThreshold:       0.30,
					QueueWasteThreshold:     0.60,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewJobAnalyticsEngine(mockClient, logger, tt.config)
			require.NoError(t, err)
			assert.NotNil(t, engine)
			assert.NotNil(t, engine.config)
			assert.NotNil(t, engine.metrics)
			assert.NotNil(t, engine.efficiencyCalc)
			assert.NotNil(t, engine.analyticsData)
			assert.NotNil(t, engine.wasteAnalysis)

			if tt.config == nil {
				assert.Equal(t, 30*time.Second, engine.config.AnalysisInterval)
				assert.Equal(t, 100, engine.config.MaxJobsPerAnalysis)
				assert.True(t, engine.config.EnableWasteDetection)
				assert.True(t, engine.config.EnablePerformanceComparison)
				assert.True(t, engine.config.EnableTrendAnalysis)
				assert.True(t, engine.config.EnableBenchmarking)
			} else {
				assert.Equal(t, tt.config.AnalysisInterval, engine.config.AnalysisInterval)
				assert.Equal(t, tt.config.MaxJobsPerAnalysis, engine.config.MaxJobsPerAnalysis)
				assert.Equal(t, tt.config.EnableWasteDetection, engine.config.EnableWasteDetection)
				assert.Equal(t, tt.config.EnablePerformanceComparison, engine.config.EnablePerformanceComparison)
			}
		})
	}
}

func TestJobAnalyticsEngine_Describe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	ch := make(chan *prometheus.Desc, 100)
	engine.Describe(ch)
	close(ch)

	descs := []string{}
	for desc := range ch {
		descs = append(descs, desc.String())
	}

	// Verify we have the expected number of metric descriptions
	assert.GreaterOrEqual(t, len(descs), 20)
}

func TestJobAnalyticsEngine_PerformJobAnalytics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, &AnalyticsConfig{
		AnalysisInterval:       30 * time.Second,
		MaxJobsPerAnalysis:     25,
		EnableWasteDetection:   true,
		EnableTrendAnalysis:    true,
		EnableBenchmarking:     true,
		BatchSize:             10,
		HistoricalDataRetention: 24 * time.Hour,
	})
	require.NoError(t, err)

	// Setup mock expectations
	now := time.Now()
	startTime := now.Add(-2 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "analytics-123",
		Name:       "analytics-test-job",
		UserName:   "analyticsuser",
		Account:    "analyticsaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       8,
		Memory:     16384, // 16GB in MB
		Nodes:      2,
		TimeLimit:  240, // 4 hours
		StartTime:  &startTime,
	}

	jobList := &slurm.JobList{
		Jobs: []*slurm.Job{testJob},
	}

	mockJobManager.On("List", mock.Anything, mock.MatchedBy(func(opts *slurm.ListJobsOptions) bool {
		return len(opts.States) >= 2 && opts.MaxCount == 25
	})).Return(jobList, nil)

	// Test analytics processing
	ctx := context.Background()
	err = engine.performJobAnalytics(ctx)
	require.NoError(t, err)

	// Verify analytics data was generated
	analyticsData, exists := engine.GetAnalyticsData("analytics-123")
	assert.True(t, exists)
	assert.NotNil(t, analyticsData)
	assert.Equal(t, "analytics-123", analyticsData.JobID)
	assert.WithinDuration(t, time.Now(), analyticsData.AnalysisTimestamp, 5*time.Second)

	// Verify resource utilization analysis
	assert.NotNil(t, analyticsData.ResourceUtilization)
	assert.NotNil(t, analyticsData.ResourceUtilization.CPUAnalysis)
	assert.NotNil(t, analyticsData.ResourceUtilization.MemoryAnalysis)
	assert.Equal(t, "cpu", analyticsData.ResourceUtilization.CPUAnalysis.ResourceType)
	assert.Equal(t, "memory", analyticsData.ResourceUtilization.MemoryAnalysis.ResourceType)

	// Verify waste analysis was performed
	assert.NotNil(t, analyticsData.WasteAnalysis)
	assert.Equal(t, "analytics-123", analyticsData.WasteAnalysis.JobID)
	assert.GreaterOrEqual(t, analyticsData.WasteAnalysis.TotalWasteScore, 0.0)
	assert.LessOrEqual(t, analyticsData.WasteAnalysis.TotalWasteScore, 1.0)

	// Verify efficiency analysis
	assert.NotNil(t, analyticsData.EfficiencyAnalysis)
	assert.GreaterOrEqual(t, analyticsData.EfficiencyAnalysis.OverallEfficiency, 0.0)
	assert.LessOrEqual(t, analyticsData.EfficiencyAnalysis.OverallEfficiency, 1.0)

	// Verify performance analysis
	assert.NotNil(t, analyticsData.PerformanceMetrics)
	assert.NotNil(t, analyticsData.PerformanceMetrics.ThroughputAnalysis)

	// Verify cost analysis
	assert.NotNil(t, analyticsData.CostAnalysis)
	assert.GreaterOrEqual(t, analyticsData.CostAnalysis.TotalCost, 0.0)

	// Verify optimization insights
	assert.NotNil(t, analyticsData.OptimizationInsights)
	assert.NotNil(t, analyticsData.ActionableRecommendations)

	// Verify overall scores
	assert.GreaterOrEqual(t, analyticsData.OverallScore, 0.0)
	assert.LessOrEqual(t, analyticsData.OverallScore, 1.0)
	assert.Contains(t, []string{"A", "B", "C", "D", "F"}, analyticsData.PerformanceGrade)
	assert.GreaterOrEqual(t, analyticsData.WasteScore, 0.0)
	assert.LessOrEqual(t, analyticsData.WasteScore, 1.0)
	assert.GreaterOrEqual(t, analyticsData.EfficiencyScore, 0.0)
	assert.LessOrEqual(t, analyticsData.EfficiencyScore, 1.0)

	mockJobManager.AssertExpectations(t)
}

func TestJobAnalyticsEngine_ResourceUtilizationAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-1 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "resource-456",
		Name:       "resource-test-job",
		UserName:   "resourceuser",
		Account:    "resourceaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       4,
		Memory:     8192, // 8GB
		Nodes:      1,
		TimeLimit:  120, // 2 hours
		StartTime:  &startTime,
	}

	// Test resource utilization data extraction
	resourceData := engine.extractResourceUtilizationData(testJob)
	assert.NotNil(t, resourceData)
	assert.Equal(t, float64(4), resourceData.CPUAllocated)
	assert.Equal(t, int64(8192*1024*1024), resourceData.MemoryAllocated)
	assert.GreaterOrEqual(t, resourceData.CPUUsed, 0.0)
	assert.LessOrEqual(t, resourceData.CPUUsed, float64(4))
	assert.GreaterOrEqual(t, resourceData.MemoryUsed, int64(0))
	assert.LessOrEqual(t, resourceData.MemoryUsed, resourceData.MemoryAllocated)
	assert.Greater(t, resourceData.WallTime, 0.0)

	// Test resource utilization analysis
	resourceAnalysis := engine.analyzeResourceUtilization(testJob, resourceData)
	assert.NotNil(t, resourceAnalysis)
	assert.NotNil(t, resourceAnalysis.CPUAnalysis)
	assert.NotNil(t, resourceAnalysis.MemoryAnalysis)
	assert.NotNil(t, resourceAnalysis.IOAnalysis)
	assert.NotNil(t, resourceAnalysis.NetworkAnalysis)
	assert.NotNil(t, resourceAnalysis.GPUAnalysis)

	// Verify CPU analysis
	cpuAnalysis := resourceAnalysis.CPUAnalysis
	assert.Equal(t, "cpu", cpuAnalysis.ResourceType)
	assert.Equal(t, float64(4), cpuAnalysis.AllocatedAmount)
	assert.GreaterOrEqual(t, cpuAnalysis.UtilizationRate, 0.0)
	assert.LessOrEqual(t, cpuAnalysis.UtilizationRate, 1.0)
	assert.GreaterOrEqual(t, cpuAnalysis.WastePercentage, 0.0)
	assert.LessOrEqual(t, cpuAnalysis.WastePercentage, 1.0)
	assert.GreaterOrEqual(t, cpuAnalysis.EfficiencyScore, 0.0)
	assert.LessOrEqual(t, cpuAnalysis.EfficiencyScore, 1.0)
	assert.Contains(t, []string{"steady", "bursty", "gradual_increase", "gradual_decrease", "variable"}, cpuAnalysis.UsagePattern)
	assert.GreaterOrEqual(t, cpuAnalysis.PatternConfidence, 0.0)
	assert.LessOrEqual(t, cpuAnalysis.PatternConfidence, 1.0)

	// Verify memory analysis
	memoryAnalysis := resourceAnalysis.MemoryAnalysis
	assert.Equal(t, "memory", memoryAnalysis.ResourceType)
	assert.Equal(t, float64(8192*1024*1024), memoryAnalysis.AllocatedAmount)
	assert.GreaterOrEqual(t, memoryAnalysis.UtilizationRate, 0.0)
	assert.LessOrEqual(t, memoryAnalysis.UtilizationRate, 1.0)

	// Verify correlations and bottlenecks
	assert.NotNil(t, resourceAnalysis.ResourceCorrelations)
	assert.Contains(t, resourceAnalysis.ResourceCorrelations, "cpu_memory")
	assert.Contains(t, resourceAnalysis.ResourceCorrelations, "cpu_io")
	assert.NotNil(t, resourceAnalysis.BottleneckChain)
	assert.NotNil(t, resourceAnalysis.CriticalPath)
}

func TestJobAnalyticsEngine_WasteAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, &AnalyticsConfig{
		EnableWasteDetection: true,
		WasteThresholds: &WasteThresholds{
			CPUWasteThreshold:       0.25,
			MemoryWasteThreshold:    0.25,
			TimeWasteThreshold:      0.20,
			ResourceOverallocation:  0.35,
			IdleTimeThreshold:       0.15,
			QueueWasteThreshold:     0.45,
		},
	})
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-30 * time.Minute)

	// Test job with potential waste
	testJob := &slurm.Job{
		JobID:      "waste-789",
		Name:       "waste-test-job",
		UserName:   "wasteuser",
		Account:    "wasteaccount",
		Partition:  "compute",
		JobState:   "RUNNING",
		CPUs:       8, // High allocation
		Memory:     16384, // 16GB - high allocation
		Nodes:      1,
		TimeLimit:  240, // 4 hours
		StartTime:  &startTime,
	}

	// Create resource data with low utilization (indicating waste)
	resourceData := &ResourceUtilizationData{
		CPUAllocated:    8.0,
		CPUUsed:         2.0,  // 25% utilization - high waste
		MemoryAllocated: 16384 * 1024 * 1024,
		MemoryUsed:      4 * 1024 * 1024 * 1024, // 25% utilization - high waste
		WallTime:        1800, // 30 minutes
		Timestamp:       now,
	}

	resourceAnalysis := engine.analyzeResourceUtilization(testJob, resourceData)

	// Test waste analysis
	wasteAnalysis := engine.performWasteAnalysis(testJob, resourceData, resourceAnalysis)
	assert.NotNil(t, wasteAnalysis)
	assert.Equal(t, "waste-789", wasteAnalysis.JobID)
	assert.GreaterOrEqual(t, wasteAnalysis.TotalWasteScore, 0.0)
	assert.LessOrEqual(t, wasteAnalysis.TotalWasteScore, 1.0)

	// Verify CPU waste detection
	assert.NotNil(t, wasteAnalysis.CPUWaste)
	cpuWaste := wasteAnalysis.CPUWaste
	assert.Equal(t, "cpu", cpuWaste.WasteType)
	assert.Equal(t, 6.0, cpuWaste.Amount) // 8 - 2 = 6 CPUs wasted
	assert.Equal(t, 0.75, cpuWaste.Percentage) // 6/8 = 75% waste
	assert.Greater(t, cpuWaste.Confidence, 0.0)
	assert.NotEmpty(t, cpuWaste.RootCauses)
	assert.NotEmpty(t, cpuWaste.Recommendations)
	assert.Greater(t, cpuWaste.EstimatedCost, 0.0)

	// Verify memory waste detection
	assert.NotNil(t, wasteAnalysis.MemoryWaste)
	memoryWaste := wasteAnalysis.MemoryWaste
	assert.Equal(t, "memory", memoryWaste.WasteType)
	assert.Greater(t, memoryWaste.Amount, 0.0)
	assert.Equal(t, 0.75, memoryWaste.Percentage) // 75% waste
	assert.NotEmpty(t, memoryWaste.RootCauses)
	assert.NotEmpty(t, memoryWaste.Recommendations)

	// Verify time waste analysis
	assert.NotNil(t, wasteAnalysis.TimeWaste)
	timeWaste := wasteAnalysis.TimeWaste
	assert.Equal(t, "time", timeWaste.WasteType)

	// Verify waste categories
	assert.NotNil(t, wasteAnalysis.WasteCategories)
	if len(wasteAnalysis.WasteCategories) > 0 {
		for categoryName, category := range wasteAnalysis.WasteCategories {
			assert.NotEmpty(t, categoryName)
			assert.NotNil(t, category)
			assert.NotEmpty(t, category.CategoryName)
			assert.GreaterOrEqual(t, category.WastePercentage, 0.0)
			assert.LessOrEqual(t, category.WastePercentage, 1.0)
			assert.Contains(t, []string{"low", "medium", "high"}, category.Severity)
			assert.NotEmpty(t, category.Description)
			assert.NotEmpty(t, category.Impact)
		}
	}

	// Verify waste impact analysis
	assert.NotNil(t, wasteAnalysis.WasteImpact)
	impact := wasteAnalysis.WasteImpact
	assert.GreaterOrEqual(t, impact.ClusterImpact, 0.0)
	assert.LessOrEqual(t, impact.ClusterImpact, 1.0)
	assert.GreaterOrEqual(t, impact.FinancialImpact, 0.0)
	assert.LessOrEqual(t, impact.FinancialImpact, 1.0)

	// Verify cost of waste
	assert.GreaterOrEqual(t, wasteAnalysis.CostOfWaste, 0.0)

	// Verify waste reduction plan
	assert.NotNil(t, wasteAnalysis.WasteReductionPlan)
	assert.Greater(t, len(wasteAnalysis.WasteReductionPlan), 0)
	for _, action := range wasteAnalysis.WasteReductionPlan {
		assert.NotEmpty(t, action.Action)
		assert.NotEmpty(t, action.WasteType)
		assert.GreaterOrEqual(t, action.ExpectedReduction, 0.0)
		assert.LessOrEqual(t, action.ExpectedReduction, 1.0)
		assert.Contains(t, []string{"low", "medium", "high"}, action.ImplementationEffort)
		assert.Contains(t, []string{"low", "medium", "high"}, action.Priority)
	}

	// Verify expected savings
	assert.NotNil(t, wasteAnalysis.ExpectedSavings)
	savings := wasteAnalysis.ExpectedSavings
	assert.GreaterOrEqual(t, savings.CostSavings, 0.0)
	assert.NotNil(t, savings.ResourceSavings)
	assert.GreaterOrEqual(t, savings.TimeSavings, 0.0)
	assert.GreaterOrEqual(t, savings.Confidence, 0.0)
	assert.LessOrEqual(t, savings.Confidence, 1.0)
}

func TestJobAnalyticsEngine_SpecificWasteAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name             string
		cpuAllocated     float64
		cpuUsed          float64
		memoryAllocated  int64
		memoryUsed       int64
		expectedCPUWaste float64
		expectedMemWaste float64
	}{
		{
			name:             "high CPU waste",
			cpuAllocated:     8.0,
			cpuUsed:          2.0,
			memoryAllocated:  8 * 1024 * 1024 * 1024,
			memoryUsed:       6 * 1024 * 1024 * 1024,
			expectedCPUWaste: 0.75, // 75% waste
			expectedMemWaste: 0.25, // 25% waste
		},
		{
			name:             "high memory waste",
			cpuAllocated:     4.0,
			cpuUsed:          3.5,
			memoryAllocated:  16 * 1024 * 1024 * 1024,
			memoryUsed:       4 * 1024 * 1024 * 1024,
			expectedCPUWaste: 0.125, // 12.5% waste
			expectedMemWaste: 0.75,  // 75% waste
		},
		{
			name:             "minimal waste",
			cpuAllocated:     4.0,
			cpuUsed:          3.2,
			memoryAllocated:  8 * 1024 * 1024 * 1024,
			memoryUsed:       6.5 * 1024 * 1024 * 1024,
			expectedCPUWaste: 0.2,    // 20% waste
			expectedMemWaste: 0.1875, // 18.75% waste
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resourceData := &ResourceUtilizationData{
				CPUAllocated:    tt.cpuAllocated,
				CPUUsed:         tt.cpuUsed,
				MemoryAllocated: tt.memoryAllocated,
				MemoryUsed:      tt.memoryUsed,
				WallTime:        3600, // 1 hour
				Timestamp:       now,
			}

			resourceAnalysis := &ResourceUtilizationAnalysis{
				CPUAnalysis:    engine.analyzeResource("cpu", tt.cpuAllocated, tt.cpuUsed, 3600),
				MemoryAnalysis: engine.analyzeResource("memory", float64(tt.memoryAllocated), float64(tt.memoryUsed), 3600),
			}

			// Test CPU waste analysis
			cpuWaste := engine.analyzeCPUWaste(resourceData, resourceAnalysis.CPUAnalysis)
			assert.InDelta(t, tt.expectedCPUWaste, cpuWaste.Percentage, 0.01, "CPU waste percentage should match expected")
			assert.Equal(t, tt.cpuAllocated-tt.cpuUsed, cpuWaste.Amount, "CPU waste amount should match")

			// Test memory waste analysis
			memoryWaste := engine.analyzeMemoryWaste(resourceData, resourceAnalysis.MemoryAnalysis)
			assert.InDelta(t, tt.expectedMemWaste, memoryWaste.Percentage, 0.01, "Memory waste percentage should match expected")
			assert.Equal(t, float64(tt.memoryAllocated-tt.memoryUsed), memoryWaste.Amount, "Memory waste amount should match")

			// Verify waste severity classification
			if cpuWaste.Percentage > 0.5 {
				assert.Equal(t, "high", cpuWaste.Severity)
			} else if cpuWaste.Percentage > 0.3 {
				assert.Equal(t, "medium", cpuWaste.Severity)
			} else {
				assert.Equal(t, "low", cpuWaste.Severity)
			}

			if memoryWaste.Percentage > 0.5 {
				assert.Equal(t, "high", memoryWaste.Severity)
			} else if memoryWaste.Percentage > 0.3 {
				assert.Equal(t, "medium", memoryWaste.Severity)
			} else {
				assert.Equal(t, "low", memoryWaste.Severity)
			}
		})
	}
}

func TestJobAnalyticsEngine_EfficiencyAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "efficiency-456",
		CPUs:       4,
		Memory:     8192,
	}

	resourceData := &ResourceUtilizationData{
		CPUAllocated:    4.0,
		CPUUsed:         3.2,
		MemoryAllocated: 8192 * 1024 * 1024,
		MemoryUsed:      6 * 1024 * 1024 * 1024,
		WallTime:        3600,
	}

	efficiencyAnalysis := engine.performEfficiencyAnalysis(testJob, resourceData)
	assert.NotNil(t, efficiencyAnalysis)
	assert.GreaterOrEqual(t, efficiencyAnalysis.OverallEfficiency, 0.0)
	assert.LessOrEqual(t, efficiencyAnalysis.OverallEfficiency, 1.0)
	assert.NotNil(t, efficiencyAnalysis.ResourceEfficiencies)
	assert.Contains(t, efficiencyAnalysis.ResourceEfficiencies, "cpu")
	assert.Contains(t, efficiencyAnalysis.ResourceEfficiencies, "memory")
	assert.Contains(t, efficiencyAnalysis.ResourceEfficiencies, "overall")

	for resourceType, efficiency := range efficiencyAnalysis.ResourceEfficiencies {
		assert.GreaterOrEqual(t, efficiency, 0.0, "Efficiency for %s should be non-negative", resourceType)
		assert.LessOrEqual(t, efficiency, 1.0, "Efficiency for %s should not exceed 1.0", resourceType)
	}
}

func TestJobAnalyticsEngine_PerformanceAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "performance-789",
		CPUs:       8,
		Memory:     16384,
	}

	resourceData := &ResourceUtilizationData{
		CPUAllocated:    8.0,
		CPUUsed:         6.4,
		MemoryAllocated: 16384 * 1024 * 1024,
		MemoryUsed:      12 * 1024 * 1024 * 1024,
		WallTime:        7200,
	}

	performanceAnalysis := engine.performPerformanceAnalysis(testJob, resourceData)
	assert.NotNil(t, performanceAnalysis)
	assert.NotNil(t, performanceAnalysis.ThroughputAnalysis)
	assert.GreaterOrEqual(t, performanceAnalysis.ThroughputAnalysis.AverageThroughput, 0.0)
	assert.NotEmpty(t, performanceAnalysis.ThroughputAnalysis.ThroughputTrend)
}

func TestJobAnalyticsEngine_CostAnalysis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	now := time.Now()
	startTime := now.Add(-2 * time.Hour)

	testJob := &slurm.Job{
		JobID:      "cost-123",
		CPUs:       4,
		Memory:     8192,
		StartTime:  &startTime,
	}

	resourceData := &ResourceUtilizationData{
		CPUAllocated:    4.0,
		CPUUsed:         2.8,
		MemoryAllocated: 8192 * 1024 * 1024,
		MemoryUsed:      6 * 1024 * 1024 * 1024,
		WallTime:        7200, // 2 hours
	}

	// Create a mock waste analysis
	wasteAnalysis := &WasteAnalysisResult{
		CostOfWaste: 0.5,
	}

	costAnalysis := engine.performCostAnalysis(testJob, resourceData, wasteAnalysis)
	assert.NotNil(t, costAnalysis)
	assert.GreaterOrEqual(t, costAnalysis.TotalCost, 0.0)
	assert.Equal(t, 0.5, costAnalysis.CostWaste)
	assert.GreaterOrEqual(t, costAnalysis.CostEfficiency, 0.0)
	assert.LessOrEqual(t, costAnalysis.CostEfficiency, 1.0)

	// Verify cost calculation is reasonable
	expectedCPUCost := 4.0 * 0.10 * 2.0  // 4 CPUs * $0.10/hour * 2 hours
	expectedMemoryCost := 8.0 * 0.02 * 2.0 // 8GB * $0.02/hour * 2 hours
	expectedTotalCost := expectedCPUCost + expectedMemoryCost
	assert.InDelta(t, expectedTotalCost, costAnalysis.TotalCost, 0.01)
}

func TestJobAnalyticsEngine_UpdateAnalyticsMetrics(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	testJob := &slurm.Job{
		JobID:      "metrics-999",
		Name:       "metrics-test-job",
		UserName:   "metricsuser",
		Account:    "metricsaccount",
		Partition:  "compute",
	}

	// Create comprehensive analytics data
	analyticsData := &JobAnalyticsData{
		JobID: "metrics-999",
		WasteAnalysis: &WasteAnalysisResult{
			TotalWasteScore: 0.4,
			CPUWaste: &WasteDetail{
				WasteType:     "cpu",
				Percentage:    0.3,
				EstimatedCost: 2.5,
			},
			MemoryWaste: &WasteDetail{
				WasteType:     "memory",
				Percentage:    0.5,
				EstimatedCost: 1.8,
			},
		},
		EfficiencyAnalysis: &EfficiencyAnalysisResult{
			OverallEfficiency: 0.75,
			ResourceEfficiencies: map[string]float64{
				"cpu":     0.8,
				"memory":  0.7,
				"overall": 0.75,
			},
		},
		PerformanceMetrics: &PerformanceAnalyticsResult{
			ThroughputAnalysis: &ThroughputAnalysis{
				AverageThroughput: 85.0,
			},
		},
		CostAnalysis: &CostAnalysisResult{
			TotalCost:      10.0,
			CostEfficiency: 0.8,
			CostWaste:      2.0,
		},
		ROIAnalysis: &ROIAnalysisResult{
			ROI: 1.5,
		},
		OptimizationInsights: &OptimizationInsights{
			QuickWins: []QuickWinOpportunity{
				{Opportunity: "Reduce CPU allocation"},
				{Opportunity: "Optimize memory usage"},
			},
			PrimaryInsights: []OptimizationInsight{
				{Category: "efficiency", Impact: 0.3},
				{Category: "cost", Impact: 0.2},
			},
		},
		ActionableRecommendations: []ActionableRecommendation{
			{
				Category: "resource_optimization",
				ExpectedImpact: &ExpectedImpact{
					PerformanceGain: 0.15,
				},
			},
		},
		OverallScore: 0.78,
	}

	// Update metrics
	engine.updateAnalyticsMetrics(testJob, analyticsData)

	// Verify waste metrics
	wasteScoreMetric := testutil.ToFloat64(engine.metrics.WasteScore.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.4, wasteScoreMetric)

	cpuWasteMetric := testutil.ToFloat64(engine.metrics.ResourceWasteDetected.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute", "cpu",
	))
	assert.Equal(t, 1.0, cpuWasteMetric)

	// Verify efficiency metrics
	efficiencyMetric := testutil.ToFloat64(engine.metrics.EfficiencyScore.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.75, efficiencyMetric)

	cpuGapMetric := testutil.ToFloat64(engine.metrics.EfficiencyGap.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute", "cpu",
	))
	assert.Equal(t, 0.0, cpuGapMetric) // 0.8 efficiency vs 0.8 target = 0 gap

	// Verify performance metrics
	performanceMetric := testutil.ToFloat64(engine.metrics.PerformanceScore.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.78, performanceMetric)

	throughputMetric := testutil.ToFloat64(engine.metrics.ThroughputAnalysis.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 85.0, throughputMetric)

	// Verify cost metrics
	costEfficiencyMetric := testutil.ToFloat64(engine.metrics.CostEfficiency.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 0.8, costEfficiencyMetric)

	costWasteMetric := testutil.ToFloat64(engine.metrics.CostWaste.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 2.0, costWasteMetric)

	// Verify ROI metrics
	roiMetric := testutil.ToFloat64(engine.metrics.ROIScore.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 1.5, roiMetric)

	// Verify optimization metrics
	quickWinsMetric := testutil.ToFloat64(engine.metrics.QuickWinsIdentified.WithLabelValues(
		"metrics-999", "metricsuser", "metricsaccount", "compute",
	))
	assert.Equal(t, 2.0, quickWinsMetric)
}

func TestJobAnalyticsEngine_ResourceAnalysisMethods(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name            string
		allocated       float64
		used            float64
		wallTime        float64
		expectedPattern string
		expectedEfficiency float64
	}{
		{
			name:            "high utilization steady pattern",
			allocated:       8.0,
			used:            6.8,
			wallTime:        3600,
			expectedPattern: "steady",
			expectedEfficiency: 0.85,
		},
		{
			name:            "low utilization bursty pattern",
			allocated:       8.0,
			used:            2.0,
			wallTime:        1800,
			expectedPattern: "bursty",
			expectedEfficiency: 0.25,
		},
		{
			name:            "medium utilization variable pattern",
			allocated:       4.0,
			used:            2.2,
			wallTime:        7200,
			expectedPattern: "variable",
			expectedEfficiency: 0.55,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := engine.analyzeResource("cpu", tt.allocated, tt.used, tt.wallTime)

			assert.Equal(t, "cpu", analysis.ResourceType)
			assert.Equal(t, tt.allocated, analysis.AllocatedAmount)
			assert.Equal(t, tt.used, analysis.AverageUsage)
			assert.Equal(t, tt.used/tt.allocated, analysis.UtilizationRate)
			assert.InDelta(t, (tt.allocated-tt.used)/tt.allocated, analysis.WastePercentage, 0.01)
			assert.Equal(t, tt.expectedPattern, analysis.UsagePattern)
			assert.GreaterOrEqual(t, analysis.PatternConfidence, 0.0)
			assert.LessOrEqual(t, analysis.PatternConfidence, 1.0)
			assert.GreaterOrEqual(t, analysis.EfficiencyScore, 0.0)
			assert.LessOrEqual(t, analysis.EfficiencyScore, 1.0)
			assert.Greater(t, analysis.OptimalAllocation, 0.0)
			assert.GreaterOrEqual(t, analysis.PotentialSavings, 0.0)
			assert.NotEmpty(t, analysis.RecommendedAction)
		})
	}
}

func TestJobAnalyticsEngine_DataManagement(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, &AnalyticsConfig{
		HistoricalDataRetention: 1 * time.Hour,
	})
	require.NoError(t, err)

	now := time.Now()

	// Add current analytics data
	currentData := &JobAnalyticsData{
		JobID:             "current-job",
		AnalysisTimestamp: now,
		OverallScore:      0.8,
	}
	engine.analyticsData["current-job"] = currentData

	// Add old analytics data
	oldData := &JobAnalyticsData{
		JobID:             "old-job",
		AnalysisTimestamp: now.Add(-2 * time.Hour),
		OverallScore:      0.6,
	}
	engine.analyticsData["old-job"] = oldData

	// Test data retrieval before cleanup
	data, exists := engine.GetAnalyticsData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", data.JobID)

	data, exists = engine.GetAnalyticsData("old-job")
	assert.True(t, exists)
	assert.Equal(t, "old-job", data.JobID)

	// Clean old data
	engine.cleanOldAnalyticsData()

	// Verify old data was removed
	_, exists = engine.GetAnalyticsData("old-job")
	assert.False(t, exists)

	// Verify current data remains
	data, exists = engine.GetAnalyticsData("current-job")
	assert.True(t, exists)
	assert.Equal(t, "current-job", data.JobID)

	// Test analytics stats
	stats := engine.GetAnalyticsStats()
	assert.Equal(t, 1, stats["jobs_analyzed"])
	assert.NotNil(t, stats["config"])
}

func TestJobAnalyticsEngine_ErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockJobManager := &MockJobManager{}
	mockClient := &MockSlurmClient{
		jobManager: mockJobManager,
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	// Test when job manager returns error
	mockJobManager.On("List", mock.Anything, mock.Anything).Return((*slurm.JobList)(nil), assert.AnError)

	ctx := context.Background()
	err = engine.performJobAnalytics(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list jobs for analytics")

	mockJobManager.AssertExpectations(t)
}

func TestJobAnalyticsEngine_ScoreCalculations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	mockClient := &MockSlurmClient{
		jobManager: &MockJobManager{},
	}

	engine, err := NewJobAnalyticsEngine(mockClient, logger, nil)
	require.NoError(t, err)

	tests := []struct {
		name                string
		efficiencyScore     float64
		wasteScore          float64
		expectedOverall     float64
		expectedGrade       string
	}{
		{
			name:            "excellent performance",
			efficiencyScore: 0.95,
			wasteScore:      0.05,
			expectedOverall: 0.95,
			expectedGrade:   "A",
		},
		{
			name:            "good performance",
			efficiencyScore: 0.85,
			wasteScore:      0.15,
			expectedOverall: 0.85,
			expectedGrade:   "B",
		},
		{
			name:            "average performance",
			efficiencyScore: 0.75,
			wasteScore:      0.25,
			expectedOverall: 0.75,
			expectedGrade:   "C",
		},
		{
			name:            "below average performance",
			efficiencyScore: 0.65,
			wasteScore:      0.35,
			expectedOverall: 0.65,
			expectedGrade:   "D",
		},
		{
			name:            "poor performance",
			efficiencyScore: 0.45,
			wasteScore:      0.55,
			expectedOverall: 0.45,
			expectedGrade:   "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			efficiency := &EfficiencyAnalysisResult{
				OverallEfficiency: tt.efficiencyScore,
			}

			waste := &WasteAnalysisResult{
				TotalWasteScore: tt.wasteScore,
			}

			performance := &PerformanceAnalyticsResult{}

			// Calculate overall score
			overallScore := engine.calculateOverallScore(efficiency, performance, waste)
			assert.InDelta(t, tt.expectedOverall, overallScore, 0.01)

			// Calculate performance grade
			grade := engine.calculatePerformanceGrade(overallScore)
			assert.Equal(t, tt.expectedGrade, grade)

			// Verify individual score calculations
			wasteScoreResult := engine.calculateWasteScore(waste)
			assert.Equal(t, tt.wasteScore, wasteScoreResult)

			efficiencyScoreResult := engine.calculateEfficiencyScore(efficiency)
			assert.Equal(t, tt.efficiencyScore, efficiencyScoreResult)
		})
	}
}