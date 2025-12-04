-- Migration 001: Initial schema
-- Creates tables for ticket and project sync state

-- Schema version tracking
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Ticket synchronization state
CREATE TABLE IF NOT EXISTS ticket_sync_state (
    ticket_key TEXT PRIMARY KEY,
    last_synced TIMESTAMP NOT NULL,
    last_modified_local TIMESTAMP NOT NULL,
    last_modified_jira TIMESTAMP NOT NULL,
    is_dirty BOOLEAN NOT NULL DEFAULT 0,
    conflict_detected BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_ticket_dirty
    ON ticket_sync_state(is_dirty)
    WHERE is_dirty = 1;

CREATE INDEX IF NOT EXISTS idx_ticket_conflict
    ON ticket_sync_state(conflict_detected)
    WHERE conflict_detected = 1;

CREATE INDEX IF NOT EXISTS idx_ticket_modified_local
    ON ticket_sync_state(last_modified_local);

-- Project synchronization state
CREATE TABLE IF NOT EXISTS project_sync_state (
    project_key TEXT PRIMARY KEY,
    last_full_sync TIMESTAMP,
    last_incremental_sync TIMESTAMP,
    ticket_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Record migration application
INSERT INTO schema_version (version) VALUES (1);
