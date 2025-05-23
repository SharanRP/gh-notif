package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// ViewMode represents different view modes for the notification UI
type ViewMode int

const (
	// CompactView shows a compact list of notifications
	CompactView ViewMode = iota
	// DetailedView shows detailed information for a single notification
	DetailedView
	// SplitView shows a list on the left and details on the right
	SplitView
	// TableView shows notifications in a table format
	TableView
)

// ColorScheme represents different color schemes
type ColorScheme int

const (
	// DarkScheme is the default dark color scheme
	DarkScheme ColorScheme = iota
	// LightScheme is a light color scheme
	LightScheme
	// HighContrastScheme is a high contrast scheme for accessibility
	HighContrastScheme
)

// Model represents the main UI model
type Model struct {
	// Data
	notifications []*github.Notification
	selected      int
	filterString  string
	filteredItems []*github.Notification

	// UI Components
	viewport  viewport.Model
	help      help.Model
	spinner   spinner.Model
	keyMap    keyMap
	statusBar StatusBar

	// State
	width       int
	height      int
	ready       bool
	loading     bool
	showHelp    bool
	viewMode    ViewMode
	colorScheme ColorScheme
	error       error
}

// StatusBar represents the status bar at the bottom of the UI
type StatusBar struct {
	text  string
	style lipgloss.Style
}

// keyMap defines the keybindings for the UI
type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	Help          key.Binding
	Quit          key.Binding
	Select        key.Binding
	MarkAsRead    key.Binding
	MarkAllAsRead key.Binding
	Filter        key.Binding
	ViewMode      key.Binding
	ColorScheme   key.Binding
	OpenInBrowser key.Binding
	Refresh       key.Binding
}

// defaultKeyMap returns the default key bindings
func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		MarkAsRead: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mark as read"),
		),
		MarkAllAsRead: key.NewBinding(
			key.WithKeys("M"),
			key.WithHelp("M", "mark all as read"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ViewMode: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "change view"),
		),
		ColorScheme: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "change colors"),
		),
		OpenInBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Select, k.MarkAsRead, k.ViewMode, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.MarkAsRead, k.MarkAllAsRead, k.OpenInBrowser},
		{k.ViewMode, k.ColorScheme, k.Filter, k.Refresh},
		{k.Help, k.Quit},
	}
}

// NewModel creates a new UI model
func NewModel(notifications []*github.Notification) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	h := help.New()
	h.ShowAll = false

	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}

	m := Model{
		notifications: notifications,
		filteredItems: notifications,
		selected:      0,
		help:          h,
		spinner:       s,
		keyMap:        defaultKeyMap(),
		viewport:      vp,
		viewMode:      CompactView,
		colorScheme:   DarkScheme,
		statusBar: StatusBar{
			text:  fmt.Sprintf("%d notifications", len(notifications)),
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		},
	}

	return m
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

// filterNotifications filters notifications based on the filter string
func (m *Model) filterNotifications() {
	if m.filterString == "" {
		m.filteredItems = m.notifications
		return
	}

	filter := strings.ToLower(m.filterString)
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
	if len(filtered) > 0 && m.selected >= len(filtered) {
		m.selected = len(filtered) - 1
	}
}

// getSelectedNotification returns the currently selected notification
func (m Model) getSelectedNotification() *github.Notification {
	if len(m.filteredItems) == 0 || m.selected < 0 || m.selected >= len(m.filteredItems) {
		return nil
	}
	return m.filteredItems[m.selected]
}

// getNotificationURL returns the URL for the selected notification
func (m Model) getNotificationURL() string {
	n := m.getSelectedNotification()
	if n == nil {
		return ""
	}

	// Try to get the HTML URL first
	if n.GetSubject().GetURL() != "" {
		return n.GetSubject().GetURL()
	}

	// Fall back to the API URL
	return n.GetURL()
}
