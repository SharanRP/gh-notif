package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

// TestTextFormatter tests the text formatter
func TestTextFormatter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(5)

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter
	formatter := NewFormatter(&buf).
		WithFormat(FormatText).
		WithNoColor(true)

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "Repository") {
		t.Errorf("Expected output to contain 'Repository', got: %s", output)
	}
	if !strings.Contains(output, "Type") {
		t.Errorf("Expected output to contain 'Type', got: %s", output)
	}
	if !strings.Contains(output, "Title") {
		t.Errorf("Expected output to contain 'Title', got: %s", output)
	}
	if !strings.Contains(output, "Updated") {
		t.Errorf("Expected output to contain 'Updated', got: %s", output)
	}
	if !strings.Contains(output, "Status") {
		t.Errorf("Expected output to contain 'Status', got: %s", output)
	}
	if !strings.Contains(output, "test/repo1") {
		t.Errorf("Expected output to contain 'test/repo1', got: %s", output)
	}
	if !strings.Contains(output, "Issue") {
		t.Errorf("Expected output to contain 'Issue', got: %s", output)
	}
	if !strings.Contains(output, "Unread") {
		t.Errorf("Expected output to contain 'Unread', got: %s", output)
	}
}

// TestJSONFormatter tests the JSON formatter
func TestJSONFormatter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(5)

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter
	formatter := NewFormatter(&buf).
		WithFormat(FormatJSON)

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "\"id\":") {
		t.Errorf("Expected output to contain '\"id\":', got: %s", output)
	}
	if !strings.Contains(output, "\"repository\":") {
		t.Errorf("Expected output to contain '\"repository\":', got: %s", output)
	}
	if !strings.Contains(output, "\"type\":") {
		t.Errorf("Expected output to contain '\"type\":', got: %s", output)
	}
	if !strings.Contains(output, "\"title\":") {
		t.Errorf("Expected output to contain '\"title\":', got: %s", output)
	}
	if !strings.Contains(output, "\"updated_at\":") {
		t.Errorf("Expected output to contain '\"updated_at\":', got: %s", output)
	}
	if !strings.Contains(output, "\"unread\":") {
		t.Errorf("Expected output to contain '\"unread\":', got: %s", output)
	}

	// Parse JSON
	var jsonOutput []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &jsonOutput)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	// Check JSON structure
	if len(jsonOutput) != 5 {
		t.Errorf("Expected 5 notifications, got %d", len(jsonOutput))
	}

	for _, n := range jsonOutput {
		if _, ok := n["id"]; !ok {
			t.Errorf("Expected notification to have 'id' field")
		}
		if _, ok := n["repository"]; !ok {
			t.Errorf("Expected notification to have 'repository' field")
		}
		if _, ok := n["type"]; !ok {
			t.Errorf("Expected notification to have 'type' field")
		}
		if _, ok := n["title"]; !ok {
			t.Errorf("Expected notification to have 'title' field")
		}
		if _, ok := n["updated_at"]; !ok {
			t.Errorf("Expected notification to have 'updated_at' field")
		}
		if _, ok := n["unread"]; !ok {
			t.Errorf("Expected notification to have 'unread' field")
		}
	}
}

// TestCSVFormatter tests the CSV formatter
func TestCSVFormatter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(5)

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter
	formatter := NewFormatter(&buf).
		WithFormat(FormatCSV)

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "Repository,Type,Title,Updated,Status") {
		t.Errorf("Expected output to contain header, got: %s", output)
	}
	if !strings.Contains(output, "test/repo1") {
		t.Errorf("Expected output to contain 'test/repo1', got: %s", output)
	}
	if !strings.Contains(output, "Issue") {
		t.Errorf("Expected output to contain 'Issue', got: %s", output)
	}
	if !strings.Contains(output, "Unread") {
		t.Errorf("Expected output to contain 'Unread', got: %s", output)
	}

	// Count lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 6 { // Header + 5 notifications
		t.Errorf("Expected 6 lines, got %d", len(lines))
	}
}

// TestTemplateFormatter tests the template formatter
func TestTemplateFormatter(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(5)

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter with template
	formatter := NewFormatter(&buf).
		WithFormat(FormatTemplate).
		WithTemplate("{{range .}}{{.GetRepository.GetFullName}}: {{.GetSubject.GetTitle}} ({{if .GetUnread}}Unread{{else}}Read{{end}})\n{{end}}")

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "test/repo1: Issue 1 (Unread)") {
		t.Errorf("Expected output to contain 'test/repo1: Issue 1 (Unread)', got: %s", output)
	}
	if !strings.Contains(output, "test/repo2: PullRequest 2 (Read)") {
		t.Errorf("Expected output to contain 'test/repo2: PullRequest 2 (Read)', got: %s", output)
	}

	// Count lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}
}

// TestCustomFields tests the custom fields option
func TestCustomFields(t *testing.T) {
	// Create test notifications
	notifications := createTestNotifications(5)

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter with custom fields
	formatter := NewFormatter(&buf).
		WithFormat(FormatText).
		WithNoColor(true).
		WithFields([]string{"repository", "type", "status"})

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "Repository") {
		t.Errorf("Expected output to contain 'Repository', got: %s", output)
	}
	if !strings.Contains(output, "Type") {
		t.Errorf("Expected output to contain 'Type', got: %s", output)
	}
	if !strings.Contains(output, "Status") {
		t.Errorf("Expected output to contain 'Status', got: %s", output)
	}
	if strings.Contains(output, "Title") {
		t.Errorf("Expected output not to contain 'Title', got: %s", output)
	}
	if strings.Contains(output, "Updated") {
		t.Errorf("Expected output not to contain 'Updated', got: %s", output)
	}
}

// TestEmptyNotifications tests formatting empty notifications
func TestEmptyNotifications(t *testing.T) {
	// Create empty notifications
	var notifications []*github.Notification

	// Create buffer for output
	var buf bytes.Buffer

	// Create formatter
	formatter := NewFormatter(&buf).
		WithFormat(FormatText).
		WithNoColor(true)

	// Format notifications
	err := formatter.Format(notifications)
	if err != nil {
		t.Fatalf("Failed to format notifications: %v", err)
	}

	// Check output
	output := buf.String()
	if !strings.Contains(output, "No notifications found") {
		t.Errorf("Expected output to contain 'No notifications found', got: %s", output)
	}
}

// Helper functions

// createTestNotifications creates test notifications for testing
func createTestNotifications(count int) []*github.Notification {
	notifications := make([]*github.Notification, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		// Alternate between repositories
		repo := "test/repo1"
		if i%2 == 1 {
			repo = "test/repo2"
		}

		// Alternate between types
		typ := "Issue"
		if i%2 == 1 {
			typ = "PullRequest"
		}

		// Alternate between read and unread
		unread := true
		if i%2 == 1 {
			unread = false
		}

		// Alternate between recent and old
		updatedAt := now.Add(-1 * time.Hour)
		if i%2 == 1 {
			updatedAt = now.Add(-72 * time.Hour)
		}

		// Create notification
		notifications[i] = &github.Notification{
			ID:      github.String(fmt.Sprintf("%d", i+1)),
			Unread:  github.Bool(unread),
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("%s %d", typ, i+1)),
				Type:  github.String(typ),
				URL:   github.String(fmt.Sprintf("https://api.github.com/repos/%s/%ss/%d", repo, strings.ToLower(typ), i+1)),
			},
			Repository: &github.Repository{
				FullName: github.String(repo),
			},
			UpdatedAt: &github.Timestamp{Time: updatedAt},
			Reason:    github.String("mention"),
		}
	}

	return notifications
}
