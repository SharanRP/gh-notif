package main

import (
	"fmt"
	"os"

	"github.com/SharanRP/gh-notif/cmd/gh-notif"
	"github.com/SharanRP/gh-notif/internal/version"
)

// Version information set by ldflags during build
var (
	versionString = "dev"
	commitString  = "unknown"
	dateString    = "unknown"
	builtByString = "unknown"
)

func main() {
	// Set version information
	version.Version = versionString
	version.Commit = commitString
	version.Date = dateString
	version.BuiltBy = builtByString

	if err := ghnotif.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
