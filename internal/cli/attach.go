package cli

import (
	"fmt"
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

		u := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%d", cfg.Port), Path: "/attach", RawQuery: "id=" + sessionID}

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
				_, message, err := c.ReadMessage()
				if err != nil {
					return
				}
				// Output is usually binary PTY data
				_, err = os.Stdout.Write(message)
				if err != nil {
					return
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
