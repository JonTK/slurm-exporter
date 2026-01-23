// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package testutil

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PerformanceTestSuite provides utilities for performance testing
type PerformanceTestSuite struct {
	BaseTestSuite
}

// BenchmarkOptions configures performance benchmarks
type BenchmarkOptions struct {
	MinIterations    int
	MaxIterations    int
	Duration         time.Duration
	WarmupIterations int
	MemoryTracking   bool
	GCControl        bool
}

// DefaultBenchmarkOptions returns sensible defaults
func DefaultBenchmarkOptions() BenchmarkOptions {
	return BenchmarkOptions{
		MinIterations:    10,
		MaxIterations:    10000,
		Duration:         5 * time.Second,
		WarmupIterations: 3,
		MemoryTracking:   true,
		GCControl:        true,
	}
}

// BenchmarkResult contains benchmark results
type BenchmarkResult struct {
	Iterations       int
	Duration         time.Duration
	NsPerOp          int64
	BytesPerOp       int64
	AllocsPerOp      int64
	MemoryBefore     runtime.MemStats
	MemoryAfter      runtime.MemStats
	ThroughputPerSec float64
}

// String returns a string representation of the benchmark result
func (br *BenchmarkResult) String() string {
	return formatBenchmarkResult(br)
}

// RunBenchmark runs a performance benchmark
func (pts *PerformanceTestSuite) RunBenchmark(name string, fn func(), opts BenchmarkOptions) *BenchmarkResult {
	pts.T().Helper()

	// Warmup
	for i := 0; i < opts.WarmupIterations; i++ {
		fn()
	}

	if opts.GCControl {
		runtime.GC()
		runtime.GC() // Double GC to ensure clean state
	}

	var memBefore, memAfter runtime.MemStats
	if opts.MemoryTracking {
		runtime.ReadMemStats(&memBefore)
	}

	start := time.Now()
	iterations := 0

	// Run benchmark
	for time.Since(start) < opts.Duration && iterations < opts.MaxIterations {
		fn()
		iterations++

		if iterations >= opts.MinIterations && time.Since(start) >= opts.Duration/2 {
			// Check if we should continue based on time
			if time.Since(start) >= opts.Duration {
				break
			}
		}
	}

	duration := time.Since(start)

	if opts.MemoryTracking {
		runtime.ReadMemStats(&memAfter)
	}

	result := &BenchmarkResult{
		Iterations:   iterations,
		Duration:     duration,
		NsPerOp:      duration.Nanoseconds() / int64(iterations),
		MemoryBefore: memBefore,
		MemoryAfter:  memAfter,
	}

	if opts.MemoryTracking {
		result.BytesPerOp = int64(memAfter.TotalAlloc-memBefore.TotalAlloc) / int64(iterations)
		result.AllocsPerOp = int64(memAfter.Mallocs-memBefore.Mallocs) / int64(iterations)
	}

	result.ThroughputPerSec = float64(iterations) / duration.Seconds()

	pts.T().Logf("Benchmark %s: %s", name, result.String())

	return result
}

// AssertPerformance asserts performance characteristics
func (pts *PerformanceTestSuite) AssertPerformance(result *BenchmarkResult, expectations PerformanceExpectations) {
	pts.T().Helper()

	if expectations.MaxNsPerOp > 0 {
		assert.LessOrEqual(pts.T(), result.NsPerOp, expectations.MaxNsPerOp,
			"Operation too slow: %d ns/op > %d ns/op", result.NsPerOp, expectations.MaxNsPerOp)
	}

	if expectations.MaxBytesPerOp > 0 {
		assert.LessOrEqual(pts.T(), result.BytesPerOp, expectations.MaxBytesPerOp,
			"Too much memory allocation: %d bytes/op > %d bytes/op",
			result.BytesPerOp, expectations.MaxBytesPerOp)
	}

	if expectations.MaxAllocsPerOp > 0 {
		assert.LessOrEqual(pts.T(), result.AllocsPerOp, expectations.MaxAllocsPerOp,
			"Too many allocations: %d allocs/op > %d allocs/op",
			result.AllocsPerOp, expectations.MaxAllocsPerOp)
	}

	if expectations.MinThroughputPerSec > 0 {
		assert.GreaterOrEqual(pts.T(), result.ThroughputPerSec, expectations.MinThroughputPerSec,
			"Throughput too low: %.2f ops/sec < %.2f ops/sec",
			result.ThroughputPerSec, expectations.MinThroughputPerSec)
	}
}

// PerformanceExpectations defines expected performance characteristics
type PerformanceExpectations struct {
	MaxNsPerOp          int64   // Maximum nanoseconds per operation
	MaxBytesPerOp       int64   // Maximum bytes allocated per operation
	MaxAllocsPerOp      int64   // Maximum allocations per operation
	MinThroughputPerSec float64 // Minimum operations per second
}

// MemorySnapshot captures memory statistics
type MemorySnapshot struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	Mallocs      uint64
	Frees        uint64
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapReleased uint64
	HeapObjects  uint64
	StackInuse   uint64
	StackSys     uint64
	NumGC        uint32
	PauseTotalNs uint64
}

// TakeMemorySnapshot captures current memory statistics
func TakeMemorySnapshot() MemorySnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemorySnapshot{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
	}
}

// MemoryDiff calculates the difference between two memory snapshots
func MemoryDiff(before, after MemorySnapshot) MemorySnapshot {
	return MemorySnapshot{
		Alloc:        after.Alloc - before.Alloc,
		TotalAlloc:   after.TotalAlloc - before.TotalAlloc,
		Sys:          after.Sys - before.Sys,
		Mallocs:      after.Mallocs - before.Mallocs,
		Frees:        after.Frees - before.Frees,
		HeapAlloc:    after.HeapAlloc - before.HeapAlloc,
		HeapSys:      after.HeapSys - before.HeapSys,
		HeapIdle:     after.HeapIdle - before.HeapIdle,
		HeapInuse:    after.HeapInuse - before.HeapInuse,
		HeapReleased: after.HeapReleased - before.HeapReleased,
		HeapObjects:  after.HeapObjects - before.HeapObjects,
		StackInuse:   after.StackInuse - before.StackInuse,
		StackSys:     after.StackSys - before.StackSys,
		NumGC:        after.NumGC - before.NumGC,
		PauseTotalNs: after.PauseTotalNs - before.PauseTotalNs,
	}
}

// LoadTest represents a load testing scenario
type LoadTest struct {
	Name        string
	Concurrency int
	Duration    time.Duration
	RampUpTime  time.Duration
	Function    func(ctx context.Context) error
}

// LoadTestResult contains load test results
type LoadTestResult struct {
	TotalRequests   int64
	SuccessfulReqs  int64
	FailedRequests  int64
	Duration        time.Duration
	RequestsPerSec  float64
	AvgResponseTime time.Duration
	MaxResponseTime time.Duration
	MinResponseTime time.Duration
	Errors          []error
}

// RunLoadTest executes a load test
func (pts *PerformanceTestSuite) RunLoadTest(test LoadTest) *LoadTestResult {
	pts.T().Helper()

	ctx, cancel := context.WithTimeout(pts.ctx, test.Duration)
	defer cancel()

	result := &LoadTestResult{
		MinResponseTime: time.Hour, // Will be updated with actual minimum
	}

	// Channel for collecting results
	type requestResult struct {
		duration time.Duration
		err      error
	}
	resultCh := make(chan requestResult, test.Concurrency*100)

	// Start workers
	for i := 0; i < test.Concurrency; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					err := test.Function(ctx)
					duration := time.Since(start)

					select {
					case resultCh <- requestResult{duration: duration, err: err}:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Collect results
	start := time.Now()
	var totalResponseTime time.Duration

	for {
		select {
		case res := <-resultCh:
			result.TotalRequests++
			totalResponseTime += res.duration

			if res.err != nil {
				result.FailedRequests++
				result.Errors = append(result.Errors, res.err)
			} else {
				result.SuccessfulReqs++
			}

			// Update response time stats
			if res.duration > result.MaxResponseTime {
				result.MaxResponseTime = res.duration
			}
			if res.duration < result.MinResponseTime {
				result.MinResponseTime = res.duration
			}

		case <-ctx.Done():
			result.Duration = time.Since(start)
			if result.TotalRequests > 0 {
				result.RequestsPerSec = float64(result.TotalRequests) / result.Duration.Seconds()
				result.AvgResponseTime = totalResponseTime / time.Duration(result.TotalRequests)
			}
			return result
		}
	}
}

// AssertLoadTestResults validates load test results
func (pts *PerformanceTestSuite) AssertLoadTestResults(result *LoadTestResult, minRPS float64, maxErrorRate float64) {
	pts.T().Helper()

	require.Greater(pts.T(), result.TotalRequests, int64(0), "No requests were made")

	errorRate := float64(result.FailedRequests) / float64(result.TotalRequests)
	assert.LessOrEqual(pts.T(), errorRate, maxErrorRate,
		"Error rate too high: %.2f%% > %.2f%%", errorRate*100, maxErrorRate*100)

	assert.GreaterOrEqual(pts.T(), result.RequestsPerSec, minRPS,
		"Request rate too low: %.2f RPS < %.2f RPS", result.RequestsPerSec, minRPS)

	pts.T().Logf("Load test results: %d requests, %.2f RPS, %.2f%% errors",
		result.TotalRequests, result.RequestsPerSec, errorRate*100)
}

// formatBenchmarkResult formats benchmark results for display
func formatBenchmarkResult(result *BenchmarkResult) string {
	base := fmt.Sprintf("%d iterations, %.2f ns/op, %.2f ops/sec",
		result.Iterations, float64(result.NsPerOp), result.ThroughputPerSec)

	if result.BytesPerOp > 0 {
		base += fmt.Sprintf(", %d B/op", result.BytesPerOp)
	}
	if result.AllocsPerOp > 0 {
		base += fmt.Sprintf(", %d allocs/op", result.AllocsPerOp)
	}

	return base
}
