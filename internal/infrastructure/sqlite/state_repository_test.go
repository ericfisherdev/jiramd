package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/esfisher/jiramd/internal/domain"
	"github.com/esfisher/jiramd/internal/domain/repository"
)

// Helper function to create a test database
func setupTestDB(t *testing.T) (*Database, func()) {
	t.Helper()

	config := DatabaseConfig{
		Path:         ":memory:", // In-memory database for tests
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// Apply migrations
	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

func TestStateRepository_SaveAndGetTicketState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	tests := []struct {
		name  string
		state *repository.TicketSyncState
	}{
		{
			name: "basic ticket state",
			state: &repository.TicketSyncState{
				TicketKey:          "JMD-123",
				LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
				LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
				LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
				IsDirty:            false,
				ConflictDetected:   false,
			},
		},
		{
			name: "dirty ticket",
			state: &repository.TicketSyncState{
				TicketKey:          "JMD-456",
				LastSynced:         time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Millisecond),
				LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
				LastModifiedJira:   time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Millisecond),
				IsDirty:            true,
				ConflictDetected:   false,
			},
		},
		{
			name: "conflicted ticket",
			state: &repository.TicketSyncState{
				TicketKey:          "JMD-789",
				LastSynced:         time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Millisecond),
				LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
				LastModifiedJira:   time.Now().Add(-30 * time.Minute).UTC().Truncate(time.Millisecond),
				IsDirty:            true,
				ConflictDetected:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save state
			err := repo.SaveTicketState(ctx, tt.state)
			if err != nil {
				t.Fatalf("SaveTicketState failed: %v", err)
			}

			// Get state
			got, err := repo.GetTicketState(ctx, tt.state.TicketKey)
			if err != nil {
				t.Fatalf("GetTicketState failed: %v", err)
			}

			// Verify
			if got.TicketKey != tt.state.TicketKey {
				t.Errorf("TicketKey: got %v, want %v", got.TicketKey, tt.state.TicketKey)
			}
			if !got.LastSynced.Equal(tt.state.LastSynced) {
				t.Errorf("LastSynced: got %v, want %v", got.LastSynced, tt.state.LastSynced)
			}
			if !got.LastModifiedLocal.Equal(tt.state.LastModifiedLocal) {
				t.Errorf("LastModifiedLocal: got %v, want %v", got.LastModifiedLocal, tt.state.LastModifiedLocal)
			}
			if !got.LastModifiedJira.Equal(tt.state.LastModifiedJira) {
				t.Errorf("LastModifiedJira: got %v, want %v", got.LastModifiedJira, tt.state.LastModifiedJira)
			}
			if got.IsDirty != tt.state.IsDirty {
				t.Errorf("IsDirty: got %v, want %v", got.IsDirty, tt.state.IsDirty)
			}
			if got.ConflictDetected != tt.state.ConflictDetected {
				t.Errorf("ConflictDetected: got %v, want %v", got.ConflictDetected, tt.state.ConflictDetected)
			}
		})
	}
}

func TestStateRepository_SaveTicketState_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	ticketKey := "JMD-100"

	// Initial save
	initial := &repository.TicketSyncState{
		TicketKey:          ticketKey,
		LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
		LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
		LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
		IsDirty:            false,
		ConflictDetected:   false,
	}
	if err := repo.SaveTicketState(ctx, initial); err != nil {
		t.Fatalf("initial save failed: %v", err)
	}

	// Update
	updated := &repository.TicketSyncState{
		TicketKey:          ticketKey,
		LastSynced:         time.Now().Add(1 * time.Hour).UTC().Truncate(time.Millisecond),
		LastModifiedLocal:  time.Now().Add(2 * time.Hour).UTC().Truncate(time.Millisecond),
		LastModifiedJira:   time.Now().Add(1 * time.Hour).UTC().Truncate(time.Millisecond),
		IsDirty:            true,
		ConflictDetected:   false,
	}
	if err := repo.SaveTicketState(ctx, updated); err != nil {
		t.Fatalf("update save failed: %v", err)
	}

	// Verify update
	got, err := repo.GetTicketState(ctx, ticketKey)
	if err != nil {
		t.Fatalf("GetTicketState failed: %v", err)
	}

	if !got.LastSynced.Equal(updated.LastSynced) {
		t.Errorf("LastSynced not updated: got %v, want %v", got.LastSynced, updated.LastSynced)
	}
	if got.IsDirty != updated.IsDirty {
		t.Errorf("IsDirty not updated: got %v, want %v", got.IsDirty, updated.IsDirty)
	}
}

func TestStateRepository_GetTicketState_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	_, err := repo.GetTicketState(ctx, "NONEXISTENT-999")
	if err == nil {
		t.Fatal("expected error for non-existent ticket, got nil")
	}

	if !domain.IsNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStateRepository_GetDirtyTickets(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	// Save multiple tickets
	now := time.Now().UTC().Truncate(time.Millisecond)
	tickets := []*repository.TicketSyncState{
		{
			TicketKey:          "JMD-1",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            true,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "JMD-2",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "JMD-3",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            true,
			ConflictDetected:   false,
		},
	}

	for _, ticket := range tickets {
		if err := repo.SaveTicketState(ctx, ticket); err != nil {
			t.Fatalf("failed to save ticket %s: %v", ticket.TicketKey, err)
		}
	}

	// Get dirty tickets
	dirty, err := repo.GetDirtyTickets(ctx)
	if err != nil {
		t.Fatalf("GetDirtyTickets failed: %v", err)
	}

	// Verify count
	if len(dirty) != 2 {
		t.Errorf("expected 2 dirty tickets, got %d", len(dirty))
	}

	// Verify all are dirty
	for _, ticket := range dirty {
		if !ticket.IsDirty {
			t.Errorf("ticket %s should be dirty", ticket.TicketKey)
		}
	}
}

func TestStateRepository_GetConflictedTickets(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	// Save multiple tickets
	now := time.Now().UTC().Truncate(time.Millisecond)
	tickets := []*repository.TicketSyncState{
		{
			TicketKey:          "JMD-1",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            true,
			ConflictDetected:   true,
		},
		{
			TicketKey:          "JMD-2",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "JMD-3",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            true,
			ConflictDetected:   true,
		},
	}

	for _, ticket := range tickets {
		if err := repo.SaveTicketState(ctx, ticket); err != nil {
			t.Fatalf("failed to save ticket %s: %v", ticket.TicketKey, err)
		}
	}

	// Get conflicted tickets
	conflicted, err := repo.GetConflictedTickets(ctx)
	if err != nil {
		t.Fatalf("GetConflictedTickets failed: %v", err)
	}

	// Verify count
	if len(conflicted) != 2 {
		t.Errorf("expected 2 conflicted tickets, got %d", len(conflicted))
	}

	// Verify all have conflicts
	for _, ticket := range conflicted {
		if !ticket.ConflictDetected {
			t.Errorf("ticket %s should have conflict", ticket.TicketKey)
		}
	}
}

func TestStateRepository_GetTicketsModifiedSince(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	tickets := []*repository.TicketSyncState{
		{
			TicketKey:          "JMD-1",
			LastSynced:         now,
			LastModifiedLocal:  twoHoursAgo,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "JMD-2",
			LastSynced:         now,
			LastModifiedLocal:  oneHourAgo,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "JMD-3",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
	}

	for _, ticket := range tickets {
		if err := repo.SaveTicketState(ctx, ticket); err != nil {
			t.Fatalf("failed to save ticket %s: %v", ticket.TicketKey, err)
		}
	}

	// Get tickets modified since 90 minutes ago
	since := now.Add(-90 * time.Minute)
	modified, err := repo.GetTicketsModifiedSince(ctx, since)
	if err != nil {
		t.Fatalf("GetTicketsModifiedSince failed: %v", err)
	}

	// Should return JMD-2 and JMD-3 (modified less than 90 minutes ago)
	if len(modified) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(modified))
	}
}

func TestStateRepository_DeleteTicketState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	// Save a ticket
	state := &repository.TicketSyncState{
		TicketKey:          "JMD-DELETE",
		LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
		LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
		LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
		IsDirty:            false,
		ConflictDetected:   false,
	}
	if err := repo.SaveTicketState(ctx, state); err != nil {
		t.Fatalf("SaveTicketState failed: %v", err)
	}

	// Delete the ticket
	if err := repo.DeleteTicketState(ctx, "JMD-DELETE"); err != nil {
		t.Fatalf("DeleteTicketState failed: %v", err)
	}

	// Verify deletion
	_, err := repo.GetTicketState(ctx, "JMD-DELETE")
	if err == nil {
		t.Fatal("expected error for deleted ticket, got nil")
	}
	if !domain.IsNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStateRepository_DeleteTicketState_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	err := repo.DeleteTicketState(ctx, "NONEXISTENT-999")
	if err == nil {
		t.Fatal("expected error for non-existent ticket, got nil")
	}
	if !domain.IsNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestStateRepository_SaveAndGetProjectState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	state := &repository.ProjectSyncState{
		ProjectKey:           "JMD",
		LastFullSync:         now,
		LastIncrementalSync:  now.Add(1 * time.Hour),
		TicketCount:          42,
	}

	// Save project state
	if err := repo.SaveProjectState(ctx, state); err != nil {
		t.Fatalf("SaveProjectState failed: %v", err)
	}

	// Get project state
	got, err := repo.GetProjectState(ctx, "JMD")
	if err != nil {
		t.Fatalf("GetProjectState failed: %v", err)
	}

	// Verify
	if got.ProjectKey != state.ProjectKey {
		t.Errorf("ProjectKey: got %v, want %v", got.ProjectKey, state.ProjectKey)
	}
	if !got.LastFullSync.Equal(state.LastFullSync) {
		t.Errorf("LastFullSync: got %v, want %v", got.LastFullSync, state.LastFullSync)
	}
	if !got.LastIncrementalSync.Equal(state.LastIncrementalSync) {
		t.Errorf("LastIncrementalSync: got %v, want %v", got.LastIncrementalSync, state.LastIncrementalSync)
	}
	if got.TicketCount != state.TicketCount {
		t.Errorf("TicketCount: got %v, want %v", got.TicketCount, state.TicketCount)
	}
}

func TestStateRepository_GetAllProjectStates(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	projects := []*repository.ProjectSyncState{
		{
			ProjectKey:           "JMD",
			LastFullSync:         now,
			LastIncrementalSync:  now,
			TicketCount:          10,
		},
		{
			ProjectKey:           "TEST",
			LastFullSync:         now,
			LastIncrementalSync:  now,
			TicketCount:          20,
		},
	}

	for _, project := range projects {
		if err := repo.SaveProjectState(ctx, project); err != nil {
			t.Fatalf("failed to save project %s: %v", project.ProjectKey, err)
		}
	}

	// Get all projects
	all, err := repo.GetAllProjectStates(ctx)
	if err != nil {
		t.Fatalf("GetAllProjectStates failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("expected 2 projects, got %d", len(all))
	}
}

func TestStateRepository_DeleteProjectState(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	// Save project and tickets
	now := time.Now().UTC().Truncate(time.Millisecond)
	project := &repository.ProjectSyncState{
		ProjectKey:           "DEL",
		LastFullSync:         now,
		LastIncrementalSync:  now,
		TicketCount:          2,
	}
	if err := repo.SaveProjectState(ctx, project); err != nil {
		t.Fatalf("SaveProjectState failed: %v", err)
	}

	tickets := []*repository.TicketSyncState{
		{
			TicketKey:          "DEL-1",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
		{
			TicketKey:          "DEL-2",
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            false,
			ConflictDetected:   false,
		},
	}
	for _, ticket := range tickets {
		if err := repo.SaveTicketState(ctx, ticket); err != nil {
			t.Fatalf("SaveTicketState failed: %v", err)
		}
	}

	// Delete project
	if err := repo.DeleteProjectState(ctx, "DEL"); err != nil {
		t.Fatalf("DeleteProjectState failed: %v", err)
	}

	// Verify project deleted
	_, err := repo.GetProjectState(ctx, "DEL")
	if err == nil {
		t.Fatal("expected error for deleted project, got nil")
	}
	if !domain.IsNotFoundError(err) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}

	// Verify tickets deleted
	for _, ticket := range tickets {
		_, err := repo.GetTicketState(ctx, ticket.TicketKey)
		if err == nil {
			t.Errorf("expected ticket %s to be deleted", ticket.TicketKey)
		}
	}
}

func TestStateRepository_Transactions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	t.Run("commit transaction", func(t *testing.T) {
		// Begin transaction
		txCtx, err := repo.BeginTransaction(ctx)
		if err != nil {
			t.Fatalf("BeginTransaction failed: %v", err)
		}

		// Save state in transaction
		state := &repository.TicketSyncState{
			TicketKey:          "JMD-TX1",
			LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
			IsDirty:            false,
			ConflictDetected:   false,
		}
		if err := repo.SaveTicketState(txCtx, state); err != nil {
			t.Fatalf("SaveTicketState failed: %v", err)
		}

		// Commit
		if err := repo.Commit(txCtx); err != nil {
			t.Fatalf("Commit failed: %v", err)
		}

		// Verify state was saved
		got, err := repo.GetTicketState(ctx, "JMD-TX1")
		if err != nil {
			t.Fatalf("GetTicketState failed: %v", err)
		}
		if got.TicketKey != "JMD-TX1" {
			t.Errorf("expected ticket JMD-TX1, got %s", got.TicketKey)
		}
	})

	t.Run("rollback transaction", func(t *testing.T) {
		// Begin transaction
		txCtx, err := repo.BeginTransaction(ctx)
		if err != nil {
			t.Fatalf("BeginTransaction failed: %v", err)
		}

		// Save state in transaction
		state := &repository.TicketSyncState{
			TicketKey:          "JMD-TX2",
			LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
			IsDirty:            false,
			ConflictDetected:   false,
		}
		if err := repo.SaveTicketState(txCtx, state); err != nil {
			t.Fatalf("SaveTicketState failed: %v", err)
		}

		// Rollback
		if err := repo.Rollback(txCtx); err != nil {
			t.Fatalf("Rollback failed: %v", err)
		}

		// Verify state was not saved
		_, err = repo.GetTicketState(ctx, "JMD-TX2")
		if err == nil {
			t.Fatal("expected error for rolled back ticket, got nil")
		}
		if !domain.IsNotFoundError(err) {
			t.Errorf("expected ErrNotFound, got: %v", err)
		}
	})
}

func TestStateRepository_ValidationErrors(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewStateRepository(db.DB(), nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		fn      func() error
		wantErr error
	}{
		{
			name: "SaveTicketState with nil state",
			fn: func() error {
				return repo.SaveTicketState(ctx, nil)
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "SaveTicketState with empty key",
			fn: func() error {
				return repo.SaveTicketState(ctx, &repository.TicketSyncState{})
			},
			wantErr: domain.ErrEmptyKey,
		},
		{
			name: "GetTicketState with empty key",
			fn: func() error {
				_, err := repo.GetTicketState(ctx, "")
				return err
			},
			wantErr: domain.ErrEmptyKey,
		},
		{
			name: "SaveProjectState with nil state",
			fn: func() error {
				return repo.SaveProjectState(ctx, nil)
			},
			wantErr: domain.ErrInvalidInput,
		},
		{
			name: "SaveProjectState with empty key",
			fn: func() error {
				return repo.SaveProjectState(ctx, &repository.ProjectSyncState{})
			},
			wantErr: domain.ErrEmptyKey,
		},
		{
			name: "GetProjectState with empty key",
			fn: func() error {
				_, err := repo.GetProjectState(ctx, "")
				return err
			},
			wantErr: domain.ErrEmptyKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !domain.IsError(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}
