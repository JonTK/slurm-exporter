package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNodeStateStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockNodeStateStreamingSLURMClient) StreamNodeStateChanges(ctx context.Context) (<-chan NodeStateChangeEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan NodeStateChangeEvent), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeStreamingConfiguration(ctx context.Context) (*NodeStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NodeStreamingConfiguration), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetActiveNodeStreams(ctx context.Context) ([]*ActiveNodeStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveNodeStream), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeEventHistory(ctx context.Context, nodeID string, duration time.Duration) ([]*NodeEvent, error) {
	args := m.Called(ctx, nodeID, duration)
	return args.Get(0).([]*NodeEvent), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeStreamingMetrics(ctx context.Context) (*NodeStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NodeStreamingMetrics), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeEventFilters(ctx context.Context) ([]*NodeEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*NodeEventFilter), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) ConfigureNodeStreaming(ctx context.Context, config *NodeStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeStreamingHealthStatus(ctx context.Context) (*NodeStreamingHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NodeStreamingHealthStatus), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeEventSubscriptions(ctx context.Context) ([]*NodeEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*NodeEventSubscription), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) ManageNodeEventSubscription(ctx context.Context, subscription *NodeEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeEventProcessingStats(ctx context.Context) (*NodeEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NodeEventProcessingStats), args.Error(1)
}

func (m *MockNodeStateStreamingSLURMClient) GetNodeStreamingPerformanceMetrics(ctx context.Context) (*NodeStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NodeStreamingPerformanceMetrics), args.Error(1)
}

func TestNewNodeStateStreamingCollector(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.nodeEventsTotal)
	assert.NotNil(t, collector.activeNodeStreams)
	assert.NotNil(t, collector.streamingHealthScore)
	assert.NotNil(t, collector.nodeStateChanges)
	assert.NotNil(t, collector.nodeCoverage)
}

func TestNodeStateStreamingCollector_Describe(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 60) // Should have many metrics
}

func TestNodeStateStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	config := &NodeStreamingConfiguration{
		StreamingEnabled:         true,
		EventBufferSize:         15000,
		EventBatchSize:          150,
		MaxConcurrentStreams:    75,
		RateLimitPerSecond:      1500,
		BackpressureThreshold:   7500,
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
		StateChangeDetection:    true,
		PerformanceMonitoring:   true,
		AlertIntegration:        true,
		EventCorrelation:        true,
		AnomalyDetection:        true,
		PredictiveAnalytics:     true,
		MaintenanceMode:         false,
		EmergencyProtocols:      true,
		AuditLogging:            true,
		ComplianceReporting:     true,
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveNodeStream{
		{
			StreamID:           "node-stream-001",
			NodeID:             "node-001",
			NodeName:           "compute01",
			PartitionName:      "compute",
			StreamStartTime:    now.Add(-time.Hour * 2),
			LastEventTime:      now,
			EventCount:         2500,
			StreamStatus:       "active",
			StreamType:         "state_change",
			ConsumerID:         "consumer-node-001",
			ConsumerEndpoint:   "http://node-monitor.example.com/events",
			StreamPriority:     8,
			BufferedEvents:     75,
			ProcessedEvents:    2425,
			FailedEvents:       12,
			RetryCount:         3,
			Bandwidth:          2048000.0,
			Latency:            time.Millisecond * 25,
			CompressionRatio:   0.72,
			EventRate:          125.5,
			ConnectionQuality:  0.98,
			StreamHealth:       "healthy",
			LastHeartbeat:      now,
			BackpressureActive: false,
			FailoverActive:     false,
			QueuedEvents:       35,
			DroppedEvents:      3,
			FilterMatches:      2200,
			ValidationErrors:   8,
			EnrichmentFailures: 4,
			DeliveryAttempts:   2437,
			AckRate:            0.995,
			ResourceUsage: map[string]float64{
				"cpu":    0.15,
				"memory": 0.25,
			},
			ConfigurationHash:   "abc123def456",
			SecurityContext:     "secure",
			ComplianceFlags:     []string{"gdpr", "hipaa"},
			PerformanceProfile:  "high_throughput",
		},
		{
			StreamID:           "node-stream-002",
			NodeID:             "node-002",
			NodeName:           "gpu01",
			PartitionName:      "gpu",
			StreamStartTime:    now.Add(-time.Minute * 45),
			LastEventTime:      now.Add(-time.Minute * 2),
			EventCount:         1200,
			StreamStatus:       "warning",
			StreamType:         "maintenance",
			ConsumerID:         "consumer-node-002",
			ConsumerEndpoint:   "http://maintenance.example.com/events",
			StreamPriority:     6,
			BufferedEvents:     150,
			ProcessedEvents:    1050,
			FailedEvents:       25,
			RetryCount:         8,
			Bandwidth:          1024000.0,
			Latency:            time.Millisecond * 75,
			CompressionRatio:   0.68,
			EventRate:          75.2,
			ConnectionQuality:  0.85,
			StreamHealth:       "warning",
			LastHeartbeat:      now.Add(-time.Minute * 2),
			BackpressureActive: true,
			FailoverActive:     false,
			QueuedEvents:       125,
			DroppedEvents:      18,
			FilterMatches:      980,
			ValidationErrors:   15,
			EnrichmentFailures: 10,
			DeliveryAttempts:   1083,
			AckRate:            0.970,
			ResourceUsage: map[string]float64{
				"cpu":    0.35,
				"memory": 0.45,
			},
			ConfigurationHash:   "def456ghi789",
			SecurityContext:     "secure",
			ComplianceFlags:     []string{"sox"},
			PerformanceProfile:  "balanced",
		},
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	streamingMetrics := &NodeStreamingMetrics{
		TotalStreams:            150,
		ActiveStreams:           45,
		PausedStreams:           8,
		FailedStreams:           3,
		EventsPerSecond:         750.8,
		AverageEventLatency:     time.Millisecond * 35,
		MaxEventLatency:         time.Millisecond * 180,
		MinEventLatency:         time.Millisecond * 8,
		TotalEventsProcessed:    2800000,
		TotalEventsDropped:      280,
		TotalEventsFailed:       140,
		AverageStreamDuration:   time.Hour * 3,
		MaxStreamDuration:       time.Hour * 12,
		TotalBandwidthUsed:      20480000.0,
		CompressionEfficiency:   0.72,
		DeduplicationRate:       0.18,
		ErrorRate:               0.003,
		SuccessRate:             0.997,
		BackpressureOccurrences: 45,
		FailoverOccurrences:     6,
		ReconnectionAttempts:    28,
		MemoryUsage:             4096000000,
		CPUUsage:                0.35,
		NetworkUsage:            0.55,
		DiskUsage:               21474836480,
		CacheHitRate:            0.88,
		QueueDepth:              225,
		ProcessingEfficiency:    0.94,
		StreamingHealth:         0.97,
		NodeCoverage:            0.92,
		StateChangeAccuracy:     0.998,
		PredictionAccuracy:      0.85,
		AlertResponseTime:       time.Second * 15,
		MaintenanceCompliance:   0.96,
		SecurityIncidents:       2,
		ComplianceViolations:    1,
		PerformanceDegradation:  0.02,
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_StreamingHealth(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	healthStatus := &NodeStreamingHealthStatus{
		OverallHealth: "healthy",
		ComponentHealth: map[string]string{
			"event_processor":    "healthy",
			"stream_manager":     "healthy",
			"filter_engine":      "warning",
			"delivery_system":    "healthy",
			"maintenance_tracker": "healthy",
			"anomaly_detector":   "healthy",
		},
		LastHealthCheck:     time.Now(),
		HealthCheckDuration: time.Millisecond * 75,
		HealthScore:         0.97,
		CriticalIssues:      []string{},
		WarningIssues:       []string{"High filter processing time in GPU partition"},
		InfoMessages:        []string{"System operating normally", "Node coverage at 92%"},
		StreamingUptime:     time.Hour * 96,
		ServiceAvailability: 0.9998,
		ResourceUtilization: map[string]float64{
			"cpu":     0.35,
			"memory":  0.55,
			"disk":    0.20,
			"network": 0.45,
		},
		PerformanceMetrics: map[string]float64{
			"throughput":              750.8,
			"latency":                 0.035,
			"error_rate":              0.003,
			"state_change_accuracy":   0.998,
			"prediction_accuracy":     0.85,
		},
		ErrorSummary: map[string]int64{
			"validation_errors":   45,
			"network_errors":      8,
			"processing_errors":   18,
			"enrichment_errors":   12,
		},
		HealthTrends: map[string]float64{
			"health_score_trend":     0.015,
			"error_rate_trend":       -0.0005,
			"performance_trend":      0.008,
		},
		PredictedIssues:    []string{},
		RecommendedActions: []string{"Monitor GPU partition filter performance", "Consider scaling node monitoring capacity"},
		SystemCapacity: map[string]float64{
			"max_streams":       2000,
			"max_throughput":    5000,
			"max_connections":   1000,
			"max_nodes":         500,
		},
		AlertThresholds: map[string]float64{
			"error_rate":             0.005,
			"latency":                0.100,
			"cpu_usage":              0.80,
			"node_coverage":          0.90,
			"prediction_accuracy":    0.80,
		},
		SLACompliance: map[string]float64{
			"availability":           0.9995,
			"performance":            0.998,
			"reliability":            0.999,
			"node_coverage":          0.92,
			"maintenance_compliance": 0.96,
		},
		DependencyStatus: map[string]string{
			"slurm_api":           "healthy",
			"prometheus":          "healthy",
			"storage":             "healthy",
			"node_agents":         "healthy",
			"maintenance_system":  "healthy",
		},
		ConfigurationValid:    true,
		SecurityStatus:        "secure",
		BackupStatus:          "active",
		MonitoringEnabled:     true,
		LoggingEnabled:        true,
		MaintenanceSchedule:   []string{"Sunday 02:00-04:00", "Monthly deep maintenance"},
		CapacityForecasts: map[string]float64{
			"node_growth":        0.15,
			"stream_growth":      0.25,
			"throughput_growth":  0.20,
		},
		RiskIndicators: map[string]float64{
			"hardware_failure_risk": 0.02,
			"capacity_risk":         0.05,
			"security_risk":         0.01,
		},
		ComplianceMetrics: map[string]float64{
			"data_retention":   1.0,
			"audit_coverage":   0.98,
			"security_score":   0.97,
		},
		PerformanceBaselines: map[string]float64{
			"baseline_throughput": 700.0,
			"baseline_latency":    0.040,
			"baseline_accuracy":   0.995,
		},
		AnomalyDetectors: map[string]bool{
			"performance_anomaly": true,
			"security_anomaly":    true,
			"capacity_anomaly":    true,
		},
		AutomationStatus: map[string]bool{
			"auto_scaling":     true,
			"auto_remediation": true,
			"auto_maintenance": false,
		},
		IntegrationHealth: map[string]string{
			"grafana":      "healthy",
			"alertmanager": "healthy",
			"elk_stack":    "healthy",
		},
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(healthStatus, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_EventFilters(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	now := time.Now()
	filters := []*NodeEventFilter{
		{
			FilterID:          "node-filter-001",
			FilterName:        "Critical Node Events",
			FilterType:        "state_change",
			FilterExpression:  "state IN ['down', 'drain', 'fail']",
			IncludePattern:    "critical,urgent,down,fail",
			ExcludePattern:    "test,maintenance",
			NodeStates:        []string{"down", "drain", "fail"},
			Partitions:        []string{"compute", "gpu"},
			Nodes:             []string{"compute[01-10]", "gpu[01-05]"},
			Features:          []string{"avx2", "gpu"},
			EventTypes:        []string{"state_change", "hardware_failure"},
			Priorities:        []int{8, 9, 10},
			TimeRange:         "last_1h",
			ResourceCriteria: map[string]interface{}{
				"min_cpu":         16,
				"min_memory":      "64GB",
				"gpu_required":    true,
			},
			CustomCriteria: map[string]interface{}{
				"criticality":     "high",
				"business_impact": "severe",
			},
			FilterEnabled:     true,
			FilterPriority:    10,
			CreatedBy:         "ops-admin",
			CreatedTime:       now.Add(-time.Hour * 48),
			ModifiedBy:        "ops-admin",
			ModifiedTime:      now.Add(-time.Hour * 2),
			UsageCount:        8500,
			LastUsedTime:      now.Add(-time.Minute),
			FilterDescription: "Filter for critical node state changes requiring immediate attention",
			FilterTags:        []string{"critical", "production", "alerting"},
			ValidationRules:   []string{"syntax_check", "performance_check", "security_check"},
			MatchCount:        45000,
			FilteredCount:     28000,
			ErrorCount:        12,
			PerformanceImpact: 0.05,
			MaintenanceWindow: "Sunday 02:00-04:00",
			EmergencyBypass:   true,
			ComplianceLevel:   "high",
			AuditTrail:        []string{"created", "modified", "activated"},
			BusinessContext:   "Critical infrastructure monitoring",
			CostImplications:  125.50,
			RiskAssessment:    "low",
		},
		{
			FilterID:          "node-filter-002",
			FilterName:        "Maintenance Events",
			FilterType:        "maintenance",
			FilterExpression:  "maintenance_mode = true OR event_type = 'scheduled_maintenance'",
			IncludePattern:    "maintenance,scheduled,planned",
			ExcludePattern:    "emergency,critical",
			NodeStates:        []string{"maint", "reserved"},
			Partitions:        []string{"compute"},
			Nodes:             []string{},
			Features:          []string{},
			EventTypes:        []string{"maintenance", "scheduled_maintenance"},
			Priorities:        []int{3, 4, 5},
			TimeRange:         "last_24h",
			ResourceCriteria: map[string]interface{}{
				"maintenance_type": "scheduled",
			},
			CustomCriteria: map[string]interface{}{
				"business_hours": false,
				"impact_level":   "low",
			},
			FilterEnabled:     true,
			FilterPriority:    5,
			CreatedBy:         "maintenance-admin",
			CreatedTime:       now.Add(-time.Hour * 72),
			ModifiedBy:        "maintenance-admin",
			ModifiedTime:      now.Add(-time.Hour * 6),
			UsageCount:        5200,
			LastUsedTime:      now.Add(-time.Minute * 15),
			FilterDescription: "Filter for scheduled maintenance events and maintenance mode nodes",
			FilterTags:        []string{"maintenance", "scheduled", "non-critical"},
			ValidationRules:   []string{"syntax_check", "schedule_check"},
			MatchCount:        18000,
			FilteredCount:     12000,
			ErrorCount:        3,
			PerformanceImpact: 0.02,
			MaintenanceWindow: "Sunday 02:00-04:00",
			EmergencyBypass:   false,
			ComplianceLevel:   "medium",
			AuditTrail:        []string{"created", "modified"},
			BusinessContext:   "Planned maintenance coordination",
			CostImplications:  45.25,
			RiskAssessment:    "very_low",
		},
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return(filters, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_ProcessingStats(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	now := time.Now()
	processingStats := &NodeEventProcessingStats{
		ProcessingStartTime:     now.Add(-time.Hour * 2),
		TotalEventsReceived:     180000,
		TotalEventsProcessed:    172000,
		TotalEventsFiltered:     145000,
		TotalEventsDropped:      3500,
		TotalProcessingTime:     time.Hour * 2,
		AverageProcessingTime:   time.Millisecond * 40,
		MaxProcessingTime:       time.Millisecond * 650,
		MinProcessingTime:       time.Millisecond * 3,
		ProcessingThroughput:    23.89,
		ErrorRate:               0.018,
		SuccessRate:             0.982,
		FilterEfficiency:        0.84,
		ValidationErrors:        850,
		EnrichmentErrors:        320,
		DeliveryErrors:          480,
		TransformationErrors:    150,
		SerializationErrors:     75,
		NetworkErrors:           280,
		AuthenticationErrors:    45,
		AuthorizationErrors:     18,
		RateLimitExceeded:       125,
		BackpressureEvents:      180,
		CircuitBreakerTrips:     8,
		RetryAttempts:           2400,
		DeadLetterEvents:        85,
		DuplicateEvents:         350,
		OutOfOrderEvents:        520,
		LateArrivingEvents:      180,
		ProcessingQueues: map[string]int64{
			"main":         1500,
			"priority":     750,
			"bulk":         3000,
			"maintenance":  500,
			"dlq":          85,
		},
		WorkerStatistics: map[string]interface{}{
			"active_workers":      15,
			"idle_workers":        8,
			"blocked_workers":     3,
			"maintenance_workers": 2,
		},
		ResourceUtilization: map[string]float64{
			"cpu":     0.55,
			"memory":  0.70,
			"disk":    0.30,
			"network": 0.45,
		},
		PerformanceCounters: map[string]int64{
			"cache_hits":         145000,
			"cache_misses":       25000,
			"gc_cycles":          75,
			"context_switches":   8500,
		},
		StateChangeAccuracy:     0.998,
		EventCorrelationRate:    0.92,
		AnomalyDetectionRate:    0.15,
		MaintenancePredictions:  45,
		AlertGenerationRate:     0.08,
		ComplianceChecks:        12000,
		SecurityScans:           850,
		DataQualityScore:        0.96,
		SystemLoadImpact:        0.25,
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(processingStats, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(&NodeStreamingPerformanceMetrics{}, nil)

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

func TestNodeStateStreamingCollector_Collect_PerformanceMetrics(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	performanceMetrics := &NodeStreamingPerformanceMetrics{
		Throughput:              750.8,
		Latency:                 time.Millisecond * 35,
		P50Latency:              time.Millisecond * 25,
		P95Latency:              time.Millisecond * 85,
		P99Latency:              time.Millisecond * 165,
		MaxLatency:              time.Millisecond * 650,
		MessageRate:             680.5,
		ByteRate:                4096000.0,
		ErrorRate:               0.003,
		SuccessRate:             0.997,
		AvailabilityPercentage:  99.98,
		UptimePercentage:        99.99,
		CPUUtilization:          0.45,
		MemoryUtilization:       0.65,
		NetworkUtilization:      0.55,
		DiskUtilization:         0.25,
		ConnectionCount:         375,
		ActiveConnections:       280,
		IdleConnections:         95,
		FailedConnections:       8,
		ConnectionPoolSize:      400,
		QueueDepth:              225,
		BufferUtilization:       0.75,
		GCPressure:              0.18,
		GCFrequency:             0.12,
		GCDuration:              time.Millisecond * 35,
		HeapSize:                4294967296,
		ThreadCount:             75,
		ContextSwitches:         25000,
		SystemCalls:             45000,
		PageFaults:              850,
		CacheMisses:             1500,
		BranchMispredictions:    3200,
		InstructionsPerSecond:   1500000.0,
		CyclesPerInstruction:    2.2,
		PerformanceScore:        0.94,
		NodeCoverageEfficiency:  0.92,
		StateDetectionAccuracy:  0.998,
		EventProcessingSpeed:    750.8,
		AlertLatency:            time.Second * 12,
		RecoveryTime:            time.Minute * 8,
		MaintenanceEfficiency:   0.96,
		CompliancePerformance:   0.98,
		SecurityResponseTime:    time.Second * 5,
		PredictiveAccuracy:      0.85,
		AutomationEffectiveness: 0.91,
	}

	client.On("GetNodeStreamingConfiguration", mock.Anything).Return(&NodeStreamingConfiguration{}, nil)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream{}, nil)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return(&NodeStreamingMetrics{}, nil)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return(&NodeStreamingHealthStatus{}, nil)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription{}, nil)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter{}, nil)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return(&NodeEventProcessingStats{}, nil)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return(performanceMetrics, nil)

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

func TestNodeStateStreamingCollector_CollectError(t *testing.T) {
	client := &MockNodeStateStreamingSLURMClient{}
	collector := NewNodeStateStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetNodeStreamingConfiguration", mock.Anything).Return((*NodeStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveNodeStreams", mock.Anything).Return([]*ActiveNodeStream(nil), assert.AnError)
	client.On("GetNodeStreamingMetrics", mock.Anything).Return((*NodeStreamingMetrics)(nil), assert.AnError)
	client.On("GetNodeStreamingHealthStatus", mock.Anything).Return((*NodeStreamingHealthStatus)(nil), assert.AnError)
	client.On("GetNodeEventSubscriptions", mock.Anything).Return([]*NodeEventSubscription(nil), assert.AnError)
	client.On("GetNodeEventFilters", mock.Anything).Return([]*NodeEventFilter(nil), assert.AnError)
	client.On("GetNodeEventProcessingStats", mock.Anything).Return((*NodeEventProcessingStats)(nil), assert.AnError)
	client.On("GetNodeStreamingPerformanceMetrics", mock.Anything).Return((*NodeStreamingPerformanceMetrics)(nil), assert.AnError)

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