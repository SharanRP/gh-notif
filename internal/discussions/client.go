package discussions

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/cache"
	"github.com/SharanRP/gh-notif/internal/config"
)

// Client handles GitHub Discussions operations
type Client struct {
	graphqlClient *GraphQLClient
	cacheManager  *cache.Manager
	configManager *config.ConfigManager
	debug         bool

	// Performance optimization
	maxConcurrent int
	timeout       time.Duration
	cacheTTL      time.Duration
}

// NewClient creates a new discussions client
func NewClient(ctx context.Context) (*Client, error) {
	// Create GraphQL client
	graphqlClient, err := NewGraphQLClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GraphQL client: %w", err)
	}

	// Get config manager
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

	return &Client{
		graphqlClient: graphqlClient,
		cacheManager:  cacheManager,
		configManager: configManager,
		debug:         config.Advanced.Debug,
		maxConcurrent: config.Advanced.MaxConcurrent,
		timeout:       time.Duration(config.API.Timeout) * time.Second,
		cacheTTL:      time.Duration(config.Advanced.CacheTTL) * time.Second,
	}, nil
}

// GetDiscussions fetches discussions with filtering and caching
func (c *Client) GetDiscussions(ctx context.Context, repositories []string, filter DiscussionFilter, options DiscussionOptions) ([]Discussion, error) {
	// Set default options
	if options.Concurrency <= 0 {
		options.Concurrency = c.maxConcurrent
	}
	if options.Timeout <= 0 {
		options.Timeout = c.timeout
	}
	if options.CacheTTL <= 0 {
		options.CacheTTL = c.cacheTTL
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Process repositories concurrently
	type result struct {
		discussions []Discussion
		err         error
	}

	resultChan := make(chan result, len(repositories))
	semaphore := make(chan struct{}, options.Concurrency)

	var wg sync.WaitGroup
	for _, repo := range repositories {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			discussions, err := c.getRepositoryDiscussions(ctx, repo, filter, options)
			resultChan <- result{discussions: discussions, err: err}
		}(repo)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var allDiscussions []Discussion
	var errors []error

	for res := range resultChan {
		if res.err != nil {
			errors = append(errors, res.err)
			continue
		}
		allDiscussions = append(allDiscussions, res.discussions...)
	}

	// Return error if any occurred
	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to fetch discussions: %v", errors)
	}

	// Apply additional filtering
	filteredDiscussions := c.applyClientSideFilters(allDiscussions, filter)

	// Sort discussions
	c.sortDiscussions(filteredDiscussions, filter)

	// Apply limit
	if filter.Limit > 0 && len(filteredDiscussions) > filter.Limit {
		filteredDiscussions = filteredDiscussions[:filter.Limit]
	}

	return filteredDiscussions, nil
}

// getRepositoryDiscussions fetches discussions for a single repository
func (c *Client) getRepositoryDiscussions(ctx context.Context, repo string, filter DiscussionFilter, options DiscussionOptions) ([]Discussion, error) {
	// Parse repository
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Check cache if enabled
	if options.UseCache {
		cacheKey := fmt.Sprintf("discussions_%s_%s_%v", owner, repoName, filter)
		if cached, found := c.cacheManager.Get(cacheKey); found {
			if discussions, ok := cached.([]Discussion); ok {
				if c.debug {
					fmt.Printf("Using cached discussions for %s (%d items)\n", repo, len(discussions))
				}
				return discussions, nil
			}
		}
	}

	// Fetch from API
	discussions, err := c.graphqlClient.GetDiscussions(ctx, owner, repoName, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discussions for %s: %w", repo, err)
	}

	// Cache the result
	if options.UseCache {
		cacheKey := fmt.Sprintf("discussions_%s_%s_%v", owner, repoName, filter)
		c.cacheManager.Set(cacheKey, discussions, options.CacheTTL)
	}

	if c.debug {
		fmt.Printf("Fetched %d discussions from %s\n", len(discussions), repo)
	}

	return discussions, nil
}

// GetDiscussion fetches a single discussion with comments
func (c *Client) GetDiscussion(ctx context.Context, repo string, number int, options DiscussionOptions) (*Discussion, error) {
	// Parse repository
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Check cache if enabled
	cacheKey := fmt.Sprintf("discussion_%s_%s_%d", owner, repoName, number)
	if options.UseCache {
		if cached, found := c.cacheManager.Get(cacheKey); found {
			if discussion, ok := cached.(*Discussion); ok {
				if c.debug {
					fmt.Printf("Using cached discussion %s#%d\n", repo, number)
				}
				return discussion, nil
			}
		}
	}

	// Fetch from API using GraphQL
	discussion, err := c.getDiscussionByNumber(ctx, owner, repoName, number, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discussion %s#%d: %w", repo, number, err)
	}

	// Cache the result
	if options.UseCache {
		c.cacheManager.Set(cacheKey, discussion, options.CacheTTL)
	}

	return discussion, nil
}

// GetDiscussionComments fetches comments for a discussion
func (c *Client) GetDiscussionComments(ctx context.Context, discussionID string, options DiscussionOptions) ([]Comment, error) {
	// Check cache if enabled
	cacheKey := fmt.Sprintf("discussion_comments_%s", discussionID)
	if options.UseCache {
		if cached, found := c.cacheManager.Get(cacheKey); found {
			if comments, ok := cached.([]Comment); ok {
				if c.debug {
					fmt.Printf("Using cached comments for discussion %s (%d items)\n", discussionID, len(comments))
				}
				return comments, nil
			}
		}
	}

	// Fetch from API
	comments, err := c.getDiscussionCommentsFromAPI(ctx, discussionID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments for discussion %s: %w", discussionID, err)
	}

	// Cache the result
	if options.UseCache {
		c.cacheManager.Set(cacheKey, comments, options.CacheTTL)
	}

	return comments, nil
}

// SearchDiscussions searches discussions across repositories
func (c *Client) SearchDiscussions(ctx context.Context, query string, repositories []string, filter DiscussionFilter, options DiscussionOptions) ([]Discussion, error) {
	// Get all discussions first
	allDiscussions, err := c.GetDiscussions(ctx, repositories, filter, options)
	if err != nil {
		return nil, err
	}

	// Perform full-text search
	var matchingDiscussions []Discussion
	queryLower := strings.ToLower(query)

	for _, discussion := range allDiscussions {
		// Search in title, body, and comments
		if c.matchesQuery(discussion, queryLower) {
			matchingDiscussions = append(matchingDiscussions, discussion)
		}
	}

	return matchingDiscussions, nil
}

// GetDiscussionCategories fetches available discussion categories for a repository
func (c *Client) GetDiscussionCategories(ctx context.Context, repo string) ([]Category, error) {
	// Parse repository
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s", repo)
	}
	owner, repoName := parts[0], parts[1]

	// Check cache
	cacheKey := fmt.Sprintf("discussion_categories_%s_%s", owner, repoName)
	if cached, found := c.cacheManager.Get(cacheKey); found {
		if categories, ok := cached.([]Category); ok {
			return categories, nil
		}
	}

	// Fetch from API
	categories, err := c.getDiscussionCategoriesFromAPI(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discussion categories for %s: %w", repo, err)
	}

	// Cache the result
	c.cacheManager.Set(cacheKey, categories, c.cacheTTL)

	return categories, nil
}

// Helper methods

// applyClientSideFilters applies filters that couldn't be applied server-side
func (c *Client) applyClientSideFilters(discussions []Discussion, filter DiscussionFilter) []Discussion {
	var filtered []Discussion

	for _, discussion := range discussions {
		// Apply author filter
		if filter.Author != "" && discussion.Author.Login != filter.Author {
			continue
		}

		// Apply time filters
		if filter.CreatedAfter != nil && discussion.CreatedAt.Before(*filter.CreatedAfter) {
			continue
		}
		if filter.CreatedBefore != nil && discussion.CreatedAt.After(*filter.CreatedBefore) {
			continue
		}
		if filter.UpdatedAfter != nil && discussion.UpdatedAt.Before(*filter.UpdatedAfter) {
			continue
		}
		if filter.UpdatedBefore != nil && discussion.UpdatedAt.After(*filter.UpdatedBefore) {
			continue
		}

		// Apply engagement filters
		if filter.MinUpvotes > 0 && discussion.UpvoteCount < filter.MinUpvotes {
			continue
		}
		if filter.MinComments > 0 && discussion.CommentCount < filter.MinComments {
			continue
		}

		// Apply answer filters
		if filter.Answered != nil {
			hasAnswer := discussion.Answer != nil
			if *filter.Answered != hasAnswer {
				continue
			}
		}

		// Apply participation filters
		if filter.Participating && !discussion.ViewerDidAuthor {
			// TODO: Check if user participated in comments
			continue
		}

		// Apply content filters
		if filter.Query != "" {
			queryLower := strings.ToLower(filter.Query)
			if !c.matchesQuery(discussion, queryLower) {
				continue
			}
		}

		filtered = append(filtered, discussion)
	}

	return filtered
}

// sortDiscussions sorts discussions based on the filter criteria
func (c *Client) sortDiscussions(discussions []Discussion, filter DiscussionFilter) {
	// Implementation would depend on the sorting requirements
	// For now, we'll keep the default order from the API
}

// matchesQuery checks if a discussion matches the search query
func (c *Client) matchesQuery(discussion Discussion, queryLower string) bool {
	// Search in title
	if strings.Contains(strings.ToLower(discussion.Title), queryLower) {
		return true
	}

	// Search in body
	if strings.Contains(strings.ToLower(discussion.Body), queryLower) {
		return true
	}

	// Search in category
	if strings.Contains(strings.ToLower(discussion.Category.Name), queryLower) {
		return true
	}

	// Search in labels
	for _, label := range discussion.Labels {
		if strings.Contains(strings.ToLower(label.Name), queryLower) {
			return true
		}
	}

	return false
}

// Placeholder methods for API calls that need to be implemented
func (c *Client) getDiscussionByNumber(ctx context.Context, owner, repo string, number int, options DiscussionOptions) (*Discussion, error) {
	// TODO: Implement GraphQL query for single discussion
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) getDiscussionCommentsFromAPI(ctx context.Context, discussionID string, options DiscussionOptions) ([]Comment, error) {
	// TODO: Implement GraphQL query for discussion comments
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) getDiscussionCategoriesFromAPI(ctx context.Context, owner, repo string) ([]Category, error) {
	// TODO: Implement GraphQL query for discussion categories
	return nil, fmt.Errorf("not implemented")
}
