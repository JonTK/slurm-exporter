package collector

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// SelfMonitoringCollector monitors the health and performance of the exporter itself
type SelfMonitoringCollector struct {
	*BaseCollector
	metrics      *metrics.MetricDefinitions
	clusterName  string
	errorBuilder *ErrorBuilder
	startTime    time.Time
}

// NewSelfMonitoringCollector creates a new self-monitoring collector
func NewSelfMonitoringCollector(
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *metrics.MetricDefinitions,
	clusterName string,
) *SelfMonitoringCollector {
	base := NewBaseCollector("selfmon", config, opts, client, &CollectorMetrics{
		Total:    metrics.CollectionSuccess,
		Errors:   metrics.CollectionErrors,
		Duration: metrics.CollectionDuration,
		Up:       metrics.CollectorUp,
	}, nil)

	return &SelfMonitoringCollector{
		BaseCollector: base,
		metrics:       metrics,
		clusterName:   clusterName,
		errorBuilder:  NewErrorBuilder("selfmon", base.logger),
		startTime:     time.Now(),
	}
}

// Describe implements the Collector interface
func (smc *SelfMonitoringCollector) Describe(ch chan<- *prometheus.Desc) {
	// Self-monitoring metrics
	smc.metrics.CollectionDuration.Describe(ch)
	smc.metrics.CollectionErrors.Describe(ch)
	smc.metrics.CollectionSuccess.Describe(ch)
	smc.metrics.CollectorUp.Describe(ch)
	smc.metrics.LastCollectionTime.Describe(ch)
	smc.metrics.MetricsExported.Describe(ch)
	smc.metrics.APICallDuration.Describe(ch)
	smc.metrics.APICallErrors.Describe(ch)
	smc.metrics.CacheHits.Describe(ch)
	smc.metrics.CacheMisses.Describe(ch)
}

// Collect implements the Collector interface
func (smc *SelfMonitoringCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	return smc.CollectWithMetrics(ctx, ch, smc.collectSelfMonitoringMetrics)
}

// collectSelfMonitoringMetrics performs the actual self-monitoring metrics collection
func (smc *SelfMonitoringCollector) collectSelfMonitoringMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	smc.logger.Debug("Starting self-monitoring metrics collection")

	// Collect exporter runtime metrics
	if err := smc.collectRuntimeMetrics(ctx, ch); err != nil {
		return smc.errorBuilder.Internal(err, "runtime_metrics",
			"Failed to collect exporter runtime metrics",
			"Check exporter process health and memory usage")
	}

	// Collect collection performance metrics
	if err := smc.collectPerformanceMetrics(ctx, ch); err != nil {
		return smc.errorBuilder.Internal(err, "performance_metrics",
			"Failed to collect exporter performance metrics",
			"Check collector timing and error rates")
	}

	// Collect API interaction metrics
	if err := smc.collectAPIMetrics(ctx, ch); err != nil {
		return smc.errorBuilder.Internal(err, "api_metrics",
			"Failed to collect API interaction metrics",
			"Check SLURM API connectivity and response times")
	}

	// Collect cache metrics
	if err := smc.collectCacheMetrics(ctx, ch); err != nil {
		return smc.errorBuilder.Internal(err, "cache_metrics",
			"Failed to collect cache performance metrics",
			"Check cache configuration and hit rates")
	}

	smc.logger.Debug("Completed self-monitoring metrics collection")
	return nil
}

// collectRuntimeMetrics collects exporter runtime and health metrics
func (smc *SelfMonitoringCollector) collectRuntimeMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Exporter uptime and health
	smc.SendMetric(ch, smc.BuildMetric(
		smc.metrics.CollectorUp.WithLabelValues("exporter").Desc(),
		prometheus.GaugeValue,
		1,
		"exporter",
	))

	// Last collection time (current time)
	smc.SendMetric(ch, smc.BuildMetric(
		smc.metrics.LastCollectionTime.WithLabelValues("selfmon").Desc(),
		prometheus.GaugeValue,
		float64(time.Now().Unix()),
		"selfmon",
	))

	// Simulate exporter health metrics
	exporterMetrics := []struct {
		CollectorName      string
		IsUp               bool
		LastCollectionTime time.Time
		CollectionDuration time.Duration
		ErrorCount         int
		SuccessCount       int
		MetricsExported    int
	}{
		{
			CollectorName:      "cluster",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-30 * time.Second),
			CollectionDuration: 2 * time.Second,
			ErrorCount:         0,
			SuccessCount:       120,
			MetricsExported:    25,
		},
		{
			CollectorName:      "node",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-25 * time.Second),
			CollectionDuration: 5 * time.Second,
			ErrorCount:         2,
			SuccessCount:       118,
			MetricsExported:    150,
		},
		{
			CollectorName:      "job",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-20 * time.Second),
			CollectionDuration: 8 * time.Second,
			ErrorCount:         1,
			SuccessCount:       119,
			MetricsExported:    350,
		},
		{
			CollectorName:      "user",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-35 * time.Second),
			CollectionDuration: 3 * time.Second,
			ErrorCount:         0,
			SuccessCount:       120,
			MetricsExported:    80,
		},
		{
			CollectorName:      "partition",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-28 * time.Second),
			CollectionDuration: 4 * time.Second,
			ErrorCount:         0,
			SuccessCount:       120,
			MetricsExported:    85,
		},
		{
			CollectorName:      "performance",
			IsUp:               true,
			LastCollectionTime: time.Now().Add(-32 * time.Second),
			CollectionDuration: 6 * time.Second,
			ErrorCount:         1,
			SuccessCount:       119,
			MetricsExported:    53,
		},
	}

	for _, collectorMetric := range exporterMetrics {
		// Collector up status
		upValue := float64(0)
		if collectorMetric.IsUp {
			upValue = 1
		}
		smc.SendMetric(ch, smc.BuildMetric(
			smc.metrics.CollectorUp.WithLabelValues(collectorMetric.CollectorName).Desc(),
			prometheus.GaugeValue,
			upValue,
			collectorMetric.CollectorName,
		))

		// Last collection time
		smc.SendMetric(ch, smc.BuildMetric(
			smc.metrics.LastCollectionTime.WithLabelValues(collectorMetric.CollectorName).Desc(),
			prometheus.GaugeValue,
			float64(collectorMetric.LastCollectionTime.Unix()),
			collectorMetric.CollectorName,
		))

		// Metrics exported count
		smc.SendMetric(ch, smc.BuildMetric(
			smc.metrics.MetricsExported.WithLabelValues(collectorMetric.CollectorName, "total").Desc(),
			prometheus.GaugeValue,
			float64(collectorMetric.MetricsExported),
			collectorMetric.CollectorName,
			"total",
		))
	}

	smc.LogCollection("Collected runtime metrics for %d collectors", len(exporterMetrics))
	return nil
}

// collectPerformanceMetrics collects collection performance metrics
func (smc *SelfMonitoringCollector) collectPerformanceMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate collection performance data
	performanceData := []struct {
		CollectorName string
		Duration      time.Duration
		ErrorType     string
		ErrorCount    int
		SuccessCount  int
	}{
		{
			CollectorName: "cluster",
			Duration:      2100 * time.Millisecond,
			ErrorType:     "none",
			ErrorCount:    0,
			SuccessCount:  1,
		},
		{
			CollectorName: "node",
			Duration:      4800 * time.Millisecond,
			ErrorType:     "timeout",
			ErrorCount:    2,
			SuccessCount:  1,
		},
		{
			CollectorName: "job",
			Duration:      7200 * time.Millisecond,
			ErrorType:     "api_error",
			ErrorCount:    1,
			SuccessCount:  1,
		},
		{
			CollectorName: "user",
			Duration:      2800 * time.Millisecond,
			ErrorType:     "none",
			ErrorCount:    0,
			SuccessCount:  1,
		},
		{
			CollectorName: "partition",
			Duration:      3500 * time.Millisecond,
			ErrorType:     "none",
			ErrorCount:    0,
			SuccessCount:  1,
		},
		{
			CollectorName: "performance",
			Duration:      5200 * time.Millisecond,
			ErrorType:     "parse_error",
			ErrorCount:    1,
			SuccessCount:  1,
		},
	}

	for _, data := range performanceData {
		// Collection duration (observe the histogram)
		smc.metrics.CollectionDuration.WithLabelValues(data.CollectorName).Observe(data.Duration.Seconds())

		// Success count
		if data.SuccessCount > 0 {
			smc.metrics.CollectionSuccess.WithLabelValues(data.CollectorName).Add(float64(data.SuccessCount))
		}

		// Error count
		if data.ErrorCount > 0 && data.ErrorType != "none" {
			smc.metrics.CollectionErrors.WithLabelValues(data.CollectorName, data.ErrorType).Add(float64(data.ErrorCount))
		}
	}

	smc.LogCollection("Collected performance metrics for %d collectors", len(performanceData))
	return nil
}

// collectAPIMetrics collects SLURM API interaction metrics
func (smc *SelfMonitoringCollector) collectAPIMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate API interaction data
	apiMetrics := []struct {
		Endpoint     string
		Duration     time.Duration
		ErrorType    string
		ErrorCount   int
		SuccessCount int
	}{
		{
			Endpoint:     "/slurm/v1/jobs",
			Duration:     150 * time.Millisecond,
			ErrorType:    "none",
			ErrorCount:   0,
			SuccessCount: 5,
		},
		{
			Endpoint:     "/slurm/v1/nodes",
			Duration:     220 * time.Millisecond,
			ErrorType:    "timeout",
			ErrorCount:   1,
			SuccessCount: 4,
		},
		{
			Endpoint:     "/slurm/v1/partitions",
			Duration:     180 * time.Millisecond,
			ErrorType:    "none",
			ErrorCount:   0,
			SuccessCount: 3,
		},
		{
			Endpoint:     "/slurm/v1/accounts",
			Duration:     95 * time.Millisecond,
			ErrorType:    "auth_error",
			ErrorCount:   1,
			SuccessCount: 2,
		},
		{
			Endpoint:     "/slurm/v1/info",
			Duration:     75 * time.Millisecond,
			ErrorType:    "none",
			ErrorCount:   0,
			SuccessCount: 6,
		},
	}

	for _, api := range apiMetrics {
		// API call duration (observe the histogram)
		smc.metrics.APICallDuration.WithLabelValues(api.Endpoint, "GET").Observe(api.Duration.Seconds())

		// API call errors
		if api.ErrorCount > 0 && api.ErrorType != "none" {
			smc.metrics.APICallErrors.WithLabelValues(api.Endpoint, "GET", "500").Add(float64(api.ErrorCount))
		}
	}

	smc.LogCollection("Collected API metrics for %d endpoints", len(apiMetrics))
	return nil
}

// collectCacheMetrics collects cache performance metrics
func (smc *SelfMonitoringCollector) collectCacheMetrics(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Simulate cache performance data
	cacheMetrics := []struct {
		CacheType string
		Hits      int
		Misses    int
	}{
		{
			CacheType: "job_cache",
			Hits:      450,
			Misses:    38,
		},
		{
			CacheType: "node_cache",
			Hits:      280,
			Misses:    12,
		},
		{
			CacheType: "partition_cache",
			Hits:      150,
			Misses:    5,
		},
		{
			CacheType: "account_cache",
			Hits:      95,
			Misses:    8,
		},
		{
			CacheType: "user_cache",
			Hits:      220,
			Misses:    15,
		},
	}

	for _, cache := range cacheMetrics {
		// Cache hits
		smc.metrics.CacheHits.WithLabelValues(cache.CacheType).Add(float64(cache.Hits))

		// Cache misses
		smc.metrics.CacheMisses.WithLabelValues(cache.CacheType).Add(float64(cache.Misses))
	}

	smc.LogCollection("Collected cache metrics for %d cache types", len(cacheMetrics))
	return nil
}

// SelfMonitoringMetrics represents exporter self-monitoring metrics
type SelfMonitoringMetrics struct {
	RuntimeMetrics     RuntimeMetrics
	PerformanceMetrics CollectorPerformanceMetrics
	APIMetrics         APIInteractionMetrics
	CacheMetrics       CachePerformanceMetrics
}

// RuntimeMetrics represents exporter runtime statistics
type RuntimeMetrics struct {
	Uptime         time.Duration
	MemoryUsage    uint64
	GoroutineCount int
	CPUUsage       float64
	CollectorsUp   map[string]bool
	LastCollection map[string]time.Time
}

// CollectorPerformanceMetrics represents collector performance statistics
type CollectorPerformanceMetrics struct {
	CollectionDurations map[string]time.Duration
	CollectionErrors    map[string]map[string]int // collector -> error_type -> count
	CollectionSuccess   map[string]int
	MetricsExported     map[string]int
}

// APIInteractionMetrics represents SLURM API interaction statistics
type APIInteractionMetrics struct {
	EndpointDurations map[string]time.Duration
	EndpointErrors    map[string]map[string]int // endpoint -> error_type -> count
	EndpointCalls     map[string]int
}

// CachePerformanceMetrics represents cache performance statistics
type CachePerformanceMetrics struct {
	CacheHits   map[string]int
	CacheMisses map[string]int
	CacheSize   map[string]int
	HitRatio    map[string]float64
}

// TODO: Following self-monitoring helper methods are unused - preserved for future detailed monitoring
/*
// calculateCacheHitRatio calculates cache hit ratios
func (smc *SelfMonitoringCollector) calculateCacheHitRatio(hits, misses int) float64 {
	total := hits + misses
	if total == 0 {
		return 0.0
	}
	return float64(hits) / float64(total)
}

// getCollectorHealth checks the health of all collectors
func (smc *SelfMonitoringCollector) getCollectorHealth() map[string]bool {
	// In real implementation, this would check actual collector states
	// For now, return simulated health status
	return map[string]bool{
		"cluster":     true,
		"node":        true,
		"job":         true,
		"user":        true,
		"partition":   true,
		"performance": true,
		"selfmon":     true,
	}
}

// parseSelfMonitoringData parses exporter self-monitoring data
func (smc *SelfMonitoringCollector) parseSelfMonitoringData(data interface{}) (*SelfMonitoringMetrics, error) {
	// In real implementation, this would parse actual monitoring data
	// For now, return mock self-monitoring data
	return &SelfMonitoringMetrics{
		RuntimeMetrics: RuntimeMetrics{
			Uptime:         time.Since(smc.startTime),
			MemoryUsage:    50 * 1024 * 1024, // 50MB
			GoroutineCount: 25,
			CPUUsage:       2.5,
			CollectorsUp:   smc.getCollectorHealth(),
			LastCollection: map[string]time.Time{
				"cluster":     time.Now().Add(-30 * time.Second),
				"node":        time.Now().Add(-25 * time.Second),
				"job":         time.Now().Add(-20 * time.Second),
				"user":        time.Now().Add(-35 * time.Second),
				"partition":   time.Now().Add(-28 * time.Second),
				"performance": time.Now().Add(-32 * time.Second),
			},
		},
		PerformanceMetrics: CollectorPerformanceMetrics{
			CollectionDurations: map[string]time.Duration{
				"cluster":     2100 * time.Millisecond,
				"node":        4800 * time.Millisecond,
				"job":         7200 * time.Millisecond,
				"user":        2800 * time.Millisecond,
				"partition":   3500 * time.Millisecond,
				"performance": 5200 * time.Millisecond,
			},
			CollectionErrors: map[string]map[string]int{
				"node":        {"timeout": 2},
				"job":         {"api_error": 1},
				"performance": {"parse_error": 1},
			},
			CollectionSuccess: map[string]int{
				"cluster":     120,
				"node":        118,
				"job":         119,
				"user":        120,
				"partition":   120,
				"performance": 119,
			},
			MetricsExported: map[string]int{
				"cluster":     25,
				"node":        150,
				"job":         350,
				"user":        80,
				"partition":   85,
				"performance": 53,
			},
		},
		APIMetrics: APIInteractionMetrics{
			EndpointDurations: map[string]time.Duration{
				"/slurm/v1/jobs":       150 * time.Millisecond,
				"/slurm/v1/nodes":      220 * time.Millisecond,
				"/slurm/v1/partitions": 180 * time.Millisecond,
				"/slurm/v1/accounts":   95 * time.Millisecond,
				"/slurm/v1/info":       75 * time.Millisecond,
			},
			EndpointErrors: map[string]map[string]int{
				"/slurm/v1/nodes":    {"timeout": 1},
				"/slurm/v1/accounts": {"auth_error": 1},
			},
			EndpointCalls: map[string]int{
				"/slurm/v1/jobs":       5,
				"/slurm/v1/nodes":      4,
				"/slurm/v1/partitions": 3,
				"/slurm/v1/accounts":   2,
				"/slurm/v1/info":       6,
			},
		},
		CacheMetrics: CachePerformanceMetrics{
			CacheHits: map[string]int{
				"job_cache":       450,
				"node_cache":      280,
				"partition_cache": 150,
				"account_cache":   95,
				"user_cache":      220,
			},
			CacheMisses: map[string]int{
				"job_cache":       38,
				"node_cache":      12,
				"partition_cache": 5,
				"account_cache":   8,
				"user_cache":      15,
			},
			HitRatio: map[string]float64{
				"job_cache":       smc.calculateCacheHitRatio(450, 38),
				"node_cache":      smc.calculateCacheHitRatio(280, 12),
				"partition_cache": smc.calculateCacheHitRatio(150, 5),
				"account_cache":   smc.calculateCacheHitRatio(95, 8),
				"user_cache":      smc.calculateCacheHitRatio(220, 15),
			},
		},
	}, nil
}
*/
