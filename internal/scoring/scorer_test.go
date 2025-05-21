package scoring

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestScorer(t *testing.T) {
	// Create a test scorer
	factors := DefaultScoreFactors()
	scorer := NewScorer(factors)

	// Create a test notification
	now := time.Now()
	notification := &github.Notification{
		ID: github.String("123"),
		Subject: &github.NotificationSubject{
			Title: github.String("Test notification"),
			Type:  github.String("PullRequest"),
		},
		Reason: github.String("mention"),
		Repository: &github.Repository{
			FullName: github.String("owner/repo"),
		},
		UpdatedAt: &github.Timestamp{Time: now},
	}

	// Test scoring a single notification
	score := scorer.scoreNotification(notification)
	if score == nil {
		t.Fatalf("Expected score, got nil")
	}

	if score.Total <= 0 || score.Total > 100 {
		t.Errorf("Expected score between 1 and 100, got %d", score.Total)
	}

	// Test scoring multiple notifications
	notifications := []*github.Notification{notification}
	ctx := context.Background()
	scores, err := scorer.Score(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to score notifications: %v", err)
	}

	if len(scores) != 1 {
		t.Errorf("Expected 1 score, got %d", len(scores))
	}

	// Test scoring with concurrency
	scorer = scorer.WithConcurrency(2).WithBatchSize(1)
	scores, err = scorer.Score(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to score notifications with concurrency: %v", err)
	}

	if len(scores) != 1 {
		t.Errorf("Expected 1 score, got %d", len(scores))
	}

	// Test scoring with timeout
	scorer = scorer.WithTimeout(1 * time.Second)
	scores, err = scorer.Score(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to score notifications with timeout: %v", err)
	}

	if len(scores) != 1 {
		t.Errorf("Expected 1 score, got %d", len(scores))
	}
}

func TestScoreFilter(t *testing.T) {
	// Create a test scorer
	factors := DefaultScoreFactors()
	scorer := NewScorer(factors)

	// Create a test notification
	now := time.Now()
	notification := &github.Notification{
		ID: github.String("123"),
		Subject: &github.NotificationSubject{
			Title: github.String("Test notification"),
			Type:  github.String("PullRequest"),
		},
		Reason: github.String("mention"),
		Repository: &github.Repository{
			FullName: github.String("owner/repo"),
		},
		UpdatedAt: &github.Timestamp{Time: now},
	}

	// Score the notification
	score := scorer.scoreNotification(notification)

	// Create a score filter with a minimum score
	filter := &ScoreFilter{
		MinScore: score.Total - 1,
		Scorer:   scorer,
		Scores: map[string]*NotificationScore{
			"123": score,
		},
	}

	// Test that the notification passes the filter
	if !filter.Apply(notification) {
		t.Errorf("Expected notification to pass filter")
	}

	// Create a score filter with a maximum score
	filter = &ScoreFilter{
		MaxScore: score.Total + 1,
		Scorer:   scorer,
		Scores: map[string]*NotificationScore{
			"123": score,
		},
	}

	// Test that the notification passes the filter
	if !filter.Apply(notification) {
		t.Errorf("Expected notification to pass filter")
	}

	// Create a score filter with a range that excludes the notification
	filter = &ScoreFilter{
		MinScore: score.Total + 1,
		MaxScore: score.Total + 10,
		Scorer:   scorer,
		Scores: map[string]*NotificationScore{
			"123": score,
		},
	}

	// Test that the notification does not pass the filter
	if filter.Apply(notification) {
		t.Errorf("Expected notification to not pass filter")
	}

	// Test the description
	description := filter.Description()
	if description == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestScoreFactors(t *testing.T) {
	// Create default score factors
	factors := DefaultScoreFactors()

	// Test that the factors are initialized
	if factors.AgeWeight <= 0 {
		t.Errorf("Expected positive age weight, got %f", factors.AgeWeight)
	}

	if factors.ActivityWeight <= 0 {
		t.Errorf("Expected positive activity weight, got %f", factors.ActivityWeight)
	}

	if factors.InvolvementWeight <= 0 {
		t.Errorf("Expected positive involvement weight, got %f", factors.InvolvementWeight)
	}

	if factors.TypeWeight <= 0 {
		t.Errorf("Expected positive type weight, got %f", factors.TypeWeight)
	}

	if factors.ReasonWeight <= 0 {
		t.Errorf("Expected positive reason weight, got %f", factors.ReasonWeight)
	}

	if factors.RepoWeight <= 0 {
		t.Errorf("Expected positive repo weight, got %f", factors.RepoWeight)
	}

	// Test custom repository weights
	if factors.CustomRepoWeights == nil {
		t.Errorf("Expected non-nil custom repo weights")
	}
}
