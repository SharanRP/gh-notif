package grouping

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestGrouper(t *testing.T) {
	// Create a test grouper
	options := DefaultGroupOptions()
	options.MinGroupSize = 1  // Set to 1 for testing
	grouper := NewGrouper(options)

	// Create test notifications
	now := time.Now()
	notifications := []*github.Notification{
		{
			ID: github.String("1"),
			Subject: &github.NotificationSubject{
				Title: github.String("Test notification 1"),
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
				Title: github.String("Test notification 2"),
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
				Title: github.String("Test notification 3"),
				Type:  github.String("PullRequest"),
			},
			Reason: github.String("mention"),
			Repository: &github.Repository{
				FullName: github.String("owner/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-2 * time.Hour)},
		},
	}

	// Test grouping by repository
	ctx := context.Background()
	groups, err := grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Test grouping by type
	options.PrimaryGrouping = GroupByType
	grouper = NewGrouper(options)
	groups, err = grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications by type: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Test grouping by reason
	options.PrimaryGrouping = GroupByReason
	grouper = NewGrouper(options)
	groups, err = grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications by reason: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Test grouping by time
	options.PrimaryGrouping = GroupByTime
	grouper = NewGrouper(options)
	groups, err = grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications by time: %v", err)
	}

	// Test grouping with secondary grouping
	options.PrimaryGrouping = GroupByRepository
	options.SecondaryGrouping = GroupByType
	grouper = NewGrouper(options)
	groups, err = grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications with secondary grouping: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Check that the first group has subgroups
	if len(groups[0].Subgroups) == 0 {
		t.Errorf("Expected subgroups, got none")
	}

	// Test smart grouping
	options.PrimaryGrouping = GroupBySmart
	options.SecondaryGrouping = ""
	grouper = NewGrouper(options)
	groups, err = grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications with smart grouping: %v", err)
	}
}

func TestTitleSimilarity(t *testing.T) {
	// Test exact match
	similarity := calculateTitleSimilarity("Test title", "Test title")
	if similarity != 1.0 {
		t.Errorf("Expected similarity 1.0 for exact match, got %f", similarity)
	}

	// Test no match
	similarity = calculateTitleSimilarity("Test title", "Completely different")
	if similarity != 0.0 {
		t.Errorf("Expected similarity 0.0 for no match, got %f", similarity)
	}

	// Test partial match
	similarity = calculateTitleSimilarity("Test title", "Test something")
	if similarity <= 0.0 || similarity >= 1.0 {
		t.Errorf("Expected similarity between 0.0 and 1.0 for partial match, got %f", similarity)
	}

	// Test case insensitivity
	similarity1 := calculateTitleSimilarity("Test title", "test title")
	similarity2 := calculateTitleSimilarity("TEST TITLE", "test title")
	if similarity1 != 1.0 || similarity2 != 1.0 {
		t.Errorf("Expected similarity 1.0 for case-insensitive match, got %f and %f", similarity1, similarity2)
	}
}

func TestGroupByOwner(t *testing.T) {
	// Create a test grouper
	options := DefaultGroupOptions()
	options.PrimaryGrouping = GroupByOwner
	options.MinGroupSize = 1  // Set to 1 for testing
	grouper := NewGrouper(options)

	// Create test notifications
	now := time.Now()
	notifications := []*github.Notification{
		{
			ID: github.String("1"),
			Repository: &github.Repository{
				FullName: github.String("owner1/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID: github.String("2"),
			Repository: &github.Repository{
				FullName: github.String("owner1/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID: github.String("3"),
			Repository: &github.Repository{
				FullName: github.String("owner2/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
	}

	// Test grouping by owner
	ctx := context.Background()
	groups, err := grouper.Group(ctx, notifications)
	if err != nil {
		t.Fatalf("Failed to group notifications by owner: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}

	// Check group names
	foundOwner1 := false
	foundOwner2 := false
	for _, group := range groups {
		if group.Name == "owner1" {
			foundOwner1 = true
			if group.Count != 2 {
				t.Errorf("Expected 2 notifications in owner1 group, got %d", group.Count)
			}
		} else if group.Name == "owner2" {
			foundOwner2 = true
			if group.Count != 1 {
				t.Errorf("Expected 1 notification in owner2 group, got %d", group.Count)
			}
		}
	}

	if !foundOwner1 || !foundOwner2 {
		t.Errorf("Expected groups for owner1 and owner2, got %v", groups)
	}
}
