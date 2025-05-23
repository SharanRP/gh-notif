package filter

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestEngine(t *testing.T) {
	// Create test notifications
	notifications := []*github.Notification{
		{
			ID:     github.String("1"),
			Unread: github.Bool(true), // unread
			Reason: github.String("mention"),
			Subject: &github.NotificationSubject{
				Type:  github.String("PullRequest"),
				Title: github.String("Fix bug in API"),
			},
			Repository: &github.Repository{
				FullName: github.String("owner/api"),
				Owner: &github.User{
					Login: github.String("owner"),
				},
			},
		},
		{
			ID:     github.String("2"),
			Unread: github.Bool(true), // unread
			Reason: github.String("assign"),
			Subject: &github.NotificationSubject{
				Type:  github.String("Issue"),
				Title: github.String("Add new feature"),
			},
			Repository: &github.Repository{
				FullName: github.String("owner/web"),
				Owner: &github.User{
					Login: github.String("owner"),
				},
			},
		},
		{
			ID:     github.String("3"),
			Unread: github.Bool(false), // read
			Reason: github.String("comment"),
			Subject: &github.NotificationSubject{
				Type:  github.String("PullRequest"),
				Title: github.String("Update documentation"),
			},
			Repository: &github.Repository{
				FullName: github.String("other/docs"),
				Owner: &github.User{
					Login: github.String("other"),
				},
			},
		},
	}

	// Test cases
	testCases := []struct {
		name          string
		filter        Filter
		expectedCount int
		expectedIDs   []string
	}{
		{
			name:          "No filter (all notifications)",
			filter:        &AllFilter{},
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name:          "Filter by unread",
			filter:        &ReadFilter{Read: false}, // unread
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name:          "Filter by type",
			filter:        NewTypeFilter("PullRequest"),
			expectedCount: 2,
			expectedIDs:   []string{"1", "3"},
		},
		{
			name:          "Filter by reason",
			filter:        &ReasonFilter{Reason: "mention"},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "Filter by repository",
			filter:        &RepoFilter{Repo: "owner/api"},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "Filter by organization",
			filter:        &OrgFilter{Org: "owner"},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name:          "Filter by text",
			filter:        &TextFilter{Text: "bug"},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "Complex filter: unread AND PullRequest",
			filter:        &AndFilter{Filters: []Filter{&ReadFilter{Read: false}, NewTypeFilter("PullRequest")}},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name:          "Complex filter: PullRequest OR Issue",
			filter:        &OrFilter{Filters: []Filter{NewTypeFilter("PullRequest"), NewTypeFilter("Issue")}},
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name:          "Complex filter: NOT PullRequest",
			filter:        &NotFilter{Filter: NewTypeFilter("PullRequest")},
			expectedCount: 1,
			expectedIDs:   []string{"2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a filter engine with the test filter
			engine := NewEngine().WithFilter(tc.filter)

			// Filter the notifications
			filtered, err := engine.Filter(context.Background(), notifications)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check the count
			if len(filtered) != tc.expectedCount {
				t.Errorf("Expected %d notifications, got %d", tc.expectedCount, len(filtered))
			}

			// Check the IDs
			for i, id := range tc.expectedIDs {
				if i < len(filtered) {
					if *filtered[i].ID != id {
						t.Errorf("Expected notification ID %s at index %d, got %s", id, i, *filtered[i].ID)
					}
				}
			}
		})
	}
}

func TestEngineWithCancellation(t *testing.T) {
	// Create a large number of test notifications
	notifications := make([]*github.Notification, 1000)
	for i := 0; i < 1000; i++ {
		notifications[i] = &github.Notification{
			ID: github.String(string(rune('0' + i%10))),
		}
	}

	// Create a filter that takes some time to process
	slowFilter := &slowTestFilter{}

	// Create a filter engine with the slow filter
	engine := NewEngine().WithFilter(slowFilter)

	// Create a context with cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Filter the notifications
	_, err := engine.Filter(ctx, notifications)

	// Check that the context was canceled
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
}

// slowTestFilter is a test filter that takes time to process
type slowTestFilter struct{}

func (f *slowTestFilter) Apply(notification *github.Notification) bool {
	// Simulate a slow operation
	time.Sleep(1 * time.Millisecond)
	return true
}

func (f *slowTestFilter) Description() string {
	return "slow test filter"
}
