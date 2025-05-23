package watch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/filter"
	githubclient "github.com/SharanRP/gh-notif/internal/github"
	"github.com/google/go-github/v60/github"
)

// NotificationEvent represents a notification event
type NotificationEvent struct {
	// Type is the event type
	Type EventType
	// Notification is the notification
	Notification *github.Notification
	// Timestamp is when the event occurred
	Timestamp time.Time
}

// EventType represents the type of notification event
type EventType string

const (
	// EventNew is a new notification
	EventNew EventType = "new"
	// EventUpdated is an updated notification
	EventUpdated EventType = "updated"
	// EventRead is a notification marked as read
	EventRead EventType = "read"
)

// WatchOptions contains options for watching notifications
type WatchOptions struct {
	// RefreshInterval is the interval between refreshes
	RefreshInterval time.Duration
	// MaxRefreshInterval is the maximum refresh interval (for backoff)
	MaxRefreshInterval time.Duration
	// BackoffFactor is the factor to increase the refresh interval by
	BackoffFactor float64
	// BackoffThreshold is the threshold for triggering backoff
	BackoffThreshold int
	// Filter is the filter to apply to notifications
	Filter filter.Filter
	// ShowDesktopNotifications determines whether to show desktop notifications
	ShowDesktopNotifications bool
	// DesktopNotificationCommand is the command to use for desktop notifications
	DesktopNotificationCommand string
	// DesktopNotificationArgs are the arguments for the desktop notification command
	DesktopNotificationArgs []string
	// EventCallback is called when a notification event occurs
	EventCallback func(event NotificationEvent)
	// ErrorCallback is called when an error occurs
	ErrorCallback func(err error)
	// StatsCallback is called with watch statistics
	StatsCallback func(stats WatchStats)
}

// DefaultWatchOptions returns the default watch options
func DefaultWatchOptions() *WatchOptions {
	return &WatchOptions{
		RefreshInterval:            30 * time.Second,
		MaxRefreshInterval:         5 * time.Minute,
		BackoffFactor:              1.5,
		BackoffThreshold:           3,
		ShowDesktopNotifications:   false,
		DesktopNotificationCommand: getDefaultNotificationCommand(),
		DesktopNotificationArgs:    getDefaultNotificationArgs(),
	}
}

// getDefaultNotificationCommand returns the default notification command for the platform
func getDefaultNotificationCommand() string {
	// This would be platform-specific
	// For now, return a placeholder
	return "echo"
}

// getDefaultNotificationArgs returns the default notification arguments for the platform
func getDefaultNotificationArgs() []string {
	// This would be platform-specific
	// For now, return a placeholder
	return []string{}
}

// WatchStats contains statistics about the watch operation
type WatchStats struct {
	// StartTime is when the watch started
	StartTime time.Time
	// LastRefreshTime is when the last refresh occurred
	LastRefreshTime time.Time
	// NextRefreshTime is when the next refresh will occur
	NextRefreshTime time.Time
	// RefreshCount is the number of refreshes
	RefreshCount int
	// NewNotificationCount is the number of new notifications
	NewNotificationCount int
	// UpdatedNotificationCount is the number of updated notifications
	UpdatedNotificationCount int
	// ReadNotificationCount is the number of read notifications
	ReadNotificationCount int
	// ErrorCount is the number of errors
	ErrorCount int
	// CurrentRefreshInterval is the current refresh interval
	CurrentRefreshInterval time.Duration
	// IdleCount is the number of consecutive idle refreshes
	IdleCount int
}

// GitHubClient is an interface for GitHub client operations
type GitHubClient interface {
	GetUnreadNotifications(options githubclient.NotificationOptions) ([]*github.Notification, error)
	GetAllNotifications(options githubclient.NotificationOptions) ([]*github.Notification, error)
}

// Watcher watches for notification changes
type Watcher struct {
	// Options are the watch options
	Options *WatchOptions
	// Client is the GitHub client
	Client GitHubClient
	// Context is the context for cancellation
	Context context.Context
	// CancelFunc is the function to cancel the context
	CancelFunc context.CancelFunc
	// Notifications are the current notifications
	Notifications []*github.Notification
	// NotificationMap is a map of notification ID to notification
	NotificationMap map[string]*github.Notification
	// Stats are the watch statistics
	Stats WatchStats
	// Mu protects the notifications and stats
	Mu sync.RWMutex
	// Running indicates whether the watcher is running
	Running bool
}

// NewWatcher creates a new watcher
func NewWatcher(client GitHubClient, options *WatchOptions) *Watcher {
	if options == nil {
		options = DefaultWatchOptions()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Watcher{
		Options:         options,
		Client:          client,
		Context:         ctx,
		CancelFunc:      cancel,
		Notifications:   make([]*github.Notification, 0),
		NotificationMap: make(map[string]*github.Notification),
		Stats: WatchStats{
			StartTime:              time.Now(),
			CurrentRefreshInterval: options.RefreshInterval,
		},
	}
}

// Start starts watching for notification changes
func (w *Watcher) Start() error {
	if w.Running {
		return fmt.Errorf("watcher is already running")
	}

	w.Running = true
	w.Stats.StartTime = time.Now()

	// Start the watch loop
	go w.watchLoop()

	return nil
}

// Stop stops watching for notification changes
func (w *Watcher) Stop() {
	if !w.Running {
		return
	}

	w.CancelFunc()
	w.Running = false
}

// watchLoop is the main watch loop
func (w *Watcher) watchLoop() {
	// Initial refresh
	w.refresh()

	// Create a ticker for refreshing
	ticker := time.NewTicker(w.Options.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Refresh notifications
			w.refresh()

			// Update the ticker interval based on backoff
			w.updateRefreshInterval()
			ticker.Reset(w.Stats.CurrentRefreshInterval)

		case <-w.Context.Done():
			// Context cancelled, stop watching
			return
		}
	}
}

// refresh refreshes notifications and detects changes
func (w *Watcher) refresh() {
	w.Mu.Lock()
	defer w.Mu.Unlock()

	// Update stats
	w.Stats.LastRefreshTime = time.Now()
	w.Stats.RefreshCount++
	w.Stats.NextRefreshTime = time.Now().Add(w.Stats.CurrentRefreshInterval)

	// Fetch notifications
	notifications, err := w.Client.GetUnreadNotifications(githubclient.NotificationOptions{
		All:      true,
		UseCache: false,
	})

	if err != nil {
		w.Stats.ErrorCount++
		if w.Options.ErrorCallback != nil {
			w.Options.ErrorCallback(err)
		}
		return
	}

	// Filter notifications if a filter is specified
	if w.Options.Filter != nil {
		var filtered []*github.Notification
		for _, n := range notifications {
			if w.Options.Filter.Apply(n) {
				filtered = append(filtered, n)
			}
		}
		notifications = filtered
	}

	// Check for changes
	changes := false
	newNotifications := make([]*github.Notification, 0)
	updatedNotifications := make([]*github.Notification, 0)
	readNotifications := make([]*github.Notification, 0)

	// Create a map of new notifications
	newNotificationMap := make(map[string]*github.Notification)
	for _, n := range notifications {
		newNotificationMap[n.GetID()] = n
	}

	// Check for new and updated notifications
	for _, n := range notifications {
		id := n.GetID()
		existing, ok := w.NotificationMap[id]
		if !ok {
			// New notification
			newNotifications = append(newNotifications, n)
			changes = true
		} else if n.GetUpdatedAt().After(existing.GetUpdatedAt().Time) {
			// Updated notification
			updatedNotifications = append(updatedNotifications, n)
			changes = true
		}
	}

	// Check for read notifications
	for id, n := range w.NotificationMap {
		if _, ok := newNotificationMap[id]; !ok {
			// Notification no longer in the list (marked as read)
			readNotifications = append(readNotifications, n)
			changes = true
		}
	}

	// Update the notification map
	w.NotificationMap = newNotificationMap
	w.Notifications = notifications

	// Update stats
	w.Stats.NewNotificationCount += len(newNotifications)
	w.Stats.UpdatedNotificationCount += len(updatedNotifications)
	w.Stats.ReadNotificationCount += len(readNotifications)

	// If there were no changes, increment the idle count
	if !changes {
		w.Stats.IdleCount++
	} else {
		w.Stats.IdleCount = 0
	}

	// Call the stats callback
	if w.Options.StatsCallback != nil {
		w.Options.StatsCallback(w.Stats)
	}

	// Process events
	w.processEvents(newNotifications, updatedNotifications, readNotifications)
}

// processEvents processes notification events
func (w *Watcher) processEvents(newNotifications, updatedNotifications, readNotifications []*github.Notification) {
	// Process new notifications
	for _, n := range newNotifications {
		if w.Options.EventCallback != nil {
			w.Options.EventCallback(NotificationEvent{
				Type:         EventNew,
				Notification: n,
				Timestamp:    time.Now(),
			})
		}

		// Show desktop notification
		if w.Options.ShowDesktopNotifications {
			w.showDesktopNotification(n, EventNew)
		}
	}

	// Process updated notifications
	for _, n := range updatedNotifications {
		if w.Options.EventCallback != nil {
			w.Options.EventCallback(NotificationEvent{
				Type:         EventUpdated,
				Notification: n,
				Timestamp:    time.Now(),
			})
		}

		// Show desktop notification
		if w.Options.ShowDesktopNotifications {
			w.showDesktopNotification(n, EventUpdated)
		}
	}

	// Process read notifications
	for _, n := range readNotifications {
		if w.Options.EventCallback != nil {
			w.Options.EventCallback(NotificationEvent{
				Type:         EventRead,
				Notification: n,
				Timestamp:    time.Now(),
			})
		}
	}
}

// showDesktopNotification shows a desktop notification
func (w *Watcher) showDesktopNotification(n *github.Notification, eventType EventType) {
	// This would use the platform-specific notification command
	// For now, just print to stdout
	fmt.Printf("Notification: %s - %s\n", eventType, n.GetSubject().GetTitle())
}

// updateRefreshInterval updates the refresh interval based on backoff
func (w *Watcher) updateRefreshInterval() {
	// If we've had several idle refreshes, increase the interval
	if w.Stats.IdleCount >= w.Options.BackoffThreshold {
		// Calculate the new interval
		newInterval := time.Duration(float64(w.Stats.CurrentRefreshInterval) * w.Options.BackoffFactor)

		// Cap at the maximum interval
		if newInterval > w.Options.MaxRefreshInterval {
			newInterval = w.Options.MaxRefreshInterval
		}

		w.Stats.CurrentRefreshInterval = newInterval
	} else if w.Stats.IdleCount == 0 {
		// If we had changes, reset to the base interval
		w.Stats.CurrentRefreshInterval = w.Options.RefreshInterval
	}
}
