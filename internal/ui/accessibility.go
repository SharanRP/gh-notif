package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// AccessibilityMode represents different accessibility modes
type AccessibilityMode int

const (
	// StandardMode is the default mode
	StandardMode AccessibilityMode = iota
	// ScreenReaderMode is optimized for screen readers
	ScreenReaderMode
	// HighContrastMode uses high contrast colors
	HighContrastMode
	// LargeTextMode uses larger text
	LargeTextMode
)

// AccessibilitySettings contains settings for accessibility
type AccessibilitySettings struct {
	Mode            AccessibilityMode
	ColorScheme     ColorScheme
	UseUnicode      bool
	UseAnimations   bool
	KeyboardOnly    bool
	DescriptiveText bool
}

// DefaultAccessibilitySettings returns the default accessibility settings
func DefaultAccessibilitySettings() AccessibilitySettings {
	return AccessibilitySettings{
		Mode:            StandardMode,
		ColorScheme:     DarkScheme,
		UseUnicode:      true,
		UseAnimations:   true,
		KeyboardOnly:    false,
		DescriptiveText: false,
	}
}

// ScreenReaderDescription generates a screen reader friendly description of a notification
func ScreenReaderDescription(notification *github.Notification) string {
	if notification == nil {
		return "No notification selected"
	}

	var sb strings.Builder

	// Add read status
	if notification.GetUnread() {
		sb.WriteString("Unread notification. ")
	} else {
		sb.WriteString("Read notification. ")
	}

	// Add type
	sb.WriteString(fmt.Sprintf("Type: %s. ", notification.GetSubject().GetType()))

	// Add repository
	sb.WriteString(fmt.Sprintf("Repository: %s. ", notification.GetRepository().GetFullName()))

	// Add title
	sb.WriteString(fmt.Sprintf("Title: %s. ", notification.GetSubject().GetTitle()))

	// Add updated time
	sb.WriteString(fmt.Sprintf("Updated: %s. ", notification.GetUpdatedAt().Format("January 2, 2006 at 3:04 PM")))

	return sb.String()
}

// GetAccessibleSymbols returns symbols appropriate for the given accessibility settings
func GetAccessibleSymbols(settings AccessibilitySettings) Symbols {
	if settings.UseUnicode {
		return DefaultSymbols()
	}

	// Use ASCII-only symbols
	return Symbols{
		UnreadIndicator: "*",
		ReadIndicator:   "o",
		PullRequest:     "PR",
		Issue:           "I",
		Release:         "R",
		Discussion:      "D",
		Commit:          "C",
		Check:           "+",
		Cross:           "x",
		Warning:         "!",
		Info:            "i",
		ArrowRight:      ">",
		ArrowLeft:       "<",
		ArrowUp:         "^",
		ArrowDown:       "v",
		Ellipsis:        "...",
		Star:            "*",
		Dot:             ".",
	}
}

// GetAccessibleTheme returns a theme appropriate for the given accessibility settings
func GetAccessibleTheme(settings AccessibilitySettings) Theme {
	switch settings.ColorScheme {
	case DarkScheme:
		if settings.Mode == HighContrastMode {
			return HighContrastTheme()
		}
		return DefaultDarkTheme()
	case LightScheme:
		return DefaultLightTheme()
	case HighContrastScheme:
		return HighContrastTheme()
	default:
		return DefaultDarkTheme()
	}
}

// GetAccessibleStyles returns styles appropriate for the given accessibility settings
func GetAccessibleStyles(settings AccessibilitySettings) Styles {
	theme := GetAccessibleTheme(settings)
	styles := NewStyles(theme)

	// Modify styles based on accessibility mode
	switch settings.Mode {
	case LargeTextMode:
		// Increase font size for all styles
		// Note: This is a simplification as terminal font size control is limited
		styles.Header = styles.Header.Bold(true).Padding(1, 2)
		styles.ListItem = styles.ListItem.Padding(1, 2)
		styles.SelectedItem = styles.SelectedItem.Padding(1, 2)
		styles.DetailHeader = styles.DetailHeader.Bold(true).Padding(1, 2)
		styles.TableHeader = styles.TableHeader.Bold(true).Padding(1, 2)
		styles.TableCell = styles.TableCell.Padding(1, 2)
	case ScreenReaderMode:
		// Optimize for screen readers
		// Remove visual styling that might interfere with screen readers
		styles.App = lipgloss.NewStyle()
		styles.Header = lipgloss.NewStyle().Bold(true)
		styles.StatusBar = lipgloss.NewStyle()
		styles.HelpBar = lipgloss.NewStyle()
		styles.List = lipgloss.NewStyle()
		styles.ListItem = lipgloss.NewStyle()
		styles.SelectedItem = lipgloss.NewStyle().Bold(true)
		styles.DetailView = lipgloss.NewStyle()
		styles.Table = lipgloss.NewStyle()
	}

	return styles
}

// GenerateAccessibleHelp generates help text appropriate for the given accessibility settings
func GenerateAccessibleHelp(settings AccessibilitySettings) string {
	var sb strings.Builder

	sb.WriteString("Keyboard Shortcuts:\n\n")

	// Navigation
	sb.WriteString("Navigation:\n")
	sb.WriteString("  Up/Down Arrow or j/k: Move selection up/down\n")
	sb.WriteString("  Left/Right Arrow or h/l: Move between panels\n")
	sb.WriteString("  Enter: Select notification\n\n")

	// Actions
	sb.WriteString("Actions:\n")
	sb.WriteString("  m: Mark selected notification as read\n")
	sb.WriteString("  M: Mark all notifications as read\n")
	sb.WriteString("  o: Open selected notification in browser\n")
	sb.WriteString("  r: Refresh notifications\n\n")

	// View controls
	sb.WriteString("View Controls:\n")
	sb.WriteString("  v: Change view mode (compact, detailed, split, table)\n")
	sb.WriteString("  c: Change color scheme (dark, light, high contrast)\n")
	sb.WriteString("  /: Filter notifications\n")
	sb.WriteString("  ?: Toggle help\n")
	sb.WriteString("  q: Quit\n\n")

	// Accessibility specific
	if settings.Mode != StandardMode {
		sb.WriteString("Accessibility Controls:\n")
		sb.WriteString("  a: Toggle accessibility modes\n")
		sb.WriteString("  A: Toggle animations\n")
		sb.WriteString("  u: Toggle Unicode/ASCII symbols\n")
		sb.WriteString("  d: Toggle descriptive text\n\n")
	}

	return sb.String()
}

// GenerateAccessibleStatusBar generates a status bar appropriate for the given accessibility settings
func GenerateAccessibleStatusBar(model Model, settings AccessibilitySettings) string {
	var sb strings.Builder

	// Basic information
	sb.WriteString(fmt.Sprintf("%d notifications", len(model.notifications)))
	
	if len(model.filteredItems) != len(model.notifications) {
		sb.WriteString(fmt.Sprintf(" (%d filtered)", len(model.filteredItems)))
	}

	// Add view mode
	switch model.viewMode {
	case CompactView:
		sb.WriteString(" | Compact View")
	case DetailedView:
		sb.WriteString(" | Detailed View")
	case SplitView:
		sb.WriteString(" | Split View")
	case TableView:
		sb.WriteString(" | Table View")
	}

	// Add accessibility mode
	switch settings.Mode {
	case ScreenReaderMode:
		sb.WriteString(" | Screen Reader Mode")
	case HighContrastMode:
		sb.WriteString(" | High Contrast Mode")
	case LargeTextMode:
		sb.WriteString(" | Large Text Mode")
	}

	// Add selected notification info if in screen reader mode
	if settings.Mode == ScreenReaderMode && model.getSelectedNotification() != nil {
		n := model.getSelectedNotification()
		sb.WriteString(fmt.Sprintf(" | Selected: %s", n.GetSubject().GetTitle()))
	}

	return sb.String()
}
