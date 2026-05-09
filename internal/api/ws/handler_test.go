package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/strickdd/refressh/internal/daemon"
)

func TestWSClient(t *testing.T) {
	// Start a test server to get a real WebSocket connection
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		client := NewWSClient("test-client", conn)

		// Test Write (BinaryMessage)
		_, _ = client.Write([]byte("binary data"))

		// Test SendStatus (TextMessage)
		_ = client.SendStatus(daemon.Status{IsPrimary: true})
	}))
	defer s.Close()

	// Connect to the test server
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Read binary message
	messageType, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Errorf("Expected BinaryMessage, got %d", messageType)
	}
	if string(p) != "binary data" {
		t.Errorf("Expected 'binary data', got '%s'", string(p))
	}

	// Read status message
	messageType, p, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read status: %v", err)
	}
	if messageType != websocket.TextMessage {
		t.Errorf("Expected TextMessage, got %d", messageType)
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
	broadcaster := daemon.NewBroadcaster()
	handler := Handler(broadcaster)
	server := httptest.NewServer(handler)
	defer server.Close()

	u := "ws" + strings.TrimPrefix(server.URL, "http") + "?id=test-session"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Wait for client to be added to broadcaster
	time.Sleep(50 * time.Millisecond)

	// Broadcast something and see if it arrives at the WebSocket
	testData := []byte("broadcast test")
	broadcaster.Broadcast(testData)

	_, p, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read broadcast: %v", err)
	}
	if string(p) != string(testData) {
		t.Errorf("Expected '%s', got '%s'", string(testData), string(p))
	}

	// Test input handling: send message from WS to Broadcaster
	inputMsg := []byte("input test")
	
	// We need to set an input writer to verify
	inputChan := make(chan []byte, 1)
	broadcaster.SetInputWriter(&chanWriter{ch: inputChan})

	if err := conn.WriteMessage(websocket.BinaryMessage, inputMsg); err != nil {
		t.Fatalf("Failed to write input: %v", err)
	}

	select {
	case received := <-inputChan:
		if string(received) != string(inputMsg) {
			t.Errorf("Expected input '%s', got '%s'", string(inputMsg), string(received))
		}
	case <-time.After(100 * time.Millisecond):
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
