package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// Build information set by ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	Date      = "unknown"
	BuiltBy   = "unknown"
	GoVersion = runtime.Version()
	Platform  = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info contains version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	BuiltBy   string `json:"built_by"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetInfo returns version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		BuiltBy:   BuiltBy,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("gh-notif version %s (%s) built on %s by %s\nGo version: %s\nPlatform: %s",
		i.Version, i.Commit, i.Date, i.BuiltBy, i.GoVersion, i.Platform)
}

// JSON returns version information as JSON
func (i Info) JSON() (string, error) {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateChecker checks for updates
type UpdateChecker struct {
	// CurrentVersion is the current version
	CurrentVersion string

	// Repository is the GitHub repository (e.g., "user/gh-notif")
	Repository string

	// HTTPClient is the HTTP client to use
	HTTPClient *http.Client

	// CheckInterval is how often to check for updates
	CheckInterval time.Duration

	// LastCheck is when we last checked for updates
	LastCheck time.Time
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker(repository string) *UpdateChecker {
	return &UpdateChecker{
		CurrentVersion: Version,
		Repository:     repository,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		CheckInterval:  24 * time.Hour,
	}
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
}

// CheckForUpdate checks for a newer version
func (uc *UpdateChecker) CheckForUpdate(ctx context.Context) (*UpdateInfo, error) {
	// Check if we should check for updates
	if time.Since(uc.LastCheck) < uc.CheckInterval {
		return nil, nil
	}

	// Update last check time
	uc.LastCheck = time.Now()

	// Get the latest release from GitHub
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", uc.Repository)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := uc.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Skip draft and prerelease versions
	if release.Draft || release.Prerelease {
		return nil, nil
	}

	// Parse versions
	currentVer, err := semver.NewVersion(strings.TrimPrefix(uc.CurrentVersion, "v"))
	if err != nil {
		return nil, fmt.Errorf("error parsing current version: %w", err)
	}

	latestVer, err := semver.NewVersion(strings.TrimPrefix(release.TagName, "v"))
	if err != nil {
		return nil, fmt.Errorf("error parsing latest version: %w", err)
	}

	// Check if there's a newer version
	if latestVer.GreaterThan(currentVer) {
		return &UpdateInfo{
			CurrentVersion: uc.CurrentVersion,
			LatestVersion:  release.TagName,
			ReleaseURL:     release.HTMLURL,
			ReleaseNotes:   release.Body,
			PublishedAt:    release.PublishedAt,
		}, nil
	}

	return nil, nil
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	CurrentVersion string    `json:"current_version"`
	LatestVersion  string    `json:"latest_version"`
	ReleaseURL     string    `json:"release_url"`
	ReleaseNotes   string    `json:"release_notes"`
	PublishedAt    time.Time `json:"published_at"`
}

// String returns a formatted update message
func (ui *UpdateInfo) String() string {
	return fmt.Sprintf(`A new version of gh-notif is available!

Current version: %s
Latest version:  %s
Released:        %s

Release notes:
%s

Download: %s

To update, run:
  gh-notif update
`,
		ui.CurrentVersion,
		ui.LatestVersion,
		ui.PublishedAt.Format("2006-01-02"),
		ui.ReleaseNotes,
		ui.ReleaseURL,
	)
}

// SelfUpdater handles self-updating the binary
type SelfUpdater struct {
	// Repository is the GitHub repository
	Repository string

	// HTTPClient is the HTTP client to use
	HTTPClient *http.Client

	// Platform is the target platform
	Platform string

	// Architecture is the target architecture
	Architecture string
}

// NewSelfUpdater creates a new self-updater
func NewSelfUpdater(repository string) *SelfUpdater {
	return &SelfUpdater{
		Repository:   repository,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
		Platform:     runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

// Update updates the binary to the latest version
func (su *SelfUpdater) Update(ctx context.Context, version string) error {
	// Construct the download URL
	var ext string
	if su.Platform == "windows" {
		ext = ".zip"
	} else {
		ext = ".tar.gz"
	}

	archName := su.Architecture
	if archName == "amd64" {
		archName = "x86_64"
	}

	filename := fmt.Sprintf("gh-notif_%s_%s_%s%s", version, su.Platform, archName, ext)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", su.Repository, version, filename)

	// Download the new binary
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("error creating download request: %w", err)
	}

	resp, err := su.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error downloading update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// TODO: Implement binary replacement logic
	// This would involve:
	// 1. Extracting the archive
	// 2. Verifying the binary
	// 3. Replacing the current binary
	// 4. Restarting the application

	return fmt.Errorf("self-update not yet implemented")
}

// IsUpdateAvailable checks if an update is available without detailed info
func IsUpdateAvailable(ctx context.Context, repository string) (bool, error) {
	checker := NewUpdateChecker(repository)
	updateInfo, err := checker.CheckForUpdate(ctx)
	if err != nil {
		return false, err
	}
	return updateInfo != nil, nil
}
