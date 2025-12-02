// Package config provides infrastructure implementation for configuration loading.
// This layer handles YAML file parsing and environment variable expansion.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
	"gopkg.in/yaml.v3"
)

// yamlConfig represents the YAML structure for configuration.
// This is separate from domain.Config to allow for YAML-specific handling.
type yamlConfig struct {
	Jira    yamlJiraConfig    `yaml:"jira"`
	Sync    yamlSyncConfig    `yaml:"sync"`
	Storage yamlStorageConfig `yaml:"storage"`
}

type yamlJiraConfig struct {
	BaseURL string `yaml:"base_url"`
	Email   string `yaml:"email"`
	Token   string `yaml:"token"`
	Project string `yaml:"project"`
}

type yamlSyncConfig struct {
	Interval     string `yaml:"interval"`
	MarkdownDir  string `yaml:"markdown_dir"`
	WatchEnabled bool   `yaml:"watch_enabled"`
}

type yamlStorageConfig struct {
	DBPath string `yaml:"db_path"`
}

// Loader implements domain.ConfigLoader interface.
type Loader struct{}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{}
}

// Load loads configuration from the specified YAML file path.
// It performs the following operations:
// 1. Reads and parses YAML file
// 2. Expands environment variables (${VAR} syntax)
// 3. Expands home directory (~)
// 4. Converts YAML structure to domain.Config
// Returns domain error if loading or parsing fails.
func (l *Loader) Load(path string) (*domain.Config, error) {
	// Expand home directory in path
	expandedPath, err := expandHomePath(path)
	if err != nil {
		return nil, domain.NewConfigError(fmt.Sprintf("failed to expand path: %v", err))
	}

	// Read YAML file
	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, domain.NewConfigError(fmt.Sprintf("failed to read config file: %v", err))
	}

	// Parse YAML
	var yamlCfg yamlConfig
	if err := yaml.Unmarshal(data, &yamlCfg); err != nil {
		return nil, domain.NewConfigError(fmt.Sprintf("failed to parse YAML: %v", err))
	}

	// Expand environment variables in all string fields
	if err := expandEnvVars(&yamlCfg); err != nil {
		return nil, domain.NewConfigError(fmt.Sprintf("failed to expand env vars: %v", err))
	}

	// Convert to domain config
	cfg, err := toDomainConfig(&yamlCfg)
	if err != nil {
		return nil, domain.NewConfigError(fmt.Sprintf("failed to convert config: %v", err))
	}

	return cfg, nil
}

// expandHomePath expands ~ to the user's home directory.
func expandHomePath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if path == "~" {
		return homeDir, nil
	}

	// Only expand ~/path, not ~username/path
	if !strings.HasPrefix(path, "~/") {
		return "", fmt.Errorf("unsupported path format: %s (only ~/ is supported)", path)
	}

	return filepath.Join(homeDir, path[2:]), nil
}

// expandEnvVars expands environment variables in the format ${VAR} or $VAR.
// It processes all string fields in the yamlConfig structure.
func expandEnvVars(cfg *yamlConfig) error {
	// Pattern matches ${VAR} or $VAR
	envVarPattern := regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

	// Expand Jira config fields
	cfg.Jira.BaseURL = expandString(cfg.Jira.BaseURL, envVarPattern)
	cfg.Jira.Email = expandString(cfg.Jira.Email, envVarPattern)
	cfg.Jira.Token = expandString(cfg.Jira.Token, envVarPattern)
	cfg.Jira.Project = expandString(cfg.Jira.Project, envVarPattern)

	// Expand Sync config fields
	cfg.Sync.MarkdownDir = expandString(cfg.Sync.MarkdownDir, envVarPattern)

	// Expand Storage config fields
	cfg.Storage.DBPath = expandString(cfg.Storage.DBPath, envVarPattern)

	// Expand home directory paths
	var err error
	cfg.Sync.MarkdownDir, err = expandHomePath(cfg.Sync.MarkdownDir)
	if err != nil {
		return fmt.Errorf("failed to expand markdown_dir: %w", err)
	}

	cfg.Storage.DBPath, err = expandHomePath(cfg.Storage.DBPath)
	if err != nil {
		return fmt.Errorf("failed to expand db_path: %w", err)
	}

	return nil
}

// expandString replaces environment variable references with their values.
func expandString(s string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name (handle both ${VAR} and $VAR)
		varName := strings.TrimPrefix(match, "$")
		varName = strings.TrimPrefix(varName, "{")
		varName = strings.TrimSuffix(varName, "}")

		// Get environment variable value
		value := os.Getenv(varName)
		if value == "" {
			// Keep original if env var not set
			return match
		}
		return value
	})
}

// toDomainConfig converts yamlConfig to domain.Config.
func toDomainConfig(yamlCfg *yamlConfig) (*domain.Config, error) {
	// Parse sync interval
	interval, err := time.ParseDuration(yamlCfg.Sync.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid sync interval '%s': %w", yamlCfg.Sync.Interval, err)
	}

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: yamlCfg.Jira.BaseURL,
			Email:   yamlCfg.Jira.Email,
			Token:   yamlCfg.Jira.Token,
			Project: yamlCfg.Jira.Project,
		},
		Sync: domain.SyncConfig{
			Interval:     interval,
			MarkdownDir:  yamlCfg.Sync.MarkdownDir,
			WatchEnabled: yamlCfg.Sync.WatchEnabled,
		},
		Storage: domain.StorageConfig{
			DBPath: yamlCfg.Storage.DBPath,
		},
	}

	return cfg, nil
}
