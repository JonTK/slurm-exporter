// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
)

// DegradationMode represents the current degradation state
type DegradationMode int

const (
	ModeNormal DegradationMode = iota
	ModePartial
	ModeUnavailable
)

// CircuitBreaker implements a circuit breaker pattern for SLURM API calls
type CircuitBreaker struct {
	name         string
	maxFailures  int
	resetTimeout time.Duration

	mu           sync.RWMutex
	failures     int
	lastFailTime time.Time
	state        CircuitBreakerState

	logger *logrus.Entry
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed   CircuitBreakerState = iota // Normal operation
	StateOpen                                // Failures exceeded, rejecting calls
	StateHalfOpen                            // Testing if service recovered
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		logger:       logrus.WithField("component", "circuit_breaker").WithField("name", name),
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check current state
	switch cb.state {
	case StateOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open")
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open for %s", cb.name)
		}

	case StateHalfOpen:
		// In half-open, we allow one call through
		cb.logger.Debug("Circuit breaker in half-open state, testing call")

	case StateClosed:
		// Normal operation, proceed with call
		cb.logger.Debug("Circuit breaker is closed, executing call")
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()

		// If we were half-open, go back to open
		if cb.state == StateHalfOpen {
			cb.state = StateOpen
			cb.logger.Warn("Circuit breaker returning to open after failed test")
		}

		return err
	}

	// Success - reset if needed
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
		cb.logger.Info("Circuit breaker closed after successful test")
	} else if cb.state == StateClosed && cb.failures > 0 {
		cb.failures = 0
		cb.logger.Debug("Circuit breaker failures reset after success")
	}

	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures && cb.state == StateClosed {
		cb.state = StateOpen
		cb.logger.WithField("failures", cb.failures).Warn("Circuit breaker opened due to excessive failures")
	}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == StateOpen
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = StateClosed
	cb.logger.Info("Circuit breaker manually reset")
}

// DegradationManager manages graceful degradation for collectors
type DegradationManager struct {
	config   *config.DegradationConfig
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex

	// Metrics
	degradationMetrics *DegradationMetrics

	// Cached values for degraded mode
	cache   map[string]*CachedMetrics
	cacheMu sync.RWMutex

	logger *logrus.Entry
}

// CachedMetrics holds cached metric values for degraded mode
type CachedMetrics struct {
	CollectorName string
	Metrics       []prometheus.Metric
	CachedAt      time.Time
	TTL           time.Duration
}

// DegradationMetrics tracks degradation events
type DegradationMetrics struct {
	CircuitBreakerState       *prometheus.GaugeVec
	DegradationMode           prometheus.Gauge
	CachedMetricsServed       *prometheus.CounterVec
	FailuresBeforeDegradation *prometheus.HistogramVec
}

// NewDegradationMetrics creates degradation metrics
func NewDegradationMetrics(namespace, subsystem string) *DegradationMetrics {
	return &DegradationMetrics{
		CircuitBreakerState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "circuit_breaker_state",
				Help:      "Current state of circuit breakers (0=closed, 1=open, 2=half-open)",
			},
			[]string{"collector"},
		),
		DegradationMode: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "degradation_mode",
				Help:      "Current degradation mode (0=normal, 1=partial, 2=unavailable)",
			},
		),
		CachedMetricsServed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "cached_metrics_served_total",
				Help:      "Total number of times cached metrics were served",
			},
			[]string{"collector"},
		),
		FailuresBeforeDegradation: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "failures_before_degradation",
				Help:      "Number of failures before entering degraded mode",
				Buckets:   []float64{1, 2, 3, 5, 10, 15, 20, 30},
			},
			[]string{"collector"},
		),
	}
}

// Register registers degradation metrics
func (m *DegradationMetrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.CircuitBreakerState,
		m.DegradationMode,
		m.CachedMetricsServed,
		m.FailuresBeforeDegradation,
	}

	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	return nil
}

// NewDegradationManager creates a new degradation manager
func NewDegradationManager(config *config.DegradationConfig, promRegistry *prometheus.Registry) (*DegradationManager, error) {
	metrics := NewDegradationMetrics("slurm", "degradation")
	if err := metrics.Register(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register degradation metrics: %w", err)
	}

	dm := &DegradationManager{
		config:             config,
		breakers:           make(map[string]*CircuitBreaker),
		degradationMetrics: metrics,
		cache:              make(map[string]*CachedMetrics),
		logger:             logrus.WithField("component", "degradation_manager"),
	}

	// Set initial degradation mode
	dm.degradationMetrics.DegradationMode.Set(float64(ModeNormal))

	return dm, nil
}

// GetCircuitBreaker returns or creates a circuit breaker for a collector
func (dm *DegradationManager) GetCircuitBreaker(collectorName string) *CircuitBreaker {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if breaker, exists := dm.breakers[collectorName]; exists {
		return breaker
	}

	// Create new circuit breaker
	breaker := NewCircuitBreaker(
		collectorName,
		dm.config.MaxFailures,
		dm.config.ResetTimeout,
	)

	dm.breakers[collectorName] = breaker
	return breaker
}

// ExecuteWithDegradation executes a collection with degradation support
func (dm *DegradationManager) ExecuteWithDegradation(
	ctx context.Context,
	collectorName string,
	collectFunc func(context.Context) ([]prometheus.Metric, error),
) ([]prometheus.Metric, error) {
	breaker := dm.GetCircuitBreaker(collectorName)

	// Try to execute with circuit breaker
	var metrics []prometheus.Metric
	var collectErr error

	err := breaker.Call(func() error {
		metrics, collectErr = collectFunc(ctx)
		return collectErr
	})

	// Update circuit breaker state metric
	dm.degradationMetrics.CircuitBreakerState.WithLabelValues(collectorName).Set(float64(breaker.GetState()))

	if err != nil {
		// Check if we should serve cached metrics
		if dm.config.UseCachedMetrics && breaker.IsOpen() {
			cached := dm.getCachedMetrics(collectorName)
			if cached != nil {
				dm.degradationMetrics.CachedMetricsServed.WithLabelValues(collectorName).Inc()
				dm.logger.WithFields(logrus.Fields{
					"collector": collectorName,
					"cached_at": cached.CachedAt,
				}).Info("Serving cached metrics due to circuit breaker open")
				return cached.Metrics, nil
			}
		}

		return nil, err
	}

	// Success - cache the metrics if enabled
	if dm.config.UseCachedMetrics && len(metrics) > 0 {
		dm.cacheMetrics(collectorName, metrics)
	}

	return metrics, nil
}

// cacheMetrics caches metrics for a collector
func (dm *DegradationManager) cacheMetrics(collectorName string, metrics []prometheus.Metric) {
	dm.cacheMu.Lock()
	defer dm.cacheMu.Unlock()

	dm.cache[collectorName] = &CachedMetrics{
		CollectorName: collectorName,
		Metrics:       metrics,
		CachedAt:      time.Now(),
		TTL:           dm.config.CacheTTL,
	}

	dm.logger.WithFields(logrus.Fields{
		"collector":    collectorName,
		"metric_count": len(metrics),
	}).Debug("Cached metrics for degraded mode")
}

// getCachedMetrics retrieves cached metrics if still valid
func (dm *DegradationManager) getCachedMetrics(collectorName string) *CachedMetrics {
	dm.cacheMu.RLock()
	defer dm.cacheMu.RUnlock()

	cached, exists := dm.cache[collectorName]
	if !exists {
		return nil
	}

	// Check if cache is still valid
	if time.Since(cached.CachedAt) > cached.TTL {
		dm.logger.WithField("collector", collectorName).Debug("Cached metrics expired")
		return nil
	}

	return cached
}

// UpdateDegradationMode updates the overall degradation mode based on circuit breaker states
func (dm *DegradationManager) UpdateDegradationMode() {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	openCount := 0
	totalCount := len(dm.breakers)

	for _, breaker := range dm.breakers {
		if breaker.IsOpen() {
			openCount++
		}
	}

	var mode DegradationMode
	if openCount == 0 {
		mode = ModeNormal
	} else if openCount < totalCount {
		mode = ModePartial
	} else {
		mode = ModeUnavailable
	}

	dm.degradationMetrics.DegradationMode.Set(float64(mode))

	if mode != ModeNormal {
		dm.logger.WithFields(logrus.Fields{
			"mode":           mode,
			"open_breakers":  openCount,
			"total_breakers": totalCount,
		}).Warn("System in degraded mode")
	}
}

// GetDegradationStats returns current degradation statistics
func (dm *DegradationManager) GetDegradationStats() map[string]DegradationStats {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	stats := make(map[string]DegradationStats)

	for name, breaker := range dm.breakers {
		breaker.mu.RLock()
		stats[name] = DegradationStats{
			CollectorName: name,
			State:         breaker.state,
			Failures:      breaker.failures,
			LastFailTime:  breaker.lastFailTime,
			HasCache:      dm.hasCachedMetrics(name),
		}
		breaker.mu.RUnlock()
	}

	return stats
}

// hasCachedMetrics checks if valid cached metrics exist
func (dm *DegradationManager) hasCachedMetrics(collectorName string) bool {
	return dm.getCachedMetrics(collectorName) != nil
}

// DegradationStats contains degradation statistics for a collector
type DegradationStats struct {
	CollectorName string
	State         CircuitBreakerState
	Failures      int
	LastFailTime  time.Time
	HasCache      bool
}

// ResetAllBreakers resets all circuit breakers
func (dm *DegradationManager) ResetAllBreakers() {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	for _, breaker := range dm.breakers {
		breaker.Reset()
	}

	dm.UpdateDegradationMode()
	dm.logger.Info("All circuit breakers reset")
}
