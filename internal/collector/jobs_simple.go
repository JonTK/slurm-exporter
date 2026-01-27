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
	jobQueueTime  *prometheus.Desc
	jobRunTime    *prometheus.Desc
	jobSubmitTime *prometheus.Desc

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

	c.jobSubmitTime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, jobsCollectorSubsystem, "submit_time"),
		"Unix timestamp when the job was submitted",
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
//
//nolint:unparam
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
	ch <- c.jobSubmitTime
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
	return c.collect(ctx, ch)
}

// collect gathers metrics from SLURM
// jobContext holds extracted job information for metric collection
type jobContext struct {
	jobID     string
	jobName   string
	userName  string
	partition string
	jobState  string
}

// extractJobContext extracts and normalizes job fields with safe defaults
func extractJobContext(job slurm.Job) jobContext {
	ctx := jobContext{
		jobID:     job.ID,
		jobName:   job.Name,
		userName:  job.UserID,
		partition: job.Partition,
		jobState:  job.State,
	}

	// Apply safe defaults
	if ctx.jobName == "" {
		ctx.jobName = "unknown"
	}
	if ctx.userName == "" {
		ctx.userName = "unknown"
	}
	if ctx.partition == "" {
		ctx.partition = "unknown"
	}
	if ctx.jobState == "" {
		ctx.jobState = "UNKNOWN"
	}

	return ctx
}

// createJobLabels creates a labels map for cardinality checking
func (ctx *jobContext) createJobLabels() map[string]string {
	return map[string]string{
		"job_id":    ctx.jobID,
		"job_name":  ctx.jobName,
		"user":      ctx.userName,
		"partition": ctx.partition,
	}
}

// createStateLabels creates a labels map including state
func (ctx *jobContext) createStateLabels() map[string]string {
	labels := ctx.createJobLabels()
	labels["state"] = ctx.jobState
	return labels
}

func (c *JobsSimpleCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
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
		jobCtx := extractJobContext(job)
		c.collectJobMetrics(ch, job, jobCtx, now)
	}

	return nil
}

// collectJobMetrics collects all metrics for a single job
func (c *JobsSimpleCollector) collectJobMetrics(ch chan<- prometheus.Metric, job slurm.Job, ctx jobContext, now time.Time) {
	c.collectJobState(ch, ctx)
	c.collectQueueTime(ch, job, ctx)
	c.collectRunTime(ch, job, ctx, now)
	c.collectSubmitTime(ch, job, ctx)
	c.collectResourceMetrics(ch, job, ctx)
	c.collectJobInfo(ch, ctx)
}

// collectJobState collects job state metric
func (c *JobsSimpleCollector) collectJobState(ch chan<- prometheus.Metric, ctx jobContext) {
	stateValue := 0.0
	if isJobActive(ctx.jobState) {
		stateValue = 1.0
	}

	if c.shouldCollectMetric("slurm_job_state", MetricTypeGauge, false, false) &&
		c.shouldCollectWithCardinality("slurm_job_state", ctx.createStateLabels()) {
		ch <- prometheus.MustNewConstMetric(
			c.jobStates,
			prometheus.GaugeValue,
			stateValue,
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition, ctx.jobState,
		)
	}
}

// collectQueueTime collects queue time metric if applicable
func (c *JobsSimpleCollector) collectQueueTime(ch chan<- prometheus.Metric, job slurm.Job, ctx jobContext) {
	if job.StartTime == nil || job.StartTime.IsZero() || job.SubmitTime.IsZero() {
		return
	}

	queueTime := job.StartTime.Sub(job.SubmitTime).Seconds()
	if queueTime < 0 {
		return // Skip negative queue times
	}

	if c.shouldCollectMetric("slurm_job_queue_time_seconds", MetricTypeGauge, true, false) &&
		c.shouldCollectWithCardinality("slurm_job_queue_time_seconds", ctx.createJobLabels()) {
		ch <- prometheus.MustNewConstMetric(
			c.jobQueueTime,
			prometheus.GaugeValue,
			queueTime,
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}
}

// collectRunTime collects run time metric if job is active
func (c *JobsSimpleCollector) collectRunTime(ch chan<- prometheus.Metric, job slurm.Job, ctx jobContext, now time.Time) {
	if job.StartTime == nil || job.StartTime.IsZero() || !isJobActive(ctx.jobState) {
		return
	}

	runTime := now.Sub(*job.StartTime).Seconds()
	if runTime < 0 {
		return // Skip negative run times
	}

	if c.shouldCollectMetric("slurm_job_run_time_seconds", MetricTypeGauge, true, false) &&
		c.shouldCollectWithCardinality("slurm_job_run_time_seconds", ctx.createJobLabels()) {
		ch <- prometheus.MustNewConstMetric(
			c.jobRunTime,
			prometheus.GaugeValue,
			runTime,
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}
}

// collectSubmitTime collects job submit time metric if applicable
func (c *JobsSimpleCollector) collectSubmitTime(ch chan<- prometheus.Metric, job slurm.Job, ctx jobContext) {
	if job.SubmitTime.IsZero() {
		return
	}

	if c.shouldCollectMetric("slurm_job_submit_time", MetricTypeGauge, false, false) &&
		c.shouldCollectWithCardinality("slurm_job_submit_time", ctx.createJobLabels()) {
		ch <- prometheus.MustNewConstMetric(
			c.jobSubmitTime,
			prometheus.GaugeValue,
			float64(job.SubmitTime.Unix()),
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}
}

// collectResourceMetrics collects CPU, memory, and node metrics
func (c *JobsSimpleCollector) collectResourceMetrics(ch chan<- prometheus.Metric, job slurm.Job, ctx jobContext) {
	labels := ctx.createJobLabels()

	// CPU count
	cpus := c.sanitizeCPUCount(job.CPUs, ctx.jobID)
	if c.shouldCollectMetric("slurm_job_cpus", MetricTypeGauge, false, true) &&
		c.shouldCollectWithCardinality("slurm_job_cpus", labels) {
		ch <- prometheus.MustNewConstMetric(
			c.jobCPUs,
			prometheus.GaugeValue,
			float64(cpus),
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}

	// Memory in bytes
	memoryBytes := c.sanitizeMemory(job.Memory)
	if c.shouldCollectMetric("slurm_job_memory_bytes", MetricTypeGauge, false, true) &&
		c.shouldCollectWithCardinality("slurm_job_memory_bytes", labels) {
		ch <- prometheus.MustNewConstMetric(
			c.jobMemory,
			prometheus.GaugeValue,
			float64(memoryBytes),
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}

	// Node count
	nodes := c.calculateNodeCount(job.Nodes)
	if c.shouldCollectMetric("slurm_job_nodes", MetricTypeGauge, false, true) &&
		c.shouldCollectWithCardinality("slurm_job_nodes", labels) {
		ch <- prometheus.MustNewConstMetric(
			c.jobNodes,
			prometheus.GaugeValue,
			float64(nodes),
			ctx.jobID, ctx.jobName, ctx.userName, ctx.partition,
		)
	}
}

// sanitizeCPUCount validates and sanitizes CPU count
func (c *JobsSimpleCollector) sanitizeCPUCount(cpus int, jobID string) int {
	if cpus <= 0 {
		return 0
	}
	if cpus >= 10000 {
		c.logger.WithFields(map[string]interface{}{
			"job_id": jobID,
			"cpus":   cpus,
		}).Warn("Unusually high CPU count detected, using 0")
		return 0
	}
	return cpus
}

// sanitizeMemory validates memory value (already in bytes from slurm-client)
func (c *JobsSimpleCollector) sanitizeMemory(memoryBytes int) int64 {
	if memoryBytes <= 0 {
		return 0
	}
	// Sanity check: reject values > 1TB (1024^4 bytes)
	maxBytes := int64(1024 * 1024 * 1024 * 1024)
	if int64(memoryBytes) >= maxBytes {
		return 0
	}
	return int64(memoryBytes)
}

// calculateNodeCount determines node count from job node list
func (c *JobsSimpleCollector) calculateNodeCount(nodes []string) int {
	if len(nodes) == 0 {
		return 1 // Default for single-node jobs
	}
	return len(nodes)
}

// collectJobInfo collects job info metric
func (c *JobsSimpleCollector) collectJobInfo(ch chan<- prometheus.Metric, ctx jobContext) {
	// These fields don't exist in interfaces.Job, use defaults
	account := "default"
	qos := "normal"

	infoLabels := map[string]string{
		"job_id":    ctx.jobID,
		"job_name":  ctx.jobName,
		"user":      ctx.userName,
		"account":   account,
		"partition": ctx.partition,
		"qos":       qos,
		"state":     ctx.jobState,
	}

	if c.shouldCollectMetric("slurm_job_info", MetricTypeGauge, false, false) &&
		c.shouldCollectWithCardinality("slurm_job_info", infoLabels) {
		ch <- prometheus.MustNewConstMetric(
			c.jobInfo,
			prometheus.GaugeValue,
			1.0, // Info metric always has value 1
			ctx.jobID, ctx.jobName, ctx.userName, account, ctx.partition, qos, ctx.jobState,
		)
	}
}

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
