package filter

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

// TestRepositoryFilter tests the repository filter
func TestRepositoryFilter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create repository filter
	filter, err := NewRepositoryFilter("test/repo1")
	if err != nil {
		t.Fatalf("Failed to create repository filter: %v", err)
	}

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}

	// Test with glob pattern
	filter, err = NewRepositoryFilter("test/*")
	if err != nil {
		t.Fatalf("Failed to create repository filter: %v", err)
	}

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 10 {
		t.Errorf("Expected 10 notifications, got %d", len(filtered))
	}
}

// TestTypeFilter tests the type filter
func TestTypeFilter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create type filter
	filter := NewTypeFilter("Issue")

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}

	// Test with multiple types
	filter = NewTypeFilter("Issue", "PullRequest")

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 10 {
		t.Errorf("Expected 10 notifications, got %d", len(filtered))
	}
}

// TestStatusFilter tests the status filter
func TestStatusFilter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create status filter
	filter := NewStatusFilter(true)

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}

	// Test with read status
	filter = NewStatusFilter(false)

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}
}

// TestTimeFilter tests the time filter
func TestTimeFilter(t *testing.T) {
	// Create test notifications with specific timestamps
	notifications := make([]*github.Notification, 10)
	now := time.Now()

	// Create 5 recent notifications (within last 24 hours)
	for i := 0; i < 5; i++ {
		notifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Recent Notification %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i) * time.Hour)},
		}
	}

	// Create 5 older notifications (older than 48 hours)
	for i := 0; i < 5; i++ {
		notifications[i+5] = &github.Notification{
			ID:      github.String(fmt.Sprintf("%d", i+6)),
			Unread:  github.Bool(false),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Old Notification %d", i+1)),
				Type:  github.String("PullRequest"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i+48) * time.Hour)},
		}
	}

	// Test with since filter (last 24 hours)
	filter := NewTimeFilter().WithSince(now.Add(-24 * time.Hour))

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results - should match the 5 recent notifications
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications within last 24 hours, got %d", len(filtered))
	}

	// Test with before filter (older than 48 hours)
	// Create specific notifications for this test
	olderNotifications := make([]*github.Notification, 10)

	// Create 4 notifications older than 48 hours
	for i := 0; i < 4; i++ {
		olderNotifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("older-%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Older than 48 hours %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			// Older than 48 hours
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(49+i) * time.Hour)},
		}
	}

	// Create 6 notifications newer than 48 hours
	for i := 0; i < 6; i++ {
		olderNotifications[i+4] = &github.Notification{
			ID:      github.String(fmt.Sprintf("newer-%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Newer than 48 hours %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			// Newer than 48 hours
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(24+i) * time.Hour)},
		}
	}

	filter = NewTimeFilter().WithBefore(now.Add(-48 * time.Hour))

	// Apply filter
	filtered = nil
	for _, n := range olderNotifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results - should match the 4 older notifications
	if len(filtered) != 4 {
		t.Errorf("Expected 4 notifications older than 48 hours, got %d", len(filtered))
		for _, n := range filtered {
			t.Logf("Filtered notification: %s, time: %s", n.GetSubject().GetTitle(), n.GetUpdatedAt().Time)
		}
	}

	// Test with both since and before (between 48 and 72 hours old)
	// Create specific notifications for this test
	specificNotifications := make([]*github.Notification, 10)

	// Create 4 notifications between 48 and 72 hours old
	for i := 0; i < 4; i++ {
		specificNotifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("between-%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Between 48-72 hours %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			// Between 48 and 72 hours old
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(48+i) * time.Hour)},
		}
	}

	// Create 3 notifications newer than 48 hours
	for i := 0; i < 3; i++ {
		specificNotifications[i+4] = &github.Notification{
			ID:      github.String(fmt.Sprintf("newer-%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Newer than 48 hours %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			// Newer than 48 hours
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(24+i) * time.Hour)},
		}
	}

	// Create 3 notifications older than 72 hours
	for i := 0; i < 3; i++ {
		specificNotifications[i+7] = &github.Notification{
			ID:      github.String(fmt.Sprintf("older-%d", i+1)),
			Unread:  github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Older than 72 hours %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			// Older than 72 hours
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(72+i) * time.Hour)},
		}
	}

	filter = NewTimeFilter().
		WithSince(now.Add(-72 * time.Hour)).
		WithBefore(now.Add(-48 * time.Hour))

	// Apply filter
	filtered = nil
	for _, n := range specificNotifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Should match exactly 4 notifications (the ones between 48 and 72 hours old)
	if len(filtered) != 4 {
		t.Errorf("Expected 4 notifications between 48 and 72 hours old, got %d", len(filtered))
		for _, n := range filtered {
			t.Logf("Filtered notification: %s, time: %s", n.GetSubject().GetTitle(), n.GetUpdatedAt().Time)
		}
	}
}

// TestRegexFilter tests the regex filter
func TestRegexFilter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create regex filter
	filter, err := NewRegexFilter("Issue.*[0-9]", "title")
	if err != nil {
		t.Fatalf("Failed to create regex filter: %v", err)
	}

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}

	// Test with repository field
	filter, err = NewRegexFilter("test/repo1", "repository")
	if err != nil {
		t.Fatalf("Failed to create regex filter: %v", err)
	}

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if filter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}
}

// TestCompositeFilter tests the composite filter
func TestCompositeFilter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(10)

	// Create repository filter
	repoFilter, err := NewRepositoryFilter("test/repo1")
	if err != nil {
		t.Fatalf("Failed to create repository filter: %v", err)
	}

	// Create type filter
	typeFilter := NewTypeFilter("Issue")

	// Create composite filter with AND
	andFilter := &CompositeFilter{
		Filters:  []Filter{repoFilter, typeFilter},
		Operator: And,
	}

	// Apply filter
	var filtered []*github.Notification
	for _, n := range notifications {
		if andFilter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results - we expect repo1 AND Issue to match 3 notifications
	expectedCount := 0
	for _, n := range notifications {
		if n.GetRepository().GetFullName() == "test/repo1" && n.GetSubject().GetType() == "Issue" {
			expectedCount++
		}
	}
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d notifications, got %d", expectedCount, len(filtered))
	}

	// Create composite filter with OR
	orFilter := &CompositeFilter{
		Filters:  []Filter{repoFilter, typeFilter},
		Operator: Or,
	}

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if orFilter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results - we expect repo1 OR Issue to match
	expectedCount = 0
	for _, n := range notifications {
		if n.GetRepository().GetFullName() == "test/repo1" || n.GetSubject().GetType() == "Issue" {
			expectedCount++
		}
	}
	if len(filtered) != expectedCount {
		t.Errorf("Expected %d notifications, got %d", expectedCount, len(filtered))
	}

	// Create composite filter with NOT
	notFilter := &CompositeFilter{
		Filters:  []Filter{repoFilter},
		Operator: Not,
	}

	// Apply filter
	filtered = nil
	for _, n := range notifications {
		if notFilter.Apply(n) {
			filtered = append(filtered, n)
		}
	}

	// Check results
	if len(filtered) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(filtered))
	}
}

// TestFilterEngine tests the filter engine
func TestFilterEngine(t *testing.T) {
	// Create test notifications with specific patterns
	notifications := make([]*github.Notification, 100)
	now := time.Now()

	// Create notifications with a known pattern for testing
	// We'll create 22 notifications for repo1 (11 Issues, 11 PRs)
	// and 22 notifications for repo2 (11 Issues, 11 PRs)

	// Create notifications for repo1 - exactly 2 Issue type
	for i := 0; i < 2; i++ {
		notifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("repo1-issue-%d", i+1)),
			Unread:  github.Bool(i%2 == 0),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Issue %d", i+1)),
				Type:  github.String("Issue"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i) * time.Hour)},
		}
	}

	// Create notifications for repo1 - PullRequest type
	for i := 0; i < 20; i++ {
		notifications[i+2] = &github.Notification{
			ID:      github.String(fmt.Sprintf("repo1-pr-%d", i+1)),
			Unread:  github.Bool(i%2 == 0),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("PullRequest %d", i+1)),
				Type:  github.String("PullRequest"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i) * time.Hour)},
		}
	}

	// Create notifications for repo2
	for i := 0; i < 22; i++ {
		notifType := "Issue"
		if i >= 11 {
			notifType = "PullRequest"
		}

		notifications[i+22] = &github.Notification{
			ID:      github.String(fmt.Sprintf("repo2-%d", i+1)),
			Unread:  github.Bool(i%2 == 0),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("%s %d", notifType, i+1)),
				Type:  github.String(notifType),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i) * time.Hour)},
		}
	}

	// Fill the rest with dummy notifications
	for i := 44; i < 100; i++ {
		notifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("dummy-%d", i+1)),
			Unread:  github.Bool(i%2 == 0),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Dummy %d", i+1)),
				Type:  github.String("Discussion"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo3"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-time.Duration(i) * time.Hour)},
		}
	}

	// Create repository filter
	repoFilter, err := NewRepositoryFilter("test/repo1")
	if err != nil {
		t.Fatalf("Failed to create repository filter: %v", err)
	}

	// Create type filter
	typeFilter := NewTypeFilter("Issue")

	// Create composite filter with AND
	andFilter := &CompositeFilter{
		Filters:  []Filter{repoFilter, typeFilter},
		Operator: And,
	}

	// Create filter engine
	engine := NewEngine().WithFilter(andFilter)

	// Apply filter
	ctx := context.Background()
	filtered, err := engine.Filter(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to filter notifications: %v", err)
	}

	// Count how many notifications match repo1 AND Issue
	expectedCount := 0
	for _, n := range notifications {
		if n.GetRepository().GetFullName() == "test/repo1" && n.GetSubject().GetType() == "Issue" {
			expectedCount++
		}
	}

	// Should be 2 (the number of Issue notifications in repo1)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(filtered))
	}

	// Test with concurrency
	engine = NewEngine().
		WithFilter(andFilter).
		WithConcurrency(4)

	// Apply filter
	filtered, err = engine.Filter(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to filter notifications: %v", err)
	}

	// Check results - should be the same as sequential filtering
	if len(filtered) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(filtered))
	}
}

// BenchmarkFilterEngine benchmarks the filter engine
func BenchmarkFilterEngine(b *testing.B) {
	// Create test notifications
	notifications := createTestNotifications(1000)

	// Create repository filter
	repoFilter, err := NewRepositoryFilter("test/repo1")
	if err != nil {
		b.Fatalf("Failed to create repository filter: %v", err)
	}

	// Create type filter
	typeFilter := NewTypeFilter("Issue")

	// Create status filter
	statusFilter := NewStatusFilter(true)

	// Create composite filter with AND
	andFilter := &CompositeFilter{
		Filters:  []Filter{repoFilter, typeFilter, statusFilter},
		Operator: And,
	}

	// Create filter engine
	engine := NewEngine().WithFilter(andFilter)

	// Benchmark sequential filtering
	b.Run("Sequential", func(b *testing.B) {
		engine.WithConcurrency(1).WithIndexing(false)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Filter(ctx, notifications)
			if err != nil {
				b.Fatalf("Failed to filter notifications: %v", err)
			}
		}
	})

	// Benchmark concurrent filtering
	b.Run("Concurrent", func(b *testing.B) {
		engine.WithConcurrency(4).WithIndexing(false)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Filter(ctx, notifications)
			if err != nil {
				b.Fatalf("Failed to filter notifications: %v", err)
			}
		}
	})

	// Benchmark indexed filtering
	b.Run("Indexed", func(b *testing.B) {
		engine.WithConcurrency(1).WithIndexing(true)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Filter(ctx, notifications)
			if err != nil {
				b.Fatalf("Failed to filter notifications: %v", err)
			}
		}
	})

	// Benchmark concurrent indexed filtering
	b.Run("ConcurrentIndexed", func(b *testing.B) {
		engine.WithConcurrency(4).WithIndexing(true)
		ctx := context.Background()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := engine.Filter(ctx, notifications)
			if err != nil {
				b.Fatalf("Failed to filter notifications: %v", err)
			}
		}
	})
}

// Helper functions

// createTestNotifications creates test notifications for testing
func createTestNotifications(count int) []*github.Notification {
	notifications := make([]*github.Notification, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		// Alternate between repositories
		repo := "test/repo1"
		if i%2 == 1 {
			repo = "test/repo2"
		}

		// Alternate between types
		typ := "Issue"
		if i%2 == 1 {
			typ = "PullRequest"
		}

		// Alternate between read and unread
		unread := true
		if i%2 == 1 {
			unread = false
		}

		// Alternate between recent and old
		updatedAt := now.Add(-1 * time.Hour)
		if i%2 == 1 {
			updatedAt = now.Add(-72 * time.Hour)
		}

		// Create notification
		notifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("%d", i+1)),
			Unread:  github.Bool(unread),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("%s %d", typ, i+1)),
				Type:  github.String(typ),
				URL:   github.String(fmt.Sprintf("https://api.github.com/repos/%s/%ss/%d", repo, strings.ToLower(typ), i+1)),
			},
			Repository: &github.Repository{
				FullName: github.String(repo),
			},
			UpdatedAt: &github.Timestamp{Time: updatedAt},
			Reason:    github.String("mention"),
		}
	}

	return notifications
}
