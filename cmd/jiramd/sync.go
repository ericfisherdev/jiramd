package main

import (
	"fmt"

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
  - Testing synchronization logic`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement sync command
		fmt.Println("sync command not yet implemented")
	},
}

func init() {
	// Add flags specific to sync command
	// syncCmd.Flags().StringP("direction", "d", "both", "Sync direction: 'to-jira', 'from-jira', or 'both'")
	// syncCmd.Flags().StringP("project", "p", "", "Limit sync to specific project key")
}
