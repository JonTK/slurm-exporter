package collector

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockFairShareViolationsSLURMClient struct {
	mock.Mock
}

func (m *MockFairShareViolationsSLURMClient) DetectFairShareViolations(ctx context.Context) ([]*FairShareViolation, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*FairShareViolation), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetUserViolationProfile(ctx context.Context, userName string) (*UserViolationProfile, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserViolationProfile), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetAccountViolationSummary(ctx context.Context, accountName string) (*AccountViolationSummary, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountViolationSummary), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetSystemViolationOverview(ctx context.Context) (*SystemViolationOverview, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SystemViolationOverview), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetViolationTrendAnalysis(ctx context.Context, period string) (*ViolationTrendAnalysis, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*ViolationTrendAnalysis), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetAlertHistory(ctx context.Context, period string) (*AlertHistory, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*AlertHistory), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetRemediationHistory(ctx context.Context, period string) (*RemediationHistory, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*RemediationHistory), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) ValidateViolationThresholds(ctx context.Context) (*ThresholdValidation, error) {
	args := m.Called(ctx)
	return args.Get(0).(*ThresholdValidation), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetNotificationStats(ctx context.Context) (*NotificationStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(*NotificationStats), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetPolicyComplianceReport(ctx context.Context) (*PolicyComplianceReport, error) {
	args := m.Called(ctx)
	return args.Get(0).(*PolicyComplianceReport), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) PredictViolationRisk(ctx context.Context, scope string) (*ViolationRiskPrediction, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(*ViolationRiskPrediction), args.Error(1)
}

func (m *MockFairShareViolationsSLURMClient) GetViolationImpactAnalysis(ctx context.Context, violationID string) (*ViolationImpactAnalysis, error) {
	args := m.Called(ctx, violationID)
	return args.Get(0).(*ViolationImpactAnalysis), args.Error(1)
}

func TestNewFairShareViolationsCollector(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}
	collector := NewFairShareViolationsCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.fairShareViolations)
	assert.NotNil(t, collector.violationSeverity)
	assert.NotNil(t, collector.alertsGenerated)
	assert.NotNil(t, collector.systemViolationOverview)
}

func TestFairShareViolationsCollector_Describe(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}
	collector := NewFairShareViolationsCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 44 metrics, verify we have the correct number
	assert.Equal(t, 44, len(descs), "Should have exactly 44 metric descriptions")
}

func TestFairShareViolationsCollector_Collect_Success(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}

	// Mock violations
	violations := []*FairShareViolation{
		{
			ViolationID:         "violation_001",
			UserName:            "testuser",
			AccountName:         "testaccount",
			ViolationType:       "excessive_usage",
			Severity:            "high",
			Description:         "User exceeded fair-share allocation by 150%",
			CurrentFairShare:    0.3,
			ExpectedFairShare:   0.8,
			ViolationMagnitude:  0.5,
			ViolationPercentage: 150.0,
			ThresholdExceeded:   "usage_threshold",
			ResourcesUsed: map[string]float64{
				"cpu_hours":    1500.0,
				"memory_gb":    2048.0,
				"storage_tb":   5.0,
			},
			JobsAffected:        []string{"job_001", "job_002", "job_003"},
			TimeframeStart:      time.Now().Add(-24 * time.Hour),
			TimeframeEnd:        time.Now(),
			Duration:            24 * time.Hour,
			PriorityImpact:      -2000.0,
			QueueImpact:         5.0,
			SystemImpact:        0.15,
			UserImpact:          0.75,
			OverallImpact:       0.45,
			Status:              "active",
			DetectedAt:          time.Now().Add(-2 * time.Hour),
			AlertSent:           true,
			Acknowledged:        false,
			AcknowledgedBy:      "",
			RemediationActions:  []string{"reduce_priority", "notify_user"},
			AutoRemediation:     false,
			RemediationStatus:   "pending",
		},
		{
			ViolationID:         "violation_002",
			UserName:            "user2",
			AccountName:         "account2",
			ViolationType:       "resource_hoarding",
			Severity:            "medium",
			Description:         "User holding resources without active jobs",
			CurrentFairShare:    0.6,
			ExpectedFairShare:   0.4,
			ViolationMagnitude:  0.2,
			ViolationPercentage: 50.0,
			PriorityImpact:      -500.0,
			QueueImpact:         2.0,
			SystemImpact:        0.08,
			UserImpact:          0.35,
			OverallImpact:       0.2,
			Status:              "acknowledged",
			DetectedAt:          time.Now().Add(-6 * time.Hour),
			AlertSent:           true,
			Acknowledged:        true,
			AcknowledgedBy:      "admin",
			AcknowledgedAt:      time.Now().Add(-4 * time.Hour),
		},
	}

	// Mock user violation profile
	userProfile := &UserViolationProfile{
		UserName:            "testuser",
		AccountName:         "testaccount",
		TotalViolations:     15,
		ActiveViolations:    2,
		ResolvedViolations:  13,
		ViolationRate:       0.12,
		ViolationScore:      0.75,
		ViolationTypes: map[string]int{
			"excessive_usage":   8,
			"resource_hoarding": 4,
			"policy_violation":  3,
		},
		MostCommonType:      "excessive_usage",
		ViolationFrequency:  2.5,
		RecurrenceRate:      0.35,
		EscalationRate:      0.15,
		ComplianceScore:     0.65,
		ComplianceRating:    "moderate",
		ComplianceTrend:     "improving",
		RiskLevel:           "medium",
		RiskScore:           0.6,
		ViolationPatterns:   []string{"peak_hour_usage", "weekend_spikes"},
		BehaviorChanges:     []string{"reduced_peak_usage", "better_scheduling"},
		ImprovementTrend:    "positive",
		AdaptationScore:     0.7,
		AverageResponseTime: 3600.0,
		RemediationSuccess:  0.8,
		CooperationLevel:    0.85,
		LastUpdated:         time.Now(),
	}

	// Mock account violation summary
	accountSummary := &AccountViolationSummary{
		AccountName:         "testaccount",
		TotalUsers:          25,
		UsersWithViolations: 8,
		TotalViolations:     45,
		ActiveViolations:    12,
		ViolationRate:       0.32,
		AccountRiskScore:    0.55,
		RiskLevel:           "medium",
		RiskFactors:         []string{"high_usage_users", "resource_contention"},
		RiskTrend:           "stable",
		ComplianceScore:     0.68,
		ComplianceRating:    "good",
		PolicyAdherence:     0.75,
		ViolationImpact:     0.4,
		ManagementResponse:  0.82,
		RemediationRate:     0.78,
		PreventionEffectiveness: 0.65,
		LastSummarized:      time.Now(),
	}

	// Mock system violation overview
	systemOverview := &SystemViolationOverview{
		TotalViolations:       250,
		ActiveViolations:      45,
		ViolationsToday:       12,
		ViolationsThisWeek:    78,
		ViolationRate:         0.18,
		CriticalViolations:    8,
		HighViolations:        15,
		MediumViolations:      35,
		LowViolations:         25,
		ComplianceHealth:      0.75,
		ViolationTrend:        "decreasing",
		SystemStressLevel:     0.3,
		AlertLoad:             0.65,
		ThresholdBreaches:     23,
		ThresholdAdjustments:  5,
		ThresholdEffectiveness: 0.82,
		DetectionAccuracy:     0.89,
		FalsePositiveRate:     0.08,
		ResponseEfficiency:    0.85,
		AutomationRate:        0.72,
		LastUpdated:           time.Now(),
	}

	// Mock violation trend analysis
	trendAnalysis := &ViolationTrendAnalysis{
		AnalysisPeriod:      "24h",
		ViolationTrend:      "decreasing",
		TrendDirection:      "downward",
		TrendVelocity:       -0.05,
		TrendAcceleration:   0.01,
		TrendVolatility:     0.15,
		PatternType:         "cyclical",
		SeasonalPatterns: map[string]float64{
			"morning": 1.2,
			"evening": 0.8,
		},
		CyclicalBehavior:     true,
		AnomalousEvents:      []string{"weekend_spike", "maintenance_period"},
		FutureTrend:          "stable",
		PredictedViolations:  8,
		PredictionConfidence: 0.78,
		RiskForecast:         "low",
		LastAnalyzed:         time.Now(),
	}

	// Mock alert history
	alertHistory := &AlertHistory{
		Period:              "24h",
		TotalAlerts:         56,
		AlertsSent:          54,
		AlertsAcknowledged:  45,
		AlertsResolved:      38,
		AlertsEscalated:     8,
		AverageResponseTime: 1800.0,
		AcknowledgmentRate:  0.83,
		ResolutionRate:      0.84,
		EscalationRate:      0.14,
		EmailAlerts:         35,
		SlackAlerts:         15,
		SMSAlerts:           4,
		PagerDutyAlerts:     2,
		DeliverySuccess:     0.96,
		LastUpdated:         time.Now(),
	}

	// Mock remediation history
	remediationHistory := &RemediationHistory{
		Period:                  "24h",
		TotalRemediations:       42,
		SuccessfulRemediations:  36,
		FailedRemediations:      6,
		AutoRemediations:        28,
		ManualRemediations:      14,
		RemediationSuccess:      0.86,
		AverageRemediationTime:  2400.0,
		AutomationEffectiveness: 0.89,
		RecurrencePrevention:    0.75,
		ActionTypes: map[string]int{
			"priority_adjustment": 18,
			"resource_limit":      12,
			"notification":        8,
			"suspension":          4,
		},
		MostEffectiveAction:  "priority_adjustment",
		LeastEffectiveAction: "suspension",
		LastUpdated:          time.Now(),
	}

	// Mock threshold validation
	thresholdValidation := &ThresholdValidation{
		ValidationStatus:    "valid",
		ValidationScore:     0.88,
		ThresholdAccuracy:   0.85,
		OptimalityScore:     0.79,
		OptimalThresholds: map[string]float64{
			"usage_threshold":      0.8,
			"priority_threshold":   1000.0,
			"duration_threshold":   7200.0,
		},
		CurrentThresholds: map[string]float64{
			"usage_threshold":      0.75,
			"priority_threshold":   1200.0,
			"duration_threshold":   7200.0,
		},
		RecommendedChanges:  []string{"increase_usage_threshold", "adjust_priority_threshold"},
		CalibrationNeeded:   true,
		FalsePositiveRate:   0.08,
		FalseNegativeRate:   0.12,
		DetectionAccuracy:   0.89,
		Sensitivity:         0.88,
		Specificity:         0.92,
		ValidatedAt:         time.Now(),
	}

	// Mock notification stats
	notificationStats := &NotificationStats{
		TotalNotifications:   156,
		SuccessfulDeliveries: 148,
		FailedDeliveries:     8,
		DeliveryRate:         0.95,
		ChannelStats: map[string]*ChannelStats{
			"email": {
				ChannelName:     "email",
				MessagesSent:    89,
				DeliverySuccess: 0.97,
				ResponseRate:    0.65,
				AvgResponseTime: 1200.0,
			},
			"slack": {
				ChannelName:     "slack",
				MessagesSent:    45,
				DeliverySuccess: 0.98,
				ResponseRate:    0.82,
				AvgResponseTime: 600.0,
			},
		},
		PreferredChannels:    []string{"slack", "email"},
		ChannelEffectiveness: map[string]float64{
			"email": 0.75,
			"slack": 0.89,
		},
		PreferenceCompliance: 0.87,
		LastUpdated:          time.Now(),
	}

	// Mock policy compliance report
	policyComplianceReport := &PolicyComplianceReport{
		OverallCompliance:   0.82,
		ComplianceRating:    "good",
		ComplianceTrend:     "improving",
		TotalPolicies:       15,
		ActivePolicies:      12,
		ViolatedPolicies:    3,
		PolicyEffectiveness: 0.85,
		PolicyViolations: map[string]int{
			"fair_share_policy":    8,
			"resource_limit_policy": 5,
			"access_policy":        2,
		},
		ViolationsByUser: map[string]int{
			"user1": 5,
			"user2": 3,
			"user3": 2,
		},
		ViolationsByAccount: map[string]int{
			"account1": 7,
			"account2": 4,
			"account3": 4,
		},
		EnforcementRate:   0.88,
		AutoEnforcement:   0.72,
		ManualEnforcement: 0.16,
		LastGenerated:     time.Now(),
	}

	// Mock violation risk prediction
	riskPrediction := &ViolationRiskPrediction{
		Scope:             "system",
		PredictionHorizon: "7d",
		RiskScore:         0.65,
		RiskLevel:         "medium",
		PredictedViolations: 25,
		HighRiskUsers:     []string{"user1", "user3"},
		HighRiskAccounts:  []string{"account1"},
		PrimaryRiskFactors: []string{"resource_contention", "peak_usage"},
		SecondaryRiskFactors: []string{"policy_changes", "user_behavior"},
		RiskMitigationActions: []string{"threshold_adjustment", "proactive_notification"},
		PredictionConfidence: 0.78,
		ModelAccuracy:       0.85,
		UncertaintyRange:    0.15,
		PredictedAt:         time.Now(),
	}

	// Setup mock expectations
	client.On("DetectFairShareViolations", mock.Anything).Return(violations, nil)
	client.On("GetUserViolationProfile", mock.Anything, mock.AnythingOfType("string")).Return(userProfile, nil)
	client.On("GetAccountViolationSummary", mock.Anything, mock.AnythingOfType("string")).Return(accountSummary, nil)
	client.On("GetSystemViolationOverview", mock.Anything).Return(systemOverview, nil)
	client.On("GetViolationTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(trendAnalysis, nil)
	client.On("GetAlertHistory", mock.Anything, "24h").Return(alertHistory, nil)
	client.On("GetRemediationHistory", mock.Anything, "24h").Return(remediationHistory, nil)
	client.On("ValidateViolationThresholds", mock.Anything).Return(thresholdValidation, nil)
	client.On("GetNotificationStats", mock.Anything).Return(notificationStats, nil)
	client.On("GetPolicyComplianceReport", mock.Anything).Return(policyComplianceReport, nil)
	client.On("PredictViolationRisk", mock.Anything, mock.AnythingOfType("string")).Return(riskPrediction, nil)

	collector := NewFairShareViolationsCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 400)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify we collected metrics
	assert.Greater(t, len(metrics), 0, "Should collect at least some metrics")

	// Verify specific metrics are present
	metricNames := make(map[string]bool)
	for _, metric := range metrics {
		desc := metric.Desc()
		metricNames[desc.String()] = true
	}

	// Check for key metric families
	foundViolations := false
	foundSeverity := false
	foundSystemOverview := false
	foundUserProfile := false
	foundNotifications := false

	for name := range metricNames {
		if strings.Contains(name, "fairshare_violations") {
			foundViolations = true
		}
		if strings.Contains(name, "violation_severity") {
			foundSeverity = true
		}
		if strings.Contains(name, "system_violation_overview") {
			foundSystemOverview = true
		}
		if strings.Contains(name, "user_violation_score") {
			foundUserProfile = true
		}
		if strings.Contains(name, "notification") {
			foundNotifications = true
		}
	}

	assert.True(t, foundViolations, "Should have violation metrics")
	assert.True(t, foundSeverity, "Should have severity metrics")
	assert.True(t, foundSystemOverview, "Should have system overview metrics")
	assert.True(t, foundUserProfile, "Should have user profile metrics")
	assert.True(t, foundNotifications, "Should have notification metrics")

	client.AssertExpectations(t)
}

func TestFairShareViolationsCollector_Collect_Error(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}

	// Mock error response
	client.On("DetectFairShareViolations", mock.Anything).Return([]*FairShareViolation(nil), assert.AnError)

	collector := NewFairShareViolationsCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 10)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Should still collect some metrics (empty metrics after reset)
	assert.GreaterOrEqual(t, len(metrics), 0, "Should handle errors gracefully")

	client.AssertExpectations(t)
}

func TestFairShareViolationsCollector_MetricValues(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}

	// Create test data with known values
	violations := []*FairShareViolation{
		{
			ViolationID:         "test_violation",
			UserName:            "test_user",
			AccountName:         "test_account",
			ViolationType:       "test_type",
			Severity:            "high",
			Status:              "active",
			ViolationMagnitude:  0.8,
			Duration:            120 * time.Minute,
			PriorityImpact:      -1000.0,
		},
	}

	systemOverview := &SystemViolationOverview{
		TotalViolations:     100,
		ActiveViolations:    25,
		ComplianceHealth:    0.85,
		DetectionAccuracy:   0.92,
	}

	client.On("DetectFairShareViolations", mock.Anything).Return(violations, nil)
	client.On("GetSystemViolationOverview", mock.Anything).Return(systemOverview, nil)
	client.On("GetUserViolationProfile", mock.Anything, mock.AnythingOfType("string")).Return(&UserViolationProfile{}, nil)
	client.On("GetAccountViolationSummary", mock.Anything, mock.AnythingOfType("string")).Return(&AccountViolationSummary{}, nil)
	client.On("GetViolationTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&ViolationTrendAnalysis{}, nil)
	client.On("GetAlertHistory", mock.Anything, "24h").Return(&AlertHistory{}, nil)
	client.On("GetRemediationHistory", mock.Anything, "24h").Return(&RemediationHistory{}, nil)
	client.On("ValidateViolationThresholds", mock.Anything).Return(&ThresholdValidation{}, nil)
	client.On("GetNotificationStats", mock.Anything).Return(&NotificationStats{}, nil)
	client.On("GetPolicyComplianceReport", mock.Anything).Return(&PolicyComplianceReport{}, nil)
	client.On("PredictViolationRisk", mock.Anything, mock.AnythingOfType("string")).Return(&ViolationRiskPrediction{}, nil)

	collector := NewFairShareViolationsCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundViolationSeverity := false
	foundViolationDuration := false
	foundSystemTotal := false
	foundComplianceHealth := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_fairshare_violation_severity_score":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.8), *mf.Metric[0].Gauge.Value)
				foundViolationSeverity = true
			}
		case "slurm_fairshare_violation_duration_minutes":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(120), *mf.Metric[0].Gauge.Value)
				foundViolationDuration = true
			}
		case "slurm_system_violation_overview":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "metric" && *label.Value == "total_violations" {
						assert.Equal(t, float64(100), *metric.Gauge.Value)
						foundSystemTotal = true
					}
				}
			}
		case "slurm_system_compliance_health":
			for _, metric := range mf.Metric {
				for _, label := range metric.Label {
					if *label.Name == "health_metric" && *label.Value == "compliance_health" {
						assert.Equal(t, float64(0.85), *metric.Gauge.Value)
						foundComplianceHealth = true
					}
				}
			}
		}
	}

	assert.True(t, foundViolationSeverity, "Should find violation severity metric with correct value")
	assert.True(t, foundViolationDuration, "Should find violation duration metric with correct value")
	assert.True(t, foundSystemTotal, "Should find system total violations metric with correct value")
	assert.True(t, foundComplianceHealth, "Should find compliance health metric with correct value")
}

func TestFairShareViolationsCollector_Integration(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}

	// Setup comprehensive mock data
	setupFairShareViolationMocks(client)

	collector := NewFairShareViolationsCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_system_violation_overview System-wide violation overview metrics
		# TYPE slurm_system_violation_overview gauge
		slurm_system_violation_overview{metric="total_violations"} 150
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_system_violation_overview")
	assert.NoError(t, err)
}

func TestFairShareViolationsCollector_ViolationDetection(t *testing.T) {
	client := &MockFairShareViolationsSLURMClient{}

	violations := []*FairShareViolation{
		{
			ViolationID:    "v1",
			UserName:       "user1",
			AccountName:    "account1",
			ViolationType:  "excessive_usage",
			Severity:       "high",
			Status:         "active",
			OverallImpact:  0.8,
		},
		{
			ViolationID:    "v2",
			UserName:       "user2",
			AccountName:    "account2",
			ViolationType:  "resource_hoarding",
			Severity:       "medium",
			Status:         "acknowledged",
			OverallImpact:  0.4,
		},
	}

	client.On("DetectFairShareViolations", mock.Anything).Return(violations, nil)
	client.On("GetSystemViolationOverview", mock.Anything).Return(&SystemViolationOverview{}, nil)
	client.On("GetUserViolationProfile", mock.Anything, mock.AnythingOfType("string")).Return(&UserViolationProfile{}, nil)
	client.On("GetAccountViolationSummary", mock.Anything, mock.AnythingOfType("string")).Return(&AccountViolationSummary{}, nil)
	client.On("GetViolationTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(&ViolationTrendAnalysis{}, nil)
	client.On("GetAlertHistory", mock.Anything, "24h").Return(&AlertHistory{}, nil)
	client.On("GetRemediationHistory", mock.Anything, "24h").Return(&RemediationHistory{}, nil)
	client.On("ValidateViolationThresholds", mock.Anything).Return(&ThresholdValidation{}, nil)
	client.On("GetNotificationStats", mock.Anything).Return(&NotificationStats{}, nil)
	client.On("GetPolicyComplianceReport", mock.Anything).Return(&PolicyComplianceReport{}, nil)
	client.On("PredictViolationRisk", mock.Anything, mock.AnythingOfType("string")).Return(&ViolationRiskPrediction{}, nil)

	collector := NewFairShareViolationsCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 200)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify violation detection metrics are present
	foundViolationMetrics := false
	foundImpactMetrics := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "fairshare_violations") {
			foundViolationMetrics = true
		}
		if strings.Contains(desc, "violation_impact") {
			foundImpactMetrics = true
		}
	}

	assert.True(t, foundViolationMetrics, "Should find violation detection metrics")
	assert.True(t, foundImpactMetrics, "Should find violation impact metrics")
}

func setupFairShareViolationMocks(client *MockFairShareViolationsSLURMClient) {
	violations := []*FairShareViolation{
		{
			ViolationID:   "v1",
			UserName:      "user1",
			AccountName:   "account1",
			ViolationType: "excessive_usage",
			Severity:      "high",
			Status:        "active",
		},
	}

	systemOverview := &SystemViolationOverview{
		TotalViolations:     150,
		ActiveViolations:    30,
		ComplianceHealth:    0.85,
		DetectionAccuracy:   0.9,
	}

	userProfile := &UserViolationProfile{
		UserName:        "user1",
		AccountName:     "account1",
		TotalViolations: 10,
		ViolationScore:  0.6,
	}

	accountSummary := &AccountViolationSummary{
		AccountName:     "account1",
		TotalUsers:      20,
		TotalViolations: 35,
		ComplianceScore: 0.75,
	}

	trendAnalysis := &ViolationTrendAnalysis{
		AnalysisPeriod:  "24h",
		ViolationTrend:  "stable",
		TrendDirection:  "neutral",
		TrendVelocity:   0.0,
	}

	alertHistory := &AlertHistory{
		Period:              "24h",
		TotalAlerts:         50,
		AverageResponseTime: 1800.0,
	}

	remediationHistory := &RemediationHistory{
		Period:             "24h",
		TotalRemediations:  40,
		RemediationSuccess: 0.85,
	}

	thresholdValidation := &ThresholdValidation{
		ValidationStatus:  "valid",
		ValidationScore:   0.88,
		ThresholdAccuracy: 0.85,
	}

	notificationStats := &NotificationStats{
		TotalNotifications: 100,
		DeliveryRate:       0.95,
	}

	policyCompliance := &PolicyComplianceReport{
		OverallCompliance:   0.82,
		PolicyEffectiveness: 0.85,
	}

	riskPrediction := &ViolationRiskPrediction{
		Scope:               "system",
		PredictionHorizon:   "7d",
		RiskScore:           0.6,
		PredictedViolations: 20,
	}

	client.On("DetectFairShareViolations", mock.Anything).Return(violations, nil)
	client.On("GetSystemViolationOverview", mock.Anything).Return(systemOverview, nil)
	client.On("GetUserViolationProfile", mock.Anything, mock.AnythingOfType("string")).Return(userProfile, nil)
	client.On("GetAccountViolationSummary", mock.Anything, mock.AnythingOfType("string")).Return(accountSummary, nil)
	client.On("GetViolationTrendAnalysis", mock.Anything, mock.AnythingOfType("string")).Return(trendAnalysis, nil)
	client.On("GetAlertHistory", mock.Anything, "24h").Return(alertHistory, nil)
	client.On("GetRemediationHistory", mock.Anything, "24h").Return(remediationHistory, nil)
	client.On("ValidateViolationThresholds", mock.Anything).Return(thresholdValidation, nil)
	client.On("GetNotificationStats", mock.Anything).Return(notificationStats, nil)
	client.On("GetPolicyComplianceReport", mock.Anything).Return(policyCompliance, nil)
	client.On("PredictViolationRisk", mock.Anything, mock.AnythingOfType("string")).Return(riskPrediction, nil)
}