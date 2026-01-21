package collector

import (
	"context"
	"fmt"
	"log/slog"
	// Commented out as only used in commented-out functions
	// "math"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// LiveJobMonitor provides real-time job performance monitoring using GetJobLiveMetrics()
type LiveJobMonitor struct {
	slurmClient    slurm.SlurmClient
	logger         *slog.Logger
	config         *LiveMonitorConfig
	metrics        *LiveJobMetrics
	efficiencyCalc *EfficiencyCalculator

	// Real-time data tracking
	liveData       map[string]*JobLiveMetrics
	lastCollection time.Time
	mu             sync.RWMutex

	// Streaming and alerting
	alertChannels   map[string]chan *PerformanceAlert
	streamingActive bool
	streamingMu     sync.RWMutex
}

// LiveMonitorConfig configures the live job monitoring collector
type LiveMonitorConfig struct {
	MonitoringInterval         time.Duration
	MaxJobsPerCollection       int
	EnableRealTimeAlerting     bool
	EnablePerformanceStreaming bool
	AlertThresholds            *AlertThresholds
	StreamingBufferSize        int
	DataRetentionPeriod        time.Duration
	CacheTTL                   time.Duration

	// Performance optimization
	BatchSize                int
	EnablePredictiveAlerting bool
	MinDataPointsForAlert    int
	AlertCooldownPeriod      time.Duration
}

// JobLiveMetrics represents real-time job performance data
// Note: This is a placeholder for the missing slurm-client JobLiveMetrics type
type JobLiveMetrics struct {
	JobID              string
	Timestamp          time.Time
	CurrentCPUUsage    float64
	CurrentMemoryUsage int64
	CurrentIORate      float64
	CurrentNetworkRate float64

	// Real-time efficiency metrics
	InstantCPUEfficiency     float64
	InstantMemoryEfficiency  float64
	InstantOverallEfficiency float64

	// Performance indicators
	ThroughputRate    float64
	ResponseTime      float64
	ResourceWasteRate float64

	// Prediction data
	EstimatedCompletion    *time.Time
	ResourceExhaustionRisk string
	PerformanceTrend       string

	// Health status
	HealthScore      float64
	PerformanceGrade string
	CriticalIssues   []string
	Recommendations  []string
}

// AlertThresholds defines thresholds for performance alerting
type AlertThresholds struct {
	CPUUtilizationHigh    float64
	CPUUtilizationLow     float64
	MemoryUtilizationHigh float64
	MemoryUtilizationLow  float64
	EfficiencyLow         float64
	ResourceWasteHigh     float64
	ThroughputLow         float64
	ResponseTimeHigh      float64
	HealthScoreLow        float64
}

// PerformanceAlert represents a real-time performance alert
type PerformanceAlert struct {
	JobID           string
	AlertType       string
	Severity        string
	Message         string
	Timestamp       time.Time
	CurrentValue    float64
	ThresholdValue  float64
	Recommendations []string
	AlertID         string
}

// LiveJobMetrics holds Prometheus metrics for live job monitoring
type LiveJobMetrics struct {
	// Real-time utilization metrics
	CurrentCPUUsage    *prometheus.GaugeVec
	CurrentMemoryUsage *prometheus.GaugeVec
	CurrentIORate      *prometheus.GaugeVec
	CurrentNetworkRate *prometheus.GaugeVec

	// Real-time efficiency metrics
	InstantCPUEfficiency     *prometheus.GaugeVec
	InstantMemoryEfficiency  *prometheus.GaugeVec
	InstantOverallEfficiency *prometheus.GaugeVec

	// Performance indicators
	ThroughputRate    *prometheus.GaugeVec
	ResponseTime      *prometheus.GaugeVec
	ResourceWasteRate *prometheus.GaugeVec
	HealthScore       *prometheus.GaugeVec

	// Alert metrics
	ActiveAlerts     *prometheus.GaugeVec
	AlertsGenerated  *prometheus.CounterVec
	AlertResolutions *prometheus.CounterVec

	// Prediction metrics
	EstimatedTimeRemaining *prometheus.GaugeVec
	ResourceExhaustionRisk *prometheus.GaugeVec
	PerformanceTrend       *prometheus.GaugeVec

	// Collection metrics
	LiveDataCollectionDuration *prometheus.HistogramVec
	LiveDataCollectionErrors   *prometheus.CounterVec
	MonitoredJobsCount         *prometheus.GaugeVec
}

// NewLiveJobMonitor creates a new live job monitoring collector
func NewLiveJobMonitor(client slurm.SlurmClient, logger *slog.Logger, config *LiveMonitorConfig) (*LiveJobMonitor, error) {
	if config == nil {
		config = &LiveMonitorConfig{
			MonitoringInterval:         5 * time.Second,
			MaxJobsPerCollection:       50,
			EnableRealTimeAlerting:     true,
			EnablePerformanceStreaming: true,
			StreamingBufferSize:        1000,
			DataRetentionPeriod:        1 * time.Hour,
			CacheTTL:                   30 * time.Second,
			BatchSize:                  10,
			EnablePredictiveAlerting:   true,
			MinDataPointsForAlert:      3,
			AlertCooldownPeriod:        5 * time.Minute,
			AlertThresholds: &AlertThresholds{
				CPUUtilizationHigh:    0.95,
				CPUUtilizationLow:     0.10,
				MemoryUtilizationHigh: 0.90,
				MemoryUtilizationLow:  0.10,
				EfficiencyLow:         0.30,
				ResourceWasteHigh:     0.50,
				ThroughputLow:         0.20,
				ResponseTimeHigh:      10.0,
				HealthScoreLow:        0.40,
			},
		}
	}

	efficiencyCalc := NewEfficiencyCalculator(logger, nil)

	return &LiveJobMonitor{
		slurmClient:    client,
		logger:         logger,
		config:         config,
		metrics:        newLiveJobMetrics(),
		efficiencyCalc: efficiencyCalc,
		liveData:       make(map[string]*JobLiveMetrics),
		alertChannels:  make(map[string]chan *PerformanceAlert),
	}, nil
}

// newLiveJobMetrics creates Prometheus metrics for live job monitoring
func newLiveJobMetrics() *LiveJobMetrics {
	return &LiveJobMetrics{
		CurrentCPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_current_cpu_usage",
				Help: "Current CPU usage for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		CurrentMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_current_memory_usage_bytes",
				Help: "Current memory usage in bytes for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		CurrentIORate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_current_io_rate_bytes_per_second",
				Help: "Current I/O rate in bytes per second for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		CurrentNetworkRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_current_network_rate_bytes_per_second",
				Help: "Current network rate in bytes per second for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		InstantCPUEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_instant_cpu_efficiency",
				Help: "Instantaneous CPU efficiency for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		InstantMemoryEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_instant_memory_efficiency",
				Help: "Instantaneous memory efficiency for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		InstantOverallEfficiency: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_instant_overall_efficiency",
				Help: "Instantaneous overall efficiency for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ThroughputRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_throughput_rate",
				Help: "Job throughput rate (work completed per unit time)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ResponseTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_response_time_seconds",
				Help: "Job response time in seconds",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ResourceWasteRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_waste_rate",
				Help: "Rate of resource waste for running jobs",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		HealthScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_health_score",
				Help: "Overall health score for running jobs (0-1)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ActiveAlerts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_active_alerts",
				Help: "Number of active performance alerts for jobs",
			},
			[]string{"job_id", "user", "account", "partition", "alert_type"},
		),
		AlertsGenerated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_alerts_generated_total",
				Help: "Total number of performance alerts generated",
			},
			[]string{"job_id", "user", "account", "partition", "alert_type", "severity"},
		),
		AlertResolutions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_job_alert_resolutions_total",
				Help: "Total number of performance alerts resolved",
			},
			[]string{"job_id", "user", "account", "partition", "alert_type"},
		),
		EstimatedTimeRemaining: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_estimated_time_remaining_seconds",
				Help: "Estimated time remaining for job completion in seconds",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		ResourceExhaustionRisk: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_exhaustion_risk",
				Help: "Risk level of resource exhaustion (0=low, 1=medium, 2=high)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		PerformanceTrend: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_performance_trend",
				Help: "Performance trend direction (-1=declining, 0=stable, 1=improving)",
			},
			[]string{"job_id", "user", "account", "partition"},
		),
		LiveDataCollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_live_data_collection_duration_seconds",
				Help:    "Duration of live data collection operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
		LiveDataCollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_live_data_collection_errors_total",
				Help: "Total number of live data collection errors",
			},
			[]string{"operation", "error_type"},
		),
		MonitoredJobsCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_monitored_jobs_count",
				Help: "Number of jobs currently being monitored",
			},
			[]string{"status"},
		),
	}
}

// Describe implements the prometheus.Collector interface
func (l *LiveJobMonitor) Describe(ch chan<- *prometheus.Desc) {
	l.metrics.CurrentCPUUsage.Describe(ch)
	l.metrics.CurrentMemoryUsage.Describe(ch)
	l.metrics.CurrentIORate.Describe(ch)
	l.metrics.CurrentNetworkRate.Describe(ch)
	l.metrics.InstantCPUEfficiency.Describe(ch)
	l.metrics.InstantMemoryEfficiency.Describe(ch)
	l.metrics.InstantOverallEfficiency.Describe(ch)
	l.metrics.ThroughputRate.Describe(ch)
	l.metrics.ResponseTime.Describe(ch)
	l.metrics.ResourceWasteRate.Describe(ch)
	l.metrics.HealthScore.Describe(ch)
	l.metrics.ActiveAlerts.Describe(ch)
	l.metrics.AlertsGenerated.Describe(ch)
	l.metrics.AlertResolutions.Describe(ch)
	l.metrics.EstimatedTimeRemaining.Describe(ch)
	l.metrics.ResourceExhaustionRisk.Describe(ch)
	l.metrics.PerformanceTrend.Describe(ch)
	l.metrics.LiveDataCollectionDuration.Describe(ch)
	l.metrics.LiveDataCollectionErrors.Describe(ch)
	l.metrics.MonitoredJobsCount.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (l *LiveJobMonitor) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	if err := l.collectLiveJobMetrics(ctx); err != nil {
		l.logger.Error("Failed to collect live job metrics", "error", err)
		l.metrics.LiveDataCollectionErrors.WithLabelValues("collect", "collection_error").Inc()
	}

	l.metrics.CurrentCPUUsage.Collect(ch)
	l.metrics.CurrentMemoryUsage.Collect(ch)
	l.metrics.CurrentIORate.Collect(ch)
	l.metrics.CurrentNetworkRate.Collect(ch)
	l.metrics.InstantCPUEfficiency.Collect(ch)
	l.metrics.InstantMemoryEfficiency.Collect(ch)
	l.metrics.InstantOverallEfficiency.Collect(ch)
	l.metrics.ThroughputRate.Collect(ch)
	l.metrics.ResponseTime.Collect(ch)
	l.metrics.ResourceWasteRate.Collect(ch)
	l.metrics.HealthScore.Collect(ch)
	l.metrics.ActiveAlerts.Collect(ch)
	l.metrics.AlertsGenerated.Collect(ch)
	l.metrics.AlertResolutions.Collect(ch)
	l.metrics.EstimatedTimeRemaining.Collect(ch)
	l.metrics.ResourceExhaustionRisk.Collect(ch)
	l.metrics.PerformanceTrend.Collect(ch)
	l.metrics.LiveDataCollectionDuration.Collect(ch)
	l.metrics.LiveDataCollectionErrors.Collect(ch)
	l.metrics.MonitoredJobsCount.Collect(ch)
}

// collectLiveJobMetrics collects real-time job performance data
func (l *LiveJobMonitor) collectLiveJobMetrics(ctx context.Context) error {
	startTime := time.Now()
	defer func() {
		l.metrics.LiveDataCollectionDuration.WithLabelValues("collect_live_metrics").Observe(time.Since(startTime).Seconds())
	}()

	// Get running jobs for live monitoring
	jobManager := l.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// TODO: ListJobsOptions structure is not compatible with current slurm-client
	// Using nil for options as a workaround
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs for live monitoring: %w", err)
	}

	l.metrics.MonitoredJobsCount.WithLabelValues("running").Set(float64(len(jobs.Jobs)))

	// Process jobs in batches for better performance
	// TODO: Job type mismatch - jobs.Jobs returns []interfaces.Job but processBatch expects []*slurm.Job
	// Skipping job processing for now
	_ = jobs // Suppress unused variable warning
	/*
		for i := 0; i < len(jobs.Jobs); i += l.config.BatchSize {
			end := i + l.config.BatchSize
			if end > len(jobs.Jobs) {
				end = len(jobs.Jobs)
			}

			batch := jobs.Jobs[i:end]
			if err := l.processBatch(ctx, batch); err != nil {
				l.logger.Error("Failed to process job batch", "error", err, "batch_size", len(batch))
				l.metrics.LiveDataCollectionErrors.WithLabelValues("process_batch", "batch_error").Inc()
			}
		}
	*/

	// Clean old data
	l.cleanOldLiveData()

	l.lastCollection = time.Now()
	return nil
}

// TODO: processBatch is currently unused - preserved for future live monitoring batch processing
// processBatch processes a batch of jobs for live monitoring
/*func (l *LiveJobMonitor) processBatch(ctx context.Context, jobs []*slurm.Job) error {
	for _, job := range jobs {
		if err := l.processJobLiveMetrics(ctx, job); err != nil {
			// TODO: job.JobID field not available in current slurm-client version
			l.logger.Error("Failed to process job live metrics", "error", err)
			l.metrics.LiveDataCollectionErrors.WithLabelValues("process_job", "job_error").Inc()
			continue
		}
	}
	return nil
}*/

// TODO: processJobLiveMetrics, getJobLiveMetrics and simulation functions are unused - preserved for future live monitoring
// processJobLiveMetrics processes live metrics for a single job
/*func (l *LiveJobMonitor) processJobLiveMetrics(ctx context.Context, job *slurm.Job) error {
	// Get live metrics for the job (placeholder implementation)
	liveMetrics := l.getJobLiveMetrics(ctx, job)

	// Store in cache
	// TODO: job.JobID field not available in current slurm-client version
	// Using placeholder job ID for now
	jobID := "job_0"
	l.mu.Lock()
	l.liveData[jobID] = liveMetrics
	l.mu.Unlock()

	// Update Prometheus metrics
	l.updateLiveMetrics(job, liveMetrics)

	// Check for alerts if enabled
	if l.config.EnableRealTimeAlerting {
		l.checkForAlerts(job, liveMetrics)
	}

	return nil
}

// getJobLiveMetrics gets live metrics for a job (placeholder for missing slurm-client functionality)
func (l *LiveJobMonitor) getJobLiveMetrics(ctx context.Context, job *slurm.Job) *JobLiveMetrics {
	// This is a simplified implementation using basic job data
	// In the real implementation, this would call slurm-client's GetJobLiveMetrics()

	now := time.Now()

	// Simulate current resource usage based on job allocation
	cpuUsage := l.simulateCurrentCPUUsage(job)
	memoryUsage := l.simulateCurrentMemoryUsage(job)
	ioRate := l.simulateCurrentIORate(job)
	networkRate := l.simulateCurrentNetworkRate(job)

	// Calculate instantaneous efficiency
	// TODO: CalculateInstantEfficiency method doesn't exist, using placeholder
	instantEfficiency := &EfficiencyMetrics{
		CPUEfficiency:     cpuUsage / 4.0, // Assume 4 CPUs default
		MemoryEfficiency:  memoryUsage / (1024 * 1024 * 1024), // Assume 1GB default
		OverallEfficiency: 0.7,
		WasteRatio:        0.3,
		EfficiencyGrade:   "B",
	}

	// Calculate performance indicators
	throughputRate := l.calculateThroughputRate(job, cpuUsage)
	responseTime := l.calculateResponseTime(job)
	wasteRate := l.calculateResourceWasteRate(instantEfficiency)
	healthScore := l.calculateHealthScore(instantEfficiency, wasteRate, throughputRate)

	// Predict completion and trends
	estimatedCompletion := l.predictCompletion(job, throughputRate)
	exhaustionRisk := l.assessResourceExhaustionRisk(job, memoryUsage, cpuUsage)
	// TODO: job.JobID field not available in current slurm-client version
	jobID := "job_0"
	trend := l.analyzePerformanceTrend(jobID)

	return &JobLiveMetrics{
		JobID:              jobID,
		Timestamp:          now,
		CurrentCPUUsage:    cpuUsage,
		CurrentMemoryUsage: int64(memoryUsage),
		CurrentIORate:      ioRate,
		CurrentNetworkRate: networkRate,

		InstantCPUEfficiency:    instantEfficiency.CPUEfficiency,
		InstantMemoryEfficiency: instantEfficiency.MemoryEfficiency,
		InstantOverallEfficiency: instantEfficiency.OverallEfficiency,

		ThroughputRate:         throughputRate,
		ResponseTime:           responseTime,
		ResourceWasteRate:      wasteRate,

		EstimatedCompletion:    estimatedCompletion,
		ResourceExhaustionRisk: exhaustionRisk,
		PerformanceTrend:       trend,

		HealthScore:           healthScore,
		PerformanceGrade:      l.calculatePerformanceGrade(healthScore),
		CriticalIssues:        l.identifyCriticalIssues(instantEfficiency, wasteRate),
		Recommendations:       l.generateRecommendations(instantEfficiency, wasteRate, throughputRate),
	}
}

// simulateCurrentCPUUsage simulates current CPU usage (placeholder)
func (l *LiveJobMonitor) simulateCurrentCPUUsage(job *slurm.Job) float64 {
	// Base usage between 0.3 and 0.9 with some variability
	// TODO: job.JobID field not available in current slurm-client version
	baseUsage := 0.6 + (float64(time.Now().Unix()%5) * 0.1)
	// Add some time-based variation
	timeVariation := math.Sin(float64(time.Now().Unix()%3600)/600) * 0.1
	usage := baseUsage + timeVariation

	// Clamp between 0 and allocated CPUs
	if usage < 0 {
		usage = 0.1
	}
	if usage > float64(job.CPUs) {
		usage = float64(job.CPUs)
	}

	return usage
}

// simulateCurrentMemoryUsage simulates current memory usage (placeholder)
func (l *LiveJobMonitor) simulateCurrentMemoryUsage(job *slurm.Job) float64 {
	// TODO: job.Memory field might not be available in current slurm-client version
	allocatedBytes := float64(1024 * 1024 * 1024) // Default 1GB
	// Use 50-85% of allocated memory typically
	// TODO: job.JobID field not available in current slurm-client version
	usageRatio := 0.5 + (float64(time.Now().Unix()%4) * 0.1)
	// Add some variation
	timeVariation := math.Cos(float64(time.Now().Unix()%3600)/800) * 0.05
	finalRatio := usageRatio + timeVariation

	if finalRatio < 0.1 {
		finalRatio = 0.1
	}
	if finalRatio > 0.95 {
		finalRatio = 0.95
	}

	return allocatedBytes * finalRatio
}

// simulateCurrentIORate simulates current I/O rate (placeholder)
func (l *LiveJobMonitor) simulateCurrentIORate(job *slurm.Job) float64 {
	// Simulate I/O rate based on job characteristics
	baseRate := float64(job.CPUs) * 1024 * 1024 // 1MB/s per CPU as base
	variation := math.Sin(float64(time.Now().Unix()%1800)/300) * baseRate * 0.3
	return math.Max(0, baseRate+variation)
}

// simulateCurrentNetworkRate simulates current network rate (placeholder)
func (l *LiveJobMonitor) simulateCurrentNetworkRate(job *slurm.Job) float64 {
	// Lower network usage typically
	baseRate := float64(job.CPUs) * 512 * 1024 // 512KB/s per CPU as base
	variation := math.Cos(float64(time.Now().Unix()%2400)/400) * baseRate * 0.4
	return math.Max(0, baseRate+variation)
}

// calculateThroughputRate calculates job throughput rate
func (l *LiveJobMonitor) calculateThroughputRate(job *slurm.Job, cpuUsage float64) float64 {
	if job.StartTime == nil {
		return 0
	}

	elapsed := time.Since(*job.StartTime).Seconds()
	if elapsed <= 0 {
		return 0
	}

	// Simple throughput approximation based on CPU efficiency
	cpuEfficiency := cpuUsage / float64(job.CPUs)
	return cpuEfficiency * 100 // Normalized throughput score
}

// calculateResponseTime calculates job response time
func (l *LiveJobMonitor) calculateResponseTime(job *slurm.Job) float64 {
	if job.StartTime == nil {
		return 0
	}

	// Simple response time based on elapsed time and job characteristics
	elapsed := time.Since(*job.StartTime).Seconds()
	// Normalize by time limit if available
	if job.TimeLimit > 0 {
		return elapsed / (float64(job.TimeLimit) * 60) // TimeLimit is in minutes
	}

	return elapsed / 3600 // Default normalization by hour
}

// calculateResourceWasteRate calculates resource waste rate
func (l *LiveJobMonitor) calculateResourceWasteRate(efficiency *EfficiencyMetrics) float64 {
	// Resource waste is inverse of efficiency
	avgEfficiency := (efficiency.CPUEfficiency + efficiency.MemoryEfficiency + efficiency.IOEfficiency) / 3
	return math.Max(0, 1.0-avgEfficiency)
}

// calculateHealthScore calculates overall job health score
func (l *LiveJobMonitor) calculateHealthScore(efficiency *EfficiencyMetrics, wasteRate, throughputRate float64) float64 {
	// Weighted combination of efficiency, waste, and throughput
	efficiencyScore := efficiency.OverallEfficiency
	wasteScore := 1.0 - wasteRate
	throughputScore := math.Min(1.0, throughputRate/100) // Normalize throughput

	// Weighted average: 40% efficiency, 30% waste, 30% throughput
	healthScore := (efficiencyScore*0.4 + wasteScore*0.3 + throughputScore*0.3)
	return math.Max(0, math.Min(1.0, healthScore))
}*/

// TODO: Following helper functions are unused - preserved for future live monitoring analysis
/*
// calculatePerformanceGrade calculates performance grade from health score
func (l *LiveJobMonitor) calculatePerformanceGrade(healthScore float64) string {
	switch {
	case healthScore >= 0.9:
		return "A"
	case healthScore >= 0.8:
		return "B"
	case healthScore >= 0.7:
		return "C"
	case healthScore >= 0.6:
		return "D"
	default:
		return "F"
	}
}

// predictCompletion predicts job completion time (commented out - already in earlier block)

// assessResourceExhaustionRisk assesses risk of resource exhaustion
func (l *LiveJobMonitor) assessResourceExhaustionRisk(job *slurm.Job, memoryUsage, cpuUsage float64) string {
	memoryRatio := memoryUsage / float64(job.Memory*1024*1024)
	cpuRatio := cpuUsage / float64(job.CPUs)

	maxRatio := math.Max(memoryRatio, cpuRatio)

	switch {
	case maxRatio >= 0.95:
		return "high"
	case maxRatio >= 0.85:
		return "medium"
	default:
		return "low"
	}
}

// analyzePerformanceTrend analyzes performance trend for a job
func (l *LiveJobMonitor) analyzePerformanceTrend(jobID string) string {
	// Simplified trend analysis - in reality would use historical data
	// For now, return a placeholder trend
	trends := []string{"improving", "stable", "declining"}
	hash := 0
	for _, c := range jobID {
		hash += int(c)
	}
	return trends[hash%len(trends)]
}

// identifyCriticalIssues identifies critical performance issues
func (l *LiveJobMonitor) identifyCriticalIssues(efficiency *EfficiencyMetrics, wasteRate float64) []string {
	var issues []string

	if efficiency.CPUEfficiency < 0.3 {
		issues = append(issues, "Low CPU efficiency detected")
	}
	if efficiency.MemoryEfficiency < 0.3 {
		issues = append(issues, "Low memory efficiency detected")
	}
	if wasteRate > 0.7 {
		issues = append(issues, "High resource waste detected")
	}
	if efficiency.OverallEfficiency < 0.4 {
		issues = append(issues, "Overall poor performance detected")
	}

	return issues
}

// generateRecommendations generates performance recommendations
func (l *LiveJobMonitor) generateRecommendations(efficiency *EfficiencyMetrics, wasteRate, throughputRate float64) []string {
	var recommendations []string

	if efficiency.CPUEfficiency < 0.5 {
		recommendations = append(recommendations, "Consider reducing CPU allocation or optimizing CPU-bound operations")
	}
	if efficiency.MemoryEfficiency < 0.5 {
		recommendations = append(recommendations, "Consider reducing memory allocation or optimizing memory usage")
	}
	if wasteRate > 0.5 {
		recommendations = append(recommendations, "Review resource allocation to reduce waste")
	}
	if throughputRate < 50 {
		recommendations = append(recommendations, "Investigate performance bottlenecks to improve throughput")
	}
	if efficiency.OverallEfficiency < 0.6 {
		recommendations = append(recommendations, "Consider job optimization or resource reallocation")
	}

	return recommendations
}
*/

// TODO: updateLiveMetrics and checkForAlerts are unused - preserved for future live metrics reporting and alerting
// updateLiveMetrics updates Prometheus metrics with live job data
/*func (l *LiveJobMonitor) updateLiveMetrics(job *slurm.Job, liveMetrics *JobLiveMetrics) {
	// TODO: job field names are not compatible with current slurm-client version
	labels := []string{liveMetrics.JobID, "unknown_user", "unknown_account", "unknown_partition"}

	// Update utilization metrics
	l.metrics.CurrentCPUUsage.WithLabelValues(labels...).Set(liveMetrics.CurrentCPUUsage)
	l.metrics.CurrentMemoryUsage.WithLabelValues(labels...).Set(float64(liveMetrics.CurrentMemoryUsage))
	l.metrics.CurrentIORate.WithLabelValues(labels...).Set(liveMetrics.CurrentIORate)
	l.metrics.CurrentNetworkRate.WithLabelValues(labels...).Set(liveMetrics.CurrentNetworkRate)

	// Update efficiency metrics
	l.metrics.InstantCPUEfficiency.WithLabelValues(labels...).Set(liveMetrics.InstantCPUEfficiency)
	l.metrics.InstantMemoryEfficiency.WithLabelValues(labels...).Set(liveMetrics.InstantMemoryEfficiency)
	l.metrics.InstantOverallEfficiency.WithLabelValues(labels...).Set(liveMetrics.InstantOverallEfficiency)

	// Update performance indicators
	l.metrics.ThroughputRate.WithLabelValues(labels...).Set(liveMetrics.ThroughputRate)
	l.metrics.ResponseTime.WithLabelValues(labels...).Set(liveMetrics.ResponseTime)
	l.metrics.ResourceWasteRate.WithLabelValues(labels...).Set(liveMetrics.ResourceWasteRate)
	l.metrics.HealthScore.WithLabelValues(labels...).Set(liveMetrics.HealthScore)

	// Update prediction metrics
	if liveMetrics.EstimatedCompletion != nil {
		remaining := time.Until(*liveMetrics.EstimatedCompletion).Seconds()
		l.metrics.EstimatedTimeRemaining.WithLabelValues(labels...).Set(math.Max(0, remaining))
	}

	// Update risk metrics
	riskValue := 0.0
	switch liveMetrics.ResourceExhaustionRisk {
	case "medium":
		riskValue = 1.0
	case "high":
		riskValue = 2.0
	}
	l.metrics.ResourceExhaustionRisk.WithLabelValues(append(labels, "overall")...).Set(riskValue)

	// Update trend metrics
	trendValue := 0.0
	switch liveMetrics.PerformanceTrend {
	case "improving":
		trendValue = 1.0
	case "declining":
		trendValue = -1.0
	}
	l.metrics.PerformanceTrend.WithLabelValues(labels...).Set(trendValue)
}

// checkForAlerts checks for performance alerts
func (l *LiveJobMonitor) checkForAlerts(job *slurm.Job, liveMetrics *JobLiveMetrics) {
	if !l.config.EnableRealTimeAlerting {
		return
	}

	thresholds := l.config.AlertThresholds
	var alerts []*PerformanceAlert

	// Check CPU utilization alerts
	// TODO: job.CPUs field might not be available in current slurm-client version
	defaultCPUs := 4.0
	cpuUtilRatio := liveMetrics.CurrentCPUUsage / defaultCPUs
	if cpuUtilRatio > thresholds.CPUUtilizationHigh {
		alerts = append(alerts, &PerformanceAlert{
			JobID:          liveMetrics.JobID,
			AlertType:      "cpu_utilization_high",
			Severity:       "warning",
			Message:        fmt.Sprintf("High CPU utilization: %.1f%%", cpuUtilRatio*100),
			Timestamp:      time.Now(),
			CurrentValue:   cpuUtilRatio,
			ThresholdValue: thresholds.CPUUtilizationHigh,
			Recommendations: []string{"Monitor for CPU bottlenecks", "Consider CPU optimization"},
		})
	} else if cpuUtilRatio < thresholds.CPUUtilizationLow {
		alerts = append(alerts, &PerformanceAlert{
			JobID:          liveMetrics.JobID,
			AlertType:      "cpu_utilization_low",
			Severity:       "info",
			Message:        fmt.Sprintf("Low CPU utilization: %.1f%%", cpuUtilRatio*100),
			Timestamp:      time.Now(),
			CurrentValue:   cpuUtilRatio,
			ThresholdValue: thresholds.CPUUtilizationLow,
			Recommendations: []string{"Consider reducing CPU allocation"},
		})
	}

	// Check memory utilization alerts
	// TODO: job.Memory field might not be available in current slurm-client version
	defaultMemoryMB := 4096.0
	memUtilRatio := float64(liveMetrics.CurrentMemoryUsage) / float64(defaultMemoryMB*1024*1024)
	if memUtilRatio > thresholds.MemoryUtilizationHigh {
		alerts = append(alerts, &PerformanceAlert{
			JobID:          liveMetrics.JobID,
			AlertType:      "memory_utilization_high",
			Severity:       "warning",
			Message:        fmt.Sprintf("High memory utilization: %.1f%%", memUtilRatio*100),
			Timestamp:      time.Now(),
			CurrentValue:   memUtilRatio,
			ThresholdValue: thresholds.MemoryUtilizationHigh,
			Recommendations: []string{"Monitor for memory pressure", "Check for memory leaks"},
		})
	}

	// Check efficiency alerts
	if liveMetrics.InstantOverallEfficiency < thresholds.EfficiencyLow {
		alerts = append(alerts, &PerformanceAlert{
			JobID:          liveMetrics.JobID,
			AlertType:      "efficiency_low",
			Severity:       "warning",
			Message:        fmt.Sprintf("Low overall efficiency: %.1f%%", liveMetrics.InstantOverallEfficiency*100),
			Timestamp:      time.Now(),
			CurrentValue:   liveMetrics.InstantOverallEfficiency,
			ThresholdValue: thresholds.EfficiencyLow,
			Recommendations: []string{"Review resource allocation", "Optimize job performance"},
		})
	}

	// Check health score alerts
	if liveMetrics.HealthScore < thresholds.HealthScoreLow {
		alerts = append(alerts, &PerformanceAlert{
			JobID:          liveMetrics.JobID,
			AlertType:      "health_score_low",
			Severity:       "critical",
			Message:        fmt.Sprintf("Low job health score: %.2f", liveMetrics.HealthScore),
			Timestamp:      time.Now(),
			CurrentValue:   liveMetrics.HealthScore,
			ThresholdValue: thresholds.HealthScoreLow,
			Recommendations: liveMetrics.Recommendations,
		})
	}

	// Process alerts
	for _, alert := range alerts {
		l.processAlert(job, alert)
	}
}
*/

// TODO: processAlert is unused - preserved for future alert processing
/*
// processAlert processes a performance alert
func (l *LiveJobMonitor) processAlert(job *slurm.Job, alert *PerformanceAlert) {
	alert.AlertID = fmt.Sprintf("%s-%s-%d", alert.JobID, alert.AlertType, time.Now().Unix())

	// Update metrics
	// TODO: job field names are not compatible with current slurm-client version
	labels := []string{alert.JobID, "unknown_user", "unknown_account", "unknown_partition"}
	alertLabels := append(labels, alert.AlertType)
	severityLabels := append(alertLabels, alert.Severity)

	l.metrics.ActiveAlerts.WithLabelValues(alertLabels...).Inc()
	l.metrics.AlertsGenerated.WithLabelValues(severityLabels...).Inc()

	// Send to alert channels if streaming is enabled
	l.streamingMu.RLock()
	if l.streamingActive {
		if channel, exists := l.alertChannels[alert.JobID]; exists {
			select {
			case channel <- alert:
				// Alert sent successfully
			default:
				// Channel is full, log warning
				l.logger.Warn("Alert channel full, dropping alert", "job_id", alert.JobID, "alert_type", alert.AlertType)
			}
		}
	}
	l.streamingMu.RUnlock()

	l.logger.Info("Performance alert generated",
		"job_id", alert.JobID,
		"alert_type", alert.AlertType,
		"severity", alert.Severity,
		"message", alert.Message,
	)
}
*/

// cleanOldLiveData removes old live data entries
func (l *LiveJobMonitor) cleanOldLiveData() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.config.DataRetentionPeriod)

	for jobID, data := range l.liveData {
		if data.Timestamp.Before(cutoff) {
			delete(l.liveData, jobID)
		}
	}
}

// GetLiveData returns live data for a specific job
func (l *LiveJobMonitor) GetLiveData(jobID string) (*JobLiveMetrics, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	data, exists := l.liveData[jobID]
	return data, exists
}

// GetAllLiveData returns all current live data
func (l *LiveJobMonitor) GetAllLiveData() map[string]*JobLiveMetrics {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make(map[string]*JobLiveMetrics)
	for k, v := range l.liveData {
		result[k] = v
	}
	return result
}

// StartStreaming starts real-time streaming for a job
func (l *LiveJobMonitor) StartStreaming(jobID string) <-chan *PerformanceAlert {
	l.streamingMu.Lock()
	defer l.streamingMu.Unlock()

	channel := make(chan *PerformanceAlert, l.config.StreamingBufferSize)
	l.alertChannels[jobID] = channel
	l.streamingActive = true

	return channel
}

// StopStreaming stops real-time streaming for a job
func (l *LiveJobMonitor) StopStreaming(jobID string) {
	l.streamingMu.Lock()
	defer l.streamingMu.Unlock()

	if channel, exists := l.alertChannels[jobID]; exists {
		close(channel)
		delete(l.alertChannels, jobID)
	}

	if len(l.alertChannels) == 0 {
		l.streamingActive = false
	}
}

// GetMonitoringStats returns monitoring statistics
func (l *LiveJobMonitor) GetMonitoringStats() map[string]interface{} {
	l.mu.RLock()
	l.streamingMu.RLock()
	defer l.mu.RUnlock()
	defer l.streamingMu.RUnlock()

	return map[string]interface{}{
		"monitored_jobs":   len(l.liveData),
		"active_streams":   len(l.alertChannels),
		"streaming_active": l.streamingActive,
		"last_collection":  l.lastCollection,
		"config":           l.config,
	}
}
