package examples

import (
	"fmt"
	"time"

	"github.com/SharanRP/gh-notif/internal/ui"
	"github.com/SharanRP/gh-notif/internal/ui/components"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
)

// DemoModel demonstrates the enhanced UI components
type DemoModel struct {
	// Components
	registry    *components.ComponentRegistry
	layout      *components.Layout
	virtualList *components.VirtualList
	form        *components.Form
	progress    *components.Progress
	markdown    *components.MarkdownRenderer

	// State
	width       int
	height      int
	currentDemo int
	demos       []Demo

	// Styling
	theme  ui.EnhancedTheme
	styles ui.EnhancedStyles
}

// Demo represents a UI component demonstration
type Demo struct {
	Name        string
	Description string
	Component   components.Component
	SetupFunc   func(*DemoModel)
}

// NewDemoModel creates a new demo model
func NewDemoModel() *DemoModel {
	// Create theme and styles
	theme := ui.NewEnhancedDarkTheme()
	theme = ui.AdaptThemeToTerminal(theme)
	styles := ui.NewEnhancedStyles(theme)

	// Create component registry
	registry := components.GetGlobalRegistry()

	model := &DemoModel{
		registry: registry,
		theme:    theme,
		styles:   styles,
	}

	// Setup demos
	model.setupDemos()

	return model
}

// setupDemos initializes all the demo components
func (m *DemoModel) setupDemos() {
	m.demos = []Demo{
		{
			Name:        "Virtual List",
			Description: "High-performance virtualized list with thousands of items",
			SetupFunc:   (*DemoModel).setupVirtualListDemo,
		},
		{
			Name:        "Interactive Form",
			Description: "Form with validation and keyboard navigation",
			SetupFunc:   (*DemoModel).setupFormDemo,
		},
		{
			Name:        "Progress Indicators",
			Description: "Various progress indicators with animations",
			SetupFunc:   (*DemoModel).setupProgressDemo,
		},
		{
			Name:        "Markdown Rendering",
			Description: "Rich markdown content with syntax highlighting",
			SetupFunc:   (*DemoModel).setupMarkdownDemo,
		},
		{
			Name:        "Layout System",
			Description: "Flexible layout system with multiple arrangements",
			SetupFunc:   (*DemoModel).setupLayoutDemo,
		},
		{
			Name:        "Enhanced Styling",
			Description: "Gradients, animations, and advanced styling",
			SetupFunc:   (*DemoModel).setupStylingDemo,
		},
	}

	// Setup the first demo
	if len(m.demos) > 0 {
		m.demos[m.currentDemo].SetupFunc(m)
	}
}

// setupVirtualListDemo sets up the virtual list demonstration
func (m *DemoModel) setupVirtualListDemo() {
	// Create sample notifications
	notifications := m.createSampleNotifications(1000)

	// Create notification items
	componentStyles := components.DefaultEnhancedStyles()
	symbols := components.DefaultSymbols()
	notificationItems := components.NewNotificationItemList(notifications, componentStyles, symbols)

	// Create virtual list
	m.virtualList = components.NewVirtualList(notificationItems.GetVirtualListItems(), 1)
	m.virtualList.SetSize(m.width-4, m.height-8)
	m.virtualList.SetFocused(true)

	// Create layout
	m.layout = components.NewLayout(components.LayoutVertical)
	m.layout.AddComponent("title", m.createTitlePanel("Virtual List Demo", "Navigate through 1000+ notifications with smooth scrolling"))
	m.layout.AddComponent("list", m.virtualList)
	m.layout.AddComponent("help", m.createHelpPanel("↑/↓: Navigate • Enter: Select • q: Quit"))
}

// setupFormDemo sets up the form demonstration
func (m *DemoModel) setupFormDemo() {
	// Create form
	m.form = components.NewForm("GitHub Notification Filter")

	// Add fields
	m.form.AddField(
		components.NewTextInputField("query", "Search Query").
			SetPlaceholder("Enter search terms...").
			SetHelp("Search in notification titles and descriptions").
			SetRequired(false),
	)

	m.form.AddField(
		components.NewTextInputField("repo", "Repository").
			SetPlaceholder("owner/repository").
			SetHelp("Filter by specific repository").
			SetValidator(func(value string) error {
				if value != "" && len(value) < 3 {
					return fmt.Errorf("repository name must be at least 3 characters")
				}
				return nil
			}),
	)

	m.form.AddField(
		components.NewTextInputField("author", "Author").
			SetPlaceholder("username").
			SetHelp("Filter by notification author").
			SetRequired(false),
	)

	m.form.SetSize(m.width-4, m.height-8)
	m.form.SetFocused(true)

	// Create layout
	m.layout = components.NewLayout(components.LayoutVertical)
	m.layout.AddComponent("title", m.createTitlePanel("Interactive Form Demo", "Tab between fields, validate input, submit with Ctrl+S"))
	m.layout.AddComponent("form", m.form)
	m.layout.AddComponent("help", m.createHelpPanel("Tab: Next field • Shift+Tab: Previous • Ctrl+S: Submit • Esc: Cancel"))
}

// setupProgressDemo sets up the progress demonstration
func (m *DemoModel) setupProgressDemo() {
	// Create different types of progress indicators
	progressBar := components.NewProgress(components.ProgressBar)
	progressBar.SetTitle("Downloading notifications...")
	progressBar.SetDescription("Fetching from GitHub API")
	progressBar.SetValue(0.65)

	progressSpinner := components.NewProgress(components.ProgressSpinner)
	progressSpinner.SetTitle("Processing notifications")
	progressSpinner.SetDescription("Applying filters and sorting")

	progressSteps := components.NewProgress(components.ProgressSteps)
	progressSteps.SetTitle("Setup Wizard")
	progressSteps.AddStep(components.ProgressStep{
		ID:          "auth",
		Title:       "Authentication",
		Description: "Connecting to GitHub",
		Status:      components.StepCompleted,
	})
	progressSteps.AddStep(components.ProgressStep{
		ID:          "fetch",
		Title:       "Fetch Notifications",
		Description: "Downloading latest notifications",
		Status:      components.StepInProgress,
	})
	progressSteps.AddStep(components.ProgressStep{
		ID:          "process",
		Title:       "Process Data",
		Description: "Organizing and filtering",
		Status:      components.StepPending,
	})

	// Create layout
	m.layout = components.NewLayout(components.LayoutVertical)
	m.layout.AddComponent("title", m.createTitlePanel("Progress Indicators Demo", "Various progress indicators with different styles"))
	m.layout.AddComponent("bar", progressBar)
	m.layout.AddComponent("spinner", progressSpinner)
	m.layout.AddComponent("steps", progressSteps)
	m.layout.AddComponent("help", m.createHelpPanel("Watch the animated progress indicators"))
}

// setupMarkdownDemo sets up the markdown demonstration
func (m *DemoModel) setupMarkdownDemo() {
	content := `# GitHub Notifications Help

Welcome to **gh-notif**, a high-performance CLI tool for managing GitHub notifications.

## Features

- 🚀 **High Performance**: Virtualized lists handle thousands of notifications
- 🎨 **Beautiful UI**: Modern terminal interface with gradients and animations
- ⌨️  **Keyboard Friendly**: Vim-style navigation and shortcuts
- 🔍 **Powerful Filtering**: Advanced search and filtering capabilities
- 📱 **Responsive**: Adapts to different terminal sizes

## Quick Start

1. **Authentication**: Run ` + "`gh-notif auth login`" + ` to authenticate
2. **List Notifications**: Use ` + "`gh-notif list`" + ` to see your notifications
3. **Interactive UI**: Launch ` + "`gh-notif tui`" + ` for the full experience

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| ` + "`j/k`" + ` | Navigate up/down |
| ` + "`Enter`" + ` | Select notification |
| ` + "`m`" + ` | Mark as read |
| ` + "`o`" + ` | Open in browser |
| ` + "`/`" + ` | Filter notifications |

## Code Example

` + "```go" + `
// Create a new GitHub client
client, err := github.NewClient(ctx)
if err != nil {
    log.Fatal(err)
}

// Fetch notifications
notifications, err := client.GetNotifications()
` + "```" + `

> **Tip**: Use the ` + "`--help`" + ` flag with any command to see detailed usage information.

---

*For more information, visit our [GitHub repository](https://github.com/SharanRP/gh-notif).*`

	m.markdown = components.NewMarkdownRenderer(content)
	m.markdown.SetSize(m.width-4, m.height-8)
	m.markdown.SetFocused(true)

	// Create layout
	m.layout = components.NewLayout(components.LayoutVertical)
	m.layout.AddComponent("title", m.createTitlePanel("Markdown Rendering Demo", "Rich text with syntax highlighting and formatting"))
	m.layout.AddComponent("markdown", m.markdown)
	m.layout.AddComponent("help", m.createHelpPanel("↑/↓: Scroll • j/k: Line by line • g/G: Top/Bottom"))
}

// setupLayoutDemo sets up the layout demonstration
func (m *DemoModel) setupLayoutDemo() {
	// Create multiple panels to demonstrate layout
	panel1 := components.NewPanel("Notifications", components.PanelPrimary)
	panel1.SetContent("📧 15 unread notifications\n📝 5 pull requests\n🐛 3 issues\n📦 2 releases")

	panel2 := components.NewPanel("Statistics", components.PanelSecondary)
	panel2.SetContent("📊 Weekly Activity:\n• Monday: 12\n• Tuesday: 8\n• Wednesday: 15\n• Thursday: 6")

	panel3 := components.NewPanel("Quick Actions", components.PanelBordered)
	panel3.SetContent("⚡ Available Actions:\n• Mark all as read\n• Archive old notifications\n• Sync with GitHub\n• Export to CSV")

	panel4 := components.NewPanel("Recent Activity", components.PanelElevated)
	panel4.SetContent("🕒 Recent Events:\n• New PR opened\n• Issue commented\n• Release published\n• Discussion started")

	// Create grid layout
	m.layout = components.NewLayout(components.LayoutGrid)
	m.layout.AddComponent("panel1", panel1)
	m.layout.AddComponent("panel2", panel2)
	m.layout.AddComponent("panel3", panel3)
	m.layout.AddComponent("panel4", panel4)
	m.layout.SetPadding(1)
	m.layout.SetMargin(1)
}

// setupStylingDemo sets up the styling demonstration
func (m *DemoModel) setupStylingDemo() {
	// Create a panel showcasing various styling features
	content := fmt.Sprintf(`%s

%s

%s

%s

%s

%s`,
		ui.CreateGradientText("Gradient Text Example", m.theme.PrimaryGradient),
		m.styles.BadgePrimary.Render("PRIMARY")+" "+
			m.styles.BadgeSecondary.Render("SECONDARY")+" "+
			m.styles.BadgeSuccess.Render("SUCCESS")+" "+
			m.styles.BadgeWarning.Render("WARNING")+" "+
			m.styles.BadgeError.Render("ERROR"),
		ui.CreateProgressBar(40, 0.75, m.theme),
		m.styles.Glow.Render("✨ Glowing Text Effect ✨"),
		m.styles.PanelElevated.Width(50).Render("Elevated Panel with Border"),
		"🎨 Colors adapt to terminal capabilities",
	)

	panel := components.NewPanel("Enhanced Styling Demo", components.PanelPrimary)
	panel.SetContent(content)
	panel.SetSize(m.width-4, m.height-8)

	// Create layout
	m.layout = components.NewLayout(components.LayoutVertical)
	m.layout.AddComponent("title", m.createTitlePanel("Enhanced Styling Demo", "Gradients, badges, progress bars, and special effects"))
	m.layout.AddComponent("panel", panel)
	m.layout.AddComponent("help", m.createHelpPanel("Observe the various styling effects and animations"))
}

// createTitlePanel creates a title panel for demos
func (m *DemoModel) createTitlePanel(title, description string) *components.Panel {
	content := fmt.Sprintf("%s\n\n%s\n\nDemo %d of %d",
		ui.CreateGradientText(title, m.theme.PrimaryGradient),
		description,
		m.currentDemo+1,
		len(m.demos),
	)

	panel := components.NewPanel("", components.PanelBordered)
	panel.SetContent(content)
	return panel
}

// createHelpPanel creates a help panel for demos
func (m *DemoModel) createHelpPanel(help string) *components.Panel {
	content := fmt.Sprintf("%s • n: Next demo • p: Previous demo • q: Quit", help)

	panel := components.NewPanel("", components.PanelDefault)
	panel.SetContent(content)
	return panel
}

// createSampleNotifications creates sample notifications for testing
func (m *DemoModel) createSampleNotifications(count int) []*github.Notification {
	notifications := make([]*github.Notification, count)

	repos := []string{"microsoft/vscode", "golang/go", "kubernetes/kubernetes", "facebook/react", "torvalds/linux"}
	types := []string{"PullRequest", "Issue", "Release", "Discussion", "Commit"}
	reasons := []string{"assign", "author", "comment", "mention", "review_requested", "subscribed"}

	for i := 0; i < count; i++ {
		repo := repos[i%len(repos)]
		notifType := types[i%len(types)]
		reason := reasons[i%len(reasons)]
		unread := i%3 == 0 // Every third notification is unread

		notification := &github.Notification{
			ID: github.String(fmt.Sprintf("notification_%d", i)),
			Repository: &github.Repository{
				FullName: github.String(repo),
			},
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("Sample notification %d: %s", i+1, notifType)),
				Type:  github.String(notifType),
				URL:   github.String(fmt.Sprintf("https://github.com/%s/issues/%d", repo, i+1)),
			},
			Reason:    github.String(reason),
			Unread:    github.Bool(unread),
			UpdatedAt: &github.Timestamp{Time: time.Now().Add(-time.Duration(i) * time.Hour)},
		}

		notifications[i] = notification
	}

	return notifications
}

// Init initializes the demo model
func (m *DemoModel) Init() tea.Cmd {
	return m.layout.Init()
}

// Update handles messages and updates the demo model
func (m *DemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(m.width, m.height)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n":
			m.nextDemo()
		case "p":
			m.prevDemo()
		default:
			var cmd tea.Cmd
			updatedLayout, cmd := m.layout.Update(msg)
			m.layout = updatedLayout.(*components.Layout)
			return m, cmd
		}
	}

	var cmd tea.Cmd
	updatedLayout, cmd := m.layout.Update(msg)
	m.layout = updatedLayout.(*components.Layout)
	return m, cmd
}

// View renders the demo model
func (m *DemoModel) View() string {
	if m.layout == nil {
		return "Loading demo..."
	}

	return m.layout.View()
}

// nextDemo switches to the next demo
func (m *DemoModel) nextDemo() {
	m.currentDemo = (m.currentDemo + 1) % len(m.demos)
	m.demos[m.currentDemo].SetupFunc(m)
	m.layout.SetSize(m.width, m.height)
}

// prevDemo switches to the previous demo
func (m *DemoModel) prevDemo() {
	m.currentDemo--
	if m.currentDemo < 0 {
		m.currentDemo = len(m.demos) - 1
	}
	m.demos[m.currentDemo].SetupFunc(m)
	m.layout.SetSize(m.width, m.height)
}

// RunEnhancedUIDemo runs the enhanced UI demonstration
func RunEnhancedUIDemo() error {
	model := NewDemoModel()

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()

	return err
}
