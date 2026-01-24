// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RockyClusterTestSuite runs integration tests against rocky9.ar.jontk.com
type RockyClusterTestSuite struct {
	suite.Suite
	exporterURL    string
	client         *http.Client
	startTime      time.Time
	collectedStats map[string]interface{}
}

// SetupSuite initializes the test suite
func (suite *RockyClusterTestSuite) SetupSuite() {
	suite.exporterURL = "http://localhost:9341"
	suite.client = &http.Client{
		Timeout: 30 * time.Second,
	}
	suite.startTime = time.Now()
	suite.collectedStats = make(map[string]interface{})

	// Wait for exporter to be ready
	suite.waitForExporter()
}

// TearDownSuite cleans up after tests
func (suite *RockyClusterTestSuite) TearDownSuite() {
	// Generate test report
	suite.generateTestReport()
}

// waitForExporter waits for the exporter to become ready
func (suite *RockyClusterTestSuite) waitForExporter() {
	// Quick check if exporter is available, skip if not
	quickCtx, quickCancel := context.WithTimeout(context.Background(), 5*time.Second)
	req, _ := http.NewRequestWithContext(quickCtx, http.MethodGet, suite.exporterURL+"/ready", nil)
	quickCancel()

	resp, err := suite.client.Do(req)
	if err != nil {
		suite.T().Skipf("Skipping integration test - exporter not running at %s: %v", suite.exporterURL, err)
		return
	}
	if resp != nil {
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			suite.T().Log("Exporter is ready")
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second) // Check more frequently
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			suite.T().Skipf("Skipping integration test - exporter not ready at %s within timeout", suite.exporterURL)
			return
		case <-ticker.C:
			reqCtx, reqCancel := context.WithTimeout(context.Background(), 5*time.Second)

			req, _ := http.NewRequestWithContext(reqCtx, http.MethodGet, suite.exporterURL+"/ready", nil)

			reqCancel()

			resp, err := suite.client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				_ = resp.Body.Close()
				suite.T().Log("Exporter is ready")
				return
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
			suite.T().Log("Waiting for exporter to become ready...")
		}
	}
}

// TestHealthEndpoints tests all health-related endpoints
func (suite *RockyClusterTestSuite) TestHealthEndpoints() {
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		checkResponse  func(body []byte) error
	}{
		{
			name:           "Health endpoint",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) error {
				var health map[string]interface{}
				if err := json.Unmarshal(body, &health); err != nil {
					return fmt.Errorf("failed to parse health response: %w", err)
				}

				status, ok := health["status"].(string)
				if !ok || status != "healthy" {
					return fmt.Errorf("expected status 'healthy', got %v", status)
				}

				// Check for required health checks
				checks, ok := health["checks"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("health checks not found in response")
				}

				requiredChecks := []string{"slurm_connectivity", "metric_collection"}
				for _, check := range requiredChecks {
					if _, exists := checks[check]; !exists {
						return fmt.Errorf("required health check '%s' not found", check)
					}
				}

				suite.collectedStats["health_checks"] = checks
				return nil
			},
		},
		{
			name:           "Ready endpoint",
			endpoint:       "/ready",
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) error {
				if !strings.Contains(string(body), "Ready") {
					return fmt.Errorf("expected 'Ready' in response, got: %s", string(body))
				}
				return nil
			},
		},
		{
			name:           "Debug config endpoint",
			endpoint:       "/debug/config",
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) error {
				var config map[string]interface{}
				if err := json.Unmarshal(body, &config); err != nil {
					return fmt.Errorf("failed to parse config response: %w", err)
				}

				// Verify SLURM configuration
				slurm, ok := config["slurm"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("slurm configuration not found")
				}

				host, ok := slurm["host"].(string)
				if !ok || host != "rocky9.ar.jontk.com" {
					return fmt.Errorf("expected host 'rocky9.ar.jontk.com', got %v", host)
				}

				suite.collectedStats["config"] = config
				return nil
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+test.endpoint, nil)

			cancel()

			resp, err := suite.client.Do(req)
			require.NoError(suite.T(), err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(suite.T(), test.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(suite.T(), err)

			if test.checkResponse != nil {
				err := test.checkResponse(body)
				assert.NoError(suite.T(), err, "Response validation failed for %s", test.name)
			}
		})
	}
}

// TestMetricsEndpoint tests the main metrics endpoint
func (suite *RockyClusterTestSuite) TestMetricsEndpoint() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/metrics", nil)

	cancel()

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	assert.Equal(suite.T(), "text/plain; version=0.0.4; charset=utf-8", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)

	metrics := string(body)
	suite.T().Logf("Metrics response size: %d bytes", len(metrics))

	// Parse metrics to validate format
	parser := &expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(metrics))
	require.NoError(suite.T(), err, "Failed to parse metrics")

	suite.collectedStats["total_metric_families"] = len(metricFamilies)

	// Count total metrics
	totalMetrics := 0
	slurmMetrics := 0
	for name, family := range metricFamilies {
		for range family.GetMetric() {
			totalMetrics++
			if strings.HasPrefix(name, "slurm_") {
				slurmMetrics++
			}
		}
	}

	suite.collectedStats["total_metrics"] = totalMetrics
	suite.collectedStats["slurm_metrics"] = slurmMetrics

	suite.T().Logf("Total metric families: %d", len(metricFamilies))
	suite.T().Logf("Total metric samples: %d", totalMetrics)
	suite.T().Logf("SLURM metric samples: %d", slurmMetrics)

	// Validate we have SLURM metrics
	assert.Greater(suite.T(), slurmMetrics, 0, "No SLURM metrics found")
	assert.Greater(suite.T(), slurmMetrics, 10, "Too few SLURM metrics, expected at least 10")
}

// TestCollectorMetrics tests specific collector metrics
func (suite *RockyClusterTestSuite) TestCollectorMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/metrics", nil)

	cancel()

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)

	parser := &expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	require.NoError(suite.T(), err)

	// Test for essential metric families
	essentialMetrics := []struct {
		name        string
		description string
		metricType  dto.MetricType
		required    bool
	}{
		{"slurm_jobs_total", "Total number of jobs", dto.MetricType_COUNTER, true},
		{"slurm_nodes_total", "Total number of nodes", dto.MetricType_GAUGE, true},
		{"slurm_partition_nodes_total", "Total nodes in partition", dto.MetricType_GAUGE, true},
		{"slurm_exporter_collect_duration_seconds", "Collection duration", dto.MetricType_HISTOGRAM, true},
		{"slurm_controller_up", "SLURM controller status", dto.MetricType_GAUGE, true},
		{"up", "Exporter up status", dto.MetricType_GAUGE, true},
	}

	collectorStats := make(map[string]interface{})

	for _, metric := range essentialMetrics {
		suite.Run(fmt.Sprintf("Metric_%s", metric.name), func() {
			family, exists := metricFamilies[metric.name]
			if metric.required {
				require.True(suite.T(), exists, "Required metric %s not found", metric.name)
			}
			if exists {
				assert.Equal(suite.T(), metric.metricType, family.GetType(), "Wrong metric type for %s", metric.name)
				assert.NotEmpty(suite.T(), family.GetMetric(), "No samples for metric %s", metric.name)

				// Count samples
				collectorStats[metric.name] = len(family.GetMetric())
			}
		})
	}

	suite.collectedStats["collector_metrics"] = collectorStats
}

// TestCollectorPerformance tests collection performance
func (suite *RockyClusterTestSuite) TestCollectorPerformance() {
	// Get initial metrics
	perfCtx, perfCancel := context.WithTimeout(context.Background(), 10*time.Second)
	perfReq, err := http.NewRequestWithContext(perfCtx, http.MethodGet, suite.exporterURL+"/metrics", nil)
	perfCancel()
	require.NoError(suite.T(), err)
	initialResp, err := suite.client.Do(perfReq)
	require.NoError(suite.T(), err)
	_ = initialResp.Body.Close()

	// Measure collection time over multiple collections
	collections := 5
	durations := make([]time.Duration, collections)

	for i := 0; i < collections; i++ {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/metrics", nil)

		cancel()

		resp, err := suite.client.Do(req)
		duration := time.Since(start)

		require.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()

		durations[i] = duration
		suite.T().Logf("Collection %d took: %v", i+1, duration)

		// Wait between collections
		time.Sleep(2 * time.Second)
	}

	// Calculate statistics
	var totalDuration time.Duration
	var maxDuration time.Duration
	minDuration := durations[0]

	for _, d := range durations {
		totalDuration += d
		if d > maxDuration {
			maxDuration = d
		}
		if d < minDuration {
			minDuration = d
		}
	}

	avgDuration := totalDuration / time.Duration(collections)

	suite.collectedStats["performance"] = map[string]interface{}{
		"collections":     collections,
		"avg_duration_ms": avgDuration.Milliseconds(),
		"min_duration_ms": minDuration.Milliseconds(),
		"max_duration_ms": maxDuration.Milliseconds(),
	}

	suite.T().Logf("Performance statistics:")
	suite.T().Logf("  Average: %v", avgDuration)
	suite.T().Logf("  Min: %v", minDuration)
	suite.T().Logf("  Max: %v", maxDuration)

	// Assert reasonable performance
	assert.Less(suite.T(), avgDuration.Seconds(), 10.0, "Average collection time too high")
	assert.Less(suite.T(), maxDuration.Seconds(), 30.0, "Max collection time too high")
}

// TestSLURMConnectivity tests SLURM-specific functionality
func (suite *RockyClusterTestSuite) TestSLURMConnectivity() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/metrics", nil)

	cancel()

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)

	parser := &expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	require.NoError(suite.T(), err)

	// Check controller status
	if controllerFamily, exists := metricFamilies["slurm_controller_up"]; exists {
		require.NotEmpty(suite.T(), controllerFamily.GetMetric(), "No controller status metrics")

		controllerUp := controllerFamily.GetMetric()[0].GetGauge().GetValue()
		assert.Equal(suite.T(), float64(1), controllerUp, "SLURM controller is down")

		suite.collectedStats["slurm_controller_up"] = controllerUp == 1
	}

	// Check for job metrics (indicates successful SLURM API calls)
	if jobsFamily, exists := metricFamilies["slurm_jobs_total"]; exists {
		require.NotEmpty(suite.T(), jobsFamily.GetMetric(), "No job metrics found")

		totalJobs := 0
		jobStates := make(map[string]int)

		for _, metric := range jobsFamily.GetMetric() {
			value := int(metric.GetCounter().GetValue())
			totalJobs += value

			// Extract state label
			for _, label := range metric.GetLabel() {
				if label.GetName() == "state" {
					jobStates[label.GetValue()] += value
				}
			}
		}

		suite.collectedStats["total_jobs"] = totalJobs
		suite.collectedStats["job_states"] = jobStates

		suite.T().Logf("Total jobs found: %d", totalJobs)
		suite.T().Logf("Job states: %+v", jobStates)
	}

	// Check for node metrics
	if nodesFamily, exists := metricFamilies["slurm_nodes_total"]; exists {
		require.NotEmpty(suite.T(), nodesFamily.GetMetric(), "No node metrics found")

		totalNodes := 0
		nodeStates := make(map[string]int)

		for _, metric := range nodesFamily.GetMetric() {
			value := int(metric.GetGauge().GetValue())
			totalNodes += value

			// Extract state label
			for _, label := range metric.GetLabel() {
				if label.GetName() == "state" {
					nodeStates[label.GetValue()] += value
				}
			}
		}

		suite.collectedStats["total_nodes"] = totalNodes
		suite.collectedStats["node_states"] = nodeStates

		suite.T().Logf("Total nodes found: %d", totalNodes)
		suite.T().Logf("Node states: %+v", nodeStates)

		assert.Greater(suite.T(), totalNodes, 0, "No nodes found in cluster")
	}
}

// TestCollectorHealth tests individual collector health
func (suite *RockyClusterTestSuite) TestCollectorHealth() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/debug/vars", nil)

	cancel()

	resp, err := suite.client.Do(req)
	require.NoError(suite.T(), err)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)

	var vars map[string]interface{}
	err = json.Unmarshal(body, &vars)
	require.NoError(suite.T(), err)

	// Check for collector-specific metrics
	slurmExporter, ok := vars["slurm_exporter"].(map[string]interface{})
	if ok {
		suite.collectedStats["exporter_vars"] = slurmExporter

		// Check collection statistics
		if collections, exists := slurmExporter["total_collections"]; exists {
			assert.Greater(suite.T(), collections, float64(0), "No collections performed")
		}

		if errors, exists := slurmExporter["collection_errors"]; exists {
			errorCount := errors.(float64)
			suite.T().Logf("Collection errors: %v", errorCount)
			suite.collectedStats["collection_errors"] = errorCount
		}
	}

	// Check memory statistics
	if memstats, ok := vars["memstats"].(map[string]interface{}); ok {
		if alloc, exists := memstats["Alloc"]; exists {
			allocBytes := alloc.(float64)
			allocMB := allocBytes / 1024 / 1024
			suite.T().Logf("Memory allocation: %.2f MB", allocMB)
			suite.collectedStats["memory_alloc_mb"] = allocMB

			// Assert reasonable memory usage (less than 1GB)
			assert.Less(suite.T(), allocMB, 1024.0, "Memory usage too high")
		}
	}
}

// TestTracingIntegration tests distributed tracing features
func (suite *RockyClusterTestSuite) TestTracingIntegration() {
	// Test tracing status endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, suite.exporterURL+"/debug/tracing/stats", nil)

	cancel()

	resp, err := suite.client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		suite.T().Skip("Tracing not available or not enabled")
		return
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err)

	var tracingStats map[string]interface{}
	err = json.Unmarshal(body, &tracingStats)
	require.NoError(suite.T(), err)

	suite.collectedStats["tracing_stats"] = tracingStats

	enabled, ok := tracingStats["enabled"].(bool)
	if ok && enabled {
		suite.T().Log("Tracing is enabled")

		// Test enabling detailed tracing
		enableReq := `{"collector": "jobs", "duration": "30s"}`
		enableCtx, enableCancel := context.WithTimeout(context.Background(), 5*time.Second)
		enableRequest, err := http.NewRequestWithContext(enableCtx, http.MethodPost, suite.exporterURL+"/debug/tracing/enable", strings.NewReader(enableReq))
		enableRequest.Header.Set("Content-Type", "application/json")
		enableCancel()
		if err == nil {
			enableResp, err := suite.client.Do(enableRequest)
			if err == nil {
				_ = enableResp.Body.Close()
				suite.T().Log("Successfully enabled detailed tracing for jobs collector")
			}
		}
	} else {
		suite.T().Log("Tracing is disabled")
	}
}

// generateTestReport creates a comprehensive test report
func (suite *RockyClusterTestSuite) generateTestReport() {
	report := map[string]interface{}{
		"test_run": map[string]interface{}{
			"start_time":   suite.startTime,
			"end_time":     time.Now(),
			"duration":     time.Since(suite.startTime).String(),
			"cluster":      "rocky9.ar.jontk.com",
			"exporter_url": suite.exporterURL,
		},
		"statistics": suite.collectedStats,
		"summary": map[string]interface{}{
			"tests_passed":       suite.T().Failed() == false,
			"cluster_accessible": suite.collectedStats["slurm_controller_up"],
		},
	}

	// Write report to file
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err == nil {
		filename := fmt.Sprintf("test-report-%s.json", time.Now().Format("20060102-150405"))
		_ = os.WriteFile(filename, reportJSON, 0644)
		suite.T().Logf("Test report written to: %s", filename)
	}

	// Print summary
	suite.T().Log("=== INTEGRATION TEST SUMMARY ===")
	suite.T().Logf("Cluster: rocky9.ar.jontk.com")
	suite.T().Logf("Test Duration: %v", time.Since(suite.startTime))
	suite.T().Logf("Controller Status: %v", suite.collectedStats["slurm_controller_up"])
	suite.T().Logf("Total Metrics: %v", suite.collectedStats["total_metrics"])
	suite.T().Logf("SLURM Metrics: %v", suite.collectedStats["slurm_metrics"])
	suite.T().Logf("Total Jobs: %v", suite.collectedStats["total_jobs"])
	suite.T().Logf("Total Nodes: %v", suite.collectedStats["total_nodes"])
	suite.T().Logf("Memory Usage: %v MB", suite.collectedStats["memory_alloc_mb"])

	if perfStats, ok := suite.collectedStats["performance"].(map[string]interface{}); ok {
		suite.T().Logf("Avg Collection Time: %v ms", perfStats["avg_duration_ms"])
	}
}

// TestRockyClusterIntegration runs the full integration test suite
func TestRockyClusterIntegration(t *testing.T) {
	// Skip if running in CI without proper setup
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Integration tests skipped")
	}

	suite.Run(t, new(RockyClusterTestSuite))
}
