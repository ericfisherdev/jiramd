// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import "time"

// Ticket represents a Jira ticket entity.
// This is a core domain entity that encapsulates ticket state and behavior.
type Ticket struct {
	// Key is the unique Jira ticket identifier (e.g., "JMD-123")
	Key string

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

	// Created is when the ticket was created
	Created time.Time

	// Updated is when the ticket was last updated
	Updated time.Time

	// Fields contains custom field values
	Fields map[string]interface{}
}

// Comment represents a comment on a Jira ticket.
type Comment struct {
	// ID is the unique comment identifier
	ID string

	// TicketKey is the key of the ticket this comment belongs to
	TicketKey string

	// Author is the user who created the comment
	Author string

	// Body is the comment text content
	Body string

	// Created is when the comment was created
	Created time.Time

	// Updated is when the comment was last updated
	Updated time.Time
}

// Project represents a Jira project entity.
type Project struct {
	// Key is the unique project key (e.g., "JMD")
	Key string

	// Name is the human-readable project name
	Name string

	// Description is the project description
	Description string
}
