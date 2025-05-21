package firstrun

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/gh-notif/internal/wizard"
)

// FirstRunOptions contains options for the first-run wizard
type FirstRunOptions struct {
	// ConfigPath is the path to the configuration file
	ConfigPath string

	// Force forces the first-run wizard to run even if it has run before
	Force bool

	// NonInteractive disables interactive mode
	NonInteractive bool

	// SkipAuth skips the authentication step
	SkipAuth bool

	// SkipDisplay skips the display settings step
	SkipDisplay bool

	// SkipNotifications skips the notification settings step
	SkipNotifications bool

	// SkipAdvanced skips the advanced settings step
	SkipAdvanced bool

	// Theme sets the color theme
	Theme string

	// NoColor disables color output
	NoColor bool
}

// DefaultFirstRunOptions returns the default first-run options
func DefaultFirstRunOptions() *FirstRunOptions {
	return &FirstRunOptions{
		ConfigPath:        "",
		Force:             false,
		NonInteractive:    false,
		SkipAuth:          false,
		SkipDisplay:       false,
		SkipNotifications: false,
		SkipAdvanced:      false,
		Theme:             "auto",
		NoColor:           false,
	}
}

// FirstRun represents the first-run wizard
type FirstRun struct {
	options *FirstRunOptions
}

// NewFirstRun creates a new first-run wizard
func NewFirstRun(options *FirstRunOptions) *FirstRun {
	if options == nil {
		options = DefaultFirstRunOptions()
	}
	return &FirstRun{
		options: options,
	}
}

// ShouldRun returns whether the first-run wizard should run
func (f *FirstRun) ShouldRun() bool {
	if f.options.Force {
		return true
	}

	// Check if the configuration file exists
	configPath := f.options.ConfigPath
	if configPath == "" {
		// Use default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return true
		}
		configPath = filepath.Join(home, ".gh-notif.yaml")
	}

	// Check if the file exists
	_, err := os.Stat(configPath)
	return os.IsNotExist(err)
}

// Run runs the first-run wizard
func (f *FirstRun) Run() error {
	fmt.Println("Welcome to gh-notif!")
	fmt.Println("This appears to be your first time running gh-notif.")
	fmt.Println("Let's set up your configuration.")
	fmt.Println()

	// Create wizard options
	wizardOptions := &wizard.WizardOptions{
		Interactive:       !f.options.NonInteractive,
		Theme:             f.options.Theme,
		NoColor:           f.options.NoColor,
		Width:             80,
		Height:            20,
		ConfigPath:        f.options.ConfigPath,
		SkipAuth:          f.options.SkipAuth,
		SkipDisplay:       f.options.SkipDisplay,
		SkipNotifications: f.options.SkipNotifications,
		SkipAdvanced:      f.options.SkipAdvanced,
	}

	// Run the wizard
	if err := wizard.RunWizard(wizardOptions); err != nil {
		return fmt.Errorf("error running setup wizard: %w", err)
	}

	// Show success message
	fmt.Println()
	fmt.Println("Setup complete!")
	fmt.Println("You can now start using gh-notif.")
	fmt.Println()
	fmt.Println("Here are some commands to get you started:")
	fmt.Println("  gh-notif list                 - List your unread notifications")
	fmt.Println("  gh-notif tutorial             - Run the interactive tutorial")
	fmt.Println("  gh-notif ui                   - Start the interactive terminal UI")
	fmt.Println("  gh-notif help                 - Show help for all commands")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/user/gh-notif")

	return nil
}

// CheckAndRun checks if the first-run wizard should run and runs it if needed
func CheckAndRun(options *FirstRunOptions) error {
	firstRun := NewFirstRun(options)
	if firstRun.ShouldRun() {
		return firstRun.Run()
	}
	return nil
}

// MarkAsRun marks the first-run wizard as having been run
func MarkAsRun() error {
	// Create a marker file to indicate that the first-run wizard has been run
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}

	// Create .gh-notif directory if it doesn't exist
	ghNotifDir := filepath.Join(home, ".gh-notif")
	if err := os.MkdirAll(ghNotifDir, 0755); err != nil {
		return fmt.Errorf("error creating .gh-notif directory: %w", err)
	}

	// Create marker file
	markerFile := filepath.Join(ghNotifDir, ".first-run-complete")
	if err := os.WriteFile(markerFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("error creating marker file: %w", err)
	}

	return nil
}

// HasRun returns whether the first-run wizard has been run before
func HasRun() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	markerFile := filepath.Join(home, ".gh-notif", ".first-run-complete")
	_, err = os.Stat(markerFile)
	return err == nil
}

// ResetFirstRun resets the first-run state, causing the wizard to run again
func ResetFirstRun() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}

	markerFile := filepath.Join(home, ".gh-notif", ".first-run-complete")
	if err := os.Remove(markerFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error removing marker file: %w", err)
	}

	return nil
}

// RunFirstTimeSetup runs the first-time setup process
func RunFirstTimeSetup(configPath string, nonInteractive bool) error {
	options := DefaultFirstRunOptions()
	options.ConfigPath = configPath
	options.NonInteractive = nonInteractive

	return CheckAndRun(options)
}
