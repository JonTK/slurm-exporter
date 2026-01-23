// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	// Commented out as only used in commented-out parser function
	// "strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// NodeCollector collects node-level metrics
type NodeCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewNodeCollector creates a new node collector
func NewNodeCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *NodeCollector {
	base := NewBaseCollector("node", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &NodeCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("node", base.logger),
	}
}

// Describe implements the Collector interface
func (nc *NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	// Node information metrics
	nc.metrics.NodeInfo.Describe(ch)
	nc.metrics.NodeCPUs.Describe(ch)
	nc.metrics.NodeMemory.Describe(ch)
	nc.metrics.NodeState.Describe(ch)
	nc.metrics.NodeAllocatedCPUs.Describe(ch)
	nc.metrics.NodeAllocatedMemory.Describe(ch)
	nc.metrics.NodeLoad.Describe(ch)
	nc.metrics.NodeUptime.Describe(ch)
	nc.metrics.NodeLastHeartbeat.Describe(ch)
	nc.metrics.NodeFeatures.Describe(ch)
	nc.metrics.NodeGPUs.Describe(ch)
	nc.metrics.NodeWeight.Describe(ch)
}

// Collect implements the Collector interface
func (nc *NodeCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return nc.CollectWithMetrics(ctx, ch, nc.collectNodeMetrics)
}

// collectNodeMetrics performs the actual node metrics collection
func (nc *NodeCollector) collectNodeMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	nc.logger.Debug("Starting node metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch node data

	// Collect individual node information
	if err := nc.collectNodeInfo(ctx, ch); err != nil {
		return nc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM node data availability",
			"Verify node information is up to date")
	}

	// Collect node states
	if err := nc.collectNodeStates(ctx, ch); err != nil {
		return nc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM node state data",
			"Verify node state information is accessible")
	}

	// Collect node resource utilization
	if err := nc.collectNodeUtilization(ctx, ch); err != nil {
		return nc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM node utilization data",
			"Verify resource utilization metrics are available")
	}

	// Collect node health metrics
	if err := nc.collectNodeHealth(ctx, ch); err != nil {
		return nc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM node health data",
			"Verify node health metrics are accessible")
	}

	nc.logger.Debug("Completed node metrics collection")
	return nil
}

// collectNodeInfo collects basic node information
func (nc *NodeCollector) collectNodeInfo(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate node data - in real implementation this would come from SLURM API
	// This represents what we might get from /slurm/v1/nodes
	nodes := []struct {
		Name         string
		Architecture string
		OS           string
		Kernel       string
		CPUs         int
		Memory       int64 // bytes
		Weight       int
	}{
		{
			Name:         "compute-001",
			Architecture: "x86_64",
			OS:           "Linux",
			Kernel:       "5.4.0-74-generic",
			CPUs:         48,
			Memory:       128 * 1024 * 1024 * 1024, // 128GB
			Weight:       1,
		},
		{
			Name:         "compute-002",
			Architecture: "x86_64",
			OS:           "Linux",
			Kernel:       "5.4.0-74-generic",
			CPUs:         48,
			Memory:       128 * 1024 * 1024 * 1024, // 128GB
			Weight:       1,
		},
		{
			Name:         "gpu-001",
			Architecture: "x86_64",
			OS:           "Linux",
			Kernel:       "5.4.0-74-generic",
			CPUs:         24,
			Memory:       256 * 1024 * 1024 * 1024, // 256GB
			Weight:       2,
		},
	}

	for _, node := range nodes {
		// Node info metric
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeInfo.WithLabelValues(
				nc.clusterName,
				node.Name,
				node.Architecture,
				node.OS,
				node.Kernel,
			).Desc(),
			prometheus.GaugeValue,
			1,
			nc.clusterName,
			node.Name,
			node.Architecture,
			node.OS,
			node.Kernel,
		))

		// Node CPUs
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeCPUs.WithLabelValues(nc.clusterName, node.Name).Desc(),
			prometheus.GaugeValue,
			float64(node.CPUs),
			nc.clusterName,
			node.Name,
		))

		// Node memory
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeMemory.WithLabelValues(nc.clusterName, node.Name).Desc(),
			prometheus.GaugeValue,
			float64(node.Memory),
			nc.clusterName,
			node.Name,
		))

		// Node weight
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeWeight.WithLabelValues(nc.clusterName, node.Name).Desc(),
			prometheus.GaugeValue,
			float64(node.Weight),
			nc.clusterName,
			node.Name,
		))
	}

	nc.LogCollectionf("Collected info for %d nodes", len(nodes))
	return nil
}

// collectNodeStates collects node state information
func (nc *NodeCollector) collectNodeStates(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate node state data
	nodeStates := []struct {
		Name  string
		State string
	}{
		{"compute-001", "idle"},
		{"compute-002", "allocated"},
		{"gpu-001", "mixed"},
	}

	for _, nodeState := range nodeStates {
		// Normalize the state
		normalizedState := nc.parseNodeState(nodeState.State)

		// Node state metric (1 for current state, 0 for others)
		states := []string{"idle", "allocated", "mixed", "down", "drain", "completing", "maintenance", "unknown"}
		for _, state := range states {
			value := float64(0)
			if state == normalizedState {
				value = 1
			}

			nc.SendMetric(ch, nc.BuildMetric(
				nc.metrics.NodeState.WithLabelValues(nc.clusterName, nodeState.Name, state).Desc(),
				prometheus.GaugeValue,
				value,
				nc.clusterName,
				nodeState.Name,
				state,
			))
		}
	}

	nc.LogCollectionf("Collected states for %d nodes", len(nodeStates))
	return nil
}

// collectNodeUtilization collects node resource utilization
func (nc *NodeCollector) collectNodeUtilization(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate node utilization data
	nodeUtilization := []struct {
		Name            string
		AllocatedCPUs   int
		AllocatedMemory int64 // bytes
	}{
		{"compute-001", 0, 0},                        // idle
		{"compute-002", 48, 64 * 1024 * 1024 * 1024}, // fully allocated
		{"gpu-001", 12, 128 * 1024 * 1024 * 1024},    // partially allocated
	}

	for _, util := range nodeUtilization {
		// Allocated CPUs
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeAllocatedCPUs.WithLabelValues(nc.clusterName, util.Name).Desc(),
			prometheus.GaugeValue,
			float64(util.AllocatedCPUs),
			nc.clusterName,
			util.Name,
		))

		// Allocated memory
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeAllocatedMemory.WithLabelValues(nc.clusterName, util.Name).Desc(),
			prometheus.GaugeValue,
			float64(util.AllocatedMemory),
			nc.clusterName,
			util.Name,
		))
	}

	nc.LogCollectionf("Collected utilization for %d nodes", len(nodeUtilization))
	return nil
}

// collectNodeHealth collects node health and performance metrics
func (nc *NodeCollector) collectNodeHealth(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate node health data
	nodeHealth := []struct {
		Name          string
		LoadAvg1      float64
		LoadAvg5      float64
		LoadAvg15     float64
		Uptime        int64 // seconds
		LastHeartbeat int64 // unix timestamp
		Features      []string
		GPUs          map[string]int // GPU type -> count
	}{
		{
			Name:          "compute-001",
			LoadAvg1:      0.1,
			LoadAvg5:      0.2,
			LoadAvg15:     0.1,
			Uptime:        86400 * 30, // 30 days
			LastHeartbeat: time.Now().Unix(),
			Features:      []string{"avx2", "sse4_2"},
			GPUs:          nil,
		},
		{
			Name:          "compute-002",
			LoadAvg1:      15.2,
			LoadAvg5:      14.8,
			LoadAvg15:     12.5,
			Uptime:        86400 * 25, // 25 days
			LastHeartbeat: time.Now().Unix(),
			Features:      []string{"avx2", "sse4_2"},
			GPUs:          nil,
		},
		{
			Name:          "gpu-001",
			LoadAvg1:      8.5,
			LoadAvg5:      9.1,
			LoadAvg15:     7.8,
			Uptime:        86400 * 20, // 20 days
			LastHeartbeat: time.Now().Unix(),
			Features:      []string{"avx2", "sse4_2", "cuda"},
			GPUs:          map[string]int{"V100": 4},
		},
	}

	for _, health := range nodeHealth {
		// Load averages
		loadIntervals := map[string]float64{
			"1m":  health.LoadAvg1,
			"5m":  health.LoadAvg5,
			"15m": health.LoadAvg15,
		}

		for interval, load := range loadIntervals {
			nc.SendMetric(ch, nc.BuildMetric(
				nc.metrics.NodeLoad.WithLabelValues(nc.clusterName, health.Name, interval).Desc(),
				prometheus.GaugeValue,
				load,
				nc.clusterName,
				health.Name,
				interval,
			))
		}

		// Uptime
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeUptime.WithLabelValues(nc.clusterName, health.Name).Desc(),
			prometheus.GaugeValue,
			float64(health.Uptime),
			nc.clusterName,
			health.Name,
		))

		// Last heartbeat
		nc.SendMetric(ch, nc.BuildMetric(
			nc.metrics.NodeLastHeartbeat.WithLabelValues(nc.clusterName, health.Name).Desc(),
			prometheus.GaugeValue,
			float64(health.LastHeartbeat),
			nc.clusterName,
			health.Name,
		))

		// Features
		for _, feature := range health.Features {
			nc.SendMetric(ch, nc.BuildMetric(
				nc.metrics.NodeFeatures.WithLabelValues(nc.clusterName, health.Name, feature).Desc(),
				prometheus.GaugeValue,
				1,
				nc.clusterName,
				health.Name,
				feature,
			))
		}

		// GPUs
		if health.GPUs != nil {
			for gpuType, count := range health.GPUs {
				nc.SendMetric(ch, nc.BuildMetric(
					nc.metrics.NodeGPUs.WithLabelValues(nc.clusterName, health.Name, gpuType).Desc(),
					prometheus.GaugeValue,
					float64(count),
					nc.clusterName,
					health.Name,
					gpuType,
				))
			}
		}
	}

	nc.LogCollectionf("Collected health metrics for %d nodes", len(nodeHealth))
	return nil
}

// parseNodeState converts SLURM node state string to normalized state
func (nc *NodeCollector) parseNodeState(slurmState string) string {
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

// TODO: parseMemorySize is unused - preserved for future data parsing needs
/*
// parseMemorySize converts SLURM memory specification to bytes
func (nc *NodeCollector) parseMemorySize(slurmMemory string) (int64, error) {
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
*/
