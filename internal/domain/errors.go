// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import "errors"

// Domain errors represent business rule violations and core domain concerns.
var (
	// ErrNotFound indicates a requested entity was not found
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict indicates a conflict in entity state
	ErrConflict = errors.New("entity conflict")

	// ErrUnauthorized indicates lack of authorization
	ErrUnauthorized = errors.New("unauthorized")

	// ErrConfig indicates a configuration error
	ErrConfig = errors.New("configuration error")
)

// ConfigError represents a configuration-specific error with details.
type ConfigError struct {
	Message string
}

// Error implements the error interface.
func (e *ConfigError) Error() string {
	return e.Message
}

// NewConfigError creates a new ConfigError.
func NewConfigError(message string) error {
	return &ConfigError{Message: message}
}
