// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
	// "github.com/jontk/slurm-exporter/internal/slurm"
)

// BaseCollector provides common functionality for all collectors
type BaseCollector struct {
	// Basic properties
	name    string
	enabled bool
	mu      sync.RWMutex

	// Configuration
	config     *config.CollectorConfig
	globalOpts *CollectorOptions

	// SLURM client
	client interface{} // Will be *slurm.Client when slurm-client build issues are fixed

	// Metrics
	metrics *CollectorMetrics

	// State tracking
	state CollectorState

	// Cardinality management
	cardinalityManager *metrics.CardinalityManager

	// Logger
	logger *logrus.Entry
}

// NewBaseCollector creates a new base collector
func NewBaseCollector(
	name string,
	config *config.CollectorConfig,
	opts *CollectorOptions,
	client interface{}, // Will be *slurm.Client when slurm-client build issues are fixed
	metrics *CollectorMetrics,
	cardinalityManager *metrics.CardinalityManager,
) *BaseCollector {
	logger := logrus.WithField("collector", name)

	return &BaseCollector{
		name:               name,
		enabled:            config.Enabled,
		config:             config,
		globalOpts:         opts,
		client:             client,
		metrics:            metrics,
		cardinalityManager: cardinalityManager,
		logger:             logger,
		state: CollectorState{
			Name:    name,
			Enabled: config.Enabled,
		},
	}
}

// Name returns the collector name
func (b *BaseCollector) Name() string {
	return b.name
}

// IsEnabled returns whether this collector is enabled
func (b *BaseCollector) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.enabled
}

// SetEnabled enables or disables the collector
func (b *BaseCollector) SetEnabled(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.enabled = enabled
	b.state.Enabled = enabled

	if enabled {
		b.logger.Info("Collector enabled")
	} else {
		b.logger.Info("Collector disabled")
	}
}

// GetState returns the current collector state
func (b *BaseCollector) GetState() CollectorState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}

// UpdateState updates the collector state
func (b *BaseCollector) UpdateState(update func(*CollectorState)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	update(&b.state)
}

// CollectWithMetrics wraps collection with common metrics tracking
func (b *BaseCollector) CollectWithMetrics(
	ctx context.Context,
	ch chan<- prometheus.Metric,
	collectFunc func(context.Context, chan<- prometheus.Metric) error,
) error {
	if !b.IsEnabled() {
		return nil
	}

	start := time.Now()
	labels := prometheus.Labels{"collector": b.name}

	// Update metrics
	b.metrics.Total.With(labels).Inc()

	// Create context with timeout
	timeout := b.config.Timeout
	if timeout <= 0 {
		timeout = b.globalOpts.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform collection
	err := collectFunc(ctx, ch)

	// Update duration metric
	duration := time.Since(start)
	b.metrics.Duration.With(labels).Observe(duration.Seconds())

	// Update state
	b.UpdateState(func(s *CollectorState) {
		s.LastCollection = start
		s.LastDuration = duration
		s.TotalCollections++

		if err != nil {
			s.LastError = err
			s.ConsecutiveErrors++
			s.TotalErrors++
			b.metrics.Errors.With(labels).Inc()
			b.metrics.Up.With(labels).Set(0)

			b.logger.WithError(err).Error("Collection failed")
		} else {
			s.LastError = nil
			s.ConsecutiveErrors = 0
			b.metrics.Up.With(labels).Set(1)

			b.logger.WithField("duration", duration).Debug("Collection completed")
		}
	})

	return err
}

// HandleError processes collection errors with retry logic
func (b *BaseCollector) HandleError(err error) error {
	if err == nil {
		return nil
	}

	state := b.GetState()
	errorHandling := b.config.ErrorHandling

	// Check if we should fail fast
	if errorHandling.FailFast {
		return &CollectionError{
			Collector: b.name,
			Err:       err,
			Timestamp: time.Now(),
		}
	}

	// Check if we've exceeded the error threshold from global config
	errorThreshold := 3 // default
	if b.config.ErrorHandling.MaxRetries > 0 {
		errorThreshold = b.config.ErrorHandling.MaxRetries
	}

	if state.ConsecutiveErrors >= errorThreshold {
		b.logger.WithError(err).Error("Collector disabled due to consecutive errors")
		b.SetEnabled(false)
		return fmt.Errorf("collector %s disabled: %w", b.name, err)
	}

	return &CollectionError{
		Collector: b.name,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// ShouldRetry determines if an operation should be retried
func (b *BaseCollector) ShouldRetry(attempt int) bool {
	if !b.IsEnabled() {
		return false
	}

	errorHandling := b.config.ErrorHandling
	return attempt < errorHandling.MaxRetries
}

// GetRetryDelay calculates the delay before the next retry
func (b *BaseCollector) GetRetryDelay(attempt int) time.Duration {
	errorHandling := b.config.ErrorHandling

	// Calculate exponential backoff
	// For attempt 0, use base delay
	// For attempt 1+, multiply by backoff factor
	multiplier := 1.0
	for i := 0; i < attempt; i++ {
		multiplier *= errorHandling.BackoffFactor
	}

	delay := time.Duration(float64(errorHandling.RetryDelay) * multiplier)

	// Cap at max delay
	if delay > errorHandling.MaxRetryDelay {
		delay = errorHandling.MaxRetryDelay
	}

	return delay
}

// WrapError wraps an error with collector context
func (b *BaseCollector) WrapError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("collector %s: %s: %w", b.name, msg, err)
}

// LogCollectionf logs collection events with printf-style formatting
func (b *BaseCollector) LogCollectionf(format string, args ...interface{}) {
	b.logger.Debugf(format, args...)
}

// BuildMetric creates a prometheus metric with common labels
func (b *BaseCollector) BuildMetric(
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	value float64,
	labelValues ...string,
) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)
}

// BuildMetricWithCardinality creates a prometheus metric with cardinality checking
func (b *BaseCollector) BuildMetricWithCardinality(
	metricName string,
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	value float64,
	labels map[string]string,
	labelValues ...string,
) (prometheus.Metric, bool) {
	// Check cardinality limits before creating metric
	if b.cardinalityManager != nil {
		if !b.cardinalityManager.ShouldCollectMetric(metricName, labels) {
			b.logger.WithFields(logrus.Fields{
				"metric": metricName,
				"labels": labels,
			}).Debug("Metric dropped due to cardinality limits")
			return nil, false
		}
	}

	metric := prometheus.MustNewConstMetric(desc, valueType, value, labelValues...)
	return metric, true
}

// SendMetric sends a metric to the channel with error handling
func (b *BaseCollector) SendMetric(ch chan<- prometheus.Metric, metric prometheus.Metric) {
	if metric == nil {
		return // Don't send nil metrics
	}

	select {
	case ch <- metric:
		// Metric sent successfully
	default:
		b.logger.Warn("Metric channel is full, dropping metric")
	}
}

// SendMetricWithCardinality sends a metric with cardinality checking
func (b *BaseCollector) SendMetricWithCardinality(
	ch chan<- prometheus.Metric,
	metricName string,
	desc *prometheus.Desc,
	valueType prometheus.ValueType,
	value float64,
	labels map[string]string,
	labelValues ...string,
) {
	metric, shouldSend := b.BuildMetricWithCardinality(metricName, desc, valueType, value, labels, labelValues...)
	if shouldSend {
		b.SendMetric(ch, metric)
	}
}

// GetCardinalityManager returns the cardinality manager
func (b *BaseCollector) GetCardinalityManager() *metrics.CardinalityManager {
	return b.cardinalityManager
}

// SetCardinalityManager sets the cardinality manager
func (b *BaseCollector) SetCardinalityManager(cm *metrics.CardinalityManager) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cardinalityManager = cm
}

// Describe implements the Collector interface
// Subclasses should override this method to provide their specific descriptors
func (b *BaseCollector) Describe(ch chan<- *prometheus.Desc) {
	// Base collector has no metrics to describe
	// Subclasses will implement their own Describe method
}

// Collect implements the Collector interface
// Subclasses should override this method to provide their specific collection logic
func (b *BaseCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Base collector has no metrics to collect
	// Subclasses will implement their own Collect method using CollectWithMetrics
	return nil
}
