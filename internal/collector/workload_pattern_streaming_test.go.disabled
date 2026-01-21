package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWorkloadPatternStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockWorkloadPatternStreamingSLURMClient) StreamWorkloadPatternEvents(ctx context.Context) (<-chan WorkloadPatternEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan WorkloadPatternEvent), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternStreamingConfiguration(ctx context.Context) (*PatternStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PatternStreamingConfiguration), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetActivePatternStreams(ctx context.Context) ([]*ActivePatternStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActivePatternStream), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternEventHistory(ctx context.Context, patternID string, duration time.Duration) ([]*PatternEvent, error) {
	args := m.Called(ctx, patternID, duration)
	return args.Get(0).([]*PatternEvent), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternStreamingMetrics(ctx context.Context) (*PatternStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PatternStreamingMetrics), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternEventFilters(ctx context.Context) ([]*PatternEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*PatternEventFilter), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) ConfigurePatternStreaming(ctx context.Context, config *PatternStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternStreamingStatus(ctx context.Context) (*PatternStreamingStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PatternStreamingStatus), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternEventSubscriptions(ctx context.Context) ([]*PatternEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*PatternEventSubscription), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) ManagePatternEventSubscription(ctx context.Context, subscription *PatternEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternEventProcessingStats(ctx context.Context) (*PatternEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PatternEventProcessingStats), args.Error(1)
}

func (m *MockWorkloadPatternStreamingSLURMClient) GetPatternStreamingPerformanceMetrics(ctx context.Context) (*PatternStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PatternStreamingPerformanceMetrics), args.Error(1)
}

func TestNewWorkloadPatternStreamingCollector(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.patternEvents)
	assert.NotNil(t, collector.activePatternStreams)
	assert.NotNil(t, collector.patternDetectionRate)
	assert.NotNil(t, collector.patternConfidence)
	assert.NotNil(t, collector.patternSignificance)
	assert.NotNil(t, collector.patternComplexity)
	assert.NotNil(t, collector.patternValue)
	assert.NotNil(t, collector.resourcePatterns)
	assert.NotNil(t, collector.userBehaviorPatterns)
	assert.NotNil(t, collector.anomalyPatterns)
	assert.NotNil(t, collector.optimizationPotential)
	assert.NotNil(t, collector.businessImpactScore)
	assert.NotNil(t, collector.predictionAccuracy)
	assert.NotNil(t, collector.mlModelPerformance)
	assert.NotNil(t, collector.analysisEfficiency)
	assert.NotNil(t, collector.patternEvolution)
	assert.NotNil(t, collector.costOptimizationMetrics)
	assert.NotNil(t, collector.compliancePatterns)
	assert.NotNil(t, collector.scalingPatterns)
}

func TestWorkloadPatternStreamingCollector_Describe(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.GreaterOrEqual(t, count, 19) // At least the base metrics
}

func TestWorkloadPatternStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	config := &PatternStreamingConfiguration{
		StreamingEnabled:          true,
		DetectionInterval:         time.Minute * 5,
		AnalysisWindowSize:        time.Hour * 24,
		EventBufferSize:           10000,
		MaxConcurrentAnalysis:     20,
		StreamTimeout:             time.Hour,
		PatternTypes:              []string{"resource", "temporal", "behavioral"},
		MinPatternConfidence:      0.75,
		MinPatternSignificance:    0.6,
		MinPatternFrequency:       5,
		MinPatternDuration:        time.Hour,
		MaxPatternAge:             time.Hour * 24 * 30,
		PatternMergeThreshold:     0.85,
		TimeSeriesAnalysis:        true,
		FrequencyAnalysis:         true,
		CorrelationAnalysis:       true,
		CausalityAnalysis:         true,
		PredictiveAnalysis:        true,
		AnomalyDetection:          true,
		TrendAnalysis:             true,
		SeasonalAnalysis:          true,
		ResourcePatternDetection:  true,
		UtilizationAnalysis:       true,
		EfficiencyAnalysis:        true,
		WasteDetection:            true,
		BottleneckDetection:       true,
		ContentionAnalysis:        true,
		ScalingAnalysis:           true,
		CapacityAnalysis:          true,
		UserBehaviorAnalysis:      true,
		SubmissionPatternAnalysis: true,
		ExecutionPatternAnalysis:  true,
		FailurePatternAnalysis:    true,
		MigrationPatternAnalysis:  true,
		CollaborationAnalysis:     true,
		LearningCurveAnalysis:     true,
		SatisfactionAnalysis:      true,
		ParallelProcessing:        true,
		GPUAcceleration:           true,
		CachingEnabled:            true,
		CompressionEnabled:        true,
		SamplingEnabled:           true,
		SamplingRate:              0.1,
		IncrementalAnalysis:       true,
		StreamProcessing:          true,
		MLEnabled:                 true,
		ModelTypes:                []string{"lstm", "random_forest", "clustering"},
		TrainingEnabled:           true,
		OnlineLearning:            true,
		ModelUpdateFrequency:      time.Hour * 6,
		FeatureEngineering:        true,
		AutoMLEnabled:             true,
		EnsembleMethods:           true,
		OptimizationAnalysis:      true,
		CostOptimization:          true,
		PerformanceOptimization:   true,
		ResourceOptimization:      true,
		SchedulingOptimization:    true,
		PlacementOptimization:     true,
		ScalingOptimization:       true,
		WorkflowOptimization:      true,
		BusinessImpactAnalysis:    true,
		ROICalculation:            true,
		CostBenefitAnalysis:       true,
		RiskAssessment:            true,
		ComplianceChecking:        true,
		SLAMonitoring:             true,
		KPITracking:               true,
		ValueStreamMapping:        true,
		AlertingEnabled:           true,
		AlertThresholds: map[string]float64{
			"anomaly_score":       0.8,
			"optimization_score":  0.7,
			"business_impact":     0.6,
		},
		AutomatedActions:          true,
		RecommendationEngine:      true,
		DecisionSupport:           true,
		PrescriptiveAnalytics:     true,
		PlaybookIntegration:       true,
		WorkflowAutomation:        true,
		DataRetentionPeriod:       time.Hour * 24 * 90,
		DataArchiving:             true,
		DataCompression:           true,
		DataEncryption:            true,
		DataAnonymization:         true,
		DataQualityChecks:         true,
		DataLineageTracking:       true,
		DataGovernance:            true,
	}

	client.On("GetPatternStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActivePatternStreams", mock.Anything).Return([]*ActivePatternStream{}, nil)
	client.On("GetPatternStreamingMetrics", mock.Anything).Return(&PatternStreamingMetrics{}, nil)
	client.On("GetPatternStreamingStatus", mock.Anything).Return(&PatternStreamingStatus{}, nil)
	client.On("GetPatternEventSubscriptions", mock.Anything).Return([]*PatternEventSubscription{}, nil)
	client.On("GetPatternEventFilters", mock.Anything).Return([]*PatternEventFilter{}, nil)
	client.On("GetPatternEventProcessingStats", mock.Anything).Return(&PatternEventProcessingStats{}, nil)
	client.On("GetPatternStreamingPerformanceMetrics", mock.Anything).Return(&PatternStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 150)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := []prometheus.Metric{}
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.Greater(t, len(metrics), 0)
	client.AssertExpectations(t)
}

func TestWorkloadPatternStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActivePatternStream{
		{
			StreamID:                  "pattern-stream-001",
			StreamName:                "workload_pattern_detector",
			StreamType:                "comprehensive",
			CreatedAt:                 now.Add(-time.Hour * 48),
			LastActivityAt:            now,
			Status:                    "healthy",
			AnalysisWindowStart:       now.Add(-time.Hour * 24),
			AnalysisWindowEnd:         now,
			EventsAnalyzed:            2500000,
			PatternsDetected:          1250,
			ActivePatterns:            85,
			EvolvingPatterns:          25,
			StablePatterns:            60,
			AnalysisRate:              125.5,
			DetectionLatency:          time.Millisecond * 250,
			ProcessingEfficiency:      0.92,
			ResourceUtilization:       0.68,
			MemoryUsage:               4294967296, // 4GB
			CPUUsage:                  0.45,
			DetectionAccuracy:         0.95,
			FalsePositiveRate:         0.03,
			FalseNegativeRate:         0.02,
			PrecisionScore:            0.97,
			RecallScore:               0.98,
			F1Score:                   0.975,
			MostFrequentPattern:       "daily_batch_submission",
			MostSignificantPattern:    "resource_scaling_burst",
			LongestPattern:            "seasonal_workload_cycle",
			MostComplexPattern:        "multi_user_collaboration",
			MostValuablePattern:       "cost_optimization_opportunity",
			EmergingPatterns:          []string{"gpu_utilization_spike", "memory_intensive_jobs"},
			DecliningPatterns:         []string{"legacy_workflow", "inefficient_submission"},
			BusinessValueIdentified:   3500000.0,
			CostSavingsPotential:      850000.0,
			EfficiencyGainsPotential:  0.25,
			OptimizationOpportunities: 125,
			RiskIndicators:            15,
			ComplianceIssues:          3,
			ModelAccuracy:             0.91,
			ModelPrecision:            0.93,
			ModelRecall:               0.89,
			ModelF1Score:              0.91,
			ModelTrainingTime:         time.Hour * 2,
			ModelInferenceTime:        time.Millisecond * 50,
			ModelVersion:              "v2.3.1",
			ModelLastUpdated:          now.Add(-time.Hour * 6),
		},
		{
			StreamID:                  "pattern-stream-002",
			StreamName:                "resource_pattern_analyzer",
			StreamType:                "resource_focused",
			CreatedAt:                 now.Add(-time.Hour * 24),
			LastActivityAt:            now.Add(-time.Minute * 5),
			Status:                    "degraded",
			AnalysisWindowStart:       now.Add(-time.Hour * 12),
			AnalysisWindowEnd:         now.Add(-time.Minute * 5),
			EventsAnalyzed:            1200000,
			PatternsDetected:          450,
			ActivePatterns:            32,
			EvolvingPatterns:          12,
			StablePatterns:            20,
			AnalysisRate:              85.3,
			DetectionLatency:          time.Millisecond * 450,
			ProcessingEfficiency:      0.85,
			ResourceUtilization:       0.82,
			MemoryUsage:               3221225472, // 3GB
			CPUUsage:                  0.65,
			DetectionAccuracy:         0.89,
			FalsePositiveRate:         0.06,
			FalseNegativeRate:         0.05,
			PrecisionScore:            0.91,
			RecallScore:               0.87,
			F1Score:                   0.89,
			MostFrequentPattern:       "cpu_burst_pattern",
			MostSignificantPattern:    "memory_leak_pattern",
			LongestPattern:            "gradual_resource_increase",
			MostComplexPattern:        "multi_resource_contention",
			MostValuablePattern:       "underutilization_pattern",
			EmergingPatterns:          []string{"new_gpu_workload", "container_scaling"},
			DecliningPatterns:         []string{"legacy_batch_jobs"},
			BusinessValueIdentified:   1800000.0,
			CostSavingsPotential:      425000.0,
			EfficiencyGainsPotential:  0.18,
			OptimizationOpportunities: 65,
			RiskIndicators:            8,
			ComplianceIssues:          1,
			ModelAccuracy:             0.87,
			ModelPrecision:            0.89,
			ModelRecall:               0.85,
			ModelF1Score:              0.87,
			ModelTrainingTime:         time.Hour * 1,
			ModelInferenceTime:        time.Millisecond * 75,
			ModelVersion:              "v2.2.0",
			ModelLastUpdated:          now.Add(-time.Hour * 12),
		},
	}

	client.On("GetPatternStreamingConfiguration", mock.Anything).Return(&PatternStreamingConfiguration{}, nil)
	client.On("GetActivePatternStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetPatternStreamingMetrics", mock.Anything).Return(&PatternStreamingMetrics{}, nil)
	client.On("GetPatternStreamingStatus", mock.Anything).Return(&PatternStreamingStatus{}, nil)
	client.On("GetPatternEventSubscriptions", mock.Anything).Return([]*PatternEventSubscription{}, nil)
	client.On("GetPatternEventFilters", mock.Anything).Return([]*PatternEventFilter{}, nil)
	client.On("GetPatternEventProcessingStats", mock.Anything).Return(&PatternEventProcessingStats{}, nil)
	client.On("GetPatternStreamingPerformanceMetrics", mock.Anything).Return(&PatternStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 150)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := []prometheus.Metric{}
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.Greater(t, len(metrics), 0)
	client.AssertExpectations(t)
}

func TestWorkloadPatternStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	streamingMetrics := &PatternStreamingMetrics{
		TotalStreams:              15,
		ActiveStreams:             12,
		HealthyStreams:            10,
		DegradedStreams:           2,
		StreamEfficiency:          0.88,
		StreamReliability:         0.95,
		TotalPatternsDetected:     50000,
		UniquePatterns:            2500,
		RecurringPatterns:         1800,
		EmergingPatterns:          350,
		StablePatterns:            1200,
		DecliningPatterns:         250,
		EventsAnalyzed:            25000000,
		AnalysisRate:              2500.5,
		AnalysisLatency:           time.Millisecond * 150,
		AnalysisAccuracy:          0.93,
		AnalysisCoverage:          0.95,
		AnalysisDepth:             0.87,
		AverageConfidence:         0.85,
		AverageSignificance:       0.78,
		PatternStability:          0.82,
		PatternDiversity:          0.75,
		PatternComplexity:         0.68,
		PatternValue:              0.88,
		CPUUtilization:            0.65,
		MemoryUtilization:         0.72,
		StorageUtilization:        0.58,
		NetworkUtilization:        0.45,
		GPUUtilization:            0.82,
		ResourceEfficiency:        0.91,
		ModelPerformance:          0.89,
		PredictionAccuracy:        0.86,
		TrainingEfficiency:        0.92,
		InferenceSpeed:            0.95,
		ModelComplexity:           0.75,
		FeatureImportance: map[string]float64{
			"cpu_usage":     0.85,
			"memory_usage":  0.78,
			"job_duration":  0.72,
			"user_behavior": 0.68,
		},
		TotalBusinessValue:        15000000.0,
		IdentifiedSavings:         3500000.0,
		OptimizationPotential:     0.35,
		RiskReduction:             0.45,
		ComplianceImprovement:     0.85,
		UserSatisfactionImpact:    0.78,
		OptimizationsIdentified:   850,
		OptimizationsImplemented:  450,
		EfficiencyGains:           0.25,
		CostReductions:            0.22,
		PerformanceImprovements:   0.28,
		ResourceSavings:           0.32,
		PatternGrowthRate:         0.15,
		ComplexityTrend:           0.08,
		ValueTrend:                0.12,
		EfficiencyTrend:           0.18,
		AdoptionRate:              0.65,
		EvolutionRate:             0.25,
	}

	client.On("GetPatternStreamingConfiguration", mock.Anything).Return(&PatternStreamingConfiguration{}, nil)
	client.On("GetActivePatternStreams", mock.Anything).Return([]*ActivePatternStream{}, nil)
	client.On("GetPatternStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetPatternStreamingStatus", mock.Anything).Return(&PatternStreamingStatus{}, nil)
	client.On("GetPatternEventSubscriptions", mock.Anything).Return([]*PatternEventSubscription{}, nil)
	client.On("GetPatternEventFilters", mock.Anything).Return([]*PatternEventFilter{}, nil)
	client.On("GetPatternEventProcessingStats", mock.Anything).Return(&PatternEventProcessingStats{}, nil)
	client.On("GetPatternStreamingPerformanceMetrics", mock.Anything).Return(&PatternStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 150)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := []prometheus.Metric{}
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.Greater(t, len(metrics), 0)
	client.AssertExpectations(t)
}

func TestWorkloadPatternStreamingCollector_CollectError(t *testing.T) {
	client := &MockWorkloadPatternStreamingSLURMClient{}
	collector := NewWorkloadPatternStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetPatternStreamingConfiguration", mock.Anything).Return((*PatternStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActivePatternStreams", mock.Anything).Return([]*ActivePatternStream(nil), assert.AnError)
	client.On("GetPatternStreamingMetrics", mock.Anything).Return((*PatternStreamingMetrics)(nil), assert.AnError)
	client.On("GetPatternStreamingStatus", mock.Anything).Return((*PatternStreamingStatus)(nil), assert.AnError)
	client.On("GetPatternEventSubscriptions", mock.Anything).Return([]*PatternEventSubscription(nil), assert.AnError)
	client.On("GetPatternEventFilters", mock.Anything).Return([]*PatternEventFilter(nil), assert.AnError)
	client.On("GetPatternEventProcessingStats", mock.Anything).Return((*PatternEventProcessingStats)(nil), assert.AnError)
	client.On("GetPatternStreamingPerformanceMetrics", mock.Anything).Return((*PatternStreamingPerformanceMetrics)(nil), assert.AnError)

	ch := make(chan prometheus.Metric, 150)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	// Should complete without panicking despite errors
	metrics := []prometheus.Metric{}
	for m := range ch {
		metrics = append(metrics, m)
	}

	client.AssertExpectations(t)
}