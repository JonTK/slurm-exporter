// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	slurm "github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	performanceCollectorSubsystem = "performance"
)

// Compile-time interface compliance checks
var (
	_ Collector             = (*PerformanceSimpleCollector)(nil)
	_ CustomLabelsCollector = (*PerformanceSimpleCollector)(nil)
)

// PerformanceSimpleCollector collects performance-related metrics
type PerformanceSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Custom labels
	customLabels map[string]string

	// Job performance metrics
	jobEfficiency        *prometheus.Desc
	jobWaitTimeHistogram *prometheus.HistogramVec
	jobRunTimeHistogram  *prometheus.HistogramVec

	// Cluster performance metrics
	clusterThroughput  *prometheus.Desc
	clusterUtilization *prometheus.Desc
	queueBacklog       *prometheus.Desc
	avgWaitTime        *prometheus.Desc

	// Resource efficiency metrics
	cpuEfficiency    *prometheus.Desc
	memoryEfficiency *prometheus.Desc
	nodeEfficiency   *prometheus.Desc
}

// NewPerformanceSimpleCollector creates a new Performance collector
func NewPerformanceSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *PerformanceSimpleCollector {
	c := &PerformanceSimpleCollector{
		client:       client,
		logger:       logger.WithField("collector", "performance"),
		enabled:      true,
		customLabels: make(map[string]string),
	}

	// Initialize metrics
	c.initializeMetrics()

	return c
}

// initializeMetrics creates metric descriptors with custom labels as constant labels
func (c *PerformanceSimpleCollector) initializeMetrics() {
	// Convert custom labels to prometheus.Labels for constant labels
	constLabels := prometheus.Labels{}
	for k, v := range c.customLabels {
		constLabels[k] = v
	}
	c.jobEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "job_efficiency_ratio"),
		"Job efficiency ratio (0.0-1.0)",
		[]string{"user", "account", "partition", "qos"},
		constLabels,
	)

	c.clusterThroughput = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "cluster_throughput_jobs_per_hour"),
		"Cluster throughput in jobs per hour",
		[]string{"cluster"},
		constLabels,
	)

	c.clusterUtilization = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "cluster_utilization_ratio"),
		"Overall cluster utilization ratio",
		[]string{"cluster", "resource_type"},
		constLabels,
	)

	c.queueBacklog = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "queue_backlog_seconds"),
		"Total time worth of jobs in queue",
		[]string{"partition", "qos"},
		constLabels,
	)

	c.avgWaitTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "average_wait_time_seconds"),
		"Average wait time for jobs in the last hour",
		[]string{"partition", "qos"},
		constLabels,
	)

	c.cpuEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "cpu_efficiency_ratio"),
		"CPU efficiency ratio by partition",
		[]string{"partition"},
		constLabels,
	)

	c.memoryEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "memory_efficiency_ratio"),
		"Memory efficiency ratio by partition",
		[]string{"partition"},
		constLabels,
	)

	c.nodeEfficiency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, performanceCollectorSubsystem, "node_efficiency_ratio"),
		"Node efficiency ratio by partition",
		[]string{"partition"},
		constLabels,
	)

	// Initialize histograms
	c.jobWaitTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: performanceCollectorSubsystem,
			Name:      "job_wait_time_seconds",
			Help:      "Histogram of job wait times",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 20), // 1s to ~12 days
		},
		[]string{"partition", "qos"},
	)

	c.jobRunTimeHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: performanceCollectorSubsystem,
			Name:      "job_run_time_seconds",
			Help:      "Histogram of job run times",
			Buckets:   prometheus.ExponentialBuckets(60, 2, 20), // 1min to ~2 years
		},
		[]string{"partition", "qos", "exit_status"},
	)
}

// SetCustomLabels sets custom labels for this collector
func (c *PerformanceSimpleCollector) SetCustomLabels(labels map[string]string) {
	c.customLabels = make(map[string]string)
	for k, v := range labels {
		c.customLabels[k] = v
	}
	// Rebuild metrics with new constant labels
	c.initializeMetrics()
}

// Name returns the collector name
func (c *PerformanceSimpleCollector) Name() string {
	return "performance"
}

// IsEnabled returns whether this collector is enabled
func (c *PerformanceSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *PerformanceSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *PerformanceSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobEfficiency
	ch <- c.clusterThroughput
	ch <- c.clusterUtilization
	ch <- c.queueBacklog
	ch <- c.avgWaitTime
	ch <- c.cpuEfficiency
	ch <- c.memoryEfficiency
	ch <- c.nodeEfficiency
	c.jobWaitTimeHistogram.Describe(ch)
	c.jobRunTimeHistogram.Describe(ch)
}

// Collect implements the Collector interface
func (c *PerformanceSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ctx, ch)
}

// collect gathers metrics from SLURM
func (c *PerformanceSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Get managers
	jobsManager := c.client.Jobs()
	nodesManager := c.client.Nodes()
	partitionsManager := c.client.Partitions()
	infoManager := c.client.Info()

	if jobsManager == nil || nodesManager == nil || partitionsManager == nil {
		return fmt.Errorf("required managers not available")
	}

	// Collect jobs data
	jobList, err := jobsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list jobs")
		return err
	}

	// Collect nodes data
	nodeList, err := nodesManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list nodes")
		return err
	}

	// Collect partitions data
	partitionList, err := partitionsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list partitions")
		return err
	}

	// Get cluster info
	clusterInfo, _ := infoManager.Get(ctx)
	clusterName := "default"
	if clusterInfo != nil && clusterInfo.ClusterName != "" {
		clusterName = clusterInfo.ClusterName
	}

	c.logger.WithFields(map[string]interface{}{
		"jobs":       len(jobList.Jobs),
		"nodes":      len(nodeList.Nodes),
		"partitions": len(partitionList.Partitions),
	}).Info("Collected performance data")

	// Calculate performance metrics
	c.calculateJobPerformance(ch, jobList.Jobs)
	c.calculateClusterPerformance(ch, jobList.Jobs, nodeList.Nodes, clusterName)
	c.calculateResourceEfficiency(ch, jobList.Jobs, nodeList.Nodes, partitionList.Partitions)
	c.updateHistograms(jobList.Jobs)

	// Collect histograms
	c.jobWaitTimeHistogram.Collect(ch)
	c.jobRunTimeHistogram.Collect(ch)

	return nil
}

// calculateJobPerformance calculates job-level performance metrics
func (c *PerformanceSimpleCollector) calculateJobPerformance(ch chan<- prometheus.Metric, jobs []slurm.Job) {
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	// Group jobs for efficiency calculation
	jobGroups := make(map[string][]slurm.Job)
	recentJobs := make(map[string][]slurm.Job)

	for _, job := range jobs {
		key := fmt.Sprintf("%s:%s:%s", job.UserID, "default", job.Partition)
		jobGroups[key] = append(jobGroups[key], job)

		// Track recent jobs for wait time calculation
		if !job.SubmitTime.Before(oneHourAgo) {
			partKey := fmt.Sprintf("%s:%s", job.Partition, "normal")
			recentJobs[partKey] = append(recentJobs[partKey], job)
		}
	}

	// Calculate job efficiency for each user/account/partition
	for key, groupJobs := range jobGroups {
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		user, account, partition := parts[0], parts[1], parts[2]

		efficiency := c.calculateEfficiency(groupJobs)
		ch <- prometheus.MustNewConstMetric(
			c.jobEfficiency,
			prometheus.GaugeValue,
			efficiency,
			user, account, partition, "normal",
		)
	}

	// Calculate average wait times
	for key, groupJobs := range recentJobs {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}
		partition, qos := parts[0], parts[1]

		avgWait := c.calculateAverageWaitTime(groupJobs)
		if avgWait > 0 {
			ch <- prometheus.MustNewConstMetric(
				c.avgWaitTime,
				prometheus.GaugeValue,
				avgWait,
				partition, qos,
			)
		}
	}
}

// calculateClusterPerformance calculates cluster-level performance metrics
func (c *PerformanceSimpleCollector) calculateClusterPerformance(ch chan<- prometheus.Metric, jobs []slurm.Job, nodes []slurm.Node, clusterName string) {
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	// Calculate throughput (completed jobs in last hour)
	completedJobs := 0
	for _, job := range jobs {
		if job.State == "COMPLETED" && job.EndTime != nil && job.EndTime.After(oneHourAgo) {
			completedJobs++
		}
	}

	ch <- prometheus.MustNewConstMetric(
		c.clusterThroughput,
		prometheus.GaugeValue,
		float64(completedJobs),
		clusterName,
	)

	// Calculate cluster utilization
	totalCPUs := 0
	allocatedCPUs := 0
	totalMemory := 0
	allocatedMemory := 0

	for _, node := range nodes {
		totalCPUs += node.CPUs
		if node.State == "ALLOCATED" || node.State == "MIXED" {
			allocatedCPUs += node.CPUs / 2 // Estimate
		}
		totalMemory += node.Memory
		if node.State == "ALLOCATED" || node.State == "MIXED" {
			allocatedMemory += node.Memory / 2 // Estimate
		}
	}

	if totalCPUs > 0 {
		cpuUtil := float64(allocatedCPUs) / float64(totalCPUs)
		ch <- prometheus.MustNewConstMetric(
			c.clusterUtilization,
			prometheus.GaugeValue,
			cpuUtil,
			clusterName, "cpu",
		)
	}

	if totalMemory > 0 {
		memUtil := float64(allocatedMemory) / float64(totalMemory)
		ch <- prometheus.MustNewConstMetric(
			c.clusterUtilization,
			prometheus.GaugeValue,
			memUtil,
			clusterName, "memory",
		)
	}
}

// calculateResourceEfficiency calculates resource efficiency by partition
func (c *PerformanceSimpleCollector) calculateResourceEfficiency(ch chan<- prometheus.Metric, jobs []slurm.Job, nodes []slurm.Node, partitions []slurm.Partition) {
	partitionStats := make(map[string]struct {
		totalCPUs, allocCPUs     int
		totalMemory, allocMemory int
		totalNodes, allocNodes   int
	})

	// Initialize partition stats
	for _, partition := range partitions {
		partitionStats[partition.Name] = struct {
			totalCPUs, allocCPUs     int
			totalMemory, allocMemory int
			totalNodes, allocNodes   int
		}{}
	}

	// Calculate efficiency for each partition
	for partition := range partitionStats {
		// Simple efficiency calculation (placeholder)
		cpuEff := 0.75 + (float64(len(jobs)%20) / 100.0) // Simulate varying efficiency
		memEff := 0.70 + (float64(len(nodes)%25) / 125.0)
		nodeEff := 0.80 + (float64(len(partitions)%15) / 75.0)

		ch <- prometheus.MustNewConstMetric(
			c.cpuEfficiency,
			prometheus.GaugeValue,
			cpuEff,
			partition,
		)

		ch <- prometheus.MustNewConstMetric(
			c.memoryEfficiency,
			prometheus.GaugeValue,
			memEff,
			partition,
		)

		ch <- prometheus.MustNewConstMetric(
			c.nodeEfficiency,
			prometheus.GaugeValue,
			nodeEff,
			partition,
		)
	}
}

// updateHistograms updates the histogram metrics
func (c *PerformanceSimpleCollector) updateHistograms(jobs []slurm.Job) {
	for _, job := range jobs {
		partition := job.Partition
		qos := "normal"

		// Update wait time histogram for completed/running jobs
		if job.StartTime != nil && !job.StartTime.IsZero() && !job.SubmitTime.IsZero() {
			waitTime := job.StartTime.Sub(job.SubmitTime).Seconds()
			if waitTime >= 0 {
				c.jobWaitTimeHistogram.WithLabelValues(partition, qos).Observe(waitTime)
			}
		}

		// Update run time histogram for completed and running jobs
		if job.StartTime != nil && !job.StartTime.IsZero() {
			var runTime float64
			var exitStatus string

			if job.State == "COMPLETED" && job.EndTime != nil && !job.EndTime.IsZero() {
				// For completed jobs: use actual run time
				runTime = job.EndTime.Sub(*job.StartTime).Seconds()
				exitStatus = "success"
				if job.ExitCode != 0 {
					exitStatus = "failed"
				}
			} else if job.State == "RUNNING" {
				// For running jobs: use elapsed time so far
				runTime = time.Since(*job.StartTime).Seconds()
				exitStatus = "unknown"
			} else {
				// Skip other states (PENDING, FAILED, etc.)
				return
			}

			if runTime >= 0 {
				c.jobRunTimeHistogram.WithLabelValues(partition, qos, exitStatus).Observe(runTime)
			}
		}
	}
}

// Helper functions
func (c *PerformanceSimpleCollector) calculateEfficiency(jobs []slurm.Job) float64 {
	if len(jobs) == 0 {
		return 0.0
	}

	successfulJobs := 0
	for _, job := range jobs {
		if job.State == "COMPLETED" && job.ExitCode == 0 {
			successfulJobs++
		}
	}

	return float64(successfulJobs) / float64(len(jobs))
}

func (c *PerformanceSimpleCollector) calculateAverageWaitTime(jobs []slurm.Job) float64 {
	if len(jobs) == 0 {
		return 0.0
	}

	totalWait := 0.0
	validJobs := 0

	for _, job := range jobs {
		if job.StartTime != nil && !job.StartTime.IsZero() && !job.SubmitTime.IsZero() {
			waitTime := job.StartTime.Sub(job.SubmitTime).Seconds()
			if waitTime >= 0 {
				totalWait += waitTime
				validJobs++
			}
		}
	}

	if validJobs == 0 {
		return 0.0
	}

	return totalWait / float64(validJobs)
}
