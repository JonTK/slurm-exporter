package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJobSchedulingStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockJobSchedulingStreamingSLURMClient) StreamJobSchedulingEvents(ctx context.Context) (<-chan JobSchedulingEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan JobSchedulingEvent), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingStreamingConfiguration(ctx context.Context) (*SchedulingStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SchedulingStreamingConfiguration), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetActiveSchedulingStreams(ctx context.Context) ([]*ActiveSchedulingStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveSchedulingStream), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingEventHistory(ctx context.Context, jobID string, duration time.Duration) ([]*SchedulingEvent, error) {
	args := m.Called(ctx, jobID, duration)
	return args.Get(0).([]*SchedulingEvent), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingStreamingMetrics(ctx context.Context) (*SchedulingStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SchedulingStreamingMetrics), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingEventFilters(ctx context.Context) ([]*SchedulingEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*SchedulingEventFilter), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) ConfigureSchedulingStreaming(ctx context.Context, config *SchedulingStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingStreamingHealthStatus(ctx context.Context) (*SchedulingStreamingHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SchedulingStreamingHealthStatus), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingEventSubscriptions(ctx context.Context) ([]*SchedulingEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*SchedulingEventSubscription), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) ManageSchedulingEventSubscription(ctx context.Context, subscription *SchedulingEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingEventProcessingStats(ctx context.Context) (*SchedulingEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SchedulingEventProcessingStats), args.Error(1)
}

func (m *MockJobSchedulingStreamingSLURMClient) GetSchedulingStreamingPerformanceMetrics(ctx context.Context) (*SchedulingStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SchedulingStreamingPerformanceMetrics), args.Error(1)
}

func TestNewJobSchedulingStreamingCollector(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.schedulingEvents)
	assert.NotNil(t, collector.activeSchedulingStreams)
	assert.NotNil(t, collector.streamingHealthScore)
	assert.NotNil(t, collector.schedulingLatency)
	assert.NotNil(t, collector.schedulingThroughput)
	assert.NotNil(t, collector.schedulingEfficiency)
	assert.NotNil(t, collector.decisionConfidence)
	assert.NotNil(t, collector.resourceAllocationScore)
	assert.NotNil(t, collector.queueOptimization)
	assert.NotNil(t, collector.backfillEffectiveness)
	assert.NotNil(t, collector.preemptionImpact)
	assert.NotNil(t, collector.fairnessScore)
	assert.NotNil(t, collector.businessValueDelivered)
	assert.NotNil(t, collector.slaCompliance)
	assert.NotNil(t, collector.costEfficiency)
	assert.NotNil(t, collector.optimizationGains)
	assert.NotNil(t, collector.anomalyDetectionRate)
	assert.NotNil(t, collector.predictionAccuracy)
	assert.NotNil(t, collector.systemUtilization)
}

func TestJobSchedulingStreamingCollector_Describe(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

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

func TestJobSchedulingStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

	config := &SchedulingStreamingConfiguration{
		StreamingEnabled:        true,
		EventBufferSize:         10000,
		EventBatchSize:          100,
		MaxConcurrentStreams:    50,
		CompressionEnabled:      true,
		CompressionLevel:        6,
		EncryptionEnabled:       true,
		BatchProcessing:         true,
		ParallelProcessing:      true,
		EventFiltering:          true,
		EventRouting:            true,
		PriorityQueuing:         true,
		GuaranteedDelivery:      true,
		OrderingGuaranteed:      true,
		DuplicateDetection:      true,
		MaxRetries:              5,
		BackoffMultiplier:       2.0,
		MetricsEnabled:          true,
		TracingEnabled:          true,
		ProfilingEnabled:        false,
		SamplingRate:            0.1,
		MaxMemoryUsage:          8589934592, // 8GB
		MaxCPUUsage:             0.8,
		MaxBandwidth:            1073741824, // 1GB/s
		MaxEventsPerSecond:      10000,
		BackpressureThreshold:   5000,
		CircuitBreakerThreshold: 0.5,
		AuthenticationRequired:  true,
		AuthorizationRequired:   true,
		AuditLoggingEnabled:     true,
		EventEnrichment:         true,
		CorrelationEnabled:      true,
		AdaptiveStreaming:       true,
		MLPrediction:            true,
		AnomalyDetection:        true,
	}

	client.On("GetSchedulingStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveSchedulingStreams", mock.Anything).Return([]*ActiveSchedulingStream{}, nil)
	client.On("GetSchedulingStreamingMetrics", mock.Anything).Return(&SchedulingStreamingMetrics{}, nil)
	client.On("GetSchedulingStreamingHealthStatus", mock.Anything).Return(&SchedulingStreamingHealthStatus{}, nil)
	client.On("GetSchedulingEventSubscriptions", mock.Anything).Return([]*SchedulingEventSubscription{}, nil)
	client.On("GetSchedulingEventFilters", mock.Anything).Return([]*SchedulingEventFilter{}, nil)
	client.On("GetSchedulingEventProcessingStats", mock.Anything).Return(&SchedulingEventProcessingStats{}, nil)
	client.On("GetSchedulingStreamingPerformanceMetrics", mock.Anything).Return(&SchedulingStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 100)
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

func TestJobSchedulingStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveSchedulingStream{
		{
			StreamID:             "sched-stream-001",
			StreamName:           "primary_scheduler",
			StreamType:           "backfill",
			CreatedAt:            now.Add(-time.Hour * 2),
			LastActivityAt:       now,
			Status:               "healthy",
			ConsumerID:           "scheduler-001",
			ConsumerName:         "Primary Scheduler",
			ConsumerType:         "scheduler",
			ConsumerEndpoint:     "tcp://scheduler1.example.com:5555",
			ConsumerHealth:       "healthy",
			LastHeartbeat:        now.Add(-time.Second * 30),
			EventsDelivered:      125000,
			EventsPending:        250,
			EventsFailed:         12,
			EventsFiltered:       3500,
			BytesTransferred:     1073741824, // 1GB
			AverageLatency:       time.Millisecond * 15,
			Throughput:           850.5,
			ErrorRate:            0.0001,
			SuccessRate:          0.9999,
			ProcessingRate:       825.3,
			BackpressureActive:   false,
			ResourceUtilization:  0.65,
			DeliveryGuarantee:    "at-least-once",
			OrderingGuarantee:    "strict",
			DuplicateRate:        0.001,
			DataQuality:          0.995,
			EnrichmentQuality:    0.98,
			CorrelationAccuracy:  0.97,
			BusinessValue:        125000.50,
			CostPerEvent:         0.002,
			ROI:                  15.5,
			SLACompliance:        0.999,
			CustomerSatisfaction: 0.95,
			ImpactScore:          0.88,
			OptimizationEnabled:  true,
			OptimizationGains:    0.25,
			ResourceSavings:      0.18,
			CostSavings:          0.22,
			EfficiencyGains:      0.28,
			PerformanceGains:     0.30,
		},
		{
			StreamID:             "sched-stream-002",
			StreamName:           "priority_scheduler",
			StreamType:           "priority",
			CreatedAt:            now.Add(-time.Hour * 1),
			LastActivityAt:       now.Add(-time.Minute * 2),
			Status:               "degraded",
			ConsumerID:           "scheduler-002",
			ConsumerName:         "Priority Scheduler",
			ConsumerType:         "scheduler",
			ConsumerEndpoint:     "tcp://scheduler2.example.com:5556",
			ConsumerHealth:       "warning",
			LastHeartbeat:        now.Add(-time.Minute * 2),
			EventsDelivered:      45000,
			EventsPending:        500,
			EventsFailed:         45,
			EventsFiltered:       1200,
			BytesTransferred:     536870912, // 512MB
			AverageLatency:       time.Millisecond * 35,
			Throughput:           425.2,
			ErrorRate:            0.001,
			SuccessRate:          0.999,
			ProcessingRate:       415.8,
			BackpressureActive:   true,
			ResourceUtilization:  0.85,
			DeliveryGuarantee:    "at-least-once",
			OrderingGuarantee:    "relaxed",
			DuplicateRate:        0.002,
			DataQuality:          0.99,
			EnrichmentQuality:    0.96,
			CorrelationAccuracy:  0.94,
			BusinessValue:        85000.25,
			CostPerEvent:         0.003,
			ROI:                  12.3,
			SLACompliance:        0.995,
			CustomerSatisfaction: 0.92,
			ImpactScore:          0.82,
			OptimizationEnabled:  true,
			OptimizationGains:    0.18,
			ResourceSavings:      0.12,
			CostSavings:          0.15,
			EfficiencyGains:      0.20,
			PerformanceGains:     0.22,
		},
	}

	client.On("GetSchedulingStreamingConfiguration", mock.Anything).Return(&SchedulingStreamingConfiguration{}, nil)
	client.On("GetActiveSchedulingStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetSchedulingStreamingMetrics", mock.Anything).Return(&SchedulingStreamingMetrics{}, nil)
	client.On("GetSchedulingStreamingHealthStatus", mock.Anything).Return(&SchedulingStreamingHealthStatus{}, nil)
	client.On("GetSchedulingEventSubscriptions", mock.Anything).Return([]*SchedulingEventSubscription{}, nil)
	client.On("GetSchedulingEventFilters", mock.Anything).Return([]*SchedulingEventFilter{}, nil)
	client.On("GetSchedulingEventProcessingStats", mock.Anything).Return(&SchedulingEventProcessingStats{}, nil)
	client.On("GetSchedulingStreamingPerformanceMetrics", mock.Anything).Return(&SchedulingStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := []prometheus.Metric{}
	for m := range ch {
		metrics = append(metrics, m)
	}

	assert.Greater(t, len(metrics), 0)

	// Verify that we have metrics for active streams
	hasActiveStreamMetric := false
	for _, m := range metrics {
		dto := &prometheus.Metric{}
		_ = m.Write(dto)
		if dto.Label != nil {
			for _, label := range dto.Label {
				if label.GetName() == "stream_type" && (label.GetValue() == "backfill" || label.GetValue() == "priority") {
					hasActiveStreamMetric = true
					break
				}
			}
		}
	}
	assert.True(t, hasActiveStreamMetric, "Should have metrics for active streams")

	client.AssertExpectations(t)
}

func TestJobSchedulingStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

	streamingMetrics := &SchedulingStreamingMetrics{
		TotalStreams:              25,
		ActiveStreams:             18,
		HealthyStreams:            15,
		DegradedStreams:           2,
		FailedStreams:             1,
		StreamCreationRate:        2.5,
		TotalEventsProcessed:      10000000,
		EventsPerSecond:           1250.5,
		AverageEventSize:          2048,
		PeakEventRate:             2500.0,
		EventBacklog:              1500,
		EventProcessingLatency:    time.Millisecond * 25,
		CPUUtilization:            0.65,
		MemoryUtilization:         0.72,
		NetworkUtilization:        0.48,
		DiskUtilization:           0.55,
		SystemLoad:                2.8,
		ResourceEfficiency:        0.88,
		Uptime:                    time.Hour * 720, // 30 days
		Availability:              0.9999,
		MTBF:                      time.Hour * 168, // 1 week
		MTTR:                      time.Minute * 5,
		ErrorRate:                 0.001,
		RecoveryRate:              0.98,
		DataAccuracy:              0.995,
		DataCompleteness:          0.998,
		DataTimeliness:            0.992,
		EnrichmentSuccess:         0.97,
		CorrelationSuccess:        0.95,
		DeduplicationRate:         0.15,
		TotalBusinessValue:        2500000.0,
		AverageCostPerEvent:       0.0015,
		ROI:                       18.5,
		CustomerSatisfaction:      0.94,
		SLACompliance:             0.998,
		BusinessImpact:            0.92,
		OptimizationOpportunities: 45,
		PotentialSavings:          125000.0,
		EfficiencyScore:           0.91,
		PerformanceScore:          0.93,
		QualityScore:              0.95,
		ValueScore:                0.89,
		CapacityUtilization:       0.78,
		ScalingEvents:             12,
		ResourceHeadroom:          0.22,
		ProjectedCapacityNeeds:    0.85,
		CapacityPlanningScore:     0.88,
		GrowthRate:                0.15,
		TotalCost:                 15000.0,
		InfrastructureCost:        8000.0,
		OperationalCost:           5000.0,
		DataTransferCost:          1500.0,
		StorageCost:               500.0,
		CostTrend:                 -0.05,
		ComplianceScore:           0.97,
		PolicyViolations:          2,
		AuditEvents:               150,
		SecurityIncidents:         0,
		PrivacyIncidents:          0,
		RegulatoryIssues:          0,
	}

	client.On("GetSchedulingStreamingConfiguration", mock.Anything).Return(&SchedulingStreamingConfiguration{}, nil)
	client.On("GetActiveSchedulingStreams", mock.Anything).Return([]*ActiveSchedulingStream{}, nil)
	client.On("GetSchedulingStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetSchedulingStreamingHealthStatus", mock.Anything).Return(&SchedulingStreamingHealthStatus{}, nil)
	client.On("GetSchedulingEventSubscriptions", mock.Anything).Return([]*SchedulingEventSubscription{}, nil)
	client.On("GetSchedulingEventFilters", mock.Anything).Return([]*SchedulingEventFilter{}, nil)
	client.On("GetSchedulingEventProcessingStats", mock.Anything).Return(&SchedulingEventProcessingStats{}, nil)
	client.On("GetSchedulingStreamingPerformanceMetrics", mock.Anything).Return(&SchedulingStreamingPerformanceMetrics{}, nil)

	ch := make(chan prometheus.Metric, 100)
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

func TestJobSchedulingStreamingCollector_CollectError(t *testing.T) {
	client := &MockJobSchedulingStreamingSLURMClient{}
	collector := NewJobSchedulingStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetSchedulingStreamingConfiguration", mock.Anything).Return((*SchedulingStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveSchedulingStreams", mock.Anything).Return([]*ActiveSchedulingStream(nil), assert.AnError)
	client.On("GetSchedulingStreamingMetrics", mock.Anything).Return((*SchedulingStreamingMetrics)(nil), assert.AnError)
	client.On("GetSchedulingStreamingHealthStatus", mock.Anything).Return((*SchedulingStreamingHealthStatus)(nil), assert.AnError)
	client.On("GetSchedulingEventSubscriptions", mock.Anything).Return([]*SchedulingEventSubscription(nil), assert.AnError)
	client.On("GetSchedulingEventFilters", mock.Anything).Return([]*SchedulingEventFilter(nil), assert.AnError)
	client.On("GetSchedulingEventProcessingStats", mock.Anything).Return((*SchedulingEventProcessingStats)(nil), assert.AnError)
	client.On("GetSchedulingStreamingPerformanceMetrics", mock.Anything).Return((*SchedulingStreamingPerformanceMetrics)(nil), assert.AnError)

	ch := make(chan prometheus.Metric, 100)
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