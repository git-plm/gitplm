package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/git-plm/gitplm/releases/latest"
	githubRelURL = "https://github.com/git-plm/gitplm/releases/latest/download"
)

// GitHubRelease represents the GitHub API release response
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
}

// CheckForUpdate checks if a newer version is available, at most once per day.
// Returns a user-facing message if an update is available, or empty string otherwise.
// Errors are silently ignored.
func CheckForUpdate(currentVersion string) string {
	current := strings.TrimPrefix(currentVersion, "v")
	if current == "Development" || current == "" {
		return ""
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	tsFile := filepath.Join(home, ".gitplm-last-update-check")

	if data, err := os.ReadFile(tsFile); err == nil {
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(string(data))); err == nil {
			if time.Since(t) < 24*time.Hour {
				return ""
			}
		}
	}

	latestVersion, err := getLatestVersion()
	if err != nil {
		return ""
	}

	// Write timestamp regardless of version comparison
	_ = os.WriteFile(tsFile, []byte(time.Now().Format(time.RFC3339)), 0644)

	latest := strings.TrimPrefix(latestVersion, "v")
	if latest != current {
		return fmt.Sprintf("A new version of gitplm is available (v%s). Run 'gitplm -update' to upgrade.", latest)
	}

	return ""
}

// Update checks for and downloads the latest version of gitplm
func Update(currentVersion string) error {
	fmt.Println("Checking for updates...")

	latestVersion, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	if current == latest {
		fmt.Printf("Already running the latest version (%s)\n", currentVersion)
		return nil
	}

	if current == "Development" {
		fmt.Printf("Running development version. Latest release is %s\n", latestVersion)
		fmt.Println("Proceeding with update...")
	} else {
		fmt.Printf("Updating from %s to %s\n", currentVersion, latestVersion)
	}

	if err := downloadAndInstall(latestVersion); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latestVersion)
	return nil
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func downloadAndInstall(version string) error {
	binaryName := getBinaryName(version)
	downloadURL := fmt.Sprintf("%s/%s", githubRelURL, binaryName)

	fmt.Printf("Downloading %s...\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(execPath), "gitplm-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write binary: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		os.Remove(oldPath)
		if err := os.Rename(execPath, oldPath); err != nil {
			return fmt.Errorf("failed to backup old binary: %w", err)
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			os.Rename(oldPath, execPath)
			return fmt.Errorf("failed to install new binary: %w", err)
		}
		os.Remove(oldPath)
	} else {
		if err := os.Rename(tmpPath, execPath); err != nil {
			return fmt.Errorf("failed to install new binary: %w", err)
		}
	}

	fmt.Println("Binary updated successfully")
	return nil
}

func getBinaryName(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	osName := goos
	if goos == "darwin" {
		osName = "macos"
	}

	archName := goarch
	if goarch == "amd64" {
		archName = "x86_64"
	} else if goarch == "386" {
		archName = "i386"
	}

	armVersion := ""
	if goarch == "arm" {
		armVersion = "7"
		if v := os.Getenv("GOARM"); v != "" {
			armVersion = v
		}
	}

	return fmt.Sprintf("gitplm-%s-%s-%s%s", version, osName, archName, armVersion)
}
