package common

import (
	"fmt"
	"time"
)

// ActionType represents the type of action
type ActionType string

const (
	// ActionMarkAsRead represents marking a notification as read
	ActionMarkAsRead ActionType = "mark_as_read"
	// ActionMarkAllAsRead represents marking all notifications as read
	ActionMarkAllAsRead ActionType = "mark_all_as_read"
	// ActionArchive represents archiving a notification
	ActionArchive ActionType = "archive"
	// ActionUnarchive represents unarchiving a notification
	ActionUnarchive ActionType = "unarchive"
	// ActionSubscribe represents subscribing to a thread
	ActionSubscribe ActionType = "subscribe"
	// ActionUnsubscribe represents unsubscribing from a thread
	ActionUnsubscribe ActionType = "unsubscribe"
	// ActionMute represents muting a repository
	ActionMute ActionType = "mute"
)

// Action represents an action performed on a notification
type Action struct {
	// Type is the type of action
	Type ActionType
	// NotificationID is the ID of the notification
	NotificationID string
	// RepositoryName is the name of the repository
	RepositoryName string
	// Timestamp is when the action was performed
	Timestamp time.Time
	// Success indicates whether the action was successful
	Success bool
	// Error is the error that occurred, if any
	Error error
	// Metadata is additional information about the action
	Metadata map[string]interface{}
}

// ActionResult represents the result of an action
type ActionResult struct {
	// Action is the action that was performed
	Action Action
	// Success indicates whether the action was successful
	Success bool
	// Error is the error that occurred, if any
	Error error
}

// BatchResult represents the result of a batch operation
type BatchResult struct {
	// TotalCount is the total number of operations
	TotalCount int
	// SuccessCount is the number of successful operations
	SuccessCount int
	// FailureCount is the number of failed operations
	FailureCount int
	// Results is the results of each operation
	Results []ActionResult
	// Errors is the errors that occurred
	Errors []error
	// Duration is how long the operation took
	Duration time.Duration
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return "less than a second"
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
}

// FormatProgress formats progress as a percentage
func FormatProgress(completed, total int) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", float64(completed)/float64(total)*100)
}

// BatchOptions represents options for batch operations
type BatchOptions struct {
	// Concurrency is the number of concurrent operations
	Concurrency int
	// ProgressCallback is called when progress is made
	ProgressCallback func(completed, total int)
	// ErrorCallback is called when an error occurs
	ErrorCallback func(notificationID string, err error)
	// ContinueOnError indicates whether to continue on error
	ContinueOnError bool
	// Timeout is the maximum time to wait for the operation to complete
	Timeout time.Duration
}

// UndoResult represents the result of an undo operation
type UndoResult struct {
	// OriginalAction is the action that was undone
	OriginalAction Action
	// UndoAction is the action that was performed to undo the original action
	UndoAction Action
	// Success indicates whether the undo was successful
	Success bool
	// Error is the error that occurred, if any
	Error error
}
