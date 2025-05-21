package grouping

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
)

// GroupType represents the type of grouping
type GroupType string

const (
	// GroupByRepository groups notifications by repository
	GroupByRepository GroupType = "repository"
	// GroupByOwner groups notifications by repository owner
	GroupByOwner GroupType = "owner"
	// GroupByType groups notifications by notification type
	GroupByType GroupType = "type"
	// GroupByReason groups notifications by notification reason
	GroupByReason GroupType = "reason"
	// GroupByThread groups notifications by thread (PR/issue/discussion)
	GroupByThread GroupType = "thread"
	// GroupByTime groups notifications by time period
	GroupByTime GroupType = "time"
	// GroupByScore groups notifications by score range
	GroupByScore GroupType = "score"
	// GroupBySmart uses an algorithm to group related notifications
	GroupBySmart GroupType = "smart"
)

// Group represents a group of notifications
type Group struct {
	// ID is a unique identifier for the group
	ID string `json:"id"`
	// Name is a human-readable name for the group
	Name string `json:"name"`
	// Type is the type of grouping
	Type GroupType `json:"type"`
	// Count is the number of notifications in the group
	Count int `json:"count"`
	// UnreadCount is the number of unread notifications in the group
	UnreadCount int `json:"unread_count"`
	// Notifications are the notifications in the group
	Notifications []*github.Notification `json:"notifications"`
	// Subgroups are optional subgroups
	Subgroups []*Group `json:"subgroups,omitempty"`
	// Parent is the parent group, if any
	Parent *Group `json:"-"`
	// Metadata is additional information about the group
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GroupOptions contains options for grouping notifications
type GroupOptions struct {
	// PrimaryGrouping is the primary grouping type
	PrimaryGrouping GroupType
	// SecondaryGrouping is an optional secondary grouping type
	SecondaryGrouping GroupType
	// MaxGroups is the maximum number of top-level groups
	MaxGroups int
	// MinGroupSize is the minimum size for a group
	MinGroupSize int
	// Concurrency is the number of goroutines to use
	Concurrency int
	// Timeout is the maximum time to spend grouping
	Timeout time.Duration
	// ScoreThresholds are the thresholds for score grouping
	ScoreThresholds []int
	// TimeThresholds are the thresholds for time grouping
	TimeThresholds []time.Duration
	// SmartGroupingThreshold is the similarity threshold for smart grouping
	SmartGroupingThreshold float64
}

// DefaultGroupOptions returns the default grouping options
func DefaultGroupOptions() *GroupOptions {
	return &GroupOptions{
		PrimaryGrouping:        GroupByRepository,
		SecondaryGrouping:      "",
		MaxGroups:              10,
		MinGroupSize:           2,
		Concurrency:            5,
		Timeout:                5 * time.Second,
		ScoreThresholds:        []int{25, 50, 75},
		TimeThresholds:         []time.Duration{24 * time.Hour, 7 * 24 * time.Hour, 30 * 24 * time.Hour},
		SmartGroupingThreshold: 0.7,
	}
}

// Grouper groups notifications
type Grouper struct {
	// Options are the grouping options
	Options *GroupOptions
}

// NewGrouper creates a new grouper
func NewGrouper(options *GroupOptions) *Grouper {
	if options == nil {
		options = DefaultGroupOptions()
	}
	return &Grouper{
		Options: options,
	}
}

// Group groups notifications
func (g *Grouper) Group(ctx context.Context, notifications []*github.Notification) ([]*Group, error) {
	if len(notifications) == 0 {
		return nil, nil
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, g.Options.Timeout)
	defer cancel()

	// Group by the primary grouping type
	groups, err := g.groupBy(ctx, notifications, g.Options.PrimaryGrouping)
	if err != nil {
		return nil, err
	}

	// Apply secondary grouping if specified
	if g.Options.SecondaryGrouping != "" {
		for _, group := range groups {
			subgroups, err := g.groupBy(ctx, group.Notifications, g.Options.SecondaryGrouping)
			if err != nil {
				return nil, err
			}
			group.Subgroups = subgroups
			for _, subgroup := range subgroups {
				subgroup.Parent = group
			}
		}
	}

	// Sort groups by count (descending)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	// Limit the number of groups
	if g.Options.MaxGroups > 0 && len(groups) > g.Options.MaxGroups {
		// Create an "Other" group for the remaining notifications
		otherGroup := &Group{
			ID:   "other",
			Name: "Other",
			Type: g.Options.PrimaryGrouping,
		}

		// Add the remaining groups to the "Other" group
		for _, group := range groups[g.Options.MaxGroups:] {
			otherGroup.Notifications = append(otherGroup.Notifications, group.Notifications...)
			otherGroup.Count += group.Count
			otherGroup.UnreadCount += group.UnreadCount
		}

		// Add the "Other" group to the list
		groups = append(groups[:g.Options.MaxGroups], otherGroup)
	}

	return groups, nil
}

// groupBy groups notifications by the specified type
func (g *Grouper) groupBy(ctx context.Context, notifications []*github.Notification, groupType GroupType) ([]*Group, error) {
	switch groupType {
	case GroupByRepository:
		return g.groupByRepository(notifications), nil
	case GroupByOwner:
		return g.groupByOwner(notifications), nil
	case GroupByType:
		return g.groupByType(notifications), nil
	case GroupByReason:
		return g.groupByReason(notifications), nil
	case GroupByThread:
		return g.groupByThread(notifications), nil
	case GroupByTime:
		return g.groupByTime(notifications), nil
	case GroupByScore:
		return g.groupByScore(notifications), nil
	case GroupBySmart:
		return g.groupBySmart(ctx, notifications), nil
	default:
		return nil, fmt.Errorf("unsupported grouping type: %s", groupType)
	}
}

// groupByRepository groups notifications by repository
func (g *Grouper) groupByRepository(notifications []*github.Notification) []*Group {
	// Create a map of repository name to notifications
	repoGroups := make(map[string][]*github.Notification)
	for _, n := range notifications {
		repo := n.GetRepository().GetFullName()
		repoGroups[repo] = append(repoGroups[repo], n)
	}

	// Create groups
	var groups []*Group
	for repo, ns := range repoGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("repo-%s", repo),
			Name:          repo,
			Type:          GroupByRepository,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// groupByOwner groups notifications by repository owner
func (g *Grouper) groupByOwner(notifications []*github.Notification) []*Group {
	// Create a map of owner name to notifications
	ownerGroups := make(map[string][]*github.Notification)
	for _, n := range notifications {
		fullName := n.GetRepository().GetFullName()
		parts := strings.Split(fullName, "/")
		if len(parts) < 2 {
			continue
		}
		owner := parts[0]
		ownerGroups[owner] = append(ownerGroups[owner], n)
	}

	// Create groups
	var groups []*Group
	for owner, ns := range ownerGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("owner-%s", owner),
			Name:          owner,
			Type:          GroupByOwner,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// groupByType groups notifications by type
func (g *Grouper) groupByType(notifications []*github.Notification) []*Group {
	// Create a map of type to notifications
	typeGroups := make(map[string][]*github.Notification)
	for _, n := range notifications {
		typ := n.GetSubject().GetType()
		typeGroups[typ] = append(typeGroups[typ], n)
	}

	// Create groups
	var groups []*Group
	for typ, ns := range typeGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("type-%s", typ),
			Name:          typ,
			Type:          GroupByType,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// groupByReason groups notifications by reason
func (g *Grouper) groupByReason(notifications []*github.Notification) []*Group {
	// Create a map of reason to notifications
	reasonGroups := make(map[string][]*github.Notification)
	for _, n := range notifications {
		reason := n.GetReason()
		reasonGroups[reason] = append(reasonGroups[reason], n)
	}

	// Create groups
	var groups []*Group
	for reason, ns := range reasonGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("reason-%s", reason),
			Name:          formatReason(reason),
			Type:          GroupByReason,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// formatReason formats a reason for display
func formatReason(reason string) string {
	switch reason {
	case "assign":
		return "Assigned"
	case "author":
		return "Authored"
	case "comment":
		return "Commented"
	case "mention":
		return "Mentioned"
	case "review_requested":
		return "Review Requested"
	case "state_change":
		return "State Changed"
	case "subscribed":
		return "Subscribed"
	case "team_mention":
		return "Team Mentioned"
	default:
		return strings.Title(reason)
	}
}

// groupByThread groups notifications by thread
func (g *Grouper) groupByThread(notifications []*github.Notification) []*Group {
	// This is a more complex grouping that requires additional API calls
	// For now, implement a simple version based on URL patterns

	// Create a map of thread ID to notifications
	threadGroups := make(map[string][]*github.Notification)
	for _, n := range notifications {
		// Extract the thread ID from the URL
		url := n.GetSubject().GetURL()
		threadID := extractThreadID(url)
		if threadID == "" {
			continue
		}
		threadGroups[threadID] = append(threadGroups[threadID], n)
	}

	// Create groups
	var groups []*Group
	for threadID, ns := range threadGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Get the thread title
		title := ns[0].GetSubject().GetTitle()

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("thread-%s", threadID),
			Name:          title,
			Type:          GroupByThread,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// extractThreadID extracts a thread ID from a URL
func extractThreadID(url string) string {
	// Extract the issue or PR number
	re := regexp.MustCompile(`/(?:issues|pull)/(\d+)$`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// groupByTime groups notifications by time period
func (g *Grouper) groupByTime(notifications []*github.Notification) []*Group {
	// Create time period groups
	now := time.Now()
	timeGroups := make(map[string][]*github.Notification)

	for _, n := range notifications {
		updatedAt := n.GetUpdatedAt().Time
		age := now.Sub(updatedAt)

		var period string
		if age < g.Options.TimeThresholds[0] {
			period = "Today"
		} else if age < g.Options.TimeThresholds[1] {
			period = "This Week"
		} else if age < g.Options.TimeThresholds[2] {
			period = "This Month"
		} else {
			period = "Older"
		}

		timeGroups[period] = append(timeGroups[period], n)
	}

	// Create groups
	var groups []*Group
	for period, ns := range timeGroups {
		// Skip small groups
		if len(ns) < g.Options.MinGroupSize {
			continue
		}

		// Count unread notifications
		unreadCount := 0
		for _, n := range ns {
			if n.GetUnread() {
				unreadCount++
			}
		}

		// Create the group
		groups = append(groups, &Group{
			ID:            fmt.Sprintf("time-%s", period),
			Name:          period,
			Type:          GroupByTime,
			Count:         len(ns),
			UnreadCount:   unreadCount,
			Notifications: ns,
		})
	}

	return groups
}

// groupByScore groups notifications by score range
func (g *Grouper) groupByScore(notifications []*github.Notification) []*Group {
	// This would require scoring the notifications first
	// For now, return an empty list
	return []*Group{}
}

// groupBySmart uses an algorithm to group related notifications
func (g *Grouper) groupBySmart(ctx context.Context, notifications []*github.Notification) []*Group {
	// This is a complex grouping that uses multiple factors
	// For now, implement a simple version based on title similarity

	// Create a map of groups
	groups := make(map[string]*Group)
	ungrouped := make([]*github.Notification, 0)

	// First pass: group by exact title match
	for _, n := range notifications {
		title := n.GetSubject().GetTitle()
		if title == "" {
			ungrouped = append(ungrouped, n)
			continue
		}

		groupID := fmt.Sprintf("smart-%s", title)
		if group, ok := groups[groupID]; ok {
			// Add to existing group
			group.Notifications = append(group.Notifications, n)
			group.Count++
			if n.GetUnread() {
				group.UnreadCount++
			}
		} else {
			// Create a new group
			groups[groupID] = &Group{
				ID:            groupID,
				Name:          title,
				Type:          GroupBySmart,
				Count:         1,
				UnreadCount:   boolToInt(n.GetUnread()),
				Notifications: []*github.Notification{n},
			}
		}
	}

	// Second pass: try to group ungrouped notifications by similarity
	for _, n := range ungrouped {
		title := n.GetSubject().GetTitle()
		if title == "" {
			continue
		}

		// Find the most similar group
		var bestGroup *Group
		bestSimilarity := 0.0

		for _, group := range groups {
			similarity := calculateTitleSimilarity(title, group.Name)
			if similarity > bestSimilarity && similarity >= g.Options.SmartGroupingThreshold {
				bestSimilarity = similarity
				bestGroup = group
			}
		}

		if bestGroup != nil {
			// Add to the best group
			bestGroup.Notifications = append(bestGroup.Notifications, n)
			bestGroup.Count++
			if n.GetUnread() {
				bestGroup.UnreadCount++
			}
		} else {
			// Create a new group
			groupID := fmt.Sprintf("smart-%s", title)
			groups[groupID] = &Group{
				ID:            groupID,
				Name:          title,
				Type:          GroupBySmart,
				Count:         1,
				UnreadCount:   boolToInt(n.GetUnread()),
				Notifications: []*github.Notification{n},
			}
		}
	}

	// Convert the map to a slice
	var result []*Group
	for _, group := range groups {
		// Skip small groups
		if group.Count < g.Options.MinGroupSize {
			continue
		}
		result = append(result, group)
	}

	return result
}

// calculateTitleSimilarity calculates the similarity between two titles
func calculateTitleSimilarity(a, b string) float64 {
	// Simple implementation using Jaccard similarity of words
	aWords := strings.Fields(strings.ToLower(a))
	bWords := strings.Fields(strings.ToLower(b))

	// Create sets of words
	aSet := make(map[string]bool)
	bSet := make(map[string]bool)

	for _, word := range aWords {
		aSet[word] = true
	}

	for _, word := range bWords {
		bSet[word] = true
	}

	// Calculate intersection and union sizes
	intersection := 0
	for word := range aSet {
		if bSet[word] {
			intersection++
		}
	}

	union := len(aSet) + len(bSet) - intersection

	// Calculate Jaccard similarity
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// boolToInt converts a bool to an int
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
