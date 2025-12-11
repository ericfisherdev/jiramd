// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/esfisher/jiramd/internal/domain"
)

// mapIssueToTicket converts a Jira API Issue to a domain Ticket entity.
// This function extracts the relevant fields from the Jira API response
// and maps them to our domain model.
func mapIssueToTicket(issue *jira.Issue) (*domain.Ticket, error) {
	if issue == nil {
		return nil, fmt.Errorf("issue cannot be nil")
	}

	// Create ticket key value object
	ticketKey, err := domain.NewTicketKey(issue.Key)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket key '%s': %w", issue.Key, err)
	}

	// Convert Jira Time to time.Time
	created := time.Time(issue.Fields.Created)
	updated := time.Time(issue.Fields.Updated)

	// Create the ticket
	ticket := domain.NewTicket(ticketKey, issue.Fields.Summary, created, updated)

	// Map optional fields
	if issue.Fields.Description != "" {
		ticket.Description = issue.Fields.Description
	}

	if issue.Fields.Status != nil {
		ticket.Status = issue.Fields.Status.Name
	}

	// Type is not a pointer in the new version
	if issue.Fields.Type.Name != "" {
		ticket.IssueType = issue.Fields.Type.Name
	}

	if issue.Fields.Priority != nil {
		ticket.Priority = issue.Fields.Priority.Name
	}

	if issue.Fields.Assignee != nil {
		ticket.Assignee = issue.Fields.Assignee.DisplayName
	}

	if issue.Fields.Reporter != nil {
		ticket.Reporter = issue.Fields.Reporter.DisplayName
	}

	// Map labels
	if issue.Fields.Labels != nil {
		ticket.Labels = issue.Fields.Labels
	} else {
		ticket.Labels = make([]string, 0)
	}

	// Initialize custom fields map
	ticket.CustomFields = make(map[string]domain.FieldValue)

	return ticket, nil
}

// mapCommentToComment converts a Jira API Comment to a domain Comment entity.
func mapCommentToComment(jiraComment *jira.Comment, ticketKey domain.TicketKey) (*domain.Comment, error) {
	if jiraComment == nil {
		return nil, fmt.Errorf("comment cannot be nil")
	}

	// Parse timestamps (Comment uses string timestamps)
	created, err := parseCommentTime(jiraComment.Created)
	if err != nil {
		return nil, fmt.Errorf("invalid comment created timestamp: %w", err)
	}

	updated, err := parseCommentTime(jiraComment.Updated)
	if err != nil {
		return nil, fmt.Errorf("invalid comment updated timestamp: %w", err)
	}

	// Extract author name
	author := ""
	if jiraComment.Author != nil {
		author = jiraComment.Author.DisplayName
	}

	// Create the comment
	comment, err := domain.NewComment(
		jiraComment.ID,
		ticketKey,
		author,
		jiraComment.Body,
		created,
		updated,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}

	return comment, nil
}

// parseCommentTime parses a Jira API comment timestamp string to time.Time.
// Comment timestamps are strings in ISO 8601 format.
func parseCommentTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	// Jira uses multiple possible timestamp formats
	formats := []string{
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000Z0700",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var lastErr error
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t.UTC(), nil
		}
		lastErr = err
	}

	return time.Time{}, fmt.Errorf("failed to parse time '%s': %w", timeStr, lastErr)
}

// mapProjectToProject converts a Jira API Project to a domain Project entity.
func mapProjectToProject(jiraProject *jira.Project) (*domain.Project, error) {
	if jiraProject == nil {
		return nil, fmt.Errorf("project cannot be nil")
	}

	// Create the project
	project, err := domain.NewProject(jiraProject.Key, jiraProject.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	// Map optional description
	if jiraProject.Description != "" {
		project.Description = jiraProject.Description
	}

	return project, nil
}


// mapCommentToCommentCreate converts a domain Comment to Jira API comment creation structure.
func mapCommentToCommentCreate(comment *domain.Comment) *jira.Comment {
	return &jira.Comment{
		Body: comment.Body,
	}
}
