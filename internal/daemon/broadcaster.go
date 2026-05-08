package daemon

import (
	"io"
	"sync"
	"time"
)

// Client represents a connected user/interface
type Client interface {
	io.Writer
	ID() string
	SendStatus(status Status) error
}

// Status represents the state of a client (Primary or View-only)
type Status struct {
	IsPrimary bool `json:"is_primary"`
}

const (
	// clientBufferSize is the number of messages to buffer per client
	clientBufferSize = 256
)

type clientState struct {
	client Client
	send   chan []byte
	quit   chan struct{}
}

// Broadcaster manages multiple clients and broadcasts data to them
type Broadcaster struct {
	clients         map[string]*clientState
	primaryClientID string
	inputWriter     io.Writer
	mu              sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]*clientState),
	}
}

func (b *Broadcaster) SetInputWriter(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inputWriter = w
}

func (b *Broadcaster) AddClient(c Client) {
	b.mu.Lock()

	// If client already exists, close the old one to avoid goroutine leaks
	if oldState, ok := b.clients[c.ID()]; ok {
		close(oldState.quit)
		delete(b.clients, c.ID())
	}

	state := &clientState{
		client: c,
		send:   make(chan []byte, clientBufferSize),
		quit:   make(chan struct{}),
	}
	b.clients[c.ID()] = state

	go b.clientWriteLoop(state)

	// If no primary client, make this one primary
	if b.primaryClientID == "" {
		b.primaryClientID = c.ID()
	}
	b.mu.Unlock()

	b.updateClientStatuses()
}

func (b *Broadcaster) RemoveClient(id string) {
	b.mu.Lock()
	if state, ok := b.clients[id]; ok {
		close(state.quit)
		delete(b.clients, id)
	}

	if b.primaryClientID == id {
		b.primaryClientID = ""
		// Promote another client if available
		for nextID := range b.clients {
			b.primaryClientID = nextID
			break
		}
	}
	b.mu.Unlock()

	b.updateClientStatuses()
}

func (b *Broadcaster) SetPrimary(id string) {
	b.mu.Lock()
	if _, ok := b.clients[id]; ok {
		b.primaryClientID = id
	}
	b.mu.Unlock()
	b.updateClientStatuses()
}

func (b *Broadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, state := range b.clients {
		close(state.quit)
		delete(b.clients, id)
	}
}

func (b *Broadcaster) updateClientStatuses() {
	b.mu.RLock()
	type target struct {
		c         Client
		isPrimary bool
	}
	var targets []target
	for id, state := range b.clients {
		targets = append(targets, target{
			c:         state.client,
			isPrimary: id == b.primaryClientID,
		})
	}
	b.mu.RUnlock()

	for _, t := range targets {
		_ = t.c.SendStatus(Status{
			IsPrimary: t.isPrimary,
		})
	}
}

func (b *Broadcaster) clientWriteLoop(state *clientState) {
	for {
		select {
		case data := <-state.send:
			// Ensure write doesn't block indefinitely
			done := make(chan struct{})
			go func() {
				_, _ = state.client.Write(data)
				close(done)
			}()

			select {
			case <-done:
				// Write completed
			case <-time.After(5 * time.Second):
				// Write timed out
			case <-state.quit:
				return
			}
		case <-state.quit:
			return
		}
	}
}

func (b *Broadcaster) Broadcast(data []byte) {
	if len(data) == 0 {
		return
	}

	// Create a copy of the data to safely share among goroutines
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	b.mu.RLock()
	var states []*clientState
	for _, state := range b.clients {
		states = append(states, state)
	}
	b.mu.RUnlock()

	for _, state := range states {
		select {
		case state.send <- dataCopy:
		default:
			// Slow client: drop the oldest message and try again to keep current data.
			select {
			case <-state.send:
			default:
			}
			select {
			case state.send <- dataCopy:
			default:
				// If it's still full, we just drop this update for this client.
			}
		}
	}
}

func (b *Broadcaster) HandleInput(clientID string, data []byte) (int, error) {
	b.mu.RLock()
	isPrimary := clientID == b.primaryClientID
	writer := b.inputWriter
	b.mu.RUnlock()

	if !isPrimary || writer == nil {
		// Ignore input from non-primary clients or if no writer is set
		return 0, nil
	}

	return writer.Write(data)
}
