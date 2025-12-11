// Package jira provides Jira API client implementation.
// This infrastructure layer implements integration with Jira Cloud API.
package jira

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/esfisher/jiramd/internal/domain"
)

// mapHTTPError maps HTTP status codes to domain errors.
// This function implements the error translation layer between infrastructure and domain.
//
// HTTP Status Code Mappings:
//   - 404 Not Found -> domain.ErrNotFound
//   - 401 Unauthorized -> domain.ErrUnauthorized
//   - 403 Forbidden -> domain.ErrUnauthorized
//   - 400 Bad Request -> domain.ErrInvalidInput
//   - 409 Conflict -> domain.ErrConflict
//   - Other errors -> wrapped original error
func mapHTTPError(err error, context string) error {
	if err == nil {
		return nil
	}

	// Try to extract HTTP status code from the error
	var statusCode int

	// Check if it's a Jira API error that wraps an HTTP response
	// The go-jira library typically returns errors that can be type-asserted
	// We'll check for common patterns in the error string
	errStr := err.Error()

	// Common Jira API error patterns
	switch {
	case containsStatusCode(errStr, "404"):
		statusCode = http.StatusNotFound
	case containsStatusCode(errStr, "401"):
		statusCode = http.StatusUnauthorized
	case containsStatusCode(errStr, "403"):
		statusCode = http.StatusForbidden
	case containsStatusCode(errStr, "400"):
		statusCode = http.StatusBadRequest
	case containsStatusCode(errStr, "409"):
		statusCode = http.StatusConflict
	default:
		// Unknown error, wrap and return
		return fmt.Errorf("%s: %w", context, err)
	}

	// Map HTTP status codes to domain errors
	switch statusCode {
	case http.StatusNotFound:
		return fmt.Errorf("%s: %w", context, domain.ErrNotFound)
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%s: %w", context, domain.ErrUnauthorized)
	case http.StatusBadRequest:
		return fmt.Errorf("%s: %w", context, domain.ErrInvalidInput)
	case http.StatusConflict:
		return fmt.Errorf("%s: %w", context, domain.ErrConflict)
	default:
		return fmt.Errorf("%s: %w", context, err)
	}
}

// containsStatusCode checks if an error string contains a specific HTTP status code.
// This is a helper function for parsing error messages from the Jira API client.
func containsStatusCode(errStr, code string) bool {
	// Look for patterns like "404" or "status code 404" or "HTTP 404"
	patterns := []string{
		code,
		"status code " + code,
		"HTTP " + code,
		"status=" + code,
	}

	for _, pattern := range patterns {
		if len(errStr) >= len(pattern) {
			for i := 0; i <= len(errStr)-len(pattern); i++ {
				if errStr[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}

// IsNotFoundError checks if an error is a "not found" error.
func IsNotFoundError(err error) bool {
	return errors.Is(err, domain.ErrNotFound)
}

// IsUnauthorizedError checks if an error is an "unauthorized" error.
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, domain.ErrUnauthorized)
}

// IsConflictError checks if an error is a "conflict" error.
func IsConflictError(err error) bool {
	return errors.Is(err, domain.ErrConflict)
}

// IsInvalidInputError checks if an error is an "invalid input" error.
func IsInvalidInputError(err error) bool {
	return errors.Is(err, domain.ErrInvalidInput)
}
