// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/health"
)

// startTime tracks when the server package was loaded
var startTime = time.Now()

// RegistryInterface defines the methods needed by the server from the registry
type RegistryInterface interface {
	GetStats() map[string]collector.CollectorState
	CollectAll(ctx context.Context) error
	GetPerformanceStats() map[string]*collector.CollectorPerformanceStats
}

// HTTPMetrics holds HTTP-related metrics
type HTTPMetrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestSize      *prometheus.HistogramVec
	responseSize     *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

// NewHTTPMetrics creates HTTP metrics
func NewHTTPMetrics() *HTTPMetrics {
	return &HTTPMetrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "slurm_exporter",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm_exporter",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		requestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm_exporter",
				Subsystem: "http",
				Name:      "request_size_bytes",
				Help:      "HTTP request size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"method", "path"},
		),
		responseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "slurm_exporter",
				Subsystem: "http",
				Name:      "response_size_bytes",
				Help:      "HTTP response size in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"method", "path"},
		),
		requestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "slurm_exporter",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),
	}
}

// Register registers HTTP metrics with Prometheus
func (m *HTTPMetrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.requestsTotal,
		m.requestDuration,
		m.requestSize,
		m.responseSize,
		m.requestsInFlight,
	}

	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	return nil
}

// Server represents the HTTP server.
type Server struct {
	config         *config.Config
	logger         *logrus.Logger
	server         *http.Server
	registry       RegistryInterface
	promRegistry   *prometheus.Registry
	httpMetrics    *HTTPMetrics
	healthChecker  *health.HealthChecker
	isShuttingDown bool
}

// New creates a new server instance.
func New(cfg *config.Config, logger *logrus.Logger, registry RegistryInterface, promRegistry *prometheus.Registry) (*Server, error) {
	// Use provided Prometheus registry or create a new one
	if promRegistry == nil {
		promRegistry = prometheus.NewRegistry()
	}

	// Create HTTP metrics
	httpMetrics := NewHTTPMetrics()
	if err := httpMetrics.Register(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register HTTP metrics: %w", err)
	}

	// Create health checker
	healthChecker := health.NewHealthChecker(logger)

	s := &Server{
		config:        cfg,
		logger:        logger,
		registry:      registry,
		promRegistry:  promRegistry,
		httpMetrics:   httpMetrics,
		healthChecker: healthChecker,
	}

	// Setup health checks
	s.setupHealthChecks()

	// Create HTTP handler and setup routes with middleware
	handler := s.setupRoutes()

	// Configure HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Configure TLS if enabled
	if cfg.Server.TLS.Enabled {
		tlsConfig, err := s.createTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		server.TLSConfig = tlsConfig
	}

	s.server = server
	return s, nil
}

// createTLSConfig creates and configures TLS settings
func (s *Server) createTLSConfig() (*tls.Config, error) {
	cfg := s.config.Server.TLS

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Default to TLS 1.2
	}

	// Configure minimum TLS version if specified (validate before loading certificates)
	if cfg.MinVersion != "" {
		switch cfg.MinVersion {
		case "1.0":
			tlsConfig.MinVersion = tls.VersionTLS10
		case "1.1":
			tlsConfig.MinVersion = tls.VersionTLS11
		case "1.2":
			tlsConfig.MinVersion = tls.VersionTLS12
		case "1.3":
			tlsConfig.MinVersion = tls.VersionTLS13
		default:
			return nil, fmt.Errorf("unsupported TLS version: %s", cfg.MinVersion)
		}
	}

	// Validate required files
	if cfg.CertFile == "" {
		return nil, fmt.Errorf("TLS cert_file is required when TLS is enabled")
	}
	if cfg.KeyFile == "" {
		return nil, fmt.Errorf("TLS key_file is required when TLS is enabled")
	}

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	// Configure cipher suites if specified
	if len(cfg.CipherSuites) > 0 {
		cipherSuites := make([]uint16, 0, len(cfg.CipherSuites))
		cipherMap := map[string]uint16{
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":       tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":       tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":     tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":     tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":      tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		}

		for _, suite := range cfg.CipherSuites {
			if cipherID, ok := cipherMap[suite]; ok {
				cipherSuites = append(cipherSuites, cipherID)
			} else {
				s.logger.WithField("cipher_suite", suite).Warn("Unknown cipher suite, skipping")
			}
		}

		if len(cipherSuites) > 0 {
			tlsConfig.CipherSuites = cipherSuites
		}
	}

	s.logger.WithFields(logrus.Fields{
		"min_version":   cfg.MinVersion,
		"cipher_suites": len(cfg.CipherSuites),
		"cert_file":     cfg.CertFile,
		"key_file":      cfg.KeyFile,
	}).Info("TLS configuration created")

	return tlsConfig, nil
}

// setupHealthChecks configures health check functions
func (s *Server) setupHealthChecks() {
	// Basic service liveness check
	s.healthChecker.RegisterCheck("service", func(ctx context.Context) health.Check {
		return health.Check{
			Status:  health.StatusHealthy,
			Message: "Service is running",
		}
	})

	// Registry availability check
	s.healthChecker.RegisterCheck("registry", func(ctx context.Context) health.Check {
		if s.registry == nil {
			return health.Check{
				Status: health.StatusUnhealthy,
				Error:  "Collector registry is not available",
			}
		}

		stats := s.registry.GetStats()
		enabledCount := 0
		for _, stat := range stats {
			if stat.Enabled {
				enabledCount++
			}
		}

		if enabledCount == 0 {
			return health.Check{
				Status:  health.StatusDegraded,
				Message: "No collectors are enabled",
			}
		}

		return health.Check{
			Status:  health.StatusHealthy,
			Message: fmt.Sprintf("%d collectors enabled", enabledCount),
			Metadata: map[string]string{
				"enabled_collectors": fmt.Sprintf("%d", enabledCount),
				"total_collectors":   fmt.Sprintf("%d", len(stats)),
			},
		}
	})

	// Metrics endpoint self-check
	s.healthChecker.RegisterCheck("metrics_endpoint", health.MetricsEndpointCheck(
		fmt.Sprintf("http://localhost%s%s", s.config.Server.Address, s.config.Server.MetricsPath),
		&http.Client{Timeout: 10 * time.Second},
	))

	// TODO: Add SLURM connectivity check when SLURM client is available
	// This would be added in the main.go where we have access to the SLURM client
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints using our health checker
	mux.Handle("/health", s.healthChecker.HealthHandler())
	mux.Handle("/ready", s.healthChecker.ReadinessHandler())
	mux.Handle("/live", s.healthChecker.LivenessHandler())

	// Legacy health endpoints (for backwards compatibility)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)

	// Metrics endpoint
	mux.Handle(s.config.Server.MetricsPath, s.createMetricsHandler())

	// Root endpoint with basic info
	mux.HandleFunc("/", s.handleRoot)

	// Debug endpoints (for troubleshooting)
	mux.HandleFunc("/debug/health", s.handleDebugHealth)
	mux.HandleFunc("/debug/collectors", s.handleDebugCollectors)
	mux.HandleFunc("/debug/performance", s.handleDebugPerformance)

	// Apply middleware to all routes
	return s.CombinedMiddleware(mux)
}

// Start starts the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	scheme := "HTTP"
	if s.config.Server.TLS.Enabled {
		scheme = "HTTPS"
	}

	s.logger.WithFields(logrus.Fields{
		"address": s.config.Server.Address,
		"scheme":  scheme,
		"tls":     s.config.Server.TLS.Enabled,
	}).Info("Starting HTTP server")

	go func() {
		<-ctx.Done()
		s.logger.Info("Context cancelled, shutting down server")
		_ = s.server.Shutdown(context.Background())
	}()

	var err error
	if s.config.Server.TLS.Enabled {
		// Start HTTPS server
		err = s.server.ListenAndServeTLS(s.config.Server.TLS.CertFile, s.config.Server.TLS.KeyFile)
	} else {
		// Start HTTP server
		err = s.server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	s.isShuttingDown = true
	return s.server.Shutdown(ctx)
}

// IsShuttingDown returns whether the server is in shutdown mode
func (s *Server) IsShuttingDown() bool {
	return s.isShuttingDown
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "health_handler")
	logger.Debug("Health check requested")

	// Check if request context is cancelled
	select {
	case <-r.Context().Done():
		logger.Debug("Request cancelled before processing")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
	}

	// Simple health check - always returns OK
	// Could be extended to check dependencies
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// handleReady handles the readiness check endpoint
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "ready_handler")
	logger.Debug("Readiness check requested")

	// Check if request context is cancelled
	select {
	case <-r.Context().Done():
		logger.Debug("Request cancelled before processing")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
	}

	// Not ready if shutting down
	if s.isShuttingDown {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("Server is shutting down"))
		return
	}

	// Check if collectors are ready
	if s.registry != nil {
		stats := s.registry.GetStats()

		// Check for cancellation during stats gathering
		select {
		case <-r.Context().Done():
			logger.Debug("Request cancelled during collector stats check")
			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		// Consider ready if at least one collector is enabled
		ready := false
		enabledCount := 0
		for _, stat := range stats {
			if stat.Enabled {
				ready = true
				enabledCount++
			}
		}

		logger.WithFields(logrus.Fields{
			"total_collectors":   len(stats),
			"enabled_collectors": enabledCount,
			"ready":              ready,
			"shutting_down":      s.isShuttingDown,
		}).Debug("Collector status checked")

		if !ready {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("No collectors enabled"))
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Ready"))
}

// handleRoot handles the root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "root_handler")
	logger.Debug("Root endpoint requested")

	// Check if request context is cancelled
	select {
	case <-r.Context().Done():
		logger.Debug("Request cancelled before processing")
		http.Error(w, "Request cancelled", http.StatusRequestTimeout)
		return
	default:
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	html := `<!DOCTYPE html>
<html>
<head>
    <title>SLURM Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 10px 0; }
        .endpoint a { text-decoration: none; color: #0066cc; }
        .endpoint a:hover { text-decoration: underline; }
        .stats { background-color: #f5f5f5; padding: 15px; margin: 20px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>SLURM Prometheus Exporter</h1>
    <p>This is a Prometheus exporter for SLURM workload manager metrics.</p>

    <h2>Available Endpoints</h2>
    <div class="endpoint">📊 <a href="%s">Metrics</a> - Prometheus metrics endpoint</div>
    <div class="endpoint">❤️ <a href="/health">Health</a> - Health check endpoint</div>
    <div class="endpoint">⚡ <a href="/ready">Ready</a> - Readiness check endpoint</div>

    <div class="stats">
        <h3>Collector Status</h3>
        %s
    </div>
</body>
</html>`

	// Generate collector status
	collectorStatus := "No collectors configured"
	if s.registry != nil {
		stats := s.registry.GetStats()

		// Check for cancellation during stats processing
		select {
		case <-r.Context().Done():
			logger.Debug("Request cancelled during collector status generation")
			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		if len(stats) > 0 {
			collectorStatus = "<ul>"
			for name, stat := range stats {
				// Check for cancellation in the loop (for large number of collectors)
				select {
				case <-r.Context().Done():
					logger.Debug("Request cancelled during collector status loop")
					http.Error(w, "Request cancelled", http.StatusRequestTimeout)
					return
				default:
				}

				status := "✅ Enabled"
				if !stat.Enabled {
					status = "❌ Disabled"
				}
				collectorStatus += fmt.Sprintf("<li><strong>%s:</strong> %s</li>", name, status)
			}
			collectorStatus += "</ul>"
		}
	}

	content := fmt.Sprintf(html, s.config.Server.MetricsPath, collectorStatus)
	_, _ = w.Write([]byte(content))
}

// createMetricsHandler creates the Prometheus metrics handler
func (s *Server) createMetricsHandler() http.Handler {
	// Create a custom gatherer that collects from our registry
	gatherer := prometheus.Gatherers{
		s.promRegistry,
		prometheus.DefaultGatherer, // Include Go runtime metrics
	}

	// Create promhttp handler with custom configuration
	handler := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		ErrorLog:      s.logger,
		ErrorHandling: promhttp.ContinueOnError,
		Timeout:       30 * time.Second,
	})

	// Wrap with collection triggering
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request context is already cancelled
		select {
		case <-r.Context().Done():
			s.logger.WithField("component", "metrics_handler").Debug("Request cancelled before processing")
			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		// Trigger collection from all collectors if registry is available
		if s.registry != nil {
			// Use the request context (which already has timeout from middleware)
			collectionCtx := r.Context()

			s.logger.WithField("component", "metrics_handler").Debug("Starting metrics collection")

			if err := s.registry.CollectAll(collectionCtx); err != nil {
				if collectionCtx.Err() == context.DeadlineExceeded {
					s.logger.WithField("component", "metrics_handler").Warn("Metrics collection timed out")
				} else if collectionCtx.Err() == context.Canceled {
					s.logger.WithField("component", "metrics_handler").Debug("Metrics collection cancelled")
					http.Error(w, "Request cancelled", http.StatusRequestTimeout)
					return
				} else {
					s.logger.WithFields(logrus.Fields{
						"component": "metrics_handler",
						"error":     err.Error(),
					}).Warn("Failed to collect all metrics")
				}
				// Continue serving cached/existing metrics even on timeout/error
			} else {
				s.logger.WithField("component", "metrics_handler").Debug("Metrics collection completed successfully")
			}
		}

		// Check again if request was cancelled during collection
		select {
		case <-r.Context().Done():
			s.logger.WithField("component", "metrics_handler").Debug("Request cancelled during collection")
			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		// Serve the metrics
		handler.ServeHTTP(w, r)
	})
}

// GetPrometheusRegistry returns the Prometheus registry
func (s *Server) GetPrometheusRegistry() *prometheus.Registry {
	return s.promRegistry
}

// SetPrometheusRegistry sets the Prometheus registry
func (s *Server) SetPrometheusRegistry(registry *prometheus.Registry) {
	s.promRegistry = registry
}

// RegisterCollector registers a Prometheus collector
func (s *Server) RegisterCollector(collector prometheus.Collector) error {
	return s.promRegistry.Register(collector)
}

// UnregisterCollector unregisters a Prometheus collector
func (s *Server) UnregisterCollector(collector prometheus.Collector) bool {
	return s.promRegistry.Unregister(collector)
}

// GetMetricsPath returns the configured metrics path
func (s *Server) GetMetricsPath() string {
	return s.config.Server.MetricsPath
}

// GetAddress returns the server address
func (s *Server) GetAddress() string {
	return s.config.Server.Address
}

// handleDebugHealth provides detailed health information for debugging
func (s *Server) handleDebugHealth(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "debug_health_handler")
	logger.Debug("Debug health check requested")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	report := s.healthChecker.CheckHealth(ctx)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"health_report": report,
		"server_info": map[string]interface{}{
			"address":      s.config.Server.Address,
			"metrics_path": s.config.Server.MetricsPath,
			"uptime":       time.Since(startTime),
		},
	}); err != nil {
		logger.WithError(err).Error("Failed to encode debug health response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDebugCollectors provides detailed collector information for debugging
func (s *Server) handleDebugCollectors(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "debug_collectors_handler")
	logger.Debug("Debug collectors requested")

	if s.registry == nil {
		http.Error(w, "Registry not available", http.StatusInternalServerError)
		return
	}

	stats := s.registry.GetStats()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	debugInfo := map[string]interface{}{
		"collectors": stats,
		"summary": map[string]interface{}{
			"total_collectors": len(stats),
			"enabled_collectors": func() int {
				enabled := 0
				for _, stat := range stats {
					if stat.Enabled {
						enabled++
					}
				}
				return enabled
			}(),
		},
		"timestamp": time.Now(),
	}

	if err := json.NewEncoder(w).Encode(debugInfo); err != nil {
		logger.WithError(err).Error("Failed to encode debug collectors response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleDebugPerformance provides detailed performance statistics for collectors
func (s *Server) handleDebugPerformance(w http.ResponseWriter, r *http.Request) {
	logger := s.logger.WithField("component", "debug_performance_handler")
	logger.Debug("Debug performance requested")

	if s.registry == nil {
		http.Error(w, "Registry not available", http.StatusInternalServerError)
		return
	}

	perfStats := s.registry.GetPerformanceStats()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")

	// Calculate aggregated statistics
	summary := map[string]interface{}{
		"total_collectors": len(perfStats),
		"total_collections": func() int64 {
			var total int64
			for _, stats := range perfStats {
				total += stats.CollectionCount
			}
			return total
		}(),
		"total_errors": func() int64 {
			var total int64
			for _, stats := range perfStats {
				total += stats.ErrorCount
			}
			return total
		}(),
		"total_metrics": func() int64 {
			var total int64
			for _, stats := range perfStats {
				total += stats.TotalMetrics
			}
			return total
		}(),
		"collectors_with_errors": func() int {
			count := 0
			for _, stats := range perfStats {
				if stats.ErrorCount > 0 {
					count++
				}
			}
			return count
		}(),
		"collectors_with_sla_violations": func() int {
			count := 0
			for _, stats := range perfStats {
				if stats.SLAViolations > 0 {
					count++
				}
			}
			return count
		}(),
	}

	// Format detailed statistics for each collector
	collectorDetails := make(map[string]interface{})
	for name, stats := range perfStats {
		avgDuration := time.Duration(0)
		if stats.CollectionCount > 0 {
			avgDuration = stats.TotalDuration / time.Duration(stats.CollectionCount)
		}

		successRate := float64(0)
		if stats.CollectionCount > 0 {
			successRate = float64(stats.SuccessCount) / float64(stats.CollectionCount) * 100
		}

		avgMetrics := int64(0)
		if stats.SuccessCount > 0 {
			avgMetrics = stats.TotalMetrics / stats.SuccessCount
		}

		collectorDetails[name] = map[string]interface{}{
			"collection_count":   stats.CollectionCount,
			"success_count":      stats.SuccessCount,
			"error_count":        stats.ErrorCount,
			"success_rate":       fmt.Sprintf("%.2f%%", successRate),
			"consecutive_errors": stats.ConsecutiveErrors,
			"last_error": func() interface{} {
				if stats.LastError != nil {
					return map[string]interface{}{
						"message": stats.LastError.Error(),
						"time":    stats.LastErrorTime,
					}
				}
				return nil
			}(),
			"timing": map[string]interface{}{
				"last_duration": stats.LastDuration.String(),
				"avg_duration":  avgDuration.String(),
				"min_duration":  stats.MinDuration.String(),
				"max_duration":  stats.MaxDuration.String(),
			},
			"metrics": map[string]interface{}{
				"total_metrics": stats.TotalMetrics,
				"last_count":    stats.LastMetricCount,
				"avg_count":     avgMetrics,
				"max_count":     stats.MaxMetricCount,
			},
			"resources": map[string]interface{}{
				"last_memory_bytes": stats.LastMemoryUsage,
				"max_memory_bytes":  stats.MaxMemoryUsage,
				"goroutines":        stats.GoroutineCount,
			},
			"sla": map[string]interface{}{
				"violations":          stats.SLAViolations,
				"last_violation_time": stats.LastSLAViolation,
			},
			"last_success_time": stats.LastSuccessTime,
		}
	}

	debugInfo := map[string]interface{}{
		"summary":    summary,
		"collectors": collectorDetails,
		"timestamp":  time.Now(),
	}

	if err := json.NewEncoder(w).Encode(debugInfo); err != nil {
		logger.WithError(err).Error("Failed to encode debug performance response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
