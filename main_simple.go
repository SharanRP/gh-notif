package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-notif",
	Short: "A high-performance GitHub notification manager",
	Long: `gh-notif is a CLI tool for managing GitHub notifications in the terminal.
It allows you to view, filter, and interact with your GitHub notifications efficiently.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to gh-notif!")
		fmt.Println("Use --help to see available commands.")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of gh-notif`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gh-notif v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
