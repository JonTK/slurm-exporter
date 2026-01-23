// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package slurm

import (
	"context"
	"fmt"
	"sync"
	"time"

	slurm "github.com/jontk/slurm-client"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/jontk/slurm-exporter/internal/config"
	authpkg "github.com/jontk/slurm-exporter/internal/slurm/auth"
)

// Client provides a wrapper around the SLURM client with additional functionality
type Client struct {
	client      slurm.SlurmClient
	config      *config.SLURMConfig
	rateLimiter *rate.Limiter
	mu          sync.RWMutex
	connected   bool
	lastError   error
	retryCount  int
}

// ConnectionPool manages a pool of SLURM client connections
type ConnectionPool struct {
	clients []*Client
	current int
	mu      sync.Mutex
	config  *config.SLURMConfig
}

// NewClient creates a new SLURM client wrapper
func NewClient(cfg *config.SLURMConfig) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid SLURM configuration: %w", err)
	}

	// Create rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond), cfg.RateLimit.BurstSize)

	// Configure authentication using the auth package
	authProvider, err := authpkg.ConfigureAuth(&cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to configure authentication: %w", err)
	}

	// Create the SLURM client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var client slurm.SlurmClient

	if cfg.UseAdapters {
		// Use adapter pattern for better version compatibility
		logrus.Info("Creating SLURM client with adapter pattern enabled")

		// When using adapters, we should use the factory pattern from slurm-client
		// For now, we'll use the standard approach but with auto-detection if no version specified
		if cfg.APIVersion == "" {
			client, err = slurm.NewClient(ctx,
				slurm.WithBaseURL(cfg.BaseURL),
				slurm.WithAuth(authProvider),
			)
		} else {
			client, err = slurm.NewClientWithVersion(ctx, cfg.APIVersion,
				slurm.WithBaseURL(cfg.BaseURL),
				slurm.WithAuth(authProvider),
			)
		}
	} else {
		// Traditional client creation
		client, err = slurm.NewClientWithVersion(ctx, cfg.APIVersion,
			slurm.WithBaseURL(cfg.BaseURL),
			slurm.WithAuth(authProvider),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create SLURM client: %w", err)
	}

	wrapper := &Client{
		client:      client,
		config:      cfg,
		rateLimiter: rateLimiter,
		connected:   false,
	}

	// Test the connection
	if err := wrapper.TestConnection(context.Background()); err != nil {
		logrus.WithError(err).Warn("Initial SLURM connection test failed")
		wrapper.lastError = err
	} else {
		wrapper.connected = true
		logrus.Info("SLURM client connection established successfully")
	}

	return wrapper, nil
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(cfg *config.SLURMConfig, poolSize int) (*ConnectionPool, error) {
	if poolSize <= 0 {
		poolSize = 5 // Default pool size
	}

	pool := &ConnectionPool{
		clients: make([]*Client, poolSize),
		config:  cfg,
	}

	// Create pool clients
	for i := 0; i < poolSize; i++ {
		client, err := NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create client %d: %w", i, err)
		}
		pool.clients[i] = client
	}

	return pool, nil
}

// GetClient returns the next available client from the pool
func (p *ConnectionPool) GetClient() *Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	client := p.clients[p.current]
	p.current = (p.current + 1) % len(p.clients)
	return client
}

// TestConnection tests the SLURM API connection
func (c *Client) TestConnection(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}

	// Create a context with timeout
	reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Test connection by pinging the cluster
	err := c.client.Info().Ping(reqCtx)
	if err != nil {
		c.connected = false
		c.lastError = err
		return fmt.Errorf("SLURM API connection test failed: %w", err)
	}

	c.connected = true
	c.lastError = nil
	c.retryCount = 0
	return nil
}

// IsConnected returns the current connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetLastError returns the last connection error
func (c *Client) GetLastError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastError
}

// executeWithRetry executes a function with retry logic
func (c *Client) executeWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Execute the operation
		if err := operation(); err != nil {
			lastErr = err

			// Don't retry on context cancellation
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Log the attempt
			logrus.WithError(err).WithField("attempt", attempt+1).Debug("SLURM API request failed")

			// If this isn't the last attempt, wait before retrying
			if attempt < c.config.RetryAttempts {
				delay := c.config.RetryDelay * time.Duration(attempt+1)
				logrus.WithField("delay", delay).Debug("Retrying SLURM API request")

				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			// Success
			c.mu.Lock()
			c.connected = true
			c.lastError = nil
			c.retryCount = 0
			c.mu.Unlock()
			return nil
		}
	}

	// All retries failed
	c.mu.Lock()
	c.connected = false
	c.lastError = lastErr
	c.retryCount++
	c.mu.Unlock()

	return fmt.Errorf("operation failed after %d attempts: %w", c.config.RetryAttempts+1, lastErr)
}

// GetJobs retrieves job information from SLURM
func (c *Client) GetJobs(ctx context.Context) (*slurm.JobList, error) {
	var result *slurm.JobList

	err := c.executeWithRetry(ctx, func() error {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		var err error
		jobList, err := c.client.Jobs().List(reqCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to get jobs: %w", err)
		}
		result = jobList

		return nil
	})

	return result, err
}

// GetNodes retrieves node information from SLURM
func (c *Client) GetNodes(ctx context.Context) (*slurm.NodeList, error) {
	var result *slurm.NodeList

	err := c.executeWithRetry(ctx, func() error {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		var err error
		nodeList, err := c.client.Nodes().List(reqCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to get nodes: %w", err)
		}
		result = nodeList

		return nil
	})

	return result, err
}

// GetPartitions retrieves partition information from SLURM
func (c *Client) GetPartitions(ctx context.Context) (*slurm.PartitionList, error) {
	var result *slurm.PartitionList

	err := c.executeWithRetry(ctx, func() error {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		var err error
		partitionList, err := c.client.Partitions().List(reqCtx, nil)
		if err != nil {
			return fmt.Errorf("failed to get partitions: %w", err)
		}
		result = partitionList

		return nil
	})

	return result, err
}

// GetInfo retrieves cluster information from SLURM
func (c *Client) GetInfo(ctx context.Context) (*slurm.ClusterInfo, error) {
	var result *slurm.ClusterInfo

	err := c.executeWithRetry(ctx, func() error {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		var err error
		clusterInfo, err := c.client.Info().Get(reqCtx)
		if err != nil {
			return fmt.Errorf("failed to get cluster info: %w", err)
		}
		result = clusterInfo

		return nil
	})

	return result, err
}

// GetStats retrieves cluster statistics from SLURM
func (c *Client) GetStats(ctx context.Context) (*slurm.ClusterStats, error) {
	var result *slurm.ClusterStats

	err := c.executeWithRetry(ctx, func() error {
		reqCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
		defer cancel()

		var err error
		clusterStats, err := c.client.Info().Stats(reqCtx)
		if err != nil {
			return fmt.Errorf("failed to get cluster stats: %w", err)
		}
		result = clusterStats

		return nil
	})

	return result, err
}

// Close cleans up the client resources
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	// No explicit cleanup needed for the HTTP client in this case
	return nil
}

// GetRetryCount returns the current retry count
func (c *Client) GetRetryCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.retryCount
}

// ResetRetryCount resets the retry counter
func (c *Client) ResetRetryCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.retryCount = 0
}

// GetSlurmClient returns the underlying SLURM client
func (c *Client) GetSlurmClient() slurm.SlurmClient {
	return c.client
}
