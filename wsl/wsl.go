// Package wsl exposes filesystem and distro operations through the live
// interactive WSL bash session managed by the boot package.
// This is the core of the application.
package wsl

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strconv"
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
	return ListDirAs(dir, "", "", false)
}

func ListDirAs(dir, distro, user string, elevated bool) ([]Entry, error) {

	if dir == "" {
		dir = "/"
	}

	out, err := RunCommandAs(listDirCmd(dir), distro, user, elevated)
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
func listDirCmd(dir string) string {
	quoted := shellQuote(dir)
	return "test -d " + quoted + " && find " + quoted + " -mindepth 1 -maxdepth 1 -printf '%f\\t%p\\t%y\\t%s\\t%TY-%Tm-%Td %TH:%TM\\t%M\\n'"
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
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 6 {
			continue
		}

		name := fields[0]
		if name == "" || name == "." || name == ".." {
			continue
		}

		size, _ := strconv.ParseInt(fields[3], 10, 64)
		full := fields[1]
		if full == "" {
			full = path.Join(dir, name)
		}
		entries = append(entries, Entry{
			Name:      name,
			Path:      full,
			IsDir:     fields[2] == "d",
			IsSymlink: fields[2] == "l",
			IsHidden:  strings.HasPrefix(name, "."),
			Size:      size,
			Modified:  fields[4],
			Perm:      fields[5],
		})
	}
	return entries, nil
}

type DefaultHome struct {
	User string `json:"user"`
	Home string `json:"home"`
}

// DefaultHomePath resolves the active WSL environment default user's name and home path.
//
// This makes the app boot from the home directory of the default login user managed by WSL2.
func DefaultHomePath() (*DefaultHome, error) {
	out, err := runSession("sh -lc 'printf \"%s\\t%s\\n\" \"$(whoami)\" \"$HOME\"'")
	if err != nil {
		return nil, err
	}
	fields := strings.SplitN(strings.TrimSpace(out), "\t", 2)
	if len(fields) != 2 || strings.TrimSpace(fields[0]) == "" || strings.TrimSpace(fields[1]) == "" {
		return nil, fmt.Errorf("could not resolve default WSL home directory")
	}
	return &DefaultHome{
		User: strings.TrimSpace(fields[0]),
		Home: strings.TrimSpace(fields[1]),
	}, nil
}

// HomePath resolves the home directory for a given user inside WSL.
func HomePath(user string) (string, error) {
	user = strings.TrimSpace(user)
	if user == "" {
		home, err := DefaultHomePath()
		if err != nil {
			return "", err
		}
		return home.Home, nil
	}

	if user == "root" {
		return "/root", nil
	}
	return "/home/" + user, nil
}

// ListDistros enumerates installed WSL distributions on Windows.
func ListDistros() ([]string, error) {

	cmd := exec.Command("wsl.exe", "-l", "-q")

	// This is necessary even though called in Boot(), because the command above doesn't go through boot.RunCommand().
	boot.HideTerminalApps(cmd)

	out, err := cmd.Output()
	if err != nil {
		return []string{"default"}, err
	}
	lines := splitNonEmpty(cleanWindowsCommandOutput(string(out)))
	if len(lines) == 0 {
		return []string{"default"}, nil
	}
	return lines, nil
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
	return ReadFileAs(p, "", "", false)
}

func ReadFileAs(p, distro, user string, elevated bool) (string, error) {
	return RunCommandAs("cat -- "+shellQuote(p)+" 2>/dev/null", distro, user, elevated)
}

// WriteFile replaces a text file's contents inside WSL.
func WriteFile(p, contents string) error {
	return WriteFileAs(p, contents, "", "", false)
}

func WriteFileAs(p, contents, distro, user string, elevated bool) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(contents))
	_, err := RunCommandAs("printf %s "+shellQuote(encoded)+" | base64 -d > "+shellQuote(p), distro, user, elevated)
	return err
}

// ReadFileBase64 returns a file's raw bytes encoded as base64 so binary
// content (images, PDFs, docx) can travel safely through the Wails bridge.
func ReadFileBase64(p string) (string, error) {
	return ReadFileBase64As(p, "", "", false)
}

func ReadFileBase64As(p, distro, user string, elevated bool) (string, error) {
	return RunCommandAs("base64 -w 0 -- "+shellQuote(p)+" 2>/dev/null", distro, user, elevated)
}

func RunCommandAs(cmd, distro, user string, elevated bool) (string, error) {
	args := []string{}
	if distro != "" && distro != "default" {
		args = append(args, "-d", distro)
	}
	if resolved := commandUser(user, elevated); resolved != "" {
		args = append(args, "-u", resolved)
	}
	args = append(args, "--", "sh", "-lc", cmd)

	wslCmd := exec.Command("wsl.exe", args...)

	// This is necessary even though called in Boot(), because the command above doesn't go through boot.RunCommand().
	boot.HideTerminalApps(wslCmd)

	out, err := wslCmd.CombinedOutput()
	clean := strings.TrimSpace(cleanCommandBytes(out))
	if err != nil && clean != "" {
		return clean, err
	}

	if elevated || user == "root" {
		return runSession("sudo -n sh -lc " + shellQuote(cmd))
	}
	return runSession(cmd)
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

func cleanWindowsCommandOutput(s string) string {
	s = strings.TrimPrefix(s, "\ufeff")
	return strings.ReplaceAll(s, "\x00", "")
}

func cleanCommandBytes(out []byte) string {
	out = bytes.TrimPrefix(out, []byte{0xff, 0xfe})
	out = bytes.TrimPrefix(out, []byte{0xfe, 0xff})
	return cleanWindowsCommandOutput(string(out))
}

func commandUser(user string, elevated bool) string {
	if elevated || user == "root" {
		return "root"
	}
	return strings.TrimSpace(user)
}
