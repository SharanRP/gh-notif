package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/common"
)

// viewSelectMode renders the select mode view
func (m ActionModel) viewSelectMode() string {
	var sb strings.Builder

	// Render header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Select Notifications")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Render list
	sb.WriteString(m.list.View())
	sb.WriteString("\n")

	// Render status bar
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(m.statusBar.text)
	sb.WriteString(statusBar)
	sb.WriteString("\n")

	// Render help
	helpView := m.help.View(m.keyMap)
	sb.WriteString(helpView)

	return sb.String()
}

// viewActionMenuMode renders the action menu mode view
func (m ActionModel) viewActionMenuMode() string {
	var sb strings.Builder

	// Render header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render(fmt.Sprintf("Choose Action for %d Selected Notifications", m.selectedCount))
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Render menu
	menuStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	menuItems := []string{
		"[r] Mark as Read",
		"[x] Archive",
		"[s] Subscribe",
		"[u] Unsubscribe",
		"[m] Mute Repository",
		"[esc] Back",
	}

	menu := menuStyle.Render(strings.Join(menuItems, "\n"))
	sb.WriteString(menu)
	sb.WriteString("\n\n")

	// Render help
	helpView := m.help.View(m.keyMap)
	sb.WriteString(helpView)

	return sb.String()
}

// viewProgressMode renders the progress mode view
func (m ActionModel) viewProgressMode() string {
	var sb strings.Builder

	// Render header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Processing Notifications")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Render spinner
	spinner := m.spinner.View()
	sb.WriteString(spinner)
	sb.WriteString(" Processing... Please wait")
	sb.WriteString("\n\n")

	// Render progress bar (placeholder)
	progressStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62"))

	progress := progressStyle.Render("[=======================>                   ] 50%")
	sb.WriteString(progress)
	sb.WriteString("\n\n")

	// Render status
	sb.WriteString("Press Esc to cancel")
	sb.WriteString("\n")

	return sb.String()
}

// viewResultMode renders the result mode view
func (m ActionModel) viewResultMode() string {
	var sb strings.Builder

	// Render header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("Operation Complete")
	sb.WriteString(header)
	sb.WriteString("\n\n")

	// Render result
	if m.result == nil {
		sb.WriteString("No result available")
	} else {
		resultStyle := lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

		// Build result text
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("Total: %d\n", m.result.TotalCount))
		resultText.WriteString(fmt.Sprintf("Success: %d\n", m.result.SuccessCount))
		resultText.WriteString(fmt.Sprintf("Failure: %d\n", m.result.FailureCount))
		resultText.WriteString(fmt.Sprintf("Duration: %s\n", common.FormatDuration(m.result.Duration)))

		if m.result.FailureCount > 0 {
			resultText.WriteString("\nErrors:\n")
			for i, err := range m.result.Errors {
				if i >= 5 {
					resultText.WriteString(fmt.Sprintf("...and %d more errors\n", len(m.result.Errors)-5))
					break
				}
				resultText.WriteString(fmt.Sprintf("- %v\n", err))
			}
		}

		result := resultStyle.Render(resultText.String())
		sb.WriteString(result)
	}
	sb.WriteString("\n\n")

	// Render help
	sb.WriteString("Press Enter to continue")
	sb.WriteString("\n")

	return sb.String()
}

// renderNotificationItem renders a notification item
func (m ActionModel) renderNotificationItem(n *github.Notification, selected bool) string {
	var sb strings.Builder

	// Render selection indicator
	if selected {
		sb.WriteString("[x] ")
	} else {
		sb.WriteString("[ ] ")
	}

	// Render notification type
	typeIcon := getTypeIcon(n.GetSubject().GetType())
	sb.WriteString(typeIcon)
	sb.WriteString(" ")

	// Render repository
	repo := n.GetRepository().GetFullName()
	sb.WriteString(repo)
	sb.WriteString(": ")

	// Render title
	title := n.GetSubject().GetTitle()
	sb.WriteString(title)

	return sb.String()
}

// getTypeIcon returns an icon for the notification type
func getTypeIcon(typ string) string {
	switch typ {
	case "Issue":
		return "🔍"
	case "PullRequest":
		return "🔀"
	case "Release":
		return "📦"
	case "Commit":
		return "📝"
	case "Discussion":
		return "💬"
	default:
		return "📋"
	}
}
