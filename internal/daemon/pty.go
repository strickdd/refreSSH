package daemon

import (
	"io"
	"os/exec"
	"sync"
)

// PTY defines the interface for a Pseudo-Terminal or its fallback.
type PTY interface {
	io.ReadWriteCloser
	Resize(rows, cols uint16) error
}

// Session represents a single terminal session managed by the daemon.
type Session struct {
	id      string
	command string
	args    []string
	pty     PTY
	cmd     *exec.Cmd
	running bool
	mu      sync.Mutex
}

// NewSession creates a new terminal session with the given ID and command.
func NewSession(id string, command string, args ...string) *Session {
	return &Session{
		id:      id,
		command: command,
		args:    args,
	}
}
