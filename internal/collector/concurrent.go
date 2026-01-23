// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ConcurrentCollector manages concurrent collection from multiple collectors
type ConcurrentCollector struct {
	registry       *Registry
	maxConcurrency int
	semaphore      chan struct{}
	wg             sync.WaitGroup

	// Metrics
	activeCollections int64
	totalCollections  int64
	failedCollections int64

	// Collection results
	resultsMu sync.Mutex
	results   map[string]*CollectionResult

	logger *logrus.Entry
}

// CollectionResult holds the result of a collection attempt
type CollectionResult struct {
	CollectorName string
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	MetricCount   int
	Error         error
	Success       bool
}

// NewConcurrentCollector creates a new concurrent collector
func NewConcurrentCollector(registry *Registry, maxConcurrency int) *ConcurrentCollector {
	if maxConcurrency <= 0 {
		maxConcurrency = 5 // Default
	}

	return &ConcurrentCollector{
		registry:       registry,
		maxConcurrency: maxConcurrency,
		semaphore:      make(chan struct{}, maxConcurrency),
		results:        make(map[string]*CollectionResult),
		logger:         logrus.WithField("component", "concurrent_collector"),
	}
}

// CollectAll performs concurrent collection from all enabled collectors
func (cc *ConcurrentCollector) CollectAll(ctx context.Context) ([]*CollectionResult, error) {
	// Get all collectors
	collectors := cc.registry.List()
	if len(collectors) == 0 {
		return nil, nil
	}

	// Reset results
	cc.resultsMu.Lock()
	cc.results = make(map[string]*CollectionResult)
	cc.resultsMu.Unlock()

	// Create result channel
	resultChan := make(chan *CollectionResult, len(collectors))

	// Start collections
	for _, name := range collectors {
		collector, exists := cc.registry.Get(name)
		if !exists || !collector.IsEnabled() {
			continue
		}

		cc.wg.Add(1)
		go cc.collectFromCollector(ctx, name, collector, resultChan)
	}

	// Wait for all collections to complete
	go func() {
		cc.wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []*CollectionResult
	for result := range resultChan {
		results = append(results, result)

		cc.resultsMu.Lock()
		cc.results[result.CollectorName] = result
		cc.resultsMu.Unlock()

		if result.Error != nil {
			atomic.AddInt64(&cc.failedCollections, 1)
		}
	}

	// Check for errors
	var errors []error
	for _, result := range results {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("%s: %w", result.CollectorName, result.Error))
		}
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("collection errors: %v", errors)
	}

	return results, nil
}

// collectFromCollector collects metrics from a single collector
func (cc *ConcurrentCollector) collectFromCollector(
	ctx context.Context,
	name string,
	collector Collector,
	resultChan chan<- *CollectionResult,
) {
	defer cc.wg.Done()

	// Acquire semaphore
	select {
	case cc.semaphore <- struct{}{}:
		defer func() { <-cc.semaphore }()
	case <-ctx.Done():
		resultChan <- &CollectionResult{
			CollectorName: name,
			StartTime:     time.Now(),
			EndTime:       time.Now(),
			Error:         ctx.Err(),
			Success:       false,
		}
		return
	}

	// Track active collections
	atomic.AddInt64(&cc.activeCollections, 1)
	defer atomic.AddInt64(&cc.activeCollections, -1)
	atomic.AddInt64(&cc.totalCollections, 1)

	// Start collection
	result := &CollectionResult{
		CollectorName: name,
		StartTime:     time.Now(),
	}

	// Create metrics channel
	metricChan := make(chan prometheus.Metric, 1000)
	var metrics []prometheus.Metric

	// Collect metrics in background
	var collectionErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		collectionErr = collector.Collect(ctx, metricChan)
		close(metricChan)
	}()

	// Collect all metrics
	for metric := range metricChan {
		metrics = append(metrics, metric)
	}

	// Wait for collection to complete
	<-done

	// Update result
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.MetricCount = len(metrics)
	result.Error = collectionErr
	result.Success = collectionErr == nil

	// Log result
	if result.Success {
		cc.logger.WithFields(logrus.Fields{
			"collector":    name,
			"duration_ms":  result.Duration.Milliseconds(),
			"metric_count": result.MetricCount,
		}).Debug("Collection completed")
	} else {
		cc.logger.WithFields(logrus.Fields{
			"collector": name,
			"error":     result.Error,
		}).Error("Collection failed")
	}

	resultChan <- result
}

// GetActiveCollections returns the number of currently active collections
func (cc *ConcurrentCollector) GetActiveCollections() int64 {
	return atomic.LoadInt64(&cc.activeCollections)
}

// GetTotalCollections returns the total number of collections performed
func (cc *ConcurrentCollector) GetTotalCollections() int64 {
	return atomic.LoadInt64(&cc.totalCollections)
}

// GetFailedCollections returns the total number of failed collections
func (cc *ConcurrentCollector) GetFailedCollections() int64 {
	return atomic.LoadInt64(&cc.failedCollections)
}

// GetLastResults returns the results from the last collection cycle
func (cc *ConcurrentCollector) GetLastResults() map[string]*CollectionResult {
	cc.resultsMu.Lock()
	defer cc.resultsMu.Unlock()

	// Create a copy to avoid race conditions
	results := make(map[string]*CollectionResult)
	for k, v := range cc.results {
		results[k] = v
	}
	return results
}

// CollectionOrchestrator manages scheduled collection cycles
type CollectionOrchestrator struct {
	collector *ConcurrentCollector
	registry  *Registry
	intervals map[string]time.Duration
	timers    map[string]*time.Timer
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *logrus.Entry
	stopped   bool // Protected by mu - prevents new collections after Stop()
}

// NewCollectionOrchestrator creates a new collection orchestrator
func NewCollectionOrchestrator(registry *Registry, maxConcurrency int) *CollectionOrchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	return &CollectionOrchestrator{
		collector: NewConcurrentCollector(registry, maxConcurrency),
		registry:  registry,
		intervals: make(map[string]time.Duration),
		timers:    make(map[string]*time.Timer),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logrus.WithField("component", "collection_orchestrator"),
	}
}

// SetCollectorInterval sets the collection interval for a specific collector
func (co *CollectionOrchestrator) SetCollectorInterval(name string, interval time.Duration) {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.intervals[name] = interval

	// Cancel existing timer if any
	if timer, exists := co.timers[name]; exists {
		timer.Stop()
		delete(co.timers, name)
	}
}

// Start begins the collection orchestration
func (co *CollectionOrchestrator) Start() {
	co.logger.Info("Starting collection orchestrator")

	// Reset stopped flag to allow collections
	co.mu.Lock()
	co.stopped = false
	co.mu.Unlock()

	// Start individual collector timers
	for _, name := range co.registry.List() {
		co.startCollectorTimer(name)
	}
}

// startCollectorTimer starts a timer for a specific collector
func (co *CollectionOrchestrator) startCollectorTimer(name string) {
	co.mu.RLock()
	interval, hasInterval := co.intervals[name]
	co.mu.RUnlock()

	if !hasInterval {
		// Use default interval from configuration
		return
	}

	collector, exists := co.registry.Get(name)
	if !exists || !collector.IsEnabled() {
		return
	}

	// Create timer
	timer := time.NewTimer(interval)

	co.mu.Lock()
	co.timers[name] = timer
	co.mu.Unlock()

	// Start timer goroutine
	go func() {
		for {
			select {
			case <-timer.C:
				// Check if stopped before collecting
				co.mu.RLock()
				isStopped := co.stopped
				co.mu.RUnlock()

				if isStopped {
					timer.Stop()
					return
				}

				// Perform collection
				ctx, cancel := context.WithTimeout(co.ctx, 30*time.Second)

				metricChan := make(chan prometheus.Metric, 1000)
				go func() {
					defer close(metricChan)
					if err := collector.Collect(ctx, metricChan); err != nil {
						co.logger.WithError(err).WithField("collector", name).Error("Scheduled collection failed")
					}
				}()

				// Drain metrics
				count := 0
				for range metricChan {
					count++
				}

				co.logger.WithFields(logrus.Fields{
					"collector":    name,
					"metric_count": count,
				}).Debug("Scheduled collection completed")

				cancel()

				// Check if stopped before resetting timer
				co.mu.RLock()
				isStopped = co.stopped
				co.mu.RUnlock()

				if isStopped {
					timer.Stop()
					return
				}

				// Reset timer only if not stopped
				timer.Reset(interval)

			case <-co.ctx.Done():
				timer.Stop()
				return
			}
		}
	}()
}

// Stop stops the collection orchestrator
func (co *CollectionOrchestrator) Stop() {
	co.logger.Info("Stopping collection orchestrator")

	// Set stopped flag to prevent new collections
	co.mu.Lock()
	co.stopped = true
	co.mu.Unlock()

	co.cancel()

	// Stop all timers
	co.mu.Lock()
	for _, timer := range co.timers {
		timer.Stop()
	}
	co.timers = make(map[string]*time.Timer)
	co.mu.Unlock()
}

// CollectNow triggers an immediate collection for a specific collector
func (co *CollectionOrchestrator) CollectNow(name string) (*CollectionResult, error) {
	collector, exists := co.registry.Get(name)
	if !exists {
		return nil, fmt.Errorf("collector %s not found", name)
	}

	if !collector.IsEnabled() {
		return nil, fmt.Errorf("collector %s is disabled", name)
	}

	resultChan := make(chan *CollectionResult, 1)
	co.collector.wg.Add(1)

	ctx, cancel := context.WithTimeout(co.ctx, 30*time.Second)
	defer cancel()

	go co.collector.collectFromCollector(ctx, name, collector, resultChan)

	result := <-resultChan
	return result, result.Error
}
