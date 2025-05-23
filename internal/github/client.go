package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/auth"
	"github.com/SharanRP/gh-notif/internal/config"
	"github.com/google/go-github/v60/github"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

// Client wraps the GitHub client with authentication and additional features
type Client struct {
	client        *github.Client
	ctx           context.Context
	rateLimiter   *rate.Limiter
	retryClient   *retryablehttp.Client
	cacheManager  *CacheManager
	configManager *config.ConfigManager
	baseURL       string
	uploadURL     string
	maxConcurrent int
	retryCount    int
	retryDelay    time.Duration
	timeout       time.Duration
	cacheTTL      time.Duration
	debug         bool

	// Object pools for memory efficiency
	notificationPool sync.Pool
	responsePool     sync.Pool
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the GitHub API
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithUploadURL sets the upload URL for the GitHub API
func WithUploadURL(uploadURL string) ClientOption {
	return func(c *Client) {
		c.uploadURL = uploadURL
	}
}

// WithMaxConcurrent sets the maximum number of concurrent requests
func WithMaxConcurrent(maxConcurrent int) ClientOption {
	return func(c *Client) {
		c.maxConcurrent = maxConcurrent
	}
}

// WithRetryCount sets the number of retries for failed requests
func WithRetryCount(retryCount int) ClientOption {
	return func(c *Client) {
		c.retryCount = retryCount
	}
}

// WithRetryDelay sets the delay between retries
func WithRetryDelay(retryDelay time.Duration) ClientOption {
	return func(c *Client) {
		c.retryDelay = retryDelay
	}
}

// WithTimeout sets the timeout for API requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithCacheTTL sets the TTL for cached responses
func WithCacheTTL(cacheTTL time.Duration) ClientOption {
	return func(c *Client) {
		c.cacheTTL = cacheTTL
	}
}

// WithDebug enables or disables debug logging
func WithDebug(debug bool) ClientOption {
	return func(c *Client) {
		c.debug = debug
	}
}

// NewClient creates a new authenticated GitHub client with enhanced features
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	// Load configuration
	cm := config.NewConfigManager()
	if err := cm.Load(); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get the config
	config := cm.GetConfig()

	// Create a client with default settings
	client := &Client{
		ctx:           ctx,
		configManager: cm,
		baseURL:       config.API.BaseURL,
		uploadURL:     config.API.UploadURL,
		maxConcurrent: config.Advanced.MaxConcurrent,
		retryCount:    config.API.RetryCount,
		retryDelay:    time.Duration(config.API.RetryDelay) * time.Second,
		timeout:       time.Duration(config.API.Timeout) * time.Second,
		cacheTTL:      time.Duration(config.Advanced.CacheTTL) * time.Second,
		debug:         config.Advanced.Debug,

		// Initialize object pools
		notificationPool: sync.Pool{
			New: func() interface{} {
				return &github.Notification{}
			},
		},
		responsePool: sync.Pool{
			New: func() interface{} {
				return &github.Response{}
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Create a rate limiter (default: 5000 requests per hour = ~1.4 requests per second)
	client.rateLimiter = rate.NewLimiter(rate.Limit(1.4), 5)

	// Create a retryable HTTP client
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = client.retryCount
	retryClient.RetryWaitMin = client.retryDelay
	retryClient.RetryWaitMax = client.retryDelay * 10
	retryClient.Logger = nil // Disable default logging
	client.retryClient = retryClient

	// Get an authenticated HTTP client
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Add the auth middleware
	middleware := auth.NewAuthMiddleware()
	httpClient.Transport = middleware.RoundTripper(httpClient.Transport)

	// Set the timeout
	httpClient.Timeout = client.timeout

	// Create the GitHub client
	ghClient := github.NewClient(httpClient)

	// Set custom base URL if provided (for GitHub Enterprise)
	if client.baseURL != "https://api.github.com" {
		baseURL, err := url.Parse(client.baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}

		uploadURL, err := url.Parse(client.uploadURL)
		if err != nil {
			return nil, fmt.Errorf("invalid upload URL: %w", err)
		}

		ghClient.BaseURL = baseURL
		ghClient.UploadURL = uploadURL
	}

	client.client = ghClient

	// Initialize the cache manager
	cacheManager, err := NewCacheManager(client, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %w", err)
	}
	client.cacheManager = cacheManager

	return client, nil
}

// NewClientOrExit creates a new authenticated GitHub client or exits on error
func NewClientOrExit(ctx context.Context, opts ...ClientOption) *Client {
	client, err := NewClient(ctx, opts...)
	if err != nil {
		fmt.Printf("Error creating GitHub client: %v\n", err)
		os.Exit(1)
	}
	return client
}

// waitForRateLimit waits for the rate limiter
func (c *Client) waitForRateLimit(ctx context.Context) error {
	// Wait for rate limiter
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit wait error: %w", err)
	}
	return nil
}

// handleRateLimit handles GitHub API rate limiting
func (c *Client) handleRateLimit(resp *github.Response) {
	if resp != nil && resp.Rate.Remaining == 0 {
		// Calculate how long to wait
		waitTime := resp.Rate.Reset.Time.Sub(time.Now())
		if waitTime > 0 {
			if c.debug {
				fmt.Printf("Rate limit exceeded, waiting for %v\n", waitTime)
			}
			time.Sleep(waitTime)
		}
	}
}

// logRequest logs API request details if debug is enabled
func (c *Client) logRequest(method, url string, body interface{}) {
	if c.debug {
		fmt.Printf("GitHub API Request: %s %s\n", method, url)
		if body != nil {
			fmt.Printf("Request Body: %+v\n", body)
		}
	}
}

// logResponse logs API response details if debug is enabled
func (c *Client) logResponse(resp *github.Response, body interface{}, err error) {
	if c.debug {
		if err != nil {
			fmt.Printf("GitHub API Error: %v\n", err)
		}
		if resp != nil {
			fmt.Printf("GitHub API Response: %d %s\n", resp.StatusCode, resp.Status)
			fmt.Printf("Rate Limit: %d/%d, Reset: %s\n",
				resp.Rate.Remaining, resp.Rate.Limit, resp.Rate.Reset.Time)
		}
		if body != nil {
			fmt.Printf("Response Body: %+v\n", body)
		}
	}
}

// isRateLimitError checks if an error is due to rate limiting
func (c *Client) isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	var rateLimitErr *github.RateLimitError
	return errors.As(err, &rateLimitErr)
}

// isTransientError checks if an error is transient and can be retried
func (c *Client) isTransientError(err error) bool {
	if err == nil {
		return false
	}

	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return true
	}

	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}

	var transportErr *url.Error
	return errors.As(err, &transportErr)
}

// ListNotifications lists GitHub notifications with rate limiting and retries
func (c *Client) ListNotifications(opts *github.NotificationListOptions) ([]*github.Notification, *github.Response, error) {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return nil, nil, err
	}

	// Log the request
	c.logRequest("GET", "notifications", opts)

	// Try the request with retries
	var notifications []*github.Notification
	var resp *github.Response
	var err error

	for attempt := 0; attempt <= c.retryCount; attempt++ {
		notifications, resp, err = c.client.Activity.ListNotifications(c.ctx, opts)

		// Log the response
		c.logResponse(resp, notifications, err)

		// If successful or not a transient error, break
		if err == nil || !c.isTransientError(err) {
			break
		}

		// If this is a rate limit error, wait for the reset time
		if c.isRateLimitError(err) {
			c.handleRateLimit(resp)
			continue
		}

		// Otherwise, wait and retry
		if attempt < c.retryCount {
			waitTime := c.retryDelay * time.Duration(1<<uint(attempt))
			if c.debug {
				fmt.Printf("Retrying after %v (attempt %d/%d)\n", waitTime, attempt+1, c.retryCount)
			}
			select {
			case <-time.After(waitTime):
				// Continue with retry
			case <-c.ctx.Done():
				return nil, nil, c.ctx.Err()
			}
		}
	}

	// Handle rate limiting for successful requests
	c.handleRateLimit(resp)

	return notifications, resp, err
}

// MarkThreadRead marks a notification thread as read
func (c *Client) MarkThreadRead(threadID string) (*github.Response, error) {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return nil, err
	}

	// Log the request
	c.logRequest("PATCH", fmt.Sprintf("notifications/threads/%s", threadID), nil)

	// Mark the thread as read
	resp, err := c.client.Activity.MarkThreadRead(c.ctx, threadID)

	// Log the response
	c.logResponse(resp, nil, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	return resp, err
}

// GetThread gets a notification thread
func (c *Client) GetThread(threadID string) (*github.Notification, *github.Response, error) {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return nil, nil, err
	}

	// Log the request
	c.logRequest("GET", fmt.Sprintf("notifications/threads/%s", threadID), nil)

	// Get the thread
	notification, resp, err := c.client.Activity.GetThread(c.ctx, threadID)

	// Log the response
	c.logResponse(resp, notification, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	return notification, resp, err
}

// WithContext returns a new Client with the given context
func (c *Client) WithContext(ctx context.Context) *Client {
	newClient := &Client{
		client:        c.client,
		ctx:           ctx,
		rateLimiter:   c.rateLimiter,
		retryClient:   c.retryClient,
		cacheManager:  c.cacheManager,
		configManager: c.configManager,
		baseURL:       c.baseURL,
		uploadURL:     c.uploadURL,
		maxConcurrent: c.maxConcurrent,
		retryCount:    c.retryCount,
		retryDelay:    c.retryDelay,
		timeout:       c.timeout,
		cacheTTL:      c.cacheTTL,
		debug:         c.debug,
		// Initialize new object pools to avoid copying sync.Pool
		notificationPool: sync.Pool{
			New: func() interface{} {
				return &github.Notification{}
			},
		},
		responsePool: sync.Pool{
			New: func() interface{} {
				return &github.Response{}
			},
		},
	}
	return newClient
}

// Do performs an HTTP request and returns the API response
func (c *Client) Do(req *http.Request, v interface{}) (*github.Response, error) {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return nil, err
	}

	// Log the request
	c.logRequest(req.Method, req.URL.String(), nil)

	// Perform the request
	resp, err := c.client.Do(c.ctx, req, v)

	// Log the response
	c.logResponse(resp, v, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	return resp, err
}

// GetRawClient returns the underlying GitHub client
func (c *Client) GetRawClient() *github.Client {
	return c.client
}

// SetRawClient sets the underlying GitHub client
func (c *Client) SetRawClient(client *github.Client) {
	c.client = client
}
