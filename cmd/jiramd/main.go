// Package main is the entry point for the jiramd CLI.
// This layer can depend on application and infrastructure layers to wire up the application.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// version is set at build time using ldflags
	version = "dev"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jiramd",
	Short: "jiramd - Jira markdown sync daemon",
	Long: `jiramd is a background daemon that bidirectionally syncs Jira Cloud tickets
to local markdown files. It eliminates AI token usage for Jira ticket
management by maintaining a local markdown cache.`,
	Version: version,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Register subcommands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(fieldCmd)
	rootCmd.AddCommand(statusCmd)

	// Global flags can be added here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jiramd.yaml)")
}
