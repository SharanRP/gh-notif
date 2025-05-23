# Enhanced Terminal UI

gh-notif features an exceptional terminal experience built with the Charm suite of Go libraries (Bubble Tea, Lip Gloss, Bubbles). The UI is designed to be modern, responsive, and delightful to use while maintaining excellent performance.

## Architecture Overview

### Component System

The UI is built around a reusable component architecture:

- **Component Registry**: Manages and creates UI components
- **Virtual Lists**: High-performance lists for thousands of notifications
- **Interactive Forms**: Forms with validation and keyboard navigation
- **Progress Indicators**: Rich progress feedback with animations
- **Markdown Rendering**: Full markdown support with syntax highlighting
- **Layout System**: Flexible layouts (vertical, horizontal, grid, split, tabs)

### Key Features

#### 🚀 High Performance
- **Virtualized Lists**: Handle thousands of notifications smoothly
- **Render Caching**: Intelligent caching for optimal performance
- **Concurrent Operations**: Background processing with progress feedback
- **Memory Efficient**: Optimized memory usage with object pooling

#### 🎨 Rich Styling & Theming
- **Multiple Themes**: Dark, light, and high contrast themes
- **Gradient Text**: Beautiful gradient effects for headers and branding
- **Adaptive Colors**: Automatically adapts to terminal capabilities
- **Custom Styling**: Badges, panels, and custom styled components
- **Animations**: Subtle animations for state changes and feedback

#### ⌨️ Exceptional Keyboard Support
- **Vim-style Navigation**: j/k for up/down, g/G for top/bottom
- **Tab Navigation**: Navigate between form fields and components
- **Custom Key Bindings**: Configurable keyboard shortcuts
- **Context-aware Help**: Dynamic help based on current context

#### ♿ Accessibility Features
- **Screen Reader Support**: Compatible with screen readers
- **High Contrast Mode**: Enhanced visibility for low vision users
- **Keyboard-only Navigation**: Full functionality without mouse
- **Configurable Text Size**: Adjustable text sizing
- **Color Blind Friendly**: Alternative color schemes for color blindness

## Component Reference

### Virtual List

High-performance virtualized list component for handling large datasets:

```go
// Create notification items
items := components.NewNotificationItemList(notifications, styles, symbols)

// Create virtual list
virtualList := components.NewVirtualList(items.GetVirtualListItems(), 1)
virtualList.SetSize(width, height)
virtualList.SetFocused(true)

// Handle selection
virtualList.Update(msg) // Returns ComponentEvent on selection
```

**Features:**
- Renders only visible items for performance
- Smooth scrolling with keyboard navigation
- Render caching for frequently accessed items
- Support for variable item heights
- Selection and focus management

### Interactive Forms

Forms with validation, keyboard navigation, and rich feedback:

```go
// Create form
form := components.NewForm("Filter Notifications")

// Add fields with validation
form.AddField(
    components.NewTextInputField("query", "Search").
        SetPlaceholder("Enter search terms...").
        SetValidator(func(value string) error {
            if len(value) < 3 {
                return fmt.Errorf("minimum 3 characters required")
            }
            return nil
        }).
        SetRequired(true),
)

// Handle submission
form.Update(msg) // Returns ComponentEvent on submit
```

**Features:**
- Tab navigation between fields
- Real-time validation with error display
- Required field indicators
- Help text and placeholders
- Submit/cancel handling

### Progress Indicators

Rich progress feedback with multiple styles:

```go
// Progress bar
progressBar := components.NewProgress(components.ProgressBar)
progressBar.SetTitle("Loading notifications...")
progressBar.SetValue(0.65) // 65% complete

// Step-based progress
progressSteps := components.NewProgress(components.ProgressSteps)
progressSteps.AddStep(components.ProgressStep{
    ID:     "fetch",
    Title:  "Fetching notifications",
    Status: components.StepInProgress,
})
```

**Progress Types:**
- **Progress Bar**: Traditional progress bar with percentage
- **Spinner**: Animated spinner with text
- **Circular**: Circular progress indicator
- **Steps**: Step-by-step progress with status

### Markdown Rendering

Full markdown support with syntax highlighting:

```go
content := `# GitHub Notifications

## Features
- **High Performance**: Virtualized lists
- **Beautiful UI**: Modern terminal interface

## Code Example
` + "```go" + `
client, err := github.NewClient(ctx)
` + "```" + `
`

markdown := components.NewMarkdownRenderer(content)
markdown.SetSize(width, height)
```

**Supported Elements:**
- Headers (H1-H6) with styling
- **Bold** and *italic* text
- `Inline code` and code blocks
- [Links](url) with highlighting
- Lists (bulleted and numbered)
- Blockquotes and horizontal rules
- Tables with borders

### Layout System

Flexible layout system for organizing components:

```go
// Vertical layout
layout := components.NewLayout(components.LayoutVertical)
layout.AddComponent("header", headerComponent)
layout.AddComponent("content", contentComponent)
layout.AddComponent("footer", footerComponent)

// Split layout with ratios
splitLayout := components.NewLayout(components.LayoutSplit)
splitLayout.SetSplits([]int{1, 2}) // 1:2 ratio
splitLayout.AddComponent("sidebar", sidebarComponent)
splitLayout.AddComponent("main", mainComponent)

// Grid layout
gridLayout := components.NewLayout(components.LayoutGrid)
// Automatically arranges components in a grid
```

**Layout Types:**
- **Vertical**: Stack components vertically
- **Horizontal**: Arrange components side by side
- **Grid**: Automatic grid arrangement
- **Split**: Resizable splits with custom ratios
- **Tabs**: Tabbed interface with navigation

## Styling System

### Enhanced Themes

```go
// Create enhanced theme
theme := ui.NewEnhancedDarkTheme()
theme = ui.AdaptThemeToTerminal(theme) // Adapt to terminal capabilities
styles := ui.NewEnhancedStyles(theme)

// Apply accessibility settings
accessibilitySettings := ui.DefaultEnhancedAccessibilitySettings()
accessibilitySettings.HighContrastMode = true
theme = ui.ApplyAccessibilitySettings(theme, accessibilitySettings)
```

### Gradient Text

```go
// Create gradient text
headerText := ui.CreateGradientText("GitHub Notifications", theme.PrimaryGradient)

// Animated text effects
pulseText := ui.CreatePulseText("Loading...", theme.PulseColors, animFrame)
shimmerText := ui.CreateShimmerText("Processing", theme.ShimmerColors, animFrame)
```

### Badges and Panels

```go
// Create badges
primaryBadge := components.NewBadge("PRIMARY", components.BadgePrimary)
successBadge := components.NewBadge("SUCCESS", components.BadgeSuccess)
errorBadge := components.NewBadge("ERROR", components.BadgeError)

// Create panels
panel := components.NewPanel("Notifications", components.PanelBordered)
panel.SetContent("Panel content here...")
```

## Accessibility

### Screen Reader Support

```go
// Enable screen reader mode
settings := ui.DefaultEnhancedAccessibilitySettings()
settings.ScreenReaderMode = true
settings.VerboseDescriptions = true

// Create accessible descriptions
description := ui.CreateAccessibleDescription(notification, settings)
// "Unread notification from repository microsoft/vscode. Type: PullRequest..."
```

### Keyboard Navigation

The UI supports multiple navigation styles:

- **Arrow Keys**: Standard arrow key navigation
- **Vim Keys**: j/k for up/down, h/l for left/right, g/G for top/bottom
- **Tab Navigation**: Tab between form fields and components
- **Custom Bindings**: User-configurable key bindings

### Visual Accessibility

```go
// High contrast theme
theme := ui.HighContrastEnhancedTheme()

// Color blind friendly theme
theme = ui.ColorBlindFriendlyTheme(theme)

// Reduced motion
settings.ReducedMotion = true // Disables animations
```

## Performance Optimizations

### Virtual Lists
- Only renders visible items
- Intelligent render caching
- Smooth scrolling with large datasets
- Memory efficient item management

### Render Caching
- Caches rendered components by content hash
- Automatic cache invalidation on changes
- Configurable cache size limits
- Memory pool for frequent allocations

### Concurrent Operations
- Background processing with progress feedback
- Non-blocking UI updates
- Efficient goroutine management
- Proper cancellation handling

## Usage Examples

### Basic Enhanced UI

```go
// Create enhanced model
model := ui.NewEnhancedModel(notifications)

// Run with enhanced features
p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
_, err := p.Run()
```

### Custom Component

```go
// Create custom component
config := components.ComponentConfig{
    Type:   components.VirtualListComponentType,
    Width:  80,
    Height: 24,
    Props: map[string]interface{}{
        "items":      notificationItems,
        "itemHeight": 1,
    },
}

component, err := registry.CreateComponent(config)
```

### Demo Application

Run the comprehensive demo to see all features:

```go
import "github.com/user/gh-notif/internal/ui/examples"

// Run the enhanced UI demo
err := examples.RunEnhancedUIDemo()
```

The demo showcases:
- Virtual list with 1000+ items
- Interactive forms with validation
- Progress indicators with animations
- Markdown rendering with syntax highlighting
- Layout system demonstrations
- Enhanced styling effects

## Best Practices

### Performance
- Use virtual lists for large datasets (>100 items)
- Enable render caching for frequently updated components
- Implement proper cancellation for long-running operations
- Use object pooling for frequently allocated objects

### Accessibility
- Always provide keyboard navigation
- Include screen reader descriptions
- Support high contrast mode
- Test with accessibility tools

### Styling
- Use semantic color names (success, error, warning)
- Adapt to terminal capabilities
- Provide theme alternatives
- Test in different terminal environments

### User Experience
- Provide clear visual feedback for actions
- Use progressive disclosure for complex features
- Include contextual help
- Maintain consistent navigation patterns

The enhanced UI system provides a solid foundation for building modern, accessible, and performant terminal applications while maintaining the flexibility to customize and extend as needed.
