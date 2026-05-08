package daemon

import (
	"io"
	"os/exec"
	"sync"
)

// PTY defines the abstract interface for terminal emulation.
// It allows refreSSH to support both real pseudo-terminals and fallback mechanisms.
type PTY interface {
	io.ReadWriteCloser
	// Resize updates the terminal dimensions.
	Resize(rows, cols uint16) error
}

// Session represents a single persistent terminal instance.
type Session struct {
	id      string
	command string
	args    []string
	pty     PTY
	cmd     *exec.Cmd
	running bool
	mu      sync.Mutex
}

// NewSession creates and initializes a new terminal Session with the specified parameters.
func NewSession(id string, command string, args ...string) *Session {
	return &Session{
		id:      id,
		command: command,
		args:    args,
	}
}
