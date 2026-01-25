// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package main

import (
	"flag"
	"os"
	"testing"
)

//nolint:paralleltest // Cannot run in parallel due to global flag.CommandLine state
func TestMainFlags(t *testing.T) {
	// Cannot run in parallel due to global flag.CommandLine state
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test version flag
	os.Args = []string{"cmd", "--version"}

	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Re-define flags
	showVersion := flag.Bool("version", false, "Show version information and exit")
	flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.String("addr", ":8080", "Address to listen on")
	flag.String("metrics-path", "/metrics", "Path for metrics endpoint")

	flag.Parse()

	if !*showVersion {
		t.Error("Expected version flag to be true")
	}

	// Test default values
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	showVersion = flag.Bool("version", false, "Show version information and exit")
	configFile = flag.String("config", "configs/config.yaml", "Path to configuration file")
	logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	addr = flag.String("addr", ":8080", "Address to listen on")
	metricsPath = flag.String("metrics-path", "/metrics", "Path for metrics endpoint")

	flag.Parse()

	if *showVersion {
		t.Error("Expected version flag to be false by default")
	}

	if *configFile != "configs/config.yaml" {
		t.Errorf("Expected default config file to be 'configs/config.yaml', got '%s'", *configFile)
	}

	if *logLevel != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", *logLevel)
	}

	if *addr != ":8080" {
		t.Errorf("Expected default address to be ':8080', got '%s'", *addr)
	}

	if *metricsPath != "/metrics" {
		t.Errorf("Expected default metrics path to be '/metrics', got '%s'", *metricsPath)
	}
}

//nolint:paralleltest // Cannot run in parallel due to global flag.CommandLine state
func TestMainFlagsCustomValues(t *testing.T) {
	// Cannot run in parallel due to global flag.CommandLine state
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test custom values
	os.Args = []string{
		"cmd",
		"--config", "/custom/config.yaml",
		"--log-level", "debug",
		"--addr", ":9090",
		"--metrics-path", "/custom/metrics",
	}

	// Reset flags for testing
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Re-define flags and capture returned pointers
	showVersion := flag.Bool("version", false, "Show version information and exit")
	configFile := flag.String("config", "configs/config.yaml", "Path to configuration file")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	addr := flag.String("addr", ":8080", "Address to listen on")
	metricsPath := flag.String("metrics-path", "/metrics", "Path for metrics endpoint")

	flag.Parse()

	if *showVersion {
		t.Error("Expected version flag to be false")
	}

	if *configFile != "/custom/config.yaml" {
		t.Errorf("Expected config file to be '/custom/config.yaml', got '%s'", *configFile)
	}

	if *logLevel != "debug" {
		t.Errorf("Expected log level to be 'debug', got '%s'", *logLevel)
	}

	if *addr != ":9090" {
		t.Errorf("Expected address to be ':9090', got '%s'", *addr)
	}

	if *metricsPath != "/custom/metrics" {
		t.Errorf("Expected metrics path to be '/custom/metrics', got '%s'", *metricsPath)
	}
}
