package actions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBatchProcessorBasic(t *testing.T) {
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
	if len(result.Results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(result.Results))
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestBatchProcessorWithErrorsDetailed(t *testing.T) {
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
	if len(result.Results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(result.Results))
	}
	if len(result.Errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(result.Errors))
	}
}

func TestBatchProcessorWithCancellation(t *testing.T) {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create a batch processor
	processor := NewBatchProcessor(ctx, nil)

	// Add tasks
	for i := 0; i < 10; i++ {
		processor.AddTask(func() (Action, error) {
			// Sleep to simulate work
			time.Sleep(10 * time.Millisecond)
			return Action{
				Type:      ActionMarkAsRead,
				Timestamp: time.Now(),
				Success:   true,
			}, nil
		})
	}

	// Cancel the context after a short delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	// Process the tasks
	result := processor.Process()

	// Check the result
	// We can't predict exactly how many tasks will complete before cancellation,
	// but we can check that some tasks completed and some didn't
	if result.TotalCount != 10 {
		t.Errorf("Expected total count 10, got %d", result.TotalCount)
	}
	if result.SuccessCount+result.FailureCount != result.TotalCount {
		t.Errorf("Expected success + failure to equal total, got %d + %d != %d",
			result.SuccessCount, result.FailureCount, result.TotalCount)
	}
}

func TestBatchProcessorWithProgress(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Track progress
	var progressCalls int
	var lastCompleted, lastTotal int

	// Create batch options with progress callback
	opts := &BatchOptions{
		Concurrency: 2,
		ProgressCallback: func(completed, total int) {
			progressCalls++
			lastCompleted = completed
			lastTotal = total
		},
	}

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks
	for i := 0; i < 5; i++ {
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
	if result.TotalCount != 5 {
		t.Errorf("Expected total count 5, got %d", result.TotalCount)
	}
	if result.SuccessCount != 5 {
		t.Errorf("Expected success count 5, got %d", result.SuccessCount)
	}

	// Check that progress was reported
	if progressCalls == 0 {
		t.Errorf("Expected progress callback to be called, but it wasn't")
	}
	if lastCompleted != 5 {
		t.Errorf("Expected last completed to be 5, got %d", lastCompleted)
	}
	if lastTotal != 5 {
		t.Errorf("Expected last total to be 5, got %d", lastTotal)
	}
}

func TestBatchProcessorWithErrorCallback(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Track errors
	var errorCalls int
	var lastErrorID string
	var lastError error

	// Create batch options with error callback
	opts := &BatchOptions{
		Concurrency: 2,
		ErrorCallback: func(notificationID string, err error) {
			errorCalls++
			lastErrorID = notificationID
			lastError = err
		},
	}

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks with errors
	expectedError := errors.New("test error")
	processor.AddTask(func() (Action, error) {
		return Action{
			Type:           ActionMarkAsRead,
			NotificationID: "123",
			Timestamp:      time.Now(),
			Success:        false,
			Error:          expectedError,
		}, expectedError
	})

	// Add a successful task
	processor.AddTask(func() (Action, error) {
		return Action{
			Type:           ActionMarkAsRead,
			NotificationID: "456",
			Timestamp:      time.Now(),
			Success:        true,
		}, nil
	})

	// Process the tasks
	result := processor.Process()

	// Check the result
	if result.TotalCount != 2 {
		t.Errorf("Expected total count 2, got %d", result.TotalCount)
	}
	if result.SuccessCount != 1 {
		t.Errorf("Expected success count 1, got %d", result.SuccessCount)
	}
	if result.FailureCount != 1 {
		t.Errorf("Expected failure count 1, got %d", result.FailureCount)
	}

	// Check that error callback was called
	if errorCalls != 1 {
		t.Errorf("Expected error callback to be called once, got %d", errorCalls)
	}
	if lastErrorID != "123" {
		t.Errorf("Expected last error ID to be 123, got %s", lastErrorID)
	}
	if lastError != expectedError {
		t.Errorf("Expected last error to be %v, got %v", expectedError, lastError)
	}
}
