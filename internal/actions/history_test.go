package actions

import (
	"sync"
	"testing"
	"time"
)

func TestActionHistory(t *testing.T) {
	// Create a new action history with a max size of 5
	history := NewActionHistory(5)

	// Check initial state
	if len(history.Actions) != 0 {
		t.Errorf("Expected empty history, got %d actions", len(history.Actions))
	}

	// Add actions
	for i := 0; i < 10; i++ {
		action := Action{
			Type:           ActionMarkAsRead,
			NotificationID: string(rune('A' + i)),
			Timestamp:      time.Now(),
			Success:        true,
		}
		history.Add(action)
	}

	// Check that the history is limited to max size
	if len(history.Actions) != 5 {
		t.Errorf("Expected history size 5, got %d", len(history.Actions))
	}

	// Check that the most recent actions are kept
	// The actions should be in reverse order (most recent first)
	for i := 0; i < 5; i++ {
		expected := string(rune('A' + 9 - i))
		if history.Actions[i].NotificationID != expected {
			t.Errorf("Expected notification ID %s at index %d, got %s",
				expected, i, history.Actions[i].NotificationID)
		}
	}

	// Test GetLast
	actions := history.GetLast(3)
	if len(actions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(actions))
	}
	for i := 0; i < 3; i++ {
		expected := string(rune('A' + 9 - i))
		if actions[i].NotificationID != expected {
			t.Errorf("Expected notification ID %s at index %d, got %s",
				expected, i, actions[i].NotificationID)
		}
	}

	// Test GetLast with a count larger than the history size
	actions = history.GetLast(10)
	if len(actions) != 5 {
		t.Errorf("Expected 5 actions, got %d", len(actions))
	}

	// Test GetLast with a negative count
	actions = history.GetLast(-1)
	if len(actions) != 5 {
		t.Errorf("Expected 5 actions, got %d", len(actions))
	}

	// Test Clear
	history.Clear()
	if len(history.Actions) != 0 {
		t.Errorf("Expected empty history after clear, got %d actions", len(history.Actions))
	}
}

func TestGetActionHistory(t *testing.T) {
	// Reset the singleton
	actionHistory = nil
	actionHistoryOnce = sync.Once{}

	// Get the action history
	history := GetActionHistory()
	if history == nil {
		t.Errorf("Expected non-nil action history")
	}

	// Check that it's a singleton
	history2 := GetActionHistory()
	if history != history2 {
		t.Errorf("Expected the same action history instance")
	}

	// Check the default max size
	if history.maxSize != 100 {
		t.Errorf("Expected default max size 100, got %d", history.maxSize)
	}
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "less than a second"},
		{1 * time.Second, "1s"},
		{1*time.Minute + 30*time.Second, "1m30s"},
		{2*time.Hour + 15*time.Minute, "2h15m0s"},
	}

	for _, tc := range testCases {
		result := FormatDuration(tc.duration)
		if result != tc.expected {
			t.Errorf("FormatDuration(%v) = %s, expected %s", tc.duration, result, tc.expected)
		}
	}
}

func TestFormatProgress(t *testing.T) {
	testCases := []struct {
		completed int
		total     int
		expected  string
	}{
		{0, 10, "0.0%"},
		{5, 10, "50.0%"},
		{10, 10, "100.0%"},
		{0, 0, "0%"},
	}

	for _, tc := range testCases {
		result := FormatProgress(tc.completed, tc.total)
		if result != tc.expected {
			t.Errorf("FormatProgress(%d, %d) = %s, expected %s", tc.completed, tc.total, result, tc.expected)
		}
	}
}
