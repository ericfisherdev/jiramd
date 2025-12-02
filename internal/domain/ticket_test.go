package domain

import (
	"testing"
	"time"
)

func TestNewTicketKey(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    string
	}{
		{
			name:    "valid ticket key",
			input:   "JMD-123",
			wantErr: false,
			want:    "JMD-123",
		},
		{
			name:    "valid with multiple digits",
			input:   "PROJ-99999",
			wantErr: false,
			want:    "PROJ-99999",
		},
		{
			name:    "valid with numbers in project key",
			input:   "P2D-456",
			wantErr: false,
			want:    "P2D-456",
		},
		{
			name:    "valid with whitespace (trimmed)",
			input:   "  JMD-123  ",
			wantErr: false,
			want:    "JMD-123",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "lowercase project key",
			input:   "jmd-123",
			wantErr: true,
		},
		{
			name:    "missing dash",
			input:   "JMD123",
			wantErr: true,
		},
		{
			name:    "missing number",
			input:   "JMD-",
			wantErr: true,
		},
		{
			name:    "missing project key",
			input:   "-123",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			input:   "JMD_123",
			wantErr: true,
		},
		{
			name:    "multiple dashes",
			input:   "JMD-123-456",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTicketKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTicketKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("NewTicketKey() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestTicketKey_ProjectKey(t *testing.T) {
	tests := []struct {
		name       string
		ticketKey  string
		wantPrefix string
	}{
		{
			name:       "simple project key",
			ticketKey:  "JMD-123",
			wantPrefix: "JMD",
		},
		{
			name:       "longer project key",
			ticketKey:  "PROJ-456",
			wantPrefix: "PROJ",
		},
		{
			name:       "project key with numbers",
			ticketKey:  "P2D-789",
			wantPrefix: "P2D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := NewTicketKey(tt.ticketKey)
			if err != nil {
				t.Fatalf("NewTicketKey() error = %v", err)
			}
			if got := key.ProjectKey(); got != tt.wantPrefix {
				t.Errorf("ProjectKey() = %v, want %v", got, tt.wantPrefix)
			}
		})
	}
}

func TestTicketKey_IsZero(t *testing.T) {
	zero := TicketKey{}
	if !zero.IsZero() {
		t.Error("Zero TicketKey should return true for IsZero()")
	}

	key, _ := NewTicketKey("JMD-123")
	if key.IsZero() {
		t.Error("Valid TicketKey should return false for IsZero()")
	}
}

func TestNewTicket(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	created := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)

	ticket := NewTicket(key, "Test ticket", created, updated)

	if ticket.Key.String() != "JMD-123" {
		t.Errorf("Key = %v, want JMD-123", ticket.Key.String())
	}
	if ticket.Summary != "Test ticket" {
		t.Errorf("Summary = %v, want 'Test ticket'", ticket.Summary)
	}
	if !ticket.Created.Equal(created) {
		t.Errorf("Created = %v, want %v", ticket.Created, created)
	}
	if !ticket.Updated.Equal(updated) {
		t.Errorf("Updated = %v, want %v", ticket.Updated, updated)
	}
	if ticket.Labels == nil {
		t.Error("Labels should be initialized")
	}
	if ticket.CustomFields == nil {
		t.Error("CustomFields should be initialized")
	}
}

func TestTicket_Validate(t *testing.T) {
	validKey, _ := NewTicketKey("JMD-123")
	validTime := time.Now()

	tests := []struct {
		name    string
		ticket  *Ticket
		wantErr bool
	}{
		{
			name:    "valid ticket",
			ticket:  NewTicket(validKey, "Valid summary", validTime, validTime),
			wantErr: false,
		},
		{
			name: "empty key",
			ticket: &Ticket{
				Key:     TicketKey{},
				Summary: "Summary",
				Created: validTime,
				Updated: validTime,
			},
			wantErr: true,
		},
		{
			name: "empty summary",
			ticket: &Ticket{
				Key:     validKey,
				Summary: "",
				Created: validTime,
				Updated: validTime,
			},
			wantErr: true,
		},
		{
			name: "whitespace summary",
			ticket: &Ticket{
				Key:     validKey,
				Summary: "   ",
				Created: validTime,
				Updated: validTime,
			},
			wantErr: true,
		},
		{
			name: "zero created time",
			ticket: &Ticket{
				Key:     validKey,
				Summary: "Summary",
				Created: time.Time{},
				Updated: validTime,
			},
			wantErr: true,
		},
		{
			name: "zero updated time",
			ticket: &Ticket{
				Key:     validKey,
				Summary: "Summary",
				Created: validTime,
				Updated: time.Time{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ticket.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTicket_ContentHash(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	now := time.Now()

	ticket1 := NewTicket(key, "Test", now, now)
	ticket1.Description = "Description"
	ticket1.Status = "In Progress"
	ticket1.Priority = "High"
	ticket1.Assignee = "user@example.com"
	ticket1.Labels = []string{"bug", "critical"}

	ticket2 := NewTicket(key, "Test", now, now)
	ticket2.Description = "Description"
	ticket2.Status = "In Progress"
	ticket2.Priority = "High"
	ticket2.Assignee = "user@example.com"
	ticket2.Labels = []string{"bug", "critical"}

	hash1 := ticket1.ContentHash()
	hash2 := ticket2.ContentHash()

	if hash1 != hash2 {
		t.Errorf("Identical tickets should have same hash: %s != %s", hash1, hash2)
	}

	// Modify ticket2
	ticket2.Status = "Done"
	hash3 := ticket2.ContentHash()

	if hash1 == hash3 {
		t.Error("Different tickets should have different hashes")
	}

	// Hash should be 32 hex characters (MD5)
	if len(hash1) != 32 {
		t.Errorf("Hash length = %d, want 32", len(hash1))
	}
}

func TestTicket_ContentHash_Deterministic(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	now := time.Now()

	// Create same ticket multiple times and verify hash is deterministic
	hashes := make([]string, 5)
	for i := 0; i < 5; i++ {
		ticket := NewTicket(key, "Test", now, now)
		ticket.Description = "Description"
		ticket.Status = "In Progress"
		ticket.Labels = []string{"a", "b", "c"}
		hashes[i] = ticket.ContentHash()
	}

	// All hashes should be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Hash %d = %s, want %s (non-deterministic)", i, hashes[i], hashes[0])
		}
	}
}
