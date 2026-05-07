//go:build windows
// +build windows

package daemon

import (
	"fmt"
)

type windowsPTY struct {
}

func (p *windowsPTY) Read(b []byte) (n int, err error) {
	return 0, fmt.Errorf("PTY not implemented on Windows yet")
}

func (p *windowsPTY) Write(b []byte) (n int, err error) {
	return 0, fmt.Errorf("PTY not implemented on Windows yet")
}

func (p *windowsPTY) Close() error {
	return nil
}

func (p *windowsPTY) Resize(rows, cols uint16) error {
	return nil
}

func (p *windowsPTY) Wait() error {
	return nil
}

func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// On Windows we might use conpty or just a plain pipe if we don't need full terminal emulation
	// For now, let's just use exec.Command and return an error if PTY is requested
	return fmt.Errorf("PTY not supported on Windows in this stub")
}
