// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"fmt"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/sirupsen/logrus"
)

// ExampleJobData represents a simplified SLURM job for demonstration
type ExampleJobData struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	State      string     `json:"state"`
	UserID     string     `json:"user_id"`
	NodeList   string     `json:"node_list"`
	SubmitTime time.Time  `json:"submit_time"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
}

// ExampleCacheUsage demonstrates how to use the intelligent cache in a real-world scenario
func ExampleCacheUsage() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Configure intelligent caching with adaptive TTL
	cacheCfg := config.CachingConfig{
		Intelligent:     true,
		BaseTTL:         2 * time.Minute,
		MaxEntries:      10000,
		CleanupInterval: 5 * time.Minute,
		ChangeTracking:  true,
		AdaptiveTTL: config.AdaptiveTTLConfig{
			Enabled:           true,
			MinTTL:            30 * time.Second,
			MaxTTL:            30 * time.Minute,
			StabilityWindow:   10 * time.Minute,
			VarianceThreshold: 0.1,
			ChangeThreshold:   0.05,
			ExtensionFactor:   2.0,
			ReductionFactor:   0.5,
		},
	}

	// Create intelligent cache
	cache := NewIntelligentCache(cacheCfg, logger)
	defer cache.Close()

	// Example 1: Cache stable data (completed jobs)
	demonstrateStableDataCaching(cache, logger)

	// Example 2: Cache volatile data (running jobs)
	demonstrateVolatileDataCaching(cache, logger)

	// Example 3: Use cache with function wrapper
	demonstrateCachedFunction(cache, logger)

	// Example 4: Monitor cache performance
	demonstrateCacheMonitoring(cache, logger)
}

// demonstrateStableDataCaching shows how the cache adapts TTL for stable data
func demonstrateStableDataCaching(cache *IntelligentCache, logger *logrus.Logger) {
	logger.Info("=== Demonstrating Stable Data Caching ===")

	// Simulate caching a completed job (stable data)
	completedJob := &ExampleJobData{
		ID:         "job_12345",
		Name:       "simulation_run",
		State:      "COMPLETED",
		UserID:     "user123",
		NodeList:   "node[001-004]",
		SubmitTime: time.Now().Add(-2 * time.Hour),
		StartTime:  timePtr(time.Now().Add(-1 * time.Hour)),
		EndTime:    timePtr(time.Now().Add(-30 * time.Minute)),
	}

	jobKey := fmt.Sprintf("job:%s", completedJob.ID)

	// Cache the job data multiple times (simulating repeated requests with same data)
	for i := 0; i < 5; i++ {
		cache.Set(jobKey, completedJob)
		time.Sleep(100 * time.Millisecond)

		// Retrieve from cache
		if cached, found := cache.Get(jobKey); found {
			job, ok := cached.(*ExampleJobData)
			if !ok {
				logger.Error("Failed to cast cached job data to ExampleJobData")
				continue
			}
			logger.WithFields(logrus.Fields{
				"job_id":    job.ID,
				"state":     job.State,
				"iteration": i + 1,
			}).Info("Retrieved stable job from cache")
		}
	}

	// Check TTL adaptation - should be extended due to stability
	if entry, exists := cache.entries[jobKey]; exists {
		logger.WithFields(logrus.Fields{
			"job_id":          completedJob.ID,
			"ttl":             entry.TTL,
			"stability_score": entry.StabilityScore,
			"change_records":  len(entry.ChangeHistory),
		}).Info("Stable data cache entry stats")
	}
}

// demonstrateVolatileDataCaching shows how the cache adapts TTL for changing data
func demonstrateVolatileDataCaching(cache *IntelligentCache, logger *logrus.Logger) {
	logger.Info("=== Demonstrating Volatile Data Caching ===")

	// Simulate caching a running job (volatile data)
	runningJob := &ExampleJobData{
		ID:         "job_67890",
		Name:       "ml_training",
		State:      "RUNNING",
		UserID:     "user456",
		NodeList:   "gpu[001-002]",
		SubmitTime: time.Now().Add(-30 * time.Minute),
		StartTime:  timePtr(time.Now().Add(-25 * time.Minute)),
	}

	jobKey := fmt.Sprintf("job:%s", runningJob.ID)

	// Cache the job data with changing state
	states := []string{"PENDING", "RUNNING", "RUNNING", "COMPLETING", "COMPLETED"}

	for i, state := range states {
		// Update job state to simulate changes
		runningJob.State = state
		if state == "COMPLETED" {
			runningJob.EndTime = timePtr(time.Now())
		}

		cache.Set(jobKey, runningJob)
		time.Sleep(200 * time.Millisecond)

		logger.WithFields(logrus.Fields{
			"job_id":    runningJob.ID,
			"new_state": state,
			"iteration": i + 1,
		}).Info("Cached changing job data")
	}

	// Check TTL adaptation - should be reduced due to instability
	if entry, exists := cache.entries[jobKey]; exists {
		logger.WithFields(logrus.Fields{
			"job_id":          runningJob.ID,
			"ttl":             entry.TTL,
			"stability_score": entry.StabilityScore,
			"change_records":  len(entry.ChangeHistory),
		}).Info("Volatile data cache entry stats")
	}
}

// demonstrateCachedFunction shows how to wrap expensive operations with caching
func demonstrateCachedFunction(cache *IntelligentCache, logger *logrus.Logger) {
	logger.Info("=== Demonstrating Cached Function Wrapper ===")

	// Simulate an expensive SLURM API call
	expensiveSlurmCall := func(ctx context.Context, jobID string) (interface{}, error) {
		logger.WithField("job_id", jobID).Info("Making expensive SLURM API call")

		// Simulate API latency
		time.Sleep(500 * time.Millisecond)

		// Return mock job data
		return &ExampleJobData{
			ID:     jobID,
			Name:   fmt.Sprintf("job_%s", jobID),
			State:  "RUNNING",
			UserID: "api_user",
		}, nil
	}

	// Wrap with cache
	cachedSlurmCall := cache.CachedFunction(expensiveSlurmCall)

	jobID := "api_job_123"

	// First call - should hit the API
	start := time.Now()
	result1, err := cachedSlurmCall(context.Background(), jobID)
	duration1 := time.Since(start)

	if err == nil {
		job, ok := result1.(*ExampleJobData)
		if !ok {
			logger.Error("Failed to cast first call result to ExampleJobData")
			return
		}
		logger.WithFields(logrus.Fields{
			"job_id":   job.ID,
			"duration": duration1,
			"source":   "API",
		}).Info("First call completed")
	}

	// Second call - should hit the cache
	start = time.Now()
	result2, err := cachedSlurmCall(context.Background(), jobID)
	duration2 := time.Since(start)

	if err == nil {
		job, ok := result2.(*ExampleJobData)
		if !ok {
			logger.Error("Failed to cast second call result to ExampleJobData")
			return
		}
		logger.WithFields(logrus.Fields{
			"job_id":   job.ID,
			"duration": duration2,
			"source":   "Cache",
			"speedup":  fmt.Sprintf("%.1fx", float64(duration1)/float64(duration2)),
		}).Info("Second call completed")
	}
}

// demonstrateCacheMonitoring shows how to monitor cache performance
func demonstrateCacheMonitoring(cache *IntelligentCache, logger *logrus.Logger) {
	logger.Info("=== Demonstrating Cache Monitoring ===")

	// Get comprehensive cache metrics
	metrics := cache.GetMetrics()
	logger.WithFields(logrus.Fields{
		"hit_rate":        fmt.Sprintf("%.2f%%", metrics.HitRate*100),
		"total_entries":   metrics.EntryCount,
		"total_size_kb":   metrics.TotalSize / 1024,
		"average_ttl":     metrics.AverageTTL,
		"ttl_extensions":  metrics.TTLExtensions,
		"ttl_reductions":  metrics.TTLReductions,
		"memory_pressure": metrics.MemoryPressure,
	}).Info("Cache metrics")

	// Get detailed stats for debugging
	stats := cache.GetStats()
	logger.WithFields(logrus.Fields{
		"adaptive_ttl_enabled":   stats["adaptive_ttl_enabled"],
		"stability_distribution": stats["stability_distribution"],
		"ttl_distribution":       stats["ttl_distribution"],
	}).Info("Cache debugging stats")

	// Get top cache entries
	topEntries := cache.GetTopEntries(3)
	logger.WithField("count", len(topEntries)).Info("Top cache entries:")
	for i, entry := range topEntries {
		logger.WithFields(logrus.Fields{
			"rank":       i + 1,
			"key":        entry.Key,
			"hit_count":  entry.HitCount,
			"stability":  fmt.Sprintf("%.3f", entry.StabilityScore),
			"ttl":        entry.TTL,
			"size_bytes": entry.Size,
		}).Info("Top entry")
	}
}

// IntegrationExample shows how to integrate intelligent cache with SLURM collectors
func IntegrationExample() {
	logger := logrus.New()

	// Create collector with cache
	cacheCfg := config.CachingConfig{
		Intelligent:     true,
		BaseTTL:         1 * time.Minute,
		MaxEntries:      5000,
		CleanupInterval: 5 * time.Minute,
		ChangeTracking:  true,
		AdaptiveTTL: config.AdaptiveTTLConfig{
			Enabled:           true,
			MinTTL:            15 * time.Second,
			MaxTTL:            15 * time.Minute,
			StabilityWindow:   5 * time.Minute,
			VarianceThreshold: 0.15,
			ChangeThreshold:   0.1,
			ExtensionFactor:   1.5,
			ReductionFactor:   0.7,
		},
	}

	collector := &CachedSlurmCollector{
		cache:  NewIntelligentCache(cacheCfg, logger),
		logger: logger.WithField("component", "cached_collector"),
	}
	defer collector.cache.Close()

	// Example: Cache job data with intelligent TTL
	collector.cacheJobData()

	// Example: Cache node data with different patterns
	collector.cacheNodeData()

	// Show final cache statistics
	metrics := collector.cache.GetMetrics()
	collector.logger.WithFields(logrus.Fields{
		"final_hit_rate":      fmt.Sprintf("%.1f%%", metrics.HitRate*100),
		"total_entries":       metrics.EntryCount,
		"adaptive_extensions": metrics.TTLExtensions,
		"adaptive_reductions": metrics.TTLReductions,
	}).Info("Final cache performance")
}

// cacheJobData demonstrates caching job data with different change patterns
func (c *CachedSlurmCollector) cacheJobData() {
	// Simulate different job states and their caching behavior
	jobs := []*ExampleJobData{
		{ID: "1001", State: "COMPLETED", UserID: "user1"}, // Stable
		{ID: "1002", State: "RUNNING", UserID: "user2"},   // Changing
		{ID: "1003", State: "PENDING", UserID: "user3"},   // Will change
	}

	for _, job := range jobs {
		key := fmt.Sprintf("job:%s", job.ID)

		// Simulate multiple collection cycles
		for cycle := 0; cycle < 3; cycle++ {
			// Stable job stays the same
			if job.State == "COMPLETED" {
				c.cache.Set(key, job)
			} else {
				// Running jobs change state
				if cycle == 1 && job.State == "PENDING" {
					job.State = "RUNNING"
				} else if cycle == 2 && job.State == "RUNNING" {
					job.State = "COMPLETED"
				}
				c.cache.Set(key, job)
			}

			time.Sleep(50 * time.Millisecond)
		}

		// Check final TTL adaptation
		if entry, exists := c.cache.entries[key]; exists {
			c.logger.WithFields(logrus.Fields{
				"job_id":      job.ID,
				"final_state": job.State,
				"ttl":         entry.TTL,
				"stability":   fmt.Sprintf("%.3f", entry.StabilityScore),
			}).Debug("Job cache adaptation")
		}
	}
}

// cacheNodeData demonstrates caching node data (typically more stable)
func (c *CachedSlurmCollector) cacheNodeData() {
	// Simulate node data (usually stable unless nodes go down/up)
	nodes := []map[string]interface{}{
		{"name": "node001", "state": "IDLE", "cpus": 32},
		{"name": "node002", "state": "ALLOCATED", "cpus": 32},
		{"name": "node003", "state": "DOWN", "cpus": 32},
	}

	for _, node := range nodes {
		key := fmt.Sprintf("node:%s", node["name"])

		// Simulate multiple collections - nodes are generally stable
		for cycle := 0; cycle < 4; cycle++ {
			// Occasionally change node state
			if cycle == 2 && node["name"] == "node003" {
				node["state"] = "IDLE" // Node comes back up
			}

			c.cache.Set(key, node)
			time.Sleep(30 * time.Millisecond)
		}
	}
}

// timePtr is a helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// CachedSlurmCollector demonstrates using intelligent cache with SLURM collectors
type CachedSlurmCollector struct {
	cache  *IntelligentCache
	logger *logrus.Entry
}
