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

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/sessions", cfg.Port))
		if err != nil {
			fmt.Println("Error connecting to daemon. Is it running?")
			return
		}
		defer resp.Body.Close()

		var sessions []*daemon.Session
		if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
			fmt.Printf("Error decoding response: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No active sessions.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tCOMMAND\tSTATUS\tSTARTED")
		for _, s := range sessions {
			status := "stopped"
			if s.Running {
				status = "running"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", s.ID, s.Command, status, s.StartTime.Format(time.RFC822))
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
