// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 SLURM Exporter Contributors

package auth

import (
	"fmt"
	"os"
	"strings"

	slurmauth "github.com/jontk/slurm-client/pkg/auth"
	"github.com/jontk/slurm-exporter/internal/config"
	"github.com/sirupsen/logrus"
)

// ConfigureAuth creates an auth provider based on the configuration
func ConfigureAuth(cfg *config.AuthConfig) (slurmauth.Provider, error) {
	switch cfg.Type {
	case "none":
		logrus.Debug("Using no authentication")
		return slurmauth.NewNoAuth(), nil

	case "jwt":
		token, err := getJWTToken(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to configure JWT auth: %w", err)
		}
		logrus.Debug("Using JWT authentication")
		return slurmauth.NewTokenAuth(token), nil

	case "basic":
		username, password, err := getBasicCredentials(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to configure basic auth: %w", err)
		}
		logrus.Debug("Using basic authentication")
		return slurmauth.NewBasicAuth(username, password), nil

	case "apikey":
		apiKey, err := getAPIKey(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to configure API key auth: %w", err)
		}
		logrus.Debug("Using API key authentication")
		// Note: The slurm-client may not have a specific API key auth provider
		// In that case, we can use token auth which is often used for API keys
		return slurmauth.NewTokenAuth(apiKey), nil

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", cfg.Type)
	}
}

// getJWTToken retrieves JWT token from configuration or file
func getJWTToken(cfg *config.AuthConfig) (string, error) {
	if cfg.Token != "" {
		return cfg.Token, nil
	}

	if cfg.TokenFile != "" {
		token, err := readSecretFile(cfg.TokenFile, "JWT token")
		if err != nil {
			return "", err
		}
		return token, nil
	}

	return "", fmt.Errorf("JWT auth requires token or token_file to be specified")
}

// getBasicCredentials retrieves username and password for basic auth
func getBasicCredentials(cfg *config.AuthConfig) (string, string, error) {
	if cfg.Username == "" {
		return "", "", fmt.Errorf("basic auth requires username")
	}

	password := cfg.Password
	if password == "" && cfg.PasswordFile != "" {
		var err error
		password, err = readSecretFile(cfg.PasswordFile, "password")
		if err != nil {
			return "", "", err
		}
	}

	if password == "" {
		return "", "", fmt.Errorf("basic auth requires password or password_file")
	}

	return cfg.Username, password, nil
}

// getAPIKey retrieves API key from configuration or file
func getAPIKey(cfg *config.AuthConfig) (string, error) {
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	if cfg.APIKeyFile != "" {
		apiKey, err := readSecretFile(cfg.APIKeyFile, "API key")
		if err != nil {
			return "", err
		}
		return apiKey, nil
	}

	return "", fmt.Errorf("API key auth requires api_key or api_key_file to be specified")
}

// readSecretFile reads a secret from a file with proper security checks
func readSecretFile(filename, secretType string) (string, error) {
	// Check file permissions
	info, err := os.Stat(filename)
	if err != nil {
		return "", fmt.Errorf("failed to stat %s file %s: %w", secretType, filename, err)
	}

	// Warn if file permissions are too open
	mode := info.Mode()
	if mode&0077 != 0 {
		logrus.Warnf("%s file %s has permissions %v, consider restricting to 600", secretType, filename, mode.Perm())
	}

	// Read the file
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read %s from file %s: %w", secretType, filename, err)
	}

	// Trim whitespace and newlines
	secret := strings.TrimSpace(string(content))
	if secret == "" {
		return "", fmt.Errorf("%s file %s is empty", secretType, filename)
	}

	return secret, nil
}

// ServiceAccountAuth provides authentication using Kubernetes service account tokens
type ServiceAccountAuth struct {
	tokenPath string
}

// NewServiceAccountAuth creates a new service account authentication provider
func NewServiceAccountAuth(tokenPath string) *ServiceAccountAuth {
	if tokenPath == "" {
		//nolint:gosec // This is the standard Kubernetes service account token path, not a hardcoded credential
		tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}
	return &ServiceAccountAuth{
		tokenPath: tokenPath,
	}
}

// GetToken reads the service account token from the mounted secret
func (s *ServiceAccountAuth) GetToken() (string, error) {
	token, err := readSecretFile(s.tokenPath, "service account token")
	if err != nil {
		return "", fmt.Errorf("failed to read service account token: %w", err)
	}
	return token, nil
}

// ToAuthProvider converts ServiceAccountAuth to a slurm auth.Provider
func (s *ServiceAccountAuth) ToAuthProvider() (slurmauth.Provider, error) {
	token, err := s.GetToken()
	if err != nil {
		return nil, err
	}
	return slurmauth.NewTokenAuth(token), nil
}

// RefreshableAuth provides an authentication mechanism that can refresh tokens
type RefreshableAuth struct {
	provider    slurmauth.Provider
	refreshFunc func() (slurmauth.Provider, error)
	// TODO: Unused field - preserved for future refresh tracking
	// lastRefresh  int64
	refreshAfter int64 // seconds
}

// NewRefreshableAuth creates a new refreshable authentication provider
func NewRefreshableAuth(initial slurmauth.Provider, refreshFunc func() (slurmauth.Provider, error), refreshAfterSeconds int64) *RefreshableAuth {
	return &RefreshableAuth{
		provider:     initial,
		refreshFunc:  refreshFunc,
		refreshAfter: refreshAfterSeconds,
	}
}

// GetProvider returns the current auth provider, refreshing if necessary
func (r *RefreshableAuth) GetProvider() (slurmauth.Provider, error) {
	// TODO: Implement refresh logic based on time
	// For now, just return the current provider
	return r.provider, nil
}
