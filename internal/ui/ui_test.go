package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
)

// TestModelCreation tests that a model can be created
func TestModelCreation(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications()

	// Create model
	model := NewModel(notifications)

	// Check that the model was created correctly
	if model.notifications == nil {
		t.Error("Model notifications is nil")
	}

	if len(model.notifications) != len(notifications) {
		t.Errorf("Model has %d notifications, expected %d", len(model.notifications), len(notifications))
	}

	if model.selected != 0 {
		t.Errorf("Model selected is %d, expected 0", model.selected)
	}

	if model.viewMode != CompactView {
		t.Errorf("Model view mode is %d, expected %d", model.viewMode, CompactView)
	}

	if model.colorScheme != DarkScheme {
		t.Errorf("Model color scheme is %d, expected %d", model.colorScheme, DarkScheme)
	}
}

// TestModelUpdate tests that the model can be updated
func TestModelUpdate(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications()

	// Create model
	model := NewModel(notifications)

	// Test key presses
	testCases := []struct {
		name     string
		key      tea.KeyType
		keyRune  rune
		expected func(m Model) bool
	}{
		{
			name:    "Down arrow selects next notification",
			key:     tea.KeyDown,
			keyRune: 0,
			expected: func(m Model) bool {
				return m.selected == 1
			},
		},
		{
			name:    "Up arrow selects previous notification",
			key:     tea.KeyUp,
			keyRune: 0,
			expected: func(m Model) bool {
				return m.selected == 0
			},
		},
		{
			name:    "v key changes view mode",
			key:     tea.KeyRunes,
			keyRune: 'v',
			expected: func(m Model) bool {
				return m.viewMode == DetailedView
			},
		},
		{
			name:    "c key changes color scheme",
			key:     tea.KeyRunes,
			keyRune: 'c',
			expected: func(m Model) bool {
				return m.colorScheme == LightScheme
			},
		},
		{
			name:    "? key toggles help",
			key:     tea.KeyRunes,
			keyRune: '?',
			expected: func(m Model) bool {
				return m.showHelp
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var keyMsg tea.KeyMsg
			if tc.key == tea.KeyRunes {
				keyMsg = tea.KeyMsg{Type: tc.key, Runes: []rune{tc.keyRune}}
			} else {
				keyMsg = tea.KeyMsg{Type: tc.key}
			}

			updatedModel, _ := model.Update(keyMsg)
			m, ok := updatedModel.(Model)
			if !ok {
				t.Fatal("Updated model is not a Model")
			}

			if !tc.expected(m) {
				t.Errorf("Expected condition not met after key press")
			}

			// Update the model for the next test
			model = m
		})
	}
}

// TestFilterNotifications tests the filter functionality
func TestFilterNotifications(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications()

	// Create model
	model := NewModel(notifications)

	// Test filtering
	testCases := []struct {
		filter   string
		expected int
	}{
		{
			filter:   "issue",
			expected: 1,
		},
		{
			filter:   "pull",
			expected: 1,
		},
		{
			filter:   "repo1",
			expected: 2,
		},
		{
			filter:   "nonexistent",
			expected: 0,
		},
		{
			filter:   "",
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filter, func(t *testing.T) {
			model.filterString = tc.filter
			model.filterNotifications()

			if len(model.filteredItems) != tc.expected {
				t.Errorf("Filter '%s' returned %d items, expected %d",
					tc.filter, len(model.filteredItems), tc.expected)
			}
		})
	}
}

// TestMarkdownRenderer tests the markdown renderer
func TestMarkdownRenderer(t *testing.T) {
	// Create styles
	theme := DefaultDarkTheme()
	styles := NewStyles(theme)

	// Create renderer
	renderer := NewMarkdownRenderer(styles, 80)

	// Test rendering
	testCases := []struct {
		name     string
		markdown string
		contains string
	}{
		{
			name:     "Header",
			markdown: "# Test Header",
			contains: "Test Header",
		},
		{
			name:     "Bold",
			markdown: "This is **bold** text",
			contains: "bold",
		},
		{
			name:     "Italic",
			markdown: "This is *italic* text",
			contains: "italic",
		},
		{
			name:     "Code",
			markdown: "This is `code` text",
			contains: "code",
		},
		{
			name:     "List",
			markdown: "- Item 1\n- Item 2",
			contains: "Item 1",
		},
		{
			name:     "Quote",
			markdown: "> This is a quote",
			contains: "This is a quote",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := renderer.Render(tc.markdown)
			if !contains(result, tc.contains) {
				t.Errorf("Rendered markdown does not contain '%s'", tc.contains)
			}
		})
	}
}

// Helper functions

// createTestNotifications creates test notifications for testing
func createTestNotifications() []*github.Notification {
	now := time.Now()

	return []*github.Notification{
		{
			ID:     github.String("1"),
			Unread: github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String("Test Issue"),
				Type:  github.String("Issue"),
				URL:   github.String("https://api.github.com/repos/test/repo1/issues/1"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now},
		},
		{
			ID:     github.String("2"),
			Unread: github.Bool(true),
			Subject: &github.NotificationSubject{
				Title: github.String("Test Pull Request"),
				Type:  github.String("PullRequest"),
				URL:   github.String("https://api.github.com/repos/test/repo1/pulls/1"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo1"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-1 * time.Hour)},
		},
		{
			ID:     github.String("3"),
			Unread: github.Bool(false),
			Subject: &github.NotificationSubject{
				Title: github.String("Test Release"),
				Type:  github.String("Release"),
				URL:   github.String("https://api.github.com/repos/test/repo2/releases/1"),
			},
			Repository: &github.Repository{
				FullName: github.String("test/repo2"),
			},
			UpdatedAt: &github.Timestamp{Time: now.Add(-24 * time.Hour)},
		},
	}
}

// contains checks if a string contains another string
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s[len(s)-1] != 0
}
