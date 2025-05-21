package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpKeyMap defines the keybindings for the help model
type HelpKeyMap struct {
	Quit key.Binding
}

// NewHelpKeyMap creates a new key map for the help model
func NewHelpKeyMap() HelpKeyMap {
	return HelpKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "close help"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k HelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k HelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit},
	}
}

// HelpModel represents the help model
type HelpModel struct {
	viewport viewport.Model
	keys     HelpKeyMap
	width    int
	height   int
	visible  bool
	context  string
}

// NewHelpModel creates a new help model
func NewHelpModel(width, height int) HelpModel {
	vp := viewport.New(width, height-4)
	vp.SetContent(getHelpContent("main"))

	return HelpModel{
		viewport: vp,
		keys:     NewHelpKeyMap(),
		width:    width,
		height:   height,
		visible:  false,
		context:  "main",
	}
}

// SetSize sets the size of the help model
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 4
}

// SetContext sets the context for the help content
func (m *HelpModel) SetContext(context string) {
	if m.context != context {
		m.context = context
		m.viewport.SetContent(getHelpContent(context))
		m.viewport.GotoTop()
	}
}

// Toggle toggles the visibility of the help model
func (m *HelpModel) Toggle() {
	m.visible = !m.visible
}

// Show shows the help model
func (m *HelpModel) Show() {
	m.visible = true
}

// Hide hides the help model
func (m *HelpModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the help model is visible
func (m *HelpModel) IsVisible() bool {
	return m.visible
}

// Update updates the help model
func (m *HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.visible = false
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return *m, cmd
}

// View renders the help model
func (m *HelpModel) View() string {
	if !m.visible {
		return ""
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1).
		Render(fmt.Sprintf("Help: %s", strings.Title(m.context)))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press q to close help, ↑/↓ to scroll")

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(m.width - 4).
		Render(fmt.Sprintf("%s\n\n%s\n\n%s", title, m.viewport.View(), footer))
}

// getHelpContent returns the help content for the given context
func getHelpContent(context string) string {
	switch context {
	case "main":
		return mainHelpContent
	case "list":
		return listHelpContent
	case "detail":
		return detailHelpContent
	case "filter":
		return filterHelpContent
	case "group":
		return groupHelpContent
	case "search":
		return searchHelpContent
	case "watch":
		return watchHelpContent
	case "action":
		return actionHelpContent
	default:
		return mainHelpContent
	}
}

// Help content for different contexts
const mainHelpContent = `
# gh-notif Help

gh-notif is a high-performance CLI tool for managing GitHub notifications in the terminal.

## Navigation

- Use arrow keys or j/k to navigate
- Press Enter to select or expand
- Press Esc to go back
- Press q to quit

## Common Commands

- r: Mark as read
- o: Open in browser
- f: Filter notifications
- g: Group notifications
- s: Search notifications
- ?: Show this help

## Views

- 1: Switch to compact view
- 2: Switch to detailed view
- 3: Switch to split view
- 4: Switch to table view

## Other

- Ctrl+R: Refresh notifications
- Ctrl+S: Save current view as a filter
- Ctrl+F: Search within the current view
- Ctrl+Z: Undo last action
`

const listHelpContent = `
# Notification List Help

This view shows a list of your GitHub notifications.

## Navigation

- ↑/k: Move up
- ↓/j: Move down
- PgUp/Ctrl+B: Page up
- PgDn/Ctrl+F: Page down
- Home/g: Go to top
- End/G: Go to bottom
- Enter: View notification details
- Esc/h: Go back

## Actions

- r: Mark as read
- u: Mark as unread
- o: Open in browser
- a: Archive
- s: Subscribe
- S: Unsubscribe
- m: Mute repository
- M: Unmute repository
- Space: Select notification
- A: Select all
- N: Select none
- I: Invert selection

## Filtering and Grouping

- f: Filter notifications
- F: Clear filter
- g: Group notifications
- G: Clear grouping
- t: Sort notifications
- T: Reverse sort order

## Views

- 1: Switch to compact view
- 2: Switch to detailed view
- 3: Switch to split view
- 4: Switch to table view

## Search

- /: Search
- n: Next search result
- N: Previous search result
`

const detailHelpContent = `
# Notification Detail Help

This view shows the details of a notification.

## Navigation

- ↑/k: Scroll up
- ↓/j: Scroll down
- PgUp/Ctrl+B: Page up
- PgDn/Ctrl+F: Page down
- Home/g: Go to top
- End/G: Go to bottom
- Esc/h: Go back to list

## Actions

- r: Mark as read
- u: Mark as unread
- o: Open in browser
- a: Archive
- s: Subscribe
- S: Unsubscribe
- c: Copy URL
- y: Copy ID

## Content

- e: Expand/collapse sections
- p: Preview in Markdown
- d: Show diff (for pull requests)
- i: Show issue details
- t: Show timeline
- m: Show comments
- w: Watch thread
`

const filterHelpContent = `
# Filter Help

This view allows you to create and apply filters to your notifications.

## Navigation

- Tab: Next field
- Shift+Tab: Previous field
- Enter: Apply filter
- Esc: Cancel

## Actions

- Ctrl+Space: Show autocomplete
- Alt+S: Save filter
- Alt+L: Load filter
- Alt+C: Clear filter
- Alt+H: Show filter history
- Alt+E: Edit filter as text
- Alt+V: Validate filter

## Filter Syntax

- repo:owner/repo - Notifications from a specific repository
- org:organization - Notifications from a specific organization
- type:PullRequest - Pull request notifications
- type:Issue - Issue notifications
- reason:mention - Notifications where you were mentioned
- is:unread - Unread notifications
- is:read - Read notifications
- author:username - Notifications from a specific author
- involves:username - Notifications involving a specific user
- label:bug - Notifications with a specific label
- state:open - Notifications for open issues/PRs
- state:closed - Notifications for closed issues/PRs
- created:>2023-01-01 - Notifications created after a date
- updated:<2023-01-01 - Notifications updated before a date

## Combining Filters

- AND - Both conditions must be true
- OR - Either condition can be true
- NOT - Negates a condition
- () - Groups conditions

## Examples

- repo:owner/repo AND type:PullRequest AND is:unread
- (type:PullRequest OR type:Issue) AND is:unread
- repo:owner/repo AND NOT type:Issue
- @my-prs - Use a saved filter
`

const groupHelpContent = `
# Group Help

This view shows notifications grouped by various criteria.

## Navigation

- ↑/k: Move up
- ↓/j: Move down
- Enter: Expand/collapse group
- Space: Select/deselect group
- Esc/h: Go back

## Actions

- r: Mark group as read
- a: Archive group
- s: Subscribe to group
- S: Unsubscribe from group
- o: Open all in group

## Grouping

- g: Change grouping
- f: Filter within group
- t: Sort group
- c: Collapse all groups
- e: Expand all groups

## Grouping Criteria

- repository: Group by repository
- owner: Group by repository owner
- type: Group by notification type
- reason: Group by notification reason
- author: Group by author
- state: Group by state (open/closed)
- smart: Use smart grouping algorithm
`

const searchHelpContent = `
# Search Help

This view allows you to search across your notifications.

## Navigation

- Tab: Next field
- Shift+Tab: Previous field
- Enter: Execute search
- Esc: Cancel search

## Search Options

- Alt+R: Toggle regex mode
- Alt+C: Toggle case sensitivity
- Alt+W: Toggle whole word
- Alt+S: Save search
- Alt+H: Show search history
- Alt+A: Advanced search options

## Search Syntax

- Simple text: Searches notification titles and bodies
- "exact phrase": Searches for an exact phrase
- author:username: Searches for notifications from a specific author
- repo:owner/repo: Limits search to a specific repository
- type:PullRequest: Limits search to pull requests
- type:Issue: Limits search to issues

## Examples

- bug fix
- "security vulnerability"
- author:octocat bug
- repo:owner/repo feature
`

const watchHelpContent = `
# Watch Help

This view shows real-time updates for your notifications.

## Controls

- Space: Pause/resume watching
- r: Mark as read
- o: Open in browser
- f: Filter notifications
- i: Show notification details
- d: Toggle desktop notifications
- +: Increase refresh interval
- -: Decrease refresh interval
- c: Clear notifications
- s: Show statistics
- Esc/q: Quit watch mode

## Watch Options

- Filter: Only watch notifications matching a filter
- Interval: How often to check for new notifications
- Desktop Notifications: Show desktop notifications for new items
- Timeout: Automatically stop watching after a period
- Backoff: Gradually increase interval if no new notifications
`

const actionHelpContent = `
# Action Help

This view allows you to perform actions on multiple notifications.

## Navigation

- Tab: Next option
- Shift+Tab: Previous option
- Enter: Execute action
- Esc: Cancel action

## Selection

- a: Select all
- n: Select none
- i: Invert selection
- f: Filter selection
- s: Save selection
- l: Load selection

## Actions

- Mark as read: Mark selected notifications as read
- Mark as unread: Mark selected notifications as unread
- Archive: Archive selected notifications
- Unarchive: Unarchive selected notifications
- Subscribe: Subscribe to selected notification threads
- Unsubscribe: Unsubscribe from selected notification threads
- Open: Open selected notifications in browser
- Mute: Mute repositories of selected notifications
- Unmute: Unmute repositories of selected notifications
`
