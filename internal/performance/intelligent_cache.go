// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/sirupsen/logrus"
)

// IntelligentCacheEntry represents a cache entry with adaptive TTL capabilities
type IntelligentCacheEntry struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
	LastUsed  time.Time     `json:"last_used"`
	HitCount  int64         `json:"hit_count"`
	Size      int           `json:"size"` // Estimated size in bytes
	TTL       time.Duration `json:"ttl"`

	// Change tracking for adaptive TTL
	ChangeHistory  []ChangeRecord `json:"change_history,omitempty"`
	LastValue      interface{}    `json:"last_value,omitempty"`
	StabilityScore float64        `json:"stability_score"`
}

// ChangeRecord tracks when and how much a cached value changed
type ChangeRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	ChangeScore float64   `json:"change_score"` // 0.0 = no change, 1.0 = complete change
	ValueHash   uint64    `json:"value_hash"`
}

// IntelligentCache implements adaptive TTL caching based on change patterns
type IntelligentCache struct {
	entries map[string]*IntelligentCacheEntry
	config  config.CachingConfig
	logger  *logrus.Entry
	mu      sync.RWMutex

	// Statistics
	hits        int64
	misses      int64
	evictions   int64
	totalSize   int64
	lastCleanup time.Time

	// Metrics
	metrics *IntelligentCacheMetrics

	// Background cleanup
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// IntelligentCacheMetrics provides detailed cache performance metrics
type IntelligentCacheMetrics struct {
	Hits           int64         `json:"hits"`
	Misses         int64         `json:"misses"`
	HitRate        float64       `json:"hit_rate"`
	Evictions      int64         `json:"evictions"`
	TotalSize      int64         `json:"total_size_bytes"`
	EntryCount     int           `json:"entry_count"`
	AverageTTL     time.Duration `json:"average_ttl"`
	TTLExtensions  int64         `json:"ttl_extensions"`
	TTLReductions  int64         `json:"ttl_reductions"`
	LastCleanup    time.Time     `json:"last_cleanup"`
	MemoryPressure bool          `json:"memory_pressure"`

	// TTL distribution
	TTLDistribution map[string]int `json:"ttl_distribution"`
}

// NewIntelligentCache creates a new intelligent cache with adaptive TTL
func NewIntelligentCache(cfg config.CachingConfig, logger *logrus.Logger) *IntelligentCache {
	cache := &IntelligentCache{
		entries:     make(map[string]*IntelligentCacheEntry, cfg.MaxEntries),
		config:      cfg,
		logger:      logger.WithField("component", "intelligent_cache"),
		lastCleanup: time.Now(),
		metrics: &IntelligentCacheMetrics{
			TTLDistribution: make(map[string]int),
		},
		stopCleanup: make(chan struct{}),
	}

	// Start background cleanup if intelligent caching is enabled
	if cfg.Intelligent {
		cache.startBackgroundCleanup()
	}

	cache.logger.WithFields(logrus.Fields{
		"max_entries":      cfg.MaxEntries,
		"base_ttl":         cfg.BaseTTL,
		"cleanup_interval": cfg.CleanupInterval,
		"intelligent":      cfg.Intelligent,
		"adaptive_ttl":     cfg.AdaptiveTTL.Enabled,
	}).Info("Intelligent cache initialized")

	return cache
}

// Get retrieves a value from the cache, updating access statistics
func (c *IntelligentCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		c.recordMiss()
		return nil, false
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.totalSize -= int64(entry.Size)
		c.mu.Unlock()

		c.recordMiss()
		c.logger.WithField("key", key).Debug("Cache entry expired")
		return nil, false
	}

	// Update access statistics
	c.mu.Lock()
	entry.LastUsed = time.Now()
	entry.HitCount++
	c.mu.Unlock()

	c.recordHit()
	return entry.Value, true
}

// Set stores a value in the cache with adaptive TTL calculation
func (c *IntelligentCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	ttl := c.calculateTTL(key, value)

	// Calculate estimated size
	size := c.estimateSize(value)

	// Check if we need to evict entries
	if len(c.entries) >= c.config.MaxEntries {
		c.evictLRU()
	}

	// Check for existing entry to track changes
	var changeHistory []ChangeRecord
	stabilityScore := 0.5 // Default stability

	if existing, exists := c.entries[key]; exists && c.config.ChangeTracking {
		changeHistory = existing.ChangeHistory

		// Record change
		changeScore := c.calculateChangeScore(existing.Value, value)
		changeRecord := ChangeRecord{
			Timestamp:   now,
			ChangeScore: changeScore,
			ValueHash:   c.hashValue(value),
		}

		changeHistory = append(changeHistory, changeRecord)

		// Keep only recent history within stability window
		cutoff := now.Add(-c.config.AdaptiveTTL.StabilityWindow)
		filtered := make([]ChangeRecord, 0, len(changeHistory))
		for _, record := range changeHistory {
			if record.Timestamp.After(cutoff) {
				filtered = append(filtered, record)
			}
		}
		changeHistory = filtered

		// Update stability score
		stabilityScore = c.calculateStabilityScore(changeHistory)

		// Remove old entry size from total
		c.totalSize -= int64(existing.Size)
	}

	// Create new entry
	entry := &IntelligentCacheEntry{
		Key:            key,
		Value:          value,
		CreatedAt:      now,
		ExpiresAt:      now.Add(ttl),
		LastUsed:       now,
		HitCount:       0,
		Size:           size,
		TTL:            ttl,
		ChangeHistory:  changeHistory,
		LastValue:      value,
		StabilityScore: stabilityScore,
	}

	c.entries[key] = entry
	c.totalSize += int64(size)

	// Update TTL distribution metrics
	c.updateTTLDistribution(ttl)

	c.logger.WithFields(logrus.Fields{
		"key":             key,
		"ttl":             ttl,
		"size":            size,
		"stability_score": stabilityScore,
		"change_records":  len(changeHistory),
	}).Debug("Cache entry stored")
}

// calculateTTL determines the optimal TTL for a cache entry using adaptive logic
func (c *IntelligentCache) calculateTTL(key string, value interface{}) time.Duration {
	// Start with base TTL
	ttl := c.config.BaseTTL

	// If adaptive TTL is not enabled, return base TTL
	if !c.config.AdaptiveTTL.Enabled {
		return ttl
	}

	// Check if we have an existing entry with change history
	if existing, exists := c.entries[key]; exists && len(existing.ChangeHistory) > 0 {
		stabilityScore := c.calculateStabilityScore(existing.ChangeHistory)

		// Calculate adaptive TTL based on stability
		if stabilityScore < c.config.AdaptiveTTL.ChangeThreshold {
			// High change rate - reduce TTL
			ttl = time.Duration(float64(ttl) * c.config.AdaptiveTTL.ReductionFactor)
			c.metrics.TTLReductions++
		} else if stabilityScore > (1.0 - c.config.AdaptiveTTL.VarianceThreshold) {
			// High stability - extend TTL
			ttl = time.Duration(float64(ttl) * c.config.AdaptiveTTL.ExtensionFactor)
			c.metrics.TTLExtensions++
		}

		// Ensure TTL is within bounds
		if ttl < c.config.AdaptiveTTL.MinTTL {
			ttl = c.config.AdaptiveTTL.MinTTL
		} else if ttl > c.config.AdaptiveTTL.MaxTTL {
			ttl = c.config.AdaptiveTTL.MaxTTL
		}
	}

	return ttl
}

// calculateStabilityScore computes how stable a cached value has been
func (c *IntelligentCache) calculateStabilityScore(history []ChangeRecord) float64 {
	if len(history) == 0 {
		return 0.5 // Default stability
	}

	// Calculate average change score
	totalChange := 0.0
	for _, record := range history {
		totalChange += record.ChangeScore
	}
	avgChange := totalChange / float64(len(history))

	// Stability is inverse of change rate
	stability := 1.0 - avgChange

	// Consider change frequency
	if len(history) > 1 {
		timespan := history[len(history)-1].Timestamp.Sub(history[0].Timestamp)
		if timespan > 0 {
			frequency := float64(len(history)) / timespan.Hours()
			// Higher frequency reduces stability
			stability *= math.Max(0.1, 1.0-frequency*0.1)
		}
	}

	return math.Max(0.0, math.Min(1.0, stability))
}

// calculateChangeScore computes how much two values differ
func (c *IntelligentCache) calculateChangeScore(oldValue, newValue interface{}) float64 {
	// Simple hash-based comparison for now
	oldHash := c.hashValue(oldValue)
	newHash := c.hashValue(newValue)

	if oldHash == newHash {
		return 0.0 // No change
	}

	// For more sophisticated comparison, we could analyze structural changes
	// For now, any difference is considered a significant change
	return 1.0
}

// hashValue creates a hash of a value for change detection
func (c *IntelligentCache) hashValue(value interface{}) uint64 {
	h := fnv.New64a()

	// Convert value to JSON for consistent hashing
	if data, err := json.Marshal(value); err == nil {
		h.Write(data)
	} else {
		// Fallback to string representation
		_, _ = fmt.Fprintf(h, "%v", value)
	}

	return h.Sum64()
}

// estimateSize provides a rough estimate of value size in bytes
func (c *IntelligentCache) estimateSize(value interface{}) int {
	if data, err := json.Marshal(value); err == nil {
		return len(data)
	}

	// Fallback size estimation
	return len(fmt.Sprintf("%v", value)) * 2 // Rough estimate
}

// evictLRU removes the least recently used entry
func (c *IntelligentCache) evictLRU() {
	if len(c.entries) == 0 {
		return
	}

	var oldestKey string
	oldestTime := time.Now()

	// Find least recently used entry
	for key, entry := range c.entries {
		if entry.LastUsed.Before(oldestTime) {
			oldestTime = entry.LastUsed
			oldestKey = key
		}
	}

	// Remove the entry
	if oldestKey != "" {
		if entry, exists := c.entries[oldestKey]; exists {
			c.totalSize -= int64(entry.Size)
			delete(c.entries, oldestKey)
			c.evictions++

			c.logger.WithFields(logrus.Fields{
				"key":       oldestKey,
				"last_used": oldestTime,
				"size":      entry.Size,
			}).Debug("Cache entry evicted (LRU)")
		}
	}
}

// startBackgroundCleanup starts the background cleanup routine
func (c *IntelligentCache) startBackgroundCleanup() {
	c.cleanupTicker = time.NewTicker(c.config.CleanupInterval)

	go func() {
		for {
			select {
			case <-c.cleanupTicker.C:
				c.cleanup()
			case <-c.stopCleanup:
				c.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes expired entries and updates metrics
func (c *IntelligentCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expired := 0

	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			c.totalSize -= int64(entry.Size)
			delete(c.entries, key)
			expired++
		}
	}

	c.lastCleanup = now
	c.updateMetrics()

	c.logger.WithFields(logrus.Fields{
		"expired_entries": expired,
		"total_entries":   len(c.entries),
		"total_size":      c.totalSize,
	}).Debug("Cache cleanup completed")
}

// updateMetrics recalculates cache metrics
func (c *IntelligentCache) updateMetrics() {
	total := c.hits + c.misses
	if total > 0 {
		c.metrics.HitRate = float64(c.hits) / float64(total)
	}

	c.metrics.Hits = c.hits
	c.metrics.Misses = c.misses
	c.metrics.Evictions = c.evictions
	c.metrics.TotalSize = c.totalSize
	c.metrics.EntryCount = len(c.entries)
	c.metrics.LastCleanup = c.lastCleanup

	// Calculate average TTL
	if len(c.entries) > 0 {
		totalTTL := time.Duration(0)
		for _, entry := range c.entries {
			totalTTL += entry.TTL
		}
		c.metrics.AverageTTL = totalTTL / time.Duration(len(c.entries))
	}

	// Check memory pressure (simple heuristic)
	maxSize := int64(c.config.MaxEntries * 1024)          // Assume 1KB average per entry
	c.metrics.MemoryPressure = c.totalSize > maxSize*8/10 // 80% threshold
}

// updateTTLDistribution updates TTL distribution metrics
func (c *IntelligentCache) updateTTLDistribution(ttl time.Duration) {
	bucket := c.getTTLBucket(ttl)
	c.metrics.TTLDistribution[bucket]++
}

// getTTLBucket categorizes TTL into buckets for metrics
func (c *IntelligentCache) getTTLBucket(ttl time.Duration) string {
	if ttl < time.Minute {
		return "< 1m"
	} else if ttl < 5*time.Minute {
		return "1-5m"
	} else if ttl < 15*time.Minute {
		return "5-15m"
	} else if ttl < time.Hour {
		return "15m-1h"
	} else {
		return "> 1h"
	}
}

// recordHit increments hit counter
func (c *IntelligentCache) recordHit() {
	c.mu.Lock()
	c.hits++
	c.mu.Unlock()
}

// recordMiss increments miss counter
func (c *IntelligentCache) recordMiss() {
	c.mu.Lock()
	c.misses++
	c.mu.Unlock()
}

// GetMetrics returns current cache metrics
func (c *IntelligentCache) GetMetrics() IntelligentCacheMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.updateMetrics()

	// Create a copy to avoid race conditions
	metrics := *c.metrics

	// Copy TTL distribution map
	metrics.TTLDistribution = make(map[string]int)
	for k, v := range c.metrics.TTLDistribution {
		metrics.TTLDistribution[k] = v
	}

	return metrics
}

// GetStats returns detailed cache statistics for debugging
func (c *IntelligentCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Calculate stability distribution
	stabilityBuckets := map[string]int{
		"very_stable":   0, // > 0.8
		"stable":        0, // 0.6-0.8
		"moderate":      0, // 0.4-0.6
		"unstable":      0, // 0.2-0.4
		"very_unstable": 0, // < 0.2
	}

	for _, entry := range c.entries {
		score := entry.StabilityScore
		if score > 0.8 {
			stabilityBuckets["very_stable"]++
		} else if score > 0.6 {
			stabilityBuckets["stable"]++
		} else if score > 0.4 {
			stabilityBuckets["moderate"]++
		} else if score > 0.2 {
			stabilityBuckets["unstable"]++
		} else {
			stabilityBuckets["very_unstable"]++
		}
	}

	return map[string]interface{}{
		"enabled":                c.config.Intelligent,
		"adaptive_ttl_enabled":   c.config.AdaptiveTTL.Enabled,
		"entry_count":            len(c.entries),
		"max_entries":            c.config.MaxEntries,
		"total_size_bytes":       c.totalSize,
		"hits":                   c.hits,
		"misses":                 c.misses,
		"hit_rate":               c.metrics.HitRate,
		"evictions":              c.evictions,
		"ttl_extensions":         c.metrics.TTLExtensions,
		"ttl_reductions":         c.metrics.TTLReductions,
		"average_ttl":            c.metrics.AverageTTL,
		"last_cleanup":           c.lastCleanup,
		"memory_pressure":        c.metrics.MemoryPressure,
		"stability_distribution": stabilityBuckets,
		"ttl_distribution":       c.metrics.TTLDistribution,
	}
}

// GetTopEntries returns the most frequently accessed cache entries
func (c *IntelligentCache) GetTopEntries(limit int) []IntelligentCacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := make([]IntelligentCacheEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		entries = append(entries, *entry)
	}

	// Sort by hit count (descending)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].HitCount > entries[j].HitCount
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	return entries
}

// Clear removes all entries from the cache
func (c *IntelligentCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*IntelligentCacheEntry, c.config.MaxEntries)
	c.totalSize = 0
	c.hits = 0
	c.misses = 0
	c.evictions = 0
	c.metrics.TTLExtensions = 0
	c.metrics.TTLReductions = 0
	c.metrics.TTLDistribution = make(map[string]int)

	c.logger.Info("Cache cleared")
}

// Close stops the background cleanup and releases resources
func (c *IntelligentCache) Close() {
	if c.cleanupTicker != nil {
		close(c.stopCleanup)
		c.cleanupTicker.Stop()
	}

	c.logger.Info("Intelligent cache closed")
}

// CachedFunction wraps a function with intelligent caching
func (c *IntelligentCache) CachedFunction(fn func(ctx context.Context, key string) (interface{}, error)) func(ctx context.Context, key string) (interface{}, error) {
	return func(ctx context.Context, key string) (interface{}, error) {
		// Try cache first
		if value, found := c.Get(key); found {
			return value, nil
		}

		// Call original function
		result, err := fn(ctx, key)
		if err != nil {
			return nil, err
		}

		// Cache the result
		c.Set(key, result)

		return result, nil
	}
}

// IntelligentCacheWrapper provides a simple interface for the intelligent cache
type IntelligentCacheWrapper interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	GetMetrics() IntelligentCacheMetrics
	GetStats() map[string]interface{}
	GetTopEntries(limit int) []IntelligentCacheEntry
	Clear()
	Close()
	CachedFunction(fn func(ctx context.Context, key string) (interface{}, error)) func(ctx context.Context, key string) (interface{}, error)
}

// Ensure IntelligentCache implements IntelligentCacheWrapper
var _ IntelligentCacheWrapper = (*IntelligentCache)(nil)
