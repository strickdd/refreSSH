package cli

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
)

var stopCmd = &cobra.Command{
	Use:   "stop [session-id]",
	Short: "Stop and terminate a session",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		sessionID := args[0]
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		apiURL := fmt.Sprintf("http://127.0.0.1:%d/sessions/%s", cfg.Port, url.PathEscape(sessionID))

		token, err := config.GetAPIToken()
		if err != nil {
			fmt.Printf("Error loading API token: %v\n", err)
			return
		}

		req, err := http.NewRequest(http.MethodDelete, apiURL, nil) //nolint:gosec
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error connecting to daemon. Is it running?")
			return
		}
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode == http.StatusNoContent {
			fmt.Printf("Session '%s' stopped.\n", sessionID)
		} else if resp.StatusCode == http.StatusNotFound {
			fmt.Printf("Session '%s' not found.\n", sessionID)
		} else {
			fmt.Printf("Error stopping session: %s\n", resp.Status)
		}
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
