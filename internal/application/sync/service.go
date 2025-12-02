// Package sync contains use cases for synchronization operations.
// This layer orchestrates domain logic and depends only on domain interfaces.
package sync

import (
	"context"

	"github.com/esfisher/jiramd/internal/domain/repository"
)

// Service handles synchronization use cases between Jira and local storage.
// It orchestrates the synchronization logic using domain entities and repository interfaces.
//
// TODO: Add Jira client abstraction (interface) to avoid tight coupling to infrastructure.
// The service should inject a JiraClient interface defined in the application layer.
//
// Error contract: Methods return domain.ErrNotFound when resources don't exist,
// domain.ErrUnauthorized for auth failures, and wrapped errors for other infra issues.
type Service struct {
	ticketRepo  repository.TicketRepository
	commentRepo repository.CommentRepository
	projectRepo repository.ProjectRepository
}

// NewService creates a new sync service with the required repositories.
func NewService(
	ticketRepo repository.TicketRepository,
	commentRepo repository.CommentRepository,
	projectRepo repository.ProjectRepository,
) *Service {
	return &Service{
		ticketRepo:  ticketRepo,
		commentRepo: commentRepo,
		projectRepo: projectRepo,
	}
}

// SyncTicket synchronizes a single ticket between Jira and local storage.
// This is a placeholder for the actual implementation.
func (s *Service) SyncTicket(ctx context.Context, ticketKey string) error {
	// TODO: Implement ticket synchronization logic
	return nil
}

// SyncProject synchronizes all tickets for a project.
// This is a placeholder for the actual implementation.
func (s *Service) SyncProject(ctx context.Context, projectKey string) error {
	// TODO: Implement project synchronization logic
	return nil
}
