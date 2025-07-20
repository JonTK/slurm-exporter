package config

import (
	"fmt"
	"time"
)

// Config represents the application configuration.
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	SLURM      SLURMConfig      `yaml:"slurm"`
	Collectors CollectorsConfig `yaml:"collectors"`
	Logging    LoggingConfig    `yaml:"logging"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Address     string        `yaml:"address"`
	MetricsPath string        `yaml:"metrics_path"`
	Timeout     time.Duration `yaml:"timeout"`
	TLS         TLSConfig     `yaml:"tls"`
}

// TLSConfig holds TLS configuration.
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// SLURMConfig holds SLURM connection configuration.
type SLURMConfig struct {
	BaseURL string      `yaml:"base_url"`
	Auth    AuthConfig  `yaml:"auth"`
	Timeout time.Duration `yaml:"timeout"`
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	Type     string `yaml:"type"`     // jwt, basic, apikey, none
	Token    string `yaml:"token"`    // For JWT
	Username string `yaml:"username"` // For basic auth
	Password string `yaml:"password"` // For basic auth
	APIKey   string `yaml:"api_key"`  // For API key auth
}

// CollectorsConfig holds configuration for metric collectors.
type CollectorsConfig struct {
	Cluster     CollectorConfig `yaml:"cluster"`
	Nodes       CollectorConfig `yaml:"nodes"`
	Jobs        CollectorConfig `yaml:"jobs"`
	Users       CollectorConfig `yaml:"users"`
	Partitions  CollectorConfig `yaml:"partitions"`
	Performance CollectorConfig `yaml:"performance"`
}

// CollectorConfig holds configuration for individual collectors.
type CollectorConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}

// LoggingConfig holds logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// Load loads configuration from a file.
func Load(filename string) (*Config, error) {
	// For now, return a default configuration
	// This will be implemented in task 2.2
	cfg := &Config{
		Server: ServerConfig{
			Address:     ":8080",
			MetricsPath: "/metrics",
			Timeout:     30 * time.Second,
		},
		SLURM: SLURMConfig{
			BaseURL: "http://localhost:6820",
			Auth: AuthConfig{
				Type: "none",
			},
			Timeout: 30 * time.Second,
		},
		Collectors: CollectorsConfig{
			Cluster: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
			},
			Nodes: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
			},
			Jobs: CollectorConfig{
				Enabled:  true,
				Interval: 15 * time.Second,
			},
			Users: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
			},
			Partitions: CollectorConfig{
				Enabled:  true,
				Interval: 60 * time.Second,
			},
			Performance: CollectorConfig{
				Enabled:  true,
				Interval: 30 * time.Second,
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	fmt.Printf("Loading config from %s (using defaults for now)\n", filename)
	return cfg, nil
}