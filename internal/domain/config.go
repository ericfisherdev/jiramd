// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"time"
)

// Config represents the application configuration value object.
// This is a value object containing immutable configuration data.
type Config struct {
	Jira    JiraConfig
	Sync    SyncConfig
	Storage StorageConfig
}

// JiraConfig contains Jira-specific configuration.
type JiraConfig struct {
	BaseURL string
	Email   string
	Token   string
	Project string
}

// SyncConfig contains synchronization-specific configuration.
type SyncConfig struct {
	Interval     time.Duration
	MarkdownDir  string
	WatchEnabled bool
}

// StorageConfig contains storage-specific configuration.
type StorageConfig struct {
	DBPath string
}

// ConfigLoader defines the interface for loading configuration.
// This interface allows infrastructure implementations while keeping domain pure.
type ConfigLoader interface {
	// Load loads configuration from the specified path.
	// Returns domain error if loading fails.
	Load(path string) (*Config, error)
}

// ConfigValidator defines the interface for validating configuration.
type ConfigValidator interface {
	// Validate validates the configuration.
	// Returns domain error if validation fails.
	Validate(config *Config) error
}
