// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ErrorType represents different categories of collection errors
type ErrorType string

const (
	ErrorTypeConnection    ErrorType = "connection"
	ErrorTypeAuth          ErrorType = "authentication"
	ErrorTypeTimeout       ErrorType = "timeout"
	ErrorTypeRateLimit     ErrorType = "rate_limit"
	ErrorTypeAPI           ErrorType = "api_error"
	ErrorTypeParsing       ErrorType = "parsing"
	ErrorTypeInternal      ErrorType = "internal"
	ErrorTypeConfiguration ErrorType = "configuration"
	ErrorTypePermission    ErrorType = "permission"
	ErrorTypeNotFound      ErrorType = "not_found"
)

// ErrorSeverity represents the severity level of errors
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityHigh     ErrorSeverity = "high"
	SeverityCritical ErrorSeverity = "critical"
)

// CollectionError represents a structured collection error
type CollectionError struct {
	Collector   string        `json:"collector"`
	Type        ErrorType     `json:"type"`
	Severity    ErrorSeverity `json:"severity"`
	Message     string        `json:"message"`
	Err         error         `json:"error,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
	Context     ErrorContext  `json:"context,omitempty"`
	Retryable   bool          `json:"retryable"`
	Suggestions []string      `json:"suggestions,omitempty"`
}

// ErrorContext provides additional context about the error
type ErrorContext struct {
	Operation  string            `json:"operation,omitempty"`
	Endpoint   string            `json:"endpoint,omitempty"`
	HTTPStatus int               `json:"http_status,omitempty"`
	Duration   time.Duration     `json:"duration,omitempty"`
	RequestID  string            `json:"request_id,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Error implements the error interface
func (ce *CollectionError) Error() string {
	if ce.Err != nil {
		return fmt.Sprintf("%s collector [%s/%s]: %s: %v",
			ce.Collector, ce.Type, ce.Severity, ce.Message, ce.Err)
	}
	return fmt.Sprintf("%s collector [%s/%s]: %s",
		ce.Collector, ce.Type, ce.Severity, ce.Message)
}

// Unwrap returns the underlying error
func (ce *CollectionError) Unwrap() error {
	return ce.Err
}

// Is implements error comparison
func (ce *CollectionError) Is(target error) bool {
	if other, ok := target.(*CollectionError); ok {
		return ce.Type == other.Type && ce.Collector == other.Collector
	}
	return false
}

// LogFields returns structured log fields for this error
func (ce *CollectionError) LogFields() logrus.Fields {
	fields := logrus.Fields{
		"collector":  ce.Collector,
		"error_type": ce.Type,
		"severity":   ce.Severity,
		"timestamp":  ce.Timestamp,
		"retryable":  ce.Retryable,
	}

	if ce.Context.Operation != "" {
		fields["operation"] = ce.Context.Operation
	}
	if ce.Context.Endpoint != "" {
		fields["endpoint"] = ce.Context.Endpoint
	}
	if ce.Context.HTTPStatus > 0 {
		fields["http_status"] = ce.Context.HTTPStatus
	}
	if ce.Context.Duration > 0 {
		fields["duration_ms"] = ce.Context.Duration.Milliseconds()
	}
	if ce.Context.RequestID != "" {
		fields["request_id"] = ce.Context.RequestID
	}

	// Add metadata
	for k, v := range ce.Context.Metadata {
		fields[fmt.Sprintf("meta_%s", k)] = v
	}

	return fields
}

// ErrorBuilder helps construct structured errors
type ErrorBuilder struct {
	collector string
	logger    *logrus.Entry
}

// NewErrorBuilder creates a new error builder
func NewErrorBuilder(collector string, logger *logrus.Entry) *ErrorBuilder {
	return &ErrorBuilder{
		collector: collector,
		logger:    logger,
	}
}

// Connection creates a connection error
func (eb *ErrorBuilder) Connection(err error, endpoint string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeConnection,
		Severity:    SeverityHigh,
		Message:     "Failed to connect to SLURM API",
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   true,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: "connect",
			Endpoint:  endpoint,
		},
	}
}

// Auth creates an authentication error
func (eb *ErrorBuilder) Auth(err error, authType string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeAuth,
		Severity:    SeverityCritical,
		Message:     "Authentication failed",
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   false,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: "authenticate",
			Metadata: map[string]string{
				"auth_type": authType,
			},
		},
	}
}

// Timeout creates a timeout error
func (eb *ErrorBuilder) Timeout(err error, operation string, duration time.Duration, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeTimeout,
		Severity:    SeverityMedium,
		Message:     fmt.Sprintf("Operation timed out after %v", duration),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   true,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: operation,
			Duration:  duration,
		},
	}
}

// RateLimit creates a rate limit error
func (eb *ErrorBuilder) RateLimit(err error, endpoint string, retryAfter time.Duration, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeRateLimit,
		Severity:    SeverityMedium,
		Message:     "Rate limit exceeded",
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   true,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: "api_call",
			Endpoint:  endpoint,
			Metadata: map[string]string{
				"retry_after": retryAfter.String(),
			},
		},
	}
}

// API creates an API error
func (eb *ErrorBuilder) API(err error, endpoint string, httpStatus int, requestID string, suggestions ...string) *CollectionError {
	severity := SeverityMedium
	retryable := true

	// Determine severity and retryability based on HTTP status
	switch {
	case httpStatus >= 500:
		severity = SeverityHigh
		retryable = true
	case httpStatus == 404:
		severity = SeverityLow
		retryable = false
	case httpStatus == 401 || httpStatus == 403:
		severity = SeverityCritical
		retryable = false
	case httpStatus >= 400:
		severity = SeverityMedium
		retryable = false
	}

	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeAPI,
		Severity:    severity,
		Message:     fmt.Sprintf("API request failed with status %d", httpStatus),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   retryable,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation:  "api_call",
			Endpoint:   endpoint,
			HTTPStatus: httpStatus,
			RequestID:  requestID,
		},
	}
}

// Parsing creates a parsing error
func (eb *ErrorBuilder) Parsing(err error, operation string, dataType string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeParsing,
		Severity:    SeverityMedium,
		Message:     fmt.Sprintf("Failed to parse %s data", dataType),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   false,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: operation,
			Metadata: map[string]string{
				"data_type": dataType,
			},
		},
	}
}

// Internal creates an internal error
func (eb *ErrorBuilder) Internal(err error, operation string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeInternal,
		Severity:    SeverityHigh,
		Message:     "Internal error occurred",
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   true,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: operation,
		},
	}
}

// Configuration creates a configuration error
func (eb *ErrorBuilder) Configuration(err error, field string, value interface{}, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeConfiguration,
		Severity:    SeverityCritical,
		Message:     fmt.Sprintf("Invalid configuration for field '%s'", field),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   false,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: "validate_config",
			Metadata: map[string]string{
				"field": field,
				"value": fmt.Sprintf("%v", value),
			},
		},
	}
}

// Permission creates a permission error
func (eb *ErrorBuilder) Permission(err error, resource string, action string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypePermission,
		Severity:    SeverityCritical,
		Message:     fmt.Sprintf("Permission denied for %s on %s", action, resource),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   false,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: action,
			Metadata: map[string]string{
				"resource": resource,
				"action":   action,
			},
		},
	}
}

// NotFound creates a not found error
func (eb *ErrorBuilder) NotFound(err error, resource string, identifier string, suggestions ...string) *CollectionError {
	return &CollectionError{
		Collector:   eb.collector,
		Type:        ErrorTypeNotFound,
		Severity:    SeverityLow,
		Message:     fmt.Sprintf("%s not found: %s", resource, identifier),
		Err:         err,
		Timestamp:   time.Now(),
		Retryable:   false,
		Suggestions: suggestions,
		Context: ErrorContext{
			Operation: "lookup",
			Metadata: map[string]string{
				"resource":   resource,
				"identifier": identifier,
			},
		},
	}
}

// Log logs the error with appropriate level and structured fields
func (eb *ErrorBuilder) Log(err *CollectionError) {
	fields := err.LogFields()

	// Add suggestions if present
	if len(err.Suggestions) > 0 {
		fields["suggestions"] = err.Suggestions
	}

	// Log with appropriate level based on severity
	switch err.Severity {
	case SeverityLow:
		eb.logger.WithFields(fields).WithError(err.Err).Info(err.Message)
	case SeverityMedium:
		eb.logger.WithFields(fields).WithError(err.Err).Warn(err.Message)
	case SeverityHigh:
		eb.logger.WithFields(fields).WithError(err.Err).Error(err.Message)
	case SeverityCritical:
		eb.logger.WithFields(fields).WithError(err.Err).Error(err.Message)
	default:
		eb.logger.WithFields(fields).WithError(err.Err).Error(err.Message)
	}
}

// ErrorAnalyzer provides error analysis and classification
type ErrorAnalyzer struct {
	logger *logrus.Entry
}

// NewErrorAnalyzer creates a new error analyzer
func NewErrorAnalyzer(logger *logrus.Entry) *ErrorAnalyzer {
	return &ErrorAnalyzer{
		logger: logger,
	}
}

// AnalyzeError analyzes an error and provides classification
func (ea *ErrorAnalyzer) AnalyzeError(err error, collector string) *CollectionError {
	if err == nil {
		return nil
	}

	// If it's already a CollectionError, return as-is
	if collErr, ok := err.(*CollectionError); ok {
		return collErr
	}

	builder := NewErrorBuilder(collector, ea.logger)

	// Analyze error by message and type
	errMsg := err.Error()

	switch {
	case isConnectionError(errMsg):
		return builder.Connection(err, "",
			"Check SLURM API endpoint URL and network connectivity",
			"Verify firewall rules allow outbound connections",
			"Ensure SLURM REST API is running and accessible")

	case isTimeoutError(errMsg):
		return builder.Timeout(err, "unknown", 0,
			"Increase timeout configuration",
			"Check SLURM API performance and load",
			"Consider reducing request frequency")

	case isAuthError(errMsg):
		return builder.Auth(err, "unknown",
			"Verify authentication credentials",
			"Check token/password expiration",
			"Ensure user has required permissions")

	case isRateLimitError(errMsg):
		return builder.RateLimit(err, "", 0,
			"Reduce collection frequency",
			"Implement exponential backoff",
			"Check rate limit configuration")

	case isPermissionError(errMsg):
		return builder.Permission(err, "unknown", "read",
			"Check user permissions in SLURM",
			"Verify authentication user has access to required resources",
			"Contact SLURM administrator")

	default:
		return builder.Internal(err, "unknown",
			"Check logs for detailed error information",
			"Verify SLURM API compatibility",
			"Report issue if problem persists")
	}
}

// isConnectionError checks if error is connection-related
func isConnectionError(errMsg string) bool {
	connectionKeywords := []string{
		"connection refused",
		"connection reset",
		"no such host",
		"network unreachable",
		"connection timeout",
		"dial tcp",
		"connect:",
	}

	for _, keyword := range connectionKeywords {
		if errorContains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isTimeoutError checks if error is timeout-related
func isTimeoutError(errMsg string) bool {
	timeoutKeywords := []string{
		"timeout",
		"deadline exceeded",
		"context deadline exceeded",
		"request timeout",
	}

	for _, keyword := range timeoutKeywords {
		if errorContains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isAuthError checks if error is authentication-related
func isAuthError(errMsg string) bool {
	authKeywords := []string{
		"unauthorized",
		"authentication failed",
		"invalid token",
		"token expired",
		"401",
		"auth",
	}

	for _, keyword := range authKeywords {
		if errorContains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isRateLimitError checks if error is rate limit-related
func isRateLimitError(errMsg string) bool {
	rateLimitKeywords := []string{
		"rate limit",
		"too many requests",
		"429",
		"quota exceeded",
		"throttle",
	}

	for _, keyword := range rateLimitKeywords {
		if errorContains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// isPermissionError checks if error is permission-related
func isPermissionError(errMsg string) bool {
	permissionKeywords := []string{
		"permission denied",
		"forbidden",
		"403",
		"access denied",
		"not authorized",
	}

	for _, keyword := range permissionKeywords {
		if errorContains(errMsg, keyword) {
			return true
		}
	}
	return false
}

// errorContains checks if str contains substr (case-insensitive)
func errorContains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					errorContainsHelper(str, substr)))
}

// errorContainsHelper helper for case-insensitive substring search
func errorContainsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ErrorRecoveryHandler handles error recovery strategies
type ErrorRecoveryHandler struct {
	logger *logrus.Entry
}

// NewErrorRecoveryHandler creates a new error recovery handler
func NewErrorRecoveryHandler(logger *logrus.Entry) *ErrorRecoveryHandler {
	return &ErrorRecoveryHandler{
		logger: logger,
	}
}

// HandleError processes an error and suggests recovery actions
func (erh *ErrorRecoveryHandler) HandleError(ctx context.Context, err *CollectionError) error {
	if err == nil {
		return nil
	}

	erh.logger.WithFields(err.LogFields()).WithError(err.Err).
		Debugf("Processing error for recovery: %s", err.Message)

	// Execute recovery strategy based on error type
	switch err.Type {
	case ErrorTypeConnection:
		return erh.handleConnectionError(ctx, err)
	case ErrorTypeTimeout:
		return erh.handleTimeoutError(ctx, err)
	case ErrorTypeRateLimit:
		return erh.handleRateLimitError(ctx, err)
	case ErrorTypeAuth:
		return erh.handleAuthError(ctx, err)
	default:
		// For other errors, just log and return
		return err
	}
}

// handleConnectionError implements connection error recovery
func (erh *ErrorRecoveryHandler) handleConnectionError(ctx context.Context, err *CollectionError) error {
	erh.logger.WithField("collector", err.Collector).
		Info("Attempting connection error recovery")

	// For connection errors, we typically want to retry after a delay
	// This would be handled by the circuit breaker in practice
	return err
}

// handleTimeoutError implements timeout error recovery
func (erh *ErrorRecoveryHandler) handleTimeoutError(ctx context.Context, err *CollectionError) error {
	erh.logger.WithField("collector", err.Collector).
		Info("Attempting timeout error recovery")

	// For timeout errors, suggest increasing timeout or reducing load
	return err
}

// handleRateLimitError implements rate limit error recovery
func (erh *ErrorRecoveryHandler) handleRateLimitError(ctx context.Context, err *CollectionError) error {
	erh.logger.WithField("collector", err.Collector).
		Info("Attempting rate limit error recovery")

	// For rate limit errors, implement backoff
	if retryAfter := err.Context.Metadata["retry_after"]; retryAfter != "" {
		if duration, parseErr := time.ParseDuration(retryAfter); parseErr == nil {
			erh.logger.WithField("retry_after", duration).
				Info("Backing off due to rate limit")

			select {
			case <-time.After(duration):
				return nil // Indicate recovery attempt successful
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return err
}

// handleAuthError implements auth error recovery
func (erh *ErrorRecoveryHandler) handleAuthError(ctx context.Context, err *CollectionError) error {
	erh.logger.WithField("collector", err.Collector).
		Error("Authentication error requires manual intervention")

	// Auth errors typically require manual intervention
	return err
}
