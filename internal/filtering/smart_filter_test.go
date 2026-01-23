// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package filtering

import (
	"context"
	"testing"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSmartFilter_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled: false,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, filter)
	assert.False(t, filter.IsEnabled())
}

func TestNewSmartFilter_Enabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	assert.NotNil(t, filter)
	assert.True(t, filter.IsEnabled())

	// Clean up
	_ = filter.Close()
}

func TestNewSmartFilter_InvalidConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	testCases := []struct {
		name string
		cfg  config.SmartFilteringConfig
	}{
		{
			name: "negative_noise_threshold",
			cfg: config.SmartFilteringConfig{
				Enabled:        true,
				NoiseThreshold: -0.1,
				CacheSize:      1000,
				LearningWindow: 50,
			},
		},
		{
			name: "zero_cache_size",
			cfg: config.SmartFilteringConfig{
				Enabled:        true,
				NoiseThreshold: 0.8,
				CacheSize:      0,
				LearningWindow: 50,
			},
		},
		{
			name: "zero_learning_window",
			cfg: config.SmartFilteringConfig{
				Enabled:        true,
				NoiseThreshold: 0.8,
				CacheSize:      1000,
				LearningWindow: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter, err := NewSmartFilter(tc.cfg, logger)
			assert.Error(t, err)
			assert.Nil(t, filter)
		})
	}
}

func TestSmartFilter_ProcessMetrics_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled: false,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)

	// Create test metrics
	metrics := createTestMetrics()

	// Process metrics
	filtered, err := filter.ProcessMetrics(context.Background(), "test_collector", metrics)
	require.NoError(t, err)

	// Should return original metrics unchanged
	assert.Equal(t, metrics, filtered)
}

func TestSmartFilter_ProcessMetrics_LearningPhase(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	// Create test metrics
	metrics := createTestMetrics()

	// Process metrics during learning phase
	filtered, err := filter.ProcessMetrics(context.Background(), "test_collector", metrics)
	require.NoError(t, err)

	// During learning phase, all metrics should be kept
	assert.Equal(t, len(metrics), len(filtered))
}

func TestSmartFilter_CreateMetricKey(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	testCases := []struct {
		name     string
		metric   *dto.Metric
		expected string
	}{
		{
			name: "no_labels",
			metric: &dto.Metric{
				Gauge: &dto.Gauge{Value: func(f float64) *float64 { return &f }(1.0)},
			},
			expected: "test_metric",
		},
		{
			name: "with_labels",
			metric: &dto.Metric{
				Label: []*dto.LabelPair{
					{Name: func(s string) *string { return &s }("job"), Value: func(s string) *string { return &s }("123")},
					{Name: func(s string) *string { return &s }("state"), Value: func(s string) *string { return &s }("running")},
				},
				Gauge: &dto.Gauge{Value: func(f float64) *float64 { return &f }(1.0)},
			},
			expected: "test_metric_job=123_state=running",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := filter.createMetricKey("test_metric", tc.metric)
			assert.Equal(t, tc.expected, key)
		})
	}
}

func TestSmartFilter_ExtractValue(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	testCases := []struct {
		name     string
		metric   *dto.Metric
		expected float64
	}{
		{
			name: "gauge",
			metric: &dto.Metric{
				Gauge: &dto.Gauge{Value: func(f float64) *float64 { return &f }(42.5)},
			},
			expected: 42.5,
		},
		{
			name: "counter",
			metric: &dto.Metric{
				Counter: &dto.Counter{Value: func(f float64) *float64 { return &f }(100.0)},
			},
			expected: 100.0,
		},
		{
			name: "untyped",
			metric: &dto.Metric{
				Untyped: &dto.Untyped{Value: func(f float64) *float64 { return &f }(25.0)},
			},
			expected: 25.0,
		},
		{
			name: "histogram",
			metric: &dto.Metric{
				Histogram: &dto.Histogram{SampleCount: func(u uint64) *uint64 { return &u }(50)},
			},
			expected: 50.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := filter.extractValue(tc.metric)
			assert.Equal(t, tc.expected, value)
		})
	}
}

func TestSmartFilter_CalculateStatistics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	// Create a pattern with known values
	pattern := &MetricPattern{
		Values: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	}

	filter.calculateStatistics(pattern)

	// Check calculated statistics
	assert.Equal(t, 3.0, pattern.Mean)
	assert.InDelta(t, 2.5, pattern.Variance, 0.01) // Sample variance
	assert.True(t, pattern.ChangeRate > 0)
	assert.True(t, pattern.NoiseScore >= 0 && pattern.NoiseScore <= 1)
	assert.True(t, pattern.Correlation >= -1 && pattern.Correlation <= 1)
}

func TestSmartFilter_CalculateCorrelation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	testCases := []struct {
		name      string
		x         []float64
		y         []float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "perfect_positive_correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{2, 4, 6, 8, 10},
			expected:  1.0,
			tolerance: 0.01,
		},
		{
			name:      "perfect_negative_correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{10, 8, 6, 4, 2},
			expected:  -1.0,
			tolerance: 0.01,
		},
		{
			name:      "no_correlation",
			x:         []float64{1, 2, 3, 4, 5},
			y:         []float64{1, 1, 1, 1, 1},
			expected:  0.0,
			tolerance: 0.01,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			correlation := filter.calculateCorrelation(tc.x, tc.y)
			assert.InDelta(t, tc.expected, correlation, tc.tolerance)
		})
	}
}

func TestSmartFilter_Cache(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      3, // Small cache for testing
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	// Test cache operations
	filter.cacheDecision("key1", ActionKeep)
	filter.cacheDecision("key2", ActionFilter)
	filter.cacheDecision("key3", ActionReduce)

	// Should find cached decisions
	action, found := filter.getCachedDecision("key1")
	assert.True(t, found)
	assert.Equal(t, ActionKeep, action)

	action, found = filter.getCachedDecision("key2")
	assert.True(t, found)
	assert.Equal(t, ActionFilter, action)

	// Test cache eviction by adding more items
	filter.cacheDecision("key4", ActionKeep)
	filter.cacheDecision("key5", ActionKeep)

	// Cache should have been partially cleared
	assert.True(t, len(filter.filterCache) <= cfg.CacheSize)
}

func TestSmartFilter_RegisterMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	registry := prometheus.NewRegistry()
	err = filter.RegisterMetrics(registry)
	assert.NoError(t, err)

	// Process some metrics to populate the metrics
	metrics := createTestMetrics()
	_, err = filter.ProcessMetrics(context.Background(), "test_collector", metrics)
	require.NoError(t, err)

	// Test that metrics are registered by gathering them
	metricFamilies, err := registry.Gather()
	assert.NoError(t, err)
	assert.True(t, len(metricFamilies) >= 5)
}

func TestSmartFilter_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	stats := filter.GetStats()
	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, true, stats["learning_phase"])
	assert.Equal(t, 0, stats["total_patterns"])
	assert.Equal(t, 0, stats["cache_size"])

	// Process some metrics
	metrics := createTestMetrics()
	_, err = filter.ProcessMetrics(context.Background(), "test_collector", metrics)
	require.NoError(t, err)

	stats = filter.GetStats()
	assert.True(t, stats["total_patterns"].(int) > 0)
}

func TestSmartFilter_GetPatterns(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	// Initially no patterns
	patterns := filter.GetPatterns()
	assert.Empty(t, patterns)

	// Process some metrics to create patterns
	metrics := createTestMetrics()
	_, err = filter.ProcessMetrics(context.Background(), "test_collector", metrics)
	require.NoError(t, err)

	patterns = filter.GetPatterns()
	assert.NotEmpty(t, patterns)
}

func TestSmartFilter_FilterAction_String(t *testing.T) {
	testCases := []struct {
		action   FilterAction
		expected string
	}{
		{ActionKeep, "keep"},
		{ActionFilter, "filter"},
		{ActionReduce, "reduce"},
		{ActionUnknown, "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.action.String())
		})
	}
}

func TestSmartFilter_NoiseScoreCalculation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.8,
		CacheSize:      1000,
		LearningWindow: 50,
		VarianceLimit:  10.0,
		CorrelationMin: 0.1,
	}

	filter, err := NewSmartFilter(cfg, logger)
	require.NoError(t, err)
	defer func() { _ = filter.Close() }()

	testCases := []struct {
		name        string
		variance    float64
		changeRate  float64
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "low_noise",
			variance:    1.0,
			changeRate:  0.1,
			expectedMin: 0.0,
			expectedMax: 0.5,
		},
		{
			name:        "high_noise",
			variance:    20.0,
			changeRate:  2.0,
			expectedMin: 0.5,
			expectedMax: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern := &MetricPattern{
				Variance:   tc.variance,
				ChangeRate: tc.changeRate,
			}

			noiseScore := filter.calculateNoiseScore(pattern)
			assert.True(t, noiseScore >= tc.expectedMin && noiseScore <= tc.expectedMax,
				"Noise score %f should be between %f and %f", noiseScore, tc.expectedMin, tc.expectedMax)
			assert.True(t, noiseScore >= 0 && noiseScore <= 1,
				"Noise score should be normalized between 0 and 1")
		})
	}
}

// Helper function to create test metrics
func createTestMetrics() []*dto.MetricFamily {
	return []*dto.MetricFamily{
		{
			Name: func(s string) *string { return &s }("test_gauge"),
			Type: dto.MetricType_GAUGE.Enum(),
			Metric: []*dto.Metric{
				{
					Label: []*dto.LabelPair{
						{Name: func(s string) *string { return &s }("job"), Value: func(s string) *string { return &s }("123")},
					},
					Gauge: &dto.Gauge{Value: func(f float64) *float64 { return &f }(42.0)},
				},
				{
					Label: []*dto.LabelPair{
						{Name: func(s string) *string { return &s }("job"), Value: func(s string) *string { return &s }("124")},
					},
					Gauge: &dto.Gauge{Value: func(f float64) *float64 { return &f }(84.0)},
				},
			},
		},
		{
			Name: func(s string) *string { return &s }("test_counter"),
			Type: dto.MetricType_COUNTER.Enum(),
			Metric: []*dto.Metric{
				{
					Counter: &dto.Counter{Value: func(f float64) *float64 { return &f }(100.0)},
				},
			},
		},
	}
}
