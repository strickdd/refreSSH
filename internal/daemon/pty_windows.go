//go:build windows
// +build windows

package daemon

import (
	"fmt"
	"io"
	"os/exec"
)

type pipePTY struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (p *pipePTY) Read(b []byte) (int, error)  { return p.stdout.Read(b) }
func (p *pipePTY) Write(b []byte) (int, error) { return p.stdin.Write(b) }
func (p *pipePTY) Close() error {
	p.stdin.Close()
	return p.stdout.Close()
}

func (p *pipePTY) Resize(rows, cols uint16) error {
	return fmt.Errorf("resize not supported on windows pipes")
}

// Start launches the terminal process and uses standard pipes as a PTY fallback on Windows.
func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("session already running")
	}

	c := exec.Command(s.command, s.args...)

	stdin, err := c.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := c.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	s.cmd = c
	s.pty = &pipePTY{stdin: stdin, stdout: stdout}
	s.running = true

	return nil
}
