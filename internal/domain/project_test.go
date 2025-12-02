package domain

import (
	"testing"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		projName string
		wantErr  bool
		wantKey  string
		wantName string
	}{
		{
			name:     "valid project",
			key:      "JMD",
			projName: "Jira Markdown Daemon",
			wantErr:  false,
			wantKey:  "JMD",
			wantName: "Jira Markdown Daemon",
		},
		{
			name:     "valid with lowercase (converted to uppercase)",
			key:      "jmd",
			projName: "Project",
			wantErr:  false,
			wantKey:  "JMD",
			wantName: "Project",
		},
		{
			name:     "valid with whitespace (trimmed)",
			key:      "  PROJ  ",
			projName: "  Project Name  ",
			wantErr:  false,
			wantKey:  "PROJ",
			wantName: "Project Name",
		},
		{
			name:     "valid with numbers",
			key:      "P2D",
			projName: "Project 2D",
			wantErr:  false,
			wantKey:  "P2D",
			wantName: "Project 2D",
		},
		{
			name:     "valid max length (10 chars)",
			key:      "ABCDEFGHIJ",
			projName: "Long Project",
			wantErr:  false,
			wantKey:  "ABCDEFGHIJ",
			wantName: "Long Project",
		},
		{
			name:     "empty key",
			key:      "",
			projName: "Project",
			wantErr:  true,
		},
		{
			name:     "empty name",
			key:      "PROJ",
			projName: "",
			wantErr:  true,
		},
		{
			name:     "single character key",
			key:      "J",
			projName: "Project",
			wantErr:  true,
		},
		{
			name:     "too long key (11 chars)",
			key:      "ABCDEFGHIJK",
			projName: "Project",
			wantErr:  true,
		},
		{
			name:     "invalid characters in key",
			key:      "JMD_TEST",
			projName: "Project",
			wantErr:  true,
		},
		{
			name:     "lowercase not at start",
			key:      "Jmd",
			projName: "Project",
			wantErr:  false, // Will be converted to uppercase
			wantKey:  "JMD",
			wantName: "Project",
		},
		{
			name:     "key with dash",
			key:      "JMD-123",
			projName: "Project",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := NewProject(tt.key, tt.projName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if proj == nil {
					t.Fatal("NewProject() returned nil")
				}
				if proj.Key != tt.wantKey {
					t.Errorf("Key = %v, want %v", proj.Key, tt.wantKey)
				}
				if proj.Name != tt.wantName {
					t.Errorf("Name = %v, want %v", proj.Name, tt.wantName)
				}
				if proj.CustomFields == nil {
					t.Error("CustomFields should be initialized")
				}
			}
		})
	}
}

func TestProject_AddCustomField(t *testing.T) {
	proj, _ := NewProject("JMD", "Test Project")

	t.Run("add valid field", func(t *testing.T) {
		field, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
		err := proj.AddCustomField(field)
		if err != nil {
			t.Errorf("AddCustomField() error = %v", err)
		}
		if len(proj.CustomFields) != 1 {
			t.Errorf("CustomFields length = %d, want 1", len(proj.CustomFields))
		}
	})

	t.Run("add duplicate field", func(t *testing.T) {
		field, _ := NewCustomField("dev", "Dev Assignment", "labels", SyncBidirectional)
		err := proj.AddCustomField(field)
		if err == nil {
			t.Error("AddCustomField() should error on duplicate field name")
		}
	})

	t.Run("add nil field", func(t *testing.T) {
		err := proj.AddCustomField(nil)
		if err == nil {
			t.Error("AddCustomField() should error on nil field")
		}
	})

	t.Run("add invalid field", func(t *testing.T) {
		invalidField := &CustomField{
			Name:          "",
			DisplayName:   "Invalid",
			Source:        "source",
			SyncDirection: SyncBidirectional,
		}
		err := proj.AddCustomField(invalidField)
		if err == nil {
			t.Error("AddCustomField() should error on invalid field")
		}
	})
}

func TestProject_GetCustomField(t *testing.T) {
	proj, _ := NewProject("JMD", "Test Project")
	field1, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
	field2, _ := NewCustomField("priority", "Priority", "priority", SyncJiraToLocal)
	proj.AddCustomField(field1)
	proj.AddCustomField(field2)

	t.Run("get existing field", func(t *testing.T) {
		got := proj.GetCustomField("dev")
		if got == nil {
			t.Fatal("GetCustomField() returned nil")
		}
		if got.Name != "dev" {
			t.Errorf("Got field name %v, want 'dev'", got.Name)
		}
	})

	t.Run("get non-existent field", func(t *testing.T) {
		got := proj.GetCustomField("nonexistent")
		if got != nil {
			t.Error("GetCustomField() should return nil for non-existent field")
		}
	})
}

func TestProject_RemoveCustomField(t *testing.T) {
	proj, _ := NewProject("JMD", "Test Project")
	field1, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
	field2, _ := NewCustomField("priority", "Priority", "priority", SyncJiraToLocal)
	proj.AddCustomField(field1)
	proj.AddCustomField(field2)

	t.Run("remove existing field", func(t *testing.T) {
		removed := proj.RemoveCustomField("dev")
		if !removed {
			t.Error("RemoveCustomField() should return true for existing field")
		}
		if len(proj.CustomFields) != 1 {
			t.Errorf("CustomFields length = %d, want 1", len(proj.CustomFields))
		}
		if proj.GetCustomField("dev") != nil {
			t.Error("Field should be removed")
		}
	})

	t.Run("remove non-existent field", func(t *testing.T) {
		removed := proj.RemoveCustomField("nonexistent")
		if removed {
			t.Error("RemoveCustomField() should return false for non-existent field")
		}
	})
}

func TestProject_BidirectionalFields(t *testing.T) {
	proj, _ := NewProject("JMD", "Test Project")
	field1, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
	field2, _ := NewCustomField("priority", "Priority", "priority", SyncJiraToLocal)
	field3, _ := NewCustomField("impl", "Implementation", "local", SyncLocalOnly)
	field4, _ := NewCustomField("status", "Status", "status", SyncBidirectional)

	proj.AddCustomField(field1)
	proj.AddCustomField(field2)
	proj.AddCustomField(field3)
	proj.AddCustomField(field4)

	bidirectional := proj.BidirectionalFields()
	if len(bidirectional) != 2 {
		t.Errorf("BidirectionalFields() length = %d, want 2", len(bidirectional))
	}

	// Verify the returned fields are bidirectional
	for _, field := range bidirectional {
		if !field.IsBidirectional() {
			t.Errorf("Field %s should be bidirectional", field.Name)
		}
	}
}

func TestProject_DerivedFields(t *testing.T) {
	proj, _ := NewProject("JMD", "Test Project")
	field1, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
	field1.Condition = "has-label('dev1','dev2')"
	field2, _ := NewCustomField("priority", "Priority", "priority", SyncJiraToLocal)
	field3, _ := NewCustomField("status", "Status", "status", SyncBidirectional)
	field3.Condition = "has-label('in-progress')"

	proj.AddCustomField(field1)
	proj.AddCustomField(field2)
	proj.AddCustomField(field3)

	derived := proj.DerivedFields()
	if len(derived) != 2 {
		t.Errorf("DerivedFields() length = %d, want 2", len(derived))
	}

	// Verify the returned fields have conditions
	for _, field := range derived {
		if !field.IsDerived() {
			t.Errorf("Field %s should be derived", field.Name)
		}
	}
}
