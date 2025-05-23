package filter

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

// TestSortByField tests sorting by different fields
func TestSortByField(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Test sorting by repository
	t.Run("SortByRepository", func(t *testing.T) {
		// Sort by repository ascending
		sorted := SortByField(notifications, SortByRepository, Ascending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			repo1 := sorted[i].GetRepository().GetFullName()
			repo2 := sorted[i+1].GetRepository().GetFullName()
			if repo1 > repo2 {
				t.Errorf("Expected %s to be before %s", repo1, repo2)
			}
		}

		// Sort by repository descending
		sorted = SortByField(notifications, SortByRepository, Descending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			repo1 := sorted[i].GetRepository().GetFullName()
			repo2 := sorted[i+1].GetRepository().GetFullName()
			if repo1 < repo2 {
				t.Errorf("Expected %s to be before %s", repo1, repo2)
			}
		}
	})

	// Test sorting by type
	t.Run("SortByType", func(t *testing.T) {
		// Sort by type ascending
		sorted := SortByField(notifications, SortByType, Ascending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			type1 := sorted[i].GetSubject().GetType()
			type2 := sorted[i+1].GetSubject().GetType()
			if type1 > type2 {
				t.Errorf("Expected %s to be before %s", type1, type2)
			}
		}

		// Sort by type descending
		sorted = SortByField(notifications, SortByType, Descending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			type1 := sorted[i].GetSubject().GetType()
			type2 := sorted[i+1].GetSubject().GetType()
			if type1 < type2 {
				t.Errorf("Expected %s to be before %s", type1, type2)
			}
		}
	})

	// Test sorting by time
	t.Run("SortByTime", func(t *testing.T) {
		// Sort by time ascending
		sorted := SortByField(notifications, SortByTime, Ascending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			time1 := sorted[i].GetUpdatedAt().Time
			time2 := sorted[i+1].GetUpdatedAt().Time
			if time1.After(time2) {
				t.Errorf("Expected %s to be before %s", time1, time2)
			}
		}

		// Sort by time descending
		sorted = SortByField(notifications, SortByTime, Descending)

		// Check results
		for i := 0; i < len(sorted)-1; i++ {
			time1 := sorted[i].GetUpdatedAt().Time
			time2 := sorted[i+1].GetUpdatedAt().Time
			if time1.Before(time2) {
				t.Errorf("Expected %s to be before %s", time1, time2)
			}
		}
	})

	// Test sorting by status
	t.Run("SortByStatus", func(t *testing.T) {
		// Sort by status ascending
		sorted := SortByField(notifications, SortByStatus, Ascending)

		// Check results - in ascending order, read comes before unread
		var readFound, unreadFound bool
		var lastReadIndex, firstUnreadIndex int

		for i, n := range sorted {
			if n.GetUnread() {
				unreadFound = true
				if firstUnreadIndex == 0 && i > 0 {
					firstUnreadIndex = i
				}
			} else {
				readFound = true
				lastReadIndex = i
			}
		}

		if readFound && unreadFound && lastReadIndex > firstUnreadIndex {
			t.Errorf("Expected all read notifications to come before unread ones in ascending order")
		}

		// Sort by status descending
		sorted = SortByField(notifications, SortByStatus, Descending)

		// Check results - in descending order, unread comes before read
		readFound, unreadFound = false, false
		lastUnreadIndex, firstReadIndex := 0, 0

		for i, n := range sorted {
			if n.GetUnread() {
				unreadFound = true
				lastUnreadIndex = i
			} else {
				readFound = true
				if firstReadIndex == 0 && i > 0 {
					firstReadIndex = i
				}
			}
		}

		if readFound && unreadFound && lastUnreadIndex > firstReadIndex {
			t.Errorf("Expected all unread notifications to come before read ones in descending order")
		}
	})
}

// TestMultiSort tests sorting by multiple criteria
func TestMultiSort(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create sorter with multiple criteria
	sorter := NewSorter().WithCriteria(
		NewSortCriterion(SortByRepository, Ascending),
		NewSortCriterion(SortByType, Descending),
	)

	// Sort notifications
	sorted := sorter.Sort(notifications)

	// Check results
	for i := 0; i < len(sorted)-1; i++ {
		repo1 := sorted[i].GetRepository().GetFullName()
		repo2 := sorted[i+1].GetRepository().GetFullName()
		if repo1 > repo2 {
			t.Errorf("Expected %s to be before %s", repo1, repo2)
		} else if repo1 == repo2 {
			type1 := sorted[i].GetSubject().GetType()
			type2 := sorted[i+1].GetSubject().GetType()
			if type1 < type2 {
				t.Errorf("Expected %s to be before %s", type1, type2)
			}
		}
	}
}

// TestParallelSort tests parallel sorting
func TestParallelSort(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(100)

	// Create sorter with parallel sorting
	sorter := NewSorter().
		WithCriteria(NewSortCriterion(SortByRepository, Ascending)).
		WithParallel(true).
		WithBatchSize(10)

	// Sort notifications
	sorted := sorter.Sort(notifications)

	// Check results
	for i := 0; i < len(sorted)-1; i++ {
		repo1 := sorted[i].GetRepository().GetFullName()
		repo2 := sorted[i+1].GetRepository().GetFullName()
		if repo1 > repo2 {
			t.Errorf("Expected %s to be before %s", repo1, repo2)
		}
	}
}

// BenchmarkSort benchmarks the sorting system
func BenchmarkSort(b *testing.B) {
	// Create test notifications
	notifications := createRandomNotifications(1000)

	// Benchmark sequential sorting
	b.Run("Sequential", func(b *testing.B) {
		sorter := NewSorter().
			WithCriteria(
				NewSortCriterion(SortByRepository, Ascending),
				NewSortCriterion(SortByType, Descending),
				NewSortCriterion(SortByTime, Descending),
			).
			WithParallel(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sorter.Sort(notifications)
		}
	})

	// Benchmark parallel sorting
	b.Run("Parallel", func(b *testing.B) {
		sorter := NewSorter().
			WithCriteria(
				NewSortCriterion(SortByRepository, Ascending),
				NewSortCriterion(SortByType, Descending),
				NewSortCriterion(SortByTime, Descending),
			).
			WithParallel(true).
			WithBatchSize(100)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sorter.Sort(notifications)
		}
	})

	// Benchmark single field sorting
	b.Run("SingleField", func(b *testing.B) {
		sorter := NewSorter().
			WithCriteria(NewSortCriterion(SortByTime, Descending)).
			WithParallel(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sorter.Sort(notifications)
		}
	})

	// Benchmark multiple field sorting
	b.Run("MultipleFields", func(b *testing.B) {
		sorter := NewSorter().
			WithCriteria(
				NewSortCriterion(SortByRepository, Ascending),
				NewSortCriterion(SortByType, Descending),
				NewSortCriterion(SortByTime, Descending),
				NewSortCriterion(SortByStatus, Descending),
				NewSortCriterion(SortByTitle, Ascending),
			).
			WithParallel(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sorter.Sort(notifications)
		}
	})

	// Benchmark sorting with different batch sizes
	sizes := []int{10, 50, 100, 200, 500}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("BatchSize_%d", size), func(b *testing.B) {
			sorter := NewSorter().
				WithCriteria(
					NewSortCriterion(SortByRepository, Ascending),
					NewSortCriterion(SortByType, Descending),
					NewSortCriterion(SortByTime, Descending),
				).
				WithParallel(true).
				WithBatchSize(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sorter.Sort(notifications)
			}
		})
	}
}

// Helper functions

// createRandomNotifications creates random notifications for benchmarking
func createRandomNotifications(count int) []*github.Notification {
	notifications := make([]*github.Notification, count)
	now := time.Now()
	rand.Seed(time.Now().UnixNano())

	repos := []string{"user/repo1", "user/repo2", "org/repo3", "org/repo4", "other/repo5"}
	types := []string{"Issue", "PullRequest", "Release", "Discussion", "Commit"}
	reasons := []string{"mention", "assign", "review_requested", "subscribed", "team_mention"}

	for i := 0; i < count; i++ {
		// Random repository
		repo := repos[rand.Intn(len(repos))]

		// Random type
		typ := types[rand.Intn(len(types))]

		// Random read/unread status
		unread := rand.Intn(2) == 0

		// Random updated time
		updatedAt := now.Add(-time.Duration(rand.Intn(168)) * time.Hour)

		// Random reason
		reason := reasons[rand.Intn(len(reasons))]

		// Create notification
		notifications[i] = &github.Notification{
			ID:     github.String(fmt.Sprintf("%d", i+1)),
			Unread: github.Bool(unread),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("%s %d", typ, i+1)),
				Type:  github.String(typ),
				URL:   github.String(fmt.Sprintf("https://api.github.com/repos/%s/%ss/%d", repo, strings.ToLower(typ), i+1)),
			},
			Repository: &github.Repository{
				FullName: github.String(repo),
			},
			UpdatedAt: &github.Timestamp{Time: updatedAt},
			Reason:    github.String(reason),
		}
	}

	return notifications
}
