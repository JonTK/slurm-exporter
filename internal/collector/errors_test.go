// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestCollectionError(t *testing.T) {
	t.Run("ErrorInterface", func(t *testing.T) {
		baseErr := errors.New("base error")
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeConnection,
			Severity:  SeverityHigh,
			Message:   "Connection failed",
			Err:       baseErr,
			Timestamp: time.Now(),
		}

		// Test Error() method
		errStr := collErr.Error()
		if errStr == "" {
			t.Error("Error string should not be empty")
		}

		// Test Unwrap() method
		if collErr.Unwrap() != baseErr {
			t.Error("Unwrap should return the base error")
		}

		// Test Is() method
		otherErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeConnection,
		}
		if !collErr.Is(otherErr) {
			t.Error("Should match errors with same collector and type")
		}

		differentErr := &CollectionError{
			Collector: "other",
			Type:      ErrorTypeConnection,
		}
		if collErr.Is(differentErr) {
			t.Error("Should not match errors with different collector")
		}
	})

	t.Run("LogFields", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeAPI,
			Severity:  SeverityMedium,
			Message:   "API error",
			Timestamp: time.Now(),
			Retryable: true,
			Context: ErrorContext{
				Operation:  "get_jobs",
				Endpoint:   "/slurm/v1/jobs",
				HTTPStatus: 500,
				Duration:   100 * time.Millisecond,
				RequestID:  "req-123",
				Metadata: map[string]string{
					"user": "testuser",
				},
			},
		}

		fields := collErr.LogFields()

		// Check required fields
		if fields["collector"] != "test" {
			t.Errorf("Expected collector 'test', got '%v'", fields["collector"])
		}
		if fields["error_type"] != ErrorTypeAPI {
			t.Errorf("Expected error_type '%s', got '%v'", ErrorTypeAPI, fields["error_type"])
		}
		if fields["severity"] != SeverityMedium {
			t.Errorf("Expected severity '%s', got '%v'", SeverityMedium, fields["severity"])
		}
		if fields["retryable"] != true {
			t.Errorf("Expected retryable true, got '%v'", fields["retryable"])
		}

		// Check context fields
		if fields["operation"] != "get_jobs" {
			t.Errorf("Expected operation 'get_jobs', got '%v'", fields["operation"])
		}
		if fields["endpoint"] != "/slurm/v1/jobs" {
			t.Errorf("Expected endpoint '/slurm/v1/jobs', got '%v'", fields["endpoint"])
		}
		if fields["http_status"] != 500 {
			t.Errorf("Expected http_status 500, got '%v'", fields["http_status"])
		}
		if fields["duration_ms"] != int64(100) {
			t.Errorf("Expected duration_ms 100, got '%v'", fields["duration_ms"])
		}
		if fields["request_id"] != "req-123" {
			t.Errorf("Expected request_id 'req-123', got '%v'", fields["request_id"])
		}
		if fields["meta_user"] != "testuser" {
			t.Errorf("Expected meta_user 'testuser', got '%v'", fields["meta_user"])
		}
	})
}

func TestErrorBuilder(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	builder := NewErrorBuilder("test_collector", logger)

	t.Run("Connection", func(t *testing.T) {
		baseErr := errors.New("connection refused")
		err := builder.Connection(baseErr, "https://slurm.example.com", "Check network")

		if err.Collector != "test_collector" {
			t.Errorf("Expected collector 'test_collector', got '%s'", err.Collector)
		}
		if err.Type != ErrorTypeConnection {
			t.Errorf("Expected type '%s', got '%s'", ErrorTypeConnection, err.Type)
		}
		if err.Severity != SeverityHigh {
			t.Errorf("Expected severity '%s', got '%s'", SeverityHigh, err.Severity)
		}
		if !err.Retryable {
			t.Error("Connection errors should be retryable")
		}
		if err.Context.Endpoint != "https://slurm.example.com" {
			t.Errorf("Expected endpoint 'https://slurm.example.com', got '%s'", err.Context.Endpoint)
		}
		if len(err.Suggestions) != 1 || err.Suggestions[0] != "Check network" {
			t.Errorf("Expected suggestions ['Check network'], got %v", err.Suggestions)
		}
	})

	t.Run("Auth", func(t *testing.T) {
		baseErr := errors.New("unauthorized")
		err := builder.Auth(baseErr, "jwt", "Check token")

		if err.Type != ErrorTypeAuth {
			t.Errorf("Expected type '%s', got '%s'", ErrorTypeAuth, err.Type)
		}
		if err.Severity != SeverityCritical {
			t.Errorf("Expected severity '%s', got '%s'", SeverityCritical, err.Severity)
		}
		if err.Retryable {
			t.Error("Auth errors should not be retryable")
		}
		if err.Context.Metadata["auth_type"] != "jwt" {
			t.Errorf("Expected auth_type 'jwt', got '%s'", err.Context.Metadata["auth_type"])
		}
	})

	t.Run("Timeout", func(t *testing.T) {
		baseErr := errors.New("context deadline exceeded")
		duration := 30 * time.Second
		err := builder.Timeout(baseErr, "fetch_jobs", duration, "Increase timeout")

		if err.Type != ErrorTypeTimeout {
			t.Errorf("Expected type '%s', got '%s'", ErrorTypeTimeout, err.Type)
		}
		if err.Severity != SeverityMedium {
			t.Errorf("Expected severity '%s', got '%s'", SeverityMedium, err.Severity)
		}
		if !err.Retryable {
			t.Error("Timeout errors should be retryable")
		}
		if err.Context.Operation != "fetch_jobs" {
			t.Errorf("Expected operation 'fetch_jobs', got '%s'", err.Context.Operation)
		}
		if err.Context.Duration != duration {
			t.Errorf("Expected duration %v, got %v", duration, err.Context.Duration)
		}
	})

	t.Run("API", func(t *testing.T) {
		baseErr := errors.New("internal server error")
		err := builder.API(baseErr, "/api/jobs", 500, "req-456", "Try again")

		if err.Type != ErrorTypeAPI {
			t.Errorf("Expected type '%s', got '%s'", ErrorTypeAPI, err.Type)
		}
		if err.Severity != SeverityHigh {
			t.Errorf("Expected severity '%s', got '%s'", SeverityHigh, err.Severity)
		}
		if !err.Retryable {
			t.Error("5xx API errors should be retryable")
		}
		if err.Context.HTTPStatus != 500 {
			t.Errorf("Expected http_status 500, got %d", err.Context.HTTPStatus)
		}
		if err.Context.RequestID != "req-456" {
			t.Errorf("Expected request_id 'req-456', got '%s'", err.Context.RequestID)
		}

		// Test 404 error (should not be retryable)
		err404 := builder.API(baseErr, "/api/jobs", 404, "req-457")
		if err404.Retryable {
			t.Error("404 API errors should not be retryable")
		}
		if err404.Severity != SeverityLow {
			t.Errorf("Expected severity '%s' for 404, got '%s'", SeverityLow, err404.Severity)
		}

		// Test 401 error (should be critical)
		err401 := builder.API(baseErr, "/api/jobs", 401, "req-458")
		if err401.Retryable {
			t.Error("401 API errors should not be retryable")
		}
		if err401.Severity != SeverityCritical {
			t.Errorf("Expected severity '%s' for 401, got '%s'", SeverityCritical, err401.Severity)
		}
	})
}

func TestErrorAnalyzer(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	analyzer := NewErrorAnalyzer(logger)

	testCases := []struct {
		name         string
		errMsg       string
		expectedType ErrorType
		retryable    bool
	}{
		{
			name:         "ConnectionRefused",
			errMsg:       "dial tcp: connection refused",
			expectedType: ErrorTypeConnection,
			retryable:    true,
		},
		{
			name:         "Timeout",
			errMsg:       "context deadline exceeded",
			expectedType: ErrorTypeTimeout,
			retryable:    true,
		},
		{
			name:         "Unauthorized",
			errMsg:       "401 unauthorized",
			expectedType: ErrorTypeAuth,
			retryable:    false,
		},
		{
			name:         "RateLimit",
			errMsg:       "429 too many requests",
			expectedType: ErrorTypeRateLimit,
			retryable:    true,
		},
		{
			name:         "Permission",
			errMsg:       "403 forbidden",
			expectedType: ErrorTypePermission,
			retryable:    false,
		},
		{
			name:         "Unknown",
			errMsg:       "some random error",
			expectedType: ErrorTypeInternal,
			retryable:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseErr := errors.New(tc.errMsg)
			collErr := analyzer.AnalyzeError(baseErr, "test_collector")

			if collErr == nil {
				t.Fatal("Expected analyzed error, got nil")
			}

			if collErr.Type != tc.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tc.expectedType, collErr.Type)
			}

			if collErr.Retryable != tc.retryable {
				t.Errorf("Expected retryable %v, got %v", tc.retryable, collErr.Retryable)
			}

			if collErr.Collector != "test_collector" {
				t.Errorf("Expected collector 'test_collector', got '%s'", collErr.Collector)
			}

			if len(collErr.Suggestions) == 0 {
				t.Error("Expected suggestions to be provided")
			}
		})
	}

	t.Run("ExistingCollectionError", func(t *testing.T) {
		originalErr := &CollectionError{
			Collector: "original",
			Type:      ErrorTypeAPI,
			Severity:  SeverityHigh,
		}

		result := analyzer.AnalyzeError(originalErr, "test_collector")

		// Should return the original error unchanged
		if result != originalErr {
			t.Error("Should return original CollectionError unchanged")
		}
	})

	t.Run("NilError", func(t *testing.T) {
		result := analyzer.AnalyzeError(nil, "test_collector")
		if result != nil {
			t.Error("Should return nil for nil error")
		}
	})
}

func TestErrorRecoveryHandler(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	handler := NewErrorRecoveryHandler(logger)
	ctx := context.Background()

	t.Run("NilError", func(t *testing.T) {
		err := handler.HandleError(ctx, nil)
		if err != nil {
			t.Error("Should return nil for nil error")
		}
	})

	t.Run("ConnectionError", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeConnection,
			Severity:  SeverityHigh,
			Message:   "Connection failed",
			Retryable: true,
		}

		err := handler.HandleError(ctx, collErr)
		// Connection errors are returned as-is for circuit breaker handling
		if err != collErr {
			t.Error("Should return original connection error")
		}
	})

	t.Run("TimeoutError", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeTimeout,
			Severity:  SeverityMedium,
			Message:   "Timeout occurred",
			Retryable: true,
		}

		err := handler.HandleError(ctx, collErr)
		// Timeout errors are returned as-is for circuit breaker handling
		if err != collErr {
			t.Error("Should return original timeout error")
		}
	})

	t.Run("RateLimitError", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeRateLimit,
			Severity:  SeverityMedium,
			Message:   "Rate limited",
			Retryable: true,
			Context: ErrorContext{
				Metadata: map[string]string{
					"retry_after": "100ms",
				},
			},
		}

		// Test with context that times out immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := handler.HandleError(ctx, collErr)
		// Should return context error due to cancellation
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("AuthError", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeAuth,
			Severity:  SeverityCritical,
			Message:   "Authentication failed",
			Retryable: false,
		}

		err := handler.HandleError(ctx, collErr)
		// Auth errors require manual intervention
		if err != collErr {
			t.Error("Should return original auth error")
		}
	})

	t.Run("OtherError", func(t *testing.T) {
		collErr := &CollectionError{
			Collector: "test",
			Type:      ErrorTypeParsing,
			Severity:  SeverityMedium,
			Message:   "Parsing failed",
			Retryable: false,
		}

		err := handler.HandleError(ctx, collErr)
		// Other errors are returned as-is
		if err != collErr {
			t.Error("Should return original parsing error")
		}
	})
}

func TestErrorClassification(t *testing.T) {
	testCases := []struct {
		name     string
		errMsg   string
		checkFn  func(string) bool
		expected bool
	}{
		// Connection errors
		{"ConnectionRefused", "connection refused", isConnectionError, true},
		{"ConnectionReset", "connection reset by peer", isConnectionError, true},
		{"NoSuchHost", "no such host", isConnectionError, true},
		{"NetworkUnreachable", "network unreachable", isConnectionError, true},
		{"DialTCP", "dial tcp 127.0.0.1:6820: connection refused", isConnectionError, true},
		{"NotConnection", "some other error", isConnectionError, false},

		// Timeout errors
		{"Timeout", "operation timeout", isTimeoutError, true},
		{"DeadlineExceeded", "context deadline exceeded", isTimeoutError, true},
		{"RequestTimeout", "request timeout", isTimeoutError, true},
		{"NotTimeout", "some other error", isTimeoutError, false},

		// Auth errors
		{"Unauthorized", "401 unauthorized", isAuthError, true},
		{"AuthFailed", "authentication failed", isAuthError, true},
		{"InvalidToken", "invalid token provided", isAuthError, true},
		{"TokenExpired", "token expired", isAuthError, true},
		{"NotAuth", "some other error", isAuthError, false},

		// Rate limit errors
		{"RateLimit", "rate limit exceeded", isRateLimitError, true},
		{"TooManyRequests", "429 too many requests", isRateLimitError, true},
		{"QuotaExceeded", "quota exceeded", isRateLimitError, true},
		{"Throttle", "request throttled", isRateLimitError, true},
		{"NotRateLimit", "some other error", isRateLimitError, false},

		// Permission errors
		{"PermissionDenied", "permission denied", isPermissionError, true},
		{"Forbidden", "403 forbidden", isPermissionError, true},
		{"AccessDenied", "access denied to resource", isPermissionError, true},
		{"NotAuthorized", "not authorized for this action", isPermissionError, true},
		{"NotPermission", "some other error", isPermissionError, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.checkFn(tc.errMsg)
			if result != tc.expected {
				t.Errorf("Expected %v for '%s', got %v", tc.expected, tc.errMsg, result)
			}
		})
	}
}

func TestContainsHelper(t *testing.T) {
	testCases := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "lo wo", true},
		{"hello world", "xyz", false},
		{"timeout", "timeout", true},
		{"connection timeout", "timeout", true},
		{"", "timeout", false},
		{"timeout", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.str+"_contains_"+tc.substr, func(t *testing.T) {
			result := contains(tc.str, tc.substr)
			if result != tc.expected {
				t.Errorf("Expected contains('%s', '%s') = %v, got %v",
					tc.str, tc.substr, tc.expected, result)
			}
		})
	}
}
