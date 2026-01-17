package collector

import (
	"strings"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricType represents the type of Prometheus metric
type MetricType int

const (
	MetricTypeCounter MetricType = iota
	MetricTypeGauge
	MetricTypeHistogram
	MetricTypeSummary
	MetricTypeInfo
)

// MetricInfo holds information about a metric for filtering
type MetricInfo struct {
	Name        string
	Type        MetricType
	IsInfo      bool
	IsTiming    bool
	IsResource  bool
	Description string
}

// MetricFilter provides metric filtering capabilities
type MetricFilter struct {
	config config.MetricFilterConfig
}

// NewMetricFilter creates a new metric filter
func NewMetricFilter(config config.MetricFilterConfig) *MetricFilter {
	return &MetricFilter{
		config: config,
	}
}

// ShouldCollectMetric determines if a metric should be collected based on filters
func (mf *MetricFilter) ShouldCollectMetric(info MetricInfo) bool {
	// If enable_all is false and no specific includes/excludes, default to true
	if !mf.config.EnableAll && len(mf.config.IncludeMetrics) == 0 && len(mf.config.ExcludeMetrics) == 0 {
		return mf.applyTypeFilters(info)
	}

	// Check include/exclude lists first
	if len(mf.config.IncludeMetrics) > 0 {
		included := false
		for _, pattern := range mf.config.IncludeMetrics {
			if mf.matchesPattern(info.Name, pattern) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check exclude list
	for _, pattern := range mf.config.ExcludeMetrics {
		if mf.matchesPattern(info.Name, pattern) {
			return false
		}
	}

	// Apply type-specific filters
	return mf.applyTypeFilters(info)
}

// applyTypeFilters applies type-specific filtering rules
func (mf *MetricFilter) applyTypeFilters(info MetricInfo) bool {
	// Type-specific filtering
	if mf.config.OnlyInfo && !info.IsInfo {
		return false
	}
	if mf.config.OnlyCounters && info.Type != MetricTypeCounter {
		return false
	}
	if mf.config.OnlyGauges && info.Type != MetricTypeGauge {
		return false
	}
	if mf.config.OnlyHistograms && info.Type != MetricTypeHistogram {
		return false
	}

	// Skip filters
	if mf.config.SkipHistograms && info.Type == MetricTypeHistogram {
		return false
	}
	if mf.config.SkipTimingMetrics && info.IsTiming {
		return false
	}
	if mf.config.SkipResourceMetrics && info.IsResource {
		return false
	}

	return true
}

// matchesPattern checks if a metric name matches a pattern (supports wildcards)
func (mf *MetricFilter) matchesPattern(name, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// Convert wildcard pattern to regex-like matching
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			return strings.HasPrefix(name, prefix)
		}
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(name, suffix)
		}
		// For patterns with * in the middle, use contains
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(name, parts[0]) && strings.HasSuffix(name, parts[1])
		}
	}

	// Exact match
	return name == pattern
}

// GetMetricType determines the metric type from a Prometheus descriptor
func GetMetricType(desc *prometheus.Desc) MetricType {
	// Since we can't directly get the type from desc, we'll use naming conventions
	metricName := desc.String()

	if strings.Contains(metricName, "_total") {
		return MetricTypeCounter
	}
	if strings.Contains(metricName, "_bucket") || strings.Contains(metricName, "_histogram") {
		return MetricTypeHistogram
	}
	if strings.Contains(metricName, "_info") {
		return MetricTypeInfo
	}
	// Default to gauge
	return MetricTypeGauge
}

// CreateMetricInfo creates MetricInfo from metric name and characteristics
func CreateMetricInfo(name string, metricType MetricType, desc string) MetricInfo {
	info := MetricInfo{
		Name:        name,
		Type:        metricType,
		Description: desc,
	}

	// Determine if it's an info metric
	info.IsInfo = strings.Contains(name, "_info") || metricType == MetricTypeInfo

	// Determine if it's a timing metric
	info.IsTiming = strings.Contains(name, "_time") ||
		strings.Contains(name, "_duration") ||
		strings.Contains(name, "_latency") ||
		strings.Contains(name, "_seconds")

	// Determine if it's a resource metric
	info.IsResource = strings.Contains(name, "_cpu") ||
		strings.Contains(name, "_memory") ||
		strings.Contains(name, "_disk") ||
		strings.Contains(name, "_bytes") ||
		strings.Contains(name, "_usage")

	return info
}

// DefaultMetricFilterConfig returns a default metric filter configuration
func DefaultMetricFilterConfig() config.MetricFilterConfig {
	return config.MetricFilterConfig{
		EnableAll:           true,
		IncludeMetrics:      []string{},
		ExcludeMetrics:      []string{},
		OnlyInfo:            false,
		OnlyCounters:        false,
		OnlyGauges:          false,
		OnlyHistograms:      false,
		SkipHistograms:      false,
		SkipTimingMetrics:   false,
		SkipResourceMetrics: false,
	}
}
