package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/collector"
	"github.com/jontk/slurm-exporter/internal/config"
)

// mockRegistry implements RegistryInterface for testing
type mockRegistry struct {
	stats      map[string]collector.CollectorState
	shouldFail bool
}

func (m *mockRegistry) GetStats() map[string]collector.CollectorState {
	if m.stats == nil {
		return map[string]collector.CollectorState{
			"test_collector": {
				Name:    "test_collector",
				Enabled: true,
			},
		}
	}
	return m.stats
}

func (m *mockRegistry) CollectAll(ctx context.Context) error {
	if m.shouldFail {
		return fmt.Errorf("mock collection error")
	}
	return nil
}

func (m *mockRegistry) GetPerformanceStats() map[string]*collector.CollectorPerformanceStats {
	return map[string]*collector.CollectorPerformanceStats{
		"test_collector": {
			LastDuration:    100 * time.Millisecond,
			TotalDuration:   100 * time.Millisecond,
			CollectionCount: 1,
			SuccessCount:    1,
			ErrorCount:      0,
		},
	}
}

func createTestConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Address:      ":8080",
			MetricsPath:  "/metrics",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce test noise
	return logger
}

func TestNew(t *testing.T) {
	t.Run("WithValidConfig", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if server == nil {
			t.Fatal("Expected server to be created")
		}

		if server.config != cfg {
			t.Error("Expected config to be set")
		}

		if server.logger != logger {
			t.Error("Expected logger to be set")
		}

		if server.registry == nil {
			t.Error("Expected registry to be set")
		}

		if server.promRegistry == nil {
			t.Error("Expected Prometheus registry to be created")
		}

		if server.server == nil {
			t.Error("Expected HTTP server to be created")
		}
	})

	t.Run("WithNilRegistry", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()

		server, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if server.registry != nil {
			t.Error("Expected registry to be nil")
		}
	})
}

func TestHealthEndpoint(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLogger()
	registry := &mockRegistry{}

	server, err := New(cfg, logger, registry, prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", string(body))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", contentType)
	}
}

func TestReadyEndpoint(t *testing.T) {
	t.Run("WithEnabledCollectors", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{
			stats: map[string]collector.CollectorState{
				"collector1": {Name: "collector1", Enabled: true},
				"collector2": {Name: "collector2", Enabled: false},
			},
		}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		server.handleReady(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != "Ready" {
			t.Errorf("Expected body 'Ready', got '%s'", string(body))
		}
	})

	t.Run("WithNoEnabledCollectors", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{
			stats: map[string]collector.CollectorState{
				"collector1": {Name: "collector1", Enabled: false},
				"collector2": {Name: "collector2", Enabled: false},
			},
		}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		server.handleReady(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		if string(body) != "No collectors enabled" {
			t.Errorf("Expected body 'No collectors enabled', got '%s'", string(body))
		}
	})

	t.Run("WithNilRegistry", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()

		server, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		server.handleReady(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestRootEndpoint(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLogger()
	registry := &mockRegistry{
		stats: map[string]collector.CollectorState{
			"test_collector": {Name: "test_collector", Enabled: true},
		},
	}

	server, err := New(cfg, logger, registry, prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.handleRoot(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	content := string(body)

	// Check for essential content
	expectedContent := []string{
		"SLURM Prometheus Exporter",
		"Available Endpoints",
		"/metrics",
		"/health",
		"/ready",
		"Collector Status",
		"test_collector",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected content to contain '%s'", expected)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type 'text/html', got '%s'", contentType)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	t.Run("WithRegistry", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Register a test metric
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "test_gauge",
			Help: "A test gauge",
		})
		gauge.Set(42)

		if err := server.RegisterCollector(gauge); err != nil {
			t.Fatalf("Failed to register test metric: %v", err)
		}

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := server.createMetricsHandler()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		content := string(body)

		// Should contain our test metric
		if !strings.Contains(content, "test_gauge") {
			t.Error("Expected metrics to contain test_gauge")
		}

		// Should contain Go runtime metrics
		if !strings.Contains(content, "go_") {
			t.Error("Expected metrics to contain Go runtime metrics")
		}
	})

	t.Run("WithCollectionError", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{shouldFail: true}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := server.createMetricsHandler()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		// Should still return 200 even if collection fails
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("WithNilRegistry", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()

		server, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		handler := server.createMetricsHandler()
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}

func TestServerLifecycle(t *testing.T) {
	t.Run("StartAndShutdown", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Address:      ":0", // Use random port for testing
				MetricsPath:  "/metrics",
				ReadTimeout:  time.Second,
				WriteTimeout: time.Second,
				IdleTimeout:  time.Second,
			},
		}
		logger := createTestLogger()
		registry := &mockRegistry{}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Start server in background
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- server.Start(ctx)
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Shutdown server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			t.Errorf("Failed to shutdown server: %v", err)
		}

		// Wait for start to complete
		select {
		case err := <-errChan:
			if err != nil {
				t.Errorf("Unexpected error from Start(): %v", err)
			}
		case <-time.After(time.Second):
			t.Error("Start() did not complete within timeout")
		}
	})
}

func TestServerConfiguration(t *testing.T) {
	t.Run("GetMethods", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server.GetMetricsPath() != cfg.Server.MetricsPath {
			t.Errorf("Expected metrics path %s, got %s", cfg.Server.MetricsPath, server.GetMetricsPath())
		}

		if server.GetAddress() != cfg.Server.Address {
			t.Errorf("Expected address %s, got %s", cfg.Server.Address, server.GetAddress())
		}

		if server.GetPrometheusRegistry() == nil {
			t.Error("Expected Prometheus registry to be available")
		}
	})

	t.Run("RegistryMethods", func(t *testing.T) {
		cfg := createTestConfig()
		logger := createTestLogger()
		registry := &mockRegistry{}

		server, err := New(cfg, logger, registry, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// Test registering and unregistering a collector
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "test_registry_gauge",
			Help: "A test gauge for registry testing",
		})

		if err := server.RegisterCollector(gauge); err != nil {
			t.Errorf("Failed to register collector: %v", err)
		}

		if !server.UnregisterCollector(gauge) {
			t.Error("Failed to unregister collector")
		}

		// Test setting a new registry
		newRegistry := prometheus.NewRegistry()
		server.SetPrometheusRegistry(newRegistry)

		if server.GetPrometheusRegistry() != newRegistry {
			t.Error("Failed to set new Prometheus registry")
		}
	})
}

func TestTLSConfiguration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("TLSDisabled", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Address:      ":8080",
				MetricsPath:  "/metrics",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				TLS: config.TLSConfig{
					Enabled: false,
				},
			},
		}

		server, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		// TLS should not be configured
		if server.server.TLSConfig != nil {
			t.Error("Expected TLS config to be nil when TLS is disabled")
		}
	})

	t.Run("TLSEnabledMissingCertFile", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Address:      ":8443",
				MetricsPath:  "/metrics",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				TLS: config.TLSConfig{
					Enabled:  true,
					CertFile: "", // Missing cert file
					KeyFile:  "/path/to/key.pem",
				},
			},
		}

		_, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err == nil {
			t.Error("Expected error when cert_file is missing")
		}
		if !strings.Contains(err.Error(), "cert_file is required") {
			t.Errorf("Expected error about cert_file, got: %v", err)
		}
	})

	t.Run("TLSEnabledMissingKeyFile", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				Address:      ":8443",
				MetricsPath:  "/metrics",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				TLS: config.TLSConfig{
					Enabled:  true,
					CertFile: "/path/to/cert.pem",
					KeyFile:  "", // Missing key file
				},
			},
		}

		_, err := New(cfg, logger, nil, prometheus.NewRegistry())
		if err == nil {
			t.Error("Expected error when key_file is missing")
		}
		if !strings.Contains(err.Error(), "key_file is required") {
			t.Errorf("Expected error about key_file, got: %v", err)
		}
	})

	t.Run("TLSMinVersionParsing", func(t *testing.T) {
		// Create a test server to test TLS config parsing
		cfg := &config.Config{
			Server: config.ServerConfig{
				TLS: config.TLSConfig{
					Enabled:    true,
					CertFile:   "/nonexistent/cert.pem", // We won't actually load the cert
					KeyFile:    "/nonexistent/key.pem",
					MinVersion: "1.3",
				},
			},
		}

		server := &Server{
			config: cfg,
			logger: logger,
		}

		// This will fail because files don't exist, but we can test the version parsing logic
		_, err := server.createTLSConfig()
		// We expect it to fail on certificate loading, not version parsing
		if !strings.Contains(err.Error(), "failed to load TLS certificate") {
			t.Errorf("Expected certificate loading error, got: %v", err)
		}
	})

	t.Run("TLSInvalidMinVersion", func(t *testing.T) {
		cfg := &config.Config{
			Server: config.ServerConfig{
				TLS: config.TLSConfig{
					Enabled:    true,
					CertFile:   "/path/to/cert.pem",
					KeyFile:    "/path/to/key.pem",
					MinVersion: "invalid",
				},
			},
		}

		server := &Server{
			config: cfg,
			logger: logger,
		}

		_, err := server.createTLSConfig()
		if err == nil {
			t.Error("Expected error for invalid TLS version")
		}
		// Now version validation happens first
		if !strings.Contains(err.Error(), "unsupported TLS version") {
			t.Errorf("Expected unsupported TLS version error, got: %v", err)
		}
	})
}

func TestSetupRoutes(t *testing.T) {
	cfg := createTestConfig()
	logger := createTestLogger()
	registry := &mockRegistry{}

	server, err := New(cfg, logger, registry, prometheus.NewRegistry())
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	mux := server.setupRoutes()

	// Test that routes are properly configured
	testCases := []struct {
		path   string
		method string
	}{
		{"/health", "GET"},
		{"/ready", "GET"},
		{"/metrics", "GET"},
		{"/", "GET"},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		// All routes should respond (not 404)
		if w.Code == http.StatusNotFound {
			t.Errorf("Route %s %s returned 404", tc.method, tc.path)
		}
	}
}
