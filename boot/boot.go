// Package boot manages the lifecycle of the WSL2 environment. It provides three main functionalities: booting a long-lived WSL session, running commands in that session, and closing the session.
package boot

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// BootWSL represents the WSL boot process, managing long-lived streams.
type BootWSL struct {
	wslCmd       *exec.Cmd
	input        io.WriteCloser
	outputWriter io.Closer
	reader       *bufio.Reader
	mu           sync.Mutex // Prevents concurrent writes/reads on the single shell instance
}

// Global session instance so system_stats.go can access it cleanly
var Session *BootWSL

// Boot starts a persistent interactive bash shell session inside WSL.
func (b *BootWSL) Boot() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var cmd *exec.Cmd
	// Launching a clean interactive shell profile
	cmd = exec.Command("wsl.exe", "sh", "-lc", "exec bash --noprofile --norc")

	HideTerminalApps(cmd)

	inputPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	outputReader, outputWriter := io.Pipe()
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter

	if err := cmd.Start(); err != nil {
		_ = outputWriter.Close()
		return "", err
	}

	b.wslCmd = cmd
	b.input = inputPipe
	b.outputWriter = outputWriter
	b.reader = bufio.NewReader(outputReader)
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

	delimiter := "___IHU_CMD_DONE_" + randomToken() + "___"

	fmt.Fprint(b.input, cmdStr+"\n")
	fmt.Fprintf(b.input, "printf '\\n%s:%%s\\n' $?\n", delimiter)

	var outputLines []string
	exitCode := 0
	for {
		line, err := b.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		cleanLine := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(cleanLine, delimiter+":"); ok {
			code, err := strconv.Atoi(after)
			if err == nil {
				exitCode = code
			}
			break
		}

		outputLines = append(outputLines, line)
	}

	out := strings.TrimSpace(strings.Join(outputLines, ""))
	if exitCode != 0 {
		if out == "" {
			out = fmt.Sprintf("command exited with status %d", exitCode)
		}
		return out, errors.New(out)
	}
	return out, nil
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
	if b.outputWriter != nil {
		_ = b.outputWriter.Close()
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
	b.outputWriter = nil
	b.reader = nil
	if Session == b {
		Session = nil
	}
}

// Random delimiter token generator.
func randomToken() string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)
}
