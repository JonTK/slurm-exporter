// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package metrics

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

// PerformanceMonitor collects runtime performance metrics about the exporter
type PerformanceMonitor struct {
	metrics  *PerformanceMetrics
	process  *process.Process
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	enabled  bool
	mu       sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(metrics *PerformanceMetrics, interval time.Duration) (*PerformanceMonitor, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &PerformanceMonitor{
		metrics:  metrics,
		process:  proc,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		enabled:  true,
	}, nil
}

// Start begins monitoring performance metrics
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.enabled {
		return
	}

	pm.wg.Add(1)
	go pm.monitorLoop()
}

// Stop stops the performance monitor
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.cancel != nil {
		pm.cancel()
	}
	pm.wg.Wait()
}

// SetEnabled enables or disables the monitor
func (pm *PerformanceMonitor) SetEnabled(enabled bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.enabled = enabled
}

// monitorLoop is the main monitoring loop
func (pm *PerformanceMonitor) monitorLoop() {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.collectMetrics()
		}
	}
}

// collectMetrics collects all performance metrics
func (pm *PerformanceMonitor) collectMetrics() {
	pm.mu.RLock()
	enabled := pm.enabled
	pm.mu.RUnlock()

	if !enabled {
		return
	}

	// Collect memory metrics
	pm.collectMemoryMetrics()

	// Collect CPU metrics
	pm.collectCPUMetrics()

	// Collect goroutine metrics
	pm.collectGoroutineMetrics()
}

// collectMemoryMetrics collects memory usage metrics
func (pm *PerformanceMonitor) collectMemoryMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Go runtime memory stats
	pm.metrics.UpdateMemoryUsage("heap_alloc", float64(m.Alloc))
	pm.metrics.UpdateMemoryUsage("heap_sys", float64(m.HeapSys))
	pm.metrics.UpdateMemoryUsage("heap_idle", float64(m.HeapIdle))
	pm.metrics.UpdateMemoryUsage("heap_inuse", float64(m.HeapInuse))
	pm.metrics.UpdateMemoryUsage("stack_inuse", float64(m.StackInuse))
	pm.metrics.UpdateMemoryUsage("stack_sys", float64(m.StackSys))
	pm.metrics.UpdateMemoryUsage("gc_sys", float64(m.GCSys))

	// Process memory (if available)
	if memInfo, err := pm.process.MemoryInfo(); err == nil {
		pm.metrics.UpdateMemoryUsage("rss", float64(memInfo.RSS))
		pm.metrics.UpdateMemoryUsage("vms", float64(memInfo.VMS))
	}

	// Memory percent (if available)
	if memPercent, err := pm.process.MemoryPercent(); err == nil {
		pm.metrics.UpdateMemoryUsage("percent", float64(memPercent))
	}
}

// collectCPUMetrics collects CPU usage metrics
func (pm *PerformanceMonitor) collectCPUMetrics() {
	// CPU percent over the monitoring interval
	if cpuPercent, err := pm.process.CPUPercent(); err == nil {
		pm.metrics.UpdateCPUUsage("total", cpuPercent)
	}

	// System-wide CPU stats (if available)
	if cpuTimes, err := cpu.Times(false); err == nil && len(cpuTimes) > 0 {
		ct := cpuTimes[0]
		total := ct.User + ct.System + ct.Idle + ct.Nice + ct.Iowait + ct.Irq + ct.Softirq + ct.Steal

		if total > 0 {
			pm.metrics.UpdateCPUUsage("user", (ct.User/total)*100)
			pm.metrics.UpdateCPUUsage("system", (ct.System/total)*100)
			pm.metrics.UpdateCPUUsage("idle", (ct.Idle/total)*100)
		}
	}
}

// collectGoroutineMetrics collects goroutine metrics
func (pm *PerformanceMonitor) collectGoroutineMetrics() {
	numGoroutines := runtime.NumGoroutine()
	pm.metrics.UpdateThreadPoolUtilization("goroutines", float64(numGoroutines))

	numCPU := runtime.NumCPU()
	if numCPU > 0 {
		utilizationPercent := (float64(numGoroutines) / float64(numCPU)) * 100
		pm.metrics.UpdateThreadPoolUtilization("cpu_ratio", utilizationPercent)
	}
}

// CardinalityTracker tracks metric cardinality for collectors
type CardinalityTracker struct {
	metrics    *PerformanceMetrics
	gatherer   prometheus.Gatherer
	mu         sync.RWMutex
	lastUpdate time.Time
	interval   time.Duration
}

// NewCardinalityTracker creates a new cardinality tracker
func NewCardinalityTracker(metrics *PerformanceMetrics, gatherer prometheus.Gatherer, interval time.Duration) *CardinalityTracker {
	return &CardinalityTracker{
		metrics:  metrics,
		gatherer: gatherer,
		interval: interval,
	}
}

// Update updates cardinality metrics if enough time has passed
func (ct *CardinalityTracker) Update() {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	if time.Since(ct.lastUpdate) < ct.interval {
		return
	}

	ct.updateCardinality()
	ct.lastUpdate = time.Now()
}

// updateCardinality collects and updates cardinality metrics
func (ct *CardinalityTracker) updateCardinality() {
	mfs, err := ct.gatherer.Gather()
	if err != nil {
		return
	}

	cardinalityMap := make(map[string]int)

	for _, mf := range mfs {
		if mf.GetName() == "" {
			continue
		}

		metricName := mf.GetName()
		count := 0

		for _, metric := range mf.GetMetric() {
			count++
			// Count unique label combinations
			if len(metric.GetLabel()) > 0 {
				// Each metric with unique label set counts as one
				count += len(metric.GetLabel())
			}
		}

		cardinalityMap[metricName] = count
	}

	// Update cardinality metrics
	for metricName, cardinality := range cardinalityMap {
		// Try to determine which collector this metric belongs to
		collector := "unknown"
		if len(metricName) > 5 && metricName[:5] == "slurm" {
			// Parse collector from metric name if possible
			collector = extractCollectorFromMetric(metricName)
		}

		ct.metrics.UpdateCardinality(metricName, collector, float64(cardinality))
	}
}

// extractCollectorFromMetric attempts to extract collector name from metric name
func extractCollectorFromMetric(metricName string) string {
	// Common patterns in SLURM exporter metrics
	if contains(metricName, "job") {
		return "jobs"
	}
	if contains(metricName, "node") {
		return "nodes"
	}
	if contains(metricName, "partition") {
		return "partitions"
	}
	if contains(metricName, "account") {
		return "accounts"
	}
	if contains(metricName, "qos") {
		return "qos"
	}
	if contains(metricName, "reservation") {
		return "reservations"
	}
	if contains(metricName, "fairshare") {
		return "fairshare"
	}
	if contains(metricName, "cluster") {
		return "cluster"
	}
	if contains(metricName, "system") {
		return "system"
	}

	return "unknown"
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && findInString(s, substr))
}

// findInString checks if substring exists in string
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CacheMetricsRecorder provides an interface for recording cache metrics
type CacheMetricsRecorder interface {
	RecordHit(cacheType, collector string)
	RecordMiss(cacheType, collector string)
	UpdateSize(cacheType string, sizeBytes float64)
}

// cacheMetricsRecorder implements CacheMetricsRecorder
type cacheMetricsRecorder struct {
	metrics *PerformanceMetrics
}

// NewCacheMetricsRecorder creates a new cache metrics recorder
func NewCacheMetricsRecorder(metrics *PerformanceMetrics) CacheMetricsRecorder {
	return &cacheMetricsRecorder{metrics: metrics}
}

// RecordHit records a cache hit
func (cmr *cacheMetricsRecorder) RecordHit(cacheType, collector string) {
	cmr.metrics.RecordCacheHit(cacheType, collector)
}

// RecordMiss records a cache miss
func (cmr *cacheMetricsRecorder) RecordMiss(cacheType, collector string) {
	cmr.metrics.RecordCacheMiss(cacheType, collector)
}

// UpdateSize updates cache size
func (cmr *cacheMetricsRecorder) UpdateSize(cacheType string, sizeBytes float64) {
	cmr.metrics.UpdateCacheSize(cacheType, sizeBytes)
}
