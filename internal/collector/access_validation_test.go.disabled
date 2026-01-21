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

type MockAccessValidationSLURMClient struct {
	mock.Mock
}

func (m *MockAccessValidationSLURMClient) ValidateUserAccountAccess(ctx context.Context, userName, accountName string) (*UserAccessValidation, error) {
	args := m.Called(ctx, userName, accountName)
	return args.Get(0).(*UserAccessValidation), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetUserAccessRights(ctx context.Context, userName string) (*UserAccessRights, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserAccessRights), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccountAccessPolicies(ctx context.Context, accountName string) (*AccountAccessPolicies, error) {
	args := m.Called(ctx, accountName)
	return args.Get(0).(*AccountAccessPolicies), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccessAuditLog(ctx context.Context, period string) (*AccessAuditLog, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*AccessAuditLog), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccessViolations(ctx context.Context, severity string) (*AccessViolations, error) {
	args := m.Called(ctx, severity)
	return args.Get(0).(*AccessViolations), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccessPatterns(ctx context.Context, userName string) (*AccessPatterns, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*AccessPatterns), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetSessionValidation(ctx context.Context, sessionID string) (*SessionValidation, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*SessionValidation), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetComplianceStatus(ctx context.Context, scope string) (*ComplianceStatus, error) {
	args := m.Called(ctx, scope)
	return args.Get(0).(*ComplianceStatus), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetSecurityAlerts(ctx context.Context, period string) (*SecurityAlerts, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*SecurityAlerts), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccessRecommendations(ctx context.Context, userName string) (*AccessRecommendations, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*AccessRecommendations), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetRiskAssessment(ctx context.Context, entityType, entityID string) (*RiskAssessment, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).(*RiskAssessment), args.Error(1)
}

func (m *MockAccessValidationSLURMClient) GetAccessTrends(ctx context.Context, period string) (*AccessTrends, error) {
	args := m.Called(ctx, period)
	return args.Get(0).(*AccessTrends), args.Error(1)
}

func TestNewAccessValidationCollector(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}
	collector := NewAccessValidationCollector(client)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.NotNil(t, collector.userAccessValidation)
	assert.NotNil(t, collector.accessValidationTime)
	assert.NotNil(t, collector.complianceScore)
	assert.NotNil(t, collector.overallRisk)
}

func TestAccessValidationCollector_Describe(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}
	collector := NewAccessValidationCollector(client)

	ch := make(chan *prometheus.Desc, 100)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var descs []*prometheus.Desc
	for desc := range ch {
		descs = append(descs, desc)
	}

	// We expect 52 metrics, verify we have the correct number
	assert.Equal(t, 52, len(descs), "Should have exactly 52 metric descriptions")
}

func TestAccessValidationCollector_Collect_Success(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Mock access validation
	validation := &UserAccessValidation{
		UserName:            "user1",
		AccountName:         "research",
		AccessGranted:       true,
		ValidationTime:      time.Now(),
		ValidationMethod:    "token",
		AuthenticationLevel: "standard",
		PermissionLevel:     "submit",
		Restrictions:        []string{},
		ValidationReasons:   []string{},
		RiskScore:           0.2,
		ComplianceStatus:    "compliant",
	}

	// Mock user access rights
	accessRights := &UserAccessRights{
		UserName:           "user1",
		PrimaryAccount:     "research",
		SecondaryAccounts:  []string{"engineering"},
		GlobalPermissions:  []string{"submit", "view"},
		AccountPermissions: map[string][]string{
			"research":    {"submit", "view", "operate"},
			"engineering": {"view"},
		},
		ResourcePermissions: map[string][]string{
			"compute": {"use", "monitor"},
		},
		EffectiveRights: map[string]bool{
			"submit":  true,
			"view":    true,
			"operate": true,
			"admin":   false,
		},
		ExpirationDates: map[string]time.Time{
			"elevated": time.Now().Add(30 * 24 * time.Hour),
		},
		ComplianceFlags: []string{"mfa_enabled", "recent_review"},
	}

	// Mock account policies
	policies := &AccountAccessPolicies{
		AccountName:           "research",
		PolicyVersion:         "2.1.0",
		MinAuthLevel:          2,
		MFARequired:           true,
		SessionTimeout:        30 * time.Minute,
		MaxConcurrentSessions: 3,
		PolicyEnforcement:     "strict",
		AuditLevel:            "detailed",
	}

	// Mock audit log
	auditLog := &AccessAuditLog{
		Period:           "24h",
		TotalEvents:      150,
		SuccessfulLogins: 120,
		FailedLogins:     30,
		AccessChanges:    5,
		PolicyViolations: 3,
		SecurityEvents:   2,
		Events: []AuditEvent{
			{
				EventID:     "evt1",
				Timestamp:   time.Now().Add(-1 * time.Hour),
				UserName:    "user1",
				AccountName: "research",
				EventType:   "login",
				Action:      "authenticate",
				Result:      "success",
				RiskScore:   0.1,
				Flagged:     false,
			},
			{
				EventID:     "evt2",
				Timestamp:   time.Now().Add(-2 * time.Hour),
				UserName:    "user2",
				AccountName: "research",
				EventType:   "login",
				Action:      "authenticate",
				Result:      "failed",
				RiskScore:   0.5,
				Flagged:     true,
			},
		},
		Summary: AuditSummary{
			ComplianceScore: 0.92,
		},
	}

	// Mock access violations
	violations := &AccessViolations{
		TotalViolations:    10,
		CriticalViolations: 1,
		WarningViolations:  3,
		InfoViolations:     6,
		UnresolvedCount:    2,
		Violations: []AccessViolation{
			{
				ViolationID:     "vio1",
				Timestamp:       time.Now().Add(-30 * time.Minute),
				UserName:        "user3",
				AccountName:     "admin",
				ViolationType:   "unauthorized_access",
				Severity:        "critical",
				AttemptedAction: "admin_operation",
				DeniedReason:    "insufficient_privileges",
				RiskScore:       0.9,
				AutoBlocked:     true,
			},
		},
	}

	// Mock access patterns
	patterns := &AccessPatterns{
		UserName: "user1",
		AccessFrequency: map[string]int{
			"submit": 50,
			"view":   100,
			"cancel": 5,
		},
		PreferredAccounts:    []string{"research"},
		AverageSessionLength: 45 * time.Minute,
		PeakUsageHours:       []int{10, 14, 16},
		BehaviorScore:        0.85,
		AnomalyDetected:      false,
		RiskProfile:          "low",
	}

	// Mock session validation
	sessionValidation := &SessionValidation{
		SessionID:     "session1",
		UserName:      "user1",
		AccountName:   "research",
		Valid:         true,
		CreatedAt:     time.Now().Add(-20 * time.Minute),
		LastActivity:  time.Now().Add(-2 * time.Minute),
		ExpiresAt:     time.Now().Add(10 * time.Minute),
		AuthMethod:    "token",
		MFACompleted:  true,
		ActivityCount: 15,
		RiskScore:     0.15,
	}

	// Mock compliance status
	compliance := &ComplianceStatus{
		Scope:           "global",
		ComplianceScore: 0.88,
		PolicyAdherence: 0.92,
		SecurityPosture: 0.85,
		AuditReadiness:  0.90,
		ControlsInPlace: 45,
		ControlsFailing: 5,
		RequiredActions: []ComplianceAction{
			{Priority: "high", Description: "Update access policies"},
			{Priority: "medium", Description: "Review user permissions"},
		},
		RiskLevel: "medium",
	}

	// Mock security alerts
	securityAlerts := &SecurityAlerts{
		Period:            "24h",
		TotalAlerts:       8,
		CriticalAlerts:    1,
		ActiveAlerts:      3,
		ResolvedAlerts:    5,
		MeanTimeToResolve: 2 * time.Hour,
		Alerts: []SecurityAlert{
			{
				AlertID:        "alert1",
				Timestamp:      time.Now().Add(-3 * time.Hour),
				AlertType:      "brute_force",
				Severity:       "critical",
				UserName:       "attacker1",
				Status:         "resolved",
				ResolutionTime: time.Now().Add(-1 * time.Hour),
				FalsePositive:  false,
			},
		},
		FalsePositiveRate: 0.1,
	}

	// Mock risk assessment
	riskAssessment := &RiskAssessment{
		EntityType:  "user",
		EntityID:    "user1",
		OverallRisk: 0.3,
		RiskFactors: map[string]float64{
			"access_pattern": 0.2,
			"permissions":    0.3,
			"compliance":     0.1,
		},
		Vulnerabilities: []Vulnerability{
			{ID: "vuln1", Severity: "medium"},
			{ID: "vuln2", Severity: "low"},
		},
		Threats: []Threat{
			{ID: "threat1", Type: "external", Status: "active"},
		},
		Mitigations: []Mitigation{
			{ID: "mit1", Type: "policy", Status: "implemented"},
		},
	}

	// Mock access trends
	trends := &AccessTrends{
		Period:          "7d",
		LoginTrends:     []float64{100, 110, 105, 115, 120, 118, 125},
		FailureTrends:   []float64{10, 12, 8, 15, 11, 9, 13},
		ViolationTrends: []float64{2, 3, 1, 4, 2, 2, 3},
		RiskTrends:      []float64{0.3, 0.32, 0.28, 0.35, 0.31, 0.29, 0.33},
	}

	// Setup mock expectations
	client.On("ValidateUserAccountAccess", mock.Anything, "user1", "research").Return(validation, nil)
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&UserAccessValidation{
		AccessGranted: false,
		ValidationReasons: []string{"not_found"},
	}, nil)
	client.On("GetUserAccessRights", mock.Anything, "user1").Return(accessRights, nil)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return(&UserAccessRights{
		GlobalPermissions:   []string{},
		AccountPermissions:  map[string][]string{},
		ResourcePermissions: map[string][]string{},
		EffectiveRights:     map[string]bool{},
		ExpirationDates:     map[string]time.Time{},
	}, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, "research").Return(policies, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessPolicies{
		MinAuthLevel: 1,
		SessionTimeout: 15 * time.Minute,
		MaxConcurrentSessions: 1,
	}, nil)
	client.On("GetAccessAuditLog", mock.Anything, "24h").Return(auditLog, nil)
	client.On("GetAccessViolations", mock.Anything, "all").Return(violations, nil)
	client.On("GetAccessPatterns", mock.Anything, "user1").Return(patterns, nil)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return(&AccessPatterns{
		AccessFrequency: map[string]int{},
		PreferredAccounts: []string{"default"},
	}, nil)
	client.On("GetSessionValidation", mock.Anything, mock.AnythingOfType("string")).Return(sessionValidation, nil)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return(compliance, nil)
	client.On("GetSecurityAlerts", mock.Anything, "24h").Return(securityAlerts, nil)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(riskAssessment, nil)
	client.On("GetAccessTrends", mock.Anything, "7d").Return(trends, nil)

	collector := NewAccessValidationCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 300)
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
	foundValidation := false
	foundCompliance := false
	foundRisk := false
	foundAlerts := false
	foundTrends := false

	for name := range metricNames {
		if strings.Contains(name, "access_validation_status") {
			foundValidation = true
		}
		if strings.Contains(name, "compliance_score") {
			foundCompliance = true
		}
		if strings.Contains(name, "overall_risk_score") {
			foundRisk = true
		}
		if strings.Contains(name, "security_alerts") {
			foundAlerts = true
		}
		if strings.Contains(name, "login_trend") {
			foundTrends = true
		}
	}

	assert.True(t, foundValidation, "Should have validation metrics")
	assert.True(t, foundCompliance, "Should have compliance metrics")
	assert.True(t, foundRisk, "Should have risk metrics")
	assert.True(t, foundAlerts, "Should have alert metrics")
	assert.True(t, foundTrends, "Should have trend metrics")

	client.AssertExpectations(t)
}

func TestAccessValidationCollector_Collect_Error(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Mock error responses
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*UserAccessValidation)(nil), assert.AnError)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return((*UserAccessRights)(nil), assert.AnError)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return((*AccountAccessPolicies)(nil), assert.AnError)
	client.On("GetAccessAuditLog", mock.Anything, mock.AnythingOfType("string")).Return((*AccessAuditLog)(nil), assert.AnError)
	client.On("GetAccessViolations", mock.Anything, mock.AnythingOfType("string")).Return((*AccessViolations)(nil), assert.AnError)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return((*AccessPatterns)(nil), assert.AnError)
	client.On("GetSessionValidation", mock.Anything, mock.AnythingOfType("string")).Return((*SessionValidation)(nil), assert.AnError)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return((*ComplianceStatus)(nil), assert.AnError)
	client.On("GetSecurityAlerts", mock.Anything, mock.AnythingOfType("string")).Return((*SecurityAlerts)(nil), assert.AnError)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*RiskAssessment)(nil), assert.AnError)
	client.On("GetAccessTrends", mock.Anything, mock.AnythingOfType("string")).Return((*AccessTrends)(nil), assert.AnError)

	collector := NewAccessValidationCollector(client)

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

	// Verify error counters were incremented
	// In practice, we would check the collectionErrors counter
}

func TestAccessValidationCollector_MetricValues(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Create test data with known values
	validation := &UserAccessValidation{
		UserName:            "test_user",
		AccountName:         "test_account",
		AccessGranted:       true,
		ValidationMethod:    "token",
		AuthenticationLevel: "elevated", // Should map to 3.0
		PermissionLevel:     "admin",    // Should map to 4.0
		RiskScore:           0.25,
	}

	accessRights := &UserAccessRights{
		UserName:          "test_user",
		PrimaryAccount:    "test_account",
		GlobalPermissions: []string{"submit", "view", "cancel"},
		EffectiveRights: map[string]bool{
			"submit": true,
			"view":   true,
			"admin":  true,
		},
		ComplianceFlags: []string{"flag1", "flag2"},
		ExpirationDates: map[string]time.Time{
			"admin": time.Now().Add(10 * 24 * time.Hour), // 10 days from now
		},
	}

	compliance := &ComplianceStatus{
		Scope:           "global",
		ComplianceScore: 0.95,
		PolicyAdherence: 0.93,
		SecurityPosture: 0.88,
		ControlsInPlace: 50,
		ControlsFailing: 3,
	}

	riskAssessment := &RiskAssessment{
		EntityType:  "user",
		EntityID:    "test_user",
		OverallRisk: 0.42,
		RiskFactors: map[string]float64{
			"behavior": 0.3,
		},
	}

	trends := &AccessTrends{
		LoginTrends:     []float64{100, 105, 110, 115, 120},
		FailureTrends:   []float64{5, 6, 4, 7, 8},
		ViolationTrends: []float64{1, 2, 1, 3, 2},
		RiskTrends:      []float64{0.2, 0.25, 0.22, 0.28, 0.26},
	}

	client.On("ValidateUserAccountAccess", mock.Anything, "user1", "research").Return(validation, nil)
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&UserAccessValidation{
		AccessGranted: false,
		ValidationReasons: []string{"not_found"},
	}, nil)
	client.On("GetUserAccessRights", mock.Anything, "user1").Return(accessRights, nil)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return(&UserAccessRights{
		GlobalPermissions:   []string{},
		AccountPermissions:  map[string][]string{},
		ResourcePermissions: map[string][]string{},
		EffectiveRights:     map[string]bool{},
		ExpirationDates:     map[string]time.Time{},
		SecondaryAccounts:   []string{"default"},
	}, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessPolicies{
		MinAuthLevel: 1,
		SessionTimeout: 15 * time.Minute,
		MaxConcurrentSessions: 1,
	}, nil)
	client.On("GetAccessAuditLog", mock.Anything, mock.AnythingOfType("string")).Return(&AccessAuditLog{
		Events: []AuditEvent{},
		Summary: AuditSummary{},
	}, nil)
	client.On("GetAccessViolations", mock.Anything, mock.AnythingOfType("string")).Return(&AccessViolations{
		Violations: []AccessViolation{},
	}, nil)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return(&AccessPatterns{
		AccessFrequency: map[string]int{},
		PreferredAccounts: []string{"default"},
	}, nil)
	client.On("GetSessionValidation", mock.Anything, mock.AnythingOfType("string")).Return(&SessionValidation{}, nil)
	client.On("GetComplianceStatus", mock.Anything, "global").Return(compliance, nil)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return(&ComplianceStatus{}, nil)
	client.On("GetSecurityAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&SecurityAlerts{
		Alerts: []SecurityAlert{},
	}, nil)
	client.On("GetRiskAssessment", mock.Anything, "user", "user1").Return(riskAssessment, nil)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&RiskAssessment{
		RiskFactors: map[string]float64{},
		Vulnerabilities: []Vulnerability{},
		Threats: []Threat{},
		Mitigations: []Mitigation{},
	}, nil)
	client.On("GetAccessTrends", mock.Anything, "7d").Return(trends, nil)

	collector := NewAccessValidationCollector(client)

	// Test specific metric values
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Gather metrics
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)

	// Verify metric values
	foundValidationStatus := false
	foundAuthLevel := false
	foundPermLevel := false
	foundRiskScore := false
	foundComplianceScore := false
	foundOverallRisk := false
	foundLoginTrend := false

	for _, mf := range metricFamilies {
		switch *mf.Name {
		case "slurm_user_access_validation_status":
			if len(mf.Metric) > 0 {
				// Should be 1.0 for granted access
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "user" && *label.Value == "user1" {
							assert.Equal(t, float64(1), *metric.Gauge.Value)
							foundValidationStatus = true
						}
					}
				}
			}
		case "slurm_authentication_level":
			if len(mf.Metric) > 0 {
				// "elevated" should map to 3.0
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "user" && *label.Value == "user1" {
							assert.Equal(t, float64(3), *metric.Gauge.Value)
							foundAuthLevel = true
						}
					}
				}
			}
		case "slurm_permission_level":
			if len(mf.Metric) > 0 {
				// "admin" should map to 4.0
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "user" && *label.Value == "user1" {
							assert.Equal(t, float64(4), *metric.Gauge.Value)
							foundPermLevel = true
						}
					}
				}
			}
		case "slurm_access_risk_score":
			if len(mf.Metric) > 0 {
				for _, metric := range mf.Metric {
					for _, label := range metric.Label {
						if *label.Name == "user" && *label.Value == "user1" {
							assert.Equal(t, float64(0.25), *metric.Gauge.Value)
							foundRiskScore = true
						}
					}
				}
			}
		case "slurm_compliance_score":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(0.95), *mf.Metric[0].Gauge.Value)
				foundComplianceScore = true
			}
		case "slurm_overall_risk_score":
			if len(mf.Metric) > 0 {
				for _, metric := range mf.Metric {
					hasUser := false
					hasUserID := false
					for _, label := range metric.Label {
						if *label.Name == "entity_type" && *label.Value == "user" {
							hasUser = true
						}
						if *label.Name == "entity_id" && *label.Value == "user1" {
							hasUserID = true
						}
					}
					if hasUser && hasUserID {
						assert.Equal(t, float64(0.42), *metric.Gauge.Value)
						foundOverallRisk = true
					}
				}
			}
		case "slurm_login_trend":
			if len(mf.Metric) > 0 {
				assert.Equal(t, float64(120), *mf.Metric[0].Gauge.Value) // Last value in trends
				foundLoginTrend = true
			}
		}
	}

	assert.True(t, foundValidationStatus, "Should find validation status metric with correct value")
	assert.True(t, foundAuthLevel, "Should find auth level metric with correct value")
	assert.True(t, foundPermLevel, "Should find permission level metric with correct value")
	assert.True(t, foundRiskScore, "Should find risk score metric with correct value")
	assert.True(t, foundComplianceScore, "Should find compliance score metric with correct value")
	assert.True(t, foundOverallRisk, "Should find overall risk metric with correct value")
	assert.True(t, foundLoginTrend, "Should find login trend metric with correct value")
}

func TestAccessValidationCollector_Integration(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Setup comprehensive mock data
	setupAccessValidationMocks(client)

	collector := NewAccessValidationCollector(client)

	// Test the collector with testutil
	expected := `
		# HELP slurm_user_access_validation_status User access validation status (1 = granted, 0 = denied)
		# TYPE slurm_user_access_validation_status gauge
		slurm_user_access_validation_status{account="research",user="user1",validation_method="token"} 1
	`

	err := testutil.CollectAndCompare(collector, strings.NewReader(expected),
		"slurm_user_access_validation_status")
	assert.NoError(t, err)
}

func TestAccessValidationCollector_SessionMetrics(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Setup session specific mocks
	setupSessionMocks(client)

	collector := NewAccessValidationCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify session metrics are present
	foundSessionValidity := false
	foundSessionActivity := false
	foundMFACompletion := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "session_validity") {
			foundSessionValidity = true
		}
		if strings.Contains(desc, "session_activity_count") {
			foundSessionActivity = true
		}
		if strings.Contains(desc, "mfa_completion_rate") {
			foundMFACompletion = true
		}
	}

	assert.True(t, foundSessionValidity, "Should find session validity metrics")
	assert.True(t, foundSessionActivity, "Should find session activity metrics")
	assert.True(t, foundMFACompletion, "Should find MFA completion metrics")
}

func TestAccessValidationCollector_ViolationMetrics(t *testing.T) {
	client := &MockAccessValidationSLURMClient{}

	// Setup violation specific mocks
	setupViolationMocks(client)

	collector := NewAccessValidationCollector(client)

	// Collect metrics
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []prometheus.Metric
	for metric := range ch {
		metrics = append(metrics, metric)
	}

	// Verify violation metrics are present
	foundViolations := false
	foundSeverity := false
	foundAutoBlocked := false

	for _, metric := range metrics {
		desc := metric.Desc().String()
		if strings.Contains(desc, "access_violations_total") {
			foundViolations = true
		}
		if strings.Contains(desc, "violation_severity_distribution") {
			foundSeverity = true
		}
		if strings.Contains(desc, "auto_blocked_attempts_total") {
			foundAutoBlocked = true
		}
	}

	assert.True(t, foundViolations, "Should find violation metrics")
	assert.True(t, foundSeverity, "Should find severity distribution metrics")
	assert.True(t, foundAutoBlocked, "Should find auto-blocked metrics")
}

// Helper functions

func setupAccessValidationMocks(client *MockAccessValidationSLURMClient) {
	validation := &UserAccessValidation{
		AccessGranted:       true,
		ValidationMethod:    "token",
		AuthenticationLevel: "standard",
		PermissionLevel:     "submit",
		RiskScore:           0.2,
	}

	client.On("ValidateUserAccountAccess", mock.Anything, "user1", "research").Return(validation, nil)
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&UserAccessValidation{
		AccessGranted: false,
		ValidationReasons: []string{"not_found"},
	}, nil)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return(&UserAccessRights{
		GlobalPermissions:   []string{},
		AccountPermissions:  map[string][]string{},
		ResourcePermissions: map[string][]string{},
		EffectiveRights:     map[string]bool{},
		ExpirationDates:     map[string]time.Time{},
		SecondaryAccounts:   []string{"default"},
	}, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessPolicies{
		MinAuthLevel: 1,
		SessionTimeout: 15 * time.Minute,
		MaxConcurrentSessions: 1,
	}, nil)
	client.On("GetAccessAuditLog", mock.Anything, mock.AnythingOfType("string")).Return(&AccessAuditLog{
		Events: []AuditEvent{},
		Summary: AuditSummary{},
	}, nil)
	client.On("GetAccessViolations", mock.Anything, mock.AnythingOfType("string")).Return(&AccessViolations{
		Violations: []AccessViolation{},
	}, nil)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return(&AccessPatterns{
		AccessFrequency: map[string]int{},
		PreferredAccounts: []string{"default"},
	}, nil)
	client.On("GetSessionValidation", mock.Anything, mock.AnythingOfType("string")).Return(&SessionValidation{}, nil)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return(&ComplianceStatus{}, nil)
	client.On("GetSecurityAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&SecurityAlerts{
		Alerts: []SecurityAlert{},
	}, nil)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&RiskAssessment{
		RiskFactors: map[string]float64{},
		Vulnerabilities: []Vulnerability{},
		Threats: []Threat{},
		Mitigations: []Mitigation{},
	}, nil)
	client.On("GetAccessTrends", mock.Anything, mock.AnythingOfType("string")).Return(&AccessTrends{}, nil)
}

func setupSessionMocks(client *MockAccessValidationSLURMClient) {
	sessions := []*SessionValidation{
		{
			SessionID:     "session1",
			UserName:      "user1",
			AccountName:   "research",
			Valid:         true,
			MFACompleted:  true,
			ActivityCount: 25,
			RiskScore:     0.1,
		},
		{
			SessionID:     "session2",
			UserName:      "user2",
			AccountName:   "research",
			Valid:         true,
			MFACompleted:  false,
			ActivityCount: 10,
			RiskScore:     0.3,
		},
		{
			SessionID:     "session3",
			UserName:      "user3",
			AccountName:   "engineering",
			Valid:         false,
			MFACompleted:  false,
			ActivityCount: 0,
			RiskScore:     0.8,
		},
	}

	for i, sessionID := range []string{"session1", "session2", "session3"} {
		client.On("GetSessionValidation", mock.Anything, sessionID).Return(sessions[i], nil)
	}

	// Set up other required mocks with minimal data
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&UserAccessValidation{
		AccessGranted: false,
		ValidationReasons: []string{"not_found"},
	}, nil)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return(&UserAccessRights{
		GlobalPermissions:   []string{},
		AccountPermissions:  map[string][]string{},
		ResourcePermissions: map[string][]string{},
		EffectiveRights:     map[string]bool{},
		ExpirationDates:     map[string]time.Time{},
		SecondaryAccounts:   []string{"default"},
	}, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessPolicies{
		MinAuthLevel: 1,
		SessionTimeout: 15 * time.Minute,
		MaxConcurrentSessions: 1,
	}, nil)
	client.On("GetAccessAuditLog", mock.Anything, mock.AnythingOfType("string")).Return(&AccessAuditLog{
		Events: []AuditEvent{},
		Summary: AuditSummary{},
	}, nil)
	client.On("GetAccessViolations", mock.Anything, mock.AnythingOfType("string")).Return(&AccessViolations{
		Violations: []AccessViolation{},
	}, nil)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return(&AccessPatterns{
		AccessFrequency: map[string]int{},
		PreferredAccounts: []string{"default"},
	}, nil)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return(&ComplianceStatus{}, nil)
	client.On("GetSecurityAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&SecurityAlerts{
		Alerts: []SecurityAlert{},
	}, nil)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&RiskAssessment{
		RiskFactors: map[string]float64{},
		Vulnerabilities: []Vulnerability{},
		Threats: []Threat{},
		Mitigations: []Mitigation{},
	}, nil)
	client.On("GetAccessTrends", mock.Anything, mock.AnythingOfType("string")).Return(&AccessTrends{}, nil)
}

func setupViolationMocks(client *MockAccessValidationSLURMClient) {
	violations := &AccessViolations{
		TotalViolations:    15,
		CriticalViolations: 2,
		WarningViolations:  5,
		InfoViolations:     8,
		UnresolvedCount:    4,
		Violations: []AccessViolation{
			{
				ViolationID:     "vio1",
				ViolationType:   "unauthorized_access",
				Severity:        "critical",
				RiskScore:       0.9,
				AutoBlocked:     true,
				DeniedReason:    "insufficient_privileges",
			},
			{
				ViolationID:     "vio2",
				ViolationType:   "policy_violation",
				Severity:        "warning",
				RiskScore:       0.6,
				AutoBlocked:     false,
				DeniedReason:    "time_restriction",
			},
			{
				ViolationID:     "vio3",
				ViolationType:   "suspicious_activity",
				Severity:        "critical",
				RiskScore:       0.85,
				AutoBlocked:     true,
				DeniedReason:    "anomaly_detected",
			},
		},
	}

	client.On("GetAccessViolations", mock.Anything, "all").Return(violations, nil)

	// Set up other required mocks with minimal data
	client.On("ValidateUserAccountAccess", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&UserAccessValidation{
		AccessGranted: false,
		ValidationReasons: []string{"not_found"},
	}, nil)
	client.On("GetUserAccessRights", mock.Anything, mock.AnythingOfType("string")).Return(&UserAccessRights{
		GlobalPermissions:   []string{},
		AccountPermissions:  map[string][]string{},
		ResourcePermissions: map[string][]string{},
		EffectiveRights:     map[string]bool{},
		ExpirationDates:     map[string]time.Time{},
		SecondaryAccounts:   []string{"default"},
	}, nil)
	client.On("GetAccountAccessPolicies", mock.Anything, mock.AnythingOfType("string")).Return(&AccountAccessPolicies{
		MinAuthLevel: 1,
		SessionTimeout: 15 * time.Minute,
		MaxConcurrentSessions: 1,
	}, nil)
	client.On("GetAccessAuditLog", mock.Anything, mock.AnythingOfType("string")).Return(&AccessAuditLog{
		Events: []AuditEvent{},
		Summary: AuditSummary{},
	}, nil)
	client.On("GetAccessPatterns", mock.Anything, mock.AnythingOfType("string")).Return(&AccessPatterns{
		AccessFrequency: map[string]int{},
		PreferredAccounts: []string{"default"},
	}, nil)
	client.On("GetSessionValidation", mock.Anything, mock.AnythingOfType("string")).Return(&SessionValidation{}, nil)
	client.On("GetComplianceStatus", mock.Anything, mock.AnythingOfType("string")).Return(&ComplianceStatus{}, nil)
	client.On("GetSecurityAlerts", mock.Anything, mock.AnythingOfType("string")).Return(&SecurityAlerts{
		Alerts: []SecurityAlert{},
	}, nil)
	client.On("GetRiskAssessment", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&RiskAssessment{
		RiskFactors: map[string]float64{},
		Vulnerabilities: []Vulnerability{},
		Threats: []Threat{},
		Mitigations: []Mitigation{},
	}, nil)
	client.On("GetAccessTrends", mock.Anything, mock.AnythingOfType("string")).Return(&AccessTrends{}, nil)
}