package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEventAggregationCorrelationSLURMClient struct {
	mock.Mock
}

func (m *MockEventAggregationCorrelationSLURMClient) StreamAggregatedEvents(ctx context.Context) (<-chan AggregatedEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan AggregatedEvent), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetAggregationConfiguration(ctx context.Context) (*AggregationConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AggregationConfiguration), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetActiveAggregations(ctx context.Context) ([]*ActiveAggregation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveAggregation), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetAggregationHistory(ctx context.Context, aggregationID string, duration time.Duration) ([]*AggregationRecord, error) {
	args := m.Called(ctx, aggregationID, duration)
	return args.Get(0).([]*AggregationRecord), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetAggregationMetrics(ctx context.Context) (*AggregationMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AggregationMetrics), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) ConfigureAggregation(ctx context.Context, config *AggregationConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockEventAggregationCorrelationSLURMClient) StreamCorrelatedEvents(ctx context.Context) (<-chan CorrelatedEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan CorrelatedEvent), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetCorrelationConfiguration(ctx context.Context) (*CorrelationConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*CorrelationConfiguration), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetActiveCorrelations(ctx context.Context) ([]*ActiveCorrelation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveCorrelation), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetCorrelationHistory(ctx context.Context, correlationID string, duration time.Duration) ([]*CorrelationRecord, error) {
	args := m.Called(ctx, correlationID, duration)
	return args.Get(0).([]*CorrelationRecord), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) GetCorrelationMetrics(ctx context.Context) (*CorrelationMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*CorrelationMetrics), args.Error(1)
}

func (m *MockEventAggregationCorrelationSLURMClient) ConfigureCorrelation(ctx context.Context, config *CorrelationConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func TestNewEventAggregationCorrelationCollector(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.aggregatedEvents)
	assert.NotNil(t, collector.activeAggregations)
	assert.NotNil(t, collector.aggregationLatency)
	assert.NotNil(t, collector.aggregationQuality)
	assert.NotNil(t, collector.aggregationThroughput)
	assert.NotNil(t, collector.aggregationErrors)
	assert.NotNil(t, collector.correlatedEvents)
	assert.NotNil(t, collector.activeCorrelations)
	assert.NotNil(t, collector.correlationLatency)
	assert.NotNil(t, collector.correlationAccuracy)
	assert.NotNil(t, collector.correlationConfidence)
	assert.NotNil(t, collector.correlationComplexity)
	assert.NotNil(t, collector.processingTime)
	assert.NotNil(t, collector.resourceUtilization)
	assert.NotNil(t, collector.dataQuality)
	assert.NotNil(t, collector.systemHealth)
	assert.NotNil(t, collector.businessValue)
	assert.NotNil(t, collector.costOptimization)
	assert.NotNil(t, collector.operationalEfficiency)
	assert.NotNil(t, collector.decisionSupport)
	assert.NotNil(t, collector.dataCompleteness)
	assert.NotNil(t, collector.analysisReliability)
	assert.NotNil(t, collector.resultValidation)
	assert.NotNil(t, collector.performanceIndex)
}

func TestEventAggregationCorrelationCollector_Describe(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	ch := make(chan *prometheus.Desc, 200)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.GreaterOrEqual(t, count, 24) // At least the base metrics
}

func TestEventAggregationCorrelationCollector_Collect_AggregationConfiguration(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	aggregationConfig := &AggregationConfiguration{
		AggregationEnabled:   true,
		AggregationInterval:  time.Minute * 5,
		WindowSize:           time.Hour * 1,
		WindowType:           "sliding",
		BufferSize:           10000,
		MaxConcurrentWindows: 50,
		ProcessingTimeout:    time.Minute * 30,
		AggregationMethods:   []string{"count", "sum", "avg", "percentile"},
		StatisticalFunctions: []string{"mean", "median", "stddev"},
		GroupingKeys:         []string{"user", "account", "partition"},
		AggregationLevel:     "cluster",
		AggregationScope:     "global",
		SourceEventTypes:     []string{"job", "node", "system"},
		RealTimeProcessing:   true,
		BatchProcessing:      true,
		StreamProcessing:     true,
		ParallelProcessing:   true,
		QualityThreshold:     0.95,
		AccuracyThreshold:    0.90,
		CompletenessThreshold: 0.85,
		TimelinessThreshold:  time.Minute * 5,
		ConsistencyChecks:    true,
		ValidationRules:      []string{"completeness", "accuracy", "consistency"},
		OutlierDetection:     true,
		NoiseReduction:       true,
		DataCleaning:         true,
		OutputFormats:        []string{"json", "csv", "parquet"},
		CompressionEnabled:   true,
		EncryptionEnabled:    true,
		AlertingEnabled:      true,
		ReportingEnabled:     true,
		RetentionPeriod:      time.Hour * 24 * 30,
		ArchiveEnabled:       true,
		BackupEnabled:        true,
		DataGovernance:       true,
		MaxMemoryUsage:       8589934592, // 8GB
		MaxCPUUsage:          0.8,
		ScalingPolicies:      []string{"horizontal", "vertical"},
		LoadBalancing:        true,
		FailoverEnabled:      true,
		AccessControl:        true,
		AuthenticationRequired: true,
		DataAnonymization:    true,
		AuditLogging:         true,
		HealthMonitoring:     true,
		PerformanceMonitoring: true,
		QualityMonitoring:    true,
	}

	correlationConfig := &CorrelationConfiguration{
		CorrelationEnabled:   true,
		CorrelationInterval:  time.Minute * 2,
		AnalysisWindow:       time.Hour * 2,
		CorrelationDepth:     5,
		MaxCorrelations:      1000,
		CorrelationTimeout:   time.Minute * 15,
		ProcessingThreads:    20,
		MemoryBufferSize:     4294967296, // 4GB
		CorrelationMethods:   []string{"temporal", "causal", "statistical", "pattern"},
		TemporalMethods:      []string{"sequence", "proximity", "overlap"},
		StatisticalMethods:   []string{"pearson", "spearman", "mutual_info"},
		PatternMethods:       []string{"frequent_patterns", "sequential_patterns"},
		MLMethods:            []string{"clustering", "classification", "anomaly_detection"},
		CausalityMethods:     []string{"granger", "pc_algorithm"},
		CorrelationThreshold: 0.7,
		SignificanceLevel:    0.05,
		ConfidenceLevel:      0.95,
		MinimumSupport:       0.1,
		MinimumConfidence:    0.8,
		MaxPatternLength:     10,
		WindowOverlap:        0.5,
		AnalysisAccuracy:     0.9,
		ValidationRequired:   true,
		CrossValidation:      true,
		EventTypes:           []string{"job", "node", "queue", "partition"},
		SourceSystems:        []string{"slurm", "monitoring", "logging"},
		MLModelsEnabled:      true,
		ModelTypes:           []string{"supervised", "unsupervised", "deep_learning"},
		TrainingEnabled:      true,
		TrainingFrequency:    time.Hour * 6,
		FeatureEngineering:   true,
		FeatureSelection:     true,
		OnlineLearning:       true,
		GraphAnalysisEnabled: true,
		GraphAlgorithms:      []string{"pagerank", "community_detection", "centrality"},
		CommunityDetection:   true,
		InfluenceAnalysis:    true,
		NetworkPropagation:   true,
		ParallelProcessing:   true,
		DistributedProcessing: true,
		GPUAcceleration:      true,
		StreamProcessing:     true,
		BatchProcessing:      true,
		CachingStrategy:      "lru",
		IndexingStrategy:     "btree",
		CompressionEnabled:   true,
		OptimizationLevel:    "high",
		ResourceManagement:   true,
		OutputFormats:        []string{"json", "graphml", "csv"},
		ReportGeneration:     true,
		VisualizationEnabled: true,
		DashboardIntegration: true,
		AlertingEnabled:      true,
		QualityAssurance:     true,
		ResultValidation:     true,
		PeerReview:           true,
		StatisticalValidation: true,
		DataPrivacy:          true,
		AccessControl:        true,
		AuditTrail:           true,
		ComplianceChecks:     true,
		DataGovernance:       true,
	}

	client.On("GetAggregationConfiguration", mock.Anything).Return(aggregationConfig, nil)
	client.On("GetCorrelationConfiguration", mock.Anything).Return(correlationConfig, nil)
	client.On("GetActiveAggregations", mock.Anything).Return([]*ActiveAggregation{}, nil)
	client.On("GetActiveCorrelations", mock.Anything).Return([]*ActiveCorrelation{}, nil)
	client.On("GetAggregationMetrics", mock.Anything).Return(&AggregationMetrics{}, nil)
	client.On("GetCorrelationMetrics", mock.Anything).Return(&CorrelationMetrics{}, nil)

	ch := make(chan prometheus.Metric, 200)
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

func TestEventAggregationCorrelationCollector_Collect_ActiveOperations(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	now := time.Now()
	activeAggregations := []*ActiveAggregation{
		{
			AggregationID:    "agg-001",
			AggregationName:  "job_performance_aggregation",
			AggregationType:  "statistical",
			Status:           "active",
			StartTime:        now.Add(-time.Hour * 2),
			EndTime:          now.Add(time.Hour * 1),
			EventsProcessed:  2500000,
			ResultsGenerated: 15000,
			ProcessingRate:   2500.5,
			QualityScore:     0.95,
			PerformanceMetrics: map[string]float64{
				"cpu_utilization":    0.75,
				"memory_utilization": 0.68,
				"throughput":         2500.5,
				"latency":            125.0,
			},
		},
		{
			AggregationID:    "agg-002",
			AggregationName:  "resource_utilization_aggregation",
			AggregationType:  "temporal",
			Status:           "healthy",
			StartTime:        now.Add(-time.Hour * 4),
			EndTime:          now.Add(time.Hour * 2),
			EventsProcessed:  5000000,
			ResultsGenerated: 25000,
			ProcessingRate:   4250.8,
			QualityScore:     0.92,
			PerformanceMetrics: map[string]float64{
				"cpu_utilization":    0.82,
				"memory_utilization": 0.74,
				"throughput":         4250.8,
				"latency":            98.5,
			},
		},
	}

	activeCorrelations := []*ActiveCorrelation{
		{
			CorrelationID:    "corr-001",
			CorrelationName:  "job_failure_correlation",
			CorrelationType:  "causal",
			Status:           "active",
			StartTime:        now.Add(-time.Hour * 3),
			EndTime:          now.Add(time.Hour * 1),
			EventsAnalyzed:   1500000,
			CorrelationsFound: 850,
			AnalysisDepth:    5,
			ConfidenceScore:  0.88,
			AccuracyMetrics: map[string]float64{
				"precision": 0.92,
				"recall":    0.85,
				"f1_score":  0.885,
				"accuracy":  0.90,
			},
		},
		{
			CorrelationID:    "corr-002",
			CorrelationName:  "performance_pattern_correlation",
			CorrelationType:  "pattern",
			Status:           "healthy",
			StartTime:        now.Add(-time.Hour * 6),
			EndTime:          now.Add(time.Hour * 3),
			EventsAnalyzed:   3000000,
			CorrelationsFound: 1250,
			AnalysisDepth:    7,
			ConfidenceScore:  0.91,
			AccuracyMetrics: map[string]float64{
				"precision": 0.94,
				"recall":    0.89,
				"f1_score":  0.915,
				"accuracy":  0.93,
			},
		},
	}

	client.On("GetAggregationConfiguration", mock.Anything).Return(&AggregationConfiguration{}, nil)
	client.On("GetCorrelationConfiguration", mock.Anything).Return(&CorrelationConfiguration{}, nil)
	client.On("GetActiveAggregations", mock.Anything).Return(activeAggregations, nil)
	client.On("GetActiveCorrelations", mock.Anything).Return(activeCorrelations, nil)
	client.On("GetAggregationMetrics", mock.Anything).Return(&AggregationMetrics{}, nil)
	client.On("GetCorrelationMetrics", mock.Anything).Return(&CorrelationMetrics{}, nil)

	ch := make(chan prometheus.Metric, 200)
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

func TestEventAggregationCorrelationCollector_Collect_SystemMetrics(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	aggregationMetrics := &AggregationMetrics{
		TotalAggregations:     500,
		ActiveAggregations:    25,
		CompletedAggregations: 450,
		FailedAggregations:    25,
		AverageProcessingTime: time.Second * 125,
		ThroughputRate:        2500.5,
		QualityScore:          0.95,
		ResourceUtilization: map[string]float64{
			"cpu":     0.75,
			"memory":  0.68,
			"storage": 0.55,
			"network": 0.42,
			"gpu":     0.85,
		},
	}

	correlationMetrics := &CorrelationMetrics{
		TotalCorrelations:     1200,
		ActiveCorrelations:    15,
		CompletedCorrelations: 1150,
		FailedCorrelations:    35,
		AverageAnalysisTime:   time.Second * 180,
		AccuracyRate:          0.92,
		ConfidenceScore:       0.88,
		ResourceUtilization: map[string]float64{
			"cpu":     0.82,
			"memory":  0.76,
			"storage": 0.62,
			"network": 0.48,
			"gpu":     0.91,
		},
	}

	client.On("GetAggregationConfiguration", mock.Anything).Return(&AggregationConfiguration{}, nil)
	client.On("GetCorrelationConfiguration", mock.Anything).Return(&CorrelationConfiguration{}, nil)
	client.On("GetActiveAggregations", mock.Anything).Return([]*ActiveAggregation{}, nil)
	client.On("GetActiveCorrelations", mock.Anything).Return([]*ActiveCorrelation{}, nil)
	client.On("GetAggregationMetrics", mock.Anything).Return(aggregationMetrics, nil)
	client.On("GetCorrelationMetrics", mock.Anything).Return(correlationMetrics, nil)

	ch := make(chan prometheus.Metric, 200)
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

func TestEventAggregationCorrelationCollector_CollectError(t *testing.T) {
	client := &MockEventAggregationCorrelationSLURMClient{}
	collector := NewEventAggregationCorrelationCollector(client)

	// Simulate errors from all client methods
	client.On("GetAggregationConfiguration", mock.Anything).Return((*AggregationConfiguration)(nil), assert.AnError)
	client.On("GetCorrelationConfiguration", mock.Anything).Return((*CorrelationConfiguration)(nil), assert.AnError)
	client.On("GetActiveAggregations", mock.Anything).Return([]*ActiveAggregation(nil), assert.AnError)
	client.On("GetActiveCorrelations", mock.Anything).Return([]*ActiveCorrelation(nil), assert.AnError)
	client.On("GetAggregationMetrics", mock.Anything).Return((*AggregationMetrics)(nil), assert.AnError)
	client.On("GetCorrelationMetrics", mock.Anything).Return((*CorrelationMetrics)(nil), assert.AnError)

	ch := make(chan prometheus.Metric, 200)
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