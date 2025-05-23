package components

import (
	"github.com/charmbracelet/lipgloss"
)

// EnhancedStyles contains enhanced styling for components
type EnhancedStyles struct {
	// Base styles
	App            lipgloss.Style
	Header         lipgloss.Style
	StatusBar      lipgloss.Style
	HelpBar        lipgloss.Style
	Spinner        lipgloss.Style
	Error          lipgloss.Style
	FilterPrompt   lipgloss.Style
	FilterInput    lipgloss.Style

	// List styles
	List           lipgloss.Style
	ListItem       lipgloss.Style
	SelectedItem   lipgloss.Style
	UnreadItem     lipgloss.Style
	ReadItem       lipgloss.Style

	// Gradient styles
	HeaderGradient    lipgloss.Style
	AccentGradient    lipgloss.Style
	ProgressGradient  lipgloss.Style

	// Badge styles
	BadgePrimary      lipgloss.Style
	BadgeSecondary    lipgloss.Style
	BadgeSuccess      lipgloss.Style
	BadgeWarning      lipgloss.Style
	BadgeError        lipgloss.Style
	BadgeInfo         lipgloss.Style

	// Panel styles
	PanelPrimary      lipgloss.Style
	PanelSecondary    lipgloss.Style
	PanelBordered     lipgloss.Style
	PanelElevated     lipgloss.Style

	// Indicator styles
	UnreadIndicator   lipgloss.Style
	ReadIndicator     lipgloss.Style
}

// Symbols contains Unicode symbols used in the UI
type Symbols struct {
	UnreadIndicator string
	ReadIndicator   string
	PullRequest     string
	Issue           string
	Release         string
	Discussion      string
	Commit          string
	Check           string
	Cross           string
	Warning         string
	Info            string
	ArrowRight      string
	ArrowLeft       string
	ArrowUp         string
	ArrowDown       string
	Ellipsis        string
	Star            string
	Dot             string
}

// DefaultSymbols returns the default symbols with beautiful icons
func DefaultSymbols() Symbols {
	return Symbols{
		UnreadIndicator: "🔴",  // Red circle for unread
		ReadIndicator:   "⚪",  // White circle for read
		PullRequest:     "🔀",  // Merge symbol for PRs
		Issue:           "🐛",  // Bug for issues
		Release:         "🚀",  // Rocket for releases
		Discussion:      "💬",  // Speech bubble for discussions
		Commit:          "📝",  // Memo for commits
		Check:           "✅",  // Green check mark
		Cross:           "❌",  // Red X
		Warning:         "⚠️",   // Warning sign
		Info:            "ℹ️",   // Information
		ArrowRight:      "▶️",   // Play button right
		ArrowLeft:       "◀️",   // Play button left
		ArrowUp:         "🔼",  // Up triangle
		ArrowDown:       "🔽",  // Down triangle
		Ellipsis:        "⋯",   // Horizontal ellipsis
		Star:            "⭐",  // Star
		Dot:             "•",   // Bullet point
	}
}

// DefaultEnhancedStyles returns default enhanced styles
func DefaultEnhancedStyles() EnhancedStyles {
	return EnhancedStyles{
		// Base styles
		App: lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E2E")).
			Foreground(lipgloss.Color("#CDD6F4")),

		Header: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true).
			Padding(0, 1),

		StatusBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Padding(0, 1),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")).
			Padding(0, 1),

		Spinner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true),

		FilterPrompt: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true),

		FilterInput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")),

		// List styles
		List: lipgloss.NewStyle().
			Padding(0, 1),

		ListItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CDD6F4")).
			Padding(0, 1),

		SelectedItem: lipgloss.NewStyle().
			Background(lipgloss.Color("#45475A")).
			Foreground(lipgloss.Color("#F5E0DC")).
			Bold(true).
			Padding(0, 1),

		UnreadItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5E0DC")).
			Bold(true),

		ReadItem: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")),

		// Gradient styles
		HeaderGradient: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89B4FA")).
			Bold(true).
			Padding(0, 1),

		AccentGradient: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1")).
			Bold(true),

		ProgressGradient: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#89DCEB")),

		// Badge styles with rounded borders and better spacing
		BadgePrimary: lipgloss.NewStyle().
			Background(lipgloss.Color("#89B4FA")).
			Foreground(lipgloss.Color("#1E1E2E")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#74C7EC")),

		BadgeSecondary: lipgloss.NewStyle().
			Background(lipgloss.Color("#6C7086")).
			Foreground(lipgloss.Color("#F5E0DC")).
			Padding(0, 2).
			Margin(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#585B70")),

		BadgeSuccess: lipgloss.NewStyle().
			Background(lipgloss.Color("#A6E3A1")).
			Foreground(lipgloss.Color("#1E1E2E")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#94E2D5")),

		BadgeWarning: lipgloss.NewStyle().
			Background(lipgloss.Color("#FAB387")).
			Foreground(lipgloss.Color("#1E1E2E")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F9E2AF")),

		BadgeError: lipgloss.NewStyle().
			Background(lipgloss.Color("#F38BA8")).
			Foreground(lipgloss.Color("#1E1E2E")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F2CDCD")),

		BadgeInfo: lipgloss.NewStyle().
			Background(lipgloss.Color("#89DCEB")).
			Foreground(lipgloss.Color("#1E1E2E")).
			Padding(0, 2).
			Margin(0, 1).
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#74C7EC")),

		// Panel styles with enhanced borders and spacing
		PanelPrimary: lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E2E")).
			Foreground(lipgloss.Color("#CDD6F4")).
			Padding(2, 3).
			Margin(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89B4FA")).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true),

		PanelSecondary: lipgloss.NewStyle().
			Background(lipgloss.Color("#181825")).
			Foreground(lipgloss.Color("#BAC2DE")).
			Padding(2, 3).
			Margin(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#45475A")).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true),

		PanelBordered: lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1E2E")).
			Foreground(lipgloss.Color("#CDD6F4")).
			Padding(2, 3).
			Margin(1, 2).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("#A6E3A1")).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true),

		PanelElevated: lipgloss.NewStyle().
			Background(lipgloss.Color("#313244")).
			Foreground(lipgloss.Color("#F5E0DC")).
			Padding(2, 3).
			Margin(1, 2).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#74C7EC")).
			BorderTop(true).
			BorderBottom(true).
			BorderLeft(true).
			BorderRight(true),

		// Indicator styles
		UnreadIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8")).
			Bold(true),

		ReadIndicator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086")),
	}
}
