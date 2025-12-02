package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// projectCmd represents the project command
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage Jira project configurations",
	Long: `Manage Jira project configurations for synchronization.

Subcommands allow you to:
  - List configured projects
  - Add a new project to sync
  - Remove a project from sync
  - View project details`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement project command or show help
		fmt.Println("project command not yet implemented")
		cmd.Help()
	},
}

func init() {
	// Add subcommands for project management
	// projectCmd.AddCommand(projectListCmd)
	// projectCmd.AddCommand(projectAddCmd)
	// projectCmd.AddCommand(projectRemoveCmd)
}
