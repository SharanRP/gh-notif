package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
)

// ArchiveNotification archives a notification
func ArchiveNotification(ctx context.Context, notificationID string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionArchive,
		NotificationID: notificationID,
		Timestamp:      time.Now(),
	}

	// First, mark the notification as read (required for archiving)
	resp, err := client.MarkThreadRead(notificationID)
	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to mark notification as read before archiving: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when marking as read", resp.StatusCode)
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, err
	}

	// Now archive the notification by updating the subscription
	// We don't actually need to parse the thread ID for this operation

	// Update the subscription to ignore the thread
	sub := &github.Subscription{
		Ignored: github.Bool(true),
	}
	_, resp, err = client.GetRawClient().Activity.SetThreadSubscription(ctx, notificationID, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to archive notification: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when archiving", resp.StatusCode)
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

// UnarchiveNotification unarchives a notification
func UnarchiveNotification(ctx context.Context, notificationID string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionUnarchive,
		NotificationID: notificationID,
		Timestamp:      time.Now(),
	}

	// Update the subscription to unignore the thread
	sub := &github.Subscription{
		Ignored: github.Bool(false),
	}
	_, resp, err := client.GetRawClient().Activity.SetThreadSubscription(ctx, notificationID, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to unarchive notification: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when unarchiving", resp.StatusCode)
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

// ArchiveMultipleNotifications archives multiple notifications
func ArchiveMultipleNotifications(ctx context.Context, notificationIDs []string, opts *BatchOptions) (*BatchResult, error) {
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
		// Capture for closure
		processor.AddTask(func() (Action, error) {
			result, err := ArchiveNotification(ctx, id)
			if err != nil {
				return Action{
					Type:           ActionArchive,
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

// UnarchiveMultipleNotifications unarchives multiple notifications
func UnarchiveMultipleNotifications(ctx context.Context, notificationIDs []string, opts *BatchOptions) (*BatchResult, error) {
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
		// Capture for closure
		processor.AddTask(func() (Action, error) {
			result, err := UnarchiveNotification(ctx, id)
			if err != nil {
				return Action{
					Type:           ActionUnarchive,
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

// Helper function to parse a thread ID
func parseThreadID(threadID string) (map[string]string, error) {
	// In a real implementation, this would parse the thread ID to extract
	// owner, repo, and thread number. For now, we'll return a placeholder.
	return map[string]string{
		"id": threadID,
	}, nil
}
