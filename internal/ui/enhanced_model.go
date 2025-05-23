package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
	"github.com/user/gh-notif/internal/ui/components"
)

// EnhancedModel represents the enhanced UI model with modern components
type EnhancedModel struct {
	// Data
	notifications []*github.Notification

	// Components
	registry      *components.ComponentRegistry
	layout        *components.Layout
	virtualList   *components.VirtualList
	statusPanel   *components.Panel
	filterForm    *components.Form
	progressBar   *components.Progress
	markdown      *components.MarkdownRenderer

	// State
	width         int
	height        int
	ready         bool
	loading       bool
	viewMode      EnhancedViewMode
	showFilter    bool
	showHelp      bool
	animFrame     int
	lastUpdate    time.Time

	// Styling
	theme         EnhancedTheme
	styles        EnhancedStyles
	symbols       Symbols

	// Key bindings
	keyMap        EnhancedKeyMap

	// Context
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

// EnhancedViewMode represents different enhanced view modes
type EnhancedViewMode int

const (
	// EnhancedListView shows a virtualized list
	EnhancedListView EnhancedViewMode = iota
	// EnhancedDetailView shows detailed notification view
	EnhancedDetailView
	// EnhancedSplitView shows split layout
	EnhancedSplitView
	// EnhancedDashboardView shows dashboard with multiple panels
	EnhancedDashboardView
)

// EnhancedKeyMap defines enhanced key bindings
type EnhancedKeyMap struct {
	// Navigation
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	PageUp        key.Binding
	PageDown      key.Binding
	Home          key.Binding
	End           key.Binding

	// Actions
	Select        key.Binding
	MarkAsRead    key.Binding
	MarkAllAsRead key.Binding
	Archive       key.Binding
	Open          key.Binding

	// UI Controls
	ToggleView    key.Binding
	ToggleTheme   key.Binding
	ToggleCompact key.Binding
	ToggleFilter  key.Binding
	ToggleHelp    key.Binding
	Refresh       key.Binding

	// System
	Quit          key.Binding
	Cancel        key.Binding
}

// DefaultEnhancedKeyMap returns default enhanced key bindings
func DefaultEnhancedKeyMap() EnhancedKeyMap {
	return EnhancedKeyMap{
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
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
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
		Archive: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "toggle view"),
		),
		ToggleTheme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle theme"),
		),
		ToggleCompact: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "toggle compact"),
		),
		ToggleFilter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "toggle filter"),
		),
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// NewEnhancedModel creates a new enhanced UI model
func NewEnhancedModel(notifications []*github.Notification) *EnhancedModel {
	ctx, cancel := context.WithCancel(context.Background())

	// Create theme and styles
	theme := NewEnhancedDarkTheme()
	theme = AdaptThemeToTerminal(theme)
	styles := NewEnhancedStyles(theme)
	symbols := DefaultSymbols()

	// Create component registry
	registry := components.GetGlobalRegistry()

	// Create main layout
	layout := components.NewLayout(components.LayoutVertical)

	// Create virtual list with notification items
	componentStyles := components.DefaultEnhancedStyles()
	componentSymbols := components.DefaultSymbols()
	notificationItems := components.NewNotificationItemList(notifications, componentStyles, componentSymbols)
	virtualList := components.NewVirtualList(notificationItems.GetVirtualListItems(), 1)

	// Create status panel
	statusPanel := components.NewPanel("Status", components.PanelDefault)
	statusPanel.SetContent(fmt.Sprintf("%d notifications", len(notifications)))

	// Create filter form
	filterForm := components.NewForm("Filter Notifications")
	filterForm.AddField(
		components.NewTextInputField("query", "Search").
			SetPlaceholder("Type to filter notifications...").
			SetHelp("Use keywords to filter notifications"),
	)
	filterForm.AddField(
		components.NewTextInputField("repo", "Repository").
			SetPlaceholder("owner/repo").
			SetHelp("Filter by repository name"),
	)

	// Create progress bar
	progressBar := components.NewProgress(components.ProgressBar)
	progressBar.SetTitle("Loading notifications...")

	// Create markdown renderer for help
	helpContent := `# GitHub Notifications Help

## Navigation
- **↑/↓ or j/k**: Navigate up/down
- **Page Up/Down**: Page navigation
- **Home/End or g/G**: Go to top/bottom

## Actions
- **Enter/Space**: Select notification
- **m**: Mark as read
- **M**: Mark all as read
- **a**: Archive notification
- **o**: Open in browser

## Views
- **v**: Toggle view mode
- **c**: Toggle compact mode
- **t**: Toggle theme
- **/**: Toggle filter
- **?**: Toggle this help

## Other
- **r**: Refresh notifications
- **q**: Quit application
- **Esc**: Cancel current action`

	markdown := components.NewMarkdownRenderer(helpContent)

	model := &EnhancedModel{
		notifications: notifications,
		registry:      registry,
		layout:        layout,
		virtualList:   virtualList,
		statusPanel:   statusPanel,
		filterForm:    filterForm,
		progressBar:   progressBar,
		markdown:      markdown,
		viewMode:      EnhancedListView,
		theme:         theme,
		styles:        styles,
		symbols:       symbols,
		keyMap:        DefaultEnhancedKeyMap(),
		ctx:           ctx,
		cancelFunc:    cancel,
		lastUpdate:    time.Now(),
	}

	// Setup layout
	model.setupLayout()

	return model
}

// setupLayout configures the main layout
func (m *EnhancedModel) setupLayout() {
	switch m.viewMode {
	case EnhancedListView:
		m.layout = components.NewLayout(components.LayoutVertical)
		m.layout.AddComponent("list", m.virtualList)
		m.layout.AddComponent("status", m.statusPanel)

	case EnhancedDetailView:
		m.layout = components.NewLayout(components.LayoutVertical)
		m.layout.AddComponent("detail", m.markdown)
		m.layout.AddComponent("status", m.statusPanel)

	case EnhancedSplitView:
		m.layout = components.NewLayout(components.LayoutSplit)
		m.layout.SetSplits([]int{1, 2}) // 1:2 ratio
		m.layout.AddComponent("list", m.virtualList)
		m.layout.AddComponent("detail", m.markdown)

	case EnhancedDashboardView:
		m.layout = components.NewLayout(components.LayoutGrid)
		m.layout.AddComponent("list", m.virtualList)
		m.layout.AddComponent("status", m.statusPanel)
		m.layout.AddComponent("progress", m.progressBar)
	}

	if m.showFilter {
		filterLayout := components.NewLayout(components.LayoutVertical)
		filterLayout.AddComponent("filter", m.filterForm)
		filterLayout.AddComponent("main", m.layout)
		m.layout = filterLayout
	}

	if m.showHelp {
		helpLayout := components.NewLayout(components.LayoutSplit)
		helpLayout.SetSplits([]int{2, 1}) // 2:1 ratio
		helpLayout.AddComponent("main", m.layout)
		helpLayout.AddComponent("help", m.markdown)
		m.layout = helpLayout
	}
}

// Init initializes the enhanced model
func (m *EnhancedModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize all components
	if cmd := m.layout.Init(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Start animation ticker
	cmds = append(cmds, m.tickAnimation())

	m.ready = true
	return tea.Batch(cmds...)
}

// Update handles messages and updates the enhanced model
func (m *EnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(m.width, m.height)

		// Send resize message to all components
		resizeMsg := components.ComponentMessage{
			Type: components.ComponentResizeMsg,
			Data: struct{ Width, Height int }{m.width, m.height},
		}

		var cmd tea.Cmd
		updatedLayout, cmd := m.layout.Update(resizeMsg)
		m.layout = updatedLayout.(*components.Layout)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Quit):
			m.cancelFunc()
			return m, tea.Quit

		case key.Matches(msg, m.keyMap.ToggleView):
			m.toggleViewMode()
			m.setupLayout()
			m.layout.SetSize(m.width, m.height)

		case key.Matches(msg, m.keyMap.ToggleTheme):
			m.toggleTheme()

		case key.Matches(msg, m.keyMap.ToggleFilter):
			m.showFilter = !m.showFilter
			m.setupLayout()
			m.layout.SetSize(m.width, m.height)

		case key.Matches(msg, m.keyMap.ToggleHelp):
			m.showHelp = !m.showHelp
			m.setupLayout()
			m.layout.SetSize(m.width, m.height)

		case key.Matches(msg, m.keyMap.Refresh):
			m.loading = true
			m.progressBar.SetValue(0.0)
			return m, m.refreshNotifications()

		default:
			// Pass message to layout
			var cmd tea.Cmd
			updatedLayout, cmd := m.layout.Update(msg)
			m.layout = updatedLayout.(*components.Layout)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case AnimationTickMsg:
		m.animFrame++
		m.lastUpdate = time.Now()
		cmds = append(cmds, m.tickAnimation())

	case components.ComponentEvent:
		switch msg.EventType {
		case "select":
			if item, ok := msg.Data.(*components.NotificationItem); ok {
				m.showNotificationDetail(item)
			}
		case "submit":
			if m.showFilter {
				m.applyFilter(msg.Data.(map[string]interface{}))
			}
		}

	default:
		// Pass message to layout
		var cmd tea.Cmd
		updatedLayout, cmd := m.layout.Update(msg)
		m.layout = updatedLayout.(*components.Layout)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the enhanced model
func (m *EnhancedModel) View() string {
	if !m.ready {
		return m.styles.Spinner.Render("Initializing enhanced UI...")
	}

	// Create animated header
	headerText := "GitHub Notifications"
	if m.loading {
		headerText = CreatePulseText(headerText, m.theme.PulseColors, m.animFrame)
	} else {
		headerText = CreateGradientText(headerText, m.theme.PrimaryGradient)
	}

	header := m.styles.HeaderGradient.Render(headerText)

	// Add view mode indicator
	viewModeText := m.getViewModeText()
	viewModeBadge := m.styles.BadgeSecondary.Render(viewModeText)

	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, header, "  ", viewModeBadge)

	// Render main content
	content := m.layout.View()

	// Create footer with key hints
	var footerParts []string
	if !m.showHelp {
		footerParts = append(footerParts,
			m.styles.BadgeInfo.Render("? Help"),
			m.styles.BadgeSecondary.Render("v View"),
			m.styles.BadgeSecondary.Render("/ Filter"),
			m.styles.BadgeSecondary.Render("r Refresh"),
			m.styles.BadgeError.Render("q Quit"),
		)
	}

	footer := lipgloss.JoinHorizontal(lipgloss.Center, footerParts...)

	// Join all parts
	view := lipgloss.JoinVertical(lipgloss.Left,
		headerLine,
		"",
		content,
		"",
		footer,
	)

	// Apply container styling
	containerStyle := m.styles.Container.
		Width(m.width).
		Height(m.height)

	return containerStyle.Render(view)
}

// toggleViewMode cycles through view modes
func (m *EnhancedModel) toggleViewMode() {
	switch m.viewMode {
	case EnhancedListView:
		m.viewMode = EnhancedDetailView
	case EnhancedDetailView:
		m.viewMode = EnhancedSplitView
	case EnhancedSplitView:
		m.viewMode = EnhancedDashboardView
	case EnhancedDashboardView:
		m.viewMode = EnhancedListView
	}
}

// toggleTheme switches between light and dark themes
func (m *EnhancedModel) toggleTheme() {
	if m.theme.Background == lipgloss.Color("#1E1E2E") {
		m.theme = NewEnhancedLightTheme()
	} else {
		m.theme = NewEnhancedDarkTheme()
	}

	m.theme = AdaptThemeToTerminal(m.theme)
	m.styles = NewEnhancedStyles(m.theme)

	// Update component styles
	componentStyles := components.ComponentStyles{
		Base:     lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")),
		Focused:  lipgloss.NewStyle().Background(lipgloss.Color("#45475A")).Foreground(lipgloss.Color("#F5E0DC")),
		Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")),
		Error:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")),
		Success:  m.styles.BadgeSuccess,
		Warning:  m.styles.BadgeWarning,
		Info:     m.styles.BadgeInfo,
	}

	m.virtualList.SetStyles(componentStyles)
	m.statusPanel.SetStyles(componentStyles)
	m.filterForm.SetStyles(componentStyles)
	m.progressBar.SetStyles(componentStyles)
	m.markdown.SetStyles(componentStyles)
}

// getViewModeText returns text for the current view mode
func (m *EnhancedModel) getViewModeText() string {
	switch m.viewMode {
	case EnhancedListView:
		return "LIST"
	case EnhancedDetailView:
		return "DETAIL"
	case EnhancedSplitView:
		return "SPLIT"
	case EnhancedDashboardView:
		return "DASHBOARD"
	default:
		return "UNKNOWN"
	}
}

// showNotificationDetail shows details for a notification
func (m *EnhancedModel) showNotificationDetail(item *components.NotificationItem) {
	notification := item.GetNotification()

	// Create markdown content for the notification
	content := fmt.Sprintf(`# %s

**Repository:** %s
**Type:** %s
**Reason:** %s
**Updated:** %s

## Description
%s

## Links
- [View on GitHub](%s)
`,
		notification.GetSubject().GetTitle(),
		notification.GetRepository().GetFullName(),
		notification.GetSubject().GetType(),
		notification.GetReason(),
		notification.GetUpdatedAt().Format("2006-01-02 15:04:05"),
		notification.GetSubject().GetTitle(), // Using title as description for now
		notification.GetSubject().GetURL(),
	)

	m.markdown.SetContent(content)
}

// applyFilter applies filter criteria to the notification list
func (m *EnhancedModel) applyFilter(values map[string]interface{}) {
	// Implementation would filter the notification list based on form values
	// For now, just close the filter
	m.showFilter = false
	m.setupLayout()
	m.layout.SetSize(m.width, m.height)
}

// refreshNotifications simulates refreshing notifications
func (m *EnhancedModel) refreshNotifications() tea.Cmd {
	return func() tea.Msg {
		// Simulate loading time
		time.Sleep(1 * time.Second)
		return RefreshCompleteMsg{}
	}
}

// tickAnimation returns a command for animation ticking
func (m *EnhancedModel) tickAnimation() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return AnimationTickMsg{Time: t}
	})
}

// Animation messages
type AnimationTickMsg struct {
	Time time.Time
}

type RefreshCompleteMsg struct{}

// DisplayEnhancedNotifications displays notifications using the enhanced UI
func DisplayEnhancedNotifications(notifications []*github.Notification) error {
	model := NewEnhancedModel(notifications)

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()

	return err
}
