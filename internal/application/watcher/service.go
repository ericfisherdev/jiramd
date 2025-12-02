// Package watcher contains use cases for file system watching operations.
// This layer orchestrates file system monitoring and sync triggering.
package watcher

import (
	"context"
	"errors"
)

// Service handles file system watching use cases.
// It monitors markdown file changes and triggers appropriate synchronization.
//
// TODO: Follow dependency injection pattern - update NewService to accept and store
// dependencies (file.Watcher interface, sync service) rather than constructing them internally.
// Document Watch/Stop semantics (idempotency, error handling, context cancellation).
type Service struct {
	// TODO: Add dependencies for file watching and sync triggering
}

// NewService creates a new watcher service.
func NewService() *Service {
	return &Service{}
}

// Watch starts watching the specified directory for changes.
// This is a placeholder for the actual implementation.
func (s *Service) Watch(ctx context.Context, dir string) error {
	// TODO: Implement file watching logic
	return errors.New("watcher.Service.Watch not implemented")
}

// Stop stops the file watcher.
// This is a placeholder for the actual implementation.
func (s *Service) Stop() error {
	// TODO: Implement stop logic
	return errors.New("watcher.Service.Stop not implemented")
}
