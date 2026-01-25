// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// Collector defines the interface for metric collectors
type Collector interface {
	// Name returns the collector name
	Name() string

	// Describe sends the super-set of all possible descriptors of metrics
	// collected by this Collector to the provided channel and returns once
	// the last descriptor has been sent.
	Describe(ch chan<- *prometheus.Desc)

	// Collect is called by the Prometheus registry when collecting metrics.
	// The implementation sends each collected metric via the provided channel
	// and returns once the last metric has been sent.
	Collect(ctx context.Context, ch chan<- prometheus.Metric) error

	// IsEnabled returns whether this collector is enabled
	IsEnabled() bool

	// SetEnabled enables or disables the collector
	SetEnabled(enabled bool)
}

// FilterableCollector extends the base Collector interface with filtering capabilities
type FilterableCollector interface {
	Collector

	// SetMetricFilter sets the metric filter for this collector
	SetMetricFilter(filter *MetricFilter)

	// GetMetricFilter returns the current metric filter
	GetMetricFilter() *MetricFilter

	// UpdateFilterConfig updates the filter configuration
	UpdateFilterConfig(config config.FilterConfig)
}

// CardinalityAwareCollector extends collectors with cardinality management
type CardinalityAwareCollector interface {
	Collector

	// SetCardinalityManager sets the cardinality manager for this collector
	SetCardinalityManager(cm *metrics.CardinalityManager)
}

// CustomLabelsCollector extends collectors with custom label support
type CustomLabelsCollector interface {
	Collector

	// SetCustomLabels sets custom labels for this collector
	SetCustomLabels(labels map[string]string)
}

// CollectorMetrics provides common metrics for all collectors
type CollectorMetrics struct {
	// Duration of the last collection
	Duration *prometheus.HistogramVec

	// Total number of collections
	Total *prometheus.CounterVec

	// Number of failed collections
	Errors *prometheus.CounterVec

	// Current status (1 = up, 0 = down)
	Up *prometheus.GaugeVec
}

// NewCollectorMetrics creates common metrics for collectors
func NewCollectorMetrics(namespace, subsystem string) *CollectorMetrics {
	constLabels := prometheus.Labels{
		"subsystem": subsystem,
	}

	return &CollectorMetrics{
		Duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "collection_duration_seconds",
				Help:        "Duration of collections by the exporter",
				ConstLabels: constLabels,
				Buckets:     prometheus.DefBuckets,
			},
			[]string{"collector"},
		),
		Total: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "collections_total",
				Help:        "Total number of collections by the exporter",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),
		Errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "collection_errors_total",
				Help:        "Total number of collection errors by the exporter",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),
		Up: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Subsystem:   subsystem,
				Name:        "collector_up",
				Help:        "Whether the collector is up (1) or down (0)",
				ConstLabels: constLabels,
			},
			[]string{"collector"},
		),
	}
}

// Register registers all collector metrics with prometheus
func (m *CollectorMetrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.Duration,
		m.Total,
		m.Errors,
		m.Up,
	}

	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	return nil
}

// CollectorOptions provides options for creating collectors
type CollectorOptions struct {
	// Namespace for metrics
	Namespace string

	// Subsystem for metrics
	Subsystem string

	// Timeout for collection operations
	Timeout time.Duration

	// Labels to add to all metrics
	ConstLabels prometheus.Labels

	// Logger for the collector
	Logger interface{}
}

// Note: CollectionError moved to errors.go with enhanced functionality

// CollectorState represents the current state of a collector
type CollectorState struct {
	// Name of the collector
	Name string

	// Whether the collector is enabled
	Enabled bool

	// Last collection time
	LastCollection time.Time

	// Last collection duration
	LastDuration time.Duration

	// Last error (if any)
	LastError error

	// Number of consecutive errors
	ConsecutiveErrors int

	// Total collections
	TotalCollections int64

	// Total errors
	TotalErrors int64
}
