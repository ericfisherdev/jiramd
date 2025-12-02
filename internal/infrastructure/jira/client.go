// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"context"

	"github.com/esfisher/jiramd/internal/domain"
)

// Client represents a Jira API client.
// It implements communication with Jira Cloud REST API.
//
// TODO: Inject http.Client (or interface) and logger via NewClient for better testability
// and control over timeouts/retries. Map HTTP status codes to domain errors (404 -> ErrNotFound,
// 401/403 -> ErrUnauthorized).
type Client struct {
	baseURL string
	email   string
	token   string
}

// NewClient creates a new Jira API client.
func NewClient(baseURL, email, token string) *Client {
	return &Client{
		baseURL: baseURL,
		email:   email,
		token:   token,
	}
}

// GetTicket retrieves a ticket from Jira.
// This is a placeholder for the actual implementation.
func (c *Client) GetTicket(ctx context.Context, key string) (*domain.Ticket, error) {
	// TODO: Implement Jira API call to get ticket
	return nil, nil
}

// UpdateTicket updates a ticket in Jira.
// This is a placeholder for the actual implementation.
func (c *Client) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	// TODO: Implement Jira API call to update ticket
	return nil
}

// GetProject retrieves a project from Jira.
// This is a placeholder for the actual implementation.
func (c *Client) GetProject(ctx context.Context, key string) (*domain.Project, error) {
	// TODO: Implement Jira API call to get project
	return nil, nil
}
