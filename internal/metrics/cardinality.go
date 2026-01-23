// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// CardinalityManager manages metric cardinality for large cluster scalability
type CardinalityManager struct {
	mu      sync.RWMutex
	limits  map[string]CardinalityLimit
	usage   map[string]*CardinalityUsage
	filters map[string]CardinalityFilter
	logger  *logrus.Logger

	// Configuration
	globalLimit    int
	checkInterval  time.Duration
	alertThreshold float64 // Percentage threshold for alerts

	// Metrics
	cardinalityMetrics *CardinalityMetrics
}

// CardinalityLimit defines limits for a specific metric or metric pattern
type CardinalityLimit struct {
	MetricPattern  string         // Regex pattern or exact name
	MaxSeries      int            // Maximum number of metric series
	MaxLabels      int            // Maximum number of unique label combinations
	MaxLabelValues map[string]int // Per-label value limits
	SamplingRate   float64        // 0.0-1.0, for high-cardinality metrics
	TTL            time.Duration  // Time-to-live for metric series
	Priority       Priority       // Priority level for enforcement
}

// CardinalityUsage tracks current usage for a metric
type CardinalityUsage struct {
	MetricName    string
	SeriesCount   int
	LabelCount    int
	LabelValues   map[string]int
	LastUpdate    time.Time
	SampledSeries int // Number of series that were sampled/dropped
}

// CardinalityFilter implements filtering strategies for high-cardinality metrics
type CardinalityFilter struct {
	Strategy         FilterStrategy
	SampleRate       float64
	LabelFilters     map[string][]string // Allowed values per label
	DropPatterns     []string            // Patterns to always drop
	PreservePatterns []string            // Patterns to always preserve
}

// FilterStrategy defines how to handle high-cardinality metrics
type FilterStrategy string

const (
	FilterStrategySample    FilterStrategy = "sample"    // Random sampling
	FilterStrategyDrop      FilterStrategy = "drop"      // Drop excess metrics
	FilterStrategyAggregate FilterStrategy = "aggregate" // Aggregate by fewer labels
	FilterStrategyLimit     FilterStrategy = "limit"     // Hard limit with FIFO
)

// Priority defines enforcement priority levels
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

// CardinalityMetrics contains Prometheus metrics for cardinality monitoring
type CardinalityMetrics struct {
	SeriesCount      *prometheus.GaugeVec
	LimitUtilization *prometheus.GaugeVec
	SampledSeries    *prometheus.CounterVec
	DroppedSeries    *prometheus.CounterVec
	LimitViolations  *prometheus.CounterVec
}

// NewCardinalityManager creates a new cardinality manager
func NewCardinalityManager(logger *logrus.Logger) *CardinalityManager {
	cm := &CardinalityManager{
		limits:         make(map[string]CardinalityLimit),
		usage:          make(map[string]*CardinalityUsage),
		filters:        make(map[string]CardinalityFilter),
		logger:         logger,
		globalLimit:    1000000, // 1M series default
		checkInterval:  30 * time.Second,
		alertThreshold: 0.8, // 80%
	}

	cm.cardinalityMetrics = cm.createCardinalityMetrics()
	cm.setupDefaultLimits()

	return cm
}

// createCardinalityMetrics creates Prometheus metrics for cardinality monitoring
func (cm *CardinalityManager) createCardinalityMetrics() *CardinalityMetrics {
	return &CardinalityMetrics{
		SeriesCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_cardinality_series",
				Help: "Current number of metric series per metric name",
			},
			[]string{"metric_name"},
		),
		LimitUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_cardinality_limit_utilization",
				Help: "Percentage of cardinality limit utilized per metric",
			},
			[]string{"metric_name", "limit_type"},
		),
		SampledSeries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_cardinality_sampled_series_total",
				Help: "Total number of metric series that were sampled due to cardinality limits",
			},
			[]string{"metric_name", "strategy"},
		),
		DroppedSeries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_cardinality_dropped_series_total",
				Help: "Total number of metric series that were dropped due to cardinality limits",
			},
			[]string{"metric_name", "reason"},
		),
		LimitViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_cardinality_limit_violations_total",
				Help: "Total number of cardinality limit violations",
			},
			[]string{"metric_name", "limit_type", "severity"},
		),
	}
}

// setupDefaultLimits configures reasonable default cardinality limits
func (cm *CardinalityManager) setupDefaultLimits() {
	// High-cardinality metrics (job-level)
	cm.limits["slurm_job_.*"] = CardinalityLimit{
		MetricPattern: "slurm_job_.*",
		MaxSeries:     50000,
		MaxLabels:     8,
		MaxLabelValues: map[string]int{
			"job_id": 10000,
			"user":   1000,
		},
		SamplingRate: 0.1, // Sample 10% of job metrics
		TTL:          24 * time.Hour,
		Priority:     PriorityMedium,
	}

	// Medium-cardinality metrics (node-level)
	cm.limits["slurm_node_.*"] = CardinalityLimit{
		MetricPattern: "slurm_node_.*",
		MaxSeries:     5000,
		MaxLabels:     6,
		MaxLabelValues: map[string]int{
			"node_name": 1000,
		},
		SamplingRate: 1.0, // Keep all node metrics
		TTL:          1 * time.Hour,
		Priority:     PriorityHigh,
	}

	// Low-cardinality metrics (cluster/partition-level)
	cm.limits["slurm_cluster_.*|slurm_partition_.*"] = CardinalityLimit{
		MetricPattern: "slurm_cluster_.*|slurm_partition_.*",
		MaxSeries:     500,
		MaxLabels:     5,
		MaxLabelValues: map[string]int{
			"partition": 100,
		},
		SamplingRate: 1.0, // Keep all cluster/partition metrics
		TTL:          5 * time.Minute,
		Priority:     PriorityCritical,
	}

	// User/account metrics with moderate limits
	cm.limits["slurm_user_.*|slurm_account_.*"] = CardinalityLimit{
		MetricPattern: "slurm_user_.*|slurm_account_.*",
		MaxSeries:     10000,
		MaxLabels:     4,
		MaxLabelValues: map[string]int{
			"user":    1000,
			"account": 500,
		},
		SamplingRate: 0.5, // Sample 50% of user metrics
		TTL:          6 * time.Hour,
		Priority:     PriorityMedium,
	}

	// Exporter self-monitoring (always preserve)
	cm.limits["slurm_exporter_.*"] = CardinalityLimit{
		MetricPattern: "slurm_exporter_.*",
		MaxSeries:     100,
		MaxLabels:     4,
		SamplingRate:  1.0,
		TTL:           1 * time.Minute,
		Priority:      PriorityCritical,
	}
}

// SetLimit configures a cardinality limit for a metric pattern
func (cm *CardinalityManager) SetLimit(pattern string, limit CardinalityLimit) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	limit.MetricPattern = pattern
	cm.limits[pattern] = limit

	cm.logger.WithFields(logrus.Fields{
		"pattern":    pattern,
		"max_series": limit.MaxSeries,
		"sampling":   limit.SamplingRate,
		"priority":   limit.Priority,
	}).Info("Cardinality limit configured")
}

// SetFilter configures a cardinality filter for a metric pattern
func (cm *CardinalityManager) SetFilter(pattern string, filter CardinalityFilter) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.filters[pattern] = filter

	cm.logger.WithFields(logrus.Fields{
		"pattern":  pattern,
		"strategy": filter.Strategy,
		"rate":     filter.SampleRate,
	}).Info("Cardinality filter configured")
}

// ShouldCollectMetric determines if a metric should be collected based on cardinality limits
func (cm *CardinalityManager) ShouldCollectMetric(metricName string, labels map[string]string) bool {
	// Find applicable limit (read-only operation)
	cm.mu.RLock()
	limit, found := cm.findApplicableLimit(metricName)
	cm.mu.RUnlock()

	if !found {
		// Update usage tracking for metrics without limits
		cm.updateUsage(metricName, labels)
		return true // No limit configured, allow collection
	}

	// Check if metric should be sampled
	if limit.SamplingRate < 1.0 {
		// Use deterministic sampling based on label hash for consistency
		if !cm.shouldSample(metricName, labels, limit.SamplingRate) {
			cm.cardinalityMetrics.SampledSeries.WithLabelValues(metricName, "sampling").Inc()
			return false
		}
	}

	// Check current usage (read-only)
	cm.mu.RLock()
	usage := cm.usage[metricName]
	cm.mu.RUnlock()

	// Check series limit
	if usage != nil && limit.MaxSeries > 0 && usage.SeriesCount >= limit.MaxSeries {
		cm.cardinalityMetrics.DroppedSeries.WithLabelValues(metricName, "series_limit").Inc()
		cm.cardinalityMetrics.LimitViolations.WithLabelValues(metricName, "series", string(limit.Priority)).Inc()

		if limit.Priority == PriorityCritical {
			cm.logger.WithFields(logrus.Fields{
				"metric": metricName,
				"series": usage.SeriesCount,
				"limit":  limit.MaxSeries,
			}).Error("Critical cardinality limit exceeded")
		}

		return false
	}

	// Check label-specific limits
	for labelName := range labels {
		if maxValues, exists := limit.MaxLabelValues[labelName]; exists {
			if usage != nil && usage.LabelValues[labelName] >= maxValues {
				cm.cardinalityMetrics.DroppedSeries.WithLabelValues(metricName, "label_limit").Inc()
				return false
			}
		}
	}

	// Update usage tracking after all checks pass
	cm.updateUsage(metricName, labels)

	return true
}

// findApplicableLimit finds the most specific limit that applies to a metric name
func (cm *CardinalityManager) findApplicableLimit(metricName string) (CardinalityLimit, bool) {
	// For now, use simple string matching. In production, would use regex
	for pattern, limit := range cm.limits {
		if cm.matchesPattern(metricName, pattern) {
			return limit, true
		}
	}
	return CardinalityLimit{}, false
}

// matchesPattern checks if a metric name matches a pattern (simplified)
func (cm *CardinalityManager) matchesPattern(metricName, pattern string) bool {
	// Simplified pattern matching - in production would use regexp
	if pattern == metricName {
		return true
	}

	// Handle simple wildcard patterns
	if len(pattern) > 2 && pattern[len(pattern)-2:] == ".*" {
		prefix := pattern[:len(pattern)-2]
		return len(metricName) >= len(prefix) && metricName[:len(prefix)] == prefix
	}

	return false
}

// shouldSample determines if a metric should be sampled based on label hash
func (cm *CardinalityManager) shouldSample(metricName string, labels map[string]string, rate float64) bool {
	// Create deterministic hash from metric name and labels
	hash := cm.hashLabels(metricName, labels)

	// Use modulo operation for consistent sampling
	return float64(hash%1000)/1000.0 < rate
}

// hashLabels creates a simple hash from metric name and labels
func (cm *CardinalityManager) hashLabels(metricName string, labels map[string]string) uint32 {
	// Simple hash function for demonstration
	var hash uint32 = 5381

	// Hash metric name
	for _, c := range metricName {
		hash = ((hash << 5) + hash) + uint32(c)
	}

	// Sort labels for consistent hashing
	var keys []string
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Hash sorted label pairs
	for _, k := range keys {
		v := labels[k]
		for _, c := range k + "=" + v {
			hash = ((hash << 5) + hash) + uint32(c)
		}
	}

	return hash
}

// updateUsage updates cardinality usage statistics
func (cm *CardinalityManager) updateUsage(metricName string, labels map[string]string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	usage, exists := cm.usage[metricName]
	if !exists {
		usage = &CardinalityUsage{
			MetricName:  metricName,
			LabelValues: make(map[string]int),
		}
		cm.usage[metricName] = usage
	}

	usage.SeriesCount++
	usage.LabelCount = len(labels)
	usage.LastUpdate = time.Now()

	// Track unique label values
	for labelName := range labels {
		usage.LabelValues[labelName]++
	}

	// Update Prometheus metrics
	cm.cardinalityMetrics.SeriesCount.WithLabelValues(metricName).Set(float64(usage.SeriesCount))
}

// GetCardinalityReport generates a comprehensive cardinality report
func (cm *CardinalityManager) GetCardinalityReport() CardinalityReport {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	report := CardinalityReport{
		GeneratedAt:     time.Now(),
		GlobalStats:     cm.calculateGlobalStats(),
		MetricStats:     make([]MetricCardinalityStats, 0, len(cm.usage)),
		Violations:      cm.findViolations(),
		Recommendations: cm.generateRecommendations(),
	}

	// Collect per-metric statistics
	for metricName, usage := range cm.usage {
		limit, hasLimit := cm.findApplicableLimit(metricName)

		stats := MetricCardinalityStats{
			MetricName:  metricName,
			SeriesCount: usage.SeriesCount,
			LabelCount:  usage.LabelCount,
			LastUpdate:  usage.LastUpdate,
			HasLimit:    hasLimit,
		}

		if hasLimit {
			stats.SeriesLimit = limit.MaxSeries
			stats.Utilization = float64(usage.SeriesCount) / float64(limit.MaxSeries) * 100
			stats.SamplingRate = limit.SamplingRate
		}

		report.MetricStats = append(report.MetricStats, stats)
	}

	// Sort by series count descending
	sort.Slice(report.MetricStats, func(i, j int) bool {
		return report.MetricStats[i].SeriesCount > report.MetricStats[j].SeriesCount
	})

	return report
}

// CardinalityReport contains comprehensive cardinality analysis
type CardinalityReport struct {
	GeneratedAt     time.Time
	GlobalStats     GlobalCardinalityStats
	MetricStats     []MetricCardinalityStats
	Violations      []CardinalityViolation
	Recommendations []CardinalityRecommendation
}

// GlobalCardinalityStats contains cluster-wide cardinality statistics
type GlobalCardinalityStats struct {
	TotalSeries            int
	TotalMetrics           int
	AverageSeriesPerMetric float64
	HighestCardinality     string
	GlobalUtilization      float64
}

// MetricCardinalityStats contains per-metric cardinality statistics
type MetricCardinalityStats struct {
	MetricName   string
	SeriesCount  int
	LabelCount   int
	LastUpdate   time.Time
	HasLimit     bool
	SeriesLimit  int
	Utilization  float64
	SamplingRate float64
}

// CardinalityViolation represents a cardinality limit violation
type CardinalityViolation struct {
	MetricName    string
	ViolationType string
	Current       int
	Limit         int
	Severity      Priority
	Message       string
}

// CardinalityRecommendation suggests cardinality optimizations
type CardinalityRecommendation struct {
	MetricName string
	Type       string
	Priority   Priority
	Message    string
	Action     string
}

// calculateGlobalStats computes global cardinality statistics
func (cm *CardinalityManager) calculateGlobalStats() GlobalCardinalityStats {
	totalSeries := 0
	maxSeries := 0
	maxMetric := ""

	for metricName, usage := range cm.usage {
		totalSeries += usage.SeriesCount
		if usage.SeriesCount > maxSeries {
			maxSeries = usage.SeriesCount
			maxMetric = metricName
		}
	}

	avgSeries := 0.0
	if len(cm.usage) > 0 {
		avgSeries = float64(totalSeries) / float64(len(cm.usage))
	}

	globalUtil := float64(totalSeries) / float64(cm.globalLimit) * 100

	return GlobalCardinalityStats{
		TotalSeries:            totalSeries,
		TotalMetrics:           len(cm.usage),
		AverageSeriesPerMetric: avgSeries,
		HighestCardinality:     maxMetric,
		GlobalUtilization:      globalUtil,
	}
}

// findViolations identifies current cardinality limit violations
func (cm *CardinalityManager) findViolations() []CardinalityViolation {
	var violations []CardinalityViolation

	for metricName, usage := range cm.usage {
		limit, hasLimit := cm.findApplicableLimit(metricName)
		if !hasLimit {
			continue
		}

		// Check series limit violations
		if limit.MaxSeries > 0 && usage.SeriesCount > limit.MaxSeries {
			violations = append(violations, CardinalityViolation{
				MetricName:    metricName,
				ViolationType: "series_limit",
				Current:       usage.SeriesCount,
				Limit:         limit.MaxSeries,
				Severity:      limit.Priority,
				Message:       fmt.Sprintf("Metric %s has %d series, exceeding limit of %d", metricName, usage.SeriesCount, limit.MaxSeries),
			})
		}

		// Check label limit violations
		if limit.MaxLabels > 0 && usage.LabelCount > limit.MaxLabels {
			violations = append(violations, CardinalityViolation{
				MetricName:    metricName,
				ViolationType: "label_limit",
				Current:       usage.LabelCount,
				Limit:         limit.MaxLabels,
				Severity:      limit.Priority,
				Message:       fmt.Sprintf("Metric %s has %d labels, exceeding limit of %d", metricName, usage.LabelCount, limit.MaxLabels),
			})
		}
	}

	return violations
}

// generateRecommendations creates cardinality optimization recommendations
func (cm *CardinalityManager) generateRecommendations() []CardinalityRecommendation {
	var recommendations []CardinalityRecommendation

	for metricName, usage := range cm.usage {
		limit, hasLimit := cm.findApplicableLimit(metricName)

		// Recommend limits for metrics without them
		if !hasLimit && usage.SeriesCount > 1000 {
			recommendations = append(recommendations, CardinalityRecommendation{
				MetricName: metricName,
				Type:       "add_limit",
				Priority:   PriorityMedium,
				Message:    fmt.Sprintf("Metric %s has high cardinality (%d series) but no limits configured", metricName, usage.SeriesCount),
				Action:     "Configure cardinality limits and sampling",
			})
		}

		// Recommend sampling for high-cardinality metrics
		if hasLimit && limit.SamplingRate == 1.0 && usage.SeriesCount > int(float64(limit.MaxSeries)*0.8) {
			recommendations = append(recommendations, CardinalityRecommendation{
				MetricName: metricName,
				Type:       "enable_sampling",
				Priority:   PriorityMedium,
				Message:    fmt.Sprintf("Metric %s is approaching cardinality limit, consider enabling sampling", metricName),
				Action:     "Set sampling rate to 0.1-0.5 depending on requirements",
			})
		}

		// Recommend aggregation for excessive label count
		if usage.LabelCount > 10 {
			recommendations = append(recommendations, CardinalityRecommendation{
				MetricName: metricName,
				Type:       "reduce_labels",
				Priority:   PriorityHigh,
				Message:    fmt.Sprintf("Metric %s has excessive labels (%d), consider aggregation", metricName, usage.LabelCount),
				Action:     "Remove non-essential labels or aggregate by fewer dimensions",
			})
		}
	}

	return recommendations
}

// StartCardinalityMonitoring begins background cardinality monitoring
func (cm *CardinalityManager) StartCardinalityMonitoring(ctx context.Context) {
	ticker := time.NewTicker(cm.checkInterval)
	defer ticker.Stop()

	cm.logger.Info("Starting cardinality monitoring")

	for {
		select {
		case <-ctx.Done():
			cm.logger.Info("Stopping cardinality monitoring")
			return
		case <-ticker.C:
			cm.performCardinalityCheck()
		}
	}
}

// performCardinalityCheck performs periodic cardinality analysis
func (cm *CardinalityManager) performCardinalityCheck() {
	report := cm.GetCardinalityReport()

	// Log global statistics
	cm.logger.WithFields(logrus.Fields{
		"total_series":  report.GlobalStats.TotalSeries,
		"total_metrics": report.GlobalStats.TotalMetrics,
		"global_util":   report.GlobalStats.GlobalUtilization,
		"highest_card":  report.GlobalStats.HighestCardinality,
	}).Info("Cardinality check completed")

	// Alert on violations
	for _, violation := range report.Violations {
		level := logrus.InfoLevel
		if violation.Severity == PriorityHigh || violation.Severity == PriorityCritical {
			level = logrus.WarnLevel
		}

		cm.logger.WithFields(logrus.Fields{
			"metric":   violation.MetricName,
			"type":     violation.ViolationType,
			"current":  violation.Current,
			"limit":    violation.Limit,
			"severity": violation.Severity,
		}).Log(level, violation.Message)
	}

	// Update utilization metrics
	for _, stats := range report.MetricStats {
		if stats.HasLimit {
			cm.cardinalityMetrics.LimitUtilization.WithLabelValues(
				stats.MetricName, "series",
			).Set(stats.Utilization)
		}
	}
}

// RegisterMetrics registers cardinality metrics with Prometheus
func (cm *CardinalityManager) RegisterMetrics(registry *prometheus.Registry) error {
	metrics := []*prometheus.GaugeVec{cm.cardinalityMetrics.SeriesCount, cm.cardinalityMetrics.LimitUtilization}
	counters := []*prometheus.CounterVec{
		cm.cardinalityMetrics.SampledSeries,
		cm.cardinalityMetrics.DroppedSeries,
		cm.cardinalityMetrics.LimitViolations,
	}

	for _, metric := range metrics {
		if err := registry.Register(metric); err != nil {
			return fmt.Errorf("failed to register cardinality gauge metric: %w", err)
		}
	}

	for _, metric := range counters {
		if err := registry.Register(metric); err != nil {
			return fmt.Errorf("failed to register cardinality counter metric: %w", err)
		}
	}

	return nil
}
