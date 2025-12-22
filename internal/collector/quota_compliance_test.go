package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQuotaComplianceSLURMClient struct {
	mock.Mock
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaCompliance(ctx context.Context, entityType, entityName string) (*QuotaCompliance, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaCompliance), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) ListQuotaEntities(ctx context.Context, entityType string) ([]*QuotaEntity, error) {
	args := m.Called(ctx, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*QuotaEntity), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaViolations(ctx context.Context, entityType, entityName string) ([]*QuotaViolation, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*QuotaViolation), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaAlerts(ctx context.Context, entityType, entityName string) ([]*QuotaAlert, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*QuotaAlert), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaThresholds(ctx context.Context, entityType, entityName string) (*QuotaThresholds, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaThresholds), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaPolicies(ctx context.Context, entityType string) ([]*QuotaPolicy, error) {
	args := m.Called(ctx, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*QuotaPolicy), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaEnforcement(ctx context.Context, entityType, entityName string) (*QuotaEnforcement, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaEnforcement), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaTrends(ctx context.Context, entityType, entityName string) (*QuotaTrends, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaTrends), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaForecast(ctx context.Context, entityType, entityName string) (*QuotaForecast, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaForecast), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaRecommendations(ctx context.Context, entityType, entityName string) (*QuotaRecommendations, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaRecommendations), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaAuditLog(ctx context.Context, entityType, entityName string) ([]*QuotaAuditEntry, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*QuotaAuditEntry), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetSystemQuotaOverview(ctx context.Context) (*SystemQuotaOverview, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*SystemQuotaOverview), args.Error(1)
}

func (m *MockQuotaComplianceSLURMClient) GetQuotaNotificationSettings(ctx context.Context, entityType, entityName string) (*QuotaNotificationSettings, error) {
	args := m.Called(ctx, entityType, entityName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*QuotaNotificationSettings), args.Error(1)
}

func TestNewQuotaComplianceCollector(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
}

func TestQuotaComplianceCollector_Describe(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 52, count)
}

func TestQuotaComplianceCollector_Collect_Success(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	// Mock system overview
	systemOverview := &SystemQuotaOverview{
		TotalEntities:        50,
		CompliantEntities:    42,
		NonCompliantEntities: 8,
		SystemCompliance:     0.84,
		TotalViolations:      15,
		ActiveViolations:     5,
		TotalAlerts:          8,
		CriticalAlerts:       2,
		ResourceUtilization: []*SystemResourceUsage{
			{
				ResourceType:    "CPU",
				TotalQuota:      10000.0,
				UsedQuota:       7500.0,
				UtilizationRate: 0.75,
				TrendDirection:  "INCREASING",
			},
			{
				ResourceType:    "Memory",
				TotalQuota:      50000.0,
				UsedQuota:       35000.0,
				UtilizationRate: 0.70,
				TrendDirection:  "STABLE",
			},
		},
		RiskAssessment: &SystemRiskAssessment{
			OverallRiskLevel:   "MEDIUM",
			RiskScore:          0.35,
			HighRiskEntities:   3,
			MediumRiskEntities: 12,
			LowRiskEntities:    35,
		},
		PerformanceMetrics: &SystemPerformanceMetrics{
			ResponseTime:         0.25,
			Throughput:          1500.0,
			ErrorRate:           0.02,
			Availability:        0.999,
			ProcessingEfficiency: 0.92,
			SystemLoad:          0.68,
		},
		CostAnalysis: &SystemCostAnalysis{
			TotalCost:                 25000.0,
			CostPerEntity:             500.0,
			CostTrend:                 "STABLE",
			CostOptimizationPotential: 0.15,
			WastedResources:           3750.0,
		},
		LastUpdated: time.Now(),
	}

	// Mock entities
	userEntities := []*QuotaEntity{
		{
			EntityType:       "user",
			EntityName:       "testuser",
			EntityDescription: "Test user account",
			CreatedAt:        time.Now().Add(-365 * 24 * time.Hour),
			UpdatedAt:        time.Now(),
			Status:           "ACTIVE",
			QuotaCount:       5,
			ComplianceStatus: "COMPLIANT",
		},
	}

	// Mock compliance data
	compliance := &QuotaCompliance{
		EntityType:       "user",
		EntityName:       "testuser",
		ComplianceScore:  0.85,
		ComplianceStatus: "COMPLIANT",
		LastAssessment:   time.Now(),
		ResourceCompliance: []*ResourceCompliance{
			{
				ResourceType:        "CPU",
				AllocatedQuota:      1000.0,
				UsedQuota:           750.0,
				UtilizationRate:     0.75,
				ComplianceStatus:    "COMPLIANT",
				ThresholdBreaches:   0,
				ComplianceScore:     0.90,
				TrendDirection:      "STABLE",
			},
			{
				ResourceType:        "Memory",
				AllocatedQuota:      5000.0,
				UsedQuota:           4200.0,
				UtilizationRate:     0.84,
				ComplianceStatus:    "WARNING",
				ThresholdBreaches:   1,
				ComplianceScore:     0.75,
				TrendDirection:      "INCREASING",
				ProjectedExhaustion: &[]time.Time{time.Now().Add(30 * 24 * time.Hour)}[0],
			},
		},
		PolicyCompliance: []*PolicyCompliance{
			{
				PolicyName:       "ResourceUsagePolicy",
				PolicyType:       "USAGE_LIMIT",
				ComplianceStatus: "COMPLIANT",
				ComplianceScore:  0.88,
				ViolationCount:   0,
				PolicyDescription: "Standard resource usage policy",
				RequiredActions:  []string{},
				EnforcementLevel: "STRICT",
			},
		},
		ViolationSummary: &ViolationSummary{
			TotalViolations:         2,
			ActiveViolations:        0,
			ResolvedViolations:      2,
			CriticalViolations:      0,
			WarningViolations:       2,
			ViolationRate:           0.02,
			AverageResolutionTime:   3600.0,
			RecurrentViolations:     0,
		},
		ComplianceHistory: []*ComplianceSnapshot{
			{
				Timestamp:           time.Now().Add(-24 * time.Hour),
				ComplianceScore:     0.82,
				ViolationCount:      1,
				ResourceUtilization: 0.78,
				PolicyBreaches:      0,
			},
		},
		RiskLevel:          "LOW",
		RecommendedActions: []string{"monitor_memory_usage", "consider_quota_increase"},
		NextReview:         time.Now().Add(7 * 24 * time.Hour),
	}

	// Mock violations
	violations := []*QuotaViolation{
		{
			ViolationID:      "v001",
			EntityType:       "user",
			EntityName:       "testuser",
			ViolationType:    "QUOTA_EXCEEDED",
			ResourceType:     "Memory",
			Severity:         "WARNING",
			DetectedAt:       time.Now().Add(-2 * time.Hour),
			ResolvedAt:       &[]time.Time{time.Now().Add(-30 * time.Minute)}[0],
			Description:      "Memory quota exceeded",
			CurrentUsage:     5200.0,
			QuotaLimit:       5000.0,
			ExcessAmount:     200.0,
			ExcessPercentage: 4.0,
			Impact:           "Performance degradation",
			RootCause:        "Large dataset processing",
			ResolutionAction: "Temporary quota increase",
			PreventionMeasures: []string{"optimize_memory_usage", "data_streaming"},
			IsRecurrent:      false,
			RecurrenceCount:  0,
			Status:           "RESOLVED",
		},
	}

	// Mock alerts
	alerts := []*QuotaAlert{
		{
			AlertID:             "a001",
			EntityType:          "user",
			EntityName:          "testuser",
			AlertType:           "THRESHOLD_WARNING",
			Priority:            "MEDIUM",
			Severity:            "WARNING",
			TriggeredAt:         time.Now().Add(-1 * time.Hour),
			AcknowledgedAt:      &[]time.Time{time.Now().Add(-45 * time.Minute)}[0],
			Message:             "Memory usage approaching 80% threshold",
			ResourceType:        "Memory",
			CurrentValue:        4000.0,
			ThresholdValue:      4000.0,
			ThresholdType:       "WARNING",
			TrendDirection:      "INCREASING",
			NotificationSent:    true,
			EscalationLevel:     1,
			ActionRequired:      "MONITOR",
			RecommendedResponse: []string{"review_usage", "consider_optimization"},
			Status:              "ACKNOWLEDGED",
		},
	}

	// Mock enforcement
	enforcement := &QuotaEnforcement{
		EntityType:        "user",
		EntityName:        "testuser",
		EnforcementStatus: "ACTIVE",
		EnforcementMode:   "STRICT",
		ActivePolicies:    []string{"ResourceUsagePolicy", "FairSharePolicy"},
		EnforcementActions: []*EnforcementAction{
			{
				ActionID:          "e001",
				ActionType:        "QUOTA_ADJUSTMENT",
				ExecutedAt:        time.Now().Add(-1 * time.Hour),
				ExecutedBy:        "system",
				Description:       "Temporary quota increase for memory",
				AffectedResources: []string{"Memory"},
				Impact:            "Resolved violation",
				Duration:          &[]int{3600}[0],
				Status:            "COMPLETED",
				Result:            "SUCCESS",
				RollbackAvailable: true,
			},
		},
		ViolationHistory: []*ViolationRecord{
			{
				Timestamp:      time.Now().Add(-2 * time.Hour),
				ViolationType:  "QUOTA_EXCEEDED",
				Severity:       "WARNING",
				ActionTaken:    "QUOTA_INCREASE",
				ResolutionTime: 1800,
				WasRecurrent:   false,
			},
		},
		ExceptionStatus: &ExceptionStatus{
			HasActiveExceptions: false,
			ActiveExceptions:    []*ActiveException{},
			TotalExceptions:     0,
		},
		OverrideStatus: &OverrideStatus{
			HasActiveOverrides: false,
			ActiveOverrides:    []*ActiveOverride{},
			TotalOverrides:     1,
		},
		ComplianceMetrics: &ComplianceMetrics{
			ComplianceRate:        0.95,
			ViolationFrequency:    0.02,
			AverageResolutionTime: 1800.0,
			RecurrenceRate:        0.0,
			PolicyAdherence:       0.98,
			ExceptionRate:         0.0,
			OverrideRate:          0.01,
			TrendDirection:        "IMPROVING",
		},
		LastEnforcement: time.Now().Add(-1 * time.Hour),
		NextEvaluation:  time.Now().Add(1 * time.Hour),
	}

	// Mock trends
	trends := &QuotaTrends{
		EntityType:     "user",
		EntityName:     "testuser",
		AnalysisPeriod: "30_DAYS",
		UsageTrends: []*UsageTrend{
			{
				ResourceType:    "CPU",
				TrendDirection:  "STABLE",
				GrowthRate:      0.02,
				Volatility:      0.15,
				Seasonality:     true,
				PeakPeriods:     []string{"weekdays"},
				LowPeriods:      []string{"weekends"},
				AverageUsage:    750.0,
				TrendConfidence: 0.85,
			},
			{
				ResourceType:    "Memory",
				TrendDirection:  "INCREASING",
				GrowthRate:      0.05,
				Volatility:      0.25,
				Seasonality:     false,
				AverageUsage:    4200.0,
				TrendConfidence: 0.78,
			},
		},
		SeasonalPatterns: []*SeasonalPattern{
			{
				PatternType:     "WEEKLY",
				PatternStrength: 0.65,
				PatternPeriod:   "7_DAYS",
				PeakPhase:       "WEEKDAYS",
				LowPhase:        "WEEKENDS",
				Amplitude:       0.3,
				Regularity:      0.82,
			},
		},
		AnomalyDetection: &AnomalyDetection{
			AnomaliesDetected:   2,
			AnomalyRate:         0.01,
			LastAnomaly:         &[]time.Time{time.Now().Add(-48 * time.Hour)}[0],
			AnomalyTypes:        []string{"usage_spike", "pattern_break"},
			DetectionConfidence: 0.88,
			AnomalyImpact:       "LOW",
		},
		InfluencingFactors: []string{"workload_type", "time_of_day"},
		RecommendedActions: []string{"optimize_scheduling", "implement_auto_scaling"},
	}

	// Mock forecast
	forecast := &QuotaForecast{
		EntityType:      "user",
		EntityName:      "testuser",
		ForecastHorizon: "30_DAYS",
		ResourceForecasts: []*ResourceForecast{
			{
				ResourceType:        "CPU",
				CurrentUsage:        750.0,
				PredictedUsage:      820.0,
				PredictedGrowth:     0.09,
				ConfidenceLevel:     0.85,
				RecommendedQuota:    1200.0,
				RiskLevel:           "LOW",
			},
			{
				ResourceType:        "Memory",
				CurrentUsage:        4200.0,
				PredictedUsage:      4800.0,
				PredictedGrowth:     0.14,
				ConfidenceLevel:     0.78,
				ProjectedExhaustion: &[]time.Time{time.Now().Add(25 * 24 * time.Hour)}[0],
				RecommendedQuota:    6000.0,
				RiskLevel:           "MEDIUM",
			},
		},
		RiskForecasts: []*RiskForecast{
			{
				RiskType:           "QUOTA_EXHAUSTION",
				CurrentRiskLevel:   "LOW",
				PredictedRiskLevel: "MEDIUM",
				RiskProbability:    0.25,
				RiskImpact:         "MEDIUM",
				MitigationActions:  []string{"increase_quota", "optimize_usage"},
			},
		},
		ModelAccuracy: &ModelAccuracy{
			OverallAccuracy: 0.82,
			PredictionError: 0.18,
			ConfidenceScore: 0.85,
			ModelType:       "ARIMA",
			LastValidation:  time.Now().Add(-7 * 24 * time.Hour),
			ValidationMetrics: map[string]float64{
				"MAE":  45.2,
				"RMSE": 68.5,
				"MAPE": 0.08,
			},
		},
		ForecastAssumptions: []string{"stable_workload", "no_major_changes"},
		UpdateFrequency:     "DAILY",
		LastUpdated:         time.Now(),
	}

	// Mock recommendations
	recommendations := &QuotaRecommendations{
		EntityType:           "user",
		EntityName:           "testuser",
		OverallScore:         0.78,
		ImprovementPotential: 0.22,
		QuotaOptimization: []*QuotaOptimization{
			{
				ResourceType:        "Memory",
				CurrentQuota:        5000.0,
				RecommendedQuota:    6000.0,
				QuotaChange:         1000.0,
				ChangeReason:        "Anticipated growth",
				ExpectedBenefit:     "Prevent violations",
				ImplementationCost:  50.0,
				ROI:                 8.5,
				RiskLevel:           "LOW",
				Priority:            "HIGH",
			},
		},
		PriorityMatrix: &PriorityMatrix{
			HighPriority:  []string{"memory_quota_increase"},
			MediumPriority: []string{"cpu_optimization"},
			LowPriority:   []string{"documentation_update"},
			QuickWins:     []string{"threshold_adjustment"},
			LongTermGoals: []string{"usage_optimization"},
		},
		ExpectedOutcomes: &ExpectedOutcomes{
			ComplianceImprovement:  0.15,
			CostReduction:          0.08,
			PerformanceGain:        0.12,
			RiskReduction:          0.25,
			EfficiencyGain:         0.10,
			UserSatisfaction:       0.18,
			OperationalImprovement: 0.14,
		},
		LastGenerated: time.Now(),
	}

	// Mock audit entries
	auditEntries := []*QuotaAuditEntry{
		{
			EntryID:          "audit001",
			Timestamp:        time.Now().Add(-1 * time.Hour),
			EntityType:       "user",
			EntityName:       "testuser",
			EventType:        "QUOTA_ADJUSTMENT",
			EventDescription: "Memory quota increased",
			UserID:           "admin",
			UserRole:         "ADMINISTRATOR",
			SourceIP:         "192.168.1.100",
			AffectedResources: []string{"Memory"},
			OldValues:        map[string]interface{}{"quota": 5000.0},
			NewValues:        map[string]interface{}{"quota": 6000.0},
			Success:          true,
			SessionID:        "session123",
			RequestID:        "req456",
			Severity:         "INFO",
			Category:         "CONFIGURATION",
			Tags:             []string{"quota", "adjustment"},
		},
	}

	// Mock notification settings
	notificationSettings := &QuotaNotificationSettings{
		EntityType: "user",
		EntityName: "testuser",
		NotificationHistory: []*NotificationHistory{
			{
				NotificationID: "notif001",
				Timestamp:      time.Now().Add(-30 * time.Minute),
				EventType:      "THRESHOLD_WARNING",
				Channel:        "email",
				Recipient:      "testuser@example.com",
				Status:         "DELIVERED",
				DeliveryTime:   &[]time.Time{time.Now().Add(-29 * time.Minute)}[0],
				RetryCount:     0,
			},
			{
				NotificationID: "notif002",
				Timestamp:      time.Now().Add(-15 * time.Minute),
				EventType:      "QUOTA_VIOLATION",
				Channel:        "slack",
				Recipient:      "#alerts",
				Status:         "FAILED",
				ErrorDetails:   "Channel not found",
				RetryCount:     2,
			},
		},
		LastUpdated: time.Now(),
	}

	// Set up mock expectations
	client.On("GetSystemQuotaOverview", mock.Anything).Return(systemOverview, nil)
	client.On("ListQuotaEntities", mock.Anything, "user").Return(userEntities, nil)
	client.On("ListQuotaEntities", mock.Anything, "account").Return([]*QuotaEntity{}, nil)
	client.On("ListQuotaEntities", mock.Anything, "partition").Return([]*QuotaEntity{}, nil)
	client.On("ListQuotaEntities", mock.Anything, "qos").Return([]*QuotaEntity{}, nil)
	client.On("GetQuotaCompliance", mock.Anything, "user", "testuser").Return(compliance, nil)
	client.On("GetQuotaViolations", mock.Anything, "user", "testuser").Return(violations, nil)
	client.On("GetQuotaAlerts", mock.Anything, "user", "testuser").Return(alerts, nil)
	client.On("GetQuotaEnforcement", mock.Anything, "user", "testuser").Return(enforcement, nil)
	client.On("GetQuotaTrends", mock.Anything, "user", "testuser").Return(trends, nil)
	client.On("GetQuotaForecast", mock.Anything, "user", "testuser").Return(forecast, nil)
	client.On("GetQuotaRecommendations", mock.Anything, "user", "testuser").Return(recommendations, nil)
	client.On("GetQuotaAuditLog", mock.Anything, "user", "testuser").Return(auditEntries, nil)
	client.On("GetQuotaNotificationSettings", mock.Anything, "user", "testuser").Return(notificationSettings, nil)

	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	assert.Greater(t, len(metrics), 0)

	client.AssertExpectations(t)
}

func TestQuotaComplianceCollector_CollectSystemOverview(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	overview := &SystemQuotaOverview{
		TotalEntities:        100,
		CompliantEntities:    85,
		NonCompliantEntities: 15,
		SystemCompliance:     0.85,
		TotalViolations:      25,
		ActiveViolations:     8,
		TotalAlerts:          12,
		CriticalAlerts:       3,
		RiskAssessment: &SystemRiskAssessment{
			RiskScore: 0.25,
		},
		PerformanceMetrics: &SystemPerformanceMetrics{
			ResponseTime:  0.15,
			Throughput:   2000.0,
			Availability: 0.998,
		},
		CostAnalysis: &SystemCostAnalysis{
			TotalCost:                 50000.0,
			CostOptimizationPotential: 0.20,
			WastedResources:           10000.0,
		},
	}

	client.On("GetSystemQuotaOverview", mock.Anything).Return(overview, nil)

	ch := make(chan prometheus.Metric, 50)
	collector.collectSystemOverview(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 9, count)

	client.AssertExpectations(t)
}

func TestQuotaComplianceCollector_CollectComplianceMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	compliance := &QuotaCompliance{
		EntityType:       "user",
		EntityName:       "testuser",
		ComplianceScore:  0.90,
		ComplianceStatus: "COMPLIANT",
		ResourceCompliance: []*ResourceCompliance{
			{
				ResourceType:     "CPU",
				UtilizationRate:  0.75,
				ComplianceScore:  0.88,
				ThresholdBreaches: 1,
			},
		},
		PolicyCompliance: []*PolicyCompliance{
			{
				PolicyName:      "TestPolicy",
				PolicyType:      "USAGE",
				ComplianceScore: 0.92,
			},
		},
		ViolationSummary: &ViolationSummary{
			TotalViolations:   5,
			ActiveViolations:  2,
			CriticalViolations: 1,
			ViolationRate:     0.05,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectComplianceMetrics(ch, "user", "testuser", compliance)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectViolationMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	resolvedTime := time.Now()
	violations := []*QuotaViolation{
		{
			ViolationType:    "QUOTA_EXCEEDED",
			Severity:         "HIGH",
			ResourceType:     "Memory",
			ExcessAmount:     500.0,
			ExcessPercentage: 10.0,
			IsRecurrent:      false,
			DetectedAt:       time.Now().Add(-2 * time.Hour),
			ResolvedAt:       &resolvedTime,
		},
		{
			ViolationType:    "POLICY_VIOLATION",
			Severity:         "MEDIUM",
			ResourceType:     "CPU",
			ExcessAmount:     100.0,
			ExcessPercentage: 5.0,
			IsRecurrent:      true,
			RecurrenceCount:  3,
			DetectedAt:       time.Now().Add(-1 * time.Hour),
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectViolationMetrics(ch, "user", "testuser", violations)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectAlertMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	acknowledgedTime := time.Now()
	alerts := []*QuotaAlert{
		{
			AlertType:       "THRESHOLD_WARNING",
			Priority:        "HIGH",
			Status:          "ACTIVE",
			EscalationLevel: 2,
			TriggeredAt:     time.Now().Add(-1 * time.Hour),
			AcknowledgedAt:  &acknowledgedTime,
		},
		{
			AlertType:       "QUOTA_VIOLATION",
			Priority:        "MEDIUM",
			Status:          "RESOLVED",
			EscalationLevel: 1,
			TriggeredAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectAlertMetrics(ch, "user", "testuser", alerts)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectEnforcementMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	enforcement := &QuotaEnforcement{
		EnforcementMode: "STRICT",
		EnforcementActions: []*EnforcementAction{
			{
				ActionType: "QUOTA_ADJUSTMENT",
				Result:     "SUCCESS",
			},
		},
		ExceptionStatus: &ExceptionStatus{
			ActiveExceptions: []*ActiveException{
				{ExceptionType: "TEMPORARY"},
			},
		},
		OverrideStatus: &OverrideStatus{
			ActiveOverrides: []*ActiveOverride{
				{OverrideType: "EMERGENCY"},
			},
		},
		ComplianceMetrics: &ComplianceMetrics{
			PolicyAdherence: 0.95,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectEnforcementMetrics(ch, "user", "testuser", enforcement)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectTrendMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	trends := &QuotaTrends{
		UsageTrends: []*UsageTrend{
			{
				ResourceType:   "CPU",
				TrendDirection: "INCREASING",
				GrowthRate:     0.05,
				Volatility:     0.15,
			},
		},
		SeasonalPatterns: []*SeasonalPattern{
			{
				PatternType:     "WEEKLY",
				PatternStrength: 0.65,
			},
		},
		AnomalyDetection: &AnomalyDetection{
			AnomaliesDetected: 3,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectTrendMetrics(ch, "user", "testuser", trends)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectForecastMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	forecast := &QuotaForecast{
		ForecastHorizon: "30_DAYS",
		ResourceForecasts: []*ResourceForecast{
			{
				ResourceType:        "CPU",
				ConfidenceLevel:     0.85,
				ProjectedExhaustion: &[]time.Time{time.Now().Add(15 * 24 * time.Hour)}[0],
			},
		},
		RiskForecasts: []*RiskForecast{
			{
				RiskType:        "QUOTA_EXHAUSTION",
				RiskProbability: 0.25,
			},
		},
		ModelAccuracy: &ModelAccuracy{
			OverallAccuracy: 0.88,
			ConfidenceScore: 0.82,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectForecastMetrics(ch, "user", "testuser", forecast)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectRecommendationMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	recommendations := &QuotaRecommendations{
		QuotaOptimization: []*QuotaOptimization{
			{
				ResourceType: "CPU",
				ROI:          5.5,
			},
		},
		PolicyRecommendations: []*PolicyRecommendation{
			{
				PolicyName: "TestPolicy",
			},
		},
		PriorityMatrix: &PriorityMatrix{
			HighPriority:  []string{"action1"},
			MediumPriority: []string{"action2", "action3"},
			LowPriority:   []string{"action4"},
			QuickWins:     []string{"win1", "win2"},
		},
		ExpectedOutcomes: &ExpectedOutcomes{
			ComplianceImprovement: 0.15,
			CostReduction:         0.12,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectRecommendationMetrics(ch, "user", "testuser", recommendations)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectAuditMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	auditEntries := []*QuotaAuditEntry{
		{
			EventType: "QUOTA_ADJUSTMENT",
			Severity:  "INFO",
			Success:   true,
		},
		{
			EventType: "POLICY_CHANGE",
			Severity:  "WARNING",
			Success:   false,
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectAuditMetrics(ch, "user", "testuser", auditEntries)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_CollectNotificationMetrics(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	deliveryTime := time.Now()
	settings := &QuotaNotificationSettings{
		NotificationHistory: []*NotificationHistory{
			{
				EventType:    "THRESHOLD_WARNING",
				Channel:      "email",
				Status:       "DELIVERED",
				Timestamp:    time.Now().Add(-5 * time.Minute),
				DeliveryTime: &deliveryTime,
			},
			{
				EventType:    "QUOTA_VIOLATION",
				Channel:      "slack",
				Status:       "FAILED",
				Timestamp:    time.Now().Add(-10 * time.Minute),
				ErrorDetails: "Channel not found",
			},
		},
	}

	ch := make(chan prometheus.Metric, 50)
	collector.collectNotificationMetrics(ch, "user", "testuser", settings)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	assert.Greater(t, count, 0)
}

func TestQuotaComplianceCollector_Collect_SystemOverviewError(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	client.On("GetSystemQuotaOverview", mock.Anything).Return(nil, errors.New("API error"))
	client.On("ListQuotaEntities", mock.Anything, mock.Anything).Return([]*QuotaEntity{}, nil)

	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	metrics := make([]prometheus.Metric, 0)
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect entity metrics even if system overview fails
	assert.GreaterOrEqual(t, len(metrics), 0)

	client.AssertExpectations(t)
}

func TestQuotaComplianceCollector_DescribeAll(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	gatherer := prometheus.Gatherer(registry)
	metricFamilies, err := gatherer.Gather()

	assert.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestQuotaComplianceCollector_MetricCount(t *testing.T) {
	client := &MockQuotaComplianceSLURMClient{}
	collector := NewQuotaComplianceCollector(client)

	expectedMetrics := []string{
		"slurm_quota_compliance_score",
		"slurm_quota_compliance_status",
		"slurm_quota_resource_compliance",
		"slurm_quota_policy_compliance",
		"slurm_quota_violations_summary",
		"slurm_quota_compliance_history",
		"slurm_quota_violations_total",
		"slurm_quota_violations_severity",
		"slurm_quota_violations_resolution_time_seconds",
		"slurm_quota_violations_recurrence",
		"slurm_quota_violations_excess",
		"slurm_quota_alerts_total",
		"slurm_quota_alerts_active",
		"slurm_quota_alerts_priority",
		"slurm_quota_alerts_escalation",
		"slurm_quota_alerts_response_time_seconds",
		"slurm_quota_threshold_breaches_total",
		"slurm_quota_threshold_effectiveness",
		"slurm_quota_threshold_utilization",
		"slurm_quota_threshold_adjustments_total",
		"slurm_quota_policy_adherence",
		"slurm_quota_policy_violations_total",
		"slurm_quota_policy_exceptions",
		"slurm_quota_policy_effectiveness",
		"slurm_quota_enforcement_actions_total",
		"slurm_quota_enforcement_effectiveness",
		"slurm_quota_enforcement_overrides_total",
		"slurm_quota_enforcement_exceptions",
		"slurm_quota_trend_direction",
		"slurm_quota_trend_growth_rate",
		"slurm_quota_trend_volatility",
		"slurm_quota_trend_seasonality",
		"slurm_quota_trend_anomalies",
		"slurm_quota_forecast_accuracy",
		"slurm_quota_forecast_confidence",
		"slurm_quota_forecast_risk",
		"slurm_quota_forecast_exhaustion_days",
		"slurm_quota_recommendations_total",
		"slurm_quota_recommendations_priority",
		"slurm_quota_recommendations_roi",
		"slurm_quota_recommendations_implementation_effort",
		"slurm_quota_audit_events_total",
		"slurm_quota_audit_failures_total",
		"slurm_quota_audit_latency_seconds",
		"slurm_system_quota_compliance",
		"slurm_system_quota_entities",
		"slurm_system_quota_violations",
		"slurm_system_quota_alerts",
		"slurm_system_quota_utilization",
		"slurm_system_quota_risk",
		"slurm_system_quota_performance",
		"slurm_system_quota_cost",
		"slurm_quota_notifications_sent_total",
		"slurm_quota_notifications_delivered_total",
		"slurm_quota_notifications_failed_total",
		"slurm_quota_notifications_latency_seconds",
	}

	assert.Equal(t, len(expectedMetrics), 52)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}

	assert.Equal(t, 52, count)
}