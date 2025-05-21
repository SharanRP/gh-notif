package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/user/gh-notif/internal/ui"
)

// createMockNotifications creates mock notifications for testing
func createMockNotifications() []*github.Notification {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	twoDaysAgo := now.Add(-48 * time.Hour)

	return []*github.Notification{
		{
			ID:         github.String("1"),
			Repository: &github.Repository{FullName: github.String("user/repo1")},
			Subject: &github.NotificationSubject{
				Title: github.String("Issue: Fix bug in authentication"),
				Type:  github.String("Issue"),
				URL:   github.String("https://api.github.com/repos/user/repo1/issues/1"),
			},
			Reason:    github.String("mention"),
			Unread:    github.Bool(true),
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID:         github.String("2"),
			Repository: &github.Repository{FullName: github.String("user/repo2")},
			Subject: &github.NotificationSubject{
				Title: github.String("PR: Add new feature"),
				Type:  github.String("PullRequest"),
				URL:   github.String("https://api.github.com/repos/user/repo2/pulls/2"),
			},
			Reason:    github.String("review_requested"),
			Unread:    github.Bool(true),
			UpdatedAt: &github.Timestamp{Time: yesterday},
		},
		{
			ID:         github.String("3"),
			Repository: &github.Repository{FullName: github.String("org/repo3")},
			Subject: &github.NotificationSubject{
				Title: github.String("Release v1.0.0"),
				Type:  github.String("Release"),
				URL:   github.String("https://api.github.com/repos/org/repo3/releases/1"),
			},
			Reason:    github.String("subscribed"),
			Unread:    github.Bool(false),
			UpdatedAt: &github.Timestamp{Time: twoDaysAgo},
		},
		{
			ID:         github.String("4"),
			Repository: &github.Repository{FullName: github.String("org/repo4")},
			Subject: &github.NotificationSubject{
				Title: github.String("Discussion: New ideas"),
				Type:  github.String("Discussion"),
				URL:   github.String("https://api.github.com/repos/org/repo4/discussions/1"),
			},
			Reason:    github.String("subscribed"),
			Unread:    github.Bool(true),
			UpdatedAt: &github.Timestamp{Time: now},
		},
	}
}

func main() {
	// Print to stderr to make sure we see output
	fmt.Fprintf(os.Stderr, "Starting test program...\n")

	// Create a file for output
	f, err := os.Create("C:/Users/SHARAN/OneDrive/Desktop/tp/ghnotify/test_results.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Write directly to the file
	writer := f

	// Print current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(writer, "Error getting current working directory: %v\n", err)
	} else {
		fmt.Fprintf(writer, "Current working directory: %s\n", cwd)
	}

	// Create mock notifications
	notifications := createMockNotifications()
	fmt.Fprintf(writer, "Created %d mock notifications\n", len(notifications))

	// Print notifications to console and file
	fmt.Fprintf(writer, "GitHub Notifications:\n")
	fmt.Fprintf(writer, "====================\n")

	for i, n := range notifications {
		fmt.Fprintf(writer, "%d. [%s] %s: %s (%s)\n",
			i+1,
			n.GetSubject().GetType(),
			n.GetRepository().GetFullName(),
			n.GetSubject().GetTitle(),
			n.GetUpdatedAt().Format(time.RFC3339))
	}

	fmt.Fprintf(writer, "\nTest completed successfully.\n")

	// Verify our UI components
	fmt.Fprintf(writer, "\nVerifying UI components:\n")
	fmt.Fprintf(writer, "======================\n")

	// Test theme creation
	theme := ui.DefaultDarkTheme()
	fmt.Fprintf(writer, "Created default dark theme successfully\n")

	// Test styles creation
	_ = ui.NewStyles(theme)
	fmt.Fprintf(writer, "Created styles successfully\n")

	// Test symbols
	_ = ui.DefaultSymbols()
	fmt.Fprintf(writer, "Created symbols successfully\n")

	fmt.Fprintf(writer, "\nAll tests completed successfully.\n")
}
