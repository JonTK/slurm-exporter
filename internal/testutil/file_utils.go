// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package testutil

import (
	"io/fs"
	"os"
)

// WriteFile writes data to a file
func WriteFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0600)
}

// RemoveFile removes a file
func RemoveFile(filename string) error {
	return os.Remove(filename)
}

// ReadFile reads a file
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CreateTempDir creates a temporary directory
func CreateTempDir(pattern string) (string, error) {
	return os.MkdirTemp("", pattern)
}

// RemoveAll removes a directory and all its contents
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Chmod changes file permissions
func Chmod(name string, mode fs.FileMode) error {
	return os.Chmod(name, mode)
}
