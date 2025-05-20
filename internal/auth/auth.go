package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	// TokenSource holds the OAuth2 token source
	TokenSource oauth2.TokenSource

	// storage is the token storage implementation
	storage Storage

	// ErrNotAuthenticated is returned when the user is not authenticated
	ErrNotAuthenticated = errors.New("not authenticated")
)

// init initializes the auth package
func init() {
	var err error
	// Create storage based on configuration
	storage, err = CreateStorage()
	if err != nil {
		// If we can't create storage, we'll initialize it when needed
		return
	}

	// Try to load the token
	token, err := storage.LoadToken()
	if err == nil && token.Valid() {
		TokenSource = oauth2.StaticTokenSource(token)
	}
}

// Login performs the GitHub OAuth2 device flow authentication
func Login(ctx context.Context) error {
	// Get client ID and scopes from config
	clientID := GetClientID()
	if clientID == "" {
		return fmt.Errorf("GitHub client ID not configured. Please set it with 'gh-notif config set auth.client_id YOUR_CLIENT_ID'")
	}

	// Get scopes from config
	scopes := GetScopes()
	if len(scopes) == 0 {
		scopes = []string{"notifications", "repo"}
	}

	// Initialize storage if needed
	if storage == nil {
		var err error
		storage, err = CreateStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize token storage: %w", err)
		}
	}

	// Start the device flow
	deviceResp, err := InitiateDeviceFlow(clientID, scopes)
	if err != nil {
		return fmt.Errorf("failed to initiate device flow: %w", err)
	}

	// Display instructions to the user
	fmt.Println("To authenticate with GitHub, please:")
	fmt.Println("1. Enter this code:", deviceResp.UserCode)
	fmt.Println("2. Visit:", deviceResp.VerificationURI)

	// Try to open the browser
	if err := OpenBrowser(deviceResp.VerificationURI); err == nil {
		fmt.Println("Browser opened automatically. If it didn't open, please visit the URL above.")
	}

	fmt.Println("Waiting for authentication...")

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(deviceResp.ExpiresIn)*time.Second)
	defer cancel()

	// Poll for the token
	token, err := PollForToken(timeoutCtx, clientID, deviceResp.DeviceCode, deviceResp.Interval)
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

// Logout removes the stored credentials
func Logout() error {
	TokenSource = nil

	// Initialize storage if needed
	if storage == nil {
		var err error
		storage, err = CreateStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize token storage: %w", err)
		}
	}

	return storage.DeleteToken()
}

// Status checks the authentication status
func Status() (bool, *oauth2.Token, error) {
	if TokenSource == nil {
		// Initialize storage if needed
		if storage == nil {
			var err error
			storage, err = CreateStorage()
			if err != nil {
				return false, nil, fmt.Errorf("failed to initialize token storage: %w", err)
			}
		}

		token, err := storage.LoadToken()
		if err != nil {
			if errors.Is(err, ErrNoToken) {
				return false, nil, nil
			}
			return false, nil, err
		}

		if token.Valid() {
			TokenSource = oauth2.StaticTokenSource(token)
			return true, token, nil
		}

		return false, token, nil
	}

	// Get the token from the token source
	token, err := TokenSource.Token()
	if err != nil {
		return false, nil, err
	}

	return token.Valid(), token, nil
}

// RefreshToken attempts to refresh the OAuth token
func RefreshToken(ctx context.Context) error {
	// Check if we have a token
	authenticated, token, err := Status()
	if err != nil {
		return err
	}

	if !authenticated {
		return ErrNotAuthenticated
	}

	// Check if the token has a refresh token
	refreshToken, ok := token.Extra("refresh_token").(string)
	if !ok || refreshToken == "" {
		return errors.New("no refresh token available, please login again")
	}

	// Get client ID from config
	clientID := GetClientID()
	if clientID == "" {
		return fmt.Errorf("GitHub client ID not configured. Please set it with 'gh-notif config set auth.client_id YOUR_CLIENT_ID'")
	}

	// Create a token source with the refresh token
	config := &oauth2.Config{
		ClientID: clientID,
		Endpoint: github.Endpoint,
	}

	// Create a new token source with the refresh token
	ts := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	// Get a new token
	newToken, err := ts.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Initialize storage if needed
	if storage == nil {
		var err error
		storage, err = CreateStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize token storage: %w", err)
		}
	}

	// Save the new token
	if err := storage.SaveToken(newToken); err != nil {
		return fmt.Errorf("failed to save refreshed token: %w", err)
	}

	// Update the token source
	TokenSource = oauth2.StaticTokenSource(newToken)

	return nil
}

// GetClient returns an HTTP client with the OAuth2 token
func GetClient(ctx context.Context) (*http.Client, error) {
	if TokenSource == nil {
		// Initialize storage if needed
		if storage == nil {
			var err error
			storage, err = CreateStorage()
			if err != nil {
				return nil, fmt.Errorf("failed to initialize token storage: %w", err)
			}
		}

		// Try to load the token
		authenticated, _, err := Status()
		if err != nil {
			return nil, err
		}

		if !authenticated {
			return nil, ErrNotAuthenticated
		}
	}

	// Create a client with the token source
	return oauth2.NewClient(ctx, TokenSource), nil
}

// GetClientOrExit returns an HTTP client or exits if not authenticated
func GetClientOrExit(ctx context.Context) *http.Client {
	client, err := GetClient(ctx)
	if err != nil {
		if errors.Is(err, ErrNotAuthenticated) {
			fmt.Fprintln(os.Stderr, "Error: Not authenticated. Please run 'gh-notif auth login' first.")
		} else {
			fmt.Fprintf(os.Stderr, "Authentication error: %v\n", err)
		}
		os.Exit(1)
	}
	return client
}
