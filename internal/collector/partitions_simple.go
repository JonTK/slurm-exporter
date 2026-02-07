// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	partitionsCollectorSubsystem = "partition"
)

// PartitionsSimpleCollector collects partition-related metrics
type PartitionsSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Partition state metrics
	partitionState *prometheus.Desc

	// Partition node metrics
	partitionNodesTotal     *prometheus.Desc
	partitionNodesAllocated *prometheus.Desc
	partitionNodesIdle      *prometheus.Desc
	partitionNodesDown      *prometheus.Desc

	// Partition CPU metrics
	partitionCPUsTotal     *prometheus.Desc
	partitionCPUsAllocated *prometheus.Desc
	partitionCPUsIdle      *prometheus.Desc

	// Partition job metrics
	partitionJobsPending *prometheus.Desc
	partitionJobsRunning *prometheus.Desc

	// Partition info
	partitionInfo *prometheus.Desc
}

// NewPartitionsSimpleCollector creates a new Partitions collector
func NewPartitionsSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *PartitionsSimpleCollector {
	c := &PartitionsSimpleCollector{
		client:  client,
		logger:  logger.WithField("collector", "partitions"),
		enabled: true,
	}

	// Initialize metrics
	c.partitionState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "state"),
		"Current state of the partition (1=up, 0=down)",
		[]string{"partition", "state"},
		nil,
	)

	c.partitionNodesTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "nodes_total"),
		"Total number of nodes in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionNodesAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "nodes_allocated"),
		"Number of allocated nodes in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionNodesIdle = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "nodes_idle"),
		"Number of idle nodes in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionNodesDown = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "nodes_down"),
		"Number of down nodes in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionCPUsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "cpus_total"),
		"Total number of CPUs in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionCPUsAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "cpus_allocated"),
		"Number of allocated CPUs in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionCPUsIdle = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "cpus_idle"),
		"Number of idle CPUs in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionJobsPending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "jobs_pending"),
		"Number of pending jobs in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionJobsRunning = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "jobs_running"),
		"Number of running jobs in the partition",
		[]string{"partition"},
		nil,
	)

	c.partitionInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, partitionsCollectorSubsystem, "info"),
		"Partition information with all labels",
		[]string{"partition", "state", "qos", "max_time", "default_time"},
		nil,
	)

	return c
}

// Name returns the collector name
func (c *PartitionsSimpleCollector) Name() string {
	return "partitions"
}

// IsEnabled returns whether this collector is enabled
func (c *PartitionsSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *PartitionsSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *PartitionsSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.partitionState
	ch <- c.partitionNodesTotal
	ch <- c.partitionNodesAllocated
	ch <- c.partitionNodesIdle
	ch <- c.partitionNodesDown
	ch <- c.partitionCPUsTotal
	ch <- c.partitionCPUsAllocated
	ch <- c.partitionCPUsIdle
	ch <- c.partitionJobsPending
	ch <- c.partitionJobsRunning
	ch <- c.partitionInfo
}

// Collect implements the Collector interface
func (c *PartitionsSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ctx, ch)
}

// partitionStats holds aggregated statistics for a partition
type partitionStats struct {
	idleNodes     int
	downNodes     int
	allocatedCPUs int
	pendingJobs   int
	runningJobs   int
}

// publishPartitionMetrics publishes all metrics for a single partition
func (c *PartitionsSimpleCollector) publishPartitionMetrics(ch chan<- prometheus.Metric, partition slurm.Partition, stats *partitionStats) {
	// Extract name from pointer
	name := ""
	if partition.Name != nil {
		name = *partition.Name
	}

	// Extract state from nested Partition.State array
	stateStr := "UNKNOWN"
	if partition.Partition != nil && len(partition.Partition.State) > 0 {
		stateStr = string(partition.Partition.State[0])
	}

	stateValue := 0.0
	if isPartitionUp(stateStr) {
		stateValue = 1.0
	}
	ch <- prometheus.MustNewConstMetric(c.partitionState, prometheus.GaugeValue, stateValue, name, stateStr)

	// Extract node total from nested Nodes.Total
	nodesTot := int32(0)
	if partition.Nodes != nil && partition.Nodes.Total != nil {
		nodesTot = *partition.Nodes.Total
	}
	ch <- prometheus.MustNewConstMetric(c.partitionNodesTotal, prometheus.GaugeValue, float64(nodesTot), name)

	allocatedNodes := nodesTot
	if nodesTot > 0 && stats != nil {
		allocatedNodes = nodesTot - int32(stats.idleNodes) - int32(stats.downNodes)
		if allocatedNodes < 0 {
			allocatedNodes = 0
		}
	}
	ch <- prometheus.MustNewConstMetric(c.partitionNodesAllocated, prometheus.GaugeValue, float64(allocatedNodes), name)

	idleNodes := 0
	downNodes := 0
	if stats != nil {
		idleNodes = stats.idleNodes
		downNodes = stats.downNodes
	}
	ch <- prometheus.MustNewConstMetric(c.partitionNodesIdle, prometheus.GaugeValue, float64(idleNodes), name)
	ch <- prometheus.MustNewConstMetric(c.partitionNodesDown, prometheus.GaugeValue, float64(downNodes), name)

	// Extract CPU total from nested CPUs.Total
	cpusTot := int32(0)
	if partition.CPUs != nil && partition.CPUs.Total != nil {
		cpusTot = *partition.CPUs.Total
	}
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsTotal, prometheus.GaugeValue, float64(cpusTot), name)

	allocatedCPUs := int32(0)
	if stats != nil {
		allocatedCPUs = int32(stats.allocatedCPUs)
	}
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsAllocated, prometheus.GaugeValue, float64(allocatedCPUs), name)

	idleCPUs := cpusTot - allocatedCPUs
	if idleCPUs < 0 {
		idleCPUs = 0
	}
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsIdle, prometheus.GaugeValue, float64(idleCPUs), name)

	pendingJobs := 0
	runningJobs := 0
	if stats != nil {
		pendingJobs = stats.pendingJobs
		runningJobs = stats.runningJobs
	}
	ch <- prometheus.MustNewConstMetric(c.partitionJobsPending, prometheus.GaugeValue, float64(pendingJobs), name)
	ch <- prometheus.MustNewConstMetric(c.partitionJobsRunning, prometheus.GaugeValue, float64(runningJobs), name)

	// Extract time limits from partition data
	maxTime := formatTimeLimit(partition.Maximums)
	defaultTime := formatDefaultTime(partition.Defaults)
	ch <- prometheus.MustNewConstMetric(c.partitionInfo, prometheus.GaugeValue, 1, name, stateStr, "default", maxTime, defaultTime)
}

// collect gathers metrics from SLURM
func (c *PartitionsSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Get Partitions manager from client
	partitionsManager := c.client.Partitions()
	if partitionsManager == nil {
		return fmt.Errorf("partitions manager not available")
	}

	// List all partitions
	partitionList, err := partitionsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list partitions")
		return err
	}

	c.logger.WithField("count", len(partitionList.Partitions)).Info("Collected partition entries")

	// Query all nodes to build partition-level aggregations
	nodesManager := c.client.Nodes()
	var nodeList *slurm.NodeList
	if nodesManager != nil {
		nodeList, err = nodesManager.List(ctx, nil)
		if err != nil {
			c.logger.WithError(err).Warn("Failed to list nodes, node metrics will be unavailable")
			nodeList = nil
		}
	}

	// Query all jobs to build partition-level aggregations
	jobsManager := c.client.Jobs()
	var jobList *slurm.JobList
	if jobsManager != nil {
		jobList, err = jobsManager.List(ctx, nil)
		if err != nil {
			c.logger.WithError(err).Warn("Failed to list jobs, job metrics will be unavailable")
			jobList = nil
		}
	}

	// Build partition statistics from node and job data
	partitionStatsMap := buildPartitionStats(nodeList, jobList)

	for _, partition := range partitionList.Partitions {
		partitionName := ""
		if partition.Name != nil {
			partitionName = *partition.Name
		}
		stats := partitionStatsMap[partitionName]
		c.publishPartitionMetrics(ch, partition, stats)
	}

	return nil
}

// isPartitionUp returns true if the partition is in an up state
func isPartitionUp(state string) bool {
	state = strings.ToUpper(state)
	return state == "UP"
}

// buildPartitionStats aggregates node and job data by partition
func buildPartitionStats(nodeList *slurm.NodeList, jobList *slurm.JobList) map[string]*partitionStats {
	statsMap := make(map[string]*partitionStats)

	// Aggregate node data by partition
	if nodeList != nil {
		for _, node := range nodeList.Nodes {
			// Each node can belong to multiple partitions
			for _, partitionName := range node.Partitions {
				if statsMap[partitionName] == nil {
					statsMap[partitionName] = &partitionStats{}
				}
				stats := statsMap[partitionName]

				// Count nodes by state
				if len(node.State) > 0 {
					nodeState := string(node.State[0])
					switch nodeState {
					case "IDLE":
						stats.idleNodes++
					case "DOWN":
						stats.downNodes++
					}
				}

				// Sum allocated CPUs for this partition
				if node.AllocCPUs != nil {
					stats.allocatedCPUs += int(*node.AllocCPUs)
				}
			}
		}
	}

	// Aggregate job data by partition
	if jobList != nil {
		for _, job := range jobList.Jobs {
			if job.Partition == nil {
				continue
			}
			partitionName := *job.Partition

			if statsMap[partitionName] == nil {
				statsMap[partitionName] = &partitionStats{}
			}
			stats := statsMap[partitionName]

			// Count jobs by state
			if len(job.JobState) > 0 {
				jobState := string(job.JobState[0])
				switch jobState {
				case "PENDING":
					stats.pendingJobs++
				case "RUNNING":
					stats.runningJobs++
				}
			}
		}
	}

	return statsMap
}

// formatTimeLimit formats the maximum time limit from partition maximums
func formatTimeLimit(maximums *slurm.PartitionMaximums) string {
	if maximums == nil || maximums.Time == nil {
		return "UNLIMITED"
	}
	minutes := *maximums.Time
	if minutes == 0xFFFFFFFF { // Special value for unlimited
		return "UNLIMITED"
	}
	return fmt.Sprintf("%d", minutes)
}

// formatDefaultTime formats the default time limit from partition defaults
func formatDefaultTime(defaults *slurm.PartitionDefaults) string {
	if defaults == nil || defaults.Time == nil {
		return "UNLIMITED"
	}
	minutes := *defaults.Time
	if minutes == 0xFFFFFFFF { // Special value for unlimited
		return "UNLIMITED"
	}
	return fmt.Sprintf("%d", minutes)
}
