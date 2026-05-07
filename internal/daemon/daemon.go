package daemon

import (
	"fmt"
	"github.com/strickdd/refressh/internal/api"
	"sync"
)

type Daemon struct {
	sessions    map[string]*Session
	broadcaster *Broadcaster
	mu          sync.RWMutex
}

func New() *Daemon {
	return &Daemon{
		sessions:    make(map[string]*Session),
		broadcaster: NewBroadcaster(),
	}
}

func (d *Daemon) Start() error {
	fmt.Println("Daemon starting...")

	// Create a default session for now
	s := NewSession("default", "bash")
	if err := s.Start(); err != nil {
		// Fallback to sh or cmd.exe
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

	// Start broadcasting loop
	go d.broadcastLoop(s)

	return api.Start()
}

func (d *Daemon) broadcastLoop(s *Session) {
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
