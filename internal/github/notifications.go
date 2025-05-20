package github

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/user/gh-notif/internal/ui"
)

// NotificationOptions contains options for filtering notifications
type NotificationOptions struct {
	All           bool      // Include all notifications, not just unread ones
	Unread        bool      // Only include unread notifications
	RepoName      string    // Filter by repository name
	OrgName       string    // Filter by organization name
	Since         time.Time // Only show notifications updated after this time
	Before        time.Time // Only show notifications updated before this time
	Participating bool      // Only show notifications in which the user is participating or mentioned
	PerPage       int       // Number of results per page
	Page          int       // Page number
	UseCache      bool      // Whether to use cached results if available
	CacheTTL      time.Duration // How long to cache results
	MaxConcurrent int       // Maximum number of concurrent requests
}

// ListNotifications fetches and displays GitHub notifications
func ListNotifications(options NotificationOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create an authenticated client with middleware
	client, err := NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Set default values if not specified
	if options.PerPage <= 0 {
		options.PerPage = 100
	}

	// Enable caching by default
	options.UseCache = true
	if options.CacheTTL <= 0 {
		options.CacheTTL = 5 * time.Minute
	}

	// Set max concurrent requests
	if options.MaxConcurrent <= 0 {
		options.MaxConcurrent = 5
	}

	// Fetch notifications using the high-performance implementation
	var notifications []*github.Notification

	// Choose the appropriate method based on the options
	if options.RepoName != "" {
		notifications, err = client.GetNotificationsByRepo(options.RepoName, options)
	} else if options.OrgName != "" {
		notifications, err = client.GetNotificationsByOrg(options.OrgName, options)
	} else if !options.All {
		notifications, err = client.GetUnreadNotifications(options)
	} else {
		notifications, err = client.GetAllNotifications(options)
	}

	if err != nil {
		return fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// Fetch additional details for the notifications
	if len(notifications) > 0 {
		if err := client.FetchNotificationDetails(notifications); err != nil {
			// Log the error but continue
			fmt.Printf("Warning: Failed to fetch some notification details: %v\n", err)
		}
	}

	// Use the UI package to display notifications
	return ui.DisplayNotifications(notifications)
}

// MarkAsRead marks a notification as read
func MarkAsRead(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create an authenticated client with middleware
	client, err := NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Mark the thread as read
	return client.MarkNotificationRead(id)
}

// MarkNotificationRead marks a notification thread as read
func (c *Client) MarkNotificationRead(threadID string) error {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return err
	}

	// Log the request
	c.logRequest("PATCH", fmt.Sprintf("notifications/threads/%s", threadID), nil)

	// Mark the thread as read
	resp, err := c.client.Activity.MarkThreadRead(c.ctx, threadID)

	// Log the response
	c.logResponse(resp, nil, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to mark thread as read: %w", err)
	}

	return nil
}

// MarkAllNotificationsRead marks all notifications as read
func (c *Client) MarkAllNotificationsRead() error {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return err
	}

	// Log the request
	c.logRequest("PUT", "notifications", nil)

	// Mark all notifications as read
	resp, err := c.client.Activity.MarkNotificationsRead(c.ctx, github.Timestamp{Time: time.Now()})

	// Log the response
	c.logResponse(resp, nil, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// MarkRepositoryNotificationsRead marks all notifications in a repository as read
func (c *Client) MarkRepositoryNotificationsRead(owner, repo string) error {
	// Wait for rate limiter
	if err := c.waitForRateLimit(c.ctx); err != nil {
		return err
	}

	// Log the request
	c.logRequest("PUT", fmt.Sprintf("repos/%s/%s/notifications", owner, repo), nil)

	// Mark all notifications in the repository as read
	resp, err := c.client.Activity.MarkRepositoryNotificationsRead(c.ctx, owner, repo, github.Timestamp{Time: time.Now()})

	// Log the response
	c.logResponse(resp, nil, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to mark repository notifications as read: %w", err)
	}

	return nil
}

// NotificationResult represents the result of a notification fetch operation
type NotificationResult struct {
	Notifications []*github.Notification
	Response      *github.Response
	Error         error
}

// GetAllNotifications fetches all notifications with pagination support
func (c *Client) GetAllNotifications(opts NotificationOptions) ([]*github.Notification, error) {
	cacheKey := fmt.Sprintf("all_notifications_%v_%v_%s_%s_%v_%v_%v_%d",
		opts.All, opts.Unread, opts.RepoName, opts.OrgName,
		opts.Since.Unix(), opts.Before.Unix(), opts.Participating, opts.PerPage)

	// Check cache if enabled
	if opts.UseCache {
		if cached, found := c.cache.Get(cacheKey); found {
			if notifications, ok := cached.([]*github.Notification); ok {
				if c.debug {
					fmt.Printf("Using cached notifications (%d items)\n", len(notifications))
				}
				return notifications, nil
			}
		}
	}

	// Set up the list options
	listOptions := &github.NotificationListOptions{
		All:           opts.All,
		Participating: opts.Participating,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if !opts.Since.IsZero() {
		listOptions.Since = opts.Since
	}

	if !opts.Before.IsZero() {
		listOptions.Before = opts.Before
	}

	// Override per page if specified
	if opts.PerPage > 0 {
		listOptions.ListOptions.PerPage = opts.PerPage
	}

	// Set page if specified
	if opts.Page > 0 {
		listOptions.ListOptions.Page = opts.Page
	}

	// Determine max concurrent requests
	maxConcurrent := 5 // Default
	if opts.MaxConcurrent > 0 {
		maxConcurrent = opts.MaxConcurrent
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	// For single page requests, just fetch directly
	if opts.Page > 0 {
		return c.fetchNotificationsWithRetry(ctx, listOptions, opts.RepoName, opts.OrgName)
	}

	// For multi-page requests, fetch the first page to get pagination info
	notifications, resp, err := c.client.Activity.ListNotifications(ctx, listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// Handle rate limiting
	c.handleRateLimit(resp)
	c.logResponse(resp, notifications, err)

	// Filter by repo or org if needed
	notifications = c.filterNotifications(notifications, opts.RepoName, opts.OrgName)

	// If there's only one page, return the results
	if resp.NextPage == 0 {
		// Cache the results if enabled
		if opts.UseCache && opts.CacheTTL > 0 {
			c.cache.Set(cacheKey, notifications, opts.CacheTTL)
		}
		return notifications, nil
	}

	// Otherwise, fetch all pages concurrently
	var allNotifications []*github.Notification
	allNotifications = append(allNotifications, notifications...)

	// Calculate the number of remaining pages
	lastPage := resp.LastPage
	if lastPage == 0 {
		// If GitHub doesn't provide LastPage, estimate based on the number of notifications
		// GitHub API doesn't provide a total count, so we'll estimate
		estimatedTotal := len(notifications) * 2 // Assume there are at least twice as many
		lastPage = (estimatedTotal + listOptions.ListOptions.PerPage - 1) / listOptions.ListOptions.PerPage
	}

	// Create a channel for results
	resultCh := make(chan NotificationResult, lastPage-1)

	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup

	// Fetch remaining pages concurrently
	for page := resp.NextPage; page <= lastPage; page++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create a copy of the list options with the current page
			pageOpts := *listOptions
			pageOpts.ListOptions.Page = pageNum

			// Fetch the page
			pageNotifications, pageResp, pageErr := c.client.Activity.ListNotifications(ctx, &pageOpts)

			// Send the result to the channel
			resultCh <- NotificationResult{
				Notifications: c.filterNotifications(pageNotifications, opts.RepoName, opts.OrgName),
				Response:      pageResp,
				Error:         pageErr,
			}
		}(page)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultCh)

	// Process the results
	var fetchErr error
	for result := range resultCh {
		if result.Error != nil {
			fetchErr = result.Error
			continue
		}

		// Handle rate limiting
		c.handleRateLimit(result.Response)
		c.logResponse(result.Response, result.Notifications, result.Error)

		// Add the notifications to the result
		allNotifications = append(allNotifications, result.Notifications...)
	}

	// If there was an error, return it
	if fetchErr != nil {
		return allNotifications, fmt.Errorf("error fetching some notifications: %w", fetchErr)
	}

	// Cache the results if enabled
	if opts.UseCache && opts.CacheTTL > 0 {
		c.cache.Set(cacheKey, allNotifications, opts.CacheTTL)
	}

	return allNotifications, nil
}

// GetUnreadNotifications fetches only unread notifications
func (c *Client) GetUnreadNotifications(opts NotificationOptions) ([]*github.Notification, error) {
	// Force unread to true
	opts.Unread = true
	opts.All = false

	return c.GetAllNotifications(opts)
}

// GetNotificationsByRepo fetches notifications for a specific repository
func (c *Client) GetNotificationsByRepo(repo string, opts NotificationOptions) ([]*github.Notification, error) {
	// Set the repo name
	opts.RepoName = repo

	return c.GetAllNotifications(opts)
}

// GetNotificationsByOrg fetches notifications for a specific organization
func (c *Client) GetNotificationsByOrg(org string, opts NotificationOptions) ([]*github.Notification, error) {
	// Set the org name
	opts.OrgName = org

	return c.GetAllNotifications(opts)
}

// fetchNotificationsWithRetry fetches notifications with retry logic
func (c *Client) fetchNotificationsWithRetry(ctx context.Context, opts *github.NotificationListOptions, repoFilter, orgFilter string) ([]*github.Notification, error) {
	var notifications []*github.Notification
	var resp *github.Response
	var err error

	// Wait for rate limiter
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	// Log the request
	c.logRequest("GET", "notifications", opts)

	// Try the request with retries
	for attempt := 0; attempt <= c.retryCount; attempt++ {
		notifications, resp, err = c.client.Activity.ListNotifications(ctx, opts)

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
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications after %d attempts: %w", c.retryCount+1, err)
	}

	// Filter the notifications if needed
	return c.filterNotifications(notifications, repoFilter, orgFilter), nil
}

// filterNotifications filters notifications by repository and organization
func (c *Client) filterNotifications(notifications []*github.Notification, repoFilter, orgFilter string) []*github.Notification {
	if repoFilter == "" && orgFilter == "" {
		return notifications
	}

	var filtered []*github.Notification
	for _, n := range notifications {
		// Filter by repository if specified
		if repoFilter != "" && n.GetRepository().GetFullName() != repoFilter {
			continue
		}

		// Filter by organization if specified
		if orgFilter != "" {
			repoFullName := n.GetRepository().GetFullName()
			parts := strings.Split(repoFullName, "/")
			if len(parts) > 0 && parts[0] != orgFilter {
				continue
			}
		}

		filtered = append(filtered, n)
	}

	return filtered
}

// FetchNotificationDetails fetches additional details for notifications in parallel
func (c *Client) FetchNotificationDetails(notifications []*github.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	// Determine max concurrent requests
	maxConcurrent := c.maxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, len(notifications))

	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)

	for _, notification := range notifications {
		wg.Add(1)
		go func(n *github.Notification) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Wait for rate limiter
			if err := c.waitForRateLimit(ctx); err != nil {
				errCh <- err
				return
			}

			// Fetch additional details based on notification type
			subjectType := n.GetSubject().GetType()
			subjectURL := n.GetSubject().GetURL()

			if subjectURL == "" {
				return // Skip if no URL
			}

			// Log the request
			c.logRequest("GET", subjectURL, nil)

			var err error
			switch subjectType {
			case "Issue":
				err = c.fetchIssueDetails(ctx, n)
			case "PullRequest":
				err = c.fetchPullRequestDetails(ctx, n)
			case "Commit":
				err = c.fetchCommitDetails(ctx, n)
			case "Release":
				err = c.fetchReleaseDetails(ctx, n)
			case "Discussion":
				err = c.fetchDiscussionDetails(ctx, n)
			}

			if err != nil {
				errCh <- fmt.Errorf("error fetching details for %s: %w", subjectType, err)
			}

		}(notification)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Check if there were any errors
	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors fetching notification details: %v", errs)
	}

	return nil
}

// fetchIssueDetails fetches details for an issue notification
func (c *Client) fetchIssueDetails(ctx context.Context, notification *github.Notification) error {
	// Parse the URL to get owner, repo, and issue number
	url := notification.GetSubject().GetURL()
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid issue URL: %s", url)
	}

	owner := parts[4]
	repo := parts[5]
	issueNumber := parts[7]

	// Convert issue number to int
	var issueNum int
	if _, err := fmt.Sscanf(issueNumber, "%d", &issueNum); err != nil {
		return fmt.Errorf("invalid issue number: %s", issueNumber)
	}

	// Fetch issue details
	issue, resp, err := c.client.Issues.Get(ctx, owner, repo, issueNum)

	// Log the response
	c.logResponse(resp, issue, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to fetch issue details: %w", err)
	}

	// Store the details in the notification (using the latest_comment_url field)
	notification.Subject.LatestCommentURL = github.String(fmt.Sprintf("%s#%d", url, issue.GetID()))

	return nil
}

// fetchPullRequestDetails fetches details for a pull request notification
func (c *Client) fetchPullRequestDetails(ctx context.Context, notification *github.Notification) error {
	// Parse the URL to get owner, repo, and PR number
	url := notification.GetSubject().GetURL()
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid pull request URL: %s", url)
	}

	owner := parts[4]
	repo := parts[5]
	prNumber := parts[7]

	// Convert PR number to int
	var prNum int
	if _, err := fmt.Sscanf(prNumber, "%d", &prNum); err != nil {
		return fmt.Errorf("invalid pull request number: %s", prNumber)
	}

	// Fetch PR details
	pr, resp, err := c.client.PullRequests.Get(ctx, owner, repo, prNum)

	// Log the response
	c.logResponse(resp, pr, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to fetch pull request details: %w", err)
	}

	// Store the details in the notification
	notification.Subject.LatestCommentURL = github.String(fmt.Sprintf("%s#%d", url, pr.GetID()))

	return nil
}

// fetchCommitDetails fetches details for a commit notification
func (c *Client) fetchCommitDetails(ctx context.Context, notification *github.Notification) error {
	// Parse the URL to get owner, repo, and commit SHA
	url := notification.GetSubject().GetURL()
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid commit URL: %s", url)
	}

	owner := parts[4]
	repo := parts[5]
	sha := parts[7]

	// Fetch commit details
	commit, resp, err := c.client.Repositories.GetCommit(ctx, owner, repo, sha, nil)

	// Log the response
	c.logResponse(resp, commit, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to fetch commit details: %w", err)
	}

	// Store the details in the notification
	notification.Subject.LatestCommentURL = github.String(fmt.Sprintf("%s#%s", url, commit.GetSHA()))

	return nil
}

// fetchReleaseDetails fetches details for a release notification
func (c *Client) fetchReleaseDetails(ctx context.Context, notification *github.Notification) error {
	// Parse the URL to get owner, repo, and release ID
	url := notification.GetSubject().GetURL()
	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid release URL: %s", url)
	}

	owner := parts[4]
	repo := parts[5]
	releaseID := parts[7]

	// Convert release ID to int64
	var id int64
	if _, err := fmt.Sscanf(releaseID, "%d", &id); err != nil {
		return fmt.Errorf("invalid release ID: %s", releaseID)
	}

	// Fetch release details
	release, resp, err := c.client.Repositories.GetRelease(ctx, owner, repo, id)

	// Log the response
	c.logResponse(resp, release, err)

	// Handle rate limiting
	c.handleRateLimit(resp)

	if err != nil {
		return fmt.Errorf("failed to fetch release details: %w", err)
	}

	// Store the details in the notification
	notification.Subject.LatestCommentURL = github.String(fmt.Sprintf("%s#%d", url, release.GetID()))

	return nil
}

// fetchDiscussionDetails fetches details for a discussion notification
func (c *Client) fetchDiscussionDetails(ctx context.Context, notification *github.Notification) error {
	// GitHub API doesn't have a direct endpoint for discussions yet
	// This is a placeholder for when it becomes available
	return nil
}
