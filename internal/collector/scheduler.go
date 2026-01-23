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

	"github.com/jontk/slurm-exporter/internal/config"
)

// Scheduler manages collection scheduling for all collectors
type Scheduler struct {
	registry      *Registry
	orchestrator  *CollectionOrchestrator
	config        *config.CollectorsConfig
	mu            sync.RWMutex
	schedules     map[string]*Schedule
	globalMetrics *SchedulerMetrics
	logger        *logrus.Entry
	ctx           context.Context
	cancel        context.CancelFunc
	stopped       bool // Protected by mu
}

// Schedule represents a collection schedule for a collector
type Schedule struct {
	CollectorName string
	Interval      time.Duration
	Timeout       time.Duration
	LastRun       time.Time
	NextRun       time.Time
	RunCount      int64
	ErrorCount    int64
	timer         *time.Timer
	enabled       bool
	mu            sync.RWMutex
}

// SchedulerMetrics tracks scheduler performance
type SchedulerMetrics struct {
	ScheduledRuns   *prometheus.CounterVec
	MissedRuns      *prometheus.CounterVec
	ScheduleLatency *prometheus.HistogramVec
}

// NewSchedulerMetrics creates metrics for the scheduler
func NewSchedulerMetrics(namespace, subsystem string) *SchedulerMetrics {
	return &SchedulerMetrics{
		ScheduledRuns: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "scheduled_collections_total",
				Help:      "Total number of scheduled collections",
			},
			[]string{"collector", "status"},
		),
		MissedRuns: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "missed_collections_total",
				Help:      "Total number of missed scheduled collections",
			},
			[]string{"collector", "reason"},
		),
		ScheduleLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "schedule_latency_seconds",
				Help:      "Latency between scheduled and actual collection time",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"collector"},
		),
	}
}

// Register registers scheduler metrics
func (m *SchedulerMetrics) Register(registry *prometheus.Registry) error {
	collectors := []prometheus.Collector{
		m.ScheduledRuns,
		m.MissedRuns,
		m.ScheduleLatency,
	}

	for _, c := range collectors {
		if err := registry.Register(c); err != nil {
			return err
		}
	}

	return nil
}

// NewScheduler creates a new collection scheduler
func NewScheduler(registry *Registry, config *config.CollectorsConfig) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create metrics
	metrics := NewSchedulerMetrics("slurm", "scheduler")
	if err := metrics.Register(registry.promRegistry); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to register scheduler metrics: %w", err)
	}

	// Create orchestrator
	orchestrator := NewCollectionOrchestrator(registry, config.Global.MaxConcurrency)

	scheduler := &Scheduler{
		registry:      registry,
		orchestrator:  orchestrator,
		config:        config,
		schedules:     make(map[string]*Schedule),
		globalMetrics: metrics,
		logger:        logrus.WithField("component", "scheduler"),
		ctx:           ctx,
		cancel:        cancel,
	}

	return scheduler, nil
}

// InitializeSchedules creates schedules from configuration
func (s *Scheduler) InitializeSchedules() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Helper function to create schedule
	createSchedule := func(name string, cfg config.CollectorConfig) error {
		interval := cfg.Interval
		if interval <= 0 {
			interval = s.config.Global.DefaultInterval
		}

		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = s.config.Global.DefaultTimeout
		}

		schedule := &Schedule{
			CollectorName: name,
			Interval:      interval,
			Timeout:       timeout,
			enabled:       cfg.Enabled,
		}

		s.schedules[name] = schedule

		// Set interval in orchestrator
		s.orchestrator.SetCollectorInterval(name, interval)

		s.logger.WithFields(logrus.Fields{
			"collector": name,
			"interval":  interval,
			"timeout":   timeout,
			"enabled":   cfg.Enabled,
		}).Info("Initialized collection schedule")

		return nil
	}

	// Create schedules for each collector type
	scheduleConfigs := map[string]config.CollectorConfig{
		"cluster":     s.config.Cluster,
		"nodes":       s.config.Nodes,
		"jobs":        s.config.Jobs,
		"users":       s.config.Users,
		"partitions":  s.config.Partitions,
		"performance": s.config.Performance,
		"system":      s.config.System,
	}

	for name, cfg := range scheduleConfigs {
		if err := createSchedule(name, cfg); err != nil {
			return fmt.Errorf("failed to create schedule for %s: %w", name, err)
		}
	}

	return nil
}

// Start begins scheduled collection
func (s *Scheduler) Start() error {
	s.logger.Info("Starting collection scheduler")

	// Reset stopped flag
	s.mu.Lock()
	s.stopped = false
	s.mu.Unlock()

	// Start orchestrator
	s.orchestrator.Start()

	// Start schedule monitoring
	s.mu.RLock()
	for name, schedule := range s.schedules {
		if schedule.enabled {
			s.startSchedule(name, schedule)
		}
	}
	s.mu.RUnlock()

	// Monitor for schedule updates
	go s.monitorSchedules()

	return nil
}

// Stop stops all scheduled collections
func (s *Scheduler) Stop() {
	s.logger.Info("Stopping collection scheduler")

	// Set stopped flag to prevent new timers from being scheduled
	s.mu.Lock()
	s.stopped = true
	s.mu.Unlock()

	// Cancel context first - this will signal all goroutines to stop
	s.cancel()

	// Stop orchestrator
	s.orchestrator.Stop()

	// Collect all timers while holding the lock briefly
	// This avoids the deadlock with runScheduledCollection which holds
	// schedule.mu and tries to acquire s.mu.RLock
	s.mu.RLock()
	timers := make([]*time.Timer, 0, len(s.schedules))
	for _, schedule := range s.schedules {
		schedule.mu.Lock()
		if schedule.timer != nil {
			timers = append(timers, schedule.timer)
			schedule.timer = nil
		}
		schedule.mu.Unlock()
	}
	s.mu.RUnlock()

	// Stop all timers outside of any locks
	for _, timer := range timers {
		timer.Stop()
	}
}

// startSchedule starts a collection schedule
func (s *Scheduler) startSchedule(name string, schedule *Schedule) {
	schedule.mu.Lock()
	defer schedule.mu.Unlock()

	// Calculate next run time
	now := time.Now()
	schedule.NextRun = now.Add(schedule.Interval)

	// Create timer
	schedule.timer = time.AfterFunc(schedule.Interval, func() {
		s.runScheduledCollection(name, schedule)
	})

	s.logger.WithFields(logrus.Fields{
		"collector": name,
		"next_run":  schedule.NextRun,
	}).Debug("Started collection schedule")
}

// runScheduledCollection executes a scheduled collection
func (s *Scheduler) runScheduledCollection(name string, schedule *Schedule) {
	// Check if still enabled
	schedule.mu.RLock()
	if !schedule.enabled {
		schedule.mu.RUnlock()
		return
	}
	expectedTime := schedule.NextRun
	schedule.mu.RUnlock()

	// Calculate latency
	actualTime := time.Now()
	latency := actualTime.Sub(expectedTime)

	// Record latency metric
	s.globalMetrics.ScheduleLatency.WithLabelValues(name).Observe(latency.Seconds())

	// Check if we're too late
	if latency > schedule.Interval {
		s.globalMetrics.MissedRuns.WithLabelValues(name, "latency").Inc()
		s.logger.WithFields(logrus.Fields{
			"collector": name,
			"latency":   latency,
		}).Warn("Missed scheduled collection due to latency")
	}

	// Run collection
	result, err := s.orchestrator.CollectNow(name)

	// Update schedule stats
	schedule.mu.Lock()
	schedule.LastRun = actualTime
	schedule.RunCount++
	if err != nil {
		schedule.ErrorCount++
		s.globalMetrics.ScheduledRuns.WithLabelValues(name, "error").Inc()
	} else {
		s.globalMetrics.ScheduledRuns.WithLabelValues(name, "success").Inc()
	}

	// Schedule next run only if not stopped
	s.mu.RLock()
	stopped := s.stopped
	s.mu.RUnlock()

	if !stopped {
		schedule.NextRun = actualTime.Add(schedule.Interval)
		schedule.timer = time.AfterFunc(schedule.Interval, func() {
			s.runScheduledCollection(name, schedule)
		})
	}
	schedule.mu.Unlock()

	// Log result
	if err != nil {
		s.logger.WithError(err).WithField("collector", name).Error("Scheduled collection failed")
	} else if result != nil {
		s.logger.WithFields(logrus.Fields{
			"collector":    name,
			"duration":     result.Duration,
			"metric_count": result.MetricCount,
		}).Debug("Scheduled collection completed")
	}
}

// monitorSchedules monitors for schedule changes
func (s *Scheduler) monitorSchedules() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkScheduleHealth()
		case <-s.ctx.Done():
			return
		}
	}
}

// checkScheduleHealth checks the health of all schedules
func (s *Scheduler) checkScheduleHealth() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()

	for name, schedule := range s.schedules {
		schedule.mu.RLock()

		// Check if schedule is running behind
		if schedule.enabled && schedule.NextRun.Before(now.Add(-schedule.Interval)) {
			s.logger.WithFields(logrus.Fields{
				"collector":   name,
				"next_run":    schedule.NextRun,
				"interval":    schedule.Interval,
				"run_count":   schedule.RunCount,
				"error_count": schedule.ErrorCount,
			}).Warn("Schedule is running behind")

			s.globalMetrics.MissedRuns.WithLabelValues(name, "behind").Inc()
		}

		schedule.mu.RUnlock()
	}
}

// UpdateSchedule updates a collector's schedule
func (s *Scheduler) UpdateSchedule(name string, interval, timeout time.Duration) error {
	s.mu.RLock()
	schedule, exists := s.schedules[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schedule for collector %s not found", name)
	}

	schedule.mu.Lock()
	defer schedule.mu.Unlock()

	// Update schedule
	schedule.Interval = interval
	schedule.Timeout = timeout

	// Cancel current timer
	if schedule.timer != nil {
		schedule.timer.Stop()
	}

	// Update orchestrator
	s.orchestrator.SetCollectorInterval(name, interval)

	// Restart schedule if enabled
	if schedule.enabled {
		schedule.NextRun = time.Now().Add(interval)
		schedule.timer = time.AfterFunc(interval, func() {
			s.runScheduledCollection(name, schedule)
		})
	}

	s.logger.WithFields(logrus.Fields{
		"collector": name,
		"interval":  interval,
		"timeout":   timeout,
	}).Info("Updated collection schedule")

	return nil
}

// EnableSchedule enables a collector's schedule
func (s *Scheduler) EnableSchedule(name string) error {
	s.mu.RLock()
	schedule, exists := s.schedules[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schedule for collector %s not found", name)
	}

	schedule.mu.Lock()
	defer schedule.mu.Unlock()

	if schedule.enabled {
		return nil // Already enabled
	}

	schedule.enabled = true

	// Start schedule
	schedule.NextRun = time.Now().Add(schedule.Interval)
	schedule.timer = time.AfterFunc(schedule.Interval, func() {
		s.runScheduledCollection(name, schedule)
	})

	s.logger.WithField("collector", name).Info("Enabled collection schedule")

	return nil
}

// DisableSchedule disables a collector's schedule
func (s *Scheduler) DisableSchedule(name string) error {
	s.mu.RLock()
	schedule, exists := s.schedules[name]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schedule for collector %s not found", name)
	}

	schedule.mu.Lock()
	defer schedule.mu.Unlock()

	if !schedule.enabled {
		return nil // Already disabled
	}

	schedule.enabled = false

	// Stop timer
	if schedule.timer != nil {
		schedule.timer.Stop()
		schedule.timer = nil
	}

	s.logger.WithField("collector", name).Info("Disabled collection schedule")

	return nil
}

// GetScheduleStats returns statistics for all schedules
func (s *Scheduler) GetScheduleStats() map[string]ScheduleStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]ScheduleStats)

	for name, schedule := range s.schedules {
		schedule.mu.RLock()
		stats[name] = ScheduleStats{
			CollectorName: schedule.CollectorName,
			Interval:      schedule.Interval,
			Timeout:       schedule.Timeout,
			Enabled:       schedule.enabled,
			LastRun:       schedule.LastRun,
			NextRun:       schedule.NextRun,
			RunCount:      schedule.RunCount,
			ErrorCount:    schedule.ErrorCount,
			ErrorRate:     float64(schedule.ErrorCount) / float64(max(schedule.RunCount, 1)),
		}
		schedule.mu.RUnlock()
	}

	return stats
}

// ScheduleStats contains statistics for a schedule
type ScheduleStats struct {
	CollectorName string
	Interval      time.Duration
	Timeout       time.Duration
	Enabled       bool
	LastRun       time.Time
	NextRun       time.Time
	RunCount      int64
	ErrorCount    int64
	ErrorRate     float64
}
