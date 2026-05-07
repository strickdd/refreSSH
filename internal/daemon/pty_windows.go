//go:build windows
// +build windows

package daemon

import (
	"errors"
	"fmt"
	"io"
)

type pipePTY struct {
	in  io.WriteCloser
	out io.ReadCloser
}

func (p *pipePTY) Read(b []byte) (int, error)  { return p.out.Read(b) }
func (p *pipePTY) Write(b []byte) (int, error) { return p.in.Write(b) }
func (p *pipePTY) Close() error {
	_ = p.in.Close()
	return p.out.Close()
}
func (p *pipePTY) Resize(rows, cols uint16) error {
	return fmt.Errorf("not supported on Windows")
}
func (p *pipePTY) Wait() error { return nil }

func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return errors.New("session already running")
	}

	// On Windows, if we don't have a real PTY implementation yet,
	// we fall back to using pipes. This allows basic interaction.
	inPipe, err := s.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	outPipe, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	s.cmd.Stderr = s.cmd.Stdout

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	s.pty = &pipePTY{in: inPipe, out: outPipe}
	s.running = true
	return nil
}
