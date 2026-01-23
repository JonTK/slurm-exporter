// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PriorityFactorsCollector collects detailed priority factor breakdown metrics
type PriorityFactorsCollector struct {
	client PriorityFactorsSLURMClient

	// Priority Factor Component Metrics
	priorityFactorWeight       *prometheus.GaugeVec
	priorityFactorValue        *prometheus.GaugeVec
	priorityFactorContribution *prometheus.GaugeVec
	priorityFactorNormalized   *prometheus.GaugeVec
	priorityFactorEffective    *prometheus.GaugeVec

	// Factor Weight Analysis
	factorWeightDistribution   *prometheus.HistogramVec
	factorWeightBalance        *prometheus.GaugeVec
	factorWeightOptimization   *prometheus.GaugeVec
	factorWeightRecommendation *prometheus.GaugeVec

	// Priority Factor Trends
	factorTrendDirection    *prometheus.GaugeVec
	factorTrendVelocity     *prometheus.GaugeVec
	factorTrendAcceleration *prometheus.GaugeVec
	factorTrendVolatility   *prometheus.GaugeVec
	factorTrendPrediction   *prometheus.GaugeVec

	// Factor Impact Analysis
	factorImpactScore       *prometheus.GaugeVec
	factorImpactChange      *prometheus.CounterVec
	factorImpactCorrelation *prometheus.GaugeVec
	factorImpactSensitivity *prometheus.GaugeVec

	// System Factor Configuration
	systemFactorConfig        *prometheus.GaugeVec
	systemFactorCalibration   *prometheus.GaugeVec
	systemFactorEffectiveness *prometheus.GaugeVec
	systemFactorOptimality    *prometheus.GaugeVec

	// User-specific Factor Analysis
	userFactorProfile        *prometheus.GaugeVec
	userFactorDeviation      *prometheus.GaugeVec
	userFactorOptimization   *prometheus.GaugeVec
	userFactorRecommendation *prometheus.GaugeVec

	// Account Factor Aggregation
	accountFactorSummary    *prometheus.GaugeVec
	accountFactorVariance   *prometheus.GaugeVec
	accountFactorBalance    *prometheus.GaugeVec
	accountFactorEfficiency *prometheus.GaugeVec

	// Partition Factor Analysis
	partitionFactorWeights      *prometheus.GaugeVec
	partitionFactorUtilization  *prometheus.GaugeVec
	partitionFactorOptimization *prometheus.GaugeVec
	partitionFactorBalance      *prometheus.GaugeVec

	// QoS Factor Tracking
	qosFactorPriority      *prometheus.GaugeVec
	qosFactorMultiplier    *prometheus.GaugeVec
	qosFactorEffectiveness *prometheus.GaugeVec
	qosFactorUtilization   *prometheus.GaugeVec

	// Temporal Factor Analysis
	temporalFactorEvolution      *prometheus.GaugeVec
	temporalFactorCycles         *prometheus.GaugeVec
	temporalFactorSeasonality    *prometheus.GaugeVec
	temporalFactorPredictability *prometheus.GaugeVec
}

// PriorityFactorsSLURMClient interface for priority factor operations
type PriorityFactorsSLURMClient interface {
	GetPriorityFactorBreakdown(ctx context.Context, jobID string) (*PriorityFactorBreakdown, error)
	GetSystemFactorWeights(ctx context.Context) (*SystemFactorWeights, error)
	GetUserFactorProfile(ctx context.Context, userName string) (*UserFactorProfile, error)
	GetAccountFactorSummary(ctx context.Context, accountName string) (*AccountFactorSummary, error)
	GetPartitionFactorAnalysis(ctx context.Context, partition string) (*PartitionFactorAnalysis, error)
	GetQoSFactorAnalysis(ctx context.Context, qosName string) (*QoSFactorAnalysis, error)
	GetFactorTrendAnalysis(ctx context.Context, factor string, period string) (*FactorTrendAnalysis, error)
	GetFactorImpactAnalysis(ctx context.Context, factor string) (*FactorImpactAnalysis, error)
	OptimizeFactorWeights(ctx context.Context, scope string) (*FactorOptimizationResult, error)
	ValidateFactorConfiguration(ctx context.Context) (*FactorConfigurationValidation, error)
}

// PriorityFactorBreakdown represents detailed factor analysis
type PriorityFactorBreakdown struct {
	JobID         string `json:"job_id"`
	UserName      string `json:"user_name"`
	AccountName   string `json:"account_name"`
	PartitionName string `json:"partition_name"`
	QoSName       string `json:"qos_name"`

	// Individual Factor Values
	AgeValue       float64 `json:"age_value"`
	FairShareValue float64 `json:"fair_share_value"`
	QoSValue       float64 `json:"qos_value"`
	PartitionValue float64 `json:"partition_value"`
	SizeValue      float64 `json:"size_value"`
	AssocValue     float64 `json:"assoc_value"`
	NiceValue      float64 `json:"nice_value"`

	// Factor Weights
	AgeWeight       float64 `json:"age_weight"`
	FairShareWeight float64 `json:"fair_share_weight"`
	QoSWeight       float64 `json:"qos_weight"`
	PartitionWeight float64 `json:"partition_weight"`
	SizeWeight      float64 `json:"size_weight"`
	AssocWeight     float64 `json:"assoc_weight"`
	NiceWeight      float64 `json:"nice_weight"`

	// Factor Contributions
	AgeContribution       float64 `json:"age_contribution"`
	FairShareContribution float64 `json:"fair_share_contribution"`
	QoSContribution       float64 `json:"qos_contribution"`
	PartitionContribution float64 `json:"partition_contribution"`
	SizeContribution      float64 `json:"size_contribution"`
	AssocContribution     float64 `json:"assoc_contribution"`
	NiceContribution      float64 `json:"nice_contribution"`

	// Normalized Values (0-1 scale)
	AgeNormalized       float64 `json:"age_normalized"`
	FairShareNormalized float64 `json:"fair_share_normalized"`
	QoSNormalized       float64 `json:"qos_normalized"`
	PartitionNormalized float64 `json:"partition_normalized"`
	SizeNormalized      float64 `json:"size_normalized"`
	AssocNormalized     float64 `json:"assoc_normalized"`
	NiceNormalized      float64 `json:"nice_normalized"`

	// Effective Contributions
	AgeEffective       float64 `json:"age_effective"`
	FairShareEffective float64 `json:"fair_share_effective"`
	QoSEffective       float64 `json:"qos_effective"`
	PartitionEffective float64 `json:"partition_effective"`
	SizeEffective      float64 `json:"size_effective"`
	AssocEffective     float64 `json:"assoc_effective"`
	NiceEffective      float64 `json:"nice_effective"`

	// Analysis Metadata
	TotalPriority      float64   `json:"total_priority"`
	DominantFactor     string    `json:"dominant_factor"`
	SecondaryFactor    string    `json:"secondary_factor"`
	FactorBalance      float64   `json:"factor_balance"`
	ConfigurationScore float64   `json:"configuration_score"`
	AnalyzedAt         time.Time `json:"analyzed_at"`
}

// SystemFactorWeights represents system-wide factor configuration
type SystemFactorWeights struct {
	// Global Weights
	AgeWeightGlobal       float64 `json:"age_weight_global"`
	FairShareWeightGlobal float64 `json:"fair_share_weight_global"`
	QoSWeightGlobal       float64 `json:"qos_weight_global"`
	PartitionWeightGlobal float64 `json:"partition_weight_global"`
	SizeWeightGlobal      float64 `json:"size_weight_global"`
	AssocWeightGlobal     float64 `json:"assoc_weight_global"`
	NiceWeightGlobal      float64 `json:"nice_weight_global"`

	// Weight Ranges
	AgeWeightRange       []float64 `json:"age_weight_range"`
	FairShareWeightRange []float64 `json:"fair_share_weight_range"`
	QoSWeightRange       []float64 `json:"qos_weight_range"`
	PartitionWeightRange []float64 `json:"partition_weight_range"`
	SizeWeightRange      []float64 `json:"size_weight_range"`

	// Configuration Analysis
	WeightBalance           float64 `json:"weight_balance"`
	WeightOptimality        float64 `json:"weight_optimality"`
	WeightEffectiveness     float64 `json:"weight_effectiveness"`
	ConfigurationVersion    string  `json:"configuration_version"`
	OptimizationRecommended bool    `json:"optimization_recommended"`

	// Calibration Data
	CalibrationAccuracy  float64   `json:"calibration_accuracy"`
	CalibrationStability float64   `json:"calibration_stability"`
	LastCalibration      time.Time `json:"last_calibration"`
	NextCalibration      time.Time `json:"next_calibration"`

	LastUpdated time.Time `json:"last_updated"`
}

// UserFactorProfile represents user-specific factor analysis
type UserFactorProfile struct {
	UserName    string `json:"user_name"`
	AccountName string `json:"account_name"`

	// User Factor Averages
	AvgAgeContribution       float64 `json:"avg_age_contribution"`
	AvgFairShareContribution float64 `json:"avg_fair_share_contribution"`
	AvgQoSContribution       float64 `json:"avg_qos_contribution"`
	AvgPartitionContribution float64 `json:"avg_partition_contribution"`
	AvgSizeContribution      float64 `json:"avg_size_contribution"`

	// Factor Variances
	AgeVariance       float64 `json:"age_variance"`
	FairShareVariance float64 `json:"fair_share_variance"`
	QoSVariance       float64 `json:"qos_variance"`
	PartitionVariance float64 `json:"partition_variance"`
	SizeVariance      float64 `json:"size_variance"`

	// User Behavior Analysis
	DominantFactorPattern       string   `json:"dominant_factor_pattern"`
	FactorStability             float64  `json:"factor_stability"`
	FactorOptimizationScore     float64  `json:"factor_optimization_score"`
	OptimizationRecommendations []string `json:"optimization_recommendations"`

	// Deviation Analysis
	DeviationFromNorm     float64 `json:"deviation_from_norm"`
	DeviationSignificance string  `json:"deviation_significance"`
	DeviationExplanation  string  `json:"deviation_explanation"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// AccountFactorSummary represents account-level factor aggregation
type AccountFactorSummary struct {
	AccountName string `json:"account_name"`

	// Aggregate Factor Metrics
	TotalUsers            int     `json:"total_users"`
	ActiveUsers           int     `json:"active_users"`
	AvgFactorBalance      float64 `json:"avg_factor_balance"`
	FactorVarianceScore   float64 `json:"factor_variance_score"`
	FactorEfficiencyScore float64 `json:"factor_efficiency_score"`

	// Factor Distribution
	AgeFactorDistribution       map[string]float64 `json:"age_factor_distribution"`
	FairShareFactorDistribution map[string]float64 `json:"fair_share_factor_distribution"`
	QoSFactorDistribution       map[string]float64 `json:"qos_factor_distribution"`

	// Account Optimization
	OptimizationPotential float64  `json:"optimization_potential"`
	RecommendedActions    []string `json:"recommended_actions"`
	EstimatedImprovement  float64  `json:"estimated_improvement"`

	LastSummarized time.Time `json:"last_summarized"`
}

// PartitionFactorAnalysis represents partition-specific factor analysis
type PartitionFactorAnalysis struct {
	PartitionName string `json:"partition_name"`

	// Partition Factor Configuration
	PartitionWeight           float64 `json:"partition_weight"`
	PartitionPriorityBonus    float64 `json:"partition_priority_bonus"`
	PartitionFactorMultiplier float64 `json:"partition_factor_multiplier"`

	// Utilization Analysis
	FactorUtilizationRate float64 `json:"factor_utilization_rate"`
	FactorEffectiveness   float64 `json:"factor_effectiveness"`
	FactorBalance         float64 `json:"factor_balance"`

	// Optimization Analysis
	OptimizationScore        float64 `json:"optimization_score"`
	BalanceRecommendation    string  `json:"balance_recommendation"`
	EfficiencyRecommendation string  `json:"efficiency_recommendation"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// QoSFactorAnalysis represents QoS-specific factor analysis
type QoSFactorAnalysis struct {
	QoSName string `json:"qos_name"`

	// QoS Factor Configuration
	QoSPriorityValue       uint64  `json:"qos_priority_value"`
	QoSWeight              float64 `json:"qos_weight"`
	QoSMultiplier          float64 `json:"qos_multiplier"`
	QoSFactorEffectiveness float64 `json:"qos_factor_effectiveness"`

	// Usage Analysis
	QoSUtilizationRate float64 `json:"qos_utilization_rate"`
	QoSFactorImpact    float64 `json:"qos_factor_impact"`
	QoSBalanceScore    float64 `json:"qos_balance_score"`

	// Performance Analysis
	QoSPerformanceScore      float64 `json:"qos_performance_score"`
	QoSOptimizationPotential float64 `json:"qos_optimization_potential"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// FactorTrendAnalysis represents temporal factor analysis
type FactorTrendAnalysis struct {
	FactorName     string `json:"factor_name"`
	AnalysisPeriod string `json:"analysis_period"`

	// Trend Metrics
	TrendDirection    string  `json:"trend_direction"`
	TrendVelocity     float64 `json:"trend_velocity"`
	TrendAcceleration float64 `json:"trend_acceleration"`
	TrendVolatility   float64 `json:"trend_volatility"`

	// Predictive Analysis
	FutureTrendPrediction string  `json:"future_trend_prediction"`
	PredictionConfidence  float64 `json:"prediction_confidence"`
	PredictionTimeframe   string  `json:"prediction_timeframe"`

	// Seasonal Analysis
	SeasonalityDetected bool    `json:"seasonality_detected"`
	SeasonalityStrength float64 `json:"seasonality_strength"`
	SeasonalityPeriod   string  `json:"seasonality_period"`

	// Cyclical Analysis
	CyclicalPatterns    []string `json:"cyclical_patterns"`
	CyclePredictability float64  `json:"cycle_predictability"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// FactorImpactAnalysis represents factor impact assessment
type FactorImpactAnalysis struct {
	FactorName string `json:"factor_name"`

	// Impact Metrics
	ImpactScore        float64 `json:"impact_score"`
	ImpactSignificance string  `json:"impact_significance"`
	ImpactCorrelation  float64 `json:"impact_correlation"`
	ImpactSensitivity  float64 `json:"impact_sensitivity"`

	// Change Analysis
	ImpactChangeRate float64 `json:"impact_change_rate"`
	ImpactStability  float64 `json:"impact_stability"`
	ImpactVolatility float64 `json:"impact_volatility"`

	// Cross-factor Analysis
	FactorInteractions map[string]float64 `json:"factor_interactions"`
	FactorSynergies    []string           `json:"factor_synergies"`
	FactorConflicts    []string           `json:"factor_conflicts"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// FactorOptimizationResult represents optimization analysis
type FactorOptimizationResult struct {
	OptimizationScope string `json:"optimization_scope"`

	// Current State
	CurrentConfiguration map[string]float64 `json:"current_configuration"`
	CurrentEffectiveness float64            `json:"current_effectiveness"`
	CurrentBalance       float64            `json:"current_balance"`

	// Recommended Configuration
	RecommendedConfiguration map[string]float64 `json:"recommended_configuration"`
	ExpectedEffectiveness    float64            `json:"expected_effectiveness"`
	ExpectedImprovement      float64            `json:"expected_improvement"`

	// Implementation Analysis
	ImplementationComplexity string   `json:"implementation_complexity"`
	ImplementationRisk       string   `json:"implementation_risk"`
	ImplementationTimeframe  string   `json:"implementation_timeframe"`
	RecommendedActions       []string `json:"recommended_actions"`

	OptimizedAt time.Time `json:"optimized_at"`
}

// FactorConfigurationValidation represents configuration validation
type FactorConfigurationValidation struct {
	ValidationScope string `json:"validation_scope"`

	// Validation Results
	IsValid            bool     `json:"is_valid"`
	ValidationScore    float64  `json:"validation_score"`
	ValidationIssues   []string `json:"validation_issues"`
	ValidationWarnings []string `json:"validation_warnings"`

	// Configuration Health
	ConfigurationHealth string  `json:"configuration_health"`
	HealthScore         float64 `json:"health_score"`
	StabilityScore      float64 `json:"stability_score"`
	OptimalityScore     float64 `json:"optimality_score"`

	// Recommendations
	ImmediateActions          []string `json:"immediate_actions"`
	LongTermActions           []string `json:"long_term_actions"`
	MonitoringRecommendations []string `json:"monitoring_recommendations"`

	ValidatedAt time.Time `json:"validated_at"`
}

// NewPriorityFactorsCollector creates a new priority factors collector
func NewPriorityFactorsCollector(client PriorityFactorsSLURMClient) *PriorityFactorsCollector {
	return &PriorityFactorsCollector{
		client: client,

		// Priority Factor Component Metrics
		priorityFactorWeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_weight",
				Help: "Weight assigned to each priority factor",
			},
			[]string{"factor", "scope", "partition", "qos"},
		),

		priorityFactorValue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_value",
				Help: "Raw value of each priority factor",
			},
			[]string{"job_id", "user", "account", "factor"},
		),

		priorityFactorContribution: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_contribution",
				Help: "Contribution of each factor to total priority",
			},
			[]string{"job_id", "user", "account", "factor"},
		),

		priorityFactorNormalized: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_normalized",
				Help: "Normalized priority factor values (0-1 scale)",
			},
			[]string{"job_id", "user", "account", "factor"},
		),

		priorityFactorEffective: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_effective",
				Help: "Effective priority factor after weighting",
			},
			[]string{"job_id", "user", "account", "factor"},
		),

		// Factor Weight Analysis
		factorWeightDistribution: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_priority_factor_weight_distribution",
				Help:    "Distribution of priority factor weights",
				Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
			[]string{"factor", "scope"},
		),

		factorWeightBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_weight_balance",
				Help: "Balance score for priority factor weights",
			},
			[]string{"scope", "configuration_version"},
		),

		factorWeightOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_weight_optimization_score",
				Help: "Optimization score for factor weight configuration",
			},
			[]string{"scope", "metric"},
		),

		factorWeightRecommendation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_weight_recommendation",
				Help: "Recommended factor weight values",
			},
			[]string{"factor", "scope", "recommendation_type"},
		),

		// Priority Factor Trends
		factorTrendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_trend_direction",
				Help: "Trend direction for priority factors",
			},
			[]string{"factor", "period", "trend"},
		),

		factorTrendVelocity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_trend_velocity",
				Help: "Rate of change for priority factors",
			},
			[]string{"factor", "period"},
		),

		factorTrendAcceleration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_trend_acceleration",
				Help: "Acceleration of priority factor trends",
			},
			[]string{"factor", "period"},
		),

		factorTrendVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_trend_volatility",
				Help: "Volatility of priority factor trends",
			},
			[]string{"factor", "period"},
		),

		factorTrendPrediction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_trend_prediction_confidence",
				Help: "Confidence in priority factor trend predictions",
			},
			[]string{"factor", "period", "prediction"},
		),

		// Factor Impact Analysis
		factorImpactScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_impact_score",
				Help: "Impact score of priority factors",
			},
			[]string{"factor", "impact_type"},
		),

		factorImpactChange: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_priority_factor_impact_changes_total",
				Help: "Total number of priority factor impact changes",
			},
			[]string{"factor", "change_type", "significance"},
		),

		factorImpactCorrelation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_impact_correlation",
				Help: "Correlation between priority factors",
			},
			[]string{"factor_a", "factor_b", "correlation_type"},
		),

		factorImpactSensitivity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_factor_impact_sensitivity",
				Help: "Sensitivity of priority factors to changes",
			},
			[]string{"factor", "sensitivity_type"},
		),

		// System Factor Configuration
		systemFactorConfig: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_factor_configuration",
				Help: "System-wide priority factor configuration",
			},
			[]string{"factor", "config_type", "version"},
		),

		systemFactorCalibration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_factor_calibration",
				Help: "System factor calibration metrics",
			},
			[]string{"metric", "version"},
		),

		systemFactorEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_factor_effectiveness",
				Help: "Effectiveness of system factor configuration",
			},
			[]string{"scope", "metric"},
		),

		systemFactorOptimality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_factor_optimality",
				Help: "Optimality score of factor configuration",
			},
			[]string{"scope", "optimization_type"},
		),

		// User-specific Factor Analysis
		userFactorProfile: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_factor_profile",
				Help: "User-specific priority factor profile",
			},
			[]string{"user", "account", "factor", "metric"},
		),

		userFactorDeviation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_factor_deviation",
				Help: "User deviation from normal factor patterns",
			},
			[]string{"user", "account", "deviation_type"},
		),

		userFactorOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_factor_optimization_score",
				Help: "User factor optimization score",
			},
			[]string{"user", "account", "optimization_type"},
		),

		userFactorRecommendation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_factor_recommendation_score",
				Help: "User factor recommendation relevance score",
			},
			[]string{"user", "account", "recommendation_type"},
		),

		// Account Factor Aggregation
		accountFactorSummary: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_factor_summary",
				Help: "Account-level factor summary metrics",
			},
			[]string{"account", "metric"},
		),

		accountFactorVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_factor_variance",
				Help: "Variance in account factor usage",
			},
			[]string{"account", "factor"},
		),

		accountFactorBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_factor_balance",
				Help: "Balance score for account factor usage",
			},
			[]string{"account"},
		),

		accountFactorEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_factor_efficiency",
				Help: "Efficiency score for account factor usage",
			},
			[]string{"account"},
		),

		// Partition Factor Analysis
		partitionFactorWeights: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_factor_weights",
				Help: "Partition-specific factor weights",
			},
			[]string{"partition", "factor"},
		),

		partitionFactorUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_factor_utilization",
				Help: "Partition factor utilization rate",
			},
			[]string{"partition", "metric"},
		),

		partitionFactorOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_factor_optimization_score",
				Help: "Partition factor optimization score",
			},
			[]string{"partition", "optimization_type"},
		),

		partitionFactorBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_factor_balance",
				Help: "Partition factor balance score",
			},
			[]string{"partition"},
		),

		// QoS Factor Tracking
		qosFactorPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_factor_priority",
				Help: "QoS factor priority values",
			},
			[]string{"qos", "metric"},
		),

		qosFactorMultiplier: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_factor_multiplier",
				Help: "QoS factor multiplier values",
			},
			[]string{"qos", "factor"},
		),

		qosFactorEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_factor_effectiveness",
				Help: "QoS factor effectiveness score",
			},
			[]string{"qos", "metric"},
		),

		qosFactorUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_factor_utilization",
				Help: "QoS factor utilization rate",
			},
			[]string{"qos"},
		),

		// Temporal Factor Analysis
		temporalFactorEvolution: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_temporal_factor_evolution",
				Help: "Evolution of factors over time",
			},
			[]string{"factor", "period", "evolution_type"},
		),

		temporalFactorCycles: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_temporal_factor_cycles",
				Help: "Cyclical patterns in factor behavior",
			},
			[]string{"factor", "cycle_type"},
		),

		temporalFactorSeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_temporal_factor_seasonality",
				Help: "Seasonal patterns in factor behavior",
			},
			[]string{"factor", "season", "metric"},
		),

		temporalFactorPredictability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_temporal_factor_predictability",
				Help: "Predictability score of factor behavior",
			},
			[]string{"factor", "timeframe"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *PriorityFactorsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.priorityFactorWeight.Describe(ch)
	c.priorityFactorValue.Describe(ch)
	c.priorityFactorContribution.Describe(ch)
	c.priorityFactorNormalized.Describe(ch)
	c.priorityFactorEffective.Describe(ch)
	c.factorWeightDistribution.Describe(ch)
	c.factorWeightBalance.Describe(ch)
	c.factorWeightOptimization.Describe(ch)
	c.factorWeightRecommendation.Describe(ch)
	c.factorTrendDirection.Describe(ch)
	c.factorTrendVelocity.Describe(ch)
	c.factorTrendAcceleration.Describe(ch)
	c.factorTrendVolatility.Describe(ch)
	c.factorTrendPrediction.Describe(ch)
	c.factorImpactScore.Describe(ch)
	c.factorImpactChange.Describe(ch)
	c.factorImpactCorrelation.Describe(ch)
	c.factorImpactSensitivity.Describe(ch)
	c.systemFactorConfig.Describe(ch)
	c.systemFactorCalibration.Describe(ch)
	c.systemFactorEffectiveness.Describe(ch)
	c.systemFactorOptimality.Describe(ch)
	c.userFactorProfile.Describe(ch)
	c.userFactorDeviation.Describe(ch)
	c.userFactorOptimization.Describe(ch)
	c.userFactorRecommendation.Describe(ch)
	c.accountFactorSummary.Describe(ch)
	c.accountFactorVariance.Describe(ch)
	c.accountFactorBalance.Describe(ch)
	c.accountFactorEfficiency.Describe(ch)
	c.partitionFactorWeights.Describe(ch)
	c.partitionFactorUtilization.Describe(ch)
	c.partitionFactorOptimization.Describe(ch)
	c.partitionFactorBalance.Describe(ch)
	c.qosFactorPriority.Describe(ch)
	c.qosFactorMultiplier.Describe(ch)
	c.qosFactorEffectiveness.Describe(ch)
	c.qosFactorUtilization.Describe(ch)
	c.temporalFactorEvolution.Describe(ch)
	c.temporalFactorCycles.Describe(ch)
	c.temporalFactorSeasonality.Describe(ch)
	c.temporalFactorPredictability.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *PriorityFactorsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Reset metrics
	c.resetMetrics()

	// Collect system factor weights
	c.collectSystemFactorWeights(ctx)

	// Get sample job IDs for factor breakdown
	sampleJobIDs := c.getSampleJobIDs()
	for _, jobID := range sampleJobIDs {
		c.collectJobFactorBreakdown(ctx, jobID)
	}

	// Get sample users for factor profiles
	sampleUsers := c.getSampleUsers()
	for _, user := range sampleUsers {
		c.collectUserFactorProfile(ctx, user)
	}

	// Get sample accounts for factor summaries
	sampleAccounts := c.getSampleAccounts()
	for _, account := range sampleAccounts {
		c.collectAccountFactorSummary(ctx, account)
	}

	// Get sample partitions for factor analysis
	samplePartitions := c.getSamplePartitions()
	for _, partition := range samplePartitions {
		c.collectPartitionFactorAnalysis(ctx, partition)
	}

	// Get sample QoS for factor analysis
	sampleQoS := c.getSampleQoS()
	for _, qos := range sampleQoS {
		c.collectQoSFactorAnalysis(ctx, qos)
	}

	// Collect factor trend analysis
	c.collectFactorTrendAnalysis(ctx)

	// Collect factor impact analysis
	c.collectFactorImpactAnalysis(ctx)

	// Collect optimization and validation metrics
	c.collectOptimizationMetrics(ctx)

	// Collect all metrics
	c.priorityFactorWeight.Collect(ch)
	c.priorityFactorValue.Collect(ch)
	c.priorityFactorContribution.Collect(ch)
	c.priorityFactorNormalized.Collect(ch)
	c.priorityFactorEffective.Collect(ch)
	c.factorWeightDistribution.Collect(ch)
	c.factorWeightBalance.Collect(ch)
	c.factorWeightOptimization.Collect(ch)
	c.factorWeightRecommendation.Collect(ch)
	c.factorTrendDirection.Collect(ch)
	c.factorTrendVelocity.Collect(ch)
	c.factorTrendAcceleration.Collect(ch)
	c.factorTrendVolatility.Collect(ch)
	c.factorTrendPrediction.Collect(ch)
	c.factorImpactScore.Collect(ch)
	c.factorImpactChange.Collect(ch)
	c.factorImpactCorrelation.Collect(ch)
	c.factorImpactSensitivity.Collect(ch)
	c.systemFactorConfig.Collect(ch)
	c.systemFactorCalibration.Collect(ch)
	c.systemFactorEffectiveness.Collect(ch)
	c.systemFactorOptimality.Collect(ch)
	c.userFactorProfile.Collect(ch)
	c.userFactorDeviation.Collect(ch)
	c.userFactorOptimization.Collect(ch)
	c.userFactorRecommendation.Collect(ch)
	c.accountFactorSummary.Collect(ch)
	c.accountFactorVariance.Collect(ch)
	c.accountFactorBalance.Collect(ch)
	c.accountFactorEfficiency.Collect(ch)
	c.partitionFactorWeights.Collect(ch)
	c.partitionFactorUtilization.Collect(ch)
	c.partitionFactorOptimization.Collect(ch)
	c.partitionFactorBalance.Collect(ch)
	c.qosFactorPriority.Collect(ch)
	c.qosFactorMultiplier.Collect(ch)
	c.qosFactorEffectiveness.Collect(ch)
	c.qosFactorUtilization.Collect(ch)
	c.temporalFactorEvolution.Collect(ch)
	c.temporalFactorCycles.Collect(ch)
	c.temporalFactorSeasonality.Collect(ch)
	c.temporalFactorPredictability.Collect(ch)
}

func (c *PriorityFactorsCollector) resetMetrics() {
	c.priorityFactorWeight.Reset()
	c.priorityFactorValue.Reset()
	c.priorityFactorContribution.Reset()
	c.priorityFactorNormalized.Reset()
	c.priorityFactorEffective.Reset()
	c.factorWeightBalance.Reset()
	c.factorWeightOptimization.Reset()
	c.factorWeightRecommendation.Reset()
	c.factorTrendDirection.Reset()
	c.factorTrendVelocity.Reset()
	c.factorTrendAcceleration.Reset()
	c.factorTrendVolatility.Reset()
	c.factorTrendPrediction.Reset()
	c.factorImpactScore.Reset()
	c.factorImpactCorrelation.Reset()
	c.factorImpactSensitivity.Reset()
	c.systemFactorConfig.Reset()
	c.systemFactorCalibration.Reset()
	c.systemFactorEffectiveness.Reset()
	c.systemFactorOptimality.Reset()
	c.userFactorProfile.Reset()
	c.userFactorDeviation.Reset()
	c.userFactorOptimization.Reset()
	c.userFactorRecommendation.Reset()
	c.accountFactorSummary.Reset()
	c.accountFactorVariance.Reset()
	c.accountFactorBalance.Reset()
	c.accountFactorEfficiency.Reset()
	c.partitionFactorWeights.Reset()
	c.partitionFactorUtilization.Reset()
	c.partitionFactorOptimization.Reset()
	c.partitionFactorBalance.Reset()
	c.qosFactorPriority.Reset()
	c.qosFactorMultiplier.Reset()
	c.qosFactorEffectiveness.Reset()
	c.qosFactorUtilization.Reset()
	c.temporalFactorEvolution.Reset()
	c.temporalFactorCycles.Reset()
	c.temporalFactorSeasonality.Reset()
	c.temporalFactorPredictability.Reset()
}

func (c *PriorityFactorsCollector) collectSystemFactorWeights(ctx context.Context) {
	weights, err := c.client.GetSystemFactorWeights(ctx)
	if err != nil {
		log.Printf("Error collecting system factor weights: %v", err)
		return
	}

	// System factor configuration
	c.systemFactorConfig.WithLabelValues("age", "weight", weights.ConfigurationVersion).Set(weights.AgeWeightGlobal)
	c.systemFactorConfig.WithLabelValues("fairshare", "weight", weights.ConfigurationVersion).Set(weights.FairShareWeightGlobal)
	c.systemFactorConfig.WithLabelValues("qos", "weight", weights.ConfigurationVersion).Set(weights.QoSWeightGlobal)
	c.systemFactorConfig.WithLabelValues("partition", "weight", weights.ConfigurationVersion).Set(weights.PartitionWeightGlobal)
	c.systemFactorConfig.WithLabelValues("size", "weight", weights.ConfigurationVersion).Set(weights.SizeWeightGlobal)
	c.systemFactorConfig.WithLabelValues("assoc", "weight", weights.ConfigurationVersion).Set(weights.AssocWeightGlobal)
	c.systemFactorConfig.WithLabelValues("nice", "weight", weights.ConfigurationVersion).Set(weights.NiceWeightGlobal)

	// Weight balance and optimization
	c.factorWeightBalance.WithLabelValues("global", weights.ConfigurationVersion).Set(weights.WeightBalance)
	c.factorWeightOptimization.WithLabelValues("global", "optimality").Set(weights.WeightOptimality)
	c.factorWeightOptimization.WithLabelValues("global", "effectiveness").Set(weights.WeightEffectiveness)

	// Calibration metrics
	c.systemFactorCalibration.WithLabelValues("accuracy", weights.ConfigurationVersion).Set(weights.CalibrationAccuracy)
	c.systemFactorCalibration.WithLabelValues("stability", weights.ConfigurationVersion).Set(weights.CalibrationStability)

	// Weight distribution
	c.factorWeightDistribution.WithLabelValues("age", "global").Observe(weights.AgeWeightGlobal)
	c.factorWeightDistribution.WithLabelValues("fairshare", "global").Observe(weights.FairShareWeightGlobal)
	c.factorWeightDistribution.WithLabelValues("qos", "global").Observe(weights.QoSWeightGlobal)
	c.factorWeightDistribution.WithLabelValues("partition", "global").Observe(weights.PartitionWeightGlobal)
	c.factorWeightDistribution.WithLabelValues("size", "global").Observe(weights.SizeWeightGlobal)
}

func (c *PriorityFactorsCollector) collectJobFactorBreakdown(ctx context.Context, jobID string) {
	breakdown, err := c.client.GetPriorityFactorBreakdown(ctx, jobID)
	if err != nil {
		log.Printf("Error collecting factor breakdown for job %s: %v", jobID, err)
		return
	}

	labels := []string{breakdown.JobID, breakdown.UserName, breakdown.AccountName}

	// Factor values
	c.priorityFactorValue.WithLabelValues(append(labels, "age")...).Set(breakdown.AgeValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "fairshare")...).Set(breakdown.FairShareValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "qos")...).Set(breakdown.QoSValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "partition")...).Set(breakdown.PartitionValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "size")...).Set(breakdown.SizeValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "assoc")...).Set(breakdown.AssocValue)
	c.priorityFactorValue.WithLabelValues(append(labels, "nice")...).Set(breakdown.NiceValue)

	// Factor contributions
	c.priorityFactorContribution.WithLabelValues(append(labels, "age")...).Set(breakdown.AgeContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "fairshare")...).Set(breakdown.FairShareContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "qos")...).Set(breakdown.QoSContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "partition")...).Set(breakdown.PartitionContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "size")...).Set(breakdown.SizeContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "assoc")...).Set(breakdown.AssocContribution)
	c.priorityFactorContribution.WithLabelValues(append(labels, "nice")...).Set(breakdown.NiceContribution)

	// Normalized values
	c.priorityFactorNormalized.WithLabelValues(append(labels, "age")...).Set(breakdown.AgeNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "fairshare")...).Set(breakdown.FairShareNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "qos")...).Set(breakdown.QoSNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "partition")...).Set(breakdown.PartitionNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "size")...).Set(breakdown.SizeNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "assoc")...).Set(breakdown.AssocNormalized)
	c.priorityFactorNormalized.WithLabelValues(append(labels, "nice")...).Set(breakdown.NiceNormalized)

	// Effective contributions
	c.priorityFactorEffective.WithLabelValues(append(labels, "age")...).Set(breakdown.AgeEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "fairshare")...).Set(breakdown.FairShareEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "qos")...).Set(breakdown.QoSEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "partition")...).Set(breakdown.PartitionEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "size")...).Set(breakdown.SizeEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "assoc")...).Set(breakdown.AssocEffective)
	c.priorityFactorEffective.WithLabelValues(append(labels, "nice")...).Set(breakdown.NiceEffective)

	// Weights
	weightLabels := []string{"job", breakdown.PartitionName, breakdown.QoSName}
	c.priorityFactorWeight.WithLabelValues(append([]string{"age"}, weightLabels...)...).Set(breakdown.AgeWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"fairshare"}, weightLabels...)...).Set(breakdown.FairShareWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"qos"}, weightLabels...)...).Set(breakdown.QoSWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"partition"}, weightLabels...)...).Set(breakdown.PartitionWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"size"}, weightLabels...)...).Set(breakdown.SizeWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"assoc"}, weightLabels...)...).Set(breakdown.AssocWeight)
	c.priorityFactorWeight.WithLabelValues(append([]string{"nice"}, weightLabels...)...).Set(breakdown.NiceWeight)
}

func (c *PriorityFactorsCollector) collectUserFactorProfile(ctx context.Context, userName string) {
	profile, err := c.client.GetUserFactorProfile(ctx, userName)
	if err != nil {
		log.Printf("Error collecting user factor profile for %s: %v", userName, err)
		return
	}

	labels := []string{profile.UserName, profile.AccountName}

	// User factor averages
	c.userFactorProfile.WithLabelValues(append(labels, []string{"age", "average"}...)...).Set(profile.AvgAgeContribution)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"fairshare", "average"}...)...).Set(profile.AvgFairShareContribution)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"qos", "average"}...)...).Set(profile.AvgQoSContribution)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"partition", "average"}...)...).Set(profile.AvgPartitionContribution)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"size", "average"}...)...).Set(profile.AvgSizeContribution)

	// User factor variances
	c.userFactorProfile.WithLabelValues(append(labels, []string{"age", "variance"}...)...).Set(profile.AgeVariance)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"fairshare", "variance"}...)...).Set(profile.FairShareVariance)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"qos", "variance"}...)...).Set(profile.QoSVariance)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"partition", "variance"}...)...).Set(profile.PartitionVariance)
	c.userFactorProfile.WithLabelValues(append(labels, []string{"size", "variance"}...)...).Set(profile.SizeVariance)

	// User behavior metrics
	c.userFactorProfile.WithLabelValues(append(labels, []string{"behavior", "stability"}...)...).Set(profile.FactorStability)
	c.userFactorOptimization.WithLabelValues(append(labels, "score")...).Set(profile.FactorOptimizationScore)
	c.userFactorDeviation.WithLabelValues(append(labels, "from_norm")...).Set(profile.DeviationFromNorm)
}

func (c *PriorityFactorsCollector) collectAccountFactorSummary(ctx context.Context, accountName string) {
	summary, err := c.client.GetAccountFactorSummary(ctx, accountName)
	if err != nil {
		log.Printf("Error collecting account factor summary for %s: %v", accountName, err)
		return
	}

	c.accountFactorSummary.WithLabelValues(summary.AccountName, "total_users").Set(float64(summary.TotalUsers))
	c.accountFactorSummary.WithLabelValues(summary.AccountName, "active_users").Set(float64(summary.ActiveUsers))
	c.accountFactorBalance.WithLabelValues(summary.AccountName).Set(summary.AvgFactorBalance)
	c.accountFactorSummary.WithLabelValues(summary.AccountName, "variance_score").Set(summary.FactorVarianceScore)
	c.accountFactorEfficiency.WithLabelValues(summary.AccountName).Set(summary.FactorEfficiencyScore)
	c.accountFactorSummary.WithLabelValues(summary.AccountName, "optimization_potential").Set(summary.OptimizationPotential)
}

func (c *PriorityFactorsCollector) collectPartitionFactorAnalysis(ctx context.Context, partition string) {
	analysis, err := c.client.GetPartitionFactorAnalysis(ctx, partition)
	if err != nil {
		log.Printf("Error collecting partition factor analysis for %s: %v", partition, err)
		return
	}

	c.partitionFactorWeights.WithLabelValues(partition, "partition_weight").Set(analysis.PartitionWeight)
	c.partitionFactorWeights.WithLabelValues(partition, "priority_bonus").Set(analysis.PartitionPriorityBonus)
	c.partitionFactorWeights.WithLabelValues(partition, "multiplier").Set(analysis.PartitionFactorMultiplier)

	c.partitionFactorUtilization.WithLabelValues(partition, "utilization_rate").Set(analysis.FactorUtilizationRate)
	c.partitionFactorUtilization.WithLabelValues(partition, "effectiveness").Set(analysis.FactorEffectiveness)
	c.partitionFactorBalance.WithLabelValues(partition).Set(analysis.FactorBalance)
	c.partitionFactorOptimization.WithLabelValues(partition, "optimization_score").Set(analysis.OptimizationScore)
}

func (c *PriorityFactorsCollector) collectQoSFactorAnalysis(ctx context.Context, qosName string) {
	analysis, err := c.client.GetQoSFactorAnalysis(ctx, qosName)
	if err != nil {
		log.Printf("Error collecting QoS factor analysis for %s: %v", qosName, err)
		return
	}

	c.qosFactorPriority.WithLabelValues(qosName, "priority_value").Set(float64(analysis.QoSPriorityValue))
	c.qosFactorPriority.WithLabelValues(qosName, "weight").Set(analysis.QoSWeight)
	c.qosFactorMultiplier.WithLabelValues(qosName, "multiplier").Set(analysis.QoSMultiplier)
	c.qosFactorEffectiveness.WithLabelValues(qosName, "effectiveness").Set(analysis.QoSFactorEffectiveness)
	c.qosFactorUtilization.WithLabelValues(qosName).Set(analysis.QoSUtilizationRate)
	c.qosFactorEffectiveness.WithLabelValues(qosName, "impact").Set(analysis.QoSFactorImpact)
	c.qosFactorEffectiveness.WithLabelValues(qosName, "balance").Set(analysis.QoSBalanceScore)
	c.qosFactorEffectiveness.WithLabelValues(qosName, "performance").Set(analysis.QoSPerformanceScore)
}

func (c *PriorityFactorsCollector) collectFactorTrendAnalysis(ctx context.Context) {
	factors := []string{"age", "fairshare", "qos", "partition", "size", "assoc", "nice"}
	period := "24h"

	for _, factor := range factors {
		trend, err := c.client.GetFactorTrendAnalysis(ctx, factor, period)
		if err != nil {
			continue
		}

		c.factorTrendDirection.WithLabelValues(factor, period, trend.TrendDirection).Set(1.0)
		c.factorTrendVelocity.WithLabelValues(factor, period).Set(trend.TrendVelocity)
		c.factorTrendAcceleration.WithLabelValues(factor, period).Set(trend.TrendAcceleration)
		c.factorTrendVolatility.WithLabelValues(factor, period).Set(trend.TrendVolatility)
		c.factorTrendPrediction.WithLabelValues(factor, period, trend.FutureTrendPrediction).Set(trend.PredictionConfidence)

		if trend.SeasonalityDetected {
			c.temporalFactorSeasonality.WithLabelValues(factor, trend.SeasonalityPeriod, "strength").Set(trend.SeasonalityStrength)
		}

		c.temporalFactorPredictability.WithLabelValues(factor, trend.PredictionTimeframe).Set(trend.CyclePredictability)
	}
}

func (c *PriorityFactorsCollector) collectFactorImpactAnalysis(ctx context.Context) {
	factors := []string{"age", "fairshare", "qos", "partition", "size", "assoc", "nice"}

	for _, factor := range factors {
		impact, err := c.client.GetFactorImpactAnalysis(ctx, factor)
		if err != nil {
			continue
		}

		c.factorImpactScore.WithLabelValues(factor, "score").Set(impact.ImpactScore)
		c.factorImpactScore.WithLabelValues(factor, "correlation").Set(impact.ImpactCorrelation)
		c.factorImpactSensitivity.WithLabelValues(factor, "sensitivity").Set(impact.ImpactSensitivity)
		c.factorImpactSensitivity.WithLabelValues(factor, "stability").Set(impact.ImpactStability)
		c.factorImpactSensitivity.WithLabelValues(factor, "volatility").Set(impact.ImpactVolatility)

		// Cross-factor correlations
		for otherFactor, correlation := range impact.FactorInteractions {
			c.factorImpactCorrelation.WithLabelValues(factor, otherFactor, "interaction").Set(correlation)
		}
	}
}

func (c *PriorityFactorsCollector) collectOptimizationMetrics(ctx context.Context) {
	// System optimization
	systemOptimization, err := c.client.OptimizeFactorWeights(ctx, "system")
	if err == nil {
		c.systemFactorOptimality.WithLabelValues("system", "current_effectiveness").Set(systemOptimization.CurrentEffectiveness)
		c.systemFactorOptimality.WithLabelValues("system", "expected_effectiveness").Set(systemOptimization.ExpectedEffectiveness)
		c.systemFactorOptimality.WithLabelValues("system", "expected_improvement").Set(systemOptimization.ExpectedImprovement)
	}

	// Configuration validation
	validation, err := c.client.ValidateFactorConfiguration(ctx)
	if err == nil {
		c.systemFactorEffectiveness.WithLabelValues("system", "validation_score").Set(validation.ValidationScore)
		c.systemFactorEffectiveness.WithLabelValues("system", "health_score").Set(validation.HealthScore)
		c.systemFactorEffectiveness.WithLabelValues("system", "stability_score").Set(validation.StabilityScore)
		c.systemFactorEffectiveness.WithLabelValues("system", "optimality_score").Set(validation.OptimalityScore)
	}
}

// Simplified mock data generators for testing purposes
func (c *PriorityFactorsCollector) getSampleJobIDs() []string {
	return []string{"12345", "12346", "12347"}
}

func (c *PriorityFactorsCollector) getSampleUsers() []string {
	return []string{"user1", "user2", "user3"}
}

func (c *PriorityFactorsCollector) getSampleAccounts() []string {
	return []string{"account1", "account2", "account3"}
}

func (c *PriorityFactorsCollector) getSamplePartitions() []string {
	return []string{"cpu", "gpu", "bigmem"}
}

func (c *PriorityFactorsCollector) getSampleQoS() []string {
	return []string{"normal", "high", "low"}
}
