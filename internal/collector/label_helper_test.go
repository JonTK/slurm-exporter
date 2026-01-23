// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jontk/slurm-exporter/internal/metrics"
)

func TestLabelHelper(t *testing.T) {
	logger, _ := test.NewNullLogger()
	logEntry := logrus.NewEntry(logger)
	labelManager := metrics.NewLabelManager()
	labelHelper := NewLabelHelper(labelManager, logEntry)

	t.Run("ValidateAndBuildMetric", func(t *testing.T) {
		desc := prometheus.NewDesc(
			"test_metric",
			"Test metric",
			[]string{"cluster_name", "job_state"},
			nil,
		)

		metric, err := labelHelper.ValidateAndBuildMetric(
			desc,
			prometheus.GaugeValue,
			1.0,
			[]string{"cluster_name", "job_state"},
			"test-cluster", "running",
		)

		if err != nil {
			t.Errorf("ValidateAndBuildMetric failed: %v", err)
		}
		if metric == nil {
			t.Error("Expected metric to be created")
		}
	})

	t.Run("ValidateAndBuildMetricLabelMismatch", func(t *testing.T) {
		desc := prometheus.NewDesc(
			"test_metric",
			"Test metric",
			[]string{"cluster_name"},
			nil,
		)

		_, err := labelHelper.ValidateAndBuildMetric(
			desc,
			prometheus.GaugeValue,
			1.0,
			[]string{"cluster_name", "job_state"},
			"test-cluster",
		)

		if err == nil {
			t.Error("Expected error for label count mismatch")
		}
	})
}

func TestStandardLabelSets(t *testing.T) {
	labelManager := metrics.NewLabelManager()
	sls := NewStandardLabelSets(labelManager)

	t.Run("ClusterLabels", func(t *testing.T) {
		labels, err := sls.ClusterLabels("test-cluster")
		if err != nil {
			t.Errorf("ClusterLabels failed: %v", err)
		}
		if labels["cluster_name"] != "test_cluster" {
			t.Errorf("Expected sanitized cluster name, got %s", labels["cluster_name"])
		}
	})

	t.Run("NodeLabels", func(t *testing.T) {
		labels, err := sls.NodeLabels("test-cluster", "node01")
		if err != nil {
			t.Errorf("NodeLabels failed: %v", err)
		}
		if labels["cluster_name"] != "test_cluster" {
			t.Errorf("Expected cluster name, got %s", labels["cluster_name"])
		}
		if labels["node_name"] != "node01" {
			t.Errorf("Expected node name, got %s", labels["node_name"])
		}
	})

	t.Run("JobLabels", func(t *testing.T) {
		labels, err := sls.JobLabels("test-cluster", "12345", "alice", "ml_team", "compute")
		if err != nil {
			t.Errorf("JobLabels failed: %v", err)
		}
		if labels["job_id"] != "12345" {
			t.Errorf("Expected job ID '12345', got '%s'", labels["job_id"])
		}
		if labels["user"] != "alice" {
			t.Errorf("Expected user, got %s", labels["user"])
		}
		if labels["account"] != "ml_team" {
			t.Errorf("Expected account, got %s", labels["account"])
		}
		if labels["partition"] != "compute" {
			t.Errorf("Expected partition, got %s", labels["partition"])
		}
	})

	t.Run("PartitionLabels", func(t *testing.T) {
		labels, err := sls.PartitionLabels("test-cluster", "gpu")
		if err != nil {
			t.Errorf("PartitionLabels failed: %v", err)
		}
		if labels["partition"] != "gpu" {
			t.Errorf("Expected partition name, got %s", labels["partition"])
		}
	})

	t.Run("UserLabels", func(t *testing.T) {
		labels, err := sls.UserLabels("test-cluster", "bob", "physics")
		if err != nil {
			t.Errorf("UserLabels failed: %v", err)
		}
		if labels["user"] != "bob" {
			t.Errorf("Expected user, got %s", labels["user"])
		}
		if labels["account"] != "physics" {
			t.Errorf("Expected account, got %s", labels["account"])
		}
	})

	t.Run("StateLabels", func(t *testing.T) {
		labels, err := sls.StateLabels("test-cluster", "job123", "running")
		if err != nil {
			t.Errorf("StateLabels failed: %v", err)
		}
		if labels["job_state"] != "running" {
			t.Errorf("Expected job_state 'running', got '%s'", labels["job_state"])
		}
	})
}

func TestLabelSetBuilder(t *testing.T) {
	labelManager := metrics.NewLabelManager()

	t.Run("FluentInterface", func(t *testing.T) {
		builder := NewLabelSetBuilder(labelManager)

		labels, err := builder.
			WithCluster("test-cluster").
			WithNode("node01").
			WithPartition("compute").
			WithState("job", "running").
			WithResource("cpu").
			WithMetricType("utilization").
			WithTimeWindow("1h").
			Build()

		if err != nil {
			t.Errorf("LabelSetBuilder failed: %v", err)
		}

		expectedLabels := map[string]string{
			"cluster_name":  "test_cluster",
			"node_name":     "node01",
			"partition":     "compute",
			"job_state":     "running",
			"resource_type": "cpu",
			"metric_type":   "utilization",
			"time_window":   "1h",
		}

		for key, expectedValue := range expectedLabels {
			if labels[key] != expectedValue {
				t.Errorf("Expected %s='%s', got '%s'", key, expectedValue, labels[key])
			}
		}
	})

	t.Run("WithJob", func(t *testing.T) {
		builder := NewLabelSetBuilder(labelManager)

		labels, err := builder.
			WithCluster("test-cluster").
			WithJob("12345", "alice", "ml_team", "gpu").
			Build()

		if err != nil {
			t.Errorf("WithJob failed: %v", err)
		}

		if labels["job_id"] != "12345" {
			t.Errorf("Expected job_id='12345', got '%s'", labels["job_id"])
		}
		if labels["user"] != "alice" {
			t.Errorf("Expected user='alice', got '%s'", labels["user"])
		}
		if labels["account"] != "ml_team" {
			t.Errorf("Expected account='ml_team', got '%s'", labels["account"])
		}
		if labels["partition"] != "gpu" {
			t.Errorf("Expected partition='gpu', got '%s'", labels["partition"])
		}
	})

	t.Run("WithCustom", func(t *testing.T) {
		builder := NewLabelSetBuilder(labelManager)

		labels, err := builder.
			WithCluster("test-cluster").
			WithCustom("custom_label", "custom_value").
			Build()

		if err != nil {
			t.Errorf("WithCustom failed: %v", err)
		}

		if labels["custom_label"] != "custom_value" {
			t.Errorf("Expected custom_label='custom_value', got '%s'", labels["custom_label"])
		}
	})

	t.Run("BuildSlice", func(t *testing.T) {
		builder := NewLabelSetBuilder(labelManager)

		names, values, err := builder.
			WithCluster("test-cluster").
			WithPartition("compute").
			BuildSlice()

		if err != nil {
			t.Errorf("BuildSlice failed: %v", err)
		}

		if len(names) != len(values) {
			t.Errorf("Names and values length mismatch: %d vs %d", len(names), len(values))
		}

		if len(names) < 2 {
			t.Errorf("Expected at least 2 labels, got %d", len(names))
		}

		// Check that cluster_name and partition are present
		foundCluster := false
		foundPartition := false
		for i, name := range names {
			if name == "cluster_name" && values[i] == "test_cluster" {
				foundCluster = true
			}
			if name == "partition" && values[i] == "compute" {
				foundPartition = true
			}
		}

		if !foundCluster {
			t.Error("Expected to find cluster_name label")
		}
		if !foundPartition {
			t.Error("Expected to find partition label")
		}
	})
}

func TestDimensionalAnalysisReporter(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logEntry := logrus.NewEntry(logger)
	labelManager := metrics.NewLabelManager()
	reporter := NewDimensionalAnalysisReporter(labelManager, logEntry)

	// Add some test data
	for i := 0; i < 50; i++ {
		labelManager.TrackLabelUsage("job_state", "running")
	}

	t.Run("LogCardinalityAnalysis", func(t *testing.T) {
		hook.Reset()
		reporter.LogCardinalityAnalysis()

		// Check that log entries were created
		if len(hook.Entries) == 0 {
			t.Error("Expected log entries for cardinality analysis")
		}

		// Check for summary log
		foundSummary := false
		for _, entry := range hook.Entries {
			if entry.Message == "Dimensional Analysis Summary:" {
				foundSummary = true
				break
			}
		}
		if !foundSummary {
			t.Error("Expected to find dimensional analysis summary log")
		}
	})

	t.Run("LogHighCardinalityWarnings", func(t *testing.T) {
		hook.Reset()
		reporter.LogHighCardinalityWarnings(30.0)

		// Should have warnings since job_state has high utilization
		foundWarning := false
		for _, entry := range hook.Entries {
			if entry.Level == logrus.WarnLevel {
				foundWarning = true
				break
			}
		}
		if !foundWarning {
			t.Error("Expected to find high cardinality warnings")
		}
	})

	t.Run("LogRecommendations", func(t *testing.T) {
		hook.Reset()
		reporter.LogRecommendations()

		// Should have recommendations logged
		if len(hook.Entries) == 0 {
			t.Error("Expected log entries for recommendations")
		}
	})
}

func TestGetLabelNames(t *testing.T) {
	t.Run("GetLabelNames", func(t *testing.T) {
		desc := prometheus.NewDesc(
			"test_metric",
			"Test metric",
			[]string{"cluster_name", "job_state"},
			nil,
		)

		names := GetLabelNames(desc)

		// The function returns a default set since we can't easily extract from desc
		if len(names) == 0 {
			t.Error("Expected at least one label name")
		}
	})
}
