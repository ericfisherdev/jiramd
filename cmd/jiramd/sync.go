package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/esfisher/jiramd/internal/infrastructure/jira"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manually trigger a one-time sync",
	Long: `Manually trigger a one-time synchronization between local markdown files
and Jira tickets.

This is useful for:
  - Initial setup and data population
  - Forcing a sync without running the daemon
  - Testing synchronization logic

Required environment variables:
  JIRAMD_BASE_URL - Jira Cloud URL (e.g., https://yoursite.atlassian.net)
  JIRAMD_USER_EMAIL - User email for Jira authentication
  JIRAMD_API_TOKEN - API token from Jira Cloud`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))

		// Get environment variables for Jira authentication
		baseURL := os.Getenv("JIRAMD_BASE_URL")
		email := os.Getenv("JIRAMD_USER_EMAIL")
		token := os.Getenv("JIRAMD_API_TOKEN")

		if baseURL == "" || email == "" || token == "" {
			logger.Error("missing required environment variables",
				"JIRAMD_BASE_URL", baseURL != "",
				"JIRAMD_USER_EMAIL", email != "",
				"JIRAMD_API_TOKEN", token != "")
			fmt.Println("Error: Missing required environment variables")
			fmt.Println("Please set JIRAMD_BASE_URL, JIRAMD_USER_EMAIL, and JIRAMD_API_TOKEN")
			os.Exit(1)
		}

		// Create Jira client
		client, err := jira.NewClient(baseURL, email, token, logger)
		if err != nil {
			logger.Error("failed to create jira client", "error", err)
			fmt.Printf("Error creating Jira client: %v\n", err)
			os.Exit(1)
		}

		// Create Jira repository
		_ = jira.NewRepository(client)

		// TODO: Implement sync command with repository
		logger.Info("sync command not yet fully implemented")
		fmt.Println("sync command not yet implemented - Jira client initialized successfully")
	},
}

func init() {
	// Add flags specific to sync command
	// syncCmd.Flags().StringP("direction", "d", "both", "Sync direction: 'to-jira', 'from-jira', or 'both'")
	// syncCmd.Flags().StringP("project", "p", "", "Limit sync to specific project key")
}
