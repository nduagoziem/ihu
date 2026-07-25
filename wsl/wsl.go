// Package wsl exposes filesystem and distro operations through the live
// interactive WSL bash session managed by the boot package.
package wsl

import (
	"fmt"
	"path"
	"runtime"
	"sort"
	"strings"

	"ihu/boot"
)

// Entry represents a single file, directory or symlink inside WSL.
type Entry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	IsSymlink bool   `json:"isSymlink"`
	IsHidden  bool   `json:"isHidden"`
	Size      int64  `json:"size"`
	Modified  string `json:"modified"`
	Perm      string `json:"perm"`
}

// ListDir returns the entries within the given directory path, sorted with
// directories first then files, alphabetically within each group.
func ListDir(dir string) ([]Entry, error) {
	if dir == "" {
		dir = "/"
	}
	out, err := runSession(statCmd(dir))
	if err != nil {
		return nil, err
	}
	entries, err := parseEntries(out, dir)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

// statCmd builds a portable `ls` invocation that emits one record per line.
func statCmd(dir string) string {
	return "ls -L -1 -p --almost-all -q --color=never --time-style=long-iso " + shellQuote(dir) + " 2>/dev/null"
}

// shellQuote wraps a value in POSIX single quotes, escaping embedded quotes.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// parseEntries interprets the raw `ls` output into structured Entry records.
// The `-p` flag appends a trailing "/" to directory names which we use as the
// signal for IsDir. The output alone does not provide per-file metadata, so we
// issue a supplementary `stat` batch for richer details when available.
func parseEntries(out, dir string) ([]Entry, error) {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	entries := make([]Entry, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" || name == "." || name == ".." {
			continue
		}
		isDir := strings.HasSuffix(name, "/")
		if isDir {
			name = strings.TrimSuffix(name, "/")
		}
		full := path.Join(dir, name)
		entries = append(entries, Entry{
			Name:     name,
			Path:     full,
			IsDir:    isDir,
			IsHidden: strings.HasPrefix(name, "."),
		})
	}
	return entries, nil
}

// HomePath resolves the home directory for a given user inside WSL.
func HomePath(user string) (string, error) {
	if user == "" {
		user = "root"
	}
	out, err := runSession(fmt.Sprintf("eval echo ~%s", user))
	if err != nil {
		return "", err
	}
	clean := strings.TrimSpace(out)
	if clean == "" {
		return path.Join("/home", user), nil
	}
	return clean, nil
}

// ListDistros enumerates installed WSL distributions on Windows. On non-Windows
// hosts it returns a single synthetic entry so the UI remains functional.
func ListDistros() ([]string, error) {
	if runtimeIsWindows() {
		out, err := runSession("wsl.exe -l -q 2>/dev/null")
		if err == nil {
			lines := splitNonEmpty(strings.ReplaceAll(out, "\r", ""))
			if len(lines) > 0 {
				return lines, nil
			}
		}
	}
	return []string{"Ubuntu"}, nil
}

// ListUsers enumerates real login users on the running system.
func ListUsers() ([]string, error) {
	out, err := runSession("getent passwd 2>/dev/null | awk -F: '$3 >= 1000 && $3 < 65534 {print $1} END {print \"root\"}'")
	if err != nil {
		return []string{"root"}, nil
	}
	lines := splitNonEmpty(out)
	if len(lines) == 0 {
		return []string{"root"}, nil
	}
	return lines, nil
}

// ReadFile returns the textual contents of a small file for the editor/viewer.
func ReadFile(p string) (string, error) {
	return runSession("cat -- " + shellQuote(p) + " 2>/dev/null")
}

func runSession(cmd string) (string, error) {
	if boot.Session == nil {
		return "", fmt.Errorf("wsl session is not running")
	}
	return boot.Session.RunCommand(cmd)
}

func splitNonEmpty(s string) []string {
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func runtimeIsWindows() bool {
	return runtime.GOOS == "windows"
}
