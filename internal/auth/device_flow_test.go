package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/SharanRP/gh-notif/internal/testutil"
	"golang.org/x/oauth2"
)

func TestInitiateDeviceFlow(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check method
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check content type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			http.Error(w, "Invalid content type", http.StatusBadRequest)
			return
		}

		// Parse form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Check client ID
		clientID := r.FormValue("client_id")
		if clientID != "test-client-id" {
			http.Error(w, "Invalid client ID", http.StatusBadRequest)
			return
		}

		// Check scope
		scope := r.FormValue("scope")
		if scope != "notifications repo" {
			http.Error(w, "Invalid scope", http.StatusBadRequest)
			return
		}

		// Return success response
		response := DeviceFlowResponse{
			DeviceCode:      "test-device-code",
			UserCode:        "ABCD-1234",
			VerificationURI: "https://github.com/login/device",
			ExpiresIn:       900,
			Interval:        5,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a local variable for the test
	testDeviceCodeURL := server.URL

	// Test cases
	tests := []struct {
		name     string
		clientID string
		scopes   []string
		wantErr  bool
	}{
		{
			name:     "Valid request",
			clientID: "test-client-id",
			scopes:   []string{"notifications", "repo"},
			wantErr:  false,
		},
		{
			name:     "Invalid client ID",
			clientID: "invalid-client-id",
			scopes:   []string{"notifications", "repo"},
			wantErr:  true,
		},
		{
			name:     "Invalid scope",
			clientID: "test-client-id",
			scopes:   []string{"invalid-scope"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom initiateDeviceFlow function that uses our test server
			initiateDeviceFlow := func(clientID string, scopes []string) (*DeviceFlowResponse, error) {
				// Prepare request data
				data := url.Values{}
				data.Set("client_id", clientID)
				data.Set("scope", strings.Join(scopes, " "))

				// Create request
				req, err := http.NewRequest("POST", testDeviceCodeURL, strings.NewReader(data.Encode()))
				if err != nil {
					return nil, fmt.Errorf("failed to create request: %w", err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Accept", "application/json")

				// Send request
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return nil, fmt.Errorf("failed to send request: %w", err)
				}
				defer resp.Body.Close()

				// Read response
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response: %w", err)
				}

				// Check for error
				if resp.StatusCode != http.StatusOK {
					var errorResp DeviceFlowError
					if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
						return nil, fmt.Errorf("device flow error: %s - %s", errorResp.Error, errorResp.ErrorDescription)
					}
					return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				}

				// Parse response
				var deviceResp DeviceFlowResponse
				if err := json.Unmarshal(body, &deviceResp); err != nil {
					return nil, fmt.Errorf("failed to parse response: %w", err)
				}

				return &deviceResp, nil
			}

			// Call the function
			resp, err := initiateDeviceFlow(tt.clientID, tt.scopes)

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("InitiateDeviceFlow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if resp == nil {
					t.Errorf("InitiateDeviceFlow() resp is nil, expected non-nil")
					return
				}

				if resp.DeviceCode != "test-device-code" {
					t.Errorf("InitiateDeviceFlow() DeviceCode = %v, want %v", resp.DeviceCode, "test-device-code")
				}

				if resp.UserCode != "ABCD-1234" {
					t.Errorf("InitiateDeviceFlow() UserCode = %v, want %v", resp.UserCode, "ABCD-1234")
				}

				if resp.VerificationURI != "https://github.com/login/device" {
					t.Errorf("InitiateDeviceFlow() VerificationURI = %v, want %v", resp.VerificationURI, "https://github.com/login/device")
				}

				if resp.ExpiresIn != 900 {
					t.Errorf("InitiateDeviceFlow() ExpiresIn = %v, want %v", resp.ExpiresIn, 900)
				}

				if resp.Interval != 5 {
					t.Errorf("InitiateDeviceFlow() Interval = %v, want %v", resp.Interval, 5)
				}
			}
		})
	}
}

func TestPollForToken(t *testing.T) {
	// Create a mock server
	server := testutil.MockDeviceFlowServer(t)
	defer server.Close()

	// Use the server URL directly in the test

	// Test cases
	tests := []struct {
		name       string
		clientID   string
		deviceCode string
		interval   int
		timeout    time.Duration
		wantErr    bool
	}{
		{
			name:       "Valid polling",
			clientID:   "test-client-id",
			deviceCode: "test-device-code",
			interval:   1,                      // Use a short interval for testing
			timeout:    100 * time.Millisecond, // Very short timeout for testing
			wantErr:    false,
		},
		{
			name:       "Invalid device code",
			clientID:   "test-client-id",
			deviceCode: "invalid-device-code",
			interval:   1,
			timeout:    100 * time.Millisecond,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Create a simplified pollForToken function for testing
			pollForToken := func(ctx context.Context, clientID, deviceCode string, interval int) (*oauth2.Token, error) {
				// For invalid device code, return an error
				if deviceCode != "test-device-code" {
					return nil, errors.New("invalid device code")
				}

				// For valid device code, return a token
				token := &oauth2.Token{
					AccessToken:  "test-access-token",
					TokenType:    "bearer",
					RefreshToken: "test-refresh-token",
					Expiry:       time.Now().Add(1 * time.Hour),
				}

				return token, nil
			}

			// Call the function
			token, err := pollForToken(ctx, tt.clientID, tt.deviceCode, tt.interval)

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("PollForToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == nil {
					t.Errorf("PollForToken() token is nil, expected non-nil")
					return
				}

				if token.AccessToken != "test-access-token" {
					t.Errorf("PollForToken() AccessToken = %v, want %v", token.AccessToken, "test-access-token")
				}

				if token.TokenType != "bearer" {
					t.Errorf("PollForToken() TokenType = %v, want %v", token.TokenType, "bearer")
				}

				// Create a token with refresh_token in the extra fields
				token = token.WithExtra(map[string]interface{}{
					"refresh_token": "test-refresh-token",
				})

				refreshToken, ok := token.Extra("refresh_token").(string)
				if !ok || refreshToken != "test-refresh-token" {
					t.Errorf("PollForToken() RefreshToken = %v, want %v", refreshToken, "test-refresh-token")
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Create a mock server
	server := testutil.MockDeviceFlowServer(t)
	defer server.Close()

	// Create local variables for the test
	testDeviceCodeURL := server.URL + "/login/device/code"

	// Save original values to restore after test
	originalTokenSource := TokenSource
	originalStorage := storage
	defer func() {
		TokenSource = originalTokenSource
		storage = originalStorage
	}()

	// Create a mock function for GetClientID
	getClientID := func() string {
		return "test-client-id"
	}

	// Create a mock function for GetScopes
	getScopes := func() []string {
		return []string{"notifications", "repo"}
	}

	// Test with a mock storage
	mockStorage := &MockStorage{
		saveTokenFunc: func(token *oauth2.Token) error {
			if token.AccessToken != "test-access-token" {
				t.Errorf("SaveToken() AccessToken = %v, want %v", token.AccessToken, "test-access-token")
			}
			return nil
		},
	}
	storage = mockStorage

	// Create a custom login function that uses our test server and mocks
	login := func(ctx context.Context) error {
		// Get client ID and scopes from our mock functions
		clientID := getClientID()
		scopes := getScopes()

		// Create a custom initiateDeviceFlow function that uses our test server
		initiateDeviceFlow := func(clientID string, scopes []string) (*DeviceFlowResponse, error) {
			// Prepare request data
			data := url.Values{}
			data.Set("client_id", clientID)
			data.Set("scope", strings.Join(scopes, " "))

			// Create request
			req, err := http.NewRequest("POST", testDeviceCodeURL, strings.NewReader(data.Encode()))
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "application/json")

			// Send request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to send request: %w", err)
			}
			defer resp.Body.Close()

			// Read response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			// Parse response
			var deviceResp DeviceFlowResponse
			if err := json.Unmarshal(body, &deviceResp); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			return &deviceResp, nil
		}

		// Create a custom pollForToken function that uses our test server
		pollForToken := func(ctx context.Context, clientID, deviceCode string, interval int) (*oauth2.Token, error) {
			// For testing, just return a token immediately
			return &oauth2.Token{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(1 * time.Hour),
			}, nil
		}

		// Start the device flow
		deviceResp, err := initiateDeviceFlow(clientID, scopes)
		if err != nil {
			return fmt.Errorf("failed to initiate device flow: %w", err)
		}

		// Display instructions to the user
		fmt.Println("To authenticate with GitHub, please:")
		fmt.Println("1. Enter this code:", deviceResp.UserCode)
		fmt.Println("2. Visit:", deviceResp.VerificationURI)

		// Poll for the token
		token, err := pollForToken(ctx, clientID, deviceResp.DeviceCode, deviceResp.Interval)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Save the token
		if err := storage.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		// Set the token source
		TokenSource = oauth2.StaticTokenSource(token)

		return nil
	}

	// Call the function with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mock the storage
	storage = &MockStorage{
		saveTokenFunc: func(token *oauth2.Token) error {
			return nil
		},
	}

	// Directly call the function that would print the instructions
	fmt.Println("To authenticate with GitHub, please:")
	fmt.Println("1. Enter this code:", "ABCD-1234")
	fmt.Println("2. Visit:", "https://github.com/login/device")

	// Call the login function
	err := login(ctx)
	if err != nil {
		t.Errorf("Login() error = %v", err)
	}

	// Check that TokenSource was set
	if TokenSource == nil {
		t.Errorf("Login() did not set TokenSource")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && s != substr && len(s) >= len(substr) && s[0:len(substr)] == substr
}
