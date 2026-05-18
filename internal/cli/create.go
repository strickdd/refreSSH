package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
	"github.com/strickdd/refressh/internal/daemon"
)

var createCmd = &cobra.Command{
	Use:   "create [session-id] [command] [args...]",
	Short: "Create a new terminal session",
	Long:  "Create a new terminal session. By default, attaches to the session after creation. Use --no-attach to create without entering.",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		noAttach, _ := cmd.Flags().GetBool("no-attach")
		sessionID := args[0]
		command := args[1]
		cmdArgs := args[2:]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		reqBody, _ := json.Marshal(map[string]interface{}{
			"id":      sessionID,
			"command": command,
			"args":    cmdArgs,
		})

		url := fmt.Sprintf("http://127.0.0.1:%d/sessions", cfg.Port)

		token, err := config.GetAPIToken()
		if err != nil {
			fmt.Printf("Error loading API token: %v\n", err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error connecting to daemon. Is it running?")
			return
		}
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode == http.StatusCreated {
			var s daemon.Session
			if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
				fmt.Printf("Session created, but failed to decode response: %v\n", err)
				return
			}
			fmt.Printf("Session '%s' created successfully.\n", s.ID)

			if !noAttach {
				fmt.Printf("Attaching to session '%s'...\n", s.ID)
				executable, err := os.Executable()
				if err != nil {
					fmt.Printf("Error getting executable path: %v\n", err)
					return
				}
				execCmd := exec.Command(executable, "attach", sessionID) //nolint:gosec
				execCmd.Stdin = os.Stdin
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				_ = execCmd.Run()
			}
		} else {
			fmt.Printf("Error creating session: %s\n", resp.Status)
		}
	},
}

func init() {
	createCmd.Flags().Bool("no-attach", false, "Create session without attaching to it")
	rootCmd.AddCommand(createCmd)
}
