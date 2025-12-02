package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
)

func TestLoader_Load_Success(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
jira:
  base_url: "https://example.atlassian.net"
  email: "test@example.com"
  token: "test-token"
  project: "TEST"

sync:
  interval: 5m
  markdown_dir: "/tmp/tickets"
  watch_enabled: true

storage:
  db_path: "/tmp/jiramd.db"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	loader := NewLoader()
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify configuration values
	if cfg.Jira.BaseURL != "https://example.atlassian.net" {
		t.Errorf("Jira.BaseURL = %v, want %v", cfg.Jira.BaseURL, "https://example.atlassian.net")
	}

	if cfg.Jira.Email != "test@example.com" {
		t.Errorf("Jira.Email = %v, want %v", cfg.Jira.Email, "test@example.com")
	}

	if cfg.Jira.Token != "test-token" {
		t.Errorf("Jira.Token = %v, want %v", cfg.Jira.Token, "test-token")
	}

	if cfg.Jira.Project != "TEST" {
		t.Errorf("Jira.Project = %v, want %v", cfg.Jira.Project, "TEST")
	}

	if cfg.Sync.Interval != 5*time.Minute {
		t.Errorf("Sync.Interval = %v, want %v", cfg.Sync.Interval, 5*time.Minute)
	}

	if cfg.Sync.MarkdownDir != "/tmp/tickets" {
		t.Errorf("Sync.MarkdownDir = %v, want %v", cfg.Sync.MarkdownDir, "/tmp/tickets")
	}

	if !cfg.Sync.WatchEnabled {
		t.Errorf("Sync.WatchEnabled = %v, want %v", cfg.Sync.WatchEnabled, true)
	}

	if cfg.Storage.DBPath != "/tmp/jiramd.db" {
		t.Errorf("Storage.DBPath = %v, want %v", cfg.Storage.DBPath, "/tmp/jiramd.db")
	}
}

func TestLoader_Load_EnvVarExpansion(t *testing.T) {
	// Set test environment variables
	os.Setenv("JIRAMD_API_TOKEN", "secret-token-from-env")
	os.Setenv("JIRAMD_BASE_URL", "https://env.atlassian.net")
	defer func() {
		os.Unsetenv("JIRAMD_API_TOKEN")
		os.Unsetenv("JIRAMD_BASE_URL")
	}()

	// Create a temporary config file with env var references
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
jira:
  base_url: "${JIRAMD_BASE_URL}"
  email: "test@example.com"
  token: "${JIRAMD_API_TOKEN}"
  project: "TEST"

sync:
  interval: 5m
  markdown_dir: "/tmp/tickets"
  watch_enabled: true

storage:
  db_path: "/tmp/jiramd.db"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	loader := NewLoader()
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify environment variables were expanded
	if cfg.Jira.BaseURL != "https://env.atlassian.net" {
		t.Errorf("Jira.BaseURL = %v, want %v", cfg.Jira.BaseURL, "https://env.atlassian.net")
	}

	if cfg.Jira.Token != "secret-token-from-env" {
		t.Errorf("Jira.Token = %v, want %v", cfg.Jira.Token, "secret-token-from-env")
	}
}

func TestLoader_Load_HomeDirExpansion(t *testing.T) {
	// Create a temporary config file with home directory references
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
jira:
  base_url: "https://example.atlassian.net"
  email: "test@example.com"
  token: "test-token"
  project: "TEST"

sync:
  interval: 5m
  markdown_dir: "~/jira-tickets"
  watch_enabled: true

storage:
  db_path: "~/.local/share/jiramd/jiramd.db"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	loader := NewLoader()
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Get expected home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	expectedMarkdownDir := filepath.Join(homeDir, "jira-tickets")
	expectedDBPath := filepath.Join(homeDir, ".local/share/jiramd/jiramd.db")

	// Verify home directory was expanded
	if cfg.Sync.MarkdownDir != expectedMarkdownDir {
		t.Errorf("Sync.MarkdownDir = %v, want %v", cfg.Sync.MarkdownDir, expectedMarkdownDir)
	}

	if cfg.Storage.DBPath != expectedDBPath {
		t.Errorf("Storage.DBPath = %v, want %v", cfg.Storage.DBPath, expectedDBPath)
	}
}

func TestLoader_Load_FileNotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}

	// Verify it's a ConfigError
	var configErr *domain.ConfigError
	if !isConfigError(err) {
		t.Errorf("Load() error type = %T, want *domain.ConfigError", err)
	}
	_ = configErr // Avoid unused variable error
}

func TestLoader_Load_InvalidYAML(t *testing.T) {
	// Create a temporary config file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
jira:
  base_url: "https://example.atlassian.net"
  email: test@example.com"  # Missing opening quote
  invalid yaml structure
`

	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	loader := NewLoader()
	_, err := loader.Load(configPath)
	if err == nil {
		t.Error("Load() expected error for invalid YAML, got nil")
	}

	// Verify it's a ConfigError
	if !isConfigError(err) {
		t.Errorf("Load() error type = %T, want *domain.ConfigError", err)
	}
}

func TestLoader_Load_InvalidDuration(t *testing.T) {
	// Create a temporary config file with invalid duration
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
jira:
  base_url: "https://example.atlassian.net"
  email: "test@example.com"
  token: "test-token"
  project: "TEST"

sync:
  interval: "invalid-duration"
  markdown_dir: "/tmp/tickets"
  watch_enabled: true

storage:
  db_path: "/tmp/jiramd.db"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load configuration
	loader := NewLoader()
	_, err := loader.Load(configPath)
	if err == nil {
		t.Error("Load() expected error for invalid duration, got nil")
	}

	// Verify it's a ConfigError
	if !isConfigError(err) {
		t.Errorf("Load() error type = %T, want *domain.ConfigError", err)
	}
}

// Helper function to check if error is a ConfigError
func isConfigError(err error) bool {
	_, ok := err.(*domain.ConfigError)
	return ok
}
