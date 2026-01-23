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

	"github.com/jontk/slurm-client"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/metrics"
)

// Registry manages multiple collectors
type Registry struct {
	// Collectors map
	collectors map[string]Collector
	mu         sync.RWMutex

	// Prometheus registry
	promRegistry *prometheus.Registry

	// Metrics for the registry itself
	metrics *CollectorMetrics

	// Configuration
	config *config.CollectorsConfig

	// Cardinality management
	cardinalityManager *metrics.CardinalityManager

	// Performance monitoring
	performanceMonitor *PerformanceMonitor

	// Logger
	logger *logrus.Entry
}

// NewRegistry creates a new collector registry
func NewRegistry(cfg *config.CollectorsConfig, promRegistry *prometheus.Registry) (*Registry, error) {
	logger := logrus.WithField("component", "collector_registry")

	// Create metrics for registry operations
	collectorMetrics := NewCollectorMetrics("slurm", "exporter")
	if err := collectorMetrics.Register(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register registry metrics: %w", err)
	}

	// Create cardinality manager
	cardinalityManager := metrics.NewCardinalityManager(logger.Logger)
	if err := cardinalityManager.RegisterMetrics(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register cardinality metrics: %w", err)
	}

	// Create performance monitor with default SLAs
	slaConfig := SLAConfig{
		MaxCollectionDuration:   30 * time.Second,
		MaxErrorRate:            0.1,               // 0.1 errors per minute
		MinSuccessRate:          95.0,              // 95% success rate
		MaxMemoryUsage:          100 * 1024 * 1024, // 100MB per collector
		MaxMetricsPerCollection: 10000,
	}
	performanceMonitor := NewPerformanceMonitor("slurm", "exporter", slaConfig, logger)
	if err := performanceMonitor.RegisterMetrics(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register performance metrics: %w", err)
	}

	registry := &Registry{
		collectors:         make(map[string]Collector),
		promRegistry:       promRegistry,
		metrics:            collectorMetrics,
		config:             cfg,
		cardinalityManager: cardinalityManager,
		performanceMonitor: performanceMonitor,
		logger:             logger,
	}

	// Register the registry itself as a collector
	if err := promRegistry.Register(registry); err != nil {
		return nil, fmt.Errorf("failed to register collector registry: %w", err)
	}

	return registry, nil
}

// Register adds a collector to the registry
func (r *Registry) Register(name string, collector Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.collectors[name]; exists {
		return fmt.Errorf("collector %s already registered", name)
	}

	// Register collector with Prometheus
	if err := r.promRegistry.Register(&collectorAdapter{
		collector:          collector,
		performanceMonitor: r.performanceMonitor,
	}); err != nil {
		return fmt.Errorf("failed to register collector %s with prometheus: %w", name, err)
	}

	r.collectors[name] = collector
	r.logger.WithField("collector", name).Info("Collector registered")

	// Initialize collector state
	labels := prometheus.Labels{"collector": name}
	if collector.IsEnabled() {
		r.metrics.Up.With(labels).Set(1)
	} else {
		r.metrics.Up.With(labels).Set(0)
	}

	return nil
}

// Unregister removes a collector from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	collector, exists := r.collectors[name]
	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}

	// Unregister from Prometheus
	if !r.promRegistry.Unregister(&collectorAdapter{collector: collector}) {
		r.logger.WithField("collector", name).Warn("Failed to unregister collector from prometheus")
	}

	delete(r.collectors, name)
	r.logger.WithField("collector", name).Info("Collector unregistered")

	// Remove metrics
	labels := prometheus.Labels{"collector": name}
	r.metrics.Up.Delete(labels)
	r.metrics.Total.Delete(labels)
	r.metrics.Errors.Delete(labels)
	r.metrics.Duration.Delete(labels)

	return nil
}

// Get returns a collector by name
func (r *Registry) Get(name string) (Collector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	collector, exists := r.collectors[name]
	return collector, exists
}

// List returns all registered collectors
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.collectors))
	for name := range r.collectors {
		names = append(names, name)
	}
	return names
}

// EnableCollector enables a specific collector
func (r *Registry) EnableCollector(name string) error {
	r.mu.RLock()
	collector, exists := r.collectors[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}

	collector.SetEnabled(true)
	labels := prometheus.Labels{"collector": name}
	r.metrics.Up.With(labels).Set(1)

	r.logger.WithField("collector", name).Info("Collector enabled")
	return nil
}

// DisableCollector disables a specific collector
func (r *Registry) DisableCollector(name string) error {
	r.mu.RLock()
	collector, exists := r.collectors[name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collector %s not found", name)
	}

	collector.SetEnabled(false)
	labels := prometheus.Labels{"collector": name}
	r.metrics.Up.With(labels).Set(0)

	r.logger.WithField("collector", name).Info("Collector disabled")
	return nil
}

// CollectAll triggers collection for all enabled collectors
func (r *Registry) CollectAll(ctx context.Context) error {
	r.mu.RLock()
	collectors := make(map[string]Collector, len(r.collectors))
	for name, collector := range r.collectors {
		collectors[name] = collector
	}
	r.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(collectors))

	// Collect from all collectors concurrently
	for name, collector := range collectors {
		if !collector.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(name string, collector Collector) {
			defer wg.Done()

			// Create a dummy channel for collection
			ch := make(chan prometheus.Metric, 1000)
			defer close(ch)

			// Drain the channel in background
			go func() {
				for range ch {
					// Discard metrics - this is just for triggering collection
				}
			}()

			if err := collector.Collect(ctx, ch); err != nil {
				errChan <- fmt.Errorf("collector %s: %w", name, err)
			}
		}(name, collector)
	}

	// Wait for all collections to complete
	wg.Wait()
	close(errChan)

	// Collect all errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("collection errors: %v", errs)
	}

	return nil
}

// GetStats returns statistics for all collectors
func (r *Registry) GetStats() map[string]CollectorState {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]CollectorState)
	for name, collector := range r.collectors {
		if bc, ok := collector.(*BaseCollector); ok {
			stats[name] = bc.GetState()
		} else {
			// For non-base collectors, provide basic info
			stats[name] = CollectorState{
				Name:    name,
				Enabled: collector.IsEnabled(),
			}
		}
	}

	return stats
}

// Describe implements prometheus.Collector for the registry itself
func (r *Registry) Describe(ch chan<- *prometheus.Desc) {
	// The registry doesn't have its own metrics to describe
	// Individual collectors will describe their own metrics
}

// Collect implements prometheus.Collector for the registry itself
func (r *Registry) Collect(ch chan<- prometheus.Metric) {
	// The registry doesn't collect its own metrics
	// Individual collectors will handle their own collection
}

// collectorAdapter adapts our Collector interface to prometheus.Collector
type collectorAdapter struct {
	collector          Collector
	performanceMonitor *PerformanceMonitor
}

// Describe implements prometheus.Collector
func (ca *collectorAdapter) Describe(ch chan<- *prometheus.Desc) {
	ca.collector.Describe(ch)
}

// Collect implements prometheus.Collector
func (ca *collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	startTime := time.Now()

	// Create a wrapper channel to count metrics
	metricsChan := make(chan prometheus.Metric, 1000)
	var metricsCount int
	done := make(chan struct{})

	// Start a goroutine to forward metrics and count them
	go func() {
		defer close(done)
		for metric := range metricsChan {
			select {
			case ch <- metric:
				metricsCount++
			case <-ctx.Done():
				// Context cancelled, stop sending metrics
				return
			}
		}
	}()

	// Collect metrics
	err := ca.collector.Collect(ctx, metricsChan)
	close(metricsChan)

	// Wait for the forwarding goroutine to finish
	<-done

	duration := time.Since(startTime)

	// Record performance metrics
	if ca.performanceMonitor != nil {
		ca.performanceMonitor.RecordCollection(ca.collector.Name(), duration, metricsCount, err)
	}

	if err != nil {
		logrus.WithError(err).WithField("collector", ca.collector.Name()).Error("Collection failed")
	}
}

// GetCardinalityManager returns the cardinality manager
func (r *Registry) GetCardinalityManager() *metrics.CardinalityManager {
	return r.cardinalityManager
}

// SetCardinalityLimit configures a cardinality limit for a metric pattern
func (r *Registry) SetCardinalityLimit(pattern string, limit metrics.CardinalityLimit) {
	if r.cardinalityManager != nil {
		r.cardinalityManager.SetLimit(pattern, limit)
	}
}

// SetCardinalityFilter configures a cardinality filter for a metric pattern
func (r *Registry) SetCardinalityFilter(pattern string, filter metrics.CardinalityFilter) {
	if r.cardinalityManager != nil {
		r.cardinalityManager.SetFilter(pattern, filter)
	}
}

// GetCardinalityReport generates a comprehensive cardinality report
func (r *Registry) GetCardinalityReport() metrics.CardinalityReport {
	if r.cardinalityManager != nil {
		return r.cardinalityManager.GetCardinalityReport()
	}
	return metrics.CardinalityReport{}
}

// StartCardinalityMonitoring begins background cardinality monitoring
func (r *Registry) StartCardinalityMonitoring(ctx context.Context) {
	if r.cardinalityManager != nil {
		go r.cardinalityManager.StartCardinalityMonitoring(ctx)
		r.logger.Info("Cardinality monitoring started")
	}
}

// StartPerformanceMonitoring begins background performance monitoring
func (r *Registry) StartPerformanceMonitoring(ctx context.Context, reportInterval time.Duration) {
	if r.performanceMonitor != nil {
		go r.performanceMonitor.StartPeriodicReporting(ctx, reportInterval)
		r.logger.Info("Performance monitoring started")
	}
}

// GetPerformanceStats returns performance statistics for all collectors
func (r *Registry) GetPerformanceStats() map[string]*CollectorPerformanceStats {
	if r.performanceMonitor != nil {
		return r.performanceMonitor.GetStats()
	}
	return make(map[string]*CollectorPerformanceStats)
}

// GetCollectorPerformanceStats returns performance statistics for a specific collector
func (r *Registry) GetCollectorPerformanceStats(collector string) (*CollectorPerformanceStats, bool) {
	if r.performanceMonitor != nil {
		return r.performanceMonitor.GetCollectorStats(collector)
	}
	return nil, false
}

// RegistryOptions provides options for creating a registry
type RegistryOptions struct {
	// Maximum concurrent collections
	MaxConcurrency int

	// Collection timeout
	CollectionTimeout time.Duration

	// Whether to continue on collector failures
	ContinueOnError bool
}

// ReconfigureCollectors updates collector configurations without restart
func (r *Registry) ReconfigureCollectors(cfg *config.CollectorsConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("Reconfiguring collectors")

	// Update registry config
	r.config = cfg

	// Update each collector's configuration
	var errors []error
	for name, collector := range r.collectors {
		// Update enabled state based on configuration
		var enabled bool
		var filterConfig config.FilterConfig
		var customLabels map[string]string

		switch name {
		case "jobs":
			enabled = cfg.Jobs.Enabled
			filterConfig = cfg.Jobs.Filters
			customLabels = cfg.Jobs.Labels
		case "nodes":
			enabled = cfg.Nodes.Enabled
			filterConfig = cfg.Nodes.Filters
			customLabels = cfg.Nodes.Labels
		case "partitions":
			enabled = cfg.Partitions.Enabled
			filterConfig = cfg.Partitions.Filters
			customLabels = cfg.Partitions.Labels
		case "accounts":
			enabled = cfg.Accounts.Enabled
			filterConfig = cfg.Accounts.Filters
			customLabels = cfg.Accounts.Labels
		case "users":
			enabled = cfg.Users.Enabled
			filterConfig = cfg.Users.Filters
			customLabels = cfg.Users.Labels
		case "qos":
			enabled = cfg.QoS.Enabled
			filterConfig = cfg.QoS.Filters
			customLabels = cfg.QoS.Labels
		case "reservations":
			enabled = cfg.Reservations.Enabled
			filterConfig = cfg.Reservations.Filters
			customLabels = cfg.Reservations.Labels
		case "associations":
			enabled = cfg.Associations.Enabled
			filterConfig = cfg.Associations.Filters
			customLabels = cfg.Associations.Labels
		case "cluster":
			enabled = cfg.Cluster.Enabled
			filterConfig = cfg.Cluster.Filters
			customLabels = cfg.Cluster.Labels
		case "performance":
			enabled = cfg.Performance.Enabled
			filterConfig = cfg.Performance.Filters
			customLabels = cfg.Performance.Labels
		case "system":
			enabled = cfg.System.Enabled
			filterConfig = cfg.System.Filters
			customLabels = cfg.System.Labels
		default:
			r.logger.WithField("collector", name).Warn("Unknown collector in registry")
			continue
		}

		// Update enabled state
		collector.SetEnabled(enabled)
		labels := prometheus.Labels{"collector": name}
		if enabled {
			r.metrics.Up.With(labels).Set(1)
		} else {
			r.metrics.Up.With(labels).Set(0)
		}

		// Update filter configuration if supported
		if filterableCollector, ok := collector.(FilterableCollector); ok {
			filterableCollector.UpdateFilterConfig(filterConfig)
			r.logger.WithField("collector", name).Debug("Updated filter configuration")
		}

		// Update custom labels if supported
		if customLabelsCollector, ok := collector.(CustomLabelsCollector); ok {
			customLabelsCollector.SetCustomLabels(customLabels)
			r.logger.WithField("collector", name).Debug("Updated custom labels")
		}
	}

	// Update cardinality limits
	// TODO: Implement cardinality limit updates when configuration structure is defined

	if len(errors) > 0 {
		return fmt.Errorf("reconfiguration errors: %v", errors)
	}

	r.logger.Info("Collectors reconfigured successfully")
	return nil
}

// CollectorFactory is a function that creates a collector
type CollectorFactory func(config *config.CollectorConfig) (Collector, error)

// RegisteredCollectorFactories holds all available collector factories
var RegisteredCollectorFactories = make(map[string]CollectorFactory)

// RegisterCollectorFactory registers a collector factory
func RegisterCollectorFactory(name string, factory CollectorFactory) {
	RegisteredCollectorFactories[name] = factory
}

// CreateCollectorsFromConfig creates and registers collectors based on configuration
func (r *Registry) CreateCollectorsFromConfig(cfg *config.CollectorsConfig, client interface{}) error {
	r.logger.Info("Creating collectors from configuration")

	// Cast client to the expected interface
	_, ok := client.(slurm.SlurmClient)
	if !ok {
		return fmt.Errorf("invalid client type, expected slurm.SlurmClient")
	}

	// Create and register enabled collectors
	slurmClient := client.(slurm.SlurmClient)
	logger := r.logger

	// Register collectors that have working constructors
	if cfg.QoS.Enabled {
		collector := NewQoSCollector(slurmClient, logger)
		if err := r.Register("qos", collector); err != nil {
			return fmt.Errorf("failed to register QoS collector: %w", err)
		}
		r.logger.Info("QoS collector registered successfully")
	}

	if cfg.Reservations.Enabled {
		collector := NewReservationCollector(slurmClient, logger)
		if err := r.Register("reservations", collector); err != nil {
			return fmt.Errorf("failed to register reservation collector: %w", err)
		}
		r.logger.Info("Reservations collector registered successfully")
	}

	// Jobs collector
	if cfg.Jobs.Enabled {
		collector := NewJobsSimpleCollector(slurmClient, logger)

		// Configure filtering
		if filterableCollector, ok := (Collector(collector)).(FilterableCollector); ok {
			filterableCollector.UpdateFilterConfig(cfg.Jobs.Filters)
		}

		// Configure cardinality management
		if cardinalityAware, ok := (Collector(collector)).(CardinalityAwareCollector); ok {
			cardinalityAware.SetCardinalityManager(r.cardinalityManager)
		}

		// Configure custom labels
		if customLabelsCollector, ok := (Collector(collector)).(CustomLabelsCollector); ok {
			customLabelsCollector.SetCustomLabels(cfg.Jobs.Labels)
		}

		if err := r.Register("jobs", collector); err != nil {
			return fmt.Errorf("failed to register jobs collector: %w", err)
		}
		r.logger.Info("Jobs collector registered successfully")
	}

	// Nodes collector
	if cfg.Nodes.Enabled {
		collector := NewNodesSimpleCollector(slurmClient, logger)

		// Configure custom labels
		if customLabelsCollector, ok := (Collector(collector)).(CustomLabelsCollector); ok {
			customLabelsCollector.SetCustomLabels(cfg.Nodes.Labels)
		}

		if err := r.Register("nodes", collector); err != nil {
			return fmt.Errorf("failed to register nodes collector: %w", err)
		}
		r.logger.Info("Nodes collector registered successfully")
	}

	// Partitions collector
	if cfg.Partitions.Enabled {
		collector := NewPartitionsSimpleCollector(slurmClient, logger)
		if err := r.Register("partitions", collector); err != nil {
			return fmt.Errorf("failed to register partitions collector: %w", err)
		}
		r.logger.Info("Partitions collector registered successfully")
	}

	// Cluster collector
	if cfg.Cluster.Enabled {
		collector := NewClusterSimpleCollector(slurmClient, logger)
		if err := r.Register("cluster", collector); err != nil {
			return fmt.Errorf("failed to register cluster collector: %w", err)
		}
		r.logger.Info("Cluster collector registered successfully")
	}

	// Users collector
	if cfg.Users.Enabled {
		collector := NewUsersSimpleCollector(slurmClient, logger)
		if err := r.Register("users", collector); err != nil {
			return fmt.Errorf("failed to register users collector: %w", err)
		}
		r.logger.Info("Users collector registered successfully")
	}

	// Accounts collector
	if cfg.Accounts.Enabled {
		collector := NewAccountsSimpleCollector(slurmClient, logger)
		if err := r.Register("accounts", collector); err != nil {
			return fmt.Errorf("failed to register accounts collector: %w", err)
		}
		r.logger.Info("Accounts collector registered successfully")
	}

	// Associations collector
	if cfg.Associations.Enabled {
		collector := NewAssociationsSimpleCollector(slurmClient, logger)
		if err := r.Register("associations", collector); err != nil {
			return fmt.Errorf("failed to register associations collector: %w", err)
		}
		r.logger.Info("Associations collector registered successfully")
	}

	// Performance collector
	if cfg.Performance.Enabled {
		collector := NewPerformanceSimpleCollector(slurmClient, logger)

		// Configure custom labels
		if customLabelsCollector, ok := (Collector(collector)).(CustomLabelsCollector); ok {
			customLabelsCollector.SetCustomLabels(cfg.Performance.Labels)
		}

		if err := r.Register("performance", collector); err != nil {
			return fmt.Errorf("failed to register performance collector: %w", err)
		}
		r.logger.Info("Performance collector registered successfully")
	}

	// System collector
	if cfg.System.Enabled {
		collector := NewSystemSimpleCollector(slurmClient, logger)

		// Configure custom labels
		if customLabelsCollector, ok := (Collector(collector)).(CustomLabelsCollector); ok {
			customLabelsCollector.SetCustomLabels(cfg.System.Labels)
		}

		if err := r.Register("system", collector); err != nil {
			return fmt.Errorf("failed to register system collector: %w", err)
		}
		r.logger.Info("System collector registered successfully")
	}

	// Register new standalone operation collectors if enabled
	// Note: These collectors may need to be added to the config structure

	// Licenses collector
	if cfg.Licenses.Enabled {
		collector := NewLicensesCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("licenses", collector); err != nil {
			return fmt.Errorf("failed to register licenses collector: %w", err)
		}
		r.logger.Info("Licenses collector registered successfully")
	}

	// Shares collector (fairshare)
	if cfg.Shares.Enabled {
		collector := NewSharesCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("shares", collector); err != nil {
			return fmt.Errorf("failed to register shares collector: %w", err)
		}
		r.logger.Info("Shares collector registered successfully")
	}

	// Diagnostics collector
	if cfg.Diagnostics.Enabled {
		collector := NewDiagnosticsCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("diagnostics", collector); err != nil {
			return fmt.Errorf("failed to register diagnostics collector: %w", err)
		}
		r.logger.Info("Diagnostics collector registered successfully")
	}

	// TRES collector
	if cfg.TRES.Enabled {
		collector := NewTRESCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("tres", collector); err != nil {
			return fmt.Errorf("failed to register TRES collector: %w", err)
		}
		r.logger.Info("TRES collector registered successfully")
	}

	// WCKeys collector
	if cfg.WCKeys.Enabled {
		collector := NewWCKeysCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("wckeys", collector); err != nil {
			return fmt.Errorf("failed to register WCKeys collector: %w", err)
		}
		r.logger.Info("WCKeys collector registered successfully")
	}

	// Clusters collector
	if cfg.Clusters.Enabled {
		collector := NewClustersCollector(slurmClient, logger, cfg.CollectionTimeout)
		if err := r.Register("clusters", collector); err != nil {
			return fmt.Errorf("failed to register clusters collector: %w", err)
		}
		r.logger.Info("Clusters collector registered successfully")
	}

	r.logger.WithField("count", len(r.collectors)).Info("Collectors created and registered")
	return nil
}
