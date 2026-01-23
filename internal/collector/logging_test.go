// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestLoggingCollector(t *testing.T) {
	// Create test logger with hook to capture logs
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.DebugLevel)
	logEntry := logrus.NewEntry(logger)

	t.Run("SuccessfulCollection", func(t *testing.T) {
		hook.Reset()

		// Create mock collector
		mockCollector := &mockCollector{
			name:    "test_collector",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				metric := prometheus.MustNewConstMetric(
					prometheus.NewDesc("test_metric", "Test metric", nil, nil),
					prometheus.GaugeValue,
					42.0,
				)
				ch <- metric
				return nil
			},
		}

		// Create logging collector
		loggingCollector := NewLoggingCollector(mockCollector, logEntry)

		// Test collection
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := loggingCollector.Collect(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Expected successful collection, got error: %v", err)
		}

		// Check logs
		entries := hook.AllEntries()
		if len(entries) < 2 {
			t.Errorf("Expected at least 2 log entries, got %d", len(entries))
		}

		// Find the completion log
		var completionLog *logrus.Entry
		for _, entry := range entries {
			if entry.Message == "Collection completed successfully" {
				completionLog = entry
				break
			}
		}

		if completionLog == nil {
			t.Error("Expected to find completion log entry")
		} else {
			if completionLog.Data["collector"] != "test_collector" {
				t.Errorf("Expected collector 'test_collector', got '%v'", completionLog.Data["collector"])
			}
			if completionLog.Data["success"] != true {
				t.Errorf("Expected success true, got '%v'", completionLog.Data["success"])
			}
			if completionLog.Data["metric_count"] != 1 {
				t.Errorf("Expected metric_count 1, got '%v'", completionLog.Data["metric_count"])
			}
		}
	})

	t.Run("FailedCollection", func(t *testing.T) {
		hook.Reset()

		// Create failing mock collector
		mockCollector := &mockCollector{
			name:    "failing_collector",
			enabled: true,
			collectFunc: func(ctx context.Context, ch chan<- prometheus.Metric) error {
				return errors.New("connection refused")
			},
		}

		// Create logging collector
		loggingCollector := NewLoggingCollector(mockCollector, logEntry)

		// Test collection
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := loggingCollector.Collect(ctx, metricChan)
		close(metricChan)

		if err == nil {
			t.Error("Expected collection to fail")
		}

		// Should be a CollectionError
		if _, ok := err.(*CollectionError); !ok {
			t.Errorf("Expected CollectionError, got %T", err)
		}

		// Check logs for error entry
		entries := hook.AllEntries()
		errorFound := false
		for _, entry := range entries {
			if entry.Level == logrus.ErrorLevel {
				errorFound = true
				if entry.Data["collector"] != "failing_collector" {
					t.Errorf("Expected collector 'failing_collector', got '%v'", entry.Data["collector"])
				}
				break
			}
		}

		if !errorFound {
			t.Error("Expected to find error log entry")
		}
	})

	t.Run("DisabledCollector", func(t *testing.T) {
		hook.Reset()

		// Create disabled mock collector
		mockCollector := &mockCollector{
			name:    "disabled_collector",
			enabled: false,
		}

		// Create logging collector
		loggingCollector := NewLoggingCollector(mockCollector, logEntry)

		// Test collection
		ctx := context.Background()
		metricChan := make(chan prometheus.Metric, 10)

		err := loggingCollector.Collect(ctx, metricChan)
		close(metricChan)

		if err != nil {
			t.Errorf("Expected no error for disabled collector, got: %v", err)
		}

		// Check for disabled message
		entries := hook.AllEntries()
		disabledFound := false
		for _, entry := range entries {
			if entry.Message == "Collector is disabled, skipping collection" {
				disabledFound = true
				break
			}
		}

		if !disabledFound {
			t.Error("Expected to find disabled collector message")
		}
	})

	t.Run("SetEnabled", func(t *testing.T) {
		hook.Reset()

		mockCollector := &mockCollector{
			name:    "toggle_collector",
			enabled: true,
		}

		loggingCollector := NewLoggingCollector(mockCollector, logEntry)

		// Disable collector
		loggingCollector.SetEnabled(false)

		// Check logs
		entries := hook.AllEntries()
		stateChangeFound := false
		for _, entry := range entries {
			if entry.Message == "Collector state changed" {
				stateChangeFound = true
				if entry.Data["old_state"] != true {
					t.Errorf("Expected old_state true, got '%v'", entry.Data["old_state"])
				}
				if entry.Data["new_state"] != false {
					t.Errorf("Expected new_state false, got '%v'", entry.Data["new_state"])
				}
				break
			}
		}

		if !stateChangeFound {
			t.Error("Expected to find state change log entry")
		}
	})

	t.Run("Describe", func(t *testing.T) {
		hook.Reset()

		mockCollector := &mockCollector{
			name:    "describe_collector",
			enabled: true,
		}

		loggingCollector := NewLoggingCollector(mockCollector, logEntry)

		// Test describe
		descChan := make(chan *prometheus.Desc, 10)
		loggingCollector.Describe(descChan)
		close(descChan)

		// Check logs
		entries := hook.AllEntries()
		if len(entries) < 2 {
			t.Errorf("Expected at least 2 log entries for describe, got %d", len(entries))
		}

		// Should have start and completion messages
		startFound := false
		completeFound := false
		for _, entry := range entries {
			if entry.Message == "Starting metric description" {
				startFound = true
			}
			if entry.Message == "Completed metric description" {
				completeFound = true
			}
		}

		if !startFound {
			t.Error("Expected to find start message")
		}
		if !completeFound {
			t.Error("Expected to find completion message")
		}
	})
}

func TestCollectorLogger(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	logEntry := logrus.NewEntry(logger)

	collectorLogger := NewCollectorLogger(logEntry)

	t.Run("LogCollection", func(t *testing.T) {
		hook.Reset()

		fields := logrus.Fields{
			"duration_ms": 100,
			"success":     true,
		}

		collectorLogger.LogCollection("test_collector", "collect", fields)

		entries := hook.AllEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		entry := entries[0]
		if entry.Data["collector"] != "test_collector" {
			t.Errorf("Expected collector 'test_collector', got '%v'", entry.Data["collector"])
		}
		if entry.Data["operation"] != "collect" {
			t.Errorf("Expected operation 'collect', got '%v'", entry.Data["operation"])
		}
		if entry.Data["duration_ms"] != 100 {
			t.Errorf("Expected duration_ms 100, got '%v'", entry.Data["duration_ms"])
		}
	})

	t.Run("LogError", func(t *testing.T) {
		hook.Reset()

		testErr := errors.New("test error")
		fields := logrus.Fields{
			"endpoint": "/api/jobs",
		}

		collectorLogger.LogError("test_collector", "fetch", testErr, fields)

		entries := hook.AllEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		entry := entries[0]
		if entry.Level != logrus.ErrorLevel {
			t.Errorf("Expected error level, got %v", entry.Level)
		}
		if entry.Data["collector"] != "test_collector" {
			t.Errorf("Expected collector 'test_collector', got '%v'", entry.Data["collector"])
		}
		if entry.Data["endpoint"] != "/api/jobs" {
			t.Errorf("Expected endpoint '/api/jobs', got '%v'", entry.Data["endpoint"])
		}
	})

	t.Run("LogMetric", func(t *testing.T) {
		hook.Reset()

		labels := map[string]string{
			"job_id": "12345",
			"state":  "running",
		}

		collectorLogger.LogMetric("test_collector", "job_count", 42, labels)

		entries := hook.AllEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		entry := entries[0]
		if entry.Level != logrus.TraceLevel {
			t.Errorf("Expected trace level, got %v", entry.Level)
		}
		if entry.Data["metric_name"] != "job_count" {
			t.Errorf("Expected metric_name 'job_count', got '%v'", entry.Data["metric_name"])
		}
		if entry.Data["value"] != 42 {
			t.Errorf("Expected value 42, got '%v'", entry.Data["value"])
		}
		if entry.Data["job_id"] != "12345" {
			t.Errorf("Expected job_id '12345', got '%v'", entry.Data["job_id"])
		}
	})

	t.Run("LogPerformance", func(t *testing.T) {
		hook.Reset()

		duration := 150 * time.Millisecond
		collectorLogger.LogPerformance("test_collector", "collect", duration, true, 5)

		entries := hook.AllEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		entry := entries[0]
		if entry.Level != logrus.InfoLevel {
			t.Errorf("Expected info level, got %v", entry.Level)
		}
		if entry.Data["duration_ms"] != int64(150) {
			t.Errorf("Expected duration_ms 150, got '%v'", entry.Data["duration_ms"])
		}
		if entry.Data["metric_count"] != 5 {
			t.Errorf("Expected metric_count 5, got '%v'", entry.Data["metric_count"])
		}
		if entry.Data["success"] != true {
			t.Errorf("Expected success true, got '%v'", entry.Data["success"])
		}
	})
}

func TestStructuredLogger(t *testing.T) {
	logger, hook := test.NewNullLogger()
	logEntry := logrus.NewEntry(logger)

	structuredLogger := NewStructuredLogger(logEntry)

	t.Run("WithCollector", func(t *testing.T) {
		entry := structuredLogger.WithCollector("test_collector")
		if entry.Data["collector"] != "test_collector" {
			t.Errorf("Expected collector 'test_collector', got '%v'", entry.Data["collector"])
		}
	})

	t.Run("WithOperation", func(t *testing.T) {
		entry := structuredLogger.WithOperation("collect")
		if entry.Data["operation"] != "collect" {
			t.Errorf("Expected operation 'collect', got '%v'", entry.Data["operation"])
		}
	})

	t.Run("WithDuration", func(t *testing.T) {
		duration := 100 * time.Millisecond
		entry := structuredLogger.WithDuration(duration)
		if entry.Data["duration_ms"] != int64(100) {
			t.Errorf("Expected duration_ms 100, got '%v'", entry.Data["duration_ms"])
		}
	})

	t.Run("WithError", func(t *testing.T) {
		// Test with regular error
		regularErr := errors.New("regular error")
		entry := structuredLogger.WithError(regularErr)
		if entry.Data["error"] != regularErr {
			t.Errorf("Expected error to be set")
		}

		// Test with CollectionError
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeAPI,
			Severity:  SeverityHigh,
			Message:   "API error",
			Timestamp: time.Now(),
		}
		entry = structuredLogger.WithError(collErr)
		if entry.Data["collector"] != "test" {
			t.Errorf("Expected collector 'test', got '%v'", entry.Data["collector"])
		}
		if entry.Data["error_type"] != ErrorTypeAPI {
			t.Errorf("Expected error_type '%s', got '%v'", ErrorTypeAPI, entry.Data["error_type"])
		}
	})

	t.Run("WithContext", func(t *testing.T) {
		// Test with context containing request ID
		//lint:ignore SA1029 test validates logger extracts string keys from context
		ctx := context.WithValue(context.Background(), "request_id", "req-123")
		entry := structuredLogger.WithContext(ctx)
		if entry.Data["request_id"] != "req-123" {
			t.Errorf("Expected request_id 'req-123', got '%v'", entry.Data["request_id"])
		}

		// Test with context containing trace ID
		//lint:ignore SA1029 test validates logger extracts string keys from context
		ctx = context.WithValue(ctx, "trace_id", "trace-456")
		entry = structuredLogger.WithContext(ctx)
		if entry.Data["trace_id"] != "trace-456" {
			t.Errorf("Expected trace_id 'trace-456', got '%v'", entry.Data["trace_id"])
		}
	})

	t.Run("LogEntry", func(t *testing.T) {
		hook.Reset()

		entry := structuredLogger.LogEntry("test_collector", "collect")
		entry.Info("Test message")

		entries := hook.AllEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 log entry, got %d", len(entries))
		}

		logEntry := entries[0]
		if logEntry.Data["collector"] != "test_collector" {
			t.Errorf("Expected collector 'test_collector', got '%v'", logEntry.Data["collector"])
		}
		if logEntry.Data["operation"] != "collect" {
			t.Errorf("Expected operation 'collect', got '%v'", logEntry.Data["operation"])
		}
		if _, exists := logEntry.Data["timestamp"]; !exists {
			t.Error("Expected timestamp to be set")
		}
	})
}
