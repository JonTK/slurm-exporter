package performance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ConnectionPool manages a pool of SLURM client connections for better performance
type ConnectionPool struct {
	logger       *logrus.Entry
	baseURL      string
	authProvider interface{} // auth.Provider

	// Pool configuration
	minConnections int
	maxConnections int
	maxIdleTime    time.Duration
	connectionTTL  time.Duration

	// Pool state
	mu sync.RWMutex
	// TODO: Unused field - preserved for future connection tracking
	// connections     []*PooledConnection
	availableConns chan *PooledConnection
	totalConns     int

	// Metrics
	activeConns prometheus.Gauge
	idleConns   prometheus.Gauge
	connCreated prometheus.Counter
	connClosed  prometheus.Counter
	connErrors  prometheus.Counter
	acquireTime prometheus.Histogram

	// Background management
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
}

// PooledConnection wraps a SLURM client with pool metadata
type PooledConnection struct {
	client     slurm.SlurmClient
	createdAt  time.Time
	lastUsedAt time.Time
	inUse      bool
	usageCount int64
	mu         sync.Mutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(baseURL string, authProvider interface{}, logger *logrus.Entry) *ConnectionPool {
	pool := &ConnectionPool{
		logger:         logger,
		baseURL:        baseURL,
		authProvider:   authProvider,
		minConnections: 2,
		maxConnections: 20,
		maxIdleTime:    30 * time.Minute,
		connectionTTL:  2 * time.Hour,
		availableConns: make(chan *PooledConnection, 20),
		stopChan:       make(chan struct{}),
	}

	pool.initMetrics()
	pool.startBackgroundCleanup()

	// Pre-populate with minimum connections
	pool.warmupPool()

	return pool
}

// initMetrics initializes Prometheus metrics for the connection pool
func (cp *ConnectionPool) initMetrics() {
	cp.activeConns = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slurm_exporter_active_connections",
		Help: "Number of active SLURM connections",
	})

	cp.idleConns = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "slurm_exporter_idle_connections",
		Help: "Number of idle SLURM connections",
	})

	cp.connCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_connections_created_total",
		Help: "Total number of SLURM connections created",
	})

	cp.connClosed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_connections_closed_total",
		Help: "Total number of SLURM connections closed",
	})

	cp.connErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "slurm_exporter_connection_errors_total",
		Help: "Total number of SLURM connection errors",
	})

	cp.acquireTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "slurm_exporter_connection_acquire_duration_seconds",
		Help:    "Time spent acquiring a connection from the pool",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
	})
}

// startBackgroundCleanup starts the background cleanup routine
func (cp *ConnectionPool) startBackgroundCleanup() {
	cp.cleanupTicker = time.NewTicker(5 * time.Minute)
	go cp.cleanupLoop()
}

// cleanupLoop runs the connection cleanup in the background
func (cp *ConnectionPool) cleanupLoop() {
	for {
		select {
		case <-cp.cleanupTicker.C:
			cp.cleanupExpiredConnections()
		case <-cp.stopChan:
			return
		}
	}
}

// warmupPool pre-creates minimum connections
func (cp *ConnectionPool) warmupPool() {
	for i := 0; i < cp.minConnections; i++ {
		if conn, err := cp.createConnection(); err == nil {
			cp.availableConns <- conn
		} else {
			cp.logger.WithError(err).Warn("Failed to create warmup connection")
		}
	}
}

// createConnection creates a new pooled connection
func (cp *ConnectionPool) createConnection() (*PooledConnection, error) {
	// Create new SLURM client - this would use the actual client constructor
	// For now, we'll use a placeholder
	client, err := cp.newSlurmClient()
	if err != nil {
		cp.connErrors.Inc()
		return nil, fmt.Errorf("failed to create SLURM client: %w", err)
	}

	conn := &PooledConnection{
		client:     client,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      false,
		usageCount: 0,
	}

	cp.mu.Lock()
	cp.totalConns++
	cp.mu.Unlock()

	cp.connCreated.Inc()
	cp.activeConns.Inc()

	cp.logger.WithField("total_connections", cp.totalConns).Debug("Created new connection")

	return conn, nil
}

// newSlurmClient creates a new SLURM client (placeholder implementation)
func (cp *ConnectionPool) newSlurmClient() (slurm.SlurmClient, error) {
	// This would be replaced with actual client creation logic
	// return slurm.NewClient(context.Background(),
	//     slurm.WithBaseURL(cp.baseURL),
	//     slurm.WithAuth(cp.authProvider),
	// )

	// For now, return a mock or interface placeholder
	return nil, fmt.Errorf("client creation not implemented")
}

// Acquire gets a connection from the pool
func (cp *ConnectionPool) Acquire(ctx context.Context) (*PooledConnection, error) {
	start := time.Now()
	defer func() {
		cp.acquireTime.Observe(time.Since(start).Seconds())
	}()

	// Try to get an available connection
	select {
	case conn := <-cp.availableConns:
		if cp.isConnectionValid(conn) {
			conn.mu.Lock()
			conn.inUse = true
			conn.lastUsedAt = time.Now()
			conn.usageCount++
			conn.mu.Unlock()

			cp.idleConns.Dec()
			return conn, nil
		}
		// Connection is invalid, close it and try to create a new one
		cp.closeConnection(conn)

	case <-ctx.Done():
		return nil, ctx.Err()

	default:
		// No available connections, try to create a new one
	}

	// Create new connection if under limit
	cp.mu.RLock()
	canCreate := cp.totalConns < cp.maxConnections
	cp.mu.RUnlock()

	if canCreate {
		conn, err := cp.createConnection()
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}

		conn.mu.Lock()
		conn.inUse = true
		conn.mu.Unlock()

		return conn, nil
	}

	// Wait for an available connection
	select {
	case conn := <-cp.availableConns:
		if cp.isConnectionValid(conn) {
			conn.mu.Lock()
			conn.inUse = true
			conn.lastUsedAt = time.Now()
			conn.usageCount++
			conn.mu.Unlock()

			cp.idleConns.Dec()
			return conn, nil
		}
		cp.closeConnection(conn)
		return nil, fmt.Errorf("no valid connections available")

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release returns a connection to the pool
func (cp *ConnectionPool) Release(conn *PooledConnection) {
	if conn == nil {
		return
	}

	conn.mu.Lock()
	conn.inUse = false
	conn.lastUsedAt = time.Now()
	conn.mu.Unlock()

	// Check if connection should be retired
	if cp.shouldRetireConnection(conn) {
		cp.closeConnection(conn)
		return
	}

	// Return to pool
	select {
	case cp.availableConns <- conn:
		cp.idleConns.Inc()
	default:
		// Pool is full, close the connection
		cp.closeConnection(conn)
	}
}

// isConnectionValid checks if a connection is still valid
func (cp *ConnectionPool) isConnectionValid(conn *PooledConnection) bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	// Check age
	if time.Since(conn.createdAt) > cp.connectionTTL {
		return false
	}

	// Check idle time
	if time.Since(conn.lastUsedAt) > cp.maxIdleTime {
		return false
	}

	// Could add health check here
	// err := conn.client.Info().Ping(context.Background())
	// return err == nil

	return true
}

// shouldRetireConnection checks if a connection should be retired
func (cp *ConnectionPool) shouldRetireConnection(conn *PooledConnection) bool {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	// Retire old connections
	if time.Since(conn.createdAt) > cp.connectionTTL {
		return true
	}

	// Retire heavily used connections
	if conn.usageCount > 10000 {
		return true
	}

	// Keep minimum connections
	cp.mu.RLock()
	totalConns := cp.totalConns
	cp.mu.RUnlock()

	return totalConns > cp.minConnections
}

// closeConnection closes and removes a connection from the pool
func (cp *ConnectionPool) closeConnection(conn *PooledConnection) {
	if conn == nil {
		return
	}

	conn.mu.Lock()
	if conn.client != nil {
		_ = conn.client.Close()
	}
	conn.mu.Unlock()

	cp.mu.Lock()
	cp.totalConns--
	cp.mu.Unlock()

	cp.connClosed.Inc()
	cp.activeConns.Dec()

	cp.logger.WithField("total_connections", cp.totalConns).Debug("Closed connection")
}

// cleanupExpiredConnections removes expired connections from the pool
func (cp *ConnectionPool) cleanupExpiredConnections() {
	cleaned := 0

	// Check available connections
	for len(cp.availableConns) > 0 {
		select {
		case conn := <-cp.availableConns:
			if !cp.isConnectionValid(conn) {
				cp.closeConnection(conn)
				cleaned++
			} else {
				// Put it back
				cp.availableConns <- conn
				goto done
			}
		default:
			goto done
		}
	}

done:
	if cleaned > 0 {
		cp.logger.WithField("cleaned", cleaned).Debug("Cleaned expired connections")
	}
}

// Stats returns current pool statistics
func (cp *ConnectionPool) Stats() PoolStats {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return PoolStats{
		TotalConnections:     cp.totalConns,
		AvailableConnections: len(cp.availableConns),
		ActiveConnections:    cp.totalConns - len(cp.availableConns),
		MinConnections:       cp.minConnections,
		MaxConnections:       cp.maxConnections,
	}
}

// SetPoolSize adjusts the pool size limits
func (cp *ConnectionPool) SetPoolSize(min, max int) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.minConnections = min
	cp.maxConnections = max

	cp.logger.WithFields(logrus.Fields{
		"min_connections": min,
		"max_connections": max,
	}).Info("Updated pool size")
}

// Close shuts down the connection pool
func (cp *ConnectionPool) Close() {
	close(cp.stopChan)
	cp.cleanupTicker.Stop()

	// Close all connections
	for len(cp.availableConns) > 0 {
		conn := <-cp.availableConns
		cp.closeConnection(conn)
	}

	cp.logger.Info("Connection pool closed")
}

// Describe implements prometheus.Collector
func (cp *ConnectionPool) Describe(ch chan<- *prometheus.Desc) {
	cp.activeConns.Describe(ch)
	cp.idleConns.Describe(ch)
	cp.connCreated.Describe(ch)
	cp.connClosed.Describe(ch)
	cp.connErrors.Describe(ch)
	cp.acquireTime.Describe(ch)
}

// Collect implements prometheus.Collector
func (cp *ConnectionPool) Collect(ch chan<- prometheus.Metric) {
	stats := cp.Stats()

	cp.activeConns.Set(float64(stats.ActiveConnections))
	cp.idleConns.Set(float64(stats.AvailableConnections))

	cp.activeConns.Collect(ch)
	cp.idleConns.Collect(ch)
	cp.connCreated.Collect(ch)
	cp.connClosed.Collect(ch)
	cp.connErrors.Collect(ch)
	cp.acquireTime.Collect(ch)
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	TotalConnections     int
	AvailableConnections int
	ActiveConnections    int
	MinConnections       int
	MaxConnections       int
}
