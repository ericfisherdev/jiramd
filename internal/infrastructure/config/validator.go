// Package config provides infrastructure implementation for configuration loading.
package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/esfisher/jiramd/internal/domain"
)

// Validator implements domain.ConfigValidator interface.
type Validator struct{}

// NewValidator creates a new configuration validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates the configuration according to business rules.
// Returns domain error if validation fails.
func (v *Validator) Validate(config *domain.Config) error {
	if err := v.validateJira(&config.Jira); err != nil {
		return err
	}

	if err := v.validateSync(&config.Sync); err != nil {
		return err
	}

	if err := v.validateStorage(&config.Storage); err != nil {
		return err
	}

	return nil
}

// validateJira validates Jira configuration fields.
func (v *Validator) validateJira(jira *domain.JiraConfig) error {
	// Validate BaseURL is present and valid
	if jira.BaseURL == "" {
		return domain.NewConfigError("jira.base_url is required")
	}

	// Validate BaseURL format
	if _, err := url.Parse(jira.BaseURL); err != nil {
		return domain.NewConfigError(fmt.Sprintf("jira.base_url is not a valid URL: %v", err))
	}

	// Ensure BaseURL uses HTTPS
	if !strings.HasPrefix(jira.BaseURL, "https://") {
		return domain.NewConfigError("jira.base_url must use https:// protocol for security")
	}

	// Validate Email is present
	if jira.Email == "" {
		return domain.NewConfigError("jira.email is required")
	}

	// Basic email format validation
	if !strings.Contains(jira.Email, "@") {
		return domain.NewConfigError("jira.email must be a valid email address")
	}

	// Validate Token is present (critical for security)
	if jira.Token == "" {
		return domain.NewConfigError("jira.token is required (set JIRAMD_API_TOKEN environment variable)")
	}

	// Validate Project is present
	if jira.Project == "" {
		return domain.NewConfigError("jira.project is required")
	}

	// Validate Project key format (Jira project keys are typically uppercase letters)
	if len(jira.Project) < 2 || len(jira.Project) > 10 {
		return domain.NewConfigError("jira.project must be between 2 and 10 characters")
	}

	return nil
}

// validateSync validates Sync configuration fields.
func (v *Validator) validateSync(sync *domain.SyncConfig) error {
	// Validate Interval is positive
	if sync.Interval <= 0 {
		return domain.NewConfigError("sync.interval must be positive")
	}

	// Validate MarkdownDir is present
	if sync.MarkdownDir == "" {
		return domain.NewConfigError("sync.markdown_dir is required")
	}

	return nil
}

// validateStorage validates Storage configuration fields.
func (v *Validator) validateStorage(storage *domain.StorageConfig) error {
	// Validate DBPath is present
	if storage.DBPath == "" {
		return domain.NewConfigError("storage.db_path is required")
	}

	return nil
}
