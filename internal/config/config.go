// Package config handles application configuration loading and validation.
// This layer depends only on domain and standard library.
package config

import (
	"time"
)

// Config represents the application configuration.
type Config struct {
	// Jira configuration
	Jira JiraConfig `yaml:"jira"`

	// Sync configuration
	Sync SyncConfig `yaml:"sync"`

	// Storage configuration
	Storage StorageConfig `yaml:"storage"`
}

// JiraConfig contains Jira-specific configuration.
type JiraConfig struct {
	// BaseURL is the Jira instance base URL (e.g., "https://example.atlassian.net")
	BaseURL string `yaml:"base_url"`

	// Email is the Jira user email for authentication
	Email string `yaml:"email"`

	// Token is the Jira API token (should be loaded from environment)
	Token string `yaml:"token"`

	// Project is the Jira project key to sync
	Project string `yaml:"project"`
}

// SyncConfig contains synchronization-specific configuration.
type SyncConfig struct {
	// Interval is the sync interval duration
	Interval time.Duration `yaml:"interval"`

	// MarkdownDir is the directory containing markdown files
	MarkdownDir string `yaml:"markdown_dir"`

	// WatchEnabled enables file system watching
	WatchEnabled bool `yaml:"watch_enabled"`
}

// StorageConfig contains storage-specific configuration.
type StorageConfig struct {
	// DBPath is the SQLite database file path
	DBPath string `yaml:"db_path"`
}

// Load loads configuration from a YAML file.
// This is a placeholder for the actual implementation.
func Load(path string) (*Config, error) {
	// TODO: Implement YAML config loading
	return nil, nil
}

// Validate validates the configuration.
// This is a placeholder for the actual implementation.
func (c *Config) Validate() error {
	// TODO: Implement validation logic
	return nil
}
