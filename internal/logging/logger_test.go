// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestNewLogger(t *testing.T) {
	t.Run("WithValidConfig", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("Expected logger to be created")
		}

		if logger.GetLevel() != logrus.InfoLevel {
			t.Errorf("Expected log level Info, got %v", logger.GetLevel())
		}

		if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
			t.Error("Expected JSON formatter")
		}
	})

	t.Run("WithNilConfig", func(t *testing.T) {
		logger, err := NewLogger(nil)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if logger == nil {
			t.Fatal("Expected logger to be created with defaults")
		}

		if logger.GetLevel() != logrus.InfoLevel {
			t.Errorf("Expected default log level Info, got %v", logger.GetLevel())
		}
	})

	t.Run("WithInvalidLevel", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "invalid",
			Format: "json",
			Output: "stdout",
		}

		_, err := NewLogger(cfg)
		if err == nil {
			t.Error("Expected error for invalid log level")
		}
	})

	t.Run("WithTextFormatter", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "debug",
			Format: "text",
			Output: "stdout",
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
			t.Error("Expected text formatter")
		}

		if logger.GetLevel() != logrus.DebugLevel {
			t.Errorf("Expected log level Debug, got %v", logger.GetLevel())
		}
	})

	t.Run("WithConstantFields", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
			Fields: map[string]string{
				"service": "slurm-exporter",
				"version": "1.0.0",
			},
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Test by capturing log output
		var buf bytes.Buffer
		logger.SetOutput(&buf)

		logger.Info("test message")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["service"] != "slurm-exporter" {
			t.Errorf("Expected service field to be 'slurm-exporter', got %v", logEntry["service"])
		}

		if logEntry["version"] != "1.0.0" {
			t.Errorf("Expected version field to be '1.0.0', got %v", logEntry["version"])
		}
	})
}

func TestLoggerFileOutput(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "logger_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("WithValidFile", func(t *testing.T) {
		logFile := filepath.Join(tmpDir, "test.log")
		cfg := &config.LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "file",
			File:       logFile,
			MaxSize:    10,
			MaxAge:     7,
			MaxBackups: 3,
			Compress:   true,
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		logger.Info("test message")

		// Check if file exists
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("Expected log file to be created")
		}

		// Read file contents
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if !strings.Contains(string(content), "test message") {
			t.Error("Expected log file to contain test message")
		}

		// Test close
		if err := logger.Close(); err != nil {
			t.Errorf("Failed to close logger: %v", err)
		}
	})

	t.Run("WithFileOutputNoPath", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "file",
			File:   "", // Empty file path
		}

		_, err := NewLogger(cfg)
		if err == nil {
			t.Error("Expected error when file output is specified without file path")
		}
	})

	t.Run("WithNestedDirectory", func(t *testing.T) {
		logFile := filepath.Join(tmpDir, "nested", "dir", "test.log")
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "file",
			File:   logFile,
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		logger.Info("test message")

		// Check if file exists
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			t.Error("Expected log file to be created in nested directory")
		}
	})
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.LoggingConfig{
		Level:        "debug",
		Format:       "json",
		Output:       "stdout",
		SuppressHTTP: true,
	}

	logger, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	logger.SetOutput(&buf)

	t.Run("WithComponent", func(t *testing.T) {
		buf.Reset()
		logger.WithComponent("test-component").Info("test message")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["component"] != "test-component" {
			t.Errorf("Expected component field to be 'test-component', got %v", logEntry["component"])
		}
	})

	t.Run("WithCollector", func(t *testing.T) {
		buf.Reset()
		logger.WithCollector("cluster-collector").Info("test message")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["collector"] != "cluster-collector" {
			t.Errorf("Expected collector field to be 'cluster-collector', got %v", logEntry["collector"])
		}
	})

	t.Run("WithRequest", func(t *testing.T) {
		buf.Reset()
		logger.WithRequest("GET", "/metrics", "prometheus/1.0").Info("test message")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["http_method"] != "GET" {
			t.Errorf("Expected http_method field to be 'GET', got %v", logEntry["http_method"])
		}

		if logEntry["http_path"] != "/metrics" {
			t.Errorf("Expected http_path field to be '/metrics', got %v", logEntry["http_path"])
		}

		if logEntry["http_user_agent"] != "prometheus/1.0" {
			t.Errorf("Expected http_user_agent field to be 'prometheus/1.0', got %v", logEntry["http_user_agent"])
		}
	})

	t.Run("WithError", func(t *testing.T) {
		buf.Reset()
		testErr := fmt.Errorf("test error")
		logger.WithError(testErr).Error("error occurred")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["error"] != "test error" {
			t.Errorf("Expected error field to be 'test error', got %v", logEntry["error"])
		}
	})

	t.Run("WithDuration", func(t *testing.T) {
		buf.Reset()
		duration := 150 * time.Millisecond
		logger.WithDuration(duration).Info("operation completed")

		var logEntry map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
			t.Fatalf("Failed to parse log JSON: %v", err)
		}

		if logEntry["duration"] == nil {
			t.Error("Expected duration field to be present")
		}
	})

	t.Run("IsHTTPSuppressed", func(t *testing.T) {
		if !logger.IsHTTPSuppressed() {
			t.Error("Expected HTTP suppression to be true")
		}

		cfg2 := &config.LoggingConfig{
			Level:        "info",
			Format:       "json",
			Output:       "stdout",
			SuppressHTTP: false,
		}

		logger2, err := NewLogger(cfg2)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		if logger2.IsHTTPSuppressed() {
			t.Error("Expected HTTP suppression to be false")
		}
	})

	t.Run("SetLevel", func(t *testing.T) {
		logger.SetLevel(logrus.ErrorLevel)
		if logger.GetLevel() != logrus.ErrorLevel {
			t.Errorf("Expected log level Error, got %v", logger.GetLevel())
		}
	})
}

func TestLogLevels(t *testing.T) {
	testCases := []struct {
		configLevel string
		logrusLevel logrus.Level
	}{
		{"debug", logrus.DebugLevel},
		{"info", logrus.InfoLevel},
		{"warn", logrus.WarnLevel},
		{"error", logrus.ErrorLevel},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Level_%s", tc.configLevel), func(t *testing.T) {
			cfg := &config.LoggingConfig{
				Level:  tc.configLevel,
				Format: "json",
				Output: "stdout",
			}

			logger, err := NewLogger(cfg)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			if logger.GetLevel() != tc.logrusLevel {
				t.Errorf("Expected log level %v, got %v", tc.logrusLevel, logger.GetLevel())
			}
		})
	}
}

func TestLogFormats(t *testing.T) {
	testCases := []struct {
		format       string
		expectedType interface{}
		shouldError  bool
	}{
		{"json", &logrus.JSONFormatter{}, false},
		{"text", &logrus.TextFormatter{}, false},
		{"invalid", nil, true}, // Should error due to config validation
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Format_%s", tc.format), func(t *testing.T) {
			cfg := &config.LoggingConfig{
				Level:  "info",
				Format: tc.format,
				Output: "stdout",
			}

			logger, err := NewLogger(cfg)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for format '%s', but got none", tc.format)
				}
				return
			}

			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}

			switch tc.expectedType.(type) {
			case *logrus.JSONFormatter:
				if _, ok := logger.Formatter.(*logrus.JSONFormatter); !ok {
					t.Errorf("Expected JSON formatter for format '%s'", tc.format)
				}
			case *logrus.TextFormatter:
				if _, ok := logger.Formatter.(*logrus.TextFormatter); !ok {
					t.Errorf("Expected text formatter for format '%s'", tc.format)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	t.Run("FileOutput", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "logger_close_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		logFile := filepath.Join(tmpDir, "test.log")
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "file",
			File:   logFile,
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		logger.Info("test message")

		if err := logger.Close(); err != nil {
			t.Errorf("Failed to close file logger: %v", err)
		}
	})

	t.Run("StdoutOutput", func(t *testing.T) {
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}

		logger, err := NewLogger(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		// Should not error when closing stdout logger
		if err := logger.Close(); err != nil {
			t.Errorf("Unexpected error closing stdout logger: %v", err)
		}
	})
}
