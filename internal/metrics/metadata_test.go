// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"strings"
	"testing"
)

func TestMetricMetadata(t *testing.T) {
	t.Run("GetMetricMetadata", func(t *testing.T) {
		// Test existing metric
		metadata, exists := GetMetricMetadata("slurm_cluster_info")
		if !exists {
			t.Error("Expected slurm_cluster_info to exist in registry")
		}
		if metadata.Name != "slurm_cluster_info" {
			t.Errorf("Expected name 'slurm_cluster_info', got '%s'", metadata.Name)
		}
		if metadata.Type != MetricTypeGauge {
			t.Errorf("Expected type %s, got %s", MetricTypeGauge, metadata.Type)
		}
		if metadata.Category != CategoryCluster {
			t.Errorf("Expected category %s, got %s", CategoryCluster, metadata.Category)
		}

		// Test non-existing metric
		_, exists = GetMetricMetadata("non_existent_metric")
		if exists {
			t.Error("Expected non_existent_metric to not exist")
		}
	})

	t.Run("GetMetricsByCategory", func(t *testing.T) {
		clusterMetrics := GetMetricsByCategory(CategoryCluster)
		if len(clusterMetrics) == 0 {
			t.Error("Expected at least one cluster metric")
		}

		// Verify all returned metrics are in the correct category
		for _, metric := range clusterMetrics {
			if metric.Category != CategoryCluster {
				t.Errorf("Expected category %s, got %s for metric %s", CategoryCluster, metric.Category, metric.Name)
			}
		}

		// Test empty category
		testMetrics := GetMetricsByCategory("non_existent_category")
		if len(testMetrics) != 0 {
			t.Error("Expected no metrics for non-existent category")
		}
	})

	t.Run("GetMetricsByStability", func(t *testing.T) {
		stableMetrics := GetMetricsByStability(StabilityStable)
		if len(stableMetrics) == 0 {
			t.Error("Expected at least one stable metric")
		}

		// Verify all returned metrics have the correct stability level
		for _, metric := range stableMetrics {
			if metric.StabilityLevel != StabilityStable {
				t.Errorf("Expected stability %s, got %s for metric %s", StabilityStable, metric.StabilityLevel, metric.Name)
			}
		}

		betaMetrics := GetMetricsByStability(StabilityBeta)
		// Beta metrics may or may not exist, just check that filtering works
		for _, metric := range betaMetrics {
			if metric.StabilityLevel != StabilityBeta {
				t.Errorf("Expected stability %s, got %s for metric %s", StabilityBeta, metric.StabilityLevel, metric.Name)
			}
		}
	})

	t.Run("ValidateMetricMetadata", func(t *testing.T) {
		issues := ValidateMetricMetadata()
		if len(issues) > 0 {
			t.Errorf("Metadata validation failed with issues: %v", issues)
		}
	})

	t.Run("GetMetricCount", func(t *testing.T) {
		count := GetMetricCount()
		if count == 0 {
			t.Error("Expected at least one metric in registry")
		}
		if count != len(MetricsRegistry) {
			t.Errorf("Expected count %d, got %d", len(MetricsRegistry), count)
		}
	})

	t.Run("GetCategoryCount", func(t *testing.T) {
		counts := GetCategoryCount()
		if len(counts) == 0 {
			t.Error("Expected at least one category")
		}

		totalCount := 0
		for _, count := range counts {
			totalCount += count
		}

		if totalCount != len(MetricsRegistry) {
			t.Errorf("Category counts don't match total metrics: %d vs %d", totalCount, len(MetricsRegistry))
		}
	})
}

func TestMetricTypes(t *testing.T) {
	t.Run("MetricType", func(t *testing.T) {
		types := []MetricType{MetricTypeGauge, MetricTypeCounter, MetricTypeHistogram, MetricTypeSummary}
		for _, metricType := range types {
			if string(metricType) == "" {
				t.Error("Metric type should not be empty")
			}
		}
	})

	t.Run("MetricCategory", func(t *testing.T) {
		categories := []MetricCategory{
			CategoryCluster, CategoryNode, CategoryJob, CategoryUser,
			CategoryAccount, CategoryPartition, CategoryPerformance, CategoryExporter,
		}
		for _, category := range categories {
			if string(category) == "" {
				t.Error("Metric category should not be empty")
			}
		}
	})

	t.Run("StabilityLevel", func(t *testing.T) {
		levels := []StabilityLevel{
			StabilityStable, StabilityBeta, StabilityAlpha,
			StabilityDeprecated, StabilityExperimental,
		}
		for _, level := range levels {
			if string(level) == "" {
				t.Error("Stability level should not be empty")
			}
		}
	})
}

func TestDocumentationGeneration(t *testing.T) {
	t.Run("GenerateDocumentation", func(t *testing.T) {
		doc := GenerateDocumentation()
		if doc == "" {
			t.Error("Expected non-empty documentation")
		}

		// Check for expected sections
		expectedSections := []string{
			"# SLURM Exporter Metrics Reference",
			"## Table of Contents",
			"## Metric Stability Levels",
		}

		for _, section := range expectedSections {
			if !strings.Contains(doc, section) {
				t.Errorf("Documentation missing expected section: %s", section)
			}
		}

		// Check that metric names appear in documentation
		for name := range MetricsRegistry {
			if !strings.Contains(doc, name) {
				t.Errorf("Documentation missing metric: %s", name)
			}
		}
	})

	t.Run("GeneratePrometheusConfig", func(t *testing.T) {
		config := GeneratePrometheusConfig()
		if config == "" {
			t.Error("Expected non-empty Prometheus config")
		}

		expectedElements := []string{
			"global:",
			"scrape_configs:",
			"job_name: 'slurm-exporter'",
			"targets:",
		}

		for _, element := range expectedElements {
			if !strings.Contains(config, element) {
				t.Errorf("Prometheus config missing expected element: %s", element)
			}
		}
	})

	t.Run("GenerateGrafanaDashboard", func(t *testing.T) {
		dashboard := GenerateGrafanaDashboard()
		if dashboard == "" {
			t.Error("Expected non-empty Grafana dashboard")
		}

		expectedElements := []string{
			"dashboard",
			"SLURM Cluster Overview",
			"panels",
			"CPU Utilization",
			"Job States",
		}

		for _, element := range expectedElements {
			if !strings.Contains(dashboard, element) {
				t.Errorf("Grafana dashboard missing expected element: %s", element)
			}
		}
	})
}

func TestMetricMetadataStructure(t *testing.T) {
	t.Run("RequiredFields", func(t *testing.T) {
		// Test that key metrics have all required fields
		keyMetrics := []string{
			"slurm_cluster_info",
			"slurm_node_state",
			"slurm_job_states",
			"slurm_exporter_collection_duration_seconds",
		}

		for _, metricName := range keyMetrics {
			metadata, exists := GetMetricMetadata(metricName)
			if !exists {
				t.Errorf("Key metric %s not found in registry", metricName)
				continue
			}

			if metadata.Name == "" {
				t.Errorf("Metric %s missing name", metricName)
			}
			if metadata.Help == "" {
				t.Errorf("Metric %s missing help", metricName)
			}
			if metadata.Description == "" {
				t.Errorf("Metric %s missing description", metricName)
			}
			if metadata.Type == "" {
				t.Errorf("Metric %s missing type", metricName)
			}
			if metadata.Category == "" {
				t.Errorf("Metric %s missing category", metricName)
			}
			if metadata.StabilityLevel == "" {
				t.Errorf("Metric %s missing stability level", metricName)
			}
		}
	})

	t.Run("ExampleValues", func(t *testing.T) {
		// Test that metrics with example values have consistent structure
		for name, metadata := range MetricsRegistry {
			if metadata.ExampleValue != "" {
				if len(metadata.ExampleLabels) > 0 && len(metadata.Labels) == 0 {
					t.Errorf("Metric %s has example labels but no label definition", name)
				}
			}
		}
	})

	t.Run("RelatedMetrics", func(t *testing.T) {
		// Test that related metrics exist in registry
		for name, metadata := range MetricsRegistry {
			for _, relatedMetric := range metadata.RelatedMetrics {
				if _, exists := MetricsRegistry[relatedMetric]; !exists {
					t.Errorf("Metric %s references non-existent related metric: %s", name, relatedMetric)
				}
			}
		}
	})

	t.Run("LabelConsistency", func(t *testing.T) {
		// Test that example labels match defined labels
		for name, metadata := range MetricsRegistry {
			if len(metadata.ExampleLabels) > 0 && len(metadata.Labels) > 0 {
				for labelName := range metadata.ExampleLabels {
					found := false
					for _, definedLabel := range metadata.Labels {
						if labelName == definedLabel {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Metric %s has example label '%s' not in defined labels", name, labelName)
					}
				}
			}
		}
	})
}

func TestDocumentationQuality(t *testing.T) {
	t.Run("HelpTextQuality", func(t *testing.T) {
		for name, metadata := range MetricsRegistry {
			// Help text should be concise
			if len(metadata.Help) > 100 {
				t.Errorf("Metric %s help text too long (%d chars): %s", name, len(metadata.Help), metadata.Help)
			}

			// Help text should not end with period (Prometheus convention)
			if strings.HasSuffix(metadata.Help, ".") {
				t.Errorf("Metric %s help text should not end with period: %s", name, metadata.Help)
			}

			// Description should be more detailed than help
			if len(metadata.Description) <= len(metadata.Help) {
				t.Errorf("Metric %s description should be more detailed than help", name)
			}
		}
	})

	t.Run("UseCaseCompleness", func(t *testing.T) {
		// Key metrics should have use cases and troubleshooting info
		keyMetrics := []string{
			"slurm_cluster_info",
			"slurm_node_state",
			"slurm_job_states",
		}

		for _, metricName := range keyMetrics {
			metadata, _ := GetMetricMetadata(metricName)
			if metadata.UseCase == "" {
				t.Errorf("Key metric %s missing use case information", metricName)
			}
			if metadata.Troubleshooting == "" {
				t.Errorf("Key metric %s missing troubleshooting information", metricName)
			}
		}
	})
}

func TestCategoryDistribution(t *testing.T) {
	t.Run("CategoryBalance", func(t *testing.T) {
		counts := GetCategoryCount()

		// Should have metrics in all major categories
		expectedCategories := []MetricCategory{
			CategoryCluster, CategoryNode, CategoryJob, CategoryExporter,
		}

		for _, category := range expectedCategories {
			if counts[category] == 0 {
				t.Errorf("Expected at least one metric in category %s", category)
			}
		}

		// Log distribution for manual review
		t.Logf("Metric distribution by category:")
		for category, count := range counts {
			t.Logf("  %s: %d metrics", category, count)
		}
	})
}
