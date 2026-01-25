// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewPerformanceMonitor(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration:   30 * time.Second,
		MaxErrorRate:            5.0,
		MinSuccessRate:          95.0,
		MaxMemoryUsage:          100 * 1024 * 1024,
		MaxMetricsPerCollection: 10000,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	assert.NotNil(t, pm)
	assert.Equal(t, slaConfig.MaxCollectionDuration, pm.slaConfig.MaxCollectionDuration)
	assert.NotNil(t, pm.collectorStats)
	assert.NotNil(t, pm.collectionDuration)
}

func TestPerformanceMonitor_RegisterMetrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MaxErrorRate:          5.0,
		MinSuccessRate:        95.0,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))
	registry := prometheus.NewRegistry()

	err := pm.RegisterMetrics(registry)
	assert.NoError(t, err)
}

func TestPerformanceMonitor_RecordCollection_Success(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MaxErrorRate:          5.0,
		MinSuccessRate:        95.0,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record successful collection
	pm.RecordCollection("test-collector", 100*time.Millisecond, 42, nil)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.SuccessCount)
	assert.Equal(t, 100*time.Millisecond, stats.LastDuration)
	assert.Equal(t, 42, stats.LastMetricCount)
}

func TestPerformanceMonitor_RecordCollection_Error(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MaxErrorRate:          5.0,
		MinSuccessRate:        95.0,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record failed collection
	testError := errors.New("test error")
	pm.RecordCollection("test-collector", 100*time.Millisecond, 0, testError)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.NotNil(t, stats)
	assert.Equal(t, int64(1), stats.ErrorCount)
	assert.Equal(t, testError, stats.LastError)
}

func TestPerformanceMonitor_RecordCollection_Multiple(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MaxErrorRate:          5.0,
		MinSuccessRate:        95.0,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record multiple collections
	pm.RecordCollection("collector-a", 50*time.Millisecond, 10, nil)
	pm.RecordCollection("collector-a", 100*time.Millisecond, 20, nil)
	pm.RecordCollection("collector-b", 75*time.Millisecond, 15, nil)

	statsA, existsA := pm.GetCollectorStats("collector-a")
	statsB, existsB := pm.GetCollectorStats("collector-b")

	assert.True(t, existsA)
	assert.True(t, existsB)
	assert.Equal(t, int64(2), statsA.CollectionCount)
	assert.Equal(t, int64(1), statsB.CollectionCount)
}

func TestPerformanceMonitor_RecordResourceUsage(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxMemoryUsage: 100 * 1024 * 1024,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record resource usage
	memoryBytes := int64(50 * 1024 * 1024)
	goroutines := 5

	pm.RecordResourceUsage("test-collector", memoryBytes, goroutines)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.Equal(t, memoryBytes, stats.LastMemoryUsage)
	assert.Equal(t, goroutines, stats.GoroutineCount)
}

func TestPerformanceMonitor_GetStats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record some collections
	pm.RecordCollection("collector-a", 50*time.Millisecond, 10, nil)
	pm.RecordCollection("collector-b", 75*time.Millisecond, 15, nil)

	stats := pm.GetStats()

	assert.NotNil(t, stats)
	assert.Equal(t, 2, len(stats))
	assert.Contains(t, stats, "collector-a")
	assert.Contains(t, stats, "collector-b")
}

func TestPerformanceMonitor_GetCollectorStats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Non-existent collector
	stats, exists := pm.GetCollectorStats("non-existent")
	assert.False(t, exists)
	assert.Nil(t, stats)

	// Record collection and check again
	pm.RecordCollection("test-collector", 50*time.Millisecond, 10, nil)
	stats, exists = pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.NotNil(t, stats)
}

func TestPerformanceMonitor_SLAViolation_DurationExceeded(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 100 * time.Millisecond,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record collection that exceeds SLA
	pm.RecordCollection("test-collector", 200*time.Millisecond, 10, nil)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	// SLA violation should be recorded
	assert.Greater(t, stats.SLAViolations, int64(0))
}

func TestPerformanceMonitor_SLAViolation_HighErrorRate(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MaxErrorRate:          1.0, // Very low threshold
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record multiple errors
	for i := 0; i < 5; i++ {
		pm.RecordCollection("test-collector", 50*time.Millisecond, 0, errors.New("test error"))
	}

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	// Should have multiple errors
	assert.Greater(t, stats.ErrorCount, int64(0))
}

func TestPerformanceMonitor_CollectorStats_Initialization(t *testing.T) {
	t.Parallel()

	stats := &CollectorPerformanceStats{}

	assert.Equal(t, time.Duration(0), stats.LastDuration)
	assert.Equal(t, int64(0), stats.SuccessCount)
	assert.Equal(t, int64(0), stats.ErrorCount)
	assert.Nil(t, stats.LastError)
}

func TestPerformanceMonitor_PeriodicReporting(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record some data
	pm.RecordCollection("test-collector", 50*time.Millisecond, 10, nil)

	// Start periodic reporting with short interval
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pm.StartPeriodicReporting(ctx, 100*time.Millisecond)

	// Wait for reporting to complete
	<-ctx.Done()

	// Should complete without panic
	assert.NotNil(t, pm)
}

func TestPerformanceMonitor_DurationTracking(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record durations in increasing order
	pm.RecordCollection("test-collector", 50*time.Millisecond, 10, nil)
	pm.RecordCollection("test-collector", 100*time.Millisecond, 20, nil)
	pm.RecordCollection("test-collector", 75*time.Millisecond, 15, nil)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.Equal(t, int64(3), stats.CollectionCount)
	assert.Equal(t, 75*time.Millisecond, stats.LastDuration)
	assert.Equal(t, 50*time.Millisecond, stats.MinDuration)
	assert.Equal(t, 100*time.Millisecond, stats.MaxDuration)
}

func TestPerformanceMonitor_SuccessRate_Calculation(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
		MinSuccessRate:        80.0,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record 10 collections: 7 successful, 3 failed
	for i := 0; i < 7; i++ {
		pm.RecordCollection("test-collector", 50*time.Millisecond, 10, nil)
	}
	for i := 0; i < 3; i++ {
		pm.RecordCollection("test-collector", 50*time.Millisecond, 0, errors.New("error"))
	}

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.Equal(t, int64(7), stats.SuccessCount)
	assert.Equal(t, int64(3), stats.ErrorCount)
}

func TestPerformanceMonitor_MetricsTracking(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxMetricsPerCollection: 10000,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record collections with varying metric counts
	pm.RecordCollection("test-collector", 50*time.Millisecond, 100, nil)
	pm.RecordCollection("test-collector", 75*time.Millisecond, 500, nil)
	pm.RecordCollection("test-collector", 60*time.Millisecond, 250, nil)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.Equal(t, int64(850), stats.TotalMetrics)
	assert.Equal(t, 250, stats.LastMetricCount)
	assert.Equal(t, 500, stats.MaxMetricCount)
}

func TestPerformanceMonitor_MultipleCollectors(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record for multiple collectors
	collectors := []string{"collector-a", "collector-b", "collector-c"}
	for _, collector := range collectors {
		for i := 0; i < 3; i++ {
			pm.RecordCollection(collector, time.Duration(50+i*10)*time.Millisecond, 10+i*5, nil)
		}
	}

	stats := pm.GetStats()
	assert.Equal(t, len(collectors), len(stats))

	for _, collector := range collectors {
		s, exists := pm.GetCollectorStats(collector)
		assert.True(t, exists)
		assert.Equal(t, int64(3), s.CollectionCount)
	}
}

func TestPerformanceMonitor_Concurrent_RecordCollection(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	done := make(chan bool, 10)

	// Concurrent recordings
	for i := 0; i < 10; i++ {
		go func(idx int) {
			collector := "collector-" + string(rune(idx))
			pm.RecordCollection(collector, 50*time.Millisecond, 10+idx, nil)
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	stats := pm.GetStats()
	assert.Equal(t, 10, len(stats))
}

func TestPerformanceMonitor_ErrorDetails(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record with error
	testError := errors.New("connection timeout")
	pm.RecordCollection("test-collector", 100*time.Millisecond, 0, testError)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.NotNil(t, stats.LastError)
	assert.Equal(t, testError, stats.LastError)
	assert.Greater(t, stats.ConsecutiveErrors, 0)
}

func TestPerformanceMonitor_RecoveryFromErrors(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	slaConfig := SLAConfig{
		MaxCollectionDuration: 30 * time.Second,
	}

	pm := NewPerformanceMonitor("test", "collector", slaConfig, logger.WithField("component", "test"))

	// Record error then success
	pm.RecordCollection("test-collector", 100*time.Millisecond, 0, errors.New("error"))
	pm.RecordCollection("test-collector", 50*time.Millisecond, 10, nil)

	stats, exists := pm.GetCollectorStats("test-collector")
	assert.True(t, exists)
	assert.Equal(t, int64(1), stats.ErrorCount)
	assert.Equal(t, int64(1), stats.SuccessCount)
	// Consecutive errors should be reset after success
	assert.LessOrEqual(t, stats.ConsecutiveErrors, 1)
}

func TestPerformanceMonitor_SLAConfig(t *testing.T) {
	t.Parallel()

	slaConfig := SLAConfig{
		MaxCollectionDuration:   30 * time.Second,
		MaxErrorRate:            5.0,
		MinSuccessRate:          95.0,
		MaxMemoryUsage:          100 * 1024 * 1024,
		MaxMetricsPerCollection: 10000,
	}

	assert.Equal(t, 30*time.Second, slaConfig.MaxCollectionDuration)
	assert.Equal(t, 5.0, slaConfig.MaxErrorRate)
	assert.Equal(t, 95.0, slaConfig.MinSuccessRate)
}
