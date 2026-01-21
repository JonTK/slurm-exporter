package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAnomalyDetectionStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockAnomalyDetectionStreamingSLURMClient) StreamAnomalyEvents(ctx context.Context) (<-chan AnomalyEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan AnomalyEvent), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyStreamingConfiguration(ctx context.Context) (*AnomalyStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AnomalyStreamingConfiguration), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetActiveAnomalyStreams(ctx context.Context) ([]*ActiveAnomalyStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveAnomalyStream), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyEventHistory(ctx context.Context, anomalyID string, duration time.Duration) ([]*AnomalyEventRecord, error) {
	args := m.Called(ctx, anomalyID, duration)
	return args.Get(0).([]*AnomalyEventRecord), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyStreamingMetrics(ctx context.Context) (*AnomalyStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AnomalyStreamingMetrics), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyEventFilters(ctx context.Context) ([]*AnomalyEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*AnomalyEventFilter), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) ConfigureAnomalyStreaming(ctx context.Context, config *AnomalyStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyStreamingStatus(ctx context.Context) (*AnomalyStreamingStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AnomalyStreamingStatus), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyEventSubscriptions(ctx context.Context) ([]*AnomalyEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*AnomalyEventSubscription), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) ManageAnomalyEventSubscription(ctx context.Context, subscription *AnomalyEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyEventProcessingStats(ctx context.Context) (*AnomalyEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AnomalyEventProcessingStats), args.Error(1)
}

func (m *MockAnomalyDetectionStreamingSLURMClient) GetAnomalyStreamingPerformanceMetrics(ctx context.Context) (*AnomalyStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*AnomalyStreamingPerformanceMetrics), args.Error(1)
}

func TestNewAnomalyDetectionStreamingCollector(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.anomalyEvents)
	assert.NotNil(t, collector.activeAnomalyStreams)
	assert.NotNil(t, collector.anomalyScore)
	assert.NotNil(t, collector.anomalyConfidence)
	assert.NotNil(t, collector.anomalySeverity)
	assert.NotNil(t, collector.anomalyFrequency)
	assert.NotNil(t, collector.detectionLatency)
	assert.NotNil(t, collector.falsePositiveRate)
	assert.NotNil(t, collector.truePositiveRate)
	assert.NotNil(t, collector.modelAccuracy)
	assert.NotNil(t, collector.businessImpact)
	assert.NotNil(t, collector.riskScore)
	assert.NotNil(t, collector.mitigationEffectiveness)
	assert.NotNil(t, collector.costAvoidance)
	assert.NotNil(t, collector.incidentsPrevented)
	assert.NotNil(t, collector.downtimePrevented)
	assert.NotNil(t, collector.complianceScore)
	assert.NotNil(t, collector.automationRate)
	assert.NotNil(t, collector.responseTime)
	assert.NotNil(t, collector.streamingPerformance)
}

func TestAnomalyDetectionStreamingCollector_Describe(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

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

func TestAnomalyDetectionStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

	config := &AnomalyStreamingConfiguration{
		StreamingEnabled:           true,
		DetectionInterval:          time.Second * 30,
		AnalysisWindowSize:         time.Hour * 24,
		EventBufferSize:            10000,
		MaxConcurrentDetection:     50,
		StreamTimeout:              time.Hour,
		DetectionMethods:           []string{"statistical", "ml", "rule_based", "hybrid"},
		StatisticalMethods:         []string{"zscore", "mad", "iqr", "grubbs", "gesd"},
		MLModels:                   []string{"isolation_forest", "autoencoder", "lstm", "prophet", "arima"},
		RuleEngines:                []string{"drools", "easy_rules", "custom"},
		HybridApproaches:           []string{"ensemble", "voting", "stacking"},
		AnomalyTypes:               []string{"point", "contextual", "collective", "trend", "seasonal"},
		ResourceAnomalies:          true,
		PerformanceAnomalies:       true,
		SecurityAnomalies:          true,
		ComplianceAnomalies:        true,
		CostAnomalies:              true,
		UserBehaviorAnomalies:      true,
		JobPatternAnomalies:        true,
		NetworkAnomalies:           true,
		StorageAnomalies:           true,
		HardwareAnomalies:          true,
		SoftwareAnomalies:          true,
		ConfigurationAnomalies:     true,
		IntegrationAnomalies:       true,
		DataQualityAnomalies:       true,
		BusinessProcessAnomalies:   true,
		SensitivityLevels:          []string{"low", "medium", "high", "critical"},
		DefaultSensitivity:         "medium",
		AdaptiveSensitivity:        true,
		ContextualAnalysis:         true,
		SeasonalityDetection:       true,
		TrendAnalysis:              true,
		CorrelationAnalysis:        true,
		CausalityAnalysis:          true,
		RootCauseAnalysis:          true,
		ImpactAnalysis:             true,
		PredictiveAnalysis:         true,
		HistoricalComparison:       true,
		PeerComparison:             true,
		BaselineComparison:         true,
		ThresholdAdaptation:        true,
		LearningRate:               0.01,
		TrainingFrequency:          time.Hour * 6,
		ModelUpdateThreshold:       0.85,
		DataRetentionPeriod:        time.Hour * 24 * 90,
		FeatureEngineering:         true,
		FeatureSelection:           true,
		DimensionalityReduction:    true,
		DataPreprocessing:          true,
		OutlierRemoval:             true,
		NoiseReduction:             true,
		DataNormalization:          true,
		DataAugmentation:           true,
		RealTimeScoring:            true,
		BatchScoring:               true,
		StreamProcessing:           true,
		ParallelProcessing:         true,
		GPUAcceleration:            true,
		CachingEnabled:             true,
		CompressionEnabled:         true,
		EncryptionEnabled:          true,
		AutoRemediation:            true,
		PlaybookIntegration:        true,
		TicketingIntegration:       true,
		NotificationChannels:       []string{"email", "slack", "pagerduty", "webhook"},
		AlertThresholds: map[string]float64{
			"anomaly_score":  0.7,
			"confidence":     0.8,
			"severity":       0.6,
			"business_impact": 0.5,
		},
		AlertSuppressionRules:      []string{"duplicate", "flapping", "scheduled", "acknowledged"},
		AlertGroupingEnabled:       true,
		AlertCorrelationEnabled:    true,
		AlertEnrichmentEnabled:     true,
		AlertPrioritization:        true,
		FalsePositiveHandling:      true,
		FeedbackLoopEnabled:        true,
		ContinuousLearning:         true,
		ModelExplainability:        true,
		AuditLogging:               true,
		ComplianceTracking:         true,
		PerformanceMonitoring:      true,
		CostTracking:               true,
		ROICalculation:             true,
		BusinessValueTracking:      true,
		DashboardIntegration:       true,
		ReportingEnabled:           true,
		APIEnabled:                 true,
		WebhookEnabled:             true,
		ExportFormats:              []string{"json", "csv", "parquet", "avro"},
		DataGovernance:             true,
		DataLineage:                true,
		DataQualityChecks:          true,
		PrivacyProtection:          true,
		AccessControl:              true,
		MultiTenancy:               true,
		ScalabilityMode:            "horizontal",
		HighAvailability:           true,
		DisasterRecovery:           true,
		BackupEnabled:              true,
	}

	client.On("GetAnomalyStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveAnomalyStreams", mock.Anything).Return([]*ActiveAnomalyStream{}, nil)
	client.On("GetAnomalyStreamingMetrics", mock.Anything).Return(&AnomalyStreamingMetrics{}, nil)
	client.On("GetAnomalyStreamingStatus", mock.Anything).Return(&AnomalyStreamingStatus{}, nil)
	client.On("GetAnomalyEventSubscriptions", mock.Anything).Return([]*AnomalyEventSubscription{}, nil)
	client.On("GetAnomalyEventFilters", mock.Anything).Return([]*AnomalyEventFilter{}, nil)
	client.On("GetAnomalyEventProcessingStats", mock.Anything).Return(&AnomalyEventProcessingStats{}, nil)
	client.On("GetAnomalyStreamingPerformanceMetrics", mock.Anything).Return(&AnomalyStreamingPerformanceMetrics{}, nil)

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

func TestAnomalyDetectionStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveAnomalyStream{
		{
			StreamID:                   "anomaly-stream-001",
			StreamName:                 "comprehensive_anomaly_detector",
			StreamType:                 "multi_method",
			CreatedAt:                  now.Add(-time.Hour * 48),
			LastActivityAt:             now,
			Status:                     "healthy",
			DetectionWindowStart:       now.Add(-time.Hour * 24),
			DetectionWindowEnd:         now,
			EventsAnalyzed:             5000000,
			AnomaliesDetected:          2500,
			CriticalAnomalies:          50,
			HighSeverityAnomalies:      250,
			MediumSeverityAnomalies:    800,
			LowSeverityAnomalies:       1400,
			DetectionRate:              0.0005,
			FalsePositives:             125,
			TruePositives:              2375,
			FalseNegatives:             75,
			TrueNegatives:              4997425,
			Precision:                  0.95,
			Recall:                     0.97,
			F1Score:                    0.96,
			Accuracy:                   0.9996,
			ModelPerformance:           0.94,
			ProcessingRate:             5000.5,
			DetectionLatency:           time.Millisecond * 150,
			ResourceUtilization:        0.72,
			MemoryUsage:                8589934592, // 8GB
			CPUUsage:                   0.65,
			GPUUsage:                   0.85,
			NetworkBandwidth:           125.5,
			ActiveModels:               8,
			ModelVersions: map[string]string{
				"isolation_forest": "v3.2.1",
				"lstm":             "v2.8.5",
				"autoencoder":      "v4.1.0",
			},
			FeatureImportance: map[string]float64{
				"cpu_usage_pattern":    0.85,
				"memory_spike":         0.78,
				"job_failure_rate":     0.72,
				"network_anomaly":      0.65,
				"user_behavior_change": 0.58,
			},
			BusinessValueProtected:     5000000.0,
			CostAvoidance:              1250000.0,
			IncidentsPrevented:         125,
			DowntimePrevented:          time.Hour * 72,
			ComplianceViolationsCaught: 15,
			SecurityThreatsCaught:      8,
			PerformanceIssuesCaught:    450,
			ResourceWasteCaught:        325,
			AutoRemediations:           850,
			AlertsGenerated:            1250,
			TicketsCreated:             125,
			EscalationsTriggered:       25,
			UserNotifications:          2500,
			LearningProgress:           0.92,
			AdaptationRate:             0.15,
			DriftDetected:              false,
			LastModelUpdate:            now.Add(-time.Hour * 12),
			NextScheduledUpdate:        now.Add(time.Hour * 6),
		},
		{
			StreamID:                   "anomaly-stream-002",
			StreamName:                 "security_anomaly_detector",
			StreamType:                 "security_focused",
			CreatedAt:                  now.Add(-time.Hour * 24),
			LastActivityAt:             now.Add(-time.Minute * 5),
			Status:                     "degraded",
			DetectionWindowStart:       now.Add(-time.Hour * 12),
			DetectionWindowEnd:         now.Add(-time.Minute * 5),
			EventsAnalyzed:             2500000,
			AnomaliesDetected:          850,
			CriticalAnomalies:          25,
			HighSeverityAnomalies:      150,
			MediumSeverityAnomalies:    350,
			LowSeverityAnomalies:       325,
			DetectionRate:              0.00034,
			FalsePositives:             85,
			TruePositives:              765,
			FalseNegatives:             45,
			TrueNegatives:              2499105,
			Precision:                  0.90,
			Recall:                     0.944,
			F1Score:                    0.922,
			Accuracy:                   0.999,
			ModelPerformance:           0.88,
			ProcessingRate:             2500.3,
			DetectionLatency:           time.Millisecond * 250,
			ResourceUtilization:        0.82,
			MemoryUsage:                6442450944, // 6GB
			CPUUsage:                   0.75,
			GPUUsage:                   0.0, // No GPU
			NetworkBandwidth:           85.3,
			ActiveModels:               5,
			ModelVersions: map[string]string{
				"security_rules":   "v5.3.2",
				"behavior_model":   "v3.1.0",
				"threat_detection": "v4.2.1",
			},
			FeatureImportance: map[string]float64{
				"auth_failures":       0.92,
				"privilege_escalation": 0.88,
				"unusual_access":      0.85,
				"data_exfiltration":   0.82,
				"port_scanning":       0.75,
			},
			BusinessValueProtected:     3500000.0,
			CostAvoidance:              850000.0,
			IncidentsPrevented:         45,
			DowntimePrevented:          time.Hour * 24,
			ComplianceViolationsCaught: 8,
			SecurityThreatsCaught:      25,
			PerformanceIssuesCaught:    150,
			ResourceWasteCaught:        85,
			AutoRemediations:           325,
			AlertsGenerated:            450,
			TicketsCreated:             85,
			EscalationsTriggered:       15,
			UserNotifications:          850,
			LearningProgress:           0.85,
			AdaptationRate:             0.12,
			DriftDetected:              true,
			LastModelUpdate:            now.Add(-time.Hour * 24),
			NextScheduledUpdate:        now.Add(time.Hour * 2),
		},
	}

	client.On("GetAnomalyStreamingConfiguration", mock.Anything).Return(&AnomalyStreamingConfiguration{}, nil)
	client.On("GetActiveAnomalyStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetAnomalyStreamingMetrics", mock.Anything).Return(&AnomalyStreamingMetrics{}, nil)
	client.On("GetAnomalyStreamingStatus", mock.Anything).Return(&AnomalyStreamingStatus{}, nil)
	client.On("GetAnomalyEventSubscriptions", mock.Anything).Return([]*AnomalyEventSubscription{}, nil)
	client.On("GetAnomalyEventFilters", mock.Anything).Return([]*AnomalyEventFilter{}, nil)
	client.On("GetAnomalyEventProcessingStats", mock.Anything).Return(&AnomalyEventProcessingStats{}, nil)
	client.On("GetAnomalyStreamingPerformanceMetrics", mock.Anything).Return(&AnomalyStreamingPerformanceMetrics{}, nil)

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

func TestAnomalyDetectionStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

	streamingMetrics := &AnomalyStreamingMetrics{
		TotalStreams:                15,
		ActiveStreams:               12,
		HealthyStreams:              10,
		DegradedStreams:             2,
		StreamEfficiency:            0.92,
		StreamReliability:           0.98,
		TotalAnomaliesDetected:      125000,
		UniqueAnomalies:             8500,
		RecurringAnomalies:          5500,
		NewAnomalies:                2500,
		CriticalAnomalies:           1250,
		HighSeverityAnomalies:       5500,
		MediumSeverityAnomalies:     25000,
		LowSeverityAnomalies:        93250,
		EventsAnalyzed:              50000000,
		AnalysisRate:                5000.5,
		DetectionLatency:            time.Millisecond * 125,
		DetectionAccuracy:           0.985,
		FalsePositiveRate:           0.05,
		FalseNegativeRate:           0.03,
		PrecisionScore:              0.95,
		RecallScore:                 0.97,
		F1Score:                     0.96,
		AverageAnomalyScore:         0.72,
		AverageConfidence:           0.88,
		ModelAccuracy:               0.93,
		ModelPrecision:              0.94,
		ModelRecall:                 0.92,
		EnsemblePerformance:         0.96,
		CPUUtilization:              0.68,
		MemoryUtilization:           0.75,
		GPUUtilization:              0.85,
		StorageUtilization:          0.62,
		NetworkUtilization:          0.55,
		ResourceEfficiency:          0.88,
		TotalBusinessValue:          25000000.0,
		TotalCostAvoidance:          6500000.0,
		TotalIncidentsPrevented:     850,
		TotalDowntimePrevented:      time.Hour * 720,
		ComplianceViolationsCaught:  125,
		SecurityThreatsCaught:       85,
		PerformanceIssuesCaught:     2500,
		ResourceWasteCaught:         1850,
		AutoRemediationSuccess:      0.92,
		AlertAccuracy:               0.88,
		ResponseTimeAverage:         time.Minute * 5,
		MitigationEffectiveness:     0.85,
		UserSatisfaction:            0.91,
		SystemUptime:                0.9995,
		DataQuality:                 0.97,
		ModelDrift:                  0.08,
		LearningRate:                0.15,
		AdaptationSuccess:           0.93,
		PredictiveAccuracy:          0.86,
		PreventiveActions:           3500,
		CostSavingsRealized:         5500000.0,
		ROI:                         4.2,
		ComplianceScore:             0.98,
		RiskReduction:               0.65,
		OperationalEfficiency:       0.89,
		InnovationIndex:             0.75,
	}

	client.On("GetAnomalyStreamingConfiguration", mock.Anything).Return(&AnomalyStreamingConfiguration{}, nil)
	client.On("GetActiveAnomalyStreams", mock.Anything).Return([]*ActiveAnomalyStream{}, nil)
	client.On("GetAnomalyStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetAnomalyStreamingStatus", mock.Anything).Return(&AnomalyStreamingStatus{}, nil)
	client.On("GetAnomalyEventSubscriptions", mock.Anything).Return([]*AnomalyEventSubscription{}, nil)
	client.On("GetAnomalyEventFilters", mock.Anything).Return([]*AnomalyEventFilter{}, nil)
	client.On("GetAnomalyEventProcessingStats", mock.Anything).Return(&AnomalyEventProcessingStats{}, nil)
	client.On("GetAnomalyStreamingPerformanceMetrics", mock.Anything).Return(&AnomalyStreamingPerformanceMetrics{}, nil)

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

func TestAnomalyDetectionStreamingCollector_CollectError(t *testing.T) {
	client := &MockAnomalyDetectionStreamingSLURMClient{}
	collector := NewAnomalyDetectionStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetAnomalyStreamingConfiguration", mock.Anything).Return((*AnomalyStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveAnomalyStreams", mock.Anything).Return([]*ActiveAnomalyStream(nil), assert.AnError)
	client.On("GetAnomalyStreamingMetrics", mock.Anything).Return((*AnomalyStreamingMetrics)(nil), assert.AnError)
	client.On("GetAnomalyStreamingStatus", mock.Anything).Return((*AnomalyStreamingStatus)(nil), assert.AnError)
	client.On("GetAnomalyEventSubscriptions", mock.Anything).Return([]*AnomalyEventSubscription(nil), assert.AnError)
	client.On("GetAnomalyEventFilters", mock.Anything).Return([]*AnomalyEventFilter(nil), assert.AnError)
	client.On("GetAnomalyEventProcessingStats", mock.Anything).Return((*AnomalyEventProcessingStats)(nil), assert.AnError)
	client.On("GetAnomalyStreamingPerformanceMetrics", mock.Anything).Return((*AnomalyStreamingPerformanceMetrics)(nil), assert.AnError)

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