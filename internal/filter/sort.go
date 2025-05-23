package filter

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-github/v60/github"
)

// SortDirection represents the direction of sorting
type SortDirection int

const (
	// Ascending sorts in ascending order
	Ascending SortDirection = iota
	// Descending sorts in descending order
	Descending
)

// SortField represents a field to sort by
type SortField string

const (
	// SortByRepository sorts by repository name
	SortByRepository SortField = "repository"
	// SortByType sorts by notification type
	SortByType SortField = "type"
	// SortByTitle sorts by notification title
	SortByTitle SortField = "title"
	// SortByTime sorts by notification time
	SortByTime SortField = "time"
	// SortByStatus sorts by notification status
	SortByStatus SortField = "status"
	// SortByReason sorts by notification reason
	SortByReason SortField = "reason"
)

// SortCriterion represents a criterion for sorting
type SortCriterion struct {
	Field     SortField
	Direction SortDirection
}

// NewSortCriterion creates a new sort criterion
func NewSortCriterion(field SortField, direction SortDirection) SortCriterion {
	return SortCriterion{
		Field:     field,
		Direction: direction,
	}
}

// Sorter sorts notifications
type Sorter struct {
	// Criteria is the list of sort criteria
	Criteria []SortCriterion
	// Parallel controls whether to use parallel sorting
	Parallel bool
	// BatchSize controls the size of notification batches for parallel sorting
	BatchSize int
}

// NewSorter creates a new sorter
func NewSorter() *Sorter {
	return &Sorter{
		Parallel:  true,
		BatchSize: 1000,
	}
}

// WithCriteria sets the sort criteria
func (s *Sorter) WithCriteria(criteria ...SortCriterion) *Sorter {
	s.Criteria = criteria
	return s
}

// WithParallel enables or disables parallel sorting
func (s *Sorter) WithParallel(parallel bool) *Sorter {
	s.Parallel = parallel
	return s
}

// WithBatchSize sets the batch size for parallel sorting
func (s *Sorter) WithBatchSize(batchSize int) *Sorter {
	if batchSize > 0 {
		s.BatchSize = batchSize
	}
	return s
}

// Sort sorts the notifications
func (s *Sorter) Sort(notifications []*github.Notification) []*github.Notification {
	if len(notifications) <= 1 || len(s.Criteria) == 0 {
		return notifications
	}

	// Make a copy to avoid modifying the original
	result := make([]*github.Notification, len(notifications))
	copy(result, notifications)

	// For small sets or when parallel is disabled, use sequential sorting
	if len(notifications) < s.BatchSize || !s.Parallel {
		s.sortSequential(result)
		return result
	}

	// Use parallel sorting for large sets
	s.sortParallel(result)
	return result
}

// sortSequential sorts notifications sequentially
func (s *Sorter) sortSequential(notifications []*github.Notification) {
	sort.Slice(notifications, func(i, j int) bool {
		return s.compare(notifications[i], notifications[j])
	})
}

// sortParallel sorts notifications in parallel
func (s *Sorter) sortParallel(notifications []*github.Notification) {
	// Divide the notifications into batches
	numBatches := (len(notifications) + s.BatchSize - 1) / s.BatchSize
	batches := make([][]*github.Notification, numBatches)

	for i := 0; i < numBatches; i++ {
		start := i * s.BatchSize
		end := (i + 1) * s.BatchSize
		if end > len(notifications) {
			end = len(notifications)
		}
		batches[i] = notifications[start:end]
	}

	// Sort each batch in parallel
	var wg sync.WaitGroup
	for i := range batches {
		wg.Add(1)
		go func(batch []*github.Notification) {
			defer wg.Done()
			sort.Slice(batch, func(i, j int) bool {
				return s.compare(batch[i], batch[j])
			})
		}(batches[i])
	}
	wg.Wait()

	// Merge the sorted batches
	s.merge(notifications, batches)
}

// merge merges sorted batches
func (s *Sorter) merge(result []*github.Notification, batches [][]*github.Notification) {
	// Simple merge implementation without using a heap
	// This is less efficient but simpler to implement

	// Create a temporary slice to hold the merged result
	temp := make([]*github.Notification, 0, len(result))

	// Create indices for each batch
	indices := make([]int, len(batches))

	// Merge until all batches are exhausted
	for len(temp) < len(result) {
		// Find the smallest item among all batches
		smallestBatch := -1
		var smallestItem *github.Notification

		for i, batch := range batches {
			// Skip exhausted batches
			if indices[i] >= len(batch) {
				continue
			}

			// Get the current item from this batch
			item := batch[indices[i]]

			// If this is the first valid item or it's smaller than the current smallest
			if smallestBatch == -1 || s.compare(item, smallestItem) {
				smallestBatch = i
				smallestItem = item
			}
		}

		// If no valid items found, we're done
		if smallestBatch == -1 {
			break
		}

		// Add the smallest item to the result
		temp = append(temp, smallestItem)

		// Move to the next item in the batch
		indices[smallestBatch]++
	}

	// Copy the merged result back to the original slice
	copy(result, temp)
}

// compare compares two notifications based on the sort criteria
func (s *Sorter) compare(a, b *github.Notification) bool {
	for _, criterion := range s.Criteria {
		var result int
		switch criterion.Field {
		case SortByRepository:
			result = strings.Compare(a.GetRepository().GetFullName(), b.GetRepository().GetFullName())
		case SortByType:
			result = strings.Compare(a.GetSubject().GetType(), b.GetSubject().GetType())
		case SortByTitle:
			result = strings.Compare(a.GetSubject().GetTitle(), b.GetSubject().GetTitle())
		case SortByTime:
			if a.GetUpdatedAt().Time.Equal(b.GetUpdatedAt().Time) {
				result = 0
			} else if a.GetUpdatedAt().Before(b.GetUpdatedAt().Time) {
				result = -1
			} else {
				result = 1
			}
		case SortByStatus:
			if a.GetUnread() == b.GetUnread() {
				result = 0
			} else if a.GetUnread() {
				result = 1 // Unread comes after read in ascending order
			} else {
				result = -1
			}
		case SortByReason:
			result = strings.Compare(a.GetReason(), b.GetReason())
		default:
			result = 0
		}

		if result != 0 {
			if criterion.Direction == Ascending {
				return result < 0
			}
			return result > 0
		}
	}

	// If all criteria are equal, maintain stable sort by comparing pointers
	return fmt.Sprintf("%p", a) < fmt.Sprintf("%p", b)
}

// SortByField sorts notifications by a single field
func SortByField(notifications []*github.Notification, field SortField, direction SortDirection) []*github.Notification {
	sorter := NewSorter().WithCriteria(NewSortCriterion(field, direction))
	return sorter.Sort(notifications)
}
