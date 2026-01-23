// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// DocumentationGenerator handles generation of various documentation formats
type DocumentationGenerator struct {
	outputDir string
}

// NewDocumentationGenerator creates a new documentation generator
func NewDocumentationGenerator(outputDir string) *DocumentationGenerator {
	return &DocumentationGenerator{
		outputDir: outputDir,
	}
}

// GenerateAll generates all documentation files
func (dg *DocumentationGenerator) GenerateAll() error {
	// Ensure output directory exists
	if err := os.MkdirAll(dg.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	generators := []struct {
		filename string
		content  func() string
	}{
		{"metrics-reference.md", GenerateDocumentation},
		{"prometheus-config.yml", GeneratePrometheusConfig},
		{"grafana-dashboard.json", GenerateGrafanaDashboard},
		{"metrics-summary.md", dg.generateMetricsSummary},
		{"troubleshooting.md", dg.generateTroubleshootingGuide},
		{"api-mapping.md", dg.generateAPIMapping},
	}

	for _, gen := range generators {
		content := gen.content()
		filePath := filepath.Join(dg.outputDir, gen.filename)

		//nolint:gosec // Documentation files need to be world-readable for web servers
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", gen.filename, err)
		}
	}

	return nil
}

// generateMetricsSummary creates a high-level summary of metrics
func (dg *DocumentationGenerator) generateMetricsSummary() string {
	var doc strings.Builder

	doc.WriteString("# SLURM Exporter Metrics Summary\n\n")
	doc.WriteString("This document provides a high-level overview of metrics exposed by the SLURM Prometheus exporter.\n\n")

	// Overview statistics
	totalMetrics := GetMetricCount()
	categoryCounts := GetCategoryCount()

	doc.WriteString("## Overview\n\n")
	doc.WriteString(fmt.Sprintf("- **Total Metrics:** %d\n", totalMetrics))
	doc.WriteString("- **Categories:** ")

	var categoryList []string
	for category, count := range categoryCounts {
		categoryList = append(categoryList, fmt.Sprintf("%s (%d)", category, count))
	}
	doc.WriteString(strings.Join(categoryList, ", "))
	doc.WriteString("\n\n")

	// Stability distribution
	doc.WriteString("## Stability Levels\n\n")
	stabilityLevels := []StabilityLevel{StabilityStable, StabilityBeta, StabilityAlpha, StabilityExperimental, StabilityDeprecated}

	titleCaser := cases.Title(language.English)
	for _, level := range stabilityLevels {
		metrics := GetMetricsByStability(level)
		if len(metrics) > 0 {
			doc.WriteString(fmt.Sprintf("- **%s:** %d metrics\n", titleCaser.String(string(level)), len(metrics)))
		}
	}
	doc.WriteString("\n")

	// Quick reference by category
	doc.WriteString("## Quick Reference by Category\n\n")

	categories := []MetricCategory{CategoryCluster, CategoryNode, CategoryJob, CategoryUser, CategoryAccount, CategoryPartition, CategoryPerformance, CategoryExporter}

	for _, category := range categories {
		metrics := GetMetricsByCategory(category)
		if len(metrics) == 0 {
			continue
		}

		doc.WriteString(fmt.Sprintf("### %s Metrics\n\n", titleCaser.String(string(category))))

		for _, metric := range metrics {
			doc.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", metric.Name, metric.Type, metric.Help))
		}
		doc.WriteString("\n")
	}

	// Collection endpoints
	doc.WriteString("## SLURM API Endpoints\n\n")
	doc.WriteString("The exporter collects data from these SLURM REST API endpoints:\n\n")

	endpoints := make(map[string][]string)
	for _, metadata := range MetricsRegistry {
		if metadata.SlurmEndpoint != "" && metadata.SlurmEndpoint != "internal" {
			endpoints[metadata.SlurmEndpoint] = append(endpoints[metadata.SlurmEndpoint], metadata.Name)
		}
	}

	for endpoint, metrics := range endpoints {
		doc.WriteString(fmt.Sprintf("- **%s** (%d metrics)\n", endpoint, len(metrics)))
	}
	doc.WriteString("\n")

	return doc.String()
}

// generateTroubleshootingGuide creates a troubleshooting guide
func (dg *DocumentationGenerator) generateTroubleshootingGuide() string {
	var doc strings.Builder

	doc.WriteString("# SLURM Exporter Troubleshooting Guide\n\n")
	doc.WriteString("This guide helps diagnose and resolve common issues with the SLURM Prometheus exporter.\n\n")

	doc.WriteString("## Common Issues\n\n")

	// Organize troubleshooting by category
	doc.WriteString("### Connection Issues\n\n")
	doc.WriteString("**Symptoms:** Missing metrics, connection timeouts, authentication errors\n\n")
	doc.WriteString("**Diagnostics:**\n")
	doc.WriteString("- Check `slurm_exporter_collector_up` metric\n")
	doc.WriteString("- Review `slurm_exporter_collection_errors_total` for error patterns\n")
	doc.WriteString("- Examine `slurm_exporter_api_call_duration_seconds` for slow responses\n\n")

	doc.WriteString("**Solutions:**\n")
	doc.WriteString("- Verify SLURM REST API is running and accessible\n")
	doc.WriteString("- Check authentication credentials and permissions\n")
	doc.WriteString("- Verify network connectivity and firewall rules\n")
	doc.WriteString("- Review SLURM daemon logs for errors\n\n")

	doc.WriteString("### Performance Issues\n\n")
	doc.WriteString("**Symptoms:** High collection times, memory usage, timeout errors\n\n")
	doc.WriteString("**Diagnostics:**\n")
	doc.WriteString("- Monitor `slurm_exporter_collection_duration_seconds` histogram\n")
	doc.WriteString("- Check cardinality of high-cardinality metrics (jobs, nodes)\n")
	doc.WriteString("- Review system resource usage (CPU, memory)\n\n")

	doc.WriteString("**Solutions:**\n")
	doc.WriteString("- Increase collection intervals for expensive collectors\n")
	doc.WriteString("- Implement metric filtering to reduce cardinality\n")
	doc.WriteString("- Configure appropriate timeouts and retry policies\n")
	doc.WriteString("- Consider running multiple exporter instances for load distribution\n\n")

	doc.WriteString("### Data Quality Issues\n\n")
	doc.WriteString("**Symptoms:** Inconsistent values, missing labels, stale data\n\n")
	doc.WriteString("**Diagnostics:**\n")
	doc.WriteString("- Compare metrics with SLURM command outputs (`sinfo`, `squeue`, `sacct`)\n")
	doc.WriteString("- Check `slurm_exporter_last_collection_timestamp` for staleness\n")
	doc.WriteString("- Review API response formats for schema changes\n\n")

	doc.WriteString("**Solutions:**\n")
	doc.WriteString("- Verify SLURM version compatibility\n")
	doc.WriteString("- Update exporter to latest version\n")
	doc.WriteString("- Review and update metric mapping configurations\n")
	doc.WriteString("- Check for SLURM configuration changes\n\n")

	// Metric-specific troubleshooting
	doc.WriteString("## Metric-Specific Troubleshooting\n\n")

	for _, metadata := range MetricsRegistry {
		if metadata.Troubleshooting != "" {
			doc.WriteString(fmt.Sprintf("### %s\n\n", metadata.Name))
			doc.WriteString(fmt.Sprintf("%s\n\n", metadata.Troubleshooting))
		}
	}

	doc.WriteString("## Diagnostic Queries\n\n")
	doc.WriteString("Useful Prometheus queries for diagnosing exporter issues:\n\n")

	queries := []struct {
		name  string
		query string
		desc  string
	}{
		{
			"Exporter Health",
			"up{job=\"slurm-exporter\"}",
			"Shows if the exporter is up and responding",
		},
		{
			"Collection Success Rate",
			"rate(slurm_exporter_collection_success_total[5m]) / rate(slurm_exporter_collection_duration_seconds_count[5m])",
			"Success rate of metric collection",
		},
		{
			"Average Collection Time",
			"rate(slurm_exporter_collection_duration_seconds_sum[5m]) / rate(slurm_exporter_collection_duration_seconds_count[5m])",
			"Average time spent collecting metrics",
		},
		{
			"High Cardinality Metrics",
			"count by (__name__) ({__name__=~\"slurm_.*\"})",
			"Number of series per metric (cardinality check)",
		},
		{
			"API Error Rate",
			"rate(slurm_exporter_api_call_errors_total[5m])",
			"Rate of SLURM API call errors",
		},
	}

	for _, query := range queries {
		doc.WriteString(fmt.Sprintf("**%s:**\n", query.name))
		doc.WriteString(fmt.Sprintf("```promql\n%s\n```\n", query.query))
		doc.WriteString(fmt.Sprintf("%s\n\n", query.desc))
	}

	return doc.String()
}

// generateAPIMapping creates API endpoint mapping documentation
func (dg *DocumentationGenerator) generateAPIMapping() string {
	var doc strings.Builder

	doc.WriteString("# SLURM API Endpoint Mapping\n\n")
	doc.WriteString("This document describes how SLURM REST API endpoints map to Prometheus metrics.\n\n")

	// Group metrics by endpoint
	endpointMetrics := make(map[string][]MetricMetadata)
	for _, metadata := range MetricsRegistry {
		if metadata.SlurmEndpoint != "" && metadata.SlurmEndpoint != "internal" {
			endpointMetrics[metadata.SlurmEndpoint] = append(endpointMetrics[metadata.SlurmEndpoint], metadata)
		}
	}

	doc.WriteString("## Endpoint Overview\n\n")
	for endpoint, metrics := range endpointMetrics {
		doc.WriteString(fmt.Sprintf("- **%s**: %d metrics\n", endpoint, len(metrics)))
	}
	doc.WriteString("\n")

	doc.WriteString("## Detailed Mapping\n\n")

	for endpoint, metrics := range endpointMetrics {
		doc.WriteString(fmt.Sprintf("### %s\n\n", endpoint))
		doc.WriteString("**Metrics derived from this endpoint:**\n\n")

		for _, metric := range metrics {
			doc.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", metric.Name, metric.Type, metric.Help))
			if metric.CalculationMethod != "" {
				doc.WriteString(fmt.Sprintf("  - *Calculation:* %s\n", metric.CalculationMethod))
			}
		}
		doc.WriteString("\n")

		doc.WriteString("**Sample API Response Structure:**\n")
		doc.WriteString("```json\n")
		doc.WriteString("{\n")
		doc.WriteString("  \"meta\": {\n")
		doc.WriteString("    \"plugin\": {\n")
		doc.WriteString("      \"type\": \"openapi/v0.0.39\",\n")
		doc.WriteString("      \"name\": \"Slurm OpenAPI\"\n")
		doc.WriteString("    }\n")
		doc.WriteString("  },\n")
		doc.WriteString("  \"data\": {\n")
		doc.WriteString("    // Endpoint-specific data structure\n")
		doc.WriteString("  }\n")
		doc.WriteString("}\n")
		doc.WriteString("```\n\n")
	}

	// Internal metrics section
	internalMetrics := []MetricMetadata{}
	for _, metadata := range MetricsRegistry {
		if metadata.SlurmEndpoint == "internal" || metadata.SlurmEndpoint == "" {
			internalMetrics = append(internalMetrics, metadata)
		}
	}

	if len(internalMetrics) > 0 {
		doc.WriteString("## Internal Metrics\n\n")
		doc.WriteString("These metrics are generated internally by the exporter and do not correspond to specific SLURM API endpoints:\n\n")

		for _, metric := range internalMetrics {
			doc.WriteString(fmt.Sprintf("- **%s**: %s\n", metric.Name, metric.Help))
		}
		doc.WriteString("\n")
	}

	doc.WriteString("## API Version Compatibility\n\n")
	doc.WriteString("The exporter supports multiple SLURM REST API versions:\n\n")
	doc.WriteString("- **v0.0.39**: Full support (recommended)\n")
	doc.WriteString("- **v0.0.40**: Full support\n")
	doc.WriteString("- **v0.0.41**: Full support\n")
	doc.WriteString("- **v0.0.42**: Full support\n")
	doc.WriteString("- **v0.0.43**: Full support\n\n")

	doc.WriteString("**Version Detection:**\n")
	doc.WriteString("The exporter automatically detects the SLURM API version and adapts its requests accordingly. ")
	doc.WriteString("Version information is included in the `slurm_cluster_info` metric labels.\n\n")

	return doc.String()
}

// GenerateHelmChart creates a Helm chart for deploying the exporter
func (dg *DocumentationGenerator) GenerateHelmChart() error {
	chartDir := filepath.Join(dg.outputDir, "helm-chart", "slurm-exporter")

	// Create chart directory structure
	dirs := []string{
		chartDir,
		filepath.Join(chartDir, "templates"),
		filepath.Join(chartDir, "templates", "tests"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Chart.yaml
	chartYaml := `apiVersion: v2
name: slurm-exporter
description: A Helm chart for SLURM Prometheus Exporter
type: application
version: 0.1.0
appVersion: "1.0.0"
keywords:
  - slurm
  - prometheus
  - monitoring
  - hpc
home: https://github.com/jontk/slurm-exporter
sources:
  - https://github.com/jontk/slurm-exporter
maintainers:
  - name: SLURM Exporter Team
    email: support@example.com
`

	// values.yaml
	valuesYaml := `# Default values for slurm-exporter
replicaCount: 1

image:
  repository: slurm-exporter
  pullPolicy: IfNotPresent
  tag: "latest"

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  create: true
  annotations: {}
  name: ""

podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "9100"
  prometheus.io/path: "/metrics"

podSecurityContext: {}
securityContext: {}

service:
  type: ClusterIP
  port: 80
  targetPort: 9100

ingress:
  enabled: false
  className: ""
  annotations: {}
  hosts:
    - host: slurm-exporter.local
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

autoscaling:
  enabled: false
  minReplicas: 1
  maxReplicas: 100
  targetCPUUtilizationPercentage: 80

nodeSelector: {}
tolerations: []
affinity: {}

config:
  slurmApi:
    baseUrl: "http://slurm-rest:6820"
    timeout: 30s
    version: "v0.0.39"
  collectors:
    cluster:
      enabled: true
      interval: 30s
    node:
      enabled: true
      interval: 30s
    job:
      enabled: true
      interval: 15s
    user:
      enabled: true
      interval: 60s
    partition:
      enabled: true
      interval: 30s
    performance:
      enabled: true
      interval: 60s
  metrics:
    port: 9100
    path: /metrics

serviceMonitor:
  enabled: false
  labels: {}
  annotations: {}
  interval: 30s
  scrapeTimeout: 10s
`

	files := map[string]string{
		filepath.Join(chartDir, "Chart.yaml"):  chartYaml,
		filepath.Join(chartDir, "values.yaml"): valuesYaml,
	}

	for path, content := range files {
		//nolint:gosec // Helm chart files need to be world-readable for deployment tools
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	return nil
}

// GenerateMetricsHTML creates an HTML metrics browser
func (dg *DocumentationGenerator) GenerateMetricsHTML() string {
	var html strings.Builder

	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SLURM Exporter Metrics Browser</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .metric { border: 1px solid #ddd; margin: 10px 0; padding: 15px; border-radius: 5px; }
        .metric-name { font-weight: bold; color: #333; }
        .metric-type { color: #666; font-style: italic; }
        .metric-help { margin: 10px 0; }
        .metric-labels { color: #555; }
        .category { background-color: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .stability { padding: 2px 6px; border-radius: 3px; font-size: 12px; }
        .stable { background-color: #d4edda; color: #155724; }
        .beta { background-color: #fff3cd; color: #856404; }
        .alpha { background-color: #f8d7da; color: #721c24; }
        .search { width: 100%; padding: 10px; margin-bottom: 20px; font-size: 16px; }
        .toc { background-color: #f8f9fa; padding: 15px; margin-bottom: 20px; }
        .example { background-color: #f8f8f8; padding: 10px; font-family: monospace; margin: 10px 0; }
    </style>
    <script>
        function filterMetrics() {
            const searchTerm = document.getElementById('search').value.toLowerCase();
            const metrics = document.querySelectorAll('.metric');

            metrics.forEach(metric => {
                const text = metric.textContent.toLowerCase();
                if (text.includes(searchTerm)) {
                    metric.style.display = 'block';
                } else {
                    metric.style.display = 'none';
                }
            });
        }
    </script>
</head>
<body>
    <h1>SLURM Exporter Metrics Browser</h1>
    <p>Interactive browser for all metrics exposed by the SLURM Prometheus exporter.</p>

    <input type="text" id="search" class="search" placeholder="Search metrics..." onkeyup="filterMetrics()">

    <div class="toc">
        <h3>Categories</h3>
        <ul>`)

	titleCaser := cases.Title(language.English)
	categories := []MetricCategory{CategoryCluster, CategoryNode, CategoryJob, CategoryUser, CategoryAccount, CategoryPartition, CategoryPerformance, CategoryExporter}
	for _, category := range categories {
		count := len(GetMetricsByCategory(category))
		if count > 0 {
			html.WriteString(fmt.Sprintf(`<li><a href="#%s">%s (%d metrics)</a></li>`, strings.ToLower(string(category)), titleCaser.String(string(category)), count))
		}
	}

	html.WriteString(`</ul>
    </div>`)

	for _, category := range categories {
		metrics := GetMetricsByCategory(category)
		if len(metrics) == 0 {
			continue
		}

		html.WriteString(fmt.Sprintf(`
    <div class="category" id="%s">
        <h2>%s Metrics</h2>`, strings.ToLower(string(category)), titleCaser.String(string(category))))

		for _, metric := range metrics {
			stabilityClass := string(metric.StabilityLevel)
			html.WriteString(fmt.Sprintf(`
        <div class="metric">
            <div class="metric-name">%s <span class="stability %s">%s</span></div>
            <div class="metric-type">Type: %s</div>
            <div class="metric-help">%s</div>
            <div class="metric-description">%s</div>`,
				metric.Name, stabilityClass, strings.ToUpper(string(metric.StabilityLevel)),
				metric.Type, metric.Help, metric.Description))

			if len(metric.Labels) > 0 {
				html.WriteString(`<div class="metric-labels"><strong>Labels:</strong> `)
				html.WriteString(strings.Join(metric.Labels, ", "))
				html.WriteString(`</div>`)
			}

			if metric.ExampleValue != "" {
				html.WriteString(`<div class="example">Example: `)
				html.WriteString(metric.Name)
				if len(metric.ExampleLabels) > 0 {
					html.WriteString(`{`)
					var labelPairs []string
					for key, value := range metric.ExampleLabels {
						labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, key, value))
					}
					html.WriteString(strings.Join(labelPairs, ","))
					html.WriteString(`}`)
				}
				html.WriteString(fmt.Sprintf(` %s</div>`, metric.ExampleValue))
			}

			html.WriteString(`</div>`)
		}

		html.WriteString(`</div>`)
	}

	html.WriteString(`
</body>
</html>`)

	return html.String()
}
