package filter

import (
	"testing"

	"github.com/google/go-github/v60/github"
)

func TestAndFilter(t *testing.T) {
	// Create test filters
	filter1 := NewTypeFilter("PullRequest")
	filter2 := &ReasonFilter{Reason: "mention"}
	filter3 := &ReadFilter{Read: false} // unread

	// Create an AND filter with these filters
	andFilter := &AndFilter{
		Filters: []Filter{filter1, filter2, filter3},
	}

	// Test cases
	testCases := []struct {
		name         string
		notification *github.Notification
		expected     bool
	}{
		{
			name: "All filters match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: true,
		},
		{
			name: "First filter doesn't match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: false,
		},
		{
			name: "Second filter doesn't match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: false,
		},
		{
			name: "Third filter doesn't match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: false,
		},
		{
			name: "No filters match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := andFilter.Apply(tc.notification)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestOrFilter(t *testing.T) {
	// Create test filters
	filter1 := NewTypeFilter("PullRequest")
	filter2 := &ReasonFilter{Reason: "mention"}
	filter3 := &ReadFilter{Read: false} // unread

	// Create an OR filter with these filters
	orFilter := &OrFilter{
		Filters: []Filter{filter1, filter2, filter3},
	}

	// Test cases
	testCases := []struct {
		name         string
		notification *github.Notification
		expected     bool
	}{
		{
			name: "All filters match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: true,
		},
		{
			name: "Only first filter matches",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: true,
		},
		{
			name: "Only second filter matches",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: true,
		},
		{
			name: "Only third filter matches",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: true,
		},
		{
			name: "No filters match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := orFilter.Apply(tc.notification)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestNotFilter(t *testing.T) {
	// Create test filters
	filter1 := NewTypeFilter("PullRequest")

	// Create a NOT filter
	notFilter := &NotFilter{
		Filter: filter1,
	}

	// Test cases
	testCases := []struct {
		name         string
		notification *github.Notification
		expected     bool
	}{
		{
			name: "Inner filter matches (should return false)",
			notification: &github.Notification{
				ID: github.String("123"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: false,
		},
		{
			name: "Inner filter doesn't match (should return true)",
			notification: &github.Notification{
				ID: github.String("123"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := notFilter.Apply(tc.notification)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestComplexFilter(t *testing.T) {
	// Create a complex filter: (PullRequest OR Issue) AND (mention OR assign) AND unread
	typeFilter1 := NewTypeFilter("PullRequest")
	typeFilter2 := NewTypeFilter("Issue")
	typeFilter := &OrFilter{Filters: []Filter{typeFilter1, typeFilter2}}

	reasonFilter1 := &ReasonFilter{Reason: "mention"}
	reasonFilter2 := &ReasonFilter{Reason: "assign"}
	reasonFilter := &OrFilter{Filters: []Filter{reasonFilter1, reasonFilter2}}

	readFilter := &ReadFilter{Read: false} // unread

	complexFilter := &AndFilter{
		Filters: []Filter{typeFilter, reasonFilter, readFilter},
	}

	// Test cases
	testCases := []struct {
		name         string
		notification *github.Notification
		expected     bool
	}{
		{
			name: "PullRequest, mention, unread - should match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: true,
		},
		{
			name: "Issue, assign, unread - should match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("assign"),
				Subject: &github.NotificationSubject{
					Type: github.String("Issue"),
				},
			},
			expected: true,
		},
		{
			name: "PullRequest, mention, read - should not match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: false,
		},
		{
			name: "PullRequest, comment, unread - should not match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("comment"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			expected: false,
		},
		{
			name: "Release, mention, unread - should not match",
			notification: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("Release"),
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := complexFilter.Apply(tc.notification)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}
