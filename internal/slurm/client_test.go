// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package slurm

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.SLURMConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond, // Short timeout for tests
				RetryAttempts: 0,                      // No retries to avoid long waits
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config - empty base URL",
			config: &config.SLURMConfig{
				BaseURL:       "",
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid config - negative timeout",
			config: &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       -1 * time.Second,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client with no error")
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

func TestNewConnectionPool(t *testing.T) {
	t.Skip("Skipping test that requires network connection - no mock SLURM server available")

	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond, // Short timeout for tests
		RetryAttempts: 0,                      // No retries to avoid long waits
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	tests := []struct {
		name     string
		poolSize int
		wantErr  bool
	}{
		{
			name:     "valid pool size",
			poolSize: 3,
			wantErr:  false,
		},
		{
			name:     "zero pool size - should use default",
			poolSize: 0,
			wantErr:  false,
		},
		{
			name:     "negative pool size - should use default",
			poolSize: -1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewConnectionPool(cfg, tt.poolSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConnectionPool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if pool == nil {
					t.Error("NewConnectionPool() returned nil pool with no error")
				} else {
					expectedSize := tt.poolSize
					if expectedSize <= 0 {
						expectedSize = 5 // Default pool size
					}
					if len(pool.clients) != expectedSize {
						t.Errorf("NewConnectionPool() pool size = %d, want %d", len(pool.clients), expectedSize)
					}
				}
			}
		})
	}
}

func TestConnectionPoolGetClient(t *testing.T) {
	t.Skip("Skipping test that requires network connection - no mock SLURM server available")

	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond, // Short timeout for tests
		RetryAttempts: 0,                      // No retries to avoid long waits
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	pool, err := NewConnectionPool(cfg, 3)
	if err != nil {
		t.Fatalf("NewConnectionPool() error = %v", err)
	}

	// Test round-robin behavior
	clients := make([]*Client, 6)
	for i := 0; i < 6; i++ {
		clients[i] = pool.GetClient()
	}

	// Verify round-robin: clients should repeat in order
	if clients[0] != clients[3] {
		t.Error("Pool should return clients in round-robin order")
	}
	if clients[1] != clients[4] {
		t.Error("Pool should return clients in round-robin order")
	}
	if clients[2] != clients[5] {
		t.Error("Pool should return clients in round-robin order")
	}
}

func TestClientConnectionStatus(t *testing.T) {
	t.Skip("Skipping test that requires network connection - no mock SLURM server available")

	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond, // Short timeout for tests
		RetryAttempts: 0,                      // No retries to avoid long waits
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Initially, connection status might be false due to failed test connection
	// but the client should still be created
	if client.GetRetryCount() < 0 {
		t.Error("Retry count should be non-negative")
	}

	// Test ResetRetryCount
	client.ResetRetryCount()
	if client.GetRetryCount() != 0 {
		t.Error("Retry count should be 0 after reset")
	}
}

func TestClientContextCancellation(t *testing.T) {
	t.Skip("Skipping test that requires network connection - no mock SLURM server available")

	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond, // Short timeout for tests
		RetryAttempts: 0,                      // No retries to avoid long waits
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = client.TestConnection(ctx)
	if err == nil {
		t.Error("TestConnection should fail with cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestClientRateLimiting(t *testing.T) {
	t.Skip("Skipping test that requires network connection - no mock SLURM server available")

	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond, // Short timeout for tests
		RetryAttempts: 0,                      // No retries for faster test
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 1.0, // Very low rate for testing
			BurstSize:         1,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Make multiple rapid requests - should be rate limited
	start := time.Now()

	ctx := context.Background()
	_ = client.TestConnection(ctx)
	_ = client.TestConnection(ctx)

	elapsed := time.Since(start)

	// With rate limiting, the second call should take at least 1 second
	if elapsed < 900*time.Millisecond {
		t.Errorf("Rate limiting not working properly, elapsed time: %v", elapsed)
	}
}

// TestClientGettersAndSetters tests simple getter/setter methods
func TestClientGettersAndSetters(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Test IsConnected
	// Note: Connection might fail in test environment, but method should work
	_ = client.IsConnected()

	// Test GetLastError
	// Note: Error might be present from failed connection test
	_ = client.GetLastError()

	// Test GetRetryCount and ResetRetryCount
	initialCount := client.GetRetryCount()
	if initialCount < 0 {
		t.Errorf("GetRetryCount() = %d, should be non-negative", initialCount)
	}

	client.ResetRetryCount()
	if client.GetRetryCount() != 0 {
		t.Errorf("GetRetryCount() after reset = %d, want 0", client.GetRetryCount())
	}

	// Test GetSlurmClient
	slurmClient := client.GetSlurmClient()
	if slurmClient == nil {
		t.Error("GetSlurmClient() returned nil")
	}
}

// TestConnectionPoolRoundRobin tests connection pool round-robin behavior without network
func TestConnectionPoolRoundRobin(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	pool, err := NewConnectionPool(cfg, 3)
	if err != nil {
		t.Fatalf("NewConnectionPool() error = %v", err)
	}

	if len(pool.clients) != 3 {
		t.Errorf("Pool size = %d, want 3", len(pool.clients))
	}

	// Test round-robin behavior
	clients := make([]*Client, 6)
	for i := 0; i < 6; i++ {
		clients[i] = pool.GetClient()
		if clients[i] == nil {
			t.Errorf("GetClient() %d returned nil", i)
		}
	}

	// Verify round-robin: clients should repeat in order
	if clients[0] != clients[3] {
		t.Error("Pool should return first client again after 3 calls")
	}
	if clients[1] != clients[4] {
		t.Error("Pool should return second client again after 3 calls")
	}
	if clients[2] != clients[5] {
		t.Error("Pool should return third client again after 3 calls")
	}
}

// TestConnectionPoolDefaultSize tests default pool size
func TestConnectionPoolDefaultSize(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	tests := []struct {
		name         string
		poolSize     int
		expectedSize int
	}{
		{
			name:         "zero pool size uses default",
			poolSize:     0,
			expectedSize: 5,
		},
		{
			name:         "negative pool size uses default",
			poolSize:     -5,
			expectedSize: 5,
		},
		{
			name:         "explicit size",
			poolSize:     10,
			expectedSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewConnectionPool(cfg, tt.poolSize)
			if err != nil {
				t.Fatalf("NewConnectionPool() error = %v", err)
			}

			if len(pool.clients) != tt.expectedSize {
				t.Errorf("Pool size = %d, want %d", len(pool.clients), tt.expectedSize)
			}
		})
	}
}

// TestClientClose tests the Close method
func TestClientClose(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// After close, connection should be marked as disconnected
	if client.IsConnected() {
		t.Error("Client should be disconnected after Close()")
	}
}

// TestClientRetryLogic tests the retry mechanism
func TestClientRetryLogic(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientTimeout tests timeout handling
func TestClientTimeout(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	cfg := &config.SLURMConfig{
		BaseURL:       slowServer.URL,
		APIVersion:    "v0.0.42",
		Timeout:       500 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    100 * time.Millisecond,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 1000.0,
			BurstSize:         100,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx := context.Background()
	_, err = client.GetJobs(ctx)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestClientErrorPropagation tests error handling and propagation
func TestClientErrorPropagation(t *testing.T) {
	tests := []struct {
		name          string
		responseCode  int
		responseBody  string
		wantErr       bool
		errorContains string
	}{
		{
			name:          "500 internal server error",
			responseCode:  500,
			responseBody:  `{"error": "Internal server error"}`,
			wantErr:       true,
			errorContains: "failed",
		},
		{
			name:          "503 service unavailable",
			responseCode:  503,
			responseBody:  `{"error": "Service unavailable"}`,
			wantErr:       true,
			errorContains: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/slurm/v0.0.42/ping" {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer testServer.Close()

			cfg := &config.SLURMConfig{
				BaseURL:       testServer.URL,
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    100 * time.Millisecond,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 1000.0,
					BurstSize:         100,
				},
			}

			client, err := NewClient(cfg)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			defer func() { _ = client.Close() }()

			ctx := context.Background()
			_, err = client.GetJobs(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetJobs() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errorContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing %q, got %v", tt.errorContains, err)
				}
			}
		})
	}
}

// TestClientConnectionStatusUpdate tests connection status updates
func TestClientConnectionStatusUpdate(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientMultipleConcurrentRequests tests concurrent request handling
func TestClientMultipleConcurrentRequests(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientRateLimitingEnforcement tests rate limiting behavior
func TestClientRateLimitingEnforcement(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientContextCancellationPropagation tests context cancellation
func TestClientContextCancellationPropagation(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slurm/v0.0.42/ping" {
			w.WriteHeader(http.StatusOK)
			return
		}
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	cfg := &config.SLURMConfig{
		BaseURL:       testServer.URL,
		APIVersion:    "v0.0.42",
		Timeout:       10 * time.Second,
		RetryAttempts: 0,
		RetryDelay:    100 * time.Millisecond,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 1000.0,
			BurstSize:         100,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	_, err = client.GetJobs(ctx)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

// TestClientNilResultHandling tests handling of nil API responses
func TestClientNilResultHandling(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientRetryWithExponentialBackoff tests exponential backoff in retries
func TestClientRetryWithExponentialBackoff(t *testing.T) {
	t.Skip("Skipping - HTTP mock approach not compatible with slurm-client library parsing")
}

// TestClientNewClientWithAdapters tests client creation with adapters enabled/disabled
func TestClientNewClientWithAdapters(t *testing.T) {
	tests := []struct {
		name        string
		useAdapters bool
		apiVersion  string
		wantErr     bool
	}{
		{
			name:        "adapters enabled with version",
			useAdapters: true,
			apiVersion:  "v0.0.42",
			wantErr:     false,
		},
		{
			name:        "adapters enabled without version",
			useAdapters: true,
			apiVersion:  "",
			wantErr:     false,
		},
		{
			name:        "adapters disabled with version",
			useAdapters: false,
			apiVersion:  "v0.0.42",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    tt.apiVersion,
				UseAdapters:   tt.useAdapters,
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client without error")
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientRetryCount tests retry count management
func TestClientRetryCount(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Initial retry count should be 0
	if client.GetRetryCount() != 0 {
		t.Errorf("Initial retry count = %d, want 0", client.GetRetryCount())
	}

	// Reset should keep it at 0
	client.ResetRetryCount()
	if client.GetRetryCount() != 0 {
		t.Errorf("Reset retry count = %d, want 0", client.GetRetryCount())
	}
}

// TestClientConnectionError tests connection error handling
func TestClientConnectionError(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "http://localhost:99999",
		APIVersion:    "v0.0.42",
		Timeout:       500 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    100 * time.Millisecond,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Should not be connected after failure
	if client.IsConnected() {
		t.Error("Client should not be connected after failed connection")
	}

	// Should have last error set
	if client.GetLastError() == nil {
		t.Error("LastError should be set after failed connection")
	}
}

// TestClientRateLimitConfig tests rate limiting configuration
func TestClientRateLimitConfig(t *testing.T) {
	tests := []struct {
		name        string
		rateLimit   config.RateLimitConfig
		wantErr     bool
		description string
	}{
		{
			name: "high rate limit",
			rateLimit: config.RateLimitConfig{
				RequestsPerSecond: 1000.0,
				BurstSize:         500,
			},
			wantErr:     false,
			description: "allows high request rate",
		},
		{
			name: "low rate limit",
			rateLimit: config.RateLimitConfig{
				RequestsPerSecond: 0.1,
				BurstSize:         1,
			},
			wantErr:     false,
			description: "allows low request rate",
		},
		{
			name: "zero burst size",
			rateLimit: config.RateLimitConfig{
				RequestsPerSecond: 10.0,
				BurstSize:         0,
			},
			wantErr:     true,
			description: "zero burst should be invalid",
		},
		{
			name: "negative rate",
			rateLimit: config.RateLimitConfig{
				RequestsPerSecond: -1.0,
				BurstSize:         10,
			},
			wantErr:     true,
			description: "negative rate should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: tt.rateLimit,
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientRetryConfiguration tests retry configuration
func TestClientRetryConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		retryAttempts int
		retryDelay    time.Duration
		wantErr       bool
	}{
		{
			name:          "no retries",
			retryAttempts: 0,
			retryDelay:    1 * time.Second,
			wantErr:       false,
		},
		{
			name:          "some retries",
			retryAttempts: 5,
			retryDelay:    1 * time.Second,
			wantErr:       false,
		},
		{
			name:          "many retries",
			retryAttempts: 10,
			retryDelay:    1 * time.Second,
			wantErr:       false,
		},
		{
			name:          "negative retry attempts",
			retryAttempts: -1,
			retryDelay:    1 * time.Second,
			wantErr:       true,
		},
		{
			name:          "zero retry delay",
			retryAttempts: 3,
			retryDelay:    0 * time.Second,
			wantErr:       true,
		},
		{
			name:          "negative retry delay",
			retryAttempts: 3,
			retryDelay:    -1 * time.Second,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: tt.retryAttempts,
				RetryDelay:    tt.retryDelay,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientTimeoutConfiguration tests timeout configuration
func TestClientTimeoutConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "very short timeout",
			timeout: 10 * time.Millisecond,
			wantErr: false,
		},
		{
			name:    "normal timeout",
			timeout: 30 * time.Second,
			wantErr: false,
		},
		{
			name:    "long timeout",
			timeout: 5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "zero timeout",
			timeout: 0 * time.Second,
			wantErr: true,
		},
		{
			name:    "negative timeout",
			timeout: -1 * time.Second,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       tt.timeout,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestConnectionPoolEdgeCases tests connection pool edge cases
func TestConnectionPoolEdgeCases(t *testing.T) {
	cfg := &config.SLURMConfig{
		BaseURL:       "https://example.com:6820",
		APIVersion:    "v0.0.42",
		Timeout:       100 * time.Millisecond,
		RetryAttempts: 0,
		RetryDelay:    1 * time.Second,
		Auth: config.AuthConfig{
			Type: "none",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
		},
	}

	t.Run("very large pool size", func(t *testing.T) {
		pool, err := NewConnectionPool(cfg, 100)
		if err != nil {
			t.Fatalf("NewConnectionPool() error = %v", err)
		}
		if len(pool.clients) != 100 {
			t.Errorf("Pool size = %d, want 100", len(pool.clients))
		}

		for i := 0; i < 110; i++ {
			client := pool.GetClient()
			if client == nil {
				t.Errorf("GetClient() returned nil at iteration %d", i)
			}
		}
	})

	t.Run("pool of size one", func(t *testing.T) {
		pool, err := NewConnectionPool(cfg, 1)
		if err != nil {
			t.Fatalf("NewConnectionPool() error = %v", err)
		}
		if len(pool.clients) != 1 {
			t.Errorf("Pool size = %d, want 1", len(pool.clients))
		}

		client1 := pool.GetClient()
		client2 := pool.GetClient()
		if client1 != client2 {
			t.Error("Pool with size 1 should always return the same client")
		}
	})
}

// TestClientAPIVersionConfiguration tests API version configuration
func TestClientAPIVersionConfiguration(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		wantErr    bool
	}{
		{
			name:       "v0.0.42",
			apiVersion: "v0.0.42",
			wantErr:    false,
		},
		{
			name:       "v0.0.43",
			apiVersion: "v0.0.43",
			wantErr:    false,
		},
		{
			name:       "empty version",
			apiVersion: "",
			wantErr:    false,
		},
		{
			name:       "invalid version format",
			apiVersion: "invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    tt.apiVersion,
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientAuthConfiguration tests authentication configuration
func TestClientAuthConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		auth    config.AuthConfig
		wantErr bool
	}{
		{
			name: "none auth",
			auth: config.AuthConfig{
				Type: "none",
			},
			wantErr: false,
		},
		{
			name: "jwt with token",
			auth: config.AuthConfig{
				Type:  "jwt",
				Token: "test-token",
			},
			wantErr: false,
		},
		{
			name: "basic auth with credentials",
			auth: config.AuthConfig{
				Type:     "basic",
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "jwt without token",
			auth: config.AuthConfig{
				Type: "jwt",
			},
			wantErr: true,
		},
		{
			name: "unsupported auth type",
			auth: config.AuthConfig{
				Type: "unsupported",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       "https://example.com:6820",
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth:          tt.auth,
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientBaseURLConfiguration tests base URL configuration
func TestClientBaseURLConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid http URL",
			baseURL: "http://example.com:6820",
			wantErr: false,
		},
		{
			name:    "valid https URL",
			baseURL: "https://example.com:6820",
			wantErr: false,
		},
		{
			name:    "URL with path",
			baseURL: "https://example.com:6820/slurm",
			wantErr: false,
		},
		{
			name:    "empty base URL",
			baseURL: "",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			baseURL: "not-a-url",
			wantErr: true,
		},
		{
			name:    "URL without scheme",
			baseURL: "example.com:6820",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.SLURMConfig{
				BaseURL:       tt.baseURL,
				APIVersion:    "v0.0.42",
				Timeout:       100 * time.Millisecond,
				RetryAttempts: 0,
				RetryDelay:    1 * time.Second,
				Auth: config.AuthConfig{
					Type: "none",
				},
				RateLimit: config.RateLimitConfig{
					RequestsPerSecond: 10.0,
					BurstSize:         20,
				},
			}

			client, err := NewClient(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}

			if client != nil {
				_ = client.Close()
			}
		})
	}
}

// TestClientTestConnection tests the TestConnection method
func TestClientTestConnection(t *testing.T) {
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a proper SLURM API ping response structure
		w.Write([]byte(`{"meta":{"plugin":{"type":"accounting_storage","name":"slurm_mysql"}}}`))
	}))
	defer successServer.Close()

	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failServer.Close()

	t.Run("successful connection test", func(t *testing.T) {
		cfg := &config.SLURMConfig{
			BaseURL:       successServer.URL,
			APIVersion:    "v0.0.42",
			Timeout:       100 * time.Millisecond,
			RetryAttempts: 0,
			RetryDelay:    1 * time.Second,
			Auth: config.AuthConfig{
				Type: "none",
			},
			RateLimit: config.RateLimitConfig{
				RequestsPerSecond: 1000.0,
				BurstSize:         100,
			},
		}

		client, err := NewClient(cfg)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer func() { _ = client.Close() }()

		ctx := context.Background()
		err = client.TestConnection(ctx)
		if err != nil {
			t.Errorf("TestConnection() error = %v", err)
		}
	})

	t.Run("failed connection test", func(t *testing.T) {
		cfg := &config.SLURMConfig{
			BaseURL:       failServer.URL,
			APIVersion:    "v0.0.42",
			Timeout:       100 * time.Millisecond,
			RetryAttempts: 0,
			RetryDelay:    1 * time.Second,
			Auth: config.AuthConfig{
				Type: "none",
			},
			RateLimit: config.RateLimitConfig{
				RequestsPerSecond: 1000.0,
				BurstSize:         100,
			},
		}

		client, err := NewClient(cfg)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer func() { _ = client.Close() }()

		ctx := context.Background()
		err = client.TestConnection(ctx)
		if err == nil {
			t.Error("TestConnection() expected error, got nil")
		}
	})

	t.Run("test connection with cancelled context", func(t *testing.T) {
		cfg := &config.SLURMConfig{
			BaseURL:       "https://example.com:6820",
			APIVersion:    "v0.0.42",
			Timeout:       100 * time.Millisecond,
			RetryAttempts: 0,
			RetryDelay:    1 * time.Second,
			Auth: config.AuthConfig{
				Type: "none",
			},
			RateLimit: config.RateLimitConfig{
				RequestsPerSecond: 1000.0,
				BurstSize:         100,
			},
		}

		client, err := NewClient(cfg)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		defer func() { _ = client.Close() }()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = client.TestConnection(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})
}
