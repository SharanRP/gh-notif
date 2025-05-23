package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/common"
)

// ActionMode represents different action modes
type ActionMode int

const (
	// SelectMode is for selecting notifications
	SelectMode ActionMode = iota
	// ActionMenuMode is for choosing an action
	ActionMenuMode
	// ProgressMode is for showing progress
	ProgressMode
	// ResultMode is for showing results
	ResultMode
)

// ActionModel represents the action UI model
type ActionModel struct {
	// Data
	notifications []*github.Notification
	selected      map[string]bool
	selectedCount int

	// UI Components
	list        list.Model
	help        help.Model
	spinner     spinner.Model
	viewport    viewport.Model
	keyMap      actionKeyMap
	statusBar   StatusBar

	// State
	width       int
	height      int
	ready       bool
	loading     bool
	mode        ActionMode
	error       error
	result      *common.BatchResult

	// Context
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// actionKeyMap defines the keybindings for the action UI
type actionKeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Left           key.Binding
	Right          key.Binding
	Help           key.Binding
	Quit           key.Binding
	Select         key.Binding
	ToggleSelect   key.Binding
	SelectAll      key.Binding
	DeselectAll    key.Binding
	MarkAsRead     key.Binding
	Archive        key.Binding
	Subscribe      key.Binding
	Unsubscribe    key.Binding
	Mute           key.Binding
	Open           key.Binding
	Cancel         key.Binding
	Back           key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k actionKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Select, k.ToggleSelect}
}

// FullHelp returns keybindings for the expanded help view.
func (k actionKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Help, k.Quit, k.Select, k.ToggleSelect},
		{k.SelectAll, k.DeselectAll, k.MarkAsRead, k.Archive},
		{k.Subscribe, k.Unsubscribe, k.Mute, k.Open},
		{k.Cancel, k.Back},
	}
}

// defaultActionKeyMap returns the default key bindings
func defaultActionKeyMap() actionKeyMap {
	return actionKeyMap{
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
		ToggleSelect: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle select"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		DeselectAll: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "deselect all"),
		),
		MarkAsRead: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "mark as read"),
		),
		Archive: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "archive"),
		),
		Subscribe: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "subscribe"),
		),
		Unsubscribe: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "unsubscribe"),
		),
		Mute: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "mute repository"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "cancel"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// NewActionModel creates a new action model
func NewActionModel(notifications []*github.Notification) ActionModel {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Create a spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Create a help model
	h := help.New()
	h.ShowAll = false

	// Create a viewport
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.KeyMap{}

	// Create a list model
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()

	// Create the model
	m := ActionModel{
		notifications: notifications,
		selected:      make(map[string]bool),
		selectedCount: 0,
		list:          l,
		help:          h,
		spinner:       s,
		viewport:      vp,
		keyMap:        defaultActionKeyMap(),
		mode:          SelectMode,
		ctx:           ctx,
		cancelFunc:    cancel,
		statusBar: StatusBar{
			text:  fmt.Sprintf("%d notifications", len(notifications)),
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		},
	}

	// Initialize the list items
	items := make([]list.Item, len(notifications))
	for i, n := range notifications {
		items[i] = notificationItem{
			notification: n,
			selected:     false,
		}
	}
	m.list.SetItems(items)

	return m
}

// notificationItem represents a notification in the list
type notificationItem struct {
	notification *github.Notification
	selected     bool
}

// Title returns the title of the notification
func (i notificationItem) Title() string {
	return i.notification.GetSubject().GetTitle()
}

// Description returns the description of the notification
func (i notificationItem) Description() string {
	repo := i.notification.GetRepository().GetFullName()
	typ := i.notification.GetSubject().GetType()
	return fmt.Sprintf("%s (%s)", repo, typ)
}

// FilterValue returns the filter value of the notification
func (i notificationItem) FilterValue() string {
	return i.notification.GetSubject().GetTitle()
}

// Init initializes the model
func (m ActionModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.list.StartSpinner(),
	)
}

// Update updates the model
func (m ActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update the list
		headerHeight := 2
		footerHeight := 3
		m.list.SetSize(msg.Width, msg.Height-headerHeight-footerHeight)

		// Update the viewport
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerHeight - footerHeight

		return m, nil

	case tea.KeyMsg:
		// Handle global key bindings
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.cancelFunc()
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		}

		// Handle mode-specific key bindings
		switch m.mode {
		case SelectMode:
			return m.updateSelectMode(msg)
		case ActionMenuMode:
			return m.updateActionMenuMode(msg)
		case ProgressMode:
			return m.updateProgressMode(msg)
		case ResultMode:
			return m.updateResultMode(msg)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case ProgressUpdateMsg:
		// Update progress
		return m, nil

	case BatchResultMsg:
		// Update with batch result
		m.result = msg.Result
		m.mode = ResultMode
		return m, nil
	}

	// Update the list
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m ActionModel) View() string {
	if !m.ready {
		return "Initializing..."
	}

	switch m.mode {
	case SelectMode:
		return m.viewSelectMode()
	case ActionMenuMode:
		return m.viewActionMenuMode()
	case ProgressMode:
		return m.viewProgressMode()
	case ResultMode:
		return m.viewResultMode()
	default:
		return "Unknown mode"
	}
}

// BatchResultMsg is a message containing a batch result
type BatchResultMsg struct {
	Result *common.BatchResult
}

// DisplayActionUI displays the action UI
func DisplayActionUI(notifications []*github.Notification) error {
	if len(notifications) == 0 {
		fmt.Println("No notifications to act on.")
		return nil
	}

	// Create the model
	model := NewActionModel(notifications)

	// Run the UI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}
