// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// BatchProcessor handles efficient batch processing of metrics
type BatchProcessor struct {
	config     BatchConfig
	logger     *logrus.Entry
	metrics    *batchMetrics
	mu         sync.RWMutex
	batches    map[string]*batch
	processors map[string]BatchProcessorFunc
	// TODO: Unused field - preserved for future timer-based flushing
	// flushTimer *time.Timer
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// BatchConfig holds configuration for batch processing
type BatchConfig struct {
	MaxBatchSize      int           `yaml:"max_batch_size"`     // Maximum items per batch
	MaxBatchWait      time.Duration `yaml:"max_batch_wait"`     // Maximum wait time before flush
	MaxConcurrency    int           `yaml:"max_concurrency"`    // Maximum concurrent batch processors
	FlushInterval     time.Duration `yaml:"flush_interval"`     // Regular flush interval
	EnableCompression bool          `yaml:"enable_compression"` // Enable batch compression
	RetryAttempts     int           `yaml:"retry_attempts"`     // Number of retry attempts
	RetryDelay        time.Duration `yaml:"retry_delay"`        // Delay between retries
}

// BatchItem represents a single item in a batch
type BatchItem interface {
	Key() string          // Unique key for deduplication
	Size() int            // Size in bytes for memory tracking
	Type() string         // Type for routing to correct processor
	Priority() int        // Priority for ordering (higher = more important)
	Timestamp() time.Time // Timestamp for age tracking
}

// BatchProcessor function type for processing a batch of items
type BatchProcessorFunc func(ctx context.Context, items []BatchItem) error

// batch represents a collection of items to be processed together
type batch struct {
	items      []BatchItem
	totalSize  int
	createTime time.Time
	lastAdd    time.Time
	priority   int
}

// batchMetrics tracks batch processing metrics
type batchMetrics struct {
	batchesProcessed   *prometheus.CounterVec
	itemsProcessed     *prometheus.CounterVec
	batchSize          *prometheus.HistogramVec
	batchWaitTime      *prometheus.HistogramVec
	processingDuration *prometheus.HistogramVec
	flushTrigger       *prometheus.CounterVec
	errors             *prometheus.CounterVec
	retries            *prometheus.CounterVec
	compressionRatio   *prometheus.GaugeVec
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(config BatchConfig, logger *logrus.Entry) (*BatchProcessor, error) {
	if config.MaxBatchSize <= 0 {
		config.MaxBatchSize = 100
	}
	if config.MaxBatchWait <= 0 {
		config.MaxBatchWait = 5 * time.Second
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 30 * time.Second
	}
	if config.MaxConcurrency <= 0 {
		config.MaxConcurrency = 4
	}
	if config.RetryAttempts <= 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	bp := &BatchProcessor{
		config:     config,
		logger:     logger,
		metrics:    createBatchMetrics(),
		batches:    make(map[string]*batch),
		processors: make(map[string]BatchProcessorFunc),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start flush timer
	bp.startFlushTimer()

	return bp, nil
}

// Add adds an item to the appropriate batch
func (bp *BatchProcessor) Add(item BatchItem) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	batchType := item.Type()
	b, exists := bp.batches[batchType]
	if !exists {
		b = &batch{
			items:      make([]BatchItem, 0, bp.config.MaxBatchSize),
			createTime: time.Now(),
			priority:   item.Priority(),
		}
		bp.batches[batchType] = b
	}

	// Check for duplicates
	for _, existing := range b.items {
		if existing.Key() == item.Key() {
			// Update existing item if newer
			if item.Timestamp().After(existing.Timestamp()) {
				// Replace logic would go here
				return nil
			}
			return nil // Skip duplicate
		}
	}

	// Add item to batch
	b.items = append(b.items, item)
	b.totalSize += item.Size()
	b.lastAdd = time.Now()

	// Update priority if higher
	if item.Priority() > b.priority {
		b.priority = item.Priority()
	}

	// Check if batch should be flushed
	if len(b.items) >= bp.config.MaxBatchSize {
		go bp.flushBatch(batchType)
	}

	return nil
}

// ProcessBatch registers a processor function for a batch type
func (bp *BatchProcessor) ProcessBatch(batchType string, processor BatchProcessorFunc) {
	bp.mu.Lock()
	bp.processors[batchType] = processor
	bp.mu.Unlock()

	bp.wg.Add(1)
	go func() {
		defer bp.wg.Done()

		for {
			select {
			case <-bp.ctx.Done():
				return
			case <-time.After(bp.config.MaxBatchWait):
				bp.checkAndFlushBatch(batchType, processor)
			}
		}
	}()
}

// checkAndFlushBatch checks if a batch needs flushing and processes it
func (bp *BatchProcessor) checkAndFlushBatch(batchType string, processor BatchProcessorFunc) {
	bp.mu.Lock()
	batch, exists := bp.batches[batchType]
	if !exists || len(batch.items) == 0 {
		bp.mu.Unlock()
		return
	}

	// Check if batch should be flushed
	shouldFlush := false
	waitTime := time.Since(batch.createTime)

	if len(batch.items) >= bp.config.MaxBatchSize {
		shouldFlush = true
		bp.metrics.flushTrigger.WithLabelValues(batchType, "size").Inc()
	} else if waitTime >= bp.config.MaxBatchWait {
		shouldFlush = true
		bp.metrics.flushTrigger.WithLabelValues(batchType, "time").Inc()
	}

	if !shouldFlush {
		bp.mu.Unlock()
		return
	}

	// Remove batch from map
	items := batch.items
	delete(bp.batches, batchType)
	bp.mu.Unlock()

	// Process batch if processor is registered
	if processor != nil {
		bp.processBatchWithRetry(batchType, items, processor, waitTime)
	} else {
		bp.logger.WithField("batch_type", batchType).Warn("No processor registered for batch type")
	}
}

// processBatchWithRetry processes a batch with retry logic
func (bp *BatchProcessor) processBatchWithRetry(batchType string, items []BatchItem, processor BatchProcessorFunc, waitTime time.Duration) {
	start := time.Now()

	// Record metrics
	bp.metrics.batchSize.WithLabelValues(batchType).Observe(float64(len(items)))
	bp.metrics.batchWaitTime.WithLabelValues(batchType).Observe(waitTime.Seconds())

	var err error
	for attempt := 0; attempt <= bp.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			bp.metrics.retries.WithLabelValues(batchType).Inc()
			time.Sleep(bp.config.RetryDelay * time.Duration(attempt))
		}

		err = processor(bp.ctx, items)
		if err == nil {
			break
		}

		bp.logger.WithFields(logrus.Fields{
			"batch_type": batchType,
			"attempt":    attempt + 1,
			"error":      err,
			"items":      len(items),
		}).Warn("Batch processing failed, retrying")
	}

	duration := time.Since(start)
	bp.metrics.processingDuration.WithLabelValues(batchType).Observe(duration.Seconds())

	if err != nil {
		bp.metrics.errors.WithLabelValues(batchType, "processing_failed").Inc()
		bp.logger.WithFields(logrus.Fields{
			"batch_type": batchType,
			"error":      err,
			"items":      len(items),
			"duration":   duration,
		}).Error("Batch processing failed after retries")
	} else {
		bp.metrics.batchesProcessed.WithLabelValues(batchType, "success").Inc()
		bp.metrics.itemsProcessed.WithLabelValues(batchType).Add(float64(len(items)))

		bp.logger.WithFields(logrus.Fields{
			"batch_type": batchType,
			"items":      len(items),
			"duration":   duration,
			"wait_time":  waitTime,
		}).Debug("Batch processed successfully")
	}
}

// flushBatch forces a flush of a specific batch type
func (bp *BatchProcessor) flushBatch(batchType string) {
	bp.mu.Lock()
	batch, exists := bp.batches[batchType]
	if !exists || len(batch.items) == 0 {
		bp.mu.Unlock()
		return
	}

	items := batch.items
	waitTime := time.Since(batch.createTime)
	delete(bp.batches, batchType)

	// Get processor
	processor := bp.processors[batchType]
	bp.mu.Unlock()

	bp.logger.WithFields(logrus.Fields{
		"batch_type": batchType,
		"items":      len(items),
		"wait_time":  waitTime,
	}).Debug("Force flushing batch")

	// Process in background if processor exists
	if processor != nil {
		go bp.processBatchWithRetry(batchType, items, processor, waitTime)
	} else {
		bp.logger.WithField("batch_type", batchType).Warn("No processor registered for batch type during flush")
	}
}

// FlushAll flushes all pending batches
func (bp *BatchProcessor) FlushAll() {
	bp.mu.Lock()
	types := make([]string, 0, len(bp.batches))
	for batchType := range bp.batches {
		types = append(types, batchType)
	}
	bp.mu.Unlock()

	for _, batchType := range types {
		bp.flushBatch(batchType)
	}
}

// startFlushTimer starts the periodic flush timer
func (bp *BatchProcessor) startFlushTimer() {
	bp.wg.Add(1)
	go func() {
		defer bp.wg.Done()
		ticker := time.NewTicker(bp.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-bp.ctx.Done():
				return
			case <-ticker.C:
				bp.FlushAll()
			}
		}
	}()
}

// GetStats returns batch processor statistics
func (bp *BatchProcessor) GetStats() map[string]interface{} {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	stats := map[string]interface{}{
		"config":      bp.config,
		"batch_count": len(bp.batches),
		"batch_types": make(map[string]interface{}),
	}

	for batchType, batch := range bp.batches {
		batchTypes, ok := stats["batch_types"].(map[string]interface{})
		if !ok {
			continue
		}
		batchTypes[batchType] = map[string]interface{}{
			"items":      len(batch.items),
			"total_size": batch.totalSize,
			"age":        time.Since(batch.createTime),
			"last_add":   time.Since(batch.lastAdd),
			"priority":   batch.priority,
		}
	}

	return stats
}

// Stop gracefully stops the batch processor
func (bp *BatchProcessor) Stop() error {
	bp.logger.Info("Stopping batch processor")

	// Flush all pending batches
	bp.FlushAll()

	// Cancel context
	bp.cancel()

	// Wait for all goroutines
	done := make(chan struct{})
	go func() {
		bp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		bp.logger.Info("Batch processor stopped")
		return nil
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting for batch processor to stop")
	}
}

// createBatchMetrics creates Prometheus metrics for batch processing
func createBatchMetrics() *batchMetrics {
	return &batchMetrics{
		batchesProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_batch_processor_batches_total",
				Help: "Total number of batches processed",
			},
			[]string{"type", "status"},
		),
		itemsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_batch_processor_items_total",
				Help: "Total number of items processed",
			},
			[]string{"type"},
		),
		batchSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_batch_processor_batch_size",
				Help:    "Size of processed batches",
				Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"type"},
		),
		batchWaitTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_batch_processor_wait_time_seconds",
				Help:    "Time items waited in batch before processing",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"type"},
		),
		processingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_batch_processor_duration_seconds",
				Help:    "Time taken to process a batch",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"type"},
		),
		flushTrigger: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_batch_processor_flush_trigger_total",
				Help: "Number of batch flushes by trigger type",
			},
			[]string{"type", "trigger"},
		),
		errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_batch_processor_errors_total",
				Help: "Total number of batch processing errors",
			},
			[]string{"type", "error"},
		),
		retries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_batch_processor_retries_total",
				Help: "Total number of batch processing retries",
			},
			[]string{"type"},
		),
		compressionRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_batch_processor_compression_ratio",
				Help: "Compression ratio achieved for batches",
			},
			[]string{"type"},
		),
	}
}

// RegisterMetrics registers all batch processor metrics
func (bp *BatchProcessor) RegisterMetrics(reg prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		bp.metrics.batchesProcessed,
		bp.metrics.itemsProcessed,
		bp.metrics.batchSize,
		bp.metrics.batchWaitTime,
		bp.metrics.processingDuration,
		bp.metrics.flushTrigger,
		bp.metrics.errors,
		bp.metrics.retries,
		bp.metrics.compressionRatio,
	}

	for _, collector := range collectors {
		if err := reg.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
