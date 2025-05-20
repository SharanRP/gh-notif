package github

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/go-github/v60/github"
)

// NotificationStream provides a streaming interface for notifications
type NotificationStream struct {
	client        *Client
	options       NotificationOptions
	ctx           context.Context
	cancel        context.CancelFunc
	notificationCh chan *github.Notification
	errorCh       chan error
	doneCh        chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	started       bool
}

// NewNotificationStream creates a new notification stream
func NewNotificationStream(client *Client, options NotificationOptions) *NotificationStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationStream{
		client:        client,
		options:       options,
		ctx:           ctx,
		cancel:        cancel,
		notificationCh: make(chan *github.Notification, 100),
		errorCh:       make(chan error, 10),
		doneCh:        make(chan struct{}),
		started:       false,
	}
}

// Start starts the notification stream
func (s *NotificationStream) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("stream already started")
	}

	s.started = true
	go s.stream()
	return nil
}

// Stop stops the notification stream
func (s *NotificationStream) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	s.cancel()
	<-s.doneCh
	s.started = false
}

// Notifications returns a channel of notifications
func (s *NotificationStream) Notifications() <-chan *github.Notification {
	return s.notificationCh
}

// Errors returns a channel of errors
func (s *NotificationStream) Errors() <-chan error {
	return s.errorCh
}

// stream fetches notifications and streams them to the channel
func (s *NotificationStream) stream() {
	defer close(s.doneCh)
	defer close(s.notificationCh)
	defer close(s.errorCh)

	// Set up the list options
	listOptions := &github.NotificationListOptions{
		All:           s.options.All,
		Participating: s.options.Participating,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	if !s.options.Since.IsZero() {
		listOptions.Since = s.options.Since
	}

	if !s.options.Before.IsZero() {
		listOptions.Before = s.options.Before
	}

	// Override per page if specified
	if s.options.PerPage > 0 {
		listOptions.ListOptions.PerPage = s.options.PerPage
	}

	// Determine max concurrent requests
	maxConcurrent := 5 // Default
	if s.options.MaxConcurrent > 0 {
		maxConcurrent = s.options.MaxConcurrent
	}

	// Fetch the first page to get pagination info
	notifications, resp, err := s.client.client.Activity.ListNotifications(s.ctx, listOptions)
	if err != nil {
		s.errorCh <- fmt.Errorf("failed to fetch notifications: %w", err)
		return
	}

	// Handle rate limiting
	s.client.handleRateLimit(resp)
	s.client.logResponse(resp, notifications, err)

	// Stream the first page of notifications
	for _, n := range s.client.filterNotifications(notifications, s.options.RepoName, s.options.OrgName) {
		select {
		case s.notificationCh <- n:
		case <-s.ctx.Done():
			return
		}
	}

	// If there's only one page, we're done
	if resp.NextPage == 0 {
		return
	}

	// Calculate the number of remaining pages
	lastPage := resp.LastPage
	if lastPage == 0 {
		// If GitHub doesn't provide LastPage, estimate based on the number of notifications
		// GitHub API doesn't provide a total count, so we'll estimate
		estimatedTotal := len(notifications) * 2 // Assume there are at least twice as many
		lastPage = (estimatedTotal + listOptions.ListOptions.PerPage - 1) / listOptions.ListOptions.PerPage
	}

	// Use a semaphore to limit concurrent requests
	semaphore := make(chan struct{}, maxConcurrent)

	// Fetch remaining pages concurrently
	for page := resp.NextPage; page <= lastPage; page++ {
		// Check if the context is done
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		s.wg.Add(1)
		go func(pageNum int) {
			defer s.wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create a copy of the list options with the current page
			pageOpts := *listOptions
			pageOpts.ListOptions.Page = pageNum

			// Wait for rate limiter
			if err := s.client.waitForRateLimit(s.ctx); err != nil {
				s.errorCh <- err
				return
			}

			// Fetch the page
			pageNotifications, pageResp, pageErr := s.client.client.Activity.ListNotifications(s.ctx, &pageOpts)
			if pageErr != nil {
				s.errorCh <- fmt.Errorf("failed to fetch page %d: %w", pageNum, pageErr)
				return
			}

			// Handle rate limiting
			s.client.handleRateLimit(pageResp)
			s.client.logResponse(pageResp, pageNotifications, pageErr)

			// Filter and stream the notifications
			filtered := s.client.filterNotifications(pageNotifications, s.options.RepoName, s.options.OrgName)
			for _, n := range filtered {
				select {
				case s.notificationCh <- n:
				case <-s.ctx.Done():
					return
				}
			}
		}(page)
	}

	// Wait for all goroutines to complete
	s.wg.Wait()
}

// CollectAll collects all notifications from the stream and returns them as a slice
func (s *NotificationStream) CollectAll() ([]*github.Notification, error) {
	var notifications []*github.Notification
	var errs []error

	// Start the stream if not already started
	if err := s.Start(); err != nil {
		return nil, err
	}

	// Collect notifications and errors
	for {
		select {
		case n, ok := <-s.Notifications():
			if !ok {
				// Channel closed, we're done
				if len(errs) > 0 {
					return notifications, fmt.Errorf("errors collecting notifications: %v", errs)
				}
				return notifications, nil
			}
			notifications = append(notifications, n)
		case err, ok := <-s.Errors():
			if !ok {
				// Channel closed
				continue
			}
			errs = append(errs, err)
		case <-s.ctx.Done():
			return notifications, fmt.Errorf("context canceled: %w", s.ctx.Err())
		}
	}
}
