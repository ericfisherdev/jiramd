// Package markdown provides markdown file parsing and generation.
// This infrastructure layer handles conversion between markdown files and domain entities.
package markdown

import (
	"context"

	"github.com/esfisher/jiramd/internal/domain"
)

// Parser handles parsing markdown files into domain entities.
type Parser struct {
	// TODO: Add template engine and configuration
}

// NewParser creates a new markdown parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseTicket parses a markdown file into a Ticket entity.
// This is a placeholder for the actual implementation.
func (p *Parser) ParseTicket(ctx context.Context, content []byte) (*domain.Ticket, error) {
	// TODO: Implement markdown parsing logic
	return nil, nil
}

// GenerateTicket generates a markdown file from a Ticket entity.
// This is a placeholder for the actual implementation.
func (p *Parser) GenerateTicket(ctx context.Context, ticket *domain.Ticket) ([]byte, error) {
	// TODO: Implement markdown generation logic
	return nil, nil
}
