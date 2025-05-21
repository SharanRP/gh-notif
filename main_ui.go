package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/gh-notif/internal/auth"
	"github.com/user/gh-notif/internal/github"
	"github.com/user/gh-notif/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "gh-notif-ui",
	Short: "A high-performance GitHub notification manager",
	Long: `gh-notif-ui is a CLI tool for managing GitHub notifications in the terminal.
It allows you to view, filter, and interact with your GitHub notifications efficiently.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to gh-notif-ui!")
		fmt.Println("Use --help to see available commands.")
	},
}

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the terminal UI",
	Long: `Launch the terminal UI for GitHub notifications.
	
This provides an interactive interface for viewing and managing your GitHub notifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a context
		ctx := context.Background()

		// Check authentication status
		authenticated, _, err := auth.Status()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking auth status: %v\n", err)
			os.Exit(1)
		}

		if !authenticated {
			fmt.Fprintln(os.Stderr, "Not authenticated. Please run 'gh-notif auth login' first.")
			os.Exit(1)
		}

		// Create a GitHub client
		client, err := github.NewClient(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating GitHub client: %v\n", err)
			os.Exit(1)
		}

		// Run the UI
		if err := ui.RunApp(ctx, client); err != nil {
			fmt.Fprintf(os.Stderr, "Error running UI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
