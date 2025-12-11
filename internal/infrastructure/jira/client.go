// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"fmt"
	"log/slog"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
)

// Client represents a Jira API client.
// It wraps the go-jira/v2 client with authentication and provides
// infrastructure-layer access to Jira Cloud REST API.
type Client struct {
	jiraClient *jira.Client
	logger     *slog.Logger
}

// NewClient creates a new Jira API client with BasicAuth transport.
// The baseURL should be the full Jira Cloud URL (e.g., "https://yoursite.atlassian.net").
// The email is the user's email address and token is the API token from Jira Cloud.
// Logger is injected for dependency inversion and better testability.
func NewClient(baseURL, email, token string, logger *slog.Logger) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Create transport with Basic Auth (email + API token)
	tp := jira.BasicAuthTransport{
		Username: email,
		APIToken: token,
	}

	// Create the Jira client with the authenticated transport
	jiraClient, err := jira.NewClient(baseURL, tp.Client())
	if err != nil {
		return nil, fmt.Errorf("failed to create jira client: %w", err)
	}

	return &Client{
		jiraClient: jiraClient,
		logger:     logger,
	}, nil
}
