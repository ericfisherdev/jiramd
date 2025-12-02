// Package repository defines interfaces for data access.
// These interfaces are part of the domain layer and define contracts
// that infrastructure implementations must fulfill.
package repository

import (
	"context"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
)

// JiraRepository defines the interface for Jira Cloud API operations.
// This interface abstracts communication with Jira Cloud REST API.
//
// Implementations must:
//   - Handle authentication with Jira Cloud (API token)
//   - Map Jira API responses to domain entities
//   - Map HTTP status codes to domain errors (404->ErrNotFound, 401/403->ErrUnauthorized)
//   - Implement pagination for large result sets
//   - Handle rate limiting and retries appropriately
//
// Domain errors that methods should return:
//   - ErrNotFound: when a ticket, comment, or project is not found
//   - ErrUnauthorized: when authentication fails or user lacks permissions
//   - ErrInvalidInput: when provided data fails validation
//   - ErrConflict: when there's an optimistic locking conflict
type JiraRepository interface {
	// FetchTicket retrieves a single ticket from Jira by its key.
	// Returns ErrNotFound if the ticket doesn't exist.
	// Returns ErrUnauthorized if the user lacks permission to view the ticket.
	FetchTicket(ctx context.Context, key string) (*domain.Ticket, error)

	// FetchTicketsModifiedSince retrieves tickets modified after the given timestamp.
	// Uses JQL: "project = X AND updated >= timestamp ORDER BY updated ASC"
	// Results should be paginated to avoid memory issues with large result sets.
	// Returns empty slice if no tickets match the criteria.
	FetchTicketsModifiedSince(ctx context.Context, projectKey string, since time.Time) ([]*domain.Ticket, error)

	// FetchAllTickets retrieves all tickets for a project.
	// Uses JQL: "project = X ORDER BY updated DESC"
	// Results should be paginated to avoid memory issues with large result sets.
	FetchAllTickets(ctx context.Context, projectKey string) ([]*domain.Ticket, error)

	// UpdateTicket pushes local ticket changes to Jira.
	// Only updates fields that have changed to minimize API calls.
	// Returns the updated ticket with the authoritative Jira timestamp for version tracking.
	// Returns ErrNotFound if the ticket no longer exists in Jira.
	// Returns ErrConflict if the ticket was modified by another user since last fetch.
	// Returns ErrUnauthorized if the user lacks permission to edit the ticket.
	UpdateTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Ticket, error)

	// FetchComments retrieves all comments for a given ticket.
	// Returns empty slice if the ticket has no comments.
	// Returns ErrNotFound if the ticket doesn't exist.
	FetchComments(ctx context.Context, ticketKey string) ([]*domain.Comment, error)

	// AddComment adds a new comment to a Jira ticket.
	// Returns the created comment with its Jira-assigned ID populated.
	// Returns ErrNotFound if the ticket doesn't exist.
	// Returns ErrUnauthorized if the user lacks permission to comment.
	AddComment(ctx context.Context, ticketKey string, comment *domain.Comment) (*domain.Comment, error)

	// FetchProject retrieves project metadata from Jira.
	// Returns ErrNotFound if the project doesn't exist.
	// Returns ErrUnauthorized if the user lacks permission to view the project.
	FetchProject(ctx context.Context, projectKey string) (*domain.Project, error)

	// FetchProjects retrieves all projects the authenticated user can access.
	// Returns empty slice if the user has no accessible projects.
	FetchProjects(ctx context.Context) ([]*domain.Project, error)
}
