// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jontk/slurm-exporter/internal/config"
)

// Logger wraps logrus.Logger with additional configuration
type Logger struct {
	*logrus.Logger
	config         *config.LoggingConfig
	constantFields logrus.Fields
}

// NewLogger creates a new configured logger instance
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	if cfg == nil {
		cfg = &config.LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	// Set formatter
	switch cfg.Format {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.DateTime,
			FullTimestamp:   true,
		})
	default:
		// Default to JSON for safety
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Set output
	var writer io.Writer
	switch cfg.Output {
	case "stdout":
		writer = os.Stdout
	case "stderr":
		writer = os.Stderr
	case "file":
		if cfg.File == "" {
			return nil, fmt.Errorf("log file path is required when output is 'file'")
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(cfg.File), 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Use lumberjack for log rotation
		writer = &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,    // MB
			MaxAge:     cfg.MaxAge,     // days
			MaxBackups: cfg.MaxBackups, // files
			Compress:   cfg.Compress,   // compress rotated files
		}
	default:
		writer = os.Stdout
	}

	logger.SetOutput(writer)

	// Store constant fields for later use
	var constantFields logrus.Fields
	if len(cfg.Fields) > 0 {
		constantFields = make(logrus.Fields)
		for k, v := range cfg.Fields {
			constantFields[k] = v
		}
	}

	return &Logger{
		Logger:         logger,
		config:         cfg,
		constantFields: constantFields,
	}, nil
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() logrus.Level {
	return l.Logger.GetLevel()
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level logrus.Level) {
	l.Logger.SetLevel(level)
}

// IsHTTPSuppressed returns whether HTTP request logging is suppressed
func (l *Logger) IsHTTPSuppressed() bool {
	return l.config.SuppressHTTP
}

// WithComponent creates a logger with a component field
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithCollector creates a logger with a collector field
func (l *Logger) WithCollector(collector string) *logrus.Entry {
	return l.WithField("collector", collector)
}

// WithRequest creates a logger with request fields
func (l *Logger) WithRequest(method, path, userAgent string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"http_method":     method,
		"http_path":       path,
		"http_user_agent": userAgent,
	})
}

// WithError creates a logger with error field
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithDuration creates a logger with duration field
func (l *Logger) WithDuration(duration interface{}) *logrus.Entry {
	return l.WithField("duration", duration)
}

// withConstantFields returns a logger entry with constant fields applied
func (l *Logger) withConstantFields() *logrus.Entry {
	if len(l.constantFields) > 0 {
		return l.WithFields(l.constantFields)
	}
	return logrus.NewEntry(l.Logger)
}

// Override key logging methods to include constant fields

// Debug logs a message at debug level with constant fields
func (l *Logger) Debug(args ...interface{}) {
	l.withConstantFields().Debug(args...)
}

// Info logs a message at info level with constant fields
func (l *Logger) Info(args ...interface{}) {
	l.withConstantFields().Info(args...)
}

// Warn logs a message at warn level with constant fields
func (l *Logger) Warn(args ...interface{}) {
	l.withConstantFields().Warn(args...)
}

// Error logs a message at error level with constant fields
func (l *Logger) Error(args ...interface{}) {
	l.withConstantFields().Error(args...)
}

// Fatal logs a message at fatal level with constant fields
func (l *Logger) Fatal(args ...interface{}) {
	l.withConstantFields().Fatal(args...)
}

// Debugf logs a formatted message at debug level with constant fields
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.withConstantFields().Debugf(format, args...)
}

// Infof logs a formatted message at info level with constant fields
func (l *Logger) Infof(format string, args ...interface{}) {
	l.withConstantFields().Infof(format, args...)
}

// Warnf logs a formatted message at warn level with constant fields
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.withConstantFields().Warnf(format, args...)
}

// Errorf logs a formatted message at error level with constant fields
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.withConstantFields().Errorf(format, args...)
}

// Fatalf logs a formatted message at fatal level with constant fields
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.withConstantFields().Fatalf(format, args...)
}

// Close closes any file outputs (for lumberjack)
func (l *Logger) Close() error {
	if l.config.Output == "file" {
		if closer, ok := l.Out.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}
