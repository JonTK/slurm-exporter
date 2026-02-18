// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Authentication type constants
const (
	AuthTypeJWT  = "jwt"
	AuthTypeNone = "none"
)

// Config represents the application configuration.
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	SLURM         SLURMConfig         `yaml:"slurm"`
	Collectors    CollectorsConfig    `yaml:"collectors"`
	Logging       LoggingConfig       `yaml:"logging"`
	Metrics       MetricsConfig       `yaml:"metrics"`
	Observability ObservabilityConfig `yaml:"observability"`
	Validation    ValidationConfig    `yaml:"validation"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Address        string          `yaml:"address"`
	MetricsPath    string          `yaml:"metrics_path"`
	HealthPath     string          `yaml:"health_path"`
	ReadyPath      string          `yaml:"ready_path"`
	Timeout        time.Duration   `yaml:"timeout"`
	ReadTimeout    time.Duration   `yaml:"read_timeout"`
	WriteTimeout   time.Duration   `yaml:"write_timeout"`
	IdleTimeout    time.Duration   `yaml:"idle_timeout"`
	TLS            TLSConfig       `yaml:"tls"`
	BasicAuth      BasicAuthConfig `yaml:"basic_auth"`
	CORS           CORSConfig      `yaml:"cors"`
	MaxRequestSize int64           `yaml:"max_request_size"`
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled      bool     `yaml:"enabled"`
	CertFile     string   `yaml:"cert_file"`
	KeyFile      string   `yaml:"key_file"`
	MinVersion   string   `yaml:"min_version"`
	CipherSuites []string `yaml:"cipher_suites"`
}

// BasicAuthConfig holds basic authentication configuration for metrics endpoint.
type BasicAuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

// SLURMConfig holds SLURM connection configuration.
type SLURMConfig struct {
	BaseURL       string          `yaml:"base_url"`
	APIVersion    string          `yaml:"api_version"`
	UseAdapters   bool            `yaml:"use_adapters"`
	Auth          AuthConfig      `yaml:"auth"`
	Timeout       time.Duration   `yaml:"timeout"`
	RetryAttempts int             `yaml:"retry_attempts"`
	RetryDelay    time.Duration   `yaml:"retry_delay"`
	TLS           SLURMTLSConfig  `yaml:"tls"`
	RateLimit     RateLimitConfig `yaml:"rate_limit"`
}

// SLURMTLSConfig holds TLS configuration for SLURM connections.
type SLURMTLSConfig struct {
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	CACertFile         string `yaml:"ca_cert_file"`
	ClientCertFile     string `yaml:"client_cert_file"`
	ClientKeyFile      string `yaml:"client_key_file"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	BurstSize         int     `yaml:"burst_size"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Type         string            `yaml:"type"`          // jwt, basic, apikey, none
	Token        string            `yaml:"token"`         // For JWT
	TokenFile    string            `yaml:"token_file"`    // For JWT from file
	Username     string            `yaml:"username"`      // For basic auth
	Password     string            `yaml:"password"`      // For basic auth
	PasswordFile string            `yaml:"password_file"` // For basic auth from file
	APIKey       string            `yaml:"api_key"`       // For API key auth
	APIKeyFile   string            `yaml:"api_key_file"`  // For API key from file
	Headers      map[string]string `yaml:"headers"`       // Custom headers
}

// CollectorsConfig holds configuration for metric collectors.
type CollectorsConfig struct {
	Global            GlobalCollectorConfig `yaml:"global"`
	Cluster           CollectorConfig       `yaml:"cluster"`
	Nodes             CollectorConfig       `yaml:"nodes"`
	Jobs              CollectorConfig       `yaml:"jobs"`
	Users             CollectorConfig       `yaml:"users"`
	Accounts          CollectorConfig       `yaml:"accounts"`
	Associations      CollectorConfig       `yaml:"associations"`
	Partitions        CollectorConfig       `yaml:"partitions"`
	Performance       CollectorConfig       `yaml:"performance"`
	System            CollectorConfig       `yaml:"system"`
	QoS               CollectorConfig       `yaml:"qos"`
	Reservations      CollectorConfig       `yaml:"reservations"`
	Licenses          CollectorConfig       `yaml:"licenses"`
	Shares            CollectorConfig       `yaml:"shares"`
	Diagnostics       CollectorConfig       `yaml:"diagnostics"`
	TRES              CollectorConfig       `yaml:"tres"`
	WCKeys            CollectorConfig       `yaml:"wckeys"`
	Clusters          CollectorConfig       `yaml:"clusters"`
	Degradation       DegradationConfig     `yaml:"degradation"`
	CollectionTimeout time.Duration         `yaml:"collection_timeout"`
}

// GlobalCollectorConfig holds global collector settings.
type GlobalCollectorConfig struct {
	DefaultInterval     time.Duration         `yaml:"default_interval"`
	DefaultTimeout      time.Duration         `yaml:"default_timeout"`
	MaxConcurrency      int                   `yaml:"max_concurrency"`
	BatchProcessing     BatchProcessingConfig `yaml:"batch_processing"`
	ErrorThreshold      int                   `yaml:"error_threshold"`
	RecoveryDelay       time.Duration         `yaml:"recovery_delay"`
	GracefulDegradation bool                  `yaml:"graceful_degradation"`
}

// CollectorConfig holds configuration for individual collectors.
type CollectorConfig struct {
	Enabled        bool                `yaml:"enabled"`
	Interval       time.Duration       `yaml:"interval"`
	Timeout        time.Duration       `yaml:"timeout"`
	MaxConcurrency int                 `yaml:"max_concurrency"`
	Labels         map[string]string   `yaml:"labels"`
	Filters        FilterConfig        `yaml:"filters"`
	ErrorHandling  ErrorHandlingConfig `yaml:"error_handling"`
}

// FilterConfig holds filtering configuration for collectors.
type FilterConfig struct {
	// Entity filters
	IncludeNodes      []string `yaml:"include_nodes"`
	ExcludeNodes      []string `yaml:"exclude_nodes"`
	IncludePartitions []string `yaml:"include_partitions"`
	ExcludePartitions []string `yaml:"exclude_partitions"`
	IncludeUsers      []string `yaml:"include_users"`
	ExcludeUsers      []string `yaml:"exclude_users"`
	IncludeAccounts   []string `yaml:"include_accounts"`
	ExcludeAccounts   []string `yaml:"exclude_accounts"`
	IncludeQoS        []string `yaml:"include_qos"`
	ExcludeQoS        []string `yaml:"exclude_qos"`
	JobStates         []string `yaml:"job_states"`
	NodeStates        []string `yaml:"node_states"`

	// Metric filters
	Metrics MetricFilterConfig `yaml:"metrics"`
}

// MetricFilterConfig holds metric-specific filtering configuration.
type MetricFilterConfig struct {
	// Enable all metrics by default, disable specific ones
	EnableAll      bool     `yaml:"enable_all"`
	IncludeMetrics []string `yaml:"include_metrics"`
	ExcludeMetrics []string `yaml:"exclude_metrics"`

	// Metric-specific configurations
	OnlyInfo            bool `yaml:"only_info"`             // Only collect info metrics
	OnlyCounters        bool `yaml:"only_counters"`         // Only collect counter metrics
	OnlyGauges          bool `yaml:"only_gauges"`           // Only collect gauge metrics
	OnlyHistograms      bool `yaml:"only_histograms"`       // Only collect histogram metrics
	SkipHistograms      bool `yaml:"skip_histograms"`       // Skip histogram metrics (reduce cardinality)
	SkipTimingMetrics   bool `yaml:"skip_timing_metrics"`   // Skip timing-related metrics
	SkipResourceMetrics bool `yaml:"skip_resource_metrics"` // Skip resource usage metrics
}

// ErrorHandlingConfig holds error handling configuration.
type ErrorHandlingConfig struct {
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	BackoffFactor float64       `yaml:"backoff_factor"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay"`
	FailFast      bool          `yaml:"fail_fast"`
}

// DegradationConfig holds graceful degradation configuration.
type DegradationConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MaxFailures      int           `yaml:"max_failures"`
	ResetTimeout     time.Duration `yaml:"reset_timeout"`
	UseCachedMetrics bool          `yaml:"use_cached_metrics"`
	CacheTTL         time.Duration `yaml:"cache_ttl"`
}

// BatchProcessingConfig holds batch processing configuration.
type BatchProcessingConfig struct {
	Enabled           bool                         `yaml:"enabled"`
	MaxBatchSize      int                          `yaml:"max_batch_size"`
	MaxBatchWait      time.Duration                `yaml:"max_batch_wait"`
	FlushInterval     time.Duration                `yaml:"flush_interval"`
	MaxConcurrency    int                          `yaml:"max_concurrency"`
	EnableCompression bool                         `yaml:"enable_compression"`
	RetryAttempts     int                          `yaml:"retry_attempts"`
	RetryDelay        time.Duration                `yaml:"retry_delay"`
	MetricBatching    MetricBatchingConfig         `yaml:"metric_batching"`
	EntityBatching    map[string]EntityBatchConfig `yaml:"entity_batching"`
}

// MetricBatchingConfig holds metric-specific batching configuration.
type MetricBatchingConfig struct {
	EnableSampling    bool               `yaml:"enable_sampling"`
	SamplingRates     map[string]float64 `yaml:"sampling_rates"`
	EnableAggregation bool               `yaml:"enable_aggregation"`
	AggregationWindow time.Duration      `yaml:"aggregation_window"`
	MaxMetricAge      time.Duration      `yaml:"max_metric_age"`
}

// EntityBatchConfig holds entity-specific batch configuration.
type EntityBatchConfig struct {
	Enabled      bool          `yaml:"enabled"`
	MaxBatchSize int           `yaml:"max_batch_size"`
	MaxBatchWait time.Duration `yaml:"max_batch_wait"`
	Priority     int           `yaml:"priority"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level        string            `yaml:"level"`         // debug, info, warn, error
	Format       string            `yaml:"format"`        // json, text
	Output       string            `yaml:"output"`        // stdout, stderr, file
	File         string            `yaml:"file"`          // Log file path
	MaxSize      int               `yaml:"max_size"`      // Max size in MB
	MaxAge       int               `yaml:"max_age"`       // Max age in days
	MaxBackups   int               `yaml:"max_backups"`   // Max backup files
	Compress     bool              `yaml:"compress"`      // Compress rotated files
	Fields       map[string]string `yaml:"fields"`        // Additional fields
	SuppressHTTP bool              `yaml:"suppress_http"` // Suppress HTTP request logs
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Namespace   string            `yaml:"namespace"`
	Subsystem   string            `yaml:"subsystem"`
	ConstLabels map[string]string `yaml:"const_labels"`
	MaxAge      time.Duration     `yaml:"max_age"`
	AgeBuckets  int               `yaml:"age_buckets"`
	Registry    RegistryConfig    `yaml:"registry"`
	Cardinality CardinalityConfig `yaml:"cardinality"`
}

// RegistryConfig holds Prometheus registry configuration.
type RegistryConfig struct {
	EnableGoCollector      bool `yaml:"enable_go_collector"`
	EnableProcessCollector bool `yaml:"enable_process_collector"`
	EnableBuildInfo        bool `yaml:"enable_build_info"`
}

// CardinalityConfig holds cardinality management configuration.
type CardinalityConfig struct {
	MaxSeries    int `yaml:"max_series"`
	MaxLabels    int `yaml:"max_labels"`
	MaxLabelSize int `yaml:"max_label_size"`
	WarnLimit    int `yaml:"warn_limit"`
}

// ValidationConfig holds validation configuration.
type ValidationConfig struct {
	AllowInsecureConnections bool `yaml:"allow_insecure_connections"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Address:        ":10341",
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			ReadyPath:      "/ready",
			Timeout:        30 * time.Second,
			ReadTimeout:    15 * time.Second, // Must be longer than max collection timeout (10s)
			WriteTimeout:   10 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxRequestSize: 1024 * 1024, // 1MB
			TLS: TLSConfig{
				Enabled: false,
			},
			BasicAuth: BasicAuthConfig{
				Enabled: false,
			},
			CORS: CORSConfig{
				Enabled: false,
			},
		},
		SLURM: SLURMConfig{
			BaseURL:       "http://localhost:6820",
			APIVersion:    "", // Empty enables auto-detection
			UseAdapters:   true,
			Timeout:       30 * time.Second,
			RetryAttempts: 3,
			RetryDelay:    5 * time.Second,
			Auth: AuthConfig{
				Type: "none",
			},
			TLS: SLURMTLSConfig{
				InsecureSkipVerify: false,
			},
			RateLimit: RateLimitConfig{
				RequestsPerSecond: 10.0,
				BurstSize:         20,
			},
		},
		Collectors: CollectorsConfig{
			Global: GlobalCollectorConfig{
				DefaultInterval:     30 * time.Second,
				DefaultTimeout:      10 * time.Second,
				MaxConcurrency:      5,
				ErrorThreshold:      5,
				RecoveryDelay:       60 * time.Second,
				GracefulDegradation: true,
				BatchProcessing: BatchProcessingConfig{
					Enabled:           true,
					MaxBatchSize:      100,
					MaxBatchWait:      5 * time.Second,
					FlushInterval:     30 * time.Second,
					MaxConcurrency:    4,
					EnableCompression: false,
					RetryAttempts:     3,
					RetryDelay:        time.Second,
					MetricBatching: MetricBatchingConfig{
						EnableSampling:    false,
						EnableAggregation: true,
						AggregationWindow: 30 * time.Second,
						MaxMetricAge:      5 * time.Minute,
					},
					EntityBatching: map[string]EntityBatchConfig{
						"jobs": {
							Enabled:      true,
							MaxBatchSize: 200,
							MaxBatchWait: 3 * time.Second,
							Priority:     10,
						},
						"nodes": {
							Enabled:      true,
							MaxBatchSize: 100,
							MaxBatchWait: 5 * time.Second,
							Priority:     5,
						},
						"partitions": {
							Enabled:      true,
							MaxBatchSize: 50,
							MaxBatchWait: 10 * time.Second,
							Priority:     3,
						},
					},
				},
			},
			Cluster: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true, // Collect all metrics by default
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Nodes: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Jobs: CollectorConfig{
				Enabled:  true,
				Interval: 15 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Users: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Partitions: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Performance: CollectorConfig{
				Enabled:  false, // Disabled - not registered in collector registry (API migration)
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			System: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			QoS: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Reservations: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Licenses: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Shares: CollectorConfig{
				Enabled:  true,
				Interval: 120 * time.Second, // Less frequent as fairshare changes slowly
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Diagnostics: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			TRES: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			WCKeys: CollectorConfig{
				Enabled:  false,             // Disabled by default as not all clusters use WCKeys
				Interval: 300 * time.Second, // 5 minutes
				Timeout:  10 * time.Second,
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Clusters: CollectorConfig{
				Enabled:  true,
				Interval: 300 * time.Second, // 5 minutes - cluster info changes rarely
				Timeout:  10 * time.Second,
				Filters: FilterConfig{
					Metrics: MetricFilterConfig{
						EnableAll: true,
					},
				},
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			CollectionTimeout: 30 * time.Second,
			Degradation: DegradationConfig{
				Enabled:          true,
				MaxFailures:      3,
				ResetTimeout:     5 * time.Minute,
				UseCachedMetrics: true,
				CacheTTL:         10 * time.Minute,
			},
		},
		Logging: LoggingConfig{
			Level:        "info",
			Format:       "json",
			Output:       "stdout",
			SuppressHTTP: false,
		},
		Metrics: MetricsConfig{
			Namespace:  "slurm",
			Subsystem:  "exporter",
			MaxAge:     5 * time.Minute,
			AgeBuckets: 5,
			Registry: RegistryConfig{
				EnableGoCollector:      true,
				EnableProcessCollector: true,
				EnableBuildInfo:        true,
			},
			Cardinality: CardinalityConfig{
				MaxSeries:    10000,
				MaxLabels:    100,
				MaxLabelSize: 1024,
				WarnLimit:    8000,
			},
		},
		Observability: ObservabilityConfig{
			PerformanceMonitoring: PerformanceMonitoringConfig{
				Enabled:           true,
				Interval:          30 * time.Second,
				MemoryThreshold:   100 * 1024 * 1024, // 100MB
				CPUThreshold:      80.0,              // 80%
				CardinalityUpdate: 60 * time.Second,  // 1 minute
			},
			Tracing: TracingConfig{
				Enabled:    false, // Disabled by default for performance
				SampleRate: 0.01,  // 1% sampling when enabled
				Endpoint:   "localhost:4317",
				Insecure:   true,
			},
			AdaptiveCollection: AdaptiveCollectionConfig{
				Enabled:      true,
				MinInterval:  30 * time.Second,
				MaxInterval:  5 * time.Minute,
				BaseInterval: 1 * time.Minute,
				ScoreWindow:  10 * time.Minute,
			},
			SmartFiltering: SmartFilteringConfig{
				Enabled:        true,
				NoiseThreshold: 0.8,
				CacheSize:      10000,
				LearningWindow: 100,
				VarianceLimit:  0.8,
				CorrelationMin: 0.2,
			},
			CircuitBreaker: CircuitBreakerConfig{
				Enabled:          true,
				FailureThreshold: 5,
				ResetTimeout:     30 * time.Second,
				HalfOpenRequests: 3,
			},
			Health: HealthConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
				Timeout:  10 * time.Second,
				Checks:   []string{"slurm_api", "collectors", "memory"},
				Endpoints: HealthEndpointsConfig{
					Enabled:   true,
					Path:      "/health",
					ReadyPath: "/ready",
					LivePath:  "/live",
				},
			},
			Debug: DebugConfig{
				Enabled:     false, // Enable only when needed
				RequireAuth: true,
				Username:    "debug",
				Password:    "",
				EnabledEndpoints: []string{
					"collectors", "tracing", "patterns", "scheduler",
					"health", "runtime", "config", "cache",
				},
			},
			Caching: CachingConfig{
				Intelligent:     true,
				BaseTTL:         1 * time.Minute,
				MaxEntries:      50000,
				CleanupInterval: 5 * time.Minute,
				ChangeTracking:  true,
				AdaptiveTTL: AdaptiveTTLConfig{
					Enabled:           true,
					MinTTL:            30 * time.Second,
					MaxTTL:            30 * time.Minute,
					StabilityWindow:   10 * time.Minute,
					VarianceThreshold: 0.1,
					ChangeThreshold:   0.05,
					ExtensionFactor:   2.0,
					ReductionFactor:   0.5,
				},
			},
			BatchProcessing: BatchProcessingConfig{
				Enabled:        true,
				MaxBatchSize:   1000,
				MaxBatchWait:   5 * time.Second,
				FlushInterval:  10 * time.Second,
				MaxConcurrency: 4,
			},
		},
	}
}

// Load loads configuration from a file.
func Load(filename string) (*Config, error) {
	// Start with default configuration
	cfg := Default()

	// If no file specified, just apply env overrides and return
	if filename == "" {
		// Apply environment variable overrides
		if err := cfg.ApplyEnvOverrides(); err != nil {
			return cfg, fmt.Errorf("failed to apply environment overrides: %w", err)
		}

		// Validate configuration with enhanced validation
		if err := cfg.ValidateEnhanced(); err != nil {
			return cfg, fmt.Errorf("configuration validation failed: %w", err)
		}

		return cfg, nil
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return cfg, fmt.Errorf("configuration file %s does not exist", filename)
	}

	// Read file content
	data, err := os.ReadFile(filename)
	if err != nil {
		return cfg, fmt.Errorf("failed to read configuration file %s: %w", filename, err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse YAML configuration: %w", err)
	}

	// Apply environment variable overrides
	if err := cfg.ApplyEnvOverrides(); err != nil {
		return cfg, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Validate configuration with enhanced validation
	if err := cfg.ValidateEnhanced(); err != nil {
		return cfg, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server configuration: %w", err)
	}

	if err := c.SLURM.Validate(); err != nil {
		return fmt.Errorf("SLURM configuration: %w", err)
	}

	if err := c.Collectors.Validate(); err != nil {
		return fmt.Errorf("collectors configuration: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging configuration: %w", err)
	}

	if err := c.Metrics.Validate(); err != nil {
		return fmt.Errorf("metrics configuration: %w", err)
	}

	if err := c.Observability.Validate(); err != nil {
		return fmt.Errorf("observability configuration: %w", err)
	}

	return nil
}

// Validate validates the server configuration.
func (s *ServerConfig) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("server.address cannot be empty (example: ':10341' or '0.0.0.0:10341')")
	}

	if s.MetricsPath == "" {
		return fmt.Errorf("server.metrics_path cannot be empty (default: '/metrics')")
	}

	if s.HealthPath == "" {
		return fmt.Errorf("server.health_path cannot be empty (default: '/health')")
	}

	if s.ReadyPath == "" {
		return fmt.Errorf("server.ready_path cannot be empty (default: '/ready')")
	}

	if s.Timeout <= 0 {
		return fmt.Errorf("server.timeout must be positive, got '%v' (example: '30s', '1m')", s.Timeout)
	}

	if s.ReadTimeout <= 0 {
		return fmt.Errorf("server.read_timeout must be positive, got '%v' (example: '30s')", s.ReadTimeout)
	}

	if s.WriteTimeout <= 0 {
		return fmt.Errorf("server.write_timeout must be positive, got '%v' (example: '30s')", s.WriteTimeout)
	}

	if s.IdleTimeout <= 0 {
		return fmt.Errorf("server.idle_timeout must be positive, got '%v' (example: '60s')", s.IdleTimeout)
	}

	if s.MaxRequestSize <= 0 {
		return fmt.Errorf("server.max_request_size must be positive, got %d (example: 1048576 for 1MB)", s.MaxRequestSize)
	}

	// Validate TLS configuration
	if s.TLS.Enabled {
		if s.TLS.CertFile == "" {
			return fmt.Errorf("server.tls.cert_file must be specified when TLS is enabled (set server.tls.enabled=false to disable TLS)")
		}
		if s.TLS.KeyFile == "" {
			return fmt.Errorf("server.tls.key_file must be specified when TLS is enabled (set server.tls.enabled=false to disable TLS)")
		}

		// Check if TLS files exist
		if _, err := os.Stat(s.TLS.CertFile); os.IsNotExist(err) {
			return fmt.Errorf("server.tls.cert_file '%s' does not exist", s.TLS.CertFile)
		}
		if _, err := os.Stat(s.TLS.KeyFile); os.IsNotExist(err) {
			return fmt.Errorf("server.tls.key_file '%s' does not exist", s.TLS.KeyFile)
		}
	}

	// Validate basic auth configuration
	if s.BasicAuth.Enabled {
		if s.BasicAuth.Username == "" {
			return fmt.Errorf("server.basic_auth.username must be specified when basic auth is enabled (set server.basic_auth.enabled=false to disable)")
		}
		if s.BasicAuth.Password == "" {
			return fmt.Errorf("server.basic_auth.password must be specified when basic auth is enabled (use environment variable SLURM_EXPORTER_SERVER_BASIC_AUTH_PASSWORD)")
		}
	}

	return nil
}

// Validate validates the SLURM configuration.
func (s *SLURMConfig) Validate() error {
	if err := validateURL(s.BaseURL, "slurm.base_url (example: 'https://slurm.example.com:6820' or use environment variable SLURM_EXPORTER_SLURM_BASE_URL)"); err != nil {
		return err
	}

	// Apply default API version if empty
	if s.APIVersion == "" {
		s.APIVersion = "v0.0.44"
	}

	if err := validateAPIVersion(s.APIVersion); err != nil {
		return fmt.Errorf("slurm.api_version: %w", err)
	}

	if s.Timeout <= 0 {
		return fmt.Errorf("slurm.timeout must be positive, got '%v' (example: '30s', '1m')", s.Timeout)
	}

	if s.RetryAttempts < 0 {
		return fmt.Errorf("slurm.retry_attempts cannot be negative, got %d (use 0 to disable retries)", s.RetryAttempts)
	}

	if s.RetryDelay <= 0 {
		return fmt.Errorf("slurm.retry_delay must be positive, got '%v' (example: '5s', '10s')", s.RetryDelay)
	}

	if err := s.Auth.Validate(); err != nil {
		return fmt.Errorf("auth configuration: %w", err)
	}

	if s.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("slurm.rate_limit.requests_per_second must be positive, got %.2f (example: 10.0)", s.RateLimit.RequestsPerSecond)
	}

	if s.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("slurm.rate_limit.burst_size must be positive, got %d (example: 20)", s.RateLimit.BurstSize)
	}

	return nil
}

// validateFileExists checks if a file exists
func validateFileExists(path, field string) error {
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Don't expose the actual file path in error messages (security)
			return fmt.Errorf("%s configuration points to a file that does not exist", field)
		}
	}
	return nil
}

// validateJWTAuth validates JWT authentication configuration
func (a *AuthConfig) validateJWTAuth() error {
	if a.Token == "" && a.TokenFile == "" {
		return fmt.Errorf("slurm.auth.token or slurm.auth.token_file must be specified when using JWT auth (set slurm.auth.type='none' to disable authentication)")
	}
	return validateFileExists(a.TokenFile, "slurm.auth.token_file")
}

// validateBasicAuth validates basic authentication configuration
func (a *AuthConfig) validateBasicAuth() error {
	if a.Username == "" {
		return fmt.Errorf("slurm.auth.username must be specified when using basic auth (current type: '%s')", a.Type)
	}
	if a.Password == "" && a.PasswordFile == "" {
		return fmt.Errorf("slurm.auth.password or slurm.auth.password_file must be specified when using basic auth (use environment variable SLURM_EXPORTER_SLURM_AUTH_PASSWORD)")
	}
	return validateFileExists(a.PasswordFile, "slurm.auth.password_file")
}

// validateAPIKeyAuth validates API key authentication configuration
func (a *AuthConfig) validateAPIKeyAuth() error {
	if a.APIKey == "" && a.APIKeyFile == "" {
		return fmt.Errorf("slurm.auth.api_key or slurm.auth.api_key_file must be specified when using API key auth (use environment variable SLURM_EXPORTER_SLURM_AUTH_API_KEY)")
	}
	return validateFileExists(a.APIKeyFile, "slurm.auth.api_key_file")
}

// Validate validates the auth configuration.
func (a *AuthConfig) Validate() error {
	switch a.Type {
	case AuthTypeNone:
		// No validation needed
	case AuthTypeJWT:
		if err := a.validateJWTAuth(); err != nil {
			return err
		}
	case "basic":
		if err := a.validateBasicAuth(); err != nil {
			return err
		}
	case "apikey":
		if err := a.validateAPIKeyAuth(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported auth type: '%s' (supported types: 'none', 'jwt', 'basic', 'apikey')", a.Type)
	}

	return nil
}

// Validate validates the collectors configuration.
func (c *CollectorsConfig) Validate() error {
	if c.Global.DefaultInterval <= 0 {
		return fmt.Errorf("collectors.global.default_interval must be positive, got '%v' (example: '30s', '1m')", c.Global.DefaultInterval)
	}

	if c.Global.DefaultTimeout <= 0 {
		return fmt.Errorf("collectors.global.default_timeout must be positive, got '%v' (example: '30s')", c.Global.DefaultTimeout)
	}

	if c.Global.MaxConcurrency <= 0 {
		return fmt.Errorf("collectors.global.max_concurrency must be positive, got %d (example: 5)", c.Global.MaxConcurrency)
	}

	if c.Global.ErrorThreshold < 0 {
		return fmt.Errorf("collectors.global.error_threshold cannot be negative, got %d (use 0 to disable error threshold)", c.Global.ErrorThreshold)
	}

	if c.Global.RecoveryDelay <= 0 {
		return fmt.Errorf("collectors.global.recovery_delay must be positive, got '%v' (example: '30s', '1m')", c.Global.RecoveryDelay)
	}

	// Validate individual collectors
	collectors := []struct {
		name      string
		collector CollectorConfig
	}{
		{"cluster", c.Cluster},
		{"nodes", c.Nodes},
		{"jobs", c.Jobs},
		{"users", c.Users},
		{"partitions", c.Partitions},
		{"performance", c.Performance},
		{"system", c.System},
		{"qos", c.QoS},
		{"reservations", c.Reservations},
	}

	for _, col := range collectors {
		if err := col.collector.Validate(); err != nil {
			return fmt.Errorf("collectors.%s: %w", col.name, err)
		}
	}

	// Validate degradation config
	if err := c.Degradation.Validate(); err != nil {
		return fmt.Errorf("collectors.degradation: %w", err)
	}

	return nil
}

// Validate validates the collector configuration.
func (c *CollectorConfig) Validate() error {
	if c.Enabled {
		if c.Interval <= 0 {
			return fmt.Errorf("interval must be positive when collector is enabled, got '%v' (example: '30s', '1m')", c.Interval)
		}

		if c.Timeout <= 0 {
			return fmt.Errorf("timeout must be positive when collector is enabled, got '%v' (example: '30s')", c.Timeout)
		}

		if c.MaxConcurrency < 0 {
			return fmt.Errorf("max_concurrency cannot be negative, got %d (use 0 for unlimited concurrency)", c.MaxConcurrency)
		}

		if err := c.ErrorHandling.Validate(); err != nil {
			return fmt.Errorf("error_handling: %w", err)
		}
	}

	return nil
}

// Validate validates the error handling configuration.
func (e *ErrorHandlingConfig) Validate() error {
	if e.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative, got %d (use 0 to disable retries)", e.MaxRetries)
	}

	if e.RetryDelay <= 0 {
		return fmt.Errorf("retry_delay must be positive, got '%v' (example: '5s', '10s')", e.RetryDelay)
	}

	if e.BackoffFactor <= 0 {
		return fmt.Errorf("backoff_factor must be positive, got %.2f (example: 1.5, 2.0)", e.BackoffFactor)
	}

	if e.MaxRetryDelay <= 0 {
		return fmt.Errorf("max_retry_delay must be positive, got '%v' (example: '30s', '1m')", e.MaxRetryDelay)
	}

	if e.MaxRetryDelay < e.RetryDelay {
		return fmt.Errorf("max_retry_delay (%v) must be >= retry_delay (%v)", e.MaxRetryDelay, e.RetryDelay)
	}

	return nil
}

// Validate validates the degradation configuration.
func (d *DegradationConfig) Validate() error {
	if d.Enabled {
		if d.MaxFailures <= 0 {
			return fmt.Errorf("max_failures must be positive when degradation is enabled, got %d (example: 3, 5)", d.MaxFailures)
		}

		if d.ResetTimeout <= 0 {
			return fmt.Errorf("reset_timeout must be positive when degradation is enabled, got '%v' (example: '5m', '10m')", d.ResetTimeout)
		}

		if d.UseCachedMetrics && d.CacheTTL <= 0 {
			return fmt.Errorf("cache_ttl must be positive when cached metrics are enabled, got '%v' (example: '10m', '30m')", d.CacheTTL)
		}
	}

	return nil
}

// Validate validates the logging configuration.
func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[l.Level] {
		return fmt.Errorf("logging.level: invalid value '%s' (supported: debug, info, warn, error)", l.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[l.Format] {
		return fmt.Errorf("logging.format: invalid value '%s' (supported: json, text)", l.Format)
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}

	if !validOutputs[l.Output] {
		return fmt.Errorf("logging.output: invalid value '%s' (supported: stdout, stderr, file)", l.Output)
	}

	if l.Output == "file" && l.File == "" {
		return fmt.Errorf("logging.file must be specified when logging.output='file' (example: '/var/log/slurm-exporter.log')")
	}

	if l.MaxSize < 0 {
		return fmt.Errorf("max size cannot be negative")
	}

	if l.MaxAge < 0 {
		return fmt.Errorf("max age cannot be negative")
	}

	if l.MaxBackups < 0 {
		return fmt.Errorf("max backups cannot be negative")
	}

	return nil
}

// Validate validates the metrics configuration.
func (m *MetricsConfig) Validate() error {
	if m.Namespace == "" {
		return fmt.Errorf("metrics.namespace cannot be empty (example: 'slurm')")
	}

	if m.MaxAge <= 0 {
		return fmt.Errorf("metrics.max_age must be positive, got '%v' (example: '5m', '1h')", m.MaxAge)
	}

	if m.AgeBuckets <= 0 {
		return fmt.Errorf("metrics.age_buckets must be positive, got %d (example: 5, 10)", m.AgeBuckets)
	}

	if err := m.Cardinality.Validate(); err != nil {
		return fmt.Errorf("metrics.cardinality: %w", err)
	}

	return nil
}

// Validate validates the cardinality configuration.
func (c *CardinalityConfig) Validate() error {
	if c.MaxSeries <= 0 {
		return fmt.Errorf("max_series must be positive, got %d (example: 10000, 50000)", c.MaxSeries)
	}

	if c.MaxLabels <= 0 {
		return fmt.Errorf("max_labels must be positive, got %d (example: 10, 20)", c.MaxLabels)
	}

	if c.MaxLabelSize <= 0 {
		return fmt.Errorf("max_label_size must be positive, got %d (example: 128, 256)", c.MaxLabelSize)
	}

	if c.WarnLimit < 0 {
		return fmt.Errorf("warn limit cannot be negative")
	}

	if c.WarnLimit > c.MaxSeries {
		return fmt.Errorf("warn limit cannot be greater than max series")
	}

	return nil
}

// ApplyEnvOverrides applies environment variable overrides to the configuration.
// Environment variables follow the pattern: SLURM_EXPORTER_<SECTION>_<FIELD>
// For nested fields, use underscores: SLURM_EXPORTER_SLURM_AUTH_TYPE
func (c *Config) ApplyEnvOverrides() error {
	prefix := "SLURM_EXPORTER_"

	// Server configuration overrides
	if err := c.applyServerEnvOverrides(prefix + "SERVER_"); err != nil {
		return fmt.Errorf("server config overrides: %w", err)
	}

	// SLURM configuration overrides
	if err := c.applySLURMEnvOverrides(prefix + "SLURM_"); err != nil {
		return fmt.Errorf("SLURM config overrides: %w", err)
	}

	// Collectors configuration overrides
	if err := c.applyCollectorsEnvOverrides(prefix + "COLLECTORS_"); err != nil {
		return fmt.Errorf("collectors config overrides: %w", err)
	}

	// Logging configuration overrides
	if err := c.applyLoggingEnvOverrides(prefix + "LOGGING_"); err != nil {
		return fmt.Errorf("logging config overrides: %w", err)
	}

	// Metrics configuration overrides
	if err := c.applyMetricsEnvOverrides(prefix + "METRICS_"); err != nil {
		return fmt.Errorf("metrics config overrides: %w", err)
	}

	return nil
}

// envString gets a string environment variable
func envString(key string, setter func(string)) {
	if val := os.Getenv(key); val != "" {
		setter(val)
	}
}

// envDuration gets a duration environment variable
func envDuration(key string, setter func(time.Duration) error) error {
	if val := os.Getenv(key); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid duration for %s: %w", key, err)
		}
		return setter(duration)
	}
	return nil
}

// envBool gets a boolean environment variable
func envBool(key string, setter func(bool) error) error {
	if val := os.Getenv(key); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid boolean for %s: %w", key, err)
		}
		return setter(enabled)
	}
	return nil
}

// envInt gets an integer environment variable
func envInt(key string, setter func(int) error) error {
	if val := os.Getenv(key); val != "" {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		return setter(intVal)
	}
	return nil
}

// envInt64 gets an int64 environment variable
func envInt64(key string, setter func(int64) error) error {
	if val := os.Getenv(key); val != "" {
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer for %s: %w", key, err)
		}
		return setter(intVal)
	}
	return nil
}

// envFloat64 gets a float64 environment variable
func envFloat64(key string, setter func(float64) error) error {
	if val := os.Getenv(key); val != "" {
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid float for %s: %w", key, err)
		}
		return setter(floatVal)
	}
	return nil
}

// applyServerEnvOverrides applies server-specific environment overrides.
func (c *Config) applyServerEnvOverrides(prefix string) error {
	// String overrides
	envString(prefix+"ADDRESS", func(v string) { c.Server.Address = v })
	envString(prefix+"METRICS_PATH", func(v string) { c.Server.MetricsPath = v })
	envString(prefix+"HEALTH_PATH", func(v string) { c.Server.HealthPath = v })
	envString(prefix+"READY_PATH", func(v string) { c.Server.ReadyPath = v })

	// Duration overrides
	if err := envDuration(prefix+"TIMEOUT", func(v time.Duration) error { c.Server.Timeout = v; return nil }); err != nil {
		return err
	}
	if err := envDuration(prefix+"READ_TIMEOUT", func(v time.Duration) error { c.Server.ReadTimeout = v; return nil }); err != nil {
		return err
	}
	if err := envDuration(prefix+"WRITE_TIMEOUT", func(v time.Duration) error { c.Server.WriteTimeout = v; return nil }); err != nil {
		return err
	}
	if err := envDuration(prefix+"IDLE_TIMEOUT", func(v time.Duration) error { c.Server.IdleTimeout = v; return nil }); err != nil {
		return err
	}

	// Integer overrides
	if err := envInt64(prefix+"MAX_REQUEST_SIZE", func(v int64) error { c.Server.MaxRequestSize = v; return nil }); err != nil {
		return err
	}

	// TLS configuration
	if err := envBool(prefix+"TLS_ENABLED", func(v bool) error { c.Server.TLS.Enabled = v; return nil }); err != nil {
		return err
	}
	envString(prefix+"TLS_CERT_FILE", func(v string) { c.Server.TLS.CertFile = v })
	envString(prefix+"TLS_KEY_FILE", func(v string) { c.Server.TLS.KeyFile = v })

	// Basic Auth configuration
	if err := envBool(prefix+"BASIC_AUTH_ENABLED", func(v bool) error { c.Server.BasicAuth.Enabled = v; return nil }); err != nil {
		return err
	}
	envString(prefix+"BASIC_AUTH_USERNAME", func(v string) { c.Server.BasicAuth.Username = v })
	envString(prefix+"BASIC_AUTH_PASSWORD", func(v string) { c.Server.BasicAuth.Password = v })

	return nil
}

// applySLURMEnvOverrides applies SLURM-specific environment overrides.
func (c *Config) applySLURMEnvOverrides(prefix string) error {
	// String overrides
	envString(prefix+"BASE_URL", func(v string) { c.SLURM.BaseURL = v })
	envString(prefix+"API_VERSION", func(v string) { c.SLURM.APIVersion = v })

	// Boolean overrides
	if err := envBool(prefix+"USE_ADAPTERS", func(v bool) error { c.SLURM.UseAdapters = v; return nil }); err != nil {
		return err
	}

	// Duration overrides
	if err := envDuration(prefix+"TIMEOUT", func(v time.Duration) error { c.SLURM.Timeout = v; return nil }); err != nil {
		return err
	}
	if err := envDuration(prefix+"RETRY_DELAY", func(v time.Duration) error { c.SLURM.RetryDelay = v; return nil }); err != nil {
		return err
	}

	// Integer overrides
	if err := envInt(prefix+"RETRY_ATTEMPTS", func(v int) error { c.SLURM.RetryAttempts = v; return nil }); err != nil {
		return err
	}

	// Authentication configuration
	envString(prefix+"AUTH_TYPE", func(v string) { c.SLURM.Auth.Type = v })
	envString(prefix+"AUTH_TOKEN", func(v string) { c.SLURM.Auth.Token = v })
	envString(prefix+"AUTH_TOKEN_FILE", func(v string) { c.SLURM.Auth.TokenFile = v })
	envString(prefix+"AUTH_USERNAME", func(v string) { c.SLURM.Auth.Username = v })
	envString(prefix+"AUTH_PASSWORD", func(v string) { c.SLURM.Auth.Password = v })
	envString(prefix+"AUTH_PASSWORD_FILE", func(v string) { c.SLURM.Auth.PasswordFile = v })
	envString(prefix+"AUTH_API_KEY", func(v string) { c.SLURM.Auth.APIKey = v })
	envString(prefix+"AUTH_API_KEY_FILE", func(v string) { c.SLURM.Auth.APIKeyFile = v })

	// TLS configuration
	if err := envBool(prefix+"TLS_INSECURE_SKIP_VERIFY", func(v bool) error { c.SLURM.TLS.InsecureSkipVerify = v; return nil }); err != nil {
		return err
	}
	envString(prefix+"TLS_CA_CERT_FILE", func(v string) { c.SLURM.TLS.CACertFile = v })
	envString(prefix+"TLS_CLIENT_CERT_FILE", func(v string) { c.SLURM.TLS.ClientCertFile = v })
	envString(prefix+"TLS_CLIENT_KEY_FILE", func(v string) { c.SLURM.TLS.ClientKeyFile = v })

	// Rate limiting configuration
	if err := envFloat64(prefix+"RATE_LIMIT_REQUESTS_PER_SECOND", func(v float64) error { c.SLURM.RateLimit.RequestsPerSecond = v; return nil }); err != nil {
		return err
	}
	if err := envInt(prefix+"RATE_LIMIT_BURST_SIZE", func(v int) error { c.SLURM.RateLimit.BurstSize = v; return nil }); err != nil {
		return err
	}

	return nil
}

// applyCollectorsEnvOverrides applies collectors-specific environment overrides.
func (c *Config) applyCollectorsEnvOverrides(prefix string) error {
	// Global collector settings
	if val := os.Getenv(prefix + "GLOBAL_DEFAULT_INTERVAL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid global default interval: %w", err)
		}
		c.Collectors.Global.DefaultInterval = duration
	}

	if val := os.Getenv(prefix + "GLOBAL_DEFAULT_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid global default timeout: %w", err)
		}
		c.Collectors.Global.DefaultTimeout = duration
	}

	if val := os.Getenv(prefix + "GLOBAL_MAX_CONCURRENCY"); val != "" {
		concurrency, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid global max concurrency: %w", err)
		}
		c.Collectors.Global.MaxConcurrency = concurrency
	}

	// Individual collector overrides
	collectors := map[string]*CollectorConfig{
		"CLUSTER":      &c.Collectors.Cluster,
		"NODES":        &c.Collectors.Nodes,
		"JOBS":         &c.Collectors.Jobs,
		"USERS":        &c.Collectors.Users,
		"PARTITIONS":   &c.Collectors.Partitions,
		"PERFORMANCE":  &c.Collectors.Performance,
		"SYSTEM":       &c.Collectors.System,
		"QOS":          &c.Collectors.QoS,
		"RESERVATIONS": &c.Collectors.Reservations,
	}

	for name, collector := range collectors {
		if err := c.applyCollectorEnvOverrides(prefix+name+"_", collector); err != nil {
			return fmt.Errorf("%s collector overrides: %w", strings.ToLower(name), err)
		}
	}

	return nil
}

// applyCollectorEnvOverrides applies environment overrides for a single collector.
func (c *Config) applyCollectorEnvOverrides(prefix string, collector *CollectorConfig) error {
	if val := os.Getenv(prefix + "ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid enabled value: %w", err)
		}
		collector.Enabled = enabled
	}

	if val := os.Getenv(prefix + "INTERVAL"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}
		collector.Interval = duration
	}

	if val := os.Getenv(prefix + "TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		collector.Timeout = duration
	}

	if val := os.Getenv(prefix + "MAX_CONCURRENCY"); val != "" {
		concurrency, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max concurrency: %w", err)
		}
		collector.MaxConcurrency = concurrency
	}

	return nil
}

// applyLoggingEnvOverrides applies logging-specific environment overrides.
func (c *Config) applyLoggingEnvOverrides(prefix string) error {
	if val := os.Getenv(prefix + "LEVEL"); val != "" {
		c.Logging.Level = val
	}

	if val := os.Getenv(prefix + "FORMAT"); val != "" {
		c.Logging.Format = val
	}

	if val := os.Getenv(prefix + "OUTPUT"); val != "" {
		c.Logging.Output = val
	}

	if val := os.Getenv(prefix + "FILE"); val != "" {
		c.Logging.File = val
	}

	if val := os.Getenv(prefix + "MAX_SIZE"); val != "" {
		size, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max size: %w", err)
		}
		c.Logging.MaxSize = size
	}

	if val := os.Getenv(prefix + "MAX_AGE"); val != "" {
		age, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max age: %w", err)
		}
		c.Logging.MaxAge = age
	}

	if val := os.Getenv(prefix + "MAX_BACKUPS"); val != "" {
		backups, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max backups: %w", err)
		}
		c.Logging.MaxBackups = backups
	}

	if val := os.Getenv(prefix + "COMPRESS"); val != "" {
		compress, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid compress value: %w", err)
		}
		c.Logging.Compress = compress
	}

	if val := os.Getenv(prefix + "SUPPRESS_HTTP"); val != "" {
		suppress, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid suppress HTTP value: %w", err)
		}
		c.Logging.SuppressHTTP = suppress
	}

	return nil
}

// applyMetricsEnvOverrides applies metrics-specific environment overrides.
func (c *Config) applyMetricsEnvOverrides(prefix string) error {
	// String overrides
	envString(prefix+"NAMESPACE", func(v string) { c.Metrics.Namespace = v })
	envString(prefix+"SUBSYSTEM", func(v string) { c.Metrics.Subsystem = v })

	// Duration overrides
	if err := envDuration(prefix+"MAX_AGE", func(v time.Duration) error { c.Metrics.MaxAge = v; return nil }); err != nil {
		return err
	}

	// Integer overrides
	if err := envInt(prefix+"AGE_BUCKETS", func(v int) error { c.Metrics.AgeBuckets = v; return nil }); err != nil {
		return err
	}

	// Registry configuration
	if err := envBool(prefix+"REGISTRY_ENABLE_GO_COLLECTOR", func(v bool) error { c.Metrics.Registry.EnableGoCollector = v; return nil }); err != nil {
		return err
	}
	if err := envBool(prefix+"REGISTRY_ENABLE_PROCESS_COLLECTOR", func(v bool) error { c.Metrics.Registry.EnableProcessCollector = v; return nil }); err != nil {
		return err
	}
	if err := envBool(prefix+"REGISTRY_ENABLE_BUILD_INFO", func(v bool) error { c.Metrics.Registry.EnableBuildInfo = v; return nil }); err != nil {
		return err
	}

	// Cardinality configuration
	if err := envInt(prefix+"CARDINALITY_MAX_SERIES", func(v int) error { c.Metrics.Cardinality.MaxSeries = v; return nil }); err != nil {
		return err
	}
	if err := envInt(prefix+"CARDINALITY_MAX_LABELS", func(v int) error { c.Metrics.Cardinality.MaxLabels = v; return nil }); err != nil {
		return err
	}
	if err := envInt(prefix+"CARDINALITY_MAX_LABEL_SIZE", func(v int) error { c.Metrics.Cardinality.MaxLabelSize = v; return nil }); err != nil {
		return err
	}
	if err := envInt(prefix+"CARDINALITY_WARN_LIMIT", func(v int) error { c.Metrics.Cardinality.WarnLimit = v; return nil }); err != nil {
		return err
	}

	return nil
}

// Reloader provides configuration hot-reloading capabilities using file watchers
type Reloader struct {
	configFile string
	watcher    *fsnotify.Watcher
	callback   func(*Config) error
	config     *Config
	mu         sync.RWMutex
}

// NewReloader creates a new configuration reloader
func NewReloader(configFile string, initialConfig *Config, callback func(*Config) error) (*Reloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	r := &Reloader{
		configFile: configFile,
		watcher:    watcher,
		callback:   callback,
		config:     initialConfig,
	}

	// Add the configuration file to the watcher
	err = watcher.Add(configFile)
	if err != nil {
		_ = watcher.Close()
		return nil, fmt.Errorf("failed to watch config file %s: %w", configFile, err)
	}

	return r, nil
}

// Start begins watching for configuration changes
func (r *Reloader) Start(ctx context.Context) error {
	for {
		select {
		case event, ok := <-r.watcher.Events:
			if !ok {
				return fmt.Errorf("watcher events channel closed")
			}

			// Only reload on write events to the config file
			if event.Has(fsnotify.Write) && event.Name == r.configFile {
				if err := r.reload(); err != nil {
					logrus.WithError(err).Error("Failed to reload configuration")
					continue
				}
				logrus.Info("Configuration reloaded successfully")
			}

		case err, ok := <-r.watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			logrus.WithError(err).Error("File watcher error")

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// reload loads the updated configuration and calls the callback
func (r *Reloader) reload() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Load the new configuration
	newConfig, err := Load(r.configFile)
	if err != nil {
		return fmt.Errorf("failed to load updated config: %w", err)
	}

	// Validate the new configuration
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("updated config validation failed: %w", err)
	}

	// Call the callback with the new configuration
	if r.callback != nil {
		if err := r.callback(newConfig); err != nil {
			return fmt.Errorf("config update callback failed: %w", err)
		}
	}

	// Update the stored configuration
	r.config = newConfig
	return nil
}

// GetConfig returns the current configuration (thread-safe)
func (r *Reloader) GetConfig() *Config {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

// Close stops the file watcher and releases resources
func (r *Reloader) Close() error {
	return r.watcher.Close()
}

// validateURL validates that a URL is properly formatted
func validateURL(url, field string) error {
	if url == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s must be a valid HTTP/HTTPS URL starting with 'http://' or 'https://', got: %s", field, url)
	}

	return nil
}

// validateAPIVersion validates that the SLURM API version is supported
func validateAPIVersion(version string) error {
	supportedVersions := []string{"v0.0.40", "v0.0.41", "v0.0.42", "v0.0.43", "v0.0.44"}

	// Allow empty version - will use default
	if version == "" {
		return nil
	}

	for _, v := range supportedVersions {
		if version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported SLURM API version '%s' (supported: %s)", version, strings.Join(supportedVersions, ", "))
}

// ObservabilityConfig holds observability configuration for the exporter
type ObservabilityConfig struct {
	// Performance monitoring
	PerformanceMonitoring PerformanceMonitoringConfig `yaml:"performance_monitoring"`

	// Collection tracing
	Tracing TracingConfig `yaml:"tracing"`

	// Adaptive collection
	AdaptiveCollection AdaptiveCollectionConfig `yaml:"adaptive_collection"`

	// Smart filtering
	SmartFiltering SmartFilteringConfig `yaml:"smart_filtering"`

	// Circuit breaker
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`

	// Health checks
	Health HealthConfig `yaml:"health"`

	// Debug endpoints
	Debug DebugConfig `yaml:"debug"`

	// Intelligent caching
	Caching CachingConfig `yaml:"caching"`

	// Batch processing
	BatchProcessing BatchProcessingConfig `yaml:"batch_processing"`
}

// PerformanceMonitoringConfig holds performance monitoring configuration
type PerformanceMonitoringConfig struct {
	Enabled           bool          `yaml:"enabled"`
	Interval          time.Duration `yaml:"interval"`
	MemoryThreshold   int64         `yaml:"memory_threshold"`   // bytes
	CPUThreshold      float64       `yaml:"cpu_threshold"`      // percentage
	CardinalityUpdate time.Duration `yaml:"cardinality_update"` // interval for cardinality updates
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled    bool    `yaml:"enabled"`
	SampleRate float64 `yaml:"sample_rate"`
	Endpoint   string  `yaml:"endpoint"`
	Insecure   bool    `yaml:"insecure"`
}

// AdaptiveCollectionConfig holds adaptive collection configuration
type AdaptiveCollectionConfig struct {
	Enabled      bool          `yaml:"enabled"`
	MinInterval  time.Duration `yaml:"min_interval"`
	MaxInterval  time.Duration `yaml:"max_interval"`
	BaseInterval time.Duration `yaml:"base_interval"`
	ScoreWindow  time.Duration `yaml:"score_window"`
}

// SmartFilteringConfig holds smart filtering configuration
type SmartFilteringConfig struct {
	Enabled        bool    `yaml:"enabled"`
	NoiseThreshold float64 `yaml:"noise_threshold"`
	CacheSize      int     `yaml:"cache_size"`
	LearningWindow int     `yaml:"learning_window"`
	VarianceLimit  float64 `yaml:"variance_limit"`
	CorrelationMin float64 `yaml:"correlation_min"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `yaml:"enabled"`
	FailureThreshold int           `yaml:"failure_threshold"`
	ResetTimeout     time.Duration `yaml:"reset_timeout"`
	HalfOpenRequests int           `yaml:"half_open_requests"`
}

// HealthConfig holds health monitoring configuration
type HealthConfig struct {
	Enabled   bool                  `yaml:"enabled"`
	Interval  time.Duration         `yaml:"interval"`
	Timeout   time.Duration         `yaml:"timeout"`
	Checks    []string              `yaml:"checks"`
	Endpoints HealthEndpointsConfig `yaml:"endpoints"`
}

// HealthEndpointsConfig holds health endpoint configuration
type HealthEndpointsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Path      string `yaml:"path"`
	ReadyPath string `yaml:"ready_path"`
	LivePath  string `yaml:"live_path"`
}

// DebugConfig holds debug configuration
type DebugConfig struct {
	Enabled          bool     `yaml:"enabled"`
	RequireAuth      bool     `yaml:"require_auth"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
	EnabledEndpoints []string `yaml:"enabled_endpoints"` // List of enabled endpoints
}

// CachingConfig holds intelligent caching configuration
type CachingConfig struct {
	Intelligent     bool          `yaml:"intelligent"`
	BaseTTL         time.Duration `yaml:"base_ttl"`
	MaxEntries      int           `yaml:"max_entries"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
	ChangeTracking  bool          `yaml:"change_tracking"`

	// Adaptive TTL configuration
	AdaptiveTTL AdaptiveTTLConfig `yaml:"adaptive_ttl"`
}

// AdaptiveTTLConfig holds configuration for adaptive TTL calculation
type AdaptiveTTLConfig struct {
	Enabled           bool          `yaml:"enabled"`
	MinTTL            time.Duration `yaml:"min_ttl"`
	MaxTTL            time.Duration `yaml:"max_ttl"`
	StabilityWindow   time.Duration `yaml:"stability_window"`
	VarianceThreshold float64       `yaml:"variance_threshold"`
	ChangeThreshold   float64       `yaml:"change_threshold"`
	ExtensionFactor   float64       `yaml:"extension_factor"`
	ReductionFactor   float64       `yaml:"reduction_factor"`
}

// Validate validates the observability configuration
func (o *ObservabilityConfig) Validate() error {
	if err := o.PerformanceMonitoring.Validate(); err != nil {
		return fmt.Errorf("performance_monitoring: %w", err)
	}

	if err := o.Tracing.Validate(); err != nil {
		return fmt.Errorf("tracing: %w", err)
	}

	if err := o.AdaptiveCollection.Validate(); err != nil {
		return fmt.Errorf("adaptive_collection: %w", err)
	}

	if err := o.SmartFiltering.Validate(); err != nil {
		return fmt.Errorf("smart_filtering: %w", err)
	}

	if err := o.CircuitBreaker.Validate(); err != nil {
		return fmt.Errorf("circuit_breaker: %w", err)
	}

	if err := o.Health.Validate(); err != nil {
		return fmt.Errorf("health: %w", err)
	}

	if err := o.Debug.Validate(); err != nil {
		return fmt.Errorf("debug: %w", err)
	}

	if err := o.Caching.Validate(); err != nil {
		return fmt.Errorf("caching: %w", err)
	}

	if err := o.BatchProcessing.Validate(); err != nil {
		return fmt.Errorf("batch_processing: %w", err)
	}

	return nil
}

// Validate validates the performance monitoring configuration
func (p *PerformanceMonitoringConfig) Validate() error {
	if p.Enabled {
		if p.Interval <= 0 {
			return fmt.Errorf("interval must be positive when enabled, got '%v' (example: '30s', '1m')", p.Interval)
		}

		if p.MemoryThreshold < 0 {
			return fmt.Errorf("memory_threshold cannot be negative, got %d bytes", p.MemoryThreshold)
		}

		if p.CPUThreshold < 0 || p.CPUThreshold > 100 {
			return fmt.Errorf("cpu_threshold must be between 0 and 100, got %.2f", p.CPUThreshold)
		}

		if p.CardinalityUpdate <= 0 {
			return fmt.Errorf("cardinality_update must be positive, got '%v' (example: '60s', '5m')", p.CardinalityUpdate)
		}
	}

	return nil
}

// Validate validates the tracing configuration
func (t *TracingConfig) Validate() error {
	if t.Enabled {
		if t.SampleRate < 0 || t.SampleRate > 1 {
			return fmt.Errorf("sample_rate must be between 0 and 1, got %.4f", t.SampleRate)
		}

		if t.Endpoint == "" {
			return fmt.Errorf("endpoint must be specified when tracing is enabled (example: 'localhost:4317')")
		}
	}

	return nil
}

// Validate validates the adaptive collection configuration
func (a *AdaptiveCollectionConfig) Validate() error {
	if a.Enabled {
		if a.MinInterval <= 0 {
			return fmt.Errorf("min_interval must be positive, got '%v' (example: '30s')", a.MinInterval)
		}

		if a.MaxInterval <= 0 {
			return fmt.Errorf("max_interval must be positive, got '%v' (example: '5m')", a.MaxInterval)
		}

		if a.BaseInterval <= 0 {
			return fmt.Errorf("base_interval must be positive, got '%v' (example: '1m')", a.BaseInterval)
		}

		if a.MinInterval >= a.MaxInterval {
			return fmt.Errorf("min_interval (%v) must be less than max_interval (%v)", a.MinInterval, a.MaxInterval)
		}

		if a.BaseInterval < a.MinInterval || a.BaseInterval > a.MaxInterval {
			return fmt.Errorf("base_interval (%v) must be between min_interval (%v) and max_interval (%v)", a.BaseInterval, a.MinInterval, a.MaxInterval)
		}

		if a.ScoreWindow <= 0 {
			return fmt.Errorf("score_window must be positive, got '%v' (example: '10m')", a.ScoreWindow)
		}
	}

	return nil
}

// Validate validates the smart filtering configuration
func (s *SmartFilteringConfig) Validate() error {
	if s.Enabled {
		if s.NoiseThreshold < 0 || s.NoiseThreshold > 1 {
			return fmt.Errorf("noise_threshold must be between 0 and 1, got %.2f", s.NoiseThreshold)
		}

		if s.CacheSize <= 0 {
			return fmt.Errorf("cache_size must be positive, got %d", s.CacheSize)
		}

		if s.LearningWindow <= 0 {
			return fmt.Errorf("learning_window must be positive, got %d", s.LearningWindow)
		}

		if s.VarianceLimit < 0 || s.VarianceLimit > 1 {
			return fmt.Errorf("variance_limit must be between 0 and 1, got %.2f", s.VarianceLimit)
		}

		if s.CorrelationMin < 0 || s.CorrelationMin > 1 {
			return fmt.Errorf("correlation_min must be between 0 and 1, got %.2f", s.CorrelationMin)
		}
	}

	return nil
}

// Validate validates the circuit breaker configuration
func (c *CircuitBreakerConfig) Validate() error {
	if c.Enabled {
		if c.FailureThreshold <= 0 {
			return fmt.Errorf("failure_threshold must be positive, got %d", c.FailureThreshold)
		}

		if c.ResetTimeout <= 0 {
			return fmt.Errorf("reset_timeout must be positive, got '%v' (example: '30s', '1m')", c.ResetTimeout)
		}

		if c.HalfOpenRequests <= 0 {
			return fmt.Errorf("half_open_requests must be positive, got %d", c.HalfOpenRequests)
		}
	}

	return nil
}

// Validate validates the health configuration
func (h *HealthConfig) Validate() error {
	if h.Enabled {
		if h.Interval <= 0 {
			return fmt.Errorf("interval must be positive, got '%v' (example: '30s')", h.Interval)
		}

		if h.Timeout <= 0 {
			return fmt.Errorf("timeout must be positive, got '%v' (example: '10s')", h.Timeout)
		}

		if h.Timeout >= h.Interval {
			return fmt.Errorf("timeout (%v) must be less than interval (%v)", h.Timeout, h.Interval)
		}

		validChecks := map[string]bool{
			"slurm_api":  true,
			"collectors": true,
			"memory":     true,
			"disk":       true,
			"network":    true,
		}

		for _, check := range h.Checks {
			if !validChecks[check] {
				return fmt.Errorf("invalid health check '%s' (valid: slurm_api, collectors, memory, disk, network)", check)
			}
		}

		if err := h.Endpoints.Validate(); err != nil {
			return fmt.Errorf("endpoints: %w", err)
		}
	}

	return nil
}

// Validate validates the health endpoints configuration
func (h *HealthEndpointsConfig) Validate() error {
	if h.Enabled {
		if h.Path == "" {
			return fmt.Errorf("path cannot be empty when endpoints are enabled")
		}

		if h.ReadyPath == "" {
			return fmt.Errorf("ready_path cannot be empty when endpoints are enabled")
		}

		if h.LivePath == "" {
			return fmt.Errorf("live_path cannot be empty when endpoints are enabled")
		}
	}

	return nil
}

// Validate validates the caching configuration
func (c *CachingConfig) Validate() error {
	if c.Intelligent {
		if c.BaseTTL <= 0 {
			return fmt.Errorf("base_ttl must be positive, got '%v' (example: '1m')", c.BaseTTL)
		}

		if c.MaxEntries <= 0 {
			return fmt.Errorf("max_entries must be positive, got %d", c.MaxEntries)
		}

		if c.CleanupInterval <= 0 {
			return fmt.Errorf("cleanup_interval must be positive, got '%v' (example: '5m')", c.CleanupInterval)
		}

		if err := c.AdaptiveTTL.Validate(); err != nil {
			return fmt.Errorf("adaptive_ttl: %w", err)
		}
	}

	return nil
}

// Validate validates the adaptive TTL configuration
func (a *AdaptiveTTLConfig) Validate() error {
	if a.Enabled {
		if a.MinTTL <= 0 {
			return fmt.Errorf("min_ttl must be positive, got '%v' (example: '30s')", a.MinTTL)
		}

		if a.MaxTTL <= 0 {
			return fmt.Errorf("max_ttl must be positive, got '%v' (example: '30m')", a.MaxTTL)
		}

		if a.MinTTL >= a.MaxTTL {
			return fmt.Errorf("min_ttl (%v) must be less than max_ttl (%v)", a.MinTTL, a.MaxTTL)
		}

		if a.StabilityWindow <= 0 {
			return fmt.Errorf("stability_window must be positive, got '%v' (example: '10m')", a.StabilityWindow)
		}

		if a.VarianceThreshold < 0 || a.VarianceThreshold > 1 {
			return fmt.Errorf("variance_threshold must be between 0 and 1, got %f", a.VarianceThreshold)
		}

		if a.ChangeThreshold < 0 || a.ChangeThreshold > 1 {
			return fmt.Errorf("change_threshold must be between 0 and 1, got %f", a.ChangeThreshold)
		}

		if a.ExtensionFactor <= 1 {
			return fmt.Errorf("extension_factor must be greater than 1, got %f", a.ExtensionFactor)
		}

		if a.ReductionFactor <= 0 || a.ReductionFactor >= 1 {
			return fmt.Errorf("reduction_factor must be between 0 and 1, got %f", a.ReductionFactor)
		}
	}

	return nil
}

// Validate validates the batch processing configuration
func (b *BatchProcessingConfig) Validate() error {
	if b.Enabled {
		if b.MaxBatchSize <= 0 {
			return fmt.Errorf("max_batch_size must be positive, got %d", b.MaxBatchSize)
		}

		if b.MaxBatchWait <= 0 {
			return fmt.Errorf("max_batch_wait must be positive, got '%v' (example: '5s')", b.MaxBatchWait)
		}

		if b.FlushInterval <= 0 {
			return fmt.Errorf("flush_interval must be positive, got '%v' (example: '30s')", b.FlushInterval)
		}

		if b.MaxConcurrency <= 0 {
			return fmt.Errorf("max_concurrency must be positive, got %d", b.MaxConcurrency)
		}
	}

	return nil
}

// Validate validates the debug configuration
func (d *DebugConfig) Validate() error {
	if d.Enabled {
		if d.RequireAuth && (d.Username == "" || d.Password == "") {
			return fmt.Errorf("username and password must be specified when auth is required")
		}

		// Validate enabled endpoints
		validEndpoints := map[string]bool{
			"collectors": true,
			"tracing":    true,
			"patterns":   true,
			"scheduler":  true,
			"health":     true,
			"runtime":    true,
			"config":     true,
		}

		for _, endpoint := range d.EnabledEndpoints {
			if !validEndpoints[endpoint] {
				return fmt.Errorf("invalid endpoint '%s', valid endpoints: collectors, tracing, patterns, scheduler, health, runtime, config", endpoint)
			}
		}
	}

	return nil
}
