// Package api provides the local-only HTTP API server for refreSSH.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/strickdd/refressh/internal/api/ui"
	"github.com/strickdd/refressh/internal/api/ws"
	"github.com/strickdd/refressh/internal/config"
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
		ReadHeaderTimeout: 5 * time.Second,  // Address G112: Potential Slowloris Attack
		ReadTimeout:       10 * time.Second, // Maximum duration for reading the entire request
		WriteTimeout:      10 * time.Second, // Maximum duration before timing out writes of the response
		IdleTimeout:       30 * time.Second, // Maximum amount of time to wait for the next request when keep-alive is enabled
	}

	return server.ListenAndServe()
}

// NewHandler creates and configures the API routes, returning an http.Handler.
func NewHandler(d *daemon.Daemon) http.Handler {
	mux := http.NewServeMux()

	// Basic health check endpoint (unauthenticated)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "OK") //nolint:errcheck
	})

	// Web UI Assets (unauthenticated, auth is handled by the UI itself requesting the API)
	uiAssets, err := ui.Assets()
	if err == nil {
		mux.Handle("/", http.FileServer(uiAssets))
	} else {
		fmt.Printf("Warning: Could not load UI assets: %v\n", err)
	}

	// Wrap API routes with auth middleware
	api := http.NewServeMux()

	// Session list endpoint
	api.HandleFunc("GET /sessions", func(w http.ResponseWriter, _ *http.Request) {
		sessions := d.Sessions()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sessions) //nolint:errcheck
	})

	// Session creation request structure
	type CreateSessionRequest struct {
		ID      string   `json:"id"`
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}

	// Session creation endpoint
	api.HandleFunc("POST /sessions", func(w http.ResponseWriter, r *http.Request) {
		// Limit request body to 1MB to prevent memory exhaustion
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		var req CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body or body too large", http.StatusBadRequest)
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
	api.HandleFunc("DELETE /sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := d.StopSession(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// WebSocket attachment endpoint
	api.HandleFunc("/attach", ws.Handler(d))

	// Mount authenticated API routes directly to the main mux
	mux.Handle("GET /sessions", authMiddleware(api))
	mux.Handle("POST /sessions", authMiddleware(api))
	mux.Handle("DELETE /sessions/{id}", authMiddleware(api))
	mux.Handle("/attach", authMiddleware(api))

	return mux
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := config.GetAPIToken()
		if err != nil {
			http.Error(w, "Internal Server Error: failed to load API token", http.StatusInternalServerError)
			return
		}

		// Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer "+token {
			// Also check 'token' query parameter for WebSockets if needed,
			// but we'll stick to headers for now as gorilla supports them.
			if r.URL.Query().Get("token") != token {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
