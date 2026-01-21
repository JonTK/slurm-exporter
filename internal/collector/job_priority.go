package collector

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// JobPriorityCollector collects job priority prediction and analysis metrics
type JobPriorityCollector struct {
	client JobPrioritySLURMClient

	// Job Priority Prediction Metrics
	jobPriorityScore     *prometheus.GaugeVec
	jobPriorityRank      *prometheus.GaugeVec
	jobPriorityAgeScore  *prometheus.GaugeVec
	jobPriorityFairShare *prometheus.GaugeVec
	jobPriorityQoSScore  *prometheus.GaugeVec
	jobPriorityPartition *prometheus.GaugeVec
	jobPrioritySize      *prometheus.GaugeVec
	jobPriorityAssoc     *prometheus.GaugeVec
	jobPriorityNice      *prometheus.GaugeVec

	// Priority Prediction Metrics
	jobEstimatedWaitTime  *prometheus.GaugeVec
	jobQueuePosition      *prometheus.GaugeVec
	jobPriorityTrend      *prometheus.GaugeVec
	jobPriorityChanges    *prometheus.CounterVec
	jobPriorityVolatility *prometheus.GaugeVec

	// System Priority Metrics
	systemPriorityStats     *prometheus.GaugeVec
	priorityDistribution    *prometheus.HistogramVec
	priorityRebalanceEvents *prometheus.CounterVec
	priorityAlgorithmStats  *prometheus.GaugeVec

	// Queue Analysis Metrics
	queueDepth           *prometheus.GaugeVec
	queueProcessingRate  *prometheus.GaugeVec
	queueStarvationRisk  *prometheus.GaugeVec
	queueEfficiencyScore *prometheus.GaugeVec

	// User Priority Behavior Metrics
	userPriorityPattern   *prometheus.GaugeVec
	userPriorityAverage   *prometheus.GaugeVec
	userPriorityVariance  *prometheus.GaugeVec
	userSubmissionPattern *prometheus.GaugeVec

	// Prediction Accuracy Metrics
	predictionAccuracy   *prometheus.GaugeVec
	predictionError      *prometheus.HistogramVec
	predictionConfidence *prometheus.GaugeVec
	predictionModel      *prometheus.GaugeVec
}

// JobPrioritySLURMClient interface for job priority operations
type JobPrioritySLURMClient interface {
	CalculateJobPriority(ctx context.Context, jobID string) (*DetailedJobPriorityFactors, error)
	PredictJobScheduling(ctx context.Context, jobID string) (*JobSchedulingPrediction, error)
	GetQueueAnalysis(ctx context.Context, partition string) (*QueueAnalysis, error)
	GetSystemPriorityStats(ctx context.Context) (*SystemPriorityStats, error)
	GetUserPriorityPattern(ctx context.Context, userName string) (*UserPriorityPattern, error)
	ValidatePriorityPrediction(ctx context.Context, jobID string) (*PriorityPredictionValidation, error)
}

// DetailedJobPriorityFactors represents detailed priority breakdown
type DetailedJobPriorityFactors struct {
	JobID         string `json:"job_id"`
	UserName      string `json:"user_name"`
	AccountName   string `json:"account_name"`
	PartitionName string `json:"partition_name"`
	QoSName       string `json:"qos_name"`

	// Priority Components
	TotalPriority   uint64  `json:"total_priority"`
	AgePriority     uint64  `json:"age_priority"`
	FairShareFactor float64 `json:"fair_share_factor"`
	QoSPriority     uint64  `json:"qos_priority"`
	PartitionPrio   uint64  `json:"partition_priority"`
	SizePriority    uint64  `json:"size_priority"`
	AssocPriority   uint64  `json:"assoc_priority"`
	NicePriority    int32   `json:"nice_priority"`

	// Priority Weighting
	AgeWeight       float64 `json:"age_weight"`
	FairShareWeight float64 `json:"fair_share_weight"`
	QoSWeight       float64 `json:"qos_weight"`
	PartitionWeight float64 `json:"partition_weight"`
	SizeWeight      float64 `json:"size_weight"`

	// Ranking Information
	QueueRank     int `json:"queue_rank"`
	PartitionRank int `json:"partition_rank"`
	UserRank      int `json:"user_rank"`

	// Temporal Data
	SubmittedAt    time.Time `json:"submitted_at"`
	LastCalculated time.Time `json:"last_calculated"`
	NextUpdate     time.Time `json:"next_update"`
}

// JobSchedulingPrediction represents scheduling predictions
type JobSchedulingPrediction struct {
	JobID              string        `json:"job_id"`
	EstimatedWaitTime  time.Duration `json:"estimated_wait_time"`
	EstimatedStartTime time.Time     `json:"estimated_start_time"`
	ConfidenceLevel    float64       `json:"confidence_level"`
	PredictionMethod   string        `json:"prediction_method"`

	// Queue Analysis
	QueuePosition  int     `json:"queue_position"`
	QueueDepth     int     `json:"queue_depth"`
	JobsAhead      int     `json:"jobs_ahead"`
	ProcessingRate float64 `json:"processing_rate"`

	// Priority Trends
	PriorityTrend      string  `json:"priority_trend"`
	PriorityVelocity   float64 `json:"priority_velocity"`
	PriorityVolatility float64 `json:"priority_volatility"`

	// Resource Availability
	ResourceAvailability  float64 `json:"resource_availability"`
	NodeAvailabilityScore float64 `json:"node_availability_score"`
	PartitionUtilization  float64 `json:"partition_utilization"`

	// Risk Factors
	StarvationRisk      float64 `json:"starvation_risk"`
	PreemptionRisk      float64 `json:"preemption_risk"`
	BackfillProbability float64 `json:"backfill_probability"`

	LastUpdated time.Time `json:"last_updated"`
}

// QueueAnalysis represents queue state analysis
type QueueAnalysis struct {
	PartitionName   string  `json:"partition_name"`
	TotalJobs       int     `json:"total_jobs"`
	PendingJobs     int     `json:"pending_jobs"`
	RunningJobs     int     `json:"running_jobs"`
	ProcessingRate  float64 `json:"processing_rate"`
	AverageWaitTime float64 `json:"average_wait_time"`
	EfficiencyScore float64 `json:"efficiency_score"`

	// Priority Distribution
	HighPriorityJobs   int `json:"high_priority_jobs"`
	MediumPriorityJobs int `json:"medium_priority_jobs"`
	LowPriorityJobs    int `json:"low_priority_jobs"`

	// Starvation Analysis
	StarvationRisk      float64 `json:"starvation_risk"`
	OldestJobAge        float64 `json:"oldest_job_age"`
	StarvationThreshold float64 `json:"starvation_threshold"`

	// Trends
	ThroughputTrend  string `json:"throughput_trend"`
	LatencyTrend     string `json:"latency_trend"`
	UtilizationTrend string `json:"utilization_trend"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// SystemPriorityStats represents system-wide priority statistics
type SystemPriorityStats struct {
	TotalJobs       int     `json:"total_jobs"`
	AveragePriority float64 `json:"average_priority"`
	PriorityStdDev  float64 `json:"priority_std_dev"`
	PriorityRange   uint64  `json:"priority_range"`
	PriorityMedian  float64 `json:"priority_median"`

	// Algorithm Performance
	AlgorithmVersion    string  `json:"algorithm_version"`
	RebalanceFrequency  float64 `json:"rebalance_frequency"`
	CalculationTime     float64 `json:"calculation_time"`
	CalculationAccuracy float64 `json:"calculation_accuracy"`

	// Factor Weights
	AgeWeightGlobal       float64 `json:"age_weight_global"`
	FairShareWeightGlobal float64 `json:"fair_share_weight_global"`
	QoSWeightGlobal       float64 `json:"qos_weight_global"`
	PartitionWeightGlobal float64 `json:"partition_weight_global"`

	// Trends
	PriorityInflation  float64 `json:"priority_inflation"`
	PriorityDeflation  float64 `json:"priority_deflation"`
	PriorityVolatility float64 `json:"priority_volatility"`

	LastUpdated time.Time `json:"last_updated"`
}

// UserPriorityPattern represents user priority behavior patterns
type UserPriorityPattern struct {
	UserName          string  `json:"user_name"`
	AccountName       string  `json:"account_name"`
	AveragePriority   float64 `json:"average_priority"`
	PriorityVariance  float64 `json:"priority_variance"`
	SubmissionPattern string  `json:"submission_pattern"`

	// Behavior Analysis
	HighPriorityRatio   float64 `json:"high_priority_ratio"`
	LowPriorityRatio    float64 `json:"low_priority_ratio"`
	PriorityConsistency float64 `json:"priority_consistency"`
	BehaviorScore       float64 `json:"behavior_score"`

	// Temporal Patterns
	PeakSubmissionHours []int   `json:"peak_submission_hours"`
	SubmissionFrequency float64 `json:"submission_frequency"`
	BatchingBehavior    float64 `json:"batching_behavior"`

	// Efficiency Metrics
	WaitTimeEfficiency   float64 `json:"wait_time_efficiency"`
	ResourceEfficiency   float64 `json:"resource_efficiency"`
	SchedulingEfficiency float64 `json:"scheduling_efficiency"`

	LastAnalyzed time.Time `json:"last_analyzed"`
}

// PriorityPredictionValidation represents prediction accuracy validation
type PriorityPredictionValidation struct {
	JobID              string  `json:"job_id"`
	PredictedWaitTime  float64 `json:"predicted_wait_time"`
	ActualWaitTime     float64 `json:"actual_wait_time"`
	PredictionError    float64 `json:"prediction_error"`
	PredictionAccuracy float64 `json:"prediction_accuracy"`
	ConfidenceLevel    float64 `json:"confidence_level"`

	// Error Analysis
	AbsoluteError float64 `json:"absolute_error"`
	RelativeError float64 `json:"relative_error"`
	ErrorCategory string  `json:"error_category"`
	ErrorSeverity string  `json:"error_severity"`

	// Model Performance
	ModelVersion    string  `json:"model_version"`
	ModelAccuracy   float64 `json:"model_accuracy"`
	ModelConfidence float64 `json:"model_confidence"`

	ValidatedAt time.Time `json:"validated_at"`
}

// NewJobPriorityCollector creates a new job priority collector
func NewJobPriorityCollector(client JobPrioritySLURMClient) *JobPriorityCollector {
	return &JobPriorityCollector{
		client: client,

		// Job Priority Prediction Metrics
		jobPriorityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_score",
				Help: "Total priority score for jobs",
			},
			[]string{"job_id", "user", "account", "partition", "qos"},
		),

		jobPriorityRank: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_rank",
				Help: "Queue rank based on priority",
			},
			[]string{"job_id", "user", "account", "partition", "rank_type"},
		),

		jobPriorityAgeScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_age_score",
				Help: "Age component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPriorityFairShare: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_fairshare_factor",
				Help: "Fair-share factor component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPriorityQoSScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_qos_score",
				Help: "QoS component of job priority",
			},
			[]string{"job_id", "user", "account", "partition", "qos"},
		),

		jobPriorityPartition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_partition_score",
				Help: "Partition component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPrioritySize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_size_score",
				Help: "Size component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPriorityAssoc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_association_score",
				Help: "Association component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPriorityNice: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_nice_score",
				Help: "Nice value component of job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		// Priority Prediction Metrics
		jobEstimatedWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_estimated_wait_time_seconds",
				Help: "Estimated wait time for job scheduling in seconds",
			},
			[]string{"job_id", "user", "account", "partition", "method"},
		),

		jobQueuePosition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_queue_position",
				Help: "Position of job in scheduling queue",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		jobPriorityTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_trend",
				Help: "Priority trend direction and velocity",
			},
			[]string{"job_id", "user", "account", "partition", "trend"},
		),

		jobPriorityChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_priority_changes_total",
				Help: "Total number of priority changes for jobs",
			},
			[]string{"job_id", "user", "account", "partition", "change_type"},
		),

		jobPriorityVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_volatility",
				Help: "Priority volatility score",
			},
			[]string{"job_id", "user", "account", "partition"},
		),

		// System Priority Metrics
		systemPriorityStats: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_priority_statistics",
				Help: "System-wide priority statistics",
			},
			[]string{"statistic", "algorithm_version"},
		),

		priorityDistribution: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_priority_distribution",
				Help:    "Distribution of job priorities",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
			},
			[]string{"partition", "qos"},
		),

		priorityRebalanceEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_priority_rebalance_events_total",
				Help: "Total number of priority rebalance events",
			},
			[]string{"event_type", "partition"},
		),

		priorityAlgorithmStats: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_priority_algorithm_performance",
				Help: "Priority algorithm performance metrics",
			},
			[]string{"metric", "algorithm_version"},
		),

		// Queue Analysis Metrics
		queueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_depth",
				Help: "Number of jobs in queue by partition",
			},
			[]string{"partition", "job_state"},
		),

		queueProcessingRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_processing_rate",
				Help: "Queue processing rate in jobs per hour",
			},
			[]string{"partition"},
		),

		queueStarvationRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_starvation_risk",
				Help: "Risk of job starvation in queue",
			},
			[]string{"partition", "risk_level"},
		),

		queueEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_efficiency_score",
				Help: "Queue efficiency score",
			},
			[]string{"partition"},
		),

		// User Priority Behavior Metrics
		userPriorityPattern: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_priority_pattern",
				Help: "User priority behavior patterns",
			},
			[]string{"user", "account", "pattern_type"},
		),

		userPriorityAverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_average_priority",
				Help: "Average priority for user submissions",
			},
			[]string{"user", "account"},
		),

		userPriorityVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_priority_variance",
				Help: "Priority variance for user submissions",
			},
			[]string{"user", "account"},
		),

		userSubmissionPattern: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_submission_pattern_score",
				Help: "User job submission pattern analysis",
			},
			[]string{"user", "account", "pattern"},
		),

		// Prediction Accuracy Metrics
		predictionAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_accuracy",
				Help: "Accuracy of scheduling predictions",
			},
			[]string{"prediction_type", "model_version"},
		),

		predictionError: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_prediction_error",
				Help:    "Distribution of prediction errors",
				Buckets: []float64{0.1, 0.2, 0.5, 1.0, 2.0, 5.0, 10.0},
			},
			[]string{"prediction_type", "error_type"},
		),

		predictionConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_confidence",
				Help: "Confidence level of predictions",
			},
			[]string{"prediction_type", "model_version"},
		),

		predictionModel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_prediction_model_performance",
				Help: "Prediction model performance metrics",
			},
			[]string{"model_version", "metric"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *JobPriorityCollector) Describe(ch chan<- *prometheus.Desc) {
	c.jobPriorityScore.Describe(ch)
	c.jobPriorityRank.Describe(ch)
	c.jobPriorityAgeScore.Describe(ch)
	c.jobPriorityFairShare.Describe(ch)
	c.jobPriorityQoSScore.Describe(ch)
	c.jobPriorityPartition.Describe(ch)
	c.jobPrioritySize.Describe(ch)
	c.jobPriorityAssoc.Describe(ch)
	c.jobPriorityNice.Describe(ch)
	c.jobEstimatedWaitTime.Describe(ch)
	c.jobQueuePosition.Describe(ch)
	c.jobPriorityTrend.Describe(ch)
	c.jobPriorityChanges.Describe(ch)
	c.jobPriorityVolatility.Describe(ch)
	c.systemPriorityStats.Describe(ch)
	c.priorityDistribution.Describe(ch)
	c.priorityRebalanceEvents.Describe(ch)
	c.priorityAlgorithmStats.Describe(ch)
	c.queueDepth.Describe(ch)
	c.queueProcessingRate.Describe(ch)
	c.queueStarvationRisk.Describe(ch)
	c.queueEfficiencyScore.Describe(ch)
	c.userPriorityPattern.Describe(ch)
	c.userPriorityAverage.Describe(ch)
	c.userPriorityVariance.Describe(ch)
	c.userSubmissionPattern.Describe(ch)
	c.predictionAccuracy.Describe(ch)
	c.predictionError.Describe(ch)
	c.predictionConfidence.Describe(ch)
	c.predictionModel.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *JobPriorityCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Reset metrics
	c.resetMetrics()

	// Collect system priority statistics
	c.collectSystemPriorityStats(ctx)

	// Get sample job IDs for priority analysis
	sampleJobIDs := c.getSampleJobIDs()
	for _, jobID := range sampleJobIDs {
		c.collectJobPriorityMetrics(ctx, jobID)
	}

	// Get sample partitions for queue analysis
	samplePartitions := c.getSamplePartitions()
	for _, partition := range samplePartitions {
		c.collectQueueAnalysis(ctx, partition)
	}

	// Get sample users for priority pattern analysis
	sampleUsers := c.getSampleUsers()
	for _, user := range sampleUsers {
		c.collectUserPriorityPatterns(ctx, user)
	}

	// Collect prediction validation metrics
	c.collectPredictionValidation(ctx)

	// Collect all metrics
	c.jobPriorityScore.Collect(ch)
	c.jobPriorityRank.Collect(ch)
	c.jobPriorityAgeScore.Collect(ch)
	c.jobPriorityFairShare.Collect(ch)
	c.jobPriorityQoSScore.Collect(ch)
	c.jobPriorityPartition.Collect(ch)
	c.jobPrioritySize.Collect(ch)
	c.jobPriorityAssoc.Collect(ch)
	c.jobPriorityNice.Collect(ch)
	c.jobEstimatedWaitTime.Collect(ch)
	c.jobQueuePosition.Collect(ch)
	c.jobPriorityTrend.Collect(ch)
	c.jobPriorityChanges.Collect(ch)
	c.jobPriorityVolatility.Collect(ch)
	c.systemPriorityStats.Collect(ch)
	c.priorityDistribution.Collect(ch)
	c.priorityRebalanceEvents.Collect(ch)
	c.priorityAlgorithmStats.Collect(ch)
	c.queueDepth.Collect(ch)
	c.queueProcessingRate.Collect(ch)
	c.queueStarvationRisk.Collect(ch)
	c.queueEfficiencyScore.Collect(ch)
	c.userPriorityPattern.Collect(ch)
	c.userPriorityAverage.Collect(ch)
	c.userPriorityVariance.Collect(ch)
	c.userSubmissionPattern.Collect(ch)
	c.predictionAccuracy.Collect(ch)
	c.predictionError.Collect(ch)
	c.predictionConfidence.Collect(ch)
	c.predictionModel.Collect(ch)
}

func (c *JobPriorityCollector) resetMetrics() {
	c.jobPriorityScore.Reset()
	c.jobPriorityRank.Reset()
	c.jobPriorityAgeScore.Reset()
	c.jobPriorityFairShare.Reset()
	c.jobPriorityQoSScore.Reset()
	c.jobPriorityPartition.Reset()
	c.jobPrioritySize.Reset()
	c.jobPriorityAssoc.Reset()
	c.jobPriorityNice.Reset()
	c.jobEstimatedWaitTime.Reset()
	c.jobQueuePosition.Reset()
	c.jobPriorityTrend.Reset()
	c.jobPriorityVolatility.Reset()
	c.systemPriorityStats.Reset()
	c.queueDepth.Reset()
	c.queueProcessingRate.Reset()
	c.queueStarvationRisk.Reset()
	c.queueEfficiencyScore.Reset()
	c.userPriorityPattern.Reset()
	c.userPriorityAverage.Reset()
	c.userPriorityVariance.Reset()
	c.userSubmissionPattern.Reset()
	c.predictionAccuracy.Reset()
	c.predictionConfidence.Reset()
	c.predictionModel.Reset()
}

func (c *JobPriorityCollector) collectSystemPriorityStats(ctx context.Context) {
	stats, err := c.client.GetSystemPriorityStats(ctx)
	if err != nil {
		log.Printf("Error collecting system priority stats: %v", err)
		return
	}

	c.systemPriorityStats.WithLabelValues("total_jobs", stats.AlgorithmVersion).Set(float64(stats.TotalJobs))
	c.systemPriorityStats.WithLabelValues("average_priority", stats.AlgorithmVersion).Set(stats.AveragePriority)
	c.systemPriorityStats.WithLabelValues("priority_std_dev", stats.AlgorithmVersion).Set(stats.PriorityStdDev)
	c.systemPriorityStats.WithLabelValues("priority_range", stats.AlgorithmVersion).Set(float64(stats.PriorityRange))
	c.systemPriorityStats.WithLabelValues("priority_median", stats.AlgorithmVersion).Set(stats.PriorityMedian)
	c.systemPriorityStats.WithLabelValues("priority_inflation", stats.AlgorithmVersion).Set(stats.PriorityInflation)
	c.systemPriorityStats.WithLabelValues("priority_volatility", stats.AlgorithmVersion).Set(stats.PriorityVolatility)

	c.priorityAlgorithmStats.WithLabelValues("rebalance_frequency", stats.AlgorithmVersion).Set(stats.RebalanceFrequency)
	c.priorityAlgorithmStats.WithLabelValues("calculation_time", stats.AlgorithmVersion).Set(stats.CalculationTime)
	c.priorityAlgorithmStats.WithLabelValues("calculation_accuracy", stats.AlgorithmVersion).Set(stats.CalculationAccuracy)
}

func (c *JobPriorityCollector) collectJobPriorityMetrics(ctx context.Context, jobID string) {
	factors, err := c.client.CalculateJobPriority(ctx, jobID)
	if err != nil {
		log.Printf("Error calculating job priority for %s: %v", jobID, err)
		return
	}

	labels := []string{factors.JobID, factors.UserName, factors.AccountName, factors.PartitionName, factors.QoSName}
	baseLabels := []string{factors.JobID, factors.UserName, factors.AccountName, factors.PartitionName}

	c.jobPriorityScore.WithLabelValues(labels...).Set(float64(factors.TotalPriority))
	c.jobPriorityAgeScore.WithLabelValues(baseLabels...).Set(float64(factors.AgePriority))
	c.jobPriorityFairShare.WithLabelValues(baseLabels...).Set(factors.FairShareFactor)
	c.jobPriorityQoSScore.WithLabelValues(labels...).Set(float64(factors.QoSPriority))
	c.jobPriorityPartition.WithLabelValues(baseLabels...).Set(float64(factors.PartitionPrio))
	c.jobPrioritySize.WithLabelValues(baseLabels...).Set(float64(factors.SizePriority))
	c.jobPriorityAssoc.WithLabelValues(baseLabels...).Set(float64(factors.AssocPriority))
	c.jobPriorityNice.WithLabelValues(baseLabels...).Set(float64(factors.NicePriority))

	c.jobPriorityRank.WithLabelValues(append(baseLabels, "queue")...).Set(float64(factors.QueueRank))
	c.jobPriorityRank.WithLabelValues(append(baseLabels, "partition")...).Set(float64(factors.PartitionRank))
	c.jobPriorityRank.WithLabelValues(append(baseLabels, "user")...).Set(float64(factors.UserRank))

	// Collect scheduling prediction
	prediction, err := c.client.PredictJobScheduling(ctx, jobID)
	if err != nil {
		log.Printf("Error predicting job scheduling for %s: %v", jobID, err)
		return
	}

	predictionLabels := append(baseLabels, prediction.PredictionMethod)
	c.jobEstimatedWaitTime.WithLabelValues(predictionLabels...).Set(prediction.EstimatedWaitTime.Seconds())
	c.jobQueuePosition.WithLabelValues(baseLabels...).Set(float64(prediction.QueuePosition))
	c.jobPriorityTrend.WithLabelValues(append(baseLabels, prediction.PriorityTrend)...).Set(prediction.PriorityVelocity)
	c.jobPriorityVolatility.WithLabelValues(baseLabels...).Set(prediction.PriorityVolatility)

	// Add to priority distribution
	c.priorityDistribution.WithLabelValues(factors.PartitionName, factors.QoSName).Observe(float64(factors.TotalPriority))
}

func (c *JobPriorityCollector) collectQueueAnalysis(ctx context.Context, partition string) {
	analysis, err := c.client.GetQueueAnalysis(ctx, partition)
	if err != nil {
		log.Printf("Error analyzing queue for partition %s: %v", partition, err)
		return
	}

	c.queueDepth.WithLabelValues(partition, "total").Set(float64(analysis.TotalJobs))
	c.queueDepth.WithLabelValues(partition, "pending").Set(float64(analysis.PendingJobs))
	c.queueDepth.WithLabelValues(partition, "running").Set(float64(analysis.RunningJobs))
	c.queueDepth.WithLabelValues(partition, "high_priority").Set(float64(analysis.HighPriorityJobs))
	c.queueDepth.WithLabelValues(partition, "medium_priority").Set(float64(analysis.MediumPriorityJobs))
	c.queueDepth.WithLabelValues(partition, "low_priority").Set(float64(analysis.LowPriorityJobs))

	c.queueProcessingRate.WithLabelValues(partition).Set(analysis.ProcessingRate)
	c.queueEfficiencyScore.WithLabelValues(partition).Set(analysis.EfficiencyScore)

	if analysis.StarvationRisk > 0.7 {
		c.queueStarvationRisk.WithLabelValues(partition, "high").Set(analysis.StarvationRisk)
	} else if analysis.StarvationRisk > 0.4 {
		c.queueStarvationRisk.WithLabelValues(partition, "medium").Set(analysis.StarvationRisk)
	} else {
		c.queueStarvationRisk.WithLabelValues(partition, "low").Set(analysis.StarvationRisk)
	}
}

func (c *JobPriorityCollector) collectUserPriorityPatterns(ctx context.Context, userName string) {
	pattern, err := c.client.GetUserPriorityPattern(ctx, userName)
	if err != nil {
		log.Printf("Error getting user priority pattern for %s: %v", userName, err)
		return
	}

	c.userPriorityAverage.WithLabelValues(pattern.UserName, pattern.AccountName).Set(pattern.AveragePriority)
	c.userPriorityVariance.WithLabelValues(pattern.UserName, pattern.AccountName).Set(pattern.PriorityVariance)

	c.userPriorityPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "high_priority_ratio").Set(pattern.HighPriorityRatio)
	c.userPriorityPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "low_priority_ratio").Set(pattern.LowPriorityRatio)
	c.userPriorityPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "consistency").Set(pattern.PriorityConsistency)
	c.userPriorityPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "behavior_score").Set(pattern.BehaviorScore)

	c.userSubmissionPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "frequency").Set(pattern.SubmissionFrequency)
	c.userSubmissionPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "batching").Set(pattern.BatchingBehavior)
	c.userSubmissionPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "wait_efficiency").Set(pattern.WaitTimeEfficiency)
	c.userSubmissionPattern.WithLabelValues(pattern.UserName, pattern.AccountName, "resource_efficiency").Set(pattern.ResourceEfficiency)
}

func (c *JobPriorityCollector) collectPredictionValidation(ctx context.Context) {
	sampleJobIDs := c.getSampleJobIDs()
	for _, jobID := range sampleJobIDs {
		validation, err := c.client.ValidatePriorityPrediction(ctx, jobID)
		if err != nil {
			continue
		}

		c.predictionAccuracy.WithLabelValues("wait_time", validation.ModelVersion).Set(validation.PredictionAccuracy)
		c.predictionConfidence.WithLabelValues("wait_time", validation.ModelVersion).Set(validation.ConfidenceLevel)
		c.predictionModel.WithLabelValues(validation.ModelVersion, "accuracy").Set(validation.ModelAccuracy)
		c.predictionModel.WithLabelValues(validation.ModelVersion, "confidence").Set(validation.ModelConfidence)

		c.predictionError.WithLabelValues("wait_time", "absolute").Observe(validation.AbsoluteError)
		c.predictionError.WithLabelValues("wait_time", "relative").Observe(validation.RelativeError)
	}
}

// Simplified mock data generators for testing purposes
func (c *JobPriorityCollector) getSampleJobIDs() []string {
	return []string{"12345", "12346", "12347"}
}

func (c *JobPriorityCollector) getSamplePartitions() []string {
	return []string{"cpu", "gpu", "bigmem"}
}

func (c *JobPriorityCollector) getSampleUsers() []string {
	return []string{"user1", "user2", "user3"}
}
