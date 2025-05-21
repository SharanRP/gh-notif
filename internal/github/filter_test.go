package github

import (
	"context"
	"testing"

	"github.com/google/go-github/v60/github"
)

func TestGetFilteredNotifications(t *testing.T) {
	// Create a mock client
	client := &Client{}

	// Mock the GetAllNotifications method
	originalGetAllNotifications := client.GetAllNotifications
	defer func() {
		client.GetAllNotifications = originalGetAllNotifications
	}()

	// Create test notifications
	testNotifications := []*github.Notification{
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
		name           string
		options        NotificationOptions
		mockFunc       func()
		expectedCount  int
		expectedIDs    []string
	}{
		{
			name:    "No filter",
			options: NotificationOptions{},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "3"},
		},
		{
			name: "Filter by repository",
			options: NotificationOptions{
				RepoName: "owner/api",
			},
			mockFunc: func() {
				client.GetNotificationsByRepo = func(repoName string, opts NotificationOptions) ([]*github.Notification, error) {
					// In a real implementation, this would filter by repository
					// For testing, we'll just return a subset
					return []*github.Notification{testNotifications[0]}, nil
				}
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name: "Filter by organization",
			options: NotificationOptions{
				OrgName: "owner",
			},
			mockFunc: func() {
				client.GetNotificationsByOrg = func(orgName string, opts NotificationOptions) ([]*github.Notification, error) {
					// In a real implementation, this would filter by organization
					// For testing, we'll just return a subset
					return []*github.Notification{testNotifications[0], testNotifications[1]}, nil
				}
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name: "Filter by unread",
			options: NotificationOptions{
				Unread: true,
			},
			mockFunc: func() {
				client.GetUnreadNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					// In a real implementation, this would filter by unread status
					// For testing, we'll just return a subset
					return []*github.Notification{testNotifications[0], testNotifications[1]}, nil
				}
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name: "Filter string: is:unread",
			options: NotificationOptions{
				FilterString: "is:unread",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name: "Filter string: type:PullRequest",
			options: NotificationOptions{
				FilterString: "type:PullRequest",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "3"},
		},
		{
			name: "Filter string: repo:owner/api",
			options: NotificationOptions{
				FilterString: "repo:owner/api",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name: "Filter string: org:owner",
			options: NotificationOptions{
				FilterString: "org:owner",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 2,
			expectedIDs:   []string{"1", "2"},
		},
		{
			name: "Filter string: reason:mention",
			options: NotificationOptions{
				FilterString: "reason:mention",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name: "Filter string: bug",
			options: NotificationOptions{
				FilterString: "bug",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
		{
			name: "Complex filter: is:unread type:PullRequest",
			options: NotificationOptions{
				FilterString: "is:unread type:PullRequest",
			},
			mockFunc: func() {
				client.GetAllNotifications = func(opts NotificationOptions) ([]*github.Notification, error) {
					return testNotifications, nil
				}
			},
			expectedCount: 1,
			expectedIDs:   []string{"1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mocks
			tc.mockFunc()

			// Call GetFilteredNotifications
			filtered, err := client.GetFilteredNotifications(tc.options)
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

func TestConvertAPIURLToWebURL(t *testing.T) {
	// Test cases
	testCases := []struct {
		name      string
		apiURL    string
		expected  string
		expectErr bool
	}{
		{
			name:      "Pull request URL",
			apiURL:    "https://api.github.com/repos/owner/repo/pulls/123",
			expected:  "https://github.com/owner/repo/pull/123",
			expectErr: false,
		},
		{
			name:      "Issue URL",
			apiURL:    "https://api.github.com/repos/owner/repo/issues/456",
			expected:  "https://github.com/owner/repo/issue/456",
			expectErr: false,
		},
		{
			name:      "Repository URL",
			apiURL:    "https://api.github.com/repos/owner/repo",
			expected:  "https://github.com/owner/repo",
			expectErr: false,
		},
		{
			name:      "Enterprise API URL",
			apiURL:    "https://github.example.com/api/v3/repos/owner/repo/pulls/123",
			expected:  "https://github.example.com/owner/repo/pull/123",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call ConvertAPIURLToWebURL
			webURL, err := ConvertAPIURLToWebURL(tc.apiURL)

			// Check for errors
			if tc.expectErr && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check the result
			if webURL != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, webURL)
			}
		})
	}
}
