// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// ReloadHandler is called when configuration changes are detected
type ReloadHandler func(*Config) error

// Watcher monitors configuration file changes and triggers reloads
type Watcher struct {
	configFile    string
	handler       ReloadHandler
	watcher       *fsnotify.Watcher
	logger        *logrus.Entry
	mu            sync.RWMutex
	currentConfig *Config
	stopChan      chan struct{}
	debounceTime  time.Duration
}

// NewWatcher creates a new configuration watcher
func NewWatcher(configFile string, handler ReloadHandler, logger *logrus.Entry) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Load initial configuration
	initialConfig, err := Load(configFile)
	if err != nil {
		_ = fsWatcher.Close()
		return nil, fmt.Errorf("failed to load initial configuration: %w", err)
	}

	w := &Watcher{
		configFile:    configFile,
		handler:       handler,
		watcher:       fsWatcher,
		logger:        logger.WithField("component", "config-watcher"),
		currentConfig: initialConfig,
		stopChan:      make(chan struct{}),
		debounceTime:  2 * time.Second, // Debounce rapid changes
	}

	// Add the config file to the watcher
	if err := fsWatcher.Add(configFile); err != nil {
		_ = fsWatcher.Close()
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	return w, nil
}

// Start begins watching for configuration changes
func (w *Watcher) Start(ctx context.Context) error {
	w.logger.WithField("file", w.configFile).Info("Starting configuration watcher")

	// Apply initial configuration
	if err := w.handler(w.currentConfig); err != nil {
		return fmt.Errorf("failed to apply initial configuration: %w", err)
	}

	go w.watch(ctx)
	return nil
}

// watch monitors for file changes
func (w *Watcher) watch(ctx context.Context) {
	var (
		timer     *time.Timer
		timerChan <-chan time.Time
	)

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Configuration watcher stopped by context")
			return

		case <-w.stopChan:
			w.logger.Info("Configuration watcher stopped")
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				w.logger.Warn("File watcher events channel closed")
				return
			}

			// Handle file events
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				w.logger.WithFields(logrus.Fields{
					"file":      event.Name,
					"operation": event.Op.String(),
				}).Debug("Configuration file changed")

				// Reset debounce timer
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(w.debounceTime)
				timerChan = timer.C
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				w.logger.Warn("File watcher errors channel closed")
				return
			}
			w.logger.WithError(err).Error("File watcher error")

		case <-timerChan:
			// Debounce period expired, reload configuration
			timerChan = nil
			w.reload()
		}
	}
}

// reload attempts to reload the configuration
func (w *Watcher) reload() {
	w.logger.Info("Reloading configuration")

	// Load new configuration
	newConfig, err := Load(w.configFile)
	if err != nil {
		w.logger.WithError(err).Error("Failed to load new configuration, keeping current")
		return
	}

	// Validate the new configuration
	if err := newConfig.ValidateEnhanced(); err != nil {
		w.logger.WithError(err).Error("New configuration validation failed, keeping current")
		return
	}

	// Check if configuration actually changed
	if configEqual(w.GetConfig(), newConfig) {
		w.logger.Debug("Configuration unchanged after reload")
		return
	}

	// Apply the new configuration
	if err := w.handler(newConfig); err != nil {
		w.logger.WithError(err).Error("Failed to apply new configuration, keeping current")
		return
	}

	// Update current configuration
	w.mu.Lock()
	w.currentConfig = newConfig
	w.mu.Unlock()

	w.logger.Info("Configuration reloaded successfully")
}

// GetConfig returns the current configuration
func (w *Watcher) GetConfig() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentConfig
}

// Stop stops the configuration watcher
func (w *Watcher) Stop() error {
	w.logger.Info("Stopping configuration watcher")

	close(w.stopChan)
	return w.watcher.Close()
}

// configEqual performs a simple equality check on configurations
// In a real implementation, this would be more sophisticated
func configEqual(a, b *Config) bool {
	// For now, we'll assume any loaded config is different
	// A proper implementation would compare all fields
	return false
}

// ReloadableRegistry defines the interface for registries that support hot-reload
type ReloadableRegistry interface {
	// ReconfigureCollectors updates collector configuration without restart
	ReconfigureCollectors(config *CollectorsConfig) error
}

// CreateReloadHandler creates a reload handler for the given registry
func CreateReloadHandler(registry ReloadableRegistry, logger *logrus.Entry) ReloadHandler {
	return func(newConfig *Config) error {
		logger.Info("Applying new configuration")

		// Update collector configurations
		if err := registry.ReconfigureCollectors(&newConfig.Collectors); err != nil {
			return fmt.Errorf("failed to reconfigure collectors: %w", err)
		}

		// Note: Some configuration changes (like server address) would require restart
		// We only support reloading collector configurations for now
		logger.Info("Configuration reload completed")
		return nil
	}
}
