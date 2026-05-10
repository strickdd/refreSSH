package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/daemon"
)

func TestWSClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close() //nolint:errcheck

		client := NewWSClient("test-client", conn)

		// Test Write (BinaryMessage)
		if _, err := client.Write([]byte("binary data")); err != nil {
			return
		}

		// Test SendStatus (TextMessage)
		if err := client.SendStatus(daemon.Status{IsPrimary: true}); err != nil {
			return
		}
	}))
	defer server.Close()

	u := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close() //nolint:errcheck

	// Read binary data
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read binary data: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Errorf("Expected BinaryMessage, got %d", messageType)
	}
	if string(p) != "binary data" {
		t.Errorf("Expected 'binary data', got '%s'", string(p))
	}

	// Read status message
	_, p, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read status: %v", err)
	}
	var status daemon.Status
	if err := json.Unmarshal(p, &status); err != nil {
		t.Fatalf("Failed to unmarshal status: %v", err)
	}
	if !status.IsPrimary {
		t.Error("Expected IsPrimary to be true")
	}
}

func TestHandler(t *testing.T) {
	d := daemon.New(nil)

	// Use a command that is likely to exist on the platform.
	cmd := "sh"
	if runtime.GOOS == "windows" {
		cmd = "cmd.exe"
	}

	// Create the session the test expects.
	s, err := d.CreateSession("test-session", cmd)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	handler := Handler(d)
	server := httptest.NewServer(handler)
	defer server.Close()

	u := "ws" + strings.TrimPrefix(server.URL, "http") + "?id=test-session"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close() //nolint:errcheck

	// Upon connection, AddClient is called, which triggers a status update.
	// We MUST consume this status message first.
	messageType, _, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read initial status: %v", err)
	}
	if messageType != websocket.TextMessage {
		t.Errorf("Expected TextMessage for status, got %d", messageType)
	}

	// Broadcast something and see if it arrives at the WebSocket
	time.Sleep(200 * time.Millisecond) // Wait for potential shell prompts
	testData := []byte("broadcast test")
	s.Broadcaster.Broadcast(testData)

	foundBroadcast := false
	for i := 0; i < 10; i++ {
		_, p, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read broadcast: %v", err)
		}
		if strings.Contains(string(p), string(testData)) {
			foundBroadcast = true
			break
		}
	}

	if !foundBroadcast {
		t.Errorf("Expected '%s' to be received in WebSocket messages", string(testData))
	}

	// Test input handling: send message from WS to Broadcaster
	inputMsg := []byte("input test")

	// We need to set an input writer to verify
	inputChan := make(chan []byte, 1)
	s.Broadcaster.SetInputWriter(&chanWriter{ch: inputChan})

	if err := conn.WriteMessage(websocket.BinaryMessage, inputMsg); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	select {
	case received := <-inputChan:
		if string(received) != string(inputMsg) {
			t.Errorf("Expected input '%s', got '%s'", string(inputMsg), string(received))
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for input to be handled")
	}
}

type chanWriter struct {
	ch chan []byte
}

func (cw *chanWriter) Write(p []byte) (n int, err error) {
	cw.ch <- p
	return len(p), nil
}
