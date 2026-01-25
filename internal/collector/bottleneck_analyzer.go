// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
)

// BottleneckAnalyzer provides advanced bottleneck detection and analysis for SLURM jobs
type BottleneckAnalyzer struct {
	slurmClient    slurm.SlurmClient
	logger         *slog.Logger
	config         *BottleneckConfig
	metrics        *BottleneckMetrics
	efficiencyCalc *EfficiencyCalculator
	mu             sync.RWMutex

	// Cache for analysis results
	analysisCache  map[string]*StepPerformanceAnalysis
	cacheTTL       time.Duration
	lastCollection time.Time
}

// BottleneckConfig holds configuration for bottleneck analysis
type BottleneckConfig struct {
	AnalysisInterval         time.Duration `yaml:"analysis_interval"`
	MaxJobsPerAnalysis       int           `yaml:"max_jobs_per_analysis"`
	EnablePredictiveAnalysis bool          `yaml:"enable_predictive_analysis"`
	EnableRootCauseAnalysis  bool          `yaml:"enable_root_cause_analysis"`
	CacheTTL                 time.Duration `yaml:"cache_ttl"`

	// Bottleneck detection thresholds
	CPUBottleneckThreshold     float64 `yaml:"cpu_bottleneck_threshold"`     // CPU usage above this indicates bottleneck
	MemoryBottleneckThreshold  float64 `yaml:"memory_bottleneck_threshold"`  // Memory usage above this indicates bottleneck
	IOBottleneckThreshold      float64 `yaml:"io_bottleneck_threshold"`      // I/O wait above this indicates bottleneck
	NetworkBottleneckThreshold float64 `yaml:"network_bottleneck_threshold"` // Network utilization above this indicates bottleneck

	// Performance degradation thresholds
	PerformanceDegradationThreshold float64 `yaml:"performance_degradation_threshold"` // Performance drop above this is significant

	// Analysis sensitivity settings
	SensitivityLevel         string  `yaml:"sensitivity_level"` // "low", "medium", "high"
	MinDataPointsForAnalysis int     `yaml:"min_data_points_for_analysis"`
	AnalysisConfidenceLevel  float64 `yaml:"analysis_confidence_level"`
}

// StepPerformanceAnalysis represents comprehensive performance analysis of a job step
type StepPerformanceAnalysis struct {
	JobID             string    `json:"job_id"`
	StepID            string    `json:"step_id"`
	AnalysisTimestamp time.Time `json:"analysis_timestamp"`

	// Bottleneck identification
	PrimaryBottleneck    string   `json:"primary_bottleneck"` // "cpu", "memory", "io", "network", "none"
	SecondaryBottlenecks []string `json:"secondary_bottlenecks"`
	BottleneckSeverity   float64  `json:"bottleneck_severity"`   // 0.0 to 1.0
	BottleneckConfidence float64  `json:"bottleneck_confidence"` // 0.0 to 1.0

	// Performance metrics
	CPUEfficiency     float64 `json:"cpu_efficiency"`
	MemoryEfficiency  float64 `json:"memory_efficiency"`
	IOEfficiency      float64 `json:"io_efficiency"`
	NetworkEfficiency float64 `json:"network_efficiency"`
	OverallEfficiency float64 `json:"overall_efficiency"`

	// Resource utilization patterns
	CPUUtilizationPattern     string `json:"cpu_utilization_pattern"` // "stable", "bursty", "declining", "increasing"
	MemoryUtilizationPattern  string `json:"memory_utilization_pattern"`
	IOUtilizationPattern      string `json:"io_utilization_pattern"`
	NetworkUtilizationPattern string `json:"network_utilization_pattern"`

	// Performance issues
	PerformanceIssues         []PerformanceIssue        `json:"performance_issues"`
	OptimizationOpportunities []OptimizationOpportunity `json:"optimization_opportunities"`

	// Predictive analysis
	PredictedCompletion    *time.Time `json:"predicted_completion,omitempty"`
	EstimatedTimeRemaining float64    `json:"estimated_time_remaining"` // seconds
	PerformanceTrend       string     `json:"performance_trend"`        // "improving", "stable", "degrading"

	// Root cause analysis
	RootCauses         []RootCause `json:"root_causes"`
	RecommendedActions []string    `json:"recommended_actions"`
}

// PerformanceIssue represents a detected performance issue
type PerformanceIssue struct {
	Type        string    `json:"type"`     // "cpu_underutilization", "memory_pressure", "io_wait", "network_congestion"
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // Description of performance impact
	FirstSeen   time.Time `json:"first_seen"`
	Frequency   float64   `json:"frequency"` // How often this issue occurs (0.0 to 1.0)
}

// Note: OptimizationOpportunity type is defined in common_types.go

// RootCause represents a root cause analysis result
type RootCause struct {
	Cause              string   `json:"cause"`
	Confidence         float64  `json:"confidence"`          // 0.0 to 1.0
	ContributionFactor float64  `json:"contribution_factor"` // How much this cause contributes to the issue (0.0 to 1.0)
	Evidence           []string `json:"evidence"`            // Supporting evidence for this root cause
}

// BottleneckMetrics holds Prometheus metrics for bottleneck analysis
type BottleneckMetrics struct {
	// Bottleneck detection metrics
	BottleneckDetected   *prometheus.GaugeVec
	BottleneckSeverity   *prometheus.GaugeVec
	BottleneckConfidence *prometheus.GaugeVec
	BottleneckType       *prometheus.GaugeVec

	// Performance analysis metrics
	PerformanceEfficiencyScore *prometheus.GaugeVec
	PerformanceIssueCount      *prometheus.GaugeVec
	OptimizationOpportunities  *prometheus.GaugeVec

	// Predictive metrics
	PredictedCompletionTime   *prometheus.GaugeVec
	EstimatedTimeRemaining    *prometheus.GaugeVec
	PerformanceTrendIndicator *prometheus.GaugeVec

	// Root cause analysis metrics
	RootCauseCount      *prometheus.GaugeVec
	RootCauseConfidence *prometheus.GaugeVec

	// Analysis performance metrics
	AnalysisDuration prometheus.Histogram
	AnalysisErrors   *prometheus.CounterVec
	StepsAnalyzed    prometheus.Counter
	BottlenecksFound prometheus.Counter
}

// NewBottleneckAnalyzer creates a new bottleneck analyzer
func NewBottleneckAnalyzer(slurmClient slurm.SlurmClient, logger *slog.Logger, config *BottleneckConfig) (*BottleneckAnalyzer, error) {
	if config == nil {
		config = &BottleneckConfig{
			AnalysisInterval:                30 * time.Second,
			MaxJobsPerAnalysis:              100,
			EnablePredictiveAnalysis:        true,
			EnableRootCauseAnalysis:         true,
			CacheTTL:                        5 * time.Minute,
			CPUBottleneckThreshold:          0.9,  // 90% CPU usage
			MemoryBottleneckThreshold:       0.85, // 85% memory usage
			IOBottleneckThreshold:           0.2,  // 20% I/O wait
			NetworkBottleneckThreshold:      0.8,  // 80% network utilization
			PerformanceDegradationThreshold: 0.2,  // 20% performance drop
			SensitivityLevel:                "medium",
			MinDataPointsForAnalysis:        3,
			AnalysisConfidenceLevel:         0.7, // 70% confidence threshold
		}
	}

	metrics := &BottleneckMetrics{
		BottleneckDetected: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_bottleneck_detected",
				Help: "Whether a bottleneck was detected for job step (1 = yes, 0 = no)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		BottleneckSeverity: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_bottleneck_severity",
				Help: "Severity of detected bottleneck (0.0 to 1.0)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "bottleneck_type"},
		),
		BottleneckConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_bottleneck_confidence",
				Help: "Confidence level of bottleneck detection (0.0 to 1.0)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "bottleneck_type"},
		),
		BottleneckType: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_bottleneck_type_info",
				Help: "Type of bottleneck detected (informational, value always 1)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "bottleneck_type"},
		),
		PerformanceEfficiencyScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_performance_efficiency_score",
				Help: "Overall performance efficiency score from bottleneck analysis",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		PerformanceIssueCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_performance_issues_total",
				Help: "Number of performance issues detected",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "severity"},
		),
		OptimizationOpportunities: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_optimization_opportunities_total",
				Help: "Number of optimization opportunities identified",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "type"},
		),
		PredictedCompletionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_predicted_completion_timestamp",
				Help: "Predicted completion time as Unix timestamp",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		EstimatedTimeRemaining: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_estimated_time_remaining_seconds",
				Help: "Estimated time remaining for job completion in seconds",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		PerformanceTrendIndicator: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_performance_trend_indicator",
				Help: "Performance trend indicator (1=StateImproving, 0=StateStable, -1=degrading)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		RootCauseCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_root_causes_total",
				Help: "Number of root causes identified for performance issues",
			},
			[]string{"job_id", "step_id", "user", "account", "partition"},
		),
		RootCauseConfidence: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_job_root_cause_confidence",
				Help: "Confidence level of root cause analysis (0.0 to 1.0)",
			},
			[]string{"job_id", "step_id", "user", "account", "partition", "cause"},
		),
		AnalysisDuration: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "slurm_bottleneck_analysis_duration_seconds",
				Help:    "Time spent performing bottleneck analysis",
				Buckets: prometheus.DefBuckets,
			},
		),
		AnalysisErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_bottleneck_analysis_errors_total",
				Help: "Total number of bottleneck analysis errors",
			},
			[]string{"error_type"},
		),
		StepsAnalyzed: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_bottleneck_analysis_steps_processed_total",
				Help: "Total number of job steps analyzed for bottlenecks",
			},
		),
		BottlenecksFound: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "slurm_bottlenecks_detected_total",
				Help: "Total number of bottlenecks detected",
			},
		),
	}

	// Create efficiency calculator for performance analysis
	efficiencyCalc := NewEfficiencyCalculator(logger, nil)

	analyzer := &BottleneckAnalyzer{
		slurmClient:    slurmClient,
		logger:         logger,
		config:         config,
		metrics:        metrics,
		efficiencyCalc: efficiencyCalc,
		analysisCache:  make(map[string]*StepPerformanceAnalysis),
		cacheTTL:       config.CacheTTL,
		lastCollection: time.Time{},
	}

	return analyzer, nil
}

// Describe implements the prometheus.Collector interface
func (b *BottleneckAnalyzer) Describe(ch chan<- *prometheus.Desc) {
	b.metrics.BottleneckDetected.Describe(ch)
	b.metrics.BottleneckSeverity.Describe(ch)
	b.metrics.BottleneckConfidence.Describe(ch)
	b.metrics.BottleneckType.Describe(ch)
	b.metrics.PerformanceEfficiencyScore.Describe(ch)
	b.metrics.PerformanceIssueCount.Describe(ch)
	b.metrics.OptimizationOpportunities.Describe(ch)
	b.metrics.PredictedCompletionTime.Describe(ch)
	b.metrics.EstimatedTimeRemaining.Describe(ch)
	b.metrics.PerformanceTrendIndicator.Describe(ch)
	b.metrics.RootCauseCount.Describe(ch)
	b.metrics.RootCauseConfidence.Describe(ch)
	b.metrics.AnalysisDuration.Describe(ch)
	b.metrics.AnalysisErrors.Describe(ch)
	b.metrics.StepsAnalyzed.Describe(ch)
	b.metrics.BottlenecksFound.Describe(ch)
}

// Collect implements the prometheus.Collector interface
func (b *BottleneckAnalyzer) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()
	defer func() {
		b.metrics.AnalysisDuration.Observe(time.Since(start).Seconds())
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := b.analyzeBottlenecks(ctx); err != nil {
		b.logger.Error("Failed to analyze bottlenecks", "error", err)
		b.metrics.AnalysisErrors.WithLabelValues("analysis_failure").Inc()
	}

	// Collect metrics
	b.metrics.BottleneckDetected.Collect(ch)
	b.metrics.BottleneckSeverity.Collect(ch)
	b.metrics.BottleneckConfidence.Collect(ch)
	b.metrics.BottleneckType.Collect(ch)
	b.metrics.PerformanceEfficiencyScore.Collect(ch)
	b.metrics.PerformanceIssueCount.Collect(ch)
	b.metrics.OptimizationOpportunities.Collect(ch)
	b.metrics.PredictedCompletionTime.Collect(ch)
	b.metrics.EstimatedTimeRemaining.Collect(ch)
	b.metrics.PerformanceTrendIndicator.Collect(ch)
	b.metrics.RootCauseCount.Collect(ch)
	b.metrics.RootCauseConfidence.Collect(ch)
	b.metrics.AnalysisDuration.Collect(ch)
	b.metrics.AnalysisErrors.Collect(ch)
	b.metrics.StepsAnalyzed.Collect(ch)
	b.metrics.BottlenecksFound.Collect(ch)
}

// analyzeBottlenecks performs comprehensive bottleneck analysis
func (b *BottleneckAnalyzer) analyzeBottlenecks(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Get job manager
	jobManager := b.slurmClient.Jobs()
	if jobManager == nil {
		return fmt.Errorf("job manager not available")
	}

	// List running jobs for analysis
	// Using nil for options as the exact structure is not clear
	jobs, err := jobManager.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	b.logger.Debug("Analyzing bottlenecks", "job_count", len(jobs.Jobs))

	// Analyze each job
	for _, job := range jobs.Jobs {
		analysis := b.performStepPerformanceAnalysis(&job)

		// Cache the analysis
		cacheKey := fmt.Sprintf("%s:0", job.ID) // Step 0 for main job step
		b.analysisCache[cacheKey] = analysis

		// Update metrics
		b.updateAnalysisMetrics(&job, analysis)

		b.metrics.StepsAnalyzed.Inc()

		if analysis.PrimaryBottleneck != StateNone {
			b.metrics.BottlenecksFound.Inc()
		}
	}

	// Clean expired cache entries
	b.cleanExpiredCache()

	b.lastCollection = time.Now()
	return nil
}

// performStepPerformanceAnalysis performs comprehensive analysis on a job step
func (b *BottleneckAnalyzer) performStepPerformanceAnalysis(job *slurm.Job) *StepPerformanceAnalysis {
	analysis := &StepPerformanceAnalysis{
		JobID:             job.ID,
		StepID:            "0", // Main job step
		AnalysisTimestamp: time.Now(),
	}

	// Create resource utilization data for efficiency calculation
	utilizationData := CreateResourceUtilizationDataFromJob(job)

	// Calculate efficiency metrics
	efficiencyMetrics, err := b.efficiencyCalc.CalculateEfficiency(utilizationData)
	if err != nil {
		b.logger.Warn("Failed to calculate efficiency for bottleneck analysis", "job_id", job.ID, "error", err)
		// Use default values
		efficiencyMetrics = &EfficiencyMetrics{
			CPUEfficiency:     0.5,
			MemoryEfficiency:  0.5,
			IOEfficiency:      0.5,
			NetworkEfficiency: 0.5,
			OverallEfficiency: 0.5,
		}
	}

	// Set efficiency metrics
	analysis.CPUEfficiency = efficiencyMetrics.CPUEfficiency
	analysis.MemoryEfficiency = efficiencyMetrics.MemoryEfficiency
	analysis.IOEfficiency = efficiencyMetrics.IOEfficiency
	analysis.NetworkEfficiency = efficiencyMetrics.NetworkEfficiency
	analysis.OverallEfficiency = efficiencyMetrics.OverallEfficiency

	// Identify primary bottleneck
	analysis.PrimaryBottleneck, analysis.BottleneckSeverity, analysis.BottleneckConfidence = b.identifyPrimaryBottleneck(utilizationData, efficiencyMetrics)

	// Identify secondary bottlenecks
	analysis.SecondaryBottlenecks = b.identifySecondaryBottlenecks(utilizationData, efficiencyMetrics, analysis.PrimaryBottleneck)

	// Analyze utilization patterns
	analysis.CPUUtilizationPattern = b.analyzeUtilizationPattern("cpu", utilizationData)
	analysis.MemoryUtilizationPattern = b.analyzeUtilizationPattern("memory", utilizationData)
	analysis.IOUtilizationPattern = b.analyzeUtilizationPattern("io", utilizationData)
	analysis.NetworkUtilizationPattern = b.analyzeUtilizationPattern("network", utilizationData)

	// Detect performance issues
	analysis.PerformanceIssues = b.detectPerformanceIssues(utilizationData, efficiencyMetrics)

	// Identify optimization opportunities
	analysis.OptimizationOpportunities = b.identifyOptimizationOpportunities(utilizationData, efficiencyMetrics, analysis.PerformanceIssues)

	// Predictive analysis
	if b.config.EnablePredictiveAnalysis {
		b.performPredictiveAnalysis(job, utilizationData, analysis)
	}

	// Root cause analysis
	if b.config.EnableRootCauseAnalysis {
		analysis.RootCauses = b.performRootCauseAnalysis(utilizationData, efficiencyMetrics, analysis.PerformanceIssues)
	}

	// Generate recommended actions
	analysis.RecommendedActions = b.generateRecommendedActions(analysis)

	return analysis
}

// identifyPrimaryBottleneck identifies the primary bottleneck in the job step
func (b *BottleneckAnalyzer) identifyPrimaryBottleneck(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) (string, float64, float64) {
	// Calculate bottleneck scores for each resource type
	scores := map[string]float64{
		"cpu":     b.calculateCPUBottleneckScore(data, efficiency),
		"memory":  b.calculateMemoryBottleneckScore(data, efficiency),
		"io":      b.calculateIOBottleneckScore(data, efficiency),
		"network": b.calculateNetworkBottleneckScore(data, efficiency),
	}

	// Find the highest scoring bottleneck
	maxScore := 0.0
	primaryBottleneck := StateNone

	for resource, score := range scores {
		if score > maxScore {
			maxScore = score
			primaryBottleneck = resource
		}
	}

	// Set minimum threshold for bottleneck detection
	if maxScore < 0.3 {
		primaryBottleneck = StateNone
		maxScore = 0.0
	}

	// Calculate confidence based on how much the primary bottleneck exceeds others
	confidence := b.calculateBottleneckConfidence(scores, primaryBottleneck)

	return primaryBottleneck, maxScore, confidence
}

// identifySecondaryBottlenecks identifies secondary bottlenecks
func (b *BottleneckAnalyzer) identifySecondaryBottlenecks(data *ResourceUtilizationData, efficiency *EfficiencyMetrics, primary string) []string {
	_ = data
	secondary := []string{}

	if primary != "cpu" && efficiency.CPUEfficiency < 0.6 {
		secondary = append(secondary, "cpu")
	}
	if primary != "memory" && efficiency.MemoryEfficiency < 0.6 {
		secondary = append(secondary, "memory")
	}
	if primary != "io" && efficiency.IOEfficiency < 0.6 {
		secondary = append(secondary, "io")
	}
	if primary != "network" && efficiency.NetworkEfficiency < 0.6 {
		secondary = append(secondary, "network")
	}

	return secondary
}

// calculateCPUBottleneckScore calculates CPU bottleneck score
func (b *BottleneckAnalyzer) calculateCPUBottleneckScore(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) float64 {
	if data.CPUAllocated <= 0 {
		return 0.0
	}

	cpuUtil := data.CPUUsed / data.CPUAllocated

	// High CPU utilization with low efficiency indicates bottleneck
	if cpuUtil > b.config.CPUBottleneckThreshold && efficiency.CPUEfficiency < 0.7 {
		return cpuUtil * (1.0 - efficiency.CPUEfficiency)
	}

	return 0.0
}

// calculateMemoryBottleneckScore calculates memory bottleneck score
func (b *BottleneckAnalyzer) calculateMemoryBottleneckScore(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) float64 {
	if data.MemoryAllocated <= 0 {
		return 0.0
	}

	memoryUtil := float64(data.MemoryUsed) / float64(data.MemoryAllocated)

	// High memory utilization indicates bottleneck
	if memoryUtil > b.config.MemoryBottleneckThreshold {
		return memoryUtil * (1.0 - efficiency.MemoryEfficiency)
	}

	return 0.0
}

// calculateIOBottleneckScore calculates I/O bottleneck score
func (b *BottleneckAnalyzer) calculateIOBottleneckScore(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) float64 {
	if data.WallTime <= 0 {
		return 0.0
	}

	ioWaitRatio := data.IOWaitTime / data.WallTime

	// High I/O wait indicates bottleneck
	if ioWaitRatio > b.config.IOBottleneckThreshold {
		return ioWaitRatio * (1.0 - efficiency.IOEfficiency)
	}

	return 0.0
}

// calculateNetworkBottleneckScore calculates network bottleneck score
func (b *BottleneckAnalyzer) calculateNetworkBottleneckScore(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) float64 {
	if data.WallTime <= 0 {
		return 0.0
	}

	totalBytes := data.NetworkRxBytes + data.NetworkTxBytes
	if totalBytes == 0 {
		return 0.0
	}

	// Calculate network utilization (simplified)
	networkThroughput := float64(totalBytes) / data.WallTime

	// If network throughput is very high and efficiency is low, it might be a bottleneck
	if networkThroughput > 100*1024*1024 && efficiency.NetworkEfficiency < 0.7 { // 100 MB/s threshold
		return (1.0 - efficiency.NetworkEfficiency) * 0.8 // Lower weight for network bottlenecks
	}

	return 0.0
}

// calculateBottleneckConfidence calculates confidence in bottleneck detection
func (b *BottleneckAnalyzer) calculateBottleneckConfidence(scores map[string]float64, primary string) float64 {
	if primary == StateNone {
		return 0.0
	}

	primaryScore := scores[primary]
	totalOtherScores := 0.0

	for resource, score := range scores {
		if resource != primary {
			totalOtherScores += score
		}
	}

	// Confidence is higher when primary bottleneck clearly dominates
	if totalOtherScores == 0 {
		return 1.0
	}

	confidence := primaryScore / (primaryScore + totalOtherScores)
	return math.Min(confidence, 1.0)
}

// Additional methods would continue here for pattern analysis, issue detection,
// optimization opportunities, predictive analysis, root cause analysis, etc.
// These are simplified implementations for the core functionality.

// analyzeUtilizationPattern analyzes resource utilization patterns
func (b *BottleneckAnalyzer) analyzeUtilizationPattern(resourceType string, data *ResourceUtilizationData) string {
	// Simplified pattern analysis - in a real implementation, this would analyze
	// time series data to determine patterns
	switch resourceType {
	case "cpu":
		if data.CPUAllocated > 0 {
			util := data.CPUUsed / data.CPUAllocated
			if util > 0.8 {
				return StateStable
			} else if util < 0.3 {
				return "low"
			}
		}
		return "moderate"
	case "memory":
		if data.MemoryAllocated > 0 {
			util := float64(data.MemoryUsed) / float64(data.MemoryAllocated)
			if util > 0.8 {
				return "high"
			} else if util < 0.3 {
				return "low"
			}
		}
		return "moderate"
	default:
		return StateUnknown
	}
}

// detectPerformanceIssues detects various performance issues
func (b *BottleneckAnalyzer) detectPerformanceIssues(data *ResourceUtilizationData, efficiency *EfficiencyMetrics) []PerformanceIssue {
	issues := []PerformanceIssue{}

	// CPU underutilization
	if efficiency.CPUEfficiency < 0.4 {
		issues = append(issues, PerformanceIssue{
			Type:        "cpu_underutilization",
			Severity:    b.getSeverityLevel(1.0 - efficiency.CPUEfficiency),
			Description: "CPU resources are significantly underutilized",
			Impact:      "Wasted computational resources and reduced cost efficiency",
			FirstSeen:   time.Now(),
			Frequency:   1.0 - efficiency.CPUEfficiency,
		})
	}

	// Memory pressure
	if data.MemoryAllocated > 0 {
		memUtil := float64(data.MemoryUsed) / float64(data.MemoryAllocated)
		if memUtil > 0.9 {
			issues = append(issues, PerformanceIssue{
				Type:        "memory_pressure",
				Severity:    b.getSeverityLevel(memUtil - 0.7),
				Description: "Memory usage is very high, indicating potential memory pressure",
				Impact:      "May cause swapping, performance degradation, or job failure",
				FirstSeen:   time.Now(),
				Frequency:   memUtil,
			})
		}
	}

	return issues
}

// identifyOptimizationOpportunities identifies optimization opportunities
func (b *BottleneckAnalyzer) identifyOptimizationOpportunities(data *ResourceUtilizationData, efficiency *EfficiencyMetrics, issues []PerformanceIssue) []OptimizationOpportunity {
	_ = issues
	opportunities := []OptimizationOpportunity{}

	// Resource reduction opportunities
	if efficiency.CPUEfficiency < 0.5 && data.CPUAllocated > 1 {
		opportunities = append(opportunities, OptimizationOpportunity{
			Type:             "resource_reduction",
			Description:      "Reduce CPU allocation to match actual usage patterns",
			PotentialSavings: (0.5 - efficiency.CPUEfficiency) * 0.8,
			Effort:           "low",
			Priority:         2,
		})
	}

	return opportunities
}

// performPredictiveAnalysis performs predictive analysis
func (b *BottleneckAnalyzer) performPredictiveAnalysis(job *slurm.Job, data *ResourceUtilizationData, analysis *StepPerformanceAnalysis) {
	// Simplified predictive analysis
	if job.StartTime != nil && data.WallTime > 0 {
		// Estimate remaining time based on current progress (simplified)
		if job.TimeLimit > 0 {
			timeLimit := float64(job.TimeLimit) * 60 // Convert to seconds
			progress := data.WallTime / timeLimit
			if progress > 0 && progress < 1 {
				estimatedTotal := data.WallTime / progress
				analysis.EstimatedTimeRemaining = estimatedTotal - data.WallTime

				predictedCompletion := time.Now().Add(time.Duration(analysis.EstimatedTimeRemaining) * time.Second)
				analysis.PredictedCompletion = &predictedCompletion
			}
		}

		// Performance trend analysis (simplified)
		if analysis.OverallEfficiency > 0.7 {
			analysis.PerformanceTrend = StateStable
		} else if analysis.OverallEfficiency < 0.4 {
			analysis.PerformanceTrend = "degrading"
		} else {
			analysis.PerformanceTrend = StateStable
		}
	}
}

// performRootCauseAnalysis performs root cause analysis
func (b *BottleneckAnalyzer) performRootCauseAnalysis(data *ResourceUtilizationData, efficiency *EfficiencyMetrics, issues []PerformanceIssue) []RootCause {
	_ = data
	_ = efficiency
	causes := []RootCause{}

	// Analyze root causes based on detected issues
	for _, issue := range issues {
		switch issue.Type {
		case "cpu_underutilization":
			causes = append(causes, RootCause{
				Cause:              "Over-allocation of CPU resources",
				Confidence:         0.8,
				ContributionFactor: 0.9,
				Evidence:           []string{"Low CPU efficiency", "High allocated vs used CPU ratio"},
			})
		case "memory_pressure":
			causes = append(causes, RootCause{
				Cause:              "Insufficient memory allocation",
				Confidence:         0.9,
				ContributionFactor: 0.95,
				Evidence:           []string{"High memory utilization", "Memory usage near allocation limit"},
			})
		}
	}

	return causes
}

// generateRecommendedActions generates recommended actions
func (b *BottleneckAnalyzer) generateRecommendedActions(analysis *StepPerformanceAnalysis) []string {
	actions := []string{}

	// Actions based on primary bottleneck
	switch analysis.PrimaryBottleneck {
	case "cpu":
		actions = append(actions, "Optimize CPU-intensive operations or increase CPU allocation")
	case "memory":
		actions = append(actions, "Increase memory allocation or optimize memory usage patterns")
	case "io":
		actions = append(actions, "Optimize I/O operations or use faster storage")
	case "network":
		actions = append(actions, "Optimize network communication or use faster network infrastructure")
	}

	// Actions based on performance issues
	for _, issue := range analysis.PerformanceIssues {
		switch issue.Type {
		case "cpu_underutilization":
			actions = append(actions, "Consider reducing CPU allocation for better resource efficiency")
		case "memory_pressure":
			actions = append(actions, "Increase memory allocation to prevent performance degradation")
		}
	}

	if len(actions) == 0 {
		actions = append(actions, "Performance is within acceptable parameters")
	}

	return actions
}

// updateAnalysisMetrics updates Prometheus metrics from analysis results
func (b *BottleneckAnalyzer) updateAnalysisMetrics(job *slurm.Job, analysis *StepPerformanceAnalysis) {
	labels := []string{
		analysis.JobID,
		analysis.StepID,
		"", // TODO: job.UserName field not available
		"", // TODO: job.Account field not available
		job.Partition,
	}

	// Bottleneck detection metrics
	bottleneckValue := 0.0
	if analysis.PrimaryBottleneck != StateNone {
		bottleneckValue = 1.0
	}
	b.metrics.BottleneckDetected.WithLabelValues(labels...).Set(bottleneckValue)

	if analysis.PrimaryBottleneck != StateNone {
		severityLabels := append(labels, analysis.PrimaryBottleneck)
		b.metrics.BottleneckSeverity.WithLabelValues(severityLabels...).Set(analysis.BottleneckSeverity)
		b.metrics.BottleneckConfidence.WithLabelValues(severityLabels...).Set(analysis.BottleneckConfidence)

		typeLabels := append(labels, analysis.PrimaryBottleneck)
		b.metrics.BottleneckType.WithLabelValues(typeLabels...).Set(1.0)
	}

	// Performance metrics
	b.metrics.PerformanceEfficiencyScore.WithLabelValues(labels...).Set(analysis.OverallEfficiency)

	// Performance issues count by severity
	severityCounts := make(map[string]int)
	for _, issue := range analysis.PerformanceIssues {
		severityCounts[issue.Severity]++
	}
	for severity, count := range severityCounts {
		issueLabels := append(labels, severity)
		b.metrics.PerformanceIssueCount.WithLabelValues(issueLabels...).Set(float64(count))
	}

	// Optimization opportunities by type
	oppTypeCounts := make(map[string]int)
	for _, opp := range analysis.OptimizationOpportunities {
		oppTypeCounts[opp.Type]++
	}
	for oppType, count := range oppTypeCounts {
		oppLabels := append(labels, oppType)
		b.metrics.OptimizationOpportunities.WithLabelValues(oppLabels...).Set(float64(count))
	}

	// Predictive metrics
	if analysis.PredictedCompletion != nil {
		b.metrics.PredictedCompletionTime.WithLabelValues(labels...).Set(float64(analysis.PredictedCompletion.Unix()))
	}
	if analysis.EstimatedTimeRemaining > 0 {
		b.metrics.EstimatedTimeRemaining.WithLabelValues(labels...).Set(analysis.EstimatedTimeRemaining)
	}

	// Performance trend
	trendValue := 0.0
	switch analysis.PerformanceTrend {
	case StateImproving:
		trendValue = 1.0
	case "degrading":
		trendValue = -1.0
	}
	b.metrics.PerformanceTrendIndicator.WithLabelValues(labels...).Set(trendValue)

	// Root cause metrics
	b.metrics.RootCauseCount.WithLabelValues(labels...).Set(float64(len(analysis.RootCauses)))
	for _, cause := range analysis.RootCauses {
		causeLabels := append(labels, cause.Cause)
		b.metrics.RootCauseConfidence.WithLabelValues(causeLabels...).Set(cause.Confidence)
	}
}

// getSeverityLevel converts numeric severity to string level
func (b *BottleneckAnalyzer) getSeverityLevel(severity float64) string {
	if severity >= 0.8 {
		return "critical"
	}
	if severity >= 0.6 {
		return "high"
	}
	if severity >= 0.4 {
		return "medium"
	}
	return "low"
}

// cleanExpiredCache removes expired analysis results from cache
func (b *BottleneckAnalyzer) cleanExpiredCache() {
	now := time.Now()
	for key, analysis := range b.analysisCache {
		if now.Sub(analysis.AnalysisTimestamp) > b.cacheTTL {
			delete(b.analysisCache, key)
		}
	}
}

// GetCacheSize returns the current size of the analysis cache
func (b *BottleneckAnalyzer) GetCacheSize() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.analysisCache)
}

// GetLastCollection returns the timestamp of the last successful analysis
func (b *BottleneckAnalyzer) GetLastCollection() time.Time {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.lastCollection
}
