// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"fmt"
	"strings"
	"time"
)

// Comment represents a comment on a Jira ticket.
// Comment is an entity owned by the Ticket aggregate.
type Comment struct {
	// ID is the unique comment identifier from Jira
	ID string

	// TicketKey is the key of the ticket this comment belongs to
	TicketKey TicketKey

	// Author is the user who created the comment (email or username)
	Author string

	// Body is the comment text content
	Body string

	// Created is when the comment was created (immutable, always UTC)
	Created time.Time

	// Updated is when the comment was last updated (always UTC)
	Updated time.Time
}

// NewComment creates a new Comment with required fields.
// All timestamps are normalized to UTC.
func NewComment(id string, ticketKey TicketKey, author, body string, created, updated time.Time) (*Comment, error) {
	c := &Comment{
		ID:        strings.TrimSpace(id),
		TicketKey: ticketKey,
		Author:    strings.TrimSpace(author),
		Body:      body,
		Created:   created.UTC(),
		Updated:   updated.UTC(),
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate checks if the comment has all required fields populated.
func (c *Comment) Validate() error {
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("%w: comment ID is required", ErrInvalidInput)
	}
	if c.TicketKey.IsZero() {
		return fmt.Errorf("%w: ticket key is required", ErrInvalidInput)
	}
	if strings.TrimSpace(c.Author) == "" {
		return fmt.Errorf("%w: comment author is required", ErrInvalidInput)
	}
	if c.Created.IsZero() {
		return fmt.Errorf("%w: created timestamp is required", ErrInvalidTimestamp)
	}
	if c.Updated.IsZero() {
		return fmt.Errorf("%w: updated timestamp is required", ErrInvalidTimestamp)
	}
	return nil
}
