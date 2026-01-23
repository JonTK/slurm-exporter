// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package adaptive

import (
	"context"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ExampleCollectorManager demonstrates how to integrate the adaptive scheduler with collectors
type ExampleCollectorManager struct {
	scheduler  *CollectorScheduler
	collectors map[string]*ExampleCollector
	logger     *logrus.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex

	// SLURM client for monitoring cluster state
	slurmClient ExampleSLURMClient
}

// ExampleSLURMClient interface for demonstration
type ExampleSLURMClient interface {
	GetJobCount(ctx context.Context) (int, error)
	GetNodeCount(ctx context.Context) (int, error)
	GetClusterMetrics(ctx context.Context) (map[string]interface{}, error)
}

// ExampleCollector represents a collector that uses adaptive scheduling
type ExampleCollector struct {
	name        string
	collectFunc func(ctx context.Context) error
	lastRun     time.Time
	nextRun     time.Time
	isRunning   bool
	mu          sync.RWMutex
}

// NewExampleCollectorManager creates a new collector manager with adaptive scheduling
func NewExampleCollectorManager(
	cfg config.AdaptiveCollectionConfig,
	slurmClient ExampleSLURMClient,
	logger *logrus.Logger,
) (*ExampleCollectorManager, error) {
	// Create adaptive scheduler
	scheduler, err := NewCollectorScheduler(cfg, 30*time.Second, logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ExampleCollectorManager{
		scheduler:   scheduler,
		collectors:  make(map[string]*ExampleCollector),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		slurmClient: slurmClient,
	}, nil
}

// RegisterCollector registers a new collector with the manager
func (cm *ExampleCollectorManager) RegisterCollector(name string, collectFunc func(ctx context.Context) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	collector := &ExampleCollector{
		name:        name,
		collectFunc: collectFunc,
		lastRun:     time.Now(),
		nextRun:     time.Now().Add(cm.scheduler.GetCollectionInterval(name)),
	}

	cm.collectors[name] = collector
	cm.scheduler.RegisterCollector(name)

	cm.logger.WithFields(logrus.Fields{
		"collector": name,
		"interval":  cm.scheduler.GetCollectionInterval(name),
	}).Info("Registered collector with adaptive manager")
}

// Start begins the adaptive collection process
func (cm *ExampleCollectorManager) Start() error {
	// Register scheduler metrics
	if err := cm.scheduler.RegisterMetrics(prometheus.DefaultRegisterer); err != nil {
		return err
	}

	// Start activity monitoring
	cm.wg.Add(1)
	go cm.monitorActivity()

	// Start collection scheduling
	cm.wg.Add(1)
	go cm.scheduleCollections()

	cm.logger.Info("Started adaptive collector manager")
	return nil
}

// Stop gracefully stops the collector manager
func (cm *ExampleCollectorManager) Stop() {
	cm.cancel()
	cm.wg.Wait()
	cm.logger.Info("Stopped adaptive collector manager")
}

// monitorActivity continuously monitors cluster activity and updates the scheduler
func (cm *ExampleCollectorManager) monitorActivity() {
	defer cm.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.updateClusterActivity()
		}
	}
}

// updateClusterActivity fetches current cluster state and updates the scheduler
func (cm *ExampleCollectorManager) updateClusterActivity() {
	ctx, cancel := context.WithTimeout(cm.ctx, 10*time.Second)
	defer cancel()

	// Get current cluster metrics
	jobCount, err := cm.slurmClient.GetJobCount(ctx)
	if err != nil {
		cm.logger.WithError(err).Warn("Failed to get job count for activity monitoring")
		return
	}

	nodeCount, err := cm.slurmClient.GetNodeCount(ctx)
	if err != nil {
		cm.logger.WithError(err).Warn("Failed to get node count for activity monitoring")
		return
	}

	clusterMetrics, err := cm.slurmClient.GetClusterMetrics(ctx)
	if err != nil {
		cm.logger.WithError(err).Warn("Failed to get cluster metrics for activity monitoring")
		clusterMetrics = make(map[string]interface{})
	}

	// Add job and node counts to metrics for change detection
	clusterMetrics["job_count"] = jobCount
	clusterMetrics["node_count"] = nodeCount

	// Update scheduler with activity data
	cm.scheduler.UpdateActivity(ctx, jobCount, nodeCount, clusterMetrics)

	cm.logger.WithFields(logrus.Fields{
		"job_count":      jobCount,
		"node_count":     nodeCount,
		"activity_score": cm.scheduler.GetCurrentScore(),
	}).Debug("Updated cluster activity")
}

// scheduleCollections manages the adaptive collection scheduling
func (cm *ExampleCollectorManager) scheduleCollections() {
	defer cm.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.runScheduledCollections()
		}
	}
}

// runScheduledCollections checks and runs any collectors that are due
func (cm *ExampleCollectorManager) runScheduledCollections() {
	now := time.Now()

	cm.mu.RLock()
	var dueCollectors []*ExampleCollector
	for _, collector := range cm.collectors {
		collector.mu.RLock()
		if !collector.isRunning && now.After(collector.nextRun) {
			dueCollectors = append(dueCollectors, collector)
		}
		collector.mu.RUnlock()
	}
	cm.mu.RUnlock()

	// Run due collectors
	for _, collector := range dueCollectors {
		cm.wg.Add(1)
		go cm.runCollector(collector, now)
	}
}

// runCollector executes a single collector with adaptive timing
func (cm *ExampleCollectorManager) runCollector(collector *ExampleCollector, scheduledTime time.Time) {
	defer cm.wg.Done()

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
		collector.mu.Unlock()
	}()

	actualStartTime := time.Now()

	// Record collection timing
	cm.scheduler.RecordCollection(collector.name, scheduledTime, actualStartTime)

	// Execute collection
	ctx, cancel := context.WithTimeout(cm.ctx, 30*time.Second)
	defer cancel()

	err := collector.collectFunc(ctx)
	duration := time.Since(actualStartTime)

	// Update collector state
	collector.mu.Lock()
	collector.lastRun = actualStartTime

	// Get current adaptive interval
	interval := cm.scheduler.GetCollectionInterval(collector.name)
	collector.nextRun = actualStartTime.Add(interval)
	collector.mu.Unlock()

	// Log collection result
	logFields := logrus.Fields{
		"collector":      collector.name,
		"duration":       duration,
		"next_run":       collector.nextRun,
		"interval":       interval,
		"activity_score": cm.scheduler.GetCurrentScore(),
	}

	if err != nil {
		logFields["error"] = err
		cm.logger.WithFields(logFields).Error("Collection failed")
	} else {
		cm.logger.WithFields(logFields).Debug("Collection completed")
	}
}

// GetCollectorStatus returns the current status of all collectors
func (cm *ExampleCollectorManager) GetCollectorStatus() map[string]CollectorStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]CollectorStatus)
	for name, collector := range cm.collectors {
		collector.mu.RLock()
		status[name] = CollectorStatus{
			Name:            name,
			LastRun:         collector.lastRun,
			NextRun:         collector.nextRun,
			IsRunning:       collector.isRunning,
			CurrentInterval: cm.scheduler.GetCollectionInterval(name),
			ActivityScore:   cm.scheduler.GetCurrentScore(),
		}
		collector.mu.RUnlock()
	}

	return status
}

// CollectorStatus represents the current status of a collector
type CollectorStatus struct {
	Name            string        `json:"name"`
	LastRun         time.Time     `json:"last_run"`
	NextRun         time.Time     `json:"next_run"`
	IsRunning       bool          `json:"is_running"`
	CurrentInterval time.Duration `json:"current_interval"`
	ActivityScore   float64       `json:"activity_score"`
}

// GetSchedulerStats returns statistics about the adaptive scheduler
func (cm *ExampleCollectorManager) GetSchedulerStats() map[string]interface{} {
	return cm.scheduler.GetStats()
}

// MockSLURMClient provides a mock implementation for demonstration
type MockSLURMClient struct {
	jobCount  int
	nodeCount int
	metrics   map[string]interface{}
	mu        sync.RWMutex
}

// NewMockSLURMClient creates a new mock SLURM client
func NewMockSLURMClient() *MockSLURMClient {
	return &MockSLURMClient{
		jobCount:  100,
		nodeCount: 50,
		metrics: map[string]interface{}{
			"queue_length":   25,
			"cluster_load":   0.7,
			"pending_jobs":   10,
			"running_jobs":   90,
			"failed_jobs":    5,
			"completed_jobs": 1000,
		},
	}
}

func (m *MockSLURMClient) GetJobCount(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.jobCount, nil
}

func (m *MockSLURMClient) GetNodeCount(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodeCount, nil
}

func (m *MockSLURMClient) GetClusterMetrics(ctx context.Context) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid concurrent map access
	metrics := make(map[string]interface{})
	for k, v := range m.metrics {
		metrics[k] = v
	}
	return metrics, nil
}

// SetJobCount updates the mock job count (for testing activity changes)
func (m *MockSLURMClient) SetJobCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobCount = count
}

// SetNodeCount updates the mock node count
func (m *MockSLURMClient) SetNodeCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodeCount = count
}

// UpdateMetric updates a specific metric value
func (m *MockSLURMClient) UpdateMetric(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[key] = value
}

// ExampleUsage demonstrates how to use the adaptive collection system
func ExampleUsage() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create adaptive configuration
	cfg := config.AdaptiveCollectionConfig{
		Enabled:      true,
		MinInterval:  10 * time.Second,
		MaxInterval:  2 * time.Minute,
		BaseInterval: 30 * time.Second,
		ScoreWindow:  5 * time.Minute,
	}

	// Create mock SLURM client
	slurmClient := NewMockSLURMClient()

	// Create collector manager
	manager, err := NewExampleCollectorManager(cfg, slurmClient, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create collector manager")
	}

	// Register some example collectors
	manager.RegisterCollector("jobs", func(ctx context.Context) error {
		logger.Info("Collecting job metrics")
		time.Sleep(500 * time.Millisecond) // Simulate collection work
		return nil
	})

	manager.RegisterCollector("nodes", func(ctx context.Context) error {
		logger.Info("Collecting node metrics")
		time.Sleep(300 * time.Millisecond) // Simulate collection work
		return nil
	})

	manager.RegisterCollector("partitions", func(ctx context.Context) error {
		logger.Info("Collecting partition metrics")
		time.Sleep(200 * time.Millisecond) // Simulate collection work
		return nil
	})

	// Start the manager
	if err := manager.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start collector manager")
	}

	// Simulate some activity changes
	go func() {
		time.Sleep(1 * time.Minute)
		logger.Info("Simulating high activity")
		slurmClient.SetJobCount(500) // Increase job count
		slurmClient.UpdateMetric("queue_length", 100)

		time.Sleep(2 * time.Minute)
		logger.Info("Simulating low activity")
		slurmClient.SetJobCount(20) // Decrease job count
		slurmClient.UpdateMetric("queue_length", 5)
	}()

	// Run for 5 minutes
	time.Sleep(5 * time.Minute)

	// Stop the manager
	manager.Stop()

	// Print final statistics
	stats := manager.GetSchedulerStats()
	logger.WithField("stats", stats).Info("Final scheduler statistics")

	status := manager.GetCollectorStatus()
	logger.WithField("status", status).Info("Final collector status")
}
