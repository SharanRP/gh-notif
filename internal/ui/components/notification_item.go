package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// NotificationItem implements VirtualListItem for GitHub notifications
type NotificationItem struct {
	notification *github.Notification
	styles       EnhancedStyles
	symbols      Symbols
	width        int
	compact      bool
}

// NewNotificationItem creates a new notification item
func NewNotificationItem(notification *github.Notification, styles EnhancedStyles, symbols Symbols) *NotificationItem {
	return &NotificationItem{
		notification: notification,
		styles:       styles,
		symbols:      symbols,
		compact:      true,
	}
}

// Render renders the notification item with the given width and style
func (ni *NotificationItem) Render(width int, style lipgloss.Style) string {
	ni.width = width

	if ni.compact {
		return ni.renderCompact(style)
	}
	return ni.renderDetailed(style)
}

// renderCompact renders a compact view of the notification
func (ni *NotificationItem) renderCompact(style lipgloss.Style) string {
	n := ni.notification

	// Status indicator with enhanced styling
	var indicator string
	if n.GetUnread() {
		indicator = ni.styles.UnreadIndicator.Render(ni.symbols.UnreadIndicator + " NEW")
	} else {
		indicator = ni.styles.ReadIndicator.Render(ni.symbols.ReadIndicator + " READ")
	}

	// Type icon with enhanced badges
	var typeIcon string
	var typeText string
	var typeStyle lipgloss.Style

	switch n.GetSubject().GetType() {
	case "PullRequest":
		typeIcon = ni.symbols.PullRequest
		typeText = "PR"
		typeStyle = ni.styles.BadgeInfo
	case "Issue":
		typeIcon = ni.symbols.Issue
		typeText = "ISSUE"
		typeStyle = ni.styles.BadgeWarning
	case "Release":
		typeIcon = ni.symbols.Release
		typeText = "RELEASE"
		typeStyle = ni.styles.BadgeSuccess
	case "Discussion":
		typeIcon = ni.symbols.Discussion
		typeText = "DISCUSS"
		typeStyle = ni.styles.BadgePrimary
	case "Commit":
		typeIcon = ni.symbols.Commit
		typeText = "COMMIT"
		typeStyle = ni.styles.BadgeSecondary
	default:
		typeIcon = ni.symbols.Dot
		typeText = "OTHER"
		typeStyle = ni.styles.BadgeSecondary
	}

	typeIndicator := typeStyle.Render(typeIcon + " " + typeText)

	// Repository name with styling
	repoName := n.GetRepository().GetFullName()
	repoStyle := ni.styles.AccentGradient.Copy()
	if len(repoName) > 30 {
		repoName = repoName[:27] + "..."
	}
	repoText := repoStyle.Render(repoName)

	// Title with truncation
	title := n.GetSubject().GetTitle()
	maxTitleWidth := ni.width - 50 // Reserve space for other elements
	if maxTitleWidth < 20 {
		maxTitleWidth = 20
	}
	if len(title) > maxTitleWidth {
		title = title[:maxTitleWidth-3] + "..."
	}

	// Time formatting with relative time
	updatedAt := n.GetUpdatedAt().Time
	timeStr := ni.formatRelativeTime(updatedAt)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	timeText := timeStyle.Render(timeStr)

	// Reason badge with icons
	reason := n.GetReason()
	var reasonBadge string
	switch reason {
	case "assign":
		reasonBadge = ni.styles.BadgeInfo.Render("👤 ASSIGNED")
	case "author":
		reasonBadge = ni.styles.BadgePrimary.Render("✍️ AUTHOR")
	case "comment":
		reasonBadge = ni.styles.BadgeSecondary.Render("💬 COMMENT")
	case "mention":
		reasonBadge = ni.styles.BadgeWarning.Render("📢 MENTION")
	case "review_requested":
		reasonBadge = ni.styles.BadgeError.Render("👀 REVIEW")
	case "subscribed":
		reasonBadge = ni.styles.BadgeSuccess.Render("🔔 SUBSCRIBED")
	default:
		reasonBadge = ni.styles.BadgeSecondary.Render("📌 " + strings.ToUpper(reason))
	}

	// Build the line
	parts := []string{
		indicator,
		typeIndicator,
		repoText,
		title,
	}

	// Add reason badge if there's space
	if ni.width > 80 {
		parts = append(parts, reasonBadge)
	}

	// Add time if there's space
	if ni.width > 100 {
		parts = append(parts, timeText)
	}

	line := lipgloss.JoinHorizontal(lipgloss.Center, parts...)

	// Apply the provided style (for selection, etc.)
	return style.Render(line)
}

// renderDetailed renders a detailed view of the notification
func (ni *NotificationItem) renderDetailed(style lipgloss.Style) string {
	n := ni.notification
	var parts []string

	// Header line with status and type
	var statusIcon string
	if n.GetUnread() {
		statusIcon = ni.styles.BadgeError.Render("● UNREAD")
	} else {
		statusIcon = ni.styles.BadgeSuccess.Render("○ READ")
	}

	typeText := ni.getTypeText(n.GetSubject().GetType())
	typeBadge := ni.styles.BadgeInfo.Render(typeText)

	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, statusIcon, " ", typeBadge)
	parts = append(parts, headerLine)

	// Repository and title
	repoStyle := ni.styles.HeaderGradient
	repoLine := repoStyle.Render(n.GetRepository().GetFullName())
	parts = append(parts, repoLine)

	titleStyle := ni.styles.Header
	titleLine := titleStyle.Render(n.GetSubject().GetTitle())
	parts = append(parts, titleLine)

	// Metadata
	updatedAt := n.GetUpdatedAt().Time
	timeStr := ni.formatRelativeTime(updatedAt)
	reason := n.GetReason()

	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	metaLine := metaStyle.Render(fmt.Sprintf("Updated %s • Reason: %s", timeStr, reason))
	parts = append(parts, metaLine)

	// URL if available
	if url := n.GetSubject().GetURL(); url != "" {
		urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Underline(true)
		urlLine := urlStyle.Render("🔗 " + url)
		parts = append(parts, urlLine)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Apply enhanced container styling with rounded borders
	containerStyle := ni.styles.PanelElevated.Copy().
		Width(ni.width-8).
		Padding(2, 3).
		Margin(1, 2)

	// Apply the provided style (for selection, etc.)
	// Note: We'll apply the style by overlaying it
	if style.GetBackground() != lipgloss.Color("") {
		containerStyle = containerStyle.Background(style.GetBackground())
	}
	if style.GetForeground() != lipgloss.Color("") {
		containerStyle = containerStyle.Foreground(style.GetForeground())
	}

	return containerStyle.Render(content)
}

// GetHeight returns the height of the item when rendered
func (ni *NotificationItem) GetHeight() int {
	if ni.compact {
		return 1
	}
	return 6 // Detailed view height
}

// GetID returns a unique identifier for the item
func (ni *NotificationItem) GetID() string {
	return ni.notification.GetID()
}

// IsSelectable returns whether the item can be selected
func (ni *NotificationItem) IsSelectable() bool {
	return true
}

// SetCompact sets whether to use compact rendering
func (ni *NotificationItem) SetCompact(compact bool) {
	ni.compact = compact
}

// IsCompact returns whether compact rendering is enabled
func (ni *NotificationItem) IsCompact() bool {
	return ni.compact
}

// GetNotification returns the underlying GitHub notification
func (ni *NotificationItem) GetNotification() *github.Notification {
	return ni.notification
}

// formatRelativeTime formats a time as a relative string
func (ni *NotificationItem) formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 365*24*time.Hour:
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		years := int(diff.Hours() / (24 * 365))
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// getTypeText returns a human-readable type text
func (ni *NotificationItem) getTypeText(notificationType string) string {
	switch notificationType {
	case "PullRequest":
		return "PR"
	case "Issue":
		return "ISSUE"
	case "Release":
		return "RELEASE"
	case "Discussion":
		return "DISCUSSION"
	case "Commit":
		return "COMMIT"
	default:
		return strings.ToUpper(notificationType)
	}
}

// NotificationItemList is a helper for creating lists of notification items
type NotificationItemList struct {
	items   []*NotificationItem
	styles  EnhancedStyles
	symbols Symbols
	compact bool
}

// NewNotificationItemList creates a new notification item list
func NewNotificationItemList(notifications []*github.Notification, styles EnhancedStyles, symbols Symbols) *NotificationItemList {
	items := make([]*NotificationItem, len(notifications))
	for i, notification := range notifications {
		items[i] = NewNotificationItem(notification, styles, symbols)
	}

	return &NotificationItemList{
		items:   items,
		styles:  styles,
		symbols: symbols,
		compact: true,
	}
}

// GetVirtualListItems returns items as VirtualListItem interface
func (nil *NotificationItemList) GetVirtualListItems() []VirtualListItem {
	items := make([]VirtualListItem, len(nil.items))
	for i, item := range nil.items {
		items[i] = item
	}
	return items
}

// SetCompact sets compact mode for all items
func (nil *NotificationItemList) SetCompact(compact bool) {
	nil.compact = compact
	for _, item := range nil.items {
		item.SetCompact(compact)
	}
}

// GetItems returns the notification items
func (nil *NotificationItemList) GetItems() []*NotificationItem {
	return nil.items
}

// GetNotifications returns the underlying GitHub notifications
func (nil *NotificationItemList) GetNotifications() []*github.Notification {
	notifications := make([]*github.Notification, len(nil.items))
	for i, item := range nil.items {
		notifications[i] = item.GetNotification()
	}
	return notifications
}

// Filter filters items based on a predicate function
func (nil *NotificationItemList) Filter(predicate func(*github.Notification) bool) *NotificationItemList {
	var filteredItems []*NotificationItem

	for _, item := range nil.items {
		if predicate(item.GetNotification()) {
			filteredItems = append(filteredItems, item)
		}
	}

	return &NotificationItemList{
		items:   filteredItems,
		styles:  nil.styles,
		symbols: nil.symbols,
		compact: nil.compact,
	}
}

// Sort sorts items based on a comparison function
func (nil *NotificationItemList) Sort(less func(*github.Notification, *github.Notification) bool) {
	// Simple bubble sort for demonstration
	n := len(nil.items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if !less(nil.items[j].GetNotification(), nil.items[j+1].GetNotification()) {
				nil.items[j], nil.items[j+1] = nil.items[j+1], nil.items[j]
			}
		}
	}
}
