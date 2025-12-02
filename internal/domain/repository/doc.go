// Package repository defines interfaces for data access in the domain layer.
//
// # Repository Pattern
//
// This package follows the Repository pattern from Domain-Driven Design (DDD).
// Repository interfaces define contracts for data access without specifying
// implementation details. This allows the domain layer to remain independent
// of infrastructure concerns.
//
// # Architecture Principles
//
// The repository interfaces in this package adhere to Clean Architecture principles:
//
//  1. Domain layer has ZERO dependencies on application or infrastructure layers
//  2. Interfaces are defined by the domain, implemented by infrastructure
//  3. Dependency direction: infrastructure -> domain (never the reverse)
//  4. All methods use domain entities and value objects, never infrastructure types
//
// # Repository Interfaces
//
// This package defines three primary repository interfaces:
//
// ## JiraRepository
//
// Abstracts communication with Jira Cloud REST API. Implementations handle:
//   - Authentication and authorization
//   - HTTP client management and retries
//   - Pagination of large result sets
//   - Rate limiting
//   - Mapping between Jira API responses and domain entities
//
// ## MarkdownRepository
//
// Abstracts reading and writing of ticket markdown files. Implementations handle:
//   - File I/O operations
//   - YAML frontmatter parsing
//   - Markdown template processing
//   - Directory structure management
//
// ## StateRepository
//
// Abstracts persistence of synchronization state. Implementations handle:
//   - Tracking sync timestamps for conflict detection
//   - Recording local modifications
//   - Managing project metadata
//   - Transaction support for atomic updates
//
// ## Legacy Interfaces (ticket.go)
//
// The TicketRepository, CommentRepository, and ProjectRepository interfaces
// provide basic CRUD operations for entity persistence. These are used by
// infrastructure implementations and may be refactored or removed in future
// iterations in favor of the more specialized interfaces above.
//
// # Error Handling
//
// All repository methods should return domain errors defined in domain/errors.go:
//   - ErrNotFound: Entity not found
//   - ErrInvalidInput: Invalid input data
//   - ErrConflict: Concurrent modification conflict
//   - ErrUnauthorized: Authentication or authorization failure
//
// Infrastructure implementations must map their specific errors (HTTP status codes,
// database errors, file system errors) to these domain errors.
//
// # Usage Example
//
//	type SyncService struct {
//		jira     repository.JiraRepository
//		markdown repository.MarkdownRepository
//		state    repository.StateRepository
//	}
//
//	func (s *SyncService) PullTicket(ctx context.Context, key string) error {
//		// Fetch from Jira
//		ticket, err := s.jira.FetchTicket(ctx, key)
//		if err != nil {
//			return fmt.Errorf("failed to fetch ticket: %w", err)
//		}
//
//		// Write to markdown
//		filePath := filepath.Join("tickets", key+".md")
//		if err := s.markdown.WriteTicket(ctx, filePath, ticket); err != nil {
//			return fmt.Errorf("failed to write markdown: %w", err)
//		}
//
//		// Update sync state
//		state := &repository.TicketSyncState{
//			TicketKey:         key,
//			LastSynced:        time.Now(),
//			LastModifiedJira:  ticket.Updated,
//			LastModifiedLocal: time.Now(),
//			IsDirty:           false,
//		}
//		return s.state.SaveTicketState(ctx, state)
//	}
package repository
