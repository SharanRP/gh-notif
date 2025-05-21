package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

func TestMarkAsReadDetailed(t *testing.T) {
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

	// Create a context
	ctx := context.Background()

	// Call MarkAsRead
	result, err := MarkAsRead(ctx, "123456")
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionMarkAsRead {
		t.Errorf("Expected action type %s, got %s", ActionMarkAsRead, result.Action.Type)
	}
	if result.Action.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.Action.NotificationID)
	}
}

func TestMarkAllAsRead(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock functions
	mockClient.MarkAllNotificationsReadFunc = func() error {
		return nil
	}

	// Create a context
	ctx := context.Background()

	// Call MarkAllAsRead
	result, err := MarkAllAsRead(ctx)
	if err != nil {
		t.Fatalf("MarkAllAsRead failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionMarkAllAsRead {
		t.Errorf("Expected action type %s, got %s", ActionMarkAllAsRead, result.Action.Type)
	}
}

func TestMarkRepositoryNotificationsAsRead(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock functions
	mockClient.MarkRepositoryNotificationsReadFunc = func(owner, repo string) error {
		if owner != "testowner" || repo != "testrepo" {
			t.Errorf("Expected owner/repo to be testowner/testrepo, got %s/%s", owner, repo)
		}
		return nil
	}

	// Create a context
	ctx := context.Background()

	// Call MarkRepositoryNotificationsAsRead
	result, err := MarkRepositoryNotificationsAsRead(ctx, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("MarkRepositoryNotificationsAsRead failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionMarkAllAsRead {
		t.Errorf("Expected action type %s, got %s", ActionMarkAllAsRead, result.Action.Type)
	}
	if result.Action.RepositoryName != "testowner/testrepo" {
		t.Errorf("Expected repository name testowner/testrepo, got %s", result.Action.RepositoryName)
	}
}

func TestMarkMultipleAsReadDetailed(t *testing.T) {
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

	// Create a context
	ctx := context.Background()

	// Call MarkMultipleAsRead
	ids := []string{"123", "456", "789"}
	result, err := MarkMultipleAsRead(ctx, ids, nil)
	if err != nil {
		t.Fatalf("MarkMultipleAsRead failed: %v", err)
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
	if markThreadReadCalls != 3 {
		t.Errorf("Expected 3 calls to MarkThreadRead, got %d", markThreadReadCalls)
	}
}
