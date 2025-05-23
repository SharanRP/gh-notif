package discussions

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/user/gh-notif/internal/cache"
)

// SearchEngine provides full-text search capabilities for discussions
type SearchEngine struct {
	client       *Client
	cacheManager *cache.Manager
	index        *SearchIndex
	debug        bool
}

// SearchIndex maintains an in-memory search index for discussions
type SearchIndex struct {
	mu           sync.RWMutex
	discussions  map[string]*IndexedDiscussion
	keywords     map[string][]string // keyword -> discussion IDs
	categories   map[string][]string // category -> discussion IDs
	authors      map[string][]string // author -> discussion IDs
	repositories map[string][]string // repo -> discussion IDs
	lastUpdated  time.Time
}

// IndexedDiscussion represents a discussion in the search index
type IndexedDiscussion struct {
	Discussion *Discussion `json:"discussion"`
	Keywords   []string    `json:"keywords"`
	Score      float64     `json:"score"`
	IndexedAt  time.Time   `json:"indexed_at"`
}

// SearchResult represents a search result
type SearchResult struct {
	Discussion *Discussion `json:"discussion"`
	Score      float64     `json:"score"`
	Highlights []string    `json:"highlights"`
	Rank       int         `json:"rank"`
}

// SearchOptions contains options for search operations
type SearchOptions struct {
	// Search parameters
	Query       string   `json:"query"`
	Repositories []string `json:"repositories,omitempty"`
	Categories  []string `json:"categories,omitempty"`
	Authors     []string `json:"authors,omitempty"`
	
	// Filters
	MinScore    float64   `json:"min_score,omitempty"`
	MaxResults  int       `json:"max_results,omitempty"`
	Since       *time.Time `json:"since,omitempty"`
	Before      *time.Time `json:"before,omitempty"`
	
	// Search behavior
	FuzzyMatch  bool `json:"fuzzy_match,omitempty"`
	CaseSensitive bool `json:"case_sensitive,omitempty"`
	WholeWords  bool `json:"whole_words,omitempty"`
	
	// Performance
	UseCache    bool          `json:"use_cache,omitempty"`
	CacheTTL    time.Duration `json:"cache_ttl,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

// NewSearchEngine creates a new search engine
func NewSearchEngine(client *Client, cacheManager *cache.Manager) *SearchEngine {
	return &SearchEngine{
		client:       client,
		cacheManager: cacheManager,
		index:        NewSearchIndex(),
		debug:        false,
	}
}

// NewSearchIndex creates a new search index
func NewSearchIndex() *SearchIndex {
	return &SearchIndex{
		discussions:  make(map[string]*IndexedDiscussion),
		keywords:     make(map[string][]string),
		categories:   make(map[string][]string),
		authors:      make(map[string][]string),
		repositories: make(map[string][]string),
		lastUpdated:  time.Now(),
	}
}

// Search performs a full-text search across discussions
func (se *SearchEngine) Search(ctx context.Context, options SearchOptions) ([]SearchResult, error) {
	// Set defaults
	if options.MaxResults <= 0 {
		options.MaxResults = 50
	}
	if options.Timeout <= 0 {
		options.Timeout = 30 * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()

	// Check cache first
	if options.UseCache {
		cacheKey := fmt.Sprintf("search_%v", options)
		if cached, found := se.cacheManager.Get(cacheKey); found {
			if results, ok := cached.([]SearchResult); ok {
				if se.debug {
					fmt.Printf("Using cached search results (%d items)\n", len(results))
				}
				return results, nil
			}
		}
	}

	// Ensure index is up to date
	if err := se.updateIndex(ctx, options.Repositories); err != nil {
		return nil, fmt.Errorf("failed to update search index: %w", err)
	}

	// Perform the search
	results := se.performSearch(options)

	// Cache the results
	if options.UseCache && options.CacheTTL > 0 {
		cacheKey := fmt.Sprintf("search_%v", options)
		se.cacheManager.Set(cacheKey, results, options.CacheTTL)
	}

	return results, nil
}

// IndexDiscussions adds discussions to the search index
func (se *SearchEngine) IndexDiscussions(discussions []Discussion) error {
	se.index.mu.Lock()
	defer se.index.mu.Unlock()

	for _, discussion := range discussions {
		indexed := &IndexedDiscussion{
			Discussion: &discussion,
			Keywords:   se.extractKeywords(discussion),
			Score:      se.calculateRelevanceScore(discussion),
			IndexedAt:  time.Now(),
		}

		// Add to main index
		se.index.discussions[discussion.ID] = indexed

		// Add to keyword index
		for _, keyword := range indexed.Keywords {
			se.index.keywords[keyword] = append(se.index.keywords[keyword], discussion.ID)
		}

		// Add to category index
		categoryKey := strings.ToLower(discussion.Category.Slug)
		se.index.categories[categoryKey] = append(se.index.categories[categoryKey], discussion.ID)

		// Add to author index
		authorKey := strings.ToLower(discussion.Author.Login)
		se.index.authors[authorKey] = append(se.index.authors[authorKey], discussion.ID)

		// Add to repository index
		repoKey := strings.ToLower(discussion.Repository.FullName)
		se.index.repositories[repoKey] = append(se.index.repositories[repoKey], discussion.ID)
	}

	se.index.lastUpdated = time.Now()
	return nil
}

// ClearIndex clears the search index
func (se *SearchEngine) ClearIndex() {
	se.index.mu.Lock()
	defer se.index.mu.Unlock()

	se.index.discussions = make(map[string]*IndexedDiscussion)
	se.index.keywords = make(map[string][]string)
	se.index.categories = make(map[string][]string)
	se.index.authors = make(map[string][]string)
	se.index.repositories = make(map[string][]string)
	se.index.lastUpdated = time.Now()
}

// GetIndexStats returns statistics about the search index
func (se *SearchEngine) GetIndexStats() map[string]interface{} {
	se.index.mu.RLock()
	defer se.index.mu.RUnlock()

	return map[string]interface{}{
		"discussions":  len(se.index.discussions),
		"keywords":     len(se.index.keywords),
		"categories":   len(se.index.categories),
		"authors":      len(se.index.authors),
		"repositories": len(se.index.repositories),
		"last_updated": se.index.lastUpdated,
	}
}

// updateIndex updates the search index with latest discussions
func (se *SearchEngine) updateIndex(ctx context.Context, repositories []string) error {
	// Check if index needs updating (every 5 minutes)
	if time.Since(se.index.lastUpdated) < 5*time.Minute {
		return nil
	}

	// Fetch recent discussions
	filter := DiscussionFilter{
		State: "all",
		Sort:  "updated",
		Limit: 1000, // Reasonable limit for indexing
	}

	options := DiscussionOptions{
		UseCache: true,
		CacheTTL: 5 * time.Minute,
	}

	discussions, err := se.client.GetDiscussions(ctx, repositories, filter, options)
	if err != nil {
		return fmt.Errorf("failed to fetch discussions for indexing: %w", err)
	}

	// Update the index
	return se.IndexDiscussions(discussions)
}

// performSearch executes the actual search against the index
func (se *SearchEngine) performSearch(options SearchOptions) []SearchResult {
	se.index.mu.RLock()
	defer se.index.mu.RUnlock()

	var candidateIDs []string
	query := strings.ToLower(options.Query)

	// Find candidate discussions based on keywords
	if options.Query != "" {
		keywords := se.tokenizeQuery(query)
		candidateMap := make(map[string]bool)

		for _, keyword := range keywords {
			if discussionIDs, exists := se.index.keywords[keyword]; exists {
				for _, id := range discussionIDs {
					candidateMap[id] = true
				}
			}
		}

		for id := range candidateMap {
			candidateIDs = append(candidateIDs, id)
		}
	} else {
		// No query, get all discussions
		for id := range se.index.discussions {
			candidateIDs = append(candidateIDs, id)
		}
	}

	// Apply filters and calculate scores
	var results []SearchResult
	for _, id := range candidateIDs {
		indexed, exists := se.index.discussions[id]
		if !exists {
			continue
		}

		discussion := indexed.Discussion

		// Apply repository filter
		if len(options.Repositories) > 0 {
			found := false
			for _, repo := range options.Repositories {
				if strings.EqualFold(discussion.Repository.FullName, repo) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply category filter
		if len(options.Categories) > 0 {
			found := false
			for _, category := range options.Categories {
				if strings.EqualFold(discussion.Category.Name, category) ||
				   strings.EqualFold(discussion.Category.Slug, category) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply author filter
		if len(options.Authors) > 0 {
			found := false
			for _, author := range options.Authors {
				if strings.EqualFold(discussion.Author.Login, author) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Apply time filters
		if options.Since != nil && discussion.CreatedAt.Before(*options.Since) {
			continue
		}
		if options.Before != nil && discussion.CreatedAt.After(*options.Before) {
			continue
		}

		// Calculate search score
		score := se.calculateSearchScore(discussion, options.Query)
		if score < options.MinScore {
			continue
		}

		// Generate highlights
		highlights := se.generateHighlights(discussion, options.Query)

		results = append(results, SearchResult{
			Discussion: discussion,
			Score:      score,
			Highlights: highlights,
		})
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Add ranks
	for i := range results {
		results[i].Rank = i + 1
	}

	// Apply limit
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results
}

// extractKeywords extracts searchable keywords from a discussion
func (se *SearchEngine) extractKeywords(discussion Discussion) []string {
	var keywords []string

	// Extract from title
	titleWords := se.tokenize(discussion.Title)
	keywords = append(keywords, titleWords...)

	// Extract from body
	bodyWords := se.tokenize(discussion.Body)
	keywords = append(keywords, bodyWords...)

	// Add category
	keywords = append(keywords, strings.ToLower(discussion.Category.Name))
	keywords = append(keywords, strings.ToLower(discussion.Category.Slug))

	// Add labels
	for _, label := range discussion.Labels {
		keywords = append(keywords, strings.ToLower(label.Name))
	}

	// Remove duplicates and filter
	return se.deduplicateAndFilter(keywords)
}

// tokenize splits text into searchable tokens
func (se *SearchEngine) tokenize(text string) []string {
	// Simple tokenization - split on whitespace and punctuation
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	})

	// Filter out short words and common stop words
	var filtered []string
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
	}

	for _, word := range words {
		if len(word) > 2 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// tokenizeQuery tokenizes a search query
func (se *SearchEngine) tokenizeQuery(query string) []string {
	return se.tokenize(query)
}

// deduplicateAndFilter removes duplicates and filters keywords
func (se *SearchEngine) deduplicateAndFilter(keywords []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, keyword := range keywords {
		if !seen[keyword] && len(keyword) > 2 {
			seen[keyword] = true
			result = append(result, keyword)
		}
	}

	return result
}

// calculateRelevanceScore calculates a base relevance score for a discussion
func (se *SearchEngine) calculateRelevanceScore(discussion Discussion) float64 {
	score := 0.0

	// Age factor (newer is better)
	age := time.Since(discussion.CreatedAt)
	ageFactor := 1.0 / (1.0 + age.Hours()/24.0/7.0) // Decay over weeks
	score += ageFactor * 0.2

	// Engagement factor
	engagementScore := float64(discussion.UpvoteCount)*0.3 + 
					  float64(discussion.CommentCount)*0.5 + 
					  float64(discussion.ReactionCount)*0.2
	score += engagementScore * 0.3

	// State factor (open discussions are more relevant)
	if discussion.State == "OPEN" {
		score += 0.2
	}

	// Answer factor (answered questions are less urgent)
	if discussion.Answer != nil {
		score += 0.1
	} else if discussion.Category.IsAnswerable {
		score += 0.2 // Unanswered questions are more relevant
	}

	// Category factor (Q&A might be more important)
	if strings.Contains(strings.ToLower(discussion.Category.Name), "q&a") ||
	   strings.Contains(strings.ToLower(discussion.Category.Name), "question") {
		score += 0.2
	}

	return score
}

// calculateSearchScore calculates a search-specific score
func (se *SearchEngine) calculateSearchScore(discussion *Discussion, query string) float64 {
	if query == "" {
		return se.calculateRelevanceScore(*discussion)
	}

	score := 0.0
	queryLower := strings.ToLower(query)
	queryWords := se.tokenizeQuery(queryLower)

	// Title match (highest weight)
	titleLower := strings.ToLower(discussion.Title)
	for _, word := range queryWords {
		if strings.Contains(titleLower, word) {
			score += 0.4
		}
	}

	// Body match
	bodyLower := strings.ToLower(discussion.Body)
	for _, word := range queryWords {
		if strings.Contains(bodyLower, word) {
			score += 0.2
		}
	}

	// Category match
	categoryLower := strings.ToLower(discussion.Category.Name)
	for _, word := range queryWords {
		if strings.Contains(categoryLower, word) {
			score += 0.3
		}
	}

	// Label match
	for _, label := range discussion.Labels {
		labelLower := strings.ToLower(label.Name)
		for _, word := range queryWords {
			if strings.Contains(labelLower, word) {
				score += 0.2
			}
		}
	}

	// Add base relevance score
	score += se.calculateRelevanceScore(*discussion) * 0.3

	return score
}

// generateHighlights generates highlighted snippets for search results
func (se *SearchEngine) generateHighlights(discussion *Discussion, query string) []string {
	if query == "" {
		return []string{}
	}

	var highlights []string
	queryWords := se.tokenizeQuery(strings.ToLower(query))

	// Check title
	if se.containsAnyWord(discussion.Title, queryWords) {
		highlights = append(highlights, "Title: "+discussion.Title)
	}

	// Check body (first 200 chars)
	body := discussion.Body
	if len(body) > 200 {
		body = body[:200] + "..."
	}
	if se.containsAnyWord(body, queryWords) {
		highlights = append(highlights, "Body: "+body)
	}

	// Check category
	if se.containsAnyWord(discussion.Category.Name, queryWords) {
		highlights = append(highlights, "Category: "+discussion.Category.Name)
	}

	return highlights
}

// containsAnyWord checks if text contains any of the given words
func (se *SearchEngine) containsAnyWord(text string, words []string) bool {
	textLower := strings.ToLower(text)
	for _, word := range words {
		if strings.Contains(textLower, word) {
			return true
		}
	}
	return false
}
