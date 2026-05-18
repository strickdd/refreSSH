// Package ws provides WebSocket-based terminal attachment handlers.
package ws
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/daemon"
)

var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Not a browser request (likely CLI)
			}
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			// Strictly allow only local origins to prevent CSWH
			return u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"
		},
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

const (
	// writeWait is the maximum time to wait for a message to be written.
	writeWait = 10 * time.Second
	// readWait is the maximum time to wait for a message to be read before closing idle connections.
	readWait = 60 * time.Second
	// pingPeriod is how often to send pings to detect dead connections.
	pingPeriod = (readWait * 9) / 10
	// pongWait is how long to wait for a pong response before declaring the connection dead.
	pongWait = readWait + 6*time.Second
)

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
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *Client) startPingLoop(conn *websocket.Conn, quit chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.mu.Lock()
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					c.mu.Unlock()
					return
				}
				c.mu.Unlock()
			case <-quit:
				return
			}
		}
	}()
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
			fmt.Printf("Upgrade failed for session %s: %v\n", sessionID, err)
			return
		}

		// Use remote address as client ID for uniqueness within this session
		clientID := r.RemoteAddr
		client := NewWSClient(clientID, conn)

		// Set initial read deadline for idle timeout
		if err := conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
			fmt.Printf("Client %s failed to set read deadline: %v\n", clientID, err)
		}

		// Set close handler to ignore close frames (we manage close lifecycle)
		conn.SetCloseHandler(func(_ int, _ string) error {
			return nil
		})

		fmt.Printf("Client %s connected to session %s\n", clientID, sessionID)

		s.Broadcaster.AddClient(client)

		// Start ping/pong keepalive goroutine
		quit := make(chan struct{})
		client.startPingLoop(conn, quit)

		// Read loop: handle incoming messages
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseAbnormalClosure) {
					fmt.Printf("Client %s WS unexpected error: %v\n", clientID, err)
				}
				break
			}

			// Reset read deadline on each successful message
			if err := conn.SetReadDeadline(time.Now().Add(readWait)); err != nil {
				fmt.Printf("Client %s failed to reset read deadline: %v\n", clientID, err)
			}

			if msgType == websocket.TextMessage {
				type ControlMessage struct {
					Action string `json:"action"`
				}
				var ctrl ControlMessage
				if err := json.Unmarshal(message, &ctrl); err == nil {
					fmt.Printf("Client %s sent control action: %s\n", clientID, ctrl.Action)
					if ctrl.Action == "request_primary" {
						s.Broadcaster.SetPrimary(clientID)
					}
				}
				continue
			}

			if _, err := s.Broadcaster.HandleInput(clientID, message); err != nil {
				fmt.Printf("Failed to handle input from client %s: %v\n", clientID, err)
				break
			}
		}

		// Signal ping loop to stop and cleanup
		close(quit)
		fmt.Printf("Client %s disconnected from session %s\n", clientID, sessionID)
		s.Broadcaster.RemoveClient(clientID)
		conn.Close() //nolint:errcheck,gosec
	}
}
