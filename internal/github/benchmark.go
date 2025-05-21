package github

import (
	"context"
	"strings"

	"github.com/google/go-github/v60/github"
)

// FilterNotifications filters notifications based on a filter string
func (c *Client) FilterNotifications(notifications []*github.Notification, filterString, sortBy string) []*github.Notification {
	if filterString == "" {
		return notifications
	}

	// Parse the filter string
	filters := parseBenchmarkFilterString(filterString)
	if len(filters) == 0 {
		return notifications
	}

	// Apply filters
	var filtered []*github.Notification
	for _, n := range notifications {
		if matchesFilters(n, filters) {
			filtered = append(filtered, n)
		}
	}

	// Sort if requested
	if sortBy != "" {
		filtered = sortNotifications(filtered, sortBy)
	}

	return filtered
}

// Filter represents a parsed filter
type Filter struct {
	Field    string
	Operator string
	Value    string
}

// parseBenchmarkFilterString parses a filter string into filters
func parseBenchmarkFilterString(filterString string) []Filter {
	var filters []Filter

	// Split by spaces, but respect quotes
	parts := strings.Fields(filterString)
	for _, part := range parts {
		// Parse field:value or field=value or field>value
		var field, operator, value string
		if strings.Contains(part, ":") {
			parts := strings.SplitN(part, ":", 2)
			field = parts[0]
			operator = ":"
			value = parts[1]
		} else if strings.Contains(part, "=") {
			parts := strings.SplitN(part, "=", 2)
			field = parts[0]
			operator = "="
			value = parts[1]
		} else if strings.Contains(part, ">") {
			parts := strings.SplitN(part, ">", 2)
			field = parts[0]
			operator = ">"
			value = parts[1]
		} else if strings.Contains(part, "<") {
			parts := strings.SplitN(part, "<", 2)
			field = parts[0]
			operator = "<"
			value = parts[1]
		} else {
			// Treat as a search term
			field = "text"
			operator = ":"
			value = part
		}

		// Add the filter
		filters = append(filters, Filter{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}

	return filters
}

// matchesFilters checks if a notification matches all filters
func matchesFilters(n *github.Notification, filters []Filter) bool {
	for _, filter := range filters {
		if !matchesFilter(n, filter) {
			return false
		}
	}
	return true
}

// matchesFilter checks if a notification matches a filter
func matchesFilter(n *github.Notification, filter Filter) bool {
	switch filter.Field {
	case "repo", "repository":
		return strings.Contains(strings.ToLower(n.GetRepository().GetFullName()), strings.ToLower(filter.Value))
	case "type":
		return strings.EqualFold(n.GetSubject().GetType(), filter.Value)
	case "reason":
		return strings.EqualFold(n.GetReason(), filter.Value)
	case "title":
		return strings.Contains(strings.ToLower(n.GetSubject().GetTitle()), strings.ToLower(filter.Value))
	case "text":
		// Search in title, repo, and type
		return strings.Contains(strings.ToLower(n.GetSubject().GetTitle()), strings.ToLower(filter.Value)) ||
			strings.Contains(strings.ToLower(n.GetRepository().GetFullName()), strings.ToLower(filter.Value)) ||
			strings.Contains(strings.ToLower(n.GetSubject().GetType()), strings.ToLower(filter.Value))
	case "unread":
		return n.GetUnread() == (filter.Value == "true" || filter.Value == "yes" || filter.Value == "1")
	default:
		return false
	}
}

// sortNotifications sorts notifications by the specified field
func sortNotifications(notifications []*github.Notification, sortBy string) []*github.Notification {
	// For now, just return the original list
	// In a real implementation, we would sort by the specified field
	return notifications
}

// GroupNotifications groups notifications by the specified field
func (c *Client) GroupNotifications(ctx context.Context, notifications []*github.Notification, groupBy string) ([]*NotificationGroup, error) {
	if len(notifications) == 0 {
		return nil, nil
	}

	// Group by the specified field
	groups := make(map[string]*NotificationGroup)

	for _, n := range notifications {
		var key string
		var name string

		switch groupBy {
		case "repository", "repo":
			key = n.GetRepository().GetFullName()
			name = key
		case "type":
			key = n.GetSubject().GetType()
			name = key
		case "reason":
			key = n.GetReason()
			name = key
		case "status":
			if n.GetUnread() {
				key = "unread"
				name = "Unread"
			} else {
				key = "read"
				name = "Read"
			}
		default:
			key = "all"
			name = "All Notifications"
		}

		// Create or update the group
		group, ok := groups[key]
		if !ok {
			group = &NotificationGroup{
				ID:            key,
				Name:          name,
				Type:          groupBy,
				Notifications: []*github.Notification{},
			}
			groups[key] = group
		}

		// Add the notification to the group
		group.Notifications = append(group.Notifications, n)
		group.Count++
		if n.GetUnread() {
			group.UnreadCount++
		}
	}

	// Convert the map to a slice
	var result []*NotificationGroup
	for _, group := range groups {
		result = append(result, group)
	}

	return result, nil
}

// NotificationGroup represents a group of notifications
type NotificationGroup struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Count         int                    `json:"count"`
	UnreadCount   int                    `json:"unread_count"`
	Notifications []*github.Notification `json:"notifications"`
}
