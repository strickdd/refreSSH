package daemon

import (
	"io"
	"os/exec"
	"sync"
)

// PTY represents a pseudo-terminal session
type PTY interface {
	io.ReadWriteCloser
	Resize(rows, cols uint16) error
	Wait() error
}

// Session manages a single PTY and its associated process
type Session struct {
	id      string
	cmd     *exec.Cmd
	pty     PTY
	mu      sync.Mutex
	running bool
}

func NewSession(id string, command string, args ...string) *Session {
	return &Session{
		id:  id,
		cmd: exec.Command(command, args...),
	}
}
