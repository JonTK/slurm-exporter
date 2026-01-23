// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

func TestCardinalityManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce test noise

	t.Run("NewCardinalityManager", func(t *testing.T) {
		cm := NewCardinalityManager(logger)
		if cm == nil {
			t.Fatal("Expected non-nil CardinalityManager")
		}
		if cm.globalLimit != 1000000 {
			t.Errorf("Expected global limit 1000000, got %d", cm.globalLimit)
		}
		if len(cm.limits) == 0 {
			t.Error("Expected default limits to be configured")
		}
	})

	t.Run("SetLimit", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		limit := CardinalityLimit{
			MaxSeries:    1000,
			MaxLabels:    5,
			SamplingRate: 0.5,
			Priority:     PriorityHigh,
		}

		cm.SetLimit("test_metric_.*", limit)

		if len(cm.limits) == 0 {
			t.Error("Expected limit to be set")
		}

		storedLimit, exists := cm.limits["test_metric_.*"]
		if !exists {
			t.Error("Expected limit to be stored")
		}
		if storedLimit.MaxSeries != 1000 {
			t.Errorf("Expected MaxSeries 1000, got %d", storedLimit.MaxSeries)
		}
	})

	t.Run("ShouldCollectMetric", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Test with no limits configured
		labels := map[string]string{"label1": "value1"}
		if !cm.ShouldCollectMetric("unknown_metric", labels) {
			t.Error("Expected to collect metric with no limits")
		}

		// Configure a limit
		limit := CardinalityLimit{
			MaxSeries:    2,
			SamplingRate: 1.0,
			Priority:     PriorityMedium,
		}
		cm.SetLimit("limited_metric", limit)

		// Should collect first few metrics
		if !cm.ShouldCollectMetric("limited_metric", map[string]string{"id": "1"}) {
			t.Error("Expected to collect first metric")
		}
		if !cm.ShouldCollectMetric("limited_metric", map[string]string{"id": "2"}) {
			t.Error("Expected to collect second metric")
		}

		// Should reject after limit
		if cm.ShouldCollectMetric("limited_metric", map[string]string{"id": "3"}) {
			t.Error("Expected to reject metric after limit")
		}
	})

	t.Run("SamplingLogic", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Configure sampling
		limit := CardinalityLimit{
			MaxSeries:    1000,
			SamplingRate: 0.5,
			Priority:     PriorityMedium,
		}
		cm.SetLimit("sampled_metric", limit)

		// Test multiple metrics with same labels (should be consistent)
		labels := map[string]string{"test": "value"}
		result1 := cm.ShouldCollectMetric("sampled_metric", labels)
		result2 := cm.ShouldCollectMetric("sampled_metric", labels)

		if result1 != result2 {
			t.Error("Sampling should be consistent for same labels")
		}
	})

	t.Run("PatternMatching", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		testCases := []struct {
			pattern     string
			metricName  string
			shouldMatch bool
		}{
			{"exact_match", "exact_match", true},
			{"prefix_.*", "prefix_test", true},
			{"prefix_.*", "prefix_another", true},
			{"prefix_.*", "other_prefix", false},
			{"slurm_job_.*", "slurm_job_states", true},
			{"slurm_job_.*", "slurm_node_states", false},
		}

		for _, tc := range testCases {
			result := cm.matchesPattern(tc.metricName, tc.pattern)
			if result != tc.shouldMatch {
				t.Errorf("Pattern %s vs metric %s: expected %v, got %v",
					tc.pattern, tc.metricName, tc.shouldMatch, result)
			}
		}
	})

	t.Run("UsageTracking", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		labels1 := map[string]string{"label1": "value1", "label2": "value2"}
		labels2 := map[string]string{"label1": "value3", "label2": "value4"}

		cm.ShouldCollectMetric("test_metric", labels1)
		cm.ShouldCollectMetric("test_metric", labels2)

		usage, exists := cm.usage["test_metric"]
		if !exists {
			t.Fatal("Expected usage to be tracked")
		}

		if usage.SeriesCount != 2 {
			t.Errorf("Expected series count 2, got %d", usage.SeriesCount)
		}
		if usage.LabelCount != 2 {
			t.Errorf("Expected label count 2, got %d", usage.LabelCount)
		}
		if usage.LabelValues["label1"] != 2 {
			t.Errorf("Expected label1 count 2, got %d", usage.LabelValues["label1"])
		}
	})

	t.Run("HashConsistency", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		labels1 := map[string]string{"a": "1", "b": "2"}
		labels2 := map[string]string{"b": "2", "a": "1"} // Different order

		hash1 := cm.hashLabels("test_metric", labels1)
		hash2 := cm.hashLabels("test_metric", labels2)

		if hash1 != hash2 {
			t.Error("Hash should be consistent regardless of label order")
		}

		// Different labels should produce different hashes
		labels3 := map[string]string{"a": "1", "b": "3"}
		hash3 := cm.hashLabels("test_metric", labels3)

		if hash1 == hash3 {
			t.Error("Different labels should produce different hashes")
		}
	})

	t.Run("GetCardinalityReport", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Add some test metrics
		cm.ShouldCollectMetric("metric1", map[string]string{"id": "1"})
		cm.ShouldCollectMetric("metric1", map[string]string{"id": "2"})
		cm.ShouldCollectMetric("metric2", map[string]string{"id": "1"})

		report := cm.GetCardinalityReport()

		if report.GlobalStats.TotalSeries != 3 {
			t.Errorf("Expected total series 3, got %d", report.GlobalStats.TotalSeries)
		}
		if report.GlobalStats.TotalMetrics != 2 {
			t.Errorf("Expected total metrics 2, got %d", report.GlobalStats.TotalMetrics)
		}
		if len(report.MetricStats) != 2 {
			t.Errorf("Expected 2 metric stats, got %d", len(report.MetricStats))
		}

		// Check sorting (highest cardinality first)
		if len(report.MetricStats) >= 2 {
			if report.MetricStats[0].SeriesCount < report.MetricStats[1].SeriesCount {
				t.Error("Metrics should be sorted by series count descending")
			}
		}
	})

	t.Run("ViolationDetection", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Configure a low limit
		limit := CardinalityLimit{
			MaxSeries: 1,
			MaxLabels: 1,
			Priority:  PriorityHigh,
		}
		cm.SetLimit("violation_test", limit)

		// Add metrics that exceed limits
		cm.ShouldCollectMetric("violation_test", map[string]string{"id": "1"})
		cm.ShouldCollectMetric("violation_test", map[string]string{"id": "2"}) // Should be rejected

		// Force add more for testing violations
		usage := &CardinalityUsage{
			MetricName:  "violation_test",
			SeriesCount: 5, // Exceeds limit of 1
			LabelCount:  3, // Exceeds limit of 1
		}
		cm.usage["violation_test"] = usage

		violations := cm.findViolations()

		if len(violations) == 0 {
			t.Error("Expected violations to be detected")
		}

		seriesViolationFound := false
		labelViolationFound := false
		for _, v := range violations {
			if v.ViolationType == "series_limit" {
				seriesViolationFound = true
			}
			if v.ViolationType == "label_limit" {
				labelViolationFound = true
			}
		}

		if !seriesViolationFound {
			t.Error("Expected series limit violation")
		}
		if !labelViolationFound {
			t.Error("Expected label limit violation")
		}
	})

	t.Run("RecommendationGeneration", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Add high-cardinality metric without limits
		usage := &CardinalityUsage{
			MetricName:  "high_cardinality",
			SeriesCount: 2000,
			LabelCount:  15,
		}
		cm.usage["high_cardinality"] = usage

		// Add metric approaching limits
		limit := CardinalityLimit{
			MaxSeries:    100,
			SamplingRate: 1.0,
			Priority:     PriorityMedium,
		}
		cm.SetLimit("approaching_limit", limit)
		usage2 := &CardinalityUsage{
			MetricName:  "approaching_limit",
			SeriesCount: 85, // 85% of limit
			LabelCount:  5,
		}
		cm.usage["approaching_limit"] = usage2

		recommendations := cm.generateRecommendations()

		if len(recommendations) == 0 {
			t.Error("Expected recommendations to be generated")
		}

		foundLimitRec := false
		foundSamplingRec := false
		foundLabelRec := false

		for _, rec := range recommendations {
			switch rec.Type {
			case "add_limit":
				foundLimitRec = true
			case "enable_sampling":
				foundSamplingRec = true
			case "reduce_labels":
				foundLabelRec = true
			}
		}

		if !foundLimitRec {
			t.Error("Expected recommendation to add limits")
		}
		if !foundSamplingRec {
			t.Error("Expected recommendation to enable sampling")
		}
		if !foundLabelRec {
			t.Error("Expected recommendation to reduce labels")
		}
	})

	t.Run("RegisterMetrics", func(t *testing.T) {
		cm := NewCardinalityManager(logger)
		registry := prometheus.NewRegistry()

		err := cm.RegisterMetrics(registry)
		if err != nil {
			t.Errorf("Failed to register metrics: %v", err)
		}

		// Add some test data to ensure metrics have values
		cm.ShouldCollectMetric("test_metric", map[string]string{"label": "value"})

		// Test that metrics can be gathered
		gathering, err := registry.Gather()
		if err != nil {
			t.Errorf("Failed to gather metrics: %v", err)
		}

		// Should have at least one metric family for our cardinality metrics
		if len(gathering) == 0 {
			t.Error("Expected metrics to be registered")
		}

		// Look for at least one of our cardinality metrics
		foundCardinalityMetric := false
		for _, mf := range gathering {
			if strings.Contains(mf.GetName(), "cardinality") {
				foundCardinalityMetric = true
				break
			}
		}

		if !foundCardinalityMetric {
			t.Error("Expected to find cardinality metrics in gathering")
		}
	})

	t.Run("MonitoringStart", func(t *testing.T) {
		cm := NewCardinalityManager(logger)
		cm.checkInterval = 10 * time.Millisecond // Fast for testing

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// This should run briefly and then stop
		cm.StartCardinalityMonitoring(ctx)

		// If we get here, monitoring started and stopped correctly
	})

	t.Run("DefaultLimitsConfiguration", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Test that default limits are properly configured
		expectedPatterns := []string{
			"slurm_job_.*",
			"slurm_node_.*",
			"slurm_cluster_.*|slurm_partition_.*",
			"slurm_user_.*|slurm_account_.*",
			"slurm_exporter_.*",
		}

		for _, pattern := range expectedPatterns {
			if _, exists := cm.limits[pattern]; !exists {
				t.Errorf("Expected default limit for pattern %s", pattern)
			}
		}

		// Test specific configurations
		jobLimit := cm.limits["slurm_job_.*"]
		if jobLimit.SamplingRate != 0.1 {
			t.Errorf("Expected job sampling rate 0.1, got %f", jobLimit.SamplingRate)
		}

		nodeLimit := cm.limits["slurm_node_.*"]
		if nodeLimit.Priority != PriorityHigh {
			t.Errorf("Expected node priority high, got %s", nodeLimit.Priority)
		}

		exporterLimit := cm.limits["slurm_exporter_.*"]
		if exporterLimit.Priority != PriorityCritical {
			t.Errorf("Expected exporter priority critical, got %s", exporterLimit.Priority)
		}
	})

	t.Run("LabelValueLimits", func(t *testing.T) {
		cm := NewCardinalityManager(logger)

		// Configure label-specific limits
		limit := CardinalityLimit{
			MaxSeries: 1000,
			MaxLabelValues: map[string]int{
				"user": 2,
			},
			SamplingRate: 1.0,
			Priority:     PriorityMedium,
		}
		cm.SetLimit("user_metric", limit)

		// Add metrics up to label limit
		cm.ShouldCollectMetric("user_metric", map[string]string{"user": "alice"})
		cm.ShouldCollectMetric("user_metric", map[string]string{"user": "bob"})

		// This should be rejected due to label value limit
		// Note: This is a simplified test - in reality the limit tracking would be more sophisticated
		usage := cm.usage["user_metric"]
		if usage == nil {
			t.Fatal("Expected usage to be tracked")
		}

		// Manually set label value count to test limit
		usage.LabelValues["user"] = 3 // Exceeds limit of 2

		result := cm.ShouldCollectMetric("user_metric", map[string]string{"user": "charlie"})
		if result {
			t.Error("Expected metric to be rejected due to label value limit")
		}
	})
}

func TestCardinalityLimitPriorities(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewCardinalityManager(logger)

	t.Run("PriorityEnforcement", func(t *testing.T) {
		// Test that different priorities are handled correctly
		priorities := []Priority{
			PriorityLow,
			PriorityMedium,
			PriorityHigh,
			PriorityCritical,
		}

		for i, priority := range priorities {
			limit := CardinalityLimit{
				MaxSeries: 1,
				Priority:  priority,
			}

			pattern := fmt.Sprintf("priority_test_%d", i)
			cm.SetLimit(pattern, limit)

			// Add metrics to exceed limit
			cm.ShouldCollectMetric(pattern, map[string]string{"id": "1"})
			shouldReject := !cm.ShouldCollectMetric(pattern, map[string]string{"id": "2"})

			if !shouldReject {
				t.Errorf("Expected metric to be rejected for priority %s", priority)
			}
		}
	})
}

func TestFilterStrategies(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewCardinalityManager(logger)

	t.Run("FilterConfiguration", func(t *testing.T) {
		filter := CardinalityFilter{
			Strategy:   FilterStrategySample,
			SampleRate: 0.5,
			LabelFilters: map[string][]string{
				"environment": {"prod", "staging"},
			},
			DropPatterns:     []string{"debug_.*"},
			PreservePatterns: []string{"critical_.*"},
		}

		cm.SetFilter("test_filter", filter)

		storedFilter, exists := cm.filters["test_filter"]
		if !exists {
			t.Error("Expected filter to be stored")
		}

		if storedFilter.Strategy != FilterStrategySample {
			t.Errorf("Expected sample strategy, got %s", storedFilter.Strategy)
		}
		if storedFilter.SampleRate != 0.5 {
			t.Errorf("Expected sample rate 0.5, got %f", storedFilter.SampleRate)
		}
	})

	t.Run("FilterStrategiesEnum", func(t *testing.T) {
		strategies := []FilterStrategy{
			FilterStrategySample,
			FilterStrategyDrop,
			FilterStrategyAggregate,
			FilterStrategyLimit,
		}

		for _, strategy := range strategies {
			if string(strategy) == "" {
				t.Errorf("Filter strategy should not be empty: %v", strategy)
			}
		}
	})
}

func TestCardinalityReport(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewCardinalityManager(logger)

	t.Run("ReportStructure", func(t *testing.T) {
		// Add test data
		cm.ShouldCollectMetric("metric_a", map[string]string{"id": "1"})
		cm.ShouldCollectMetric("metric_a", map[string]string{"id": "2"})
		cm.ShouldCollectMetric("metric_b", map[string]string{"id": "1"})

		report := cm.GetCardinalityReport()

		// Verify report structure
		if report.GeneratedAt.IsZero() {
			t.Error("Expected report to have generation timestamp")
		}

		if report.GlobalStats.TotalSeries == 0 {
			t.Error("Expected global stats to be populated")
		}

		if len(report.MetricStats) == 0 {
			t.Error("Expected metric stats to be populated")
		}

		// Verify each metric stat has required fields
		for _, stat := range report.MetricStats {
			if stat.MetricName == "" {
				t.Error("Expected metric name to be set")
			}
			if stat.LastUpdate.IsZero() {
				t.Error("Expected last update to be set")
			}
		}
	})

	t.Run("ReportSorting", func(t *testing.T) {
		// Clear previous usage
		cm.usage = make(map[string]*CardinalityUsage)

		// Add metrics with different cardinalities
		for i := 1; i <= 5; i++ {
			for j := 1; j <= i; j++ {
				cm.ShouldCollectMetric(fmt.Sprintf("metric_%d", i), map[string]string{"id": fmt.Sprintf("%d", j)})
			}
		}

		report := cm.GetCardinalityReport()

		// Verify sorting (descending by series count)
		for i := 1; i < len(report.MetricStats); i++ {
			if report.MetricStats[i-1].SeriesCount < report.MetricStats[i].SeriesCount {
				t.Error("Metric stats should be sorted by series count descending")
			}
		}
	})
}

func BenchmarkCardinalityCheck(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewCardinalityManager(logger)

	// Setup test data
	for i := 0; i < 1000; i++ {
		labels := map[string]string{
			"metric_id": fmt.Sprintf("metric_%d", i),
			"user":      fmt.Sprintf("user_%d", i%100),
			"partition": fmt.Sprintf("partition_%d", i%10),
		}
		cm.ShouldCollectMetric("benchmark_metric", labels)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		labels := map[string]string{
			"metric_id": fmt.Sprintf("metric_%d", i),
			"user":      fmt.Sprintf("user_%d", i%100),
			"partition": fmt.Sprintf("partition_%d", i%10),
		}
		cm.ShouldCollectMetric("benchmark_metric", labels)
	}
}

func BenchmarkHashLabels(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	cm := NewCardinalityManager(logger)

	labels := map[string]string{
		"cluster_name": "production",
		"node_name":    "compute-001",
		"job_id":       "12345",
		"user":         "alice",
		"partition":    "gpu",
		"state":        "running",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cm.hashLabels("test_metric", labels)
	}
}
