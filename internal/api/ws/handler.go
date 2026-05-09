package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/daemon"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Strictly local-only as per GEMINI.md, but ListenAndServe on 127.0.0.1 handles this.
		// For extra safety, we could check r.RemoteAddr.
		return true
	},
}

// WSClient wraps a WebSocket connection to satisfy the daemon.Client interface.
type WSClient struct {
	id   string
	conn *websocket.Conn
	mu   sync.Mutex
}

// NewWSClient creates a new WSClient instance.
func NewWSClient(id string, conn *websocket.Conn) *WSClient {
	return &WSClient{
		id:   id,
		conn: conn,
	}
}

// ID returns the unique identifier for the client.
func (c *WSClient) ID() string {
	return c.id
}

// Write sends binary data (PTY output) to the WebSocket client.
func (c *WSClient) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err = c.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// SendStatus sends a status update to the client as a JSON message.
func (c *WSClient) SendStatus(status daemon.Status) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// Handler returns an HTTP handler for WebSocket attachment.
func Handler(broadcaster *daemon.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("id")
		if sessionID == "" {
			http.Error(w, "Session ID required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Use remote address as client ID for uniqueness within this session
		clientID := r.RemoteAddr
		client := NewWSClient(clientID, conn)

		broadcaster.AddClient(client)
		defer broadcaster.RemoveClient(clientID)

		// Input loop: forward messages from WebSocket to Broadcaster
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			_, _ = broadcaster.HandleInput(clientID, message)
		}
	}
}
