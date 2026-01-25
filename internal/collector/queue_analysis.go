// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// QueueAnalysisCollector collects queue position analysis and wait time prediction metrics
type QueueAnalysisCollector struct {
	client QueueAnalysisSLURMClient

	// Queue Position Metrics
	queuePosition        *prometheus.GaugeVec
	queueDepth           *prometheus.GaugeVec
	queueMovement        *prometheus.GaugeVec
	queueAdvancementRate *prometheus.GaugeVec
	queueStagnation      *prometheus.GaugeVec

	// Wait Time Prediction Metrics
	predictedWaitTime          *prometheus.GaugeVec
	waitTimePredictionAccuracy *prometheus.GaugeVec
	waitTimeVariance           *prometheus.GaugeVec
	waitTimeConfidence         *prometheus.GaugeVec
	waitTimeModelPerformance   *prometheus.GaugeVec

	// Historical Wait Time Analysis
	actualWaitTime      *prometheus.HistogramVec
	waitTimePercentiles *prometheus.GaugeVec
	waitTimeTrends      *prometheus.GaugeVec
	waitTimeAnomalies   *prometheus.GaugeVec
	waitTimePatterns    *prometheus.GaugeVec

	// Queue Efficiency Metrics
	queueThroughput      *prometheus.GaugeVec
	queueProcessingRate  *prometheus.GaugeVec
	queueUtilization     *prometheus.GaugeVec
	queueBottlenecks     *prometheus.GaugeVec
	queueEfficiencyScore *prometheus.GaugeVec

	// Resource-Based Queue Analysis
	resourceQueueDepth     *prometheus.GaugeVec
	resourceWaitTime       *prometheus.GaugeVec
	resourceAvailability   *prometheus.GaugeVec
	resourceContention     *prometheus.GaugeVec
	resourceAllocationRate *prometheus.GaugeVec

	// Priority-Based Queue Analysis
	priorityQueueMetrics *prometheus.GaugeVec
	priorityWaitTime     *prometheus.GaugeVec
	priorityAdvantage    *prometheus.GaugeVec
	priorityInversion    *prometheus.GaugeVec
	priorityStarvation   *prometheus.GaugeVec

	// Partition Queue Analysis
	partitionQueueDepth  *prometheus.GaugeVec
	partitionWaitTime    *prometheus.GaugeVec
	partitionThroughput  *prometheus.GaugeVec
	partitionEfficiency  *prometheus.GaugeVec
	partitionLoadBalance *prometheus.GaugeVec

	// User Queue Experience
	userQueueMetrics       *prometheus.GaugeVec
	userWaitExperience     *prometheus.GaugeVec
	userQueueBehavior      *prometheus.GaugeVec
	userSubmissionStrategy *prometheus.GaugeVec
	userQueueOptimization  *prometheus.GaugeVec

	// Backfill Analysis
	backfillOpportunities *prometheus.GaugeVec
	backfillSuccess       *prometheus.CounterVec
	backfillEfficiency    *prometheus.GaugeVec
	backfillImpact        *prometheus.GaugeVec
	backfillOptimization  *prometheus.GaugeVec

	// Queue State Transitions
	queueStateChanges      *prometheus.CounterVec
	queueTransitionTime    *prometheus.HistogramVec
	queueStateDistribution *prometheus.GaugeVec
	queueHealthMetrics     *prometheus.GaugeVec

	// Prediction Model Metrics
	modelAccuracy          *prometheus.GaugeVec
	modelPredictionError   *prometheus.HistogramVec
	modelCalibration       *prometheus.GaugeVec
	modelFeatureImportance *prometheus.GaugeVec
	modelUpdateFrequency   *prometheus.CounterVec

	// System Load Impact
	systemLoadImpact           *prometheus.GaugeVec
	loadBalancingEffectiveness *prometheus.GaugeVec
	capacityUtilization        *prometheus.GaugeVec
	resourceFragmentation      *prometheus.GaugeVec
	schedulingOverhead         *prometheus.GaugeVec
}

// QueueAnalysisSLURMClient interface for queue analysis operations
type QueueAnalysisSLURMClient interface {
	GetQueuePositionAnalysis(ctx context.Context, jobID string) (*QueuePositionAnalysis, error)
	PredictWaitTime(ctx context.Context, jobID string) (*QueueWaitTimePrediction, error)
	GetQueueMetrics(ctx context.Context, partition string) (*QueueAnalysisMetrics, error)
	GetHistoricalWaitTimes(ctx context.Context, filters *WaitTimeFilters) (*HistoricalWaitTimes, error)
	GetQueueEfficiencyAnalysis(ctx context.Context, partition string) (*QueueEfficiencyAnalysis, error)
	GetResourceQueueAnalysis(ctx context.Context, resourceType string) (*ResourceQueueAnalysis, error)
	GetPriorityQueueAnalysis(ctx context.Context) (*PriorityQueueAnalysis, error)
	GetUserQueueExperience(ctx context.Context, userName string) (*UserQueueExperience, error)
	GetBackfillAnalysis(ctx context.Context, partition string) (*BackfillAnalysis, error)
	GetQueueStateTransitions(ctx context.Context, period string) (*QueueStateTransitions, error)
	ValidatePredictionModel(ctx context.Context) (*PredictionModelValidation, error)
	GetSystemLoadImpact(ctx context.Context) (*SystemLoadImpact, error)
}

// QueuePositionAnalysis represents queue position analysis for a job
type QueuePositionAnalysis struct {
	JobID         string `json:"job_id"`
	UserName      string `json:"user_name"`
	AccountName   string `json:"account_name"`
	PartitionName string `json:"partition_name"`

	// Current Position
	CurrentPosition    int     `json:"current_position"`
	TotalQueueDepth    int     `json:"total_queue_depth"`
	PositionPercentile float64 `json:"position_percentile"`
	JobsAhead          int     `json:"jobs_ahead"`
	JobsBehind         int     `json:"jobs_behind"`

	// Position History
	InitialPosition int           `json:"initial_position"`
	PositionChanges []int         `json:"position_changes"`
	AdvancementRate float64       `json:"advancement_rate"`
	TimeInQueue     time.Duration `json:"time_in_queue"`
	StagnationTime  time.Duration `json:"stagnation_time"`

	// Movement Analysis
	MovementDirection    string  `json:"movement_direction"`
	MovementVelocity     float64 `json:"movement_velocity"`
	MovementAcceleration float64 `json:"movement_acceleration"`
	MovementPrediction   string  `json:"movement_prediction"`

	// Queue Context
	QueueClass    string    `json:"queue_class"`
	PriorityClass string    `json:"priority_class"`
	ResourceClass string    `json:"resource_class"`
	SubmittedAt   time.Time `json:"submitted_at"`

	LastUpdated time.Time `json:"last_updated"`
}

// QueueWaitTimePrediction represents wait time prediction for a job
type QueueWaitTimePrediction struct {
	JobID         string `json:"job_id"`
	UserName      string `json:"user_name"`
	PartitionName string `json:"partition_name"`

	// Primary Predictions
	PredictedWaitTime  time.Duration `json:"predicted_wait_time"`
	EstimatedStartTime time.Time     `json:"estimated_start_time"`
	ConfidenceLevel    float64       `json:"confidence_level"`
	PredictionMethod   string        `json:"prediction_method"`

	// Prediction Ranges
	MinWaitTime      time.Duration `json:"min_wait_time"`
	MaxWaitTime      time.Duration `json:"max_wait_time"`
	MedianWaitTime   time.Duration `json:"median_wait_time"`
	P90WaitTime      time.Duration `json:"p90_wait_time"`
	UncertaintyRange time.Duration `json:"uncertainty_range"`

	// Factors Influencing Prediction
	QueuePositionFactor float64 `json:"queue_position_factor"`
	PriorityFactor      float64 `json:"priority_factor"`
	ResourceFactor      float64 `json:"resource_factor"`
	HistoricalFactor    float64 `json:"historical_factor"`
	SystemLoadFactor    float64 `json:"system_load_factor"`

	// Model Information
	ModelVersion     string             `json:"model_version"`
	ModelAccuracy    float64            `json:"model_accuracy"`
	FeatureWeights   map[string]float64 `json:"feature_weights"`
	TrainingDataSize int                `json:"training_data_size"`

	// Update Information
	PredictedAt     time.Time     `json:"predicted_at"`
	NextUpdate      time.Time     `json:"next_update"`
	UpdateFrequency time.Duration `json:"update_frequency"`
}

// QueueAnalysisMetrics represents comprehensive queue metrics
type QueueAnalysisMetrics struct {
	PartitionName string `json:"partition_name"`

	// Basic Queue Stats
	TotalJobs     int `json:"total_jobs"`
	PendingJobs   int `json:"pending_jobs"`
	RunningJobs   int `json:"running_jobs"`
	CompletedJobs int `json:"completed_jobs"`
	FailedJobs    int `json:"failed_jobs"`

	// Queue Depth Analysis
	AverageQueueDepth  float64 `json:"average_queue_depth"`
	MaxQueueDepth      int     `json:"max_queue_depth"`
	MinQueueDepth      int     `json:"min_queue_depth"`
	QueueDepthVariance float64 `json:"queue_depth_variance"`
	QueueDepthTrend    string  `json:"queue_depth_trend"`

	// Throughput Metrics
	JobsPerHour           float64 `json:"jobs_per_hour"`
	JobsPerDay            float64 `json:"jobs_per_day"`
	ThroughputTrend       string  `json:"throughput_trend"`
	ThroughputVariability float64 `json:"throughput_variability"`
	PeakThroughput        float64 `json:"peak_throughput"`

	// Processing Metrics
	AverageProcessingTime float64 `json:"average_processing_time"`
	ProcessingEfficiency  float64 `json:"processing_efficiency"`
	ResourceUtilization   float64 `json:"resource_utilization"`
	QueueUtilization      float64 `json:"queue_utilization"`

	// Health Indicators
	QueueHealth        float64 `json:"queue_health"`
	BottleneckSeverity float64 `json:"bottleneck_severity"`
	StarvationRisk     float64 `json:"starvation_risk"`
	OverloadRisk       float64 `json:"overload_risk"`

	LastUpdated time.Time `json:"last_updated"`
}

// HistoricalWaitTimes represents historical wait time analysis
type HistoricalWaitTimes struct {
	AnalysisPeriod string `json:"analysis_period"`
	TotalJobs      int    `json:"total_jobs"`

	// Wait Time Statistics
	MeanWaitTime   float64 `json:"mean_wait_time"`
	MedianWaitTime float64 `json:"median_wait_time"`
	P90WaitTime    float64 `json:"p90_wait_time"`
	P95WaitTime    float64 `json:"p95_wait_time"`
	P99WaitTime    float64 `json:"p99_wait_time"`
	MaxWaitTime    float64 `json:"max_wait_time"`
	MinWaitTime    float64 `json:"min_wait_time"`
	StdDevWaitTime float64 `json:"stddev_wait_time"`

	// Trends and Patterns
	WaitTimeTrend    string             `json:"wait_time_trend"`
	SeasonalPatterns map[string]float64 `json:"seasonal_patterns"`
	DailyPatterns    map[string]float64 `json:"daily_patterns"`
	WeeklyPatterns   map[string]float64 `json:"weekly_patterns"`

	// Categorized Analysis
	WaitTimeByPriority  map[string]float64 `json:"wait_time_by_priority"`
	WaitTimeByPartition map[string]float64 `json:"wait_time_by_partition"`
	WaitTimeByUser      map[string]float64 `json:"wait_time_by_user"`
	WaitTimeByResource  map[string]float64 `json:"wait_time_by_resource"`

	// Anomalies
	AnomalousJobs    []string `json:"anomalous_jobs"`
	AnomalyThreshold float64  `json:"anomaly_threshold"`
	AnomalyRate      float64  `json:"anomaly_rate"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// QueueEfficiencyAnalysis represents queue efficiency analysis
type QueueEfficiencyAnalysis struct {
	PartitionName string `json:"partition_name"`

	// Efficiency Metrics
	OverallEfficiency    float64 `json:"overall_efficiency"`
	SchedulingEfficiency float64 `json:"scheduling_efficiency"`
	ResourceEfficiency   float64 `json:"resource_efficiency"`
	ThroughputEfficiency float64 `json:"throughput_efficiency"`
	WaitTimeEfficiency   float64 `json:"wait_time_efficiency"`

	// Bottleneck Analysis
	BottleneckType     string   `json:"bottleneck_type"`
	BottleneckSeverity float64  `json:"bottleneck_severity"`
	BottleneckImpact   float64  `json:"bottleneck_impact"`
	BottleneckSources  []string `json:"bottleneck_sources"`

	// Optimization Opportunities
	OptimizationPotential float64  `json:"optimization_potential"`
	RecommendedActions    []string `json:"recommended_actions"`
	ExpectedImprovement   float64  `json:"expected_improvement"`
	ImplementationCost    string   `json:"implementation_cost"`

	// Performance Indicators
	ResponseTime     float64 `json:"response_time"`
	ServiceLevel     float64 `json:"service_level"`
	UserSatisfaction float64 `json:"user_satisfaction"`
	SystemStress     float64 `json:"system_stress"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// ResourceQueueAnalysis represents resource-specific queue analysis
type ResourceQueueAnalysis struct {
	ResourceType string `json:"resource_type"`

	// Resource Availability
	TotalCapacity       float64 `json:"total_capacity"`
	AvailableCapacity   float64 `json:"available_capacity"`
	UtilizedCapacity    float64 `json:"utilized_capacity"`
	ReservedCapacity    float64 `json:"reserved_capacity"`
	CapacityUtilization float64 `json:"capacity_utilization"`

	// Queue Metrics by Resource
	JobsWaitingForResource  int     `json:"jobs_waiting_for_resource"`
	AverageResourceWaitTime float64 `json:"average_resource_wait_time"`
	ResourceContention      float64 `json:"resource_contention"`
	ResourceFragmentation   float64 `json:"resource_fragmentation"`

	// Allocation Efficiency
	AllocationRate       float64 `json:"allocation_rate"`
	AllocationEfficiency float64 `json:"allocation_efficiency"`
	WasteRate            float64 `json:"waste_rate"`
	FragmentationImpact  float64 `json:"fragmentation_impact"`

	// Predictions
	FutureAvailability float64 `json:"future_availability"`
	ExpectedWaitTime   float64 `json:"expected_wait_time"`
	AllocationForecast string  `json:"allocation_forecast"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// PriorityQueueAnalysis represents priority-based queue analysis
type PriorityQueueAnalysis struct {
	// Priority Level Analysis
	PriorityLevels       map[string]*PriorityLevelMetrics `json:"priority_levels"`
	PriorityDistribution map[string]float64               `json:"priority_distribution"`

	// Priority Fairness
	FairnessScore     float64 `json:"fairness_score"`
	PriorityAdvantage float64 `json:"priority_advantage"`
	StarvationRisk    float64 `json:"starvation_risk"`
	PriorityInversion float64 `json:"priority_inversion"`

	// Queue Dynamics
	PriorityMobility   float64 `json:"priority_mobility"`
	AgingEffectiveness float64 `json:"aging_effectiveness"`
	PreemptionRate     float64 `json:"preemption_rate"`
	EscalationRate     float64 `json:"escalation_rate"`

	// System Health
	PrioritySystemHealth float64 `json:"priority_system_health"`
	PolicyEffectiveness  float64 `json:"policy_effectiveness"`
	OptimizationNeeded   bool    `json:"optimization_needed"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// PriorityLevelMetrics represents metrics for a specific priority level
type PriorityLevelMetrics struct {
	PriorityLevel   string  `json:"priority_level"`
	JobCount        int     `json:"job_count"`
	AverageWaitTime float64 `json:"average_wait_time"`
	MedianWaitTime  float64 `json:"median_wait_time"`
	ThroughputRate  float64 `json:"throughput_rate"`
	QueuePosition   float64 `json:"queue_position"`
	StarvationRisk  float64 `json:"starvation_risk"`
	AdvancementRate float64 `json:"advancement_rate"`
}

// UserQueueExperience represents user-specific queue experience
type UserQueueExperience struct {
	UserName    string `json:"user_name"`
	AccountName string `json:"account_name"`

	// User Queue Statistics
	TotalSubmissions     int     `json:"total_submissions"`
	AverageWaitTime      float64 `json:"average_wait_time"`
	MedianWaitTime       float64 `json:"median_wait_time"`
	WaitTimeVariability  float64 `json:"wait_time_variability"`
	QueueExperienceScore float64 `json:"queue_experience_score"`

	// Behavioral Analysis
	SubmissionPatterns     []string `json:"submission_patterns"`
	OptimalSubmissionTimes []string `json:"optimal_submission_times"`
	SubmissionStrategy     string   `json:"submission_strategy"`
	BehaviorOptimization   float64  `json:"behavior_optimization"`

	// Performance Metrics
	JobSuccessRate      float64 `json:"job_success_rate"`
	ResourceEfficiency  float64 `json:"resource_efficiency"`
	QueuePosition       float64 `json:"queue_position"`
	PriorityUtilization float64 `json:"priority_utilization"`

	// Recommendations
	WaitTimeReduction float64  `json:"wait_time_reduction"`
	OptimizationTips  []string `json:"optimization_tips"`
	BestPractices     []string `json:"best_practices"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// BackfillAnalysis represents backfill scheduling analysis
type BackfillAnalysis struct {
	PartitionName string `json:"partition_name"`

	// Backfill Opportunities
	TotalOpportunities    int     `json:"total_opportunities"`
	UtilizedOpportunities int     `json:"utilized_opportunities"`
	MissedOpportunities   int     `json:"missed_opportunities"`
	BackfillRate          float64 `json:"backfill_rate"`
	BackfillEfficiency    float64 `json:"backfill_efficiency"`

	// Impact Analysis
	JobsBackfilled         int     `json:"jobs_backfilled"`
	WaitTimeReduction      float64 `json:"wait_time_reduction"`
	UtilizationImprovement float64 `json:"utilization_improvement"`
	ThroughputIncrease     float64 `json:"throughput_increase"`

	// Backfill Characteristics
	AverageBackfillDuration float64        `json:"average_backfill_duration"`
	BackfillJobSizes        map[string]int `json:"backfill_job_sizes"`
	BackfillSuccess         float64        `json:"backfill_success"`

	// Optimization Potential
	OptimizationScore    float64  `json:"optimization_score"`
	ImprovementPotential float64  `json:"improvement_potential"`
	RecommendedChanges   []string `json:"recommended_changes"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// QueueStateTransitions represents queue state transition analysis
type QueueStateTransitions struct {
	AnalysisPeriod string `json:"analysis_period"`

	// State Transition Counts
	StateTransitions  map[string]map[string]int `json:"state_transitions"`
	TransitionRates   map[string]float64        `json:"transition_rates"`
	StateDistribution map[string]float64        `json:"state_distribution"`

	// Transition Time Analysis
	AverageTransitionTime  map[string]float64 `json:"average_transition_time"`
	TransitionTimeVariance map[string]float64 `json:"transition_time_variance"`

	// Queue Health Indicators
	HealthyTransitions     float64 `json:"healthy_transitions"`
	ProblematicTransitions float64 `json:"problematic_transitions"`
	TransitionEfficiency   float64 `json:"transition_efficiency"`
	StateStability         float64 `json:"state_stability"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// PredictionModelValidation represents prediction model validation
type PredictionModelValidation struct {
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`

	// Accuracy Metrics
	OverallAccuracy     float64 `json:"overall_accuracy"`
	PredictionError     float64 `json:"prediction_error"`
	MeanAbsoluteError   float64 `json:"mean_absolute_error"`
	RootMeanSquareError float64 `json:"root_mean_square_error"`
	R2Score             float64 `json:"r2_score"`

	// Calibration Metrics
	CalibrationScore   float64 `json:"calibration_score"`
	ConfidenceAccuracy float64 `json:"confidence_accuracy"`
	UncertaintyCapture float64 `json:"uncertainty_capture"`

	// Feature Analysis
	FeatureImportance   map[string]float64 `json:"feature_importance"`
	FeatureCorrelations map[string]float64 `json:"feature_correlations"`
	FeatureStability    float64            `json:"feature_stability"`

	// Performance Analysis
	PredictionLatency float64 `json:"prediction_latency"`
	ModelComplexity   float64 `json:"model_complexity"`
	UpdateFrequency   float64 `json:"update_frequency"`
	RetrainingNeeded  bool    `json:"retraining_needed"`

	ValidatedAt time.Time `json:"validated_at"`
}

// SystemLoadImpact represents system load impact on queue performance
type SystemLoadImpact struct {
	// System Load Metrics
	CPULoad           float64 `json:"cpu_load"`
	MemoryLoad        float64 `json:"memory_load"`
	NetworkLoad       float64 `json:"network_load"`
	StorageLoad       float64 `json:"storage_load"`
	OverallSystemLoad float64 `json:"overall_system_load"`

	// Impact on Queue Performance
	QueuePerformanceImpact float64 `json:"queue_performance_impact"`
	WaitTimeIncrease       float64 `json:"wait_time_increase"`
	ThroughputDecrease     float64 `json:"throughput_decrease"`
	EfficiencyReduction    float64 `json:"efficiency_reduction"`

	// Load Balancing
	LoadBalanceScore         float64            `json:"load_balance_score"`
	LoadDistribution         map[string]float64 `json:"load_distribution"`
	LoadBalanceEffectiveness float64            `json:"load_balance_effectiveness"`

	// Capacity Management
	CapacityUtilization   float64 `json:"capacity_utilization"`
	CapacityFragmentation float64 `json:"capacity_fragmentation"`
	CapacityOptimization  float64 `json:"capacity_optimization"`

	LastMeasured time.Time `json:"last_measured"`
}

// WaitTimeFilters represents filters for historical wait time analysis
type WaitTimeFilters struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	UserName      string    `json:"user_name"`
	AccountName   string    `json:"account_name"`
	PartitionName string    `json:"partition_name"`
	MinPriority   int       `json:"min_priority"`
	MaxPriority   int       `json:"max_priority"`
	JobSizeMin    int       `json:"job_size_min"`
	JobSizeMax    int       `json:"job_size_max"`
}

// NewQueueAnalysisCollector creates a new queue analysis collector
func NewQueueAnalysisCollector(client QueueAnalysisSLURMClient) *QueueAnalysisCollector {
	return &QueueAnalysisCollector{
		client: client,

		// Queue Position Metrics
		queuePosition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_position",
				Help: "Current position of jobs in queue",
			},
			[]string{"job_id", "user", "account", "partition", "priority_class"},
		),

		queueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_depth",
				Help: "Current depth of job queue",
			},
			[]string{"partition", "queue_class", "metric"},
		),

		queueMovement: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_movement",
				Help: "Queue position movement metrics",
			},
			[]string{"job_id", "user", "account", "partition", "movement_type"},
		),

		queueAdvancementRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_advancement_rate",
				Help: "Rate of advancement in queue",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		queueStagnation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_stagnation_time_minutes",
				Help: "Time job has been stagnant in queue in minutes",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		// Wait Time Prediction Metrics
		predictedWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_predicted_wait_time_seconds",
				Help: "Predicted wait time for jobs in seconds",
			},
			[]string{"job_id", "user", "account", "partition", "prediction_method"},
		),

		waitTimePredictionAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_prediction_accuracy",
				Help: "Accuracy of wait time predictions",
			},
			[]string{"prediction_method", "model_version", "time_range"},
		),

		waitTimeVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_variance",
				Help: "Variance in wait time predictions",
			},
			[]string{"job_id", "user", "account", "partition", "variance_type"},
		),

		waitTimeConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_prediction_confidence",
				Help: "Confidence level of wait time predictions",
			},
			[]string{"job_id", "user", "account", "partition", "prediction_method"},
		),

		waitTimeModelPerformance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_model_performance",
				Help: "Performance metrics of wait time prediction models",
			},
			[]string{"model_version", "performance_metric"},
		),

		// Historical Wait Time Analysis
		actualWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_actual_wait_time_seconds",
				Help:    "Distribution of actual wait times in seconds",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 28800, 86400},
			},
			[]string{"user", "account", "partition", "priority_class"},
		),

		waitTimePercentiles: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_percentiles_seconds",
				Help: "Wait time percentiles in seconds",
			},
			[]string{"partition", "priority_class", "percentile"},
		),

		waitTimeTrends: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_trends",
				Help: "Wait time trend analysis",
			},
			[]string{"partition", "trend_type", "period"},
		),

		waitTimeAnomalies: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_anomalies",
				Help: "Wait time anomaly detection metrics",
			},
			[]string{"partition", "anomaly_type", "severity"},
		),

		waitTimePatterns: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_wait_time_patterns",
				Help: "Wait time pattern analysis",
			},
			[]string{"partition", "pattern_type", "time_period"},
		),

		// Queue Efficiency Metrics
		queueThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_throughput_jobs_per_hour",
				Help: "Queue throughput in jobs per hour",
			},
			[]string{"partition", "throughput_type"},
		),

		queueProcessingRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_processing_rate",
				Help: "Queue processing rate",
			},
			[]string{"partition", "rate_type"},
		),

		queueUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_utilization",
				Help: "Queue utilization metrics",
			},
			[]string{"partition", "utilization_type"},
		),

		queueBottlenecks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_bottlenecks",
				Help: "Queue bottleneck analysis",
			},
			[]string{"partition", "bottleneck_type", "severity"},
		),

		queueEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_efficiency_score",
				Help: "Overall queue efficiency score",
			},
			[]string{"partition", "efficiency_type"},
		),

		// Resource-Based Queue Analysis
		resourceQueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_queue_depth",
				Help: "Queue depth by resource type",
			},
			[]string{"resource_type", "partition"},
		),

		resourceWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_wait_time_seconds",
				Help: "Wait time by resource type in seconds",
			},
			[]string{"resource_type", "partition", "wait_type"},
		),

		resourceAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_availability",
				Help: "Resource availability metrics",
			},
			[]string{"resource_type", "partition", "availability_type"},
		),

		resourceContention: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_contention",
				Help: "Resource contention metrics",
			},
			[]string{"resource_type", "partition"},
		),

		resourceAllocationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_allocation_rate",
				Help: "Resource allocation rate",
			},
			[]string{"resource_type", "partition"},
		),

		// Priority-Based Queue Analysis
		priorityQueueMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_queue_metrics",
				Help: "Priority-based queue metrics",
			},
			[]string{"priority_level", "metric_type"},
		),

		priorityWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_wait_time_seconds",
				Help: "Wait time by priority level in seconds",
			},
			[]string{"priority_level", "wait_type"},
		),

		priorityAdvantage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_advantage",
				Help: "Priority advantage metrics",
			},
			[]string{"priority_level", "advantage_type"},
		),

		priorityInversion: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_inversion",
				Help: "Priority inversion detection",
			},
			[]string{"partition", "inversion_type"},
		),

		priorityStarvation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_starvation_risk",
				Help: "Priority starvation risk assessment",
			},
			[]string{"priority_level", "partition"},
		),

		// Partition Queue Analysis
		partitionQueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_queue_depth",
				Help: "Queue depth by partition",
			},
			[]string{"partition", "job_state"},
		),

		partitionWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_wait_time_seconds",
				Help: "Wait time by partition in seconds",
			},
			[]string{"partition", "wait_type"},
		),

		partitionThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_throughput",
				Help: "Throughput by partition",
			},
			[]string{"partition", "throughput_type"},
		),

		partitionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_efficiency",
				Help: "Efficiency by partition",
			},
			[]string{"partition", "efficiency_type"},
		),

		partitionLoadBalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_partition_load_balance",
				Help: "Load balance across partitions",
			},
			[]string{"partition", "balance_type"},
		),

		// User Queue Experience
		userQueueMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_queue_metrics",
				Help: "User-specific queue metrics",
			},
			[]string{"user", "account", "metric_type"},
		),

		userWaitExperience: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_wait_experience",
				Help: "User wait experience metrics",
			},
			[]string{"user", "account", "experience_type"},
		),

		userQueueBehavior: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_queue_behavior",
				Help: "User queue behavior analysis",
			},
			[]string{"user", "account", "behavior_type"},
		),

		userSubmissionStrategy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_submission_strategy",
				Help: "User submission strategy analysis",
			},
			[]string{"user", "account", "strategy_type"},
		),

		userQueueOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_queue_optimization",
				Help: "User queue optimization potential",
			},
			[]string{"user", "account", "optimization_type"},
		),

		// Backfill Analysis
		backfillOpportunities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_backfill_opportunities",
				Help: "Backfill scheduling opportunities",
			},
			[]string{"partition", "opportunity_type"},
		),

		backfillSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_backfill_success_total",
				Help: "Total backfill successes",
			},
			[]string{"partition", "success_type"},
		),

		backfillEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_backfill_efficiency",
				Help: "Backfill scheduling efficiency",
			},
			[]string{"partition", "efficiency_type"},
		),

		backfillImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_backfill_impact",
				Help: "Impact of backfill scheduling",
			},
			[]string{"partition", "impact_type"},
		),

		backfillOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_backfill_optimization",
				Help: "Backfill optimization potential",
			},
			[]string{"partition", "optimization_type"},
		),

		// Queue State Transitions
		queueStateChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_state_changes_total",
				Help: "Total queue state changes",
			},
			[]string{"from_state", "to_state", "partition"},
		),

		queueTransitionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_transition_time_seconds",
				Help:    "Time for queue state transitions in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800},
			},
			[]string{"from_state", "to_state", "partition"},
		),

		queueStateDistribution: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_state_distribution",
				Help: "Distribution of jobs across queue states",
			},
			[]string{"state", "partition"},
		),

		queueHealthMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_health_metrics",
				Help: "Queue health and stability metrics",
			},
			[]string{"partition", "health_metric"},
		),

		// Prediction Model Metrics
		modelAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_model_accuracy",
				Help: "Prediction model accuracy metrics",
			},
			[]string{"model_name", "model_version", "accuracy_type"},
		),

		modelPredictionError: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_prediction_model_error",
				Help:    "Prediction model error distribution",
				Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"model_name", "model_version", "error_type"},
		),

		modelCalibration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_model_calibration",
				Help: "Prediction model calibration metrics",
			},
			[]string{"model_name", "model_version", "calibration_type"},
		),

		modelFeatureImportance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_model_feature_importance",
				Help: "Feature importance in prediction models",
			},
			[]string{"model_name", "model_version", "feature"},
		),

		modelUpdateFrequency: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_prediction_model_updates_total",
				Help: "Total prediction model updates",
			},
			[]string{"model_name", "update_type"},
		),

		// System Load Impact
		systemLoadImpact: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_load_impact",
				Help: "Impact of system load on queue performance",
			},
			[]string{"load_type", "impact_metric"},
		),

		loadBalancingEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_load_balancing_effectiveness",
				Help: "Effectiveness of load balancing",
			},
			[]string{"balancing_type", "effectiveness_metric"},
		),

		capacityUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_capacity_utilization",
				Help: "System capacity utilization",
			},
			[]string{"capacity_type", "utilization_metric"},
		),

		resourceFragmentation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_resource_fragmentation",
				Help: "Resource fragmentation metrics",
			},
			[]string{"resource_type", "fragmentation_type"},
		),

		schedulingOverhead: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_scheduling_overhead",
				Help: "Scheduling overhead metrics",
			},
			[]string{"overhead_type"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *QueueAnalysisCollector) Describe(ch chan<- *prometheus.Desc) {
	c.queuePosition.Describe(ch)
	c.queueDepth.Describe(ch)
	c.queueMovement.Describe(ch)
	c.queueAdvancementRate.Describe(ch)
	c.queueStagnation.Describe(ch)
	c.predictedWaitTime.Describe(ch)
	c.waitTimePredictionAccuracy.Describe(ch)
	c.waitTimeVariance.Describe(ch)
	c.waitTimeConfidence.Describe(ch)
	c.waitTimeModelPerformance.Describe(ch)
	c.actualWaitTime.Describe(ch)
	c.waitTimePercentiles.Describe(ch)
	c.waitTimeTrends.Describe(ch)
	c.waitTimeAnomalies.Describe(ch)
	c.waitTimePatterns.Describe(ch)
	c.queueThroughput.Describe(ch)
	c.queueProcessingRate.Describe(ch)
	c.queueUtilization.Describe(ch)
	c.queueBottlenecks.Describe(ch)
	c.queueEfficiencyScore.Describe(ch)
	c.resourceQueueDepth.Describe(ch)
	c.resourceWaitTime.Describe(ch)
	c.resourceAvailability.Describe(ch)
	c.resourceContention.Describe(ch)
	c.resourceAllocationRate.Describe(ch)
	c.priorityQueueMetrics.Describe(ch)
	c.priorityWaitTime.Describe(ch)
	c.priorityAdvantage.Describe(ch)
	c.priorityInversion.Describe(ch)
	c.priorityStarvation.Describe(ch)
	c.partitionQueueDepth.Describe(ch)
	c.partitionWaitTime.Describe(ch)
	c.partitionThroughput.Describe(ch)
	c.partitionEfficiency.Describe(ch)
	c.partitionLoadBalance.Describe(ch)
	c.userQueueMetrics.Describe(ch)
	c.userWaitExperience.Describe(ch)
	c.userQueueBehavior.Describe(ch)
	c.userSubmissionStrategy.Describe(ch)
	c.userQueueOptimization.Describe(ch)
	c.backfillOpportunities.Describe(ch)
	c.backfillSuccess.Describe(ch)
	c.backfillEfficiency.Describe(ch)
	c.backfillImpact.Describe(ch)
	c.backfillOptimization.Describe(ch)
	c.queueStateChanges.Describe(ch)
	c.queueTransitionTime.Describe(ch)
	c.queueStateDistribution.Describe(ch)
	c.queueHealthMetrics.Describe(ch)
	c.modelAccuracy.Describe(ch)
	c.modelPredictionError.Describe(ch)
	c.modelCalibration.Describe(ch)
	c.modelFeatureImportance.Describe(ch)
	c.modelUpdateFrequency.Describe(ch)
	c.systemLoadImpact.Describe(ch)
	c.loadBalancingEffectiveness.Describe(ch)
	c.capacityUtilization.Describe(ch)
	c.resourceFragmentation.Describe(ch)
	c.schedulingOverhead.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *QueueAnalysisCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Reset metrics
	c.resetMetrics()

	// Get sample partitions for analysis
	samplePartitions := c.getSamplePartitions()
	for _, partition := range samplePartitions {
		c.collectPartitionQueueMetrics(ctx, partition)
	}

	// Get sample job IDs for position analysis
	sampleJobIDs := c.getSampleJobIDs()
	for _, jobID := range sampleJobIDs {
		c.collectJobQueuePosition(ctx, jobID)
		c.collectWaitTimePrediction(ctx, jobID)
	}

	// Get sample users for experience analysis
	sampleUsers := c.getSampleUsers()
	for _, user := range sampleUsers {
		c.collectUserQueueExperience(ctx, user)
	}

	// Collect historical wait time analysis
	c.collectHistoricalWaitTimes(ctx)

	// Collect resource-based analysis
	c.collectResourceQueueAnalysis(ctx)

	// Collect priority-based analysis
	c.collectPriorityQueueAnalysis(ctx)

	// Collect backfill analysis
	c.collectBackfillAnalysis(ctx)

	// Collect queue state transitions
	c.collectQueueStateTransitions(ctx)

	// Collect prediction model validation
	c.collectPredictionModelValidation(ctx)

	// Collect system load impact
	c.collectSystemLoadImpact(ctx)

	// Collect all metrics
	c.queuePosition.Collect(ch)
	c.queueDepth.Collect(ch)
	c.queueMovement.Collect(ch)
	c.queueAdvancementRate.Collect(ch)
	c.queueStagnation.Collect(ch)
	c.predictedWaitTime.Collect(ch)
	c.waitTimePredictionAccuracy.Collect(ch)
	c.waitTimeVariance.Collect(ch)
	c.waitTimeConfidence.Collect(ch)
	c.waitTimeModelPerformance.Collect(ch)
	c.actualWaitTime.Collect(ch)
	c.waitTimePercentiles.Collect(ch)
	c.waitTimeTrends.Collect(ch)
	c.waitTimeAnomalies.Collect(ch)
	c.waitTimePatterns.Collect(ch)
	c.queueThroughput.Collect(ch)
	c.queueProcessingRate.Collect(ch)
	c.queueUtilization.Collect(ch)
	c.queueBottlenecks.Collect(ch)
	c.queueEfficiencyScore.Collect(ch)
	c.resourceQueueDepth.Collect(ch)
	c.resourceWaitTime.Collect(ch)
	c.resourceAvailability.Collect(ch)
	c.resourceContention.Collect(ch)
	c.resourceAllocationRate.Collect(ch)
	c.priorityQueueMetrics.Collect(ch)
	c.priorityWaitTime.Collect(ch)
	c.priorityAdvantage.Collect(ch)
	c.priorityInversion.Collect(ch)
	c.priorityStarvation.Collect(ch)
	c.partitionQueueDepth.Collect(ch)
	c.partitionWaitTime.Collect(ch)
	c.partitionThroughput.Collect(ch)
	c.partitionEfficiency.Collect(ch)
	c.partitionLoadBalance.Collect(ch)
	c.userQueueMetrics.Collect(ch)
	c.userWaitExperience.Collect(ch)
	c.userQueueBehavior.Collect(ch)
	c.userSubmissionStrategy.Collect(ch)
	c.userQueueOptimization.Collect(ch)
	c.backfillOpportunities.Collect(ch)
	c.backfillSuccess.Collect(ch)
	c.backfillEfficiency.Collect(ch)
	c.backfillImpact.Collect(ch)
	c.backfillOptimization.Collect(ch)
	c.queueStateChanges.Collect(ch)
	c.queueTransitionTime.Collect(ch)
	c.queueStateDistribution.Collect(ch)
	c.queueHealthMetrics.Collect(ch)
	c.modelAccuracy.Collect(ch)
	c.modelPredictionError.Collect(ch)
	c.modelCalibration.Collect(ch)
	c.modelFeatureImportance.Collect(ch)
	c.modelUpdateFrequency.Collect(ch)
	c.systemLoadImpact.Collect(ch)
	c.loadBalancingEffectiveness.Collect(ch)
	c.capacityUtilization.Collect(ch)
	c.resourceFragmentation.Collect(ch)
	c.schedulingOverhead.Collect(ch)
}

// resetQueueMetrics resets queue-related metrics
func (c *QueueAnalysisCollector) resetQueueMetrics() {
	c.queuePosition.Reset()
	c.queueDepth.Reset()
	c.queueMovement.Reset()
	c.queueAdvancementRate.Reset()
	c.queueStagnation.Reset()
	c.queueThroughput.Reset()
	c.queueProcessingRate.Reset()
	c.queueUtilization.Reset()
	c.queueBottlenecks.Reset()
	c.queueEfficiencyScore.Reset()
	c.queueStateDistribution.Reset()
	c.queueHealthMetrics.Reset()
}

// resetWaitTimeMetrics resets wait time analysis metrics
func (c *QueueAnalysisCollector) resetWaitTimeMetrics() {
	c.predictedWaitTime.Reset()
	c.waitTimePredictionAccuracy.Reset()
	c.waitTimeVariance.Reset()
	c.waitTimeConfidence.Reset()
	c.waitTimeModelPerformance.Reset()
	c.waitTimePercentiles.Reset()
	c.waitTimeTrends.Reset()
	c.waitTimeAnomalies.Reset()
	c.waitTimePatterns.Reset()
}

// resetResourceMetrics resets resource-related metrics
func (c *QueueAnalysisCollector) resetResourceMetrics() {
	c.resourceQueueDepth.Reset()
	c.resourceWaitTime.Reset()
	c.resourceAvailability.Reset()
	c.resourceContention.Reset()
	c.resourceAllocationRate.Reset()
}

// resetPriorityMetrics resets priority-related metrics
func (c *QueueAnalysisCollector) resetPriorityMetrics() {
	c.priorityQueueMetrics.Reset()
	c.priorityWaitTime.Reset()
	c.priorityAdvantage.Reset()
	c.priorityInversion.Reset()
	c.priorityStarvation.Reset()
}

// resetPartitionMetrics resets partition-related metrics
func (c *QueueAnalysisCollector) resetPartitionMetrics() {
	c.partitionQueueDepth.Reset()
	c.partitionWaitTime.Reset()
	c.partitionThroughput.Reset()
	c.partitionEfficiency.Reset()
	c.partitionLoadBalance.Reset()
}

// resetUserMetrics resets user-related metrics
func (c *QueueAnalysisCollector) resetUserMetrics() {
	c.userQueueMetrics.Reset()
	c.userWaitExperience.Reset()
	c.userQueueBehavior.Reset()
	c.userSubmissionStrategy.Reset()
	c.userQueueOptimization.Reset()
}

// resetBackfillMetrics resets backfill-related metrics
func (c *QueueAnalysisCollector) resetBackfillMetrics() {
	c.backfillOpportunities.Reset()
	c.backfillEfficiency.Reset()
	c.backfillImpact.Reset()
	c.backfillOptimization.Reset()
}

// resetModelMetrics resets model accuracy and performance metrics
func (c *QueueAnalysisCollector) resetModelMetrics() {
	c.modelAccuracy.Reset()
	c.modelCalibration.Reset()
	c.modelFeatureImportance.Reset()
}

// resetSystemMetrics resets system-level metrics
func (c *QueueAnalysisCollector) resetSystemMetrics() {
	c.systemLoadImpact.Reset()
	c.loadBalancingEffectiveness.Reset()
	c.capacityUtilization.Reset()
	c.resourceFragmentation.Reset()
	c.schedulingOverhead.Reset()
}

func (c *QueueAnalysisCollector) resetMetrics() {
	c.resetQueueMetrics()
	c.resetWaitTimeMetrics()
	c.resetResourceMetrics()
	c.resetPriorityMetrics()
	c.resetPartitionMetrics()
	c.resetUserMetrics()
	c.resetBackfillMetrics()
	c.resetModelMetrics()
	c.resetSystemMetrics()
}

func (c *QueueAnalysisCollector) collectPartitionQueueMetrics(ctx context.Context, partition string) {
	metrics, err := c.client.GetQueueMetrics(ctx, partition)
	if err != nil {
		log.Printf("Error collecting queue metrics for partition %s: %v", partition, err)
		return
	}

	c.queueDepth.WithLabelValues(partition, "all", "total").Set(float64(metrics.TotalJobs))
	c.queueDepth.WithLabelValues(partition, "all", "pending").Set(float64(metrics.PendingJobs))
	c.queueDepth.WithLabelValues(partition, "all", "running").Set(float64(metrics.RunningJobs))
	c.queueDepth.WithLabelValues(partition, "all", "completed").Set(float64(metrics.CompletedJobs))
	c.queueDepth.WithLabelValues(partition, "all", "failed").Set(float64(metrics.FailedJobs))

	c.queueDepth.WithLabelValues(partition, "depth", "average").Set(metrics.AverageQueueDepth)
	c.queueDepth.WithLabelValues(partition, "depth", "max").Set(float64(metrics.MaxQueueDepth))
	c.queueDepth.WithLabelValues(partition, "depth", "min").Set(float64(metrics.MinQueueDepth))
	c.queueDepth.WithLabelValues(partition, "depth", "variance").Set(metrics.QueueDepthVariance)

	c.queueThroughput.WithLabelValues(partition, "hourly").Set(metrics.JobsPerHour)
	c.queueThroughput.WithLabelValues(partition, "daily").Set(metrics.JobsPerDay)
	c.queueThroughput.WithLabelValues(partition, "peak").Set(metrics.PeakThroughput)

	c.queueProcessingRate.WithLabelValues(partition, "average").Set(metrics.AverageProcessingTime)
	c.queueProcessingRate.WithLabelValues(partition, "efficiency").Set(metrics.ProcessingEfficiency)

	c.queueUtilization.WithLabelValues(partition, "resource").Set(metrics.ResourceUtilization)
	c.queueUtilization.WithLabelValues(partition, "queue").Set(metrics.QueueUtilization)

	c.queueHealthMetrics.WithLabelValues(partition, "health").Set(metrics.QueueHealth)
	c.queueBottlenecks.WithLabelValues(partition, "severity", "moderate").Set(metrics.BottleneckSeverity)
	c.queueHealthMetrics.WithLabelValues(partition, "starvation_risk").Set(metrics.StarvationRisk)
	c.queueHealthMetrics.WithLabelValues(partition, "overload_risk").Set(metrics.OverloadRisk)

	// Collect queue efficiency analysis
	efficiency, err := c.client.GetQueueEfficiencyAnalysis(ctx, partition)
	if err == nil {
		c.queueEfficiencyScore.WithLabelValues(partition, "overall").Set(efficiency.OverallEfficiency)
		c.queueEfficiencyScore.WithLabelValues(partition, "scheduling").Set(efficiency.SchedulingEfficiency)
		c.queueEfficiencyScore.WithLabelValues(partition, "resource").Set(efficiency.ResourceEfficiency)
		c.queueEfficiencyScore.WithLabelValues(partition, "throughput").Set(efficiency.ThroughputEfficiency)
		c.queueEfficiencyScore.WithLabelValues(partition, "wait_time").Set(efficiency.WaitTimeEfficiency)

		c.queueBottlenecks.WithLabelValues(partition, efficiency.BottleneckType, "detected").Set(efficiency.BottleneckSeverity)
		c.queueEfficiencyScore.WithLabelValues(partition, "optimization_potential").Set(efficiency.OptimizationPotential)
	}
}

func (c *QueueAnalysisCollector) collectJobQueuePosition(ctx context.Context, jobID string) {
	position, err := c.client.GetQueuePositionAnalysis(ctx, jobID)
	if err != nil {
		log.Printf("Error collecting queue position for job %s: %v", jobID, err)
		return
	}

	labels := []string{position.JobID, position.UserName, position.AccountName, position.PartitionName, position.PriorityClass}
	c.queuePosition.WithLabelValues(labels...).Set(float64(position.CurrentPosition))

	movementLabels := []string{position.JobID, position.UserName, position.AccountName, position.PartitionName}
	c.queueMovement.WithLabelValues(append(movementLabels, "velocity")...).Set(position.MovementVelocity)
	c.queueMovement.WithLabelValues(append(movementLabels, "acceleration")...).Set(position.MovementAcceleration)
	c.queueAdvancementRate.WithLabelValues(movementLabels...).Set(position.AdvancementRate)
	c.queueStagnation.WithLabelValues(movementLabels...).Set(position.StagnationTime.Minutes())
}

func (c *QueueAnalysisCollector) collectWaitTimePrediction(ctx context.Context, jobID string) {
	prediction, err := c.client.PredictWaitTime(ctx, jobID)
	if err != nil {
		log.Printf("Error predicting wait time for job %s: %v", jobID, err)
		return
	}

	labels := []string{prediction.JobID, prediction.UserName, prediction.PartitionName, prediction.PredictionMethod}
	c.predictedWaitTime.WithLabelValues(labels...).Set(prediction.PredictedWaitTime.Seconds())

	confidenceLabels := []string{prediction.JobID, prediction.UserName, prediction.PartitionName, prediction.PredictionMethod}
	c.waitTimeConfidence.WithLabelValues(confidenceLabels...).Set(prediction.ConfidenceLevel)

	varianceLabels := []string{prediction.JobID, prediction.UserName, prediction.PartitionName}
	c.waitTimeVariance.WithLabelValues(append(varianceLabels, "min")...).Set(prediction.MinWaitTime.Seconds())
	c.waitTimeVariance.WithLabelValues(append(varianceLabels, "max")...).Set(prediction.MaxWaitTime.Seconds())
	c.waitTimeVariance.WithLabelValues(append(varianceLabels, "median")...).Set(prediction.MedianWaitTime.Seconds())
	c.waitTimeVariance.WithLabelValues(append(varianceLabels, "p90")...).Set(prediction.P90WaitTime.Seconds())
	c.waitTimeVariance.WithLabelValues(append(varianceLabels, "uncertainty")...).Set(prediction.UncertaintyRange.Seconds())

	c.waitTimeModelPerformance.WithLabelValues(prediction.ModelVersion, "accuracy").Set(prediction.ModelAccuracy)
	c.waitTimeModelPerformance.WithLabelValues(prediction.ModelVersion, "training_size").Set(float64(prediction.TrainingDataSize))

	for feature, weight := range prediction.FeatureWeights {
		c.modelFeatureImportance.WithLabelValues("wait_time", prediction.ModelVersion, feature).Set(weight)
	}
}

func (c *QueueAnalysisCollector) collectUserQueueExperience(ctx context.Context, userName string) {
	experience, err := c.client.GetUserQueueExperience(ctx, userName)
	if err != nil {
		log.Printf("Error collecting user queue experience for %s: %v", userName, err)
		return
	}

	labels := []string{experience.UserName, experience.AccountName}

	c.userQueueMetrics.WithLabelValues(append(labels, "total_submissions")...).Set(float64(experience.TotalSubmissions))
	c.userWaitExperience.WithLabelValues(append(labels, "average_wait")...).Set(experience.AverageWaitTime)
	c.userWaitExperience.WithLabelValues(append(labels, "median_wait")...).Set(experience.MedianWaitTime)
	c.userWaitExperience.WithLabelValues(append(labels, "wait_variability")...).Set(experience.WaitTimeVariability)
	c.userWaitExperience.WithLabelValues(append(labels, "experience_score")...).Set(experience.QueueExperienceScore)

	c.userQueueBehavior.WithLabelValues(append(labels, "behavior_optimization")...).Set(experience.BehaviorOptimization)
	c.userQueueMetrics.WithLabelValues(append(labels, "success_rate")...).Set(experience.JobSuccessRate)
	c.userQueueMetrics.WithLabelValues(append(labels, "resource_efficiency")...).Set(experience.ResourceEfficiency)
	c.userQueueMetrics.WithLabelValues(append(labels, "queue_position")...).Set(experience.QueuePosition)
	c.userQueueMetrics.WithLabelValues(append(labels, "priority_utilization")...).Set(experience.PriorityUtilization)

	c.userQueueOptimization.WithLabelValues(append(labels, "wait_time_reduction")...).Set(experience.WaitTimeReduction)
}

func (c *QueueAnalysisCollector) collectHistoricalWaitTimes(ctx context.Context) {
	filters := &WaitTimeFilters{
		StartTime: time.Now().Add(-24 * time.Hour),
		EndTime:   time.Now(),
	}

	historical, err := c.client.GetHistoricalWaitTimes(ctx, filters)
	if err != nil {
		log.Printf("Error collecting historical wait times: %v", err)
		return
	}

	c.waitTimePercentiles.WithLabelValues("all", "all", "mean").Set(historical.MeanWaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "median").Set(historical.MedianWaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "p90").Set(historical.P90WaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "p95").Set(historical.P95WaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "p99").Set(historical.P99WaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "max").Set(historical.MaxWaitTime)
	c.waitTimePercentiles.WithLabelValues("all", "all", "min").Set(historical.MinWaitTime)

	c.waitTimeTrends.WithLabelValues("all", "overall", historical.AnalysisPeriod).Set(1.0)
	c.waitTimeAnomalies.WithLabelValues("all", "rate", "moderate").Set(historical.AnomalyRate)
	c.waitTimeAnomalies.WithLabelValues("all", "threshold", "moderate").Set(historical.AnomalyThreshold)

	for pattern, value := range historical.SeasonalPatterns {
		c.waitTimePatterns.WithLabelValues("all", "seasonal", pattern).Set(value)
	}

	for pattern, value := range historical.DailyPatterns {
		c.waitTimePatterns.WithLabelValues("all", "daily", pattern).Set(value)
	}

	for pattern, value := range historical.WeeklyPatterns {
		c.waitTimePatterns.WithLabelValues("all", "weekly", pattern).Set(value)
	}
}

func (c *QueueAnalysisCollector) collectResourceQueueAnalysis(ctx context.Context) {
	resourceTypes := []string{"cpu", "memory", "gpu", "storage"}
	for _, resourceType := range resourceTypes {
		analysis, err := c.client.GetResourceQueueAnalysis(ctx, resourceType)
		if err != nil {
			continue
		}

		c.resourceQueueDepth.WithLabelValues(resourceType, "all").Set(float64(analysis.JobsWaitingForResource))
		c.resourceWaitTime.WithLabelValues(resourceType, "all", "average").Set(analysis.AverageResourceWaitTime)
		c.resourceWaitTime.WithLabelValues(resourceType, "all", "expected").Set(analysis.ExpectedWaitTime)

		c.resourceAvailability.WithLabelValues(resourceType, "all", "total").Set(analysis.TotalCapacity)
		c.resourceAvailability.WithLabelValues(resourceType, "all", "available").Set(analysis.AvailableCapacity)
		c.resourceAvailability.WithLabelValues(resourceType, "all", "utilized").Set(analysis.UtilizedCapacity)
		c.resourceAvailability.WithLabelValues(resourceType, "all", "reserved").Set(analysis.ReservedCapacity)
		c.resourceAvailability.WithLabelValues(resourceType, "all", "utilization").Set(analysis.CapacityUtilization)

		c.resourceContention.WithLabelValues(resourceType, "all").Set(analysis.ResourceContention)
		c.resourceAllocationRate.WithLabelValues(resourceType, "all").Set(analysis.AllocationRate)
		c.resourceFragmentation.WithLabelValues(resourceType, "fragmentation").Set(analysis.ResourceFragmentation)
	}
}

func (c *QueueAnalysisCollector) collectPriorityQueueAnalysis(ctx context.Context) {
	analysis, err := c.client.GetPriorityQueueAnalysis(ctx)
	if err != nil {
		log.Printf("Error collecting priority queue analysis: %v", err)
		return
	}

	c.priorityAdvantage.WithLabelValues("system", "fairness_score").Set(analysis.FairnessScore)
	c.priorityAdvantage.WithLabelValues("system", "priority_advantage").Set(analysis.PriorityAdvantage)
	c.priorityStarvation.WithLabelValues("system", "all").Set(analysis.StarvationRisk)
	c.priorityInversion.WithLabelValues("all", "priority_inversion").Set(analysis.PriorityInversion)

	c.priorityQueueMetrics.WithLabelValues("system", "mobility").Set(analysis.PriorityMobility)
	c.priorityQueueMetrics.WithLabelValues("system", "aging_effectiveness").Set(analysis.AgingEffectiveness)
	c.priorityQueueMetrics.WithLabelValues("system", "preemption_rate").Set(analysis.PreemptionRate)
	c.priorityQueueMetrics.WithLabelValues("system", "escalation_rate").Set(analysis.EscalationRate)

	c.priorityQueueMetrics.WithLabelValues("system", "health").Set(analysis.PrioritySystemHealth)
	c.priorityQueueMetrics.WithLabelValues("system", "policy_effectiveness").Set(analysis.PolicyEffectiveness)

	for level, metrics := range analysis.PriorityLevels {
		c.priorityQueueMetrics.WithLabelValues(level, "job_count").Set(float64(metrics.JobCount))
		c.priorityWaitTime.WithLabelValues(level, "average").Set(metrics.AverageWaitTime)
		c.priorityWaitTime.WithLabelValues(level, "median").Set(metrics.MedianWaitTime)
		c.priorityQueueMetrics.WithLabelValues(level, "throughput_rate").Set(metrics.ThroughputRate)
		c.priorityQueueMetrics.WithLabelValues(level, "advancement_rate").Set(metrics.AdvancementRate)
		c.priorityStarvation.WithLabelValues(level, "all").Set(metrics.StarvationRisk)
	}
}

func (c *QueueAnalysisCollector) collectBackfillAnalysis(ctx context.Context) {
	partitions := c.getSamplePartitions()
	for _, partition := range partitions {
		analysis, err := c.client.GetBackfillAnalysis(ctx, partition)
		if err != nil {
			continue
		}

		c.backfillOpportunities.WithLabelValues(partition, "total").Set(float64(analysis.TotalOpportunities))
		c.backfillOpportunities.WithLabelValues(partition, "utilized").Set(float64(analysis.UtilizedOpportunities))
		c.backfillOpportunities.WithLabelValues(partition, "missed").Set(float64(analysis.MissedOpportunities))

		c.backfillEfficiency.WithLabelValues(partition, "rate").Set(analysis.BackfillRate)
		c.backfillEfficiency.WithLabelValues(partition, "efficiency").Set(analysis.BackfillEfficiency)
		c.backfillEfficiency.WithLabelValues(partition, "success").Set(analysis.BackfillSuccess)

		c.backfillImpact.WithLabelValues(partition, "jobs_backfilled").Set(float64(analysis.JobsBackfilled))
		c.backfillImpact.WithLabelValues(partition, "wait_time_reduction").Set(analysis.WaitTimeReduction)
		c.backfillImpact.WithLabelValues(partition, "utilization_improvement").Set(analysis.UtilizationImprovement)
		c.backfillImpact.WithLabelValues(partition, "throughput_increase").Set(analysis.ThroughputIncrease)

		c.backfillOptimization.WithLabelValues(partition, "optimization_score").Set(analysis.OptimizationScore)
		c.backfillOptimization.WithLabelValues(partition, "improvement_potential").Set(analysis.ImprovementPotential)
	}
}

func (c *QueueAnalysisCollector) collectQueueStateTransitions(ctx context.Context) {
	transitions, err := c.client.GetQueueStateTransitions(ctx, "24h")
	if err != nil {
		log.Printf("Error collecting queue state transitions: %v", err)
		return
	}

	for state, distribution := range transitions.StateDistribution {
		c.queueStateDistribution.WithLabelValues(state, "all").Set(distribution)
	}

	c.queueHealthMetrics.WithLabelValues("all", "healthy_transitions").Set(transitions.HealthyTransitions)
	c.queueHealthMetrics.WithLabelValues("all", "problematic_transitions").Set(transitions.ProblematicTransitions)
	c.queueHealthMetrics.WithLabelValues("all", "transition_efficiency").Set(transitions.TransitionEfficiency)
	c.queueHealthMetrics.WithLabelValues("all", "state_stability").Set(transitions.StateStability)
}

func (c *QueueAnalysisCollector) collectPredictionModelValidation(ctx context.Context) {
	validation, err := c.client.ValidatePredictionModel(ctx)
	if err != nil {
		log.Printf("Error validating prediction model: %v", err)
		return
	}

	c.modelAccuracy.WithLabelValues(validation.ModelName, validation.ModelVersion, "overall").Set(validation.OverallAccuracy)
	c.modelAccuracy.WithLabelValues(validation.ModelName, validation.ModelVersion, "prediction_error").Set(validation.PredictionError)
	c.modelAccuracy.WithLabelValues(validation.ModelName, validation.ModelVersion, "mae").Set(validation.MeanAbsoluteError)
	c.modelAccuracy.WithLabelValues(validation.ModelName, validation.ModelVersion, "rmse").Set(validation.RootMeanSquareError)
	c.modelAccuracy.WithLabelValues(validation.ModelName, validation.ModelVersion, "r2").Set(validation.R2Score)

	c.modelCalibration.WithLabelValues(validation.ModelName, validation.ModelVersion, "calibration").Set(validation.CalibrationScore)
	c.modelCalibration.WithLabelValues(validation.ModelName, validation.ModelVersion, "confidence").Set(validation.ConfidenceAccuracy)
	c.modelCalibration.WithLabelValues(validation.ModelName, validation.ModelVersion, "uncertainty").Set(validation.UncertaintyCapture)

	for feature, importance := range validation.FeatureImportance {
		c.modelFeatureImportance.WithLabelValues(validation.ModelName, validation.ModelVersion, feature).Set(importance)
	}

	c.waitTimeModelPerformance.WithLabelValues(validation.ModelVersion, "prediction_latency").Set(validation.PredictionLatency)
	c.waitTimeModelPerformance.WithLabelValues(validation.ModelVersion, "model_complexity").Set(validation.ModelComplexity)
	c.waitTimeModelPerformance.WithLabelValues(validation.ModelVersion, "feature_stability").Set(validation.FeatureStability)
}

func (c *QueueAnalysisCollector) collectSystemLoadImpact(ctx context.Context) {
	impact, err := c.client.GetSystemLoadImpact(ctx)
	if err != nil {
		log.Printf("Error collecting system load impact: %v", err)
		return
	}

	c.systemLoadImpact.WithLabelValues("cpu", "load").Set(impact.CPULoad)
	c.systemLoadImpact.WithLabelValues("memory", "load").Set(impact.MemoryLoad)
	c.systemLoadImpact.WithLabelValues("network", "load").Set(impact.NetworkLoad)
	c.systemLoadImpact.WithLabelValues("storage", "load").Set(impact.StorageLoad)
	c.systemLoadImpact.WithLabelValues("overall", "load").Set(impact.OverallSystemLoad)

	c.systemLoadImpact.WithLabelValues("queue", "performance_impact").Set(impact.QueuePerformanceImpact)
	c.systemLoadImpact.WithLabelValues("wait_time", "increase").Set(impact.WaitTimeIncrease)
	c.systemLoadImpact.WithLabelValues("throughput", "decrease").Set(impact.ThroughputDecrease)
	c.systemLoadImpact.WithLabelValues("efficiency", "reduction").Set(impact.EfficiencyReduction)

	c.loadBalancingEffectiveness.WithLabelValues("load_balance", "score").Set(impact.LoadBalanceScore)
	c.loadBalancingEffectiveness.WithLabelValues("load_balance", "effectiveness").Set(impact.LoadBalanceEffectiveness)

	c.capacityUtilization.WithLabelValues("capacity", "utilization").Set(impact.CapacityUtilization)
	c.capacityUtilization.WithLabelValues("capacity", "fragmentation").Set(impact.CapacityFragmentation)
	c.capacityUtilization.WithLabelValues("capacity", "optimization").Set(impact.CapacityOptimization)

	for resource, distribution := range impact.LoadDistribution {
		c.loadBalancingEffectiveness.WithLabelValues(resource, "distribution").Set(distribution)
	}
}

// Simplified mock data generators for testing purposes
func (c *QueueAnalysisCollector) getSamplePartitions() []string {
	return []string{"cpu", "gpu", "bigmem"}
}

func (c *QueueAnalysisCollector) getSampleJobIDs() []string {
	return []string{"12345", "12346", "12347"}
}

func (c *QueueAnalysisCollector) getSampleUsers() []string {
	return []string{"user1", "user2", "user3"}
}
