package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/browser"
	"github.com/SharanRP/gh-notif/internal/common"
	"github.com/SharanRP/gh-notif/internal/operations"
)

// updateSelectMode handles updates in select mode
func (m ActionModel) updateSelectMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, m.keyMap.ToggleSelect):
		// Toggle selection of the current item
		i := m.list.Index()
		if i < 0 || i >= len(m.notifications) {
			return m, nil
		}

		// Get the notification
		notification := m.notifications[i]
		id := notification.GetID()

		// Toggle selection
		if m.selected[id] {
			m.selected[id] = false
			m.selectedCount--
		} else {
			m.selected[id] = true
			m.selectedCount++
		}

		// Update the status bar
		m.statusBar.text = fmt.Sprintf("%d/%d notifications selected", m.selectedCount, len(m.notifications))

		// Update the list item
		items := m.list.Items()
		items[i] = notificationItem{
			notification: notification,
			selected:     m.selected[id],
		}
		m.list.SetItems(items)

		return m, nil

	case key.Matches(msg, m.keyMap.SelectAll):
		// Select all notifications
		for i, n := range m.notifications {
			id := n.GetID()
			m.selected[id] = true

			// Update the list item
			items := m.list.Items()
			items[i] = notificationItem{
				notification: n,
				selected:     true,
			}
			m.list.SetItems(items)
		}
		m.selectedCount = len(m.notifications)
		m.statusBar.text = fmt.Sprintf("%d/%d notifications selected", m.selectedCount, len(m.notifications))
		return m, nil

	case key.Matches(msg, m.keyMap.DeselectAll):
		// Deselect all notifications
		for i, n := range m.notifications {
			id := n.GetID()
			m.selected[id] = false

			// Update the list item
			items := m.list.Items()
			items[i] = notificationItem{
				notification: n,
				selected:     false,
			}
			m.list.SetItems(items)
		}
		m.selectedCount = 0
		m.statusBar.text = fmt.Sprintf("%d/%d notifications selected", m.selectedCount, len(m.notifications))
		return m, nil

	case key.Matches(msg, m.keyMap.MarkAsRead):
		// Mark selected notifications as read
		if m.selectedCount == 0 {
			// If no notifications are selected, mark the current one
			i := m.list.Index()
			if i < 0 || i >= len(m.notifications) {
				return m, nil
			}
			notification := m.notifications[i]
			m.selected[notification.GetID()] = true
			m.selectedCount = 1
		}

		// Switch to action menu mode
		m.mode = ActionMenuMode
		return m, nil

	case key.Matches(msg, m.keyMap.Archive):
		// Archive selected notifications
		if m.selectedCount == 0 {
			// If no notifications are selected, archive the current one
			i := m.list.Index()
			if i < 0 || i >= len(m.notifications) {
				return m, nil
			}
			notification := m.notifications[i]
			m.selected[notification.GetID()] = true
			m.selectedCount = 1
		}

		// Switch to action menu mode
		m.mode = ActionMenuMode
		return m, nil

	case key.Matches(msg, m.keyMap.Subscribe):
		// Subscribe to selected threads
		if m.selectedCount == 0 {
			// If no notifications are selected, subscribe to the current one
			i := m.list.Index()
			if i < 0 || i >= len(m.notifications) {
				return m, nil
			}
			notification := m.notifications[i]
			m.selected[notification.GetID()] = true
			m.selectedCount = 1
		}

		// Switch to action menu mode
		m.mode = ActionMenuMode
		return m, nil

	case key.Matches(msg, m.keyMap.Unsubscribe):
		// Unsubscribe from selected threads
		if m.selectedCount == 0 {
			// If no notifications are selected, unsubscribe from the current one
			i := m.list.Index()
			if i < 0 || i >= len(m.notifications) {
				return m, nil
			}
			notification := m.notifications[i]
			m.selected[notification.GetID()] = true
			m.selectedCount = 1
		}

		// Switch to action menu mode
		m.mode = ActionMenuMode
		return m, nil

	case key.Matches(msg, m.keyMap.Mute):
		// Mute repositories of selected notifications
		if m.selectedCount == 0 {
			// If no notifications are selected, mute the current one's repository
			i := m.list.Index()
			if i < 0 || i >= len(m.notifications) {
				return m, nil
			}
			notification := m.notifications[i]
			m.selected[notification.GetID()] = true
			m.selectedCount = 1
		}

		// Switch to action menu mode
		m.mode = ActionMenuMode
		return m, nil

	case key.Matches(msg, m.keyMap.Open):
		// Open the current notification in the browser
		i := m.list.Index()
		if i < 0 || i >= len(m.notifications) {
			return m, nil
		}
		notification := m.notifications[i]
		url := notification.GetSubject().GetURL()

		// Convert API URL to web URL
		webURL := convertAPIURLToWebURL(url)

		// Open the URL in the browser
		return m, func() tea.Msg {
			browser.OpenURL(webURL)
			return nil
		}
	}

	// Update the list
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateActionMenuMode handles updates in action menu mode
func (m ActionModel) updateActionMenuMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Back):
		// Go back to select mode
		m.mode = SelectMode
		return m, nil

	case key.Matches(msg, m.keyMap.MarkAsRead):
		// Mark selected notifications as read
		return m, m.performMarkAsRead()

	case key.Matches(msg, m.keyMap.Archive):
		// Archive selected notifications
		return m, m.performArchive()

	case key.Matches(msg, m.keyMap.Subscribe):
		// Subscribe to selected threads
		return m, m.performSubscribe()

	case key.Matches(msg, m.keyMap.Unsubscribe):
		// Unsubscribe from selected threads
		return m, m.performUnsubscribe()

	case key.Matches(msg, m.keyMap.Mute):
		// Mute repositories of selected notifications
		return m, m.performMute()
	}

	return m, nil
}

// updateProgressMode handles updates in progress mode
func (m ActionModel) updateProgressMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Cancel):
		// Cancel the operation
		m.cancelFunc()

		// Create a new context
		ctx, cancel := context.WithCancel(context.Background())
		m.ctx = ctx
		m.cancelFunc = cancel

		// Go back to select mode
		m.mode = SelectMode
		return m, nil
	}

	return m, nil
}

// updateResultMode handles updates in result mode
func (m ActionModel) updateResultMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Back), key.Matches(msg, m.keyMap.Select):
		// Go back to select mode
		m.mode = SelectMode
		return m, nil
	}

	return m, nil
}

// performMarkAsRead marks selected notifications as read
func (m ActionModel) performMarkAsRead() tea.Cmd {
	// Get selected notification IDs
	ids := m.getSelectedIDs()
	if len(ids) == 0 {
		return nil
	}

	// Switch to progress mode
	m.mode = ProgressMode

	// Create batch options
	opts := &common.BatchOptions{
		Concurrency: 5,
		ProgressCallback: func(completed, total int) {
			// This will be called from a goroutine, so we need to send a message
			// to update the UI
			// In a real implementation, we would send a message to update the progress
		},
		ContinueOnError: true,
		Timeout:         30 * time.Second,
	}

	// Perform the operation in a goroutine
	return func() tea.Msg {
		result, _ := operations.MarkMultipleAsRead(m.ctx, ids, opts)
		return BatchResultMsg{Result: result}
	}
}

// performArchive archives selected notifications
func (m ActionModel) performArchive() tea.Cmd {
	// Get selected notification IDs
	ids := m.getSelectedIDs()
	if len(ids) == 0 {
		return nil
	}

	// Switch to progress mode
	m.mode = ProgressMode

	// Create batch options
	opts := &common.BatchOptions{
		Concurrency: 5,
		ProgressCallback: func(completed, total int) {
			// This will be called from a goroutine, so we need to send a message
			// to update the UI
			// In a real implementation, we would send a message to update the progress
		},
		ContinueOnError: true,
		Timeout:         30 * time.Second,
	}

	// Perform the operation in a goroutine
	return func() tea.Msg {
		result, _ := operations.ArchiveMultipleNotifications(m.ctx, ids, opts)
		return BatchResultMsg{Result: result}
	}
}

// performSubscribe subscribes to selected threads
func (m ActionModel) performSubscribe() tea.Cmd {
	// Get selected notification IDs
	ids := m.getSelectedIDs()
	if len(ids) == 0 {
		return nil
	}

	// Switch to progress mode
	m.mode = ProgressMode

	// Create batch options
	opts := &common.BatchOptions{
		Concurrency: 5,
		ProgressCallback: func(completed, total int) {
			// This will be called from a goroutine, so we need to send a message
			// to update the UI
			// In a real implementation, we would send a message to update the progress
		},
		ContinueOnError: true,
		Timeout:         30 * time.Second,
	}

	// Perform the operation in a goroutine
	return func() tea.Msg {
		result, _ := operations.SubscribeToMultipleThreads(m.ctx, ids, opts)
		return BatchResultMsg{Result: result}
	}
}

// performUnsubscribe unsubscribes from selected threads
func (m ActionModel) performUnsubscribe() tea.Cmd {
	// Get selected notification IDs
	ids := m.getSelectedIDs()
	if len(ids) == 0 {
		return nil
	}

	// Switch to progress mode
	m.mode = ProgressMode

	// Create batch options
	opts := &common.BatchOptions{
		Concurrency: 5,
		ProgressCallback: func(completed, total int) {
			// This will be called from a goroutine, so we need to send a message
			// to update the UI
			// In a real implementation, we would send a message to update the progress
		},
		ContinueOnError: true,
		Timeout:         30 * time.Second,
	}

	// Perform the operation in a goroutine
	return func() tea.Msg {
		result, _ := operations.UnsubscribeFromMultipleThreads(m.ctx, ids, opts)
		return BatchResultMsg{Result: result}
	}
}

// performMute mutes repositories of selected notifications
func (m ActionModel) performMute() tea.Cmd {
	// Get selected notification IDs
	ids := m.getSelectedIDs()
	if len(ids) == 0 {
		return nil
	}

	// Get unique repository names
	repos := make(map[string]bool)
	for _, id := range ids {
		for _, n := range m.notifications {
			if n.GetID() == id {
				repos[n.GetRepository().GetFullName()] = true
				break
			}
		}
	}

	// Convert to slice
	repoNames := make([]string, 0, len(repos))
	for repo := range repos {
		repoNames = append(repoNames, repo)
	}

	// Switch to progress mode
	m.mode = ProgressMode

	// Create batch options
	opts := &common.BatchOptions{
		Concurrency: 5,
		ProgressCallback: func(completed, total int) {
			// This will be called from a goroutine, so we need to send a message
			// to update the UI
			// In a real implementation, we would send a message to update the progress
		},
		ContinueOnError: true,
		Timeout:         30 * time.Second,
	}

	// Perform the operation in a goroutine
	return func() tea.Msg {
		result, _ := operations.MuteMultipleRepositories(m.ctx, repoNames, opts)
		return BatchResultMsg{Result: result}
	}
}

// getSelectedIDs returns the IDs of selected notifications
func (m ActionModel) getSelectedIDs() []string {
	ids := make([]string, 0, m.selectedCount)
	for _, n := range m.notifications {
		id := n.GetID()
		if m.selected[id] {
			ids = append(ids, id)
		}
	}
	return ids
}

// convertAPIURLToWebURL converts a GitHub API URL to a web URL
func convertAPIURLToWebURL(apiURL string) string {
	// This is a simplified implementation
	// In a real implementation, we would parse the URL and convert it properly
	return strings.Replace(apiURL, "api.github.com", "github.com", 1)
}
