package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show synchronization status",
	Long: `Show the current synchronization status between local markdown files
and Jira tickets.

Displays:
  - Last sync timestamp
  - Number of tickets synchronized
  - Any pending changes or conflicts
  - Daemon running status`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement status command
		fmt.Println("status command not yet implemented")
	},
}

func init() {
	// Add flags specific to status command
	// statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed status information")
	// statusCmd.Flags().StringP("project", "p", "", "Show status for specific project only")
}
