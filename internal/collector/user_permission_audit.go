// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type UserPermissionAuditSLURMClient interface {
	GetUserPermissions(ctx context.Context, userName string) (*UserPermissions, error)
	GetUserAccessAuditLog(ctx context.Context, userName string, startTime, endTime time.Time) ([]*UserAccessAuditEntry, error)
	GetUserAccessPatterns(ctx context.Context, userName string) (*UserAccessPatterns, error)
	GetUserPermissionHistory(ctx context.Context, userName string, startTime, endTime time.Time) ([]*UserPermissionHistoryEntry, error)
	GetUserAccessViolations(ctx context.Context, userName string) ([]*UserAccessViolation, error)
	GetUserSessionAudit(ctx context.Context, userName string) (*UserSessionAudit, error)
	GetUserComplianceStatus(ctx context.Context, userName string) (*UserComplianceStatus, error)
	GetUserSecurityEvents(ctx context.Context, userName string) ([]*UserSecurityEvent, error)
	GetUserRoleAssignments(ctx context.Context, userName string) ([]*UserRoleAssignment, error)
	GetUserPrivilegeEscalations(ctx context.Context, userName string) ([]*UserPrivilegeEscalation, error)
	GetUserDataAccess(ctx context.Context, userName string) (*UserDataAccess, error)
	GetUserAccountAssociations(ctx context.Context, userName string) ([]*UserAccountAssociation, error)
	GetUserAuthenticationEvents(ctx context.Context, userName string) ([]*UserAuthenticationEvent, error)
	GetUserResourceAccess(ctx context.Context, userName string) (*UserResourceAccess, error)
}

type UserPermissions struct {
	UserName               string
	ActivePermissions      []string
	InheritedPermissions   []string
	ExplicitPermissions    []string
	DeniedPermissions      []string
	TemporaryPermissions   []string
	ConditionalPermissions []string
	PermissionLevel        string
	EffectivePermissions   map[string]bool
	PermissionScope        string
	ExpirationTime         *time.Time
	GrantedBy              string
	GrantedTime            time.Time
	LastValidated          time.Time
	ValidationStatus       string
	PermissionSource       string
	ComplianceStatus       string
	RiskLevel              string
	LastModified           time.Time
	ModifiedBy             string
	AuditTrail             []string
}

type UserAccessAuditEntry struct {
	EntryID              string
	UserName             string
	AccessType           string
	ResourceAccessed     string
	Action               string
	Result               string
	Timestamp            time.Time
	SourceIP             string
	UserAgent            string
	SessionID            string
	RequestMethod        string
	ResponseCode         int
	ResponseSize         int64
	ProcessingTime       time.Duration
	ErrorMessage         string
	SecurityContext      string
	AuthenticationMethod string
	AuthorizationResult  string
	DataSensitivity      string
	ComplianceFlags      []string
	RiskScore            float64
	AnomalyScore         float64
	GeoLocation          string
	DeviceInfo           string
	NetworkInfo          string
}

type UserAccessPatterns struct {
	UserName               string
	LoginFrequency         float64
	AverageSessionDuration time.Duration
	PeakAccessHours        []int
	AccessDays             []string
	ResourceUsagePatterns  map[string]float64
	AccessLocations        []string
	DevicePatterns         []string
	CommandPatterns        []string
	DataAccessPatterns     map[string]float64
	NetworkPatterns        []string
	BehaviorBaseline       map[string]float64
	AnomalyCount           int
	LastAnomaly            time.Time
	PatternStability       float64
	RiskIndicators         []string
	CompliancePatterns     map[string]float64
	EfficiencyMetrics      map[string]float64
	UsageVariability       float64
	SeasonalPatterns       map[string]float64
	PredictiveInsights     map[string]float64
}

type UserPermissionHistoryEntry struct {
	HistoryID        string
	UserName         string
	PermissionName   string
	Action           string
	OldValue         string
	NewValue         string
	ChangeReason     string
	ChangedBy        string
	ChangeTime       time.Time
	ApprovalStatus   string
	ApprovedBy       string
	ApprovalTime     time.Time
	EffectiveTime    time.Time
	ExpirationTime   *time.Time
	RollbackTime     *time.Time
	RollbackReason   string
	ImpactAssessment string
	ComplianceImpact string
	RiskAssessment   string
	ValidationStatus string
	AuditNotes       string
}

type UserAccessViolation struct {
	ViolationID         string
	UserName            string
	ViolationType       string
	Severity            string
	Description         string
	DetectedTime        time.Time
	ResourceInvolved    string
	PolicyViolated      string
	AccessAttempted     string
	ActualPermissions   []string
	RequiredPermissions []string
	ViolationSource     string
	DetectionMethod     string
	InvestigationStatus string
	ResolutionStatus    string
	ResolutionTime      *time.Time
	ResolutionAction    string
	ResponsibleParty    string
	ImpactAssessment    string
	RemediationSteps    []string
	PreventionMeasures  []string
	ComplianceImpact    string
	RiskScore           float64
	RecurrenceCount     int
	RelatedViolations   []string
	EvidenceCollected   []string
	AuditTrail          []string
}

type UserSessionAudit struct {
	UserName              string
	ActiveSessions        int
	TotalSessions         int
	AverageSessionTime    time.Duration
	LongestSession        time.Duration
	ShortestSession       time.Duration
	SessionLocations      []string
	SessionDevices        []string
	ConcurrentSessions    int
	MaxConcurrentSessions int
	SessionAnomalies      int
	SuspiciousSessions    int
	SecurityViolations    int
	ComplianceIssues      int
	LastSessionStart      time.Time
	LastSessionEnd        time.Time
	SessionEfficiency     float64
	ResourceUsage         map[string]float64
	NetworkActivity       map[string]float64
	DataTransfer          map[string]int64
	SessionRiskScore      float64
	SessionPatterns       map[string]float64
}

type UserComplianceStatus struct {
	UserName                string
	OverallComplianceScore  float64
	PolicyCompliance        map[string]float64
	RequiredCertifications  []string
	CompletedCertifications []string
	ExpiredCertifications   []string
	PendingCertifications   []string
	TrainingRequirements    []string
	CompletedTraining       []string
	PendingTraining         []string
	AcknowledgedPolicies    []string
	PendingAcknowledgments  []string
	ComplianceViolations    int
	LastComplianceReview    time.Time
	NextComplianceReview    time.Time
	ComplianceRiskLevel     string
	RemediationActions      []string
	ComplianceExceptions    []string
	AuditFindings           []string
	CorrectiveActions       []string
	ComplianceTrends        map[string]float64
	ComplianceHistory       []string
}

type UserSecurityEvent struct {
	EventID             string
	UserName            string
	EventType           string
	Severity            string
	EventTime           time.Time
	SourceSystem        string
	EventDescription    string
	SourceIP            string
	TargetResource      string
	ActionTaken         string
	EventOutcome        string
	ThreatLevel         string
	AttackVector        string
	Indicators          []string
	AffectedSystems     []string
	DataInvolved        string
	BusinessImpact      string
	TechnicalImpact     string
	ResponseTime        time.Duration
	ResolutionTime      time.Duration
	InvestigationStatus string
	ForensicEvidence    []string
	ResponseActions     []string
	LessonsLearned      []string
	RelatedEvents       []string
	ComplianceImpact    string
}

type UserRoleAssignment struct {
	AssignmentID          string
	UserName              string
	RoleName              string
	RoleType              string
	AssignmentType        string
	AssignedBy            string
	AssignedTime          time.Time
	EffectiveTime         time.Time
	ExpirationTime        *time.Time
	AssignmentStatus      string
	ApprovalWorkflow      string
	ApprovedBy            string
	ApprovalTime          time.Time
	RolePermissions       []string
	RoleScope             string
	RoleLevel             string
	InheritedRoles        []string
	ConflictingRoles      []string
	RoleDependencies      []string
	BusinessJustification string
	RiskAssessment        string
	ComplianceImpact      string
	ReviewFrequency       string
	LastReviewed          time.Time
	NextReview            time.Time
	ReviewStatus          string
	UsageMetrics          map[string]float64
}

type UserPrivilegeEscalation struct {
	EscalationID        string
	UserName            string
	EscalationType      string
	RequestTime         time.Time
	RequestedBy         string
	OriginalPrivileges  []string
	RequestedPrivileges []string
	GrantedPrivileges   []string
	Justification       string
	BusinessNeed        string
	ApprovalRequired    bool
	ApprovedBy          string
	ApprovalTime        time.Time
	EffectiveTime       time.Time
	ExpirationTime      *time.Time
	DurationRequested   time.Duration
	ActualDuration      time.Duration
	AutoRevocation      bool
	RevocationTime      *time.Time
	RevocationReason    string
	RiskAssessment      string
	ImpactAnalysis      string
	MonitoringLevel     string
	UsageTracking       bool
	ComplianceReview    bool
	AuditRequirement    string
	EscalationPattern   string
	AbuseIndicators     []string
	SecurityValidation  string
}

type UserDataAccess struct {
	UserName                   string
	DataClassifications        []string
	AccessibleDataSets         []string
	RestrictedDataSets         []string
	DataAccessLevel            string
	SensitiveDataAccess        bool
	PIIAccess                  bool
	PHIAccess                  bool
	FinancialDataAccess        bool
	IntellectualPropertyAccess bool
	ClassifiedDataAccess       bool
	DataDownloadCount          int64
	DataUploadCount            int64
	DataViewCount              int64
	DataModificationCount      int64
	DataDeletionCount          int64
	DataSharingCount           int64
	DataExportCount            int64
	LastDataAccess             time.Time
	DataAccessPatterns         map[string]float64
	DataRetentionCompliance    bool
	DataMinimizationCompliance bool
	ConsentManagement          map[string]bool
	DataBreachInvolvement      []string
	DataGovernanceScore        float64
	DataComplianceStatus       string
}

type UserPermissionAccountAssociation struct {
	AssociationID           string
	UserName                string
	AccountName             string
	AssociationType         string
	RelationshipType        string
	PrimaryAccount          bool
	DefaultAccount          bool
	AssociationStatus       string
	AssociatedBy            string
	AssociationTime         time.Time
	EffectiveTime           time.Time
	ExpirationTime          *time.Time
	AccountPermissions      []string
	InheritedPermissions    []string
	PermissionSource        string
	AccessLevel             string
	AccountRole             string
	BillingResponsibility   bool
	ResourceQuotas          map[string]float64
	UsageMetrics            map[string]float64
	AccountComplianceStatus string
	LastActivity            time.Time
	ActivityMetrics         map[string]float64
	PerformanceMetrics      map[string]float64
	CostMetrics             map[string]float64
	RiskMetrics             map[string]float64
	AssociationHistory      []string
}

type UserAuthenticationEvent struct {
	EventID                string
	UserName               string
	AuthenticationType     string
	AuthenticationMethod   string
	EventTime              time.Time
	SourceIP               string
	SourceLocation         string
	DeviceInfo             string
	UserAgent              string
	AuthenticationResult   string
	FailureReason          string
	AttemptCount           int
	SessionID              string
	TokenUsed              string
	CertificateUsed        string
	BiometricUsed          bool
	MultiFactorUsed        bool
	RiskScore              float64
	TrustScore             float64
	AnomalyScore           float64
	GeoLocationRisk        float64
	DeviceRisk             float64
	BehaviorRisk           float64
	TimeRisk               float64
	NetworkRisk            float64
	ThreatIntelligence     []string
	SecurityFlags          []string
	ComplianceFlags        []string
	BusinessContext        string
	AuthenticationDuration time.Duration
	TokenLifetime          time.Duration
	SessionDuration        time.Duration
}

type UserResourceAccess struct {
	UserName                string
	AccessibleResources     []string
	RestrictedResources     []string
	ResourcePermissions     map[string][]string
	ResourceUsage           map[string]float64
	ResourceQuotas          map[string]float64
	QuotaUtilization        map[string]float64
	ResourceEfficiency      map[string]float64
	ResourceWaste           map[string]float64
	ResourceCosts           map[string]float64
	PeakResourceUsage       map[string]float64
	ResourceAccessFrequency map[string]int64
	LastResourceAccess      map[string]time.Time
	ResourceAccessPatterns  map[string]float64
	ResourceViolations      map[string]int
	ResourceCompliance      map[string]float64
	ResourceSecurity        map[string]float64
	ResourceAvailability    map[string]float64
	ResourcePerformance     map[string]float64
	ResourceOptimization    map[string]float64
	ResourceRecommendations map[string][]string
	ResourceRiskAssessment  map[string]float64
	ResourceTrends          map[string]float64
}

type UserPermissionAuditCollector struct {
	client UserPermissionAuditSLURMClient
	mutex  sync.RWMutex

	// Permission metrics
	activePermissions      *prometheus.GaugeVec
	inheritedPermissions   *prometheus.GaugeVec
	explicitPermissions    *prometheus.GaugeVec
	deniedPermissions      *prometheus.GaugeVec
	temporaryPermissions   *prometheus.GaugeVec
	conditionalPermissions *prometheus.GaugeVec
	permissionLevel        *prometheus.GaugeVec
	permissionScope        *prometheus.GaugeVec
	permissionRiskLevel    *prometheus.GaugeVec
	permissionCompliance   *prometheus.GaugeVec

	// Audit metrics
	auditEntriesTotal        *prometheus.CounterVec
	auditEntryResponseCode   *prometheus.HistogramVec
	auditEntryProcessingTime *prometheus.HistogramVec
	auditEntryRiskScore      *prometheus.GaugeVec
	auditEntryAnomalyScore   *prometheus.GaugeVec
	dataAccessTotal          *prometheus.CounterVec
	securityViolationsTotal  *prometheus.CounterVec

	// Access pattern metrics
	loginFrequency         *prometheus.GaugeVec
	averageSessionDuration *prometheus.GaugeVec
	accessAnomalies        *prometheus.GaugeVec
	patternStability       *prometheus.GaugeVec
	usageVariability       *prometheus.GaugeVec
	behaviorRiskIndicators *prometheus.GaugeVec
	accessEfficiency       *prometheus.GaugeVec

	// Permission history metrics
	permissionChangesTotal   *prometheus.CounterVec
	permissionApprovalsTotal *prometheus.CounterVec
	permissionRollbacksTotal *prometheus.CounterVec
	permissionChangeRisk     *prometheus.GaugeVec
	permissionChangeImpact   *prometheus.GaugeVec

	// Violation metrics
	accessViolationsTotal   *prometheus.CounterVec
	violationSeverity       *prometheus.GaugeVec
	violationRiskScore      *prometheus.GaugeVec
	violationRecurrence     *prometheus.GaugeVec
	violationResolutionTime *prometheus.HistogramVec
	violationImpactScore    *prometheus.GaugeVec

	// Session metrics
	activeSessions        *prometheus.GaugeVec
	totalSessions         *prometheus.GaugeVec
	maxConcurrentSessions *prometheus.GaugeVec
	sessionAnomalies      *prometheus.GaugeVec
	suspiciousSessions    *prometheus.GaugeVec
	sessionEfficiency     *prometheus.GaugeVec
	sessionRiskScore      *prometheus.GaugeVec

	// Compliance metrics
	complianceScore         *prometheus.GaugeVec
	certificationCompliance *prometheus.GaugeVec
	trainingCompliance      *prometheus.GaugeVec
	policyCompliance        *prometheus.GaugeVec
	complianceViolations    *prometheus.GaugeVec
	complianceRiskLevel     *prometheus.GaugeVec
	remediationActionsTotal *prometheus.CounterVec

	// Security event metrics
	securityEventsTotal   *prometheus.CounterVec
	securityEventSeverity *prometheus.GaugeVec
	threatLevel           *prometheus.GaugeVec
	attackVectorCount     *prometheus.CounterVec
	responseTime          *prometheus.HistogramVec
	resolutionTime        *prometheus.HistogramVec
	businessImpactScore   *prometheus.GaugeVec

	// Role assignment metrics
	roleAssignmentsTotal *prometheus.CounterVec
	activeRoles          *prometheus.GaugeVec
	roleConflicts        *prometheus.GaugeVec
	roleUsageMetrics     *prometheus.GaugeVec
	roleRiskAssessment   *prometheus.GaugeVec
	roleComplianceScore  *prometheus.GaugeVec

	// Privilege escalation metrics
	privilegeEscalationsTotal *prometheus.CounterVec
	escalationRiskScore       *prometheus.GaugeVec
	escalationDuration        *prometheus.HistogramVec
	abuseIndicators           *prometheus.GaugeVec
	autoRevocations           *prometheus.CounterVec
	escalationPatterns        *prometheus.GaugeVec

	// Data access metrics
	dataClassificationAccess *prometheus.GaugeVec
	sensitiveDataAccess      *prometheus.GaugeVec
	dataOperationsTotal      *prometheus.CounterVec
	dataGovernanceScore      *prometheus.GaugeVec
	dataComplianceStatus     *prometheus.GaugeVec
	dataBreachInvolvement    *prometheus.GaugeVec

	// Account association metrics
	accountAssociationsTotal  *prometheus.CounterVec
	primaryAccountAssignments *prometheus.GaugeVec
	accountPermissionLevel    *prometheus.GaugeVec
	accountUsageMetrics       *prometheus.GaugeVec
	accountCostMetrics        *prometheus.GaugeVec
	accountRiskMetrics        *prometheus.GaugeVec

	// Authentication metrics
	authenticationEventsTotal *prometheus.CounterVec
	authenticationRiskScore   *prometheus.GaugeVec
	authenticationTrustScore  *prometheus.GaugeVec
	authenticationAnomalies   *prometheus.GaugeVec
	multiFactorUsage          *prometheus.GaugeVec
	authenticationDuration    *prometheus.HistogramVec
	tokenLifetime             *prometheus.HistogramVec

	// Resource access metrics
	resourceAccessTotal      *prometheus.CounterVec
	resourceQuotaUtilization *prometheus.GaugeVec
	resourceEfficiency       *prometheus.GaugeVec
	resourceWaste            *prometheus.GaugeVec
	resourceCosts            *prometheus.GaugeVec
	resourceViolations       *prometheus.CounterVec
	resourceCompliance       *prometheus.GaugeVec
	resourceSecurity         *prometheus.GaugeVec
	resourcePerformance      *prometheus.GaugeVec
	resourceOptimization     *prometheus.GaugeVec
}

func NewUserPermissionAuditCollector(client UserPermissionAuditSLURMClient) *UserPermissionAuditCollector {
	return &UserPermissionAuditCollector{
		client: client,

		// Permission metrics
		activePermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_active_permissions",
				Help: "Number of active permissions for user",
			},
			[]string{"user"},
		),
		inheritedPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_inherited_permissions",
				Help: "Number of inherited permissions for user",
			},
			[]string{"user"},
		),
		explicitPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_explicit_permissions",
				Help: "Number of explicit permissions for user",
			},
			[]string{"user"},
		),
		deniedPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_denied_permissions",
				Help: "Number of denied permissions for user",
			},
			[]string{"user"},
		),
		temporaryPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_temporary_permissions",
				Help: "Number of temporary permissions for user",
			},
			[]string{"user"},
		),
		conditionalPermissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_conditional_permissions",
				Help: "Number of conditional permissions for user",
			},
			[]string{"user"},
		),
		permissionLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_level",
				Help: "Permission level for user (0=basic, 1=standard, 2=advanced, 3=admin)",
			},
			[]string{"user", "level"},
		),
		permissionScope: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_scope",
				Help: "Permission scope for user (0=limited, 1=departmental, 2=organizational, 3=global)",
			},
			[]string{"user", "scope"},
		),
		permissionRiskLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_risk_level",
				Help: "Permission risk level for user (0=low, 1=medium, 2=high, 3=critical)",
			},
			[]string{"user", "risk_level"},
		),
		permissionCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_compliance_score",
				Help: "Permission compliance score for user",
			},
			[]string{"user"},
		),

		// Audit metrics
		auditEntriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_audit_entries_total",
				Help: "Total number of audit entries for user",
			},
			[]string{"user", "access_type", "result"},
		),
		auditEntryResponseCode: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_audit_response_code",
				Help:    "Distribution of response codes in audit entries",
				Buckets: []float64{200, 300, 400, 401, 403, 404, 500},
			},
			[]string{"user"},
		),
		auditEntryProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_audit_processing_time_seconds",
				Help:    "Processing time for audit entries",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"user", "access_type"},
		),
		auditEntryRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_audit_risk_score",
				Help: "Risk score for audit entries",
			},
			[]string{"user"},
		),
		auditEntryAnomalyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_audit_anomaly_score",
				Help: "Anomaly score for audit entries",
			},
			[]string{"user"},
		),
		dataAccessTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_data_access_total",
				Help: "Total data access operations by user",
			},
			[]string{"user", "data_type", "operation"},
		),
		securityViolationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_security_violations_total",
				Help: "Total security violations by user",
			},
			[]string{"user", "violation_type"},
		),

		// Access pattern metrics
		loginFrequency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_login_frequency",
				Help: "Login frequency for user",
			},
			[]string{"user"},
		),
		averageSessionDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_average_session_duration_seconds",
				Help: "Average session duration for user",
			},
			[]string{"user"},
		),
		accessAnomalies: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_access_anomalies",
				Help: "Number of access anomalies for user",
			},
			[]string{"user"},
		),
		patternStability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_pattern_stability",
				Help: "Pattern stability score for user",
			},
			[]string{"user"},
		),
		usageVariability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_usage_variability",
				Help: "Usage variability score for user",
			},
			[]string{"user"},
		),
		behaviorRiskIndicators: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_behavior_risk_indicators",
				Help: "Number of behavior risk indicators for user",
			},
			[]string{"user"},
		),
		accessEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_access_efficiency",
				Help: "Access efficiency score for user",
			},
			[]string{"user"},
		),

		// Permission history metrics
		permissionChangesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_permission_changes_total",
				Help: "Total permission changes for user",
			},
			[]string{"user", "change_type"},
		),
		permissionApprovalsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_permission_approvals_total",
				Help: "Total permission approvals for user",
			},
			[]string{"user", "approval_status"},
		),
		permissionRollbacksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_permission_rollbacks_total",
				Help: "Total permission rollbacks for user",
			},
			[]string{"user"},
		),
		permissionChangeRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_change_risk_score",
				Help: "Risk score for permission changes",
			},
			[]string{"user"},
		),
		permissionChangeImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_permission_change_impact_score",
				Help: "Impact score for permission changes",
			},
			[]string{"user"},
		),

		// Violation metrics
		accessViolationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_access_violations_total",
				Help: "Total access violations for user",
			},
			[]string{"user", "violation_type", "severity"},
		),
		violationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_severity_score",
				Help: "Violation severity score for user",
			},
			[]string{"user", "severity"},
		),
		violationRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_risk_score",
				Help: "Violation risk score for user",
			},
			[]string{"user"},
		),
		violationRecurrence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_recurrence_count",
				Help: "Violation recurrence count for user",
			},
			[]string{"user", "violation_type"},
		),
		violationResolutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_violation_resolution_time_seconds",
				Help:    "Time to resolve violations",
				Buckets: []float64{3600, 7200, 14400, 28800, 86400, 172800, 604800},
			},
			[]string{"user", "violation_type"},
		),
		violationImpactScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_impact_score",
				Help: "Violation impact score for user",
			},
			[]string{"user"},
		),

		// Session metrics
		activeSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_active_sessions",
				Help: "Number of active sessions for user",
			},
			[]string{"user"},
		),
		totalSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_total_sessions",
				Help: "Total number of sessions for user",
			},
			[]string{"user"},
		),
		maxConcurrentSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_max_concurrent_sessions",
				Help: "Maximum concurrent sessions for user",
			},
			[]string{"user"},
		),
		sessionAnomalies: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_session_anomalies",
				Help: "Number of session anomalies for user",
			},
			[]string{"user"},
		),
		suspiciousSessions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_suspicious_sessions",
				Help: "Number of suspicious sessions for user",
			},
			[]string{"user"},
		),
		sessionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_session_efficiency",
				Help: "Session efficiency score for user",
			},
			[]string{"user"},
		),
		sessionRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_session_risk_score",
				Help: "Session risk score for user",
			},
			[]string{"user"},
		),

		// Compliance metrics
		complianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_compliance_score",
				Help: "Overall compliance score for user",
			},
			[]string{"user"},
		),
		certificationCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_certification_compliance",
				Help: "Certification compliance percentage for user",
			},
			[]string{"user"},
		),
		trainingCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_training_compliance",
				Help: "Training compliance percentage for user",
			},
			[]string{"user"},
		),
		policyCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_policy_compliance",
				Help: "Policy compliance score for user",
			},
			[]string{"user", "policy_type"},
		),
		complianceViolations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_compliance_violations",
				Help: "Number of compliance violations for user",
			},
			[]string{"user"},
		),
		complianceRiskLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_compliance_risk_level",
				Help: "Compliance risk level for user (0=low, 1=medium, 2=high, 3=critical)",
			},
			[]string{"user", "risk_level"},
		),
		remediationActionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_remediation_actions_total",
				Help: "Total remediation actions for user",
			},
			[]string{"user", "action_type"},
		),

		// Security event metrics
		securityEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_security_events_total",
				Help: "Total security events for user",
			},
			[]string{"user", "event_type", "severity"},
		),
		securityEventSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_security_event_severity_score",
				Help: "Security event severity score for user",
			},
			[]string{"user", "severity"},
		),
		threatLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_threat_level",
				Help: "Threat level for user (0=low, 1=medium, 2=high, 3=critical)",
			},
			[]string{"user", "threat_level"},
		),
		attackVectorCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_attack_vectors_total",
				Help: "Total attack vectors detected for user",
			},
			[]string{"user", "attack_vector"},
		),
		responseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_security_response_time_seconds",
				Help:    "Security incident response time",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400},
			},
			[]string{"user", "event_type"},
		),
		resolutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_security_resolution_time_seconds",
				Help:    "Security incident resolution time",
				Buckets: []float64{3600, 7200, 14400, 28800, 86400, 172800, 604800},
			},
			[]string{"user", "event_type"},
		),
		businessImpactScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_security_business_impact_score",
				Help: "Security event business impact score",
			},
			[]string{"user"},
		),

		// Role assignment metrics
		roleAssignmentsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_role_assignments_total",
				Help: "Total role assignments for user",
			},
			[]string{"user", "role_type", "assignment_type"},
		),
		activeRoles: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_active_roles",
				Help: "Number of active roles for user",
			},
			[]string{"user"},
		),
		roleConflicts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_role_conflicts",
				Help: "Number of role conflicts for user",
			},
			[]string{"user"},
		),
		roleUsageMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_role_usage_score",
				Help: "Role usage score for user",
			},
			[]string{"user", "role_name"},
		),
		roleRiskAssessment: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_role_risk_score",
				Help: "Role risk assessment score for user",
			},
			[]string{"user", "role_name"},
		),
		roleComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_role_compliance_score",
				Help: "Role compliance score for user",
			},
			[]string{"user", "role_name"},
		),

		// Privilege escalation metrics
		privilegeEscalationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_privilege_escalations_total",
				Help: "Total privilege escalations for user",
			},
			[]string{"user", "escalation_type"},
		),
		escalationRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_escalation_risk_score",
				Help: "Privilege escalation risk score for user",
			},
			[]string{"user"},
		),
		escalationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_escalation_duration_seconds",
				Help:    "Duration of privilege escalations",
				Buckets: []float64{3600, 7200, 14400, 28800, 86400, 172800, 604800},
			},
			[]string{"user", "escalation_type"},
		),
		abuseIndicators: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_abuse_indicators",
				Help: "Number of abuse indicators for user",
			},
			[]string{"user"},
		),
		autoRevocations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_auto_revocations_total",
				Help: "Total automatic revocations for user",
			},
			[]string{"user"},
		),
		escalationPatterns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_escalation_patterns",
				Help: "Escalation pattern score for user",
			},
			[]string{"user"},
		),

		// Data access metrics
		dataClassificationAccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_data_classification_access",
				Help: "Data classification access level for user",
			},
			[]string{"user", "classification"},
		),
		sensitiveDataAccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_sensitive_data_access",
				Help: "Sensitive data access permissions for user",
			},
			[]string{"user", "data_type"},
		),
		dataOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_data_operations_total",
				Help: "Total data operations by user",
			},
			[]string{"user", "operation_type"},
		),
		dataGovernanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_data_governance_score",
				Help: "Data governance score for user",
			},
			[]string{"user"},
		),
		dataComplianceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_data_compliance_status",
				Help: "Data compliance status for user",
			},
			[]string{"user", "compliance_type"},
		),
		dataBreachInvolvement: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_data_breach_involvement",
				Help: "Data breach involvement score for user",
			},
			[]string{"user"},
		),

		// Account association metrics
		accountAssociationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_account_associations_total",
				Help: "Total account associations for user",
			},
			[]string{"user", "association_type"},
		),
		primaryAccountAssignments: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_primary_account_assignments",
				Help: "Number of primary account assignments for user",
			},
			[]string{"user"},
		),
		accountPermissionLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_account_permission_level",
				Help: "Account permission level for user",
			},
			[]string{"user", "account"},
		),
		accountUsageMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_account_usage_score",
				Help: "Account usage score for user",
			},
			[]string{"user", "account"},
		),
		accountCostMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_account_cost_score",
				Help: "Account cost score for user",
			},
			[]string{"user", "account"},
		),
		accountRiskMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_account_risk_score",
				Help: "Account risk score for user",
			},
			[]string{"user", "account"},
		),

		// Authentication metrics
		authenticationEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_authentication_events_total",
				Help: "Total authentication events for user",
			},
			[]string{"user", "auth_type", "result"},
		),
		authenticationRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_authentication_risk_score",
				Help: "Authentication risk score for user",
			},
			[]string{"user"},
		),
		authenticationTrustScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_authentication_trust_score",
				Help: "Authentication trust score for user",
			},
			[]string{"user"},
		),
		authenticationAnomalies: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_authentication_anomalies",
				Help: "Authentication anomaly score for user",
			},
			[]string{"user"},
		),
		multiFactorUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_multi_factor_usage",
				Help: "Multi-factor authentication usage rate for user",
			},
			[]string{"user"},
		),
		authenticationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_authentication_duration_seconds",
				Help:    "Authentication duration for user",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"user", "auth_type"},
		),
		tokenLifetime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_user_token_lifetime_seconds",
				Help:    "Token lifetime for user",
				Buckets: []float64{300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"user"},
		),

		// Resource access metrics
		resourceAccessTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_resource_access_total",
				Help: "Total resource access operations by user",
			},
			[]string{"user", "resource_type"},
		),
		resourceQuotaUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_quota_utilization",
				Help: "Resource quota utilization for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_efficiency",
				Help: "Resource efficiency score for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceWaste: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_waste",
				Help: "Resource waste score for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceCosts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_costs",
				Help: "Resource costs for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_user_resource_violations_total",
				Help: "Total resource violations for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_compliance",
				Help: "Resource compliance score for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceSecurity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_security",
				Help: "Resource security score for user",
			},
			[]string{"user", "resource_type"},
		),
		resourcePerformance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_performance",
				Help: "Resource performance score for user",
			},
			[]string{"user", "resource_type"},
		),
		resourceOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_optimization",
				Help: "Resource optimization score for user",
			},
			[]string{"user", "resource_type"},
		),
	}
}

func (c *UserPermissionAuditCollector) Describe(ch chan<- *prometheus.Desc) {
	c.activePermissions.Describe(ch)
	c.inheritedPermissions.Describe(ch)
	c.explicitPermissions.Describe(ch)
	c.deniedPermissions.Describe(ch)
	c.temporaryPermissions.Describe(ch)
	c.conditionalPermissions.Describe(ch)
	c.permissionLevel.Describe(ch)
	c.permissionScope.Describe(ch)
	c.permissionRiskLevel.Describe(ch)
	c.permissionCompliance.Describe(ch)
	c.auditEntriesTotal.Describe(ch)
	c.auditEntryResponseCode.Describe(ch)
	c.auditEntryProcessingTime.Describe(ch)
	c.auditEntryRiskScore.Describe(ch)
	c.auditEntryAnomalyScore.Describe(ch)
	c.dataAccessTotal.Describe(ch)
	c.securityViolationsTotal.Describe(ch)
	c.loginFrequency.Describe(ch)
	c.averageSessionDuration.Describe(ch)
	c.accessAnomalies.Describe(ch)
	c.patternStability.Describe(ch)
	c.usageVariability.Describe(ch)
	c.behaviorRiskIndicators.Describe(ch)
	c.accessEfficiency.Describe(ch)
	c.permissionChangesTotal.Describe(ch)
	c.permissionApprovalsTotal.Describe(ch)
	c.permissionRollbacksTotal.Describe(ch)
	c.permissionChangeRisk.Describe(ch)
	c.permissionChangeImpact.Describe(ch)
	c.accessViolationsTotal.Describe(ch)
	c.violationSeverity.Describe(ch)
	c.violationRiskScore.Describe(ch)
	c.violationRecurrence.Describe(ch)
	c.violationResolutionTime.Describe(ch)
	c.violationImpactScore.Describe(ch)
	c.activeSessions.Describe(ch)
	c.totalSessions.Describe(ch)
	c.maxConcurrentSessions.Describe(ch)
	c.sessionAnomalies.Describe(ch)
	c.suspiciousSessions.Describe(ch)
	c.sessionEfficiency.Describe(ch)
	c.sessionRiskScore.Describe(ch)
	c.complianceScore.Describe(ch)
	c.certificationCompliance.Describe(ch)
	c.trainingCompliance.Describe(ch)
	c.policyCompliance.Describe(ch)
	c.complianceViolations.Describe(ch)
	c.complianceRiskLevel.Describe(ch)
	c.remediationActionsTotal.Describe(ch)
	c.securityEventsTotal.Describe(ch)
	c.securityEventSeverity.Describe(ch)
	c.threatLevel.Describe(ch)
	c.attackVectorCount.Describe(ch)
	c.responseTime.Describe(ch)
	c.resolutionTime.Describe(ch)
	c.businessImpactScore.Describe(ch)
	c.roleAssignmentsTotal.Describe(ch)
	c.activeRoles.Describe(ch)
	c.roleConflicts.Describe(ch)
	c.roleUsageMetrics.Describe(ch)
	c.roleRiskAssessment.Describe(ch)
	c.roleComplianceScore.Describe(ch)
	c.privilegeEscalationsTotal.Describe(ch)
	c.escalationRiskScore.Describe(ch)
	c.escalationDuration.Describe(ch)
	c.abuseIndicators.Describe(ch)
	c.autoRevocations.Describe(ch)
	c.escalationPatterns.Describe(ch)
	c.dataClassificationAccess.Describe(ch)
	c.sensitiveDataAccess.Describe(ch)
	c.dataOperationsTotal.Describe(ch)
	c.dataGovernanceScore.Describe(ch)
	c.dataComplianceStatus.Describe(ch)
	c.dataBreachInvolvement.Describe(ch)
	c.accountAssociationsTotal.Describe(ch)
	c.primaryAccountAssignments.Describe(ch)
	c.accountPermissionLevel.Describe(ch)
	c.accountUsageMetrics.Describe(ch)
	c.accountCostMetrics.Describe(ch)
	c.accountRiskMetrics.Describe(ch)
	c.authenticationEventsTotal.Describe(ch)
	c.authenticationRiskScore.Describe(ch)
	c.authenticationTrustScore.Describe(ch)
	c.authenticationAnomalies.Describe(ch)
	c.multiFactorUsage.Describe(ch)
	c.authenticationDuration.Describe(ch)
	c.tokenLifetime.Describe(ch)
	c.resourceAccessTotal.Describe(ch)
	c.resourceQuotaUtilization.Describe(ch)
	c.resourceEfficiency.Describe(ch)
	c.resourceWaste.Describe(ch)
	c.resourceCosts.Describe(ch)
	c.resourceViolations.Describe(ch)
	c.resourceCompliance.Describe(ch)
	c.resourceSecurity.Describe(ch)
	c.resourcePerformance.Describe(ch)
	c.resourceOptimization.Describe(ch)
}

func (c *UserPermissionAuditCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()
	users := []string{"user1", "user2", "user3"}

	for _, user := range users {
		c.collectPermissionMetrics(ctx, user, ch)
		c.collectAuditMetrics(ctx, user, ch)
		c.collectAccessPatterns(ctx, user, ch)
		c.collectPermissionHistory(ctx, user, ch)
		c.collectViolationMetrics(ctx, user, ch)
		c.collectSessionMetrics(ctx, user, ch)
		c.collectComplianceMetrics(ctx, user, ch)
		c.collectSecurityEvents(ctx, user, ch)
		c.collectRoleMetrics(ctx, user, ch)
		c.collectEscalationMetrics(ctx, user, ch)
		c.collectDataAccessMetrics(ctx, user, ch)
		c.collectAccountAssociationMetrics(ctx, user, ch)
		c.collectAuthenticationMetrics(ctx, user, ch)
		c.collectResourceAccessMetrics(ctx, user, ch)
	}

	c.activePermissions.Collect(ch)
	c.inheritedPermissions.Collect(ch)
	c.explicitPermissions.Collect(ch)
	c.deniedPermissions.Collect(ch)
	c.temporaryPermissions.Collect(ch)
	c.conditionalPermissions.Collect(ch)
	c.permissionLevel.Collect(ch)
	c.permissionScope.Collect(ch)
	c.permissionRiskLevel.Collect(ch)
	c.permissionCompliance.Collect(ch)
	c.auditEntriesTotal.Collect(ch)
	c.auditEntryResponseCode.Collect(ch)
	c.auditEntryProcessingTime.Collect(ch)
	c.auditEntryRiskScore.Collect(ch)
	c.auditEntryAnomalyScore.Collect(ch)
	c.dataAccessTotal.Collect(ch)
	c.securityViolationsTotal.Collect(ch)
	c.loginFrequency.Collect(ch)
	c.averageSessionDuration.Collect(ch)
	c.accessAnomalies.Collect(ch)
	c.patternStability.Collect(ch)
	c.usageVariability.Collect(ch)
	c.behaviorRiskIndicators.Collect(ch)
	c.accessEfficiency.Collect(ch)
	c.permissionChangesTotal.Collect(ch)
	c.permissionApprovalsTotal.Collect(ch)
	c.permissionRollbacksTotal.Collect(ch)
	c.permissionChangeRisk.Collect(ch)
	c.permissionChangeImpact.Collect(ch)
	c.accessViolationsTotal.Collect(ch)
	c.violationSeverity.Collect(ch)
	c.violationRiskScore.Collect(ch)
	c.violationRecurrence.Collect(ch)
	c.violationResolutionTime.Collect(ch)
	c.violationImpactScore.Collect(ch)
	c.activeSessions.Collect(ch)
	c.totalSessions.Collect(ch)
	c.maxConcurrentSessions.Collect(ch)
	c.sessionAnomalies.Collect(ch)
	c.suspiciousSessions.Collect(ch)
	c.sessionEfficiency.Collect(ch)
	c.sessionRiskScore.Collect(ch)
	c.complianceScore.Collect(ch)
	c.certificationCompliance.Collect(ch)
	c.trainingCompliance.Collect(ch)
	c.policyCompliance.Collect(ch)
	c.complianceViolations.Collect(ch)
	c.complianceRiskLevel.Collect(ch)
	c.remediationActionsTotal.Collect(ch)
	c.securityEventsTotal.Collect(ch)
	c.securityEventSeverity.Collect(ch)
	c.threatLevel.Collect(ch)
	c.attackVectorCount.Collect(ch)
	c.responseTime.Collect(ch)
	c.resolutionTime.Collect(ch)
	c.businessImpactScore.Collect(ch)
	c.roleAssignmentsTotal.Collect(ch)
	c.activeRoles.Collect(ch)
	c.roleConflicts.Collect(ch)
	c.roleUsageMetrics.Collect(ch)
	c.roleRiskAssessment.Collect(ch)
	c.roleComplianceScore.Collect(ch)
	c.privilegeEscalationsTotal.Collect(ch)
	c.escalationRiskScore.Collect(ch)
	c.escalationDuration.Collect(ch)
	c.abuseIndicators.Collect(ch)
	c.autoRevocations.Collect(ch)
	c.escalationPatterns.Collect(ch)
	c.dataClassificationAccess.Collect(ch)
	c.sensitiveDataAccess.Collect(ch)
	c.dataOperationsTotal.Collect(ch)
	c.dataGovernanceScore.Collect(ch)
	c.dataComplianceStatus.Collect(ch)
	c.dataBreachInvolvement.Collect(ch)
	c.accountAssociationsTotal.Collect(ch)
	c.primaryAccountAssignments.Collect(ch)
	c.accountPermissionLevel.Collect(ch)
	c.accountUsageMetrics.Collect(ch)
	c.accountCostMetrics.Collect(ch)
	c.accountRiskMetrics.Collect(ch)
	c.authenticationEventsTotal.Collect(ch)
	c.authenticationRiskScore.Collect(ch)
	c.authenticationTrustScore.Collect(ch)
	c.authenticationAnomalies.Collect(ch)
	c.multiFactorUsage.Collect(ch)
	c.authenticationDuration.Collect(ch)
	c.tokenLifetime.Collect(ch)
	c.resourceAccessTotal.Collect(ch)
	c.resourceQuotaUtilization.Collect(ch)
	c.resourceEfficiency.Collect(ch)
	c.resourceWaste.Collect(ch)
	c.resourceCosts.Collect(ch)
	c.resourceViolations.Collect(ch)
	c.resourceCompliance.Collect(ch)
	c.resourceSecurity.Collect(ch)
	c.resourcePerformance.Collect(ch)
	c.resourceOptimization.Collect(ch)
}

func (c *UserPermissionAuditCollector) collectPermissionMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	permissions, err := c.client.GetUserPermissions(ctx, user)
	if err != nil {
		log.Printf("Error collecting permissions for user %s: %v", user, err)
		return
	}

	c.activePermissions.WithLabelValues(user).Set(float64(len(permissions.ActivePermissions)))
	c.inheritedPermissions.WithLabelValues(user).Set(float64(len(permissions.InheritedPermissions)))
	c.explicitPermissions.WithLabelValues(user).Set(float64(len(permissions.ExplicitPermissions)))
	c.deniedPermissions.WithLabelValues(user).Set(float64(len(permissions.DeniedPermissions)))
	c.temporaryPermissions.WithLabelValues(user).Set(float64(len(permissions.TemporaryPermissions)))
	c.conditionalPermissions.WithLabelValues(user).Set(float64(len(permissions.ConditionalPermissions)))

	var levelValue float64
	switch permissions.PermissionLevel {
	case "basic":
		levelValue = 0
	case "standard":
		levelValue = 1
	case "advanced":
		levelValue = 2
	case "admin":
		levelValue = 3
	}
	c.permissionLevel.WithLabelValues(user, permissions.PermissionLevel).Set(levelValue)

	var scopeValue float64
	switch permissions.PermissionScope {
	case "limited":
		scopeValue = 0
	case "departmental":
		scopeValue = 1
	case "organizational":
		scopeValue = 2
	case "global":
		scopeValue = 3
	}
	c.permissionScope.WithLabelValues(user, permissions.PermissionScope).Set(scopeValue)

	var riskValue float64
	switch permissions.RiskLevel {
	case "low":
		riskValue = 0
	case "medium":
		riskValue = 1
	case "high":
		riskValue = 2
	case "critical":
		riskValue = 3
	}
	c.permissionRiskLevel.WithLabelValues(user, permissions.RiskLevel).Set(riskValue)

	var complianceScore float64
	switch permissions.ComplianceStatus {
	case "compliant":
		complianceScore = 1.0
	case "warning":
		complianceScore = 0.7
	case "violation":
		complianceScore = 0.3
	default:
		complianceScore = 0.0
	}
	c.permissionCompliance.WithLabelValues(user).Set(complianceScore)
}

func (c *UserPermissionAuditCollector) collectAuditMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	auditEntries, err := c.client.GetUserAccessAuditLog(ctx, user, startTime, endTime)
	if err != nil {
		log.Printf("Error collecting audit log for user %s: %v", user, err)
		return
	}

	accessTypeCounts := make(map[string]map[string]int)
	var totalRiskScore, totalAnomalyScore float64
	var entryCount int

	for _, entry := range auditEntries {
		if accessTypeCounts[entry.AccessType] == nil {
			accessTypeCounts[entry.AccessType] = make(map[string]int)
		}
		accessTypeCounts[entry.AccessType][entry.Result]++

		c.auditEntryResponseCode.WithLabelValues(user).Observe(float64(entry.ResponseCode))
		c.auditEntryProcessingTime.WithLabelValues(user, entry.AccessType).Observe(entry.ProcessingTime.Seconds())

		totalRiskScore += entry.RiskScore
		totalAnomalyScore += entry.AnomalyScore
		entryCount++

		if entry.DataSensitivity != "" {
			c.dataAccessTotal.WithLabelValues(user, entry.DataSensitivity, entry.Action).Inc()
		}

		if len(entry.ComplianceFlags) > 0 {
			c.securityViolationsTotal.WithLabelValues(user, "compliance").Inc()
		}
	}

	for accessType, resultMap := range accessTypeCounts {
		for result, count := range resultMap {
			c.auditEntriesTotal.WithLabelValues(user, accessType, result).Add(float64(count))
		}
	}

	if entryCount > 0 {
		c.auditEntryRiskScore.WithLabelValues(user).Set(totalRiskScore / float64(entryCount))
		c.auditEntryAnomalyScore.WithLabelValues(user).Set(totalAnomalyScore / float64(entryCount))
	}
}

func (c *UserPermissionAuditCollector) collectAccessPatterns(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	patterns, err := c.client.GetUserAccessPatterns(ctx, user)
	if err != nil {
		log.Printf("Error collecting access patterns for user %s: %v", user, err)
		return
	}

	c.loginFrequency.WithLabelValues(user).Set(patterns.LoginFrequency)
	c.averageSessionDuration.WithLabelValues(user).Set(patterns.AverageSessionDuration.Seconds())
	c.accessAnomalies.WithLabelValues(user).Set(float64(patterns.AnomalyCount))
	c.patternStability.WithLabelValues(user).Set(patterns.PatternStability)
	c.usageVariability.WithLabelValues(user).Set(patterns.UsageVariability)
	c.behaviorRiskIndicators.WithLabelValues(user).Set(float64(len(patterns.RiskIndicators)))

	if efficiency, ok := patterns.EfficiencyMetrics["overall"]; ok {
		c.accessEfficiency.WithLabelValues(user).Set(efficiency)
	}
}

func (c *UserPermissionAuditCollector) collectPermissionHistory(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	endTime := time.Now()
	startTime := endTime.Add(-30 * 24 * time.Hour)
	history, err := c.client.GetUserPermissionHistory(ctx, user, startTime, endTime)
	if err != nil {
		log.Printf("Error collecting permission history for user %s: %v", user, err)
		return
	}

	changeCounts := make(map[string]int)
	approvalCounts := make(map[string]int)
	rollbackCount := 0
	var totalRisk, totalImpact float64
	var historyCount int

	for _, entry := range history {
		changeCounts[entry.Action]++
		approvalCounts[entry.ApprovalStatus]++

		if entry.RollbackTime != nil {
			rollbackCount++
		}

		if entry.RiskAssessment != "" {
			var riskScore float64
			switch entry.RiskAssessment {
			case "low":
				riskScore = 0.25
			case "medium":
				riskScore = 0.5
			case "high":
				riskScore = 0.75
			case "critical":
				riskScore = 1.0
			}
			totalRisk += riskScore
		}

		if entry.ImpactAssessment != "" {
			var impactScore float64
			switch entry.ImpactAssessment {
			case "low":
				impactScore = 0.25
			case "medium":
				impactScore = 0.5
			case "high":
				impactScore = 0.75
			case "critical":
				impactScore = 1.0
			}
			totalImpact += impactScore
		}

		historyCount++
	}

	for changeType, count := range changeCounts {
		c.permissionChangesTotal.WithLabelValues(user, changeType).Add(float64(count))
	}

	for status, count := range approvalCounts {
		c.permissionApprovalsTotal.WithLabelValues(user, status).Add(float64(count))
	}

	c.permissionRollbacksTotal.WithLabelValues(user).Add(float64(rollbackCount))

	if historyCount > 0 {
		c.permissionChangeRisk.WithLabelValues(user).Set(totalRisk / float64(historyCount))
		c.permissionChangeImpact.WithLabelValues(user).Set(totalImpact / float64(historyCount))
	}
}

func (c *UserPermissionAuditCollector) collectViolationMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	violations, err := c.client.GetUserAccessViolations(ctx, user)
	if err != nil {
		log.Printf("Error collecting violations for user %s: %v", user, err)
		return
	}

	violationCounts := make(map[string]map[string]int)
	severityCounts := make(map[string]int)
	recurrenceCounts := make(map[string]int)
	var totalRisk, totalImpact float64
	var violationCount int

	for _, violation := range violations {
		if violationCounts[violation.ViolationType] == nil {
			violationCounts[violation.ViolationType] = make(map[string]int)
		}
		violationCounts[violation.ViolationType][violation.Severity]++

		severityCounts[violation.Severity]++
		recurrenceCounts[violation.ViolationType] += violation.RecurrenceCount

		totalRisk += violation.RiskScore

		var impactScore float64
		switch violation.ImpactAssessment {
		case "low":
			impactScore = 0.25
		case "medium":
			impactScore = 0.5
		case "high":
			impactScore = 0.75
		case "critical":
			impactScore = 1.0
		}
		totalImpact += impactScore

		if violation.ResolutionTime != nil {
			resolutionDuration := violation.ResolutionTime.Sub(violation.DetectedTime)
			c.violationResolutionTime.WithLabelValues(user, violation.ViolationType).Observe(resolutionDuration.Seconds())
		}

		violationCount++
	}

	for violationType, severityMap := range violationCounts {
		for severity, count := range severityMap {
			c.accessViolationsTotal.WithLabelValues(user, violationType, severity).Add(float64(count))
		}
	}

	for severity, count := range severityCounts {
		var severityScore float64
		switch severity {
		case "low":
			severityScore = 0.25
		case "medium":
			severityScore = 0.5
		case "high":
			severityScore = 0.75
		case "critical":
			severityScore = 1.0
		}
		c.violationSeverity.WithLabelValues(user, severity).Set(severityScore * float64(count))
	}

	for violationType, count := range recurrenceCounts {
		c.violationRecurrence.WithLabelValues(user, violationType).Set(float64(count))
	}

	if violationCount > 0 {
		c.violationRiskScore.WithLabelValues(user).Set(totalRisk / float64(violationCount))
		c.violationImpactScore.WithLabelValues(user).Set(totalImpact / float64(violationCount))
	}
}

func (c *UserPermissionAuditCollector) collectSessionMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	sessionAudit, err := c.client.GetUserSessionAudit(ctx, user)
	if err != nil {
		log.Printf("Error collecting session audit for user %s: %v", user, err)
		return
	}

	c.activeSessions.WithLabelValues(user).Set(float64(sessionAudit.ActiveSessions))
	c.totalSessions.WithLabelValues(user).Set(float64(sessionAudit.TotalSessions))
	c.maxConcurrentSessions.WithLabelValues(user).Set(float64(sessionAudit.MaxConcurrentSessions))
	c.sessionAnomalies.WithLabelValues(user).Set(float64(sessionAudit.SessionAnomalies))
	c.suspiciousSessions.WithLabelValues(user).Set(float64(sessionAudit.SuspiciousSessions))
	c.sessionEfficiency.WithLabelValues(user).Set(sessionAudit.SessionEfficiency)
	c.sessionRiskScore.WithLabelValues(user).Set(sessionAudit.SessionRiskScore)
}

func (c *UserPermissionAuditCollector) collectComplianceMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	compliance, err := c.client.GetUserComplianceStatus(ctx, user)
	if err != nil {
		log.Printf("Error collecting compliance status for user %s: %v", user, err)
		return
	}

	c.complianceScore.WithLabelValues(user).Set(compliance.OverallComplianceScore)

	certificationCompliance := float64(len(compliance.CompletedCertifications)) / float64(len(compliance.RequiredCertifications)+1) * 100
	c.certificationCompliance.WithLabelValues(user).Set(certificationCompliance)

	trainingCompliance := float64(len(compliance.CompletedTraining)) / float64(len(compliance.TrainingRequirements)+1) * 100
	c.trainingCompliance.WithLabelValues(user).Set(trainingCompliance)

	for policyType, score := range compliance.PolicyCompliance {
		c.policyCompliance.WithLabelValues(user, policyType).Set(score)
	}

	c.complianceViolations.WithLabelValues(user).Set(float64(compliance.ComplianceViolations))

	var riskValue float64
	switch compliance.ComplianceRiskLevel {
	case "low":
		riskValue = 0
	case "medium":
		riskValue = 1
	case "high":
		riskValue = 2
	case "critical":
		riskValue = 3
	}
	c.complianceRiskLevel.WithLabelValues(user, compliance.ComplianceRiskLevel).Set(riskValue)

	c.remediationActionsTotal.WithLabelValues(user, "compliance").Add(float64(len(compliance.RemediationActions)))
}

func (c *UserPermissionAuditCollector) collectSecurityEvents(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	events, err := c.client.GetUserSecurityEvents(ctx, user)
	if err != nil {
		log.Printf("Error collecting security events for user %s: %v", user, err)
		return
	}

	eventCounts := make(map[string]map[string]int)
	severityCounts := make(map[string]int)
	threatCounts := make(map[string]int)
	attackVectorCounts := make(map[string]int)
	var totalBusinessImpact float64
	var eventCount int

	for _, event := range events {
		if eventCounts[event.EventType] == nil {
			eventCounts[event.EventType] = make(map[string]int)
		}
		eventCounts[event.EventType][event.Severity]++

		severityCounts[event.Severity]++
		threatCounts[event.ThreatLevel]++
		if event.AttackVector != "" {
			attackVectorCounts[event.AttackVector]++
		}

		var businessImpact float64
		switch event.BusinessImpact {
		case "low":
			businessImpact = 0.25
		case "medium":
			businessImpact = 0.5
		case "high":
			businessImpact = 0.75
		case "critical":
			businessImpact = 1.0
		}
		totalBusinessImpact += businessImpact

		if event.ResponseTime > 0 {
			c.responseTime.WithLabelValues(user, event.EventType).Observe(event.ResponseTime.Seconds())
		}

		if event.ResolutionTime > 0 {
			c.resolutionTime.WithLabelValues(user, event.EventType).Observe(event.ResolutionTime.Seconds())
		}

		eventCount++
	}

	for eventType, severityMap := range eventCounts {
		for severity, count := range severityMap {
			c.securityEventsTotal.WithLabelValues(user, eventType, severity).Add(float64(count))
		}
	}

	for severity, count := range severityCounts {
		var severityScore float64
		switch severity {
		case "low":
			severityScore = 0.25
		case "medium":
			severityScore = 0.5
		case "high":
			severityScore = 0.75
		case "critical":
			severityScore = 1.0
		}
		c.securityEventSeverity.WithLabelValues(user, severity).Set(severityScore * float64(count))
	}

	for threatLevel, count := range threatCounts {
		var threatValue float64
		switch threatLevel {
		case "low":
			threatValue = 0
		case "medium":
			threatValue = 1
		case "high":
			threatValue = 2
		case "critical":
			threatValue = 3
		}
		c.threatLevel.WithLabelValues(user, threatLevel).Set(threatValue * float64(count))
	}

	for attackVector, count := range attackVectorCounts {
		c.attackVectorCount.WithLabelValues(user, attackVector).Add(float64(count))
	}

	if eventCount > 0 {
		c.businessImpactScore.WithLabelValues(user).Set(totalBusinessImpact / float64(eventCount))
	}
}

func (c *UserPermissionAuditCollector) collectRoleMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	roles, err := c.client.GetUserRoleAssignments(ctx, user)
	if err != nil {
		log.Printf("Error collecting role assignments for user %s: %v", user, err)
		return
	}

	assignmentCounts := make(map[string]map[string]int)
	activeRoleCount := 0
	conflictCount := 0
	var totalUsage, totalRisk, totalCompliance float64
	var roleCount int

	for _, role := range roles {
		if assignmentCounts[role.RoleType] == nil {
			assignmentCounts[role.RoleType] = make(map[string]int)
		}
		assignmentCounts[role.RoleType][role.AssignmentType]++

		if role.AssignmentStatus == "active" {
			activeRoleCount++
		}

		conflictCount += len(role.ConflictingRoles)

		if usage, ok := role.UsageMetrics["overall"]; ok {
			totalUsage += usage
			c.roleUsageMetrics.WithLabelValues(user, role.RoleName).Set(usage)
		}

		var riskScore float64
		switch role.RiskAssessment {
		case "low":
			riskScore = 0.25
		case "medium":
			riskScore = 0.5
		case "high":
			riskScore = 0.75
		case "critical":
			riskScore = 1.0
		}
		totalRisk += riskScore
		c.roleRiskAssessment.WithLabelValues(user, role.RoleName).Set(riskScore)

		var complianceScore float64
		switch role.ComplianceImpact {
		case "compliant":
			complianceScore = 1.0
		case "warning":
			complianceScore = 0.7
		case "violation":
			complianceScore = 0.3
		default:
			complianceScore = 0.0
		}
		totalCompliance += complianceScore
		c.roleComplianceScore.WithLabelValues(user, role.RoleName).Set(complianceScore)

		roleCount++
	}

	for roleType, assignmentMap := range assignmentCounts {
		for assignmentType, count := range assignmentMap {
			c.roleAssignmentsTotal.WithLabelValues(user, roleType, assignmentType).Add(float64(count))
		}
	}

	c.activeRoles.WithLabelValues(user).Set(float64(activeRoleCount))
	c.roleConflicts.WithLabelValues(user).Set(float64(conflictCount))
}

func (c *UserPermissionAuditCollector) collectEscalationMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	escalations, err := c.client.GetUserPrivilegeEscalations(ctx, user)
	if err != nil {
		log.Printf("Error collecting privilege escalations for user %s: %v", user, err)
		return
	}

	escalationCounts := make(map[string]int)
	autoRevocationCount := 0
	var totalRisk, totalPattern float64
	var escalationCount int

	for _, escalation := range escalations {
		escalationCounts[escalation.EscalationType]++

		if escalation.AutoRevocation {
			autoRevocationCount++
		}

		var riskScore float64
		switch escalation.RiskAssessment {
		case "low":
			riskScore = 0.25
		case "medium":
			riskScore = 0.5
		case "high":
			riskScore = 0.75
		case "critical":
			riskScore = 1.0
		}
		totalRisk += riskScore

		var patternScore float64
		switch escalation.EscalationPattern {
		case "normal":
			patternScore = 0.0
		case "suspicious":
			patternScore = 0.5
		case "anomalous":
			patternScore = 1.0
		}
		totalPattern += patternScore

		if escalation.ActualDuration > 0 {
			c.escalationDuration.WithLabelValues(user, escalation.EscalationType).Observe(escalation.ActualDuration.Seconds())
		}

		escalationCount++
	}

	for escalationType, count := range escalationCounts {
		c.privilegeEscalationsTotal.WithLabelValues(user, escalationType).Add(float64(count))
	}

	c.autoRevocations.WithLabelValues(user).Add(float64(autoRevocationCount))

	if escalationCount > 0 {
		c.escalationRiskScore.WithLabelValues(user).Set(totalRisk / float64(escalationCount))
		c.escalationPatterns.WithLabelValues(user).Set(totalPattern / float64(escalationCount))
	}

	c.abuseIndicators.WithLabelValues(user).Set(0) // Would need abuse indicator count from escalations
}

func (c *UserPermissionAuditCollector) collectDataAccessMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	dataAccess, err := c.client.GetUserDataAccess(ctx, user)
	if err != nil {
		log.Printf("Error collecting data access for user %s: %v", user, err)
		return
	}

	for _, classification := range dataAccess.DataClassifications {
		c.dataClassificationAccess.WithLabelValues(user, classification).Set(1)
	}

	sensitiveDataTypes := map[string]bool{
		"pii":                   dataAccess.PIIAccess,
		"phi":                   dataAccess.PHIAccess,
		"financial":             dataAccess.FinancialDataAccess,
		"intellectual_property": dataAccess.IntellectualPropertyAccess,
		"classified":            dataAccess.ClassifiedDataAccess,
	}

	for dataType, hasAccess := range sensitiveDataTypes {
		var accessValue float64
		if hasAccess {
			accessValue = 1
		}
		c.sensitiveDataAccess.WithLabelValues(user, dataType).Set(accessValue)
	}

	operationCounts := map[string]int64{
		"download":     dataAccess.DataDownloadCount,
		"upload":       dataAccess.DataUploadCount,
		"view":         dataAccess.DataViewCount,
		"modification": dataAccess.DataModificationCount,
		"deletion":     dataAccess.DataDeletionCount,
		"sharing":      dataAccess.DataSharingCount,
		"export":       dataAccess.DataExportCount,
	}

	for operation, count := range operationCounts {
		c.dataOperationsTotal.WithLabelValues(user, operation).Add(float64(count))
	}

	c.dataGovernanceScore.WithLabelValues(user).Set(dataAccess.DataGovernanceScore)

	var complianceValue float64
	switch dataAccess.DataComplianceStatus {
	case "compliant":
		complianceValue = 1.0
	case "warning":
		complianceValue = 0.7
	case "violation":
		complianceValue = 0.3
	default:
		complianceValue = 0.0
	}
	c.dataComplianceStatus.WithLabelValues(user, dataAccess.DataComplianceStatus).Set(complianceValue)

	c.dataBreachInvolvement.WithLabelValues(user).Set(float64(len(dataAccess.DataBreachInvolvement)))
}

func (c *UserPermissionAuditCollector) collectAccountAssociationMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	associations, err := c.client.GetUserAccountAssociations(ctx, user)
	if err != nil {
		log.Printf("Error collecting account associations for user %s: %v", user, err)
		return
	}

	associationCounts := make(map[string]int)
	primaryCount := 0

	for _, association := range associations {
		// Count by access level instead of non-existent AssociationType
		associationCounts[association.AccessLevel]++

		// Check if this is the primary account
		if association.PrimaryAccount != "" {
			primaryCount++
		}

		var levelValue float64
		switch association.AccessLevel {
		case "read":
			levelValue = 1
		case "write":
			levelValue = 2
		case "admin":
			levelValue = 3
		}
		// Use PrimaryAccount as the account name
		c.accountPermissionLevel.WithLabelValues(user, association.PrimaryAccount).Set(levelValue)

		// Check account usage if available
		if association.AccountUsage != nil && association.AccountUsage[association.PrimaryAccount] != nil {
			usage := association.AccountUsage[association.PrimaryAccount].UsagePercentage
			c.accountUsageMetrics.WithLabelValues(user, association.PrimaryAccount).Set(usage)
		}

		// Skip cost and risk metrics as they don't exist in the struct
	}

	for associationType, count := range associationCounts {
		c.accountAssociationsTotal.WithLabelValues(user, associationType).Add(float64(count))
	}

	c.primaryAccountAssignments.WithLabelValues(user).Set(float64(primaryCount))
}

func (c *UserPermissionAuditCollector) collectAuthenticationMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	events, err := c.client.GetUserAuthenticationEvents(ctx, user)
	if err != nil {
		log.Printf("Error collecting authentication events for user %s: %v", user, err)
		return
	}

	authCounts := make(map[string]map[string]int)
	var totalRisk, totalTrust, totalAnomaly float64
	mfaUsageCount := 0
	var eventCount int

	for _, event := range events {
		if authCounts[event.AuthenticationType] == nil {
			authCounts[event.AuthenticationType] = make(map[string]int)
		}
		authCounts[event.AuthenticationType][event.AuthenticationResult]++

		totalRisk += event.RiskScore
		totalTrust += event.TrustScore
		totalAnomaly += event.AnomalyScore

		if event.MultiFactorUsed {
			mfaUsageCount++
		}

		c.authenticationDuration.WithLabelValues(user, event.AuthenticationType).Observe(event.AuthenticationDuration.Seconds())
		c.tokenLifetime.WithLabelValues(user).Observe(event.TokenLifetime.Seconds())

		eventCount++
	}

	for authType, resultMap := range authCounts {
		for result, count := range resultMap {
			c.authenticationEventsTotal.WithLabelValues(user, authType, result).Add(float64(count))
		}
	}

	if eventCount > 0 {
		c.authenticationRiskScore.WithLabelValues(user).Set(totalRisk / float64(eventCount))
		c.authenticationTrustScore.WithLabelValues(user).Set(totalTrust / float64(eventCount))
		c.authenticationAnomalies.WithLabelValues(user).Set(totalAnomaly / float64(eventCount))
		c.multiFactorUsage.WithLabelValues(user).Set(float64(mfaUsageCount) / float64(eventCount))
	}
}

func (c *UserPermissionAuditCollector) collectResourceAccessMetrics(ctx context.Context, user string, ch chan<- prometheus.Metric) {
	resourceAccess, err := c.client.GetUserResourceAccess(ctx, user)
	if err != nil {
		log.Printf("Error collecting resource access for user %s: %v", user, err)
		return
	}

	for resourceType, frequency := range resourceAccess.ResourceAccessFrequency {
		c.resourceAccessTotal.WithLabelValues(user, resourceType).Add(float64(frequency))
	}

	for resourceType, utilization := range resourceAccess.QuotaUtilization {
		c.resourceQuotaUtilization.WithLabelValues(user, resourceType).Set(utilization)
	}

	for resourceType, efficiency := range resourceAccess.ResourceEfficiency {
		c.resourceEfficiency.WithLabelValues(user, resourceType).Set(efficiency)
	}

	for resourceType, waste := range resourceAccess.ResourceWaste {
		c.resourceWaste.WithLabelValues(user, resourceType).Set(waste)
	}

	for resourceType, cost := range resourceAccess.ResourceCosts {
		c.resourceCosts.WithLabelValues(user, resourceType).Set(cost)
	}

	for resourceType, violations := range resourceAccess.ResourceViolations {
		c.resourceViolations.WithLabelValues(user, resourceType).Add(float64(violations))
	}

	for resourceType, compliance := range resourceAccess.ResourceCompliance {
		c.resourceCompliance.WithLabelValues(user, resourceType).Set(compliance)
	}

	for resourceType, security := range resourceAccess.ResourceSecurity {
		c.resourceSecurity.WithLabelValues(user, resourceType).Set(security)
	}

	for resourceType, performance := range resourceAccess.ResourcePerformance {
		c.resourcePerformance.WithLabelValues(user, resourceType).Set(performance)
	}

	for resourceType, optimization := range resourceAccess.ResourceOptimization {
		c.resourceOptimization.WithLabelValues(user, resourceType).Set(optimization)
	}
}
