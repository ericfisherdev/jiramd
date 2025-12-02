package domain

import (
	"testing"
	"time"
)

func TestNewComment(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	created := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		ticketKey TicketKey
		author    string
		body      string
		created   time.Time
		updated   time.Time
		wantErr   bool
	}{
		{
			name:      "valid comment",
			id:        "10001",
			ticketKey: key,
			author:    "user@example.com",
			body:      "This is a comment",
			created:   created,
			updated:   updated,
			wantErr:   false,
		},
		{
			name:      "valid with whitespace (trimmed)",
			id:        "  10002  ",
			ticketKey: key,
			author:    "  user@example.com  ",
			body:      "Comment body",
			created:   created,
			updated:   updated,
			wantErr:   false,
		},
		{
			name:      "empty id",
			id:        "",
			ticketKey: key,
			author:    "user@example.com",
			body:      "Comment",
			created:   created,
			updated:   updated,
			wantErr:   true,
		},
		{
			name:      "zero ticket key",
			id:        "10001",
			ticketKey: TicketKey{},
			author:    "user@example.com",
			body:      "Comment",
			created:   created,
			updated:   updated,
			wantErr:   true,
		},
		{
			name:      "empty author",
			id:        "10001",
			ticketKey: key,
			author:    "",
			body:      "Comment",
			created:   created,
			updated:   updated,
			wantErr:   true,
		},
		{
			name:      "zero created time",
			id:        "10001",
			ticketKey: key,
			author:    "user@example.com",
			body:      "Comment",
			created:   time.Time{},
			updated:   updated,
			wantErr:   true,
		},
		{
			name:      "zero updated time",
			id:        "10001",
			ticketKey: key,
			author:    "user@example.com",
			body:      "Comment",
			created:   created,
			updated:   time.Time{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := NewComment(tt.id, tt.ticketKey, tt.author, tt.body, tt.created, tt.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if comment == nil {
					t.Fatal("NewComment() returned nil comment")
				}
				// Verify timestamps are in UTC
				if comment.Created.Location() != time.UTC {
					t.Error("Created timestamp should be in UTC")
				}
				if comment.Updated.Location() != time.UTC {
					t.Error("Updated timestamp should be in UTC")
				}
			}
		})
	}
}

func TestComment_Validate(t *testing.T) {
	validKey, _ := NewTicketKey("JMD-123")
	validTime := time.Now()

	tests := []struct {
		name    string
		comment *Comment
		wantErr bool
	}{
		{
			name: "valid comment",
			comment: &Comment{
				ID:        "10001",
				TicketKey: validKey,
				Author:    "user@example.com",
				Body:      "Comment body",
				Created:   validTime,
				Updated:   validTime,
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			comment: &Comment{
				ID:        "",
				TicketKey: validKey,
				Author:    "user@example.com",
				Body:      "Comment",
				Created:   validTime,
				Updated:   validTime,
			},
			wantErr: true,
		},
		{
			name: "whitespace ID",
			comment: &Comment{
				ID:        "   ",
				TicketKey: validKey,
				Author:    "user@example.com",
				Body:      "Comment",
				Created:   validTime,
				Updated:   validTime,
			},
			wantErr: true,
		},
		{
			name: "zero ticket key",
			comment: &Comment{
				ID:        "10001",
				TicketKey: TicketKey{},
				Author:    "user@example.com",
				Body:      "Comment",
				Created:   validTime,
				Updated:   validTime,
			},
			wantErr: true,
		},
		{
			name: "empty author",
			comment: &Comment{
				ID:        "10001",
				TicketKey: validKey,
				Author:    "",
				Body:      "Comment",
				Created:   validTime,
				Updated:   validTime,
			},
			wantErr: true,
		},
		{
			name: "zero created time",
			comment: &Comment{
				ID:        "10001",
				TicketKey: validKey,
				Author:    "user@example.com",
				Body:      "Comment",
				Created:   time.Time{},
				Updated:   validTime,
			},
			wantErr: true,
		},
		{
			name: "zero updated time",
			comment: &Comment{
				ID:        "10001",
				TicketKey: validKey,
				Author:    "user@example.com",
				Body:      "Comment",
				Created:   validTime,
				Updated:   time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.comment.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
