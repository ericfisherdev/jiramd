// Package sqlite provides SQLite-based implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
)

// Migration represents a database schema migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

var (
	//go:embed migrations/001_initial_schema.sql
	migration001 string
)

// migrations contains all available migrations in order.
var migrations = []Migration{
	{
		Version: 1,
		Name:    "initial_schema",
		SQL:     migration001,
	},
}

// MigrationManager handles database schema migrations.
type MigrationManager struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewMigrationManager creates a new migration manager.
func NewMigrationManager(db *sql.DB, logger *slog.Logger) *MigrationManager {
	if logger == nil {
		logger = slog.Default()
	}
	return &MigrationManager{
		db:     db,
		logger: logger,
	}
}

// Migrate applies all pending migrations.
// Migrations are applied in a transaction and rolled back on error.
// Returns the current schema version after migration.
func (m *MigrationManager) Migrate(ctx context.Context) (int, error) {
	m.logger.Info("starting database migrations")

	// Get current schema version
	currentVersion, err := m.getCurrentVersion(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get current schema version: %w", err)
	}

	m.logger.Info("current schema version", "version", currentVersion)

	// Apply pending migrations
	appliedCount := 0
	for _, migration := range migrations {
		if migration.Version <= currentVersion {
			m.logger.Debug("skipping applied migration",
				"version", migration.Version,
				"name", migration.Name)
			continue
		}

		m.logger.Info("applying migration",
			"version", migration.Version,
			"name", migration.Name)

		if err := m.applyMigration(ctx, migration); err != nil {
			return currentVersion, fmt.Errorf("failed to apply migration %d (%s): %w",
				migration.Version, migration.Name, err)
		}

		appliedCount++
		currentVersion = migration.Version
	}

	if appliedCount > 0 {
		m.logger.Info("migrations completed",
			"applied_count", appliedCount,
			"current_version", currentVersion)
	} else {
		m.logger.Info("no pending migrations")
	}

	return currentVersion, nil
}

// getCurrentVersion returns the current schema version.
// Returns 0 if no migrations have been applied yet.
func (m *MigrationManager) getCurrentVersion(ctx context.Context) (int, error) {
	// Check if schema_version table exists
	var tableExists bool
	err := m.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM sqlite_master
			WHERE type = 'table' AND name = 'schema_version'
		)
	`).Scan(&tableExists)
	if err != nil {
		return 0, fmt.Errorf("failed to check schema_version table: %w", err)
	}

	if !tableExists {
		return 0, nil
	}

	// Get latest version
	var version int
	err = m.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0)
		FROM schema_version
	`).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to query schema version: %w", err)
	}

	return version, nil
}

// applyMigration applies a single migration within a transaction.
func (m *MigrationManager) applyMigration(ctx context.Context, migration Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Execute migration SQL
	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}

// Reset drops all tables and reapplies all migrations.
// WARNING: This will delete all data. Use only for testing.
func (m *MigrationManager) Reset(ctx context.Context) error {
	m.logger.Warn("resetting database - all data will be lost")

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get all table names
	rows, err := tx.QueryContext(ctx, `
		SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to iterate tables: %w", err)
	}

	// Drop all tables
	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reset: %w", err)
	}

	// Reapply migrations
	_, err = m.Migrate(ctx)
	return err
}
