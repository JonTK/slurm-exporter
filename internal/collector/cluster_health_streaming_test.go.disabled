package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClusterHealthStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockClusterHealthStreamingSLURMClient) StreamClusterHealthEvents(ctx context.Context) (<-chan ClusterHealthEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan ClusterHealthEvent), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthStreamingConfiguration(ctx context.Context) (*HealthStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HealthStreamingConfiguration), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetActiveHealthStreams(ctx context.Context) ([]*ActiveHealthStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveHealthStream), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthEventHistory(ctx context.Context, componentID string, duration time.Duration) ([]*HealthEvent, error) {
	args := m.Called(ctx, componentID, duration)
	return args.Get(0).([]*HealthEvent), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthStreamingMetrics(ctx context.Context) (*HealthStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HealthStreamingMetrics), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthEventFilters(ctx context.Context) ([]*HealthEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*HealthEventFilter), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) ConfigureHealthStreaming(ctx context.Context, config *HealthStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthStreamingStatus(ctx context.Context) (*HealthStreamingStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HealthStreamingStatus), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthEventSubscriptions(ctx context.Context) ([]*HealthEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*HealthEventSubscription), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) ManageHealthEventSubscription(ctx context.Context, subscription *HealthEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthEventProcessingStats(ctx context.Context) (*HealthEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HealthEventProcessingStats), args.Error(1)
}

func (m *MockClusterHealthStreamingSLURMClient) GetHealthStreamingPerformanceMetrics(ctx context.Context) (*HealthStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*HealthStreamingPerformanceMetrics), args.Error(1)
}

func TestNewClusterHealthStreamingCollector(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.healthEvents)
	assert.NotNil(t, collector.activeHealthStreams)
	assert.NotNil(t, collector.clusterHealthScore)
	assert.NotNil(t, collector.componentHealth)
	assert.NotNil(t, collector.nodeAvailability)
	assert.NotNil(t, collector.jobHealthMetrics)
	assert.NotNil(t, collector.resourceUtilization)
	assert.NotNil(t, collector.serviceAvailability)
	assert.NotNil(t, collector.anomalyDetection)
	assert.NotNil(t, collector.predictiveHealth)
	assert.NotNil(t, collector.businessImpact)
	assert.NotNil(t, collector.recoveryMetrics)
	assert.NotNil(t, collector.streamingPerformance)
	assert.NotNil(t, collector.detectionAccuracy)
	assert.NotNil(t, collector.automationEffectiveness)
	assert.NotNil(t, collector.costOptimization)
	assert.NotNil(t, collector.complianceStatus)
	assert.NotNil(t, collector.riskAssessment)
	assert.NotNil(t, collector.capacityForecast)
}

func TestClusterHealthStreamingCollector_Describe(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

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

func TestClusterHealthStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

	config := &HealthStreamingConfiguration{
		StreamingEnabled:         true,
		HealthCheckInterval:      time.Minute * 5,
		EventBufferSize:          10000,
		EventBatchSize:           100,
		MaxConcurrentStreams:     50,
		StreamTimeout:            time.Minute * 30,
		ComponentMonitoring:      true,
		NodeMonitoring:           true,
		JobMonitoring:            true,
		QueueMonitoring:          true,
		ResourceMonitoring:       true,
		NetworkMonitoring:        true,
		StorageMonitoring:        true,
		ServiceMonitoring:        true,
		CriticalThreshold:        0.9,
		WarningThreshold:         0.7,
		InfoThreshold:            0.5,
		AnomalyThreshold:         0.8,
		PredictionThreshold:      0.75,
		ImpactThreshold:          0.6,
		RecoveryThreshold:        0.85,
		FaultDetection:           true,
		AnomalyDetection:         true,
		PredictiveAnalysis:       true,
		TrendAnalysis:            true,
		PatternRecognition:       true,
		CorrelationAnalysis:      true,
		RootCauseAnalysis:        true,
		CompressionEnabled:       true,
		EncryptionEnabled:        true,
		DeduplicationEnabled:     true,
		FilteringEnabled:         true,
		AggregationEnabled:       true,
		SamplingEnabled:          true,
		CachingEnabled:           true,
		GuaranteedDelivery:       true,
		OrderPreservation:        true,
		DuplicateHandling:        true,
		RetryEnabled:             true,
		CircuitBreakerEnabled:    true,
		BackpressureEnabled:      true,
		FailoverEnabled:          true,
		AlertingEnabled:          true,
		NotificationEnabled:      true,
		WebhookEnabled:           true,
		MachineLearning:          true,
		AutoRemediation:          true,
		CapacityPlanning:         true,
		CostOptimization:         true,
		ComplianceMonitoring:     true,
		SecurityMonitoring:       true,
		PerformanceOptimization:  true,
		SLATracking:              true,
		BusinessMetrics:          true,
		UserExperience:           true,
		AutoScaling:              true,
		AutoHealing:              true,
		AutoOptimization:         true,
		AutoFailover:             true,
		AuditLogging:             true,
		ComplianceReporting:      true,
		DataRetention:            time.Hour * 24 * 30, // 30 days
		DataPrivacy:              true,
		Encryption:               true,
		AccessControl:            true,
		ChangeTracking:           true,
	}

	client.On("GetHealthStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveHealthStreams", mock.Anything).Return([]*ActiveHealthStream{}, nil)
	client.On("GetHealthStreamingMetrics", mock.Anything).Return(&HealthStreamingMetrics{}, nil)
	client.On("GetHealthStreamingStatus", mock.Anything).Return(&HealthStreamingStatus{}, nil)
	client.On("GetHealthEventSubscriptions", mock.Anything).Return([]*HealthEventSubscription{}, nil)
	client.On("GetHealthEventFilters", mock.Anything).Return([]*HealthEventFilter{}, nil)
	client.On("GetHealthEventProcessingStats", mock.Anything).Return(&HealthEventProcessingStats{}, nil)
	client.On("GetHealthStreamingPerformanceMetrics", mock.Anything).Return(&HealthStreamingPerformanceMetrics{}, nil)

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

func TestClusterHealthStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveHealthStream{
		{
			StreamID:              "health-stream-001",
			StreamName:            "primary_health_monitor",
			StreamType:            "comprehensive",
			CreatedAt:             now.Add(-time.Hour * 24),
			LastUpdateAt:          now,
			Status:                "healthy",
			ConsumerID:            "monitor-001",
			ConsumerName:          "Primary Health Monitor",
			ConsumerType:          "monitoring",
			ConsumerEndpoint:      "https://monitor1.example.com:8443",
			LastHeartbeat:         now.Add(-time.Second * 15),
			ConnectionHealth:      "excellent",
			EventsDelivered:       1500000,
			EventsPending:         150,
			EventsFiltered:        50000,
			EventsAggregated:      25000,
			BytesTransferred:      10737418240, // 10GB
			CompressionRatio:      0.65,
			Throughput:            1250.5,
			Latency:               time.Millisecond * 25,
			ErrorRate:             0.0001,
			ProcessingRate:        1200.3,
			DeliveryRate:          1195.8,
			BacklogSize:           150,
			DataQuality:           0.995,
			DataCompleteness:      0.998,
			DataAccuracy:          0.997,
			DataTimeliness:        0.996,
			SignalToNoise:         0.92,
			FalsePositiveRate:     0.02,
			ComponentsCovered:     500,
			HealthChecksPerformed: 8640000,
			IssuesDetected:        1250,
			IssuesResolved:        1200,
			CriticalAlerts:        15,
			MTBF:                  time.Hour * 720,
			MTTR:                  time.Minute * 15,
			BusinessValue:         2500000.0,
			CostSavings:           850000.0,
			DowntimePrevented:     time.Hour * 48,
			IncidentsPrevented:    125,
			SLACompliance:         0.999,
			CustomerSatisfaction:  0.95,
			ROI:                   3.5,
		},
		{
			StreamID:              "health-stream-002",
			StreamName:            "critical_health_monitor",
			StreamType:            "critical",
			CreatedAt:             now.Add(-time.Hour * 12),
			LastUpdateAt:          now.Add(-time.Minute * 1),
			Status:                "degraded",
			ConsumerID:            "monitor-002",
			ConsumerName:          "Critical Health Monitor",
			ConsumerType:          "monitoring",
			ConsumerEndpoint:      "https://monitor2.example.com:8444",
			LastHeartbeat:         now.Add(-time.Minute * 1),
			ConnectionHealth:      "fair",
			EventsDelivered:       750000,
			EventsPending:         500,
			EventsFiltered:        25000,
			EventsAggregated:      12500,
			BytesTransferred:      5368709120, // 5GB
			CompressionRatio:      0.60,
			Throughput:            625.2,
			Latency:               time.Millisecond * 45,
			ErrorRate:             0.001,
			ProcessingRate:        615.8,
			DeliveryRate:          610.5,
			BacklogSize:           500,
			DataQuality:           0.98,
			DataCompleteness:      0.99,
			DataAccuracy:          0.985,
			DataTimeliness:        0.975,
			SignalToNoise:         0.88,
			FalsePositiveRate:     0.04,
			ComponentsCovered:     250,
			HealthChecksPerformed: 4320000,
			IssuesDetected:        850,
			IssuesResolved:        800,
			CriticalAlerts:        25,
			MTBF:                  time.Hour * 360,
			MTTR:                  time.Minute * 25,
			BusinessValue:         1800000.0,
			CostSavings:           650000.0,
			DowntimePrevented:     time.Hour * 36,
			IncidentsPrevented:    95,
			SLACompliance:         0.995,
			CustomerSatisfaction:  0.92,
			ROI:                   2.8,
		},
	}

	client.On("GetHealthStreamingConfiguration", mock.Anything).Return(&HealthStreamingConfiguration{}, nil)
	client.On("GetActiveHealthStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetHealthStreamingMetrics", mock.Anything).Return(&HealthStreamingMetrics{}, nil)
	client.On("GetHealthStreamingStatus", mock.Anything).Return(&HealthStreamingStatus{}, nil)
	client.On("GetHealthEventSubscriptions", mock.Anything).Return([]*HealthEventSubscription{}, nil)
	client.On("GetHealthEventFilters", mock.Anything).Return([]*HealthEventFilter{}, nil)
	client.On("GetHealthEventProcessingStats", mock.Anything).Return(&HealthEventProcessingStats{}, nil)
	client.On("GetHealthStreamingPerformanceMetrics", mock.Anything).Return(&HealthStreamingPerformanceMetrics{}, nil)

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

func TestClusterHealthStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

	streamingMetrics := &HealthStreamingMetrics{
		TotalStreams:              25,
		ActiveStreams:             20,
		HealthyStreams:            18,
		DegradedStreams:           2,
		StreamUtilization:         0.85,
		StreamEfficiency:          0.92,
		TotalEventsProcessed:      25000000,
		CriticalEvents:            2500,
		WarningEvents:             12500,
		InfoEvents:                250000,
		EventsPerSecond:           2500.5,
		EventProcessingLatency:    time.Millisecond * 30,
		ClusterHealthScore:        0.95,
		ComponentHealthScore:      0.93,
		ServiceHealthScore:        0.96,
		OverallAvailability:       0.9995,
		OverallReliability:        0.998,
		OverallPerformance:        0.94,
		IssuesDetected:            5000,
		IssuesResolved:            4800,
		AnomaliesDetected:         350,
		PredictionsGenerated:      1200,
		FalsePositives:            50,
		FalseNegatives:            25,
		DetectionAccuracy:         0.985,
		ResponseTime:              time.Millisecond * 250,
		ProcessingThroughput:      2450.5,
		ResourceUtilization:       0.72,
		NetworkUtilization:        0.65,
		StorageUtilization:        0.68,
		ComputeUtilization:        0.75,
		DowntimeAverted:           time.Hour * 120,
		CostSavingsRealized:       2500000.0,
		SLAAttainment:             0.998,
		CustomerImpactMinimized:   0.95,
		ProductivityMaintained:    0.97,
		RevenueProtected:          15000000.0,
		AutoRemediations:          850,
		AutoScalingEvents:         125,
		AutoFailovers:             15,
		AutoOptimizations:         450,
		AutomationSuccessRate:     0.96,
		ManualInterventions:       85,
		PredictiveAccuracy:        0.88,
		PreventiveActions:         650,
		PlannedMaintenances:       45,
		UnplannedOutages:          3,
		MTTF:                      time.Hour * 2160, // 90 days
		PredictedFailures:         125,
		DataQuality:               0.97,
		MonitoringCoverage:        0.95,
		AlertQuality:              0.93,
		RecommendationQuality:     0.91,
		ReportingQuality:          0.94,
		AnalysisDepth:             0.89,
		CapacityHeadroom:          0.35,
		GrowthRate:                0.15,
		ScalingEfficiency:         0.92,
		ResourceEfficiency:        0.88,
		OptimizationPotential:     0.25,
		WasteReduction:            0.22,
	}

	client.On("GetHealthStreamingConfiguration", mock.Anything).Return(&HealthStreamingConfiguration{}, nil)
	client.On("GetActiveHealthStreams", mock.Anything).Return([]*ActiveHealthStream{}, nil)
	client.On("GetHealthStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetHealthStreamingStatus", mock.Anything).Return(&HealthStreamingStatus{}, nil)
	client.On("GetHealthEventSubscriptions", mock.Anything).Return([]*HealthEventSubscription{}, nil)
	client.On("GetHealthEventFilters", mock.Anything).Return([]*HealthEventFilter{}, nil)
	client.On("GetHealthEventProcessingStats", mock.Anything).Return(&HealthEventProcessingStats{}, nil)
	client.On("GetHealthStreamingPerformanceMetrics", mock.Anything).Return(&HealthStreamingPerformanceMetrics{}, nil)

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

func TestClusterHealthStreamingCollector_CollectError(t *testing.T) {
	client := &MockClusterHealthStreamingSLURMClient{}
	collector := NewClusterHealthStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetHealthStreamingConfiguration", mock.Anything).Return((*HealthStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveHealthStreams", mock.Anything).Return([]*ActiveHealthStream(nil), assert.AnError)
	client.On("GetHealthStreamingMetrics", mock.Anything).Return((*HealthStreamingMetrics)(nil), assert.AnError)
	client.On("GetHealthStreamingStatus", mock.Anything).Return((*HealthStreamingStatus)(nil), assert.AnError)
	client.On("GetHealthEventSubscriptions", mock.Anything).Return([]*HealthEventSubscription(nil), assert.AnError)
	client.On("GetHealthEventFilters", mock.Anything).Return([]*HealthEventFilter(nil), assert.AnError)
	client.On("GetHealthEventProcessingStats", mock.Anything).Return((*HealthEventProcessingStats)(nil), assert.AnError)
	client.On("GetHealthStreamingPerformanceMetrics", mock.Anything).Return((*HealthStreamingPerformanceMetrics)(nil), assert.AnError)

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