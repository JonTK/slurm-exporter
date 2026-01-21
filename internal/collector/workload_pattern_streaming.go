package collector

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// WorkloadPatternStreamingSLURMClient defines the interface for workload pattern streaming operations
type WorkloadPatternStreamingSLURMClient interface {
	StreamWorkloadPatternEvents(ctx context.Context) (<-chan WorkloadPatternEvent, error)
	GetPatternStreamingConfiguration(ctx context.Context) (*PatternStreamingConfiguration, error)
	GetActivePatternStreams(ctx context.Context) ([]*ActivePatternStream, error)
	GetPatternEventHistory(ctx context.Context, patternID string, duration time.Duration) ([]*PatternEvent, error)
	GetPatternStreamingMetrics(ctx context.Context) (*PatternStreamingMetrics, error)
	GetPatternEventFilters(ctx context.Context) ([]*PatternEventFilter, error)
	ConfigurePatternStreaming(ctx context.Context, config *PatternStreamingConfiguration) error
	GetPatternStreamingStatus(ctx context.Context) (*PatternStreamingStatus, error)
	GetPatternEventSubscriptions(ctx context.Context) ([]*PatternEventSubscription, error)
	ManagePatternEventSubscription(ctx context.Context, subscription *PatternEventSubscription) error
	GetPatternEventProcessingStats(ctx context.Context) (*PatternEventProcessingStats, error)
	GetPatternStreamingPerformanceMetrics(ctx context.Context) (*PatternStreamingPerformanceMetrics, error)
}

// WorkloadPatternEvent represents a comprehensive workload pattern detection event
type WorkloadPatternEvent struct {
	// Event identification
	EventID        string
	PatternID      string
	PatternName    string
	PatternType    string
	EventType      string
	EventTimestamp time.Time
	DetectionTime  time.Time
	SequenceNumber int64
	CorrelationID  string

	// Pattern characteristics
	PatternCategory    string
	PatternSubtype     string
	PatternSignature   string
	PatternFingerprint string
	Confidence         float64
	Significance       float64
	Frequency          float64
	Periodicity        time.Duration

	// Temporal characteristics
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	RecurrenceCount   int
	RecurrencePattern string
	TimeOfDay         string
	DayOfWeek         string
	SeasonalFactors   []string

	// Workload metrics
	JobCount         int
	UserCount        int
	AccountCount     int
	PartitionCount   int
	TotalCPUHours    float64
	TotalMemoryHours float64
	TotalGPUHours    float64
	AverageJobSize   float64

	// Resource patterns
	CPUPattern        string
	MemoryPattern     string
	GPUPattern        string
	IOPattern         string
	NetworkPattern    string
	StoragePattern    string
	ResourceProfile   string
	ResourceSignature string

	// Job characteristics
	JobSizeDistribution     map[string]int
	JobDurationDistribution map[string]int
	JobTypeDistribution     map[string]int
	JobPriorityDistribution map[string]int
	QueueDistribution       map[string]int
	PartitionDistribution   map[string]int
	UserDistribution        map[string]int
	AccountDistribution     map[string]int

	// Submission patterns
	SubmissionRate         float64
	SubmissionBurstiness   float64
	SubmissionRegularity   float64
	PeakSubmissionTime     time.Time
	SubmissionTimePattern  string
	BatchSubmissions       int
	InteractiveSubmissions int
	ArrayJobSubmissions    int

	// Execution patterns
	ExecutionEfficiency float64
	CompletionRate      float64
	FailureRate         float64
	CancellationRate    float64
	ResubmissionRate    float64
	WaitTimePattern     string
	RuntimePattern      string
	EfficiencyPattern   string

	// Resource utilization patterns
	CPUUtilizationPattern    string
	MemoryUtilizationPattern string
	GPUUtilizationPattern    string
	OverallUtilization       float64
	PeakUtilization          float64
	ValleyUtilization        float64
	UtilizationVariance      float64
	UtilizationTrend         string

	// Scaling patterns
	ScalingPattern      string
	ScalingFrequency    float64
	ScalingAmplitude    float64
	ScalingEfficiency   float64
	ElasticityPattern   string
	BurstingPattern     string
	ContractionPattern  string
	OptimalScalingLevel float64

	// Performance patterns
	PerformanceProfile           string
	ThroughputPattern            string
	LatencyPattern               string
	PerformanceEfficiencyPattern string
	SpeedupPattern               string
	ScalabilityPattern           string
	BottleneckPattern            string
	ContentionPattern            string

	// Cost patterns
	CostPattern               string
	CostPerJobPattern         string
	CostEfficiencyPattern     string
	BudgetUtilizationPattern  string
	CostOptimizationPotential float64
	WastedResourceCost        float64
	OptimalCostConfiguration  string
	CostTrend                 string

	// User behavior patterns
	UserActivityPattern      string
	UserSubmissionPattern    string
	UserResourcePattern      string
	UserEfficiencyPattern    string
	UserCollaborationPattern string
	UserMigrationPattern     string
	UserLearningCurve        string
	UserSatisfactionPattern  string

	// Anomaly indicators
	AnomalyScore            float64
	DeviationFromNormal     float64
	StatisticalSignificance float64
	OutlierCount            int
	AnomalyType             string
	AnomalyDescription      string
	BaselineComparison      float64
	TrendDeviation          float64

	// Predictive insights
	PredictedNextOccurrence *time.Time
	PredictedDuration       time.Duration
	PredictedImpact         float64
	PredictedResourceNeeds  map[string]float64
	PredictedCost           float64
	PredictedBottlenecks    []string
	PredictionConfidence    float64
	PredictionHorizon       time.Duration

	// Business impact
	BusinessImpact           float64
	ProductivityImpact       float64
	CostImpact               float64
	SLAImpact                float64
	UserSatisfactionImpact   float64
	ResourceEfficiencyImpact float64
	CapacityPlanningImpact   float64
	StrategicValue           float64

	// Optimization opportunities
	OptimizationPotential    float64
	RecommendedActions       []string
	ExpectedImprovement      float64
	ImplementationComplexity float64
	ROIEstimate              float64
	RiskAssessment           float64
	PriorityScore            float64
	AutomationCandidate      bool

	// Pattern evolution
	EvolutionStage       string
	MaturityLevel        float64
	StabilityScore       float64
	ChangeProbability    float64
	AdaptationRate       float64
	LearningRate         float64
	OptimizationProgress float64
	ConvergenceStatus    string

	// Correlations
	CorrelatedPatterns   []string
	CausalRelationships  []string
	DependentPatterns    []string
	InfluencingFactors   []string
	ExternalTriggers     []string
	EnvironmentalFactors []string
	SystemicInfluences   []string
	FeedbackLoops        []string

	// Classification
	PrimaryClassification    string
	SecondaryClassifications []string
	IndustryCategory         string
	WorkloadClass            string
	ComputeParadigm          string
	ApplicationDomain        string
	ResearchArea             string
	UseCaseType              string

	// Metadata
	DataSources         []string
	AnalysisMethods     []string
	ConfidenceIntervals map[string][2]float64
	StatisticalMetrics  map[string]float64
	QualityScore        float64
	ReliabilityScore    float64
	CompletenessScore   float64
	ValidationStatus    string
}

// PatternStreamingConfiguration represents pattern streaming configuration
type PatternStreamingConfiguration struct {
	// Basic settings
	StreamingEnabled      bool
	DetectionInterval     time.Duration
	AnalysisWindowSize    time.Duration
	EventBufferSize       int
	MaxConcurrentAnalysis int
	StreamTimeout         time.Duration

	// Pattern detection settings
	PatternTypes           []string
	MinPatternConfidence   float64
	MinPatternSignificance float64
	MinPatternFrequency    int
	MinPatternDuration     time.Duration
	MaxPatternAge          time.Duration
	PatternMergeThreshold  float64

	// Analysis settings
	TimeSeriesAnalysis  bool
	FrequencyAnalysis   bool
	CorrelationAnalysis bool
	CausalityAnalysis   bool
	PredictiveAnalysis  bool
	AnomalyDetection    bool
	TrendAnalysis       bool
	SeasonalAnalysis    bool

	// Resource analysis
	ResourcePatternDetection bool
	UtilizationAnalysis      bool
	EfficiencyAnalysis       bool
	WasteDetection           bool
	BottleneckDetection      bool
	ContentionAnalysis       bool
	ScalingAnalysis          bool
	CapacityAnalysis         bool

	// Behavioral analysis
	UserBehaviorAnalysis      bool
	SubmissionPatternAnalysis bool
	ExecutionPatternAnalysis  bool
	FailurePatternAnalysis    bool
	MigrationPatternAnalysis  bool
	CollaborationAnalysis     bool
	LearningCurveAnalysis     bool
	SatisfactionAnalysis      bool

	// Performance settings
	ParallelProcessing  bool
	GPUAcceleration     bool
	CachingEnabled      bool
	CompressionEnabled  bool
	SamplingEnabled     bool
	SamplingRate        float64
	IncrementalAnalysis bool
	StreamProcessing    bool

	// Machine learning settings
	MLEnabled            bool
	ModelTypes           []string
	TrainingEnabled      bool
	OnlineLearning       bool
	ModelUpdateFrequency time.Duration
	FeatureEngineering   bool
	AutoMLEnabled        bool
	EnsembleMethods      bool

	// Optimization settings
	OptimizationAnalysis    bool
	CostOptimization        bool
	PerformanceOptimization bool
	ResourceOptimization    bool
	SchedulingOptimization  bool
	PlacementOptimization   bool
	ScalingOptimization     bool
	WorkflowOptimization    bool

	// Business analysis
	BusinessImpactAnalysis bool
	ROICalculation         bool
	CostBenefitAnalysis    bool
	RiskAssessment         bool
	ComplianceChecking     bool
	SLAMonitoring          bool
	KPITracking            bool
	ValueStreamMapping     bool

	// Alerting and actions
	AlertingEnabled       bool
	AlertThresholds       map[string]float64
	AutomatedActions      bool
	RecommendationEngine  bool
	DecisionSupport       bool
	PrescriptiveAnalytics bool
	PlaybookIntegration   bool
	WorkflowAutomation    bool

	// Integration settings
	ExternalDataSources  []string
	StreamingPlatforms   []string
	AnalyticsPlatforms   []string
	VisualizationTools   []string
	NotificationChannels []string
	StorageBackends      []string
	APIEndpoints         []string
	WebhookEndpoints     []string

	// Data management
	DataRetentionPeriod time.Duration
	DataArchiving       bool
	DataCompression     bool
	DataEncryption      bool
	DataAnonymization   bool
	DataQualityChecks   bool
	DataLineageTracking bool
	DataGovernance      bool
}

// ActivePatternStream represents an active pattern detection stream
type ActivePatternStream struct {
	// Stream identification
	StreamID       string
	StreamName     string
	StreamType     string
	CreatedAt      time.Time
	LastActivityAt time.Time
	Status         string

	// Analysis state
	AnalysisWindowStart time.Time
	AnalysisWindowEnd   time.Time
	EventsAnalyzed      int64
	PatternsDetected    int64
	ActivePatterns      int
	EvolvingPatterns    int
	StablePatterns      int

	// Performance metrics
	AnalysisRate         float64
	DetectionLatency     time.Duration
	ProcessingEfficiency float64
	ResourceUtilization  float64
	MemoryUsage          int64
	CPUUsage             float64

	// Quality metrics
	DetectionAccuracy float64
	FalsePositiveRate float64
	FalseNegativeRate float64
	PrecisionScore    float64
	RecallScore       float64
	F1Score           float64

	// Pattern statistics
	MostFrequentPattern    string
	MostSignificantPattern string
	LongestPattern         string
	MostComplexPattern     string
	MostValuablePattern    string
	EmergingPatterns       []string
	DecliningPatterns      []string

	// Business metrics
	BusinessValueIdentified   float64
	CostSavingsPotential      float64
	EfficiencyGainsPotential  float64
	OptimizationOpportunities int
	RiskIndicators            int
	ComplianceIssues          int

	// ML model performance
	ModelAccuracy      float64
	ModelPrecision     float64
	ModelRecall        float64
	ModelF1Score       float64
	ModelTrainingTime  time.Duration
	ModelInferenceTime time.Duration
	ModelVersion       string
	ModelLastUpdated   time.Time
}

// PatternEvent represents an individual pattern event
type PatternEvent struct {
	Event             WorkloadPatternEvent
	AnalysisMetadata  map[string]interface{}
	DetectionMetadata map[string]interface{}
	ValidationResults map[string]interface{}
}

// PatternStreamingMetrics represents overall pattern streaming metrics
type PatternStreamingMetrics struct {
	// Stream metrics
	TotalStreams      int
	ActiveStreams     int
	HealthyStreams    int
	DegradedStreams   int
	StreamEfficiency  float64
	StreamReliability float64

	// Pattern detection metrics
	TotalPatternsDetected int64
	UniquePatterns        int64
	RecurringPatterns     int64
	EmergingPatterns      int64
	StablePatterns        int64
	DecliningPatterns     int64

	// Analysis metrics
	EventsAnalyzed   int64
	AnalysisRate     float64
	AnalysisLatency  time.Duration
	AnalysisAccuracy float64
	AnalysisCoverage float64
	AnalysisDepth    float64

	// Pattern quality metrics
	AverageConfidence   float64
	AverageSignificance float64
	PatternStability    float64
	PatternDiversity    float64
	PatternComplexity   float64
	PatternValue        float64

	// Resource metrics
	CPUUtilization     float64
	MemoryUtilization  float64
	StorageUtilization float64
	NetworkUtilization float64
	GPUUtilization     float64
	ResourceEfficiency float64

	// ML metrics
	ModelPerformance   float64
	PredictionAccuracy float64
	TrainingEfficiency float64
	InferenceSpeed     float64
	ModelComplexity    float64
	FeatureImportance  map[string]float64

	// Business impact metrics
	TotalBusinessValue     float64
	IdentifiedSavings      float64
	OptimizationPotential  float64
	RiskReduction          float64
	ComplianceImprovement  float64
	UserSatisfactionImpact float64

	// Optimization metrics
	OptimizationsIdentified  int64
	OptimizationsImplemented int64
	EfficiencyGains          float64
	CostReductions           float64
	PerformanceImprovements  float64
	ResourceSavings          float64

	// Trend metrics
	PatternGrowthRate float64
	ComplexityTrend   float64
	ValueTrend        float64
	EfficiencyTrend   float64
	AdoptionRate      float64
	EvolutionRate     float64
}

// PatternEventFilter represents pattern event filtering configuration
type PatternEventFilter struct {
	FilterID   string
	FilterName string
	FilterType string
	Enabled    bool
	Priority   int

	// Pattern filters
	PatternTypes      []string
	PatternCategories []string
	MinConfidence     float64
	MinSignificance   float64
	MinFrequency      int
	MinDuration       time.Duration

	// Resource filters
	ResourceTypes       []string
	MinResourceUsage    float64
	MaxResourceUsage    float64
	UtilizationRange    [2]float64
	EfficiencyThreshold float64
	WasteThreshold      float64

	// Business filters
	MinBusinessImpact   float64
	MinCostImpact       float64
	MinROI              float64
	RequireOptimization bool
	RequireAnomaly      bool
	RequireCompliance   bool

	// Time filters
	TimeWindows        []string
	RecurrenceRequired bool
	MinRecurrence      int
	SeasonalOnly       bool
	TrendRequired      bool
	RecentOnly         bool

	// User filters
	UserGroups      []string
	Accounts        []string
	Departments     []string
	Projects        []string
	Applications    []string
	WorkloadClasses []string

	// Statistics
	EventsMatched    int64
	EventsRejected   int64
	ProcessingTime   time.Duration
	FilterEfficiency float64
	LastMatch        time.Time
	MatchRate        float64
}

// PatternStreamingStatus represents pattern streaming system status
type PatternStreamingStatus struct {
	// System status
	OverallHealth    float64
	SystemStatus     string
	LastStatusCheck  time.Time
	UptimePercentage float64
	PerformanceScore float64
	ReliabilityScore float64

	// Detection status
	DetectionCoverage  float64
	DetectionAccuracy  float64
	DetectionLatency   time.Duration
	ActiveDetectors    int
	FailedDetectors    int
	DetectorEfficiency float64

	// Analysis status
	AnalysisBacklog    int64
	AnalysisCapacity   float64
	AnalysisThroughput float64
	AnalysisQuality    float64
	ModelStatus        string
	ModelPerformance   float64

	// Pattern status
	ActivePatterns    int
	MonitoredPatterns int
	EvolvingPatterns  int
	StablePatterns    int
	CriticalPatterns  int
	ValuePatterns     int

	// Resource status
	ResourceUtilization float64
	MemoryPressure      float64
	CPUPressure         float64
	StoragePressure     float64
	NetworkPressure     float64
	ScalingNeeded       bool

	// Business status
	BusinessValueDelivered   float64
	OptimizationsSuggested   int
	OptimizationsImplemented int
	CostSavingsRealized      float64
	EfficiencyGains          float64
	ComplianceScore          float64

	// Risk status
	RiskLevel           float64
	CriticalRisks       int
	MitigatedRisks      int
	EmergingRisks       int
	RiskTrend           string
	RiskMitigationScore float64

	// Recommendations
	ImmediateActions          []string
	OptimizationOpportunities []string
	CapacityRecommendations   []string
	ConfigurationSuggestions  []string
	TrainingRecommendations   []string
	IntegrationOptions        []string
}

// PatternEventSubscription represents a pattern event subscription
type PatternEventSubscription struct {
	// Subscription details
	SubscriptionID   string
	SubscriberID     string
	SubscriptionName string
	SubscriptionType string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Status           string

	// Pattern interests
	PatternTypes          []string
	PatternCategories     []string
	MinConfidence         float64
	MinSignificance       float64
	MinBusinessImpact     float64
	NotificationThreshold float64

	// Delivery settings
	DeliveryMethod string
	Endpoint       string
	Format         string
	Batching       bool
	BatchSize      int
	BatchInterval  time.Duration
	Compression    bool

	// Filtering
	Filters                []string
	RequireValidation      bool
	RequireOptimization    bool
	RequireAnomaly         bool
	RequirePrediction      bool
	IncludeMetadata        bool
	IncludeRecommendations bool

	// Quality settings
	DeliveryGuarantee  string
	MaxRetries         int
	RetryBackoff       time.Duration
	TimeoutDuration    time.Duration
	DeadLetterHandling string
	ErrorHandling      string

	// Rate limiting
	RateLimit        int
	BurstLimit       int
	QuotaLimit       int64
	QuotaWindow      time.Duration
	ThrottleStrategy string
	PriorityLevel    int

	// Business settings
	CostCenter         string
	BudgetLimit        float64
	UsageTracking      bool
	BillingEnabled     bool
	SLALevel           string
	ComplianceRequired bool

	// Statistics
	EventsDelivered int64
	EventsFailed    int64
	BytesDelivered  int64
	AverageLatency  time.Duration
	SuccessRate     float64
	LastDelivery    time.Time
	TotalCost       float64
	ValueDelivered  float64
}

// PatternEventProcessingStats represents pattern processing statistics
type PatternEventProcessingStats struct {
	// Processing metrics
	TotalEventsProcessed int64
	SuccessfulProcessing int64
	FailedProcessing     int64
	PartialProcessing    int64
	SkippedProcessing    int64
	ReprocessedEvents    int64

	// Pattern detection metrics
	PatternsDetected  int64
	PatternsValidated int64
	PatternsRejected  int64
	PatternsMerged    int64
	PatternsEvolved   int64
	PatternsExpired   int64

	// Performance metrics
	AverageDetectionTime time.Duration
	MinDetectionTime     time.Duration
	MaxDetectionTime     time.Duration
	DetectionTimeP50     time.Duration
	DetectionTimeP95     time.Duration
	DetectionTimeP99     time.Duration

	// Analysis metrics
	AnalysisDepth         float64
	AnalysisAccuracy      float64
	AnalysisCoverage      float64
	FeatureExtractionTime time.Duration
	ModelInferenceTime    time.Duration
	ValidationTime        time.Duration

	// Quality metrics
	ConfidenceDistribution   map[string]int64
	SignificanceDistribution map[string]int64
	ComplexityDistribution   map[string]int64
	ValueDistribution        map[string]int64
	QualityScore             float64
	ReliabilityScore         float64

	// Resource metrics
	CPUTimeConsumed    time.Duration
	MemoryAllocated    int64
	StorageUsed        int64
	NetworkBandwidth   int64
	GPUTimeConsumed    time.Duration
	ResourceEfficiency float64

	// Business metrics
	BusinessValueGenerated    float64
	OptimizationsIdentified   int64
	CostSavingsIdentified     float64
	EfficiencyGainsIdentified float64
	RisksIdentified           int64
	ComplianceIssuesFound     int64

	// ML metrics
	ModelUpdates             int64
	ModelRetraining          int64
	FeatureImportanceChanges int64
	PredictionAccuracy       float64
	ModelDrift               float64
	DataDrift                float64
}

// PatternStreamingPerformanceMetrics represents performance metrics
type PatternStreamingPerformanceMetrics struct {
	// Latency metrics
	EndToEndLatency        time.Duration
	DetectionLatency       time.Duration
	AnalysisLatency        time.Duration
	ValidationLatency      time.Duration
	DeliveryLatency        time.Duration
	TotalProcessingLatency time.Duration

	// Throughput metrics
	EventThroughput    float64
	PatternThroughput  float64
	AnalysisThroughput float64
	DetectionRate      float64
	ProcessingRate     float64
	DeliveryRate       float64

	// Accuracy metrics
	DetectionAccuracy      float64
	ClassificationAccuracy float64
	PredictionAccuracy     float64
	ValidationAccuracy     float64
	OverallAccuracy        float64
	ErrorRate              float64

	// Efficiency metrics
	ProcessingEfficiency float64
	DetectionEfficiency  float64
	AnalysisEfficiency   float64
	ResourceEfficiency   float64
	CostEfficiency       float64
	TimeEfficiency       float64

	// Quality metrics
	PatternQuality        float64
	DataQuality           float64
	AnalysisQuality       float64
	RecommendationQuality float64
	ServiceQuality        float64
	OutputQuality         float64

	// Coverage metrics
	EventCoverage    float64
	PatternCoverage  float64
	UserCoverage     float64
	ResourceCoverage float64
	TimeCoverage     float64
	FeatureCoverage  float64

	// Scalability metrics
	ScalabilityIndex   float64
	ElasticityScore    float64
	LoadHandling       float64
	GrowthCapability   float64
	PerformanceScaling float64
	ResourceScaling    float64

	// Business metrics
	ValueGeneration  float64
	ROI              float64
	CostPerPattern   float64
	ValuePerPattern  float64
	BusinessImpact   float64
	UserSatisfaction float64

	// Reliability metrics
	Availability   float64
	Reliability    float64
	Durability     float64
	FaultTolerance float64
	RecoveryTime   time.Duration
	DataIntegrity  float64
}

// WorkloadPatternStreamingCollector collects workload pattern streaming metrics
type WorkloadPatternStreamingCollector struct {
	client                  WorkloadPatternStreamingSLURMClient
	patternEvents           *prometheus.Desc
	activePatternStreams    *prometheus.Desc
	patternDetectionRate    *prometheus.Desc
	patternConfidence       *prometheus.Desc
	patternSignificance     *prometheus.Desc
	patternComplexity       *prometheus.Desc
	patternValue            *prometheus.Desc
	resourcePatterns        *prometheus.Desc
	userBehaviorPatterns    *prometheus.Desc
	anomalyPatterns         *prometheus.Desc
	optimizationPotential   *prometheus.Desc
	businessImpactScore     *prometheus.Desc
	predictionAccuracy      *prometheus.Desc
	mlModelPerformance      *prometheus.Desc
	analysisEfficiency      *prometheus.Desc
	patternEvolution        *prometheus.Desc
	costOptimizationMetrics *prometheus.Desc
	compliancePatterns      *prometheus.Desc
	scalingPatterns         *prometheus.Desc
	streamConfigMetrics     map[string]*prometheus.Desc
	eventFilterMetrics      map[string]*prometheus.Desc
	subscriptionMetrics     map[string]*prometheus.Desc
	processingStatsMetrics  map[string]*prometheus.Desc
	performanceMetrics      map[string]*prometheus.Desc
}

// NewWorkloadPatternStreamingCollector creates a new workload pattern streaming collector
func NewWorkloadPatternStreamingCollector(client WorkloadPatternStreamingSLURMClient) *WorkloadPatternStreamingCollector {
	return &WorkloadPatternStreamingCollector{
		client: client,
		patternEvents: prometheus.NewDesc(
			"slurm_pattern_events_total",
			"Total number of workload pattern events",
			[]string{"pattern_type", "category", "confidence_level", "significance"},
			nil,
		),
		activePatternStreams: prometheus.NewDesc(
			"slurm_active_pattern_streams",
			"Number of active pattern detection streams",
			[]string{"stream_type", "status", "analysis_type"},
			nil,
		),
		patternDetectionRate: prometheus.NewDesc(
			"slurm_pattern_detection_rate",
			"Rate of pattern detection per second",
			[]string{"pattern_type", "detection_method"},
			nil,
		),
		patternConfidence: prometheus.NewDesc(
			"slurm_pattern_confidence_score",
			"Confidence score of detected patterns",
			[]string{"pattern_type", "pattern_id"},
			nil,
		),
		patternSignificance: prometheus.NewDesc(
			"slurm_pattern_significance_score",
			"Significance score of detected patterns",
			[]string{"pattern_type", "impact_level"},
			nil,
		),
		patternComplexity: prometheus.NewDesc(
			"slurm_pattern_complexity_score",
			"Complexity score of detected patterns",
			[]string{"pattern_type", "complexity_level"},
			nil,
		),
		patternValue: prometheus.NewDesc(
			"slurm_pattern_business_value",
			"Business value of detected patterns",
			[]string{"pattern_type", "value_category"},
			nil,
		),
		resourcePatterns: prometheus.NewDesc(
			"slurm_resource_pattern_metrics",
			"Resource utilization pattern metrics",
			[]string{"resource_type", "pattern_name", "metric"},
			nil,
		),
		userBehaviorPatterns: prometheus.NewDesc(
			"slurm_user_behavior_patterns",
			"User behavior pattern metrics",
			[]string{"behavior_type", "user_group", "pattern"},
			nil,
		),
		anomalyPatterns: prometheus.NewDesc(
			"slurm_anomaly_pattern_score",
			"Anomaly score for detected patterns",
			[]string{"anomaly_type", "severity", "pattern_id"},
			nil,
		),
		optimizationPotential: prometheus.NewDesc(
			"slurm_pattern_optimization_potential",
			"Optimization potential identified from patterns",
			[]string{"optimization_type", "resource_type", "impact_area"},
			nil,
		),
		businessImpactScore: prometheus.NewDesc(
			"slurm_pattern_business_impact",
			"Business impact score of patterns",
			[]string{"impact_type", "pattern_category"},
			nil,
		),
		predictionAccuracy: prometheus.NewDesc(
			"slurm_pattern_prediction_accuracy",
			"Accuracy of pattern-based predictions",
			[]string{"prediction_type", "time_horizon"},
			nil,
		),
		mlModelPerformance: prometheus.NewDesc(
			"slurm_pattern_ml_model_performance",
			"Machine learning model performance metrics",
			[]string{"model_type", "metric", "version"},
			nil,
		),
		analysisEfficiency: prometheus.NewDesc(
			"slurm_pattern_analysis_efficiency",
			"Efficiency of pattern analysis",
			[]string{"analysis_type", "efficiency_metric"},
			nil,
		),
		patternEvolution: prometheus.NewDesc(
			"slurm_pattern_evolution_metrics",
			"Pattern evolution and stability metrics",
			[]string{"evolution_stage", "stability_level"},
			nil,
		),
		costOptimizationMetrics: prometheus.NewDesc(
			"slurm_pattern_cost_optimization",
			"Cost optimization opportunities from patterns",
			[]string{"cost_type", "optimization_category"},
			nil,
		),
		compliancePatterns: prometheus.NewDesc(
			"slurm_pattern_compliance_score",
			"Compliance-related pattern metrics",
			[]string{"compliance_type", "pattern_category"},
			nil,
		),
		scalingPatterns: prometheus.NewDesc(
			"slurm_scaling_pattern_metrics",
			"Scaling pattern metrics",
			[]string{"scaling_type", "pattern_name", "efficiency"},
			nil,
		),
		streamConfigMetrics:    make(map[string]*prometheus.Desc),
		eventFilterMetrics:     make(map[string]*prometheus.Desc),
		subscriptionMetrics:    make(map[string]*prometheus.Desc),
		processingStatsMetrics: make(map[string]*prometheus.Desc),
		performanceMetrics:     make(map[string]*prometheus.Desc),
	}
}

// Describe implements prometheus.Collector
func (c *WorkloadPatternStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.patternEvents
	ch <- c.activePatternStreams
	ch <- c.patternDetectionRate
	ch <- c.patternConfidence
	ch <- c.patternSignificance
	ch <- c.patternComplexity
	ch <- c.patternValue
	ch <- c.resourcePatterns
	ch <- c.userBehaviorPatterns
	ch <- c.anomalyPatterns
	ch <- c.optimizationPotential
	ch <- c.businessImpactScore
	ch <- c.predictionAccuracy
	ch <- c.mlModelPerformance
	ch <- c.analysisEfficiency
	ch <- c.patternEvolution
	ch <- c.costOptimizationMetrics
	ch <- c.compliancePatterns
	ch <- c.scalingPatterns

	// Dynamic metrics
	for _, desc := range c.streamConfigMetrics {
		ch <- desc
	}
	for _, desc := range c.eventFilterMetrics {
		ch <- desc
	}
	for _, desc := range c.subscriptionMetrics {
		ch <- desc
	}
	for _, desc := range c.processingStatsMetrics {
		ch <- desc
	}
	for _, desc := range c.performanceMetrics {
		ch <- desc
	}
}

// Collect implements prometheus.Collector
func (c *WorkloadPatternStreamingCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Collect streaming configuration metrics
	c.collectStreamingConfiguration(ctx, ch)

	// Collect active streams metrics
	c.collectActiveStreams(ctx, ch)

	// Collect streaming metrics
	c.collectStreamingMetrics(ctx, ch)

	// Collect pattern status metrics
	c.collectPatternStatus(ctx, ch)

	// Collect event subscription metrics
	c.collectEventSubscriptions(ctx, ch)

	// Collect event filter metrics
	c.collectEventFilters(ctx, ch)

	// Collect processing statistics
	c.collectProcessingStats(ctx, ch)

	// Collect performance metrics
	c.collectPerformanceMetrics(ctx, ch)
}

func (c *WorkloadPatternStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetPatternStreamingConfiguration(ctx)
	if err != nil {
		return
	}

	// Configuration enabled/disabled metrics
	configurations := map[string]bool{
		"streaming_enabled":          config.StreamingEnabled,
		"time_series_analysis":       config.TimeSeriesAnalysis,
		"frequency_analysis":         config.FrequencyAnalysis,
		"correlation_analysis":       config.CorrelationAnalysis,
		"causality_analysis":         config.CausalityAnalysis,
		"predictive_analysis":        config.PredictiveAnalysis,
		"anomaly_detection":          config.AnomalyDetection,
		"trend_analysis":             config.TrendAnalysis,
		"seasonal_analysis":          config.SeasonalAnalysis,
		"resource_pattern_detection": config.ResourcePatternDetection,
		"utilization_analysis":       config.UtilizationAnalysis,
		"efficiency_analysis":        config.EfficiencyAnalysis,
		"waste_detection":            config.WasteDetection,
		"bottleneck_detection":       config.BottleneckDetection,
		"user_behavior_analysis":     config.UserBehaviorAnalysis,
		"ml_enabled":                 config.MLEnabled,
		"training_enabled":           config.TrainingEnabled,
		"online_learning":            config.OnlineLearning,
		"optimization_analysis":      config.OptimizationAnalysis,
		"business_impact_analysis":   config.BusinessImpactAnalysis,
		"alerting_enabled":           config.AlertingEnabled,
		"automated_actions":          config.AutomatedActions,
		"recommendation_engine":      config.RecommendationEngine,
	}

	for name, enabled := range configurations {
		value := 0.0
		if enabled {
			value = 1.0
		}

		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_pattern_stream_config_"+name,
				"Pattern streaming configuration for "+name,
				nil,
				nil,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.streamConfigMetrics[name],
			prometheus.GaugeValue,
			value,
		)
	}

	// Numeric configuration metrics
	numericConfigs := map[string]float64{
		"min_pattern_confidence":   config.MinPatternConfidence,
		"min_pattern_significance": config.MinPatternSignificance,
		"min_pattern_frequency":    float64(config.MinPatternFrequency),
		"pattern_merge_threshold":  config.PatternMergeThreshold,
		"sampling_rate":            config.SamplingRate,
		"event_buffer_size":        float64(config.EventBufferSize),
		"max_concurrent_analysis":  float64(config.MaxConcurrentAnalysis),
	}

	for name, value := range numericConfigs {
		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_pattern_stream_config_"+name,
				"Pattern streaming configuration value for "+name,
				nil,
				nil,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.streamConfigMetrics[name],
			prometheus.GaugeValue,
			value,
		)
	}
}

func (c *WorkloadPatternStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActivePatternStreams(ctx)
	if err != nil {
		return
	}

	// Count streams by type and status
	streamCounts := make(map[string]int)
	for _, stream := range streams {
		key := stream.StreamType + "_" + stream.Status + "_pattern"
		streamCounts[key]++

		// Individual stream metrics
		ch <- prometheus.MustNewConstMetric(
			c.patternDetectionRate,
			prometheus.GaugeValue,
			stream.AnalysisRate,
			stream.StreamType, "streaming",
		)

		ch <- prometheus.MustNewConstMetric(
			c.analysisEfficiency,
			prometheus.GaugeValue,
			stream.ProcessingEfficiency,
			stream.StreamType, "processing",
		)

		ch <- prometheus.MustNewConstMetric(
			c.mlModelPerformance,
			prometheus.GaugeValue,
			stream.ModelAccuracy,
			"accuracy", stream.StreamType, stream.ModelVersion,
		)

		ch <- prometheus.MustNewConstMetric(
			c.businessImpactScore,
			prometheus.GaugeValue,
			stream.BusinessValueIdentified,
			"value_identified", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.optimizationPotential,
			prometheus.GaugeValue,
			stream.CostSavingsPotential,
			"cost_savings", "all", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.optimizationPotential,
			prometheus.GaugeValue,
			stream.EfficiencyGainsPotential,
			"efficiency_gains", "all", stream.StreamType,
		)
	}

	// Aggregate stream counts
	for key, count := range streamCounts {
		parts := strings.Split(key, "_")
		if len(parts) >= 3 {
			ch <- prometheus.MustNewConstMetric(
				c.activePatternStreams,
				prometheus.GaugeValue,
				float64(count),
				parts[0], parts[1], parts[2],
			)
		}
	}
}

func (c *WorkloadPatternStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetPatternStreamingMetrics(ctx)
	if err != nil {
		return
	}

	// Pattern detection metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.CounterValue,
		float64(metrics.TotalPatternsDetected),
		"all", "all", "all", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.GaugeValue,
		float64(metrics.UniquePatterns),
		"unique", "all", "all", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.GaugeValue,
		float64(metrics.RecurringPatterns),
		"recurring", "all", "all", "current",
	)

	// Pattern quality metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternConfidence,
		prometheus.GaugeValue,
		metrics.AverageConfidence,
		"all", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternSignificance,
		prometheus.GaugeValue,
		metrics.AverageSignificance,
		"all", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternComplexity,
		prometheus.GaugeValue,
		metrics.PatternComplexity,
		"all", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternValue,
		prometheus.GaugeValue,
		metrics.PatternValue,
		"all", "average",
	)

	// Analysis metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		metrics.AnalysisAccuracy,
		"overall", "accuracy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		metrics.AnalysisCoverage,
		"overall", "coverage",
	)

	// ML metrics
	ch <- prometheus.MustNewConstMetric(
		c.mlModelPerformance,
		prometheus.GaugeValue,
		metrics.ModelPerformance,
		"overall", "performance", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictionAccuracy,
		prometheus.GaugeValue,
		metrics.PredictionAccuracy,
		"overall", "all",
	)

	// Business impact metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpactScore,
		prometheus.GaugeValue,
		metrics.TotalBusinessValue,
		"total_value", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimizationMetrics,
		prometheus.GaugeValue,
		metrics.IdentifiedSavings,
		"identified", "total",
	)

	ch <- prometheus.MustNewConstMetric(
		c.optimizationPotential,
		prometheus.GaugeValue,
		metrics.OptimizationPotential,
		"overall", "all", "potential",
	)

	// Trend metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternEvolution,
		prometheus.GaugeValue,
		metrics.PatternGrowthRate,
		"growth_rate", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvolution,
		prometheus.GaugeValue,
		metrics.EvolutionRate,
		"evolution_rate", "overall",
	)
}

func (c *WorkloadPatternStreamingCollector) collectPatternStatus(ctx context.Context, ch chan<- prometheus.Metric) {
	status, err := c.client.GetPatternStreamingStatus(ctx)
	if err != nil {
		return
	}

	// System status
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		status.OverallHealth,
		"system", "health",
	)

	// Detection status
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		status.DetectionCoverage,
		"detection", "coverage",
	)

	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		status.DetectionAccuracy,
		"detection", "accuracy",
	)

	// Pattern status
	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.GaugeValue,
		float64(status.ActivePatterns),
		"active", "all", "all", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.GaugeValue,
		float64(status.StablePatterns),
		"stable", "all", "all", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.GaugeValue,
		float64(status.EvolvingPatterns),
		"evolving", "all", "all", "current",
	)

	// Resource status
	ch <- prometheus.MustNewConstMetric(
		c.resourcePatterns,
		prometheus.GaugeValue,
		status.ResourceUtilization,
		"overall", "system", "utilization",
	)

	// Business status
	ch <- prometheus.MustNewConstMetric(
		c.businessImpactScore,
		prometheus.GaugeValue,
		status.BusinessValueDelivered,
		"delivered", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimizationMetrics,
		prometheus.GaugeValue,
		status.CostSavingsRealized,
		"realized", "actual",
	)

	ch <- prometheus.MustNewConstMetric(
		c.compliancePatterns,
		prometheus.GaugeValue,
		status.ComplianceScore,
		"overall", "score",
	)

	// Risk status
	ch <- prometheus.MustNewConstMetric(
		c.anomalyPatterns,
		prometheus.GaugeValue,
		status.RiskLevel,
		"risk", "overall", "system",
	)
}

func (c *WorkloadPatternStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetPatternEventSubscriptions(ctx)
	if err != nil {
		return
	}

	for _, sub := range subscriptions {
		// Subscription metrics
		ch <- prometheus.MustNewConstMetric(
			c.patternEvents,
			prometheus.CounterValue,
			float64(sub.EventsDelivered),
			"subscription", sub.SubscriptionType, "delivered", sub.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			c.patternEvents,
			prometheus.CounterValue,
			float64(sub.EventsFailed),
			"subscription", sub.SubscriptionType, "failed", sub.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			c.businessImpactScore,
			prometheus.GaugeValue,
			sub.ValueDelivered,
			"subscription_value", sub.SubscriptionID,
		)
	}
}

func (c *WorkloadPatternStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetPatternEventFilters(ctx)
	if err != nil {
		return
	}

	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.patternEvents,
			prometheus.CounterValue,
			float64(filter.EventsMatched),
			"filter", filter.FilterType, "matched", "active",
		)

		ch <- prometheus.MustNewConstMetric(
			c.analysisEfficiency,
			prometheus.GaugeValue,
			filter.FilterEfficiency,
			"filter", filter.FilterID,
		)
	}
}

func (c *WorkloadPatternStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetPatternEventProcessingStats(ctx)
	if err != nil {
		return
	}

	// Processing metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.CounterValue,
		float64(stats.TotalEventsProcessed),
		"processing", "all", "all", "total",
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternEvents,
		prometheus.CounterValue,
		float64(stats.PatternsDetected),
		"detected", "all", "all", "total",
	)

	// Performance metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		stats.AverageDetectionTime.Seconds(),
		"detection_time", "average",
	)

	// Quality metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		stats.QualityScore,
		"quality", "overall",
	)

	// ML metrics
	ch <- prometheus.MustNewConstMetric(
		c.predictionAccuracy,
		prometheus.GaugeValue,
		stats.PredictionAccuracy,
		"processing", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.mlModelPerformance,
		prometheus.GaugeValue,
		stats.ModelDrift,
		"drift", "model", "current",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpactScore,
		prometheus.GaugeValue,
		stats.BusinessValueGenerated,
		"processing_value", "generated",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimizationMetrics,
		prometheus.GaugeValue,
		stats.CostSavingsIdentified,
		"processing", "identified",
	)
}

func (c *WorkloadPatternStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perf, err := c.client.GetPatternStreamingPerformanceMetrics(ctx)
	if err != nil {
		return
	}

	// Latency metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.EndToEndLatency.Seconds(),
		"latency", "end_to_end",
	)

	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.DetectionLatency.Seconds(),
		"latency", "detection",
	)

	// Throughput metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternDetectionRate,
		prometheus.GaugeValue,
		perf.PatternThroughput,
		"overall", "throughput",
	)

	// Accuracy metrics
	ch <- prometheus.MustNewConstMetric(
		c.predictionAccuracy,
		prometheus.GaugeValue,
		perf.DetectionAccuracy,
		"detection", "performance",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictionAccuracy,
		prometheus.GaugeValue,
		perf.ClassificationAccuracy,
		"classification", "performance",
	)

	// Efficiency metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.ProcessingEfficiency,
		"performance", "processing",
	)

	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.ResourceEfficiency,
		"performance", "resource",
	)

	// Quality metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.PatternQuality,
		"quality", "patterns",
	)

	// Coverage metrics
	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.EventCoverage,
		"coverage", "events",
	)

	ch <- prometheus.MustNewConstMetric(
		c.analysisEfficiency,
		prometheus.GaugeValue,
		perf.PatternCoverage,
		"coverage", "patterns",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpactScore,
		prometheus.GaugeValue,
		perf.ValueGeneration,
		"performance_value", "generation_rate",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimizationMetrics,
		prometheus.GaugeValue,
		perf.CostPerPattern,
		"performance", "cost_per_pattern",
	)

	// Scalability metrics
	ch <- prometheus.MustNewConstMetric(
		c.scalingPatterns,
		prometheus.GaugeValue,
		perf.ScalabilityIndex,
		"performance", "scalability", "index",
	)
}

// TODO: Following resource pattern collection method is unused - preserved for future pattern analysis
/*
func (c *WorkloadPatternStreamingCollector) collectResourcePatterns(event WorkloadPatternEvent, ch chan<- prometheus.Metric) {
	// Resource utilization patterns
	resourceMetrics := map[string]float64{
		"cpu_hours":    event.TotalCPUHours,
		"memory_hours": event.TotalMemoryHours,
		"gpu_hours":    event.TotalGPUHours,
	}

	for resource, value := range resourceMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.resourcePatterns,
			prometheus.GaugeValue,
			value,
			resource, event.PatternName, "total_hours",
		)
	}

	// Pattern-specific metrics
	ch <- prometheus.MustNewConstMetric(
		c.patternConfidence,
		prometheus.GaugeValue,
		event.Confidence,
		event.PatternType, event.PatternID,
	)

	ch <- prometheus.MustNewConstMetric(
		c.patternSignificance,
		prometheus.GaugeValue,
		event.Significance,
		event.PatternType, "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.anomalyPatterns,
		prometheus.GaugeValue,
		event.AnomalyScore,
		event.AnomalyType, "medium", event.PatternID,
	)

	ch <- prometheus.MustNewConstMetric(
		c.optimizationPotential,
		prometheus.GaugeValue,
		event.OptimizationPotential,
		"pattern", event.PatternType, event.PatternID,
	)

	ch <- prometheus.MustNewConstMetric(
		c.businessImpactScore,
		prometheus.GaugeValue,
		event.BusinessImpact,
		"pattern", event.PatternCategory,
	)
}
*/
