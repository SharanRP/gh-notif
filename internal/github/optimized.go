package github

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/SharanRP/gh-notif/internal/cache"
	"github.com/google/go-github/v60/github"
)

// OptimizedGetAllNotifications is an optimized version of GetAllNotifications
func (c *Client) OptimizedGetAllNotifications(opts NotificationOptions) ([]*github.Notification, error) {
	cacheKey := fmt.Sprintf("all_notifications_%v_%v_%s_%s_%v_%v_%v_%d",
		opts.All, opts.Unread, opts.RepoName, opts.OrgName,
		opts.Since.Unix(), opts.Before.Unix(), opts.Participating, opts.PerPage)

	// Check cache if enabled
	if opts.UseCache && c.cacheManager != nil {
		if cached, found := c.cacheManager.Manager.Get(cacheKey); found {
			if notifications, ok := cached.([]*github.Notification); ok {
				if c.debug {
					fmt.Printf("Using cached notifications (%d items)\n", len(notifications))
				}

				// Queue background refresh if enabled
				if opts.BackgroundRefresh {
					c.cacheManager.Manager.Prefetch(cache.PrefetchRequest{
						Key:      cacheKey,
						Priority: 1,
						Callback: func(ctx context.Context) (interface{}, error) {
							// This would refresh the notifications in the background
							return c.fetchAllNotificationsWithETag(ctx, opts, cacheKey)
						},
					})
				}

				return notifications, nil
			}
		}
	}

	// Create notification list options
	listOptions := &github.NotificationListOptions{
		All:           opts.All,
		Participating: opts.Participating,
		ListOptions: github.ListOptions{
			PerPage: opts.PerPage,
			Page:    opts.Page,
		},
	}

	// Set since/before if provided
	if !opts.Since.IsZero() {
		listOptions.Since = opts.Since
	}
	if !opts.Before.IsZero() {
		listOptions.Before = opts.Before
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(c.ctx, c.timeout)
	defer cancel()

	// For single page requests, just fetch directly
	if opts.Page > 0 {
		return c.fetchNotificationsWithRetry(ctx, listOptions, opts.RepoName, opts.OrgName)
	}

	// For multi-page requests, use optimized fetching
	return c.fetchAllNotificationsWithETag(ctx, opts, cacheKey)
}

// fetchAllNotificationsWithETag fetches all notifications with ETag support
func (c *Client) fetchAllNotificationsWithETag(ctx context.Context, opts NotificationOptions, cacheKey string) ([]*github.Notification, error) {
	// Create notification list options
	listOptions := &github.NotificationListOptions{
		All:           opts.All,
		Participating: opts.Participating,
		ListOptions: github.ListOptions{
			PerPage: opts.PerPage,
		},
	}

	// Set since/before if provided
	if !opts.Since.IsZero() {
		listOptions.Since = opts.Since
	}
	if !opts.Before.IsZero() {
		listOptions.Before = opts.Before
	}

	// Get ETag from cache metadata if available
	var etag string
	if opts.UseCache && c.cacheManager != nil {
		// In a real implementation, we would store and retrieve the ETag
		// For now, we'll just use a placeholder
		etag = ""
	}

	// Create a request for the first page
	req, err := c.client.NewRequest("GET", "notifications", listOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set If-None-Match header if we have an ETag
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	// Execute the request
	var notifications []*github.Notification
	resp, err := c.client.Do(ctx, req, &notifications)

	// Handle 304 Not Modified
	if resp != nil && resp.StatusCode == http.StatusNotModified {
		// Use cached data
		if c.cacheManager != nil {
			if cached, found := c.cacheManager.Manager.Get(cacheKey); found {
				if cachedNotifications, ok := cached.([]*github.Notification); ok {
					if c.debug {
						fmt.Printf("Using cached notifications (304 Not Modified)\n")
					}
					return cachedNotifications, nil
				}
			}
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// Handle rate limiting
	c.handleRateLimit(resp)
	c.logResponse(resp, notifications, err)

	// Filter by repo or org if needed
	notifications = c.filterNotifications(notifications, opts.RepoName, opts.OrgName)

	// Store the ETag for future requests
	newETag := resp.Header.Get("ETag")
	if newETag != "" {
		// In a real implementation, we would store this ETag
		// For now, we'll just log it
		if c.debug {
			fmt.Printf("New ETag: %s\n", newETag)
		}
	}

	// If there's only one page, return the results
	if resp.NextPage == 0 {
		// Cache the results if enabled
		if opts.UseCache && opts.CacheTTL > 0 && c.cacheManager != nil {
			c.cacheManager.Manager.Set(cacheKey, notifications, opts.CacheTTL)
		}
		return notifications, nil
	}

	// Otherwise, fetch all pages concurrently with optimized batching
	return c.fetchRemainingPagesOptimized(ctx, notifications, resp, listOptions, opts, cacheKey)
}

// fetchRemainingPagesOptimized fetches remaining pages with optimized batching
func (c *Client) fetchRemainingPagesOptimized(
	ctx context.Context,
	firstPageNotifications []*github.Notification,
	firstPageResp *github.Response,
	listOptions *github.NotificationListOptions,
	opts NotificationOptions,
	cacheKey string,
) ([]*github.Notification, error) {
	// Start with the first page results
	allNotifications := make([]*github.Notification, len(firstPageNotifications))
	copy(allNotifications, firstPageNotifications)

	// Determine the last page
	lastPage := firstPageResp.LastPage
	if lastPage == 0 {
		// If GitHub doesn't tell us the last page, estimate it
		// Estimate based on first page and per page
		totalCount := len(firstPageNotifications) * 10 // Assume 10x the first page
		lastPage = (totalCount + opts.PerPage - 1) / opts.PerPage
	}

	// Optimize batch size based on the number of pages
	batchSize := 1
	if lastPage-firstPageResp.NextPage > 10 {
		// For many pages, use larger batches
		batchSize = 5
	}

	// Create a channel for results with appropriate buffer size
	resultCh := make(chan NotificationResult, (lastPage-firstPageResp.NextPage+batchSize-1)/batchSize)

	// Create a semaphore to limit concurrent requests
	maxConcurrent := c.maxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 5
	}
	semaphore := make(chan struct{}, maxConcurrent)

	// Create a wait group to wait for all goroutines
	var wg sync.WaitGroup

	// Fetch remaining pages in batches
	for page := firstPageResp.NextPage; page <= lastPage; page += batchSize {
		wg.Add(1)
		go func(startPage int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Determine end page for this batch
			endPage := startPage + batchSize - 1
			if endPage > lastPage {
				endPage = lastPage
			}

			// Fetch pages in this batch
			batchNotifications, err := c.fetchPageBatch(ctx, listOptions, startPage, endPage, opts.RepoName, opts.OrgName)

			// Send the result to the channel
			resultCh <- NotificationResult{
				Notifications: batchNotifications,
				Error:         err,
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

		// Add the notifications to the result
		allNotifications = append(allNotifications, result.Notifications...)
	}

	// If there was an error, return it
	if fetchErr != nil {
		return allNotifications, fmt.Errorf("error fetching some notifications: %w", fetchErr)
	}

	// Cache the results if enabled
	if opts.UseCache && opts.CacheTTL > 0 && c.cacheManager != nil {
		c.cacheManager.Manager.Set(cacheKey, allNotifications, opts.CacheTTL)
	}

	return allNotifications, nil
}

// fetchPageBatch fetches a batch of pages
func (c *Client) fetchPageBatch(
	ctx context.Context,
	baseOptions *github.NotificationListOptions,
	startPage, endPage int,
	repoFilter, orgFilter string,
) ([]*github.Notification, error) {
	var allNotifications []*github.Notification

	// Fetch each page in the batch
	for page := startPage; page <= endPage; page++ {
		// Create a copy of the options with the current page
		pageOpts := *baseOptions
		pageOpts.Page = page

		// Fetch the page
		notifications, resp, err := c.client.Activity.ListNotifications(ctx, &pageOpts)
		if err != nil {
			return allNotifications, err
		}

		// Handle rate limiting
		c.handleRateLimit(resp)
		c.logResponse(resp, notifications, nil)

		// Filter and add to results
		filtered := c.filterNotifications(notifications, repoFilter, orgFilter)
		allNotifications = append(allNotifications, filtered...)
	}

	return allNotifications, nil
}
