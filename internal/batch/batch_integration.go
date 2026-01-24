// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/performance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// BatchedCollector wraps a collector with batch processing capabilities
type BatchedCollector struct {
	collector  collector.Collector
	batcher    *performance.MetricBatcher
	processor  *performance.BatchProcessor
	logger     *logrus.Entry
	mu         sync.RWMutex
	enabled    bool
	batchItems chan batchItem
}

// batchItem represents an item to be batch processed
type batchItem struct {
	itemType  string
	data      interface{}
	timestamp time.Time
}

func (b *batchItem) Key() string {
	switch b.itemType {
	case "job":
		if job, ok := b.data.(*slurm.Job); ok {
			return fmt.Sprintf("job_%s", job.ID)
		}
	case "node":
		if node, ok := b.data.(*slurm.Node); ok {
			return fmt.Sprintf("node_%s", node.Name)
		}
	case "partition":
		if partition, ok := b.data.(*slurm.Partition); ok {
			return fmt.Sprintf("partition_%s", partition.Name)
		}
	}
	return fmt.Sprintf("%s_%v", b.itemType, b.data)
}

func (b *batchItem) Size() int {
	// Estimate based on type
	switch b.itemType {
	case "job":
		return 1024 // Approximate job size
	case "node":
		return 512
	case "partition":
		return 256
	default:
		return 100
	}
}

func (b *batchItem) Type() string {
	return b.itemType
}

func (b *batchItem) Priority() int {
	// Jobs have highest priority
	switch b.itemType {
	case "job":
		return 10
	case "node":
		return 5
	case "partition":
		return 3
	default:
		return 1
	}
}

func (b *batchItem) Timestamp() time.Time {
	return b.timestamp
}

// NewBatchedCollector creates a new batched collector wrapper
func NewBatchedCollector(
	coll collector.Collector,
	batchConfig performance.BatchConfig,
	logger *logrus.Entry,
) (*BatchedCollector, error) {
	processor, err := performance.NewBatchProcessor(batchConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("creating batch processor: %w", err)
	}

	metricConfig := performance.MetricBatcherConfig{
		BatchConfig:       batchConfig,
		EnableSampling:    true,
		EnableAggregation: true,
		AggregationWindow: 30 * time.Second,
		MaxMetricAge:      5 * time.Minute,
	}

	batcher, err := performance.NewMetricBatcher(metricConfig, nil, logger)
	if err != nil {
		return nil, fmt.Errorf("creating metric batcher: %w", err)
	}

	bc := &BatchedCollector{
		collector:  coll,
		batcher:    batcher,
		processor:  processor,
		logger:     logger,
		enabled:    true,
		batchItems: make(chan batchItem, 1000),
	}

	// Register batch processors
	bc.registerProcessors()

	// Start batch item processor
	go bc.processBatchItems()

	return bc, nil
}

// registerProcessors registers batch processors for different entity types
func (bc *BatchedCollector) registerProcessors() {
	// Job batch processor
	bc.processor.ProcessBatch("job", func(ctx context.Context, items []performance.BatchItem) error {
		jobs := make([]*slurm.Job, 0, len(items))
		for _, item := range items {
			if bi, ok := item.(*batchItem); ok {
				if job, ok := bi.data.(*slurm.Job); ok {
					jobs = append(jobs, job)
				}
			}
		}
		return bc.processJobBatch(ctx, jobs)
	})

	// Node batch processor
	bc.processor.ProcessBatch("node", func(ctx context.Context, items []performance.BatchItem) error {
		nodes := make([]*slurm.Node, 0, len(items))
		for _, item := range items {
			if bi, ok := item.(*batchItem); ok {
				if node, ok := bi.data.(*slurm.Node); ok {
					nodes = append(nodes, node)
				}
			}
		}
		return bc.processNodeBatch(ctx, nodes)
	})

	// Partition batch processor
	bc.processor.ProcessBatch("partition", func(ctx context.Context, items []performance.BatchItem) error {
		partitions := make([]*slurm.Partition, 0, len(items))
		for _, item := range items {
			if bi, ok := item.(*batchItem); ok {
				if partition, ok := bi.data.(*slurm.Partition); ok {
					partitions = append(partitions, partition)
				}
			}
		}
		return bc.processPartitionBatch(ctx, partitions)
	})
}

// processBatchItems processes items from the channel
func (bc *BatchedCollector) processBatchItems() {
	for item := range bc.batchItems {
		if err := bc.processor.Add(&item); err != nil {
			bc.logger.WithError(err).WithField("type", item.itemType).Error("Failed to add item to batch")
		}
	}
}

// processJobBatch processes a batch of jobs
func (bc *BatchedCollector) processJobBatch(ctx context.Context, jobs []*slurm.Job) error {
	_ = ctx
	bc.logger.WithField("count", len(jobs)).Debug("Processing job batch")

	// Group jobs by state for efficient metric creation
	jobsByState := make(map[string][]*slurm.Job)
	for _, job := range jobs {
		jobsByState[job.State] = append(jobsByState[job.State], job)
	}

	// Create metrics for each state group
	for state, stateJobs := range jobsByState {
		// Batch metric creation
		labels := prometheus.Labels{
			"state": state,
		}

		if err := bc.batcher.BatchMetric(
			"slurm_jobs_total",
			labels,
			float64(len(stateJobs)),
			"gauge",
		); err != nil {
			bc.logger.WithError(err).Error("Failed to batch job metric")
		}

		// Calculate aggregate metrics
		var totalCPUs, totalMemory int
		for _, job := range stateJobs {
			totalCPUs += job.CPUs
			totalMemory += job.Memory
		}

		// Batch resource metrics
		_ = bc.batcher.BatchMetric("slurm_jobs_cpus_total", labels, float64(totalCPUs), "gauge")
		_ = bc.batcher.BatchMetric("slurm_jobs_memory_bytes_total", labels, float64(totalMemory)*1024*1024, "gauge")
	}

	return nil
}

// processNodeBatch processes a batch of nodes
func (bc *BatchedCollector) processNodeBatch(ctx context.Context, nodes []*slurm.Node) error {
	_ = ctx
	bc.logger.WithField("count", len(nodes)).Debug("Processing node batch")

	// Group nodes by state
	nodesByState := make(map[string][]*slurm.Node)
	for _, node := range nodes {
		nodesByState[node.State] = append(nodesByState[node.State], node)
	}

	// Create aggregated metrics
	for state, stateNodes := range nodesByState {
		labels := prometheus.Labels{
			"state": state,
		}

		_ = bc.batcher.BatchMetric("slurm_nodes_total", labels, float64(len(stateNodes)), "gauge")

		// Calculate resource totals
		var totalCPUs, totalMemory int
		for _, node := range stateNodes {
			totalCPUs += node.CPUs
			totalMemory += node.Memory
		}

		_ = bc.batcher.BatchMetric("slurm_nodes_cpus_total", labels, float64(totalCPUs), "gauge")
		_ = bc.batcher.BatchMetric("slurm_nodes_memory_bytes_total", labels, float64(totalMemory)*1024*1024, "gauge")
	}

	return nil
}

// processPartitionBatch processes a batch of partitions
func (bc *BatchedCollector) processPartitionBatch(ctx context.Context, partitions []*slurm.Partition) error {
	_ = ctx
	bc.logger.WithField("count", len(partitions)).Debug("Processing partition batch")

	for _, partition := range partitions {
		labels := prometheus.Labels{
			"partition": partition.Name,
			"state":     partition.State,
		}

		// Batch partition metrics
		_ = bc.batcher.BatchMetric("slurm_partition_nodes_total", labels, float64(partition.TotalNodes), "gauge")
		_ = bc.batcher.BatchMetric("slurm_partition_nodes_available", labels, float64(partition.AvailableNodes), "gauge")
		_ = bc.batcher.BatchMetric("slurm_partition_cpus_total", labels, float64(partition.TotalCPUs), "gauge")
		_ = bc.batcher.BatchMetric("slurm_partition_cpus_idle", labels, float64(partition.IdleCPUs), "gauge")
	}

	return nil
}

// QueueJob queues a job for batch processing
func (bc *BatchedCollector) QueueJob(job *slurm.Job) {
	select {
	case bc.batchItems <- batchItem{
		itemType:  "job",
		data:      job,
		timestamp: time.Now(),
	}:
	default:
		bc.logger.Warn("Batch item queue full, dropping job")
	}
}

// QueueNode queues a node for batch processing
func (bc *BatchedCollector) QueueNode(node *slurm.Node) {
	select {
	case bc.batchItems <- batchItem{
		itemType:  "node",
		data:      node,
		timestamp: time.Now(),
	}:
	default:
		bc.logger.Warn("Batch item queue full, dropping node")
	}
}

// QueuePartition queues a partition for batch processing
func (bc *BatchedCollector) QueuePartition(partition *slurm.Partition) {
	select {
	case bc.batchItems <- batchItem{
		itemType:  "partition",
		data:      partition,
		timestamp: time.Now(),
	}:
	default:
		bc.logger.Warn("Batch item queue full, dropping partition")
	}
}

// Describe implements prometheus.Collector
func (bc *BatchedCollector) Describe(ch chan<- *prometheus.Desc) {
	bc.collector.Describe(ch)
}

// Collect implements prometheus.Collector
func (bc *BatchedCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Flush any pending batches
	bc.processor.FlushAll()

	// Collect batched metrics
	metrics, err := bc.batcher.CollectBatchedMetrics()
	if err != nil {
		return err
	}

	// Convert and send metrics
	for _, mf := range metrics {
		for range mf.GetMetric() {
			// This would convert dto metrics to prometheus metrics
			bc.logger.WithField("metric", mf.GetName()).Debug("Collected batched metric")
		}
	}

	// Also collect from underlying collector
	return bc.collector.Collect(ctx, ch)
}

// GetStats returns batch collector statistics
func (bc *BatchedCollector) GetStats() map[string]interface{} {
	bc.mu.RLock()
	queueLen := len(bc.batchItems)
	bc.mu.RUnlock()

	return map[string]interface{}{
		"enabled":         bc.enabled,
		"queue_length":    queueLen,
		"queue_capacity":  cap(bc.batchItems),
		"processor_stats": bc.processor.GetStats(),
		"batcher_stats":   bc.batcher.GetStats(),
	}
}

// Stop stops the batched collector
func (bc *BatchedCollector) Stop() error {
	bc.mu.Lock()
	bc.enabled = false
	bc.mu.Unlock()

	// Close channel to stop processor
	close(bc.batchItems)

	// Stop components
	if err := bc.processor.Stop(); err != nil {
		bc.logger.WithError(err).Error("Error stopping batch processor")
	}

	if err := bc.batcher.Stop(); err != nil {
		bc.logger.WithError(err).Error("Error stopping metric batcher")
	}

	return nil
}
