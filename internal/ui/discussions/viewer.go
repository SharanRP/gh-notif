package discussions

import (
	"fmt"
	"strings"
	"time"

	"github.com/SharanRP/gh-notif/internal/discussions"
	"github.com/SharanRP/gh-notif/internal/ui"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiscussionViewer provides a terminal UI for viewing discussions
type DiscussionViewer struct {
	discussion   *discussions.Discussion
	comments     []discussions.Comment
	viewport     viewport.Model
	width        int
	height       int
	styles       EnhancedDiscussionStyles
	theme        ui.EnhancedTheme
	showComments bool
	currentView  ViewMode
	keyMap       DiscussionKeyMap

	// Animation state
	animationFrame int
	lastUpdate     time.Time

	// Interactive state
	focused      bool
	selectedTab  int
	scrollOffset int
}

// ViewMode represents different viewing modes
type ViewMode int

const (
	ViewDiscussion ViewMode = iota
	ViewComments
	ViewAnalytics
	ViewMetadata
)

// tickMsg represents a tick message for animations
type tickMsg struct{}

// DiscussionKeyMap defines key bindings for the discussion viewer
type DiscussionKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Tab        key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Help       key.Binding
	Quit       key.Binding
	ToggleView key.Binding
	Refresh    key.Binding
}

// DefaultDiscussionKeyMap returns default key bindings
func DefaultDiscussionKeyMap() DiscussionKeyMap {
	return DiscussionKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "previous tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "next tab"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch view"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		ToggleView: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "toggle comments"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// EnhancedDiscussionStyles contains enhanced styling for the discussion viewer
type EnhancedDiscussionStyles struct {
	// Base styles
	App       lipgloss.Style
	Container lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style

	// Content styles
	Title         lipgloss.Style
	TitleGradient lipgloss.Style
	Subtitle      lipgloss.Style
	Author        lipgloss.Style
	AuthorBadge   lipgloss.Style
	Metadata      lipgloss.Style
	Body          lipgloss.Style
	BodyQuote     lipgloss.Style

	// Comment styles
	Comment       lipgloss.Style
	CommentHeader lipgloss.Style
	CommentBody   lipgloss.Style
	CommentMeta   lipgloss.Style
	CommentThread lipgloss.Style

	// Special elements
	Answer        lipgloss.Style
	AnswerBadge   lipgloss.Style
	Category      lipgloss.Style
	CategoryBadge lipgloss.Style
	Label         lipgloss.Style
	Reaction      lipgloss.Style
	ReactionCount lipgloss.Style

	// Interactive elements
	Tab         lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Button      lipgloss.Style
	ButtonHover lipgloss.Style

	// Layout styles
	Border        lipgloss.Style
	BorderActive  lipgloss.Style
	Panel         lipgloss.Style
	PanelElevated lipgloss.Style
	Separator     lipgloss.Style

	// Status styles
	StatusOpen     lipgloss.Style
	StatusClosed   lipgloss.Style
	StatusAnswered lipgloss.Style

	// Animation styles
	Pulse   lipgloss.Style
	Shimmer lipgloss.Style
	Glow    lipgloss.Style

	// Highlight styles
	Highlight     lipgloss.Style
	HighlightText lipgloss.Style
	Selection     lipgloss.Style
}

// NewDiscussionViewer creates a new discussion viewer
func NewDiscussionViewer(discussion *discussions.Discussion, comments []discussions.Comment) *DiscussionViewer {
	vp := viewport.New(80, 20)
	theme := ui.NewEnhancedDarkTheme()

	return &DiscussionViewer{
		discussion:   discussion,
		comments:     comments,
		viewport:     vp,
		showComments: true,
		currentView:  ViewDiscussion,
		styles:       NewEnhancedDiscussionStyles(theme),
		theme:        theme,
		keyMap:       DefaultDiscussionKeyMap(),
		focused:      true,
		selectedTab:  0,
		lastUpdate:   time.Now(),
	}
}

// NewEnhancedDiscussionStyles creates enhanced styles for discussions
func NewEnhancedDiscussionStyles(theme ui.EnhancedTheme) EnhancedDiscussionStyles {
	return EnhancedDiscussionStyles{
		// Base styles
		App: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Foreground).
			Padding(1),

		Container: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Foreground).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderColor).
			Padding(1, 2),

		Header: lipgloss.NewStyle().
			Background(theme.AccentColor).
			Foreground(theme.Background).
			Bold(true).
			Padding(0, 2).
			MarginBottom(1),

		Footer: lipgloss.NewStyle().
			Foreground(theme.DimmedText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(theme.BorderColor).
			Padding(1, 0, 0, 0),

		// Content styles with gradients
		Title: lipgloss.NewStyle().
			Foreground(theme.PrimaryGradient[0]).
			Bold(true).
			MarginBottom(1),

		TitleGradient: lipgloss.NewStyle().
			Foreground(theme.PrimaryGradient[1]).
			Bold(true).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(theme.AccentGradient[0]).
			Italic(true),

		Author: lipgloss.NewStyle().
			Foreground(theme.AccentGradient[1]).
			Bold(true),

		AuthorBadge: lipgloss.NewStyle().
			Background(theme.AccentGradient[0]).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		Metadata: lipgloss.NewStyle().
			Foreground(theme.DimmedText).
			Italic(true),

		Body: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			MarginTop(1).
			MarginBottom(1).
			Padding(1, 0),

		BodyQuote: lipgloss.NewStyle().
			Foreground(theme.DimmedText).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.AccentColor).
			PaddingLeft(2).
			Italic(true),

		// Enhanced comment styles
		Comment: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Foreground).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1),

		CommentHeader: lipgloss.NewStyle().
			Foreground(theme.AccentGradient[0]).
			Bold(true).
			MarginBottom(1),

		CommentBody: lipgloss.NewStyle().
			Foreground(theme.Foreground),

		CommentMeta: lipgloss.NewStyle().
			Foreground(theme.DimmedText).
			Italic(true).
			MarginTop(1),

		CommentThread: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.AccentGradient[2]).
			PaddingLeft(2),

		// Special elements with enhanced styling
		Answer: lipgloss.NewStyle().
			Background(theme.SuccessColor).
			Foreground(theme.Background).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.SuccessColor).
			Padding(1, 2).
			MarginTop(1).
			MarginBottom(1).
			Bold(true),

		AnswerBadge: lipgloss.NewStyle().
			Background(theme.SuccessColor).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		Category: lipgloss.NewStyle().
			Background(theme.InfoColor).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		CategoryBadge: lipgloss.NewStyle().
			Background(theme.AccentGradient[1]).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		Label: lipgloss.NewStyle().
			Background(theme.DimmedText).
			Foreground(theme.Background).
			Padding(0, 1),

		Reaction: lipgloss.NewStyle().
			Background(theme.AccentGradient[2]).
			Foreground(theme.Background).
			Padding(0, 1),

		ReactionCount: lipgloss.NewStyle().
			Foreground(theme.AccentGradient[1]).
			Bold(true),

		// Interactive elements
		Tab: lipgloss.NewStyle().
			Padding(0, 2).
			MarginRight(1),

		TabActive: lipgloss.NewStyle().
			Background(theme.AccentColor).
			Foreground(theme.Background).
			Padding(0, 2).
			MarginRight(1).
			Bold(true),

		TabInactive: lipgloss.NewStyle().
			Background(theme.DimmedText).
			Foreground(theme.Background).
			Padding(0, 2).
			MarginRight(1),

		Button: lipgloss.NewStyle().
			Background(theme.AccentColor).
			Foreground(theme.Background).
			Padding(0, 2).
			Bold(true),

		ButtonHover: lipgloss.NewStyle().
			Background(theme.HighlightText).
			Foreground(theme.Background).
			Padding(0, 2).
			Bold(true),

		// Layout styles
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.BorderColor),

		BorderActive: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.AccentColor),

		Panel: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Foreground).
			Border(lipgloss.NormalBorder()).
			BorderForeground(theme.BorderColor).
			Padding(1, 2),

		PanelElevated: lipgloss.NewStyle().
			Background(theme.Background).
			Foreground(theme.Foreground).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(theme.AccentColor).
			Padding(1, 2),

		Separator: lipgloss.NewStyle().
			Foreground(theme.BorderColor),

		// Status styles
		StatusOpen: lipgloss.NewStyle().
			Background(theme.SuccessColor).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		StatusClosed: lipgloss.NewStyle().
			Background(theme.ErrorColor).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		StatusAnswered: lipgloss.NewStyle().
			Background(theme.InfoColor).
			Foreground(theme.Background).
			Padding(0, 1).
			Bold(true),

		// Animation styles
		Pulse: lipgloss.NewStyle().
			Foreground(theme.PulseColors[0]),

		Shimmer: lipgloss.NewStyle().
			Foreground(theme.ShimmerColors[0]),

		Glow: lipgloss.NewStyle().
			Foreground(theme.GlowColor).
			Bold(true),

		// Highlight styles
		Highlight: lipgloss.NewStyle().
			Background(theme.AccentColor).
			Foreground(theme.Background),

		HighlightText: lipgloss.NewStyle().
			Foreground(theme.HighlightText).
			Bold(true),

		Selection: lipgloss.NewStyle().
			Background(theme.SelectedBg).
			Foreground(theme.SelectedFg),
	}
}

// Init initializes the discussion viewer
func (dv *DiscussionViewer) Init() tea.Cmd {
	return nil
}

// Update handles messages for the discussion viewer
func (dv *DiscussionViewer) Update(msg tea.Msg) (*DiscussionViewer, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dv.width = msg.Width
		dv.height = msg.Height
		dv.viewport.Width = msg.Width - 6
		dv.viewport.Height = msg.Height - 8
		dv.updateContent()

	case tea.KeyMsg:
		if !dv.focused {
			return dv, nil
		}

		switch {
		case key.Matches(msg, dv.keyMap.Quit):
			return dv, tea.Quit
		case key.Matches(msg, dv.keyMap.Escape):
			return dv, tea.Quit
		case key.Matches(msg, dv.keyMap.Left):
			dv.previousTab()
			dv.updateContent()
		case key.Matches(msg, dv.keyMap.Right):
			dv.nextTab()
			dv.updateContent()
		case key.Matches(msg, dv.keyMap.Tab):
			dv.nextTab()
			dv.updateContent()
		case key.Matches(msg, dv.keyMap.ToggleView):
			dv.showComments = !dv.showComments
			dv.updateContent()
		case key.Matches(msg, dv.keyMap.Refresh):
			// Trigger refresh animation
			dv.animationFrame = 0
			dv.lastUpdate = time.Now()
			dv.updateContent()
		}

	case tickMsg:
		// Update animation frame
		dv.animationFrame++
		if dv.animationFrame > 10 {
			dv.animationFrame = 0
		}
		dv.lastUpdate = time.Now()
		return dv, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg{}
		})
	}

	dv.viewport, cmd = dv.viewport.Update(msg)
	return dv, cmd
}

// nextTab switches to the next tab
func (dv *DiscussionViewer) nextTab() {
	dv.selectedTab = (dv.selectedTab + 1) % 4
	dv.currentView = ViewMode(dv.selectedTab)
}

// previousTab switches to the previous tab
func (dv *DiscussionViewer) previousTab() {
	dv.selectedTab = (dv.selectedTab - 1 + 4) % 4
	dv.currentView = ViewMode(dv.selectedTab)
}

// View renders the discussion viewer
func (dv *DiscussionViewer) View() string {
	if dv.discussion == nil {
		return dv.styles.Container.Render(
			dv.styles.Title.Render("🚫 No Discussion Available") + "\n\n" +
				dv.styles.Metadata.Render("No discussion data to display."),
		)
	}

	// Create the main layout
	header := dv.renderEnhancedHeader()
	tabs := dv.renderTabs()
	content := dv.styles.BorderActive.Render(dv.viewport.View())
	footer := dv.renderEnhancedFooter()

	// Combine all elements
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)

	return dv.styles.App.Render(mainContent)
}

// renderEnhancedHeader renders an enhanced header with beautiful styling
func (dv *DiscussionViewer) renderEnhancedHeader() string {
	if dv.discussion == nil {
		return ""
	}

	// Create animated title with gradient effect
	titleText := dv.discussion.Title
	if len(titleText) > 60 {
		titleText = titleText[:57] + "..."
	}

	// Add animation effect based on frame
	var titleStyle lipgloss.Style
	if dv.animationFrame%2 == 0 {
		titleStyle = dv.styles.Title
	} else {
		titleStyle = dv.styles.TitleGradient
	}

	title := titleStyle.Render("💬 " + titleText)

	// Discussion number with enhanced styling
	discussionNumber := dv.styles.AuthorBadge.Render(fmt.Sprintf("#%d", dv.discussion.Number))

	// Status badge with appropriate color
	var statusBadge string
	if dv.discussion.Answer != nil {
		statusBadge = dv.styles.StatusAnswered.Render("✅ ANSWERED")
	} else {
		statusBadge = dv.styles.StatusOpen.Render("🟢 OPEN")
	}

	// Repository info with enhanced styling
	repoInfo := dv.styles.Subtitle.Render(
		fmt.Sprintf("📁 %s", dv.discussion.Repository.FullName),
	)

	// Category with emoji and enhanced styling
	categoryInfo := dv.styles.CategoryBadge.Render(
		fmt.Sprintf("%s %s", dv.discussion.Category.Emoji, dv.discussion.Category.Name),
	)

	// Author information with badge
	authorInfo := lipgloss.JoinHorizontal(
		lipgloss.Left,
		dv.styles.Metadata.Render("by "),
		dv.styles.AuthorBadge.Render("@"+dv.discussion.Author.Login),
		dv.styles.Metadata.Render(" • "+dv.formatTimeAgo(dv.discussion.CreatedAt)),
	)

	// Engagement metrics with beautiful icons and styling
	engagementMetrics := lipgloss.JoinHorizontal(
		lipgloss.Left,
		dv.styles.ReactionCount.Render(fmt.Sprintf("👍 %d", dv.discussion.UpvoteCount)),
		dv.styles.Metadata.Render(" • "),
		dv.styles.ReactionCount.Render(fmt.Sprintf("💬 %d", dv.discussion.CommentCount)),
		dv.styles.Metadata.Render(" • "),
		dv.styles.ReactionCount.Render(fmt.Sprintf("🔥 %d", dv.discussion.ReactionCount)),
	)

	// Combine header elements
	headerLine1 := lipgloss.JoinHorizontal(lipgloss.Left, title, " ", discussionNumber, " ", statusBadge)
	headerLine2 := lipgloss.JoinHorizontal(lipgloss.Left, repoInfo, " • ", categoryInfo)
	headerLine3 := authorInfo
	headerLine4 := engagementMetrics

	headerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		headerLine1,
		headerLine2,
		headerLine3,
		headerLine4,
	)

	return dv.styles.Header.Render(headerContent)
}

// renderTabs renders the navigation tabs
func (dv *DiscussionViewer) renderTabs() string {
	tabs := []string{"💬 Discussion", "💭 Comments", "📊 Analytics", "📋 Metadata"}
	var renderedTabs []string

	for i, tab := range tabs {
		if i == dv.selectedTab {
			renderedTabs = append(renderedTabs, dv.styles.TabActive.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, dv.styles.TabInactive.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...)
}

// renderEnhancedFooter renders an enhanced footer with help text and status
func (dv *DiscussionViewer) renderEnhancedFooter() string {
	// Key bindings help
	helpItems := []string{
		dv.styles.HighlightText.Render("←→") + dv.styles.Metadata.Render(" navigate"),
		dv.styles.HighlightText.Render("t") + dv.styles.Metadata.Render(" toggle"),
		dv.styles.HighlightText.Render("r") + dv.styles.Metadata.Render(" refresh"),
		dv.styles.HighlightText.Render("q") + dv.styles.Metadata.Render(" quit"),
	}

	helpText := lipgloss.JoinHorizontal(
		lipgloss.Left,
		dv.styles.Metadata.Render("🎮 "),
		strings.Join(helpItems, dv.styles.Metadata.Render(" • ")),
	)

	// Status information
	statusText := dv.styles.Metadata.Render(
		fmt.Sprintf("📍 View: %s | 🕒 Updated: %s",
			dv.getViewName(),
			dv.formatTimeAgo(dv.lastUpdate)),
	)

	// Combine help and status
	footerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		helpText,
		statusText,
	)

	return dv.styles.Footer.Render(footerContent)
}

// getViewName returns the current view name
func (dv *DiscussionViewer) getViewName() string {
	switch dv.currentView {
	case ViewDiscussion:
		return "Discussion"
	case ViewComments:
		return "Comments"
	case ViewAnalytics:
		return "Analytics"
	case ViewMetadata:
		return "Metadata"
	default:
		return "Unknown"
	}
}

// updateContent updates the viewport content based on current view
func (dv *DiscussionViewer) updateContent() {
	var content string

	switch dv.currentView {
	case ViewDiscussion:
		content = dv.renderEnhancedDiscussionContent()
	case ViewComments:
		content = dv.renderEnhancedCommentsContent()
	case ViewAnalytics:
		content = dv.renderEnhancedAnalyticsContent()
	case ViewMetadata:
		content = dv.renderEnhancedMetadataContent()
	}

	dv.viewport.SetContent(content)
}

// renderEnhancedDiscussionContent renders the main discussion content with enhanced styling
func (dv *DiscussionViewer) renderEnhancedDiscussionContent() string {
	if dv.discussion == nil {
		return dv.styles.Panel.Render("No discussion content available")
	}

	var sections []string

	// Discussion body with enhanced formatting
	if dv.discussion.Body != "" {
		bodyContent := dv.renderMarkdownContent(dv.discussion.Body)
		bodyPanel := dv.styles.Panel.Render(bodyContent)
		sections = append(sections, bodyPanel)
	}

	// Answer section if present
	if dv.discussion.Answer != nil {
		answerSection := dv.renderAnswerSection()
		sections = append(sections, answerSection)
	}

	// Labels section if any
	if len(dv.discussion.Labels) > 0 {
		labelsSection := dv.renderLabelsSection()
		sections = append(sections, labelsSection)
	}

	// Quick comments preview if enabled
	if dv.showComments && len(dv.comments) > 0 {
		commentsPreview := dv.renderCommentsPreview()
		sections = append(sections, commentsPreview)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderAnswerSection renders the accepted answer with enhanced styling
func (dv *DiscussionViewer) renderAnswerSection() string {
	if dv.discussion.Answer == nil {
		return ""
	}

	answer := dv.discussion.Answer

	// Answer header with badge
	answerHeader := lipgloss.JoinHorizontal(
		lipgloss.Left,
		dv.styles.AnswerBadge.Render("✅ ACCEPTED ANSWER"),
		dv.styles.Metadata.Render(" by "),
		dv.styles.AuthorBadge.Render("@"+answer.Author.Login),
		dv.styles.Metadata.Render(" • "+dv.formatTimeAgo(answer.CreatedAt)),
	)

	// Answer body with enhanced formatting
	answerBody := dv.renderMarkdownContent(answer.Body)

	answerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		answerHeader,
		"",
		answerBody,
	)

	return dv.styles.Answer.Render(answerContent)
}

// renderLabelsSection renders discussion labels
func (dv *DiscussionViewer) renderLabelsSection() string {
	if len(dv.discussion.Labels) == 0 {
		return ""
	}

	var labels []string
	for _, label := range dv.discussion.Labels {
		labelStyle := dv.styles.Label.Copy()
		if label.Color != "" {
			labelStyle = labelStyle.Background(lipgloss.Color("#" + label.Color))
		}
		labels = append(labels, labelStyle.Render(label.Name))
	}

	labelsContent := lipgloss.JoinHorizontal(lipgloss.Left, labels...)

	return dv.styles.Panel.Render(
		dv.styles.Subtitle.Render("🏷️ Labels") + "\n" + labelsContent,
	)
}

// renderCommentsPreview renders a preview of comments
func (dv *DiscussionViewer) renderCommentsPreview() string {
	if len(dv.comments) == 0 {
		return ""
	}

	previewCount := 3
	if len(dv.comments) < previewCount {
		previewCount = len(dv.comments)
	}

	var commentPreviews []string
	for i := 0; i < previewCount; i++ {
		comment := dv.comments[i]
		preview := dv.renderCommentPreview(comment, i+1)
		commentPreviews = append(commentPreviews, preview)
	}

	moreText := ""
	if len(dv.comments) > previewCount {
		moreText = dv.styles.Metadata.Render(
			fmt.Sprintf("\n... and %d more comments (switch to Comments tab to see all)",
				len(dv.comments)-previewCount),
		)
	}

	previewContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render(fmt.Sprintf("💭 Comments Preview (%d total)", len(dv.comments))),
		"",
		lipgloss.JoinVertical(lipgloss.Left, commentPreviews...),
		moreText,
	)

	return dv.styles.Panel.Render(previewContent)
}

// renderEnhancedCommentsContent renders the full comments view
func (dv *DiscussionViewer) renderEnhancedCommentsContent() string {
	if len(dv.comments) == 0 {
		return dv.styles.Panel.Render(
			dv.styles.Subtitle.Render("💭 No Comments Yet") + "\n\n" +
				dv.styles.Metadata.Render("Be the first to comment on this discussion!"),
		)
	}

	var commentSections []string

	// Comments header
	header := dv.styles.Subtitle.Render(
		fmt.Sprintf("💭 Comments (%d)", len(dv.comments)),
	)
	commentSections = append(commentSections, header)

	// Render all comments
	for i, comment := range dv.comments {
		commentSection := dv.renderEnhancedComment(comment, i+1)
		commentSections = append(commentSections, commentSection)
	}

	return lipgloss.JoinVertical(lipgloss.Left, commentSections...)
}

// renderEnhancedComment renders a single comment with enhanced styling
func (dv *DiscussionViewer) renderEnhancedComment(comment discussions.Comment, index int) string {
	// Comment header with author and metadata
	commentHeader := lipgloss.JoinHorizontal(
		lipgloss.Left,
		dv.styles.CommentHeader.Render(fmt.Sprintf("#%d", index)),
		dv.styles.Metadata.Render(" by "),
		dv.styles.AuthorBadge.Render("@"+comment.Author.Login),
		dv.styles.Metadata.Render(" • "+dv.formatTimeAgo(comment.CreatedAt)),
	)

	// Special badge for answer
	if comment.IsAnswer {
		answerBadge := dv.styles.AnswerBadge.Render("✅ ANSWER")
		commentHeader = lipgloss.JoinHorizontal(lipgloss.Left, commentHeader, " ", answerBadge)
	}

	// Comment body
	commentBody := dv.renderMarkdownContent(comment.Body)

	// Comment metadata (reactions, etc.)
	var metaParts []string
	if comment.UpvoteCount > 0 {
		metaParts = append(metaParts, fmt.Sprintf("👍 %d", comment.UpvoteCount))
	}
	if comment.ReactionCount > 0 {
		metaParts = append(metaParts, fmt.Sprintf("🔥 %d", comment.ReactionCount))
	}

	var commentMeta string
	if len(metaParts) > 0 {
		commentMeta = dv.styles.CommentMeta.Render(strings.Join(metaParts, " • "))
	}

	// Combine comment elements
	commentContent := lipgloss.JoinVertical(
		lipgloss.Left,
		commentHeader,
		"",
		commentBody,
	)

	if commentMeta != "" {
		commentContent = lipgloss.JoinVertical(lipgloss.Left, commentContent, "", commentMeta)
	}

	return dv.styles.Comment.Render(commentContent)
}

// renderCommentPreview renders a preview of a comment
func (dv *DiscussionViewer) renderCommentPreview(comment discussions.Comment, index int) string {
	// Shorter preview version
	preview := comment.Body
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}

	commentPreview := fmt.Sprintf("#%d @%s: %s",
		index, comment.Author.Login, preview)

	return dv.styles.CommentMeta.Render(commentPreview)
}

// renderEnhancedAnalyticsContent renders analytics view
func (dv *DiscussionViewer) renderEnhancedAnalyticsContent() string {
	if dv.discussion == nil {
		return dv.styles.Panel.Render("No analytics available")
	}

	var sections []string

	// Basic metrics section
	metricsSection := dv.renderMetricsSection()
	sections = append(sections, metricsSection)

	// Engagement analysis
	engagementSection := dv.renderEngagementSection()
	sections = append(sections, engagementSection)

	// Timeline section
	timelineSection := dv.renderTimelineSection()
	sections = append(sections, timelineSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderMetricsSection renders the metrics section
func (dv *DiscussionViewer) renderMetricsSection() string {
	metrics := []string{
		fmt.Sprintf("👍 Upvotes: %d", dv.discussion.UpvoteCount),
		fmt.Sprintf("💬 Comments: %d", dv.discussion.CommentCount),
		fmt.Sprintf("🔥 Reactions: %d", dv.discussion.ReactionCount),
		"👁️ Views: N/A", // Not available in current API
	}

	metricsContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("📊 Engagement Metrics"),
		"",
		strings.Join(metrics, "\n"),
	)

	return dv.styles.Panel.Render(metricsContent)
}

// renderEngagementSection renders engagement analysis
func (dv *DiscussionViewer) renderEngagementSection() string {
	// Calculate engagement score
	engagementScore := float64(dv.discussion.UpvoteCount)*0.3 +
		float64(dv.discussion.CommentCount)*0.5 +
		float64(dv.discussion.ReactionCount)*0.2

	var engagementLevel string
	var engagementColor lipgloss.Style

	switch {
	case engagementScore >= 50:
		engagementLevel = "🔥 Very High"
		engagementColor = dv.styles.StatusAnswered
	case engagementScore >= 20:
		engagementLevel = "📈 High"
		engagementColor = dv.styles.StatusOpen
	case engagementScore >= 5:
		engagementLevel = "📊 Medium"
		engagementColor = dv.styles.CategoryBadge
	default:
		engagementLevel = "📉 Low"
		engagementColor = dv.styles.Metadata
	}

	analysisContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("🎯 Engagement Analysis"),
		"",
		fmt.Sprintf("Score: %.1f", engagementScore),
		engagementColor.Render("Level: "+engagementLevel),
	)

	return dv.styles.Panel.Render(analysisContent)
}

// renderTimelineSection renders timeline information
func (dv *DiscussionViewer) renderTimelineSection() string {
	timelineItems := []string{
		fmt.Sprintf("📅 Created: %s", dv.formatTimeAgo(dv.discussion.CreatedAt)),
		fmt.Sprintf("🔄 Updated: %s", dv.formatTimeAgo(dv.discussion.UpdatedAt)),
	}

	if dv.discussion.Answer != nil {
		timelineItems = append(timelineItems,
			fmt.Sprintf("✅ Answered: %s", dv.formatTimeAgo(dv.discussion.Answer.CreatedAt)))
	}

	timelineContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("⏰ Timeline"),
		"",
		strings.Join(timelineItems, "\n"),
	)

	return dv.styles.Panel.Render(timelineContent)
}

// renderEnhancedMetadataContent renders metadata view
func (dv *DiscussionViewer) renderEnhancedMetadataContent() string {
	if dv.discussion == nil {
		return dv.styles.Panel.Render("No metadata available")
	}

	var sections []string

	// Repository information
	repoSection := dv.renderRepositorySection()
	sections = append(sections, repoSection)

	// Category information
	categorySection := dv.renderCategorySection()
	sections = append(sections, categorySection)

	// Author information
	authorSection := dv.renderAuthorSection()
	sections = append(sections, authorSection)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderRepositorySection renders repository metadata
func (dv *DiscussionViewer) renderRepositorySection() string {
	repo := dv.discussion.Repository

	repoInfo := []string{
		fmt.Sprintf("Name: %s", repo.Name),
		fmt.Sprintf("Full Name: %s", repo.FullName),
		fmt.Sprintf("Owner: @%s", repo.Owner.Login),
		fmt.Sprintf("Private: %t", repo.Private),
		fmt.Sprintf("URL: %s", repo.URL),
	}

	repoContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("📁 Repository"),
		"",
		strings.Join(repoInfo, "\n"),
	)

	return dv.styles.Panel.Render(repoContent)
}

// renderCategorySection renders category metadata
func (dv *DiscussionViewer) renderCategorySection() string {
	category := dv.discussion.Category

	categoryInfo := []string{
		fmt.Sprintf("Name: %s %s", category.Emoji, category.Name),
		fmt.Sprintf("Description: %s", category.Description),
		fmt.Sprintf("Answerable: %t", category.IsAnswerable),
		fmt.Sprintf("Created: %s", dv.formatTimeAgo(category.CreatedAt)),
	}

	categoryContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("📂 Category"),
		"",
		strings.Join(categoryInfo, "\n"),
	)

	return dv.styles.Panel.Render(categoryContent)
}

// renderAuthorSection renders author metadata
func (dv *DiscussionViewer) renderAuthorSection() string {
	author := dv.discussion.Author

	authorInfo := []string{
		fmt.Sprintf("Username: @%s", author.Login),
		fmt.Sprintf("Name: %s", author.Name),
		fmt.Sprintf("Profile: %s", author.URL),
	}

	authorContent := lipgloss.JoinVertical(
		lipgloss.Left,
		dv.styles.Subtitle.Render("👤 Author"),
		"",
		strings.Join(authorInfo, "\n"),
	)

	return dv.styles.Panel.Render(authorContent)
}

// renderMarkdownContent renders markdown content with basic formatting
func (dv *DiscussionViewer) renderMarkdownContent(content string) string {
	// Basic markdown rendering - could be enhanced with a proper markdown renderer
	lines := strings.Split(content, "\n")
	var renderedLines []string

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			renderedLines = append(renderedLines, dv.styles.BodyQuote.Render(line))
			continue
		}

		// Handle quotes
		if strings.HasPrefix(line, ">") {
			renderedLines = append(renderedLines, dv.styles.BodyQuote.Render(line))
			continue
		}

		// Handle headers
		if strings.HasPrefix(line, "#") {
			renderedLines = append(renderedLines, dv.styles.HighlightText.Render(line))
			continue
		}

		// Regular text
		renderedLines = append(renderedLines, dv.styles.Body.Render(line))
	}

	return strings.Join(renderedLines, "\n")
}

// formatTimeAgo formats a time for display (reuse from existing code)
func (dv *DiscussionViewer) formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	} else {
		return t.Format("Jan 2, 2006")
	}
}

// renderCommentsSection renders the comments section
func (dv *DiscussionViewer) renderCommentsSection() string {
	var parts []string

	header := dv.styles.Highlight.Render(fmt.Sprintf("💬 Comments (%d)", len(dv.comments)))
	parts = append(parts, header)

	for i, comment := range dv.comments {
		commentContent := dv.renderComment(comment, i)
		parts = append(parts, commentContent)
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderComment renders a single comment
func (dv *DiscussionViewer) renderComment(comment discussions.Comment, index int) string {
	// Comment header
	header := fmt.Sprintf("#%d @%s • %s",
		index+1,
		comment.Author.Login,
		dv.formatTime(comment.CreatedAt))

	// Comment body
	body := comment.Body
	if comment.IsAnswer {
		body = "✅ " + body
	}

	// Reactions if any
	var reactions string
	if comment.UpvoteCount > 0 {
		reactions = fmt.Sprintf("👍 %d", comment.UpvoteCount)
	}

	commentContent := lipgloss.JoinVertical(lipgloss.Left,
		dv.styles.CommentMeta.Render(header),
		body,
		dv.styles.Metadata.Render(reactions))

	return dv.styles.Comment.Render(commentContent)
}

// renderAnalyticsContent renders analytics view
func (dv *DiscussionViewer) renderAnalyticsContent() string {
	if dv.discussion == nil {
		return "No analytics available"
	}

	var parts []string

	// Basic metrics
	parts = append(parts, dv.styles.Highlight.Render("📊 Discussion Analytics"))

	metrics := []string{
		fmt.Sprintf("Upvotes: %d", dv.discussion.UpvoteCount),
		fmt.Sprintf("Comments: %d", dv.discussion.CommentCount),
		fmt.Sprintf("Reactions: %d", dv.discussion.ReactionCount),
		fmt.Sprintf("Created: %s", dv.formatTime(dv.discussion.CreatedAt)),
		fmt.Sprintf("Updated: %s", dv.formatTime(dv.discussion.UpdatedAt)),
	}

	for _, metric := range metrics {
		parts = append(parts, dv.styles.Metadata.Render("• "+metric))
	}

	// Category info
	parts = append(parts, "")
	parts = append(parts, dv.styles.Highlight.Render("📂 Category"))
	categoryInfo := fmt.Sprintf("%s %s - %s",
		dv.discussion.Category.Emoji,
		dv.discussion.Category.Name,
		dv.discussion.Category.Description)
	parts = append(parts, dv.styles.Metadata.Render(categoryInfo))

	// Participation info
	if dv.discussion.ViewerDidAuthor {
		parts = append(parts, "")
		parts = append(parts, dv.styles.Highlight.Render("👤 You are the author"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// formatTime formats a time for display
func (dv *DiscussionViewer) formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	} else {
		return t.Format("Jan 2, 2006")
	}
}
