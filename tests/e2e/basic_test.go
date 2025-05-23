package e2e

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

// TestBasicFunctionality tests basic CLI functionality without authentication
func TestBasicFunctionality(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping basic functionality tests in short mode")
	}

	tmpDir := setupBasicTestEnvironment(t)
	defer cleanupBasicTestEnvironment(t, tmpDir)

	binaryPath := buildBasicTestBinary(t, tmpDir)

	t.Run("Version Command", func(t *testing.T) {
		testVersionCommand(t, binaryPath)
	})

	t.Run("Help Commands", func(t *testing.T) {
		testHelpCommands(t, binaryPath)
	})

	t.Run("Config Commands", func(t *testing.T) {
		testConfigCommands(t, binaryPath)
	})

	t.Run("Auth Commands", func(t *testing.T) {
		testAuthCommands(t, binaryPath)
	})

	t.Run("Completion Commands", func(t *testing.T) {
		testCompletionCommands(t, binaryPath)
	})
}

func setupBasicTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-basic-test-*")
	require.NoError(t, err)

	// Set environment variables for isolated testing
	os.Setenv("GH_NOTIF_CONFIG", filepath.Join(tmpDir, "config.yaml"))
	os.Setenv("GH_NOTIF_CACHE_DIR", filepath.Join(tmpDir, "cache"))

	return tmpDir
}

func cleanupBasicTestEnvironment(t *testing.T, tmpDir string) {
	os.RemoveAll(tmpDir)
	os.Unsetenv("GH_NOTIF_CONFIG")
	os.Unsetenv("GH_NOTIF_CACHE_DIR")
}

func buildBasicTestBinary(t *testing.T, tmpDir string) string {
	binaryName := "gh-notif-basic-test"
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

func testVersionCommand(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Version command should succeed: %s", output)

	outputStr := string(output)
	assert.Contains(t, outputStr, "gh-notif", "Version output should contain gh-notif")
	assert.Contains(t, outputStr, "version", "Version output should contain version info")
}

func testHelpCommands(t *testing.T, binaryPath string) {
	commands := []string{
		"--help",
		"help",
		"auth --help",
		"config --help",
		"version --help",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			parts := strings.Fields(cmd)
			execCmd := exec.CommandContext(ctx, binaryPath)
			execCmd.Args = append(execCmd.Args, parts...)

			output, err := execCmd.CombinedOutput()
			require.NoError(t, err, "Help command should succeed: %s", output)

			outputStr := string(output)
			assert.Contains(t, outputStr, "Usage:", "Help should contain usage information")
		})
	}
}

func testConfigCommands(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test config list
	cmd := exec.CommandContext(ctx, binaryPath, "config", "list")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Config list should succeed: %s", output)

	// Test config get
	cmd = exec.CommandContext(ctx, binaryPath, "config", "get", "display.theme")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Config get should succeed: %s", output)

	// Test config set
	cmd = exec.CommandContext(ctx, binaryPath, "config", "set", "display.theme", "light")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Config set should succeed: %s", output)

	// Verify the change
	cmd = exec.CommandContext(ctx, binaryPath, "config", "get", "display.theme")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Config get after set should succeed: %s", output)
	assert.Contains(t, string(output), "light", "Config should return the set value")

	// Test invalid config
	cmd = exec.CommandContext(ctx, binaryPath, "config", "set", "invalid.key", "value")
	output, err = cmd.CombinedOutput()
	assert.Error(t, err, "Invalid config should fail")
	assert.Contains(t, string(output), "invalid", "Error message should mention invalid")
}

func testAuthCommands(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test auth status (should be unauthenticated)
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "status")
	output, err := cmd.CombinedOutput()
	assert.Error(t, err, "Should be unauthenticated initially")
	outputStr := strings.ToLower(string(output))
	assert.Contains(t, outputStr, "not authenticated", "Should indicate not authenticated")

	// Test auth logout (should succeed even if not logged in)
	cmd = exec.CommandContext(ctx, binaryPath, "auth", "logout")
	output, err = cmd.CombinedOutput()
	// This might succeed or fail depending on implementation, both are acceptable
	t.Logf("Auth logout output: %s", output)
}

func testCompletionCommands(t *testing.T, binaryPath string) {
	shells := []string{"bash", "zsh", "fish", "powershell"}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, "completion", shell)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Completion for %s should succeed: %s", shell, output)

			outputStr := string(output)
			assert.NotEmpty(t, outputStr, "Completion output should not be empty")

			// Basic validation that it looks like shell completion
			switch shell {
			case "bash":
				assert.Contains(t, outputStr, "bash", "Bash completion should contain bash-specific content")
			case "zsh":
				assert.Contains(t, outputStr, "zsh", "Zsh completion should contain zsh-specific content")
			case "fish":
				assert.Contains(t, outputStr, "complete", "Fish completion should contain complete commands")
			case "powershell":
				assert.Contains(t, outputStr, "Register-ArgumentCompleter", "PowerShell completion should contain Register-ArgumentCompleter")
			}
		})
	}
}

// TestErrorHandling tests error handling for invalid commands
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling tests in short mode")
	}

	tmpDir := setupBasicTestEnvironment(t)
	defer cleanupBasicTestEnvironment(t, tmpDir)

	binaryPath := buildBasicTestBinary(t, tmpDir)

	t.Run("Invalid Command", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "invalidcommand")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "Invalid command should fail")

		outputStr := string(output)
		assert.Contains(t, outputStr, "unknown command", "Should indicate unknown command")
	})

	t.Run("Invalid Flag", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "--invalid-flag")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "Invalid flag should fail")

		outputStr := string(output)
		assert.Contains(t, outputStr, "unknown flag", "Should indicate unknown flag")
	})

	t.Run("Missing Required Args", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, binaryPath, "config", "set")
		output, err := cmd.CombinedOutput()
		assert.Error(t, err, "Missing required args should fail")

		outputStr := string(output)
		// Should contain usage information or error about missing arguments
		assert.True(t,
			strings.Contains(outputStr, "Usage:") ||
			strings.Contains(outputStr, "required") ||
			strings.Contains(outputStr, "argument"),
			"Should indicate missing arguments or show usage")
	})
}

// TestCLIIntegration tests CLI integration scenarios
func TestCLIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration tests in short mode")
	}

	tmpDir := setupBasicTestEnvironment(t)
	defer cleanupBasicTestEnvironment(t, tmpDir)

	binaryPath := buildBasicTestBinary(t, tmpDir)

	t.Run("Config Workflow", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Set multiple config values
		configs := map[string]string{
			"display.theme":    "dark",
			"api.timeout":      "60",
			"advanced.debug":   "true",
		}

		for key, value := range configs {
			cmd := exec.CommandContext(ctx, binaryPath, "config", "set", key, value)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Config set %s should succeed: %s", key, output)
		}

		// Verify all values
		for key, expectedValue := range configs {
			cmd := exec.CommandContext(ctx, binaryPath, "config", "get", key)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Config get %s should succeed: %s", key, output)
			assert.Contains(t, string(output), expectedValue, "Config should return the set value for %s", key)
		}

		// List all config
		cmd := exec.CommandContext(ctx, binaryPath, "config", "list")
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "Config list should succeed: %s", output)

		outputStr := string(output)
		for key, value := range configs {
			assert.Contains(t, outputStr, value, "Config list should contain %s: %s", key, value)
		}
	})

	t.Run("Help System", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Test that help is available for all main commands
		commands := []string{"auth", "config", "version", "completion"}

		for _, cmd := range commands {
			execCmd := exec.CommandContext(ctx, binaryPath, cmd, "--help")
			output, err := execCmd.CombinedOutput()
			require.NoError(t, err, "Help for %s should succeed: %s", cmd, output)

			outputStr := string(output)
			assert.Contains(t, outputStr, "Usage:", "Help should contain usage for %s", cmd)
			assert.Contains(t, outputStr, cmd, "Help should mention the command %s", cmd)
		}
	})
}
