// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"strings"
	"testing"
)

func TestLabelManager(t *testing.T) {
	t.Run("ValidateLabels", func(t *testing.T) {
		lm := NewLabelManager()
		validLabels := map[string]string{
			"cluster_name": "production",
			"job_id":       "12345",
			"user":         "alice",
			"partition":    "compute",
			"job_state":    "running",
		}

		err := lm.ValidateLabels(validLabels)
		if err != nil {
			t.Errorf("Valid labels should not produce error: %v", err)
		}
	})

	t.Run("ValidateInvalidLabels", func(t *testing.T) {
		lm := NewLabelManager()
		invalidLabels := map[string]string{
			"job_state": "invalid_state",
		}

		err := lm.ValidateLabels(invalidLabels)
		if err == nil {
			t.Error("Invalid labels should produce error")
		}
	})

	t.Run("ValidateLabel", func(t *testing.T) {
		lm := NewLabelManager()
		testCases := []struct {
			labelName  string
			labelValue string
			shouldFail bool
		}{
			{"cluster_name", "prod-cluster", false},
			{"job_id", "12345", false},
			{"job_id", "12345_0", false},
			{"job_id", "12345.0", false},
			{"job_state", "running", false},
			{"job_state", "invalid", true},
			{"node_state", "idle", false},
			{"node_state", "invalid_state", true},
			{"unknown_label", "value", false}, // Now allows unknown labels
		}

		for _, tc := range testCases {
			err := lm.ValidateLabel(tc.labelName, tc.labelValue)
			if tc.shouldFail && err == nil {
				t.Errorf("Expected validation to fail for %s=%s", tc.labelName, tc.labelValue)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected validation to pass for %s=%s, got error: %v", tc.labelName, tc.labelValue, err)
			}
		}
	})

	t.Run("SanitizeLabels", func(t *testing.T) {
		lm := NewLabelManager()
		labels := map[string]string{
			"cluster_name": "Production Cluster",
			"user":         "alice.smith",
			"partition":    "gpu-nodes",
		}

		sanitized := lm.SanitizeLabels(labels)

		expected := map[string]string{
			"cluster_name": "production_cluster",
			"user":         "alice.smith",
			"partition":    "gpu_nodes",
		}

		for key, expectedValue := range expected {
			if sanitized[key] != expectedValue {
				t.Errorf("Expected sanitized %s='%s', got '%s'", key, expectedValue, sanitized[key])
			}
		}
	})

	t.Run("TrackLabelUsage", func(t *testing.T) {
		lm := NewLabelManager()
		lm.TrackLabelUsage("cluster_name", "test")
		lm.TrackLabelUsage("cluster_name", "prod")
		lm.TrackLabelUsage("job_state", "running")

		analysis := lm.GetDimensionalAnalysis()

		if analysis["cluster_name"].CurrentCardinality != 2 {
			t.Errorf("Expected cluster_name cardinality 2, got %d", analysis["cluster_name"].CurrentCardinality)
		}
		if analysis["job_state"].CurrentCardinality != 1 {
			t.Errorf("Expected job_state cardinality 1, got %d", analysis["job_state"].CurrentCardinality)
		}
	})

	t.Run("GetDimensionalAnalysis", func(t *testing.T) {
		lm := NewLabelManager()
		analysis := lm.GetDimensionalAnalysis()

		// Check that all standard labels are present
		expectedLabels := []string{"cluster_name", "node_name", "job_id", "user", "account", "partition"}
		for _, labelName := range expectedLabels {
			if _, exists := analysis[labelName]; !exists {
				t.Errorf("Expected label %s in dimensional analysis", labelName)
			}
		}

		// Check structure of analysis
		for labelName, info := range analysis {
			if info.LabelName != labelName {
				t.Errorf("Label name mismatch: expected %s, got %s", labelName, info.LabelName)
			}
			if info.MaxCardinality <= 0 {
				t.Errorf("Max cardinality should be positive for %s", labelName)
			}
		}
	})

	t.Run("GetHighCardinalityLabels", func(t *testing.T) {
		lm := NewLabelManager()
		// Simulate high cardinality usage
		for i := 0; i < 15; i++ {
			lm.TrackLabelUsage("job_state", "running")
		}

		highCardinality := lm.GetHighCardinalityLabels(50.0)

		// job_state has max cardinality 10, so 15 uses = 150% utilization
		found := false
		for _, info := range highCardinality {
			if info.LabelName == "job_state" && info.UtilizationPercent > 50.0 {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find job_state in high cardinality labels")
		}
	})

	t.Run("BuildLabelSet", func(t *testing.T) {
		lm := NewLabelManager()
		labels := map[string]string{
			"cluster_name": "Test Cluster",
			"job_state":    "running",
			"user":         "alice",
		}

		labelSet, err := lm.BuildLabelSet(labels)
		if err != nil {
			t.Errorf("BuildLabelSet failed: %v", err)
		}

		if labelSet["cluster_name"] != "test_cluster" {
			t.Errorf("Expected sanitized cluster_name, got %s", labelSet["cluster_name"])
		}
		if labelSet["job_state"] != "running" {
			t.Errorf("Expected job_state to remain 'running', got %s", labelSet["job_state"])
		}
	})

	t.Run("GetLabelRecommendations", func(t *testing.T) {
		lm := NewLabelManager()
		// Create high cardinality scenario
		for i := 0; i < 90; i++ {
			lm.TrackLabelUsage("partition", "compute")
		}

		recommendations := lm.GetLabelRecommendations()

		found := false
		for _, rec := range recommendations {
			if rec.LabelName == "partition" && rec.Type == RecommendationTypeCardinality {
				found = true
				if rec.Severity != SeverityWarning {
					t.Errorf("Expected warning severity for high cardinality, got %s", rec.Severity)
				}
				break
			}
		}

		if !found {
			t.Error("Expected cardinality recommendation for partition label")
		}
	})
}

func TestLabelDimension(t *testing.T) {
	t.Run("StandardLabels", func(t *testing.T) {
		// Test that all standard labels are properly defined
		requiredFields := []string{"cluster_name", "job_state", "node_state", "partition_state"}

		for _, fieldName := range requiredFields {
			dimension, exists := StandardLabels[fieldName]
			if !exists {
				t.Errorf("Standard label %s not found", fieldName)
				continue
			}

			if dimension.Name != fieldName {
				t.Errorf("Label name mismatch for %s", fieldName)
			}
			if dimension.Description == "" {
				t.Errorf("Description missing for label %s", fieldName)
			}
			if dimension.MaxCardinality <= 0 {
				t.Errorf("Max cardinality should be positive for %s", fieldName)
			}
		}
	})

	t.Run("LabelTypes", func(t *testing.T) {
		expectedTypes := []LabelType{
			LabelTypeCluster,
			LabelTypeNode,
			LabelTypeJob,
			LabelTypeUser,
			LabelTypeAccount,
			LabelTypePartition,
			LabelTypeState,
			LabelTypeResource,
			LabelTypeError,
			LabelTypeTemporal,
		}

		for _, labelType := range expectedTypes {
			if string(labelType) == "" {
				t.Errorf("Label type should not be empty")
			}
		}
	})
}

func TestLabelValidation(t *testing.T) {
	lm := NewLabelManager()

	t.Run("ClusterNameValidation", func(t *testing.T) {
		validNames := []string{"cluster1", "prod-cluster", "test_env", "c123", "123", "cluster with spaces"}
		invalidNames := []string{"", "-invalid"}

		for _, name := range validNames {
			if err := lm.ValidateLabel("cluster_name", name); err != nil {
				t.Errorf("Valid cluster name '%s' should pass validation: %v", name, err)
			}
		}

		for _, name := range invalidNames {
			if err := lm.ValidateLabel("cluster_name", name); err == nil {
				t.Errorf("Invalid cluster name '%s' should fail validation", name)
			}
		}
	})

	t.Run("JobIDValidation", func(t *testing.T) {
		validIDs := []string{"123", "12345_0", "123.456", "999_1.0"}
		invalidIDs := []string{"", "abc", "123-456", "job_123"}

		for _, id := range validIDs {
			if err := lm.ValidateLabel("job_id", id); err != nil {
				t.Errorf("Valid job ID '%s' should pass validation: %v", id, err)
			}
		}

		for _, id := range invalidIDs {
			if err := lm.ValidateLabel("job_id", id); err == nil {
				t.Errorf("Invalid job ID '%s' should fail validation", id)
			}
		}
	})

	t.Run("StateValidation", func(t *testing.T) {
		validJobStates := []string{"pending", "running", "completed", "failed"}
		invalidJobStates := []string{"RUNNING", "unknown", "queued"}

		for _, state := range validJobStates {
			if err := lm.ValidateLabel("job_state", state); err != nil {
				t.Errorf("Valid job state '%s' should pass validation: %v", state, err)
			}
		}

		for _, state := range invalidJobStates {
			if err := lm.ValidateLabel("job_state", state); err == nil {
				t.Errorf("Invalid job state '%s' should fail validation", state)
			}
		}
	})
}

func TestLabelSanitization(t *testing.T) {
	lm := NewLabelManager()

	t.Run("ValueSanitization", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"Simple", "simple"},
			{"with spaces", "with_spaces"},
			{"with-dashes", "with_dashes"},
			{"123numbers", "123numbers"},
			{"MixedCase", "mixedcase"},
			{"special@chars", "special@chars"}, // Only spaces and dashes are replaced
		}

		for _, tc := range testCases {
			result := lm.sanitizeValue(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeValue('%s') = '%s', expected '%s'", tc.input, result, tc.expected)
			}
		}
	})

	t.Run("SensitiveLabelHandling", func(t *testing.T) {
		labels := map[string]string{
			"user":         "alice.smith",
			"cluster_name": "production",
		}

		sanitized := lm.SanitizeLabels(labels)

		// User is marked as sensitive in StandardLabels
		if sanitized["user"] == "" {
			t.Error("Sensitive labels should still be sanitized, not removed")
		}
		if sanitized["cluster_name"] != "production" {
			t.Error("Non-sensitive labels should be preserved")
		}
	})
}

func TestDimensionalAnalysisReport(t *testing.T) {
	lm := NewLabelManager()

	// Add some test data
	for i := 0; i < 50; i++ {
		lm.TrackLabelUsage("job_state", "running")
	}
	for i := 0; i < 20; i++ {
		lm.TrackLabelUsage("cluster_name", "prod")
	}

	t.Run("ExportDimensionalAnalysis", func(t *testing.T) {
		report := lm.ExportDimensionalAnalysis()

		if !strings.Contains(report, "SLURM Exporter Dimensional Analysis Report") {
			t.Error("Report should contain title")
		}
		if !strings.Contains(report, "job_state") {
			t.Error("Report should contain job_state information")
		}
		if !strings.Contains(report, "cluster_name") {
			t.Error("Report should contain cluster_name information")
		}
		if !strings.Contains(report, "Label Cardinality Summary") {
			t.Error("Report should contain cardinality summary")
		}
		if !strings.Contains(report, "Recommendations") {
			t.Error("Report should contain recommendations section")
		}
	})
}

func TestCardinalityInfo(t *testing.T) {
	t.Run("CardinalityInfoFields", func(t *testing.T) {
		info := CardinalityInfo{
			LabelName:          "test_label",
			CurrentCardinality: 50,
			MaxCardinality:     100,
			UtilizationPercent: 50.0,
			LabelType:          LabelTypeJob,
			Required:           true,
			Sensitive:          false,
		}

		if info.LabelName != "test_label" {
			t.Errorf("Expected LabelName 'test_label', got '%s'", info.LabelName)
		}
		if info.UtilizationPercent != 50.0 {
			t.Errorf("Expected UtilizationPercent 50.0, got %.1f", info.UtilizationPercent)
		}
		if info.LabelType != LabelTypeJob {
			t.Errorf("Expected LabelType %s, got %s", LabelTypeJob, info.LabelType)
		}
	})
}

func TestLabelRecommendations(t *testing.T) {
	t.Run("RecommendationTypes", func(t *testing.T) {
		rec := LabelRecommendation{
			LabelName:  "test_label",
			Type:       RecommendationTypeCardinality,
			Severity:   SeverityWarning,
			Message:    "Test message",
			Suggestion: "Test suggestion",
		}

		if rec.Type != RecommendationTypeCardinality {
			t.Errorf("Expected type %s, got %s", RecommendationTypeCardinality, rec.Type)
		}
		if rec.Severity != SeverityWarning {
			t.Errorf("Expected severity %s, got %s", SeverityWarning, rec.Severity)
		}
	})

	t.Run("SeverityLevels", func(t *testing.T) {
		severities := []Severity{SeverityInfo, SeverityWarning, SeverityCritical}
		for _, severity := range severities {
			if string(severity) == "" {
				t.Error("Severity should not be empty")
			}
		}
	})
}
