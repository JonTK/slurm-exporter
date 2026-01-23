// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ClusterHealthStreamingSLURMClient defines the interface for cluster health streaming operations
type ClusterHealthStreamingSLURMClient interface {
	StreamClusterHealthEvents(ctx context.Context) (<-chan ClusterHealthEvent, error)
	GetHealthStreamingConfiguration(ctx context.Context) (*HealthStreamingConfiguration, error)
	GetActiveHealthStreams(ctx context.Context) ([]*ActiveHealthStream, error)
	GetHealthEventHistory(ctx context.Context, componentID string, duration time.Duration) ([]*HealthEvent, error)
	GetHealthStreamingMetrics(ctx context.Context) (*HealthStreamingMetrics, error)
	GetHealthEventFilters(ctx context.Context) ([]*HealthEventFilter, error)
	ConfigureHealthStreaming(ctx context.Context, config *HealthStreamingConfiguration) error
	GetHealthStreamingStatus(ctx context.Context) (*HealthStreamingStatus, error)
	GetHealthEventSubscriptions(ctx context.Context) ([]*HealthEventSubscription, error)
	ManageHealthEventSubscription(ctx context.Context, subscription *HealthEventSubscription) error
	GetHealthEventProcessingStats(ctx context.Context) (*HealthEventProcessingStats, error)
	GetHealthStreamingPerformanceMetrics(ctx context.Context) (*HealthStreamingPerformanceMetrics, error)
}

// ClusterHealthEvent represents a comprehensive cluster health event
type ClusterHealthEvent struct {
	// Event identification
	EventID        string
	ComponentID    string
	ComponentName  string
	ComponentType  string
	EventType      string
	EventTimestamp time.Time
	SequenceNumber int64
	CorrelationID  string

	// Health status
	HealthStatus     string
	HealthScore      float64
	PreviousStatus   string
	StatusDuration   time.Duration
	StatusChangeTime time.Time
	Severity         string
	Impact           string

	// Component metrics
	CPUUtilization     float64
	MemoryUtilization  float64
	DiskUtilization    float64
	NetworkUtilization float64
	IOUtilization      float64
	PowerConsumption   float64
	Temperature        float64

	// Performance metrics
	ResponseTime time.Duration
	Throughput   float64
	ErrorRate    float64
	SuccessRate  float64
	Availability float64
	Reliability  float64
	ServiceLevel float64

	// Capacity metrics
	TotalCapacity       int64
	UsedCapacity        int64
	AvailableCapacity   int64
	ReservedCapacity    int64
	CapacityUtilization float64
	CapacityTrend       float64
	ProjectedExhaustion *time.Time

	// Node health
	TotalNodes       int
	HealthyNodes     int
	DegradedNodes    int
	FailedNodes      int
	MaintenanceNodes int
	DrainedNodes     int
	NodeAvailability float64

	// Job health
	RunningJobs    int
	PendingJobs    int
	FailedJobs     int
	CompletedJobs  int
	JobSuccessRate float64
	AverageJobTime time.Duration
	JobThroughput  float64

	// Queue health
	QueueDepth           int
	QueueLatency         time.Duration
	QueueThroughput      float64
	BacklogSize          int
	BacklogAge           time.Duration
	QueueEfficiency      float64
	SchedulingEfficiency float64

	// Resource health
	ResourceFragmentation float64
	ResourceEfficiency    float64
	ResourceContention    float64
	ResourceWaste         float64
	AllocationEfficiency  float64
	UtilizationBalance    float64
	OptimalUtilization    float64

	// Network health
	NetworkLatency     time.Duration
	PacketLoss         float64
	NetworkBandwidth   float64
	NetworkErrors      int64
	NetworkRetries     int64
	NetworkCongestion  float64
	NetworkReliability float64

	// Storage health
	StorageCapacity   int64
	StorageUsed       int64
	StorageAvailable  int64
	StorageIOPS       float64
	StorageLatency    time.Duration
	StorageThroughput float64
	StorageErrors     int64

	// Service health
	ServiceUptime       time.Duration
	ServiceAvailability float64
	ServiceResponseTime time.Duration
	ServiceErrors       int64
	ServiceRequests     int64
	ServiceSLA          float64
	ServiceDependencies []string

	// Fault detection
	FaultDetected       bool
	FaultType           string
	FaultLocation       string
	FaultImpact         float64
	FaultProbability    float64
	RecoveryTime        time.Duration
	RecoveryProbability float64

	// Anomaly detection
	AnomalyDetected    bool
	AnomalyType        string
	AnomalySeverity    float64
	AnomalyConfidence  float64
	AnomalyDescription string
	BaselineDeviation  float64
	TrendDeviation     float64

	// Predictive health
	PredictedFailure    bool
	FailureProbability  float64
	TimeToFailure       *time.Duration
	PreventiveAction    string
	MaintenanceRequired bool
	MaintenanceUrgency  string
	EstimatedDowntime   time.Duration

	// Business impact
	BusinessImpact     float64
	UserImpact         float64
	SLAImpact          float64
	CostImpact         float64
	ProductivityImpact float64
	ReputationImpact   float64
	ComplianceImpact   float64

	// Recovery metrics
	RecoveryInProgress bool
	RecoveryStartTime  *time.Time
	EstimatedRecovery  *time.Time
	RecoveryProgress   float64
	RecoveryActions    []string
	RecoverySuccess    float64
	FailoverAvailable  bool

	// Dependency health
	DependencyCount      int
	HealthyDependencies  int
	DegradedDependencies int
	FailedDependencies   int
	DependencyImpact     float64
	CascadeRisk          float64
	IsolationLevel       float64

	// Compliance and audit
	ComplianceStatus     bool
	ComplianceViolations []string
	SecurityStatus       string
	SecurityViolations   []string
	AuditRequired        bool
	RegulatoryFlags      []string
	CertificationStatus  []string

	// Recommendations
	RecommendedActions  []string
	OptimizationOptions []string
	PreventiveMeasures  []string
	ImprovementAreas    []string
	BestPractices       []string
	AutomationOptions   []string
	CostSavingOptions   []string

	// Event metadata
	Source        string
	Priority      int
	Tags          []string
	Labels        map[string]string
	Annotations   map[string]string
	RelatedEvents []string
	ParentEvent   string
	ChildEvents   []string
}

// HealthStreamingConfiguration represents health streaming configuration
type HealthStreamingConfiguration struct {
	// Basic settings
	StreamingEnabled     bool
	HealthCheckInterval  time.Duration
	EventBufferSize      int
	EventBatchSize       int
	MaxConcurrentStreams int
	StreamTimeout        time.Duration

	// Health monitoring
	ComponentMonitoring bool
	NodeMonitoring      bool
	JobMonitoring       bool
	QueueMonitoring     bool
	ResourceMonitoring  bool
	NetworkMonitoring   bool
	StorageMonitoring   bool
	ServiceMonitoring   bool

	// Thresholds
	CriticalThreshold   float64
	WarningThreshold    float64
	InfoThreshold       float64
	AnomalyThreshold    float64
	PredictionThreshold float64
	ImpactThreshold     float64
	RecoveryThreshold   float64

	// Detection settings
	FaultDetection      bool
	AnomalyDetection    bool
	PredictiveAnalysis  bool
	TrendAnalysis       bool
	PatternRecognition  bool
	CorrelationAnalysis bool
	RootCauseAnalysis   bool

	// Performance settings
	CompressionEnabled   bool
	EncryptionEnabled    bool
	DeduplicationEnabled bool
	FilteringEnabled     bool
	AggregationEnabled   bool
	SamplingEnabled      bool
	CachingEnabled       bool

	// Reliability settings
	GuaranteedDelivery    bool
	OrderPreservation     bool
	DuplicateHandling     bool
	RetryEnabled          bool
	CircuitBreakerEnabled bool
	BackpressureEnabled   bool
	FailoverEnabled       bool

	// Integration settings
	AlertingEnabled     bool
	NotificationEnabled bool
	WebhookEnabled      bool
	SyslogEnabled       bool
	SNMPEnabled         bool
	PrometheusEnabled   bool
	GrafanaEnabled      bool

	// Advanced features
	MachineLearning         bool
	AutoRemediation         bool
	CapacityPlanning        bool
	CostOptimization        bool
	ComplianceMonitoring    bool
	SecurityMonitoring      bool
	PerformanceOptimization bool

	// Business settings
	SLATracking         bool
	BusinessMetrics     bool
	UserExperience      bool
	CustomerImpact      bool
	RevenueImpact       bool
	ProductivityMetrics bool
	QualityMetrics      bool

	// Automation settings
	AutoScaling      bool
	AutoHealing      bool
	AutoOptimization bool
	AutoFailover     bool
	AutoBackup       bool
	AutoUpdate       bool
	AutoReporting    bool

	// Compliance settings
	AuditLogging        bool
	ComplianceReporting bool
	DataRetention       time.Duration
	DataPrivacy         bool
	Encryption          bool
	AccessControl       bool
	ChangeTracking      bool
}

// ActiveHealthStream represents an active health event stream
type ActiveHealthStream struct {
	// Stream identification
	StreamID     string
	StreamName   string
	StreamType   string
	CreatedAt    time.Time
	LastUpdateAt time.Time
	Status       string

	// Consumer information
	ConsumerID       string
	ConsumerName     string
	ConsumerType     string
	ConsumerEndpoint string
	LastHeartbeat    time.Time
	ConnectionHealth string

	// Stream statistics
	EventsDelivered  int64
	EventsPending    int64
	EventsFiltered   int64
	EventsAggregated int64
	BytesTransferred int64
	CompressionRatio float64

	// Performance metrics
	Throughput     float64
	Latency        time.Duration
	ErrorRate      float64
	ProcessingRate float64
	DeliveryRate   float64
	BacklogSize    int64

	// Quality metrics
	DataQuality       float64
	DataCompleteness  float64
	DataAccuracy      float64
	DataTimeliness    float64
	SignalToNoise     float64
	FalsePositiveRate float64

	// Health monitoring
	ComponentsCovered     int
	HealthChecksPerformed int64
	IssuesDetected        int64
	IssuesResolved        int64
	CriticalAlerts        int64
	MTBF                  time.Duration
	MTTR                  time.Duration

	// Business metrics
	BusinessValue        float64
	CostSavings          float64
	DowntimePrevented    time.Duration
	IncidentsPrevented   int64
	SLACompliance        float64
	CustomerSatisfaction float64
	ROI                  float64
}

// HealthEvent represents an individual health event
type HealthEvent struct {
	Event           ClusterHealthEvent
	ProcessingInfo  map[string]interface{}
	DeliveryInfo    map[string]interface{}
	AnalysisResults map[string]interface{}
}

// HealthStreamingMetrics represents overall health streaming metrics
type HealthStreamingMetrics struct {
	// Stream metrics
	TotalStreams      int
	ActiveStreams     int
	HealthyStreams    int
	DegradedStreams   int
	StreamUtilization float64
	StreamEfficiency  float64

	// Event metrics
	TotalEventsProcessed   int64
	CriticalEvents         int64
	WarningEvents          int64
	InfoEvents             int64
	EventsPerSecond        float64
	EventProcessingLatency time.Duration

	// Health metrics
	ClusterHealthScore   float64
	ComponentHealthScore float64
	ServiceHealthScore   float64
	OverallAvailability  float64
	OverallReliability   float64
	OverallPerformance   float64

	// Detection metrics
	IssuesDetected       int64
	IssuesResolved       int64
	AnomaliesDetected    int64
	PredictionsGenerated int64
	FalsePositives       int64
	FalseNegatives       int64
	DetectionAccuracy    float64

	// Performance metrics
	ResponseTime         time.Duration
	ProcessingThroughput float64
	ResourceUtilization  float64
	NetworkUtilization   float64
	StorageUtilization   float64
	ComputeUtilization   float64

	// Business metrics
	DowntimeAverted         time.Duration
	CostSavingsRealized     float64
	SLAAttainment           float64
	CustomerImpactMinimized float64
	ProductivityMaintained  float64
	RevenueProtected        float64

	// Automation metrics
	AutoRemediations      int64
	AutoScalingEvents     int64
	AutoFailovers         int64
	AutoOptimizations     int64
	AutomationSuccessRate float64
	ManualInterventions   int64

	// Predictive metrics
	PredictiveAccuracy  float64
	PreventiveActions   int64
	PlannedMaintenances int64
	UnplannedOutages    int64
	MTTF                time.Duration
	PredictedFailures   int64

	// Quality metrics
	DataQuality           float64
	MonitoringCoverage    float64
	AlertQuality          float64
	RecommendationQuality float64
	ReportingQuality      float64
	AnalysisDepth         float64

	// Capacity metrics
	CapacityHeadroom      float64
	GrowthRate            float64
	ScalingEfficiency     float64
	ResourceEfficiency    float64
	OptimizationPotential float64
	WasteReduction        float64
}

// HealthEventFilter represents health event filtering configuration
type HealthEventFilter struct {
	FilterID   string
	FilterName string
	FilterType string
	Enabled    bool
	Priority   int

	// Component filters
	ComponentTypes    []string
	ComponentNames    []string
	ComponentStatuses []string
	HealthScoreRange  [2]float64
	SeverityLevels    []string
	ImpactLevels      []string

	// Event filters
	EventTypes          []string
	EventSources        []string
	EventTags           []string
	TimeWindow          time.Duration
	FrequencyThreshold  int
	CorrelationRequired bool

	// Metric filters
	MetricThresholds   map[string]float64
	TrendDirection     string
	DeviationThreshold float64
	AnomalyRequired    bool
	PredictionRequired bool
	BaselineComparison bool

	// Business filters
	BusinessImpactMin  float64
	CostImpactMin      float64
	SLAViolationOnly   bool
	CustomerImpactOnly bool
	ComplianceOnly     bool
	SecurityOnly       bool

	// Statistics
	EventsMatched    int64
	EventsRejected   int64
	ProcessingTime   time.Duration
	FilterEfficiency float64
	LastMatch        time.Time
	FalsePositives   int64
}

// HealthStreamingStatus represents streaming system health status
type HealthStreamingStatus struct {
	// System status
	SystemHealth     float64
	SystemStatus     string
	LastStatusCheck  time.Time
	UptimePercentage float64
	PerformanceScore float64
	ReliabilityScore float64

	// Component status
	ComponentsMonitored  int
	ComponentsHealthy    int
	ComponentsDegraded   int
	ComponentsFailed     int
	ComponentCoverage    float64
	ComponentReliability float64

	// Stream status
	StreamsActive     int
	StreamsHealthy    int
	StreamsDegraded   int
	StreamsFailed     int
	StreamReliability float64
	StreamPerformance float64

	// Processing status
	ProcessingCapacity   float64
	ProcessingLoad       float64
	ProcessingBacklog    int64
	ProcessingLatency    time.Duration
	ProcessingErrors     int64
	ProcessingEfficiency float64

	// Detection status
	DetectionCoverage   float64
	DetectionAccuracy   float64
	DetectionLatency    time.Duration
	FalsePositiveRate   float64
	FalseNegativeRate   float64
	DetectionEfficiency float64

	// Business status
	BusinessValue         float64
	SLACompliance         float64
	CostEfficiency        float64
	CustomerSatisfaction  float64
	OperationalExcellence float64
	StrategicAlignment    float64

	// Risk status
	OverallRisk      float64
	OperationalRisk  float64
	SecurityRisk     float64
	ComplianceRisk   float64
	FinancialRisk    float64
	ReputationalRisk float64

	// Recommendations
	ImmediateActions    []string
	PlannedActions      []string
	OptimizationOptions []string
	RiskMitigations     []string
	CapacityPlanning    []string
	CostOptimizations   []string
}

// HealthEventSubscription represents a health event subscription
type HealthEventSubscription struct {
	// Subscription details
	SubscriptionID   string
	SubscriberID     string
	SubscriptionName string
	SubscriptionType string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Status           string

	// Delivery settings
	DeliveryMethod string
	Endpoint       string
	Protocol       string
	Format         string
	Batching       bool
	Compression    bool
	Encryption     bool

	// Filter settings
	ComponentFilters   []string
	SeverityFilters    []string
	EventTypeFilters   []string
	ImpactThreshold    float64
	FrequencyThreshold int
	CorrelationWindow  time.Duration

	// Quality settings
	DeliveryGuarantee  string
	RetryPolicy        string
	MaxRetries         int
	BackoffStrategy    string
	DeadLetterHandling string
	TimeoutDuration    time.Duration

	// Rate limiting
	RateLimit            int
	BurstLimit           int
	QuotaLimit           int64
	QuotaWindow          time.Duration
	ThrottleStrategy     string
	BackpressureHandling string

	// Security settings
	Authentication     string
	Authorization      string
	APIKey             string
	Certificate        string
	IPWhitelist        []string
	DataClassification string

	// Business settings
	SLALevel           string
	Priority           int
	CostCenter         string
	BillingEnabled     bool
	UsageTracking      bool
	ComplianceRequired bool

	// Statistics
	EventsDelivered int64
	EventsFailed    int64
	BytesDelivered  int64
	AverageLatency  time.Duration
	SuccessRate     float64
	LastDelivery    time.Time

	// Monitoring
	HealthCheckEnabled  bool
	HealthCheckInterval time.Duration
	AlertingEnabled     bool
	MetricsEnabled      bool
	LoggingEnabled      bool
	AuditingEnabled     bool
}

// HealthEventProcessingStats represents health event processing statistics
type HealthEventProcessingStats struct {
	// Processing counts
	TotalEventsProcessed    int64
	CriticalEventsProcessed int64
	WarningEventsProcessed  int64
	InfoEventsProcessed     int64
	SuccessfulProcessing    int64
	FailedProcessing        int64

	// Performance metrics
	AverageProcessingTime time.Duration
	MinProcessingTime     time.Duration
	MaxProcessingTime     time.Duration
	ProcessingTimeP50     time.Duration
	ProcessingTimeP95     time.Duration
	ProcessingTimeP99     time.Duration

	// Throughput metrics
	EventsPerSecond       float64
	PeakThroughput        float64
	SustainedThroughput   float64
	ThroughputVariance    float64
	ThroughputEfficiency  float64
	ThroughputUtilization float64

	// Analysis metrics
	AnalysisDepth        float64
	CorrelationsFound    int64
	PatternsIdentified   int64
	AnomaliesDetected    int64
	PredictionsGenerated int64
	RootCausesIdentified int64

	// Quality metrics
	DataQualityScore   float64
	AnalysisAccuracy   float64
	DetectionPrecision float64
	DetectionRecall    float64
	F1Score            float64
	FalsePositiveRate  float64

	// Resource metrics
	CPUUsage           float64
	MemoryUsage        int64
	NetworkBandwidth   int64
	StorageUsage       int64
	ResourceEfficiency float64
	CostPerEvent       float64

	// Business metrics
	BusinessValueGenerated  float64
	IssuesPrevented         int64
	DowntimeAverted         time.Duration
	CostSavingsRealized     float64
	SLAsMaintained          int64
	CustomerImpactMinimized float64

	// Optimization metrics
	OptimizationsApplied    int64
	EfficiencyGains         float64
	PerformanceImprovements float64
	CostReductions          float64
	QualityImprovements     float64
	AutomationGains         float64
}

// HealthStreamingPerformanceMetrics represents performance metrics
type HealthStreamingPerformanceMetrics struct {
	// Latency metrics
	EndToEndLatency  time.Duration
	DetectionLatency time.Duration
	AnalysisLatency  time.Duration
	DeliveryLatency  time.Duration
	ResponseLatency  time.Duration
	RecoveryLatency  time.Duration

	// Throughput metrics
	EventThroughput       float64
	AnalysisThroughput    float64
	AlertThroughput       float64
	RemediationThroughput float64
	ReportingThroughput   float64
	OverallThroughput     float64

	// Efficiency metrics
	ProcessingEfficiency  float64
	DetectionEfficiency   float64
	AnalysisEfficiency    float64
	RemediationEfficiency float64
	ResourceEfficiency    float64
	CostEfficiency        float64

	// Accuracy metrics
	DetectionAccuracy      float64
	PredictionAccuracy     float64
	AnalysisAccuracy       float64
	CorrelationAccuracy    float64
	RootCauseAccuracy      float64
	RecommendationAccuracy float64

	// Coverage metrics
	MonitoringCoverage  float64
	ComponentCoverage   float64
	MetricCoverage      float64
	EventCoverage       float64
	IssueCoverage       float64
	RemediationCoverage float64

	// Quality metrics
	DataQuality     float64
	AlertQuality    float64
	AnalysisQuality float64
	ReportQuality   float64
	ServiceQuality  float64
	OverallQuality  float64

	// Business metrics
	ValueDelivered         float64
	CostSaved              float64
	DowntimePrevented      time.Duration
	IncidentsPrevented     int64
	ProductivityMaintained float64
	CustomerSatisfaction   float64

	// Scalability metrics
	ScalabilityIndex   float64
	ElasticityScore    float64
	GrowthCapability   float64
	LoadBalancing      float64
	ResourceScaling    float64
	PerformanceScaling float64
}

// ClusterHealthStreamingCollector collects cluster health streaming metrics
type ClusterHealthStreamingCollector struct {
	client                  ClusterHealthStreamingSLURMClient
	healthEvents            *prometheus.Desc
	activeHealthStreams     *prometheus.Desc
	clusterHealthScore      *prometheus.Desc
	componentHealth         *prometheus.Desc
	nodeAvailability        *prometheus.Desc
	jobHealthMetrics        *prometheus.Desc
	resourceUtilization     *prometheus.Desc
	serviceAvailability     *prometheus.Desc
	anomalyDetection        *prometheus.Desc
	predictiveHealth        *prometheus.Desc
	businessImpact          *prometheus.Desc
	recoveryMetrics         *prometheus.Desc
	streamingPerformance    *prometheus.Desc
	detectionAccuracy       *prometheus.Desc
	automationEffectiveness *prometheus.Desc
	costOptimization        *prometheus.Desc
	complianceStatus        *prometheus.Desc
	riskAssessment          *prometheus.Desc
	capacityForecast        *prometheus.Desc
	streamConfigMetrics     map[string]*prometheus.Desc
	eventFilterMetrics      map[string]*prometheus.Desc
	subscriptionMetrics     map[string]*prometheus.Desc
	processingStatsMetrics  map[string]*prometheus.Desc
	performanceMetrics      map[string]*prometheus.Desc
}

// NewClusterHealthStreamingCollector creates a new cluster health streaming collector
func NewClusterHealthStreamingCollector(client ClusterHealthStreamingSLURMClient) *ClusterHealthStreamingCollector {
	return &ClusterHealthStreamingCollector{
		client: client,
		healthEvents: prometheus.NewDesc(
			"slurm_health_events_total",
			"Total number of cluster health events",
			[]string{"event_type", "severity", "component_type", "status"},
			nil,
		),
		activeHealthStreams: prometheus.NewDesc(
			"slurm_active_health_streams",
			"Number of active health monitoring streams",
			[]string{"stream_type", "status", "consumer_type"},
			nil,
		),
		clusterHealthScore: prometheus.NewDesc(
			"slurm_cluster_health_score",
			"Overall cluster health score (0-1)",
			[]string{"component", "dimension"},
			nil,
		),
		componentHealth: prometheus.NewDesc(
			"slurm_component_health_status",
			"Health status of cluster components",
			[]string{"component_type", "component_name", "status"},
			nil,
		),
		nodeAvailability: prometheus.NewDesc(
			"slurm_node_availability_ratio",
			"Node availability ratio",
			[]string{"state", "partition"},
			nil,
		),
		jobHealthMetrics: prometheus.NewDesc(
			"slurm_job_health_metrics",
			"Job health metrics",
			[]string{"metric_type", "status"},
			nil,
		),
		resourceUtilization: prometheus.NewDesc(
			"slurm_resource_utilization_health",
			"Resource utilization health metrics",
			[]string{"resource_type", "metric"},
			nil,
		),
		serviceAvailability: prometheus.NewDesc(
			"slurm_service_availability_ratio",
			"Service availability ratio",
			[]string{"service_name", "dependency"},
			nil,
		),
		anomalyDetection: prometheus.NewDesc(
			"slurm_anomaly_detection_rate",
			"Rate of anomaly detection",
			[]string{"anomaly_type", "severity", "confidence"},
			nil,
		),
		predictiveHealth: prometheus.NewDesc(
			"slurm_predictive_health_score",
			"Predictive health score and failure probability",
			[]string{"component", "prediction_type"},
			nil,
		),
		businessImpact: prometheus.NewDesc(
			"slurm_health_business_impact",
			"Business impact of health issues",
			[]string{"impact_type", "severity"},
			nil,
		),
		recoveryMetrics: prometheus.NewDesc(
			"slurm_recovery_metrics",
			"Recovery and remediation metrics",
			[]string{"recovery_type", "status"},
			nil,
		),
		streamingPerformance: prometheus.NewDesc(
			"slurm_health_streaming_performance",
			"Health streaming performance metrics",
			[]string{"metric_type", "stream_type"},
			nil,
		),
		detectionAccuracy: prometheus.NewDesc(
			"slurm_health_detection_accuracy",
			"Health issue detection accuracy",
			[]string{"detection_type", "metric"},
			nil,
		),
		automationEffectiveness: prometheus.NewDesc(
			"slurm_health_automation_effectiveness",
			"Effectiveness of automated health management",
			[]string{"automation_type", "metric"},
			nil,
		),
		costOptimization: prometheus.NewDesc(
			"slurm_health_cost_optimization",
			"Cost optimization through health management",
			[]string{"optimization_type", "metric"},
			nil,
		),
		complianceStatus: prometheus.NewDesc(
			"slurm_health_compliance_status",
			"Compliance status metrics",
			[]string{"compliance_type", "status"},
			nil,
		),
		riskAssessment: prometheus.NewDesc(
			"slurm_health_risk_assessment",
			"Risk assessment scores",
			[]string{"risk_type", "severity"},
			nil,
		),
		capacityForecast: prometheus.NewDesc(
			"slurm_health_capacity_forecast",
			"Capacity forecast metrics",
			[]string{"resource_type", "forecast_type"},
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
func (c *ClusterHealthStreamingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.healthEvents
	ch <- c.activeHealthStreams
	ch <- c.clusterHealthScore
	ch <- c.componentHealth
	ch <- c.nodeAvailability
	ch <- c.jobHealthMetrics
	ch <- c.resourceUtilization
	ch <- c.serviceAvailability
	ch <- c.anomalyDetection
	ch <- c.predictiveHealth
	ch <- c.businessImpact
	ch <- c.recoveryMetrics
	ch <- c.streamingPerformance
	ch <- c.detectionAccuracy
	ch <- c.automationEffectiveness
	ch <- c.costOptimization
	ch <- c.complianceStatus
	ch <- c.riskAssessment
	ch <- c.capacityForecast

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
func (c *ClusterHealthStreamingCollector) Collect(ch chan<- prometheus.Metric) {
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

func (c *ClusterHealthStreamingCollector) collectStreamingConfiguration(ctx context.Context, ch chan<- prometheus.Metric) {
	config, err := c.client.GetHealthStreamingConfiguration(ctx)
	if err != nil {
		return
	}

	// Configuration enabled/disabled metrics
	configurations := map[string]bool{
		"streaming_enabled":     config.StreamingEnabled,
		"component_monitoring":  config.ComponentMonitoring,
		"node_monitoring":       config.NodeMonitoring,
		"job_monitoring":        config.JobMonitoring,
		"queue_monitoring":      config.QueueMonitoring,
		"resource_monitoring":   config.ResourceMonitoring,
		"network_monitoring":    config.NetworkMonitoring,
		"storage_monitoring":    config.StorageMonitoring,
		"service_monitoring":    config.ServiceMonitoring,
		"fault_detection":       config.FaultDetection,
		"anomaly_detection":     config.AnomalyDetection,
		"predictive_analysis":   config.PredictiveAnalysis,
		"trend_analysis":        config.TrendAnalysis,
		"pattern_recognition":   config.PatternRecognition,
		"correlation_analysis":  config.CorrelationAnalysis,
		"root_cause_analysis":   config.RootCauseAnalysis,
		"compression_enabled":   config.CompressionEnabled,
		"encryption_enabled":    config.EncryptionEnabled,
		"guaranteed_delivery":   config.GuaranteedDelivery,
		"alerting_enabled":      config.AlertingEnabled,
		"machine_learning":      config.MachineLearning,
		"auto_remediation":      config.AutoRemediation,
		"auto_scaling":          config.AutoScaling,
		"auto_healing":          config.AutoHealing,
		"compliance_monitoring": config.ComplianceMonitoring,
		"security_monitoring":   config.SecurityMonitoring,
	}

	for name, enabled := range configurations {
		value := 0.0
		if enabled {
			value = 1.0
		}

		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_health_stream_config_"+name,
				"Health streaming configuration for "+name,
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

	// Threshold metrics
	thresholds := map[string]float64{
		"critical_threshold":   config.CriticalThreshold,
		"warning_threshold":    config.WarningThreshold,
		"info_threshold":       config.InfoThreshold,
		"anomaly_threshold":    config.AnomalyThreshold,
		"prediction_threshold": config.PredictionThreshold,
		"impact_threshold":     config.ImpactThreshold,
		"recovery_threshold":   config.RecoveryThreshold,
	}

	for name, value := range thresholds {
		if _, exists := c.streamConfigMetrics[name]; !exists {
			c.streamConfigMetrics[name] = prometheus.NewDesc(
				"slurm_health_stream_config_"+name,
				"Health streaming threshold for "+name,
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

func (c *ClusterHealthStreamingCollector) collectActiveStreams(ctx context.Context, ch chan<- prometheus.Metric) {
	streams, err := c.client.GetActiveHealthStreams(ctx)
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
			c.streamingPerformance,
			prometheus.GaugeValue,
			stream.Throughput,
			"throughput", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			stream.Latency.Seconds(),
			"latency", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			stream.DataQuality,
			"data_quality", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.businessImpact,
			prometheus.GaugeValue,
			stream.BusinessValue,
			"value", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.businessImpact,
			prometheus.GaugeValue,
			stream.CostSavings,
			"cost_savings", stream.StreamType,
		)

		ch <- prometheus.MustNewConstMetric(
			c.serviceAvailability,
			prometheus.GaugeValue,
			stream.SLACompliance,
			stream.StreamName, "sla",
		)
	}

	// Aggregate stream counts
	for key, count := range streamCounts {
		parts := strings.Split(key, "_")
		if len(parts) >= 3 {
			ch <- prometheus.MustNewConstMetric(
				c.activeHealthStreams,
				prometheus.GaugeValue,
				float64(count),
				parts[0], parts[1], parts[2],
			)
		}
	}
}

func (c *ClusterHealthStreamingCollector) collectStreamingMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	metrics, err := c.client.GetHealthStreamingMetrics(ctx)
	if err != nil {
		return
	}

	// Event metrics
	ch <- prometheus.MustNewConstMetric(
		c.healthEvents,
		prometheus.CounterValue,
		float64(metrics.TotalEventsProcessed),
		"all", "all", "all", "processed",
	)

	ch <- prometheus.MustNewConstMetric(
		c.healthEvents,
		prometheus.CounterValue,
		float64(metrics.CriticalEvents),
		"health", "critical", "all", "processed",
	)

	ch <- prometheus.MustNewConstMetric(
		c.healthEvents,
		prometheus.CounterValue,
		float64(metrics.WarningEvents),
		"health", "warning", "all", "processed",
	)

	// Health scores
	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.ClusterHealthScore,
		"cluster", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.ComponentHealthScore,
		"components", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.ServiceHealthScore,
		"services", "overall",
	)

	// Performance metrics
	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.OverallAvailability,
		"availability", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.OverallReliability,
		"reliability", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		metrics.OverallPerformance,
		"performance", "overall",
	)

	// Detection metrics
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		metrics.DetectionAccuracy,
		"overall", "accuracy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.anomalyDetection,
		prometheus.CounterValue,
		float64(metrics.AnomaliesDetected),
		"all", "all", "detected",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictiveHealth,
		prometheus.CounterValue,
		float64(metrics.PredictionsGenerated),
		"all", "predictions",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		metrics.DowntimeAverted.Hours(),
		"downtime_averted", "hours",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimization,
		prometheus.GaugeValue,
		metrics.CostSavingsRealized,
		"realized", "total",
	)

	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		metrics.SLAAttainment,
		"sla_attainment", "percentage",
	)

	// Automation metrics
	ch <- prometheus.MustNewConstMetric(
		c.automationEffectiveness,
		prometheus.CounterValue,
		float64(metrics.AutoRemediations),
		"remediation", "count",
	)

	ch <- prometheus.MustNewConstMetric(
		c.automationEffectiveness,
		prometheus.GaugeValue,
		metrics.AutomationSuccessRate,
		"overall", "success_rate",
	)

	// Predictive metrics
	ch <- prometheus.MustNewConstMetric(
		c.predictiveHealth,
		prometheus.GaugeValue,
		metrics.PredictiveAccuracy,
		"overall", "accuracy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictiveHealth,
		prometheus.CounterValue,
		float64(metrics.PreventiveActions),
		"actions", "preventive",
	)

	// Capacity metrics
	ch <- prometheus.MustNewConstMetric(
		c.capacityForecast,
		prometheus.GaugeValue,
		metrics.CapacityHeadroom,
		"overall", "headroom",
	)

	ch <- prometheus.MustNewConstMetric(
		c.capacityForecast,
		prometheus.GaugeValue,
		metrics.GrowthRate,
		"overall", "growth_rate",
	)
}

func (c *ClusterHealthStreamingCollector) collectHealthStatus(ctx context.Context, ch chan<- prometheus.Metric) {
	status, err := c.client.GetHealthStreamingStatus(ctx)
	if err != nil {
		return
	}

	// System health metrics
	ch <- prometheus.MustNewConstMetric(
		c.clusterHealthScore,
		prometheus.GaugeValue,
		status.SystemHealth,
		"system", "status",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		status.PerformanceScore,
		"system_performance", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		status.ReliabilityScore,
		"system_reliability", "overall",
	)

	// Component health
	ch <- prometheus.MustNewConstMetric(
		c.componentHealth,
		prometheus.GaugeValue,
		float64(status.ComponentsHealthy),
		"all", "all", "healthy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.componentHealth,
		prometheus.GaugeValue,
		float64(status.ComponentsDegraded),
		"all", "all", "degraded",
	)

	ch <- prometheus.MustNewConstMetric(
		c.componentHealth,
		prometheus.GaugeValue,
		float64(status.ComponentsFailed),
		"all", "all", "failed",
	)

	// Processing status
	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		status.ProcessingCapacity,
		"processing_capacity", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		status.ProcessingLatency.Seconds(),
		"processing_latency", "overall",
	)

	// Detection status
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		status.DetectionCoverage,
		"detection", "coverage",
	)

	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		status.DetectionAccuracy,
		"detection", "accuracy",
	)

	// Business status
	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		status.BusinessValue,
		"business_value", "overall",
	)

	ch <- prometheus.MustNewConstMetric(
		c.complianceStatus,
		prometheus.GaugeValue,
		status.SLACompliance,
		"sla", "overall",
	)

	// Risk metrics
	ch <- prometheus.MustNewConstMetric(
		c.riskAssessment,
		prometheus.GaugeValue,
		status.OverallRisk,
		"overall", "all",
	)

	ch <- prometheus.MustNewConstMetric(
		c.riskAssessment,
		prometheus.GaugeValue,
		status.OperationalRisk,
		"operational", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.riskAssessment,
		prometheus.GaugeValue,
		status.SecurityRisk,
		"security", "current",
	)

	ch <- prometheus.MustNewConstMetric(
		c.riskAssessment,
		prometheus.GaugeValue,
		status.ComplianceRisk,
		"compliance", "current",
	)
}

func (c *ClusterHealthStreamingCollector) collectEventSubscriptions(ctx context.Context, ch chan<- prometheus.Metric) {
	subscriptions, err := c.client.GetHealthEventSubscriptions(ctx)
	if err != nil {
		return
	}

	for _, sub := range subscriptions {
		// Subscription metrics
		ch <- prometheus.MustNewConstMetric(
			c.healthEvents,
			prometheus.CounterValue,
			float64(sub.EventsDelivered),
			"subscription", "delivered", sub.SubscriptionType, sub.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			c.healthEvents,
			prometheus.CounterValue,
			float64(sub.EventsFailed),
			"subscription", "failed", sub.SubscriptionType, sub.Status,
		)

		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			sub.SuccessRate,
			"subscription_success", sub.SubscriptionID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			sub.AverageLatency.Seconds(),
			"subscription_latency", sub.SubscriptionID,
		)
	}
}

func (c *ClusterHealthStreamingCollector) collectEventFilters(ctx context.Context, ch chan<- prometheus.Metric) {
	filters, err := c.client.GetHealthEventFilters(ctx)
	if err != nil {
		return
	}

	for _, filter := range filters {
		if !filter.Enabled {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.healthEvents,
			prometheus.CounterValue,
			float64(filter.EventsMatched),
			"filter", "matched", filter.FilterType, "active",
		)

		ch <- prometheus.MustNewConstMetric(
			c.healthEvents,
			prometheus.CounterValue,
			float64(filter.EventsRejected),
			"filter", "rejected", filter.FilterType, "active",
		)

		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			filter.FilterEfficiency,
			"filter_efficiency", filter.FilterID,
		)
	}
}

func (c *ClusterHealthStreamingCollector) collectProcessingStats(ctx context.Context, ch chan<- prometheus.Metric) {
	stats, err := c.client.GetHealthEventProcessingStats(ctx)
	if err != nil {
		return
	}

	// Processing counts
	ch <- prometheus.MustNewConstMetric(
		c.healthEvents,
		prometheus.CounterValue,
		float64(stats.TotalEventsProcessed),
		"processing", "all", "all", "total",
	)

	ch <- prometheus.MustNewConstMetric(
		c.healthEvents,
		prometheus.CounterValue,
		float64(stats.SuccessfulProcessing),
		"processing", "all", "all", "successful",
	)

	// Performance metrics
	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		stats.AverageProcessingTime.Seconds(),
		"processing_time", "average",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		stats.ProcessingTimeP95.Seconds(),
		"processing_time", "p95",
	)

	// Throughput metrics
	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		stats.EventsPerSecond,
		"throughput", "events_per_second",
	)

	// Analysis metrics
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		stats.AnalysisAccuracy,
		"analysis", "accuracy",
	)

	ch <- prometheus.MustNewConstMetric(
		c.anomalyDetection,
		prometheus.CounterValue,
		float64(stats.AnomaliesDetected),
		"processing", "all", "detected",
	)

	// Quality metrics
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		stats.F1Score,
		"quality", "f1_score",
	)

	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		stats.FalsePositiveRate,
		"quality", "false_positive_rate",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		stats.BusinessValueGenerated,
		"processing_value", "generated",
	)

	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		stats.DowntimeAverted.Hours(),
		"processing_downtime", "averted_hours",
	)

	// Optimization metrics
	ch <- prometheus.MustNewConstMetric(
		c.automationEffectiveness,
		prometheus.GaugeValue,
		stats.EfficiencyGains,
		"optimization", "efficiency_gains",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimization,
		prometheus.GaugeValue,
		stats.CostReductions,
		"optimization", "cost_reductions",
	)
}

func (c *ClusterHealthStreamingCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) {
	perf, err := c.client.GetHealthStreamingPerformanceMetrics(ctx)
	if err != nil {
		return
	}

	// Latency metrics
	latencyMetrics := map[string]time.Duration{
		"end_to_end": perf.EndToEndLatency,
		"detection":  perf.DetectionLatency,
		"analysis":   perf.AnalysisLatency,
		"delivery":   perf.DeliveryLatency,
		"response":   perf.ResponseLatency,
		"recovery":   perf.RecoveryLatency,
	}

	for name, latency := range latencyMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			latency.Seconds(),
			"latency_"+name, "measured",
		)
	}

	// Throughput metrics
	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		perf.EventThroughput,
		"throughput", "events",
	)

	ch <- prometheus.MustNewConstMetric(
		c.streamingPerformance,
		prometheus.GaugeValue,
		perf.OverallThroughput,
		"throughput", "overall",
	)

	// Efficiency metrics
	efficiencyMetrics := map[string]float64{
		"processing":  perf.ProcessingEfficiency,
		"detection":   perf.DetectionEfficiency,
		"analysis":    perf.AnalysisEfficiency,
		"remediation": perf.RemediationEfficiency,
		"resource":    perf.ResourceEfficiency,
		"cost":        perf.CostEfficiency,
	}

	for name, efficiency := range efficiencyMetrics {
		ch <- prometheus.MustNewConstMetric(
			c.streamingPerformance,
			prometheus.GaugeValue,
			efficiency,
			"efficiency_"+name, "ratio",
		)
	}

	// Accuracy metrics
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		perf.DetectionAccuracy,
		"performance", "detection",
	)

	ch <- prometheus.MustNewConstMetric(
		c.predictiveHealth,
		prometheus.GaugeValue,
		perf.PredictionAccuracy,
		"performance", "prediction",
	)

	// Coverage metrics
	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		perf.MonitoringCoverage,
		"coverage", "monitoring",
	)

	ch <- prometheus.MustNewConstMetric(
		c.detectionAccuracy,
		prometheus.GaugeValue,
		perf.RemediationCoverage,
		"coverage", "remediation",
	)

	// Business metrics
	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		perf.ValueDelivered,
		"performance_value", "delivered",
	)

	ch <- prometheus.MustNewConstMetric(
		c.costOptimization,
		prometheus.GaugeValue,
		perf.CostSaved,
		"performance_cost", "saved",
	)

	ch <- prometheus.MustNewConstMetric(
		c.businessImpact,
		prometheus.GaugeValue,
		perf.CustomerSatisfaction,
		"performance", "customer_satisfaction",
	)

	// Scalability metrics
	ch <- prometheus.MustNewConstMetric(
		c.capacityForecast,
		prometheus.GaugeValue,
		perf.ScalabilityIndex,
		"scalability", "index",
	)

	ch <- prometheus.MustNewConstMetric(
		c.capacityForecast,
		prometheus.GaugeValue,
		perf.GrowthCapability,
		"scalability", "growth_capability",
	)
}
