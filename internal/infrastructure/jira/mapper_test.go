package jira

import (
	"testing"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/esfisher/jiramd/internal/domain"
)

func TestMapIssueToTicket(t *testing.T) {
	createdTime := jira.Time(time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC))
	updatedTime := jira.Time(time.Date(2023, 1, 16, 14, 45, 0, 0, time.UTC))

	tests := []struct {
		name    string
		issue   *jira.Issue
		wantErr bool
	}{
		{
			name: "valid issue with all fields",
			issue: &jira.Issue{
				Key: "JMD-123",
				Fields: &jira.IssueFields{
					Summary:     "Test issue",
					Description: "Test description",
					Created:     createdTime,
					Updated:     updatedTime,
					Status: &jira.Status{
						Name: "In Progress",
					},
					Type: jira.IssueType{
						Name: "Story",
					},
					Priority: &jira.Priority{
						Name: "High",
					},
					Assignee: &jira.User{
						DisplayName: "John Doe",
					},
					Reporter: &jira.User{
						DisplayName: "Jane Smith",
					},
					Labels: []string{"backend", "api"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid issue with minimal fields",
			issue: &jira.Issue{
				Key: "PROJ-1",
				Fields: &jira.IssueFields{
					Summary: "Minimal issue",
					Created: createdTime,
					Updated: updatedTime,
				},
			},
			wantErr: false,
		},
		{
			name:    "nil issue",
			issue:   nil,
			wantErr: true,
		},
		{
			name: "invalid ticket key",
			issue: &jira.Issue{
				Key: "invalid-key",
				Fields: &jira.IssueFields{
					Summary: "Test",
					Created: createdTime,
					Updated: updatedTime,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ticket, err := mapIssueToTicket(tt.issue)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapIssueToTicket() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if ticket == nil {
					t.Fatal("mapIssueToTicket() returned nil ticket")
				}
				if ticket.Key.String() != tt.issue.Key {
					t.Errorf("ticket.Key = %v, want %v", ticket.Key.String(), tt.issue.Key)
				}
				if ticket.Summary != tt.issue.Fields.Summary {
					t.Errorf("ticket.Summary = %v, want %v", ticket.Summary, tt.issue.Fields.Summary)
				}
			}
		})
	}
}

func TestMapCommentToComment(t *testing.T) {
	createdTime := "2023-01-15T10:30:00.000-0700"
	updatedTime := "2023-01-16T14:45:00.000-0700"

	ticketKey, _ := domain.NewTicketKey("JMD-123")

	tests := []struct {
		name        string
		jiraComment *jira.Comment
		ticketKey   domain.TicketKey
		wantErr     bool
	}{
		{
			name: "valid comment",
			jiraComment: &jira.Comment{
				ID:      "10001",
				Body:    "This is a test comment",
				Created: createdTime,
				Updated: updatedTime,
				Author: &jira.User{
					DisplayName: "John Doe",
				},
			},
			ticketKey: ticketKey,
			wantErr:   false,
		},
		{
			name:        "nil comment",
			jiraComment: nil,
			ticketKey:   ticketKey,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := mapCommentToComment(tt.jiraComment, tt.ticketKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapCommentToComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if comment == nil {
					t.Fatal("mapCommentToComment() returned nil comment")
				}
				if comment.ID != tt.jiraComment.ID {
					t.Errorf("comment.ID = %v, want %v", comment.ID, tt.jiraComment.ID)
				}
				if comment.Body != tt.jiraComment.Body {
					t.Errorf("comment.Body = %v, want %v", comment.Body, tt.jiraComment.Body)
				}
			}
		})
	}
}

func TestMapProjectToProject(t *testing.T) {
	tests := []struct {
		name        string
		jiraProject *jira.Project
		wantErr     bool
	}{
		{
			name: "valid project",
			jiraProject: &jira.Project{
				Key:         "JMD",
				Name:        "Jira Markdown Sync",
				Description: "A project for syncing Jira to markdown",
			},
			wantErr: false,
		},
		{
			name: "valid project without description",
			jiraProject: &jira.Project{
				Key:  "PROJ",
				Name: "Test Project",
			},
			wantErr: false,
		},
		{
			name:        "nil project",
			jiraProject: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := mapProjectToProject(tt.jiraProject)
			if (err != nil) != tt.wantErr {
				t.Errorf("mapProjectToProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if project == nil {
					t.Fatal("mapProjectToProject() returned nil project")
				}
				if project.Key != tt.jiraProject.Key {
					t.Errorf("project.Key = %v, want %v", project.Key, tt.jiraProject.Key)
				}
				if project.Name != tt.jiraProject.Name {
					t.Errorf("project.Name = %v, want %v", project.Name, tt.jiraProject.Name)
				}
			}
		})
	}
}


func TestMapCommentToCommentCreate(t *testing.T) {
	ticketKey, _ := domain.NewTicketKey("JMD-123")
	comment, _ := domain.NewComment("1", ticketKey, "John Doe", "Test comment body", time.Now(), time.Now())

	jiraComment := mapCommentToCommentCreate(comment)

	if jiraComment == nil {
		t.Fatal("mapCommentToCommentCreate() returned nil")
	}

	if jiraComment.Body != "Test comment body" {
		t.Errorf("jiraComment.Body = %v, want Test comment body", jiraComment.Body)
	}
}
