package collector

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// JobStepPerformanceCollector collects job step-level performance metrics and detects bottlenecks
type JobStepPerformanceCollector struct {
	slurmClient    slurm.SlurmClient
	logger         *slog.Logger
	config         *JobStepConfig
	metrics        *JobStepMetrics
	lastCollection time.Time
	mu             sync.RWMutex

	// Cache for step data
	stepCache       map[string]*JobStepDetails
	bottleneckCache map[string]*BottleneckAnalysis
	cacheTTL        time.Duration
}

// JobStepDetails represents detailed job step information
type JobStepDetails struct {
	JobID      string
	StepID     string
	StepName   string
	State      string
	StartTime  *time.Time
	EndTime    *time.Time
	CPUs       int
	Memory     int64
	Tasks      int
	Nodes      int
	CPUTime    float64
	UserTime   float64
	SystemTime float64
}

// JobStepConfig holds configuration for job step performance collection
type JobStepConfig struct {
	CollectionInterval        time.Duration         `yaml:"collection_interval"`
	MaxJobsPerCollection      int                   `yaml:"max_jobs_per_collection"`
	EnableBottleneckDetection bool                  `yaml:"enable_bottleneck_detection"`
	BottleneckThresholds      *BottleneckThresholds `yaml:"bottleneck_thresholds"`
	CacheTTL                  time.Duration         `yaml:"cache_ttl"`
	OnlyRunningJobs           bool                  `yaml:"only_running_jobs"`
}

// BottleneckThresholds defines thresholds for bottleneck detection
type BottleneckThresholds struct {
	CPUUtilizationLow      float64 `yaml:"cpu_utilization_low"`      // Below this is CPU underutilization
	MemoryUtilizationHigh  float64 `yaml:"memory_utilization_high"`  // Above this is memory pressure
	IOWaitHigh             float64 `yaml:"io_wait_high"`             // Above this is I/O bottleneck
	NetworkUtilizationHigh float64 `yaml:"network_utilization_high"` // Above this is network bottleneck
	LoadAverageHigh        float64 `yaml:"load_average_high"`        // Above this is CPU overload
}

// JobStepMetrics holds Prometheus metrics for job step performance
type JobStepMetrics struct {
	// Step-level resource utilization
	StepCPUUtilization     *prometheus.GaugeVec
	StepMemoryUtilization  *prometheus.GaugeVec
	StepIOUtilization      *prometheus.GaugeVec
	StepNetworkUtilization *prometheus.GaugeVec
	StepLoadAverage        *prometheus.GaugeVec

	// Step timing and performance
	StepDuration            *prometheus.GaugeVec
	StepExecutionEfficiency *prometheus.GaugeVec
	StepResourceEfficiency  *prometheus.GaugeVec

	// Bottleneck detection
	StepBottleneckDetected *prometheus.GaugeVec
	StepBottleneckSeverity *prometheus.GaugeVec
	StepBottleneckType     *prometheus.GaugeVec

	// Step state and progress
	StepsByState      *prometheus.GaugeVec
	StepProgressRatio *prometheus.GaugeVec

	// Collection performance
	CollectionDuration  prometheus.Histogram
	CollectionErrors    *prometheus.CounterVec
	StepsProcessed      prometheus.Counter
	BottlenecksDetected prometheus.Counter
}

// Note: Reusing JobStepDetails type defined earlier in this file

type BottleneckAnalysis struct {
	JobID           string
	StepID          string
	BottleneckType  string  // "cpu", "memory", "io", "network", "none"
	Severity        float64 // 0.0 to 1.0
	Detected        bool
	LastAnalyzed    time.Time
	Recommendations []string
}

// NewJobStepPerformanceCollector creates a new job step performance collector
func NewJobStepPerformanceCollector(slurmClient slurm.SlurmClient, logger *slog.Logger, config *JobStepConfig) (*JobStepPerformanceCollector, error) {
	if config == nil {
		config = &JobStepConfig{
			CollectionInterval:        30 * time.Second,
			MaxJobsPerCollection:      500,
			EnableBottleneckDetection: true,
			BottleneckThresholds: &BottleneckThresholds{
				CPUUtilizationLow:      0.3,  // Less than 30% CPU usage
				MemoryUtilizationHigh:  0.85, // More than 85% memory usage
				IOWaitHigh:             0.2,  // More than 20% I/O wait
				NetworkUtilizationHigh: 0.8,  // More than 80% network usage
				LoadAverageHigh:        0.9,  // Load average > 90% of cores
			},
			CacheTTL:        5 * time.Minute,
			OnlyRunningJobs: true,
		}
	}

	metrics := &JobStepMetrics{
		StepCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_cpu_utilization_ratio",
				Help: "CPU utilization ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepMemoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_memory_utilization_ratio",
				Help: "Memory utilization ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepIOUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_io_utilization_ratio",
				Help: "I/O utilization ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition", "io_type"},
		),
		StepNetworkUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_network_utilization_ratio",
				Help: "Network utilization ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition", "direction"},
		),
		StepLoadAverage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_load_average",
				Help: "Load average for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_duration_seconds",
				Help: "Job step duration in seconds",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition", "state"},
		),
		StepExecutionEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_execution_efficiency_ratio",
				Help: "Execution efficiency ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepResourceEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_resource_efficiency_ratio",
				Help: "Resource efficiency ratio for job steps",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepBottleneckDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_bottleneck_detected",
				Help: "Whether a bottleneck was detected for job step (1 = yes, 0 = no)",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		StepBottleneckSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_bottleneck_severity",
				Help: "Severity of detected bottleneck (0.0 to 1.0)",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition", "bottleneck_type"},
		),
		StepBottleneckType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_bottleneck_type_info",
				Help: "Type of bottleneck detected (informational, value always 1)",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition", "bottleneck_type"},
		),
		StepsByState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_steps_by_state_total",
				Help: "Number of job steps in each state",
			},
			[]string{"state", "partition"},
		),
		StepProgressRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_step_progress_ratio",
				Help: "Progress ratio for running job steps (estimated)",
			},
			[]string{"job_id", "step_id", "step_name", "user", "account", "partition"},
		),
		CollectionDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "slurm_job_step_performance_collection_duration_seconds",
				Help:    "Time spent collecting job step performance metrics",
				Buckets: prometheus.DefBuckets,
			},
		),
		CollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_step_performance_collection_errors_total",
				Help: "Total number of job step performance collection errors",
			},
			[]string{"error_type"},
		),
		StepsProcessed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_job_step_performance_steps_processed_total",
				Help: "Total number of job steps processed for performance metrics",
			},
		),
		BottlenecksDetected: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_job_step_bottlenecks_detected_total",
				Help: "Total number of bottlenecks detected in job steps",
			},
		),
	}

	collector := &JobStepPerformanceCollector{
		slurmClient:     slurmClient,
		logger:          logger,
		config:          config,
		metrics:         metrics,
		stepCache:       make(map[string]*JobStepDetails),
		bottleneckCache: make(map[string]*BottleneckAnalysis),
		cacheTTL:        config.CacheTTL,
		lastCollection:  time.Time{},
	}

	return collector, nil
}

// Describe implements the prometheus.Collector interface
func (c *JobStepPerformanceCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.StepCPUUtilization.Describe(ch)
	c.metrics.StepMemoryUtilization.Describe(ch)
	c.metrics.StepIOUtilization.Describe(ch)
	c.metrics.StepNetworkUtilization.Describe(ch)
	c.metrics.StepLoadAverage.Describe(ch)
	c.metrics.StepDuration.Describe(ch)
	c.metrics.StepExecutionEfficiency.Describe(ch)
	c.metrics.StepResourceEfficiency.Describe(ch)
	c.metrics.StepBottleneckDetected.Describe(ch)
	c.metrics.StepBottleneckSeverity.Describe(ch)
	c.metrics.StepBottleneckType.Describe(ch)
	c.metrics.StepsByState.Describe(ch)
	c.metrics.StepProgressRatio.Describe(ch)
	c.metrics.CollectionDuration.Describe(ch)
	c.metrics.CollectionErrors.Describe(ch)
	c.metrics.StepsProcessed.Describe(ch)
	c.metrics.BottlenecksDetected.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *JobStepPerformanceCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		c.metrics.CollectionDuration.Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.collectJobStepMetrics(ctx); err != nil {
		c.logger.Error("Failed to collect job step metrics", "error", err)
		c.metrics.CollectionErrors.WithLabelValues("step_collection").Inc()
	}

	// Collect metrics from all registered collectors
	c.metrics.StepCPUUtilization.Collect(ch)
	c.metrics.StepMemoryUtilization.Collect(ch)
	c.metrics.StepIOUtilization.Collect(ch)
	c.metrics.StepNetworkUtilization.Collect(ch)
	c.metrics.StepLoadAverage.Collect(ch)
	c.metrics.StepDuration.Collect(ch)
	c.metrics.StepExecutionEfficiency.Collect(ch)
	c.metrics.StepResourceEfficiency.Collect(ch)
	c.metrics.StepBottleneckDetected.Collect(ch)
	c.metrics.StepBottleneckSeverity.Collect(ch)
	c.metrics.StepBottleneckType.Collect(ch)
	c.metrics.StepsByState.Collect(ch)
	c.metrics.StepProgressRatio.Collect(ch)
	c.metrics.CollectionDuration.Collect(ch)
	c.metrics.CollectionErrors.Collect(ch)
	c.metrics.StepsProcessed.Collect(ch)
	c.metrics.BottlenecksDetected.Collect(ch)
}

// collectJobStepMetrics collects step-level performance metrics
func (c *JobStepPerformanceCollector) collectJobStepMetrics(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get job manager from SLURM client
	jobManager := c.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// List jobs to analyze steps for
	// TODO: ListJobsOptions structure is not compatible with current slurm-client
	// Using nil for options as a workaround
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	c.logger.Debug("Collecting job step metrics", "job_count", len(jobs.Jobs))

	// Reset aggregation metrics
	c.metrics.StepsByState.Reset()

	stepStateCounts := make(map[string]map[string]int) // partition -> state -> count

	// Process each job to get step details
	// TODO: Job type and field names are not compatible with current slurm-client version
	// Skipping job step processing for now
	_ = jobs // Suppress unused variable warning
	/*
		for _, job := range jobs.Jobs {
			// For now, create simplified step analysis since GetJobStepDetails doesn't exist yet
			// This would use jobManager.GetJobStepDetails(ctx, job.JobID) when available
			stepDetails := c.createSimplifiedStepDetails(job)

			// Cache the step details
			stepKey := fmt.Sprintf("%s:0", job.JobID) // Step 0 for main job step
			c.stepCache[stepKey] = stepDetails

			// Update metrics from step details
			c.updateMetricsFromStepDetails(job, stepDetails)

			// Perform bottleneck analysis if enabled
			if c.config.EnableBottleneckDetection {
				bottleneckAnalysis := c.analyzeBottlenecks(job, stepDetails)
				c.bottleneckCache[stepKey] = bottleneckAnalysis
				c.updateBottleneckMetrics(job, stepDetails, bottleneckAnalysis)
			}

			// Update aggregation counters
			c.updateStepStateCounts(job, stepDetails, stepStateCounts)
			c.metrics.StepsProcessed.Inc()
		}
	*/

	// Update aggregation metrics
	c.updateStepAggregationMetrics(stepStateCounts)

	// Clean expired cache entries
	c.cleanExpiredCache()

	c.lastCollection = time.Now()
	return nil
}

// TODO: Following functions are unused - preserved for future job step performance analysis features
// These implement detailed step-level performance tracking, bottleneck analysis, and metrics reporting.
/*
// createSimplifiedStepDetails creates step details from basic job data
func (c *JobStepPerformanceCollector) createSimplifiedStepDetails(job *slurm.Job) *JobStepDetails {
	// TODO: Job field names are not compatible with current slurm-client version
	// Returning empty step details for now
	return &JobStepDetails{
		JobID:      "0",
		StepID:     "0",
		StepName:   "unknown",
		State:      "UNKNOWN",
		StartTime:  nil,
		EndTime:    nil,
		CPUs:       0,
		Memory:     0,
		Nodes:      0,
		CPUTime:    0,
		UserTime:   0,
		SystemTime: 0,
	}
}

// updateMetricsFromStepDetails updates Prometheus metrics from step details
func (c *JobStepPerformanceCollector) updateMetricsFromStepDetails(job *slurm.Job, step *JobStepDetails) {
	// TODO: Job field names are not compatible with current slurm-client version
	// Using placeholder labels for now
	labels := []string{
		step.JobID,
		step.StepID,
		step.StepName,
		"unknown_user",
		"unknown_account",
		"unknown_partition",
	}

	// Step duration
	if step.EndTime != nil && step.StartTime != nil {
		duration := step.EndTime.Sub(*step.StartTime).Seconds()
		durationLabels := append(labels, step.State)
		c.metrics.StepDuration.WithLabelValues(durationLabels...).Set(duration)
	}

	// Basic resource utilization estimates (would be actual data when available)
	if step.CPUs > 0 {
		// Simplified CPU utilization estimation (would be real data from step utilization)
		estimatedCPUUtil := 0.75 // Placeholder - would come from actual step data
		c.metrics.StepCPUUtilization.WithLabelValues(labels...).Set(estimatedCPUUtil)
	}

	if step.Memory > 0 {
		// Simplified memory utilization estimation
		estimatedMemoryUtil := 0.65 // Placeholder - would come from actual step data
		c.metrics.StepMemoryUtilization.WithLabelValues(labels...).Set(estimatedMemoryUtil)
	}

	// Execution efficiency (simplified calculation)
	if step.CPUTime > 0 && step.StartTime != nil && step.EndTime != nil {
		wallTime := step.EndTime.Sub(*step.StartTime).Seconds()
		if wallTime > 0 {
			efficiency := step.CPUTime / (wallTime * float64(step.CPUs))
			c.metrics.StepExecutionEfficiency.WithLabelValues(labels...).Set(efficiency)
		}
	}

	// Progress ratio for running steps
	if step.State == "RUNNING" && step.StartTime != nil {
		elapsed := time.Since(*step.StartTime).Seconds()
		// Simplified progress estimation (would use actual job time limit)
		estimatedProgress := elapsed / (2 * 3600) // Assume 2-hour jobs for estimation
		if estimatedProgress > 1.0 {
			estimatedProgress = 1.0
		}
		c.metrics.StepProgressRatio.WithLabelValues(labels...).Set(estimatedProgress)
	}
}

// analyzeBottlenecks performs bottleneck analysis on job step
func (c *JobStepPerformanceCollector) analyzeBottlenecks(job *slurm.Job, step *JobStepDetails) *BottleneckAnalysis {
	analysis := &BottleneckAnalysis{
		JobID:        step.JobID,
		StepID:       step.StepID,
		LastAnalyzed: time.Now(),
	}

	// Simplified bottleneck detection (would use actual performance data)
	// This is placeholder logic until real step utilization data is available

	// Simulate CPU underutilization detection
	estimatedCPUUtil := 0.25 // This would come from real step data
	if estimatedCPUUtil < c.config.BottleneckThresholds.CPUUtilizationLow {
		analysis.BottleneckType = "cpu_underutilization"
		analysis.Severity = (c.config.BottleneckThresholds.CPUUtilizationLow - estimatedCPUUtil) / c.config.BottleneckThresholds.CPUUtilizationLow
		analysis.Detected = true
		analysis.Recommendations = []string{
			"Consider reducing CPU allocation for future similar jobs",
			"Review job parallelization efficiency",
		}
	}

	// Simulate memory pressure detection
	estimatedMemoryUtil := 0.90 // This would come from real step data
	if estimatedMemoryUtil > c.config.BottleneckThresholds.MemoryUtilizationHigh {
		analysis.BottleneckType = "memory_pressure"
		analysis.Severity = (estimatedMemoryUtil - c.config.BottleneckThresholds.MemoryUtilizationHigh) / (1.0 - c.config.BottleneckThresholds.MemoryUtilizationHigh)
		analysis.Detected = true
		analysis.Recommendations = []string{
			"Consider increasing memory allocation",
			"Review memory usage patterns for optimization",
		}
	}

	// If no bottleneck detected
	if !analysis.Detected {
		analysis.BottleneckType = "none"
		analysis.Severity = 0.0
	}

	return analysis
}

// updateBottleneckMetrics updates bottleneck-related metrics
func (c *JobStepPerformanceCollector) updateBottleneckMetrics(job *slurm.Job, step *JobStepDetails, analysis *BottleneckAnalysis) {
	// TODO: Job field names are not compatible with current slurm-client version
	labels := []string{
		step.JobID,
		step.StepID,
		step.StepName,
		"unknown_user",
		"unknown_account",
		"unknown_partition",
	}

	// Bottleneck detected flag
	detectedValue := 0.0
	if analysis.Detected {
		detectedValue = 1.0
		c.metrics.BottlenecksDetected.Inc()
	}
	c.metrics.StepBottleneckDetected.WithLabelValues(labels...).Set(detectedValue)

	// Bottleneck severity and type
	if analysis.Detected {
		severityLabels := append(labels, analysis.BottleneckType)
		c.metrics.StepBottleneckSeverity.WithLabelValues(severityLabels...).Set(analysis.Severity)

		typeLabels := append(labels, analysis.BottleneckType)
		c.metrics.StepBottleneckType.WithLabelValues(typeLabels...).Set(1.0)
	}
}

// updateStepStateCounts updates state count tracking
func (c *JobStepPerformanceCollector) updateStepStateCounts(job *slurm.Job, step *JobStepDetails, stateCounts map[string]map[string]int) {
	// TODO: Job field names are not compatible with current slurm-client version
	partition := "unknown_partition"
	if stateCounts[partition] == nil {
		stateCounts[partition] = make(map[string]int)
	}
	stateCounts[partition][step.State]++
}
*/

// updateStepAggregationMetrics updates aggregated step metrics
func (c *JobStepPerformanceCollector) updateStepAggregationMetrics(stateCounts map[string]map[string]int) {
	for partition, states := range stateCounts {
		for state, count := range states {
			c.metrics.StepsByState.WithLabelValues(state, partition).Set(float64(count))
		}
	}
}

// cleanExpiredCache removes expired entries from caches
func (c *JobStepPerformanceCollector) cleanExpiredCache() {
	now := time.Now()

	// Clean step cache
	for key, step := range c.stepCache {
		if step.StartTime != nil && now.Sub(*step.StartTime) > c.cacheTTL {
			delete(c.stepCache, key)
		}
	}

	// Clean bottleneck cache
	for key, analysis := range c.bottleneckCache {
		if now.Sub(analysis.LastAnalyzed) > c.cacheTTL {
			delete(c.bottleneckCache, key)
		}
	}
}

// GetCacheSize returns the current size of the step cache
func (c *JobStepPerformanceCollector) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.stepCache)
}

// GetBottleneckCacheSize returns the current size of the bottleneck cache
func (c *JobStepPerformanceCollector) GetBottleneckCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.bottleneckCache)
}

// GetLastCollection returns the timestamp of the last successful collection
func (c *JobStepPerformanceCollector) GetLastCollection() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCollection
}
