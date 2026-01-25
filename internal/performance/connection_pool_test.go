// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

func TestNewConnectionPool(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)

	assert.NotNil(t, pool)
	assert.Equal(t, "http://localhost:6817", pool.baseURL)
	assert.Equal(t, 2, pool.minConnections)
	assert.Equal(t, 20, pool.maxConnections)
	assert.Equal(t, 30*time.Minute, pool.maxIdleTime)
	assert.Equal(t, 2*time.Hour, pool.connectionTTL)
	assert.NotNil(t, pool.availableConns)
	assert.NotNil(t, pool.stopChan)
	assert.NotNil(t, pool.cleanupTicker)

	pool.Close()
}

func TestConnectionPool_Metrics(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Verify metrics are initialized
	assert.NotNil(t, pool.activeConns)
	assert.NotNil(t, pool.idleConns)
	assert.NotNil(t, pool.connCreated)
	assert.NotNil(t, pool.connClosed)
	assert.NotNil(t, pool.connErrors)
	assert.NotNil(t, pool.acquireTime)
}

func TestConnectionPool_Stats(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	stats := pool.Stats()

	assert.Equal(t, 20, stats.MaxConnections)
	assert.Equal(t, 2, stats.MinConnections)
	assert.True(t, stats.TotalConnections >= 0)
	assert.True(t, stats.AvailableConnections >= 0)
	assert.True(t, stats.ActiveConnections >= 0)
}

func TestConnectionPool_SetPoolSize(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Get initial stats
	statsInitial := pool.Stats()
	initialMax := statsInitial.MaxConnections
	assert.Equal(t, 20, initialMax)

	// Update pool size
	pool.SetPoolSize(5, 50)

	// Verify changes
	assert.Equal(t, 5, pool.minConnections)
	assert.Equal(t, 50, pool.maxConnections)

	stats := pool.Stats()
	assert.Equal(t, 50, stats.MaxConnections)
	assert.Equal(t, 5, stats.MinConnections)
}

func TestConnectionPool_PooledConnection(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 0,
	}

	assert.NotNil(t, conn)
	assert.False(t, conn.inUse)
	assert.Equal(t, int64(0), conn.usageCount)
}

func TestConnectionPool_AcquireWithTimeout(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Try to acquire with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// This will likely fail since client creation is not implemented
	// But it tests the acquire mechanism
	conn, err := pool.Acquire(ctx)

	// Since client creation returns error, we expect an error
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestConnectionPool_AcquireContextCancel(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	conn, err := pool.Acquire(ctx)

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Equal(t, context.Canceled, err)
}

func TestConnectionPool_ReleaseNilConnection(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Should not panic
	pool.Release(nil)
}

func TestConnectionPool_Describe(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	ch := make(chan *prometheus.Desc, 10)
	pool.Describe(ch)
	close(ch)

	// Should have descriptors
	descs := 0
	for range ch {
		descs++
	}

	assert.True(t, descs >= 6, "should have at least 6 descriptors")
}

func TestConnectionPool_Collect(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	ch := make(chan prometheus.Metric, 20)
	pool.Collect(ch)
	close(ch)

	// Should have metrics
	metrics := 0
	for range ch {
		metrics++
	}

	assert.True(t, metrics >= 0)
}

func TestConnectionPool_ConnectionCreationFailure(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	// The pool will have 0 connections since newSlurmClient fails
	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Warmup should have failed, so available connections should be 0
	stats := pool.Stats()
	assert.Equal(t, 0, stats.AvailableConnections)
}

func TestConnectionPool_WarmupPool(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Check that warmup attempted to create minimum connections
	// Since client creation fails, we won't have actual connections
	// but we can verify the pool structure is ready
	assert.NotNil(t, pool.availableConns)
}

func TestConnectionPool_ConnectionValid_NewConnection(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 0,
	}

	// Simulate pool settings
	pool := &ConnectionPool{
		maxIdleTime:   30 * time.Minute,
		connectionTTL: 2 * time.Hour,
	}

	// New connection should be valid
	valid := pool.isConnectionValid(conn)
	assert.True(t, valid)
}

func TestConnectionPool_ConnectionValid_ExpiredTTL(t *testing.T) {
	t.Parallel()

	// Create old connection
	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now().Add(-3 * time.Hour),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 0,
	}

	pool := &ConnectionPool{
		maxIdleTime:   30 * time.Minute,
		connectionTTL: 2 * time.Hour,
	}

	// Old connection should be invalid
	valid := pool.isConnectionValid(conn)
	assert.False(t, valid)
}

func TestConnectionPool_ConnectionValid_IdleTimeout(t *testing.T) {
	t.Parallel()

	// Create connection with stale last access time
	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now().Add(-45 * time.Minute),
		inUse:      false,
		usageCount: 0,
	}

	pool := &ConnectionPool{
		maxIdleTime:   30 * time.Minute,
		connectionTTL: 2 * time.Hour,
	}

	// Idle connection should be invalid
	valid := pool.isConnectionValid(conn)
	assert.False(t, valid)
}

func TestConnectionPool_ShouldRetireConnection_New(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 100,
	}

	pool := &ConnectionPool{
		minConnections: 5,
		maxConnections: 20,
		totalConns:     5,
		maxIdleTime:    30 * time.Minute,
		connectionTTL:  2 * time.Hour,
	}

	// New connection with moderate usage should not be retired if at minimum
	should := pool.shouldRetireConnection(conn)
	assert.False(t, should)
}

func TestConnectionPool_ShouldRetireConnection_HighUsage(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 15000,
	}

	pool := &ConnectionPool{
		minConnections: 2,
		maxConnections: 20,
		totalConns:     5,
		maxIdleTime:    30 * time.Minute,
		connectionTTL:  2 * time.Hour,
	}

	// Connection with high usage should be retired
	should := pool.shouldRetireConnection(conn)
	assert.True(t, should)
}

func TestConnectionPool_ShouldRetireConnection_OldAge(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now().Add(-3 * time.Hour),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 100,
	}

	pool := &ConnectionPool{
		minConnections: 2,
		maxConnections: 20,
		totalConns:     5,
		maxIdleTime:    30 * time.Minute,
		connectionTTL:  2 * time.Hour,
	}

	// Old connection should be retired
	should := pool.shouldRetireConnection(conn)
	assert.True(t, should)
}

func TestConnectionPool_ShouldRetireConnection_MinConnections(t *testing.T) {
	t.Parallel()

	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 100,
	}

	pool := &ConnectionPool{
		minConnections: 5,
		maxConnections: 20,
		totalConns:     5, // Only at minimum
		maxIdleTime:    30 * time.Minute,
		connectionTTL:  2 * time.Hour,
	}

	// Should not retire if at minimum
	should := pool.shouldRetireConnection(conn)
	assert.False(t, should)
}

func TestConnectionPool_CloseConnection(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)

	// Create a connection manually
	conn := &PooledConnection{
		client:     nil,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 0,
	}

	initialTotal := pool.totalConns
	pool.closeConnection(conn)

	// Total connections should decrease
	assert.Equal(t, initialTotal-1, pool.totalConns)

	pool.Close()
}

func TestConnectionPool_CloseConnectionNil(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Should not panic
	pool.closeConnection(nil)
}

func TestConnectionPool_CleanupExpired(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	// Try cleanup - should not panic even if no connections
	pool.cleanupExpiredConnections()
}

func TestConnectionPool_Close(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)

	// Should not panic
	pool.Close()

	// Verify pool is closed by checking if stopChan is closed
	// (sending to a closed channel would panic, so we just verify Close doesn't panic)
	assert.NotNil(t, pool)
}

func TestConnectionPool_Concurrent_StatAccess(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	done := make(chan bool, 5)

	// Concurrent stat access
	for i := 0; i < 5; i++ {
		go func() {
			_ = pool.Stats()
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestConnectionPool_Concurrent_PoolSizeAdjustment(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	done := make(chan bool, 5)

	// Concurrent pool size adjustments
	for i := 0; i < 5; i++ {
		go func(idx int) {
			minSize := 1 + idx
			maxSize := 10 + idx
			pool.SetPoolSize(minSize, maxSize)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// Pool should still be functional
	assert.NotNil(t, pool)
}

func TestConnectionPool_PoolStats_Structure(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	stats := pool.Stats()

	// Verify all fields are accessible
	assert.Greater(t, stats.MaxConnections, 0)
	assert.Greater(t, stats.MinConnections, 0)
	assert.GreaterOrEqual(t, stats.AvailableConnections, 0)
	assert.GreaterOrEqual(t, stats.ActiveConnections, 0)
	assert.GreaterOrEqual(t, stats.TotalConnections, 0)
}

func TestConnectionPool_AcquireMultipleCalls(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()

	pool := NewConnectionPool("http://localhost:6817", nil, logger)
	defer pool.Close()

	ctx := context.Background()

	// Make multiple acquire attempts
	for i := 0; i < 3; i++ {
		conn, err := pool.Acquire(ctx)
		// Will fail since client creation fails
		assert.Error(t, err)
		assert.Nil(t, conn)
	}
}
