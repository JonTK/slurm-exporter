package adaptive

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ActivityScore represents the cluster activity score for adaptation
type ActivityScore struct {
	Score       float64   `json:"score"`
	Timestamp   time.Time `json:"timestamp"`
	JobCount    int       `json:"job_count"`
	NodeCount   int       `json:"node_count"`
	ChangeRate  float64   `json:"change_rate"`
	Description string    `json:"description"`
}

// CollectorScheduler manages adaptive collection intervals based on cluster activity
type CollectorScheduler struct {
	config  config.AdaptiveCollectionConfig
	logger  *logrus.Logger
	metrics *SchedulerMetrics
	mu      sync.RWMutex

	// Activity tracking
	activityHistory []ActivityScore
	lastCollection  map[string]time.Time
	lastMetrics     map[string]interface{}

	// Collector intervals
	collectorIntervals map[string]time.Duration
	defaultInterval    time.Duration

	// Internal state
	enabled   bool
	startTime time.Time
}

// SchedulerMetrics holds Prometheus metrics for the adaptive scheduler
type SchedulerMetrics struct {
	activityScore       *prometheus.GaugeVec
	adaptedInterval     *prometheus.GaugeVec
	intervalAdjustments *prometheus.CounterVec
	collectionDelay     *prometheus.HistogramVec
	adaptationEvents    *prometheus.CounterVec
}

// NewCollectorScheduler creates a new adaptive collection scheduler
func NewCollectorScheduler(cfg config.AdaptiveCollectionConfig, defaultInterval time.Duration, logger *logrus.Logger) (*CollectorScheduler, error) {
	if !cfg.Enabled {
		logger.Info("Adaptive collection scheduler disabled")
		return &CollectorScheduler{
			config:             cfg,
			logger:             logger,
			enabled:            false,
			collectorIntervals: make(map[string]time.Duration),
			defaultInterval:    defaultInterval,
		}, nil
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid adaptive collection config: %w", err)
	}

	metrics := &SchedulerMetrics{
		activityScore: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_scheduler_activity_score",
				Help: "Current cluster activity score used for adaptive collection",
			},
			[]string{"collector"},
		),
		adaptedInterval: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "slurm_exporter_scheduler_interval_seconds",
				Help: "Current adaptive collection interval in seconds",
			},
			[]string{"collector"},
		),
		intervalAdjustments: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_scheduler_adjustments_total",
				Help: "Total number of interval adjustments made",
			},
			[]string{"collector", "direction"},
		),
		collectionDelay: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_exporter_scheduler_collection_delay_seconds",
				Help:    "Delay between scheduled and actual collection times",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"collector"},
		),
		adaptationEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_exporter_scheduler_events_total",
				Help: "Total number of adaptation events by type",
			},
			[]string{"event_type"},
		),
	}

	scheduler := &CollectorScheduler{
		config:             cfg,
		logger:             logger,
		metrics:            metrics,
		enabled:            true,
		activityHistory:    make([]ActivityScore, 0, 100), // Keep last 100 scores
		lastCollection:     make(map[string]time.Time),
		lastMetrics:        make(map[string]interface{}),
		collectorIntervals: make(map[string]time.Duration),
		defaultInterval:    defaultInterval,
		startTime:          time.Now(),
	}

	logger.WithFields(logrus.Fields{
		"min_interval":  cfg.MinInterval,
		"max_interval":  cfg.MaxInterval,
		"base_interval": cfg.BaseInterval,
		"score_window":  cfg.ScoreWindow,
	}).Info("Adaptive collection scheduler initialized")

	return scheduler, nil
}

// validateConfig validates the adaptive collection configuration
func validateConfig(cfg config.AdaptiveCollectionConfig) error {
	if cfg.MinInterval <= 0 {
		return fmt.Errorf("min_interval must be positive")
	}
	if cfg.MaxInterval <= 0 {
		return fmt.Errorf("max_interval must be positive")
	}
	if cfg.BaseInterval <= 0 {
		return fmt.Errorf("base_interval must be positive")
	}
	if cfg.MinInterval >= cfg.MaxInterval {
		return fmt.Errorf("min_interval must be less than max_interval")
	}
	if cfg.BaseInterval < cfg.MinInterval || cfg.BaseInterval > cfg.MaxInterval {
		return fmt.Errorf("base_interval must be between min_interval and max_interval")
	}
	if cfg.ScoreWindow <= 0 {
		return fmt.Errorf("score_window must be positive")
	}
	return nil
}

// RegisterMetrics registers the scheduler metrics with Prometheus
func (s *CollectorScheduler) RegisterMetrics(registry prometheus.Registerer) error {
	if !s.enabled {
		s.logger.Debug("Adaptive scheduler disabled, not registering metrics")
		return nil
	}

	if s.metrics == nil {
		return fmt.Errorf("metrics not initialized")
	}

	collectors := []prometheus.Collector{
		s.metrics.activityScore,
		s.metrics.adaptedInterval,
		s.metrics.intervalAdjustments,
		s.metrics.collectionDelay,
		s.metrics.adaptationEvents,
	}

	s.logger.WithField("count", len(collectors)).Debug("Registering adaptive scheduler metrics")

	for i, collector := range collectors {
		if err := registry.Register(collector); err != nil {
			return fmt.Errorf("failed to register adaptive scheduler metric %d: %w", i, err)
		}
	}

	s.logger.Info("Successfully registered adaptive scheduler metrics")
	return nil
}

// GetCollectionInterval returns the adaptive interval for a collector
func (s *CollectorScheduler) GetCollectionInterval(collector string) time.Duration {
	if !s.enabled {
		return s.defaultInterval
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if interval, exists := s.collectorIntervals[collector]; exists {
		return interval
	}

	// Use base interval as default for new collectors
	return s.config.BaseInterval
}

// UpdateActivity updates the cluster activity score and adjusts collection intervals
func (s *CollectorScheduler) UpdateActivity(ctx context.Context, jobCount, nodeCount int, changeData map[string]interface{}) {
	if !s.enabled {
		return
	}

	// Calculate activity score
	score := s.calculateActivityScore(jobCount, nodeCount, changeData)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to history
	activityScore := ActivityScore{
		Score:       score.Score,
		Timestamp:   time.Now(),
		JobCount:    jobCount,
		NodeCount:   nodeCount,
		ChangeRate:  score.ChangeRate,
		Description: score.Description,
	}

	s.activityHistory = append(s.activityHistory, activityScore)

	// Keep only recent history within score window
	cutoff := time.Now().Add(-s.config.ScoreWindow)
	var filtered []ActivityScore
	for _, entry := range s.activityHistory {
		if entry.Timestamp.After(cutoff) {
			filtered = append(filtered, entry)
		}
	}
	s.activityHistory = filtered

	// Update metrics
	s.metrics.activityScore.WithLabelValues("global").Set(score.Score)

	// Adapt collection intervals for all collectors
	s.adaptIntervals(score.Score)

	s.logger.WithFields(logrus.Fields{
		"activity_score": score.Score,
		"job_count":      jobCount,
		"node_count":     nodeCount,
		"change_rate":    score.ChangeRate,
		"description":    score.Description,
	}).Debug("Updated cluster activity score")
}

// ActivityScoreResult holds the calculated activity score and metadata
type ActivityScoreResult struct {
	Score       float64
	ChangeRate  float64
	Description string
}

// calculateActivityScore computes an activity score from 0.0 (low) to 1.0 (high)
func (s *CollectorScheduler) calculateActivityScore(jobCount, nodeCount int, changeData map[string]interface{}) ActivityScoreResult {
	var score float64
	var changeRate float64
	var factors []string

	// Factor 1: Absolute cluster size (normalized)
	clusterSize := float64(jobCount + nodeCount)
	sizeScore := math.Min(clusterSize/1000.0, 1.0) // Normalize to 1000 jobs+nodes = 1.0
	score += sizeScore * 0.3                       // 30% weight
	if sizeScore > 0.5 {
		factors = append(factors, "large-cluster")
	}

	// Factor 2: Job density (jobs per node)
	var densityScore float64
	if nodeCount > 0 {
		density := float64(jobCount) / float64(nodeCount)
		densityScore = math.Min(density/10.0, 1.0) // Normalize to 10 jobs/node = 1.0
		score += densityScore * 0.2                // 20% weight
		if densityScore > 0.7 {
			factors = append(factors, "high-density")
		}
	}

	// Factor 3: Change rate from previous collection
	if len(s.lastMetrics) > 0 {
		changeRate = s.calculateChangeRate(changeData)
		changeScore := math.Min(changeRate, 1.0)
		score += changeScore * 0.4 // 40% weight
		if changeRate > 0.5 {
			factors = append(factors, "high-change-rate")
		}
	} else {
		// First collection - assume moderate activity
		score += 0.5 * 0.4
		factors = append(factors, "initial")
	}

	// Factor 4: Time-based adjustment (more active during business hours)
	timeScore := s.getTimeBasedScore()
	score += timeScore * 0.1 // 10% weight

	// Store current metrics for next calculation
	s.lastMetrics = changeData

	score = math.Max(0.0, math.Min(1.0, score))

	description := "normal"
	if len(factors) > 0 {
		description = fmt.Sprintf("%v", factors)
	}

	return ActivityScoreResult{
		Score:       score,
		ChangeRate:  changeRate,
		Description: description,
	}
}

// calculateChangeRate computes the rate of change from previous metrics
func (s *CollectorScheduler) calculateChangeRate(current map[string]interface{}) float64 {
	if len(s.lastMetrics) == 0 {
		return 0.5 // Assume moderate change for first measurement
	}

	var totalChange float64
	var comparisons int

	for key, currentVal := range current {
		if lastVal, exists := s.lastMetrics[key]; exists {
			change := s.compareValues(lastVal, currentVal)
			totalChange += change
			comparisons++
		}
	}

	if comparisons == 0 {
		return 0.5
	}

	return totalChange / float64(comparisons)
}

// compareValues compares two values and returns a change rate (0.0 to 1.0)
func (s *CollectorScheduler) compareValues(old, new interface{}) float64 {
	// Try to convert to numbers for comparison
	oldFloat, oldOk := s.toFloat64(old)
	newFloat, newOk := s.toFloat64(new)

	if oldOk && newOk {
		if oldFloat == 0 {
			if newFloat == 0 {
				return 0.0 // No change
			}
			return 1.0 // New value appeared
		}
		change := math.Abs(newFloat-oldFloat) / oldFloat
		return math.Min(change, 1.0)
	}

	// For non-numeric values, check if they're different
	if fmt.Sprintf("%v", old) != fmt.Sprintf("%v", new) {
		return 1.0
	}
	return 0.0
}

// toFloat64 attempts to convert a value to float64
func (s *CollectorScheduler) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	default:
		return 0, false
	}
}

// getTimeBasedScore returns a score based on time of day (business hours = higher activity)
func (s *CollectorScheduler) getTimeBasedScore() float64 {
	now := time.Now()
	hour := now.Hour()

	// Assume business hours 8 AM to 6 PM are higher activity
	if hour >= 8 && hour <= 18 {
		return 0.7
	}
	// Evening hours still some activity
	if hour >= 19 && hour <= 22 {
		return 0.4
	}
	// Night/early morning - low activity
	return 0.2
}

// adaptIntervals adjusts collection intervals based on activity score
func (s *CollectorScheduler) adaptIntervals(activityScore float64) {
	// Calculate target interval using inverse relationship
	// High activity (score → 1.0) = shorter intervals (→ min_interval)
	// Low activity (score → 0.0) = longer intervals (→ max_interval)

	minInterval := s.config.MinInterval.Seconds()
	maxInterval := s.config.MaxInterval.Seconds()

	// Use exponential scaling for smoother adaptation
	targetInterval := maxInterval - (maxInterval-minInterval)*math.Pow(activityScore, 0.5)
	targetDuration := time.Duration(targetInterval * float64(time.Second))

	// Apply the same interval to all collectors
	// In a more sophisticated implementation, different collectors could have different adaptations
	for collector := range s.collectorIntervals {
		oldInterval := s.collectorIntervals[collector]
		s.collectorIntervals[collector] = targetDuration

		// Record metrics
		s.metrics.adaptedInterval.WithLabelValues(collector).Set(targetDuration.Seconds())

		// Record adjustment direction
		if targetDuration < oldInterval {
			s.metrics.intervalAdjustments.WithLabelValues(collector, "decrease").Inc()
		} else if targetDuration > oldInterval {
			s.metrics.intervalAdjustments.WithLabelValues(collector, "increase").Inc()
		}
	}

	// Update global metrics for overall tracking
	s.metrics.adaptedInterval.WithLabelValues("global").Set(targetDuration.Seconds())

	s.metrics.adaptationEvents.WithLabelValues("interval_update").Inc()
}

// RegisterCollector registers a collector with the scheduler
func (s *CollectorScheduler) RegisterCollector(name string) {
	if !s.enabled {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Start with base interval
	s.collectorIntervals[name] = s.config.BaseInterval
	s.lastCollection[name] = time.Now()

	s.metrics.adaptedInterval.WithLabelValues(name).Set(s.config.BaseInterval.Seconds())

	s.logger.WithFields(logrus.Fields{
		"collector": name,
		"interval":  s.config.BaseInterval,
	}).Debug("Registered collector with adaptive scheduler")
}

// RecordCollection records that a collection occurred for timing analysis
func (s *CollectorScheduler) RecordCollection(collector string, scheduledTime, actualTime time.Time) {
	if !s.enabled {
		return
	}

	s.mu.Lock()
	s.lastCollection[collector] = actualTime
	s.mu.Unlock()

	// Record collection delay
	delay := actualTime.Sub(scheduledTime)
	if delay > 0 {
		s.metrics.collectionDelay.WithLabelValues(collector).Observe(delay.Seconds())
	}
}

// GetActivityHistory returns recent activity history
func (s *CollectorScheduler) GetActivityHistory() []ActivityScore {
	if !s.enabled {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]ActivityScore, len(s.activityHistory))
	copy(history, s.activityHistory)
	return history
}

// GetCurrentScore returns the most recent activity score
func (s *CollectorScheduler) GetCurrentScore() float64 {
	if !s.enabled {
		return 0.5
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.activityHistory) == 0 {
		return 0.5 // Default moderate score
	}

	return s.activityHistory[len(s.activityHistory)-1].Score
}

// IsEnabled returns whether the adaptive scheduler is enabled
func (s *CollectorScheduler) IsEnabled() bool {
	return s.enabled
}

// GetConfig returns the current configuration
func (s *CollectorScheduler) GetConfig() config.AdaptiveCollectionConfig {
	return s.config
}

// GetStats returns statistics about the scheduler
func (s *CollectorScheduler) GetStats() map[string]interface{} {
	if !s.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":               true,
		"uptime_seconds":        time.Since(s.startTime).Seconds(),
		"activity_entries":      len(s.activityHistory),
		"registered_collectors": len(s.collectorIntervals),
		"current_score":         s.GetCurrentScore(),
		"config": map[string]interface{}{
			"min_interval":  s.config.MinInterval.String(),
			"max_interval":  s.config.MaxInterval.String(),
			"base_interval": s.config.BaseInterval.String(),
			"score_window":  s.config.ScoreWindow.String(),
		},
	}

	// Add current intervals
	intervals := make(map[string]string)
	for collector, interval := range s.collectorIntervals {
		intervals[collector] = interval.String()
	}
	stats["current_intervals"] = intervals

	return stats
}
