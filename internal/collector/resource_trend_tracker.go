package collector

import (
	"context"
	"fmt"
	"log/slog"
	// Commented out as only used in commented-out trend analysis functions
	// "math"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// ResourceTrendTracker tracks and analyzes resource utilization trends for pattern identification
type ResourceTrendTracker struct {
	slurmClient slurm.SlurmClient
	logger      *slog.Logger
	config      *TrendConfig
	metrics     *TrendMetrics
	mu          sync.RWMutex

	// Trend data storage
	trendData      map[string]*JobResourceTrends
	historicalData map[string][]*ResourceSnapshot
	patternCache   map[string]*IdentifiedPattern
	cacheTTL       time.Duration
	lastCollection time.Time
}

// TrendConfig holds configuration for trend tracking
type TrendConfig struct {
	TrackingInterval         time.Duration `yaml:"tracking_interval"`
	MaxJobsPerCollection     int           `yaml:"max_jobs_per_collection"`
	HistoryRetentionPeriod   time.Duration `yaml:"history_retention_period"`
	MinDataPointsForTrend    int           `yaml:"min_data_points_for_trend"`
	TrendSensitivity         float64       `yaml:"trend_sensitivity"` // 0.0 to 1.0
	PatternDetectionWindow   time.Duration `yaml:"pattern_detection_window"`
	EnablePredictiveAnalysis bool          `yaml:"enable_predictive_analysis"`
	CacheTTL                 time.Duration `yaml:"cache_ttl"`

	// Trend analysis parameters
	ShortTermWindow  time.Duration `yaml:"short_term_window"`  // 5 minutes
	MediumTermWindow time.Duration `yaml:"medium_term_window"` // 30 minutes
	LongTermWindow   time.Duration `yaml:"long_term_window"`   // 2 hours

	// Pattern detection thresholds
	PeriodicPatternThreshold  float64 `yaml:"periodic_pattern_threshold"`  // 0.7
	SeasonalPatternThreshold  float64 `yaml:"seasonal_pattern_threshold"`  // 0.8
	AnomalyDetectionThreshold float64 `yaml:"anomaly_detection_threshold"` // 0.9
	TrendChangeThreshold      float64 `yaml:"trend_change_threshold"`      // 0.3
}

// JobResourceTrends represents comprehensive resource trend analysis for a job
type JobResourceTrends struct {
	JobID       string    `json:"job_id"`
	LastUpdated time.Time `json:"last_updated"`

	// Resource trend analysis
	CPUTrend     *ResourceTrend `json:"cpu_trend"`
	MemoryTrend  *ResourceTrend `json:"memory_trend"`
	IOTrend      *ResourceTrend `json:"io_trend"`
	NetworkTrend *ResourceTrend `json:"network_trend"`

	// Overall trend indicators
	OverallTrend    string  `json:"overall_trend"`    // "increasing", "decreasing", "stable", "volatile"
	TrendStrength   float64 `json:"trend_strength"`   // 0.0 to 1.0
	TrendConfidence float64 `json:"trend_confidence"` // 0.0 to 1.0

	// Pattern identification
	IdentifiedPatterns []IdentifiedPattern `json:"identified_patterns"`

	// Predictive metrics
	PredictedPeakUsage    *PeakUsagePrediction `json:"predicted_peak_usage,omitempty"`
	ResourceExhaustionETA *time.Time           `json:"resource_exhaustion_eta,omitempty"`

	// Anomaly detection
	DetectedAnomalies []ResourceAnomaly `json:"detected_anomalies"`
}

// ResourceTrend represents trend analysis for a specific resource type
type ResourceTrend struct {
	ResourceType          string    `json:"resource_type"`     // "cpu", "memory", "io", "network"
	Direction             string    `json:"direction"`         // "increasing", "decreasing", "stable"
	Slope                 float64   `json:"slope"`             // Rate of change
	RSquared              float64   `json:"r_squared"`         // Trend fit quality (0.0 to 1.0)
	ShortTermTrend        string    `json:"short_term_trend"`  // Last 5 minutes
	MediumTermTrend       string    `json:"medium_term_trend"` // Last 30 minutes
	LongTermTrend         string    `json:"long_term_trend"`   // Last 2 hours
	Volatility            float64   `json:"volatility"`        // Measure of variance
	LastSignificantChange time.Time `json:"last_significant_change"`
}

// ResourceSnapshot represents a point-in-time resource utilization measurement
type ResourceSnapshot struct {
	Timestamp          time.Time `json:"timestamp"`
	CPUUtilization     float64   `json:"cpu_utilization"`
	MemoryUtilization  float64   `json:"memory_utilization"`
	IOUtilization      float64   `json:"io_utilization"`
	NetworkUtilization float64   `json:"network_utilization"`
	CPUAllocated       float64   `json:"cpu_allocated"`
	MemoryAllocated    int64     `json:"memory_allocated"`
}

// IdentifiedPattern represents a detected usage pattern
type IdentifiedPattern struct {
	Type              string        `json:"type"`             // "periodic", "seasonal", "burst", "gradual_increase"
	Confidence        float64       `json:"confidence"`       // 0.0 to 1.0
	Period            time.Duration `json:"period,omitempty"` // For periodic patterns
	Amplitude         float64       `json:"amplitude"`        // Pattern strength
	Phase             float64       `json:"phase,omitempty"`  // For periodic patterns
	Description       string        `json:"description"`
	FirstDetected     time.Time     `json:"first_detected"`
	LastSeen          time.Time     `json:"last_seen"`
	ResourcesAffected []string      `json:"resources_affected"` // Which resources show this pattern
}

// PeakUsagePrediction represents predicted peak resource usage
type PeakUsagePrediction struct {
	PredictedTime    time.Time `json:"predicted_time"`
	CPUPeakUsage     float64   `json:"cpu_peak_usage"`
	MemoryPeakUsage  float64   `json:"memory_peak_usage"`
	IOPeakUsage      float64   `json:"io_peak_usage"`
	NetworkPeakUsage float64   `json:"network_peak_usage"`
	Confidence       float64   `json:"confidence"`
}

// ResourceAnomaly represents a detected anomaly in resource usage
type ResourceAnomaly struct {
	Type          string        `json:"type"` // "spike", "drop", "plateau", "oscillation"
	ResourceType  string        `json:"resource_type"`
	Severity      string        `json:"severity"` // "low", "medium", "high", "critical"
	DetectedAt    time.Time     `json:"detected_at"`
	Duration      time.Duration `json:"duration"`
	ExpectedValue float64       `json:"expected_value"`
	ActualValue   float64       `json:"actual_value"`
	Deviation     float64       `json:"deviation"` // How far from expected
	Description   string        `json:"description"`
}

// TrendMetrics holds Prometheus metrics for trend tracking
type TrendMetrics struct {
	// Trend direction metrics
	ResourceTrendDirection *prometheus.GaugeVec
	TrendStrength          *prometheus.GaugeVec
	TrendConfidence        *prometheus.GaugeVec
	ResourceVolatility     *prometheus.GaugeVec

	// Pattern detection metrics
	PatternsDetected  *prometheus.GaugeVec
	PatternConfidence *prometheus.GaugeVec

	// Predictive metrics
	PredictedPeakUsage     *prometheus.GaugeVec
	ResourceExhaustionTime *prometheus.GaugeVec

	// Anomaly detection metrics
	AnomaliesDetected *prometheus.GaugeVec
	AnomalySeverity   *prometheus.GaugeVec

	// Trend analysis metrics
	ShortTermTrendIndicator  *prometheus.GaugeVec
	MediumTermTrendIndicator *prometheus.GaugeVec
	LongTermTrendIndicator   *prometheus.GaugeVec

	// Collection performance metrics
	TrendAnalysisDuration prometheus.Histogram
	TrendAnalysisErrors   *prometheus.CounterVec
	JobsTrendTracked      prometheus.Counter
	DataPointsCollected   prometheus.Counter
}

// NewResourceTrendTracker creates a new resource trend tracker
func NewResourceTrendTracker(slurmClient slurm.SlurmClient, logger *slog.Logger, config *TrendConfig) (*ResourceTrendTracker, error) {
	if config == nil {
		config = &TrendConfig{
			TrackingInterval:          30 * time.Second,
			MaxJobsPerCollection:      100,
			HistoryRetentionPeriod:    24 * time.Hour,
			MinDataPointsForTrend:     5,
			TrendSensitivity:          0.1,
			PatternDetectionWindow:    2 * time.Hour,
			EnablePredictiveAnalysis:  true,
			CacheTTL:                  5 * time.Minute,
			ShortTermWindow:           5 * time.Minute,
			MediumTermWindow:          30 * time.Minute,
			LongTermWindow:            2 * time.Hour,
			PeriodicPatternThreshold:  0.7,
			SeasonalPatternThreshold:  0.8,
			AnomalyDetectionThreshold: 0.9,
			TrendChangeThreshold:      0.3,
		}
	}

	metrics := &TrendMetrics{
		ResourceTrendDirection: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_trend_direction",
				Help: "Resource trend direction (1=increasing, 0=stable, -1=decreasing)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		TrendStrength: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_trend_strength",
				Help: "Strength of detected trend (0.0 to 1.0)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		TrendConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_trend_confidence",
				Help: "Confidence in trend analysis (0.0 to 1.0)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		ResourceVolatility: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_volatility",
				Help: "Resource usage volatility measure",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		PatternsDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_patterns_detected_total",
				Help: "Number of usage patterns detected",
			},
			[]string{"job_id", "user", "account", "partition", "pattern_type"},
		),
		PatternConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_pattern_confidence",
				Help: "Confidence in detected patterns (0.0 to 1.0)",
			},
			[]string{"job_id", "user", "account", "partition", "pattern_type"},
		),
		PredictedPeakUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_predicted_peak_usage",
				Help: "Predicted peak resource usage",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		ResourceExhaustionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_resource_exhaustion_eta_timestamp",
				Help: "Predicted resource exhaustion time as Unix timestamp",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		AnomaliesDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_anomalies_detected_total",
				Help: "Number of resource usage anomalies detected",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type", "anomaly_type"},
		),
		AnomalySeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_anomaly_severity",
				Help: "Severity of detected anomaly (1=low, 2=medium, 3=high, 4=critical)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type", "anomaly_type"},
		),
		ShortTermTrendIndicator: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_short_term_trend_indicator",
				Help: "Short-term trend indicator (1=increasing, 0=stable, -1=decreasing)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		MediumTermTrendIndicator: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_medium_term_trend_indicator",
				Help: "Medium-term trend indicator (1=increasing, 0=stable, -1=decreasing)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		LongTermTrendIndicator: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_long_term_trend_indicator",
				Help: "Long-term trend indicator (1=increasing, 0=stable, -1=decreasing)",
			},
			[]string{"job_id", "user", "account", "partition", "resource_type"},
		),
		TrendAnalysisDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "slurm_trend_analysis_duration_seconds",
				Help:    "Time spent performing trend analysis",
				Buckets: prometheus.DefBuckets,
			},
		),
		TrendAnalysisErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_trend_analysis_errors_total",
				Help: "Total number of trend analysis errors",
			},
			[]string{"error_type"},
		),
		JobsTrendTracked: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_jobs_trend_tracked_total",
				Help: "Total number of jobs tracked for trends",
			},
		),
		DataPointsCollected: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_trend_data_points_collected_total",
				Help: "Total number of trend data points collected",
			},
		),
	}

	tracker := &ResourceTrendTracker{
		slurmClient:    slurmClient,
		logger:         logger,
		config:         config,
		metrics:        metrics,
		trendData:      make(map[string]*JobResourceTrends),
		historicalData: make(map[string][]*ResourceSnapshot),
		patternCache:   make(map[string]*IdentifiedPattern),
		cacheTTL:       config.CacheTTL,
		lastCollection: time.Time{},
	}

	return tracker, nil
}

// Describe implements the prometheus.Collector interface
func (t *ResourceTrendTracker) Describe(ch chan<- *prometheus.Desc) {
	t.metrics.ResourceTrendDirection.Describe(ch)
	t.metrics.TrendStrength.Describe(ch)
	t.metrics.TrendConfidence.Describe(ch)
	t.metrics.ResourceVolatility.Describe(ch)
	t.metrics.PatternsDetected.Describe(ch)
	t.metrics.PatternConfidence.Describe(ch)
	t.metrics.PredictedPeakUsage.Describe(ch)
	t.metrics.ResourceExhaustionTime.Describe(ch)
	t.metrics.AnomaliesDetected.Describe(ch)
	t.metrics.AnomalySeverity.Describe(ch)
	t.metrics.ShortTermTrendIndicator.Describe(ch)
	t.metrics.MediumTermTrendIndicator.Describe(ch)
	t.metrics.LongTermTrendIndicator.Describe(ch)
	t.metrics.TrendAnalysisDuration.Describe(ch)
	t.metrics.TrendAnalysisErrors.Describe(ch)
	t.metrics.JobsTrendTracked.Describe(ch)
	t.metrics.DataPointsCollected.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (t *ResourceTrendTracker) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		t.metrics.TrendAnalysisDuration.Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := t.trackResourceTrends(ctx); err != nil {
		t.logger.Error("Failed to track resource trends", "error", err)
		t.metrics.TrendAnalysisErrors.WithLabelValues("tracking_failure").Inc()
	}

	// Collect metrics
	t.metrics.ResourceTrendDirection.Collect(ch)
	t.metrics.TrendStrength.Collect(ch)
	t.metrics.TrendConfidence.Collect(ch)
	t.metrics.ResourceVolatility.Collect(ch)
	t.metrics.PatternsDetected.Collect(ch)
	t.metrics.PatternConfidence.Collect(ch)
	t.metrics.PredictedPeakUsage.Collect(ch)
	t.metrics.ResourceExhaustionTime.Collect(ch)
	t.metrics.AnomaliesDetected.Collect(ch)
	t.metrics.AnomalySeverity.Collect(ch)
	t.metrics.ShortTermTrendIndicator.Collect(ch)
	t.metrics.MediumTermTrendIndicator.Collect(ch)
	t.metrics.LongTermTrendIndicator.Collect(ch)
	t.metrics.TrendAnalysisDuration.Collect(ch)
	t.metrics.TrendAnalysisErrors.Collect(ch)
	t.metrics.JobsTrendTracked.Collect(ch)
	t.metrics.DataPointsCollected.Collect(ch)
}

// trackResourceTrends performs resource trend tracking and analysis
func (t *ResourceTrendTracker) trackResourceTrends(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get job manager
	jobManager := t.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// List running jobs for trend tracking
	// TODO: ListJobsOptions structure is not compatible with current slurm-client
	// Using nil for options as a workaround
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	t.logger.Debug("Tracking resource trends", "job_count", len(jobs.Jobs))

	// Process each job
	// TODO: Job type mismatch - jobs.Jobs returns []interfaces.Job but functions expect *slurm.Job
	// Skipping job processing for now
	_ = jobs // Suppress unused variable warning
	/*
		for _, job := range jobs.Jobs {
			// Create current snapshot
			snapshot := t.createResourceSnapshot(job)

			// Add to historical data
			t.addHistoricalSnapshot(job.JobID, snapshot)

			// Analyze trends if we have enough data
			if len(t.historicalData[job.JobID]) >= t.config.MinDataPointsForTrend {
				trends := t.analyzeJobResourceTrends(job)
				t.trendData[job.JobID] = trends

				// Update metrics
				t.updateTrendMetrics(job, trends)
			}

			t.metrics.JobsTrendTracked.Inc()
			t.metrics.DataPointsCollected.Inc()
		}
	*/

	// Clean old historical data
	t.cleanOldHistoricalData()

	t.lastCollection = time.Now()
	return nil
}

// TODO: Following resource trend analysis methods are unused - preserved for future trend tracking implementation
/*
// createResourceSnapshot creates a resource snapshot from current job data
func (t *ResourceTrendTracker) createResourceSnapshot(job *slurm.Job) *ResourceSnapshot {
	snapshot := &ResourceSnapshot{
		Timestamp:       time.Now(),
		CPUAllocated:    float64(job.CPUs),
		MemoryAllocated: int64(job.Memory) * 1024 * 1024, // Convert MB to bytes
	}

	// Create utilization data for calculation
	utilizationData := CreateResourceUtilizationDataFromJob(job)

	// Calculate current utilization
	if utilizationData.CPUAllocated > 0 {
		snapshot.CPUUtilization = utilizationData.CPUUsed / utilizationData.CPUAllocated
	}

	if utilizationData.MemoryAllocated > 0 {
		snapshot.MemoryUtilization = float64(utilizationData.MemoryUsed) / float64(utilizationData.MemoryAllocated)
	}

	// I/O and network utilization (simplified estimates)
	if utilizationData.WallTime > 0 {
		totalIOBytes := utilizationData.IOReadBytes + utilizationData.IOWriteBytes
		snapshot.IOUtilization = float64(totalIOBytes) / utilizationData.WallTime / (100 * 1024 * 1024) // Normalize to 100MB/s baseline

		totalNetworkBytes := utilizationData.NetworkRxBytes + utilizationData.NetworkTxBytes
		snapshot.NetworkUtilization = float64(totalNetworkBytes) / utilizationData.WallTime / (1024 * 1024 * 1024) // Normalize to 1GB/s baseline
	}

	return snapshot
}

// addHistoricalSnapshot adds a snapshot to historical data
func (t *ResourceTrendTracker) addHistoricalSnapshot(jobID string, snapshot *ResourceSnapshot) {
	if t.historicalData[jobID] == nil {
		t.historicalData[jobID] = []*ResourceSnapshot{}
	}

	t.historicalData[jobID] = append(t.historicalData[jobID], snapshot)

	// Limit historical data size
	maxHistorySize := int(t.config.HistoryRetentionPeriod / t.config.TrackingInterval)
	if len(t.historicalData[jobID]) > maxHistorySize {
		t.historicalData[jobID] = t.historicalData[jobID][len(t.historicalData[jobID])-maxHistorySize:]
	}
}

// analyzeJobResourceTrends analyzes trends for a specific job
func (t *ResourceTrendTracker) analyzeJobResourceTrends(job *slurm.Job) *JobResourceTrends {
	// TODO: job.JobID field not available in current slurm-client version
	jobID := "job_0"
	snapshots := t.historicalData[jobID]
	if len(snapshots) < t.config.MinDataPointsForTrend {
		return nil
	}

	trends := &JobResourceTrends{
		JobID:       jobID,
		LastUpdated: time.Now(),
	}

	// Analyze trends for each resource type
	trends.CPUTrend = t.analyzeResourceTrend("cpu", snapshots)
	trends.MemoryTrend = t.analyzeResourceTrend("memory", snapshots)
	trends.IOTrend = t.analyzeResourceTrend("io", snapshots)
	trends.NetworkTrend = t.analyzeResourceTrend("network", snapshots)

	// Calculate overall trend indicators
	t.calculateOverallTrend(trends)

	// Detect patterns
	trends.IdentifiedPatterns = t.detectUsagePatterns(snapshots)

	// Predictive analysis
	if t.config.EnablePredictiveAnalysis {
		trends.PredictedPeakUsage = t.predictPeakUsage(snapshots)
		trends.ResourceExhaustionETA = t.predictResourceExhaustion(snapshots, job)
	}

	// Anomaly detection
	trends.DetectedAnomalies = t.detectAnomalies(snapshots)

	return trends
}

// analyzeResourceTrend analyzes trend for a specific resource type
func (t *ResourceTrendTracker) analyzeResourceTrend(resourceType string, snapshots []*ResourceSnapshot) *ResourceTrend {
	if len(snapshots) < 2 {
		return nil
	}

	trend := &ResourceTrend{
		ResourceType: resourceType,
	}

	// Extract values for the specific resource
	values := make([]float64, len(snapshots))
	timestamps := make([]float64, len(snapshots))

	for i, snapshot := range snapshots {
		timestamps[i] = float64(snapshot.Timestamp.Unix())
		switch resourceType {
		case "cpu":
			values[i] = snapshot.CPUUtilization
		case "memory":
			values[i] = snapshot.MemoryUtilization
		case "io":
			values[i] = snapshot.IOUtilization
		case "network":
			values[i] = snapshot.NetworkUtilization
		}
	}

	// Calculate linear regression
	slope, rSquared := t.calculateLinearRegression(timestamps, values)
	trend.Slope = slope
	trend.RSquared = rSquared

	// Determine trend direction
	if math.Abs(slope) < t.config.TrendSensitivity {
		trend.Direction = "stable"
	} else if slope > 0 {
		trend.Direction = "increasing"
	} else {
		trend.Direction = "decreasing"
	}

	// Calculate volatility
	trend.Volatility = t.calculateVolatility(values)

	// Analyze short, medium, and long-term trends
	trend.ShortTermTrend = t.analyzeWindowTrend(snapshots, t.config.ShortTermWindow, resourceType)
	trend.MediumTermTrend = t.analyzeWindowTrend(snapshots, t.config.MediumTermWindow, resourceType)
	trend.LongTermTrend = t.analyzeWindowTrend(snapshots, t.config.LongTermWindow, resourceType)

	// Find last significant change
	trend.LastSignificantChange = t.findLastSignificantChange(snapshots, resourceType)

	return trend
}

// calculateLinearRegression calculates linear regression slope and R-squared
func (t *ResourceTrendTracker) calculateLinearRegression(x, y []float64) (slope, rSquared float64) {
	if len(x) != len(y) || len(x) < 2 {
		return 0, 0
	}

	n := float64(len(x))

	// Calculate means
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / n
	meanY := sumY / n

	// Calculate slope and correlation
	var numerator, denominatorX, denominatorY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denominatorX += dx * dx
		denominatorY += dy * dy
	}

	if denominatorX == 0 {
		return 0, 0
	}

	slope = numerator / denominatorX

	// Calculate R-squared
	if denominatorY == 0 {
		rSquared = 1.0 // Perfect fit if no variance in y
	} else {
		correlation := numerator / math.Sqrt(denominatorX*denominatorY)
		rSquared = correlation * correlation
	}

	return slope, rSquared
}

// calculateVolatility calculates the volatility (standard deviation) of values
func (t *ResourceTrendTracker) calculateVolatility(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values))

	return math.Sqrt(variance)
}

// analyzeWindowTrend analyzes trend within a specific time window
func (t *ResourceTrendTracker) analyzeWindowTrend(snapshots []*ResourceSnapshot, window time.Duration, resourceType string) string {
	if len(snapshots) < 2 {
		return "unknown"
	}

	// Get snapshots within the window
	cutoff := time.Now().Add(-window)
	var windowSnapshots []*ResourceSnapshot

	for _, snapshot := range snapshots {
		if snapshot.Timestamp.After(cutoff) {
			windowSnapshots = append(windowSnapshots, snapshot)
		}
	}

	if len(windowSnapshots) < 2 {
		return "insufficient_data"
	}

	// Extract values for analysis
	values := make([]float64, len(windowSnapshots))
	timestamps := make([]float64, len(windowSnapshots))

	for i, snapshot := range windowSnapshots {
		timestamps[i] = float64(snapshot.Timestamp.Unix())
		switch resourceType {
		case "cpu":
			values[i] = snapshot.CPUUtilization
		case "memory":
			values[i] = snapshot.MemoryUtilization
		case "io":
			values[i] = snapshot.IOUtilization
		case "network":
			values[i] = snapshot.NetworkUtilization
		}
	}

	slope, _ := t.calculateLinearRegression(timestamps, values)

	if math.Abs(slope) < t.config.TrendSensitivity {
		return "stable"
	} else if slope > 0 {
		return "increasing"
	} else {
		return "decreasing"
	}
}

// calculateOverallTrend calculates overall trend indicators
func (t *ResourceTrendTracker) calculateOverallTrend(trends *JobResourceTrends) {
	directions := []string{}
	strengths := []float64{}
	confidences := []float64{}

	// Collect trend information
	for _, trend := range []*ResourceTrend{trends.CPUTrend, trends.MemoryTrend, trends.IOTrend, trends.NetworkTrend} {
		if trend != nil {
			directions = append(directions, trend.Direction)
			strengths = append(strengths, math.Abs(trend.Slope))
			confidences = append(confidences, trend.RSquared)
		}
	}

	if len(directions) == 0 {
		trends.OverallTrend = "unknown"
		return
	}

	// Count direction occurrences
	directionCounts := make(map[string]int)
	for _, dir := range directions {
		directionCounts[dir]++
	}

	// Find most common direction
	maxCount := 0
	for _, count := range directionCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	// Determine overall trend
	if directionCounts["stable"] == maxCount && maxCount > len(directions)/2 {
		trends.OverallTrend = "stable"
	} else if directionCounts["increasing"] == maxCount {
		trends.OverallTrend = "increasing"
	} else if directionCounts["decreasing"] == maxCount {
		trends.OverallTrend = "decreasing"
	} else {
		trends.OverallTrend = "volatile"
	}

	// Calculate average strength and confidence
	if len(strengths) > 0 {
		var sumStrength, sumConfidence float64
		for i := range strengths {
			sumStrength += strengths[i]
			sumConfidence += confidences[i]
		}
		trends.TrendStrength = sumStrength / float64(len(strengths))
		trends.TrendConfidence = sumConfidence / float64(len(confidences))
	}
}

// detectUsagePatterns detects various usage patterns
func (t *ResourceTrendTracker) detectUsagePatterns(snapshots []*ResourceSnapshot) []IdentifiedPattern {
	patterns := []IdentifiedPattern{}

	// Detect periodic patterns
	if periodicPattern := t.detectPeriodicPattern(snapshots); periodicPattern != nil {
		patterns = append(patterns, *periodicPattern)
	}

	// Detect burst patterns
	if burstPattern := t.detectBurstPattern(snapshots); burstPattern != nil {
		patterns = append(patterns, *burstPattern)
	}

	// Detect gradual increase patterns
	if gradualPattern := t.detectGradualIncreasePattern(snapshots); gradualPattern != nil {
		patterns = append(patterns, *gradualPattern)
	}

	return patterns
}

// detectPeriodicPattern detects periodic usage patterns
func (t *ResourceTrendTracker) detectPeriodicPattern(snapshots []*ResourceSnapshot) *IdentifiedPattern {
	if len(snapshots) < 10 {
		return nil
	}

	// Simplified periodic detection (would use FFT in real implementation)
	// Check for repeating patterns in CPU utilization
	cpuValues := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		cpuValues[i] = snapshot.CPUUtilization
	}

	// Look for repeating patterns
	if t.hasPeriodicBehavior(cpuValues) {
		return &IdentifiedPattern{
			Type:              "periodic",
			Confidence:        0.8,
			Period:            15 * time.Minute, // Simplified estimate
			Amplitude:         t.calculateVolatility(cpuValues),
			Description:       "Periodic CPU usage pattern detected",
			FirstDetected:     snapshots[0].Timestamp,
			LastSeen:          snapshots[len(snapshots)-1].Timestamp,
			ResourcesAffected: []string{"cpu"},
		}
	}

	return nil
}

// detectBurstPattern detects burst usage patterns
func (t *ResourceTrendTracker) detectBurstPattern(snapshots []*ResourceSnapshot) *IdentifiedPattern {
	if len(snapshots) < 5 {
		return nil
	}

	// Look for sudden spikes in usage
	cpuValues := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		cpuValues[i] = snapshot.CPUUtilization
	}

	mean := t.calculateMean(cpuValues)
	stdDev := t.calculateVolatility(cpuValues)

	burstCount := 0
	for _, value := range cpuValues {
		if value > mean+2*stdDev { // Values more than 2 standard deviations above mean
			burstCount++
		}
	}

	if burstCount > 0 && float64(burstCount)/float64(len(cpuValues)) > 0.1 { // More than 10% burst points
		return &IdentifiedPattern{
			Type:              "burst",
			Confidence:        0.7,
			Amplitude:         stdDev,
			Description:       "Burst usage pattern detected with periodic spikes",
			FirstDetected:     snapshots[0].Timestamp,
			LastSeen:          snapshots[len(snapshots)-1].Timestamp,
			ResourcesAffected: []string{"cpu"},
		}
	}

	return nil
}

// detectGradualIncreasePattern detects gradual increase patterns
func (t *ResourceTrendTracker) detectGradualIncreasePattern(snapshots []*ResourceSnapshot) *IdentifiedPattern {
	if len(snapshots) < 5 {
		return nil
	}

	cpuValues := make([]float64, len(snapshots))
	timestamps := make([]float64, len(snapshots))

	for i, snapshot := range snapshots {
		cpuValues[i] = snapshot.CPUUtilization
		timestamps[i] = float64(snapshot.Timestamp.Unix())
	}

	slope, rSquared := t.calculateLinearRegression(timestamps, cpuValues)

	if slope > t.config.TrendSensitivity && rSquared > 0.7 {
		return &IdentifiedPattern{
			Type:              "gradual_increase",
			Confidence:        rSquared,
			Amplitude:         slope,
			Description:       "Gradual increase in resource usage over time",
			FirstDetected:     snapshots[0].Timestamp,
			LastSeen:          snapshots[len(snapshots)-1].Timestamp,
			ResourcesAffected: []string{"cpu"},
		}
	}

	return nil
}

// Additional helper methods for pattern detection, anomaly detection, predictive analysis, etc.
// These are simplified implementations for core functionality

// hasPeriodicBehavior checks if values show periodic behavior
func (t *ResourceTrendTracker) hasPeriodicBehavior(values []float64) bool {
	// Simplified periodic detection - would use autocorrelation or FFT in real implementation
	if len(values) < 6 {
		return false
	}

	// Check for repeating patterns by comparing segments
	segmentSize := len(values) / 3
	if segmentSize < 2 {
		return false
	}

	segment1 := values[:segmentSize]
	segment2 := values[segmentSize : 2*segmentSize]

	correlation := t.calculateCorrelation(segment1, segment2)
	return correlation > t.config.PeriodicPatternThreshold
}

// calculateCorrelation calculates correlation between two series
func (t *ResourceTrendTracker) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	meanX := t.calculateMean(x)
	meanY := t.calculateMean(y)

	var numerator, denomX, denomY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		denomX += dx * dx
		denomY += dy * dy
	}

	if denomX == 0 || denomY == 0 {
		return 0
	}

	return numerator / math.Sqrt(denomX*denomY)
}

// calculateMean calculates the mean of values
func (t *ResourceTrendTracker) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// predictPeakUsage predicts peak usage based on trends
func (t *ResourceTrendTracker) predictPeakUsage(snapshots []*ResourceSnapshot) *PeakUsagePrediction {
	if len(snapshots) < 3 {
		return nil
	}

	// Simple prediction based on current trend
	latest := snapshots[len(snapshots)-1]

	return &PeakUsagePrediction{
		PredictedTime:    time.Now().Add(1 * time.Hour), // Simplified prediction
		CPUPeakUsage:     math.Min(latest.CPUUtilization*1.2, 1.0),
		MemoryPeakUsage:  math.Min(latest.MemoryUtilization*1.15, 1.0),
		IOPeakUsage:      latest.IOUtilization * 1.1,
		NetworkPeakUsage: latest.NetworkUtilization * 1.1,
		Confidence:       0.6,
	}
}

// predictResourceExhaustion predicts when resources might be exhausted
func (t *ResourceTrendTracker) predictResourceExhaustion(snapshots []*ResourceSnapshot, job *slurm.Job) *time.Time {
	if len(snapshots) < 3 {
		return nil
	}

	// Simple exhaustion prediction for memory
	memValues := make([]float64, len(snapshots))
	timestamps := make([]float64, len(snapshots))

	for i, snapshot := range snapshots {
		memValues[i] = snapshot.MemoryUtilization
		timestamps[i] = float64(snapshot.Timestamp.Unix())
	}

	slope, rSquared := t.calculateLinearRegression(timestamps, memValues)

	if slope > 0 && rSquared > 0.5 {
		// Calculate when usage might reach 100%
		currentUsage := memValues[len(memValues)-1]
		remainingCapacity := 1.0 - currentUsage

		if slope > 0 {
			timeToExhaustion := remainingCapacity / slope
			exhaustionTime := time.Now().Add(time.Duration(timeToExhaustion) * time.Second)
			return &exhaustionTime
		}
	}

	return nil
}

// detectAnomalies detects anomalies in resource usage
func (t *ResourceTrendTracker) detectAnomalies(snapshots []*ResourceSnapshot) []ResourceAnomaly {
	anomalies := []ResourceAnomaly{}

	if len(snapshots) < 5 {
		return anomalies
	}

	// Detect CPU usage anomalies
	cpuAnomalies := t.detectResourceAnomalies("cpu", snapshots)
	anomalies = append(anomalies, cpuAnomalies...)

	// Detect memory usage anomalies
	memoryAnomalies := t.detectResourceAnomalies("memory", snapshots)
	anomalies = append(anomalies, memoryAnomalies...)

	return anomalies
}

// detectResourceAnomalies detects anomalies for a specific resource
func (t *ResourceTrendTracker) detectResourceAnomalies(resourceType string, snapshots []*ResourceSnapshot) []ResourceAnomaly {
	anomalies := []ResourceAnomaly{}

	values := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		switch resourceType {
		case "cpu":
			values[i] = snapshot.CPUUtilization
		case "memory":
			values[i] = snapshot.MemoryUtilization
		case "io":
			values[i] = snapshot.IOUtilization
		case "network":
			values[i] = snapshot.NetworkUtilization
		}
	}

	mean := t.calculateMean(values)
	stdDev := t.calculateVolatility(values)

	// Detect spikes (values significantly above normal)
	for i, value := range values {
		if value > mean+3*stdDev && stdDev > 0.1 { // 3 sigma rule with minimum threshold
			anomalies = append(anomalies, ResourceAnomaly{
				Type:          "spike",
				ResourceType:  resourceType,
				Severity:      t.calculateAnomalySeverity(value, mean, stdDev),
				DetectedAt:    snapshots[i].Timestamp,
				Duration:      t.config.TrackingInterval, // Simplified
				ExpectedValue: mean,
				ActualValue:   value,
				Deviation:     (value - mean) / stdDev,
				Description:   fmt.Sprintf("Unusual spike in %s usage", resourceType),
			})
		}
	}

	return anomalies
}

// calculateAnomalySeverity calculates severity of an anomaly
func (t *ResourceTrendTracker) calculateAnomalySeverity(value, mean, stdDev float64) string {
	if stdDev == 0 {
		return "low"
	}

	deviations := math.Abs(value-mean) / stdDev

	if deviations > 4 {
		return "critical"
	} else if deviations > 3 {
		return "high"
	} else if deviations > 2 {
		return "medium"
	} else {
		return "low"
	}
}

// findLastSignificantChange finds the last significant change in resource usage
func (t *ResourceTrendTracker) findLastSignificantChange(snapshots []*ResourceSnapshot, resourceType string) time.Time {
	if len(snapshots) < 2 {
		return time.Time{}
	}

	values := make([]float64, len(snapshots))
	for i, snapshot := range snapshots {
		switch resourceType {
		case "cpu":
			values[i] = snapshot.CPUUtilization
		case "memory":
			values[i] = snapshot.MemoryUtilization
		case "io":
			values[i] = snapshot.IOUtilization
		case "network":
			values[i] = snapshot.NetworkUtilization
		}
	}

	// Look for significant changes (simplified)
	for i := len(values) - 1; i > 0; i-- {
		change := math.Abs(values[i] - values[i-1])
		if change > t.config.TrendChangeThreshold {
			return snapshots[i].Timestamp
		}
	}

	return snapshots[0].Timestamp
}

// updateTrendMetrics updates Prometheus metrics from trend analysis
func (t *ResourceTrendTracker) updateTrendMetrics(job *slurm.Job, trends *JobResourceTrends) {
	// TODO: job field names are not compatible with current slurm-client version
	labels := []string{
		trends.JobID,
		"unknown_user",
		"unknown_account",
		"unknown_partition",
	}

	// Update trend direction metrics for each resource
	resourceTrends := map[string]*ResourceTrend{
		"cpu":     trends.CPUTrend,
		"memory":  trends.MemoryTrend,
		"io":      trends.IOTrend,
		"network": trends.NetworkTrend,
	}

	for resourceType, trend := range resourceTrends {
		if trend == nil {
			continue
		}

		resourceLabels := append(labels, resourceType)

		// Trend direction
		var direction float64
		switch trend.Direction {
		case "increasing":
			direction = 1.0
		case "decreasing":
			direction = -1.0
		default:
			direction = 0.0
		}
		t.metrics.ResourceTrendDirection.WithLabelValues(resourceLabels...).Set(direction)

		// Trend strength and confidence
		t.metrics.TrendStrength.WithLabelValues(resourceLabels...).Set(math.Abs(trend.Slope))
		t.metrics.TrendConfidence.WithLabelValues(resourceLabels...).Set(trend.RSquared)
		t.metrics.ResourceVolatility.WithLabelValues(resourceLabels...).Set(trend.Volatility)

		// Short, medium, long-term trends
		t.metrics.ShortTermTrendIndicator.WithLabelValues(resourceLabels...).Set(t.trendToFloat(trend.ShortTermTrend))
		t.metrics.MediumTermTrendIndicator.WithLabelValues(resourceLabels...).Set(t.trendToFloat(trend.MediumTermTrend))
		t.metrics.LongTermTrendIndicator.WithLabelValues(resourceLabels...).Set(t.trendToFloat(trend.LongTermTrend))
	}

	// Pattern detection metrics
	patternCounts := make(map[string]int)
	for _, pattern := range trends.IdentifiedPatterns {
		patternCounts[pattern.Type]++

		patternLabels := append(labels, pattern.Type)
		t.metrics.PatternsDetected.WithLabelValues(patternLabels...).Set(float64(patternCounts[pattern.Type]))
		t.metrics.PatternConfidence.WithLabelValues(patternLabels...).Set(pattern.Confidence)
	}

	// Predictive metrics
	if trends.PredictedPeakUsage != nil {
		for resourceType, peakValue := range map[string]float64{
			"cpu":     trends.PredictedPeakUsage.CPUPeakUsage,
			"memory":  trends.PredictedPeakUsage.MemoryPeakUsage,
			"io":      trends.PredictedPeakUsage.IOPeakUsage,
			"network": trends.PredictedPeakUsage.NetworkPeakUsage,
		} {
			peakLabels := append(labels, resourceType)
			t.metrics.PredictedPeakUsage.WithLabelValues(peakLabels...).Set(peakValue)
		}
	}

	if trends.ResourceExhaustionETA != nil {
		exhaustionLabels := append(labels, "memory") // Simplified to memory exhaustion
		t.metrics.ResourceExhaustionTime.WithLabelValues(exhaustionLabels...).Set(float64(trends.ResourceExhaustionETA.Unix()))
	}

	// Anomaly metrics
	anomalyCounts := make(map[string]map[string]int) // resourceType -> anomalyType -> count
	for _, anomaly := range trends.DetectedAnomalies {
		if anomalyCounts[anomaly.ResourceType] == nil {
			anomalyCounts[anomaly.ResourceType] = make(map[string]int)
		}
		anomalyCounts[anomaly.ResourceType][anomaly.Type]++

		anomalyLabels := append(labels, anomaly.ResourceType, anomaly.Type)
		t.metrics.AnomaliesDetected.WithLabelValues(anomalyLabels...).Set(float64(anomalyCounts[anomaly.ResourceType][anomaly.Type]))

		severityValue := t.severityToFloat(anomaly.Severity)
		t.metrics.AnomalySeverity.WithLabelValues(anomalyLabels...).Set(severityValue)
	}
}

// trendToFloat converts trend string to float value
func (t *ResourceTrendTracker) trendToFloat(trend string) float64 {
	switch trend {
	case "increasing":
		return 1.0
	case "decreasing":
		return -1.0
	default:
		return 0.0
	}
}

// severityToFloat converts severity string to float value
func (t *ResourceTrendTracker) severityToFloat(severity string) float64 {
	switch severity {
	case "critical":
		return 4.0
	case "high":
		return 3.0
	case "medium":
		return 2.0
	case "low":
		return 1.0
	default:
		return 0.0
	}
}
*/

// cleanOldHistoricalData removes old historical data beyond retention period
func (t *ResourceTrendTracker) cleanOldHistoricalData() {
	cutoff := time.Now().Add(-t.config.HistoryRetentionPeriod)

	for jobID, snapshots := range t.historicalData {
		var filteredSnapshots []*ResourceSnapshot
		for _, snapshot := range snapshots {
			if snapshot.Timestamp.After(cutoff) {
				filteredSnapshots = append(filteredSnapshots, snapshot)
			}
		}

		if len(filteredSnapshots) == 0 {
			delete(t.historicalData, jobID)
		} else {
			t.historicalData[jobID] = filteredSnapshots
		}
	}
}

// GetTrendData returns trend data for a specific job
func (t *ResourceTrendTracker) GetTrendData(jobID string) *JobResourceTrends {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.trendData[jobID]
}

// GetHistoricalDataSize returns the size of historical data for a job
func (t *ResourceTrendTracker) GetHistoricalDataSize(jobID string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.historicalData[jobID])
}

// GetLastCollection returns the timestamp of the last successful collection
func (t *ResourceTrendTracker) GetLastCollection() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastCollection
}
