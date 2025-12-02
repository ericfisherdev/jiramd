// Package field contains use cases for field mapping and management operations.
// This layer handles custom field mapping configuration and validation.
package field

import (
	"context"
	"errors"
)

// Service handles field mapping use cases.
// It manages the mapping between Jira custom fields and markdown representation.
type Service struct {
	// TODO: Add dependencies for field mapping storage
}

// NewService creates a new field service.
func NewService() *Service {
	return &Service{}
}

// GetMapping retrieves the field mapping for a project.
// This is a placeholder for the actual implementation.
func (s *Service) GetMapping(ctx context.Context, projectKey string) (map[string]string, error) {
	// TODO: Implement field mapping retrieval
	return nil, errors.New("field.Service.GetMapping not implemented")
}

// SetMapping sets the field mapping for a project.
// This is a placeholder for the actual implementation.
func (s *Service) SetMapping(ctx context.Context, projectKey string, mapping map[string]string) error {
	// TODO: Implement field mapping storage
	return errors.New("field.Service.SetMapping not implemented")
}
