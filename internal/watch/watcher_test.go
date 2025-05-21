package watch

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

// MockClient is a mock GitHub client for testing
type MockClient struct {
	notifications []*github.Notification
	callCount     int
	client        *github.Client
	ctx           context.Context
}

// GetUnreadNotifications returns mock notifications
func (m *MockClient) GetUnreadNotifications(options githubclient.NotificationOptions) ([]*github.Notification, error) {
	m.callCount++
	return m.notifications, nil
}

// GetAllNotifications returns mock notifications
func (m *MockClient) GetAllNotifications(options githubclient.NotificationOptions) ([]*github.Notification, error) {
	m.callCount++
	return m.notifications, nil
}

// GetNotificationsByRepo returns mock notifications
func (m *MockClient) GetNotificationsByRepo(repo string, options githubclient.NotificationOptions) ([]*github.Notification, error) {
	m.callCount++
	return m.notifications, nil
}

// GetNotificationsByOrg returns mock notifications
func (m *MockClient) GetNotificationsByOrg(org string, options githubclient.NotificationOptions) ([]*github.Notification, error) {
	m.callCount++
	return m.notifications, nil
}

// MarkNotificationRead marks a notification as read
func (m *MockClient) MarkNotificationRead(threadID string) error {
	m.callCount++
	return nil
}

// MarkAllNotificationsRead marks all notifications as read
func (m *MockClient) MarkAllNotificationsRead() error {
	m.callCount++
	return nil
}

// MarkRepositoryNotificationsRead marks all notifications in a repository as read
func (m *MockClient) MarkRepositoryNotificationsRead(owner, repo string) error {
	m.callCount++
	return nil
}

// FetchNotificationDetails fetches additional details for notifications
func (m *MockClient) FetchNotificationDetails(notifications []*github.Notification) error {
	m.callCount++
	return nil
}

// GetRawClient returns the underlying GitHub client
func (m *MockClient) GetRawClient() *github.Client {
	return m.client
}

// SetRawClient sets the underlying GitHub client
func (m *MockClient) SetRawClient(client *github.Client) {
	m.client = client
}

// WithContext returns a new Client with the given context
func (m *MockClient) WithContext(ctx context.Context) *MockClient {
	m.ctx = ctx
	return m
}

func TestWatcher(t *testing.T) {
	// Create a mock client
	mockClient := &MockClient{
		notifications: []*github.Notification{
			{
				ID: github.String("1"),
				Subject: &github.NotificationSubject{
					Title: github.String("Test notification"),
					Type:  github.String("PullRequest"),
				},
				Reason: github.String("mention"),
				Repository: &github.Repository{
					FullName: github.String("owner/repo"),
				},
				UpdatedAt: &github.Timestamp{Time: time.Now()},
			},
		},
	}

	// Create watch options
	options := DefaultWatchOptions()
	options.RefreshInterval = 100 * time.Millisecond
	options.MaxRefreshInterval = 500 * time.Millisecond
	options.BackoffFactor = 2.0
	options.BackoffThreshold = 2

	// Create event and error channels
	eventCh := make(chan NotificationEvent, 10)
	errorCh := make(chan error, 10)
	statsCh := make(chan WatchStats, 10)

	// Set up callbacks
	options.EventCallback = func(event NotificationEvent) {
		eventCh <- event
	}
	options.ErrorCallback = func(err error) {
		errorCh <- err
	}
	options.StatsCallback = func(stats WatchStats) {
		statsCh <- stats
	}

	// Create a watcher
	watcher := NewWatcher(mockClient, options)

	// Start the watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Wait for the first refresh
	select {
	case stats := <-statsCh:
		if stats.RefreshCount != 1 {
			t.Errorf("Expected refresh count 1, got %d", stats.RefreshCount)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Timed out waiting for first refresh")
	}

	// Wait for the second refresh
	select {
	case stats := <-statsCh:
		if stats.RefreshCount != 2 {
			t.Errorf("Expected refresh count 2, got %d", stats.RefreshCount)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("Timed out waiting for second refresh")
	}

	// Stop the watcher
	watcher.Stop()

	// Check that the client was called
	if mockClient.callCount < 2 {
		t.Errorf("Expected client to be called at least twice, got %d", mockClient.callCount)
	}
}

func TestWatcherBackoff(t *testing.T) {
	// Create a mock client
	mockClient := &MockClient{
		notifications: []*github.Notification{},
	}

	// Create watch options
	options := DefaultWatchOptions()
	options.RefreshInterval = 100 * time.Millisecond
	options.MaxRefreshInterval = 500 * time.Millisecond
	options.BackoffFactor = 2.0
	options.BackoffThreshold = 2

	// Create stats channel
	statsCh := make(chan WatchStats, 10)

	// Set up callback
	options.StatsCallback = func(stats WatchStats) {
		statsCh <- stats
	}

	// Create a watcher
	watcher := NewWatcher(mockClient, options)

	// Start the watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Wait for several refreshes
	var stats WatchStats
	for i := 0; i < 5; i++ {
		select {
		case stats = <-statsCh:
			// Continue
		case <-time.After(1 * time.Second):
			t.Fatalf("Timed out waiting for refresh %d", i+1)
		}
	}

	// Check that backoff was applied
	if stats.CurrentRefreshInterval <= options.RefreshInterval {
		t.Errorf("Expected backoff to increase refresh interval, got %v", stats.CurrentRefreshInterval)
	}

	// Stop the watcher
	watcher.Stop()
}

func TestWatcherEvents(t *testing.T) {
	// Create initial notifications
	initialNotifications := []*github.Notification{
		{
			ID: github.String("1"),
			Subject: &github.NotificationSubject{
				Title: github.String("Test notification 1"),
				Type:  github.String("PullRequest"),
			},
			Repository: &github.Repository{
				FullName: github.String("owner/repo"),
			},
			UpdatedAt: &github.Timestamp{Time: time.Now()},
		},
	}

	// Create updated notifications (with a new notification)
	updatedNotifications := []*github.Notification{
		{
			ID: github.String("1"),
			Subject: &github.NotificationSubject{
				Title: github.String("Test notification 1"),
				Type:  github.String("PullRequest"),
			},
			Repository: &github.Repository{
				FullName: github.String("owner/repo"),
			},
			UpdatedAt: &github.Timestamp{Time: time.Now()},
		},
		{
			ID: github.String("2"),
			Subject: &github.NotificationSubject{
				Title: github.String("Test notification 2"),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("owner/repo"),
			},
			UpdatedAt: &github.Timestamp{Time: time.Now()},
		},
	}

	// Create a mock client that returns different notifications on each call
	mockClient := &MockClient{
		notifications: initialNotifications,
	}

	// Create watch options
	options := DefaultWatchOptions()
	options.RefreshInterval = 100 * time.Millisecond

	// Create event channel
	eventCh := make(chan NotificationEvent, 10)

	// Set up callback
	options.EventCallback = func(event NotificationEvent) {
		eventCh <- event
	}

	// Create a watcher
	watcher := NewWatcher(mockClient, options)

	// Start the watcher
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Wait for the first refresh
	time.Sleep(200 * time.Millisecond)

	// Update the mock client to return the updated notifications
	mockClient.notifications = updatedNotifications

	// Wait for the second refresh
	var foundNewNotification bool
	timeout := time.After(1 * time.Second)

	for !foundNewNotification {
		select {
		case event := <-eventCh:
			if event.Type == EventNew && event.Notification.GetID() == "2" {
				foundNewNotification = true
			}
		case <-timeout:
			if !foundNewNotification {
				t.Fatalf("Timed out waiting for new notification event")
			}
		}
	}

	// Stop the watcher
	watcher.Stop()
}
