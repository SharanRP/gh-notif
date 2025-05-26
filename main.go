package main

import (
	"fmt"
	"os"

	"github.com/SharanRP/gh-notif/internal/config"
	"github.com/SharanRP/gh-notif/internal/version"
	"github.com/spf13/cobra"
)

// Version information set by ldflags during build
var (
	versionString = "dev"
	commitString  = "unknown"
	dateString    = "unknown"
	builtByString = "unknown"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "gh-notif",
		Short: "A high-performance GitHub notification manager",
		Long: `gh-notif is a CLI tool for managing GitHub notifications in the terminal.
It allows you to view, filter, and interact with your GitHub notifications efficiently.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
				return nil
			}

			configManager := config.NewConfigManager()
			if cfgFile != "" {
				// TODO: Set custom config file
			}

			if err := configManager.Load(); err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			return nil
		},
	}
)

func main() {
	version.Version = versionString
	version.Commit = commitString
	version.Date = dateString
	version.BuiltBy = builtByString

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gh-notif.yaml)")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Show version information for gh-notif.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("gh-notif version %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Date: %s\n", version.Date)
			fmt.Printf("Built by: %s\n", version.BuiltBy)
		},
	}
	rootCmd.AddCommand(versionCmd)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List GitHub notifications",
		Long:  `List GitHub notifications with filtering options.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gh-notif list command")
			fmt.Println("This is a placeholder implementation.")
			fmt.Println("Run 'gh-notif version' to see version information.")
		},
	}
	rootCmd.AddCommand(listCmd)
}
