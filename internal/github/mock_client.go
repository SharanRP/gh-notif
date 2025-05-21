package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/time/rate"
)

// NewMockClient creates a new client for testing
func NewMockClient(ctx context.Context) (*Client, *httptest.Server, error) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default response for any request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))

	// Create a client with default settings
	client := &Client{
		ctx:           ctx,
		baseURL:       server.URL,
		uploadURL:     server.URL,
		maxConcurrent: 5,
		retryCount:    3,
		retryDelay:    1 * time.Second,
		timeout:       30 * time.Second,
		cacheTTL:      5 * time.Minute,
		debug:         false,
		rateLimiter:   nil, // Will be set below
	}

	// Create a rate limiter with a high limit for tests
	client.rateLimiter = rate.NewLimiter(rate.Inf, 100) // Infinite rate for tests

	// Create a standard HTTP client
	httpClient := &http.Client{
		Timeout: client.timeout,
	}

	// Create the GitHub client
	ghClient := github.NewClient(httpClient)

	// Set the base URL to the mock server
	baseURL, _ := ghClient.BaseURL.Parse(server.URL + "/")
	uploadURL, _ := ghClient.UploadURL.Parse(server.URL + "/")
	ghClient.BaseURL = baseURL
	ghClient.UploadURL = uploadURL

	client.client = ghClient

	return client, server, nil
}

// NewTestClient creates a new client for testing with a custom handler
func NewTestClient(ctx context.Context, handler http.Handler) (*Client, *httptest.Server, error) {
	// Create a mock server with the provided handler
	server := httptest.NewServer(handler)

	// Create a client with default settings
	client := &Client{
		ctx:           ctx,
		baseURL:       server.URL,
		uploadURL:     server.URL,
		maxConcurrent: 5,
		retryCount:    3,
		retryDelay:    1 * time.Millisecond, // Use a short delay for testing
		timeout:       30 * time.Second,
		cacheTTL:      5 * time.Minute,
		debug:         false,
		rateLimiter:   nil, // Will be set below
	}

	// Create a rate limiter with a high limit for tests
	client.rateLimiter = rate.NewLimiter(rate.Inf, 100) // Infinite rate for tests

	// Create a standard HTTP client
	httpClient := &http.Client{
		Timeout: client.timeout,
	}

	// Create the GitHub client
	ghClient := github.NewClient(httpClient)

	// Set the base URL to the mock server
	baseURL, _ := ghClient.BaseURL.Parse(server.URL + "/")
	uploadURL, _ := ghClient.UploadURL.Parse(server.URL + "/")
	ghClient.BaseURL = baseURL
	ghClient.UploadURL = uploadURL

	client.client = ghClient

	return client, server, nil
}
