// +build integration

package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/esfisher/jiramd/internal/domain/repository"
)

func TestIntegration_PersistenceBetweenConnections(t *testing.T) {
	// Create temp database file
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First connection - save state
	{
		config := DatabaseConfig{
			Path:         dbPath,
			MaxOpenConns: 1,
			BusyTimeout:  5 * time.Second,
		}

		db, err := NewDatabase(config, nil)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}

		ctx := context.Background()
		if err := db.Migrate(ctx); err != nil {
			t.Fatalf("failed to migrate: %v", err)
		}

		repo := NewStateRepository(db.DB(), nil)

		state := &repository.TicketSyncState{
			TicketKey:          "JMD-PERSIST",
			LastSynced:         time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedLocal:  time.Now().UTC().Truncate(time.Millisecond),
			LastModifiedJira:   time.Now().UTC().Truncate(time.Millisecond),
			IsDirty:            true,
			ConflictDetected:   false,
		}

		if err := repo.SaveTicketState(ctx, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		db.Close()
	}

	// Second connection - verify state persisted
	{
		config := DatabaseConfig{
			Path:         dbPath,
			MaxOpenConns: 1,
			BusyTimeout:  5 * time.Second,
		}

		db, err := NewDatabase(config, nil)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		if err := db.Migrate(ctx); err != nil {
			t.Fatalf("failed to migrate: %v", err)
		}

		repo := NewStateRepository(db.DB(), nil)

		state, err := repo.GetTicketState(ctx, "JMD-PERSIST")
		if err != nil {
			t.Fatalf("failed to get state: %v", err)
		}

		if state.TicketKey != "JMD-PERSIST" {
			t.Errorf("expected JMD-PERSIST, got %s", state.TicketKey)
		}

		if !state.IsDirty {
			t.Error("expected dirty flag to be true")
		}
	}
}

func TestIntegration_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Set file permissions
	if err := db.SetFilePermissions(); err != nil {
		t.Fatalf("failed to set file permissions: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	perm := info.Mode().Perm()
	expected := os.FileMode(0600)

	if perm != expected {
		t.Errorf("expected permissions %o, got %o", expected, perm)
	}
}

func TestIntegration_ConcurrentReads(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := NewStateRepository(db.DB(), nil)

	// Save some test data
	now := time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 10; i++ {
		state := &repository.TicketSyncState{
			TicketKey:          "JMD-" + string(rune('0'+i)),
			LastSynced:         now,
			LastModifiedLocal:  now,
			LastModifiedJira:   now,
			IsDirty:            i%2 == 0,
			ConflictDetected:   false,
		}
		if err := repo.SaveTicketState(ctx, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}
	}

	// Concurrent reads
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			ticketKey := "JMD-" + string(rune('0'+idx))
			_, err := repo.GetTicketState(ctx, ticketKey)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent read error: %v", err)
	}
}

func TestIntegration_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := NewStateRepository(db.DB(), nil)

	// Concurrent writes to different tickets
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	now := time.Now().UTC().Truncate(time.Millisecond)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			state := &repository.TicketSyncState{
				TicketKey:          "JMD-CONCURRENT-" + string(rune('0'+idx)),
				LastSynced:         now,
				LastModifiedLocal:  now,
				LastModifiedJira:   now,
				IsDirty:            false,
				ConflictDetected:   false,
			}

			if err := repo.SaveTicketState(ctx, state); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent write error: %v", err)
	}

	// Verify all writes succeeded
	dirty, err := repo.GetDirtyTickets(ctx)
	if err != nil {
		t.Fatalf("failed to get dirty tickets: %v", err)
	}

	// We wrote 10 non-dirty tickets, so dirty should be empty
	if len(dirty) != 0 {
		t.Errorf("expected 0 dirty tickets, got %d", len(dirty))
	}
}

func TestIntegration_ConcurrentUpdates(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := NewStateRepository(db.DB(), nil)

	// Create initial state
	now := time.Now().UTC().Truncate(time.Millisecond)
	initial := &repository.TicketSyncState{
		TicketKey:          "JMD-UPDATE",
		LastSynced:         now,
		LastModifiedLocal:  now,
		LastModifiedJira:   now,
		IsDirty:            false,
		ConflictDetected:   false,
	}
	if err := repo.SaveTicketState(ctx, initial); err != nil {
		t.Fatalf("failed to save initial state: %v", err)
	}

	// Concurrent updates to same ticket
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			state := &repository.TicketSyncState{
				TicketKey:          "JMD-UPDATE",
				LastSynced:         now.Add(time.Duration(idx) * time.Second),
				LastModifiedLocal:  now.Add(time.Duration(idx) * time.Second),
				LastModifiedJira:   now.Add(time.Duration(idx) * time.Second),
				IsDirty:            idx%2 == 0,
				ConflictDetected:   false,
			}

			if err := repo.SaveTicketState(ctx, state); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent update error: %v", err)
	}

	// Verify final state (one of the updates should have won)
	final, err := repo.GetTicketState(ctx, "JMD-UPDATE")
	if err != nil {
		t.Fatalf("failed to get final state: %v", err)
	}

	if final.TicketKey != "JMD-UPDATE" {
		t.Errorf("expected ticket key JMD-UPDATE, got %s", final.TicketKey)
	}
}

func TestIntegration_DatabaseHealth(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	db, err := NewDatabase(config, nil)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Check health
	if err := db.Health(ctx); err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	// Check stats
	stats := db.Stats()
	if stats.OpenConnections == 0 {
		t.Error("expected at least one open connection")
	}
}

func TestIntegration_MigrationIdempotence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	config := DatabaseConfig{
		Path:         dbPath,
		MaxOpenConns: 1,
		BusyTimeout:  5 * time.Second,
	}

	// First migration
	{
		db, err := NewDatabase(config, nil)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}

		ctx := context.Background()
		if err := db.Migrate(ctx); err != nil {
			t.Fatalf("first migration failed: %v", err)
		}

		db.Close()
	}

	// Second migration (should be no-op)
	{
		db, err := NewDatabase(config, nil)
		if err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer db.Close()

		ctx := context.Background()
		if err := db.Migrate(ctx); err != nil {
			t.Fatalf("second migration failed: %v", err)
		}

		// Verify schema version is still 1
		var version int
		err = db.DB().QueryRowContext(ctx, "SELECT MAX(version) FROM schema_version").Scan(&version)
		if err != nil {
			t.Fatalf("failed to query version: %v", err)
		}

		if version != 1 {
			t.Errorf("expected version 1, got %d", version)
		}
	}
}
