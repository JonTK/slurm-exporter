package collector

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// JobSchedulingStreamingSLURMClient defines the interface for job scheduling streaming operations
type JobSchedulingStreamingSLURMClient interface {
	StreamJobSchedulingEvents(ctx context.Context) (<-chan JobSchedulingEvent, error)
	GetSchedulingStreamingConfiguration(ctx context.Context) (*SchedulingStreamingConfiguration, error)
	GetActiveSchedulingStreams(ctx context.Context) ([]*ActiveSchedulingStream, error)
	GetSchedulingEventHistory(ctx context.Context, jobID string, duration time.Duration) ([]*SchedulingEvent, error)
	GetSchedulingStreamingMetrics(ctx context.Context) (*SchedulingStreamingMetrics, error)
	GetSchedulingEventFilters(ctx context.Context) ([]*SchedulingEventFilter, error)
	ConfigureSchedulingStreaming(ctx context.Context, config *SchedulingStreamingConfiguration) error
	GetSchedulingStreamingHealthStatus(ctx context.Context) (*SchedulingStreamingHealthStatus, error)
	GetSchedulingEventSubscriptions(ctx context.Context) ([]*SchedulingEventSubscription, error)
	ManageSchedulingEventSubscription(ctx context.Context, subscription *SchedulingEventSubscription) error
	GetSchedulingEventProcessingStats(ctx context.Context) (*SchedulingEventProcessingStats, error)
	GetSchedulingStreamingPerformanceMetrics(ctx context.Context) (*SchedulingStreamingPerformanceMetrics, error)
}

// JobSchedulingEvent represents a comprehensive job scheduling event
type JobSchedulingEvent struct {
	// Event identification
	EventID        string
	JobID          string
	JobName        string
	EventType      string
	EventTimestamp time.Time
	SequenceNumber int64
	CorrelationID  string
	ProcessingID   string

	// Scheduling decision data
	SchedulingDecision  string
	SchedulingReason    string
	SchedulingAlgorithm string
	SchedulingPriority  int
	SchedulingScore     float64
	DecisionTimestamp   time.Time
	DecisionLatency     time.Duration

	// Resource allocation
	AllocatedNodes     []string
	AllocatedCPUs      int
	AllocatedMemory    int64
	AllocatedGPUs      int
	AllocatedStorage   int64
	ResourceEfficiency float64
	AllocationScore    float64

	// Queue information
	SourceQueue       string
	TargetQueue       string
	QueuePosition     int
	QueueLength       int
	EstimatedWaitTime time.Duration
	ActualWaitTime    time.Duration
	QueueEfficiency   float64

	// Priority factors
	UserPriority      int
	GroupPriority     int
	AccountPriority   int
	QoSPriority       int
	PartitionPriority int
	FairShareFactor   float64
	AgeFactor         float64
	SizeFactor        float64

	// Constraints and requirements
	Constraints        map[string]string
	Features           []string
	Licenses           []string
	Dependencies       []string
	Reservations       []string
	ExclusiveAccess    bool
	PreemptionEligible bool

	// Backfill information
	BackfillEligible      bool
	BackfillWindow        time.Duration
	BackfillScore         float64
	BackfillPriority      int
	BackfillOpportunities int
	BackfillSuccess       bool

	// Preemption data
	PreemptionRequired  bool
	PreemptedJobs       []string
	PreemptionCost      float64
	PreemptionBenefit   float64
	PreemptionScore     float64
	PreemptionJustified bool

	// Performance metrics
	SchedulingOverhead  time.Duration
	OptimizationCycles  int
	CandidatesEvaluated int
	ConstraintChecks    int
	ScoringsPerformed   int
	DecisionConfidence  float64

	// Impact assessment
	ClusterImpact         float64
	QueueImpact           float64
	UserImpact            float64
	SystemEfficiency      float64
	ResourceFragmentation float64
	SchedulingEfficiency  float64

	// Business metrics
	EstimatedCost    float64
	BusinessPriority int
	SLACompliance    bool
	DeadlineRisk     float64
	ValueScore       float64
	ROIEstimate      float64

	// Optimization data
	OptimizationAttempts int
	OptimizationSuccess  bool
	OptimizationGain     float64
	AlternativeOptions   int
	BestAlternativeScore float64
	OptimizationTime     time.Duration

	// Prediction data
	PredictedRuntime     time.Duration
	PredictedSuccess     float64
	PredictedEfficiency  float64
	PredictedCost        float64
	PredictionConfidence float64
	PredictionAccuracy   float64

	// Anomaly detection
	AnomalyDetected    bool
	AnomalyType        string
	AnomalySeverity    float64
	AnomalyDescription string
	AnomalyMitigation  string
	AnomalyImpact      float64

	// Compliance and audit
	PolicyCompliant   bool
	PolicyViolations  []string
	AuditRequired     bool
	ComplianceScore   float64
	RegulatoryFlags   []string
	SecurityClearance string

	// Stream metadata
	StreamID         string
	ConsumerID       string
	DeliveryAttempts int
	ProcessingTime   time.Duration
	AcknowledgedAt   *time.Time
	RetryCount       int

	// Error handling
	ErrorOccurred     bool
	ErrorMessage      string
	ErrorCode         string
	RecoveryAttempted bool
	RecoverySuccess   bool
	FallbackUsed      bool

	// Event enrichment
	EnrichmentApplied bool
	EnrichmentSources []string
	EnrichmentLatency time.Duration
	AdditionalContext map[string]interface{}
	RelatedEvents     []string
	EventChain        []string
}

// SchedulingStreamingConfiguration represents streaming configuration
type SchedulingStreamingConfiguration struct {
	// Basic settings
	StreamingEnabled     bool
	EventBufferSize      int
	EventBatchSize       int
	EventBatchTimeout    time.Duration
	MaxConcurrentStreams int
	StreamTimeout        time.Duration

	// Performance settings
	CompressionEnabled  bool
	CompressionLevel    int
	EncryptionEnabled   bool
	EncryptionAlgorithm string
	BatchProcessing     bool
	ParallelProcessing  bool

	// Filtering and routing
	EventFiltering  bool
	FilterRules     []string
	EventRouting    bool
	RoutingRules    map[string]string
	PriorityQueuing bool
	DeadLetterQueue bool

	// Reliability settings
	GuaranteedDelivery bool
	OrderingGuaranteed bool
	DuplicateDetection bool
	RetryPolicy        string
	MaxRetries         int
	BackoffMultiplier  float64

	// Monitoring settings
	MetricsEnabled   bool
	TracingEnabled   bool
	ProfilingEnabled bool
	DebugMode        bool
	LogLevel         string
	SamplingRate     float64

	// Resource limits
	MaxMemoryUsage          int64
	MaxCPUUsage             float64
	MaxBandwidth            int64
	MaxEventsPerSecond      int
	BackpressureThreshold   int
	CircuitBreakerThreshold float64

	// Security settings
	AuthenticationRequired bool
	AuthorizationRequired  bool
	AuditLoggingEnabled    bool
	DataMaskingEnabled     bool
	PIIRedactionEnabled    bool
	ComplianceMode         string

	// Advanced features
	EventEnrichment      bool
	CorrelationEnabled   bool
	AggregationEnabled   bool
	WindowingEnabled     bool
	StatefulProcessing   bool
	CheckpointingEnabled bool

	// Integration settings
	WebhookEnabled   bool
	WebhookEndpoints []string
	KafkaEnabled     bool
	KafkaTopics      []string
	SQSEnabled       bool
	SQSQueues        []string

	// Optimization settings
	AdaptiveStreaming    bool
	LoadBalancing        bool
	AutoScaling          bool
	ResourceOptimization bool
	CostOptimization     bool
	PerformanceTuning    bool

	// Business logic
	SLAEnforcement   bool
	BusinessRules    []string
	CostTracking     bool
	ValueTracking    bool
	ROICalculation   bool
	ImpactAssessment bool

	// Machine learning
	MLPrediction         bool
	AnomalyDetection     bool
	PatternRecognition   bool
	TrendAnalysis        bool
	ForecastingEnabled   bool
	RecommendationEngine bool

	// Compliance settings
	DataRetention        time.Duration
	DataArchiving        bool
	LegalHold            bool
	RegulatoryCompliance []string
	PrivacyCompliance    []string
	IndustryStandards    []string
}

// ActiveSchedulingStream represents an active scheduling event stream
type ActiveSchedulingStream struct {
	// Stream identification
	StreamID       string
	StreamName     string
	StreamType     string
	CreatedAt      time.Time
	LastActivityAt time.Time
	Status         string

	// Consumer information
	ConsumerID       string
	ConsumerName     string
	ConsumerType     string
	ConsumerEndpoint string
	ConsumerHealth   string
	LastHeartbeat    time.Time

	// Stream statistics
	EventsDelivered  int64
	EventsPending    int64
	EventsFailed     int64
	EventsFiltered   int64
	BytesTransferred int64
	AverageLatency   time.Duration

	// Performance metrics
	Throughput          float64
	ErrorRate           float64
	SuccessRate         float64
	ProcessingRate      float64
	BackpressureActive  bool
	ResourceUtilization float64

	// Quality metrics
	DeliveryGuarantee   string
	OrderingGuarantee   string
	DuplicateRate       float64
	DataQuality         float64
	EnrichmentQuality   float64
	CorrelationAccuracy float64

	// Business metrics
	BusinessValue        float64
	CostPerEvent         float64
	ROI                  float64
	SLACompliance        float64
	CustomerSatisfaction float64
	ImpactScore          float64

	// Optimization metrics
	OptimizationEnabled bool
	OptimizationGains   float64
	ResourceSavings     float64
	CostSavings         float64
	EfficiencyGains     float64
	PerformanceGains    float64
}

// SchedulingEvent represents an individual scheduling event
type SchedulingEvent struct {
	Event              JobSchedulingEvent
	ProcessingMetadata map[string]interface{}
	DeliveryStatus     string
	RetryInformation   map[string]interface{}
}

// SchedulingStreamingMetrics represents overall streaming metrics
type SchedulingStreamingMetrics struct {
	// Stream metrics
	TotalStreams       int
	ActiveStreams      int
	HealthyStreams     int
	DegradedStreams    int
	FailedStreams      int
	StreamCreationRate float64

	// Event metrics
	TotalEventsProcessed   int64
	EventsPerSecond        float64
	AverageEventSize       int64
	PeakEventRate          float64
	EventBacklog           int64
	EventProcessingLatency time.Duration

	// Performance metrics
	CPUUtilization     float64
	MemoryUtilization  float64
	NetworkUtilization float64
	DiskUtilization    float64
	SystemLoad         float64
	ResourceEfficiency float64

	// Reliability metrics
	Uptime       time.Duration
	Availability float64
	MTBF         time.Duration
	MTTR         time.Duration
	ErrorRate    float64
	RecoveryRate float64

	// Quality metrics
	DataAccuracy       float64
	DataCompleteness   float64
	DataTimeliness     float64
	EnrichmentSuccess  float64
	CorrelationSuccess float64
	DeduplicationRate  float64

	// Business metrics
	TotalBusinessValue   float64
	AverageCostPerEvent  float64
	ROI                  float64
	CustomerSatisfaction float64
	SLACompliance        float64
	BusinessImpact       float64

	// Optimization metrics
	OptimizationOpportunities int
	PotentialSavings          float64
	EfficiencyScore           float64
	PerformanceScore          float64
	QualityScore              float64
	ValueScore                float64

	// Capacity metrics
	CapacityUtilization    float64
	ScalingEvents          int
	ResourceHeadroom       float64
	ProjectedCapacityNeeds float64
	CapacityPlanningScore  float64
	GrowthRate             float64

	// Cost metrics
	TotalCost          float64
	InfrastructureCost float64
	OperationalCost    float64
	DataTransferCost   float64
	StorageCost        float64
	CostTrend          float64

	// Compliance metrics
	ComplianceScore   float64
	PolicyViolations  int
	AuditEvents       int
	SecurityIncidents int
	PrivacyIncidents  int
	RegulatoryIssues  int
}

// SchedulingEventFilter represents event filtering configuration
type SchedulingEventFilter struct {
	FilterID   string
	FilterName string
	FilterType string
	Priority   int
	Enabled    bool

	// Filter criteria
	JobPatterns   []string
	UserPatterns  []string
	QueuePatterns []string
	EventTypes    []string
	MinPriority   int
	MaxPriority   int

	// Performance criteria
	MinEfficiency    float64
	MaxLatency       time.Duration
	MinThroughput    float64
	AnomalyOnly      bool
	ViolationsOnly   bool
	OptimizationOnly bool

	// Business criteria
	MinBusinessValue float64
	MaxCost          float64
	SLAOnly          bool
	CriticalOnly     bool
	ComplianceOnly   bool
	SecurityOnly     bool

	// Statistics
	MatchCount     int64
	RejectCount    int64
	ProcessingTime time.Duration
	Efficiency     float64
	LastMatch      time.Time
	ErrorCount     int
}

// SchedulingStreamingHealthStatus represents streaming health
type SchedulingStreamingHealthStatus struct {
	// Overall health
	HealthScore     float64
	HealthStatus    string
	LastHealthCheck time.Time
	HealthTrend     string
	NextHealthCheck time.Time
	HealthHistory   []float64

	// Component health
	StreamHealth     float64
	ProcessingHealth float64
	StorageHealth    float64
	NetworkHealth    float64
	SecurityHealth   float64
	ComplianceHealth float64

	// Performance health
	LatencyHealth     float64
	ThroughputHealth  float64
	ErrorRateHealth   float64
	ResourceHealth    float64
	ScalabilityHealth float64
	EfficiencyHealth  float64

	// Business health
	SLAHealth      float64
	CostHealth     float64
	ValueHealth    float64
	CustomerHealth float64
	ROIHealth      float64
	ImpactHealth   float64

	// Risk indicators
	RiskScore        float64
	SecurityRisk     float64
	OperationalRisk  float64
	ComplianceRisk   float64
	FinancialRisk    float64
	ReputationalRisk float64

	// Recommendations
	IssuesDetected    int
	CriticalIssues    int
	Recommendations   []string
	AutomatedActions  []string
	RequiredActions   []string
	PreventiveActions []string

	// Predictive health
	PredictedHealth    float64
	HealthForecast     []float64
	DegradationRisk    float64
	FailureRisk        float64
	MaintenanceNeeded  bool
	OptimizationNeeded bool
}

// SchedulingEventSubscription represents an event subscription
type SchedulingEventSubscription struct {
	// Subscription details
	SubscriptionID   string
	SubscriberID     string
	SubscriptionType string
	CreatedAt        time.Time
	ExpiresAt        *time.Time
	Status           string

	// Delivery configuration
	DeliveryMethod string
	Endpoint       string
	Protocol       string
	Format         string
	Compression    bool
	Encryption     bool

	// Filter configuration
	EventFilters        []string
	QueueFilters        []string
	UserFilters         []string
	PriorityThreshold   int
	EfficiencyThreshold float64
	BusinessValueMin    float64

	// Quality of service
	GuaranteedDelivery bool
	OrderingRequired   bool
	MaxRetries         int
	RetryBackoff       time.Duration
	DeadLetterQueue    string
	TimeoutDuration    time.Duration

	// Rate limiting
	RateLimitEnabled bool
	EventsPerSecond  int
	BurstSize        int
	QuotaLimit       int64
	QuotaUsed        int64
	QuotaResetTime   time.Time

	// Security
	AuthRequired          bool
	AuthMethod            string
	APIKey                string
	CertificateThumbprint string
	IPWhitelist           []string
	EncryptionKey         string

	// Monitoring
	MetricsEnabled      bool
	TracingEnabled      bool
	LoggingEnabled      bool
	AlertingEnabled     bool
	HealthCheckEnabled  bool
	HealthCheckInterval time.Duration

	// Statistics
	EventsDelivered int64
	EventsFailed    int64
	LastDelivery    time.Time
	AverageLatency  time.Duration
	SuccessRate     float64
	ErrorRate       float64

	// Cost and billing
	BillingEnabled bool
	CostPerEvent   float64
	TotalCost      float64
	BillingCycle   string
	PaymentMethod  string
	CreditLimit    float64

	// Compliance
	DataRetention   time.Duration
	PIIHandling     string
	AuditLogging    bool
	ComplianceFlags []string
	LegalHold       bool
	DataResidency   string
}

// SchedulingEventProcessingStats represents processing statistics
type SchedulingEventProcessingStats struct {
	// Processing metrics
	TotalEventsProcessed int64
	SuccessfulProcessing int64
	FailedProcessing     int64
	PartialProcessing    int64
	SkippedProcessing    int64
	RetryProcessing      int64

	// Performance metrics
	AverageProcessingTime time.Duration
	MinProcessingTime     time.Duration
	MaxProcessingTime     time.Duration
	P50ProcessingTime     time.Duration
	P95ProcessingTime     time.Duration
	P99ProcessingTime     time.Duration

	// Throughput metrics
	CurrentThroughput    float64
	PeakThroughput       float64
	AverageThroughput    float64
	ThroughputTrend      float64
	ThroughputVariance   float64
	ThroughputEfficiency float64

	// Resource metrics
	CPUTimeConsumed      time.Duration
	MemoryAllocated      int64
	NetworkBandwidthUsed int64
	DiskIOOperations     int64
	ResourceEfficiency   float64
	ResourceCost         float64

	// Queue metrics
	QueueDepth         int64
	QueueLatency       time.Duration
	QueueThroughput    float64
	BackpressureEvents int64
	DroppedEvents      int64
	QueueEfficiency    float64

	// Error metrics
	ErrorsByType          map[string]int64
	ErrorRate             float64
	ErrorTrend            float64
	RecoveryRate          float64
	MeanTimeBetweenErrors time.Duration
	MeanTimeToRecover     time.Duration

	// Business impact
	BusinessEventsProcessed int64
	CriticalEventsProcessed int64
	ValueGenerated          float64
	CostIncurred            float64
	ROI                     float64
	CustomerImpact          float64

	// Optimization metrics
	OptimizationApplied    int64
	OptimizationGains      float64
	EfficiencyImprovement  float64
	CostReduction          float64
	PerformanceImprovement float64
	QualityImprovement     float64

	// Compliance metrics
	ComplianceChecks     int64
	ComplianceViolations int64
	AuditEvents          int64
	SecurityEvents       int64
	PrivacyEvents        int64
	ComplianceRate       float64
}

// SchedulingStreamingPerformanceMetrics represents performance metrics
type SchedulingStreamingPerformanceMetrics struct {
	// Latency metrics
	EndToEndLatency   time.Duration
	ProcessingLatency time.Duration
	QueueingLatency   time.Duration
	NetworkLatency    time.Duration
	StorageLatency    time.Duration
	TotalLatency      time.Duration

	// Latency percentiles
	P50Latency  time.Duration
	P90Latency  time.Duration
	P95Latency  time.Duration
	P99Latency  time.Duration
	P999Latency time.Duration
	MaxLatency  time.Duration

	// Throughput metrics
	EventsPerSecond       float64
	BytesPerSecond        float64
	RecordsPerSecond      float64
	TransactionsPerSecond float64
	OperationsPerSecond   float64
	PeakThroughput        float64

	// Efficiency metrics
	ProcessingEfficiency float64
	ResourceEfficiency   float64
	CostEfficiency       float64
	EnergyEfficiency     float64
	SpaceEfficiency      float64
	TimeEfficiency       float64

	// Scalability metrics
	ScalabilityIndex        float64
	ElasticityScore         float64
	LoadBalancingEfficiency float64
	PartitioningEfficiency  float64
	ShardingEfficiency      float64
	ReplicationEfficiency   float64

	// Quality metrics
	DataQualityScore    float64
	ProcessingAccuracy  float64
	DeliveryReliability float64
	ConsistencyScore    float64
	CompletenessScore   float64
	TimelinessScore     float64

	// Optimization metrics
	OptimizationScore float64
	AutomationLevel   float64
	IntelligenceScore float64
	AdaptabilityScore float64
	LearningRate      float64
	ImprovementRate   float64

	// Business metrics
	BusinessValueDelivered float64
	CustomerSatisfaction   float64
	SLAAttainment          float64
	CostPerformanceRatio   float64
	ROIScore               float64
	CompetitiveAdvantage   float64

	// Capacity metrics
	CapacityUtilization float64
	HeadroomAvailable   float64
	ScalingPotential    float64
	GrowthCapability    float64
	BurstCapacity       float64
	SustainedCapacity   float64

	// Reliability metrics
	Availability        float64
	Reliability         float64
	Durability          float64
	FaultTolerance      float64
	RecoverabilityScore float64
	ResilienceScore     float64
}

// JobSchedulingStreamingCollector collects job scheduling streaming metrics
type JobSchedulingStreamingCollector struct {
	client                  JobSchedulingStreamingSLURMClient
	schedulingEvents        *prometheus.Desc
	activeSchedulingStreams *prometheus.Desc
	streamingHealthScore    *prometheus.Desc
	schedulingLatency       *prometheus.Desc
	schedulingThroughput    *prometheus.Desc
	schedulingEfficiency    *prometheus.Desc
	decisionConfidence      *prometheus.Desc
	resourceAllocationScore *prometheus.Desc
	queueOptimization       *prometheus.Desc
	backfillEffectiveness   *prometheus.Desc
	preemptionImpact        *prometheus.Desc
	fairnessScore           *prometheus.Desc
	businessValueDelivered  *prometheus.Desc
	slaCompliance           *prometheus.Desc
	costEfficiency          *prometheus.Desc
	optimizationGains       *prometheus.Desc
	anomalyDetectionRate    *prometheus.Desc
	predictionAccuracy      *prometheus.Desc
	systemUtilization       *prometheus.Desc
	streamConfigMetrics     map[string]*prometheus.Desc
	eventFilterMetrics      map[string]*prometheus.Desc
	subscriptionMetrics     map[string]*prometheus.Desc
	processingStatsMetrics  map[string]*prometheus.Desc
	performanceMetrics      map[string]*prometheus.Desc
}

// NewJobSchedulingStreamingCollector creates a new job scheduling streaming collector
func NewJobSchedulingStreamingCollector(client JobSchedulingStreamingSLURMClient) *JobSchedulingStreamingCollector {
	return &JobSchedulingStreamingCollector{
		client: client,
		schedulingEvents: prometheus.NewDesc(
			"slurm_scheduling_events_total",
			"Total number of job scheduling events processed",
			[]string{"event_type", "decision", "queue", "algorithm"},
			nil,
		),
		activeSchedulingStreams: prometheus.NewDesc(
			"slurm_active_scheduling_streams",
			"Number of active scheduling event streams",
			[]string{"stream_type", "status", "consumer_type"},
			nil,
		),
		streamingHealthScore: prometheus.NewDesc(
			"slurm_scheduling_streaming_health_score",
			"Overall health score of the scheduling streaming system",
			[]string{"component", "status"},
			nil,
		),
		schedulingLatency: prometheus.NewDesc(
			"slurm_scheduling_latency_seconds",
			"Job scheduling decision latency",
			[]string{"queue", "algorithm", "complexity"},
			nil,
		),
		schedulingThroughput: prometheus.NewDesc(
			"slurm_scheduling_throughput_per_second",
			"Number of scheduling decisions per second",
			[]string{"queue", "algorithm"},
			nil,
		),
		schedulingEfficiency: prometheus.NewDesc(
			"slurm_scheduling_efficiency_ratio",
			"Efficiency of scheduling decisions",
			[]string{"queue", "metric_type"},
			nil,
		),
		decisionConfidence: prometheus.NewDesc(
			"slurm_scheduling_decision_confidence",
			"Confidence score of scheduling decisions",
			[]string{"queue", "algorithm"},
			nil,
		),
		resourceAllocationScore: prometheus.NewDesc(
			"slurm_resource_allocation_score",
			"Quality score of resource allocations",
			[]string{"resource_type", "queue"},
			nil,
		),
		queueOptimization: prometheus.NewDesc(
			"slurm_queue_optimization_score",
			"Queue optimization effectiveness",
			[]string{"queue", "optimization_type"},
			nil,
		),
		backfillEffectiveness: prometheus.NewDesc(
			"slurm_backfill_effectiveness_ratio",
			"Effectiveness of backfill scheduling",
			[]string{"queue"},
			nil,
		),
		preemptionImpact: prometheus.NewDesc(
			"slurm_preemption_impact_score",
			"Impact score of job preemptions",
			[]string{"queue", "impact_type"},
			nil,
		),
		fairnessScore: prometheus.NewDesc(
			"slurm_scheduling_fairness_score",
			"Fairness score of scheduling decisions",
			[]string{"queue", "fairness_metric"},
			nil,
		),
		businessValueDelivered: prometheus.NewDesc(
			"slurm_scheduling_business_value",
			"Business value delivered through scheduling",
			[]string{"queue", "value_type"},
			nil,
		),
		slaCompliance: prometheus.NewDesc(
			"slurm_scheduling_sla_compliance_ratio",
			"SLA compliance rate for scheduled jobs",
			[]string{"queue", "sla_type"},
			nil,
		),
		costEfficiency: prometheus.NewDesc(
			"slurm_scheduling_cost_efficiency",
			"Cost efficiency of scheduling decisions",
			[]string{"queue", "cost_type"},
			nil,
		),
		optimizationGains: prometheus.NewDesc(
			"slurm_scheduling_optimization_gains",
			"Gains from scheduling optimizations",
			[]string{"optimization_type", "metric"},
			nil,
		),
		anomalyDetectionRate: prometheus.NewDesc(
			"slurm_scheduling_anomaly_detection_rate",
			"Rate of anomaly detection in scheduling",
			[]string{"anomaly_type", "severity"},
			nil,
		),
		predictionAccuracy: prometheus.NewDesc(
			"slurm_scheduling_prediction_accuracy",
			"Accuracy of scheduling predictions",
			[]string{"prediction_type", "queue"},
			nil,
		),
		systemUtilization: prometheus.NewDesc(
			"slurm_scheduling_system_utilization_ratio",
			"System utilization from scheduling decisions",
			[]string{"resource_type"},
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
func (c *JobSchedulingStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.schedulingEvents
	ch <- c.activeSchedulingStreams
	ch <- c.streamingHealthScore
	ch <- c.schedulingLatency
	ch <- c.schedulingThroughput
	ch <- c.schedulingEfficiency
	ch <- c.decisionConfidence
	ch <- c.resourceAllocationScore
	ch <- c.queueOptimization
	ch <- c.backfillEffectiveness
	ch <- c.preemptionImpact
	ch <- c.fairnessScore
	ch <- c.businessValueDelivered
	ch <- c.slaCompliance
	ch <- c.costEfficiency
	ch <- c.optimizationGains
	ch <- c.anomalyDetectionRate
	ch <- c.predictionAccuracy
	ch <- c.systemUtilization

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
func (c *JobSchedulingStreamingCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Collect streaming configuration metrics
	c.collectStreamingConfiguration(ctx, ch)

	// Collect active streams metrics
	c.collectActiveStreams(ctx, ch)

	// Collect streaming metrics
	c.collectStreamingMetrics(ctx, ch)

	// Collect health status metrics
	c.collectHealthStatus(ctx, ch)

	// Collect event subscription metrics
	c.collectEventSubscriptions(ctx, ch)

	// Collect event filter metrics
	c.collectEventFilters(ctx, ch)

	// Collect processing statistics
	c.collectProcessingStats(ctx, ch)

	// Collect performance metrics
	c.collectPerformanceMetrics(ctx, ch)
}

func (c *JobSchedulingStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetSchedulingStreamingConfiguration(ctx)
	if err != nil {
		return
	}

	// Configuration enabled/disabled metrics
	configurations := map[string]bool{
		"streaming_enabled":       config.StreamingEnabled,
		"compression_enabled":     config.CompressionEnabled,
		"encryption_enabled":      config.EncryptionEnabled,
		"batch_processing":        config.BatchProcessing,
		"parallel_processing":     config.ParallelProcessing,
		"event_filtering":         config.EventFiltering,
		"event_routing":           config.EventRouting,
		"priority_queuing":        config.PriorityQueuing,
		"guaranteed_delivery":     config.GuaranteedDelivery,
		"ordering_guaranteed":     config.OrderingGuaranteed,
		"duplicate_detection":     config.DuplicateDetection,
		"metrics_enabled":         config.MetricsEnabled,
		"tracing_enabled":         config.TracingEnabled,
		"profiling_enabled":       config.ProfilingEnabled,
		"authentication_required": config.AuthenticationRequired,
		"authorization_required":  config.AuthorizationRequired,
		"audit_logging_enabled":   config.AuditLoggingEnabled,
		"event_enrichment":        config.EventEnrichment,
		"correlation_enabled":     config.CorrelationEnabled,
		"adaptive_streaming":      config.AdaptiveStreaming,
		"ml_prediction":           config.MLPrediction,
		"anomaly_detection":       config.AnomalyDetection,
	}

	for name, enabled := range configurations {
		value := 0.0
		if enabled {
			value = 1.0
		}

		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_scheduling_stream_config_"+name,
				"Scheduling streaming configuration for "+name,
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
		"event_buffer_size":         float64(config.EventBufferSize),
		"event_batch_size":          float64(config.EventBatchSize),
		"max_concurrent_streams":    float64(config.MaxConcurrentStreams),
		"compression_level":         float64(config.CompressionLevel),
		"max_retries":               float64(config.MaxRetries),
		"backoff_multiplier":        config.BackoffMultiplier,
		"sampling_rate":             config.SamplingRate,
		"max_memory_usage":          float64(config.MaxMemoryUsage),
		"max_cpu_usage":             config.MaxCPUUsage,
		"max_bandwidth":             float64(config.MaxBandwidth),
		"max_events_per_second":     float64(config.MaxEventsPerSecond),
		"backpressure_threshold":    float64(config.BackpressureThreshold),
		"circuit_breaker_threshold": config.CircuitBreakerThreshold,
	}

	for name, value := range numericConfigs {
		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_scheduling_stream_config_"+name,
				"Scheduling streaming configuration value for "+name,
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

func (c *JobSchedulingStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActiveSchedulingStreams(ctx)
	if err != nil {
		return
	}

	// Count streams by type and status
	streamCounts := make(map[string]int)
	for _, stream := range streams {
		key := stream.StreamType + "_" + stream.Status + "_" + stream.ConsumerType
		streamCounts[key]++

		// Individual stream metrics
		ch <- prometheus.MustNewConstMetric(
			c.schedulingThroughput,
			prometheus.GaugeValue,
			stream.Throughput,
			stream.StreamType, "active",
		)

		ch <- prometheus.MustNewConstMetric(
			c.businessValueDelivered,
			prometheus.GaugeValue,
			stream.BusinessValue,
			stream.StreamType, "stream_value",
		)

		ch <- prometheus.MustNewConstMetric(
			c.costEfficiency,
			prometheus.GaugeValue,
			stream.CostPerEvent,
			stream.StreamType, "per_event",
		)

		ch <- prometheus.MustNewConstMetric(
			c.slaCompliance,
			prometheus.GaugeValue,
			stream.SLACompliance,
			stream.StreamType, "stream",
		)

		ch <- prometheus.MustNewConstMetric(
			c.optimizationGains,
			prometheus.GaugeValue,
			stream.OptimizationGains,
			"stream", stream.StreamID,
		)
	}

	// Aggregate stream counts
	for key, count := range streamCounts {
		parts := strings.Split(key, "_")
		if len(parts) >= 3 {
			ch <- prometheus.MustNewConstMetric(
				c.activeSchedulingStreams,
				prometheus.GaugeValue,
				float64(count),
				parts[0], parts[1], parts[2],
			)
		}
	}
}

func (c *JobSchedulingStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetSchedulingStreamingMetrics(ctx)
	if err != nil {
		return
	}

	// Core streaming metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingEvents,
		prometheus.CounterValue,
		float64(metrics.TotalEventsProcessed),
		"all", "all", "all", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingThroughput,
		prometheus.GaugeValue,
		metrics.EventsPerSecond,
		"system", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingLatency,
		prometheus.GaugeValue,
		metrics.EventProcessingLatency.Seconds(),
		"system", "overall", "average",
	)

	// System utilization metrics
	ch <- prometheus.MustNewConstMetric(
		c.systemUtilization,
		prometheus.GaugeValue,
		metrics.CPUUtilization,
		"cpu",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemUtilization,
		prometheus.GaugeValue,
		metrics.MemoryUtilization,
		"memory",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemUtilization,
		prometheus.GaugeValue,
		metrics.NetworkUtilization,
		"network",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemUtilization,
		prometheus.GaugeValue,
		metrics.DiskUtilization,
		"disk",
	)

	// Quality and efficiency metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		metrics.ResourceEfficiency,
		"system", "resource",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		metrics.EfficiencyScore,
		"system", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		metrics.PerformanceScore,
		"system", "performance",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessValueDelivered,
		prometheus.GaugeValue,
		metrics.TotalBusinessValue,
		"system", "total",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costEfficiency,
		prometheus.GaugeValue,
		metrics.AverageCostPerEvent,
		"system", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.slaCompliance,
		prometheus.GaugeValue,
		metrics.SLACompliance,
		"system", "overall",
	)

	// Optimization metrics
	ch <- prometheus.MustNewConstMetric(
		c.optimizationGains,
		prometheus.GaugeValue,
		metrics.PotentialSavings,
		"potential", "cost",
	)

	ch <- prometheus.MustNewConstMetric(
		c.optimizationGains,
		prometheus.GaugeValue,
		float64(metrics.OptimizationOpportunities),
		"opportunities", "count",
	)
}

func (c *JobSchedulingStreamingCollector) collectHealthStatus(ctx context.Context, ch chan<- prometheus.Metric) {
	health, err := c.client.GetSchedulingStreamingHealthStatus(ctx)
	if err != nil {
		return
	}

	// Overall health score
	ch <- prometheus.MustNewConstMetric(
		c.streamingHealthScore,
		prometheus.GaugeValue,
		health.HealthScore,
		"overall", health.HealthStatus,
	)

	// Component health scores
	componentHealth := map[string]float64{
		"stream":     health.StreamHealth,
		"processing": health.ProcessingHealth,
		"storage":    health.StorageHealth,
		"network":    health.NetworkHealth,
		"security":   health.SecurityHealth,
		"compliance": health.ComplianceHealth,
	}

	for component, score := range componentHealth {
		ch <- prometheus.MustNewConstMetric(
			c.streamingHealthScore,
			prometheus.GaugeValue,
			score,
			component, "active",
		)
	}

	// Performance health metrics
	performanceHealth := map[string]float64{
		"latency":     health.LatencyHealth,
		"throughput":  health.ThroughputHealth,
		"error_rate":  health.ErrorRateHealth,
		"resource":    health.ResourceHealth,
		"scalability": health.ScalabilityHealth,
		"efficiency":  health.EfficiencyHealth,
	}

	for metric, score := range performanceHealth {
		ch <- prometheus.MustNewConstMetric(
			c.streamingHealthScore,
			prometheus.GaugeValue,
			score,
			"performance_"+metric, "active",
		)
	}

	// Business health metrics
	businessHealth := map[string]float64{
		"sla":      health.SLAHealth,
		"cost":     health.CostHealth,
		"value":    health.ValueHealth,
		"customer": health.CustomerHealth,
		"roi":      health.ROIHealth,
		"impact":   health.ImpactHealth,
	}

	for metric, score := range businessHealth {
		ch <- prometheus.MustNewConstMetric(
			c.streamingHealthScore,
			prometheus.GaugeValue,
			score,
			"business_"+metric, "active",
		)
	}

	// Risk indicators
	riskMetrics := map[string]float64{
		"overall":      health.RiskScore,
		"security":     health.SecurityRisk,
		"operational":  health.OperationalRisk,
		"compliance":   health.ComplianceRisk,
		"financial":    health.FinancialRisk,
		"reputational": health.ReputationalRisk,
	}

	for risk, score := range riskMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.streamingHealthScore,
			prometheus.GaugeValue,
			1.0-score, // Invert for health score
			"risk_"+risk, "active",
		)
	}
}

func (c *JobSchedulingStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetSchedulingEventSubscriptions(ctx)
	if err != nil {
		return
	}

	for _, sub := range subscriptions {
		// Subscription metrics
		// TODO: Remove unused variable
		// labels := []string{sub.SubscriptionType, sub.Status, sub.DeliveryMethod}

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEvents,
			prometheus.CounterValue,
			float64(sub.EventsDelivered),
			"subscription", "delivered", sub.SubscriptionID, sub.SubscriptionType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEvents,
			prometheus.CounterValue,
			float64(sub.EventsFailed),
			"subscription", "failed", sub.SubscriptionID, sub.SubscriptionType,
		)

		if sub.RateLimitEnabled {
			ch <- prometheus.MustNewConstMetric(
				c.schedulingThroughput,
				prometheus.GaugeValue,
				float64(sub.EventsPerSecond),
				"subscription_limit", sub.SubscriptionID,
			)
		}

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEfficiency,
			prometheus.GaugeValue,
			sub.SuccessRate,
			"subscription", sub.SubscriptionID,
		)

		if sub.BillingEnabled {
			ch <- prometheus.MustNewConstMetric(
				c.costEfficiency,
				prometheus.GaugeValue,
				sub.TotalCost,
				"subscription", sub.SubscriptionID,
			)
		}
	}
}

func (c *JobSchedulingStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetSchedulingEventFilters(ctx)
	if err != nil {
		return
	}

	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEvents,
			prometheus.CounterValue,
			float64(filter.MatchCount),
			"filter", "matched", filter.FilterID, filter.FilterType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEvents,
			prometheus.CounterValue,
			float64(filter.RejectCount),
			"filter", "rejected", filter.FilterID, filter.FilterType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.schedulingEfficiency,
			prometheus.GaugeValue,
			filter.Efficiency,
			"filter", filter.FilterID,
		)
	}
}

func (c *JobSchedulingStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetSchedulingEventProcessingStats(ctx)
	if err != nil {
		return
	}

	// Processing metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingEvents,
		prometheus.CounterValue,
		float64(stats.TotalEventsProcessed),
		"processing", "total", "all", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingEvents,
		prometheus.CounterValue,
		float64(stats.SuccessfulProcessing),
		"processing", "successful", "all", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingEvents,
		prometheus.CounterValue,
		float64(stats.FailedProcessing),
		"processing", "failed", "all", "all",
	)

	// Performance metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingLatency,
		prometheus.GaugeValue,
		stats.AverageProcessingTime.Seconds(),
		"processing", "overall", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingLatency,
		prometheus.GaugeValue,
		stats.P95ProcessingTime.Seconds(),
		"processing", "overall", "p95",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingLatency,
		prometheus.GaugeValue,
		stats.P99ProcessingTime.Seconds(),
		"processing", "overall", "p99",
	)

	// Throughput metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingThroughput,
		prometheus.GaugeValue,
		stats.CurrentThroughput,
		"processing", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingThroughput,
		prometheus.GaugeValue,
		stats.PeakThroughput,
		"processing", "peak",
	)

	// Efficiency metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		stats.ResourceEfficiency,
		"processing", "resource",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		stats.QueueEfficiency,
		"processing", "queue",
	)

	// Business impact
	ch <- prometheus.MustNewConstMetric(
		c.businessValueDelivered,
		prometheus.GaugeValue,
		stats.ValueGenerated,
		"processing", "generated",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costEfficiency,
		prometheus.GaugeValue,
		stats.CostIncurred,
		"processing", "incurred",
	)

	// Optimization metrics
	ch <- prometheus.MustNewConstMetric(
		c.optimizationGains,
		prometheus.GaugeValue,
		stats.OptimizationGains,
		"processing", "gains",
	)

	ch <- prometheus.MustNewConstMetric(
		c.optimizationGains,
		prometheus.GaugeValue,
		stats.EfficiencyImprovement,
		"processing", "efficiency",
	)

	// Compliance metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingEfficiency,
		prometheus.GaugeValue,
		stats.ComplianceRate,
		"processing", "compliance",
	)
}

func (c *JobSchedulingStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perf, err := c.client.GetSchedulingStreamingPerformanceMetrics(ctx)
	if err != nil {
		return
	}

	// Latency metrics
	latencyMetrics := map[string]time.Duration{
		"end_to_end": perf.EndToEndLatency,
		"processing": perf.ProcessingLatency,
		"queueing":   perf.QueueingLatency,
		"network":    perf.NetworkLatency,
		"storage":    perf.StorageLatency,
		"p50":        perf.P50Latency,
		"p90":        perf.P90Latency,
		"p95":        perf.P95Latency,
		"p99":        perf.P99Latency,
		"p999":       perf.P999Latency,
	}

	for name, latency := range latencyMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.schedulingLatency,
			prometheus.GaugeValue,
			latency.Seconds(),
			"performance", name, "measured",
		)
	}

	// Throughput metrics
	ch <- prometheus.MustNewConstMetric(
		c.schedulingThroughput,
		prometheus.GaugeValue,
		perf.EventsPerSecond,
		"performance", "events",
	)

	ch <- prometheus.MustNewConstMetric(
		c.schedulingThroughput,
		prometheus.GaugeValue,
		perf.TransactionsPerSecond,
		"performance", "transactions",
	)

	// Efficiency metrics
	efficiencyMetrics := map[string]float64{
		"processing": perf.ProcessingEfficiency,
		"resource":   perf.ResourceEfficiency,
		"cost":       perf.CostEfficiency,
		"energy":     perf.EnergyEfficiency,
		"space":      perf.SpaceEfficiency,
		"time":       perf.TimeEfficiency,
	}

	for name, efficiency := range efficiencyMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.schedulingEfficiency,
			prometheus.GaugeValue,
			efficiency,
			"performance", name,
		)
	}

	// Quality metrics
	ch <- prometheus.MustNewConstMetric(
		c.decisionConfidence,
		prometheus.GaugeValue,
		perf.ProcessingAccuracy,
		"performance", "accuracy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.resourceAllocationScore,
		prometheus.GaugeValue,
		perf.DataQualityScore,
		"performance", "overall",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessValueDelivered,
		prometheus.GaugeValue,
		perf.BusinessValueDelivered,
		"performance", "delivered",
	)

	ch <- prometheus.MustNewConstMetric(
		c.slaCompliance,
		prometheus.GaugeValue,
		perf.SLAAttainment,
		"performance", "attainment",
	)

	// Optimization metrics
	ch <- prometheus.MustNewConstMetric(
		c.optimizationGains,
		prometheus.GaugeValue,
		perf.OptimizationScore,
		"performance", "score",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictionAccuracy,
		prometheus.GaugeValue,
		perf.LearningRate,
		"performance", "learning",
	)

	// Reliability metrics
	ch <- prometheus.MustNewConstMetric(
		c.streamingHealthScore,
		prometheus.GaugeValue,
		perf.Availability,
		"reliability", "availability",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingHealthScore,
		prometheus.GaugeValue,
		perf.ResilienceScore,
		"reliability", "resilience",
	)
}
