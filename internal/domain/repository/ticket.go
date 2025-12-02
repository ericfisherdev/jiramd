// Package repository defines interfaces for data access.
// These interfaces are part of the domain layer and define contracts
// that infrastructure implementations must fulfill.
package repository

import (
	"context"

	"github.com/esfisher/jiramd/internal/domain"
)

// TicketRepository defines the interface for ticket persistence operations.
// Implementations of this interface will be provided by the infrastructure layer.
type TicketRepository interface {
	// Save persists a ticket to storage
	Save(ctx context.Context, ticket *domain.Ticket) error

	// FindByKey retrieves a ticket by its key
	FindByKey(ctx context.Context, key string) (*domain.Ticket, error)

	// FindAll retrieves all tickets
	FindAll(ctx context.Context) ([]*domain.Ticket, error)

	// Delete removes a ticket from storage
	Delete(ctx context.Context, key string) error

	// Update updates an existing ticket
	Update(ctx context.Context, ticket *domain.Ticket) error
}

// CommentRepository defines the interface for comment persistence operations.
type CommentRepository interface {
	// Save persists a comment to storage
	Save(ctx context.Context, comment *domain.Comment) error

	// FindByTicketKey retrieves all comments for a ticket
	FindByTicketKey(ctx context.Context, ticketKey string) ([]*domain.Comment, error)

	// FindByID retrieves a comment by its ID
	FindByID(ctx context.Context, id string) (*domain.Comment, error)

	// Delete removes a comment from storage
	Delete(ctx context.Context, id string) error
}

// ProjectRepository defines the interface for project persistence operations.
type ProjectRepository interface {
	// Save persists a project to storage
	Save(ctx context.Context, project *domain.Project) error

	// FindByKey retrieves a project by its key
	FindByKey(ctx context.Context, key string) (*domain.Project, error)

	// FindAll retrieves all projects
	FindAll(ctx context.Context) ([]*domain.Project, error)

	// Delete removes a project from storage
	Delete(ctx context.Context, key string) error
}
