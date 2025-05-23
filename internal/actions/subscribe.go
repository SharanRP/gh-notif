package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
)

// SubscribeToThread subscribes to a notification thread
func SubscribeToThread(ctx context.Context, notificationID string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionSubscribe,
		NotificationID: notificationID,
		Timestamp:      time.Now(),
	}

	// Update the subscription to subscribe to the thread
	sub := &github.Subscription{
		Subscribed: github.Bool(true),
		Ignored:    github.Bool(false),
	}
	_, resp, err := client.GetRawClient().Activity.SetThreadSubscription(ctx, notificationID, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to subscribe to thread: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when subscribing", resp.StatusCode)
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, err
	}

	// Record the successful action
	action.Success = true

	// Add to history if available
	if history := GetActionHistory(); history != nil {
		history.Add(action)
	}

	return &ActionResult{
		Action:  action,
		Success: true,
	}, nil
}

// UnsubscribeFromThread unsubscribes from a notification thread
func UnsubscribeFromThread(ctx context.Context, notificationID string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionUnsubscribe,
		NotificationID: notificationID,
		Timestamp:      time.Now(),
	}

	// Update the subscription to unsubscribe from the thread
	sub := &github.Subscription{
		Subscribed: github.Bool(false),
	}
	_, resp, err := client.GetRawClient().Activity.SetThreadSubscription(ctx, notificationID, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to unsubscribe from thread: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when unsubscribing", resp.StatusCode)
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, err
	}

	// Record the successful action
	action.Success = true

	// Add to history if available
	if history := GetActionHistory(); history != nil {
		history.Add(action)
	}

	return &ActionResult{
		Action:  action,
		Success: true,
	}, nil
}

// SubscribeToMultipleThreads subscribes to multiple notification threads
func SubscribeToMultipleThreads(ctx context.Context, notificationIDs []string, opts *BatchOptions) (*BatchResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks to the processor
	for _, id := range notificationIDs {
		id := id // Capture for closure
		processor.AddTask(func() (Action, error) {
			result, err := SubscribeToThread(ctx, id)
			if err != nil {
				return Action{
					Type:           ActionSubscribe,
					NotificationID: id,
					Timestamp:      time.Now(),
					Success:        false,
					Error:          err,
				}, err
			}
			return result.Action, nil
		})
	}

	// Process the tasks
	return processor.Process(), nil
}

// UnsubscribeFromMultipleThreads unsubscribes from multiple notification threads
func UnsubscribeFromMultipleThreads(ctx context.Context, notificationIDs []string, opts *BatchOptions) (*BatchResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks to the processor
	for _, id := range notificationIDs {
		id := id // Capture for closure
		processor.AddTask(func() (Action, error) {
			result, err := UnsubscribeFromThread(ctx, id)
			if err != nil {
				return Action{
					Type:           ActionUnsubscribe,
					NotificationID: id,
					Timestamp:      time.Now(),
					Success:        false,
					Error:          err,
				}, err
			}
			return result.Action, nil
		})
	}

	// Process the tasks
	return processor.Process(), nil
}
