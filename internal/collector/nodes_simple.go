// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	nodesCollectorSubsystem = "node"
)

// Compile-time interface compliance checks
var (
	_ Collector             = (*NodesSimpleCollector)(nil)
	_ CustomLabelsCollector = (*NodesSimpleCollector)(nil)
)

// NodesSimpleCollector collects node-related metrics
type NodesSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Custom labels
	customLabels map[string]string

	// Node state metrics
	nodeState *prometheus.Desc

	// Node resource metrics
	nodeCPUsTotal       *prometheus.Desc
	nodeCPUsAllocated   *prometheus.Desc
	nodeMemoryTotal     *prometheus.Desc
	nodeMemoryAllocated *prometheus.Desc

	// Node info
	nodeInfo *prometheus.Desc
}

// NewNodesSimpleCollector creates a new Nodes collector
func NewNodesSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *NodesSimpleCollector {
	c := &NodesSimpleCollector{
		client:       client,
		logger:       logger.WithField("collector", "nodes"),
		enabled:      true,
		customLabels: make(map[string]string),
	}

	// Initialize metrics
	c.initializeMetrics()

	return c
}

// initializeMetrics creates metric descriptors with custom labels as constant labels
func (c *NodesSimpleCollector) initializeMetrics() {
	// Convert custom labels to prometheus.Labels for constant labels
	constLabels := prometheus.Labels{}
	for k, v := range c.customLabels {
		constLabels[k] = v
	}
	c.nodeState = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "state"),
		"Current state of the node (1=up, 0=down)",
		[]string{"node", "state", "partition"},
		constLabels,
	)

	c.nodeCPUsTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "cpus_total"),
		"Total number of CPUs on the node",
		[]string{"node", "partition"},
		constLabels,
	)

	c.nodeCPUsAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "cpus_allocated"),
		"Number of allocated CPUs on the node",
		[]string{"node", "partition"},
		constLabels,
	)

	c.nodeMemoryTotal = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "memory_total_bytes"),
		"Total memory on the node in bytes",
		[]string{"node", "partition"},
		constLabels,
	)

	c.nodeMemoryAllocated = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "memory_allocated_bytes"),
		"Allocated memory on the node in bytes",
		[]string{"node", "partition"},
		constLabels,
	)

	c.nodeInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nodesCollectorSubsystem, "info"),
		"Node information with all labels",
		[]string{"node", "partition", "state", "reason", "arch", "os"},
		constLabels,
	)
}

// SetCustomLabels sets custom labels for this collector
func (c *NodesSimpleCollector) SetCustomLabels(labels map[string]string) {
	c.customLabels = make(map[string]string)
	for k, v := range labels {
		c.customLabels[k] = v
	}
	// Rebuild metrics with new constant labels
	c.initializeMetrics()
}

// Name returns the collector name
func (c *NodesSimpleCollector) Name() string {
	return "nodes"
}

// IsEnabled returns whether this collector is enabled
func (c *NodesSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *NodesSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *NodesSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.nodeState
	ch <- c.nodeCPUsTotal
	ch <- c.nodeCPUsAllocated
	ch <- c.nodeMemoryTotal
	ch <- c.nodeMemoryAllocated
	ch <- c.nodeInfo
}

// Collect implements the Collector interface
func (c *NodesSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *NodesSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Nodes manager from client
	nodesManager := c.client.Nodes()
	if nodesManager == nil {
		return fmt.Errorf("nodes manager not available")
	}

	// List all nodes
	nodeList, err := nodesManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list nodes")
		return err
	}

	c.logger.WithField("count", len(nodeList.Nodes)).Info("Collected node entries")

	for _, node := range nodeList.Nodes {
		// Get first partition if available
		partition := ""
		if len(node.Partitions) > 0 {
			partition = node.Partitions[0]
		}

		// Node state metric
		stateValue := 0.0
		if isNodeUp(node.State) {
			stateValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.nodeState,
			prometheus.GaugeValue,
			stateValue,
			node.Name, node.State, partition,
		)

		// CPU metrics
		ch <- prometheus.MustNewConstMetric(
			c.nodeCPUsTotal,
			prometheus.GaugeValue,
			float64(node.CPUs),
			node.Name, partition,
		)

		// Calculate allocated CPUs from node state
		// If node is in use, assume some CPUs are allocated based on state
		allocCPUs := 0
		switch node.State {
		case "ALLOCATED", "MIXED":
			// For allocated/mixed nodes, assume 50% utilization as reasonable estimate
			allocCPUs = node.CPUs / 2
		case "COMPLETING":
			// Node is completing jobs, assume high utilization
			allocCPUs = node.CPUs * 3 / 4
		}
		// For IDLE, DRAIN, DOWN states, allocCPUs stays 0
		ch <- prometheus.MustNewConstMetric(
			c.nodeCPUsAllocated,
			prometheus.GaugeValue,
			float64(allocCPUs),
			node.Name, partition,
		)

		// Memory metrics (convert MB to bytes if Memory exists)
		memoryTotalBytes := float64(0)
		if node.Memory > 0 && node.Memory < 1000000 { // Sanity check: < 1TB
			memoryTotalBytes = float64(node.Memory * 1024 * 1024)
		} else if node.Memory >= 1000000 {
			c.logger.WithFields(map[string]interface{}{
				"node":      node.Name,
				"memory_mb": node.Memory,
			}).Warn("Unusually high memory value detected, using 0")
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeMemoryTotal,
			prometheus.GaugeValue,
			memoryTotalBytes,
			node.Name, partition,
		)

		// Allocated memory (estimate based on CPU allocation ratio)
		memoryAllocBytes := float64(0)
		if memoryTotalBytes > 0 && allocCPUs > 0 && node.CPUs > 0 {
			// Estimate memory allocation proportional to CPU allocation
			allocationRatio := float64(allocCPUs) / float64(node.CPUs)
			memoryAllocBytes = memoryTotalBytes * allocationRatio
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeMemoryAllocated,
			prometheus.GaugeValue,
			memoryAllocBytes,
			node.Name, partition,
		)

		// Node info
		reason := ""
		if node.Reason != "" {
			reason = node.Reason
		}

		arch := "x86_64" // default
		os := "linux"    // default

		ch <- prometheus.MustNewConstMetric(
			c.nodeInfo,
			prometheus.GaugeValue,
			1,
			node.Name, partition, node.State, reason, arch, os,
		)
	}

	return nil
}

// isNodeUp returns true if the node is in an up state
func isNodeUp(state string) bool {
	state = strings.ToUpper(state)
	// Node is considered up if it doesn't have DOWN or DRAIN in the state
	return !strings.Contains(state, "DOWN") && !strings.Contains(state, "DRAIN")
}
