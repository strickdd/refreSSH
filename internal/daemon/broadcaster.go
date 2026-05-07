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

// Broadcaster manages multiple clients and broadcasts data to them
type Broadcaster struct {
	clients         map[string]Client
	primaryClientID string
	mu              sync.RWMutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]Client),
	}
}

func (b *Broadcaster) AddClient(c Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c.ID()] = c
	
	// If no primary client, make this one primary
	if b.primaryClientID == "" {
		b.primaryClientID = c.ID()
	}
	
	b.updateClientStatuses()
}

func (b *Broadcaster) RemoveClient(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, id)
	
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

func (b *Broadcaster) updateClientStatuses() {
	for id, c := range b.clients {
		c.SendStatus(Status{
			IsPrimary: id == b.primaryClientID,
		})
	}
}

func (b *Broadcaster) Broadcast(data []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, c := range b.clients {
		_, _ = c.Write(data)
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
