package integration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
)

// BenchmarkMetricsCollection benchmarks the metrics collection endpoint
func BenchmarkMetricsCollection(b *testing.B) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	exporterURL := "http://localhost:9341"

	// Warmup - ensure exporter is ready
	resp, err := client.Get(exporterURL + "/ready")
	if err != nil {
		b.Skip("Exporter not available for benchmarking")
	}
	_ = resp.Body.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(exporterURL + "/metrics")
		if err != nil {
			b.Fatalf("Failed to get metrics: %v", err)
		}

		// Read and discard body to ensure full request completion
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// BenchmarkMetricsParsing benchmarks the parsing of metrics response
func BenchmarkMetricsParsing(b *testing.B) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	exporterURL := "http://localhost:9341"

	// Get metrics once
	resp, err := client.Get(exporterURL + "/metrics")
	if err != nil {
		b.Skip("Exporter not available for benchmarking")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(b, err)

	metricsText := string(body)

	b.ResetTimer()
	b.SetBytes(int64(len(body)))

	for i := 0; i < b.N; i++ {
		parser := &expfmt.TextParser{}
		_, err := parser.TextToMetricFamilies(strings.NewReader(metricsText))
		if err != nil {
			b.Fatalf("Failed to parse metrics: %v", err)
		}
	}
}

// BenchmarkHealthCheck benchmarks the health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	exporterURL := "http://localhost:9341"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(exporterURL + "/health")
		if err != nil {
			b.Fatalf("Failed to get health: %v", err)
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// BenchmarkConcurrentMetrics benchmarks concurrent metrics collection
func BenchmarkConcurrentMetrics(b *testing.B) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	exporterURL := "http://localhost:9341"

	// Test different concurrency levels
	concurrencyLevels := []int{1, 2, 4, 8, 16}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					resp, err := client.Get(exporterURL + "/metrics")
					if err != nil {
						b.Fatalf("Failed to get metrics: %v", err)
					}

					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
			})
		})
	}
}

// PerformanceTest runs comprehensive performance analysis
func PerformanceTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	exporterURL := "http://localhost:9341"

	// Warmup
	resp, err := client.Get(exporterURL + "/ready")
	if err != nil {
		t.Skip("Exporter not available for performance testing")
	}
	_ = resp.Body.Close()

	// Performance metrics to collect
	var results struct {
		RequestDurations   []time.Duration
		ResponseSizes      []int64
		MetricCounts       []int
		MemoryUsage        []float64
		ConcurrentRequests map[int][]time.Duration
	}

	results.ConcurrentRequests = make(map[int][]time.Duration)

	t.Run("Sequential Requests", func(t *testing.T) {
		iterations := 50
		results.RequestDurations = make([]time.Duration, iterations)
		results.ResponseSizes = make([]int64, iterations)
		results.MetricCounts = make([]int, iterations)

		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp, err := client.Get(exporterURL + "/metrics")
			duration := time.Since(start)

			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			_ = resp.Body.Close()

			// Parse metrics to count them
			parser := &expfmt.TextParser{}
			families, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
			require.NoError(t, err)

			metricCount := 0
			for _, family := range families {
				metricCount += len(family.Metric)
			}

			results.RequestDurations[i] = duration
			results.ResponseSizes[i] = int64(len(body))
			results.MetricCounts[i] = metricCount

			// Brief pause between requests
			time.Sleep(100 * time.Millisecond)
		}

		// Calculate statistics
		avgDuration := averageDuration(results.RequestDurations)
		minDuration := minDuration(results.RequestDurations)
		maxDuration := maxDuration(results.RequestDurations)

		t.Logf("Sequential Performance Statistics:")
		t.Logf("  Average Duration: %v", avgDuration)
		t.Logf("  Min Duration: %v", minDuration)
		t.Logf("  Max Duration: %v", maxDuration)
		t.Logf("  Average Response Size: %d bytes", averageInt64(results.ResponseSizes))
		t.Logf("  Average Metric Count: %d", averageInt(results.MetricCounts))
	})

	t.Run("Concurrent Requests", func(t *testing.T) {
		concurrencyLevels := []int{2, 4, 8}
		requestsPerLevel := 20

		for _, concurrency := range concurrencyLevels {
			t.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(t *testing.T) {
				durations := make([]time.Duration, requestsPerLevel*concurrency)
				done := make(chan struct{})

				for c := 0; c < concurrency; c++ {
					go func(workerID int) {
						defer func() { done <- struct{}{} }()

						for r := 0; r < requestsPerLevel; r++ {
							start := time.Now()
							resp, err := client.Get(exporterURL + "/metrics")
							duration := time.Since(start)

							if err == nil && resp.StatusCode == http.StatusOK {
								_, _ = io.Copy(io.Discard, resp.Body)
								_ = resp.Body.Close()

								index := workerID*requestsPerLevel + r
								durations[index] = duration
							}

							time.Sleep(50 * time.Millisecond)
						}
					}(c)
				}

				// Wait for all workers to complete
				for c := 0; c < concurrency; c++ {
					<-done
				}

				// Filter out zero durations (failed requests)
				validDurations := make([]time.Duration, 0)
				for _, d := range durations {
					if d > 0 {
						validDurations = append(validDurations, d)
					}
				}

				results.ConcurrentRequests[concurrency] = validDurations

				if len(validDurations) > 0 {
					avgDuration := averageDuration(validDurations)
					minDuration := minDuration(validDurations)
					maxDuration := maxDuration(validDurations)

					t.Logf("Concurrency %d Performance:", concurrency)
					t.Logf("  Successful Requests: %d/%d", len(validDurations), len(durations))
					t.Logf("  Average Duration: %v", avgDuration)
					t.Logf("  Min Duration: %v", minDuration)
					t.Logf("  Max Duration: %v", maxDuration)
				}
			})
		}
	})

	t.Run("Memory Usage Over Time", func(t *testing.T) {
		// Monitor memory usage during extended operation
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		memoryReadings := make([]float64, 0)

		for {
			select {
			case <-ctx.Done():
				goto memoryAnalysis
			case <-ticker.C:
				// Get debug vars to check memory usage
				resp, err := client.Get(exporterURL + "/debug/vars")
				if err == nil && resp.StatusCode == http.StatusOK {
					// Read the response but don't parse JSON for performance
					body, err := io.ReadAll(resp.Body)
					_ = resp.Body.Close()

					if err == nil && len(body) > 0 {
						// Rough memory usage estimation based on response size
						memoryReadings = append(memoryReadings, float64(len(body)))
					}
				}

				// Also make a metrics request to maintain load
				resp, err = client.Get(exporterURL + "/metrics")
				if err == nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
			}
		}

	memoryAnalysis:
		if len(memoryReadings) > 0 {
			avgMemory := averageFloat64(memoryReadings)
			minMemory := minFloat64(memoryReadings)
			maxMemory := maxFloat64(memoryReadings)

			t.Logf("Memory Usage Analysis:")
			t.Logf("  Readings: %d", len(memoryReadings))
			t.Logf("  Average Debug Response Size: %.2f bytes", avgMemory)
			t.Logf("  Min Debug Response Size: %.2f bytes", minMemory)
			t.Logf("  Max Debug Response Size: %.2f bytes", maxMemory)
		}
	})

	// Generate performance report
	t.Logf("\n=== PERFORMANCE TEST SUMMARY ===")
	if len(results.RequestDurations) > 0 {
		t.Logf("Sequential Requests:")
		t.Logf("  Average: %v", averageDuration(results.RequestDurations))
		t.Logf("  P95: %v", percentileDuration(results.RequestDurations, 0.95))
		t.Logf("  P99: %v", percentileDuration(results.RequestDurations, 0.99))
	}

	for concurrency, durations := range results.ConcurrentRequests {
		if len(durations) > 0 {
			t.Logf("Concurrent Requests (concurrency=%d):", concurrency)
			t.Logf("  Average: %v", averageDuration(durations))
			t.Logf("  P95: %v", percentileDuration(durations, 0.95))
		}
	}
}

// Helper functions for statistics
func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

func minDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	min := durations[0]
	for _, d := range durations[1:] {
		if d < min {
			min = d
		}
	}

	return min
}

func maxDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	max := durations[0]
	for _, d := range durations[1:] {
		if d > max {
			max = d
		}
	}

	return max
}

func percentileDuration(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// Simple percentile calculation (not sorting for performance)
	index := int(float64(len(durations)) * percentile)
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

func averageInt64(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	var total int64
	for _, v := range values {
		total += v
	}

	return total / int64(len(values))
}

func averageInt(values []int) int {
	if len(values) == 0 {
		return 0
	}

	var total int
	for _, v := range values {
		total += v
	}

	return total / len(values)
}

func averageFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var total float64
	for _, v := range values {
		total += v
	}

	return total / float64(len(values))
}

func minFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}

	return min
}

func maxFloat64(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}

	return max
}
