package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Open the refreSSH Web UI in your default browser",
	RunE: func(_ *cobra.Command, _ []string) error {
		// Ensure daemon is running
		if err := ensureDaemonRunning(); err != nil {
			return fmt.Errorf("failed to ensure daemon is running: %w", err)
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		token, err := config.GetAPIToken()
		if err != nil {
			return fmt.Errorf("failed to load API token: %w", err)
		}

		url := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", cfg.Port, token)
		fmt.Printf("Opening Web UI at %s\n", fmt.Sprintf("http://127.0.0.1:%d", cfg.Port))

		return openBrowser(url)
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
