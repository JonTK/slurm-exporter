// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	t.Parallel()
	info := Get()

	// Check that fields are populated
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Check that GoVersion matches runtime
	expectedGoVersion := runtime.Version()
	if info.GoVersion != expectedGoVersion {
		t.Errorf("Expected GoVersion %s, got %s", expectedGoVersion, info.GoVersion)
	}

	// Check platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Expected Platform %s, got %s", expectedPlatform, info.Platform)
	}
}

func TestString(t *testing.T) {
	t.Parallel()
	info := Get()
	result := info.String()

	// Check that the string contains expected components
	expectedComponents := []string{
		"Version:",
		"BuildTime:",
		"GitCommit:",
		"GoVersion:",
		"Platform:",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(result, component) {
			t.Errorf("String() result should contain %s, got: %s", component, result)
		}
	}
}

func TestShort(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "dev version",
			version:  "dev",
			expected: "dev (dev build)",
		},
		{
			name:     "release version",
			version:  "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "prerelease version",
			version:  "v1.0.0-beta.1",
			expected: "v1.0.0-beta.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily set the version
			originalVersion := Version
			Version = tt.version
			defer func() { Version = originalVersion }()

			info := Get()
			result := info.Short()

			if result != tt.expected {
				t.Errorf("Expected Short() to return %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestInfoStructure(t *testing.T) {
	t.Parallel()
	info := Get()

	// Verify all fields are accessible
	_ = info.Version
	_ = info.BuildTime
	_ = info.GitCommit
	_ = info.GoVersion
	_ = info.Platform

	// Test that Info can be used as expected
	if info.GoVersion != runtime.Version() {
		t.Error("GoVersion should match runtime.Version()")
	}
}
