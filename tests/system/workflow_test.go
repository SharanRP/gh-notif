package system

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompleteWorkflow tests the complete user workflow
func TestCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping workflow tests in short mode")
	}

	// Skip if no GitHub token is available
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN environment variable required for workflow tests")
	}

	// Setup test environment
	tmpDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir)

	// Build the binary for testing
	binaryPath := buildTestBinary(t, tmpDir)

	// Test workflow steps
	t.Run("Authentication", func(t *testing.T) {
		testAuthentication(t, binaryPath, tmpDir, token)
	})

	t.Run("Configuration", func(t *testing.T) {
		testConfiguration(t, binaryPath, tmpDir)
	})

	t.Run("Notification Listing", func(t *testing.T) {
		testNotificationListing(t, binaryPath, tmpDir)
	})

	t.Run("Filtering", func(t *testing.T) {
		testFiltering(t, binaryPath, tmpDir)
	})

	t.Run("Grouping", func(t *testing.T) {
		testGrouping(t, binaryPath, tmpDir)
	})

	t.Run("Actions", func(t *testing.T) {
		testActions(t, binaryPath, tmpDir)
	})

	t.Run("Search", func(t *testing.T) {
		testSearch(t, binaryPath, tmpDir)
	})

	t.Run("Export", func(t *testing.T) {
		testExport(t, binaryPath, tmpDir)
	})
}

func setupTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-workflow-*")
	require.NoError(t, err)

	// Set environment variables for testing
	os.Setenv("GH_NOTIF_CONFIG", filepath.Join(tmpDir, "config.yaml"))
	os.Setenv("GH_NOTIF_CACHE_DIR", filepath.Join(tmpDir, "cache"))

	return tmpDir
}

func cleanupTestEnvironment(t *testing.T, tmpDir string) {
	os.RemoveAll(tmpDir)
	os.Unsetenv("GH_NOTIF_CONFIG")
	os.Unsetenv("GH_NOTIF_CACHE_DIR")
}

func buildTestBinary(t *testing.T, tmpDir string) string {
	binaryName := "gh-notif-test"
	if strings.Contains(os.Getenv("OS"), "Windows") {
		binaryName += ".exe"
	}

	binaryPath := filepath.Join(tmpDir, binaryName)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, ".")
	cmd.Dir = "../../" // Assuming we're in tests/system

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build binary: %s", output)

	return binaryPath
}

func testAuthentication(t *testing.T, binaryPath, tmpDir, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test auth status (should be unauthenticated)
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "status")
	output, err := cmd.CombinedOutput()
	assert.Error(t, err, "Should be unauthenticated initially")
	assert.Contains(t, string(output), "not authenticated", "Should indicate not authenticated")

	// Test auth login with token
	cmd = exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", token)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Auth login should succeed: %s", output)

	// Test auth status (should be authenticated)
	cmd = exec.CommandContext(ctx, binaryPath, "auth", "status")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Auth status should succeed: %s", output)
	assert.Contains(t, string(output), "authenticated", "Should indicate authenticated")

	// Test token validation
	assert.Contains(t, string(output), "valid", "Token should be valid")
}

func testConfiguration(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test config list
	cmd := exec.CommandContext(ctx, binaryPath, "config", "list")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Config list should succeed: %s", output)

	// Test config set
	cmd = exec.CommandContext(ctx, binaryPath, "config", "set", "display.limit", "50")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Config set should succeed: %s", output)

	// Test config get
	cmd = exec.CommandContext(ctx, binaryPath, "config", "get", "display.limit")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Config get should succeed: %s", output)
	assert.Contains(t, string(output), "50", "Should return set value")

	// Test config validation
	cmd = exec.CommandContext(ctx, binaryPath, "config", "set", "display.limit", "invalid")
	output, err = cmd.CombinedOutput()
	assert.Error(t, err, "Invalid config should fail")
}

func testNotificationListing(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test basic list
	cmd := exec.CommandContext(ctx, binaryPath, "list")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "List should succeed: %s", output)

	// Test list with limit
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--limit", "5")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "List with limit should succeed: %s", output)

	// Test list all
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--all")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "List all should succeed: %s", output)

	// Test JSON output
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--format", "json", "--limit", "1")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "List JSON should succeed: %s", output)

	// Validate JSON format
	var notifications []map[string]interface{}
	err = json.Unmarshal(output, &notifications)
	assert.NoError(t, err, "Output should be valid JSON")
}

func testFiltering(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test repository filter
	cmd := exec.CommandContext(ctx, binaryPath, "list", "--filter", "is:unread", "--limit", "5")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Filter should succeed: %s", output)

	// Test type filter
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--filter", "type:PullRequest", "--limit", "5")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Type filter should succeed: %s", output)

	// Test complex filter
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--filter", "is:unread AND type:PullRequest", "--limit", "5")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Complex filter should succeed: %s", output)

	// Test invalid filter
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--filter", "invalid:syntax")
	output, err = cmd.CombinedOutput()
	assert.Error(t, err, "Invalid filter should fail")
}

func testGrouping(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test group by repository
	cmd := exec.CommandContext(ctx, binaryPath, "group", "--by", "repository", "--limit", "10")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Group by repository should succeed: %s", output)

	// Test group by type
	cmd = exec.CommandContext(ctx, binaryPath, "group", "--by", "type", "--limit", "10")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Group by type should succeed: %s", output)

	// Test group by reason
	cmd = exec.CommandContext(ctx, binaryPath, "group", "--by", "reason", "--limit", "10")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Group by reason should succeed: %s", output)

	// Test smart grouping
	cmd = exec.CommandContext(ctx, binaryPath, "group", "--by", "smart", "--limit", "10")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Smart grouping should succeed: %s", output)
}

func testActions(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get a notification ID for testing
	cmd := exec.CommandContext(ctx, binaryPath, "list", "--format", "json", "--limit", "1")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Should get notifications for testing: %s", output)

	var notifications []map[string]interface{}
	err = json.Unmarshal(output, &notifications)
	require.NoError(t, err, "Should parse JSON")

	if len(notifications) == 0 {
		t.Skip("No notifications available for action testing")
	}

	notificationID := notifications[0]["id"].(string)

	// Test mark as read (dry run)
	cmd = exec.CommandContext(ctx, binaryPath, "read", notificationID, "--dry-run")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Mark as read (dry run) should succeed: %s", output)

	// Test open (dry run)
	cmd = exec.CommandContext(ctx, binaryPath, "open", notificationID, "--dry-run")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Open (dry run) should succeed: %s", output)

	// Test subscribe (dry run)
	cmd = exec.CommandContext(ctx, binaryPath, "subscribe", notificationID, "--dry-run")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Subscribe (dry run) should succeed: %s", output)
}

func testSearch(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test basic search
	cmd := exec.CommandContext(ctx, binaryPath, "search", "test", "--limit", "5")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Search should succeed: %s", output)

	// Test search with filter
	cmd = exec.CommandContext(ctx, binaryPath, "search", "bug", "--filter", "is:unread", "--limit", "5")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Search with filter should succeed: %s", output)

	// Test regex search
	cmd = exec.CommandContext(ctx, binaryPath, "search", "fix.*bug", "--regex", "--limit", "5")
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "Regex search should succeed: %s", output)
}

func testExport(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test JSON export
	exportFile := filepath.Join(tmpDir, "export.json")
	cmd := exec.CommandContext(ctx, binaryPath, "list", "--format", "json", "--limit", "5", "--output", exportFile)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "JSON export should succeed: %s", output)

	// Verify export file exists and is valid JSON
	data, err := os.ReadFile(exportFile)
	require.NoError(t, err, "Export file should exist")

	var notifications []map[string]interface{}
	err = json.Unmarshal(data, &notifications)
	assert.NoError(t, err, "Export file should contain valid JSON")

	// Test CSV export
	csvFile := filepath.Join(tmpDir, "export.csv")
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--format", "csv", "--limit", "5", "--output", csvFile)
	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "CSV export should succeed: %s", output)

	// Verify CSV file exists
	_, err = os.Stat(csvFile)
	assert.NoError(t, err, "CSV export file should exist")
}

// TestCrossPlatformCompatibility tests platform-specific behavior
func TestCrossPlatformCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-platform tests in short mode")
	}

	tmpDir := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("Path Handling", func(t *testing.T) {
		testPathHandling(t, binaryPath, tmpDir)
	})

	t.Run("File Permissions", func(t *testing.T) {
		testFilePermissions(t, binaryPath, tmpDir)
	})

	t.Run("Environment Variables", func(t *testing.T) {
		testEnvironmentVariables(t, binaryPath, tmpDir)
	})
}

func testPathHandling(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with different path separators
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	cmd := exec.CommandContext(ctx, binaryPath, "config", "set", "display.limit", "10", "--config", configPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Path handling should work: %s", output)

	// Verify config file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Config file should be created")
}

func testFilePermissions(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test config file permissions
	configPath := filepath.Join(tmpDir, "permissions-test.yaml")

	cmd := exec.CommandContext(ctx, binaryPath, "config", "set", "test.value", "test", "--config", configPath)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Should create config file: %s", output)

	// Check file permissions (Unix-like systems)
	if strings.Contains(os.Getenv("OS"), "Windows") == false {
		info, err := os.Stat(configPath)
		require.NoError(t, err, "Should stat config file")

		mode := info.Mode()
		assert.Equal(t, os.FileMode(0600), mode&0777, "Config file should have restrictive permissions")
	}
}

func testEnvironmentVariables(t *testing.T, binaryPath, tmpDir string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with environment variable override
	customConfig := filepath.Join(tmpDir, "env-config.yaml")

	cmd := exec.CommandContext(ctx, binaryPath, "config", "set", "env.test", "value")
	cmd.Env = append(os.Environ(), fmt.Sprintf("GH_NOTIF_CONFIG=%s", customConfig))

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Environment variable should work: %s", output)

	// Verify config was written to custom location
	_, err = os.Stat(customConfig)
	assert.NoError(t, err, "Custom config file should exist")
}
