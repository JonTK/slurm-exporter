// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AccountUsagePatternsSLURMClient interface {
	GetAccountUsagePatterns(ctx context.Context, accountName string) (*AccountUsagePatterns, error)
	ListAccounts(ctx context.Context) ([]*AccountInfo, error)
	GetAccountAnomalies(ctx context.Context, accountName string) ([]*AccountAnomaly, error)
	GetAccountBehaviorProfile(ctx context.Context, accountName string) (*AccountBehaviorProfile, error)
	GetAccountTrendAnalysis(ctx context.Context, accountName string) (*AccountTrendAnalysis, error)
	GetAccountSeasonality(ctx context.Context, accountName string) (*AccountSeasonality, error)
	GetAccountPredictions(ctx context.Context, accountName string) (*AccountPredictions, error)
	GetAccountClassification(ctx context.Context, accountName string) (*AccountClassification, error)
	GetAccountComparison(ctx context.Context, accountA, accountB string) (*AccountComparison, error)
	GetAccountRiskAnalysis(ctx context.Context, accountName string) (*AccountRiskAnalysis, error)
	GetAccountOptimizationSuggestions(ctx context.Context, accountName string) (*AccountOptimizationSuggestions, error)
	GetAccountBenchmarks(ctx context.Context, accountName string) (*AccountBenchmarks, error)
	GetSystemUsageOverview(ctx context.Context) (*SystemUsageOverview, error)
}

type AccountUsagePatterns struct {
	AccountName          string                 `json:"account_name"`
	UsageTimeSeries      []*TimeSeriesPoint     `json:"usage_time_series"`
	DailyPatterns        []*DailyUsagePattern   `json:"daily_patterns"`
	WeeklyPatterns       []*WeeklyUsagePattern  `json:"weekly_patterns"`
	MonthlyPatterns      []*MonthlyUsagePattern `json:"monthly_patterns"`
	ResourceDistribution *ResourceDistribution  `json:"resource_distribution"`
	JobSubmissionProfile *JobSubmissionProfile  `json:"job_submission_profile"`
	UserActivityProfile  *UserActivityProfile   `json:"user_activity_profile"`
	CostAnalysis         *CostAnalysis          `json:"cost_analysis"`
	EfficiencyMetrics    *EfficiencyMetrics     `json:"efficiency_metrics"`
	LastUpdated          time.Time              `json:"last_updated"`
}

// Note: TimeSeriesPoint type is defined in common_types.go

type DailyUsagePattern struct {
	Hour             int     `json:"hour"`
	AverageCPUHours  float64 `json:"average_cpu_hours"`
	AverageMemoryGB  float64 `json:"average_memory_gb"`
	AverageGPUHours  float64 `json:"average_gpu_hours"`
	AverageJobCount  float64 `json:"average_job_count"`
	PeakUsageHour    bool    `json:"peak_usage_hour"`
	OffPeakHour      bool    `json:"off_peak_hour"`
	UtilizationScore float64 `json:"utilization_score"`
}

type WeeklyUsagePattern struct {
	DayOfWeek        int     `json:"day_of_week"`
	AverageCPUHours  float64 `json:"average_cpu_hours"`
	AverageMemoryGB  float64 `json:"average_memory_gb"`
	AverageGPUHours  float64 `json:"average_gpu_hours"`
	AverageJobCount  float64 `json:"average_job_count"`
	WeekdayPattern   bool    `json:"weekday_pattern"`
	WeekendPattern   bool    `json:"weekend_pattern"`
	UtilizationScore float64 `json:"utilization_score"`
}

type MonthlyUsagePattern struct {
	Month            int     `json:"month"`
	AverageCPUHours  float64 `json:"average_cpu_hours"`
	AverageMemoryGB  float64 `json:"average_memory_gb"`
	AverageGPUHours  float64 `json:"average_gpu_hours"`
	AverageJobCount  float64 `json:"average_job_count"`
	SeasonalTrend    string  `json:"seasonal_trend"`
	GrowthRate       float64 `json:"growth_rate"`
	UtilizationScore float64 `json:"utilization_score"`
}

type ResourceDistribution struct {
	CPUDistribution      *ResourceUsageDistribution `json:"cpu_distribution"`
	MemoryDistribution   *ResourceUsageDistribution `json:"memory_distribution"`
	GPUDistribution      *ResourceUsageDistribution `json:"gpu_distribution"`
	JobSizeDistribution  *JobSizeDistribution       `json:"job_size_distribution"`
	DurationDistribution *DurationDistribution      `json:"duration_distribution"`
}

type ResourceUsageDistribution struct {
	Percentile25 float64 `json:"percentile_25"`
	Percentile50 float64 `json:"percentile_50"`
	Percentile75 float64 `json:"percentile_75"`
	Percentile90 float64 `json:"percentile_90"`
	Percentile95 float64 `json:"percentile_95"`
	Percentile99 float64 `json:"percentile_99"`
	Mean         float64 `json:"mean"`
	StdDev       float64 `json:"std_dev"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
}

type JobSizeDistribution struct {
	SmallJobs  float64 `json:"small_jobs"`
	MediumJobs float64 `json:"medium_jobs"`
	LargeJobs  float64 `json:"large_jobs"`
	XLargeJobs float64 `json:"xlarge_jobs"`
	SingleCore float64 `json:"single_core"`
	MultiCore  float64 `json:"multi_core"`
	HighMemory float64 `json:"high_memory"`
	GPUJobs    float64 `json:"gpu_jobs"`
}

type DurationDistribution struct {
	ShortJobs       float64 `json:"short_jobs"`
	MediumJobs      float64 `json:"medium_jobs"`
	LongJobs        float64 `json:"long_jobs"`
	VeryLongJobs    float64 `json:"very_long_jobs"`
	InteractiveJobs float64 `json:"interactive_jobs"`
	BatchJobs       float64 `json:"batch_jobs"`
}

type JobSubmissionProfile struct {
	SubmissionPatterns  []*SubmissionPattern `json:"submission_patterns"`
	BurstPatterns       []*BurstPattern      `json:"burst_patterns"`
	QueueingBehavior    *QueueingBehavior    `json:"queueing_behavior"`
	RetryPatterns       *RetryPatterns       `json:"retry_patterns"`
	CancellationPattern *CancellationPattern `json:"cancellation_pattern"`
}

type SubmissionPattern struct {
	TimeWindow         string  `json:"time_window"`
	AverageSubmissions float64 `json:"average_submissions"`
	PeakSubmissions    int64   `json:"peak_submissions"`
	SubmissionRate     float64 `json:"submission_rate"`
	Regularity         float64 `json:"regularity"`
}

type BurstPattern struct {
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	JobCount       int64     `json:"job_count"`
	TriggerEvent   string    `json:"trigger_event"`
	BurstIntensity float64   `json:"burst_intensity"`
}

type QueueingBehavior struct {
	AverageQueueTime     float64 `json:"average_queue_time"`
	QueueTimeVariance    float64 `json:"queue_time_variance"`
	QueueJumpFrequency   float64 `json:"queue_jump_frequency"`
	PriorityOptimization float64 `json:"priority_optimization"`
}

type RetryPatterns struct {
	RetryFrequency     float64 `json:"retry_frequency"`
	AverageRetryDelay  float64 `json:"average_retry_delay"`
	SuccessAfterRetry  float64 `json:"success_after_retry"`
	RetryEffectiveness float64 `json:"retry_effectiveness"`
}

type CancellationPattern struct {
	CancellationRate    float64  `json:"cancellation_rate"`
	EarlyCancellations  float64  `json:"early_cancellations"`
	LateCancellations   float64  `json:"late_cancellations"`
	CancellationReasons []string `json:"cancellation_reasons"`
	PreemptionRate      float64  `json:"preemption_rate"`
}

type UserActivityProfile struct {
	ActiveUsers        int64               `json:"active_users"`
	PowerUsers         int64               `json:"power_users"`
	OccasionalUsers    int64               `json:"occasional_users"`
	NewUsers           int64               `json:"new_users"`
	UserRetention      float64             `json:"user_retention"`
	UserGrowthRate     float64             `json:"user_growth_rate"`
	UserCollaboration  *UserCollaboration  `json:"user_collaboration"`
	UserSegmentation   *UserSegmentation   `json:"user_segmentation"`
	UserBehaviorTrends *UserBehaviorTrends `json:"user_behavior_trends"`
}

type UserCollaboration struct {
	SharedProjects    int64   `json:"shared_projects"`
	CollaborationRate float64 `json:"collaboration_rate"`
	ResourceSharing   float64 `json:"resource_sharing"`
	KnowledgeTransfer float64 `json:"knowledge_transfer"`
}

type UserSegmentation struct {
	ResearchUsers     int64   `json:"research_users"`
	DevelopmentUsers  int64   `json:"development_users"`
	ProductionUsers   int64   `json:"production_users"`
	EducationalUsers  int64   `json:"educational_users"`
	ExternalUsers     int64   `json:"external_users"`
	SegmentationScore float64 `json:"segmentation_score"`
}

type UserBehaviorTrends struct {
	SkillProgression    float64 `json:"skill_progression"`
	LearningCurve       float64 `json:"learning_curve"`
	AdaptationRate      float64 `json:"adaptation_rate"`
	EfficiencyGrowth    float64 `json:"efficiency_growth"`
	BehaviorStability   float64 `json:"behavior_stability"`
	TrendPredictability float64 `json:"trend_predictability"`
}

type CostAnalysis struct {
	TotalCost        float64           `json:"total_cost"`
	CostPerCPUHour   float64           `json:"cost_per_cpu_hour"`
	CostPerGPUHour   float64           `json:"cost_per_gpu_hour"`
	CostPerJob       float64           `json:"cost_per_job"`
	CostPerUser      float64           `json:"cost_per_user"`
	CostTrends       *CostTrends       `json:"cost_trends"`
	CostOptimization *CostOptimization `json:"cost_optimization"`
	BudgetAnalysis   *BudgetAnalysis   `json:"budget_analysis"`
	CostPrediction   *CostPrediction   `json:"cost_prediction"`
}

type CostTrends struct {
	DailyCostTrend   float64 `json:"daily_cost_trend"`
	WeeklyCostTrend  float64 `json:"weekly_cost_trend"`
	MonthlyCostTrend float64 `json:"monthly_cost_trend"`
	CostGrowthRate   float64 `json:"cost_growth_rate"`
	CostVolatility   float64 `json:"cost_volatility"`
}

type CostOptimization struct {
	PotentialSavings         float64  `json:"potential_savings"`
	OptimizationAreas        []string `json:"optimization_areas"`
	RightSizingOpportunities float64  `json:"right_sizing_opportunities"`
	SchedulingOptimization   float64  `json:"scheduling_optimization"`
	ResourcePooling          float64  `json:"resource_pooling"`
}

type BudgetAnalysis struct {
	BudgetUtilization float64 `json:"budget_utilization"`
	BudgetRemaining   float64 `json:"budget_remaining"`
	BurnRate          float64 `json:"burn_rate"`
	ProjectedOverrun  float64 `json:"projected_overrun"`
	BudgetRisk        string  `json:"budget_risk"`
}

type CostPrediction struct {
	NextMonthCost      float64 `json:"next_month_cost"`
	NextQuarterCost    float64 `json:"next_quarter_cost"`
	NextYearCost       float64 `json:"next_year_cost"`
	PredictionAccuracy float64 `json:"prediction_accuracy"`
	ConfidenceInterval float64 `json:"confidence_interval"`
}

// Note: EfficiencyMetrics, ResourceWaste, and PerformanceMetrics types are defined in common_types.go

type AccountInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UserCount   int64     `json:"user_count"`
	JobCount    int64     `json:"job_count"`
}

type AccountAnomaly struct {
	AccountName     string     `json:"account_name"`
	AnomalyType     string     `json:"anomaly_type"`
	Severity        string     `json:"severity"`
	DetectedAt      time.Time  `json:"detected_at"`
	Description     string     `json:"description"`
	AffectedMetrics []string   `json:"affected_metrics"`
	DeviationScore  float64    `json:"deviation_score"`
	Impact          string     `json:"impact"`
	Recommendation  string     `json:"recommendation"`
	RootCause       string     `json:"root_cause"`
	IsResolved      bool       `json:"is_resolved"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
}

type AccountBehaviorProfile struct {
	AccountName         string  `json:"account_name"`
	BehaviorType        string  `json:"behavior_type"`
	ActivityLevel       string  `json:"activity_level"`
	ResourcePreference  string  `json:"resource_preference"`
	SchedulingPattern   string  `json:"scheduling_pattern"`
	CollaborationLevel  string  `json:"collaboration_level"`
	AdaptabilityScore   float64 `json:"adaptability_score"`
	ConsistencyScore    float64 `json:"consistency_score"`
	EfficiencyScore     float64 `json:"efficiency_score"`
	InnovationScore     float64 `json:"innovation_score"`
	RiskTolerance       string  `json:"risk_tolerance"`
	LearningRate        float64 `json:"learning_rate"`
	BehaviorStability   float64 `json:"behavior_stability"`
	PredictabilityScore float64 `json:"predictability_score"`
}

type AccountTrendAnalysis struct {
	AccountName         string               `json:"account_name"`
	TrendDirection      string               `json:"trend_direction"`
	GrowthRate          float64              `json:"growth_rate"`
	Seasonality         *Seasonality         `json:"seasonality"`
	CyclicalPatterns    []*CyclicalPattern   `json:"cyclical_patterns"`
	TrendConfidence     float64              `json:"trend_confidence"`
	ForecastAccuracy    float64              `json:"forecast_accuracy"`
	InflectionPoints    []*InflectionPoint   `json:"inflection_points"`
	OutlierDetection    *OutlierDetection    `json:"outlier_detection"`
	ChangePointAnalysis *ChangePointAnalysis `json:"change_point_analysis"`
}

type Seasonality struct {
	HasSeasonality   bool      `json:"has_seasonality"`
	SeasonalPeriod   int       `json:"seasonal_period"`
	SeasonalStrength float64   `json:"seasonal_strength"`
	SeasonalFactors  []float64 `json:"seasonal_factors"`
	PeakSeason       string    `json:"peak_season"`
	LowSeason        string    `json:"low_season"`
}

type CyclicalPattern struct {
	PatternType     string    `json:"pattern_type"`
	CycleDuration   int       `json:"cycle_duration"`
	CycleStrength   float64   `json:"cycle_strength"`
	LastCycleStart  time.Time `json:"last_cycle_start"`
	NextCycleStart  time.Time `json:"next_cycle_start"`
	CycleRegularity float64   `json:"cycle_regularity"`
}

type InflectionPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	ChangeType      string    `json:"change_type"`
	ChangeIntensity float64   `json:"change_intensity"`
	AffectedMetrics []string  `json:"affected_metrics"`
	CauseAnalysis   string    `json:"cause_analysis"`
}

type OutlierDetection struct {
	OutlierCount        int64     `json:"outlier_count"`
	OutlierRate         float64   `json:"outlier_rate"`
	OutlierTypes        []string  `json:"outlier_types"`
	LastOutlierDetected time.Time `json:"last_outlier_detected"`
	OutlierTrend        string    `json:"outlier_trend"`
}

type ChangePointAnalysis struct {
	ChangePointCount     int64     `json:"change_point_count"`
	LastChangePoint      time.Time `json:"last_change_point"`
	ChangePointFrequency float64   `json:"change_point_frequency"`
	ChangePointTypes     []string  `json:"change_point_types"`
	ChangePointImpact    float64   `json:"change_point_impact"`
}

type AccountSeasonality struct {
	AccountName        string               `json:"account_name"`
	DailySeasonality   *DailySeasonality    `json:"daily_seasonality"`
	WeeklySeasonality  *WeeklySeasonality   `json:"weekly_seasonality"`
	MonthlySeasonality *MonthlySeasonality  `json:"monthly_seasonality"`
	YearlySeasonality  *YearlySeasonality   `json:"yearly_seasonality"`
	CustomSeasonality  []*CustomSeasonality `json:"custom_seasonality"`
	SeasonalityScore   float64              `json:"seasonality_score"`
	PredictablePeriods []string             `json:"predictable_periods"`
	VolatilePeriods    []string             `json:"volatile_periods"`
}

type DailySeasonality struct {
	PeakHours           []int   `json:"peak_hours"`
	OffPeakHours        []int   `json:"off_peak_hours"`
	SeasonalityStrength float64 `json:"seasonality_strength"`
	RegularityScore     float64 `json:"regularity_score"`
}

type WeeklySeasonality struct {
	PeakDays            []int   `json:"peak_days"`
	OffPeakDays         []int   `json:"off_peak_days"`
	WeekendPattern      bool    `json:"weekend_pattern"`
	SeasonalityStrength float64 `json:"seasonality_strength"`
	RegularityScore     float64 `json:"regularity_score"`
}

type MonthlySeasonality struct {
	PeakMonths          []int   `json:"peak_months"`
	OffPeakMonths       []int   `json:"off_peak_months"`
	QuarterlyPattern    bool    `json:"quarterly_pattern"`
	SeasonalityStrength float64 `json:"seasonality_strength"`
	RegularityScore     float64 `json:"regularity_score"`
}

type YearlySeasonality struct {
	AcademicCalendar    bool    `json:"academic_calendar"`
	FiscalYearPattern   bool    `json:"fiscal_year_pattern"`
	SeasonalityStrength float64 `json:"seasonality_strength"`
	RegularityScore     float64 `json:"regularity_score"`
}

type CustomSeasonality struct {
	PatternName        string  `json:"pattern_name"`
	PatternDescription string  `json:"pattern_description"`
	PatternPeriod      int     `json:"pattern_period"`
	PatternStrength    float64 `json:"pattern_strength"`
	PatternRegularity  float64 `json:"pattern_regularity"`
}

type AccountPredictions struct {
	AccountName           string                 `json:"account_name"`
	ShortTermPredictions  *ShortTermPredictions  `json:"short_term_predictions"`
	MediumTermPredictions *MediumTermPredictions `json:"medium_term_predictions"`
	LongTermPredictions   *LongTermPredictions   `json:"long_term_predictions"`
	PredictionAccuracy    *PredictionAccuracy    `json:"prediction_accuracy"`
	UncertaintyAnalysis   *UncertaintyAnalysis   `json:"uncertainty_analysis"`
	ScenarioAnalysis      *ScenarioAnalysis      `json:"scenario_analysis"`
}

type ShortTermPredictions struct {
	NextDayUsage      *UsagePrediction `json:"next_day_usage"`
	NextWeekUsage     *UsagePrediction `json:"next_week_usage"`
	PredictionHorizon int              `json:"prediction_horizon"`
	ConfidenceLevel   float64          `json:"confidence_level"`
}

type MediumTermPredictions struct {
	NextMonthUsage    *UsagePrediction `json:"next_month_usage"`
	NextQuarterUsage  *UsagePrediction `json:"next_quarter_usage"`
	PredictionHorizon int              `json:"prediction_horizon"`
	ConfidenceLevel   float64          `json:"confidence_level"`
}

type LongTermPredictions struct {
	NextYearUsage     *UsagePrediction `json:"next_year_usage"`
	MultiYearTrend    *UsagePrediction `json:"multi_year_trend"`
	PredictionHorizon int              `json:"prediction_horizon"`
	ConfidenceLevel   float64          `json:"confidence_level"`
}

type UsagePrediction struct {
	PredictedCPUHours  float64 `json:"predicted_cpu_hours"`
	PredictedMemoryGB  float64 `json:"predicted_memory_gb"`
	PredictedGPUHours  float64 `json:"predicted_gpu_hours"`
	PredictedJobCount  int64   `json:"predicted_job_count"`
	PredictedUserCount int64   `json:"predicted_user_count"`
	PredictedCost      float64 `json:"predicted_cost"`
}

type PredictionAccuracy struct {
	MAE              float64 `json:"mae"`
	RMSE             float64 `json:"rmse"`
	MAPE             float64 `json:"mape"`
	R2Score          float64 `json:"r2_score"`
	AccuracyScore    float64 `json:"accuracy_score"`
	ModelPerformance string  `json:"model_performance"`
}

type UncertaintyAnalysis struct {
	ConfidenceIntervals *ConfidenceIntervals `json:"confidence_intervals"`
	PredictionVariance  float64              `json:"prediction_variance"`
	UncertaintyScore    float64              `json:"uncertainty_score"`
	RiskFactors         []string             `json:"risk_factors"`
}

type ConfidenceIntervals struct {
	LowerBound      float64 `json:"lower_bound"`
	UpperBound      float64 `json:"upper_bound"`
	IntervalWidth   float64 `json:"interval_width"`
	ConfidenceLevel float64 `json:"confidence_level"`
}

type ScenarioAnalysis struct {
	BestCaseScenario      *UsagePrediction   `json:"best_case_scenario"`
	WorstCaseScenario     *UsagePrediction   `json:"worst_case_scenario"`
	MostLikelyScenario    *UsagePrediction   `json:"most_likely_scenario"`
	ScenarioProbabilities map[string]float64 `json:"scenario_probabilities"`
}

type AccountClassification struct {
	AccountName         string                  `json:"account_name"`
	PrimaryClass        string                  `json:"primary_class"`
	SecondaryClasses    []string                `json:"secondary_classes"`
	ClassificationScore float64                 `json:"classification_score"`
	ClassificationModel string                  `json:"classification_model"`
	ClassFeatures       *ClassificationFeatures `json:"class_features"`
	Confidence          float64                 `json:"confidence"`
	SimilarAccounts     []string                `json:"similar_accounts"`
	ClassTrends         *ClassTrends            `json:"class_trends"`
}

type ClassificationFeatures struct {
	UsageIntensity          string `json:"usage_intensity"`
	ResourcePreference      string `json:"resource_preference"`
	SchedulingBehavior      string `json:"scheduling_behavior"`
	CollaborationLevel      string `json:"collaboration_level"`
	TechnicalSophistication string `json:"technical_sophistication"`
	ProjectComplexity       string `json:"project_complexity"`
	CostSensitivity         string `json:"cost_sensitivity"`
	InnovationLevel         string `json:"innovation_level"`
}

type ClassTrends struct {
	ClassStability     float64 `json:"class_stability"`
	TrendDirection     string  `json:"trend_direction"`
	EvolutionSpeed     float64 `json:"evolution_speed"`
	ClassMigrationRisk float64 `json:"class_migration_risk"`
}

type AccountComparison struct {
	AccountA              string                 `json:"account_a"`
	AccountB              string                 `json:"account_b"`
	SimilarityScore       float64                `json:"similarity_score"`
	DifferenceAnalysis    *DifferenceAnalysis    `json:"difference_analysis"`
	PerformanceComparison *PerformanceComparison `json:"performance_comparison"`
	ResourceComparison    *ResourceComparison    `json:"resource_comparison"`
	BehaviorComparison    *BehaviorComparison    `json:"behavior_comparison"`
	RecommendedActions    []string               `json:"recommended_actions"`
	BenchmarkAnalysis     *BenchmarkAnalysis     `json:"benchmark_analysis"`
}

type DifferenceAnalysis struct {
	KeyDifferences []string `json:"key_differences"`
	StrengthsA     []string `json:"strengths_a"`
	StrengthsB     []string `json:"strengths_b"`
	WeaknessesA    []string `json:"weaknesses_a"`
	WeaknessesB    []string `json:"weaknesses_b"`
	OpportunitiesA []string `json:"opportunities_a"`
	OpportunitiesB []string `json:"opportunities_b"`
}

type PerformanceComparison struct {
	EfficiencyA       float64 `json:"efficiency_a"`
	EfficiencyB       float64 `json:"efficiency_b"`
	ThroughputA       float64 `json:"throughput_a"`
	ThroughputB       float64 `json:"throughput_b"`
	CostEfficiencyA   float64 `json:"cost_efficiency_a"`
	CostEfficiencyB   float64 `json:"cost_efficiency_b"`
	PerformanceWinner string  `json:"performance_winner"`
}

type ResourceComparison struct {
	CPUUsageA             float64 `json:"cpu_usage_a"`
	CPUUsageB             float64 `json:"cpu_usage_b"`
	MemoryUsageA          float64 `json:"memory_usage_a"`
	MemoryUsageB          float64 `json:"memory_usage_b"`
	GPUUsageA             float64 `json:"gpu_usage_a"`
	GPUUsageB             float64 `json:"gpu_usage_b"`
	ResourceOptimizationA float64 `json:"resource_optimization_a"`
	ResourceOptimizationB float64 `json:"resource_optimization_b"`
}

type BehaviorComparison struct {
	SubmissionPatternA string  `json:"submission_pattern_a"`
	SubmissionPatternB string  `json:"submission_pattern_b"`
	CollaborationA     float64 `json:"collaboration_a"`
	CollaborationB     float64 `json:"collaboration_b"`
	AdaptabilityA      float64 `json:"adaptability_a"`
	AdaptabilityB      float64 `json:"adaptability_b"`
	InnovationA        float64 `json:"innovation_a"`
	InnovationB        float64 `json:"innovation_b"`
}

type BenchmarkAnalysis struct {
	RelativePerformanceA float64 `json:"relative_performance_a"`
	RelativePerformanceB float64 `json:"relative_performance_b"`
	IndustryBenchmark    float64 `json:"industry_benchmark"`
	PeerGroupAverage     float64 `json:"peer_group_average"`
	BestPracticeGap      float64 `json:"best_practice_gap"`
}

type AccountRiskAnalysis struct {
	AccountName          string                `json:"account_name"`
	OverallRiskScore     float64               `json:"overall_risk_score"`
	RiskLevel            string                `json:"risk_level"`
	RiskFactors          []*RiskFactor         `json:"risk_factors"`
	OperationalRisks     *OperationalRisks     `json:"operational_risks"`
	FinancialRisks       *FinancialRisks       `json:"financial_risks"`
	SecurityRisks        *SecurityRisks        `json:"security_risks"`
	ComplianceRisks      *ComplianceRisks      `json:"compliance_risks"`
	MitigationStrategies []*MitigationStrategy `json:"mitigation_strategies"`
	RiskTrends           *RiskTrends           `json:"risk_trends"`
}

type RiskFactor struct {
	FactorName  string  `json:"factor_name"`
	RiskScore   float64 `json:"risk_score"`
	Impact      string  `json:"impact"`
	Probability string  `json:"probability"`
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
}

type OperationalRisks struct {
	ResourceExhaustion     float64 `json:"resource_exhaustion"`
	QueueBottlenecks       float64 `json:"queue_bottlenecks"`
	PerformanceDegradation float64 `json:"performance_degradation"`
	ServiceDisruption      float64 `json:"service_disruption"`
	CapacityOverflow       float64 `json:"capacity_overflow"`
}

type FinancialRisks struct {
	BudgetOverrun        float64 `json:"budget_overrun"`
	CostSpikes           float64 `json:"cost_spikes"`
	UncontrolledSpending float64 `json:"uncontrolled_spending"`
	ResourceWaste        float64 `json:"resource_waste"`
	ROIDeterioration     float64 `json:"roi_deterioration"`
}

type SecurityRisks struct {
	UnauthorizedAccess  float64 `json:"unauthorized_access"`
	DataBreach          float64 `json:"data_breach"`
	PrivilegeEscalation float64 `json:"privilege_escalation"`
	MaliciousActivity   float64 `json:"malicious_activity"`
	CompromisedAccounts float64 `json:"compromised_accounts"`
}

type ComplianceRisks struct {
	PolicyViolations float64 `json:"policy_violations"`
	RegulatoryBreach float64 `json:"regulatory_breach"`
	AuditFailures    float64 `json:"audit_failures"`
	DataGovernance   float64 `json:"data_governance"`
	ComplianceDrift  float64 `json:"compliance_drift"`
}

type MitigationStrategy struct {
	StrategyName       string  `json:"strategy_name"`
	TargetRisk         string  `json:"target_risk"`
	ImplementationCost float64 `json:"implementation_cost"`
	Effectiveness      float64 `json:"effectiveness"`
	Priority           string  `json:"priority"`
	Timeline           string  `json:"timeline"`
	ResponsibleParty   string  `json:"responsible_party"`
}

type RiskTrends struct {
	RiskDirection    string   `json:"risk_direction"`
	RiskVelocity     float64  `json:"risk_velocity"`
	RiskVolatility   float64  `json:"risk_volatility"`
	EmergingRisks    []string `json:"emerging_risks"`
	DiminishingRisks []string `json:"diminishing_risks"`
}

type AccountOptimizationSuggestions struct {
	AccountName             string                   `json:"account_name"`
	OverallScore            float64                  `json:"overall_score"`
	ImprovementPotential    float64                  `json:"improvement_potential"`
	ResourceOptimization    *ResourceOptimization    `json:"resource_optimization"`
	CostOptimization        *CostOptimization        `json:"cost_optimization"`
	PerformanceOptimization *PerformanceOptimization `json:"performance_optimization"`
	WorkflowOptimization    *WorkflowOptimization    `json:"workflow_optimization"`
	UserOptimization        *UserOptimization        `json:"user_optimization"`
	SchedulingOptimization  *SchedulingOptimization  `json:"scheduling_optimization"`
	PriorityActions         []string                 `json:"priority_actions"`
	QuickWins               []string                 `json:"quick_wins"`
	LongTermGoals           []string                 `json:"long_term_goals"`
}

type ResourceOptimization struct {
	RightSizing      *RightSizing      `json:"right_sizing"`
	ResourcePooling  *ResourcePooling  `json:"resource_pooling"`
	LoadBalancing    *LoadBalancing    `json:"load_balancing"`
	CapacityPlanning *CapacityPlanning `json:"capacity_planning"`
}

type RightSizing struct {
	OversizedJobs    int64    `json:"oversized_jobs"`
	UndersizedJobs   int64    `json:"undersized_jobs"`
	OptimalSizing    []string `json:"optimal_sizing"`
	SavingsPotential float64  `json:"savings_potential"`
}

type ResourcePooling struct {
	PoolingOpportunities []string `json:"pooling_opportunities"`
	SharedResources      []string `json:"shared_resources"`
	PoolingBenefits      float64  `json:"pooling_benefits"`
}

type LoadBalancing struct {
	BalancingOpportunities []string `json:"balancing_opportunities"`
	BottleneckResources    []string `json:"bottleneck_resources"`
	BalancingStrategies    []string `json:"balancing_strategies"`
}

type CapacityPlanning struct {
	FutureNeeds        *UsagePrediction `json:"future_needs"`
	ExpansionPlan      []string         `json:"expansion_plan"`
	InvestmentRequired float64          `json:"investment_required"`
}

type PerformanceOptimization struct {
	PerformanceGaps     []string `json:"performance_gaps"`
	OptimizationTargets []string `json:"optimization_targets"`
	PerformanceActions  []string `json:"performance_actions"`
	ExpectedGains       float64  `json:"expected_gains"`
}

type WorkflowOptimization struct {
	WorkflowIssues          []string `json:"workflow_issues"`
	ProcessImprovements     []string `json:"process_improvements"`
	AutomationOpportunities []string `json:"automation_opportunities"`
	EfficiencyGains         float64  `json:"efficiency_gains"`
}

type UserOptimization struct {
	TrainingNeeds []string `json:"training_needs"`
	SkillGaps     []string `json:"skill_gaps"`
	BestPractices []string `json:"best_practices"`
	UserSupport   []string `json:"user_support"`
}

type SchedulingOptimization struct {
	SchedulingIssues    []string `json:"scheduling_issues"`
	OptimalTiming       []string `json:"optimal_timing"`
	PriorityAdjustments []string `json:"priority_adjustments"`
	QueueImprovements   []string `json:"queue_improvements"`
}

type AccountBenchmarks struct {
	AccountName         string               `json:"account_name"`
	BenchmarkSuite      string               `json:"benchmark_suite"`
	OverallRanking      int                  `json:"overall_ranking"`
	PeerGroupRanking    int                  `json:"peer_group_ranking"`
	IndustryRanking     int                  `json:"industry_ranking"`
	BenchmarkScores     *BenchmarkScores     `json:"benchmark_scores"`
	CompetitiveAnalysis *CompetitiveAnalysis `json:"competitive_analysis"`
	BenchmarkTrends     *BenchmarkTrends     `json:"benchmark_trends"`
	ImprovementAreas    []string             `json:"improvement_areas"`
	BestPractices       []string             `json:"best_practices"`
}

type BenchmarkScores struct {
	EfficiencyScore        float64 `json:"efficiency_score"`
	PerformanceScore       float64 `json:"performance_score"`
	CostEffectivenessScore float64 `json:"cost_effectiveness_score"`
	InnovationScore        float64 `json:"innovation_score"`
	SustainabilityScore    float64 `json:"sustainability_score"`
	UserSatisfactionScore  float64 `json:"user_satisfaction_score"`
}

type CompetitiveAnalysis struct {
	TopPerformers         []string `json:"top_performers"`
	CompetitiveGaps       []string `json:"competitive_gaps"`
	CompetitiveAdvantages []string `json:"competitive_advantages"`
	MarketPosition        string   `json:"market_position"`
}

type BenchmarkTrends struct {
	PerformanceTrend   string  `json:"performance_trend"`
	RankingTrend       string  `json:"ranking_trend"`
	ImprovementRate    float64 `json:"improvement_rate"`
	BenchmarkStability float64 `json:"benchmark_stability"`
}

type SystemUsageOverview struct {
	TotalAccounts     int64              `json:"total_accounts"`
	ActiveAccounts    int64              `json:"active_accounts"`
	InactiveAccounts  int64              `json:"inactive_accounts"`
	SystemUtilization *SystemUtilization `json:"system_utilization"`
	GlobalPatterns    *GlobalPatterns    `json:"global_patterns"`
	SystemHealth      *SystemHealth      `json:"system_health"`
	CapacityStatus    *CapacityStatus    `json:"capacity_status"`
	SystemTrends      *SystemTrends      `json:"system_trends"`
	UsageDistribution *UsageDistribution `json:"usage_distribution"`
	SystemAnomalies   []*SystemAnomaly   `json:"system_anomalies"`
	SystemAlerts      []*SystemAlert     `json:"system_alerts"`
}

type SystemUtilization struct {
	CPUUtilization     float64 `json:"cpu_utilization"`
	MemoryUtilization  float64 `json:"memory_utilization"`
	GPUUtilization     float64 `json:"gpu_utilization"`
	StorageUtilization float64 `json:"storage_utilization"`
	NetworkUtilization float64 `json:"network_utilization"`
	OverallUtilization float64 `json:"overall_utilization"`
}

type GlobalPatterns struct {
	PeakUsageHours       []int    `json:"peak_usage_hours"`
	OffPeakHours         []int    `json:"off_peak_hours"`
	SeasonalPatterns     []string `json:"seasonal_patterns"`
	RegionalPatterns     []string `json:"regional_patterns"`
	DepartmentalPatterns []string `json:"departmental_patterns"`
}

type SystemHealth struct {
	HealthScore       float64 `json:"health_score"`
	PerformanceScore  float64 `json:"performance_score"`
	StabilityScore    float64 `json:"stability_score"`
	AvailabilityScore float64 `json:"availability_score"`
	ReliabilityScore  float64 `json:"reliability_score"`
}

type CapacityStatus struct {
	TotalCapacity       *ResourceCapacity `json:"total_capacity"`
	UsedCapacity        *ResourceCapacity `json:"used_capacity"`
	AvailableCapacity   *ResourceCapacity `json:"available_capacity"`
	ReservedCapacity    *ResourceCapacity `json:"reserved_capacity"`
	CapacityUtilization float64           `json:"capacity_utilization"`
}

type ResourceCapacity struct {
	CPUCores  int64   `json:"cpu_cores"`
	MemoryGB  float64 `json:"memory_gb"`
	GPUCount  int64   `json:"gpu_count"`
	StorageTB float64 `json:"storage_tb"`
	NodeCount int64   `json:"node_count"`
}

type SystemTrends struct {
	UsageTrend       string  `json:"usage_trend"`
	GrowthRate       float64 `json:"growth_rate"`
	CapacityTrend    string  `json:"capacity_trend"`
	PerformanceTrend string  `json:"performance_trend"`
	CostTrend        string  `json:"cost_trend"`
}

type UsageDistribution struct {
	TopUsers           []string `json:"top_users"`
	TopAccounts        []string `json:"top_accounts"`
	ResourceHogs       []string `json:"resource_hogs"`
	LightUsers         []string `json:"light_users"`
	UsageConcentration float64  `json:"usage_concentration"`
}

type SystemAnomaly struct {
	AnomalyType        string    `json:"anomaly_type"`
	Severity           string    `json:"severity"`
	DetectedAt         time.Time `json:"detected_at"`
	AffectedComponents []string  `json:"affected_components"`
	Impact             string    `json:"impact"`
	RootCause          string    `json:"root_cause"`
	IsResolved         bool      `json:"is_resolved"`
}

type SystemAlert struct {
	AlertType        string    `json:"alert_type"`
	Priority         string    `json:"priority"`
	Message          string    `json:"message"`
	TriggeredAt      time.Time `json:"triggered_at"`
	AffectedEntities []string  `json:"affected_entities"`
	ActionRequired   string    `json:"action_required"`
	IsAcknowledged   bool      `json:"is_acknowledged"`
}

type AccountUsagePatternsCollector struct {
	client AccountUsagePatternsSLURMClient
	mutex  sync.RWMutex

	// Usage Pattern Metrics
	accountUsageScore        *prometheus.GaugeVec
	accountCPUUtilization    *prometheus.GaugeVec
	accountMemoryUtilization *prometheus.GaugeVec
	accountGPUUtilization    *prometheus.GaugeVec
	accountJobSubmissions    *prometheus.GaugeVec
	accountActiveUsers       *prometheus.GaugeVec
	accountCostAnalysis      *prometheus.GaugeVec
	accountEfficiencyScore   *prometheus.GaugeVec

	// Anomaly Detection Metrics
	accountAnomaliesTotal        *prometheus.GaugeVec
	accountAnomalySeverity       *prometheus.GaugeVec
	accountAnomalyDeviationScore *prometheus.GaugeVec
	accountAnomalyResolutionTime *prometheus.GaugeVec

	// Behavior Profile Metrics
	accountBehaviorType        *prometheus.GaugeVec
	accountActivityLevel       *prometheus.GaugeVec
	accountAdaptabilityScore   *prometheus.GaugeVec
	accountConsistencyScore    *prometheus.GaugeVec
	accountInnovationScore     *prometheus.GaugeVec
	accountPredictabilityScore *prometheus.GaugeVec

	// Trend Analysis Metrics
	accountGrowthRate           *prometheus.GaugeVec
	accountTrendConfidence      *prometheus.GaugeVec
	accountForecastAccuracy     *prometheus.GaugeVec
	accountSeasonalityStrength  *prometheus.GaugeVec
	accountOutlierRate          *prometheus.GaugeVec
	accountChangePointFrequency *prometheus.GaugeVec

	// Seasonality Metrics
	accountDailySeasonality   *prometheus.GaugeVec
	accountWeeklySeasonality  *prometheus.GaugeVec
	accountMonthlySeasonality *prometheus.GaugeVec
	accountYearlySeasonality  *prometheus.GaugeVec
	accountSeasonalityScore   *prometheus.GaugeVec

	// Prediction Metrics
	accountPredictionAccuracy   *prometheus.GaugeVec
	accountPredictionConfidence *prometheus.GaugeVec
	accountUncertaintyScore     *prometheus.GaugeVec
	accountPredictionVariance   *prometheus.GaugeVec

	// Classification Metrics
	accountClassificationScore *prometheus.GaugeVec
	accountClassConfidence     *prometheus.GaugeVec
	accountClassStability      *prometheus.GaugeVec
	accountClassMigrationRisk  *prometheus.GaugeVec

	// Comparison Metrics
	accountSimilarityScore  *prometheus.GaugeVec
	accountPerformanceRatio *prometheus.GaugeVec
	accountResourceRatio    *prometheus.GaugeVec
	accountBenchmarkScore   *prometheus.GaugeVec

	// Risk Analysis Metrics
	accountRiskScore       *prometheus.GaugeVec
	accountOperationalRisk *prometheus.GaugeVec
	accountFinancialRisk   *prometheus.GaugeVec
	accountSecurityRisk    *prometheus.GaugeVec
	accountComplianceRisk  *prometheus.GaugeVec
	accountRiskVelocity    *prometheus.GaugeVec

	// Optimization Metrics
	accountOptimizationScore        *prometheus.GaugeVec
	accountImprovementPotential     *prometheus.GaugeVec
	accountRightSizingSavings       *prometheus.GaugeVec
	accountResourcePoolingBenefits  *prometheus.GaugeVec
	accountPerformanceGainPotential *prometheus.GaugeVec

	// Benchmark Metrics
	accountOverallRanking      *prometheus.GaugeVec
	accountPeerGroupRanking    *prometheus.GaugeVec
	accountIndustryRanking     *prometheus.GaugeVec
	accountBenchmarkTrend      *prometheus.GaugeVec
	accountCompetitivePosition *prometheus.GaugeVec

	// System Overview Metrics
	systemTotalAccounts       prometheus.Gauge
	systemActiveAccounts      prometheus.Gauge
	systemUtilization         *prometheus.GaugeVec
	systemHealthScore         prometheus.Gauge
	systemCapacityUtilization prometheus.Gauge
	systemAnomaliesTotal      prometheus.Gauge
	systemAlertsTotal         prometheus.Gauge
}

func NewAccountUsagePatternsCollector(client AccountUsagePatternsSLURMClient) *AccountUsagePatternsCollector {
	return &AccountUsagePatternsCollector{
		client: client,

		// Usage Pattern Metrics
		accountUsageScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_usage_score",
				Help: "Overall usage score for account indicating utilization level",
			},
			[]string{"account", "time_period"},
		),
		accountCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cpu_utilization",
				Help: "CPU utilization percentage for account",
			},
			[]string{"account", "resource_type"},
		),
		accountMemoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_memory_utilization",
				Help: "Memory utilization percentage for account",
			},
			[]string{"account", "resource_type"},
		),
		accountGPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_gpu_utilization",
				Help: "GPU utilization percentage for account",
			},
			[]string{"account", "resource_type"},
		),
		accountJobSubmissions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_job_submissions",
				Help: "Number of job submissions for account",
			},
			[]string{"account", "time_period", "submission_type"},
		),
		accountActiveUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_active_users",
				Help: "Number of active users in account",
			},
			[]string{"account", "user_type"},
		),
		accountCostAnalysis: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_cost_analysis",
				Help: "Cost analysis metrics for account",
			},
			[]string{"account", "cost_type", "currency"},
		),
		accountEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_efficiency_score",
				Help: "Efficiency score for account resource utilization",
			},
			[]string{"account", "efficiency_type"},
		),

		// Anomaly Detection Metrics
		accountAnomaliesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_anomalies_total",
				Help: "Total number of anomalies detected for account",
			},
			[]string{"account", "anomaly_type", "severity"},
		),
		accountAnomalySeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_anomaly_severity",
				Help: "Severity score of account anomalies",
			},
			[]string{"account", "anomaly_type"},
		),
		accountAnomalyDeviationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_anomaly_deviation_score",
				Help: "Deviation score for detected anomalies",
			},
			[]string{"account", "anomaly_type"},
		),
		accountAnomalyResolutionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_anomaly_resolution_time_seconds",
				Help: "Time taken to resolve anomalies in seconds",
			},
			[]string{"account", "anomaly_type"},
		),

		// Behavior Profile Metrics
		accountBehaviorType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_behavior_type",
				Help: "Behavior type classification for account (encoded as numeric)",
			},
			[]string{"account", "behavior_type"},
		),
		accountActivityLevel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_activity_level",
				Help: "Activity level classification for account (encoded as numeric)",
			},
			[]string{"account", "activity_level"},
		),
		accountAdaptabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_adaptability_score",
				Help: "Adaptability score for account behavior",
			},
			[]string{"account"},
		),
		accountConsistencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_consistency_score",
				Help: "Consistency score for account behavior patterns",
			},
			[]string{"account"},
		),
		accountInnovationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_innovation_score",
				Help: "Innovation score for account usage patterns",
			},
			[]string{"account"},
		),
		accountPredictabilityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_predictability_score",
				Help: "Predictability score for account behavior",
			},
			[]string{"account"},
		),

		// Trend Analysis Metrics
		accountGrowthRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_growth_rate",
				Help: "Growth rate for account usage metrics",
			},
			[]string{"account", "metric_type"},
		),
		accountTrendConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_trend_confidence",
				Help: "Confidence level in trend analysis",
			},
			[]string{"account", "trend_type"},
		),
		accountForecastAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_forecast_accuracy",
				Help: "Accuracy of forecasting models for account",
			},
			[]string{"account", "forecast_type"},
		),
		accountSeasonalityStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_seasonality_strength",
				Help: "Strength of seasonal patterns in account usage",
			},
			[]string{"account", "season_type"},
		),
		accountOutlierRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_outlier_rate",
				Help: "Rate of outlier detection in account usage",
			},
			[]string{"account", "outlier_type"},
		),
		accountChangePointFrequency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_change_point_frequency",
				Help: "Frequency of change points in account behavior",
			},
			[]string{"account"},
		),

		// Seasonality Metrics
		accountDailySeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_daily_seasonality",
				Help: "Daily seasonality patterns for account usage",
			},
			[]string{"account", "hour"},
		),
		accountWeeklySeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_weekly_seasonality",
				Help: "Weekly seasonality patterns for account usage",
			},
			[]string{"account", "day_of_week"},
		),
		accountMonthlySeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_monthly_seasonality",
				Help: "Monthly seasonality patterns for account usage",
			},
			[]string{"account", "month"},
		),
		accountYearlySeasonality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_yearly_seasonality",
				Help: "Yearly seasonality patterns for account usage",
			},
			[]string{"account", "quarter"},
		),
		accountSeasonalityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_seasonality_score",
				Help: "Overall seasonality score for account",
			},
			[]string{"account", "seasonality_type"},
		),

		// Prediction Metrics
		accountPredictionAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_prediction_accuracy",
				Help: "Accuracy of predictions for account usage",
			},
			[]string{"account", "prediction_type", "metric"},
		),
		accountPredictionConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_prediction_confidence",
				Help: "Confidence level in predictions for account",
			},
			[]string{"account", "prediction_horizon"},
		),
		accountUncertaintyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_uncertainty_score",
				Help: "Uncertainty score for account predictions",
			},
			[]string{"account", "prediction_type"},
		),
		accountPredictionVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_prediction_variance",
				Help: "Variance in account usage predictions",
			},
			[]string{"account", "prediction_type"},
		),

		// Classification Metrics
		accountClassificationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_classification_score",
				Help: "Classification score for account categorization",
			},
			[]string{"account", "classification_model"},
		),
		accountClassConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_class_confidence",
				Help: "Confidence in account classification",
			},
			[]string{"account", "primary_class"},
		),
		accountClassStability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_class_stability",
				Help: "Stability of account classification over time",
			},
			[]string{"account", "class_type"},
		),
		accountClassMigrationRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_class_migration_risk",
				Help: "Risk of account migrating to different class",
			},
			[]string{"account", "current_class", "target_class"},
		),

		// Comparison Metrics
		accountSimilarityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_similarity_score",
				Help: "Similarity score between accounts",
			},
			[]string{"account_a", "account_b", "comparison_type"},
		),
		accountPerformanceRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_ratio",
				Help: "Performance ratio between accounts",
			},
			[]string{"account_a", "account_b", "metric_type"},
		),
		accountResourceRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_ratio",
				Help: "Resource usage ratio between accounts",
			},
			[]string{"account_a", "account_b", "resource_type"},
		),
		accountBenchmarkScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_score",
				Help: "Benchmark score for account performance",
			},
			[]string{"account", "benchmark_type"},
		),

		// Risk Analysis Metrics
		accountRiskScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_risk_score",
				Help: "Overall risk score for account",
			},
			[]string{"account", "risk_type"},
		),
		accountOperationalRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_operational_risk",
				Help: "Operational risk factors for account",
			},
			[]string{"account", "risk_factor"},
		),
		accountFinancialRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_financial_risk",
				Help: "Financial risk factors for account",
			},
			[]string{"account", "risk_factor"},
		),
		accountSecurityRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_security_risk",
				Help: "Security risk factors for account",
			},
			[]string{"account", "risk_factor"},
		),
		accountComplianceRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_compliance_risk",
				Help: "Compliance risk factors for account",
			},
			[]string{"account", "risk_factor"},
		),
		accountRiskVelocity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_risk_velocity",
				Help: "Rate of change in account risk levels",
			},
			[]string{"account", "risk_type"},
		),

		// Optimization Metrics
		accountOptimizationScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_optimization_score",
				Help: "Overall optimization score for account",
			},
			[]string{"account", "optimization_type"},
		),
		accountImprovementPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_improvement_potential",
				Help: "Potential for improvement in account performance",
			},
			[]string{"account", "improvement_area"},
		),
		accountRightSizingSavings: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_right_sizing_savings",
				Help: "Potential savings from right-sizing resources",
			},
			[]string{"account", "resource_type"},
		),
		accountResourcePoolingBenefits: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_resource_pooling_benefits",
				Help: "Benefits from resource pooling strategies",
			},
			[]string{"account", "pooling_type"},
		),
		accountPerformanceGainPotential: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_performance_gain_potential",
				Help: "Potential performance gains from optimization",
			},
			[]string{"account", "optimization_area"},
		),

		// Benchmark Metrics
		accountOverallRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_overall_ranking",
				Help: "Overall ranking of account compared to peers",
			},
			[]string{"account", "ranking_type"},
		),
		accountPeerGroupRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_peer_group_ranking",
				Help: "Ranking within peer group",
			},
			[]string{"account", "peer_group"},
		),
		accountIndustryRanking: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_industry_ranking",
				Help: "Industry-wide ranking of account",
			},
			[]string{"account", "industry_category"},
		),
		accountBenchmarkTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_benchmark_trend",
				Help: "Trend in benchmark performance over time",
			},
			[]string{"account", "benchmark_metric"},
		),
		accountCompetitivePosition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_competitive_position",
				Help: "Competitive position score for account",
			},
			[]string{"account", "competitive_dimension"},
		),

		// System Overview Metrics
		systemTotalAccounts: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_total_accounts",
				Help: "Total number of accounts in the system",
			},
		),
		systemActiveAccounts: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_active_accounts",
				Help: "Number of active accounts in the system",
			},
		),
		systemUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_utilization",
				Help: "System-wide resource utilization",
			},
			[]string{"resource_type"},
		),
		systemHealthScore: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_health_score",
				Help: "Overall system health score",
			},
		),
		systemCapacityUtilization: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_capacity_utilization",
				Help: "Overall system capacity utilization",
			},
		),
		systemAnomaliesTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_anomalies_total",
				Help: "Total number of system-wide anomalies detected",
			},
		),
		systemAlertsTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_system_alerts_total",
				Help: "Total number of active system alerts",
			},
		),
	}
}

func (c *AccountUsagePatternsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.accountUsageScore.Describe(ch)
	c.accountCPUUtilization.Describe(ch)
	c.accountMemoryUtilization.Describe(ch)
	c.accountGPUUtilization.Describe(ch)
	c.accountJobSubmissions.Describe(ch)
	c.accountActiveUsers.Describe(ch)
	c.accountCostAnalysis.Describe(ch)
	c.accountEfficiencyScore.Describe(ch)
	c.accountAnomaliesTotal.Describe(ch)
	c.accountAnomalySeverity.Describe(ch)
	c.accountAnomalyDeviationScore.Describe(ch)
	c.accountAnomalyResolutionTime.Describe(ch)
	c.accountBehaviorType.Describe(ch)
	c.accountActivityLevel.Describe(ch)
	c.accountAdaptabilityScore.Describe(ch)
	c.accountConsistencyScore.Describe(ch)
	c.accountInnovationScore.Describe(ch)
	c.accountPredictabilityScore.Describe(ch)
	c.accountGrowthRate.Describe(ch)
	c.accountTrendConfidence.Describe(ch)
	c.accountForecastAccuracy.Describe(ch)
	c.accountSeasonalityStrength.Describe(ch)
	c.accountOutlierRate.Describe(ch)
	c.accountChangePointFrequency.Describe(ch)
	c.accountDailySeasonality.Describe(ch)
	c.accountWeeklySeasonality.Describe(ch)
	c.accountMonthlySeasonality.Describe(ch)
	c.accountYearlySeasonality.Describe(ch)
	c.accountSeasonalityScore.Describe(ch)
	c.accountPredictionAccuracy.Describe(ch)
	c.accountPredictionConfidence.Describe(ch)
	c.accountUncertaintyScore.Describe(ch)
	c.accountPredictionVariance.Describe(ch)
	c.accountClassificationScore.Describe(ch)
	c.accountClassConfidence.Describe(ch)
	c.accountClassStability.Describe(ch)
	c.accountClassMigrationRisk.Describe(ch)
	c.accountSimilarityScore.Describe(ch)
	c.accountPerformanceRatio.Describe(ch)
	c.accountResourceRatio.Describe(ch)
	c.accountBenchmarkScore.Describe(ch)
	c.accountRiskScore.Describe(ch)
	c.accountOperationalRisk.Describe(ch)
	c.accountFinancialRisk.Describe(ch)
	c.accountSecurityRisk.Describe(ch)
	c.accountComplianceRisk.Describe(ch)
	c.accountRiskVelocity.Describe(ch)
	c.accountOptimizationScore.Describe(ch)
	c.accountImprovementPotential.Describe(ch)
	c.accountRightSizingSavings.Describe(ch)
	c.accountResourcePoolingBenefits.Describe(ch)
	c.accountPerformanceGainPotential.Describe(ch)
	c.accountOverallRanking.Describe(ch)
	c.accountPeerGroupRanking.Describe(ch)
	c.accountIndustryRanking.Describe(ch)
	c.accountBenchmarkTrend.Describe(ch)
	c.accountCompetitivePosition.Describe(ch)
	c.systemTotalAccounts.Describe(ch)
	c.systemActiveAccounts.Describe(ch)
	c.systemUtilization.Describe(ch)
	c.systemHealthScore.Describe(ch)
	c.systemCapacityUtilization.Describe(ch)
	c.systemAnomaliesTotal.Describe(ch)
	c.systemAlertsTotal.Describe(ch)
}

func (c *AccountUsagePatternsCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()

	accounts, err := c.client.ListAccounts(ctx)
	if err != nil {
		return
	}

	for _, account := range accounts {
		c.collectAccountMetrics(ch, account.Name)
	}

	c.collectSystemOverviewMetrics(ch)
}

func (c *AccountUsagePatternsCollector) collectAccountMetrics(ch chan<- prometheus.Metric, accountName string) {
	ctx := context.Background()

	// Collect usage patterns
	if patterns, err := c.client.GetAccountUsagePatterns(ctx, accountName); err == nil {
		c.collectUsagePatternsMetrics(ch, accountName, patterns)
	}

	// Collect anomalies
	if anomalies, err := c.client.GetAccountAnomalies(ctx, accountName); err == nil {
		c.collectAnomalyMetrics(ch, accountName, anomalies)
	}

	// Collect behavior profile
	if profile, err := c.client.GetAccountBehaviorProfile(ctx, accountName); err == nil {
		c.collectBehaviorProfileMetrics(ch, accountName, profile)
	}

	// Collect trend analysis
	if trends, err := c.client.GetAccountTrendAnalysis(ctx, accountName); err == nil {
		c.collectTrendAnalysisMetrics(ch, accountName, trends)
	}

	// Collect seasonality
	if seasonality, err := c.client.GetAccountSeasonality(ctx, accountName); err == nil {
		c.collectSeasonalityMetrics(ch, accountName, seasonality)
	}

	// Collect predictions
	if predictions, err := c.client.GetAccountPredictions(ctx, accountName); err == nil {
		c.collectPredictionMetrics(ch, accountName, predictions)
	}

	// Collect classification
	if classification, err := c.client.GetAccountClassification(ctx, accountName); err == nil {
		c.collectClassificationMetrics(ch, accountName, classification)
	}

	// Collect risk analysis
	if risk, err := c.client.GetAccountRiskAnalysis(ctx, accountName); err == nil {
		c.collectRiskAnalysisMetrics(ch, accountName, risk)
	}

	// Collect optimization suggestions
	if optimization, err := c.client.GetAccountOptimizationSuggestions(ctx, accountName); err == nil {
		c.collectOptimizationMetrics(ch, accountName, optimization)
	}

	// Collect benchmarks
	if benchmarks, err := c.client.GetAccountBenchmarks(ctx, accountName); err == nil {
		c.collectBenchmarkMetrics(ch, accountName, benchmarks)
	}
}

func (c *AccountUsagePatternsCollector) collectUsagePatternsMetrics(ch chan<- prometheus.Metric, accountName string, patterns *AccountUsagePatterns) {
	// Cost analysis metrics
	if patterns.CostAnalysis != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountCostAnalysis.WithLabelValues(accountName, "total", "USD").Desc(),
			prometheus.GaugeValue,
			patterns.CostAnalysis.TotalCost,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountCostAnalysis.WithLabelValues(accountName, "cpu_hour", "USD").Desc(),
			prometheus.GaugeValue,
			patterns.CostAnalysis.CostPerCPUHour,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountCostAnalysis.WithLabelValues(accountName, "gpu_hour", "USD").Desc(),
			prometheus.GaugeValue,
			patterns.CostAnalysis.CostPerGPUHour,
		)
	}

	// Efficiency metrics
	if patterns.EfficiencyMetrics != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountEfficiencyScore.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			patterns.EfficiencyMetrics.OverallEfficiency,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountEfficiencyScore.WithLabelValues(accountName, "cpu").Desc(),
			prometheus.GaugeValue,
			patterns.EfficiencyMetrics.CPUEfficiency,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountEfficiencyScore.WithLabelValues(accountName, "memory").Desc(),
			prometheus.GaugeValue,
			patterns.EfficiencyMetrics.MemoryEfficiency,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountEfficiencyScore.WithLabelValues(accountName, "gpu").Desc(),
			prometheus.GaugeValue,
			patterns.EfficiencyMetrics.GPUEfficiency,
		)
	}

	// User activity metrics
	if patterns.UserActivityProfile != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountActiveUsers.WithLabelValues(accountName, "active").Desc(),
			prometheus.GaugeValue,
			float64(patterns.UserActivityProfile.ActiveUsers),
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountActiveUsers.WithLabelValues(accountName, "power").Desc(),
			prometheus.GaugeValue,
			float64(patterns.UserActivityProfile.PowerUsers),
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountActiveUsers.WithLabelValues(accountName, "occasional").Desc(),
			prometheus.GaugeValue,
			float64(patterns.UserActivityProfile.OccasionalUsers),
		)
	}
}

func (c *AccountUsagePatternsCollector) collectAnomalyMetrics(ch chan<- prometheus.Metric, accountName string, anomalies []*AccountAnomaly) {
	// Count anomalies by type and severity
	anomalyCount := make(map[string]map[string]int)
	totalDeviation := 0.0

	for _, anomaly := range anomalies {
		if anomalyCount[anomaly.AnomalyType] == nil {
			anomalyCount[anomaly.AnomalyType] = make(map[string]int)
		}
		anomalyCount[anomaly.AnomalyType][anomaly.Severity]++
		totalDeviation += anomaly.DeviationScore

		// Individual anomaly metrics
		ch <- prometheus.MustNewConstMetric(
			c.accountAnomalyDeviationScore.WithLabelValues(accountName, anomaly.AnomalyType).Desc(),
			prometheus.GaugeValue,
			anomaly.DeviationScore,
		)

		if anomaly.IsResolved && anomaly.ResolvedAt != nil {
			resolutionTime := anomaly.ResolvedAt.Sub(anomaly.DetectedAt).Seconds()
			ch <- prometheus.MustNewConstMetric(
				c.accountAnomalyResolutionTime.WithLabelValues(accountName, anomaly.AnomalyType).Desc(),
				prometheus.GaugeValue,
				resolutionTime,
			)
		}
	}

	// Aggregate anomaly metrics
	for anomalyType, severities := range anomalyCount {
		for severity, count := range severities {
			ch <- prometheus.MustNewConstMetric(
				c.accountAnomaliesTotal.WithLabelValues(accountName, anomalyType, severity).Desc(),
				prometheus.GaugeValue,
				float64(count),
			)
		}
	}
}

func (c *AccountUsagePatternsCollector) collectBehaviorProfileMetrics(ch chan<- prometheus.Metric, accountName string, profile *AccountBehaviorProfile) {
	ch <- prometheus.MustNewConstMetric(
		c.accountAdaptabilityScore.WithLabelValues(accountName).Desc(),
		prometheus.GaugeValue,
		profile.AdaptabilityScore,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountConsistencyScore.WithLabelValues(accountName).Desc(),
		prometheus.GaugeValue,
		profile.ConsistencyScore,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountInnovationScore.WithLabelValues(accountName).Desc(),
		prometheus.GaugeValue,
		profile.InnovationScore,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountPredictabilityScore.WithLabelValues(accountName).Desc(),
		prometheus.GaugeValue,
		profile.PredictabilityScore,
	)
}

func (c *AccountUsagePatternsCollector) collectTrendAnalysisMetrics(ch chan<- prometheus.Metric, accountName string, trends *AccountTrendAnalysis) {
	ch <- prometheus.MustNewConstMetric(
		c.accountGrowthRate.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		trends.GrowthRate,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountTrendConfidence.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		trends.TrendConfidence,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountForecastAccuracy.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		trends.ForecastAccuracy,
	)

	if trends.OutlierDetection != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountOutlierRate.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			trends.OutlierDetection.OutlierRate,
		)
	}

	if trends.ChangePointAnalysis != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountChangePointFrequency.WithLabelValues(accountName).Desc(),
			prometheus.GaugeValue,
			trends.ChangePointAnalysis.ChangePointFrequency,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectSeasonalityMetrics(ch chan<- prometheus.Metric, accountName string, seasonality *AccountSeasonality) {
	ch <- prometheus.MustNewConstMetric(
		c.accountSeasonalityScore.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		seasonality.SeasonalityScore,
	)

	if seasonality.DailySeasonality != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountSeasonalityScore.WithLabelValues(accountName, "daily").Desc(),
			prometheus.GaugeValue,
			seasonality.DailySeasonality.SeasonalityStrength,
		)
	}

	if seasonality.WeeklySeasonality != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountSeasonalityScore.WithLabelValues(accountName, "weekly").Desc(),
			prometheus.GaugeValue,
			seasonality.WeeklySeasonality.SeasonalityStrength,
		)
	}

	if seasonality.MonthlySeasonality != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountSeasonalityScore.WithLabelValues(accountName, "monthly").Desc(),
			prometheus.GaugeValue,
			seasonality.MonthlySeasonality.SeasonalityStrength,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectPredictionMetrics(ch chan<- prometheus.Metric, accountName string, predictions *AccountPredictions) {
	if predictions.PredictionAccuracy != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountPredictionAccuracy.WithLabelValues(accountName, "overall", "mae").Desc(),
			prometheus.GaugeValue,
			predictions.PredictionAccuracy.MAE,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountPredictionAccuracy.WithLabelValues(accountName, "overall", "rmse").Desc(),
			prometheus.GaugeValue,
			predictions.PredictionAccuracy.RMSE,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountPredictionAccuracy.WithLabelValues(accountName, "overall", "r2").Desc(),
			prometheus.GaugeValue,
			predictions.PredictionAccuracy.R2Score,
		)
	}

	if predictions.ShortTermPredictions != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountPredictionConfidence.WithLabelValues(accountName, "short_term").Desc(),
			prometheus.GaugeValue,
			predictions.ShortTermPredictions.ConfidenceLevel,
		)
	}

	if predictions.UncertaintyAnalysis != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountUncertaintyScore.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			predictions.UncertaintyAnalysis.UncertaintyScore,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountPredictionVariance.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			predictions.UncertaintyAnalysis.PredictionVariance,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectClassificationMetrics(ch chan<- prometheus.Metric, accountName string, classification *AccountClassification) {
	ch <- prometheus.MustNewConstMetric(
		c.accountClassificationScore.WithLabelValues(accountName, classification.ClassificationModel).Desc(),
		prometheus.GaugeValue,
		classification.ClassificationScore,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountClassConfidence.WithLabelValues(accountName, classification.PrimaryClass).Desc(),
		prometheus.GaugeValue,
		classification.Confidence,
	)

	if classification.ClassTrends != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountClassStability.WithLabelValues(accountName, classification.PrimaryClass).Desc(),
			prometheus.GaugeValue,
			classification.ClassTrends.ClassStability,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountClassMigrationRisk.WithLabelValues(accountName, classification.PrimaryClass, "unknown").Desc(),
			prometheus.GaugeValue,
			classification.ClassTrends.ClassMigrationRisk,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectRiskAnalysisMetrics(ch chan<- prometheus.Metric, accountName string, risk *AccountRiskAnalysis) {
	ch <- prometheus.MustNewConstMetric(
		c.accountRiskScore.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		risk.OverallRiskScore,
	)

	if risk.OperationalRisks != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountOperationalRisk.WithLabelValues(accountName, "resource_exhaustion").Desc(),
			prometheus.GaugeValue,
			risk.OperationalRisks.ResourceExhaustion,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountOperationalRisk.WithLabelValues(accountName, "queue_bottlenecks").Desc(),
			prometheus.GaugeValue,
			risk.OperationalRisks.QueueBottlenecks,
		)
	}

	if risk.FinancialRisks != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountFinancialRisk.WithLabelValues(accountName, "budget_overrun").Desc(),
			prometheus.GaugeValue,
			risk.FinancialRisks.BudgetOverrun,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountFinancialRisk.WithLabelValues(accountName, "cost_spikes").Desc(),
			prometheus.GaugeValue,
			risk.FinancialRisks.CostSpikes,
		)
	}

	if risk.SecurityRisks != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountSecurityRisk.WithLabelValues(accountName, "unauthorized_access").Desc(),
			prometheus.GaugeValue,
			risk.SecurityRisks.UnauthorizedAccess,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountSecurityRisk.WithLabelValues(accountName, "data_breach").Desc(),
			prometheus.GaugeValue,
			risk.SecurityRisks.DataBreach,
		)
	}

	if risk.ComplianceRisks != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountComplianceRisk.WithLabelValues(accountName, "policy_violations").Desc(),
			prometheus.GaugeValue,
			risk.ComplianceRisks.PolicyViolations,
		)
	}

	if risk.RiskTrends != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountRiskVelocity.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			risk.RiskTrends.RiskVelocity,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectOptimizationMetrics(ch chan<- prometheus.Metric, accountName string, optimization *AccountOptimizationSuggestions) {
	ch <- prometheus.MustNewConstMetric(
		c.accountOptimizationScore.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		optimization.OverallScore,
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountImprovementPotential.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		optimization.ImprovementPotential,
	)

	if optimization.ResourceOptimization != nil && optimization.ResourceOptimization.RightSizing != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountRightSizingSavings.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			optimization.ResourceOptimization.RightSizing.SavingsPotential,
		)
	}

	if optimization.ResourceOptimization != nil && optimization.ResourceOptimization.ResourcePooling != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountResourcePoolingBenefits.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			optimization.ResourceOptimization.ResourcePooling.PoolingBenefits,
		)
	}

	if optimization.PerformanceOptimization != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountPerformanceGainPotential.WithLabelValues(accountName, "overall").Desc(),
			prometheus.GaugeValue,
			optimization.PerformanceOptimization.ExpectedGains,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectBenchmarkMetrics(ch chan<- prometheus.Metric, accountName string, benchmarks *AccountBenchmarks) {
	ch <- prometheus.MustNewConstMetric(
		c.accountOverallRanking.WithLabelValues(accountName, "overall").Desc(),
		prometheus.GaugeValue,
		float64(benchmarks.OverallRanking),
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountPeerGroupRanking.WithLabelValues(accountName, "peer_group").Desc(),
		prometheus.GaugeValue,
		float64(benchmarks.PeerGroupRanking),
	)
	ch <- prometheus.MustNewConstMetric(
		c.accountIndustryRanking.WithLabelValues(accountName, "industry").Desc(),
		prometheus.GaugeValue,
		float64(benchmarks.IndustryRanking),
	)

	if benchmarks.BenchmarkScores != nil {
		ch <- prometheus.MustNewConstMetric(
			c.accountBenchmarkScore.WithLabelValues(accountName, "efficiency").Desc(),
			prometheus.GaugeValue,
			benchmarks.BenchmarkScores.EfficiencyScore,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountBenchmarkScore.WithLabelValues(accountName, "performance").Desc(),
			prometheus.GaugeValue,
			benchmarks.BenchmarkScores.PerformanceScore,
		)
		ch <- prometheus.MustNewConstMetric(
			c.accountBenchmarkScore.WithLabelValues(accountName, "cost_effectiveness").Desc(),
			prometheus.GaugeValue,
			benchmarks.BenchmarkScores.CostEffectivenessScore,
		)
	}
}

func (c *AccountUsagePatternsCollector) collectSystemOverviewMetrics(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	overview, err := c.client.GetSystemUsageOverview(ctx)
	if err != nil {
		return
	}

	c.systemTotalAccounts.Set(float64(overview.TotalAccounts))
	c.systemActiveAccounts.Set(float64(overview.ActiveAccounts))

	if overview.SystemUtilization != nil {
		ch <- prometheus.MustNewConstMetric(
			c.systemUtilization.WithLabelValues("cpu").Desc(),
			prometheus.GaugeValue,
			overview.SystemUtilization.CPUUtilization,
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemUtilization.WithLabelValues("memory").Desc(),
			prometheus.GaugeValue,
			overview.SystemUtilization.MemoryUtilization,
		)
		ch <- prometheus.MustNewConstMetric(
			c.systemUtilization.WithLabelValues("gpu").Desc(),
			prometheus.GaugeValue,
			overview.SystemUtilization.GPUUtilization,
		)
	}

	if overview.SystemHealth != nil {
		c.systemHealthScore.Set(overview.SystemHealth.HealthScore)
	}

	if overview.CapacityStatus != nil {
		c.systemCapacityUtilization.Set(overview.CapacityStatus.CapacityUtilization)
	}

	c.systemAnomaliesTotal.Set(float64(len(overview.SystemAnomalies)))
	c.systemAlertsTotal.Set(float64(len(overview.SystemAlerts)))

	ch <- prometheus.MustNewConstMetric(c.systemTotalAccounts.Desc(), prometheus.GaugeValue, float64(overview.TotalAccounts))
	ch <- prometheus.MustNewConstMetric(c.systemActiveAccounts.Desc(), prometheus.GaugeValue, float64(overview.ActiveAccounts))
	ch <- prometheus.MustNewConstMetric(c.systemHealthScore.Desc(), prometheus.GaugeValue, overview.SystemHealth.HealthScore)
	ch <- prometheus.MustNewConstMetric(c.systemCapacityUtilization.Desc(), prometheus.GaugeValue, overview.CapacityStatus.CapacityUtilization)
	ch <- prometheus.MustNewConstMetric(c.systemAnomaliesTotal.Desc(), prometheus.GaugeValue, float64(len(overview.SystemAnomalies)))
	ch <- prometheus.MustNewConstMetric(c.systemAlertsTotal.Desc(), prometheus.GaugeValue, float64(len(overview.SystemAlerts)))
}
