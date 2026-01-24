// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package config

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jontk/slurm-exporter/internal/testutil"
)

// MockReloadHandler for testing
type MockReloadHandler struct {
	mu          sync.Mutex
	reloadCount int
	lastConfig  *Config
	lastError   error
}

func (m *MockReloadHandler) Handle(config *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reloadCount++
	m.lastConfig = config
	return m.lastError
}

func (m *MockReloadHandler) GetReloadCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reloadCount
}

func (m *MockReloadHandler) GetLastConfig() *Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastConfig
}

func TestWatcher_New(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(`
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)

	_ = watcher.Stop()
}

func TestWatcher_Start_Stop(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(`
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Stop watcher
	_ = watcher.Stop()

	// Should have loaded initial config
	assert.Equal(t, 1, handler.GetReloadCount())
	assert.NoError(t, handler.lastError) // Direct access is OK for error check in defer
	assert.NotNil(t, handler.GetLastConfig())
}

func TestWatcher_ConfigChange(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	initialConfig := `
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`
	err = os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)
	watcher.debounceTime = 100 * time.Millisecond // Set short debounce for testing

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	// Give it time to start
	time.Sleep(200 * time.Millisecond)

	// Modify config file
	modifiedConfig := `
server:
  address: ":9090"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 45s
`
	err = os.WriteFile(configFile, []byte(modifiedConfig), 0644)
	require.NoError(t, err)

	// Wait for file change to be detected and debounce timer to fire
	time.Sleep(300 * time.Millisecond)

	_ = watcher.Stop()

	// Should have reloaded config
	assert.True(t, handler.GetReloadCount() >= 2, "should have reloaded config at least twice (initial + change)")
	assert.NoError(t, handler.lastError) // Direct access is OK for error check in defer
	assert.NotNil(t, handler.GetLastConfig())
	assert.Equal(t, ":9090", handler.GetLastConfig().Server.Address)
}

func TestWatcher_InvalidConfig(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	validConfig := `
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`
	err = os.WriteFile(configFile, []byte(validConfig), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)
	watcher.debounceTime = 100 * time.Millisecond // Set short debounce for testing

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	// Give it time to start
	time.Sleep(200 * time.Millisecond)

	// Write invalid YAML
	invalidConfig := `
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: invalid_duration
  invalid_yaml: [
`
	err = os.WriteFile(configFile, []byte(invalidConfig), 0644)
	require.NoError(t, err)

	// Wait for file change to be detected and debounce timer to fire
	time.Sleep(300 * time.Millisecond)

	_ = watcher.Stop()

	// Handler should only have been called once (initial load)
	// Invalid config should be rejected and handler not called again
	assert.Equal(t, 1, handler.reloadCount, "handler should only be called for initial valid config")
	assert.NoError(t, handler.lastError, "last successful handler call should have no error")
	// The config should still be the initial valid config
	assert.Equal(t, ":8080", handler.GetLastConfig().Server.Address)
}

func TestWatcher_NonExistentFile(t *testing.T) {
	t.Parallel()
	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	_, err := NewWatcher("/nonexistent/config.yaml", handler.Handle, logger)
	assert.Error(t, err)
}

func TestWatcher_GetConfig(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(`
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)

	// Start watcher briefly
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(200 * time.Millisecond)

	// Get current config
	currentConfig := watcher.GetConfig()
	assert.NotNil(t, currentConfig)
	assert.Equal(t, ":8080", currentConfig.Server.Address)

	_ = watcher.Stop()
}

func TestWatcher_DebounceMultipleChanges(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	initialConfig := `
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`
	err = os.WriteFile(configFile, []byte(initialConfig), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	// Create watcher with short debounce time for testing
	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)
	watcher.debounceTime = 100 * time.Millisecond

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	// Give it time to start
	time.Sleep(200 * time.Millisecond)
	initialCount := handler.GetReloadCount()

	// Make multiple rapid changes
	for i := 0; i < 5; i++ {
		config := `
server:
  address: ":808` + string(rune('0'+i)) + `"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`
		err = os.WriteFile(configFile, []byte(config), 0644)
		require.NoError(t, err)
		time.Sleep(20 * time.Millisecond) // Rapid changes
	}

	// Wait for debounce to settle
	time.Sleep(300 * time.Millisecond)

	_ = watcher.Stop()

	// Should have debounced the changes (not reload for every single change)
	reloadCount := handler.GetReloadCount() - initialCount
	assert.True(t, reloadCount < 5, "should have debounced rapid changes, got %d reloads", reloadCount)
	assert.True(t, reloadCount >= 1, "should have at least one reload")
}

func TestWatcher_ConcurrentAccess(t *testing.T) {
	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "watcher_test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(`
server:
  address: ":8080"
  metrics_path: "/metrics"
slurm:
  base_url: "http://localhost:6820"
  timeout: 30s
`), 0644)
	require.NoError(t, err)

	logger := testutil.GetTestLogger()
	handler := &MockReloadHandler{}

	watcher, err := NewWatcher(configFile, handler.Handle, logger)
	require.NoError(t, err)

	// Start watcher
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		err := watcher.Start(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(100 * time.Millisecond)

	// Concurrent access to GetConfig
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				config := watcher.GetConfig()
				assert.NotNil(t, config)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	_ = watcher.Stop()
}
