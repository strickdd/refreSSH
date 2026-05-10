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

	// Bind to local loopback ONLY as per GEMINI.md
	server := &http.Server{
		Addr:              addr,
		Handler:           NewHandler(d),
		ReadHeaderTimeout: 5 * time.Second, // Address G112: Potential Slowloris Attack
	}

	return server.ListenAndServe()
}

// NewHandler creates and configures the API routes, returning an http.Handler.
func NewHandler(d *daemon.Daemon) http.Handler {
	mux := http.NewServeMux()

	// Basic health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "OK") //nolint:errcheck
	})

	// Session list endpoint
	mux.HandleFunc("GET /sessions", func(w http.ResponseWriter, _ *http.Request) {
		sessions := d.Sessions()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Session creation request structure
	type CreateSessionRequest struct {
		ID      string   `json:"id"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	// Session creation endpoint
	mux.HandleFunc("POST /sessions", func(w http.ResponseWriter, r *http.Request) {
		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.ID == "" {
			http.Error(w, "Session ID is required", http.StatusBadRequest)
			return
		}

		if req.Command == "" {
			http.Error(w, "Command is required", http.StatusBadRequest)
			return
		}

		s, err := d.CreateSession(req.ID, req.Command, req.Args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(s) //nolint:errcheck
	})

	// Session termination endpoint
	mux.HandleFunc("DELETE /sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := d.StopSession(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// WebSocket attachment endpoint
	mux.HandleFunc("/attach", ws.Handler(d))

	return mux
}
