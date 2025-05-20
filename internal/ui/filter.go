package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// FilterModel represents the filter input model
type FilterModel struct {
	textInput      textinput.Model
	notifications  []*github.Notification
	filteredItems  []*github.Notification
	filterActive   bool
	width          int
	height         int
	styles         Styles
	onFilterChange func(string, []*github.Notification)
	onExit         func()
}

// NewFilterModel creates a new filter model
func NewFilterModel(notifications []*github.Notification, styles Styles, onFilterChange func(string, []*github.Notification), onExit func()) FilterModel {
	ti := textinput.New()
	ti.Placeholder = "Type to filter notifications..."
	ti.Focus()
	ti.Width = 40
	ti.Prompt = "Filter: "
	ti.PromptStyle = styles.FilterPrompt
	ti.TextStyle = styles.FilterInput

	return FilterModel{
		textInput:      ti,
		notifications:  notifications,
		filteredItems:  notifications,
		filterActive:   true,
		styles:         styles,
		onFilterChange: onFilterChange,
		onExit:         onExit,
	}
}

// Init initializes the filter model
func (m FilterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles events for the filter model
func (m FilterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc:
			// Exit filter mode
			m.filterActive = false
			if m.onExit != nil {
				m.onExit()
			}
			return m, nil
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = m.width - 20
	}

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)

	// Filter notifications based on input
	m.filterNotifications(m.textInput.Value())

	return m, cmd
}

// View renders the filter model
func (m FilterModel) View() string {
	if !m.filterActive {
		return ""
	}

	// Render filter input
	inputView := m.textInput.View()

	// Render filter results preview
	var previewView string
	if m.textInput.Value() != "" {
		previewView = m.renderFilterPreview()
	}

	// Render help text
	helpView := m.styles.HelpBar.Render("Enter: Apply filter • Esc: Cancel")

	// Join all components
	return lipgloss.JoinVertical(lipgloss.Left,
		inputView,
		"",
		previewView,
		"",
		helpView,
	)
}

// renderFilterPreview renders a preview of the filtered results
func (m FilterModel) renderFilterPreview() string {
	if len(m.filteredItems) == 0 {
		return m.styles.NoNotifications.Render("No matching notifications")
	}

	// Show a preview of the first few filtered items
	maxPreviewItems := 5
	if len(m.filteredItems) < maxPreviewItems {
		maxPreviewItems = len(m.filteredItems)
	}

	var sb strings.Builder
	sb.WriteString(m.styles.DetailHeader.Render(
		lipgloss.JoinHorizontal(lipgloss.Center,
			"Filter Preview ",
			m.styles.StatusBar.Render(
				lipgloss.JoinHorizontal(lipgloss.Center,
					"(",
					m.styles.UnreadIndicator.Render(
						lipgloss.JoinHorizontal(lipgloss.Center,
							"Showing ",
							lipgloss.NewStyle().Bold(true).Render(
								lipgloss.JoinHorizontal(lipgloss.Center,
									"1-",
									lipgloss.NewStyle().Bold(true).Render(
										lipgloss.JoinHorizontal(lipgloss.Center,
											string(rune(maxPreviewItems+'0')),
											" of ",
											string(rune(len(m.filteredItems)+'0')),
										),
									),
								),
							),
							" results",
						),
					),
					")",
				),
			),
		),
	))
	sb.WriteString("\n\n")

	for i := 0; i < maxPreviewItems; i++ {
		n := m.filteredItems[i]

		// Render repository name
		repo := n.GetRepository().GetFullName()

		// Render title with smart truncation
		title := n.GetSubject().GetTitle()
		maxTitleLen := m.width - len(repo) - 10
		if len(title) > maxTitleLen && maxTitleLen > 3 {
			title = title[:maxTitleLen-3] + "..."
		}

		// Highlight matching text
		filter := strings.ToLower(m.textInput.Value())
		if filter != "" {
			// Highlight in repo
			repoLower := strings.ToLower(repo)
			if idx := strings.Index(repoLower, filter); idx >= 0 {
				repo = repo[:idx] +
					m.styles.FilterPrompt.Render(repo[idx:idx+len(filter)]) +
					repo[idx+len(filter):]
			}

			// Highlight in title
			titleLower := strings.ToLower(title)
			if idx := strings.Index(titleLower, filter); idx >= 0 {
				title = title[:idx] +
					m.styles.FilterPrompt.Render(title[idx:idx+len(filter)]) +
					title[idx+len(filter):]
			}
		}

		// Join all parts
		line := lipgloss.JoinHorizontal(lipgloss.Left,
			m.styles.ListItem.Render(repo + ": "),
			m.styles.ListItem.Render(title),
		)

		sb.WriteString(line + "\n")
	}

	if len(m.filteredItems) > maxPreviewItems {
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render("...and more"))
	}

	return sb.String()
}

// filterNotifications filters notifications based on the filter string
func (m *FilterModel) filterNotifications(filter string) {
	if filter == "" {
		m.filteredItems = m.notifications
		if m.onFilterChange != nil {
			m.onFilterChange(filter, m.filteredItems)
		}
		return
	}

	filter = strings.ToLower(filter)
	var filtered []*github.Notification

	for _, n := range m.notifications {
		// Search in repository name
		repoName := strings.ToLower(n.GetRepository().GetFullName())
		// Search in title
		title := strings.ToLower(n.GetSubject().GetTitle())
		// Search in type
		typeName := strings.ToLower(n.GetSubject().GetType())

		if strings.Contains(repoName, filter) ||
		   strings.Contains(title, filter) ||
		   strings.Contains(typeName, filter) {
			filtered = append(filtered, n)
		}
	}

	m.filteredItems = filtered
	if m.onFilterChange != nil {
		m.onFilterChange(filter, m.filteredItems)
	}
}

// IsActive returns whether the filter is active
func (m FilterModel) IsActive() bool {
	return m.filterActive
}

// GetFilteredItems returns the filtered items
func (m FilterModel) GetFilteredItems() []*github.Notification {
	return m.filteredItems
}

// GetFilterString returns the current filter string
func (m FilterModel) GetFilterString() string {
	return m.textInput.Value()
}
