// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type QueueStateStreamingSLURMClient interface {
	StreamQueueStateChanges(ctx context.Context) (<-chan QueueStateChangeEvent, error)
	GetQueueStreamingConfiguration(ctx context.Context) (*QueueStreamingConfiguration, error)
	GetActiveQueueStreams(ctx context.Context) ([]*ActiveQueueStream, error)
	GetQueueEventHistory(ctx context.Context, queueID string, duration time.Duration) ([]*QueueEvent, error)
	GetQueueStreamingMetrics(ctx context.Context) (*QueueStreamingMetrics, error)
	GetQueueEventFilters(ctx context.Context) ([]*QueueEventFilter, error)
	ConfigureQueueStreaming(ctx context.Context, config *QueueStreamingConfiguration) error
	GetQueueStreamingHealthStatus(ctx context.Context) (*QueueStreamingHealthStatus, error)
	GetQueueEventSubscriptions(ctx context.Context) ([]*QueueEventSubscription, error)
	ManageQueueEventSubscription(ctx context.Context, subscription *QueueEventSubscription) error
	GetQueueEventProcessingStats(ctx context.Context) (*QueueEventProcessingStats, error)
	GetQueueStreamingPerformanceMetrics(ctx context.Context) (*QueueStreamingPerformanceMetrics, error)
}

type QueueStateChangeEvent struct {
	EventID              string
	QueueID              string
	QueueName            string
	PartitionName        string
	PreviousState        string
	CurrentState         string
	StateChangeTime      time.Time
	StateChangeReason    string
	EventType            string
	Priority             int
	QueueLength          int
	PendingJobs          int
	RunningJobs          int
	CompletedJobs        int
	FailedJobs           int
	CancelledJobs        int
	AvgWaitTime          time.Duration
	MaxWaitTime          time.Duration
	MinWaitTime          time.Duration
	MedianWaitTime       time.Duration
	TotalCPUsPending     int
	TotalMemoryPending   int64
	TotalGPUsPending     int
	TotalCPUsRunning     int
	TotalMemoryRunning   int64
	TotalGPUsRunning     int
	BackfillJobs         int
	BackfillEfficiency   float64
	PreemptedJobs        int
	PreemptionRate       float64
	JobSubmissionRate    float64
	JobCompletionRate    float64
	JobFailureRate       float64
	JobCancellationRate  float64
	ThroughputRate       float64
	UtilizationRate      float64
	EfficiencyScore      float64
	QueuePriority        int
	QueueWeight          float64
	QueueLimits          map[string]interface{}
	QueueQuotas          map[string]interface{}
	ResourceAvailability map[string]interface{}
	CapacityMetrics      map[string]interface{}
	PerformanceMetrics   map[string]float64
	HealthMetrics        map[string]float64
	UserCounts           map[string]int
	AccountCounts        map[string]int
	QoSCounts            map[string]int
	PriorityDistribution map[string]int
	JobSizeDistribution  map[string]int
	DurationDistribution map[string]int
	ResourceDistribution map[string]interface{}
	TimeInQueue          time.Duration
	ProcessingTime       time.Duration
	EventSequence        int64
	CorrelationID        string
	CausationID          string
	EventMetadata        map[string]interface{}
	StreamingSource      string
	TriggerType          string
	AlertLevel           string
	AlertReason          string
	ActionRequired       bool
	RecommendedActions   []string
	ImpactAssessment     string
	BusinessImpact       string
	OperationalImpact    string
	UserImpact           string
	SLAImpact            string
	CostImpact           float64
	RiskLevel            string
	ComplianceStatus     string
	AuditTrail           []string
	SecurityContext      map[string]interface{}
	ConfigurationDrift   bool
	PolicyViolations     []string
	QualityMetrics       map[string]float64
	PredictiveMetrics    map[string]float64
	AnomalyIndicators    []string
	TrendAnalysis        map[string]interface{}
	SeasonalPatterns     map[string]interface{}
	ForecastData         map[string]interface{}
}

type QueueStreamingConfiguration struct {
	StreamingEnabled         bool
	EventBufferSize          int
	EventBatchSize           int
	EventFlushInterval       time.Duration
	FilterCriteria           []string
	IncludedQueueStates      []string
	ExcludedQueueStates      []string
	IncludedPartitions       []string
	ExcludedPartitions       []string
	IncludedQueues           []string
	ExcludedQueues           []string
	EventRetentionPeriod     time.Duration
	MaxConcurrentStreams     int
	StreamingProtocol        string
	CompressionEnabled       bool
	EncryptionEnabled        bool
	AuthenticationRequired   bool
	RateLimitPerSecond       int
	BackpressureThreshold    int
	FailoverEnabled          bool
	DeduplicationEnabled     bool
	EventValidationEnabled   bool
	MetricsCollectionEnabled bool
	DebugLoggingEnabled      bool
	PriorityBasedStreaming   bool
	EventEnrichmentEnabled   bool
	CustomEventHandlers      []string
	StreamingEndpoints       []string
	HealthCheckInterval      time.Duration
	ReconnectionAttempts     int
	ReconnectionDelay        time.Duration
	QueueMonitoring          bool
	BackfillAnalysis         bool
	PreemptionTracking       bool
	ResourceTracking         bool
	PerformanceAnalysis      bool
	TrendAnalysis            bool
	AnomalyDetection         bool
	PredictiveAnalytics      bool
	AlertIntegration         bool
	NotificationChannels     []string
	SLAMonitoring            bool
	ComplianceTracking       bool
	QualityAssurance         bool
	CostTracking             bool
	CapacityPlanning         bool
	LoadBalancing            bool
	OptimizationEnabled      bool
	AutoScaling              bool
	MaintenanceMode          bool
	EmergencyProtocols       bool
	DataRetention            string
	ArchivalPolicy           string
	PrivacySettings          map[string]bool
	AuditLogging             bool
	SecurityMonitoring       bool
	IntegrationSettings      map[string]interface{}
	CustomMetrics            []string
	DashboardIntegration     bool
	ReportingEnabled         bool
}

type ActiveQueueStream struct {
	StreamID               string
	QueueID                string
	QueueName              string
	PartitionName          string
	StreamStartTime        time.Time
	LastEventTime          time.Time
	EventCount             int64
	StreamStatus           string
	StreamType             string
	ConsumerID             string
	ConsumerEndpoint       string
	StreamPriority         int
	BufferedEvents         int
	ProcessedEvents        int64
	FailedEvents           int64
	RetryCount             int
	LastError              string
	StreamMetadata         map[string]interface{}
	Bandwidth              float64
	Latency                time.Duration
	CompressionRatio       float64
	EventRate              float64
	ConnectionQuality      float64
	StreamHealth           string
	LastHeartbeat          time.Time
	BackpressureActive     bool
	FailoverActive         bool
	QueuedEvents           int
	DroppedEvents          int64
	FilterMatches          int64
	ValidationErrors       int64
	EnrichmentFailures     int64
	DeliveryAttempts       int64
	AckRate                float64
	ResourceUsage          map[string]float64
	ConfigurationHash      string
	SecurityContext        string
	ComplianceFlags        []string
	PerformanceProfile     string
	QualityScore           float64
	AlertsGenerated        int64
	ActionsTriggered       int64
	OptimizationsSuggested int64
	PredictionsAccuracy    float64
	AnomaliesDetected      int64
	TrendsIdentified       int64
	CostImpact             float64
	BusinessValue          float64
	UserSatisfaction       float64
	SLACompliance          float64
	SystemImpact           float64
}

type QueueEvent struct {
	EventID            string
	QueueID            string
	EventType          string
	EventTime          time.Time
	EventData          map[string]interface{}
	EventSource        string
	EventPriority      int
	ProcessingDelay    time.Duration
	EventSize          int64
	EventVersion       string
	CorrelationID      string
	CausationID        string
	Metadata           map[string]interface{}
	Tags               []string
	Fingerprint        string
	ProcessedBy        string
	ProcessedTime      time.Time
	ValidationStatus   string
	EnrichmentData     map[string]interface{}
	DeliveryAttempts   int
	DeliveryStatus     string
	AcknowledgedAt     *time.Time
	ExpiresAt          *time.Time
	RetryPolicy        string
	ErrorDetails       string
	ImpactLevel        string
	ResolutionTime     *time.Time
	ResolutionBy       string
	ResolutionNotes    string
	EscalationLevel    int
	RootCause          string
	RelatedEvents      []string
	Dependencies       []string
	Notifications      []string
	Workflows          []string
	ComplianceData     map[string]interface{}
	SecurityContext    map[string]interface{}
	QualityMetrics     map[string]float64
	PerformanceData    map[string]interface{}
	BusinessContext    string
	OperationalContext string
	TechnicalContext   string
	UserContext        string
	AlertData          map[string]interface{}
	ActionData         map[string]interface{}
	OptimizationData   map[string]interface{}
	PredictionData     map[string]interface{}
	AnomalyData        map[string]interface{}
	TrendData          map[string]interface{}
	CostData           map[string]interface{}
}

type QueueStreamingMetrics struct {
	TotalStreams              int64
	ActiveStreams             int64
	PausedStreams             int64
	FailedStreams             int64
	EventsPerSecond           float64
	AverageEventLatency       time.Duration
	MaxEventLatency           time.Duration
	MinEventLatency           time.Duration
	TotalEventsProcessed      int64
	TotalEventsDropped        int64
	TotalEventsFailed         int64
	AverageStreamDuration     time.Duration
	MaxStreamDuration         time.Duration
	TotalBandwidthUsed        float64
	CompressionEfficiency     float64
	DeduplicationRate         float64
	ErrorRate                 float64
	SuccessRate               float64
	BackpressureOccurrences   int64
	FailoverOccurrences       int64
	ReconnectionAttempts      int64
	MemoryUsage               int64
	CPUUsage                  float64
	NetworkUsage              float64
	DiskUsage                 int64
	CacheHitRate              float64
	QueueDepth                int64
	ProcessingEfficiency      float64
	StreamingHealth           float64
	QueueCoverage             float64
	StateChangeAccuracy       float64
	PredictionAccuracy        float64
	AlertResponseTime         time.Duration
	OptimizationEffectiveness float64
	QualityScore              float64
	ComplianceScore           float64
	SecurityScore             float64
	PerformanceScore          float64
	BusinessValue             float64
	UserSatisfaction          float64
	CostEfficiency            float64
	ResourceUtilization       float64
	ThroughputOptimization    float64
	LatencyOptimization       float64
	CapacityUtilization       float64
	LoadBalanceEfficiency     float64
	BackfillEfficiency        float64
	PreemptionEfficiency      float64
	SLACompliance             float64
	AnomalyDetectionRate      float64
	TrendPredictionAccuracy   float64
	ForecastAccuracy          float64
	OptimizationImpact        float64
	AutomationEffectiveness   float64
}

type QueueEventFilter struct {
	FilterID             string
	FilterName           string
	FilterType           string
	FilterExpression     string
	IncludePattern       string
	ExcludePattern       string
	QueueStates          []string
	Partitions           []string
	Queues               []string
	EventTypes           []string
	Priorities           []int
	QueueLengthRange     []int
	WaitTimeRange        []time.Duration
	ResourceCriteria     map[string]interface{}
	PerformanceCriteria  map[string]interface{}
	QualityCriteria      map[string]interface{}
	ComplianceCriteria   map[string]interface{}
	BusinessCriteria     map[string]interface{}
	CustomCriteria       map[string]interface{}
	FilterEnabled        bool
	FilterPriority       int
	CreatedBy            string
	CreatedTime          time.Time
	ModifiedBy           string
	ModifiedTime         time.Time
	UsageCount           int64
	LastUsedTime         time.Time
	FilterDescription    string
	FilterTags           []string
	ValidationRules      []string
	MatchCount           int64
	FilteredCount        int64
	ErrorCount           int64
	PerformanceImpact    float64
	MaintenanceWindow    string
	EmergencyBypass      bool
	ComplianceLevel      string
	AuditTrail           []string
	BusinessContext      string
	TechnicalContext     string
	OperationalContext   string
	CostImplications     float64
	RiskAssessment       string
	QualityImpact        float64
	UserImpact           float64
	SystemImpact         float64
	SecurityImplications string
	DataRetention        string
	PrivacySettings      map[string]bool
}

type QueueStreamingHealthStatus struct {
	OverallHealth             string
	ComponentHealth           map[string]string
	LastHealthCheck           time.Time
	HealthCheckDuration       time.Duration
	HealthScore               float64
	CriticalIssues            []string
	WarningIssues             []string
	InfoMessages              []string
	StreamingUptime           time.Duration
	ServiceAvailability       float64
	ResourceUtilization       map[string]float64
	PerformanceMetrics        map[string]float64
	ErrorSummary              map[string]int64
	HealthTrends              map[string]float64
	PredictedIssues           []string
	RecommendedActions        []string
	SystemCapacity            map[string]float64
	AlertThresholds           map[string]float64
	SLACompliance             map[string]float64
	DependencyStatus          map[string]string
	ConfigurationValid        bool
	SecurityStatus            string
	BackupStatus              string
	MonitoringEnabled         bool
	LoggingEnabled            bool
	MaintenanceSchedule       []string
	CapacityForecasts         map[string]float64
	RiskIndicators            map[string]float64
	ComplianceMetrics         map[string]float64
	QualityMetrics            map[string]float64
	PerformanceBaselines      map[string]float64
	AnomalyDetectors          map[string]bool
	AutomationStatus          map[string]bool
	IntegrationHealth         map[string]string
	BusinessMetrics           map[string]float64
	UserExperience            map[string]float64
	CostMetrics               map[string]float64
	EfficiencyMetrics         map[string]float64
	OptimizationOpportunities []string
	TrendAnalysis             map[string]interface{}
	PredictiveInsights        map[string]interface{}
	RecommendationEngine      map[string]interface{}
	AlertingEffectiveness     float64
	ResponseTimeMetrics       map[string]float64
	EscalationEffectiveness   float64
	ResolutionRates           map[string]float64
	PreventativeActions       []string
}

type QueueEventSubscription struct {
	SubscriptionID          string
	SubscriberName          string
	SubscriberEndpoint      string
	SubscriptionType        string
	EventTypes              []string
	FilterCriteria          string
	DeliveryMethod          string
	DeliveryFormat          string
	SubscriptionStatus      string
	CreatedTime             time.Time
	LastDeliveryTime        time.Time
	DeliveryCount           int64
	FailedDeliveries        int64
	RetryPolicy             string
	MaxRetries              int
	RetryDelay              time.Duration
	ExpirationTime          *time.Time
	Priority                int
	BatchDelivery           bool
	BatchSize               int
	BatchTimeout            time.Duration
	CompressionEnabled      bool
	EncryptionEnabled       bool
	AuthenticationToken     string
	CallbackURL             string
	ErrorHandling           string
	DeliveryGuarantee       string
	Metadata                map[string]interface{}
	Tags                    []string
	SubscriberContact       string
	BusinessContext         string
	TechnicalContext        string
	OperationalContext      string
	UsageQuota              int64
	UsedQuota               int64
	BandwidthLimit          float64
	CostCenter              string
	ServiceLevel            string
	MaintenanceWindow       string
	EmergencyContacts       []string
	EscalationProcedure     string
	ComplianceRequirements  []string
	QualityRequirements     []string
	PerformanceRequirements []string
	SecurityRequirements    []string
	AuditSettings           map[string]bool
	DataRetention           string
	PrivacySettings         map[string]bool
	IntegrationSettings     map[string]interface{}
	CustomSettings          map[string]interface{}
	AlertingEnabled         bool
	MonitoringEnabled       bool
	ReportingEnabled        bool
	DashboardEnabled        bool
	AnalyticsEnabled        bool
	OptimizationEnabled     bool
	PredictiveEnabled       bool
	AnomalyDetectionEnabled bool
	TrendAnalysisEnabled    bool
	ForecastingEnabled      bool
	RecommendationEnabled   bool
	AutomationEnabled       bool
}

type QueueEventProcessingStats struct {
	ProcessingStartTime   time.Time
	TotalEventsReceived   int64
	TotalEventsProcessed  int64
	TotalEventsFiltered   int64
	TotalEventsDropped    int64
	TotalProcessingTime   time.Duration
	AverageProcessingTime time.Duration
	MaxProcessingTime     time.Duration
	MinProcessingTime     time.Duration
	ProcessingThroughput  float64
	ErrorRate             float64
	SuccessRate           float64
	FilterEfficiency      float64
	ValidationErrors      int64
	EnrichmentErrors      int64
	DeliveryErrors        int64
	TransformationErrors  int64
	SerializationErrors   int64
	NetworkErrors         int64
	AuthenticationErrors  int64
	AuthorizationErrors   int64
	RateLimitExceeded     int64
	BackpressureEvents    int64
	CircuitBreakerTrips   int64
	RetryAttempts         int64
	DeadLetterEvents      int64
	DuplicateEvents       int64
	OutOfOrderEvents      int64
	LateArrivingEvents    int64
	ProcessingQueues      map[string]int64
	WorkerStatistics      map[string]interface{}
	ResourceUtilization   map[string]float64
	PerformanceCounters   map[string]int64
	QueueStateAccuracy    float64
	EventCorrelationRate  float64
	AnomalyDetectionRate  float64
	TrendDetectionRate    float64
	PredictionAccuracy    float64
	OptimizationImpact    float64
	AlertGenerationRate   float64
	ActionExecutionRate   float64
	ComplianceChecks      int64
	QualityChecks         int64
	SecurityScans         int64
	PerformanceAnalysis   int64
	BusinessAnalysis      int64
	CostAnalysis          int64
	UserAnalysis          int64
	SystemAnalysis        int64
	DataQualityScore      float64
	SystemLoadImpact      float64
	BusinessImpact        float64
	UserImpact            float64
	OperationalImpact     float64
	FinancialImpact       float64
	StrategicImpact       float64
	CompetitiveImpact     float64
	RiskImpact            float64
	ComplianceImpact      float64
	QualityImpact         float64
	PerformanceImpact     float64
	EfficiencyImpact      float64
	ProductivityImpact    float64
	InnovationImpact      float64
	SustainabilityImpact  float64
}

type QueueStreamingPerformanceMetrics struct {
	Throughput                float64
	Latency                   time.Duration
	P50Latency                time.Duration
	P95Latency                time.Duration
	P99Latency                time.Duration
	MaxLatency                time.Duration
	MessageRate               float64
	ByteRate                  float64
	ErrorRate                 float64
	SuccessRate               float64
	AvailabilityPercentage    float64
	UptimePercentage          float64
	CPUUtilization            float64
	MemoryUtilization         float64
	NetworkUtilization        float64
	DiskUtilization           float64
	ConnectionCount           int64
	ActiveConnections         int64
	IdleConnections           int64
	FailedConnections         int64
	ConnectionPoolSize        int64
	QueueDepth                int64
	BufferUtilization         float64
	GCPressure                float64
	GCFrequency               float64
	GCDuration                time.Duration
	HeapSize                  int64
	ThreadCount               int64
	ContextSwitches           int64
	SystemCalls               int64
	PageFaults                int64
	CacheMisses               int64
	BranchMispredictions      int64
	InstructionsPerSecond     float64
	CyclesPerInstruction      float64
	PerformanceScore          float64
	QueueCoverageEfficiency   float64
	StateDetectionAccuracy    float64
	EventProcessingSpeed      float64
	AlertLatency              time.Duration
	ResponseTime              time.Duration
	ResolutionTime            time.Duration
	OptimizationEffectiveness float64
	PredictionAccuracy        float64
	AnomalyDetectionAccuracy  float64
	TrendAnalysisAccuracy     float64
	ForecastAccuracy          float64
	RecommendationAccuracy    float64
	AutomationEffectiveness   float64
	BusinessValue             float64
	CostEffectiveness         float64
	QualityScore              float64
	ComplianceScore           float64
	SecurityScore             float64
	UserSatisfactionScore     float64
	SystemEfficiency          float64
	ResourceOptimization      float64
	CapacityUtilization       float64
	LoadBalancingEfficiency   float64
	ThroughputOptimization    float64
	LatencyOptimization       float64
	ErrorReduction            float64
	AvailabilityImprovement   float64
	ReliabilityImprovement    float64
	ScalabilityImprovement    float64
	MaintainabilityScore      float64
	MonitorabilityScore       float64
	ObservabilityScore        float64
	DebuggabilityScore        float64
	TestabilityScore          float64
	ExtensibilityScore        float64
	PortabilityScore          float64
	UsabilityScore            float64
	AccessibilityScore        float64
	InteroperabilityScore     float64
	CompatibilityScore        float64
	StandardsComplianceScore  float64
	BestPracticesScore        float64
	InnovationScore           float64
	SustainabilityScore       float64
	EthicsScore               float64
	TransparencyScore         float64
	AccountabilityScore       float64
	ResponsivenessScore       float64
	AdaptabilityScore         float64
	ResilienceScore           float64
	RobustnessScore           float64
	ReliabilityScore          float64
	ConsistencyScore          float64
	PredictabilityScore       float64
	StabilityScore            float64
	MaturityScore             float64
	EvolutionScore            float64
	GrowthScore               float64
	ImpactScore               float64
}

type QueueStateStreamingCollector struct {
	client QueueStateStreamingSLURMClient
	mutex  sync.RWMutex

	// Queue state event metrics
	queueEventsTotal     *prometheus.CounterVec
	queueEventRate       *prometheus.GaugeVec
	queueEventLatency    *prometheus.HistogramVec
	queueEventSize       *prometheus.HistogramVec
	queueEventsDropped   *prometheus.CounterVec
	queueEventsFailed    *prometheus.CounterVec
	queueEventsProcessed *prometheus.CounterVec

	// Queue state metrics
	queueStateChanges  *prometheus.CounterVec
	queueLength        *prometheus.GaugeVec
	queueWaitTime      *prometheus.HistogramVec
	queueThroughput    *prometheus.GaugeVec
	queueUtilization   *prometheus.GaugeVec
	queueEfficiency    *prometheus.GaugeVec
	backfillJobs       *prometheus.CounterVec
	backfillEfficiency *prometheus.GaugeVec
	preemptedJobs      *prometheus.CounterVec
	preemptionRate     *prometheus.GaugeVec

	// Job flow metrics
	jobSubmissionRate   *prometheus.GaugeVec
	jobCompletionRate   *prometheus.GaugeVec
	jobFailureRate      *prometheus.GaugeVec
	jobCancellationRate *prometheus.GaugeVec
	pendingJobs         *prometheus.GaugeVec
	runningJobs         *prometheus.GaugeVec
	completedJobs       *prometheus.CounterVec
	failedJobs          *prometheus.CounterVec
	cancelledJobs       *prometheus.CounterVec

	// Resource metrics
	pendingCPUs         *prometheus.GaugeVec
	pendingMemory       *prometheus.GaugeVec
	pendingGPUs         *prometheus.GaugeVec
	runningCPUs         *prometheus.GaugeVec
	runningMemory       *prometheus.GaugeVec
	runningGPUs         *prometheus.GaugeVec
	resourceUtilization *prometheus.GaugeVec
	capacityMetrics     *prometheus.GaugeVec

	// Stream metrics
	activeQueueStreams       *prometheus.GaugeVec
	queueStreamDuration      *prometheus.HistogramVec
	queueStreamBandwidth     *prometheus.GaugeVec
	queueStreamLatency       *prometheus.GaugeVec
	queueStreamHealth        *prometheus.GaugeVec
	queueStreamBackpressure  *prometheus.CounterVec
	queueStreamFailover      *prometheus.CounterVec
	queueStreamReconnections *prometheus.CounterVec

	// Configuration metrics
	streamingEnabled      *prometheus.GaugeVec
	maxConcurrentStreams  *prometheus.GaugeVec
	eventBufferSize       *prometheus.GaugeVec
	eventBatchSize        *prometheus.GaugeVec
	rateLimitPerSecond    *prometheus.GaugeVec
	backpressureThreshold *prometheus.GaugeVec

	// Performance metrics
	streamingThroughput   *prometheus.GaugeVec
	streamingCPUUsage     *prometheus.GaugeVec
	streamingMemoryUsage  *prometheus.GaugeVec
	streamingNetworkUsage *prometheus.GaugeVec
	streamingDiskUsage    *prometheus.GaugeVec
	compressionEfficiency *prometheus.GaugeVec
	deduplicationRate     *prometheus.GaugeVec
	cacheHitRate          *prometheus.GaugeVec
	queueDepth            *prometheus.GaugeVec
	processingEfficiency  *prometheus.GaugeVec

	// Health metrics
	streamingHealthScore *prometheus.GaugeVec
	serviceAvailability  *prometheus.GaugeVec
	streamingUptime      *prometheus.CounterVec
	criticalIssues       *prometheus.GaugeVec
	warningIssues        *prometheus.GaugeVec
	healthCheckDuration  *prometheus.HistogramVec
	slaCompliance        *prometheus.GaugeVec

	// Subscription metrics
	eventSubscriptionsTotal *prometheus.CounterVec
	subscriptionDeliveries  *prometheus.CounterVec
	subscriptionFailures    *prometheus.CounterVec
	subscriptionRetries     *prometheus.CounterVec
	subscriptionLatency     *prometheus.HistogramVec

	// Filter metrics
	eventFiltersActive   *prometheus.GaugeVec
	filterMatchCount     *prometheus.CounterVec
	filterProcessingTime *prometheus.HistogramVec
	filterEfficiency     *prometheus.GaugeVec

	// Error metrics
	streamingErrors      *prometheus.CounterVec
	validationErrors     *prometheus.CounterVec
	enrichmentErrors     *prometheus.CounterVec
	deliveryErrors       *prometheus.CounterVec
	transformationErrors *prometheus.CounterVec
	networkErrors        *prometheus.CounterVec
	authenticationErrors *prometheus.CounterVec

	// Queue-specific metrics
	queueCoverage             *prometheus.GaugeVec
	stateChangeAccuracy       *prometheus.GaugeVec
	predictionAccuracy        *prometheus.GaugeVec
	optimizationEffectiveness *prometheus.GaugeVec
	qualityScore              *prometheus.GaugeVec
	complianceScore           *prometheus.GaugeVec
	businessValue             *prometheus.GaugeVec
	userSatisfaction          *prometheus.GaugeVec
	costEfficiency            *prometheus.GaugeVec
}

func NewQueueStateStreamingCollector(client QueueStateStreamingSLURMClient) *QueueStateStreamingCollector {
	return &QueueStateStreamingCollector{
		client: client,

		// Queue state event metrics
		queueEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_events_total",
				Help: "Total number of queue events processed",
			},
			[]string{"event_type", "queue_state", "partition", "queue_name"},
		),
		queueEventRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_event_rate",
				Help: "Rate of queue events per second",
			},
			[]string{"event_type"},
		),
		queueEventLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_event_latency_seconds",
				Help:    "Latency of queue event processing",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
			},
			[]string{"event_type", "processing_stage"},
		),
		queueEventSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_event_size_bytes",
				Help:    "Size of queue events in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			},
			[]string{"event_type"},
		),
		queueEventsDropped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_events_dropped_total",
				Help: "Total number of queue events dropped",
			},
			[]string{"reason", "event_type"},
		),
		queueEventsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_events_failed_total",
				Help: "Total number of queue events that failed processing",
			},
			[]string{"error_type", "event_type"},
		),
		queueEventsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_events_processed_total",
				Help: "Total number of queue events successfully processed",
			},
			[]string{"event_type", "processing_stage"},
		),

		// Queue state metrics
		queueStateChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_state_changes_total",
				Help: "Total number of queue state changes",
			},
			[]string{"queue_name", "from_state", "to_state", "partition"},
		),
		queueLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_length",
				Help: "Current length of job queues",
			},
			[]string{"queue_name", "partition", "job_state"},
		),
		queueWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_wait_time_seconds",
				Help:    "Distribution of job wait times in queues",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"queue_name", "partition", "wait_type"},
		),
		queueThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_throughput_jobs_per_second",
				Help: "Queue throughput in jobs per second",
			},
			[]string{"queue_name", "partition", "throughput_type"},
		),
		queueUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_utilization_ratio",
				Help: "Queue utilization ratio",
			},
			[]string{"queue_name", "partition", "resource_type"},
		),
		queueEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_efficiency_score",
				Help: "Queue efficiency score",
			},
			[]string{"queue_name", "partition", "efficiency_type"},
		),
		backfillJobs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_backfill_jobs_total",
				Help: "Total number of backfilled jobs",
			},
			[]string{"queue_name", "partition"},
		),
		backfillEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_backfill_efficiency",
				Help: "Queue backfill efficiency",
			},
			[]string{"queue_name", "partition"},
		),
		preemptedJobs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_preempted_jobs_total",
				Help: "Total number of preempted jobs",
			},
			[]string{"queue_name", "partition", "preemption_reason"},
		),
		preemptionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_preemption_rate",
				Help: "Queue preemption rate",
			},
			[]string{"queue_name", "partition"},
		),

		// Job flow metrics
		jobSubmissionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_job_submission_rate",
				Help: "Job submission rate to queues",
			},
			[]string{"queue_name", "partition"},
		),
		jobCompletionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_job_completion_rate",
				Help: "Job completion rate from queues",
			},
			[]string{"queue_name", "partition"},
		),
		jobFailureRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_job_failure_rate",
				Help: "Job failure rate in queues",
			},
			[]string{"queue_name", "partition"},
		),
		jobCancellationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_job_cancellation_rate",
				Help: "Job cancellation rate in queues",
			},
			[]string{"queue_name", "partition"},
		),
		pendingJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_pending_jobs",
				Help: "Number of pending jobs in queues",
			},
			[]string{"queue_name", "partition"},
		),
		runningJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_running_jobs",
				Help: "Number of running jobs from queues",
			},
			[]string{"queue_name", "partition"},
		),
		completedJobs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_completed_jobs_total",
				Help: "Total number of completed jobs from queues",
			},
			[]string{"queue_name", "partition"},
		),
		failedJobs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_failed_jobs_total",
				Help: "Total number of failed jobs from queues",
			},
			[]string{"queue_name", "partition"},
		),
		cancelledJobs: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_cancelled_jobs_total",
				Help: "Total number of cancelled jobs from queues",
			},
			[]string{"queue_name", "partition"},
		),

		// Resource metrics
		pendingCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_pending_cpus",
				Help: "Number of CPUs requested by pending jobs",
			},
			[]string{"queue_name", "partition"},
		),
		pendingMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_pending_memory_bytes",
				Help: "Amount of memory requested by pending jobs",
			},
			[]string{"queue_name", "partition"},
		),
		pendingGPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_pending_gpus",
				Help: "Number of GPUs requested by pending jobs",
			},
			[]string{"queue_name", "partition"},
		),
		runningCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_running_cpus",
				Help: "Number of CPUs used by running jobs",
			},
			[]string{"queue_name", "partition"},
		),
		runningMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_running_memory_bytes",
				Help: "Amount of memory used by running jobs",
			},
			[]string{"queue_name", "partition"},
		),
		runningGPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_running_gpus",
				Help: "Number of GPUs used by running jobs",
			},
			[]string{"queue_name", "partition"},
		),
		resourceUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_resource_utilization",
				Help: "Queue resource utilization percentage",
			},
			[]string{"queue_name", "partition", "resource_type"},
		),
		capacityMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_capacity_metrics",
				Help: "Queue capacity metrics",
			},
			[]string{"queue_name", "partition", "capacity_type"},
		),

		// Stream metrics
		activeQueueStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_active_queue_streams",
				Help: "Number of active queue streams",
			},
			[]string{"stream_type", "stream_status"},
		),
		queueStreamDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_stream_duration_seconds",
				Help:    "Duration of queue streams",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"stream_type"},
		),
		queueStreamBandwidth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_stream_bandwidth_bytes_per_second",
				Help: "Bandwidth usage of queue streams",
			},
			[]string{"stream_id", "consumer_id"},
		),
		queueStreamLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_stream_latency_seconds",
				Help: "Latency of queue streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		queueStreamHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_stream_health_score",
				Help: "Health score of queue streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		queueStreamBackpressure: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_stream_backpressure_total",
				Help: "Total backpressure events in queue streams",
			},
			[]string{"stream_id"},
		),
		queueStreamFailover: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_stream_failover_total",
				Help: "Total failover events in queue streams",
			},
			[]string{"stream_id", "failover_reason"},
		),
		queueStreamReconnections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_stream_reconnections_total",
				Help: "Total reconnection attempts for queue streams",
			},
			[]string{"stream_id", "reconnection_reason"},
		),

		// Configuration metrics
		streamingEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_enabled",
				Help: "Whether queue streaming is enabled (1=enabled, 0=disabled)",
			},
			[]string{},
		),
		maxConcurrentStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_max_concurrent_streams",
				Help: "Maximum number of concurrent queue streams allowed",
			},
			[]string{},
		),
		eventBufferSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_event_buffer_size",
				Help: "Size of the event buffer for queue streaming",
			},
			[]string{},
		),
		eventBatchSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_event_batch_size",
				Help: "Batch size for queue event processing",
			},
			[]string{},
		),
		rateLimitPerSecond: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_rate_limit_per_second",
				Help: "Rate limit for queue streaming per second",
			},
			[]string{},
		),
		backpressureThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_backpressure_threshold",
				Help: "Backpressure threshold for queue streaming",
			},
			[]string{},
		),

		// Performance metrics
		streamingThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_throughput",
				Help: "Throughput of queue streaming system",
			},
			[]string{"metric_type"},
		),
		streamingCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_cpu_usage",
				Help: "CPU usage of queue streaming system",
			},
			[]string{},
		),
		streamingMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_memory_usage_bytes",
				Help: "Memory usage of queue streaming system",
			},
			[]string{},
		),
		streamingNetworkUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_network_usage",
				Help: "Network usage of queue streaming system",
			},
			[]string{},
		),
		streamingDiskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_disk_usage_bytes",
				Help: "Disk usage of queue streaming system",
			},
			[]string{},
		),
		compressionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_compression_efficiency",
				Help: "Compression efficiency of queue streaming",
			},
			[]string{},
		),
		deduplicationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_deduplication_rate",
				Help: "Deduplication rate of queue streaming",
			},
			[]string{},
		),
		cacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_cache_hit_rate",
				Help: "Cache hit rate of queue streaming",
			},
			[]string{},
		),
		queueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_queue_depth",
				Help: "Queue depth of queue streaming system",
			},
			[]string{"queue_type"},
		),
		processingEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_processing_efficiency",
				Help: "Processing efficiency of queue streaming",
			},
			[]string{},
		),

		// Health metrics
		streamingHealthScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_health_score",
				Help: "Overall health score of queue streaming system",
			},
			[]string{},
		),
		serviceAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_service_availability",
				Help: "Service availability of queue streaming system",
			},
			[]string{},
		),
		streamingUptime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_streaming_uptime_seconds_total",
				Help: "Total uptime of queue streaming system",
			},
			[]string{},
		),
		criticalIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_critical_issues",
				Help: "Number of critical issues in queue streaming system",
			},
			[]string{},
		),
		warningIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_warning_issues",
				Help: "Number of warning issues in queue streaming system",
			},
			[]string{},
		),
		healthCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_streaming_health_check_duration_seconds",
				Help:    "Duration of health checks for queue streaming system",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"component"},
		),
		slaCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_sla_compliance",
				Help: "SLA compliance of queue streaming system",
			},
			[]string{"sla_type"},
		),

		// Subscription metrics
		eventSubscriptionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_event_subscriptions_total",
				Help: "Total number of queue event subscriptions",
			},
			[]string{"subscription_type", "status"},
		),
		subscriptionDeliveries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_subscription_deliveries_total",
				Help: "Total number of subscription deliveries",
			},
			[]string{"subscription_id", "delivery_method"},
		),
		subscriptionFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_subscription_failures_total",
				Help: "Total number of subscription delivery failures",
			},
			[]string{"subscription_id", "failure_reason"},
		),
		subscriptionRetries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_subscription_retries_total",
				Help: "Total number of subscription delivery retries",
			},
			[]string{"subscription_id"},
		),
		subscriptionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_subscription_latency_seconds",
				Help:    "Latency of subscription deliveries",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"subscription_id", "delivery_method"},
		),

		// Filter metrics
		eventFiltersActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_event_filters_active",
				Help: "Number of active queue event filters",
			},
			[]string{"filter_type"},
		),
		filterMatchCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_filter_matches_total",
				Help: "Total number of filter matches",
			},
			[]string{"filter_id", "filter_type"},
		),
		filterProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_queue_filter_processing_time_seconds",
				Help:    "Processing time for queue event filters",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"filter_id", "filter_type"},
		),
		filterEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_filter_efficiency",
				Help: "Efficiency of queue event filters",
			},
			[]string{"filter_id"},
		),

		// Error metrics
		streamingErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_streaming_errors_total",
				Help: "Total number of queue streaming errors",
			},
			[]string{"error_type", "component"},
		),
		validationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_validation_errors_total",
				Help: "Total number of queue event validation errors",
			},
			[]string{"validation_type"},
		),
		enrichmentErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_enrichment_errors_total",
				Help: "Total number of queue event enrichment errors",
			},
			[]string{"enrichment_type"},
		),
		deliveryErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_delivery_errors_total",
				Help: "Total number of queue event delivery errors",
			},
			[]string{"delivery_method", "error_code"},
		),
		transformationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_transformation_errors_total",
				Help: "Total number of queue event transformation errors",
			},
			[]string{"transformation_type"},
		),
		networkErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_network_errors_total",
				Help: "Total number of queue streaming network errors",
			},
			[]string{"network_operation"},
		),
		authenticationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_queue_authentication_errors_total",
				Help: "Total number of queue streaming authentication errors",
			},
			[]string{"authentication_method"},
		),

		// Queue-specific metrics
		queueCoverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_streaming_coverage",
				Help: "Percentage of queues covered by streaming",
			},
			[]string{"partition"},
		),
		stateChangeAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_state_change_accuracy",
				Help: "Accuracy of queue state change detection",
			},
			[]string{"state_type"},
		),
		predictionAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_prediction_accuracy",
				Help: "Accuracy of queue predictions",
			},
			[]string{"prediction_type"},
		),
		optimizationEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_optimization_effectiveness",
				Help: "Effectiveness of queue optimizations",
			},
			[]string{"optimization_type"},
		),
		qualityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_quality_score",
				Help: "Queue quality score",
			},
			[]string{"queue_name", "partition"},
		),
		complianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_compliance_score",
				Help: "Queue compliance score",
			},
			[]string{"queue_name", "partition"},
		),
		businessValue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_business_value",
				Help: "Queue business value score",
			},
			[]string{"queue_name", "partition"},
		),
		userSatisfaction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_user_satisfaction",
				Help: "Queue user satisfaction score",
			},
			[]string{"queue_name", "partition"},
		),
		costEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_queue_cost_efficiency",
				Help: "Queue cost efficiency score",
			},
			[]string{"queue_name", "partition"},
		),
	}
}

func (c *QueueStateStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	c.queueEventsTotal.Describe(ch)
	c.queueEventRate.Describe(ch)
	c.queueEventLatency.Describe(ch)
	c.queueEventSize.Describe(ch)
	c.queueEventsDropped.Describe(ch)
	c.queueEventsFailed.Describe(ch)
	c.queueEventsProcessed.Describe(ch)
	c.queueStateChanges.Describe(ch)
	c.queueLength.Describe(ch)
	c.queueWaitTime.Describe(ch)
	c.queueThroughput.Describe(ch)
	c.queueUtilization.Describe(ch)
	c.queueEfficiency.Describe(ch)
	c.backfillJobs.Describe(ch)
	c.backfillEfficiency.Describe(ch)
	c.preemptedJobs.Describe(ch)
	c.preemptionRate.Describe(ch)
	c.jobSubmissionRate.Describe(ch)
	c.jobCompletionRate.Describe(ch)
	c.jobFailureRate.Describe(ch)
	c.jobCancellationRate.Describe(ch)
	c.pendingJobs.Describe(ch)
	c.runningJobs.Describe(ch)
	c.completedJobs.Describe(ch)
	c.failedJobs.Describe(ch)
	c.cancelledJobs.Describe(ch)
	c.pendingCPUs.Describe(ch)
	c.pendingMemory.Describe(ch)
	c.pendingGPUs.Describe(ch)
	c.runningCPUs.Describe(ch)
	c.runningMemory.Describe(ch)
	c.runningGPUs.Describe(ch)
	c.resourceUtilization.Describe(ch)
	c.capacityMetrics.Describe(ch)
	c.activeQueueStreams.Describe(ch)
	c.queueStreamDuration.Describe(ch)
	c.queueStreamBandwidth.Describe(ch)
	c.queueStreamLatency.Describe(ch)
	c.queueStreamHealth.Describe(ch)
	c.queueStreamBackpressure.Describe(ch)
	c.queueStreamFailover.Describe(ch)
	c.queueStreamReconnections.Describe(ch)
	c.streamingEnabled.Describe(ch)
	c.maxConcurrentStreams.Describe(ch)
	c.eventBufferSize.Describe(ch)
	c.eventBatchSize.Describe(ch)
	c.rateLimitPerSecond.Describe(ch)
	c.backpressureThreshold.Describe(ch)
	c.streamingThroughput.Describe(ch)
	c.streamingCPUUsage.Describe(ch)
	c.streamingMemoryUsage.Describe(ch)
	c.streamingNetworkUsage.Describe(ch)
	c.streamingDiskUsage.Describe(ch)
	c.compressionEfficiency.Describe(ch)
	c.deduplicationRate.Describe(ch)
	c.cacheHitRate.Describe(ch)
	c.queueDepth.Describe(ch)
	c.processingEfficiency.Describe(ch)
	c.streamingHealthScore.Describe(ch)
	c.serviceAvailability.Describe(ch)
	c.streamingUptime.Describe(ch)
	c.criticalIssues.Describe(ch)
	c.warningIssues.Describe(ch)
	c.healthCheckDuration.Describe(ch)
	c.slaCompliance.Describe(ch)
	c.eventSubscriptionsTotal.Describe(ch)
	c.subscriptionDeliveries.Describe(ch)
	c.subscriptionFailures.Describe(ch)
	c.subscriptionRetries.Describe(ch)
	c.subscriptionLatency.Describe(ch)
	c.eventFiltersActive.Describe(ch)
	c.filterMatchCount.Describe(ch)
	c.filterProcessingTime.Describe(ch)
	c.filterEfficiency.Describe(ch)
	c.streamingErrors.Describe(ch)
	c.validationErrors.Describe(ch)
	c.enrichmentErrors.Describe(ch)
	c.deliveryErrors.Describe(ch)
	c.transformationErrors.Describe(ch)
	c.networkErrors.Describe(ch)
	c.authenticationErrors.Describe(ch)
	c.queueCoverage.Describe(ch)
	c.stateChangeAccuracy.Describe(ch)
	c.predictionAccuracy.Describe(ch)
	c.optimizationEffectiveness.Describe(ch)
	c.qualityScore.Describe(ch)
	c.complianceScore.Describe(ch)
	c.businessValue.Describe(ch)
	c.userSatisfaction.Describe(ch)
	c.costEfficiency.Describe(ch)
}

func (c *QueueStateStreamingCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	ctx := context.Background()

	c.collectStreamingConfiguration(ctx, ch)
	c.collectActiveStreams(ctx, ch)
	c.collectStreamingMetrics(ctx, ch)
	c.collectStreamingHealth(ctx, ch)
	c.collectEventSubscriptions(ctx, ch)
	c.collectEventFilters(ctx, ch)
	c.collectProcessingStats(ctx, ch)
	c.collectPerformanceMetrics(ctx, ch)

	c.queueEventsTotal.Collect(ch)
	c.queueEventRate.Collect(ch)
	c.queueEventLatency.Collect(ch)
	c.queueEventSize.Collect(ch)
	c.queueEventsDropped.Collect(ch)
	c.queueEventsFailed.Collect(ch)
	c.queueEventsProcessed.Collect(ch)
	c.queueStateChanges.Collect(ch)
	c.queueLength.Collect(ch)
	c.queueWaitTime.Collect(ch)
	c.queueThroughput.Collect(ch)
	c.queueUtilization.Collect(ch)
	c.queueEfficiency.Collect(ch)
	c.backfillJobs.Collect(ch)
	c.backfillEfficiency.Collect(ch)
	c.preemptedJobs.Collect(ch)
	c.preemptionRate.Collect(ch)
	c.jobSubmissionRate.Collect(ch)
	c.jobCompletionRate.Collect(ch)
	c.jobFailureRate.Collect(ch)
	c.jobCancellationRate.Collect(ch)
	c.pendingJobs.Collect(ch)
	c.runningJobs.Collect(ch)
	c.completedJobs.Collect(ch)
	c.failedJobs.Collect(ch)
	c.cancelledJobs.Collect(ch)
	c.pendingCPUs.Collect(ch)
	c.pendingMemory.Collect(ch)
	c.pendingGPUs.Collect(ch)
	c.runningCPUs.Collect(ch)
	c.runningMemory.Collect(ch)
	c.runningGPUs.Collect(ch)
	c.resourceUtilization.Collect(ch)
	c.capacityMetrics.Collect(ch)
	c.activeQueueStreams.Collect(ch)
	c.queueStreamDuration.Collect(ch)
	c.queueStreamBandwidth.Collect(ch)
	c.queueStreamLatency.Collect(ch)
	c.queueStreamHealth.Collect(ch)
	c.queueStreamBackpressure.Collect(ch)
	c.queueStreamFailover.Collect(ch)
	c.queueStreamReconnections.Collect(ch)
	c.streamingEnabled.Collect(ch)
	c.maxConcurrentStreams.Collect(ch)
	c.eventBufferSize.Collect(ch)
	c.eventBatchSize.Collect(ch)
	c.rateLimitPerSecond.Collect(ch)
	c.backpressureThreshold.Collect(ch)
	c.streamingThroughput.Collect(ch)
	c.streamingCPUUsage.Collect(ch)
	c.streamingMemoryUsage.Collect(ch)
	c.streamingNetworkUsage.Collect(ch)
	c.streamingDiskUsage.Collect(ch)
	c.compressionEfficiency.Collect(ch)
	c.deduplicationRate.Collect(ch)
	c.cacheHitRate.Collect(ch)
	c.queueDepth.Collect(ch)
	c.processingEfficiency.Collect(ch)
	c.streamingHealthScore.Collect(ch)
	c.serviceAvailability.Collect(ch)
	c.streamingUptime.Collect(ch)
	c.criticalIssues.Collect(ch)
	c.warningIssues.Collect(ch)
	c.healthCheckDuration.Collect(ch)
	c.slaCompliance.Collect(ch)
	c.eventSubscriptionsTotal.Collect(ch)
	c.subscriptionDeliveries.Collect(ch)
	c.subscriptionFailures.Collect(ch)
	c.subscriptionRetries.Collect(ch)
	c.subscriptionLatency.Collect(ch)
	c.eventFiltersActive.Collect(ch)
	c.filterMatchCount.Collect(ch)
	c.filterProcessingTime.Collect(ch)
	c.filterEfficiency.Collect(ch)
	c.streamingErrors.Collect(ch)
	c.validationErrors.Collect(ch)
	c.enrichmentErrors.Collect(ch)
	c.deliveryErrors.Collect(ch)
	c.transformationErrors.Collect(ch)
	c.networkErrors.Collect(ch)
	c.authenticationErrors.Collect(ch)
	c.queueCoverage.Collect(ch)
	c.stateChangeAccuracy.Collect(ch)
	c.predictionAccuracy.Collect(ch)
	c.optimizationEffectiveness.Collect(ch)
	c.qualityScore.Collect(ch)
	c.complianceScore.Collect(ch)
	c.businessValue.Collect(ch)
	c.userSatisfaction.Collect(ch)
	c.costEfficiency.Collect(ch)
}

func (c *QueueStateStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetQueueStreamingConfiguration(ctx)
	if err != nil {
		log.Printf("Error collecting queue streaming configuration: %v", err)
		return
	}

	if config.StreamingEnabled {
		c.streamingEnabled.WithLabelValues().Set(1)
	} else {
		c.streamingEnabled.WithLabelValues().Set(0)
	}

	c.maxConcurrentStreams.WithLabelValues().Set(float64(config.MaxConcurrentStreams))
	c.eventBufferSize.WithLabelValues().Set(float64(config.EventBufferSize))
	c.eventBatchSize.WithLabelValues().Set(float64(config.EventBatchSize))
	c.rateLimitPerSecond.WithLabelValues().Set(float64(config.RateLimitPerSecond))
	c.backpressureThreshold.WithLabelValues().Set(float64(config.BackpressureThreshold))
}

func (c *QueueStateStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActiveQueueStreams(ctx)
	if err != nil {
		log.Printf("Error collecting active queue streams: %v", err)
		return
	}

	streamCounts := make(map[string]map[string]int)
	for _, stream := range streams {
		if streamCounts[stream.StreamType] == nil {
			streamCounts[stream.StreamType] = make(map[string]int)
		}
		streamCounts[stream.StreamType][stream.StreamStatus]++

		c.queueStreamBandwidth.WithLabelValues(stream.StreamID, stream.ConsumerID).Set(stream.Bandwidth)
		c.queueStreamLatency.WithLabelValues(stream.StreamID, stream.StreamType).Set(stream.Latency.Seconds())

		var healthScore float64
		switch stream.StreamHealth {
		case "healthy":
			healthScore = 1.0
		case "warning":
			healthScore = 0.7
		case "critical":
			healthScore = 0.3
		case "failed":
			healthScore = 0.0
		}
		c.queueStreamHealth.WithLabelValues(stream.StreamID, stream.StreamType).Set(healthScore)

		if stream.BackpressureActive {
			c.queueStreamBackpressure.WithLabelValues(stream.StreamID).Inc()
		}

		if stream.FailoverActive {
			c.queueStreamFailover.WithLabelValues(stream.StreamID, "active").Inc()
		}

		streamDuration := time.Since(stream.StreamStartTime)
		c.queueStreamDuration.WithLabelValues(stream.StreamType).Observe(streamDuration.Seconds())
	}

	for streamType, statusMap := range streamCounts {
		for status, count := range statusMap {
			c.activeQueueStreams.WithLabelValues(streamType, status).Set(float64(count))
		}
	}
}

func (c *QueueStateStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetQueueStreamingMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting queue streaming metrics: %v", err)
		return
	}

	c.queueEventRate.WithLabelValues("all").Set(metrics.EventsPerSecond)
	c.streamingThroughput.WithLabelValues("events_per_second").Set(metrics.EventsPerSecond)
	c.streamingThroughput.WithLabelValues("bytes_per_second").Set(metrics.TotalBandwidthUsed)

	c.streamingCPUUsage.WithLabelValues().Set(metrics.CPUUsage)
	c.streamingMemoryUsage.WithLabelValues().Set(float64(metrics.MemoryUsage))
	c.streamingNetworkUsage.WithLabelValues().Set(metrics.NetworkUsage)
	c.streamingDiskUsage.WithLabelValues().Set(float64(metrics.DiskUsage))

	c.compressionEfficiency.WithLabelValues().Set(metrics.CompressionEfficiency)
	c.deduplicationRate.WithLabelValues().Set(metrics.DeduplicationRate)
	c.cacheHitRate.WithLabelValues().Set(metrics.CacheHitRate)
	c.queueDepth.WithLabelValues("main").Set(float64(metrics.QueueDepth))
	c.processingEfficiency.WithLabelValues().Set(metrics.ProcessingEfficiency)

	c.queueEventsProcessed.WithLabelValues("all", "total").Add(float64(metrics.TotalEventsProcessed))
	c.queueEventsDropped.WithLabelValues("system", "all").Add(float64(metrics.TotalEventsDropped))
	c.queueEventsFailed.WithLabelValues("processing", "all").Add(float64(metrics.TotalEventsFailed))

	// Queue-specific metrics
	c.queueCoverage.WithLabelValues("all").Set(metrics.QueueCoverage)
	c.stateChangeAccuracy.WithLabelValues("all").Set(metrics.StateChangeAccuracy)
	c.predictionAccuracy.WithLabelValues("throughput").Set(metrics.PredictionAccuracy)
	c.optimizationEffectiveness.WithLabelValues("all").Set(metrics.OptimizationEffectiveness)
	c.qualityScore.WithLabelValues("system", "all").Set(metrics.QualityScore)
	c.complianceScore.WithLabelValues("system", "all").Set(metrics.ComplianceScore)
	c.businessValue.WithLabelValues("system", "all").Set(metrics.BusinessValue)
	c.userSatisfaction.WithLabelValues("system", "all").Set(metrics.UserSatisfaction)
	c.costEfficiency.WithLabelValues("system", "all").Set(metrics.CostEfficiency)
}

func (c *QueueStateStreamingCollector) collectStreamingHealth(ctx context.Context, ch chan<- prometheus.Metric) {
	health, err := c.client.GetQueueStreamingHealthStatus(ctx)
	if err != nil {
		log.Printf("Error collecting queue streaming health status: %v", err)
		return
	}

	c.streamingHealthScore.WithLabelValues().Set(health.HealthScore)
	c.serviceAvailability.WithLabelValues().Set(health.ServiceAvailability)
	c.streamingUptime.WithLabelValues().Add(health.StreamingUptime.Seconds())
	c.criticalIssues.WithLabelValues().Set(float64(len(health.CriticalIssues)))
	c.warningIssues.WithLabelValues().Set(float64(len(health.WarningIssues)))
	c.healthCheckDuration.WithLabelValues("overall").Observe(health.HealthCheckDuration.Seconds())

	for slaType, compliance := range health.SLACompliance {
		c.slaCompliance.WithLabelValues(slaType).Set(compliance)
	}
}

func (c *QueueStateStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetQueueEventSubscriptions(ctx)
	if err != nil {
		log.Printf("Error collecting queue event subscriptions: %v", err)
		return
	}

	subscriptionCounts := make(map[string]map[string]int)
	for _, sub := range subscriptions {
		if subscriptionCounts[sub.SubscriptionType] == nil {
			subscriptionCounts[sub.SubscriptionType] = make(map[string]int)
		}
		subscriptionCounts[sub.SubscriptionType][sub.SubscriptionStatus]++

		c.subscriptionDeliveries.WithLabelValues(sub.SubscriptionID, sub.DeliveryMethod).Add(float64(sub.DeliveryCount))
		c.subscriptionFailures.WithLabelValues(sub.SubscriptionID, "delivery_failure").Add(float64(sub.FailedDeliveries))
	}

	for subType, statusMap := range subscriptionCounts {
		for status, count := range statusMap {
			c.eventSubscriptionsTotal.WithLabelValues(subType, status).Add(float64(count))
		}
	}
}

func (c *QueueStateStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetQueueEventFilters(ctx)
	if err != nil {
		log.Printf("Error collecting queue event filters: %v", err)
		return
	}

	filterCounts := make(map[string]int)
	for _, filter := range filters {
		if filter.FilterEnabled {
			filterCounts[filter.FilterType]++
		}

		c.filterMatchCount.WithLabelValues(filter.FilterID, filter.FilterType).Add(float64(filter.MatchCount))

		var efficiency float64
		if filter.MatchCount > 0 {
			efficiency = float64(filter.FilteredCount) / float64(filter.MatchCount)
		}
		c.filterEfficiency.WithLabelValues(filter.FilterID).Set(efficiency)
	}

	for filterType, count := range filterCounts {
		c.eventFiltersActive.WithLabelValues(filterType).Set(float64(count))
	}
}

func (c *QueueStateStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetQueueEventProcessingStats(ctx)
	if err != nil {
		log.Printf("Error collecting queue event processing stats: %v", err)
		return
	}

	c.queueEventsProcessed.WithLabelValues("all", "received").Add(float64(stats.TotalEventsReceived))
	c.queueEventsProcessed.WithLabelValues("all", "processed").Add(float64(stats.TotalEventsProcessed))
	c.queueEventsProcessed.WithLabelValues("all", "filtered").Add(float64(stats.TotalEventsFiltered))
	c.queueEventsDropped.WithLabelValues("processing", "all").Add(float64(stats.TotalEventsDropped))

	c.validationErrors.WithLabelValues("validation").Add(float64(stats.ValidationErrors))
	c.enrichmentErrors.WithLabelValues("enrichment").Add(float64(stats.EnrichmentErrors))
	c.deliveryErrors.WithLabelValues("http", "delivery").Add(float64(stats.DeliveryErrors))
	c.transformationErrors.WithLabelValues("transformation").Add(float64(stats.TransformationErrors))
	c.networkErrors.WithLabelValues("network").Add(float64(stats.NetworkErrors))
	c.authenticationErrors.WithLabelValues("token").Add(float64(stats.AuthenticationErrors))

	for queueType, depth := range stats.ProcessingQueues {
		c.queueDepth.WithLabelValues(queueType).Set(float64(depth))
	}

	// Queue-specific processing stats
	c.stateChangeAccuracy.WithLabelValues("detection").Set(stats.QueueStateAccuracy)
	c.predictionAccuracy.WithLabelValues("trend").Set(stats.PredictionAccuracy)
	c.optimizationEffectiveness.WithLabelValues("impact").Set(stats.OptimizationImpact)
}

func (c *QueueStateStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perfMetrics, err := c.client.GetQueueStreamingPerformanceMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting queue streaming performance metrics: %v", err)
		return
	}

	c.queueEventLatency.WithLabelValues("all", "p50").Observe(perfMetrics.P50Latency.Seconds())
	c.queueEventLatency.WithLabelValues("all", "p95").Observe(perfMetrics.P95Latency.Seconds())
	c.queueEventLatency.WithLabelValues("all", "p99").Observe(perfMetrics.P99Latency.Seconds())
	c.queueEventLatency.WithLabelValues("all", "max").Observe(perfMetrics.MaxLatency.Seconds())

	c.streamingThroughput.WithLabelValues("message_rate").Set(perfMetrics.MessageRate)
	c.streamingThroughput.WithLabelValues("byte_rate").Set(perfMetrics.ByteRate)

	// Queue-specific performance metrics
	c.queueCoverage.WithLabelValues("efficiency").Set(perfMetrics.QueueCoverageEfficiency)
	c.stateChangeAccuracy.WithLabelValues("detection").Set(perfMetrics.StateDetectionAccuracy)
	c.optimizationEffectiveness.WithLabelValues("overall").Set(perfMetrics.OptimizationEffectiveness)
	c.predictionAccuracy.WithLabelValues("overall").Set(perfMetrics.PredictionAccuracy)
	c.qualityScore.WithLabelValues("performance", "all").Set(perfMetrics.QualityScore)
	c.complianceScore.WithLabelValues("performance", "all").Set(perfMetrics.ComplianceScore)
	c.businessValue.WithLabelValues("performance", "all").Set(perfMetrics.BusinessValue)
	c.userSatisfaction.WithLabelValues("performance", "all").Set(perfMetrics.UserSatisfactionScore)
	c.costEfficiency.WithLabelValues("performance", "all").Set(perfMetrics.CostEffectiveness)
}
