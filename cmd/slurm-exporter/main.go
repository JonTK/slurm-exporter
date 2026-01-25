// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/logging"
	"github.com/jontk/slurm-exporter/internal/server"
	"github.com/jontk/slurm-exporter/internal/slurm"
	"github.com/jontk/slurm-exporter/pkg/version"
)

// Timeout constants
const (
	// performanceMonitoringInterval is the interval for performance metric collection
	performanceMonitoringInterval = 5 * time.Minute
	// gracefulShutdownTimeout is the maximum duration to wait for graceful shutdown
	gracefulShutdownTimeout = 30 * time.Second
	// healthCheckTimeout is the timeout for health check operations
	healthCheckTimeout = 5 * time.Second
)

var (
	configFile  = flag.String("config", "configs/config.yaml", "Path to configuration file")
	logLevel    = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	showVersion = flag.Bool("version", false, "Show version information and exit")
	healthCheck = flag.Bool("health-check", false, "Perform health check and exit")
	addr        = flag.String("addr", ":8080", "Address to listen on")
	metricsPath = flag.String("metrics-path", "/metrics", "Path for metrics endpoint")
)

func main() {
	flag.Parse()

	// Show version information if requested
	if *showVersion {
		versionInfo := version.Get()
		fmt.Println(versionInfo.String())
		os.Exit(0)
	}

	// Perform health check if requested
	if *healthCheck {
		if err := performHealthCheck(); err != nil {
			fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Health check passed")
		os.Exit(0)
	}

	// Load configuration first
	cfg, err := config.Load(*configFile)
	if err != nil {
		// Use basic logrus for configuration errors
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Override log level from command line if provided
	if *logLevel != "info" {
		cfg.Logging.Level = *logLevel
	}

	// Set up structured logging
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize logger")
	}

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Override config with command line flags if provided
	if *addr != ":8080" {
		cfg.Server.Address = *addr
	}
	if *metricsPath != "/metrics" {
		cfg.Server.MetricsPath = *metricsPath
	}

	logger.WithComponent("main").WithFields(logrus.Fields{
		"version":      version.Get().Short(),
		"config_file":  *configFile,
		"address":      cfg.Server.Address,
		"metrics_path": cfg.Server.MetricsPath,
		"log_level":    cfg.Logging.Level,
		"log_format":   cfg.Logging.Format,
	}).Info("Starting SLURM Prometheus Exporter")

	// Create Prometheus registry for collectors
	promRegistry := prometheus.NewRegistry()

	// Create collector registry
	registry, err := collector.NewRegistry(&cfg.Collectors, promRegistry)
	if err != nil {
		logger.WithComponent("main").WithError(err).Fatal("Failed to create collector registry")
	}

	// Create SLURM client using our wrapper
	slurmWrapper, err := slurm.NewClient(&cfg.SLURM)
	if err != nil {
		// Don't log the error directly as it may contain sensitive config information
		logger.WithComponent("main").Fatal("Failed to create SLURM client (check configuration for details)")
	}

	// Get the underlying client for collectors
	slurmClient := slurmWrapper.GetSlurmClient()

	// Create and register collectors based on configuration
	if err := registry.CreateCollectorsFromConfig(&cfg.Collectors, slurmClient); err != nil {
		logger.WithComponent("main").WithError(err).Fatal("Failed to create collectors")
	}

	// Start performance monitoring with configured reporting interval
	registry.StartPerformanceMonitoring(ctx, performanceMonitoringInterval)
	logger.WithComponent("main").Info("Performance monitoring started")

	// Create and start the server
	srv, err := server.New(cfg, logger.Logger, registry, promRegistry)
	if err != nil {
		logger.WithComponent("main").WithError(err).Fatal("Failed to create server")
	}

	// Setup graceful shutdown handling
	shutdown := NewShutdownManager(logger.Logger, gracefulShutdownTimeout)

	// Register shutdown hooks for proper cleanup
	shutdown.AddShutdownHook("server", func(ctx context.Context) error {
		logger.WithComponent("shutdown").Info("Shutting down HTTP server")
		return srv.Shutdown(ctx)
	})

	shutdown.AddShutdownHook("logger", func(ctx context.Context) error {
		logger.WithComponent("shutdown").Info("Closing logger")
		return logger.Close()
	})

	// Setup config watcher for hot-reload
	configWatcher, err := config.NewWatcher(*configFile, config.CreateReloadHandler(registry, logger.WithComponent("config-watcher")), logger.WithComponent("config-watcher"))
	if err != nil {
		logger.WithComponent("main").WithError(err).Error("Failed to create config watcher, hot-reload disabled")
		// Continue without hot-reload
	} else {
		if err := configWatcher.Start(ctx); err != nil {
			logger.WithComponent("main").WithError(err).Error("Failed to start config watcher")
		} else {
			logger.WithComponent("main").Info("Configuration hot-reload enabled")
			shutdown.AddShutdownHook("config-watcher", func(ctx context.Context) error {
				logger.WithComponent("shutdown").Info("Stopping config watcher")
				return configWatcher.Stop()
			})
		}
	}

	// Start the shutdown manager
	shutdown.Start(ctx)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := srv.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	var exitCode int
	select {
	case sig := <-shutdown.SignalChan():
		logger.WithComponent("main").WithField("signal", sig).Info("Received shutdown signal")
		exitCode = 0
	case err := <-errChan:
		logger.WithComponent("main").WithError(err).Error("Server error")
		exitCode = 1
	}

	// Trigger graceful shutdown
	logger.WithComponent("main").Info("Initiating graceful shutdown...")
	if err := shutdown.Shutdown(); err != nil {
		logger.WithComponent("main").WithError(err).Error("Shutdown completed with errors")
		exitCode = 1
	} else {
		logger.WithComponent("main").Info("Graceful shutdown completed successfully")
	}

	// Explicit cleanup before exit
	cancel()
	if err := logger.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
	}
	os.Exit(exitCode)
}

// performHealthCheck performs a simple health check by trying to connect to the health endpoint
func performHealthCheck() error {
	// Default address for health check
	healthURL := "http://localhost:8080/health"

	// Try to read address from environment or config
	if envAddr := os.Getenv("SLURM_EXPORTER_ADDRESS"); envAddr != "" {
		healthURL = fmt.Sprintf("http://%s/health", envAddr)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: healthCheckTimeout,
	}

	// Create request with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Perform health check request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to health endpoint: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", resp.StatusCode)
	}

	return nil
}
