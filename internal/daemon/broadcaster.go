package daemon

import (
	"io"
	"sync"
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
	mu              sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]*clientState),
	}
}

func (b *Broadcaster) AddClient(c Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

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

	b.updateClientStatuses()
}

func (b *Broadcaster) RemoveClient(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

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

	b.updateClientStatuses()
}

func (b *Broadcaster) SetPrimary(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[id]; ok {
		b.primaryClientID = id
		b.updateClientStatuses()
	}
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
	for id, state := range b.clients {
		_ = state.client.SendStatus(Status{
			IsPrimary: id == b.primaryClientID,
		})
	}
}

func (b *Broadcaster) clientWriteLoop(state *clientState) {
	for {
		select {
		case data := <-state.send:
			_, _ = state.client.Write(data)
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
	defer b.mu.RUnlock()

	for _, state := range b.clients {
		select {
		case state.send <- dataCopy:
		default:
			// Slow client: drop the oldest message and try again to keep current data.
			// This ensures the client eventually catches up with the latest state
			// even if they miss some intermediate output.
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
	b.mu.RUnlock()

	if !isPrimary {
		// Ignore input from non-primary clients
		return 0, nil
	}

	// This would write to the PTY
	return len(data), nil
}
