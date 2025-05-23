package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/SharanRP/gh-notif/internal/discovery"
)

// TipModel represents a feature tip
type TipModel struct {
	featureID   string
	title       string
	description string
	shortcut    string
	width       int
	height      int
	visible     bool
	dismissed   bool
	manager     *discovery.DiscoveryManager
}

// NewTipModel creates a new tip model
func NewTipModel(featureID, title, description, shortcut string, manager *discovery.DiscoveryManager) TipModel {
	return TipModel{
		featureID:   featureID,
		title:       title,
		description: description,
		shortcut:    shortcut,
		width:       40,
		height:      10,
		visible:     false,
		dismissed:   false,
		manager:     manager,
	}
}

// SetSize sets the size of the tip model
func (m *TipModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Show shows the tip
func (m *TipModel) Show() {
	if m.manager != nil && m.manager.ShouldShowFeatureTip(m.featureID) {
		m.visible = true
	}
}

// Hide hides the tip
func (m *TipModel) Hide() {
	m.visible = false
}

// Dismiss dismisses the tip
func (m *TipModel) Dismiss() {
	m.dismissed = true
	m.visible = false
	if m.manager != nil {
		m.manager.DismissFeatureTip(m.featureID)
	}
}

// IsVisible returns whether the tip is visible
func (m *TipModel) IsVisible() bool {
	return m.visible && !m.dismissed
}

// Update updates the tip model
func (m *TipModel) Update(msg tea.Msg) (TipModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "enter", "space"))):
			m.Dismiss()
		}
	}

	return *m, nil
}

// View renders the tip model
func (m *TipModel) View() string {
	if !m.IsVisible() {
		return ""
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1).
		Render(fmt.Sprintf("💡 %s", m.title))

	description := lipgloss.NewStyle().
		Padding(0, 1).
		Render(m.description)

	shortcut := ""
	if m.shortcut != "" {
		shortcut = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Padding(0, 1).
			Render(fmt.Sprintf("Shortcut: %s", m.shortcut))
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press Enter to dismiss")

	content := fmt.Sprintf("%s\n\n%s", title, description)
	if shortcut != "" {
		content = fmt.Sprintf("%s\n\n%s", content, shortcut)
	}
	content = fmt.Sprintf("%s\n\n%s", content, footer)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(m.width - 4).
		Render(content)
}

// FeatureTips contains all feature tips
var FeatureTips = map[string]struct {
	Title       string
	Description string
	Shortcut    string
}{
	"filter.basic": {
		Title:       "Filter Notifications",
		Description: "You can filter notifications by repository, type, and more.",
		Shortcut:    "f",
	},
	"filter.complex": {
		Title:       "Advanced Filtering",
		Description: "Use complex filter expressions with AND, OR, NOT operators.",
		Shortcut:    "f",
	},
	"filter.save": {
		Title:       "Save Filters",
		Description: "Save filters for later use with the @name syntax.",
		Shortcut:    "Ctrl+S",
	},
	"group.basic": {
		Title:       "Group Notifications",
		Description: "Group notifications by repository, type, or other criteria.",
		Shortcut:    "g",
	},
	"group.smart": {
		Title:       "Smart Grouping",
		Description: "Use smart grouping to automatically organize related notifications.",
		Shortcut:    "g then s",
	},
	"search.basic": {
		Title:       "Search Notifications",
		Description: "Search across all notification content.",
		Shortcut:    "/",
	},
	"search.regex": {
		Title:       "Regex Search",
		Description: "Use regular expressions for more powerful searches.",
		Shortcut:    "/ then Alt+R",
	},
	"watch.basic": {
		Title:       "Watch Notifications",
		Description: "Watch for new notifications in real-time.",
		Shortcut:    "w",
	},
	"watch.desktop": {
		Title:       "Desktop Notifications",
		Description: "Get desktop notifications for new items.",
		Shortcut:    "w then d",
	},
	"ui.keyboard": {
		Title:       "Keyboard Shortcuts",
		Description: "Use keyboard shortcuts for faster navigation.",
		Shortcut:    "?",
	},
	"ui.batch": {
		Title:       "Batch Actions",
		Description: "Select multiple notifications with Space and perform batch actions.",
		Shortcut:    "Space",
	},
	"ui.views": {
		Title:       "View Modes",
		Description: "Switch between different view modes for notifications.",
		Shortcut:    "1-4",
	},
	"actions.read": {
		Title:       "Mark as Read",
		Description: "Mark notifications as read to keep track of what you've seen.",
		Shortcut:    "r",
	},
	"actions.open": {
		Title:       "Open in Browser",
		Description: "Open notifications in your browser to view details.",
		Shortcut:    "o",
	},
	"actions.archive": {
		Title:       "Archive Notifications",
		Description: "Archive notifications to keep your list clean.",
		Shortcut:    "a",
	},
	"actions.subscribe": {
		Title:       "Subscribe to Threads",
		Description: "Subscribe to notification threads to stay updated.",
		Shortcut:    "s",
	},
	"actions.undo": {
		Title:       "Undo Actions",
		Description: "Undo your last action if you make a mistake.",
		Shortcut:    "Ctrl+Z",
	},
	"config.basic": {
		Title:       "Configuration",
		Description: "Configure gh-notif to suit your preferences.",
		Shortcut:    "F2",
	},
	"tutorial.basic": {
		Title:       "Interactive Tutorial",
		Description: "Learn how to use gh-notif with the interactive tutorial.",
		Shortcut:    "F1",
	},
	"help.contextual": {
		Title:       "Contextual Help",
		Description: "Get context-specific help for the current view.",
		Shortcut:    "?",
	},
}

// CreateTipModels creates tip models for all feature tips
func CreateTipModels(manager *discovery.DiscoveryManager) map[string]TipModel {
	tips := make(map[string]TipModel)
	for id, tip := range FeatureTips {
		tips[id] = NewTipModel(id, tip.Title, tip.Description, tip.Shortcut, manager)
	}
	return tips
}
