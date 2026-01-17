package slurm

import (
	"context"
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
				Timeout:       2 * time.Second, // Short timeout for tests
				RetryAttempts: 0,               // No retries to avoid long waits
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
				Timeout:       2 * time.Second,
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
		Timeout:       2 * time.Second, // Short timeout for tests
		RetryAttempts: 0,               // No retries to avoid long waits
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
		Timeout:       2 * time.Second, // Short timeout for tests
		RetryAttempts: 0,               // No retries to avoid long waits
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
		Timeout:       2 * time.Second, // Short timeout for tests
		RetryAttempts: 0,               // No retries to avoid long waits
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
		Timeout:       2 * time.Second, // Short timeout for tests
		RetryAttempts: 0,               // No retries to avoid long waits
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
		Timeout:       2 * time.Second, // Short timeout for tests
		RetryAttempts: 0,               // No retries for faster test
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
