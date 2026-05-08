package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	DefaultPort                = 8080
	DefaultAutoShutdownMinutes = 0 // 0 means no auto-shutdown
	ConfigFileName             = "config.json"
)

type Config struct {
	Port                int    `json:"port"`
	DefaultTerminal     string `json:"defaultTerminal"`
	PrimaryColor        string `json:"primaryColor"`
	AccentColor         string `json:"accentColor"`
	AutoShutdownMinutes int    `json:"autoShutdownMinutes"`
}

// NewDefaultConfig returns a Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		Port:                DefaultPort,
		DefaultTerminal:     getDefaultTerminal(),
		PrimaryColor:        "#4A90E2", // Default blue
		AccentColor:         "#F5A623", // Default orange
		AutoShutdownMinutes: DefaultAutoShutdownMinutes,
	}
}

func getDefaultTerminal() string {
	if runtime.GOOS == "windows" {
		// Check for pwsh, then powershell, then cmd
		paths := []string{"pwsh.exe", "powershell.exe", "cmd.exe"}
		for _, p := range paths {
			if _, err := exec.LookPath(p); err == nil {
				return p
			}
		}
		return "cmd.exe"
	}

	// On Unix-like systems, check SHELL environment variable first
	if shell := os.Getenv("SHELL"); shell != "" {
		if _, err := exec.LookPath(shell); err == nil {
			return shell
		}
	}

	// Fallback to common shells
	shells := []string{"bash", "zsh", "sh"}
	for _, s := range shells {
		if _, err := exec.LookPath(s); err == nil {
			return s
		}
	}

	return "/bin/sh"
}

// GetConfigDir returns the OS-standard configuration directory path.
func GetConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	return filepath.Join(configDir, "refreSSH"), nil
}

// Load loads the configuration from the standard path.
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, ConfigFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return NewDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := NewDefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to the standard path.
func (c *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create directory with restrictive permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Enforce directory permissions even if it already existed
	if err := os.Chmod(configDir, 0700); err != nil {
		// Log error but continue as this might fail on some filesystems
		fmt.Fprintf(os.Stderr, "Warning: failed to set permissions on config directory: %v\n", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := filepath.Join(configDir, ConfigFileName)

	// Write with restrictive permissions
	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Enforce file permissions even if it already existed
	if err := os.Chmod(configPath, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set permissions on config file: %v\n", err)
	}

	return nil
}
