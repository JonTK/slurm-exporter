package health

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"os"
)

// SLURMClient defines the interface for SLURM client operations needed for health checks
type SLURMClient interface {
	Ping(ctx context.Context) error
	GetInfo(ctx context.Context) (interface{}, error)
}

// CollectorRegistry defines the interface for collector registry operations
type CollectorRegistry interface {
	GetStates() map[string]CollectorState
}

// CollectorState represents the state of a collector
type CollectorState struct {
	Name              string
	Enabled           bool
	LastCollection    time.Time
	LastDuration      time.Duration
	LastError         error
	ConsecutiveErrors int
	TotalCollections  int64
	TotalErrors       int64
}

// NewSLURMAPIHealthCheck creates a comprehensive SLURM API health check
func NewSLURMAPIHealthCheck(client SLURMClient, timeout time.Duration) CheckFunc {
	return func(ctx context.Context) Check {
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		check := Check{
			Status: StatusUnknown,
			Metadata: map[string]string{
				"timeout": timeout.String(),
			},
		}

		// Test basic connectivity with ping
		if err := client.Ping(checkCtx); err != nil {
			check.Status = StatusUnhealthy
			check.Error = fmt.Sprintf("SLURM API ping failed: %v", err)
			check.Message = "Cannot reach SLURM API"
			check.Metadata["ping_error"] = err.Error()
			return check
		}

		check.Metadata["ping_successful"] = "true"

		// Test API functionality
		info, err := client.GetInfo(checkCtx)
		if err != nil {
			check.Status = StatusDegraded
			check.Error = fmt.Sprintf("SLURM API info call failed: %v", err)
			check.Message = "SLURM API ping successful but info call failed"
			check.Metadata["info_error"] = err.Error()
			return check
		}

		check.Status = StatusHealthy
		check.Message = "SLURM API is responding normally"
		check.Metadata["info_successful"] = "true"

		// Add some info details if available
		if info != nil {
			check.Metadata["slurm_info_available"] = "true"
		}

		return check
	}
}

// NewCollectorsHealthCheck creates a health check for metric collectors
func NewCollectorsHealthCheck(registry CollectorRegistry, threshold time.Duration) CheckFunc {
	return func(ctx context.Context) Check {
		check := Check{
			Status:   StatusHealthy,
			Message:  "All collectors are healthy",
			Metadata: make(map[string]string),
		}

		states := registry.GetStates()

		var unhealthy, degraded, healthy []string
		var totalCollectors, enabledCollectors, errorCollectors int

		for name, state := range states {
			totalCollectors++

			if !state.Enabled {
				check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "disabled"
				continue
			}

			enabledCollectors++

			// Check if collector is stale
			age := time.Since(state.LastCollection)
			if age > threshold {
				unhealthy = append(unhealthy, fmt.Sprintf("%s (stale: %v)", name, age))
				check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "stale"
				check.Metadata[fmt.Sprintf("collector_%s_age", name)] = age.String()
				continue
			}

			// Check for consecutive errors
			if state.ConsecutiveErrors > 3 {
				unhealthy = append(unhealthy, fmt.Sprintf("%s (%d consecutive errors)", name, state.ConsecutiveErrors))
				errorCollectors++
				check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "error_critical"
				check.Metadata[fmt.Sprintf("collector_%s_consecutive_errors", name)] = fmt.Sprintf("%d", state.ConsecutiveErrors)
				continue
			} else if state.ConsecutiveErrors > 0 {
				degraded = append(degraded, fmt.Sprintf("%s (%d consecutive errors)", name, state.ConsecutiveErrors))
				errorCollectors++
				check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "error_degraded"
				check.Metadata[fmt.Sprintf("collector_%s_consecutive_errors", name)] = fmt.Sprintf("%d", state.ConsecutiveErrors)
				continue
			}

			// Check error rate
			if state.TotalCollections > 0 {
				errorRate := float64(state.TotalErrors) / float64(state.TotalCollections)
				if errorRate > 0.5 {
					degraded = append(degraded, fmt.Sprintf("%s (%.1f%% error rate)", name, errorRate*100))
					check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "high_error_rate"
					check.Metadata[fmt.Sprintf("collector_%s_error_rate", name)] = fmt.Sprintf("%.1f", errorRate*100)
					continue
				}
			}

			healthy = append(healthy, name)
			check.Metadata[fmt.Sprintf("collector_%s_status", name)] = "healthy"
		}

		// Set overall metadata
		check.Metadata["total_collectors"] = fmt.Sprintf("%d", totalCollectors)
		check.Metadata["enabled_collectors"] = fmt.Sprintf("%d", enabledCollectors)
		check.Metadata["healthy_collectors"] = fmt.Sprintf("%d", len(healthy))
		check.Metadata["degraded_collectors"] = fmt.Sprintf("%d", len(degraded))
		check.Metadata["unhealthy_collectors"] = fmt.Sprintf("%d", len(unhealthy))
		check.Metadata["collectors_with_errors"] = fmt.Sprintf("%d", errorCollectors)

		// Determine overall status and message
		if len(unhealthy) > 0 {
			check.Status = StatusUnhealthy
			check.Message = fmt.Sprintf("Unhealthy collectors: %v", unhealthy)
			check.Error = fmt.Sprintf("%d collectors are unhealthy", len(unhealthy))
		} else if len(degraded) > 0 {
			check.Status = StatusDegraded
			check.Message = fmt.Sprintf("Degraded collectors: %v", degraded)
		} else {
			check.Status = StatusHealthy
			check.Message = fmt.Sprintf("All %d enabled collectors are healthy", len(healthy))
		}

		return check
	}
}

// NewMemoryHealthCheck creates a memory usage health check
func NewMemoryHealthCheck(cfg config.PerformanceMonitoringConfig) CheckFunc {
	return func(ctx context.Context) Check {
		check := Check{
			Status:   StatusHealthy,
			Metadata: make(map[string]string),
		}

		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		// Get process memory info if available
		proc, err := process.NewProcess(int32(os.Getpid()))
		var rss, vms uint64
		var memPercent float32

		if err == nil {
			if memInfo, err := proc.MemoryInfo(); err == nil {
				rss = memInfo.RSS
				vms = memInfo.VMS
			}
			if percent, err := proc.MemoryPercent(); err == nil {
				memPercent = percent
			}
		}

		// Record memory statistics
		heapUsed := memStats.Alloc
		check.Metadata["heap_alloc_mb"] = fmt.Sprintf("%d", heapUsed/1024/1024)
		check.Metadata["heap_sys_mb"] = fmt.Sprintf("%d", memStats.HeapSys/1024/1024)
		check.Metadata["heap_idle_mb"] = fmt.Sprintf("%d", memStats.HeapIdle/1024/1024)
		check.Metadata["heap_inuse_mb"] = fmt.Sprintf("%d", memStats.HeapInuse/1024/1024)
		check.Metadata["stack_inuse_mb"] = fmt.Sprintf("%d", memStats.StackInuse/1024/1024)
		check.Metadata["gc_sys_mb"] = fmt.Sprintf("%d", memStats.GCSys/1024/1024)
		check.Metadata["num_gc"] = fmt.Sprintf("%d", memStats.NumGC)
		check.Metadata["gc_cpu_fraction"] = fmt.Sprintf("%.6f", memStats.GCCPUFraction)

		if rss > 0 {
			check.Metadata["rss_mb"] = fmt.Sprintf("%d", rss/1024/1024)
		}
		if vms > 0 {
			check.Metadata["vms_mb"] = fmt.Sprintf("%d", vms/1024/1024)
		}
		if memPercent > 0 {
			check.Metadata["memory_percent"] = fmt.Sprintf("%.1f", memPercent)
		}

		// Determine status based on heap usage and threshold
		if cfg.MemoryThreshold > 0 && heapUsed > uint64(cfg.MemoryThreshold) {
			check.Status = StatusUnhealthy
			check.Message = fmt.Sprintf("Memory usage too high: %d MB (threshold: %d MB)",
				heapUsed/1024/1024, cfg.MemoryThreshold/1024/1024)
			check.Error = "Memory usage exceeds threshold"
		} else if cfg.MemoryThreshold > 0 && heapUsed > uint64(float64(cfg.MemoryThreshold)*0.8) {
			check.Status = StatusDegraded
			check.Message = fmt.Sprintf("Memory usage elevated: %d MB (threshold: %d MB)",
				heapUsed/1024/1024, cfg.MemoryThreshold/1024/1024)
		} else {
			check.Status = StatusHealthy
			check.Message = fmt.Sprintf("Memory usage normal: %d MB", heapUsed/1024/1024)
		}

		check.Metadata["threshold_mb"] = fmt.Sprintf("%d", cfg.MemoryThreshold/1024/1024)

		return check
	}
}

// NewDiskHealthCheck creates a disk usage health check
func NewDiskHealthCheck(path string, threshold float64) CheckFunc {
	return func(ctx context.Context) Check {
		check := Check{
			Status: StatusHealthy,
			Metadata: map[string]string{
				"path":      path,
				"threshold": fmt.Sprintf("%.1f", threshold),
			},
		}

		usage, err := disk.Usage(path)
		if err != nil {
			check.Status = StatusUnhealthy
			check.Error = fmt.Sprintf("Failed to get disk usage for %s: %v", path, err)
			check.Message = "Disk usage check failed"
			return check
		}

		usedPercent := usage.UsedPercent
		check.Metadata["total_gb"] = fmt.Sprintf("%.1f", float64(usage.Total)/1024/1024/1024)
		check.Metadata["used_gb"] = fmt.Sprintf("%.1f", float64(usage.Used)/1024/1024/1024)
		check.Metadata["free_gb"] = fmt.Sprintf("%.1f", float64(usage.Free)/1024/1024/1024)
		check.Metadata["used_percent"] = fmt.Sprintf("%.1f", usedPercent)

		if usedPercent > threshold {
			check.Status = StatusUnhealthy
			check.Message = fmt.Sprintf("Disk usage too high: %.1f%% (threshold: %.1f%%)", usedPercent, threshold)
			check.Error = "Disk usage exceeds threshold"
		} else if usedPercent > threshold*0.8 {
			check.Status = StatusDegraded
			check.Message = fmt.Sprintf("Disk usage elevated: %.1f%% (threshold: %.1f%%)", usedPercent, threshold)
		} else {
			check.Status = StatusHealthy
			check.Message = fmt.Sprintf("Disk usage normal: %.1f%%", usedPercent)
		}

		return check
	}
}

// NewNetworkHealthCheck creates a network connectivity health check
func NewNetworkHealthCheck() CheckFunc {
	return func(ctx context.Context) Check {
		check := Check{
			Status:   StatusHealthy,
			Metadata: make(map[string]string),
		}

		// Get network interface statistics
		stats, err := net.IOCounters(false)
		if err != nil {
			check.Status = StatusDegraded
			check.Error = fmt.Sprintf("Failed to get network stats: %v", err)
			check.Message = "Network statistics unavailable"
			return check
		}

		if len(stats) == 0 {
			check.Status = StatusUnhealthy
			check.Message = "No network interfaces found"
			check.Error = "No active network interfaces"
			return check
		}

		check.Status = StatusHealthy
		check.Message = fmt.Sprintf("Network interfaces active: %d", len(stats))
		check.Metadata["interface_count"] = fmt.Sprintf("%d", len(stats))

		// Add interface statistics
		for i, stat := range stats {
			if i >= 3 { // Limit to first 3 interfaces to avoid too much metadata
				break
			}
			prefix := fmt.Sprintf("interface_%s", stat.Name)
			check.Metadata[prefix+"_bytes_sent"] = fmt.Sprintf("%d", stat.BytesSent)
			check.Metadata[prefix+"_bytes_recv"] = fmt.Sprintf("%d", stat.BytesRecv)
			check.Metadata[prefix+"_packets_sent"] = fmt.Sprintf("%d", stat.PacketsSent)
			check.Metadata[prefix+"_packets_recv"] = fmt.Sprintf("%d", stat.PacketsRecv)
		}

		return check
	}
}

// NewCircuitBreakerHealthCheck creates a health check for circuit breakers
func NewCircuitBreakerHealthCheck(getStatuses func() map[string]interface{}) CheckFunc {
	return func(ctx context.Context) Check {
		check := Check{
			Status:   StatusHealthy,
			Metadata: make(map[string]string),
		}

		statuses := getStatuses()

		var unhealthy, degraded []string
		healthyCount := 0

		for name, status := range statuses {
			// Assume status has a State field that can be compared
			if statusMap, ok := status.(map[string]interface{}); ok {
				if state, exists := statusMap["state"]; exists {
					stateStr := fmt.Sprintf("%v", state)
					check.Metadata[fmt.Sprintf("circuit_breaker_%s_state", name)] = stateStr

					switch stateStr {
					case "open":
						unhealthy = append(unhealthy, name)
					case "half-open":
						degraded = append(degraded, name)
					case "closed":
						healthyCount++
					}
				}
			}
		}

		check.Metadata["total_circuit_breakers"] = fmt.Sprintf("%d", len(statuses))
		check.Metadata["healthy_circuit_breakers"] = fmt.Sprintf("%d", healthyCount)
		check.Metadata["degraded_circuit_breakers"] = fmt.Sprintf("%d", len(degraded))
		check.Metadata["unhealthy_circuit_breakers"] = fmt.Sprintf("%d", len(unhealthy))

		if len(unhealthy) > 0 {
			check.Status = StatusUnhealthy
			check.Message = fmt.Sprintf("Circuit breakers open: %v", unhealthy)
			check.Error = fmt.Sprintf("%d circuit breakers are open", len(unhealthy))
		} else if len(degraded) > 0 {
			check.Status = StatusDegraded
			check.Message = fmt.Sprintf("Circuit breakers half-open: %v", degraded)
		} else {
			check.Status = StatusHealthy
			check.Message = fmt.Sprintf("All %d circuit breakers are closed", healthyCount)
		}

		return check
	}
}
