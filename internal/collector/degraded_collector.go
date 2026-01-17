package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
)

// DegradedCollector wraps a collector with degradation support
type DegradedCollector struct {
	collector          Collector
	degradationManager *DegradationManager
	logger             *logrus.Entry
	// TODO: Unused field - preserved for future thread safety
	// mu                 sync.RWMutex
}

// NewDegradedCollector creates a new collector with degradation support
func NewDegradedCollector(collector Collector, degradationManager *DegradationManager) *DegradedCollector {
	return &DegradedCollector{
		collector:          collector,
		degradationManager: degradationManager,
		logger:             logrus.WithField("component", "degraded_collector").WithField("collector", collector.Name()),
	}
}

// Name returns the collector name
func (dc *DegradedCollector) Name() string {
	return dc.collector.Name()
}

// Describe forwards to the wrapped collector
func (dc *DegradedCollector) Describe(ch chan<- *prometheus.Desc) {
	dc.collector.Describe(ch)
}

// Collect performs collection with degradation support
func (dc *DegradedCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	// Create a function that collects metrics and returns them
	collectFunc := func(ctx context.Context) ([]prometheus.Metric, error) {
		// Create a buffered channel to collect metrics
		metricChan := make(chan prometheus.Metric, 1000)
		var metrics []prometheus.Metric
		var collectionErr error

		// Collect in a goroutine
		done := make(chan struct{})
		go func() {
			defer close(done)
			collectionErr = dc.collector.Collect(ctx, metricChan)
			close(metricChan)
		}()

		// Gather all metrics
		for metric := range metricChan {
			metrics = append(metrics, metric)
		}

		// Wait for collection to complete
		<-done

		return metrics, collectionErr
	}

	// Execute with degradation support
	metrics, err := dc.degradationManager.ExecuteWithDegradation(ctx, dc.collector.Name(), collectFunc)

	// Send collected metrics to the channel
	for _, metric := range metrics {
		select {
		case ch <- metric:
			// Metric sent successfully
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

// IsEnabled returns whether this collector is enabled
func (dc *DegradedCollector) IsEnabled() bool {
	return dc.collector.IsEnabled()
}

// SetEnabled enables or disables the collector
func (dc *DegradedCollector) SetEnabled(enabled bool) {
	dc.collector.SetEnabled(enabled)
}

// DegradedRegistry wraps a collector registry with degradation support
type DegradedRegistry struct {
	registry           *Registry
	degradationManager *DegradationManager
	degradedCollectors map[string]*DegradedCollector
	mu                 sync.RWMutex
	logger             *logrus.Entry
}

// NewDegradedRegistry creates a new registry with degradation support
func NewDegradedRegistry(
	config *config.CollectorsConfig,
	promRegistry *prometheus.Registry,
) (*DegradedRegistry, error) {
	// Create base registry
	registry, err := NewRegistry(config, promRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create base registry: %w", err)
	}

	// Create degradation manager
	degradationManager, err := NewDegradationManager(&config.Degradation, promRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create degradation manager: %w", err)
	}

	dr := &DegradedRegistry{
		registry:           registry,
		degradationManager: degradationManager,
		degradedCollectors: make(map[string]*DegradedCollector),
		logger:             logrus.WithField("component", "degraded_registry"),
	}

	return dr, nil
}

// Register registers a collector with degradation support
func (dr *DegradedRegistry) Register(name string, collector Collector) error {
	// Register with base registry
	if err := dr.registry.Register(name, collector); err != nil {
		return err
	}

	// Wrap with degradation support if enabled
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if dr.degradationManager.config.Enabled {
		degradedCollector := NewDegradedCollector(collector, dr.degradationManager)
		dr.degradedCollectors[name] = degradedCollector
		dr.logger.WithField("collector", name).Info("Registered collector with degradation support")
	}

	return nil
}

// Unregister removes a collector
func (dr *DegradedRegistry) Unregister(name string) error {
	dr.mu.Lock()
	delete(dr.degradedCollectors, name)
	dr.mu.Unlock()

	return dr.registry.Unregister(name)
}

// Get returns a collector by name
func (dr *DegradedRegistry) Get(name string) (Collector, bool) {
	// Check if we have a degraded wrapper
	dr.mu.RLock()
	if degraded, exists := dr.degradedCollectors[name]; exists {
		dr.mu.RUnlock()
		return degraded, true
	}
	dr.mu.RUnlock()

	// Fall back to base registry
	return dr.registry.Get(name)
}

// List returns all registered collector names
func (dr *DegradedRegistry) List() []string {
	return dr.registry.List()
}

// CollectAll performs concurrent collection from all collectors
func (dr *DegradedRegistry) CollectAll(ctx context.Context) ([]*CollectionResult, error) {
	// Use the base registry for concurrent collection since it implements the interface correctly
	cc := NewConcurrentCollector(dr.registry, dr.registry.config.Global.MaxConcurrency)
	return cc.CollectAll(ctx)
}

// GetDegradationManager returns the degradation manager
func (dr *DegradedRegistry) GetDegradationManager() *DegradationManager {
	return dr.degradationManager
}

// UpdateDegradationMode updates the overall degradation mode
func (dr *DegradedRegistry) UpdateDegradationMode() {
	dr.degradationManager.UpdateDegradationMode()
}

// GetDegradationStats returns degradation statistics
func (dr *DegradedRegistry) GetDegradationStats() map[string]DegradationStats {
	return dr.degradationManager.GetDegradationStats()
}

// ResetAllBreakers resets all circuit breakers
func (dr *DegradedRegistry) ResetAllBreakers() {
	dr.degradationManager.ResetAllBreakers()
}

// PrometheusCollector returns a Prometheus collector for this registry
func (dr *DegradedRegistry) PrometheusCollector() prometheus.Collector {
	return &degradedRegistryCollector{registry: dr}
}

// degradedRegistryCollector adapts DegradedRegistry to prometheus.Collector
type degradedRegistryCollector struct {
	registry *DegradedRegistry
}

// Describe implements prometheus.Collector
func (drc *degradedRegistryCollector) Describe(ch chan<- *prometheus.Desc) {
	// Collect descriptors from all registered collectors
	for _, name := range drc.registry.List() {
		if collector, exists := drc.registry.Get(name); exists {
			collector.Describe(ch)
		}
	}
}

// Collect implements prometheus.Collector
func (drc *degradedRegistryCollector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()
	for _, name := range drc.registry.List() {
		if collector, exists := drc.registry.Get(name); exists && collector.IsEnabled() {
			// Create a sub-channel to avoid blocking
			subCh := make(chan prometheus.Metric, 1000)
			go func() {
				defer close(subCh)
				_ = collector.Collect(ctx, subCh)
			}()

			// Forward metrics
			for metric := range subCh {
				ch <- metric
			}
		}
	}
}
