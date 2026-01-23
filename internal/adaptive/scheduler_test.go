// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package adaptive

import (
	"context"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollectorScheduler_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled: false,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.False(t, scheduler.IsEnabled())
}

func TestNewCollectorScheduler_Enabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)
	assert.NotNil(t, scheduler)
	assert.True(t, scheduler.IsEnabled())
	assert.Equal(t, cfg, scheduler.GetConfig())
}

func TestNewCollectorScheduler_InvalidConfig(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	testCases := []struct {
		name string
		cfg  config.AdaptiveCollectionConfig
	}{
		{
			name: "min_interval_zero",
			cfg: config.AdaptiveCollectionConfig{
				Enabled:      true,
				MinInterval:  0,
				MaxInterval:  120 * time.Second,
				BaseInterval: 30 * time.Second,
				ScoreWindow:  300 * time.Second,
			},
		},
		{
			name: "min_greater_than_max",
			cfg: config.AdaptiveCollectionConfig{
				Enabled:      true,
				MinInterval:  60 * time.Second,
				MaxInterval:  30 * time.Second,
				BaseInterval: 45 * time.Second,
				ScoreWindow:  300 * time.Second,
			},
		},
		{
			name: "base_outside_range",
			cfg: config.AdaptiveCollectionConfig{
				Enabled:      true,
				MinInterval:  30 * time.Second,
				MaxInterval:  120 * time.Second,
				BaseInterval: 150 * time.Second,
				ScoreWindow:  300 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scheduler, err := NewCollectorScheduler(tc.cfg, 30*time.Second, logger)
			assert.Error(t, err)
			assert.Nil(t, scheduler)
		})
	}
}

func TestCollectorScheduler_RegisterCollector(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	// Register a collector
	scheduler.RegisterCollector("test_collector")

	// Should get base interval
	interval := scheduler.GetCollectionInterval("test_collector")
	assert.Equal(t, cfg.BaseInterval, interval)

	// Stats should show the registered collector
	stats := scheduler.GetStats()
	assert.Equal(t, 1, stats["registered_collectors"])
}

func TestCollectorScheduler_GetCollectionInterval_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled: false,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	// Should return default interval
	interval := scheduler.GetCollectionInterval("test_collector")
	assert.Equal(t, 30*time.Second, interval)
}

func TestCollectorScheduler_UpdateActivity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	// Register a collector
	scheduler.RegisterCollector("test_collector")

	// Test high activity scenario
	changeData := map[string]interface{}{
		"job_count":  1000,
		"node_count": 100,
		"queue_size": 50,
	}

	scheduler.UpdateActivity(context.Background(), 1000, 100, changeData)

	// Should have recorded activity
	history := scheduler.GetActivityHistory()
	assert.Len(t, history, 1)
	assert.True(t, history[0].Score > 0)

	// Get current score
	score := scheduler.GetCurrentScore()
	assert.True(t, score > 0)
	assert.True(t, score <= 1.0)
}

func TestCollectorScheduler_ActivityScoreCalculation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		jobCount    int
		nodeCount   int
		changeData  map[string]interface{}
		expectedMin float64
		expectedMax float64
	}{
		{
			name:      "low_activity",
			jobCount:  10,
			nodeCount: 10,
			changeData: map[string]interface{}{
				"jobs": 10,
			},
			expectedMin: 0.0,
			expectedMax: 0.6,
		},
		{
			name:      "high_activity",
			jobCount:  1000,
			nodeCount: 100,
			changeData: map[string]interface{}{
				"jobs": 1000,
			},
			expectedMin: 0.4,
			expectedMax: 1.0,
		},
		{
			name:      "very_high_density",
			jobCount:  2000,
			nodeCount: 50,
			changeData: map[string]interface{}{
				"jobs": 2000,
			},
			expectedMin: 0.6,
			expectedMax: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := scheduler.calculateActivityScore(tc.jobCount, tc.nodeCount, tc.changeData)
			assert.True(t, result.Score >= tc.expectedMin,
				"Score %f should be >= %f", result.Score, tc.expectedMin)
			assert.True(t, result.Score <= tc.expectedMax,
				"Score %f should be <= %f", result.Score, tc.expectedMax)
			assert.True(t, result.Score >= 0.0 && result.Score <= 1.0,
				"Score should be normalized between 0 and 1")
		})
	}
}

func TestCollectorScheduler_IntervalAdaptation(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	scheduler.RegisterCollector("test_collector")

	// Test high activity should decrease interval
	highActivityData := map[string]interface{}{
		"jobs":  1000,
		"nodes": 100,
	}

	scheduler.UpdateActivity(context.Background(), 1000, 100, highActivityData)
	highActivityInterval := scheduler.GetCollectionInterval("test_collector")

	// Test low activity should increase interval
	lowActivityData := map[string]interface{}{
		"jobs":  5,
		"nodes": 10,
	}

	scheduler.UpdateActivity(context.Background(), 5, 10, lowActivityData)
	lowActivityInterval := scheduler.GetCollectionInterval("test_collector")

	// High activity should result in shorter intervals
	assert.True(t, highActivityInterval < lowActivityInterval,
		"High activity interval (%v) should be less than low activity interval (%v)",
		highActivityInterval, lowActivityInterval)

	// Intervals should be within configured bounds
	assert.True(t, highActivityInterval >= cfg.MinInterval)
	assert.True(t, lowActivityInterval <= cfg.MaxInterval)
}

func TestCollectorScheduler_RecordCollection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	scheduler.RegisterCollector("test_collector")

	// Record a collection with some delay
	scheduledTime := time.Now()
	actualTime := scheduledTime.Add(5 * time.Second)

	scheduler.RecordCollection("test_collector", scheduledTime, actualTime)

	// Should not panic and should record the collection time
	// (detailed verification would require access to internal state)
}

func TestCollectorScheduler_GetStats(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	stats := scheduler.GetStats()
	assert.Equal(t, true, stats["enabled"])
	assert.Equal(t, 0, stats["activity_entries"])
	assert.Equal(t, 0, stats["registered_collectors"])

	// Register collector and update activity
	scheduler.RegisterCollector("test_collector")
	scheduler.UpdateActivity(context.Background(), 100, 50, map[string]interface{}{"test": 1})

	stats = scheduler.GetStats()
	assert.Equal(t, 1, stats["activity_entries"])
	assert.Equal(t, 1, stats["registered_collectors"])
	assert.NotNil(t, stats["current_intervals"])
}

func TestCollectorScheduler_RegisterMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	registry := prometheus.NewRegistry()
	err = scheduler.RegisterMetrics(registry)
	assert.NoError(t, err)

	// Register a collector and trigger some activity to populate metrics
	scheduler.RegisterCollector("test_collector")
	scheduler.UpdateActivity(context.Background(), 100, 50, map[string]interface{}{"test": 1})

	// Record a collection to populate the delay histogram
	scheduledTime := time.Now()
	actualTime := scheduledTime.Add(100 * time.Millisecond)
	scheduler.RecordCollection("test_collector", scheduledTime, actualTime)

	// Test that metrics are registered by gathering them
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.True(t, len(metrics) >= 5, "Expected at least 5 metric families, got %d", len(metrics))
}

func TestCollectorScheduler_RegisterMetrics_Disabled(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled: false,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	registry := prometheus.NewRegistry()
	err = scheduler.RegisterMetrics(registry)
	assert.NoError(t, err)

	// Should not register any metrics when disabled
	metrics, err := registry.Gather()
	assert.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestCollectorScheduler_CompareValues(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		old      interface{}
		new      interface{}
		expected float64
	}{
		{
			name:     "no_change_int",
			old:      100,
			new:      100,
			expected: 0.0,
		},
		{
			name:     "small_increase",
			old:      100,
			new:      110,
			expected: 0.1,
		},
		{
			name:     "large_increase",
			old:      100,
			new:      300,
			expected: 1.0, // Capped at 1.0
		},
		{
			name:     "string_change",
			old:      "running",
			new:      "completed",
			expected: 1.0,
		},
		{
			name:     "string_no_change",
			old:      "running",
			new:      "running",
			expected: 0.0,
		},
		{
			name:     "zero_to_value",
			old:      0,
			new:      50,
			expected: 1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := scheduler.compareValues(tc.old, tc.new)
			assert.InDelta(t, tc.expected, result, 0.01,
				"Expected %f, got %f for comparing %v to %v", tc.expected, result, tc.old, tc.new)
		})
	}
}

func TestCollectorScheduler_TimeBasedScore(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  120 * time.Second,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  300 * time.Second,
	}

	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	require.NoError(t, err)

	score := scheduler.getTimeBasedScore()
	assert.True(t, score >= 0.0 && score <= 1.0, "Time-based score should be between 0 and 1")
}
