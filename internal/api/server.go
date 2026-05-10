// Package api provides the local-only HTTP API server for refreSSH.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/strickdd/refressh/internal/api/ws"
	"github.com/strickdd/refressh/internal/daemon"
)

// Start initializes and starts the local HTTP API server on the specified port.
// It binds strictly to 127.0.0.1 to ensure local-only access.
func Start(port int, d *daemon.Daemon) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("API Server starting on %s...\n", addr)

	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			// Suppress error to satisfy errcheck
			_ = err
		}
	})

	// Session list endpoint
	http.HandleFunc("/sessions", func(w http.ResponseWriter, _ *http.Request) {
		sessions := d.Sessions()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// WebSocket attachment endpoint
	http.HandleFunc("/attach", ws.Handler(d.Broadcaster()))

	// Bind to local loopback ONLY as per GEMINI.md
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second, // Address G112: Potential Slowloris Attack
	}

	return server.ListenAndServe()
}
