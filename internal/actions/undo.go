package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/SharanRP/gh-notif/internal/common"
)

// UndoResult represents the result of an undo operation
type UndoResult = common.UndoResult

// UndoLastAction undoes the last action
func UndoLastAction(ctx context.Context) (*UndoResult, error) {
	// Get the action history
	history := GetActionHistory()
	if history == nil {
		return nil, fmt.Errorf("action history not available")
	}

	// Get the last action
	actions := history.GetLast(1)
	if len(actions) == 0 {
		return nil, fmt.Errorf("no actions to undo")
	}

	lastAction := actions[0]
	return UndoAction(ctx, lastAction)
}

// UndoAction undoes a specific action
func UndoAction(ctx context.Context, action Action) (*UndoResult, error) {
	// Create a result with the original action
	result := &UndoResult{
		OriginalAction: action,
	}

	// Check if the action was successful
	if !action.Success {
		return nil, fmt.Errorf("cannot undo a failed action")
	}

	var undoAction Action

	// Perform the appropriate undo action based on the action type
	switch action.Type {
	case ActionMarkAsRead:
		// Cannot directly mark as unread, but we can update the subscription
		// to make it appear in the unread list again
		undoResult, err := SubscribeToThread(ctx, action.NotificationID)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, fmt.Errorf("failed to undo mark as read: %w", err)
		}
		undoAction = undoResult.Action

	case ActionMarkAllAsRead:
		// Cannot directly undo marking all as read
		return nil, fmt.Errorf("cannot undo marking all notifications as read")

	case ActionArchive:
		// Undo archiving by unarchiving
		undoResult, err := UnarchiveNotification(ctx, action.NotificationID)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, fmt.Errorf("failed to undo archive: %w", err)
		}
		undoAction = undoResult.Action

	case ActionUnarchive:
		// Undo unarchiving by archiving
		undoResult, err := ArchiveNotification(ctx, action.NotificationID)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, fmt.Errorf("failed to undo unarchive: %w", err)
		}
		undoAction = undoResult.Action

	case ActionSubscribe:
		// Undo subscribing by unsubscribing
		undoResult, err := UnsubscribeFromThread(ctx, action.NotificationID)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, fmt.Errorf("failed to undo subscribe: %w", err)
		}
		undoAction = undoResult.Action

	case ActionUnsubscribe:
		// Undo unsubscribing by subscribing
		undoResult, err := SubscribeToThread(ctx, action.NotificationID)
		if err != nil {
			result.Success = false
			result.Error = err
			return result, fmt.Errorf("failed to undo unsubscribe: %w", err)
		}
		undoAction = undoResult.Action

	case ActionMute:
		// Check if this was a mute or unmute action
		if action.Metadata != nil && action.Metadata["unmute"] == true {
			// This was an unmute action, undo by muting
			undoResult, err := MuteRepository(ctx, action.RepositoryName)
			if err != nil {
				result.Success = false
				result.Error = err
				return result, fmt.Errorf("failed to undo unmute: %w", err)
			}
			undoAction = undoResult.Action
		} else {
			// This was a mute action, undo by unmuting
			undoResult, err := UnmuteRepository(ctx, action.RepositoryName)
			if err != nil {
				result.Success = false
				result.Error = err
				return result, fmt.Errorf("failed to undo mute: %w", err)
			}
			undoAction = undoResult.Action
		}

	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}

	// Set the undo action in the result
	result.UndoAction = undoAction
	result.Success = true

	return result, nil
}

// UndoMultipleActions undoes multiple actions
func UndoMultipleActions(ctx context.Context, actions []Action, opts *BatchOptions) (*BatchResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks to the processor
	for _, action := range actions {
		action := action // Capture for closure
		processor.AddTask(func() (Action, error) {
			undoResult, err := UndoAction(ctx, action)
			if err != nil {
				return common.Action{
					Type:           common.ActionType("undo_" + string(action.Type)),
					NotificationID: action.NotificationID,
					RepositoryName: action.RepositoryName,
					Timestamp:      time.Now(),
					Success:        false,
					Error:          err,
					Metadata: map[string]interface{}{
						"original_action": action,
					},
				}, err
			}
			return undoResult.UndoAction, nil
		})
	}

	// Process the tasks
	return processor.Process(), nil
}

// UndoLastNActions undoes the last N actions
func UndoLastNActions(ctx context.Context, n int, opts *BatchOptions) (*BatchResult, error) {
	// Get the action history
	history := GetActionHistory()
	if history == nil {
		return nil, fmt.Errorf("action history not available")
	}

	// Get the last N actions
	actions := history.GetLast(n)
	if len(actions) == 0 {
		return nil, fmt.Errorf("no actions to undo")
	}

	// Undo the actions
	return UndoMultipleActions(ctx, actions, opts)
}
