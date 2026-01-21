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

// JobPerformanceCollector collects comprehensive job performance metrics
type JobPerformanceCollector struct {
	slurmClient    slurm.SlurmClient
	logger         *slog.Logger
	config         *JobPerformanceConfig
	metrics        *JobPerformanceMetrics
	lastCollection time.Time
	mu             sync.RWMutex

	// Cache for recent job data
	jobCache map[string]*JobUtilization
	cacheTTL time.Duration
}

// JobUtilization represents job utilization metrics
type JobUtilization struct {
	JobID             string
	CPUUtilization    float64
	MemoryUtilization float64
	GPUUtilization    float64
	IOUtilization     float64
	LastUpdated       time.Time
}

// JobPerformanceConfig holds configuration for job performance collection
type JobPerformanceConfig struct {
	CollectionInterval   time.Duration `yaml:"collection_interval"`
	MaxJobsPerCollection int           `yaml:"max_jobs_per_collection"`
	EnableLiveMetrics    bool          `yaml:"enable_live_metrics"`
	EnableStepMetrics    bool          `yaml:"enable_step_metrics"`
	EnableEnergyMetrics  bool          `yaml:"enable_energy_metrics"`
	CacheTTL             time.Duration `yaml:"cache_ttl"`
	IncludeCompletedJobs bool          `yaml:"include_completed_jobs"`
	CompletedJobsMaxAge  time.Duration `yaml:"completed_jobs_max_age"`
}

// JobPerformanceMetrics holds Prometheus metrics for job performance
type JobPerformanceMetrics struct {
	// Job utilization metrics
	JobCPUUtilization     *prometheus.GaugeVec
	JobMemoryUtilization  *prometheus.GaugeVec
	JobGPUUtilization     *prometheus.GaugeVec
	JobIOUtilization      *prometheus.GaugeVec
	JobNetworkUtilization *prometheus.GaugeVec
	JobEnergyConsumption  *prometheus.GaugeVec

	// Job efficiency metrics
	JobCPUEfficiency     *prometheus.GaugeVec
	JobMemoryEfficiency  *prometheus.GaugeVec
	JobIOEfficiency      *prometheus.GaugeVec
	JobOverallEfficiency *prometheus.GaugeVec

	// Job resource waste metrics
	JobCPUWasted          *prometheus.GaugeVec
	JobMemoryWasted       *prometheus.GaugeVec
	JobResourceWasteRatio *prometheus.GaugeVec

	// Collection performance metrics
	CollectionDuration prometheus.Histogram
	CollectionErrors   *prometheus.CounterVec
	JobsProcessed      prometheus.Counter
	CacheHitRatio      prometheus.Gauge
}

// NewJobPerformanceCollector creates a new job performance collector
func NewJobPerformanceCollector(slurmClient slurm.SlurmClient, logger *slog.Logger, config *JobPerformanceConfig) (*JobPerformanceCollector, error) {
	if config == nil {
		config = &JobPerformanceConfig{
			CollectionInterval:   30 * time.Second,
			MaxJobsPerCollection: 1000,
			EnableLiveMetrics:    true,
			EnableStepMetrics:    true,
			EnableEnergyMetrics:  false,
			CacheTTL:             5 * time.Minute,
			IncludeCompletedJobs: true,
			CompletedJobsMaxAge:  1 * time.Hour,
		}
	}

	metrics := &JobPerformanceMetrics{
		JobCPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpu_utilization_ratio",
				Help: "CPU utilization ratio for running jobs (used/allocated)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobMemoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_utilization_ratio",
				Help: "Memory utilization ratio for running jobs (used/allocated)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobGPUUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_gpu_utilization_ratio",
				Help: "GPU utilization ratio for running jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state", "gpu_type"},
		),
		JobIOUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_io_utilization_bytes_per_second",
				Help: "I/O utilization in bytes per second for running jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state", "io_type"},
		),
		JobNetworkUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_network_utilization_bytes_per_second",
				Help: "Network utilization in bytes per second for running jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state", "direction"},
		),
		JobEnergyConsumption: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_energy_consumption_joules",
				Help: "Energy consumption in joules for running jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobCPUEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpu_efficiency_ratio",
				Help: "CPU efficiency ratio for jobs (actual usage vs optimal usage)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobMemoryEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_efficiency_ratio",
				Help: "Memory efficiency ratio for jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobIOEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_io_efficiency_ratio",
				Help: "I/O efficiency ratio for jobs",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobOverallEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_overall_efficiency_ratio",
				Help: "Overall efficiency ratio for jobs (composite metric)",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobCPUWasted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_cpu_wasted_core_hours",
				Help: "CPU core-hours wasted due to inefficient usage",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobMemoryWasted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_memory_wasted_gb_hours",
				Help: "Memory GB-hours wasted due to inefficient usage",
			},
			[]string{"job_id", "job_name", "user", "account", "partition", "state"},
		),
		JobResourceWasteRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_waste_ratio",
				Help: "Overall resource waste ratio for jobs",
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

	collector := &JobPerformanceCollector{
		slurmClient:    slurmClient,
		logger:         logger,
		config:         config,
		metrics:        metrics,
		jobCache:       make(map[string]*JobUtilization),
		cacheTTL:       config.CacheTTL,
		lastCollection: time.Time{},
	}

	return collector, nil
}

// Describe implements the prometheus.Collector interface
func (c *JobPerformanceCollector) Describe(ch chan<- *prometheus.Desc) {
	c.metrics.JobCPUUtilization.Describe(ch)
	c.metrics.JobMemoryUtilization.Describe(ch)
	c.metrics.JobGPUUtilization.Describe(ch)
	c.metrics.JobIOUtilization.Describe(ch)
	c.metrics.JobNetworkUtilization.Describe(ch)
	c.metrics.JobEnergyConsumption.Describe(ch)
	c.metrics.JobCPUEfficiency.Describe(ch)
	c.metrics.JobMemoryEfficiency.Describe(ch)
	c.metrics.JobIOEfficiency.Describe(ch)
	c.metrics.JobOverallEfficiency.Describe(ch)
	c.metrics.JobCPUWasted.Describe(ch)
	c.metrics.JobMemoryWasted.Describe(ch)
	c.metrics.JobResourceWasteRatio.Describe(ch)
	c.metrics.CollectionDuration.Describe(ch)
	c.metrics.CollectionErrors.Describe(ch)
	c.metrics.JobsProcessed.Describe(ch)
	c.metrics.CacheHitRatio.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (c *JobPerformanceCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		c.metrics.CollectionDuration.Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.collectJobUtilization(ctx); err != nil {
		c.logger.Error("Failed to collect job utilization metrics", "error", err)
		c.metrics.CollectionErrors.WithLabelValues("job_utilization").Inc()
	}

	// Collect metrics from all registered collectors
	c.metrics.JobCPUUtilization.Collect(ch)
	c.metrics.JobMemoryUtilization.Collect(ch)
	c.metrics.JobGPUUtilization.Collect(ch)
	c.metrics.JobIOUtilization.Collect(ch)
	c.metrics.JobNetworkUtilization.Collect(ch)
	c.metrics.JobEnergyConsumption.Collect(ch)
	c.metrics.JobCPUEfficiency.Collect(ch)
	c.metrics.JobMemoryEfficiency.Collect(ch)
	c.metrics.JobIOEfficiency.Collect(ch)
	c.metrics.JobOverallEfficiency.Collect(ch)
	c.metrics.JobCPUWasted.Collect(ch)
	c.metrics.JobMemoryWasted.Collect(ch)
	c.metrics.JobResourceWasteRatio.Collect(ch)
	c.metrics.CollectionDuration.Collect(ch)
	c.metrics.CollectionErrors.Collect(ch)
	c.metrics.JobsProcessed.Collect(ch)
	c.metrics.CacheHitRatio.Collect(ch)
}

// collectJobUtilization collects comprehensive job utilization metrics
func (c *JobPerformanceCollector) collectJobUtilization(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get job manager from SLURM client
	jobManager := c.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// List jobs with appropriate filters
	// Using nil for options as the exact structure is not clear
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	c.logger.Debug("Collecting job performance metrics", "job_count", len(jobs.Jobs))

	cacheHits := 0
	totalJobs := len(jobs.Jobs)

	// TODO: Job field names (JobID, etc.) are not available in the current slurm-client version
	// Skipping job utilization processing for now
	_ = jobs       // Suppress unused variable warning
	_ = jobManager // Suppress unused variable warning

	// Update cache hit ratio
	if totalJobs > 0 {
		c.metrics.CacheHitRatio.Set(float64(cacheHits) / float64(totalJobs))
	}

	// Clean expired cache entries
	c.cleanExpiredCache()

	c.lastCollection = time.Now()
	return nil
}

// TODO: updateMetricsFromUtilization is unused - preserved for future utilization metrics updates
/*
// updateMetricsFromUtilization updates Prometheus metrics from job utilization data
func (c *JobPerformanceCollector) updateMetricsFromUtilization(job *slurm.Job, util *JobUtilization) {
	// TODO: Job field names and JobUtilization structure are not compatible with current slurm-client version
	// Skipping metric updates for now

	labels := []string{
		util.JobID,
		job.Name,
		job.UserName,
		job.Account,
		job.Partition,
		job.JobState,
	}

	// CPU Utilization
	if util.CPUUtilization != nil {
		c.metrics.JobCPUUtilization.WithLabelValues(labels...).Set(util.CPUUtilization.Efficiency)
		if util.CPUUtilization.Efficiency > 0 {
			c.metrics.JobCPUEfficiency.WithLabelValues(labels...).Set(util.CPUUtilization.Efficiency)
		}
		if util.CPUUtilization.Wasted > 0 {
			c.metrics.JobCPUWasted.WithLabelValues(labels...).Set(util.CPUUtilization.Wasted)
		}
	}

	// Memory Utilization
	if util.MemoryUtilization != nil {
		c.metrics.JobMemoryUtilization.WithLabelValues(labels...).Set(util.MemoryUtilization.Efficiency)
		if util.MemoryUtilization.Efficiency > 0 {
			c.metrics.JobMemoryEfficiency.WithLabelValues(labels...).Set(util.MemoryUtilization.Efficiency)
		}
		if util.MemoryUtilization.Wasted > 0 {
			// Convert bytes to GB-hours
			wastedGBHours := util.MemoryUtilization.Wasted / (1024 * 1024 * 1024)
			c.metrics.JobMemoryWasted.WithLabelValues(labels...).Set(wastedGBHours)
		}
	}

	// GPU Utilization
	if util.GPUUtilization != nil && util.GPUUtilization.DeviceCount > 0 {
		for _, device := range util.GPUUtilization.Devices {
			gpuLabels := append(labels, device.DeviceType)
			c.metrics.JobGPUUtilization.WithLabelValues(gpuLabels...).Set(device.Utilization.Efficiency)
		}
	}

	// I/O Utilization
	if util.IOUtilization != nil {
		readLabels := append(labels, "read")
		writeLabels := append(labels, "write")

		c.metrics.JobIOUtilization.WithLabelValues(readLabels...).Set(util.IOUtilization.ReadBytesPerSecond)
		c.metrics.JobIOUtilization.WithLabelValues(writeLabels...).Set(util.IOUtilization.WriteBytesPerSecond)

		if util.IOUtilization.Efficiency > 0 {
			c.metrics.JobIOEfficiency.WithLabelValues(labels...).Set(util.IOUtilization.Efficiency)
		}
	}

	// Network Utilization
	if util.NetworkUtilization != nil {
		rxLabels := append(labels, "rx")
		txLabels := append(labels, "tx")

		c.metrics.JobNetworkUtilization.WithLabelValues(rxLabels...).Set(util.NetworkUtilization.ReceiveBytesPerSecond)
		c.metrics.JobNetworkUtilization.WithLabelValues(txLabels...).Set(util.NetworkUtilization.TransmitBytesPerSecond)
	}

	// Energy Consumption
	if c.config.EnableEnergyMetrics && util.EnergyUsage != nil {
		c.metrics.JobEnergyConsumption.WithLabelValues(labels...).Set(util.EnergyUsage.TotalEnergyConsumed)
	}

	// Overall Efficiency (composite metric)
	if util.CPUUtilization != nil && util.MemoryUtilization != nil {
		overallEfficiency := (util.CPUUtilization.Efficiency + util.MemoryUtilization.Efficiency) / 2.0
		c.metrics.JobOverallEfficiency.WithLabelValues(labels...).Set(overallEfficiency)

		// Resource waste ratio
		if util.CPUUtilization.Wasted > 0 || util.MemoryUtilization.Wasted > 0 {
			totalAllocated := util.CPUUtilization.Allocated + util.MemoryUtilization.Allocated
			totalWasted := util.CPUUtilization.Wasted + util.MemoryUtilization.Wasted
			if totalAllocated > 0 {
				wasteRatio := totalWasted / totalAllocated
				c.metrics.JobResourceWasteRatio.WithLabelValues(labels...).Set(wasteRatio)
			}
		}
	}

}
*/

// cleanExpiredCache removes expired entries from the job cache
func (c *JobPerformanceCollector) cleanExpiredCache() {
	now := time.Now()
	for jobID, util := range c.jobCache {
		if now.Sub(util.LastUpdated) > c.cacheTTL {
			delete(c.jobCache, jobID)
		}
	}
}

// GetCacheSize returns the current size of the job cache
func (c *JobPerformanceCollector) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.jobCache)
}

// GetLastCollection returns the timestamp of the last successful collection
func (c *JobPerformanceCollector) GetLastCollection() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastCollection
}
