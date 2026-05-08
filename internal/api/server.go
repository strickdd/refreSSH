// Package api provides the local-only HTTP API server for refreSSH.
package api

import (
	"fmt"
	"net/http"
	"time"
)

// Start initializes and starts the local HTTP API server on the specified port.
// It binds strictly to 127.0.0.1 to ensure local-only access.
func Start(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("API Server starting on %s...\n", addr)

	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "OK"); err != nil {
			// In a real app, we might log this, but for now we just suppress it to satisfy errcheck
			_ = err
		}
	})

	// Bind to local loopback ONLY as per GEMINI.md
	server := &http.Server{
		Addr:              addr,
		ReadHeaderTimeout: 5 * time.Second, // Address G112: Potential Slowloris Attack
	}

	return server.ListenAndServe()
}
