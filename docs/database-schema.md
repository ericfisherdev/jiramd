# Database Schema

This document describes the SQLite database schema used by jiramd for state persistence.

## Overview

The database stores synchronization state for tickets and projects, enabling:
- Conflict detection via timestamp comparison
- Dirty tracking for efficient push operations
- Project-level sync metadata for incremental updates
- Transaction support for atomic state updates

## Database File

- **Location**: Configured via `JIRAMD_DB_PATH` or `~/.jiramd/state.db` by default
- **Format**: SQLite 3.x
- **Permissions**: 0600 (owner read/write only)
- **Connection**: Single connection with WAL mode for concurrent reads

## Tables

### schema_version

Tracks database schema version for migrations.

```sql
CREATE TABLE schema_version (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

- **version**: Schema version number (monotonically increasing)
- **applied_at**: When this migration was applied

### ticket_sync_state

Stores synchronization state for individual tickets.

```sql
CREATE TABLE ticket_sync_state (
    ticket_key TEXT PRIMARY KEY,
    last_synced TIMESTAMP NOT NULL,
    last_modified_local TIMESTAMP NOT NULL,
    last_modified_jira TIMESTAMP NOT NULL,
    is_dirty BOOLEAN NOT NULL DEFAULT 0,
    conflict_detected BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ticket_dirty ON ticket_sync_state(is_dirty) WHERE is_dirty = 1;
CREATE INDEX idx_ticket_conflict ON ticket_sync_state(conflict_detected) WHERE conflict_detected = 1;
CREATE INDEX idx_ticket_modified_local ON ticket_sync_state(last_modified_local);
```

**Columns:**
- **ticket_key**: Unique Jira ticket identifier (e.g., "JMD-123")
- **last_synced**: Timestamp of last successful bidirectional sync
- **last_modified_local**: Timestamp of last local markdown file modification
- **last_modified_jira**: Timestamp from Jira's updated field
- **is_dirty**: Flag indicating unsynced local changes (1 = dirty, 0 = clean)
- **conflict_detected**: Flag indicating both local and Jira modified since last sync
- **created_at**: Record creation timestamp
- **updated_at**: Record last update timestamp

**Indexes:**
- Partial index on `is_dirty` for efficient dirty ticket queries
- Partial index on `conflict_detected` for conflict resolution queries
- Index on `last_modified_local` for time-based queries

### project_sync_state

Stores synchronization state for projects.

```sql
CREATE TABLE project_sync_state (
    project_key TEXT PRIMARY KEY,
    last_full_sync TIMESTAMP,
    last_incremental_sync TIMESTAMP,
    ticket_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Columns:**
- **project_key**: Unique Jira project identifier (e.g., "JMD")
- **last_full_sync**: Timestamp of last complete project sync
- **last_incremental_sync**: Timestamp of last incremental sync (updates only)
- **ticket_count**: Total number of tickets tracked for this project
- **created_at**: Record creation timestamp
- **updated_at**: Record last update timestamp

## Schema Evolution

### Migration Strategy

1. **Version Tracking**: `schema_version` table tracks applied migrations
2. **Up Migrations**: Forward migrations apply schema changes
3. **Transaction Safety**: Each migration runs in a transaction
4. **Idempotency**: Migrations can be safely re-run

### Migration Files

Migrations are embedded in the binary and applied automatically on startup.

Location: `internal/infrastructure/sqlite/migrations/`

Naming: `001_initial_schema.sql`, `002_add_indexes.sql`, etc.

### Current Version

**Version 1**: Initial schema (ticket_sync_state, project_sync_state, schema_version)

## Timestamp Handling

All timestamps are stored in UTC using SQLite's TIMESTAMP type:
- Format: `YYYY-MM-DD HH:MM:SS.SSS` (RFC3339 compatible)
- Timezone: Always UTC (converted on read/write)
- Precision: Millisecond precision for conflict detection
- NULL Handling: Use `time.Time{}` zero value; check with `IsZero()` on read

## Transaction Isolation

- **Isolation Level**: SERIALIZABLE (SQLite default)
- **WAL Mode**: Enabled for concurrent reads during writes
- **Lock Timeout**: 5 seconds (configurable)
- **Transaction Context**: Uses context.Context for transaction tracking

## Constraints

- **Primary Keys**: All tables use natural keys (ticket_key, project_key)
- **NOT NULL**: Required fields use NOT NULL constraints
- **Defaults**: Timestamps and counters have sensible defaults
- **Foreign Keys**: Not used (domain layer enforces referential integrity)

## Performance Considerations

### Query Optimization

- Partial indexes on boolean flags reduce index size
- Covering indexes avoid table lookups for common queries
- ANALYZE runs periodically to update query planner statistics

### Connection Pooling

- Single connection (SQLite limitation)
- Connection reuse via sql.DB
- Prepared statements cached automatically

### Write Optimization

- WAL mode enables concurrent reads
- Batch updates in transactions
- PRAGMA synchronous = NORMAL for durability/performance balance

## Backup and Recovery

### Backup Strategy

- Online backup via SQLite backup API
- Scheduled backups configurable via daemon
- Backup retention policy (keep last N backups)

### Recovery

- Database corruption: Restore from backup
- Data loss: Full resync from Jira
- Conflict resolution: Manual intervention via CLI

## Security

- File permissions: 0600 (owner only)
- No sensitive data stored (API tokens in environment)
- SQL injection prevention: Prepared statements only
- Input validation: Domain layer validates all inputs

## Example Queries

### Get all dirty tickets

```sql
SELECT ticket_key, last_modified_local, last_synced
FROM ticket_sync_state
WHERE is_dirty = 1
ORDER BY last_modified_local DESC;
```

### Get conflicted tickets

```sql
SELECT ticket_key, last_modified_local, last_modified_jira, last_synced
FROM ticket_sync_state
WHERE conflict_detected = 1;
```

### Get project sync status

```sql
SELECT project_key, last_full_sync, last_incremental_sync, ticket_count
FROM project_sync_state
ORDER BY project_key;
```

### Get tickets modified since timestamp

```sql
SELECT ticket_key, last_modified_local
FROM ticket_sync_state
WHERE last_modified_local > ?
ORDER BY last_modified_local DESC;
```
