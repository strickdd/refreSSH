package daemon

import (
	"io"
	"os/exec"
	"sync"
	"time"
)

// PTY defines the abstract interface for terminal emulation in refreSSH.
type PTY interface {
	io.ReadWriteCloser
	// Resize sets the terminal width and height.
	Resize(rows, cols uint16) error
}

// Session represents a unique, persistent terminal session managed by the daemon.
type Session struct {
	ID          string       `json:"id"`
	Command     string       `json:"command"`
	Args        []string     `json:"args"`
	StartTime   time.Time    `json:"start_time"`
	Running     bool         `json:"running"`
	Broadcaster *Broadcaster `json:"-"`
	pty         PTY
	cmd         *exec.Cmd
	mu          sync.Mutex
}

// NewSession creates and returns a new terminal Session with the specified ID and command.
func NewSession(id string, command string, args ...string) *Session {
	return &Session{
		ID:          id,
		Command:     command,
		Args:        args,
		StartTime:   time.Now(),
		Broadcaster: NewBroadcaster(),
	}
}
