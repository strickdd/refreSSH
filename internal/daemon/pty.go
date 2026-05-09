package daemon

import (
	"io"
	"os/exec"
	"sync"
)

// PTY defines the abstract interface for terminal emulation in refreSSH.
type PTY interface {
	io.ReadWriteCloser
	// Resize sets the terminal width and height.
	Resize(rows, cols uint16) error
}

// Session represents a unique, persistent terminal session managed by the daemon.
type Session struct {
	id      string
	command string
	args    []string
	pty     PTY
	cmd     *exec.Cmd
	running bool
	mu      sync.Mutex
}

// NewSession creates and returns a new terminal Session with the specified ID and command.
func NewSession(id string, command string, args ...string) *Session {
	return &Session{
		id:      id,
		command: command,
		args:    args,
	}
}
