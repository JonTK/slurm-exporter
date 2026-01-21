package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jontk/slurm-exporter/internal/config"
)

func TestConfigureAuth(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.AuthConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "none auth",
			config: &config.AuthConfig{
				Type: "none",
			},
			wantErr: false,
		},
		{
			name: "jwt with token",
			config: &config.AuthConfig{
				Type:  "jwt",
				Token: "test-jwt-token",
			},
			wantErr: false,
		},
		{
			name: "jwt without token",
			config: &config.AuthConfig{
				Type: "jwt",
			},
			wantErr: true,
			errMsg:  "JWT auth requires token or token_file",
		},
		{
			name: "basic auth with credentials",
			config: &config.AuthConfig{
				Type:     "basic",
				Username: "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "basic auth without username",
			config: &config.AuthConfig{
				Type:     "basic",
				Password: "testpass",
			},
			wantErr: true,
			errMsg:  "basic auth requires username",
		},
		{
			name: "basic auth without password",
			config: &config.AuthConfig{
				Type:     "basic",
				Username: "testuser",
			},
			wantErr: true,
			errMsg:  "basic auth requires password",
		},
		{
			name: "apikey with key",
			config: &config.AuthConfig{
				Type:   "apikey",
				APIKey: "test-api-key",
			},
			wantErr: false,
		},
		{
			name: "unsupported auth type",
			config: &config.AuthConfig{
				Type: "oauth2",
			},
			wantErr: true,
			errMsg:  "unsupported auth type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := ConfigureAuth(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigureAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ConfigureAuth() error message = %v, want containing %v", err.Error(), tt.errMsg)
				}
			}
			if !tt.wantErr && provider == nil {
				t.Error("ConfigureAuth() returned nil provider without error")
			}
		})
	}
}

func TestReadSecretFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "auth-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name        string
		filename    string
		content     string
		permissions os.FileMode
		wantErr     bool
		wantValue   string
	}{
		{
			name:        "valid secret file",
			filename:    "token.txt",
			content:     "secret-token-123",
			permissions: 0600,
			wantErr:     false,
			wantValue:   "secret-token-123",
		},
		{
			name:        "secret with newlines",
			filename:    "token-newline.txt",
			content:     "secret-token-456\n\n",
			permissions: 0600,
			wantErr:     false,
			wantValue:   "secret-token-456",
		},
		{
			name:        "empty file",
			filename:    "empty.txt",
			content:     "",
			permissions: 0600,
			wantErr:     true,
		},
		{
			name:        "file with spaces only",
			filename:    "spaces.txt",
			content:     "   \n\t\n   ",
			permissions: 0600,
			wantErr:     true,
		},
		{
			name:        "overly permissive file",
			filename:    "permissive.txt",
			content:     "secret-token-789",
			permissions: 0644,
			wantErr:     false,
			wantValue:   "secret-token-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filepath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filepath, []byte(tt.content), tt.permissions)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test readSecretFile
			value, err := readSecretFile(filepath, "test secret")
			if (err != nil) != tt.wantErr {
				t.Errorf("readSecretFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && value != tt.wantValue {
				t.Errorf("readSecretFile() = %v, want %v", value, tt.wantValue)
			}
		})
	}

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		_, err := readSecretFile("/non/existent/file", "test")
		if err == nil {
			t.Error("readSecretFile() expected error for non-existent file")
		}
	})
}

func TestConfigureAuthWithFiles(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "auth-file-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files
	tokenFile := filepath.Join(tmpDir, "jwt-token")
	passwordFile := filepath.Join(tmpDir, "password")
	apiKeyFile := filepath.Join(tmpDir, "api-key")

	_ = os.WriteFile(tokenFile, []byte("jwt-token-from-file"), 0600)
	_ = os.WriteFile(passwordFile, []byte("password-from-file"), 0600)
	_ = os.WriteFile(apiKeyFile, []byte("api-key-from-file"), 0600)

	tests := []struct {
		name    string
		config  *config.AuthConfig
		wantErr bool
	}{
		{
			name: "jwt with token file",
			config: &config.AuthConfig{
				Type:      "jwt",
				TokenFile: tokenFile,
			},
			wantErr: false,
		},
		{
			name: "basic auth with password file",
			config: &config.AuthConfig{
				Type:         "basic",
				Username:     "testuser",
				PasswordFile: passwordFile,
			},
			wantErr: false,
		},
		{
			name: "apikey with key file",
			config: &config.AuthConfig{
				Type:       "apikey",
				APIKeyFile: apiKeyFile,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := ConfigureAuth(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigureAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("ConfigureAuth() returned nil provider without error")
			}
		})
	}
}

func TestServiceAccountAuth(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "sa-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test token file
	tokenPath := filepath.Join(tmpDir, "token")
	tokenContent := "k8s-service-account-token"
	err = os.WriteFile(tokenPath, []byte(tokenContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create token file: %v", err)
	}

	t.Run("custom token path", func(t *testing.T) {
		sa := NewServiceAccountAuth(tokenPath)
		token, err := sa.GetToken()
		if err != nil {
			t.Errorf("GetToken() error = %v", err)
			return
		}
		if token != tokenContent {
			t.Errorf("GetToken() = %v, want %v", token, tokenContent)
		}

		provider, err := sa.ToAuthProvider()
		if err != nil {
			t.Errorf("ToAuthProvider() error = %v", err)
			return
		}
		if provider == nil {
			t.Error("ToAuthProvider() returned nil provider")
		}
	})

	t.Run("default token path", func(t *testing.T) {
		sa := NewServiceAccountAuth("")
		// This will fail as the default path doesn't exist in test environment
		_, err := sa.GetToken()
		if err == nil {
			t.Error("GetToken() expected error for non-existent default path")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s[0:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(substr) > 0 && len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
