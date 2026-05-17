// Package daemon manages the background process, terminal sessions, and client communication.
package daemon

import (
	"fmt"
	"io"
	"sync"
)

// Client represents a connected user or automated interface.
type Client interface {
	io.Writer
	ID() string
	SendStatus(Status) error
}

// Status represents the current connection state of a client.
type Status struct {
	IsPrimary bool `json:"is_primary"`
}

const (
	// clientBufferSize is the number of messages to buffer per client before dropping data.
	clientBufferSize = 256
	// maxScrollbackSize is the maximum number of bytes to keep in the scrollback buffer.
	maxScrollbackSize = 64 * 1024
)

type clientState struct {
	client Client
	send   chan []byte
	quit   chan struct{}
	remove chan struct{}
	wg     sync.WaitGroup
}

// Broadcaster manages a set of connected clients and handles asynchronous PTY output distribution.
type Broadcaster struct {
	clients         map[string]*clientState
	primaryClientID string
	inputWriter     io.Writer
	scrollback      []byte
	mu              sync.RWMutex
	closed          bool
}

// NewBroadcaster creates and initializes a new Broadcaster instance.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients:    make(map[string]*clientState),
		scrollback: make([]byte, 0, maxScrollbackSize),
	}
}

// SetInputWriter specifies the writer (usually a PTY) where primary client input should be directed.
func (b *Broadcaster) SetInputWriter(w io.Writer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inputWriter = w
}

// AddClient registers a new client with the broadcaster and starts its individual write loop.
func (b *Broadcaster) AddClient(c Client) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}

	// If client already exists, close the old one and wait for it to exit
	if oldState, ok := b.clients[c.ID()]; ok {
		close(oldState.quit)
		b.mu.Unlock()
		oldState.wg.Wait()
		b.mu.Lock()
		delete(b.clients, c.ID())
	}

	state := &clientState{
		client: c,
		send:   make(chan []byte, clientBufferSize),
		quit:   make(chan struct{}),
		remove: make(chan struct{}),
	}
	b.clients[c.ID()] = state

	// Send current scrollback to the new client first
	if len(b.scrollback) > 0 {
		sbCopy := make([]byte, len(b.scrollback))
		copy(sbCopy, b.scrollback)
		state.send <- sbCopy
	}

	state.wg.Add(2)
	go b.clientWriteLoop(state)

	go b.cleanupLoop(state)

	// If no primary client, make this one primary
	if b.primaryClientID == "" {
		b.primaryClientID = c.ID()
	}
	b.mu.Unlock()

	b.updateClientStatuses()
}

// RemoveClient unregisters a client by its unique identifier and cleans up its resources.
func (b *Broadcaster) RemoveClient(id string) {
	b.mu.Lock()
	if state, ok := b.clients[id]; ok {
		close(state.quit)
		b.mu.Unlock()
		state.wg.Wait()
		b.mu.Lock()
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

// SetPrimary designates a specific connected client as the primary controller.
func (b *Broadcaster) SetPrimary(id string) {
	b.mu.Lock()
	if _, ok := b.clients[id]; ok {
		fmt.Printf("Promoting client %s to primary\n", id)
		b.primaryClientID = id
	}
	b.mu.Unlock()
	b.updateClientStatuses()
}

// Close terminates all client write loops and clears the client registry.
func (b *Broadcaster) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true

	var states []*clientState
	for _, state := range b.clients {
		states = append(states, state)
		close(state.quit)
	}
	b.mu.Unlock()

	for _, state := range states {
		state.wg.Wait()
	}

	b.mu.Lock()
	b.clients = make(map[string]*clientState)
	b.primaryClientID = ""
	b.mu.Unlock()
}

func (b *Broadcaster) updateClientStatuses() {
	b.mu.RLock()
	type target struct {
		c         Client
		isPrimary bool
		state     *clientState
	}
	var targets []target
	for id, state := range b.clients {
		targets = append(targets, target{
			c:         state.client,
			isPrimary: id == b.primaryClientID,
			state:     state,
		})
	}
	b.mu.RUnlock()

	for _, t := range targets {
		if err := t.c.SendStatus(Status{
			IsPrimary: t.isPrimary,
		}); err != nil {
			select {
			case t.state.remove <- struct{}{}:
			default:
			}
		}
	}
}

func (b *Broadcaster) clientWriteLoop(state *clientState) {
	defer state.wg.Done()
	for {
		select {
		case data := <-state.send:
			if _, err := state.client.Write(data); err != nil {
				select {
				case state.remove <- struct{}{}:
				default:
				}
				return
			}
		case <-state.quit:
			return
		}
	}
}

func (b *Broadcaster) cleanupLoop(state *clientState) {
	defer state.wg.Done()
	for {
		select {
		case <-state.remove:
			b.mu.Lock()
			close(state.quit)
			if b.clients[state.client.ID()] == state {
				delete(b.clients, state.client.ID())
			}
			if b.primaryClientID == state.client.ID() {
				b.primaryClientID = ""
				for nextID := range b.clients {
					b.primaryClientID = nextID
					break
				}
			}
			primary := b.primaryClientID
			b.mu.Unlock()

			b.mu.RLock()
			var targets []Client
			for _, s := range b.clients {
				targets = append(targets, s.client)
			}
			b.mu.RUnlock()
			for _, t := range targets {
				_ = t.SendStatus(Status{IsPrimary: t.ID() == primary}) //nolint:errcheck
			}
		case <-state.quit:
			return
		}
	}
}

// Scrollback returns a copy of the current scrollback buffer.
func (b *Broadcaster) Scrollback() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.scrollback) == 0 {
		return nil
	}

	sbCopy := make([]byte, len(b.scrollback))
	copy(sbCopy, b.scrollback)
	return sbCopy
}

// Broadcast distributes a copy of the data slice to all currently registered clients.
func (b *Broadcaster) Broadcast(data []byte) {
	if len(data) == 0 {
		return
	}

	// Create a copy of the data to safely share among goroutines
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}

	// Update scrollback buffer
	b.scrollback = append(b.scrollback, dataCopy...)
	if len(b.scrollback) > maxScrollbackSize {
		// Truncate from the beginning to maintain size
		excess := len(b.scrollback) - maxScrollbackSize
		b.scrollback = b.scrollback[excess:]
	}

	var states []*clientState
	for _, state := range b.clients {
		states = append(states, state)
	}
	b.mu.Unlock()

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

// HandleInput routes input from a specific client to the designated PTY writer.
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
