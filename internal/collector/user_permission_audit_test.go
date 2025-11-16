package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserPermissionAuditSLURMClient struct {
	mock.Mock
}

func (m *MockUserPermissionAuditSLURMClient) GetUserPermissions(ctx context.Context, userName string) (*UserPermissions, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserPermissions), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserAccessAuditLog(ctx context.Context, userName string, startTime, endTime time.Time) ([]*UserAccessAuditEntry, error) {
	args := m.Called(ctx, userName, startTime, endTime)
	return args.Get(0).([]*UserAccessAuditEntry), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserAccessPatterns(ctx context.Context, userName string) (*UserAccessPatterns, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserAccessPatterns), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserPermissionHistory(ctx context.Context, userName string, startTime, endTime time.Time) ([]*UserPermissionHistoryEntry, error) {
	args := m.Called(ctx, userName, startTime, endTime)
	return args.Get(0).([]*UserPermissionHistoryEntry), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserAccessViolations(ctx context.Context, userName string) ([]*UserAccessViolation, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserAccessViolation), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserSessionAudit(ctx context.Context, userName string) (*UserSessionAudit, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserSessionAudit), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserComplianceStatus(ctx context.Context, userName string) (*UserComplianceStatus, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserComplianceStatus), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserSecurityEvents(ctx context.Context, userName string) ([]*UserSecurityEvent, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserSecurityEvent), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserRoleAssignments(ctx context.Context, userName string) ([]*UserRoleAssignment, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserRoleAssignment), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserPrivilegeEscalations(ctx context.Context, userName string) ([]*UserPrivilegeEscalation, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserPrivilegeEscalation), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserDataAccess(ctx context.Context, userName string) (*UserDataAccess, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserDataAccess), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserAccountAssociations(ctx context.Context, userName string) ([]*UserAccountAssociation, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserAccountAssociation), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserAuthenticationEvents(ctx context.Context, userName string) ([]*UserAuthenticationEvent, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).([]*UserAuthenticationEvent), args.Error(1)
}

func (m *MockUserPermissionAuditSLURMClient) GetUserResourceAccess(ctx context.Context, userName string) (*UserResourceAccess, error) {
	args := m.Called(ctx, userName)
	return args.Get(0).(*UserResourceAccess), args.Error(1)
}

func TestUserPermissionAuditCollector_PermissionMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	permissions := &UserPermissions{
		UserName:                "test-user",
		ActivePermissions:       []string{"read", "write", "execute"},
		InheritedPermissions:    []string{"read", "list"},
		ExplicitPermissions:     []string{"write", "execute"},
		DeniedPermissions:       []string{"admin", "delete"},
		TemporaryPermissions:    []string{"debug"},
		ConditionalPermissions:  []string{"backup"},
		PermissionLevel:         "advanced",
		PermissionScope:         "departmental",
		RiskLevel:              "medium",
		ComplianceStatus:       "compliant",
		ValidationStatus:       "valid",
		LastValidated:          time.Now(),
		LastModified:           time.Now(),
	}

	accessPatterns := &UserAccessPatterns{
		UserName:               "test-user",
		LoginFrequency:         5.2,
		AverageSessionDuration: 2 * time.Hour,
		AnomalyCount:           3,
		PatternStability:       0.85,
		UsageVariability:       0.25,
		RiskIndicators:         []string{"unusual_hours", "new_location"},
		EfficiencyMetrics:      map[string]float64{"overall": 0.78},
	}

	sessionAudit := &UserSessionAudit{
		UserName:               "test-user",
		ActiveSessions:         2,
		TotalSessions:          150,
		MaxConcurrentSessions:  3,
		SessionAnomalies:       5,
		SuspiciousSessions:     1,
		SessionEfficiency:      0.82,
		SessionRiskScore:       0.3,
	}

	compliance := &UserComplianceStatus{
		UserName:                "test-user",
		OverallComplianceScore:  0.92,
		RequiredCertifications:  []string{"security", "privacy", "data_handling"},
		CompletedCertifications: []string{"security", "privacy"},
		TrainingRequirements:    []string{"security_awareness", "data_protection"},
		CompletedTraining:       []string{"security_awareness"},
		PolicyCompliance:        map[string]float64{"security": 0.95, "privacy": 0.88},
		ComplianceViolations:    2,
		ComplianceRiskLevel:     "low",
		RemediationActions:      []string{"complete_training", "update_certifications"},
	}

	dataAccess := &UserDataAccess{
		UserName:              "test-user",
		DataClassifications:   []string{"public", "internal", "confidential"},
		PIIAccess:             true,
		PHIAccess:             false,
		FinancialDataAccess:   true,
		IntellectualPropertyAccess: false,
		ClassifiedDataAccess:  false,
		DataDownloadCount:     125,
		DataUploadCount:       45,
		DataViewCount:         1250,
		DataModificationCount: 78,
		DataDeletionCount:     5,
		DataSharingCount:      12,
		DataExportCount:       8,
		DataGovernanceScore:   0.87,
		DataComplianceStatus:  "compliant",
		DataBreachInvolvement: []string{},
	}

	resourceAccess := &UserResourceAccess{
		UserName:                "test-user",
		ResourceAccessFrequency: map[string]int64{"compute": 150, "storage": 85, "network": 200},
		QuotaUtilization:        map[string]float64{"compute": 0.75, "storage": 0.60, "network": 0.45},
		ResourceEfficiency:      map[string]float64{"compute": 0.82, "storage": 0.78, "network": 0.91},
		ResourceWaste:           map[string]float64{"compute": 0.18, "storage": 0.22, "network": 0.09},
		ResourceCosts:           map[string]float64{"compute": 1250.50, "storage": 320.75, "network": 85.25},
		ResourceViolations:      map[string]int{"compute": 2, "storage": 0, "network": 1},
		ResourceCompliance:      map[string]float64{"compute": 0.95, "storage": 1.0, "network": 0.88},
		ResourceSecurity:        map[string]float64{"compute": 0.92, "storage": 0.95, "network": 0.87},
		ResourcePerformance:     map[string]float64{"compute": 0.85, "storage": 0.78, "network": 0.93},
		ResourceOptimization:    map[string]float64{"compute": 0.72, "storage": 0.68, "network": 0.89},
	}

	mockClient.On("GetUserPermissions", mock.Anything, "user1").Return(permissions, nil)
	mockClient.On("GetUserPermissions", mock.Anything, "user2").Return(permissions, nil)
	mockClient.On("GetUserPermissions", mock.Anything, "user3").Return(permissions, nil)

	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, "user1").Return(accessPatterns, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, "user2").Return(accessPatterns, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, "user3").Return(accessPatterns, nil)

	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)

	mockClient.On("GetUserSessionAudit", mock.Anything, "user1").Return(sessionAudit, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, "user2").Return(sessionAudit, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, "user3").Return(sessionAudit, nil)

	mockClient.On("GetUserComplianceStatus", mock.Anything, "user1").Return(compliance, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, "user2").Return(compliance, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, "user3").Return(compliance, nil)

	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)

	mockClient.On("GetUserDataAccess", mock.Anything, "user1").Return(dataAccess, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, "user2").Return(dataAccess, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, "user3").Return(dataAccess, nil)

	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)

	mockClient.On("GetUserResourceAccess", mock.Anything, "user1").Return(resourceAccess, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, "user2").Return(resourceAccess, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, "user3").Return(resourceAccess, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_AuditLogMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	auditEntries := []*UserAccessAuditEntry{
		{
			EntryID:            "entry-001",
			UserName:           "test-user",
			AccessType:         "file_access",
			ResourceAccessed:   "/home/user/data.txt",
			Action:             "read",
			Result:             "success",
			Timestamp:          time.Now(),
			ResponseCode:       200,
			ProcessingTime:     500 * time.Millisecond,
			RiskScore:          0.3,
			AnomalyScore:       0.1,
			DataSensitivity:    "confidential",
			ComplianceFlags:    []string{"gdpr"},
		},
		{
			EntryID:            "entry-002",
			UserName:           "test-user",
			AccessType:         "system_access",
			ResourceAccessed:   "/admin/config",
			Action:             "write",
			Result:             "denied",
			Timestamp:          time.Now().Add(-time.Hour),
			ResponseCode:       403,
			ProcessingTime:     200 * time.Millisecond,
			RiskScore:          0.8,
			AnomalyScore:       0.7,
			DataSensitivity:    "restricted",
			ComplianceFlags:    []string{"security_violation"},
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, "user1", mock.Anything, mock.Anything).Return(auditEntries, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect audit metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_ViolationMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	violations := []*UserAccessViolation{
		{
			ViolationID:        "violation-001",
			UserName:           "test-user",
			ViolationType:      "unauthorized_access",
			Severity:           "high",
			Description:        "Attempted access to restricted resource",
			DetectedTime:       time.Now().Add(-2 * time.Hour),
			ResourceInvolved:   "/admin/secrets",
			PolicyViolated:     "access_control_policy",
			RiskScore:          0.85,
			RecurrenceCount:    3,
			ResolutionTime:     func() *time.Time { t := time.Now().Add(-time.Hour); return &t }(),
			ResolutionStatus:   "resolved",
			ImpactAssessment:   "medium",
		},
		{
			ViolationID:        "violation-002",
			UserName:           "test-user",
			ViolationType:      "policy_violation",
			Severity:           "medium",
			Description:        "Data handling policy violation",
			DetectedTime:       time.Now().Add(-24 * time.Hour),
			ResourceInvolved:   "/data/customer_records",
			PolicyViolated:     "data_protection_policy",
			RiskScore:          0.65,
			RecurrenceCount:    1,
			ResolutionStatus:   "investigating",
			ImpactAssessment:   "low",
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, "user1").Return(violations, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect violation metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_SecurityEventMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	securityEvents := []*UserSecurityEvent{
		{
			EventID:         "event-001",
			UserName:        "test-user",
			EventType:       "suspicious_login",
			Severity:        "high",
			EventTime:       time.Now().Add(-time.Hour),
			EventDescription: "Login from unusual location",
			ThreatLevel:     "medium",
			AttackVector:    "geographic_anomaly",
			ResponseTime:    15 * time.Minute,
			ResolutionTime:  2 * time.Hour,
			BusinessImpact:  "low",
		},
		{
			EventID:         "event-002",
			UserName:        "test-user",
			EventType:       "privilege_escalation",
			Severity:        "critical",
			EventTime:       time.Now().Add(-3 * time.Hour),
			EventDescription: "Unauthorized privilege escalation attempt",
			ThreatLevel:     "high",
			AttackVector:    "privilege_abuse",
			ResponseTime:    5 * time.Minute,
			ResolutionTime:  30 * time.Minute,
			BusinessImpact:  "medium",
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, "user1").Return(securityEvents, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect security event metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_RoleAssignmentMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	roleAssignments := []*UserRoleAssignment{
		{
			AssignmentID:     "role-001",
			UserName:         "test-user",
			RoleName:         "data_analyst",
			RoleType:         "functional",
			AssignmentType:   "permanent",
			AssignmentStatus: "active",
			RolePermissions:  []string{"read_data", "analyze_data", "generate_reports"},
			ConflictingRoles: []string{},
			UsageMetrics:     map[string]float64{"overall": 0.85},
			RiskAssessment:   "low",
			ComplianceImpact: "compliant",
		},
		{
			AssignmentID:     "role-002",
			UserName:         "test-user",
			RoleName:         "backup_admin",
			RoleType:         "administrative",
			AssignmentType:   "temporary",
			AssignmentStatus: "active",
			RolePermissions:  []string{"backup_data", "restore_data", "admin_backup_system"},
			ConflictingRoles: []string{"data_analyst"},
			UsageMetrics:     map[string]float64{"overall": 0.45},
			RiskAssessment:   "medium",
			ComplianceImpact: "warning",
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, "user1").Return(roleAssignments, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect role assignment metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_PrivilegeEscalationMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	escalations := []*UserPrivilegeEscalation{
		{
			EscalationID:       "esc-001",
			UserName:           "test-user",
			EscalationType:     "temporary_admin",
			RequestTime:        time.Now().Add(-4 * time.Hour),
			OriginalPrivileges: []string{"read", "write"},
			RequestedPrivileges: []string{"admin", "delete"},
			GrantedPrivileges:  []string{"admin"},
			ActualDuration:     2 * time.Hour,
			AutoRevocation:     true,
			RiskAssessment:     "medium",
			EscalationPattern:  "normal",
		},
		{
			EscalationID:       "esc-002",
			UserName:           "test-user",
			EscalationType:     "emergency_access",
			RequestTime:        time.Now().Add(-24 * time.Hour),
			OriginalPrivileges: []string{"read"},
			RequestedPrivileges: []string{"emergency_admin"},
			GrantedPrivileges:  []string{"emergency_admin"},
			ActualDuration:     30 * time.Minute,
			AutoRevocation:     true,
			RiskAssessment:     "high",
			EscalationPattern:  "suspicious",
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, "user1").Return(escalations, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return([]*UserAuthenticationEvent{}, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect privilege escalation metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_AuthenticationEventMetrics(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	authEvents := []*UserAuthenticationEvent{
		{
			EventID:              "auth-001",
			UserName:             "test-user",
			AuthenticationType:   "password",
			AuthenticationMethod: "local",
			AuthenticationResult: "success",
			EventTime:            time.Now(),
			RiskScore:            0.2,
			TrustScore:           0.9,
			AnomalyScore:         0.1,
			MultiFactorUsed:      true,
			AuthenticationDuration: 2 * time.Second,
			TokenLifetime:        24 * time.Hour,
		},
		{
			EventID:              "auth-002",
			UserName:             "test-user",
			AuthenticationType:   "sso",
			AuthenticationMethod: "saml",
			AuthenticationResult: "success",
			EventTime:            time.Now().Add(-time.Hour),
			RiskScore:            0.1,
			TrustScore:           0.95,
			AnomalyScore:         0.05,
			MultiFactorUsed:      true,
			AuthenticationDuration: 1 * time.Second,
			TokenLifetime:        8 * time.Hour,
		},
		{
			EventID:              "auth-003",
			UserName:             "test-user",
			AuthenticationType:   "password",
			AuthenticationMethod: "local",
			AuthenticationResult: "failure",
			EventTime:            time.Now().Add(-2 * time.Hour),
			RiskScore:            0.8,
			TrustScore:           0.3,
			AnomalyScore:         0.7,
			MultiFactorUsed:      false,
			AuthenticationDuration: 5 * time.Second,
			TokenLifetime:        0,
		},
	}

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return(&UserPermissions{}, nil)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserAccessAuditEntry{}, nil)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return(&UserAccessPatterns{}, nil)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*UserPermissionHistoryEntry{}, nil)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return([]*UserAccessViolation{}, nil)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return(&UserSessionAudit{}, nil)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return(&UserComplianceStatus{}, nil)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return([]*UserSecurityEvent{}, nil)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return([]*UserRoleAssignment{}, nil)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return([]*UserPrivilegeEscalation{}, nil)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return(&UserDataAccess{}, nil)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return([]*UserAccountAssociation{}, nil)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, "user1").Return(authEvents, nil)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return(&UserResourceAccess{}, nil)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Greater(t, metricCount, 0, "Should collect authentication event metrics")

	mockClient.AssertExpectations(t)
}

func TestUserPermissionAuditCollector_ErrorHandling(t *testing.T) {
	mockClient := new(MockUserPermissionAuditSLURMClient)
	collector := NewUserPermissionAuditCollector(mockClient)

	mockClient.On("GetUserPermissions", mock.Anything, mock.Anything).Return((*UserPermissions)(nil), assert.AnError)
	mockClient.On("GetUserAccessAuditLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(([]*UserAccessAuditEntry)(nil), assert.AnError)
	mockClient.On("GetUserAccessPatterns", mock.Anything, mock.Anything).Return((*UserAccessPatterns)(nil), assert.AnError)
	mockClient.On("GetUserPermissionHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(([]*UserPermissionHistoryEntry)(nil), assert.AnError)
	mockClient.On("GetUserAccessViolations", mock.Anything, mock.Anything).Return(([]*UserAccessViolation)(nil), assert.AnError)
	mockClient.On("GetUserSessionAudit", mock.Anything, mock.Anything).Return((*UserSessionAudit)(nil), assert.AnError)
	mockClient.On("GetUserComplianceStatus", mock.Anything, mock.Anything).Return((*UserComplianceStatus)(nil), assert.AnError)
	mockClient.On("GetUserSecurityEvents", mock.Anything, mock.Anything).Return(([]*UserSecurityEvent)(nil), assert.AnError)
	mockClient.On("GetUserRoleAssignments", mock.Anything, mock.Anything).Return(([]*UserRoleAssignment)(nil), assert.AnError)
	mockClient.On("GetUserPrivilegeEscalations", mock.Anything, mock.Anything).Return(([]*UserPrivilegeEscalation)(nil), assert.AnError)
	mockClient.On("GetUserDataAccess", mock.Anything, mock.Anything).Return((*UserDataAccess)(nil), assert.AnError)
	mockClient.On("GetUserAccountAssociations", mock.Anything, mock.Anything).Return(([]*UserAccountAssociation)(nil), assert.AnError)
	mockClient.On("GetUserAuthenticationEvents", mock.Anything, mock.Anything).Return(([]*UserAuthenticationEvent)(nil), assert.AnError)
	mockClient.On("GetUserResourceAccess", mock.Anything, mock.Anything).Return((*UserResourceAccess)(nil), assert.AnError)

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	metricCount := testutil.CollectAndCount(collector)
	assert.Equal(t, 0, metricCount, "Should not collect metrics on error")

	mockClient.AssertExpectations(t)
}