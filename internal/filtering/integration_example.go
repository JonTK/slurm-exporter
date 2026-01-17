package filtering

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
)

// FilteredCollectorManager demonstrates how to integrate smart filtering with collectors
type FilteredCollectorManager struct {
	filter     *SmartFilter
	collectors map[string]*FilteredCollector
	logger     *logrus.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// FilteredCollector represents a collector that uses smart filtering
type FilteredCollector struct {
	name        string
	collectFunc func(ctx context.Context) ([]*dto.MetricFamily, error)
	interval    time.Duration
	lastRun     time.Time
	nextRun     time.Time
	isRunning   bool
	mu          sync.RWMutex
}

// NewFilteredCollectorManager creates a new collector manager with smart filtering
func NewFilteredCollectorManager(
	filterCfg config.SmartFilteringConfig,
	logger *logrus.Logger,
) (*FilteredCollectorManager, error) {

	// Create smart filter
	filter, err := NewSmartFilter(filterCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create smart filter: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &FilteredCollectorManager{
		filter:     filter,
		collectors: make(map[string]*FilteredCollector),
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// RegisterCollector registers a new collector with the manager
func (fcm *FilteredCollectorManager) RegisterCollector(
	name string,
	interval time.Duration,
	collectFunc func(ctx context.Context) ([]*dto.MetricFamily, error),
) {
	fcm.mu.Lock()
	defer fcm.mu.Unlock()

	collector := &FilteredCollector{
		name:        name,
		collectFunc: collectFunc,
		interval:    interval,
		lastRun:     time.Now(),
		nextRun:     time.Now().Add(interval),
	}

	fcm.collectors[name] = collector

	fcm.logger.WithFields(logrus.Fields{
		"collector": name,
		"interval":  interval,
	}).Info("Registered collector with filtered manager")
}

// Start begins the filtered collection process
func (fcm *FilteredCollectorManager) Start() error {
	// Register filter metrics
	if err := fcm.filter.RegisterMetrics(prometheus.DefaultRegisterer); err != nil {
		return fmt.Errorf("failed to register filter metrics: %w", err)
	}

	// Start collection scheduling
	fcm.wg.Add(1)
	go fcm.scheduleCollections()

	fcm.logger.Info("Started filtered collector manager")
	return nil
}

// Stop gracefully stops the collector manager
func (fcm *FilteredCollectorManager) Stop() {
	fcm.cancel()
	fcm.wg.Wait()

	if fcm.filter != nil {
		_ = fcm.filter.Close()
	}

	fcm.logger.Info("Stopped filtered collector manager")
}

// scheduleCollections manages the collection scheduling
func (fcm *FilteredCollectorManager) scheduleCollections() {
	defer fcm.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-fcm.ctx.Done():
			return
		case <-ticker.C:
			fcm.runScheduledCollections()
		}
	}
}

// runScheduledCollections checks and runs any collectors that are due
func (fcm *FilteredCollectorManager) runScheduledCollections() {
	now := time.Now()

	fcm.mu.RLock()
	var dueCollectors []*FilteredCollector
	for _, collector := range fcm.collectors {
		collector.mu.RLock()
		if !collector.isRunning && now.After(collector.nextRun) {
			dueCollectors = append(dueCollectors, collector)
		}
		collector.mu.RUnlock()
	}
	fcm.mu.RUnlock()

	// Run due collectors
	for _, collector := range dueCollectors {
		fcm.wg.Add(1)
		go fcm.runFilteredCollector(collector)
	}
}

// runFilteredCollector executes a single collector with smart filtering
func (fcm *FilteredCollectorManager) runFilteredCollector(collector *FilteredCollector) {
	defer fcm.wg.Done()

	collector.mu.Lock()
	if collector.isRunning {
		collector.mu.Unlock()
		return
	}
	collector.isRunning = true
	collector.mu.Unlock()

	defer func() {
		collector.mu.Lock()
		collector.isRunning = false
		collector.nextRun = time.Now().Add(collector.interval)
		collector.mu.Unlock()
	}()

	startTime := time.Now()

	// Execute collection
	ctx, cancel := context.WithTimeout(fcm.ctx, 30*time.Second)
	defer cancel()

	rawMetrics, err := collector.collectFunc(ctx)
	if err != nil {
		fcm.logger.WithFields(logrus.Fields{
			"collector": collector.name,
			"error":     err,
		}).Error("Collection failed")
		return
	}

	// Apply smart filtering
	filteredMetrics, err := fcm.filter.ProcessMetrics(ctx, collector.name, rawMetrics)
	if err != nil {
		fcm.logger.WithFields(logrus.Fields{
			"collector": collector.name,
			"error":     err,
		}).Error("Filtering failed")
		return
	}

	duration := time.Since(startTime)

	// Calculate filtering statistics
	rawCount := fcm.countMetrics(rawMetrics)
	filteredCount := fcm.countMetrics(filteredMetrics)
	filterRate := 0.0
	if rawCount > 0 {
		filterRate = float64(rawCount-filteredCount) / float64(rawCount) * 100
	}

	// Update collector state
	collector.mu.Lock()
	collector.lastRun = startTime
	collector.mu.Unlock()

	// Log collection result
	fcm.logger.WithFields(logrus.Fields{
		"collector":        collector.name,
		"duration":         duration,
		"raw_metrics":      rawCount,
		"filtered_metrics": filteredCount,
		"filter_rate":      fmt.Sprintf("%.1f%%", filterRate),
	}).Info("Filtered collection completed")
}

// countMetrics counts the total number of metric samples
func (fcm *FilteredCollectorManager) countMetrics(families []*dto.MetricFamily) int {
	count := 0
	for _, family := range families {
		count += len(family.Metric)
	}
	return count
}

// GetCollectorStatus returns the current status of all collectors
func (fcm *FilteredCollectorManager) GetCollectorStatus() map[string]CollectorStatus {
	fcm.mu.RLock()
	defer fcm.mu.RUnlock()

	status := make(map[string]CollectorStatus)
	for name, collector := range fcm.collectors {
		collector.mu.RLock()
		status[name] = CollectorStatus{
			Name:      name,
			LastRun:   collector.lastRun,
			NextRun:   collector.nextRun,
			IsRunning: collector.isRunning,
			Interval:  collector.interval,
		}
		collector.mu.RUnlock()
	}

	return status
}

// CollectorStatus represents the current status of a collector
type CollectorStatus struct {
	Name      string        `json:"name"`
	LastRun   time.Time     `json:"last_run"`
	NextRun   time.Time     `json:"next_run"`
	IsRunning bool          `json:"is_running"`
	Interval  time.Duration `json:"interval"`
}

// GetFilterStats returns statistics about the smart filter
func (fcm *FilteredCollectorManager) GetFilterStats() map[string]interface{} {
	return fcm.filter.GetStats()
}

// GetFilterPatterns returns learned patterns from the filter
func (fcm *FilteredCollectorManager) GetFilterPatterns() map[string]MetricPattern {
	return fcm.filter.GetPatterns()
}

// MockCollectorFunction creates a mock collector function for demonstration
func MockCollectorFunction(name string, metricCount int, withNoise bool) func(ctx context.Context) ([]*dto.MetricFamily, error) {
	return func(ctx context.Context) ([]*dto.MetricFamily, error) {
		var families []*dto.MetricFamily

		// Create metrics with varying noise patterns
		for i := 0; i < metricCount; i++ {
			metricName := fmt.Sprintf("%s_metric_%d", name, i)

			var value float64
			if withNoise && i%3 == 0 {
				// Add noisy metrics (high variance)
				value = float64(time.Now().UnixNano()%1000) + float64(i)*10
			} else {
				// Stable metrics
				value = float64(i) * 10.0
			}

			family := &dto.MetricFamily{
				Name: &metricName,
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Label: []*dto.LabelPair{
							{
								Name:  stringPtr("instance"),
								Value: stringPtr(fmt.Sprintf("node_%d", i%5)),
							},
							{
								Name:  stringPtr("job"),
								Value: stringPtr(fmt.Sprintf("job_%d", i%10)),
							},
						},
						Gauge: &dto.Gauge{
							Value: &value,
						},
					},
				},
			}
			families = append(families, family)
		}

		return families, nil
	}
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// SimulateNoisyMetrics creates metrics that will be identified as noisy
func SimulateNoisyMetrics() []*dto.MetricFamily {
	var families []*dto.MetricFamily

	// Create a metric with high variance (should be filtered)
	noisyValue := float64(time.Now().UnixNano() % 10000)
	noisyFamily := &dto.MetricFamily{
		Name: stringPtr("noisy_metric"),
		Type: dto.MetricType_GAUGE.Enum(),
		Metric: []*dto.Metric{
			{
				Gauge: &dto.Gauge{Value: &noisyValue},
			},
		},
	}
	families = append(families, noisyFamily)

	// Create a stable metric (should be kept)
	stableValue := 42.0
	stableFamily := &dto.MetricFamily{
		Name: stringPtr("stable_metric"),
		Type: dto.MetricType_GAUGE.Enum(),
		Metric: []*dto.Metric{
			{
				Gauge: &dto.Gauge{Value: &stableValue},
			},
		},
	}
	families = append(families, stableFamily)

	return families
}

// ExampleUsage demonstrates how to use the filtered collection system
func ExampleUsage() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create smart filtering configuration
	filterCfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.7, // Filter metrics with noise score > 0.7
		CacheSize:      1000,
		LearningWindow: 20, // Learn from last 20 samples
		VarianceLimit:  100.0,
		CorrelationMin: 0.1,
	}

	// Create filtered collector manager
	manager, err := NewFilteredCollectorManager(filterCfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create filtered collector manager")
	}

	// Register some example collectors
	manager.RegisterCollector("jobs", 30*time.Second,
		MockCollectorFunction("jobs", 20, true)) // With noise

	manager.RegisterCollector("nodes", 45*time.Second,
		MockCollectorFunction("nodes", 15, false)) // Stable

	manager.RegisterCollector("partitions", 60*time.Second,
		MockCollectorFunction("partitions", 10, true)) // With noise

	// Start the manager
	if err := manager.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start filtered collector manager")
	}

	// Let it run for a while to learn patterns
	logger.Info("Running filtered collections for 5 minutes...")
	time.Sleep(5 * time.Minute)

	// Print statistics
	stats := manager.GetFilterStats()
	logger.WithField("filter_stats", stats).Info("Filter statistics")

	status := manager.GetCollectorStatus()
	logger.WithField("collector_status", status).Info("Collector status")

	// Print learned patterns
	patterns := manager.GetFilterPatterns()
	logger.WithField("pattern_count", len(patterns)).Info("Learned patterns")

	for key, pattern := range patterns {
		if len(pattern.Values) > 5 { // Only show patterns with enough data
			logger.WithFields(logrus.Fields{
				"metric":      key,
				"noise_score": pattern.NoiseScore,
				"mean":        pattern.Mean,
				"variance":    pattern.Variance,
				"samples":     len(pattern.Values),
				"action":      pattern.FilterRecommend.String(),
			}).Info("Pattern learned")
		}
	}

	// Stop the manager
	manager.Stop()
}

// AdvancedFilteringExample demonstrates advanced filtering scenarios
func AdvancedFilteringExample() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create aggressive filtering configuration
	filterCfg := config.SmartFilteringConfig{
		Enabled:        true,
		NoiseThreshold: 0.5, // More aggressive filtering
		CacheSize:      500,
		LearningWindow: 10, // Faster learning
		VarianceLimit:  50.0,
		CorrelationMin: 0.2,
	}

	filter, err := NewSmartFilter(filterCfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create smart filter")
	}
	defer func() { _ = filter.Close() }()

	// Simulate collection cycles with different noise patterns
	for cycle := 0; cycle < 20; cycle++ {
		logger.WithField("cycle", cycle).Info("Processing collection cycle")

		// Generate metrics with different patterns
		metrics := []*dto.MetricFamily{}

		// Add stable metrics
		for i := 0; i < 5; i++ {
			value := float64(i * 10) // Constant values
			metric := &dto.MetricFamily{
				Name: stringPtr(fmt.Sprintf("stable_metric_%d", i)),
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Gauge: &dto.Gauge{Value: &value},
					},
				},
			}
			metrics = append(metrics, metric)
		}

		// Add noisy metrics
		for i := 0; i < 5; i++ {
			value := float64(time.Now().UnixNano()%1000 + int64(i*100)) // Random values
			metric := &dto.MetricFamily{
				Name: stringPtr(fmt.Sprintf("noisy_metric_%d", i)),
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Gauge: &dto.Gauge{Value: &value},
					},
				},
			}
			metrics = append(metrics, metric)
		}

		// Add trending metrics
		for i := 0; i < 3; i++ {
			value := float64(cycle*5 + i) // Increasing trend
			metric := &dto.MetricFamily{
				Name: stringPtr(fmt.Sprintf("trending_metric_%d", i)),
				Type: dto.MetricType_GAUGE.Enum(),
				Metric: []*dto.Metric{
					{
						Gauge: &dto.Gauge{Value: &value},
					},
				},
			}
			metrics = append(metrics, metric)
		}

		// Process through filter
		filtered, err := filter.ProcessMetrics(context.Background(), "test_collector", metrics)
		if err != nil {
			logger.WithError(err).Error("Filtering failed")
			continue
		}

		rawCount := 0
		filteredCount := 0
		for _, family := range metrics {
			rawCount += len(family.Metric)
		}
		for _, family := range filtered {
			filteredCount += len(family.Metric)
		}

		filterRate := 0.0
		if rawCount > 0 {
			filterRate = float64(rawCount-filteredCount) / float64(rawCount) * 100
		}

		logger.WithFields(logrus.Fields{
			"cycle":            cycle,
			"raw_metrics":      rawCount,
			"filtered_metrics": filteredCount,
			"filter_rate":      fmt.Sprintf("%.1f%%", filterRate),
		}).Info("Cycle completed")

		// Small delay between cycles
		time.Sleep(500 * time.Millisecond)
	}

	// Print final statistics
	stats := filter.GetStats()
	logger.WithField("final_stats", stats).Info("Final filter statistics")

	patterns := filter.GetPatterns()
	logger.WithField("total_patterns", len(patterns)).Info("Total patterns learned")
}
