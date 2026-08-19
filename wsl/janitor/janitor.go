// Package janitor reclaims disk space that the WSL2 virtual disk (ext4.vhdx)
// has borrowed from Windows but no longer needs.
//
// The vhdx grows as the Linux filesystem is used but never shrinks on its own,
// even after files are deleted inside the distro. Reclaiming space is a
// three-part dance:
//
//  1. fstrim — tell the Linux filesystem to mark deleted blocks as free so the
//     host can see them as unused.
//  2. wsl --shutdown — tear down the utility VM. Closing our bash session is not
//     enough; the VM keeps an exclusive handle on the vhdx until it is shut down.
//  3. diskpart compact — shrink the (now unlocked) vhdx back down. This needs
//     Administrator rights, so it is launched elevated via a UAC prompt.
//
// Afterwards the long-lived boot session is revived so the rest of the app keeps
// working.
package janitor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"ihu/boot"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// ProgressFunc receives human-readable status messages as a reclaim run
// proceeds. It may be nil.
type ProgressFunc func(string)

// Result summarizes a completed reclaim run. Sizes are the on-disk bytes of the
// vhdx file measured immediately before and after compaction.
type Result struct {
	VhdxPath    string `json:"vhdxPath"`
	BeforeBytes int64  `json:"beforeBytes"`
	AfterBytes  int64  `json:"afterBytes"`
	SavedBytes  int64  `json:"savedBytes"`
}

// unlockTimeout bounds how long we wait for the host to release the vhdx handle
// after shutting down the WSL VM.
const unlockTimeout = 20 * time.Second

// Clean trims and compacts the WSL virtual disk for the given distro, returning
// how much space was reclaimed. Pass an empty string (or "default") to target
// the default distribution. onProgress, if non-nil, receives status updates.
func Clean(distro string, onProgress ProgressFunc) (*Result, error) {
	report := func(format string, a ...any) {
		if onProgress != nil {
			onProgress(fmt.Sprintf(format, a...))
		}
	}

	vhdxPath, err := locateVhdx(distro)
	if err != nil {
		return nil, fmt.Errorf("could not locate WSL vhdx file: %w", err)
	}
	report("Found WSL virtual disk: %s", vhdxPath)

	// Step 1: free deleted blocks inside Linux so compaction can reclaim them.
	// Run Windows-side as root, independent of the bash session we tear down next.
	report("Trimming unused blocks inside Linux (fstrim)...")
	if out, ferr := runFstrim(distro); ferr != nil {
		// Non-fatal: compaction still reclaims previously-freed blocks.
		report("fstrim skipped: %v %s", ferr, out)
	} else if out != "" {
		report("%s", out)
	}

	// Step 2: end our bash session cleanly, then shut down the whole VM. Closing
	// the session alone does NOT release the vhdx handle — only --shutdown does.
	sess := boot.Session
	if sess != nil {
		report("Closing background WSL session...")
		sess.Close()
	}

	report("Shutting down the WSL virtual machine...")
	if err := runWSLShutdown(); err != nil {
		reviveSession(sess, report)
		return nil, fmt.Errorf("failed to shut down WSL: %w", err)
	}

	report("Waiting for the virtual disk to be released...")
	if err := waitForFileUnlocked(vhdxPath, unlockTimeout); err != nil {
		// The handle may still free up; surface it and let diskpart try anyway.
		report("warning: %v", err)
	}

	before := fileSize(vhdxPath)

	// Step 3: compact the vhdx via diskpart (elevated → triggers a UAC prompt).
	report("Compacting virtual disk (needs administrator; this can take a while)...")
	if err := compactVhdx(vhdxPath); err != nil {
		reviveSession(sess, report)
		return nil, fmt.Errorf("compaction failed: %w", err)
	}

	after := fileSize(vhdxPath)

	// Step 4: revive the SAME session struct so the stats poller, file browser,
	// and any cached *boot.BootWSL pointer (App.boot) keep working.
	reviveSession(sess, report)

	res := &Result{VhdxPath: vhdxPath, BeforeBytes: before, AfterBytes: after}
	if before > 0 && after > 0 && before > after {
		res.SavedBytes = before - after
	}
	report("Done. Reclaimed %s.", humanBytes(res.SavedBytes))
	return res, nil
}

// reviveSession restarts the same boot session struct that was closed, keeping
// the package-global boot.Session and any external pointer to it valid.
func reviveSession(sess *boot.BootWSL, report func(string, ...any)) {
	if sess == nil {
		return
	}
	report("Restarting background WSL session...")
	if _, err := sess.Boot(); err != nil {
		report("warning: could not restart WSL session: %v", err)
	}
}

// runFstrim marks freed blocks inside the target distro as unused. It runs as
// root through wsl.exe so it does not depend on the long-lived bash session.
func runFstrim(distro string) (string, error) {
	args := []string{}
	if d := strings.TrimSpace(distro); d != "" && d != "default" {
		args = append(args, "-d", d)
	}
	args = append(args, "-u", "root", "--", "fstrim", "-v", "/")

	cmd := exec.Command("wsl.exe", args...)
	boot.HideTerminalApps(cmd)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(cleanWSLBytes(out)), err
}

// runWSLShutdown terminates the WSL2 utility VM, releasing all vhdx handles.
func runWSLShutdown() error {
	cmd := exec.Command("wsl.exe", "--shutdown")
	boot.HideTerminalApps(cmd)
	return cmd.Run()
}

// waitForFileUnlocked polls until the vhdx can be opened with exclusive access
// (share mode 0), which is only possible once the VM has released its handle.
func waitForFileUnlocked(path string, timeout time.Duration) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(timeout)
	for {
		h, err := windows.CreateFile(
			p,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0, // exclusive: no sharing
			nil,
			windows.OPEN_EXISTING,
			windows.FILE_ATTRIBUTE_NORMAL,
			0,
		)
		if err == nil {
			_ = windows.CloseHandle(h)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("virtual disk still locked after %s: %w", timeout, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// compactVhdx shrinks the vhdx using a diskpart script, launched elevated.
func compactVhdx(path string) error {
	script := fmt.Sprintf(
		"select vdisk file=\"%s\"\r\nattach vdisk readonly\r\ncompact vdisk\r\ndetach vdisk\r\n",
		path,
	)

	tmp, err := os.CreateTemp("", "ihu_diskpart_*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	return runElevatedDiskpart(tmp.Name())
}

// runElevatedDiskpart runs `diskpart /s <script>` with Administrator rights.
// It shells out to PowerShell's Start-Process -Verb RunAs, which raises the UAC
// prompt, then -Wait/-PassThru so we block until diskpart exits and can forward
// its exit code. The unelevated app cannot run diskpart directly.
func runElevatedDiskpart(scriptPath string) error {
	psScript := fmt.Sprintf(
		"$ErrorActionPreference='Stop'; "+
			"$p = Start-Process -FilePath 'diskpart.exe' -ArgumentList @('/s','%s') "+
			"-Verb RunAs -WindowStyle Hidden -Wait -PassThru; exit $p.ExitCode",
		psEscapeSingleQuoted(scriptPath),
	)

	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", psScript)
	boot.HideTerminalApps(cmd)

	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// psEscapeSingleQuoted escapes a value for a PowerShell single-quoted string.
func psEscapeSingleQuoted(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

const lxssKeyPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Lxss`

// locateVhdx resolves the ext4.vhdx path for the given distro, preferring the
// authoritative WSL registry and falling back to the Store install layout.
func locateVhdx(distro string) (string, error) {
	if p, err := vhdxFromRegistry(distro); err == nil && p != "" {
		return p, nil
	}
	if p := vhdxFromGlob(); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("no ext4.vhdx found for distro %q", distro)
}

// vhdxFromRegistry looks up a distro's BasePath under HKCU\...\Lxss. An empty or
// "default" distro resolves via the DefaultDistribution GUID.
func vhdxFromRegistry(distro string) (string, error) {
	root, err := registry.OpenKey(registry.CURRENT_USER, lxssKeyPath, registry.READ)
	if err != nil {
		return "", err
	}
	defer root.Close()

	want := strings.TrimSpace(distro)
	if want == "" || want == "default" {
		if def, _, derr := root.GetStringValue("DefaultDistribution"); derr == nil {
			if p, perr := vhdxFromGUID(def); perr == nil {
				return p, nil
			}
		}
	}

	guids, err := root.ReadSubKeyNames(-1)
	if err != nil {
		return "", err
	}
	for _, guid := range guids {
		sub, err := registry.OpenKey(registry.CURRENT_USER, lxssKeyPath+`\`+guid, registry.READ)
		if err != nil {
			continue
		}
		name, _, nerr := sub.GetStringValue("DistributionName")
		base, _, berr := sub.GetStringValue("BasePath")
		sub.Close()
		if nerr != nil || berr != nil {
			continue
		}
		if want == "" || want == "default" || strings.EqualFold(name, want) {
			if p := vhdxPathFromBase(base); fileExists(p) {
				return p, nil
			}
		}
	}
	return "", fmt.Errorf("distribution %q not found in registry", distro)
}

// vhdxFromGUID resolves the vhdx path for a specific Lxss GUID subkey.
func vhdxFromGUID(guid string) (string, error) {
	sub, err := registry.OpenKey(registry.CURRENT_USER, lxssKeyPath+`\`+guid, registry.READ)
	if err != nil {
		return "", err
	}
	defer sub.Close()

	base, _, err := sub.GetStringValue("BasePath")
	if err != nil {
		return "", err
	}
	p := vhdxPathFromBase(base)
	if !fileExists(p) {
		return "", fmt.Errorf("vhdx not present at %s", p)
	}
	return p, nil
}

// vhdxPathFromBase joins a registry BasePath (which may carry a \\?\ prefix)
// with the standard vhdx filename.
func vhdxPathFromBase(base string) string {
	base = strings.TrimPrefix(base, `\\?\`)
	return filepath.Join(base, "ext4.vhdx")
}

// vhdxFromGlob is the fallback for Store-installed distros.
func vhdxFromGlob() string {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return ""
	}
	matches, err := filepath.Glob(filepath.Join(local, "Packages", "*", "LocalState", "ext4.vhdx"))
	if err != nil || len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func fileSize(p string) int64 {
	info, err := os.Stat(p)
	if err != nil {
		return 0
	}
	return info.Size()
}

// cleanWSLBytes strips the UTF-16/UTF-8 BOM and embedded NUL bytes that wsl.exe
// management output carries, yielding plain text.
func cleanWSLBytes(b []byte) string {
	b = bytes.TrimPrefix(b, []byte{0xff, 0xfe})
	b = bytes.TrimPrefix(b, []byte{0xfe, 0xff})
	s := strings.ReplaceAll(string(b), "\x00", "")
	return strings.TrimPrefix(s, "\ufeff")
}

// humanBytes renders a byte count in binary units (KiB, MiB, ...).
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
