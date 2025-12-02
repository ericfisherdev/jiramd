// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// SyncTimestamp is a value object representing a sync-related timestamp.
// All sync timestamps are stored in UTC for consistency.
type SyncTimestamp struct {
	value time.Time
}

// NewSyncTimestamp creates a new SyncTimestamp, normalizing to UTC.
func NewSyncTimestamp(t time.Time) SyncTimestamp {
	return SyncTimestamp{value: t.UTC()}
}

// Time returns the underlying time value.
func (st SyncTimestamp) Time() time.Time {
	return st.value
}

// IsZero returns true if this is the zero time.
func (st SyncTimestamp) IsZero() bool {
	return st.value.IsZero()
}

// Before returns true if this timestamp is before the other.
func (st SyncTimestamp) Before(other SyncTimestamp) bool {
	return st.value.Before(other.value)
}

// After returns true if this timestamp is after the other.
func (st SyncTimestamp) After(other SyncTimestamp) bool {
	return st.value.After(other.value)
}

// SyncStatus represents the current sync status of a ticket.
type SyncStatus string

const (
	// SyncStatusInSync indicates local and remote are synchronized
	SyncStatusInSync SyncStatus = "in_sync"

	// SyncStatusLocalModified indicates local changes not yet pushed
	SyncStatusLocalModified SyncStatus = "local_modified"

	// SyncStatusRemoteModified indicates remote changes not yet pulled
	SyncStatusRemoteModified SyncStatus = "remote_modified"

	// SyncStatusConflict indicates both local and remote modified
	SyncStatusConflict SyncStatus = "conflict"

	// SyncStatusPending indicates sync operation is pending
	SyncStatusPending SyncStatus = "pending"

	// SyncStatusError indicates a sync error occurred
	SyncStatusError SyncStatus = "error"
)

// SyncState represents the synchronization state entity for tracking sync status.
// This entity tracks the overall sync state for a project.
type SyncState struct {
	// ProjectKey identifies which project this state belongs to
	ProjectKey string

	// LastFullSync is when the last full sync completed
	LastFullSync SyncTimestamp

	// LastIncrementalSync is when the last incremental sync completed
	LastIncrementalSync SyncTimestamp

	// TicketCount is the number of tickets currently tracked
	TicketCount int
}

// NewSyncState creates a new SyncState for a project.
func NewSyncState(projectKey string) (*SyncState, error) {
	projectKey = strings.TrimSpace(projectKey)
	if projectKey == "" {
		return nil, fmt.Errorf("%w: project key is required", ErrEmptyKey)
	}

	return &SyncState{
		ProjectKey:          projectKey,
		LastFullSync:        NewSyncTimestamp(time.Time{}),
		LastIncrementalSync: NewSyncTimestamp(time.Time{}),
		TicketCount:         0,
	}, nil
}

// UpdateFullSync updates the last full sync timestamp to now.
func (ss *SyncState) UpdateFullSync() {
	ss.LastFullSync = NewSyncTimestamp(time.Now())
}

// UpdateIncrementalSync updates the last incremental sync timestamp to now.
func (ss *SyncState) UpdateIncrementalSync() {
	ss.LastIncrementalSync = NewSyncTimestamp(time.Now())
}

// TicketState represents the synchronization state for a specific ticket.
// This entity tracks sync metadata for conflict detection and sync orchestration.
type TicketState struct {
	// ProjectKey identifies which project this ticket belongs to
	ProjectKey string

	// TicketKey is the unique ticket identifier
	TicketKey TicketKey

	// JiraUpdated is when Jira last modified this ticket
	JiraUpdated SyncTimestamp

	// LocalModified is when the local file was last modified (nil if never modified locally)
	LocalModified *SyncTimestamp

	// LastSynced is when this ticket was last successfully synced
	LastSynced SyncTimestamp

	// ContentHash is the MD5 hash of the ticket content at last sync
	ContentHash string

	// Status is the current sync status
	Status SyncStatus
}

// NewTicketState creates a new TicketState for a ticket.
func NewTicketState(projectKey string, ticketKey TicketKey, jiraUpdated time.Time) (*TicketState, error) {
	projectKey = strings.TrimSpace(projectKey)
	if projectKey == "" {
		return nil, fmt.Errorf("%w: project key is required", ErrEmptyKey)
	}
	if ticketKey.IsZero() {
		return nil, fmt.Errorf("%w: ticket key is required", ErrInvalidTicketKey)
	}

	return &TicketState{
		ProjectKey:    projectKey,
		TicketKey:     ticketKey,
		JiraUpdated:   NewSyncTimestamp(jiraUpdated),
		LocalModified: nil,
		LastSynced:    NewSyncTimestamp(time.Now()),
		ContentHash:   "",
		Status:        SyncStatusInSync,
	}, nil
}

// MarkLocalModified marks the ticket as locally modified at the given time.
func (ts *TicketState) MarkLocalModified(modifiedAt time.Time) {
	timestamp := NewSyncTimestamp(modifiedAt)
	ts.LocalModified = &timestamp
	ts.Status = SyncStatusLocalModified
}

// DetectConflict checks if there is a sync conflict.
// A conflict occurs when both local and remote have been modified since last sync.
func (ts *TicketState) DetectConflict() bool {
	if ts.LocalModified == nil {
		return false // No local modifications
	}

	// Conflict if both local and Jira modified after last sync
	localModifiedAfterSync := ts.LocalModified.After(ts.LastSynced)
	jiraModifiedAfterSync := ts.JiraUpdated.After(ts.LastSynced)

	if localModifiedAfterSync && jiraModifiedAfterSync {
		ts.Status = SyncStatusConflict
		return true
	}

	return false
}

// UpdateSynced updates the state after a successful sync.
func (ts *TicketState) UpdateSynced(contentHash string, jiraUpdated time.Time) {
	ts.LastSynced = NewSyncTimestamp(time.Now())
	ts.ContentHash = contentHash
	ts.JiraUpdated = NewSyncTimestamp(jiraUpdated)
	ts.LocalModified = nil
	ts.Status = SyncStatusInSync
}

// SyncResult represents the result of a sync operation.
// This is a value object that encapsulates sync operation outcomes.
type SyncResult struct {
	// TicketKey identifies which ticket was synced
	TicketKey TicketKey

	// Success indicates if the sync succeeded
	Success bool

	// Error contains the error message if sync failed
	Error string

	// ConflictDetected indicates if a conflict was detected
	ConflictDetected bool

	// OperationsPerformed lists which operations were performed
	OperationsPerformed []string
}

// NewSyncResult creates a successful sync result.
func NewSyncResult(ticketKey TicketKey) *SyncResult {
	return &SyncResult{
		TicketKey:           ticketKey,
		Success:             true,
		ConflictDetected:    false,
		OperationsPerformed: make([]string, 0),
	}
}

// MarkFailed marks the sync result as failed with an error.
func (sr *SyncResult) MarkFailed(err error) {
	sr.Success = false
	if err != nil {
		sr.Error = err.Error()
	}
}

// MarkConflict marks the sync result as having a conflict.
func (sr *SyncResult) MarkConflict() {
	sr.ConflictDetected = true
	sr.AddOperation("conflict_detected")
}

// AddOperation adds an operation to the list of performed operations.
func (sr *SyncResult) AddOperation(operation string) {
	sr.OperationsPerformed = append(sr.OperationsPerformed, operation)
}

// OperationType defines the type of pending operation.
type OperationType string

const (
	// OpPushStatus indicates a status field push operation
	OpPushStatus OperationType = "push_status"

	// OpPushField indicates a custom field push operation
	OpPushField OperationType = "push_field"

	// OpPostComment indicates a comment post operation
	OpPostComment OperationType = "post_comment"

	// OpPullTicket indicates a ticket pull operation
	OpPullTicket OperationType = "pull_ticket"

	// OpPullComments indicates a comments pull operation
	OpPullComments OperationType = "pull_comments"
)

// PendingOperation represents a queued sync operation that needs to be performed.
// This entity tracks operations that failed or are scheduled for later execution.
type PendingOperation struct {
	// ID is the unique identifier for this operation
	ID int64

	// ProjectKey identifies which project this operation belongs to
	ProjectKey string

	// TicketKey identifies which ticket this operation affects
	TicketKey TicketKey

	// Operation specifies what type of operation to perform
	Operation OperationType

	// Payload contains operation-specific data (JSON serialized)
	Payload string

	// CreatedAt is when this operation was queued
	CreatedAt SyncTimestamp

	// Attempts is how many times this operation has been attempted
	Attempts int

	// LastError contains the error from the last attempt (if any)
	LastError string
}

// NewPendingOperation creates a new pending operation.
func NewPendingOperation(projectKey string, ticketKey TicketKey, operation OperationType, payload string) (*PendingOperation, error) {
	projectKey = strings.TrimSpace(projectKey)
	if projectKey == "" {
		return nil, fmt.Errorf("%w: project key is required", ErrEmptyKey)
	}
	if ticketKey.IsZero() {
		return nil, fmt.Errorf("%w: ticket key is required", ErrInvalidTicketKey)
	}

	// Validate operation type
	switch operation {
	case OpPushStatus, OpPushField, OpPostComment, OpPullTicket, OpPullComments:
		// Valid
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidOperation, operation)
	}

	return &PendingOperation{
		ProjectKey: projectKey,
		TicketKey:  ticketKey,
		Operation:  operation,
		Payload:    payload,
		CreatedAt:  NewSyncTimestamp(time.Now()),
		Attempts:   0,
		LastError:  "",
	}, nil
}

// RecordAttempt records a failed attempt with the given error.
func (po *PendingOperation) RecordAttempt(err error) {
	po.Attempts++
	if err != nil {
		po.LastError = err.Error()
	}
}

// ShouldRetry determines if this operation should be retried.
// Returns true if attempts < max retries (currently 3).
func (po *PendingOperation) ShouldRetry() bool {
	const maxRetries = 3
	return po.Attempts < maxRetries
}
