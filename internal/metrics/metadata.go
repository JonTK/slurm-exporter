// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// MetricMetadata contains comprehensive metadata for SLURM exporter metrics
type MetricMetadata struct {
	Name              string
	Type              MetricType
	Category          MetricCategory
	Help              string
	Description       string
	Labels            []string
	Unit              string
	StabilityLevel    StabilityLevel
	ExampleValue      string
	ExampleLabels     map[string]string
	RelatedMetrics    []string
	SlurmEndpoint     string
	CalculationMethod string
	UseCase           string
	Troubleshooting   string
}

// MetricType represents the Prometheus metric type
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// MetricCategory represents the functional category of the metric
type MetricCategory string

const (
	CategoryCluster     MetricCategory = "cluster"
	CategoryNode        MetricCategory = "node"
	CategoryJob         MetricCategory = "job"
	CategoryUser        MetricCategory = "user"
	CategoryAccount     MetricCategory = "account"
	CategoryPartition   MetricCategory = "partition"
	CategoryPerformance MetricCategory = "performance"
	CategoryExporter    MetricCategory = "exporter"
)

// StabilityLevel represents the stability level of the metric
type StabilityLevel string

const (
	StabilityStable       StabilityLevel = "stable"
	StabilityBeta         StabilityLevel = "beta"
	StabilityAlpha        StabilityLevel = "alpha"
	StabilityDeprecated   StabilityLevel = "deprecated"
	StabilityExperimental StabilityLevel = "experimental"
)

// MetricsRegistry contains metadata for all metrics
var MetricsRegistry = map[string]MetricMetadata{
	"slurm_cluster_info": {
		Name:           "slurm_cluster_info",
		Type:           MetricTypeGauge,
		Category:       CategoryCluster,
		Help:           "Information about the SLURM cluster",
		Description:    "Provides basic information about the SLURM cluster including version, configuration, and status. This metric always has a value of 1 and is used primarily for its labels which contain the actual information.",
		Labels:         []string{"cluster_name", "slurm_version", "config_path", "status"},
		Unit:           "",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1",
		ExampleLabels: map[string]string{
			"cluster_name":  "production",
			"slurm_version": "22.05.2",
			"config_path":   "/etc/slurm/slurm.conf",
			"status":        "active",
		},
		RelatedMetrics:    []string{"slurm_cluster_capacity_cpus", "slurm_cluster_allocated_cpus"},
		SlurmEndpoint:     "/slurm/v0.0.39/diag",
		CalculationMethod: "Direct mapping from SLURM cluster status API",
		UseCase:           "Monitoring cluster health, tracking SLURM version updates, identifying configuration changes",
		Troubleshooting:   "If this metric is missing, check SLURM daemon connectivity. If labels are incorrect, verify SLURM API response format.",
	},
	"slurm_cluster_capacity_cpus": {
		Name:           "slurm_cluster_capacity_cpus",
		Type:           MetricTypeGauge,
		Category:       CategoryCluster,
		Help:           "Total CPU capacity of the cluster",
		Description:    "Total number of CPU cores available in the cluster across all nodes. This includes both allocated and idle CPUs.",
		Labels:         []string{"cluster_name"},
		Unit:           "cores",
		StabilityLevel: StabilityStable,
		ExampleValue:   "2400",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
		},
		RelatedMetrics:    []string{"slurm_cluster_allocated_cpus"},
		SlurmEndpoint:     "/slurm/v0.0.39/nodes",
		CalculationMethod: "Sum of CPUs across all nodes in the cluster",
		UseCase:           "Capacity planning, resource allocation monitoring, cluster sizing decisions",
		Troubleshooting:   "Sudden drops may indicate node failures. Compare with node-level metrics to identify specific issues.",
	},
	"slurm_cluster_allocated_cpus": {
		Name:           "slurm_cluster_allocated_cpus",
		Type:           MetricTypeGauge,
		Category:       CategoryCluster,
		Help:           "Number of allocated CPUs in the cluster",
		Description:    "Number of CPU cores currently allocated to running jobs across the entire cluster.",
		Labels:         []string{"cluster_name"},
		Unit:           "cores",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1680",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
		},
		RelatedMetrics:    []string{"slurm_cluster_capacity_cpus"},
		SlurmEndpoint:     "/slurm/v0.0.39/jobs",
		CalculationMethod: "Sum of CPU allocations across all running jobs",
		UseCase:           "Real-time utilization monitoring, load balancing, resource efficiency analysis",
		Troubleshooting:   "High values near capacity indicate resource contention. Use job-level metrics to identify resource-heavy jobs.",
	},
	"slurm_node_info": {
		Name:           "slurm_node_info",
		Type:           MetricTypeGauge,
		Category:       CategoryNode,
		Help:           "Information about cluster nodes",
		Description:    "Provides detailed information about individual compute nodes including hardware specifications, features, and configuration.",
		Labels:         []string{"cluster_name", "node_name", "partition", "arch", "os", "features"},
		Unit:           "",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"node_name":    "node001",
			"partition":    "compute",
			"arch":         "x86_64",
			"os":           "Linux",
			"features":     "avx2,fma",
		},
		RelatedMetrics:    []string{"slurm_node_state"},
		SlurmEndpoint:     "/slurm/v0.0.39/nodes",
		CalculationMethod: "Direct mapping from SLURM node information API",
		UseCase:           "Inventory management, hardware tracking, capacity planning by node type",
		Troubleshooting:   "Missing nodes indicate communication issues. Inconsistent features suggest configuration drift.",
	},
	"slurm_node_state": {
		Name:           "slurm_node_state",
		Type:           MetricTypeGauge,
		Category:       CategoryNode,
		Help:           "Current state of cluster nodes",
		Description:    "Binary indicator (0 or 1) for each possible node state. A value of 1 indicates the node is in that state.",
		Labels:         []string{"cluster_name", "node_name", "state"},
		Unit:           "",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"node_name":    "node001",
			"state":        "idle",
		},
		RelatedMetrics:    []string{"slurm_node_info"},
		SlurmEndpoint:     "/slurm/v0.0.39/nodes",
		CalculationMethod: "Binary encoding of node state from SLURM API (idle=1, allocated=1, down=1, etc.)",
		UseCase:           "Node health monitoring, failure detection, maintenance scheduling",
		Troubleshooting:   "Persistent down/drain states require administrator intervention. Use with node allocation metrics to understand usage patterns.",
	},
	"slurm_job_info": {
		Name:           "slurm_job_info",
		Type:           MetricTypeGauge,
		Category:       CategoryJob,
		Help:           "Information about SLURM jobs",
		Description:    "Comprehensive information about individual jobs including resource requests, ownership, and scheduling details.",
		Labels:         []string{"cluster_name", "job_id", "user", "account", "partition", "job_name", "qos"},
		Unit:           "",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"job_id":       "12345",
			"user":         "alice",
			"account":      "ml_team",
			"partition":    "gpu",
			"job_name":     "training_job",
			"qos":          "normal",
		},
		RelatedMetrics:    []string{"slurm_job_states"},
		SlurmEndpoint:     "/slurm/v0.0.39/jobs",
		CalculationMethod: "Direct mapping from SLURM job information API",
		UseCase:           "Job tracking, resource accountability, user activity monitoring",
		Troubleshooting:   "Missing jobs may indicate API connectivity issues. Use job state metrics to understand job lifecycle.",
	},
	"slurm_job_states": {
		Name:           "slurm_job_states",
		Type:           MetricTypeGauge,
		Category:       CategoryJob,
		Help:           "Count of jobs in each state",
		Description:    "Number of jobs currently in each possible job state (pending, running, completed, etc.). Provides overview of job queue status.",
		Labels:         []string{"cluster_name", "state", "partition"},
		Unit:           "jobs",
		StabilityLevel: StabilityStable,
		ExampleValue:   "25",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"state":        "running",
			"partition":    "compute",
		},
		RelatedMetrics:    []string{"slurm_job_info"},
		SlurmEndpoint:     "/slurm/v0.0.39/jobs",
		CalculationMethod: "Count of jobs grouped by state and partition",
		UseCase:           "Queue monitoring, workload analysis, scheduling efficiency assessment",
		Troubleshooting:   "High pending counts indicate resource constraints. Compare with capacity metrics to identify bottlenecks.",
	},
	"slurm_user_job_count": {
		Name:           "slurm_user_job_count",
		Type:           MetricTypeGauge,
		Category:       CategoryUser,
		Help:           "Number of jobs per user",
		Description:    "Total number of jobs (all states) currently associated with each user. Useful for tracking user activity and resource usage patterns.",
		Labels:         []string{"cluster_name", "user", "account"},
		Unit:           "jobs",
		StabilityLevel: StabilityStable,
		ExampleValue:   "8",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"user":         "alice",
			"account":      "ml_team",
		},
		RelatedMetrics:    []string{},
		SlurmEndpoint:     "/slurm/v0.0.39/jobs",
		CalculationMethod: "Count of jobs grouped by user and account",
		UseCase:           "User activity monitoring, fair-share enforcement, resource allocation tracking",
		Troubleshooting:   "Unusually high counts for individual users may indicate job submission issues or resource hogging.",
	},
	"slurm_partition_info": {
		Name:           "slurm_partition_info",
		Type:           MetricTypeGauge,
		Category:       CategoryPartition,
		Help:           "Information about SLURM partitions",
		Description:    "Configuration and status information for SLURM partitions including access controls, limits, and scheduling policies.",
		Labels:         []string{"cluster_name", "partition", "state", "allow_groups", "allow_users"},
		Unit:           "",
		StabilityLevel: StabilityStable,
		ExampleValue:   "1",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"partition":    "gpu",
			"state":        "UP",
			"allow_groups": "gpu_users",
			"allow_users":  "all",
		},
		RelatedMetrics:    []string{},
		SlurmEndpoint:     "/slurm/v0.0.39/partitions",
		CalculationMethod: "Direct mapping from SLURM partition configuration API",
		UseCase:           "Partition management, access control monitoring, resource organization",
		Troubleshooting:   "DOWN state indicates partition issues. Check allow_groups/allow_users for access problems.",
	},
	"slurm_system_throughput": {
		Name:           "slurm_system_throughput",
		Type:           MetricTypeGauge,
		Category:       CategoryPerformance,
		Help:           "System throughput metrics",
		Description:    "Various throughput measurements including jobs completed per hour, CPU hours delivered, and submission rates. Indicates cluster productivity.",
		Labels:         []string{"cluster_name", "metric_type", "time_window", "partition"},
		Unit:           "varies",
		StabilityLevel: StabilityBeta,
		ExampleValue:   "45.5",
		ExampleLabels: map[string]string{
			"cluster_name": "production",
			"metric_type":  "jobs_completed",
			"time_window":  "1h",
			"partition":    "compute",
		},
		RelatedMetrics:    []string{},
		SlurmEndpoint:     "/slurm/v0.0.39/jobs",
		CalculationMethod: "Time-based aggregation of job completion, submission, and resource delivery metrics",
		UseCase:           "Performance monitoring, capacity planning, SLA tracking",
		Troubleshooting:   "Declining throughput may indicate scheduling issues, resource constraints, or job submission problems.",
	},
	"slurm_exporter_collection_duration_seconds": {
		Name:           "slurm_exporter_collection_duration_seconds",
		Type:           MetricTypeHistogram,
		Category:       CategoryExporter,
		Help:           "Time spent collecting metrics from SLURM",
		Description:    "Histogram of time taken to collect metrics from SLURM APIs. Helps identify performance issues and API response times.",
		Labels:         []string{"collector"},
		Unit:           "seconds",
		StabilityLevel: StabilityStable,
		ExampleValue:   "2.3",
		ExampleLabels: map[string]string{
			"collector": "job",
		},
		RelatedMetrics:    []string{},
		SlurmEndpoint:     "internal",
		CalculationMethod: "Measurement of actual collection time with histogram buckets",
		UseCase:           "Exporter performance monitoring, API response time analysis, troubleshooting collection issues",
		Troubleshooting:   "High values indicate slow SLURM API responses. Check network connectivity and SLURM daemon health.",
	},
}

// GetMetricMetadata returns metadata for a specific metric
func GetMetricMetadata(metricName string) (MetricMetadata, bool) {
	metadata, exists := MetricsRegistry[metricName]
	return metadata, exists
}

// GetMetricsByCategory returns all metrics in a specific category
func GetMetricsByCategory(category MetricCategory) []MetricMetadata {
	var metrics []MetricMetadata
	for _, metadata := range MetricsRegistry {
		if metadata.Category == category {
			metrics = append(metrics, metadata)
		}
	}

	// Sort by name for consistent output
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics
}

// GetMetricsByStability returns all metrics with a specific stability level
func GetMetricsByStability(stability StabilityLevel) []MetricMetadata {
	var metrics []MetricMetadata
	for _, metadata := range MetricsRegistry {
		if metadata.StabilityLevel == stability {
			metrics = append(metrics, metadata)
		}
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics
}

// GenerateDocumentation generates comprehensive documentation for all metrics
func GenerateDocumentation() string {
	var doc strings.Builder
	titleCaser := cases.Title(language.English)

	doc.WriteString("# SLURM Exporter Metrics Reference\n\n")
	doc.WriteString("This document provides comprehensive documentation for all metrics exposed by the SLURM Prometheus exporter.\n\n")

	// Table of contents
	doc.WriteString("## Table of Contents\n\n")
	categories := []MetricCategory{
		CategoryCluster, CategoryNode, CategoryJob, CategoryUser,
		CategoryAccount, CategoryPartition, CategoryPerformance, CategoryExporter,
	}

	for _, category := range categories {
		doc.WriteString(fmt.Sprintf("- [%s Metrics](#%s-metrics)\n", titleCaser.String(string(category)), strings.ToLower(string(category))))
	}
	doc.WriteString("\n")

	// Generate documentation by category
	for _, category := range categories {
		metrics := GetMetricsByCategory(category)
		if len(metrics) == 0 {
			continue
		}

		doc.WriteString(fmt.Sprintf("## %s Metrics\n\n", titleCaser.String(string(category))))

		for _, metric := range metrics {
			doc.WriteString(generateMetricDocumentation(metric))
			doc.WriteString("\n---\n\n")
		}
	}

	// Stability levels section
	doc.WriteString("## Metric Stability Levels\n\n")
	doc.WriteString("Metrics are classified by stability level:\n\n")
	doc.WriteString("- **Stable**: Production-ready metrics with guaranteed backward compatibility\n")
	doc.WriteString("- **Beta**: Well-tested metrics that may have minor changes\n")
	doc.WriteString("- **Alpha**: Early-stage metrics that may change significantly\n")
	doc.WriteString("- **Experimental**: Proof-of-concept metrics that may be removed\n")
	doc.WriteString("- **Deprecated**: Metrics scheduled for removal in future versions\n\n")

	for _, stability := range []StabilityLevel{StabilityStable, StabilityBeta, StabilityAlpha, StabilityExperimental, StabilityDeprecated} {
		metrics := GetMetricsByStability(stability)
		if len(metrics) > 0 {
			doc.WriteString(fmt.Sprintf("### %s Metrics (%d)\n\n", titleCaser.String(string(stability)), len(metrics)))
			for _, metric := range metrics {
				doc.WriteString(fmt.Sprintf("- `%s`\n", metric.Name))
			}
			doc.WriteString("\n")
		}
	}

	return doc.String()
}

// generateMetricDocumentation generates documentation for a single metric
func generateMetricDocumentation(metadata MetricMetadata) string {
	var doc strings.Builder

	doc.WriteString(fmt.Sprintf("### %s\n\n", metadata.Name))
	doc.WriteString(fmt.Sprintf("**Type:** %s  \n", metadata.Type))
	doc.WriteString(fmt.Sprintf("**Category:** %s  \n", metadata.Category))
	doc.WriteString(fmt.Sprintf("**Stability:** %s  \n", metadata.StabilityLevel))
	if metadata.Unit != "" {
		doc.WriteString(fmt.Sprintf("**Unit:** %s  \n", metadata.Unit))
	}
	doc.WriteString("\n")

	doc.WriteString(fmt.Sprintf("**Description:** %s\n\n", metadata.Description))

	if len(metadata.Labels) > 0 {
		doc.WriteString("**Labels:**\n")
		for _, label := range metadata.Labels {
			doc.WriteString(fmt.Sprintf("- `%s`\n", label))
		}
		doc.WriteString("\n")
	}

	if metadata.ExampleValue != "" {
		doc.WriteString("**Example:**\n```\n")
		doc.WriteString(metadata.Name)
		if len(metadata.ExampleLabels) > 0 {
			doc.WriteString("{")
			var labelPairs []string
			for key, value := range metadata.ExampleLabels {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, key, value))
			}
			sort.Strings(labelPairs)
			doc.WriteString(strings.Join(labelPairs, ","))
			doc.WriteString("}")
		}
		doc.WriteString(fmt.Sprintf(" %s\n", metadata.ExampleValue))
		doc.WriteString("```\n\n")
	}

	if len(metadata.RelatedMetrics) > 0 {
		doc.WriteString("**Related Metrics:**\n")
		for _, related := range metadata.RelatedMetrics {
			doc.WriteString(fmt.Sprintf("- `%s`\n", related))
		}
		doc.WriteString("\n")
	}

	if metadata.SlurmEndpoint != "" && metadata.SlurmEndpoint != "internal" {
		doc.WriteString(fmt.Sprintf("**SLURM API Endpoint:** `%s`\n\n", metadata.SlurmEndpoint))
	}

	if metadata.CalculationMethod != "" {
		doc.WriteString(fmt.Sprintf("**Calculation:** %s\n\n", metadata.CalculationMethod))
	}

	if metadata.UseCase != "" {
		doc.WriteString(fmt.Sprintf("**Use Cases:** %s\n\n", metadata.UseCase))
	}

	if metadata.Troubleshooting != "" {
		doc.WriteString(fmt.Sprintf("**Troubleshooting:** %s\n\n", metadata.Troubleshooting))
	}

	return doc.String()
}

// GeneratePrometheusConfig generates a sample Prometheus configuration
func GeneratePrometheusConfig() string {
	var config strings.Builder

	config.WriteString("# Sample Prometheus configuration for SLURM Exporter\n")
	config.WriteString("global:\n")
	config.WriteString("  scrape_interval: 30s\n")
	config.WriteString("  evaluation_interval: 30s\n\n")

	config.WriteString("scrape_configs:\n")
	config.WriteString("  - job_name: 'slurm-exporter'\n")
	config.WriteString("    static_configs:\n")
	config.WriteString("      - targets: ['localhost:9100']\n")
	config.WriteString("    scrape_interval: 30s\n")
	config.WriteString("    scrape_timeout: 10s\n")
	config.WriteString("    metrics_path: /metrics\n\n")

	config.WriteString("# Sample recording rules for common aggregations\n")
	config.WriteString("rule_files:\n")
	config.WriteString("  - \"slurm_rules.yml\"\n\n")

	return config.String()
}

// GenerateGrafanaDashboard generates a sample Grafana dashboard configuration
func GenerateGrafanaDashboard() string {
	return `{
  "dashboard": {
    "id": null,
    "title": "SLURM Cluster Overview",
    "tags": ["slurm", "hpc"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Cluster CPU Utilization",
        "type": "stat",
        "targets": [
          {
            "expr": "slurm_cluster_allocated_cpus / slurm_cluster_capacity_cpus * 100",
            "legendFormat": "CPU Utilization %"
          }
        ]
      },
      {
        "title": "Job States",
        "type": "piechart",
        "targets": [
          {
            "expr": "slurm_job_states",
            "legendFormat": "{{state}}"
          }
        ]
      },
      {
        "title": "Node States",
        "type": "bargraph",
        "targets": [
          {
            "expr": "sum by (state) (slurm_node_state)",
            "legendFormat": "{{state}}"
          }
        ]
      },
      {
        "title": "Queue Depth",
        "type": "graph",
        "targets": [
          {
            "expr": "slurm_queue_depth",
            "legendFormat": "{{partition}}"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}`
}

// ValidateMetricMetadata validates the completeness of metric metadata
func ValidateMetricMetadata() []string {
	var issues []string

	for name, metadata := range MetricsRegistry {
		if metadata.Name != name {
			issues = append(issues, fmt.Sprintf("Metric %s: name mismatch in metadata", name))
		}

		if metadata.Help == "" {
			issues = append(issues, fmt.Sprintf("Metric %s: missing help text", name))
		}

		if metadata.Description == "" {
			issues = append(issues, fmt.Sprintf("Metric %s: missing description", name))
		}

		if metadata.StabilityLevel == "" {
			issues = append(issues, fmt.Sprintf("Metric %s: missing stability level", name))
		}

		if metadata.Category == "" {
			issues = append(issues, fmt.Sprintf("Metric %s: missing category", name))
		}

		if metadata.Type == "" {
			issues = append(issues, fmt.Sprintf("Metric %s: missing type", name))
		}
	}

	return issues
}

// GetMetricCount returns the total number of registered metrics
func GetMetricCount() int {
	return len(MetricsRegistry)
}

// GetCategoryCount returns the number of metrics in each category
func GetCategoryCount() map[MetricCategory]int {
	counts := make(map[MetricCategory]int)
	for _, metadata := range MetricsRegistry {
		counts[metadata.Category]++
	}
	return counts
}
