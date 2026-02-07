// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type QuotaComplianceSLURMClient interface {
	GetQuotaCompliance(ctx context.Context, entityType, entityName string) (*QuotaComplianceReport, error)
	ListQuotaEntities(ctx context.Context, entityType string) ([]*QuotaEntity, error)
	GetQuotaViolations(ctx context.Context, entityType, entityName string) ([]*QuotaComplianceViolation, error)
	GetQuotaAlerts(ctx context.Context, entityType, entityName string) ([]*QuotaComplianceAlert, error)
	GetQuotaThresholds(ctx context.Context, entityType, entityName string) (*QuotaThresholds, error)
	GetQuotaPolicies(ctx context.Context, entityType string) ([]*QuotaPolicy, error)
	GetQuotaEnforcement(ctx context.Context, entityType, entityName string) (*QuotaEnforcement, error)
	GetQuotaTrends(ctx context.Context, entityType, entityName string) (*QuotaTrends, error)
	GetQuotaForecast(ctx context.Context, entityType, entityName string) (*QuotaForecast, error)
	GetQuotaRecommendations(ctx context.Context, entityType, entityName string) (*QuotaRecommendations, error)
	GetQuotaAuditLog(ctx context.Context, entityType, entityName string) ([]*QuotaAuditEntry, error)
	GetSystemQuotaOverview(ctx context.Context) (*SystemQuotaOverview, error)
	GetQuotaNotificationSettings(ctx context.Context, entityType, entityName string) (*QuotaNotificationSettings, error)
}

type QuotaComplianceReport struct {
	EntityType         string                `json:"entity_type"`
	EntityName         string                `json:"entity_name"`
	ComplianceScore    float64               `json:"compliance_score"`
	ComplianceStatus   string                `json:"compliance_status"`
	LastAssessment     time.Time             `json:"last_assessment"`
	ResourceCompliance []*ResourceCompliance `json:"resource_compliance"`
	PolicyCompliance   []*PolicyCompliance   `json:"policy_compliance"`
	ViolationSummary   *ViolationSummary     `json:"violation_summary"`
	ComplianceHistory  []*ComplianceSnapshot `json:"compliance_history"`
	RiskLevel          string                `json:"risk_level"`
	RecommendedActions []string              `json:"recommended_actions"`
	NextReview         time.Time             `json:"next_review"`
}

type ResourceCompliance struct {
	ResourceType        string     `json:"resource_type"`
	AllocatedQuota      float64    `json:"allocated_quota"`
	UsedQuota           float64    `json:"used_quota"`
	UtilizationRate     float64    `json:"utilization_rate"`
	ComplianceStatus    string     `json:"compliance_status"`
	ThresholdBreaches   int        `json:"threshold_breaches"`
	LastBreach          *time.Time `json:"last_breach,omitempty"`
	ComplianceScore     float64    `json:"compliance_score"`
	TrendDirection      string     `json:"trend_direction"`
	ProjectedExhaustion *time.Time `json:"projected_exhaustion,omitempty"`
}

type PolicyCompliance struct {
	PolicyName        string         `json:"policy_name"`
	PolicyType        string         `json:"policy_type"`
	ComplianceStatus  string         `json:"compliance_status"`
	ComplianceScore   float64        `json:"compliance_score"`
	ViolationCount    int            `json:"violation_count"`
	LastViolation     *time.Time     `json:"last_violation,omitempty"`
	PolicyDescription string         `json:"policy_description"`
	RequiredActions   []string       `json:"required_actions"`
	GracePeriod       *time.Duration `json:"grace_period,omitempty"`
	EnforcementLevel  string         `json:"enforcement_level"`
}

type ViolationSummary struct {
	TotalViolations       int        `json:"total_violations"`
	ActiveViolations      int        `json:"active_violations"`
	ResolvedViolations    int        `json:"resolved_violations"`
	CriticalViolations    int        `json:"critical_violations"`
	WarningViolations     int        `json:"warning_violations"`
	LastViolation         *time.Time `json:"last_violation,omitempty"`
	ViolationRate         float64    `json:"violation_rate"`
	AverageResolutionTime float64    `json:"average_resolution_time"`
	RecurrentViolations   int        `json:"recurrent_violations"`
}

type ComplianceSnapshot struct {
	Timestamp           time.Time `json:"timestamp"`
	ComplianceScore     float64   `json:"compliance_score"`
	ViolationCount      int       `json:"violation_count"`
	ResourceUtilization float64   `json:"resource_utilization"`
	PolicyBreaches      int       `json:"policy_breaches"`
}

type QuotaEntity struct {
	EntityType        string    `json:"entity_type"`
	EntityName        string    `json:"entity_name"`
	EntityDescription string    `json:"entity_description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Status            string    `json:"status"`
	ParentEntity      string    `json:"parent_entity,omitempty"`
	QuotaCount        int       `json:"quota_count"`
	ComplianceStatus  string    `json:"compliance_status"`
}

type QuotaComplianceViolation struct {
	ViolationID        string     `json:"violation_id"`
	EntityType         string     `json:"entity_type"`
	EntityName         string     `json:"entity_name"`
	ViolationType      string     `json:"violation_type"`
	ResourceType       string     `json:"resource_type"`
	Severity           string     `json:"severity"`
	DetectedAt         time.Time  `json:"detected_at"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	Description        string     `json:"description"`
	CurrentUsage       float64    `json:"current_usage"`
	QuotaLimit         float64    `json:"quota_limit"`
	ExcessAmount       float64    `json:"excess_amount"`
	ExcessPercentage   float64    `json:"excess_percentage"`
	Impact             string     `json:"impact"`
	RootCause          string     `json:"root_cause"`
	ResolutionAction   string     `json:"resolution_action"`
	PreventionMeasures []string   `json:"prevention_measures"`
	IsRecurrent        bool       `json:"is_recurrent"`
	RecurrenceCount    int        `json:"recurrence_count"`
	LastOccurrence     *time.Time `json:"last_occurrence,omitempty"`
	Status             string     `json:"status"`
}

type QuotaComplianceAlert struct {
	AlertID             string     `json:"alert_id"`
	EntityType          string     `json:"entity_type"`
	EntityName          string     `json:"entity_name"`
	AlertType           string     `json:"alert_type"`
	Priority            string     `json:"priority"`
	Severity            string     `json:"severity"`
	TriggeredAt         time.Time  `json:"triggered_at"`
	AcknowledgedAt      *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty"`
	Message             string     `json:"message"`
	ResourceType        string     `json:"resource_type"`
	CurrentValue        float64    `json:"current_value"`
	ThresholdValue      float64    `json:"threshold_value"`
	ThresholdType       string     `json:"threshold_type"`
	TrendDirection      string     `json:"trend_direction"`
	PredictedExhaustion *time.Time `json:"predicted_exhaustion,omitempty"`
	NotificationSent    bool       `json:"notification_sent"`
	EscalationLevel     int        `json:"escalation_level"`
	ActionRequired      string     `json:"action_required"`
	RecommendedResponse []string   `json:"recommended_response"`
	Status              string     `json:"status"`
}

type QuotaThresholds struct {
	EntityType         string               `json:"entity_type"`
	EntityName         string               `json:"entity_name"`
	ResourceThresholds []*ResourceThreshold `json:"resource_thresholds"`
	AlertThresholds    []*AlertThreshold    `json:"alert_thresholds"`
	PolicyThresholds   []*PolicyThreshold   `json:"policy_thresholds"`
	GlobalThresholds   []*GlobalThreshold   `json:"global_thresholds"`
	CustomThresholds   []*CustomThreshold   `json:"custom_thresholds"`
	LastUpdated        time.Time            `json:"last_updated"`
	UpdatedBy          string               `json:"updated_by"`
}

type ResourceThreshold struct {
	ResourceType      string     `json:"resource_type"`
	WarningThreshold  float64    `json:"warning_threshold"`
	CriticalThreshold float64    `json:"critical_threshold"`
	MaxThreshold      float64    `json:"max_threshold"`
	ThresholdUnit     string     `json:"threshold_unit"`
	EvaluationPeriod  int        `json:"evaluation_period"`
	CooldownPeriod    int        `json:"cooldown_period"`
	IsEnabled         bool       `json:"is_enabled"`
	LastTriggered     *time.Time `json:"last_triggered,omitempty"`
	TriggerCount      int        `json:"trigger_count"`
}

type AlertThreshold struct {
	AlertType          string   `json:"alert_type"`
	Threshold          float64  `json:"threshold"`
	ComparisonOperator string   `json:"comparison_operator"`
	EvaluationWindow   int      `json:"evaluation_window"`
	TriggerCondition   string   `json:"trigger_condition"`
	NotificationDelay  int      `json:"notification_delay"`
	EscalationRules    []string `json:"escalation_rules"`
	IsActive           bool     `json:"is_active"`
}

type PolicyThreshold struct {
	PolicyName         string `json:"policy_name"`
	ViolationThreshold int    `json:"violation_threshold"`
	TimeWindow         int    `json:"time_window"`
	ActionThreshold    string `json:"action_threshold"`
	EnforcementDelay   int    `json:"enforcement_delay"`
	GracePeriod        int    `json:"grace_period"`
	IsStrict           bool   `json:"is_strict"`
}

type GlobalThreshold struct {
	ThresholdName      string   `json:"threshold_name"`
	ThresholdValue     float64  `json:"threshold_value"`
	ApplicableEntities []string `json:"applicable_entities"`
	Priority           int      `json:"priority"`
	OverrideAllowed    bool     `json:"override_allowed"`
	InheritanceRules   []string `json:"inheritance_rules"`
}

type CustomThreshold struct {
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Condition          string    `json:"condition"`
	ThresholdValue     float64   `json:"threshold_value"`
	Metric             string    `json:"metric"`
	EvaluationInterval int       `json:"evaluation_interval"`
	CreatedBy          string    `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	IsActive           bool      `json:"is_active"`
}

type QuotaPolicy struct {
	PolicyID           string              `json:"policy_id"`
	PolicyName         string              `json:"policy_name"`
	PolicyType         string              `json:"policy_type"`
	Description        string              `json:"description"`
	ApplicableEntities []string            `json:"applicable_entities"`
	ResourceRules      []*ResourceRule     `json:"resource_rules"`
	EnforcementRules   []*EnforcementRule  `json:"enforcement_rules"`
	ViolationHandling  *ViolationHandling  `json:"violation_handling"`
	NotificationRules  []*NotificationRule `json:"notification_rules"`
	ExceptionRules     []*ExceptionRule    `json:"exception_rules"`
	ValidityPeriod     *ValidityPeriod     `json:"validity_period"`
	Priority           int                 `json:"priority"`
	IsActive           bool                `json:"is_active"`
	CreatedBy          string              `json:"created_by"`
	CreatedAt          time.Time           `json:"created_at"`
	LastModified       time.Time           `json:"last_modified"`
	VersionNumber      int                 `json:"version_number"`
}

type ResourceRule struct {
	ResourceType     string   `json:"resource_type"`
	MinQuota         float64  `json:"min_quota"`
	MaxQuota         float64  `json:"max_quota"`
	DefaultQuota     float64  `json:"default_quota"`
	AllocationRules  []string `json:"allocation_rules"`
	SharingRules     []string `json:"sharing_rules"`
	InheritanceRules []string `json:"inheritance_rules"`
	RenewlRules      []string `json:"renewal_rules"`
}

type EnforcementRule struct {
	RuleName         string   `json:"rule_name"`
	EnforcementType  string   `json:"enforcement_type"`
	TriggerCondition string   `json:"trigger_condition"`
	Action           string   `json:"action"`
	GracePeriod      int      `json:"grace_period"`
	EscalationSteps  []string `json:"escalation_steps"`
	AutoRemediation  bool     `json:"auto_remediation"`
	ManualOverride   bool     `json:"manual_override"`
}

type ViolationHandling struct {
	ImmediateActions          []string `json:"immediate_actions"`
	EscalationActions         []string `json:"escalation_actions"`
	ResolutionActions         []string `json:"resolution_actions"`
	PreventionMeasures        []string `json:"prevention_measures"`
	DocumentationRequirements []string `json:"documentation_requirements"`
	ApprovalRequirements      []string `json:"approval_requirements"`
}

type QuotaNotificationRule struct {
	RuleName            string   `json:"rule_name"`
	TriggerEvents       []string `json:"trigger_events"`
	Recipients          []string `json:"recipients"`
	NotificationMethods []string `json:"notification_methods"`
	MessageTemplate     string   `json:"message_template"`
	Frequency           string   `json:"frequency"`
	EscalationDelay     int      `json:"escalation_delay"`
	IsActive            bool     `json:"is_active"`
}

type ExceptionRule struct {
	RuleName              string   `json:"rule_name"`
	Conditions            []string `json:"conditions"`
	ExceptionType         string   `json:"exception_type"`
	Duration              int      `json:"duration"`
	ApprovalRequired      bool     `json:"approval_required"`
	JustificationRequired bool     `json:"justification_required"`
	ReviewPeriod          int      `json:"review_period"`
	IsTemporary           bool     `json:"is_temporary"`
}

type ValidityPeriod struct {
	StartDate         time.Time   `json:"start_date"`
	EndDate           *time.Time  `json:"end_date,omitempty"`
	Timezone          string      `json:"timezone"`
	RecurrencePattern string      `json:"recurrence_pattern"`
	ExceptionDates    []time.Time `json:"exception_dates"`
}

type QuotaEnforcement struct {
	EntityType         string               `json:"entity_type"`
	EntityName         string               `json:"entity_name"`
	EnforcementStatus  string               `json:"enforcement_status"`
	EnforcementMode    string               `json:"enforcement_mode"`
	ActivePolicies     []string             `json:"active_policies"`
	EnforcementActions []*EnforcementAction `json:"enforcement_actions"`
	ViolationHistory   []*ViolationRecord   `json:"violation_history"`
	ExceptionStatus    *ExceptionStatus     `json:"exception_status"`
	OverrideStatus     *OverrideStatus      `json:"override_status"`
	ComplianceMetrics  *ComplianceMetrics   `json:"compliance_metrics"`
	LastEnforcement    time.Time            `json:"last_enforcement"`
	NextEvaluation     time.Time            `json:"next_evaluation"`
}

type EnforcementAction struct {
	ActionID          string    `json:"action_id"`
	ActionType        string    `json:"action_type"`
	ExecutedAt        time.Time `json:"executed_at"`
	ExecutedBy        string    `json:"executed_by"`
	Description       string    `json:"description"`
	AffectedResources []string  `json:"affected_resources"`
	Impact            string    `json:"impact"`
	Duration          *int      `json:"duration,omitempty"`
	Status            string    `json:"status"`
	Result            string    `json:"result"`
	RollbackAvailable bool      `json:"rollback_available"`
}

type ViolationRecord struct {
	Timestamp      time.Time `json:"timestamp"`
	ViolationType  string    `json:"violation_type"`
	Severity       string    `json:"severity"`
	ActionTaken    string    `json:"action_taken"`
	ResolutionTime int       `json:"resolution_time"`
	WasRecurrent   bool      `json:"was_recurrent"`
}

type ExceptionStatus struct {
	HasActiveExceptions bool               `json:"has_active_exceptions"`
	ActiveExceptions    []*ActiveException `json:"active_exceptions"`
	ExceptionHistory    []*ExceptionRecord `json:"exception_history"`
	TotalExceptions     int                `json:"total_exceptions"`
	LastException       *time.Time         `json:"last_exception,omitempty"`
}

type ActiveException struct {
	ExceptionID         string    `json:"exception_id"`
	ExceptionType       string    `json:"exception_type"`
	GrantedAt           time.Time `json:"granted_at"`
	ExpiresAt           time.Time `json:"expires_at"`
	GrantedBy           string    `json:"granted_by"`
	Justification       string    `json:"justification"`
	ApplicableResources []string  `json:"applicable_resources"`
	IsTemporary         bool      `json:"is_temporary"`
}

type ExceptionRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	ExceptionType string    `json:"exception_type"`
	Duration      int       `json:"duration"`
	WasApproved   bool      `json:"was_approved"`
	WasRevoked    bool      `json:"was_revoked"`
}

type OverrideStatus struct {
	HasActiveOverrides bool              `json:"has_active_overrides"`
	ActiveOverrides    []*ActiveOverride `json:"active_overrides"`
	OverrideHistory    []*OverrideRecord `json:"override_history"`
	TotalOverrides     int               `json:"total_overrides"`
	LastOverride       *time.Time        `json:"last_override,omitempty"`
}

type ActiveOverride struct {
	OverrideID       string     `json:"override_id"`
	OverrideType     string     `json:"override_type"`
	AppliedAt        time.Time  `json:"applied_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	AppliedBy        string     `json:"applied_by"`
	Reason           string     `json:"reason"`
	AffectedPolicies []string   `json:"affected_policies"`
	IsPermanent      bool       `json:"is_permanent"`
}

type OverrideRecord struct {
	Timestamp     time.Time `json:"timestamp"`
	OverrideType  string    `json:"override_type"`
	Duration      *int      `json:"duration,omitempty"`
	WasAuthorized bool      `json:"was_authorized"`
	WasReverted   bool      `json:"was_reverted"`
}

type ComplianceMetrics struct {
	ComplianceRate        float64 `json:"compliance_rate"`
	ViolationFrequency    float64 `json:"violation_frequency"`
	AverageResolutionTime float64 `json:"average_resolution_time"`
	RecurrenceRate        float64 `json:"recurrence_rate"`
	PolicyAdherence       float64 `json:"policy_adherence"`
	ExceptionRate         float64 `json:"exception_rate"`
	OverrideRate          float64 `json:"override_rate"`
	TrendDirection        string  `json:"trend_direction"`
}

type QuotaComplianceTrends struct {
	EntityType         string             `json:"entity_type"`
	EntityName         string             `json:"entity_name"`
	AnalysisPeriod     string             `json:"analysis_period"`
	UsageTrends        []*UsageTrend      `json:"usage_trends"`
	ComplianceTrends   []*ComplianceTrend `json:"compliance_trends"`
	ViolationTrends    []*ViolationTrend  `json:"violation_trends"`
	ThresholdTrends    []*ThresholdTrend  `json:"threshold_trends"`
	PolicyTrends       []*PolicyTrend     `json:"policy_trends"`
	SeasonalPatterns   []*SeasonalPattern `json:"seasonal_patterns"`
	AnomalyDetection   *AnomalyDetection  `json:"anomaly_detection"`
	TrendPredictions   []*TrendPrediction `json:"trend_predictions"`
	InfluencingFactors []string           `json:"influencing_factors"`
	RecommendedActions []string           `json:"recommended_actions"`
}

type UsageTrend struct {
	ResourceType    string   `json:"resource_type"`
	TrendDirection  string   `json:"trend_direction"`
	GrowthRate      float64  `json:"growth_rate"`
	Volatility      float64  `json:"volatility"`
	Seasonality     bool     `json:"seasonality"`
	PeakPeriods     []string `json:"peak_periods"`
	LowPeriods      []string `json:"low_periods"`
	AverageUsage    float64  `json:"average_usage"`
	TrendConfidence float64  `json:"trend_confidence"`
}

type ComplianceTrend struct {
	TrendPeriod       string  `json:"trend_period"`
	ComplianceScore   float64 `json:"compliance_score"`
	ScoreChange       float64 `json:"score_change"`
	TrendDirection    string  `json:"trend_direction"`
	ImprovementRate   float64 `json:"improvement_rate"`
	DeteriorationRate float64 `json:"deterioration_rate"`
	StabilityScore    float64 `json:"stability_score"`
}

type ViolationTrend struct {
	ViolationType   string  `json:"violation_type"`
	Frequency       float64 `json:"frequency"`
	FrequencyChange float64 `json:"frequency_change"`
	Severity        string  `json:"severity"`
	ResolutionTime  float64 `json:"resolution_time"`
	RecurrenceRate  float64 `json:"recurrence_rate"`
	TrendDirection  string  `json:"trend_direction"`
}

type ThresholdTrend struct {
	ThresholdType          string   `json:"threshold_type"`
	BreachFrequency        float64  `json:"breach_frequency"`
	AverageExcess          float64  `json:"average_excess"`
	TimeToThreshold        float64  `json:"time_to_threshold"`
	ThresholdEffectiveness float64  `json:"threshold_effectiveness"`
	AdjustmentHistory      []string `json:"adjustment_history"`
}

type PolicyTrend struct {
	PolicyName         string  `json:"policy_name"`
	AdherenceRate      float64 `json:"adherence_rate"`
	ViolationCount     int     `json:"violation_count"`
	ExceptionCount     int     `json:"exception_count"`
	EffectivenessScore float64 `json:"effectiveness_score"`
	TrendDirection     string  `json:"trend_direction"`
}

type QuotaSeasonalPattern struct {
	PatternType     string  `json:"pattern_type"`
	PatternStrength float64 `json:"pattern_strength"`
	PatternPeriod   string  `json:"pattern_period"`
	PeakPhase       string  `json:"peak_phase"`
	LowPhase        string  `json:"low_phase"`
	Amplitude       float64 `json:"amplitude"`
	Regularity      float64 `json:"regularity"`
}

type AnomalyDetection struct {
	AnomaliesDetected   int        `json:"anomalies_detected"`
	AnomalyRate         float64    `json:"anomaly_rate"`
	LastAnomaly         *time.Time `json:"last_anomaly,omitempty"`
	AnomalyTypes        []string   `json:"anomaly_types"`
	DetectionConfidence float64    `json:"detection_confidence"`
	AnomalyImpact       string     `json:"anomaly_impact"`
}

type TrendPrediction struct {
	PredictionHorizon  string   `json:"prediction_horizon"`
	PredictedValue     float64  `json:"predicted_value"`
	ConfidenceInterval float64  `json:"confidence_interval"`
	PredictionAccuracy float64  `json:"prediction_accuracy"`
	RiskLevel          string   `json:"risk_level"`
	RecommendedActions []string `json:"recommended_actions"`
}

type QuotaForecast struct {
	EntityType          string                `json:"entity_type"`
	EntityName          string                `json:"entity_name"`
	ForecastHorizon     string                `json:"forecast_horizon"`
	ResourceForecasts   []*ResourceForecast   `json:"resource_forecasts"`
	ComplianceForecasts []*ComplianceForecast `json:"compliance_forecasts"`
	ViolationForecasts  []*ViolationForecast  `json:"violation_forecasts"`
	CapacityForecasts   []*CapacityForecast   `json:"capacity_forecasts"`
	CostForecasts       []*CostForecast       `json:"cost_forecasts"`
	RiskForecasts       []*RiskForecast       `json:"risk_forecasts"`
	ScenarioAnalysis    []*ScenarioForecast   `json:"scenario_analysis"`
	ModelAccuracy       *ModelAccuracy        `json:"model_accuracy"`
	ForecastAssumptions []string              `json:"forecast_assumptions"`
	UpdateFrequency     string                `json:"update_frequency"`
	LastUpdated         time.Time             `json:"last_updated"`
}

type ResourceForecast struct {
	ResourceType        string     `json:"resource_type"`
	CurrentUsage        float64    `json:"current_usage"`
	PredictedUsage      float64    `json:"predicted_usage"`
	PredictedGrowth     float64    `json:"predicted_growth"`
	ConfidenceLevel     float64    `json:"confidence_level"`
	ProjectedExhaustion *time.Time `json:"projected_exhaustion,omitempty"`
	RecommendedQuota    float64    `json:"recommended_quota"`
	RiskLevel           string     `json:"risk_level"`
}

type ComplianceForecast struct {
	ForecastPeriod  string   `json:"forecast_period"`
	PredictedScore  float64  `json:"predicted_score"`
	ScoreChange     float64  `json:"score_change"`
	RiskLevel       string   `json:"risk_level"`
	ConfidenceLevel float64  `json:"confidence_level"`
	RequiredActions []string `json:"required_actions"`
}

type ViolationForecast struct {
	ViolationType      string     `json:"violation_type"`
	PredictedFrequency float64    `json:"predicted_frequency"`
	PredictedSeverity  string     `json:"predicted_severity"`
	ProbabilityScore   float64    `json:"probability_score"`
	TimeToViolation    *time.Time `json:"time_to_violation,omitempty"`
	PreventionMeasures []string   `json:"prevention_measures"`
}

type CapacityForecast struct {
	ResourceType        string     `json:"resource_type"`
	CurrentCapacity     float64    `json:"current_capacity"`
	DemandForecast      float64    `json:"demand_forecast"`
	CapacityGap         float64    `json:"capacity_gap"`
	RecommendedIncrease float64    `json:"recommended_increase"`
	TimeToCapacity      *time.Time `json:"time_to_capacity,omitempty"`
	ExpansionCost       float64    `json:"expansion_cost"`
}

type CostForecast struct {
	CostType              string   `json:"cost_type"`
	CurrentCost           float64  `json:"current_cost"`
	PredictedCost         float64  `json:"predicted_cost"`
	CostChange            float64  `json:"cost_change"`
	CostDrivers           []string `json:"cost_drivers"`
	OptimizationPotential float64  `json:"optimization_potential"`
}

type RiskForecast struct {
	RiskType           string   `json:"risk_type"`
	CurrentRiskLevel   string   `json:"current_risk_level"`
	PredictedRiskLevel string   `json:"predicted_risk_level"`
	RiskProbability    float64  `json:"risk_probability"`
	RiskImpact         string   `json:"risk_impact"`
	MitigationActions  []string `json:"mitigation_actions"`
}

type ScenarioForecast struct {
	ScenarioName        string   `json:"scenario_name"`
	ScenarioDescription string   `json:"scenario_description"`
	Probability         float64  `json:"probability"`
	Impact              string   `json:"impact"`
	PredictedOutcome    string   `json:"predicted_outcome"`
	RequiredActions     []string `json:"required_actions"`
}

type ModelAccuracy struct {
	OverallAccuracy   float64            `json:"overall_accuracy"`
	PredictionError   float64            `json:"prediction_error"`
	ConfidenceScore   float64            `json:"confidence_score"`
	ModelType         string             `json:"model_type"`
	LastValidation    time.Time          `json:"last_validation"`
	ValidationMetrics map[string]float64 `json:"validation_metrics"`
}

type QuotaComplianceRecommendations struct {
	EntityType              string                    `json:"entity_type"`
	EntityName              string                    `json:"entity_name"`
	OverallScore            float64                   `json:"overall_score"`
	ImprovementPotential    float64                   `json:"improvement_potential"`
	QuotaOptimization       []*QuotaOptimization      `json:"quota_optimization"`
	PolicyRecommendations   []*PolicyRecommendation   `json:"policy_recommendations"`
	ThresholdAdjustments    []*ThresholdAdjustment    `json:"threshold_adjustments"`
	ComplianceActions       []*ComplianceAction       `json:"compliance_actions"`
	RiskMitigation          []*RiskMitigationAction   `json:"risk_mitigation"`
	CostOptimization        []*CostOptimizationAction `json:"cost_optimization"`
	PerformanceActions      []*PerformanceAction      `json:"performance_actions"`
	AutomationOpportunities []*AutomationOpportunity  `json:"automation_opportunities"`
	PriorityMatrix          *PriorityMatrix           `json:"priority_matrix"`
	ImplementationPlan      *ImplementationPlan       `json:"implementation_plan"`
	ExpectedOutcomes        *ExpectedOutcomes         `json:"expected_outcomes"`
	LastGenerated           time.Time                 `json:"last_generated"`
}

type QuotaOptimization struct {
	ResourceType       string  `json:"resource_type"`
	CurrentQuota       float64 `json:"current_quota"`
	RecommendedQuota   float64 `json:"recommended_quota"`
	QuotaChange        float64 `json:"quota_change"`
	ChangeReason       string  `json:"change_reason"`
	ExpectedBenefit    string  `json:"expected_benefit"`
	ImplementationCost float64 `json:"implementation_cost"`
	ROI                float64 `json:"roi"`
	RiskLevel          string  `json:"risk_level"`
	Priority           string  `json:"priority"`
}

type QuotaPolicyRecommendation struct {
	PolicyName           string `json:"policy_name"`
	RecommendationType   string `json:"recommendation_type"`
	Description          string `json:"description"`
	CurrentState         string `json:"current_state"`
	RecommendedState     string `json:"recommended_state"`
	Justification        string `json:"justification"`
	ExpectedImpact       string `json:"expected_impact"`
	ImplementationEffort string `json:"implementation_effort"`
	Priority             string `json:"priority"`
}

type ThresholdAdjustment struct {
	ThresholdType        string  `json:"threshold_type"`
	ResourceType         string  `json:"resource_type"`
	CurrentThreshold     float64 `json:"current_threshold"`
	RecommendedThreshold float64 `json:"recommended_threshold"`
	AdjustmentReason     string  `json:"adjustment_reason"`
	ExpectedImprovement  string  `json:"expected_improvement"`
	ValidationPeriod     int     `json:"validation_period"`
	RollbackPlan         string  `json:"rollback_plan"`
}

type QuotaComplianceAction struct {
	ActionType        string   `json:"action_type"`
	Description       string   `json:"description"`
	TargetCompliance  float64  `json:"target_compliance"`
	CurrentGap        float64  `json:"current_gap"`
	RequiredResources []string `json:"required_resources"`
	Timeline          string   `json:"timeline"`
	SuccessMetrics    []string `json:"success_metrics"`
	DependsOn         []string `json:"depends_on"`
}

type RiskMitigationAction struct {
	RiskType           string  `json:"risk_type"`
	RiskLevel          string  `json:"risk_level"`
	MitigationStrategy string  `json:"mitigation_strategy"`
	Action             string  `json:"action"`
	EffectivenessScore float64 `json:"effectiveness_score"`
	ImplementationCost float64 `json:"implementation_cost"`
	Timeline           string  `json:"timeline"`
	ResponsibleParty   string  `json:"responsible_party"`
}

type CostOptimizationAction struct {
	OptimizationType   string  `json:"optimization_type"`
	CurrentCost        float64 `json:"current_cost"`
	PotentialSavings   float64 `json:"potential_savings"`
	SavingsPercentage  float64 `json:"savings_percentage"`
	Action             string  `json:"action"`
	ImplementationCost float64 `json:"implementation_cost"`
	PaybackPeriod      int     `json:"payback_period"`
	RiskLevel          string  `json:"risk_level"`
}

type PerformanceAction struct {
	PerformanceMetric   string   `json:"performance_metric"`
	CurrentPerformance  float64  `json:"current_performance"`
	TargetPerformance   float64  `json:"target_performance"`
	ImprovementAction   string   `json:"improvement_action"`
	ExpectedImprovement float64  `json:"expected_improvement"`
	ImplementationTime  int      `json:"implementation_time"`
	RequiredResources   []string `json:"required_resources"`
}

type AutomationOpportunity struct {
	ProcessName              string  `json:"process_name"`
	AutomationType           string  `json:"automation_type"`
	CurrentEffort            int     `json:"current_effort"`
	AutomatedEffort          int     `json:"automated_effort"`
	EffortSavings            int     `json:"effort_savings"`
	AccuracyImprovement      float64 `json:"accuracy_improvement"`
	ImplementationComplexity string  `json:"implementation_complexity"`
	ROI                      float64 `json:"roi"`
}

type PriorityMatrix struct {
	HighPriority    []string `json:"high_priority"`
	MediumPriority  []string `json:"medium_priority"`
	LowPriority     []string `json:"low_priority"`
	QuickWins       []string `json:"quick_wins"`
	LongTermGoals   []string `json:"long_term_goals"`
	PriorityFactors []string `json:"priority_factors"`
}

type QuotaImplementationPlan struct {
	Phase1Actions        []string       `json:"phase1_actions"`
	Phase2Actions        []string       `json:"phase2_actions"`
	Phase3Actions        []string       `json:"phase3_actions"`
	CriticalPath         []string       `json:"critical_path"`
	ResourceRequirements map[string]int `json:"resource_requirements"`
	Timeline             string         `json:"timeline"`
	Dependencies         []string       `json:"dependencies"`
	RiskFactors          []string       `json:"risk_factors"`
}

type ExpectedOutcomes struct {
	ComplianceImprovement  float64 `json:"compliance_improvement"`
	CostReduction          float64 `json:"cost_reduction"`
	PerformanceGain        float64 `json:"performance_gain"`
	RiskReduction          float64 `json:"risk_reduction"`
	EfficiencyGain         float64 `json:"efficiency_gain"`
	UserSatisfaction       float64 `json:"user_satisfaction"`
	OperationalImprovement float64 `json:"operational_improvement"`
}

type QuotaAuditEntry struct {
	EntryID           string                 `json:"entry_id"`
	Timestamp         time.Time              `json:"timestamp"`
	EntityType        string                 `json:"entity_type"`
	EntityName        string                 `json:"entity_name"`
	EventType         string                 `json:"event_type"`
	EventDescription  string                 `json:"event_description"`
	UserID            string                 `json:"user_id"`
	UserRole          string                 `json:"user_role"`
	SourceIP          string                 `json:"source_ip"`
	AffectedResources []string               `json:"affected_resources"`
	OldValues         map[string]interface{} `json:"old_values"`
	NewValues         map[string]interface{} `json:"new_values"`
	Success           bool                   `json:"success"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	SessionID         string                 `json:"session_id"`
	RequestID         string                 `json:"request_id"`
	Severity          string                 `json:"severity"`
	Category          string                 `json:"category"`
	Tags              []string               `json:"tags"`
}

type SystemQuotaOverview struct {
	TotalEntities         int                       `json:"total_entities"`
	CompliantEntities     int                       `json:"compliant_entities"`
	NonCompliantEntities  int                       `json:"non_compliant_entities"`
	SystemCompliance      float64                   `json:"system_compliance"`
	TotalViolations       int                       `json:"total_violations"`
	ActiveViolations      int                       `json:"active_violations"`
	TotalAlerts           int                       `json:"total_alerts"`
	CriticalAlerts        int                       `json:"critical_alerts"`
	ResourceUtilization   []*SystemResourceUsage    `json:"resource_utilization"`
	PolicyOverview        *PolicyOverview           `json:"policy_overview"`
	ComplianceTrends      *SystemComplianceTrends   `json:"compliance_trends"`
	RiskAssessment        *SystemRiskAssessment     `json:"risk_assessment"`
	PerformanceMetrics    *SystemPerformanceMetrics `json:"performance_metrics"`
	CapacityStatus        *SystemCapacityStatus     `json:"capacity_status"`
	CostAnalysis          *SystemCostAnalysis       `json:"cost_analysis"`
	RecommendationSummary *RecommendationSummary    `json:"recommendation_summary"`
	LastUpdated           time.Time                 `json:"last_updated"`
}

type SystemResourceUsage struct {
	ResourceType        string     `json:"resource_type"`
	TotalQuota          float64    `json:"total_quota"`
	UsedQuota           float64    `json:"used_quota"`
	AvailableQuota      float64    `json:"available_quota"`
	UtilizationRate     float64    `json:"utilization_rate"`
	TrendDirection      string     `json:"trend_direction"`
	ProjectedExhaustion *time.Time `json:"projected_exhaustion,omitempty"`
}

type PolicyOverview struct {
	TotalPolicies       int     `json:"total_policies"`
	ActivePolicies      int     `json:"active_policies"`
	PolicyCompliance    float64 `json:"policy_compliance"`
	PolicyViolations    int     `json:"policy_violations"`
	PolicyExceptions    int     `json:"policy_exceptions"`
	PolicyUpdatesNeeded int     `json:"policy_updates_needed"`
}

type SystemComplianceTrends struct {
	ComplianceDirection string  `json:"compliance_direction"`
	ComplianceVelocity  float64 `json:"compliance_velocity"`
	ImprovementRate     float64 `json:"improvement_rate"`
	DeteriorationRate   float64 `json:"deterioration_rate"`
	TrendStability      float64 `json:"trend_stability"`
	ForecastAccuracy    float64 `json:"forecast_accuracy"`
}

type SystemRiskAssessment struct {
	OverallRiskLevel   string   `json:"overall_risk_level"`
	RiskScore          float64  `json:"risk_score"`
	HighRiskEntities   int      `json:"high_risk_entities"`
	MediumRiskEntities int      `json:"medium_risk_entities"`
	LowRiskEntities    int      `json:"low_risk_entities"`
	EmergingRisks      []string `json:"emerging_risks"`
	TopRiskFactors     []string `json:"top_risk_factors"`
}

type QuotaSystemPerformanceMetrics struct {
	ResponseTime         float64 `json:"response_time"`
	Throughput           float64 `json:"throughput"`
	ErrorRate            float64 `json:"error_rate"`
	Availability         float64 `json:"availability"`
	ProcessingEfficiency float64 `json:"processing_efficiency"`
	SystemLoad           float64 `json:"system_load"`
}

type SystemCapacityStatus struct {
	OverallCapacity      float64    `json:"overall_capacity"`
	UsedCapacity         float64    `json:"used_capacity"`
	AvailableCapacity    float64    `json:"available_capacity"`
	CapacityGrowthRate   float64    `json:"capacity_growth_rate"`
	TimeToCapacity       *time.Time `json:"time_to_capacity,omitempty"`
	ExpansionRecommended bool       `json:"expansion_recommended"`
}

type SystemCostAnalysis struct {
	TotalCost                 float64  `json:"total_cost"`
	CostPerEntity             float64  `json:"cost_per_entity"`
	CostTrend                 string   `json:"cost_trend"`
	CostOptimizationPotential float64  `json:"cost_optimization_potential"`
	WastedResources           float64  `json:"wasted_resources"`
	EfficiencyOpportunities   []string `json:"efficiency_opportunities"`
}

type RecommendationSummary struct {
	TotalRecommendations int     `json:"total_recommendations"`
	HighPriorityActions  int     `json:"high_priority_actions"`
	QuickWins            int     `json:"quick_wins"`
	LongTermProjects     int     `json:"long_term_projects"`
	EstimatedSavings     float64 `json:"estimated_savings"`
	ImplementationEffort int     `json:"implementation_effort"`
}

type QuotaNotificationSettings struct {
	EntityType          string                 `json:"entity_type"`
	EntityName          string                 `json:"entity_name"`
	NotificationRules   []*NotificationRule    `json:"notification_rules"`
	AlertChannels       []*AlertChannel        `json:"alert_channels"`
	EscalationMatrix    []*EscalationRule      `json:"escalation_matrix"`
	QuietHours          []*QuietPeriod         `json:"quiet_hours"`
	FrequencyLimits     []*FrequencyLimit      `json:"frequency_limits"`
	ContentTemplates    []*ContentTemplate     `json:"content_templates"`
	DeliveryPreferences *DeliveryPreferences   `json:"delivery_preferences"`
	SubscriptionStatus  *SubscriptionStatus    `json:"subscription_status"`
	NotificationHistory []*NotificationHistory `json:"notification_history"`
	LastUpdated         time.Time              `json:"last_updated"`
}

type AlertChannel struct {
	ChannelID       string     `json:"channel_id"`
	ChannelType     string     `json:"channel_type"`
	ChannelName     string     `json:"channel_name"`
	Endpoint        string     `json:"endpoint"`
	IsActive        bool       `json:"is_active"`
	Priority        int        `json:"priority"`
	SupportedEvents []string   `json:"supported_events"`
	FailureHandling string     `json:"failure_handling"`
	RetryPolicy     string     `json:"retry_policy"`
	LastUsed        *time.Time `json:"last_used,omitempty"`
}

type QuotaEscalationRule struct {
	RuleID             string   `json:"rule_id"`
	TriggerCondition   string   `json:"trigger_condition"`
	EscalationLevel    int      `json:"escalation_level"`
	DelayMinutes       int      `json:"delay_minutes"`
	TargetRecipients   []string `json:"target_recipients"`
	NotificationMethod string   `json:"notification_method"`
	IsActive           bool     `json:"is_active"`
	MaxEscalations     int      `json:"max_escalations"`
}

type QuietPeriod struct {
	PeriodID        string   `json:"period_id"`
	StartTime       string   `json:"start_time"`
	EndTime         string   `json:"end_time"`
	DaysOfWeek      []int    `json:"days_of_week"`
	Timezone        string   `json:"timezone"`
	ExceptionEvents []string `json:"exception_events"`
	IsActive        bool     `json:"is_active"`
}

type FrequencyLimit struct {
	LimitType          string   `json:"limit_type"`
	MaxNotifications   int      `json:"max_notifications"`
	TimeWindow         int      `json:"time_window"`
	ApplicableEvents   []string `json:"applicable_events"`
	OverrideConditions []string `json:"override_conditions"`
	IsStrict           bool     `json:"is_strict"`
}

type ContentTemplate struct {
	TemplateID   string            `json:"template_id"`
	TemplateName string            `json:"template_name"`
	EventType    string            `json:"event_type"`
	Subject      string            `json:"subject"`
	Body         string            `json:"body"`
	Format       string            `json:"format"`
	Variables    []string          `json:"variables"`
	Localization map[string]string `json:"localization"`
	IsDefault    bool              `json:"is_default"`
}

type DeliveryPreferences struct {
	PreferredChannels    []string `json:"preferred_channels"`
	BackupChannels       []string `json:"backup_channels"`
	DeliveryMode         string   `json:"delivery_mode"`
	BatchingEnabled      bool     `json:"batching_enabled"`
	BatchingWindow       int      `json:"batching_window"`
	DeduplicationEnabled bool     `json:"deduplication_enabled"`
	DeduplicationWindow  int      `json:"deduplication_window"`
}

type SubscriptionStatus struct {
	IsSubscribed      bool      `json:"is_subscribed"`
	SubscriptionLevel string    `json:"subscription_level"`
	SubscribedEvents  []string  `json:"subscribed_events"`
	ExcludedEvents    []string  `json:"excluded_events"`
	SubscriptionDate  time.Time `json:"subscription_date"`
	LastModified      time.Time `json:"last_modified"`
	OptOutReason      string    `json:"opt_out_reason,omitempty"`
}

type NotificationHistory struct {
	NotificationID string     `json:"notification_id"`
	Timestamp      time.Time  `json:"timestamp"`
	EventType      string     `json:"event_type"`
	Channel        string     `json:"channel"`
	Recipient      string     `json:"recipient"`
	Status         string     `json:"status"`
	DeliveryTime   *time.Time `json:"delivery_time,omitempty"`
	ReadTime       *time.Time `json:"read_time,omitempty"`
	ResponseTime   *time.Time `json:"response_time,omitempty"`
	ErrorDetails   string     `json:"error_details,omitempty"`
	RetryCount     int        `json:"retry_count"`
}

type QuotaComplianceCollector struct {
	client QuotaComplianceSLURMClient
	mutex  sync.RWMutex

	// Compliance Metrics
	quotaComplianceScore    *prometheus.GaugeVec
	quotaComplianceStatus   *prometheus.GaugeVec
	quotaResourceCompliance *prometheus.GaugeVec
	quotaPolicyCompliance   *prometheus.GaugeVec
	quotaViolationsSummary  *prometheus.GaugeVec
	quotaComplianceHistory  *prometheus.GaugeVec

	// Violation Metrics
	quotaViolationsTotal          *prometheus.CounterVec
	quotaViolationsSeverity       *prometheus.GaugeVec
	quotaViolationsResolutionTime *prometheus.HistogramVec
	quotaViolationsRecurrence     *prometheus.GaugeVec
	quotaViolationsExcess         *prometheus.GaugeVec

	// Alert Metrics
	quotaAlertsTotal      *prometheus.GaugeVec
	quotaAlertsActive     *prometheus.GaugeVec
	quotaAlertsPriority   *prometheus.GaugeVec
	quotaAlertsEscalation *prometheus.GaugeVec
	quotaAlertsResponse   *prometheus.HistogramVec

	// Threshold Metrics
	quotaThresholdBreaches      *prometheus.CounterVec
	quotaThresholdEffectiveness *prometheus.GaugeVec
	quotaThresholdUtilization   *prometheus.GaugeVec
	quotaThresholdAdjustments   *prometheus.CounterVec

	// Policy Metrics
	quotaPolicyAdherence     *prometheus.GaugeVec
	quotaPolicyViolations    *prometheus.CounterVec
	quotaPolicyExceptions    *prometheus.GaugeVec
	quotaPolicyEffectiveness *prometheus.GaugeVec

	// Enforcement Metrics
	quotaEnforcementActions       *prometheus.CounterVec
	quotaEnforcementEffectiveness *prometheus.GaugeVec
	quotaEnforcementOverrides     *prometheus.CounterVec
	quotaEnforcementExceptions    *prometheus.GaugeVec

	// Trend Metrics
	quotaTrendDirection   *prometheus.GaugeVec
	quotaTrendGrowthRate  *prometheus.GaugeVec
	quotaTrendVolatility  *prometheus.GaugeVec
	quotaTrendSeasonality *prometheus.GaugeVec
	quotaTrendAnomalies   *prometheus.GaugeVec

	// Forecast Metrics
	quotaForecastAccuracy   *prometheus.GaugeVec
	quotaForecastConfidence *prometheus.GaugeVec
	quotaForecastRisk       *prometheus.GaugeVec
	quotaForecastExhaustion *prometheus.GaugeVec

	// Recommendation Metrics
	quotaRecommendationsTotal          *prometheus.GaugeVec
	quotaRecommendationsPriority       *prometheus.GaugeVec
	quotaRecommendationsROI            *prometheus.GaugeVec
	quotaRecommendationsImplementation *prometheus.GaugeVec

	// Audit Metrics
	quotaAuditEvents   *prometheus.CounterVec
	quotaAuditFailures *prometheus.CounterVec
	quotaAuditLatency  *prometheus.HistogramVec

	// System Overview Metrics
	systemQuotaCompliance  prometheus.Gauge
	systemQuotaEntities    *prometheus.GaugeVec
	systemQuotaViolations  prometheus.Gauge
	systemQuotaAlerts      prometheus.Gauge
	systemQuotaUtilization *prometheus.GaugeVec
	systemQuotaRisk        prometheus.Gauge
	systemQuotaPerformance *prometheus.GaugeVec
	systemQuotaCost        *prometheus.GaugeVec

	// Notification Metrics
	quotaNotificationsSent      *prometheus.CounterVec
	quotaNotificationsDelivered *prometheus.CounterVec
	quotaNotificationsFailed    *prometheus.CounterVec
	quotaNotificationsLatency   *prometheus.HistogramVec
}

func NewQuotaComplianceCollector(client QuotaComplianceSLURMClient) *QuotaComplianceCollector {
	return &QuotaComplianceCollector{
		client: client,

		// Compliance Metrics
		quotaComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_compliance_score",
				Help: "Overall compliance score for entity quotas",
			},
			[]string{"entity_type", "entity_name"},
		),
		quotaComplianceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_compliance_status",
				Help: "Compliance status for entity quotas (0=non-compliant, 1=compliant)",
			},
			[]string{"entity_type", "entity_name", "status"},
		),
		quotaResourceCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_resource_compliance",
				Help: "Resource-specific compliance metrics",
			},
			[]string{"entity_type", "entity_name", "resource_type", "metric"},
		),
		quotaPolicyCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_policy_compliance",
				Help: "Policy-specific compliance metrics",
			},
			[]string{"entity_type", "entity_name", "policy_name", "policy_type"},
		),
		quotaViolationsSummary: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_violations_summary",
				Help: "Summary of quota violations",
			},
			[]string{"entity_type", "entity_name", "violation_category"},
		),
		quotaComplianceHistory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_compliance_history",
				Help: "Historical compliance trends",
			},
			[]string{"entity_type", "entity_name", "time_period"},
		),

		// Violation Metrics
		quotaViolationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_violations_total",
				Help: "Total number of quota violations",
			},
			[]string{"entity_type", "entity_name", "violation_type", "severity"},
		),
		quotaViolationsSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_violations_severity",
				Help: "Severity distribution of quota violations",
			},
			[]string{"entity_type", "entity_name", "severity"},
		),
		quotaViolationsResolutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_quota_violations_resolution_time_seconds",
				Help:    "Time taken to resolve quota violations",
				Buckets: prometheus.ExponentialBuckets(60, 2, 10),
			},
			[]string{"entity_type", "entity_name", "violation_type"},
		),
		quotaViolationsRecurrence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_violations_recurrence",
				Help: "Recurrence rate of quota violations",
			},
			[]string{"entity_type", "entity_name", "violation_type"},
		),
		quotaViolationsExcess: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_violations_excess",
				Help: "Amount of excess usage in violations",
			},
			[]string{"entity_type", "entity_name", "resource_type", "metric"},
		),

		// Alert Metrics
		quotaAlertsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_alerts_total",
				Help: "Total number of quota alerts",
			},
			[]string{"entity_type", "entity_name", "alert_type"},
		),
		quotaAlertsActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_alerts_active",
				Help: "Number of active quota alerts",
			},
			[]string{"entity_type", "entity_name", "priority"},
		),
		quotaAlertsPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_alerts_priority",
				Help: "Distribution of alerts by priority",
			},
			[]string{"entity_type", "entity_name", "priority"},
		),
		quotaAlertsEscalation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_alerts_escalation",
				Help: "Alert escalation levels",
			},
			[]string{"entity_type", "entity_name", "escalation_level"},
		),
		quotaAlertsResponse: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_quota_alerts_response_time_seconds",
				Help:    "Time taken to respond to quota alerts",
				Buckets: prometheus.ExponentialBuckets(60, 2, 8),
			},
			[]string{"entity_type", "entity_name", "alert_type"},
		),

		// Threshold Metrics
		quotaThresholdBreaches: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_threshold_breaches_total",
				Help: "Total number of threshold breaches",
			},
			[]string{"entity_type", "entity_name", "resource_type", "threshold_type"},
		),
		quotaThresholdEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_threshold_effectiveness",
				Help: "Effectiveness score of quota thresholds",
			},
			[]string{"entity_type", "entity_name", "threshold_type"},
		),
		quotaThresholdUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_threshold_utilization",
				Help: "Current utilization relative to thresholds",
			},
			[]string{"entity_type", "entity_name", "resource_type", "threshold_level"},
		),
		quotaThresholdAdjustments: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_threshold_adjustments_total",
				Help: "Total number of threshold adjustments",
			},
			[]string{"entity_type", "entity_name", "threshold_type", "adjustment_type"},
		),

		// Policy Metrics
		quotaPolicyAdherence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_policy_adherence",
				Help: "Policy adherence rates",
			},
			[]string{"entity_type", "entity_name", "policy_name"},
		),
		quotaPolicyViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_policy_violations_total",
				Help: "Total number of policy violations",
			},
			[]string{"entity_type", "entity_name", "policy_name", "violation_type"},
		),
		quotaPolicyExceptions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_policy_exceptions",
				Help: "Number of active policy exceptions",
			},
			[]string{"entity_type", "entity_name", "policy_name", "exception_type"},
		),
		quotaPolicyEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_policy_effectiveness",
				Help: "Policy effectiveness scores",
			},
			[]string{"entity_type", "entity_name", "policy_name"},
		),

		// Enforcement Metrics
		quotaEnforcementActions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_enforcement_actions_total",
				Help: "Total number of enforcement actions",
			},
			[]string{"entity_type", "entity_name", "action_type", "result"},
		),
		quotaEnforcementEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_enforcement_effectiveness",
				Help: "Enforcement effectiveness scores",
			},
			[]string{"entity_type", "entity_name", "enforcement_mode"},
		),
		quotaEnforcementOverrides: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_enforcement_overrides_total",
				Help: "Total number of enforcement overrides",
			},
			[]string{"entity_type", "entity_name", "override_type", "authorization"},
		),
		quotaEnforcementExceptions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_enforcement_exceptions",
				Help: "Number of active enforcement exceptions",
			},
			[]string{"entity_type", "entity_name", "exception_type"},
		),

		// Trend Metrics
		quotaTrendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_trend_direction",
				Help: "Trend direction for quota metrics (-1=decreasing, 0=stable, 1=increasing)",
			},
			[]string{"entity_type", "entity_name", "metric_type"},
		),
		quotaTrendGrowthRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_trend_growth_rate",
				Help: "Growth rate for quota metrics",
			},
			[]string{"entity_type", "entity_name", "metric_type"},
		),
		quotaTrendVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_trend_volatility",
				Help: "Volatility measure for quota trends",
			},
			[]string{"entity_type", "entity_name", "metric_type"},
		),
		quotaTrendSeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_trend_seasonality",
				Help: "Seasonality strength for quota patterns",
			},
			[]string{"entity_type", "entity_name", "pattern_type"},
		),
		quotaTrendAnomalies: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_trend_anomalies",
				Help: "Number of anomalies detected in quota trends",
			},
			[]string{"entity_type", "entity_name", "anomaly_type"},
		),

		// Forecast Metrics
		quotaForecastAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_forecast_accuracy",
				Help: "Accuracy of quota forecasting models",
			},
			[]string{"entity_type", "entity_name", "forecast_type", "horizon"},
		),
		quotaForecastConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_forecast_confidence",
				Help: "Confidence level in quota forecasts",
			},
			[]string{"entity_type", "entity_name", "forecast_type"},
		),
		quotaForecastRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_forecast_risk",
				Help: "Risk level in quota forecasts",
			},
			[]string{"entity_type", "entity_name", "risk_type"},
		),
		quotaForecastExhaustion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_forecast_exhaustion_days",
				Help: "Days until predicted quota exhaustion",
			},
			[]string{"entity_type", "entity_name", "resource_type"},
		),

		// Recommendation Metrics
		quotaRecommendationsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_recommendations_total",
				Help: "Total number of quota recommendations",
			},
			[]string{"entity_type", "entity_name", "recommendation_type"},
		),
		quotaRecommendationsPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_recommendations_priority",
				Help: "Distribution of recommendations by priority",
			},
			[]string{"entity_type", "entity_name", "priority"},
		),
		quotaRecommendationsROI: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_recommendations_roi",
				Help: "Expected ROI from quota recommendations",
			},
			[]string{"entity_type", "entity_name", "recommendation_type"},
		),
		quotaRecommendationsImplementation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_quota_recommendations_implementation_effort",
				Help: "Implementation effort for quota recommendations",
			},
			[]string{"entity_type", "entity_name", "recommendation_type"},
		),

		// Audit Metrics
		quotaAuditEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_audit_events_total",
				Help: "Total number of quota audit events",
			},
			[]string{"entity_type", "entity_name", "event_type", "severity"},
		),
		quotaAuditFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_audit_failures_total",
				Help: "Total number of quota audit failures",
			},
			[]string{"entity_type", "entity_name", "event_type"},
		),
		quotaAuditLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_quota_audit_latency_seconds",
				Help:    "Latency of quota audit operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"entity_type", "entity_name", "operation"},
		),

		// System Overview Metrics
		systemQuotaCompliance: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_compliance",
				Help: "Overall system quota compliance score",
			},
		),
		systemQuotaEntities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_entities",
				Help: "Number of quota entities by status",
			},
			[]string{"entity_type", "status"},
		),
		systemQuotaViolations: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_violations",
				Help: "Total active quota violations in system",
			},
		),
		systemQuotaAlerts: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_alerts",
				Help: "Total active quota alerts in system",
			},
		),
		systemQuotaUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_utilization",
				Help: "System-wide quota utilization by resource type",
			},
			[]string{"resource_type"},
		),
		systemQuotaRisk: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_risk",
				Help: "Overall system quota risk score",
			},
		),
		systemQuotaPerformance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_performance",
				Help: "System quota management performance metrics",
			},
			[]string{"metric"},
		),
		systemQuotaCost: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_quota_cost",
				Help: "System quota cost analysis metrics",
			},
			[]string{"cost_type"},
		),

		// Notification Metrics
		quotaNotificationsSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_notifications_sent_total",
				Help: "Total number of quota notifications sent",
			},
			[]string{"entity_type", "entity_name", "notification_type", "channel"},
		),
		quotaNotificationsDelivered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_notifications_delivered_total",
				Help: "Total number of quota notifications delivered",
			},
			[]string{"entity_type", "entity_name", "notification_type", "channel"},
		),
		quotaNotificationsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_quota_notifications_failed_total",
				Help: "Total number of quota notification failures",
			},
			[]string{"entity_type", "entity_name", "notification_type", "channel", "failure_reason"},
		),
		quotaNotificationsLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_quota_notifications_latency_seconds",
				Help:    "Latency of quota notification delivery",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"entity_type", "entity_name", "notification_type", "channel"},
		),
	}
}

func (c *QuotaComplianceCollector) Describe(ch chan<- *prometheus.Desc) {
	c.quotaComplianceScore.Describe(ch)
	c.quotaComplianceStatus.Describe(ch)
	c.quotaResourceCompliance.Describe(ch)
	c.quotaPolicyCompliance.Describe(ch)
	c.quotaViolationsSummary.Describe(ch)
	c.quotaComplianceHistory.Describe(ch)
	c.quotaViolationsTotal.Describe(ch)
	c.quotaViolationsSeverity.Describe(ch)
	c.quotaViolationsResolutionTime.Describe(ch)
	c.quotaViolationsRecurrence.Describe(ch)
	c.quotaViolationsExcess.Describe(ch)
	c.quotaAlertsTotal.Describe(ch)
	c.quotaAlertsActive.Describe(ch)
	c.quotaAlertsPriority.Describe(ch)
	c.quotaAlertsEscalation.Describe(ch)
	c.quotaAlertsResponse.Describe(ch)
	c.quotaThresholdBreaches.Describe(ch)
	c.quotaThresholdEffectiveness.Describe(ch)
	c.quotaThresholdUtilization.Describe(ch)
	c.quotaThresholdAdjustments.Describe(ch)
	c.quotaPolicyAdherence.Describe(ch)
	c.quotaPolicyViolations.Describe(ch)
	c.quotaPolicyExceptions.Describe(ch)
	c.quotaPolicyEffectiveness.Describe(ch)
	c.quotaEnforcementActions.Describe(ch)
	c.quotaEnforcementEffectiveness.Describe(ch)
	c.quotaEnforcementOverrides.Describe(ch)
	c.quotaEnforcementExceptions.Describe(ch)
	c.quotaTrendDirection.Describe(ch)
	c.quotaTrendGrowthRate.Describe(ch)
	c.quotaTrendVolatility.Describe(ch)
	c.quotaTrendSeasonality.Describe(ch)
	c.quotaTrendAnomalies.Describe(ch)
	c.quotaForecastAccuracy.Describe(ch)
	c.quotaForecastConfidence.Describe(ch)
	c.quotaForecastRisk.Describe(ch)
	c.quotaForecastExhaustion.Describe(ch)
	c.quotaRecommendationsTotal.Describe(ch)
	c.quotaRecommendationsPriority.Describe(ch)
	c.quotaRecommendationsROI.Describe(ch)
	c.quotaRecommendationsImplementation.Describe(ch)
	c.quotaAuditEvents.Describe(ch)
	c.quotaAuditFailures.Describe(ch)
	c.quotaAuditLatency.Describe(ch)
	c.systemQuotaCompliance.Describe(ch)
	c.systemQuotaEntities.Describe(ch)
	c.systemQuotaViolations.Describe(ch)
	c.systemQuotaAlerts.Describe(ch)
	c.systemQuotaUtilization.Describe(ch)
	c.systemQuotaRisk.Describe(ch)
	c.systemQuotaPerformance.Describe(ch)
	c.systemQuotaCost.Describe(ch)
	c.quotaNotificationsSent.Describe(ch)
	c.quotaNotificationsDelivered.Describe(ch)
	c.quotaNotificationsFailed.Describe(ch)
	c.quotaNotificationsLatency.Describe(ch)
}

func (c *QuotaComplianceCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()

	// Collect system overview first
	c.collectSystemOverview(ch)

	// Collect for each entity type
	entityTypes := []string{"user", "account", "partition", "qos"}
	for _, entityType := range entityTypes {
		entities, err := c.client.ListQuotaEntities(ctx, entityType)
		if err != nil {
			continue
		}

		for _, entity := range entities {
			c.collectEntityMetrics(ch, entity.EntityType, entity.EntityName)
		}
	}
}

func (c *QuotaComplianceCollector) collectSystemOverview(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	overview, err := c.client.GetSystemQuotaOverview(ctx)
	if err != nil {
		return
	}

	c.systemQuotaCompliance.Set(overview.SystemCompliance)
	c.systemQuotaViolations.Set(float64(overview.ActiveViolations))
	c.systemQuotaAlerts.Set(float64(overview.CriticalAlerts))

	ch <- prometheus.MustNewConstMetric(c.systemQuotaCompliance.Desc(), prometheus.GaugeValue, overview.SystemCompliance)
	ch <- prometheus.MustNewConstMetric(c.systemQuotaViolations.Desc(), prometheus.GaugeValue, float64(overview.ActiveViolations))
	ch <- prometheus.MustNewConstMetric(c.systemQuotaAlerts.Desc(), prometheus.GaugeValue, float64(overview.CriticalAlerts))

	ch <- prometheus.MustNewConstMetric(
		c.systemQuotaEntities.WithLabelValues("all", "total").Desc(),
		prometheus.GaugeValue,
		float64(overview.TotalEntities),
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemQuotaEntities.WithLabelValues("all", "compliant").Desc(),
		prometheus.GaugeValue,
		float64(overview.CompliantEntities),
	)
	ch <- prometheus.MustNewConstMetric(
		c.systemQuotaEntities.WithLabelValues("all", "non_compliant").Desc(),
		prometheus.GaugeValue,
		float64(overview.NonCompliantEntities),
	)

	if overview.ResourceUtilization != nil {
		for _, usage := range overview.ResourceUtilization {
			ch <- prometheus.MustNewConstMetric(
				c.systemQuotaUtilization.WithLabelValues(usage.ResourceType).Desc(),
				prometheus.GaugeValue,
				usage.UtilizationRate,
			)
		}
	}

	if overview.RiskAssessment != nil {
		c.systemQuotaRisk.Set(overview.RiskAssessment.RiskScore)
		ch <- prometheus.MustNewConstMetric(c.systemQuotaRisk.Desc(), prometheus.GaugeValue, overview.RiskAssessment.RiskScore)
	}

	if overview.PerformanceMetrics != nil {
		// Map available metrics to the expected labels
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaPerformance.WithLabelValues("response_time").Desc(),
			prometheus.GaugeValue,
			overview.PerformanceMetrics.QueueMetrics.AverageWaitTime.Seconds(), // Use queue wait time as proxy
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaPerformance.WithLabelValues("throughput").Desc(),
			prometheus.GaugeValue,
			overview.PerformanceMetrics.ThroughputMetrics.JobsPerHour,
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaPerformance.WithLabelValues("availability").Desc(),
			prometheus.GaugeValue,
			overview.PerformanceMetrics.EfficiencyMetrics.NodeUtilization, // Use node utilization as proxy for availability
		)
	}

	if overview.CostAnalysis != nil {
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaCost.WithLabelValues("total").Desc(),
			prometheus.GaugeValue,
			overview.CostAnalysis.TotalCost,
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaCost.WithLabelValues("optimization_potential").Desc(),
			prometheus.GaugeValue,
			overview.CostAnalysis.CostOptimizationPotential,
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemQuotaCost.WithLabelValues("wasted_resources").Desc(),
			prometheus.GaugeValue,
			overview.CostAnalysis.WastedResources,
		)
	}
}

func (c *QuotaComplianceCollector) collectEntityMetrics(ch chan<- prometheus.Metric, entityType, entityName string) {
	ctx := context.Background()

	// Collect compliance metrics
	if compliance, err := c.client.GetQuotaCompliance(ctx, entityType, entityName); err == nil {
		c.collectComplianceMetrics(ch, entityType, entityName, compliance)
	}

	// Collect violations
	if violations, err := c.client.GetQuotaViolations(ctx, entityType, entityName); err == nil {
		c.collectViolationMetrics(ch, entityType, entityName, violations)
	}

	// Collect alerts
	if alerts, err := c.client.GetQuotaAlerts(ctx, entityType, entityName); err == nil {
		c.collectAlertMetrics(ch, entityType, entityName, alerts)
	}

	// Collect enforcement metrics
	if enforcement, err := c.client.GetQuotaEnforcement(ctx, entityType, entityName); err == nil {
		c.collectEnforcementMetrics(ch, entityType, entityName, enforcement)
	}

	// Collect trend metrics
	if trends, err := c.client.GetQuotaTrends(ctx, entityType, entityName); err == nil {
		c.collectTrendMetrics(ch, entityType, entityName, trends)
	}

	// Collect forecast metrics
	if forecast, err := c.client.GetQuotaForecast(ctx, entityType, entityName); err == nil {
		c.collectForecastMetrics(ch, entityType, entityName, forecast)
	}

	// Collect recommendations
	if recommendations, err := c.client.GetQuotaRecommendations(ctx, entityType, entityName); err == nil {
		c.collectRecommendationMetrics(ch, entityType, entityName, recommendations)
	}

	// Collect audit events
	if auditEntries, err := c.client.GetQuotaAuditLog(ctx, entityType, entityName); err == nil {
		c.collectAuditMetrics(ch, entityType, entityName, auditEntries)
	}

	// Collect notification metrics
	if notificationSettings, err := c.client.GetQuotaNotificationSettings(ctx, entityType, entityName); err == nil {
		c.collectNotificationMetrics(ch, entityType, entityName, notificationSettings)
	}
}

func (c *QuotaComplianceCollector) collectComplianceMetrics(ch chan<- prometheus.Metric, entityType, entityName string, compliance *QuotaComplianceReport) {
	ch <- prometheus.MustNewConstMetric(
		c.quotaComplianceScore.WithLabelValues(entityType, entityName).Desc(),
		prometheus.GaugeValue,
		compliance.ComplianceScore,
	)

	statusValue := 0.0
	if compliance.ComplianceStatus == "COMPLIANT" {
		statusValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(
		c.quotaComplianceStatus.WithLabelValues(entityType, entityName, compliance.ComplianceStatus).Desc(),
		prometheus.GaugeValue,
		statusValue,
	)

	// Resource compliance metrics
	for _, rc := range compliance.ResourceCompliance {
		ch <- prometheus.MustNewConstMetric(
			c.quotaResourceCompliance.WithLabelValues(entityType, entityName, rc.ResourceType, "utilization_rate").Desc(),
			prometheus.GaugeValue,
			rc.UtilizationRate,
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaResourceCompliance.WithLabelValues(entityType, entityName, rc.ResourceType, "compliance_score").Desc(),
			prometheus.GaugeValue,
			rc.ComplianceScore,
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaResourceCompliance.WithLabelValues(entityType, entityName, rc.ResourceType, "threshold_breaches").Desc(),
			prometheus.GaugeValue,
			float64(rc.ThresholdBreaches),
		)
	}

	// Policy compliance metrics
	for _, pc := range compliance.PolicyCompliance {
		ch <- prometheus.MustNewConstMetric(
			c.quotaPolicyCompliance.WithLabelValues(entityType, entityName, pc.PolicyName, pc.PolicyType).Desc(),
			prometheus.GaugeValue,
			pc.ComplianceScore,
		)
	}

	// Violation summary metrics
	if vs := compliance.ViolationSummary; vs != nil {
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsSummary.WithLabelValues(entityType, entityName, "total").Desc(),
			prometheus.GaugeValue,
			float64(vs.TotalViolations),
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsSummary.WithLabelValues(entityType, entityName, "active").Desc(),
			prometheus.GaugeValue,
			float64(vs.ActiveViolations),
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsSummary.WithLabelValues(entityType, entityName, "critical").Desc(),
			prometheus.GaugeValue,
			float64(vs.CriticalViolations),
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsSummary.WithLabelValues(entityType, entityName, "violation_rate").Desc(),
			prometheus.GaugeValue,
			vs.ViolationRate,
		)
	}
}

func (c *QuotaComplianceCollector) collectViolationMetrics(ch chan<- prometheus.Metric, entityType, entityName string, violations []*QuotaComplianceViolation) {
	for _, violation := range violations {
		// Count violations by type and severity
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsTotal.WithLabelValues(entityType, entityName, violation.ViolationType, violation.Severity).Desc(),
			prometheus.CounterValue,
			1,
		)

		// Track excess amounts
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsExcess.WithLabelValues(entityType, entityName, violation.ResourceType, "absolute").Desc(),
			prometheus.GaugeValue,
			violation.ExcessAmount,
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsExcess.WithLabelValues(entityType, entityName, violation.ResourceType, "percentage").Desc(),
			prometheus.GaugeValue,
			violation.ExcessPercentage,
		)

		// Track recurrence
		if violation.IsRecurrent {
			ch <- prometheus.MustNewConstMetric(
				c.quotaViolationsRecurrence.WithLabelValues(entityType, entityName, violation.ViolationType).Desc(),
				prometheus.GaugeValue,
				float64(violation.RecurrenceCount),
			)
		}

		// Track resolution time for resolved violations
		if violation.ResolvedAt != nil {
			resolutionTime := violation.ResolvedAt.Sub(violation.DetectedAt).Seconds()
			c.quotaViolationsResolutionTime.WithLabelValues(entityType, entityName, violation.ViolationType).Observe(resolutionTime)
		}
	}

	// Count violations by severity
	severityCounts := make(map[string]int)
	for _, violation := range violations {
		severityCounts[violation.Severity]++
	}
	for severity, count := range severityCounts {
		ch <- prometheus.MustNewConstMetric(
			c.quotaViolationsSeverity.WithLabelValues(entityType, entityName, severity).Desc(),
			prometheus.GaugeValue,
			float64(count),
		)
	}
}

func (c *QuotaComplianceCollector) collectAlertMetrics(ch chan<- prometheus.Metric, entityType, entityName string, alerts []*QuotaComplianceAlert) {
	alertCounts := make(map[string]int)
	priorityCounts := make(map[string]int)
	activeAlerts := 0

	for _, alert := range alerts {
		alertCounts[alert.AlertType]++
		priorityCounts[alert.Priority]++

		if alert.Status == "ACTIVE" {
			activeAlerts++
		}

		// Track escalation levels
		ch <- prometheus.MustNewConstMetric(
			c.quotaAlertsEscalation.WithLabelValues(entityType, entityName, string(rune(alert.EscalationLevel))).Desc(),
			prometheus.GaugeValue,
			float64(alert.EscalationLevel),
		)

		// Track response times for acknowledged alerts
		if alert.AcknowledgedAt != nil {
			responseTime := alert.AcknowledgedAt.Sub(alert.TriggeredAt).Seconds()
			c.quotaAlertsResponse.WithLabelValues(entityType, entityName, alert.AlertType).Observe(responseTime)
		}
	}

	// Report alert counts by type
	for alertType, count := range alertCounts {
		ch <- prometheus.MustNewConstMetric(
			c.quotaAlertsTotal.WithLabelValues(entityType, entityName, alertType).Desc(),
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Report alert counts by priority
	for priority, count := range priorityCounts {
		ch <- prometheus.MustNewConstMetric(
			c.quotaAlertsPriority.WithLabelValues(entityType, entityName, priority).Desc(),
			prometheus.GaugeValue,
			float64(count),
		)
	}

	// Report active alerts
	ch <- prometheus.MustNewConstMetric(
		c.quotaAlertsActive.WithLabelValues(entityType, entityName, "all").Desc(),
		prometheus.GaugeValue,
		float64(activeAlerts),
	)
}

func (c *QuotaComplianceCollector) collectEnforcementMetrics(ch chan<- prometheus.Metric, entityType, entityName string, enforcement *QuotaEnforcement) {
	// Enforcement effectiveness
	if enforcement.ComplianceMetrics != nil {
		ch <- prometheus.MustNewConstMetric(
			c.quotaEnforcementEffectiveness.WithLabelValues(entityType, entityName, enforcement.EnforcementMode).Desc(),
			prometheus.GaugeValue,
			enforcement.ComplianceMetrics.PolicyAdherence,
		)
	}

	// Enforcement actions
	for _, action := range enforcement.EnforcementActions {
		ch <- prometheus.MustNewConstMetric(
			c.quotaEnforcementActions.WithLabelValues(entityType, entityName, action.ActionType, action.Result).Desc(),
			prometheus.CounterValue,
			1,
		)
	}

	// Active exceptions
	if enforcement.ExceptionStatus != nil {
		ch <- prometheus.MustNewConstMetric(
			c.quotaEnforcementExceptions.WithLabelValues(entityType, entityName, "active").Desc(),
			prometheus.GaugeValue,
			float64(len(enforcement.ExceptionStatus.ActiveExceptions)),
		)
	}

	// Active overrides
	if enforcement.OverrideStatus != nil {
		ch <- prometheus.MustNewConstMetric(
			c.quotaEnforcementOverrides.WithLabelValues(entityType, entityName, "active", "authorized").Desc(),
			prometheus.CounterValue,
			float64(len(enforcement.OverrideStatus.ActiveOverrides)),
		)
	}
}

func (c *QuotaComplianceCollector) collectTrendMetrics(ch chan<- prometheus.Metric, entityType, entityName string, trends *QuotaTrends) {
	// TODO: QuotaTrends struct is missing expected fields (UsageTrends, SeasonalPatterns, AnomalyDetection)
	// Skipping trend metrics collection for now
	/*
		// Usage trends
		for _, usageTrend := range trends.UsageTrends {
			trendDirection := 0.0
			switch usageTrend.TrendDirection {
			case "INCREASING":
				trendDirection = 1.0
			case "DECREASING":
				trendDirection = -1.0
			default:
				trendDirection = 0.0
			}

			ch <- prometheus.MustNewConstMetric(
				c.quotaTrendDirection.WithLabelValues(entityType, entityName, usageTrend.ResourceType).Desc(),
				prometheus.GaugeValue,
				trendDirection,
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaTrendGrowthRate.WithLabelValues(entityType, entityName, usageTrend.ResourceType).Desc(),
				prometheus.GaugeValue,
				usageTrend.GrowthRate,
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaTrendVolatility.WithLabelValues(entityType, entityName, usageTrend.ResourceType).Desc(),
				prometheus.GaugeValue,
				usageTrend.Volatility,
			)
		}

		// Seasonal patterns
		for _, pattern := range trends.SeasonalPatterns {
			ch <- prometheus.MustNewConstMetric(
				c.quotaTrendSeasonality.WithLabelValues(entityType, entityName, pattern.PatternType).Desc(),
				prometheus.GaugeValue,
				pattern.PatternStrength,
			)
		}

		// Anomaly detection
		if trends.AnomalyDetection != nil {
			ch <- prometheus.MustNewConstMetric(
				c.quotaTrendAnomalies.WithLabelValues(entityType, entityName, "detected").Desc(),
				prometheus.GaugeValue,
				float64(trends.AnomalyDetection.AnomaliesDetected),
			)
		}
	*/
}

func (c *QuotaComplianceCollector) collectForecastMetrics(ch chan<- prometheus.Metric, entityType, entityName string, forecast *QuotaForecast) {
	// Model accuracy
	if forecast.ModelAccuracy != nil {
		ch <- prometheus.MustNewConstMetric(
			c.quotaForecastAccuracy.WithLabelValues(entityType, entityName, "overall", forecast.ForecastHorizon).Desc(),
			prometheus.GaugeValue,
			forecast.ModelAccuracy.OverallAccuracy,
		)
		ch <- prometheus.MustNewConstMetric(
			c.quotaForecastConfidence.WithLabelValues(entityType, entityName, "model").Desc(),
			prometheus.GaugeValue,
			forecast.ModelAccuracy.ConfidenceScore,
		)
	}

	// Resource forecasts
	for _, resourceForecast := range forecast.ResourceForecasts {
		ch <- prometheus.MustNewConstMetric(
			c.quotaForecastConfidence.WithLabelValues(entityType, entityName, resourceForecast.ResourceType).Desc(),
			prometheus.GaugeValue,
			resourceForecast.ConfidenceLevel,
		)

		if resourceForecast.ProjectedExhaustion != nil {
			daysToExhaustion := time.Until(*resourceForecast.ProjectedExhaustion).Hours() / 24
			ch <- prometheus.MustNewConstMetric(
				c.quotaForecastExhaustion.WithLabelValues(entityType, entityName, resourceForecast.ResourceType).Desc(),
				prometheus.GaugeValue,
				daysToExhaustion,
			)
		}
	}

	// Risk forecasts
	for _, riskForecast := range forecast.RiskForecasts {
		ch <- prometheus.MustNewConstMetric(
			c.quotaForecastRisk.WithLabelValues(entityType, entityName, riskForecast.RiskType).Desc(),
			prometheus.GaugeValue,
			riskForecast.RiskProbability,
		)
	}
}

func (c *QuotaComplianceCollector) collectRecommendationMetrics(ch chan<- prometheus.Metric, entityType, entityName string, recommendations *QuotaRecommendations) {
	// TODO: QuotaRecommendations struct is missing expected fields (QuotaOptimization, PolicyRecommendations, PriorityMatrix)
	// Skipping recommendation metrics collection for now
	/*
		// Overall recommendations metrics
		ch <- prometheus.MustNewConstMetric(
			c.quotaRecommendationsTotal.WithLabelValues(entityType, entityName, "overall").Desc(),
			prometheus.GaugeValue,
			float64(len(recommendations.QuotaOptimization)+len(recommendations.PolicyRecommendations)),
		)

		// Priority matrix
		if recommendations.PriorityMatrix != nil {
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsPriority.WithLabelValues(entityType, entityName, "high").Desc(),
				prometheus.GaugeValue,
				float64(len(recommendations.PriorityMatrix.HighPriority)),
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsPriority.WithLabelValues(entityType, entityName, "medium").Desc(),
				prometheus.GaugeValue,
				float64(len(recommendations.PriorityMatrix.MediumPriority)),
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsPriority.WithLabelValues(entityType, entityName, "low").Desc(),
				prometheus.GaugeValue,
				float64(len(recommendations.PriorityMatrix.LowPriority)),
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsPriority.WithLabelValues(entityType, entityName, "quick_wins").Desc(),
				prometheus.GaugeValue,
				float64(len(recommendations.PriorityMatrix.QuickWins)),
			)
		}

		// ROI metrics
		for _, optimization := range recommendations.QuotaOptimization {
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsROI.WithLabelValues(entityType, entityName, optimization.ResourceType).Desc(),
				prometheus.GaugeValue,
				optimization.ROI,
			)
		}

		// Expected outcomes
		if recommendations.ExpectedOutcomes != nil {
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsROI.WithLabelValues(entityType, entityName, "compliance_improvement").Desc(),
				prometheus.GaugeValue,
				recommendations.ExpectedOutcomes.ComplianceImprovement,
			)
			ch <- prometheus.MustNewConstMetric(
				c.quotaRecommendationsROI.WithLabelValues(entityType, entityName, "cost_reduction").Desc(),
				prometheus.GaugeValue,
				recommendations.ExpectedOutcomes.CostReduction,
			)
		}
	*/
}

func (c *QuotaComplianceCollector) collectAuditMetrics(ch chan<- prometheus.Metric, entityType, entityName string, auditEntries []*QuotaAuditEntry) {
	for _, entry := range auditEntries {
		ch <- prometheus.MustNewConstMetric(
			c.quotaAuditEvents.WithLabelValues(entityType, entityName, entry.EventType, entry.Severity).Desc(),
			prometheus.CounterValue,
			1,
		)

		if !entry.Success {
			ch <- prometheus.MustNewConstMetric(
				c.quotaAuditFailures.WithLabelValues(entityType, entityName, entry.EventType).Desc(),
				prometheus.CounterValue,
				1,
			)
		}
	}
}

func (c *QuotaComplianceCollector) collectNotificationMetrics(ch chan<- prometheus.Metric, entityType, entityName string, settings *QuotaNotificationSettings) {
	// Count notification history by status
	for _, history := range settings.NotificationHistory {
		channel := history.Channel
		if channel == "" {
			channel = "unknown"
		}

		ch <- prometheus.MustNewConstMetric(
			c.quotaNotificationsSent.WithLabelValues(entityType, entityName, history.EventType, channel).Desc(),
			prometheus.CounterValue,
			1,
		)

		switch history.Status {
		case "DELIVERED":
			ch <- prometheus.MustNewConstMetric(
				c.quotaNotificationsDelivered.WithLabelValues(entityType, entityName, history.EventType, channel).Desc(),
				prometheus.CounterValue,
				1,
			)
			if history.DeliveryTime != nil {
				latency := history.DeliveryTime.Sub(history.Timestamp).Seconds()
				c.quotaNotificationsLatency.WithLabelValues(entityType, entityName, history.EventType, channel).Observe(latency)
			}
		case "FAILED":
			failureReason := "unknown"
			if history.ErrorDetails != "" {
				failureReason = "error"
			}
			ch <- prometheus.MustNewConstMetric(
				c.quotaNotificationsFailed.WithLabelValues(entityType, entityName, history.EventType, channel, failureReason).Desc(),
				prometheus.CounterValue,
				1,
			)
		}
	}
}
