package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/SharanRP/gh-notif/internal/filter"
	"github.com/google/go-github/v60/github"
)

// GetFilteredNotifications fetches notifications matching a filter
func (c *Client) GetFilteredNotifications(opts NotificationOptions) ([]*github.Notification, error) {
	// Fetch all notifications first
	var notifications []*github.Notification
	var err error

	// Choose the appropriate method based on the options
	if opts.RepoName != "" {
		notifications, err = c.GetNotificationsByRepo(opts.RepoName, opts)
	} else if opts.OrgName != "" {
		notifications, err = c.GetNotificationsByOrg(opts.OrgName, opts)
	} else if !opts.All {
		notifications, err = c.GetUnreadNotifications(opts)
	} else {
		notifications, err = c.GetAllNotifications(opts)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// If no filter string is provided, return all notifications
	if opts.FilterString == "" {
		return notifications, nil
	}

	// Parse the filter string
	filterExpr, err := parseFilterString(opts.FilterString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filter: %w", err)
	}

	// Create a filter engine
	engine := filter.NewEngine().WithFilter(filterExpr)

	// Apply the filter
	filtered, err := engine.Filter(context.Background(), notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to apply filter: %w", err)
	}

	return filtered, nil
}

// GetNotification fetches a single notification by ID
func (c *Client) GetNotification(id string) (*github.Notification, error) {
	// Fetch all notifications
	notifications, err := c.GetAllNotifications(NotificationOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	// Find the notification with the given ID
	for _, n := range notifications {
		if n.GetID() == id {
			return n, nil
		}
	}

	return nil, fmt.Errorf("notification not found: %s", id)
}

// parseFilterString parses a filter string into a filter expression
func parseFilterString(filterStr string) (filter.Filter, error) {
	// This is a simplified implementation
	// In a real implementation, we would parse the filter string into a proper filter expression

	// Split the filter string into parts
	parts := strings.Fields(filterStr)

	// Create a composite filter
	filters := make([]filter.Filter, 0, len(parts))

	for _, part := range parts {
		// Check for key:value format
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			key := kv[0]
			value := kv[1]

			switch key {
			case "is":
				if value == "read" {
					filters = append(filters, &filter.ReadFilter{Read: true})
				} else if value == "unread" {
					filters = append(filters, &filter.ReadFilter{Read: false})
				}
			case "repo":
				filters = append(filters, &filter.RepoFilter{Repo: value})
			case "org":
				filters = append(filters, &filter.OrgFilter{Org: value})
			case "type":
				filters = append(filters, filter.NewTypeFilter(value))
			case "reason":
				filters = append(filters, &filter.ReasonFilter{Reason: value})
			}
		} else {
			// Simple text search
			filters = append(filters, &filter.TextFilter{Text: part})
		}
	}

	// If no filters were created, return a filter that matches everything
	if len(filters) == 0 {
		return &filter.AllFilter{}, nil
	}

	// If there's only one filter, return it
	if len(filters) == 1 {
		return filters[0], nil
	}

	// Otherwise, return a composite filter
	return &filter.AndFilter{Filters: filters}, nil
}

// ConvertAPIURLToWebURL converts a GitHub API URL to a web URL
func ConvertAPIURLToWebURL(apiURL string) (string, error) {
	// This is a simplified implementation
	// In a real implementation, we would parse the URL and convert it properly
	webURL := strings.Replace(apiURL, "api.github.com", "github.com", 1)
	webURL = strings.Replace(webURL, "/repos/", "/", 1)

	// Convert specific endpoints
	webURL = strings.Replace(webURL, "/pulls/", "/pull/", 1)
	webURL = strings.Replace(webURL, "/issues/", "/issue/", 1)

	return webURL, nil
}
