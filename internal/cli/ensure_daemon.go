package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/strickdd/refressh/internal/config"
)

// ensureDaemonRunning checks if the refreSSH daemon is responsive and starts it if not.
func ensureDaemonRunning() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 1. Check if daemon is already running
	if isDaemonHealthy(cfg.Port) {
		return nil
	}

	// 2. Not running, start it detached
	fmt.Println("refreSSH daemon is not running. Starting it now...")

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := exec.Command(executable, "start") //nolint:gosec

	// Detach the process from the current terminal
	if err := startDetached(cmd); err != nil {
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// 3. Wait for it to become healthy (up to 5 seconds)
	for i := 0; i < 50; i++ {
		if isDaemonHealthy(cfg.Port) {
			fmt.Println("Daemon started successfully.")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("daemon started but failed to become responsive within 5 seconds")
}

func isDaemonHealthy(port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := http.Client{
		Timeout: 500 * time.Millisecond,
	}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close() //nolint:errcheck
	return resp.StatusCode == http.StatusOK
}
