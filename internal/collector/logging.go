// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// LoggingCollector wraps a collector with enhanced structured logging
type LoggingCollector struct {
	collector    Collector
	logger       *logrus.Entry
	errorBuilder *ErrorBuilder
	analyzer     *ErrorAnalyzer
	recovery     *ErrorRecoveryHandler
}

// NewLoggingCollector creates a new collector with enhanced logging
func NewLoggingCollector(collector Collector, logger *logrus.Entry) *LoggingCollector {
	collectorLogger := logger.WithField("collector", collector.Name())

	return &LoggingCollector{
		collector:    collector,
		logger:       collectorLogger,
		errorBuilder: NewErrorBuilder(collector.Name(), collectorLogger),
		analyzer:     NewErrorAnalyzer(collectorLogger),
		recovery:     NewErrorRecoveryHandler(collectorLogger),
	}
}

// Name returns the collector name
func (lc *LoggingCollector) Name() string {
	return lc.collector.Name()
}

// Describe forwards to the wrapped collector
func (lc *LoggingCollector) Describe(ch chan<- *prometheus.Desc) {
	lc.logger.Debug("Starting metric description")
	start := time.Now()

	lc.collector.Describe(ch)

	duration := time.Since(start)
	lc.logger.WithField("duration_ms", duration.Milliseconds()).
		Debug("Completed metric description")
}

// Collect performs collection with enhanced logging and error handling
func (lc *LoggingCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	start := time.Now()

	lc.logger.WithFields(logrus.Fields{
		"operation": "collect",
		"enabled":   lc.collector.IsEnabled(),
	}).Debug("Starting metric collection")

	// Check if collector is enabled
	if !lc.collector.IsEnabled() {
		lc.logger.Debug("Collector is disabled, skipping collection")
		return nil
	}

	// Create a channel to count metrics
	metricChan := make(chan prometheus.Metric, 1000)
	var metricCount int

	// Collect metrics in background
	done := make(chan error, 1)
	go func() {
		defer close(metricChan)
		done <- lc.collector.Collect(ctx, metricChan)
	}()

	// Forward metrics and count them
	metricDone := make(chan struct{})
	go func() {
		defer close(metricDone)
		for metric := range metricChan {
			metricCount++
			select {
			case ch <- metric:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for collection to complete
	err := <-done
	// Wait for all metrics to be forwarded
	<-metricDone
	duration := time.Since(start)

	// Log the result
	fields := logrus.Fields{
		"operation":    "collect",
		"duration_ms":  duration.Milliseconds(),
		"metric_count": metricCount,
		"success":      err == nil,
	}

	if err != nil {
		// Analyze and enhance the error
		collErr := lc.analyzer.AnalyzeError(err, lc.collector.Name())

		// Log with structured error information
		lc.errorBuilder.Log(collErr)

		// Attempt recovery
		recoveryErr := lc.recovery.HandleError(ctx, collErr)
		if recoveryErr != nil {
			lc.logger.WithFields(fields).WithError(recoveryErr).
				Error("Collection failed and recovery unsuccessful")
		} else {
			lc.logger.WithFields(fields).
				Info("Collection failed but recovery attempted")
		}

		return collErr
	}
	lc.logger.WithFields(fields).Info("Collection completed successfully")
	return nil
}

// IsEnabled returns whether this collector is enabled
func (lc *LoggingCollector) IsEnabled() bool {
	return lc.collector.IsEnabled()
}

// SetEnabled enables or disables the collector with logging
func (lc *LoggingCollector) SetEnabled(enabled bool) {
	oldState := lc.collector.IsEnabled()
	lc.collector.SetEnabled(enabled)

	lc.logger.WithFields(logrus.Fields{
		"old_state": oldState,
		"new_state": enabled,
		"operation": "set_enabled",
	}).Info("Collector state changed")
}

// CollectorLogger provides centralized logging for collectors
type CollectorLogger struct {
	logger *logrus.Entry
}

// NewCollectorLogger creates a new collector logger
func NewCollectorLogger(logger *logrus.Entry) *CollectorLogger {
	return &CollectorLogger{
		logger: logger.WithField("component", "collector_logger"),
	}
}

// LogCollection logs a collection event
func (cl *CollectorLogger) LogCollection(collector string, operation string, fields logrus.Fields) {
	cl.logger.WithField("collector", collector).
		WithField("operation", operation).
		WithFields(fields).
		Debug("Collection event")
}

// LogError logs an error with context
func (cl *CollectorLogger) LogError(collector string, operation string, err error, fields logrus.Fields) {
	allFields := logrus.Fields{
		"collector": collector,
		"operation": operation,
	}

	// Merge additional fields
	for k, v := range fields {
		allFields[k] = v
	}

	cl.logger.WithFields(allFields).WithError(err).Error("Collector error")
}

// LogMetric logs a metric generation event
func (cl *CollectorLogger) LogMetric(collector string, metricName string, value interface{}, labels map[string]string) {
	fields := logrus.Fields{
		"collector":   collector,
		"metric_name": metricName,
		"value":       value,
		"operation":   "emit_metric",
	}

	// Add labels
	for k, v := range labels {
		fields[k] = v
	}

	cl.logger.WithFields(fields).Trace("Metric emitted")
}

// LogPerformance logs performance metrics
func (cl *CollectorLogger) LogPerformance(collector string, operation string, duration time.Duration, success bool, metricCount int) {
	fields := logrus.Fields{
		"collector":    collector,
		"operation":    operation,
		"duration_ms":  duration.Milliseconds(),
		"success":      success,
		"metric_count": metricCount,
	}

	if success {
		cl.logger.WithFields(fields).Info("Collection performance")
	} else {
		cl.logger.WithFields(fields).Warn("Collection failed")
	}
}

// LogConfigChange logs configuration changes
func (cl *CollectorLogger) LogConfigChange(collector string, field string, oldValue, newValue interface{}) {
	cl.logger.WithFields(logrus.Fields{
		"collector": collector,
		"field":     field,
		"old_value": oldValue,
		"new_value": newValue,
		"operation": "config_change",
	}).Info("Configuration changed")
}

// LogStateChange logs state transitions
func (cl *CollectorLogger) LogStateChange(collector string, oldState, newState string, reason string) {
	cl.logger.WithFields(logrus.Fields{
		"collector": collector,
		"old_state": oldState,
		"new_state": newState,
		"reason":    reason,
		"operation": "state_change",
	}).Info("State transition")
}

// StructuredLogger provides structured logging utilities
type StructuredLogger struct {
	logger *logrus.Entry
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(logger *logrus.Entry) *StructuredLogger {
	return &StructuredLogger{
		logger: logger,
	}
}

// WithCollector adds collector context to logger
func (sl *StructuredLogger) WithCollector(name string) *logrus.Entry {
	return sl.logger.WithField("collector", name)
}

// WithOperation adds operation context to logger
func (sl *StructuredLogger) WithOperation(operation string) *logrus.Entry {
	return sl.logger.WithField("operation", operation)
}

// WithDuration adds duration context to logger
func (sl *StructuredLogger) WithDuration(duration time.Duration) *logrus.Entry {
	return sl.logger.WithField("duration_ms", duration.Milliseconds())
}

// WithError adds error context to logger
func (sl *StructuredLogger) WithError(err error) *logrus.Entry {
	var collErr *CollectionError
	if errors.As(err, &collErr) {
		return sl.logger.WithFields(collErr.LogFields()).WithError(collErr.Err)
	}
	return sl.logger.WithError(err)
}

// WithContext adds context fields to logger
func (sl *StructuredLogger) WithContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	// Extract request ID if available
	if reqID := ctx.Value("request_id"); reqID != nil {
		fields["request_id"] = reqID
	}

	// Extract trace ID if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields["trace_id"] = traceID
	}

	return sl.logger.WithFields(fields)
}

// LogEntry creates a log entry with common fields
func (sl *StructuredLogger) LogEntry(collector, operation string) *logrus.Entry {
	return sl.logger.WithFields(logrus.Fields{
		"collector": collector,
		"operation": operation,
		"timestamp": time.Now(),
	})
}
