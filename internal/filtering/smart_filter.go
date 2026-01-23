// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package filtering

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
)

// MetricPattern represents a learned pattern for a specific metric
type MetricPattern struct {
	Name            string            `json:"name"`
	Labels          map[string]string `json:"labels"`
	Values          []float64         `json:"values"`
	Timestamps      []time.Time       `json:"timestamps"`
	Mean            float64           `json:"mean"`
	Variance        float64           `json:"variance"`
	ChangeRate      float64           `json:"change_rate"`
	NoiseScore      float64           `json:"noise_score"`
	Correlation     float64           `json:"correlation"`
	LastUpdated     time.Time         `json:"last_updated"`
	SampleCount     int               `json:"sample_count"`
	FilterRecommend FilterAction      `json:"filter_recommend"`
}

// FilterAction represents the recommended action for a metric
type FilterAction int

const (
	ActionKeep FilterAction = iota
	ActionFilter
	ActionReduce
	ActionUnknown
)

func (a FilterAction) String() string {
	switch a {
	case ActionKeep:
		return "keep"
	case ActionFilter:
		return "filter"
	case ActionReduce:
		return "reduce"
	default:
		return "unknown"
	}
}

// SmartFilter implements intelligent metric filtering with pattern learning
type SmartFilter struct {
	config  config.SmartFilteringConfig
	logger  *logrus.Logger
	metrics *FilterMetrics
	mu      sync.RWMutex

	// Pattern storage
	patterns       map[string]*MetricPattern
	patternHistory map[string][]float64

	// Learning state
	enabled         bool
	learningPhase   bool
	learningSince   time.Time
	totalSamples    int64
	filteredSamples int64

	// Cache for quick lookups
	filterCache map[string]FilterAction
	cacheExpiry time.Time

	// Background processing
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// FilterMetrics holds Prometheus metrics for the smart filter
type FilterMetrics struct {
	patternsLearned    *prometheus.GaugeVec
	metricsFiltered    *prometheus.CounterVec
	noiseScore         *prometheus.HistogramVec
	filterDecisions    *prometheus.CounterVec
	learningPhase      prometheus.Gauge
	cacheHitRate       prometheus.Gauge
	processingDuration *prometheus.HistogramVec
}

// NewSmartFilter creates a new smart filtering system
func NewSmartFilter(cfg config.SmartFilteringConfig, logger *logrus.Logger) (*SmartFilter, error) {
	if !cfg.Enabled {
		logger.Info("Smart filtering disabled")
		return &SmartFilter{
			config:      cfg,
			logger:      logger,
			enabled:     false,
			patterns:    make(map[string]*MetricPattern),
			filterCache: make(map[string]FilterAction),
		}, nil
	}

	// Validate configuration
	if err := validateFilterConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid smart filtering config: %w", err)
	}

	metrics := &FilterMetrics{
		patternsLearned: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_filter_patterns_learned",
				Help: "Number of metric patterns learned",
			},
			[]string{"action"},
		),
		metricsFiltered: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_filter_metrics_processed_total",
				Help: "Total number of metrics processed by smart filter",
			},
			[]string{"collector", "action"},
		),
		noiseScore: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_exporter_filter_noise_score",
				Help:    "Distribution of noise scores for metrics",
				Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
			},
			[]string{"collector"},
		),
		filterDecisions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_filter_decisions_total",
				Help: "Total number of filter decisions made",
			},
			[]string{"decision", "reason"},
		),
		learningPhase: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_filter_learning_phase",
				Help: "Whether the filter is in learning phase (1) or filtering phase (0)",
			},
		),
		cacheHitRate: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_filter_cache_hit_rate",
				Help: "Cache hit rate for filter decisions",
			},
		),
		processingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_exporter_filter_processing_duration_seconds",
				Help:    "Time spent processing metrics through the filter",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation"},
		),
	}

	ctx, cancel := context.WithCancel(context.Background())

	filter := &SmartFilter{
		config:         cfg,
		logger:         logger,
		metrics:        metrics,
		enabled:        true,
		learningPhase:  true,
		learningSince:  time.Now(),
		patterns:       make(map[string]*MetricPattern),
		patternHistory: make(map[string][]float64),
		filterCache:    make(map[string]FilterAction),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Start background processing
	filter.wg.Add(1)
	go filter.backgroundProcessor()

	// Set initial learning phase metric
	filter.metrics.learningPhase.Set(1)

	logger.WithFields(logrus.Fields{
		"noise_threshold": cfg.NoiseThreshold,
		"cache_size":      cfg.CacheSize,
		"learning_window": cfg.LearningWindow,
		"variance_limit":  cfg.VarianceLimit,
		"correlation_min": cfg.CorrelationMin,
	}).Info("Smart filtering initialized in learning mode")

	return filter, nil
}

// validateFilterConfig validates the smart filtering configuration
func validateFilterConfig(cfg config.SmartFilteringConfig) error {
	if cfg.NoiseThreshold < 0 || cfg.NoiseThreshold > 1 {
		return fmt.Errorf("noise_threshold must be between 0 and 1")
	}
	if cfg.CacheSize <= 0 {
		return fmt.Errorf("cache_size must be positive")
	}
	if cfg.LearningWindow <= 0 {
		return fmt.Errorf("learning_window must be positive")
	}
	if cfg.VarianceLimit < 0 {
		return fmt.Errorf("variance_limit must be non-negative")
	}
	if cfg.CorrelationMin < -1 || cfg.CorrelationMin > 1 {
		return fmt.Errorf("correlation_min must be between -1 and 1")
	}
	return nil
}

// RegisterMetrics registers the filter metrics with Prometheus
func (sf *SmartFilter) RegisterMetrics(registry prometheus.Registerer) error {
	if !sf.enabled {
		return nil
	}

	collectors := []prometheus.Collector{
		sf.metrics.patternsLearned,
		sf.metrics.metricsFiltered,
		sf.metrics.noiseScore,
		sf.metrics.filterDecisions,
		sf.metrics.learningPhase,
		sf.metrics.cacheHitRate,
		sf.metrics.processingDuration,
	}

	for i, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return fmt.Errorf("failed to register smart filter metric %d: %w", i, err)
		}
	}

	sf.logger.Info("Successfully registered smart filter metrics")
	return nil
}

// ProcessMetrics processes a set of metrics through the smart filter
func (sf *SmartFilter) ProcessMetrics(ctx context.Context, collector string, metrics []*dto.MetricFamily) ([]*dto.MetricFamily, error) {
	if !sf.enabled {
		return metrics, nil
	}

	start := time.Now()
	defer func() {
		sf.metrics.processingDuration.WithLabelValues("filter").Observe(time.Since(start).Seconds())
	}()

	var filteredFamilies []*dto.MetricFamily
	var totalMetrics, filteredCount int

	for _, family := range metrics {
		filteredFamily, count, familyFilteredCount := sf.processMetricFamily(ctx, collector, family)
		if filteredFamily != nil {
			filteredFamilies = append(filteredFamilies, filteredFamily)
		}
		totalMetrics += count
		filteredCount += familyFilteredCount
	}

	// Update metrics
	sf.metrics.metricsFiltered.WithLabelValues(collector, "kept").Add(float64(totalMetrics - filteredCount))
	sf.metrics.metricsFiltered.WithLabelValues(collector, "filtered").Add(float64(filteredCount))

	sf.mu.Lock()
	sf.totalSamples += int64(totalMetrics)
	sf.filteredSamples += int64(filteredCount)
	sf.mu.Unlock()

	sf.logger.WithFields(logrus.Fields{
		"collector":     collector,
		"total_metrics": totalMetrics,
		"filtered":      filteredCount,
		"kept":          totalMetrics - filteredCount,
		"filter_rate":   float64(filteredCount) / float64(totalMetrics) * 100,
	}).Debug("Processed metrics through smart filter")

	return filteredFamilies, nil
}

// processMetricFamily processes a single metric family
func (sf *SmartFilter) processMetricFamily(ctx context.Context, collector string, family *dto.MetricFamily) (*dto.MetricFamily, int, int) {
	if family == nil || len(family.GetMetric()) == 0 {
		return family, 0, 0
	}

	var filteredMetrics []*dto.Metric
	totalCount := len(family.GetMetric())
	filteredCount := 0

	for _, metric := range family.GetMetric() {
		if sf.shouldKeepMetric(ctx, collector, family.GetName(), metric) {
			filteredMetrics = append(filteredMetrics, metric)
		} else {
			filteredCount++
		}
	}

	if len(filteredMetrics) == 0 {
		return nil, totalCount, filteredCount
	}

	// Create new family with filtered metrics
	filteredFamily := &dto.MetricFamily{
		Name:   family.Name,
		Help:   family.Help,
		Type:   family.Type,
		Metric: filteredMetrics,
	}

	return filteredFamily, totalCount, filteredCount
}

// shouldKeepMetric determines if a metric should be kept or filtered
func (sf *SmartFilter) shouldKeepMetric(ctx context.Context, collector, metricName string, metric *dto.Metric) bool {
	// Create unique key for this metric
	key := sf.createMetricKey(metricName, metric)

	// Check cache first
	if action, found := sf.getCachedDecision(key); found {
		sf.recordDecision(action, "cache_hit")
		return action == ActionKeep
	}

	// Get or create pattern for this metric
	pattern := sf.getOrCreatePattern(key, metricName, metric)

	// Update pattern with current value
	value := sf.extractValue(metric)
	sf.updatePattern(pattern, value)

	// Make filtering decision
	action := sf.makeFilterDecision(pattern, collector)

	// Cache the decision
	sf.cacheDecision(key, action)

	// Record metrics
	sf.metrics.noiseScore.WithLabelValues(collector).Observe(pattern.NoiseScore)
	sf.recordDecision(action, "computed")

	return action == ActionKeep
}

// createMetricKey creates a unique key for a metric
func (sf *SmartFilter) createMetricKey(name string, metric *dto.Metric) string {
	if len(metric.GetLabel()) == 0 {
		return name
	}

	// Create deterministic key from labels
	key := name
	for _, label := range metric.GetLabel() {
		key += fmt.Sprintf("_%s=%s", label.GetName(), label.GetValue())
	}
	return key
}

// getCachedDecision retrieves a cached filter decision
func (sf *SmartFilter) getCachedDecision(key string) (FilterAction, bool) {
	sf.mu.RLock()
	defer sf.mu.RUnlock()

	// Check if cache is expired
	if time.Now().After(sf.cacheExpiry) {
		return ActionUnknown, false
	}

	action, found := sf.filterCache[key]
	return action, found
}

// cacheDecision caches a filter decision
func (sf *SmartFilter) cacheDecision(key string, action FilterAction) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	// Limit cache size
	if len(sf.filterCache) >= sf.config.CacheSize {
		// Simple eviction: clear half the cache
		for k := range sf.filterCache {
			delete(sf.filterCache, k)
			if len(sf.filterCache) <= sf.config.CacheSize/2 {
				break
			}
		}
	}

	sf.filterCache[key] = action

	// Set cache expiry if not set
	if sf.cacheExpiry.IsZero() {
		sf.cacheExpiry = time.Now().Add(5 * time.Minute)
	}
}

// getOrCreatePattern gets an existing pattern or creates a new one
func (sf *SmartFilter) getOrCreatePattern(key, name string, metric *dto.Metric) *MetricPattern {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	pattern, exists := sf.patterns[key]
	if !exists {
		// Extract labels
		labels := make(map[string]string)
		for _, label := range metric.GetLabel() {
			labels[label.GetName()] = label.GetValue()
		}

		pattern = &MetricPattern{
			Name:            name,
			Labels:          labels,
			Values:          make([]float64, 0, sf.config.LearningWindow),
			Timestamps:      make([]time.Time, 0, sf.config.LearningWindow),
			LastUpdated:     time.Now(),
			FilterRecommend: ActionUnknown,
		}
		sf.patterns[key] = pattern
	}

	return pattern
}

// extractValue extracts a numeric value from a metric
func (sf *SmartFilter) extractValue(metric *dto.Metric) float64 {
	switch {
	case metric.GetGauge() != nil:
		return metric.GetGauge().GetValue()
	case metric.GetCounter() != nil:
		return metric.GetCounter().GetValue()
	case metric.GetUntyped() != nil:
		return metric.GetUntyped().GetValue()
	case metric.GetHistogram() != nil:
		return float64(metric.GetHistogram().GetSampleCount())
	case metric.GetSummary() != nil:
		return float64(metric.GetSummary().GetSampleCount())
	default:
		return 0
	}
}

// updatePattern updates a metric pattern with a new value
func (sf *SmartFilter) updatePattern(pattern *MetricPattern, value float64) {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now()

	// Add new value
	pattern.Values = append(pattern.Values, value)
	pattern.Timestamps = append(pattern.Timestamps, now)
	pattern.LastUpdated = now
	pattern.SampleCount++

	// Maintain window size
	if len(pattern.Values) > sf.config.LearningWindow {
		pattern.Values = pattern.Values[1:]
		pattern.Timestamps = pattern.Timestamps[1:]
	}

	// Recalculate statistics if we have enough samples
	if len(pattern.Values) >= 3 {
		sf.calculateStatistics(pattern)
	}
}

// calculateStatistics calculates statistical measures for a pattern
func (sf *SmartFilter) calculateStatistics(pattern *MetricPattern) {
	values := pattern.Values
	n := len(values)

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	pattern.Mean = sum / float64(n)

	// Calculate variance
	variance := 0.0
	for _, v := range values {
		diff := v - pattern.Mean
		variance += diff * diff
	}
	pattern.Variance = variance / float64(n-1)

	// Calculate change rate (how much the metric changes)
	if n > 1 {
		changes := 0.0
		for i := 1; i < n; i++ {
			if values[i-1] != 0 {
				change := math.Abs((values[i] - values[i-1]) / values[i-1])
				changes += change
			}
		}
		pattern.ChangeRate = changes / float64(n-1)
	}

	// Calculate noise score (combination of variance and change rate)
	pattern.NoiseScore = sf.calculateNoiseScore(pattern)

	// Calculate correlation with time (trend detection)
	pattern.Correlation = sf.calculateTimeCorrelation(pattern)
}

// calculateNoiseScore calculates a noise score for a pattern
func (sf *SmartFilter) calculateNoiseScore(pattern *MetricPattern) float64 {
	// Normalize variance (higher variance = more noise)
	normalizedVariance := math.Min(pattern.Variance/sf.config.VarianceLimit, 1.0)

	// Normalize change rate (higher change rate = more noise)
	normalizedChangeRate := math.Min(pattern.ChangeRate, 1.0)

	// Combine factors (weighted average)
	noiseScore := 0.6*normalizedVariance + 0.4*normalizedChangeRate

	return math.Min(noiseScore, 1.0)
}

// calculateTimeCorrelation calculates correlation between values and time
func (sf *SmartFilter) calculateTimeCorrelation(pattern *MetricPattern) float64 {
	values := pattern.Values
	n := len(values)

	if n < 3 {
		return 0
	}

	// Create time series (0, 1, 2, ...)
	timeValues := make([]float64, n)
	for i := range timeValues {
		timeValues[i] = float64(i)
	}

	// Calculate correlation coefficient
	return sf.calculateCorrelation(timeValues, values)
}

// calculateCorrelation calculates Pearson correlation coefficient
func (sf *SmartFilter) calculateCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return 0
	}

	// Calculate means
	meanX := 0.0
	meanY := 0.0
	for i := 0; i < n; i++ {
		meanX += x[i]
		meanY += y[i]
	}
	meanX /= float64(n)
	meanY /= float64(n)

	// Calculate correlation
	numerator := 0.0
	sumSqX := 0.0
	sumSqY := 0.0

	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumSqX += dx * dx
		sumSqY += dy * dy
	}

	denominator := math.Sqrt(sumSqX * sumSqY)
	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// makeFilterDecision makes a filtering decision for a pattern
func (sf *SmartFilter) makeFilterDecision(pattern *MetricPattern, collector string) FilterAction {
	// During learning phase, keep everything
	if sf.learningPhase {
		return ActionKeep
	}

	// Check noise threshold
	if pattern.NoiseScore > sf.config.NoiseThreshold {
		// High noise - filter it
		return ActionFilter
	}

	// Check if metric has low variance and low correlation (constant value)
	if pattern.Variance < sf.config.VarianceLimit*0.1 && math.Abs(pattern.Correlation) < sf.config.CorrelationMin {
		// Constant metrics might be noise
		return ActionReduce
	}

	// Keep everything else
	return ActionKeep
}

// recordDecision records a filter decision for metrics
func (sf *SmartFilter) recordDecision(action FilterAction, reason string) {
	sf.metrics.filterDecisions.WithLabelValues(action.String(), reason).Inc()
}

// backgroundProcessor handles background tasks for the filter
func (sf *SmartFilter) backgroundProcessor() {
	defer sf.wg.Done()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-sf.ctx.Done():
			return
		case <-ticker.C:
			sf.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (sf *SmartFilter) performMaintenance() {
	sf.mu.Lock()
	defer sf.mu.Unlock()

	now := time.Now()

	// Check if we should exit learning phase
	if sf.learningPhase && now.Sub(sf.learningSince) > 10*time.Minute {
		sf.learningPhase = false
		sf.metrics.learningPhase.Set(0)
		sf.logger.Info("Smart filter exited learning phase and started filtering")
	}

	// Update pattern count metrics
	keepCount := 0
	filterCount := 0
	reduceCount := 0

	for _, pattern := range sf.patterns {
		switch pattern.FilterRecommend {
		case ActionKeep:
			keepCount++
		case ActionFilter:
			filterCount++
		case ActionReduce:
			reduceCount++
		}
	}

	sf.metrics.patternsLearned.WithLabelValues("keep").Set(float64(keepCount))
	sf.metrics.patternsLearned.WithLabelValues("filter").Set(float64(filterCount))
	sf.metrics.patternsLearned.WithLabelValues("reduce").Set(float64(reduceCount))

	// Calculate cache hit rate
	totalDecisions := sf.totalSamples
	if totalDecisions > 0 {
		// This is a simplified calculation; in practice you'd track cache hits separately
		cacheHitRate := 0.7 // Placeholder - implement proper cache hit tracking
		sf.metrics.cacheHitRate.Set(cacheHitRate)
	}

	// Clear expired cache entries
	if now.After(sf.cacheExpiry) {
		sf.filterCache = make(map[string]FilterAction)
		sf.cacheExpiry = time.Time{}
	}

	// Clean up old patterns
	sf.cleanupOldPatterns(now)
}

// cleanupOldPatterns removes patterns that haven't been updated recently
func (sf *SmartFilter) cleanupOldPatterns(now time.Time) {
	maxAge := 24 * time.Hour

	for key, pattern := range sf.patterns {
		if now.Sub(pattern.LastUpdated) > maxAge {
			delete(sf.patterns, key)
		}
	}
}

// GetStats returns statistics about the smart filter
func (sf *SmartFilter) GetStats() map[string]interface{} {
	if !sf.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	sf.mu.RLock()
	defer sf.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":          true,
		"learning_phase":   sf.learningPhase,
		"total_patterns":   len(sf.patterns),
		"cache_size":       len(sf.filterCache),
		"total_samples":    sf.totalSamples,
		"filtered_samples": sf.filteredSamples,
	}

	if sf.totalSamples > 0 {
		stats["filter_rate"] = float64(sf.filteredSamples) / float64(sf.totalSamples)
	}

	return stats
}

// GetPatterns returns a copy of learned patterns for inspection
func (sf *SmartFilter) GetPatterns() map[string]MetricPattern {
	if !sf.enabled {
		return nil
	}

	sf.mu.RLock()
	defer sf.mu.RUnlock()

	patterns := make(map[string]MetricPattern)
	for key, pattern := range sf.patterns {
		// Create a copy to prevent external modification
		patternCopy := *pattern
		patterns[key] = patternCopy
	}

	return patterns
}

// IsEnabled returns whether smart filtering is enabled
func (sf *SmartFilter) IsEnabled() bool {
	return sf.enabled
}

// Close gracefully shuts down the smart filter
func (sf *SmartFilter) Close() error {
	if sf.cancel != nil {
		sf.cancel()
		sf.wg.Wait()
	}

	sf.logger.Info("Smart filter shutdown complete")
	return nil
}
