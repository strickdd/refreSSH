// Package daemon implements the core background process and terminal management for refreSSH.
package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	// Attempt to recover previous sessions
	if err := d.loadState(); err != nil {
		fmt.Printf("Warning: failed to load previous session state: %v\n", err)
	}

	// Create a default session if no sessions exist
	d.mu.RLock()
	sessionCount := len(d.sessions)
	d.mu.RUnlock()

	if sessionCount == 0 {
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
	} else {
		// Restart recovered sessions
		d.mu.RLock()
		var sessionsToRestart []*Session
		for _, s := range d.sessions {
			sessionsToRestart = append(sessionsToRestart, s)
		}
		d.mu.RUnlock()

		for _, s := range sessionsToRestart {
			if err := s.Start(); err != nil {
				fmt.Printf("Warning: failed to restart session %s: %v\n", s.ID, err)
				continue
			}
			s.Broadcaster.SetInputWriter(s.pty)
			go d.broadcastLoop(s)
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

	fmt.Printf("Creating session %s: %s %v\n", id, command, args)
	s := NewSession(id, command, args...)
	if err := s.Start(); err != nil {
		return nil, err
	}

	// Link the session's broadcaster to its PTY
	s.Broadcaster.SetInputWriter(s.pty)

	d.mu.Lock()
	d.sessions[id] = s
	d.mu.Unlock()

	// Persist state
	if err := d.saveState(); err != nil {
		fmt.Printf("Warning: failed to save session state: %v\n", err)
	}

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

	fmt.Printf("Stopping session %s\n", id)
	// Persist state
	if err := d.saveState(); err != nil {
		fmt.Printf("Warning: failed to save session state: %v\n", err)
	}

	s.mu.Lock()
	if s.Running {
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Process.Kill() //nolint:errcheck
		}
		s.Running = false
	}
	if s.pty != nil {
		_ = s.pty.Close() //nolint:errcheck
	}
	s.mu.Unlock()

	s.Broadcaster.Close()

	return nil
}

// saveState persists the current list of sessions to disk.
func (d *Daemon) saveState() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	statePath := filepath.Join(configDir, "sessions.json")

	d.mu.RLock()
	sessions := make([]*Session, 0, len(d.sessions))
	for _, s := range d.sessions {
		sessions = append(sessions, s)
	}
	d.mu.RUnlock()

	data, err := json.MarshalIndent(sessions, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0600)
}

// loadState reads the session list from disk.
func (d *Daemon) loadState() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(configDir, "sessions.json")
	data, err := os.ReadFile(filepath.Clean(statePath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var sessions []*Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	for _, s := range sessions {
		// Initialize broadcaster and other non-serialized fields
		s.Broadcaster = NewBroadcaster()
		s.Running = false // Will be set to true if successfully restarted
		d.sessions[s.ID] = s
	}

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
				_ = s.pty.Close() //nolint:errcheck
			}
			if s.cmd != nil && s.cmd.Process != nil {
				_ = s.cmd.Wait() //nolint:errcheck
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
