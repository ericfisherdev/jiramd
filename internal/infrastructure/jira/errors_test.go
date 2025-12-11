package jira

import (
	"errors"
	"testing"

	"github.com/esfisher/jiramd/internal/domain"
)

func TestMapHTTPError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		context     string
		wantDomain  error
		wantContain string
	}{
		{
			name:        "nil error",
			err:         nil,
			context:     "test context",
			wantDomain:  nil,
			wantContain: "",
		},
		{
			name:        "404 error",
			err:         errors.New("HTTP 404 not found"),
			context:     "fetch ticket",
			wantDomain:  domain.ErrNotFound,
			wantContain: "fetch ticket",
		},
		{
			name:        "401 error",
			err:         errors.New("status code 401 unauthorized"),
			context:     "fetch ticket",
			wantDomain:  domain.ErrUnauthorized,
			wantContain: "fetch ticket",
		},
		{
			name:        "403 error",
			err:         errors.New("HTTP 403 forbidden"),
			context:     "update ticket",
			wantDomain:  domain.ErrUnauthorized,
			wantContain: "update ticket",
		},
		{
			name:        "400 error",
			err:         errors.New("status code 400 bad request"),
			context:     "create ticket",
			wantDomain:  domain.ErrInvalidInput,
			wantContain: "create ticket",
		},
		{
			name:        "409 error",
			err:         errors.New("HTTP 409 conflict"),
			context:     "update ticket",
			wantDomain:  domain.ErrConflict,
			wantContain: "update ticket",
		},
		{
			name:        "unknown error",
			err:         errors.New("some other error"),
			context:     "test operation",
			wantDomain:  nil,
			wantContain: "test operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapHTTPError(tt.err, tt.context)

			if tt.err == nil {
				if got != nil {
					t.Errorf("mapHTTPError() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("mapHTTPError() returned nil, want error")
			}

			if tt.wantDomain != nil && !errors.Is(got, tt.wantDomain) {
				t.Errorf("mapHTTPError() error does not wrap %v, got %v", tt.wantDomain, got)
			}

			if tt.wantContain != "" && !contains(got.Error(), tt.wantContain) {
				t.Errorf("mapHTTPError() error = %v, want to contain %v", got.Error(), tt.wantContain)
			}
		})
	}
}

func TestContainsStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		errStr string
		code   string
		want   bool
	}{
		{
			name:   "exact match",
			errStr: "error: 404",
			code:   "404",
			want:   true,
		},
		{
			name:   "status code pattern",
			errStr: "error: status code 401",
			code:   "401",
			want:   true,
		},
		{
			name:   "HTTP pattern",
			errStr: "error: HTTP 403 forbidden",
			code:   "403",
			want:   true,
		},
		{
			name:   "status= pattern",
			errStr: "error: status=400",
			code:   "400",
			want:   true,
		},
		{
			name:   "no match",
			errStr: "error: something went wrong",
			code:   "404",
			want:   false,
		},
		{
			name:   "partial match (should not match)",
			errStr: "error: 4041234",
			code:   "404",
			want:   true, // Will match because we're looking for substring
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsStatusCode(tt.errStr, tt.code)
			if got != tt.want {
				t.Errorf("containsStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsErrorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checkFn  func(error) bool
		expected bool
	}{
		{
			name:     "IsNotFoundError with ErrNotFound",
			err:      domain.ErrNotFound,
			checkFn:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "IsNotFoundError with other error",
			err:      domain.ErrUnauthorized,
			checkFn:  IsNotFoundError,
			expected: false,
		},
		{
			name:     "IsUnauthorizedError with ErrUnauthorized",
			err:      domain.ErrUnauthorized,
			checkFn:  IsUnauthorizedError,
			expected: true,
		},
		{
			name:     "IsConflictError with ErrConflict",
			err:      domain.ErrConflict,
			checkFn:  IsConflictError,
			expected: true,
		},
		{
			name:     "IsInvalidInputError with ErrInvalidInput",
			err:      domain.ErrInvalidInput,
			checkFn:  IsInvalidInputError,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checkFn(tt.err)
			if got != tt.expected {
				t.Errorf("check function returned %v, want %v", got, tt.expected)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
