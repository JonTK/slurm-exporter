// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// EventAggregationCorrelationSLURMClient defines the interface for event aggregation and correlation operations
type EventAggregationCorrelationSLURMClient interface {
	// Event Aggregation Operations
	StreamAggregatedEvents(ctx context.Context) (<-chan AggregatedEvent, error)
	GetAggregationConfiguration(ctx context.Context) (*AggregationConfiguration, error)
	GetActiveAggregations(ctx context.Context) ([]*ActiveAggregation, error)
	GetAggregationHistory(ctx context.Context, aggregationID string, duration time.Duration) ([]*AggregationRecord, error)
	GetAggregationMetrics(ctx context.Context) (*AggregationMetrics, error)
	ConfigureAggregation(ctx context.Context, config *AggregationConfiguration) error

	// Event Correlation Operations
	StreamCorrelatedEvents(ctx context.Context) (<-chan CorrelatedEvent, error)
	GetCorrelationConfiguration(ctx context.Context) (*CorrelationConfiguration, error)
	GetActiveCorrelations(ctx context.Context) ([]*ActiveCorrelation, error)
	GetCorrelationHistory(ctx context.Context, correlationID string, duration time.Duration) ([]*CorrelationRecord, error)
	GetCorrelationMetrics(ctx context.Context) (*CorrelationMetrics, error)
	ConfigureCorrelation(ctx context.Context, config *CorrelationConfiguration) error
}

// AggregatedEvent represents a comprehensive aggregated event with analysis and correlation data
type AggregatedEvent struct {
	// Core Identification
	EventID         string        `json:"event_id"`
	AggregationID   string        `json:"aggregation_id"`
	AggregationType string        `json:"aggregation_type"`
	AggregationName string        `json:"aggregation_name"`
	Timestamp       time.Time     `json:"timestamp"`
	TimeWindow      time.Duration `json:"time_window"`
	WindowStart     time.Time     `json:"window_start"`
	WindowEnd       time.Time     `json:"window_end"`

	// Source Events
	SourceEventCount   int64         `json:"source_event_count"`
	SourceEventTypes   []string      `json:"source_event_types"`
	SourceEventSources []string      `json:"source_event_sources"`
	FirstEventTime     time.Time     `json:"first_event_time"`
	LastEventTime      time.Time     `json:"last_event_time"`
	EventSpread        time.Duration `json:"event_spread"`
	EventRate          float64       `json:"event_rate"`
	EventVelocity      float64       `json:"event_velocity"`
	EventAcceleration  float64       `json:"event_acceleration"`

	// Aggregation Metrics
	AggregationMethod       string        `json:"aggregation_method"`
	AggregationRules        []string      `json:"aggregation_rules"`
	AggregationFunction     string        `json:"aggregation_function"`
	GroupingKeys            []string      `json:"grouping_keys"`
	AggregationLevel        string        `json:"aggregation_level"`
	AggregationScope        string        `json:"aggregation_scope"`
	AggregationAccuracy     float64       `json:"aggregation_accuracy"`
	AggregationConfidence   float64       `json:"aggregation_confidence"`
	AggregationCompleteness float64       `json:"aggregation_completeness"`
	AggregationLatency      time.Duration `json:"aggregation_latency"`

	// Statistical Analysis
	StatisticalMean         float64   `json:"statistical_mean"`
	StatisticalMedian       float64   `json:"statistical_median"`
	StatisticalMode         float64   `json:"statistical_mode"`
	StatisticalStdDev       float64   `json:"statistical_std_dev"`
	StatisticalVariance     float64   `json:"statistical_variance"`
	StatisticalMin          float64   `json:"statistical_min"`
	StatisticalMax          float64   `json:"statistical_max"`
	StatisticalRange        float64   `json:"statistical_range"`
	StatisticalQuartiles    []float64 `json:"statistical_quartiles"`
	StatisticalOutliers     int64     `json:"statistical_outliers"`
	StatisticalDistribution string    `json:"statistical_distribution"`
	StatisticalTrends       []string  `json:"statistical_trends"`

	// Resource Aggregations
	TotalCPUTime          float64 `json:"total_cpu_time"`
	AverageCPUUtilization float64 `json:"average_cpu_utilization"`
	PeakCPUUtilization    float64 `json:"peak_cpu_utilization"`
	TotalMemoryUsage      int64   `json:"total_memory_usage"`
	AverageMemoryUsage    int64   `json:"average_memory_usage"`
	PeakMemoryUsage       int64   `json:"peak_memory_usage"`
	TotalStorageIO        int64   `json:"total_storage_io"`
	AverageStorageIO      int64   `json:"average_storage_io"`
	PeakStorageIO         int64   `json:"peak_storage_io"`
	TotalNetworkTraffic   int64   `json:"total_network_traffic"`
	AverageNetworkTraffic int64   `json:"average_network_traffic"`
	PeakNetworkTraffic    int64   `json:"peak_network_traffic"`
	TotalGPUTime          float64 `json:"total_gpu_time"`
	AverageGPUUtilization float64 `json:"average_gpu_utilization"`
	PeakGPUUtilization    float64 `json:"peak_gpu_utilization"`

	// Job Aggregations
	TotalJobs          int64         `json:"total_jobs"`
	CompletedJobs      int64         `json:"completed_jobs"`
	FailedJobs         int64         `json:"failed_jobs"`
	CancelledJobs      int64         `json:"cancelled_jobs"`
	RunningJobs        int64         `json:"running_jobs"`
	QueuedJobs         int64         `json:"queued_jobs"`
	AverageJobDuration time.Duration `json:"average_job_duration"`
	AverageWaitTime    time.Duration `json:"average_wait_time"`
	AverageQueueTime   time.Duration `json:"average_queue_time"`
	JobSuccessRate     float64       `json:"job_success_rate"`
	JobFailureRate     float64       `json:"job_failure_rate"`
	JobThroughput      float64       `json:"job_throughput"`
	JobEfficiency      float64       `json:"job_efficiency"`

	// User and Account Aggregations
	UniqueUsers         int64    `json:"unique_users"`
	UniqueAccounts      int64    `json:"unique_accounts"`
	TopUsers            []string `json:"top_users"`
	TopAccounts         []string `json:"top_accounts"`
	UserActivityPattern string   `json:"user_activity_pattern"`
	AccountUsagePattern string   `json:"account_usage_pattern"`
	FairShareImpact     float64  `json:"fair_share_impact"`
	PriorityImpact      float64  `json:"priority_impact"`
	ResourceSharing     float64  `json:"resource_sharing"`
	CollaborationLevel  float64  `json:"collaboration_level"`

	// System Performance Aggregations
	SystemLoad         float64  `json:"system_load"`
	SystemUtilization  float64  `json:"system_utilization"`
	SystemEfficiency   float64  `json:"system_efficiency"`
	SystemThroughput   float64  `json:"system_throughput"`
	SystemBottlenecks  []string `json:"system_bottlenecks"`
	SystemHealth       float64  `json:"system_health"`
	SystemAvailability float64  `json:"system_availability"`
	SystemReliability  float64  `json:"system_reliability"`
	SystemPerformance  float64  `json:"system_performance"`
	SystemOptimization float64  `json:"system_optimization"`

	// Quality and Reliability Metrics
	DataQuality           float64 `json:"data_quality"`
	DataCompleteness      float64 `json:"data_completeness"`
	DataAccuracy          float64 `json:"data_accuracy"`
	DataConsistency       float64 `json:"data_consistency"`
	DataTimeliness        float64 `json:"data_timeliness"`
	DataReliability       float64 `json:"data_reliability"`
	AggregationQuality    float64 `json:"aggregation_quality"`
	ProcessingReliability float64 `json:"processing_reliability"`
	ResultValidation      bool    `json:"result_validation"`
	QualityScore          float64 `json:"quality_score"`

	// Business Impact Metrics
	BusinessValue          float64       `json:"business_value"`
	CostImpact             float64       `json:"cost_impact"`
	EfficiencyGain         float64       `json:"efficiency_gain"`
	ProductivityImpact     float64       `json:"productivity_impact"`
	ResourceSavings        float64       `json:"resource_savings"`
	TimeToInsight          time.Duration `json:"time_to_insight"`
	DecisionSupport        float64       `json:"decision_support"`
	OperationalImprovement float64       `json:"operational_improvement"`
	ROIMeasurement         float64       `json:"roi_measurement"`
	ValueRealization       float64       `json:"value_realization"`

	// Alerting and Notifications
	AlertsGenerated      int64    `json:"alerts_generated"`
	CriticalAlerts       int64    `json:"critical_alerts"`
	WarningAlerts        int64    `json:"warning_alerts"`
	InfoAlerts           int64    `json:"info_alerts"`
	AlertsTriggered      []string `json:"alerts_triggered"`
	NotificationsSent    int64    `json:"notifications_sent"`
	EscalationsTriggered int64    `json:"escalations_triggered"`
	TicketsCreated       int64    `json:"tickets_created"`
	AutomatedResponses   int64    `json:"automated_responses"`
	ManualInterventions  int64    `json:"manual_interventions"`

	// Performance and Optimization
	ProcessingTime    time.Duration `json:"processing_time"`
	MemoryFootprint   int64         `json:"memory_footprint"`
	CPUUsage          float64       `json:"cpu_usage"`
	IOOperations      int64         `json:"io_operations"`
	NetworkRequests   int64         `json:"network_requests"`
	CacheHitRate      float64       `json:"cache_hit_rate"`
	CompressionRatio  float64       `json:"compression_ratio"`
	OptimizationLevel float64       `json:"optimization_level"`
	ScalabilityFactor float64       `json:"scalability_factor"`
	PerformanceIndex  float64       `json:"performance_index"`

	// Metadata and Context
	Tags            map[string]string      `json:"tags"`
	Attributes      map[string]interface{} `json:"attributes"`
	Context         map[string]string      `json:"context"`
	Environment     string                 `json:"environment"`
	Version         string                 `json:"version"`
	CreatedBy       string                 `json:"created_by"`
	LastModified    time.Time              `json:"last_modified"`
	ExpirationTime  time.Time              `json:"expiration_time"`
	RetentionPolicy string                 `json:"retention_policy"`
	ComplianceFlags []string               `json:"compliance_flags"`
}

// CorrelatedEvent represents a comprehensive correlated event with relationship analysis
type CorrelatedEvent struct {
	// Core Identification
	EventID           string        `json:"event_id"`
	CorrelationID     string        `json:"correlation_id"`
	CorrelationType   string        `json:"correlation_type"`
	CorrelationName   string        `json:"correlation_name"`
	Timestamp         time.Time     `json:"timestamp"`
	CorrelationWindow time.Duration `json:"correlation_window"`
	WindowStart       time.Time     `json:"window_start"`
	WindowEnd         time.Time     `json:"window_end"`

	// Event Relationships
	PrimaryEventID        string   `json:"primary_event_id"`
	RelatedEventIDs       []string `json:"related_event_ids"`
	RelatedEventCount     int64    `json:"related_event_count"`
	RelationshipTypes     []string `json:"relationship_types"`
	CorrelationStrength   float64  `json:"correlation_strength"`
	CorrelationConfidence float64  `json:"correlation_confidence"`
	CausalityScore        float64  `json:"causality_score"`
	TemporalRelationship  string   `json:"temporal_relationship"`
	SpatialRelationship   string   `json:"spatial_relationship"`
	LogicalRelationship   string   `json:"logical_relationship"`

	// Correlation Analysis
	CorrelationMethod    string        `json:"correlation_method"`
	CorrelationAlgorithm string        `json:"correlation_algorithm"`
	AnalysisDepth        int32         `json:"analysis_depth"`
	AnalysisScope        string        `json:"analysis_scope"`
	AnalysisAccuracy     float64       `json:"analysis_accuracy"`
	AnalysisConfidence   float64       `json:"analysis_confidence"`
	AnalysisLatency      time.Duration `json:"analysis_latency"`
	AnalysisComplexity   float64       `json:"analysis_complexity"`
	AnalysisReliability  float64       `json:"analysis_reliability"`
	AnalysisValidation   bool          `json:"analysis_validation"`

	// Pattern Recognition
	PatternID             string  `json:"pattern_id"`
	PatternType           string  `json:"pattern_type"`
	PatternSignificance   float64 `json:"pattern_significance"`
	PatternFrequency      int64   `json:"pattern_frequency"`
	PatternStability      float64 `json:"pattern_stability"`
	PatternPredictability float64 `json:"pattern_predictability"`
	PatternEvolution      string  `json:"pattern_evolution"`
	PatternAnomaly        bool    `json:"pattern_anomaly"`
	PatternTrend          string  `json:"pattern_trend"`
	PatternClassification string  `json:"pattern_classification"`

	// Root Cause Analysis
	RootCauseIdentified  bool     `json:"root_cause_identified"`
	RootCauseEventID     string   `json:"root_cause_event_id"`
	RootCauseType        string   `json:"root_cause_type"`
	RootCauseCategory    string   `json:"root_cause_category"`
	RootCauseConfidence  float64  `json:"root_cause_confidence"`
	ContributingFactors  []string `json:"contributing_factors"`
	ImpactChain          []string `json:"impact_chain"`
	PropagationPath      []string `json:"propagation_path"`
	CascadeEffects       []string `json:"cascade_effects"`
	RootCauseDescription string   `json:"root_cause_description"`

	// Impact Analysis
	ImpactScope          string        `json:"impact_scope"`
	ImpactSeverity       string        `json:"impact_severity"`
	ImpactDuration       time.Duration `json:"impact_duration"`
	AffectedSystems      []string      `json:"affected_systems"`
	AffectedUsers        []string      `json:"affected_users"`
	AffectedJobs         []string      `json:"affected_jobs"`
	AffectedResources    []string      `json:"affected_resources"`
	BusinessImpact       float64       `json:"business_impact"`
	FinancialImpact      float64       `json:"financial_impact"`
	OperationalImpact    float64       `json:"operational_impact"`
	UserExperienceImpact float64       `json:"user_experience_impact"`

	// Predictive Analysis
	PredictiveModel        string        `json:"predictive_model"`
	PredictionAccuracy     float64       `json:"prediction_accuracy"`
	PredictionConfidence   float64       `json:"prediction_confidence"`
	PredictionHorizon      time.Duration `json:"prediction_horizon"`
	LikelihoodScore        float64       `json:"likelihood_score"`
	RiskScore              float64       `json:"risk_score"`
	FutureEventPrediction  []string      `json:"future_event_prediction"`
	TrendPrediction        string        `json:"trend_prediction"`
	AnomalyPrediction      bool          `json:"anomaly_prediction"`
	EarlyWarningIndicators []string      `json:"early_warning_indicators"`

	// Machine Learning Insights
	MLModelUsed           string             `json:"ml_model_used"`
	MLModelVersion        string             `json:"ml_model_version"`
	MLModelAccuracy       float64            `json:"ml_model_accuracy"`
	MLModelConfidence     float64            `json:"ml_model_confidence"`
	FeatureImportance     map[string]float64 `json:"feature_importance"`
	ClusteringResults     []string           `json:"clustering_results"`
	ClassificationResults []string           `json:"classification_results"`
	AnomalyScore          float64            `json:"anomaly_score"`
	SimilarityScore       float64            `json:"similarity_score"`
	LearningProgress      float64            `json:"learning_progress"`

	// Remediation and Response
	RemediationRequired      bool     `json:"remediation_required"`
	RemediationPriority      string   `json:"remediation_priority"`
	RemediationActions       []string `json:"remediation_actions"`
	AutomatedRemediation     bool     `json:"automated_remediation"`
	RemediationStatus        string   `json:"remediation_status"`
	RemediationEffectiveness float64  `json:"remediation_effectiveness"`
	PreventiveMeasures       []string `json:"preventive_measures"`
	BestPractices            []string `json:"best_practices"`
	LessonsLearned           []string `json:"lessons_learned"`
	PlaybookRecommendations  []string `json:"playbook_recommendations"`

	// Quality and Validation
	CorrelationQuality      float64   `json:"correlation_quality"`
	DataQuality             float64   `json:"data_quality"`
	AnalysisQuality         float64   `json:"analysis_quality"`
	ResultValidation        bool      `json:"result_validation"`
	PeerValidation          bool      `json:"peer_validation"`
	ExpertValidation        bool      `json:"expert_validation"`
	ConfidenceInterval      []float64 `json:"confidence_interval"`
	StatisticalSignificance float64   `json:"statistical_significance"`
	ReliabilityScore        float64   `json:"reliability_score"`
	QualityAssurance        bool      `json:"quality_assurance"`

	// Performance Metrics
	ProcessingTime      time.Duration `json:"processing_time"`
	MemoryUsage         int64         `json:"memory_usage"`
	CPUUsage            float64       `json:"cpu_usage"`
	IOOperations        int64         `json:"io_operations"`
	NetworkTraffic      int64         `json:"network_traffic"`
	CacheUtilization    float64       `json:"cache_utilization"`
	ComputeEfficiency   float64       `json:"compute_efficiency"`
	ResourceUtilization float64       `json:"resource_utilization"`
	OptimizationLevel   float64       `json:"optimization_level"`
	PerformanceIndex    float64       `json:"performance_index"`

	// Metadata and Context
	Tags            map[string]string      `json:"tags"`
	Attributes      map[string]interface{} `json:"attributes"`
	Context         map[string]string      `json:"context"`
	Environment     string                 `json:"environment"`
	Version         string                 `json:"version"`
	CreatedBy       string                 `json:"created_by"`
	LastModified    time.Time              `json:"last_modified"`
	ExpirationTime  time.Time              `json:"expiration_time"`
	RetentionPolicy string                 `json:"retention_policy"`
	ComplianceFlags []string               `json:"compliance_flags"`
}

// AggregationConfiguration represents configuration for event aggregation
type AggregationConfiguration struct {
	// Basic Configuration
	AggregationEnabled   bool          `json:"aggregation_enabled"`
	AggregationInterval  time.Duration `json:"aggregation_interval"`
	WindowSize           time.Duration `json:"window_size"`
	WindowType           string        `json:"window_type"` // fixed, sliding, session
	OverlapSize          time.Duration `json:"overlap_size"`
	BufferSize           int32         `json:"buffer_size"`
	MaxConcurrentWindows int32         `json:"max_concurrent_windows"`
	ProcessingTimeout    time.Duration `json:"processing_timeout"`

	// Aggregation Methods
	AggregationMethods   []string `json:"aggregation_methods"`   // count, sum, avg, min, max, percentile
	StatisticalFunctions []string `json:"statistical_functions"` // mean, median, mode, stddev, variance
	CustomFunctions      []string `json:"custom_functions"`
	GroupingKeys         []string `json:"grouping_keys"`
	AggregationLevel     string   `json:"aggregation_level"` // event, job, user, account, node, cluster
	AggregationScope     string   `json:"aggregation_scope"` // local, global, distributed

	// Data Sources
	SourceEventTypes []string          `json:"source_event_types"`
	SourceFilters    []string          `json:"source_filters"`
	IncludePatterns  []string          `json:"include_patterns"`
	ExcludePatterns  []string          `json:"exclude_patterns"`
	PriorityFilters  []string          `json:"priority_filters"`
	QualityFilters   []string          `json:"quality_filters"`
	TimeFilters      map[string]string `json:"time_filters"`
	ContentFilters   map[string]string `json:"content_filters"`

	// Processing Options
	RealTimeProcessing    bool `json:"real_time_processing"`
	BatchProcessing       bool `json:"batch_processing"`
	StreamProcessing      bool `json:"stream_processing"`
	ParallelProcessing    bool `json:"parallel_processing"`
	DistributedProcessing bool `json:"distributed_processing"`
	GPUAcceleration       bool `json:"gpu_acceleration"`
	MemoryOptimization    bool `json:"memory_optimization"`
	DiskOptimization      bool `json:"disk_optimization"`
	NetworkOptimization   bool `json:"network_optimization"`
	CachingEnabled        bool `json:"caching_enabled"`

	// Quality Control
	QualityThreshold      float64       `json:"quality_threshold"`
	AccuracyThreshold     float64       `json:"accuracy_threshold"`
	CompletenessThreshold float64       `json:"completeness_threshold"`
	TimelinessThreshold   time.Duration `json:"timeliness_threshold"`
	ConsistencyChecks     bool          `json:"consistency_checks"`
	ValidationRules       []string      `json:"validation_rules"`
	OutlierDetection      bool          `json:"outlier_detection"`
	NoiseReduction        bool          `json:"noise_reduction"`
	DataCleaning          bool          `json:"data_cleaning"`
	ErrorHandling         string        `json:"error_handling"`

	// Output Configuration
	OutputFormats        []string `json:"output_formats"` // json, csv, parquet, avro
	CompressionEnabled   bool     `json:"compression_enabled"`
	EncryptionEnabled    bool     `json:"encryption_enabled"`
	OutputDestinations   []string `json:"output_destinations"`
	DeliveryMethods      []string `json:"delivery_methods"`
	NotificationChannels []string `json:"notification_channels"`
	AlertingEnabled      bool     `json:"alerting_enabled"`
	ReportingEnabled     bool     `json:"reporting_enabled"`
	DashboardIntegration bool     `json:"dashboard_integration"`
	APIEnabled           bool     `json:"api_enabled"`

	// Storage and Retention
	StorageBackend     string            `json:"storage_backend"`
	RetentionPeriod    time.Duration     `json:"retention_period"`
	ArchiveEnabled     bool              `json:"archive_enabled"`
	ArchivePeriod      time.Duration     `json:"archive_period"`
	BackupEnabled      bool              `json:"backup_enabled"`
	BackupFrequency    time.Duration     `json:"backup_frequency"`
	DataLifecycle      string            `json:"data_lifecycle"`
	PurgePolicy        string            `json:"purge_policy"`
	ComplianceSettings map[string]string `json:"compliance_settings"`
	DataGovernance     bool              `json:"data_governance"`

	// Performance Tuning
	MaxMemoryUsage      int64                  `json:"max_memory_usage"`
	MaxCPUUsage         float64                `json:"max_cpu_usage"`
	MaxIOOperations     int64                  `json:"max_io_operations"`
	MaxNetworkBandwidth int64                  `json:"max_network_bandwidth"`
	ResourceLimits      map[string]interface{} `json:"resource_limits"`
	PerformanceTargets  map[string]float64     `json:"performance_targets"`
	ScalingPolicies     []string               `json:"scaling_policies"`
	LoadBalancing       bool                   `json:"load_balancing"`
	FailoverEnabled     bool                   `json:"failover_enabled"`
	DisasterRecovery    bool                   `json:"disaster_recovery"`

	// Security and Privacy
	AccessControl             bool     `json:"access_control"`
	AuthenticationRequired    bool     `json:"authentication_required"`
	AuthorizationRules        []string `json:"authorization_rules"`
	DataAnonymization         bool     `json:"data_anonymization"`
	DataMasking               bool     `json:"data_masking"`
	PIIProtection             bool     `json:"pii_protection"`
	AuditLogging              bool     `json:"audit_logging"`
	SecurityCompliance        []string `json:"security_compliance"`
	PrivacyPolicies           []string `json:"privacy_policies"`
	CertificationRequirements []string `json:"certification_requirements"`

	// Monitoring and Alerting
	HealthMonitoring      bool               `json:"health_monitoring"`
	PerformanceMonitoring bool               `json:"performance_monitoring"`
	QualityMonitoring     bool               `json:"quality_monitoring"`
	CostMonitoring        bool               `json:"cost_monitoring"`
	UsageMonitoring       bool               `json:"usage_monitoring"`
	AlertThresholds       map[string]float64 `json:"alert_thresholds"`
	NotificationRules     []string           `json:"notification_rules"`
	EscalationPolicies    []string           `json:"escalation_policies"`
	IncidentManagement    bool               `json:"incident_management"`
	SLAMonitoring         bool               `json:"sla_monitoring"`
}

// CorrelationConfiguration represents configuration for event correlation
type CorrelationConfiguration struct {
	// Basic Configuration
	CorrelationEnabled  bool          `json:"correlation_enabled"`
	CorrelationInterval time.Duration `json:"correlation_interval"`
	AnalysisWindow      time.Duration `json:"analysis_window"`
	CorrelationDepth    int32         `json:"correlation_depth"`
	MaxCorrelations     int32         `json:"max_correlations"`
	CorrelationTimeout  time.Duration `json:"correlation_timeout"`
	ProcessingThreads   int32         `json:"processing_threads"`
	MemoryBufferSize    int64         `json:"memory_buffer_size"`

	// Correlation Methods
	CorrelationMethods []string `json:"correlation_methods"` // temporal, causal, statistical, pattern, ml
	TemporalMethods    []string `json:"temporal_methods"`    // sequence, proximity, overlap, causality
	StatisticalMethods []string `json:"statistical_methods"` // pearson, spearman, kendall, mutual_info
	PatternMethods     []string `json:"pattern_methods"`     // frequent_patterns, sequential_patterns, association_rules
	MLMethods          []string `json:"ml_methods"`          // clustering, classification, anomaly_detection, graph_neural_networks
	CausalityMethods   []string `json:"causality_methods"`   // granger, pc_algorithm, ges, fci

	// Analysis Configuration
	CorrelationThreshold float64 `json:"correlation_threshold"`
	SignificanceLevel    float64 `json:"significance_level"`
	ConfidenceLevel      float64 `json:"confidence_level"`
	MinimumSupport       float64 `json:"minimum_support"`
	MinimumConfidence    float64 `json:"minimum_confidence"`
	MaxPatternLength     int32   `json:"max_pattern_length"`
	WindowOverlap        float64 `json:"window_overlap"`
	AnalysisAccuracy     float64 `json:"analysis_accuracy"`
	ValidationRequired   bool    `json:"validation_required"`
	CrossValidation      bool    `json:"cross_validation"`

	// Data Sources and Filters
	EventTypes          []string                 `json:"event_types"`
	SourceSystems       []string                 `json:"source_systems"`
	IncludeFilters      []string                 `json:"include_filters"`
	ExcludeFilters      []string                 `json:"exclude_filters"`
	PriorityLevels      []string                 `json:"priority_levels"`
	QualityRequirements []string                 `json:"quality_requirements"`
	TemporalConstraints map[string]time.Duration `json:"temporal_constraints"`
	SpatialConstraints  map[string]string        `json:"spatial_constraints"`
	ContentConstraints  map[string]string        `json:"content_constraints"`
	BusinessRules       []string                 `json:"business_rules"`

	// Machine Learning Configuration
	MLModelsEnabled      bool          `json:"ml_models_enabled"`
	ModelTypes           []string      `json:"model_types"` // supervised, unsupervised, reinforcement, deep_learning
	TrainingEnabled      bool          `json:"training_enabled"`
	TrainingFrequency    time.Duration `json:"training_frequency"`
	TrainingDataSize     int64         `json:"training_data_size"`
	FeatureEngineering   bool          `json:"feature_engineering"`
	FeatureSelection     bool          `json:"feature_selection"`
	HyperparameterTuning bool          `json:"hyperparameter_tuning"`
	ModelValidation      bool          `json:"model_validation"`
	ModelDeployment      bool          `json:"model_deployment"`
	OnlineLearning       bool          `json:"online_learning"`

	// Graph Analysis
	GraphAnalysisEnabled bool     `json:"graph_analysis_enabled"`
	GraphAlgorithms      []string `json:"graph_algorithms"` // pagerank, community_detection, centrality, shortest_path
	NodeAttributes       []string `json:"node_attributes"`
	EdgeAttributes       []string `json:"edge_attributes"`
	GraphMetrics         []string `json:"graph_metrics"`
	CommunityDetection   bool     `json:"community_detection"`
	InfluenceAnalysis    bool     `json:"influence_analysis"`
	NetworkPropagation   bool     `json:"network_propagation"`
	GraphVisualization   bool     `json:"graph_visualization"`
	DynamicGraphs        bool     `json:"dynamic_graphs"`
	GraphDatabase        string   `json:"graph_database"`

	// Performance and Optimization
	ParallelProcessing    bool   `json:"parallel_processing"`
	DistributedProcessing bool   `json:"distributed_processing"`
	GPUAcceleration       bool   `json:"gpu_acceleration"`
	StreamProcessing      bool   `json:"stream_processing"`
	BatchProcessing       bool   `json:"batch_processing"`
	CachingStrategy       string `json:"caching_strategy"`
	IndexingStrategy      string `json:"indexing_strategy"`
	CompressionEnabled    bool   `json:"compression_enabled"`
	OptimizationLevel     string `json:"optimization_level"`
	ResourceManagement    bool   `json:"resource_management"`

	// Output and Reporting
	OutputFormats        []string `json:"output_formats"`
	ReportGeneration     bool     `json:"report_generation"`
	VisualizationEnabled bool     `json:"visualization_enabled"`
	DashboardIntegration bool     `json:"dashboard_integration"`
	AlertingEnabled      bool     `json:"alerting_enabled"`
	NotificationChannels []string `json:"notification_channels"`
	ExportOptions        []string `json:"export_options"`
	APIEnabled           bool     `json:"api_enabled"`
	WebhookSupport       bool     `json:"webhook_support"`
	IntegrationPoints    []string `json:"integration_points"`

	// Quality and Validation
	QualityAssurance      bool     `json:"quality_assurance"`
	ResultValidation      bool     `json:"result_validation"`
	PeerReview            bool     `json:"peer_review"`
	ExpertValidation      bool     `json:"expert_validation"`
	StatisticalValidation bool     `json:"statistical_validation"`
	QualityMetrics        []string `json:"quality_metrics"`
	AccuracyMeasures      []string `json:"accuracy_measures"`
	ReliabilityChecks     bool     `json:"reliability_checks"`
	ConsistencyValidation bool     `json:"consistency_validation"`
	ErrorDetection        bool     `json:"error_detection"`

	// Security and Compliance
	DataPrivacy               bool     `json:"data_privacy"`
	AccessControl             bool     `json:"access_control"`
	AuditTrail                bool     `json:"audit_trail"`
	ComplianceChecks          bool     `json:"compliance_checks"`
	DataGovernance            bool     `json:"data_governance"`
	RetentionPolicies         []string `json:"retention_policies"`
	PrivacyProtection         bool     `json:"privacy_protection"`
	AnonymizationRules        []string `json:"anonymization_rules"`
	SecurityPolicies          []string `json:"security_policies"`
	CertificationRequirements []string `json:"certification_requirements"`
}

// Supporting types for aggregation and correlation operations
type ActiveAggregation struct {
	AggregationID      string             `json:"aggregation_id"`
	AggregationName    string             `json:"aggregation_name"`
	AggregationType    string             `json:"aggregation_type"`
	Status             string             `json:"status"`
	StartTime          time.Time          `json:"start_time"`
	EndTime            time.Time          `json:"end_time"`
	EventsProcessed    int64              `json:"events_processed"`
	ResultsGenerated   int64              `json:"results_generated"`
	ProcessingRate     float64            `json:"processing_rate"`
	QualityScore       float64            `json:"quality_score"`
	PerformanceMetrics map[string]float64 `json:"performance_metrics"`
}

type ActiveCorrelation struct {
	CorrelationID     string             `json:"correlation_id"`
	CorrelationName   string             `json:"correlation_name"`
	CorrelationType   string             `json:"correlation_type"`
	Status            string             `json:"status"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           time.Time          `json:"end_time"`
	EventsAnalyzed    int64              `json:"events_analyzed"`
	CorrelationsFound int64              `json:"correlations_found"`
	AnalysisDepth     int32              `json:"analysis_depth"`
	ConfidenceScore   float64            `json:"confidence_score"`
	AccuracyMetrics   map[string]float64 `json:"accuracy_metrics"`
}

type AggregationRecord struct {
	RecordID      string                 `json:"record_id"`
	AggregationID string                 `json:"aggregation_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Results       map[string]interface{} `json:"results"`
	Metadata      map[string]string      `json:"metadata"`
}

type CorrelationRecord struct {
	RecordID      string                 `json:"record_id"`
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Relationships map[string]interface{} `json:"relationships"`
	Confidence    float64                `json:"confidence"`
	Metadata      map[string]string      `json:"metadata"`
}

type AggregationMetrics struct {
	TotalAggregations     int64              `json:"total_aggregations"`
	ActiveAggregations    int64              `json:"active_aggregations"`
	CompletedAggregations int64              `json:"completed_aggregations"`
	FailedAggregations    int64              `json:"failed_aggregations"`
	AverageProcessingTime time.Duration      `json:"average_processing_time"`
	ThroughputRate        float64            `json:"throughput_rate"`
	QualityScore          float64            `json:"quality_score"`
	ResourceUtilization   map[string]float64 `json:"resource_utilization"`
}

type CorrelationMetrics struct {
	TotalCorrelations     int64              `json:"total_correlations"`
	ActiveCorrelations    int64              `json:"active_correlations"`
	CompletedCorrelations int64              `json:"completed_correlations"`
	FailedCorrelations    int64              `json:"failed_correlations"`
	AverageAnalysisTime   time.Duration      `json:"average_analysis_time"`
	AccuracyRate          float64            `json:"accuracy_rate"`
	ConfidenceScore       float64            `json:"confidence_score"`
	ResourceUtilization   map[string]float64 `json:"resource_utilization"`
}

// EventAggregationCorrelationCollector collects event aggregation and correlation metrics
type EventAggregationCorrelationCollector struct {
	client EventAggregationCorrelationSLURMClient

	// Aggregation metrics
	aggregatedEvents      *prometheus.CounterVec
	activeAggregations    *prometheus.GaugeVec
	aggregationLatency    *prometheus.HistogramVec
	aggregationQuality    *prometheus.GaugeVec
	aggregationThroughput *prometheus.GaugeVec
	aggregationErrors     *prometheus.CounterVec

	// Correlation metrics
	correlatedEvents      *prometheus.CounterVec
	activeCorrelations    *prometheus.GaugeVec
	correlationLatency    *prometheus.HistogramVec
	correlationAccuracy   *prometheus.GaugeVec
	correlationConfidence *prometheus.GaugeVec
	correlationComplexity *prometheus.GaugeVec

	// Performance metrics
	processingTime      *prometheus.HistogramVec
	resourceUtilization *prometheus.GaugeVec
	dataQuality         *prometheus.GaugeVec
	systemHealth        *prometheus.GaugeVec

	// Business impact metrics
	businessValue         *prometheus.GaugeVec
	costOptimization      *prometheus.GaugeVec
	operationalEfficiency *prometheus.GaugeVec
	decisionSupport       *prometheus.GaugeVec

	// Quality and reliability metrics
	dataCompleteness    *prometheus.GaugeVec
	analysisReliability *prometheus.GaugeVec
	resultValidation    *prometheus.GaugeVec
	performanceIndex    *prometheus.GaugeVec
}

// NewEventAggregationCorrelationCollector creates a new collector for event aggregation and correlation metrics
func NewEventAggregationCorrelationCollector(client EventAggregationCorrelationSLURMClient) *EventAggregationCorrelationCollector {
	return &EventAggregationCorrelationCollector{
		client: client,

		// Aggregation metrics
		aggregatedEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "events_total",
				Help:      "Total number of aggregated events processed",
			},
			[]string{"aggregation_type", "aggregation_method", "status", "quality_level"},
		),

		activeAggregations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "active_aggregations",
				Help:      "Number of currently active event aggregations",
			},
			[]string{"aggregation_type", "scope", "status"},
		),

		aggregationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "latency_seconds",
				Help:      "Time taken to complete event aggregation operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"aggregation_type", "complexity_level"},
		),

		aggregationQuality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "quality_score",
				Help:      "Quality score of aggregation results",
			},
			[]string{"aggregation_type", "metric_type"},
		),

		aggregationThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "throughput_events_per_second",
				Help:      "Event aggregation processing throughput",
			},
			[]string{"aggregation_type", "processing_mode"},
		),

		aggregationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "slurm",
				Subsystem: "event_aggregation",
				Name:      "errors_total",
				Help:      "Total number of aggregation errors",
			},
			[]string{"error_type", "aggregation_type", "severity"},
		),

		// Correlation metrics
		correlatedEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "events_total",
				Help:      "Total number of correlated events processed",
			},
			[]string{"correlation_type", "correlation_method", "relationship_type", "confidence_level"},
		),

		activeCorrelations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "active_correlations",
				Help:      "Number of currently active event correlations",
			},
			[]string{"correlation_type", "analysis_depth", "status"},
		),

		correlationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "latency_seconds",
				Help:      "Time taken to complete event correlation analysis",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"correlation_type", "complexity_level"},
		),

		correlationAccuracy: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "accuracy_score",
				Help:      "Accuracy score of correlation analysis",
			},
			[]string{"correlation_method", "analysis_type"},
		),

		correlationConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "confidence_score",
				Help:      "Confidence score of correlation results",
			},
			[]string{"correlation_type", "relationship_type"},
		),

		correlationComplexity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_correlation",
				Name:      "complexity_score",
				Help:      "Complexity score of correlation analysis",
			},
			[]string{"correlation_method", "analysis_depth"},
		),

		// Performance metrics
		processingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm",
				Subsystem: "event_processing",
				Name:      "processing_time_seconds",
				Help:      "Time taken to process aggregation and correlation operations",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"operation_type", "processing_mode"},
		),

		resourceUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_processing",
				Name:      "resource_utilization_ratio",
				Help:      "Resource utilization for event processing operations",
			},
			[]string{"resource_type", "operation_type"},
		),

		dataQuality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_processing",
				Name:      "data_quality_score",
				Help:      "Data quality score for processed events",
			},
			[]string{"quality_dimension", "data_source"},
		),

		systemHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_processing",
				Name:      "system_health_score",
				Help:      "Overall system health score for event processing",
			},
			[]string{"component", "health_aspect"},
		),

		// Business impact metrics
		businessValue: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_business",
				Name:      "value_score",
				Help:      "Business value score from event aggregation and correlation",
			},
			[]string{"value_type", "business_area"},
		),

		costOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_business",
				Name:      "cost_optimization_score",
				Help:      "Cost optimization score from event insights",
			},
			[]string{"optimization_type", "cost_category"},
		),

		operationalEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_business",
				Name:      "operational_efficiency_score",
				Help:      "Operational efficiency score from event analysis",
			},
			[]string{"efficiency_type", "operational_area"},
		),

		decisionSupport: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_business",
				Name:      "decision_support_score",
				Help:      "Decision support quality score from event insights",
			},
			[]string{"decision_type", "support_level"},
		),

		// Quality and reliability metrics
		dataCompleteness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_quality",
				Name:      "data_completeness_ratio",
				Help:      "Data completeness ratio for event processing",
			},
			[]string{"data_type", "completeness_aspect"},
		),

		analysisReliability: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_quality",
				Name:      "analysis_reliability_score",
				Help:      "Reliability score of event analysis operations",
			},
			[]string{"analysis_type", "reliability_aspect"},
		),

		resultValidation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_quality",
				Name:      "result_validation_score",
				Help:      "Validation score for event processing results",
			},
			[]string{"validation_type", "result_category"},
		),

		performanceIndex: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "slurm",
				Subsystem: "event_quality",
				Name:      "performance_index",
				Help:      "Overall performance index for event aggregation and correlation",
			},
			[]string{"performance_aspect", "measurement_type"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (c *EventAggregationCorrelationCollector) Describe(ch chan<- *prometheus.Desc) {
	c.aggregatedEvents.Describe(ch)
	c.activeAggregations.Describe(ch)
	c.aggregationLatency.Describe(ch)
	c.aggregationQuality.Describe(ch)
	c.aggregationThroughput.Describe(ch)
	c.aggregationErrors.Describe(ch)
	c.correlatedEvents.Describe(ch)
	c.activeCorrelations.Describe(ch)
	c.correlationLatency.Describe(ch)
	c.correlationAccuracy.Describe(ch)
	c.correlationConfidence.Describe(ch)
	c.correlationComplexity.Describe(ch)
	c.processingTime.Describe(ch)
	c.resourceUtilization.Describe(ch)
	c.dataQuality.Describe(ch)
	c.systemHealth.Describe(ch)
	c.businessValue.Describe(ch)
	c.costOptimization.Describe(ch)
	c.operationalEfficiency.Describe(ch)
	c.decisionSupport.Describe(ch)
	c.dataCompleteness.Describe(ch)
	c.analysisReliability.Describe(ch)
	c.resultValidation.Describe(ch)
	c.performanceIndex.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *EventAggregationCorrelationCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	// Collect aggregation configuration metrics
	if config, err := c.client.GetAggregationConfiguration(ctx); err == nil && config != nil {
		c.collectAggregationConfiguration(config)
	}

	// Collect correlation configuration metrics
	if config, err := c.client.GetCorrelationConfiguration(ctx); err == nil && config != nil {
		c.collectCorrelationConfiguration(config)
	}

	// Collect active aggregations
	if aggregations, err := c.client.GetActiveAggregations(ctx); err == nil {
		c.collectActiveAggregations(aggregations)
	}

	// Collect active correlations
	if correlations, err := c.client.GetActiveCorrelations(ctx); err == nil {
		c.collectActiveCorrelations(correlations)
	}

	// Collect aggregation metrics
	if metrics, err := c.client.GetAggregationMetrics(ctx); err == nil && metrics != nil {
		c.collectAggregationMetrics(metrics)
	}

	// Collect correlation metrics
	if metrics, err := c.client.GetCorrelationMetrics(ctx); err == nil && metrics != nil {
		c.collectCorrelationMetrics(metrics)
	}

	// Export all metrics
	c.aggregatedEvents.Collect(ch)
	c.activeAggregations.Collect(ch)
	c.aggregationLatency.Collect(ch)
	c.aggregationQuality.Collect(ch)
	c.aggregationThroughput.Collect(ch)
	c.aggregationErrors.Collect(ch)
	c.correlatedEvents.Collect(ch)
	c.activeCorrelations.Collect(ch)
	c.correlationLatency.Collect(ch)
	c.correlationAccuracy.Collect(ch)
	c.correlationConfidence.Collect(ch)
	c.correlationComplexity.Collect(ch)
	c.processingTime.Collect(ch)
	c.resourceUtilization.Collect(ch)
	c.dataQuality.Collect(ch)
	c.systemHealth.Collect(ch)
	c.businessValue.Collect(ch)
	c.costOptimization.Collect(ch)
	c.operationalEfficiency.Collect(ch)
	c.decisionSupport.Collect(ch)
	c.dataCompleteness.Collect(ch)
	c.analysisReliability.Collect(ch)
	c.resultValidation.Collect(ch)
	c.performanceIndex.Collect(ch)
}

func (c *EventAggregationCorrelationCollector) collectAggregationConfiguration(config *AggregationConfiguration) {
	// Collect aggregation configuration metrics
	if config.AggregationEnabled {
		methodsStr := strings.Join(config.AggregationMethods, ",")
		c.aggregationQuality.WithLabelValues("configuration", "enabled").Set(1.0)
		c.aggregationThroughput.WithLabelValues(methodsStr, "configuration").Set(float64(config.MaxConcurrentWindows))
	}
}

func (c *EventAggregationCorrelationCollector) collectCorrelationConfiguration(config *CorrelationConfiguration) {
	// Collect correlation configuration metrics
	if config.CorrelationEnabled {
		methodsStr := strings.Join(config.CorrelationMethods, ",")
		c.correlationAccuracy.WithLabelValues(methodsStr, "configuration").Set(config.CorrelationThreshold)
		c.correlationConfidence.WithLabelValues("configuration", "threshold").Set(config.ConfidenceLevel)
	}
}

func (c *EventAggregationCorrelationCollector) collectActiveAggregations(aggregations []*ActiveAggregation) {
	// Reset active aggregation metrics
	c.activeAggregations.Reset()

	for _, agg := range aggregations {
		c.activeAggregations.WithLabelValues(agg.AggregationType, "active", agg.Status).Set(1.0)
		c.aggregationQuality.WithLabelValues(agg.AggregationType, "quality").Set(agg.QualityScore)
		c.aggregationThroughput.WithLabelValues(agg.AggregationType, "processing").Set(agg.ProcessingRate)
	}
}

func (c *EventAggregationCorrelationCollector) collectActiveCorrelations(correlations []*ActiveCorrelation) {
	// Reset active correlation metrics
	c.activeCorrelations.Reset()

	for _, corr := range correlations {
		depthStr := strconv.Itoa(int(corr.AnalysisDepth))
		c.activeCorrelations.WithLabelValues(corr.CorrelationType, depthStr, corr.Status).Set(1.0)
		c.correlationConfidence.WithLabelValues(corr.CorrelationType, "analysis").Set(corr.ConfidenceScore)
		c.correlationComplexity.WithLabelValues(corr.CorrelationType, depthStr).Set(float64(corr.AnalysisDepth))
	}
}

func (c *EventAggregationCorrelationCollector) collectAggregationMetrics(metrics *AggregationMetrics) {
	c.aggregatedEvents.WithLabelValues("total", "all", "completed", "high").Add(float64(metrics.CompletedAggregations))
	c.aggregationLatency.WithLabelValues("standard", "normal").Observe(metrics.AverageProcessingTime.Seconds())
	c.aggregationThroughput.WithLabelValues("system", "overall").Set(metrics.ThroughputRate)
	c.aggregationQuality.WithLabelValues("system", "overall").Set(metrics.QualityScore)

	for resource, utilization := range metrics.ResourceUtilization {
		c.resourceUtilization.WithLabelValues(resource, "aggregation").Set(utilization)
	}
}

func (c *EventAggregationCorrelationCollector) collectCorrelationMetrics(metrics *CorrelationMetrics) {
	c.correlatedEvents.WithLabelValues("total", "all", "completed", "high").Add(float64(metrics.CompletedCorrelations))
	c.correlationLatency.WithLabelValues("standard", "normal").Observe(metrics.AverageAnalysisTime.Seconds())
	c.correlationAccuracy.WithLabelValues("system", "overall").Set(metrics.AccuracyRate)
	c.correlationConfidence.WithLabelValues("system", "overall").Set(metrics.ConfidenceScore)

	for resource, utilization := range metrics.ResourceUtilization {
		c.resourceUtilization.WithLabelValues(resource, "correlation").Set(utilization)
	}
}
