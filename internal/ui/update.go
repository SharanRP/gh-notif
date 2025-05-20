package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
	"github.com/pkg/browser"
)

// Messages for UI updates
type (
	// ErrorMsg represents an error message
	ErrorMsg struct{ err error }

	// LoadingMsg indicates loading state
	LoadingMsg struct{ loading bool }

	// FilterMsg updates the filter string
	FilterMsg struct{ filter string }

	// ViewModeMsg changes the view mode
	ViewModeMsg struct{ mode ViewMode }

	// ColorSchemeMsg changes the color scheme
	ColorSchemeMsg struct{ scheme ColorScheme }

	// MarkAsReadMsg marks a notification as read
	MarkAsReadMsg struct{ id string }

	// MarkAllAsReadMsg marks all notifications as read
	MarkAllAsReadMsg struct{}

	// RefreshMsg refreshes the notifications
	RefreshMsg struct{}
)

// Update handles UI events and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key presses
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keyMap.Up):
			if m.selected > 0 {
				m.selected--
				m.viewport.SetYOffset(m.selected * 2) // Adjust based on item height
			}

		case key.Matches(msg, m.keyMap.Down):
			if m.selected < len(m.filteredItems)-1 {
				m.selected++
				m.viewport.SetYOffset(m.selected * 2) // Adjust based on item height
			}

		case key.Matches(msg, m.keyMap.Select):
			if n := m.getSelectedNotification(); n != nil {
				// Open the notification in the browser
				url := m.getNotificationURL()
				if url != "" {
					cmds = append(cmds, openBrowserCmd(url))
				}
			}

		case key.Matches(msg, m.keyMap.MarkAsRead):
			if n := m.getSelectedNotification(); n != nil {
				// Mark the notification as read
				cmds = append(cmds, markAsReadCmd(n.GetID()))
			}

		case key.Matches(msg, m.keyMap.MarkAllAsRead):
			// Mark all notifications as read
			cmds = append(cmds, markAllAsReadCmd())

		case key.Matches(msg, m.keyMap.Filter):
			// Enter filter mode
			// This would typically activate a text input component
			// For now, we'll just toggle a filter
			if m.filterString == "" {
				m.filterString = "issue" // Example filter
			} else {
				m.filterString = ""
			}
			m.filterNotifications()

		case key.Matches(msg, m.keyMap.ViewMode):
			// Cycle through view modes
			m.viewMode = (m.viewMode + 1) % 4

		case key.Matches(msg, m.keyMap.ColorScheme):
			// Cycle through color schemes
			m.colorScheme = (m.colorScheme + 1) % 3

		case key.Matches(msg, m.keyMap.OpenInBrowser):
			if n := m.getSelectedNotification(); n != nil {
				// Open the notification in the browser
				url := m.getNotificationURL()
				if url != "" {
					cmds = append(cmds, openBrowserCmd(url))
				}
			}

		case key.Matches(msg, m.keyMap.Refresh):
			// Refresh notifications
			m.loading = true
			cmds = append(cmds, refreshNotificationsCmd())
		}

	case tea.WindowSizeMsg:
		// Handle window resize
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update viewport dimensions
		headerHeight := 2
		footerHeight := 2
		helpHeight := 0
		if m.showHelp {
			helpHeight = 6
		}

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight - helpHeight

		// Update status bar text
		m.statusBar.text = fmt.Sprintf("%d notifications (%d filtered)",
			len(m.notifications), len(m.filteredItems))

	case spinner.TickMsg:
		// Update spinner
		var spinnerCmd tea.Cmd
		m.spinner, spinnerCmd = m.spinner.Update(msg)
		if m.loading {
			cmds = append(cmds, spinnerCmd)
		}

	case ErrorMsg:
		// Handle error
		m.error = msg.err
		m.loading = false

	case LoadingMsg:
		// Update loading state
		m.loading = msg.loading
		if m.loading {
			cmds = append(cmds, m.spinner.Tick)
		}

	case FilterMsg:
		// Update filter
		m.filterString = msg.filter
		m.filterNotifications()

	case ViewModeMsg:
		// Update view mode
		m.viewMode = msg.mode

	case ColorSchemeMsg:
		// Update color scheme
		m.colorScheme = msg.scheme

	case MarkAsReadMsg:
		// Handle notification marked as read
		for i, n := range m.notifications {
			if n.GetID() == msg.id {
				// Update the notification
				m.notifications[i].Unread = github.Bool(false)
				break
			}
		}
		m.filterNotifications()

	case MarkAllAsReadMsg:
		// Handle all notifications marked as read
		for i := range m.notifications {
			m.notifications[i].Unread = github.Bool(false)
		}
		m.filterNotifications()

	case RefreshMsg:
		// Handle refreshed notifications
		// This would typically update the notifications list
		// For now, we'll just update the status
		m.loading = false
		m.statusBar.text = fmt.Sprintf("%d notifications refreshed", len(m.notifications))
	}

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// openBrowserCmd opens a URL in the browser
func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if err := browser.OpenURL(url); err != nil {
			return ErrorMsg{err: fmt.Errorf("failed to open URL: %w", err)}
		}
		return nil
	}
}

// markAsReadCmd marks a notification as read
func markAsReadCmd(id string) tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would call the GitHub API
		// For now, we'll just return a message
		return MarkAsReadMsg{id: id}
	}
}

// markAllAsReadCmd marks all notifications as read
func markAllAsReadCmd() tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would call the GitHub API
		// For now, we'll just return a message
		return MarkAllAsReadMsg{}
	}
}

// refreshNotificationsCmd refreshes the notifications
func refreshNotificationsCmd() tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would fetch new notifications
		// For now, we'll just return a message
		return RefreshMsg{}
	}
}
