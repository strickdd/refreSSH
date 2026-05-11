package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
	"golang.org/x/term"
)

var attachCmd = &cobra.Command{
	Use:   "attach [session-id]",
	Short: "Attach to an existing session",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		sessionID := args[0]
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			return
		}

		token, err := config.GetAPIToken()
		if err != nil {
			fmt.Printf("Error loading API token: %v\n", err)
			return
		}

		params := url.Values{}
		params.Add("id", sessionID)
		u := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%d", cfg.Port), Path: "/attach", RawQuery: params.Encode()}

		header := http.Header{}
		header.Add("Authorization", "Bearer "+token)

		c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
		if err != nil {
			fmt.Printf("Error connecting to daemon: %v\n", err)
			return
		}
		defer c.Close() //nolint:errcheck

		// Set terminal to raw mode
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Printf("Error setting raw mode: %v\n", err)
			return
		}
		defer func() {
			if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
				fmt.Printf("\nError restoring terminal: %v\n", err)
			}
		}()

		// Channel to catch interrupt signals
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

		// Goroutine to read from WebSocket and write to stdout
		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				msgType, message, err := c.ReadMessage()
				if err != nil {
					return
				}
				
				if msgType == websocket.TextMessage {
					var status struct {
						IsPrimary bool `json:"is_primary"`
					}
					if err := json.Unmarshal(message, &status); err == nil {
						// Update terminal window title
						title := fmt.Sprintf("\033]0;[View Only] refreSSH: %s (Ctrl+Space to request control)\007", sessionID)
						if status.IsPrimary {
							title = fmt.Sprintf("\033]0;[Primary] refreSSH: %s\007", sessionID)
						}
						_, _ = os.Stdout.Write([]byte(title)) //nolint:errcheck
					}
				} else {
					_, err = os.Stdout.Write(message)
					if err != nil {
						return
					}
				}
			}
		}()

		// Goroutine to read from stdin and write to WebSocket
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					// Check for Ctrl+Space (0x00) to request primary control
					if n == 1 && buf[0] == 0x00 {
						req := map[string]string{"action": "request_primary"}
						reqBytes, _ := json.Marshal(req)
						_ = c.WriteMessage(websocket.TextMessage, reqBytes) //nolint:errcheck
						continue
					}

					err = c.WriteMessage(websocket.BinaryMessage, buf[:n])
					if err != nil {
						return
					}
				}
			}
		}()

		select {
		case <-done:
			// Session closed or connection lost
		case <-interrupt:
			// User pressed Ctrl+C (though in raw mode it might be captured and sent to PTY)
		}
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
