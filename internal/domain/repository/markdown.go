// Package repository defines interfaces for data access.
// These interfaces are part of the domain layer and define contracts
// that infrastructure implementations must fulfill.
package repository

import (
	"context"

	"github.com/esfisher/jiramd/internal/domain"
)

// MarkdownRepository defines the interface for markdown file operations.
// This interface abstracts reading, writing, and parsing of ticket markdown files.
//
// Implementations must:
//   - Parse YAML frontmatter containing ticket metadata
//   - Extract ticket fields from markdown content
//   - Generate markdown files from domain entities using templates
//   - Handle file I/O errors appropriately
//   - Preserve markdown formatting and user customizations
//
// Domain errors that methods should return:
//   - ErrNotFound: when a markdown file doesn't exist
//   - ErrInvalidInput: when markdown content is malformed or unparseable
type MarkdownRepository interface {
	// ReadTicket reads and parses a markdown file into a Ticket entity.
	// Parses YAML frontmatter for metadata and extracts description from markdown body.
	// Returns ErrNotFound if the file doesn't exist.
	// Returns ErrInvalidInput if the markdown is malformed or missing required fields.
	ReadTicket(ctx context.Context, filePath string) (*domain.Ticket, error)

	// WriteTicket generates and writes a Ticket entity to a markdown file.
	// Uses the configured template to generate the markdown content.
	// Creates parent directories if they don't exist.
	// Returns ErrInvalidInput if the ticket data is invalid.
	WriteTicket(ctx context.Context, filePath string, ticket *domain.Ticket) error

	// ReadComments reads comments from a ticket's markdown file.
	// Comments are typically stored in a dedicated section of the ticket markdown.
	// Returns empty slice if the file has no comments.
	// Returns ErrNotFound if the file doesn't exist.
	ReadComments(ctx context.Context, filePath string) ([]*domain.Comment, error)

	// WriteComments updates the comments section of a ticket's markdown file.
	// Preserves the rest of the markdown content.
	// Returns ErrNotFound if the file doesn't exist.
	WriteComments(ctx context.Context, filePath string, comments []*domain.Comment) error

	// ListTicketFiles returns all markdown files in the configured tickets directory.
	// Files are identified by .md extension and proper frontmatter structure.
	// Returns empty slice if no ticket files exist.
	ListTicketFiles(ctx context.Context, directory string) ([]string, error)

	// GenerateIndex creates an index.md file with a summary of all tickets.
	// Uses the configured index template.
	// Returns ErrInvalidInput if the tickets data is invalid.
	GenerateIndex(ctx context.Context, indexPath string, tickets []*domain.Ticket) error

	// ValidateTemplate validates a markdown template file syntax.
	// Templates use Go's text/template syntax.
	// Returns ErrNotFound if the template file doesn't exist.
	// Returns ErrInvalidInput if the template syntax is invalid.
	// Note: Implementations should cache the parsed template internally for later use.
	ValidateTemplate(ctx context.Context, templatePath string) error
}
