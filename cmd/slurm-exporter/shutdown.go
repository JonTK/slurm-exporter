// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// ShutdownHook represents a function to execute during shutdown
type ShutdownHook func(ctx context.Context) error

// ShutdownManager manages graceful shutdown of the application
type ShutdownManager struct {
	logger     *logrus.Logger
	timeout    time.Duration
	hooks      map[string]ShutdownHook
	hookOrder  []string
	sigChan    chan os.Signal
	shutdownCh chan struct{}
	mu         sync.RWMutex
	started    bool
}

// NewShutdownManager creates a new shutdown manager
func NewShutdownManager(logger *logrus.Logger, timeout time.Duration) *ShutdownManager {
	return &ShutdownManager{
		logger:     logger,
		timeout:    timeout,
		hooks:      make(map[string]ShutdownHook),
		hookOrder:  make([]string, 0),
		sigChan:    make(chan os.Signal, 1),
		shutdownCh: make(chan struct{}),
	}
}

// AddShutdownHook adds a shutdown hook with a given name
func (sm *ShutdownManager) AddShutdownHook(name string, hook ShutdownHook) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.hooks[name]; !exists {
		sm.hookOrder = append(sm.hookOrder, name)
	}
	sm.hooks[name] = hook

	sm.logger.WithField("hook", name).Debug("Shutdown hook registered")
}

// RemoveShutdownHook removes a shutdown hook by name
func (sm *ShutdownManager) RemoveShutdownHook(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.hooks[name]; exists {
		delete(sm.hooks, name)

		// Remove from order slice
		for i, hookName := range sm.hookOrder {
			if hookName == name {
				sm.hookOrder = append(sm.hookOrder[:i], sm.hookOrder[i+1:]...)
				break
			}
		}

		sm.logger.WithField("hook", name).Debug("Shutdown hook removed")
	}
}

// Start begins listening for shutdown signals
func (sm *ShutdownManager) Start(ctx context.Context) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.started {
		return
	}

	// Listen for shutdown signals
	signal.Notify(sm.sigChan,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination request
		syscall.SIGQUIT, // Quit request
		syscall.SIGHUP,  // Hang up
	)

	sm.started = true
	sm.logger.Info("Shutdown manager started, listening for signals")
}

// SignalChan returns the signal channel for external monitoring
func (sm *ShutdownManager) SignalChan() <-chan os.Signal {
	return sm.sigChan
}

// Shutdown executes all registered shutdown hooks in reverse order
func (sm *ShutdownManager) Shutdown() error {
	sm.logger.Info("Starting graceful shutdown process")
	start := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	// Stop listening for signals
	signal.Stop(sm.sigChan)
	close(sm.shutdownCh)

	sm.mu.RLock()
	hooks := make([]string, len(sm.hookOrder))
	copy(hooks, sm.hookOrder)
	sm.mu.RUnlock()

	// Execute hooks in reverse order (LIFO - last registered, first executed)
	var errors []error
	for i := len(hooks) - 1; i >= 0; i-- {
		hookName := hooks[i]

		sm.mu.RLock()
		hook, exists := sm.hooks[hookName]
		sm.mu.RUnlock()

		if !exists {
			continue
		}

		hookStart := time.Now()
		sm.logger.WithField("hook", hookName).Info("Executing shutdown hook")

		if err := sm.executeHookWithTimeout(ctx, hookName, hook); err != nil {
			sm.logger.WithError(err).WithField("hook", hookName).Error("Shutdown hook failed")
			errors = append(errors, fmt.Errorf("hook %s: %w", hookName, err))
		} else {
			duration := time.Since(hookStart)
			sm.logger.WithFields(logrus.Fields{
				"hook":     hookName,
				"duration": duration,
			}).Info("Shutdown hook completed successfully")
		}
	}

	totalDuration := time.Since(start)
	sm.logger.WithFields(logrus.Fields{
		"total_duration": totalDuration,
		"hooks_count":    len(hooks),
		"errors_count":   len(errors),
	}).Info("Graceful shutdown process completed")

	// Return combined error if any hooks failed
	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// executeHookWithTimeout executes a hook with timeout protection
func (sm *ShutdownManager) executeHookWithTimeout(ctx context.Context, name string, hook ShutdownHook) error {
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic in shutdown hook %s: %v", name, r)
			}
		}()
		done <- hook(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown hook %s timed out", name)
	}
}

// IsShuttingDown returns true if shutdown has been initiated
func (sm *ShutdownManager) IsShuttingDown() bool {
	select {
	case <-sm.shutdownCh:
		return true
	default:
		return false
	}
}

// Wait blocks until shutdown is initiated
func (sm *ShutdownManager) Wait() {
	<-sm.shutdownCh
}

// GetRegisteredHooks returns a list of registered hook names in order
func (sm *ShutdownManager) GetRegisteredHooks() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	hooks := make([]string, len(sm.hookOrder))
	copy(hooks, sm.hookOrder)
	return hooks
}
