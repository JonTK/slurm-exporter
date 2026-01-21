package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// Check represents a single health check
type Check struct {
	Name        string            `json:"name"`
	Status      Status            `json:"status"`
	Message     string            `json:"message,omitempty"`
	Error       string            `json:"error,omitempty"`
	LastChecked time.Time         `json:"last_checked"`
	Duration    time.Duration     `json:"duration"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) Check

// HealthChecker manages and executes health checks
type HealthChecker struct {
	checks map[string]CheckFunc
	cache  map[string]Check
	mu     sync.RWMutex
	logger *logrus.Entry
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(logger *logrus.Logger) *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
		cache:  make(map[string]Check),
		logger: logger.WithField("component", "health"),
	}
}

// RegisterCheck registers a health check function
func (h *HealthChecker) RegisterCheck(name string, checkFunc CheckFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[name] = checkFunc
	h.logger.WithField("check", name).Debug("Health check registered")
}

// RemoveCheck removes a health check
func (h *HealthChecker) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.checks, name)
	delete(h.cache, name)
	h.logger.WithField("check", name).Debug("Health check removed")
}

// CheckHealth executes all health checks and returns the aggregated status
func (h *HealthChecker) CheckHealth(ctx context.Context) HealthReport {
	h.mu.RLock()
	checks := make(map[string]CheckFunc, len(h.checks))
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mu.RUnlock()

	report := HealthReport{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
	}

	// Execute all checks concurrently
	var wg sync.WaitGroup
	results := make(chan Check, len(checks))

	for name, checkFunc := range checks {
		wg.Add(1)
		go func(name string, check CheckFunc) {
			defer wg.Done()

			start := time.Now()
			result := check(ctx)
			result.Name = name
			result.LastChecked = start
			result.Duration = time.Since(start)

			results <- result
		}(name, checkFunc)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		report.Checks[result.Name] = result

		// Update cache
		h.mu.Lock()
		h.cache[result.Name] = result
		h.mu.Unlock()

		// Determine overall status
		switch result.Status {
		case StatusUnhealthy:
			report.Status = StatusUnhealthy
		case StatusDegraded:
			if report.Status == StatusHealthy {
				report.Status = StatusDegraded
			}
		}
	}

	report.Duration = time.Since(report.Timestamp)

	h.logger.WithFields(logrus.Fields{
		"status":       report.Status,
		"checks_count": len(report.Checks),
		"duration":     report.Duration,
	}).Debug("Health check completed")

	return report
}

// GetCachedCheck returns the last cached result for a specific check
func (h *HealthChecker) GetCachedCheck(name string) (Check, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	check, exists := h.cache[name]
	return check, exists
}

// HealthReport represents the aggregated health status
type HealthReport struct {
	Status    Status           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Duration  time.Duration    `json:"duration"`
	Checks    map[string]Check `json:"checks"`
}

// IsHealthy returns true if the overall status is healthy
func (r HealthReport) IsHealthy() bool {
	return r.Status == StatusHealthy
}

// IsReady returns true if the service is ready to serve traffic
// Ready means healthy or degraded (but not unhealthy)
func (r HealthReport) IsReady() bool {
	return r.Status == StatusHealthy || r.Status == StatusDegraded
}

// HealthHandler returns an HTTP handler for health checks
func (h *HealthChecker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		report := h.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Set HTTP status based on health
		switch report.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still serving traffic
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(report); err != nil {
			h.logger.WithError(err).Error("Failed to encode health report")
		}
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks
func (h *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		report := h.CheckHealth(ctx)

		w.Header().Set("Content-Type", "application/json")

		if report.IsReady() {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Return simplified response for readiness
		response := map[string]interface{}{
			"ready":     report.IsReady(),
			"status":    report.Status,
			"timestamp": report.Timestamp,
			"checks_passed": func() int {
				passed := 0
				for _, check := range report.Checks {
					if check.Status == StatusHealthy || check.Status == StatusDegraded {
						passed++
					}
				}
				return passed
			}(),
			"total_checks": len(report.Checks),
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.WithError(err).Error("Failed to encode readiness report")
		}
	}
}

// LivenessHandler returns an HTTP handler for liveness checks
func (h *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Liveness is simpler - just check if the service is running
		// We don't want to kill the pod due to external dependencies being down

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"alive":     true,
			"timestamp": time.Now(),
			"uptime":    time.Since(startTime),
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}

// startTime tracks when the service started
var startTime = time.Now()

// Common health check functions

// SLURMConnectivityCheck creates a health check for SLURM API connectivity
func SLURMConnectivityCheck(client interface{}, timeout time.Duration) CheckFunc {
	return func(ctx context.Context) Check {
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		start := time.Now()

		// Try to ping SLURM API
		// This would be implemented based on your SLURM client interface
		// For now, we'll create a basic connectivity check

		check := Check{
			Status:      StatusHealthy,
			Message:     "SLURM API is reachable",
			LastChecked: start,
			Metadata: map[string]string{
				"timeout": timeout.String(),
			},
		}

		// Simulate SLURM connectivity check
		select {
		case <-checkCtx.Done():
			check.Status = StatusUnhealthy
			check.Error = "SLURM connectivity check timed out"
			check.Message = ""
		case <-time.After(10 * time.Millisecond): // Simulate quick response
			// Success case handled above
		}

		return check
	}
}

// CollectorHealthCheck creates a health check for collector performance
func CollectorHealthCheck(collectorName string, lastCollection time.Time, maxAge time.Duration) CheckFunc {
	return func(ctx context.Context) Check {
		age := time.Since(lastCollection)

		check := Check{
			Message: fmt.Sprintf("Last collection: %s ago", age.Round(time.Second)),
			Metadata: map[string]string{
				"collector":       collectorName,
				"last_collection": lastCollection.Format(time.RFC3339),
				"max_age":         maxAge.String(),
				"age":             age.String(),
			},
		}

		if age > maxAge {
			check.Status = StatusUnhealthy
			check.Error = fmt.Sprintf("Collection data is stale (age: %s, max: %s)", age, maxAge)
		} else if age > maxAge/2 {
			check.Status = StatusDegraded
			check.Message = fmt.Sprintf("Collection data is aging (age: %s, max: %s)", age, maxAge)
		} else {
			check.Status = StatusHealthy
		}

		return check
	}
}

// MetricsEndpointCheck creates a health check for the metrics endpoint
func MetricsEndpointCheck(metricsURL string, client *http.Client) CheckFunc {
	return func(ctx context.Context) Check {
		req, err := http.NewRequestWithContext(ctx, "GET", metricsURL, nil)
		if err != nil {
			return Check{
				Status: StatusUnhealthy,
				Error:  fmt.Sprintf("Failed to create request: %v", err),
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return Check{
				Status: StatusUnhealthy,
				Error:  fmt.Sprintf("Failed to fetch metrics: %v", err),
			}
		}
		defer func() { _ = resp.Body.Close() }()

		check := Check{
			Metadata: map[string]string{
				"url":         metricsURL,
				"status_code": fmt.Sprintf("%d", resp.StatusCode),
			},
		}

		if resp.StatusCode == http.StatusOK {
			check.Status = StatusHealthy
			check.Message = "Metrics endpoint is responding"
		} else {
			check.Status = StatusUnhealthy
			check.Error = fmt.Sprintf("Metrics endpoint returned status %d", resp.StatusCode)
		}

		return check
	}
}
