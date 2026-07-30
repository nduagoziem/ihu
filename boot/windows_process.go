package boot

import (
	"os/exec"
	"syscall"
)

// HideTerminalApps disables a terminal window from popping up when the app is launched or used.
//
// Terminal apps usually pop up a window when the app is launched, which is not desirable for GUI apps. If they are exited, the app stops functioning because it uses the shell session under the hood.
//
// Better to keep them hidden and let the app manage them in the background.
func HideTerminalApps(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
