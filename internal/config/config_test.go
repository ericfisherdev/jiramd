package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_Success(t *testing.T) {
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
	cfg, err := Load(configPath)
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

func TestLoad_ValidationFailure(t *testing.T) {
	// Create a temporary config file with invalid data
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
jira:
  base_url: "http://example.atlassian.net"  # Invalid: must use HTTPS
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
	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() expected validation error for non-HTTPS base_url, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load() expected error for non-existent file, got nil")
	}
}
