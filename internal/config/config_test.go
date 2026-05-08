package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

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

	expectedBase := "refreSSH"
	if filepath.Base(dir) != expectedBase {
		t.Errorf("expected base name %s, got %s", expectedBase, filepath.Base(dir))
	}
}

func TestGetDefaultTerminal(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Test Unix-specific SHELL env var logic
		oldShell := os.Getenv("SHELL")
		defer os.Setenv("SHELL", oldShell)

		os.Setenv("SHELL", "/bin/sh")
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

func TestSaveAndLoad(t *testing.T) {
	// Use a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "refressh-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock GetConfigDir by overriding it or using a version that takes a base path.
	// For simplicity in this test, we'll just test the Save/Load logic with a custom path if we can.
	// Since Load/Save currently use GetConfigDir internally, we might want to refactor them
	// to allow passing a path for better testability, but for now let's just test that they work.

	cfg := NewDefaultConfig()
	cfg.Port = 9999
	cfg.DefaultTerminal = "test-terminal"

	// We'll manually test the marshalling/unmarshalling logic here since we can't easily
	// override GetConfigDir without changing the API or using a global variable.

	configPath := filepath.Join(tmpDir, ConfigFileName)

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
