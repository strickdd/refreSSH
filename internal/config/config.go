// Package config handles the persistent configuration for refreSSH.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// DefaultPort is the default port the API server listens on.
	DefaultPort = 42069
	// DefaultAutoShutdownMinutes is the default time before the daemon shuts down automatically (0 = never).
	DefaultAutoShutdownMinutes = 0
)

// Config represents the application configuration schema.
type Config struct {
	// Port is the local TCP port the API server listens on.
	Port int `json:"port"`
	// DefaultTerminal is the path to the shell/terminal to use if not specified.
	DefaultTerminal string `json:"defaultTerminal"`
	// PrimaryColor is the hex color code for primary UI elements.
	PrimaryColor string `json:"primaryColor"`
	// AccentColor is the hex color code for accent UI elements.
	AccentColor string `json:"accentColor"`
	// AutoShutdownMinutes is the idle time in minutes before the daemon shuts down.
	AutoShutdownMinutes int `json:"autoShutdownMinutes"`
}

// NewDefaultConfig returns a Config struct with reasonable default values.
func NewDefaultConfig() *Config {
	return &Config{
		Port:                DefaultPort,
		DefaultTerminal:     getDefaultTerminal(),
		PrimaryColor:        "#43ED0F",
		AccentColor:         "#ED0F43",
		AutoShutdownMinutes: DefaultAutoShutdownMinutes,
	}
}

func getDefaultTerminal() string {
	if runtime.GOOS == "windows" {
		return "pwsh.exe"
	}
	// On Unix, try $SHELL before falling back to sh
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "sh"
}

var configDirOverride string

// SetConfigDirOverride sets a custom directory for configuration files, primarily for testing.
func SetConfigDirOverride(dir string) {
	configDirOverride = dir
}

// GetConfigDir returns the platform-specific directory for refreSSH configuration.
// Windows: AppData\.refreSSH, Unix: ~/.refreSSH
func GetConfigDir() (string, error) {
	if configDirOverride != "" {
		return configDirOverride, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	var configDir string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, ".refreSSH")
	} else {
		configDir = filepath.Join(home, ".refreSSH")
	}

	return configDir, nil
}

// GetAPIToken returns the local API token, generating one if it doesn't exist.
func GetAPIToken() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	tokenPath := filepath.Join(configDir, "token")

	// 1. Try to read existing token
	data, err := os.ReadFile(filepath.Clean(tokenPath))
	if err == nil {
		return string(data), nil
	}

	// 2. Generate new random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	token := hex.EncodeToString(b)

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	// 3. Save token with secure permissions
	if err := os.WriteFile(tokenPath, []byte(token), 0600); err != nil {
		return "", fmt.Errorf("failed to save API token: %w", err)
	}

	return token, nil
}

// Load retrieves the configuration from the standard OS location.
// If the file does not exist, it creates a default configuration.
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")

	// Ensure directory exists with secure permissions (0700)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	// Address G304: Potential file inclusion via variable
	configPath = filepath.Clean(configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Save default config if not found
			cfg := NewDefaultConfig()
			if err := cfg.Save(); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save persists the configuration to the standard OS location.
func (c *Config) Save() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file with secure permissions (0600)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
