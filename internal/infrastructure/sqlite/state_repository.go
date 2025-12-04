// Package sqlite provides SQLite-based implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
	"github.com/esfisher/jiramd/internal/domain/repository"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// txContextKey is the context key for storing transaction.
	txContextKey contextKey = "sqlite_tx"
)

// StateRepository implements repository.StateRepository using SQLite.
type StateRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewStateRepository creates a new SQLite-backed StateRepository.
// The database connection must be initialized and migrations applied before use.
func NewStateRepository(db *sql.DB, logger *slog.Logger) *StateRepository {
	if logger == nil {
		logger = slog.Default()
	}
	return &StateRepository{
		db:     db,
		logger: logger,
	}
}

// SaveTicketState persists the synchronization state of a ticket.
// Implements repository.StateRepository.SaveTicketState.
func (r *StateRepository) SaveTicketState(ctx context.Context, state *repository.TicketSyncState) error {
	if state == nil {
		return fmt.Errorf("%w: state cannot be nil", domain.ErrInvalidInput)
	}
	if state.TicketKey == "" {
		return fmt.Errorf("%w: ticket key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	query := `
		INSERT INTO ticket_sync_state (
			ticket_key,
			last_synced,
			last_modified_local,
			last_modified_jira,
			is_dirty,
			conflict_detected,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(ticket_key) DO UPDATE SET
			last_synced = excluded.last_synced,
			last_modified_local = excluded.last_modified_local,
			last_modified_jira = excluded.last_modified_jira,
			is_dirty = excluded.is_dirty,
			conflict_detected = excluded.conflict_detected,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := exec.ExecContext(ctx, query,
		state.TicketKey,
		formatTimestamp(state.LastSynced),
		formatTimestamp(state.LastModifiedLocal),
		formatTimestamp(state.LastModifiedJira),
		state.IsDirty,
		state.ConflictDetected,
	)
	if err != nil {
		r.logger.Error("failed to save ticket state",
			"ticket_key", state.TicketKey,
			"error", err)
		return fmt.Errorf("failed to save ticket state: %w", err)
	}

	r.logger.Debug("saved ticket state",
		"ticket_key", state.TicketKey,
		"is_dirty", state.IsDirty,
		"conflict_detected", state.ConflictDetected)

	return nil
}

// GetTicketState retrieves the synchronization state of a ticket.
// Implements repository.StateRepository.GetTicketState.
func (r *StateRepository) GetTicketState(ctx context.Context, ticketKey string) (*repository.TicketSyncState, error) {
	if ticketKey == "" {
		return nil, fmt.Errorf("%w: ticket key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	query := `
		SELECT
			ticket_key,
			last_synced,
			last_modified_local,
			last_modified_jira,
			is_dirty,
			conflict_detected
		FROM ticket_sync_state
		WHERE ticket_key = ?
	`

	var state repository.TicketSyncState
	var lastSynced, lastModifiedLocal, lastModifiedJira string

	err := exec.QueryRowContext(ctx, query, ticketKey).Scan(
		&state.TicketKey,
		&lastSynced,
		&lastModifiedLocal,
		&lastModifiedJira,
		&state.IsDirty,
		&state.ConflictDetected,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: ticket state not found for key %s", domain.ErrNotFound, ticketKey)
		}
		r.logger.Error("failed to get ticket state",
			"ticket_key", ticketKey,
			"error", err)
		return nil, fmt.Errorf("failed to get ticket state: %w", err)
	}

	// Parse timestamps
	state.LastSynced = parseTimestamp(lastSynced)
	state.LastModifiedLocal = parseTimestamp(lastModifiedLocal)
	state.LastModifiedJira = parseTimestamp(lastModifiedJira)

	return &state, nil
}

// GetTicketsModifiedSince retrieves all tickets with local modifications after the given time.
// Implements repository.StateRepository.GetTicketsModifiedSince.
func (r *StateRepository) GetTicketsModifiedSince(ctx context.Context, since time.Time) ([]*repository.TicketSyncState, error) {
	exec := r.getExecutor(ctx)

	query := `
		SELECT
			ticket_key,
			last_synced,
			last_modified_local,
			last_modified_jira,
			is_dirty,
			conflict_detected
		FROM ticket_sync_state
		WHERE last_modified_local > ?
		ORDER BY last_modified_local DESC
	`

	rows, err := exec.QueryContext(ctx, query, formatTimestamp(since))
	if err != nil {
		r.logger.Error("failed to query tickets modified since",
			"since", since,
			"error", err)
		return nil, fmt.Errorf("failed to query tickets modified since: %w", err)
	}
	defer rows.Close()

	return r.scanTicketStates(rows)
}

// GetDirtyTickets retrieves all tickets marked as dirty.
// Implements repository.StateRepository.GetDirtyTickets.
func (r *StateRepository) GetDirtyTickets(ctx context.Context) ([]*repository.TicketSyncState, error) {
	exec := r.getExecutor(ctx)

	query := `
		SELECT
			ticket_key,
			last_synced,
			last_modified_local,
			last_modified_jira,
			is_dirty,
			conflict_detected
		FROM ticket_sync_state
		WHERE is_dirty = 1
		ORDER BY last_modified_local DESC
	`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to query dirty tickets", "error", err)
		return nil, fmt.Errorf("failed to query dirty tickets: %w", err)
	}
	defer rows.Close()

	return r.scanTicketStates(rows)
}

// GetConflictedTickets retrieves all tickets with detected conflicts.
// Implements repository.StateRepository.GetConflictedTickets.
func (r *StateRepository) GetConflictedTickets(ctx context.Context) ([]*repository.TicketSyncState, error) {
	exec := r.getExecutor(ctx)

	query := `
		SELECT
			ticket_key,
			last_synced,
			last_modified_local,
			last_modified_jira,
			is_dirty,
			conflict_detected
		FROM ticket_sync_state
		WHERE conflict_detected = 1
		ORDER BY last_modified_local DESC
	`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to query conflicted tickets", "error", err)
		return nil, fmt.Errorf("failed to query conflicted tickets: %w", err)
	}
	defer rows.Close()

	return r.scanTicketStates(rows)
}

// DeleteTicketState removes the synchronization state for a ticket.
// Implements repository.StateRepository.DeleteTicketState.
func (r *StateRepository) DeleteTicketState(ctx context.Context, ticketKey string) error {
	if ticketKey == "" {
		return fmt.Errorf("%w: ticket key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	query := `DELETE FROM ticket_sync_state WHERE ticket_key = ?`

	result, err := exec.ExecContext(ctx, query, ticketKey)
	if err != nil {
		r.logger.Error("failed to delete ticket state",
			"ticket_key", ticketKey,
			"error", err)
		return fmt.Errorf("failed to delete ticket state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: ticket state not found for key %s", domain.ErrNotFound, ticketKey)
	}

	r.logger.Debug("deleted ticket state", "ticket_key", ticketKey)
	return nil
}

// SaveProjectState persists the synchronization state of a project.
// Implements repository.StateRepository.SaveProjectState.
func (r *StateRepository) SaveProjectState(ctx context.Context, state *repository.ProjectSyncState) error {
	if state == nil {
		return fmt.Errorf("%w: state cannot be nil", domain.ErrInvalidInput)
	}
	if state.ProjectKey == "" {
		return fmt.Errorf("%w: project key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	query := `
		INSERT INTO project_sync_state (
			project_key,
			last_full_sync,
			last_incremental_sync,
			ticket_count,
			updated_at
		) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(project_key) DO UPDATE SET
			last_full_sync = excluded.last_full_sync,
			last_incremental_sync = excluded.last_incremental_sync,
			ticket_count = excluded.ticket_count,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := exec.ExecContext(ctx, query,
		state.ProjectKey,
		formatTimestampNullable(state.LastFullSync),
		formatTimestampNullable(state.LastIncrementalSync),
		state.TicketCount,
	)
	if err != nil {
		r.logger.Error("failed to save project state",
			"project_key", state.ProjectKey,
			"error", err)
		return fmt.Errorf("failed to save project state: %w", err)
	}

	r.logger.Debug("saved project state",
		"project_key", state.ProjectKey,
		"ticket_count", state.TicketCount)

	return nil
}

// GetProjectState retrieves the synchronization state of a project.
// Implements repository.StateRepository.GetProjectState.
func (r *StateRepository) GetProjectState(ctx context.Context, projectKey string) (*repository.ProjectSyncState, error) {
	if projectKey == "" {
		return nil, fmt.Errorf("%w: project key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	query := `
		SELECT
			project_key,
			last_full_sync,
			last_incremental_sync,
			ticket_count
		FROM project_sync_state
		WHERE project_key = ?
	`

	var state repository.ProjectSyncState
	var lastFullSync, lastIncrementalSync sql.NullString

	err := exec.QueryRowContext(ctx, query, projectKey).Scan(
		&state.ProjectKey,
		&lastFullSync,
		&lastIncrementalSync,
		&state.TicketCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: project state not found for key %s", domain.ErrNotFound, projectKey)
		}
		r.logger.Error("failed to get project state",
			"project_key", projectKey,
			"error", err)
		return nil, fmt.Errorf("failed to get project state: %w", err)
	}

	// Parse nullable timestamps
	if lastFullSync.Valid {
		state.LastFullSync = parseTimestamp(lastFullSync.String)
	}
	if lastIncrementalSync.Valid {
		state.LastIncrementalSync = parseTimestamp(lastIncrementalSync.String)
	}

	return &state, nil
}

// GetAllProjectStates retrieves all project states.
// Implements repository.StateRepository.GetAllProjectStates.
func (r *StateRepository) GetAllProjectStates(ctx context.Context) ([]*repository.ProjectSyncState, error) {
	exec := r.getExecutor(ctx)

	query := `
		SELECT
			project_key,
			last_full_sync,
			last_incremental_sync,
			ticket_count
		FROM project_sync_state
		ORDER BY project_key
	`

	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("failed to query all project states", "error", err)
		return nil, fmt.Errorf("failed to query all project states: %w", err)
	}
	defer rows.Close()

	var states []*repository.ProjectSyncState
	for rows.Next() {
		var state repository.ProjectSyncState
		var lastFullSync, lastIncrementalSync sql.NullString

		if err := rows.Scan(
			&state.ProjectKey,
			&lastFullSync,
			&lastIncrementalSync,
			&state.TicketCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan project state: %w", err)
		}

		// Parse nullable timestamps
		if lastFullSync.Valid {
			state.LastFullSync = parseTimestamp(lastFullSync.String)
		}
		if lastIncrementalSync.Valid {
			state.LastIncrementalSync = parseTimestamp(lastIncrementalSync.String)
		}

		states = append(states, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate project states: %w", err)
	}

	return states, nil
}

// DeleteProjectState removes the synchronization state for a project.
// Implements repository.StateRepository.DeleteProjectState.
func (r *StateRepository) DeleteProjectState(ctx context.Context, projectKey string) error {
	if projectKey == "" {
		return fmt.Errorf("%w: project key cannot be empty", domain.ErrEmptyKey)
	}

	exec := r.getExecutor(ctx)

	// Delete in transaction if not already in one
	inTransaction := r.isInTransaction(ctx)
	if !inTransaction {
		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()
		exec = tx
	}

	// Delete all ticket states for this project first
	// Note: This assumes ticket keys start with project key (e.g., "JMD-123")
	deleteTicketsQuery := `DELETE FROM ticket_sync_state WHERE ticket_key LIKE ? || '-%'`
	if _, err := exec.ExecContext(ctx, deleteTicketsQuery, projectKey); err != nil {
		r.logger.Error("failed to delete project ticket states",
			"project_key", projectKey,
			"error", err)
		return fmt.Errorf("failed to delete project ticket states: %w", err)
	}

	// Delete project state
	deleteProjectQuery := `DELETE FROM project_sync_state WHERE project_key = ?`
	result, err := exec.ExecContext(ctx, deleteProjectQuery, projectKey)
	if err != nil {
		r.logger.Error("failed to delete project state",
			"project_key", projectKey,
			"error", err)
		return fmt.Errorf("failed to delete project state: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: project state not found for key %s", domain.ErrNotFound, projectKey)
	}

	// Commit if we started the transaction
	if !inTransaction {
		if err := exec.(*sql.Tx).Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	r.logger.Debug("deleted project state", "project_key", projectKey)
	return nil
}

// BeginTransaction starts a new transaction.
// Implements repository.StateRepository.BeginTransaction.
func (r *StateRepository) BeginTransaction(ctx context.Context) (context.Context, error) {
	if r.isInTransaction(ctx) {
		return nil, fmt.Errorf("%w: transaction already active", domain.ErrInvalidInput)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	r.logger.Debug("transaction started")
	return context.WithValue(ctx, txContextKey, tx), nil
}

// Commit commits the current transaction.
// Implements repository.StateRepository.Commit.
func (r *StateRepository) Commit(ctx context.Context) error {
	tx := r.getTransaction(ctx)
	if tx == nil {
		return fmt.Errorf("%w: no active transaction", domain.ErrInvalidInput)
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Debug("transaction committed")
	return nil
}

// Rollback rolls back the current transaction.
// Implements repository.StateRepository.Rollback.
func (r *StateRepository) Rollback(ctx context.Context) error {
	tx := r.getTransaction(ctx)
	if tx == nil {
		return fmt.Errorf("%w: no active transaction", domain.ErrInvalidInput)
	}

	if err := tx.Rollback(); err != nil {
		r.logger.Error("failed to rollback transaction", "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	r.logger.Debug("transaction rolled back")
	return nil
}

// Helper functions

// getTransaction extracts transaction from context.
func (r *StateRepository) getTransaction(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txContextKey).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// isInTransaction checks if context has an active transaction.
func (r *StateRepository) isInTransaction(ctx context.Context) bool {
	return r.getTransaction(ctx) != nil
}

// getExecutor returns the appropriate executor (transaction or database).
func (r *StateRepository) getExecutor(ctx context.Context) executor {
	if tx := r.getTransaction(ctx); tx != nil {
		return tx
	}
	return r.db
}

// executor is an interface that both *sql.DB and *sql.Tx implement.
type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// scanTicketStates is a helper function to scan multiple ticket states from rows.
func (r *StateRepository) scanTicketStates(rows *sql.Rows) ([]*repository.TicketSyncState, error) {
	var states []*repository.TicketSyncState

	for rows.Next() {
		var state repository.TicketSyncState
		var lastSynced, lastModifiedLocal, lastModifiedJira string

		if err := rows.Scan(
			&state.TicketKey,
			&lastSynced,
			&lastModifiedLocal,
			&lastModifiedJira,
			&state.IsDirty,
			&state.ConflictDetected,
		); err != nil {
			return nil, fmt.Errorf("failed to scan ticket state: %w", err)
		}

		// Parse timestamps
		state.LastSynced = parseTimestamp(lastSynced)
		state.LastModifiedLocal = parseTimestamp(lastModifiedLocal)
		state.LastModifiedJira = parseTimestamp(lastModifiedJira)

		states = append(states, &state)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate ticket states: %w", err)
	}

	return states, nil
}

// formatTimestamp converts time.Time to SQLite timestamp string.
func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "1970-01-01 00:00:00"
	}
	return t.UTC().Format("2006-01-02 15:04:05.000")
}

// formatTimestampNullable converts time.Time to nullable SQLite timestamp.
func formatTimestampNullable(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return formatTimestamp(t)
}

// parseTimestamp converts SQLite timestamp string to time.Time.
func parseTimestamp(s string) time.Time {
	if s == "" || s == "1970-01-01 00:00:00" {
		return time.Time{}
	}

	// Try RFC3339 format first (what SQLite may return)
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t.UTC()
	}

	// Try parsing with milliseconds
	t, err = time.Parse("2006-01-02 15:04:05.000", s)
	if err == nil {
		return t.UTC()
	}

	// Fall back to seconds precision
	t, err = time.Parse("2006-01-02 15:04:05", s)
	if err == nil {
		return t.UTC()
	}

	// Log warning and return zero time
	slog.Warn("failed to parse timestamp, using zero time",
		"timestamp", s,
		"error", err)
	return time.Time{}
}
