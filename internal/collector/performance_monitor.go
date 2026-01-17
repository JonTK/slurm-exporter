package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// PerformanceMonitor tracks collector performance metrics and SLAs
type PerformanceMonitor struct {
	// Metrics
	collectionDuration   *prometheus.HistogramVec
	collectionErrors     *prometheus.CounterVec
	collectionSuccess    *prometheus.CounterVec
	metricsCollected     *prometheus.CounterVec
	scrapeTimeout        *prometheus.CounterVec
	lastCollectionTime   *prometheus.GaugeVec
	slaViolations        *prometheus.CounterVec
	collectorMemoryUsage *prometheus.GaugeVec
	goroutineCount       *prometheus.GaugeVec

	// SLA configuration
	slaConfig SLAConfig

	// Tracking state
	mu             sync.RWMutex
	collectorStats map[string]*CollectorPerformanceStats
	logger         *logrus.Entry
}

// SLAConfig defines performance SLAs for collectors
type SLAConfig struct {
	// Maximum collection duration before SLA violation
	MaxCollectionDuration time.Duration

	// Maximum error rate (errors per minute)
	MaxErrorRate float64

	// Minimum success rate (percentage)
	MinSuccessRate float64

	// Maximum memory usage per collector (bytes)
	MaxMemoryUsage int64

	// Maximum metrics per collection
	MaxMetricsPerCollection int
}

// CollectorPerformanceStats tracks performance statistics for a collector
type CollectorPerformanceStats struct {
	// Timing statistics
	LastDuration    time.Duration
	TotalDuration   time.Duration
	CollectionCount int64
	MinDuration     time.Duration
	MaxDuration     time.Duration

	// Error tracking
	ErrorCount        int64
	LastError         error
	LastErrorTime     time.Time
	ConsecutiveErrors int

	// Success tracking
	SuccessCount    int64
	LastSuccessTime time.Time

	// Metrics tracking
	TotalMetrics    int64
	LastMetricCount int
	MaxMetricCount  int

	// Resource usage
	LastMemoryUsage int64
	MaxMemoryUsage  int64
	GoroutineCount  int

	// SLA tracking
	SLAViolations    int64
	LastSLAViolation time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(namespace, subsystem string, slaConfig SLAConfig, logger *logrus.Entry) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		collectionDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_duration_seconds",
				Help:      "Time spent collecting metrics from SLURM",
				Buckets:   []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"collector", "status"},
		),
		collectionErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_errors_total",
				Help:      "Total number of collection errors",
			},
			[]string{"collector", "error_type"},
		),
		collectionSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_success_total",
				Help:      "Total number of successful collections",
			},
			[]string{"collector"},
		),
		metricsCollected: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "metrics_collected_total",
				Help:      "Total number of metrics collected",
			},
			[]string{"collector", "metric_type"},
		),
		scrapeTimeout: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_timeout_total",
				Help:      "Total number of collection timeouts",
			},
			[]string{"collector"},
		),
		lastCollectionTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "last_collection_timestamp_seconds",
				Help:      "Timestamp of last successful collection",
			},
			[]string{"collector"},
		),
		slaViolations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "sla_violations_total",
				Help:      "Total number of SLA violations",
			},
			[]string{"collector", "violation_type"},
		),
		collectorMemoryUsage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_memory_bytes",
				Help:      "Memory usage of collector in bytes",
			},
			[]string{"collector"},
		),
		goroutineCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "collector_goroutines",
				Help:      "Number of goroutines used by collector",
			},
			[]string{"collector"},
		),
		slaConfig:      slaConfig,
		collectorStats: make(map[string]*CollectorPerformanceStats),
		logger:         logger.WithField("component", "performance_monitor"),
	}

	return pm
}

// RegisterMetrics registers all performance metrics
func (pm *PerformanceMonitor) RegisterMetrics(reg prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		pm.collectionDuration,
		pm.collectionErrors,
		pm.collectionSuccess,
		pm.metricsCollected,
		pm.scrapeTimeout,
		pm.lastCollectionTime,
		pm.slaViolations,
		pm.collectorMemoryUsage,
		pm.goroutineCount,
	}

	for _, metric := range metrics {
		if err := reg.Register(metric); err != nil {
			return fmt.Errorf("failed to register performance metric: %w", err)
		}
	}

	return nil
}

// RecordCollection records a collection attempt
func (pm *PerformanceMonitor) RecordCollection(collector string, duration time.Duration, metricsCount int, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Get or create stats for collector
	stats, exists := pm.collectorStats[collector]
	if !exists {
		stats = &CollectorPerformanceStats{
			MinDuration: duration,
		}
		pm.collectorStats[collector] = stats
	}

	// Update timing statistics
	stats.LastDuration = duration
	stats.TotalDuration += duration
	stats.CollectionCount++
	if duration < stats.MinDuration {
		stats.MinDuration = duration
	}
	if duration > stats.MaxDuration {
		stats.MaxDuration = duration
	}

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
		pm.recordError(collector, stats, err)
	} else {
		pm.recordSuccess(collector, stats, metricsCount)
	}

	pm.collectionDuration.WithLabelValues(collector, status).Observe(duration.Seconds())

	// Check SLAs
	pm.checkSLAViolations(collector, stats, duration, err)
}

// recordError updates error statistics
func (pm *PerformanceMonitor) recordError(collector string, stats *CollectorPerformanceStats, err error) {
	stats.ErrorCount++
	stats.LastError = err
	stats.LastErrorTime = time.Now()
	stats.ConsecutiveErrors++

	// Categorize error type
	errorType := "unknown"
	switch {
	case isTimeoutError(err.Error()):
		errorType = "timeout"
		pm.scrapeTimeout.WithLabelValues(collector).Inc()
	case isConnectionError(err.Error()):
		errorType = "connection"
	case isAuthError(err.Error()):
		errorType = "authentication"
	case isAPIError(err):
		errorType = "api_error"
	}

	pm.collectionErrors.WithLabelValues(collector, errorType).Inc()

	pm.logger.WithFields(logrus.Fields{
		"collector":   collector,
		"error":       err.Error(),
		"error_type":  errorType,
		"consecutive": stats.ConsecutiveErrors,
	}).Warn("Collection error recorded")
}

// recordSuccess updates success statistics
func (pm *PerformanceMonitor) recordSuccess(collector string, stats *CollectorPerformanceStats, metricsCount int) {
	stats.SuccessCount++
	stats.LastSuccessTime = time.Now()
	stats.ConsecutiveErrors = 0
	stats.LastMetricCount = metricsCount
	stats.TotalMetrics += int64(metricsCount)

	if metricsCount > stats.MaxMetricCount {
		stats.MaxMetricCount = metricsCount
	}

	pm.collectionSuccess.WithLabelValues(collector).Inc()
	pm.lastCollectionTime.WithLabelValues(collector).Set(float64(time.Now().Unix()))
	pm.metricsCollected.WithLabelValues(collector, "total").Add(float64(metricsCount))
}

// checkSLAViolations checks for SLA violations
func (pm *PerformanceMonitor) checkSLAViolations(collector string, stats *CollectorPerformanceStats, duration time.Duration, err error) {
	violations := []string{}

	// Check collection duration SLA
	if duration > pm.slaConfig.MaxCollectionDuration {
		violations = append(violations, "duration")
		pm.slaViolations.WithLabelValues(collector, "duration").Inc()
	}

	// Check error rate SLA (errors per minute)
	if stats.CollectionCount > 0 {
		errorRate := float64(stats.ErrorCount) / (float64(stats.TotalDuration) / float64(time.Minute))
		if errorRate > pm.slaConfig.MaxErrorRate {
			violations = append(violations, "error_rate")
			pm.slaViolations.WithLabelValues(collector, "error_rate").Inc()
		}

		// Check success rate SLA
		successRate := float64(stats.SuccessCount) / float64(stats.CollectionCount) * 100
		if successRate < pm.slaConfig.MinSuccessRate {
			violations = append(violations, "success_rate")
			pm.slaViolations.WithLabelValues(collector, "success_rate").Inc()
		}
	}

	// Check metric count SLA
	if stats.LastMetricCount > pm.slaConfig.MaxMetricsPerCollection {
		violations = append(violations, "metric_count")
		pm.slaViolations.WithLabelValues(collector, "metric_count").Inc()
	}

	// Log violations
	if len(violations) > 0 {
		stats.SLAViolations++
		stats.LastSLAViolation = time.Now()

		pm.logger.WithFields(logrus.Fields{
			"collector":  collector,
			"violations": violations,
			"duration":   duration,
			"metrics":    stats.LastMetricCount,
			"error":      err != nil,
		}).Warn("SLA violations detected")
	}
}

// RecordResourceUsage records resource usage for a collector
func (pm *PerformanceMonitor) RecordResourceUsage(collector string, memoryBytes int64, goroutines int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	stats, exists := pm.collectorStats[collector]
	if !exists {
		stats = &CollectorPerformanceStats{}
		pm.collectorStats[collector] = stats
	}

	stats.LastMemoryUsage = memoryBytes
	stats.GoroutineCount = goroutines

	if memoryBytes > stats.MaxMemoryUsage {
		stats.MaxMemoryUsage = memoryBytes
	}

	// Update metrics
	pm.collectorMemoryUsage.WithLabelValues(collector).Set(float64(memoryBytes))
	pm.goroutineCount.WithLabelValues(collector).Set(float64(goroutines))

	// Check memory SLA
	if memoryBytes > pm.slaConfig.MaxMemoryUsage {
		pm.slaViolations.WithLabelValues(collector, "memory").Inc()
		stats.SLAViolations++
		stats.LastSLAViolation = time.Now()

		pm.logger.WithFields(logrus.Fields{
			"collector": collector,
			"memory":    memoryBytes,
			"limit":     pm.slaConfig.MaxMemoryUsage,
		}).Warn("Memory SLA violation")
	}
}

// GetStats returns performance statistics for all collectors
func (pm *PerformanceMonitor) GetStats() map[string]*CollectorPerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := make(map[string]*CollectorPerformanceStats)
	for name, stats := range pm.collectorStats {
		statsCopy[name] = &CollectorPerformanceStats{
			LastDuration:      stats.LastDuration,
			TotalDuration:     stats.TotalDuration,
			CollectionCount:   stats.CollectionCount,
			MinDuration:       stats.MinDuration,
			MaxDuration:       stats.MaxDuration,
			ErrorCount:        stats.ErrorCount,
			LastError:         stats.LastError,
			LastErrorTime:     stats.LastErrorTime,
			ConsecutiveErrors: stats.ConsecutiveErrors,
			SuccessCount:      stats.SuccessCount,
			LastSuccessTime:   stats.LastSuccessTime,
			TotalMetrics:      stats.TotalMetrics,
			LastMetricCount:   stats.LastMetricCount,
			MaxMetricCount:    stats.MaxMetricCount,
			LastMemoryUsage:   stats.LastMemoryUsage,
			MaxMemoryUsage:    stats.MaxMemoryUsage,
			GoroutineCount:    stats.GoroutineCount,
			SLAViolations:     stats.SLAViolations,
			LastSLAViolation:  stats.LastSLAViolation,
		}
	}

	return statsCopy
}

// GetCollectorStats returns statistics for a specific collector
func (pm *PerformanceMonitor) GetCollectorStats(collector string) (*CollectorPerformanceStats, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats, exists := pm.collectorStats[collector]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &CollectorPerformanceStats{
		LastDuration:      stats.LastDuration,
		TotalDuration:     stats.TotalDuration,
		CollectionCount:   stats.CollectionCount,
		MinDuration:       stats.MinDuration,
		MaxDuration:       stats.MaxDuration,
		ErrorCount:        stats.ErrorCount,
		LastError:         stats.LastError,
		LastErrorTime:     stats.LastErrorTime,
		ConsecutiveErrors: stats.ConsecutiveErrors,
		SuccessCount:      stats.SuccessCount,
		LastSuccessTime:   stats.LastSuccessTime,
		TotalMetrics:      stats.TotalMetrics,
		LastMetricCount:   stats.LastMetricCount,
		MaxMetricCount:    stats.MaxMetricCount,
		LastMemoryUsage:   stats.LastMemoryUsage,
		MaxMemoryUsage:    stats.MaxMemoryUsage,
		GoroutineCount:    stats.GoroutineCount,
		SLAViolations:     stats.SLAViolations,
		LastSLAViolation:  stats.LastSLAViolation,
	}, true
}

// StartPeriodicReporting starts periodic performance reporting
func (pm *PerformanceMonitor) StartPeriodicReporting(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Info("Stopping performance reporting")
			return
		case <-ticker.C:
			pm.generateReport()
		}
	}
}

// generateReport generates a performance report
func (pm *PerformanceMonitor) generateReport() {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pm.logger.Info("Performance Report")

	for collector, stats := range pm.collectorStats {
		avgDuration := time.Duration(0)
		if stats.CollectionCount > 0 {
			avgDuration = stats.TotalDuration / time.Duration(stats.CollectionCount)
		}

		successRate := float64(0)
		if stats.CollectionCount > 0 {
			successRate = float64(stats.SuccessCount) / float64(stats.CollectionCount) * 100
		}

		avgMetrics := int64(0)
		if stats.SuccessCount > 0 {
			avgMetrics = stats.TotalMetrics / stats.SuccessCount
		}

		pm.logger.WithFields(logrus.Fields{
			"collector":       collector,
			"collections":     stats.CollectionCount,
			"success_rate":    fmt.Sprintf("%.2f%%", successRate),
			"avg_duration":    avgDuration,
			"min_duration":    stats.MinDuration,
			"max_duration":    stats.MaxDuration,
			"avg_metrics":     avgMetrics,
			"max_metrics":     stats.MaxMetricCount,
			"memory_usage":    stats.LastMemoryUsage,
			"goroutines":      stats.GoroutineCount,
			"sla_violations":  stats.SLAViolations,
			"consecutive_err": stats.ConsecutiveErrors,
		}).Info("Collector performance summary")
	}
}

// Error type checking helper
func isAPIError(err error) bool {
	if err == nil {
		return false
	}
	// Default to API error for other errors
	return true
}
