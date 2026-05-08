// Package config_test contains tests for the config package.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestNewDefaultConfig verifies that NewDefaultConfig returns a config with expected defaults.
func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.Port != DefaultPort {
		t.Errorf("expected port %d, got %d", DefaultPort, cfg.Port)
	}
	if cfg.DefaultTerminal == "" {
		t.Error("expected default terminal to be set")
	}
	if cfg.PrimaryColor == "" {
		t.Error("expected primary color to be set")
	}
	if cfg.AccentColor == "" {
		t.Error("expected accent color to be set")
	}
}

// TestGetConfigDir verifies that GetConfigDir returns a valid absolute path.
func TestGetConfigDir(t *testing.T) {
	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("failed to get config dir: %v", err)
	}
	if dir == "" {
		t.Fatal("expected config dir to be non-empty")
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("expected absolute path, got %s", dir)
	}

	expectedBase := ".refreSSH"
	if filepath.Base(dir) != expectedBase {
		t.Errorf("expected base name %s, got %s", expectedBase, filepath.Base(dir))
	}
}

// TestGetDefaultTerminal verifies that a terminal is detected.
func TestGetDefaultTerminal(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Test Unix-specific SHELL env var logic
		oldShell := os.Getenv("SHELL")
		defer func() {
			if err := os.Setenv("SHELL", oldShell); err != nil {
				t.Errorf("failed to restore SHELL env var: %v", err)
			}
		}()

		if err := os.Setenv("SHELL", "/bin/sh"); err != nil {
			t.Fatalf("failed to set SHELL env var: %v", err)
		}
		term := getDefaultTerminal()
		if term == "" {
			t.Error("expected terminal to be non-empty")
		}
	} else {
		term := getDefaultTerminal()
		if term == "" {
			t.Error("expected terminal to be non-empty")
		}
	}
}

// TestSaveAndLoad verifies the config marshalling and unmarshalling.
func TestSaveAndLoad(t *testing.T) {
	// Use a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "refressh-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("failed to remove temp dir: %v", err)
		}
	}()

	cfg := NewDefaultConfig()
	cfg.Port = 9999
	cfg.DefaultTerminal = "test-terminal"

	configPath := filepath.Join(tmpDir, "config.json")

	// Test saving to custom path
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Test loading from custom path
	loadedData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var loadedCfg Config
	if err := json.Unmarshal(loadedData, &loadedCfg); err != nil {
		t.Fatalf("failed to parse config file: %v", err)
	}

	if loadedCfg.Port != cfg.Port {
		t.Errorf("expected port %d, got %d", cfg.Port, loadedCfg.Port)
	}
	if loadedCfg.DefaultTerminal != cfg.DefaultTerminal {
		t.Errorf("expected terminal %s, got %s", cfg.DefaultTerminal, loadedCfg.DefaultTerminal)
	}
}
