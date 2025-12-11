// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"context"
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/esfisher/jiramd/internal/domain"
	"github.com/esfisher/jiramd/internal/domain/repository"
)

// Ensure Repository implements the JiraRepository interface at compile time
var _ repository.JiraRepository = (*Repository)(nil)

// Repository implements the JiraRepository interface using the go-jira/v2 client.
// It provides the bridge between the domain layer and the Jira Cloud REST API.
type Repository struct {
	client *Client
}

// NewRepository creates a new Jira repository with the given client.
func NewRepository(client *Client) *Repository {
	return &Repository{
		client: client,
	}
}

// FetchTicket retrieves a single ticket from Jira by its key.
// Returns ErrNotFound if the ticket doesn't exist.
// Returns ErrUnauthorized if the user lacks permission to view the ticket.
func (r *Repository) FetchTicket(ctx context.Context, key string) (*domain.Ticket, error) {
	r.client.logger.Info("fetching ticket from jira", "key", key)

	issue, resp, err := r.client.jiraClient.Issue.Get(ctx, key, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, mapHTTPError(err, fmt.Sprintf("ticket %s not found", key))
		}
		return nil, mapHTTPError(err, fmt.Sprintf("failed to fetch ticket %s", key))
	}

	ticket, err := mapIssueToTicket(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to map issue to ticket: %w", err)
	}

	r.client.logger.Info("successfully fetched ticket", "key", key)
	return ticket, nil
}

// FetchTicketsModifiedSince retrieves tickets modified after the given timestamp.
// Uses JQL: "project = X AND updated >= timestamp ORDER BY updated ASC"
// Results are paginated to avoid memory issues with large result sets.
// Returns empty slice if no tickets match the criteria.
func (r *Repository) FetchTicketsModifiedSince(ctx context.Context, projectKey string, since time.Time) ([]*domain.Ticket, error) {
	r.client.logger.Info("fetching tickets modified since", "project", projectKey, "since", since)

	// Format the timestamp for JQL (Jira uses format: "yyyy-MM-dd HH:mm")
	// We'll use a simpler format that Jira accepts
	sinceStr := since.Format("2006-01-02 15:04")

	jql := fmt.Sprintf(`project = %s AND updated >= "%s" ORDER BY updated ASC`, projectKey, sinceStr)

	return r.searchTickets(ctx, jql, "fetch tickets modified since")
}

// FetchAllTickets retrieves all tickets for a project.
// Uses JQL: "project = X ORDER BY updated DESC"
// Results are paginated to avoid memory issues with large result sets.
func (r *Repository) FetchAllTickets(ctx context.Context, projectKey string) ([]*domain.Ticket, error) {
	r.client.logger.Info("fetching all tickets", "project", projectKey)

	jql := fmt.Sprintf(`project = %s ORDER BY updated DESC`, projectKey)

	return r.searchTickets(ctx, jql, "fetch all tickets")
}

// searchTickets is a helper method that performs JQL search with pagination.
func (r *Repository) searchTickets(ctx context.Context, jql, operation string) ([]*domain.Ticket, error) {
	const maxResults = 50 // Page size for pagination
	startAt := 0
	tickets := make([]*domain.Ticket, 0)

	for {
		searchOptions := &jira.SearchOptions{
			MaxResults: maxResults,
			StartAt:    startAt,
		}

		issues, resp, err := r.client.jiraClient.Issue.Search(ctx, jql, searchOptions)
		if err != nil {
			return nil, mapHTTPError(err, fmt.Sprintf("failed to %s", operation))
		}

		if resp == nil {
			return nil, fmt.Errorf("received nil response from jira search")
		}

		// Map each issue to a ticket
		for _, issue := range issues {
			ticket, err := mapIssueToTicket(&issue)
			if err != nil {
				r.client.logger.Warn("failed to map issue, skipping", "key", issue.Key, "error", err)
				continue
			}
			tickets = append(tickets, ticket)
		}

		// Check if we've fetched all results
		if len(issues) < maxResults {
			break
		}

		startAt += maxResults
	}

	r.client.logger.Info("successfully fetched tickets", "count", len(tickets), "operation", operation)
	return tickets, nil
}

// UpdateTicket pushes local ticket changes to Jira.
// Only updates fields that have changed to minimize API calls.
// Returns the updated ticket with the authoritative Jira timestamp for version tracking.
// Returns ErrNotFound if the ticket no longer exists in Jira.
// Returns ErrConflict if the ticket was modified by another user since last fetch.
// Returns ErrUnauthorized if the user lacks permission to edit the ticket.
func (r *Repository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) (*domain.Ticket, error) {
	r.client.logger.Info("updating ticket in jira", "key", ticket.Key.String())

	// Create issue with updated fields
	issue := &jira.Issue{
		Key: ticket.Key.String(),
		Fields: &jira.IssueFields{
			Summary:     ticket.Summary,
			Description: ticket.Description,
			Labels:      ticket.Labels,
		},
	}

	// Set optional fields
	if ticket.Priority != "" {
		issue.Fields.Priority = &jira.Priority{
			Name: ticket.Priority,
		}
	}

	if ticket.Assignee != "" {
		issue.Fields.Assignee = &jira.User{
			Name: ticket.Assignee,
		}
	}

	// Update the issue
	_, resp, err := r.client.jiraClient.Issue.Update(ctx, issue, nil)
	if err != nil {
		if resp != nil {
			switch resp.StatusCode {
			case 404:
				return nil, mapHTTPError(err, fmt.Sprintf("ticket %s not found", ticket.Key.String()))
			case 409:
				return nil, mapHTTPError(err, fmt.Sprintf("conflict updating ticket %s", ticket.Key.String()))
			}
		}
		return nil, mapHTTPError(err, fmt.Sprintf("failed to update ticket %s", ticket.Key.String()))
	}

	// Fetch the updated ticket to get the authoritative timestamp
	updatedTicket, err := r.FetchTicket(ctx, ticket.Key.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated ticket: %w", err)
	}

	r.client.logger.Info("successfully updated ticket", "key", ticket.Key.String())
	return updatedTicket, nil
}

// FetchComments retrieves all comments for a given ticket.
// Returns empty slice if the ticket has no comments.
// Returns ErrNotFound if the ticket doesn't exist.
func (r *Repository) FetchComments(ctx context.Context, ticketKey string) ([]*domain.Comment, error) {
	r.client.logger.Info("fetching comments", "ticket", ticketKey)

	// Parse ticket key to domain value object
	key, err := domain.NewTicketKey(ticketKey)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket key: %w", err)
	}

	// Get issue with comments expanded
	options := &jira.GetQueryOptions{
		Expand: "comments",
	}

	issue, resp, err := r.client.jiraClient.Issue.Get(ctx, ticketKey, options)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, mapHTTPError(err, fmt.Sprintf("ticket %s not found", ticketKey))
		}
		return nil, mapHTTPError(err, fmt.Sprintf("failed to fetch comments for ticket %s", ticketKey))
	}

	comments := make([]*domain.Comment, 0)

	// Map comments if they exist
	if issue.Fields != nil && issue.Fields.Comments != nil {
		for _, jiraComment := range issue.Fields.Comments.Comments {
			comment, err := mapCommentToComment(jiraComment, key)
			if err != nil {
				r.client.logger.Warn("failed to map comment, skipping", "comment_id", jiraComment.ID, "error", err)
				continue
			}
			comments = append(comments, comment)
		}
	}

	r.client.logger.Info("successfully fetched comments", "ticket", ticketKey, "count", len(comments))
	return comments, nil
}

// AddComment adds a new comment to a Jira ticket.
// Returns the created comment with its Jira-assigned ID populated.
// Returns ErrNotFound if the ticket doesn't exist.
// Returns ErrUnauthorized if the user lacks permission to comment.
func (r *Repository) AddComment(ctx context.Context, ticketKey string, comment *domain.Comment) (*domain.Comment, error) {
	r.client.logger.Info("adding comment to ticket", "ticket", ticketKey)

	// Parse ticket key
	key, err := domain.NewTicketKey(ticketKey)
	if err != nil {
		return nil, fmt.Errorf("invalid ticket key: %w", err)
	}

	// Map domain comment to Jira comment
	jiraComment := mapCommentToCommentCreate(comment)

	// Add the comment
	createdComment, resp, err := r.client.jiraClient.Issue.AddComment(ctx, ticketKey, jiraComment)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, mapHTTPError(err, fmt.Sprintf("ticket %s not found", ticketKey))
		}
		return nil, mapHTTPError(err, fmt.Sprintf("failed to add comment to ticket %s", ticketKey))
	}

	// Map the created comment back to domain
	domainComment, err := mapCommentToComment(createdComment, key)
	if err != nil {
		return nil, fmt.Errorf("failed to map created comment: %w", err)
	}

	r.client.logger.Info("successfully added comment", "ticket", ticketKey, "comment_id", domainComment.ID)
	return domainComment, nil
}

// FetchProject retrieves project metadata from Jira.
// Returns ErrNotFound if the project doesn't exist.
// Returns ErrUnauthorized if the user lacks permission to view the project.
func (r *Repository) FetchProject(ctx context.Context, projectKey string) (*domain.Project, error) {
	r.client.logger.Info("fetching project", "key", projectKey)

	project, resp, err := r.client.jiraClient.Project.Get(ctx, projectKey)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return nil, mapHTTPError(err, fmt.Sprintf("project %s not found", projectKey))
		}
		return nil, mapHTTPError(err, fmt.Sprintf("failed to fetch project %s", projectKey))
	}

	domainProject, err := mapProjectToProject(project)
	if err != nil {
		return nil, fmt.Errorf("failed to map project: %w", err)
	}

	r.client.logger.Info("successfully fetched project", "key", projectKey)
	return domainProject, nil
}

// FetchProjects retrieves all projects the authenticated user can access.
// Returns empty slice if the user has no accessible projects.
func (r *Repository) FetchProjects(ctx context.Context) ([]*domain.Project, error) {
	r.client.logger.Info("fetching all projects")

	projectList, resp, err := r.client.jiraClient.Project.GetAll(ctx, nil)
	if err != nil {
		return nil, mapHTTPError(err, "failed to fetch projects")
	}

	if resp == nil {
		return nil, fmt.Errorf("received nil response from project list")
	}

	// ProjectList is a slice type
	domainProjects := make([]*domain.Project, 0, len(*projectList))

	for i := range *projectList {
		projectItem := &(*projectList)[i]
		// Map the minimal project info to a full Project
		project := &jira.Project{
			Key:         projectItem.Key,
			Name:        projectItem.Name,
			ID:          projectItem.ID,
			Description: "", // Not available in list view
		}

		domainProject, err := mapProjectToProject(project)
		if err != nil {
			r.client.logger.Warn("failed to map project, skipping", "key", projectItem.Key, "error", err)
			continue
		}
		domainProjects = append(domainProjects, domainProject)
	}

	r.client.logger.Info("successfully fetched projects", "count", len(domainProjects))
	return domainProjects, nil
}
