package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette for the UI
type Theme struct {
	// Base colors
	Background    lipgloss.Color
	Foreground    lipgloss.Color
	DimmedText    lipgloss.Color
	HighlightText lipgloss.Color
	AccentColor   lipgloss.Color
	ErrorColor    lipgloss.Color
	SuccessColor  lipgloss.Color
	WarningColor  lipgloss.Color
	InfoColor     lipgloss.Color

	// UI element colors
	BorderColor      lipgloss.Color
	SelectedBg       lipgloss.Color
	SelectedFg       lipgloss.Color
	HeaderColor      lipgloss.Color
	StatusBarColor   lipgloss.Color
	HelpTextColor    lipgloss.Color
	UnreadColor      lipgloss.Color
	ReadColor        lipgloss.Color
	FilterMatchColor lipgloss.Color

	// Notification type colors
	IssueColor       lipgloss.Color
	PullRequestColor lipgloss.Color
	ReleaseColor     lipgloss.Color
	DiscussionColor  lipgloss.Color
	CommitColor      lipgloss.Color
}

// Styles contains all the styles for the UI
type Styles struct {
	// Base styles
	App             lipgloss.Style
	Header          lipgloss.Style
	StatusBar       lipgloss.Style
	HelpBar         lipgloss.Style
	Spinner         lipgloss.Style
	Error           lipgloss.Style
	FilterPrompt    lipgloss.Style
	FilterInput     lipgloss.Style
	NoNotifications lipgloss.Style

	// List styles
	List         lipgloss.Style
	ListItem     lipgloss.Style
	SelectedItem lipgloss.Style
	UnreadItem   lipgloss.Style
	ReadItem     lipgloss.Style

	// Detail view styles
	DetailView   lipgloss.Style
	DetailHeader lipgloss.Style
	DetailBody   lipgloss.Style
	DetailFooter lipgloss.Style

	// Split view styles
	SplitLeft    lipgloss.Style
	SplitRight   lipgloss.Style
	SplitDivider lipgloss.Style

	// Table styles
	Table            lipgloss.Style
	TableHeader      lipgloss.Style
	TableCell        lipgloss.Style
	TableSelectedRow lipgloss.Style

	// Notification type styles
	Issue       lipgloss.Style
	PullRequest lipgloss.Style
	Release     lipgloss.Style
	Discussion  lipgloss.Style
	Commit      lipgloss.Style

	// Status indicators
	UnreadIndicator lipgloss.Style
	ReadIndicator   lipgloss.Style
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

// DefaultSymbols returns the default symbols
func DefaultSymbols() Symbols {
	return Symbols{
		UnreadIndicator: "●",
		ReadIndicator:   "○",
		PullRequest:     "⟳",
		Issue:           "◉",
		Release:         "⬇",
		Discussion:      "💬",
		Commit:          "◯",
		Check:           "✓",
		Cross:           "✗",
		Warning:         "⚠",
		Info:            "ℹ",
		ArrowRight:      "→",
		ArrowLeft:       "←",
		ArrowUp:         "↑",
		ArrowDown:       "↓",
		Ellipsis:        "…",
		Star:            "★",
		Dot:             "•",
	}
}

// DefaultDarkTheme returns the default dark theme
func DefaultDarkTheme() Theme {
	return Theme{
		Background:       lipgloss.Color("#1E1E2E"),
		Foreground:       lipgloss.Color("#CDD6F4"),
		DimmedText:       lipgloss.Color("#6C7086"),
		HighlightText:    lipgloss.Color("#F5E0DC"),
		AccentColor:      lipgloss.Color("#89B4FA"),
		ErrorColor:       lipgloss.Color("#F38BA8"),
		SuccessColor:     lipgloss.Color("#A6E3A1"),
		WarningColor:     lipgloss.Color("#FAB387"),
		InfoColor:        lipgloss.Color("#89DCEB"),
		BorderColor:      lipgloss.Color("#313244"),
		SelectedBg:       lipgloss.Color("#45475A"),
		SelectedFg:       lipgloss.Color("#F5E0DC"),
		HeaderColor:      lipgloss.Color("#89B4FA"),
		StatusBarColor:   lipgloss.Color("#6C7086"),
		HelpTextColor:    lipgloss.Color("#6C7086"),
		UnreadColor:      lipgloss.Color("#A6E3A1"),
		ReadColor:        lipgloss.Color("#6C7086"),
		FilterMatchColor: lipgloss.Color("#F9E2AF"),
		IssueColor:       lipgloss.Color("#F38BA8"),
		PullRequestColor: lipgloss.Color("#89B4FA"),
		ReleaseColor:     lipgloss.Color("#A6E3A1"),
		DiscussionColor:  lipgloss.Color("#CBA6F7"),
		CommitColor:      lipgloss.Color("#FAB387"),
	}
}

// DefaultLightTheme returns the default light theme
func DefaultLightTheme() Theme {
	return Theme{
		Background:       lipgloss.Color("#EFF1F5"),
		Foreground:       lipgloss.Color("#4C4F69"),
		DimmedText:       lipgloss.Color("#9CA0B0"),
		HighlightText:    lipgloss.Color("#DC8A78"),
		AccentColor:      lipgloss.Color("#1E66F5"),
		ErrorColor:       lipgloss.Color("#D20F39"),
		SuccessColor:     lipgloss.Color("#40A02B"),
		WarningColor:     lipgloss.Color("#FE640B"),
		InfoColor:        lipgloss.Color("#209FB5"),
		BorderColor:      lipgloss.Color("#DCE0E8"),
		SelectedBg:       lipgloss.Color("#CCD0DA"),
		SelectedFg:       lipgloss.Color("#DC8A78"),
		HeaderColor:      lipgloss.Color("#1E66F5"),
		StatusBarColor:   lipgloss.Color("#9CA0B0"),
		HelpTextColor:    lipgloss.Color("#9CA0B0"),
		UnreadColor:      lipgloss.Color("#40A02B"),
		ReadColor:        lipgloss.Color("#9CA0B0"),
		FilterMatchColor: lipgloss.Color("#DF8E1D"),
		IssueColor:       lipgloss.Color("#D20F39"),
		PullRequestColor: lipgloss.Color("#1E66F5"),
		ReleaseColor:     lipgloss.Color("#40A02B"),
		DiscussionColor:  lipgloss.Color("#8839EF"),
		CommitColor:      lipgloss.Color("#FE640B"),
	}
}

// HighContrastTheme returns a high contrast theme for accessibility
func HighContrastTheme() Theme {
	return Theme{
		Background:       lipgloss.Color("#000000"),
		Foreground:       lipgloss.Color("#FFFFFF"),
		DimmedText:       lipgloss.Color("#AAAAAA"),
		HighlightText:    lipgloss.Color("#FFFF00"),
		AccentColor:      lipgloss.Color("#00FFFF"),
		ErrorColor:       lipgloss.Color("#FF0000"),
		SuccessColor:     lipgloss.Color("#00FF00"),
		WarningColor:     lipgloss.Color("#FFAA00"),
		InfoColor:        lipgloss.Color("#00AAFF"),
		BorderColor:      lipgloss.Color("#FFFFFF"),
		SelectedBg:       lipgloss.Color("#0000AA"),
		SelectedFg:       lipgloss.Color("#FFFFFF"),
		HeaderColor:      lipgloss.Color("#FFFF00"),
		StatusBarColor:   lipgloss.Color("#FFFFFF"),
		HelpTextColor:    lipgloss.Color("#FFFFFF"),
		UnreadColor:      lipgloss.Color("#00FF00"),
		ReadColor:        lipgloss.Color("#AAAAAA"),
		FilterMatchColor: lipgloss.Color("#FFFF00"),
		IssueColor:       lipgloss.Color("#FF0000"),
		PullRequestColor: lipgloss.Color("#00FFFF"),
		ReleaseColor:     lipgloss.Color("#00FF00"),
		DiscussionColor:  lipgloss.Color("#FF00FF"),
		CommitColor:      lipgloss.Color("#FFAA00"),
	}
}

// NewStyles creates a new Styles instance based on the given theme
func NewStyles(theme Theme) Styles {
	s := Styles{}

	// Base styles
	s.App = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground)

	s.Header = lipgloss.NewStyle().
		Foreground(theme.HeaderColor).
		Bold(true).
		Padding(0, 1)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(theme.StatusBarColor).
		Padding(0, 1)

	s.HelpBar = lipgloss.NewStyle().
		Foreground(theme.HelpTextColor).
		Padding(0, 1)

	s.Spinner = lipgloss.NewStyle().
		Foreground(theme.AccentColor)

	s.Error = lipgloss.NewStyle().
		Foreground(theme.ErrorColor).
		Bold(true)

	s.FilterPrompt = lipgloss.NewStyle().
		Foreground(theme.AccentColor).
		Bold(true)

	s.FilterInput = lipgloss.NewStyle().
		Foreground(theme.Foreground)

	s.NoNotifications = lipgloss.NewStyle().
		Foreground(theme.DimmedText).
		Italic(true).
		Align(lipgloss.Center)

	// List styles
	s.List = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor).
		Padding(0, 1)

	s.ListItem = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Padding(0, 1)

	s.SelectedItem = lipgloss.NewStyle().
		Background(theme.SelectedBg).
		Foreground(theme.SelectedFg).
		Bold(true).
		Padding(0, 1)

	s.UnreadItem = lipgloss.NewStyle().
		Foreground(theme.UnreadColor).
		Bold(true)

	s.ReadItem = lipgloss.NewStyle().
		Foreground(theme.ReadColor)

	// Detail view styles
	s.DetailView = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor).
		Padding(1, 2)

	s.DetailHeader = lipgloss.NewStyle().
		Foreground(theme.HeaderColor).
		Bold(true).
		Underline(true).
		Padding(0, 0, 1, 0)

	s.DetailBody = lipgloss.NewStyle().
		Foreground(theme.Foreground)

	s.DetailFooter = lipgloss.NewStyle().
		Foreground(theme.DimmedText).
		Italic(true)

	// Split view styles
	s.SplitLeft = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor).
		Padding(0, 1)

	s.SplitRight = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor).
		Padding(1, 2)

	s.SplitDivider = lipgloss.NewStyle().
		Foreground(theme.BorderColor)

	// Table styles
	s.Table = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor)

	s.TableHeader = lipgloss.NewStyle().
		Foreground(theme.HeaderColor).
		Bold(true).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(theme.BorderColor)

	s.TableCell = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Padding(0, 1)

	s.TableSelectedRow = lipgloss.NewStyle().
		Background(theme.SelectedBg).
		Foreground(theme.SelectedFg).
		Bold(true)

	// Notification type styles
	s.Issue = lipgloss.NewStyle().
		Foreground(theme.IssueColor)

	s.PullRequest = lipgloss.NewStyle().
		Foreground(theme.PullRequestColor)

	s.Release = lipgloss.NewStyle().
		Foreground(theme.ReleaseColor)

	s.Discussion = lipgloss.NewStyle().
		Foreground(theme.DiscussionColor)

	s.Commit = lipgloss.NewStyle().
		Foreground(theme.CommitColor)

	// Status indicators
	s.UnreadIndicator = lipgloss.NewStyle().
		Foreground(theme.UnreadColor).
		Bold(true)

	s.ReadIndicator = lipgloss.NewStyle().
		Foreground(theme.ReadColor)

	return s
}
