// Package domain contains the core business logic and entities for jiramd.
//
// # Domain-Driven Design (DDD) Principles
//
// This package follows strict DDD principles and has ZERO external dependencies
// (only Go standard library is allowed). The domain layer represents the heart
// of the application's business logic and defines the ubiquitous language.
//
// # Architecture Rules
//
//   - NO imports from application or infrastructure layers
//   - NO external dependencies (only stdlib: time, strings, fmt, regexp, crypto/md5, encoding/hex, errors)
//   - All domain logic is self-contained and testable in isolation
//   - Entities and value objects are immutable where appropriate
//   - All timestamps are stored in UTC
//
// # Core Concepts
//
// ## Entities
//
// Entities have identity and lifecycle. Two entities are equal if they have
// the same identity, even if their attributes differ.
//
//   - Ticket: A Jira ticket (aggregate root), identified by TicketKey
//   - Comment: A comment on a ticket, identified by ID
//   - Project: A Jira project, identified by project key
//   - SyncState: Sync state for a project, identified by project key
//   - TicketState: Sync state for a ticket, identified by (project key, ticket key)
//   - PendingOperation: A queued sync operation, identified by ID
//
// ## Value Objects
//
// Value objects have no identity. Two value objects are equal if all their
// attributes are equal. Value objects should be immutable.
//
//   - TicketKey: A validated Jira ticket identifier (e.g., "JMD-123")
//   - SyncTimestamp: A UTC timestamp for sync operations
//   - FieldValue: A field's value with type-safe access
//   - CustomField: Configuration for a custom field
//   - DerivedField: A field computed from other fields
//   - SyncResult: Result of a sync operation
//
// ## Aggregates
//
// Ticket is an aggregate root that owns Comments and CustomFields.
// External references should only hold the TicketKey, not the full Ticket.
//
// # Ubiquitous Language
//
// The following terms define the domain language and must be used consistently
// throughout the codebase:
//
//   - Ticket: A Jira work item (not "Issue")
//   - Pull: Synchronize from Jira to local (Jira → Local)
//   - Push: Synchronize from local to Jira (Local → Jira)
//   - Sync: Bidirectional synchronization operation
//   - Conflict: Both local and remote modified since last sync
//   - Staging: A new comment file pending sync to Jira
//   - Custom Field: User-defined field configuration
//   - Derived Field: Field computed from other fields using DSL
//   - Bidirectional: Syncs both directions (Jira ↔ Local)
//   - Local-Only: Never synced to Jira
//   - Content Hash: MD5 hash for conflict detection
//
// # Domain Errors
//
// Domain errors represent business rule violations. Application layer should
// check for these specific errors:
//
//   - ErrNotFound: Entity not found
//   - ErrInvalidInput: Invalid input data
//   - ErrInvalidTicketKey: Ticket key format invalid
//   - ErrInvalidFieldValue: Field value not in valid values
//   - ErrSyncConflict: Sync conflict detected
//   - ErrInvalidTimestamp: Invalid or zero timestamp
//   - ErrEmptyKey: Empty or whitespace-only key
//   - ErrInvalidOperation: Invalid operation type
//
// # Invariants
//
// The domain layer enforces these invariants:
//
//   - TicketKey must match format: ^[A-Z][A-Z0-9]+-\d+$
//   - Project key must match format: ^[A-Z][A-Z0-9]{1,9}$
//   - All timestamps must be in UTC
//   - Ticket must have Key, Summary, Created, and Updated
//   - Comment must have ID, TicketKey, Author, Created, and Updated
//   - CustomField must have Name, DisplayName, Source, and SyncDirection
//   - PendingOperation attempts must be <= max retries (3)
//
// # Usage Example
//
//	// Create a ticket key (value object)
//	key, err := domain.NewTicketKey("JMD-123")
//	if err != nil {
//	    // Handle invalid format
//	}
//
//	// Create a ticket (entity)
//	ticket := domain.NewTicket(key, "Implement feature", time.Now(), time.Now())
//	ticket.Status = "In Progress"
//	ticket.Priority = "High"
//
//	// Compute content hash for conflict detection
//	hash := ticket.ContentHash()
//
//	// Validate ticket
//	if err := ticket.Validate(); err != nil {
//	    // Handle validation error
//	}
//
//	// Create a custom field (value object)
//	field, err := domain.NewCustomField("dev_assignment", "Dev Assignment", "labels", domain.SyncBidirectional)
//	if err != nil {
//	    // Handle invalid field
//	}
//
//	// Add to project
//	project, _ := domain.NewProject("JMD", "Jira Markdown Daemon")
//	project.AddCustomField(field)
package domain
