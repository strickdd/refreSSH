package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
)

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Manage isolated containerized Sandbox deployments",
}

var sandboxStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the refreSSH Sandbox container",
	RunE: func(_ *cobra.Command, _ []string) error {
		configDir, err := config.GetConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config dir: %w", err)
		}

		sandboxDir := filepath.Join(configDir, "sandbox")
		if err := os.MkdirAll(sandboxDir, 0700); err != nil {
			return fmt.Errorf("failed to create sandbox directory: %w", err)
		}

		// Write Dockerfile
		dockerfileContent := `FROM golang:1.24-alpine
RUN apk add --no-cache bash curl git openssh-client ca-certificates
RUN go install github.com/strickdd/refressh/cmd/refressh@latest
EXPOSE 9000
ENTRYPOINT ["/go/bin/refressh", "daemon", "start"]
`
		dockerfilePath := filepath.Join(sandboxDir, "Dockerfile")
		if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0600); err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}

		// Write docker-compose.yml
		composeContent := `services:
  sandbox:
    build: .
    container_name: refressh-sandbox
    environment:
      - REFRESSH_API_LISTEN=0.0.0.0:9000
    ports:
      - "127.0.0.1:9000:9000"
    restart: unless-stopped
    volumes:
      - refressh-data:/root/.refreSSH

volumes:
  refressh-data:
`
		composePath := filepath.Join(sandboxDir, "docker-compose.yml")
		if err := os.WriteFile(composePath, []byte(composeContent), 0600); err != nil {
			return fmt.Errorf("failed to write docker-compose.yml: %w", err)
		}

		fmt.Println("Starting refreSSH Sandbox...")
		//nolint:gosec // execution of docker is intentional for sandbox management
		cmd := exec.Command("docker", "compose", "-f", composePath, "up", "-d", "--build")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start sandbox container (is docker installed and running?): %w", err)
		}

		fmt.Println("\nSandbox started successfully!")
		fmt.Println("You can now connect to it by pointing your CLI or UI to 127.0.0.1:9000")
		return nil
	},
}

var sandboxStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the refreSSH Sandbox container",
	RunE: func(_ *cobra.Command, _ []string) error {
		configDir, err := config.GetConfigDir()
		if err != nil {
			return err
		}

		composePath := filepath.Join(configDir, "sandbox", "docker-compose.yml")
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			fmt.Println("Sandbox is not currently configured or running.")
			return nil
		}

		fmt.Println("Stopping refreSSH Sandbox...")
		//nolint:gosec // execution of docker is intentional for sandbox management
		cmd := exec.Command("docker", "compose", "-f", composePath, "down")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop sandbox: %w", err)
		}

		fmt.Println("Sandbox stopped.")
		return nil
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxStartCmd)
	sandboxCmd.AddCommand(sandboxStopCmd)
	rootCmd.AddCommand(sandboxCmd)
}
