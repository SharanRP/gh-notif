package subscriptions

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/auth"
)

// GitHubClientImpl implements GitHubClient using go-github
type GitHubClientImpl struct {
	client *github.Client
}

// GetRawClient returns the underlying GitHub client
func (gc *GitHubClientImpl) GetRawClient() *github.Client {
	return gc.client
}

// NewGitHubClient creates a new GitHub client for subscriptions
func NewGitHubClient(ctx context.Context) (*GitHubClientImpl, error) {
	// Get authenticated HTTP client
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Create GitHub client
	client := github.NewClient(httpClient)

	return &GitHubClientImpl{
		client: client,
	}, nil
}

// GetRepository retrieves repository information
func (gc *GitHubClientImpl) GetRepository(ctx context.Context, owner, repo string) (*github.Repository, error) {
	repository, _, err := gc.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repo, err)
	}

	return repository, nil
}

// ListUserRepositories lists repositories for a user
func (gc *GitHubClientImpl) ListUserRepositories(ctx context.Context, user string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := gc.client.Repositories.List(ctx, user, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for user %s: %w", user, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// ListOrganizationRepositories lists repositories for an organization
func (gc *GitHubClientImpl) ListOrganizationRepositories(ctx context.Context, org string) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := gc.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for organization %s: %w", org, err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// CheckRepositoryAccess checks if we have access to a repository
func (gc *GitHubClientImpl) CheckRepositoryAccess(ctx context.Context, owner, repo string) (bool, error) {
	// Try to get the repository
	_, resp, err := gc.client.Repositories.Get(ctx, owner, repo)

	if err != nil {
		// Check if it's a 404 (not found) or 403 (forbidden)
		if resp != nil {
			switch resp.StatusCode {
			case 404:
				return false, fmt.Errorf("repository not found or no access")
			case 403:
				return false, fmt.Errorf("access forbidden")
			}
		}
		return false, err
	}

	return true, nil
}

// GetRepositorySubscription gets the current subscription status for a repository
func (gc *GitHubClientImpl) GetRepositorySubscription(ctx context.Context, owner, repo string) (*github.Subscription, error) {
	subscription, _, err := gc.client.Activity.GetRepositorySubscription(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository subscription: %w", err)
	}

	return subscription, nil
}

// SetRepositorySubscription sets the subscription status for a repository
func (gc *GitHubClientImpl) SetRepositorySubscription(ctx context.Context, owner, repo string, subscription *github.Subscription) (*github.Subscription, error) {
	sub, _, err := gc.client.Activity.SetRepositorySubscription(ctx, owner, repo, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to set repository subscription: %w", err)
	}

	return sub, nil
}

// DeleteRepositorySubscription deletes the subscription for a repository
func (gc *GitHubClientImpl) DeleteRepositorySubscription(ctx context.Context, owner, repo string) error {
	_, err := gc.client.Activity.DeleteRepositorySubscription(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to delete repository subscription: %w", err)
	}

	return nil
}

// ListWatchedRepositories lists repositories the authenticated user is watching
func (gc *GitHubClientImpl) ListWatchedRepositories(ctx context.Context) ([]*github.Repository, error) {
	var allRepos []*github.Repository

	opts := &github.ListOptions{PerPage: 100}

	for {
		repos, resp, err := gc.client.Activity.ListWatched(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list watched repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// GetAuthenticatedUser gets information about the authenticated user
func (gc *GitHubClientImpl) GetAuthenticatedUser(ctx context.Context) (*github.User, error) {
	user, _, err := gc.client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	return user, nil
}

// ListUserOrganizations lists organizations for the authenticated user
func (gc *GitHubClientImpl) ListUserOrganizations(ctx context.Context) ([]*github.Organization, error) {
	var allOrgs []*github.Organization

	opts := &github.ListOptions{PerPage: 100}

	for {
		orgs, resp, err := gc.client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list user organizations: %w", err)
		}

		allOrgs = append(allOrgs, orgs...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allOrgs, nil
}

// ValidateRepositoryPattern validates if a repository pattern is accessible
func (gc *GitHubClientImpl) ValidateRepositoryPattern(ctx context.Context, pattern string) ([]string, error) {
	// For now, we'll implement basic organization validation
	// In a full implementation, this would expand the pattern and validate access

	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}

	// This is a simplified implementation
	// A full implementation would parse the pattern and validate access
	return []string{}, nil
}

// GetRepositoryEvents gets recent events for a repository
func (gc *GitHubClientImpl) GetRepositoryEvents(ctx context.Context, owner, repo string) ([]*github.Event, error) {
	opts := &github.ListOptions{PerPage: 30}

	events, _, err := gc.client.Activity.ListRepositoryEvents(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository events: %w", err)
	}

	return events, nil
}

// GetRepositoryNotifications gets notifications for a repository
func (gc *GitHubClientImpl) GetRepositoryNotifications(ctx context.Context, owner, repo string) ([]*github.Notification, error) {
	opts := &github.NotificationListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	notifications, _, err := gc.client.Activity.ListRepositoryNotifications(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository notifications: %w", err)
	}

	return notifications, nil
}
