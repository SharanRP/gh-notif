package integration

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/SharanRP/gh-notif/cmd/gh-notif/auth"
	"github.com/SharanRP/gh-notif/internal/testutil"
	"github.com/spf13/cobra"
)

// setupAuthTest sets up a test environment for auth commands
func setupAuthTest(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory for testing
	tempDir, cleanup := testutil.TempDir(t)

	// Save original home directory and restore after test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	// Create a cleanup function
	cleanupAll := func() {
		os.Setenv("HOME", originalHome)
		cleanup()
	}

	return tempDir, cleanupAll
}

// executeCommand executes a cobra command for testing
func executeCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()

	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Set args
	cmd.SetArgs(args)

	// Execute the command
	err := cmd.Execute()

	// Return the output
	return buf.String(), err
}

func TestAuthLoginCommand(t *testing.T) {
	// Set up test environment
	_, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create the auth command
	rootCmd := &cobra.Command{Use: "gh-notif"}
	authCmd := auth.NewAuthCmd()
	rootCmd.AddCommand(authCmd)

	// Execute the login command
	output, err := executeCommand(t, rootCmd, "auth", "login")
	if err != nil {
		t.Fatalf("login command failed: %v", err)
	}

	// Check the output
	if !strings.Contains(output, "To authenticate with GitHub, please:") {
		t.Errorf("login output does not contain expected instructions")
	}
	if !strings.Contains(output, "Enter this code:") {
		t.Errorf("login output does not contain user code")
	}
}

func TestAuthStatusCommand(t *testing.T) {
	// Set up test environment
	_, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create the auth command
	rootCmd := &cobra.Command{Use: "gh-notif"}
	authCmd := auth.NewAuthCmd()
	rootCmd.AddCommand(authCmd)

	// Execute the status command
	output, err := executeCommand(t, rootCmd, "auth", "status")
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	// Check the output
	if !strings.Contains(output, "Authenticated") {
		t.Errorf("status output does not indicate authenticated")
	}
	if !strings.Contains(output, "test-access-token") {
		t.Errorf("status output does not contain access token")
	}
}

func TestAuthLogoutCommand(t *testing.T) {
	// Set up test environment
	_, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create the auth command
	rootCmd := &cobra.Command{Use: "gh-notif"}
	authCmd := auth.NewAuthCmd()
	rootCmd.AddCommand(authCmd)

	// Execute the logout command
	output, err := executeCommand(t, rootCmd, "auth", "logout")
	if err != nil {
		t.Fatalf("logout command failed: %v", err)
	}

	// Check the output
	if !strings.Contains(output, "Logged out") {
		t.Errorf("logout output does not indicate success")
	}
}

func TestAuthRefreshCommand(t *testing.T) {
	// Set up test environment
	_, cleanup := setupAuthTest(t)
	defer cleanup()

	// Create the auth command
	rootCmd := &cobra.Command{Use: "gh-notif"}
	authCmd := auth.NewAuthCmd()
	rootCmd.AddCommand(authCmd)

	// Execute the refresh command
	output, err := executeCommand(t, rootCmd, "auth", "refresh")
	if err != nil {
		t.Fatalf("refresh command failed: %v", err)
	}

	// Check the output
	if !strings.Contains(output, "Token refreshed") {
		t.Errorf("refresh output does not indicate success")
	}
}
