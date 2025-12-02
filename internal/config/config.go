// Package config handles application configuration loading and validation.
// This package provides a convenient wrapper around the infrastructure config implementation.
package config

import (
	"github.com/esfisher/jiramd/internal/domain"
	infraConfig "github.com/esfisher/jiramd/internal/infrastructure/config"
)

// Load loads and validates configuration from a YAML file.
// This is a convenience function that uses the infrastructure implementation.
// It performs the following operations:
// 1. Loads configuration from YAML file
// 2. Expands environment variables
// 3. Validates configuration
// Returns domain.Config and error if loading or validation fails.
func Load(path string) (*domain.Config, error) {
	// Create loader and validator
	loader := infraConfig.NewLoader()
	validator := infraConfig.NewValidator()

	// Load configuration
	cfg, err := loader.Load(path)
	if err != nil {
		return nil, err
	}

	// Validate configuration
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
