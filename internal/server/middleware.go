package server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// LoggingMiddleware provides structured HTTP request logging
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status and size
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log the request
		duration := time.Since(start)

		logEntry := s.logger.WithFields(logrus.Fields{
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"remote_addr":   r.RemoteAddr,
			"user_agent":    r.Header.Get("User-Agent"),
			"referer":       r.Header.Get("Referer"),
			"status":        rw.statusCode,
			"response_size": rw.written,
			"duration":      duration,
			"duration_ms":   float64(duration.Nanoseconds()) / 1000000,
		})

		// Add request ID if present
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			logEntry = logEntry.WithField("request_id", requestID)
		}

		// Log at different levels based on status code
		message := "HTTP request"
		switch {
		case rw.statusCode >= 500:
			logEntry.Error(message)
		case rw.statusCode >= 400:
			logEntry.Warn(message)
		default:
			logEntry.Info(message)
		}
	})
}

// MetricsMiddleware adds metrics collection for HTTP requests
func (s *Server) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Increment in-flight requests
		s.httpMetrics.requestsInFlight.Inc()
		defer s.httpMetrics.requestsInFlight.Dec()

		// Wrap the response writer to capture metrics
		rw := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     0,
			written:        0,
		}

		// Record request size
		requestSize := float64(r.ContentLength)
		if requestSize < 0 {
			requestSize = 0
		}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start)
		method := r.Method
		path := s.normalizePath(r.URL.Path)
		status := rw.statusCode
		if status == 0 {
			status = 200 // Default to 200 if not set
		}

		// Update metrics
		labels := []string{method, path}
		statusLabels := []string{method, path, strconv.Itoa(status)}

		s.httpMetrics.requestsTotal.WithLabelValues(statusLabels...).Inc()
		s.httpMetrics.requestDuration.WithLabelValues(labels...).Observe(duration.Seconds())
		s.httpMetrics.requestSize.WithLabelValues(labels...).Observe(requestSize)
		s.httpMetrics.responseSize.WithLabelValues(labels...).Observe(float64(rw.written))

		s.logger.WithFields(logrus.Fields{
			"component":     "http_metrics",
			"method":        method,
			"path":          path,
			"status":        status,
			"duration":      duration,
			"request_size":  requestSize,
			"response_size": rw.written,
		}).Debug("HTTP metrics recorded")
	})
}

// normalizePath normalizes URL paths for metrics to avoid cardinality explosion
func (s *Server) normalizePath(path string) string {
	switch path {
	case s.config.Server.MetricsPath:
		return "/metrics"
	case s.config.Server.HealthPath:
		return "/health"
	case s.config.Server.ReadyPath:
		return "/ready"
	case "/":
		return "/"
	default:
		return "/other"
	}
}

// metricsResponseWriter wraps http.ResponseWriter to capture metrics
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (mw *metricsResponseWriter) WriteHeader(code int) {
	mw.statusCode = code
	mw.ResponseWriter.WriteHeader(code)
}

func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	if mw.statusCode == 0 {
		mw.statusCode = http.StatusOK
	}
	n, err := mw.ResponseWriter.Write(b)
	mw.written += int64(n)
	return n, err
}

// HeadersMiddleware adds standard security and info headers
func (s *Server) HeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Add server info
		w.Header().Set("Server", "slurm-exporter")

		// Add cache control for metrics endpoint
		if r.URL.Path == s.config.Server.MetricsPath {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		next.ServeHTTP(w, r)
	})
}

// RecoveryMiddleware recovers from panics and logs them
func (s *Server) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithFields(logrus.Fields{
					"component":   "recovery",
					"method":      r.Method,
					"path":        r.URL.Path,
					"remote_addr": r.RemoteAddr,
					"panic":       err,
				}).Error("HTTP handler panic recovered")

				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds request timeout handling with context cancellation
func (s *Server) TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the incoming context is already cancelled
		select {
		case <-r.Context().Done():
			s.logger.WithFields(logrus.Fields{
				"component": "timeout_middleware",
				"path":      r.URL.Path,
				"method":    r.Method,
				"error":     r.Context().Err(),
			}).Debug("Request context already cancelled")

			http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			return
		default:
		}

		// Create a context with timeout based on the request type
		var timeout time.Duration

		// Different timeouts for different endpoints
		switch r.URL.Path {
		case s.config.Server.MetricsPath:
			// Metrics endpoint gets longer timeout for collection
			timeout = 30 * time.Second
		case "/health":
			// Health check should be very fast
			timeout = 5 * time.Second
		case "/ready":
			// Readiness check may need to check collectors
			timeout = 10 * time.Second
		default:
			// Default timeout for other endpoints
			timeout = 15 * time.Second
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		// Add timeout information to request context
		r = r.WithContext(ctx)

		// Create a channel to handle completion
		done := make(chan struct{})

		// Run the request handler in a goroutine
		go func() {
			defer close(done)
			next.ServeHTTP(w, r)
		}()

		// Wait for completion or timeout
		select {
		case <-done:
			// Request completed normally
			return
		case <-ctx.Done():
			// Request timed out or was cancelled
			if ctx.Err() == context.DeadlineExceeded {
				s.logger.WithFields(logrus.Fields{
					"component": "timeout_middleware",
					"path":      r.URL.Path,
					"method":    r.Method,
					"timeout":   timeout,
				}).Warn("Request timeout exceeded")

				http.Error(w, "Request timeout", http.StatusGatewayTimeout)
			} else {
				s.logger.WithFields(logrus.Fields{
					"component": "timeout_middleware",
					"path":      r.URL.Path,
					"method":    r.Method,
					"error":     ctx.Err(),
				}).Debug("Request cancelled")

				http.Error(w, "Request cancelled", http.StatusRequestTimeout)
			}
			return
		}
	})
}

// CombinedMiddleware applies all middleware in the correct order
func (s *Server) CombinedMiddleware(next http.Handler) http.Handler {
	// Apply middleware in reverse order (last applied = first executed)
	handler := next
	handler = s.MetricsMiddleware(handler)
	handler = s.LoggingMiddleware(handler)
	handler = s.TimeoutMiddleware(handler)
	handler = s.HeadersMiddleware(handler)
	handler = s.RecoveryMiddleware(handler)

	return handler
}
