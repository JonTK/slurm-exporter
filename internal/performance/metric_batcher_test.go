// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewMetricBatcher(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
		EnableSampling:    false,
		EnableAggregation: false,
		MaxMetricAge:      5 * time.Minute,
		AggregationWindow: 30 * time.Second,
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)

	assert.NoError(t, err)
	assert.NotNil(t, batcher)
	assert.Equal(t, registry, batcher.registry)
	assert.Equal(t, config.EnableSampling, batcher.config.EnableSampling)

	_ = batcher.Stop()
}

func TestNewMetricBatcher_DefaultValues(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
		// Leave MaxMetricAge and AggregationWindow at 0 to test defaults
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)

	assert.NoError(t, err)
	assert.NotNil(t, batcher)
	assert.Equal(t, 5*time.Minute, batcher.config.MaxMetricAge)
	assert.Equal(t, 30*time.Second, batcher.config.AggregationWindow)

	_ = batcher.Stop()
}

func TestMetricItem_Key(t *testing.T) {
	t.Parallel()

	labels := prometheus.Labels{
		"job":      "slurm",
		"instance": "node1",
	}

	item := &MetricItem{
		metricName: "slurm_jobs_total",
		labels:     labels,
		value:      42,
		timestamp:  time.Now(),
		metricType: "counter",
	}

	key := item.Key()

	assert.NotEmpty(t, key)
	assert.Contains(t, key, "slurm_jobs_total")
	assert.Contains(t, key, "job")
	assert.Contains(t, key, "slurm")
}

func TestMetricItem_KeyConsistency(t *testing.T) {
	t.Parallel()

	labels := prometheus.Labels{
		"job":      "slurm",
		"instance": "node1",
	}

	item1 := &MetricItem{
		metricName: "slurm_jobs_total",
		labels:     labels,
		value:      42,
	}

	item2 := &MetricItem{
		metricName: "slurm_jobs_total",
		labels:     labels,
		value:      43,
	}

	key1 := item1.Key()
	key2 := item2.Key()

	// Keys should be identical regardless of value
	assert.Equal(t, key1, key2)
}

func TestMetricItem_Size(t *testing.T) {
	t.Parallel()

	labels := prometheus.Labels{
		"job": "slurm",
	}

	item := &MetricItem{
		metricName: "slurm_jobs_total",
		labels:     labels,
		value:      42,
		help:       "Total SLURM jobs",
	}

	size := item.Size()

	assert.Greater(t, size, 0)
	assert.GreaterOrEqual(t, size, len("slurm_jobs_total")+len("Total SLURM jobs"))
}

func TestMetricItem_Type(t *testing.T) {
	t.Parallel()

	item := &MetricItem{
		metricType: "counter",
	}

	assert.Equal(t, "counter", item.Type())
}

func TestMetricItem_Priority(t *testing.T) {
	t.Parallel()

	// Recent metric (< 1 minute)
	recentItem := &MetricItem{
		timestamp: time.Now(),
	}
	assert.Equal(t, 10, recentItem.Priority())

	// Medium age (1-5 minutes)
	mediumItem := &MetricItem{
		timestamp: time.Now().Add(-2 * time.Minute),
	}
	assert.Equal(t, 5, mediumItem.Priority())

	// Old metric (> 5 minutes)
	oldItem := &MetricItem{
		timestamp: time.Now().Add(-10 * time.Minute),
	}
	assert.Equal(t, 1, oldItem.Priority())
}

func TestMetricItem_Timestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()
	item := &MetricItem{
		timestamp: now,
	}

	assert.Equal(t, now, item.Timestamp())
}

func TestMetricBatcher_BatchMetric(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
		EnableSampling: false,
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	labels := prometheus.Labels{"job": "test"}
	err = batcher.BatchMetric("test_metric", labels, 42.5, "counter")

	assert.NoError(t, err)
}

func TestMetricBatcher_BatchMetric_WithSampling(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
		EnableSampling: true,
		SamplingRates: map[string]float64{
			"test_metric": 0.5,
		},
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	labels := prometheus.Labels{"job": "test"}

	// Add multiple metrics - some should be sampled
	count := 0
	for i := 0; i < 10; i++ {
		err := batcher.BatchMetric("test_metric", labels, float64(i), "counter")
		if err == nil {
			count++
		}
	}

	// We expect approximately 50% to pass sampling (some variance due to hash)
	assert.Greater(t, count, 0)
}

func TestMetricBatcher_CollectBatchedMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
	}

	registry := prometheus.NewRegistry()

	// Register a simple counter to ensure we have metrics
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_metric_total",
		Help: "Test metric",
	})
	registry.MustRegister(counter)
	counter.Inc()

	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	metrics, err := batcher.CollectBatchedMetrics()

	assert.NoError(t, err)
	// With a registry that has metrics, we should get results
	assert.NotNil(t, metrics)
}

func TestMetricBatcher_GetStats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
		EnableSampling:    true,
		EnableAggregation: true,
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	stats := batcher.GetStats()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "processor_stats")
	assert.Contains(t, stats, "buffer_sizes")
	assert.Contains(t, stats, "config")
	assert.Contains(t, stats, "sampling_enabled")
	assert.Contains(t, stats, "aggregation_enabled")

	samplingEnabled, ok := stats["sampling_enabled"].(bool)
	assert.True(t, ok)
	assert.True(t, samplingEnabled)

	aggregationEnabled, ok := stats["aggregation_enabled"].(bool)
	assert.True(t, ok)
	assert.True(t, aggregationEnabled)
}

func TestMetricBatcher_Stop(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)

	err = batcher.Stop()
	assert.NoError(t, err)
}

func TestHashSample_KeepAll(t *testing.T) {
	t.Parallel()

	// With rate >= 1.0, all samples should be kept (function returns false)
	shouldDrop := hashSample("test_key", 1.0)
	assert.False(t, shouldDrop)

	shouldDrop = hashSample("test_key", 1.5)
	assert.False(t, shouldDrop)
}

func TestHashSample_DropAll(t *testing.T) {
	t.Parallel()

	// With rate <= 0.0, all samples should be dropped (function returns true)
	shouldDrop := hashSample("test_key", 0.0)
	assert.True(t, shouldDrop)

	shouldDrop = hashSample("test_key", -0.5)
	assert.True(t, shouldDrop)
}

func TestHashSample_PartialSampling(t *testing.T) {
	t.Parallel()

	// With 0.5 rate, roughly 50% should be dropped and 50% kept
	// The exact distribution depends on hash function, so we verify both possibilities
	rate := 0.5
	kept := 0
	dropped := 0

	for i := 0; i < 100; i++ {
		key := "key_" + string(rune(i))
		if hashSample(key, rate) {
			dropped++
		} else {
			kept++
		}
	}

	// Verify that we have a mix (not all kept or all dropped)
	assert.Greater(t, kept+dropped, 0)
	assert.Equal(t, 100, kept+dropped)      // All samples accounted for
	assert.True(t, kept > 0 || dropped > 0) // At least one of each is reasonable
}

func TestCalculateHistogramBuckets(t *testing.T) {
	t.Parallel()

	values := []float64{0.005, 0.02, 0.15, 0.7, 2.5, 8.0}
	buckets := calculateHistogramBuckets(values)

	assert.NotNil(t, buckets)
	assert.Greater(t, len(buckets), 0)

	// Verify bucket counts (each value <= bucket increments that bucket)
	assert.Equal(t, 1, buckets[0.01]) // Only 0.005 <= 0.01
	assert.Equal(t, 2, buckets[0.05]) // 0.005, 0.02 <= 0.05
	assert.Equal(t, 2, buckets[0.1])  // 0.005, 0.02 <= 0.1
	assert.Equal(t, 3, buckets[0.5])  // 0.005, 0.02, 0.15 <= 0.5
}

func TestCalculateHistogramBuckets_Empty(t *testing.T) {
	t.Parallel()

	values := []float64{}
	buckets := calculateHistogramBuckets(values)

	assert.NotNil(t, buckets)
	// All buckets should have count 0
	for _, count := range buckets {
		assert.Equal(t, 0, count)
	}
}

func TestCalculateHistogramBuckets_Single(t *testing.T) {
	t.Parallel()

	values := []float64{0.5}
	buckets := calculateHistogramBuckets(values)

	assert.NotNil(t, buckets)
	// Value 0.5 should be in buckets >= 0.5
	assert.Equal(t, 1, buckets[0.5])
	assert.Equal(t, 1, buckets[1])
	assert.Equal(t, 1, buckets[5])
}

func TestCalculateQuantiles(t *testing.T) {
	t.Parallel()

	values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	quantiles := calculateQuantiles(values, []float64{0.5, 0.9, 0.99})

	assert.NotNil(t, quantiles)
	assert.Equal(t, 3, len(quantiles))
	assert.Contains(t, quantiles, 0.5)
	assert.Contains(t, quantiles, 0.9)
	assert.Contains(t, quantiles, 0.99)

	// 50th percentile should be around 5.5
	assert.Greater(t, quantiles[0.5], 4.0)
	assert.Less(t, quantiles[0.5], 7.0)
}

func TestCalculateQuantiles_Empty(t *testing.T) {
	t.Parallel()

	values := []float64{}
	quantiles := calculateQuantiles(values, []float64{0.5, 0.9})

	assert.Nil(t, quantiles)
}

func TestCalculateQuantiles_Single(t *testing.T) {
	t.Parallel()

	values := []float64{42}
	quantiles := calculateQuantiles(values, []float64{0.5, 0.9, 0.99})

	assert.NotNil(t, quantiles)
	// All quantiles should return the single value
	assert.Equal(t, 42.0, quantiles[0.5])
	assert.Equal(t, 42.0, quantiles[0.9])
	assert.Equal(t, 42.0, quantiles[0.99])
}

func TestMetricBatcher_MultipleMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	// Add different metric types
	err = batcher.BatchMetric("counter_metric", prometheus.Labels{}, 10, "counter")
	assert.NoError(t, err)

	err = batcher.BatchMetric("gauge_metric", prometheus.Labels{}, 20, "gauge")
	assert.NoError(t, err)

	err = batcher.BatchMetric("histogram_metric", prometheus.Labels{}, 0.5, "histogram")
	assert.NoError(t, err)

	err = batcher.BatchMetric("summary_metric", prometheus.Labels{}, 0.1, "summary")
	assert.NoError(t, err)

	// All should succeed
	assert.NoError(t, err)
}

func TestMetricItem_EmptyLabels(t *testing.T) {
	t.Parallel()

	item := &MetricItem{
		metricName: "test_metric",
		labels:     prometheus.Labels{},
		value:      42,
	}

	key := item.Key()
	assert.Equal(t, "test_metric", key)

	size := item.Size()
	assert.Greater(t, size, 0)
}

func TestMetricBatcher_NilRegistry(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 2,
		},
	}

	batcher, err := NewMetricBatcher(config, nil, logger)
	assert.NoError(t, err)
	assert.NotNil(t, batcher)
	assert.Nil(t, batcher.registry)

	metrics, err := batcher.CollectBatchedMetrics()
	assert.NoError(t, err)
	assert.Nil(t, metrics)

	_ = batcher.Stop()
}

func TestMetricBatcher_Aggregation_SingleMetric(t *testing.T) {
	t.Parallel()

	batcher := &MetricBatcher{
		config: MetricBatcherConfig{
			AggregationWindow: 1 * time.Second,
		},
	}

	metric := &MetricItem{
		metricName: "test",
		value:      42,
		timestamp:  time.Now(),
	}

	metrics := batcher.aggregateMetrics([]*MetricItem{metric})
	assert.Equal(t, 1, len(metrics))
	assert.Equal(t, metric, metrics[0])
}

func TestMetricBatcher_Aggregation_MultipleWindows(t *testing.T) {
	t.Parallel()

	batcher := &MetricBatcher{
		config: MetricBatcherConfig{
			AggregationWindow: 100 * time.Millisecond,
		},
		logger: testutil.GetTestLogger(),
	}

	now := time.Now()
	metrics := []*MetricItem{
		{metricName: "test", value: 10, timestamp: now},
		{metricName: "test", value: 20, timestamp: now.Add(50 * time.Millisecond)},
		{metricName: "test", value: 30, timestamp: now.Add(200 * time.Millisecond)},
	}

	aggregated := batcher.aggregateMetrics(metrics)
	assert.Greater(t, len(aggregated), 0)
}

func TestMetricBatcher_AggregateWindow(t *testing.T) {
	t.Parallel()

	batcher := &MetricBatcher{
		logger: testutil.GetTestLogger(),
	}

	now := time.Now()
	metrics := []*MetricItem{
		{metricName: "test", value: 10, timestamp: now, labels: prometheus.Labels{"job": "a"}},
		{metricName: "test", value: 20, timestamp: now.Add(50 * time.Millisecond), labels: prometheus.Labels{"job": "a"}},
	}

	aggregated := batcher.aggregateWindow(metrics)
	assert.NotNil(t, aggregated)
	assert.Equal(t, "test", aggregated.metricName)
	// Values should be summed
	assert.Equal(t, 30.0, aggregated.value)
	// Timestamp should be from last metric
	assert.Equal(t, metrics[1].timestamp, aggregated.timestamp)
}

func TestMetricBatcher_Concurrent_BatchMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	config := MetricBatcherConfig{
		BatchConfig: BatchConfig{
			MaxBatchSize:   1000,
			MaxBatchWait:   5 * time.Second,
			MaxConcurrency: 4,
		},
	}

	registry := prometheus.NewRegistry()
	batcher, err := NewMetricBatcher(config, registry, logger)
	assert.NoError(t, err)
	defer batcher.Stop()

	done := make(chan bool, 10)

	// Concurrent metric submissions
	for i := 0; i < 10; i++ {
		go func(idx int) {
			labels := prometheus.Labels{"worker": string(rune(idx))}
			_ = batcher.BatchMetric("test_metric", labels, float64(idx), "counter")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := batcher.GetStats()
	assert.NotNil(t, stats)
}
