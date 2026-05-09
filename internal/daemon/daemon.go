// Package daemon implements the core background process and terminal management for refreSSH.
package daemon

import (
	"fmt"
	"sync"

	"github.com/strickdd/refressh/internal/api"
	"github.com/strickdd/refressh/internal/config"
)

// Daemon is the primary service that manages pseudo-terminal sessions and coordinates client broadcasting.
type Daemon struct {
	config      *config.Config
	sessions    map[string]*Session
	broadcaster *Broadcaster
	mu          sync.RWMutex
}

// New creates and initializes a new refreSSH Daemon instance.
func New(cfg *config.Config) *Daemon {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	return &Daemon{
		config:      cfg,
		sessions:    make(map[string]*Session),
		broadcaster: NewBroadcaster(),
	}
}

// Start launches the background daemon and its local API server.
func (d *Daemon) Start() error {
	fmt.Println("Daemon starting...")

	// Create a default session for the daemon's own shell
	s := NewSession("default", d.config.DefaultTerminal)
	if err := s.Start(); err != nil {
		// Attempt fallbacks if the configured shell is unavailable
		s = NewSession("default", "sh")
		if err := s.Start(); err != nil {
			s = NewSession("default", "cmd.exe")
			if err := s.Start(); err != nil {
				return fmt.Errorf("failed to start any shell: %w", err)
			}
		}
	}

	d.mu.Lock()
	d.sessions[s.id] = s
	d.mu.Unlock()

	// Direct primary client input to the PTY
	d.broadcaster.SetInputWriter(s.pty)

	// Start reading from PTY and broadcasting to clients
	go d.broadcastLoop(s)

	return api.Start(d.config.Port)
}

func (d *Daemon) broadcastLoop(s *Session) {
	defer func() {
		s.mu.Lock()
		s.running = false
		if s.pty != nil {
			_ = s.pty.Close()
		}
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Wait()
		}
		s.mu.Unlock()
	}()

	buf := make([]byte, 1024)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			d.broadcaster.Broadcast(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
