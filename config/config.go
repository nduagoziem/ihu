package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type WSLConfig struct {
	WelcomeDisabled    bool     `json:"welcomeDisabled"`
	DefaultLinuxPath   string   `json:"defaultLinuxPath"`
	DefaultLinuxUser   string   `json:"defaultLinuxUser"`
	DefaultLinuxDistro string   `json:"defaultLinuxDistro"`
	PinnedFolders      []string `json:"pinnedFolders"`
	BackgroundImage    string   `json:"backgroundImage"`
	BackgroundMode     string   `json:"backgroundMode"`
}

// DefaultWSLConfig provides a default WSLConfig.
func DefaultWSLConfig() *WSLConfig {
	return &WSLConfig{
		DefaultLinuxPath:   "/root",
		DefaultLinuxUser:   "root",
		DefaultLinuxDistro: "",
		WelcomeDisabled:    false,
		PinnedFolders:      []string{},
		BackgroundImage:    "",
		BackgroundMode:     "gradient",
	}
}

// LoadWSLConfig loads the app configuration from the config.json file. If the file does not exist or is invalid, it returns the default configuration.
func LoadWSLConfig() (*WSLConfig, error) {
	config := DefaultWSLConfig()

	data, err := os.ReadFile(configPath())

	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveWSLConfig saves the app configuration to the config.json file. It ensures that the configuration directory exists and writes the configuration in a pretty-printed JSON format.
func SaveWSLConfig(config *WSLConfig) error {

	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// SetDefaultLinuxPath gives users the ability to set their choice directory in the app homepage.
func SetDefaultLinuxPath(path string) error {
	config, err := LoadWSLConfig()
	if err != nil {
		return err
	}
	config.DefaultLinuxPath = path
	return SaveWSLConfig(config)
}

// SetDefaultLinuxUser gives users the ability to set their choice user as default.
func SetDefaultLinuxUser(user string) error {
	config, err := LoadWSLConfig()
	if err != nil {
		return err
	}
	config.DefaultLinuxUser = user
	return SaveWSLConfig(config)
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

// SetWelcomeDisabled gives users the ability to disable the welcome message on app launch.
func SetWelcomeDisabled(disabled bool) error {
	config, err := LoadWSLConfig()
	if err != nil {
		return err
	}
	config.WelcomeDisabled = disabled
	return SaveWSLConfig(config)
}

// Helper function to get the path to the config.json configuration file. It constructs the path based on the user's configuration directory, ensuring that the application can read and write its configuration in a consistent location.
func configPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = "."
	}
	return filepath.Join(dir, "ihu", "config.json")
}
