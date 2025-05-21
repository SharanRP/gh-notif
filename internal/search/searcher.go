package search

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

// SearchOptions contains options for searching notifications
type SearchOptions struct {
	// Query is the search query
	Query string
	// Fields are the fields to search
	Fields []string
	// CaseSensitive determines whether the search is case-sensitive
	CaseSensitive bool
	// UseRegex determines whether to use regex matching
	UseRegex bool
	// MaxResults is the maximum number of results to return
	MaxResults int
	// Concurrency is the number of goroutines to use
	Concurrency int
	// Timeout is the maximum time to spend searching
	Timeout time.Duration
	// HighlightMatches determines whether to highlight matches
	HighlightMatches bool
	// HighlightPrefix is the prefix for highlighted text
	HighlightPrefix string
	// HighlightSuffix is the suffix for highlighted text
	HighlightSuffix string
}

// DefaultSearchOptions returns the default search options
func DefaultSearchOptions() *SearchOptions {
	return &SearchOptions{
		Fields:          []string{"title", "repository", "type", "reason"},
		CaseSensitive:   false,
		UseRegex:        false,
		MaxResults:      100,
		Concurrency:     5,
		Timeout:         5 * time.Second,
		HighlightMatches: true,
		HighlightPrefix:  "\033[1;33m", // Yellow bold
		HighlightSuffix:  "\033[0m",    // Reset
	}
}

// SearchResult represents a search result
type SearchResult struct {
	// Notification is the matching notification
	Notification *github.Notification
	// Score is the relevance score (higher is better)
	Score float64
	// Matches are the matching fields and positions
	Matches map[string][]Match
}

// Match represents a match in a field
type Match struct {
	// Start is the start position of the match
	Start int
	// End is the end position of the match
	End int
	// Text is the matched text
	Text string
}

// Searcher searches notifications
type Searcher struct {
	// Options are the search options
	Options *SearchOptions
	// Index is the search index
	Index *Index
}

// NewSearcher creates a new searcher
func NewSearcher(options *SearchOptions) *Searcher {
	if options == nil {
		options = DefaultSearchOptions()
	}

	return &Searcher{
		Options: options,
		Index:   NewIndex(),
	}
}

// Search searches notifications
func (s *Searcher) Search(ctx context.Context, notifications []*github.Notification, query string) ([]*SearchResult, error) {
	if len(notifications) == 0 {
		return nil, nil
	}

	// Update the index
	s.Index.Update(notifications)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.Options.Timeout)
	defer cancel()

	// Parse the query
	parsedQuery, err := s.parseQuery(query)
	if err != nil {
		return nil, err
	}

	// For small sets, don't bother with concurrency
	if len(notifications) < 100 {
		return s.searchSequential(notifications, parsedQuery), nil
	}

	return s.searchConcurrent(ctx, notifications, parsedQuery)
}

// parseQuery parses a search query
func (s *Searcher) parseQuery(query string) (interface{}, error) {
	// For now, just use the query string directly
	// In a more advanced implementation, this would parse complex queries
	if s.Options.UseRegex {
		// Compile the regex
		flags := ""
		if !s.Options.CaseSensitive {
			flags = "(?i)"
		}
		re, err := regexp.Compile(flags + query)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
		return re, nil
	}

	// For simple string search, just return the query
	if !s.Options.CaseSensitive {
		return strings.ToLower(query), nil
	}
	return query, nil
}

// searchSequential searches notifications sequentially
func (s *Searcher) searchSequential(notifications []*github.Notification, query interface{}) []*SearchResult {
	var results []*SearchResult
	for _, n := range notifications {
		result := s.searchNotification(n, query)
		if result != nil {
			results = append(results, result)
		}
	}

	// Sort results by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit the number of results
	if s.Options.MaxResults > 0 && len(results) > s.Options.MaxResults {
		results = results[:s.Options.MaxResults]
	}

	return results
}

// searchConcurrent searches notifications concurrently
func (s *Searcher) searchConcurrent(ctx context.Context, notifications []*github.Notification, query interface{}) ([]*SearchResult, error) {
	// Create channels for input and output
	input := make(chan *github.Notification, len(notifications))
	output := make(chan *SearchResult, len(notifications))
	done := make(chan struct{})

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < s.Options.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range input {
				result := s.searchNotification(n, query)
				if result != nil {
					select {
					case output <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Start a goroutine to close the output channel when all workers are done
	go func() {
		wg.Wait()
		close(output)
		close(done)
	}()

	// Feed input channel with notifications
	go func() {
		defer close(input)
		for _, n := range notifications {
			select {
			case input <- n:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	var results []*SearchResult
	for {
		select {
		case result, ok := <-output:
			if !ok {
				// Sort results by score (descending)
				sort.Slice(results, func(i, j int) bool {
					return results[i].Score > results[j].Score
				})

				// Limit the number of results
				if s.Options.MaxResults > 0 && len(results) > s.Options.MaxResults {
					results = results[:s.Options.MaxResults]
				}

				return results, nil
			}
			results = append(results, result)
		case <-ctx.Done():
			return results, ctx.Err()
		case <-done:
			// Sort results by score (descending)
			sort.Slice(results, func(i, j int) bool {
				return results[i].Score > results[j].Score
			})

			// Limit the number of results
			if s.Options.MaxResults > 0 && len(results) > s.Options.MaxResults {
				results = results[:s.Options.MaxResults]
			}

			return results, nil
		}
	}
}

// searchNotification searches a notification
func (s *Searcher) searchNotification(n *github.Notification, query interface{}) *SearchResult {
	// Check each field for matches
	matches := make(map[string][]Match)
	var score float64

	for _, field := range s.Options.Fields {
		var value string
		switch field {
		case "title":
			value = n.GetSubject().GetTitle()
		case "repository", "repo":
			value = n.GetRepository().GetFullName()
		case "type":
			value = n.GetSubject().GetType()
		case "reason":
			value = n.GetReason()
		default:
			continue
		}

		// Search for matches
		fieldMatches := s.findMatches(value, query)
		if len(fieldMatches) > 0 {
			matches[field] = fieldMatches
			// Add to the score based on the number of matches and field importance
			fieldScore := float64(len(fieldMatches)) * s.getFieldWeight(field)
			score += fieldScore
		}
	}

	// If there are no matches, return nil
	if len(matches) == 0 {
		return nil
	}

	// Create a search result
	return &SearchResult{
		Notification: n,
		Score:        score,
		Matches:      matches,
	}
}

// findMatches finds matches in a string
func (s *Searcher) findMatches(value string, query interface{}) []Match {
	if value == "" {
		return nil
	}

	var matches []Match

	switch q := query.(type) {
	case *regexp.Regexp:
		// Regex search
		for _, match := range q.FindAllStringIndex(value, -1) {
			matches = append(matches, Match{
				Start: match[0],
				End:   match[1],
				Text:  value[match[0]:match[1]],
			})
		}
	case string:
		// Simple string search
		searchValue := value
		searchQuery := q

		if !s.Options.CaseSensitive {
			searchValue = strings.ToLower(value)
			// searchQuery is already lowercase if case-insensitive
		}

		// Find all occurrences
		for i := 0; i <= len(searchValue)-len(searchQuery); i++ {
			if searchValue[i:i+len(searchQuery)] == searchQuery {
				matches = append(matches, Match{
					Start: i,
					End:   i + len(searchQuery),
					Text:  value[i : i+len(searchQuery)],
				})
			}
		}
	}

	return matches
}

// getFieldWeight returns the weight for a field
func (s *Searcher) getFieldWeight(field string) float64 {
	switch field {
	case "title":
		return 1.0
	case "repository", "repo":
		return 0.8
	case "type":
		return 0.6
	case "reason":
		return 0.7
	default:
		return 0.5
	}
}

// HighlightMatches highlights matches in a string
func (s *Searcher) HighlightMatches(value string, matches []Match) string {
	if !s.Options.HighlightMatches || len(matches) == 0 {
		return value
	}

	// Sort matches by start position
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Start < matches[j].Start
	})

	// Build the highlighted string
	var result strings.Builder
	lastEnd := 0

	for _, match := range matches {
		// Add text before the match
		result.WriteString(value[lastEnd:match.Start])
		// Add highlighted match
		result.WriteString(s.Options.HighlightPrefix)
		result.WriteString(value[match.Start:match.End])
		result.WriteString(s.Options.HighlightSuffix)
		lastEnd = match.End
	}

	// Add remaining text
	result.WriteString(value[lastEnd:])

	return result.String()
}
