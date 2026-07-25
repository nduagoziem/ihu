package main

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"time"
=======
>>>>>>> 2e4cadf (Refactor WSL session handling in app and wsl packages)

	"ihu/boot"
	"ihu/config"
	"ihu/wsl"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
<<<<<<< HEAD
	sess := &boot.BootWSL{}
	if _, err := sess.Boot(); err != nil {
=======
	if _, err := a.boot.Boot(); err != nil {
>>>>>>> 2e4cadf (Refactor WSL session handling in app and wsl packages)
		fmt.Println("boot:", err)
	}
}

// shutdown closes the live WSL session when the window is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.boot != nil {
		a.boot.Close()
	}
}

// GetBootData returns system statistics and a timestamp for the welcome screen.
func (a *App) GetBootData() boot.BootData {
	return boot.GetBootData(a.boot)
}

// --- Config ---------------------------------------------------------------

// GetConfig returns the persisted app configuration.
func (a *App) GetConfig() (*config.WSLConfig, error) {
	return config.LoadWSLConfig()
}

// SetWelcomeDisabled toggles whether the welcome screen shows at startup.
func (a *App) SetWelcomeDisabled(disabled bool) (*config.WSLConfig, error) {
	if err := config.SetWelcomeDisabled(disabled); err != nil {
		return nil, err
	}
	return config.LoadWSLConfig()
}

// SetDefaultLinuxPath persists the chosen default landing directory.
func (a *App) SetDefaultLinuxPath(path string) (*config.WSLConfig, error) {
	if err := config.SetDefaultLinuxPath(path); err != nil {
		return nil, err
	}
	return config.LoadWSLConfig()
}

// SetDefaultLinuxUser persists the chosen default user.
func (a *App) SetDefaultLinuxUser(user string) (*config.WSLConfig, error) {
	if err := config.SetDefaultLinuxUser(user); err != nil {
		return nil, err
	}
	return config.LoadWSLConfig()
}

// SetDefaultLinuxDistro persists the chosen default distro.
func (a *App) SetDefaultLinuxDistro(distro string) (*config.WSLConfig, error) {
	if err := config.SetDefaultLinuxDistro(distro); err != nil {
		return nil, err
	}
	return config.LoadWSLConfig()
}

// TogglePinnedFolder adds or removes a folder from the pinned list.
func (a *App) TogglePinnedFolder(path string) (*config.WSLConfig, error) {
	return config.TogglePinnedFolder(path)
}

// SetBackground persists the chosen background image and mode.
func (a *App) SetBackground(image, mode string) (*config.WSLConfig, error) {
	return config.SetBackground(image, mode)
}

// --- Filesystem -----------------------------------------------------------

// ListDir returns the entries within a WSL directory.
func (a *App) ListDir(dir string) ([]wsl.Entry, error) {
	return wsl.ListDir(a.boot, dir)
}

// HomePath resolves the home directory for a user.
func (a *App) HomePath(user string) (string, error) {
	return wsl.HomePath(a.boot, user)
}

// ListDistros enumerates installed WSL distributions.
func (a *App) ListDistros() ([]string, error) {
	return wsl.ListDistros(a.boot)
}

// ListUsers enumerates login users on the system.
func (a *App) ListUsers() ([]string, error) {
	return wsl.ListUsers(a.boot)
}

// ReadFile returns the contents of a file for the editor/viewer.
func (a *App) ReadFile(path string) (string, error) {
<<<<<<< HEAD
	return wsl.ReadFile(path)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
=======
	return wsl.ReadFile(a.boot, path)
}

// RunWSLCommand sends a command to the long-lived WSL shell and returns stdout/stderr.
func (a *App) RunWSLCommand(command string) (string, error) {
	if a.boot == nil {
		return "", fmt.Errorf("wsl session is not initialized")
	}
	return a.boot.RunCommand(command)
}

func (a *App) Greet() string {
	return "Good Morning from Go"
}

// func (a *App) Greet(name string) string {
// 	return fmt.Sprintf("Hello %s, It's show time!", name)
// }
>>>>>>> 2e4cadf (Refactor WSL session handling in app and wsl packages)
