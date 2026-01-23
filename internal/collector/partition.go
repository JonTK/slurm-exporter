// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	// Commented out as only used in commented-out parser functions
	// "strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// PartitionCollector collects partition-level metrics
type PartitionCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewPartitionCollector creates a new partition collector
func NewPartitionCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *PartitionCollector {
	base := NewBaseCollector("partition", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &PartitionCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("partition", base.logger),
	}
}

// Describe implements the Collector interface
func (pc *PartitionCollector) Describe(ch chan<- *prometheus.Desc) {
	// Partition metrics
	pc.metrics.PartitionInfo.Describe(ch)
	pc.metrics.PartitionNodes.Describe(ch)
	pc.metrics.PartitionCPUs.Describe(ch)
	pc.metrics.PartitionMemory.Describe(ch)
	pc.metrics.PartitionState.Describe(ch)
	pc.metrics.PartitionJobCount.Describe(ch)
	pc.metrics.PartitionAllocatedCPUs.Describe(ch)
	pc.metrics.PartitionAllocatedMemory.Describe(ch)
	pc.metrics.PartitionUtilization.Describe(ch)
	pc.metrics.PartitionMaxTime.Describe(ch)
	pc.metrics.PartitionDefaultTime.Describe(ch)
	pc.metrics.PartitionPriority.Describe(ch)
}

// Collect implements the Collector interface
func (pc *PartitionCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return pc.CollectWithMetrics(ctx, ch, pc.collectPartitionMetrics)
}

// collectPartitionMetrics performs the actual partition metrics collection
func (pc *PartitionCollector) collectPartitionMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	pc.logger.Debug("Starting partition metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch partition data

	// Collect partition information and configuration
	if err := pc.collectPartitionInfo(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/partitions", 500, "",
			"Check SLURM partition data availability",
			"Verify partition information is accessible")
	}

	// Collect partition resource allocation and utilization
	if err := pc.collectPartitionUtilization(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/partitions", 500, "",
			"Check SLURM partition utilization data",
			"Verify partition resource information is accessible")
	}

	// Collect partition policies and limits
	if err := pc.collectPartitionPolicies(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/partitions", 500, "",
			"Check SLURM partition policy data",
			"Verify partition configuration is accessible")
	}

	pc.logger.Debug("Completed partition metrics collection")
	return nil
}

// collectPartitionInfo collects basic partition information and configuration
func (pc *PartitionCollector) collectPartitionInfo(ctx context.Context, ch chan<- prometheus.Metric) error {
	partitions := getTestPartitions()

	for _, partition := range partitions {
		pc.sendPartitionInfoMetric(ch, partition)
		pc.sendPartitionResourceMetrics(ch, partition)
		pc.sendPartitionStateMetrics(ch, partition)
	}

	pc.LogCollectionf("Collected info for %d partitions", len(partitions))
	return nil
}

// partitionInfo holds partition information for metric collection
type partitionInfo struct {
	Name        string
	State       string
	TotalNodes  int
	TotalCPUs   int
	TotalMemory int64 // bytes
	AllowGroups string
	AllowUsers  string
	Default     bool
	Hidden      bool
	RootOnly    bool
}

// sendPartitionInfoMetric sends the partition info metric with all labels
func (pc *PartitionCollector) sendPartitionInfoMetric(ch chan<- prometheus.Metric, partition partitionInfo) {
	pc.SendMetric(ch, pc.BuildMetric(
		pc.metrics.PartitionInfo.WithLabelValues(
			pc.clusterName,
			partition.Name,
			partition.State,
			partition.AllowGroups,
			partition.AllowUsers,
		).Desc(),
		prometheus.GaugeValue,
		1,
		pc.clusterName,
		partition.Name,
		partition.State,
		partition.AllowGroups,
		partition.AllowUsers,
	))
}

// sendPartitionResourceMetrics sends node, CPU, and memory metrics
func (pc *PartitionCollector) sendPartitionResourceMetrics(ch chan<- prometheus.Metric, partition partitionInfo) {
	// Partition nodes
	pc.SendMetric(ch, pc.BuildMetric(
		pc.metrics.PartitionNodes.WithLabelValues(pc.clusterName, partition.Name).Desc(),
		prometheus.GaugeValue,
		float64(partition.TotalNodes),
		pc.clusterName,
		partition.Name,
	))

	// Partition CPUs
	pc.SendMetric(ch, pc.BuildMetric(
		pc.metrics.PartitionCPUs.WithLabelValues(pc.clusterName, partition.Name).Desc(),
		prometheus.GaugeValue,
		float64(partition.TotalCPUs),
		pc.clusterName,
		partition.Name,
	))

	// Partition memory
	pc.SendMetric(ch, pc.BuildMetric(
		pc.metrics.PartitionMemory.WithLabelValues(pc.clusterName, partition.Name).Desc(),
		prometheus.GaugeValue,
		float64(partition.TotalMemory),
		pc.clusterName,
		partition.Name,
	))
}

// sendPartitionStateMetrics sends binary state metrics for all possible states
func (pc *PartitionCollector) sendPartitionStateMetrics(ch chan<- prometheus.Metric, partition partitionInfo) {
	states := []string{"UP", "DOWN", "DRAIN", "INACTIVE"}
	for _, state := range states {
		value := float64(0)
		if partition.State == state {
			value = 1
		}

		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionState.WithLabelValues(pc.clusterName, partition.Name, state).Desc(),
			prometheus.GaugeValue,
			value,
			pc.clusterName,
			partition.Name,
			state,
		))
	}
}

// getTestPartitions returns test data for partition information
// In real implementation this would come from SLURM API
func getTestPartitions() []partitionInfo {
	return []partitionInfo{
		{
			Name:        "compute",
			State:       "UP",
			TotalNodes:  50,
			TotalCPUs:   2400,
			TotalMemory: 512 * 1024 * 1024 * 1024, // 512GB
			AllowGroups: "all",
			AllowUsers:  "all",
			Default:     true,
			Hidden:      false,
			RootOnly:    false,
		},
		{
			Name:        "gpu",
			State:       "UP",
			TotalNodes:  10,
			TotalCPUs:   480,
			TotalMemory: 256 * 1024 * 1024 * 1024, // 256GB
			AllowGroups: "gpu_users",
			AllowUsers:  "all",
			Default:     false,
			Hidden:      false,
			RootOnly:    false,
		},
		{
			Name:        "highmem",
			State:       "UP",
			TotalNodes:  5,
			TotalCPUs:   240,
			TotalMemory: 512 * 1024 * 1024 * 1024, // 512GB
			AllowGroups: "highmem_users",
			AllowUsers:  "all",
			Default:     false,
			Hidden:      false,
			RootOnly:    false,
		},
		{
			Name:        "debug",
			State:       "UP",
			TotalNodes:  2,
			TotalCPUs:   96,
			TotalMemory: 64 * 1024 * 1024 * 1024, // 64GB
			AllowGroups: "all",
			AllowUsers:  "all",
			Default:     false,
			Hidden:      false,
			RootOnly:    false,
		},
		{
			Name:        "maintenance",
			State:       "DOWN",
			TotalNodes:  3,
			TotalCPUs:   144,
			TotalMemory: 96 * 1024 * 1024 * 1024, // 96GB
			AllowGroups: "admin",
			AllowUsers:  "root",
			Default:     false,
			Hidden:      true,
			RootOnly:    true,
		},
	}
}

// collectPartitionUtilization collects partition resource utilization
func (pc *PartitionCollector) collectPartitionUtilization(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate partition utilization data
	partitionUtilization := []struct {
		Name            string
		JobCount        int
		AllocatedNodes  int
		AllocatedCPUs   int
		AllocatedMemory int64 // bytes
		TotalNodes      int
		TotalCPUs       int
		TotalMemory     int64 // bytes
	}{
		{
			Name:            "compute",
			JobCount:        120,
			AllocatedNodes:  35,
			AllocatedCPUs:   1680,
			AllocatedMemory: 336 * 1024 * 1024 * 1024, // 336GB
			TotalNodes:      50,
			TotalCPUs:       2400,
			TotalMemory:     512 * 1024 * 1024 * 1024, // 512GB
		},
		{
			Name:            "gpu",
			JobCount:        25,
			AllocatedNodes:  8,
			AllocatedCPUs:   384,
			AllocatedMemory: 192 * 1024 * 1024 * 1024, // 192GB
			TotalNodes:      10,
			TotalCPUs:       480,
			TotalMemory:     256 * 1024 * 1024 * 1024, // 256GB
		},
		{
			Name:            "highmem",
			JobCount:        8,
			AllocatedNodes:  3,
			AllocatedCPUs:   144,
			AllocatedMemory: 384 * 1024 * 1024 * 1024, // 384GB
			TotalNodes:      5,
			TotalCPUs:       240,
			TotalMemory:     512 * 1024 * 1024 * 1024, // 512GB
		},
		{
			Name:            "debug",
			JobCount:        3,
			AllocatedNodes:  1,
			AllocatedCPUs:   48,
			AllocatedMemory: 32 * 1024 * 1024 * 1024, // 32GB
			TotalNodes:      2,
			TotalCPUs:       96,
			TotalMemory:     64 * 1024 * 1024 * 1024, // 64GB
		},
		{
			Name:            "maintenance",
			JobCount:        0,
			AllocatedNodes:  0,
			AllocatedCPUs:   0,
			AllocatedMemory: 0,
			TotalNodes:      3,
			TotalCPUs:       144,
			TotalMemory:     96 * 1024 * 1024 * 1024, // 96GB
		},
	}

	for _, util := range partitionUtilization {
		// Job count
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionJobCount.WithLabelValues(pc.clusterName, util.Name).Desc(),
			prometheus.GaugeValue,
			float64(util.JobCount),
			pc.clusterName,
			util.Name,
		))

		// Allocated resources
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionAllocatedCPUs.WithLabelValues(pc.clusterName, util.Name).Desc(),
			prometheus.GaugeValue,
			float64(util.AllocatedCPUs),
			pc.clusterName,
			util.Name,
		))

		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionAllocatedMemory.WithLabelValues(pc.clusterName, util.Name).Desc(),
			prometheus.GaugeValue,
			float64(util.AllocatedMemory),
			pc.clusterName,
			util.Name,
		))

		// Utilization ratios
		cpuUtilization := float64(0)
		memoryUtilization := float64(0)
		nodeUtilization := float64(0)

		if util.TotalCPUs > 0 {
			cpuUtilization = float64(util.AllocatedCPUs) / float64(util.TotalCPUs)
		}
		if util.TotalMemory > 0 {
			memoryUtilization = float64(util.AllocatedMemory) / float64(util.TotalMemory)
		}
		if util.TotalNodes > 0 {
			nodeUtilization = float64(util.AllocatedNodes) / float64(util.TotalNodes)
		}

		utilizationTypes := map[string]float64{
			"cpu":    cpuUtilization,
			"memory": memoryUtilization,
			"nodes":  nodeUtilization,
		}

		for utilizationType, utilization := range utilizationTypes {
			pc.SendMetric(ch, pc.BuildMetric(
				pc.metrics.PartitionUtilization.WithLabelValues(pc.clusterName, util.Name, utilizationType).Desc(),
				prometheus.GaugeValue,
				utilization,
				pc.clusterName,
				util.Name,
				utilizationType,
			))
		}
	}

	pc.LogCollectionf("Collected utilization for %d partitions", len(partitionUtilization))
	return nil
}

// collectPartitionPolicies collects partition policies and limits
func (pc *PartitionCollector) collectPartitionPolicies(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate partition policy data
	partitionPolicies := []struct {
		Name              string
		MaxTime           int64 // minutes
		DefaultTime       int64 // minutes
		Priority          int
		PreemptMode       string
		OversubscribeMode string
		MaxNodesPerJob    int
		MaxCPUsPerJob     int
		MaxMemoryPerJob   int64 // bytes
	}{
		{
			Name:              "compute",
			MaxTime:           2880, // 48 hours
			DefaultTime:       480,  // 8 hours
			Priority:          1,
			PreemptMode:       "OFF",
			OversubscribeMode: "NO",
			MaxNodesPerJob:    16,
			MaxCPUsPerJob:     768,
			MaxMemoryPerJob:   1024 * 1024 * 1024 * 1024, // 1TB
		},
		{
			Name:              "gpu",
			MaxTime:           1440, // 24 hours
			DefaultTime:       240,  // 4 hours
			Priority:          10,
			PreemptMode:       "REQUEUE",
			OversubscribeMode: "NO",
			MaxNodesPerJob:    4,
			MaxCPUsPerJob:     192,
			MaxMemoryPerJob:   512 * 1024 * 1024 * 1024, // 512GB
		},
		{
			Name:              "highmem",
			MaxTime:           720, // 12 hours
			DefaultTime:       120, // 2 hours
			Priority:          5,
			PreemptMode:       "OFF",
			OversubscribeMode: "NO",
			MaxNodesPerJob:    2,
			MaxCPUsPerJob:     96,
			MaxMemoryPerJob:   512 * 1024 * 1024 * 1024, // 512GB
		},
		{
			Name:              "debug",
			MaxTime:           60, // 1 hour
			DefaultTime:       15, // 15 minutes
			Priority:          100,
			PreemptMode:       "REQUEUE",
			OversubscribeMode: "YES",
			MaxNodesPerJob:    1,
			MaxCPUsPerJob:     48,
			MaxMemoryPerJob:   32 * 1024 * 1024 * 1024, // 32GB
		},
		{
			Name:              "maintenance",
			MaxTime:           10080, // 7 days
			DefaultTime:       480,   // 8 hours
			Priority:          1000,
			PreemptMode:       "OFF",
			OversubscribeMode: "NO",
			MaxNodesPerJob:    3,
			MaxCPUsPerJob:     144,
			MaxMemoryPerJob:   96 * 1024 * 1024 * 1024, // 96GB
		},
	}

	for _, policy := range partitionPolicies {
		// Max time limit (in seconds for Prometheus)
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionMaxTime.WithLabelValues(pc.clusterName, policy.Name).Desc(),
			prometheus.GaugeValue,
			float64(policy.MaxTime*60), // convert minutes to seconds
			pc.clusterName,
			policy.Name,
		))

		// Default time limit (in seconds for Prometheus)
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionDefaultTime.WithLabelValues(pc.clusterName, policy.Name).Desc(),
			prometheus.GaugeValue,
			float64(policy.DefaultTime*60), // convert minutes to seconds
			pc.clusterName,
			policy.Name,
		))

		// Priority
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.PartitionPriority.WithLabelValues(pc.clusterName, policy.Name).Desc(),
			prometheus.GaugeValue,
			float64(policy.Priority),
			pc.clusterName,
			policy.Name,
		))
	}

	pc.LogCollectionf("Collected policies for %d partitions", len(partitionPolicies))
	return nil
}

// PartitionInfo represents partition information structure
type PartitionInfo struct {
	Name        string
	State       string
	Nodes       []string
	TotalNodes  int
	TotalCPUs   int
	TotalMemory int64
	AllowGroups []string
	AllowUsers  []string
	DenyGroups  []string
	DenyUsers   []string
	Default     bool
	Hidden      bool
	RootOnly    bool
	Shared      bool
	Preempt     bool
	Priority    int
	Limits      PartitionLimits
	Policies    PartitionPolicies
}

// PartitionLimits represents resource limits for a partition
type PartitionLimits struct {
	MaxTime         time.Duration
	DefaultTime     time.Duration
	MaxNodes        int
	MaxNodesPerJob  int
	MaxCPUs         int
	MaxCPUsPerJob   int
	MaxMemory       int64
	MaxMemoryPerJob int64
	MaxJobsPerUser  int
	MaxJobsTotal    int
}

// PartitionPolicies represents policies for a partition
type PartitionPolicies struct {
	PreemptMode       string
	OversubscribeMode string
	Priority          int
	GraceTime         time.Duration
	PreemptType       string
	SelectType        string
	DefMemPerCPU      int64
	MaxMemPerCPU      int64
}

// TODO: Following parser functions are unused - preserved for future data parsing needs
/*
// parsePartitionState converts SLURM partition state string to normalized state
func (pc *PartitionCollector) parsePartitionState(slurmState string) string {
	// SLURM partition states: UP, DOWN, DRAIN, INACTIVE
	switch slurmState {
	case "UP":
		return "UP"
	case "DOWN":
		return "DOWN"
	case "DRAIN":
		return "DRAIN"
	case "INACTIVE":
		return "INACTIVE"
	default:
		return "UNKNOWN"
	}
}

// parseTimeLimit converts SLURM time limit string to duration
func (pc *PartitionCollector) parseTimeLimit(slurmTime string) (time.Duration, error) {
	if slurmTime == "" || slurmTime == "UNLIMITED" {
		return 0, nil
	}

	// SLURM time formats: "minutes", "hours:minutes", "days-hours:minutes:seconds"
	// For now, simple implementation assuming minutes
	if minutes, err := strconv.Atoi(slurmTime); err == nil {
		return time.Duration(minutes) * time.Minute, nil
	}

	return 0, pc.errorBuilder.Parsing(nil, "parse_time_limit", "time_limit",
		"Check SLURM time limit format",
		"Verify time limit configuration")
}

// parseMemorySize converts SLURM memory specification to bytes
func (pc *PartitionCollector) parseMemorySize(slurmMemory string) (int64, error) {
	// SLURM memory can be in various formats: "1024M", "2G", "500000K", etc.
	if slurmMemory == "" || slurmMemory == "UNLIMITED" {
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
