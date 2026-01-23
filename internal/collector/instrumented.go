// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// InstrumentedCollector provides automatic performance monitoring for collectors
type InstrumentedCollector struct {
	name              string
	collector         Collector
	perfMetrics       *metrics.PerformanceMetrics
	logger            *logrus.Logger
	enabled           bool
	lastCollection    time.Time
	lastDuration      time.Duration
	lastError         error
	consecutiveErrors int
	totalCollections  int64
	totalErrors       int64
	isCollecting      bool
	mu                sync.RWMutex
}

// NewInstrumentedCollector wraps a collector with performance monitoring
func NewInstrumentedCollector(name string, collector Collector, perfMetrics *metrics.PerformanceMetrics, logger *logrus.Logger) *InstrumentedCollector {
	return &InstrumentedCollector{
		name:        name,
		collector:   collector,
		perfMetrics: perfMetrics,
		logger:      logger,
		enabled:     true,
	}
}

// Name returns the collector name
func (ic *InstrumentedCollector) Name() string {
	return ic.name
}

// Describe passes through to the wrapped collector
func (ic *InstrumentedCollector) Describe(ch chan<- *prometheus.Desc) {
	ic.collector.Describe(ch)
}

// Collect collects metrics with automatic instrumentation
func (ic *InstrumentedCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	ic.mu.Lock()
	if ic.isCollecting {
		ic.mu.Unlock()
		ic.logger.WithField("collector", ic.name).Warn("Collection already in progress, skipping")
		return nil
	}
	ic.isCollecting = true
	ic.mu.Unlock()

	defer func() {
		ic.mu.Lock()
		ic.isCollecting = false
		ic.mu.Unlock()
	}()

	// Start performance timing
	timer := ic.perfMetrics.NewCollectionTimer(ic.name, "total")
	startTime := time.Now()

	// Update queue depth (0 since we're now processing)
	ic.perfMetrics.UpdateQueueDepth(ic.name, 0)

	var collectErr error
	defer func() {
		// Stop timer and record duration
		duration := timer.Stop()

		ic.mu.Lock()
		ic.lastCollection = startTime
		ic.lastDuration = duration
		ic.lastError = collectErr
		ic.totalCollections++

		if collectErr != nil {
			ic.consecutiveErrors++
			ic.totalErrors++
			ic.perfMetrics.RecordCollectionError(ic.name, classifyError(collectErr))
			ic.logger.WithFields(logrus.Fields{
				"collector":          ic.name,
				"error":              collectErr,
				"consecutive_errors": ic.consecutiveErrors,
				"duration":           duration,
			}).Error("Collection failed")
		} else {
			ic.consecutiveErrors = 0
			ic.logger.WithFields(logrus.Fields{
				"collector": ic.name,
				"duration":  duration,
			}).Debug("Collection completed successfully")
		}
		ic.mu.Unlock()
	}()

	// Execute the actual collection
	collectErr = ic.collector.Collect(ctx, ch)

	return collectErr
}

// IsEnabled returns whether this collector is enabled
func (ic *InstrumentedCollector) IsEnabled() bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.enabled && ic.collector.IsEnabled()
}

// SetEnabled enables or disables the collector
func (ic *InstrumentedCollector) SetEnabled(enabled bool) {
	ic.mu.Lock()
	ic.enabled = enabled
	ic.mu.Unlock()
	ic.collector.SetEnabled(enabled)
}

// GetState returns the current state of the collector
func (ic *InstrumentedCollector) GetState() CollectorState {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	return CollectorState{
		Name:              ic.name,
		Enabled:           ic.enabled && ic.collector.IsEnabled(),
		LastCollection:    ic.lastCollection,
		LastDuration:      ic.lastDuration,
		LastError:         ic.lastError,
		ConsecutiveErrors: ic.consecutiveErrors,
		TotalCollections:  ic.totalCollections,
		TotalErrors:       ic.totalErrors,
	}
}

// LastCollectionTime returns when this collector last ran
func (ic *InstrumentedCollector) LastCollectionTime() time.Time {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.lastCollection
}

// LastDuration returns how long the last collection took
func (ic *InstrumentedCollector) LastDuration() time.Duration {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.lastDuration
}

// ErrorCount returns the total number of errors
func (ic *InstrumentedCollector) ErrorCount() int64 {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.totalErrors
}

// MetricCount returns the number of metrics produced (if trackable)
func (ic *InstrumentedCollector) MetricCount() int {
	// This would need to be implemented by counting metrics sent through the channel
	// For now, return -1 to indicate not implemented
	return -1
}

// IsCollecting returns whether a collection is currently in progress
func (ic *InstrumentedCollector) IsCollecting() bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.isCollecting
}

// APIInstrumentedCollector extends InstrumentedCollector with API call monitoring
type APIInstrumentedCollector struct {
	*InstrumentedCollector
}

// NewAPIInstrumentedCollector creates an instrumented collector with API monitoring
func NewAPIInstrumentedCollector(name string, collector Collector, perfMetrics *metrics.PerformanceMetrics, logger *logrus.Logger) *APIInstrumentedCollector {
	return &APIInstrumentedCollector{
		InstrumentedCollector: NewInstrumentedCollector(name, collector, perfMetrics, logger),
	}
}

// TimeAPICall provides a helper to time API calls
func (aic *APIInstrumentedCollector) TimeAPICall(endpoint, method string, fn func() error) error {
	timer := aic.perfMetrics.NewAPITimer(endpoint, method)

	err := fn()

	status := "success"
	if err != nil {
		status = "error"
		statusCode := extractStatusCode(err)
		errorType := classifyError(err)
		aic.perfMetrics.RecordAPIError(endpoint, statusCode, errorType)

		aic.logger.WithFields(logrus.Fields{
			"collector":  aic.name,
			"endpoint":   endpoint,
			"method":     method,
			"error":      err,
			"error_type": errorType,
		}).Error("API call failed")
	}

	timer.StopWithStatus(status)
	return err
}

// classifyError classifies errors into types for metrics
func classifyError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()

	// Network errors
	if contains(errStr, "connection refused") || contains(errStr, "no route to host") {
		return "network"
	}

	// Timeout errors
	if contains(errStr, "timeout") || contains(errStr, "deadline exceeded") {
		return "timeout"
	}

	// Authentication errors
	if contains(errStr, "unauthorized") || contains(errStr, "forbidden") || contains(errStr, "401") || contains(errStr, "403") {
		return "auth"
	}

	// Rate limiting
	if contains(errStr, "rate limit") || contains(errStr, "too many requests") || contains(errStr, "429") {
		return "rate_limit"
	}

	// Server errors
	if contains(errStr, "500") || contains(errStr, "502") || contains(errStr, "503") || contains(errStr, "504") {
		return "server"
	}

	// Client errors
	if contains(errStr, "400") || contains(errStr, "404") || contains(errStr, "422") {
		return "client"
	}

	// Parse errors
	if contains(errStr, "json") || contains(errStr, "xml") || contains(errStr, "parse") {
		return "parse"
	}

	// Configuration errors
	if contains(errStr, "config") || contains(errStr, "invalid") {
		return "config"
	}

	return "unknown"
}

// extractStatusCode attempts to extract HTTP status code from error
func extractStatusCode(err error) string {
	if err == nil {
		return "200"
	}

	errStr := err.Error()

	// Common HTTP status codes
	statusCodes := []string{"400", "401", "403", "404", "422", "429", "500", "502", "503", "504"}

	for _, code := range statusCodes {
		if contains(errStr, code) {
			return code
		}
	}

	return "unknown"
}

// contains checks if string contains substring (helper function)
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CollectorRegistry manages instrumented collectors
type CollectorRegistry struct {
	collectors  map[string]*InstrumentedCollector
	perfMetrics *metrics.PerformanceMetrics
	logger      *logrus.Logger
	mu          sync.RWMutex
}

// NewCollectorRegistry creates a new collector registry
func NewCollectorRegistry(perfMetrics *metrics.PerformanceMetrics, logger *logrus.Logger) *CollectorRegistry {
	return &CollectorRegistry{
		collectors:  make(map[string]*InstrumentedCollector),
		perfMetrics: perfMetrics,
		logger:      logger,
	}
}

// Register registers a collector with instrumentation
func (cr *CollectorRegistry) Register(name string, collector Collector) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	instrumented := NewInstrumentedCollector(name, collector, cr.perfMetrics, cr.logger)
	cr.collectors[name] = instrumented
}

// RegisterAPI registers a collector with API monitoring
func (cr *CollectorRegistry) RegisterAPI(name string, collector Collector) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	instrumented := NewAPIInstrumentedCollector(name, collector, cr.perfMetrics, cr.logger)
	cr.collectors[name] = instrumented.InstrumentedCollector
}

// Get returns a collector by name
func (cr *CollectorRegistry) Get(name string) (*InstrumentedCollector, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	collector, exists := cr.collectors[name]
	return collector, exists
}

// List returns all registered collectors
func (cr *CollectorRegistry) List() map[string]*InstrumentedCollector {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	result := make(map[string]*InstrumentedCollector)
	for name, collector := range cr.collectors {
		result[name] = collector
	}
	return result
}

// GetStates returns the state of all collectors
func (cr *CollectorRegistry) GetStates() map[string]CollectorState {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	states := make(map[string]CollectorState)
	for name, collector := range cr.collectors {
		states[name] = collector.GetState()
	}
	return states
}
