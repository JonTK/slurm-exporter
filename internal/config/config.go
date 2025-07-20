package config

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	SLURM      SLURMConfig      `yaml:"slurm"`
	Collectors CollectorsConfig `yaml:"collectors"`
	Logging    LoggingConfig    `yaml:"logging"`
	Metrics    MetricsConfig    `yaml:"metrics"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Address       string        `yaml:"address"`
	MetricsPath   string        `yaml:"metrics_path"`
	HealthPath    string        `yaml:"health_path"`
	ReadyPath     string        `yaml:"ready_path"`
	Timeout       time.Duration `yaml:"timeout"`
	ReadTimeout   time.Duration `yaml:"read_timeout"`
	WriteTimeout  time.Duration `yaml:"write_timeout"`
	IdleTimeout   time.Duration `yaml:"idle_timeout"`
	TLS           TLSConfig     `yaml:"tls"`
	BasicAuth     BasicAuthConfig `yaml:"basic_auth"`
	CORS          CORSConfig    `yaml:"cors"`
	MaxRequestSize int64        `yaml:"max_request_size"`
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	MinVersion string `yaml:"min_version"`
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
	BaseURL       string        `yaml:"base_url"`
	APIVersion    string        `yaml:"api_version"`
	Auth          AuthConfig    `yaml:"auth"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	TLS           SLURMTLSConfig `yaml:"tls"`
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
	Type         string            `yaml:"type"`         // jwt, basic, apikey, none
	Token        string            `yaml:"token"`        // For JWT
	TokenFile    string            `yaml:"token_file"`   // For JWT from file
	Username     string            `yaml:"username"`     // For basic auth
	Password     string            `yaml:"password"`     // For basic auth
	PasswordFile string            `yaml:"password_file"` // For basic auth from file
	APIKey       string            `yaml:"api_key"`      // For API key auth
	APIKeyFile   string            `yaml:"api_key_file"` // For API key from file
	Headers      map[string]string `yaml:"headers"`      // Custom headers
}

// CollectorsConfig holds configuration for metric collectors.
type CollectorsConfig struct {
	Global      GlobalCollectorConfig `yaml:"global"`
	Cluster     CollectorConfig       `yaml:"cluster"`
	Nodes       CollectorConfig       `yaml:"nodes"`
	Jobs        CollectorConfig       `yaml:"jobs"`
	Users       CollectorConfig       `yaml:"users"`
	Partitions  CollectorConfig       `yaml:"partitions"`
	Performance CollectorConfig       `yaml:"performance"`
	System      CollectorConfig       `yaml:"system"`
}

// GlobalCollectorConfig holds global collector settings.
type GlobalCollectorConfig struct {
	DefaultInterval    time.Duration `yaml:"default_interval"`
	DefaultTimeout     time.Duration `yaml:"default_timeout"`
	MaxConcurrency     int           `yaml:"max_concurrency"`
	ErrorThreshold     int           `yaml:"error_threshold"`
	RecoveryDelay      time.Duration `yaml:"recovery_delay"`
	GracefulDegradation bool         `yaml:"graceful_degradation"`
}

// CollectorConfig holds configuration for individual collectors.
type CollectorConfig struct {
	Enabled         bool          `yaml:"enabled"`
	Interval        time.Duration `yaml:"interval"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxConcurrency  int           `yaml:"max_concurrency"`
	Labels          map[string]string `yaml:"labels"`
	Filters         FilterConfig  `yaml:"filters"`
	ErrorHandling   ErrorHandlingConfig `yaml:"error_handling"`
}

// FilterConfig holds filtering configuration for collectors.
type FilterConfig struct {
	IncludeNodes      []string `yaml:"include_nodes"`
	ExcludeNodes      []string `yaml:"exclude_nodes"`
	IncludePartitions []string `yaml:"include_partitions"`
	ExcludePartitions []string `yaml:"exclude_partitions"`
	IncludeUsers      []string `yaml:"include_users"`
	ExcludeUsers      []string `yaml:"exclude_users"`
	JobStates         []string `yaml:"job_states"`
	NodeStates        []string `yaml:"node_states"`
}

// ErrorHandlingConfig holds error handling configuration.
type ErrorHandlingConfig struct {
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	BackoffFactor float64       `yaml:"backoff_factor"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay"`
	FailFast      bool          `yaml:"fail_fast"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level          string            `yaml:"level"`          // debug, info, warn, error
	Format         string            `yaml:"format"`         // json, text
	Output         string            `yaml:"output"`         // stdout, stderr, file
	File           string            `yaml:"file"`           // Log file path
	MaxSize        int               `yaml:"max_size"`       // Max size in MB
	MaxAge         int               `yaml:"max_age"`        // Max age in days
	MaxBackups     int               `yaml:"max_backups"`    // Max backup files
	Compress       bool              `yaml:"compress"`       // Compress rotated files
	Fields         map[string]string `yaml:"fields"`         // Additional fields
	SuppressHTTP   bool              `yaml:"suppress_http"`  // Suppress HTTP request logs
}

// MetricsConfig holds metrics configuration.
type MetricsConfig struct {
	Namespace        string            `yaml:"namespace"`
	Subsystem        string            `yaml:"subsystem"`
	ConstLabels      map[string]string `yaml:"const_labels"`
	MaxAge           time.Duration     `yaml:"max_age"`
	AgeBuckets       int               `yaml:"age_buckets"`
	Registry         RegistryConfig    `yaml:"registry"`
	Cardinality      CardinalityConfig `yaml:"cardinality"`
}

// RegistryConfig holds Prometheus registry configuration.
type RegistryConfig struct {
	EnableGoCollector     bool `yaml:"enable_go_collector"`
	EnableProcessCollector bool `yaml:"enable_process_collector"`
	EnableBuildInfo       bool `yaml:"enable_build_info"`
}

// CardinalityConfig holds cardinality management configuration.
type CardinalityConfig struct {
	MaxSeries    int `yaml:"max_series"`
	MaxLabels    int `yaml:"max_labels"`
	MaxLabelSize int `yaml:"max_label_size"`
	WarnLimit    int `yaml:"warn_limit"`
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Address:        ":8080",
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			ReadyPath:      "/ready",
			Timeout:        30 * time.Second,
			ReadTimeout:    10 * time.Second,
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
			APIVersion:    "v0.0.42",
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
			},
			Cluster: CollectorConfig{
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
			Nodes: CollectorConfig{
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
			Jobs: CollectorConfig{
				Enabled:  true,
				Interval: 15 * time.Second,
				Timeout:  10 * time.Second,
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
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
			Performance: CollectorConfig{
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
			System: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
				Timeout:  10 * time.Second,
				ErrorHandling: ErrorHandlingConfig{
					MaxRetries:    3,
					RetryDelay:    5 * time.Second,
					BackoffFactor: 2.0,
					MaxRetryDelay: 60 * time.Second,
				},
			},
		},
		Logging: LoggingConfig{
			Level:        "info",
			Format:       "json",
			Output:       "stdout",
			SuppressHTTP: false,
		},
		Metrics: MetricsConfig{
			Namespace: "slurm",
			Subsystem: "exporter",
			MaxAge:    5 * time.Minute,
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

		// Validate configuration
		if err := cfg.Validate(); err != nil {
			return cfg, fmt.Errorf("configuration validation failed: %w", err)
		}

		return cfg, nil
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return cfg, fmt.Errorf("configuration file %s does not exist", filename)
	}

	// Read file content
	data, err := ioutil.ReadFile(filename)
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
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

	return nil
}

// Validate validates the server configuration.
func (s *ServerConfig) Validate() error {
	if s.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}

	if s.MetricsPath == "" {
		return fmt.Errorf("metrics path cannot be empty")
	}

	if s.HealthPath == "" {
		return fmt.Errorf("health path cannot be empty")
	}

	if s.ReadyPath == "" {
		return fmt.Errorf("ready path cannot be empty")
	}

	if s.Timeout <= 0 {
		return fmt.Errorf("server timeout must be positive")
	}

	if s.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if s.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if s.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	if s.MaxRequestSize <= 0 {
		return fmt.Errorf("max request size must be positive")
	}

	// Validate TLS configuration
	if s.TLS.Enabled {
		if s.TLS.CertFile == "" {
			return fmt.Errorf("TLS cert file must be specified when TLS is enabled")
		}
		if s.TLS.KeyFile == "" {
			return fmt.Errorf("TLS key file must be specified when TLS is enabled")
		}
	}

	// Validate basic auth configuration
	if s.BasicAuth.Enabled {
		if s.BasicAuth.Username == "" {
			return fmt.Errorf("basic auth username must be specified when basic auth is enabled")
		}
		if s.BasicAuth.Password == "" {
			return fmt.Errorf("basic auth password must be specified when basic auth is enabled")
		}
	}

	return nil
}

// Validate validates the SLURM configuration.
func (s *SLURMConfig) Validate() error {
	if s.BaseURL == "" {
		return fmt.Errorf("SLURM base URL cannot be empty")
	}

	if s.APIVersion == "" {
		return fmt.Errorf("SLURM API version cannot be empty")
	}

	if s.Timeout <= 0 {
		return fmt.Errorf("SLURM timeout must be positive")
	}

	if s.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if s.RetryDelay <= 0 {
		return fmt.Errorf("retry delay must be positive")
	}

	if err := s.Auth.Validate(); err != nil {
		return fmt.Errorf("auth configuration: %w", err)
	}

	if s.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("requests per second must be positive")
	}

	if s.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("burst size must be positive")
	}

	return nil
}

// Validate validates the auth configuration.
func (a *AuthConfig) Validate() error {
	switch a.Type {
	case "none":
		// No validation needed
	case "jwt":
		if a.Token == "" && a.TokenFile == "" {
			return fmt.Errorf("JWT token or token file must be specified")
		}
	case "basic":
		if a.Username == "" {
			return fmt.Errorf("basic auth username must be specified")
		}
		if a.Password == "" && a.PasswordFile == "" {
			return fmt.Errorf("basic auth password or password file must be specified")
		}
	case "apikey":
		if a.APIKey == "" && a.APIKeyFile == "" {
			return fmt.Errorf("API key or API key file must be specified")
		}
	default:
		return fmt.Errorf("unsupported auth type: %s", a.Type)
	}

	return nil
}

// Validate validates the collectors configuration.
func (c *CollectorsConfig) Validate() error {
	if c.Global.DefaultInterval <= 0 {
		return fmt.Errorf("default interval must be positive")
	}

	if c.Global.DefaultTimeout <= 0 {
		return fmt.Errorf("default timeout must be positive")
	}

	if c.Global.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	if c.Global.ErrorThreshold < 0 {
		return fmt.Errorf("error threshold cannot be negative")
	}

	if c.Global.RecoveryDelay <= 0 {
		return fmt.Errorf("recovery delay must be positive")
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
	}

	for _, col := range collectors {
		if err := col.collector.Validate(); err != nil {
			return fmt.Errorf("%s collector: %w", col.name, err)
		}
	}

	return nil
}

// Validate validates the collector configuration.
func (c *CollectorConfig) Validate() error {
	if c.Enabled {
		if c.Interval <= 0 {
			return fmt.Errorf("interval must be positive when collector is enabled")
		}

		if c.Timeout <= 0 {
			return fmt.Errorf("timeout must be positive when collector is enabled")
		}

		if c.MaxConcurrency < 0 {
			return fmt.Errorf("max concurrency cannot be negative")
		}

		if err := c.ErrorHandling.Validate(); err != nil {
			return fmt.Errorf("error handling: %w", err)
		}
	}

	return nil
}

// Validate validates the error handling configuration.
func (e *ErrorHandlingConfig) Validate() error {
	if e.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if e.RetryDelay <= 0 {
		return fmt.Errorf("retry delay must be positive")
	}

	if e.BackoffFactor <= 0 {
		return fmt.Errorf("backoff factor must be positive")
	}

	if e.MaxRetryDelay <= 0 {
		return fmt.Errorf("max retry delay must be positive")
	}

	if e.MaxRetryDelay < e.RetryDelay {
		return fmt.Errorf("max retry delay must be greater than or equal to retry delay")
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
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", l.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s (must be json or text)", l.Format)
	}

	validOutputs := map[string]bool{
		"stdout": true,
		"stderr": true,
		"file":   true,
	}

	if !validOutputs[l.Output] {
		return fmt.Errorf("invalid log output: %s (must be stdout, stderr, or file)", l.Output)
	}

	if l.Output == "file" && l.File == "" {
		return fmt.Errorf("log file must be specified when output is file")
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
		return fmt.Errorf("metrics namespace cannot be empty")
	}

	if m.MaxAge <= 0 {
		return fmt.Errorf("max age must be positive")
	}

	if m.AgeBuckets <= 0 {
		return fmt.Errorf("age buckets must be positive")
	}

	if err := m.Cardinality.Validate(); err != nil {
		return fmt.Errorf("cardinality: %w", err)
	}

	return nil
}

// Validate validates the cardinality configuration.
func (c *CardinalityConfig) Validate() error {
	if c.MaxSeries <= 0 {
		return fmt.Errorf("max series must be positive")
	}

	if c.MaxLabels <= 0 {
		return fmt.Errorf("max labels must be positive")
	}

	if c.MaxLabelSize <= 0 {
		return fmt.Errorf("max label size must be positive")
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

// applyServerEnvOverrides applies server-specific environment overrides.
func (c *Config) applyServerEnvOverrides(prefix string) error {
	if val := os.Getenv(prefix + "ADDRESS"); val != "" {
		c.Server.Address = val
	}

	if val := os.Getenv(prefix + "METRICS_PATH"); val != "" {
		c.Server.MetricsPath = val
	}

	if val := os.Getenv(prefix + "HEALTH_PATH"); val != "" {
		c.Server.HealthPath = val
	}

	if val := os.Getenv(prefix + "READY_PATH"); val != "" {
		c.Server.ReadyPath = val
	}

	if val := os.Getenv(prefix + "TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid server timeout: %w", err)
		}
		c.Server.Timeout = duration
	}

	if val := os.Getenv(prefix + "READ_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid read timeout: %w", err)
		}
		c.Server.ReadTimeout = duration
	}

	if val := os.Getenv(prefix + "WRITE_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid write timeout: %w", err)
		}
		c.Server.WriteTimeout = duration
	}

	if val := os.Getenv(prefix + "IDLE_TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid idle timeout: %w", err)
		}
		c.Server.IdleTimeout = duration
	}

	if val := os.Getenv(prefix + "MAX_REQUEST_SIZE"); val != "" {
		size, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid max request size: %w", err)
		}
		c.Server.MaxRequestSize = size
	}

	// TLS configuration
	if val := os.Getenv(prefix + "TLS_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid TLS enabled value: %w", err)
		}
		c.Server.TLS.Enabled = enabled
	}

	if val := os.Getenv(prefix + "TLS_CERT_FILE"); val != "" {
		c.Server.TLS.CertFile = val
	}

	if val := os.Getenv(prefix + "TLS_KEY_FILE"); val != "" {
		c.Server.TLS.KeyFile = val
	}

	// Basic Auth configuration
	if val := os.Getenv(prefix + "BASIC_AUTH_ENABLED"); val != "" {
		enabled, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid basic auth enabled value: %w", err)
		}
		c.Server.BasicAuth.Enabled = enabled
	}

	if val := os.Getenv(prefix + "BASIC_AUTH_USERNAME"); val != "" {
		c.Server.BasicAuth.Username = val
	}

	if val := os.Getenv(prefix + "BASIC_AUTH_PASSWORD"); val != "" {
		c.Server.BasicAuth.Password = val
	}

	return nil
}

// applySLURMEnvOverrides applies SLURM-specific environment overrides.
func (c *Config) applySLURMEnvOverrides(prefix string) error {
	if val := os.Getenv(prefix + "BASE_URL"); val != "" {
		c.SLURM.BaseURL = val
	}

	if val := os.Getenv(prefix + "API_VERSION"); val != "" {
		c.SLURM.APIVersion = val
	}

	if val := os.Getenv(prefix + "TIMEOUT"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid SLURM timeout: %w", err)
		}
		c.SLURM.Timeout = duration
	}

	if val := os.Getenv(prefix + "RETRY_ATTEMPTS"); val != "" {
		attempts, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid retry attempts: %w", err)
		}
		c.SLURM.RetryAttempts = attempts
	}

	if val := os.Getenv(prefix + "RETRY_DELAY"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid retry delay: %w", err)
		}
		c.SLURM.RetryDelay = duration
	}

	// Authentication configuration
	if val := os.Getenv(prefix + "AUTH_TYPE"); val != "" {
		c.SLURM.Auth.Type = val
	}

	if val := os.Getenv(prefix + "AUTH_TOKEN"); val != "" {
		c.SLURM.Auth.Token = val
	}

	if val := os.Getenv(prefix + "AUTH_TOKEN_FILE"); val != "" {
		c.SLURM.Auth.TokenFile = val
	}

	if val := os.Getenv(prefix + "AUTH_USERNAME"); val != "" {
		c.SLURM.Auth.Username = val
	}

	if val := os.Getenv(prefix + "AUTH_PASSWORD"); val != "" {
		c.SLURM.Auth.Password = val
	}

	if val := os.Getenv(prefix + "AUTH_PASSWORD_FILE"); val != "" {
		c.SLURM.Auth.PasswordFile = val
	}

	if val := os.Getenv(prefix + "AUTH_API_KEY"); val != "" {
		c.SLURM.Auth.APIKey = val
	}

	if val := os.Getenv(prefix + "AUTH_API_KEY_FILE"); val != "" {
		c.SLURM.Auth.APIKeyFile = val
	}

	// TLS configuration
	if val := os.Getenv(prefix + "TLS_INSECURE_SKIP_VERIFY"); val != "" {
		skip, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid TLS insecure skip verify value: %w", err)
		}
		c.SLURM.TLS.InsecureSkipVerify = skip
	}

	if val := os.Getenv(prefix + "TLS_CA_CERT_FILE"); val != "" {
		c.SLURM.TLS.CACertFile = val
	}

	if val := os.Getenv(prefix + "TLS_CLIENT_CERT_FILE"); val != "" {
		c.SLURM.TLS.ClientCertFile = val
	}

	if val := os.Getenv(prefix + "TLS_CLIENT_KEY_FILE"); val != "" {
		c.SLURM.TLS.ClientKeyFile = val
	}

	// Rate limiting configuration
	if val := os.Getenv(prefix + "RATE_LIMIT_REQUESTS_PER_SECOND"); val != "" {
		rps, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("invalid requests per second: %w", err)
		}
		c.SLURM.RateLimit.RequestsPerSecond = rps
	}

	if val := os.Getenv(prefix + "RATE_LIMIT_BURST_SIZE"); val != "" {
		burst, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid burst size: %w", err)
		}
		c.SLURM.RateLimit.BurstSize = burst
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
		"CLUSTER":     &c.Collectors.Cluster,
		"NODES":       &c.Collectors.Nodes,
		"JOBS":        &c.Collectors.Jobs,
		"USERS":       &c.Collectors.Users,
		"PARTITIONS":  &c.Collectors.Partitions,
		"PERFORMANCE": &c.Collectors.Performance,
		"SYSTEM":      &c.Collectors.System,
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
	if val := os.Getenv(prefix + "NAMESPACE"); val != "" {
		c.Metrics.Namespace = val
	}

	if val := os.Getenv(prefix + "SUBSYSTEM"); val != "" {
		c.Metrics.Subsystem = val
	}

	if val := os.Getenv(prefix + "MAX_AGE"); val != "" {
		duration, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid max age: %w", err)
		}
		c.Metrics.MaxAge = duration
	}

	if val := os.Getenv(prefix + "AGE_BUCKETS"); val != "" {
		buckets, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid age buckets: %w", err)
		}
		c.Metrics.AgeBuckets = buckets
	}

	// Registry configuration
	if val := os.Getenv(prefix + "REGISTRY_ENABLE_GO_COLLECTOR"); val != "" {
		enable, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid enable go collector value: %w", err)
		}
		c.Metrics.Registry.EnableGoCollector = enable
	}

	if val := os.Getenv(prefix + "REGISTRY_ENABLE_PROCESS_COLLECTOR"); val != "" {
		enable, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid enable process collector value: %w", err)
		}
		c.Metrics.Registry.EnableProcessCollector = enable
	}

	if val := os.Getenv(prefix + "REGISTRY_ENABLE_BUILD_INFO"); val != "" {
		enable, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("invalid enable build info value: %w", err)
		}
		c.Metrics.Registry.EnableBuildInfo = enable
	}

	// Cardinality configuration
	if val := os.Getenv(prefix + "CARDINALITY_MAX_SERIES"); val != "" {
		maxSeries, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max series: %w", err)
		}
		c.Metrics.Cardinality.MaxSeries = maxSeries
	}

	if val := os.Getenv(prefix + "CARDINALITY_MAX_LABELS"); val != "" {
		maxLabels, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max labels: %w", err)
		}
		c.Metrics.Cardinality.MaxLabels = maxLabels
	}

	if val := os.Getenv(prefix + "CARDINALITY_MAX_LABEL_SIZE"); val != "" {
		maxLabelSize, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid max label size: %w", err)
		}
		c.Metrics.Cardinality.MaxLabelSize = maxLabelSize
	}

	if val := os.Getenv(prefix + "CARDINALITY_WARN_LIMIT"); val != "" {
		warnLimit, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid warn limit: %w", err)
		}
		c.Metrics.Cardinality.WarnLimit = warnLimit
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
		watcher.Close()
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