package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPartitionResourceStreamingSLURMClient struct {
	mock.Mock
}

func (m *MockPartitionResourceStreamingSLURMClient) StreamPartitionResourceChanges(ctx context.Context) (<-chan PartitionResourceChangeEvent, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan PartitionResourceChangeEvent), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionStreamingConfiguration(ctx context.Context) (*PartitionStreamingConfiguration, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PartitionStreamingConfiguration), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetActivePartitionStreams(ctx context.Context) ([]*ActivePartitionStream, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*ActivePartitionStream), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionEventHistory(ctx context.Context, partitionID string, duration time.Duration) ([]*PartitionEvent, error) {
	args := m.Called(ctx, partitionID, duration)
	return args.Get(0).([]*PartitionEvent), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionStreamingMetrics(ctx context.Context) (*PartitionStreamingMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PartitionStreamingMetrics), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionEventFilters(ctx context.Context) ([]*PartitionEventFilter, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*PartitionEventFilter), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) ConfigurePartitionStreaming(ctx context.Context, config *PartitionStreamingConfiguration) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionStreamingHealthStatus(ctx context.Context) (*PartitionStreamingHealthStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PartitionStreamingHealthStatus), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionEventSubscriptions(ctx context.Context) ([]*PartitionEventSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*PartitionEventSubscription), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) ManagePartitionEventSubscription(ctx context.Context, subscription *PartitionEventSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionEventProcessingStats(ctx context.Context) (*PartitionEventProcessingStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PartitionEventProcessingStats), args.Error(1)
}

func (m *MockPartitionResourceStreamingSLURMClient) GetPartitionStreamingPerformanceMetrics(ctx context.Context) (*PartitionStreamingPerformanceMetrics, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PartitionStreamingPerformanceMetrics), args.Error(1)
}

func TestNewPartitionResourceStreamingCollector(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.partitionResourceEvents)
	assert.NotNil(t, collector.activePartitionStreams)
	assert.NotNil(t, collector.streamingHealthScore)
	assert.NotNil(t, collector.resourceUtilization)
	assert.NotNil(t, collector.capacityEfficiency)
}

func TestPartitionResourceStreamingCollector_Describe(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	ch := make(chan *prometheus.Desc, 150)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 100) // Should have many metrics
}

func TestPartitionResourceStreamingCollector_Collect_StreamingConfiguration(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	config := &PartitionStreamingConfiguration{
		StreamingEnabled:           true,
		ResourceEventBufferSize:   20000,
		ResourceEventBatchSize:    200,
		MaxConcurrentStreams:      100,
		RateLimitPerSecond:        2000,
		BackpressureThreshold:     10000,
		CompressionEnabled:        true,
		EncryptionEnabled:         true,
		AuthenticationRequired:    true,
		DeduplicationEnabled:      true,
		EventValidationEnabled:    true,
		MetricsCollectionEnabled:  true,
		DebugLoggingEnabled:       false,
		PriorityBasedStreaming:    true,
		EventEnrichmentEnabled:    true,
		FailoverEnabled:           true,
		ResourceChangeDetection:   true,
		CapacityMonitoring:        true,
		PerformanceMonitoring:     true,
		AlertIntegration:          true,
		EventCorrelation:          true,
		AnomalyDetection:          true,
		PredictiveAnalytics:       true,
		MaintenanceMode:           false,
		EmergencyProtocols:        true,
		AuditLogging:              true,
		ComplianceReporting:       true,
		ResourceOptimization:      true,
		CostTracking:              true,
		BusinessValueAnalysis:     true,
		SLAMonitoring:             true,
		CapacityPlanning:          true,
		WorkloadPrediction:        true,
		AutoScaling:               true,
		ResourceRebalancing:       true,
		UtilizationOptimization:   true,
		EfficiencyTracking:        true,
		FragmentationDetection:    true,
		IdleResourceDetection:     true,
		OversubscriptionAwareness: true,
		QueueOptimization:         true,
		FairShareEnforcement:      true,
		PriorityManagement:        true,
		PreemptionStrategy:        true,
		JobPackingOptimization:    true,
		NodeAffinityOptimization:  true,
		NetworkOptimization:       true,
		StorageOptimization:       true,
		EnergyOptimization:        true,
		CoolingOptimization:       true,
		MaintenanceOptimization:   true,
		LicenseOptimization:       true,
		SecurityMonitoring:        true,
		ComplianceChecking:        true,
		IntegrationHealthChecks:   true,
		DependencyMonitoring:      true,
		BackupValidation:          true,
		DisasterRecovery:          true,
		HighAvailability:          true,
		LoadBalancing:             true,
		ServiceDiscovery:          true,
		ConfigurationManagement:   true,
		VersionControl:            true,
		RollbackCapability:        true,
		BlueGreenDeployment:       true,
		CanaryReleases:            true,
		FeatureToggling:           true,
		A_BTestingSupport:         true,
		UserExperienceTracking:    true,
		CustomerSatisfaction:      true,
		BusinessImpactAnalysis:    true,
		ROICalculation:            true,
		CostBenefitAnalysis:       true,
		ValueStreamMapping:        true,
		ProcessOptimization:       true,
		WorkflowAutomation:        true,
		DecisionSupport:           true,
		RecommendationEngine:      true,
		IntelligentAlerting:       true,
		PredictiveMaintenance:     true,
		RootCauseAnalysis:         true,
		TrendAnalysis:             true,
		SeasonalAdjustments:       true,
		BehavioralAnalytics:       true,
		UsagePatternRecognition:   true,
		AnomalyExplanation:        true,
		CausalInference:           true,
		ImpactAssessment:          true,
		RiskManagement:            true,
		ComplianceAutomation:      true,
		AuditAutomation:           true,
		ReportGeneration:          true,
		DashboardAutomation:       true,
		NotificationManagement:    true,
		EscalationProcedures:      true,
		IncidentManagement:        true,
		ChangeManagement:          true,
		ReleaseManagement:         true,
		CapacityManagement:        true,
		AvailabilityManagement:    true,
		ContinuityManagement:      true,
		ServiceLevelManagement:    true,
		SupplierManagement:        true,
		AssetManagement:           true,
		ConfigurationAuditing:     true,
		SecurityAuditing:          true,
		PerformanceAuditing:       true,
		ComplianceAuditing:        true,
		DataGovernance:            true,
		DataQuality:               true,
		DataLineage:               true,
		DataPrivacy:               true,
		DataRetention:             true,
		DataArchiving:             true,
		DataPurging:               true,
		DataClassification:        true,
		DataEncryption:            true,
		DataMasking:               true,
		DataTokenization:          true,
		DataAnonymization:         true,
		DataPseudonymization:      true,
		ConsentManagement:         true,
		RightToErasure:            true,
		DataPortability:           true,
		TransparencyReporting:     true,
		PrivacyImpactAssessment:   true,
		LegalCompliance:           true,
		RegulatoryCompliance:      true,
		IndustryStandards:         true,
		BestPractices:             true,
		ContinuousImprovement:     true,
	}

	client.On("GetPartitionStreamingConfiguration", mock.Anything).Return(config, nil)
	client.On("GetActivePartitionStreams", mock.Anything).Return([]*ActivePartitionStream{}, nil)
	client.On("GetPartitionStreamingMetrics", mock.Anything).Return(&PartitionStreamingMetrics{}, nil)
	client.On("GetPartitionStreamingHealthStatus", mock.Anything).Return(&PartitionStreamingHealthStatus{}, nil)
	client.On("GetPartitionEventSubscriptions", mock.Anything).Return([]*PartitionEventSubscription{}, nil)
	client.On("GetPartitionEventFilters", mock.Anything).Return([]*PartitionEventFilter{}, nil)
	client.On("GetPartitionEventProcessingStats", mock.Anything).Return(&PartitionEventProcessingStats{}, nil)
	client.On("GetPartitionStreamingPerformanceMetrics", mock.Anything).Return(&PartitionStreamingPerformanceMetrics{}, nil)

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

func TestPartitionResourceStreamingCollector_Collect_ActiveStreams(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	now := time.Now()
	activeStreams := []*ActivePartitionStream{
		{
			StreamID:               "partition-stream-001",
			PartitionID:            "partition-001",
			PartitionName:          "compute",
			StreamStartTime:        now.Add(-time.Hour * 3),
			LastEventTime:          now,
			EventCount:             5500,
			StreamStatus:           "active",
			StreamType:             "resource_availability",
			ConsumerID:             "consumer-partition-001",
			ConsumerEndpoint:       "http://capacity-monitor.example.com/events",
			StreamPriority:         9,
			BufferedEvents:         125,
			ProcessedEvents:        5375,
			FailedEvents:           18,
			RetryCount:             5,
			Bandwidth:              4096000.0,
			Latency:                time.Millisecond * 30,
			CompressionRatio:       0.75,
			EventRate:              185.5,
			ConnectionQuality:      0.99,
			StreamHealth:           "healthy",
			LastHeartbeat:          now,
			BackpressureActive:     false,
			FailoverActive:         false,
			QueuedEvents:           65,
			DroppedEvents:          5,
			FilterMatches:          4800,
			ValidationErrors:       12,
			EnrichmentFailures:     6,
			DeliveryAttempts:       5393,
			AckRate:                0.997,
			ResourceUtilization:    0.85,
			CapacityUtilization:    0.78,
			EfficiencyScore:        0.92,
			FragmentationLevel:     0.15,
			OptimizationPotential:  0.18,
			CostEfficiency:         0.88,
			BusinessValueScore:     0.91,
			UserSatisfactionScore:  0.89,
			SLACompliance:          0.98,
			SecurityScore:          0.96,
			ComplianceScore:        0.97,
			ReliabilityScore:       0.95,
			PerformanceScore:       0.94,
			QualityScore:           0.93,
			ResourceUsage: map[string]float64{
				"cpu":    0.25,
				"memory": 0.40,
			},
			ConfigurationHash:    "def456ghi789abc123",
			SecurityContext:      "secure",
			ComplianceFlags:      []string{"gdpr", "hipaa", "sox"},
			PerformanceProfile:   "high_performance",
			MaintenanceWindow:    "Sunday 02:00-04:00",
			ServiceLevelTargets:  map[string]float64{"availability": 0.999, "latency": 0.050},
			BusinessContext:      "Critical compute resource monitoring",
			CostImplications:     245.75,
			RiskAssessment:       "low",
			AutomationLevel:      "high",
			MonitoringLevel:      "comprehensive",
			AlertingEnabled:      true,
			ScalingEnabled:       true,
			OptimizationEnabled:  true,
			PredictionEnabled:    true,
			AnomalyDetection:     true,
			TrendAnalysis:        true,
			CapacityPlanning:     true,
			WorkloadPrediction:   true,
			ResourceForecasting:  true,
			DemandPrediction:     true,
			UtilizationPrediction: true,
			EfficiencyPrediction: true,
			CostPrediction:       true,
			ValuePrediction:      true,
			RiskPrediction:       true,
			ImpactAnalysis:       true,
		},
		{
			StreamID:               "partition-stream-002",
			PartitionID:            "partition-002",
			PartitionName:          "gpu",
			StreamStartTime:        now.Add(-time.Hour * 1),
			LastEventTime:          now.Add(-time.Minute * 3),
			EventCount:             1800,
			StreamStatus:           "warning",
			StreamType:             "resource_optimization",
			ConsumerID:             "consumer-partition-002",
			ConsumerEndpoint:       "http://gpu-optimizer.example.com/events",
			StreamPriority:         8,
			BufferedEvents:         200,
			ProcessedEvents:        1600,
			FailedEvents:           35,
			RetryCount:             12,
			Bandwidth:              2048000.0,
			Latency:                time.Millisecond * 85,
			CompressionRatio:       0.70,
			EventRate:              95.2,
			ConnectionQuality:      0.88,
			StreamHealth:           "warning",
			LastHeartbeat:          now.Add(-time.Minute * 3),
			BackpressureActive:     true,
			FailoverActive:         false,
			QueuedEvents:           175,
			DroppedEvents:          25,
			FilterMatches:          1450,
			ValidationErrors:       22,
			EnrichmentFailures:     13,
			DeliveryAttempts:       1635,
			AckRate:                0.985,
			ResourceUtilization:    0.95,
			CapacityUtilization:    0.88,
			EfficiencyScore:        0.78,
			FragmentationLevel:     0.35,
			OptimizationPotential:  0.45,
			CostEfficiency:         0.72,
			BusinessValueScore:     0.82,
			UserSatisfactionScore:  0.75,
			SLACompliance:          0.92,
			SecurityScore:          0.94,
			ComplianceScore:        0.95,
			ReliabilityScore:       0.87,
			PerformanceScore:       0.85,
			QualityScore:           0.83,
			ResourceUsage: map[string]float64{
				"cpu":    0.55,
				"memory": 0.75,
			},
			ConfigurationHash:    "ghi789abc123def456",
			SecurityContext:      "secure",
			ComplianceFlags:      []string{"gdpr"},
			PerformanceProfile:   "balanced",
			MaintenanceWindow:    "Sunday 02:00-04:00",
			ServiceLevelTargets:  map[string]float64{"availability": 0.995, "latency": 0.100},
			BusinessContext:      "GPU resource optimization monitoring",
			CostImplications:     385.50,
			RiskAssessment:       "medium",
			AutomationLevel:      "medium",
			MonitoringLevel:      "standard",
			AlertingEnabled:      true,
			ScalingEnabled:       true,
			OptimizationEnabled:  true,
			PredictionEnabled:    true,
			AnomalyDetection:     true,
			TrendAnalysis:        true,
			CapacityPlanning:     true,
			WorkloadPrediction:   true,
			ResourceForecasting:  true,
			DemandPrediction:     true,
			UtilizationPrediction: true,
			EfficiencyPrediction: true,
			CostPrediction:       true,
			ValuePrediction:      true,
			RiskPrediction:       true,
			ImpactAnalysis:       true,
		},
	}

	client.On("GetPartitionStreamingConfiguration", mock.Anything).Return(&PartitionStreamingConfiguration{}, nil)
	client.On("GetActivePartitionStreams", mock.Anything).Return(activeStreams, nil)
	client.On("GetPartitionStreamingMetrics", mock.Anything).Return(&PartitionStreamingMetrics{}, nil)
	client.On("GetPartitionStreamingHealthStatus", mock.Anything).Return(&PartitionStreamingHealthStatus{}, nil)
	client.On("GetPartitionEventSubscriptions", mock.Anything).Return([]*PartitionEventSubscription{}, nil)
	client.On("GetPartitionEventFilters", mock.Anything).Return([]*PartitionEventFilter{}, nil)
	client.On("GetPartitionEventProcessingStats", mock.Anything).Return(&PartitionEventProcessingStats{}, nil)
	client.On("GetPartitionStreamingPerformanceMetrics", mock.Anything).Return(&PartitionStreamingPerformanceMetrics{}, nil)

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

func TestPartitionResourceStreamingCollector_Collect_StreamingMetrics(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	streamingMetrics := &PartitionStreamingMetrics{
		TotalStreams:                    200,
		ActiveStreams:                   75,
		PausedStreams:                   12,
		FailedStreams:                   5,
		ResourceEventsPerSecond:         950.5,
		AverageResourceEventLatency:     time.Millisecond * 28,
		MaxResourceEventLatency:         time.Millisecond * 150,
		MinResourceEventLatency:         time.Millisecond * 5,
		TotalResourceEventsProcessed:    4500000,
		TotalResourceEventsDropped:     450,
		TotalResourceEventsFailed:      225,
		AverageStreamDuration:           time.Hour * 4,
		MaxStreamDuration:               time.Hour * 16,
		TotalBandwidthUsed:              40960000.0,
		CompressionEfficiency:           0.74,
		DeduplicationRate:               0.22,
		ErrorRate:                       0.002,
		SuccessRate:                     0.998,
		BackpressureOccurrences:         65,
		FailoverOccurrences:             8,
		ReconnectionAttempts:            45,
		MemoryUsage:                     8192000000,
		CPUUsage:                        0.42,
		NetworkUsage:                    0.58,
		DiskUsage:                       42949672960,
		CacheHitRate:                    0.91,
		QueueDepth:                      350,
		ProcessingEfficiency:            0.96,
		StreamingHealth:                 0.98,
		PartitionCoverage:               0.95,
		ResourceChangeAccuracy:          0.999,
		CapacityPredictionAccuracy:      0.88,
		AlertResponseTime:               time.Second * 12,
		ComplianceScore:                 0.97,
		SecurityIncidents:               1,
		ComplianceViolations:            0,
		PerformanceDegradation:          0.015,
		ResourceUtilizationEfficiency:   0.92,
		CapacityOptimizationScore:       0.89,
		CostEfficiencyScore:             0.85,
		BusinessValueScore:              0.93,
		UserSatisfactionScore:           0.91,
		SLAComplianceScore:              0.96,
		ReliabilityScore:                0.98,
		AvailabilityScore:               0.995,
		PerformanceScore:                0.94,
		QualityScore:                    0.95,
		EfficiencyScore:                 0.93,
		OptimizationScore:               0.87,
		InnovationScore:                 0.78,
		AdaptabilityScore:               0.82,
		ScalabilityScore:                0.89,
		SustainabilityScore:             0.75,
		FragmentationLevel:              0.18,
		IdleResourcePercentage:          0.12,
		OversubscriptionLevel:           0.05,
		LoadBalancingEfficiency:         0.94,
		SchedulingEfficiency:            0.91,
		AllocationAccuracy:              0.97,
		DeallocationSpeed:               0.95,
		ResourceTurnoverRate:            0.25,
		CapacityGrowthRate:              0.15,
		DemandGrowthRate:                0.22,
		UtilizationTrendSlope:           0.08,
		EfficiencyTrendSlope:            0.05,
		CostTrendSlope:                  -0.03,
		ValueTrendSlope:                 0.12,
		RiskTrendSlope:                  -0.02,
		ThroughputTrendSlope:            0.15,
		LatencyTrendSlope:               -0.08,
		ErrorTrendSlope:                 -0.05,
		MaintenanceCompliance:           0.98,
		SecurityCompliance:              0.96,
		DataGovernanceCompliance:        0.94,
		PrivacyCompliance:               0.93,
		RegulatoryCompliance:            0.95,
		IndustryStandardsCompliance:     0.92,
		BestPracticesAdherence:          0.88,
		ContinuousImprovementScore:      0.85,
		InnovationAdoptionRate:          0.65,
		TechnologyMatureness:            0.78,
		DigitalTransformationProgress:   0.72,
		AutomationLevel:                 0.82,
		IntegrationHealthScore:          0.94,
		DependencyStabilityScore:        0.91,
		InteroperabilityScore:           0.89,
		StandardizationLevel:            0.86,
		ModularityScore:                 0.83,
		ReusabilityScore:                0.80,
		MaintainabilityScore:            0.87,
		TestabilityScore:                0.84,
		DeployabilityScore:              0.88,
		MonitorabilityScore:             0.92,
		ObservabilityScore:              0.90,
		DebuggabilityScore:              0.85,
		TroubleshootingEfficiency:       0.83,
		IncidentResponseTime:            time.Minute * 5,
		ProblemResolutionTime:           time.Hour * 2,
		ChangeSuccessRate:               0.94,
		RollbackSuccessRate:             0.98,
		DeploymentFrequency:             12.5,
		LeadTimeForChanges:              time.Hour * 6,
		TimeToRestore:                   time.Minute * 15,
		ChangeFailureRate:               0.03,
	}

	client.On("GetPartitionStreamingConfiguration", mock.Anything).Return(&PartitionStreamingConfiguration{}, nil)
	client.On("GetActivePartitionStreams", mock.Anything).Return([]*ActivePartitionStream{}, nil)
	client.On("GetPartitionStreamingMetrics", mock.Anything).Return(streamingMetrics, nil)
	client.On("GetPartitionStreamingHealthStatus", mock.Anything).Return(&PartitionStreamingHealthStatus{}, nil)
	client.On("GetPartitionEventSubscriptions", mock.Anything).Return([]*PartitionEventSubscription{}, nil)
	client.On("GetPartitionEventFilters", mock.Anything).Return([]*PartitionEventFilter{}, nil)
	client.On("GetPartitionEventProcessingStats", mock.Anything).Return(&PartitionEventProcessingStats{}, nil)
	client.On("GetPartitionStreamingPerformanceMetrics", mock.Anything).Return(&PartitionStreamingPerformanceMetrics{}, nil)

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

func TestPartitionResourceStreamingCollector_CollectError(t *testing.T) {
	client := &MockPartitionResourceStreamingSLURMClient{}
	collector := NewPartitionResourceStreamingCollector(client)

	// Simulate errors from all client methods
	client.On("GetPartitionStreamingConfiguration", mock.Anything).Return((*PartitionStreamingConfiguration)(nil), assert.AnError)
	client.On("GetActivePartitionStreams", mock.Anything).Return([]*ActivePartitionStream(nil), assert.AnError)
	client.On("GetPartitionStreamingMetrics", mock.Anything).Return((*PartitionStreamingMetrics)(nil), assert.AnError)
	client.On("GetPartitionStreamingHealthStatus", mock.Anything).Return((*PartitionStreamingHealthStatus)(nil), assert.AnError)
	client.On("GetPartitionEventSubscriptions", mock.Anything).Return([]*PartitionEventSubscription(nil), assert.AnError)
	client.On("GetPartitionEventFilters", mock.Anything).Return([]*PartitionEventFilter(nil), assert.AnError)
	client.On("GetPartitionEventProcessingStats", mock.Anything).Return((*PartitionEventProcessingStats)(nil), assert.AnError)
	client.On("GetPartitionStreamingPerformanceMetrics", mock.Anything).Return((*PartitionStreamingPerformanceMetrics)(nil), assert.AnError)

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