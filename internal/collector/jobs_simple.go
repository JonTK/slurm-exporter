// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	jobsCollectorSubsystem = "job"
)

// Compile-time interface compliance checks
var (
	_ Collector                 = (*JobsSimpleCollector)(nil)
	_ FilterableCollector       = (*JobsSimpleCollector)(nil)
	_ CardinalityAwareCollector = (*JobsSimpleCollector)(nil)
	_ CustomLabelsCollector     = (*JobsSimpleCollector)(nil)
)

// JobsSimpleCollector collects job-related metrics
type JobsSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Metric filtering
	metricFilter *MetricFilter
	filterConfig config.FilterConfig

	// Cardinality management
	cardinalityManager *metrics.CardinalityManager

	// Custom labels
	customLabels map[string]string

	// Job state metrics
	jobStates *prometheus.Desc

	// Job timing metrics
	jobQueueTime *prometheus.Desc
	jobRunTime   *prometheus.Desc

	// Job resource metrics
	jobCPUs   *prometheus.Desc
	jobMemory *prometheus.Desc
	jobNodes  *prometheus.Desc

	// Job info metric
	jobInfo *prometheus.Desc
}

// NewJobsSimpleCollector creates a new Jobs collector
func NewJobsSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *JobsSimpleCollector {
	c := &JobsSimpleCollector{
		client:       client,
		logger:       logger.WithField("collector", "jobs"),
		enabled:      true,
		metricFilter: NewMetricFilter(DefaultMetricFilterConfig()),
		customLabels: make(map[string]string),
	}

	// Initialize metrics
	c.initializeMetrics()

	return c
}

// initializeMetrics creates metric descriptors with custom labels as constant labels
func (c *JobsSimpleCollector) initializeMetrics() {
	// Convert custom labels to prometheus.Labels for constant labels
	constLabels := prometheus.Labels{}
	for k, v := range c.customLabels {
		constLabels[k] = v
	}

	c.jobStates = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "state"),
		"Current state of SLURM jobs (1=active, 0=inactive)",
		[]string{"job_id", "job_name", "user", "partition", "state"},
		constLabels,
	)

	c.jobQueueTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "queue_time_seconds"),
		"Time spent in queue before job started",
		[]string{"job_id", "job_name", "user", "partition"},
		constLabels,
	)

	c.jobRunTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "run_time_seconds"),
		"Time spent running the job",
		[]string{"job_id", "job_name", "user", "partition"},
		constLabels,
	)

	c.jobCPUs = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "cpus"),
		"Number of CPUs allocated to the job",
		[]string{"job_id", "job_name", "user", "partition"},
		constLabels,
	)

	c.jobMemory = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "memory_bytes"),
		"Memory allocated to the job in bytes",
		[]string{"job_id", "job_name", "user", "partition"},
		constLabels,
	)

	c.jobNodes = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "nodes"),
		"Number of nodes allocated to the job",
		[]string{"job_id", "job_name", "user", "partition"},
		constLabels,
	)

	c.jobInfo = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "info"),
		"Job information",
		[]string{"job_id", "job_name", "user", "account", "partition", "qos", "state"},
		constLabels,
	)
}

// Name returns the collector name
func (c *JobsSimpleCollector) Name() string {
	return "jobs"
}

// IsEnabled returns whether this collector is enabled
func (c *JobsSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *JobsSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// SetMetricFilter sets the metric filter for this collector
func (c *JobsSimpleCollector) SetMetricFilter(filter *MetricFilter) {
	c.metricFilter = filter
}

// GetMetricFilter returns the current metric filter
func (c *JobsSimpleCollector) GetMetricFilter() *MetricFilter {
	return c.metricFilter
}

// UpdateFilterConfig updates the filter configuration
func (c *JobsSimpleCollector) UpdateFilterConfig(config config.FilterConfig) {
	c.filterConfig = config
	if c.metricFilter != nil {
		c.metricFilter = NewMetricFilter(config.Metrics)
	}
}

// SetCustomLabels sets custom labels for this collector
func (c *JobsSimpleCollector) SetCustomLabels(labels map[string]string) {
	c.customLabels = make(map[string]string)
	for k, v := range labels {
		c.customLabels[k] = v
	}
	// Rebuild metrics with new constant labels
	c.initializeMetrics()
}

// SetCardinalityManager sets the cardinality manager for this collector
func (c *JobsSimpleCollector) SetCardinalityManager(cm *metrics.CardinalityManager) {
	c.cardinalityManager = cm
}

// shouldCollectMetric checks if a metric should be collected based on filters
func (c *JobsSimpleCollector) shouldCollectMetric(name string, metricType MetricType, isTiming bool, isResource bool) bool {
	if c.metricFilter == nil {
		return true
	}

	info := MetricInfo{
		Name:       name,
		Type:       metricType,
		IsTiming:   isTiming,
		IsResource: isResource,
		IsInfo:     strings.Contains(name, "_info"),
	}

	return c.metricFilter.ShouldCollectMetric(info)
}

// shouldCollectWithCardinality checks cardinality limits before collecting a metric
func (c *JobsSimpleCollector) shouldCollectWithCardinality(metricName string, labels map[string]string) bool {
	if c.cardinalityManager == nil {
		return true
	}

	// Merge metric labels with custom labels
	allLabels := make(map[string]string)
	for k, v := range c.customLabels {
		allLabels[k] = v
	}
	for k, v := range labels {
		allLabels[k] = v
	}

	return c.cardinalityManager.ShouldCollectMetric(metricName, allLabels)
}

// Describe implements prometheus.Collector
func (c *JobsSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.jobStates
	ch <- c.jobQueueTime
	ch <- c.jobRunTime
	ch <- c.jobCPUs
	ch <- c.jobMemory
	ch <- c.jobNodes
	ch <- c.jobInfo
}

// Collect implements the Collector interface
func (c *JobsSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *JobsSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	// Get Jobs manager from client
	jobsManager := c.client.Jobs()
	if jobsManager == nil {
		return fmt.Errorf("jobs manager not available")
	}

	// List all jobs
	jobList, err := jobsManager.List(ctx, nil)
	if err != nil {
		c.logger.WithError(err).Error("Failed to list jobs")
		return err
	}

	c.logger.WithField("count", len(jobList.Jobs)).Info("Collected job entries")

	now := time.Now()

	for _, job := range jobList.Jobs {
		// Convert job ID to string
		jobID := job.ID

		// Get job details with safe defaults
		jobName := job.Name
		if jobName == "" {
			jobName = "unknown"
		}

		userName := job.UserID
		if userName == "" {
			userName = "unknown"
		}

		partition := "unknown"
		if job.Partition != "" {
			partition = job.Partition
		}

		jobState := "UNKNOWN"
		if job.State != "" {
			jobState = job.State
		}

		// Job state metric
		stateValue := 0.0
		if isJobActive(jobState) {
			stateValue = 1.0
		}

		// Create labels map for cardinality checking
		stateLabels := map[string]string{
			"job_id":    jobID,
			"job_name":  jobName,
			"user":      userName,
			"partition": partition,
			"state":     jobState,
		}

		if c.shouldCollectMetric("slurm_job_state", MetricTypeGauge, false, false) &&
			c.shouldCollectWithCardinality("slurm_job_state", stateLabels) {
			ch <- prometheus.MustNewConstMetric(
				c.jobStates,
				prometheus.GaugeValue,
				stateValue,
				jobID, jobName, userName, partition, jobState,
			)
		}

		// Queue time (if job has started)
		if job.StartTime != nil && !job.StartTime.IsZero() && !job.SubmitTime.IsZero() {
			queueTime := job.StartTime.Sub(job.SubmitTime).Seconds()
			// Only record positive queue times (sanity check)
			if queueTime >= 0 {
				timeLabels := map[string]string{
					"job_id":    jobID,
					"job_name":  jobName,
					"user":      userName,
					"partition": partition,
				}
				if c.shouldCollectMetric("slurm_job_queue_time_seconds", MetricTypeGauge, true, false) &&
					c.shouldCollectWithCardinality("slurm_job_queue_time_seconds", timeLabels) {
					ch <- prometheus.MustNewConstMetric(
						c.jobQueueTime,
						prometheus.GaugeValue,
						queueTime,
						jobID, jobName, userName, partition,
					)
				}
			}
		}

		// Run time (if job is running)
		if job.StartTime != nil && !job.StartTime.IsZero() && isJobActive(jobState) {
			runTime := now.Sub(*job.StartTime).Seconds()
			// Only record positive run times
			if runTime >= 0 {
				runLabels := map[string]string{
					"job_id":    jobID,
					"job_name":  jobName,
					"user":      userName,
					"partition": partition,
				}
				if c.shouldCollectMetric("slurm_job_run_time_seconds", MetricTypeGauge, true, false) &&
					c.shouldCollectWithCardinality("slurm_job_run_time_seconds", runLabels) {
					ch <- prometheus.MustNewConstMetric(
						c.jobRunTime,
						prometheus.GaugeValue,
						runTime,
						jobID, jobName, userName, partition,
					)
				}
			}
		}

		// Resource metrics - use safe defaults
		cpus := float64(0)
		if job.CPUs > 0 && job.CPUs < 10000 { // Sanity check for reasonable CPU count
			cpus = float64(job.CPUs)
		} else if job.CPUs >= 10000 {
			c.logger.WithFields(map[string]interface{}{
				"job_id": jobID,
				"cpus":   job.CPUs,
			}).Warn("Unusually high CPU count detected, using 0")
		}
		resourceLabels := map[string]string{
			"job_id":    jobID,
			"job_name":  jobName,
			"user":      userName,
			"partition": partition,
		}
		if c.shouldCollectMetric("slurm_job_cpus", MetricTypeGauge, false, true) &&
			c.shouldCollectWithCardinality("slurm_job_cpus", resourceLabels) {
			ch <- prometheus.MustNewConstMetric(
				c.jobCPUs,
				prometheus.GaugeValue,
				cpus,
				jobID, jobName, userName, partition,
			)
		}

		// Memory in bytes (from job.Memory field if available)
		memoryBytes := float64(0)
		if job.Memory > 0 && job.Memory < 1000000 { // Sanity check: < 1TB
			memoryBytes = float64(job.Memory * 1024 * 1024) // Convert MB to bytes
		}
		if c.shouldCollectMetric("slurm_job_memory_bytes", MetricTypeGauge, false, true) &&
			c.shouldCollectWithCardinality("slurm_job_memory_bytes", resourceLabels) {
			ch <- prometheus.MustNewConstMetric(
				c.jobMemory,
				prometheus.GaugeValue,
				memoryBytes,
				jobID, jobName, userName, partition,
			)
		}

		// Nodes (calculate from available data)
		nodes := float64(1) // Default for single-node jobs
		if len(job.Nodes) > 0 {
			nodes = float64(len(job.Nodes))
		}
		if c.shouldCollectMetric("slurm_job_nodes", MetricTypeGauge, false, true) &&
			c.shouldCollectWithCardinality("slurm_job_nodes", resourceLabels) {
			ch <- prometheus.MustNewConstMetric(
				c.jobNodes,
				prometheus.GaugeValue,
				nodes,
				jobID, jobName, userName, partition,
			)
		}

		// Job info
		// These fields don't exist in interfaces.Job, use defaults
		account := "default"
		qos := "normal"

		infoLabels := map[string]string{
			"job_id":    jobID,
			"job_name":  jobName,
			"user":      userName,
			"account":   account,
			"partition": partition,
			"qos":       qos,
			"state":     jobState,
		}
		if c.shouldCollectMetric("slurm_job_info", MetricTypeInfo, false, false) &&
			c.shouldCollectWithCardinality("slurm_job_info", infoLabels) {
			ch <- prometheus.MustNewConstMetric(
				c.jobInfo,
				prometheus.GaugeValue,
				1,
				jobID, jobName, userName, account, partition, qos, jobState,
			)
		}
	}

	return nil
}

// isJobActive returns true if the job is in an active state
func isJobActive(state string) bool {
	state = strings.ToUpper(state)
	switch state {
	case "RUNNING", "COMPLETING":
		return true
	case "PENDING", "COMPLETED", "FAILED", "CANCELLED", "TIMEOUT":
		return false
	default:
		// Default to considering unknown states as inactive
		return false
	}
}
