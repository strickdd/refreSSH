package daemon

import (
	"fmt"
	"sync"

	"github.com/strickdd/refressh/internal/api"
	"github.com/strickdd/refressh/internal/config"
)

// Daemon represents the core background process of refreSSH.
// It manages terminal sessions and distributes their output to connected clients.
type Daemon struct {
	config      *config.Config
	sessions    map[string]*Session
	broadcaster *Broadcaster
	mu          sync.RWMutex
}

// New creates a new Daemon instance with the provided configuration.
// If cfg is nil, a default configuration is used.
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

// Start launches the daemon, initializes the default terminal session, and starts the API server.
func (d *Daemon) Start() error {
	fmt.Println("Daemon starting...")

	// Create a default session using configured terminal
	s := NewSession("default", d.config.DefaultTerminal)
	if err := s.Start(); err != nil {
		// Fallback to sh or cmd.exe if the configured one fails
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

	// Connect broadcaster to PTY input
	d.broadcaster.SetInputWriter(s.pty)

	// Start broadcasting loop
	go d.broadcastLoop(s)

	return api.Start(d.config.Port)
}

func (d *Daemon) broadcastLoop(s *Session) {
	defer func() {
		s.mu.Lock()
		if s.pty != nil {
			_ = s.pty.Close()
		}
		if s.cmd != nil {
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
