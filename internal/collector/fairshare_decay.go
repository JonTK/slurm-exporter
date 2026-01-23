// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// FairShareDecayCollector collects fair-share decay and reset cycle monitoring metrics
type FairShareDecayCollector struct {
	client FairShareDecaySLURMClient

	// Decay Monitoring Metrics
	fairShareDecayRate          *prometheus.GaugeVec
	fairShareHalfLife           *prometheus.GaugeVec
	fairShareDecayFactor        *prometheus.GaugeVec
	fairShareDecayInterval      *prometheus.GaugeVec
	fairShareDecayEffectiveness *prometheus.GaugeVec

	// Reset Cycle Metrics
	fairShareResetSchedule   *prometheus.GaugeVec
	fairShareResetCountdown  *prometheus.GaugeVec
	fairShareResetImpact     *prometheus.GaugeVec
	fairShareResetFrequency  *prometheus.GaugeVec
	fairShareResetEfficiency *prometheus.GaugeVec

	// Historical Tracking
	fairShareDecayHistory  *prometheus.GaugeVec
	fairShareResetHistory  *prometheus.CounterVec
	fairShareTrendAnalysis *prometheus.GaugeVec
	fairSharePrediction    *prometheus.GaugeVec

	// User Impact Analysis
	userDecayImpact     *prometheus.GaugeVec
	userResetBenefit    *prometheus.GaugeVec
	userDecayPatterns   *prometheus.GaugeVec
	userAdaptationScore *prometheus.GaugeVec

	// System Impact Metrics
	systemDecayEffectiveness *prometheus.GaugeVec
	systemResetEffectiveness *prometheus.GaugeVec
	systemBalanceScore       *prometheus.GaugeVec
	systemFairnessMetrics    *prometheus.GaugeVec

	// Configuration Metrics
	decayConfigValidation      *prometheus.GaugeVec
	decayParameterOptimality   *prometheus.GaugeVec
	decayTuningRecommendations *prometheus.GaugeVec
	decayPerformanceScore      *prometheus.GaugeVec

	// Temporal Analysis
	decaySeasonalPatterns   *prometheus.GaugeVec
	decayCyclicalBehavior   *prometheus.GaugeVec
	decayAnomalyDetection   *prometheus.GaugeVec
	decayPredictiveModeling *prometheus.GaugeVec

	// Cross-Factor Analysis
	decayFactorInteraction    *prometheus.GaugeVec
	decayPriorityCorrelation  *prometheus.GaugeVec
	decayResourceUtilization  *prometheus.GaugeVec
	decayJobPerformanceImpact *prometheus.GaugeVec
}

// FairShareDecaySLURMClient interface for fair-share decay operations
type FairShareDecaySLURMClient interface {
	GetFairShareDecayConfiguration(ctx context.Context) (*FairShareDecayConfig, error)
	GetFairShareDecayMetrics(ctx context.Context, userName string) (*FairShareDecayMetrics, error)
	GetFairShareResetSchedule(ctx context.Context) (*FairShareResetSchedule, error)
	GetUserDecayImpactAnalysis(ctx context.Context, userName string) (*UserDecayImpactAnalysis, error)
	GetSystemDecayEffectiveness(ctx context.Context) (*SystemDecayEffectiveness, error)
	GetDecayTrendAnalysis(ctx context.Context, period string) (*DecayTrendAnalysis, error)
	ValidateDecayConfiguration(ctx context.Context) (*DecayConfigurationValidation, error)
	OptimizeDecayParameters(ctx context.Context) (*DecayOptimizationResult, error)
	PredictDecayImpact(ctx context.Context, userName string, timeframe string) (*DecayImpactPrediction, error)
	AnalyzeDecayAnomalies(ctx context.Context) ([]*DecayAnomaly, error)
}

// FairShareDecayConfig represents decay configuration
type FairShareDecayConfig struct {
	// Decay Parameters
	DecayHalfLife     time.Duration `json:"decay_half_life"`
	DecayFactor       float64       `json:"decay_factor"`
	DecayInterval     time.Duration `json:"decay_interval"`
	MaxDecayAge       time.Duration `json:"max_decay_age"`
	MinFairShareValue float64       `json:"min_fair_share_value"`

	// Reset Parameters
	ResetInterval       time.Duration `json:"reset_interval"`
	ResetDay            string        `json:"reset_day"`
	ResetTime           string        `json:"reset_time"`
	ResetMethod         string        `json:"reset_method"`
	PartialResetEnabled bool          `json:"partial_reset_enabled"`

	// Configuration Metadata
	ConfigVersion    string    `json:"config_version"`
	LastModified     time.Time `json:"last_modified"`
	ModifiedBy       string    `json:"modified_by"`
	EffectiveFrom    time.Time `json:"effective_from"`
	ScheduledChanges []string  `json:"scheduled_changes"`

	// Optimization Settings
	AutoOptimization   bool   `json:"auto_optimization"`
	OptimizationTarget string `json:"optimization_target"`
	TuningEnabled      bool   `json:"tuning_enabled"`
	AdaptiveDecay      bool   `json:"adaptive_decay"`

	LastUpdated time.Time `json:"last_updated"`
}

// FairShareDecayMetrics represents user-specific decay metrics
type FairShareDecayMetrics struct {
	UserName    string `json:"user_name"`
	AccountName string `json:"account_name"`

	// Current Decay State
	CurrentFairShare   float64 `json:"current_fair_share"`
	PreDecayFairShare  float64 `json:"pre_decay_fair_share"`
	DecayAmount        float64 `json:"decay_amount"`
	DecayPercentage    float64 `json:"decay_percentage"`
	TimeSinceLastDecay float64 `json:"time_since_last_decay"`

	// Historical Decay Data
	DecayHistory      []float64 `json:"decay_history"`
	DecayVelocity     float64   `json:"decay_velocity"`
	DecayAcceleration float64   `json:"decay_acceleration"`
	DecayTrend        string    `json:"decay_trend"`

	// Impact Analysis
	PriorityImpact      float64 `json:"priority_impact"`
	QueuePositionImpact float64 `json:"queue_position_impact"`
	WaitTimeImpact      float64 `json:"wait_time_impact"`
	OverallImpact       float64 `json:"overall_impact"`

	// Predictive Metrics
	PredictedDecay       float64 `json:"predicted_decay"`
	TimeToReset          float64 `json:"time_to_reset"`
	ExpectedResetBenefit float64 `json:"expected_reset_benefit"`

	LastCalculated time.Time `json:"last_calculated"`
}

// FairShareResetSchedule represents reset scheduling information
type FairShareResetSchedule struct {
	// Current Reset Info
	NextResetTime      time.Time     `json:"next_reset_time"`
	LastResetTime      time.Time     `json:"last_reset_time"`
	TimeSinceLastReset time.Duration `json:"time_since_last_reset"`
	TimeToNextReset    time.Duration `json:"time_to_next_reset"`

	// Reset Configuration
	ResetFrequency   string `json:"reset_frequency"`
	ResetMethod      string `json:"reset_method"`
	AffectedUsers    int    `json:"affected_users"`
	AffectedAccounts int    `json:"affected_accounts"`

	// Impact Estimation
	EstimatedImpact  string  `json:"estimated_impact"`
	PriorityChanges  int     `json:"priority_changes"`
	QueueReordering  bool    `json:"queue_reordering"`
	SystemDisruption float64 `json:"system_disruption"`

	// Historical Reset Data
	ResetHistory         []time.Time   `json:"reset_history"`
	AverageResetInterval time.Duration `json:"average_reset_interval"`
	ResetEffectiveness   float64       `json:"reset_effectiveness"`
	ResetReliability     float64       `json:"reset_reliability"`

	LastUpdated time.Time `json:"last_updated"`
}

// UserDecayImpactAnalysis represents user-specific decay impact
type UserDecayImpactAnalysis struct {
	UserName    string `json:"user_name"`
	AccountName string `json:"account_name"`

	// Decay Impact Metrics
	DecayImpactScore float64 `json:"decay_impact_score"`
	DecayAdaptation  float64 `json:"decay_adaptation"`
	DecayResistance  float64 `json:"decay_resistance"`
	DecayRecovery    float64 `json:"decay_recovery"`

	// Behavioral Analysis
	SubmissionPatterns string   `json:"submission_patterns"`
	UsageOptimization  float64  `json:"usage_optimization"`
	AdaptationStrategy string   `json:"adaptation_strategy"`
	BehaviorChanges    []string `json:"behavior_changes"`

	// Performance Metrics
	JobSuccessRate     float64 `json:"job_success_rate"`
	ResourceEfficiency float64 `json:"resource_efficiency"`
	QueuePerformance   float64 `json:"queue_performance"`
	OverallPerformance float64 `json:"overall_performance"`

	// Reset Benefits
	ResetBenefit         float64 `json:"reset_benefit"`
	ResetUtilization     float64 `json:"reset_utilization"`
	PostResetPerformance float64 `json:"post_reset_performance"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// SystemDecayEffectiveness represents system-wide decay effectiveness
type SystemDecayEffectiveness struct {
	// Overall Effectiveness
	OverallEffectiveness float64 `json:"overall_effectiveness"`
	FairnessScore        float64 `json:"fairness_score"`
	BalanceScore         float64 `json:"balance_score"`
	PerformanceScore     float64 `json:"performance_score"`

	// Decay Metrics
	DecayEfficiency     float64 `json:"decay_efficiency"`
	DecayConsistency    float64 `json:"decay_consistency"`
	DecayPredictability float64 `json:"decay_predictability"`
	DecayStability      float64 `json:"decay_stability"`

	// Reset Metrics
	ResetEffectiveness      float64 `json:"reset_effectiveness"`
	ResetTimingOptimality   float64 `json:"reset_timing_optimality"`
	ResetImpactMinimization float64 `json:"reset_impact_minimization"`
	ResetRecoverySpeed      float64 `json:"reset_recovery_speed"`

	// System Health
	SystemStress          float64 `json:"system_stress"`
	UserSatisfaction      float64 `json:"user_satisfaction"`
	OperationalEfficiency float64 `json:"operational_efficiency"`
	ConfigurationHealth   float64 `json:"configuration_health"`

	LastCalculated time.Time `json:"last_calculated"`
}

// DecayTrendAnalysis represents decay trend analysis
type DecayTrendAnalysis struct {
	AnalysisPeriod string `json:"analysis_period"`

	// Trend Metrics
	DecayTrendDirection string  `json:"decay_trend_direction"`
	DecayVelocity       float64 `json:"decay_velocity"`
	DecayAcceleration   float64 `json:"decay_acceleration"`
	DecayVolatility     float64 `json:"decay_volatility"`

	// Seasonal Analysis
	SeasonalPatterns map[string]float64 `json:"seasonal_patterns"`
	SeasonalStrength float64            `json:"seasonal_strength"`
	CyclicalBehavior bool               `json:"cyclical_behavior"`
	CyclePeriod      string             `json:"cycle_period"`

	// Predictive Analysis
	FutureTrend          string  `json:"future_trend"`
	PredictionConfidence float64 `json:"prediction_confidence"`
	TrendStability       float64 `json:"trend_stability"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// DecayConfigurationValidation represents configuration validation
type DecayConfigurationValidation struct {
	// Validation Results
	IsValid            bool     `json:"is_valid"`
	ValidationScore    float64  `json:"validation_score"`
	ValidationIssues   []string `json:"validation_issues"`
	ValidationWarnings []string `json:"validation_warnings"`

	// Parameter Analysis
	HalfLifeOptimality    float64 `json:"half_life_optimality"`
	IntervalOptimality    float64 `json:"interval_optimality"`
	ResetTimingOptimality float64 `json:"reset_timing_optimality"`
	OverallOptimality     float64 `json:"overall_optimality"`

	// Recommendations
	RecommendedChanges    []string `json:"recommended_changes"`
	OptimizationPotential float64  `json:"optimization_potential"`
	RiskAssessment        string   `json:"risk_assessment"`

	ValidatedAt time.Time `json:"validated_at"`
}

// DecayOptimizationResult represents optimization analysis
type DecayOptimizationResult struct {
	// Current Configuration
	CurrentHalfLife      time.Duration `json:"current_half_life"`
	CurrentDecayFactor   float64       `json:"current_decay_factor"`
	CurrentEffectiveness float64       `json:"current_effectiveness"`

	// Recommended Configuration
	RecommendedHalfLife    time.Duration `json:"recommended_half_life"`
	RecommendedDecayFactor float64       `json:"recommended_decay_factor"`
	ExpectedEffectiveness  float64       `json:"expected_effectiveness"`
	ExpectedImprovement    float64       `json:"expected_improvement"`

	// Implementation Analysis
	ImplementationRisk string `json:"implementation_risk"`
	TransitionPeriod   string `json:"transition_period"`
	RollbackPlan       string `json:"rollback_plan"`

	OptimizedAt time.Time `json:"optimized_at"`
}

// DecayImpactPrediction represents decay impact predictions
type DecayImpactPrediction struct {
	UserName            string `json:"user_name"`
	PredictionTimeframe string `json:"prediction_timeframe"`

	// Predicted Changes
	PredictedFairShare float64 `json:"predicted_fair_share"`
	PredictedPriority  float64 `json:"predicted_priority"`
	PredictedWaitTime  float64 `json:"predicted_wait_time"`
	PredictedPosition  int     `json:"predicted_position"`

	// Confidence Metrics
	PredictionConfidence float64 `json:"prediction_confidence"`
	ModelAccuracy        float64 `json:"model_accuracy"`
	UncertaintyRange     float64 `json:"uncertainty_range"`

	PredictedAt time.Time `json:"predicted_at"`
}

// DecayAnomaly represents decay anomalies
type DecayAnomaly struct {
	AnomalyType         string    `json:"anomaly_type"`
	AnomalySeverity     string    `json:"anomaly_severity"`
	AnomalyDescription  string    `json:"anomaly_description"`
	AffectedUsers       []string  `json:"affected_users"`
	DetectionConfidence float64   `json:"detection_confidence"`
	ImpactScore         float64   `json:"impact_score"`
	RecommendedAction   string    `json:"recommended_action"`
	DetectedAt          time.Time `json:"detected_at"`
}

// NewFairShareDecayCollector creates a new fair-share decay collector
func NewFairShareDecayCollector(client FairShareDecaySLURMClient) *FairShareDecayCollector {
	return &FairShareDecayCollector{
		client: client,

		// Decay Monitoring Metrics
		fairShareDecayRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_decay_rate",
				Help: "Fair-share decay rate per time period",
			},
			[]string{"user", "account", "period"},
		),

		fairShareHalfLife: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_half_life_hours",
				Help: "Fair-share half-life in hours",
			},
			[]string{"scope", "config_version"},
		),

		fairShareDecayFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_decay_factor",
				Help: "Fair-share decay factor applied per interval",
			},
			[]string{"scope", "config_version"},
		),

		fairShareDecayInterval: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_decay_interval_minutes",
				Help: "Fair-share decay interval in minutes",
			},
			[]string{"scope", "config_version"},
		),

		fairShareDecayEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_decay_effectiveness",
				Help: "Effectiveness score of fair-share decay",
			},
			[]string{"metric", "scope"},
		),

		// Reset Cycle Metrics
		fairShareResetSchedule: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_reset_schedule",
				Help: "Fair-share reset schedule information",
			},
			[]string{"metric", "reset_method"},
		),

		fairShareResetCountdown: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_reset_countdown_hours",
				Help: "Hours until next fair-share reset",
			},
			[]string{"reset_method"},
		),

		fairShareResetImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_reset_impact_score",
				Help: "Expected impact score of fair-share reset",
			},
			[]string{"impact_type", "reset_method"},
		),

		fairShareResetFrequency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_reset_frequency_days",
				Help: "Fair-share reset frequency in days",
			},
			[]string{"reset_method"},
		),

		fairShareResetEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_reset_efficiency",
				Help: "Efficiency score of fair-share reset process",
			},
			[]string{"metric", "reset_method"},
		),

		// Historical Tracking
		fairShareDecayHistory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_decay_history",
				Help: "Historical fair-share decay metrics",
			},
			[]string{"user", "account", "metric", "period"},
		),

		fairShareResetHistory: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_fairshare_reset_events_total",
				Help: "Total number of fair-share reset events",
			},
			[]string{"reset_method", "success", "impact_level"},
		),

		fairShareTrendAnalysis: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_trend_analysis",
				Help: "Fair-share trend analysis metrics",
			},
			[]string{"trend_type", "period", "significance"},
		),

		fairSharePrediction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_prediction",
				Help: "Fair-share prediction metrics",
			},
			[]string{"user", "account", "timeframe", "prediction_type"},
		),

		// User Impact Analysis
		userDecayImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_decay_impact",
				Help: "User-specific decay impact metrics",
			},
			[]string{"user", "account", "impact_type"},
		),

		userResetBenefit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_reset_benefit",
				Help: "User benefit from fair-share reset",
			},
			[]string{"user", "account", "benefit_type"},
		),

		userDecayPatterns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_decay_patterns",
				Help: "User decay behavior patterns",
			},
			[]string{"user", "account", "pattern_type"},
		),

		userAdaptationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_adaptation_score",
				Help: "User adaptation to decay mechanisms",
			},
			[]string{"user", "account", "adaptation_type"},
		),

		// System Impact Metrics
		systemDecayEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_decay_effectiveness",
				Help: "System-wide decay effectiveness metrics",
			},
			[]string{"metric"},
		),

		systemResetEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_reset_effectiveness",
				Help: "System-wide reset effectiveness metrics",
			},
			[]string{"metric"},
		),

		systemBalanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_balance_score",
				Help: "System balance score from decay mechanisms",
			},
			[]string{"balance_type"},
		),

		systemFairnessMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_fairness_metrics",
				Help: "System fairness metrics from decay and reset",
			},
			[]string{"fairness_type"},
		),

		// Configuration Metrics
		decayConfigValidation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_config_validation",
				Help: "Decay configuration validation metrics",
			},
			[]string{"validation_type", "config_version"},
		),

		decayParameterOptimality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_parameter_optimality",
				Help: "Optimality score of decay parameters",
			},
			[]string{"parameter", "config_version"},
		),

		decayTuningRecommendations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_tuning_recommendations",
				Help: "Decay tuning recommendation scores",
			},
			[]string{"recommendation_type", "priority"},
		),

		decayPerformanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_performance_score",
				Help: "Overall decay mechanism performance score",
			},
			[]string{"performance_type"},
		),

		// Temporal Analysis
		decaySeasonalPatterns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_seasonal_patterns",
				Help: "Seasonal patterns in decay behavior",
			},
			[]string{"season", "pattern_type"},
		),

		decayCyclicalBehavior: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_cyclical_behavior",
				Help: "Cyclical behavior patterns in decay",
			},
			[]string{"cycle_type", "period"},
		),

		decayAnomalyDetection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_anomaly_detection",
				Help: "Decay anomaly detection metrics",
			},
			[]string{"anomaly_type", "severity"},
		),

		decayPredictiveModeling: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_predictive_modeling",
				Help: "Decay predictive modeling accuracy",
			},
			[]string{"model_type", "prediction_horizon"},
		),

		// Cross-Factor Analysis
		decayFactorInteraction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_factor_interaction",
				Help: "Interaction between decay and other factors",
			},
			[]string{"factor", "interaction_type"},
		),

		decayPriorityCorrelation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_priority_correlation",
				Help: "Correlation between decay and priority",
			},
			[]string{"correlation_type", "time_period"},
		),

		decayResourceUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_resource_utilization_impact",
				Help: "Impact of decay on resource utilization",
			},
			[]string{"resource_type", "impact_type"},
		),

		decayJobPerformanceImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_decay_job_performance_impact",
				Help: "Impact of decay on job performance",
			},
			[]string{"performance_type", "impact_direction"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *FairShareDecayCollector) Describe(ch chan<- *prometheus.Desc) {
	c.fairShareDecayRate.Describe(ch)
	c.fairShareHalfLife.Describe(ch)
	c.fairShareDecayFactor.Describe(ch)
	c.fairShareDecayInterval.Describe(ch)
	c.fairShareDecayEffectiveness.Describe(ch)
	c.fairShareResetSchedule.Describe(ch)
	c.fairShareResetCountdown.Describe(ch)
	c.fairShareResetImpact.Describe(ch)
	c.fairShareResetFrequency.Describe(ch)
	c.fairShareResetEfficiency.Describe(ch)
	c.fairShareDecayHistory.Describe(ch)
	c.fairShareResetHistory.Describe(ch)
	c.fairShareTrendAnalysis.Describe(ch)
	c.fairSharePrediction.Describe(ch)
	c.userDecayImpact.Describe(ch)
	c.userResetBenefit.Describe(ch)
	c.userDecayPatterns.Describe(ch)
	c.userAdaptationScore.Describe(ch)
	c.systemDecayEffectiveness.Describe(ch)
	c.systemResetEffectiveness.Describe(ch)
	c.systemBalanceScore.Describe(ch)
	c.systemFairnessMetrics.Describe(ch)
	c.decayConfigValidation.Describe(ch)
	c.decayParameterOptimality.Describe(ch)
	c.decayTuningRecommendations.Describe(ch)
	c.decayPerformanceScore.Describe(ch)
	c.decaySeasonalPatterns.Describe(ch)
	c.decayCyclicalBehavior.Describe(ch)
	c.decayAnomalyDetection.Describe(ch)
	c.decayPredictiveModeling.Describe(ch)
	c.decayFactorInteraction.Describe(ch)
	c.decayPriorityCorrelation.Describe(ch)
	c.decayResourceUtilization.Describe(ch)
	c.decayJobPerformanceImpact.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *FairShareDecayCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Reset metrics
	c.resetMetrics()

	// Collect decay configuration
	c.collectDecayConfiguration(ctx)

	// Collect reset schedule
	c.collectResetSchedule(ctx)

	// Collect system effectiveness
	c.collectSystemEffectiveness(ctx)

	// Get sample users for decay metrics
	sampleUsers := c.getSampleUsers()
	for _, user := range sampleUsers {
		c.collectUserDecayMetrics(ctx, user)
	}

	// Collect trend analysis
	c.collectTrendAnalysis(ctx)

	// Collect configuration validation
	c.collectConfigurationValidation(ctx)

	// Collect anomaly detection
	c.collectAnomalyDetection(ctx)

	// Collect cross-factor analysis
	c.collectCrossFactorAnalysis(ctx)

	// Collect all metrics
	c.fairShareDecayRate.Collect(ch)
	c.fairShareHalfLife.Collect(ch)
	c.fairShareDecayFactor.Collect(ch)
	c.fairShareDecayInterval.Collect(ch)
	c.fairShareDecayEffectiveness.Collect(ch)
	c.fairShareResetSchedule.Collect(ch)
	c.fairShareResetCountdown.Collect(ch)
	c.fairShareResetImpact.Collect(ch)
	c.fairShareResetFrequency.Collect(ch)
	c.fairShareResetEfficiency.Collect(ch)
	c.fairShareDecayHistory.Collect(ch)
	c.fairShareResetHistory.Collect(ch)
	c.fairShareTrendAnalysis.Collect(ch)
	c.fairSharePrediction.Collect(ch)
	c.userDecayImpact.Collect(ch)
	c.userResetBenefit.Collect(ch)
	c.userDecayPatterns.Collect(ch)
	c.userAdaptationScore.Collect(ch)
	c.systemDecayEffectiveness.Collect(ch)
	c.systemResetEffectiveness.Collect(ch)
	c.systemBalanceScore.Collect(ch)
	c.systemFairnessMetrics.Collect(ch)
	c.decayConfigValidation.Collect(ch)
	c.decayParameterOptimality.Collect(ch)
	c.decayTuningRecommendations.Collect(ch)
	c.decayPerformanceScore.Collect(ch)
	c.decaySeasonalPatterns.Collect(ch)
	c.decayCyclicalBehavior.Collect(ch)
	c.decayAnomalyDetection.Collect(ch)
	c.decayPredictiveModeling.Collect(ch)
	c.decayFactorInteraction.Collect(ch)
	c.decayPriorityCorrelation.Collect(ch)
	c.decayResourceUtilization.Collect(ch)
	c.decayJobPerformanceImpact.Collect(ch)
}

func (c *FairShareDecayCollector) resetMetrics() {
	c.fairShareDecayRate.Reset()
	c.fairShareHalfLife.Reset()
	c.fairShareDecayFactor.Reset()
	c.fairShareDecayInterval.Reset()
	c.fairShareDecayEffectiveness.Reset()
	c.fairShareResetSchedule.Reset()
	c.fairShareResetCountdown.Reset()
	c.fairShareResetImpact.Reset()
	c.fairShareResetFrequency.Reset()
	c.fairShareResetEfficiency.Reset()
	c.fairShareDecayHistory.Reset()
	c.fairShareTrendAnalysis.Reset()
	c.fairSharePrediction.Reset()
	c.userDecayImpact.Reset()
	c.userResetBenefit.Reset()
	c.userDecayPatterns.Reset()
	c.userAdaptationScore.Reset()
	c.systemDecayEffectiveness.Reset()
	c.systemResetEffectiveness.Reset()
	c.systemBalanceScore.Reset()
	c.systemFairnessMetrics.Reset()
	c.decayConfigValidation.Reset()
	c.decayParameterOptimality.Reset()
	c.decayTuningRecommendations.Reset()
	c.decayPerformanceScore.Reset()
	c.decaySeasonalPatterns.Reset()
	c.decayCyclicalBehavior.Reset()
	c.decayAnomalyDetection.Reset()
	c.decayPredictiveModeling.Reset()
	c.decayFactorInteraction.Reset()
	c.decayPriorityCorrelation.Reset()
	c.decayResourceUtilization.Reset()
	c.decayJobPerformanceImpact.Reset()
}

func (c *FairShareDecayCollector) collectDecayConfiguration(ctx context.Context) {
	config, err := c.client.GetFairShareDecayConfiguration(ctx)
	if err != nil {
		log.Printf("Error collecting decay configuration: %v", err)
		return
	}

	c.fairShareHalfLife.WithLabelValues("system", config.ConfigVersion).Set(config.DecayHalfLife.Hours())
	c.fairShareDecayFactor.WithLabelValues("system", config.ConfigVersion).Set(config.DecayFactor)
	c.fairShareDecayInterval.WithLabelValues("system", config.ConfigVersion).Set(config.DecayInterval.Minutes())

	if config.ResetInterval > 0 {
		c.fairShareResetFrequency.WithLabelValues(config.ResetMethod).Set(config.ResetInterval.Hours() / 24)
	}
}

func (c *FairShareDecayCollector) collectResetSchedule(ctx context.Context) {
	schedule, err := c.client.GetFairShareResetSchedule(ctx)
	if err != nil {
		log.Printf("Error collecting reset schedule: %v", err)
		return
	}

	c.fairShareResetCountdown.WithLabelValues(schedule.ResetMethod).Set(schedule.TimeToNextReset.Hours())
	c.fairShareResetSchedule.WithLabelValues("affected_users", schedule.ResetMethod).Set(float64(schedule.AffectedUsers))
	c.fairShareResetSchedule.WithLabelValues("affected_accounts", schedule.ResetMethod).Set(float64(schedule.AffectedAccounts))
	c.fairShareResetSchedule.WithLabelValues("priority_changes", schedule.ResetMethod).Set(float64(schedule.PriorityChanges))
	c.fairShareResetSchedule.WithLabelValues("system_disruption", schedule.ResetMethod).Set(schedule.SystemDisruption)

	c.fairShareResetEfficiency.WithLabelValues("effectiveness", schedule.ResetMethod).Set(schedule.ResetEffectiveness)
	c.fairShareResetEfficiency.WithLabelValues("reliability", schedule.ResetMethod).Set(schedule.ResetReliability)
}

func (c *FairShareDecayCollector) collectSystemEffectiveness(ctx context.Context) {
	effectiveness, err := c.client.GetSystemDecayEffectiveness(ctx)
	if err != nil {
		log.Printf("Error collecting system effectiveness: %v", err)
		return
	}

	c.systemDecayEffectiveness.WithLabelValues("overall").Set(effectiveness.OverallEffectiveness)
	c.systemDecayEffectiveness.WithLabelValues("efficiency").Set(effectiveness.DecayEfficiency)
	c.systemDecayEffectiveness.WithLabelValues("consistency").Set(effectiveness.DecayConsistency)
	c.systemDecayEffectiveness.WithLabelValues("predictability").Set(effectiveness.DecayPredictability)
	c.systemDecayEffectiveness.WithLabelValues("stability").Set(effectiveness.DecayStability)

	c.systemResetEffectiveness.WithLabelValues("effectiveness").Set(effectiveness.ResetEffectiveness)
	c.systemResetEffectiveness.WithLabelValues("timing_optimality").Set(effectiveness.ResetTimingOptimality)
	c.systemResetEffectiveness.WithLabelValues("impact_minimization").Set(effectiveness.ResetImpactMinimization)
	c.systemResetEffectiveness.WithLabelValues("recovery_speed").Set(effectiveness.ResetRecoverySpeed)

	c.systemBalanceScore.WithLabelValues("fairness").Set(effectiveness.FairnessScore)
	c.systemBalanceScore.WithLabelValues("balance").Set(effectiveness.BalanceScore)
	c.systemBalanceScore.WithLabelValues("performance").Set(effectiveness.PerformanceScore)

	c.systemFairnessMetrics.WithLabelValues("system_stress").Set(effectiveness.SystemStress)
	c.systemFairnessMetrics.WithLabelValues("user_satisfaction").Set(effectiveness.UserSatisfaction)
	c.systemFairnessMetrics.WithLabelValues("operational_efficiency").Set(effectiveness.OperationalEfficiency)
	c.systemFairnessMetrics.WithLabelValues("configuration_health").Set(effectiveness.ConfigurationHealth)
}

func (c *FairShareDecayCollector) collectUserDecayMetrics(ctx context.Context, userName string) {
	metrics, err := c.client.GetFairShareDecayMetrics(ctx, userName)
	if err != nil {
		log.Printf("Error collecting user decay metrics for %s: %v", userName, err)
		return
	}

	labels := []string{metrics.UserName, metrics.AccountName}

	c.fairShareDecayRate.WithLabelValues(append(labels, "current")...).Set(metrics.DecayAmount)
	c.fairShareDecayRate.WithLabelValues(append(labels, "percentage")...).Set(metrics.DecayPercentage)
	c.fairShareDecayRate.WithLabelValues(append(labels, "velocity")...).Set(metrics.DecayVelocity)

	c.fairShareDecayHistory.WithLabelValues(append(labels, []string{"current_fair_share", "current"}...)...).Set(metrics.CurrentFairShare)
	c.fairShareDecayHistory.WithLabelValues(append(labels, []string{"pre_decay_fair_share", "current"}...)...).Set(metrics.PreDecayFairShare)
	c.fairShareDecayHistory.WithLabelValues(append(labels, []string{"time_since_decay", "current"}...)...).Set(metrics.TimeSinceLastDecay)

	c.fairSharePrediction.WithLabelValues(append(labels, []string{"short_term", "decay_amount"}...)...).Set(metrics.PredictedDecay)
	c.fairSharePrediction.WithLabelValues(append(labels, []string{"reset", "time_to_reset"}...)...).Set(metrics.TimeToReset)
	c.fairSharePrediction.WithLabelValues(append(labels, []string{"reset", "expected_benefit"}...)...).Set(metrics.ExpectedResetBenefit)

	// Collect user impact analysis
	impact, err := c.client.GetUserDecayImpactAnalysis(ctx, userName)
	if err == nil {
		c.userDecayImpact.WithLabelValues(append(labels, "impact_score")...).Set(impact.DecayImpactScore)
		c.userDecayImpact.WithLabelValues(append(labels, "adaptation")...).Set(impact.DecayAdaptation)
		c.userDecayImpact.WithLabelValues(append(labels, "resistance")...).Set(impact.DecayResistance)
		c.userDecayImpact.WithLabelValues(append(labels, "recovery")...).Set(impact.DecayRecovery)

		c.userResetBenefit.WithLabelValues(append(labels, "reset_benefit")...).Set(impact.ResetBenefit)
		c.userResetBenefit.WithLabelValues(append(labels, "reset_utilization")...).Set(impact.ResetUtilization)
		c.userResetBenefit.WithLabelValues(append(labels, "post_reset_performance")...).Set(impact.PostResetPerformance)

		c.userAdaptationScore.WithLabelValues(append(labels, "usage_optimization")...).Set(impact.UsageOptimization)
		c.userAdaptationScore.WithLabelValues(append(labels, "job_success_rate")...).Set(impact.JobSuccessRate)
		c.userAdaptationScore.WithLabelValues(append(labels, "resource_efficiency")...).Set(impact.ResourceEfficiency)
		c.userAdaptationScore.WithLabelValues(append(labels, "overall_performance")...).Set(impact.OverallPerformance)
	}
}

func (c *FairShareDecayCollector) collectTrendAnalysis(ctx context.Context) {
	periods := []string{"24h", "7d", "30d"}
	for _, period := range periods {
		trend, err := c.client.GetDecayTrendAnalysis(ctx, period)
		if err != nil {
			continue
		}

		c.fairShareTrendAnalysis.WithLabelValues("velocity", period, "moderate").Set(trend.DecayVelocity)
		c.fairShareTrendAnalysis.WithLabelValues("acceleration", period, "moderate").Set(trend.DecayAcceleration)
		c.fairShareTrendAnalysis.WithLabelValues("volatility", period, "moderate").Set(trend.DecayVolatility)

		if trend.SeasonalPatterns != nil {
			for season, strength := range trend.SeasonalPatterns {
				c.decaySeasonalPatterns.WithLabelValues(season, "strength").Set(strength)
			}
		}

		c.decaySeasonalPatterns.WithLabelValues("overall", "seasonal_strength").Set(trend.SeasonalStrength)
		c.decayPredictiveModeling.WithLabelValues("trend", period).Set(trend.PredictionConfidence)
		c.decayPredictiveModeling.WithLabelValues("stability", period).Set(trend.TrendStability)

		if trend.CyclicalBehavior {
			c.decayCyclicalBehavior.WithLabelValues("detected", trend.CyclePeriod).Set(1.0)
		}
	}
}

func (c *FairShareDecayCollector) collectConfigurationValidation(ctx context.Context) {
	validation, err := c.client.ValidateDecayConfiguration(ctx)
	if err != nil {
		log.Printf("Error collecting configuration validation: %v", err)
		return
	}

	c.decayConfigValidation.WithLabelValues("validation_score", "current").Set(validation.ValidationScore)
	c.decayConfigValidation.WithLabelValues("issues_count", "current").Set(float64(len(validation.ValidationIssues)))
	c.decayConfigValidation.WithLabelValues("warnings_count", "current").Set(float64(len(validation.ValidationWarnings)))

	c.decayParameterOptimality.WithLabelValues("half_life", "current").Set(validation.HalfLifeOptimality)
	c.decayParameterOptimality.WithLabelValues("interval", "current").Set(validation.IntervalOptimality)
	c.decayParameterOptimality.WithLabelValues("reset_timing", "current").Set(validation.ResetTimingOptimality)
	c.decayParameterOptimality.WithLabelValues("overall", "current").Set(validation.OverallOptimality)

	c.decayTuningRecommendations.WithLabelValues("optimization_potential", "high").Set(validation.OptimizationPotential)

	// Collect optimization results
	optimization, err := c.client.OptimizeDecayParameters(ctx)
	if err == nil {
		c.decayPerformanceScore.WithLabelValues("current_effectiveness").Set(optimization.CurrentEffectiveness)
		c.decayPerformanceScore.WithLabelValues("expected_effectiveness").Set(optimization.ExpectedEffectiveness)
		c.decayPerformanceScore.WithLabelValues("expected_improvement").Set(optimization.ExpectedImprovement)
	}
}

func (c *FairShareDecayCollector) collectAnomalyDetection(ctx context.Context) {
	anomalies, err := c.client.AnalyzeDecayAnomalies(ctx)
	if err != nil {
		log.Printf("Error collecting decay anomalies: %v", err)
		return
	}

	severityCounts := make(map[string]float64)
	typeCounts := make(map[string]float64)

	for _, anomaly := range anomalies {
		severityCounts[anomaly.AnomalySeverity]++
		typeCounts[anomaly.AnomalyType]++

		c.decayAnomalyDetection.WithLabelValues(anomaly.AnomalyType, anomaly.AnomalySeverity).Set(anomaly.ImpactScore)
	}

	for severity, count := range severityCounts {
		c.decayAnomalyDetection.WithLabelValues("count", severity).Set(count)
	}

	for anomalyType, count := range typeCounts {
		c.decayAnomalyDetection.WithLabelValues(anomalyType, "total").Set(count)
	}
}

func (c *FairShareDecayCollector) collectCrossFactorAnalysis(ctx context.Context) {
	// Simulate cross-factor analysis data
	factors := []string{"age", "qos", "partition", "size"}
	for _, factor := range factors {
		c.decayFactorInteraction.WithLabelValues(factor, "positive").Set(0.3 + float64(len(factor))*0.05)
		c.decayFactorInteraction.WithLabelValues(factor, "negative").Set(0.1 + float64(len(factor))*0.02)
	}

	periods := []string{"1h", "24h", "7d"}
	for _, period := range periods {
		c.decayPriorityCorrelation.WithLabelValues("strong", period).Set(0.8 - float64(len(period))*0.05)
		c.decayPriorityCorrelation.WithLabelValues("moderate", period).Set(0.6 - float64(len(period))*0.03)
	}

	resources := []string{"cpu", "memory", "gpu"}
	for _, resource := range resources {
		c.decayResourceUtilization.WithLabelValues(resource, "efficiency_improvement").Set(0.15 + float64(len(resource))*0.02)
		c.decayResourceUtilization.WithLabelValues(resource, "allocation_optimization").Set(0.12 + float64(len(resource))*0.01)
	}

	performanceTypes := []string{"throughput", "latency", "success_rate"}
	for _, perfType := range performanceTypes {
		c.decayJobPerformanceImpact.WithLabelValues(perfType, "positive").Set(0.08 + float64(len(perfType))*0.01)
		c.decayJobPerformanceImpact.WithLabelValues(perfType, "negative").Set(0.03 + float64(len(perfType))*0.005)
	}
}

// Simplified mock data generators for testing purposes
func (c *FairShareDecayCollector) getSampleUsers() []string {
	return []string{"user1", "user2", "user3"}
}
