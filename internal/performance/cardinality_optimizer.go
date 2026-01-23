// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// CardinalityOptimizer manages metric cardinality to prevent memory issues
type CardinalityOptimizer struct {
	logger *logrus.Entry

	// Configuration
	maxCardinality int
	sampleRate     float64
	enableSampling bool

	// Cardinality tracking
	mu                sync.RWMutex
	metricCardinality map[string]int
	labelCardinality  map[string]map[string]int
	lastCleanup       time.Time
	cleanupInterval   time.Duration

	// Sampling state
	samplingSeeds map[string]uint64

	// Metrics
	cardinalityTotal    prometheus.Gauge
	cardinalityByMetric *prometheus.GaugeVec
	sampledMetrics      prometheus.Counter
	droppedMetrics      prometheus.Counter
	cleanupDuration     prometheus.Histogram
}

// NewCardinalityOptimizer creates a new cardinality optimizer
func NewCardinalityOptimizer(maxCardinality int, sampleRate float64, logger *logrus.Entry) *CardinalityOptimizer {
	co := &CardinalityOptimizer{
		logger:            logger,
		maxCardinality:    maxCardinality,
		sampleRate:        sampleRate,
		enableSampling:    sampleRate < 1.0,
		metricCardinality: make(map[string]int),
		labelCardinality:  make(map[string]map[string]int),
		samplingSeeds:     make(map[string]uint64),
		cleanupInterval:   10 * time.Minute,
		lastCleanup:       time.Now(),
	}

	co.initMetrics()

	return co
}

// initMetrics initializes Prometheus metrics for cardinality monitoring
func (co *CardinalityOptimizer) initMetrics() {
	co.cardinalityTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slurm_exporter_cardinality_total",
		Help: "Total number of unique metric series",
	})

	co.cardinalityByMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slurm_exporter_cardinality_by_metric",
			Help: "Cardinality per metric name",
		},
		[]string{"metric_name"},
	)

	co.sampledMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_sampled_metrics_total",
		Help: "Total number of metrics that were sampled",
	})

	co.droppedMetrics = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_dropped_metrics_total",
		Help: "Total number of metrics dropped due to cardinality limits",
	})

	co.cleanupDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "slurm_exporter_cardinality_cleanup_duration_seconds",
		Help:    "Time spent cleaning up cardinality tracking",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
	})
}

// ShouldCollectMetric determines if a metric should be collected based on cardinality limits
func (co *CardinalityOptimizer) ShouldCollectMetric(metricName string, labels map[string]string) bool {
	co.mu.Lock()
	defer co.mu.Unlock()

	// Check total cardinality limit
	totalCardinality := co.getTotalCardinality()
	if totalCardinality >= co.maxCardinality {
		// Apply deterministic sampling based on metric signature
		if co.enableSampling {
			return co.shouldSampleMetric(metricName, labels)
		}

		co.droppedMetrics.Inc()
		return false
	}

	// Update cardinality tracking
	co.updateCardinality(metricName, labels)

	return true
}

// shouldSampleMetric determines if a metric should be sampled
func (co *CardinalityOptimizer) shouldSampleMetric(metricName string, labels map[string]string) bool {
	// Create deterministic hash from metric name and labels
	hash := co.hashMetric(metricName, labels)

	// Get or create sampling seed for this metric
	seed, exists := co.samplingSeeds[metricName]
	if !exists {
		seed = co.hashString(metricName)
		co.samplingSeeds[metricName] = seed
	}

	// Combine hash with seed for deterministic sampling
	finalHash := hash ^ seed
	threshold := uint64(float64(^uint64(0)) * co.sampleRate)

	shouldSample := finalHash <= threshold
	if shouldSample {
		co.sampledMetrics.Inc()
		co.updateCardinality(metricName, labels)
	} else {
		co.droppedMetrics.Inc()
	}

	return shouldSample
}

// hashMetric creates a deterministic hash for a metric and its labels
func (co *CardinalityOptimizer) hashMetric(metricName string, labels map[string]string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(metricName))

	// Sort labels for deterministic hashing
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(labels[k]))
	}

	return h.Sum64()
}

// hashString creates a hash from a string
func (co *CardinalityOptimizer) hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// updateCardinality updates cardinality tracking for a metric
func (co *CardinalityOptimizer) updateCardinality(metricName string, labels map[string]string) {
	// Update metric cardinality
	co.metricCardinality[metricName]++

	// Update label cardinality
	if co.labelCardinality[metricName] == nil {
		co.labelCardinality[metricName] = make(map[string]int)
	}

	for labelName, labelValue := range labels {
		key := fmt.Sprintf("%s=%s", labelName, labelValue)
		co.labelCardinality[metricName][key]++
	}
}

// getTotalCardinality calculates total cardinality across all metrics
func (co *CardinalityOptimizer) getTotalCardinality() int {
	total := 0
	for _, count := range co.metricCardinality {
		total += count
	}
	return total
}

// GetCardinalityStats returns current cardinality statistics
func (co *CardinalityOptimizer) GetCardinalityStats() CardinalityStats {
	co.mu.RLock()
	defer co.mu.RUnlock()

	stats := CardinalityStats{
		TotalCardinality: co.getTotalCardinality(),
		MaxCardinality:   co.maxCardinality,
		SampleRate:       co.sampleRate,
		MetricCounts:     make(map[string]int),
		TopMetrics:       make([]MetricCardinality, 0),
	}

	// Copy metric cardinality
	for metric, count := range co.metricCardinality {
		stats.MetricCounts[metric] = count
	}

	// Get top metrics by cardinality
	type metricCount struct {
		name  string
		count int
	}

	metrics := make([]metricCount, 0, len(co.metricCardinality))
	for name, count := range co.metricCardinality {
		metrics = append(metrics, metricCount{name, count})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].count > metrics[j].count
	})

	// Take top 10
	limit := 10
	if len(metrics) < limit {
		limit = len(metrics)
	}

	for i := 0; i < limit; i++ {
		stats.TopMetrics = append(stats.TopMetrics, MetricCardinality{
			MetricName:  metrics[i].name,
			Cardinality: metrics[i].count,
		})
	}

	return stats
}

// OptimizeCardinality performs cardinality optimization operations
func (co *CardinalityOptimizer) OptimizeCardinality() {
	co.mu.Lock()
	defer co.mu.Unlock()

	start := time.Now()
	defer func() {
		co.cleanupDuration.Observe(time.Since(start).Seconds())
	}()

	totalCardinality := co.getTotalCardinality()

	// If we're over the limit, increase sampling rate
	if totalCardinality > co.maxCardinality {
		oldRate := co.sampleRate

		// Calculate new sample rate to get within limits
		targetCardinality := float64(co.maxCardinality) * 0.9 // 90% of limit
		newRate := targetCardinality / float64(totalCardinality)

		// Don't reduce too aggressively
		minRate := oldRate * 0.5
		if newRate < minRate {
			newRate = minRate
		}

		co.sampleRate = newRate
		co.enableSampling = true

		co.logger.WithFields(logrus.Fields{
			"old_rate":           oldRate,
			"new_rate":           newRate,
			"total_cardinality":  totalCardinality,
			"target_cardinality": targetCardinality,
		}).Info("Adjusted sampling rate due to high cardinality")
	}

	// Periodic cleanup of old tracking data
	if time.Since(co.lastCleanup) > co.cleanupInterval {
		co.cleanupOldMetrics()
		co.lastCleanup = time.Now()
	}
}

// cleanupOldMetrics removes tracking data for metrics that haven't been seen recently
func (co *CardinalityOptimizer) cleanupOldMetrics() {
	// This is a simplified cleanup - in practice, you'd want to track
	// last seen times for metrics and remove old ones

	// Reset cardinality counts periodically to avoid unbounded growth
	resetThreshold := len(co.metricCardinality) > 1000
	if resetThreshold {
		co.metricCardinality = make(map[string]int)
		co.labelCardinality = make(map[string]map[string]int)
		co.logger.Info("Reset cardinality tracking due to size")
	}
}

// SetSampleRate updates the sampling rate
func (co *CardinalityOptimizer) SetSampleRate(rate float64) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.sampleRate = rate
	co.enableSampling = rate < 1.0

	co.logger.WithField("sample_rate", rate).Info("Updated sampling rate")
}

// SetMaxCardinality updates the maximum cardinality limit
func (co *CardinalityOptimizer) SetMaxCardinality(maxLimit int) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.maxCardinality = maxLimit

	co.logger.WithField("max_cardinality", maxLimit).Info("Updated maximum cardinality")
}

// Describe implements prometheus.Collector
func (co *CardinalityOptimizer) Describe(ch chan<- *prometheus.Desc) {
	co.cardinalityTotal.Describe(ch)
	co.cardinalityByMetric.Describe(ch)
	co.sampledMetrics.Describe(ch)
	co.droppedMetrics.Describe(ch)
	co.cleanupDuration.Describe(ch)
}

// Collect implements prometheus.Collector
func (co *CardinalityOptimizer) Collect(ch chan<- prometheus.Metric) {
	stats := co.GetCardinalityStats()

	co.cardinalityTotal.Set(float64(stats.TotalCardinality))

	// Update per-metric cardinality
	for metricName, count := range stats.MetricCounts {
		co.cardinalityByMetric.WithLabelValues(metricName).Set(float64(count))
	}

	co.cardinalityTotal.Collect(ch)
	co.cardinalityByMetric.Collect(ch)
	co.sampledMetrics.Collect(ch)
	co.droppedMetrics.Collect(ch)
	co.cleanupDuration.Collect(ch)
}

// CardinalityStats represents cardinality statistics
type CardinalityStats struct {
	TotalCardinality int
	MaxCardinality   int
	SampleRate       float64
	MetricCounts     map[string]int
	TopMetrics       []MetricCardinality
}

// MetricCardinality represents cardinality for a specific metric
type MetricCardinality struct {
	MetricName  string
	Cardinality int
}
