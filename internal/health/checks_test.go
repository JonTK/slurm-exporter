package health

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock SLURM client for testing
type MockSLURMClient struct {
	mock.Mock
}

func (m *MockSLURMClient) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSLURMClient) GetInfo(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

// Mock collector registry for testing
type MockCollectorRegistry struct {
	states map[string]CollectorState
}

func (m *MockCollectorRegistry) GetStates() map[string]CollectorState {
	return m.states
}

func TestSLURMAPIHealthCheck_Healthy(t *testing.T) {
	client := new(MockSLURMClient)
	client.On("Ping", mock.Anything).Return(nil)
	client.On("GetInfo", mock.Anything).Return(map[string]string{"version": "1.0"}, nil)

	checkFunc := NewSLURMAPIHealthCheck(client, 5*time.Second)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "responding normally")
	assert.Equal(t, "true", result.Metadata["ping_successful"])
	assert.Equal(t, "true", result.Metadata["info_successful"])

	client.AssertExpectations(t)
}

func TestSLURMAPIHealthCheck_PingFails(t *testing.T) {
	client := new(MockSLURMClient)
	client.On("Ping", mock.Anything).Return(errors.New("connection refused"))

	checkFunc := NewSLURMAPIHealthCheck(client, 5*time.Second)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Error, "ping failed")
	assert.Contains(t, result.Error, "connection refused")

	client.AssertExpectations(t)
}

func TestSLURMAPIHealthCheck_InfoFails(t *testing.T) {
	client := new(MockSLURMClient)
	client.On("Ping", mock.Anything).Return(nil)
	client.On("GetInfo", mock.Anything).Return(nil, errors.New("unauthorized"))

	checkFunc := NewSLURMAPIHealthCheck(client, 5*time.Second)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, result.Status)
	assert.Contains(t, result.Error, "info call failed")
	assert.Contains(t, result.Error, "unauthorized")
	assert.Equal(t, "true", result.Metadata["ping_successful"])

	client.AssertExpectations(t)
}

func TestCollectorsHealthCheck_AllHealthy(t *testing.T) {
	registry := &MockCollectorRegistry{
		states: map[string]CollectorState{
			"collector1": {
				Name:              "collector1",
				Enabled:           true,
				LastCollection:    time.Now().Add(-30 * time.Second),
				ConsecutiveErrors: 0,
				TotalCollections:  100,
				TotalErrors:       2,
			},
			"collector2": {
				Name:              "collector2",
				Enabled:           true,
				LastCollection:    time.Now().Add(-45 * time.Second),
				ConsecutiveErrors: 0,
				TotalCollections:  50,
				TotalErrors:       0,
			},
		},
	}

	checkFunc := NewCollectorsHealthCheck(registry, 2*time.Minute)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "All 2 enabled collectors are healthy")
	assert.Equal(t, "2", result.Metadata["enabled_collectors"])
	assert.Equal(t, "2", result.Metadata["healthy_collectors"])
	assert.Equal(t, "0", result.Metadata["unhealthy_collectors"])
}

func TestCollectorsHealthCheck_StaleCollector(t *testing.T) {
	registry := &MockCollectorRegistry{
		states: map[string]CollectorState{
			"collector1": {
				Name:           "collector1",
				Enabled:        true,
				LastCollection: time.Now().Add(-5 * time.Minute), // Stale
			},
			"collector2": {
				Name:           "collector2",
				Enabled:        true,
				LastCollection: time.Now().Add(-30 * time.Second), // Fresh
			},
		},
	}

	checkFunc := NewCollectorsHealthCheck(registry, 2*time.Minute)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "Unhealthy collectors")
	assert.Contains(t, result.Message, "collector1")
	assert.Contains(t, result.Message, "stale")
	assert.Equal(t, "stale", result.Metadata["collector_collector1_status"])
}

func TestCollectorsHealthCheck_ConsecutiveErrors(t *testing.T) {
	registry := &MockCollectorRegistry{
		states: map[string]CollectorState{
			"collector1": {
				Name:              "collector1",
				Enabled:           true,
				LastCollection:    time.Now().Add(-30 * time.Second),
				ConsecutiveErrors: 5, // Too many errors
			},
			"collector2": {
				Name:              "collector2",
				Enabled:           true,
				LastCollection:    time.Now().Add(-30 * time.Second),
				ConsecutiveErrors: 2, // Some errors but not critical
			},
		},
	}

	checkFunc := NewCollectorsHealthCheck(registry, 2*time.Minute)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "collector1")
	assert.Contains(t, result.Message, "5 consecutive errors")
	assert.Equal(t, "error_critical", result.Metadata["collector_collector1_status"])
	assert.Equal(t, "error_degraded", result.Metadata["collector_collector2_status"])
}

func TestCollectorsHealthCheck_DisabledCollectors(t *testing.T) {
	registry := &MockCollectorRegistry{
		states: map[string]CollectorState{
			"collector1": {
				Name:    "collector1",
				Enabled: false, // Disabled
			},
			"collector2": {
				Name:           "collector2",
				Enabled:        true,
				LastCollection: time.Now().Add(-30 * time.Second),
			},
		},
	}

	checkFunc := NewCollectorsHealthCheck(registry, 2*time.Minute)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Equal(t, "2", result.Metadata["total_collectors"])
	assert.Equal(t, "1", result.Metadata["enabled_collectors"])
	assert.Equal(t, "disabled", result.Metadata["collector_collector1_status"])
	assert.Equal(t, "healthy", result.Metadata["collector_collector2_status"])
}

func TestMemoryHealthCheck_Normal(t *testing.T) {
	cfg := config.PerformanceMonitoringConfig{
		MemoryThreshold: 500 * 1024 * 1024, // 500MB threshold
	}

	checkFunc := NewMemoryHealthCheck(cfg)
	result := checkFunc(context.Background())

	// Should be healthy since actual memory usage is likely much less than 500MB in tests
	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "Memory usage normal")
	assert.NotEmpty(t, result.Metadata["heap_alloc_mb"])
	assert.NotEmpty(t, result.Metadata["heap_sys_mb"])
	assert.Equal(t, "500", result.Metadata["threshold_mb"])
}

func TestMemoryHealthCheck_NoThreshold(t *testing.T) {
	cfg := config.PerformanceMonitoringConfig{
		MemoryThreshold: 0, // No threshold
	}

	checkFunc := NewMemoryHealthCheck(cfg)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "Memory usage normal")
	assert.NotEmpty(t, result.Metadata["heap_alloc_mb"])
}

func TestDiskHealthCheck_Normal(t *testing.T) {
	// Use OS temp directory which works on all platforms
	tmpDir := os.TempDir()
	checkFunc := NewDiskHealthCheck(tmpDir, 90.0) // 90% threshold
	result := checkFunc(context.Background())

	// Most systems should have temp directory with < 90% usage
	assert.True(t, result.Status == StatusHealthy || result.Status == StatusDegraded)
	assert.NotEmpty(t, result.Metadata["used_percent"])
	assert.Equal(t, tmpDir, result.Metadata["path"])
	assert.Equal(t, "90.0", result.Metadata["threshold"])
}

func TestDiskHealthCheck_InvalidPath(t *testing.T) {
	checkFunc := NewDiskHealthCheck("/nonexistent/path", 90.0)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Error, "Failed to get disk usage")
}

func TestNetworkHealthCheck_Normal(t *testing.T) {
	checkFunc := NewNetworkHealthCheck()
	result := checkFunc(context.Background())

	// Most systems should have at least one network interface
	assert.True(t, result.Status == StatusHealthy || result.Status == StatusDegraded)
	assert.NotEmpty(t, result.Metadata["interface_count"])
}

func TestCircuitBreakerHealthCheck_AllClosed(t *testing.T) {
	getStatuses := func() map[string]interface{} {
		return map[string]interface{}{
			"cb1": map[string]interface{}{"state": "closed"},
			"cb2": map[string]interface{}{"state": "closed"},
		}
	}

	checkFunc := NewCircuitBreakerHealthCheck(getStatuses)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusHealthy, result.Status)
	assert.Contains(t, result.Message, "All 2 circuit breakers are closed")
	assert.Equal(t, "2", result.Metadata["total_circuit_breakers"])
	assert.Equal(t, "2", result.Metadata["healthy_circuit_breakers"])
	assert.Equal(t, "0", result.Metadata["unhealthy_circuit_breakers"])
}

func TestCircuitBreakerHealthCheck_SomeOpen(t *testing.T) {
	getStatuses := func() map[string]interface{} {
		return map[string]interface{}{
			"cb1": map[string]interface{}{"state": "open"},
			"cb2": map[string]interface{}{"state": "closed"},
			"cb3": map[string]interface{}{"state": "half-open"},
		}
	}

	checkFunc := NewCircuitBreakerHealthCheck(getStatuses)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusUnhealthy, result.Status)
	assert.Contains(t, result.Message, "Circuit breakers open")
	assert.Contains(t, result.Message, "cb1")
	assert.Equal(t, "3", result.Metadata["total_circuit_breakers"])
	assert.Equal(t, "1", result.Metadata["healthy_circuit_breakers"])
	assert.Equal(t, "1", result.Metadata["degraded_circuit_breakers"])
	assert.Equal(t, "1", result.Metadata["unhealthy_circuit_breakers"])
}

func TestCircuitBreakerHealthCheck_HalfOpen(t *testing.T) {
	getStatuses := func() map[string]interface{} {
		return map[string]interface{}{
			"cb1": map[string]interface{}{"state": "half-open"},
			"cb2": map[string]interface{}{"state": "closed"},
		}
	}

	checkFunc := NewCircuitBreakerHealthCheck(getStatuses)
	result := checkFunc(context.Background())

	assert.Equal(t, StatusDegraded, result.Status)
	assert.Contains(t, result.Message, "Circuit breakers half-open")
	assert.Contains(t, result.Message, "cb1")
	assert.Equal(t, "2", result.Metadata["total_circuit_breakers"])
	assert.Equal(t, "1", result.Metadata["healthy_circuit_breakers"])
	assert.Equal(t, "1", result.Metadata["degraded_circuit_breakers"])
	assert.Equal(t, "0", result.Metadata["unhealthy_circuit_breakers"])
}
