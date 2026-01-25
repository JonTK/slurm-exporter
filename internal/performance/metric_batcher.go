// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
)

// MetricBatcher provides efficient batching for Prometheus metrics
type MetricBatcher struct {
	processor    *BatchProcessor
	logger       *logrus.Entry
	mu           sync.RWMutex
	metricBuffer map[string][]*MetricItem
	registry     prometheus.Gatherer
	config       MetricBatcherConfig
}

// MetricBatcherConfig holds configuration for metric batching
type MetricBatcherConfig struct {
	BatchConfig       BatchConfig        `yaml:"batch_config"`
	EnableSampling    bool               `yaml:"enable_sampling"`
	SamplingRates     map[string]float64 `yaml:"sampling_rates"`
	EnableAggregation bool               `yaml:"enable_aggregation"`
	AggregationWindow time.Duration      `yaml:"aggregation_window"`
	MaxMetricAge      time.Duration      `yaml:"max_metric_age"`
}

// MetricItem implements BatchItem for Prometheus metrics
type MetricItem struct {
	metricName string
	labels     prometheus.Labels
	value      float64
	timestamp  time.Time
	metricType string
	help       string
}

func (m *MetricItem) Key() string {
	// Create unique key from metric name and sorted labels
	key := m.metricName

	// Sort label keys for consistent ordering
	keys := make([]string, 0, len(m.labels))
	for k := range m.labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Append label values
	for _, k := range keys {
		key += fmt.Sprintf("_%s_%s", k, m.labels[k])
	}

	return key
}

func (m *MetricItem) Size() int {
	// Estimate size based on metric name and labels
	size := len(m.metricName) + len(m.help) + 8 // 8 bytes for float64 value
	for k, v := range m.labels {
		size += len(k) + len(v)
	}
	return size
}

func (m *MetricItem) Type() string {
	return m.metricType
}

func (m *MetricItem) Priority() int {
	// Higher priority for more recent metrics
	age := time.Since(m.timestamp)
	if age < time.Minute {
		return 10
	} else if age < 5*time.Minute {
		return 5
	}
	return 1
}

func (m *MetricItem) Timestamp() time.Time {
	return m.timestamp
}

// NewMetricBatcher creates a new metric batcher
func NewMetricBatcher(config MetricBatcherConfig, registry prometheus.Gatherer, logger *logrus.Entry) (*MetricBatcher, error) {
	processor, err := NewBatchProcessor(config.BatchConfig, logger.WithField("component", "batch_processor"))
	if err != nil {
		return nil, fmt.Errorf("creating batch processor: %w", err)
	}

	if config.MaxMetricAge <= 0 {
		config.MaxMetricAge = 5 * time.Minute
	}
	if config.AggregationWindow <= 0 {
		config.AggregationWindow = 30 * time.Second
	}

	mb := &MetricBatcher{
		processor:    processor,
		logger:       logger,
		metricBuffer: make(map[string][]*MetricItem),
		registry:     registry,
		config:       config,
	}

	// Register metric processors
	mb.registerProcessors()

	return mb, nil
}

// registerProcessors registers batch processors for different metric types
func (mb *MetricBatcher) registerProcessors() {
	// Counter processor
	mb.processor.ProcessBatch("counter", mb.processCounterBatch)

	// Gauge processor
	mb.processor.ProcessBatch("gauge", mb.processGaugeBatch)

	// Histogram processor
	mb.processor.ProcessBatch("histogram", mb.processHistogramBatch)

	// Summary processor
	mb.processor.ProcessBatch("summary", mb.processSummaryBatch)
}

// BatchMetric adds a metric to the batch
func (mb *MetricBatcher) BatchMetric(name string, labels prometheus.Labels, value float64, metricType string) error {
	item := &MetricItem{
		metricName: name,
		labels:     labels,
		value:      value,
		timestamp:  time.Now(),
		metricType: metricType,
	}

	// Apply sampling if enabled
	if mb.config.EnableSampling {
		rate, exists := mb.config.SamplingRates[name]
		if !exists {
			rate = 1.0 // Default to no sampling
		}

		// Simple sampling based on hash of key
		if hashSample(item.Key(), rate) {
			return nil // Skip this sample
		}
	}

	return mb.processor.Add(item)
}

// processCounterBatch processes a batch of counter metrics
func (mb *MetricBatcher) processCounterBatch(ctx context.Context, items []BatchItem) error {
	mb.logger.WithField("items", len(items)).Debug("Processing counter batch")

	// Group by metric name
	grouped := make(map[string][]*MetricItem)
	for _, item := range items {
		metric, ok := item.(*MetricItem)
		if !ok {
			mb.logger.Warn("Item is not a MetricItem, skipping")
			continue
		}
		grouped[metric.metricName] = append(grouped[metric.metricName], metric)
	}

	// Process each metric group
	for metricName, metrics := range grouped {
		if mb.config.EnableAggregation {
			metrics = mb.aggregateMetrics(metrics)
		}

		// Create or update counter
		for _, metric := range metrics {
			// In a real implementation, this would update the actual Prometheus metrics
			mb.logger.WithFields(logrus.Fields{
				"metric": metricName,
				"labels": metric.labels,
				"value":  metric.value,
			}).Debug("Counter metric processed")
		}
	}

	return nil
}

// processGaugeBatch processes a batch of gauge metrics
func (mb *MetricBatcher) processGaugeBatch(ctx context.Context, items []BatchItem) error {
	mb.logger.WithField("items", len(items)).Debug("Processing gauge batch")

	// For gauges, we typically want the most recent value
	latest := make(map[string]*MetricItem)

	for _, item := range items {
		metric, ok := item.(*MetricItem)
		if !ok {
			mb.logger.Warn("Item is not a MetricItem, skipping")
			continue
		}
		key := metric.Key()

		existing, exists := latest[key]
		if !exists || metric.timestamp.After(existing.timestamp) {
			latest[key] = metric
		}
	}

	// Process latest values
	for _, metric := range latest {
		mb.logger.WithFields(logrus.Fields{
			"metric": metric.metricName,
			"labels": metric.labels,
			"value":  metric.value,
		}).Debug("Gauge metric processed")
	}

	return nil
}

// processHistogramBatch processes a batch of histogram metrics
func (mb *MetricBatcher) processHistogramBatch(ctx context.Context, items []BatchItem) error {
	mb.logger.WithField("items", len(items)).Debug("Processing histogram batch")

	// Group observations by metric and labels
	observations := make(map[string][]float64)

	for _, item := range items {
		metric, ok := item.(*MetricItem)
		if !ok {
			mb.logger.Warn("Item is not a MetricItem, skipping")
			continue
		}
		key := metric.Key()
		observations[key] = append(observations[key], metric.value)
	}

	// Process histogram observations
	for key, values := range observations {
		// Calculate histogram buckets
		buckets := calculateHistogramBuckets(values)

		mb.logger.WithFields(logrus.Fields{
			"key":          key,
			"observations": len(values),
			"buckets":      len(buckets),
		}).Debug("Histogram metric processed")
	}

	return nil
}

// processSummaryBatch processes a batch of summary metrics
func (mb *MetricBatcher) processSummaryBatch(ctx context.Context, items []BatchItem) error {
	mb.logger.WithField("items", len(items)).Debug("Processing summary batch")

	// Similar to histogram but calculate quantiles
	observations := make(map[string][]float64)

	for _, item := range items {
		metric, ok := item.(*MetricItem)
		if !ok {
			mb.logger.Warn("Item is not a MetricItem, skipping")
			continue
		}
		key := metric.Key()
		observations[key] = append(observations[key], metric.value)
	}

	// Process summary observations
	for key, values := range observations {
		// Calculate quantiles
		quantiles := calculateQuantiles(values, []float64{0.5, 0.9, 0.99})

		mb.logger.WithFields(logrus.Fields{
			"key":          key,
			"observations": len(values),
			"quantiles":    quantiles,
		}).Debug("Summary metric processed")
	}

	return nil
}

// aggregateMetrics aggregates metrics within the time window
func (mb *MetricBatcher) aggregateMetrics(metrics []*MetricItem) []*MetricItem {
	if len(metrics) <= 1 {
		return metrics
	}

	// Sort by timestamp
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].timestamp.Before(metrics[j].timestamp)
	})

	aggregated := make([]*MetricItem, 0)
	windowStart := metrics[0].timestamp
	windowMetrics := []*MetricItem{metrics[0]}

	for i := 1; i < len(metrics); i++ {
		if metrics[i].timestamp.Sub(windowStart) <= mb.config.AggregationWindow {
			windowMetrics = append(windowMetrics, metrics[i])
		} else {
			// Aggregate window
			agg := mb.aggregateWindow(windowMetrics)
			aggregated = append(aggregated, agg)

			// Start new window
			windowStart = metrics[i].timestamp
			windowMetrics = []*MetricItem{metrics[i]}
		}
	}

	// Don't forget last window
	if len(windowMetrics) > 0 {
		agg := mb.aggregateWindow(windowMetrics)
		aggregated = append(aggregated, agg)
	}

	return aggregated
}

// aggregateWindow aggregates metrics within a single window
func (mb *MetricBatcher) aggregateWindow(metrics []*MetricItem) *MetricItem {
	if len(metrics) == 1 {
		return metrics[0]
	}

	// For counters, sum the values
	sum := 0.0
	for _, m := range metrics {
		sum += m.value
	}

	// Return aggregated metric with latest timestamp
	return &MetricItem{
		metricName: metrics[0].metricName,
		labels:     metrics[0].labels,
		value:      sum,
		timestamp:  metrics[len(metrics)-1].timestamp,
		metricType: metrics[0].metricType,
		help:       metrics[0].help,
	}
}

// CollectBatchedMetrics collects all batched metrics
func (mb *MetricBatcher) CollectBatchedMetrics() ([]*dto.MetricFamily, error) {
	// Force flush all pending batches
	mb.processor.FlushAll()

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Collect metrics from registry
	if mb.registry != nil {
		metrics, err := mb.registry.Gather()
		if err != nil {
			return nil, fmt.Errorf("failed to gather batched metrics: %w", err)
		}
		return metrics, nil
	}

	return nil, nil
}

// GetStats returns batcher statistics
func (mb *MetricBatcher) GetStats() map[string]interface{} {
	processorStats := mb.processor.GetStats()

	mb.mu.RLock()
	bufferSizes := make(map[string]int)
	for metricType, items := range mb.metricBuffer {
		bufferSizes[metricType] = len(items)
	}
	mb.mu.RUnlock()

	return map[string]interface{}{
		"processor_stats":     processorStats,
		"buffer_sizes":        bufferSizes,
		"config":              mb.config,
		"sampling_enabled":    mb.config.EnableSampling,
		"aggregation_enabled": mb.config.EnableAggregation,
	}
}

// Stop stops the metric batcher
func (mb *MetricBatcher) Stop() error {
	return mb.processor.Stop()
}

// Helper functions

// hashSample determines if a sample should be kept based on consistent hashing
func hashSample(key string, rate float64) bool {
	if rate >= 1.0 {
		return false // Keep all samples
	}
	if rate <= 0.0 {
		return true // Drop all samples
	}

	// Simple hash-based sampling
	hash := uint32(0)
	for _, b := range []byte(key) {
		hash = hash*31 + uint32(b)
	}

	threshold := uint32(rate * float64(^uint32(0)))
	return hash > threshold
}

// calculateHistogramBuckets calculates histogram buckets from observations
func calculateHistogramBuckets(values []float64) map[float64]int {
	buckets := map[float64]int{
		0.01: 0, 0.05: 0, 0.1: 0, 0.5: 0, 1: 0, 5: 0, 10: 0,
	}

	for _, v := range values {
		for bucket := range buckets {
			if v <= bucket {
				buckets[bucket]++
			}
		}
	}

	return buckets
}

// calculateQuantiles calculates quantiles from observations
func calculateQuantiles(values []float64, quantiles []float64) map[float64]float64 {
	if len(values) == 0 {
		return nil
	}

	sort.Float64s(values)
	result := make(map[float64]float64)

	for _, q := range quantiles {
		idx := int(q * float64(len(values)-1))
		result[q] = values[idx]
	}

	return result
}
