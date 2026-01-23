// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Simplified validation tests that focus on functionality rather than exact error message formats
func TestEnhancedValidation_Functional(t *testing.T) {
	t.Run("validation system is working", func(t *testing.T) {
		// Just test that validation doesn't panic and returns structured errors
		cfg := &Config{}
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "empty config should fail validation")
		assert.Contains(t, err.Error(), "configuration validation failed", "should return structured validation error")
	})

	t.Run("invalid server address should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Server.Address = "invalid-address"
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "invalid server address should fail validation")
	})

	t.Run("missing metrics path should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Server.MetricsPath = ""
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "missing metrics path should fail validation")
	})

	t.Run("invalid metrics path should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Server.MetricsPath = "no-leading-slash"
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "invalid metrics path should fail validation")
	})

	t.Run("missing SLURM base URL should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.SLURM.BaseURL = ""
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "missing SLURM base URL should fail validation")
	})

	t.Run("invalid auth type should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.SLURM.Auth.Type = "invalid-auth-type"
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "invalid auth type should fail validation")
	})

	t.Run("short collector timeout should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Collectors.Jobs.Timeout = 10 * time.Millisecond
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "short collector timeout should fail validation")
	})

	t.Run("TLS enabled without cert should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Server.TLS.Enabled = true
		cfg.Server.TLS.CertFile = ""
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "TLS enabled without cert should fail validation")
	})

	t.Run("conflicting metric filters should fail", func(t *testing.T) {
		cfg := getValidTestConfig()
		cfg.Collectors.Jobs.Filters.Metrics.OnlyCounters = true
		cfg.Collectors.Jobs.Filters.Metrics.OnlyGauges = true
		err := cfg.ValidateEnhanced()
		assert.Error(t, err, "conflicting metric filters should fail validation")
	})
}

// getValidTestConfig returns a configuration that should pass all validation
func getValidTestConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:        ":8080",
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			ReadyPath:      "/ready",
			Timeout:        30 * time.Second,
			IdleTimeout:    60 * time.Second,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxRequestSize: 10 * 1024 * 1024,
		},
		SLURM: SLURMConfig{
			BaseURL:    "http://localhost:6820",
			Timeout:    30 * time.Second,
			APIVersion: "v0.0.43",
			Auth: AuthConfig{
				Type:  "jwt",
				Token: "valid-jwt-token",
			},
		},
		Collectors: CollectorsConfig{
			Global: GlobalCollectorConfig{
				DefaultInterval: 15 * time.Second,
			},
			Jobs: CollectorConfig{
				Enabled: true,
				Timeout: 15 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
			},
		},
		Metrics: MetricsConfig{
			Cardinality: CardinalityConfig{
				WarnLimit: 5000,
				MaxSeries: 10000,
			},
		},
	}
}
