// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"log/slog"
	// Commented out as only used in commented-out task simulation functions
	// "math"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// TaskUtilizationMonitor provides job step task-level utilization monitoring
type TaskUtilizationMonitor struct {
	slurmClient slurm.SlurmClient
	logger      *slog.Logger
	config      *TaskMonitorConfig
	metrics     *TaskUtilizationMetrics

	// Task-level data tracking
	taskData       map[string]*JobStepTaskData
	stepData       map[string]*JobStepData
	lastCollection time.Time
	mu             sync.RWMutex

	// Task performance analysis
	performanceAnalyzer *TaskPerformanceAnalyzer

	// Load balancing analysis
	loadBalancer *TaskLoadBalancer

	// Task efficiency tracking
	efficiencyTracker *TaskEfficiencyTracker
}

// TaskMonitorConfig configures the task utilization monitoring collector
type TaskMonitorConfig struct {
	MonitoringInterval        time.Duration
	MaxJobsPerCollection      int
	MaxStepsPerJob            int
	MaxTasksPerStep           int
	EnableTaskLoadBalancing   bool
	EnablePerformanceAnalysis bool
	EnableEfficiencyTracking  bool

	// Task detection parameters
	TaskDetectionMethod  string // "process_monitoring", "resource_tracking", "estimated"
	MinTaskDuration      time.Duration
	TaskSamplingInterval time.Duration

	// Load balancing parameters
	LoadImbalanceThreshold float64 // Threshold for detecting load imbalance
	TaskMigrationEnabled   bool
	LoadBalancingStrategy  string // "round_robin", "least_loaded", "performance_based"

	// Performance analysis parameters
	PerformanceVarianceThreshold float64
	TaskBottleneckDetection      bool
	EnableTaskProfiling          bool

	// Data retention
	TaskDataRetention time.Duration
	StepDataRetention time.Duration

	// Processing optimization
	BatchSize                int
	EnableParallelProcessing bool
	MaxConcurrentAnalyses    int
}

// JobStepTaskData contains task-level utilization data for a job step
type JobStepTaskData struct {
	JobID     string
	StepID    string
	TaskID    string
	NodeID    string
	Timestamp time.Time

	// Task resource utilization
	CPUUtilization     float64 // CPU utilization for this task
	MemoryUtilization  float64 // Memory utilization for this task
	IOUtilization      float64 // I/O utilization for this task
	NetworkUtilization float64 // Network utilization for this task

	// Task allocation and usage
	AllocatedCPUs       int     // CPUs allocated to this task
	AllocatedMemoryMB   int     // Memory allocated to this task in MB
	ActualCPUUsage      float64 // Actual CPU cores used
	ActualMemoryUsageMB int64   // Actual memory used in MB

	// Task performance metrics
	TaskThroughput float64 // Work units completed per second
	TaskLatency    float64 // Average task response time
	TaskEfficiency float64 // Overall task efficiency score
	TaskWasteScore float64 // Resource waste score for this task

	// Task timing information
	TaskStartTime       time.Time
	TaskDuration        time.Duration
	EstimatedCompletion *time.Time

	// Task state and status
	TaskState     string  // "running", "completed", "failed", "idle"
	TaskProgress  float64 // Progress percentage (0-1)
	TaskErrorRate float64 // Error rate for this task

	// Load balancing metrics
	LoadScore               float64 // Current load score
	LoadImbalance           float64 // Load imbalance ratio
	BalancingRecommendation string  // Load balancing recommendation

	// Communication patterns
	MessagesSent      int64   // Number of messages sent
	MessagesReceived  int64   // Number of messages received
	DataTransferredMB float64 // Data transferred in MB

	// Task dependencies
	DependentTasks []string // List of dependent task IDs
	BlockingTasks  []string // List of tasks blocking this task
	CriticalPath   bool     // Whether this task is on the critical path
}

// JobStepData contains aggregated data for a job step
type JobStepData struct {
	JobID     string
	StepID    string
	Timestamp time.Time

	// Step-level aggregations
	TotalTasks     int
	ActiveTasks    int
	CompletedTasks int
	FailedTasks    int

	// Resource utilization aggregations
	AvgCPUUtilization    float64
	MinCPUUtilization    float64
	MaxCPUUtilization    float64
	StdDevCPUUtilization float64

	AvgMemoryUtilization    float64
	MinMemoryUtilization    float64
	MaxMemoryUtilization    float64
	StdDevMemoryUtilization float64

	// Performance aggregations
	AvgTaskThroughput float64
	AvgTaskLatency    float64
	AvgTaskEfficiency float64
	TotalTaskWaste    float64

	// Load balancing metrics
	LoadImbalanceScore    float64
	LoadVarianceScore     float64
	TaskDistributionScore float64

	// Step performance analysis
	BottleneckTasks           []string
	CriticalPathTasks         []string
	PerformanceIssues         []TaskPerformanceIssue
	OptimizationOpportunities []TaskOptimizationOpportunity

	// Communication analysis
	TotalMessagesExchanged  int64
	TotalDataTransferredMB  float64
	CommunicationEfficiency float64

	// Progress and timing
	StepProgress            float64
	EstimatedStepCompletion *time.Time
	StepBottlenecks         []TaskBottleneck
}

// TaskPerformanceIssue represents a detected performance issue at task level
type TaskPerformanceIssue struct {
	IssueID             string
	TaskID              string
	IssueType           string // "cpu_underutilization", "memory_pressure", "load_imbalance", etc.
	Severity            string // "low", "medium", "high", "critical"
	Description         string
	Impact              string
	Detection           time.Time
	Recommendations     []string
	ExpectedImprovement float64
}

// TaskOptimizationOpportunity represents an optimization opportunity
type TaskOptimizationOpportunity struct {
	OpportunityID        string
	OpportunityType      string // "load_rebalancing", "resource_reallocation", "task_migration"
	AffectedTasks        []string
	ExpectedBenefit      float64
	ImplementationEffort string
	Priority             string
	Description          string
	ActionSteps          []string
}

// TaskBottleneck represents a task-level bottleneck
type TaskBottleneck struct {
	TaskID             string
	BottleneckType     string // "cpu", "memory", "io", "network", "synchronization"
	Severity           float64
	Duration           time.Duration
	Impact             string
	AffectedTasks      []string
	RecommendedActions []string
}

// TaskPerformanceAnalyzer analyzes task-level performance
type TaskPerformanceAnalyzer struct {
	config          *TaskMonitorConfig
	logger          *slog.Logger
	performanceData map[string]*TaskPerformanceData
	analysisResults map[string]*TaskAnalysisResult
	issueHistory    map[string][]TaskPerformanceIssue
}

// TaskPerformanceData contains performance data for analysis
type TaskPerformanceData struct {
	TaskID               string
	PerformanceHistory   []TaskPerformancePoint
	ResourceHistory      []TaskResourcePoint
	CommunicationHistory []TaskCommunicationPoint
	LoadHistory          []TaskLoadPoint
}

// TaskPerformancePoint represents a single performance measurement
type TaskPerformancePoint struct {
	Timestamp  time.Time
	Throughput float64
	Latency    float64
	Efficiency float64
	ErrorRate  float64
}

// TaskResourcePoint represents a single resource measurement
type TaskResourcePoint struct {
	Timestamp          time.Time
	CPUUtilization     float64
	MemoryUtilization  float64
	IOUtilization      float64
	NetworkUtilization float64
}

// TaskCommunicationPoint represents a single communication measurement
type TaskCommunicationPoint struct {
	Timestamp        time.Time
	MessagesSent     int64
	MessagesReceived int64
	DataTransferred  float64
	Bandwidth        float64
}

// TaskLoadPoint represents a single load measurement
type TaskLoadPoint struct {
	Timestamp  time.Time
	LoadScore  float64
	Imbalance  float64
	QueueDepth int
	WaitTime   time.Duration
}

// TaskAnalysisResult contains analysis results for a task
type TaskAnalysisResult struct {
	TaskID              string
	AnalysisTimestamp   time.Time
	PerformanceGrade    string
	EfficiencyScore     float64
	LoadBalanceScore    float64
	CommunicationScore  float64
	OverallScore        float64
	DetectedIssues      []TaskPerformanceIssue
	Recommendations     []string
	PredictedCompletion *time.Time
}

// TaskLoadBalancer handles task load balancing analysis
type TaskLoadBalancer struct {
	config               *TaskMonitorConfig
	logger               *slog.Logger
	loadData             map[string]*TaskLoadData
	balancingHistory     map[string][]LoadBalancingEvent
	migrationSuggestions []TaskMigrationSuggestion
}

// TaskLoadData contains load balancing data
type TaskLoadData struct {
	TaskID         string
	CurrentLoad    float64
	AverageLoad    float64
	LoadTrend      string
	LoadVariance   float64
	Capacity       float64
	Utilization    float64
	QueueLength    int
	ProcessingRate float64
}

// LoadBalancingEvent represents a load balancing event
type LoadBalancingEvent struct {
	EventID       string
	Timestamp     time.Time
	EventType     string // "imbalance_detected", "migration_suggested", "rebalancing_performed"
	AffectedTasks []string
	Action        string
	Result        string
	Impact        float64
}

// TaskMigrationSuggestion represents a task migration suggestion
type TaskMigrationSuggestion struct {
	SuggestionID    string
	TaskID          string
	SourceNode      string
	TargetNode      string
	Reason          string
	ExpectedBenefit float64
	MigrationCost   float64
	Priority        string
	Timestamp       time.Time
}

// TaskEfficiencyTracker tracks task efficiency over time
type TaskEfficiencyTracker struct {
	config         *TaskMonitorConfig
	logger         *slog.Logger
	efficiencyData map[string]*TaskEfficiencyData
	benchmarkData  map[string]float64
	// TODO: EfficiencyTrend type is not defined - commented out for now
	// trends            map[string]*EfficiencyTrend
}

// TaskEfficiencyData contains efficiency tracking data
type TaskEfficiencyData struct {
	TaskID              string
	EfficiencyHistory   []EfficiencyPoint
	BenchmarkComparison float64
	EfficiencyTrend     string
	ImprovementRate     float64
	TargetEfficiency    float64
}

// EfficiencyPoint represents a single efficiency measurement
type EfficiencyPoint struct {
	Timestamp  time.Time
	Efficiency float64
	CPUEff     float64
	MemoryEff  float64
	IOEff      float64
	NetworkEff float64
}

// TaskUtilizationMetrics holds Prometheus metrics for task utilization monitoring
type TaskUtilizationMetrics struct {
	// Task-level utilization metrics
	TaskCPUUtilization     *prometheus.GaugeVec
	TaskMemoryUtilization  *prometheus.GaugeVec
	TaskIOUtilization      *prometheus.GaugeVec
	TaskNetworkUtilization *prometheus.GaugeVec

	// Task performance metrics
	TaskThroughput *prometheus.GaugeVec
	TaskLatency    *prometheus.GaugeVec
	TaskEfficiency *prometheus.GaugeVec
	TaskWasteScore *prometheus.GaugeVec
	TaskErrorRate  *prometheus.GaugeVec

	// Task load balancing metrics
	TaskLoadScore            *prometheus.GaugeVec
	TaskLoadImbalance        *prometheus.GaugeVec
	LoadVarianceScore        *prometheus.GaugeVec
	TaskMigrationSuggestions *prometheus.GaugeVec

	// Step-level aggregation metrics
	StepAvgCPUUtilization    *prometheus.GaugeVec
	StepMaxCPUUtilization    *prometheus.GaugeVec
	StepCPUVariance          *prometheus.GaugeVec
	StepAvgMemoryUtilization *prometheus.GaugeVec
	StepMemoryVariance       *prometheus.GaugeVec

	// Step performance metrics
	StepAvgThroughput      *prometheus.GaugeVec
	StepAvgLatency         *prometheus.GaugeVec
	StepAvgEfficiency      *prometheus.GaugeVec
	StepLoadImbalanceScore *prometheus.GaugeVec

	// Task communication metrics
	TaskMessagesExchanged   *prometheus.CounterVec
	TaskDataTransferred     *prometheus.CounterVec
	CommunicationEfficiency *prometheus.GaugeVec

	// Task state metrics
	TasksRunning   *prometheus.GaugeVec
	TasksCompleted *prometheus.CounterVec
	TasksFailed    *prometheus.CounterVec
	TaskProgress   *prometheus.GaugeVec

	// Performance issue metrics
	TaskPerformanceIssues     *prometheus.GaugeVec
	TaskBottlenecks           *prometheus.GaugeVec
	OptimizationOpportunities *prometheus.GaugeVec

	// Critical path metrics
	CriticalPathTasks       *prometheus.GaugeVec
	StepCriticalPathLength  *prometheus.GaugeVec
	CriticalPathBottlenecks *prometheus.GaugeVec

	// Collection metrics
	TaskDataCollectionDuration *prometheus.HistogramVec
	TaskDataCollectionErrors   *prometheus.CounterVec
	MonitoredTasksCount        *prometheus.GaugeVec
}

// NewTaskUtilizationMonitor creates a new task utilization monitoring collector
func NewTaskUtilizationMonitor(client slurm.SlurmClient, logger *slog.Logger, config *TaskMonitorConfig) (*TaskUtilizationMonitor, error) {
	if config == nil {
		config = &TaskMonitorConfig{
			MonitoringInterval:           30 * time.Second,
			MaxJobsPerCollection:         50,
			MaxStepsPerJob:               10,
			MaxTasksPerStep:              100,
			EnableTaskLoadBalancing:      true,
			EnablePerformanceAnalysis:    true,
			EnableEfficiencyTracking:     true,
			TaskDetectionMethod:          "resource_tracking",
			MinTaskDuration:              1 * time.Second,
			TaskSamplingInterval:         5 * time.Second,
			LoadImbalanceThreshold:       0.2,
			TaskMigrationEnabled:         false,
			LoadBalancingStrategy:        "performance_based",
			PerformanceVarianceThreshold: 0.3,
			TaskBottleneckDetection:      true,
			EnableTaskProfiling:          true,
			TaskDataRetention:            2 * time.Hour,
			StepDataRetention:            24 * time.Hour,
			BatchSize:                    20,
			EnableParallelProcessing:     true,
			MaxConcurrentAnalyses:        3,
		}
	}

	performanceAnalyzer := &TaskPerformanceAnalyzer{
		config:          config,
		logger:          logger,
		performanceData: make(map[string]*TaskPerformanceData),
		analysisResults: make(map[string]*TaskAnalysisResult),
		issueHistory:    make(map[string][]TaskPerformanceIssue),
	}

	loadBalancer := &TaskLoadBalancer{
		config:               config,
		logger:               logger,
		loadData:             make(map[string]*TaskLoadData),
		balancingHistory:     make(map[string][]LoadBalancingEvent),
		migrationSuggestions: []TaskMigrationSuggestion{},
	}

	efficiencyTracker := &TaskEfficiencyTracker{
		config:         config,
		logger:         logger,
		efficiencyData: make(map[string]*TaskEfficiencyData),
		benchmarkData:  getTaskBenchmarks(),
		// trends:         make(map[string]*EfficiencyTrend), // Commented out - EfficiencyTrend type not defined
	}

	return &TaskUtilizationMonitor{
		slurmClient:         client,
		logger:              logger,
		config:              config,
		metrics:             newTaskUtilizationMetrics(),
		taskData:            make(map[string]*JobStepTaskData),
		stepData:            make(map[string]*JobStepData),
		performanceAnalyzer: performanceAnalyzer,
		loadBalancer:        loadBalancer,
		efficiencyTracker:   efficiencyTracker,
	}, nil
}

// newTaskUtilizationMetrics creates Prometheus metrics for task utilization monitoring
func newTaskUtilizationMetrics() *TaskUtilizationMetrics {
	return &TaskUtilizationMetrics{
		TaskCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_cpu_utilization",
				Help: "CPU utilization for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskMemoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_memory_utilization",
				Help: "Memory utilization for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskIOUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_io_utilization",
				Help: "I/O utilization for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskNetworkUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_network_utilization",
				Help: "Network utilization for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_throughput",
				Help: "Throughput for individual tasks (work units per second)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_latency_seconds",
				Help: "Latency for individual tasks in seconds",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_efficiency",
				Help: "Overall efficiency score for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskWasteScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_waste_score",
				Help: "Resource waste score for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskErrorRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_error_rate",
				Help: "Error rate for individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskLoadScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_load_score",
				Help: "Load score for individual tasks",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		TaskLoadImbalance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_load_imbalance",
				Help: "Load imbalance ratio for individual tasks",
			},
			[]string{"job_id", "step_id", "task_id", "node_id", "user", "account", "partition"},
		),
		LoadVarianceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_load_variance_score",
				Help: "Load variance score for job steps",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TaskMigrationSuggestions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_migration_suggestions",
				Help: "Number of task migration suggestions",
			},
			[]string{"job_id", "step_id", "source_node", "target_node", "reason"},
		),
		StepAvgCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_avg_cpu_utilization",
				Help: "Average CPU utilization across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepMaxCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_max_cpu_utilization",
				Help: "Maximum CPU utilization across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepCPUVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_cpu_variance",
				Help: "CPU utilization variance across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepAvgMemoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_avg_memory_utilization",
				Help: "Average memory utilization across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepMemoryVariance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_memory_variance",
				Help: "Memory utilization variance across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepAvgThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_avg_throughput",
				Help: "Average throughput across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepAvgLatency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_avg_latency_seconds",
				Help: "Average latency across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepAvgEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_avg_efficiency",
				Help: "Average efficiency across all tasks in a step",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepLoadImbalanceScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_load_imbalance_score",
				Help: "Load imbalance score for job steps",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TaskMessagesExchanged: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_task_messages_exchanged_total",
				Help: "Total number of messages exchanged by tasks",
			},
			[]string{"job_id", "step_id", "task_id", "direction"},
		),
		TaskDataTransferred: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_task_data_transferred_bytes_total",
				Help: "Total data transferred by tasks in bytes",
			},
			[]string{"job_id", "step_id", "task_id", "direction"},
		),
		CommunicationEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_communication_efficiency",
				Help: "Communication efficiency for job steps",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TasksRunning: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_tasks_running",
				Help: "Number of currently running tasks",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TasksCompleted: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_tasks_completed_total",
				Help: "Total number of completed tasks",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TasksFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_tasks_failed_total",
				Help: "Total number of failed tasks",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		TaskProgress: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_progress",
				Help: "Progress of individual tasks (0-1)",
			},
			[]string{"job_id", "step_id", "task_id", "user", "account", "partition"},
		),
		TaskPerformanceIssues: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_performance_issues",
				Help: "Number of performance issues detected for tasks",
			},
			[]string{"job_id", "step_id", "task_id", "issue_type", "severity"},
		),
		TaskBottlenecks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_bottlenecks",
				Help: "Number of bottlenecks detected for tasks",
			},
			[]string{"job_id", "step_id", "task_id", "bottleneck_type"},
		),
		OptimizationOpportunities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_task_optimization_opportunities",
				Help: "Number of optimization opportunities for tasks",
			},
			[]string{"job_id", "step_id", "opportunity_type"},
		),
		CriticalPathTasks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_critical_path_tasks",
				Help: "Number of tasks on the critical path",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		StepCriticalPathLength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_step_critical_path_length_seconds",
				Help: "Length of the critical path for job steps in seconds",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		CriticalPathBottlenecks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_critical_path_bottlenecks",
				Help: "Number of bottlenecks on the critical path",
			},
			[]string{"job_id", "step_id", "bottleneck_type"},
		),
		TaskDataCollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_task_data_collection_duration_seconds",
				Help:    "Duration of task data collection operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		TaskDataCollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_task_data_collection_errors_total",
				Help: "Total number of task data collection errors",
			},
			[]string{"operation", "error_type"},
		),
		MonitoredTasksCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_monitored_tasks_count",
				Help: "Number of tasks currently being monitored",
			},
			[]string{"status"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (t *TaskUtilizationMonitor) Describe(ch chan<- *prometheus.Desc) {
	t.metrics.TaskCPUUtilization.Describe(ch)
	t.metrics.TaskMemoryUtilization.Describe(ch)
	t.metrics.TaskIOUtilization.Describe(ch)
	t.metrics.TaskNetworkUtilization.Describe(ch)
	t.metrics.TaskThroughput.Describe(ch)
	t.metrics.TaskLatency.Describe(ch)
	t.metrics.TaskEfficiency.Describe(ch)
	t.metrics.TaskWasteScore.Describe(ch)
	t.metrics.TaskErrorRate.Describe(ch)
	t.metrics.TaskLoadScore.Describe(ch)
	t.metrics.TaskLoadImbalance.Describe(ch)
	t.metrics.LoadVarianceScore.Describe(ch)
	t.metrics.TaskMigrationSuggestions.Describe(ch)
	t.metrics.StepAvgCPUUtilization.Describe(ch)
	t.metrics.StepMaxCPUUtilization.Describe(ch)
	t.metrics.StepCPUVariance.Describe(ch)
	t.metrics.StepAvgMemoryUtilization.Describe(ch)
	t.metrics.StepMemoryVariance.Describe(ch)
	t.metrics.StepAvgThroughput.Describe(ch)
	t.metrics.StepAvgLatency.Describe(ch)
	t.metrics.StepAvgEfficiency.Describe(ch)
	t.metrics.StepLoadImbalanceScore.Describe(ch)
	t.metrics.TaskMessagesExchanged.Describe(ch)
	t.metrics.TaskDataTransferred.Describe(ch)
	t.metrics.CommunicationEfficiency.Describe(ch)
	t.metrics.TasksRunning.Describe(ch)
	t.metrics.TasksCompleted.Describe(ch)
	t.metrics.TasksFailed.Describe(ch)
	t.metrics.TaskProgress.Describe(ch)
	t.metrics.TaskPerformanceIssues.Describe(ch)
	t.metrics.TaskBottlenecks.Describe(ch)
	t.metrics.OptimizationOpportunities.Describe(ch)
	t.metrics.CriticalPathTasks.Describe(ch)
	t.metrics.StepCriticalPathLength.Describe(ch)
	t.metrics.CriticalPathBottlenecks.Describe(ch)
	t.metrics.TaskDataCollectionDuration.Describe(ch)
	t.metrics.TaskDataCollectionErrors.Describe(ch)
	t.metrics.MonitoredTasksCount.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (t *TaskUtilizationMonitor) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if err := t.collectTaskUtilizationMetrics(ctx); err != nil {
		t.logger.Error("Failed to collect task utilization metrics", "error", err)
		t.metrics.TaskDataCollectionErrors.WithLabelValues("collect", "collection_error").Inc()
	}

	t.metrics.TaskCPUUtilization.Collect(ch)
	t.metrics.TaskMemoryUtilization.Collect(ch)
	t.metrics.TaskIOUtilization.Collect(ch)
	t.metrics.TaskNetworkUtilization.Collect(ch)
	t.metrics.TaskThroughput.Collect(ch)
	t.metrics.TaskLatency.Collect(ch)
	t.metrics.TaskEfficiency.Collect(ch)
	t.metrics.TaskWasteScore.Collect(ch)
	t.metrics.TaskErrorRate.Collect(ch)
	t.metrics.TaskLoadScore.Collect(ch)
	t.metrics.TaskLoadImbalance.Collect(ch)
	t.metrics.LoadVarianceScore.Collect(ch)
	t.metrics.TaskMigrationSuggestions.Collect(ch)
	t.metrics.StepAvgCPUUtilization.Collect(ch)
	t.metrics.StepMaxCPUUtilization.Collect(ch)
	t.metrics.StepCPUVariance.Collect(ch)
	t.metrics.StepAvgMemoryUtilization.Collect(ch)
	t.metrics.StepMemoryVariance.Collect(ch)
	t.metrics.StepAvgThroughput.Collect(ch)
	t.metrics.StepAvgLatency.Collect(ch)
	t.metrics.StepAvgEfficiency.Collect(ch)
	t.metrics.StepLoadImbalanceScore.Collect(ch)
	t.metrics.TaskMessagesExchanged.Collect(ch)
	t.metrics.TaskDataTransferred.Collect(ch)
	t.metrics.CommunicationEfficiency.Collect(ch)
	t.metrics.TasksRunning.Collect(ch)
	t.metrics.TasksCompleted.Collect(ch)
	t.metrics.TasksFailed.Collect(ch)
	t.metrics.TaskProgress.Collect(ch)
	t.metrics.TaskPerformanceIssues.Collect(ch)
	t.metrics.TaskBottlenecks.Collect(ch)
	t.metrics.OptimizationOpportunities.Collect(ch)
	t.metrics.CriticalPathTasks.Collect(ch)
	t.metrics.StepCriticalPathLength.Collect(ch)
	t.metrics.CriticalPathBottlenecks.Collect(ch)
	t.metrics.TaskDataCollectionDuration.Collect(ch)
	t.metrics.TaskDataCollectionErrors.Collect(ch)
	t.metrics.MonitoredTasksCount.Collect(ch)
}

// collectTaskUtilizationMetrics collects task-level utilization data
func (t *TaskUtilizationMonitor) collectTaskUtilizationMetrics(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		t.metrics.TaskDataCollectionDuration.WithLabelValues("collect_task_metrics").Observe(time.Since(startTime).Seconds())
	}()

	// Get running jobs for task monitoring
	jobManager := t.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// TODO: ListJobsOptions structure is not compatible with current slurm-client
	// Using nil for options as a workaround
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs for task monitoring: %w", err)
	}

	totalTasks := 0

	// Process each job for task-level monitoring
	// TODO: Job type mismatch - jobs.Jobs returns []interfaces.Job but functions expect *slurm.Job
	// Skipping job processing for now
	_ = jobs // Suppress unused variable warning
	/*
		for _, job := range jobs.Jobs {
			if err := t.processJobStepTasks(ctx, job); err != nil {
				t.logger.Error("Failed to process job step tasks", "job_id", job.JobID, "error", err)
				t.metrics.TaskDataCollectionErrors.WithLabelValues("process_job", "job_error").Inc()
				continue
			}

			// Count tasks for this job
			jobTasks := t.countJobTasks(job.JobID)
			totalTasks += jobTasks
		}
	*/

	t.metrics.MonitoredTasksCount.WithLabelValues("total").Set(float64(totalTasks))

	// Perform cross-task analysis
	if t.config.EnablePerformanceAnalysis {
		if err := t.performanceAnalyzer.analyzeTaskPerformance(); err != nil {
			t.logger.Error("Failed to analyze task performance", "error", err)
			t.metrics.TaskDataCollectionErrors.WithLabelValues("analyze_performance", "analysis_error").Inc()
		}
	}

	// Perform load balancing analysis
	if t.config.EnableTaskLoadBalancing {
		if err := t.loadBalancer.analyzeLoadBalancing(); err != nil {
			t.logger.Error("Failed to analyze load balancing", "error", err)
			t.metrics.TaskDataCollectionErrors.WithLabelValues("analyze_load_balancing", "analysis_error").Inc()
		}
	}

	// Clean old data
	t.cleanOldTaskData()

	t.lastCollection = time.Now()
	return nil
}

// TODO: Following task processing methods are unused - preserved for future task-level monitoring
/*
// processJobStepTasks processes task-level data for all steps in a job
func (t *TaskUtilizationMonitor) processJobStepTasks(ctx context.Context, job *slurm.Job) error {
	// TODO: Job field names are not compatible with current slurm-client version
	// Returning nil for now
	return nil
}

// processStepTasks processes task-level data for a single step
func (t *TaskUtilizationMonitor) processStepTasks(ctx context.Context, job *slurm.Job, step *JobStepInfo) error {
	// Get tasks for this step (simplified simulation)
	tasks := t.generateStepTasks(job, step)

	var stepTaskData []*JobStepTaskData

	// TODO: job.JobID field not available in current slurm-client version
	jobID := fmt.Sprintf("job_%s", job.ID)

	for i, task := range tasks {
		if i >= t.config.MaxTasksPerStep {
			break
		}

		taskData := t.collectTaskData(job, step, task)
		stepTaskData = append(stepTaskData, taskData)

		// Store task data
		taskKey := fmt.Sprintf("%s:%s:%s", jobID, step.StepID, task.TaskID)
		t.mu.Lock()
		t.taskData[taskKey] = taskData
		t.mu.Unlock()

		// Update task metrics
		t.updateTaskMetrics(job, taskData)

		// Track performance data if enabled
		if t.config.EnablePerformanceAnalysis {
			t.performanceAnalyzer.trackTaskPerformance(taskData)
		}

		// Track load data if enabled
		if t.config.EnableTaskLoadBalancing {
			t.loadBalancer.trackTaskLoad(taskData)
		}

		// Track efficiency if enabled
		if t.config.EnableEfficiencyTracking {
			t.efficiencyTracker.trackTaskEfficiency(taskData)
		}
	}

	// Aggregate step-level data
	stepData := t.aggregateStepData(job, step, stepTaskData)
	stepKey := fmt.Sprintf("%s:%s", jobID, step.StepID)
	t.mu.Lock()
	t.stepData[stepKey] = stepData
	t.mu.Unlock()

	// Update step metrics
	t.updateStepMetrics(job, stepData)

	return nil
}
*/

// JobStepInfo represents basic step information (placeholder)
type JobStepInfo struct {
	StepID    string
	StepName  string
	NodeCount int
	TaskCount int
}

// TaskInfo represents basic task information (placeholder)
type TaskInfo struct {
	TaskID   string
	NodeID   string
	CPUCount int
	MemoryMB int
}

/*
// getJobSteps gets job steps (simplified simulation)
func (t *TaskUtilizationMonitor) getJobSteps(job *slurm.Job) []*JobStepInfo {
	// TODO: Job field types are not compatible with current slurm-client version
	// Using placeholder values for now
	stepCount := 1
	defaultCPUs := 4

	var steps []*JobStepInfo
	for i := 0; i < stepCount; i++ {
		step := &JobStepInfo{
			StepID:    fmt.Sprintf("%d", i),
			StepName:  fmt.Sprintf("step_%d", i),
			NodeCount: 1, // Default node count
			TaskCount: defaultCPUs / stepCount,
		}
		steps = append(steps, step)
	}

	return steps
}
*/

// TODO: Following task-level utility and helper methods are unused - preserved for future task monitoring implementation
/*
// generateStepTasks generates tasks for a step (simplified simulation)
func (t *TaskUtilizationMonitor) generateStepTasks(job *slurm.Job, step *JobStepInfo) []*TaskInfo {
	var tasks []*TaskInfo

	tasksPerNode := step.TaskCount / step.NodeCount
	if tasksPerNode == 0 {
		tasksPerNode = 1
	}

	cpusPerTask := job.CPUs / step.TaskCount
	if cpusPerTask == 0 {
		cpusPerTask = 1
	}

	memoryPerTask := job.Memory / step.TaskCount
	if memoryPerTask == 0 {
		memoryPerTask = 1024 // 1GB minimum
	}

	for nodeIdx := 0; nodeIdx < step.NodeCount; nodeIdx++ {
		for taskIdx := 0; taskIdx < tasksPerNode; taskIdx++ {
			taskID := fmt.Sprintf("%d_%d", nodeIdx, taskIdx)
			nodeID := fmt.Sprintf("node%d", nodeIdx)

			task := &TaskInfo{
				TaskID:   taskID,
				NodeID:   nodeID,
				CPUCount: cpusPerTask,
				MemoryMB: memoryPerTask,
			}
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// collectTaskData collects utilization data for a single task
func (t *TaskUtilizationMonitor) collectTaskData(job *slurm.Job, step *JobStepInfo, task *TaskInfo) *JobStepTaskData {
	now := time.Now()

	// Simulate task resource utilization
	cpuUtil := t.simulateTaskCPUUtilization(job, task)
	memoryUtil := t.simulateTaskMemoryUtilization(job, task)
	ioUtil := t.simulateTaskIOUtilization(job, task)
	networkUtil := t.simulateTaskNetworkUtilization(job, task)

	// Calculate actual usage
	actualCPUUsage := cpuUtil * float64(task.CPUCount)
	actualMemoryUsage := int64(memoryUtil * float64(task.MemoryMB) * 1024 * 1024) // Convert to bytes

	// Calculate performance metrics
	throughput := t.calculateTaskThroughput(cpuUtil, memoryUtil)
	latency := t.calculateTaskLatency(cpuUtil, ioUtil)
	efficiency := t.calculateTaskEfficiency(cpuUtil, memoryUtil, ioUtil, networkUtil)
	wasteScore := t.calculateTaskWasteScore(cpuUtil, memoryUtil, efficiency)
	errorRate := t.simulateTaskErrorRate(job, task)

	// Calculate load metrics
	loadScore := t.calculateTaskLoadScore(cpuUtil, memoryUtil, throughput)
	// TODO: job.JobID field not available in current slurm-client version
	jobID := "job_0"
	loadImbalance := t.calculateTaskLoadImbalance(jobID, step.StepID, loadScore)

	// Simulate communication metrics
	messagesSent := t.simulateMessagesSent(job, task)
	messagesReceived := t.simulateMessagesReceived(job, task)
	dataTransferred := t.simulateDataTransferred(job, task)

	// Calculate task timing
	var taskStartTime time.Time
	var taskDuration time.Duration
	if job.StartTime != nil {
		taskStartTime = job.StartTime.Add(time.Duration(len(task.TaskID)) * time.Second) // Stagger start times
		taskDuration = time.Since(taskStartTime)
	} else {
		taskStartTime = now.Add(-time.Hour)
		taskDuration = time.Hour
	}

	// Calculate progress and state
	progress := t.calculateTaskProgress(taskDuration, job.TimeLimit)
	state := t.determineTaskState(progress, errorRate)

	// Determine critical path
	criticalPath := t.isTaskOnCriticalPath(job, step, task)

	return &JobStepTaskData{
		JobID:         jobID,
		StepID:        step.StepID,
		TaskID:        task.TaskID,
		NodeID:        task.NodeID,
		Timestamp:     now,

		CPUUtilization:     cpuUtil,
		MemoryUtilization:  memoryUtil,
		IOUtilization:      ioUtil,
		NetworkUtilization: networkUtil,

		AllocatedCPUs:      task.CPUCount,
		AllocatedMemoryMB:  task.MemoryMB,
		ActualCPUUsage:     actualCPUUsage,
		ActualMemoryUsageMB: actualMemoryUsage,

		TaskThroughput:     throughput,
		TaskLatency:        latency,
		TaskEfficiency:     efficiency,
		TaskWasteScore:     wasteScore,

		TaskStartTime:      taskStartTime,
		TaskDuration:       taskDuration,
		EstimatedCompletion: t.estimateTaskCompletion(taskStartTime, progress, job.TimeLimit),

		TaskState:          state,
		TaskProgress:       progress,
		TaskErrorRate:      errorRate,

		LoadScore:          loadScore,
		LoadImbalance:      loadImbalance,
		BalancingRecommendation: t.generateBalancingRecommendation(loadScore, loadImbalance),

		MessagesSent:       messagesSent,
		MessagesReceived:   messagesReceived,
		DataTransferredMB:  dataTransferred,

		CriticalPath:       criticalPath,
	}
}

// simulateTaskCPUUtilization simulates CPU utilization for a task
func (t *TaskUtilizationMonitor) simulateTaskCPUUtilization(job *slurm.Job, task *TaskInfo) float64 {
	// Base utilization with task-specific variation
	baseUtil := 0.7 + (float64(len(task.TaskID)%5) * 0.05)

	// Add job-specific factors
	if job.CPUs > 16 {
		baseUtil += 0.1 // Higher utilization for larger jobs
	}

	// Add time-based variation
	if job.StartTime != nil {
		elapsed := time.Since(*job.StartTime).Hours()
		timeVariation := math.Sin(elapsed/2) * 0.1
		baseUtil += timeVariation
	}

	// Clamp between 0.1 and 1.0
	return math.Max(0.1, math.Min(1.0, baseUtil))
}

// simulateTaskMemoryUtilization simulates memory utilization for a task
func (t *TaskUtilizationMonitor) simulateTaskMemoryUtilization(job *slurm.Job, task *TaskInfo) float64 {
	// Memory utilization tends to be more stable than CPU
	baseUtil := 0.6 + (float64(len(task.TaskID)%4) * 0.08)

	// Memory-intensive jobs have higher utilization
	memoryPerCPU := float64(task.MemoryMB) / float64(task.CPUCount)
	if memoryPerCPU > 4000 { // > 4GB per CPU
		baseUtil += 0.15
	}

	// Clamp between 0.2 and 0.95
	return math.Max(0.2, math.Min(0.95, baseUtil))
}

// simulateTaskIOUtilization simulates I/O utilization for a task
func (t *TaskUtilizationMonitor) simulateTaskIOUtilization(job *slurm.Job, task *TaskInfo) float64 {
	// I/O utilization is typically lower and more variable
	baseUtil := 0.3 + (float64(len(task.TaskID)%6) * 0.06)

	// Multi-node jobs may have higher I/O
	if len(job.Nodes) > 1 {
		baseUtil += 0.2
	}

	// Add random spikes for I/O-intensive periods
	if len(task.TaskID)%7 == 0 {
		baseUtil += 0.3 // I/O spike
	}

	return math.Max(0.05, math.Min(0.8, baseUtil))
}

// simulateTaskNetworkUtilization simulates network utilization for a task
func (t *TaskUtilizationMonitor) simulateTaskNetworkUtilization(job *slurm.Job, task *TaskInfo) float64 {
	// Network utilization depends on job communication patterns
	baseUtil := 0.2 + (float64(len(task.TaskID)%3) * 0.1)

	// Multi-node jobs have higher network utilization
	nodeCount := len(job.Nodes)
	if nodeCount > 1 {
		baseUtil += 0.3 * float64(nodeCount-1) / 10.0
	}

	// High-CPU jobs may be communication-intensive
	if job.CPUs > 32 {
		baseUtil += 0.2
	}

	return math.Max(0.05, math.Min(0.7, baseUtil))
}

// calculateTaskThroughput calculates task throughput based on utilization
func (t *TaskUtilizationMonitor) calculateTaskThroughput(cpuUtil, memoryUtil float64) float64 {
	// Simple throughput model based on resource utilization
	avgUtil := (cpuUtil + memoryUtil) / 2.0
	return avgUtil * 100.0 // Normalize to 0-100 scale
}

// calculateTaskLatency calculates task latency
func (t *TaskUtilizationMonitor) calculateTaskLatency(cpuUtil, ioUtil float64) float64 {
	// Latency inversely related to CPU utilization, positively related to I/O utilization
	cpuLatency := (1.0 - cpuUtil) * 0.5
	ioLatency := ioUtil * 0.3
	return cpuLatency + ioLatency
}

// calculateTaskEfficiency calculates overall task efficiency
func (t *TaskUtilizationMonitor) calculateTaskEfficiency(cpuUtil, memoryUtil, ioUtil, networkUtil float64) float64 {
	// Weighted efficiency calculation
	weights := []float64{0.4, 0.3, 0.2, 0.1} // CPU, Memory, I/O, Network
	utils := []float64{cpuUtil, memoryUtil, ioUtil, networkUtil}

	totalEfficiency := 0.0
	for i, util := range utils {
		totalEfficiency += util * weights[i]
	}

	return totalEfficiency
}

// calculateTaskWasteScore calculates resource waste score
func (t *TaskUtilizationMonitor) calculateTaskWasteScore(cpuUtil, memoryUtil, efficiency float64) float64 {
	// Waste is high when utilization is low or efficiency is poor
	avgUtil := (cpuUtil + memoryUtil) / 2.0
	wasteFromUtil := 1.0 - avgUtil
	wasteFromEff := 1.0 - efficiency

	return (wasteFromUtil + wasteFromEff) / 2.0
}

// simulateTaskErrorRate simulates task error rate
func (t *TaskUtilizationMonitor) simulateTaskErrorRate(job *slurm.Job, task *TaskInfo) float64 {
	// Most tasks have low error rates
	baseErrorRate := 0.01 + (float64(len(task.TaskID)%10) * 0.005)

	// Increase error rate for longer-running jobs
	if job.StartTime != nil {
		elapsed := time.Since(*job.StartTime).Hours()
		if elapsed > 24 {
			baseErrorRate += 0.02
		}
	}

	return math.Min(0.1, baseErrorRate)
}

// calculateTaskLoadScore calculates load score for a task
func (t *TaskUtilizationMonitor) calculateTaskLoadScore(cpuUtil, memoryUtil, throughput float64) float64 {
	// Load score combines utilization and throughput
	utilScore := (cpuUtil + memoryUtil) / 2.0
	throughputScore := throughput / 100.0

	return (utilScore + throughputScore) / 2.0
}

// calculateTaskLoadImbalance calculates load imbalance for a task
func (t *TaskUtilizationMonitor) calculateTaskLoadImbalance(jobID, stepID string, taskLoadScore float64) float64 {
	// Simple load imbalance calculation
	// In reality, this would compare against other tasks in the same step

	avgLoadScore := 0.7 // Assume average load score
	imbalance := math.Abs(taskLoadScore - avgLoadScore) / avgLoadScore

	return math.Min(1.0, imbalance)
}

// simulateMessagesSent simulates messages sent by a task
func (t *TaskUtilizationMonitor) simulateMessagesSent(job *slurm.Job, task *TaskInfo) int64 {
	baseMessages := int64(100 + len(task.TaskID)%500)

	// Multi-node jobs send more messages
	nodeCount := len(job.Nodes)
	if nodeCount > 1 {
		baseMessages *= int64(nodeCount)
	}

	return baseMessages
}

// simulateMessagesReceived simulates messages received by a task
func (t *TaskUtilizationMonitor) simulateMessagesReceived(job *slurm.Job, task *TaskInfo) int64 {
	// Usually similar to messages sent
	sent := t.simulateMessagesSent(job, task)
	return sent + int64(len(task.TaskID)%100) - 50 // Add some variation
}

// simulateDataTransferred simulates data transferred by a task
func (t *TaskUtilizationMonitor) simulateDataTransferred(job *slurm.Job, task *TaskInfo) float64 {
	baseMB := 10.0 + float64(len(task.TaskID)%100)

	// Larger jobs transfer more data
	if job.CPUs > 16 {
		baseMB *= 2.0
	}

	nodeCount := len(job.Nodes)
	if nodeCount > 1 {
		baseMB *= float64(nodeCount) * 0.5
	}

	return baseMB
}

// calculateTaskProgress calculates task progress
func (t *TaskUtilizationMonitor) calculateTaskProgress(duration time.Duration, timeLimit int) float64 {
	if timeLimit <= 0 {
		return 0.5 // Default progress if no time limit
	}

	timeLimitDuration := time.Duration(timeLimit) * time.Minute
	progress := duration.Seconds() / timeLimitDuration.Seconds()

	// Add some randomness for more realistic progress
	variation := float64(len(fmt.Sprintf("%.0f", duration.Seconds()))%20) / 100.0
	progress += variation - 0.1

	return math.Max(0.0, math.Min(1.0, progress))
}

// determineTaskState determines task state based on progress and error rate
func (t *TaskUtilizationMonitor) determineTaskState(progress, errorRate float64) string {
	if errorRate > 0.05 {
		return "failed"
	}

	if progress >= 1.0 {
		return "completed"
	}

	if progress > 0.01 {
		return "running"
	}

	return "idle"
}

// isTaskOnCriticalPath determines if a task is on the critical path
func (t *TaskUtilizationMonitor) isTaskOnCriticalPath(job *slurm.Job, step *JobStepInfo, task *TaskInfo) bool {
	// Simple heuristic - tasks with high CPU allocation or first/last tasks
	if task.CPUCount >= job.CPUs/2 {
		return true
	}

	if task.TaskID == "0_0" { // First task
		return true
	}

	return false
}

// estimateTaskCompletion estimates task completion time
func (t *TaskUtilizationMonitor) estimateTaskCompletion(startTime time.Time, progress float64, timeLimit int) *time.Time {
	if progress <= 0 || timeLimit <= 0 {
		return nil
	}

	elapsed := time.Since(startTime)
	estimatedTotal := elapsed.Seconds() / progress
	completion := startTime.Add(time.Duration(estimatedTotal) * time.Second)

	return &completion
}

// generateBalancingRecommendation generates load balancing recommendation
func (t *TaskUtilizationMonitor) generateBalancingRecommendation(loadScore, loadImbalance float64) string {
	if loadImbalance > 0.3 {
		if loadScore > 0.8 {
			return "migrate_to_less_loaded_node"
		} else {
			return "increase_task_allocation"
		}
	}

	if loadScore < 0.3 {
		return "reduce_task_allocation"
	}

	return "optimal"
}

// aggregateStepData aggregates task data into step-level metrics
func (t *TaskUtilizationMonitor) aggregateStepData(job *slurm.Job, step *JobStepInfo, tasks []*JobStepTaskData) *JobStepData {
	if len(tasks) == 0 {
		return &JobStepData{
			JobID:     fmt.Sprintf("job_%s", job.ID),
			StepID:    step.StepID,
			Timestamp: time.Now(),
		}
	}

	// Count task states
	totalTasks := len(tasks)
	activeTasks := 0
	completedTasks := 0
	failedTasks := 0

	// Accumulate metrics for averaging
	var cpuUtils, memoryUtils, throughputs, latencies, efficiencies, wastes []float64
	var loadScores []float64
	var criticalPathTasks []string
	var messagesTotal int64
	var dataTotal float64

	for _, task := range tasks {
		switch task.TaskState {
		case "running":
			activeTasks++
		case "completed":
			completedTasks++
		case "failed":
			failedTasks++
		}

		cpuUtils = append(cpuUtils, task.CPUUtilization)
		memoryUtils = append(memoryUtils, task.MemoryUtilization)
		throughputs = append(throughputs, task.TaskThroughput)
		latencies = append(latencies, task.TaskLatency)
		efficiencies = append(efficiencies, task.TaskEfficiency)
		wastes = append(wastes, task.TaskWasteScore)
		loadScores = append(loadScores, task.LoadScore)

		messagesTotal += task.MessagesSent + task.MessagesReceived
		dataTotal += task.DataTransferredMB

		if task.CriticalPath {
			criticalPathTasks = append(criticalPathTasks, task.TaskID)
		}
	}

	// Calculate averages and statistics
	avgCPU := average(cpuUtils)
	minCPU := minimum(cpuUtils)
	maxCPU := maximum(cpuUtils)
	stdDevCPU := standardDeviation(cpuUtils)

	avgMemory := average(memoryUtils)
	minMemory := minimum(memoryUtils)
	maxMemory := maximum(memoryUtils)
	stdDevMemory := standardDeviation(memoryUtils)

	avgThroughput := average(throughputs)
	avgLatency := average(latencies)
	avgEfficiency := average(efficiencies)
	totalWaste := sum(wastes)

	// Calculate load balancing metrics
	loadImbalanceScore := standardDeviation(loadScores) / average(loadScores)
	loadVarianceScore := variance(loadScores)
	taskDistributionScore := t.calculateTaskDistributionScore(tasks)

	// Calculate communication efficiency
	commEfficiency := t.calculateCommunicationEfficiency(tasks)

	// Calculate step progress
	progresses := make([]float64, len(tasks))
	for i, task := range tasks {
		progresses[i] = task.TaskProgress
	}
	stepProgress := average(progresses)

	// Estimate step completion
	var stepCompletion *time.Time
	if stepProgress > 0 {
		maxCompletion := time.Time{}
		for _, task := range tasks {
			if task.EstimatedCompletion != nil && task.EstimatedCompletion.After(maxCompletion) {
				maxCompletion = *task.EstimatedCompletion
			}
		}
		if !maxCompletion.IsZero() {
			stepCompletion = &maxCompletion
		}
	}

	return &JobStepData{
		JobID:         fmt.Sprintf("job_%s", job.ID),
		StepID:        step.StepID,
		Timestamp:     time.Now(),

		TotalTasks:     totalTasks,
		ActiveTasks:    activeTasks,
		CompletedTasks: completedTasks,
		FailedTasks:    failedTasks,

		AvgCPUUtilization:     avgCPU,
		MinCPUUtilization:     minCPU,
		MaxCPUUtilization:     maxCPU,
		StdDevCPUUtilization:  stdDevCPU,

		AvgMemoryUtilization:  avgMemory,
		MinMemoryUtilization:  minMemory,
		MaxMemoryUtilization:  maxMemory,
		StdDevMemoryUtilization: stdDevMemory,

		AvgTaskThroughput:     avgThroughput,
		AvgTaskLatency:        avgLatency,
		AvgTaskEfficiency:     avgEfficiency,
		TotalTaskWaste:        totalWaste,

		LoadImbalanceScore:    loadImbalanceScore,
		LoadVarianceScore:     loadVarianceScore,
		TaskDistributionScore: taskDistributionScore,

		CriticalPathTasks:     criticalPathTasks,

		TotalMessagesExchanged: messagesTotal,
		TotalDataTransferredMB: dataTotal,
		CommunicationEfficiency: commEfficiency,

		StepProgress:          stepProgress,
		EstimatedStepCompletion: stepCompletion,
	}
}

// Helper functions for statistical calculations

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func minimum(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

func maximum(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	avg := average(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

func standardDeviation(values []float64) float64 {
	return math.Sqrt(variance(values))
}

// calculateTaskDistributionScore calculates how evenly tasks are distributed
func (t *TaskUtilizationMonitor) calculateTaskDistributionScore(tasks []*JobStepTaskData) float64 {
	if len(tasks) == 0 {
		return 1.0
	}

	// Count tasks per node
	nodeTaskCount := make(map[string]int)
	for _, task := range tasks {
		nodeTaskCount[task.NodeID]++
	}

	// Calculate distribution variance
	var counts []float64
	for _, count := range nodeTaskCount {
		counts = append(counts, float64(count))
	}

	if len(counts) <= 1 {
		return 1.0
	}

	// Lower variance = better distribution
	avgCount := average(counts)
	varianceScore := variance(counts)
	distributionScore := 1.0 / (1.0 + varianceScore/avgCount)

	return distributionScore
}

// calculateCommunicationEfficiency calculates communication efficiency
func (t *TaskUtilizationMonitor) calculateCommunicationEfficiency(tasks []*JobStepTaskData) float64 {
	if len(tasks) == 0 {
		return 1.0
	}

	totalMessages := int64(0)
	totalData := 0.0
	totalThroughput := 0.0

	for _, task := range tasks {
		totalMessages += task.MessagesSent + task.MessagesReceived
		totalData += task.DataTransferredMB
		totalThroughput += task.TaskThroughput
	}

	if totalMessages == 0 || totalData == 0 {
		return 1.0
	}

	// Simple efficiency model: higher throughput with lower communication overhead
	avgThroughput := totalThroughput / float64(len(tasks))
	commOverhead := float64(totalMessages) / totalData

	efficiency := avgThroughput / (1.0 + commOverhead)
	return math.Min(1.0, efficiency/100.0)
}

// countJobTasks counts total tasks for a job
func (t *TaskUtilizationMonitor) countJobTasks(jobID string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	count := 0
	for key := range t.taskData {
		if len(key) > len(jobID) && key[:len(jobID)] == jobID {
			count++
		}
	}
	return count
}

// updateTaskMetrics updates Prometheus metrics for a task
func (t *TaskUtilizationMonitor) updateTaskMetrics(job *slurm.Job, taskData *JobStepTaskData) {
	labels := []string{
		taskData.JobID, taskData.StepID, taskData.TaskID, taskData.NodeID,
		"", "", job.Partition, // TODO: job.User and job.Account fields not available
	}

	// Update utilization metrics
	t.metrics.TaskCPUUtilization.WithLabelValues(labels...).Set(taskData.CPUUtilization)
	t.metrics.TaskMemoryUtilization.WithLabelValues(labels...).Set(taskData.MemoryUtilization)
	t.metrics.TaskIOUtilization.WithLabelValues(labels...).Set(taskData.IOUtilization)
	t.metrics.TaskNetworkUtilization.WithLabelValues(labels...).Set(taskData.NetworkUtilization)

	// Update performance metrics
	t.metrics.TaskThroughput.WithLabelValues(labels...).Set(taskData.TaskThroughput)
	t.metrics.TaskLatency.WithLabelValues(labels...).Set(taskData.TaskLatency)
	t.metrics.TaskEfficiency.WithLabelValues(labels...).Set(taskData.TaskEfficiency)
	t.metrics.TaskWasteScore.WithLabelValues(labels...).Set(taskData.TaskWasteScore)
	t.metrics.TaskErrorRate.WithLabelValues(labels...).Set(taskData.TaskErrorRate)

	// Update load balancing metrics
	t.metrics.TaskLoadScore.WithLabelValues(labels...).Set(taskData.LoadScore)
	t.metrics.TaskLoadImbalance.WithLabelValues(labels...).Set(taskData.LoadImbalance)

	// Update communication metrics
	t.metrics.TaskMessagesExchanged.WithLabelValues(taskData.JobID, taskData.StepID, taskData.TaskID, "sent").Add(float64(taskData.MessagesSent))
	t.metrics.TaskMessagesExchanged.WithLabelValues(taskData.JobID, taskData.StepID, taskData.TaskID, "received").Add(float64(taskData.MessagesReceived))
	t.metrics.TaskDataTransferred.WithLabelValues(taskData.JobID, taskData.StepID, taskData.TaskID, "total").Add(taskData.DataTransferredMB * 1024 * 1024) // Convert to bytes

	// Update task state metrics
	taskLabels := []string{taskData.JobID, taskData.StepID, "", "", job.Partition} // TODO: job.User and job.Account fields not available
	switch taskData.TaskState {
	case "running":
		t.metrics.TasksRunning.WithLabelValues(taskLabels...).Inc()
	case "completed":
		t.metrics.TasksCompleted.WithLabelValues(taskLabels...).Inc()
	case "failed":
		t.metrics.TasksFailed.WithLabelValues(taskLabels...).Inc()
	}

	// Update progress
	progressLabels := []string{taskData.JobID, taskData.StepID, taskData.TaskID, "", "", job.Partition} // TODO: job.User and job.Account fields not available
	t.metrics.TaskProgress.WithLabelValues(progressLabels...).Set(taskData.TaskProgress)
}

// updateStepMetrics updates Prometheus metrics for a step
func (t *TaskUtilizationMonitor) updateStepMetrics(job *slurm.Job, stepData *JobStepData) {
	labels := []string{stepData.JobID, stepData.StepID, "", "", job.Partition} // TODO: job.User and job.Account fields not available

	// Update step-level aggregation metrics
	t.metrics.StepAvgCPUUtilization.WithLabelValues(labels...).Set(stepData.AvgCPUUtilization)
	t.metrics.StepMaxCPUUtilization.WithLabelValues(labels...).Set(stepData.MaxCPUUtilization)
	t.metrics.StepCPUVariance.WithLabelValues(labels...).Set(stepData.StdDevCPUUtilization)
	t.metrics.StepAvgMemoryUtilization.WithLabelValues(labels...).Set(stepData.AvgMemoryUtilization)
	t.metrics.StepMemoryVariance.WithLabelValues(labels...).Set(stepData.StdDevMemoryUtilization)

	// Update step performance metrics
	t.metrics.StepAvgThroughput.WithLabelValues(labels...).Set(stepData.AvgTaskThroughput)
	t.metrics.StepAvgLatency.WithLabelValues(labels...).Set(stepData.AvgTaskLatency)
	t.metrics.StepAvgEfficiency.WithLabelValues(labels...).Set(stepData.AvgTaskEfficiency)
	t.metrics.StepLoadImbalanceScore.WithLabelValues(labels...).Set(stepData.LoadImbalanceScore)

	// Update load balancing metrics
	t.metrics.LoadVarianceScore.WithLabelValues(labels...).Set(stepData.LoadVarianceScore)

	// Update communication efficiency
	t.metrics.CommunicationEfficiency.WithLabelValues(labels...).Set(stepData.CommunicationEfficiency)

	// Update critical path metrics
	t.metrics.CriticalPathTasks.WithLabelValues(labels...).Set(float64(len(stepData.CriticalPathTasks)))
}
*/

// Performance analysis methods (simplified implementations)

func (a *TaskPerformanceAnalyzer) analyzeTaskPerformance() error {
	// Placeholder for task performance analysis
	return nil
}

/*
func (a *TaskPerformanceAnalyzer) trackTaskPerformance(taskData *JobStepTaskData) {
	// Placeholder for tracking task performance data
}
*/

func (l *TaskLoadBalancer) analyzeLoadBalancing() error {
	// Placeholder for load balancing analysis
	return nil
}

/*
func (l *TaskLoadBalancer) trackTaskLoad(taskData *JobStepTaskData) {
	// Placeholder for tracking task load data
}

func (e *TaskEfficiencyTracker) trackTaskEfficiency(taskData *JobStepTaskData) {
	// Placeholder for tracking task efficiency
}
*/

// cleanOldTaskData removes old task and step data
func (t *TaskUtilizationMonitor) cleanOldTaskData() {
	t.mu.Lock()
	defer t.mu.Unlock()

	taskCutoff := time.Now().Add(-t.config.TaskDataRetention)
	stepCutoff := time.Now().Add(-t.config.StepDataRetention)

	for key, data := range t.taskData {
		if data.Timestamp.Before(taskCutoff) {
			delete(t.taskData, key)
		}
	}

	for key, data := range t.stepData {
		if data.Timestamp.Before(stepCutoff) {
			delete(t.stepData, key)
		}
	}
}

// getTaskBenchmarks returns task performance benchmarks
func getTaskBenchmarks() map[string]float64 {
	return map[string]float64{
		"cpu_efficiency":           0.80,
		"memory_efficiency":        0.75,
		"io_efficiency":            0.70,
		"network_efficiency":       0.65,
		"overall_efficiency":       0.75,
		"load_balance":             0.90,
		"communication_efficiency": 0.85,
	}
}

// GetTaskData returns task data for a specific task
func (t *TaskUtilizationMonitor) GetTaskData(jobID, stepID, taskID string) (*JobStepTaskData, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", jobID, stepID, taskID)
	data, exists := t.taskData[key]
	return data, exists
}

// GetStepData returns step data for a specific step
func (t *TaskUtilizationMonitor) GetStepData(jobID, stepID string) (*JobStepData, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", jobID, stepID)
	data, exists := t.stepData[key]
	return data, exists
}

// GetTaskMonitoringStats returns task monitoring statistics
func (t *TaskUtilizationMonitor) GetTaskMonitoringStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return map[string]interface{}{
		"monitored_tasks": len(t.taskData),
		"monitored_steps": len(t.stepData),
		"last_collection": t.lastCollection,
		"config":          t.config,
	}
}
