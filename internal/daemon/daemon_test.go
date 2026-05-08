package daemon

import (
	"github.com/strickdd/refressh/internal/config"
	"testing"
)

func TestNewDaemon(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.Port = 1234
	cfg.DefaultTerminal = "test-term"

	d := New(cfg)

	if d.config.Port != 1234 {
		t.Errorf("expected port 1234, got %d", d.config.Port)
	}

	if d.config.DefaultTerminal != "test-term" {
		t.Errorf("expected terminal test-term, got %s", d.config.DefaultTerminal)
	}
}

func TestNewDaemonNilConfig(t *testing.T) {
	d := New(nil)

	if d.config == nil {
		t.Fatal("expected default config to be created when nil is passed")
	}

	if d.config.Port != config.DefaultPort {
		t.Errorf("expected default port %d, got %d", config.DefaultPort, d.config.Port)
	}
}
