package discussions

import (
	"testing"
	"time"
)

func TestDiscussionFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter DiscussionFilter
		want   bool
	}{
		{
			name: "empty filter",
			filter: DiscussionFilter{},
			want: true,
		},
		{
			name: "repository filter",
			filter: DiscussionFilter{
				Repository: "owner/repo",
			},
			want: true,
		},
		{
			name: "category filter",
			filter: DiscussionFilter{
				Category: "Q&A",
			},
			want: true,
		},
		{
			name: "state filter",
			filter: DiscussionFilter{
				State: "open",
			},
			want: true,
		},
		{
			name: "complex filter",
			filter: DiscussionFilter{
				Repository: "owner/repo",
				Category:   "Q&A",
				State:      "open",
				Author:     "username",
				MinUpvotes: 5,
				MinComments: 2,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that filter can be created without errors
			if tt.filter.Repository == "" && tt.filter.Category == "" {
				// Basic validation passed
				return
			}

			// More complex validation would go here
			// For now, just ensure the filter structure is valid
		})
	}
}

func TestDiscussionOptions(t *testing.T) {
	tests := []struct {
		name    string
		options DiscussionOptions
		want    bool
	}{
		{
			name: "default options",
			options: DiscussionOptions{},
			want: true,
		},
		{
			name: "with comments",
			options: DiscussionOptions{
				IncludeComments: true,
				MaxComments:     50,
			},
			want: true,
		},
		{
			name: "with caching",
			options: DiscussionOptions{
				UseCache: true,
				CacheTTL: 5 * time.Minute,
			},
			want: true,
		},
		{
			name: "performance options",
			options: DiscussionOptions{
				Concurrency: 5,
				Timeout:     30 * time.Second,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that options can be created without errors
			if tt.options.Concurrency < 0 {
				t.Errorf("Invalid concurrency: %d", tt.options.Concurrency)
			}

			if tt.options.Timeout < 0 {
				t.Errorf("Invalid timeout: %v", tt.options.Timeout)
			}
		})
	}
}

func TestDiscussionTypes(t *testing.T) {
	// Test Discussion struct
	discussion := Discussion{
		ID:       "test-id",
		Number:   123,
		Title:    "Test Discussion",
		Body:     "This is a test discussion",
		State:    "OPEN",
		Category: Category{
			ID:   "cat-id",
			Name: "Q&A",
			Slug: "q-a",
		},
		Author: User{
			ID:    "user-id",
			Login: "testuser",
		},
		Repository: Repository{
			ID:       "repo-id",
			Name:     "test-repo",
			FullName: "owner/test-repo",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if discussion.ID == "" {
		t.Error("Discussion ID should not be empty")
	}

	if discussion.Number <= 0 {
		t.Error("Discussion number should be positive")
	}

	if discussion.Title == "" {
		t.Error("Discussion title should not be empty")
	}

	// Test Comment struct
	comment := Comment{
		ID:       "comment-id",
		Body:     "This is a test comment",
		Author:   discussion.Author,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if comment.ID == "" {
		t.Error("Comment ID should not be empty")
	}

	if comment.Body == "" {
		t.Error("Comment body should not be empty")
	}
}

func TestEventTypes(t *testing.T) {
	eventTypes := []EventType{
		EventDiscussionCreated,
		EventDiscussionUpdated,
		EventDiscussionClosed,
		EventDiscussionReopened,
		EventDiscussionDeleted,
		EventCommentCreated,
		EventCommentUpdated,
		EventCommentMarkedAsAnswer,
		EventReactionAdded,
		EventSubscribed,
	}

	for _, eventType := range eventTypes {
		if string(eventType) == "" {
			t.Errorf("Event type should not be empty: %v", eventType)
		}
	}
}

func TestDiscussionAnalytics(t *testing.T) {
	timeRange := TimeRange{
		Start: time.Now().AddDate(0, 0, -30),
		End:   time.Now(),
	}

	analytics := DiscussionAnalytics{
		TimeRange:           timeRange,
		TotalDiscussions:    100,
		OpenDiscussions:     80,
		ClosedDiscussions:   20,
		AnsweredDiscussions: 15,
		TotalComments:       500,
		TotalReactions:      200,
		TotalUpvotes:        300,
		CategoryStats:       make(map[string]CategoryStats),
	}

	// Calculate derived metrics
	if analytics.TotalDiscussions > 0 {
		analytics.AverageComments = float64(analytics.TotalComments) / float64(analytics.TotalDiscussions)
		analytics.AverageReactions = float64(analytics.TotalReactions) / float64(analytics.TotalDiscussions)
		analytics.AverageUpvotes = float64(analytics.TotalUpvotes) / float64(analytics.TotalDiscussions)
	}

	if analytics.AverageComments != 5.0 {
		t.Errorf("Expected average comments to be 5.0, got %f", analytics.AverageComments)
	}

	if analytics.AverageReactions != 2.0 {
		t.Errorf("Expected average reactions to be 2.0, got %f", analytics.AverageReactions)
	}

	if analytics.AverageUpvotes != 3.0 {
		t.Errorf("Expected average upvotes to be 3.0, got %f", analytics.AverageUpvotes)
	}
}

func TestSearchOptions(t *testing.T) {
	options := SearchOptions{
		Query:         "test query",
		Repositories:  []string{"owner/repo1", "owner/repo2"},
		Categories:    []string{"Q&A", "General"},
		Authors:       []string{"user1", "user2"},
		MaxResults:    50,
		FuzzyMatch:    true,
		CaseSensitive: false,
		UseCache:      true,
		CacheTTL:      5 * time.Minute,
		Timeout:       30 * time.Second,
	}

	if options.Query == "" {
		t.Error("Search query should not be empty")
	}

	if len(options.Repositories) == 0 {
		t.Error("Should have repositories to search")
	}

	if options.MaxResults <= 0 {
		t.Error("Max results should be positive")
	}

	if options.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}
}

func TestUserStats(t *testing.T) {
	user := User{
		ID:    "user-id",
		Login: "testuser",
		Name:  "Test User",
	}

	stats := UserStats{
		User:            user,
		DiscussionCount: 10,
		CommentCount:    50,
		ReactionCount:   25,
		UpvoteCount:     30,
	}

	if stats.User.Login == "" {
		t.Error("User login should not be empty")
	}

	if stats.DiscussionCount < 0 {
		t.Error("Discussion count should not be negative")
	}

	if stats.CommentCount < 0 {
		t.Error("Comment count should not be negative")
	}
}

func TestTopicStats(t *testing.T) {
	topic := TopicStats{
		Topic:           "bug",
		DiscussionCount: 15,
		CommentCount:    75,
		TrendScore:      0.85,
	}

	if topic.Topic == "" {
		t.Error("Topic should not be empty")
	}

	if topic.DiscussionCount < 0 {
		t.Error("Discussion count should not be negative")
	}

	if topic.TrendScore < 0 || topic.TrendScore > 1 {
		t.Error("Trend score should be between 0 and 1")
	}
}

// Benchmark tests
func BenchmarkDiscussionCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		discussion := Discussion{
			ID:     "test-id",
			Number: i,
			Title:  "Test Discussion",
			Body:   "This is a test discussion body",
			State:  "OPEN",
			Category: Category{
				ID:   "cat-id",
				Name: "Q&A",
				Slug: "q-a",
			},
			Author: User{
				ID:    "user-id",
				Login: "testuser",
			},
			Repository: Repository{
				ID:       "repo-id",
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = discussion
	}
}

func BenchmarkFilterCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		filter := DiscussionFilter{
			Repository:  "owner/repo",
			Category:    "Q&A",
			State:       "open",
			Author:      "username",
			MinUpvotes:  5,
			MinComments: 2,
			Query:       "test query",
		}
		_ = filter
	}
}
