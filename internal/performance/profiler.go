// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package performance

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Profiler provides profiling capabilities for collectors
type Profiler struct {
	config   ProfilerConfig
	logger   *logrus.Entry
	storage  ProfileStorage
	metrics  *profilerMetrics
	mu       sync.RWMutex
	profiles map[string]*CollectorProfile
	enabled  bool
}

// ProfilerConfig holds profiler configuration
type ProfilerConfig struct {
	Enabled              bool                    `yaml:"enabled"`
	CPUProfileRate       int                     `yaml:"cpu_profile_rate"`       // Hz
	MemProfileRate       int                     `yaml:"memory_profile_rate"`    // bytes
	BlockProfileRate     int                     `yaml:"block_profile_rate"`     // nanoseconds
	MutexProfileFraction int                     `yaml:"mutex_profile_fraction"` // 1/n
	AutoProfile          AutoProfileConfig       `yaml:"auto_profile"`
	Storage              ProfileStorageConfig    `yaml:"storage"`
	ContinuousProfiling  ContinuousProfileConfig `yaml:"continuous_profiling"`
}

// AutoProfileConfig holds automatic profiling triggers
type AutoProfileConfig struct {
	Enabled                 bool          `yaml:"enabled"`
	DurationThreshold       time.Duration `yaml:"duration_threshold"`
	MemoryThreshold         int64         `yaml:"memory_threshold"`     // bytes
	ErrorRateThreshold      float64       `yaml:"error_rate_threshold"` // percentage
	CPUUsageThreshold       float64       `yaml:"cpu_usage_threshold"`  // percentage
	ProfileOnSlowCollection bool          `yaml:"profile_on_slow_collection"`
}

// ProfileStorageConfig holds profile storage configuration
type ProfileStorageConfig struct {
	Type      string        `yaml:"type"`      // "memory", "file", "s3"
	Path      string        `yaml:"path"`      // base path for profiles
	MaxSize   int64         `yaml:"max_size"`  // max total size in bytes
	Retention time.Duration `yaml:"retention"` // how long to keep profiles
}

// ContinuousProfileConfig holds continuous profiling configuration
type ContinuousProfileConfig struct {
	Enabled          bool          `yaml:"enabled"`
	Interval         time.Duration `yaml:"interval"`
	CPUDuration      time.Duration `yaml:"cpu_duration"`
	IncludeHeap      bool          `yaml:"include_heap"`
	IncludeGoroutine bool          `yaml:"include_goroutine"`
}

// CollectorProfile represents a profile for a specific collector
type CollectorProfile struct {
	CollectorName    string
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	CPUProfile       *bytes.Buffer
	HeapProfile      *bytes.Buffer
	GoroutineProfile *bytes.Buffer
	BlockProfile     *bytes.Buffer
	MutexProfile     *bytes.Buffer
	TraceData        *bytes.Buffer
	Phases           map[string]*ProfilePhase
	Metadata         map[string]interface{}
}

// ProfilePhase represents a phase within a collection
type ProfilePhase struct {
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Allocations int64
	CPUNanos    int64
}

// OperationProfile tracks a single operation
type OperationProfile struct {
	profiler      *Profiler
	collectorName string
	profile       *CollectorProfile
	currentPhase  *ProfilePhase
	startMem      runtime.MemStats
	startTime     time.Time
	// TODO: Unused field - preserved for future CPU profiling
	// cpuProfile    *pprof.Profile
	traceCtx  context.Context
	traceTask *trace.Task
}

// profilerMetrics holds profiler metrics
type profilerMetrics struct {
	profilesCreated     *prometheus.CounterVec
	profilesSaved       *prometheus.CounterVec
	profileSize         *prometheus.HistogramVec
	profileDuration     *prometheus.HistogramVec
	autoProfileTriggers *prometheus.CounterVec
	storageUsage        prometheus.Gauge
	activeProfiles      prometheus.Gauge
}

// NewProfiler creates a new profiler
func NewProfiler(config ProfilerConfig, logger *logrus.Entry) (*Profiler, error) {
	if !config.Enabled {
		return &Profiler{enabled: false}, nil
	}

	// Set runtime profiling rates
	if config.CPUProfileRate > 0 {
		runtime.SetCPUProfileRate(config.CPUProfileRate)
	}
	if config.MemProfileRate > 0 {
		runtime.MemProfileRate = config.MemProfileRate
	}
	if config.BlockProfileRate > 0 {
		runtime.SetBlockProfileRate(config.BlockProfileRate)
	}
	if config.MutexProfileFraction > 0 {
		runtime.SetMutexProfileFraction(config.MutexProfileFraction)
	}

	storage, err := NewProfileStorage(config.Storage)
	if err != nil {
		return nil, fmt.Errorf("creating profile storage: %w", err)
	}

	p := &Profiler{
		config:   config,
		logger:   logger,
		storage:  storage,
		metrics:  createProfilerMetrics(),
		profiles: make(map[string]*CollectorProfile),
		enabled:  true,
	}

	// Start continuous profiling if enabled
	if config.ContinuousProfiling.Enabled {
		go p.continuousProfiling()
	}

	return p, nil
}

// StartOperation starts profiling an operation
func (p *Profiler) StartOperation(collectorName string) *OperationProfile {
	if !p.enabled {
		return &OperationProfile{profiler: p}
	}

	profile := &CollectorProfile{
		CollectorName: collectorName,
		StartTime:     time.Now(),
		Phases:        make(map[string]*ProfilePhase),
		Metadata:      make(map[string]interface{}),
	}

	// Start CPU profiling
	var cpuBuf bytes.Buffer
	if p.config.AutoProfile.ProfileOnSlowCollection {
		profile.CPUProfile = &cpuBuf
		if err := pprof.StartCPUProfile(&cpuBuf); err != nil {
			p.logger.WithError(err).Warn("Failed to start CPU profiling")
		}
	}

	// Start trace if enabled
	var traceBuf bytes.Buffer
	ctx, task := trace.NewTask(context.Background(), collectorName)
	profile.TraceData = &traceBuf

	op := &OperationProfile{
		profiler:      p,
		collectorName: collectorName,
		profile:       profile,
		startTime:     time.Now(),
		traceCtx:      ctx,
		traceTask:     task,
	}

	// Get initial memory stats
	runtime.ReadMemStats(&op.startMem)

	p.mu.Lock()
	p.profiles[collectorName] = profile
	p.mu.Unlock()

	p.metrics.activeProfiles.Inc()
	p.metrics.profilesCreated.WithLabelValues(collectorName).Inc()

	return op
}

// Phase starts a new phase in the operation
func (op *OperationProfile) Phase(name string) {
	if !op.profiler.enabled || op.profile == nil {
		return
	}

	// End current phase if any
	if op.currentPhase != nil {
		op.currentPhase.EndTime = time.Now()
		op.currentPhase.Duration = op.currentPhase.EndTime.Sub(op.currentPhase.StartTime)
	}

	// Start new phase
	phase := &ProfilePhase{
		Name:      name,
		StartTime: time.Now(),
	}

	trace.WithRegion(op.traceCtx, name, func() {
		op.currentPhase = phase
		op.profile.Phases[name] = phase
	})
}

// PhaseEnd ends the current phase
func (op *OperationProfile) PhaseEnd(name string) {
	if !op.profiler.enabled || op.profile == nil {
		return
	}

	if phase, exists := op.profile.Phases[name]; exists {
		phase.EndTime = time.Now()
		phase.Duration = phase.EndTime.Sub(phase.StartTime)

		// Calculate resource usage for this phase
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		phase.Allocations = int64(memStats.TotalAlloc - op.startMem.TotalAlloc)
	}

	if op.currentPhase != nil && op.currentPhase.Name == name {
		op.currentPhase = nil
	}
}

// RecordAllocation records memory allocation
func (op *OperationProfile) RecordAllocation(obj interface{}) {
	if !op.profiler.enabled || op.currentPhase == nil {
		return
	}

	// This is a simplified version - in production you'd use reflection
	// or type assertions to get actual size
	op.currentPhase.Allocations += 1024 // Placeholder
}

// Stop stops the operation profiling
func (op *OperationProfile) Stop() {
	if !op.profiler.enabled || op.profile == nil {
		return
	}

	op.profile.EndTime = time.Now()
	op.profile.Duration = op.profile.EndTime.Sub(op.profile.StartTime)

	// Stop CPU profiling
	if op.profile.CPUProfile != nil {
		pprof.StopCPUProfile()
	}

	// End trace task
	if op.traceTask != nil {
		op.traceTask.End()
	}

	// Collect final profiles
	op.collectFinalProfiles()

	// Check if we should save the profile
	if op.shouldSaveProfile() {
		_ = op.Save()
	}

	op.profiler.mu.Lock()
	delete(op.profiler.profiles, op.collectorName)
	op.profiler.mu.Unlock()

	op.profiler.metrics.activeProfiles.Dec()
	op.profiler.metrics.profileDuration.WithLabelValues(op.collectorName).Observe(op.profile.Duration.Seconds())
}

// collectFinalProfiles collects all profile data
func (op *OperationProfile) collectFinalProfiles() {
	if !op.profiler.enabled {
		return
	}

	// Heap profile
	if op.profile.HeapProfile == nil {
		op.profile.HeapProfile = &bytes.Buffer{}
	}
	_ = pprof.WriteHeapProfile(op.profile.HeapProfile)

	// Goroutine profile
	if op.profile.GoroutineProfile == nil {
		op.profile.GoroutineProfile = &bytes.Buffer{}
	}
	_ = pprof.Lookup("goroutine").WriteTo(op.profile.GoroutineProfile, 0)

	// Block profile
	if op.profiler.config.BlockProfileRate > 0 {
		op.profile.BlockProfile = &bytes.Buffer{}
		_ = pprof.Lookup("block").WriteTo(op.profile.BlockProfile, 0)
	}

	// Mutex profile
	if op.profiler.config.MutexProfileFraction > 0 {
		op.profile.MutexProfile = &bytes.Buffer{}
		_ = pprof.Lookup("mutex").WriteTo(op.profile.MutexProfile, 0)
	}
}

// shouldSaveProfile determines if the profile should be saved
func (op *OperationProfile) shouldSaveProfile() bool {
	if !op.profiler.enabled || !op.profiler.config.AutoProfile.Enabled {
		return false
	}

	cfg := op.profiler.config.AutoProfile

	// Check duration threshold
	if cfg.DurationThreshold > 0 && op.profile.Duration >= cfg.DurationThreshold {
		op.profiler.metrics.autoProfileTriggers.WithLabelValues(op.collectorName, "duration").Inc()
		return true
	}

	// Check memory threshold
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memUsed := int64(memStats.Alloc - op.startMem.Alloc)
	if cfg.MemoryThreshold > 0 && memUsed >= cfg.MemoryThreshold {
		op.profiler.metrics.autoProfileTriggers.WithLabelValues(op.collectorName, "memory").Inc()
		return true
	}

	return false
}

// Save saves the profile
func (op *OperationProfile) Save() error {
	if !op.profiler.enabled || op.profile == nil {
		return nil
	}

	err := op.profiler.storage.Save(op.profile)
	if err != nil {
		op.profiler.logger.WithError(err).Error("Failed to save profile")
		return err
	}

	op.profiler.metrics.profilesSaved.WithLabelValues(op.collectorName).Inc()

	// Calculate total size
	totalSize := 0
	if op.profile.CPUProfile != nil {
		totalSize += op.profile.CPUProfile.Len()
	}
	if op.profile.HeapProfile != nil {
		totalSize += op.profile.HeapProfile.Len()
	}
	if op.profile.GoroutineProfile != nil {
		totalSize += op.profile.GoroutineProfile.Len()
	}

	op.profiler.metrics.profileSize.WithLabelValues(op.collectorName).Observe(float64(totalSize))

	return nil
}

// Duration returns the operation duration
func (op *OperationProfile) Duration() time.Duration {
	if op.profile == nil {
		return 0
	}
	return op.profile.Duration
}

// SaveCPUProfile saves a CPU profile
func (op *OperationProfile) SaveCPUProfile() {
	if !op.profiler.enabled {
		return
	}
	// CPU profile is already being collected if enabled
}

// SaveHeapProfile saves a heap profile
func (op *OperationProfile) SaveHeapProfile() {
	if !op.profiler.enabled || op.profile == nil {
		return
	}

	if op.profile.HeapProfile == nil {
		op.profile.HeapProfile = &bytes.Buffer{}
	}
	_ = pprof.WriteHeapProfile(op.profile.HeapProfile)
}

// GenerateReport generates a profile report
func (op *OperationProfile) GenerateReport() string {
	if !op.profiler.enabled || op.profile == nil {
		return ""
	}

	report := fmt.Sprintf("Profile Report for %s\n", op.collectorName)
	report += fmt.Sprintf("Duration: %v\n", op.profile.Duration)
	report += fmt.Sprintf("Start: %v, End: %v\n", op.profile.StartTime, op.profile.EndTime)

	report += "\nPhases:\n"
	for name, phase := range op.profile.Phases {
		report += fmt.Sprintf("  %s: %v (Allocations: %d bytes)\n",
			name, phase.Duration, phase.Allocations)
	}

	return report
}

// continuousProfiling runs continuous profiling
func (p *Profiler) continuousProfiling() {
	ticker := time.NewTicker(p.config.ContinuousProfiling.Interval)
	defer ticker.Stop()

	for range ticker.C {
		p.collectContinuousProfile()
	}
}

// collectContinuousProfile collects a continuous profile
func (p *Profiler) collectContinuousProfile() {
	profile := &CollectorProfile{
		CollectorName: "continuous",
		StartTime:     time.Now(),
		Metadata: map[string]interface{}{
			"type": "continuous",
		},
	}

	// CPU profile
	if p.config.ContinuousProfiling.CPUDuration > 0 {
		cpuBuf := &bytes.Buffer{}
		profile.CPUProfile = cpuBuf

		if err := pprof.StartCPUProfile(cpuBuf); err != nil {
			p.logger.WithError(err).Warn("Failed to start CPU profiling")
		} else {
			time.Sleep(p.config.ContinuousProfiling.CPUDuration)
			pprof.StopCPUProfile()
		}
	}

	// Heap profile
	if p.config.ContinuousProfiling.IncludeHeap {
		profile.HeapProfile = &bytes.Buffer{}
		_ = pprof.WriteHeapProfile(profile.HeapProfile)
	}

	// Goroutine profile
	if p.config.ContinuousProfiling.IncludeGoroutine {
		profile.GoroutineProfile = &bytes.Buffer{}
		_ = pprof.Lookup("goroutine").WriteTo(profile.GoroutineProfile, 0)
	}

	profile.EndTime = time.Now()
	profile.Duration = profile.EndTime.Sub(profile.StartTime)

	// Save the profile
	if err := p.storage.Save(profile); err != nil {
		p.logger.WithError(err).Error("Failed to save continuous profile")
	}
}

// GetProfile returns a profile by name
func (p *Profiler) GetProfile(collectorName string) *CollectorProfile {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.profiles[collectorName]
}

// ListProfiles lists all stored profiles
func (p *Profiler) ListProfiles() ([]*ProfileMetadata, error) {
	return p.storage.List()
}

// LoadProfile loads a profile from storage
func (p *Profiler) LoadProfile(id string) (*CollectorProfile, error) {
	return p.storage.Load(id)
}

// GetStats returns profiler statistics
func (p *Profiler) GetStats() map[string]interface{} {
	p.mu.RLock()
	activeCount := len(p.profiles)
	p.mu.RUnlock()

	storageStats := p.storage.GetStats()

	return map[string]interface{}{
		"enabled":         p.enabled,
		"active_profiles": activeCount,
		"config":          p.config,
		"storage":         storageStats,
	}
}

// createProfilerMetrics creates profiler metrics
func createProfilerMetrics() *profilerMetrics {
	return &profilerMetrics{
		profilesCreated: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_profiler_profiles_created_total",
				Help: "Total number of profiles created",
			},
			[]string{"collector"},
		),
		profilesSaved: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_profiler_profiles_saved_total",
				Help: "Total number of profiles saved",
			},
			[]string{"collector"},
		),
		profileSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_profiler_profile_size_bytes",
				Help:    "Size of saved profiles in bytes",
				Buckets: []float64{1024, 10240, 102400, 1048576, 10485760},
			},
			[]string{"collector"},
		),
		profileDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "slurm_profiler_profile_duration_seconds",
				Help:    "Duration of profiled operations",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"collector"},
		),
		autoProfileTriggers: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "slurm_profiler_auto_triggers_total",
				Help: "Total number of automatic profile triggers",
			},
			[]string{"collector", "trigger"},
		),
		storageUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_profiler_storage_usage_bytes",
				Help: "Current profile storage usage in bytes",
			},
		),
		activeProfiles: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "slurm_profiler_active_profiles",
				Help: "Number of currently active profiles",
			},
		),
	}
}

// RegisterMetrics registers profiler metrics
func (p *Profiler) RegisterMetrics(reg prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		p.metrics.profilesCreated,
		p.metrics.profilesSaved,
		p.metrics.profileSize,
		p.metrics.profileDuration,
		p.metrics.autoProfileTriggers,
		p.metrics.storageUsage,
		p.metrics.activeProfiles,
	}

	for _, collector := range collectors {
		if err := reg.Register(collector); err != nil {
			return err
		}
	}

	return nil
}
