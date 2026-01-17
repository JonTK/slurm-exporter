package collector

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// FairShareViolationsCollector collects fair-share violation detection and alerting metrics
type FairShareViolationsCollector struct {
	client FairShareViolationsSLURMClient

	// Violation Detection Metrics
	fairShareViolations *prometheus.GaugeVec
	violationSeverity   *prometheus.GaugeVec
	violationDuration   *prometheus.GaugeVec
	violationType       *prometheus.GaugeVec
	violationImpact     *prometheus.GaugeVec

	// Violation Trends
	violationTrends     *prometheus.GaugeVec
	violationFrequency  *prometheus.CounterVec
	violationEscalation *prometheus.CounterVec
	violationResolution *prometheus.CounterVec
	violationRecurrence *prometheus.GaugeVec

	// User Violation Metrics
	userViolationScore   *prometheus.GaugeVec
	userViolationHistory *prometheus.GaugeVec
	userViolationPattern *prometheus.GaugeVec
	userComplianceScore  *prometheus.GaugeVec
	userRiskAssessment   *prometheus.GaugeVec

	// Account Violation Metrics
	accountViolationAggregate *prometheus.GaugeVec
	accountViolationRisk      *prometheus.GaugeVec
	accountComplianceMetrics  *prometheus.GaugeVec
	accountViolationImpact    *prometheus.GaugeVec

	// System Violation Metrics
	systemViolationOverview   *prometheus.GaugeVec
	systemViolationThresholds *prometheus.GaugeVec
	systemComplianceHealth    *prometheus.GaugeVec
	systemViolationPrediction *prometheus.GaugeVec

	// Alert Management Metrics
	alertsGenerated       *prometheus.CounterVec
	alertsAcknowledged    *prometheus.CounterVec
	alertsResolved        *prometheus.CounterVec
	alertResponseTime     *prometheus.HistogramVec
	alertEscalationEvents *prometheus.CounterVec

	// Notification Metrics
	notificationsSent          *prometheus.CounterVec
	notificationDeliveryStatus *prometheus.GaugeVec
	notificationChannels       *prometheus.GaugeVec
	notificationPreferences    *prometheus.GaugeVec

	// Remediation Tracking
	remediationActions       *prometheus.CounterVec
	remediationEffectiveness *prometheus.GaugeVec
	remediationTime          *prometheus.HistogramVec
	automatedRemediation     *prometheus.CounterVec
	remediationSuccess       *prometheus.GaugeVec

	// Policy Enforcement
	policyViolations        *prometheus.CounterVec
	policyEffectiveness     *prometheus.GaugeVec
	policyComplianceRate    *prometheus.GaugeVec
	policyUpdateFrequency   *prometheus.CounterVec
	policyValidationResults *prometheus.GaugeVec

	// Threshold Management
	thresholdBreaches          *prometheus.CounterVec
	thresholdOptimality        *prometheus.GaugeVec
	thresholdCalibration       *prometheus.GaugeVec
	dynamicThresholdAdjustment *prometheus.CounterVec
	thresholdEffectiveness     *prometheus.GaugeVec
}

// FairShareViolationsSLURMClient interface for fair-share violation operations
type FairShareViolationsSLURMClient interface {
	DetectFairShareViolations(ctx context.Context) ([]*FairShareViolationRecord, error)
	GetUserViolationProfile(ctx context.Context, userName string) (*UserViolationProfile, error)
	GetAccountViolationSummary(ctx context.Context, accountName string) (*AccountViolationSummary, error)
	GetSystemViolationOverview(ctx context.Context) (*SystemViolationOverview, error)
	GetViolationTrendAnalysis(ctx context.Context, period string) (*ViolationTrendAnalysis, error)
	GetAlertHistory(ctx context.Context, period string) (*AlertHistory, error)
	GetRemediationHistory(ctx context.Context, period string) (*RemediationHistory, error)
	ValidateViolationThresholds(ctx context.Context) (*ThresholdValidation, error)
	GetNotificationStats(ctx context.Context) (*NotificationStats, error)
	GetPolicyComplianceReport(ctx context.Context) (*PolicyComplianceReport, error)
	PredictViolationRisk(ctx context.Context, scope string) (*ViolationRiskPrediction, error)
	GetViolationImpactAnalysis(ctx context.Context, violationID string) (*ViolationImpactAnalysis, error)
}

// FairShareViolationRecord represents a detected violation
type FairShareViolationRecord struct {
	ViolationID   string `json:"violation_id"`
	UserName      string `json:"user_name"`
	AccountName   string `json:"account_name"`
	ViolationType string `json:"violation_type"`
	Severity      string `json:"severity"`
	Description   string `json:"description"`

	// Violation Details
	CurrentFairShare    float64 `json:"current_fair_share"`
	ExpectedFairShare   float64 `json:"expected_fair_share"`
	ViolationMagnitude  float64 `json:"violation_magnitude"`
	ViolationPercentage float64 `json:"violation_percentage"`
	ThresholdExceeded   string  `json:"threshold_exceeded"`

	// Context Information
	ResourcesUsed  map[string]float64 `json:"resources_used"`
	JobsAffected   []string           `json:"jobs_affected"`
	TimeframeStart time.Time          `json:"timeframe_start"`
	TimeframeEnd   time.Time          `json:"timeframe_end"`
	Duration       time.Duration      `json:"duration"`

	// Impact Assessment
	PriorityImpact float64 `json:"priority_impact"`
	QueueImpact    float64 `json:"queue_impact"`
	SystemImpact   float64 `json:"system_impact"`
	UserImpact     float64 `json:"user_impact"`
	OverallImpact  float64 `json:"overall_impact"`

	// Status and Resolution
	Status         string    `json:"status"`
	DetectedAt     time.Time `json:"detected_at"`
	AlertSent      bool      `json:"alert_sent"`
	Acknowledged   bool      `json:"acknowledged"`
	AcknowledgedBy string    `json:"acknowledged_by"`
	AcknowledgedAt time.Time `json:"acknowledged_at"`
	ResolvedAt     time.Time `json:"resolved_at"`
	Resolution     string    `json:"resolution"`

	// Remediation
	RemediationActions []string `json:"remediation_actions"`
	AutoRemediation    bool     `json:"auto_remediation"`
	RemediationStatus  string   `json:"remediation_status"`
}

// UserViolationProfile represents user-specific violation profile
type UserViolationProfile struct {
	UserName    string `json:"user_name"`
	AccountName string `json:"account_name"`

	// Violation Statistics
	TotalViolations    int     `json:"total_violations"`
	ActiveViolations   int     `json:"active_violations"`
	ResolvedViolations int     `json:"resolved_violations"`
	ViolationRate      float64 `json:"violation_rate"`
	ViolationScore     float64 `json:"violation_score"`

	// Violation Patterns
	ViolationTypes     map[string]int `json:"violation_types"`
	MostCommonType     string         `json:"most_common_type"`
	ViolationFrequency float64        `json:"violation_frequency"`
	RecurrenceRate     float64        `json:"recurrence_rate"`
	EscalationRate     float64        `json:"escalation_rate"`

	// Compliance Metrics
	ComplianceScore  float64 `json:"compliance_score"`
	ComplianceRating string  `json:"compliance_rating"`
	ComplianceTrend  string  `json:"compliance_trend"`
	RiskLevel        string  `json:"risk_level"`
	RiskScore        float64 `json:"risk_score"`

	// Behavioral Analysis
	ViolationPatterns []string `json:"violation_patterns"`
	BehaviorChanges   []string `json:"behavior_changes"`
	ImprovementTrend  string   `json:"improvement_trend"`
	AdaptationScore   float64  `json:"adaptation_score"`

	// Response Metrics
	AverageResponseTime float64 `json:"average_response_time"`
	RemediationSuccess  float64 `json:"remediation_success"`
	CooperationLevel    float64 `json:"cooperation_level"`

	LastUpdated time.Time `json:"last_updated"`
}

// AccountViolationSummary represents account-level violation summary
type AccountViolationSummary struct {
	AccountName string `json:"account_name"`

	// Aggregate Statistics
	TotalUsers          int     `json:"total_users"`
	UsersWithViolations int     `json:"users_with_violations"`
	TotalViolations     int     `json:"total_violations"`
	ActiveViolations    int     `json:"active_violations"`
	ViolationRate       float64 `json:"violation_rate"`

	// Risk Assessment
	AccountRiskScore float64  `json:"account_risk_score"`
	RiskLevel        string   `json:"risk_level"`
	RiskFactors      []string `json:"risk_factors"`
	RiskTrend        string   `json:"risk_trend"`

	// Compliance Metrics
	ComplianceScore  float64 `json:"compliance_score"`
	ComplianceRating string  `json:"compliance_rating"`
	PolicyAdherence  float64 `json:"policy_adherence"`
	ViolationImpact  float64 `json:"violation_impact"`

	// Management Metrics
	ManagementResponse      float64 `json:"management_response"`
	RemediationRate         float64 `json:"remediation_rate"`
	PreventionEffectiveness float64 `json:"prevention_effectiveness"`

	LastSummarized time.Time `json:"last_summarized"`
}

// SystemViolationOverview represents system-wide violation overview
type SystemViolationOverview struct {
	// Overall Statistics
	TotalViolations    int     `json:"total_violations"`
	ActiveViolations   int     `json:"active_violations"`
	ViolationsToday    int     `json:"violations_today"`
	ViolationsThisWeek int     `json:"violations_this_week"`
	ViolationRate      float64 `json:"violation_rate"`

	// Severity Breakdown
	CriticalViolations int `json:"critical_violations"`
	HighViolations     int `json:"high_violations"`
	MediumViolations   int `json:"medium_violations"`
	LowViolations      int `json:"low_violations"`

	// System Health
	ComplianceHealth  float64 `json:"compliance_health"`
	ViolationTrend    string  `json:"violation_trend"`
	SystemStressLevel float64 `json:"system_stress_level"`
	AlertLoad         float64 `json:"alert_load"`

	// Thresholds
	ThresholdBreaches      int     `json:"threshold_breaches"`
	ThresholdAdjustments   int     `json:"threshold_adjustments"`
	ThresholdEffectiveness float64 `json:"threshold_effectiveness"`

	// Performance Metrics
	DetectionAccuracy  float64 `json:"detection_accuracy"`
	FalsePositiveRate  float64 `json:"false_positive_rate"`
	ResponseEfficiency float64 `json:"response_efficiency"`
	AutomationRate     float64 `json:"automation_rate"`

	LastUpdated time.Time `json:"last_updated"`
}

// ViolationTrendAnalysis represents violation trend analysis
type ViolationTrendAnalysis struct {
	AnalysisPeriod string `json:"analysis_period"`

	// Trend Metrics
	ViolationTrend    string  `json:"violation_trend"`
	TrendDirection    string  `json:"trend_direction"`
	TrendVelocity     float64 `json:"trend_velocity"`
	TrendAcceleration float64 `json:"trend_acceleration"`
	TrendVolatility   float64 `json:"trend_volatility"`

	// Pattern Analysis
	PatternType      string             `json:"pattern_type"`
	SeasonalPatterns map[string]float64 `json:"seasonal_patterns"`
	CyclicalBehavior bool               `json:"cyclical_behavior"`
	AnomalousEvents  []string           `json:"anomalous_events"`

	// Predictions
	FutureTrend          string  `json:"future_trend"`
	PredictedViolations  int     `json:"predicted_violations"`
	PredictionConfidence float64 `json:"prediction_confidence"`
	RiskForecast         string  `json:"risk_forecast"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// AlertHistory represents alert history and statistics
type AlertHistory struct {
	Period string `json:"period"`

	// Alert Statistics
	TotalAlerts        int `json:"total_alerts"`
	AlertsSent         int `json:"alerts_sent"`
	AlertsAcknowledged int `json:"alerts_acknowledged"`
	AlertsResolved     int `json:"alerts_resolved"`
	AlertsEscalated    int `json:"alerts_escalated"`

	// Response Metrics
	AverageResponseTime float64 `json:"average_response_time"`
	AcknowledgmentRate  float64 `json:"acknowledgment_rate"`
	ResolutionRate      float64 `json:"resolution_rate"`
	EscalationRate      float64 `json:"escalation_rate"`

	// Channel Statistics
	EmailAlerts     int     `json:"email_alerts"`
	SlackAlerts     int     `json:"slack_alerts"`
	SMSAlerts       int     `json:"sms_alerts"`
	PagerDutyAlerts int     `json:"pager_duty_alerts"`
	DeliverySuccess float64 `json:"delivery_success"`

	LastUpdated time.Time `json:"last_updated"`
}

// RemediationHistory represents remediation history and effectiveness
type RemediationHistory struct {
	Period string `json:"period"`

	// Remediation Statistics
	TotalRemediations      int `json:"total_remediations"`
	SuccessfulRemediations int `json:"successful_remediations"`
	FailedRemediations     int `json:"failed_remediations"`
	AutoRemediations       int `json:"auto_remediations"`
	ManualRemediations     int `json:"manual_remediations"`

	// Effectiveness Metrics
	RemediationSuccess      float64 `json:"remediation_success"`
	AverageRemediationTime  float64 `json:"average_remediation_time"`
	AutomationEffectiveness float64 `json:"automation_effectiveness"`
	RecurrencePrevention    float64 `json:"recurrence_prevention"`

	// Action Types
	ActionTypes          map[string]int `json:"action_types"`
	MostEffectiveAction  string         `json:"most_effective_action"`
	LeastEffectiveAction string         `json:"least_effective_action"`

	LastUpdated time.Time `json:"last_updated"`
}

// ThresholdValidation represents threshold validation results
type ThresholdValidation struct {
	// Validation Results
	ValidationStatus  string  `json:"validation_status"`
	ValidationScore   float64 `json:"validation_score"`
	ThresholdAccuracy float64 `json:"threshold_accuracy"`
	OptimalityScore   float64 `json:"optimality_score"`

	// Threshold Analysis
	OptimalThresholds  map[string]float64 `json:"optimal_thresholds"`
	CurrentThresholds  map[string]float64 `json:"current_thresholds"`
	RecommendedChanges []string           `json:"recommended_changes"`
	CalibrationNeeded  bool               `json:"calibration_needed"`

	// Performance Metrics
	FalsePositiveRate float64 `json:"false_positive_rate"`
	FalseNegativeRate float64 `json:"false_negative_rate"`
	DetectionAccuracy float64 `json:"detection_accuracy"`
	Sensitivity       float64 `json:"sensitivity"`
	Specificity       float64 `json:"specificity"`

	ValidatedAt time.Time `json:"validated_at"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	// Delivery Statistics
	TotalNotifications   int     `json:"total_notifications"`
	SuccessfulDeliveries int     `json:"successful_deliveries"`
	FailedDeliveries     int     `json:"failed_deliveries"`
	DeliveryRate         float64 `json:"delivery_rate"`

	// Channel Performance
	ChannelStats         map[string]*ChannelStats `json:"channel_stats"`
	PreferredChannels    []string                 `json:"preferred_channels"`
	ChannelEffectiveness map[string]float64       `json:"channel_effectiveness"`

	// User Preferences
	UserPreferences      map[string]*UserNotificationPrefs `json:"user_preferences"`
	PreferenceCompliance float64                           `json:"preference_compliance"`

	LastUpdated time.Time `json:"last_updated"`
}

// ChannelStats represents per-channel statistics
type ChannelStats struct {
	ChannelName     string  `json:"channel_name"`
	MessagesSent    int     `json:"messages_sent"`
	DeliverySuccess float64 `json:"delivery_success"`
	ResponseRate    float64 `json:"response_rate"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// UserNotificationPrefs represents user notification preferences
type UserNotificationPrefs struct {
	UserName          string   `json:"user_name"`
	PreferredChannels []string `json:"preferred_channels"`
	Severity          string   `json:"severity"`
	TimeWindows       []string `json:"time_windows"`
	Enabled           bool     `json:"enabled"`
}

// PolicyComplianceReport represents policy compliance report
type PolicyComplianceReport struct {
	// Overall Compliance
	OverallCompliance float64 `json:"overall_compliance"`
	ComplianceRating  string  `json:"compliance_rating"`
	ComplianceTrend   string  `json:"compliance_trend"`

	// Policy Metrics
	TotalPolicies       int     `json:"total_policies"`
	ActivePolicies      int     `json:"active_policies"`
	ViolatedPolicies    int     `json:"violated_policies"`
	PolicyEffectiveness float64 `json:"policy_effectiveness"`

	// Violation Breakdown
	PolicyViolations    map[string]int `json:"policy_violations"`
	ViolationsByUser    map[string]int `json:"violations_by_user"`
	ViolationsByAccount map[string]int `json:"violations_by_account"`

	// Enforcement Metrics
	EnforcementRate   float64 `json:"enforcement_rate"`
	AutoEnforcement   float64 `json:"auto_enforcement"`
	ManualEnforcement float64 `json:"manual_enforcement"`

	LastGenerated time.Time `json:"last_generated"`
}

// ViolationRiskPrediction represents violation risk prediction
type ViolationRiskPrediction struct {
	Scope             string `json:"scope"`
	PredictionHorizon string `json:"prediction_horizon"`

	// Risk Metrics
	RiskScore           float64  `json:"risk_score"`
	RiskLevel           string   `json:"risk_level"`
	PredictedViolations int      `json:"predicted_violations"`
	HighRiskUsers       []string `json:"high_risk_users"`
	HighRiskAccounts    []string `json:"high_risk_accounts"`

	// Risk Factors
	PrimaryRiskFactors    []string `json:"primary_risk_factors"`
	SecondaryRiskFactors  []string `json:"secondary_risk_factors"`
	RiskMitigationActions []string `json:"risk_mitigation_actions"`

	// Confidence Metrics
	PredictionConfidence float64 `json:"prediction_confidence"`
	ModelAccuracy        float64 `json:"model_accuracy"`
	UncertaintyRange     float64 `json:"uncertainty_range"`

	PredictedAt time.Time `json:"predicted_at"`
}

// ViolationImpactAnalysis represents violation impact analysis
type ViolationImpactAnalysis struct {
	ViolationID string `json:"violation_id"`

	// Direct Impact
	DirectImpact   float64 `json:"direct_impact"`
	IndirectImpact float64 `json:"indirect_impact"`
	TotalImpact    float64 `json:"total_impact"`
	ImpactRadius   float64 `json:"impact_radius"`

	// Resource Impact
	CPUImpact     float64 `json:"cpu_impact"`
	MemoryImpact  float64 `json:"memory_impact"`
	StorageImpact float64 `json:"storage_impact"`
	NetworkImpact float64 `json:"network_impact"`

	// User Impact
	AffectedUsers    []string           `json:"affected_users"`
	UserImpactScores map[string]float64 `json:"user_impact_scores"`
	JobsDelayed      int                `json:"jobs_delayed"`
	WaitTimeIncrease float64            `json:"wait_time_increase"`

	// System Impact
	SystemPerformance float64 `json:"system_performance"`
	FairnessReduction float64 `json:"fairness_reduction"`
	EfficiencyLoss    float64 `json:"efficiency_loss"`

	AnalyzedAt time.Time `json:"analyzed_at"`
}

// NewFairShareViolationsCollector creates a new fair-share violations collector
func NewFairShareViolationsCollector(client FairShareViolationsSLURMClient) *FairShareViolationsCollector {
	return &FairShareViolationsCollector{
		client: client,

		// Violation Detection Metrics
		fairShareViolations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violations",
				Help: "Number of fair-share violations by type and severity",
			},
			[]string{"user", "account", "violation_type", "severity", "status"},
		),

		violationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_severity_score",
				Help: "Severity score of fair-share violations",
			},
			[]string{"violation_id", "user", "account", "violation_type"},
		),

		violationDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_duration_minutes",
				Help: "Duration of fair-share violations in minutes",
			},
			[]string{"violation_id", "user", "account", "violation_type"},
		),

		violationType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_type_count",
				Help: "Count of violations by type",
			},
			[]string{"violation_type", "severity"},
		),

		violationImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_impact_score",
				Help: "Impact score of fair-share violations",
			},
			[]string{"violation_id", "user", "account", "impact_type"},
		),

		// Violation Trends
		violationTrends: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_trends",
				Help: "Fair-share violation trend analysis",
			},
			[]string{"trend_type", "period", "direction"},
		),

		violationFrequency: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_fairshare_violation_frequency_total",
				Help: "Total frequency of fair-share violations",
			},
			[]string{"user", "account", "violation_type", "time_period"},
		),

		violationEscalation: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_fairshare_violation_escalations_total",
				Help: "Total number of violation escalations",
			},
			[]string{"violation_type", "from_severity", "to_severity"},
		),

		violationResolution: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_fairshare_violation_resolutions_total",
				Help: "Total number of violation resolutions",
			},
			[]string{"violation_type", "resolution_type", "success"},
		),

		violationRecurrence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violation_recurrence_rate",
				Help: "Recurrence rate of fair-share violations",
			},
			[]string{"user", "account", "violation_type"},
		),

		// User Violation Metrics
		userViolationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_score",
				Help: "User violation score and risk assessment",
			},
			[]string{"user", "account", "score_type"},
		),

		userViolationHistory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_history",
				Help: "User violation history metrics",
			},
			[]string{"user", "account", "metric", "period"},
		),

		userViolationPattern: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_violation_pattern",
				Help: "User violation behavior patterns",
			},
			[]string{"user", "account", "pattern_type"},
		),

		userComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_compliance_score",
				Help: "User compliance score and rating",
			},
			[]string{"user", "account", "compliance_type"},
		),

		userRiskAssessment: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_risk_assessment",
				Help: "User risk assessment for violations",
			},
			[]string{"user", "account", "risk_type"},
		),

		// Account Violation Metrics
		accountViolationAggregate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_violation_aggregate",
				Help: "Account-level violation aggregates",
			},
			[]string{"account", "metric"},
		),

		accountViolationRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_violation_risk",
				Help: "Account violation risk assessment",
			},
			[]string{"account", "risk_type"},
		),

		accountComplianceMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_compliance_metrics",
				Help: "Account compliance metrics and scores",
			},
			[]string{"account", "compliance_metric"},
		),

		accountViolationImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_violation_impact",
				Help: "Account violation impact assessment",
			},
			[]string{"account", "impact_type"},
		),

		// System Violation Metrics
		systemViolationOverview: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_violation_overview",
				Help: "System-wide violation overview metrics",
			},
			[]string{"metric"},
		),

		systemViolationThresholds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_violation_thresholds",
				Help: "System violation threshold metrics",
			},
			[]string{"threshold_type", "metric"},
		),

		systemComplianceHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_compliance_health",
				Help: "System compliance health metrics",
			},
			[]string{"health_metric"},
		),

		systemViolationPrediction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_violation_prediction",
				Help: "System violation prediction metrics",
			},
			[]string{"prediction_type", "horizon"},
		),

		// Alert Management Metrics
		alertsGenerated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_alerts_generated_total",
				Help: "Total number of violation alerts generated",
			},
			[]string{"alert_type", "severity", "channel"},
		),

		alertsAcknowledged: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_alerts_acknowledged_total",
				Help: "Total number of violation alerts acknowledged",
			},
			[]string{"alert_type", "severity", "acknowledged_by"},
		),

		alertsResolved: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_alerts_resolved_total",
				Help: "Total number of violation alerts resolved",
			},
			[]string{"alert_type", "resolution_type"},
		),

		alertResponseTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_violation_alert_response_time_seconds",
				Help:    "Response time for violation alerts",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400},
			},
			[]string{"alert_type", "severity"},
		),

		alertEscalationEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_alert_escalations_total",
				Help: "Total number of alert escalation events",
			},
			[]string{"alert_type", "escalation_level"},
		),

		// Notification Metrics
		notificationsSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_notifications_sent_total",
				Help: "Total number of violation notifications sent",
			},
			[]string{"channel", "type", "severity"},
		),

		notificationDeliveryStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_notification_delivery_rate",
				Help: "Notification delivery success rate by channel",
			},
			[]string{"channel", "status"},
		),

		notificationChannels: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_notification_channels",
				Help: "Notification channel statistics",
			},
			[]string{"channel", "metric"},
		),

		notificationPreferences: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_notification_preferences",
				Help: "User notification preferences",
			},
			[]string{"user", "channel", "preference_type"},
		),

		// Remediation Tracking
		remediationActions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_remediation_actions_total",
				Help: "Total number of remediation actions taken",
			},
			[]string{"action_type", "trigger", "success"},
		),

		remediationEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_remediation_effectiveness",
				Help: "Effectiveness of remediation actions",
			},
			[]string{"action_type", "metric"},
		),

		remediationTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_violation_remediation_time_seconds",
				Help:    "Time taken for remediation actions",
				Buckets: []float64{300, 900, 1800, 3600, 7200, 14400, 28800},
			},
			[]string{"action_type", "automation_level"},
		),

		automatedRemediation: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_automated_remediation_total",
				Help: "Total number of automated remediation actions",
			},
			[]string{"action_type", "success", "trigger"},
		),

		remediationSuccess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_remediation_success_rate",
				Help: "Success rate of remediation actions",
			},
			[]string{"action_type", "time_period"},
		),

		// Policy Enforcement
		policyViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_policy_violations_total",
				Help: "Total number of policy violations",
			},
			[]string{"policy_name", "policy_type", "violation_type"},
		),

		policyEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_policy_effectiveness",
				Help: "Policy effectiveness metrics",
			},
			[]string{"policy_name", "metric"},
		),

		policyComplianceRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_policy_compliance_rate",
				Help: "Policy compliance rate",
			},
			[]string{"policy_name", "scope"},
		),

		policyUpdateFrequency: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_policy_updates_total",
				Help: "Total number of policy updates",
			},
			[]string{"policy_name", "update_type"},
		),

		policyValidationResults: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_policy_validation_results",
				Help: "Policy validation results",
			},
			[]string{"policy_name", "validation_type"},
		),

		// Threshold Management
		thresholdBreaches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_threshold_breaches_total",
				Help: "Total number of threshold breaches",
			},
			[]string{"threshold_type", "severity"},
		),

		thresholdOptimality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_threshold_optimality",
				Help: "Optimality score of violation thresholds",
			},
			[]string{"threshold_type", "metric"},
		),

		thresholdCalibration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_threshold_calibration",
				Help: "Threshold calibration metrics",
			},
			[]string{"threshold_type", "calibration_metric"},
		),

		dynamicThresholdAdjustment: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_dynamic_threshold_adjustments_total",
				Help: "Total number of dynamic threshold adjustments",
			},
			[]string{"threshold_type", "adjustment_direction"},
		),

		thresholdEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_threshold_effectiveness",
				Help: "Effectiveness of violation thresholds",
			},
			[]string{"threshold_type", "effectiveness_metric"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *FairShareViolationsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.fairShareViolations.Describe(ch)
	c.violationSeverity.Describe(ch)
	c.violationDuration.Describe(ch)
	c.violationType.Describe(ch)
	c.violationImpact.Describe(ch)
	c.violationTrends.Describe(ch)
	c.violationFrequency.Describe(ch)
	c.violationEscalation.Describe(ch)
	c.violationResolution.Describe(ch)
	c.violationRecurrence.Describe(ch)
	c.userViolationScore.Describe(ch)
	c.userViolationHistory.Describe(ch)
	c.userViolationPattern.Describe(ch)
	c.userComplianceScore.Describe(ch)
	c.userRiskAssessment.Describe(ch)
	c.accountViolationAggregate.Describe(ch)
	c.accountViolationRisk.Describe(ch)
	c.accountComplianceMetrics.Describe(ch)
	c.accountViolationImpact.Describe(ch)
	c.systemViolationOverview.Describe(ch)
	c.systemViolationThresholds.Describe(ch)
	c.systemComplianceHealth.Describe(ch)
	c.systemViolationPrediction.Describe(ch)
	c.alertsGenerated.Describe(ch)
	c.alertsAcknowledged.Describe(ch)
	c.alertsResolved.Describe(ch)
	c.alertResponseTime.Describe(ch)
	c.alertEscalationEvents.Describe(ch)
	c.notificationsSent.Describe(ch)
	c.notificationDeliveryStatus.Describe(ch)
	c.notificationChannels.Describe(ch)
	c.notificationPreferences.Describe(ch)
	c.remediationActions.Describe(ch)
	c.remediationEffectiveness.Describe(ch)
	c.remediationTime.Describe(ch)
	c.automatedRemediation.Describe(ch)
	c.remediationSuccess.Describe(ch)
	c.policyViolations.Describe(ch)
	c.policyEffectiveness.Describe(ch)
	c.policyComplianceRate.Describe(ch)
	c.policyUpdateFrequency.Describe(ch)
	c.policyValidationResults.Describe(ch)
	c.thresholdBreaches.Describe(ch)
	c.thresholdOptimality.Describe(ch)
	c.thresholdCalibration.Describe(ch)
	c.dynamicThresholdAdjustment.Describe(ch)
	c.thresholdEffectiveness.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *FairShareViolationsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Reset metrics
	c.resetMetrics()

	// Collect violations
	c.collectViolations(ctx)

	// Collect system overview
	c.collectSystemOverview(ctx)

	// Get sample users for violation profiles
	sampleUsers := c.getSampleUsers()
	for _, user := range sampleUsers {
		c.collectUserViolationProfile(ctx, user)
	}

	// Get sample accounts for violation summaries
	sampleAccounts := c.getSampleAccounts()
	for _, account := range sampleAccounts {
		c.collectAccountViolationSummary(ctx, account)
	}

	// Collect trend analysis
	c.collectTrendAnalysis(ctx)

	// Collect alert and notification stats
	c.collectAlertStats(ctx)

	// Collect remediation history
	c.collectRemediationHistory(ctx)

	// Collect policy compliance
	c.collectPolicyCompliance(ctx)

	// Collect threshold validation
	c.collectThresholdValidation(ctx)

	// Collect violation predictions
	c.collectViolationPredictions(ctx)

	// Collect all metrics
	c.fairShareViolations.Collect(ch)
	c.violationSeverity.Collect(ch)
	c.violationDuration.Collect(ch)
	c.violationType.Collect(ch)
	c.violationImpact.Collect(ch)
	c.violationTrends.Collect(ch)
	c.violationFrequency.Collect(ch)
	c.violationEscalation.Collect(ch)
	c.violationResolution.Collect(ch)
	c.violationRecurrence.Collect(ch)
	c.userViolationScore.Collect(ch)
	c.userViolationHistory.Collect(ch)
	c.userViolationPattern.Collect(ch)
	c.userComplianceScore.Collect(ch)
	c.userRiskAssessment.Collect(ch)
	c.accountViolationAggregate.Collect(ch)
	c.accountViolationRisk.Collect(ch)
	c.accountComplianceMetrics.Collect(ch)
	c.accountViolationImpact.Collect(ch)
	c.systemViolationOverview.Collect(ch)
	c.systemViolationThresholds.Collect(ch)
	c.systemComplianceHealth.Collect(ch)
	c.systemViolationPrediction.Collect(ch)
	c.alertsGenerated.Collect(ch)
	c.alertsAcknowledged.Collect(ch)
	c.alertsResolved.Collect(ch)
	c.alertResponseTime.Collect(ch)
	c.alertEscalationEvents.Collect(ch)
	c.notificationsSent.Collect(ch)
	c.notificationDeliveryStatus.Collect(ch)
	c.notificationChannels.Collect(ch)
	c.notificationPreferences.Collect(ch)
	c.remediationActions.Collect(ch)
	c.remediationEffectiveness.Collect(ch)
	c.remediationTime.Collect(ch)
	c.automatedRemediation.Collect(ch)
	c.remediationSuccess.Collect(ch)
	c.policyViolations.Collect(ch)
	c.policyEffectiveness.Collect(ch)
	c.policyComplianceRate.Collect(ch)
	c.policyUpdateFrequency.Collect(ch)
	c.policyValidationResults.Collect(ch)
	c.thresholdBreaches.Collect(ch)
	c.thresholdOptimality.Collect(ch)
	c.thresholdCalibration.Collect(ch)
	c.dynamicThresholdAdjustment.Collect(ch)
	c.thresholdEffectiveness.Collect(ch)
}

func (c *FairShareViolationsCollector) resetMetrics() {
	c.fairShareViolations.Reset()
	c.violationSeverity.Reset()
	c.violationDuration.Reset()
	c.violationType.Reset()
	c.violationImpact.Reset()
	c.violationTrends.Reset()
	c.violationRecurrence.Reset()
	c.userViolationScore.Reset()
	c.userViolationHistory.Reset()
	c.userViolationPattern.Reset()
	c.userComplianceScore.Reset()
	c.userRiskAssessment.Reset()
	c.accountViolationAggregate.Reset()
	c.accountViolationRisk.Reset()
	c.accountComplianceMetrics.Reset()
	c.accountViolationImpact.Reset()
	c.systemViolationOverview.Reset()
	c.systemViolationThresholds.Reset()
	c.systemComplianceHealth.Reset()
	c.systemViolationPrediction.Reset()
	c.notificationDeliveryStatus.Reset()
	c.notificationChannels.Reset()
	c.notificationPreferences.Reset()
	c.remediationEffectiveness.Reset()
	c.remediationSuccess.Reset()
	c.policyEffectiveness.Reset()
	c.policyComplianceRate.Reset()
	c.policyValidationResults.Reset()
	c.thresholdOptimality.Reset()
	c.thresholdCalibration.Reset()
	c.thresholdEffectiveness.Reset()
}

func (c *FairShareViolationsCollector) collectViolations(ctx context.Context) {
	violations, err := c.client.DetectFairShareViolations(ctx)
	if err != nil {
		log.Printf("Error detecting violations: %v", err)
		return
	}

	severityCounts := make(map[string]float64)
	typeCounts := make(map[string]float64)

	for _, violation := range violations {
		labels := []string{violation.UserName, violation.AccountName, violation.ViolationType, violation.Severity, violation.Status}
		c.fairShareViolations.WithLabelValues(labels...).Set(1.0)

		violationLabels := []string{violation.ViolationID, violation.UserName, violation.AccountName, violation.ViolationType}
		c.violationSeverity.WithLabelValues(violationLabels...).Set(violation.ViolationMagnitude)
		c.violationDuration.WithLabelValues(violationLabels...).Set(violation.Duration.Minutes())

		impactLabels := []string{violation.ViolationID, violation.UserName, violation.AccountName}
		c.violationImpact.WithLabelValues(append(impactLabels, "priority")...).Set(violation.PriorityImpact)
		c.violationImpact.WithLabelValues(append(impactLabels, "queue")...).Set(violation.QueueImpact)
		c.violationImpact.WithLabelValues(append(impactLabels, "system")...).Set(violation.SystemImpact)
		c.violationImpact.WithLabelValues(append(impactLabels, "user")...).Set(violation.UserImpact)
		c.violationImpact.WithLabelValues(append(impactLabels, "overall")...).Set(violation.OverallImpact)

		severityCounts[violation.Severity]++
		typeCounts[violation.ViolationType]++
	}

	for severity := range severityCounts {
		for violationType, typeCount := range typeCounts {
			c.violationType.WithLabelValues(violationType, severity).Set(typeCount)
		}
	}
}

func (c *FairShareViolationsCollector) collectSystemOverview(ctx context.Context) {
	overview, err := c.client.GetSystemViolationOverview(ctx)
	if err != nil {
		log.Printf("Error collecting system overview: %v", err)
		return
	}

	c.systemViolationOverview.WithLabelValues("total_violations").Set(float64(overview.TotalViolations))
	c.systemViolationOverview.WithLabelValues("active_violations").Set(float64(overview.ActiveViolations))
	c.systemViolationOverview.WithLabelValues("violations_today").Set(float64(overview.ViolationsToday))
	c.systemViolationOverview.WithLabelValues("violations_this_week").Set(float64(overview.ViolationsThisWeek))
	c.systemViolationOverview.WithLabelValues("violation_rate").Set(overview.ViolationRate)

	c.systemViolationOverview.WithLabelValues("critical_violations").Set(float64(overview.CriticalViolations))
	c.systemViolationOverview.WithLabelValues("high_violations").Set(float64(overview.HighViolations))
	c.systemViolationOverview.WithLabelValues("medium_violations").Set(float64(overview.MediumViolations))
	c.systemViolationOverview.WithLabelValues("low_violations").Set(float64(overview.LowViolations))

	c.systemComplianceHealth.WithLabelValues("compliance_health").Set(overview.ComplianceHealth)
	c.systemComplianceHealth.WithLabelValues("system_stress_level").Set(overview.SystemStressLevel)
	c.systemComplianceHealth.WithLabelValues("alert_load").Set(overview.AlertLoad)

	c.systemViolationThresholds.WithLabelValues("threshold", "breaches").Set(float64(overview.ThresholdBreaches))
	c.systemViolationThresholds.WithLabelValues("threshold", "adjustments").Set(float64(overview.ThresholdAdjustments))
	c.systemViolationThresholds.WithLabelValues("threshold", "effectiveness").Set(overview.ThresholdEffectiveness)

	c.systemViolationOverview.WithLabelValues("detection_accuracy").Set(overview.DetectionAccuracy)
	c.systemViolationOverview.WithLabelValues("false_positive_rate").Set(overview.FalsePositiveRate)
	c.systemViolationOverview.WithLabelValues("response_efficiency").Set(overview.ResponseEfficiency)
	c.systemViolationOverview.WithLabelValues("automation_rate").Set(overview.AutomationRate)
}

func (c *FairShareViolationsCollector) collectUserViolationProfile(ctx context.Context, userName string) {
	profile, err := c.client.GetUserViolationProfile(ctx, userName)
	if err != nil {
		log.Printf("Error collecting user violation profile for %s: %v", userName, err)
		return
	}

	labels := []string{profile.UserName, profile.AccountName}

	c.userViolationScore.WithLabelValues(append(labels, "violation_score")...).Set(profile.ViolationScore)
	c.userViolationScore.WithLabelValues(append(labels, "violation_rate")...).Set(profile.ViolationRate)
	c.userViolationScore.WithLabelValues(append(labels, "risk_score")...).Set(profile.RiskScore)

	c.userViolationHistory.WithLabelValues(append(labels, []string{"total_violations", "all_time"}...)...).Set(float64(profile.TotalViolations))
	c.userViolationHistory.WithLabelValues(append(labels, []string{"active_violations", "current"}...)...).Set(float64(profile.ActiveViolations))
	c.userViolationHistory.WithLabelValues(append(labels, []string{"resolved_violations", "all_time"}...)...).Set(float64(profile.ResolvedViolations))

	c.userViolationPattern.WithLabelValues(append(labels, "violation_frequency")...).Set(profile.ViolationFrequency)
	c.userViolationPattern.WithLabelValues(append(labels, "recurrence_rate")...).Set(profile.RecurrenceRate)
	c.userViolationPattern.WithLabelValues(append(labels, "escalation_rate")...).Set(profile.EscalationRate)

	c.userComplianceScore.WithLabelValues(append(labels, "compliance_score")...).Set(profile.ComplianceScore)
	c.userComplianceScore.WithLabelValues(append(labels, "adaptation_score")...).Set(profile.AdaptationScore)

	c.userRiskAssessment.WithLabelValues(append(labels, "cooperation_level")...).Set(profile.CooperationLevel)
	c.userRiskAssessment.WithLabelValues(append(labels, "remediation_success")...).Set(profile.RemediationSuccess)
	c.userRiskAssessment.WithLabelValues(append(labels, "average_response_time")...).Set(profile.AverageResponseTime)
}

func (c *FairShareViolationsCollector) collectAccountViolationSummary(ctx context.Context, accountName string) {
	summary, err := c.client.GetAccountViolationSummary(ctx, accountName)
	if err != nil {
		log.Printf("Error collecting account violation summary for %s: %v", accountName, err)
		return
	}

	c.accountViolationAggregate.WithLabelValues(summary.AccountName, "total_users").Set(float64(summary.TotalUsers))
	c.accountViolationAggregate.WithLabelValues(summary.AccountName, "users_with_violations").Set(float64(summary.UsersWithViolations))
	c.accountViolationAggregate.WithLabelValues(summary.AccountName, "total_violations").Set(float64(summary.TotalViolations))
	c.accountViolationAggregate.WithLabelValues(summary.AccountName, "active_violations").Set(float64(summary.ActiveViolations))
	c.accountViolationAggregate.WithLabelValues(summary.AccountName, "violation_rate").Set(summary.ViolationRate)

	c.accountViolationRisk.WithLabelValues(summary.AccountName, "risk_score").Set(summary.AccountRiskScore)
	c.accountComplianceMetrics.WithLabelValues(summary.AccountName, "compliance_score").Set(summary.ComplianceScore)
	c.accountComplianceMetrics.WithLabelValues(summary.AccountName, "policy_adherence").Set(summary.PolicyAdherence)
	c.accountViolationImpact.WithLabelValues(summary.AccountName, "violation_impact").Set(summary.ViolationImpact)

	c.accountComplianceMetrics.WithLabelValues(summary.AccountName, "management_response").Set(summary.ManagementResponse)
	c.accountComplianceMetrics.WithLabelValues(summary.AccountName, "remediation_rate").Set(summary.RemediationRate)
	c.accountComplianceMetrics.WithLabelValues(summary.AccountName, "prevention_effectiveness").Set(summary.PreventionEffectiveness)
}

func (c *FairShareViolationsCollector) collectTrendAnalysis(ctx context.Context) {
	periods := []string{"24h", "7d", "30d"}
	for _, period := range periods {
		trend, err := c.client.GetViolationTrendAnalysis(ctx, period)
		if err != nil {
			continue
		}

		c.violationTrends.WithLabelValues("velocity", period, trend.TrendDirection).Set(trend.TrendVelocity)
		c.violationTrends.WithLabelValues("acceleration", period, trend.TrendDirection).Set(trend.TrendAcceleration)
		c.violationTrends.WithLabelValues("volatility", period, trend.TrendDirection).Set(trend.TrendVolatility)

		c.systemViolationPrediction.WithLabelValues("predicted_violations", period).Set(float64(trend.PredictedViolations))
		c.systemViolationPrediction.WithLabelValues("prediction_confidence", period).Set(trend.PredictionConfidence)
	}
}

func (c *FairShareViolationsCollector) collectAlertStats(ctx context.Context) {
	period := "24h"
	alertHistory, err := c.client.GetAlertHistory(ctx, period)
	if err != nil {
		log.Printf("Error collecting alert history: %v", err)
		return
	}

	c.alertResponseTime.WithLabelValues("all", "medium").Observe(alertHistory.AverageResponseTime)

	// Notification stats
	notificationStats, err := c.client.GetNotificationStats(ctx)
	if err == nil {
		c.notificationDeliveryStatus.WithLabelValues("all", "delivered").Set(notificationStats.DeliveryRate)
		c.notificationChannels.WithLabelValues("all", "total_notifications").Set(float64(notificationStats.TotalNotifications))
		c.notificationChannels.WithLabelValues("all", "successful_deliveries").Set(float64(notificationStats.SuccessfulDeliveries))
		c.notificationChannels.WithLabelValues("all", "failed_deliveries").Set(float64(notificationStats.FailedDeliveries))

		for channelName, stats := range notificationStats.ChannelStats {
			c.notificationChannels.WithLabelValues(channelName, "messages_sent").Set(float64(stats.MessagesSent))
			c.notificationChannels.WithLabelValues(channelName, "delivery_success").Set(stats.DeliverySuccess)
			c.notificationChannels.WithLabelValues(channelName, "response_rate").Set(stats.ResponseRate)
		}
	}
}

func (c *FairShareViolationsCollector) collectRemediationHistory(ctx context.Context) {
	period := "24h"
	remediationHistory, err := c.client.GetRemediationHistory(ctx, period)
	if err != nil {
		log.Printf("Error collecting remediation history: %v", err)
		return
	}

	c.remediationEffectiveness.WithLabelValues("all", "success_rate").Set(remediationHistory.RemediationSuccess)
	c.remediationEffectiveness.WithLabelValues("all", "automation_effectiveness").Set(remediationHistory.AutomationEffectiveness)
	c.remediationEffectiveness.WithLabelValues("all", "recurrence_prevention").Set(remediationHistory.RecurrencePrevention)

	c.remediationSuccess.WithLabelValues("all", period).Set(remediationHistory.RemediationSuccess)
	c.remediationTime.WithLabelValues("all", "automated").Observe(remediationHistory.AverageRemediationTime)
}

func (c *FairShareViolationsCollector) collectPolicyCompliance(ctx context.Context) {
	complianceReport, err := c.client.GetPolicyComplianceReport(ctx)
	if err != nil {
		log.Printf("Error collecting policy compliance: %v", err)
		return
	}

	c.policyComplianceRate.WithLabelValues("all", "system").Set(complianceReport.OverallCompliance)
	c.policyEffectiveness.WithLabelValues("all", "effectiveness").Set(complianceReport.PolicyEffectiveness)
	c.policyEffectiveness.WithLabelValues("all", "enforcement_rate").Set(complianceReport.EnforcementRate)
	c.policyEffectiveness.WithLabelValues("all", "auto_enforcement").Set(complianceReport.AutoEnforcement)
}

func (c *FairShareViolationsCollector) collectThresholdValidation(ctx context.Context) {
	validation, err := c.client.ValidateViolationThresholds(ctx)
	if err != nil {
		log.Printf("Error collecting threshold validation: %v", err)
		return
	}

	c.thresholdOptimality.WithLabelValues("all", "validation_score").Set(validation.ValidationScore)
	c.thresholdOptimality.WithLabelValues("all", "threshold_accuracy").Set(validation.ThresholdAccuracy)
	c.thresholdOptimality.WithLabelValues("all", "optimality_score").Set(validation.OptimalityScore)

	c.thresholdCalibration.WithLabelValues("all", "false_positive_rate").Set(validation.FalsePositiveRate)
	c.thresholdCalibration.WithLabelValues("all", "false_negative_rate").Set(validation.FalseNegativeRate)
	c.thresholdCalibration.WithLabelValues("all", "detection_accuracy").Set(validation.DetectionAccuracy)
	c.thresholdCalibration.WithLabelValues("all", "sensitivity").Set(validation.Sensitivity)
	c.thresholdCalibration.WithLabelValues("all", "specificity").Set(validation.Specificity)
}

func (c *FairShareViolationsCollector) collectViolationPredictions(ctx context.Context) {
	scopes := []string{"system", "user", "account"}
	for _, scope := range scopes {
		prediction, err := c.client.PredictViolationRisk(ctx, scope)
		if err != nil {
			continue
		}

		c.systemViolationPrediction.WithLabelValues("risk_score", prediction.PredictionHorizon).Set(prediction.RiskScore)
		c.systemViolationPrediction.WithLabelValues("predicted_violations", prediction.PredictionHorizon).Set(float64(prediction.PredictedViolations))
		c.systemViolationPrediction.WithLabelValues("prediction_confidence", prediction.PredictionHorizon).Set(prediction.PredictionConfidence)
	}
}

// Simplified mock data generators for testing purposes
func (c *FairShareViolationsCollector) getSampleUsers() []string {
	return []string{"user1", "user2", "user3"}
}

func (c *FairShareViolationsCollector) getSampleAccounts() []string {
	return []string{"account1", "account2", "account3"}
}
