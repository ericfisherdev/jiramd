// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ticketKeyPattern defines the valid format for Jira ticket keys (PROJECT-123)
// Project key must be 2-10 characters (matching projectKeyPattern constraints)
var ticketKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}-\d+$`)

// TicketKey is a value object representing a valid Jira ticket identifier.
// It enforces the format: PROJECT-NUMBER (e.g., "JMD-123").
// TicketKey is immutable and comparable by value.
type TicketKey struct {
	value string
}

// NewTicketKey creates a new TicketKey after validating the format.
// Returns ErrInvalidTicketKey if the key doesn't match the expected pattern.
func NewTicketKey(key string) (TicketKey, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return TicketKey{}, fmt.Errorf("%w: empty ticket key", ErrEmptyKey)
	}

	if !ticketKeyPattern.MatchString(key) {
		return TicketKey{}, fmt.Errorf("%w: %s (expected format: PROJECT-123)", ErrInvalidTicketKey, key)
	}

	return TicketKey{value: key}, nil
}

// String returns the string representation of the ticket key.
func (tk TicketKey) String() string {
	return tk.value
}

// ProjectKey extracts the project portion of the ticket key (e.g., "JMD" from "JMD-123").
func (tk TicketKey) ProjectKey() string {
	parts := strings.Split(tk.value, "-")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// IsZero returns true if this is the zero value (empty ticket key).
func (tk TicketKey) IsZero() bool {
	return tk.value == ""
}

// Ticket represents a Jira ticket entity.
// This is a core domain entity (aggregate root) that encapsulates ticket state and behavior.
// Ticket has identity defined by its TicketKey and maintains its lifecycle.
type Ticket struct {
	// Key is the unique Jira ticket identifier (immutable)
	Key TicketKey

	// Summary is the ticket title/summary
	Summary string

	// Description is the detailed ticket description
	Description string

	// Status represents the current ticket status (e.g., "To Do", "In Progress", "Done")
	Status string

	// IssueType is the type of ticket (e.g., "Story", "Bug", "Task")
	IssueType string

	// Priority is the ticket priority (e.g., "High", "Medium", "Low")
	Priority string

	// Assignee is the user assigned to the ticket
	Assignee string

	// Reporter is the user who created the ticket
	Reporter string

	// Labels contains ticket labels
	Labels []string

	// Created is when the ticket was created (immutable, always UTC)
	Created time.Time

	// Updated is when the ticket was last updated in Jira (always UTC)
	Updated time.Time

	// CustomFields contains custom field values (flexible storage for extension)
	CustomFields map[string]FieldValue
}

// NewTicket creates a new Ticket with required fields.
// All timestamps are normalized to UTC.
func NewTicket(key TicketKey, summary string, created, updated time.Time) *Ticket {
	return &Ticket{
		Key:          key,
		Summary:      summary,
		Created:      created.UTC(),
		Updated:      updated.UTC(),
		Labels:       make([]string, 0),
		CustomFields: make(map[string]FieldValue),
	}
}

// ContentHash computes an MD5 hash of the ticket content for conflict detection.
// This includes all mutable fields that can be modified locally.
func (t *Ticket) ContentHash() string {
	h := md5.New()
	// Include all fields that can be modified
	fmt.Fprintf(h, "summary:%s\n", t.Summary)
	fmt.Fprintf(h, "description:%s\n", t.Description)
	fmt.Fprintf(h, "status:%s\n", t.Status)
	fmt.Fprintf(h, "priority:%s\n", t.Priority)
	fmt.Fprintf(h, "assignee:%s\n", t.Assignee)
	fmt.Fprintf(h, "labels:%s\n", strings.Join(t.Labels, ","))

	// Sort custom field keys for deterministic hash
	keys := make([]string, 0, len(t.CustomFields))
	for k := range t.CustomFields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Include custom fields in sorted order for deterministic hash
	for _, k := range keys {
		v := t.CustomFields[k]
		fmt.Fprintf(h, "custom:%s=%v\n", k, v.Raw())
	}

	return hex.EncodeToString(h.Sum(nil))
}

// Validate checks if the ticket has all required fields populated.
func (t *Ticket) Validate() error {
	if t.Key.IsZero() {
		return fmt.Errorf("%w: ticket key is required", ErrInvalidInput)
	}
	if strings.TrimSpace(t.Summary) == "" {
		return fmt.Errorf("%w: ticket summary is required", ErrInvalidInput)
	}
	if t.Created.IsZero() {
		return fmt.Errorf("%w: created timestamp is required", ErrInvalidTimestamp)
	}
	if t.Updated.IsZero() {
		return fmt.Errorf("%w: updated timestamp is required", ErrInvalidTimestamp)
	}
	return nil
}
