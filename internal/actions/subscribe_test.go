package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

func TestSubscribeToThread(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call SubscribeToThread
	result, err := SubscribeToThread(ctx, "123456")
	if err != nil {
		t.Fatalf("SubscribeToThread failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionSubscribe {
		t.Errorf("Expected action type %s, got %s", ActionSubscribe, result.Action.Type)
	}
	if result.Action.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.Action.NotificationID)
	}
}

func TestUnsubscribeFromThread(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call UnsubscribeFromThread
	result, err := UnsubscribeFromThread(ctx, "123456")
	if err != nil {
		t.Fatalf("UnsubscribeFromThread failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionUnsubscribe {
		t.Errorf("Expected action type %s, got %s", ActionUnsubscribe, result.Action.Type)
	}
	if result.Action.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.Action.NotificationID)
	}
}

func TestSubscribeToMultipleThreads(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call SubscribeToMultipleThreads
	ids := []string{"123", "456", "789"}
	result, err := SubscribeToMultipleThreads(ctx, ids, nil)
	if err != nil {
		t.Fatalf("SubscribeToMultipleThreads failed: %v", err)
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
