package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

func TestArchiveNotification(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock functions
	mockClient.MarkThreadReadFunc = func(threadID string) (*github.Response, error) {
		if threadID != "123456" {
			t.Errorf("Expected thread ID 123456, got %s", threadID)
		}
		return &github.Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
			},
		}, nil
	}

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call ArchiveNotification
	result, err := ArchiveNotification(ctx, "123456")
	if err != nil {
		t.Fatalf("ArchiveNotification failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionArchive {
		t.Errorf("Expected action type %s, got %s", ActionArchive, result.Action.Type)
	}
	if result.Action.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.Action.NotificationID)
	}
}

func TestUnarchiveNotification(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call UnarchiveNotification
	result, err := UnarchiveNotification(ctx, "123456")
	if err != nil {
		t.Fatalf("UnarchiveNotification failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionUnarchive {
		t.Errorf("Expected action type %s, got %s", ActionUnarchive, result.Action.Type)
	}
	if result.Action.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.Action.NotificationID)
	}
}

func TestArchiveMultipleNotifications(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Track the number of calls
	markThreadReadCalls := 0

	// Set up the mock functions
	mockClient.MarkThreadReadFunc = func(threadID string) (*github.Response, error) {
		markThreadReadCalls++
		return &github.Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
			},
		}, nil
	}

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call ArchiveMultipleNotifications
	ids := []string{"123", "456", "789"}
	result, err := ArchiveMultipleNotifications(ctx, ids, nil)
	if err != nil {
		t.Fatalf("ArchiveMultipleNotifications failed: %v", err)
	}

	// Check the result
	if result.TotalCount != 3 {
		t.Errorf("Expected total count 3, got %d", result.TotalCount)
	}
	if result.SuccessCount != 3 {
		t.Errorf("Expected success count 3, got %d", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", result.FailureCount)
	}
}

func TestUnarchiveMultipleNotifications(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call UnarchiveMultipleNotifications
	ids := []string{"123", "456", "789"}
	result, err := UnarchiveMultipleNotifications(ctx, ids, nil)
	if err != nil {
		t.Fatalf("UnarchiveMultipleNotifications failed: %v", err)
	}

	// Check the result
	if result.TotalCount != 3 {
		t.Errorf("Expected total count 3, got %d", result.TotalCount)
	}
	if result.SuccessCount != 3 {
		t.Errorf("Expected success count 3, got %d", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", result.FailureCount)
	}
}
