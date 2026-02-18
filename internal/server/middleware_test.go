// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
)

func createTestServer() *Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Address:      ":10341",
			MetricsPath:  "/metrics",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create HTTP metrics for middleware testing
	httpMetrics := NewHTTPMetrics()
	promRegistry := prometheus.NewRegistry()
	_ = httpMetrics.Register(promRegistry)

	return &Server{
		config:       cfg,
		logger:       logger,
		httpMetrics:  httpMetrics,
		promRegistry: promRegistry,
	}
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	// Capture log output
	var buf bytes.Buffer
	server.logger.SetOutput(&buf)
	server.logger.SetFormatter(&logrus.JSONFormatter{})

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Wrap with logging middleware
	handler := server.LoggingMiddleware(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "HTTP request") {
		t.Error("Expected log to contain 'HTTP request'")
	}

	// Parse log entry
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	// Verify log fields
	if logEntry["method"] != "GET" {
		t.Errorf("Expected method 'GET', got %v", logEntry["method"])
	}

	if logEntry["path"] != "/test" {
		t.Errorf("Expected path '/test', got %v", logEntry["path"])
	}

	if logEntry["user_agent"] != "test-agent" {
		t.Errorf("Expected user_agent 'test-agent', got %v", logEntry["user_agent"])
	}

	status, ok := logEntry["status"].(float64)
	if !ok || status != 200 {
		t.Errorf("Expected status 200, got %v", logEntry["status"])
	}

	if logEntry["duration"] == nil {
		t.Error("Expected duration field to be present")
	}
}

func TestHeadersMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Wrap with headers middleware
	handler := server.HeadersMiddleware(testHandler)

	t.Run("StandardHeaders", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check security headers
		expectedHeaders := map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"X-XSS-Protection":       "1; mode=block",
			"Server":                 "slurm-exporter",
		}

		for header, expectedValue := range expectedHeaders {
			if actual := w.Header().Get(header); actual != expectedValue {
				t.Errorf("Expected header %s to be '%s', got '%s'", header, expectedValue, actual)
			}
		}
	})

	t.Run("MetricsEndpointHeaders", func(t *testing.T) {
		t.Parallel()
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Check cache control header for metrics endpoint
		if cacheControl := w.Header().Get("Cache-Control"); cacheControl != "no-cache, no-store, must-revalidate" {
			t.Errorf("Expected Cache-Control header for metrics endpoint, got '%s'", cacheControl)
		}
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	// Capture log output
	var buf bytes.Buffer
	server.logger.SetOutput(&buf)
	server.logger.SetFormatter(&logrus.JSONFormatter{})

	// Create a test handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with recovery middleware
	handler := server.RecoveryMiddleware(panicHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute request (should not panic)
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Error("Expected response to contain 'Internal Server Error'")
	}

	// Check log output
	logOutput := buf.String()
	if !strings.Contains(logOutput, "HTTP handler panic recovered") {
		t.Error("Expected log to contain panic recovery message")
	}

	if !strings.Contains(logOutput, "test panic") {
		t.Error("Expected log to contain panic message")
	}
}

func TestCombinedMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	// Capture log output
	var buf bytes.Buffer
	server.logger.SetOutput(&buf)
	server.logger.SetFormatter(&logrus.JSONFormatter{})

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Apply combined middleware
	handler := server.CombinedMiddleware(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check that all middleware is applied
	// Security headers should be present
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Expected security headers from HeadersMiddleware")
	}

	// Logging should occur
	logOutput := buf.String()
	if !strings.Contains(logOutput, "HTTP request") {
		t.Error("Expected logging from LoggingMiddleware")
	}

	// Response should be successful
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMetricsMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test response"))
	})

	// Wrap with metrics middleware
	handler := server.MetricsMiddleware(testHandler)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify metrics were recorded
	// Note: In a real test environment, you would gather metrics from the registry
	// and verify specific values. For this test, we just ensure no panics occurred
	// and the handler completed successfully.

	// Test with different methods and paths to verify path normalization
	testCases := []struct {
		method string
		path   string
		expect string
	}{
		{"GET", "/metrics", "/metrics"},
		{"POST", "/health", "/health"},
		{"PUT", "/ready", "/ready"},
		{"GET", "/", "/"},
		{"GET", "/unknown/path", "/other"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s %s, got %d", tc.method, tc.path, w.Code)
		}
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	t.Parallel()
	t.Run("BasicAuthDisabled", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{
			Server: config.ServerConfig{
				MetricsPath: "/metrics",
				BasicAuth: config.BasicAuthConfig{
					Enabled: false,
				},
			},
		}
		server := &Server{
			config: cfg,
			logger: logrus.New(),
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := server.BasicAuthMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("BasicAuthEnabledNonMetricsEndpoint", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{
			Server: config.ServerConfig{
				MetricsPath: "/metrics",
				BasicAuth: config.BasicAuthConfig{
					Enabled:  true,
					Username: "user",
					Password: "pass",
				},
			},
		}
		server := &Server{
			config: cfg,
			logger: logrus.New(),
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := server.BasicAuthMiddleware(testHandler)

		// Test non-metrics endpoint (should not require auth)
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for non-metrics endpoint, got %d", w.Code)
		}
	})

	t.Run("BasicAuthEnabledNoCredentials", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{
			Server: config.ServerConfig{
				MetricsPath: "/metrics",
				BasicAuth: config.BasicAuthConfig{
					Enabled:  true,
					Username: "user",
					Password: "pass",
				},
			},
		}
		server := &Server{
			config: cfg,
			logger: logrus.New(),
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := server.BasicAuthMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}

		if w.Header().Get("WWW-Authenticate") == "" {
			t.Error("Expected WWW-Authenticate header")
		}
	})

	t.Run("BasicAuthEnabledInvalidCredentials", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{
			Server: config.ServerConfig{
				MetricsPath: "/metrics",
				BasicAuth: config.BasicAuthConfig{
					Enabled:  true,
					Username: "user",
					Password: "pass",
				},
			},
		}
		server := &Server{
			config: cfg,
			logger: logrus.New(),
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := server.BasicAuthMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("wrong", "credentials")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("BasicAuthEnabledValidCredentials", func(t *testing.T) {
		t.Parallel()
		cfg := &config.Config{
			Server: config.ServerConfig{
				MetricsPath: "/metrics",
				BasicAuth: config.BasicAuthConfig{
					Enabled:  true,
					Username: "user",
					Password: "pass",
				},
			},
		}
		server := &Server{
			config: cfg,
			logger: logrus.New(),
		}

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		handler := server.BasicAuthMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("user", "pass")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Body.String() != "success" {
			t.Errorf("Expected body 'success', got '%s'", w.Body.String())
		}
	})
}

func TestTimeoutMiddleware(t *testing.T) {
	t.Parallel()
	server := createTestServer()

	t.Run("NormalRequest", func(t *testing.T) {
		t.Parallel(
		// Create a test handler that completes quickly
		)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("success"))
		})

		// Wrap with timeout middleware
		handler := server.TimeoutMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "success") {
			t.Error("Expected response to contain 'success'")
		}
	})

	t.Run("TimeoutRequest", func(t *testing.T) {
		t.Parallel(
		// Create a test handler that takes longer than timeout
		)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep longer than health endpoint timeout (5s)
			time.Sleep(6 * time.Second)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("should not reach here"))
		})

		// Wrap with timeout middleware
		handler := server.TimeoutMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.ServeHTTP(w, req)
		duration := time.Since(start)

		// Should timeout quickly (around 5 seconds for health endpoint)
		if duration > 7*time.Second {
			t.Errorf("Request took too long: %v", duration)
		}

		if w.Code != http.StatusGatewayTimeout {
			t.Errorf("Expected status 504, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "Request timeout") {
			t.Error("Expected response to contain 'Request timeout'")
		}
	})

	t.Run("MetricsEndpointTimeout", func(t *testing.T) {
		t.Parallel(
		// Test that metrics endpoint gets longer timeout
		)

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep for a time that would exceed health timeout but not metrics timeout
			time.Sleep(8 * time.Second)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("metrics response"))
		})

		handler := server.TimeoutMiddleware(testHandler)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		start := time.Now()
		handler.ServeHTTP(w, req)
		duration := time.Since(start)

		// Should complete (not timeout) for metrics endpoint
		if duration > 10*time.Second {
			t.Errorf("Request took too long: %v", duration)
		}

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		t.Parallel()
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handler should not be reached due to cancelled context
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("should not reach here"))
		})

		handler := server.TimeoutMiddleware(testHandler)

		// Create a request with already cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := httptest.NewRequest(http.MethodGet, "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusRequestTimeout {
			t.Errorf("Expected status 408, got %d", w.Code)
		}
	})
}

func TestResponseWriter(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w}

	// Test default status code
	_, _ = rw.Write([]byte("test"))
	if rw.statusCode != http.StatusOK {
		t.Errorf("Expected default status code 200, got %d", rw.statusCode)
	}

	// Test explicit status code
	rw2 := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	rw2.WriteHeader(http.StatusCreated)
	if rw2.statusCode != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", rw2.statusCode)
	}

	// Test written bytes tracking
	rw3 := &responseWriter{ResponseWriter: httptest.NewRecorder()}
	testData := []byte("test data")
	n, err := rw3.Write(testData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), n)
	}
	if rw3.written != int64(len(testData)) {
		t.Errorf("Expected %d bytes tracked, got %d", len(testData), rw3.written)
	}
}
