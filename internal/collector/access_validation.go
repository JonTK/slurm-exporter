// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// AccessValidationSLURMClient defines the interface for SLURM client operations related to access validation
type AccessValidationSLURMClient interface {
	ValidateUserAccountAccess(ctx context.Context, userName, accountName string) (*UserAccessValidation, error)
	GetUserAccessRights(ctx context.Context, userName string) (*UserAccessRights, error)
	GetAccountAccessPolicies(ctx context.Context, accountName string) (*AccountAccessPolicies, error)
	GetAccessAuditLog(ctx context.Context, period string) (*AccessAuditLog, error)
	GetAccessViolations(ctx context.Context, severity string) (*AccessViolations, error)
	GetAccessPatterns(ctx context.Context, userName string) (*AccessPatterns, error)
	GetSessionValidation(ctx context.Context, sessionID string) (*SessionValidation, error)
	GetComplianceStatus(ctx context.Context, scope string) (*ComplianceStatus, error)
	GetSecurityAlerts(ctx context.Context, period string) (*SecurityAlerts, error)
	GetAccessRecommendations(ctx context.Context, userName string) (*AccessRecommendations, error)
	GetRiskAssessment(ctx context.Context, entityType, entityID string) (*RiskAssessment, error)
	GetAccessTrends(ctx context.Context, period string) (*AccessTrends, error)
}

// UserAccessValidation represents the result of user access validation
type UserAccessValidation struct {
	UserName            string
	AccountName         string
	AccessGranted       bool
	ValidationTime      time.Time
	ValidationMethod    string
	AuthenticationLevel string
	PermissionLevel     string
	Restrictions        []string
	ValidationReasons   []string
	TTL                 time.Duration
	SessionID           string
	SourceIP            string
	UserAgent           string
	RiskScore           float64
	ComplianceStatus    string
	RequiredActions     []string
}

// UserAccessRights represents comprehensive user access rights
type UserAccessRights struct {
	UserName            string
	PrimaryAccount      string
	SecondaryAccounts   []string
	GlobalPermissions   []string
	AccountPermissions  map[string][]string
	ResourcePermissions map[string][]string
	EffectiveRights     map[string]bool
	Restrictions        map[string][]string
	ExpirationDates     map[string]time.Time
	LastReview          time.Time
	NextReview          time.Time
	ReviewedBy          string
	ComplianceFlags     []string
}

// AccountAccessPolicies represents access policies for an account
type AccountAccessPolicies struct {
	AccountName           string
	PolicyVersion         string
	RequiredAuth          []string
	MinAuthLevel          int
	AllowedUsers          []string
	DeniedUsers           []string
	AllowedGroups         []string
	DeniedGroups          []string
	TimeRestrictions      []TimeRestriction
	LocationRestrictions  []LocationRestriction
	ResourceLimits        map[string]int
	MFARequired           bool
	SessionTimeout        time.Duration
	MaxConcurrentSessions int
	PolicyEnforcement     string
	AuditLevel            string
	LastModified          time.Time
	ModifiedBy            string
}

// TimeRestriction represents time-based access restrictions
type TimeRestriction struct {
	StartTime   string
	EndTime     string
	DaysOfWeek  []string
	Timezone    string
	Description string
}

// LocationRestriction represents location-based access restrictions
type LocationRestriction struct {
	AllowedIPs       []string
	DeniedIPs        []string
	AllowedCountries []string
	DeniedCountries  []string
	GeoFencing       bool
}

// AccessAuditLog represents access audit trail
type AccessAuditLog struct {
	Period           string
	TotalEvents      int
	SuccessfulLogins int
	FailedLogins     int
	AccessChanges    int
	PolicyViolations int
	SecurityEvents   int
	Events           []AuditEvent
	Summary          AuditSummary
}

// AuditEvent represents a single audit event
type AuditEvent struct {
	EventID     string
	Timestamp   time.Time
	UserName    string
	AccountName string
	EventType   string
	Action      string
	Result      string
	SourceIP    string
	UserAgent   string
	Details     map[string]interface{}
	RiskScore   float64
	Flagged     bool
}

// AuditSummary represents audit log summary
type AuditSummary struct {
	TopUsers        map[string]int
	TopAccounts     map[string]int
	TopViolations   map[string]int
	RiskTrends      []float64
	ComplianceScore float64
}

// AccessViolations represents detected access violations
type AccessViolations struct {
	TotalViolations     int
	CriticalViolations  int
	WarningViolations   int
	InfoViolations      int
	UnresolvedCount     int
	Violations          []AccessViolation
	ViolationsByType    map[string]int
	ViolationsByUser    map[string]int
	ViolationsByAccount map[string]int
	TrendData           []int
}

// AccessViolation represents a single access violation
type AccessViolation struct {
	ViolationID      string
	Timestamp        time.Time
	UserName         string
	AccountName      string
	ViolationType    string
	Severity         string
	Description      string
	AttemptedAction  string
	DeniedReason     string
	SourceIP         string
	RiskScore        float64
	AutoBlocked      bool
	ResolutionStatus string
	ResolutionTime   time.Time
	ResolvedBy       string
}

// AccessPatterns represents user access patterns
type AccessPatterns struct {
	UserName                string
	LoginTimes              []time.Time
	AccessFrequency         map[string]int
	PreferredAccounts       []string
	CommonOperations        []string
	AverageSessionLength    time.Duration
	PeakUsageHours          []int
	AccessLocations         []string
	DeviceFingerprints      []string
	BehaviorScore           float64
	AnomalyDetected         bool
	RiskProfile             string
	RecommendedRestrictions []string
}

// SessionValidation represents session validation results
type SessionValidation struct {
	SessionID         string
	UserName          string
	AccountName       string
	Valid             bool
	CreatedAt         time.Time
	LastActivity      time.Time
	ExpiresAt         time.Time
	SourceIP          string
	UserAgent         string
	AuthMethod        string
	MFACompleted      bool
	PermissionsCached []string
	ActivityCount     int
	RiskScore         float64
	Flags             []string
}

// ComplianceStatus represents compliance monitoring
type ComplianceStatus struct {
	Scope             string
	ComplianceScore   float64
	PolicyAdherence   float64
	SecurityPosture   float64
	AuditReadiness    float64
	ControlsInPlace   int
	ControlsFailing   int
	RequiredActions   []ComplianceAction
	Certifications    []Certification
	LastAudit         time.Time
	NextAudit         time.Time
	ComplianceOfficer string
	RiskLevel         string
}

// ComplianceAction represents required compliance action
type ComplianceAction struct {
	ActionID    string
	Description string
	Priority    string
	Deadline    time.Time
	AssignedTo  string
	Status      string
	Impact      string
}

// Certification represents compliance certification
type Certification struct {
	Name         string
	Status       string
	ValidUntil   time.Time
	Issuer       string
	Requirements []string
}

// SecurityAlerts represents security alerts
type SecurityAlerts struct {
	Period            string
	TotalAlerts       int
	CriticalAlerts    int
	ActiveAlerts      int
	ResolvedAlerts    int
	MeanTimeToResolve time.Duration
	Alerts            []SecurityAlert
	AlertsByType      map[string]int
	AlertsByUser      map[string]int
	FalsePositiveRate float64
}

// SecurityAlert represents a single security alert
type SecurityAlert struct {
	AlertID           string
	Timestamp         time.Time
	AlertType         string
	Severity          string
	UserName          string
	AccountName       string
	Description       string
	Indicators        []string
	RecommendedAction string
	AutoResponse      string
	Status            string
	AssignedTo        string
	ResolutionTime    time.Time
	FalsePositive     bool
}

// AccessRecommendations represents access optimization recommendations
type AccessRecommendations struct {
	UserName              string
	CurrentAccess         []string
	RecommendedAccess     []string
	RemovalCandidates     []string
	AdditionalNeeded      []string
	RiskReduction         float64
	ComplianceImprovement float64
	Justifications        map[string]string
	ImpactAnalysis        map[string]interface{}
	ImplementationSteps   []string
}

// RiskAssessment represents risk assessment results
type RiskAssessment struct {
	EntityType      string
	EntityID        string
	OverallRisk     float64
	RiskFactors     map[string]float64
	Vulnerabilities []Vulnerability
	Threats         []Threat
	Mitigations     []Mitigation
	TrendDirection  string
	PreviousScore   float64
	AssessmentDate  time.Time
	NextAssessment  time.Time
	Recommendations []string
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string
	Type        string
	Severity    string
	Description string
	Impact      float64
	Likelihood  float64
	Mitigation  string
}

// Threat represents a security threat
type Threat struct {
	ID          string
	Type        string
	Source      string
	Target      string
	Probability float64
	Impact      float64
	Status      string
}

// Mitigation represents a risk mitigation
type Mitigation struct {
	ID             string
	Type           string
	Description    string
	Effectiveness  float64
	Cost           float64
	Implementation string
	Status         string
}

// AccessTrends represents access trend analysis
type AccessTrends struct {
	Period                 string
	LoginTrends            []float64
	FailureTrends          []float64
	ViolationTrends        []float64
	RiskTrends             []float64
	NewUserTrends          []float64
	InactiveUserTrends     []float64
	PermissionChangeTrends []float64
	SessionLengthTrends    []float64
	GeographicTrends       map[string][]float64
	DeviceTrends           map[string][]float64
	Predictions            map[string]float64
}

// AccessValidationCollector collects access validation metrics from SLURM
type AccessValidationCollector struct {
	client AccessValidationSLURMClient
	mutex  sync.RWMutex

	// Access validation metrics
	userAccessValidation *prometheus.GaugeVec
	accessValidationTime *prometheus.HistogramVec
	accessGranted        *prometheus.GaugeVec
	accessDenied         *prometheus.CounterVec
	authenticationLevel  *prometheus.GaugeVec
	permissionLevel      *prometheus.GaugeVec
	riskScore            *prometheus.GaugeVec

	// Access rights metrics
	userAccessRights     *prometheus.GaugeVec
	effectivePermissions *prometheus.GaugeVec
	accountPermissions   *prometheus.GaugeVec
	permissionExpiry     *prometheus.GaugeVec
	complianceFlags      *prometheus.GaugeVec

	// Policy metrics
	policyVersion         *prometheus.GaugeVec
	minAuthLevel          *prometheus.GaugeVec
	mfaRequired           *prometheus.GaugeVec
	sessionTimeout        *prometheus.GaugeVec
	maxConcurrentSessions *prometheus.GaugeVec
	policyViolations      *prometheus.CounterVec

	// Audit metrics
	auditEvents          *prometheus.CounterVec
	successfulLogins     *prometheus.CounterVec
	failedLogins         *prometheus.CounterVec
	securityEvents       *prometheus.CounterVec
	auditComplianceScore *prometheus.GaugeVec

	// Violation metrics
	accessViolations     *prometheus.CounterVec
	violationSeverity    *prometheus.GaugeVec
	unresolvedViolations *prometheus.GaugeVec
	autoBlockedAttempts  *prometheus.CounterVec
	violationRiskScore   *prometheus.GaugeVec

	// Pattern metrics
	accessFrequency *prometheus.GaugeVec
	sessionLength   *prometheus.HistogramVec
	behaviorScore   *prometheus.GaugeVec
	anomalyDetected *prometheus.GaugeVec
	peakUsageHours  *prometheus.GaugeVec

	// Session metrics
	activeSessions    *prometheus.GaugeVec
	sessionValidity   *prometheus.GaugeVec
	sessionActivity   *prometheus.GaugeVec
	sessionRiskScore  *prometheus.GaugeVec
	mfaCompletionRate *prometheus.GaugeVec

	// Compliance metrics
	complianceScore *prometheus.GaugeVec
	policyAdherence *prometheus.GaugeVec
	securityPosture *prometheus.GaugeVec
	controlsInPlace *prometheus.GaugeVec
	controlsFailing *prometheus.GaugeVec
	requiredActions *prometheus.GaugeVec

	// Security alert metrics
	securityAlerts    *prometheus.CounterVec
	alertSeverity     *prometheus.GaugeVec
	activeAlerts      *prometheus.GaugeVec
	meanTimeToResolve *prometheus.HistogramVec
	falsePositiveRate *prometheus.GaugeVec

	// Risk assessment metrics
	overallRisk     *prometheus.GaugeVec
	riskFactors     *prometheus.GaugeVec
	vulnerabilities *prometheus.GaugeVec
	threats         *prometheus.GaugeVec
	mitigations     *prometheus.GaugeVec

	// Trend metrics
	loginTrend     *prometheus.GaugeVec
	failureTrend   *prometheus.GaugeVec
	violationTrend *prometheus.GaugeVec
	riskTrend      *prometheus.GaugeVec

	// Collection metrics
	collectionDuration *prometheus.HistogramVec
	collectionErrors   *prometheus.CounterVec
	lastCollectionTime *prometheus.GaugeVec
}

// NewAccessValidationCollector creates a new access validation collector
func NewAccessValidationCollector(client AccessValidationSLURMClient) *AccessValidationCollector {
	return &AccessValidationCollector{
		client: client,

		// Access validation metrics
		userAccessValidation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_access_validation_status",
				Help: "User access validation status (1 = granted, 0 = denied)",
			},
			[]string{"user", "account", "validation_method"},
		),
		accessValidationTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_access_validation_duration_seconds",
				Help:    "Time taken to validate access",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"validation_method"},
		),
		accessGranted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_granted_total",
				Help: "Total number of access grants",
			},
			[]string{"account", "permission_level"},
		),
		accessDenied: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_access_denied_total",
				Help: "Total number of access denials",
			},
			[]string{"account", "reason"},
		),
		authenticationLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_authentication_level",
				Help: "Current authentication level for user",
			},
			[]string{"user", "method"},
		),
		permissionLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_permission_level",
				Help: "Permission level for user on account",
			},
			[]string{"user", "account"},
		),
		riskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_risk_score",
				Help: "Risk score for access attempt",
			},
			[]string{"user", "account"},
		),

		// Access rights metrics
		userAccessRights: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_access_rights_count",
				Help: "Number of access rights per user",
			},
			[]string{"user", "right_type"},
		),
		effectivePermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_effective_permissions_count",
				Help: "Number of effective permissions",
			},
			[]string{"user", "account"},
		),
		accountPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_permissions_count",
				Help: "Number of permissions per account",
			},
			[]string{"user", "account", "permission_type"},
		),
		permissionExpiry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_permission_expiry_days",
				Help: "Days until permission expires",
			},
			[]string{"user", "account", "permission"},
		),
		complianceFlags: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_compliance_flags_count",
				Help: "Number of compliance flags",
			},
			[]string{"user"},
		),

		// Policy metrics
		policyVersion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_policy_version",
				Help: "Current policy version number",
			},
			[]string{"account"},
		),
		minAuthLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_min_auth_level",
				Help: "Minimum authentication level required",
			},
			[]string{"account"},
		),
		mfaRequired: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_mfa_required",
				Help: "Whether MFA is required (1 = yes, 0 = no)",
			},
			[]string{"account"},
		),
		sessionTimeout: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_session_timeout_seconds",
				Help: "Session timeout duration in seconds",
			},
			[]string{"account"},
		),
		maxConcurrentSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_max_concurrent_sessions",
				Help: "Maximum allowed concurrent sessions",
			},
			[]string{"account"},
		),
		policyViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_policy_violations_total",
				Help: "Total policy violations",
			},
			[]string{"account", "policy_type"},
		),

		// Audit metrics
		auditEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_audit_events_total",
				Help: "Total audit events",
			},
			[]string{"event_type", "result"},
		),
		successfulLogins: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_successful_logins_total",
				Help: "Total successful login attempts",
			},
			[]string{"account"},
		),
		failedLogins: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_failed_logins_total",
				Help: "Total failed login attempts",
			},
			[]string{"account", "reason"},
		),
		securityEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_security_events_total",
				Help: "Total security events",
			},
			[]string{"event_type", "severity"},
		),
		auditComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_audit_compliance_score",
				Help: "Audit compliance score (0-1)",
			},
			[]string{},
		),

		// Violation metrics
		accessViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_access_violations_total",
				Help: "Total access violations",
			},
			[]string{"violation_type", "severity"},
		),
		violationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_severity_distribution",
				Help: "Distribution of violations by severity",
			},
			[]string{"severity"},
		),
		unresolvedViolations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_unresolved_violations_count",
				Help: "Number of unresolved violations",
			},
			[]string{"severity"},
		),
		autoBlockedAttempts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_auto_blocked_attempts_total",
				Help: "Total auto-blocked access attempts",
			},
			[]string{"reason"},
		),
		violationRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_risk_score",
				Help: "Risk score of violations",
			},
			[]string{"violation_type"},
		),

		// Pattern metrics
		accessFrequency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_frequency",
				Help: "Access frequency per user",
			},
			[]string{"user", "account", "operation"},
		),
		sessionLength: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_session_length_seconds",
				Help:    "Session length distribution",
				Buckets: prometheus.ExponentialBuckets(60, 2, 10), // 1min to ~17hrs
			},
			[]string{"user"},
		),
		behaviorScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_behavior_score",
				Help: "User behavior score (0-1)",
			},
			[]string{"user"},
		),
		anomalyDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_anomaly_detected",
				Help: "Whether anomaly is detected (1 = yes, 0 = no)",
			},
			[]string{"user", "anomaly_type"},
		),
		peakUsageHours: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_peak_usage_hour",
				Help: "Peak usage hour of day",
			},
			[]string{"user"},
		),

		// Session metrics
		activeSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_active_sessions_count",
				Help: "Number of active sessions",
			},
			[]string{"account"},
		),
		sessionValidity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_session_validity",
				Help: "Session validity status (1 = valid, 0 = invalid)",
			},
			[]string{"session_id", "user"},
		),
		sessionActivity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_session_activity_count",
				Help: "Number of activities in session",
			},
			[]string{"session_id"},
		),
		sessionRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_session_risk_score",
				Help: "Risk score for active session",
			},
			[]string{"session_id", "user"},
		),
		mfaCompletionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_mfa_completion_rate",
				Help: "MFA completion rate",
			},
			[]string{"account"},
		),

		// Compliance metrics
		complianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_compliance_score",
				Help: "Overall compliance score (0-1)",
			},
			[]string{"scope"},
		),
		policyAdherence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_policy_adherence_score",
				Help: "Policy adherence score (0-1)",
			},
			[]string{"scope"},
		),
		securityPosture: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_security_posture_score",
				Help: "Security posture score (0-1)",
			},
			[]string{"scope"},
		),
		controlsInPlace: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_controls_in_place_count",
				Help: "Number of security controls in place",
			},
			[]string{"scope"},
		),
		controlsFailing: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_controls_failing_count",
				Help: "Number of failing security controls",
			},
			[]string{"scope"},
		),
		requiredActions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_required_compliance_actions_count",
				Help: "Number of required compliance actions",
			},
			[]string{"priority"},
		),

		// Security alert metrics
		securityAlerts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_security_alerts_total",
				Help: "Total security alerts",
			},
			[]string{"alert_type", "severity"},
		),
		alertSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_alert_severity_distribution",
				Help: "Distribution of alerts by severity",
			},
			[]string{"severity"},
		),
		activeAlerts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_active_alerts_count",
				Help: "Number of active security alerts",
			},
			[]string{"severity"},
		),
		meanTimeToResolve: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_alert_mttr_seconds",
				Help:    "Mean time to resolve alerts",
				Buckets: prometheus.ExponentialBuckets(60, 2, 10),
			},
			[]string{"alert_type"},
		),
		falsePositiveRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_alert_false_positive_rate",
				Help: "False positive rate for alerts",
			},
			[]string{"alert_type"},
		),

		// Risk assessment metrics
		overallRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_overall_risk_score",
				Help: "Overall risk score (0-1)",
			},
			[]string{"entity_type", "entity_id"},
		),
		riskFactors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_risk_factor_score",
				Help: "Individual risk factor scores",
			},
			[]string{"entity_type", "entity_id", "factor"},
		),
		vulnerabilities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_vulnerabilities_count",
				Help: "Number of identified vulnerabilities",
			},
			[]string{"severity"},
		),
		threats: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_threats_count",
				Help: "Number of identified threats",
			},
			[]string{"threat_type", "status"},
		),
		mitigations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_mitigations_count",
				Help: "Number of mitigations in place",
			},
			[]string{"mitigation_type", "status"},
		),

		// Trend metrics
		loginTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_login_trend",
				Help: "Login trend value",
			},
			[]string{"period"},
		),
		failureTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_failure_trend",
				Help: "Failure trend value",
			},
			[]string{"period"},
		),
		violationTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_trend",
				Help: "Violation trend value",
			},
			[]string{"period"},
		),
		riskTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_risk_trend",
				Help: "Risk trend value",
			},
			[]string{"period"},
		),

		// Collection metrics
		collectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_access_validation_collection_duration_seconds",
				Help:    "Time spent collecting access validation metrics",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		collectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_access_validation_collection_errors_total",
				Help: "Total number of errors during access validation collection",
			},
			[]string{"operation", "error_type"},
		),
		lastCollectionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_access_validation_last_collection_timestamp",
				Help: "Timestamp of last successful collection",
			},
			[]string{"metric_type"},
		),
	}
}

// Describe sends the super-set of all possible descriptors
func (c *AccessValidationCollector) Describe(ch chan<- *prometheus.Desc) {
	c.userAccessValidation.Describe(ch)
	c.accessValidationTime.Describe(ch)
	c.accessGranted.Describe(ch)
	c.accessDenied.Describe(ch)
	c.authenticationLevel.Describe(ch)
	c.permissionLevel.Describe(ch)
	c.riskScore.Describe(ch)
	c.userAccessRights.Describe(ch)
	c.effectivePermissions.Describe(ch)
	c.accountPermissions.Describe(ch)
	c.permissionExpiry.Describe(ch)
	c.complianceFlags.Describe(ch)
	c.policyVersion.Describe(ch)
	c.minAuthLevel.Describe(ch)
	c.mfaRequired.Describe(ch)
	c.sessionTimeout.Describe(ch)
	c.maxConcurrentSessions.Describe(ch)
	c.policyViolations.Describe(ch)
	c.auditEvents.Describe(ch)
	c.successfulLogins.Describe(ch)
	c.failedLogins.Describe(ch)
	c.securityEvents.Describe(ch)
	c.auditComplianceScore.Describe(ch)
	c.accessViolations.Describe(ch)
	c.violationSeverity.Describe(ch)
	c.unresolvedViolations.Describe(ch)
	c.autoBlockedAttempts.Describe(ch)
	c.violationRiskScore.Describe(ch)
	c.accessFrequency.Describe(ch)
	c.sessionLength.Describe(ch)
	c.behaviorScore.Describe(ch)
	c.anomalyDetected.Describe(ch)
	c.peakUsageHours.Describe(ch)
	c.activeSessions.Describe(ch)
	c.sessionValidity.Describe(ch)
	c.sessionActivity.Describe(ch)
	c.sessionRiskScore.Describe(ch)
	c.mfaCompletionRate.Describe(ch)
	c.complianceScore.Describe(ch)
	c.policyAdherence.Describe(ch)
	c.securityPosture.Describe(ch)
	c.controlsInPlace.Describe(ch)
	c.controlsFailing.Describe(ch)
	c.requiredActions.Describe(ch)
	c.securityAlerts.Describe(ch)
	c.alertSeverity.Describe(ch)
	c.activeAlerts.Describe(ch)
	c.meanTimeToResolve.Describe(ch)
	c.falsePositiveRate.Describe(ch)
	c.overallRisk.Describe(ch)
	c.riskFactors.Describe(ch)
	c.vulnerabilities.Describe(ch)
	c.threats.Describe(ch)
	c.mitigations.Describe(ch)
	c.loginTrend.Describe(ch)
	c.failureTrend.Describe(ch)
	c.violationTrend.Describe(ch)
	c.riskTrend.Describe(ch)
	c.collectionDuration.Describe(ch)
	c.collectionErrors.Describe(ch)
	c.lastCollectionTime.Describe(ch)
}

// Collect fetches the stats and delivers them as Prometheus metrics
func (c *AccessValidationCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset metrics
	c.userAccessValidation.Reset()
	c.accessGranted.Reset()
	c.authenticationLevel.Reset()
	c.permissionLevel.Reset()
	c.riskScore.Reset()
	c.userAccessRights.Reset()
	c.effectivePermissions.Reset()
	c.accountPermissions.Reset()
	c.permissionExpiry.Reset()
	c.complianceFlags.Reset()
	c.policyVersion.Reset()
	c.minAuthLevel.Reset()
	c.mfaRequired.Reset()
	c.sessionTimeout.Reset()
	c.maxConcurrentSessions.Reset()
	c.violationSeverity.Reset()
	c.unresolvedViolations.Reset()
	c.violationRiskScore.Reset()
	c.accessFrequency.Reset()
	c.behaviorScore.Reset()
	c.anomalyDetected.Reset()
	c.peakUsageHours.Reset()
	c.activeSessions.Reset()
	c.sessionValidity.Reset()
	c.sessionActivity.Reset()
	c.sessionRiskScore.Reset()
	c.mfaCompletionRate.Reset()
	c.complianceScore.Reset()
	c.policyAdherence.Reset()
	c.securityPosture.Reset()
	c.controlsInPlace.Reset()
	c.controlsFailing.Reset()
	c.requiredActions.Reset()
	c.alertSeverity.Reset()
	c.activeAlerts.Reset()
	c.falsePositiveRate.Reset()
	c.overallRisk.Reset()
	c.riskFactors.Reset()
	c.vulnerabilities.Reset()
	c.threats.Reset()
	c.mitigations.Reset()
	c.loginTrend.Reset()
	c.failureTrend.Reset()
	c.violationTrend.Reset()
	c.riskTrend.Reset()

	// Collect various metrics
	c.collectAccessValidationMetrics()
	c.collectAccessRightsMetrics()
	c.collectPolicyMetrics()
	c.collectAuditMetrics()
	c.collectViolationMetrics()
	c.collectPatternMetrics()
	c.collectSessionMetrics()
	c.collectComplianceMetrics()
	c.collectSecurityAlertMetrics()
	c.collectRiskAssessmentMetrics()
	c.collectTrendMetrics()

	// Collect all metrics
	c.userAccessValidation.Collect(ch)
	c.accessValidationTime.Collect(ch)
	c.accessGranted.Collect(ch)
	c.accessDenied.Collect(ch)
	c.authenticationLevel.Collect(ch)
	c.permissionLevel.Collect(ch)
	c.riskScore.Collect(ch)
	c.userAccessRights.Collect(ch)
	c.effectivePermissions.Collect(ch)
	c.accountPermissions.Collect(ch)
	c.permissionExpiry.Collect(ch)
	c.complianceFlags.Collect(ch)
	c.policyVersion.Collect(ch)
	c.minAuthLevel.Collect(ch)
	c.mfaRequired.Collect(ch)
	c.sessionTimeout.Collect(ch)
	c.maxConcurrentSessions.Collect(ch)
	c.policyViolations.Collect(ch)
	c.auditEvents.Collect(ch)
	c.successfulLogins.Collect(ch)
	c.failedLogins.Collect(ch)
	c.securityEvents.Collect(ch)
	c.auditComplianceScore.Collect(ch)
	c.accessViolations.Collect(ch)
	c.violationSeverity.Collect(ch)
	c.unresolvedViolations.Collect(ch)
	c.autoBlockedAttempts.Collect(ch)
	c.violationRiskScore.Collect(ch)
	c.accessFrequency.Collect(ch)
	c.sessionLength.Collect(ch)
	c.behaviorScore.Collect(ch)
	c.anomalyDetected.Collect(ch)
	c.peakUsageHours.Collect(ch)
	c.activeSessions.Collect(ch)
	c.sessionValidity.Collect(ch)
	c.sessionActivity.Collect(ch)
	c.sessionRiskScore.Collect(ch)
	c.mfaCompletionRate.Collect(ch)
	c.complianceScore.Collect(ch)
	c.policyAdherence.Collect(ch)
	c.securityPosture.Collect(ch)
	c.controlsInPlace.Collect(ch)
	c.controlsFailing.Collect(ch)
	c.requiredActions.Collect(ch)
	c.securityAlerts.Collect(ch)
	c.alertSeverity.Collect(ch)
	c.activeAlerts.Collect(ch)
	c.meanTimeToResolve.Collect(ch)
	c.falsePositiveRate.Collect(ch)
	c.overallRisk.Collect(ch)
	c.riskFactors.Collect(ch)
	c.vulnerabilities.Collect(ch)
	c.threats.Collect(ch)
	c.mitigations.Collect(ch)
	c.loginTrend.Collect(ch)
	c.failureTrend.Collect(ch)
	c.violationTrend.Collect(ch)
	c.riskTrend.Collect(ch)
	c.collectionDuration.Collect(ch)
	c.collectionErrors.Collect(ch)
	c.lastCollectionTime.Collect(ch)
}

func (c *AccessValidationCollector) collectAccessValidationMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample user-account pairs for validation
	testPairs := []struct {
		user    string
		account string
	}{
		{"user1", "research"},
		{"user2", "engineering"},
		{"admin1", "admin"},
		{"guest1", "public"},
	}

	for _, pair := range testPairs {
		validationStart := time.Now()
		validation, err := c.client.ValidateUserAccountAccess(ctx, pair.user, pair.account)
		validationDuration := time.Since(validationStart).Seconds()

		if err != nil {
			c.collectionErrors.WithLabelValues("validation", "access_validation_error").Inc()
			continue
		}

		if validation != nil {
			// Set validation status
			status := 0.0
			if validation.AccessGranted {
				status = 1.0
				c.accessGranted.WithLabelValues(pair.account, validation.PermissionLevel).Inc()
			} else {
				for _, reason := range validation.ValidationReasons {
					c.accessDenied.WithLabelValues(pair.account, reason).Inc()
				}
			}
			c.userAccessValidation.WithLabelValues(pair.user, pair.account, validation.ValidationMethod).Set(status)

			// Record validation time
			c.accessValidationTime.WithLabelValues(validation.ValidationMethod).Observe(validationDuration)

			// Set authentication level
			authLevel := getAuthLevelNumeric(validation.AuthenticationLevel)
			c.authenticationLevel.WithLabelValues(pair.user, validation.ValidationMethod).Set(authLevel)

			// Set permission level
			permLevel := getPermissionLevelNumeric(validation.PermissionLevel)
			c.permissionLevel.WithLabelValues(pair.user, pair.account).Set(permLevel)

			// Set risk score
			c.riskScore.WithLabelValues(pair.user, pair.account).Set(validation.RiskScore)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("access_validation").Observe(duration)
	c.lastCollectionTime.WithLabelValues("access_validation").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectAccessRightsMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample users for access rights
	sampleUsers := []string{"user1", "user2", "admin1", "operator1"}

	for _, userName := range sampleUsers {
		rights, err := c.client.GetUserAccessRights(ctx, userName)
		if err != nil {
			c.collectionErrors.WithLabelValues("rights", "access_rights_error").Inc()
			continue
		}

		if rights != nil {
			// Count different types of rights
			c.userAccessRights.WithLabelValues(userName, "global").Set(float64(len(rights.GlobalPermissions)))
			c.userAccessRights.WithLabelValues(userName, "account").Set(float64(len(rights.AccountPermissions)))
			c.userAccessRights.WithLabelValues(userName, "resource").Set(float64(len(rights.ResourcePermissions)))

			// Count effective permissions
			effectiveCount := 0
			for _, hasPermission := range rights.EffectiveRights {
				if hasPermission {
					effectiveCount++
				}
			}
			c.effectivePermissions.WithLabelValues(userName, rights.PrimaryAccount).Set(float64(effectiveCount))

			// Account-specific permissions
			for account, perms := range rights.AccountPermissions {
				c.accountPermissions.WithLabelValues(userName, account, "total").Set(float64(len(perms)))
			}

			// Permission expiry
			for perm, expiry := range rights.ExpirationDates {
				daysUntilExpiry := time.Until(expiry).Hours() / 24
				c.permissionExpiry.WithLabelValues(userName, rights.PrimaryAccount, perm).Set(daysUntilExpiry)
			}

			// Compliance flags
			c.complianceFlags.WithLabelValues(userName).Set(float64(len(rights.ComplianceFlags)))
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("access_rights").Observe(duration)
	c.lastCollectionTime.WithLabelValues("access_rights").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectPolicyMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample accounts for policy metrics
	sampleAccounts := []string{"research", "engineering", "admin", "public"}

	for _, accountName := range sampleAccounts {
		policies, err := c.client.GetAccountAccessPolicies(ctx, accountName)
		if err != nil {
			c.collectionErrors.WithLabelValues("policy", "access_policy_error").Inc()
			continue
		}

		if policies != nil {
			// Parse and set policy version
			version := parsePolicyVersion(policies.PolicyVersion)
			c.policyVersion.WithLabelValues(accountName).Set(version)

			// Set authentication requirements
			c.minAuthLevel.WithLabelValues(accountName).Set(float64(policies.MinAuthLevel))

			// MFA requirement
			mfaValue := 0.0
			if policies.MFARequired {
				mfaValue = 1.0
			}
			c.mfaRequired.WithLabelValues(accountName).Set(mfaValue)

			// Session settings
			c.sessionTimeout.WithLabelValues(accountName).Set(policies.SessionTimeout.Seconds())
			c.maxConcurrentSessions.WithLabelValues(accountName).Set(float64(policies.MaxConcurrentSessions))
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("policy").Observe(duration)
	c.lastCollectionTime.WithLabelValues("policy").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectAuditMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get audit log for last 24 hours
	auditLog, err := c.client.GetAccessAuditLog(ctx, "24h")
	if err != nil {
		c.collectionErrors.WithLabelValues("audit", "audit_log_error").Inc()
		return
	}

	if auditLog != nil {
		// Process audit events
		for _, event := range auditLog.Events {
			c.auditEvents.WithLabelValues(event.EventType, event.Result).Inc()

			if event.EventType == "login" {
				if event.Result == "success" {
					c.successfulLogins.WithLabelValues(event.AccountName).Inc()
				} else {
					c.failedLogins.WithLabelValues(event.AccountName, event.Result).Inc()
				}
			}

			if event.Flagged {
				severity := getEventSeverity(event.RiskScore)
				c.securityEvents.WithLabelValues(event.EventType, severity).Inc()
			}
		}

		// Set compliance score
		if auditLog.Summary.ComplianceScore > 0 {
			c.auditComplianceScore.WithLabelValues().Set(auditLog.Summary.ComplianceScore)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("audit").Observe(duration)
	c.lastCollectionTime.WithLabelValues("audit").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectViolationMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get all violations
	violations, err := c.client.GetAccessViolations(ctx, "all")
	if err != nil {
		c.collectionErrors.WithLabelValues("violations", "access_violations_error").Inc()
		return
	}

	if violations != nil {
		// Count violations by type and severity
		for _, violation := range violations.Violations {
			c.accessViolations.WithLabelValues(violation.ViolationType, violation.Severity).Inc()

			if violation.AutoBlocked {
				c.autoBlockedAttempts.WithLabelValues(violation.DeniedReason).Inc()
			}

			// Track risk scores by violation type
			c.violationRiskScore.WithLabelValues(violation.ViolationType).Set(violation.RiskScore)
		}

		// Set severity distribution
		c.violationSeverity.WithLabelValues("critical").Set(float64(violations.CriticalViolations))
		c.violationSeverity.WithLabelValues("warning").Set(float64(violations.WarningViolations))
		c.violationSeverity.WithLabelValues("info").Set(float64(violations.InfoViolations))

		// Unresolved violations
		c.unresolvedViolations.WithLabelValues("all").Set(float64(violations.UnresolvedCount))
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("violations").Observe(duration)
	c.lastCollectionTime.WithLabelValues("violations").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectPatternMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample users for pattern analysis
	sampleUsers := []string{"user1", "user2", "admin1"}

	for _, userName := range sampleUsers {
		patterns, err := c.client.GetAccessPatterns(ctx, userName)
		if err != nil {
			c.collectionErrors.WithLabelValues("patterns", "access_patterns_error").Inc()
			continue
		}

		if patterns != nil {
			// Access frequency
			for operation, freq := range patterns.AccessFrequency {
				c.accessFrequency.WithLabelValues(userName, patterns.PreferredAccounts[0], operation).Set(float64(freq))
			}

			// Session length
			c.sessionLength.WithLabelValues(userName).Observe(patterns.AverageSessionLength.Seconds())

			// Behavior score
			c.behaviorScore.WithLabelValues(userName).Set(patterns.BehaviorScore)

			// Anomaly detection
			anomalyValue := 0.0
			if patterns.AnomalyDetected {
				anomalyValue = 1.0
			}
			c.anomalyDetected.WithLabelValues(userName, patterns.RiskProfile).Set(anomalyValue)

			// Peak usage hours
			if len(patterns.PeakUsageHours) > 0 {
				c.peakUsageHours.WithLabelValues(userName).Set(float64(patterns.PeakUsageHours[0]))
			}
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("patterns").Observe(duration)
	c.lastCollectionTime.WithLabelValues("patterns").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectSessionMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample session IDs (in practice, these would come from active session list)
	sampleSessions := []string{"session1", "session2", "session3"}

	activeSessionsByAccount := make(map[string]int)
	mfaCompletedSessions := make(map[string]int)
	totalSessionsByAccount := make(map[string]int)

	for _, sessionID := range sampleSessions {
		session, err := c.client.GetSessionValidation(ctx, sessionID)
		if err != nil {
			continue
		}

		if session != nil {
			// Session validity
			validity := 0.0
			if session.Valid {
				validity = 1.0
				activeSessionsByAccount[session.AccountName]++
			}
			c.sessionValidity.WithLabelValues(sessionID, session.UserName).Set(validity)

			// Session activity
			c.sessionActivity.WithLabelValues(sessionID).Set(float64(session.ActivityCount))

			// Session risk score
			c.sessionRiskScore.WithLabelValues(sessionID, session.UserName).Set(session.RiskScore)

			// Track MFA completion
			totalSessionsByAccount[session.AccountName]++
			if session.MFACompleted {
				mfaCompletedSessions[session.AccountName]++
			}
		}
	}

	// Set active sessions count
	for account, count := range activeSessionsByAccount {
		c.activeSessions.WithLabelValues(account).Set(float64(count))
	}

	// Calculate and set MFA completion rate
	for account, total := range totalSessionsByAccount {
		if total > 0 {
			completed := mfaCompletedSessions[account]
			rate := float64(completed) / float64(total)
			c.mfaCompletionRate.WithLabelValues(account).Set(rate)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("sessions").Observe(duration)
	c.lastCollectionTime.WithLabelValues("sessions").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectComplianceMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get compliance status for different scopes
	scopes := []string{"global", "accounts", "users", "policies"}

	for _, scope := range scopes {
		compliance, err := c.client.GetComplianceStatus(ctx, scope)
		if err != nil {
			c.collectionErrors.WithLabelValues("compliance", "compliance_status_error").Inc()
			continue
		}

		if compliance != nil {
			c.complianceScore.WithLabelValues(scope).Set(compliance.ComplianceScore)
			c.policyAdherence.WithLabelValues(scope).Set(compliance.PolicyAdherence)
			c.securityPosture.WithLabelValues(scope).Set(compliance.SecurityPosture)
			c.controlsInPlace.WithLabelValues(scope).Set(float64(compliance.ControlsInPlace))
			c.controlsFailing.WithLabelValues(scope).Set(float64(compliance.ControlsFailing))

			// Count required actions by priority
			priorityCount := make(map[string]int)
			for _, action := range compliance.RequiredActions {
				priorityCount[action.Priority]++
			}

			for priority, count := range priorityCount {
				c.requiredActions.WithLabelValues(priority).Set(float64(count))
			}
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("compliance").Observe(duration)
	c.lastCollectionTime.WithLabelValues("compliance").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectSecurityAlertMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get security alerts for last 24 hours
	alerts, err := c.client.GetSecurityAlerts(ctx, "24h")
	if err != nil {
		c.collectionErrors.WithLabelValues("alerts", "security_alerts_error").Inc()
		return
	}

	if alerts != nil {
		// Count alerts by type and severity
		severityCount := make(map[string]int)
		activeCount := make(map[string]int)

		for _, alert := range alerts.Alerts {
			c.securityAlerts.WithLabelValues(alert.AlertType, alert.Severity).Inc()

			severityCount[alert.Severity]++

			if alert.Status == "active" {
				activeCount[alert.Severity]++
			}

			// Track resolution time
			if alert.Status == "resolved" && !alert.ResolutionTime.IsZero() {
				resolutionTime := alert.ResolutionTime.Sub(alert.Timestamp).Seconds()
				c.meanTimeToResolve.WithLabelValues(alert.AlertType).Observe(resolutionTime)
			}
		}

		// Set severity distribution
		for severity, count := range severityCount {
			c.alertSeverity.WithLabelValues(severity).Set(float64(count))
		}

		// Set active alerts
		for severity, count := range activeCount {
			c.activeAlerts.WithLabelValues(severity).Set(float64(count))
		}

		// Set false positive rate
		c.falsePositiveRate.WithLabelValues("all").Set(alerts.FalsePositiveRate)
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("security_alerts").Observe(duration)
	c.lastCollectionTime.WithLabelValues("security_alerts").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectRiskAssessmentMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample entities for risk assessment
	entities := []struct {
		entityType string
		entityID   string
	}{
		{"user", "user1"},
		{"account", "research"},
		{"system", "cluster1"},
	}

	for _, entity := range entities {
		risk, err := c.client.GetRiskAssessment(ctx, entity.entityType, entity.entityID)
		if err != nil {
			c.collectionErrors.WithLabelValues("risk", "risk_assessment_error").Inc()
			continue
		}

		if risk != nil {
			// Overall risk score
			c.overallRisk.WithLabelValues(entity.entityType, entity.entityID).Set(risk.OverallRisk)

			// Individual risk factors
			for factor, score := range risk.RiskFactors {
				c.riskFactors.WithLabelValues(entity.entityType, entity.entityID, factor).Set(score)
			}

			// Count vulnerabilities by severity
			vulnCount := make(map[string]int)
			for _, vuln := range risk.Vulnerabilities {
				vulnCount[vuln.Severity]++
			}
			for severity, count := range vulnCount {
				c.vulnerabilities.WithLabelValues(severity).Set(float64(count))
			}

			// Count threats
			threatCount := make(map[string]map[string]int)
			for _, threat := range risk.Threats {
				if threatCount[threat.Type] == nil {
					threatCount[threat.Type] = make(map[string]int)
				}
				threatCount[threat.Type][threat.Status]++
			}
			for threatType, statusCounts := range threatCount {
				for status, count := range statusCounts {
					c.threats.WithLabelValues(threatType, status).Set(float64(count))
				}
			}

			// Count mitigations
			mitigationCount := make(map[string]map[string]int)
			for _, mitigation := range risk.Mitigations {
				if mitigationCount[mitigation.Type] == nil {
					mitigationCount[mitigation.Type] = make(map[string]int)
				}
				mitigationCount[mitigation.Type][mitigation.Status]++
			}
			for mitigationType, statusCounts := range mitigationCount {
				for status, count := range statusCounts {
					c.mitigations.WithLabelValues(mitigationType, status).Set(float64(count))
				}
			}
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("risk_assessment").Observe(duration)
	c.lastCollectionTime.WithLabelValues("risk_assessment").Set(float64(time.Now().Unix()))
}

func (c *AccessValidationCollector) collectTrendMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get access trends for different periods
	trends, err := c.client.GetAccessTrends(ctx, "7d")
	if err != nil {
		c.collectionErrors.WithLabelValues("trends", "access_trends_error").Inc()
		return
	}

	if trends != nil {
		// Set latest trend values
		if len(trends.LoginTrends) > 0 {
			c.loginTrend.WithLabelValues("7d").Set(trends.LoginTrends[len(trends.LoginTrends)-1])
		}
		if len(trends.FailureTrends) > 0 {
			c.failureTrend.WithLabelValues("7d").Set(trends.FailureTrends[len(trends.FailureTrends)-1])
		}
		if len(trends.ViolationTrends) > 0 {
			c.violationTrend.WithLabelValues("7d").Set(trends.ViolationTrends[len(trends.ViolationTrends)-1])
		}
		if len(trends.RiskTrends) > 0 {
			c.riskTrend.WithLabelValues("7d").Set(trends.RiskTrends[len(trends.RiskTrends)-1])
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("trends").Observe(duration)
	c.lastCollectionTime.WithLabelValues("trends").Set(float64(time.Now().Unix()))
}

// Helper functions

func getAuthLevelNumeric(level string) float64 {
	levels := map[string]float64{
		"none":     0,
		"basic":    1,
		"standard": 2,
		"elevated": 3,
		"admin":    4,
	}
	if val, ok := levels[level]; ok {
		return val
	}
	return 0
}

func getPermissionLevelNumeric(level string) float64 {
	levels := map[string]float64{
		"none":    0,
		"view":    1,
		"submit":  2,
		"operate": 3,
		"admin":   4,
		"owner":   5,
	}
	if val, ok := levels[level]; ok {
		return val
	}
	return 0
}

func parsePolicyVersion(version string) float64 {
	// Simple version parsing - in practice would be more sophisticated
	// Assumes format like "1.2.3" and returns 1.23
	parts := strings.Split(version, ".")
	if len(parts) >= 2 {
		major := 0.0
		minor := 0.0
		// Parse major and minor version numbers
		// Simplified for example
		return major + minor/100
	}
	return 0
}

func getEventSeverity(riskScore float64) string {
	if riskScore >= 0.8 {
		return "critical"
	} else if riskScore >= 0.6 {
		return "high"
	} else if riskScore >= 0.4 {
		return "medium"
	} else if riskScore >= 0.2 {
		return "low"
	}
	return "info"
}
