package ui

import (
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// EnhancedTheme extends the basic theme with advanced styling capabilities
type EnhancedTheme struct {
	Theme

	// Gradient colors
	PrimaryGradient   []lipgloss.Color
	SecondaryGradient []lipgloss.Color
	AccentGradient    []lipgloss.Color

	// Animation colors
	PulseColors   []lipgloss.Color
	ShimmerColors []lipgloss.Color

	// Special effects
	GlowColor   lipgloss.Color
	ShadowColor lipgloss.Color

	// Adaptive colors (based on terminal capabilities)
	SupportsTrue  bool
	Supports256   bool
	SupportsBasic bool
}

// EnhancedStyles extends the basic styles with advanced effects
type EnhancedStyles struct {
	Styles

	// Gradient styles
	HeaderGradient   lipgloss.Style
	AccentGradient   lipgloss.Style
	ProgressGradient lipgloss.Style

	// Animation styles
	Pulse   lipgloss.Style
	Shimmer lipgloss.Style
	Glow    lipgloss.Style

	// Badge styles
	BadgePrimary   lipgloss.Style
	BadgeSecondary lipgloss.Style
	BadgeSuccess   lipgloss.Style
	BadgeWarning   lipgloss.Style
	BadgeError     lipgloss.Style
	BadgeInfo      lipgloss.Style

	// Panel styles
	PanelPrimary   lipgloss.Style
	PanelSecondary lipgloss.Style
	PanelBordered  lipgloss.Style
	PanelElevated  lipgloss.Style

	// Interactive styles
	ButtonPrimary   lipgloss.Style
	ButtonSecondary lipgloss.Style
	ButtonDisabled  lipgloss.Style
	ButtonHover     lipgloss.Style
	ButtonPressed   lipgloss.Style

	// Form styles
	InputFocused lipgloss.Style
	InputError   lipgloss.Style
	InputSuccess lipgloss.Style
	Label        lipgloss.Style
	Placeholder  lipgloss.Style

	// Layout styles
	Container   lipgloss.Style
	Sidebar     lipgloss.Style
	MainContent lipgloss.Style
	Footer      lipgloss.Style
}

// AnimationState tracks animation state
type AnimationState struct {
	StartTime    time.Time
	Duration     time.Duration
	CurrentFrame int
	TotalFrames  int
	Easing       EasingFunction
}

// EasingFunction defines animation easing
type EasingFunction func(t float64) float64

// Common easing functions
var (
	EaseLinear = func(t float64) float64 { return t }
	EaseInOut  = func(t float64) float64 { return t * t * (3.0 - 2.0*t) }
	EaseIn     = func(t float64) float64 { return t * t }
	EaseOut    = func(t float64) float64 { return 1.0 - (1.0-t)*(1.0-t) }
)

// NewEnhancedDarkTheme creates an enhanced dark theme
func NewEnhancedDarkTheme() EnhancedTheme {
	base := DefaultDarkTheme()

	return EnhancedTheme{
		Theme: base,
		PrimaryGradient: []lipgloss.Color{
			lipgloss.Color("#89B4FA"),
			lipgloss.Color("#74C7EC"),
			lipgloss.Color("#89DCEB"),
		},
		SecondaryGradient: []lipgloss.Color{
			lipgloss.Color("#F5E0DC"),
			lipgloss.Color("#F2CDCD"),
			lipgloss.Color("#F5C2E7"),
		},
		AccentGradient: []lipgloss.Color{
			lipgloss.Color("#A6E3A1"),
			lipgloss.Color("#94E2D5"),
			lipgloss.Color("#89DCEB"),
		},
		PulseColors: []lipgloss.Color{
			lipgloss.Color("#89B4FA"),
			lipgloss.Color("#74C7EC"),
			lipgloss.Color("#89B4FA"),
		},
		ShimmerColors: []lipgloss.Color{
			lipgloss.Color("#313244"),
			lipgloss.Color("#45475A"),
			lipgloss.Color("#585B70"),
			lipgloss.Color("#45475A"),
			lipgloss.Color("#313244"),
		},
		GlowColor:     lipgloss.Color("#89B4FA"),
		ShadowColor:   lipgloss.Color("#11111B"),
		SupportsTrue:  true,
		Supports256:   true,
		SupportsBasic: true,
	}
}

// NewEnhancedLightTheme creates an enhanced light theme
func NewEnhancedLightTheme() EnhancedTheme {
	base := DefaultLightTheme()

	return EnhancedTheme{
		Theme: base,
		PrimaryGradient: []lipgloss.Color{
			lipgloss.Color("#1E66F5"),
			lipgloss.Color("#209FB5"),
			lipgloss.Color("#04A5E5"),
		},
		SecondaryGradient: []lipgloss.Color{
			lipgloss.Color("#DC8A78"),
			lipgloss.Color("#EA76CB"),
			lipgloss.Color("#8839EF"),
		},
		AccentGradient: []lipgloss.Color{
			lipgloss.Color("#40A02B"),
			lipgloss.Color("#209FB5"),
			lipgloss.Color("#1E66F5"),
		},
		PulseColors: []lipgloss.Color{
			lipgloss.Color("#1E66F5"),
			lipgloss.Color("#209FB5"),
			lipgloss.Color("#1E66F5"),
		},
		ShimmerColors: []lipgloss.Color{
			lipgloss.Color("#DCE0E8"),
			lipgloss.Color("#CCD0DA"),
			lipgloss.Color("#BCC0CC"),
			lipgloss.Color("#CCD0DA"),
			lipgloss.Color("#DCE0E8"),
		},
		GlowColor:     lipgloss.Color("#1E66F5"),
		ShadowColor:   lipgloss.Color("#E6E9EF"),
		SupportsTrue:  true,
		Supports256:   true,
		SupportsBasic: true,
	}
}

// NewEnhancedStyles creates enhanced styles from a theme
func NewEnhancedStyles(theme EnhancedTheme) EnhancedStyles {
	base := NewStyles(theme.Theme)

	styles := EnhancedStyles{
		Styles: base,
	}

	// Create gradient styles
	styles.HeaderGradient = lipgloss.NewStyle().
		Foreground(theme.PrimaryGradient[0]).
		Bold(true).
		Padding(0, 1)

	styles.AccentGradient = lipgloss.NewStyle().
		Foreground(theme.AccentGradient[0]).
		Bold(true)

	styles.ProgressGradient = lipgloss.NewStyle().
		Foreground(theme.AccentGradient[1])

	// Create animation styles
	styles.Pulse = lipgloss.NewStyle().
		Foreground(theme.PulseColors[0])

	styles.Shimmer = lipgloss.NewStyle().
		Foreground(theme.ShimmerColors[0])

	styles.Glow = lipgloss.NewStyle().
		Foreground(theme.GlowColor).
		Bold(true)

	// Create badge styles
	styles.BadgePrimary = lipgloss.NewStyle().
		Background(theme.AccentColor).
		Foreground(theme.Background).
		Padding(0, 1).
		Bold(true)

	styles.BadgeSecondary = lipgloss.NewStyle().
		Background(theme.DimmedText).
		Foreground(theme.Background).
		Padding(0, 1)

	styles.BadgeSuccess = lipgloss.NewStyle().
		Background(theme.SuccessColor).
		Foreground(theme.Background).
		Padding(0, 1).
		Bold(true)

	styles.BadgeWarning = lipgloss.NewStyle().
		Background(theme.WarningColor).
		Foreground(theme.Background).
		Padding(0, 1).
		Bold(true)

	styles.BadgeError = lipgloss.NewStyle().
		Background(theme.ErrorColor).
		Foreground(theme.Background).
		Padding(0, 1).
		Bold(true)

	styles.BadgeInfo = lipgloss.NewStyle().
		Background(theme.InfoColor).
		Foreground(theme.Background).
		Padding(0, 1).
		Bold(true)

	// Create panel styles
	styles.PanelPrimary = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.BorderColor)

	styles.PanelSecondary = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.DimmedText).
		Padding(1, 2).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BorderColor)

	styles.PanelBordered = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		Padding(1, 2).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.AccentColor)

	styles.PanelElevated = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		Padding(1, 2).
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(theme.BorderColor)

	// Create interactive styles
	styles.ButtonPrimary = lipgloss.NewStyle().
		Background(theme.AccentColor).
		Foreground(theme.Background).
		Padding(0, 2).
		Bold(true)

	styles.ButtonSecondary = lipgloss.NewStyle().
		Background(theme.DimmedText).
		Foreground(theme.Background).
		Padding(0, 2)

	styles.ButtonDisabled = lipgloss.NewStyle().
		Background(theme.DimmedText).
		Foreground(theme.Background).
		Padding(0, 2).
		Faint(true)

	styles.ButtonHover = lipgloss.NewStyle().
		Background(theme.HighlightText).
		Foreground(theme.Background).
		Padding(0, 2).
		Bold(true)

	styles.ButtonPressed = lipgloss.NewStyle().
		Background(theme.SelectedBg).
		Foreground(theme.SelectedFg).
		Padding(0, 2).
		Bold(true)

	// Create form styles
	styles.InputFocused = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Background).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.AccentColor).
		Padding(0, 1)

	styles.InputError = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Background).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.ErrorColor).
		Padding(0, 1)

	styles.InputSuccess = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Background(theme.Background).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.SuccessColor).
		Padding(0, 1)

	styles.Label = lipgloss.NewStyle().
		Foreground(theme.Foreground).
		Bold(true)

	styles.Placeholder = lipgloss.NewStyle().
		Foreground(theme.DimmedText).
		Italic(true)

	// Create layout styles
	styles.Container = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		Padding(1)

	styles.Sidebar = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BorderColor).
		BorderRight(true).
		Padding(1)

	styles.MainContent = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.Foreground).
		Padding(1)

	styles.Footer = lipgloss.NewStyle().
		Background(theme.Background).
		Foreground(theme.DimmedText).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.BorderColor).
		BorderTop(true).
		Padding(0, 1)

	return styles
}

// CreateGradientText creates text with a gradient effect
func CreateGradientText(text string, colors []lipgloss.Color) string {
	if len(colors) < 2 {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	var result strings.Builder

	for i, r := range runes {
		// Calculate position in gradient (0.0 to 1.0)
		pos := float64(i) / float64(len(runes)-1)

		// Find the two colors to interpolate between
		colorIndex := pos * float64(len(colors)-1)
		index1 := int(math.Floor(colorIndex))
		index2 := int(math.Ceil(colorIndex))

		if index1 == index2 {
			// Exact color match
			style := lipgloss.NewStyle().Foreground(colors[index1])
			result.WriteString(style.Render(string(r)))
		} else {
			// Interpolate between colors (simplified)
			style := lipgloss.NewStyle().Foreground(colors[index1])
			result.WriteString(style.Render(string(r)))
		}
	}

	return result.String()
}

// CreatePulseText creates pulsing text animation
func CreatePulseText(text string, colors []lipgloss.Color, frame int) string {
	if len(colors) == 0 {
		return text
	}

	colorIndex := frame % len(colors)
	style := lipgloss.NewStyle().Foreground(colors[colorIndex])
	return style.Render(text)
}

// CreateShimmerText creates shimmering text animation
func CreateShimmerText(text string, colors []lipgloss.Color, frame int) string {
	if len(colors) == 0 {
		return text
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	var result strings.Builder

	for i, r := range runes {
		colorIndex := (i + frame) % len(colors)
		style := lipgloss.NewStyle().Foreground(colors[colorIndex])
		result.WriteString(style.Render(string(r)))
	}

	return result.String()
}

// CreateProgressBar creates an enhanced progress bar
func CreateProgressBar(width int, progress float64, theme EnhancedTheme) string {
	if width <= 0 {
		return ""
	}

	filled := int(progress * float64(width))
	if filled > width {
		filled = width
	}

	var bar strings.Builder

	// Create gradient for filled portion
	for i := 0; i < filled; i++ {
		pos := float64(i) / float64(width)
		colorIndex := int(pos * float64(len(theme.AccentGradient)-1))
		if colorIndex >= len(theme.AccentGradient) {
			colorIndex = len(theme.AccentGradient) - 1
		}

		style := lipgloss.NewStyle().Foreground(theme.AccentGradient[colorIndex])
		bar.WriteString(style.Render("█"))
	}

	// Add empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(theme.DimmedText)
	for i := filled; i < width; i++ {
		bar.WriteString(emptyStyle.Render("░"))
	}

	return bar.String()
}

// DetectTerminalCapabilities detects terminal color capabilities
func DetectTerminalCapabilities() (trueColor, color256, basic bool) {
	// For now, assume basic capabilities
	// In a real implementation, you would detect the terminal's capabilities
	return true, true, true
}

// AdaptThemeToTerminal adapts a theme to terminal capabilities
func AdaptThemeToTerminal(theme EnhancedTheme) EnhancedTheme {
	trueColor, color256, basic := DetectTerminalCapabilities()

	theme.SupportsTrue = trueColor
	theme.Supports256 = color256
	theme.SupportsBasic = basic

	if !trueColor && !color256 {
		// Fallback to basic colors
		theme.AccentColor = lipgloss.Color("4")  // Blue
		theme.ErrorColor = lipgloss.Color("1")   // Red
		theme.SuccessColor = lipgloss.Color("2") // Green
		theme.WarningColor = lipgloss.Color("3") // Yellow
		theme.InfoColor = lipgloss.Color("6")    // Cyan
	}

	return theme
}
