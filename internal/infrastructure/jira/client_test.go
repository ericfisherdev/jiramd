package jira

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name    string
		baseURL string
		email   string
		token   string
		logger  *slog.Logger
		wantErr bool
	}{
		{
			name:    "valid client creation",
			baseURL: "https://test.atlassian.net",
			email:   "test@example.com",
			token:   "test-token-123",
			logger:  logger,
			wantErr: false,
		},
		{
			name:    "empty baseURL",
			baseURL: "",
			email:   "test@example.com",
			token:   "test-token-123",
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "empty email",
			baseURL: "https://test.atlassian.net",
			email:   "",
			token:   "test-token-123",
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "empty token",
			baseURL: "https://test.atlassian.net",
			email:   "test@example.com",
			token:   "",
			logger:  logger,
			wantErr: true,
		},
		{
			name:    "nil logger",
			baseURL: "https://test.atlassian.net",
			email:   "test@example.com",
			token:   "test-token-123",
			logger:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.email, tt.token, tt.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client when error was not expected")
			}
			if !tt.wantErr && client.jiraClient == nil {
				t.Error("NewClient() returned client with nil jiraClient")
			}
			if !tt.wantErr && client.logger == nil {
				t.Error("NewClient() returned client with nil logger")
			}
		})
	}
}
