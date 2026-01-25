// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// MetricDefinitions holds all Prometheus metric definitions for the SLURM exporter
type MetricDefinitions struct {
	// Label management
	LabelManager *LabelManager
	// Cluster Overview Metrics
	ClusterInfo              *prometheus.GaugeVec
	ClusterCapacityCPUs      *prometheus.GaugeVec
	ClusterCapacityMemory    *prometheus.GaugeVec
	ClusterCapacityNodes     *prometheus.GaugeVec
	ClusterAllocatedCPUs     *prometheus.GaugeVec
	ClusterAllocatedMemory   *prometheus.GaugeVec
	ClusterAllocatedNodes    *prometheus.GaugeVec
	ClusterUtilizationCPUs   *prometheus.GaugeVec
	ClusterUtilizationMemory *prometheus.GaugeVec
	ClusterUtilizationNodes  *prometheus.GaugeVec
	ClusterNodeStates        *prometheus.GaugeVec
	ClusterPartitionCount    *prometheus.GaugeVec
	ClusterJobCount          *prometheus.GaugeVec
	ClusterUserCount         *prometheus.GaugeVec

	// Node-Level Metrics
	NodeInfo            *prometheus.GaugeVec
	NodeCPUs            *prometheus.GaugeVec
	NodeMemory          *prometheus.GaugeVec
	NodeState           *prometheus.GaugeVec
	NodeAllocatedCPUs   *prometheus.GaugeVec
	NodeAllocatedMemory *prometheus.GaugeVec
	NodeLoad            *prometheus.GaugeVec
	NodeUptime          *prometheus.GaugeVec
	NodeLastHeartbeat   *prometheus.GaugeVec
	NodeFeatures        *prometheus.GaugeVec
	NodeGPUs            *prometheus.GaugeVec
	NodeWeight          *prometheus.GaugeVec

	// Job Metrics
	JobInfo            *prometheus.GaugeVec
	JobStates          *prometheus.GaugeVec
	JobQueueTime       *prometheus.HistogramVec
	JobRunTime         *prometheus.HistogramVec
	JobWaitTime        *prometheus.HistogramVec
	JobCPURequested    *prometheus.GaugeVec
	JobMemoryRequested *prometheus.GaugeVec
	JobNodesRequested  *prometheus.GaugeVec
	JobCPUAllocated    *prometheus.GaugeVec
	JobMemoryAllocated *prometheus.GaugeVec
	JobNodesAllocated  *prometheus.GaugeVec
	JobPriority        *prometheus.GaugeVec
	JobStartTime       *prometheus.GaugeVec
	JobEndTime         *prometheus.GaugeVec
	JobExitCode        *prometheus.GaugeVec

	// User and Account Metrics
	UserJobCount           *prometheus.GaugeVec
	UserCPUAllocated       *prometheus.GaugeVec
	UserMemoryAllocated    *prometheus.GaugeVec
	UserJobStates          *prometheus.GaugeVec
	AccountInfo            *prometheus.GaugeVec
	AccountJobCount        *prometheus.GaugeVec
	AccountCPUAllocated    *prometheus.GaugeVec
	AccountMemoryAllocated *prometheus.GaugeVec
	AccountFairShare       *prometheus.GaugeVec
	AccountUsage           *prometheus.GaugeVec

	// Partition Metrics
	PartitionInfo            *prometheus.GaugeVec
	PartitionNodes           *prometheus.GaugeVec
	PartitionCPUs            *prometheus.GaugeVec
	PartitionMemory          *prometheus.GaugeVec
	PartitionState           *prometheus.GaugeVec
	PartitionJobCount        *prometheus.GaugeVec
	PartitionAllocatedCPUs   *prometheus.GaugeVec
	PartitionAllocatedMemory *prometheus.GaugeVec
	PartitionUtilization     *prometheus.GaugeVec
	PartitionMaxTime         *prometheus.GaugeVec
	PartitionDefaultTime     *prometheus.GaugeVec
	PartitionPriority        *prometheus.GaugeVec

	// Performance and Efficiency Metrics
	ClusterThroughput   *prometheus.GaugeVec
	ClusterEfficiency   *prometheus.GaugeVec
	QueueBacklog        *prometheus.GaugeVec
	AverageQueueTime    *prometheus.GaugeVec
	AverageWaitTime     *prometheus.GaugeVec
	JobCompletionRate   *prometheus.GaugeVec
	ResourceWastage     *prometheus.GaugeVec
	SchedulingLatency   *prometheus.HistogramVec
	SystemThroughput    *prometheus.GaugeVec
	SystemEfficiency    *prometheus.GaugeVec
	QueueWaitTime       *prometheus.GaugeVec
	QueueDepth          *prometheus.GaugeVec
	ResourceUtilization *prometheus.GaugeVec
	JobTurnover         *prometheus.GaugeVec

	// System Metrics
	SystemLoad        *prometheus.GaugeVec
	SystemMemoryUsage *prometheus.GaugeVec
	SystemDiskUsage   *prometheus.GaugeVec
	SystemNetworkIO   *prometheus.GaugeVec
	SystemCPUUsage    *prometheus.GaugeVec

	// Exporter Self-Monitoring
	CollectionDuration *prometheus.HistogramVec
	CollectionErrors   *prometheus.CounterVec
	CollectionSuccess  *prometheus.CounterVec
	CollectorUp        *prometheus.GaugeVec
	LastCollectionTime *prometheus.GaugeVec
	MetricsExported    *prometheus.CounterVec
	APICallDuration    *prometheus.HistogramVec
	APICallErrors      *prometheus.CounterVec
	CacheHits          *prometheus.CounterVec
	CacheMisses        *prometheus.CounterVec
}

// NewMetricDefinitions creates all metric definitions with the given namespace and subsystem
func NewMetricDefinitions(namespace, subsystem string, constLabels prometheus.Labels) *MetricDefinitions {
	return &MetricDefinitions{
		// Initialize label manager
		LabelManager: NewLabelManager(),
		// Cluster Overview Metrics
		ClusterInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_info",
				Help:        "Information about the SLURM cluster (version, build time, etc.)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "version", "build_time", "controller_host"},
		),

		ClusterCapacityCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_capacity_cpus",
				Help:        "Total CPU capacity of the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterCapacityMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_capacity_memory_bytes",
				Help:        "Total memory capacity of the cluster in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterCapacityNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_capacity_nodes",
				Help:        "Total number of nodes in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterAllocatedCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_allocated_cpus",
				Help:        "Number of allocated CPUs in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterAllocatedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_allocated_memory_bytes",
				Help:        "Amount of allocated memory in the cluster in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterAllocatedNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_allocated_nodes",
				Help:        "Number of allocated nodes in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterUtilizationCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_utilization_cpus_ratio",
				Help:        "CPU utilization ratio of the cluster (0-1)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterUtilizationMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_utilization_memory_ratio",
				Help:        "Memory utilization ratio of the cluster (0-1)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterUtilizationNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_utilization_nodes_ratio",
				Help:        "Node utilization ratio of the cluster (0-1)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterNodeStates: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_node_states",
				Help:        "Number of nodes in each state",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "state"},
		),

		ClusterPartitionCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_partitions_total",
				Help:        "Total number of partitions in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		ClusterJobCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_jobs_total",
				Help:        "Total number of jobs in the cluster by state",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "state"},
		),

		ClusterUserCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "cluster_users_total",
				Help:        "Total number of active users in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name"},
		),

		// Node-Level Metrics
		NodeInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_info",
				Help:        "Information about cluster nodes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name", "architecture", "os", "kernel"},
		),

		NodeCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_cpus",
				Help:        "Number of CPUs on the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_memory_bytes",
				Help:        "Amount of memory on the node in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_state",
				Help:        "Current state of the node (1 for the current state, 0 otherwise)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name", "state"},
		),

		NodeAllocatedCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_allocated_cpus",
				Help:        "Number of allocated CPUs on the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeAllocatedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_allocated_memory_bytes",
				Help:        "Amount of allocated memory on the node in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeLoad: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_load_average",
				Help:        "Load average of the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name", "interval"},
		),

		NodeUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_uptime_seconds",
				Help:        "Uptime of the node in seconds",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeLastHeartbeat: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_last_heartbeat_timestamp",
				Help:        "Timestamp of the last heartbeat from the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		NodeFeatures: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_features",
				Help:        "Features available on the node (1 if feature is available, 0 otherwise)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name", "feature"},
		),

		NodeGPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_gpus",
				Help:        "Number of GPUs on the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name", "gpu_type"},
		),

		NodeWeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "node_weight",
				Help:        "Scheduling weight of the node",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "node_name"},
		),

		// Job Metrics
		JobInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_info",
				Help:        "Information about jobs in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "job_name", "user", "account", "partition", "state"},
		),

		JobStates: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_states",
				Help:        "Number of jobs in each state",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "state"},
		),

		JobQueueTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_queue_time_seconds",
				Help:        "Time jobs spend in the queue before starting",
				ConstLabels: constLabels,
				Buckets:     []float64{30, 60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
			},
			[]string{"cluster_name", "partition", "account"},
		),

		JobRunTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_run_time_seconds",
				Help:        "Time jobs spend running",
				ConstLabels: constLabels,
				Buckets:     []float64{60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400, 172800},
			},
			[]string{"cluster_name", "partition", "account"},
		),

		JobWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_wait_time_seconds",
				Help:        "Total time jobs wait (queue + any delays)",
				ConstLabels: constLabels,
				Buckets:     []float64{30, 60, 300, 600, 1800, 3600, 7200, 14400, 28800, 86400},
			},
			[]string{"cluster_name", "partition", "account"},
		),

		JobCPURequested: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_cpus_requested",
				Help:        "Number of CPUs requested by the job",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobMemoryRequested: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_memory_requested_bytes",
				Help:        "Amount of memory requested by the job in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobNodesRequested: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_nodes_requested",
				Help:        "Number of nodes requested by the job",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_cpus_allocated",
				Help:        "Number of CPUs allocated to the job",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_memory_allocated_bytes",
				Help:        "Amount of memory allocated to the job in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobNodesAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_nodes_allocated",
				Help:        "Number of nodes allocated to the job",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_priority",
				Help:        "Priority of the job",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobStartTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_start_timestamp",
				Help:        "Start time of the job as Unix timestamp",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobEndTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_end_timestamp",
				Help:        "End time of the job as Unix timestamp",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		JobExitCode: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_exit_code",
				Help:        "Exit code of completed jobs",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "job_id", "user", "account", "partition"},
		),

		// User and Account Metrics
		UserJobCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "user_job_count",
				Help:        "Number of jobs for each user",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "user", "account"},
		),

		UserCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "user_cpu_allocated",
				Help:        "Number of CPUs allocated to each user",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "user", "account"},
		),

		UserMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "user_memory_allocated_bytes",
				Help:        "Amount of memory allocated to each user in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "user", "account"},
		),

		UserJobStates: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "user_job_states",
				Help:        "Number of jobs in each state for each user",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "user", "account", "state"},
		),

		AccountInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_info",
				Help:        "Information about accounts in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account", "organization", "description", "parent"},
		),

		AccountJobCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_job_count",
				Help:        "Number of jobs for each account",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account"},
		),

		AccountCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_cpu_allocated",
				Help:        "Number of CPUs allocated to each account",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account"},
		),

		AccountMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_memory_allocated_bytes",
				Help:        "Amount of memory allocated to each account in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account"},
		),

		AccountFairShare: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_fair_share",
				Help:        "Fair-share score for each account",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account"},
		),

		AccountUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "account_usage",
				Help:        "Resource usage for each account by usage type",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "account", "usage_type"},
		),

		// Partition Metrics
		PartitionInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_info",
				Help:        "Information about partitions in the cluster",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition", "state", "allow_groups", "allow_users"},
		),

		PartitionNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_nodes",
				Help:        "Number of nodes in each partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_cpus",
				Help:        "Number of CPUs in each partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_memory_bytes",
				Help:        "Amount of memory in each partition in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_state",
				Help:        "Current state of each partition (1 for current state, 0 otherwise)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition", "state"},
		),

		PartitionJobCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_job_count",
				Help:        "Number of jobs in each partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionAllocatedCPUs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_allocated_cpus",
				Help:        "Number of allocated CPUs in each partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionAllocatedMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_allocated_memory_bytes",
				Help:        "Amount of allocated memory in each partition in bytes",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_utilization_ratio",
				Help:        "Resource utilization ratio for each partition by resource type",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition", "resource_type"},
		),

		PartitionMaxTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_max_time_seconds",
				Help:        "Maximum time limit for jobs in each partition in seconds",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionDefaultTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_default_time_seconds",
				Help:        "Default time limit for jobs in each partition in seconds",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		PartitionPriority: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "partition_priority",
				Help:        "Priority of each partition for scheduling",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		// Performance Metrics
		SystemThroughput: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "system_throughput",
				Help:        "System throughput metrics (jobs/hour, CPU hours/hour, etc.)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "metric_type", "time_window", "partition"},
		),

		SystemEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "system_efficiency",
				Help:        "System efficiency metrics (resource utilization ratios)",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "efficiency_type", "partition", "time_window"},
		),

		QueueWaitTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "queue_wait_time_seconds",
				Help:        "Queue wait time statistics in seconds",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition", "statistic", "priority"},
		),

		QueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "queue_depth",
				Help:        "Number of jobs waiting in queue by partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition", "priority"},
		),

		ResourceUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "resource_utilization_ratio",
				Help:        "Resource utilization ratio by type and partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "resource_type", "partition"},
		),

		JobTurnover: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "job_turnover_rate",
				Help:        "Job turnover rate (jobs completed per hour) by partition",
				ConstLabels: constLabels,
			},
			[]string{"cluster_name", "partition"},
		),

		// Exporter Self-Monitoring
		CollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "collection_duration_seconds",
				Help:        "Time spent collecting metrics from SLURM",
				ConstLabels: constLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"collector"},
		),

		CollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "collection_errors_total",
				Help:        "Total number of collection errors",
				ConstLabels: constLabels,
			},
			[]string{"collector", "error_type"},
		),

		CollectionSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "collection_success_total",
				Help:        "Total number of successful collections",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),

		CollectorUp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "collector_up",
				Help:        "Whether the collector is up and working (1 = up, 0 = down)",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),

		LastCollectionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "last_collection_timestamp",
				Help:        "Unix timestamp of the last successful collection",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),

		MetricsExported: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "metrics_exported_total",
				Help:        "Total number of metrics exported",
				ConstLabels: constLabels,
			},
			[]string{"collector", "metric_name"},
		),

		APICallDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "api_call_duration_seconds",
				Help:        "Time spent on SLURM API calls",
				ConstLabels: constLabels,
				Buckets:     []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"endpoint", "method"},
		),

		APICallErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "api_call_errors_total",
				Help:        "Total number of SLURM API call errors",
				ConstLabels: constLabels,
			},
			[]string{"endpoint", "method", "status_code"},
		),

		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "cache_hits_total",
				Help:        "Total number of cache hits",
				ConstLabels: constLabels,
			},
			[]string{"cache_type"},
		),

		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   "exporter",
				Name:        "cache_misses_total",
				Help:        "Total number of cache misses",
				ConstLabels: constLabels,
			},
			[]string{"cache_type"},
		),
	}
}

// allMetrics returns all metrics as a slice
func (md *MetricDefinitions) allMetrics() []prometheus.Collector {
	return []prometheus.Collector{
		// Cluster metrics
		md.ClusterInfo,
		md.ClusterCapacityCPUs,
		md.ClusterCapacityMemory,
		md.ClusterCapacityNodes,
		md.ClusterAllocatedCPUs,
		md.ClusterAllocatedMemory,
		md.ClusterAllocatedNodes,
		md.ClusterUtilizationCPUs,
		md.ClusterUtilizationMemory,
		md.ClusterUtilizationNodes,
		md.ClusterNodeStates,
		md.ClusterPartitionCount,
		md.ClusterJobCount,
		md.ClusterUserCount,

		// Node metrics
		md.NodeInfo,
		md.NodeCPUs,
		md.NodeMemory,
		md.NodeState,
		md.NodeAllocatedCPUs,
		md.NodeAllocatedMemory,
		md.NodeLoad,
		md.NodeUptime,
		md.NodeLastHeartbeat,
		md.NodeFeatures,
		md.NodeGPUs,
		md.NodeWeight,

		// Job metrics
		md.JobInfo,
		md.JobStates,
		md.JobQueueTime,
		md.JobRunTime,
		md.JobWaitTime,
		md.JobCPURequested,
		md.JobMemoryRequested,
		md.JobNodesRequested,
		md.JobCPUAllocated,
		md.JobMemoryAllocated,
		md.JobNodesAllocated,
		md.JobPriority,
		md.JobStartTime,
		md.JobEndTime,
		md.JobExitCode,

		// User and account metrics
		md.UserJobCount,
		md.UserCPUAllocated,
		md.UserMemoryAllocated,
		md.UserJobStates,
		md.AccountInfo,
		md.AccountJobCount,
		md.AccountCPUAllocated,
		md.AccountMemoryAllocated,
		md.AccountFairShare,
		md.AccountUsage,

		// Partition metrics
		md.PartitionInfo,
		md.PartitionNodes,
		md.PartitionCPUs,
		md.PartitionMemory,
		md.PartitionState,
		md.PartitionJobCount,
		md.PartitionAllocatedCPUs,
		md.PartitionAllocatedMemory,
		md.PartitionUtilization,
		md.PartitionMaxTime,
		md.PartitionDefaultTime,
		md.PartitionPriority,

		// Performance metrics
		md.SystemThroughput,
		md.SystemEfficiency,
		md.QueueWaitTime,
		md.QueueDepth,
		md.ResourceUtilization,
		md.JobTurnover,

		// Exporter metrics
		md.CollectionDuration,
		md.CollectionErrors,
		md.CollectionSuccess,
		md.CollectorUp,
		md.LastCollectionTime,
		md.MetricsExported,
		md.APICallDuration,
		md.APICallErrors,
		md.CacheHits,
		md.CacheMisses,
	}
}

// Register registers all metrics with the given registry
func (md *MetricDefinitions) Register(registry *prometheus.Registry) error {
	for _, metric := range md.allMetrics() {
		if err := registry.Register(metric); err != nil {
			return err
		}
	}

	return nil
}

// Unregister unregisters all metrics from the given registry
func (md *MetricDefinitions) Unregister(registry *prometheus.Registry) {
	for _, metric := range md.allMetrics() {
		registry.Unregister(metric)
	}
}
