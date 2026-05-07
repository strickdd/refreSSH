package api

import (
	"fmt"
	"net/http"
)

func Start() error {
	fmt.Println("API Server starting on 127.0.0.1:9000...")
	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	// Bind to local loopback ONLY as per GEMINI.md
	return nil // Stub for now, we'll implement the actual listener later
}
