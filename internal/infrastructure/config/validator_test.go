package config

import (
	"testing"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
)

func TestValidator_Validate_Success(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err != nil {
		t.Errorf("Validate() error = %v, want nil", err)
	}
}

func TestValidator_Validate_MissingJiraBaseURL(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "", // Missing
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing base_url, got nil")
	}
}

func TestValidator_Validate_InvalidJiraBaseURL(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "not-a-url",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for invalid base_url, got nil")
	}
}

func TestValidator_Validate_NonHTTPSBaseURL(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "http://example.atlassian.net", // HTTP instead of HTTPS
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for non-HTTPS base_url, got nil")
	}
}

func TestValidator_Validate_MissingJiraEmail(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "", // Missing
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing email, got nil")
	}
}

func TestValidator_Validate_InvalidJiraEmail(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "not-an-email", // Invalid email format
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for invalid email, got nil")
	}
}

func TestValidator_Validate_MissingJiraToken(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "", // Missing - CRITICAL SECURITY ISSUE
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing token, got nil")
	}
}

func TestValidator_Validate_MissingJiraProject(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "", // Missing
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing project, got nil")
	}
}

func TestValidator_Validate_InvalidJiraProjectLength(t *testing.T) {
	validator := NewValidator()

	testCases := []struct {
		name    string
		project string
	}{
		{"too short", "A"},
		{"too long", "VERYLONGPROJECTKEY"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &domain.Config{
				Jira: domain.JiraConfig{
					BaseURL: "https://example.atlassian.net",
					Email:   "test@example.com",
					Token:   "test-token",
					Project: tc.project,
				},
				Sync: domain.SyncConfig{
					Interval:     5 * time.Minute,
					MarkdownDir:  "/tmp/tickets",
					WatchEnabled: true,
				},
				Storage: domain.StorageConfig{
					DBPath: "/tmp/jiramd.db",
				},
			}

			err := validator.Validate(cfg)
			if err == nil {
				t.Errorf("Validate() expected error for project '%s', got nil", tc.project)
			}
		})
	}
}

func TestValidator_Validate_NegativeSyncInterval(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     -1 * time.Minute, // Negative
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for negative interval, got nil")
	}
}

func TestValidator_Validate_MissingSyncMarkdownDir(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "", // Missing
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "/tmp/jiramd.db",
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing markdown_dir, got nil")
	}
}

func TestValidator_Validate_MissingStorageDBPath(t *testing.T) {
	validator := NewValidator()

	cfg := &domain.Config{
		Jira: domain.JiraConfig{
			BaseURL: "https://example.atlassian.net",
			Email:   "test@example.com",
			Token:   "test-token",
			Project: "TEST",
		},
		Sync: domain.SyncConfig{
			Interval:     5 * time.Minute,
			MarkdownDir:  "/tmp/tickets",
			WatchEnabled: true,
		},
		Storage: domain.StorageConfig{
			DBPath: "", // Missing
		},
	}

	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Validate() expected error for missing db_path, got nil")
	}
}
