package performance

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// MemoryOptimizer manages memory usage and garbage collection optimization
type MemoryOptimizer struct {
	logger     *logrus.Entry
	gcPercent  int
	memLimit   uint64
	lastGC     time.Time
	gcInterval time.Duration
	mu         sync.RWMutex

	// Metrics
	memoryUsage    prometheus.Gauge
	gcDuration     prometheus.Histogram
	gcCount        prometheus.Counter
	allocationRate prometheus.Histogram
	objectPools    map[string]*sync.Pool
}

// NewMemoryOptimizer creates a new memory optimizer
func NewMemoryOptimizer(logger *logrus.Entry) *MemoryOptimizer {
	mo := &MemoryOptimizer{
		logger:      logger,
		gcPercent:   100, // Default Go GC target
		gcInterval:  30 * time.Second,
		lastGC:      time.Now(),
		objectPools: make(map[string]*sync.Pool),
	}

	mo.initMetrics()
	mo.setupObjectPools()
	mo.optimizeGCSettings()

	return mo
}

// initMetrics initializes Prometheus metrics for memory monitoring
func (mo *MemoryOptimizer) initMetrics() {
	mo.memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slurm_exporter_memory_usage_bytes",
		Help: "Current memory usage in bytes",
	})

	mo.gcDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "slurm_exporter_gc_duration_seconds",
		Help:    "Time spent in garbage collection",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	})

	mo.gcCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_gc_total",
		Help: "Total number of garbage collections",
	})

	mo.allocationRate = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "slurm_exporter_allocation_rate_bytes_per_second",
		Help:    "Memory allocation rate in bytes per second",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 20),
	})
}

// setupObjectPools creates object pools for frequently allocated objects
func (mo *MemoryOptimizer) setupObjectPools() {
	// Pool for metric slices
	mo.objectPools["metrics"] = &sync.Pool{
		New: func() interface{} {
			return make([]prometheus.Metric, 0, 100)
		},
	}

	// Pool for label slices
	mo.objectPools["labels"] = &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 10)
		},
	}

	// Pool for string builders
	mo.objectPools["strings"] = &sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 50)
		},
	}

	// Pool for metric channels
	mo.objectPools["channels"] = &sync.Pool{
		New: func() interface{} {
			return make(chan prometheus.Metric, 1000)
		},
	}
}

// optimizeGCSettings configures garbage collection for better performance
func (mo *MemoryOptimizer) optimizeGCSettings() {
	// Set GC target percentage based on available memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Adjust GC target based on heap size
	if m.HeapAlloc > 100*1024*1024 { // > 100MB
		mo.gcPercent = 50 // More aggressive GC for large heaps
	} else {
		mo.gcPercent = 200 // Less aggressive for small heaps
	}

	debug.SetGCPercent(mo.gcPercent)

	// Set memory limit if configured
	if mo.memLimit > 0 {
		debug.SetMemoryLimit(int64(mo.memLimit))
	}

	mo.logger.WithFields(logrus.Fields{
		"gc_percent":   mo.gcPercent,
		"memory_limit": mo.memLimit,
		"heap_alloc":   m.HeapAlloc,
	}).Info("Optimized GC settings")
}

// GetMetricPool returns a pooled metric slice
func (mo *MemoryOptimizer) GetMetricPool() []prometheus.Metric {
	return mo.objectPools["metrics"].Get().([]prometheus.Metric)[:0]
}

// PutMetricPool returns a metric slice to the pool
func (mo *MemoryOptimizer) PutMetricPool(metrics []prometheus.Metric) {
	if cap(metrics) < 1000 { // Only pool reasonably sized slices
		mo.objectPools["metrics"].Put(&metrics)
	}
}

// GetLabelPool returns a pooled label slice
func (mo *MemoryOptimizer) GetLabelPool() []string {
	return mo.objectPools["labels"].Get().([]string)[:0]
}

// PutLabelPool returns a label slice to the pool
func (mo *MemoryOptimizer) PutLabelPool(labels []string) {
	if cap(labels) < 100 {
		mo.objectPools["labels"].Put(&labels)
	}
}

// GetChannelPool returns a pooled metric channel
func (mo *MemoryOptimizer) GetChannelPool() chan prometheus.Metric {
	ch := mo.objectPools["channels"].Get().(chan prometheus.Metric)
	// Drain any leftover metrics
	for len(ch) > 0 {
		<-ch
	}
	return ch
}

// PutChannelPool returns a metric channel to the pool
func (mo *MemoryOptimizer) PutChannelPool(ch chan prometheus.Metric) {
	if cap(ch) <= 2000 {
		mo.objectPools["channels"].Put(ch)
	}
}

// ForceGC triggers garbage collection if enough time has passed
func (mo *MemoryOptimizer) ForceGC() {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	if time.Since(mo.lastGC) > mo.gcInterval {
		start := time.Now()
		runtime.GC()
		duration := time.Since(start)

		mo.lastGC = time.Now()
		mo.gcCount.Inc()
		mo.gcDuration.Observe(duration.Seconds())

		mo.logger.WithField("duration", duration).Debug("Forced garbage collection")
	}
}

// UpdateMemoryStats updates memory usage metrics
func (mo *MemoryOptimizer) UpdateMemoryStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mo.memoryUsage.Set(float64(m.Alloc))

	// Calculate allocation rate
	if mo.lastGC.IsZero() {
		mo.lastGC = time.Now()
	} else {
		timeDiff := time.Since(mo.lastGC).Seconds()
		if timeDiff > 0 {
			allocRate := float64(m.TotalAlloc) / timeDiff
			mo.allocationRate.Observe(allocRate)
		}
	}
}

// GetMemoryStats returns current memory statistics
func (mo *MemoryOptimizer) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		GCCPUFraction: m.GCCPUFraction,
		NumGC:         m.NumGC,
		LastGC:        time.Unix(0, int64(m.LastGC)),
	}
}

// SetMemoryLimit sets the memory limit for the optimizer
func (mo *MemoryOptimizer) SetMemoryLimit(limit uint64) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.memLimit = limit
	if limit > 0 {
		debug.SetMemoryLimit(int64(limit))
		mo.logger.WithField("limit", limit).Info("Set memory limit")
	}
}

// SetGCPercent sets the garbage collection target percentage
func (mo *MemoryOptimizer) SetGCPercent(percent int) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	mo.gcPercent = percent
	debug.SetGCPercent(percent)
	mo.logger.WithField("percent", percent).Info("Set GC target percentage")
}

// OptimizeForHighThroughput optimizes settings for high-throughput scenarios
func (mo *MemoryOptimizer) OptimizeForHighThroughput() {
	mo.SetGCPercent(300) // Less frequent GC
	mo.gcInterval = 60 * time.Second
	mo.logger.Info("Optimized for high throughput")
}

// OptimizeForLowLatency optimizes settings for low-latency scenarios
func (mo *MemoryOptimizer) OptimizeForLowLatency() {
	mo.SetGCPercent(50) // More frequent GC
	mo.gcInterval = 10 * time.Second
	mo.logger.Info("Optimized for low latency")
}

// Describe implements prometheus.Collector
func (mo *MemoryOptimizer) Describe(ch chan<- *prometheus.Desc) {
	mo.memoryUsage.Describe(ch)
	mo.gcDuration.Describe(ch)
	mo.gcCount.Describe(ch)
	mo.allocationRate.Describe(ch)
}

// Collect implements prometheus.Collector
func (mo *MemoryOptimizer) Collect(ch chan<- prometheus.Metric) {
	mo.UpdateMemoryStats()

	mo.memoryUsage.Collect(ch)
	mo.gcDuration.Collect(ch)
	mo.gcCount.Collect(ch)
	mo.allocationRate.Collect(ch)
}

// MemoryStats represents memory statistics
type MemoryStats struct {
	Alloc         uint64    // Bytes allocated and in use
	TotalAlloc    uint64    // Total bytes allocated (even if freed)
	Sys           uint64    // Bytes obtained from OS
	HeapAlloc     uint64    // Bytes allocated in heap
	HeapSys       uint64    // Bytes obtained from OS for heap
	HeapInuse     uint64    // Bytes in in-use spans
	HeapReleased  uint64    // Bytes released to OS
	GCCPUFraction float64   // Fraction of CPU time used by GC
	NumGC         uint32    // Number of completed GC cycles
	LastGC        time.Time // Time of last GC
}

// String returns a string representation of memory stats
func (ms MemoryStats) String() string {
	return fmt.Sprintf(
		"Alloc=%dMB TotalAlloc=%dMB Sys=%dMB HeapAlloc=%dMB GCCPUFraction=%.2f%% NumGC=%d",
		ms.Alloc/1024/1024,
		ms.TotalAlloc/1024/1024,
		ms.Sys/1024/1024,
		ms.HeapAlloc/1024/1024,
		ms.GCCPUFraction*100,
		ms.NumGC,
	)
}
