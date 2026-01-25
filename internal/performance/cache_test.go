// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewCacheManager(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	cm := NewCacheManager(logger)

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.stores)
}

func TestCacheManager_CreateStore(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	assert.NotNil(t, store)
	assert.Equal(t, "test-store", store.name)
	assert.Equal(t, 100, store.maxSize)
	assert.Equal(t, 30*time.Second, store.defaultTTL)
}

func TestCacheManager_GetStore(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	// Create a store
	created := cm.CreateStore("test-store", 100, 30*time.Second)

	// Retrieve it
	retrieved := cm.GetStore("test-store")

	assert.NotNil(t, retrieved)
	assert.Equal(t, created.name, retrieved.name)
}

func TestCacheManager_SetAndGet(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	cm.CreateStore("test-store", 100, 30*time.Second)

	// Set a value
	cm.Set("test-store", "key1", "value1")

	// Get it back
	value, exists := cm.Get("test-store", "key1")

	assert.True(t, exists)
	assert.Equal(t, "value1", value)
}

func TestCacheManager_SetWithTTL(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	cm.CreateStore("test-store", 100, 30*time.Second)

	// Set value with custom TTL
	cm.SetWithTTL("test-store", "key1", "value1", 1*time.Second)

	// Should be immediately available
	value, exists := cm.Get("test-store", "key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Should be expired
	_, exists = cm.Get("test-store", "key1")
	assert.False(t, exists)
}

func TestCacheManager_Delete(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	cm.CreateStore("test-store", 100, 30*time.Second)
	cm.Set("test-store", "key1", "value1")

	// Delete the key
	cm.Delete("test-store", "key1")

	// Should not exist
	_, exists := cm.Get("test-store", "key1")
	assert.False(t, exists)
}

func TestCacheManager_Clear(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	cm.CreateStore("test-store", 100, 30*time.Second)
	cm.Set("test-store", "key1", "value1")
	cm.Set("test-store", "key2", "value2")

	// Clear the store
	cm.Clear("test-store")

	// Both should be gone
	_, exists1 := cm.Get("test-store", "key1")
	_, exists2 := cm.Get("test-store", "key2")
	assert.False(t, exists1)
	assert.False(t, exists2)
}

func TestCacheStore_Size(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	// Initial size should be 0
	assert.Equal(t, 0, store.Size())

	// Add items
	store.Set("key1", "value1", 30*time.Second)
	assert.Equal(t, 1, store.Size())

	store.Set("key2", "value2", 30*time.Second)
	assert.Equal(t, 2, store.Size())

	// Delete item
	store.Delete("key1")
	assert.Equal(t, 1, store.Size())
}

func TestCacheStore_Stats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	// Add items
	store.Set("key1", "value1", 30*time.Second)
	store.Set("key2", "value2", 30*time.Second)

	// Get items (should increment hits)
	store.Get("key1")
	store.Get("key2")

	// Try to get non-existent item (should increment misses)
	store.Get("non-existent")

	stats := store.Stats()
	assert.Equal(t, int64(2), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, 2, stats.Size)
}

func TestCacheStore_LRUEviction(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	// Create small store to force evictions
	store := cm.CreateStore("small-store", 2, 30*time.Second)

	// Add items
	store.Set("key1", "value1", 30*time.Second)
	store.Set("key2", "value2", 30*time.Second)
	assert.Equal(t, 2, store.Size())

	// Access key1 to mark it as recently used
	store.Get("key1")

	// Add key3 - should evict key2 (LRU)
	store.Set("key3", "value3", 30*time.Second)
	assert.Equal(t, 2, store.Size())

	// key1 should exist
	_, exists := store.Get("key1")
	assert.True(t, exists)

	// key2 should be evicted
	_, exists = store.Get("key2")
	assert.False(t, exists)

	// key3 should exist
	_, exists = store.Get("key3")
	assert.True(t, exists)
}

func TestCacheStore_ExpiredEntries(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	// Add entry with short TTL
	store.Set("short-ttl", "value", 100*time.Millisecond)

	// Should exist immediately
	_, exists := store.Get("short-ttl")
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, exists = store.Get("short-ttl")
	assert.False(t, exists)
}

func TestCacheStore_MultipleValues(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	// Add multiple values of different types
	store.Set("string", "hello", 30*time.Second)
	store.Set("number", 42, 30*time.Second)
	store.Set("float", 3.14, 30*time.Second)

	// Retrieve and verify
	str, _ := store.Get("string")
	assert.Equal(t, "hello", str)

	num, _ := store.Get("number")
	assert.Equal(t, 42, num)

	f, _ := store.Get("float")
	assert.Equal(t, 3.14, f)
}

func TestCacheStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 1000, 30*time.Second)

	// Concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			key := "key" + string(rune(index))
			store.Set(key, index, 30*time.Second)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all items were added
	assert.Equal(t, 10, store.Size())
}

func TestCacheManager_MultipleStores(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	// Create multiple stores
	store1 := cm.CreateStore("store1", 100, 30*time.Second)
	store2 := cm.CreateStore("store2", 100, 30*time.Second)
	store3 := cm.CreateStore("store3", 100, 30*time.Second)

	// Add values to different stores
	store1.Set("key", "value1", 30*time.Second)
	store2.Set("key", "value2", 30*time.Second)
	store3.Set("key", "value3", 30*time.Second)

	// Verify isolation
	v1, _ := store1.Get("key")
	v2, _ := store2.Get("key")
	v3, _ := store3.Get("key")

	assert.Equal(t, "value1", v1)
	assert.Equal(t, "value2", v2)
	assert.Equal(t, "value3", v3)
}

func TestCacheStore_Clear(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	cm := NewCacheManager(logger)

	store := cm.CreateStore("test-store", 100, 30*time.Second)

	// Add values
	store.Set("key1", "value1", 30*time.Second)
	store.Set("key2", "value2", 30*time.Second)
	store.Set("key3", "value3", 30*time.Second)

	assert.Equal(t, 3, store.Size())

	// Clear
	store.Clear()

	// Should be empty
	assert.Equal(t, 0, store.Size())

	// All items should be gone
	_, exists1 := store.Get("key1")
	_, exists2 := store.Get("key2")
	_, exists3 := store.Get("key3")

	assert.False(t, exists1)
	assert.False(t, exists2)
	assert.False(t, exists3)
}
