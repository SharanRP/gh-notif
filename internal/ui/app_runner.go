package ui

import (
	"context"
	"fmt"

	githubclient "github.com/user/gh-notif/internal/github"
)

// RunApp runs the main application UI
func RunApp(ctx context.Context, client *githubclient.Client) error {
	// Fetch notifications
	options := githubclient.NotificationOptions{
		All:      false,
		UseCache: true,
	}

	notifications, err := client.GetUnreadNotifications(options)
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	// Display notifications with enhanced UI
	return DisplayEnhancedNotifications(notifications)
}
