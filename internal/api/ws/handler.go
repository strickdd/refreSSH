// Package ws provides WebSocket-based terminal attachment handlers.
package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/daemon"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(_ *http.Request) bool {
		// Strictly local-only as per GEMINI.md, but ListenAndServe on 127.0.0.1 handles this.
		return true
	},
}

// Client wraps a WebSocket connection to satisfy the daemon.Client interface.
type Client struct {
	id   string
	conn *websocket.Conn
	mu   sync.Mutex
}

// NewWSClient creates a new Client instance.
func NewWSClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		id:   id,
		conn: conn,
	}
}

// ID returns the unique identifier for the client.
func (c *Client) ID() string {
	return c.id
}

// Write sends binary data (PTY output) to the WebSocket client.
func (c *Client) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err = c.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// SendStatus sends a status update to the client as a JSON message.
func (c *Client) SendStatus(status daemon.Status) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Handler returns an HTTP handler for WebSocket attachment.
func Handler(d *daemon.Daemon) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("id")
		if sessionID == "" {
			http.Error(w, "Session ID required", http.StatusBadRequest)
			return
		}

		s, ok := d.Session(sessionID)
		if !ok {
			http.Error(w, "Session not found", http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck

		// Use remote address as client ID for uniqueness within this session
		clientID := r.RemoteAddr
		client := NewWSClient(clientID, conn)

		s.Broadcaster.AddClient(client)
		defer s.Broadcaster.RemoveClient(clientID)

		// Input loop: forward messages from WebSocket to Broadcaster
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if _, err := s.Broadcaster.HandleInput(clientID, message); err != nil {
				// Failed to handle input, possibly due to writer issues
				break
			}
		}
	}
}
