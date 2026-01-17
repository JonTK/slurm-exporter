package collector

import (
	"context"
	// Commented out as they're only used in commented-out parser functions
	// "strconv"
	// "strings"
	// "time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// ClusterCollector collects cluster-level overview metrics
type ClusterCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewClusterCollector creates a new cluster collector
func NewClusterCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *ClusterCollector {
	base := NewBaseCollector("cluster", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &ClusterCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("cluster", base.logger),
	}
}

// Describe implements the Collector interface
func (cc *ClusterCollector) Describe(ch chan<- *prometheus.Desc) {
	// Cluster overview metrics
	cc.metrics.ClusterInfo.Describe(ch)
	cc.metrics.ClusterCapacityCPUs.Describe(ch)
	cc.metrics.ClusterCapacityMemory.Describe(ch)
	cc.metrics.ClusterCapacityNodes.Describe(ch)
	cc.metrics.ClusterAllocatedCPUs.Describe(ch)
	cc.metrics.ClusterAllocatedMemory.Describe(ch)
	cc.metrics.ClusterAllocatedNodes.Describe(ch)
	cc.metrics.ClusterUtilizationCPUs.Describe(ch)
	cc.metrics.ClusterUtilizationMemory.Describe(ch)
	cc.metrics.ClusterUtilizationNodes.Describe(ch)
	cc.metrics.ClusterNodeStates.Describe(ch)
	cc.metrics.ClusterPartitionCount.Describe(ch)
	cc.metrics.ClusterJobCount.Describe(ch)
	cc.metrics.ClusterUserCount.Describe(ch)
}

// Collect implements the Collector interface
func (cc *ClusterCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return cc.CollectWithMetrics(ctx, ch, cc.collectClusterMetrics)
}

// collectClusterMetrics performs the actual cluster metrics collection
func (cc *ClusterCollector) collectClusterMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	cc.logger.Debug("Starting cluster metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch data

	// Collect cluster information
	if err := cc.collectClusterInfo(ctx, ch); err != nil {
		return cc.errorBuilder.Internal(err, "collect_cluster_info",
			"Check SLURM controller connectivity",
			"Verify SLURM REST API is accessible")
	}

	// Collect node information for cluster aggregates
	if err := cc.collectNodeSummary(ctx, ch); err != nil {
		return cc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM node data availability",
			"Verify node information is up to date")
	}

	// Collect job information for cluster aggregates
	if err := cc.collectJobSummary(ctx, ch); err != nil {
		return cc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM job data availability",
			"Verify job information is accessible")
	}

	// Collect partition information
	if err := cc.collectPartitionSummary(ctx, ch); err != nil {
		return cc.errorBuilder.API(err, "/slurm/v1/partitions", 500, "",
			"Check SLURM partition data availability",
			"Verify partition configuration is accessible")
	}

	cc.logger.Debug("Completed cluster metrics collection")
	return nil
}

// collectClusterInfo collects basic cluster information
func (cc *ClusterCollector) collectClusterInfo(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate cluster info - in real implementation this would come from SLURM API
	clusterInfo := map[string]string{
		"cluster_name":    cc.clusterName,
		"version":         "23.02.0", // Simulated version
		"build_time":      "2023-02-15",
		"controller_host": "slurmctld.example.com",
	}

	// Send cluster info metric
	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterInfo.WithLabelValues(
			clusterInfo["cluster_name"],
			clusterInfo["version"],
			clusterInfo["build_time"],
			clusterInfo["controller_host"],
		).Desc(),
		prometheus.GaugeValue,
		1,
		clusterInfo["cluster_name"],
		clusterInfo["version"],
		clusterInfo["build_time"],
		clusterInfo["controller_host"],
	))

	cc.LogCollection("Collected cluster info: version=%s", clusterInfo["version"])
	return nil
}

// collectNodeSummary collects node summary information for cluster aggregates
func (cc *ClusterCollector) collectNodeSummary(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate node data - in real implementation this would come from SLURM API
	// This represents what we might get from /slurm/v1/nodes
	nodeData := struct {
		TotalNodes      int
		TotalCPUs       int
		TotalMemory     int64 // bytes
		AllocatedNodes  int
		AllocatedCPUs   int
		AllocatedMemory int64 // bytes
		NodeStates      map[string]int
	}{
		TotalNodes:      100,
		TotalCPUs:       4800,
		TotalMemory:     1024 * 1024 * 1024 * 1024, // 1TB
		AllocatedNodes:  75,
		AllocatedCPUs:   3600,
		AllocatedMemory: 768 * 1024 * 1024 * 1024, // 768GB
		NodeStates: map[string]int{
			"idle":      20,
			"allocated": 75,
			"down":      3,
			"drain":     2,
		},
	}

	// Cluster capacity metrics
	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterCapacityNodes.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.TotalNodes),
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterCapacityCPUs.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.TotalCPUs),
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterCapacityMemory.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.TotalMemory),
		cc.clusterName,
	))

	// Cluster allocation metrics
	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterAllocatedNodes.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.AllocatedNodes),
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterAllocatedCPUs.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.AllocatedCPUs),
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterAllocatedMemory.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(nodeData.AllocatedMemory),
		cc.clusterName,
	))

	// Cluster utilization metrics
	cpuUtilization := float64(nodeData.AllocatedCPUs) / float64(nodeData.TotalCPUs)
	memoryUtilization := float64(nodeData.AllocatedMemory) / float64(nodeData.TotalMemory)
	nodeUtilization := float64(nodeData.AllocatedNodes) / float64(nodeData.TotalNodes)

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterUtilizationCPUs.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		cpuUtilization,
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterUtilizationMemory.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		memoryUtilization,
		cc.clusterName,
	))

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterUtilizationNodes.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		nodeUtilization,
		cc.clusterName,
	))

	// Node states metrics
	for state, count := range nodeData.NodeStates {
		cc.SendMetric(ch, cc.BuildMetric(
			cc.metrics.ClusterNodeStates.WithLabelValues(cc.clusterName, state).Desc(),
			prometheus.GaugeValue,
			float64(count),
			cc.clusterName,
			state,
		))
	}

	cc.LogCollection("Collected node summary: %d total nodes, %.1f%% CPU utilization",
		nodeData.TotalNodes, cpuUtilization*100)

	return nil
}

// collectJobSummary collects job summary information for cluster aggregates
func (cc *ClusterCollector) collectJobSummary(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate job data - in real implementation this would come from SLURM API
	// This represents what we might get from /slurm/v1/jobs
	jobStates := map[string]int{
		"pending":   150,
		"running":   320,
		"completed": 1200,
		"cancelled": 45,
		"failed":    23,
		"timeout":   12,
	}

	// Extract unique users count (simulated)
	uniqueUsers := 85

	// Job states metrics
	for state, count := range jobStates {
		cc.SendMetric(ch, cc.BuildMetric(
			cc.metrics.ClusterJobCount.WithLabelValues(cc.clusterName, state).Desc(),
			prometheus.GaugeValue,
			float64(count),
			cc.clusterName,
			state,
		))
	}

	// User count metric
	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterUserCount.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(uniqueUsers),
		cc.clusterName,
	))

	totalJobs := 0
	for _, count := range jobStates {
		totalJobs += count
	}

	cc.LogCollection("Collected job summary: %d total jobs, %d unique users", totalJobs, uniqueUsers)
	return nil
}

// collectPartitionSummary collects partition summary information
func (cc *ClusterCollector) collectPartitionSummary(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate partition data - in real implementation this would come from SLURM API
	// This represents what we might get from /slurm/v1/partitions
	partitionCount := 8 // Example: compute, gpu, highmem, debug, etc.

	cc.SendMetric(ch, cc.BuildMetric(
		cc.metrics.ClusterPartitionCount.WithLabelValues(cc.clusterName).Desc(),
		prometheus.GaugeValue,
		float64(partitionCount),
		cc.clusterName,
	))

	cc.LogCollection("Collected partition summary: %d partitions", partitionCount)
	return nil
}

// Helper methods for simulating SLURM API responses
// In real implementation, these would be replaced with actual SLURM client calls

// TODO: Following parser functions are unused - preserved for future data parsing needs
/*
// parseNodeState converts SLURM node state string to normalized state
func (cc *ClusterCollector) parseNodeState(slurmState string) string {
	// SLURM node states can be complex (e.g., "idle+cloud", "allocated+completing")
	// Normalize to primary states for metrics
	switch {
	case strings.Contains(slurmState, "idle"):
		return "idle"
	case strings.Contains(slurmState, "allocated") || strings.Contains(slurmState, "mixed"):
		return "allocated"
	case strings.Contains(slurmState, "down"):
		return "down"
	case strings.Contains(slurmState, "drain"):
		return "drain"
	case strings.Contains(slurmState, "completing"):
		return "completing"
	case strings.Contains(slurmState, "maint"):
		return "maintenance"
	default:
		return "unknown"
	}
}

// parseJobState converts SLURM job state string to normalized state
func (cc *ClusterCollector) parseJobState(slurmState string) string {
	// SLURM job states: PENDING, RUNNING, SUSPENDED, COMPLETED, CANCELLED, FAILED, TIMEOUT, etc.
	switch slurmState {
	case "PENDING", "PD":
		return "pending"
	case "RUNNING", "R":
		return "running"
	case "SUSPENDED", "S":
		return "suspended"
	case "COMPLETED", "CD":
		return "completed"
	case "CANCELLED", "CA":
		return "cancelled"
	case "FAILED", "F":
		return "failed"
	case "TIMEOUT", "TO":
		return "timeout"
	case "NODE_FAIL", "NF":
		return "node_fail"
	case "PREEMPTED", "PR":
		return "preempted"
	case "BOOT_FAIL", "BF":
		return "boot_fail"
	case "DEADLINE", "DL":
		return "deadline"
	case "OUT_OF_MEMORY", "OOM":
		return "out_of_memory"
	default:
		return "unknown"
	}
}

// parseMemorySize converts SLURM memory specification to bytes
func (cc *ClusterCollector) parseMemorySize(slurmMemory string) (int64, error) {
	// SLURM memory can be in various formats: "1024M", "2G", "500000K", etc.
	if slurmMemory == "" {
		return 0, nil
	}

	// Extract numeric part and unit
	var value float64
	var unit string

	// Simple parsing - in real implementation, use more robust parsing
	if len(slurmMemory) > 0 {
		lastChar := slurmMemory[len(slurmMemory)-1]
		switch lastChar {
		case 'K', 'k':
			unit = "K"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'M', 'm':
			unit = "M"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'G', 'g':
			unit = "G"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		case 'T', 't':
			unit = "T"
			if val, err := strconv.ParseFloat(slurmMemory[:len(slurmMemory)-1], 64); err == nil {
				value = val
			}
		default:
			// Assume MB if no unit
			unit = "M"
			if val, err := strconv.ParseFloat(slurmMemory, 64); err == nil {
				value = val
			}
		}
	}

	// Convert to bytes
	switch unit {
	case "K":
		return int64(value * 1024), nil
	case "M":
		return int64(value * 1024 * 1024), nil
	case "G":
		return int64(value * 1024 * 1024 * 1024), nil
	case "T":
		return int64(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return int64(value), nil
	}
}

// parseTimeString converts SLURM time string to Unix timestamp
func (cc *ClusterCollector) parseTimeString(slurmTime string) (int64, error) {
	if slurmTime == "" || slurmTime == "Unknown" {
		return 0, nil
	}

	// SLURM typically uses format: "2023-02-15T10:30:25"
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, slurmTime); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, cc.errorBuilder.Parsing(nil, "parse_time", "timestamp",
		"Check SLURM time format configuration",
		"Verify time zone settings")
}
*/
