// boot.go
package boot

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// BootWSL represents the WSL boot process, managing long-lived streams.
type BootWSL struct {
	wslCmd *exec.Cmd
	input  io.WriteCloser
	reader *bufio.Reader
	mu     sync.Mutex // Prevents concurrent writes/reads on the single shell instance
}

// BootData contains the necessary WSL data returned to the frontend at app launch.
type BootData struct {
	SystemStats *SystemStats `json:"systemStats"`
	BootedAt    string       `json:"bootedAt"`
}

// Global session instance so systemstats.go can access it cleanly
var Session *BootWSL

// Boot starts a persistent interactive bash shell session inside WSL.
func (b *BootWSL) Boot() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Launching a clean interactive shell profile
		cmd = exec.Command("wsl.exe", "sh", "-lc", "exec bash --noprofile --norc")
	} else {
		cmd = exec.Command("bash", "--noprofile", "--norc")
	}

	inputPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	// Redirect Stderr to Stdout so we don't lock up on unexpected error channels
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", err
	}

	b.wslCmd = cmd
	b.input = inputPipe
	b.reader = bufio.NewReader(outputPipe)
	Session = b // Set the global singleton

	return "WSL Session initialized successfully", nil
}

// RunCommand sends a command to the live shell session and blocks until it reads the full response.
func (b *BootWSL) RunCommand(cmdStr string) (string, error) {
	if b.input == nil || b.reader == nil {
		return "", fmt.Errorf("wsl session is not running")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Unique token to detect when this specific command concludes
	delimiter := "___CMD_DONE___"

	// Format: Execute the command, then echo our delimiter token on a new line
	fmt.Fprint(b.input, cmdStr+"\n")
	fmt.Fprint(b.input, "echo "+delimiter+"\n")

	var outputLines []string
	for {
		line, err := b.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		cleanLine := strings.TrimSpace(line)
		if cleanLine == delimiter {
			break
		}

		// Collect raw output lines, ignoring the command echo match itself if reflected
		if !strings.Contains(line, delimiter) {
			outputLines = append(outputLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(outputLines, "")), nil
}

// Close gracefully terminates the running WSL instance.
func (b *BootWSL) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.wslCmd == nil {
		return
	}

	if b.input != nil {
		_, _ = fmt.Fprint(b.input, "exit\n")
		_ = b.input.Close()
	}

	done := make(chan error, 1)
	go func() {
		done <- b.wslCmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		if b.wslCmd.Process != nil {
			_ = b.wslCmd.Process.Kill()
		}
		<-done
	}

	b.wslCmd = nil
	b.input = nil
	b.reader = nil
	if Session == b {
		Session = nil
	}
}

func GetBootData() BootData {
	return BootData{
		SystemStats: GetStats(),
		BootedAt:    time.Now().Local().Format(time.RFC1123),
	}
}
