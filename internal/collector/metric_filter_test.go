// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"testing"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricFilter(t *testing.T) {
	t.Parallel()
	t.Run("NewMetricFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			EnableAll: true,
		}
		filter := NewMetricFilter(cfg)

		if filter == nil {
			t.Error("Expected non-nil filter")
			return
		}
		if filter.config.EnableAll != true {
			t.Error("Filter config not set correctly")
		}
	})
}

func TestShouldCollectMetric(t *testing.T) {
	t.Parallel()
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultMetricFilterConfig()
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Counter metric",
				metric: MetricInfo{
					Name: "test_counter_total",
					Type: MetricTypeCounter,
				},
				expected: true,
			},
			{
				name: "Gauge metric",
				metric: MetricInfo{
					Name: "test_gauge",
					Type: MetricTypeGauge,
				},
				expected: true,
			},
			{
				name: "Histogram metric",
				metric: MetricInfo{
					Name: "test_histogram",
					Type: MetricTypeHistogram,
				},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("IncludeFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			IncludeMetrics: []string{"slurm_node_*", "slurm_job_*"},
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   string
			expected bool
		}{
			{"Include node metrics", "slurm_node_cpus", true},
			{"Include job metrics", "slurm_job_count", true},
			{"Exclude other metrics", "other_metric", false},
			{"Exclude partition metrics", "slurm_partition_nodes", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(MetricInfo{Name: tc.metric})
				if result != tc.expected {
					t.Errorf("For metric '%s': expected %v, got %v", tc.metric, tc.expected, result)
				}
			})
		}
	})

	t.Run("ExcludeFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			ExcludeMetrics: []string{"*_info", "*_debug"},
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   string
			expected bool
		}{
			{"Exclude info metrics", "slurm_node_info", false},
			{"Exclude debug metrics", "slurm_debug", false},
			{"Include other metrics", "slurm_node_cpus", true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(MetricInfo{Name: tc.metric})
				if result != tc.expected {
					t.Errorf("For metric '%s': expected %v, got %v", tc.metric, tc.expected, result)
				}
			})
		}
	})

	t.Run("OnlyInfoFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			OnlyInfo: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Info metric",
				metric: MetricInfo{
					Name:   "test_info",
					IsInfo: true,
				},
				expected: true,
			},
			{
				name: "Counter metric",
				metric: MetricInfo{
					Name:   "test_counter",
					IsInfo: false,
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("OnlyCountersFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			OnlyCounters: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Counter metric",
				metric: MetricInfo{
					Name: "test_counter_total",
					Type: MetricTypeCounter,
				},
				expected: true,
			},
			{
				name: "Gauge metric",
				metric: MetricInfo{
					Name: "test_gauge",
					Type: MetricTypeGauge,
				},
				expected: false,
			},
			{
				name: "Histogram metric",
				metric: MetricInfo{
					Name: "test_histogram",
					Type: MetricTypeHistogram,
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("OnlyGaugesFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			OnlyGauges: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Gauge metric",
				metric: MetricInfo{
					Name: "test_gauge",
					Type: MetricTypeGauge,
				},
				expected: true,
			},
			{
				name: "Counter metric",
				metric: MetricInfo{
					Name: "test_counter_total",
					Type: MetricTypeCounter,
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("SkipHistogramsFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			SkipHistograms: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Histogram metric",
				metric: MetricInfo{
					Name: "test_histogram",
					Type: MetricTypeHistogram,
				},
				expected: false,
			},
			{
				name: "Gauge metric",
				metric: MetricInfo{
					Name: "test_gauge",
					Type: MetricTypeGauge,
				},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("SkipTimingMetricsFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			SkipTimingMetrics: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Timing metric",
				metric: MetricInfo{
					Name:     "test_duration_seconds",
					IsTiming: true,
				},
				expected: false,
			},
			{
				name: "Non-timing metric",
				metric: MetricInfo{
					Name:     "test_count",
					IsTiming: false,
				},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})

	t.Run("SkipResourceMetricsFilter", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			SkipResourceMetrics: true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "CPU metric",
				metric: MetricInfo{
					Name:       "test_cpu",
					IsResource: true,
				},
				expected: false,
			},
			{
				name: "Memory metric",
				metric: MetricInfo{
					Name:       "test_memory",
					IsResource: true,
				},
				expected: false,
			},
			{
				name: "Non-resource metric",
				metric: MetricInfo{
					Name:       "test_count",
					IsResource: false,
				},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})
}

func TestMatchesPattern(t *testing.T) {
	t.Parallel()
	cfg := DefaultMetricFilterConfig()
	filter := NewMetricFilter(cfg)

	testCases := []struct {
		name     string
		metric   string
		pattern  string
		expected bool
	}{
		{"Wildcard all", "any_metric", "*", true},
		{"Wildcard suffix", "slurm_node_cpus", "slurm_*", true},
		{"Wildcard prefix", "node_cpus", "*_cpus", true},
		{"Wildcard both sides", "slurm_node_cpus", "slurm_*_cpus", true},
		{"Exact match", "slurm_node_cpus", "slurm_node_cpus", true},
		{"No match", "other_metric", "slurm_*", false},
		{"Suffix wildcard mismatch", "slurm_node_memory", "*_cpus", false},
		{"Prefix wildcard mismatch", "node_cpus", "slurm_*", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filter.matchesPattern(tc.metric, tc.pattern)
			if result != tc.expected {
				t.Errorf("Pattern '%s' on metric '%s': expected %v, got %v",
					tc.pattern, tc.metric, tc.expected, result)
			}
		})
	}
}

func TestGetMetricType(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		metric   string
		expected MetricType
	}{
		{"Counter metric", "slurm_jobs_total", MetricTypeCounter},
		{"Histogram bucket", "slurm_duration_bucket", MetricTypeHistogram},
		{"Histogram metric", "slurm_request_duration_histogram", MetricTypeHistogram},
		{"Info metric", "slurm_node_info", MetricTypeInfo},
		{"Gauge metric", "slurm_node_cpus", MetricTypeGauge},
		{"Default gauge", "slurm_up", MetricTypeGauge},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			desc := prometheus.NewDesc(tc.metric, "test metric", nil, nil)
			result := GetMetricType(desc)
			if result != tc.expected {
				t.Errorf("Expected type %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestCreateMetricInfo(t *testing.T) {
	t.Parallel()
	t.Run("Counter metric", func(t *testing.T) {
		info := CreateMetricInfo("slurm_jobs_total", MetricTypeCounter, "Total jobs")

		if info.Name != "slurm_jobs_total" {
			t.Errorf("Expected name 'slurm_jobs_total', got '%s'", info.Name)
		}
		if info.Type != MetricTypeCounter {
			t.Errorf("Expected type Counter, got %v", info.Type)
		}
		if info.IsInfo != false {
			t.Error("Counter should not be info metric")
		}
		if info.IsTiming != false {
			t.Error("Counter should not be timing metric")
		}
		if info.IsResource != false {
			t.Error("Counter should not be resource metric")
		}
	})

	t.Run("Info metric", func(t *testing.T) {
		info := CreateMetricInfo("slurm_node_info", MetricTypeInfo, "Node information")

		if info.IsInfo != true {
			t.Error("Info metric should have IsInfo true")
		}
	})

	t.Run("Timing metric", func(t *testing.T) {
		testCases := []struct {
			name     string
			metric   string
			isTiming bool
		}{
			{"Duration metric", "slurm_duration_seconds", true},
			{"Time metric", "slurm_request_time", true},
			{"Latency metric", "slurm_latency", true},
			{"Non-timing metric", "slurm_count", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				info := CreateMetricInfo(tc.metric, MetricTypeGauge, "Test metric")
				if info.IsTiming != tc.isTiming {
					t.Errorf("Metric '%s': expected IsTiming %v, got %v",
						tc.metric, tc.isTiming, info.IsTiming)
				}
			})
		}
	})

	t.Run("Resource metric", func(t *testing.T) {
		testCases := []struct {
			name       string
			metric     string
			isResource bool
		}{
			{"CPU metric", "slurm_node_cpus", true},
			{"Memory metric", "slurm_node_memory", true},
			{"Disk metric", "slurm_disk_usage", true},
			{"Bytes metric", "slurm_bytes", true},
			{"Usage metric", "slurm_usage", true},
			{"Non-resource metric", "slurm_count", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				info := CreateMetricInfo(tc.metric, MetricTypeGauge, "Test metric")
				if info.IsResource != tc.isResource {
					t.Errorf("Metric '%s': expected IsResource %v, got %v",
						tc.metric, tc.isResource, info.IsResource)
				}
			})
		}
	})
}

func TestDefaultMetricFilterConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultMetricFilterConfig()

	if !cfg.EnableAll {
		t.Error("EnableAll should be true by default")
	}
	if cfg.OnlyInfo {
		t.Error("OnlyInfo should be false by default")
	}
	if cfg.OnlyCounters {
		t.Error("OnlyCounters should be false by default")
	}
	if cfg.OnlyGauges {
		t.Error("OnlyGauges should be false by default")
	}
	if cfg.OnlyHistograms {
		t.Error("OnlyHistograms should be false by default")
	}
	if cfg.SkipHistograms {
		t.Error("SkipHistograms should be false by default")
	}
	if cfg.SkipTimingMetrics {
		t.Error("SkipTimingMetrics should be false by default")
	}
	if cfg.SkipResourceMetrics {
		t.Error("SkipResourceMetrics should be false by default")
	}
}

func TestMetricFilterCombined(t *testing.T) {
	t.Parallel()
	t.Run("Include and Exclude combined", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			IncludeMetrics: []string{"slurm_*"},
			ExcludeMetrics: []string{"*_debug", "*_info"},
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   string
			expected bool
		}{
			{"Include node CPU metric", "slurm_node_cpus", true},
			{"Exclude debug metric", "slurm_debug", false},
			{"Exclude info metric", "slurm_node_info", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(MetricInfo{Name: tc.metric})
				if result != tc.expected {
					t.Errorf("For metric '%s': expected %v, got %v", tc.metric, tc.expected, result)
				}
			})
		}
	})

	t.Run("Type and pattern filters combined", func(t *testing.T) {
		cfg := config.MetricFilterConfig{
			IncludeMetrics: []string{"slurm_*"},
			OnlyCounters:   true,
		}
		filter := NewMetricFilter(cfg)

		testCases := []struct {
			name     string
			metric   MetricInfo
			expected bool
		}{
			{
				name: "Include counter",
				metric: MetricInfo{
					Name: "slurm_jobs_total",
					Type: MetricTypeCounter,
				},
				expected: true,
			},
			{
				name: "Exclude gauge",
				metric: MetricInfo{
					Name: "slurm_node_cpus",
					Type: MetricTypeGauge,
				},
				expected: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := filter.ShouldCollectMetric(tc.metric)
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
			})
		}
	})
}
