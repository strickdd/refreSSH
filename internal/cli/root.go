package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "refressh",
	Short: "refreSSH - Persistent terminal session manager",
	Long: `refreSSH is a user-level background daemon designed to host and manage 
persistent terminal sessions, optimized for AI CLI agents.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Root flags can be added here
}
