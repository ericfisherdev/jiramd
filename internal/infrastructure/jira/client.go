// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"context"
	"fmt"
	"net/http"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/esfisher/jiramd/internal/domain"
)

// Client represents a Jira API client.
// It implements communication with Jira Cloud REST API.
//
// The client uses the go-jira v2 library for API communication and implements
// proper retry logic with exponential backoff for rate limiting (429 responses).
// HTTP timeouts are set to 30 seconds for all requests.
type Client struct {
	baseURL    string
	email      string
	token      string
	httpClient *http.Client
	jiraClient *jira.Client
}

// NewClient creates a new Jira API client with proper HTTP configuration.
// The HTTP client is configured with:
//   - 30 second timeout for all requests
//   - Exponential backoff retry logic for 429 rate limit responses
//
// Returns an error if the Jira client cannot be initialized.
func NewClient(baseURL, email, token string) (*Client, error) {
	// Create HTTP client with 30 second timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &retryTransport{
			transport:     http.DefaultTransport,
			maxRetries:    3,
			initialDelay:  1 * time.Second,
			maxDelay:      10 * time.Second,
			retryStatuses: map[int]bool{429: true}, // Retry on rate limit
		},
	}

	// Create Jira client using go-jira v2 library
	tp := jira.BasicAuthTransport{
		Username: email,
		APIToken: token,
	}
	tp.Transport = httpClient.Transport

	jiraClient, err := jira.NewClient(baseURL, tp.Client())
	if err != nil {
		return nil, fmt.Errorf("failed to create jira client: %w", err)
	}

	return &Client{
		baseURL:    baseURL,
		email:      email,
		token:      token,
		httpClient: httpClient,
		jiraClient: jiraClient,
	}, nil
}

// retryTransport implements HTTP retry logic with exponential backoff.
// It wraps the default transport and retries requests that fail with specific status codes.
type retryTransport struct {
	transport     http.RoundTripper
	maxRetries    int
	initialDelay  time.Duration
	maxDelay      time.Duration
	retryStatuses map[int]bool
}

// RoundTrip implements http.RoundTripper interface with retry logic.
func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	delay := rt.initialDelay

	for attempt := 0; attempt <= rt.maxRetries; attempt++ {
		// Make the request
		resp, err = rt.transport.RoundTrip(req)

		// If no error and not a retry status, return immediately
		if err == nil && !rt.retryStatuses[resp.StatusCode] {
			return resp, nil
		}

		// If this was the last attempt, return the result
		if attempt == rt.maxRetries {
			return resp, err
		}

		// Check if we should retry based on status code
		if resp != nil && rt.retryStatuses[resp.StatusCode] {
			// Wait before retrying with exponential backoff
			time.Sleep(delay)
			delay *= 2
			if delay > rt.maxDelay {
				delay = rt.maxDelay
			}
			continue
		}

		// For other errors, don't retry
		return resp, err
	}

	return resp, err
}

// GetTicket retrieves a ticket from Jira.
// This is a placeholder for the actual implementation.
func (c *Client) GetTicket(ctx context.Context, key string) (*domain.Ticket, error) {
	// TODO: Implement Jira API call to get ticket
	return nil, fmt.Errorf("jira.Client.GetTicket not implemented")
}

// UpdateTicket pushes local ticket changes to Jira.
// Only updates fields that have changed to minimize API calls.
// Returns the updated ticket with the authoritative Jira timestamp for version tracking.
//
// Domain errors returned:
//   - ErrNotFound: if the ticket no longer exists in Jira (404)
//   - ErrConflict: if the ticket was modified by another user since last fetch (409)
//   - ErrUnauthorized: if the user lacks permission to edit the ticket (401/403)
//   - ErrInvalidInput: if the ticket data is invalid
func (c *Client) UpdateTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Ticket, error) {
	// Validate input
	if err := ticket.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}

	// Build the Jira issue with only fields that should be updated
	jiraIssue := &jira.Issue{
		Key: ticket.Key.String(),
		Fields: &jira.IssueFields{},
	}

	// Only include fields that should be updated
	if err := addFieldUpdates(ticket, jiraIssue.Fields); err != nil {
		return nil, err
	}

	// Make the API call using go-jira v2
	updatedIssue, resp, err := c.jiraClient.Issue.Update(ctx, jiraIssue, nil)
	if err != nil {
		return nil, mapHTTPError(err, resp)
	}

	// Convert the updated issue back to domain entity
	updatedTicket, err := c.issueToDomainTicket(updatedIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert updated issue to domain ticket: %w", err)
	}

	return updatedTicket, nil
}

// addFieldUpdates adds field updates to the IssueFields struct.
// Only includes fields that should be sent to Jira (not read-only fields).
func addFieldUpdates(ticket *domain.Ticket, fields *jira.IssueFields) error {
	// Summary
	if ticket.Summary != "" {
		fields.Summary = ticket.Summary
	}

	// Description
	if ticket.Description != "" {
		fields.Description = ticket.Description
	}

	// Priority - convert to Jira priority object
	if ticket.Priority != "" {
		fields.Priority = &jira.Priority{
			Name: ticket.Priority,
		}
	}

	// Assignee - convert to Jira user object with account ID
	if ticket.Assignee != "" {
		fields.Assignee = &jira.User{
			AccountID: ticket.Assignee,
		}
	}

	// Labels
	if len(ticket.Labels) > 0 {
		fields.Labels = ticket.Labels
	}

	// Note: Status updates require a separate transition API call (not part of update)
	// This will be handled in a separate method if needed

	return nil
}

// mapHTTPError maps HTTP errors to domain errors.
func mapHTTPError(err error, resp *jira.Response) error {
	if err == nil {
		return nil
	}

	// If we don't have a response, return the original error
	if resp == nil || resp.Response == nil {
		return fmt.Errorf("jira api error: %w", err)
	}

	// Map HTTP status codes to domain errors
	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%w: ticket not found in jira", domain.ErrNotFound)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: insufficient permissions", domain.ErrUnauthorized)
	case http.StatusConflict:
		return fmt.Errorf("%w: ticket was modified by another user", domain.ErrConflict)
	case http.StatusBadRequest:
		return fmt.Errorf("%w: invalid ticket data: %v", domain.ErrInvalidInput, err)
	default:
		return fmt.Errorf("jira api error (status %d): %w", resp.StatusCode, err)
	}
}

// issueToDomainTicket converts a Jira issue to a domain Ticket entity.
// This is a placeholder for the actual conversion logic.
func (c *Client) issueToDomainTicket(issue *jira.Issue) (*domain.Ticket, error) {
	// TODO: This will be implemented properly in JMD-13 (Jira Authentication/Fetch)
	// For now, return a basic conversion
	if issue == nil {
		return nil, fmt.Errorf("issue is nil")
	}

	key, err := domain.NewTicketKey(issue.Key)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket key: %w", err)
	}

	// Convert Jira Time to time.Time
	created := time.Time(issue.Fields.Created)
	updated := time.Time(issue.Fields.Updated)

	ticket := domain.NewTicket(
		key,
		issue.Fields.Summary,
		created,
		updated,
	)

	// Set optional fields
	if issue.Fields.Description != "" {
		ticket.Description = issue.Fields.Description
	}

	if issue.Fields.Status != nil {
		ticket.Status = issue.Fields.Status.Name
	}

	if issue.Fields.Type.Name != "" {
		ticket.IssueType = issue.Fields.Type.Name
	}

	if issue.Fields.Priority != nil {
		ticket.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.Assignee != nil {
		ticket.Assignee = issue.Fields.Assignee.AccountID
	}

	if issue.Fields.Reporter != nil {
		ticket.Reporter = issue.Fields.Reporter.AccountID
	}

	if len(issue.Fields.Labels) > 0 {
		ticket.Labels = issue.Fields.Labels
	}

	return ticket, nil
}

// GetProject retrieves a project from Jira.
// This is a placeholder for the actual implementation.
func (c *Client) GetProject(ctx context.Context, key string) (*domain.Project, error) {
	// TODO: Implement Jira API call to get project
	return nil, fmt.Errorf("jira.Client.GetProject not implemented")
}
