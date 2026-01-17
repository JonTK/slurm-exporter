package collector

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// SimplifiedJobPerformanceCollector provides basic job performance metrics using existing SLURM client types
type SimplifiedJobPerformanceCollector struct {
	slurmClient    slurm.SlurmClient
	logger         *slog.Logger
	config         *JobPerformanceConfig
	metrics        *SimplifiedJobMetrics
	efficiencyCalc *EfficiencyCalculator
	lastCollection time.Time
	mu             sync.RWMutex

	// Cache for recent job data
	jobCache        map[string]*slurm.Job
	efficiencyCache map[string]*EfficiencyMetrics
	cacheTTL        time.Duration
}

// SimplifiedJobMetrics holds basic Prometheus metrics for job performance
type SimplifiedJobMetrics struct {
	// Basic job metrics from existing data
	JobDuration        *prometheus.GaugeVec
	JobCPUAllocated    *prometheus.GaugeVec
	JobMemoryAllocated *prometheus.GaugeVec
	JobNodesAllocated  *prometheus.GaugeVec
	JobGPUAllocated    *prometheus.GaugeVec
	JobQueueTime       *prometheus.GaugeVec
	JobStartDelay      *prometheus.GaugeVec

	// Job state tracking
	JobStateTransitions *prometheus.CounterVec
	JobsByState         *prometheus.GaugeVec
	JobsByUser          *prometheus.GaugeVec
	JobsByAccount       *prometheus.GaugeVec
	JobsByPartition     *prometheus.GaugeVec

	// Performance indicators (estimated)
	JobResourceRatio      *prometheus.GaugeVec
	JobEfficiencyEstimate *prometheus.GaugeVec

	// Efficiency metrics from calculator
	JobCPUEfficiencyScore     *prometheus.GaugeVec
	JobMemoryEfficiencyScore  *prometheus.GaugeVec
	JobOverallEfficiencyScore *prometheus.GaugeVec
	JobEfficiencyGrade        *prometheus.GaugeVec
	JobWasteRatio             *prometheus.GaugeVec

	// Collection performance metrics
	CollectionDuration prometheus.Histogram
	CollectionErrors   *prometheus.CounterVec
	JobsProcessed      prometheus.Counter
	CacheHitRatio      prometheus.Gauge
}

// NewSimplifiedJobPerformanceCollector creates a new simplified job performance collector
func NewSimplifiedJobPerformanceCollector(slurmClient slurm.SlurmClient, logger *slog.Logger, config *JobPerformanceConfig) (*SimplifiedJobPerformanceCollector, error) {
	if config == nil {
		config = &JobPerformanceConfig{
			CollectionInterval:   30 * time.Second,
			MaxJobsPerCollection: 1000,
			CacheTTL:             5 * time.Minute,
		}
	}

	metrics := &SimplifiedJobMetrics{
		JobDuration: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_duration_seconds",
				Help: "Job duration in seconds",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobCPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpus_allocated_total",
				Help: "Number of CPUs allocated to job",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobMemoryAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_allocated_bytes",
				Help: "Memory allocated to job in bytes",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobNodesAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_nodes_allocated_total",
				Help: "Number of nodes allocated to job",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobGPUAllocated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_gpus_allocated_total",
				Help: "Number of GPUs allocated to job",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobQueueTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_queue_time_seconds",
				Help: "Time job spent in queue before starting",
			},
			[]string{"job_id", "job_name", "user", "account", "partition"},
		),
		JobStartDelay: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_start_delay_seconds",
				Help: "Delay between job submission and start",
			},
			[]string{"job_id", "job_name", "user", "account", "partition"},
		),
		JobStateTransitions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_state_transitions_total",
				Help: "Total number of job state transitions",
			},
			[]string{"user", "account", "partition", "from_state", "to_state"},
		),
		JobsByState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_jobs_by_state_total",
				Help: "Number of jobs in each state",
			},
			[]string{"state", "partition"},
		),
		JobsByUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_jobs_by_user_total",
				Help: "Number of jobs per user",
			},
			[]string{"user", "account", "state"},
		),
		JobsByAccount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_jobs_by_account_total",
				Help: "Number of jobs per account",
			},
			[]string{"account", "state"},
		),
		JobsByPartition: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_jobs_by_partition_total",
				Help: "Number of jobs per partition",
			},
			[]string{"partition", "state"},
		),
		JobResourceRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_ratio",
				Help: "Job resource allocation ratio (nodes * cpus_per_node)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobEfficiencyEstimate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_efficiency_estimate_ratio",
				Help: "Estimated job efficiency based on runtime vs time limit",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobCPUEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpu_efficiency_score",
				Help: "CPU efficiency score calculated by efficiency algorithm",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobMemoryEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_efficiency_score",
				Help: "Memory efficiency score calculated by efficiency algorithm",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobOverallEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_overall_efficiency_score",
				Help: "Overall efficiency score calculated by efficiency algorithm",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobEfficiencyGrade: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_efficiency_grade_info",
				Help: "Job efficiency grade (A=5, B=4, C=3, D=2, F=1)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state", "grade"},
		),
		JobWasteRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_waste_ratio",
				Help: "Resource waste ratio calculated by efficiency algorithm",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		CollectionDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "slurm_job_performance_collection_duration_seconds",
				Help:    "Time spent collecting job performance metrics",
				Buckets: prometheus.DefBuckets,
			},
		),
		CollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_performance_collection_errors_total",
				Help: "Total number of job performance collection errors",
			},
			[]string{"error_type"},
		),
		JobsProcessed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_job_performance_jobs_processed_total",
				Help: "Total number of jobs processed for performance metrics",
			},
		),
		CacheHitRatio: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_job_performance_cache_hit_ratio",
				Help: "Cache hit ratio for job performance data",
			},
		),
	}

	// Create efficiency calculator
	efficiencyCalc := NewEfficiencyCalculator(logger, nil)

	collector := &SimplifiedJobPerformanceCollector{
		slurmClient:     slurmClient,
		logger:          logger,
		config:          config,
		metrics:         metrics,
		efficiencyCalc:  efficiencyCalc,
		jobCache:        make(map[string]*slurm.Job),
		efficiencyCache: make(map[string]*EfficiencyMetrics),
		cacheTTL:        config.CacheTTL,
		lastCollection:  time.Time{},
	}

	return collector, nil
}

// Describe implements the prometheus.Collector interface
func (c *SimplifiedJobPerformanceCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.JobDuration.Describe(ch)
	c.metrics.JobCPUAllocated.Describe(ch)
	c.metrics.JobMemoryAllocated.Describe(ch)
	c.metrics.JobNodesAllocated.Describe(ch)
	c.metrics.JobGPUAllocated.Describe(ch)
	c.metrics.JobQueueTime.Describe(ch)
	c.metrics.JobStartDelay.Describe(ch)
	c.metrics.JobStateTransitions.Describe(ch)
	c.metrics.JobsByState.Describe(ch)
	c.metrics.JobsByUser.Describe(ch)
	c.metrics.JobsByAccount.Describe(ch)
	c.metrics.JobsByPartition.Describe(ch)
	c.metrics.JobResourceRatio.Describe(ch)
	c.metrics.JobEfficiencyEstimate.Describe(ch)
	c.metrics.JobCPUEfficiencyScore.Describe(ch)
	c.metrics.JobMemoryEfficiencyScore.Describe(ch)
	c.metrics.JobOverallEfficiencyScore.Describe(ch)
	c.metrics.JobEfficiencyGrade.Describe(ch)
	c.metrics.JobWasteRatio.Describe(ch)
	c.metrics.CollectionDuration.Describe(ch)
	c.metrics.CollectionErrors.Describe(ch)
	c.metrics.JobsProcessed.Describe(ch)
	c.metrics.CacheHitRatio.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *SimplifiedJobPerformanceCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		c.metrics.CollectionDuration.Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.collectJobMetrics(ctx); err != nil {
		c.logger.Error("Failed to collect job metrics", "error", err)
		c.metrics.CollectionErrors.WithLabelValues("job_collection").Inc()
	}

	// Collect metrics from all registered collectors
	c.metrics.JobDuration.Collect(ch)
	c.metrics.JobCPUAllocated.Collect(ch)
	c.metrics.JobMemoryAllocated.Collect(ch)
	c.metrics.JobNodesAllocated.Collect(ch)
	c.metrics.JobGPUAllocated.Collect(ch)
	c.metrics.JobQueueTime.Collect(ch)
	c.metrics.JobStartDelay.Collect(ch)
	c.metrics.JobStateTransitions.Collect(ch)
	c.metrics.JobsByState.Collect(ch)
	c.metrics.JobsByUser.Collect(ch)
	c.metrics.JobsByAccount.Collect(ch)
	c.metrics.JobsByPartition.Collect(ch)
	c.metrics.JobResourceRatio.Collect(ch)
	c.metrics.JobEfficiencyEstimate.Collect(ch)
	c.metrics.JobCPUEfficiencyScore.Collect(ch)
	c.metrics.JobMemoryEfficiencyScore.Collect(ch)
	c.metrics.JobOverallEfficiencyScore.Collect(ch)
	c.metrics.JobEfficiencyGrade.Collect(ch)
	c.metrics.JobWasteRatio.Collect(ch)
	c.metrics.CollectionDuration.Collect(ch)
	c.metrics.CollectionErrors.Collect(ch)
	c.metrics.JobsProcessed.Collect(ch)
	c.metrics.CacheHitRatio.Collect(ch)
}

// collectJobMetrics collects basic job metrics using existing SLURM client
func (c *SimplifiedJobPerformanceCollector) collectJobMetrics(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get job manager from SLURM client
	jobManager := c.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// List jobs with filters
	// Using nil for options as the exact structure is not clear
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	c.logger.Debug("Collecting basic job metrics", "job_count", len(jobs.Jobs))

	// Reset aggregation metrics
	c.resetAggregationMetrics()

	cacheHits := 0
	totalJobs := len(jobs.Jobs)

	// Track state counts for aggregation
	stateCounts := make(map[string]map[string]int)     // partition -> state -> count
	userStateCounts := make(map[string]map[string]int) // user -> state -> count
	accountStateCounts := make(map[string]int)         // account -> count

	// TODO: Job field names (JobID, etc.) are not available in the current slurm-client version
	// Skipping job processing for now
	_ = jobs // Suppress unused variable warning

	// Update aggregation metrics
	c.updateAggregationMetrics(stateCounts, userStateCounts, accountStateCounts)

	// Update cache hit ratio
	if totalJobs > 0 {
		c.metrics.CacheHitRatio.Set(float64(cacheHits) / float64(totalJobs))
	}

	// Clean expired cache entries
	c.cleanExpiredCache()

	c.lastCollection = time.Now()
	return nil
}

// TODO: updateMetricsFromJob is unused - preserved for future metrics updates
/*
// updateMetricsFromJob updates Prometheus metrics from basic job data
func (c *SimplifiedJobPerformanceCollector) updateMetricsFromJob(job *slurm.Job) {
	// TODO: Job field names are not compatible with current slurm-client version
	// Skipping metric updates for now

	labels := []string{
		job.JobID,
		job.Name,
		job.UserName,
		job.Account,
		job.Partition,
		job.JobState,
	}

	// Job duration (for completed jobs)
	if job.EndTime != nil && job.StartTime != nil {
		duration := job.EndTime.Sub(*job.StartTime).Seconds()
		c.metrics.JobDuration.WithLabelValues(labels...).Set(duration)
	}

	// Resource allocation
	if job.CPUs > 0 {
		c.metrics.JobCPUAllocated.WithLabelValues(labels...).Set(float64(job.CPUs))
	}

	if job.Memory > 0 {
		// Convert MB to bytes
		memoryBytes := float64(job.Memory) * 1024 * 1024
		c.metrics.JobMemoryAllocated.WithLabelValues(labels...).Set(memoryBytes)
	}

	if job.Nodes > 0 {
		c.metrics.JobNodesAllocated.WithLabelValues(labels...).Set(float64(job.Nodes))
	}

	// GPU allocation (if available)
	if job.TresCPU != "" {
		// Parse TRES for GPU count - simplified parsing
		// In real implementation, would parse TRES string properly
		c.metrics.JobGPUAllocated.WithLabelValues(labels...).Set(0) // Placeholder
	}

	// Queue time calculation
	if job.StartTime != nil && job.SubmitTime != nil {
		queueTime := job.StartTime.Sub(*job.SubmitTime).Seconds()
		queueLabels := labels[:5] // Exclude state for queue metrics
		c.metrics.JobQueueTime.WithLabelValues(queueLabels...).Set(queueTime)
		c.metrics.JobStartDelay.WithLabelValues(queueLabels...).Set(queueTime)
	}

	// Resource ratio (rough efficiency indicator)
	if job.CPUs > 0 && job.Nodes > 0 {
		resourceRatio := float64(job.CPUs) / float64(job.Nodes)
		c.metrics.JobResourceRatio.WithLabelValues(labels...).Set(resourceRatio)
	}

	// Efficiency estimate based on runtime vs time limit
	if job.EndTime != nil && job.StartTime != nil && job.TimeLimit > 0 {
		runtime := job.EndTime.Sub(*job.StartTime).Seconds()
		timeLimit := float64(job.TimeLimit) * 60 // Convert minutes to seconds
		if timeLimit > 0 {
			efficiencyEstimate := runtime / timeLimit
			c.metrics.JobEfficiencyEstimate.WithLabelValues(labels...).Set(efficiencyEstimate)
		}
	}

}
*/

// TODO: updateAggregationCounters is unused - preserved for future aggregation metrics
/*
// updateAggregationCounters updates counters for aggregation metrics
func (c *SimplifiedJobPerformanceCollector) updateAggregationCounters(
	job *slurm.Job,
	stateCounts map[string]map[string]int,
	userStateCounts map[string]map[string]int,
	accountStateCounts map[string]int,
) {
	// TODO: Job field names are not compatible with current slurm-client version
	// Skipping aggregation counter updates for now

	// State counts by partition
	if stateCounts[job.Partition] == nil {
		stateCounts[job.Partition] = make(map[string]int)
	}
	stateCounts[job.Partition][job.JobState]++

	// User state counts
	userKey := fmt.Sprintf("%s:%s", job.UserName, job.Account)
	if userStateCounts[userKey] == nil {
		userStateCounts[userKey] = make(map[string]int)
	}
	userStateCounts[userKey][job.JobState]++

	// Account counts
	accountStateCounts[job.Account]++

}
*/

// resetAggregationMetrics resets aggregation metrics before collection
func (c *SimplifiedJobPerformanceCollector) resetAggregationMetrics() {
	c.metrics.JobsByState.Reset()
	c.metrics.JobsByUser.Reset()
	c.metrics.JobsByAccount.Reset()
	c.metrics.JobsByPartition.Reset()
}

// updateAggregationMetrics updates aggregation metrics from collected counts
func (c *SimplifiedJobPerformanceCollector) updateAggregationMetrics(
	stateCounts map[string]map[string]int,
	userStateCounts map[string]map[string]int,
	accountStateCounts map[string]int,
) {
	// Jobs by state and partition
	for partition, states := range stateCounts {
		for state, count := range states {
			c.metrics.JobsByState.WithLabelValues(state, partition).Set(float64(count))
			c.metrics.JobsByPartition.WithLabelValues(partition, state).Set(float64(count))
		}
	}

	// Jobs by user and account
	for userKey, states := range userStateCounts {
		// Parse user:account key
		parts := strings.SplitN(userKey, ":", 2)
		if len(parts) == 2 {
			user, account := parts[0], parts[1]
			for state, count := range states {
				c.metrics.JobsByUser.WithLabelValues(user, account, state).Set(float64(count))
			}
		}
	}

	// Jobs by account
	for account, count := range accountStateCounts {
		c.metrics.JobsByAccount.WithLabelValues(account, "total").Set(float64(count))
	}
}

// cleanExpiredCache removes expired entries from the job cache
func (c *SimplifiedJobPerformanceCollector) cleanExpiredCache() {
	// TODO: Job cache uses slurm.Job type which has incompatible field names
	// For now, just clear old entries based on time
	if time.Since(c.lastCollection) > c.cacheTTL {
		c.jobCache = make(map[string]*slurm.Job)
	}
}

// GetCacheSize returns the current size of the job cache
func (c *SimplifiedJobPerformanceCollector) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.jobCache)
}

// TODO: updateEfficiencyMetrics, setEfficiencyMetrics, and gradeToNumeric are unused - preserved for future efficiency calculations
/*
// updateEfficiencyMetrics calculates and updates efficiency metrics for a job
func (c *SimplifiedJobPerformanceCollector) updateEfficiencyMetrics(job *slurm.Job) {
	// TODO: Job field names are not compatible with current slurm-client version
	// Skipping efficiency metric updates for now

	// Check efficiency cache first
	if cachedEfficiency, exists := c.efficiencyCache[job.JobID]; exists {
		c.setEfficiencyMetrics(job, cachedEfficiency)
		return
	}

	// Create resource utilization data from job
	utilizationData := CreateResourceUtilizationDataFromJob(job)

	// Calculate efficiency metrics
	efficiencyMetrics, err := c.efficiencyCalc.CalculateEfficiency(utilizationData)
	if err != nil {
		c.logger.Warn("Failed to calculate efficiency metrics", "job_id", job.JobID, "error", err)
		return
	}

	// Cache the efficiency metrics
	c.efficiencyCache[job.JobID] = efficiencyMetrics

	// Update Prometheus metrics
	c.setEfficiencyMetrics(job, efficiencyMetrics)

}

// setEfficiencyMetrics sets Prometheus metrics from efficiency calculations
func (c *SimplifiedJobPerformanceCollector) setEfficiencyMetrics(job *slurm.Job, effMetrics *EfficiencyMetrics) {
	// TODO: Job field names are not compatible with current slurm-client version
	// Skipping metric updates for now

	labels := []string{
		job.JobID,
		job.Name,
		job.UserName,
		job.Account,
		job.Partition,
		job.JobState,
	}

	// Set individual efficiency scores
	c.metrics.JobCPUEfficiencyScore.WithLabelValues(labels...).Set(effMetrics.CPUEfficiency)
	c.metrics.JobMemoryEfficiencyScore.WithLabelValues(labels...).Set(effMetrics.MemoryEfficiency)
	c.metrics.JobOverallEfficiencyScore.WithLabelValues(labels...).Set(effMetrics.OverallEfficiency)

	// Set waste ratio
	c.metrics.JobWasteRatio.WithLabelValues(labels...).Set(effMetrics.WasteRatio)

	// Set efficiency grade as numeric value
	gradeLabels := append(labels, effMetrics.EfficiencyGrade)
	gradeValue := c.gradeToNumeric(effMetrics.EfficiencyGrade)
	c.metrics.JobEfficiencyGrade.WithLabelValues(gradeLabels...).Set(gradeValue)

}

// gradeToNumeric converts letter grade to numeric value
func (c *SimplifiedJobPerformanceCollector) gradeToNumeric(grade string) float64 {
	switch grade {
	case "A":
		return 5.0
	case "B":
		return 4.0
	case "C":
		return 3.0
	case "D":
		return 2.0
	case "F":
		return 1.0
	default:
		return 0.0
	}
}
*/

// GetLastCollection returns the timestamp of the last successful collection
func (c *SimplifiedJobPerformanceCollector) GetLastCollection() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCollection
}
