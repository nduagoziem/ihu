package main

import (
	"context"
	"fmt"
<<<<<<< HEAD
	"time"

	"ihu/boot"
	"ihu/config"
	"ihu/wsl"
=======
	"ihu/boot"
	"ihu/config"
	"log"
>>>>>>> 9bc2b59 (wip: progress on wsl boot and session.)
)

// App struct
type App struct {
	ctx  context.Context
	boot *boot.BootWSL
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		boot: &boot.BootWSL{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
<<<<<<< HEAD
	sess := &boot.BootWSL{}
	if _, err := sess.Boot(); err != nil {
		fmt.Println("boot:", err)
	}
}

// shutdown closes the live WSL session when the window is closing.
func (a *App) shutdown(ctx context.Context) {
	if boot.Session != nil {
		boot.Session.Close()
	}
}

// BootData returns the launch-time stats used by the welcome screen.
type BootData struct {
	SystemStats *boot.SystemStats `json:"systemStats"`
	BootedAt    string            `json:"bootedAt"`
}

// GetBootData returns system statistics and a timestamp for the welcome screen.
func (a *App) GetBootData() BootData {
	return BootData{
		SystemStats: boot.GetStats(),
		BootedAt:    time.Now().Local().Format(time.RFC1123),
	}
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
	return wsl.ListDir(dir)
}

// HomePath resolves the home directory for a user.
func (a *App) HomePath(user string) (string, error) {
	return wsl.HomePath(user)
}

// ListDistros enumerates installed WSL distributions.
func (a *App) ListDistros() ([]string, error) {
	return wsl.ListDistros()
}

// ListUsers enumerates login users on the system.
func (a *App) ListUsers() ([]string, error) {
	return wsl.ListUsers()
}

// ReadFile returns the contents of a file for the editor/viewer.
func (a *App) ReadFile(path string) (string, error) {
	return wsl.ReadFile(path)
=======

	if _, err := a.boot.Boot(); err != nil {
		log.Printf("failed to boot WSL session: %v", err)
	}
>>>>>>> 9bc2b59 (wip: progress on wsl boot and session.)
}

// The shutdown method will be called by Wails right at the end of the shutdown process.
// This is a good place to deallocate memory and perform any shutdown tasks.
func (a *App) shutdown(ctx context.Context) {
	if a.boot != nil {
		a.boot.Close()
	}
}

// GetBootData returns the WSL stats used by the UI.
func (a *App) GetBootData() boot.BootData {
	return boot.GetBootData(a.boot)
}

// RunWSLCommand sends a command to the long-lived WSL shell and returns stdout/stderr.
func (a *App) RunWSLCommand(command string) (string, error) {
	if a.boot == nil {
		return "", fmt.Errorf("wsl session is not initialized")
	}
	return a.boot.RunCommand(command)
}

// SetWelcomeDisabled updates the configuration to disable the welcome screen on app launch.
func (a *App) SetWelcomeDisabled(disabled bool) error {
	return config.SetWelcomeDisabled(disabled)
}

func (a *App) Greet() string {
	return "Good Morning from Go"
}

// func (a *App) Greet(name string) string {
// 	return fmt.Sprintf("Hello %s, It's show time!", name)
// }
