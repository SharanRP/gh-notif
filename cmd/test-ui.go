package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/ui/examples"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/test-ui.go demo     - Run the enhanced UI demo")
		fmt.Println("  go run cmd/test-ui.go notif    - Test with sample notifications")
		fmt.Println("  go run cmd/test-ui.go help     - Show this help")
		return
	}

	switch os.Args[1] {
	case "demo":
		fmt.Println("🚀 Starting Enhanced UI Demo...")
		fmt.Println("This will showcase all the enhanced UI components:")
		fmt.Println("- Virtual Lists with 1000+ items")
		fmt.Println("- Interactive Forms with validation")
		fmt.Println("- Progress indicators with animations")
		fmt.Println("- Markdown rendering with syntax highlighting")
		fmt.Println("- Layout system demonstrations")
		fmt.Println("- Enhanced styling effects")
		fmt.Println()
		fmt.Println("Navigation:")
		fmt.Println("- n: Next demo")
		fmt.Println("- p: Previous demo")
		fmt.Println("- q: Quit")
		fmt.Println()
		fmt.Println("Press Enter to continue...")
		fmt.Scanln()

		if err := examples.RunEnhancedUIDemo(); err != nil {
			log.Fatalf("Error running demo: %v", err)
		}

	case "notif":
		fmt.Println("🔔 Testing Enhanced UI with Sample Notifications...")
		notifications := createSampleNotifications(50)
		
		// Import the enhanced UI function
		// Note: This would normally be imported from internal/ui
		fmt.Printf("Created %d sample notifications\n", len(notifications))
		fmt.Println("Press Enter to launch enhanced UI...")
		fmt.Scanln()

		// For now, just show that we have the notifications
		fmt.Println("Enhanced notification UI would launch here with:")
		for i, notif := range notifications[:5] { // Show first 5
			fmt.Printf("  %d. %s - %s (%s)\n", 
				i+1, 
				notif.GetRepository().GetFullName(),
				notif.GetSubject().GetTitle(),
				notif.GetSubject().GetType())
		}
		fmt.Printf("  ... and %d more notifications\n", len(notifications)-5)

	case "help":
		fmt.Println("🎨 Enhanced UI Test Runner")
		fmt.Println()
		fmt.Println("This tool demonstrates the enhanced terminal UI features:")
		fmt.Println()
		fmt.Println("Features Demonstrated:")
		fmt.Println("  🚀 High Performance")
		fmt.Println("    - Virtualized lists for thousands of items")
		fmt.Println("    - Render caching for optimal performance")
		fmt.Println("    - Concurrent operations with progress feedback")
		fmt.Println()
		fmt.Println("  🎨 Rich Styling & Theming")
		fmt.Println("    - Multiple themes (dark, light, high contrast)")
		fmt.Println("    - Gradient text effects")
		fmt.Println("    - Adaptive colors based on terminal capabilities")
		fmt.Println("    - Custom badges and panels")
		fmt.Println()
		fmt.Println("  ⌨️ Exceptional Keyboard Support")
		fmt.Println("    - Vim-style navigation (j/k, g/G)")
		fmt.Println("    - Tab navigation between components")
		fmt.Println("    - Custom key bindings")
		fmt.Println("    - Context-aware help")
		fmt.Println()
		fmt.Println("  ♿ Accessibility Features")
		fmt.Println("    - Screen reader support")
		fmt.Println("    - High contrast mode")
		fmt.Println("    - Keyboard-only navigation")
		fmt.Println("    - Color blind friendly themes")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  demo  - Interactive demo of all components")
		fmt.Println("  notif - Test with sample GitHub notifications")
		fmt.Println("  help  - Show this help message")

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Run 'go run cmd/test-ui.go help' for usage information")
	}
}

// createSampleNotifications creates sample GitHub notifications for testing
func createSampleNotifications(count int) []*github.Notification {
	notifications := make([]*github.Notification, count)
	
	repos := []string{
		"microsoft/vscode",
		"golang/go", 
		"kubernetes/kubernetes",
		"facebook/react",
		"torvalds/linux",
		"nodejs/node",
		"rust-lang/rust",
		"python/cpython",
		"tensorflow/tensorflow",
		"docker/docker",
	}
	
	types := []string{"PullRequest", "Issue", "Release", "Discussion", "Commit"}
	reasons := []string{"assign", "author", "comment", "mention", "review_requested", "subscribed"}
	
	titles := []string{
		"Fix critical security vulnerability in authentication",
		"Add support for new API endpoints",
		"Improve performance of database queries",
		"Update documentation with latest examples",
		"Refactor legacy code for better maintainability",
		"Add comprehensive test coverage",
		"Fix memory leak in background processing",
		"Implement new feature requested by users",
		"Update dependencies to latest versions",
		"Fix compatibility issues with latest OS",
		"Add internationalization support",
		"Optimize build process for faster CI/CD",
		"Fix race condition in concurrent operations",
		"Add dark mode support to UI components",
		"Improve error handling and logging",
	}
	
	for i := 0; i < count; i++ {
		repo := repos[i%len(repos)]
		notifType := types[i%len(types)]
		reason := reasons[i%len(reasons)]
		title := titles[i%len(titles)]
		unread := i%3 == 0 // Every third notification is unread
		
		notification := &github.Notification{
			ID: github.String(fmt.Sprintf("notification_%d", i)),
			Repository: &github.Repository{
				FullName: github.String(repo),
			},
			Subject: &github.NotificationSubject{
				Title: github.String(fmt.Sprintf("%s (#%d)", title, i+1)),
				Type:  github.String(notifType),
				URL:   github.String(fmt.Sprintf("https://github.com/%s/issues/%d", repo, i+1)),
			},
			Reason:    github.String(reason),
			Unread:    github.Bool(unread),
			UpdatedAt: &github.Timestamp{Time: time.Now().Add(-time.Duration(i) * time.Hour)},
		}
		
		notifications[i] = notification
	}
	
	return notifications
}
