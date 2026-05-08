package api

import (
	"fmt"
	"net/http"
)

func Start(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("API Server starting on %s...\n", addr)

	// Basic health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	// Bind to local loopback ONLY as per GEMINI.md
	server := &http.Server{
		Addr: addr,
	}

	return server.ListenAndServe()
}
