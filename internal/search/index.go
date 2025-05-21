package search

import (
	"strings"
	"sync"

	"github.com/google/go-github/v60/github"
)

// Index is a search index for notifications
type Index struct {
	// titleIndex is an inverted index for notification titles
	titleIndex map[string][]string
	// repoIndex is an inverted index for repository names
	repoIndex map[string][]string
	// typeIndex is an inverted index for notification types
	typeIndex map[string][]string
	// reasonIndex is an inverted index for notification reasons
	reasonIndex map[string][]string
	// notificationMap is a map of notification ID to notification
	notificationMap map[string]*github.Notification
	// mu protects the index
	mu sync.RWMutex
}

// NewIndex creates a new search index
func NewIndex() *Index {
	return &Index{
		titleIndex:      make(map[string][]string),
		repoIndex:       make(map[string][]string),
		typeIndex:       make(map[string][]string),
		reasonIndex:     make(map[string][]string),
		notificationMap: make(map[string]*github.Notification),
	}
}

// Update updates the index with notifications
func (i *Index) Update(notifications []*github.Notification) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Clear the index
	i.titleIndex = make(map[string][]string)
	i.repoIndex = make(map[string][]string)
	i.typeIndex = make(map[string][]string)
	i.reasonIndex = make(map[string][]string)
	i.notificationMap = make(map[string]*github.Notification)

	// Index each notification
	for _, n := range notifications {
		i.indexNotification(n)
	}
}

// indexNotification indexes a notification
func (i *Index) indexNotification(n *github.Notification) {
	id := n.GetID()
	i.notificationMap[id] = n

	// Index the title
	title := n.GetSubject().GetTitle()
	if title != "" {
		words := tokenize(title)
		for _, word := range words {
			i.titleIndex[word] = append(i.titleIndex[word], id)
		}
	}

	// Index the repository
	repo := n.GetRepository().GetFullName()
	if repo != "" {
		words := tokenize(repo)
		for _, word := range words {
			i.repoIndex[word] = append(i.repoIndex[word], id)
		}
	}

	// Index the type
	typ := n.GetSubject().GetType()
	if typ != "" {
		i.typeIndex[strings.ToLower(typ)] = append(i.typeIndex[strings.ToLower(typ)], id)
	}

	// Index the reason
	reason := n.GetReason()
	if reason != "" {
		i.reasonIndex[strings.ToLower(reason)] = append(i.reasonIndex[strings.ToLower(reason)], id)
	}
}

// Search searches the index for notifications matching a query
func (i *Index) Search(query string) []*github.Notification {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Tokenize the query
	queryWords := tokenize(query)
	if len(queryWords) == 0 {
		return nil
	}

	// Search each index
	titleMatches := i.searchIndex(i.titleIndex, queryWords)
	repoMatches := i.searchIndex(i.repoIndex, queryWords)
	typeMatches := i.searchIndex(i.typeIndex, queryWords)
	reasonMatches := i.searchIndex(i.reasonIndex, queryWords)

	// Combine the results
	resultMap := make(map[string]bool)
	for _, id := range titleMatches {
		resultMap[id] = true
	}
	for _, id := range repoMatches {
		resultMap[id] = true
	}
	for _, id := range typeMatches {
		resultMap[id] = true
	}
	for _, id := range reasonMatches {
		resultMap[id] = true
	}

	// Convert to a list of notifications
	var results []*github.Notification
	for id := range resultMap {
		if n, ok := i.notificationMap[id]; ok {
			results = append(results, n)
		}
	}

	return results
}

// searchIndex searches an index for notifications matching query words
func (i *Index) searchIndex(index map[string][]string, queryWords []string) []string {
	if len(queryWords) == 0 {
		return nil
	}

	// Find notifications that match all query words
	var resultIDs []string
	firstWord := true

	for _, word := range queryWords {
		ids, ok := index[word]
		if !ok {
			continue
		}

		if firstWord {
			// For the first word, use all matching IDs
			resultIDs = append(resultIDs, ids...)
			firstWord = false
		} else {
			// For subsequent words, intersect with existing results
			resultIDs = intersect(resultIDs, ids)
		}
	}

	return resultIDs
}

// intersect returns the intersection of two slices
func intersect(a, b []string) []string {
	// Create a map for faster lookups
	bMap := make(map[string]bool)
	for _, id := range b {
		bMap[id] = true
	}

	// Find the intersection
	var result []string
	for _, id := range a {
		if bMap[id] {
			result = append(result, id)
		}
	}

	return result
}

// tokenize splits a string into tokens
func tokenize(s string) []string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace non-alphanumeric characters with spaces
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return ' '
	}, s)

	// Split into words
	words := strings.Fields(s)

	// Filter out short words and duplicates
	result := make([]string, 0, len(words))
	seen := make(map[string]bool)
	for _, word := range words {
		// Skip very short words (a, an, the, etc.)
		if len(word) <= 1 {
			continue
		}

		// Skip common stop words
		if word == "an" || word == "the" || word == "is" || word == "are" || word == "was" || word == "were" {
			continue
		}

		if !seen[word] {
			result = append(result, word)
			seen[word] = true
		}
	}

	return result
}

// GetNotification gets a notification by ID
func (i *Index) GetNotification(id string) (*github.Notification, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	n, ok := i.notificationMap[id]
	return n, ok
}

// GetNotifications gets all notifications
func (i *Index) GetNotifications() []*github.Notification {
	i.mu.RLock()
	defer i.mu.RUnlock()
	var notifications []*github.Notification
	for _, n := range i.notificationMap {
		notifications = append(notifications, n)
	}
	return notifications
}

// Size returns the number of notifications in the index
func (i *Index) Size() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return len(i.notificationMap)
}
