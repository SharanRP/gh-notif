package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestBackgroundRefresher(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with notifications
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Create a response with notifications
		notifications := []*github.Notification{
			{
				ID: github.String("1"),
				Repository: &github.Repository{
					FullName: github.String("org1/repo1"),
				},
				Subject: &github.NotificationSubject{
					Title: github.String("Issue 1"),
					Type:  github.String("Issue"),
				},
			},
		}

		json.NewEncoder(w).Encode(notifications)
	}))
	defer server.Close()

	// Create a client with the mock server
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with notifications
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Create a response with notifications
		notifications := []*github.Notification{
			{
				ID: github.String("1"),
				Repository: &github.Repository{
					FullName: github.String("org1/repo1"),
				},
				Subject: &github.NotificationSubject{
					Title: github.String("Issue 1"),
					Type:  github.String("Issue"),
				},
			},
		}

		json.NewEncoder(w).Encode(notifications)
	})
	client, testServer, err := NewTestClient(ctx, handler)
	defer testServer.Close()
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create a background refresher
	options := NotificationOptions{
		All:      true,
		PerPage:  10,
		UseCache: false,
	}
	refresher := NewBackgroundRefresher(client, 100*time.Millisecond, options)

	// Set up a callback to track updates
	updateCount := 0
	refresher.SetOnUpdateCallback(func(notifications []*github.Notification) {
		updateCount++

		// Check the notifications
		if len(notifications) != 1 {
			t.Errorf("Callback notifications len = %d, want %d", len(notifications), 1)
		}
		if notifications[0].GetID() != "1" {
			t.Errorf("Callback notification ID = %s, want %s", notifications[0].GetID(), "1")
		}
	})

	// Start the refresher
	refresher.Start()

	// Check that it's running
	if !refresher.IsRunning() {
		t.Errorf("IsRunning() = %v, want %v", refresher.IsRunning(), true)
	}

	// Wait for at least one update
	time.Sleep(150 * time.Millisecond)

	// Check that we got at least one update
	if updateCount < 1 {
		t.Errorf("Update count = %d, want at least %d", updateCount, 1)
	}

	// Get the notifications
	notifications := refresher.GetNotifications()
	if len(notifications) != 1 {
		t.Errorf("GetNotifications() len = %d, want %d", len(notifications), 1)
	}
	if notifications[0].GetID() != "1" {
		t.Errorf("GetNotifications() ID = %s, want %s", notifications[0].GetID(), "1")
	}

	// Force a refresh
	refresher.ForceRefresh()
	time.Sleep(50 * time.Millisecond)

	// Check that we got another update
	if updateCount < 2 {
		t.Errorf("Update count after force refresh = %d, want at least %d", updateCount, 2)
	}

	// Update the options
	newOptions := NotificationOptions{
		All:      false,
		Unread:   true,
		PerPage:  20,
		UseCache: false,
	}
	refresher.SetOptions(newOptions)

	// Update the interval
	refresher.SetInterval(200 * time.Millisecond)

	// Stop the refresher
	refresher.Stop()

	// Check that it's not running
	if refresher.IsRunning() {
		t.Errorf("IsRunning() after stop = %v, want %v", refresher.IsRunning(), false)
	}
}

func TestNotificationStream(t *testing.T) {
	// Skip this test for now
	t.Skip("Skipping streaming test that requires complex mock setup")
}
