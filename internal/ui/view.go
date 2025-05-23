package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// View renders the UI
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if len(m.notifications) == 0 {
		return "No notifications found."
	}

	// Get the appropriate theme and styles
	var theme Theme
	switch m.colorScheme {
	case DarkScheme:
		theme = DefaultDarkTheme()
	case LightScheme:
		theme = DefaultLightTheme()
	case HighContrastScheme:
		theme = HighContrastTheme()
	}
	styles := NewStyles(theme)
	symbols := DefaultSymbols()

	// Render the appropriate view
	var content string
	switch m.viewMode {
	case CompactView:
		content = m.renderCompactView(styles, symbols)
	case DetailedView:
		content = m.renderDetailedView(styles, symbols)
	case SplitView:
		content = m.renderSplitView(styles, symbols)
	case TableView:
		content = m.renderTableView(styles, symbols)
	}

	// Render header
	header := styles.Header.Render("GitHub Notifications")
	if m.loading {
		header = lipgloss.JoinHorizontal(lipgloss.Center,
			m.spinner.View(), " ", header)
	}

	// Render filter indicator if active
	if m.filterString != "" {
		filterInfo := styles.FilterPrompt.Render(
			fmt.Sprintf("Filter: %s", m.filterString))
		header = lipgloss.JoinHorizontal(lipgloss.Center,
			header, "  ", filterInfo)
	}

	// Render status bar
	status := styles.StatusBar.Render(m.statusBar.text)

	// Render help
	var help string
	if m.showHelp {
		help = m.help.FullHelpView(m.keyMap.FullHelp())
	} else {
		help = m.help.ShortHelpView(m.keyMap.ShortHelp())
	}
	help = styles.HelpBar.Render(help)

	// Render error if any
	var errorView string
	if m.error != nil {
		errorView = styles.Error.Render(fmt.Sprintf("Error: %v", m.error))
	}

	// Join all components
	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
		"",
		status,
		help,
		errorView,
	)

	return view
}

// renderCompactView renders a compact list of notifications
func (m Model) renderCompactView(styles Styles, symbols Symbols) string {
	var sb strings.Builder

	for i, n := range m.filteredItems {
		var itemStyle lipgloss.Style
		if i == m.selected {
			itemStyle = styles.SelectedItem
		} else {
			itemStyle = styles.ListItem
		}

		// Render status indicator
		var indicator string
		if n.GetUnread() {
			indicator = styles.UnreadIndicator.Render(symbols.UnreadIndicator)
		} else {
			indicator = styles.ReadIndicator.Render(symbols.ReadIndicator)
		}

		// Render notification type icon
		var typeIcon string
		switch n.GetSubject().GetType() {
		case "Issue":
			typeIcon = styles.Issue.Render(symbols.Issue)
		case "PullRequest":
			typeIcon = styles.PullRequest.Render(symbols.PullRequest)
		case "Release":
			typeIcon = styles.Release.Render(symbols.Release)
		case "Discussion":
			typeIcon = styles.Discussion.Render(symbols.Discussion)
		case "Commit":
			typeIcon = styles.Commit.Render(symbols.Commit)
		default:
			typeIcon = styles.Commit.Render(symbols.Dot)
		}

		// Render repository name
		repo := n.GetRepository().GetFullName()

		// Render title with smart truncation
		title := n.GetSubject().GetTitle()
		maxTitleLen := m.width - len(repo) - 10
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-3] + symbols.Ellipsis
		}

		// Render time
		timeStr := formatTimeForView(n.GetUpdatedAt().Time)

		// Join all parts
		line := fmt.Sprintf("%s %s %s: %s (%s)",
			indicator,
			typeIcon,
			repo,
			title,
			timeStr,
		)

		sb.WriteString(itemStyle.Render(line) + "\n")
	}

	return styles.List.
		Width(m.width - 4).
		Height(m.height - 6).
		Render(sb.String())
}

// renderDetailedView renders detailed information for the selected notification
func (m Model) renderDetailedView(styles Styles, symbols Symbols) string {
	n := m.getSelectedNotification()
	if n == nil {
		return styles.NoNotifications.Render("No notification selected")
	}

	var sb strings.Builder

	// Render header with repository and type
	header := fmt.Sprintf("%s %s",
		n.GetRepository().GetFullName(),
		n.GetSubject().GetType(),
	)
	sb.WriteString(styles.DetailHeader.Render(header) + "\n\n")

	// Render title
	title := n.GetSubject().GetTitle()
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

	// Render status
	var status string
	if n.GetUnread() {
		status = styles.UnreadIndicator.Render("Unread")
	} else {
		status = styles.ReadIndicator.Render("Read")
	}
	sb.WriteString(status + "\n\n")

	// Render updated time
	updated := fmt.Sprintf("Updated: %s", n.GetUpdatedAt().Format(time.RFC1123))
	sb.WriteString(updated + "\n\n")

	// Render URL
	url := m.getNotificationURL()
	sb.WriteString(fmt.Sprintf("URL: %s\n\n", url))

	// Render actions
	actions := "Press 'o' to open in browser, 'm' to mark as read"
	sb.WriteString(styles.DetailFooter.Render(actions))

	return styles.DetailView.
		Width(m.width - 4).
		Height(m.height - 6).
		Render(sb.String())
}

// renderSplitView renders a split view with list on the left and details on the right
func (m Model) renderSplitView(styles Styles, symbols Symbols) string {
	// Calculate dimensions
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 3 // Account for divider and padding
	height := m.height - 6

	// Render the list on the left
	var leftSb strings.Builder
	for i, n := range m.filteredItems {
		var itemStyle lipgloss.Style
		if i == m.selected {
			itemStyle = styles.SelectedItem
		} else {
			itemStyle = styles.ListItem
		}

		// Render status indicator
		var indicator string
		if n.GetUnread() {
			indicator = styles.UnreadIndicator.Render(symbols.UnreadIndicator)
		} else {
			indicator = styles.ReadIndicator.Render(symbols.ReadIndicator)
		}

		// Render notification type icon
		var typeIcon string
		switch n.GetSubject().GetType() {
		case "Issue":
			typeIcon = styles.Issue.Render(symbols.Issue)
		case "PullRequest":
			typeIcon = styles.PullRequest.Render(symbols.PullRequest)
		case "Release":
			typeIcon = styles.Release.Render(symbols.Release)
		case "Discussion":
			typeIcon = styles.Discussion.Render(symbols.Discussion)
		case "Commit":
			typeIcon = styles.Commit.Render(symbols.Commit)
		default:
			typeIcon = styles.Commit.Render(symbols.Dot)
		}

		// Render title with smart truncation
		title := n.GetSubject().GetTitle()
		maxTitleLen := leftWidth - 10
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-3] + symbols.Ellipsis
		}

		// Join all parts
		line := fmt.Sprintf("%s %s %s",
			indicator,
			typeIcon,
			title,
		)

		leftSb.WriteString(itemStyle.Render(line) + "\n")
	}

	// Render the details on the right
	n := m.getSelectedNotification()
	var rightContent string
	if n == nil {
		rightContent = styles.NoNotifications.Render("No notification selected")
	} else {
		var rightSb strings.Builder

		// Render header with repository and type
		header := fmt.Sprintf("%s %s",
			n.GetRepository().GetFullName(),
			n.GetSubject().GetType(),
		)
		rightSb.WriteString(styles.DetailHeader.Render(header) + "\n\n")

		// Render title
		title := n.GetSubject().GetTitle()
		rightSb.WriteString(lipgloss.NewStyle().Bold(true).Render(title) + "\n\n")

		// Render status
		var status string
		if n.GetUnread() {
			status = styles.UnreadIndicator.Render("Unread")
		} else {
			status = styles.ReadIndicator.Render("Read")
		}
		rightSb.WriteString(status + "\n\n")

		// Render updated time
		updated := fmt.Sprintf("Updated: %s", n.GetUpdatedAt().Format(time.RFC1123))
		rightSb.WriteString(updated + "\n\n")

		// Render URL
		url := m.getNotificationURL()
		rightSb.WriteString(fmt.Sprintf("URL: %s\n\n", url))

		rightContent = rightSb.String()
	}

	// Render both panels
	left := styles.SplitLeft.
		Width(leftWidth).
		Height(height).
		Render(leftSb.String())

	right := styles.SplitRight.
		Width(rightWidth).
		Height(height).
		Render(rightContent)

	divider := styles.SplitDivider.Render(strings.Repeat("│", height))

	return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, right)
}

// renderTableView renders notifications in a table format
func (m Model) renderTableView(styles Styles, symbols Symbols) string {
	// Define column widths
	idWidth := 3
	typeWidth := 6
	repoWidth := 20
	titleWidth := m.width - idWidth - typeWidth - repoWidth - 20 // Account for padding and time
	timeWidth := 16

	// Render header
	header := lipgloss.JoinHorizontal(lipgloss.Top,
		styles.TableHeader.Width(idWidth).Render("#"),
		styles.TableHeader.Width(typeWidth).Render("Type"),
		styles.TableHeader.Width(repoWidth).Render("Repository"),
		styles.TableHeader.Width(titleWidth).Render("Title"),
		styles.TableHeader.Width(timeWidth).Render("Updated"),
	)

	// Render rows
	var rows []string
	for i, n := range m.filteredItems {
		var rowStyle lipgloss.Style
		if i == m.selected {
			rowStyle = styles.TableSelectedRow
		} else {
			rowStyle = styles.TableCell
		}

		// Render ID
		id := fmt.Sprintf("%d", i+1)
		idCell := rowStyle.Copy().Width(idWidth).Render(id)

		// Render type with icon
		var typeIcon string
		switch n.GetSubject().GetType() {
		case "Issue":
			typeIcon = symbols.Issue
		case "PullRequest":
			typeIcon = symbols.PullRequest
		case "Release":
			typeIcon = symbols.Release
		case "Discussion":
			typeIcon = symbols.Discussion
		case "Commit":
			typeIcon = symbols.Commit
		default:
			typeIcon = symbols.Dot
		}
		typeCell := rowStyle.Copy().Width(typeWidth).Render(typeIcon)

		// Render repository with truncation
		repo := n.GetRepository().GetFullName()
		if len(repo) > repoWidth-3 {
			repo = repo[:repoWidth-3] + symbols.Ellipsis
		}
		repoCell := rowStyle.Copy().Width(repoWidth).Render(repo)

		// Render title with truncation
		title := n.GetSubject().GetTitle()
		if len(title) > titleWidth-3 {
			title = title[:titleWidth-3] + symbols.Ellipsis
		}
		titleCell := rowStyle.Copy().Width(titleWidth).Render(title)

		// Render time
		timeStr := formatTimeForView(n.GetUpdatedAt().Time)
		timeCell := rowStyle.Copy().Width(timeWidth).Render(timeStr)

		// Join cells into a row
		row := lipgloss.JoinHorizontal(lipgloss.Top,
			idCell, typeCell, repoCell, titleCell, timeCell)
		rows = append(rows, row)
	}

	// Join header and rows
	table := lipgloss.JoinVertical(lipgloss.Left,
		header,
		strings.Join(rows, "\n"),
	)

	return styles.Table.
		Width(m.width - 4).
		Height(m.height - 6).
		Render(table)
}

// formatTimeForView formats a time.Time into a human-readable string for the view
func formatTimeForView(t time.Time) string {
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
