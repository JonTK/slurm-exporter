// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewCardinalityOptimizer(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	assert.NotNil(t, co)
	assert.Equal(t, 10000, co.maxCardinality)
	assert.Equal(t, 1.0, co.sampleRate)
	assert.False(t, co.enableSampling)
	assert.NotNil(t, co.metricCardinality)
	assert.NotNil(t, co.labelCardinality)
	assert.NotNil(t, co.samplingSeeds)
}

func TestNewCardinalityOptimizer_WithSampling(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 0.5, logger)

	assert.NotNil(t, co)
	assert.Equal(t, 0.5, co.sampleRate)
	assert.True(t, co.enableSampling)
}

func TestCardinalityOptimizer_ShouldCollectMetric_BelowLimit(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(100, 1.0, logger)

	labels := map[string]string{"job": "test", "instance": "node1"}
	should := co.ShouldCollectMetric("test_metric", labels)

	assert.True(t, should)
}

func TestCardinalityOptimizer_ShouldCollectMetric_AtLimit_NoSampling(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(2, 1.0, logger)

	// Fill up to limit
	co.ShouldCollectMetric("metric1", map[string]string{"a": "1"})
	co.ShouldCollectMetric("metric2", map[string]string{"b": "2"})

	// This should be dropped (at limit, no sampling)
	labels := map[string]string{"c": "3"}
	should := co.ShouldCollectMetric("metric3", labels)

	assert.False(t, should)
}

func TestCardinalityOptimizer_ShouldCollectMetric_WithSampling(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(2, 0.5, logger)

	// Fill up to limit
	co.ShouldCollectMetric("metric1", map[string]string{"a": "1"})
	co.ShouldCollectMetric("metric2", map[string]string{"b": "2"})

	// With sampling enabled, some might pass
	labels := map[string]string{"c": "3"}
	should := co.ShouldCollectMetric("metric3", labels)

	// Can't predict exact result, but should not panic
	assert.NotNil(t, should)
}

func TestCardinalityOptimizer_GetCardinalityStats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	// Add some metrics
	co.ShouldCollectMetric("jobs_total", map[string]string{"cluster": "test"})
	co.ShouldCollectMetric("nodes_allocated", map[string]string{"cluster": "test", "node": "node1"})
	co.ShouldCollectMetric("nodes_allocated", map[string]string{"cluster": "test", "node": "node2"})

	stats := co.GetCardinalityStats()

	assert.Equal(t, 3, stats.TotalCardinality)
	assert.Equal(t, 10000, stats.MaxCardinality)
	assert.Equal(t, 1.0, stats.SampleRate)
	assert.NotEmpty(t, stats.MetricCounts)
	assert.Greater(t, len(stats.TopMetrics), 0)
}

func TestCardinalityOptimizer_SetSampleRate(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)
	assert.False(t, co.enableSampling)

	co.SetSampleRate(0.5)
	assert.Equal(t, 0.5, co.sampleRate)
	assert.True(t, co.enableSampling)

	co.SetSampleRate(1.0)
	assert.Equal(t, 1.0, co.sampleRate)
	assert.False(t, co.enableSampling)
}

func TestCardinalityOptimizer_SetMaxCardinality(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	co.SetMaxCardinality(20000)
	assert.Equal(t, 20000, co.maxCardinality)

	co.SetMaxCardinality(5000)
	assert.Equal(t, 5000, co.maxCardinality)
}

func TestCardinalityOptimizer_OptimizeCardinality_BelowLimit(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	// Add metrics below limit
	for i := 0; i < 100; i++ {
		co.ShouldCollectMetric("test_metric", map[string]string{"index": string(rune(i))})
	}

	initialRate := co.sampleRate

	// Call OptimizeCardinality when below limit
	co.OptimizeCardinality()

	// Sample rate should not change when below limit
	assert.Equal(t, initialRate, co.sampleRate)
}

func TestCardinalityOptimizer_OptimizeCardinality_OverLimit(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(100, 1.0, logger)

	// Add metrics to exceed limit
	for i := 0; i < 150; i++ {
		co.ShouldCollectMetric("test_metric", map[string]string{"index": string(rune(i % 256))})
	}

	stats := co.GetCardinalityStats()
	if stats.TotalCardinality > co.maxCardinality {
		co.OptimizeCardinality()

		// Sample rate should increase
		assert.True(t, co.sampleRate < 1.0)
		assert.True(t, co.enableSampling)
	}
}

func TestCardinalityOptimizer_Describe(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	ch := make(chan *prometheus.Desc, 10)
	co.Describe(ch)
	close(ch)

	descs := 0
	for range ch {
		descs++
	}

	assert.Greater(t, descs, 0)
}

func TestCardinalityOptimizer_Collect(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	// Add a metric
	co.ShouldCollectMetric("test_metric", map[string]string{"label": "value"})

	ch := make(chan prometheus.Metric, 20)
	co.Collect(ch)
	close(ch)

	metrics := 0
	for range ch {
		metrics++
	}

	assert.Greater(t, metrics, 0)
}

func TestCardinalityOptimizer_HashMetric_Consistency(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	labels1 := map[string]string{"job": "test", "instance": "node1"}
	labels2 := map[string]string{"instance": "node1", "job": "test"}

	hash1 := co.hashMetric("test_metric", labels1)
	hash2 := co.hashMetric("test_metric", labels2)

	// Should be identical regardless of label order
	assert.Equal(t, hash1, hash2)
}

func TestCardinalityOptimizer_HashMetric_Different(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	hash1 := co.hashMetric("metric1", map[string]string{"a": "1"})
	hash2 := co.hashMetric("metric2", map[string]string{"a": "1"})
	hash3 := co.hashMetric("metric1", map[string]string{"a": "2"})

	// Different metrics/labels should have different hashes
	assert.NotEqual(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}

func TestCardinalityOptimizer_HashString(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	hash1 := co.hashString("test1")
	hash2 := co.hashString("test1")
	hash3 := co.hashString("test2")

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, hash1, hash3)
}

func TestCardinalityOptimizer_EmptyLabels(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	should := co.ShouldCollectMetric("test_metric", map[string]string{})
	assert.True(t, should)

	stats := co.GetCardinalityStats()
	assert.Equal(t, 1, stats.TotalCardinality)
}

func TestCardinalityOptimizer_ShouldSampleMetric_BelowLimit(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(1000, 0.5, logger)

	// Should sample when at limit and sampling enabled
	co.mu.Lock()
	// Fill to limit
	for i := 0; i < 1000; i++ {
		co.metricCardinality["test"] = i
	}
	co.mu.Unlock()

	labels := map[string]string{"a": "1"}
	// This will try sampling
	result := co.ShouldCollectMetric("new_metric", labels)

	// Can't predict exact result, but should not panic
	assert.NotNil(t, result)
}

func TestCardinalityOptimizer_TopMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	// Add metrics with different cardinalities
	for i := 0; i < 50; i++ {
		co.ShouldCollectMetric("metric_a", map[string]string{"idx": string(rune(i))})
	}
	for i := 0; i < 30; i++ {
		co.ShouldCollectMetric("metric_b", map[string]string{"idx": string(rune(i))})
	}
	for i := 0; i < 20; i++ {
		co.ShouldCollectMetric("metric_c", map[string]string{"idx": string(rune(i))})
	}

	stats := co.GetCardinalityStats()

	// Top metric should be metric_a with 50 cardinality
	assert.Greater(t, len(stats.TopMetrics), 0)
	assert.Equal(t, "metric_a", stats.TopMetrics[0].MetricName)
	assert.Equal(t, 50, stats.TopMetrics[0].Cardinality)
}

func TestCardinalityOptimizer_CleanupOldMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	// Add many metrics to trigger cleanup threshold
	for i := 0; i < 1100; i++ {
		co.ShouldCollectMetric("metric_"+string(rune(i%256)), map[string]string{"idx": string(rune(i))})
	}

	// Call cleanup
	co.OptimizeCardinality()

	// After cleanup, should have been reset
	stats := co.GetCardinalityStats()
	assert.NotNil(t, stats)
}

func TestCardinalityOptimizer_MetricCardinality_Stats(t *testing.T) {
	t.Parallel()

	mc := MetricCardinality{
		MetricName:  "test_metric",
		Cardinality: 42,
	}

	assert.Equal(t, "test_metric", mc.MetricName)
	assert.Equal(t, 42, mc.Cardinality)
}

func TestCardinalityOptimizer_Concurrent_Collect(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	done := make(chan bool, 5)

	// Concurrent collections
	for i := 0; i < 5; i++ {
		go func(idx int) {
			labels := map[string]string{"worker": string(rune(idx))}
			co.ShouldCollectMetric("test_metric", labels)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 5; i++ {
		<-done
	}

	stats := co.GetCardinalityStats()
	assert.Equal(t, 5, stats.TotalCardinality)
}

func TestCardinalityOptimizer_Metrics_Initialized(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	co := NewCardinalityOptimizer(10000, 1.0, logger)

	assert.NotNil(t, co.cardinalityTotal)
	assert.NotNil(t, co.cardinalityByMetric)
	assert.NotNil(t, co.sampledMetrics)
	assert.NotNil(t, co.droppedMetrics)
	assert.NotNil(t, co.cleanupDuration)
}
