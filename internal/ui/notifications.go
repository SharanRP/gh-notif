package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
	"github.com/pkg/browser"
)

// DisplayNotifications shows the notifications in a terminal UI
func DisplayNotifications(notifications []*github.Notification) error {
	if len(notifications) == 0 {
		fmt.Println("No notifications found.")
		return nil
	}

	// Create the model
	model := NewModel(notifications)

	// Run the UI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// DisplayNotificationsWithOptions shows notifications with custom options
func DisplayNotificationsWithOptions(notifications []*github.Notification, options DisplayOptions) error {
	if len(notifications) == 0 {
		fmt.Println("No notifications found.")
		return nil
	}

	// Create the model
	model := NewModel(notifications)

	// Apply options
	model.viewMode = options.InitialViewMode
	model.colorScheme = options.ColorScheme

	// Set up accessibility if needed
	if options.AccessibilityMode != StandardMode {
		settings := DefaultAccessibilitySettings()
		settings.Mode = options.AccessibilityMode
		settings.ColorScheme = options.ColorScheme
		settings.UseUnicode = options.UseUnicode
		settings.UseAnimations = options.UseAnimations
	}

	// Run the UI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// DisplayOptions contains options for customizing the notification display
type DisplayOptions struct {
	InitialViewMode   ViewMode
	ColorScheme       ColorScheme
	AccessibilityMode AccessibilityMode
	UseUnicode        bool
	UseAnimations     bool
}

// DefaultDisplayOptions returns the default display options
func DefaultDisplayOptions() DisplayOptions {
	return DisplayOptions{
		InitialViewMode:   CompactView,
		ColorScheme:       DarkScheme,
		AccessibilityMode: StandardMode,
		UseUnicode:        true,
		UseAnimations:     true,
	}
}

// CreateTableView creates a table view for notifications
func CreateTableView(notifications []*github.Notification, width, height int) table.Model {
	// Create table columns
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Type", Width: 10},
		{Title: "Repository", Width: 30},
		{Title: "Title", Width: width - 85}, // Adjust based on other columns
		{Title: "Updated", Width: 20},
	}

	// Create table rows
	rows := []table.Row{}
	for i, n := range notifications {
		id := fmt.Sprintf("%d", i+1)
		notifType := n.GetSubject().GetType()
		repo := n.GetRepository().GetFullName()
		title := n.GetSubject().GetTitle()
		updated := formatTime(n.GetUpdatedAt().Time)

		rows = append(rows, table.Row{id, notifType, repo, title, updated})
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows), height-6)),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return t
}

// OpenNotificationInBrowser opens a notification in the browser
func OpenNotificationInBrowser(notification *github.Notification) error {
	if notification == nil {
		return fmt.Errorf("no notification selected")
	}

	// Try to get the HTML URL first
	url := notification.GetSubject().GetURL()
	if url == "" {
		// Fall back to the API URL
		url = notification.GetURL()
	}

	if url == "" {
		return fmt.Errorf("notification has no URL")
	}

	// Convert API URL to HTML URL if needed
	url = convertAPIURLToHTMLURL(url)

	// Open the URL in the browser
	return browser.OpenURL(url)
}

// convertAPIURLToHTMLURL converts a GitHub API URL to an HTML URL
func convertAPIURLToHTMLURL(apiURL string) string {
	// This is a simplified conversion
	// In a real implementation, you would need to handle different types of resources
	return apiURL
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatTime formats a time.Time into a human-readable string
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2")
	}
}
