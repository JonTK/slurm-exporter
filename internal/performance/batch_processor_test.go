// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockBatchItem implements BatchItem for testing
type mockBatchItem struct {
	key       string
	size      int
	itemType  string
	priority  int
	timestamp time.Time
}

func (m *mockBatchItem) Key() string          { return m.key }
func (m *mockBatchItem) Size() int            { return m.size }
func (m *mockBatchItem) Type() string         { return m.itemType }
func (m *mockBatchItem) Priority() int        { return m.priority }
func (m *mockBatchItem) Timestamp() time.Time { return m.timestamp }

func TestNewBatchProcessor(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())

	tests := []struct {
		name   string
		config BatchConfig
		want   BatchConfig
	}{
		{
			name:   "default values",
			config: BatchConfig{},
			want: BatchConfig{
				MaxBatchSize:   100,
				MaxBatchWait:   5 * time.Second,
				FlushInterval:  30 * time.Second,
				MaxConcurrency: 4,
				RetryAttempts:  3,
				RetryDelay:     time.Second,
			},
		},
		{
			name: "custom values",
			config: BatchConfig{
				MaxBatchSize:  50,
				MaxBatchWait:  2 * time.Second,
				FlushInterval: 10 * time.Second,
			},
			want: BatchConfig{
				MaxBatchSize:   50,
				MaxBatchWait:   2 * time.Second,
				FlushInterval:  10 * time.Second,
				MaxConcurrency: 4,
				RetryAttempts:  3,
				RetryDelay:     time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp, err := NewBatchProcessor(tt.config, logger)
			require.NoError(t, err)
			defer func() { _ = bp.Stop() }()

			assert.Equal(t, tt.want.MaxBatchSize, bp.config.MaxBatchSize)
			assert.Equal(t, tt.want.MaxBatchWait, bp.config.MaxBatchWait)
			assert.Equal(t, tt.want.FlushInterval, bp.config.FlushInterval)
		})
	}
}

func TestBatchProcessor_Add(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10, // Increased to prevent auto-flush when adding 3 items
		MaxBatchWait: 100 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Add items
	items := []BatchItem{
		&mockBatchItem{key: "item1", size: 10, itemType: "test", priority: 1, timestamp: time.Now()},
		&mockBatchItem{key: "item2", size: 20, itemType: "test", priority: 2, timestamp: time.Now()},
		&mockBatchItem{key: "item3", size: 30, itemType: "test", priority: 3, timestamp: time.Now()},
	}

	for _, item := range items {
		err := bp.Add(item)
		assert.NoError(t, err)
	}

	// Check batch state
	bp.mu.RLock()
	batch, exists := bp.batches["test"]
	bp.mu.RUnlock()

	assert.True(t, exists)
	if exists {
		assert.Len(t, batch.items, 3)
		assert.Equal(t, 60, batch.totalSize)
		assert.Equal(t, 3, batch.priority)
	}
}

func TestBatchProcessor_Deduplication(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxBatchWait: 100 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Add duplicate items
	item1 := &mockBatchItem{key: "dup", size: 10, itemType: "test", priority: 1, timestamp: time.Now()}
	item2 := &mockBatchItem{key: "dup", size: 20, itemType: "test", priority: 2, timestamp: time.Now().Add(time.Second)}

	err = bp.Add(item1)
	assert.NoError(t, err)
	err = bp.Add(item2)
	assert.NoError(t, err)

	// Check only one item exists
	bp.mu.RLock()
	batch, exists := bp.batches["test"]
	bp.mu.RUnlock()

	assert.True(t, exists)
	assert.Len(t, batch.items, 1)
}

func TestBatchProcessor_AutoFlush(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 2,
		MaxBatchWait: 50 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	processedCount := int32(0)
	processor := func(ctx context.Context, items []BatchItem) error {
		atomic.AddInt32(&processedCount, int32(len(items)))
		return nil
	}

	// Register processor
	bp.ProcessBatch("test", processor)

	// Add items to trigger size-based flush
	items := []BatchItem{
		&mockBatchItem{key: "item1", size: 10, itemType: "test", priority: 1, timestamp: time.Now()},
		&mockBatchItem{key: "item2", size: 20, itemType: "test", priority: 2, timestamp: time.Now()},
	}

	for _, item := range items {
		err := bp.Add(item)
		assert.NoError(t, err)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Check items were processed
	count := atomic.LoadInt32(&processedCount)
	assert.Equal(t, int32(2), count)
}

func TestBatchProcessor_TimeBasedFlush(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxBatchWait: 50 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	processedCount := int32(0)
	processor := func(ctx context.Context, items []BatchItem) error {
		atomic.AddInt32(&processedCount, int32(len(items)))
		return nil
	}

	// Register processor
	bp.ProcessBatch("test", processor)

	// Add single item
	item := &mockBatchItem{key: "item1", size: 10, itemType: "test", priority: 1, timestamp: time.Now()}
	err = bp.Add(item)
	assert.NoError(t, err)

	// Wait for time-based flush
	time.Sleep(100 * time.Millisecond)

	// Check item was processed
	count := atomic.LoadInt32(&processedCount)
	assert.Equal(t, int32(1), count)
}

func TestBatchProcessor_RetryLogic(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize:  2,
		MaxBatchWait:  50 * time.Millisecond,
		RetryAttempts: 2,
		RetryDelay:    10 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	attemptCount := int32(0)
	processor := func(ctx context.Context, items []BatchItem) error {
		attempt := atomic.AddInt32(&attemptCount, 1)
		if attempt < 3 {
			return errors.New("simulated error")
		}
		return nil
	}

	// Process batch directly
	items := []BatchItem{
		&mockBatchItem{key: "item1", size: 10, itemType: "test", priority: 1, timestamp: time.Now()},
	}

	bp.processBatchWithRetry("test", items, processor, time.Second)

	// Check retry count
	attempts := atomic.LoadInt32(&attemptCount)
	assert.Equal(t, int32(3), attempts)
}

func TestBatchProcessor_ConcurrentAdd(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 100,
		MaxBatchWait: 100 * time.Millisecond,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Concurrent adds
	var wg sync.WaitGroup
	itemCount := 50

	for i := 0; i < itemCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			item := &mockBatchItem{
				key:       string(rune('a' + idx)),
				size:      10,
				itemType:  "test",
				priority:  idx,
				timestamp: time.Now(),
			}
			err := bp.Add(item)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Check all items added
	bp.mu.RLock()
	batch, exists := bp.batches["test"]
	bp.mu.RUnlock()

	assert.True(t, exists)
	assert.Len(t, batch.items, itemCount)
}

func TestBatchProcessor_FlushAll(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxBatchWait: time.Minute, // Long wait
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Add items to multiple batch types
	types := []string{"type1", "type2", "type3"}
	for _, bType := range types {
		for i := 0; i < 3; i++ {
			item := &mockBatchItem{
				key:       string(rune('a' + i)),
				size:      10,
				itemType:  bType,
				priority:  1,
				timestamp: time.Now(),
			}
			err := bp.Add(item)
			assert.NoError(t, err)
		}
	}

	// Verify batches exist
	bp.mu.RLock()
	batchCount := len(bp.batches)
	bp.mu.RUnlock()
	assert.Equal(t, 3, batchCount)

	// Flush all
	bp.FlushAll()

	// Give time for async flush
	time.Sleep(100 * time.Millisecond)

	// Verify batches cleared
	bp.mu.RLock()
	batchCount = len(bp.batches)
	bp.mu.RUnlock()
	assert.Equal(t, 0, batchCount)
}

func TestBatchProcessor_GetStats(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxBatchWait: time.Minute,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Add items
	item1 := &mockBatchItem{key: "item1", size: 100, itemType: "test", priority: 5, timestamp: time.Now()}
	item2 := &mockBatchItem{key: "item2", size: 200, itemType: "test", priority: 10, timestamp: time.Now()}

	_ = bp.Add(item1)
	_ = bp.Add(item2)

	// Get stats
	stats := bp.GetStats()

	assert.Equal(t, 1, stats["batch_count"])

	batchTypes, ok := stats["batch_types"].(map[string]interface{})
	require.True(t, ok, "batch_types should be map[string]interface{}")
	testBatch, ok := batchTypes["test"].(map[string]interface{})
	require.True(t, ok, "test batch should be map[string]interface{}")

	assert.Equal(t, 2, testBatch["items"])
	assert.Equal(t, 300, testBatch["total_size"])
	assert.Equal(t, 10, testBatch["priority"])
}

func TestBatchProcessor_Stop(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{
		MaxBatchSize: 10,
		MaxBatchWait: time.Minute,
	}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)

	// Add items
	for i := 0; i < 5; i++ {
		item := &mockBatchItem{
			key:       string(rune('a' + i)),
			size:      10,
			itemType:  "test",
			priority:  1,
			timestamp: time.Now(),
		}
		_ = bp.Add(item)
	}

	// Stop should flush pending batches
	err = bp.Stop()
	assert.NoError(t, err)

	// Verify clean stop
	select {
	case <-bp.ctx.Done():
	// Context should be canceled
	default:
		t.Error("Context not canceled after stop")
	}
}

func TestBatchProcessor_RegisterMetrics(t *testing.T) {
	t.Parallel()
	logger := logrus.NewEntry(logrus.New())
	config := BatchConfig{}

	bp, err := NewBatchProcessor(config, logger)
	require.NoError(t, err)
	defer func() { _ = bp.Stop() }()

	// Register metrics
	reg := prometheus.NewRegistry()
	err = bp.RegisterMetrics(reg)
	assert.NoError(t, err)

	// Create some metric data so they appear in the registry (Prometheus only
	// shows metrics in Gather() after they have at least one data point)
	bp.metrics.batchesProcessed.WithLabelValues("test", "success").Add(0)
	bp.metrics.itemsProcessed.WithLabelValues("test").Add(0)
	bp.metrics.batchSize.WithLabelValues("test").Observe(1)
	bp.metrics.batchWaitTime.WithLabelValues("test").Observe(0.1)
	bp.metrics.processingDuration.WithLabelValues("test").Observe(0.01)

	// Verify metrics registered
	mfs, err := reg.Gather()
	assert.NoError(t, err)
	assert.True(t, len(mfs) > 0)

	// Check for specific metrics
	metricNames := make(map[string]bool)
	for _, mf := range mfs {
		metricNames[mf.GetName()] = true
	}

	expectedMetrics := []string{
		"slurm_batch_processor_batches_total",
		"slurm_batch_processor_items_total",
		"slurm_batch_processor_batch_size",
		"slurm_batch_processor_wait_time_seconds",
		"slurm_batch_processor_duration_seconds",
	}

	for _, expected := range expectedMetrics {
		assert.True(t, metricNames[expected], "Missing metric: %s", expected)
	}
}
