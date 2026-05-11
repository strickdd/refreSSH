package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
	"github.com/strickdd/refressh/internal/daemon"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active sessions",
	Run: func(_ *cobra.Command, _ []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		url := fmt.Sprintf("http://127.0.0.1:%d/sessions", cfg.Port)
		resp, err := http.Get(url) //nolint:gosec
		if err != nil {
			fmt.Println("Error connecting to daemon. Is it running?")
			return
		}
		defer resp.Body.Close() //nolint:errcheck

		var sessions []daemon.Session
		if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No active sessions.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tCOMMAND\tSTATUS\tSTARTED") //nolint:errcheck
		for i := range sessions {
			s := &sessions[i]
			status := "stopped"
			if s.Running {
				status = "running"
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.Command, status, s.StartTime.Format(time.RFC822)) //nolint:errcheck
		}
		_ = w.Flush() //nolint:errcheck
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
