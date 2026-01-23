// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	// Commented out as only used in commented-out calculation function
	// "math"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// PerformanceCollector collects system-wide performance metrics
type PerformanceCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
}

// NewPerformanceCollector creates a new performance collector
func NewPerformanceCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *PerformanceCollector {
	base := NewBaseCollector("performance", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &PerformanceCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("performance", base.logger),
	}
}

// Describe implements the Collector interface
func (pc *PerformanceCollector) Describe(ch chan<- *prometheus.Desc) {
	// Performance metrics
	pc.metrics.SystemThroughput.Describe(ch)
	pc.metrics.SystemEfficiency.Describe(ch)
	pc.metrics.QueueWaitTime.Describe(ch)
	pc.metrics.QueueDepth.Describe(ch)
	pc.metrics.ResourceUtilization.Describe(ch)
	pc.metrics.JobTurnover.Describe(ch)
}

// Collect implements the Collector interface
func (pc *PerformanceCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return pc.CollectWithMetrics(ctx, ch, pc.collectPerformanceMetrics)
}

// collectPerformanceMetrics performs the actual performance metrics collection
func (pc *PerformanceCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	pc.logger.Debug("Starting performance metrics collection")

	// Since SLURM client is not available, we'll simulate the metrics for now
	// In real implementation, this would use the SLURM client to fetch performance data

	// Collect system throughput metrics
	if err := pc.collectThroughputMetrics(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM job throughput data availability",
			"Verify job completion and submission rates are accessible")
	}

	// Collect system efficiency metrics
	if err := pc.collectEfficiencyMetrics(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/nodes", 500, "",
			"Check SLURM resource efficiency data",
			"Verify resource utilization information is accessible")
	}

	// Collect queue analysis metrics
	if err := pc.collectQueueMetrics(ctx, ch); err != nil {
		return pc.errorBuilder.API(err, "/slurm/v1/jobs", 500, "",
			"Check SLURM queue data availability",
			"Verify job queue information is accessible")
	}

	pc.logger.Debug("Completed performance metrics collection")
	return nil
}

// collectThroughputMetrics collects system throughput metrics
func (pc *PerformanceCollector) collectThroughputMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate throughput data - in real implementation this would come from SLURM API
	// This represents jobs completed/submitted per hour, CPU hours delivered, etc.
	throughputData := []struct {
		MetricType    string
		Value         float64
		TimeWindow    string
		PartitionName string
	}{
		{
			MetricType:    "jobs_completed",
			Value:         45.5, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "compute",
		},
		{
			MetricType:    "jobs_completed",
			Value:         12.3, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "gpu",
		},
		{
			MetricType:    "jobs_completed",
			Value:         8.7, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "highmem",
		},
		{
			MetricType:    "jobs_submitted",
			Value:         52.1, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "compute",
		},
		{
			MetricType:    "jobs_submitted",
			Value:         15.2, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "gpu",
		},
		{
			MetricType:    "jobs_submitted",
			Value:         9.8, // jobs per hour
			TimeWindow:    "1h",
			PartitionName: "highmem",
		},
		{
			MetricType:    "cpu_hours_delivered",
			Value:         1890.5, // CPU hours per hour
			TimeWindow:    "1h",
			PartitionName: "compute",
		},
		{
			MetricType:    "cpu_hours_delivered",
			Value:         420.3, // CPU hours per hour
			TimeWindow:    "1h",
			PartitionName: "gpu",
		},
		{
			MetricType:    "cpu_hours_delivered",
			Value:         185.7, // CPU hours per hour
			TimeWindow:    "1h",
			PartitionName: "highmem",
		},
	}

	for _, data := range throughputData {
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.SystemThroughput.WithLabelValues(
				pc.clusterName,
				data.MetricType,
				data.TimeWindow,
				data.PartitionName,
			).Desc(),
			prometheus.GaugeValue,
			data.Value,
			pc.clusterName,
			data.MetricType,
			data.TimeWindow,
			data.PartitionName,
		))
	}

	pc.LogCollectionf("Collected throughput metrics for %d data points", len(throughputData))
	return nil
}

// collectEfficiencyMetrics collects system efficiency metrics
func (pc *PerformanceCollector) collectEfficiencyMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate efficiency data
	efficiencyData := []struct {
		EfficiencyType string
		Value          float64
		PartitionName  string
		TimeWindow     string
	}{
		{
			EfficiencyType: "cpu_utilization",
			Value:          0.78, // 78% CPU utilization
			PartitionName:  "compute",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "memory_utilization",
			Value:          0.65, // 65% memory utilization
			PartitionName:  "compute",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "node_utilization",
			Value:          0.82, // 82% node utilization
			PartitionName:  "compute",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "cpu_utilization",
			Value:          0.85, // 85% CPU utilization
			PartitionName:  "gpu",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "memory_utilization",
			Value:          0.72, // 72% memory utilization
			PartitionName:  "gpu",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "node_utilization",
			Value:          0.90, // 90% node utilization
			PartitionName:  "gpu",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "cpu_utilization",
			Value:          0.91, // 91% CPU utilization
			PartitionName:  "highmem",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "memory_utilization",
			Value:          0.88, // 88% memory utilization
			PartitionName:  "highmem",
			TimeWindow:     "1h",
		},
		{
			EfficiencyType: "node_utilization",
			Value:          0.75, // 75% node utilization
			PartitionName:  "highmem",
			TimeWindow:     "1h",
		},
	}

	for _, data := range efficiencyData {
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.SystemEfficiency.WithLabelValues(
				pc.clusterName,
				data.EfficiencyType,
				data.PartitionName,
				data.TimeWindow,
			).Desc(),
			prometheus.GaugeValue,
			data.Value,
			pc.clusterName,
			data.EfficiencyType,
			data.PartitionName,
			data.TimeWindow,
		))
	}

	// Collect resource utilization metrics (more detailed breakdowns)
	resourceData := []struct {
		ResourceType  string
		Utilization   float64
		PartitionName string
	}{
		{"cpu", 0.78, "compute"},
		{"memory", 0.65, "compute"},
		{"storage", 0.45, "compute"},
		{"network", 0.23, "compute"},
		{"cpu", 0.85, "gpu"},
		{"memory", 0.72, "gpu"},
		{"gpu", 0.92, "gpu"},
		{"storage", 0.38, "gpu"},
		{"cpu", 0.91, "highmem"},
		{"memory", 0.88, "highmem"},
		{"storage", 0.52, "highmem"},
	}

	for _, data := range resourceData {
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.ResourceUtilization.WithLabelValues(
				pc.clusterName,
				data.ResourceType,
				data.PartitionName,
			).Desc(),
			prometheus.GaugeValue,
			data.Utilization,
			pc.clusterName,
			data.ResourceType,
			data.PartitionName,
		))
	}

	pc.LogCollectionf("Collected efficiency metrics for %d efficiency points and %d resource points",
		len(efficiencyData), len(resourceData))
	return nil
}

// collectQueueMetrics collects queue analysis metrics
func (pc *PerformanceCollector) collectQueueMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate queue analysis data
	queueData := []struct {
		PartitionName   string
		QueueDepth      int
		AvgWaitTime     time.Duration
		MedianWaitTime  time.Duration
		MaxWaitTime     time.Duration
		P95WaitTime     time.Duration
		JobTurnoverRate float64 // jobs per hour
		Priority        string
	}{
		{
			PartitionName:   "compute",
			QueueDepth:      125,
			AvgWaitTime:     45 * time.Minute,
			MedianWaitTime:  32 * time.Minute,
			MaxWaitTime:     3*time.Hour + 15*time.Minute,
			P95WaitTime:     2*time.Hour + 30*time.Minute,
			JobTurnoverRate: 42.5,
			Priority:        "normal",
		},
		{
			PartitionName:   "gpu",
			QueueDepth:      38,
			AvgWaitTime:     2*time.Hour + 15*time.Minute,
			MedianWaitTime:  1*time.Hour + 45*time.Minute,
			MaxWaitTime:     8*time.Hour + 30*time.Minute,
			P95WaitTime:     6*time.Hour + 45*time.Minute,
			JobTurnoverRate: 12.3,
			Priority:        "high",
		},
		{
			PartitionName:   "highmem",
			QueueDepth:      22,
			AvgWaitTime:     3*time.Hour + 45*time.Minute,
			MedianWaitTime:  2*time.Hour + 30*time.Minute,
			MaxWaitTime:     12 * time.Hour,
			P95WaitTime:     9*time.Hour + 15*time.Minute,
			JobTurnoverRate: 8.7,
			Priority:        "high",
		},
		{
			PartitionName:   "debug",
			QueueDepth:      8,
			AvgWaitTime:     5 * time.Minute,
			MedianWaitTime:  3 * time.Minute,
			MaxWaitTime:     25 * time.Minute,
			P95WaitTime:     18 * time.Minute,
			JobTurnoverRate: 15.2,
			Priority:        "urgent",
		},
	}

	for _, data := range queueData {
		// Queue depth
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.QueueDepth.WithLabelValues(pc.clusterName, data.PartitionName, data.Priority).Desc(),
			prometheus.GaugeValue,
			float64(data.QueueDepth),
			pc.clusterName,
			data.PartitionName,
			data.Priority,
		))

		// Wait time metrics (different percentiles)
		waitTimeTypes := map[string]time.Duration{
			"average": data.AvgWaitTime,
			"median":  data.MedianWaitTime,
			"max":     data.MaxWaitTime,
			"p95":     data.P95WaitTime,
		}

		for waitType, duration := range waitTimeTypes {
			pc.SendMetric(ch, pc.BuildMetric(
				pc.metrics.QueueWaitTime.WithLabelValues(
					pc.clusterName,
					data.PartitionName,
					waitType,
					data.Priority,
				).Desc(),
				prometheus.GaugeValue,
				duration.Seconds(),
				pc.clusterName,
				data.PartitionName,
				waitType,
				data.Priority,
			))
		}

		// Job turnover rate
		pc.SendMetric(ch, pc.BuildMetric(
			pc.metrics.JobTurnover.WithLabelValues(pc.clusterName, data.PartitionName).Desc(),
			prometheus.GaugeValue,
			data.JobTurnoverRate,
			pc.clusterName,
			data.PartitionName,
		))
	}

	pc.LogCollectionf("Collected queue analysis for %d partitions", len(queueData))
	return nil
}

// SystemPerformanceMetrics represents system performance metrics
type SystemPerformanceMetrics struct {
	ThroughputMetrics ThroughputMetrics
	EfficiencyMetrics SystemEfficiencyMetrics
	QueueMetrics      QueueMetrics
}

// ThroughputMetrics represents throughput statistics
type ThroughputMetrics struct {
	JobsPerHour     float64
	CPUHoursPerHour float64
	CompletionRate  float64
	SubmissionRate  float64
}

// SystemEfficiencyMetrics represents efficiency statistics
type SystemEfficiencyMetrics struct {
	CPUUtilization     float64
	MemoryUtilization  float64
	NodeUtilization    float64
	StorageUtilization float64
	NetworkUtilization float64
	OverallEfficiency  float64
}

// QueueMetrics represents queue analysis statistics
type QueueMetrics struct {
	QueueDepth      int
	AverageWaitTime time.Duration
	MedianWaitTime  time.Duration
	MaxWaitTime     time.Duration
	P95WaitTime     time.Duration
	TurnoverRate    float64
}

// TODO: Following calculation and parser functions are unused - preserved for future performance analysis
/*
// calculateEfficiency calculates overall system efficiency
func (pc *PerformanceCollector) calculateEfficiency(metrics *EfficiencyMetrics) float64 {
	// Weighted average of different resource utilizations
	weights := map[string]float64{
		"cpu":     0.4,
		"memory":  0.3,
		"node":    0.2,
		"storage": 0.05,
		"network": 0.05,
	}

	// Map efficiency metrics to utilization names
	utilizations := map[string]float64{
		"cpu":     metrics.CPUEfficiency,
		"memory":  metrics.MemoryEfficiency,
		"node":    metrics.ResourceEfficiency,    // Use resource efficiency as proxy for node
		"storage": metrics.IOEfficiency,          // Use IO efficiency as proxy for storage
		"network": metrics.NetworkEfficiency,
	}

	efficiency := 0.0
	for resource, weight := range weights {
		if util, ok := utilizations[resource]; ok {
			efficiency += util * weight
		}
	}

	return math.Min(efficiency, 1.0) // Cap at 100%
}

// parsePerformanceData parses SLURM performance data
func (pc *PerformanceCollector) parsePerformanceData(data interface{}) (*PerformanceMetrics, error) {
	// In real implementation, this would parse actual SLURM API response
	// For now, return mock performance data
	// Return simplified performance metrics based on the new structure
	return &PerformanceMetrics{
		Throughput:      45.5,  // Jobs per hour
		Latency:         120.0, // Average latency in seconds
		QueueTime:       300.0, // Average queue time in seconds
		ExecutionTime:   1800.0, // Average execution time in seconds
		TurnaroundTime:  2100.0, // Total turnaround time in seconds
		SuccessRate:     0.92,
		ErrorRate:       0.08,
	}, nil
}

// parseEfficiencyData parses efficiency data (now separate from performance)
func (pc *PerformanceCollector) parseEfficiencyData(data interface{}) (*EfficiencyMetrics, error) {
	// Return efficiency metrics using the correct structure
	return &EfficiencyMetrics{
		CPUEfficiency:       0.78,
		MemoryEfficiency:    0.65,
		ResourceEfficiency:  0.715, // Average of CPU and memory
		IOEfficiency:        0.45,
		NetworkEfficiency:   0.23,
		OverallEfficiency:   0.725,
		WasteRatio:          0.275,
		OptimizationScore:   0.68,
		EfficiencyGrade:     "B",
		EfficiencyCategory:  "Good",
		QueueEfficiency:     0.82,
		Recommendations: []string{
			"Consider optimizing memory allocation",
			"Network utilization is low - review network requirements",
		},
	}, nil
}
*/
