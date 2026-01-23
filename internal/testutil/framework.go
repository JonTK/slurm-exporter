// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BaseTestSuite provides common testing utilities for all test suites
type BaseTestSuite struct {
	suite.Suite
	ctrl    *gomock.Controller
	metrics *TestMetricsRegistry
	clock   *TestClock
	logger  *TestLogger
	ctx     context.Context
	cancel  context.CancelFunc
}

// SetupTest initializes common test resources
func (s *BaseTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.metrics = NewTestMetricsRegistry()
	s.clock = NewTestClock()
	s.logger = NewTestLogger()
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 30*time.Second)
}

// TearDownTest cleans up test resources
func (s *BaseTestSuite) TearDownTest() {
	if s.ctrl != nil {
		s.ctrl.Finish()
	}
	if s.metrics != nil {
		s.metrics.Reset()
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// TestMetricsRegistry provides a test-friendly metrics registry
type TestMetricsRegistry struct {
	mu      sync.RWMutex
	metrics map[string]*TestMetric
	reg     *prometheus.Registry
}

// TestMetric represents a metric in testing
type TestMetric struct {
	Name   string
	Type   string
	Value  float64
	Labels map[string]string
	Help   string
}

// NewTestMetricsRegistry creates a new test metrics registry
func NewTestMetricsRegistry() *TestMetricsRegistry {
	return &TestMetricsRegistry{
		metrics: make(map[string]*TestMetric),
		reg:     prometheus.NewRegistry(),
	}
}

// Register registers a metric
func (tmr *TestMetricsRegistry) Register(c prometheus.Collector) error {
	return tmr.reg.Register(c)
}

// MustRegister registers a metric and panics on error
func (tmr *TestMetricsRegistry) MustRegister(c prometheus.Collector) {
	tmr.reg.MustRegister(c)
}

// Unregister removes a metric
func (tmr *TestMetricsRegistry) Unregister(c prometheus.Collector) bool {
	return tmr.reg.Unregister(c)
}

// Gather gathers all metrics
func (tmr *TestMetricsRegistry) Gather() ([]*dto.MetricFamily, error) {
	return tmr.reg.Gather()
}

// Reset clears all metrics
func (tmr *TestMetricsRegistry) Reset() {
	tmr.mu.Lock()
	defer tmr.mu.Unlock()
	tmr.metrics = make(map[string]*TestMetric)
	tmr.reg = prometheus.NewRegistry()
}

// RecordMetric records a metric value for testing
func (tmr *TestMetricsRegistry) RecordMetric(name string, value float64, labels map[string]string) {
	tmr.mu.Lock()
	defer tmr.mu.Unlock()

	tmr.metrics[name] = &TestMetric{
		Name:   name,
		Value:  value,
		Labels: labels,
	}
}

// GetMetric retrieves a metric by name
func (tmr *TestMetricsRegistry) GetMetric(name string) *TestMetric {
	tmr.mu.RLock()
	defer tmr.mu.RUnlock()
	return tmr.metrics[name]
}

// GetMetricValue gets the value of a metric
func (tmr *TestMetricsRegistry) GetMetricValue(name string) float64 {
	if metric := tmr.GetMetric(name); metric != nil {
		return metric.Value
	}
	return 0
}

// HasMetric checks if a metric exists
func (tmr *TestMetricsRegistry) HasMetric(name string) bool {
	tmr.mu.RLock()
	defer tmr.mu.RUnlock()
	_, exists := tmr.metrics[name]
	return exists
}

// TestClock provides a controllable clock for testing
type TestClock struct {
	mu   sync.RWMutex
	time time.Time
}

// NewTestClock creates a new test clock
func NewTestClock() *TestClock {
	return &TestClock{
		time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Now returns the current test time
func (tc *TestClock) Now() time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.time
}

// Set sets the test time
func (tc *TestClock) Set(t time.Time) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.time = t
}

// Add advances the test time
func (tc *TestClock) Add(d time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.time = tc.time.Add(d)
}

// TestLogger provides a test logger that captures log entries
type TestLogger struct {
	mu      sync.RWMutex
	entries []TestLogEntry
	logger  *logrus.Logger
}

// TestLogEntry represents a log entry in testing
type TestLogEntry struct {
	Level   logrus.Level
	Message string
	Fields  logrus.Fields
	Time    time.Time
}

// NewTestLogger creates a new test logger
func NewTestLogger() *TestLogger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	tl := &TestLogger{
		entries: make([]TestLogEntry, 0),
		logger:  logger,
	}

	// Add hook to capture log entries
	logger.AddHook(&testLogHook{logger: tl})

	return tl
}

// GetLogger returns the underlying logrus logger
func (tl *TestLogger) GetLogger() *logrus.Logger {
	return tl.logger
}

// GetEntry returns a logger entry
func (tl *TestLogger) GetEntry() *logrus.Entry {
	return logrus.NewEntry(tl.logger)
}

// GetEntries returns all captured log entries
func (tl *TestLogger) GetEntries() []TestLogEntry {
	tl.mu.RLock()
	defer tl.mu.RUnlock()

	entries := make([]TestLogEntry, len(tl.entries))
	copy(entries, tl.entries)
	return entries
}

// GetEntriesWithLevel returns log entries for a specific level
func (tl *TestLogger) GetEntriesWithLevel(level logrus.Level) []TestLogEntry {
	entries := tl.GetEntries()
	var filtered []TestLogEntry

	for _, entry := range entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// HasLogWithMessage checks if any log entry contains the message
func (tl *TestLogger) HasLogWithMessage(message string) bool {
	entries := tl.GetEntries()
	for _, entry := range entries {
		if entry.Message == message {
			return true
		}
	}
	return false
}

// HasLogWithLevel checks if any log entry has the specified level
func (tl *TestLogger) HasLogWithLevel(level logrus.Level) bool {
	entries := tl.GetEntries()
	for _, entry := range entries {
		if entry.Level == level {
			return true
		}
	}
	return false
}

// Reset clears all captured log entries
func (tl *TestLogger) Reset() {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.entries = tl.entries[:0]
}

// addEntry adds a log entry (used by hook)
func (tl *TestLogger) addEntry(entry TestLogEntry) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.entries = append(tl.entries, entry)
}

// testLogHook captures log entries for testing
type testLogHook struct {
	logger *TestLogger
}

func (h *testLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *testLogHook) Fire(entry *logrus.Entry) error {
	fields := make(logrus.Fields)
	for k, v := range entry.Data {
		fields[k] = v
	}

	h.logger.addEntry(TestLogEntry{
		Level:   entry.Level,
		Message: entry.Message,
		Fields:  fields,
		Time:    entry.Time,
	})

	return nil
}

// TestHelpers provides utility functions for testing
type TestHelpers struct{}

// AssertMetricValue asserts that a metric has the expected value
func (h TestHelpers) AssertMetricValue(t *testing.T, registry *TestMetricsRegistry, name string, expected float64) {
	actual := registry.GetMetricValue(name)
	assert.InDelta(t, expected, actual, 0.001,
		"Metric %s: expected %f, got %f", name, expected, actual)
}

// AssertMetricExists asserts that a metric exists
func (h TestHelpers) AssertMetricExists(t *testing.T, registry *TestMetricsRegistry, name string) {
	assert.True(t, registry.HasMetric(name), "Metric %s should exist", name)
}

// AssertMetricLabels asserts that a metric has expected labels
func (h TestHelpers) AssertMetricLabels(t *testing.T, registry *TestMetricsRegistry, name string, expectedLabels map[string]string) {
	metric := registry.GetMetric(name)
	require.NotNil(t, metric, "Metric %s should exist", name)

	for k, v := range expectedLabels {
		assert.Equal(t, v, metric.Labels[k],
			"Label %s mismatch for metric %s: expected %s, got %s", k, name, v, metric.Labels[k])
	}
}

// AssertLogMessage asserts that a log message was recorded
func (h TestHelpers) AssertLogMessage(t *testing.T, logger *TestLogger, message string) {
	assert.True(t, logger.HasLogWithMessage(message),
		"Expected log message not found: %s", message)
}

// AssertLogLevel asserts that a log entry with the specified level was recorded
func (h TestHelpers) AssertLogLevel(t *testing.T, logger *TestLogger, level logrus.Level) {
	assert.True(t, logger.HasLogWithLevel(level),
		"Expected log level not found: %s", level)
}

// AssertNoErrors asserts that no error logs were recorded
func (h TestHelpers) AssertNoErrors(t *testing.T, logger *TestLogger) {
	errorEntries := logger.GetEntriesWithLevel(logrus.ErrorLevel)
	assert.Empty(t, errorEntries, "Unexpected error logs: %+v", errorEntries)
}

// WaitForCondition waits for a condition to be true with timeout
func (h TestHelpers) WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			t.Fatalf("Condition not met within timeout: %s", message)
		}
	}
}

// Eventually polls a condition until it's true or timeout
func (h TestHelpers) Eventually(t *testing.T, condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}

// CreateTempConfig creates a temporary configuration for testing
func (h TestHelpers) CreateTempConfig(t *testing.T, config string) string {
	tempFile := fmt.Sprintf("/tmp/test-config-%d.yaml", time.Now().UnixNano())

	err := WriteFile(tempFile, []byte(config))
	require.NoError(t, err)

	// Cleanup on test completion
	t.Cleanup(func() {
		_ = RemoveFile(tempFile)
	})

	return tempFile
}

// Global test helpers instance
var Helpers = TestHelpers{}
