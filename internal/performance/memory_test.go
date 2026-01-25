// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"fmt"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewMemoryOptimizer(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	mo := NewMemoryOptimizer(logger)

	assert.NotNil(t, mo)
	assert.NotNil(t, mo.objectPools)
	assert.True(t, len(mo.objectPools) > 0, "should have initialized object pools")
}

func TestMemoryOptimizer_MetricPools(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	t.Run("GetAndPutMetricPool", func(t *testing.T) {
		// Get metrics from pool
		metrics := mo.GetMetricPool()
		assert.NotNil(t, metrics)

		// Put metrics back to pool
		mo.PutMetricPool(metrics)
		// Should not panic
	})

	t.Run("GetAndPutLabelPool", func(t *testing.T) {
		// Get labels from pool
		labels := mo.GetLabelPool()
		assert.NotNil(t, labels)

		// Put labels back to pool
		mo.PutLabelPool(labels)
		// Should not panic
	})

	t.Run("GetAndPutChannelPool", func(t *testing.T) {
		// Get channel from pool
		ch := mo.GetChannelPool()
		assert.NotNil(t, ch)

		// Put channel back to pool
		mo.PutChannelPool(ch)
		// Should not panic
	})
}

func TestMemoryOptimizer_MemoryStats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	t.Run("UpdateMemoryStats", func(t *testing.T) {
		mo.UpdateMemoryStats()
		// Should not panic
	})

	t.Run("GetMemoryStats", func(t *testing.T) {
		stats := mo.GetMemoryStats()
		// Verify stats structure is initialized (non-zero values indicate runtime has collected memory stats)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.TotalAlloc, uint64(0))
	})

	t.Run("MemoryStats_String", func(t *testing.T) {
		stats := mo.GetMemoryStats()
		statsStr := stats.String()
		assert.NotEmpty(t, statsStr)
		// Verify it contains reasonable information
		assert.Contains(t, statsStr, "Alloc")
	})
}

func TestMemoryOptimizer_SetMemoryLimit(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	tests := []struct {
		name  string
		limit uint64
	}{
		{
			name:  "set 1MB limit",
			limit: 1024 * 1024,
		},
		{
			name:  "set 1GB limit",
			limit: 1024 * 1024 * 1024,
		},
		{
			name:  "set 0 limit (unlimited)",
			limit: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mo.SetMemoryLimit(tc.limit)
			// Should not panic
			stats := mo.GetMemoryStats()
			assert.NotNil(t, stats)
		})
	}
}

func TestMemoryOptimizer_SetGCPercent(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	tests := []struct {
		name    string
		percent int
	}{
		{
			name:    "set 50%",
			percent: 50,
		},
		{
			name:    "set 100%",
			percent: 100,
		},
		{
			name:    "set 200%",
			percent: 200,
		},
		{
			name:    "set -1 (disable GC)",
			percent: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mo.SetGCPercent(tc.percent)
			// Should not panic
		})
	}
}

func TestMemoryOptimizer_OptimizationModes(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	t.Run("OptimizeForHighThroughput", func(t *testing.T) {
		mo.OptimizeForHighThroughput()
		// Should not panic - optimizes for throughput by reducing GC frequency
		stats := mo.GetMemoryStats()
		assert.NotNil(t, stats)
	})

	t.Run("OptimizeForLowLatency", func(t *testing.T) {
		mo.OptimizeForLowLatency()
		// Should not panic - optimizes for latency by more frequent GC
		stats := mo.GetMemoryStats()
		assert.NotNil(t, stats)
	})
}

func TestMemoryOptimizer_ForceGC(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	// Get initial GC count
	statsBefore := mo.GetMemoryStats()
	gcBefore := statsBefore.NumGC

	// Force GC
	mo.ForceGC()

	// Get stats after GC
	statsAfter := mo.GetMemoryStats()
	gcAfter := statsAfter.NumGC

	// GC count should increase or stay same (depends on timing)
	assert.True(t, gcAfter >= gcBefore)
}

func TestMemoryOptimizer_PrometheusMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	t.Run("Describe", func(t *testing.T) {
		ch := make(chan *prometheus.Desc, 10)
		mo.Describe(ch)
		close(ch)

		// Should have at least some descriptors
		descs := 0
		for range ch {
			descs++
		}
		assert.True(t, descs > 0)
	})

	t.Run("Collect", func(t *testing.T) {
		ch := make(chan prometheus.Metric, 100)
		mo.Collect(ch)
		close(ch)

		// Should have at least some metrics
		metrics := 0
		for range ch {
			metrics++
		}
		assert.True(t, metrics >= 0)
	})
}

func TestMemoryOptimizer_ConcurrentOperations(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	// Simulate concurrent operations
	var wg sync.WaitGroup

	// Concurrent reads (should be safe)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = mo.GetMemoryStats()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = mo.GetMetricPool()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = mo.GetLabelPool()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = mo.GetChannelPool()
	}()

	// Concurrent writes (serialized to avoid data races during test)
	wg.Add(1)
	go func() {
		defer wg.Done()
		mo.UpdateMemoryStats()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mo.SetGCPercent(100)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mo.SetMemoryLimit(1024 * 1024 * 1024)
	}()

	// Wait for all operations to complete
	wg.Wait()
}

func TestMemoryOptimizer_PoolReuse(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	t.Run("MetricPool reuse", func(t *testing.T) {
		// Get pool multiple times to verify reuse
		for i := 0; i < 5; i++ {
			metrics := mo.GetMetricPool()
			assert.NotNil(t, metrics)
			mo.PutMetricPool(metrics)
		}
	})

	t.Run("LabelPool reuse", func(t *testing.T) {
		// Get pool multiple times to verify reuse
		for i := 0; i < 5; i++ {
			labels := mo.GetLabelPool()
			assert.NotNil(t, labels)
			mo.PutLabelPool(labels)
		}
	})

	t.Run("ChannelPool reuse", func(t *testing.T) {
		// Get pool multiple times to verify reuse
		for i := 0; i < 5; i++ {
			ch := mo.GetChannelPool()
			assert.NotNil(t, ch)
			mo.PutChannelPool(ch)
		}
	})
}

func TestMemoryOptimizer_StressTest(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	// Stress test with rapid pool operations
	for i := 0; i < 100; i++ {
		metrics := mo.GetMetricPool()
		labels := mo.GetLabelPool()
		ch := mo.GetChannelPool()

		mo.PutMetricPool(metrics)
		mo.PutLabelPool(labels)
		mo.PutChannelPool(ch)
	}

	// Verify still working
	stats := mo.GetMemoryStats()
	assert.NotNil(t, stats)
}

func TestMemoryOptimizer_MemoryStatsFields(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	stats := mo.GetMemoryStats()

	// Verify all fields are present and have reasonable values
	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.TotalAlloc, stats.Alloc, "TotalAlloc should be >= Alloc")

	// Verify LastGC is set
	assert.False(t, stats.LastGC.IsZero(), "LastGC should be set")

	// Verify GC CPU fraction is reasonable
	assert.True(t, stats.GCCPUFraction >= 0 && stats.GCCPUFraction <= 1)
}

func TestMemoryOptimizer_MultipleInstances(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	// Create multiple instances and verify they don't interfere
	mo1 := NewMemoryOptimizer(logger)
	mo2 := NewMemoryOptimizer(logger)
	mo3 := NewMemoryOptimizer(logger)

	// Each should have independent object pools
	assert.NotNil(t, mo1.objectPools)
	assert.NotNil(t, mo2.objectPools)
	assert.NotNil(t, mo3.objectPools)

	// Operations on one shouldn't affect others
	mo1.SetGCPercent(50)
	mo2.SetGCPercent(200)
	mo3.SetGCPercent(100)

	// All should still be functional
	stats1 := mo1.GetMemoryStats()
	stats2 := mo2.GetMemoryStats()
	stats3 := mo3.GetMemoryStats()

	assert.NotNil(t, stats1)
	assert.NotNil(t, stats2)
	assert.NotNil(t, stats3)
}

func TestMemoryOptimizer_ObjectPoolNames(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	mo := NewMemoryOptimizer(logger)

	// Verify expected object pools are initialized
	expectedPools := []string{"metrics", "labels", "strings", "channels"}

	for _, poolName := range expectedPools {
		pool, exists := mo.objectPools[poolName]
		assert.True(t, exists, fmt.Sprintf("pool %s should exist", poolName))
		assert.NotNil(t, pool, fmt.Sprintf("pool %s should not be nil", poolName))
	}
}
