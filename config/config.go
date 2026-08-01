// Package config provides the functionalities for creating, reading and updating the app app configuration for the WSL2 environment.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type WSLConfig struct {
	DefaultLinuxDistro string   `json:"defaultLinuxDistro"`
	PinnedFolders      []string `json:"pinnedFolders"`
	BackgroundImage    string   `json:"backgroundImage"`
	BackgroundMode     string   `json:"backgroundMode"`
}

// DefaultWSLConfig provides a default WSLConfig.
func DefaultWSLConfig() *WSLConfig {
	return &WSLConfig{
		DefaultLinuxDistro: "default",
		PinnedFolders:      []string{},
		BackgroundImage:    "",
		BackgroundMode:     "gradient",
	}
}

// LoadWSLConfig loads the app configuration from the config.json file.
func LoadWSLConfig() (*WSLConfig, error) {
	cfg := DefaultWSLConfig()
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, SaveWSLConfig(cfg)
		}
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	normalize(cfg)
	return cfg, nil
}

// EnsureWSLConfig creates the config file when it is missing.
func EnsureWSLConfig() error {
	_, err := LoadWSLConfig()
	return err
}

// SaveWSLConfig saves the app configuration to the config.json file.
func SaveWSLConfig(cfg *WSLConfig) error {
	if cfg == nil {
		cfg = DefaultWSLConfig()
	}
	normalize(cfg)

	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func normalize(cfg *WSLConfig) {
	defaults := DefaultWSLConfig()
	if strings.TrimSpace(cfg.DefaultLinuxDistro) == "" {
		cfg.DefaultLinuxDistro = defaults.DefaultLinuxDistro
	}
	if strings.TrimSpace(cfg.BackgroundMode) == "" {
		cfg.BackgroundMode = defaults.BackgroundMode
	}
	if cfg.PinnedFolders == nil {
		cfg.PinnedFolders = []string{}
	}
}

// SetDefaultLinuxDistro gives users the ability to set their choice distro as default.
func SetDefaultLinuxDistro(distro string) error {
	config, err := LoadWSLConfig()
	if err != nil {
		return err
	}
	config.DefaultLinuxDistro = distro
	return SaveWSLConfig(config)
}

// TogglePinnedFolder adds or removes a folder path from the pinned list.
func TogglePinnedFolder(path string) (*WSLConfig, error) {
	config, err := LoadWSLConfig()
	if err != nil {
		return nil, err
	}
	for i, p := range config.PinnedFolders {
		if p == path {
			config.PinnedFolders = append(config.PinnedFolders[:i], config.PinnedFolders[i+1:]...)
			return config, SaveWSLConfig(config)
		}
	}
	config.PinnedFolders = append(config.PinnedFolders, path)
	return config, SaveWSLConfig(config)
}

// SetBackground persists the chosen background image and mode.
func SetBackground(image, mode string) (*WSLConfig, error) {
	config, err := LoadWSLConfig()
	if err != nil {
		return nil, err
	}
	config.BackgroundImage = image
	config.BackgroundMode = mode
	return config, SaveWSLConfig(config)
}

// Helper function to get the path to the config.json configuration file. It constructs the path based on the user's configuration directory, ensuring that the application can read and write its configuration in a consistent location.
func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "ihu", "config.json")
}
