// Package main is the entry point for the jiramd CLI.
// This layer can depend on application and infrastructure layers to wire up the application.
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// run is the main application logic.
// This is a placeholder for the actual implementation.
func run() error {
	// TODO: Implement CLI command routing
	// TODO: Wire up dependencies (repositories, services, etc.)
	// TODO: Handle commands: serve, sync, project, field, status
	fmt.Println("jiramd - Jira markdown sync daemon")
	return nil
}
