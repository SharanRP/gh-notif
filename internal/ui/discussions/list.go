package discussions

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/SharanRP/gh-notif/internal/discussions"
	"github.com/SharanRP/gh-notif/internal/ui"
)

// DiscussionList provides an enhanced list view for discussions
type DiscussionList struct {
	list         list.Model
	discussions  []discussions.Discussion
	width        int
	height       int
	styles       EnhancedDiscussionStyles
	theme        ui.EnhancedTheme
	keyMap       DiscussionListKeyMap

	// State
	focused      bool
	loading      bool
	searchMode   bool
	searchQuery  string

	// Animation
	animationFrame int
	lastUpdate     time.Time
}

// DiscussionListKeyMap defines key bindings for the discussion list
type DiscussionListKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Search   key.Binding
	Filter   key.Binding
	Refresh  key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// DefaultDiscussionListKeyMap returns default key bindings for the list
func DefaultDiscussionListKeyMap() DiscussionListKeyMap {
	return DiscussionListKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Filter: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "filter"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "quit"),
		),
	}
}

// DiscussionItem represents a discussion item in the list
type DiscussionItem struct {
	Discussion discussions.Discussion
	Index      int
}

// FilterValue implements list.Item
func (d DiscussionItem) FilterValue() string {
	return d.Discussion.Title + " " + d.Discussion.Body + " " + d.Discussion.Author.Login
}

// NewDiscussionList creates a new enhanced discussion list
func NewDiscussionList(discussions []discussions.Discussion) *DiscussionList {
	theme := ui.NewEnhancedDarkTheme()
	styles := NewEnhancedDiscussionStyles(theme)

	// Create list items
	items := make([]list.Item, len(discussions))
	for i, discussion := range discussions {
		items[i] = DiscussionItem{
			Discussion: discussion,
			Index:      i,
		}
	}

	// Create the list model
	delegate := NewDiscussionItemDelegate(styles, theme)
	l := list.New(items, delegate, 0, 0)
	l.Title = "🗣️  GitHub Discussions"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.Title
	l.Styles.PaginationStyle = styles.Metadata
	l.Styles.HelpStyle = styles.Metadata

	return &DiscussionList{
		list:        l,
		discussions: discussions,
		styles:      styles,
		theme:       theme,
		keyMap:      DefaultDiscussionListKeyMap(),
		focused:     true,
		lastUpdate:  time.Now(),
	}
}

// DiscussionItemDelegate handles rendering of discussion items
type DiscussionItemDelegate struct {
	styles EnhancedDiscussionStyles
	theme  ui.EnhancedTheme
}

// NewDiscussionItemDelegate creates a new item delegate
func NewDiscussionItemDelegate(styles EnhancedDiscussionStyles, theme ui.EnhancedTheme) *DiscussionItemDelegate {
	return &DiscussionItemDelegate{
		styles: styles,
		theme:  theme,
	}
}

// Height returns the height of each item
func (d *DiscussionItemDelegate) Height() int {
	return 4
}

// Spacing returns the spacing between items
func (d *DiscussionItemDelegate) Spacing() int {
	return 1
}

// Update handles item updates
func (d *DiscussionItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders a discussion item
func (d *DiscussionItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	discussionItem, ok := item.(DiscussionItem)
	if !ok {
		return
	}

	discussion := discussionItem.Discussion
	isSelected := index == m.Index()

	// Status indicator
	statusIcon := "🟢"
	statusStyle := d.styles.StatusOpen
	if discussion.Answer != nil {
		statusIcon = "✅"
		statusStyle = d.styles.StatusAnswered
	}

	// Title with truncation
	title := discussion.Title
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = d.styles.HighlightText
	} else {
		titleStyle = d.styles.Title
	}

	// Repository and category info
	repoInfo := fmt.Sprintf("📁 %s", discussion.Repository.FullName)
	categoryInfo := fmt.Sprintf("%s %s", discussion.Category.Emoji, discussion.Category.Name)

	// Author and time
	authorInfo := fmt.Sprintf("👤 @%s • 🕒 %s",
		discussion.Author.Login,
		formatTimeAgo(discussion.CreatedAt))

	// Engagement metrics
	engagement := fmt.Sprintf("👍 %d • 💬 %d • 🔥 %d",
		discussion.UpvoteCount,
		discussion.CommentCount,
		discussion.ReactionCount)

	// Build the item content
	line1 := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusStyle.Render(statusIcon),
		" ",
		titleStyle.Render(title),
	)

	line2 := d.styles.Subtitle.Render(repoInfo + " • " + categoryInfo)
	line3 := d.styles.Metadata.Render(authorInfo)
	line4 := d.styles.ReactionCount.Render(engagement)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		line1,
		line2,
		line3,
		line4,
	)

	// Apply selection styling
	var finalContent string
	if isSelected {
		finalContent = d.styles.Selection.Render(content)
	} else {
		finalContent = d.styles.Panel.Render(content)
	}

	fmt.Fprint(w, finalContent)
}

// Init initializes the discussion list
func (dl *DiscussionList) Init() tea.Cmd {
	return nil
}

// Update handles messages for the discussion list
func (dl *DiscussionList) Update(msg tea.Msg) (*DiscussionList, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		dl.width = msg.Width
		dl.height = msg.Height
		dl.list.SetWidth(msg.Width - 4)
		dl.list.SetHeight(msg.Height - 4)

	case tea.KeyMsg:
		if !dl.focused {
			return dl, nil
		}

		switch {
		case key.Matches(msg, dl.keyMap.Quit):
			return dl, tea.Quit
		case key.Matches(msg, dl.keyMap.Enter):
			// Handle selection
			if item, ok := dl.list.SelectedItem().(DiscussionItem); ok {
				// Return a custom message or command to open the discussion
				return dl, tea.Sequence(
					tea.Printf("Opening discussion #%d", item.Discussion.Number),
				)
			}
		case key.Matches(msg, dl.keyMap.Refresh):
			dl.loading = true
			dl.animationFrame = 0
			dl.lastUpdate = time.Now()
			// Return refresh command
			return dl, tea.Batch(
				tea.Printf("Refreshing discussions..."),
				tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
					return tickMsg{}
				}),
			)
		}

	case tickMsg:
		dl.animationFrame++
		if dl.animationFrame > 10 {
			dl.animationFrame = 0
		}
		dl.lastUpdate = time.Now()
		return dl, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
			return tickMsg{}
		})
	}

	dl.list, cmd = dl.list.Update(msg)
	return dl, cmd
}

// View renders the discussion list
func (dl *DiscussionList) View() string {
	if dl.loading {
		loadingText := dl.styles.Glow.Render("🔄 Loading discussions...")
		return dl.styles.Container.Render(loadingText)
	}

	// Header with stats
	header := dl.renderHeader()

	// Main list
	listView := dl.list.View()

	// Footer with help
	footer := dl.renderFooter()

	// Combine all elements
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		listView,
		footer,
	)

	return dl.styles.App.Render(content)
}

// renderHeader renders the list header with statistics
func (dl *DiscussionList) renderHeader() string {
	totalCount := len(dl.discussions)
	answeredCount := 0
	for _, d := range dl.discussions {
		if d.Answer != nil {
			answeredCount++
		}
	}

	stats := fmt.Sprintf("📊 Total: %d • ✅ Answered: %d • ❓ Unanswered: %d",
		totalCount, answeredCount, totalCount-answeredCount)

	return dl.styles.Header.Render(stats)
}

// renderFooter renders the list footer with help
func (dl *DiscussionList) renderFooter() string {
	helpItems := []string{
		dl.styles.HighlightText.Render("↑↓") + dl.styles.Metadata.Render(" navigate"),
		dl.styles.HighlightText.Render("enter") + dl.styles.Metadata.Render(" select"),
		dl.styles.HighlightText.Render("/") + dl.styles.Metadata.Render(" search"),
		dl.styles.HighlightText.Render("r") + dl.styles.Metadata.Render(" refresh"),
		dl.styles.HighlightText.Render("q") + dl.styles.Metadata.Render(" quit"),
	}

	helpText := strings.Join(helpItems, dl.styles.Metadata.Render(" • "))

	return dl.styles.Footer.Render("🎮 " + helpText)
}

// SetDiscussions updates the discussions in the list
func (dl *DiscussionList) SetDiscussions(discussions []discussions.Discussion) {
	dl.discussions = discussions

	// Update list items
	items := make([]list.Item, len(discussions))
	for i, discussion := range discussions {
		items[i] = DiscussionItem{
			Discussion: discussion,
			Index:      i,
		}
	}

	dl.list.SetItems(items)
}

// GetSelectedDiscussion returns the currently selected discussion
func (dl *DiscussionList) GetSelectedDiscussion() *discussions.Discussion {
	if item, ok := dl.list.SelectedItem().(DiscussionItem); ok {
		return &item.Discussion
	}
	return nil
}

// Helper function for time formatting
func formatTimeAgo(t time.Time) string {
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
		return t.Format("Jan 2")
	}
}
