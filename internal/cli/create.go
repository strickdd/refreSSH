package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
	"github.com/strickdd/refressh/internal/daemon"
)

var createCmd = &cobra.Command{
	Use:   "create [session-id] [command] [args...]",
	Short: "Create a new terminal session",
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
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
		} else {
			fmt.Printf("Error creating session: %s\n", resp.Status)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
