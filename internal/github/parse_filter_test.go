package github

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/filter"
)

func TestParseFilterString(t *testing.T) {
	// Test cases
	testCases := []struct {
		name       string
		filterStr  string
		testNotif  *github.Notification
		shouldMatch bool
	}{
		{
			name:      "Empty filter string",
			filterStr: "",
			testNotif: &github.Notification{
				ID: github.String("123"),
			},
			shouldMatch: true,
		},
		{
			name:      "is:read filter",
			filterStr: "is:read",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(false), // read
			},
			shouldMatch: true,
		},
		{
			name:      "is:unread filter",
			filterStr: "is:unread",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
			},
			shouldMatch: true,
		},
		{
			name:      "repo filter",
			filterStr: "repo:owner/repo",
			testNotif: &github.Notification{
				ID: github.String("123"),
				Repository: &github.Repository{
					FullName: github.String("owner/repo"),
				},
			},
			shouldMatch: true,
		},
		{
			name:      "org filter",
			filterStr: "org:owner",
			testNotif: &github.Notification{
				ID: github.String("123"),
				Repository: &github.Repository{
					Owner: &github.User{
						Login: github.String("owner"),
					},
				},
			},
			shouldMatch: true,
		},
		{
			name:      "type filter",
			filterStr: "type:PullRequest",
			testNotif: &github.Notification{
				ID: github.String("123"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			shouldMatch: true,
		},
		{
			name:      "reason filter",
			filterStr: "reason:mention",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Reason: github.String("mention"),
			},
			shouldMatch: true,
		},
		{
			name:      "text filter",
			filterStr: "bug",
			testNotif: &github.Notification{
				ID: github.String("123"),
				Subject: &github.NotificationSubject{
					Title: github.String("Fix critical bug"),
				},
			},
			shouldMatch: true,
		},
		{
			name:      "multiple filters (AND)",
			filterStr: "is:unread type:PullRequest",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			shouldMatch: true,
		},
		{
			name:      "multiple filters (AND) - one doesn't match",
			filterStr: "is:unread type:Issue",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
			},
			shouldMatch: false,
		},
		{
			name:      "complex filter",
			filterStr: "is:unread repo:owner/repo type:PullRequest reason:mention",
			testNotif: &github.Notification{
				ID:     github.String("123"),
				Unread: github.Bool(true), // unread
				Reason: github.String("mention"),
				Subject: &github.NotificationSubject{
					Type: github.String("PullRequest"),
				},
				Repository: &github.Repository{
					FullName: github.String("owner/repo"),
				},
			},
			shouldMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the filter string
			f, err := parseFilterString(tc.filterStr)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Apply the filter to the test notification
			result := f.Apply(tc.testNotif)

			// Check the result
			if result != tc.shouldMatch {
				t.Errorf("Expected filter to return %v, got %v", tc.shouldMatch, result)
			}
		})
	}
}

func TestParseFilterStringTypes(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		filterStr string
		expected  reflect.Type
	}{
		{
			name:      "Empty filter string",
			filterStr: "",
			expected:  reflect.TypeOf(&filter.AllFilter{}),
		},
		{
			name:      "Single is:read filter",
			filterStr: "is:read",
			expected:  reflect.TypeOf(&filter.ReadFilter{}),
		},
		{
			name:      "Single repo filter",
			filterStr: "repo:owner/repo",
			expected:  reflect.TypeOf(&filter.RepoFilter{}),
		},
		{
			name:      "Multiple filters",
			filterStr: "is:unread type:PullRequest",
			expected:  reflect.TypeOf(&filter.AndFilter{}),
		},
		{
			name:      "Text search",
			filterStr: "bug",
			expected:  reflect.TypeOf(&filter.TextFilter{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the filter string
			f, err := parseFilterString(tc.filterStr)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check the type
			actualType := reflect.TypeOf(f)
			if actualType != tc.expected {
				t.Errorf("Expected filter type %v, got %v", tc.expected, actualType)
			}
		})
	}
}

func TestParseFilterStringWithInvalidInput(t *testing.T) {
	// This test is just to ensure that the function doesn't panic with invalid input
	// Since our implementation is simplified, it should return an AllFilter for any input
	
	// Test cases with potentially problematic input
	testCases := []string{
		":",                  // Empty key-value
		"invalid:",           // Missing value
		":invalid",           // Missing key
		"is:invalid",         // Invalid value for is
		"repo:",              // Missing repo name
		"type:InvalidType",   // Invalid type
		"reason:InvalidReason", // Invalid reason
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			// Parse the filter string
			_, err := parseFilterString(tc)
			
			// We don't expect errors in our simplified implementation
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
