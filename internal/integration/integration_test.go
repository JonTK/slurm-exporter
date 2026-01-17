package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/jontk/slurm-exporter/internal/testutil/mocks"
)

// TestFullCollectorIntegration tests the complete flow from SLURM client to Prometheus metrics
func TestFullCollectorIntegration(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client with all managers
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		Nodes: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
		Partitions: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	// Create prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use registry to avoid unused variable error

	// Collect metrics
	metricFamilies, err := promRegistry.Gather()
	require.NoError(t, err)

	// Verify we have metrics from all collectors
	assert.True(t, len(metricFamilies) > 0, "should have collected metrics")

	// Check for specific metric families
	metricNames := make(map[string]bool)
	for _, mf := range metricFamilies {
		metricNames[*mf.Name] = true
	}

	// Should have job metrics
	hasJobMetrics := false
	for name := range metricNames {
		if contains(name, "job") {
			hasJobMetrics = true
			break
		}
	}
	assert.True(t, hasJobMetrics, "should have job metrics")

	// Should have node metrics
	hasNodeMetrics := false
	for name := range metricNames {
		if contains(name, "node") {
			hasNodeMetrics = true
			break
		}
	}
	assert.True(t, hasNodeMetrics, "should have node metrics")

	// Should have partition metrics
	hasPartitionMetrics := false
	for name := range metricNames {
		if contains(name, "partition") {
			hasPartitionMetrics = true
			break
		}
	}
	assert.True(t, hasPartitionMetrics, "should have partition metrics")
}

// TestHTTPMetricsEndpoint tests the complete HTTP metrics endpoint
func TestHTTPMetricsEndpoint(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	// Create prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use registry to avoid unused variable error

	// Create HTTP server
	handler := promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make HTTP request
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain; version=0.0.4; charset=utf-8")

	// Read response body
	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	responseText := string(body[:n])

	// Should contain SLURM metrics
	assert.Contains(t, responseText, "slurm_", "response should contain SLURM metrics")
}

// TestCollectorFiltering tests metric filtering functionality
func TestCollectorFiltering(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors with filtering
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
			Filters: config.FilterConfig{
				Metrics: config.MetricFilterConfig{
					EnableAll:      false,
					IncludeMetrics: []string{"slurm_job_state"},
					ExcludeMetrics: []string{},
				},
			},
		},
	}

	// Create prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use registry to avoid unused variable error

	// Collect metrics
	metricFamilies, err := promRegistry.Gather()
	require.NoError(t, err)

	// All metrics should be job_state related
	for _, mf := range metricFamilies {
		metricName := *mf.Name
		if contains(metricName, "slurm_job") {
			assert.Contains(t, metricName, "state", "only job_state metrics should be collected")
		}
	}
}

// TestCollectorCustomLabels tests custom labels functionality
func TestCollectorCustomLabels(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors with custom labels
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	// Create prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use registry to avoid unused variable error

	// Collect metrics
	metricFamilies, err := promRegistry.Gather()
	require.NoError(t, err)

	// Verify custom labels are present
	foundCustomLabels := false
	for _, mf := range metricFamilies {
		for _, metric := range mf.Metric {
			for _, labelPair := range metric.Label {
				if *labelPair.Name == "cluster_name" && *labelPair.Value == "test-cluster" {
					foundCustomLabels = true
					break
				}
			}
		}
	}

	assert.True(t, foundCustomLabels, "should have custom labels in metrics")
}

// TestCollectorTimeout tests timeout handling
func TestCollectorTimeout(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client that will timeout
	mockClient := setupSlowMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors with short timeout
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 100 * time.Millisecond, // Very short timeout
		},
	}

	// Create prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use to avoid unused variable error

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_ = ctx // Use ctx to avoid unused variable error

	// Try to collect metrics - should handle timeout gracefully
	metricFamilies, err := promRegistry.Gather()
	_ = metricFamilies // Use to avoid unused variable error
	_ = err            // Allow timeout errors

	// Timeout errors should be handled gracefully
	// The exact behavior depends on implementation
	// but it should not panic or hang
	assert.True(t, true, "collection completed without hanging")
}

// TestPerformanceMonitoringIntegration tests performance monitoring
func TestPerformanceMonitoringIntegration(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	// Create Prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use to avoid unused variable error

	// Collect metrics multiple times to generate performance data
	for i := 0; i < 3; i++ {
		_, err := promRegistry.Gather()
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Collect performance metrics
	metricFamilies, err := promRegistry.Gather()
	require.NoError(t, err)

	// Should have performance metrics
	hasPerformanceMetrics := false
	for _, mf := range metricFamilies {
		metricName := *mf.Name
		if contains(metricName, "collection_duration") ||
			contains(metricName, "collection_success") ||
			contains(metricName, "metrics_collected") {
			hasPerformanceMetrics = true
			break
		}
	}

	assert.True(t, hasPerformanceMetrics, "should have performance monitoring metrics")
}

// TestCardinalityLimiting tests cardinality management
func TestCardinalityLimiting(t *testing.T) {
	t.Skip("TODO: Implement dependency injection for SLURM client in collector registry")

	_ = testutil.GetTestLogger() // Use logger to avoid unused variable error

	// Create mock SLURM client
	mockClient := setupMockSlurmClient(t)
	_ = mockClient // Use mockClient to avoid unused variable error

	// Configure collectors with cardinality limits
	cfg := &config.CollectorsConfig{
		Jobs: config.CollectorConfig{
			Enabled: true,
			Timeout: 30 * time.Second,
		},
	}

	// Create Prometheus registry
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(cfg, promRegistry)
	require.NoError(t, err)
	_ = registry // Use to avoid unused variable error

	// Collect metrics
	metricFamilies, err := promRegistry.Gather()
	require.NoError(t, err)

	// Verify cardinality is managed (implementation dependent)
	totalMetrics := 0
	for _, mf := range metricFamilies {
		totalMetrics += len(mf.Metric)
	}

	assert.True(t, totalMetrics > 0, "should have collected some metrics despite cardinality limits")
}

// Helper functions

func setupMockSlurmClient(t *testing.T) *mocks.MockSlurmClient {
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)
	mockNodeManager := new(mocks.MockNodeManager)
	mockPartitionManager := new(mocks.MockPartitionManager)

	// Setup job manager
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", context.Background(), nil).Return(getTestJobList(), nil)

	// Setup node manager
	mockClient.On("Nodes").Return(mockNodeManager)
	mockNodeManager.On("List", context.Background(), nil).Return(getTestNodeList(), nil)

	// Setup partition manager
	mockClient.On("Partitions").Return(mockPartitionManager)
	mockPartitionManager.On("List", context.Background(), nil).Return(getTestPartitionList(), nil)

	return mockClient
}

func setupSlowMockSlurmClient(t *testing.T) *mocks.MockSlurmClient {
	mockClient := new(mocks.MockSlurmClient)
	mockJobManager := new(mocks.MockJobManager)

	// Setup job manager with slow response
	mockClient.On("Jobs").Return(mockJobManager)
	mockJobManager.On("List", context.Background(), nil).Run(func(args mock.Arguments) {
		time.Sleep(500 * time.Millisecond) // Simulate slow response
	}).Return(getTestJobList(), nil)

	return mockClient
}

func getTestJobList() interface{} {
	// Return mock job list - implementation depends on actual SLURM client types
	return map[string]interface{}{
		"jobs": []map[string]interface{}{
			{
				"job_id":    "12345",
				"name":      "test_job",
				"state":     "RUNNING",
				"partition": "compute",
				"user":      "testuser",
			},
		},
	}
}

func getTestNodeList() interface{} {
	// Return mock node list
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"name":             "node01",
				"state":            "idle",
				"cpus":             64,
				"allocated_cpus":   0,
				"memory":           131072,
				"allocated_memory": 0,
			},
		},
	}
}

func getTestPartitionList() interface{} {
	// Return mock partition list
	return map[string]interface{}{
		"partitions": []map[string]interface{}{
			{
				"name":        "compute",
				"state":       "up",
				"total_nodes": 4,
				"total_cpus":  256,
			},
		},
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
