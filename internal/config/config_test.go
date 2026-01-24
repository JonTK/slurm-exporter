// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Parallel()
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Expected no error loading default config, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config to be non-nil")
	}

	// Test server config defaults
	if cfg.Server.Address != ":8080" {
		t.Errorf("Expected server address to be ':8080', got: %s", cfg.Server.Address)
	}

	if cfg.Server.MetricsPath != "/metrics" {
		t.Errorf("Expected metrics path to be '/metrics', got: %s", cfg.Server.MetricsPath)
	}

	if cfg.Server.HealthPath != "/health" {
		t.Errorf("Expected health path to be '/health', got: %s", cfg.Server.HealthPath)
	}

	if cfg.Server.ReadyPath != "/ready" {
		t.Errorf("Expected ready path to be '/ready', got: %s", cfg.Server.ReadyPath)
	}

	// Test SLURM config defaults
	if cfg.SLURM.BaseURL != "http://localhost:6820" {
		t.Errorf("Expected SLURM base URL to be 'http://localhost:6820', got: %s", cfg.SLURM.BaseURL)
	}

	if cfg.SLURM.APIVersion != "v0.0.43" {
		t.Errorf("Expected SLURM API version to be 'v0.0.43', got: %s", cfg.SLURM.APIVersion)
	}

	if cfg.SLURM.Auth.Type != "none" {
		t.Errorf("Expected auth type to be 'none', got: %s", cfg.SLURM.Auth.Type)
	}

	// Test collector defaults
	if !cfg.Collectors.Cluster.Enabled {
		t.Error("Expected cluster collector to be enabled")
	}

	if cfg.Collectors.Cluster.Interval != 30*time.Second {
		t.Errorf("Expected cluster collector interval to be 30s, got: %v", cfg.Collectors.Cluster.Interval)
	}

	if !cfg.Collectors.Jobs.Enabled {
		t.Error("Expected jobs collector to be enabled")
	}

	if cfg.Collectors.Jobs.Interval != 15*time.Second {
		t.Errorf("Expected jobs collector interval to be 15s, got: %v", cfg.Collectors.Jobs.Interval)
	}

	// Test logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected log level to be 'info', got: %s", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "json" {
		t.Errorf("Expected log format to be 'json', got: %s", cfg.Logging.Format)
	}

	// Test metrics config defaults
	if cfg.Metrics.Namespace != "slurm" {
		t.Errorf("Expected metrics namespace to be 'slurm', got: %s", cfg.Metrics.Namespace)
	}

	if cfg.Metrics.Subsystem != "exporter" {
		t.Errorf("Expected metrics subsystem to be 'exporter', got: %s", cfg.Metrics.Subsystem)
	}
}

func TestServerConfig(t *testing.T) {
	t.Parallel()
	cfg, _ := Load("test-config.yaml")

	// Test timeout values
	if cfg.Server.Timeout != 30*time.Second {
		t.Errorf("Expected server timeout to be 30s, got: %v", cfg.Server.Timeout)
	}

	if cfg.Server.ReadTimeout != 15*time.Second {
		t.Errorf("Expected read timeout to be 15s, got: %v", cfg.Server.ReadTimeout)
	}

	if cfg.Server.WriteTimeout != 10*time.Second {
		t.Errorf("Expected write timeout to be 10s, got: %v", cfg.Server.WriteTimeout)
	}

	if cfg.Server.IdleTimeout != 60*time.Second {
		t.Errorf("Expected idle timeout to be 60s, got: %v", cfg.Server.IdleTimeout)
	}

	// Test TLS defaults
	if cfg.Server.TLS.Enabled {
		t.Error("Expected TLS to be disabled by default")
	}

	// Test BasicAuth defaults
	if cfg.Server.BasicAuth.Enabled {
		t.Error("Expected BasicAuth to be disabled by default")
	}

	// Test CORS defaults
	if cfg.Server.CORS.Enabled {
		t.Error("Expected CORS to be disabled by default")
	}
}

func TestSLURMConfig(t *testing.T) {
	t.Parallel()
	cfg, _ := Load("test-config.yaml")

	// Test retry settings
	if cfg.SLURM.RetryAttempts != 3 {
		t.Errorf("Expected retry attempts to be 3, got: %d", cfg.SLURM.RetryAttempts)
	}

	if cfg.SLURM.RetryDelay != 5*time.Second {
		t.Errorf("Expected retry delay to be 5s, got: %v", cfg.SLURM.RetryDelay)
	}

	// Test TLS settings
	if cfg.SLURM.TLS.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be false by default")
	}

	// Test rate limit settings
	if cfg.SLURM.RateLimit.RequestsPerSecond != 10.0 {
		t.Errorf("Expected requests per second to be 10.0, got: %f", cfg.SLURM.RateLimit.RequestsPerSecond)
	}

	if cfg.SLURM.RateLimit.BurstSize != 20 {
		t.Errorf("Expected burst size to be 20, got: %d", cfg.SLURM.RateLimit.BurstSize)
	}
}

func TestCollectorsConfig(t *testing.T) {
	t.Parallel()
	cfg, _ := Load("test-config.yaml")

	// Test global collector settings
	if cfg.Collectors.Global.DefaultInterval != 30*time.Second {
		t.Errorf("Expected default interval to be 30s, got: %v", cfg.Collectors.Global.DefaultInterval)
	}

	if cfg.Collectors.Global.MaxConcurrency != 5 {
		t.Errorf("Expected max concurrency to be 5, got: %d", cfg.Collectors.Global.MaxConcurrency)
	}

	if !cfg.Collectors.Global.GracefulDegradation {
		t.Error("Expected graceful degradation to be enabled")
	}

	// Test individual collector settings
	collectors := []CollectorConfig{
		cfg.Collectors.Cluster,
		cfg.Collectors.Nodes,
		cfg.Collectors.Jobs,
		cfg.Collectors.Users,
		cfg.Collectors.Partitions,
		cfg.Collectors.Performance,
		cfg.Collectors.System,
	}

	for i, collector := range collectors {
		if !collector.Enabled {
			t.Errorf("Expected collector %d to be enabled", i)
		}

		if collector.Timeout != 10*time.Second {
			t.Errorf("Expected collector %d timeout to be 10s, got: %v", i, collector.Timeout)
		}

		// Test error handling settings
		if collector.ErrorHandling.MaxRetries != 3 {
			t.Errorf("Expected collector %d max retries to be 3, got: %d", i, collector.ErrorHandling.MaxRetries)
		}

		if collector.ErrorHandling.BackoffFactor != 2.0 {
			t.Errorf("Expected collector %d backoff factor to be 2.0, got: %f", i, collector.ErrorHandling.BackoffFactor)
		}
	}

	// Test specific interval settings
	if cfg.Collectors.Jobs.Interval != 15*time.Second {
		t.Errorf("Expected jobs collector interval to be 15s, got: %v", cfg.Collectors.Jobs.Interval)
	}

	if cfg.Collectors.Users.Interval != 60*time.Second {
		t.Errorf("Expected users collector interval to be 60s, got: %v", cfg.Collectors.Users.Interval)
	}
}

func TestMetricsConfig(t *testing.T) {
	t.Parallel()
	cfg, _ := Load("test-config.yaml")

	// Test metrics settings
	if cfg.Metrics.MaxAge != 5*time.Minute {
		t.Errorf("Expected max age to be 5m, got: %v", cfg.Metrics.MaxAge)
	}

	if cfg.Metrics.AgeBuckets != 5 {
		t.Errorf("Expected age buckets to be 5, got: %d", cfg.Metrics.AgeBuckets)
	}

	// Test registry settings
	if !cfg.Metrics.Registry.EnableGoCollector {
		t.Error("Expected Go collector to be enabled")
	}

	if !cfg.Metrics.Registry.EnableProcessCollector {
		t.Error("Expected process collector to be enabled")
	}

	if !cfg.Metrics.Registry.EnableBuildInfo {
		t.Error("Expected build info to be enabled")
	}

	// Test cardinality settings
	if cfg.Metrics.Cardinality.MaxSeries != 10000 {
		t.Errorf("Expected max series to be 10000, got: %d", cfg.Metrics.Cardinality.MaxSeries)
	}

	if cfg.Metrics.Cardinality.WarnLimit != 8000 {
		t.Errorf("Expected warn limit to be 8000, got: %d", cfg.Metrics.Cardinality.WarnLimit)
	}
}

func TestConfigStructures(t *testing.T) {
	t.Parallel(
	// Test that all config structures can be instantiated
	)

	var cfg Config
	var serverCfg ServerConfig
	var slurmCfg SLURMConfig
	var authCfg AuthConfig
	var collectorsCfg CollectorsConfig
	var collectorCfg CollectorConfig
	var loggingCfg LoggingConfig
	var metricsCfg MetricsConfig

	// Test that structures have the expected zero values
	if cfg.Server.Address != "" {
		t.Error("Expected zero value for server address")
	}

	if serverCfg.TLS.Enabled {
		t.Error("Expected TLS to be disabled by default")
	}

	if slurmCfg.RetryAttempts != 0 {
		t.Error("Expected zero retry attempts by default")
	}

	if authCfg.Type != "" {
		t.Error("Expected empty auth type by default")
	}

	if collectorsCfg.Global.MaxConcurrency != 0 {
		t.Error("Expected zero max concurrency by default")
	}

	if collectorCfg.Enabled {
		t.Error("Expected collector to be disabled by default")
	}

	if loggingCfg.Level != "" {
		t.Error("Expected empty log level by default")
	}

	if metricsCfg.Namespace != "" {
		t.Error("Expected empty metrics namespace by default")
	}
}

func TestLoadFromYAMLFile(t *testing.T) {
	t.Parallel(
	// Create a temporary YAML config file
	)

	yamlContent := `
server:
  address: ":9090"
  metrics_path: "/custom/metrics"

slurm:
  base_url: "https://custom-slurm:6820"
  api_version: "v0.0.43"
  auth:
    type: "jwt"
    token: "test-token"

collectors:
  cluster:
    enabled: true
    interval: "60s"
  jobs:
    enabled: false

logging:
  level: "debug"
  format: "text"

metrics:
  namespace: "custom_slurm"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error loading YAML config, got: %v", err)
	}

	// Test that YAML values override defaults
	if cfg.Server.Address != ":9090" {
		t.Errorf("Expected server address to be ':9090', got: %s", cfg.Server.Address)
	}

	if cfg.Server.MetricsPath != "/custom/metrics" {
		t.Errorf("Expected metrics path to be '/custom/metrics', got: %s", cfg.Server.MetricsPath)
	}

	if cfg.SLURM.BaseURL != "https://custom-slurm:6820" {
		t.Errorf("Expected SLURM base URL to be 'https://custom-slurm:6820', got: %s", cfg.SLURM.BaseURL)
	}

	if cfg.SLURM.APIVersion != "v0.0.43" {
		t.Errorf("Expected API version to be 'v0.0.43', got: %s", cfg.SLURM.APIVersion)
	}

	if cfg.SLURM.Auth.Type != "jwt" {
		t.Errorf("Expected auth type to be 'jwt', got: %s", cfg.SLURM.Auth.Type)
	}

	if cfg.SLURM.Auth.Token != "test-token" {
		t.Errorf("Expected auth token to be 'test-token', got: %s", cfg.SLURM.Auth.Token)
	}

	if cfg.Collectors.Cluster.Interval != 60*time.Second {
		t.Errorf("Expected cluster interval to be 60s, got: %v", cfg.Collectors.Cluster.Interval)
	}

	if cfg.Collectors.Jobs.Enabled {
		t.Error("Expected jobs collector to be disabled")
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level to be 'debug', got: %s", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "text" {
		t.Errorf("Expected log format to be 'text', got: %s", cfg.Logging.Format)
	}

	if cfg.Metrics.Namespace != "custom_slurm" {
		t.Errorf("Expected metrics namespace to be 'custom_slurm', got: %s", cfg.Metrics.Namespace)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	t.Parallel()
	_, err := Load("non-existent-file.yaml")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	t.Parallel()
	invalidYAML := `
server:
  address: ":8080"
  invalid_yaml: [unclosed
`

	tmpFile, err := os.CreateTemp("", "invalid-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(invalidYAML); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

func TestValidateServerConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config ServerConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: ServerConfig{
				Address:        ":8080",
				MetricsPath:    "/metrics",
				HealthPath:     "/health",
				ReadyPath:      "/ready",
				Timeout:        30 * time.Second,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				IdleTimeout:    60 * time.Second,
				MaxRequestSize: 1024,
			},
			valid: true,
		},
		{
			name: "empty address",
			config: ServerConfig{
				Address:        "",
				MetricsPath:    "/metrics",
				HealthPath:     "/health",
				ReadyPath:      "/ready",
				Timeout:        30 * time.Second,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				IdleTimeout:    60 * time.Second,
				MaxRequestSize: 1024,
			},
			valid: false,
		},
		{
			name: "negative timeout",
			config: ServerConfig{
				Address:        ":8080",
				MetricsPath:    "/metrics",
				HealthPath:     "/health",
				ReadyPath:      "/ready",
				Timeout:        -1 * time.Second,
				ReadTimeout:    10 * time.Second,
				WriteTimeout:   10 * time.Second,
				IdleTimeout:    60 * time.Second,
				MaxRequestSize: 1024,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestValidateAuthConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config AuthConfig
		valid  bool
	}{
		{
			name: "none auth",
			config: AuthConfig{
				Type: "none",
			},
			valid: true,
		},
		{
			name: "jwt with token",
			config: AuthConfig{
				Type:  "jwt",
				Token: "test-token",
			},
			valid: true,
		},
		{
			name: "jwt without token",
			config: AuthConfig{
				Type: "jwt",
			},
			valid: false,
		},
		{
			name: "basic auth with credentials",
			config: AuthConfig{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			valid: true,
		},
		{
			name: "basic auth without username",
			config: AuthConfig{
				Type:     "basic",
				Password: "pass",
			},
			valid: false,
		},
		{
			name: "invalid auth type",
			config: AuthConfig{
				Type: "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestValidateLoggingConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config LoggingConfig
		valid  bool
	}{
		{
			name: "valid logging config",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			valid: true,
		},
		{
			name: "invalid log level",
			config: LoggingConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			valid: false,
		},
		{
			name: "invalid log format",
			config: LoggingConfig{
				Level:  "info",
				Format: "invalid",
				Output: "stdout",
			},
			valid: false,
		},
		{
			name: "file output without file",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "file",
				File:   "",
			},
			valid: false,
		},
		{
			name: "file output with file",
			config: LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "file",
				File:   "/var/log/app.log",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestValidateCollectorConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config CollectorConfig
		valid  bool
	}{
		{
			name: "disabled collector",
			config: CollectorConfig{
				Enabled: false,
			},
			valid: true,
		},
		{
			name: "enabled collector with valid config",
			config: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			valid: true,
		},
		{
			name: "enabled collector with invalid interval",
			config: CollectorConfig{
				Enabled:  true,
				Interval: -1 * time.Second,
				Timeout:  10 * time.Second,
			},
			valid: false,
		},
		{
			name: "enabled collector with invalid timeout",
			config: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  -1 * time.Second,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.config.Validate()
			if tt.valid && err != nil {
				t.Errorf("Expected config to be valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected config to be invalid, got no error")
			}
		})
	}
}

func TestDefault(t *testing.T) {
	t.Parallel()
	cfg := Default()
	if cfg == nil {
		t.Fatal("Expected default config to be non-nil")
	}

	// Test that default config is valid
	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected default config to be valid, got error: %v", err)
	}

	// Test some key defaults
	if cfg.Server.Address != ":8080" {
		t.Errorf("Expected default server address to be ':8080', got: %s", cfg.Server.Address)
	}

	if cfg.SLURM.APIVersion != "v0.0.43" {
		t.Errorf("Expected default API version to be 'v0.0.43', got: %s", cfg.SLURM.APIVersion)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level to be 'info', got: %s", cfg.Logging.Level)
	}
}

func TestEnvOverrides(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"SLURM_EXPORTER_SERVER_ADDRESS":           ":9999",
		"SLURM_EXPORTER_SERVER_METRICS_PATH":      "/custom/metrics",
		"SLURM_EXPORTER_SLURM_BASE_URL":           "https://env-slurm:6820",
		"SLURM_EXPORTER_SLURM_AUTH_TYPE":          "jwt",
		"SLURM_EXPORTER_SLURM_AUTH_TOKEN":         "env-token",
		"SLURM_EXPORTER_COLLECTORS_JOBS_ENABLED":  "false",
		"SLURM_EXPORTER_COLLECTORS_JOBS_INTERVAL": "5m",
		"SLURM_EXPORTER_LOGGING_LEVEL":            "debug",
		"SLURM_EXPORTER_LOGGING_FORMAT":           "text",
		"SLURM_EXPORTER_METRICS_NAMESPACE":        "custom_env",
	}

	// Set environment variables
	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		// Clean up environment variables
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Expected no error loading config with env overrides, got: %v", err)
	}

	// Test that environment variables override defaults
	if cfg.Server.Address != ":9999" {
		t.Errorf("Expected server address to be ':9999', got: %s", cfg.Server.Address)
	}

	if cfg.Server.MetricsPath != "/custom/metrics" {
		t.Errorf("Expected metrics path to be '/custom/metrics', got: %s", cfg.Server.MetricsPath)
	}

	if cfg.SLURM.BaseURL != "https://env-slurm:6820" {
		t.Errorf("Expected SLURM base URL to be 'https://env-slurm:6820', got: %s", cfg.SLURM.BaseURL)
	}

	if cfg.SLURM.Auth.Type != "jwt" {
		t.Errorf("Expected auth type to be 'jwt', got: %s", cfg.SLURM.Auth.Type)
	}

	if cfg.SLURM.Auth.Token != "env-token" {
		t.Errorf("Expected auth token to be 'env-token', got: %s", cfg.SLURM.Auth.Token)
	}

	if cfg.Collectors.Jobs.Enabled {
		t.Error("Expected jobs collector to be disabled")
	}

	if cfg.Collectors.Jobs.Interval != 5*time.Minute {
		t.Errorf("Expected jobs collector interval to be 5m, got: %v", cfg.Collectors.Jobs.Interval)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level to be 'debug', got: %s", cfg.Logging.Level)
	}

	if cfg.Logging.Format != "text" {
		t.Errorf("Expected log format to be 'text', got: %s", cfg.Logging.Format)
	}

	if cfg.Metrics.Namespace != "custom_env" {
		t.Errorf("Expected metrics namespace to be 'custom_env', got: %s", cfg.Metrics.Namespace)
	}
}

func TestEnvOverridesPrecedence(t *testing.T) {
	// Create a YAML config file
	yamlContent := `
server:
  address: ":8888"
  metrics_path: "/yaml/metrics"

slurm:
  auth:
    type: "basic"
    username: "yaml-user"
    password: "yaml-pass"

logging:
  level: "warn"
`

	tmpFile, err := os.CreateTemp("", "precedence-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpFile.Close()

	// Set environment variables that should override YAML
	envVars := map[string]string{
		"SLURM_EXPORTER_SERVER_ADDRESS":   ":7777",
		"SLURM_EXPORTER_SLURM_AUTH_TYPE":  "jwt",
		"SLURM_EXPORTER_SLURM_AUTH_TOKEN": "jwt-token",
		"SLURM_EXPORTER_LOGGING_LEVEL":    "error",
	}

	for key, value := range envVars {
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key := range envVars {
			_ = os.Unsetenv(key)
		}
	}()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error loading config, got: %v", err)
	}

	// Test that environment variables override YAML values
	if cfg.Server.Address != ":7777" {
		t.Errorf("Expected env var to override YAML, server address should be ':7777', got: %s", cfg.Server.Address)
	}

	// Test that YAML values are used when no env var is set
	if cfg.Server.MetricsPath != "/yaml/metrics" {
		t.Errorf("Expected YAML value to be used, metrics path should be '/yaml/metrics', got: %s", cfg.Server.MetricsPath)
	}

	// Test that environment variables override YAML auth config
	if cfg.SLURM.Auth.Type != "jwt" {
		t.Errorf("Expected env var to override YAML, auth type should be 'jwt', got: %s", cfg.SLURM.Auth.Type)
	}

	// Test that environment variables override YAML logging config
	if cfg.Logging.Level != "error" {
		t.Errorf("Expected env var to override YAML, log level should be 'error', got: %s", cfg.Logging.Level)
	}
}

func TestInvalidEnvOverrides(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		value  string
	}{
		{
			name:   "invalid server timeout",
			envVar: "SLURM_EXPORTER_SERVER_TIMEOUT",
			value:  "invalid-duration",
		},
		{
			name:   "invalid TLS enabled",
			envVar: "SLURM_EXPORTER_SERVER_TLS_ENABLED",
			value:  "invalid-bool",
		},
		{
			name:   "invalid retry attempts",
			envVar: "SLURM_EXPORTER_SLURM_RETRY_ATTEMPTS",
			value:  "invalid-int",
		},
		{
			name:   "invalid requests per second",
			envVar: "SLURM_EXPORTER_SLURM_RATE_LIMIT_REQUESTS_PER_SECOND",
			value:  "invalid-float",
		},
		{
			name:   "invalid collector enabled",
			envVar: "SLURM_EXPORTER_COLLECTORS_JOBS_ENABLED",
			value:  "invalid-bool",
		},
		{
			name:   "invalid collector interval",
			envVar: "SLURM_EXPORTER_COLLECTORS_JOBS_INTERVAL",
			value:  "invalid-duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(tt.envVar, tt.value)
			defer func() { _ = os.Unsetenv(tt.envVar) }()

			_, err := Load("")
			if err == nil {
				t.Error("Expected error when loading config with invalid env var")
			}
		})
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	cfg := Default()

	// Test that ApplyEnvOverrides doesn't fail on default config
	if err := cfg.ApplyEnvOverrides(); err != nil {
		t.Errorf("Expected no error applying env overrides to default config, got: %v", err)
	}

	// Test with specific environment variable
	_ = os.Setenv("SLURM_EXPORTER_SERVER_ADDRESS", ":5555")
	defer func() { _ = os.Unsetenv("SLURM_EXPORTER_SERVER_ADDRESS") }()

	if err := cfg.ApplyEnvOverrides(); err != nil {
		t.Errorf("Expected no error applying env overrides, got: %v", err)
	}

	if cfg.Server.Address != ":5555" {
		t.Errorf("Expected server address to be ':5555', got: %s", cfg.Server.Address)
	}
}
