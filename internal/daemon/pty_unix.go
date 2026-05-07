// +build !windows

package daemon

import (
	"github.com/creack/pty"
	"os"
)

type unixPTY struct {
	f *os.File
}

func (p *unixPTY) Read(b []byte) (n int, err error) {
	return p.f.Read(b)
}

func (p *unixPTY) Write(b []byte) (n int, err error) {
	return p.f.Write(b)
}

func (p *unixPTY) Close() error {
	return p.f.Close()
}

func (p *unixPTY) Resize(rows, cols uint16) error {
	return p.Setsize(rows, cols)
}

func (p *unixPTY) Setsize(rows, cols uint16) error {
	return pty.Setsize(p.f, &pty.Winsize{
		Rows: rows,
		Cols: cols,
	})
}

func (p *unixPTY) Wait() error {
	// This is usually handled by the cmd.Wait() in Session
	return nil
}

func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := pty.Start(s.cmd)
	if err != nil {
		return err
	}

	s.pty = &unixPTY{f: f}
	s.running = true
	return nil
}
