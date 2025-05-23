package discussions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/cache"
	"github.com/SharanRP/gh-notif/internal/config"
)

// Manager provides a high-level interface for discussion operations
type Manager struct {
	client        *Client
	analytics     *AnalyticsEngine
	searchEngine  *SearchEngine
	watcher       *DiscussionWatcher
	configManager *config.ConfigManager
	cacheManager  *cache.Manager

	// State management
	mu            sync.RWMutex
	isInitialized bool
	repositories  []string

	// Configuration
	defaultOptions DiscussionOptions
	debug          bool
}

// ManagerOptions contains configuration for the discussion manager
type ManagerOptions struct {
	// Default repositories to monitor
	Repositories []string `json:"repositories"`

	// Default options for operations
	DefaultOptions DiscussionOptions `json:"default_options"`

	// Enable debug logging
	Debug bool `json:"debug"`

	// Cache configuration
	CacheEnabled bool          `json:"cache_enabled"`
	CacheTTL     time.Duration `json:"cache_ttl"`

	// Performance settings
	MaxConcurrency int           `json:"max_concurrency"`
	Timeout        time.Duration `json:"timeout"`
}

// NewManager creates a new discussion manager
func NewManager(ctx context.Context, options ManagerOptions) (*Manager, error) {
	// Create the client
	client, err := NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create discussions client: %w", err)
	}

	// Get config and cache managers
	configManager := config.NewConfigManager()
	if err := configManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	config := configManager.GetConfig()

	// Create cache manager
	cacheOpts := &cache.Options{
		CacheDir:   config.Advanced.CacheDir,
		DefaultTTL: time.Duration(config.Advanced.CacheTTL) * time.Second,
	}
	cacheImpl, err := cache.NewCache(cache.MemoryCacheType, cacheOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	managerOpts := &cache.ManagerOptions{
		DefaultTTL: time.Duration(config.Advanced.CacheTTL) * time.Second,
	}
	cacheManager := cache.NewManager(cacheImpl, managerOpts)

	// Create analytics engine
	analytics := NewAnalyticsEngine(client)

	// Create search engine
	searchEngine := NewSearchEngine(client, cacheManager)

	// Create watcher
	watcher := NewDiscussionWatcher(client, analytics, searchEngine)

	// Set default options
	defaultOptions := options.DefaultOptions
	if defaultOptions.UseCache && defaultOptions.CacheTTL == 0 {
		defaultOptions.CacheTTL = options.CacheTTL
	}
	if defaultOptions.Concurrency == 0 {
		defaultOptions.Concurrency = options.MaxConcurrency
	}
	if defaultOptions.Timeout == 0 {
		defaultOptions.Timeout = options.Timeout
	}

	return &Manager{
		client:         client,
		analytics:      analytics,
		searchEngine:   searchEngine,
		watcher:        watcher,
		configManager:  configManager,
		cacheManager:   cacheManager,
		repositories:   options.Repositories,
		defaultOptions: defaultOptions,
		debug:          options.Debug,
	}, nil
}

// Initialize initializes the discussion manager
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isInitialized {
		return nil
	}

	// Initialize search index if we have repositories
	if len(m.repositories) > 0 {
		if m.debug {
			fmt.Printf("Initializing discussion manager with %d repositories\n", len(m.repositories))
		}

		// Fetch initial discussions for indexing
		filter := DiscussionFilter{
			State: "all",
			Limit: 500, // Reasonable initial limit
		}

		discussions, err := m.client.GetDiscussions(ctx, m.repositories, filter, m.defaultOptions)
		if err != nil {
			return fmt.Errorf("failed to fetch initial discussions: %w", err)
		}

		// Index the discussions
		if err := m.searchEngine.IndexDiscussions(discussions); err != nil {
			return fmt.Errorf("failed to index discussions: %w", err)
		}

		if m.debug {
			fmt.Printf("Indexed %d discussions\n", len(discussions))
		}
	}

	m.isInitialized = true
	return nil
}

// GetDiscussions retrieves discussions with the default options
func (m *Manager) GetDiscussions(ctx context.Context, filter DiscussionFilter) ([]Discussion, error) {
	repositories := m.repositories
	if filter.Repository != "" {
		repositories = []string{filter.Repository}
	}

	return m.client.GetDiscussions(ctx, repositories, filter, m.defaultOptions)
}

// SearchDiscussions searches discussions using the search engine
func (m *Manager) SearchDiscussions(ctx context.Context, query string, options SearchOptions) ([]SearchResult, error) {
	// Use default repositories if none specified
	if len(options.Repositories) == 0 {
		options.Repositories = m.repositories
	}

	// Set default options
	if options.MaxResults == 0 {
		options.MaxResults = 50
	}
	if options.Timeout == 0 {
		options.Timeout = m.defaultOptions.Timeout
	}
	if !options.UseCache {
		options.UseCache = m.defaultOptions.UseCache
		options.CacheTTL = m.defaultOptions.CacheTTL
	}

	options.Query = query
	return m.searchEngine.Search(ctx, options)
}

// GetTrendingDiscussions gets trending discussions
func (m *Manager) GetTrendingDiscussions(ctx context.Context, timeRange TimeRange, limit int) ([]Discussion, error) {
	return m.analytics.GetTrendingDiscussions(ctx, m.repositories, timeRange, limit)
}

// GetUnansweredQuestions gets unanswered questions
func (m *Manager) GetUnansweredQuestions(ctx context.Context, maxAge time.Duration) ([]Discussion, error) {
	return m.analytics.GetUnansweredQuestions(ctx, m.repositories, maxAge)
}

// GenerateAnalytics generates analytics for discussions
func (m *Manager) GenerateAnalytics(ctx context.Context, timeRange TimeRange) (*DiscussionAnalytics, error) {
	return m.analytics.GenerateAnalytics(ctx, m.repositories, timeRange)
}

// StartWatching starts watching for discussion changes
func (m *Manager) StartWatching(ctx context.Context, options WatcherOptions) error {
	// Use default repositories if none specified
	if len(options.Repositories) == 0 {
		options.Repositories = m.repositories
	}

	// Set default options
	if !options.Options.UseCache {
		options.Options = m.defaultOptions
	}

	return m.watcher.Start(ctx, options)
}

// StopWatching stops watching for discussion changes
func (m *Manager) StopWatching() {
	m.watcher.Stop()
}

// IsWatching returns whether the manager is currently watching
func (m *Manager) IsWatching() bool {
	return m.watcher.IsWatching()
}

// GetWatcherEvents returns the watcher event channel
func (m *Manager) GetWatcherEvents() <-chan DiscussionEvent {
	return m.watcher.GetEventChannel()
}

// GetWatcherErrors returns the watcher error channel
func (m *Manager) GetWatcherErrors() <-chan error {
	return m.watcher.GetErrorChannel()
}

// AddRepository adds a repository to monitor
func (m *Manager) AddRepository(repository string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already exists
	for _, repo := range m.repositories {
		if repo == repository {
			return
		}
	}

	m.repositories = append(m.repositories, repository)

	if m.debug {
		fmt.Printf("Added repository: %s\n", repository)
	}
}

// RemoveRepository removes a repository from monitoring
func (m *Manager) RemoveRepository(repository string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, repo := range m.repositories {
		if repo == repository {
			m.repositories = append(m.repositories[:i], m.repositories[i+1:]...)
			if m.debug {
				fmt.Printf("Removed repository: %s\n", repository)
			}
			return
		}
	}
}

// GetRepositories returns the list of monitored repositories
func (m *Manager) GetRepositories() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]string, len(m.repositories))
	copy(result, m.repositories)
	return result
}

// GetStats returns statistics about the manager
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"initialized":   m.isInitialized,
		"repositories":  len(m.repositories),
		"is_watching":   m.watcher.IsWatching(),
		"search_index":  m.searchEngine.GetIndexStats(),
		"watcher_stats": m.watcher.GetStats(),
	}

	return stats
}

// RefreshIndex refreshes the search index
func (m *Manager) RefreshIndex(ctx context.Context) error {
	if len(m.repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	// Clear existing index
	m.searchEngine.ClearIndex()

	// Fetch fresh discussions
	filter := DiscussionFilter{
		State: "all",
		Limit: 1000,
	}

	discussions, err := m.client.GetDiscussions(ctx, m.repositories, filter, m.defaultOptions)
	if err != nil {
		return fmt.Errorf("failed to fetch discussions for index refresh: %w", err)
	}

	// Re-index
	if err := m.searchEngine.IndexDiscussions(discussions); err != nil {
		return fmt.Errorf("failed to re-index discussions: %w", err)
	}

	if m.debug {
		fmt.Printf("Refreshed index with %d discussions\n", len(discussions))
	}

	return nil
}

// Close closes the manager and cleans up resources
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop watching if active
	if m.watcher.IsWatching() {
		m.watcher.Stop()
	}

	// Clear search index
	m.searchEngine.ClearIndex()

	m.isInitialized = false

	if m.debug {
		fmt.Println("Discussion manager closed")
	}

	return nil
}

// DefaultManagerOptions returns default options for the manager
func DefaultManagerOptions() ManagerOptions {
	return ManagerOptions{
		Repositories: []string{},
		DefaultOptions: DiscussionOptions{
			UseCache:        true,
			CacheTTL:        5 * time.Minute,
			IncludeComments: false,
			Concurrency:     5,
			Timeout:         30 * time.Second,
		},
		Debug:          false,
		CacheEnabled:   true,
		CacheTTL:       5 * time.Minute,
		MaxConcurrency: 5,
		Timeout:        30 * time.Second,
	}
}
