//go:build windows
// +build windows

package daemon

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

type winPTY struct {
	f *os.File
}

func (p *winPTY) Read(b []byte) (int, error)  { return p.f.Read(b) }
func (p *winPTY) Write(b []byte) (int, error) { return p.f.Write(b) }
func (p *winPTY) Close() error                { return p.f.Close() }

func (p *winPTY) Resize(rows, cols uint16) error {
	return pty.Setsize(p.f, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

type pipePTY struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (p *pipePTY) Read(b []byte) (int, error)  { return p.stdout.Read(b) }
func (p *pipePTY) Write(b []byte) (int, error) { return p.stdin.Write(b) }
func (p *pipePTY) Close() error {
	_ = p.stdin.Close() //nolint:errcheck
	return p.stdout.Close()
}
func (p *pipePTY) Resize(_, _ uint16) error { return fmt.Errorf("resize not supported on pipes") }

// Start launches the terminal process and attaches it to a Windows Pseudo Console (ConPTY),
// falling back to standard pipes if ConPTY is unavailable.
func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Running {
		return fmt.Errorf("session already running")
	}

	c := exec.Command(s.Command, s.Args...)
	f, err := pty.Start(c)
	if err == nil {
		s.cmd = c
		s.pty = &winPTY{f: f}
		s.Running = true
		return nil
	}

	// Fallback for CI/older Windows
	fmt.Printf("ConPTY failed (%v), falling back to pipes\n", err)
	
	stdin, err := c.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		_ = stdin.Close() //nolint:errcheck
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := c.Start(); err != nil {
		_ = stdin.Close()  //nolint:errcheck
		_ = stdout.Close() //nolint:errcheck
		return fmt.Errorf("failed to start process: %w", err)
	}

	s.cmd = c
	s.pty = &pipePTY{stdin: stdin, stdout: stdout}
	s.Running = true

	return nil
}
