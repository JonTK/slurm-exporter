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
	return c.collect(ctx, ch)
}

// collect gathers metrics from SLURM
func (c *NodesSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
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

		// Extract node properties safely
		nodeName := "unknown"
		if node.Name != nil {
			nodeName = *node.Name
		}
		nodeStateStr := "UNKNOWN"
		if len(node.State) > 0 {
			nodeStateStr = string(node.State[0])
		}
		nodeCPUs := int32(0)
		if node.CPUs != nil {
			nodeCPUs = *node.CPUs
		}
		nodeMemory := int64(0)
		if node.RealMemory != nil {
			nodeMemory = *node.RealMemory
		}

		// Node state metric
		stateValue := 0.0
		if isNodeUp(nodeStateStr) {
			stateValue = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			c.nodeState,
			prometheus.GaugeValue,
			stateValue,
			nodeName, nodeStateStr, partition,
		)

		// CPU metrics
		ch <- prometheus.MustNewConstMetric(
			c.nodeCPUsTotal,
			prometheus.GaugeValue,
			float64(nodeCPUs),
			nodeName, partition,
		)

		// Get actual allocated CPUs from node data (API provides this)
		allocCPUs := int32(0)
		if node.AllocCPUs != nil {
			allocCPUs = *node.AllocCPUs
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeCPUsAllocated,
			prometheus.GaugeValue,
			float64(allocCPUs),
			nodeName, partition,
		)

		// Memory metrics (convert MB to bytes if Memory exists)
		memoryTotalBytes := float64(0)
		if nodeMemory > 0 && nodeMemory < 1000000 { // Sanity check: < 1TB
			memoryTotalBytes = float64(nodeMemory * 1024 * 1024)
		} else if nodeMemory >= 1000000 {
			c.logger.WithFields(map[string]interface{}{
				"node":      nodeName,
				"memory_mb": nodeMemory,
			}).Warn("Unusually high memory value detected, using 0")
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeMemoryTotal,
			prometheus.GaugeValue,
			memoryTotalBytes,
			nodeName, partition,
		)

		// Get actual allocated memory from node data (API provides this in MB)
		memoryAllocBytes := float64(0)
		if node.AllocMemory != nil {
			allocMemoryMB := *node.AllocMemory
			if allocMemoryMB > 0 && allocMemoryMB < 1000000 { // Sanity check: < 1TB
				memoryAllocBytes = float64(allocMemoryMB * 1024 * 1024)
			}
		}
		ch <- prometheus.MustNewConstMetric(
			c.nodeMemoryAllocated,
			prometheus.GaugeValue,
			memoryAllocBytes,
			nodeName, partition,
		)

		// Node info
		reason := ""
		if node.Reason != nil && *node.Reason != "" {
			reason = *node.Reason
		}

		// Get actual architecture and OS from node data
		arch := "x86_64" // default
		if node.Architecture != nil && *node.Architecture != "" {
			arch = *node.Architecture
		}
		os := "linux" // default
		if node.OperatingSystem != nil && *node.OperatingSystem != "" {
			os = *node.OperatingSystem
		}

		ch <- prometheus.MustNewConstMetric(
			c.nodeInfo,
			prometheus.GaugeValue,
			1,
			nodeName, partition, nodeStateStr, reason, arch, os,
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
