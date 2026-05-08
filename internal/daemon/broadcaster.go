// Package daemon implements the background service and session management for refreSSH.
package daemon

import (
	"io"
	"sync"
)

// Status represents the current connection status of a client.
type Status struct {
	ID        string `json:"id"`
	IsPrimary bool   `json:"isPrimary"`
}

// Client defines the interface for a connected client (CLI or Web).
type Client interface {
	ID() string
	Write([]byte) (int, error)
	SendStatus(Status) error
}

const (
	// clientBufferSize is the number of messages a client channel can hold before dropping data.
	clientBufferSize = 100
)

type clientState struct {
	client Client
	send   chan []byte
	quit   chan struct{}
}

// Broadcaster manages multiple connected clients and handles PTY output distribution.
type Broadcaster struct {
	clients         map[string]*clientState
	primaryClientID string
	inputWriter     io.Writer
	mu              sync.RWMutex
}

// NewBroadcaster creates a new Broadcaster instance.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[string]*clientState),
	}
}

// SetInputWriter sets the writer to which primary client input is forwarded.
func (b *Broadcaster) SetInputWriter(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inputWriter = w
}

// AddClient registers a new client for receiving broadcasted output.
func (b *Broadcaster) AddClient(c Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Handle duplicate IDs
	if old, exists := b.clients[c.ID()]; exists {
		close(old.quit)
		delete(b.clients, c.ID())
	}

	state := &clientState{
		client: c,
		send:   make(chan []byte, clientBufferSize),
		quit:   make(chan struct{}),
	}
	b.clients[c.ID()] = state

	if b.primaryClientID == "" {
		b.primaryClientID = c.ID()
	}

	go b.clientWriteLoop(state)
	go b.updateClientStatuses()
}

func (b *Broadcaster) clientWriteLoop(state *clientState) {
	for {
		select {
		case data := <-state.send:
			if _, err := state.client.Write(data); err != nil {
				b.RemoveClient(state.client.ID())
				return
			}
		case <-state.quit:
			return
		}
	}
}

// RemoveClient unregisters a client by its ID.
func (b *Broadcaster) RemoveClient(id string) {
	b.mu.Lock()
	state, exists := b.clients[id]
	if !exists {
		b.mu.Unlock()
		return
	}

	close(state.quit)
	delete(b.clients, id)

	if b.primaryClientID == id {
		b.primaryClientID = ""
		// Assign a new primary if any clients remain
		for newID := range b.clients {
			b.primaryClientID = newID
			break
		}
	}
	b.mu.Unlock()
	b.updateClientStatuses()
}

// SetPrimary assigns a specific client as the primary controller.
func (b *Broadcaster) SetPrimary(id string) {
	b.mu.Lock()
	if _, exists := b.clients[id]; exists {
		b.primaryClientID = id
	}
	b.mu.Unlock()
	b.updateClientStatuses()
}

// Close terminates all client connections and stops broadcasting.
func (b *Broadcaster) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, state := range b.clients {
		close(state.quit)
		delete(b.clients, id)
	}
	b.primaryClientID = ""
}

func (b *Broadcaster) updateClientStatuses() {
	b.mu.RLock()
	type target struct {
		client    Client
		isPrimary bool
	}
	targets := make([]target, 0, len(b.clients))
	for id, state := range b.clients {
		targets = append(targets, target{
			client:    state.client,
			isPrimary: id == b.primaryClientID,
		})
	}
	b.mu.RUnlock()

	for _, t := range targets {
		_ = t.client.SendStatus(Status{
			ID:        t.client.ID(),
			IsPrimary: t.isPrimary,
		})
	}
}

// Broadcast sends data to all registered clients.
func (b *Broadcaster) Broadcast(data []byte) {
	b.mu.RLock()
	states := make([]*clientState, 0, len(b.clients))
	for _, state := range b.clients {
		states = append(states, state)
	}
	b.mu.RUnlock()

	for _, state := range states {
		select {
		case state.send <- data:
		default:
			// Buffer full, drop data for this client
		}
	}
}

// HandleInput processes input from a client and forwards it to the PTY if the client is primary.
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
