// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// projectKeyPattern defines the valid format for Jira project keys (2-10 uppercase letters/numbers)
var projectKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

// Project represents a Jira project entity.
// This is a core domain entity that represents a Jira project being synced.
type Project struct {
	// Key is the unique project key (e.g., "JMD", "PROJ")
	Key string

	// Name is the human-readable project name
	Name string

	// Description is the project description (optional)
	Description string

	// CustomFields contains project-specific custom field configurations
	CustomFields []*CustomField
}

// NewProject creates a new Project with required fields.
func NewProject(key, name string) (*Project, error) {
	p := &Project{
		Key:          strings.TrimSpace(strings.ToUpper(key)),
		Name:         strings.TrimSpace(name),
		Description:  "",
		CustomFields: make([]*CustomField, 0),
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}

	return p, nil
}

// Validate checks if the project has all required fields populated correctly.
func (p *Project) Validate() error {
	if strings.TrimSpace(p.Key) == "" {
		return fmt.Errorf("%w: project key is required", ErrEmptyKey)
	}

	if !projectKeyPattern.MatchString(p.Key) {
		return fmt.Errorf("%w: project key '%s' (expected format: 2-10 uppercase letters/numbers)", ErrInvalidProject, p.Key)
	}

	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("%w: project name is required", ErrInvalidProject)
	}

	return nil
}

// AddCustomField adds a custom field configuration to this project.
// Returns an error if a field with the same name already exists.
func (p *Project) AddCustomField(field *CustomField) error {
	if field == nil {
		return fmt.Errorf("%w: custom field cannot be nil", ErrInvalidInput)
	}

	if err := field.Validate(); err != nil {
		return fmt.Errorf("invalid custom field: %w", err)
	}

	// Check for duplicate field names
	for _, existing := range p.CustomFields {
		if existing.Name == field.Name {
			return fmt.Errorf("%w: custom field '%s' already exists", ErrInvalidInput, field.Name)
		}
	}

	p.CustomFields = append(p.CustomFields, field)
	return nil
}

// GetCustomField retrieves a custom field by name.
// Returns nil if the field doesn't exist.
func (p *Project) GetCustomField(name string) *CustomField {
	for _, field := range p.CustomFields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

// RemoveCustomField removes a custom field by name.
// Returns true if the field was found and removed.
func (p *Project) RemoveCustomField(name string) bool {
	for i, field := range p.CustomFields {
		if field.Name == name {
			// Remove by slicing
			p.CustomFields = append(p.CustomFields[:i], p.CustomFields[i+1:]...)
			return true
		}
	}
	return false
}

// BidirectionalFields returns all custom fields configured for bidirectional sync.
func (p *Project) BidirectionalFields() []*CustomField {
	result := make([]*CustomField, 0)
	for _, field := range p.CustomFields {
		if field.IsBidirectional() {
			result = append(result, field)
		}
	}
	return result
}

// DerivedFields returns all custom fields that use derived conditions.
func (p *Project) DerivedFields() []*CustomField {
	result := make([]*CustomField, 0)
	for _, field := range p.CustomFields {
		if field.IsDerived() {
			result = append(result, field)
		}
	}
	return result
}
