// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"fmt"
	// Commented out as only used in commented-out field
	// "sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// PerformanceMetrics provides comprehensive performance monitoring for the exporter
type PerformanceMetrics struct {
	// Collection duration by collector with detailed buckets
	CollectionDuration *prometheus.HistogramVec

	// API call duration by endpoint and method
	APICallDuration *prometheus.HistogramVec

	// Collection errors by collector and error type
	CollectionErrors *prometheus.CounterVec

	// API errors by endpoint and status code
	APIErrors *prometheus.CounterVec

	// Metric cardinality by metric name
	MetricCardinality *prometheus.GaugeVec

	// Cache performance metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec
	CacheSize   *prometheus.GaugeVec

	// Memory usage tracking
	MemoryUsage *prometheus.GaugeVec

	// CPU usage tracking
	CPUUsage *prometheus.GaugeVec

	// Collection rate by collector
	CollectionRate *prometheus.CounterVec

	// Concurrent collections gauge
	ConcurrentCollections *prometheus.GaugeVec

	// Queue depth for pending collections
	QueueDepth *prometheus.GaugeVec

	// Adaptive interval tracking
	AdaptiveInterval *prometheus.GaugeVec

	// Circuit breaker state
	CircuitBreakerState *prometheus.GaugeVec

	// Thread pool utilization
	ThreadPoolUtilization *prometheus.GaugeVec

	// TODO: Unused field - preserved for future thread safety
	// mu sync.RWMutex
}

// NewPerformanceMetrics creates a new performance metrics instance
func NewPerformanceMetrics(namespace string) *PerformanceMetrics {
	return &PerformanceMetrics{
		CollectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "collection_duration_seconds",
				Help:      "Duration of collections by collector with detailed timing",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
			},
			[]string{"collector", "phase"},
		),

		APICallDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "api_duration_seconds",
				Help:      "Duration of SLURM API calls",
				Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
			},
			[]string{"endpoint", "method", "status"},
		),

		CollectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "collection_errors_total",
				Help:      "Total number of collection errors by type",
			},
			[]string{"collector", "error_type"},
		),

		APIErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "api_errors_total",
				Help:      "Total number of API errors by endpoint and status",
			},
			[]string{"endpoint", "status_code", "error_type"},
		),

		MetricCardinality: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "metric_cardinality",
				Help:      "Current cardinality of metrics by name",
			},
			[]string{"metric_name", "collector"},
		),

		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache_type", "collector"},
		),

		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache_type", "collector"},
		),

		CacheSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "cache_size_bytes",
				Help:      "Current cache size in bytes",
			},
			[]string{"cache_type"},
		),

		MemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "memory_usage_bytes",
				Help:      "Current memory usage in bytes",
			},
			[]string{"type"}, // heap, stack, gc
		),

		CPUUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "cpu_usage_percent",
				Help:      "Current CPU usage percentage",
			},
			[]string{"type"}, // user, system, total
		),

		CollectionRate: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "collections_per_second",
				Help:      "Rate of collections per second",
			},
			[]string{"collector"},
		),

		ConcurrentCollections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "concurrent_collections",
				Help:      "Number of concurrent collections in progress",
			},
			[]string{"collector"},
		),

		QueueDepth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "queue_depth",
				Help:      "Number of pending collections in queue",
			},
			[]string{"collector"},
		),

		AdaptiveInterval: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "adaptive_interval_seconds",
				Help:      "Current adaptive collection interval in seconds",
			},
			[]string{"collector"},
		),

		CircuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "circuit_breaker_state",
				Help:      "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"name"},
		),

		ThreadPoolUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "exporter",
				Name:      "thread_pool_utilization_percent",
				Help:      "Thread pool utilization percentage",
			},
			[]string{"pool_name"},
		),
	}
}

// Register registers all performance metrics with the prometheus registry
func (pm *PerformanceMetrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		pm.CollectionDuration,
		pm.APICallDuration,
		pm.CollectionErrors,
		pm.APIErrors,
		pm.MetricCardinality,
		pm.CacheHits,
		pm.CacheMisses,
		pm.CacheSize,
		pm.MemoryUsage,
		pm.CPUUsage,
		pm.CollectionRate,
		pm.ConcurrentCollections,
		pm.QueueDepth,
		pm.AdaptiveInterval,
		pm.CircuitBreakerState,
		pm.ThreadPoolUtilization,
	}

	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			return fmt.Errorf("failed to register performance metric: %w", err)
		}
	}

	return nil
}

// Timer provides a convenient way to time operations
type Timer struct {
	start    time.Time
	observer prometheus.Observer
}

// NewTimer creates a new timer that will observe the given histogram
func NewTimer(observer prometheus.Observer) *Timer {
	return &Timer{
		start:    time.Now(),
		observer: observer,
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.observer.Observe(duration.Seconds())
	return duration
}

// CollectionTimer provides specialized timing for collection operations
type CollectionTimer struct {
	collector string
	phase     string
	metrics   *PerformanceMetrics
	timer     *Timer
}

// NewCollectionTimer creates a timer for collection operations
func (pm *PerformanceMetrics) NewCollectionTimer(collector, phase string) *CollectionTimer {
	timer := NewTimer(pm.CollectionDuration.WithLabelValues(collector, phase))

	// Increment concurrent collections
	pm.ConcurrentCollections.WithLabelValues(collector).Inc()

	return &CollectionTimer{
		collector: collector,
		phase:     phase,
		metrics:   pm,
		timer:     timer,
	}
}

// Stop stops the collection timer and updates metrics
func (ct *CollectionTimer) Stop() time.Duration {
	duration := ct.timer.Stop()

	// Decrement concurrent collections
	ct.metrics.ConcurrentCollections.WithLabelValues(ct.collector).Dec()

	// Update collection rate
	ct.metrics.CollectionRate.WithLabelValues(ct.collector).Inc()

	return duration
}

// APITimer provides specialized timing for API operations
type APITimer struct {
	endpoint  string
	method    string
	metrics   *PerformanceMetrics
	startTime time.Time
}

// NewAPITimer creates a timer for API operations
func (pm *PerformanceMetrics) NewAPITimer(endpoint, method string) *APITimer {
	return &APITimer{
		endpoint:  endpoint,
		method:    method,
		metrics:   pm,
		startTime: time.Now(), // Capture start time immediately
	}
}

// StopWithStatus stops the API timer with the given status
func (at *APITimer) StopWithStatus(status string) time.Duration {
	duration := time.Since(at.startTime)
	// Observe the duration with the correct labels
	at.metrics.APICallDuration.WithLabelValues(at.endpoint, at.method, status).Observe(duration.Seconds())
	return duration
}

// RecordError records an API error
func (pm *PerformanceMetrics) RecordAPIError(endpoint, statusCode, errorType string) {
	pm.APIErrors.WithLabelValues(endpoint, statusCode, errorType).Inc()
}

// RecordCollectionError records a collection error
func (pm *PerformanceMetrics) RecordCollectionError(collector, errorType string) {
	pm.CollectionErrors.WithLabelValues(collector, errorType).Inc()
}

// UpdateCardinality updates the cardinality metric for a given metric
func (pm *PerformanceMetrics) UpdateCardinality(metricName, collector string, cardinality float64) {
	pm.MetricCardinality.WithLabelValues(metricName, collector).Set(cardinality)
}

// RecordCacheHit records a cache hit
func (pm *PerformanceMetrics) RecordCacheHit(cacheType, collector string) {
	pm.CacheHits.WithLabelValues(cacheType, collector).Inc()
}

// RecordCacheMiss records a cache miss
func (pm *PerformanceMetrics) RecordCacheMiss(cacheType, collector string) {
	pm.CacheMisses.WithLabelValues(cacheType, collector).Inc()
}

// UpdateCacheSize updates the cache size metric
func (pm *PerformanceMetrics) UpdateCacheSize(cacheType string, sizeBytes float64) {
	pm.CacheSize.WithLabelValues(cacheType).Set(sizeBytes)
}

// UpdateMemoryUsage updates memory usage metrics
func (pm *PerformanceMetrics) UpdateMemoryUsage(usageType string, bytes float64) {
	pm.MemoryUsage.WithLabelValues(usageType).Set(bytes)
}

// UpdateCPUUsage updates CPU usage metrics
func (pm *PerformanceMetrics) UpdateCPUUsage(usageType string, percent float64) {
	pm.CPUUsage.WithLabelValues(usageType).Set(percent)
}

// UpdateQueueDepth updates the queue depth metric
func (pm *PerformanceMetrics) UpdateQueueDepth(collector string, depth float64) {
	pm.QueueDepth.WithLabelValues(collector).Set(depth)
}

// UpdateAdaptiveInterval updates the adaptive interval metric
func (pm *PerformanceMetrics) UpdateAdaptiveInterval(collector string, intervalSeconds float64) {
	pm.AdaptiveInterval.WithLabelValues(collector).Set(intervalSeconds)
}

// UpdateCircuitBreakerState updates the circuit breaker state
func (pm *PerformanceMetrics) UpdateCircuitBreakerState(name string, state float64) {
	pm.CircuitBreakerState.WithLabelValues(name).Set(state)
}

// UpdateThreadPoolUtilization updates thread pool utilization
func (pm *PerformanceMetrics) UpdateThreadPoolUtilization(poolName string, percent float64) {
	pm.ThreadPoolUtilization.WithLabelValues(poolName).Set(percent)
}
