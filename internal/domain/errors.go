// Package domain contains the core business logic and entities.
// This layer has zero dependencies on application or infrastructure layers.
package domain

import "errors"

// Domain errors represent business rule violations and core domain concerns.
// These errors should be used by domain entities and checked by application layer.
var (
	// ErrNotFound indicates a requested entity was not found
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidInput indicates invalid input data
	ErrInvalidInput = errors.New("invalid input")

	// ErrConflict indicates a conflict in entity state
	ErrConflict = errors.New("entity conflict")

	// ErrUnauthorized indicates lack of authorization
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidTicketKey indicates a ticket key does not match expected format
	ErrInvalidTicketKey = errors.New("invalid ticket key format")

	// ErrInvalidFieldValue indicates a field value is invalid
	ErrInvalidFieldValue = errors.New("invalid field value")

	// ErrInvalidProject indicates a project entity is invalid
	ErrInvalidProject = errors.New("invalid project")

	// ErrSyncConflict indicates a sync conflict between local and remote
	ErrSyncConflict = errors.New("sync conflict detected")

	// ErrInvalidTimestamp indicates an invalid or zero timestamp
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrEmptyKey indicates an empty or whitespace-only key
	ErrEmptyKey = errors.New("empty key")

	// ErrInvalidOperation indicates an invalid pending operation type
	ErrInvalidOperation = errors.New("invalid operation type")
)
