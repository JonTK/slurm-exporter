package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type NodeStateStreamingSLURMClient interface {
	StreamNodeStateChanges(ctx context.Context) (<-chan NodeStateChangeEvent, error)
	GetNodeStreamingConfiguration(ctx context.Context) (*NodeStreamingConfiguration, error)
	GetActiveNodeStreams(ctx context.Context) ([]*ActiveNodeStream, error)
	GetNodeEventHistory(ctx context.Context, nodeID string, duration time.Duration) ([]*NodeEvent, error)
	GetNodeStreamingMetrics(ctx context.Context) (*NodeStreamingMetrics, error)
	GetNodeEventFilters(ctx context.Context) ([]*NodeEventFilter, error)
	ConfigureNodeStreaming(ctx context.Context, config *NodeStreamingConfiguration) error
	GetNodeStreamingHealthStatus(ctx context.Context) (*NodeStreamingHealthStatus, error)
	GetNodeEventSubscriptions(ctx context.Context) ([]*NodeEventSubscription, error)
	ManageNodeEventSubscription(ctx context.Context, subscription *NodeEventSubscription) error
	GetNodeEventProcessingStats(ctx context.Context) (*NodeEventProcessingStats, error)
	GetNodeStreamingPerformanceMetrics(ctx context.Context) (*NodeStreamingPerformanceMetrics, error)
}

type NodeStateChangeEvent struct {
	EventID            string
	NodeID             string
	NodeName           string
	PreviousState      string
	CurrentState       string
	StateChangeTime    time.Time
	StateChangeReason  string
	EventType          string
	Priority           int
	PartitionName      string
	TotalCPUs          int
	AllocatedCPUs      int
	FreeCPUs           int
	TotalMemory        int64
	AllocatedMemory    int64
	FreeMemory         int64
	TotalGPUs          int
	AllocatedGPUs      int
	FreeGPUs           int
	Features           []string
	ActiveFeatures     []string
	Architecture       string
	OS                 string
	KernelVersion      string
	LoadAverage        []float64
	BootTime           time.Time
	LastResponseTime   time.Time
	SlurmdPID          int
	SlurmdVersion      string
	SlurmdStartTime    time.Time
	MaintenanceMode    bool
	MaintenanceReason  string
	PowerState         string
	PowerConsumption   float64
	Temperature        float64
	NodeWeight         int
	NodeOwner          string
	NodeComment        string
	NodeLocation       string
	NodeAddress        string
	NodeHostname       string
	NodePort           int
	NodeProtocol       string
	JobList            []string
	StepList           []string
	ReservationList    []string
	EnergyConsumed     float64
	PowerLimit         float64
	ThermalState       string
	NetworkInterfaces  []string
	StorageInfo        map[string]interface{}
	HardwareInfo       map[string]interface{}
	EventMetadata      map[string]interface{}
	StreamingSource    string
	ProcessingTime     time.Duration
	EventSequence      int64
	CorrelationID      string
	CausationID        string
	ClusterID          string
	TriggerType        string
	AutomaticRecovery  bool
	ManualIntervention bool
	ImpactAssessment   string
	RecoveryPlan       string
	EstimatedDowntime  time.Duration
	ActualDowntime     time.Duration
	DiagnosticInfo     map[string]interface{}
	HealthStatus       string
	PerformanceMetrics map[string]float64
	ConfigurationDrift bool
	ComplianceStatus   string
	SecurityEvents     []string
}

type NodeStreamingConfiguration struct {
	StreamingEnabled         bool
	EventBufferSize          int
	EventBatchSize           int
	EventFlushInterval       time.Duration
	FilterCriteria           []string
	IncludedNodeStates       []string
	ExcludedNodeStates       []string
	IncludedPartitions       []string
	ExcludedPartitions       []string
	IncludedNodes            []string
	ExcludedNodes            []string
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
	StateChangeDetection     bool
	PerformanceMonitoring    bool
	ResourceThresholds       map[string]float64
	AlertIntegration         bool
	NotificationChannels     []string
	EventCorrelation         bool
	AnomalyDetection         bool
	PredictiveAnalytics      bool
	MaintenanceMode          bool
	EmergencyProtocols       bool
	DataRetention            string
	ArchivalPolicy           string
	PrivacySettings          map[string]bool
	AuditLogging             bool
	ComplianceReporting      bool
}

type ActiveNodeStream struct {
	StreamID           string
	NodeID             string
	NodeName           string
	PartitionName      string
	StreamStartTime    time.Time
	LastEventTime      time.Time
	EventCount         int64
	StreamStatus       string
	StreamType         string
	ConsumerID         string
	ConsumerEndpoint   string
	StreamPriority     int
	BufferedEvents     int
	ProcessedEvents    int64
	FailedEvents       int64
	RetryCount         int
	LastError          string
	StreamMetadata     map[string]interface{}
	Bandwidth          float64
	Latency            time.Duration
	CompressionRatio   float64
	EventRate          float64
	ConnectionQuality  float64
	StreamHealth       string
	LastHeartbeat      time.Time
	BackpressureActive bool
	FailoverActive     bool
	QueuedEvents       int
	DroppedEvents      int64
	FilterMatches      int64
	ValidationErrors   int64
	EnrichmentFailures int64
	DeliveryAttempts   int64
	AckRate            float64
	ResourceUsage      map[string]float64
	ConfigurationHash  string
	SecurityContext    string
	ComplianceFlags    []string
	PerformanceProfile string
}

type NodeEvent struct {
	EventID          string
	NodeID           string
	EventType        string
	EventTime        time.Time
	EventData        map[string]interface{}
	EventSource      string
	EventPriority    int
	ProcessingDelay  time.Duration
	EventSize        int64
	EventVersion     string
	CorrelationID    string
	CausationID      string
	Metadata         map[string]interface{}
	Tags             []string
	Fingerprint      string
	ProcessedBy      string
	ProcessedTime    time.Time
	ValidationStatus string
	EnrichmentData   map[string]interface{}
	DeliveryAttempts int
	DeliveryStatus   string
	AcknowledgedAt   *time.Time
	ExpiresAt        *time.Time
	RetryPolicy      string
	ErrorDetails     string
	ImpactLevel      string
	ResolutionTime   *time.Time
	ResolutionBy     string
	ResolutionNotes  string
	EscalationLevel  int
	RootCause        string
	RelatedEvents    []string
	Dependencies     []string
	Notifications    []string
	Workflows        []string
	ComplianceData   map[string]interface{}
	SecurityContext  map[string]interface{}
}

type NodeStreamingMetrics struct {
	TotalStreams            int64
	ActiveStreams           int64
	PausedStreams           int64
	FailedStreams           int64
	EventsPerSecond         float64
	AverageEventLatency     time.Duration
	MaxEventLatency         time.Duration
	MinEventLatency         time.Duration
	TotalEventsProcessed    int64
	TotalEventsDropped      int64
	TotalEventsFailed       int64
	AverageStreamDuration   time.Duration
	MaxStreamDuration       time.Duration
	TotalBandwidthUsed      float64
	CompressionEfficiency   float64
	DeduplicationRate       float64
	ErrorRate               float64
	SuccessRate             float64
	BackpressureOccurrences int64
	FailoverOccurrences     int64
	ReconnectionAttempts    int64
	MemoryUsage             int64
	CPUUsage                float64
	NetworkUsage            float64
	DiskUsage               int64
	CacheHitRate            float64
	QueueDepth              int64
	ProcessingEfficiency    float64
	StreamingHealth         float64
	NodeCoverage            float64
	StateChangeAccuracy     float64
	PredictionAccuracy      float64
	AlertResponseTime       time.Duration
	MaintenanceCompliance   float64
	SecurityIncidents       int64
	ComplianceViolations    int64
	PerformanceDegradation  float64
}

type NodeEventFilter struct {
	FilterID          string
	FilterName        string
	FilterType        string
	FilterExpression  string
	IncludePattern    string
	ExcludePattern    string
	NodeStates        []string
	Partitions        []string
	Nodes             []string
	Features          []string
	EventTypes        []string
	Priorities        []int
	TimeRange         string
	ResourceCriteria  map[string]interface{}
	CustomCriteria    map[string]interface{}
	FilterEnabled     bool
	FilterPriority    int
	CreatedBy         string
	CreatedTime       time.Time
	ModifiedBy        string
	ModifiedTime      time.Time
	UsageCount        int64
	LastUsedTime      time.Time
	FilterDescription string
	FilterTags        []string
	ValidationRules   []string
	MatchCount        int64
	FilteredCount     int64
	ErrorCount        int64
	PerformanceImpact float64
	MaintenanceWindow string
	EmergencyBypass   bool
	ComplianceLevel   string
	AuditTrail        []string
	BusinessContext   string
	CostImplications  float64
	RiskAssessment    string
}

type NodeStreamingHealthStatus struct {
	OverallHealth        string
	ComponentHealth      map[string]string
	LastHealthCheck      time.Time
	HealthCheckDuration  time.Duration
	HealthScore          float64
	CriticalIssues       []string
	WarningIssues        []string
	InfoMessages         []string
	StreamingUptime      time.Duration
	ServiceAvailability  float64
	ResourceUtilization  map[string]float64
	PerformanceMetrics   map[string]float64
	ErrorSummary         map[string]int64
	HealthTrends         map[string]float64
	PredictedIssues      []string
	RecommendedActions   []string
	SystemCapacity       map[string]float64
	AlertThresholds      map[string]float64
	SLACompliance        map[string]float64
	DependencyStatus     map[string]string
	ConfigurationValid   bool
	SecurityStatus       string
	BackupStatus         string
	MonitoringEnabled    bool
	LoggingEnabled       bool
	MaintenanceSchedule  []string
	CapacityForecasts    map[string]float64
	RiskIndicators       map[string]float64
	ComplianceMetrics    map[string]float64
	PerformanceBaselines map[string]float64
	AnomalyDetectors     map[string]bool
	AutomationStatus     map[string]bool
	IntegrationHealth    map[string]string
}

type NodeEventSubscription struct {
	SubscriptionID         string
	SubscriberName         string
	SubscriberEndpoint     string
	SubscriptionType       string
	EventTypes             []string
	FilterCriteria         string
	DeliveryMethod         string
	DeliveryFormat         string
	SubscriptionStatus     string
	CreatedTime            time.Time
	LastDeliveryTime       time.Time
	DeliveryCount          int64
	FailedDeliveries       int64
	RetryPolicy            string
	MaxRetries             int
	RetryDelay             time.Duration
	ExpirationTime         *time.Time
	Priority               int
	BatchDelivery          bool
	BatchSize              int
	BatchTimeout           time.Duration
	CompressionEnabled     bool
	EncryptionEnabled      bool
	AuthenticationToken    string
	CallbackURL            string
	ErrorHandling          string
	DeliveryGuarantee      string
	Metadata               map[string]interface{}
	Tags                   []string
	SubscriberContact      string
	BusinessContext        string
	UsageQuota             int64
	UsedQuota              int64
	BandwidthLimit         float64
	CostCenter             string
	ServiceLevel           string
	MaintenanceWindow      string
	EmergencyContacts      []string
	EscalationProcedure    string
	ComplianceRequirements []string
	AuditSettings          map[string]bool
	DataRetention          string
	PrivacySettings        map[string]bool
}

type NodeEventProcessingStats struct {
	ProcessingStartTime    time.Time
	TotalEventsReceived    int64
	TotalEventsProcessed   int64
	TotalEventsFiltered    int64
	TotalEventsDropped     int64
	TotalProcessingTime    time.Duration
	AverageProcessingTime  time.Duration
	MaxProcessingTime      time.Duration
	MinProcessingTime      time.Duration
	ProcessingThroughput   float64
	ErrorRate              float64
	SuccessRate            float64
	FilterEfficiency       float64
	ValidationErrors       int64
	EnrichmentErrors       int64
	DeliveryErrors         int64
	TransformationErrors   int64
	SerializationErrors    int64
	NetworkErrors          int64
	AuthenticationErrors   int64
	AuthorizationErrors    int64
	RateLimitExceeded      int64
	BackpressureEvents     int64
	CircuitBreakerTrips    int64
	RetryAttempts          int64
	DeadLetterEvents       int64
	DuplicateEvents        int64
	OutOfOrderEvents       int64
	LateArrivingEvents     int64
	ProcessingQueues       map[string]int64
	WorkerStatistics       map[string]interface{}
	ResourceUtilization    map[string]float64
	PerformanceCounters    map[string]int64
	StateChangeAccuracy    float64
	EventCorrelationRate   float64
	AnomalyDetectionRate   float64
	MaintenancePredictions int64
	AlertGenerationRate    float64
	ComplianceChecks       int64
	SecurityScans          int64
	DataQualityScore       float64
	SystemLoadImpact       float64
}

type NodeStreamingPerformanceMetrics struct {
	Throughput              float64
	Latency                 time.Duration
	P50Latency              time.Duration
	P95Latency              time.Duration
	P99Latency              time.Duration
	MaxLatency              time.Duration
	MessageRate             float64
	ByteRate                float64
	ErrorRate               float64
	SuccessRate             float64
	AvailabilityPercentage  float64
	UptimePercentage        float64
	CPUUtilization          float64
	MemoryUtilization       float64
	NetworkUtilization      float64
	DiskUtilization         float64
	ConnectionCount         int64
	ActiveConnections       int64
	IdleConnections         int64
	FailedConnections       int64
	ConnectionPoolSize      int64
	QueueDepth              int64
	BufferUtilization       float64
	GCPressure              float64
	GCFrequency             float64
	GCDuration              time.Duration
	HeapSize                int64
	ThreadCount             int64
	ContextSwitches         int64
	SystemCalls             int64
	PageFaults              int64
	CacheMisses             int64
	BranchMispredictions    int64
	InstructionsPerSecond   float64
	CyclesPerInstruction    float64
	PerformanceScore        float64
	NodeCoverageEfficiency  float64
	StateDetectionAccuracy  float64
	EventProcessingSpeed    float64
	AlertLatency            time.Duration
	RecoveryTime            time.Duration
	MaintenanceEfficiency   float64
	CompliancePerformance   float64
	SecurityResponseTime    time.Duration
	PredictiveAccuracy      float64
	AutomationEffectiveness float64
}

type NodeStateStreamingCollector struct {
	client NodeStateStreamingSLURMClient
	mutex  sync.RWMutex

	// Node state event metrics
	nodeEventsTotal     *prometheus.CounterVec
	nodeEventRate       *prometheus.GaugeVec
	nodeEventLatency    *prometheus.HistogramVec
	nodeEventSize       *prometheus.HistogramVec
	nodeEventsDropped   *prometheus.CounterVec
	nodeEventsFailed    *prometheus.CounterVec
	nodeEventsProcessed *prometheus.CounterVec

	// Node state change metrics
	nodeStateChanges      *prometheus.CounterVec
	nodeStateDuration     *prometheus.HistogramVec
	nodeStateTransitions  *prometheus.CounterVec
	nodeDowntime          *prometheus.HistogramVec
	nodeRecoveryTime      *prometheus.HistogramVec
	nodeMaintenanceEvents *prometheus.CounterVec

	// Stream metrics
	activeNodeStreams       *prometheus.GaugeVec
	nodeStreamDuration      *prometheus.HistogramVec
	nodeStreamBandwidth     *prometheus.GaugeVec
	nodeStreamLatency       *prometheus.GaugeVec
	nodeStreamHealth        *prometheus.GaugeVec
	nodeStreamBackpressure  *prometheus.CounterVec
	nodeStreamFailover      *prometheus.CounterVec
	nodeStreamReconnections *prometheus.CounterVec

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

	// Node-specific metrics
	nodeCoverage           *prometheus.GaugeVec
	stateChangeAccuracy    *prometheus.GaugeVec
	predictiveAccuracy     *prometheus.GaugeVec
	maintenanceCompliance  *prometheus.GaugeVec
	securityIncidents      *prometheus.CounterVec
	complianceViolations   *prometheus.CounterVec
	performanceDegradation *prometheus.GaugeVec
}

func NewNodeStateStreamingCollector(client NodeStateStreamingSLURMClient) *NodeStateStreamingCollector {
	return &NodeStateStreamingCollector{
		client: client,

		// Node state event metrics
		nodeEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_events_total",
				Help: "Total number of node events processed",
			},
			[]string{"event_type", "node_state", "partition", "node_name"},
		),
		nodeEventRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_event_rate",
				Help: "Rate of node events per second",
			},
			[]string{"event_type"},
		),
		nodeEventLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_event_latency_seconds",
				Help:    "Latency of node event processing",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
			},
			[]string{"event_type", "processing_stage"},
		),
		nodeEventSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_event_size_bytes",
				Help:    "Size of node events in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			},
			[]string{"event_type"},
		),
		nodeEventsDropped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_events_dropped_total",
				Help: "Total number of node events dropped",
			},
			[]string{"reason", "event_type"},
		),
		nodeEventsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_events_failed_total",
				Help: "Total number of node events that failed processing",
			},
			[]string{"error_type", "event_type"},
		),
		nodeEventsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_events_processed_total",
				Help: "Total number of node events successfully processed",
			},
			[]string{"event_type", "processing_stage"},
		),

		// Node state change metrics
		nodeStateChanges: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_state_changes_total",
				Help: "Total number of node state changes",
			},
			[]string{"node_name", "from_state", "to_state", "partition"},
		),
		nodeStateDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_state_duration_seconds",
				Help:    "Duration of time nodes spend in each state",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"node_state", "partition"},
		),
		nodeStateTransitions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_state_transitions_total",
				Help: "Total number of node state transitions by type",
			},
			[]string{"transition_type", "partition"},
		),
		nodeDowntime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_downtime_seconds",
				Help:    "Duration of node downtime events",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"node_name", "downtime_reason", "partition"},
		),
		nodeRecoveryTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_recovery_time_seconds",
				Help:    "Time taken for nodes to recover from failures",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"node_name", "recovery_type", "partition"},
		),
		nodeMaintenanceEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_maintenance_events_total",
				Help: "Total number of node maintenance events",
			},
			[]string{"node_name", "maintenance_type", "partition"},
		),

		// Stream metrics
		activeNodeStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_active_node_streams",
				Help: "Number of active node streams",
			},
			[]string{"stream_type", "stream_status"},
		),
		nodeStreamDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_stream_duration_seconds",
				Help:    "Duration of node streams",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"stream_type"},
		),
		nodeStreamBandwidth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_stream_bandwidth_bytes_per_second",
				Help: "Bandwidth usage of node streams",
			},
			[]string{"stream_id", "consumer_id"},
		),
		nodeStreamLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_stream_latency_seconds",
				Help: "Latency of node streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		nodeStreamHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_stream_health_score",
				Help: "Health score of node streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		nodeStreamBackpressure: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_stream_backpressure_total",
				Help: "Total backpressure events in node streams",
			},
			[]string{"stream_id"},
		),
		nodeStreamFailover: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_stream_failover_total",
				Help: "Total failover events in node streams",
			},
			[]string{"stream_id", "failover_reason"},
		),
		nodeStreamReconnections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_stream_reconnections_total",
				Help: "Total reconnection attempts for node streams",
			},
			[]string{"stream_id", "reconnection_reason"},
		),

		// Configuration metrics
		streamingEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_enabled",
				Help: "Whether node streaming is enabled (1=enabled, 0=disabled)",
			},
			[]string{},
		),
		maxConcurrentStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_max_concurrent_streams",
				Help: "Maximum number of concurrent node streams allowed",
			},
			[]string{},
		),
		eventBufferSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_event_buffer_size",
				Help: "Size of the event buffer for node streaming",
			},
			[]string{},
		),
		eventBatchSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_event_batch_size",
				Help: "Batch size for node event processing",
			},
			[]string{},
		),
		rateLimitPerSecond: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_rate_limit_per_second",
				Help: "Rate limit for node streaming per second",
			},
			[]string{},
		),
		backpressureThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_backpressure_threshold",
				Help: "Backpressure threshold for node streaming",
			},
			[]string{},
		),

		// Performance metrics
		streamingThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_throughput",
				Help: "Throughput of node streaming system",
			},
			[]string{"metric_type"},
		),
		streamingCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_cpu_usage",
				Help: "CPU usage of node streaming system",
			},
			[]string{},
		),
		streamingMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_memory_usage_bytes",
				Help: "Memory usage of node streaming system",
			},
			[]string{},
		),
		streamingNetworkUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_network_usage",
				Help: "Network usage of node streaming system",
			},
			[]string{},
		),
		streamingDiskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_disk_usage_bytes",
				Help: "Disk usage of node streaming system",
			},
			[]string{},
		),
		compressionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_compression_efficiency",
				Help: "Compression efficiency of node streaming",
			},
			[]string{},
		),
		deduplicationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_deduplication_rate",
				Help: "Deduplication rate of node streaming",
			},
			[]string{},
		),
		cacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_cache_hit_rate",
				Help: "Cache hit rate of node streaming",
			},
			[]string{},
		),
		queueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_queue_depth",
				Help: "Queue depth of node streaming system",
			},
			[]string{"queue_type"},
		),
		processingEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_processing_efficiency",
				Help: "Processing efficiency of node streaming",
			},
			[]string{},
		),

		// Health metrics
		streamingHealthScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_health_score",
				Help: "Overall health score of node streaming system",
			},
			[]string{},
		),
		serviceAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_service_availability",
				Help: "Service availability of node streaming system",
			},
			[]string{},
		),
		streamingUptime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_streaming_uptime_seconds_total",
				Help: "Total uptime of node streaming system",
			},
			[]string{},
		),
		criticalIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_critical_issues",
				Help: "Number of critical issues in node streaming system",
			},
			[]string{},
		),
		warningIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_warning_issues",
				Help: "Number of warning issues in node streaming system",
			},
			[]string{},
		),
		healthCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_streaming_health_check_duration_seconds",
				Help:    "Duration of health checks for node streaming system",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"component"},
		),
		slaCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_sla_compliance",
				Help: "SLA compliance of node streaming system",
			},
			[]string{"sla_type"},
		),

		// Subscription metrics
		eventSubscriptionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_event_subscriptions_total",
				Help: "Total number of node event subscriptions",
			},
			[]string{"subscription_type", "status"},
		),
		subscriptionDeliveries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_subscription_deliveries_total",
				Help: "Total number of subscription deliveries",
			},
			[]string{"subscription_id", "delivery_method"},
		),
		subscriptionFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_subscription_failures_total",
				Help: "Total number of subscription delivery failures",
			},
			[]string{"subscription_id", "failure_reason"},
		),
		subscriptionRetries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_subscription_retries_total",
				Help: "Total number of subscription delivery retries",
			},
			[]string{"subscription_id"},
		),
		subscriptionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_subscription_latency_seconds",
				Help:    "Latency of subscription deliveries",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"subscription_id", "delivery_method"},
		),

		// Filter metrics
		eventFiltersActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_event_filters_active",
				Help: "Number of active node event filters",
			},
			[]string{"filter_type"},
		),
		filterMatchCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_filter_matches_total",
				Help: "Total number of filter matches",
			},
			[]string{"filter_id", "filter_type"},
		),
		filterProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_node_filter_processing_time_seconds",
				Help:    "Processing time for node event filters",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"filter_id", "filter_type"},
		),
		filterEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_filter_efficiency",
				Help: "Efficiency of node event filters",
			},
			[]string{"filter_id"},
		),

		// Error metrics
		streamingErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_streaming_errors_total",
				Help: "Total number of node streaming errors",
			},
			[]string{"error_type", "component"},
		),
		validationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_validation_errors_total",
				Help: "Total number of node event validation errors",
			},
			[]string{"validation_type"},
		),
		enrichmentErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_enrichment_errors_total",
				Help: "Total number of node event enrichment errors",
			},
			[]string{"enrichment_type"},
		),
		deliveryErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_delivery_errors_total",
				Help: "Total number of node event delivery errors",
			},
			[]string{"delivery_method", "error_code"},
		),
		transformationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_transformation_errors_total",
				Help: "Total number of node event transformation errors",
			},
			[]string{"transformation_type"},
		),
		networkErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_network_errors_total",
				Help: "Total number of node streaming network errors",
			},
			[]string{"network_operation"},
		),
		authenticationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_authentication_errors_total",
				Help: "Total number of node streaming authentication errors",
			},
			[]string{"authentication_method"},
		),

		// Node-specific metrics
		nodeCoverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_streaming_coverage",
				Help: "Percentage of nodes covered by streaming",
			},
			[]string{"partition"},
		),
		stateChangeAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_state_change_accuracy",
				Help: "Accuracy of node state change detection",
			},
			[]string{"state_type"},
		),
		predictiveAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_predictive_accuracy",
				Help: "Accuracy of node failure predictions",
			},
			[]string{"prediction_type"},
		),
		maintenanceCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_maintenance_compliance",
				Help: "Compliance with maintenance schedules",
			},
			[]string{"partition"},
		),
		securityIncidents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_security_incidents_total",
				Help: "Total number of node security incidents",
			},
			[]string{"incident_type", "severity"},
		),
		complianceViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_node_compliance_violations_total",
				Help: "Total number of node compliance violations",
			},
			[]string{"violation_type", "severity"},
		),
		performanceDegradation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_node_performance_degradation",
				Help: "Level of node performance degradation",
			},
			[]string{"node_name", "degradation_type"},
		),
	}
}

func (c *NodeStateStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	c.nodeEventsTotal.Describe(ch)
	c.nodeEventRate.Describe(ch)
	c.nodeEventLatency.Describe(ch)
	c.nodeEventSize.Describe(ch)
	c.nodeEventsDropped.Describe(ch)
	c.nodeEventsFailed.Describe(ch)
	c.nodeEventsProcessed.Describe(ch)
	c.nodeStateChanges.Describe(ch)
	c.nodeStateDuration.Describe(ch)
	c.nodeStateTransitions.Describe(ch)
	c.nodeDowntime.Describe(ch)
	c.nodeRecoveryTime.Describe(ch)
	c.nodeMaintenanceEvents.Describe(ch)
	c.activeNodeStreams.Describe(ch)
	c.nodeStreamDuration.Describe(ch)
	c.nodeStreamBandwidth.Describe(ch)
	c.nodeStreamLatency.Describe(ch)
	c.nodeStreamHealth.Describe(ch)
	c.nodeStreamBackpressure.Describe(ch)
	c.nodeStreamFailover.Describe(ch)
	c.nodeStreamReconnections.Describe(ch)
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
	c.nodeCoverage.Describe(ch)
	c.stateChangeAccuracy.Describe(ch)
	c.predictiveAccuracy.Describe(ch)
	c.maintenanceCompliance.Describe(ch)
	c.securityIncidents.Describe(ch)
	c.complianceViolations.Describe(ch)
	c.performanceDegradation.Describe(ch)
}

func (c *NodeStateStreamingCollector) Collect(ch chan<- prometheus.Metric) {
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

	c.nodeEventsTotal.Collect(ch)
	c.nodeEventRate.Collect(ch)
	c.nodeEventLatency.Collect(ch)
	c.nodeEventSize.Collect(ch)
	c.nodeEventsDropped.Collect(ch)
	c.nodeEventsFailed.Collect(ch)
	c.nodeEventsProcessed.Collect(ch)
	c.nodeStateChanges.Collect(ch)
	c.nodeStateDuration.Collect(ch)
	c.nodeStateTransitions.Collect(ch)
	c.nodeDowntime.Collect(ch)
	c.nodeRecoveryTime.Collect(ch)
	c.nodeMaintenanceEvents.Collect(ch)
	c.activeNodeStreams.Collect(ch)
	c.nodeStreamDuration.Collect(ch)
	c.nodeStreamBandwidth.Collect(ch)
	c.nodeStreamLatency.Collect(ch)
	c.nodeStreamHealth.Collect(ch)
	c.nodeStreamBackpressure.Collect(ch)
	c.nodeStreamFailover.Collect(ch)
	c.nodeStreamReconnections.Collect(ch)
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
	c.nodeCoverage.Collect(ch)
	c.stateChangeAccuracy.Collect(ch)
	c.predictiveAccuracy.Collect(ch)
	c.maintenanceCompliance.Collect(ch)
	c.securityIncidents.Collect(ch)
	c.complianceViolations.Collect(ch)
	c.performanceDegradation.Collect(ch)
}

func (c *NodeStateStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetNodeStreamingConfiguration(ctx)
	if err != nil {
		log.Printf("Error collecting node streaming configuration: %v", err)
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

func (c *NodeStateStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActiveNodeStreams(ctx)
	if err != nil {
		log.Printf("Error collecting active node streams: %v", err)
		return
	}

	streamCounts := make(map[string]map[string]int)
	for _, stream := range streams {
		if streamCounts[stream.StreamType] == nil {
			streamCounts[stream.StreamType] = make(map[string]int)
		}
		streamCounts[stream.StreamType][stream.StreamStatus]++

		c.nodeStreamBandwidth.WithLabelValues(stream.StreamID, stream.ConsumerID).Set(stream.Bandwidth)
		c.nodeStreamLatency.WithLabelValues(stream.StreamID, stream.StreamType).Set(stream.Latency.Seconds())

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
		c.nodeStreamHealth.WithLabelValues(stream.StreamID, stream.StreamType).Set(healthScore)

		if stream.BackpressureActive {
			c.nodeStreamBackpressure.WithLabelValues(stream.StreamID).Inc()
		}

		if stream.FailoverActive {
			c.nodeStreamFailover.WithLabelValues(stream.StreamID, "active").Inc()
		}

		streamDuration := time.Since(stream.StreamStartTime)
		c.nodeStreamDuration.WithLabelValues(stream.StreamType).Observe(streamDuration.Seconds())
	}

	for streamType, statusMap := range streamCounts {
		for status, count := range statusMap {
			c.activeNodeStreams.WithLabelValues(streamType, status).Set(float64(count))
		}
	}
}

func (c *NodeStateStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetNodeStreamingMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting node streaming metrics: %v", err)
		return
	}

	c.nodeEventRate.WithLabelValues("all").Set(metrics.EventsPerSecond)
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

	c.nodeEventsProcessed.WithLabelValues("all", "total").Add(float64(metrics.TotalEventsProcessed))
	c.nodeEventsDropped.WithLabelValues("system", "all").Add(float64(metrics.TotalEventsDropped))
	c.nodeEventsFailed.WithLabelValues("processing", "all").Add(float64(metrics.TotalEventsFailed))

	// Node-specific metrics
	c.nodeCoverage.WithLabelValues("all").Set(metrics.NodeCoverage)
	c.stateChangeAccuracy.WithLabelValues("all").Set(metrics.StateChangeAccuracy)
	c.predictiveAccuracy.WithLabelValues("failure").Set(metrics.PredictionAccuracy)
	c.maintenanceCompliance.WithLabelValues("all").Set(metrics.MaintenanceCompliance)
	c.securityIncidents.WithLabelValues("all", "all").Add(float64(metrics.SecurityIncidents))
	c.complianceViolations.WithLabelValues("all", "all").Add(float64(metrics.ComplianceViolations))
}

func (c *NodeStateStreamingCollector) collectStreamingHealth(ctx context.Context, ch chan<- prometheus.Metric) {
	health, err := c.client.GetNodeStreamingHealthStatus(ctx)
	if err != nil {
		log.Printf("Error collecting node streaming health status: %v", err)
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

func (c *NodeStateStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetNodeEventSubscriptions(ctx)
	if err != nil {
		log.Printf("Error collecting node event subscriptions: %v", err)
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

func (c *NodeStateStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetNodeEventFilters(ctx)
	if err != nil {
		log.Printf("Error collecting node event filters: %v", err)
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

func (c *NodeStateStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetNodeEventProcessingStats(ctx)
	if err != nil {
		log.Printf("Error collecting node event processing stats: %v", err)
		return
	}

	c.nodeEventsProcessed.WithLabelValues("all", "received").Add(float64(stats.TotalEventsReceived))
	c.nodeEventsProcessed.WithLabelValues("all", "processed").Add(float64(stats.TotalEventsProcessed))
	c.nodeEventsProcessed.WithLabelValues("all", "filtered").Add(float64(stats.TotalEventsFiltered))
	c.nodeEventsDropped.WithLabelValues("processing", "all").Add(float64(stats.TotalEventsDropped))

	c.validationErrors.WithLabelValues("validation").Add(float64(stats.ValidationErrors))
	c.enrichmentErrors.WithLabelValues("enrichment").Add(float64(stats.EnrichmentErrors))
	c.deliveryErrors.WithLabelValues("http", "delivery").Add(float64(stats.DeliveryErrors))
	c.transformationErrors.WithLabelValues("transformation").Add(float64(stats.TransformationErrors))
	c.networkErrors.WithLabelValues("network").Add(float64(stats.NetworkErrors))
	c.authenticationErrors.WithLabelValues("token").Add(float64(stats.AuthenticationErrors))

	for queueType, depth := range stats.ProcessingQueues {
		c.queueDepth.WithLabelValues(queueType).Set(float64(depth))
	}

	// Node-specific processing stats
	c.stateChangeAccuracy.WithLabelValues("detection").Set(stats.StateChangeAccuracy)
}

func (c *NodeStateStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perfMetrics, err := c.client.GetNodeStreamingPerformanceMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting node streaming performance metrics: %v", err)
		return
	}

	c.nodeEventLatency.WithLabelValues("all", "p50").Observe(perfMetrics.P50Latency.Seconds())
	c.nodeEventLatency.WithLabelValues("all", "p95").Observe(perfMetrics.P95Latency.Seconds())
	c.nodeEventLatency.WithLabelValues("all", "p99").Observe(perfMetrics.P99Latency.Seconds())
	c.nodeEventLatency.WithLabelValues("all", "max").Observe(perfMetrics.MaxLatency.Seconds())

	c.streamingThroughput.WithLabelValues("message_rate").Set(perfMetrics.MessageRate)
	c.streamingThroughput.WithLabelValues("byte_rate").Set(perfMetrics.ByteRate)

	// Node-specific performance metrics
	c.nodeCoverage.WithLabelValues("efficiency").Set(perfMetrics.NodeCoverageEfficiency)
	c.stateChangeAccuracy.WithLabelValues("detection").Set(perfMetrics.StateDetectionAccuracy)
	c.maintenanceCompliance.WithLabelValues("efficiency").Set(perfMetrics.MaintenanceEfficiency)
	c.predictiveAccuracy.WithLabelValues("overall").Set(perfMetrics.PredictiveAccuracy)
}
