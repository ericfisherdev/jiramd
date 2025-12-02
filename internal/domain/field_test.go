package domain

import (
	"testing"
)

func TestFieldValue(t *testing.T) {
	tests := []struct {
		name       string
		value      interface{}
		wantString string
		wantIsZero bool
	}{
		{
			name:       "string value",
			value:      "test",
			wantString: "test",
			wantIsZero: false,
		},
		{
			name:       "int value",
			value:      123,
			wantString: "123",
			wantIsZero: false,
		},
		{
			name:       "nil value",
			value:      nil,
			wantString: "",
			wantIsZero: true,
		},
		{
			name:       "empty string",
			value:      "",
			wantString: "",
			wantIsZero: false, // Empty string is not nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fv := NewFieldValue(tt.value)

			if got := fv.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}

			if got := fv.IsZero(); got != tt.wantIsZero {
				t.Errorf("IsZero() = %v, want %v", got, tt.wantIsZero)
			}

			if got := fv.Raw(); got != tt.value {
				t.Errorf("Raw() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestNewCustomField(t *testing.T) {
	tests := []struct {
		name          string
		fieldName     string
		displayName   string
		source        string
		syncDirection SyncDirection
		wantErr       bool
	}{
		{
			name:          "valid bidirectional field",
			fieldName:     "dev_assignment",
			displayName:   "Dev Assignment",
			source:        "labels",
			syncDirection: SyncBidirectional,
			wantErr:       false,
		},
		{
			name:          "valid jira-to-local field",
			fieldName:     "story_points",
			displayName:   "Story Points",
			source:        "customfield_10001",
			syncDirection: SyncJiraToLocal,
			wantErr:       false,
		},
		{
			name:          "valid local-only field",
			fieldName:     "implementation_plan",
			displayName:   "Implementation Plan",
			source:        "local",
			syncDirection: SyncLocalOnly,
			wantErr:       false,
		},
		{
			name:          "empty name",
			fieldName:     "",
			displayName:   "Display",
			source:        "source",
			syncDirection: SyncBidirectional,
			wantErr:       true,
		},
		{
			name:          "empty display name",
			fieldName:     "field",
			displayName:   "",
			source:        "source",
			syncDirection: SyncBidirectional,
			wantErr:       true,
		},
		{
			name:          "empty source",
			fieldName:     "field",
			displayName:   "Display",
			source:        "",
			syncDirection: SyncBidirectional,
			wantErr:       true,
		},
		{
			name:          "invalid sync direction",
			fieldName:     "field",
			displayName:   "Display",
			source:        "source",
			syncDirection: SyncDirection("invalid"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, err := NewCustomField(tt.fieldName, tt.displayName, tt.source, tt.syncDirection)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCustomField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if cf == nil {
					t.Fatal("NewCustomField() returned nil")
				}
				if cf.Name != tt.fieldName {
					t.Errorf("Name = %v, want %v", cf.Name, tt.fieldName)
				}
				if cf.ValidValues == nil {
					t.Error("ValidValues should be initialized")
				}
			}
		})
	}
}

func TestCustomField_ValidateValue(t *testing.T) {
	tests := []struct {
		name        string
		validValues []string
		value       string
		wantErr     bool
	}{
		{
			name:        "no validation (empty valid values)",
			validValues: []string{},
			value:       "anything",
			wantErr:     false,
		},
		{
			name:        "valid value in list",
			validValues: []string{"dev1", "dev2", "dev3"},
			value:       "dev2",
			wantErr:     false,
		},
		{
			name:        "invalid value not in list",
			validValues: []string{"dev1", "dev2", "dev3"},
			value:       "dev4",
			wantErr:     true,
		},
		{
			name:        "case sensitive validation",
			validValues: []string{"dev1", "dev2"},
			value:       "DEV1",
			wantErr:     true,
		},
		{
			name:        "value with leading whitespace",
			validValues: []string{"dev1", "dev2", "dev3"},
			value:       "  dev2",
			wantErr:     false,
		},
		{
			name:        "value with trailing whitespace",
			validValues: []string{"dev1", "dev2", "dev3"},
			value:       "dev2  ",
			wantErr:     false,
		},
		{
			name:        "value with both leading and trailing whitespace",
			validValues: []string{"dev1", "dev2", "dev3"},
			value:       "  dev2  ",
			wantErr:     false,
		},
		{
			name:        "valid entry with whitespace",
			validValues: []string{"dev1", "  dev2  ", "dev3"},
			value:       "dev2",
			wantErr:     false,
		},
		{
			name:        "both value and valid have whitespace",
			validValues: []string{"dev1", "  dev2  ", "dev3"},
			value:       "  dev2  ",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, _ := NewCustomField("test", "Test", "labels", SyncBidirectional)
			cf.ValidValues = tt.validValues

			err := cf.ValidateValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomField_IsBidirectional(t *testing.T) {
	tests := []struct {
		name          string
		syncDirection SyncDirection
		want          bool
	}{
		{
			name:          "bidirectional",
			syncDirection: SyncBidirectional,
			want:          true,
		},
		{
			name:          "jira to local",
			syncDirection: SyncJiraToLocal,
			want:          false,
		},
		{
			name:          "local only",
			syncDirection: SyncLocalOnly,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, _ := NewCustomField("test", "Test", "source", tt.syncDirection)
			if got := cf.IsBidirectional(); got != tt.want {
				t.Errorf("IsBidirectional() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomField_IsDerived(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		want      bool
	}{
		{
			name:      "has condition",
			condition: "has-label('dev1','dev2')",
			want:      true,
		},
		{
			name:      "no condition",
			condition: "",
			want:      false,
		},
		{
			name:      "whitespace condition",
			condition: "   ",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf, _ := NewCustomField("test", "Test", "source", SyncBidirectional)
			cf.Condition = tt.condition
			if got := cf.IsDerived(); got != tt.want {
				t.Errorf("IsDerived() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDerivedField(t *testing.T) {
	cf, _ := NewCustomField("dev", "Dev", "labels", SyncBidirectional)
	cf.DefaultValue = "none"

	t.Run("new derived field uses default", func(t *testing.T) {
		df := NewDerivedField(cf)
		if df.Value() != "none" {
			t.Errorf("Value() = %v, want 'none'", df.Value())
		}
		if df.UsedDefault {
			t.Error("UsedDefault should be false initially (not explicitly set)")
		}
	})

	t.Run("set matched value", func(t *testing.T) {
		df := NewDerivedField(cf)
		df.SetMatchedValue("dev1")
		if df.Value() != "dev1" {
			t.Errorf("Value() = %v, want 'dev1'", df.Value())
		}
		if df.UsedDefault {
			t.Error("UsedDefault should be false after setting matched value")
		}
	})

	t.Run("set default explicitly", func(t *testing.T) {
		df := NewDerivedField(cf)
		df.SetMatchedValue("dev1")
		df.SetDefault()
		if df.Value() != "none" {
			t.Errorf("Value() = %v, want 'none'", df.Value())
		}
		if !df.UsedDefault {
			t.Error("UsedDefault should be true after SetDefault()")
		}
	})
}
