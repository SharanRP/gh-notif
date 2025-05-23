package tutorial

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// TutorialOptions contains options for the tutorial
type TutorialOptions struct {
	// Interactive enables interactive mode
	Interactive bool

	// SkipAuth skips the authentication section
	SkipAuth bool

	// SkipBasics skips the basics section
	SkipBasics bool

	// SkipAdvanced skips the advanced section
	SkipAdvanced bool

	// Theme sets the color theme
	Theme string

	// NoColor disables color output
	NoColor bool

	// Width sets the width of the tutorial
	Width int

	// Height sets the height of the tutorial
	Height int
}

// DefaultTutorialOptions returns the default tutorial options
func DefaultTutorialOptions() *TutorialOptions {
	return &TutorialOptions{
		Interactive: true,
		SkipAuth:    false,
		SkipBasics:  false,
		SkipAdvanced: false,
		Theme:       "dark",
		NoColor:     false,
		Width:       80,
		Height:      20,
	}
}

// Tutorial represents the tutorial
type Tutorial struct {
	options *TutorialOptions
	steps   []TutorialStep
	current int
}

// TutorialStep represents a step in the tutorial
type TutorialStep struct {
	Title       string
	Description string
	Example     string
	Action      func() error
}

// NewTutorial creates a new tutorial
func NewTutorial(options *TutorialOptions) *Tutorial {
	if options == nil {
		options = DefaultTutorialOptions()
	}

	t := &Tutorial{
		options: options,
		steps:   make([]TutorialStep, 0),
		current: 0,
	}

	// Add tutorial steps
	t.addBasicSteps()
	t.addAdvancedSteps()

	return t
}

// addBasicSteps adds basic tutorial steps
func (t *Tutorial) addBasicSteps() {
	if t.options.SkipBasics {
		return
	}

	// Introduction
	t.steps = append(t.steps, TutorialStep{
		Title: "Welcome to gh-notif",
		Description: `
gh-notif is a high-performance CLI tool for managing GitHub notifications in the terminal.

This tutorial will guide you through the basic features of gh-notif and help you get started.

Press 'n' to go to the next step, 'p' to go to the previous step, or 'q' to quit the tutorial.
`,
		Example: "",
		Action:  nil,
	})

	// Authentication
	if !t.options.SkipAuth {
		t.steps = append(t.steps, TutorialStep{
			Title: "Authentication",
			Description: `
Before using gh-notif, you need to authenticate with GitHub.

To authenticate, run:
`,
			Example: "gh-notif auth login",
			Action:  nil,
		})
	}

	// Listing notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Listing Notifications",
		Description: `
The most basic command is 'list', which shows your unread notifications.

You can filter notifications by repository, type, and more.
`,
		Example: `# List all unread notifications
gh-notif list

# List notifications for a specific repository
gh-notif list --repo="owner/repo"

# List notifications of a specific type
gh-notif list --type="PullRequest"`,
		Action: nil,
	})

	// Filtering notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Filtering Notifications",
		Description: `
gh-notif provides powerful filtering capabilities.

You can use simple flags or complex filter expressions.
`,
		Example: `# Use a simple filter
gh-notif list --repo="owner/repo" --type="PullRequest"

# Use a complex filter expression
gh-notif list --filter="repo:owner/repo AND type:PullRequest AND is:unread"`,
		Action: nil,
	})

	// Saving filters
	t.steps = append(t.steps, TutorialStep{
		Title: "Saving Filters",
		Description: `
You can save filters for later use.

Saved filters can be used with the @name syntax.
`,
		Example: `# Save a filter
gh-notif filter save my-prs "repo:owner/repo type:PullRequest is:unread"

# Use a saved filter
gh-notif list --filter="@my-prs"`,
		Action: nil,
	})

	// Marking as read
	t.steps = append(t.steps, TutorialStep{
		Title: "Marking Notifications as Read",
		Description: `
You can mark notifications as read using the 'read' command.

You can mark individual notifications or all notifications matching a filter.
`,
		Example: `# Mark a notification as read
gh-notif read <notification-id>

# Mark all notifications as read
gh-notif mark-read

# Mark notifications matching a filter as read
gh-notif mark-read --filter="repo:owner/repo"`,
		Action: nil,
	})

	// Opening notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Opening Notifications",
		Description: `
You can open notifications in your browser using the 'open' command.
`,
		Example: `# Open a notification in the browser
gh-notif open <notification-id>`,
		Action: nil,
	})
}

// addAdvancedSteps adds advanced tutorial steps
func (t *Tutorial) addAdvancedSteps() {
	if t.options.SkipAdvanced {
		return
	}

	// Grouping notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Grouping Notifications",
		Description: `
You can group notifications by various criteria using the 'group' command.

This is useful for organizing large numbers of notifications.
`,
		Example: `# Group by repository
gh-notif group --by repository

# Group by type
gh-notif group --by type

# Group with smart grouping
gh-notif group --by smart`,
		Action: nil,
	})

	// Searching notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Searching Notifications",
		Description: `
You can search notifications using the 'search' command.

This performs a full-text search across all notification content.
`,
		Example: `# Search for text
gh-notif search "bug fix"

# Search with regex
gh-notif search "bug.*fix" --regex

# Search in interactive mode
gh-notif search --interactive`,
		Action: nil,
	})

	// Watching notifications
	t.steps = append(t.steps, TutorialStep{
		Title: "Watching Notifications",
		Description: `
You can watch for new notifications using the 'watch' command.

This will show real-time updates as new notifications arrive.
`,
		Example: `# Watch all notifications
gh-notif watch

# Watch with a filter
gh-notif watch --filter="repo:owner/repo"

# Watch with desktop notifications
gh-notif watch --desktop-notification`,
		Action: nil,
	})

	// Terminal UI
	t.steps = append(t.steps, TutorialStep{
		Title: "Terminal UI",
		Description: `
gh-notif provides an interactive terminal UI for managing notifications.

This is the most powerful way to interact with gh-notif.
`,
		Example: `# Start the terminal UI
gh-notif ui

# Start with a filter
gh-notif ui --filter="repo:owner/repo"`,
		Action: nil,
	})

	// Configuration
	t.steps = append(t.steps, TutorialStep{
		Title: "Configuration",
		Description: `
gh-notif is highly configurable.

You can manage configuration using the 'config' command.
`,
		Example: `# List all configuration values
gh-notif config list

# Get a configuration value
gh-notif config get display.theme

# Set a configuration value
gh-notif config set display.theme dark

# Edit the configuration file
gh-notif config edit`,
		Action: nil,
	})

	// Conclusion
	t.steps = append(t.steps, TutorialStep{
		Title: "Conclusion",
		Description: `
Congratulations! You've completed the gh-notif tutorial.

You now know the basics of using gh-notif to manage your GitHub notifications.

For more information, check out the documentation:
https://github.com/SharanRP/gh-notif

Or run 'gh-notif help' to see all available commands.
`,
		Example: "",
		Action:  nil,
	})
}

// Run runs the tutorial
func (t *Tutorial) Run() error {
	if t.options.Interactive {
		return t.runInteractive()
	}
	return t.runNonInteractive()
}

// runNonInteractive runs the tutorial in non-interactive mode
func (t *Tutorial) runNonInteractive() error {
	for _, step := range t.steps {
		fmt.Printf("# %s\n\n", step.Title)
		fmt.Println(strings.TrimSpace(step.Description))
		if step.Example != "" {
			fmt.Printf("\n```\n%s\n```\n", step.Example)
		}
		fmt.Println()
	}
	return nil
}

// runInteractive runs the tutorial in interactive mode
func (t *Tutorial) runInteractive() error {
	p := tea.NewProgram(newTutorialModel(t), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// tutorialModel is the Bubble Tea model for the tutorial
type tutorialModel struct {
	tutorial *Tutorial
	viewport viewport.Model
	help     help.Model
	keys     tutorialKeyMap
	quitting bool
}

// tutorialKeyMap defines the keybindings for the tutorial
type tutorialKeyMap struct {
	Next  key.Binding
	Prev  key.Binding
	Quit  key.Binding
	Help  key.Binding
	Enter key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k tutorialKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view.
func (k tutorialKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Quit},
		{k.Help, k.Enter},
	}
}

// newTutorialKeyMap creates a new key map for the tutorial
func newTutorialKeyMap() tutorialKeyMap {
	return tutorialKeyMap{
		Next: key.NewBinding(
			key.WithKeys("n", "right", "space"),
			key.WithHelp("n/→/space", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("p", "left"),
			key.WithHelp("p/←", "previous"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q/esc", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "run example"),
		),
	}
}

// newTutorialModel creates a new model for the tutorial
func newTutorialModel(t *Tutorial) tutorialModel {
	vp := viewport.New(t.options.Width, t.options.Height)
	vp.SetContent(formatStep(t.steps[0]))

	helpModel := help.New()
	helpModel.Width = t.options.Width

	return tutorialModel{
		tutorial: t,
		viewport: vp,
		help:     helpModel,
		keys:     newTutorialKeyMap(),
		quitting: false,
	}
}

// formatStep formats a tutorial step for display
func formatStep(step TutorialStep) string {
	var sb strings.Builder

	// Title
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("# " + step.Title))
	sb.WriteString("\n\n")

	// Description
	sb.WriteString(wordwrap.String(strings.TrimSpace(step.Description), 80))
	sb.WriteString("\n\n")

	// Example
	if step.Example != "" {
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Render("Example:"))
		sb.WriteString("\n\n")
		sb.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")).
			Background(lipgloss.Color("236")).
			Padding(1).
			Width(78).
			Render(step.Example))
		sb.WriteString("\n\n")
	}

	return sb.String()
}

// Init initializes the model
func (m tutorialModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m tutorialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Next):
			if m.tutorial.current < len(m.tutorial.steps)-1 {
				m.tutorial.current++
				m.viewport.SetContent(formatStep(m.tutorial.steps[m.tutorial.current]))
				m.viewport.GotoTop()
			}
		case key.Matches(msg, m.keys.Prev):
			if m.tutorial.current > 0 {
				m.tutorial.current--
				m.viewport.SetContent(formatStep(m.tutorial.steps[m.tutorial.current]))
				m.viewport.GotoTop()
			}
		case key.Matches(msg, m.keys.Enter):
			step := m.tutorial.steps[m.tutorial.current]
			if step.Action != nil {
				step.Action()
			}
		}
	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 3
		verticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMarginHeight
		m.help.Width = msg.Width
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m tutorialModel) View() string {
	if m.quitting {
		return ""
	}

	progress := fmt.Sprintf(" %d/%d ", m.tutorial.current+1, len(m.tutorial.steps))

	helpView := m.help.View(m.keys)

	return fmt.Sprintf("%s\n%s\n%s",
		m.viewport.View(),
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(progress),
		helpView)
}

// RunTutorial runs the tutorial with the given options
func RunTutorial(options *TutorialOptions) error {
	tutorial := NewTutorial(options)
	return tutorial.Run()
}
