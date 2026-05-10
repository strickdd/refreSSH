// Package daemon implements the core background process and terminal management for refreSSH.
package daemon

import (
	"fmt"
	"sync"

	"github.com/strickdd/refressh/internal/config"
)

// Daemon is the primary service that manages pseudo-terminal sessions and coordinates client broadcasting.
type Daemon struct {
	config   *config.Config
	sessions map[string]*Session
	mu       sync.RWMutex
}

// New creates and initializes a new refreSSH Daemon instance.
func New(cfg *config.Config) *Daemon {
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}
	return &Daemon{
		config:   cfg,
		sessions: make(map[string]*Session),
	}
}

// Start launches the background daemon and its local API server.
func (d *Daemon) Start() error {
	fmt.Println("Daemon starting...")

	// Create a default session for the daemon's own shell
	_, err := d.CreateSession("default", d.config.DefaultTerminal)
	if err != nil {
		// Attempt fallbacks if the configured shell is unavailable
		_, err = d.CreateSession("default", "sh")
		if err != nil {
			_, err = d.CreateSession("default", "cmd.exe")
			if err != nil {
				return fmt.Errorf("failed to start any shell: %w", err)
			}
		}
	}

	return nil
}

// CreateSession creates, starts, and registers a new terminal session.
func (d *Daemon) CreateSession(id string, command string, args ...string) (*Session, error) {
	d.mu.Lock()
	if _, exists := d.sessions[id]; exists {
		d.mu.Unlock()
		return nil, fmt.Errorf("session already exists: %s", id)
	}
	d.mu.Unlock()

	s := NewSession(id, command, args...)
	if err := s.Start(); err != nil {
		return nil, err
	}

	// Link the session's broadcaster to its PTY
	s.Broadcaster.SetInputWriter(s.pty)

	d.mu.Lock()
	d.sessions[id] = s
	d.mu.Unlock()

	// Start reading from PTY and broadcasting to clients
	go d.broadcastLoop(s)

	return s, nil
}

// StopSession terminates a session and removes it from the daemon.
func (d *Daemon) StopSession(id string) error {
	d.mu.Lock()
	s, exists := d.sessions[id]
	if !exists {
		d.mu.Unlock()
		return fmt.Errorf("session not found: %s", id)
	}
	delete(d.sessions, id)
	d.mu.Unlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Running {
		if s.cmd != nil && s.cmd.Process != nil {
			// Try to terminate gracefully, then kill
			_ = s.cmd.Process.Kill()
		}
		s.Running = false
	}

	if s.pty != nil {
		_ = s.pty.Close()
	}

	s.Broadcaster.Close()

	return nil
}

// Session returns a specific session by its ID.
func (d *Daemon) Session(id string) (*Session, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	s, ok := d.sessions[id]
	return s, ok
}

// Sessions returns a list of all currently managed sessions.
func (d *Daemon) Sessions() []*Session {
	d.mu.RLock()
	defer d.mu.RUnlock()

	sessions := make([]*Session, 0, len(d.sessions))
	for _, s := range d.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (d *Daemon) broadcastLoop(s *Session) {
	defer func() {
		s.mu.Lock()
		if s.Running {
			s.Running = false
			if s.pty != nil {
				_ = s.pty.Close()
			}
			if s.cmd != nil && s.cmd.Process != nil {
				_ = s.cmd.Wait()
			}
		}
		s.mu.Unlock()
		s.Broadcaster.Close()
	}()

	buf := make([]byte, 1024)
	for {
		n, err := s.pty.Read(buf)
		if n > 0 {
			s.Broadcaster.Broadcast(buf[:n])
		}
		if err != nil {
			break
		}
	}
}
