package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRealtimeJobStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockRealtimeJobStreamingSLURMClient) StreamJobStatusUpdates(ctx context.Context) (<-chan JobStatusEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan JobStatusEvent), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobStreamingConfiguration(ctx context.Context) (*JobStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*JobStreamingConfiguration), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetActiveJobStreams(ctx context.Context) ([]*ActiveJobStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActiveJobStream), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobEventHistory(ctx context.Context, jobID string, duration time.Duration) ([]*JobEvent, error) {
	args := m.Called(ctx, jobID, duration)
	return args.Get(0).([]*JobEvent), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobStreamingMetrics(ctx context.Context) (*JobStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*JobStreamingMetrics), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobEventFilters(ctx context.Context) ([]*JobEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*JobEventFilter), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) ConfigureJobStreaming(ctx context.Context, config *JobStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetStreamingHealthStatus(ctx context.Context) (*StreamingHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*StreamingHealthStatus), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobEventSubscriptions(ctx context.Context) ([]*JobEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*JobEventSubscription), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) ManageJobEventSubscription(ctx context.Context, subscription *JobEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobEventProcessingStats(ctx context.Context) (*JobEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*JobEventProcessingStats), args.Error(1)
}

func (m *MockRealtimeJobStreamingSLURMClient) GetJobStreamingPerformanceMetrics(ctx context.Context) (*JobStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*JobStreamingPerformanceMetrics), args.Error(1)
}

func TestNewRealtimeJobStreamingCollector(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.jobEventsTotal)
	assert.NotNil(t, collector.activeJobStreams)
	assert.NotNil(t, collector.streamingHealthScore)
}

func TestRealtimeJobStreamingCollector_Describe(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 40) // Should have many metrics
}

func TestRealtimeJobStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	config := &JobStreamingConfiguration{
		StreamingEnabled:         true,
		EventBufferSize:         10000,
		EventBatchSize:          100,
		MaxConcurrentStreams:    50,
		RateLimitPerSecond:      1000,
		BackpressureThreshold:   5000,
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
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActiveJobStream{
		{
			StreamID:           "stream-001",
			JobID:              "12345",
			UserID:             "user1",
			AccountName:        "research",
			PartitionName:      "gpu",
			StreamStartTime:    now.Add(-time.Hour),
			LastEventTime:      now,
			EventCount:         1500,
			StreamStatus:       "active",
			StreamType:         "job_status",
			ConsumerID:         "consumer-001",
			ConsumerEndpoint:   "http://consumer.example.com/events",
			StreamPriority:     5,
			BufferedEvents:     50,
			ProcessedEvents:    1450,
			FailedEvents:       5,
			RetryCount:         2,
			Bandwidth:          1024000.0,
			Latency:            time.Millisecond * 50,
			CompressionRatio:   0.65,
			EventRate:          100.5,
			ConnectionQuality:  0.95,
			StreamHealth:       "healthy",
			LastHeartbeat:      now,
			BackpressureActive: false,
			FailoverActive:     false,
			QueuedEvents:       25,
			DroppedEvents:      2,
		},
		{
			StreamID:           "stream-002",
			JobID:              "12346",
			UserID:             "user2",
			AccountName:        "production",
			PartitionName:      "cpu",
			StreamStartTime:    now.Add(-time.Minute*30),
			LastEventTime:      now.Add(-time.Minute),
			EventCount:         800,
			StreamStatus:       "warning",
			StreamType:         "job_events",
			ConsumerID:         "consumer-002",
			ConsumerEndpoint:   "http://consumer2.example.com/events",
			StreamPriority:     3,
			BufferedEvents:     100,
			ProcessedEvents:    700,
			FailedEvents:       10,
			RetryCount:         5,
			Bandwidth:          512000.0,
			Latency:            time.Millisecond * 100,
			CompressionRatio:   0.70,
			EventRate:          50.2,
			ConnectionQuality:  0.80,
			StreamHealth:       "warning",
			LastHeartbeat:      now.Add(-time.Minute),
			BackpressureActive: true,
			FailoverActive:     false,
			QueuedEvents:       75,
			DroppedEvents:      8,
		},
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	streamingMetrics := &JobStreamingMetrics{
		TotalStreams:            100,
		ActiveStreams:           25,
		PausedStreams:           5,
		FailedStreams:           2,
		EventsPerSecond:         500.5,
		AverageEventLatency:     time.Millisecond * 45,
		MaxEventLatency:         time.Millisecond * 200,
		MinEventLatency:         time.Millisecond * 10,
		TotalEventsProcessed:    1500000,
		TotalEventsDropped:      150,
		TotalEventsFailed:       75,
		AverageStreamDuration:   time.Hour * 2,
		MaxStreamDuration:       time.Hour * 8,
		TotalBandwidthUsed:      10240000.0,
		CompressionEfficiency:   0.68,
		DeduplicationRate:       0.15,
		ErrorRate:               0.005,
		SuccessRate:             0.995,
		BackpressureOccurrences: 25,
		FailoverOccurrences:     3,
		ReconnectionAttempts:    15,
		MemoryUsage:             2048000000,
		CPUUsage:                0.25,
		NetworkUsage:            0.40,
		DiskUsage:               10737418240,
		CacheHitRate:            0.85,
		QueueDepth:              150,
		ProcessingEfficiency:    0.92,
		StreamingHealth:         0.95,
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_StreamingHealth(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	healthStatus := &StreamingHealthStatus{
		OverallHealth: "healthy",
		ComponentHealth: map[string]string{
			"event_processor": "healthy",
			"stream_manager":  "healthy",
			"filter_engine":   "warning",
			"delivery_system": "healthy",
		},
		LastHealthCheck:     time.Now(),
		HealthCheckDuration: time.Millisecond * 50,
		HealthScore:         0.95,
		CriticalIssues:      []string{},
		WarningIssues:       []string{"High filter processing time"},
		InfoMessages:        []string{"System operating normally"},
		StreamingUptime:     time.Hour * 72,
		ServiceAvailability: 0.9995,
		ResourceUtilization: map[string]float64{
			"cpu":    0.25,
			"memory": 0.40,
			"disk":   0.15,
			"network": 0.30,
		},
		PerformanceMetrics: map[string]float64{
			"throughput": 500.5,
			"latency":    0.045,
			"error_rate": 0.005,
		},
		ErrorSummary: map[string]int64{
			"validation_errors":   25,
			"network_errors":      5,
			"processing_errors":   10,
		},
		HealthTrends: map[string]float64{
			"health_score_trend": 0.02,
			"error_rate_trend":   -0.001,
		},
		PredictedIssues:    []string{},
		RecommendedActions: []string{"Monitor filter performance"},
		SystemCapacity: map[string]float64{
			"max_streams":     1000,
			"max_throughput":  2000,
			"max_connections": 500,
		},
		AlertThresholds: map[string]float64{
			"error_rate":    0.01,
			"latency":       0.100,
			"cpu_usage":     0.80,
		},
		SLACompliance: map[string]float64{
			"availability": 0.999,
			"performance":  0.995,
			"reliability":  0.998,
		},
		DependencyStatus: map[string]string{
			"slurm_api":    "healthy",
			"prometheus":   "healthy",
			"storage":      "healthy",
		},
		ConfigurationValid: true,
		SecurityStatus:     "secure",
		BackupStatus:       "active",
		MonitoringEnabled:  true,
		LoggingEnabled:     true,
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(healthStatus, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_EventSubscriptions(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	now := time.Now()
	subscriptions := []*JobEventSubscription{
		{
			SubscriptionID:       "sub-001",
			SubscriberName:       "Job Monitor Service",
			SubscriberEndpoint:   "http://monitor.example.com/events",
			SubscriptionType:     "job_status",
			EventTypes:           []string{"job_submitted", "job_started", "job_completed"},
			FilterCriteria:       "account=research AND partition=gpu",
			DeliveryMethod:       "webhook",
			DeliveryFormat:       "json",
			SubscriptionStatus:   "active",
			CreatedTime:          now.Add(-time.Hour * 24),
			LastDeliveryTime:     now.Add(-time.Minute),
			DeliveryCount:        15000,
			FailedDeliveries:     25,
			RetryPolicy:          "exponential_backoff",
			MaxRetries:           5,
			RetryDelay:           time.Second * 30,
			Priority:             5,
			BatchDelivery:        true,
			BatchSize:            100,
			BatchTimeout:         time.Second * 30,
			CompressionEnabled:   true,
			EncryptionEnabled:    true,
			AuthenticationToken:  "token_abc123",
			CallbackURL:          "http://monitor.example.com/callback",
			ErrorHandling:        "retry_and_dlq",
			DeliveryGuarantee:    "at_least_once",
			Metadata: map[string]interface{}{
				"service_name": "job_monitor",
				"version":      "1.2.0",
			},
			Tags:              []string{"monitoring", "production"},
			SubscriberContact: "admin@example.com",
			BusinessContext:   "Job monitoring for research workloads",
			UsageQuota:        100000,
			UsedQuota:         75000,
			BandwidthLimit:    10240000.0,
		},
		{
			SubscriptionID:       "sub-002",
			SubscriberName:       "Alert Manager",
			SubscriberEndpoint:   "http://alerts.example.com/hooks",
			SubscriptionType:     "alerts",
			EventTypes:           []string{"job_failed", "job_timeout", "node_down"},
			FilterCriteria:       "priority=high",
			DeliveryMethod:       "webhook",
			DeliveryFormat:       "prometheus",
			SubscriptionStatus:   "active",
			CreatedTime:          now.Add(-time.Hour * 48),
			LastDeliveryTime:     now.Add(-time.Minute * 5),
			DeliveryCount:        500,
			FailedDeliveries:     5,
			RetryPolicy:          "immediate",
			MaxRetries:           3,
			RetryDelay:           time.Second * 10,
			Priority:             10,
			BatchDelivery:        false,
			BatchSize:            1,
			BatchTimeout:         time.Second * 5,
			CompressionEnabled:   false,
			EncryptionEnabled:    true,
			AuthenticationToken:  "token_xyz789",
			CallbackURL:          "http://alerts.example.com/ack",
			ErrorHandling:        "immediate_alert",
			DeliveryGuarantee:    "exactly_once",
			Metadata: map[string]interface{}{
				"service_name": "alertmanager",
				"version":      "2.1.0",
			},
			Tags:              []string{"alerting", "critical"},
			SubscriberContact: "ops@example.com",
			BusinessContext:   "Critical alert management",
			UsageQuota:        10000,
			UsedQuota:         2500,
			BandwidthLimit:    1048576.0,
		},
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return(subscriptions, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_EventFilters(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	now := time.Now()
	filters := []*JobEventFilter{
		{
			FilterID:          "filter-001",
			FilterName:        "High Priority Jobs",
			FilterType:        "priority",
			FilterExpression:  "priority >= 5000",
			IncludePattern:    "priority:high,urgent",
			ExcludePattern:    "test,debug",
			JobStates:         []string{"running", "pending"},
			Accounts:          []string{"research", "production"},
			Partitions:        []string{"gpu", "cpu"},
			Users:             []string{"user1", "user2"},
			QOSNames:          []string{"high", "normal"},
			Priorities:        []int{5000, 6000, 7000},
			TimeRange:         "last_24h",
			ResourceCriteria: map[string]interface{}{
				"min_cpu":    4,
				"min_memory": "8GB",
				"gpu_type":   "V100",
			},
			CustomCriteria: map[string]interface{}{
				"department":    "engineering",
				"project_type":  "research",
			},
			FilterEnabled:     true,
			FilterPriority:    10,
			CreatedBy:         "admin",
			CreatedTime:       now.Add(-time.Hour * 24),
			ModifiedBy:        "admin",
			ModifiedTime:      now.Add(-time.Hour),
			UsageCount:        5000,
			LastUsedTime:      now.Add(-time.Minute),
			FilterDescription: "Filter for high priority production jobs",
			FilterTags:        []string{"production", "high-priority"},
			ValidationRules:   []string{"syntax_check", "performance_check"},
			MatchCount:        25000,
			FilteredCount:     15000,
			ErrorCount:        5,
		},
		{
			FilterID:          "filter-002",
			FilterName:        "GPU Jobs",
			FilterType:        "resource",
			FilterExpression:  "gres LIKE '%gpu%'",
			IncludePattern:    "gpu,cuda,ml",
			ExcludePattern:    "test",
			JobStates:         []string{"running", "completed"},
			Accounts:          []string{"ai", "ml"},
			Partitions:        []string{"gpu"},
			Users:             []string{},
			QOSNames:          []string{"gpu"},
			Priorities:        []int{},
			TimeRange:         "last_7d",
			ResourceCriteria: map[string]interface{}{
				"gpu_count": ">= 1",
				"gpu_type":  "any",
			},
			CustomCriteria: map[string]interface{}{
				"workload_type": "ml_training",
			},
			FilterEnabled:     true,
			FilterPriority:    8,
			CreatedBy:         "ml-admin",
			CreatedTime:       now.Add(-time.Hour * 48),
			ModifiedBy:        "ml-admin",
			ModifiedTime:      now.Add(-time.Hour * 2),
			UsageCount:        3000,
			LastUsedTime:      now.Add(-time.Minute * 5),
			FilterDescription: "Filter for GPU-based machine learning jobs",
			FilterTags:        []string{"gpu", "ml", "training"},
			ValidationRules:   []string{"syntax_check", "resource_check"},
			MatchCount:        12000,
			FilteredCount:     8000,
			ErrorCount:        2,
		},
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return(filters, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_ProcessingStats(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	now := time.Now()
	processingStats := &JobEventProcessingStats{
		ProcessingStartTime:     now.Add(-time.Hour),
		TotalEventsReceived:     100000,
		TotalEventsProcessed:    95000,
		TotalEventsFiltered:     80000,
		TotalEventsDropped:      2000,
		TotalProcessingTime:     time.Hour,
		AverageProcessingTime:   time.Millisecond * 36,
		MaxProcessingTime:       time.Millisecond * 500,
		MinProcessingTime:       time.Millisecond * 5,
		ProcessingThroughput:    26.39,
		ErrorRate:               0.02,
		SuccessRate:             0.98,
		FilterEfficiency:        0.84,
		ValidationErrors:        500,
		EnrichmentErrors:        200,
		DeliveryErrors:          300,
		TransformationErrors:    100,
		SerializationErrors:     50,
		NetworkErrors:           150,
		AuthenticationErrors:    25,
		AuthorizationErrors:     10,
		RateLimitExceeded:       75,
		BackpressureEvents:      100,
		CircuitBreakerTrips:     5,
		RetryAttempts:           1500,
		DeadLetterEvents:        50,
		DuplicateEvents:         200,
		OutOfOrderEvents:        300,
		LateArrivingEvents:      100,
		ProcessingQueues: map[string]int64{
			"main":       1000,
			"priority":   500,
			"bulk":       2000,
			"dlq":        50,
		},
		WorkerStatistics: map[string]interface{}{
			"active_workers":  10,
			"idle_workers":    5,
			"blocked_workers": 2,
		},
		ResourceUtilization: map[string]float64{
			"cpu":    0.45,
			"memory": 0.60,
			"disk":   0.25,
			"network": 0.35,
		},
		PerformanceCounters: map[string]int64{
			"cache_hits":    85000,
			"cache_misses":  15000,
			"gc_cycles":     50,
		},
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(processingStats, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(&JobStreamingPerformanceMetrics{}, nil)

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

func TestRealtimeJobStreamingCollector_Collect_PerformanceMetrics(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	performanceMetrics := &JobStreamingPerformanceMetrics{
		Throughput:              500.5,
		Latency:                 time.Millisecond * 45,
		P50Latency:              time.Millisecond * 30,
		P95Latency:              time.Millisecond * 80,
		P99Latency:              time.Millisecond * 150,
		MaxLatency:              time.Millisecond * 500,
		MessageRate:             450.2,
		ByteRate:                2048000.0,
		ErrorRate:               0.005,
		SuccessRate:             0.995,
		AvailabilityPercentage:  99.95,
		UptimePercentage:        99.98,
		CPUUtilization:          0.35,
		MemoryUtilization:       0.55,
		NetworkUtilization:      0.40,
		DiskUtilization:         0.20,
		ConnectionCount:         250,
		ActiveConnections:       180,
		IdleConnections:         70,
		FailedConnections:       5,
		ConnectionPoolSize:      300,
		QueueDepth:              150,
		BufferUtilization:       0.65,
		GCPressure:              0.15,
		GCFrequency:             0.1,
		GCDuration:              time.Millisecond * 25,
		HeapSize:                2147483648,
		ThreadCount:             50,
		ContextSwitches:         15000,
		SystemCalls:             25000,
		PageFaults:              500,
		CacheMisses:             1000,
		BranchMispredictions:    2000,
		InstructionsPerSecond:   1000000.0,
		CyclesPerInstruction:    2.5,
		PerformanceScore:        0.92,
	}

	client.On("GetJobStreamingConfiguration", mock.Anything).Return(&JobStreamingConfiguration{}, nil)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream{}, nil)
	client.On("GetJobStreamingMetrics", mock.Anything).Return(&JobStreamingMetrics{}, nil)
	client.On("GetStreamingHealthStatus", mock.Anything).Return(&StreamingHealthStatus{}, nil)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription{}, nil)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter{}, nil)
	client.On("GetJobEventProcessingStats", mock.Anything).Return(&JobEventProcessingStats{}, nil)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return(performanceMetrics, nil)

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

func TestRealtimeJobStreamingCollector_CollectError(t *testing.T) {
	client := &MockRealtimeJobStreamingSLURMClient{}
	collector := NewRealtimeJobStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetJobStreamingConfiguration", mock.Anything).Return((*JobStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActiveJobStreams", mock.Anything).Return([]*ActiveJobStream(nil), assert.AnError)
	client.On("GetJobStreamingMetrics", mock.Anything).Return((*JobStreamingMetrics)(nil), assert.AnError)
	client.On("GetStreamingHealthStatus", mock.Anything).Return((*StreamingHealthStatus)(nil), assert.AnError)
	client.On("GetJobEventSubscriptions", mock.Anything).Return([]*JobEventSubscription(nil), assert.AnError)
	client.On("GetJobEventFilters", mock.Anything).Return([]*JobEventFilter(nil), assert.AnError)
	client.On("GetJobEventProcessingStats", mock.Anything).Return((*JobEventProcessingStats)(nil), assert.AnError)
	client.On("GetJobStreamingPerformanceMetrics", mock.Anything).Return((*JobStreamingPerformanceMetrics)(nil), assert.AnError)

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