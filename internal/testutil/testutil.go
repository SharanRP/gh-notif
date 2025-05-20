package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

// TempDir creates a temporary directory for testing and returns a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "gh-notif-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

// SetupTestConfig creates a test configuration in a temporary directory
func SetupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tempDir, cleanup := TempDir(t)

	// Create a test config file
	configPath := filepath.Join(tempDir, ".gh-notif.yaml")

	// Create a new viper instance
	v := viper.New()
	v.SetConfigFile(configPath)

	// Set some test values
	v.Set("auth.client_id", "test-client-id")
	v.Set("auth.client_secret", "test-client-secret")
	v.Set("auth.scopes", []string{"notifications", "repo"})
	v.Set("auth.token_storage", "file")

	// Save the config
	if err := v.WriteConfigAs(configPath); err != nil {
		cleanup()
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set environment variable to point to our test config
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	// Return a cleanup function that restores the environment
	cleanupAll := func() {
		os.Setenv("HOME", oldHome)
		cleanup()
	}

	return tempDir, cleanupAll
}

// CreateTestToken creates a test OAuth2 token
func CreateTestToken(t *testing.T, expired bool) *oauth2.Token {
	t.Helper()

	var expiry time.Time
	if expired {
		expiry = time.Now().Add(-1 * time.Hour)
	} else {
		expiry = time.Now().Add(1 * time.Hour)
	}

	return &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       expiry,
	}
}

// MockGitHubAPI creates a mock GitHub API server
func MockGitHubAPI(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Add handlers
	for path, handler := range handlers {
		mux.HandleFunc(path, handler)
	}

	// Create a test server
	server := httptest.NewServer(mux)

	return server
}

// MockDeviceFlowServer creates a mock server for GitHub's device flow
func MockDeviceFlowServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Device code endpoint
	mux.HandleFunc("/login/device/code", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"device_code":      "test-device-code",
			"user_code":        "ABCD-1234",
			"verification_uri": "https://github.com/login/device",
			"expires_in":       900,
			"interval":         5,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Token endpoint
	var tokenRequests int
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tokenRequests++

		// Parse the request
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check device code
		deviceCode := r.FormValue("device_code")
		if deviceCode != "test-device-code" {
			response := map[string]string{
				"error": "invalid_device_code",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// First request: authorization pending
		if tokenRequests == 1 {
			response := map[string]string{
				"error": "authorization_pending",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Second request: slow down
		if tokenRequests == 2 {
			response := map[string]string{
				"error": "slow_down",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Third request: success
		response := map[string]interface{}{
			"access_token":  "test-access-token",
			"token_type":    "bearer",
			"scope":         "notifications repo",
			"refresh_token": "test-refresh-token",
			"expires_in":    3600,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	return httptest.NewServer(mux)
}

// ReadTestFile reads a test file from the testdata directory
func ReadTestFile(t *testing.T, filename string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filename, err)
	}

	return data
}

// CaptureOutput captures stdout and stderr during a test
func CaptureOutput(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()

	// Redirect stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	os.Stdout = wOut
	os.Stderr = wErr

	// Run the function
	f()

	// Restore stdout and stderr
	wOut.Close()
	wErr.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read captured output
	outBytes, err := io.ReadAll(rOut)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}

	errBytes, err := io.ReadAll(rErr)
	if err != nil {
		t.Fatalf("Failed to read stderr: %v", err)
	}

	return string(outBytes), string(errBytes)
}

// SetEnvVars sets environment variables for testing and returns a cleanup function
func SetEnvVars(t *testing.T, vars map[string]string) func() {
	t.Helper()

	// Save old values
	oldValues := make(map[string]string)
	for k := range vars {
		oldValues[k] = os.Getenv(k)
	}

	// Set new values
	for k, v := range vars {
		os.Setenv(k, v)
	}

	// Return cleanup function
	return func() {
		for k, v := range oldValues {
			os.Setenv(k, v)
		}
	}
}
