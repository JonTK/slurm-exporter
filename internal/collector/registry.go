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

	slurm "github.com/jontk/slurm-client"
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

// registerCollector registers a collector and logs success
func (r *Registry) registerCollector(name string, collector Collector) error {
	if err := r.Register(name, collector); err != nil {
		return fmt.Errorf("failed to register %s collector: %w", name, err)
	}
	r.logger.Infof("%s collector registered successfully", name)
	return nil
}

// configureCollectorFeatures configures optional collector features (filtering, cardinality, custom labels)
func (r *Registry) configureCollectorFeatures(collector Collector, filterConfig config.FilterConfig, customLabels map[string]string) {
	// Configure filtering
	if filterableCollector, ok := collector.(FilterableCollector); ok {
		hasFilters := len(filterConfig.Metrics.IncludeMetrics) > 0 || len(filterConfig.Metrics.ExcludeMetrics) > 0
		if hasFilters {
			filterableCollector.UpdateFilterConfig(filterConfig)
		}
	}

	// Configure cardinality management
	if cardinalityAware, ok := collector.(CardinalityAwareCollector); ok {
		cardinalityAware.SetCardinalityManager(r.cardinalityManager)
	}

	// Configure custom labels
	if customLabelsCollector, ok := collector.(CustomLabelsCollector); ok {
		if len(customLabels) > 0 {
			customLabelsCollector.SetCustomLabels(customLabels)
		}
	}
}

// registerSimpleCollectors registers collectors with no extra configuration
func (r *Registry) registerSimpleCollectors(cfg *config.CollectorsConfig, client slurm.SlurmClient, logger *logrus.Entry) error {
	collectors := []struct {
		enabled bool
		name    string
		factory func() Collector
	}{
		{cfg.QoS.Enabled, "qos", func() Collector { return NewQoSCollector(client, logger) }},
		{cfg.Reservations.Enabled, "reservations", func() Collector { return NewReservationCollector(client, logger) }},
		{cfg.Partitions.Enabled, "partitions", func() Collector { return NewPartitionsSimpleCollector(client, logger) }},
		{cfg.Cluster.Enabled, "cluster", func() Collector { return NewClusterSimpleCollector(client, logger) }},
		{cfg.Users.Enabled, "users", func() Collector { return NewUsersSimpleCollector(client, logger) }},
		// {cfg.Accounts.Enabled, "accounts", func() Collector { return NewAccountsSimpleCollector(client, logger) }},
		{cfg.Associations.Enabled, "associations", func() Collector { return NewAssociationsSimpleCollector(client, logger) }},
	}

	for _, c := range collectors {
		if c.enabled {
			if err := r.registerCollector(c.name, c.factory()); err != nil {
				return err
			}
		}
	}
	return nil
}

// registerTimeoutCollectors registers collectors that require timeout parameter
func (r *Registry) registerTimeoutCollectors(cfg *config.CollectorsConfig, client slurm.SlurmClient, logger *logrus.Entry, timeout time.Duration) error {
	collectors := []struct {
		enabled bool
		name    string
		factory func() Collector
	}{
		{cfg.Licenses.Enabled, "licenses", func() Collector { return NewLicensesCollector(client, logger, timeout) }},
		{cfg.Shares.Enabled, "shares", func() Collector { return NewSharesCollector(client, logger, timeout) }},
		{cfg.Diagnostics.Enabled, "diagnostics", func() Collector { return NewDiagnosticsCollector(client, logger, timeout) }},
		{cfg.TRES.Enabled, "tres", func() Collector { return NewTRESCollector(client, logger, timeout) }},
		{cfg.WCKeys.Enabled, "wckeys", func() Collector { return NewWCKeysCollector(client, logger, timeout) }},
		{cfg.Clusters.Enabled, "clusters", func() Collector { return NewClustersCollector(client, logger, timeout) }},
	}

	for _, c := range collectors {
		if c.enabled {
			if err := r.registerCollector(c.name, c.factory()); err != nil {
				return err
			}
		}
	}
	return nil
}

// registerConfigurableCollectors registers collectors with custom labels and/or filtering
func (r *Registry) registerConfigurableCollectors(cfg *config.CollectorsConfig, client slurm.SlurmClient) error {
	configurableCollectors := []struct {
		enabled      bool
		name         string
		factory      func() Collector
		filterConfig config.FilterConfig
		labels       map[string]string
	}{
		{cfg.Jobs.Enabled, "jobs", func() Collector { return NewJobsSimpleCollector(client, r.logger) }, cfg.Jobs.Filters, cfg.Jobs.Labels},
		{cfg.Nodes.Enabled, "nodes", func() Collector { return NewNodesSimpleCollector(client, r.logger) }, config.FilterConfig{}, cfg.Nodes.Labels},
		// Temporarily disabled during API migration:
		// {cfg.Performance.Enabled, "performance", func() Collector { return NewPerformanceSimpleCollector(client, r.logger) }, config.FilterConfig{}, cfg.Performance.Labels},
		{cfg.System.Enabled, "system", func() Collector { return NewSystemSimpleCollector(client, r.logger) }, config.FilterConfig{}, cfg.System.Labels},
	}

	for _, c := range configurableCollectors {
		if c.enabled {
			collector := c.factory()
			r.configureCollectorFeatures(collector, c.filterConfig, c.labels)
			if err := r.registerCollector(c.name, collector); err != nil {
				return err
			}
		}
	}
	return nil
}

// CreateCollectorsFromConfig creates and registers collectors based on configuration
func (r *Registry) CreateCollectorsFromConfig(cfg *config.CollectorsConfig, client interface{}) error {
	r.logger.Info("Creating collectors from configuration")

	// Cast client to the expected interface
	slurmClient, ok := client.(slurm.SlurmClient)
	if !ok {
		return fmt.Errorf("invalid client type, expected slurm.SlurmClient")
	}

	// Register simple collectors (no extra configuration)
	if err := r.registerSimpleCollectors(cfg, slurmClient, r.logger); err != nil {
		return err
	}

	// Register collectors with custom labels and/or filtering
	if err := r.registerConfigurableCollectors(cfg, slurmClient); err != nil {
		return err
	}

	// Register collectors that require timeout parameter
	if err := r.registerTimeoutCollectors(cfg, slurmClient, r.logger, cfg.CollectionTimeout); err != nil {
		return err
	}

	r.logger.WithField("count", len(r.collectors)).Info("Collectors created and registered")
	return nil
}
