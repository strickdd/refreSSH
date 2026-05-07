package daemon

import (
	"bytes"
	"sync"
	"testing"
	"time"
)

type mockClient struct {
	id        string
	mu        sync.Mutex
	buf       bytes.Buffer
	status    Status
	written   chan struct{}
	WriteFunc func(p []byte) (int, error)
}

func newMockClient(id string) *mockClient {
	return &mockClient{
		id:      id,
		written: make(chan struct{}, 1000),
	}
}

func (m *mockClient) ID() string { return m.id }
func (m *mockClient) Write(p []byte) (n int, err error) {
	if m.WriteFunc != nil {
		return m.WriteFunc(p)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n, err = m.buf.Write(p)
	m.written <- struct{}{}
	return n, err
}
func (m *mockClient) SendStatus(status Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = status
	return nil
}

func (m *mockClient) getStatus() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *mockClient) getBuffer() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.buf.Bytes()
}

func TestBroadcaster(t *testing.T) {
	b := NewBroadcaster()
	c1 := newMockClient("client1")
	c2 := newMockClient("client2")

	// Test AddClient
	b.AddClient(c1)
	time.Sleep(10 * time.Millisecond)
	if !c1.getStatus().IsPrimary {
		t.Error("First client should be primary")
	}

	b.AddClient(c2)
	time.Sleep(10 * time.Millisecond)
	if c2.getStatus().IsPrimary {
		t.Error("Second client should not be primary")
	}

	// Test Broadcast
	msg := []byte("hello")
	b.Broadcast(msg)

	// Wait for delivery
	waitChan := func(c *mockClient) {
		select {
		case <-c.written:
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Timeout waiting for %s broadcast", c.id)
		}
	}
	waitChan(c1)
	waitChan(c2)

	if !bytes.Equal(c1.getBuffer(), msg) {
		t.Errorf("Expected %s, got %s", msg, string(c1.getBuffer()))
	}
	if !bytes.Equal(c2.getBuffer(), msg) {
		t.Errorf("Expected %s, got %s", msg, string(c2.getBuffer()))
	}

	// Test Primary Handoff
	b.SetPrimary("client2")
	time.Sleep(10 * time.Millisecond)
	if c1.getStatus().IsPrimary {
		t.Error("client1 should no longer be primary")
	}
	if !c2.getStatus().IsPrimary {
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
	time.Sleep(10 * time.Millisecond)
	if !c1.getStatus().IsPrimary {
		t.Error("client1 should be promoted back to primary")
	}
}

func TestBroadcaster_SlowClient(t *testing.T) {
	b := NewBroadcaster()

	// c1 is slow and blocks
	c1 := newMockClient("slow")
	blocked := make(chan struct{})
	c1.WriteFunc = func(p []byte) (int, error) {
		<-blocked
		return len(p), nil
	}

	// c2 is fast
	c2 := newMockClient("fast")

	b.AddClient(c1)
	b.AddClient(c2)

	// Broadcast multiple messages
	for i := 0; i < 10; i++ {
		b.Broadcast([]byte("data"))
	}

	// c2 should receive all of them immediately despite c1 being blocked
	for i := 0; i < 10; i++ {
		select {
		case <-c2.written:
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Fast client c2 was blocked at message %d", i)
		}
	}

	// Unblock c1
	close(blocked)
}

func TestBroadcaster_DropData(t *testing.T) {
	b := NewBroadcaster()

	// Client with a blocking write
	c := newMockClient("slow")
	writeCalled := make(chan struct{})
	c.WriteFunc = func(p []byte) (int, error) {
		writeCalled <- struct{}{}
		time.Sleep(100 * time.Millisecond)
		return len(p), nil
	}

	b.AddClient(c)

	// Fill the buffer + 2 (one in flight, one in buffer, one to trigger drop)
	// Actually clientBufferSize is 256.
	// Let's just broadcast many messages.

	for i := 0; i < clientBufferSize+10; i++ {
		b.Broadcast([]byte("data"))
	}

	// Broadcast should not block
	done := make(chan struct{})
	go func() {
		b.Broadcast([]byte("last"))
		close(done)
	}()

	select {
	case <-done:
		// Success: Broadcast didn't block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Broadcast blocked on slow client")
	}
}

func TestBroadcaster_Close(t *testing.T) {
	b := NewBroadcaster()
	c := newMockClient("c")
	b.AddClient(c)

	b.Close()

	// Wait a bit for the goroutine to exit
	time.Sleep(10 * time.Millisecond)

	b.Broadcast([]byte("test"))

	// c should not receive anything because its loop is closed
	select {
	case <-c.written:
		t.Error("Client received message after broadcaster was closed")
	case <-time.After(50 * time.Millisecond):
		// Success
	}
}
