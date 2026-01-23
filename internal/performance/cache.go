// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// CacheManager manages multiple cache strategies for SLURM data
type CacheManager struct {
	logger *logrus.Entry

	// Cache stores
	stores map[string]*CacheStore
	mu     sync.RWMutex

	// Metrics
	cacheHits      *prometheus.CounterVec
	cacheMisses    *prometheus.CounterVec
	cacheSize      *prometheus.GaugeVec
	cacheEvictions *prometheus.CounterVec
}

// CacheStore represents a single cache with TTL and size limits
type CacheStore struct {
	name       string
	maxSize    int
	defaultTTL time.Duration

	// Storage
	data        map[string]*CacheEntry
	accessOrder []string
	mu          sync.RWMutex

	// Metrics
	hitCount   int64
	missCount  int64
	evictCount int64
}

// CacheEntry represents a cached item
type CacheEntry struct {
	value       interface{}
	createdAt   time.Time
	ttl         time.Duration
	lastAccess  time.Time
	accessCount int64
}

// NewCacheManager creates a new cache manager
func NewCacheManager(logger *logrus.Entry) *CacheManager {
	cm := &CacheManager{
		logger: logger,
		stores: make(map[string]*CacheStore),
	}

	cm.initMetrics()
	cm.setupDefaultCaches()

	return cm
}

// initMetrics initializes Prometheus metrics for cache monitoring
func (cm *CacheManager) initMetrics() {
	cm.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slurm_exporter_cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_name"},
	)

	cm.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slurm_exporter_cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_name"},
	)

	cm.cacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slurm_exporter_cache_entries",
			Help: "Number of entries in cache",
		},
		[]string{"cache_name"},
	)

	cm.cacheEvictions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slurm_exporter_cache_evictions_total",
			Help: "Total number of cache evictions",
		},
		[]string{"cache_name"},
	)
}

// setupDefaultCaches creates default cache stores for different data types
func (cm *CacheManager) setupDefaultCaches() {
	// Job data cache - shorter TTL for dynamic data
	cm.CreateStore("jobs", 1000, 30*time.Second)

	// Node data cache - medium TTL for semi-static data
	cm.CreateStore("nodes", 500, 2*time.Minute)

	// Partition data cache - longer TTL for static data
	cm.CreateStore("partitions", 100, 5*time.Minute)

	// User/account data cache - long TTL for rarely changing data
	cm.CreateStore("accounts", 200, 10*time.Minute)

	// QoS data cache - very long TTL for configuration data
	cm.CreateStore("qos", 50, 30*time.Minute)

	// Metric metadata cache - cache computed metrics
	cm.CreateStore("metrics", 2000, 1*time.Minute)
}

// CreateStore creates a new cache store
func (cm *CacheManager) CreateStore(name string, maxSize int, defaultTTL time.Duration) *CacheStore {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	store := &CacheStore{
		name:        name,
		maxSize:     maxSize,
		defaultTTL:  defaultTTL,
		data:        make(map[string]*CacheEntry),
		accessOrder: make([]string, 0, maxSize),
	}

	cm.stores[name] = store

	cm.logger.WithFields(logrus.Fields{
		"store":       name,
		"max_size":    maxSize,
		"default_ttl": defaultTTL,
	}).Debug("Created cache store")

	return store
}

// GetStore returns a cache store by name
func (cm *CacheManager) GetStore(name string) *CacheStore {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.stores[name]
}

// Get retrieves a value from the specified cache store
func (cm *CacheManager) Get(storeName, key string) (interface{}, bool) {
	store := cm.GetStore(storeName)
	if store == nil {
		return nil, false
	}

	value, found := store.Get(key)
	if found {
		cm.cacheHits.WithLabelValues(storeName).Inc()
	} else {
		cm.cacheMisses.WithLabelValues(storeName).Inc()
	}

	return value, found
}

// Set stores a value in the specified cache store
func (cm *CacheManager) Set(storeName, key string, value interface{}) {
	cm.SetWithTTL(storeName, key, value, 0) // Use default TTL
}

// SetWithTTL stores a value with a custom TTL
func (cm *CacheManager) SetWithTTL(storeName, key string, value interface{}, ttl time.Duration) {
	store := cm.GetStore(storeName)
	if store == nil {
		return
	}

	store.Set(key, value, ttl)
	cm.cacheSize.WithLabelValues(storeName).Set(float64(store.Size()))
}

// Delete removes a value from the cache
func (cm *CacheManager) Delete(storeName, key string) {
	store := cm.GetStore(storeName)
	if store == nil {
		return
	}

	store.Delete(key)
	cm.cacheSize.WithLabelValues(storeName).Set(float64(store.Size()))
}

// Clear removes all entries from a cache store
func (cm *CacheManager) Clear(storeName string) {
	store := cm.GetStore(storeName)
	if store == nil {
		return
	}

	store.Clear()
	cm.cacheSize.WithLabelValues(storeName).Set(0)
}

// CacheStore methods

// Get retrieves a value from the cache store
func (cs *CacheStore) Get(key string) (interface{}, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	entry, exists := cs.data[key]
	if !exists {
		cs.missCount++
		return nil, false
	}

	// Check if entry has expired
	if cs.isExpired(entry) {
		delete(cs.data, key)
		cs.removeFromAccessOrder(key)
		cs.missCount++
		return nil, false
	}

	// Update access information
	entry.lastAccess = time.Now()
	entry.accessCount++
	cs.updateAccessOrder(key)
	cs.hitCount++

	return entry.value, true
}

// Set stores a value in the cache store
func (cs *CacheStore) Set(key string, value interface{}, ttl time.Duration) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if ttl == 0 {
		ttl = cs.defaultTTL
	}

	// Check if we need to evict entries
	if len(cs.data) >= cs.maxSize {
		cs.evictLRU()
	}

	// Create new entry
	entry := &CacheEntry{
		value:       value,
		createdAt:   time.Now(),
		ttl:         ttl,
		lastAccess:  time.Now(),
		accessCount: 1,
	}

	cs.data[key] = entry
	cs.updateAccessOrder(key)
}

// Delete removes a value from the cache store
func (cs *CacheStore) Delete(key string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.data, key)
	cs.removeFromAccessOrder(key)
}

// Clear removes all entries from the cache store
func (cs *CacheStore) Clear() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.data = make(map[string]*CacheEntry)
	cs.accessOrder = cs.accessOrder[:0]
}

// Size returns the number of entries in the cache
func (cs *CacheStore) Size() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	return len(cs.data)
}

// Stats returns cache statistics
func (cs *CacheStore) Stats() CacheStats {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	totalRequests := cs.hitCount + cs.missCount
	var hitRate float64
	if totalRequests > 0 {
		hitRate = float64(cs.hitCount) / float64(totalRequests)
	}

	return CacheStats{
		Name:       cs.name,
		Size:       len(cs.data),
		MaxSize:    cs.maxSize,
		HitCount:   cs.hitCount,
		MissCount:  cs.missCount,
		EvictCount: cs.evictCount,
		HitRate:    hitRate,
		DefaultTTL: cs.defaultTTL,
	}
}

// isExpired checks if a cache entry has expired
func (cs *CacheStore) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.createdAt) > entry.ttl
}

// evictLRU evicts the least recently used entry
func (cs *CacheStore) evictLRU() {
	if len(cs.accessOrder) == 0 {
		return
	}

	// Find the oldest entry that's not recently accessed
	oldestKey := cs.accessOrder[0]
	delete(cs.data, oldestKey)
	cs.accessOrder = cs.accessOrder[1:]
	cs.evictCount++
}

// updateAccessOrder updates the access order for LRU tracking
func (cs *CacheStore) updateAccessOrder(key string) {
	// Remove from current position
	cs.removeFromAccessOrder(key)

	// Add to end (most recently used)
	cs.accessOrder = append(cs.accessOrder, key)
}

// removeFromAccessOrder removes a key from the access order
func (cs *CacheStore) removeFromAccessOrder(key string) {
	for i, k := range cs.accessOrder {
		if k == key {
			cs.accessOrder = append(cs.accessOrder[:i], cs.accessOrder[i+1:]...)
			break
		}
	}
}

// CleanupExpired removes expired entries from all caches
func (cm *CacheManager) CleanupExpired() {
	cm.mu.RLock()
	stores := make([]*CacheStore, 0, len(cm.stores))
	for _, store := range cm.stores {
		stores = append(stores, store)
	}
	cm.mu.RUnlock()

	for _, store := range stores {
		cleaned := store.cleanupExpired()
		if cleaned > 0 {
			cm.cacheEvictions.WithLabelValues(store.name).Add(float64(cleaned))
			cm.cacheSize.WithLabelValues(store.name).Set(float64(store.Size()))
		}
	}
}

// cleanupExpired removes expired entries from a cache store
func (cs *CacheStore) cleanupExpired() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cleaned := 0
	for key, entry := range cs.data {
		if cs.isExpired(entry) {
			delete(cs.data, key)
			cs.removeFromAccessOrder(key)
			cleaned++
		}
	}

	return cleaned
}

// GetAllStats returns statistics for all cache stores
func (cm *CacheManager) GetAllStats() []CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stats := make([]CacheStats, 0, len(cm.stores))
	for _, store := range cm.stores {
		stats = append(stats, store.Stats())
	}

	return stats
}

// Describe implements prometheus.Collector
func (cm *CacheManager) Describe(ch chan<- *prometheus.Desc) {
	cm.cacheHits.Describe(ch)
	cm.cacheMisses.Describe(ch)
	cm.cacheSize.Describe(ch)
	cm.cacheEvictions.Describe(ch)
}

// Collect implements prometheus.Collector
func (cm *CacheManager) Collect(ch chan<- prometheus.Metric) {
	// Update cache size metrics
	stats := cm.GetAllStats()
	for _, stat := range stats {
		cm.cacheSize.WithLabelValues(stat.Name).Set(float64(stat.Size))
	}

	cm.cacheHits.Collect(ch)
	cm.cacheMisses.Collect(ch)
	cm.cacheSize.Collect(ch)
	cm.cacheEvictions.Collect(ch)
}

// CachedFunction wraps a function with caching
func (cm *CacheManager) CachedFunction(storeName string, fn func(ctx context.Context, key string) (interface{}, error)) func(ctx context.Context, key string) (interface{}, error) {
	return func(ctx context.Context, key string) (interface{}, error) {
		// Try cache first
		if value, found := cm.Get(storeName, key); found {
			return value, nil
		}

		// Call original function
		result, err := fn(ctx, key)
		if err != nil {
			return nil, err
		}

		// Cache the result
		cm.Set(storeName, key, result)

		return result, nil
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	Name       string
	Size       int
	MaxSize    int
	HitCount   int64
	MissCount  int64
	EvictCount int64
	HitRate    float64
	DefaultTTL time.Duration
}
