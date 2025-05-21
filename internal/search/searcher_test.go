package search

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestSearcher(t *testing.T) {
	// Create a test searcher
	options := DefaultSearchOptions()
	searcher := NewSearcher(options)

	// Create test notifications
	now := time.Now()
	notifications := []*github.Notification{
		{
			ID: github.String("1"),
			Subject: &github.NotificationSubject{
				Title: github.String("Fix bug in login page"),
				Type:  github.String("PullRequest"),
			},
			Reason: github.String("mention"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID: github.String("2"),
			Subject: &github.NotificationSubject{
				Title: github.String("Add new feature to dashboard"),
				Type:  github.String("Issue"),
			},
			Reason: github.String("assign"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-1 * time.Hour)},
		},
		{
			ID: github.String("3"),
			Subject: &github.NotificationSubject{
				Title: github.String("Update documentation"),
				Type:  github.String("PullRequest"),
			},
			Reason: github.String("mention"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-2 * time.Hour)},
		},
	}

	// Test searching for "bug"
	ctx := context.Background()
	results, err := searcher.Search(ctx, notifications, "bug")
	if err != nil {
		t.Fatalf("Failed to search notifications: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test searching for "feature"
	results, err = searcher.Search(ctx, notifications, "feature")
	if err != nil {
		t.Fatalf("Failed to search notifications: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test searching for "repo1"
	results, err = searcher.Search(ctx, notifications, "repo1")
	if err != nil {
		t.Fatalf("Failed to search notifications: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test searching with regex
	options.UseRegex = true
	searcher = NewSearcher(options)
	results, err = searcher.Search(ctx, notifications, "bug|feature")
	if err != nil {
		t.Fatalf("Failed to search notifications with regex: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Test searching with case sensitivity
	options.UseRegex = false
	options.CaseSensitive = true
	searcher = NewSearcher(options)
	results, err = searcher.Search(ctx, notifications, "Bug")
	if err != nil {
		t.Fatalf("Failed to search notifications with case sensitivity: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}

	// Test searching with max results
	options.CaseSensitive = false
	options.MaxResults = 1
	searcher = NewSearcher(options)
	results, err = searcher.Search(ctx, notifications, "repo")
	if err != nil {
		t.Fatalf("Failed to search notifications with max results: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test highlighting matches
	options.MaxResults = 10
	options.HighlightMatches = true
	searcher = NewSearcher(options)
	results, err = searcher.Search(ctx, notifications, "bug")
	if err != nil {
		t.Fatalf("Failed to search notifications with highlighting: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test highlighting
	title := "Fix bug in login page"
	matches := []Match{
		{
			Start: 4,
			End:   7,
			Text:  "bug",
		},
	}
	highlighted := searcher.HighlightMatches(title, matches)
	if highlighted == title {
		t.Errorf("Expected highlighted text to be different from original")
	}
}

func TestIndex(t *testing.T) {
	// Create a test index
	index := NewIndex()

	// Create test notifications
	now := time.Now()
	notifications := []*github.Notification{
		{
			ID: github.String("1"),
			Subject: &github.NotificationSubject{
				Title: github.String("Fix bug in login page"),
				Type:  github.String("PullRequest"),
			},
			Reason: github.String("mention"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID: github.String("2"),
			Subject: &github.NotificationSubject{
				Title: github.String("Add new feature to dashboard"),
				Type:  github.String("Issue"),
			},
			Reason: github.String("assign"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-1 * time.Hour)},
		},
	}

	// Update the index
	index.Update(notifications)

	// Test the size
	if index.Size() != 2 {
		t.Errorf("Expected index size 2, got %d", index.Size())
	}

	// Test searching
	results := index.Search("bug")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Test getting a notification
	notification, ok := index.GetNotification("1")
	if !ok {
		t.Errorf("Expected to find notification 1")
	}
	if notification.GetID() != "1" {
		t.Errorf("Expected notification ID 1, got %s", notification.GetID())
	}

	// Test getting all notifications
	allNotifications := index.GetNotifications()
	if len(allNotifications) != 2 {
		t.Errorf("Expected 2 notifications, got %d", len(allNotifications))
	}
}

func TestTokenize(t *testing.T) {
	// Test tokenizing a simple string
	tokens := tokenize("Hello world")
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(tokens))
	}
	if tokens[0] != "hello" || tokens[1] != "world" {
		t.Errorf("Expected tokens [hello world], got %v", tokens)
	}

	// Test tokenizing with punctuation
	tokens = tokenize("Hello, world! How are you?")
	// Note: "are" is filtered out as a stop word
	if len(tokens) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(tokens))
	}
	if tokens[0] != "hello" || tokens[1] != "world" || tokens[2] != "how" || tokens[3] != "you" {
		t.Errorf("Expected tokens [hello world how you], got %v", tokens)
	}

	// Test tokenizing with duplicates
	tokens = tokenize("hello hello world")
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens (duplicates removed), got %d", len(tokens))
	}

	// Test tokenizing with short words
	tokens = tokenize("a an the hello world")
	if len(tokens) != 2 && tokens[0] != "hello" && tokens[1] != "world" {
		t.Errorf("Expected tokens [hello world], got %v", tokens)
	}
}
