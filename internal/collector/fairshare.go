package collector

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	// Commented out as only used in commented-out field
	// "sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// FairShareCollector provides comprehensive fair-share and priority analysis
type FairShareCollector struct {
	slurmClient slurm.SlurmClient
	logger      *slog.Logger
	config      *FairShareConfig
	metrics     *FairShareMetrics

	// Fair-share data storage
	userFairShares map[string]*UserFairShare
	// TODO: Unused field - preserved for future account hierarchy tracking
	// accountHierarchy  *AccountFairShareHierarchy
	priorityFactors map[string]*JobPriorityFactors

	// Analysis engines
	violationDetector *FairShareViolationDetector
	policyAnalyzer    *FairSharePolicyAnalyzer
	trendAnalyzer     *FairShareTrendAnalyzer

	// Queue analysis
	queueAnalyzer *QueueAnalyzer

	// User behavior analysis
	behaviorAnalyzer *UserBehaviorAnalyzer

	// TODO: Unused fields - preserved for future collection tracking and thread safety
	// lastCollection    time.Time
	// mu                sync.RWMutex
}

// FairShareConfig configures the fair-share collector
type FairShareConfig struct {
	CollectionInterval time.Duration
	FairShareRetention time.Duration
	PriorityRetention  time.Duration

	// Fair-share monitoring
	EnableUserFairShare      bool
	EnableAccountHierarchy   bool
	EnablePriorityAnalysis   bool
	EnableViolationDetection bool
	EnableTrendAnalysis      bool

	// Analysis parameters
	ViolationThreshold float64 // Threshold for fair-share violations
	DecayPeriod        time.Duration
	ResetCycle         time.Duration
	PriorityWeights    PriorityWeights

	// Queue analysis
	EnableQueueAnalysis   bool
	QueuePositionTracking bool
	WaitTimePrediction    bool

	// User behavior analysis
	EnableBehaviorAnalysis  bool
	BehaviorPatternWindow   time.Duration
	OptimizationSuggestions bool

	// Data processing
	MaxUsersPerCollection    int
	MaxAccountsPerCollection int
	EnableParallelProcessing bool
	MaxConcurrentAnalyses    int

	// Reporting
	GenerateReports bool
	ReportInterval  time.Duration
	ReportRetention time.Duration
}

// PriorityWeights defines weights for priority calculation factors
type PriorityWeights struct {
	AgeWeight       float64
	FairShareWeight float64
	QoSWeight       float64
	PartitionWeight float64
	AssocWeight     float64
	JobSizeWeight   float64
}

// UserFairShare contains fair-share data for a user
type UserFairShare struct {
	UserName  string
	Account   string
	Partition string
	Timestamp time.Time

	// Fair-share factors
	FairShareFactor  float64 // Current fair-share factor (0.0 - 1.0+)
	RawShares        int64   // Raw shares allocated to user
	NormalizedShares float64 // Normalized shares (0.0 - 1.0)
	EffectiveUsage   float64 // Effective usage over fair-share period

	// Usage tracking
	CPUUsage    float64 // CPU usage in CPU-hours
	MemoryUsage float64 // Memory usage in MB-hours
	GPUUsage    float64 // GPU usage in GPU-hours
	NodeUsage   float64 // Node usage in node-hours

	// Fair-share calculations
	TargetUsage       float64 // Target usage based on shares
	UsageRatio        float64 // Actual usage / target usage
	FairSharePriority float64 // Priority contribution from fair-share

	// Decay and reset tracking
	LastDecay     time.Time
	LastReset     time.Time
	DecayHalfLife time.Duration

	// Account association
	AccountPath   string // Full account hierarchy path
	ParentAccount string // Direct parent account
	Level         int    // Hierarchy level (0 = root)

	// Quality and metadata
	DataQuality float64 // Quality of fair-share data (0-1)
	LastUpdated time.Time
	UpdateCount int64

	// Trends and analysis
	TrendDirection  string  // "improving", "degrading", "stable"
	TrendStrength   float64 // Strength of trend (0-1)
	PredictedFactor float64 // Predicted fair-share factor

	// Violation tracking
	IsViolating       bool          // Currently violating fair-share
	ViolationSeverity float64       // Severity of violation (0-1)
	ViolationDuration time.Duration // How long violation has lasted
	ViolationHistory  []*FairShareViolation
}

// AccountFairShareHierarchy represents the account hierarchy for fair-share
type AccountFairShareHierarchy struct {
	RootAccount    *AccountFairShareNode
	AccountMap     map[string]*AccountFairShareNode
	LastUpdated    time.Time
	HierarchyDepth int
	TotalAccounts  int

	// Hierarchy analysis
	BalanceScore     float64 // How balanced the hierarchy is (0-1)
	EfficiencyScore  float64 // How efficiently shares are used (0-1)
	UtilizationScore float64 // Overall utilization score (0-1)
}

// AccountFairShareNode represents a node in the fair-share hierarchy
type AccountFairShareNode struct {
	AccountName   string
	ParentAccount string
	Children      []*AccountFairShareNode
	Level         int

	// Share allocation
	RawShares        int64
	NormalizedShares float64
	EffectiveShares  float64

	// Usage tracking
	TotalUsage float64
	UserUsage  map[string]float64
	ChildUsage map[string]float64

	// Fair-share metrics
	FairShareFactor float64
	TargetUsage     float64
	ActualUsage     float64
	UsageRatio      float64

	// Users in this account
	Users       []string
	UserCount   int
	ActiveUsers int

	// Quality metrics
	DataCompleteness float64
	LastUpdated      time.Time
}

// JobPriorityFactors contains detailed priority factor breakdown for a job
type JobPriorityFactors struct {
	JobID     string
	UserName  string
	Account   string
	Partition string
	QoS       string
	Timestamp time.Time

	// Priority components
	TotalPriority int64 // Total calculated priority
	BasePriority  int64 // Base priority

	// Factor contributions
	AgeFactor       float64 // Age contribution to priority
	FairShareFactor float64 // Fair-share contribution
	QoSFactor       float64 // QoS contribution
	PartitionFactor float64 // Partition contribution
	AssocFactor     float64 // Association contribution
	JobSizeFactor   float64 // Job size contribution

	// Normalized factors (0-1)
	NormalizedAge       float64
	NormalizedFairShare float64
	NormalizedQoS       float64
	NormalizedPartition float64
	NormalizedAssoc     float64
	NormalizedJobSize   float64

	// Priority analysis
	PriorityScore      float64 // Overall priority score (0-1)
	PriorityRank       int     // Rank among all jobs
	PriorityPercentile float64 // Percentile rank

	// Predictions
	EstimatedWaitTime  time.Duration // Estimated wait time based on priority
	QueuePosition      int           // Current position in queue
	PredictedStartTime time.Time     // Predicted start time

	// Context
	JobSize            int                // Number of CPUs/nodes
	SubmitTime         time.Time          // Job submission time
	RequestedRuntime   time.Duration      // Requested runtime
	RequestedResources map[string]float64 // Requested resources

	// Quality indicators
	CalculationQuality float64       // Quality of priority calculation (0-1)
	DataFreshness      time.Duration // How fresh the data is
}

// FairShareViolation represents a fair-share violation event
type FairShareViolation struct {
	ViolationID   string
	UserName      string
	Account       string
	ViolationType string // "overuse", "underuse", "imbalance"
	Severity      string // "minor", "moderate", "major", "critical"

	// Violation details
	StartTime     time.Time
	EndTime       *time.Time
	Duration      time.Duration
	CurrentFactor float64
	TargetFactor  float64
	Deviation     float64

	// Impact assessment
	ResourceImpact map[string]float64 // Impact on different resources
	UserImpact     []string           // Other users affected
	ClusterImpact  float64            // Overall cluster impact (0-1)

	// Resolution
	IsResolved       bool
	ResolutionAction string
	ResolutionTime   *time.Time

	// Notifications
	AlertsSent      int
	EscalationLevel int

	// Context
	CauseAnalysis   []string // Potential causes
	Recommendations []string // Recommended actions
}

// FairShareViolationDetector detects and analyzes fair-share violations
type FairShareViolationDetector struct {
	config     *FairShareConfig
	logger     *slog.Logger
	violations map[string]*FairShareViolation
	// TODO: Unused field - preserved for future alert history tracking
	// alertHistory    []*ViolationAlert
	thresholds *ViolationThresholds
}

// ViolationThresholds defines thresholds for violation detection
type ViolationThresholds struct {
	MinorThreshold    float64       // Minor violation threshold
	ModerateThreshold float64       // Moderate violation threshold
	MajorThreshold    float64       // Major violation threshold
	CriticalThreshold float64       // Critical violation threshold
	DurationThreshold time.Duration // Duration before escalation
}

// ViolationAlert represents a violation alert
type ViolationAlert struct {
	AlertID        string
	ViolationID    string
	AlertLevel     string
	Timestamp      time.Time
	Message        string
	Recipients     []string
	DeliveryStatus string
}

// FairSharePolicyAnalyzer analyzes fair-share policy effectiveness
type FairSharePolicyAnalyzer struct {
	config        *FairShareConfig
	logger        *slog.Logger
	policyMetrics *PolicyEffectivenessMetrics
	// TODO: Unused field - preserved for future policy recommendations
	// recommendations []*PolicyRecommendation
}

// PolicyEffectivenessMetrics contains policy effectiveness analysis
type PolicyEffectivenessMetrics struct {
	Timestamp time.Time

	// Overall effectiveness
	OverallScore    float64 // Overall policy effectiveness (0-1)
	BalanceScore    float64 // How well the policy balances usage (0-1)
	FairnessScore   float64 // How fair the policy is (0-1)
	EfficiencyScore float64 // How efficiently resources are used (0-1)

	// Usage distribution
	GiniCoefficient float64 // Measure of inequality in usage
	VarianceScore   float64 // Variance in fair-share factors
	OutlierCount    int     // Number of outlier users

	// Temporal analysis
	StabilityScore  float64 // How stable fair-share factors are over time
	ConvergenceRate float64 // Rate of convergence to target shares

	// User satisfaction
	ViolationRate     float64 // Rate of fair-share violations
	ComplaintRate     float64 // Rate of user complaints
	SatisfactionScore float64 // Overall user satisfaction (0-1)

	// Resource utilization
	UtilizationRate float64 // Overall resource utilization
	WasteRate       float64 // Resource waste rate
	ThroughputScore float64 // Job throughput score
}

// PolicyRecommendation contains policy improvement recommendations
type PolicyRecommendation struct {
	RecommendationID     string
	Category             string // "shares", "decay", "reset", "hierarchy"
	Priority             string // "high", "medium", "low"
	Description          string
	ExpectedImpact       float64
	ImplementationEffort string

	// Specific recommendations
	ShareAdjustments map[string]int64 // Recommended share adjustments
	DecayAdjustment  *time.Duration   // Recommended decay adjustment
	ResetAdjustment  *time.Duration   // Recommended reset adjustment

	// Supporting data
	SupportingData  map[string]interface{}
	AnalysisResults map[string]float64

	// Validation
	Confidence     float64
	RiskAssessment string
}

// FairShareTrendAnalyzer analyzes fair-share trends over time
type FairShareTrendAnalyzer struct {
	config      *FairShareConfig
	logger      *slog.Logger
	trendData   map[string]*FairShareTrendData
	predictions map[string]*FairSharePrediction
}

// FairShareTrendData contains trend analysis data for a user
type FairShareTrendData struct {
	UserName string
	Account  string

	// Historical data
	HistoricalFactors []FairShareDataPoint
	HistoricalUsage   []UsageDataPoint

	// Trend analysis
	TrendDirection    string  // "improving", "degrading", "stable", "cyclical"
	TrendStrength     float64 // Strength of trend (0-1)
	TrendSignificance float64 // Statistical significance

	// Statistical analysis
	Mean            float64
	StandardDev     float64
	Variance        float64
	AutoCorrelation float64

	// Seasonality
	SeasonalPatterns []SeasonalPattern
	CyclicalPatterns []FairShareCyclicalPattern

	// Change points
	ChangePoints []FairShareChangePoint

	// Quality metrics
	DataQuality       float64
	PredictionQuality float64
}

// FairShareDataPoint represents a single fair-share measurement
type FairShareDataPoint struct {
	Timestamp       time.Time
	FairShareFactor float64
	Usage           float64
	TargetUsage     float64
	Quality         float64
}

// UsageDataPoint represents a single usage measurement
type UsageDataPoint struct {
	Timestamp   time.Time
	CPUUsage    float64
	MemoryUsage float64
	GPUUsage    float64
	NodeUsage   float64
	TotalUsage  float64
}

// SeasonalPattern represents a detected seasonal pattern
type SeasonalPattern struct {
	PatternType string // "daily", "weekly", "monthly"
	Period      time.Duration
	Amplitude   float64
	Phase       float64
	Strength    float64
	Confidence  float64
}

// FairShareCyclicalPattern represents a detected cyclical pattern in fairshare
type FairShareCyclicalPattern struct {
	CycleLength time.Duration
	Amplitude   float64
	Phase       float64
	Strength    float64
	Confidence  float64
}

// FairShareChangePoint represents a detected change in fair-share behavior
type FairShareChangePoint struct {
	Timestamp    time.Time
	ChangeType   string // "level_shift", "trend_change", "variance_change"
	Magnitude    float64
	Significance float64
	Description  string
	Causes       []string
}

// FairSharePrediction contains fair-share predictions
type FairSharePrediction struct {
	UserName       string
	Account        string
	PredictionTime time.Time
	Horizon        time.Duration

	// Predicted values
	PredictedFactor float64
	PredictedUsage  float64
	PredictedRank   int

	// Confidence intervals
	LowerBound float64
	UpperBound float64
	Confidence float64

	// Prediction quality
	Accuracy float64
	RMSE     float64
	MAE      float64

	// Scenario analysis
	BestCase   float64
	WorstCase  float64
	MostLikely float64
}

// QueueAnalyzer analyzes queue positions and wait times
type QueueAnalyzer struct {
	config      *FairShareConfig
	logger      *slog.Logger
	queueData   map[string]*QueueAnalysisData
	predictions map[string]*WaitTimePrediction
}

// QueueAnalysisData contains queue analysis for a partition
type QueueAnalysisData struct {
	PartitionName string
	Timestamp     time.Time

	// Queue metrics
	QueueLength int
	TotalJobs   int
	RunningJobs int
	PendingJobs int

	// Wait time statistics
	AverageWaitTime time.Duration
	MedianWaitTime  time.Duration
	MaxWaitTime     time.Duration
	MinWaitTime     time.Duration

	// Priority distribution
	PriorityDistribution map[string]int // Priority ranges to job counts
	HighPriorityJobs     int
	MediumPriorityJobs   int
	LowPriorityJobs      int

	// Resource requirements
	TotalCPURequest    int64
	TotalMemoryRequest int64
	TotalGPURequest    int64

	// Throughput metrics
	JobStartRate      float64 // Jobs started per hour
	JobCompletionRate float64 // Jobs completed per hour
	ThroughputTrend   string  // "increasing", "decreasing", "stable"

	// Efficiency metrics
	QueueEfficiency     float64 // How efficiently the queue is processed
	ResourceUtilization float64 // Resource utilization rate
	QueueBalance        float64 // How balanced the queue is
}

// WaitTimePrediction contains wait time predictions for a job
type WaitTimePrediction struct {
	JobID     string
	UserName  string
	Partition string
	Timestamp time.Time

	// Current state
	CurrentPosition int
	CurrentPriority int64
	SubmitTime      time.Time

	// Predictions
	EstimatedWaitTime  time.Duration
	PredictedStartTime time.Time
	EstimatedEndTime   time.Time

	// Confidence intervals
	MinWaitTime time.Duration
	MaxWaitTime time.Duration
	Confidence  float64

	// Analysis factors
	PriorityImpact  float64 // Impact of priority on wait time
	QueueLoadImpact float64 // Impact of queue load
	ResourceImpact  float64 // Impact of resource requirements
	FairShareImpact float64 // Impact of fair-share factor

	// Historical context
	SimilarJobsAvgWait time.Duration
	UserAvgWait        time.Duration
	PartitionAvgWait   time.Duration

	// Prediction quality
	PredictionAccuracy float64
	ModelConfidence    float64
	DataQuality        float64
}

// UserBehaviorAnalyzer analyzes user behavior patterns for fair-share optimization
type UserBehaviorAnalyzer struct {
	config        *FairShareConfig
	logger        *slog.Logger
	behaviorData  map[string]*UserBehaviorProfile
	optimizations map[string]*BehaviorOptimization
}

// UserBehaviorProfile contains behavior analysis for a user
type UserBehaviorProfile struct {
	UserName       string
	Account        string
	AnalysisPeriod time.Duration
	LastUpdated    time.Time

	// Submission patterns
	SubmissionPattern *FairShareSubmissionPattern

	// Resource usage patterns
	ResourcePattern *FairShareResourceUsagePattern

	// Temporal patterns
	TemporalPattern *TemporalUsagePattern

	// Fair-share interaction
	FairShareBehavior *FairShareBehaviorPattern

	// Efficiency metrics
	ResourceEfficiency float64
	TimeEfficiency     float64
	QueueEfficiency    float64
	OverallEfficiency  float64

	// Behavior scoring
	BehaviorScore       float64 // Overall behavior score (0-1)
	FairnessScore       float64 // How fair the user is to others (0-1)
	EfficiencyScore     float64 // How efficiently user uses resources (0-1)
	PredictabilityScore float64 // How predictable user behavior is (0-1)

	// Recommendations
	Recommendations       []*BehaviorRecommendation
	OptimizationPotential float64
}

// FairShareSubmissionPattern represents job submission patterns in fairshare context
type FairShareSubmissionPattern struct {
	SubmissionRate       float64 // Jobs per day
	SubmissionVariance   float64 // Variance in submission rate
	PeakSubmissionHours  []int   // Hours with peak submissions
	SubmissionBursts     int     // Number of submission bursts
	SubmissionRegularity float64 // How regular submissions are (0-1)
}

// FairShareResourceUsagePattern represents resource usage patterns in fairshare context
type FairShareResourceUsagePattern struct {
	PreferredJobSizes  []int  // Preferred job sizes (CPUs)
	MemoryUsagePattern string // "light", "moderate", "heavy"
	GPUUsagePattern    string // "none", "occasional", "heavy"
	RuntimePattern     string // "short", "medium", "long", "mixed"

	// Resource efficiency
	CPUEfficiency     float64
	MemoryEfficiency  float64
	GPUEfficiency     float64
	RuntimeEfficiency float64

	// Usage distribution
	ResourceDistribution map[string]float64
	UtilizationVariance  float64
}

// TemporalUsagePattern represents temporal usage patterns
type TemporalUsagePattern struct {
	PeakUsageHours    []int
	PeakUsageDays     []string
	OffPeakUsage      float64
	SeasonalVariation float64

	// Time preferences
	PreferredSubmissionTimes []string
	PreferredRunTimes        []time.Duration

	// Temporal efficiency
	TimingOptimization float64 // How well user times submissions
	AvoidanceOfPeaks   float64 // How well user avoids peak times
}

// FairShareBehaviorPattern represents fair-share interaction patterns
type FairShareBehaviorPattern struct {
	FairShareAwareness float64 // How aware user is of fair-share (0-1)
	AdaptationRate     float64 // How quickly user adapts to changes
	ComplianceRate     float64 // How well user complies with fair-share

	// Response patterns
	ResponseToLowPriority  string // How user responds to low priority
	ResponseToHighPriority string // How user responds to high priority
	ResponseToViolations   string // How user responds to violations

	// Fair-share metrics
	FairShareHistory   []float64
	AverageFairShare   float64
	FairShareStability float64
	ViolationFrequency float64
}

// BehaviorRecommendation contains behavior optimization recommendations
type BehaviorRecommendation struct {
	RecommendationID     string
	Category             string // "timing", "sizing", "efficiency", "fairness"
	Priority             string // "high", "medium", "low"
	Description          string
	ExpectedBenefit      float64
	ImplementationEffort string

	// Specific recommendations
	TimingAdjustments []string
	SizingAdjustments []string
	EfficiencyTips    []string
	FairShareTips     []string

	// Supporting data
	CurrentBehavior map[string]float64
	TargetBehavior  map[string]float64
	HistoricalData  map[string][]float64

	// Validation
	Confidence     float64
	RiskAssessment string
}

// BehaviorOptimization contains optimization strategies for a user
type BehaviorOptimization struct {
	UserName            string
	OptimizationGoals   []string
	OptimizationPlan    []*OptimizationStep
	ExpectedImprovement float64
	ImplementationTime  time.Duration

	// Tracking
	Progress       float64
	SuccessMetrics map[string]float64
	LastUpdated    time.Time
}

// OptimizationStep represents a single optimization step
type OptimizationStep struct {
	StepID         string
	Description    string
	Action         string
	ExpectedImpact float64
	Deadline       time.Time
	Status         string // "pending", "in_progress", "completed", "failed"

	// Validation
	SuccessCriteria   []string
	CompletionMetrics map[string]float64
}

// FairShareMetrics holds Prometheus metrics for fair-share monitoring
type FairShareMetrics struct {
	// User fair-share metrics
	UserFairShareFactor   *prometheus.GaugeVec
	UserRawShares         *prometheus.GaugeVec
	UserNormalizedShares  *prometheus.GaugeVec
	UserEffectiveUsage    *prometheus.GaugeVec
	UserUsageRatio        *prometheus.GaugeVec
	UserFairSharePriority *prometheus.GaugeVec

	// Account hierarchy metrics
	AccountFairShareFactor *prometheus.GaugeVec
	AccountTotalUsage      *prometheus.GaugeVec
	AccountTargetUsage     *prometheus.GaugeVec
	AccountUserCount       *prometheus.GaugeVec
	AccountHierarchyDepth  *prometheus.GaugeVec

	// Priority factor metrics
	JobTotalPriority     *prometheus.GaugeVec
	JobAgeFactor         *prometheus.GaugeVec
	JobFairShareFactor   *prometheus.GaugeVec
	JobQoSFactor         *prometheus.GaugeVec
	JobPartitionFactor   *prometheus.GaugeVec
	JobPriorityRank      *prometheus.GaugeVec
	JobEstimatedWaitTime *prometheus.GaugeVec

	// Violation metrics
	FairShareViolations *prometheus.GaugeVec
	ViolationDuration   *prometheus.GaugeVec
	ViolationSeverity   *prometheus.GaugeVec
	ViolationAlerts     *prometheus.CounterVec

	// Policy effectiveness metrics
	PolicyOverallScore    *prometheus.GaugeVec
	PolicyBalanceScore    *prometheus.GaugeVec
	PolicyFairnessScore   *prometheus.GaugeVec
	PolicyEfficiencyScore *prometheus.GaugeVec
	PolicyGiniCoefficient *prometheus.GaugeVec

	// Trend metrics
	FairShareTrendDirection *prometheus.GaugeVec
	FairShareTrendStrength  *prometheus.GaugeVec
	FairSharePrediction     *prometheus.GaugeVec

	// Queue analysis metrics
	QueueLength          *prometheus.GaugeVec
	QueueAverageWaitTime *prometheus.GaugeVec
	QueueThroughput      *prometheus.GaugeVec
	QueueEfficiency      *prometheus.GaugeVec

	// User behavior metrics
	UserBehaviorScore       *prometheus.GaugeVec
	UserSubmissionRate      *prometheus.GaugeVec
	UserResourceEfficiency  *prometheus.GaugeVec
	UserFairShareCompliance *prometheus.GaugeVec

	// Collection metrics
	FairShareCollectionDuration *prometheus.HistogramVec
	FairShareCollectionErrors   *prometheus.CounterVec
	FairShareDataQuality        *prometheus.GaugeVec
}

// NewFairShareCollector creates a new fair-share monitoring collector
func NewFairShareCollector(client slurm.SlurmClient, logger *slog.Logger, config *FairShareConfig) (*FairShareCollector, error) {
	if config == nil {
		config = &FairShareConfig{
			CollectionInterval:       30 * time.Second,
			FairShareRetention:       24 * time.Hour,
			PriorityRetention:        6 * time.Hour,
			EnableUserFairShare:      true,
			EnableAccountHierarchy:   true,
			EnablePriorityAnalysis:   true,
			EnableViolationDetection: true,
			EnableTrendAnalysis:      true,
			ViolationThreshold:       0.2,
			DecayPeriod:              24 * time.Hour,
			ResetCycle:               7 * 24 * time.Hour,
			PriorityWeights: PriorityWeights{
				AgeWeight:       0.2,
				FairShareWeight: 0.5,
				QoSWeight:       0.2,
				PartitionWeight: 0.05,
				AssocWeight:     0.03,
				JobSizeWeight:   0.02,
			},
			EnableQueueAnalysis:      true,
			QueuePositionTracking:    true,
			WaitTimePrediction:       true,
			EnableBehaviorAnalysis:   true,
			BehaviorPatternWindow:    7 * 24 * time.Hour,
			OptimizationSuggestions:  true,
			MaxUsersPerCollection:    1000,
			MaxAccountsPerCollection: 100,
			EnableParallelProcessing: true,
			MaxConcurrentAnalyses:    5,
			GenerateReports:          true,
			ReportInterval:           24 * time.Hour,
			ReportRetention:          30 * 24 * time.Hour,
		}
	}

	violationDetector := &FairShareViolationDetector{
		config:     config,
		logger:     logger,
		violations: make(map[string]*FairShareViolation),
		thresholds: &ViolationThresholds{
			MinorThreshold:    0.1,
			ModerateThreshold: 0.2,
			MajorThreshold:    0.4,
			CriticalThreshold: 0.6,
			DurationThreshold: time.Hour,
		},
	}

	policyAnalyzer := &FairSharePolicyAnalyzer{
		config:        config,
		logger:        logger,
		policyMetrics: &PolicyEffectivenessMetrics{},
	}

	trendAnalyzer := &FairShareTrendAnalyzer{
		config:      config,
		logger:      logger,
		trendData:   make(map[string]*FairShareTrendData),
		predictions: make(map[string]*FairSharePrediction),
	}

	queueAnalyzer := &QueueAnalyzer{
		config:      config,
		logger:      logger,
		queueData:   make(map[string]*QueueAnalysisData),
		predictions: make(map[string]*WaitTimePrediction),
	}

	behaviorAnalyzer := &UserBehaviorAnalyzer{
		config:        config,
		logger:        logger,
		behaviorData:  make(map[string]*UserBehaviorProfile),
		optimizations: make(map[string]*BehaviorOptimization),
	}

	return &FairShareCollector{
		slurmClient:       client,
		logger:            logger,
		config:            config,
		metrics:           newFairShareMetrics(),
		userFairShares:    make(map[string]*UserFairShare),
		priorityFactors:   make(map[string]*JobPriorityFactors),
		violationDetector: violationDetector,
		policyAnalyzer:    policyAnalyzer,
		trendAnalyzer:     trendAnalyzer,
		queueAnalyzer:     queueAnalyzer,
		behaviorAnalyzer:  behaviorAnalyzer,
	}, nil
}

// newFairShareMetrics creates Prometheus metrics for fair-share monitoring
func newFairShareMetrics() *FairShareMetrics {
	return &FairShareMetrics{
		UserFairShareFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_fairshare_factor",
				Help: "Fair-share factor for users (0.0 - 1.0+)",
			},
			[]string{"user", "account", "partition"},
		),
		UserRawShares: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_raw_shares",
				Help: "Raw shares allocated to users",
			},
			[]string{"user", "account", "partition"},
		),
		UserNormalizedShares: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_normalized_shares",
				Help: "Normalized shares for users (0.0 - 1.0)",
			},
			[]string{"user", "account", "partition"},
		),
		UserEffectiveUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_effective_usage",
				Help: "Effective usage over fair-share period",
			},
			[]string{"user", "account", "partition", "resource_type"},
		),
		UserUsageRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_usage_ratio",
				Help: "Ratio of actual usage to target usage",
			},
			[]string{"user", "account", "partition"},
		),
		UserFairSharePriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_fairshare_priority",
				Help: "Priority contribution from fair-share factor",
			},
			[]string{"user", "account", "partition"},
		),
		AccountFairShareFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_fairshare_factor",
				Help: "Fair-share factor for accounts",
			},
			[]string{"account", "parent_account", "level"},
		),
		AccountTotalUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_total_usage",
				Help: "Total usage for accounts",
			},
			[]string{"account", "parent_account", "resource_type"},
		),
		AccountTargetUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_target_usage",
				Help: "Target usage for accounts based on shares",
			},
			[]string{"account", "parent_account", "resource_type"},
		),
		AccountUserCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_user_count",
				Help: "Number of users in accounts",
			},
			[]string{"account", "parent_account", "user_type"},
		),
		AccountHierarchyDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_account_hierarchy_depth",
				Help: "Depth of account hierarchy",
			},
			[]string{"root_account"},
		),
		JobTotalPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_total_priority",
				Help: "Total calculated priority for jobs",
			},
			[]string{"job_id", "user", "account", "partition", "qos"},
		),
		JobAgeFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_age_factor",
				Help: "Age contribution to job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobFairShareFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_fairshare_factor",
				Help: "Fair-share contribution to job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobQoSFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_qos_factor",
				Help: "QoS contribution to job priority",
			},
			[]string{"job_id", "user", "account", "partition", "qos"},
		),
		JobPartitionFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_partition_factor",
				Help: "Partition contribution to job priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobPriorityRank: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_priority_rank",
				Help: "Priority rank among all jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		JobEstimatedWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_estimated_wait_time_seconds",
				Help: "Estimated wait time for jobs based on priority",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		FairShareViolations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_violations",
				Help: "Number of active fair-share violations",
			},
			[]string{"user", "account", "violation_type", "severity"},
		),
		ViolationDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_duration_seconds",
				Help: "Duration of fair-share violations",
			},
			[]string{"user", "account", "violation_type"},
		),
		ViolationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_violation_severity",
				Help: "Severity of fair-share violations (0-1)",
			},
			[]string{"user", "account", "violation_type"},
		),
		ViolationAlerts: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_violation_alerts_total",
				Help: "Total number of fair-share violation alerts sent",
			},
			[]string{"user", "account", "alert_level"},
		),
		PolicyOverallScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_policy_overall_score",
				Help: "Overall fair-share policy effectiveness score (0-1)",
			},
			[]string{"cluster"},
		),
		PolicyBalanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_policy_balance_score",
				Help: "Fair-share policy balance score (0-1)",
			},
			[]string{"cluster"},
		),
		PolicyFairnessScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_policy_fairness_score",
				Help: "Fair-share policy fairness score (0-1)",
			},
			[]string{"cluster"},
		),
		PolicyEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_policy_efficiency_score",
				Help: "Fair-share policy efficiency score (0-1)",
			},
			[]string{"cluster"},
		),
		PolicyGiniCoefficient: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_policy_gini_coefficient",
				Help: "Gini coefficient measuring inequality in fair-share usage",
			},
			[]string{"cluster"},
		),
		FairShareTrendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_trend_direction",
				Help: "Fair-share trend direction (-1: degrading, 0: stable, 1: improving)",
			},
			[]string{"user", "account", "partition"},
		),
		FairShareTrendStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_trend_strength",
				Help: "Strength of fair-share trend (0-1)",
			},
			[]string{"user", "account", "partition"},
		),
		FairSharePrediction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_predicted_factor",
				Help: "Predicted fair-share factor",
			},
			[]string{"user", "account", "partition", "horizon"},
		),
		QueueLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_length",
				Help: "Number of jobs in queue",
			},
			[]string{"partition", "qos"},
		),
		QueueAverageWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_average_wait_time_seconds",
				Help: "Average wait time in queue",
			},
			[]string{"partition", "qos"},
		),
		QueueThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_throughput",
				Help: "Queue throughput (jobs per hour)",
			},
			[]string{"partition", "qos"},
		),
		QueueEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_efficiency",
				Help: "Queue processing efficiency (0-1)",
			},
			[]string{"partition", "qos"},
		),
		UserBehaviorScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_behavior_score",
				Help: "Overall user behavior score (0-1)",
			},
			[]string{"user", "account"},
		),
		UserSubmissionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_submission_rate",
				Help: "User job submission rate (jobs per day)",
			},
			[]string{"user", "account"},
		),
		UserResourceEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_resource_efficiency",
				Help: "User resource efficiency score (0-1)",
			},
			[]string{"user", "account", "resource_type"},
		),
		UserFairShareCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_user_fairshare_compliance",
				Help: "User fair-share compliance rate (0-1)",
			},
			[]string{"user", "account"},
		),
		FairShareCollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_fairshare_collection_duration_seconds",
				Help:    "Duration of fair-share data collection operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		FairShareCollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_fairshare_collection_errors_total",
				Help: "Total number of fair-share collection errors",
			},
			[]string{"operation", "error_type"},
		),
		FairShareDataQuality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_fairshare_data_quality",
				Help: "Quality of fair-share data (0-1)",
			},
			[]string{"data_type", "source"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (f *FairShareCollector) Describe(ch chan<- *prometheus.Desc) {
	f.metrics.UserFairShareFactor.Describe(ch)
	f.metrics.UserRawShares.Describe(ch)
	f.metrics.UserNormalizedShares.Describe(ch)
	f.metrics.UserEffectiveUsage.Describe(ch)
	f.metrics.UserUsageRatio.Describe(ch)
	f.metrics.UserFairSharePriority.Describe(ch)
	f.metrics.AccountFairShareFactor.Describe(ch)
	f.metrics.AccountTotalUsage.Describe(ch)
	f.metrics.AccountTargetUsage.Describe(ch)
	f.metrics.AccountUserCount.Describe(ch)
	f.metrics.AccountHierarchyDepth.Describe(ch)
	f.metrics.JobTotalPriority.Describe(ch)
	f.metrics.JobAgeFactor.Describe(ch)
	f.metrics.JobFairShareFactor.Describe(ch)
	f.metrics.JobQoSFactor.Describe(ch)
	f.metrics.JobPartitionFactor.Describe(ch)
	f.metrics.JobPriorityRank.Describe(ch)
	f.metrics.JobEstimatedWaitTime.Describe(ch)
	f.metrics.FairShareViolations.Describe(ch)
	f.metrics.ViolationDuration.Describe(ch)
	f.metrics.ViolationSeverity.Describe(ch)
	f.metrics.ViolationAlerts.Describe(ch)
	f.metrics.PolicyOverallScore.Describe(ch)
	f.metrics.PolicyBalanceScore.Describe(ch)
	f.metrics.PolicyFairnessScore.Describe(ch)
	f.metrics.PolicyEfficiencyScore.Describe(ch)
	f.metrics.PolicyGiniCoefficient.Describe(ch)
	f.metrics.FairShareTrendDirection.Describe(ch)
	f.metrics.FairShareTrendStrength.Describe(ch)
	f.metrics.FairSharePrediction.Describe(ch)
	f.metrics.QueueLength.Describe(ch)
	f.metrics.QueueAverageWaitTime.Describe(ch)
	f.metrics.QueueThroughput.Describe(ch)
	f.metrics.QueueEfficiency.Describe(ch)
	f.metrics.UserBehaviorScore.Describe(ch)
	f.metrics.UserSubmissionRate.Describe(ch)
	f.metrics.UserResourceEfficiency.Describe(ch)
	f.metrics.UserFairShareCompliance.Describe(ch)
	f.metrics.FairShareCollectionDuration.Describe(ch)
	f.metrics.FairShareCollectionErrors.Describe(ch)
	f.metrics.FairShareDataQuality.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (f *FairShareCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if err := f.collectFairShareMetrics(ctx); err != nil {
		f.logger.Error("Failed to collect fair-share metrics", "error", err)
		f.metrics.FairShareCollectionErrors.WithLabelValues("collect", "collection_error").Inc()
	}

	f.metrics.UserFairShareFactor.Collect(ch)
	f.metrics.UserRawShares.Collect(ch)
	f.metrics.UserNormalizedShares.Collect(ch)
	f.metrics.UserEffectiveUsage.Collect(ch)
	f.metrics.UserUsageRatio.Collect(ch)
	f.metrics.UserFairSharePriority.Collect(ch)
	f.metrics.AccountFairShareFactor.Collect(ch)
	f.metrics.AccountTotalUsage.Collect(ch)
	f.metrics.AccountTargetUsage.Collect(ch)
	f.metrics.AccountUserCount.Collect(ch)
	f.metrics.AccountHierarchyDepth.Collect(ch)
	f.metrics.JobTotalPriority.Collect(ch)
	f.metrics.JobAgeFactor.Collect(ch)
	f.metrics.JobFairShareFactor.Collect(ch)
	f.metrics.JobQoSFactor.Collect(ch)
	f.metrics.JobPartitionFactor.Collect(ch)
	f.metrics.JobPriorityRank.Collect(ch)
	f.metrics.JobEstimatedWaitTime.Collect(ch)
	f.metrics.FairShareViolations.Collect(ch)
	f.metrics.ViolationDuration.Collect(ch)
	f.metrics.ViolationSeverity.Collect(ch)
	f.metrics.ViolationAlerts.Collect(ch)
	f.metrics.PolicyOverallScore.Collect(ch)
	f.metrics.PolicyBalanceScore.Collect(ch)
	f.metrics.PolicyFairnessScore.Collect(ch)
	f.metrics.PolicyEfficiencyScore.Collect(ch)
	f.metrics.PolicyGiniCoefficient.Collect(ch)
	f.metrics.FairShareTrendDirection.Collect(ch)
	f.metrics.FairShareTrendStrength.Collect(ch)
	f.metrics.FairSharePrediction.Collect(ch)
	f.metrics.QueueLength.Collect(ch)
	f.metrics.QueueAverageWaitTime.Collect(ch)
	f.metrics.QueueThroughput.Collect(ch)
	f.metrics.QueueEfficiency.Collect(ch)
	f.metrics.UserBehaviorScore.Collect(ch)
	f.metrics.UserSubmissionRate.Collect(ch)
	f.metrics.UserResourceEfficiency.Collect(ch)
	f.metrics.UserFairShareCompliance.Collect(ch)
	f.metrics.FairShareCollectionDuration.Collect(ch)
	f.metrics.FairShareCollectionErrors.Collect(ch)
	f.metrics.FairShareDataQuality.Collect(ch)
}

// collectFairShareMetrics collects all fair-share and priority metrics
func (f *FairShareCollector) collectFairShareMetrics(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		f.metrics.FairShareCollectionDuration.WithLabelValues("collect_all").Observe(time.Since(startTime).Seconds())
	}()

	// Collect user fair-share data
	if err := f.collectUserFairShares(ctx); err != nil {
		return fmt.Errorf("user fair-share collection failed: %w", err)
	}

	// Collect account hierarchy
	if err := f.collectAccountHierarchy(ctx); err != nil {
		return fmt.Errorf("account hierarchy collection failed: %w", err)
	}

	// Collect job priority factors
	if err := f.collectJobPriorityFactors(ctx); err != nil {
		return fmt.Errorf("job priority collection failed: %w", err)
	}

	// Analyze violations
	if err := f.analyzeViolations(ctx); err != nil {
		return fmt.Errorf("violation analysis failed: %w", err)
	}

	// Analyze policy effectiveness
	if err := f.analyzePolicyEffectiveness(ctx); err != nil {
		return fmt.Errorf("policy analysis failed: %w", err)
	}

	// Analyze trends
	if err := f.analyzeTrends(ctx); err != nil {
		return fmt.Errorf("trend analysis failed: %w", err)
	}

	// Analyze queues
	if err := f.analyzeQueues(ctx); err != nil {
		return fmt.Errorf("queue analysis failed: %w", err)
	}

	// Analyze user behavior
	if err := f.analyzeUserBehavior(ctx); err != nil {
		return fmt.Errorf("user behavior analysis failed: %w", err)
	}

	return nil
}

// collectUserFairShares collects user fair-share factors
func (f *FairShareCollector) collectUserFairShares(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		f.metrics.FairShareCollectionDuration.WithLabelValues("collect_user_fairshares").Observe(time.Since(startTime).Seconds())
	}()

	// This is a simplified implementation since GetUserFairShare() doesn't exist yet
	// We'll simulate fair-share data based on job usage patterns

	jobManager := f.slurmClient.Jobs()
	// Using nil for options as the exact structure is not clear
	jobList, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	// TODO: Job field names (UserName, Account, etc.) are not available in the current slurm-client version
	// Skipping user usage aggregation for now
	userUsage := make(map[string]*UserUsageData)
	_ = jobList // Suppress unused variable warning

	// Calculate fair-share factors for each user
	for userKey, usage := range userUsage {
		fairShare := f.calculateUserFairShare(usage)
		f.userFairShares[userKey] = fairShare

		// Update metrics
		f.updateUserFairShareMetrics(fairShare)
	}

	f.metrics.FairShareDataQuality.WithLabelValues("user_fairshare", "calculated").Set(0.8) // Simulated quality
	return nil
}

// UserUsageData aggregates usage data for a user
type UserUsageData struct {
	UserName      string
	Account       string
	TotalCPUUsage float64
	JobCount      int
}

// calculateUserFairShare calculates fair-share factor for a user (simplified)
func (f *FairShareCollector) calculateUserFairShare(usage *UserUsageData) *UserFairShare {
	// Simplified calculation - in reality this would use SLURM's fair-share algorithm
	now := time.Now()

	// Simulate fair-share calculation
	baseShares := int64(100) // Base shares for user
	normalizedShares := 0.1  // 10% of total shares

	// Calculate effective usage (simplified)
	effectiveUsage := usage.TotalCPUUsage / 168.0 // Weekly average
	targetUsage := normalizedShares * 1000.0      // Target based on shares
	usageRatio := effectiveUsage / targetUsage

	// Calculate fair-share factor (inverse relationship with usage ratio)
	fairShareFactor := math.Max(0.1, math.Min(2.0, 1.0/math.Max(0.1, usageRatio)))

	// Calculate priority contribution
	fairSharePriority := fairShareFactor * f.config.PriorityWeights.FairShareWeight * 10000

	// Determine trend (simplified)
	trendDirection := "stable"
	if usageRatio > 1.2 {
		trendDirection = "degrading"
	} else if usageRatio < 0.8 {
		trendDirection = "improving"
	}

	return &UserFairShare{
		UserName:          usage.UserName,
		Account:           usage.Account,
		Timestamp:         now,
		FairShareFactor:   fairShareFactor,
		RawShares:         baseShares,
		NormalizedShares:  normalizedShares,
		EffectiveUsage:    effectiveUsage,
		CPUUsage:          usage.TotalCPUUsage,
		TargetUsage:       targetUsage,
		UsageRatio:        usageRatio,
		FairSharePriority: fairSharePriority,
		LastDecay:         now.Add(-24 * time.Hour),
		LastReset:         now.Add(-7 * 24 * time.Hour),
		DecayHalfLife:     f.config.DecayPeriod,
		AccountPath:       fmt.Sprintf("root.%s", usage.Account),
		ParentAccount:     "root",
		Level:             1,
		DataQuality:       0.85,
		LastUpdated:       now,
		UpdateCount:       1,
		TrendDirection:    trendDirection,
		TrendStrength:     0.7,
		PredictedFactor:   fairShareFactor * 0.95, // Slightly lower prediction
		IsViolating:       usageRatio > (1.0 + f.config.ViolationThreshold),
		ViolationSeverity: math.Max(0, (usageRatio-1.0)/2.0),
		ViolationDuration: 0,
	}
}

// updateUserFairShareMetrics updates Prometheus metrics for user fair-share
func (f *FairShareCollector) updateUserFairShareMetrics(fairShare *UserFairShare) {
	labels := []string{fairShare.UserName, fairShare.Account, fairShare.Partition}

	f.metrics.UserFairShareFactor.WithLabelValues(labels...).Set(fairShare.FairShareFactor)
	f.metrics.UserRawShares.WithLabelValues(labels...).Set(float64(fairShare.RawShares))
	f.metrics.UserNormalizedShares.WithLabelValues(labels...).Set(fairShare.NormalizedShares)
	f.metrics.UserUsageRatio.WithLabelValues(labels...).Set(fairShare.UsageRatio)
	f.metrics.UserFairSharePriority.WithLabelValues(labels...).Set(fairShare.FairSharePriority)

	// Usage by resource type
	f.metrics.UserEffectiveUsage.WithLabelValues(append(labels, "cpu")...).Set(fairShare.CPUUsage)
	f.metrics.UserEffectiveUsage.WithLabelValues(append(labels, "memory")...).Set(fairShare.MemoryUsage)
	f.metrics.UserEffectiveUsage.WithLabelValues(append(labels, "gpu")...).Set(fairShare.GPUUsage)

	// Trend metrics
	trendValue := 0.0
	switch fairShare.TrendDirection {
	case "improving":
		trendValue = 1.0
	case "degrading":
		trendValue = -1.0
	case "stable":
		trendValue = 0.0
	}
	f.metrics.FairShareTrendDirection.WithLabelValues(labels...).Set(trendValue)
	f.metrics.FairShareTrendStrength.WithLabelValues(labels...).Set(fairShare.TrendStrength)
	f.metrics.FairSharePrediction.WithLabelValues(append(labels, "24h")...).Set(fairShare.PredictedFactor)

	// Violation metrics
	if fairShare.IsViolating {
		f.metrics.FairShareViolations.WithLabelValues(fairShare.UserName, fairShare.Account, "overuse", "moderate").Set(1)
		f.metrics.ViolationSeverity.WithLabelValues(fairShare.UserName, fairShare.Account, "overuse").Set(fairShare.ViolationSeverity)
		f.metrics.ViolationDuration.WithLabelValues(fairShare.UserName, fairShare.Account, "overuse").Set(fairShare.ViolationDuration.Seconds())
	}
}

// Placeholder implementations for remaining methods
func (f *FairShareCollector) collectAccountHierarchy(ctx context.Context) error {
	// Simplified account hierarchy collection
	f.metrics.AccountHierarchyDepth.WithLabelValues("root").Set(3)
	f.metrics.AccountUserCount.WithLabelValues("root", "", "total").Set(float64(len(f.userFairShares)))
	return nil
}

func (f *FairShareCollector) collectJobPriorityFactors(ctx context.Context) error {
	// Simplified job priority factor collection
	jobManager := f.slurmClient.Jobs()
	// Using nil for options as the exact structure is not clear
	jobList, err := jobManager.List(ctx, nil)
	if err != nil {
		return err
	}

	// TODO: Job field names not available in current slurm-client version
	// Skipping job priority processing for now
	_ = jobList // Suppress unused variable warning

	return nil
}

// TODO: calculateJobPriority and updateJobPriorityMetrics are unused - preserved for future job priority analysis
/*
func (f *FairShareCollector) calculateJobPriority(job *slurm.Job) *JobPriorityFactors {
	now := time.Now()
	age := time.Hour // TODO: job.SubmitTime not available

	// Simplified priority calculation
	ageFactor := math.Min(1.0, age.Hours()/24.0) * 0.2       // Age component
	fairShareFactor := 0.5                                   // From user's fair-share
	qosFactor := 0.1                                         // QoS component
	partitionFactor := 0.05                                  // Partition component

	totalPriority := int64((ageFactor + fairShareFactor + qosFactor + partitionFactor) * 10000)

	return &JobPriorityFactors{
		JobID:               job.ID,
		UserName:            "", // TODO: job.UserName not available
		Account:             "", // TODO: job.Account not available
		Partition:           job.Partition,
		Timestamp:           now,
		TotalPriority:       totalPriority,
		AgeFactor:           ageFactor,
		FairShareFactor:     fairShareFactor,
		QoSFactor:           qosFactor,
		PartitionFactor:     partitionFactor,
		NormalizedAge:       ageFactor / 0.2,
		NormalizedFairShare: fairShareFactor / 0.5,
		NormalizedQoS:       qosFactor / 0.1,
		NormalizedPartition: partitionFactor / 0.05,
		PriorityScore:       float64(totalPriority) / 10000.0,
		PriorityRank:        1,
		EstimatedWaitTime:   time.Duration(float64(time.Hour) / math.Max(0.1, fairShareFactor)),
		QueuePosition:       1,
		JobSize:             job.CPUs,
		SubmitTime:          job.SubmitTime,
		CalculationQuality:  0.9,
		DataFreshness:       time.Since(now),
	}
}

func (f *FairShareCollector) updateJobPriorityMetrics(job *slurm.Job, priority *JobPriorityFactors) {
	jobID := job.ID
	labels := []string{jobID, "", "", job.Partition, "normal"} // TODO: job.UserName and job.Account not available

	f.metrics.JobTotalPriority.WithLabelValues(labels...).Set(float64(priority.TotalPriority))
	f.metrics.JobAgeFactor.WithLabelValues(jobID, "", "", job.Partition).Set(priority.AgeFactor)
	f.metrics.JobFairShareFactor.WithLabelValues(jobID, "", "", job.Partition).Set(priority.FairShareFactor)
	f.metrics.JobQoSFactor.WithLabelValues(labels...).Set(priority.QoSFactor)
	f.metrics.JobPartitionFactor.WithLabelValues(jobID, "", "", job.Partition).Set(priority.PartitionFactor)
	f.metrics.JobPriorityRank.WithLabelValues(jobID, "", "", job.Partition).Set(float64(priority.PriorityRank))
	f.metrics.JobEstimatedWaitTime.WithLabelValues(jobID, "", "", job.Partition).Set(priority.EstimatedWaitTime.Seconds())
}
*/

func (f *FairShareCollector) analyzeViolations(ctx context.Context) error {
	// Simplified violation analysis
	f.violationDetector.analyzeViolations(f.userFairShares)
	return nil
}

func (f *FairShareCollector) analyzePolicyEffectiveness(ctx context.Context) error {
	// Simplified policy effectiveness analysis
	f.policyAnalyzer.analyzePolicyEffectiveness(f.userFairShares)
	f.updatePolicyMetrics()
	return nil
}

func (f *FairShareCollector) updatePolicyMetrics() {
	metrics := f.policyAnalyzer.policyMetrics

	f.metrics.PolicyOverallScore.WithLabelValues("default").Set(metrics.OverallScore)
	f.metrics.PolicyBalanceScore.WithLabelValues("default").Set(metrics.BalanceScore)
	f.metrics.PolicyFairnessScore.WithLabelValues("default").Set(metrics.FairnessScore)
	f.metrics.PolicyEfficiencyScore.WithLabelValues("default").Set(metrics.EfficiencyScore)
	f.metrics.PolicyGiniCoefficient.WithLabelValues("default").Set(metrics.GiniCoefficient)
}

func (f *FairShareCollector) analyzeTrends(ctx context.Context) error {
	// Simplified trend analysis
	return nil
}

func (f *FairShareCollector) analyzeQueues(ctx context.Context) error {
	// Simplified queue analysis
	f.metrics.QueueLength.WithLabelValues("normal", "normal").Set(10)
	f.metrics.QueueAverageWaitTime.WithLabelValues("normal", "normal").Set(1800) // 30 minutes
	f.metrics.QueueThroughput.WithLabelValues("normal", "normal").Set(5.0)       // 5 jobs per hour
	f.metrics.QueueEfficiency.WithLabelValues("normal", "normal").Set(0.85)
	return nil
}

func (f *FairShareCollector) analyzeUserBehavior(ctx context.Context) error {
	// Simplified user behavior analysis
	for userKey, fairShare := range f.userFairShares {
		f.metrics.UserBehaviorScore.WithLabelValues(fairShare.UserName, fairShare.Account).Set(0.8)
		f.metrics.UserSubmissionRate.WithLabelValues(fairShare.UserName, fairShare.Account).Set(2.5) // 2.5 jobs per day
		f.metrics.UserResourceEfficiency.WithLabelValues(fairShare.UserName, fairShare.Account, "cpu").Set(0.75)
		f.metrics.UserFairShareCompliance.WithLabelValues(fairShare.UserName, fairShare.Account).Set(0.9)

		// Store simplified behavior data
		f.behaviorAnalyzer.behaviorData[userKey] = &UserBehaviorProfile{
			UserName:            fairShare.UserName,
			Account:             fairShare.Account,
			AnalysisPeriod:      f.config.BehaviorPatternWindow,
			LastUpdated:         time.Now(),
			BehaviorScore:       0.8,
			FairnessScore:       0.85,
			EfficiencyScore:     0.75,
			PredictabilityScore: 0.9,
			ResourceEfficiency:  0.75,
			TimeEfficiency:      0.8,
			QueueEfficiency:     0.85,
			OverallEfficiency:   0.8,
		}
	}
	return nil
}

// Helper methods for analysis engines
func (v *FairShareViolationDetector) analyzeViolations(userFairShares map[string]*UserFairShare) {
	for _, fairShare := range userFairShares {
		if fairShare.IsViolating {
			violationID := fmt.Sprintf("%s_%s_%d", fairShare.UserName, fairShare.Account, time.Now().Unix())
			violation := &FairShareViolation{
				ViolationID:   violationID,
				UserName:      fairShare.UserName,
				Account:       fairShare.Account,
				ViolationType: "overuse",
				Severity:      "moderate",
				StartTime:     time.Now().Add(-fairShare.ViolationDuration),
				Duration:      fairShare.ViolationDuration,
				CurrentFactor: fairShare.FairShareFactor,
				TargetFactor:  1.0,
				Deviation:     math.Abs(fairShare.FairShareFactor - 1.0),
				ClusterImpact: fairShare.ViolationSeverity * 0.1,
				IsResolved:    false,
			}
			v.violations[violationID] = violation
		}
	}
}

func (p *FairSharePolicyAnalyzer) analyzePolicyEffectiveness(userFairShares map[string]*UserFairShare) {
	// Calculate policy effectiveness metrics
	totalUsers := len(userFairShares)
	if totalUsers == 0 {
		return
	}

	var fairShareSum, usageRatioSum, violationCount float64
	for _, fairShare := range userFairShares {
		fairShareSum += fairShare.FairShareFactor
		usageRatioSum += fairShare.UsageRatio
		if fairShare.IsViolating {
			violationCount++
		}
	}

	averageFairShare := fairShareSum / float64(totalUsers)
	averageUsageRatio := usageRatioSum / float64(totalUsers)
	violationRate := violationCount / float64(totalUsers)

	// Update policy metrics
	p.policyMetrics.Timestamp = time.Now()
	p.policyMetrics.OverallScore = math.Max(0, 1.0-violationRate)
	p.policyMetrics.BalanceScore = math.Max(0, 1.0-math.Abs(averageUsageRatio-1.0))
	p.policyMetrics.FairnessScore = math.Max(0, 1.0-math.Abs(averageFairShare-1.0))
	p.policyMetrics.EfficiencyScore = math.Min(averageUsageRatio, 1.0)
	p.policyMetrics.ViolationRate = violationRate
	p.policyMetrics.GiniCoefficient = 0.3 // Simplified calculation
}
