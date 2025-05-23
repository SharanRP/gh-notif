package distribution

import (
	"context"
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

// PackageTest represents a package distribution test
type PackageTest struct {
	Name        string
	Platform    string
	PackageType string
	InstallCmd  []string
	VerifyCmd   []string
	UpdateCmd   []string
	UninstallCmd []string
	Prerequisites []string
}

// TestPackageDistribution tests all package distribution methods
func TestPackageDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping package distribution tests in short mode")
	}

	tests := getPackageTests()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Platform != "all" && test.Platform != runtime.GOOS {
				t.Skipf("Skipping %s test on %s", test.PackageType, runtime.GOOS)
			}

			// Check prerequisites
			if !checkPrerequisites(t, test.Prerequisites) {
				t.Skipf("Prerequisites not met for %s", test.Name)
			}

			// Run the package test
			runPackageTest(t, test)
		})
	}
}

func getPackageTests() []PackageTest {
	tests := []PackageTest{
		{
			Name:        "Docker",
			Platform:    "all",
			PackageType: "docker",
			InstallCmd:  []string{"docker", "pull", "ghcr.io/sharanrp/gh-notif:latest"},
			VerifyCmd:   []string{"docker", "run", "--rm", "ghcr.io/sharanrp/gh-notif:latest", "--version"},
			UpdateCmd:   []string{"docker", "pull", "ghcr.io/sharanrp/gh-notif:latest"},
			UninstallCmd: []string{"docker", "rmi", "ghcr.io/sharanrp/gh-notif:latest"},
			Prerequisites: []string{"docker"},
		},
		{
			Name:        "Go Install",
			Platform:    "all",
			PackageType: "go",
			InstallCmd:  []string{"go", "install", "github.com/SharanRP/gh-notif@latest"},
			VerifyCmd:   []string{"gh-notif", "--version"},
			UpdateCmd:   []string{"go", "install", "github.com/SharanRP/gh-notif@latest"},
			UninstallCmd: []string{"rm", "-f", "$(go env GOPATH)/bin/gh-notif"},
			Prerequisites: []string{"go"},
		},
	}

	// Add platform-specific tests
	switch runtime.GOOS {
	case "darwin":
		tests = append(tests, PackageTest{
			Name:        "Homebrew",
			Platform:    "darwin",
			PackageType: "brew",
			InstallCmd:  []string{"brew", "install", "SharanRP/tap/gh-notif"},
			VerifyCmd:   []string{"gh-notif", "--version"},
			UpdateCmd:   []string{"brew", "upgrade", "gh-notif"},
			UninstallCmd: []string{"brew", "uninstall", "gh-notif"},
			Prerequisites: []string{"brew"},
		})
	case "windows":
		tests = append(tests, PackageTest{
			Name:        "Scoop",
			Platform:    "windows",
			PackageType: "scoop",
			InstallCmd:  []string{"scoop", "install", "gh-notif"},
			VerifyCmd:   []string{"gh-notif", "--version"},
			UpdateCmd:   []string{"scoop", "update", "gh-notif"},
			UninstallCmd: []string{"scoop", "uninstall", "gh-notif"},
			Prerequisites: []string{"scoop"},
		})
	case "linux":
		tests = append(tests, []PackageTest{
			{
				Name:        "Snap",
				Platform:    "linux",
				PackageType: "snap",
				InstallCmd:  []string{"sudo", "snap", "install", "gh-notif"},
				VerifyCmd:   []string{"gh-notif", "--version"},
				UpdateCmd:   []string{"sudo", "snap", "refresh", "gh-notif"},
				UninstallCmd: []string{"sudo", "snap", "remove", "gh-notif"},
				Prerequisites: []string{"snap"},
			},
			{
				Name:        "DEB Package",
				Platform:    "linux",
				PackageType: "deb",
				InstallCmd:  []string{"sudo", "dpkg", "-i", "gh-notif_amd64.deb"},
				VerifyCmd:   []string{"gh-notif", "--version"},
				UpdateCmd:   []string{"sudo", "dpkg", "-i", "gh-notif_amd64.deb"},
				UninstallCmd: []string{"sudo", "dpkg", "-r", "gh-notif"},
				Prerequisites: []string{"dpkg"},
			},
		}...)
	}

	return tests
}

func checkPrerequisites(t *testing.T, prerequisites []string) bool {
	for _, prereq := range prerequisites {
		if _, err := exec.LookPath(prereq); err != nil {
			t.Logf("Prerequisite %s not found", prereq)
			return false
		}
	}
	return true
}

func runPackageTest(t *testing.T, test PackageTest) {
	// Setup test environment
	tmpDir := setupPackageTestEnvironment(t)
	defer cleanupPackageTestEnvironment(t, tmpDir)

	// Download package if needed
	if test.PackageType == "deb" {
		downloadDebPackage(t, tmpDir)
		// Update install command with actual path
		test.InstallCmd[len(test.InstallCmd)-1] = filepath.Join(tmpDir, "gh-notif_amd64.deb")
	}

	// Test installation
	t.Run("Install", func(t *testing.T) {
		testPackageInstall(t, test)
	})

	// Test verification
	t.Run("Verify", func(t *testing.T) {
		testPackageVerify(t, test)
	})

	// Test update mechanism
	t.Run("Update", func(t *testing.T) {
		testPackageUpdate(t, test)
	})

	// Test uninstallation
	t.Run("Uninstall", func(t *testing.T) {
		testPackageUninstall(t, test)
	})

	// Test cleanup verification
	t.Run("Cleanup", func(t *testing.T) {
		testPackageCleanup(t, test)
	})
}

func setupPackageTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-package-test-*")
	require.NoError(t, err)
	return tmpDir
}

func cleanupPackageTestEnvironment(t *testing.T, tmpDir string) {
	os.RemoveAll(tmpDir)
}

func downloadDebPackage(t *testing.T, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	downloadURL := "https://github.com/SharanRP/gh-notif/releases/latest/download/gh-notif_amd64.deb"
	outputPath := filepath.Join(tmpDir, "gh-notif_amd64.deb")

	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", outputPath, downloadURL)
	err := cmd.Run()
	require.NoError(t, err, "Failed to download DEB package")
}

func testPackageInstall(t *testing.T, test PackageTest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Special handling for package managers that need setup
	if test.PackageType == "brew" {
		// Add tap first
		cmd := exec.CommandContext(ctx, "brew", "tap", "SharanRP/tap")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Tap add output: %s", output)
			// Continue even if tap already exists
		}
	} else if test.PackageType == "scoop" {
		// Add bucket first
		cmd := exec.CommandContext(ctx, "scoop", "bucket", "add", "SharanRP", "https://github.com/SharanRP/scoop-bucket")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Bucket add output: %s", output)
			// Continue even if bucket already exists
		}
	}

	// Run installation command
	cmd := exec.CommandContext(ctx, test.InstallCmd[0], test.InstallCmd[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Install command output: %s", output)
		t.Logf("Install command error: %v", err)
	}

	require.NoError(t, err, "Package installation should succeed")
}

func testPackageVerify(t *testing.T, test PackageTest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, test.VerifyCmd[0], test.VerifyCmd[1:]...)
	output, err := cmd.CombinedOutput()

	require.NoError(t, err, "Package verification should succeed: %s", output)
	assert.Contains(t, string(output), "gh-notif", "Version output should contain gh-notif")
}

func testPackageUpdate(t *testing.T, test PackageTest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, test.UpdateCmd[0], test.UpdateCmd[1:]...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Update command output: %s", output)
		t.Logf("Update command error: %v", err)
	}

	// Update might fail if already at latest version, which is acceptable
	if err != nil && !strings.Contains(string(output), "already") && !strings.Contains(string(output), "latest") {
		t.Errorf("Package update failed: %v", err)
	}
}

func testPackageUninstall(t *testing.T, test PackageTest) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Handle special uninstall commands
	var cmd *exec.Cmd
	if test.PackageType == "go" {
		// For Go install, we need to find and remove the binary
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			// Get default GOPATH
			cmd = exec.CommandContext(ctx, "go", "env", "GOPATH")
			output, err := cmd.Output()
			require.NoError(t, err, "Should get GOPATH")
			gopath = strings.TrimSpace(string(output))
		}

		binaryPath := filepath.Join(gopath, "bin", "gh-notif")
		if runtime.GOOS == "windows" {
			binaryPath += ".exe"
		}

		cmd = exec.CommandContext(ctx, "rm", "-f", binaryPath)
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "del", "/f", binaryPath)
		}
	} else {
		cmd = exec.CommandContext(ctx, test.UninstallCmd[0], test.UninstallCmd[1:]...)
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Uninstall command output: %s", output)
		t.Logf("Uninstall command error: %v", err)
	}

	require.NoError(t, err, "Package uninstallation should succeed")
}

func testPackageCleanup(t *testing.T, test PackageTest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Verify the package is no longer available
	cmd := exec.CommandContext(ctx, test.VerifyCmd[0], test.VerifyCmd[1:]...)
	err := cmd.Run()

	// Should fail since package is uninstalled
	assert.Error(t, err, "Package should not be available after uninstall")
}

// TestUpdateMechanism tests the update mechanism functionality
func TestUpdateMechanism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping update mechanism tests in short mode")
	}

	tmpDir := setupPackageTestEnvironment(t)
	defer cleanupPackageTestEnvironment(t, tmpDir)

	// Build a test binary
	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("Version Check", func(t *testing.T) {
		testVersionCheck(t, binaryPath)
	})

	t.Run("Update Check", func(t *testing.T) {
		testUpdateCheck(t, binaryPath)
	})

	t.Run("Self Update", func(t *testing.T) {
		testSelfUpdate(t, binaryPath)
	})
}

func buildTestBinary(t *testing.T, tmpDir string) string {
	binaryName := "gh-notif-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tmpDir, binaryName)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, "../../")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build test binary: %s", output)

	return binaryPath
}

func testVersionCheck(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Version command should succeed: %s", output)

	assert.Contains(t, string(output), "gh-notif", "Version output should contain gh-notif")
	assert.Contains(t, string(output), "version", "Version output should contain version info")
}

func testUpdateCheck(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "version", "--check-update")
	output, err := cmd.CombinedOutput()

	// Command might succeed or fail depending on network and current version
	t.Logf("Update check output: %s", output)

	if err == nil {
		// If successful, should contain update information
		assert.True(t,
			strings.Contains(string(output), "latest") ||
			strings.Contains(string(output), "update") ||
			strings.Contains(string(output), "current"),
			"Update check should provide version information")
	}
}

func testSelfUpdate(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Test self-update in dry-run mode
	cmd := exec.CommandContext(ctx, binaryPath, "version", "--update", "--dry-run")
	output, err := cmd.CombinedOutput()

	t.Logf("Self-update output: %s", output)

	// Self-update might not be implemented yet, so we just log the result
	if err != nil {
		t.Logf("Self-update not available: %v", err)
	} else {
		assert.Contains(t, string(output), "update", "Self-update should provide update information")
	}
}

// TestDependencyRequirements tests installation dependencies
func TestDependencyRequirements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping dependency tests in short mode")
	}

	t.Run("System Dependencies", func(t *testing.T) {
		testSystemDependencies(t)
	})

	t.Run("Runtime Dependencies", func(t *testing.T) {
		testRuntimeDependencies(t)
	})
}

func testSystemDependencies(t *testing.T) {
	// Test that gh-notif doesn't require unexpected system dependencies
	dependencies := []string{
		"git", // Should be available for GitHub integration
	}

	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			t.Logf("Optional dependency %s not found: %v", dep, err)
		} else {
			t.Logf("Dependency %s found", dep)
		}
	}
}

func testRuntimeDependencies(t *testing.T) {
	tmpDir := setupPackageTestEnvironment(t)
	defer cleanupPackageTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	// Test that the binary runs without additional runtime dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--help")
	err := cmd.Run()
	require.NoError(t, err, "Binary should run without additional dependencies")
}
