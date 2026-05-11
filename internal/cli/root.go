// Package cli implements the command-line interface for refreSSH.
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refressh",
	Short: "refreSSH - Persistent terminal session manager",
	Long: `refreSSH is a user-level background daemon designed to host and manage 
persistent terminal sessions, optimized for AI CLI agents.`,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// Skip auto-start for 'daemon' commands
		if cmd.Name() == "start" || cmd.Parent().Name() == "daemon" || cmd.Name() == "daemon" {
			return nil
		}
		return ensureDaemonRunning()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Root flags can be added here
}
