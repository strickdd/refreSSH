package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of refreSSH",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("refreSSH %s\n", config.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
