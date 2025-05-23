package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstallationMethods tests all supported installation methods
func TestInstallationMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping installation tests in short mode")
	}

	tests := []struct {
		name     string
		method   string
		platform string
		setup    func(t *testing.T) string
		install  func(t *testing.T, workdir string) error
		verify   func(t *testing.T, workdir string) error
		cleanup  func(t *testing.T, workdir string) error
	}{
		{
			name:     "Direct Binary Download",
			method:   "binary",
			platform: "all",
			setup:    setupTempDir,
			install:  installFromBinary,
			verify:   verifyBinaryInstallation,
			cleanup:  cleanupBinary,
		},
		{
			name:     "Go Install",
			method:   "go",
			platform: "all",
			setup:    setupGoEnvironment,
			install:  installFromGo,
			verify:   verifyGoInstallation,
			cleanup:  cleanupGoInstallation,
		},
		{
			name:     "Docker",
			method:   "docker",
			platform: "all",
			setup:    setupDockerEnvironment,
			install:  installFromDocker,
			verify:   verifyDockerInstallation,
			cleanup:  cleanupDockerInstallation,
		},
	}

	// Add platform-specific tests
	switch runtime.GOOS {
	case "darwin":
		tests = append(tests, struct {
			name     string
			method   string
			platform string
			setup    func(t *testing.T) string
			install  func(t *testing.T, workdir string) error
			verify   func(t *testing.T, workdir string) error
			cleanup  func(t *testing.T, workdir string) error
		}{
			name:     "Homebrew",
			method:   "brew",
			platform: "darwin",
			setup:    setupBrewEnvironment,
			install:  installFromBrew,
			verify:   verifyBrewInstallation,
			cleanup:  cleanupBrewInstallation,
		})
	case "windows":
		tests = append(tests, struct {
			name     string
			method   string
			platform string
			setup    func(t *testing.T) string
			install  func(t *testing.T, workdir string) error
			verify   func(t *testing.T, workdir string) error
			cleanup  func(t *testing.T, workdir string) error
		}{
			name:     "Scoop",
			method:   "scoop",
			platform: "windows",
			setup:    setupScoopEnvironment,
			install:  installFromScoop,
			verify:   verifyScoopInstallation,
			cleanup:  cleanupScoopInstallation,
		})
	case "linux":
		tests = append(tests, []struct {
			name     string
			method   string
			platform string
			setup    func(t *testing.T) string
			install  func(t *testing.T, workdir string) error
			verify   func(t *testing.T, workdir string) error
			cleanup  func(t *testing.T, workdir string) error
		}{
			{
				name:     "Snap",
				method:   "snap",
				platform: "linux",
				setup:    setupSnapEnvironment,
				install:  installFromSnap,
				verify:   verifySnapInstallation,
				cleanup:  cleanupSnapInstallation,
			},
			{
				name:     "DEB Package",
				method:   "deb",
				platform: "linux",
				setup:    setupDebEnvironment,
				install:  installFromDeb,
				verify:   verifyDebInstallation,
				cleanup:  cleanupDebInstallation,
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.platform != "all" && tt.platform != runtime.GOOS {
				t.Skipf("Skipping %s test on %s", tt.method, runtime.GOOS)
			}

			// Setup
			workdir := tt.setup(t)
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup(t, workdir)
				}
			}()

			// Install
			err := tt.install(t, workdir)
			require.NoError(t, err, "Installation should succeed")

			// Verify
			err = tt.verify(t, workdir)
			assert.NoError(t, err, "Verification should succeed")
		})
	}
}

// Setup functions
func setupTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-test-*")
	require.NoError(t, err)
	return tmpDir
}

func setupGoEnvironment(t *testing.T) string {
	// Check if Go is available
	_, err := exec.LookPath("go")
	require.NoError(t, err, "Go must be installed for this test")
	
	return setupTempDir(t)
}

func setupDockerEnvironment(t *testing.T) string {
	// Check if Docker is available
	_, err := exec.LookPath("docker")
	require.NoError(t, err, "Docker must be installed for this test")
	
	return setupTempDir(t)
}

func setupBrewEnvironment(t *testing.T) string {
	// Check if Homebrew is available
	_, err := exec.LookPath("brew")
	require.NoError(t, err, "Homebrew must be installed for this test")
	
	return setupTempDir(t)
}

func setupScoopEnvironment(t *testing.T) string {
	// Check if Scoop is available
	_, err := exec.LookPath("scoop")
	require.NoError(t, err, "Scoop must be installed for this test")
	
	return setupTempDir(t)
}

func setupSnapEnvironment(t *testing.T) string {
	// Check if Snap is available
	_, err := exec.LookPath("snap")
	require.NoError(t, err, "Snap must be installed for this test")
	
	return setupTempDir(t)
}

func setupDebEnvironment(t *testing.T) string {
	// Check if dpkg is available
	_, err := exec.LookPath("dpkg")
	require.NoError(t, err, "dpkg must be installed for this test")
	
	return setupTempDir(t)
}

// Installation functions
func installFromBinary(t *testing.T, workdir string) error {
	// Download the latest release binary
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Determine platform and architecture
	platform := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x86_64"
	}

	var ext string
	if platform == "windows" {
		ext = ".zip"
	} else {
		ext = ".tar.gz"
	}

	filename := fmt.Sprintf("gh-notif_%s_%s%s", strings.Title(platform), arch, ext)
	downloadURL := fmt.Sprintf("https://github.com/user/gh-notif/releases/latest/download/%s", filename)

	// Download
	downloadPath := filepath.Join(workdir, filename)
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", downloadPath, downloadURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Extract
	if platform == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-Command", 
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s'", downloadPath, workdir))
	} else {
		cmd = exec.CommandContext(ctx, "tar", "-xzf", downloadPath, "-C", workdir)
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	return nil
}

func installFromGo(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Set GOPATH to workdir
	env := append(os.Environ(), fmt.Sprintf("GOPATH=%s", workdir))
	
	cmd := exec.CommandContext(ctx, "go", "install", "github.com/user/gh-notif@latest")
	cmd.Env = env
	
	return cmd.Run()
}

func installFromDocker(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Pull the Docker image
	cmd := exec.CommandContext(ctx, "docker", "pull", "ghcr.io/user/gh-notif:latest")
	return cmd.Run()
}

func installFromBrew(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Add tap and install
	cmd := exec.CommandContext(ctx, "brew", "tap", "user/tap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	cmd = exec.CommandContext(ctx, "brew", "install", "gh-notif")
	return cmd.Run()
}

func installFromScoop(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Add bucket and install
	cmd := exec.CommandContext(ctx, "scoop", "bucket", "add", "user", "https://github.com/user/scoop-bucket")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add bucket: %w", err)
	}

	cmd = exec.CommandContext(ctx, "scoop", "install", "gh-notif")
	return cmd.Run()
}

func installFromSnap(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", "snap", "install", "gh-notif")
	return cmd.Run()
}

func installFromDeb(t *testing.T, workdir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Download DEB package
	downloadURL := "https://github.com/user/gh-notif/releases/latest/download/gh-notif_amd64.deb"
	debPath := filepath.Join(workdir, "gh-notif.deb")
	
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", debPath, downloadURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download DEB package: %w", err)
	}

	cmd = exec.CommandContext(ctx, "sudo", "dpkg", "-i", debPath)
	return cmd.Run()
}

// Verification functions
func verifyBinaryInstallation(t *testing.T, workdir string) error {
	// Find the binary
	var binaryPath string
	err := filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(info.Name(), "gh-notif") && !strings.Contains(info.Name(), ".") {
			binaryPath = path
			return filepath.SkipDir
		}
		return nil
	})
	
	if err != nil || binaryPath == "" {
		return fmt.Errorf("binary not found in %s", workdir)
	}

	// Test the binary
	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run binary: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifyGoInstallation(t *testing.T, workdir string) error {
	// Check if binary is in GOPATH/bin
	binaryPath := filepath.Join(workdir, "bin", "gh-notif")
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run binary: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifyDockerInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("docker", "run", "--rm", "ghcr.io/user/gh-notif:latest", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run Docker container: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifyBrewInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("gh-notif", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run gh-notif: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifyScoopInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("gh-notif", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run gh-notif: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifySnapInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("gh-notif", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run gh-notif: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

func verifyDebInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("gh-notif", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run gh-notif: %w", err)
	}

	if !strings.Contains(string(output), "gh-notif") {
		return fmt.Errorf("unexpected version output: %s", output)
	}

	return nil
}

// Cleanup functions
func cleanupBinary(t *testing.T, workdir string) error {
	return os.RemoveAll(workdir)
}

func cleanupGoInstallation(t *testing.T, workdir string) error {
	return os.RemoveAll(workdir)
}

func cleanupDockerInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("docker", "rmi", "ghcr.io/user/gh-notif:latest")
	cmd.Run() // Ignore errors
	return nil
}

func cleanupBrewInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("brew", "uninstall", "gh-notif")
	cmd.Run() // Ignore errors
	cmd = exec.Command("brew", "untap", "user/tap")
	cmd.Run() // Ignore errors
	return nil
}

func cleanupScoopInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("scoop", "uninstall", "gh-notif")
	cmd.Run() // Ignore errors
	cmd = exec.Command("scoop", "bucket", "rm", "user")
	cmd.Run() // Ignore errors
	return nil
}

func cleanupSnapInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("sudo", "snap", "remove", "gh-notif")
	cmd.Run() // Ignore errors
	return nil
}

func cleanupDebInstallation(t *testing.T, workdir string) error {
	cmd := exec.Command("sudo", "dpkg", "-r", "gh-notif")
	cmd.Run() // Ignore errors
	return os.RemoveAll(workdir)
}
