// Package sqlite provides SQLite-based implementations of repository interfaces.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // SQLite driver
)

// DatabaseConfig holds configuration for SQLite database.
type DatabaseConfig struct {
	// Path to the SQLite database file
	Path string

	// MaxOpenConns is the maximum number of open connections (SQLite only supports 1)
	MaxOpenConns int

	// ConnMaxLifetime is the maximum lifetime of a connection
	ConnMaxLifetime time.Duration

	// BusyTimeout is how long to wait for a lock before returning SQLITE_BUSY
	BusyTimeout time.Duration
}

// DefaultConfig returns the default database configuration.
func DefaultConfig() DatabaseConfig {
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, ".jiramd", "state.db")

	return DatabaseConfig{
		Path:            defaultPath,
		MaxOpenConns:    1, // SQLite only supports single writer
		ConnMaxLifetime: 0, // No max lifetime
		BusyTimeout:     5 * time.Second,
	}
}

// Database wraps sql.DB with jiramd-specific functionality.
type Database struct {
	db     *sql.DB
	config DatabaseConfig
	logger *slog.Logger
}

// NewDatabase creates a new database connection with the given configuration.
// It ensures the database directory exists and applies migrations.
func NewDatabase(config DatabaseConfig, logger *slog.Logger) (*Database, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(config.Path)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Build connection string with pragmas
	connStr := fmt.Sprintf("file:%s?_pragma=busy_timeout(%d)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)",
		config.Path,
		int(config.BusyTimeout.Milliseconds()),
	)

	// Open database connection
	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxOpenConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{
		db:     db,
		config: config,
		logger: logger,
	}

	logger.Info("database connection established",
		"path", config.Path,
		"busy_timeout", config.BusyTimeout)

	return database, nil
}

// Migrate applies all pending database migrations.
func (d *Database) Migrate(ctx context.Context) error {
	migrator := NewMigrationManager(d.db, d.logger)
	version, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	d.logger.Info("database migrations complete", "version", version)
	return nil
}

// DB returns the underlying sql.DB.
func (d *Database) DB() *sql.DB {
	return d.db
}

// Close closes the database connection.
func (d *Database) Close() error {
	if d.db != nil {
		d.logger.Info("closing database connection")
		return d.db.Close()
	}
	return nil
}

// Health checks the database health.
func (d *Database) Health(ctx context.Context) error {
	// Ping database
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Query a simple statement
	var result int
	if err := d.db.QueryRowContext(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

// Stats returns database statistics.
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// SetFilePermissions sets the database file permissions to 0600 (owner read/write only).
func (d *Database) SetFilePermissions() error {
	if err := os.Chmod(d.config.Path, 0600); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}
	d.logger.Debug("set database file permissions to 0600")
	return nil
}
