// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"fmt"
	"strings"
)

// SyncDirection defines which direction a field should be synchronized.
type SyncDirection string

const (
	// SyncBidirectional means the field syncs both Jira <-> Local
	SyncBidirectional SyncDirection = "bidirectional"

	// SyncJiraToLocal means the field only syncs Jira -> Local (read-only from Jira)
	SyncJiraToLocal SyncDirection = "jira_to_local"

	// SyncLocalOnly means the field is local-only and never synced to Jira
	SyncLocalOnly SyncDirection = "local_only"
)

// FieldValue is a value object representing a field's value with type safety.
// It wraps the underlying value and provides type-safe access methods.
type FieldValue struct {
	raw interface{}
}

// NewFieldValue creates a new FieldValue from any value.
func NewFieldValue(value interface{}) FieldValue {
	return FieldValue{raw: value}
}

// Raw returns the underlying raw value.
func (fv FieldValue) Raw() interface{} {
	return fv.raw
}

// String returns the field value as a string.
// Returns empty string if value is nil.
func (fv FieldValue) String() string {
	if fv.raw == nil {
		return ""
	}
	return fmt.Sprintf("%v", fv.raw)
}

// IsZero returns true if the field value is nil.
func (fv FieldValue) IsZero() bool {
	return fv.raw == nil
}

// CustomField represents a user-defined custom field configuration.
// This is a value object that defines how a custom field should behave.
type CustomField struct {
	// Name is the internal field name (e.g., "dev_assignment")
	Name string

	// DisplayName is the human-readable display name (e.g., "Dev Assignment")
	DisplayName string

	// Source identifies where the field value comes from (e.g., "labels", "custom_field_10001")
	Source string

	// Condition is an optional DSL expression for deriving the value (e.g., "has-label('dev1','dev2')")
	Condition string

	// DefaultValue is the default value when condition doesn't match or source is empty
	DefaultValue string

	// ValidValues is a whitelist of allowed values (empty means no validation)
	ValidValues []string

	// SyncDirection determines how this field is synchronized
	SyncDirection SyncDirection
}

// NewCustomField creates a new CustomField with required fields.
func NewCustomField(name, displayName, source string, syncDirection SyncDirection) (*CustomField, error) {
	cf := &CustomField{
		Name:          strings.TrimSpace(name),
		DisplayName:   strings.TrimSpace(displayName),
		Source:        strings.TrimSpace(source),
		SyncDirection: syncDirection,
		ValidValues:   make([]string, 0),
	}

	if err := cf.Validate(); err != nil {
		return nil, err
	}

	return cf, nil
}

// Validate checks if the custom field configuration is valid.
func (cf *CustomField) Validate() error {
	if strings.TrimSpace(cf.Name) == "" {
		return fmt.Errorf("%w: custom field name is required", ErrInvalidInput)
	}
	if strings.TrimSpace(cf.DisplayName) == "" {
		return fmt.Errorf("%w: custom field display name is required", ErrInvalidInput)
	}
	if strings.TrimSpace(cf.Source) == "" {
		return fmt.Errorf("%w: custom field source is required", ErrInvalidInput)
	}

	// Validate sync direction
	switch cf.SyncDirection {
	case SyncBidirectional, SyncJiraToLocal, SyncLocalOnly:
		// Valid
	default:
		return fmt.Errorf("%w: invalid sync direction: %s", ErrInvalidInput, cf.SyncDirection)
	}

	return nil
}

// ValidateValue checks if a value is in the ValidValues whitelist.
// Returns nil if ValidValues is empty (no validation) or if value is in the list.
func (cf *CustomField) ValidateValue(value string) error {
	if len(cf.ValidValues) == 0 {
		return nil // No validation required
	}

	for _, valid := range cf.ValidValues {
		if value == valid {
			return nil
		}
	}

	return fmt.Errorf("%w: value '%s' not in valid values for field '%s': %v",
		ErrInvalidFieldValue, value, cf.Name, cf.ValidValues)
}

// IsBidirectional returns true if this field syncs in both directions.
func (cf *CustomField) IsBidirectional() bool {
	return cf.SyncDirection == SyncBidirectional
}

// IsDerived returns true if this field uses a condition to derive its value.
func (cf *CustomField) IsDerived() bool {
	return strings.TrimSpace(cf.Condition) != ""
}

// DerivedField represents a field whose value is computed from other fields.
// For MVP, this primarily supports the has-label() DSL condition.
type DerivedField struct {
	// CustomField is the underlying field configuration
	CustomField *CustomField

	// MatchedValue is the value that matched the condition (if any)
	MatchedValue string

	// UsedDefault indicates whether the default value was used
	UsedDefault bool
}

// NewDerivedField creates a new DerivedField from a CustomField.
func NewDerivedField(cf *CustomField) *DerivedField {
	if cf == nil {
		return nil
	}
	return &DerivedField{
		CustomField:  cf,
		UsedDefault:  false,
		MatchedValue: "",
	}
}

// SetMatchedValue sets the matched value from the condition evaluation.
func (df *DerivedField) SetMatchedValue(value string) {
	df.MatchedValue = value
	df.UsedDefault = false
}

// SetDefault marks that the default value should be used.
func (df *DerivedField) SetDefault() {
	df.MatchedValue = df.CustomField.DefaultValue
	df.UsedDefault = true
}

// Value returns the final computed value (either matched or default).
func (df *DerivedField) Value() string {
	if df.MatchedValue != "" {
		return df.MatchedValue
	}
	return df.CustomField.DefaultValue
}
