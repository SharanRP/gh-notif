package subscriptions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

// Manager handles repository subscription operations
type Manager struct {
	storage Storage
	client  GitHubClient
}

// GitHubClient interface for GitHub operations
type GitHubClient interface {
	GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error)
	ListUserRepositories(ctx context.Context, user string) ([]*github.Repository, error)
	ListOrganizationRepositories(ctx context.Context, org string) ([]*github.Repository, error)
	CheckRepositoryAccess(ctx context.Context, owner, repo string) (bool, error)
}

// NewManager creates a new subscription manager
func NewManager(storage Storage, client GitHubClient) *Manager {
	return &Manager{
		storage: storage,
		client:  client,
	}
}

// Subscribe adds a new repository subscription
func (m *Manager) Subscribe(ctx context.Context, repository string, priority Priority, config SubscriptionConfig) error {
	// Validate the repository pattern
	if err := m.validateRepository(repository); err != nil {
		return fmt.Errorf("invalid repository: %w", err)
	}

	// Check if it's a pattern
	isPattern := strings.Contains(repository, "*")

	// Validate access for non-pattern repositories
	if !isPattern {
		hasAccess, err := m.validateAccess(ctx, repository)
		if err != nil {
			return fmt.Errorf("failed to validate access: %w", err)
		}
		if !hasAccess {
			return fmt.Errorf("no access to repository %s", repository)
		}
	}

	// Create the subscription
	subscription := RepositorySubscription{
		Repository:      repository,
		IsPattern:       isPattern,
		Priority:        priority,
		Config:          config,
		Active:          true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LastAccessCheck: time.Now(),
		HasAccess:       !isPattern, // Patterns don't have direct access
	}

	// Validate the subscription
	if result := m.validateSubscription(subscription); !result.Valid {
		return fmt.Errorf("invalid subscription: %v", result.Errors)
	}

	// Save the subscription
	return m.storage.AddSubscription(subscription)
}

// Unsubscribe removes a repository subscription
func (m *Manager) Unsubscribe(repository string) error {
	return m.storage.RemoveSubscription(repository)
}

// GetSubscription retrieves a specific subscription
func (m *Manager) GetSubscription(repository string) (*RepositorySubscription, error) {
	return m.storage.GetSubscription(repository)
}

// ListSubscriptions returns all subscriptions
func (m *Manager) ListSubscriptions() ([]RepositorySubscription, error) {
	return m.storage.ListSubscriptions()
}

// UpdateSubscription updates an existing subscription
func (m *Manager) UpdateSubscription(repository string, updates SubscriptionUpdates) error {
	subscription, err := m.storage.GetSubscription(repository)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Apply updates
	if updates.Priority != nil {
		subscription.Priority = *updates.Priority
	}
	if updates.Config != nil {
		subscription.Config = *updates.Config
	}
	if updates.Active != nil {
		subscription.Active = *updates.Active
	}

	subscription.UpdatedAt = time.Now()

	// Validate the updated subscription
	if result := m.validateSubscription(*subscription); !result.Valid {
		return fmt.Errorf("invalid subscription update: %v", result.Errors)
	}

	return m.storage.UpdateSubscription(*subscription)
}

// GetStats returns subscription statistics
func (m *Manager) GetStats() (*SubscriptionStats, error) {
	subscriptions, err := m.storage.ListSubscriptions()
	if err != nil {
		return nil, err
	}

	stats := &SubscriptionStats{
		Total:       len(subscriptions),
		LastUpdated: time.Now(),
	}

	for _, sub := range subscriptions {
		if sub.Active {
			stats.Active++
		} else {
			stats.Inactive++
		}

		if sub.IsPattern {
			stats.Patterns++
		} else {
			stats.Repositories++
		}

		switch sub.Priority {
		case PriorityCritical:
			stats.Critical++
		case PriorityNormal:
			stats.Normal++
		case PriorityLow:
			stats.Low++
		}

		if !sub.HasAccess && sub.AccessError != "" {
			stats.AccessErrors++
		}
	}

	return stats, nil
}

// ValidateAccess checks access to all subscribed repositories
func (m *Manager) ValidateAccess(ctx context.Context) error {
	subscriptions, err := m.storage.ListSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		if sub.IsPattern {
			continue // Skip patterns
		}

		hasAccess, err := m.validateAccess(ctx, sub.Repository)
		if err != nil {
			// Update subscription with error
			sub.HasAccess = false
			sub.AccessError = err.Error()
			sub.LastAccessCheck = time.Now()
			m.storage.UpdateSubscription(sub)
			continue
		}

		// Update subscription with access status
		sub.HasAccess = hasAccess
		sub.AccessError = ""
		sub.LastAccessCheck = time.Now()
		m.storage.UpdateSubscription(sub)
	}

	return nil
}

// ExpandPatterns expands wildcard patterns to actual repositories
func (m *Manager) ExpandPatterns(ctx context.Context) ([]string, error) {
	subscriptions, err := m.storage.ListSubscriptions()
	if err != nil {
		return nil, err
	}

	var expanded []string
	for _, sub := range subscriptions {
		if !sub.IsPattern || !sub.Active {
			continue
		}

		repos, err := m.expandPattern(ctx, sub.Repository)
		if err != nil {
			continue // Skip patterns that can't be expanded
		}

		expanded = append(expanded, repos...)
	}

	return expanded, nil
}

// validateRepository validates a repository name or pattern
func (m *Manager) validateRepository(repository string) error {
	if repository == "" {
		return fmt.Errorf("repository cannot be empty")
	}

	// Check for valid pattern
	if strings.Contains(repository, "*") {
		parts := strings.Split(repository, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid pattern format, expected 'owner/*'")
		}
		if parts[1] != "*" {
			return fmt.Errorf("only 'owner/*' patterns are supported")
		}
		return nil
	}

	// Check for valid repository name
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected 'owner/repo'")
	}

	return nil
}

// validateAccess checks if we have access to a repository
func (m *Manager) validateAccess(ctx context.Context, repository string) (bool, error) {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid repository format")
	}

	return m.client.CheckRepositoryAccess(ctx, parts[0], parts[1])
}

// expandPattern expands a wildcard pattern to actual repositories
func (m *Manager) expandPattern(ctx context.Context, pattern string) ([]string, error) {
	parts := strings.Split(pattern, "/")
	if len(parts) != 2 || parts[1] != "*" {
		return nil, fmt.Errorf("invalid pattern")
	}

	org := parts[0]
	repos, err := m.client.ListOrganizationRepositories(ctx, org)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, repo := range repos {
		result = append(result, repo.GetFullName())
	}

	return result, nil
}

// validateSubscription validates a subscription configuration
func (m *Manager) validateSubscription(subscription RepositorySubscription) ValidationResult {
	var errors []ValidationError

	// Validate repository
	if subscription.Repository == "" {
		errors = append(errors, ValidationError{
			Field:   "repository",
			Message: "repository cannot be empty",
		})
	}

	// Validate activity types
	if len(subscription.Config.ActivityTypes) == 0 {
		errors = append(errors, ValidationError{
			Field:   "activity_types",
			Message: "at least one activity type must be specified",
		})
	}

	// Validate branch filter
	if !subscription.Config.BranchFilter.All && len(subscription.Config.BranchFilter.Patterns) == 0 && !subscription.Config.BranchFilter.MainOnly {
		errors = append(errors, ValidationError{
			Field:   "branch_filter",
			Message: "branch filter must specify 'all', 'main_only', or patterns",
		})
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// SubscriptionUpdates represents updates to a subscription
type SubscriptionUpdates struct {
	Priority *Priority            `json:"priority,omitempty"`
	Config   *SubscriptionConfig  `json:"config,omitempty"`
	Active   *bool                `json:"active,omitempty"`
}
