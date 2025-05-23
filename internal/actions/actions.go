package actions

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/common"
	githubclient "github.com/SharanRP/gh-notif/internal/github"
)

// Common errors
var (
	ErrNotFound          = errors.New("notification not found")
	ErrOperationCanceled = errors.New("operation canceled")
	ErrPartialFailure    = errors.New("some operations failed")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrNotAuthenticated  = errors.New("not authenticated")
)

// Action types
const (
	// ActionMarkAsRead represents marking a notification as read
	ActionMarkAsRead = common.ActionMarkAsRead
	// ActionMarkAllAsRead represents marking all notifications as read
	ActionMarkAllAsRead = common.ActionMarkAllAsRead
	// ActionArchive represents archiving a notification
	ActionArchive = common.ActionArchive
	// ActionUnarchive represents unarchiving a notification
	ActionUnarchive = common.ActionUnarchive
	// ActionSubscribe represents subscribing to a thread
	ActionSubscribe = common.ActionSubscribe
	// ActionUnsubscribe represents unsubscribing from a thread
	ActionUnsubscribe = common.ActionUnsubscribe
	// ActionMute represents muting a repository
	ActionMute = common.ActionMute
)

// Action represents an action performed on a notification
type Action = common.Action

// BatchOptions contains options for batch operations
type BatchOptions struct {
	// Concurrency is the maximum number of concurrent operations
	Concurrency int
	// ProgressCallback is called to report progress
	ProgressCallback func(completed, total int)
	// ErrorCallback is called when an error occurs
	ErrorCallback func(notificationID string, err error)
	// ContinueOnError determines whether to continue on error
	ContinueOnError bool
	// Timeout is the maximum time to wait for operations to complete
	Timeout time.Duration
}

// DefaultBatchOptions returns the default batch options
func DefaultBatchOptions() *BatchOptions {
	return &BatchOptions{
		Concurrency:      5,
		ProgressCallback: nil,
		ErrorCallback:    nil,
		ContinueOnError:  true,
		Timeout:          30 * time.Second,
	}
}

// ActionResult represents the result of an action
type ActionResult = common.ActionResult

// BatchResult represents the result of a batch operation
type BatchResult = common.BatchResult

// ActionHistory tracks actions for potential undo
type ActionHistory struct {
	// Actions is a list of actions performed
	Actions []Action
	// mu protects the actions list
	mu sync.RWMutex
	// maxSize is the maximum number of actions to store
	maxSize int
}

// NewActionHistory creates a new action history
func NewActionHistory(maxSize int) *ActionHistory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &ActionHistory{
		Actions: make([]Action, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds an action to the history
func (h *ActionHistory) Add(action Action) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add the action to the beginning of the list
	h.Actions = append([]Action{action}, h.Actions...)

	// Trim the list if it exceeds the maximum size
	if len(h.Actions) > h.maxSize {
		h.Actions = h.Actions[:h.maxSize]
	}
}

// GetLast returns the last n actions
func (h *ActionHistory) GetLast(n int) []Action {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.Actions) {
		n = len(h.Actions)
	}

	result := make([]Action, n)
	copy(result, h.Actions[:n])
	return result
}

// Clear clears the action history
func (h *ActionHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Actions = make([]Action, 0, h.maxSize)
}

// Singleton instance of ActionHistory
var (
	actionHistory     *ActionHistory
	actionHistoryOnce sync.Once
)

// GetActionHistory returns the singleton instance of ActionHistory
func GetActionHistory() *ActionHistory {
	actionHistoryOnce.Do(func() {
		actionHistory = NewActionHistory(100)
	})
	return actionHistory
}

// GetClient is a function that returns a GitHub client for performing actions
var GetClient = func(ctx context.Context) (*githubclient.Client, error) {
	return githubclient.NewClient(ctx)
}
