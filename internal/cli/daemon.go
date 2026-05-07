package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/daemon"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the refreSSH daemon",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the refreSSH daemon",
	Run: func(cmd *cobra.Command, args []string) {
		d := daemon.New()
		if err := d.Start(); err != nil {
			fmt.Printf("Error starting daemon: %v\n", err)
		}
	},
}

func init() {
	daemonCmd.AddCommand(startCmd)
	rootCmd.AddCommand(daemonCmd)
}
