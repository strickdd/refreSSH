package daemon

import (
	"fmt"
	"sync"

	"github.com/strickdd/refressh/internal/api"
	"github.com/strickdd/refressh/internal/config"
)

type Daemon struct {
	config      *config.Config
	sessions    map[string]*Session
	broadcaster *Broadcaster
	mu          sync.RWMutex
}

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
				return fmt.Errorf("failed to start any shell: %v", err)
			}
		}
	}

	d.mu.Lock()
	d.sessions[s.id] = s
	d.mu.Unlock()

	// Set the input writer so Broadcaster can forward input to the PTY
	d.broadcaster.SetInputWriter(s.pty)

	// Start broadcasting loop
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
