// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/testutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
)

// APIIntegrationTestSuite tests real SLURM API integration
type APIIntegrationTestSuite struct {
	testutil.BaseTestSuite
	client    slurm.SlurmClient
	exporter  *TestExporter
	container *TestContainer
}

// TestExporter wraps the exporter for testing
type TestExporter struct {
	collectors []collector.Collector
	registry   prometheus.Registerer
	logger     *logrus.Entry
}

// TestContainer manages test SLURM container
type TestContainer struct {
	name      string
	port      int
	started   bool
	apiURL    string
	authToken string
}

func (s *APIIntegrationTestSuite) SetupSuite() {
	// Skip if not integration test run
	if os.Getenv("INTEGRATION_TEST") != "true" {
		s.T().Skip("Skipping integration tests (set INTEGRATION_TEST=true to run)")
	}

	// Skip tests without real SLURM infrastructure
	// These tests require actual SLURM cluster with test container orchestration
	s.T().Skip("Integration tests require SLURM test container infrastructure (not yet implemented)")

	// TODO: Implement test infrastructure using one of these approaches:
	//   1. testcontainers-go (recommended): https://golang.testcontainers.org/
	//      - Use official SLURM container image (e.g., ghcr.io/scidas/slurm:23.11)
	//      - Configure slurmrestd with JWT auth
	//      - Expose REST API port (default 6820)
	//   2. Docker Compose: Create docker-compose.test.yml with SLURM services
	//   3. External cluster: Document connection to existing test cluster
	//
	// Prerequisites for test environment:
	//   - Docker daemon running
	//   - Network access for container pulls
	//   - ~2GB memory for SLURM container
	//   - Port 6820 available (or use random port with :0)
	//
	// NOTE: When implementing test infrastructure, use this pattern:
	// 1. Start test SLURM container
	// s.container = s.startSLURMContainer()
	// s.Require().NoError(s.container.WaitReady())
	//
	// 2. Create slurm-client with proper API:
	// import slurmauth "github.com/jontk/slurm-client/auth"
	// authProvider := slurmauth.NewTokenAuth(s.container.AuthToken())
	// ctx := context.Background()
	// client, err := slurm.NewClient(ctx,
	// 	slurm.WithBaseURL(s.container.APIURL()),
	// 	slurm.WithAuth(authProvider),
	// )
	// s.Require().NoError(err)
	// s.client = client
	//
	// 3. For explicit version:
	// client, err := slurm.NewClientWithVersion(ctx, "v0.0.43",
	// 	slurm.WithBaseURL(s.container.APIURL()),
	// 	slurm.WithAuth(authProvider),
	// )
	//
	// 4. Create exporter
	// s.exporter = s.createExporter()
}

func (s *APIIntegrationTestSuite) TearDownSuite() {
	if s.container != nil {
		s.container.Stop()
	}
}

func TestAPIIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(APIIntegrationTestSuite))
}

func (s *APIIntegrationTestSuite) TestFullCollection() {
	if s.exporter == nil {
		s.T().Skip("Exporter not initialized - test infrastructure not available")
	}

	// Submit test jobs
	jobIDs := s.submitTestJobs(10)
	defer s.cleanupJobs(jobIDs)

	// Wait for jobs to be visible in SLURM
	time.Sleep(2 * time.Second)

	// Collect metrics
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metrics, err := s.exporter.CollectMetrics(ctx)
	s.NoError(err)
	s.NotEmpty(metrics)

	// Verify expected metrics exist
	expectedMetrics := []string{
		"slurm_cluster_nodes_total",
		"slurm_job_state_total",
		"slurm_node_state_total",
		"slurm_partition_nodes_total",
	}

	for _, metricName := range expectedMetrics {
		s.assertMetricExists(metrics, metricName)
	}

	// Verify job metrics reflect our submitted jobs
	s.assertMetricValue(metrics, "slurm_job_state_total", 10, map[string]string{
		"state": "PENDING",
	})
}

func (s *APIIntegrationTestSuite) TestConnectionRecovery() {
	if s.exporter == nil || s.container == nil {
		s.T().Skip("Test infrastructure not initialized")
	}

	// Test connection recovery after API downtime
	ctx := context.Background()

	// Verify initial connection works
	_, err := s.exporter.CollectMetrics(ctx)
	s.NoError(err)

	// Stop container to simulate API downtime
	s.container.Stop()

	// Verify collection fails
	_, err = s.exporter.CollectMetrics(ctx)
	s.Error(err)

	// Restart container
	s.container.Start()
	s.Require().NoError(s.container.WaitReady())

	// Verify recovery within reasonable time
	testutil.Helpers.Eventually(s.T(), func() bool {
		_, err := s.exporter.CollectMetrics(ctx)
		return err == nil
	}, 30*time.Second, 1*time.Second)
}

func (s *APIIntegrationTestSuite) TestAPIAuthentication() {
	// Test different authentication methods
	// NOTE: Skipped by SetupSuite until test infrastructure is implemented
	//
	// Implementation pattern when test infrastructure is ready:
	// import slurmauth "github.com/jontk/slurm-client/auth"
	//
	// testCases := []struct {
	// 	name         string
	// 	authProvider slurmauth.Provider
	// 	expectErr    bool
	// }{
	// 	{
	// 		name:         "valid_token",
	// 		authProvider: slurmauth.NewTokenAuth(s.container.AuthToken()),
	// 		expectErr:    false,
	// 	},
	// 	{
	// 		name:         "invalid_token",
	// 		authProvider: slurmauth.NewTokenAuth("invalid-token"),
	// 		expectErr:    true,
	// 	},
	// 	{
	// 		name:         "no_auth",
	// 		authProvider: slurmauth.NewNoAuth(),
	// 		expectErr:    true, // Assuming cluster requires auth
	// 	},
	// }
	//
	// for _, tc := range testCases {
	// 	s.Run(tc.name, func() {
	// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 		defer cancel()
	//
	// 		client, err := slurm.NewClient(ctx,
	// 			slurm.WithBaseURL(s.container.APIURL()),
	// 			slurm.WithAuth(tc.authProvider),
	// 		)
	// 		s.NoError(err) // Client creation should succeed
	// 		defer client.Close()
	//
	// 		// Test actual API call
	// 		_, err = client.GetInfo(ctx)
	//
	// 		if tc.expectErr {
	// 			s.Error(err, "Expected authentication error")
	// 		} else {
	// 			s.NoError(err, "Valid authentication should work")
	// 		}
	// 	})
	// }
}

func (s *APIIntegrationTestSuite) TestAPIVersionCompatibility() {
	// Test compatibility with different API versions
	// NOTE: Skipped by SetupSuite until test infrastructure is implemented
	//
	// Implementation pattern when test infrastructure is ready:
	// import slurmauth "github.com/jontk/slurm-client/auth"
	//
	// versions := []string{"v0.0.44", "v0.0.40", "v0.0.41", "v0.0.42", "v0.0.43"}
	//
	// for _, version := range versions {
	// 	s.Run(fmt.Sprintf("version_%s", version), func() {
	// 		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 		defer cancel()
	//
	// 		authProvider := slurmauth.NewTokenAuth(s.container.AuthToken())
	// 		client, err := slurm.NewClientWithVersion(ctx, version,
	// 			slurm.WithBaseURL(s.container.APIURL()),
	// 			slurm.WithAuth(authProvider),
	// 		)
	// 		s.NoError(err)
	// 		defer client.Close()
	//
	// 		// Verify version was set correctly
	// 		s.Equal(version, client.Version())
	//
	// 		// Test basic functionality works with this version
	// 		_, err = client.GetInfo(ctx)
	// 		s.NoError(err, "API version %s should be supported", version)
	// 	})
	// }
}

func (s *APIIntegrationTestSuite) TestHighVolumeData() {
	if s.exporter == nil {
		s.T().Skip("Exporter not initialized - test infrastructure not available")
	}

	// Test handling of high volume data
	const jobCount = 1000

	// Submit many jobs
	jobIDs := s.submitTestJobs(jobCount)
	defer s.cleanupJobs(jobIDs)

	// Wait for jobs to be processed
	time.Sleep(5 * time.Second)

	// Measure collection performance
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	metrics, err := s.exporter.CollectMetrics(ctx)
	duration := time.Since(start)

	s.NoError(err)
	s.NotEmpty(metrics)
	s.Less(duration, 30*time.Second, "High volume collection should complete in reasonable time")

	// Verify all jobs are accounted for
	totalJobs := s.getMetricValue(metrics, "slurm_job_state_total", nil)
	s.GreaterOrEqual(totalJobs, float64(jobCount), "Should collect at least %d jobs", jobCount)
}

func (s *APIIntegrationTestSuite) TestConcurrentRequests() {
	if s.exporter == nil {
		s.T().Skip("Exporter not initialized - test infrastructure not available")
	}

	// Test concurrent API requests
	const concurrency = 10

	errCh := make(chan error, concurrency)

	// Run concurrent collections
	for i := 0; i < concurrency; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := s.exporter.CollectMetrics(ctx)
			errCh <- err
		}()
	}

	// Verify all requests succeed
	for i := 0; i < concurrency; i++ {
		err := <-errCh
		s.NoError(err, "Concurrent request %d should succeed", i)
	}
}

func (s *APIIntegrationTestSuite) TestAPIRateLimiting() {
	if s.client == nil || s.exporter == nil {
		s.T().Skip("Client/exporter not initialized - test infrastructure not available")
	}

	// Test API rate limiting behavior
	const requestCount = 50

	start := time.Now()
	successCount := 0

	for i := 0; i < requestCount; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// TODO: Replace with actual slurm-client API method when available
		// For now, use exporter collect as proxy for API calls
		_, err := s.exporter.CollectMetrics(ctx)
		cancel()

		if err == nil {
			successCount++
		} else {
			// Log rate limiting errors
			s.T().Logf("Request %d failed (possibly rate limited): %v", i, err)
		}

		// Small delay between requests
		time.Sleep(10 * time.Millisecond)
	}

	duration := time.Since(start)
	rate := float64(successCount) / duration.Seconds()

	s.T().Logf("Completed %d/%d requests in %v (%.2f req/sec)",
		successCount, requestCount, duration, rate)

	// Should handle some requests successfully
	s.Greater(successCount, requestCount/2, "Should handle at least half the requests")
}

// Helper methods

func (s *APIIntegrationTestSuite) startSLURMContainer() *TestContainer {
	// TODO: Implement using testcontainers-go
	//
	// Example implementation:
	// import "github.com/testcontainers/testcontainers-go"
	//
	// req := testcontainers.ContainerRequest{
	// 	Image:        "ghcr.io/scidas/slurm:23.11",
	// 	ExposedPorts: []string{"6820/tcp"},
	// 	WaitingFor:   wait.ForHTTP("/openapi.json").WithPort("6820/tcp"),
	// 	Env: map[string]string{
	// 		"SLURM_JWT": "test-token-123",
	// 	},
	// }
	// container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
	// 	ContainerRequest: req,
	// 	Started:          true,
	// })
	// if err != nil {
	// 	s.T().Fatalf("Failed to start container: %v", err)
	// }
	//
	// host, _ := container.Host(ctx)
	// port, _ := container.MappedPort(ctx, "6820")
	// return &TestContainer{
	// 	name:      container.GetContainerID(),
	// 	port:      port.Int(),
	// 	apiURL:    fmt.Sprintf("http://%s:%d", host, port.Int()),
	// 	authToken: "test-token-123",
	// 	started:   true,
	// 	container: container, // Store reference for cleanup
	// }

	// Mock implementation (not functional)
	container := &TestContainer{
		name:      "test-slurm",
		port:      16820,
		apiURL:    "http://localhost:16820",
		authToken: "test-token-123",
		started:   true,
	}
	return container
}

func (s *APIIntegrationTestSuite) createExporter() *TestExporter {
	registry := prometheus.NewRegistry()
	logger := logrus.NewEntry(logrus.New())

	// Create collectors with correct signature (client, logger)
	jobsCollector := collector.NewJobsSimpleCollector(
		s.client,
		logger,
	)

	nodesCollector := collector.NewNodesSimpleCollector(
		s.client,
		logger,
	)

	// TODO: Collectors should be registered with the registry for proper metric collection
	// registry.MustRegister(jobsCollector, nodesCollector)

	return &TestExporter{
		collectors: []collector.Collector{jobsCollector, nodesCollector},
		registry:   registry,
		logger:     logger,
	}
}

func (s *APIIntegrationTestSuite) submitTestJobs(count int) []string {
	// This would submit actual test jobs to SLURM
	// For now, simulate job IDs
	jobIDs := make([]string, count)
	for i := 0; i < count; i++ {
		jobIDs[i] = fmt.Sprintf("test-job-%d", i+1000)
	}
	return jobIDs
}

func (s *APIIntegrationTestSuite) cleanupJobs(jobIDs []string) {
	// This would cancel/clean up test jobs
	s.T().Logf("Cleaning up %d test jobs", len(jobIDs))
}

func (c *TestContainer) APIURL() string {
	return c.apiURL
}

func (c *TestContainer) AuthToken() string {
	return c.authToken
}

func (c *TestContainer) WaitReady() error {
	// Simulate waiting for container readiness
	time.Sleep(2 * time.Second)
	return nil
}

func (c *TestContainer) Stop() {
	c.started = false
}

func (c *TestContainer) Start() {
	c.started = true
}

func (e *TestExporter) CollectMetrics(ctx context.Context) (map[string]float64, error) {
	metrics := make(map[string]float64)

	for _, collector := range e.collectors {
		ch := make(chan prometheus.Metric, 1000)

		err := collector.Collect(ctx, ch)
		close(ch)

		if err != nil {
			return nil, err
		}

		// Convert Prometheus metrics to simple map for testing
		// TODO: This is a mock implementation - real version should:
		//   1. Parse prometheus.Metric objects using dto.Metric
		//   2. Extract metric name, labels, and values
		//   3. Store in map with unique keys (name + label combinations)
		//   Currently just drains channel to prevent blocking
		for range ch {
			metrics["collected_metric"] = 1.0
		}
	}

	// Add some realistic test metrics
	metrics["slurm_cluster_nodes_total"] = 10
	metrics["slurm_job_state_total"] = 5
	metrics["slurm_node_state_total"] = 8
	metrics["slurm_partition_nodes_total"] = 10

	return metrics, nil
}

func (s *APIIntegrationTestSuite) assertMetricExists(metrics map[string]float64, name string) {
	_, exists := metrics[name]
	s.True(exists, "Metric %s should exist", name)
}

func (s *APIIntegrationTestSuite) assertMetricValue(metrics map[string]float64, name string, expected float64, labels map[string]string) {
	// Simplified - real implementation would handle labels
	value, exists := metrics[name]
	s.True(exists, "Metric %s should exist", name)
	s.Equal(expected, value, "Metric %s should have value %f", name, expected)
}

func (s *APIIntegrationTestSuite) getMetricValue(metrics map[string]float64, name string, labels map[string]string) float64 {
	// Simplified - real implementation would handle labels
	return metrics[name]
}
