//go:build !windows
// +build !windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type unixPTY struct {
	f *os.File
}

func (p *unixPTY) Read(b []byte) (int, error)  { return p.f.Read(b) }
func (p *unixPTY) Write(b []byte) (int, error) { return p.f.Write(b) }
func (p *unixPTY) Close() error                { return p.f.Close() }

func (p *unixPTY) Resize(rows, cols uint16) error {
	return pty.Setsize(p.f, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

// Start launches the terminal process and attaches it to a PTY on Unix-like systems.
func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("session already running")
	}

	// #nosec G204 - Subprocess launching is the core purpose of refreSSH
	c := exec.Command(s.command, s.args...)
	f, err := pty.Start(c)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}

	s.cmd = c
	s.pty = &unixPTY{f: f}
	s.running = true

	return nil
}
