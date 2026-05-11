package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/strickdd/refressh/internal/config"
	"golang.org/x/term"
)

// ringBuffer is a simple thread-safe byte ring buffer to keep local scrollback history.
type ringBuffer struct {
	buf []byte
	mu  sync.Mutex
	max int
}

func newRingBuffer(max int) *ringBuffer {
	return &ringBuffer{
		buf: make([]byte, 0, max),
		max: max,
	}
}

func (r *ringBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf = append(r.buf, p...)
	if len(r.buf) > r.max {
		r.buf = r.buf[len(r.buf)-r.max:]
	}
	return len(p), nil
}

func (r *ringBuffer) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]byte, len(r.buf))
	copy(res, r.buf)
	return res
}

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

		fd := int(os.Stdin.Fd())
		// Set terminal to raw mode
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			fmt.Printf("Error setting raw mode: %v\n", err)
			return
		}
		
		restoreRaw := func() {
			_ = term.Restore(fd, oldState)
		}
		defer restoreRaw()

		// Local scrollback buffer (up to 1MB)
		localScrollback := newRingBuffer(1024 * 1024)

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
					_, _ = localScrollback.Write(message)
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
			inCommandMode := false

			for {
				n, err := os.Stdin.Read(buf)
				if err != nil {
					return
				}
				if n > 0 {
					// Handle Command Mode State Machine
					if inCommandMode {
						inCommandMode = false
						if n == 1 {
							switch buf[0] {
							case 'd', 'D': // Detach
								_, _ = os.Stdout.Write([]byte("\r\n[Detached from session]\r\n")) //nolint:errcheck
								close(done)
								return
							case 's', 'S': // Scrollback Search via Pager
								restoreRaw()
								
								// Create temp file
								tmpFile, err := os.CreateTemp("", "refressh-scrollback-*.txt")
								if err == nil {
									// Strip basic ANSI escapes (optional, but raw might be messy in simple pagers)
									// For now, write raw so 'less -R' works
									_, _ = io.Copy(tmpFile, bytes.NewReader(localScrollback.Bytes()))
									tmpFile.Close()

									pager := "less"
									args := []string{"-R", tmpFile.Name()}
									if runtime.GOOS == "windows" {
										pager = "more"
										args = []string{tmpFile.Name()}
									}
									if p := os.Getenv("PAGER"); p != "" {
										pager = p
										args = []string{tmpFile.Name()}
									}

									cmd := exec.Command(pager, args...)
									cmd.Stdin = os.Stdin
									cmd.Stdout = os.Stdout
									cmd.Stderr = os.Stderr
									_ = cmd.Run() // Wait for pager to finish
									
									_ = os.Remove(tmpFile.Name())
								}
								
								_, _ = term.MakeRaw(fd) // Re-enter raw mode
								continue
							case 0x02: // Ctrl+B again sends a literal Ctrl+B
								err = c.WriteMessage(websocket.BinaryMessage, []byte{0x02})
								if err != nil {
									return
								}
							default:
								// Exit command mode, do nothing
								_, _ = os.Stdout.Write([]byte("\r\n[Command mode exited]\r\n")) //nolint:errcheck
							}
						}
						continue
					}

					// Terminal Mode
					if n == 1 {
						if buf[0] == 0x02 { // Ctrl+B
							inCommandMode = true
							_, _ = os.Stdout.Write([]byte("\r\n[Command Mode: 'd' to detach, 's' to search scrollback, 'Ctrl+B' to send literal]\r\n")) //nolint:errcheck
							continue
						}
						if buf[0] == 0x00 { // Ctrl+Space
							req := map[string]string{"action": "request_primary"}
							reqBytes, _ := json.Marshal(req)
							_ = c.WriteMessage(websocket.TextMessage, reqBytes) //nolint:errcheck
							continue
						}
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
