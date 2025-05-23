package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

// MuteRepository mutes notifications for a repository
func MuteRepository(ctx context.Context, repoFullName string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Parse the repository name
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name format, expected 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

	// Create the action
	action := Action{
		Type:           ActionMute,
		RepositoryName: repoFullName,
		Timestamp:      time.Now(),
	}

	// First, mark all notifications in the repository as read
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

	// Now update the repository subscription to ignore
	sub := &github.Subscription{
		Subscribed: github.Bool(false),
		Ignored:    github.Bool(true),
	}
	_, resp, err := client.GetRawClient().Activity.SetRepositorySubscription(ctx, owner, repo, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to mute repository: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when muting repository", resp.StatusCode)
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

// UnmuteRepository unmutes notifications for a repository
func UnmuteRepository(ctx context.Context, repoFullName string) (*ActionResult, error) {
	// Create a client
	client, err := GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Parse the repository name
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository name format, expected 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

	// Create the action
	action := Action{
		Type:           ActionMute,
		RepositoryName: repoFullName,
		Timestamp:      time.Now(),
		Metadata: map[string]interface{}{
			"unmute": true,
		},
	}

	// Update the repository subscription to unignore
	sub := &github.Subscription{
		Subscribed: github.Bool(true),
		Ignored:    github.Bool(false),
	}
	_, resp, err := client.GetRawClient().Activity.SetRepositorySubscription(ctx, owner, repo, sub)

	if err != nil {
		action.Success = false
		action.Error = err
		return &ActionResult{
			Action:  action,
			Success: false,
			Error:   err,
		}, fmt.Errorf("failed to unmute repository: %w", err)
	}

	// Check the response status
	if resp != nil && resp.StatusCode >= 400 {
		err := fmt.Errorf("server returned status %d when unmuting repository", resp.StatusCode)
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

// MuteMultipleRepositories mutes notifications for multiple repositories
func MuteMultipleRepositories(ctx context.Context, repoNames []string, opts *BatchOptions) (*BatchResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	// Create a batch processor
	processor := NewBatchProcessor(ctx, opts)

	// Add tasks to the processor
	for _, repoName := range repoNames {
		repoName := repoName // Capture for closure
		processor.AddTask(func() (Action, error) {
			result, err := MuteRepository(ctx, repoName)
			if err != nil {
				return Action{
					Type:           ActionMute,
					RepositoryName: repoName,
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
