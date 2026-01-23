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

type RealtimeJobStreamingSLURMClient interface {
	StreamJobStatusUpdates(ctx context.Context) (<-chan JobStatusEvent, error)
	GetJobStreamingConfiguration(ctx context.Context) (*JobStreamingConfiguration, error)
	GetActiveJobStreams(ctx context.Context) ([]*ActiveJobStream, error)
	GetJobEventHistory(ctx context.Context, jobID string, duration time.Duration) ([]*JobEvent, error)
	GetJobStreamingMetrics(ctx context.Context) (*JobStreamingMetrics, error)
	GetJobEventFilters(ctx context.Context) ([]*JobEventFilter, error)
	ConfigureJobStreaming(ctx context.Context, config *JobStreamingConfiguration) error
	GetStreamingHealthStatus(ctx context.Context) (*StreamingHealthStatus, error)
	GetJobEventSubscriptions(ctx context.Context) ([]*JobEventSubscription, error)
	ManageJobEventSubscription(ctx context.Context, subscription *JobEventSubscription) error
	GetJobEventProcessingStats(ctx context.Context) (*JobEventProcessingStats, error)
	GetJobStreamingPerformanceMetrics(ctx context.Context) (*JobStreamingPerformanceMetrics, error)
}

type JobStatusEvent struct {
	EventID          string
	JobID            string
	UserID           string
	AccountName      string
	PartitionName    string
	PreviousState    string
	CurrentState     string
	StateChangeTime  time.Time
	EventType        string
	Priority         int
	SubmitTime       time.Time
	StartTime        *time.Time
	EndTime          *time.Time
	ExitCode         *int
	NodeList         []string
	CPUsAllocated    int
	MemoryAllocated  int64
	GPUsAllocated    int
	WorkingDirectory string
	ExecutablePath   string
	JobScript        string
	EnvironmentVars  map[string]string
	ResourceLimits   map[string]interface{}
	QOSName          string
	ReservationName  string
	ArrayJobID       *string
	ArrayTaskID      *string
	DependencyInfo   []string
	Comment          string
	Reason           string
	AdminComment     string
	SystemComment    string
	EventMetadata    map[string]interface{}
	StreamingSource  string
	ProcessingTime   time.Duration
	EventSequence    int64
	BatchFlag        bool
	Requeued         bool
	Preempted        bool
	TimeLimit        time.Duration
	TimeLimitRaw     string
}

type JobStreamingConfiguration struct {
	StreamingEnabled         bool
	EventBufferSize          int
	EventBatchSize           int
	EventFlushInterval       time.Duration
	FilterCriteria           []string
	IncludedJobStates        []string
	ExcludedJobStates        []string
	IncludedAccounts         []string
	ExcludedAccounts         []string
	IncludedPartitions       []string
	ExcludedPartitions       []string
	IncludedUsers            []string
	ExcludedUsers            []string
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
}

type ActiveJobStream struct {
	StreamID           string
	JobID              string
	UserID             string
	AccountName        string
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
}

type JobEvent struct {
	EventID          string
	JobID            string
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
}

type JobStreamingMetrics struct {
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
}

type JobEventFilter struct {
	FilterID          string
	FilterName        string
	FilterType        string
	FilterExpression  string
	IncludePattern    string
	ExcludePattern    string
	JobStates         []string
	Accounts          []string
	Partitions        []string
	Users             []string
	QOSNames          []string
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
}

type StreamingHealthStatus struct {
	OverallHealth       string
	ComponentHealth     map[string]string
	LastHealthCheck     time.Time
	HealthCheckDuration time.Duration
	HealthScore         float64
	CriticalIssues      []string
	WarningIssues       []string
	InfoMessages        []string
	StreamingUptime     time.Duration
	ServiceAvailability float64
	ResourceUtilization map[string]float64
	PerformanceMetrics  map[string]float64
	ErrorSummary        map[string]int64
	HealthTrends        map[string]float64
	PredictedIssues     []string
	RecommendedActions  []string
	SystemCapacity      map[string]float64
	AlertThresholds     map[string]float64
	SLACompliance       map[string]float64
	DependencyStatus    map[string]string
	ConfigurationValid  bool
	SecurityStatus      string
	BackupStatus        string
	MonitoringEnabled   bool
	LoggingEnabled      bool
}

type JobEventSubscription struct {
	SubscriptionID      string
	SubscriberName      string
	SubscriberEndpoint  string
	SubscriptionType    string
	EventTypes          []string
	FilterCriteria      string
	DeliveryMethod      string
	DeliveryFormat      string
	SubscriptionStatus  string
	CreatedTime         time.Time
	LastDeliveryTime    time.Time
	DeliveryCount       int64
	FailedDeliveries    int64
	RetryPolicy         string
	MaxRetries          int
	RetryDelay          time.Duration
	ExpirationTime      *time.Time
	Priority            int
	BatchDelivery       bool
	BatchSize           int
	BatchTimeout        time.Duration
	CompressionEnabled  bool
	EncryptionEnabled   bool
	AuthenticationToken string
	CallbackURL         string
	ErrorHandling       string
	DeliveryGuarantee   string
	Metadata            map[string]interface{}
	Tags                []string
	SubscriberContact   string
	BusinessContext     string
	UsageQuota          int64
	UsedQuota           int64
	BandwidthLimit      float64
}

type JobEventProcessingStats struct {
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
}

type JobStreamingPerformanceMetrics struct {
	Throughput             float64
	Latency                time.Duration
	P50Latency             time.Duration
	P95Latency             time.Duration
	P99Latency             time.Duration
	MaxLatency             time.Duration
	MessageRate            float64
	ByteRate               float64
	ErrorRate              float64
	SuccessRate            float64
	AvailabilityPercentage float64
	UptimePercentage       float64
	CPUUtilization         float64
	MemoryUtilization      float64
	NetworkUtilization     float64
	DiskUtilization        float64
	ConnectionCount        int64
	ActiveConnections      int64
	IdleConnections        int64
	FailedConnections      int64
	ConnectionPoolSize     int64
	QueueDepth             int64
	BufferUtilization      float64
	GCPressure             float64
	GCFrequency            float64
	GCDuration             time.Duration
	HeapSize               int64
	ThreadCount            int64
	ContextSwitches        int64
	SystemCalls            int64
	PageFaults             int64
	CacheMisses            int64
	BranchMispredictions   int64
	InstructionsPerSecond  float64
	CyclesPerInstruction   float64
	PerformanceScore       float64
}

type RealtimeJobStreamingCollector struct {
	client RealtimeJobStreamingSLURMClient
	mutex  sync.RWMutex

	// Event metrics
	jobEventsTotal     *prometheus.CounterVec
	jobEventRate       *prometheus.GaugeVec
	jobEventLatency    *prometheus.HistogramVec
	jobEventSize       *prometheus.HistogramVec
	jobEventsDropped   *prometheus.CounterVec
	jobEventsFailed    *prometheus.CounterVec
	jobEventsProcessed *prometheus.CounterVec

	// Stream metrics
	activeJobStreams       *prometheus.GaugeVec
	jobStreamDuration      *prometheus.HistogramVec
	jobStreamBandwidth     *prometheus.GaugeVec
	jobStreamLatency       *prometheus.GaugeVec
	jobStreamHealth        *prometheus.GaugeVec
	jobStreamBackpressure  *prometheus.CounterVec
	jobStreamFailover      *prometheus.CounterVec
	jobStreamReconnections *prometheus.CounterVec

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
}

func NewRealtimeJobStreamingCollector(client RealtimeJobStreamingSLURMClient) *RealtimeJobStreamingCollector {
	return &RealtimeJobStreamingCollector{
		client: client,

		// Event metrics
		jobEventsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_events_total",
				Help: "Total number of job events processed",
			},
			[]string{"event_type", "job_state", "account", "partition"},
		),
		jobEventRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_event_rate",
				Help: "Rate of job events per second",
			},
			[]string{"event_type"},
		),
		jobEventLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_event_latency_seconds",
				Help:    "Latency of job event processing",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
			},
			[]string{"event_type", "processing_stage"},
		),
		jobEventSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_event_size_bytes",
				Help:    "Size of job events in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
			},
			[]string{"event_type"},
		),
		jobEventsDropped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_events_dropped_total",
				Help: "Total number of job events dropped",
			},
			[]string{"reason", "event_type"},
		),
		jobEventsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_events_failed_total",
				Help: "Total number of job events that failed processing",
			},
			[]string{"error_type", "event_type"},
		),
		jobEventsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_events_processed_total",
				Help: "Total number of job events successfully processed",
			},
			[]string{"event_type", "processing_stage"},
		),

		// Stream metrics
		activeJobStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_active_job_streams",
				Help: "Number of active job streams",
			},
			[]string{"stream_type", "stream_status"},
		),
		jobStreamDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_stream_duration_seconds",
				Help:    "Duration of job streams",
				Buckets: []float64{60, 300, 900, 1800, 3600, 7200, 14400, 86400},
			},
			[]string{"stream_type"},
		),
		jobStreamBandwidth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_stream_bandwidth_bytes_per_second",
				Help: "Bandwidth usage of job streams",
			},
			[]string{"stream_id", "consumer_id"},
		),
		jobStreamLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_stream_latency_seconds",
				Help: "Latency of job streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		jobStreamHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_stream_health_score",
				Help: "Health score of job streams",
			},
			[]string{"stream_id", "stream_type"},
		),
		jobStreamBackpressure: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_stream_backpressure_total",
				Help: "Total backpressure events in job streams",
			},
			[]string{"stream_id"},
		),
		jobStreamFailover: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_stream_failover_total",
				Help: "Total failover events in job streams",
			},
			[]string{"stream_id", "failover_reason"},
		),
		jobStreamReconnections: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_stream_reconnections_total",
				Help: "Total reconnection attempts for job streams",
			},
			[]string{"stream_id", "reconnection_reason"},
		),

		// Configuration metrics
		streamingEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_enabled",
				Help: "Whether job streaming is enabled (1=enabled, 0=disabled)",
			},
			[]string{},
		),
		maxConcurrentStreams: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_max_concurrent_streams",
				Help: "Maximum number of concurrent job streams allowed",
			},
			[]string{},
		),
		eventBufferSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_event_buffer_size",
				Help: "Size of the event buffer for job streaming",
			},
			[]string{},
		),
		eventBatchSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_event_batch_size",
				Help: "Batch size for job event processing",
			},
			[]string{},
		),
		rateLimitPerSecond: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_rate_limit_per_second",
				Help: "Rate limit for job streaming per second",
			},
			[]string{},
		),
		backpressureThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_backpressure_threshold",
				Help: "Backpressure threshold for job streaming",
			},
			[]string{},
		),

		// Performance metrics
		streamingThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_throughput",
				Help: "Throughput of job streaming system",
			},
			[]string{"metric_type"},
		),
		streamingCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_cpu_usage",
				Help: "CPU usage of job streaming system",
			},
			[]string{},
		),
		streamingMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_memory_usage_bytes",
				Help: "Memory usage of job streaming system",
			},
			[]string{},
		),
		streamingNetworkUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_network_usage",
				Help: "Network usage of job streaming system",
			},
			[]string{},
		),
		streamingDiskUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_disk_usage_bytes",
				Help: "Disk usage of job streaming system",
			},
			[]string{},
		),
		compressionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_compression_efficiency",
				Help: "Compression efficiency of job streaming",
			},
			[]string{},
		),
		deduplicationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_deduplication_rate",
				Help: "Deduplication rate of job streaming",
			},
			[]string{},
		),
		cacheHitRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_cache_hit_rate",
				Help: "Cache hit rate of job streaming",
			},
			[]string{},
		),
		queueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_queue_depth",
				Help: "Queue depth of job streaming system",
			},
			[]string{"queue_type"},
		),
		processingEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_processing_efficiency",
				Help: "Processing efficiency of job streaming",
			},
			[]string{},
		),

		// Health metrics
		streamingHealthScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_health_score",
				Help: "Overall health score of job streaming system",
			},
			[]string{},
		),
		serviceAvailability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_service_availability",
				Help: "Service availability of job streaming system",
			},
			[]string{},
		),
		streamingUptime: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_streaming_uptime_seconds_total",
				Help: "Total uptime of job streaming system",
			},
			[]string{},
		),
		criticalIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_critical_issues",
				Help: "Number of critical issues in job streaming system",
			},
			[]string{},
		),
		warningIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_warning_issues",
				Help: "Number of warning issues in job streaming system",
			},
			[]string{},
		),
		healthCheckDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_streaming_health_check_duration_seconds",
				Help:    "Duration of health checks for job streaming system",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
			},
			[]string{"component"},
		),
		slaCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_streaming_sla_compliance",
				Help: "SLA compliance of job streaming system",
			},
			[]string{"sla_type"},
		),

		// Subscription metrics
		eventSubscriptionsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_event_subscriptions_total",
				Help: "Total number of job event subscriptions",
			},
			[]string{"subscription_type", "status"},
		),
		subscriptionDeliveries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_subscription_deliveries_total",
				Help: "Total number of subscription deliveries",
			},
			[]string{"subscription_id", "delivery_method"},
		),
		subscriptionFailures: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_subscription_failures_total",
				Help: "Total number of subscription delivery failures",
			},
			[]string{"subscription_id", "failure_reason"},
		),
		subscriptionRetries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_subscription_retries_total",
				Help: "Total number of subscription delivery retries",
			},
			[]string{"subscription_id"},
		),
		subscriptionLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_subscription_latency_seconds",
				Help:    "Latency of subscription deliveries",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"subscription_id", "delivery_method"},
		),

		// Filter metrics
		eventFiltersActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_event_filters_active",
				Help: "Number of active job event filters",
			},
			[]string{"filter_type"},
		),
		filterMatchCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_filter_matches_total",
				Help: "Total number of filter matches",
			},
			[]string{"filter_id", "filter_type"},
		),
		filterProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_job_filter_processing_time_seconds",
				Help:    "Processing time for job event filters",
				Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
			},
			[]string{"filter_id", "filter_type"},
		),
		filterEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_filter_efficiency",
				Help: "Efficiency of job event filters",
			},
			[]string{"filter_id"},
		),

		// Error metrics
		streamingErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_streaming_errors_total",
				Help: "Total number of job streaming errors",
			},
			[]string{"error_type", "component"},
		),
		validationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_validation_errors_total",
				Help: "Total number of job event validation errors",
			},
			[]string{"validation_type"},
		),
		enrichmentErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_enrichment_errors_total",
				Help: "Total number of job event enrichment errors",
			},
			[]string{"enrichment_type"},
		),
		deliveryErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_delivery_errors_total",
				Help: "Total number of job event delivery errors",
			},
			[]string{"delivery_method", "error_code"},
		),
		transformationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_transformation_errors_total",
				Help: "Total number of job event transformation errors",
			},
			[]string{"transformation_type"},
		),
		networkErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_network_errors_total",
				Help: "Total number of job streaming network errors",
			},
			[]string{"network_operation"},
		),
		authenticationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_authentication_errors_total",
				Help: "Total number of job streaming authentication errors",
			},
			[]string{"authentication_method"},
		),
	}
}

func (c *RealtimeJobStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	c.jobEventsTotal.Describe(ch)
	c.jobEventRate.Describe(ch)
	c.jobEventLatency.Describe(ch)
	c.jobEventSize.Describe(ch)
	c.jobEventsDropped.Describe(ch)
	c.jobEventsFailed.Describe(ch)
	c.jobEventsProcessed.Describe(ch)
	c.activeJobStreams.Describe(ch)
	c.jobStreamDuration.Describe(ch)
	c.jobStreamBandwidth.Describe(ch)
	c.jobStreamLatency.Describe(ch)
	c.jobStreamHealth.Describe(ch)
	c.jobStreamBackpressure.Describe(ch)
	c.jobStreamFailover.Describe(ch)
	c.jobStreamReconnections.Describe(ch)
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
}

func (c *RealtimeJobStreamingCollector) Collect(ch chan<- prometheus.Metric) {
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

	c.jobEventsTotal.Collect(ch)
	c.jobEventRate.Collect(ch)
	c.jobEventLatency.Collect(ch)
	c.jobEventSize.Collect(ch)
	c.jobEventsDropped.Collect(ch)
	c.jobEventsFailed.Collect(ch)
	c.jobEventsProcessed.Collect(ch)
	c.activeJobStreams.Collect(ch)
	c.jobStreamDuration.Collect(ch)
	c.jobStreamBandwidth.Collect(ch)
	c.jobStreamLatency.Collect(ch)
	c.jobStreamHealth.Collect(ch)
	c.jobStreamBackpressure.Collect(ch)
	c.jobStreamFailover.Collect(ch)
	c.jobStreamReconnections.Collect(ch)
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
}

func (c *RealtimeJobStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetJobStreamingConfiguration(ctx)
	if err != nil {
		log.Printf("Error collecting job streaming configuration: %v", err)
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

func (c *RealtimeJobStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActiveJobStreams(ctx)
	if err != nil {
		log.Printf("Error collecting active job streams: %v", err)
		return
	}

	streamCounts := make(map[string]map[string]int)
	for _, stream := range streams {
		if streamCounts[stream.StreamType] == nil {
			streamCounts[stream.StreamType] = make(map[string]int)
		}
		streamCounts[stream.StreamType][stream.StreamStatus]++

		c.jobStreamBandwidth.WithLabelValues(stream.StreamID, stream.ConsumerID).Set(stream.Bandwidth)
		c.jobStreamLatency.WithLabelValues(stream.StreamID, stream.StreamType).Set(stream.Latency.Seconds())

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
		c.jobStreamHealth.WithLabelValues(stream.StreamID, stream.StreamType).Set(healthScore)

		if stream.BackpressureActive {
			c.jobStreamBackpressure.WithLabelValues(stream.StreamID).Inc()
		}

		if stream.FailoverActive {
			c.jobStreamFailover.WithLabelValues(stream.StreamID, "active").Inc()
		}

		streamDuration := time.Since(stream.StreamStartTime)
		c.jobStreamDuration.WithLabelValues(stream.StreamType).Observe(streamDuration.Seconds())
	}

	for streamType, statusMap := range streamCounts {
		for status, count := range statusMap {
			c.activeJobStreams.WithLabelValues(streamType, status).Set(float64(count))
		}
	}
}

func (c *RealtimeJobStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetJobStreamingMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting job streaming metrics: %v", err)
		return
	}

	c.jobEventRate.WithLabelValues("all").Set(metrics.EventsPerSecond)
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

	c.jobEventsProcessed.WithLabelValues("all", "total").Add(float64(metrics.TotalEventsProcessed))
	c.jobEventsDropped.WithLabelValues("system", "all").Add(float64(metrics.TotalEventsDropped))
	c.jobEventsFailed.WithLabelValues("processing", "all").Add(float64(metrics.TotalEventsFailed))
}

func (c *RealtimeJobStreamingCollector) collectStreamingHealth(ctx context.Context, ch chan<- prometheus.Metric) {
	health, err := c.client.GetStreamingHealthStatus(ctx)
	if err != nil {
		log.Printf("Error collecting streaming health status: %v", err)
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

func (c *RealtimeJobStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetJobEventSubscriptions(ctx)
	if err != nil {
		log.Printf("Error collecting job event subscriptions: %v", err)
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

func (c *RealtimeJobStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetJobEventFilters(ctx)
	if err != nil {
		log.Printf("Error collecting job event filters: %v", err)
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

func (c *RealtimeJobStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetJobEventProcessingStats(ctx)
	if err != nil {
		log.Printf("Error collecting job event processing stats: %v", err)
		return
	}

	c.jobEventsProcessed.WithLabelValues("all", "received").Add(float64(stats.TotalEventsReceived))
	c.jobEventsProcessed.WithLabelValues("all", "processed").Add(float64(stats.TotalEventsProcessed))
	c.jobEventsProcessed.WithLabelValues("all", "filtered").Add(float64(stats.TotalEventsFiltered))
	c.jobEventsDropped.WithLabelValues("processing", "all").Add(float64(stats.TotalEventsDropped))

	c.validationErrors.WithLabelValues("validation").Add(float64(stats.ValidationErrors))
	c.enrichmentErrors.WithLabelValues("enrichment").Add(float64(stats.EnrichmentErrors))
	c.deliveryErrors.WithLabelValues("http", "delivery").Add(float64(stats.DeliveryErrors))
	c.transformationErrors.WithLabelValues("transformation").Add(float64(stats.TransformationErrors))
	c.networkErrors.WithLabelValues("network").Add(float64(stats.NetworkErrors))
	c.authenticationErrors.WithLabelValues("token").Add(float64(stats.AuthenticationErrors))

	for queueType, depth := range stats.ProcessingQueues {
		c.queueDepth.WithLabelValues(queueType).Set(float64(depth))
	}
}

func (c *RealtimeJobStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perfMetrics, err := c.client.GetJobStreamingPerformanceMetrics(ctx)
	if err != nil {
		log.Printf("Error collecting job streaming performance metrics: %v", err)
		return
	}

	c.jobEventLatency.WithLabelValues("all", "p50").Observe(perfMetrics.P50Latency.Seconds())
	c.jobEventLatency.WithLabelValues("all", "p95").Observe(perfMetrics.P95Latency.Seconds())
	c.jobEventLatency.WithLabelValues("all", "p99").Observe(perfMetrics.P99Latency.Seconds())
	c.jobEventLatency.WithLabelValues("all", "max").Observe(perfMetrics.MaxLatency.Seconds())

	c.streamingThroughput.WithLabelValues("message_rate").Set(perfMetrics.MessageRate)
	c.streamingThroughput.WithLabelValues("byte_rate").Set(perfMetrics.ByteRate)
}
