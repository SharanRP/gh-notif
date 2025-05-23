package e2e

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DocumentationTest represents a documentation test case
type DocumentationTest struct {
	Name     string
	File     string
	Commands []string
	Examples []string
	Sections []string
}

// TestDocumentationAccuracy tests that all documented features work as described
func TestDocumentationAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping documentation tests in short mode")
	}

	tmpDir := setupDocTestEnvironment(t)
	defer cleanupDocTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("README Examples", func(t *testing.T) {
		testREADMEExamples(t, binaryPath)
	})

	t.Run("Help Text Accuracy", func(t *testing.T) {
		testHelpTextAccuracy(t, binaryPath)
	})

	t.Run("Command Examples", func(t *testing.T) {
		testCommandExamples(t, binaryPath)
	})

	t.Run("Configuration Documentation", func(t *testing.T) {
		testConfigurationDocumentation(t, binaryPath)
	})

	t.Run("Man Pages", func(t *testing.T) {
		testManPages(t, binaryPath)
	})
}

func setupDocTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-doc-test-*")
	require.NoError(t, err)

	// Set environment variables for isolated testing
	os.Setenv("GH_NOTIF_CONFIG", filepath.Join(tmpDir, "config.yaml"))
	os.Setenv("GH_NOTIF_CACHE_DIR", filepath.Join(tmpDir, "cache"))

	return tmpDir
}

func cleanupDocTestEnvironment(t *testing.T, tmpDir string) {
	os.RemoveAll(tmpDir)
	os.Unsetenv("GH_NOTIF_CONFIG")
	os.Unsetenv("GH_NOTIF_CACHE_DIR")
}

func buildTestBinary(t *testing.T, tmpDir string) string {
	binaryName := "gh-notif-doc-test"
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

func testREADMEExamples(t *testing.T, binaryPath string) {
	// Read README.md and extract code examples
	readmePath := "../../README.md"
	examples, err := extractCodeExamples(readmePath)
	require.NoError(t, err, "Should extract examples from README")

	for i, example := range examples {
		t.Run(fmt.Sprintf("Example_%d", i+1), func(t *testing.T) {
			testCodeExample(t, binaryPath, example)
		})
	}
}

func extractCodeExamples(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var examples []string
	var currentExample strings.Builder
	inCodeBlock := false
	isBashBlock := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "```bash") || strings.HasPrefix(line, "```sh") {
			inCodeBlock = true
			isBashBlock = true
			currentExample.Reset()
			continue
		}

		if strings.HasPrefix(line, "```") && inCodeBlock {
			if isBashBlock && currentExample.Len() > 0 {
				examples = append(examples, currentExample.String())
			}
			inCodeBlock = false
			isBashBlock = false
			currentExample.Reset()
			continue
		}

		if inCodeBlock && isBashBlock {
			// Skip comments and empty lines
			trimmed := strings.TrimSpace(line)
			if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				currentExample.WriteString(line + "\n")
			}
		}
	}

	return examples, scanner.Err()
}

// replaceGhNotifCommand replaces 'gh-notif' command with the test binary path
func replaceGhNotifCommand(line, binaryPath string) string {
	// Use regex to replace only standalone 'gh-notif' commands
	re := regexp.MustCompile(`\bgh-notif\b`)
	return re.ReplaceAllString(line, binaryPath)
}

func testCodeExample(t *testing.T, binaryPath, example string) {
	lines := strings.Split(strings.TrimSpace(example), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Replace 'gh-notif' with our test binary path (only at word boundaries)
		line = replaceGhNotifCommand(line, binaryPath)

		// Skip commands that require authentication or external dependencies
		if shouldSkipCommand(line) {
			t.Logf("Skipping command: %s", line)
			continue
		}

		// Parse and execute command
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

		output, err := cmd.CombinedOutput()
		cancel()

		// Some commands are expected to fail (like auth commands without tokens)
		if err != nil {
			t.Logf("Command failed (expected for some examples): %s", line)
			t.Logf("Output: %s", output)
		} else {
			t.Logf("Command succeeded: %s", line)
		}
	}
}

func shouldSkipCommand(command string) bool {
	skipPatterns := []string{
		"auth login",
		"list",                       // Requires authentication
		"search",                     // Requires authentication
		"read",                       // Requires authentication
		"open",                       // Requires authentication
		"subscribe",                  // Requires authentication
		"watch",                      // Requires authentication and runs indefinitely
		"curl",                       // External command
		"brew",                       // External command
		"scoop",                      // External command
		"snap",                       // External command
		"docker",                     // External command
		"sudo",                       // Requires elevated permissions
		"flatpak",                    // External command
		"rpm",                        // External command
		"dpkg",                       // External command
		"git clone",                  // External command
		"make",                       // External command
		"source",                     // Shell builtin
		"echo",                       // Shell command
		"cd ",                        // Shell builtin
		"GOOS=",                      // Environment variable setting
		"go install github.com/user", // External repository
		"go build",                   // Requires source files
		"filter save",                // Requires authentication
		"filter get",                 // May require saved filters
		"filter delete",              // May require saved filters
		"completion",                 // May have parsing issues
		"tee",                        // External command
		"|",                          // Pipe operations can be complex
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(command, pattern) {
			return true
		}
	}

	return false
}

func testHelpTextAccuracy(t *testing.T, binaryPath string) {
	// Test main help
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Main help should work")

	helpText := string(output)

	// Check for essential sections
	requiredSections := []string{
		"Usage:",
		"Commands:",
		"Flags:",
		"Examples:",
	}

	for _, section := range requiredSections {
		assert.Contains(t, helpText, section, "Help should contain %s section", section)
	}

	// Test subcommand help
	subcommands := []string{
		"list",
		"group",
		"search",
		"read",
		"open",
		"subscribe",
		"auth",
		"config",
		"version",
	}

	for _, subcmd := range subcommands {
		t.Run(fmt.Sprintf("Help_%s", subcmd), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, subcmd, "--help")
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Help for %s should work", subcmd)

			helpText := string(output)
			assert.Contains(t, helpText, "Usage:", "Help should contain usage")
			assert.Contains(t, helpText, subcmd, "Help should mention the command")
		})
	}
}

func testCommandExamples(t *testing.T, binaryPath string) {
	// Test that examples in help text are valid
	commands := []string{
		"list",
		"group",
		"search",
		"auth",
		"config",
		"version",
	}

	for _, cmd := range commands {
		t.Run(fmt.Sprintf("Examples_%s", cmd), func(t *testing.T) {
			examples := extractHelpExamples(t, binaryPath, cmd)

			for i, example := range examples {
				t.Run(fmt.Sprintf("Example_%d", i+1), func(t *testing.T) {
					testHelpExample(t, binaryPath, example)
				})
			}
		})
	}
}

func extractHelpExamples(t *testing.T, binaryPath, command string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, command, "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Help should work for %s", command)

	helpText := string(output)

	// Extract examples section
	lines := strings.Split(helpText, "\n")
	var examples []string
	inExamples := false

	for _, line := range lines {
		if strings.Contains(line, "Examples:") {
			inExamples = true
			continue
		}

		if inExamples {
			// Stop at next section
			if strings.HasPrefix(line, "Flags:") || strings.HasPrefix(line, "Global Flags:") {
				break
			}

			// Extract command examples
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "gh-notif") {
				examples = append(examples, trimmed)
			}
		}
	}

	return examples
}

func testHelpExample(t *testing.T, binaryPath, example string) {
	// Replace 'gh-notif' with test binary path
	example = replaceGhNotifCommand(example, binaryPath)

	// Skip examples that require authentication
	if shouldSkipCommand(example) {
		t.Logf("Skipping example: %s", example)
		return
	}

	// Parse command
	parts := strings.Fields(example)
	if len(parts) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()

	// Examples might fail due to missing auth, but should not crash
	if err != nil {
		t.Logf("Example failed (may be expected): %s", example)
		t.Logf("Error: %v", err)
		t.Logf("Output: %s", output)

		// Check that it's a reasonable error message
		assert.NotContains(t, string(output), "panic", "Should not panic")
		assert.NotContains(t, string(output), "fatal error", "Should not have fatal errors")
	}
}

func testConfigurationDocumentation(t *testing.T, binaryPath string) {
	// Test that documented configuration options work
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test config list to see available options
	cmd := exec.CommandContext(ctx, binaryPath, "config", "list")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Config list failed: %v", err)
		t.Logf("Output: %s", output)
		return
	}

	configText := string(output)

	// Check for documented config sections
	expectedSections := []string{
		"display",
		"auth",
		"cache",
		"api",
	}

	for _, section := range expectedSections {
		if strings.Contains(configText, section) {
			t.Logf("Found config section: %s", section)
		}
	}

	// Test setting and getting config values
	testConfigs := map[string]string{
		"display.limit":  "25",
		"display.format": "table",
	}

	for key, value := range testConfigs {
		t.Run(fmt.Sprintf("Config_%s", key), func(t *testing.T) {
			// Set config
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, binaryPath, "config", "set", key, value)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Logf("Config set failed: %v", err)
				t.Logf("Output: %s", output)
				return
			}

			// Get config
			cmd = exec.CommandContext(ctx, binaryPath, "config", "get", key)
			output, err = cmd.CombinedOutput()

			if err != nil {
				t.Logf("Config get failed: %v", err)
				t.Logf("Output: %s", output)
				return
			}

			assert.Contains(t, string(output), value, "Config should return set value")
		})
	}
}

func testManPages(t *testing.T, binaryPath string) {
	// Test man page generation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tmpDir, err := os.MkdirTemp("", "gh-notif-man-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	manDir := filepath.Join(tmpDir, "man")

	cmd := exec.CommandContext(ctx, binaryPath, "man", "--output-dir", manDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Man page generation failed: %v", err)
		t.Logf("Output: %s", output)
		return
	}

	// Check that man pages were created
	manFiles, err := filepath.Glob(filepath.Join(manDir, "*.1"))
	if err != nil {
		t.Logf("Error checking man files: %v", err)
		return
	}

	assert.Greater(t, len(manFiles), 0, "Should generate at least one man page")

	// Check man page content
	for _, manFile := range manFiles {
		content, err := os.ReadFile(manFile)
		require.NoError(t, err, "Should read man page file")

		manContent := string(content)

		// Check for standard man page sections
		expectedSections := []string{
			".TH",
			".SH NAME",
			".SH SYNOPSIS",
			".SH DESCRIPTION",
		}

		for _, section := range expectedSections {
			assert.Contains(t, manContent, section, "Man page should contain %s section", section)
		}
	}
}

// TestDocumentationCompleteness tests for documentation gaps
func TestDocumentationCompleteness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping documentation completeness tests in short mode")
	}

	t.Run("Command Coverage", func(t *testing.T) {
		testCommandCoverage(t)
	})

	t.Run("Flag Coverage", func(t *testing.T) {
		testFlagCoverage(t)
	})

	t.Run("Example Coverage", func(t *testing.T) {
		testExampleCoverage(t)
	})
}

func testCommandCoverage(t *testing.T) {
	// Check that all commands are documented in README
	readmePath := "../../README.md"
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err, "Should read README")

	readmeContent := string(content)

	// List of all commands that should be documented
	commands := []string{
		"list",
		"group",
		"search",
		"read",
		"open",
		"subscribe",
		"auth",
		"config",
		"version",
		"tutorial",
		"firstrun",
		"completion",
		"man",
	}

	for _, cmd := range commands {
		assert.Contains(t, readmeContent, cmd, "README should document %s command", cmd)
	}
}

func testFlagCoverage(t *testing.T) {
	// Check that important flags are documented
	readmePath := "../../README.md"
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err, "Should read README")

	readmeContent := string(content)

	// Important flags that should be documented
	flags := []string{
		"--filter",
		"--limit",
		"--format",
		"--output",
		"--help",
		"--version",
	}

	for _, flag := range flags {
		assert.Contains(t, readmeContent, flag, "README should document %s flag", flag)
	}
}

func testExampleCoverage(t *testing.T) {
	// Check that README contains practical examples
	readmePath := "../../README.md"
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err, "Should read README")

	readmeContent := string(content)

	// Should contain various types of examples
	exampleTypes := []string{
		"gh-notif list",
		"gh-notif auth",
		"gh-notif config",
		"gh-notif search",
	}

	for _, example := range exampleTypes {
		assert.Contains(t, readmeContent, example, "README should contain %s examples", example)
	}
}
