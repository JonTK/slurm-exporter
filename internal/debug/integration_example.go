// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package debug

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/adaptive"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/filtering"
	"github.com/jontk/slurm-exporter/internal/health"
	"github.com/jontk/slurm-exporter/internal/tracing"
	"github.com/sirupsen/logrus"
)

// ExampleDebugServer demonstrates how to integrate debug endpoints with a complete system
type ExampleDebugServer struct {
	handler    *DebugHandler
	scheduler  *adaptive.CollectorScheduler
	filter     *filtering.SmartFilter
	tracer     *tracing.CollectionTracer
	healthMgr  *ExampleHealthManager
	collectors *ExampleCollectorManager
	logger     *logrus.Logger
}

// ExampleCollectorManager implements the CollectorDebugger interface
type ExampleCollectorManager struct {
	states  map[string]CollectorState
	metrics map[string]CollectorMetrics
	mu      sync.RWMutex
}

func NewExampleCollectorManager() *ExampleCollectorManager {
	return &ExampleCollectorManager{
		states:  make(map[string]CollectorState),
		metrics: make(map[string]CollectorMetrics),
	}
}

func (ecm *ExampleCollectorManager) GetStates() map[string]CollectorState {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	states := make(map[string]CollectorState)
	for k, v := range ecm.states {
		states[k] = v
	}
	return states
}

func (ecm *ExampleCollectorManager) GetMetrics() map[string]CollectorMetrics {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	metrics := make(map[string]CollectorMetrics)
	for k, v := range ecm.metrics {
		metrics[k] = v
	}
	return metrics
}

func (ecm *ExampleCollectorManager) AddCollector(name string, enabled bool) {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	ecm.states[name] = CollectorState{
		Name:              name,
		Enabled:           enabled,
		LastCollection:    time.Now().Add(-time.Duration(len(ecm.states)) * time.Minute),
		LastDuration:      time.Duration(50+len(ecm.states)*10) * time.Millisecond,
		ConsecutiveErrors: len(ecm.states) % 3, // Vary error counts
		TotalCollections:  int64(100 + len(ecm.states)*50),
		TotalErrors:       int64(len(ecm.states) * 2),
		CurrentInterval:   time.Duration(30+len(ecm.states)*15) * time.Second,
	}

	ecm.metrics[name] = CollectorMetrics{
		MetricsCollected: int64(1000 + len(ecm.metrics)*500),
		MetricsFiltered:  int64(50 + len(ecm.metrics)*25),
		AvgDuration:      time.Duration(75+len(ecm.metrics)*25) * time.Millisecond,
		SuccessRate:      0.95 - float64(len(ecm.metrics))*0.05,
		LastSuccessful:   time.Now().Add(-time.Duration(len(ecm.metrics)*2) * time.Minute),
	}
}

// NewExampleDebugServer creates a complete debug server setup
func NewExampleDebugServer() (*ExampleDebugServer, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create debug configuration
	debugCfg := config.DebugConfig{
		Enabled:     true,
		RequireAuth: false, // Simplified for example
		EnabledEndpoints: []string{
			"collectors", "tracing", "patterns", "scheduler", "health", "runtime", "config",
		},
	}

	// Create debug handler
	handler := NewDebugHandler(debugCfg, logger)
	if !handler.IsEnabled() {
		return nil, fmt.Errorf("debug handler not enabled")
	}

	// Create mock components for demonstration

	// 1. Adaptive Scheduler
	schedulerCfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  2 * time.Minute,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  5 * time.Minute,
	}
	scheduler, err := adaptive.NewCollectorScheduler(schedulerCfg, 30*time.Second, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// 2. Smart Filter
	filterCfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.7,
		CacheSize:      1000,
		LearningWindow: 20,
		VarianceLimit:  100.0,
		CorrelationMin: 0.1,
	}
	filter, err := filtering.NewSmartFilter(filterCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create filter: %w", err)
	}

	// 3. Tracing
	tracingCfg := config.TracingConfig{
		Enabled:    true,
		SampleRate: 0.1,
		Endpoint:   "localhost:4317",
		Insecure:   true,
	}
	tracer, err := tracing.NewCollectionTracer(tracingCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	// 4. Health Manager (mock)
	healthMgr := &ExampleHealthManager{
		checks: make(map[string]health.Check),
		stats:  make(map[string]interface{}),
	}
	healthMgr.initializeChecks()

	// 5. Collector Manager
	collectors := NewExampleCollectorManager()
	collectors.AddCollector("jobs", true)
	collectors.AddCollector("nodes", true)
	collectors.AddCollector("partitions", true)
	collectors.AddCollector("qos", false)

	// Register components with debug handler (cache can be nil for this example)
	handler.RegisterComponents(scheduler, filter, tracer, healthMgr, collectors, nil)

	server := &ExampleDebugServer{
		handler:    handler,
		scheduler:  scheduler,
		filter:     filter,
		tracer:     tracer,
		healthMgr:  healthMgr,
		collectors: collectors,
		logger:     logger,
	}

	logger.Info("Example debug server created successfully")
	return server, nil
}

// Start starts the debug server
func (eds *ExampleDebugServer) Start(addr string) error {
	mux := http.NewServeMux()

	// Register debug routes
	eds.handler.RegisterRoutes(mux)

	// Add a simple status endpoint
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"status":"ok","debug_enabled":%t}`, eds.handler.IsEnabled())
	})

	// Start simulating activity for demonstration
	go eds.simulateActivity()

	eds.logger.WithField("addr", addr).Info("Starting debug server")
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}

// simulateActivity simulates system activity for demonstration
func (eds *ExampleDebugServer) simulateActivity() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	jobCount := 100
	nodeCount := 50

	for range ticker.C {
		// Simulate cluster activity changes
		jobCount += (len(eds.collectors.states) * 10) - 20 // Some variation
		if jobCount < 10 {
			jobCount = 10
		}

		// Update scheduler with activity
		clusterMetrics := map[string]interface{}{
			"job_count":    jobCount,
			"node_count":   nodeCount,
			"queue_length": jobCount / 4,
			"load_avg":     0.5 + float64(jobCount)/200.0,
		}

		eds.scheduler.UpdateActivity(context.Background(), jobCount, nodeCount, clusterMetrics)

		// Update collector states
		eds.updateCollectorStates()

		eds.logger.WithFields(logrus.Fields{
			"job_count":  jobCount,
			"node_count": nodeCount,
		}).Debug("Simulated activity update")
	}
}

// updateCollectorStates updates collector states for demonstration
func (eds *ExampleDebugServer) updateCollectorStates() {
	eds.collectors.mu.Lock()
	defer eds.collectors.mu.Unlock()

	for name, state := range eds.collectors.states {
		if !state.Enabled {
			continue
		}

		// Simulate collection
		state.LastCollection = time.Now()
		state.LastDuration = time.Duration(50+len(name)*5) * time.Millisecond
		state.TotalCollections++

		// Occasionally simulate errors
		if time.Now().Second()%30 == 0 && name == "qos" {
			state.ConsecutiveErrors++
			state.TotalErrors++
		} else if state.ConsecutiveErrors > 0 {
			state.ConsecutiveErrors = 0 // Recovery
		}

		// Update current interval from scheduler
		state.CurrentInterval = eds.scheduler.GetCollectionInterval(name)

		eds.collectors.states[name] = state

		// Update metrics
		if metrics, exists := eds.collectors.metrics[name]; exists {
			metrics.MetricsCollected += 50
			if state.ConsecutiveErrors == 0 {
				metrics.MetricsFiltered += 5
				metrics.LastSuccessful = time.Now()
			}
			metrics.SuccessRate = float64(state.TotalCollections-state.TotalErrors) / float64(state.TotalCollections)
			eds.collectors.metrics[name] = metrics
		}
	}
}

// ExampleHealthManager implements the HealthDebugger interface
type ExampleHealthManager struct {
	checks map[string]health.Check
	stats  map[string]interface{}
	mu     sync.RWMutex
}

func (ehm *ExampleHealthManager) GetStats() map[string]interface{} {
	ehm.mu.RLock()
	defer ehm.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range ehm.stats {
		stats[k] = v
	}
	return stats
}

func (ehm *ExampleHealthManager) GetChecks() map[string]health.Check {
	ehm.mu.RLock()
	defer ehm.mu.RUnlock()

	checks := make(map[string]health.Check)
	for k, v := range ehm.checks {
		checks[k] = v
	}
	return checks
}

func (ehm *ExampleHealthManager) initializeChecks() {
	ehm.mu.Lock()
	defer ehm.mu.Unlock()

	// Create sample health checks
	ehm.checks["slurm_api"] = health.Check{
		Status:      health.StatusHealthy,
		Message:     "SLURM API is responding normally",
		LastChecked: time.Now(),
		Metadata: map[string]string{
			"endpoint": "https://slurm.cluster.local:6820",
			"version":  "23.02.0",
		},
	}

	ehm.checks["memory"] = health.Check{
		Status:      health.StatusDegraded,
		Message:     "Memory usage elevated: 856 MB (threshold: 1024 MB)",
		LastChecked: time.Now(),
		Metadata: map[string]string{
			"current_mb":   "856",
			"threshold_mb": "1024",
			"usage_pct":    "83.6",
		},
	}

	ehm.checks["disk"] = health.Check{
		Status:      health.StatusHealthy,
		Message:     "Disk usage normal: 45.2%",
		LastChecked: time.Now(),
		Metadata: map[string]string{
			"path":      "/var/lib/slurm-exporter",
			"used_pct":  "45.2",
			"threshold": "85.0",
		},
	}

	ehm.stats = map[string]interface{}{
		"total_checks":     len(ehm.checks),
		"healthy_checks":   1,
		"degraded_checks":  1,
		"unhealthy_checks": 0,
		"last_updated":     time.Now(),
	}
}

// ExampleUsage demonstrates how to use the debug endpoints
func ExampleUsage() {
	// Create and start the example debug server
	server, err := NewExampleDebugServer()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create debug server")
	}

	// Start the server (this will block)
	addr := ":8080"
	logrus.WithField("addr", addr).Info("Starting example debug server")
	logrus.Info("Available endpoints:")
	logrus.Info("  http://localhost:8080/debug - Main debug page")
	logrus.Info("  http://localhost:8080/debug/collectors - Collector status")
	logrus.Info("  http://localhost:8080/debug/tracing - Tracing information")
	logrus.Info("  http://localhost:8080/debug/patterns - Smart filter patterns")
	logrus.Info("  http://localhost:8080/debug/scheduler - Adaptive scheduler state")
	logrus.Info("  http://localhost:8080/debug/health - Health check results")
	logrus.Info("  http://localhost:8080/debug/runtime - Go runtime statistics")
	logrus.Info("  http://localhost:8080/debug/config - Configuration information")
	logrus.Info("  Add ?format=json to any endpoint for JSON output")

	if err := server.Start(addr); err != nil {
		logrus.WithError(err).Fatal("Debug server failed")
	}
}

// ProductionIntegrationExample shows how to integrate debug endpoints in a production setup
func ProductionIntegrationExample() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Production debug configuration (with authentication)
	debugCfg := config.DebugConfig{
		Enabled:     true,
		RequireAuth: true,
		Username:    "admin",
		Password:    "secure-debug-password",
		EnabledEndpoints: []string{
			"collectors", "health", "runtime", // Limited endpoints for production
		},
	}

	handler := NewDebugHandler(debugCfg, logger)
	if !handler.IsEnabled() {
		logger.Warn("Debug endpoints disabled")
		return
	}

	// In production, you would register your actual components
	// handler.RegisterComponents(realScheduler, realFilter, realTracer, realHealth, realCollectors, realCache)

	// Create a separate debug server (don't mix with metrics server)
	debugMux := http.NewServeMux()
	handler.RegisterRoutes(debugMux)

	// Add custom production debug endpoints
	debugMux.HandleFunc("/debug/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"version":"1.0.0","build":"production","debug_enabled":true}`)
	})

	// Start debug server on a different port
	debugAddr := ":8081"
	logger.WithField("addr", debugAddr).Info("Starting production debug server")

	go func() {
		debugServer := &http.Server{
			Addr:              debugAddr,
			Handler:           debugMux,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		}
		if err := debugServer.ListenAndServe(); err != nil {
			logger.WithError(err).Error("Debug server failed")
		}
	}()

	// Your main application would continue here...
	logger.Info("Production debug endpoints available on :8081 (auth required)")
	logger.Info("Use: curl -u admin:secure-debug-password http://localhost:8081/debug")
}
