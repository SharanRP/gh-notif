package github

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/cache"
	"github.com/SharanRP/gh-notif/internal/config"
)

// CacheManager manages the GitHub API cache
type CacheManager struct {
	// Manager is the cache manager
	Manager *cache.Manager

	// Client is the GitHub client
	Client *Client

	// Config is the configuration
	Config *config.Config
}

// NewCacheManager creates a new cache manager
func NewCacheManager(client *Client, cfg *config.Config) (*CacheManager, error) {
	// Create cache options
	cacheOpts := &cache.Options{
		CacheDir:          ".gh-notif-cache", // Use a local directory for testing
		DefaultTTL:        time.Duration(cfg.Advanced.CacheTTL) * time.Second,
		MaxSize:           1 * 1024 * 1024 * 1024, // 1GB
		MemoryLimit:       100 * 1024 * 1024,      // 100MB
		EnablePrefetching: true,
		EnableCompression: true,
	}

	// Create the cache
	cacheImpl, err := cache.NewCache(cache.BadgerCacheType, cacheOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Create manager options
	managerOpts := &cache.ManagerOptions{
		PrefetchConcurrency: cfg.Advanced.MaxConcurrent,
		EnablePrefetching:   true,
		EnableInvalidation:  true,
		DefaultTTL:          time.Duration(cfg.Advanced.CacheTTL) * time.Second,
		RefreshBeforeExpiry: 5 * time.Minute,
	}

	// Create the manager
	manager := cache.NewManager(cacheImpl, managerOpts)

	// Add invalidation patterns
	manager.AddInvalidationPattern(cache.InvalidationPattern{
		Pattern:   "notifications_*",
		Action:    cache.InvalidateDelete,
		Condition: cache.OnWrite,
	})

	return &CacheManager{
		Manager: manager,
		Client:  client,
		Config:  cfg,
	}, nil
}

// Close closes the cache manager
func (cm *CacheManager) Close() error {
	return cm.Manager.Close()
}

// GetNotifications gets notifications from cache or API
func (cm *CacheManager) GetNotifications(ctx context.Context, opts *github.NotificationListOptions) ([]*github.Notification, *github.Response, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("notifications_%v_%v_%v_%v_%d",
		opts.All, opts.Participating, opts.Since, opts.Before, opts.PerPage)

	// Try to get from cache
	if cached, found := cm.Manager.Get(cacheKey); found {
		if notifications, ok := cached.([]*github.Notification); ok {
			// Create a mock response
			resp := &github.Response{
				NextPage: 0,
				LastPage: 0,
				Rate: github.Rate{
					Limit:     5000,
					Remaining: 5000,
					Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Hour)},
				},
			}
			return notifications, resp, nil
		}
	}

	// Not in cache, fetch from API
	notifications, resp, err := cm.Client.client.Activity.ListNotifications(ctx, opts)
	if err != nil {
		return nil, resp, err
	}

	// Cache the result
	cm.Manager.Set(cacheKey, notifications, time.Duration(cm.Config.Advanced.CacheTTL)*time.Second)

	return notifications, resp, nil
}

// GetRepositoryNotifications gets repository notifications from cache or API
func (cm *CacheManager) GetRepositoryNotifications(ctx context.Context, owner, repo string, opts *github.NotificationListOptions) ([]*github.Notification, *github.Response, error) {
	// Create cache key
	cacheKey := fmt.Sprintf("repo_notifications_%s_%s_%v_%v_%v_%v_%d",
		owner, repo, opts.All, opts.Participating, opts.Since, opts.Before, opts.PerPage)

	// Try to get from cache
	if cached, found := cm.Manager.Get(cacheKey); found {
		if notifications, ok := cached.([]*github.Notification); ok {
			// Create a mock response
			resp := &github.Response{
				NextPage: 0,
				LastPage: 0,
				Rate: github.Rate{
					Limit:     5000,
					Remaining: 5000,
					Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Hour)},
				},
			}
			return notifications, resp, nil
		}
	}

	// Not in cache, fetch from API
	notifications, resp, err := cm.Client.client.Activity.ListRepositoryNotifications(ctx, owner, repo, opts)
	if err != nil {
		return nil, resp, err
	}

	// Cache the result
	cm.Manager.Set(cacheKey, notifications, time.Duration(cm.Config.Advanced.CacheTTL)*time.Second)

	return notifications, resp, nil
}

// InvalidateNotificationsCache invalidates the notifications cache
func (cm *CacheManager) InvalidateNotificationsCache() {
	// Delete all notification cache entries
	// In a real implementation, we would use a pattern match
	cm.Manager.Delete("notifications_*")
}

// PrefetchNotificationDetails prefetches notification details
func (cm *CacheManager) PrefetchNotificationDetails(notifications []*github.Notification) {
	for _, notification := range notifications {
		subjectType := notification.GetSubject().GetType()
		subjectURL := notification.GetSubject().GetURL()

		if subjectURL == "" {
			continue
		}

		// Create a prefetch request based on the subject type
		var cacheKey string
		switch subjectType {
		case "Issue":
			cacheKey = fmt.Sprintf("issue_%s_%s_%s",
				notification.GetRepository().GetOwner().GetLogin(),
				notification.GetRepository().GetName(),
				filepath.Base(subjectURL))
		case "PullRequest":
			cacheKey = fmt.Sprintf("pr_%s_%s_%s",
				notification.GetRepository().GetOwner().GetLogin(),
				notification.GetRepository().GetName(),
				filepath.Base(subjectURL))
		default:
			continue
		}

		// Queue the prefetch request
		cm.Manager.Prefetch(cache.PrefetchRequest{
			Key:      cacheKey,
			Priority: 1,
			Callback: func(ctx context.Context) (interface{}, error) {
				// This would fetch the actual data
				// For now, just return nil
				return nil, nil
			},
		})
	}
}
