// Package repository defines interfaces for data access.
// These interfaces are part of the domain layer and define contracts
// that infrastructure implementations must fulfill.
package repository

import (
	"context"
	"time"
)

// TicketSyncState represents the synchronization state of a single ticket.
// This tracks when the ticket was last synced and any local modifications.
type TicketSyncState struct {
	// TicketKey is the unique Jira ticket identifier
	TicketKey string

	// LastSynced is when the ticket was last successfully synced with Jira
	LastSynced time.Time

	// LastModifiedLocal is when the local markdown file was last modified
	LastModifiedLocal time.Time

	// LastModifiedJira is the last update timestamp from Jira
	LastModifiedJira time.Time

	// IsDirty indicates if the local file has unsynced changes
	IsDirty bool

	// ConflictDetected indicates if both local and Jira were modified since last sync
	ConflictDetected bool
}

// ProjectSyncState represents the synchronization state of a project.
type ProjectSyncState struct {
	// ProjectKey is the unique project identifier
	ProjectKey string

	// LastFullSync is when the project had its last complete sync
	LastFullSync time.Time

	// LastIncrementalSync is when the project had its last incremental sync
	LastIncrementalSync time.Time

	// TicketCount is the total number of tickets tracked for this project
	TicketCount int
}

// StateRepository defines the interface for sync state persistence.
// This interface abstracts storage of synchronization state and metadata.
//
// Implementations must:
//   - Persist sync timestamps for conflict detection
//   - Track local modifications to trigger push operations
//   - Store project metadata for incremental sync
//   - Provide transactional guarantees for state updates
//   - Handle concurrent access safely
//
// Domain errors that methods should return:
//   - ErrNotFound: when a state record doesn't exist
//   - ErrInvalidInput: when state data is invalid
//   - ErrConflict: when concurrent state updates conflict
type StateRepository interface {
	// SaveTicketState persists the synchronization state of a ticket.
	// Creates a new record if the ticket state doesn't exist, updates if it does.
	// Returns ErrInvalidInput if the state data is invalid.
	SaveTicketState(ctx context.Context, state *TicketSyncState) error

	// GetTicketState retrieves the synchronization state of a ticket.
	// Returns ErrNotFound if no state exists for the given ticket key.
	GetTicketState(ctx context.Context, ticketKey string) (*TicketSyncState, error)

	// GetTicketsModifiedSince retrieves all tickets with local modifications after the given time.
	// Used to identify tickets that need to be pushed to Jira.
	// Returns empty slice if no tickets have been modified.
	GetTicketsModifiedSince(ctx context.Context, since time.Time) ([]*TicketSyncState, error)

	// GetDirtyTickets retrieves all tickets marked as dirty (having unsynced local changes).
	// Used during sync operations to identify tickets requiring push.
	// Returns empty slice if no dirty tickets exist.
	GetDirtyTickets(ctx context.Context) ([]*TicketSyncState, error)

	// GetConflictedTickets retrieves all tickets with detected conflicts.
	// Returns empty slice if no conflicts exist.
	GetConflictedTickets(ctx context.Context) ([]*TicketSyncState, error)

	// DeleteTicketState removes the synchronization state for a ticket.
	// Used when a ticket is deleted from both Jira and local storage.
	// Returns ErrNotFound if the state doesn't exist.
	DeleteTicketState(ctx context.Context, ticketKey string) error

	// SaveProjectState persists the synchronization state of a project.
	// Creates a new record if the project state doesn't exist, updates if it does.
	SaveProjectState(ctx context.Context, state *ProjectSyncState) error

	// GetProjectState retrieves the synchronization state of a project.
	// Returns ErrNotFound if no state exists for the given project key.
	GetProjectState(ctx context.Context, projectKey string) (*ProjectSyncState, error)

	// GetAllProjectStates retrieves all project states.
	// Used to iterate over all tracked projects during sync operations.
	// Returns empty slice if no projects are tracked.
	GetAllProjectStates(ctx context.Context) ([]*ProjectSyncState, error)

	// DeleteProjectState removes the synchronization state for a project.
	// Also removes all associated ticket states.
	// Returns ErrNotFound if the state doesn't exist.
	DeleteProjectState(ctx context.Context, projectKey string) error

	// BeginTransaction starts a new transaction for atomic state updates.
	// Multiple state operations can be grouped to ensure consistency.
	// The returned context must be used for all operations within the transaction.
	// Call Commit() to persist changes or Rollback() to discard them.
	BeginTransaction(ctx context.Context) (context.Context, error)

	// Commit commits the current transaction.
	// All state changes made within the transaction become permanent.
	// Returns ErrInvalidInput if called without an active transaction.
	Commit(ctx context.Context) error

	// Rollback rolls back the current transaction.
	// All state changes made within the transaction are discarded.
	// Returns ErrInvalidInput if called without an active transaction.
	Rollback(ctx context.Context) error
}
