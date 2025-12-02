package domain

import (
	"testing"
	"time"
)

func TestNewSyncTimestamp(t *testing.T) {
	localTime := time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local)
	st := NewSyncTimestamp(localTime)

	if st.Time().Location() != time.UTC {
		t.Error("SyncTimestamp should be in UTC")
	}

	if st.IsZero() {
		t.Error("Non-zero time should not return IsZero() = true")
	}

	zeroSt := NewSyncTimestamp(time.Time{})
	if !zeroSt.IsZero() {
		t.Error("Zero time should return IsZero() = true")
	}
}

func TestSyncTimestamp_Comparison(t *testing.T) {
	t1 := NewSyncTimestamp(time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC))
	t2 := NewSyncTimestamp(time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC))

	if !t1.Before(t2) {
		t.Error("t1 should be before t2")
	}

	if t2.Before(t1) {
		t.Error("t2 should not be before t1")
	}

	if !t2.After(t1) {
		t.Error("t2 should be after t1")
	}

	if t1.After(t2) {
		t.Error("t1 should not be after t2")
	}
}

func TestNewSyncState(t *testing.T) {
	tests := []struct {
		name       string
		projectKey string
		wantErr    bool
	}{
		{
			name:       "valid project key",
			projectKey: "JMD",
			wantErr:    false,
		},
		{
			name:       "empty project key",
			projectKey: "",
			wantErr:    true,
		},
		{
			name:       "whitespace project key",
			projectKey: "   ",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, err := NewSyncState(tt.projectKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSyncState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ss == nil {
					t.Fatal("NewSyncState() returned nil")
				}
				if ss.TicketCount != 0 {
					t.Errorf("TicketCount = %d, want 0", ss.TicketCount)
				}
				if !ss.LastFullSync.IsZero() {
					t.Error("LastFullSync should be zero initially")
				}
				if !ss.LastIncrementalSync.IsZero() {
					t.Error("LastIncrementalSync should be zero initially")
				}
			}
		})
	}
}

func TestSyncState_UpdateTimestamps(t *testing.T) {
	ss, _ := NewSyncState("JMD")

	before := time.Now()
	ss.UpdateFullSync()
	after := time.Now()

	if ss.LastFullSync.IsZero() {
		t.Error("LastFullSync should not be zero after update")
	}

	fullSyncTime := ss.LastFullSync.Time()
	if fullSyncTime.Before(before) || fullSyncTime.After(after) {
		t.Error("LastFullSync timestamp out of expected range")
	}

	before = time.Now()
	ss.UpdateIncrementalSync()
	after = time.Now()

	if ss.LastIncrementalSync.IsZero() {
		t.Error("LastIncrementalSync should not be zero after update")
	}

	incSyncTime := ss.LastIncrementalSync.Time()
	if incSyncTime.Before(before) || incSyncTime.After(after) {
		t.Error("LastIncrementalSync timestamp out of expected range")
	}
}

func TestNewTicketState(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	jiraUpdated := time.Now()

	tests := []struct {
		name        string
		projectKey  string
		ticketKey   TicketKey
		jiraUpdated time.Time
		wantErr     bool
	}{
		{
			name:        "valid ticket state",
			projectKey:  "JMD",
			ticketKey:   key,
			jiraUpdated: jiraUpdated,
			wantErr:     false,
		},
		{
			name:        "empty project key",
			projectKey:  "",
			ticketKey:   key,
			jiraUpdated: jiraUpdated,
			wantErr:     true,
		},
		{
			name:        "zero ticket key",
			projectKey:  "JMD",
			ticketKey:   TicketKey{},
			jiraUpdated: jiraUpdated,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := NewTicketState(tt.projectKey, tt.ticketKey, tt.jiraUpdated)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTicketState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if ts == nil {
					t.Fatal("NewTicketState() returned nil")
				}
				if ts.Status != SyncStatusInSync {
					t.Errorf("Status = %v, want %v", ts.Status, SyncStatusInSync)
				}
				if ts.LocalModified != nil {
					t.Error("LocalModified should be nil initially")
				}
				if ts.JiraUpdated.Time().Location() != time.UTC {
					t.Error("JiraUpdated should be in UTC")
				}
			}
		})
	}
}

func TestTicketState_MarkLocalModified(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	ts, _ := NewTicketState("JMD", key, time.Now())

	modTime := time.Now()
	ts.MarkLocalModified(modTime)

	if ts.LocalModified == nil {
		t.Fatal("LocalModified should not be nil after marking")
	}

	if ts.LocalModified.Time().Location() != time.UTC {
		t.Error("LocalModified should be in UTC")
	}

	if ts.Status != SyncStatusLocalModified {
		t.Errorf("Status = %v, want %v", ts.Status, SyncStatusLocalModified)
	}
}

func TestTicketState_DetectConflict(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")

	t.Run("no conflict - no local modifications", func(t *testing.T) {
		ts, _ := NewTicketState("JMD", key, time.Now())
		if ts.DetectConflict() {
			t.Error("Should not detect conflict without local modifications")
		}
	})

	t.Run("no conflict - only local modified", func(t *testing.T) {
		pastTime := time.Now().Add(-10 * time.Minute)
		ts, _ := NewTicketState("JMD", key, pastTime)
		ts.MarkLocalModified(time.Now())

		if ts.DetectConflict() {
			t.Error("Should not detect conflict when only local is modified")
		}
	})

	t.Run("no conflict - only jira modified", func(t *testing.T) {
		pastTime := time.Now().Add(-10 * time.Minute)
		ts, _ := NewTicketState("JMD", key, pastTime)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		ts.JiraUpdated = NewSyncTimestamp(time.Now())

		if ts.DetectConflict() {
			t.Error("Should not detect conflict when only Jira is modified")
		}
	})

	t.Run("conflict - both modified", func(t *testing.T) {
		pastTime := time.Now().Add(-10 * time.Minute)
		ts, _ := NewTicketState("JMD", key, pastTime)

		// Modify local after last sync
		ts.MarkLocalModified(time.Now())

		// Simulate Jira update after last sync
		ts.JiraUpdated = NewSyncTimestamp(time.Now())

		if !ts.DetectConflict() {
			t.Error("Should detect conflict when both are modified")
		}

		if ts.Status != SyncStatusConflict {
			t.Errorf("Status = %v, want %v", ts.Status, SyncStatusConflict)
		}
	})
}

func TestTicketState_UpdateSynced(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	ts, _ := NewTicketState("JMD", key, time.Now())
	ts.MarkLocalModified(time.Now())

	contentHash := "abc123def456"
	jiraUpdated := time.Now()

	before := time.Now()
	ts.UpdateSynced(contentHash, jiraUpdated)
	after := time.Now()

	if ts.ContentHash != contentHash {
		t.Errorf("ContentHash = %v, want %v", ts.ContentHash, contentHash)
	}

	if ts.Status != SyncStatusInSync {
		t.Errorf("Status = %v, want %v", ts.Status, SyncStatusInSync)
	}

	if ts.LocalModified != nil {
		t.Error("LocalModified should be nil after sync")
	}

	syncTime := ts.LastSynced.Time()
	if syncTime.Before(before) || syncTime.After(after) {
		t.Error("LastSynced timestamp out of expected range")
	}
}

func TestNewSyncResult(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	sr := NewSyncResult(key)

	if !sr.Success {
		t.Error("New SyncResult should be successful by default")
	}

	if sr.ConflictDetected {
		t.Error("ConflictDetected should be false by default")
	}

	if sr.TicketKey.String() != "JMD-123" {
		t.Errorf("TicketKey = %v, want JMD-123", sr.TicketKey.String())
	}

	if sr.OperationsPerformed == nil {
		t.Error("OperationsPerformed should be initialized")
	}
}

func TestSyncResult_MarkFailed(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	sr := NewSyncResult(key)

	err := ErrSyncConflict
	sr.MarkFailed(err)

	if sr.Success {
		t.Error("Success should be false after MarkFailed")
	}

	if sr.Error != err.Error() {
		t.Errorf("Error = %v, want %v", sr.Error, err.Error())
	}
}

func TestSyncResult_MarkConflict(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	sr := NewSyncResult(key)

	sr.MarkConflict()

	if !sr.ConflictDetected {
		t.Error("ConflictDetected should be true after MarkConflict")
	}

	if len(sr.OperationsPerformed) != 1 {
		t.Errorf("OperationsPerformed length = %d, want 1", len(sr.OperationsPerformed))
	}

	if sr.OperationsPerformed[0] != "conflict_detected" {
		t.Errorf("Operation = %v, want 'conflict_detected'", sr.OperationsPerformed[0])
	}
}

func TestSyncResult_AddOperation(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	sr := NewSyncResult(key)

	sr.AddOperation("pull_ticket")
	sr.AddOperation("pull_comments")

	if len(sr.OperationsPerformed) != 2 {
		t.Errorf("OperationsPerformed length = %d, want 2", len(sr.OperationsPerformed))
	}

	if sr.OperationsPerformed[0] != "pull_ticket" {
		t.Errorf("Operation[0] = %v, want 'pull_ticket'", sr.OperationsPerformed[0])
	}
}

func TestNewPendingOperation(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")

	tests := []struct {
		name       string
		projectKey string
		ticketKey  TicketKey
		operation  OperationType
		payload    string
		wantErr    bool
	}{
		{
			name:       "valid push status",
			projectKey: "JMD",
			ticketKey:  key,
			operation:  OpPushStatus,
			payload:    `{"status":"Done"}`,
			wantErr:    false,
		},
		{
			name:       "valid post comment",
			projectKey: "JMD",
			ticketKey:  key,
			operation:  OpPostComment,
			payload:    `{"body":"test"}`,
			wantErr:    false,
		},
		{
			name:       "empty project key",
			projectKey: "",
			ticketKey:  key,
			operation:  OpPushStatus,
			payload:    "{}",
			wantErr:    true,
		},
		{
			name:       "zero ticket key",
			projectKey: "JMD",
			ticketKey:  TicketKey{},
			operation:  OpPushStatus,
			payload:    "{}",
			wantErr:    true,
		},
		{
			name:       "invalid operation type",
			projectKey: "JMD",
			ticketKey:  key,
			operation:  OperationType("invalid"),
			payload:    "{}",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po, err := NewPendingOperation(tt.projectKey, tt.ticketKey, tt.operation, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPendingOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if po == nil {
					t.Fatal("NewPendingOperation() returned nil")
				}
				if po.Attempts != 0 {
					t.Errorf("Attempts = %d, want 0", po.Attempts)
				}
				if po.LastError != "" {
					t.Errorf("LastError = %v, want empty", po.LastError)
				}
				if po.CreatedAt.IsZero() {
					t.Error("CreatedAt should not be zero")
				}
			}
		})
	}
}

func TestPendingOperation_RecordAttempt(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	po, _ := NewPendingOperation("JMD", key, OpPushStatus, "{}")

	err := ErrSyncConflict
	po.RecordAttempt(err)

	if po.Attempts != 1 {
		t.Errorf("Attempts = %d, want 1", po.Attempts)
	}

	if po.LastError != err.Error() {
		t.Errorf("LastError = %v, want %v", po.LastError, err.Error())
	}

	// Record another attempt
	po.RecordAttempt(nil)

	if po.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", po.Attempts)
	}
}

func TestPendingOperation_ShouldRetry(t *testing.T) {
	key, _ := NewTicketKey("JMD-123")
	po, _ := NewPendingOperation("JMD", key, OpPushStatus, "{}")

	// Should retry initially
	if !po.ShouldRetry() {
		t.Error("ShouldRetry() should be true initially")
	}

	// Record attempts
	po.RecordAttempt(ErrSyncConflict)
	if !po.ShouldRetry() {
		t.Error("ShouldRetry() should be true after 1 attempt")
	}

	po.RecordAttempt(ErrSyncConflict)
	if !po.ShouldRetry() {
		t.Error("ShouldRetry() should be true after 2 attempts")
	}

	po.RecordAttempt(ErrSyncConflict)
	if po.ShouldRetry() {
		t.Error("ShouldRetry() should be false after 3 attempts")
	}
}
