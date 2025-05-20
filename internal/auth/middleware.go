package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// AuthMiddleware provides middleware for authenticated requests
type AuthMiddleware struct {
	// MaxRetries is the maximum number of retries for token refresh
	MaxRetries int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
	// refreshFunc is the function to call to refresh the token
	refreshFunc func(ctx context.Context) error
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		refreshFunc: RefreshToken,
	}
}

// RoundTripper returns an http.RoundTripper that handles authentication
func (m *AuthMiddleware) RoundTripper(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return &authRoundTripper{
		base:       base,
		middleware: m,
	}
}

// authRoundTripper is an http.RoundTripper that handles authentication
type authRoundTripper struct {
	base       http.RoundTripper
	middleware *AuthMiddleware
}

// RoundTrip implements http.RoundTripper
func (rt *authRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Make a copy of the request to avoid modifying the original
	reqCopy := req.Clone(req.Context())

	// Try the request
	resp, err := rt.tryRequest(reqCopy, 0)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// tryRequest tries to make a request, refreshing the token if needed
func (rt *authRoundTripper) tryRequest(req *http.Request, retryCount int) (*http.Response, error) {
	// Make the request
	resp, err := rt.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Check if we need to refresh the token
	if resp.StatusCode == http.StatusUnauthorized && retryCount < rt.middleware.MaxRetries {
		// Close the response body to avoid leaking resources
		resp.Body.Close()

		// Refresh the token
		refreshFunc := rt.middleware.refreshFunc
		if refreshFunc == nil {
			refreshFunc = RefreshToken
		}
		if err := refreshFunc(req.Context()); err != nil {
			if errors.Is(err, ErrNotAuthenticated) {
				return nil, fmt.Errorf("not authenticated: %w", err)
			}
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		// Wait before retrying
		time.Sleep(rt.middleware.RetryDelay)

		// We don't need to get a new client here, just clear the Authorization header
		// so that the RoundTripper will add the new token
		req.Header.Del("Authorization")

		// Try again with the new token
		return rt.tryRequest(req, retryCount+1)
	}

	return resp, nil
}

// WithAuthClient adds an authenticated client to the context
func WithAuthClient(ctx context.Context) (context.Context, error) {
	client, err := GetClient(ctx)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, clientKey, client), nil
}

// ClientFromContext gets the authenticated client from the context
func ClientFromContext(ctx context.Context) (*http.Client, bool) {
	client, ok := ctx.Value(clientKey).(*http.Client)
	return client, ok
}

// contextKey is a type for context keys
type contextKey int

const (
	// clientKey is the key for the client in the context
	clientKey contextKey = iota
)
