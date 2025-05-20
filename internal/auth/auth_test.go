package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/user/gh-notif/internal/testutil"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

func TestStatus(t *testing.T) {
	// Save original values to restore after test
	originalTokenSource := TokenSource
	originalStorage := storage
	defer func() {
		TokenSource = originalTokenSource
		storage = originalStorage
	}()

	tests := []struct {
		name           string
		setupTokenSource func()
		setupStorage    func() Storage
		wantAuthenticated bool
		wantToken       bool
		wantErr         bool
	}{
		{
			name: "Valid token in TokenSource",
			setupTokenSource: func() {
				token := testutil.CreateTestToken(t, false)
				TokenSource = oauth2.StaticTokenSource(token)
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, errors.New("should not be called")
					},
				}
			},
			wantAuthenticated: true,
			wantToken:       true,
			wantErr:         false,
		},
		{
			name: "Expired token in TokenSource",
			setupTokenSource: func() {
				token := testutil.CreateTestToken(t, true)
				TokenSource = oauth2.StaticTokenSource(token)
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, errors.New("should not be called")
					},
				}
			},
			wantAuthenticated: false,
			wantToken:       true,
			wantErr:         false,
		},
		{
			name: "Valid token in storage",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return testutil.CreateTestToken(t, false), nil
					},
				}
			},
			wantAuthenticated: true,
			wantToken:       true,
			wantErr:         false,
		},
		{
			name: "Expired token in storage",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return testutil.CreateTestToken(t, true), nil
					},
				}
			},
			wantAuthenticated: false,
			wantToken:       true,
			wantErr:         false,
		},
		{
			name: "No token",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, ErrNoToken
					},
				}
			},
			wantAuthenticated: false,
			wantToken:       false,
			wantErr:         false,
		},
		{
			name: "Storage error",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, errors.New("storage error")
					},
				}
			},
			wantAuthenticated: false,
			wantToken:       false,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupTokenSource()
			storage = tt.setupStorage()

			// Call the function
			authenticated, token, err := Status()

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Status() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if authenticated != tt.wantAuthenticated {
				t.Errorf("Status() authenticated = %v, want %v", authenticated, tt.wantAuthenticated)
			}
			if (token != nil) != tt.wantToken {
				t.Errorf("Status() token = %v, want token present: %v", token, tt.wantToken)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	// Save original values to restore after test
	originalTokenSource := TokenSource
	originalStorage := storage
	defer func() {
		TokenSource = originalTokenSource
		storage = originalStorage
	}()

	tests := []struct {
		name        string
		setupStorage func() Storage
		wantErr     bool
	}{
		{
			name: "Successful logout",
			setupStorage: func() Storage {
				return &MockStorage{
					deleteTokenFunc: func() error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Error during logout",
			setupStorage: func() Storage {
				return &MockStorage{
					deleteTokenFunc: func() error {
						return errors.New("delete error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			TokenSource = oauth2.StaticTokenSource(testutil.CreateTestToken(t, false))
			storage = tt.setupStorage()

			// Call the function
			err := Logout()

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if TokenSource != nil {
				t.Errorf("Logout() did not clear TokenSource")
			}
		})
	}
}

func TestGetClient(t *testing.T) {
	// Save original values to restore after test
	originalTokenSource := TokenSource
	originalStorage := storage
	defer func() {
		TokenSource = originalTokenSource
		storage = originalStorage
	}()

	tests := []struct {
		name           string
		setupTokenSource func()
		setupStorage    func() Storage
		wantErr         bool
	}{
		{
			name: "Valid token in TokenSource",
			setupTokenSource: func() {
				token := testutil.CreateTestToken(t, false)
				TokenSource = oauth2.StaticTokenSource(token)
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, errors.New("should not be called")
					},
				}
			},
			wantErr: false,
		},
		{
			name: "Valid token in storage",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return testutil.CreateTestToken(t, false), nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "No token",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, ErrNoToken
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupTokenSource()
			storage = tt.setupStorage()

			// Call the function
			client, err := GetClient(context.Background())

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Errorf("GetClient() client is nil, expected non-nil")
			}
		})
	}
}

// MockStorage is a mock implementation of the Storage interface for testing
type MockStorage struct {
	saveTokenFunc  func(token *oauth2.Token) error
	loadTokenFunc  func() (*oauth2.Token, error)
	deleteTokenFunc func() error
}

func (m *MockStorage) SaveToken(token *oauth2.Token) error {
	if m.saveTokenFunc != nil {
		return m.saveTokenFunc(token)
	}
	return nil
}

func (m *MockStorage) LoadToken() (*oauth2.Token, error) {
	if m.loadTokenFunc != nil {
		return m.loadTokenFunc()
	}
	return nil, nil
}

func (m *MockStorage) DeleteToken() error {
	if m.deleteTokenFunc != nil {
		return m.deleteTokenFunc()
	}
	return nil
}

func TestRefreshToken(t *testing.T) {
	// Save original values to restore after test
	originalTokenSource := TokenSource
	originalStorage := storage
	originalGetClientID := GetClientID
	defer func() {
		TokenSource = originalTokenSource
		storage = originalStorage
		GetClientID = originalGetClientID
	}()

	// Create a test token with a refresh token
	token := testutil.CreateTestToken(t, false)
	token = token.WithExtra(map[string]interface{}{
		"refresh_token": "test-refresh-token",
	})

	// Mock GetClientID
	GetClientID = func() string {
		return "test-client-id"
	}

	// Create a mock HTTP server for token refresh
	server := testutil.MockGitHubAPI(t, map[string]http.HandlerFunc{
		"/login/oauth/access_token": func(w http.ResponseWriter, r *http.Request) {
			// Check method
			if r.Method != "POST" {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

			// Check refresh token
			refreshToken := r.FormValue("refresh_token")
			if refreshToken != "test-refresh-token" {
				http.Error(w, "Invalid refresh token", http.StatusBadRequest)
				return
			}

			// Return a new token
			response := map[string]interface{}{
				"access_token":  "new-access-token",
				"token_type":    "bearer",
				"scope":         "notifications repo",
				"refresh_token": "new-refresh-token",
				"expires_in":    3600,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		},
	})
	defer server.Close()

	// Override the GitHub endpoint for testing
	originalEndpoint := github.Endpoint
	github.Endpoint = oauth2.Endpoint{
		TokenURL: server.URL + "/login/oauth/access_token",
	}
	defer func() {
		github.Endpoint = originalEndpoint
	}()

	tests := []struct {
		name           string
		setupTokenSource func()
		setupStorage    func() Storage
		wantErr         bool
	}{
		{
			name: "Valid refresh token",
			setupTokenSource: func() {
				TokenSource = oauth2.StaticTokenSource(token)
			},
			setupStorage: func() Storage {
				return &MockStorage{
					saveTokenFunc: func(token *oauth2.Token) error {
						if token.AccessToken != "new-access-token" {
							t.Errorf("SaveToken() AccessToken = %v, want %v", token.AccessToken, "new-access-token")
						}
						refreshToken, ok := token.Extra("refresh_token").(string)
						if !ok || refreshToken != "new-refresh-token" {
							t.Errorf("SaveToken() RefreshToken = %v, want %v", refreshToken, "new-refresh-token")
						}
						return nil
					},
				}
			},
			wantErr: false,
		},
		{
			name: "No token source",
			setupTokenSource: func() {
				TokenSource = nil
			},
			setupStorage: func() Storage {
				return &MockStorage{
					loadTokenFunc: func() (*oauth2.Token, error) {
						return nil, ErrNoToken
					},
				}
			},
			wantErr: true,
		},
		{
			name: "No refresh token",
			setupTokenSource: func() {
				TokenSource = oauth2.StaticTokenSource(testutil.CreateTestToken(t, false))
			},
			setupStorage: func() Storage {
				return &MockStorage{}
			},
			wantErr: true,
		},
		{
			name: "Storage error",
			setupTokenSource: func() {
				TokenSource = oauth2.StaticTokenSource(token)
			},
			setupStorage: func() Storage {
				return &MockStorage{
					saveTokenFunc: func(token *oauth2.Token) error {
						return errors.New("storage error")
					},
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupTokenSource()
			storage = tt.setupStorage()

			// Call the function
			err := RefreshToken(context.Background())

			// Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For successful refresh, check that TokenSource was updated
			if !tt.wantErr {
				if TokenSource == nil {
					t.Errorf("RefreshToken() did not update TokenSource")
					return
				}

				// Get the token from the token source
				newToken, err := TokenSource.Token()
				if err != nil {
					t.Errorf("Token() error = %v", err)
					return
				}

				if newToken.AccessToken != "new-access-token" {
					t.Errorf("RefreshToken() AccessToken = %v, want %v", newToken.AccessToken, "new-access-token")
				}
			}
		})
	}
}
