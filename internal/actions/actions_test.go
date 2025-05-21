package actions

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

// TestMarkAsRead tests the MarkAsRead function
func TestMarkAsRead(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock function
	mockClient.MarkThreadReadFunc = func(threadID string) (*github.Response, error) {
		if threadID != "123" {
			t.Errorf("Expected thread ID 123, got %s", threadID)
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
	result, err := MarkAsRead(ctx, "123")
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
	if result.Action.NotificationID != "123" {
		t.Errorf("Expected notification ID 123, got %s", result.Action.NotificationID)
	}
}

// TestMarkAsReadError tests the MarkAsRead function with an error
func TestMarkAsReadError(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock function
	mockClient.MarkThreadReadFunc = func(threadID string) (*github.Response, error) {
		return &github.Response{
			Response: &http.Response{
				StatusCode: http.StatusInternalServerError,
			},
		}, errors.New("server error")
	}

	// Create a context
	ctx := context.Background()

	// Call MarkAsRead
	result, err := MarkAsRead(ctx, "123")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	// Check the result
	if result.Success {
		t.Errorf("Expected failure, got success")
	}
}

// TestMarkMultipleAsRead tests the MarkMultipleAsRead function
func TestMarkMultipleAsRead(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Track the number of calls
	calls := 0

	// Set up the mock function
	mockClient.MarkThreadReadFunc = func(threadID string) (*github.Response, error) {
		calls++
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
	if calls != 3 {
		t.Errorf("Expected 3 calls to MarkThreadRead, got %d", calls)
	}
}

// TestBatchProcessor tests the BatchProcessor
func TestBatchProcessor(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, nil)

	// Add tasks
	for i := 0; i < 10; i++ {
		processor.AddTask(func() (Action, error) {
			return Action{
				Type:      ActionMarkAsRead,
				Timestamp: time.Now(),
				Success:   true,
			}, nil
		})
	}

	// Process the tasks
	result := processor.Process()

	// Check the result
	if result.TotalCount != 10 {
		t.Errorf("Expected total count 10, got %d", result.TotalCount)
	}
	if result.SuccessCount != 10 {
		t.Errorf("Expected success count 10, got %d", result.SuccessCount)
	}
	if result.FailureCount != 0 {
		t.Errorf("Expected failure count 0, got %d", result.FailureCount)
	}
}

// TestBatchProcessorWithErrors tests the BatchProcessor with errors
func TestBatchProcessorWithErrors(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, nil)

	// Add tasks with some errors
	for i := 0; i < 10; i++ {
		i := i // Capture for closure
		processor.AddTask(func() (Action, error) {
			if i%2 == 0 {
				return Action{
					Type:      ActionMarkAsRead,
					Timestamp: time.Now(),
					Success:   true,
				}, nil
			}
			return Action{
				Type:      ActionMarkAsRead,
				Timestamp: time.Now(),
				Success:   false,
				Error:     errors.New("test error"),
			}, errors.New("test error")
		})
	}

	// Process the tasks
	result := processor.Process()

	// Check the result
	if result.TotalCount != 10 {
		t.Errorf("Expected total count 10, got %d", result.TotalCount)
	}
	if result.SuccessCount != 5 {
		t.Errorf("Expected success count 5, got %d", result.SuccessCount)
	}
	if result.FailureCount != 5 {
		t.Errorf("Expected failure count 5, got %d", result.FailureCount)
	}
	if len(result.Errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(result.Errors))
	}
}


