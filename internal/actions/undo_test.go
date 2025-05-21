package actions

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

func TestUndoAction(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Test cases for different action types
	testCases := []struct {
		name        string
		action      Action
		expectError bool
	}{
		{
			name: "Undo Archive",
			action: Action{
				Type:           ActionArchive,
				NotificationID: "123456",
				Timestamp:      time.Now(),
				Success:        true,
			},
			expectError: false,
		},
		{
			name: "Undo Unarchive",
			action: Action{
				Type:           ActionUnarchive,
				NotificationID: "123456",
				Timestamp:      time.Now(),
				Success:        true,
			},
			expectError: false,
		},
		{
			name: "Undo Subscribe",
			action: Action{
				Type:           ActionSubscribe,
				NotificationID: "123456",
				Timestamp:      time.Now(),
				Success:        true,
			},
			expectError: false,
		},
		{
			name: "Undo Unsubscribe",
			action: Action{
				Type:           ActionUnsubscribe,
				NotificationID: "123456",
				Timestamp:      time.Now(),
				Success:        true,
			},
			expectError: false,
		},
		{
			name: "Undo Mute",
			action: Action{
				Type:           ActionMute,
				RepositoryName: "testowner/testrepo",
				Timestamp:      time.Now(),
				Success:        true,
			},
			expectError: false,
		},
		{
			name: "Undo Unmute",
			action: Action{
				Type:           ActionMute,
				RepositoryName: "testowner/testrepo",
				Timestamp:      time.Now(),
				Success:        true,
				Metadata: map[string]interface{}{
					"unmute": true,
				},
			},
			expectError: false,
		},
		{
			name: "Undo Mark All As Read",
			action: Action{
				Type:      ActionMarkAllAsRead,
				Timestamp: time.Now(),
				Success:   true,
			},
			expectError: true,
		},
		{
			name: "Undo Failed Action",
			action: Action{
				Type:           ActionArchive,
				NotificationID: "123456",
				Timestamp:      time.Now(),
				Success:        false,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up repository subscription mock if needed
			if tc.action.Type == ActionMute {
				mockActivityService.setRepositorySubscriptionFunc = func(ctx context.Context, owner, repo string, subscription *github.Subscription) (*github.Subscription, *github.Response, error) {
					return &github.Subscription{
						Subscribed: subscription.Subscribed,
						Ignored:    subscription.Ignored,
					}, &github.Response{
						Response: &http.Response{
							StatusCode: http.StatusOK,
						},
					}, nil
				}
			}

			// Call UndoAction
			result, err := UndoAction(ctx, tc.action)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("UndoAction failed: %v", err)
			}

			// Check the result
			if !result.Success {
				t.Errorf("Expected success, got failure")
			}

			// Check that the original action is stored
			if result.OriginalAction.Type != tc.action.Type {
				t.Errorf("Expected original action type %s, got %s", tc.action.Type, result.OriginalAction.Type)
			}
		})
	}
}

func TestUndoLastAction(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Create a test action history
	history := NewActionHistory(10)

	// Add a test action
	testAction := Action{
		Type:           ActionArchive,
		NotificationID: "123456",
		Timestamp:      time.Now(),
		Success:        true,
	}
	history.Add(testAction)

	// Override the GetActionHistory function
	originalGetActionHistory := GetActionHistory
	defer func() { GetActionHistory = originalGetActionHistory }()
	GetActionHistory = func() *ActionHistory {
		return history
	}

	// Call UndoLastAction
	result, err := UndoLastAction(ctx)
	if err != nil {
		t.Fatalf("UndoLastAction failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.OriginalAction.Type != ActionArchive {
		t.Errorf("Expected original action type %s, got %s", ActionArchive, result.OriginalAction.Type)
	}
	if result.OriginalAction.NotificationID != "123456" {
		t.Errorf("Expected notification ID 123456, got %s", result.OriginalAction.NotificationID)
	}
}

func TestUndoMultipleActions(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Create test actions
	actions := []Action{
		{
			Type:           ActionArchive,
			NotificationID: "123",
			Timestamp:      time.Now(),
			Success:        true,
		},
		{
			Type:           ActionSubscribe,
			NotificationID: "456",
			Timestamp:      time.Now(),
			Success:        true,
		},
		{
			Type:           ActionUnsubscribe,
			NotificationID: "789",
			Timestamp:      time.Now(),
			Success:        true,
		},
	}

	// Call UndoMultipleActions
	result, err := UndoMultipleActions(ctx, actions, nil)
	if err != nil {
		t.Fatalf("UndoMultipleActions failed: %v", err)
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
