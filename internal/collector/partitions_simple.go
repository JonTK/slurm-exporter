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

// publishPartitionMetrics publishes all metrics for a single partition
func (c *PartitionsSimpleCollector) publishPartitionMetrics(ch chan<- prometheus.Metric, partition slurm.Partition) {
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
	if nodesTot > 0 {
		allocatedNodes = nodesTot - int32(getIdleNodes(partition)) - int32(getDownNodes(partition))
		if allocatedNodes < 0 {
			allocatedNodes = 0
		}
	}
	ch <- prometheus.MustNewConstMetric(c.partitionNodesAllocated, prometheus.GaugeValue, float64(allocatedNodes), name)
	ch <- prometheus.MustNewConstMetric(c.partitionNodesIdle, prometheus.GaugeValue, float64(getIdleNodes(partition)), name)
	ch <- prometheus.MustNewConstMetric(c.partitionNodesDown, prometheus.GaugeValue, float64(getDownNodes(partition)), name)

	// Extract CPU total from nested CPUs.Total
	cpusTot := int32(0)
	if partition.CPUs != nil && partition.CPUs.Total != nil {
		cpusTot = *partition.CPUs.Total
	}
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsTotal, prometheus.GaugeValue, float64(cpusTot), name)

	allocatedCPUs := int32(getPartitionAllocatedCPUs(partition))
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsAllocated, prometheus.GaugeValue, float64(allocatedCPUs), name)

	idleCPUs := cpusTot - allocatedCPUs
	if idleCPUs < 0 {
		idleCPUs = 0
	}
	ch <- prometheus.MustNewConstMetric(c.partitionCPUsIdle, prometheus.GaugeValue, float64(idleCPUs), name)

	ch <- prometheus.MustNewConstMetric(c.partitionJobsPending, prometheus.GaugeValue, float64(getPartitionPendingJobs(partition)), name)
	ch <- prometheus.MustNewConstMetric(c.partitionJobsRunning, prometheus.GaugeValue, float64(getPartitionRunningJobs(partition)), name)

	// These helper functions may need updating, for now use placeholder values
	maxTime := "UNLIMITED"
	defaultTime := "UNLIMITED"
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

	for _, partition := range partitionList.Partitions {
		c.publishPartitionMetrics(ch, partition)
	}

	return nil
}

// isPartitionUp returns true if the partition is in an up state
func isPartitionUp(state string) bool {
	state = strings.ToUpper(state)
	return state == "UP"
}

// Helper functions to extract metrics from partition data
// These may need adjustment based on actual API response structure

func getIdleNodes(p slurm.Partition) int {
	// This would depend on the actual partition struct fields
	// Some APIs provide this directly, others require calculation
	return 0 // Placeholder
}

func getDownNodes(p slurm.Partition) int {
	// This would depend on the actual partition struct fields
	return 0 // Placeholder
}

func getPartitionAllocatedCPUs(p slurm.Partition) int {
	// This would depend on the actual partition struct fields
	return 0 // Placeholder
}

func getPartitionPendingJobs(p slurm.Partition) int {
	// This would depend on the actual partition struct fields
	return 0 // Placeholder
}

func getPartitionRunningJobs(p slurm.Partition) int {
	// This would depend on the actual partition struct fields
	return 0 // Placeholder
}

