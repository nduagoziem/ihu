// Package terminal is a PTY (Pseudo-Terminal) that provides an interface for running linux commands directly on the app.
//
// The shell session here is different from the WSL2 shell session in boot/boot.go. The WSL2 shell session in boot/boot.go is a long-lived session that runs commands in the background for visualizing the WSL2 environment, while the terminal shell session is a short-lived session where users can run linux commands in the foreground.
//
// The session is backed by a Windows ConPTY (see conpty_windows.go). Attaching
// `wsl.exe` to a pseudoconsole gives it a real TTY, so WSL2 allocates a genuine
// Linux pty inside the distro. That is what makes Ctrl-C (SIGINT), isatty()
// detection, and \r-based progress output (curl, dnf) behave correctly —
// something the pipe-backed boot session cannot do.
package terminal

import (
	"context"
	"fmt"
	"strings"
	"sync"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

// Event names emitted to the frontend over the Wails runtime bus.
const (
	EventData = "terminal:data" // raw pseudoconsole output (ANSI/VT bytes) as a string
	EventExit = "terminal:exit" // the shell session ended
)

// Terminal owns a single live ConPTY-backed WSL shell session and streams its
// output to the frontend. Only one session is active at a time; starting a new
// one replaces any existing session.
type Terminal struct {
	ctx     context.Context
	mu      sync.Mutex
	pty     *conPTY
	running bool
}

// New creates a Terminal bound to the Wails application context used for events.
func New(ctx context.Context) *Terminal {
	return &Terminal{ctx: ctx}
}

// Start launches a fresh interactive bash session inside WSL, attached to a
// pseudoconsole of the given size. Any previously running session is torn down
// first. distro/user/elevated mirror the semantics of wsl.RunCommandAs; cwd, if
// set, becomes the shell's starting directory via `wsl.exe --cd`.
func (t *Terminal) Start(distro, user, cwd string, elevated bool, cols, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.pty != nil {
		t.pty.close()
		t.pty = nil
		t.running = false
	}

	commandLine := windows.ComposeCommandLine(buildArgs(distro, user, cwd, elevated))

	p, err := newConPTY(int16(cols), int16(rows))
	if err != nil {
		return err
	}
	if err := p.spawn(commandLine); err != nil {
		p.close()
		return err
	}

	t.pty = p
	t.running = true
	go t.readLoop(p)
	return nil
}

// Write forwards raw bytes (keystrokes, including Ctrl-C as 0x03) to the shell.
func (t *Terminal) Write(data string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pty == nil {
		return fmt.Errorf("terminal is not running")
	}
	_, err := t.pty.inW.WriteString(data)
	return err
}

// Resize updates the pseudoconsole viewport to match the frontend dimensions.
func (t *Terminal) Resize(cols, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pty == nil {
		return nil
	}
	return t.pty.resize(int16(cols), int16(rows))
}

// Stop terminates the current session, if any. Safe to call more than once.
func (t *Terminal) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.pty != nil {
		t.pty.close()
		t.pty = nil
		t.running = false
	}
	return nil
}

// readLoop pumps pseudoconsole output to the frontend until the pipe closes
// (shell exit, or teardown via Stop/Start). It then cleans up if this pty is
// still the active one and signals the frontend that the session ended.
func (t *Terminal) readLoop(p *conPTY) {
	buf := make([]byte, 4096)
	for {
		n, err := p.outR.Read(buf)
		if n > 0 {
			wruntime.EventsEmit(t.ctx, EventData, string(buf[:n]))
		}
		if err != nil {
			break
		}
	}

	t.mu.Lock()
	if t.pty == p {
		p.close()
		t.pty = nil
		t.running = false
	}
	t.mu.Unlock()

	wruntime.EventsEmit(t.ctx, EventExit)
}

// buildArgs assembles the wsl.exe argument vector (argv[0] included) mirroring
// wsl.RunCommandAs: an explicit distro when not the default, the resolved user,
// an optional starting directory, then a clean interactive bash.
func buildArgs(distro, user, cwd string, elevated bool) []string {
	args := []string{"wsl.exe"}
	if d := strings.TrimSpace(distro); d != "" && d != "default" {
		args = append(args, "-d", d)
	}
	if u := commandUser(user, elevated); u != "" {
		args = append(args, "-u", u)
	}
	if c := strings.TrimSpace(cwd); c != "" {
		args = append(args, "--cd", c)
	}
	args = append(args, "--", "bash", "--noprofile", "--norc", "-i")
	return args
}

// commandUser resolves the effective login user, matching wsl.commandUser.
func commandUser(user string, elevated bool) string {
	if elevated || user == "root" {
		return "root"
	}
	return strings.TrimSpace(user)
}
