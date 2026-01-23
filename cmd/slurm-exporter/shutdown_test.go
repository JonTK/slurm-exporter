// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func createTestShutdownManager() *ShutdownManager {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce test noise
	return NewShutdownManager(logger, 5*time.Second)
}

func TestNewShutdownManager(t *testing.T) {
	logger := logrus.New()
	timeout := 10 * time.Second

	sm := NewShutdownManager(logger, timeout)

	if sm == nil {
		t.Fatal("Expected shutdown manager to be created")
	}

	if sm.logger != logger {
		t.Error("Expected logger to be set")
	}

	if sm.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, sm.timeout)
	}

	if sm.hooks == nil {
		t.Error("Expected hooks map to be initialized")
	}

	if sm.hookOrder == nil {
		t.Error("Expected hook order slice to be initialized")
	}

	if sm.sigChan == nil {
		t.Error("Expected signal channel to be initialized")
	}
}

func TestAddShutdownHook(t *testing.T) {
	sm := createTestShutdownManager()

	var executed bool
	hook := func(ctx context.Context) error {
		executed = true
		return nil
	}

	// Add hook
	sm.AddShutdownHook("test", hook)

	// Verify hook was added
	hooks := sm.GetRegisteredHooks()
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook, got %d", len(hooks))
	}

	if hooks[0] != "test" {
		t.Errorf("Expected hook name 'test', got '%s'", hooks[0])
	}

	// Test hook replacement
	var executed2 bool
	hook2 := func(ctx context.Context) error {
		executed2 = true
		return nil
	}

	sm.AddShutdownHook("test", hook2)

	// Should still have only one hook
	hooks = sm.GetRegisteredHooks()
	if len(hooks) != 1 {
		t.Errorf("Expected 1 hook after replacement, got %d", len(hooks))
	}

	// Execute shutdown to verify the second hook was registered
	if err := sm.Shutdown(); err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}

	if executed {
		t.Error("First hook should not have been executed")
	}

	if !executed2 {
		t.Error("Second hook should have been executed")
	}
}

func TestRemoveShutdownHook(t *testing.T) {
	sm := createTestShutdownManager()

	// Add multiple hooks
	sm.AddShutdownHook("hook1", func(ctx context.Context) error { return nil })
	sm.AddShutdownHook("hook2", func(ctx context.Context) error { return nil })
	sm.AddShutdownHook("hook3", func(ctx context.Context) error { return nil })

	// Verify all hooks added
	hooks := sm.GetRegisteredHooks()
	if len(hooks) != 3 {
		t.Errorf("Expected 3 hooks, got %d", len(hooks))
	}

	// Remove middle hook
	sm.RemoveShutdownHook("hook2")

	hooks = sm.GetRegisteredHooks()
	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks after removal, got %d", len(hooks))
	}

	expectedHooks := []string{"hook1", "hook3"}
	for i, expected := range expectedHooks {
		if hooks[i] != expected {
			t.Errorf("Expected hook %d to be '%s', got '%s'", i, expected, hooks[i])
		}
	}

	// Remove non-existent hook (should not panic)
	sm.RemoveShutdownHook("nonexistent")

	hooks = sm.GetRegisteredHooks()
	if len(hooks) != 2 {
		t.Errorf("Expected 2 hooks after removing non-existent, got %d", len(hooks))
	}
}

func TestShutdownHookExecution(t *testing.T) {
	sm := createTestShutdownManager()

	var executionOrder []string
	var mu sync.Mutex

	// Add hooks in order
	for i := 1; i <= 3; i++ {
		hookName := fmt.Sprintf("hook%d", i)
		sm.AddShutdownHook(hookName, func(name string) ShutdownHook {
			return func(ctx context.Context) error {
				mu.Lock()
				executionOrder = append(executionOrder, name)
				mu.Unlock()
				return nil
			}
		}(hookName))
	}

	// Execute shutdown
	if err := sm.Shutdown(); err != nil {
		t.Errorf("Unexpected shutdown error: %v", err)
	}

	// Verify execution order (should be reverse of registration order)
	expectedOrder := []string{"hook3", "hook2", "hook1"}
	if len(executionOrder) != len(expectedOrder) {
		t.Errorf("Expected %d hooks executed, got %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Errorf("Expected execution order %d to be '%s', got '%s'", i, expected, executionOrder[i])
		}
	}
}

func TestShutdownHookError(t *testing.T) {
	sm := createTestShutdownManager()

	// Add hooks with one that fails
	sm.AddShutdownHook("success1", func(ctx context.Context) error { return nil })
	sm.AddShutdownHook("failure", func(ctx context.Context) error { return errors.New("hook failed") })
	sm.AddShutdownHook("success2", func(ctx context.Context) error { return nil })

	// Execute shutdown
	err := sm.Shutdown()

	// Should return error
	if err == nil {
		t.Error("Expected shutdown to return error when hook fails")
	}

	if !strings.Contains(err.Error(), "hook failed") {
		t.Errorf("Expected error to contain 'hook failed', got: %v", err)
	}
}

func TestShutdownHookTimeout(t *testing.T) {
	// Create manager with short timeout
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	sm := NewShutdownManager(logger, 100*time.Millisecond)

	// Add hook that takes longer than timeout
	sm.AddShutdownHook("slow", func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	// Execute shutdown
	start := time.Now()
	err := sm.Shutdown()
	duration := time.Since(start)

	// Should return timeout error
	if err == nil {
		t.Error("Expected shutdown to return timeout error")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected error to contain 'timed out', got: %v", err)
	}

	// Should not take much longer than timeout
	if duration > 500*time.Millisecond {
		t.Errorf("Shutdown took too long: %v", duration)
	}
}

func TestShutdownHookPanic(t *testing.T) {
	sm := createTestShutdownManager()

	// Add hook that panics
	sm.AddShutdownHook("panic", func(ctx context.Context) error {
		panic("test panic")
	})

	// Execute shutdown
	err := sm.Shutdown()

	// Should return panic error
	if err == nil {
		t.Error("Expected shutdown to return panic error")
	}

	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("Expected error to contain 'panic', got: %v", err)
	}
}

func TestSignalHandling(t *testing.T) {
	// Skip on Windows - SIGUSR1 is Unix-only
	if runtime.GOOS == "windows" {
		t.Skip("Skipping signal handling test on Windows - SIGUSR1 not available")
	}

	sm := createTestShutdownManager()
	ctx := context.Background()

	// Start signal handling
	sm.Start(ctx)

	// Verify we can get signal channel
	sigChan := sm.SignalChan()
	if sigChan == nil {
		t.Error("Expected signal channel to be available")
	}

	// Test sending signal (in a separate goroutine to avoid blocking)
	go func() {
		time.Sleep(10 * time.Millisecond)
		// Send signal to current process
		process, _ := os.FindProcess(os.Getpid())
		_ = process.Signal(syscall.SIGINT) // Use cross-platform signal for testing
	}()

	// Note: This test is tricky because we can't easily test real signal handling
	// in unit tests without affecting the test process itself
}

func TestIsShuttingDown(t *testing.T) {
	sm := createTestShutdownManager()

	// Initially should not be shutting down
	if sm.IsShuttingDown() {
		t.Error("Expected not to be shutting down initially")
	}

	// After shutdown, should be shutting down
	_ = sm.Shutdown()

	if !sm.IsShuttingDown() {
		t.Error("Expected to be shutting down after Shutdown() called")
	}
}

func TestConcurrentAccess(t *testing.T) {
	sm := createTestShutdownManager()

	var wg sync.WaitGroup

	// Add hooks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			hookName := fmt.Sprintf("hook%d", index)
			sm.AddShutdownHook(hookName, func(ctx context.Context) error {
				return nil
			})
		}(i)
	}

	// Remove hooks concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			hookName := fmt.Sprintf("hook%d", index)
			sm.RemoveShutdownHook(hookName)
		}(i)
	}

	wg.Wait()

	// Should not panic and should have some hooks remaining
	hooks := sm.GetRegisteredHooks()
	if len(hooks) == 0 {
		t.Error("Expected some hooks to remain after concurrent operations")
	}

	// Shutdown should work
	if err := sm.Shutdown(); err != nil {
		t.Errorf("Unexpected error during shutdown: %v", err)
	}
}
