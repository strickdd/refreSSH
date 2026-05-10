package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/api"
	"github.com/strickdd/refressh/internal/config"
	"github.com/strickdd/refressh/internal/daemon"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the refreSSH daemon",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the refreSSH daemon",
	Run: func(_ *cobra.Command, _ []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		d := daemon.New(cfg)
		if err := d.Start(); err != nil {
			fmt.Printf("Error starting daemon: %v\n", err)
			return
		}

		if err := api.Start(cfg.Port, d); err != nil {
			fmt.Printf("Error starting API: %v\n", err)
		}
	},
}

func init() {
	daemonCmd.AddCommand(startCmd)
	rootCmd.AddCommand(daemonCmd)
}
