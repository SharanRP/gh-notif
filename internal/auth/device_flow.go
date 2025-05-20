package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

const (
	// GitHub Device Flow endpoints
	deviceCodeURL = "https://github.com/login/device/code"
	tokenURL      = "https://github.com/login/oauth/access_token"
)

// DeviceFlowResponse represents the response from the device flow initiation
type DeviceFlowResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceFlowError represents an error from the device flow
type DeviceFlowError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// InitiateDeviceFlow starts the GitHub OAuth device flow
func InitiateDeviceFlow(clientID string, scopes []string) (*DeviceFlowResponse, error) {
	// Prepare request data
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", strings.Join(scopes, " "))

	// Create request
	req, err := http.NewRequest("POST", deviceCodeURL, strings.NewReader(data.Encode()))
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

// PollForToken polls GitHub for an access token using the device code
func PollForToken(ctx context.Context, clientID, deviceCode string, interval int) (*oauth2.Token, error) {
	// Prepare request data
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Create request
			req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
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

			// Read response
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			// Parse response
			var tokenResp struct {
				AccessToken  string `json:"access_token"`
				TokenType    string `json:"token_type"`
				Scope        string `json:"scope"`
				RefreshToken string `json:"refresh_token,omitempty"`
				ExpiresIn    int    `json:"expires_in,omitempty"`
				Error        string `json:"error,omitempty"`
			}
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			// Check for error
			if tokenResp.Error != "" {
				if tokenResp.Error == "authorization_pending" {
					// User hasn't authorized yet, continue polling
					continue
				}
				if tokenResp.Error == "slow_down" {
					// GitHub is asking us to slow down, increase the interval
					interval += 5
					ticker.Reset(time.Duration(interval) * time.Second)
					continue
				}
				if tokenResp.Error == "expired_token" {
					return nil, errors.New("device code expired, please try again")
				}
				return nil, fmt.Errorf("token error: %s", tokenResp.Error)
			}

			// Success! Create token
			token := &oauth2.Token{
				AccessToken: tokenResp.AccessToken,
				TokenType:   tokenResp.TokenType,
			}

			// Set expiry if provided
			if tokenResp.ExpiresIn > 0 {
				token.Expiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
			}

			// Set refresh token if provided
			if tokenResp.RefreshToken != "" {
				token = token.WithExtra(map[string]interface{}{
					"refresh_token": tokenResp.RefreshToken,
				})
			}

			return token, nil
		}
	}
}

// OpenBrowser attempts to open the verification URL in the default browser
func OpenBrowser(url string) error {
	return browser.OpenURL(url)
}
