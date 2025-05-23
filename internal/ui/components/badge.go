package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BadgeType represents different types of badges
type BadgeType int

const (
	// BadgePrimary is the primary badge type
	BadgePrimary BadgeType = iota
	// BadgeSecondary is the secondary badge type
	BadgeSecondary
	// BadgeSuccess indicates success
	BadgeSuccess
	// BadgeWarning indicates warning
	BadgeWarning
	// BadgeError indicates error
	BadgeError
	// BadgeInfo indicates information
	BadgeInfo
	// BadgeCustom allows custom styling
	BadgeCustom
)

// Badge represents a badge component
type Badge struct {
	// Configuration
	text      string
	badgeType BadgeType

	// Custom styling (for BadgeCustom type)
	customStyle lipgloss.Style

	// State
	focused bool

	// Styling
	styles ComponentStyles
}

// NewBadge creates a new badge component
func NewBadge(text string, badgeType BadgeType) *Badge {
	return &Badge{
		text:      text,
		badgeType: badgeType,
	}
}

// NewBadgeComponentFactory creates a badge component factory
func NewBadgeComponentFactory(config ComponentConfig) Component {
	text, ok := config.Props["text"].(string)
	if !ok {
		text = "Badge"
	}

	badgeType, ok := config.Props["type"].(BadgeType)
	if !ok {
		badgeType = BadgePrimary
	}

	badge := NewBadge(text, badgeType)
	badge.SetStyles(config.Styles)

	if customStyle, ok := config.Props["customStyle"].(lipgloss.Style); ok {
		badge.SetCustomStyle(customStyle)
	}

	return badge
}

// SetText sets the badge text
func (b *Badge) SetText(text string) {
	b.text = text
}

// GetText returns the badge text
func (b *Badge) GetText() string {
	return b.text
}

// SetBadgeType sets the badge type
func (b *Badge) SetBadgeType(badgeType BadgeType) {
	b.badgeType = badgeType
}

// GetBadgeType returns the badge type
func (b *Badge) GetBadgeType() BadgeType {
	return b.badgeType
}

// SetCustomStyle sets a custom style (only used with BadgeCustom type)
func (b *Badge) SetCustomStyle(style lipgloss.Style) {
	b.customStyle = style
}

// Init initializes the badge component
func (b *Badge) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the badge state
func (b *Badge) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case ComponentMessage:
		switch msg.Type {
		case "text":
			if text, ok := msg.Data.(string); ok {
				b.SetText(text)
			}
		case "type":
			if badgeType, ok := msg.Data.(BadgeType); ok {
				b.SetBadgeType(badgeType)
			}
		}
	}

	return b, nil
}

// View renders the badge
func (b *Badge) View() string {
	var style lipgloss.Style

	switch b.badgeType {
	case BadgePrimary:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Bold(true)
	case BadgeSecondary:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("8")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)
	case BadgeSuccess:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("2")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Bold(true)
	case BadgeWarning:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("3")).
			Foreground(lipgloss.Color("0")).
			Padding(0, 1).
			Bold(true)
	case BadgeError:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("1")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Bold(true)
	case BadgeInfo:
		style = lipgloss.NewStyle().
			Background(lipgloss.Color("6")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1).
			Bold(true)
	case BadgeCustom:
		style = b.customStyle
	default:
		style = b.styles.Base
	}

	return style.Render(b.text)
}

// SetSize sets the component dimensions (not applicable for badges)
func (b *Badge) SetSize(width, height int) {
	// Badges don't have fixed dimensions
}

// GetSize returns the component dimensions
func (b *Badge) GetSize() (width, height int) {
	// Calculate size based on text length
	return len(b.text) + 2, 1 // +2 for padding
}

// SetStyles sets the component styles
func (b *Badge) SetStyles(styles ComponentStyles) {
	b.styles = styles
}

// GetType returns the component type
func (b *Badge) GetType() ComponentType {
	return BadgeComponentType
}

// SetFocused sets the focus state
func (b *Badge) SetFocused(focused bool) {
	b.focused = focused
}

// IsFocused returns the focus state
func (b *Badge) IsFocused() bool {
	return b.focused
}

// Panel represents a panel component for grouping content
type Panel struct {
	// Configuration
	width     int
	height    int
	title     string
	content   string
	panelType PanelType

	// State
	focused bool

	// Styling
	styles ComponentStyles
}

// PanelType represents different types of panels
type PanelType int

const (
	// PanelDefault is the default panel type
	PanelDefault PanelType = iota
	// PanelPrimary is the primary panel type
	PanelPrimary
	// PanelSecondary is the secondary panel type
	PanelSecondary
	// PanelBordered has a prominent border
	PanelBordered
	// PanelElevated appears elevated
	PanelElevated
)

// NewPanel creates a new panel component
func NewPanel(title string, panelType PanelType) *Panel {
	return &Panel{
		title:     title,
		panelType: panelType,
	}
}

// NewPanelComponentFactory creates a panel component factory
func NewPanelComponentFactory(config ComponentConfig) Component {
	title, ok := config.Props["title"].(string)
	if !ok {
		title = ""
	}

	panelType, ok := config.Props["type"].(PanelType)
	if !ok {
		panelType = PanelDefault
	}

	panel := NewPanel(title, panelType)
	panel.SetSize(config.Width, config.Height)
	panel.SetStyles(config.Styles)

	if content, ok := config.Props["content"].(string); ok {
		panel.SetContent(content)
	}

	return panel
}

// SetTitle sets the panel title
func (p *Panel) SetTitle(title string) {
	p.title = title
}

// GetTitle returns the panel title
func (p *Panel) GetTitle() string {
	return p.title
}

// SetContent sets the panel content
func (p *Panel) SetContent(content string) {
	p.content = content
}

// GetContent returns the panel content
func (p *Panel) GetContent() string {
	return p.content
}

// SetPanelType sets the panel type
func (p *Panel) SetPanelType(panelType PanelType) {
	p.panelType = panelType
}

// GetPanelType returns the panel type
func (p *Panel) GetPanelType() PanelType {
	return p.panelType
}

// Init initializes the panel component
func (p *Panel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the panel state
func (p *Panel) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				p.SetSize(size.Width, size.Height)
			}
		case "title":
			if title, ok := msg.Data.(string); ok {
				p.SetTitle(title)
			}
		case "content":
			if content, ok := msg.Data.(string); ok {
				p.SetContent(content)
			}
		case "type":
			if panelType, ok := msg.Data.(PanelType); ok {
				p.SetPanelType(panelType)
			}
		}
	}

	return p, nil
}

// View renders the panel
func (p *Panel) View() string {
	var style lipgloss.Style

	switch p.panelType {
	case PanelDefault:
		style = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8"))
	case PanelPrimary:
		style = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("4"))
	case PanelSecondary:
		style = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8"))
	case PanelBordered:
		style = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("4"))
	case PanelElevated:
		style = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("8"))
	}

	// Apply focus styling
	if p.focused {
		style = style.BorderForeground(lipgloss.Color("6"))
	}

	// Set dimensions
	if p.width > 0 {
		style = style.Width(p.width - 4) // Account for border and padding
	}
	if p.height > 0 {
		style = style.Height(p.height - 4) // Account for border and padding
	}

	// Build content
	var parts []string

	if p.title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true)
		if p.focused {
			titleStyle = titleStyle.Foreground(lipgloss.Color("6"))
		}
		parts = append(parts, titleStyle.Render(p.title))

		if p.content != "" {
			parts = append(parts, "")
		}
	}

	if p.content != "" {
		parts = append(parts, p.content)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return style.Render(content)
}

// SetSize sets the component dimensions
func (p *Panel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// GetSize returns the component dimensions
func (p *Panel) GetSize() (width, height int) {
	return p.width, p.height
}

// SetStyles sets the component styles
func (p *Panel) SetStyles(styles ComponentStyles) {
	p.styles = styles
}

// GetType returns the component type
func (p *Panel) GetType() ComponentType {
	return PanelComponentType
}

// SetFocused sets the focus state
func (p *Panel) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns the focus state
func (p *Panel) IsFocused() bool {
	return p.focused
}
