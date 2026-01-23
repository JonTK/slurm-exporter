// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// QoSLimitsSLURMClient defines the interface for SLURM client operations related to QoS limits
type QoSLimitsSLURMClient interface {
	GetQoSLimits(ctx context.Context, qosName string) (*QoSResourceLimits, error)
	GetQoSViolations(ctx context.Context, opts *QoSViolationOptions) (*QoSViolations, error)
	GetQoSUsage(ctx context.Context, qosName string) (*QoSUsage, error)
	GetQoSConfiguration(ctx context.Context, qosName string) (*QoSConfiguration, error)
	GetQoSHierarchy(ctx context.Context) (*QoSHierarchy, error)
	GetQoSPriorities(ctx context.Context) (*QoSPriorities, error)
	GetQoSAssignments(ctx context.Context, entityType string) (*QoSAssignments, error)
	GetQoSEnforcement(ctx context.Context, qosName string) (*QoSEnforcement, error)
	GetQoSStatistics(ctx context.Context, qosName string, period string) (*QoSStatistics, error)
	GetQoSEffectiveness(ctx context.Context, qosName string) (*QoSEffectiveness, error)
	GetQoSConflicts(ctx context.Context) (*QoSConflicts, error)
	GetSystemQoSOverview(ctx context.Context) (*SystemQoSOverview, error)
}

// QoSResourceLimits represents Quality of Service resource limits
type QoSResourceLimits struct {
	QoSName              string
	Description          string
	Priority             int
	UsageFactor          float64
	UsageThreshold       float64
	GrpCPUs              int
	GrpCPUsRunning       int
	GrpMem               int64
	GrpMemRunning        int64
	GrpNodes             int
	GrpNodesRunning      int
	GrpJobs              int
	GrpJobsRunning       int
	GrpSubmitJobs        int
	GrpWall              time.Duration
	MaxCPUs              int
	MaxCPUsPerUser       int
	MaxMem               int64
	MaxMemPerUser        int64
	MaxNodes             int
	MaxNodesPerUser      int
	MaxJobs              int
	MaxJobsPerUser       int
	MaxSubmitJobs        int
	MaxSubmitJobsPerUser int
	MaxWall              time.Duration
	MaxWallPerJob        time.Duration
	MinCPUs              int
	MinMem               int64
	MinNodes             int
	Preempt              []string
	PreemptMode          string
	Flags                []string
	GraceTime            time.Duration
	OverPartQoS          string
	PartitionQoS         string
	CreatedAt            time.Time
	ModifiedAt           time.Time
	LastUsed             time.Time
	UserCount            int
	AccountCount         int
	ActiveJobs           int
	PendingJobs          int
	RunningJobs          int
}

// QoSViolationOptions represents options for querying QoS violations
type QoSViolationOptions struct {
	QoSName    string
	Severity   string
	TimeRange  string
	EntityType string
	EntityID   string
	Status     string
}

// QoSViolations represents detected QoS violations
type QoSViolations struct {
	TotalViolations    int
	ActiveViolations   int
	ResolvedViolations int
	CriticalViolations int
	WarningViolations  int
	InfoViolations     int
	Violations         []QoSLimitViolation
	ViolationsByQoS    map[string]int
	ViolationsByType   map[string]int
	ViolationsByUser   map[string]int
	ViolationTrends    []int
	ResolutionStats    ViolationResolutionStats
}

// QoSLimitViolation represents a single QoS violation
type QoSLimitViolation struct {
	ViolationID      string
	Timestamp        time.Time
	QoSName          string
	ViolationType    string
	Severity         string
	EntityType       string
	EntityID         string
	LimitType        string
	LimitValue       interface{}
	ActualValue      interface{}
	ExcessAmount     float64
	Duration         time.Duration
	Status           string
	AutoResolved     bool
	ResolutionTime   time.Time
	ResolutionAction string
	Impact           string
	RootCause        string
	PreventionSteps  []string
}

// ViolationResolutionStats represents violation resolution statistics
type ViolationResolutionStats struct {
	MeanResolutionTime   time.Duration
	MedianResolutionTime time.Duration
	AutoResolutionRate   float64
	ManualResolutionRate float64
	EscalationRate       float64
	RecurrenceRate       float64
}

// QoSUsage represents current QoS usage
type QoSUsage struct {
	QoSName               string
	CPUsInUse             int
	CPUsAllocated         int
	CPUUtilization        float64
	MemoryInUse           int64
	MemoryAllocated       int64
	MemoryUtilization     float64
	NodesInUse            int
	NodesAllocated        int
	NodeUtilization       float64
	JobsRunning           int
	JobsPending           int
	JobsCompleted         int
	JobsFailed            int
	UsersActive           int
	AccountsActive        int
	WalltimeConsumed      time.Duration
	WalltimeAllocated     time.Duration
	WalltimeUtilization   float64
	EfficiencyScore       float64
	ThroughputScore       float64
	LoadFactor            float64
	QueueDepth            int
	WaitTimeAverage       time.Duration
	WaitTimeMedian        time.Duration
	TurnaroundTimeAverage time.Duration
}

// QoSConfiguration represents QoS configuration details
type QoSConfiguration struct {
	QoSName           string
	ConfigVersion     string
	LastModified      time.Time
	ModifiedBy        string
	InheritanceRules  map[string]bool
	OverrideRules     map[string]interface{}
	EnforcementPolicy string
	ViolationHandling string
	EscalationRules   []EscalationRule
	NotificationRules []NotificationRule
	AuditSettings     AuditSettings
	ComplianceLevel   string
	PerformanceTarget float64
}

// EscalationRule represents escalation configuration
type EscalationRule struct {
	TriggerCondition string
	EscalationLevel  int
	WaitTime         time.Duration
	Action           string
	NotifyUsers      []string
}

// NotificationRule represents notification configuration
type NotificationRule struct {
	EventType     string
	Severity      string
	Recipients    []string
	MessageFormat string
	Frequency     string
	Channels      []string
}

// AuditSettings represents audit configuration
type AuditSettings struct {
	Enabled         bool
	LogLevel        string
	RetentionPeriod time.Duration
	ArchiveLocation string
	ComplianceFlags []string
}

// QoSHierarchy represents the QoS hierarchy structure
type QoSHierarchy struct {
	RootQoS          []string
	QoSTree          map[string][]string
	InheritanceChain map[string][]string
	PriorityOrder    []string
	ConflictRules    map[string]string
	DefaultQoS       string
	MaxDepth         int
	TotalQoS         int
}

// QoSPriorities represents QoS priority assignments
type QoSPriorities struct {
	PriorityMap          map[string]int
	NormalizedPriorities map[string]float64
	PriorityGroups       map[string][]string
	PreemptionMatrix     map[string][]string
	PriorityTrends       map[string][]float64
	LastUpdated          time.Time
}

// QoSAssignments represents QoS assignments to entities
type QoSAssignments struct {
	EntityType             string
	Assignments            []QoSAssignment
	TotalAssignments       int
	DefaultAssigned        int
	ExplicitAssigned       int
	InheritedAssigned      int
	ConflictingAssignments []AssignmentConflict
}

// QoSAssignment represents a single QoS assignment
type QoSAssignment struct {
	EntityID       string
	QoSName        string
	AssignmentType string
	EffectiveFrom  time.Time
	EffectiveTo    time.Time
	InheritedFrom  string
	Override       bool
	Priority       int
	Status         string
}

// AssignmentConflict represents conflicting QoS assignments
type AssignmentConflict struct {
	EntityID       string
	ConflictingQoS []string
	ConflictType   string
	ResolutionRule string
	ResolvedQoS    string
}

// QoSEnforcement represents QoS enforcement status
type QoSEnforcement struct {
	QoSName            string
	EnforcementEnabled bool
	EnforcementLevel   string
	ViolationActions   map[string]string
	GracePeriods       map[string]time.Duration
	WarningThresholds  map[string]float64
	CriticalThresholds map[string]float64
	AutoRemediation    bool
	ManualIntervention int
	EnforcementStats   EnforcementStats
}

// EnforcementStats represents enforcement statistics
type EnforcementStats struct {
	ActionsTotal       int
	ActionsSuccessful  int
	ActionsFailed      int
	PreemptionsTotal   int
	JobsTerminated     int
	JobsSuspended      int
	JobsRequeued       int
	UsersWarned        int
	UsersBlocked       int
	EffectivenessScore float64
}

// QoSStatistics represents QoS performance statistics
type QoSStatistics struct {
	QoSName               string
	Period                string
	JobsSubmitted         int
	JobsCompleted         int
	JobsFailed            int
	JobsCanceled          int
	JobsPreempted         int
	AverageQueueTime      time.Duration
	AverageRunTime        time.Duration
	AverageTurnaroundTime time.Duration
	ResourceEfficiency    float64
	Throughput            float64
	UtilizationRate       float64
	SLACompliance         float64
	PerformanceScore      float64
	UserSatisfaction      float64
	CostEffectiveness     float64
	TrendDirection        string
	PreviousPeriodChange  float64
}

// QoSEffectiveness represents QoS effectiveness analysis
type QoSEffectiveness struct {
	QoSName              string
	OverallEffectiveness float64
	ResourceOptimization float64
	FairnessScore        float64
	PerformanceImpact    float64
	UserProductivity     float64
	SystemThroughput     float64
	PolicyCompliance     float64
	CostBenefitRatio     float64
	RecommendedActions   []string
	OptimizationTips     []string
	ComparisonBaseline   string
	EffectivenessTrend   []float64
	NextReviewDate       time.Time
}

// QoSConflicts represents detected QoS conflicts
type QoSConflicts struct {
	TotalConflicts    int
	CriticalConflicts int
	ConflictsByType   map[string]int
	Conflicts         []QoSConflict
	ResolutionMatrix  map[string]string
	LastAnalyzed      time.Time
}

// QoSConflict represents a single QoS conflict
type QoSConflict struct {
	ConflictID       string
	ConflictType     string
	QoSNames         []string
	AffectedEntities []string
	Description      string
	Impact           string
	ResolutionRule   string
	Status           string
	DetectedAt       time.Time
	ResolvedAt       time.Time
}

// SystemQoSOverview represents system-wide QoS overview
type SystemQoSOverview struct {
	TotalQoS                 int
	ActiveQoS                int
	DefaultQoS               string
	HighestPriorityQoS       string
	LowestPriorityQoS        string
	QoSUtilization           map[string]float64
	SystemLoad               float64
	OverallEfficiency        float64
	ViolationRate            float64
	ComplianceScore          float64
	RecommendedOptimizations []string
	SystemHealth             string
	LastUpdated              time.Time
}

// QoSLimitsCollector collects QoS limits and violations from SLURM
type QoSLimitsCollector struct {
	client QoSLimitsSLURMClient
	mutex  sync.RWMutex

	// QoS limit metrics
	qosLimitCPUs        *prometheus.GaugeVec
	qosLimitMemory      *prometheus.GaugeVec
	qosLimitNodes       *prometheus.GaugeVec
	qosLimitJobs        *prometheus.GaugeVec
	qosLimitWalltime    *prometheus.GaugeVec
	qosMaxCPUsPerUser   *prometheus.GaugeVec
	qosMaxMemoryPerUser *prometheus.GaugeVec
	qosMaxJobsPerUser   *prometheus.GaugeVec
	qosMinCPUs          *prometheus.GaugeVec
	qosMinMemory        *prometheus.GaugeVec
	qosMinNodes         *prometheus.GaugeVec

	// QoS priority and configuration metrics
	qosPriority       *prometheus.GaugeVec
	qosUsageFactor    *prometheus.GaugeVec
	qosUsageThreshold *prometheus.GaugeVec
	qosGraceTime      *prometheus.GaugeVec
	qosUserCount      *prometheus.GaugeVec
	qosAccountCount   *prometheus.GaugeVec

	// QoS usage metrics
	qosResourceUsage       *prometheus.GaugeVec
	qosResourceUtilization *prometheus.GaugeVec
	qosJobsRunning         *prometheus.GaugeVec
	qosJobsPending         *prometheus.GaugeVec
	qosJobsCompleted       *prometheus.CounterVec
	qosJobsFailed          *prometheus.CounterVec
	qosUsersActive         *prometheus.GaugeVec
	qosWalltimeConsumed    *prometheus.CounterVec
	qosEfficiencyScore     *prometheus.GaugeVec

	// QoS violation metrics
	qosViolations              *prometheus.CounterVec
	qosViolationSeverity       *prometheus.GaugeVec
	qosActiveViolations        *prometheus.GaugeVec
	qosViolationDuration       *prometheus.HistogramVec
	qosViolationResolutionTime *prometheus.HistogramVec
	qosAutoResolutionRate      *prometheus.GaugeVec
	qosRecurrenceRate          *prometheus.GaugeVec

	// QoS enforcement metrics
	qosEnforcementEnabled       *prometheus.GaugeVec
	qosEnforcementActions       *prometheus.CounterVec
	qosPreemptions              *prometheus.CounterVec
	qosJobsTerminated           *prometheus.CounterVec
	qosJobsSuspended            *prometheus.CounterVec
	qosUsersWarned              *prometheus.CounterVec
	qosEnforcementEffectiveness *prometheus.GaugeVec

	// QoS performance metrics
	qosAverageQueueTime  *prometheus.GaugeVec
	qosAverageRunTime    *prometheus.GaugeVec
	qosThroughput        *prometheus.GaugeVec
	qosSLACompliance     *prometheus.GaugeVec
	qosPerformanceScore  *prometheus.GaugeVec
	qosUserSatisfaction  *prometheus.GaugeVec
	qosCostEffectiveness *prometheus.GaugeVec

	// QoS hierarchy metrics
	qosHierarchyDepth         *prometheus.GaugeVec
	qosTotalCount             *prometheus.GaugeVec
	qosInheritanceChainLength *prometheus.GaugeVec
	qosConflicts              *prometheus.GaugeVec

	// QoS effectiveness metrics
	qosOverallEffectiveness *prometheus.GaugeVec
	qosResourceOptimization *prometheus.GaugeVec
	qosFairnessScore        *prometheus.GaugeVec
	qosSystemThroughput     *prometheus.GaugeVec
	qosPolicyCompliance     *prometheus.GaugeVec
	qosCostBenefitRatio     *prometheus.GaugeVec

	// System QoS metrics
	systemQoSUtilization     *prometheus.GaugeVec
	systemQoSLoad            *prometheus.GaugeVec
	systemQoSEfficiency      *prometheus.GaugeVec
	systemQoSViolationRate   *prometheus.GaugeVec
	systemQoSComplianceScore *prometheus.GaugeVec

	// Collection metrics
	collectionDuration *prometheus.HistogramVec
	collectionErrors   *prometheus.CounterVec
	lastCollectionTime *prometheus.GaugeVec
}

// NewQoSLimitsCollector creates a new QoS limits collector
func NewQoSLimitsCollector(client QoSLimitsSLURMClient) *QoSLimitsCollector {
	return &QoSLimitsCollector{
		client: client,

		// QoS limit metrics
		qosLimitCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limit_cpus",
				Help: "CPU limit for QoS",
			},
			[]string{"qos", "limit_type"},
		),
		qosLimitMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limit_memory_bytes",
				Help: "Memory limit for QoS in bytes",
			},
			[]string{"qos", "limit_type"},
		),
		qosLimitNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limit_nodes",
				Help: "Node limit for QoS",
			},
			[]string{"qos", "limit_type"},
		),
		qosLimitJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limit_jobs",
				Help: "Job limit for QoS",
			},
			[]string{"qos", "limit_type"},
		),
		qosLimitWalltime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limit_walltime_seconds",
				Help: "Walltime limit for QoS in seconds",
			},
			[]string{"qos", "limit_type"},
		),
		qosMaxCPUsPerUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_max_cpus_per_user",
				Help: "Maximum CPUs per user for QoS",
			},
			[]string{"qos"},
		),
		qosMaxMemoryPerUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_max_memory_per_user_bytes",
				Help: "Maximum memory per user for QoS in bytes",
			},
			[]string{"qos"},
		),
		qosMaxJobsPerUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_max_jobs_per_user",
				Help: "Maximum jobs per user for QoS",
			},
			[]string{"qos"},
		),
		qosMinCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_min_cpus",
				Help: "Minimum CPUs for QoS",
			},
			[]string{"qos"},
		),
		qosMinMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_min_memory_bytes",
				Help: "Minimum memory for QoS in bytes",
			},
			[]string{"qos"},
		),
		qosMinNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_min_nodes",
				Help: "Minimum nodes for QoS",
			},
			[]string{"qos"},
		),

		// QoS priority and configuration metrics
		qosPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_priority",
				Help: "Priority value for QoS",
			},
			[]string{"qos"},
		),
		qosUsageFactor: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_usage_factor",
				Help: "Usage factor for QoS",
			},
			[]string{"qos"},
		),
		qosUsageThreshold: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_usage_threshold",
				Help: "Usage threshold for QoS",
			},
			[]string{"qos"},
		),
		qosGraceTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_grace_time_seconds",
				Help: "Grace time for QoS in seconds",
			},
			[]string{"qos"},
		),
		qosUserCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_user_count",
				Help: "Number of users assigned to QoS",
			},
			[]string{"qos"},
		),
		qosAccountCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_account_count",
				Help: "Number of accounts assigned to QoS",
			},
			[]string{"qos"},
		),

		// QoS usage metrics
		qosResourceUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_resource_usage",
				Help: "Current resource usage for QoS",
			},
			[]string{"qos", "resource", "usage_type"},
		),
		qosResourceUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_resource_utilization_ratio",
				Help: "Resource utilization ratio for QoS (0-1)",
			},
			[]string{"qos", "resource"},
		),
		qosJobsRunning: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_jobs_running",
				Help: "Number of running jobs for QoS",
			},
			[]string{"qos"},
		),
		qosJobsPending: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_jobs_pending",
				Help: "Number of pending jobs for QoS",
			},
			[]string{"qos"},
		),
		qosJobsCompleted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_jobs_completed_total",
				Help: "Total completed jobs for QoS",
			},
			[]string{"qos"},
		),
		qosJobsFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_jobs_failed_total",
				Help: "Total failed jobs for QoS",
			},
			[]string{"qos"},
		),
		qosUsersActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_users_active",
				Help: "Number of active users for QoS",
			},
			[]string{"qos"},
		),
		qosWalltimeConsumed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_walltime_consumed_seconds_total",
				Help: "Total walltime consumed for QoS in seconds",
			},
			[]string{"qos"},
		),
		qosEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_efficiency_score",
				Help: "Efficiency score for QoS (0-1)",
			},
			[]string{"qos"},
		),

		// QoS violation metrics
		qosViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_violations_total",
				Help: "Total QoS violations",
			},
			[]string{"qos", "violation_type", "severity"},
		),
		qosViolationSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_violation_severity_distribution",
				Help: "Distribution of QoS violations by severity",
			},
			[]string{"qos", "severity"},
		),
		qosActiveViolations: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_active_violations",
				Help: "Number of active QoS violations",
			},
			[]string{"qos", "violation_type"},
		),
		qosViolationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_qos_violation_duration_seconds",
				Help:    "Duration of QoS violations",
				Buckets: prometheus.ExponentialBuckets(60, 2, 10),
			},
			[]string{"qos", "violation_type"},
		),
		qosViolationResolutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_qos_violation_resolution_time_seconds",
				Help:    "Time to resolve QoS violations",
				Buckets: prometheus.ExponentialBuckets(60, 2, 10),
			},
			[]string{"qos", "violation_type"},
		),
		qosAutoResolutionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_auto_resolution_rate",
				Help: "Automatic resolution rate for QoS violations",
			},
			[]string{"qos"},
		),
		qosRecurrenceRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_violation_recurrence_rate",
				Help: "Recurrence rate for QoS violations",
			},
			[]string{"qos"},
		),

		// QoS enforcement metrics
		qosEnforcementEnabled: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_enforcement_enabled",
				Help: "Whether QoS enforcement is enabled (1 = yes, 0 = no)",
			},
			[]string{"qos"},
		),
		qosEnforcementActions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_enforcement_actions_total",
				Help: "Total QoS enforcement actions",
			},
			[]string{"qos", "action_type", "result"},
		),
		qosPreemptions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_preemptions_total",
				Help: "Total QoS preemptions",
			},
			[]string{"qos", "preempt_mode"},
		),
		qosJobsTerminated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_jobs_terminated_total",
				Help: "Total jobs terminated due to QoS violations",
			},
			[]string{"qos", "reason"},
		),
		qosJobsSuspended: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_jobs_suspended_total",
				Help: "Total jobs suspended due to QoS violations",
			},
			[]string{"qos", "reason"},
		),
		qosUsersWarned: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_users_warned_total",
				Help: "Total users warned for QoS violations",
			},
			[]string{"qos", "warning_type"},
		),
		qosEnforcementEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_enforcement_effectiveness",
				Help: "Effectiveness of QoS enforcement (0-1)",
			},
			[]string{"qos"},
		),

		// QoS performance metrics
		qosAverageQueueTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_average_queue_time_seconds",
				Help: "Average queue time for QoS in seconds",
			},
			[]string{"qos"},
		),
		qosAverageRunTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_average_run_time_seconds",
				Help: "Average run time for QoS in seconds",
			},
			[]string{"qos"},
		),
		qosThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_throughput_jobs_per_hour",
				Help: "Job throughput for QoS in jobs per hour",
			},
			[]string{"qos"},
		),
		qosSLACompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_sla_compliance_ratio",
				Help: "SLA compliance ratio for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosPerformanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_performance_score",
				Help: "Performance score for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosUserSatisfaction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_user_satisfaction_score",
				Help: "User satisfaction score for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosCostEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_cost_effectiveness_ratio",
				Help: "Cost effectiveness ratio for QoS",
			},
			[]string{"qos"},
		),

		// QoS hierarchy metrics
		qosHierarchyDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_hierarchy_depth",
				Help: "Depth of QoS hierarchy",
			},
			[]string{},
		),
		qosTotalCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_total_count",
				Help: "Total number of QoS policies",
			},
			[]string{},
		),
		qosInheritanceChainLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_inheritance_chain_length",
				Help: "Length of QoS inheritance chain",
			},
			[]string{"qos"},
		),
		qosConflicts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_conflicts_total",
				Help: "Number of QoS conflicts",
			},
			[]string{"conflict_type", "severity"},
		),

		// QoS effectiveness metrics
		qosOverallEffectiveness: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_overall_effectiveness",
				Help: "Overall effectiveness of QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosResourceOptimization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_resource_optimization_score",
				Help: "Resource optimization score for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosFairnessScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_fairness_score",
				Help: "Fairness score for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosSystemThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_system_throughput_impact",
				Help: "System throughput impact of QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosPolicyCompliance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_policy_compliance_score",
				Help: "Policy compliance score for QoS (0-1)",
			},
			[]string{"qos"},
		),
		qosCostBenefitRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_cost_benefit_ratio",
				Help: "Cost benefit ratio for QoS",
			},
			[]string{"qos"},
		),

		// System QoS metrics
		systemQoSUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_qos_utilization_ratio",
				Help: "System-wide QoS utilization ratio",
			},
			[]string{"qos"},
		),
		systemQoSLoad: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_qos_load",
				Help: "System-wide QoS load factor",
			},
			[]string{},
		),
		systemQoSEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_qos_efficiency",
				Help: "System-wide QoS efficiency (0-1)",
			},
			[]string{},
		),
		systemQoSViolationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_qos_violation_rate",
				Help: "System-wide QoS violation rate",
			},
			[]string{},
		),
		systemQoSComplianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_system_qos_compliance_score",
				Help: "System-wide QoS compliance score (0-1)",
			},
			[]string{},
		),

		// Collection metrics
		collectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_qos_limits_collection_duration_seconds",
				Help:    "Time spent collecting QoS limits metrics",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		collectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_qos_limits_collection_errors_total",
				Help: "Total number of errors during QoS limits collection",
			},
			[]string{"operation", "error_type"},
		),
		lastCollectionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_qos_limits_last_collection_timestamp",
				Help: "Timestamp of last successful collection",
			},
			[]string{"metric_type"},
		),
	}
}

// Describe sends the super-set of all possible descriptors
func (c *QoSLimitsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.qosLimitCPUs.Describe(ch)
	c.qosLimitMemory.Describe(ch)
	c.qosLimitNodes.Describe(ch)
	c.qosLimitJobs.Describe(ch)
	c.qosLimitWalltime.Describe(ch)
	c.qosMaxCPUsPerUser.Describe(ch)
	c.qosMaxMemoryPerUser.Describe(ch)
	c.qosMaxJobsPerUser.Describe(ch)
	c.qosMinCPUs.Describe(ch)
	c.qosMinMemory.Describe(ch)
	c.qosMinNodes.Describe(ch)
	c.qosPriority.Describe(ch)
	c.qosUsageFactor.Describe(ch)
	c.qosUsageThreshold.Describe(ch)
	c.qosGraceTime.Describe(ch)
	c.qosUserCount.Describe(ch)
	c.qosAccountCount.Describe(ch)
	c.qosResourceUsage.Describe(ch)
	c.qosResourceUtilization.Describe(ch)
	c.qosJobsRunning.Describe(ch)
	c.qosJobsPending.Describe(ch)
	c.qosJobsCompleted.Describe(ch)
	c.qosJobsFailed.Describe(ch)
	c.qosUsersActive.Describe(ch)
	c.qosWalltimeConsumed.Describe(ch)
	c.qosEfficiencyScore.Describe(ch)
	c.qosViolations.Describe(ch)
	c.qosViolationSeverity.Describe(ch)
	c.qosActiveViolations.Describe(ch)
	c.qosViolationDuration.Describe(ch)
	c.qosViolationResolutionTime.Describe(ch)
	c.qosAutoResolutionRate.Describe(ch)
	c.qosRecurrenceRate.Describe(ch)
	c.qosEnforcementEnabled.Describe(ch)
	c.qosEnforcementActions.Describe(ch)
	c.qosPreemptions.Describe(ch)
	c.qosJobsTerminated.Describe(ch)
	c.qosJobsSuspended.Describe(ch)
	c.qosUsersWarned.Describe(ch)
	c.qosEnforcementEffectiveness.Describe(ch)
	c.qosAverageQueueTime.Describe(ch)
	c.qosAverageRunTime.Describe(ch)
	c.qosThroughput.Describe(ch)
	c.qosSLACompliance.Describe(ch)
	c.qosPerformanceScore.Describe(ch)
	c.qosUserSatisfaction.Describe(ch)
	c.qosCostEffectiveness.Describe(ch)
	c.qosHierarchyDepth.Describe(ch)
	c.qosTotalCount.Describe(ch)
	c.qosInheritanceChainLength.Describe(ch)
	c.qosConflicts.Describe(ch)
	c.qosOverallEffectiveness.Describe(ch)
	c.qosResourceOptimization.Describe(ch)
	c.qosFairnessScore.Describe(ch)
	c.qosSystemThroughput.Describe(ch)
	c.qosPolicyCompliance.Describe(ch)
	c.qosCostBenefitRatio.Describe(ch)
	c.systemQoSUtilization.Describe(ch)
	c.systemQoSLoad.Describe(ch)
	c.systemQoSEfficiency.Describe(ch)
	c.systemQoSViolationRate.Describe(ch)
	c.systemQoSComplianceScore.Describe(ch)
	c.collectionDuration.Describe(ch)
	c.collectionErrors.Describe(ch)
	c.lastCollectionTime.Describe(ch)
}

// Collect fetches the stats and delivers them as Prometheus metrics
func (c *QoSLimitsCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Reset metrics
	c.qosLimitCPUs.Reset()
	c.qosLimitMemory.Reset()
	c.qosLimitNodes.Reset()
	c.qosLimitJobs.Reset()
	c.qosLimitWalltime.Reset()
	c.qosMaxCPUsPerUser.Reset()
	c.qosMaxMemoryPerUser.Reset()
	c.qosMaxJobsPerUser.Reset()
	c.qosMinCPUs.Reset()
	c.qosMinMemory.Reset()
	c.qosMinNodes.Reset()
	c.qosPriority.Reset()
	c.qosUsageFactor.Reset()
	c.qosUsageThreshold.Reset()
	c.qosGraceTime.Reset()
	c.qosUserCount.Reset()
	c.qosAccountCount.Reset()
	c.qosResourceUsage.Reset()
	c.qosResourceUtilization.Reset()
	c.qosJobsRunning.Reset()
	c.qosJobsPending.Reset()
	c.qosUsersActive.Reset()
	c.qosEfficiencyScore.Reset()
	c.qosViolationSeverity.Reset()
	c.qosActiveViolations.Reset()
	c.qosAutoResolutionRate.Reset()
	c.qosRecurrenceRate.Reset()
	c.qosEnforcementEnabled.Reset()
	c.qosEnforcementEffectiveness.Reset()
	c.qosAverageQueueTime.Reset()
	c.qosAverageRunTime.Reset()
	c.qosThroughput.Reset()
	c.qosSLACompliance.Reset()
	c.qosPerformanceScore.Reset()
	c.qosUserSatisfaction.Reset()
	c.qosCostEffectiveness.Reset()
	c.qosHierarchyDepth.Reset()
	c.qosTotalCount.Reset()
	c.qosInheritanceChainLength.Reset()
	c.qosConflicts.Reset()
	c.qosOverallEffectiveness.Reset()
	c.qosResourceOptimization.Reset()
	c.qosFairnessScore.Reset()
	c.qosSystemThroughput.Reset()
	c.qosPolicyCompliance.Reset()
	c.qosCostBenefitRatio.Reset()
	c.systemQoSUtilization.Reset()
	c.systemQoSLoad.Reset()
	c.systemQoSEfficiency.Reset()
	c.systemQoSViolationRate.Reset()
	c.systemQoSComplianceScore.Reset()

	// Collect various metrics
	c.collectQoSLimitsMetrics()
	c.collectQoSUsageMetrics()
	c.collectQoSViolationMetrics()
	c.collectQoSEnforcementMetrics()
	c.collectQoSPerformanceMetrics()
	c.collectQoSHierarchyMetrics()
	c.collectQoSEffectivenessMetrics()
	c.collectSystemQoSMetrics()

	// Collect all metrics
	c.qosLimitCPUs.Collect(ch)
	c.qosLimitMemory.Collect(ch)
	c.qosLimitNodes.Collect(ch)
	c.qosLimitJobs.Collect(ch)
	c.qosLimitWalltime.Collect(ch)
	c.qosMaxCPUsPerUser.Collect(ch)
	c.qosMaxMemoryPerUser.Collect(ch)
	c.qosMaxJobsPerUser.Collect(ch)
	c.qosMinCPUs.Collect(ch)
	c.qosMinMemory.Collect(ch)
	c.qosMinNodes.Collect(ch)
	c.qosPriority.Collect(ch)
	c.qosUsageFactor.Collect(ch)
	c.qosUsageThreshold.Collect(ch)
	c.qosGraceTime.Collect(ch)
	c.qosUserCount.Collect(ch)
	c.qosAccountCount.Collect(ch)
	c.qosResourceUsage.Collect(ch)
	c.qosResourceUtilization.Collect(ch)
	c.qosJobsRunning.Collect(ch)
	c.qosJobsPending.Collect(ch)
	c.qosJobsCompleted.Collect(ch)
	c.qosJobsFailed.Collect(ch)
	c.qosUsersActive.Collect(ch)
	c.qosWalltimeConsumed.Collect(ch)
	c.qosEfficiencyScore.Collect(ch)
	c.qosViolations.Collect(ch)
	c.qosViolationSeverity.Collect(ch)
	c.qosActiveViolations.Collect(ch)
	c.qosViolationDuration.Collect(ch)
	c.qosViolationResolutionTime.Collect(ch)
	c.qosAutoResolutionRate.Collect(ch)
	c.qosRecurrenceRate.Collect(ch)
	c.qosEnforcementEnabled.Collect(ch)
	c.qosEnforcementActions.Collect(ch)
	c.qosPreemptions.Collect(ch)
	c.qosJobsTerminated.Collect(ch)
	c.qosJobsSuspended.Collect(ch)
	c.qosUsersWarned.Collect(ch)
	c.qosEnforcementEffectiveness.Collect(ch)
	c.qosAverageQueueTime.Collect(ch)
	c.qosAverageRunTime.Collect(ch)
	c.qosThroughput.Collect(ch)
	c.qosSLACompliance.Collect(ch)
	c.qosPerformanceScore.Collect(ch)
	c.qosUserSatisfaction.Collect(ch)
	c.qosCostEffectiveness.Collect(ch)
	c.qosHierarchyDepth.Collect(ch)
	c.qosTotalCount.Collect(ch)
	c.qosInheritanceChainLength.Collect(ch)
	c.qosConflicts.Collect(ch)
	c.qosOverallEffectiveness.Collect(ch)
	c.qosResourceOptimization.Collect(ch)
	c.qosFairnessScore.Collect(ch)
	c.qosSystemThroughput.Collect(ch)
	c.qosPolicyCompliance.Collect(ch)
	c.qosCostBenefitRatio.Collect(ch)
	c.systemQoSUtilization.Collect(ch)
	c.systemQoSLoad.Collect(ch)
	c.systemQoSEfficiency.Collect(ch)
	c.systemQoSViolationRate.Collect(ch)
	c.systemQoSComplianceScore.Collect(ch)
	c.collectionDuration.Collect(ch)
	c.collectionErrors.Collect(ch)
	c.lastCollectionTime.Collect(ch)
}

func (c *QoSLimitsCollector) collectQoSLimitsMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample QoS names for collection
	sampleQoS := []string{"normal", "high", "low", "preempt", "debug"}

	for _, qosName := range sampleQoS {
		limits, err := c.client.GetQoSLimits(ctx, qosName)
		if err != nil {
			c.collectionErrors.WithLabelValues("limits", "qos_limits_error").Inc()
			continue
		}

		if limits != nil {
			// Set limit metrics
			c.qosLimitCPUs.WithLabelValues(qosName, "group").Set(float64(limits.GrpCPUs))
			c.qosLimitCPUs.WithLabelValues(qosName, "group_running").Set(float64(limits.GrpCPUsRunning))
			c.qosLimitCPUs.WithLabelValues(qosName, "max").Set(float64(limits.MaxCPUs))

			c.qosLimitMemory.WithLabelValues(qosName, "group").Set(float64(limits.GrpMem * 1024 * 1024)) // Convert MB to bytes
			c.qosLimitMemory.WithLabelValues(qosName, "group_running").Set(float64(limits.GrpMemRunning * 1024 * 1024))
			c.qosLimitMemory.WithLabelValues(qosName, "max").Set(float64(limits.MaxMem * 1024 * 1024))

			c.qosLimitNodes.WithLabelValues(qosName, "group").Set(float64(limits.GrpNodes))
			c.qosLimitNodes.WithLabelValues(qosName, "group_running").Set(float64(limits.GrpNodesRunning))
			c.qosLimitNodes.WithLabelValues(qosName, "max").Set(float64(limits.MaxNodes))

			c.qosLimitJobs.WithLabelValues(qosName, "group").Set(float64(limits.GrpJobs))
			c.qosLimitJobs.WithLabelValues(qosName, "group_running").Set(float64(limits.GrpJobsRunning))
			c.qosLimitJobs.WithLabelValues(qosName, "max").Set(float64(limits.MaxJobs))
			c.qosLimitJobs.WithLabelValues(qosName, "group_submit").Set(float64(limits.GrpSubmitJobs))

			c.qosLimitWalltime.WithLabelValues(qosName, "group").Set(limits.GrpWall.Seconds())
			c.qosLimitWalltime.WithLabelValues(qosName, "max").Set(limits.MaxWall.Seconds())
			c.qosLimitWalltime.WithLabelValues(qosName, "max_per_job").Set(limits.MaxWallPerJob.Seconds())

			// Per-user limits
			c.qosMaxCPUsPerUser.WithLabelValues(qosName).Set(float64(limits.MaxCPUsPerUser))
			c.qosMaxMemoryPerUser.WithLabelValues(qosName).Set(float64(limits.MaxMemPerUser * 1024 * 1024))
			c.qosMaxJobsPerUser.WithLabelValues(qosName).Set(float64(limits.MaxJobsPerUser))

			// Minimum requirements
			c.qosMinCPUs.WithLabelValues(qosName).Set(float64(limits.MinCPUs))
			c.qosMinMemory.WithLabelValues(qosName).Set(float64(limits.MinMem * 1024 * 1024))
			c.qosMinNodes.WithLabelValues(qosName).Set(float64(limits.MinNodes))

			// Configuration metrics
			c.qosPriority.WithLabelValues(qosName).Set(float64(limits.Priority))
			c.qosUsageFactor.WithLabelValues(qosName).Set(limits.UsageFactor)
			c.qosUsageThreshold.WithLabelValues(qosName).Set(limits.UsageThreshold)
			c.qosGraceTime.WithLabelValues(qosName).Set(limits.GraceTime.Seconds())
			c.qosUserCount.WithLabelValues(qosName).Set(float64(limits.UserCount))
			c.qosAccountCount.WithLabelValues(qosName).Set(float64(limits.AccountCount))
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("limits").Observe(duration)
	c.lastCollectionTime.WithLabelValues("limits").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSUsageMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample QoS names for usage collection
	sampleQoS := []string{"normal", "high", "low", "preempt", "debug"}

	for _, qosName := range sampleQoS {
		usage, err := c.client.GetQoSUsage(ctx, qosName)
		if err != nil {
			c.collectionErrors.WithLabelValues("usage", "qos_usage_error").Inc()
			continue
		}

		if usage != nil {
			// Resource usage metrics
			c.qosResourceUsage.WithLabelValues(qosName, "cpu", "in_use").Set(float64(usage.CPUsInUse))
			c.qosResourceUsage.WithLabelValues(qosName, "cpu", "allocated").Set(float64(usage.CPUsAllocated))
			c.qosResourceUsage.WithLabelValues(qosName, "memory", "in_use").Set(float64(usage.MemoryInUse))
			c.qosResourceUsage.WithLabelValues(qosName, "memory", "allocated").Set(float64(usage.MemoryAllocated))
			c.qosResourceUsage.WithLabelValues(qosName, "nodes", "in_use").Set(float64(usage.NodesInUse))
			c.qosResourceUsage.WithLabelValues(qosName, "nodes", "allocated").Set(float64(usage.NodesAllocated))

			// Utilization ratios
			c.qosResourceUtilization.WithLabelValues(qosName, "cpu").Set(usage.CPUUtilization)
			c.qosResourceUtilization.WithLabelValues(qosName, "memory").Set(usage.MemoryUtilization)
			c.qosResourceUtilization.WithLabelValues(qosName, "nodes").Set(usage.NodeUtilization)
			c.qosResourceUtilization.WithLabelValues(qosName, "walltime").Set(usage.WalltimeUtilization)

			// Job metrics
			c.qosJobsRunning.WithLabelValues(qosName).Set(float64(usage.JobsRunning))
			c.qosJobsPending.WithLabelValues(qosName).Set(float64(usage.JobsPending))
			c.qosJobsCompleted.WithLabelValues(qosName).Add(float64(usage.JobsCompleted))
			c.qosJobsFailed.WithLabelValues(qosName).Add(float64(usage.JobsFailed))

			// User metrics
			c.qosUsersActive.WithLabelValues(qosName).Set(float64(usage.UsersActive))

			// Walltime metrics
			c.qosWalltimeConsumed.WithLabelValues(qosName).Add(usage.WalltimeConsumed.Seconds())

			// Performance metrics
			c.qosEfficiencyScore.WithLabelValues(qosName).Set(usage.EfficiencyScore)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("usage").Observe(duration)
	c.lastCollectionTime.WithLabelValues("usage").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSViolationMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Get violations for all QoS
	opts := &QoSViolationOptions{
		Severity:  "all",
		TimeRange: "24h",
		Status:    "all",
	}

	violations, err := c.client.GetQoSViolations(ctx, opts)
	if err != nil {
		c.collectionErrors.WithLabelValues("violations", "qos_violations_error").Inc()
		return
	}

	if violations != nil {
		// Process individual violations
		for _, violation := range violations.Violations {
			c.qosViolations.WithLabelValues(violation.QoSName, violation.ViolationType, violation.Severity).Inc()

			if violation.Status == "active" {
				c.qosActiveViolations.WithLabelValues(violation.QoSName, violation.ViolationType).Inc()
			}

			// Record violation duration
			c.qosViolationDuration.WithLabelValues(violation.QoSName, violation.ViolationType).Observe(violation.Duration.Seconds())

			// Record resolution time if resolved
			if violation.Status == "resolved" && !violation.ResolutionTime.IsZero() {
				resolutionTime := violation.ResolutionTime.Sub(violation.Timestamp).Seconds()
				c.qosViolationResolutionTime.WithLabelValues(violation.QoSName, violation.ViolationType).Observe(resolutionTime)
			}
		}

		// Set violation severity distribution
		for qosName, count := range violations.ViolationsByQoS {
			c.qosViolationSeverity.WithLabelValues(qosName, "critical").Set(float64(violations.CriticalViolations))
			c.qosViolationSeverity.WithLabelValues(qosName, "warning").Set(float64(violations.WarningViolations))
			c.qosViolationSeverity.WithLabelValues(qosName, "info").Set(float64(violations.InfoViolations))
			// Simple approximation - in practice would be per-QoS
			_ = count
		}

		// Set resolution statistics
		c.qosAutoResolutionRate.WithLabelValues("all").Set(violations.ResolutionStats.AutoResolutionRate)
		c.qosRecurrenceRate.WithLabelValues("all").Set(violations.ResolutionStats.RecurrenceRate)
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("violations").Observe(duration)
	c.lastCollectionTime.WithLabelValues("violations").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSEnforcementMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample QoS names for enforcement collection
	sampleQoS := []string{"normal", "high", "low", "preempt", "debug"}

	for _, qosName := range sampleQoS {
		enforcement, err := c.client.GetQoSEnforcement(ctx, qosName)
		if err != nil {
			c.collectionErrors.WithLabelValues("enforcement", "qos_enforcement_error").Inc()
			continue
		}

		if enforcement != nil {
			// Enforcement enabled
			enabledValue := 0.0
			if enforcement.EnforcementEnabled {
				enabledValue = 1.0
			}
			c.qosEnforcementEnabled.WithLabelValues(qosName).Set(enabledValue)

			// Enforcement actions
			stats := enforcement.EnforcementStats
			c.qosEnforcementActions.WithLabelValues(qosName, "total", "success").Add(float64(stats.ActionsSuccessful))
			c.qosEnforcementActions.WithLabelValues(qosName, "total", "failed").Add(float64(stats.ActionsFailed))

			// Preemptions
			c.qosPreemptions.WithLabelValues(qosName, "all").Add(float64(stats.PreemptionsTotal))

			// Job actions
			c.qosJobsTerminated.WithLabelValues(qosName, "limit_violation").Add(float64(stats.JobsTerminated))
			c.qosJobsSuspended.WithLabelValues(qosName, "limit_violation").Add(float64(stats.JobsSuspended))

			// User actions
			c.qosUsersWarned.WithLabelValues(qosName, "limit_warning").Add(float64(stats.UsersWarned))

			// Effectiveness
			c.qosEnforcementEffectiveness.WithLabelValues(qosName).Set(stats.EffectivenessScore)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("enforcement").Observe(duration)
	c.lastCollectionTime.WithLabelValues("enforcement").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSPerformanceMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample QoS names for performance collection
	sampleQoS := []string{"normal", "high", "low", "preempt", "debug"}

	for _, qosName := range sampleQoS {
		stats, err := c.client.GetQoSStatistics(ctx, qosName, "24h")
		if err != nil {
			c.collectionErrors.WithLabelValues("performance", "qos_statistics_error").Inc()
			continue
		}

		if stats != nil {
			// Performance timing metrics
			c.qosAverageQueueTime.WithLabelValues(qosName).Set(stats.AverageQueueTime.Seconds())
			c.qosAverageRunTime.WithLabelValues(qosName).Set(stats.AverageRunTime.Seconds())

			// Throughput and efficiency
			c.qosThroughput.WithLabelValues(qosName).Set(stats.Throughput)
			c.qosSLACompliance.WithLabelValues(qosName).Set(stats.SLACompliance)
			c.qosPerformanceScore.WithLabelValues(qosName).Set(stats.PerformanceScore)
			c.qosUserSatisfaction.WithLabelValues(qosName).Set(stats.UserSatisfaction)
			c.qosCostEffectiveness.WithLabelValues(qosName).Set(stats.CostEffectiveness)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("performance").Observe(duration)
	c.lastCollectionTime.WithLabelValues("performance").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSHierarchyMetrics() {
	ctx := context.Background()
	start := time.Now()

	hierarchy, err := c.client.GetQoSHierarchy(ctx)
	if err != nil {
		c.collectionErrors.WithLabelValues("hierarchy", "qos_hierarchy_error").Inc()
		return
	}

	if hierarchy != nil {
		c.qosHierarchyDepth.WithLabelValues().Set(float64(hierarchy.MaxDepth))
		c.qosTotalCount.WithLabelValues().Set(float64(hierarchy.TotalQoS))

		// Inheritance chain lengths
		for qosName, chain := range hierarchy.InheritanceChain {
			c.qosInheritanceChainLength.WithLabelValues(qosName).Set(float64(len(chain)))
		}
	}

	// Get conflicts
	conflicts, err := c.client.GetQoSConflicts(ctx)
	if err == nil && conflicts != nil {
		c.qosConflicts.WithLabelValues("total", "all").Set(float64(conflicts.TotalConflicts))
		c.qosConflicts.WithLabelValues("critical", "high").Set(float64(conflicts.CriticalConflicts))

		for conflictType, count := range conflicts.ConflictsByType {
			c.qosConflicts.WithLabelValues(conflictType, "medium").Set(float64(count))
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("hierarchy").Observe(duration)
	c.lastCollectionTime.WithLabelValues("hierarchy").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectQoSEffectivenessMetrics() {
	ctx := context.Background()
	start := time.Now()

	// Sample QoS names for effectiveness analysis
	sampleQoS := []string{"normal", "high", "low", "preempt", "debug"}

	for _, qosName := range sampleQoS {
		effectiveness, err := c.client.GetQoSEffectiveness(ctx, qosName)
		if err != nil {
			c.collectionErrors.WithLabelValues("effectiveness", "qos_effectiveness_error").Inc()
			continue
		}

		if effectiveness != nil {
			c.qosOverallEffectiveness.WithLabelValues(qosName).Set(effectiveness.OverallEffectiveness)
			c.qosResourceOptimization.WithLabelValues(qosName).Set(effectiveness.ResourceOptimization)
			c.qosFairnessScore.WithLabelValues(qosName).Set(effectiveness.FairnessScore)
			c.qosSystemThroughput.WithLabelValues(qosName).Set(effectiveness.SystemThroughput)
			c.qosPolicyCompliance.WithLabelValues(qosName).Set(effectiveness.PolicyCompliance)
			c.qosCostBenefitRatio.WithLabelValues(qosName).Set(effectiveness.CostBenefitRatio)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("effectiveness").Observe(duration)
	c.lastCollectionTime.WithLabelValues("effectiveness").Set(float64(time.Now().Unix()))
}

func (c *QoSLimitsCollector) collectSystemQoSMetrics() {
	ctx := context.Background()
	start := time.Now()

	overview, err := c.client.GetSystemQoSOverview(ctx)
	if err != nil {
		c.collectionErrors.WithLabelValues("system", "system_qos_error").Inc()
		return
	}

	if overview != nil {
		// System-wide metrics
		c.systemQoSLoad.WithLabelValues().Set(overview.SystemLoad)
		c.systemQoSEfficiency.WithLabelValues().Set(overview.OverallEfficiency)
		c.systemQoSViolationRate.WithLabelValues().Set(overview.ViolationRate)
		c.systemQoSComplianceScore.WithLabelValues().Set(overview.ComplianceScore)

		// Per-QoS utilization
		for qosName, utilization := range overview.QoSUtilization {
			c.systemQoSUtilization.WithLabelValues(qosName).Set(utilization)
		}
	}

	duration := time.Since(start).Seconds()
	c.collectionDuration.WithLabelValues("system").Observe(duration)
	c.lastCollectionTime.WithLabelValues("system").Set(float64(time.Now().Unix()))
}
