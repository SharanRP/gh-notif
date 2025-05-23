package wizard

import (
	"fmt"
	"strings"

	"github.com/SharanRP/gh-notif/internal/config"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardOptions contains options for the setup wizard
type WizardOptions struct {
	// Interactive enables interactive mode
	Interactive bool

	// Theme sets the color theme
	Theme string

	// NoColor disables color output
	NoColor bool

	// Width sets the width of the wizard
	Width int

	// Height sets the height of the wizard
	Height int

	// ConfigPath is the path to the configuration file
	ConfigPath string

	// SkipAuth skips the authentication step
	SkipAuth bool

	// SkipDisplay skips the display settings step
	SkipDisplay bool

	// SkipNotifications skips the notification settings step
	SkipNotifications bool

	// SkipAdvanced skips the advanced settings step
	SkipAdvanced bool
}

// DefaultWizardOptions returns the default wizard options
func DefaultWizardOptions() *WizardOptions {
	return &WizardOptions{
		Interactive:       true,
		Theme:             "dark",
		NoColor:           false,
		Width:             80,
		Height:            20,
		ConfigPath:        "",
		SkipAuth:          false,
		SkipDisplay:       false,
		SkipNotifications: false,
		SkipAdvanced:      false,
	}
}

// Wizard represents the setup wizard
type Wizard struct {
	options *WizardOptions
	config  *config.Config
	steps   []WizardStep
	current int
}

// WizardStep represents a step in the wizard
type WizardStep struct {
	Title       string
	Description string
	Fields      []WizardField
	Action      func() error
}

// WizardField represents a field in a wizard step
type WizardField struct {
	Name        string
	Label       string
	Value       string
	Description string
	Required    bool
	Options     []string
	Validate    func(string) error
}

// NewWizard creates a new setup wizard
func NewWizard(options *WizardOptions) *Wizard {
	if options == nil {
		options = DefaultWizardOptions()
	}

	// Load or create config
	cfg := config.DefaultConfig()
	if options.ConfigPath != "" {
		// Load config from file
		configManager := config.NewConfigManager()
		if err := configManager.Load(); err == nil {
			cfg = configManager.GetConfig()
		}
	}

	w := &Wizard{
		options: options,
		config:  cfg,
		steps:   make([]WizardStep, 0),
		current: 0,
	}

	// Add wizard steps
	w.addWelcomeStep()
	if !options.SkipAuth {
		w.addAuthStep()
	}
	if !options.SkipDisplay {
		w.addDisplayStep()
	}
	if !options.SkipNotifications {
		w.addNotificationsStep()
	}
	if !options.SkipAdvanced {
		w.addAdvancedStep()
	}
	w.addSummaryStep()

	return w
}

// addWelcomeStep adds the welcome step
func (w *Wizard) addWelcomeStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Welcome to gh-notif",
		Description: `
This wizard will help you set up gh-notif with the optimal configuration for your needs.

You'll be guided through the following steps:
1. Authentication settings
2. Display preferences
3. Notification settings
4. Advanced options

Press 'n' to continue, or 'q' to quit.
`,
		Fields: []WizardField{},
		Action: nil,
	})
}

// addAuthStep adds the authentication step
func (w *Wizard) addAuthStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Authentication Settings",
		Description: `
gh-notif needs to authenticate with GitHub to access your notifications.

You can either use the OAuth device flow (recommended) or provide a personal access token.
`,
		Fields: []WizardField{
			{
				Name:        "auth.token_storage",
				Label:       "Token Storage Method",
				Value:       w.config.Auth.TokenStorage,
				Description: "How to store your authentication token",
				Required:    true,
				Options:     []string{"keyring", "file", "auto"},
				Validate: func(value string) error {
					if value != "keyring" && value != "file" && value != "auto" {
						return fmt.Errorf("invalid token storage method: must be 'keyring', 'file', or 'auto'")
					}
					return nil
				},
			},
			{
				Name:        "auth.scopes",
				Label:       "OAuth Scopes",
				Value:       strings.Join(w.config.Auth.Scopes, ","),
				Description: "Comma-separated list of OAuth scopes to request",
				Required:    true,
				Validate: func(value string) error {
					if value == "" {
						return fmt.Errorf("OAuth scopes cannot be empty")
					}
					return nil
				},
			},
		},
		Action: nil,
	})
}

// addDisplayStep adds the display settings step
func (w *Wizard) addDisplayStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Display Settings",
		Description: `
Configure how gh-notif displays information in the terminal.
`,
		Fields: []WizardField{
			{
				Name:        "display.theme",
				Label:       "Color Theme",
				Value:       w.config.Display.Theme,
				Description: "Color theme for the terminal UI",
				Required:    true,
				Options:     []string{"dark", "light", "auto"},
				Validate: func(value string) error {
					if value != "dark" && value != "light" && value != "auto" {
						return fmt.Errorf("invalid theme: must be 'dark', 'light', or 'auto'")
					}
					return nil
				},
			},
			{
				Name:        "display.date_format",
				Label:       "Date Format",
				Value:       w.config.Display.DateFormat,
				Description: "How to display dates",
				Required:    true,
				Options:     []string{"relative", "absolute", "iso"},
				Validate: func(value string) error {
					if value != "relative" && value != "absolute" && value != "iso" {
						return fmt.Errorf("invalid date format: must be 'relative', 'absolute', or 'iso'")
					}
					return nil
				},
			},
			{
				Name:        "display.output_format",
				Label:       "Default Output Format",
				Value:       w.config.Display.OutputFormat,
				Description: "Default format for command output",
				Required:    true,
				Options:     []string{"table", "json", "yaml", "text"},
				Validate: func(value string) error {
					if value != "table" && value != "json" && value != "yaml" && value != "text" {
						return fmt.Errorf("invalid output format: must be 'table', 'json', 'yaml', or 'text'")
					}
					return nil
				},
			},
			{
				Name:        "display.show_emojis",
				Label:       "Show Emojis",
				Value:       fmt.Sprintf("%t", w.config.Display.ShowEmojis),
				Description: "Whether to show emojis in the output",
				Required:    true,
				Options:     []string{"true", "false"},
				Validate: func(value string) error {
					if value != "true" && value != "false" {
						return fmt.Errorf("invalid value: must be 'true' or 'false'")
					}
					return nil
				},
			},
			{
				Name:        "display.compact_mode",
				Label:       "Compact Mode",
				Value:       fmt.Sprintf("%t", w.config.Display.CompactMode),
				Description: "Whether to use compact mode for output",
				Required:    true,
				Options:     []string{"true", "false"},
				Validate: func(value string) error {
					if value != "true" && value != "false" {
						return fmt.Errorf("invalid value: must be 'true' or 'false'")
					}
					return nil
				},
			},
		},
		Action: nil,
	})
}

// addNotificationsStep adds the notifications settings step
func (w *Wizard) addNotificationsStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Notification Settings",
		Description: `
Configure how gh-notif handles notifications.
`,
		Fields: []WizardField{
			{
				Name:        "notifications.default_filter",
				Label:       "Default Filter",
				Value:       w.config.Notifications.DefaultFilter,
				Description: "Default filter to apply when listing notifications",
				Required:    true,
				Options:     []string{"all", "unread", "participating"},
				Validate: func(value string) error {
					if value != "all" && value != "unread" && value != "participating" {
						return fmt.Errorf("invalid default filter: must be 'all', 'unread', or 'participating'")
					}
					return nil
				},
			},
			{
				Name:        "notifications.auto_refresh",
				Label:       "Auto Refresh",
				Value:       fmt.Sprintf("%t", w.config.Notifications.AutoRefresh),
				Description: "Whether to automatically refresh notifications",
				Required:    true,
				Options:     []string{"true", "false"},
				Validate: func(value string) error {
					if value != "true" && value != "false" {
						return fmt.Errorf("invalid value: must be 'true' or 'false'")
					}
					return nil
				},
			},
			{
				Name:        "notifications.refresh_interval",
				Label:       "Refresh Interval",
				Value:       fmt.Sprintf("%d", w.config.Notifications.RefreshInterval),
				Description: "Interval in seconds to refresh notifications",
				Required:    true,
				Validate: func(value string) error {
					// Validate that the value is a positive integer
					var interval int
					if _, err := fmt.Sscanf(value, "%d", &interval); err != nil {
						return fmt.Errorf("invalid refresh interval: must be a number")
					}
					if interval < 0 {
						return fmt.Errorf("invalid refresh interval: must be non-negative")
					}
					return nil
				},
			},
		},
		Action: nil,
	})
}

// addAdvancedStep adds the advanced settings step
func (w *Wizard) addAdvancedStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Advanced Settings",
		Description: `
Configure advanced settings for gh-notif.

These settings affect performance and behavior.
`,
		Fields: []WizardField{
			{
				Name:        "advanced.debug",
				Label:       "Debug Mode",
				Value:       fmt.Sprintf("%t", w.config.Advanced.Debug),
				Description: "Enable debug logging",
				Required:    true,
				Options:     []string{"true", "false"},
				Validate: func(value string) error {
					if value != "true" && value != "false" {
						return fmt.Errorf("invalid value: must be 'true' or 'false'")
					}
					return nil
				},
			},
			{
				Name:        "advanced.max_concurrent",
				Label:       "Max Concurrent Requests",
				Value:       fmt.Sprintf("%d", w.config.Advanced.MaxConcurrent),
				Description: "Maximum number of concurrent API requests",
				Required:    true,
				Validate: func(value string) error {
					var maxConcurrent int
					if _, err := fmt.Sscanf(value, "%d", &maxConcurrent); err != nil {
						return fmt.Errorf("invalid max concurrent: must be a number")
					}
					if maxConcurrent <= 0 {
						return fmt.Errorf("invalid max concurrent: must be positive")
					}
					return nil
				},
			},
			{
				Name:        "advanced.cache_ttl",
				Label:       "Cache TTL",
				Value:       fmt.Sprintf("%d", w.config.Advanced.CacheTTL),
				Description: "Time-to-live in seconds for cached data",
				Required:    true,
				Validate: func(value string) error {
					var cacheTTL int
					if _, err := fmt.Sscanf(value, "%d", &cacheTTL); err != nil {
						return fmt.Errorf("invalid cache TTL: must be a number")
					}
					if cacheTTL < 0 {
						return fmt.Errorf("invalid cache TTL: must be non-negative")
					}
					return nil
				},
			},
			{
				Name:        "advanced.cache_dir",
				Label:       "Cache Directory",
				Value:       w.config.Advanced.CacheDir,
				Description: "Directory to store cached data (leave empty for default)",
				Required:    false,
				Validate:    nil,
			},
		},
		Action: nil,
	})
}

// addSummaryStep adds the summary step
func (w *Wizard) addSummaryStep() {
	w.steps = append(w.steps, WizardStep{
		Title: "Configuration Summary",
		Description: `
Your configuration is ready to be saved.

Press 'Enter' to save the configuration and complete the setup.
`,
		Fields: []WizardField{},
		Action: func() error {
			// Save configuration
			configManager := config.NewConfigManager()
			if w.options.ConfigPath != "" {
				// TODO: Set config path
			}
			// TODO: Update config with wizard values
			return configManager.Save()
		},
	})
}

// Run runs the wizard
func (w *Wizard) Run() error {
	if w.options.Interactive {
		return w.runInteractive()
	}
	return w.runNonInteractive()
}

// runNonInteractive runs the wizard in non-interactive mode
func (w *Wizard) runNonInteractive() error {
	// TODO: Implement non-interactive mode
	return fmt.Errorf("non-interactive mode not implemented")
}

// runInteractive runs the wizard in interactive mode
func (w *Wizard) runInteractive() error {
	p := tea.NewProgram(newWizardModel(w), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// wizardModel is the Bubble Tea model for the wizard
type wizardModel struct {
	wizard   *Wizard
	viewport viewport.Model
	help     help.Model
	inputs   []textinput.Model
	keys     wizardKeyMap
	quitting bool
	err      error
}

// wizardKeyMap defines the keybindings for the wizard
type wizardKeyMap struct {
	Next  key.Binding
	Prev  key.Binding
	Quit  key.Binding
	Help  key.Binding
	Enter key.Binding
	Tab   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k wizardKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view.
func (k wizardKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Quit},
		{k.Help, k.Enter, k.Tab},
	}
}

// newWizardKeyMap creates a new key map for the wizard
func newWizardKeyMap() wizardKeyMap {
	return wizardKeyMap{
		Next: key.NewBinding(
			key.WithKeys("n", "right"),
			key.WithHelp("n/→", "next"),
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
			key.WithHelp("enter", "confirm"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
	}
}

// newWizardModel creates a new model for the wizard
func newWizardModel(w *Wizard) wizardModel {
	vp := viewport.New(w.options.Width, w.options.Height)
	vp.SetContent(formatWizardStep(w.steps[0]))

	helpModel := help.New()
	helpModel.Width = w.options.Width

	inputs := make([]textinput.Model, 0)
	for _, field := range w.steps[0].Fields {
		input := textinput.New()
		input.Placeholder = field.Label
		input.SetValue(field.Value)
		input.Focus()
		inputs = append(inputs, input)
	}

	return wizardModel{
		wizard:   w,
		viewport: vp,
		help:     helpModel,
		inputs:   inputs,
		keys:     newWizardKeyMap(),
		quitting: false,
		err:      nil,
	}
}

// formatWizardStep formats a wizard step for display
func formatWizardStep(step WizardStep) string {
	var sb strings.Builder

	// Title
	sb.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("# " + step.Title))
	sb.WriteString("\n\n")

	// Description
	sb.WriteString(strings.TrimSpace(step.Description))
	sb.WriteString("\n\n")

	return sb.String()
}

// Init initializes the model
func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update updates the model
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.wizard.current < len(m.wizard.steps)-1 {
				m.wizard.current++
				m.viewport.SetContent(formatWizardStep(m.wizard.steps[m.wizard.current]))
				m.viewport.GotoTop()

				// Update inputs for the new step
				m.inputs = make([]textinput.Model, 0)
				for _, field := range m.wizard.steps[m.wizard.current].Fields {
					input := textinput.New()
					input.Placeholder = field.Label
					input.SetValue(field.Value)
					input.Focus()
					m.inputs = append(m.inputs, input)
				}
			}
		case key.Matches(msg, m.keys.Prev):
			if m.wizard.current > 0 {
				m.wizard.current--
				m.viewport.SetContent(formatWizardStep(m.wizard.steps[m.wizard.current]))
				m.viewport.GotoTop()

				// Update inputs for the new step
				m.inputs = make([]textinput.Model, 0)
				for _, field := range m.wizard.steps[m.wizard.current].Fields {
					input := textinput.New()
					input.Placeholder = field.Label
					input.SetValue(field.Value)
					input.Focus()
					m.inputs = append(m.inputs, input)
				}
			}
		case key.Matches(msg, m.keys.Enter):
			step := m.wizard.steps[m.wizard.current]
			if step.Action != nil {
				if err := step.Action(); err != nil {
					m.err = err
				}
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

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Update inputs
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m wizardModel) View() string {
	if m.quitting {
		return ""
	}

	step := m.wizard.steps[m.wizard.current]
	progress := fmt.Sprintf(" %d/%d ", m.wizard.current+1, len(m.wizard.steps))

	// Render inputs
	var inputsView string
	if len(m.inputs) > 0 {
		var sb strings.Builder
		for i, input := range m.inputs {
			field := step.Fields[i]
			sb.WriteString(fmt.Sprintf("%s: %s\n", field.Label, input.View()))
			if field.Description != "" {
				sb.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("240")).
					Render("  " + field.Description))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
		inputsView = sb.String()
	}

	// Render error
	var errorView string
	if m.err != nil {
		errorView = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("Error: %v", m.err))
		errorView += "\n\n"
	}

	helpView := m.help.View(m.keys)

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		m.viewport.View(),
		inputsView,
		errorView,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(progress),
		helpView)
}

// RunWizard runs the wizard with the given options
func RunWizard(options *WizardOptions) error {
	wizard := NewWizard(options)
	return wizard.Run()
}
