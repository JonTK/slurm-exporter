// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package debug

import (
	"encoding/json"
	"html/template"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jontk/slurm-exporter/internal/adaptive"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/jontk/slurm-exporter/internal/filtering"
	"github.com/jontk/slurm-exporter/internal/health"
	"github.com/jontk/slurm-exporter/internal/tracing"
	"github.com/sirupsen/logrus"
)

// Format type constants
const (
	FormatJSON = "json"
)

// DebugHandler manages debug endpoints for troubleshooting
type DebugHandler struct {
	config config.DebugConfig
	logger *logrus.Logger
	mu     sync.RWMutex

	// Component references for inspection
	scheduler  SchedulerDebugger
	filter     FilterDebugger
	tracer     TracerDebugger
	health     HealthDebugger
	collectors CollectorDebugger
	cache      CacheDebugger

	// Debug state
	enabled      bool
	startTime    time.Time
	requestCount int64

	// Templates for HTML output
	templates *template.Template
}

// Component interfaces for debugging
type SchedulerDebugger interface {
	GetStats() map[string]interface{}
	GetActivityHistory() []adaptive.ActivityScore
	IsEnabled() bool
}

type FilterDebugger interface {
	GetStats() map[string]interface{}
	GetPatterns() map[string]filtering.MetricPattern
	IsEnabled() bool
}

type TracerDebugger interface {
	GetStats() tracing.TracingStats
	IsEnabled() bool
	GetConfig() config.TracingConfig
}

type HealthDebugger interface {
	GetStats() map[string]interface{}
	GetChecks() map[string]health.Check
}

type CollectorDebugger interface {
	GetStates() map[string]CollectorState
	GetMetrics() map[string]CollectorMetrics
}

type CacheDebugger interface {
	GetStats() map[string]interface{}
	GetMetrics() interface{} // Returns cache-specific metrics
}

// CollectorState represents the current state of a collector
type CollectorState struct {
	Name              string        `json:"name"`
	Enabled           bool          `json:"enabled"`
	LastCollection    time.Time     `json:"last_collection"`
	LastDuration      time.Duration `json:"last_duration"`
	LastError         error         `json:"last_error,omitempty"`
	ConsecutiveErrors int           `json:"consecutive_errors"`
	TotalCollections  int64         `json:"total_collections"`
	TotalErrors       int64         `json:"total_errors"`
	CurrentInterval   time.Duration `json:"current_interval"`
}

// CollectorMetrics represents metrics for a collector
type CollectorMetrics struct {
	MetricsCollected int64         `json:"metrics_collected"`
	MetricsFiltered  int64         `json:"metrics_filtered"`
	AvgDuration      time.Duration `json:"avg_duration"`
	SuccessRate      float64       `json:"success_rate"`
	LastSuccessful   time.Time     `json:"last_successful"`
}

// NewDebugHandler creates a new debug endpoint handler
func NewDebugHandler(cfg config.DebugConfig, logger *logrus.Logger) *DebugHandler {
	if !cfg.Enabled {
		logger.Info("Debug endpoints disabled")
		return &DebugHandler{
			config:  cfg,
			logger:  logger,
			enabled: false,
		}
	}

	// Initialize templates
	templates := template.Must(template.New("debug").Parse(debugTemplates))

	handler := &DebugHandler{
		config:    cfg,
		logger:    logger,
		enabled:   true,
		startTime: time.Now(),
		templates: templates,
	}

	logger.WithFields(logrus.Fields{
		"endpoints":     cfg.EnabledEndpoints,
		"auth_required": cfg.RequireAuth,
	}).Info("Debug endpoints initialized")

	return handler
}

// RegisterComponents registers debug interfaces for various components
func (dh *DebugHandler) RegisterComponents(
	scheduler SchedulerDebugger,
	filter FilterDebugger,
	tracer TracerDebugger,
	health HealthDebugger,
	collectors CollectorDebugger,
	cache CacheDebugger,
) {
	if !dh.enabled {
		return
	}

	dh.mu.Lock()
	defer dh.mu.Unlock()

	dh.scheduler = scheduler
	dh.filter = filter
	dh.tracer = tracer
	dh.health = health
	dh.collectors = collectors
	dh.cache = cache

	dh.logger.Info("Debug components registered")
}

// RegisterRoutes registers debug endpoints with an HTTP mux
func (dh *DebugHandler) RegisterRoutes(mux *http.ServeMux) {
	if !dh.enabled {
		return
	}

	// Main debug index
	mux.HandleFunc("/debug", dh.withAuth(dh.handleIndex))
	mux.HandleFunc("/debug/", dh.withAuth(dh.handleIndex))

	// Individual component endpoints
	if dh.isEndpointEnabled("collectors") {
		mux.HandleFunc("/debug/collectors", dh.withAuth(dh.handleCollectors))
	}
	if dh.isEndpointEnabled("tracing") {
		mux.HandleFunc("/debug/tracing", dh.withAuth(dh.handleTracing))
	}
	if dh.isEndpointEnabled("patterns") {
		mux.HandleFunc("/debug/patterns", dh.withAuth(dh.handlePatterns))
	}
	if dh.isEndpointEnabled("scheduler") {
		mux.HandleFunc("/debug/scheduler", dh.withAuth(dh.handleScheduler))
	}
	if dh.isEndpointEnabled("health") {
		mux.HandleFunc("/debug/health", dh.withAuth(dh.handleHealth))
	}
	if dh.isEndpointEnabled("runtime") {
		mux.HandleFunc("/debug/runtime", dh.withAuth(dh.handleRuntime))
	}
	if dh.isEndpointEnabled("config") {
		mux.HandleFunc("/debug/config", dh.withAuth(dh.handleConfig))
	}
	if dh.isEndpointEnabled("cache") {
		mux.HandleFunc("/debug/cache", dh.withAuth(dh.handleCache))
	}

	dh.logger.WithField("endpoints", len(dh.config.EnabledEndpoints)).Info("Debug routes registered")
}

// withAuth wraps handlers with authentication if required
func (dh *DebugHandler) withAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dh.mu.Lock()
		dh.requestCount++
		dh.mu.Unlock()

		// Check authentication if required
		if dh.config.RequireAuth {
			username, password, ok := r.BasicAuth()
			if !ok || username != dh.config.Username || password != dh.config.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Debug Endpoints"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Log access (sanitize user input to prevent log injection)
		dh.logger.WithFields(logrus.Fields{
			"endpoint": sanitizeLogField(r.URL.Path),
			"method":   sanitizeLogField(r.Method),
			"remote":   sanitizeLogField(r.RemoteAddr),
		}).Debug("Debug endpoint accessed")

		handler(w, r)
	}
}

// isEndpointEnabled checks if a specific endpoint is enabled
func (dh *DebugHandler) isEndpointEnabled(endpoint string) bool {
	if len(dh.config.EnabledEndpoints) == 0 {
		return true // Default: all enabled
	}

	for _, enabled := range dh.config.EnabledEndpoints {
		if enabled == endpoint {
			return true
		}
	}
	return false
}

// handleIndex serves the main debug page
func (dh *DebugHandler) handleIndex(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":        "SLURM Exporter Debug",
		"StartTime":    dh.startTime,
		"Uptime":       time.Since(dh.startTime),
		"RequestCount": dh.requestCount,
		"Endpoints":    dh.getAvailableEndpoints(),
		"Components":   dh.getComponentStatus(),
	}

	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, "index", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render debug index")
		}
	}
}

// handleCollectors serves collector debug information
func (dh *DebugHandler) handleCollectors(w http.ResponseWriter, r *http.Request) {
	if dh.collectors == nil {
		http.Error(w, "Collectors debugger not available", http.StatusServiceUnavailable)
		return
	}

	states := dh.collectors.GetStates()
	metrics := dh.collectors.GetMetrics()

	data := map[string]interface{}{
		"Title":   "Collector Debug",
		"States":  states,
		"Metrics": metrics,
		"Summary": dh.getCollectorSummary(states, metrics),
	}

	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, "collectors", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render collectors debug")
		}
	}
}

// renderDebugData renders debug data as JSON or template based on format query parameter
func (dh *DebugHandler) renderDebugData(w http.ResponseWriter, r *http.Request, templateName string, data map[string]interface{}) {
	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, templateName, data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render " + templateName + " debug")
		}
	}
}

// handleTracing serves tracing debug information
func (dh *DebugHandler) handleTracing(w http.ResponseWriter, r *http.Request) {
	if dh.tracer == nil {
		http.Error(w, "Tracer debugger not available", http.StatusServiceUnavailable)
		return
	}

	data := map[string]interface{}{
		"Title":   "Tracing Debug",
		"Stats":   dh.tracer.GetStats(),
		"Config":  dh.tracer.GetConfig(),
		"Enabled": dh.tracer.IsEnabled(),
	}

	dh.renderDebugData(w, r, "tracing", data)
}

// handlePatterns serves smart filter patterns
func (dh *DebugHandler) handlePatterns(w http.ResponseWriter, r *http.Request) {
	if dh.filter == nil {
		http.Error(w, "Filter debugger not available", http.StatusServiceUnavailable)
		return
	}

	patterns := dh.filter.GetPatterns()
	filterAction := r.URL.Query().Get("action")
	minSamples, _ := strconv.Atoi(r.URL.Query().Get("min_samples"))
	sortBy := r.URL.Query().Get("sort")

	data := map[string]interface{}{
		"Title":    "Smart Filter Patterns",
		"Patterns": dh.sortPatterns(dh.filterPatterns(patterns, filterAction, minSamples), sortBy),
		"Stats":    dh.filter.GetStats(),
		"Enabled":  dh.filter.IsEnabled(),
		"Filter": map[string]interface{}{
			"Action":     filterAction,
			"MinSamples": minSamples,
			"SortBy":     sortBy,
		},
	}

	dh.renderDebugData(w, r, "patterns", data)
}

// handleScheduler serves adaptive scheduler debug information
func (dh *DebugHandler) handleScheduler(w http.ResponseWriter, r *http.Request) {
	if dh.scheduler == nil {
		http.Error(w, "Scheduler debugger not available", http.StatusServiceUnavailable)
		return
	}

	data := map[string]interface{}{
		"Title":   "Adaptive Scheduler Debug",
		"Stats":   dh.scheduler.GetStats(),
		"History": dh.scheduler.GetActivityHistory(),
		"Enabled": dh.scheduler.IsEnabled(),
	}

	dh.renderDebugData(w, r, "scheduler", data)
}

// handleHealth serves health check debug information
func (dh *DebugHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if dh.health == nil {
		http.Error(w, "Health debugger not available", http.StatusServiceUnavailable)
		return
	}

	data := map[string]interface{}{
		"Title":  "Health Debug",
		"Stats":  dh.health.GetStats(),
		"Checks": dh.health.GetChecks(),
	}

	dh.renderDebugData(w, r, "health", data)
}

// handleRuntime serves Go runtime debug information
func (dh *DebugHandler) handleRuntime(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	data := map[string]interface{}{
		"Title": "Runtime Debug",
		"Memory": map[string]interface{}{
			"Alloc":         memStats.Alloc,
			"TotalAlloc":    memStats.TotalAlloc,
			"Sys":           memStats.Sys,
			"NumGC":         memStats.NumGC,
			"GCCPUFraction": memStats.GCCPUFraction,
			"HeapAlloc":     memStats.HeapAlloc,
			"HeapInuse":     memStats.HeapInuse,
			"HeapIdle":      memStats.HeapIdle,
			"StackInuse":    memStats.StackInuse,
		},
		"Runtime": map[string]interface{}{
			"Version":      runtime.Version(),
			"NumGoroutine": runtime.NumGoroutine(),
			"NumCPU":       runtime.NumCPU(),
			"GOMAXPROCS":   runtime.GOMAXPROCS(0),
		},
	}

	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, "runtime", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render runtime debug")
		}
	}
}

// handleConfig serves configuration debug information
func (dh *DebugHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Note: Be careful not to expose sensitive configuration
	data := map[string]interface{}{
		"Title": "Configuration Debug",
		"Debug": dh.config,
		// Add other non-sensitive config sections as needed
	}

	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, "config", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render config debug")
		}
	}
}

// handleCache serves intelligent cache debug information
func (dh *DebugHandler) handleCache(w http.ResponseWriter, r *http.Request) {
	if dh.cache == nil {
		http.Error(w, "Cache debugger not available", http.StatusServiceUnavailable)
		return
	}

	stats := dh.cache.GetStats()
	metrics := dh.cache.GetMetrics()

	data := map[string]interface{}{
		"Title":   "Intelligent Cache Debug",
		"Stats":   stats,
		"Metrics": metrics,
	}

	format := r.URL.Query().Get("format")
	if format == FormatJSON {
		dh.writeJSON(w, data)
	} else {
		if err := dh.renderTemplate(w, "cache", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			dh.logger.WithError(err).Error("Failed to render cache debug")
		}
	}
}

// Helper methods

func (dh *DebugHandler) getAvailableEndpoints() []string {
	endpoints := []string{"collectors", "tracing", "patterns", "scheduler", "health", "runtime", "config"}
	var available []string

	for _, endpoint := range endpoints {
		if dh.isEndpointEnabled(endpoint) {
			available = append(available, endpoint)
		}
	}

	return available
}

func (dh *DebugHandler) getComponentStatus() map[string]bool {
	return map[string]bool{
		"scheduler":  dh.scheduler != nil,
		"filter":     dh.filter != nil,
		"tracer":     dh.tracer != nil,
		"health":     dh.health != nil,
		"collectors": dh.collectors != nil,
		"cache":      dh.cache != nil,
	}
}

func (dh *DebugHandler) getCollectorSummary(states map[string]CollectorState, metrics map[string]CollectorMetrics) map[string]interface{} {
	_ = metrics
	total := len(states)
	enabled := 0
	errors := 0

	for _, state := range states {
		if state.Enabled {
			enabled++
		}
		if state.ConsecutiveErrors > 0 {
			errors++
		}
	}

	return map[string]interface{}{
		"total":   total,
		"enabled": enabled,
		"errors":  errors,
	}
}

func (dh *DebugHandler) filterPatterns(patterns map[string]filtering.MetricPattern, action string, minSamples int) map[string]filtering.MetricPattern {
	filtered := make(map[string]filtering.MetricPattern)

	for key, pattern := range patterns {
		// Filter by action
		if action != "" && pattern.FilterRecommend.String() != action {
			continue
		}

		// Filter by minimum samples
		if minSamples > 0 && pattern.SampleCount < minSamples {
			continue
		}

		filtered[key] = pattern
	}

	return filtered
}

func (dh *DebugHandler) sortPatterns(patterns map[string]filtering.MetricPattern, sortBy string) []filtering.MetricPattern {
	var sorted []filtering.MetricPattern //nolint:prealloc

	for _, pattern := range patterns {
		sorted = append(sorted, pattern)
	}

	switch sortBy {
	case "noise":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].NoiseScore > sorted[j].NoiseScore
		})
	case "samples":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].SampleCount > sorted[j].SampleCount
		})
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
	default:
		// Sort by last updated (most recent first)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].LastUpdated.After(sorted[j].LastUpdated)
		})
	}

	return sorted
}

func (dh *DebugHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		dh.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (dh *DebugHandler) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return dh.templates.ExecuteTemplate(w, name, data)
}

// IsEnabled returns whether debug endpoints are enabled
func (dh *DebugHandler) IsEnabled() bool {
	return dh.enabled
}

// GetStats returns debug handler statistics
func (dh *DebugHandler) GetStats() map[string]interface{} {
	if !dh.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	dh.mu.RLock()
	defer dh.mu.RUnlock()

	return map[string]interface{}{
		"enabled":       true,
		"uptime":        time.Since(dh.startTime),
		"request_count": dh.requestCount,
		"endpoints":     dh.getAvailableEndpoints(),
		"components":    dh.getComponentStatus(),
	}
}

// sanitizeLogField removes control characters and newlines from log fields to prevent log injection
func sanitizeLogField(input string) string {
	var result strings.Builder
	for _, r := range input {
		// Allow printable ASCII and common safe characters
		if (r >= 32 && r <= 126) || r == '\t' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
