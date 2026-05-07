package daemon

import (
	"bytes"
	"testing"
)

type mockClient struct {
	id     string
	buf    bytes.Buffer
	status Status
}

func (m *mockClient) ID() string { return m.id }
func (m *mockClient) Write(p []byte) (n int, err error) {
	return m.buf.Write(p)
}
func (m *mockClient) SendStatus(status Status) error {
	m.status = status
	return nil
}

func TestBroadcaster(t *testing.T) {
	b := NewBroadcaster()
	c1 := &mockClient{id: "client1"}
	c2 := &mockClient{id: "client2"}

	// Test AddClient
	b.AddClient(c1)
	if !c1.status.IsPrimary {
		t.Error("First client should be primary")
	}

	b.AddClient(c2)
	if c2.status.IsPrimary {
		t.Error("Second client should not be primary")
	}

	// Test Broadcast
	msg := []byte("hello")
	b.Broadcast(msg)

	if !bytes.Equal(c1.buf.Bytes(), msg) {
		t.Errorf("Expected %s, got %s", msg, c1.buf.String())
	}
	if !bytes.Equal(c2.buf.Bytes(), msg) {
		t.Errorf("Expected %s, got %s", msg, c2.buf.String())
	}

	// Test Primary Handoff
	b.SetPrimary("client2")
	if c1.status.IsPrimary {
		t.Error("client1 should no longer be primary")
	}
	if !c2.status.IsPrimary {
		t.Error("client2 should now be primary")
	}

	// Test Input Handling
	n, _ := b.HandleInput("client1", []byte("bad"))
	if n != 0 {
		t.Error("Non-primary client should not be able to send input")
	}

	n, _ = b.HandleInput("client2", []byte("good"))
	if n != 4 {
		t.Error("Primary client should be able to send input")
	}

	// Test RemoveClient (Handoff)
	b.RemoveClient("client2")
	if !c1.status.IsPrimary {
		t.Error("client1 should be promoted back to primary")
	}
}
