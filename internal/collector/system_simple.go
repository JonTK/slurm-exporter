// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	systemCollectorSubsystem = "system"
)

// Compile-time interface compliance checks
var (
	_ Collector             = (*SystemSimpleCollector)(nil)
	_ CustomLabelsCollector = (*SystemSimpleCollector)(nil)
)

// SystemSimpleCollector collects system-related metrics
type SystemSimpleCollector struct {
	logger  *logrus.Entry
	client  slurm.SlurmClient
	enabled bool

	// Custom labels
	customLabels map[string]string

	// SLURM daemon health metrics
	slurmDaemonUp   *prometheus.Desc
	slurmAPILatency *prometheus.Desc
	slurmAPIErrors  *prometheus.Desc

	// Database connectivity
	slurmDBConnections *prometheus.Desc
	slurmDBLatency     *prometheus.Desc

	// System resource metrics
	systemLoadAvg     *prometheus.Desc
	systemMemoryUsage *prometheus.Desc
	systemDiskUsage   *prometheus.Desc

	// SLURM configuration metrics
	configLastModified *prometheus.Desc
	activeControllers  *prometheus.Desc

	// Performance counters
	lastCollectionTime time.Time
	apiCallCount       *prometheus.CounterVec
	collectionDuration *prometheus.HistogramVec
}

// NewSystemSimpleCollector creates a new System collector
func NewSystemSimpleCollector(client slurm.SlurmClient, logger *logrus.Entry) *SystemSimpleCollector {
	c := &SystemSimpleCollector{
		client:             client,
		logger:             logger.WithField("collector", "system"),
		enabled:            true,
		customLabels:       make(map[string]string),
		lastCollectionTime: time.Now(),
	}

	// Initialize metrics
	c.initializeMetrics()

	return c
}

// initializeMetrics creates metric descriptors with custom labels as constant labels
func (c *SystemSimpleCollector) initializeMetrics() {
	// Convert custom labels to prometheus.Labels for constant labels
	constLabels := prometheus.Labels{}
	for k, v := range c.customLabels {
		constLabels[k] = v
	}
	c.slurmDaemonUp = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "slurm_daemon_up"),
		"Whether SLURM daemon is responding (1=up, 0=down)",
		[]string{"cluster", "daemon_type"},
		constLabels,
	)

	c.slurmAPILatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "slurm_api_latency_seconds"),
		"SLURM API call latency in seconds",
		[]string{"cluster", "endpoint"},
		constLabels,
	)

	c.slurmAPIErrors = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "slurm_api_errors_total"),
		"Total number of SLURM API errors",
		[]string{"cluster", "endpoint", "error_type"},
		constLabels,
	)

	c.slurmDBConnections = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "slurm_db_connections"),
		"Number of active SLURM database connections",
		[]string{"cluster"},
		constLabels,
	)

	c.slurmDBLatency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "slurm_db_latency_seconds"),
		"SLURM database query latency in seconds",
		[]string{"cluster", "query_type"},
		constLabels,
	)

	c.systemLoadAvg = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "load_average"),
		"System load average",
		[]string{"period"},
		constLabels,
	)

	c.systemMemoryUsage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "memory_usage_bytes"),
		"System memory usage in bytes",
		[]string{"type"},
		constLabels,
	)

	c.systemDiskUsage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "disk_usage_bytes"),
		"System disk usage in bytes",
		[]string{"mountpoint", "type"},
		constLabels,
	)

	c.configLastModified = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "config_last_modified_timestamp"),
		"Timestamp when SLURM configuration was last modified",
		[]string{"cluster", "config_type"},
		constLabels,
	)

	c.activeControllers = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, systemCollectorSubsystem, "active_controllers"),
		"Number of active SLURM controllers",
		[]string{"cluster"},
		constLabels,
	)

	// Performance counters
	c.apiCallCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: systemCollectorSubsystem,
			Name:      "api_calls_total",
			Help:      "Total number of SLURM API calls made",
		},
		[]string{"endpoint", "status"},
	)

	c.collectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: systemCollectorSubsystem,
			Name:      "collection_duration_seconds",
			Help:      "Time spent collecting metrics from SLURM",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"collector"},
	)
}

// SetCustomLabels sets custom labels for this collector
func (c *SystemSimpleCollector) SetCustomLabels(labels map[string]string) {
	c.customLabels = make(map[string]string)
	for k, v := range labels {
		c.customLabels[k] = v
	}
	// Rebuild metrics with new constant labels
	c.initializeMetrics()
}

// Name returns the collector name
func (c *SystemSimpleCollector) Name() string {
	return "system"
}

// IsEnabled returns whether this collector is enabled
func (c *SystemSimpleCollector) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables the collector
func (c *SystemSimpleCollector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// Describe implements prometheus.Collector
func (c *SystemSimpleCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.slurmDaemonUp
	ch <- c.slurmAPILatency
	ch <- c.slurmAPIErrors
	ch <- c.slurmDBConnections
	ch <- c.slurmDBLatency
	ch <- c.systemLoadAvg
	ch <- c.systemMemoryUsage
	ch <- c.systemDiskUsage
	ch <- c.configLastModified
	ch <- c.activeControllers
	c.apiCallCount.Describe(ch)
	c.collectionDuration.Describe(ch)
}

// Collect implements the Collector interface
func (c *SystemSimpleCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	if !c.enabled {
		return nil
	}
	return c.collect(ch)
}

// collect gathers metrics from SLURM
func (c *SystemSimpleCollector) collect(ch chan<- prometheus.Metric) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		c.collectionDuration.WithLabelValues("system").Observe(duration)
	}()

	ctx := context.Background()

	// Get cluster info
	clusterName := "default"
	infoManager := c.client.Info()
	if infoManager != nil {
		if info, err := infoManager.Get(ctx); err == nil && info != nil {
			clusterName = info.ClusterName
		}
	}

	c.logger.Info("Collected system metrics")

	// Check SLURM daemon health
	c.checkSlurmHealth(ch, ctx, clusterName)

	// Collect system metrics
	c.collectSystemMetrics(ch)

	// Collect SLURM-specific system info
	c.collectSlurmSystemInfo(ch, ctx, clusterName)

	// Collect performance counters
	c.apiCallCount.Collect(ch)
	c.collectionDuration.Collect(ch)

	c.lastCollectionTime = time.Now()
	return nil
}

// checkSlurmHealth checks SLURM daemon health
func (c *SystemSimpleCollector) checkSlurmHealth(ch chan<- prometheus.Metric, ctx context.Context, clusterName string) {
	// Test connectivity to different SLURM services
	services := map[string]func() error{
		"slurmctld": func() error {
			if infoManager := c.client.Info(); infoManager != nil {
				err := infoManager.Ping(ctx)
				return err
			}
			return fmt.Errorf("info manager not available")
		},
		"slurmdbd": func() error {
			if accountsManager := c.client.Accounts(); accountsManager != nil {
				_, err := accountsManager.List(ctx, nil)
				return err
			}
			return fmt.Errorf("accounts manager not available")
		},
	}

	for serviceName, healthCheck := range services {
		start := time.Now()
		err := healthCheck()
		latency := time.Since(start).Seconds()

		// Health status
		status := 1.0
		if err != nil {
			status = 0.0
			c.apiCallCount.WithLabelValues(serviceName, "error").Inc()
		} else {
			c.apiCallCount.WithLabelValues(serviceName, "success").Inc()
		}

		ch <- prometheus.MustNewConstMetric(
			c.slurmDaemonUp,
			prometheus.GaugeValue,
			status,
			clusterName, serviceName,
		)

		// API latency
		ch <- prometheus.MustNewConstMetric(
			c.slurmAPILatency,
			prometheus.GaugeValue,
			latency,
			clusterName, serviceName,
		)
	}
}

// collectSystemMetrics collects system-level metrics
func (c *SystemSimpleCollector) collectSystemMetrics(ch chan<- prometheus.Metric) {
	// Get Go runtime memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// System memory metrics
	ch <- prometheus.MustNewConstMetric(
		c.systemMemoryUsage,
		prometheus.GaugeValue,
		float64(memStats.Alloc),
		"allocated",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemMemoryUsage,
		prometheus.GaugeValue,
		float64(memStats.Sys),
		"system",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemMemoryUsage,
		prometheus.GaugeValue,
		float64(memStats.HeapAlloc),
		"heap",
	)

	// Simulate load average (in real implementation, read from /proc/loadavg)
	ch <- prometheus.MustNewConstMetric(
		c.systemLoadAvg,
		prometheus.GaugeValue,
		0.5, // Simulated 1-minute load average
		"1m",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemLoadAvg,
		prometheus.GaugeValue,
		0.4, // Simulated 5-minute load average
		"5m",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemLoadAvg,
		prometheus.GaugeValue,
		0.3, // Simulated 15-minute load average
		"15m",
	)

	// Simulate disk usage (in real implementation, use syscall.Statfs_t)
	ch <- prometheus.MustNewConstMetric(
		c.systemDiskUsage,
		prometheus.GaugeValue,
		1024*1024*1024*100, // 100GB used
		"/", "used",
	)

	ch <- prometheus.MustNewConstMetric(
		c.systemDiskUsage,
		prometheus.GaugeValue,
		1024*1024*1024*400, // 400GB total
		"/", "total",
	)
}

// collectSlurmSystemInfo collects SLURM-specific system information
func (c *SystemSimpleCollector) collectSlurmSystemInfo(ch chan<- prometheus.Metric, ctx context.Context, clusterName string) {
	// Number of active controllers (simulated)
	ch <- prometheus.MustNewConstMetric(
		c.activeControllers,
		prometheus.GaugeValue,
		1, // Single controller
		clusterName,
	)

	// Configuration last modified (simulated)
	configTime := time.Now().Add(-24 * time.Hour).Unix() // Config modified 24 hours ago
	ch <- prometheus.MustNewConstMetric(
		c.configLastModified,
		prometheus.GaugeValue,
		float64(configTime),
		clusterName, "slurm.conf",
	)

	ch <- prometheus.MustNewConstMetric(
		c.configLastModified,
		prometheus.GaugeValue,
		float64(configTime-3600), // DB config modified 25 hours ago
		clusterName, "slurmdbd.conf",
	)

	// Database connections (simulated)
	ch <- prometheus.MustNewConstMetric(
		c.slurmDBConnections,
		prometheus.GaugeValue,
		5, // 5 active connections
		clusterName,
	)

	// Database latency (simulated)
	ch <- prometheus.MustNewConstMetric(
		c.slurmDBLatency,
		prometheus.GaugeValue,
		0.025, // 25ms query latency
		clusterName, "select",
	)

	ch <- prometheus.MustNewConstMetric(
		c.slurmDBLatency,
		prometheus.GaugeValue,
		0.050, // 50ms insert latency
		clusterName, "insert",
	)
}
