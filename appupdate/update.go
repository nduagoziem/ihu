package appupdate

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/minio/selfupdate"
)

const (
	OWNER        = "nduagoziem"
	REPO         = "ihu"
	GITHUB_API   = "https://api.github.com"
	USER_AGENT   = "ihu-updater"
	MAX_BODY_MEG = 512
)

type Info struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	Available      bool   `json:"available"`
	ReleaseName    string `json:"releaseName"`
	ReleaseNotes   string `json:"releaseNotes"`
	ReleaseURL     string `json:"releaseUrl"`
	PublishedAt    string `json:"publishedAt"`
	AssetName      string `json:"assetName"`
	AssetURL       string `json:"assetUrl"`
	AssetSize      int64  `json:"assetSize"`
}

type Result struct {
	CurrentVersion  string `json:"currentVersion"`
	PreviousVersion string `json:"previousVersion"`
	RestartRequired bool   `json:"restartRequired"`
	Message         string `json:"message"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	HTMLURL     string        `json:"html_url"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func Check(currentVersion string) (*Info, error) {
	release, err := latestRelease()
	if err != nil {
		return nil, err
	}
	if release.Draft {
		return nil, errors.New("latest GitHub release is still a draft")
	}

	latest := strings.TrimSpace(release.TagName)
	info := &Info{
		CurrentVersion: currentVersion,
		LatestVersion:  latest,
		Available:      isUpdateAvailable(currentVersion, latest),
		ReleaseName:    release.Name,
		ReleaseNotes:   release.Body,
		ReleaseURL:     release.HTMLURL,
		PublishedAt:    release.PublishedAt,
	}

	if !info.Available {
		return info, nil
	}

	asset, err := chooseAsset(release.Assets)
	if err != nil {
		return nil, err
	}
	info.AssetName = asset.Name
	info.AssetURL = asset.BrowserDownloadURL
	info.AssetSize = asset.Size
	return info, nil
}

func Install(currentVersion string) (*Result, error) {
	info, err := Check(currentVersion)
	if err != nil {
		return nil, err
	}
	if !info.Available {
		return &Result{
			CurrentVersion:  currentVersion,
			PreviousVersion: currentVersion,
			RestartRequired: false,
			Message:         "Already up to date.",
		}, nil
	}

	resp, err := get(info.AssetURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download update: GitHub returned %s", resp.Status)
	}

	if strings.HasSuffix(strings.ToLower(info.AssetName), ".zip") {
		err = applyZip(resp.Body)
	} else {
		err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	}
	if err != nil {
		if rerr := selfupdate.RollbackError(err); rerr != nil {
			return nil, fmt.Errorf("update failed and rollback failed: %v; rollback: %w", err, rerr)
		}
		return nil, err
	}

	return &Result{
		CurrentVersion:  info.LatestVersion,
		PreviousVersion: currentVersion,
		RestartRequired: true,
		Message:         "Update installed. Restart required.",
	}, nil
}

func Restart() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Dir = filepath.Dir(exe)
	return cmd.Start()
}

func latestRelease() (*githubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", GITHUB_API, OWNER, REPO)
	resp, err := get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check update: GitHub returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, MAX_BODY_MEG<<20)).Decode(&release); err != nil {
		return nil, err
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, errors.New("latest GitHub release did not include a tag")
	}
	return &release, nil
}

func get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", USER_AGENT)

	client := &http.Client{Timeout: 2 * time.Minute}
	return client.Do(req)
}

func applyZip(r io.Reader) error {
	tmp, err := os.CreateTemp("", "ihu-update-*.zip")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	exeName := strings.ToLower(filepath.Base(os.Args[0]))
	for _, file := range zr.File {
		name := strings.ToLower(filepath.Base(file.Name))
		if file.FileInfo().IsDir() || !strings.HasSuffix(name, ".exe") {
			continue
		}
		if name != exeName && name != "ihu.exe" {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()
		return selfupdate.Apply(src, selfupdate.Options{})
	}
	return errors.New("update archive did not contain ihu.exe")
}

func chooseAsset(assets []githubAsset) (*githubAsset, error) {
	for _, pref := range []func(githubAsset) bool{
		func(a githubAsset) bool { return isZip(a.Name) && matchesPlatform(a.Name) },
		func(a githubAsset) bool { return isZip(a.Name) && mentionsWindows(a.Name) },
		func(a githubAsset) bool {
			return isExecutable(a.Name) && matchesPlatform(a.Name) && !looksLikeInstaller(a.Name)
		},
		func(a githubAsset) bool { return isExecutable(a.Name) && !looksLikeInstaller(a.Name) },
	} {
		for i := range assets {
			if pref(assets[i]) && strings.TrimSpace(assets[i].BrowserDownloadURL) != "" {
				return &assets[i], nil
			}
		}
	}
	return nil, fmt.Errorf("no compatible %s/%s update asset found", runtime.GOOS, runtime.GOARCH)
}

func isZip(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".zip")
}

func isExecutable(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".exe")
}

func mentionsWindows(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "windows") || strings.Contains(lower, "win")
}

func matchesPlatform(name string) bool {
	lower := strings.ToLower(name)
	return mentionsWindows(lower) && matchesArch(lower)
}

func matchesArch(name string) bool {
	for _, alias := range archAliases(runtime.GOARCH) {
		if strings.Contains(name, alias) {
			return true
		}
	}
	return false
}

func archAliases(arch string) []string {
	switch arch {
	case "amd64":
		return []string{"amd64", "x64", "x86_64"}
	case "386":
		return []string{"386", "x86", "ia32"}
	case "arm64":
		return []string{"arm64", "aarch64"}
	default:
		return []string{arch}
	}
}

func looksLikeInstaller(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "setup") || strings.Contains(lower, "installer") || strings.Contains(lower, "install")
}

func isUpdateAvailable(current, latest string) bool {
	current = normalizeVersion(current)
	latest = normalizeVersion(latest)
	if latest == "" {
		return false
	}
	if current == "" || current == "dev" {
		return true
	}
	return compareVersions(latest, current) > 0
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	v = strings.TrimPrefix(v, "version")
	v = strings.TrimSpace(v)
	return strings.TrimPrefix(v, "v")
}

func compareVersions(a, b string) int {
	ap := versionParts(a)
	bp := versionParts(b)
	for i := 0; i < len(ap) || i < len(bp); i++ {
		ai, bi := 0, 0
		if i < len(ap) {
			ai = ap[i]
		}
		if i < len(bp) {
			bi = bp[i]
		}
		if ai > bi {
			return 1
		}
		if ai < bi {
			return -1
		}
	}
	return 0
}

func versionParts(v string) []int {
	fields := strings.FieldsFunc(v, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == '+'
	})
	parts := make([]int, 0, len(fields))
	for _, field := range fields {
		n := leadingNumber(field)
		parts = append(parts, n)
	}
	return parts
}

func leadingNumber(s string) int {
	var b strings.Builder
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		b.WriteRune(r)
	}
	if b.Len() == 0 {
		return 0
	}
	n, _ := strconv.Atoi(b.String())
	return n
}
