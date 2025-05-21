package actions

import (
	"context"
	"fmt"
	"time"
)

// MarkAsRead marks a notification as read
func MarkAsRead(ctx context.Context, notificationID string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionMarkAsRead,
		NotificationID: notificationID,
		Timestamp:      time.Now(),
	}

	// Mark the notification as read
	resp, err := client.MarkThreadRead(notificationID)
	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d", resp.StatusCode)
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

// MarkMultipleAsRead marks multiple notifications as read
func MarkMultipleAsRead(ctx context.Context, notificationIDs []string, opts *BatchOptions) (*BatchResult, error) {
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
			// Create a client
			client, err := GetClient(ctx)
			if err != nil {
				return Action{
					Type:           ActionMarkAsRead,
					NotificationID: id,
					Timestamp:      time.Now(),
					Success:        false,
					Error:          err,
				}, err
			}

			// Mark the notification as read
			resp, err := client.MarkThreadRead(id)

			// Create the action
			action := Action{
				Type:           ActionMarkAsRead,
				NotificationID: id,
				Timestamp:      time.Now(),
			}

			if err != nil {
				action.Success = false
				action.Error = err
				return action, err
			}

			// Check the response status
			if resp != nil && resp.StatusCode >= 400 {
				err := fmt.Errorf("server returned status %d", resp.StatusCode)
				action.Success = false
				action.Error = err
				return action, err
			}

			action.Success = true
			return action, nil
		})
	}

	// Process the tasks
	result := processor.Process()

	// Add successful actions to history
	if history := GetActionHistory(); history != nil {
		for _, r := range result.Results {
			if r.Success {
				history.Add(r.Action)
			}
		}
	}

	return result, nil
}

// MarkAllAsRead marks all notifications as read
func MarkAllAsRead(ctx context.Context) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:      ActionMarkAllAsRead,
		Timestamp: time.Now(),
	}

	// Mark all notifications as read
	err = client.MarkAllNotificationsRead()
	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to mark all notifications as read: %w", err)
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

// MarkRepositoryNotificationsAsRead marks all notifications in a repository as read
func MarkRepositoryNotificationsAsRead(ctx context.Context, owner, repo string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Create the action
	action := Action{
		Type:           ActionMarkAllAsRead,
		RepositoryName: fmt.Sprintf("%s/%s", owner, repo),
		Timestamp:      time.Now(),
	}

	// Mark all notifications in the repository as read
	err = client.MarkRepositoryNotificationsRead(owner, repo)
	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to mark repository notifications as read: %w", err)
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
