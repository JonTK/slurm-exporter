package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQueueStateStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockQueueStateStreamingSLURMClient) StreamQueueStateChanges(ctx context.Context) (<-chan QueueStateChangeEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan QueueStateChangeEvent), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueStreamingConfiguration(ctx context.Context) (*QueueStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueStreamingConfiguration), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetActiveQueueStreams(ctx context.Context) ([]*ActiveQueueStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveQueueStream), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueEventHistory(ctx context.Context, queueID string, duration time.Duration) ([]*QueueEvent, error) {
	args := m.Called(ctx, queueID, duration)
	return args.Get(0).([]*QueueEvent), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueStreamingMetrics(ctx context.Context) (*QueueStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueStreamingMetrics), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueEventFilters(ctx context.Context) ([]*QueueEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*QueueEventFilter), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) ConfigureQueueStreaming(ctx context.Context, config *QueueStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueStreamingHealthStatus(ctx context.Context) (*QueueStreamingHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueStreamingHealthStatus), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueEventSubscriptions(ctx context.Context) ([]*QueueEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*QueueEventSubscription), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) ManageQueueEventSubscription(ctx context.Context, subscription *QueueEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueEventProcessingStats(ctx context.Context) (*QueueEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueEventProcessingStats), args.Error(1)
}

func (m *MockQueueStateStreamingSLURMClient) GetQueueStreamingPerformanceMetrics(ctx context.Context) (*QueueStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*QueueStreamingPerformanceMetrics), args.Error(1)
}

func TestNewQueueStateStreamingCollector(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.queueEventsTotal)
	assert.NotNil(t, collector.activeQueueStreams)
	assert.NotNil(t, collector.streamingHealthScore)
	assert.NotNil(t, collector.queueStateChanges)
	assert.NotNil(t, collector.queueLength)
	assert.NotNil(t, collector.backfillJobs)
	assert.NotNil(t, collector.preemptedJobs)
	assert.NotNil(t, collector.pendingJobs)
	assert.NotNil(t, collector.queueCoverage)
}

func TestQueueStateStreamingCollector_Describe(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	ch := make(chan *prometheus.Desc, 200)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 85) // Should have many metrics
}

func TestQueueStateStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	config := &QueueStreamingConfiguration{
		StreamingEnabled:         true,
		EventBufferSize:         20000,
		EventBatchSize:          200,
		MaxConcurrentStreams:    100,
		RateLimitPerSecond:      2000,
		BackpressureThreshold:   10000,
		CompressionEnabled:      true,
		EncryptionEnabled:       true,
		AuthenticationRequired:  true,
		DeduplicationEnabled:    true,
		EventValidationEnabled:  true,
		MetricsCollectionEnabled: true,
		DebugLoggingEnabled:     false,
		PriorityBasedStreaming:  true,
		EventEnrichmentEnabled:  true,
		FailoverEnabled:         true,
		QueueMonitoring:         true,
		BackfillAnalysis:        true,
		PreemptionTracking:      true,
		ResourceTracking:        true,
		PerformanceAnalysis:     true,
		TrendAnalysis:           true,
		AnomalyDetection:        true,
		PredictiveAnalytics:     true,
		AlertIntegration:        true,
		SLAMonitoring:           true,
		ComplianceTracking:      true,
		QualityAssurance:        true,
		CostTracking:            true,
		CapacityPlanning:        true,
		LoadBalancing:           true,
		OptimizationEnabled:     true,
		AutoScaling:             true,
		MaintenanceMode:         false,
		EmergencyProtocols:      true,
		AuditLogging:            true,
		SecurityMonitoring:      true,
		DashboardIntegration:    true,
		ReportingEnabled:        true,
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveQueueStream{
		{
			StreamID:             "queue-stream-001",
			QueueID:              "queue-001",
			QueueName:            "main",
			PartitionName:        "compute",
			StreamStartTime:      now.Add(-time.Hour * 3),
			LastEventTime:        now,
			EventCount:           5000,
			StreamStatus:         "active",
			StreamType:           "queue_monitoring",
			ConsumerID:           "consumer-queue-001",
			ConsumerEndpoint:     "http://queue-monitor.example.com/events",
			StreamPriority:       9,
			BufferedEvents:       150,
			ProcessedEvents:      4850,
			FailedEvents:         25,
			RetryCount:           5,
			Bandwidth:            4096000.0,
			Latency:              time.Millisecond * 20,
			CompressionRatio:     0.75,
			EventRate:            185.2,
			ConnectionQuality:    0.99,
			StreamHealth:         "healthy",
			LastHeartbeat:        now,
			BackpressureActive:   false,
			FailoverActive:       false,
			QueuedEvents:         75,
			DroppedEvents:        8,
			FilterMatches:        4200,
			ValidationErrors:     12,
			EnrichmentFailures:   6,
			DeliveryAttempts:     4863,
			AckRate:              0.998,
			ResourceUsage: map[string]float64{
				"cpu":    0.12,
				"memory": 0.22,
			},
			ConfigurationHash:        "def789ghi012",
			SecurityContext:          "secure",
			ComplianceFlags:          []string{"sox", "gdpr"},
			PerformanceProfile:       "high_performance",
			QualityScore:             0.96,
			AlertsGenerated:          45,
			ActionsTriggered:         12,
			OptimizationsSuggested:   8,
			PredictionsAccuracy:      0.88,
			AnomaliesDetected:        3,
			TrendsIdentified:         15,
			CostImpact:               245.75,
			BusinessValue:            892.50,
			UserSatisfaction:         0.94,
			SLACompliance:            0.998,
			SystemImpact:             0.85,
		},
		{
			StreamID:             "queue-stream-002",
			QueueID:              "queue-002",
			QueueName:            "gpu",
			PartitionName:        "gpu",
			StreamStartTime:      now.Add(-time.Hour),
			LastEventTime:        now.Add(-time.Minute * 3),
			EventCount:           2200,
			StreamStatus:         "warning",
			StreamType:           "backfill_analysis",
			ConsumerID:           "consumer-queue-002",
			ConsumerEndpoint:     "http://backfill-analyzer.example.com/events",
			StreamPriority:       7,
			BufferedEvents:       280,
			ProcessedEvents:      1920,
			FailedEvents:         45,
			RetryCount:           12,
			Bandwidth:            2048000.0,
			Latency:              time.Millisecond * 65,
			CompressionRatio:     0.70,
			EventRate:            95.8,
			ConnectionQuality:    0.88,
			StreamHealth:         "warning",
			LastHeartbeat:        now.Add(-time.Minute * 3),
			BackpressureActive:   true,
			FailoverActive:       false,
			QueuedEvents:         200,
			DroppedEvents:        28,
			FilterMatches:        1850,
			ValidationErrors:     22,
			EnrichmentFailures:   18,
			DeliveryAttempts:     1983,
			AckRate:              0.975,
			ResourceUsage: map[string]float64{
				"cpu":    0.28,
				"memory": 0.42,
			},
			ConfigurationHash:        "ghi012jkl345",
			SecurityContext:          "secure",
			ComplianceFlags:          []string{"hipaa"},
			PerformanceProfile:       "balanced",
			QualityScore:             0.88,
			AlertsGenerated:          85,
			ActionsTriggered:         25,
			OptimizationsSuggested:   18,
			PredictionsAccuracy:      0.82,
			AnomaliesDetected:        8,
			TrendsIdentified:         22,
			CostImpact:               125.25,
			BusinessValue:            456.75,
			UserSatisfaction:         0.86,
			SLACompliance:            0.985,
			SystemImpact:             0.72,
		},
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	streamingMetrics := &QueueStreamingMetrics{
		TotalStreams:             200,
		ActiveStreams:            65,
		PausedStreams:            12,
		FailedStreams:            5,
		EventsPerSecond:          1250.5,
		AverageEventLatency:      time.Millisecond * 28,
		MaxEventLatency:          time.Millisecond * 150,
		MinEventLatency:          time.Millisecond * 5,
		TotalEventsProcessed:     5800000,
		TotalEventsDropped:       580,
		TotalEventsFailed:        290,
		AverageStreamDuration:    time.Hour * 4,
		MaxStreamDuration:        time.Hour * 16,
		TotalBandwidthUsed:       40960000.0,
		CompressionEfficiency:    0.75,
		DeduplicationRate:        0.22,
		ErrorRate:                0.002,
		SuccessRate:              0.998,
		BackpressureOccurrences:  75,
		FailoverOccurrences:      12,
		ReconnectionAttempts:     45,
		MemoryUsage:              8192000000,
		CPUUsage:                 0.42,
		NetworkUsage:             0.65,
		DiskUsage:                42949672960,
		CacheHitRate:             0.91,
		QueueDepth:               350,
		ProcessingEfficiency:     0.96,
		StreamingHealth:          0.98,
		QueueCoverage:            0.95,
		StateChangeAccuracy:      0.999,
		PredictionAccuracy:       0.88,
		AlertResponseTime:        time.Second * 8,
		OptimizationEffectiveness: 0.85,
		QualityScore:             0.94,
		ComplianceScore:          0.97,
		SecurityScore:            0.98,
		PerformanceScore:         0.92,
		BusinessValue:            0.89,
		UserSatisfaction:         0.91,
		CostEfficiency:           0.87,
		ResourceUtilization:      0.78,
		ThroughputOptimization:   0.83,
		LatencyOptimization:      0.89,
		CapacityUtilization:      0.72,
		LoadBalanceEfficiency:    0.86,
		BackfillEfficiency:       0.79,
		PreemptionEfficiency:     0.82,
		SLACompliance:            0.996,
		AnomalyDetectionRate:     0.92,
		TrendPredictionAccuracy:  0.84,
		ForecastAccuracy:         0.81,
		OptimizationImpact:       0.78,
		AutomationEffectiveness:  0.88,
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_StreamingHealth(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	healthStatus := &QueueStreamingHealthStatus{
		OverallHealth: "healthy",
		ComponentHealth: map[string]string{
			"event_processor":      "healthy",
			"stream_manager":       "healthy",
			"filter_engine":        "healthy",
			"delivery_system":      "healthy",
			"queue_monitor":        "healthy",
			"backfill_analyzer":    "warning",
			"preemption_tracker":   "healthy",
			"resource_tracker":     "healthy",
			"performance_analyzer": "healthy",
			"trend_analyzer":       "healthy",
			"anomaly_detector":     "healthy",
			"optimization_engine":  "healthy",
		},
		LastHealthCheck:     time.Now(),
		HealthCheckDuration: time.Millisecond * 85,
		HealthScore:         0.98,
		CriticalIssues:      []string{},
		WarningIssues:       []string{"Backfill analyzer showing decreased efficiency in GPU partition"},
		InfoMessages:        []string{"System operating optimally", "Queue coverage at 95%", "All SLAs met"},
		StreamingUptime:     time.Hour * 168,
		ServiceAvailability: 0.9999,
		ResourceUtilization: map[string]float64{
			"cpu":     0.42,
			"memory":  0.65,
			"disk":    0.25,
			"network": 0.55,
		},
		PerformanceMetrics: map[string]float64{
			"throughput":               1250.5,
			"latency":                  0.028,
			"error_rate":               0.002,
			"queue_coverage":           0.95,
			"state_change_accuracy":    0.999,
			"prediction_accuracy":      0.88,
			"optimization_effectiveness": 0.85,
		},
		ErrorSummary: map[string]int64{
			"validation_errors":   75,
			"network_errors":      12,
			"processing_errors":   28,
			"enrichment_errors":   18,
			"delivery_errors":     22,
		},
		HealthTrends: map[string]float64{
			"health_score_trend":          0.012,
			"error_rate_trend":            -0.0003,
			"performance_trend":           0.015,
			"optimization_trend":          0.008,
		},
		PredictedIssues:    []string{},
		RecommendedActions: []string{"Monitor GPU partition backfill efficiency", "Consider queue rebalancing for optimal throughput"},
		SystemCapacity: map[string]float64{
			"max_streams":       5000,
			"max_throughput":    10000,
			"max_connections":   2000,
			"max_queues":        1000,
		},
		AlertThresholds: map[string]float64{
			"error_rate":               0.005,
			"latency":                  0.100,
			"cpu_usage":                0.80,
			"queue_coverage":           0.90,
			"prediction_accuracy":      0.80,
			"optimization_effectiveness": 0.75,
		},
		SLACompliance: map[string]float64{
			"availability":            0.9995,
			"performance":             0.998,
			"reliability":             0.999,
			"queue_coverage":          0.95,
			"optimization_effectiveness": 0.85,
			"user_satisfaction":       0.91,
		},
		DependencyStatus: map[string]string{
			"slurm_api":           "healthy",
			"prometheus":          "healthy",
			"storage":             "healthy",
			"queue_managers":      "healthy",
			"backfill_scheduler":  "warning",
			"resource_manager":    "healthy",
		},
		ConfigurationValid:  true,
		SecurityStatus:      "secure",
		BackupStatus:        "active",
		MonitoringEnabled:   true,
		LoggingEnabled:      true,
		MaintenanceSchedule: []string{"Sunday 01:00-03:00", "Monthly queue optimization"},
		CapacityForecasts: map[string]float64{
			"queue_growth":        0.18,
			"stream_growth":       0.22,
			"throughput_growth":   0.25,
			"resource_demand":     0.20,
		},
		RiskIndicators: map[string]float64{
			"queue_saturation_risk": 0.15,
			"capacity_risk":         0.08,
			"performance_risk":      0.05,
			"security_risk":         0.02,
		},
		ComplianceMetrics: map[string]float64{
			"data_retention":     1.0,
			"audit_coverage":     0.99,
			"security_score":     0.98,
			"policy_compliance":  0.97,
		},
		QualityMetrics: map[string]float64{
			"data_quality":       0.96,
			"stream_quality":     0.94,
			"analysis_quality":   0.92,
			"prediction_quality": 0.88,
		},
		PerformanceBaselines: map[string]float64{
			"baseline_throughput":      1000.0,
			"baseline_latency":         0.035,
			"baseline_accuracy":        0.995,
			"baseline_optimization":    0.80,
		},
		AnomalyDetectors: map[string]bool{
			"performance_anomaly":   true,
			"security_anomaly":      true,
			"capacity_anomaly":      true,
			"queue_anomaly":         true,
			"backfill_anomaly":      true,
		},
		AutomationStatus: map[string]bool{
			"auto_scaling":          true,
			"auto_remediation":      true,
			"auto_optimization":     true,
			"auto_load_balancing":   true,
			"auto_maintenance":      false,
		},
		IntegrationHealth: map[string]string{
			"grafana":           "healthy",
			"alertmanager":      "healthy",
			"elk_stack":         "healthy",
			"queue_dashboard":   "healthy",
			"optimization_ui":   "healthy",
		},
		BusinessMetrics: map[string]float64{
			"business_value":     0.89,
			"cost_efficiency":    0.87,
			"user_satisfaction":  0.91,
			"operational_impact": 0.85,
		},
		UserExperience: map[string]float64{
			"queue_satisfaction": 0.91,
			"wait_time_satisfaction": 0.88,
			"throughput_satisfaction": 0.93,
			"reliability_satisfaction": 0.95,
		},
		CostMetrics: map[string]float64{
			"operational_cost":   2450.75,
			"optimization_savings": 485.25,
			"efficiency_gains":   0.15,
			"roi":                1.85,
		},
		EfficiencyMetrics: map[string]float64{
			"resource_efficiency":    0.78,
			"throughput_efficiency":  0.83,
			"backfill_efficiency":    0.79,
			"preemption_efficiency":  0.82,
		},
		OptimizationOpportunities: []string{
			"GPU partition backfill optimization",
			"Cross-partition load balancing",
			"Predictive scaling for peak hours",
		},
		TrendAnalysis: map[string]interface{}{
			"throughput_trend":     "increasing",
			"latency_trend":        "stable",
			"error_trend":          "decreasing",
			"efficiency_trend":     "improving",
		},
		PredictiveInsights: map[string]interface{}{
			"capacity_forecast":    "moderate_growth",
			"performance_forecast": "stable_improvement",
			"risk_forecast":        "low_risk",
		},
		RecommendationEngine: map[string]interface{}{
			"optimization_recommendations": []string{"backfill_tuning", "load_balancing"},
			"capacity_recommendations":     []string{"gpu_expansion", "cpu_reallocation"},
			"performance_recommendations":  []string{"queue_rebalancing", "priority_tuning"},
		},
		AlertingEffectiveness:   0.94,
		ResponseTimeMetrics: map[string]float64{
			"alert_response_time":      8.5,
			"issue_resolution_time":    45.2,
			"optimization_deploy_time": 125.8,
		},
		EscalationEffectiveness: 0.91,
		ResolutionRates: map[string]float64{
			"automatic_resolution": 0.75,
			"assisted_resolution":  0.92,
			"manual_resolution":    0.98,
		},
		PreventativeActions: []string{
			"Proactive queue rebalancing",
			"Predictive capacity scaling",
			"Automated optimization tuning",
		},
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(healthStatus, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_EventFilters(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	now := time.Now()
	filters := []*QueueEventFilter{
		{
			FilterID:            "queue-filter-001",
			FilterName:          "High Priority Queue Events",
			FilterType:          "priority_based",
			FilterExpression:    "queue_priority >= 8 AND wait_time > 300",
			IncludePattern:      "priority,urgent,critical",
			ExcludePattern:      "test,debug,maintenance",
			QueueStates:         []string{"active", "draining"},
			Partitions:          []string{"compute", "gpu"},
			Queues:              []string{"main", "priority", "urgent"},
			EventTypes:          []string{"queue_length_change", "backfill_opportunity"},
			Priorities:          []int{8, 9, 10},
			QueueLengthRange:    []int{50, 1000},
			WaitTimeRange:       []time.Duration{time.Minute * 5, time.Hour * 2},
			ResourceCriteria: map[string]interface{}{
				"min_cpus":        32,
				"min_memory":      "128GB",
				"gpu_required":    true,
			},
			PerformanceCriteria: map[string]interface{}{
				"throughput_threshold": 100.0,
				"efficiency_threshold": 0.80,
			},
			QualityCriteria: map[string]interface{}{
				"data_quality_min":      0.95,
				"prediction_accuracy":   0.85,
			},
			ComplianceCriteria: map[string]interface{}{
				"sla_compliance":        0.99,
				"security_compliance":   true,
			},
			BusinessCriteria: map[string]interface{}{
				"business_value":        0.80,
				"cost_threshold":        1000.0,
			},
			CustomCriteria: map[string]interface{}{
				"department":            "research",
				"criticality":           "high",
				"optimization_target":   "throughput",
			},
			FilterEnabled:      true,
			FilterPriority:     10,
			CreatedBy:          "queue-admin",
			CreatedTime:        now.Add(-time.Hour * 72),
			ModifiedBy:         "queue-admin",
			ModifiedTime:       now.Add(-time.Hour * 3),
			UsageCount:         12500,
			LastUsedTime:       now.Add(-time.Minute),
			FilterDescription:  "Filter for high priority queue events requiring immediate optimization",
			FilterTags:         []string{"critical", "production", "optimization"},
			ValidationRules:    []string{"syntax_check", "performance_check", "business_check"},
			MatchCount:         75000,
			FilteredCount:      52000,
			ErrorCount:         18,
			PerformanceImpact:  0.08,
			MaintenanceWindow:  "Sunday 01:00-03:00",
			EmergencyBypass:    true,
			ComplianceLevel:    "high",
			AuditTrail:         []string{"created", "modified", "activated", "optimized"},
			BusinessContext:    "High-priority queue optimization and SLA compliance",
			TechnicalContext:   "Real-time queue monitoring and backfill optimization",
			OperationalContext: "Production queue management and performance tuning",
			CostImplications:   185.75,
			RiskAssessment:     "low",
			QualityImpact:      0.92,
			UserImpact:         0.88,
			SystemImpact:       0.85,
			SecurityImplications: "medium",
			DataRetention:      "30_days",
			PrivacySettings: map[string]bool{
				"anonymize_user_data": true,
				"encrypt_sensitive":   true,
			},
		},
		{
			FilterID:            "queue-filter-002",
			FilterName:          "Backfill Analysis Events",
			FilterType:          "backfill_optimization",
			FilterExpression:    "event_type = 'backfill_opportunity' OR backfill_efficiency < 0.70",
			IncludePattern:      "backfill,opportunity,optimization",
			ExcludePattern:      "maintenance,test",
			QueueStates:         []string{"active"},
			Partitions:          []string{"compute", "gpu", "memory"},
			Queues:              []string{},
			EventTypes:          []string{"backfill_opportunity", "preemption_event"},
			Priorities:          []int{4, 5, 6, 7},
			QueueLengthRange:    []int{10, 500},
			WaitTimeRange:       []time.Duration{time.Minute, time.Hour},
			ResourceCriteria: map[string]interface{}{
				"backfill_efficiency": 0.70,
				"resource_fragmentation": 0.30,
			},
			PerformanceCriteria: map[string]interface{}{
				"optimization_potential": 0.20,
				"improvement_opportunity": 0.15,
			},
			QualityCriteria: map[string]interface{}{
				"analysis_confidence":   0.80,
				"prediction_reliability": 0.75,
			},
			ComplianceCriteria: map[string]interface{}{
				"optimization_policy":   true,
				"resource_policy":       true,
			},
			BusinessCriteria: map[string]interface{}{
				"efficiency_impact":     0.10,
				"cost_optimization":     true,
			},
			CustomCriteria: map[string]interface{}{
				"optimization_type":     "backfill",
				"analysis_scope":        "partition",
				"automation_eligible":   true,
			},
			FilterEnabled:      true,
			FilterPriority:     7,
			CreatedBy:          "optimization-admin",
			CreatedTime:        now.Add(-time.Hour * 96),
			ModifiedBy:         "optimization-admin",
			ModifiedTime:       now.Add(-time.Hour * 8),
			UsageCount:         8500,
			LastUsedTime:       now.Add(-time.Minute * 10),
			FilterDescription:  "Filter for backfill optimization opportunities and efficiency analysis",
			FilterTags:         []string{"backfill", "optimization", "efficiency"},
			ValidationRules:    []string{"syntax_check", "efficiency_check"},
			MatchCount:         35000,
			FilteredCount:      28000,
			ErrorCount:         8,
			PerformanceImpact:  0.04,
			MaintenanceWindow:  "Sunday 01:00-03:00",
			EmergencyBypass:    false,
			ComplianceLevel:    "medium",
			AuditTrail:         []string{"created", "modified", "optimized"},
			BusinessContext:    "Backfill optimization and resource efficiency improvement",
			TechnicalContext:   "Automated backfill analysis and opportunity detection",
			OperationalContext: "Continuous queue optimization and performance tuning",
			CostImplications:   95.25,
			RiskAssessment:     "very_low",
			QualityImpact:      0.88,
			UserImpact:         0.82,
			SystemImpact:       0.78,
			SecurityImplications: "low",
			DataRetention:      "7_days",
			PrivacySettings: map[string]bool{
				"anonymize_user_data": false,
				"encrypt_sensitive":   true,
			},
		},
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return(filters, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_ProcessingStats(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	now := time.Now()
	processingStats := &QueueEventProcessingStats{
		ProcessingStartTime:      now.Add(-time.Hour * 4),
		TotalEventsReceived:      450000,
		TotalEventsProcessed:     432000,
		TotalEventsFiltered:      378000,
		TotalEventsDropped:       8500,
		TotalProcessingTime:      time.Hour * 4,
		AverageProcessingTime:    time.Millisecond * 32,
		MaxProcessingTime:       time.Millisecond * 750,
		MinProcessingTime:       time.Millisecond * 2,
		ProcessingThroughput:    30.0,
		ErrorRate:               0.015,
		SuccessRate:             0.985,
		FilterEfficiency:        0.87,
		ValidationErrors:        1850,
		EnrichmentErrors:        650,
		DeliveryErrors:          980,
		TransformationErrors:    285,
		SerializationErrors:     125,
		NetworkErrors:           485,
		AuthenticationErrors:    85,
		AuthorizationErrors:     32,
		RateLimitExceeded:       225,
		BackpressureEvents:      350,
		CircuitBreakerTrips:     15,
		RetryAttempts:           4500,
		DeadLetterEvents:        125,
		DuplicateEvents:         680,
		OutOfOrderEvents:        920,
		LateArrivingEvents:      350,
		ProcessingQueues: map[string]int64{
			"main":         2500,
			"priority":     1250,
			"bulk":         4500,
			"backfill":     850,
			"optimization": 650,
			"dlq":          125,
		},
		WorkerStatistics: map[string]interface{}{
			"active_workers":        25,
			"idle_workers":          12,
			"blocked_workers":       5,
			"optimization_workers":  8,
			"analysis_workers":      6,
		},
		ResourceUtilization: map[string]float64{
			"cpu":     0.65,
			"memory":  0.78,
			"disk":    0.35,
			"network": 0.58,
		},
		PerformanceCounters: map[string]int64{
			"cache_hits":         325000,
			"cache_misses":       45000,
			"gc_cycles":          125,
			"context_switches":   15500,
			"optimization_runs":  285,
		},
		QueueStateAccuracy:      0.999,
		EventCorrelationRate:    0.94,
		AnomalyDetectionRate:    0.18,
		TrendDetectionRate:      0.25,
		PredictionAccuracy:      0.88,
		OptimizationImpact:      0.15,
		AlertGenerationRate:     0.12,
		ActionExecutionRate:     0.85,
		ComplianceChecks:        25000,
		QualityChecks:          18500,
		SecurityScans:          1850,
		PerformanceAnalysis:     8500,
		BusinessAnalysis:        4500,
		CostAnalysis:           2850,
		UserAnalysis:           6500,
		SystemAnalysis:         9500,
		DataQualityScore:        0.96,
		SystemLoadImpact:        0.35,
		BusinessImpact:          0.82,
		UserImpact:             0.88,
		OperationalImpact:       0.85,
		FinancialImpact:         0.78,
		StrategicImpact:         0.72,
		CompetitiveImpact:       0.68,
		RiskImpact:             0.25,
		ComplianceImpact:        0.95,
		QualityImpact:          0.92,
		PerformanceImpact:      0.88,
		EfficiencyImpact:       0.85,
		ProductivityImpact:     0.82,
		InnovationImpact:       0.75,
		SustainabilityImpact:   0.78,
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(processingStats, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(&QueueStreamingPerformanceMetrics{}, nil)

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

func TestQueueStateStreamingCollector_Collect_PerformanceMetrics(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	performanceMetrics := &QueueStreamingPerformanceMetrics{
		Throughput:                1250.5,
		Latency:                   time.Millisecond * 28,
		P50Latency:                time.Millisecond * 20,
		P95Latency:                time.Millisecond * 75,
		P99Latency:                time.Millisecond * 140,
		MaxLatency:                time.Millisecond * 750,
		MessageRate:               1180.2,
		ByteRate:                  8192000.0,
		ErrorRate:                 0.002,
		SuccessRate:               0.998,
		AvailabilityPercentage:    99.99,
		UptimePercentage:          99.995,
		CPUUtilization:            0.52,
		MemoryUtilization:         0.72,
		NetworkUtilization:        0.68,
		DiskUtilization:           0.32,
		ConnectionCount:           650,
		ActiveConnections:         485,
		IdleConnections:           165,
		FailedConnections:         12,
		ConnectionPoolSize:        700,
		QueueDepth:                350,
		BufferUtilization:         0.82,
		GCPressure:                0.22,
		GCFrequency:               0.15,
		GCDuration:                time.Millisecond * 45,
		HeapSize:                  8589934592,
		ThreadCount:               125,
		ContextSwitches:           45000,
		SystemCalls:               85000,
		PageFaults:                1850,
		CacheMisses:               2500,
		BranchMispredictions:      5200,
		InstructionsPerSecond:     2500000.0,
		CyclesPerInstruction:      2.1,
		PerformanceScore:          0.96,
		QueueCoverageEfficiency:   0.95,
		StateDetectionAccuracy:    0.999,
		EventProcessingSpeed:      1250.5,
		AlertLatency:              time.Second * 8,
		ResponseTime:              time.Second * 15,
		ResolutionTime:            time.Minute * 5,
		OptimizationEffectiveness: 0.85,
		PredictionAccuracy:        0.88,
		AnomalyDetectionAccuracy:  0.92,
		TrendAnalysisAccuracy:     0.84,
		ForecastAccuracy:          0.81,
		RecommendationAccuracy:    0.86,
		AutomationEffectiveness:   0.88,
		BusinessValue:             0.89,
		CostEffectiveness:         0.87,
		QualityScore:              0.94,
		ComplianceScore:           0.97,
		SecurityScore:             0.98,
		UserSatisfactionScore:     0.91,
		SystemEfficiency:          0.89,
		ResourceOptimization:      0.83,
		CapacityUtilization:       0.72,
		LoadBalancingEfficiency:   0.86,
		ThroughputOptimization:    0.88,
		LatencyOptimization:       0.92,
		ErrorReduction:            0.85,
		AvailabilityImprovement:   0.15,
		ReliabilityImprovement:    0.12,
		ScalabilityImprovement:    0.25,
		MaintainabilityScore:      0.88,
		MonitorabilityScore:       0.94,
		ObservabilityScore:        0.92,
		DebuggabilityScore:        0.86,
		TestabilityScore:          0.84,
		ExtensibilityScore:        0.82,
		PortabilityScore:          0.78,
		UsabilityScore:            0.91,
		AccessibilityScore:        0.89,
		InteroperabilityScore:     0.87,
		CompatibilityScore:        0.85,
		StandardsComplianceScore:  0.95,
		BestPracticesScore:        0.92,
		InnovationScore:           0.78,
		SustainabilityScore:       0.82,
		EthicsScore:               0.95,
		TransparencyScore:         0.88,
		AccountabilityScore:       0.91,
		ResponsivenessScore:       0.89,
		AdaptabilityScore:         0.85,
		ResilienceScore:           0.87,
		RobustnessScore:           0.89,
		ReliabilityScore:          0.94,
		ConsistencyScore:          0.92,
		PredictabilityScore:       0.88,
		StabilityScore:            0.91,
		MaturityScore:             0.86,
		EvolutionScore:            0.78,
		GrowthScore:               0.82,
		ImpactScore:               0.89,
	}

	client.On("GetQueueStreamingConfiguration", mock.Anything).Return(&QueueStreamingConfiguration{}, nil)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream{}, nil)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return(&QueueStreamingMetrics{}, nil)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return(&QueueStreamingHealthStatus{}, nil)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription{}, nil)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter{}, nil)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return(&QueueEventProcessingStats{}, nil)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return(performanceMetrics, nil)

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

func TestQueueStateStreamingCollector_CollectError(t *testing.T) {
	client := &MockQueueStateStreamingSLURMClient{}
	collector := NewQueueStateStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetQueueStreamingConfiguration", mock.Anything).Return((*QueueStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveQueueStreams", mock.Anything).Return([]*ActiveQueueStream(nil), assert.AnError)
	client.On("GetQueueStreamingMetrics", mock.Anything).Return((*QueueStreamingMetrics)(nil), assert.AnError)
	client.On("GetQueueStreamingHealthStatus", mock.Anything).Return((*QueueStreamingHealthStatus)(nil), assert.AnError)
	client.On("GetQueueEventSubscriptions", mock.Anything).Return([]*QueueEventSubscription(nil), assert.AnError)
	client.On("GetQueueEventFilters", mock.Anything).Return([]*QueueEventFilter(nil), assert.AnError)
	client.On("GetQueueEventProcessingStats", mock.Anything).Return((*QueueEventProcessingStats)(nil), assert.AnError)
	client.On("GetQueueStreamingPerformanceMetrics", mock.Anything).Return((*QueueStreamingPerformanceMetrics)(nil), assert.AnError)

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