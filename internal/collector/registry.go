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

	registry := &Registry{
		collectors:         make(map[string]Collector),
		promRegistry:       promRegistry,
		metrics:            collectorMetrics,
		config:             cfg,
		cardinalityManager: cardinalityManager,
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
	if err := r.promRegistry.Register(&collectorAdapter{collector: collector}); err != nil {
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
	collector Collector
}

// Describe implements prometheus.Collector
func (ca *collectorAdapter) Describe(ch chan<- *prometheus.Desc) {
	ca.collector.Describe(ch)
}

// Collect implements prometheus.Collector
func (ca *collectorAdapter) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	if err := ca.collector.Collect(ctx, ch); err != nil {
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

// RegistryOptions provides options for creating a registry
type RegistryOptions struct {
	// Maximum concurrent collections
	MaxConcurrency int

	// Collection timeout
	CollectionTimeout time.Duration

	// Whether to continue on collector failures
	ContinueOnError bool
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

	// Add other collectors as they are implemented
	// if cfg.Performance.Enabled { ... }
	// if cfg.System.Enabled { ... }

	r.logger.WithField("count", len(r.collectors)).Info("Collectors created and registered")
	return nil
}
