// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/performance"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ProfiledCollector wraps a collector with profiling capabilities
type ProfiledCollector struct {
	collector Collector
	profiler  *performance.Profiler
	logger    *logrus.Entry
	mu        sync.RWMutex
	enabled   bool
}

// NewProfiledCollector creates a new profiled collector wrapper
func NewProfiledCollector(
	coll Collector,
	profiler *performance.Profiler,
	logger *logrus.Entry,
) (*ProfiledCollector, error) {
	if coll == nil {
		return nil, fmt.Errorf("collector cannot be nil")
	}

	return &ProfiledCollector{
		collector: coll,
		profiler:  profiler,
		logger:    logger.WithField("collector", coll.Name()),
		enabled:   true,
	}, nil
}

// Name returns the collector name
func (pc *ProfiledCollector) Name() string {
	return pc.collector.Name()
}

// Describe implements prometheus.Collector
func (pc *ProfiledCollector) Describe(ch chan<- *prometheus.Desc) {
	pc.collector.Describe(ch)
}

// Collect implements the Collector interface with profiling
func (pc *ProfiledCollector) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	pc.mu.RLock()
	if !pc.enabled {
		pc.mu.RUnlock()
		if err := pc.collector.Collect(ctx, ch); err != nil {
			return fmt.Errorf("profiled collector (disabled) failed: %w", err)
		}
		return nil
	}
	pc.mu.RUnlock()

	// Start profiling operation
	op := pc.profiler.StartOperation(pc.collector.Name())
	defer op.Stop()

	// Phase: Pre-collection
	op.Phase("pre_collection")
	pc.logger.Debug("Starting profiled collection")
	op.PhaseEnd("pre_collection")

	// Phase: Collection
	op.Phase("collection")
	err := pc.collector.Collect(ctx, ch)
	op.PhaseEnd("collection")

	// Phase: Post-collection
	op.Phase("post_collection")

	// Log collection results
	if err != nil {
		pc.logger.WithError(err).Error("Collection failed")
		// Metadata will be added when profile is saved
		op.PhaseEnd("post_collection")
		return fmt.Errorf("profiled collection failed: %w", err)
	}

	pc.logger.WithField("duration", op.Duration()).Debug("Collection completed")
	// Metadata will be added when profile is saved

	op.PhaseEnd("post_collection")

	// Generate and log report if collection was slow
	duration := op.Duration()
	if duration > 5*time.Second {
		report := op.GenerateReport()
		pc.logger.WithField("report", report).Warn("Slow collection detected")
	}

	return nil
}

// IsEnabled returns whether this collector is enabled
func (pc *ProfiledCollector) IsEnabled() bool {
	return pc.collector.IsEnabled()
}

// SetEnabled enables or disables the collector
func (pc *ProfiledCollector) SetEnabled(enabled bool) {
	pc.collector.SetEnabled(enabled)
}

// SetProfilingEnabled enables or disables profiling for this collector
func (pc *ProfiledCollector) SetProfilingEnabled(enabled bool) {
	pc.mu.Lock()
	pc.enabled = enabled
	pc.mu.Unlock()
}

// GetProfile returns the current profile if one is active
func (pc *ProfiledCollector) GetProfile() *performance.CollectorProfile {
	return pc.profiler.GetProfile(pc.collector.Name())
}

// ListProfiles returns all stored profiles for this collector
func (pc *ProfiledCollector) ListProfiles() ([]*performance.ProfileMetadata, error) {
	allProfiles, err := pc.profiler.ListProfiles()
	if err != nil {
		return nil, err
	}

	// Filter profiles for this collector
	var profiles []*performance.ProfileMetadata
	for _, p := range allProfiles {
		if p.CollectorName == pc.collector.Name() {
			profiles = append(profiles, p)
		}
	}

	return profiles, nil
}

// ProfiledCollectorManager manages profiled collectors
type ProfiledCollectorManager struct {
	profiler   *performance.Profiler
	logger     *logrus.Entry
	mu         sync.RWMutex
	collectors map[string]*ProfiledCollector
}

// NewProfiledCollectorManager creates a new profiled collector manager
func NewProfiledCollectorManager(
	profiler *performance.Profiler,
	logger *logrus.Entry,
) *ProfiledCollectorManager {
	return &ProfiledCollectorManager{
		profiler:   profiler,
		logger:     logger,
		collectors: make(map[string]*ProfiledCollector),
	}
}

// WrapCollector wraps a collector with profiling
func (pcm *ProfiledCollectorManager) WrapCollector(coll Collector) (Collector, error) {
	pcm.mu.Lock()
	defer pcm.mu.Unlock()

	// Check if already wrapped
	if pc, exists := pcm.collectors[coll.Name()]; exists {
		return pc, nil
	}

	// Create profiled collector
	pc, err := NewProfiledCollector(coll, pcm.profiler, pcm.logger)
	if err != nil {
		return nil, err
	}

	pcm.collectors[coll.Name()] = pc
	return pc, nil
}

// SetProfilingEnabled enables or disables profiling for a specific collector
func (pcm *ProfiledCollectorManager) SetProfilingEnabled(collectorName string, enabled bool) error {
	pcm.mu.RLock()
	pc, exists := pcm.collectors[collectorName]
	pcm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collector not found: %s", collectorName)
	}

	pc.SetProfilingEnabled(enabled)
	return nil
}

// SetProfilingEnabledAll enables or disables profiling for all collectors
func (pcm *ProfiledCollectorManager) SetProfilingEnabledAll(enabled bool) {
	pcm.mu.RLock()
	collectors := make([]*ProfiledCollector, 0, len(pcm.collectors))
	for _, pc := range pcm.collectors {
		collectors = append(collectors, pc)
	}
	pcm.mu.RUnlock()

	for _, pc := range collectors {
		pc.SetProfilingEnabled(enabled)
	}
}

// GetCollectorProfiles returns profiles for a specific collector
func (pcm *ProfiledCollectorManager) GetCollectorProfiles(collectorName string) ([]*performance.ProfileMetadata, error) {
	pcm.mu.RLock()
	pc, exists := pcm.collectors[collectorName]
	pcm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collector not found: %s", collectorName)
	}

	return pc.ListProfiles()
}

// GetAllProfiles returns profiles for all collectors
func (pcm *ProfiledCollectorManager) GetAllProfiles() (map[string][]*performance.ProfileMetadata, error) {
	pcm.mu.RLock()
	collectors := make(map[string]*ProfiledCollector)
	for name, pc := range pcm.collectors {
		collectors[name] = pc
	}
	pcm.mu.RUnlock()

	result := make(map[string][]*performance.ProfileMetadata)
	for name, pc := range collectors {
		profiles, err := pc.ListProfiles()
		if err != nil {
			pcm.logger.WithError(err).WithField("collector", name).Error("Failed to list profiles")
			continue
		}
		result[name] = profiles
	}

	return result, nil
}

// GetStats returns profiling statistics
func (pcm *ProfiledCollectorManager) GetStats() map[string]interface{} {
	pcm.mu.RLock()
	defer pcm.mu.RUnlock()

	collectorStats := make(map[string]interface{})
	for name, pc := range pcm.collectors {
		collectorStats[name] = map[string]interface{}{
			"profiling_enabled": pc.enabled,
			"collector_enabled": pc.IsEnabled(),
		}
	}

	return map[string]interface{}{
		"total_collectors": len(pcm.collectors),
		"collectors":       collectorStats,
		"profiler_stats":   pcm.profiler.GetStats(),
	}
}
