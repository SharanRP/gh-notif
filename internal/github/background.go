package github

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

// BackgroundRefresher manages background refresh of notifications
type BackgroundRefresher struct {
	client        *Client
	interval      time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
	notifications []*github.Notification
	mu            sync.RWMutex
	options       NotificationOptions
	running       bool
	onUpdate      func([]*github.Notification)
	lastError     error
	errorMu       sync.RWMutex
}

// NewBackgroundRefresher creates a new background refresher
func NewBackgroundRefresher(client *Client, interval time.Duration, options NotificationOptions) *BackgroundRefresher {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundRefresher{
		client:   client,
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		options:  options,
		running:  false,
	}
}

// Start starts the background refresh
func (b *BackgroundRefresher) Start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return
	}

	b.running = true
	go b.refreshLoop()
}

// Stop stops the background refresh
func (b *BackgroundRefresher) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return
	}

	b.cancel()
	b.running = false
}

// GetNotifications returns the current notifications
func (b *BackgroundRefresher) GetNotifications() []*github.Notification {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*github.Notification, len(b.notifications))
	copy(result, b.notifications)
	return result
}

// SetOnUpdateCallback sets a callback function to be called when notifications are updated
func (b *BackgroundRefresher) SetOnUpdateCallback(callback func([]*github.Notification)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.onUpdate = callback
}

// GetLastError returns the last error encountered during refresh
func (b *BackgroundRefresher) GetLastError() error {
	b.errorMu.RLock()
	defer b.errorMu.RUnlock()

	return b.lastError
}

// IsRunning returns whether the background refresh is running
func (b *BackgroundRefresher) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.running
}

// refreshLoop is the main loop for background refresh
func (b *BackgroundRefresher) refreshLoop() {
	// Do an initial refresh
	b.refresh()

	// Set up a ticker for periodic refresh
	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.refresh()
		case <-b.ctx.Done():
			return
		}
	}
}

// refresh fetches notifications and updates the internal state
func (b *BackgroundRefresher) refresh() {
	var notifications []*github.Notification
	var err error

	// Choose the appropriate method based on the options
	if b.options.RepoName != "" {
		notifications, err = b.client.GetNotificationsByRepo(b.options.RepoName, b.options)
	} else if b.options.OrgName != "" {
		notifications, err = b.client.GetNotificationsByOrg(b.options.OrgName, b.options)
	} else if !b.options.All {
		notifications, err = b.client.GetUnreadNotifications(b.options)
	} else {
		notifications, err = b.client.GetAllNotifications(b.options)
	}

	// Update the last error
	b.errorMu.Lock()
	b.lastError = err
	b.errorMu.Unlock()

	if err != nil {
		fmt.Printf("Error refreshing notifications: %v\n", err)
		return
	}

	// Fetch additional details for the notifications
	if len(notifications) > 0 {
		if err := b.client.FetchNotificationDetails(notifications); err != nil {
			// Log the error but continue
			fmt.Printf("Warning: Failed to fetch some notification details: %v\n", err)
		}
	}

	// Update the notifications
	b.mu.Lock()
	b.notifications = notifications

	// Call the callback if set
	onUpdate := b.onUpdate
	b.mu.Unlock()

	// Call the callback outside the lock
	if onUpdate != nil {
		onUpdate(notifications)
	}
}

// ForceRefresh forces an immediate refresh
func (b *BackgroundRefresher) ForceRefresh() {
	go b.refresh()
}

// SetOptions updates the notification options
func (b *BackgroundRefresher) SetOptions(options NotificationOptions) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.options = options
}

// SetInterval updates the refresh interval
func (b *BackgroundRefresher) SetInterval(interval time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.interval = interval
}
