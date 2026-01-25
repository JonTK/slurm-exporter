// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package debug

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/adaptive"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/filtering"
	"github.com/jontk/slurm-exporter/internal/health"
	"github.com/jontk/slurm-exporter/internal/tracing"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockSchedulerDebugger struct {
	enabled bool
	stats   map[string]interface{}
	history []adaptive.ActivityScore
}

func (m *mockSchedulerDebugger) GetStats() map[string]interface{} {
	return m.stats
}

func (m *mockSchedulerDebugger) GetActivityHistory() []adaptive.ActivityScore {
	return m.history
}

func (m *mockSchedulerDebugger) IsEnabled() bool {
	return m.enabled
}

type mockFilterDebugger struct {
	enabled  bool
	stats    map[string]interface{}
	patterns map[string]filtering.MetricPattern
}

func (m *mockFilterDebugger) GetStats() map[string]interface{} {
	return m.stats
}

func (m *mockFilterDebugger) GetPatterns() map[string]filtering.MetricPattern {
	return m.patterns
}

func (m *mockFilterDebugger) IsEnabled() bool {
	return m.enabled
}

type mockTracerDebugger struct {
	enabled bool
	stats   tracing.TracingStats
	config  config.TracingConfig
}

func (m *mockTracerDebugger) GetStats() tracing.TracingStats {
	return m.stats
}

func (m *mockTracerDebugger) IsEnabled() bool {
	return m.enabled
}

func (m *mockTracerDebugger) GetConfig() config.TracingConfig {
	return m.config
}

type mockHealthDebugger struct {
	stats  map[string]interface{}
	checks map[string]health.Check
}

func (m *mockHealthDebugger) GetStats() map[string]interface{} {
	return m.stats
}

func (m *mockHealthDebugger) GetChecks() map[string]health.Check {
	return m.checks
}

type mockCollectorDebugger struct {
	states  map[string]CollectorState
	metrics map[string]CollectorMetrics
}

func (m *mockCollectorDebugger) GetStates() map[string]CollectorState {
	return m.states
}

func (m *mockCollectorDebugger) GetMetrics() map[string]CollectorMetrics {
	return m.metrics
}

func TestNewDebugHandler_Disabled(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: false,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)
	assert.False(t, handler.IsEnabled())
}

func TestNewDebugHandler_Enabled(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled:          true,
		EnabledEndpoints: []string{"collectors", "tracing"},
		RequireAuth:      false,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)
	assert.True(t, handler.IsEnabled())
}

func TestDebugHandler_RegisterComponents(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Create mock components
	scheduler := &mockSchedulerDebugger{enabled: true}
	filter := &mockFilterDebugger{enabled: true}
	tracer := &mockTracerDebugger{enabled: true}
	health := &mockHealthDebugger{}
	collectors := &mockCollectorDebugger{}

	// Register components
	handler.RegisterComponents(scheduler, filter, tracer, health, collectors, nil)

	// Verify components are registered
	stats := handler.GetStats()
	components, ok := stats["components"].(map[string]bool)
	require.True(t, ok, "components should be map[string]bool")
	assert.True(t, components["scheduler"])
	assert.True(t, components["filter"])
	assert.True(t, components["tracer"])
	assert.True(t, components["health"])
	assert.True(t, components["collectors"])
}

func TestDebugHandler_HandleIndex(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Test HTML response
	req := httptest.NewRequest(http.MethodGet, "/debug", nil)
	w := httptest.NewRecorder()

	handler.handleIndex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "SLURM Exporter Debug")

	// Test JSON response
	req = httptest.NewRequest(http.MethodGet, "/debug?format=json", nil)
	w = httptest.NewRecorder()

	handler.handleIndex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "SLURM Exporter Debug", data["Title"])
}

func TestDebugHandler_HandleCollectors(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Register mock collector debugger
	collectors := &mockCollectorDebugger{
		states: map[string]CollectorState{
			"test_collector": {
				Name:              "test_collector",
				Enabled:           true,
				LastCollection:    time.Now(),
				LastDuration:      100 * time.Millisecond,
				ConsecutiveErrors: 0,
				TotalCollections:  10,
				TotalErrors:       1,
				CurrentInterval:   30 * time.Second,
			},
		},
		metrics: map[string]CollectorMetrics{
			"test_collector": {
				MetricsCollected: 100,
				MetricsFiltered:  10,
				AvgDuration:      100 * time.Millisecond,
				SuccessRate:      0.9,
				LastSuccessful:   time.Now(),
			},
		},
	}

	handler.RegisterComponents(nil, nil, nil, nil, collectors, nil)

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/debug/collectors?format=json", nil)
	w := httptest.NewRecorder()

	handler.handleCollectors(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Collector Debug", data["Title"])
	assert.NotNil(t, data["States"])
	assert.NotNil(t, data["Metrics"])
}

func TestDebugHandler_HandleTracing(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Register mock tracer debugger
	tracer := &mockTracerDebugger{
		enabled: true,
		stats: tracing.TracingStats{
			Enabled:    true,
			SampleRate: 0.1,
			Endpoint:   "localhost:4317",
		},
		config: config.TracingConfig{
			Enabled:    true,
			SampleRate: 0.1,
			Endpoint:   "localhost:4317",
			Insecure:   true,
		},
	}

	handler.RegisterComponents(nil, nil, tracer, nil, nil, nil)

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/debug/tracing?format=json", nil)
	w := httptest.NewRecorder()

	handler.handleTracing(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Tracing Debug", data["Title"])
	enabled, ok := data["Enabled"].(bool)
	require.True(t, ok, "Enabled should be bool")
	assert.True(t, enabled)
}

func TestDebugHandler_HandlePatterns(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Register mock filter debugger
	filter := &mockFilterDebugger{
		enabled: true,
		stats: map[string]interface{}{
			"enabled":          true,
			"learning_phase":   false,
			"total_patterns":   2,
			"total_samples":    100,
			"filtered_samples": 10,
		},
		patterns: map[string]filtering.MetricPattern{
			"test_metric": {
				Name:            "test_metric",
				SampleCount:     10,
				Mean:            50.0,
				Variance:        25.0,
				NoiseScore:      0.3,
				ChangeRate:      0.1,
				Correlation:     0.8,
				LastUpdated:     time.Now(),
				FilterRecommend: filtering.ActionKeep,
			},
		},
	}

	handler.RegisterComponents(nil, filter, nil, nil, nil, nil)

	// Test JSON response with filters
	req := httptest.NewRequest(http.MethodGet, "/debug/patterns?format=json&action=keep&min_samples=5", nil)
	w := httptest.NewRecorder()

	handler.handlePatterns(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Smart Filter Patterns", data["Title"])
	enabled, ok := data["Enabled"].(bool)
	require.True(t, ok, "Enabled should be bool")
	assert.True(t, enabled)
}

func TestDebugHandler_HandleScheduler(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Register mock scheduler debugger
	scheduler := &mockSchedulerDebugger{
		enabled: true,
		stats: map[string]interface{}{
			"enabled":               true,
			"uptime_seconds":        3600.0,
			"activity_entries":      10,
			"registered_collectors": 3,
			"current_score":         0.5,
		},
		history: []adaptive.ActivityScore{
			{
				Score:       0.5,
				Timestamp:   time.Now(),
				JobCount:    100,
				NodeCount:   50,
				ChangeRate:  0.1,
				Description: "normal activity",
			},
		},
	}

	handler.RegisterComponents(scheduler, nil, nil, nil, nil, nil)

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/debug/scheduler?format=json", nil)
	w := httptest.NewRecorder()

	handler.handleScheduler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Adaptive Scheduler Debug", data["Title"])
	enabled, ok := data["Enabled"].(bool)
	require.True(t, ok, "Enabled should be bool")
	assert.True(t, enabled)
}

func TestDebugHandler_HandleHealth(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Register mock health debugger
	healthDebugger := &mockHealthDebugger{
		stats: map[string]interface{}{
			"total_checks":     3,
			"healthy_checks":   2,
			"unhealthy_checks": 1,
		},
		checks: map[string]health.Check{
			"slurm_api": {
				Status:      health.StatusHealthy,
				Message:     "SLURM API is responding",
				LastChecked: time.Now(),
			},
			"memory": {
				Status:      health.StatusDegraded,
				Message:     "Memory usage elevated",
				LastChecked: time.Now(),
			},
		},
	}

	handler.RegisterComponents(nil, nil, nil, healthDebugger, nil, nil)

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/debug/health?format=json", nil)
	w := httptest.NewRecorder()

	handler.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Health Debug", data["Title"])
	assert.NotNil(t, data["Checks"])
}

func TestDebugHandler_HandleRuntime(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Test JSON response
	req := httptest.NewRequest(http.MethodGet, "/debug/runtime?format=json", nil)
	w := httptest.NewRecorder()

	handler.handleRuntime(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	assert.NoError(t, err)
	assert.Equal(t, "Runtime Debug", data["Title"])
	assert.NotNil(t, data["Memory"])
	assert.NotNil(t, data["Runtime"])
}

func TestDebugHandler_WithAuth(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled:     true,
		RequireAuth: true,
		Username:    "debug",
		Password:    "secret123",
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	// Test unauthorized access
	req := httptest.NewRequest(http.MethodGet, "/debug", nil)
	w := httptest.NewRecorder()

	authWrappedHandler := handler.withAuth(handler.handleIndex)
	authWrappedHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Equal(t, `Basic realm="Debug Endpoints"`, w.Header().Get("WWW-Authenticate"))

	// Test authorized access
	req = httptest.NewRequest(http.MethodGet, "/debug", nil)
	req.SetBasicAuth("debug", "secret123")
	w = httptest.NewRecorder()

	authWrappedHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDebugHandler_IsEndpointEnabled(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	testCases := []struct {
		name             string
		enabledEndpoints []string
		endpoint         string
		expectedEnabled  bool
	}{
		{
			name:             "all_enabled_by_default",
			enabledEndpoints: nil,
			endpoint:         "collectors",
			expectedEnabled:  true,
		},
		{
			name:             "specific_endpoint_enabled",
			enabledEndpoints: []string{"collectors", "tracing"},
			endpoint:         "collectors",
			expectedEnabled:  true,
		},
		{
			name:             "specific_endpoint_disabled",
			enabledEndpoints: []string{"collectors", "tracing"},
			endpoint:         "health",
			expectedEnabled:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := config.DebugConfig{
				Enabled:          true,
				EnabledEndpoints: tc.enabledEndpoints,
			}

			handler := NewDebugHandler(cfg, logger)
			require.NotNil(t, handler)

			enabled := handler.isEndpointEnabled(tc.endpoint)
			assert.Equal(t, tc.expectedEnabled, enabled)
		})
	}
}

func TestDebugHandler_FilterPatterns(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	patterns := map[string]filtering.MetricPattern{
		"keep_pattern": {
			Name:            "keep_metric",
			SampleCount:     10,
			FilterRecommend: filtering.ActionKeep,
		},
		"filter_pattern": {
			Name:            "filter_metric",
			SampleCount:     5,
			FilterRecommend: filtering.ActionFilter,
		},
		"low_samples": {
			Name:            "low_samples_metric",
			SampleCount:     2,
			FilterRecommend: filtering.ActionKeep,
		},
	}

	// Test filter by action
	filtered := handler.filterPatterns(patterns, "keep", 0)
	assert.Len(t, filtered, 2) // keep_pattern and low_samples

	// Test filter by minimum samples
	filtered = handler.filterPatterns(patterns, "", 5)
	assert.Len(t, filtered, 2) // keep_pattern and filter_pattern

	// Test combined filters
	filtered = handler.filterPatterns(patterns, "keep", 5)
	assert.Len(t, filtered, 1) // only keep_pattern
}

func TestDebugHandler_GetStats(t *testing.T) {
	t.Parallel()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	cfg := config.DebugConfig{
		Enabled: true,
	}

	handler := NewDebugHandler(cfg, logger)
	require.NotNil(t, handler)

	stats := handler.GetStats()
	assert.Equal(t, true, stats["enabled"])
	assert.NotNil(t, stats["uptime"])
	assert.Equal(t, int64(0), stats["request_count"])
	assert.NotNil(t, stats["endpoints"])
	assert.NotNil(t, stats["components"])
}
