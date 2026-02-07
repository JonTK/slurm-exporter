// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"os"
	"runtime"
	"testing"
)

func TestReadLoadAverage(t *testing.T) {
	// This test only works on Linux systems with /proc/loadavg
	if _, err := os.Stat("/proc/loadavg"); os.IsNotExist(err) {
		t.Skip("Skipping test: /proc/loadavg not available")
	}

	loadAvgs, err := readLoadAverage()
	if err != nil {
		t.Fatalf("readLoadAverage() failed: %v", err)
	}

	if len(loadAvgs) != 3 {
		t.Errorf("Expected 3 load averages, got %d", len(loadAvgs))
	}

	// Load averages should be non-negative
	for i, val := range loadAvgs {
		if val < 0 {
			t.Errorf("Load average[%d] is negative: %f", i, val)
		}
	}
}

func TestReadDiskUsage(t *testing.T) {
	// Disk usage monitoring is not supported on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test: disk usage monitoring not supported on Windows")
	}

	// Test reading disk usage for root filesystem
	stats, err := readDiskUsage("/")
	if err != nil {
		t.Fatalf("readDiskUsage() failed: %v", err)
	}

	if stats.Total == 0 {
		t.Error("Total disk space is 0")
	}

	if stats.Used > stats.Total {
		t.Errorf("Used space (%d) exceeds total space (%d)", stats.Used, stats.Total)
	}

	if stats.Free > stats.Total {
		t.Errorf("Free space (%d) exceeds total space (%d)", stats.Free, stats.Total)
	}
}

func TestReadDiskUsageNonExistentPath(t *testing.T) {
	_, err := readDiskUsage("/this/path/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}
}

func TestGetFileModTime(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-modtime-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Get the modification time
	modTime, err := getFileModTime(tmpFile.Name())
	if err != nil {
		t.Fatalf("getFileModTime() failed: %v", err)
	}

	if modTime.IsZero() {
		t.Error("Modification time is zero")
	}
}

func TestGetFileModTimeNonExistent(t *testing.T) {
	_, err := getFileModTime("/this/file/does/not/exist.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}
