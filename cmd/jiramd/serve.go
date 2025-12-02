package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the jiramd daemon",
	Long: `Start the jiramd daemon which watches for changes to markdown files
and Jira tickets, synchronizing them bidirectionally.

The daemon will:
  - Watch local markdown files for changes
  - Poll Jira for ticket updates
  - Synchronize changes bidirectionally
  - Maintain conflict resolution state`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement serve command
		fmt.Println("serve command not yet implemented")
	},
}

func init() {
	// Add flags specific to serve command
	// serveCmd.Flags().StringP("config", "c", "", "Path to config file")
	// serveCmd.Flags().IntP("poll-interval", "p", 60, "Jira poll interval in seconds")
}
