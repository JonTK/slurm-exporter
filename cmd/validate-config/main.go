// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/jontk/slurm-exporter/internal/config"
)

var (
	configFile = flag.String("config", "", "Path to configuration file to validate")
	format     = flag.String("format", "text", "Output format: text, json")
	strict     = flag.Bool("strict", false, "Enable strict validation (warnings treated as errors)")
)

func main() {
	flag.Parse()

	if *configFile == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -config <config-file>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load and validate configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		if *format == "json" {
			output := map[string]interface{}{
				"valid":  false,
				"errors": []string{err.Error()},
			}
			if encErr := json.NewEncoder(os.Stdout).Encode(output); encErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to encode JSON output: %v\n", encErr)
			}
		} else {
			fmt.Fprintf(os.Stderr, "❌ Configuration validation failed:\n%s\n", err.Error())
		}
		os.Exit(1)
	}

	// Output validation results
	if *format == "json" {
		output := map[string]interface{}{
			"valid":       true,
			"config_file": *configFile,
			"strict_mode": *strict,
			"message":     "Configuration validation passed",
		}
		if encErr := json.NewEncoder(os.Stdout).Encode(output); encErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to encode JSON output: %v\n", encErr)
			os.Exit(1)
		}
	} else {
		fmt.Printf("✅ Configuration validation passed!\n")
		fmt.Printf("📄 Config file: %s\n", *configFile)
		fmt.Printf("🔧 Strict mode: %t\n", *strict)

		// Show some key configuration info
		fmt.Printf("\n📊 Configuration Summary:\n")
		fmt.Printf("  • Server: %s\n", cfg.Server.Address)
		fmt.Printf("  • SLURM: %s\n", cfg.SLURM.BaseURL)

		enabledCollectors := getEnabledCollectors(cfg)
		fmt.Printf("  • Collectors: %d enabled (%v)\n", len(enabledCollectors), enabledCollectors)

		if cfg.Metrics.Cardinality.MaxSeries > 0 {
			fmt.Printf("  • Max series: %d\n", cfg.Metrics.Cardinality.MaxSeries)
		}
	}
}

func getEnabledCollectors(cfg *config.Config) []string {
	var enabled []string
	collectors := map[string]bool{
		"jobs":         cfg.Collectors.Jobs.Enabled,
		"nodes":        cfg.Collectors.Nodes.Enabled,
		"partitions":   cfg.Collectors.Partitions.Enabled,
		"cluster":      cfg.Collectors.Cluster.Enabled,
		"users":        cfg.Collectors.Users.Enabled,
		"qos":          cfg.Collectors.QoS.Enabled,
		"reservations": cfg.Collectors.Reservations.Enabled,
		"accounts":     cfg.Collectors.Accounts.Enabled,
		"associations": cfg.Collectors.Associations.Enabled,
		"performance":  cfg.Collectors.Performance.Enabled,
		"system":       cfg.Collectors.System.Enabled,
	}

	for name, isEnabled := range collectors {
		if isEnabled {
			enabled = append(enabled, name)
		}
	}

	return enabled
}
