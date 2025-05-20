package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	// Authentication settings
	Auth AuthConfig `mapstructure:"auth"`

	// Display settings
	Display DisplayConfig `mapstructure:"display"`

	// Notification settings
	Notifications NotificationConfig `mapstructure:"notifications"`

	// API settings
	API APIConfig `mapstructure:"api"`

	// Advanced settings
	Advanced AdvancedConfig `mapstructure:"advanced"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	// ClientID is the GitHub OAuth client ID
	ClientID string `mapstructure:"client_id"`

	// ClientSecret is the GitHub OAuth client secret
	ClientSecret string `mapstructure:"client_secret"`

	// Scopes are the OAuth scopes to request
	Scopes []string `mapstructure:"scopes"`

	// TokenStorage defines how to store the OAuth token
	// Options: "keyring", "file", "auto"
	TokenStorage string `mapstructure:"token_storage"`
}

// DisplayConfig holds display-related configuration
type DisplayConfig struct {
	// Theme defines the color theme to use
	// Options: "dark", "light", "auto"
	Theme string `mapstructure:"theme"`

	// DateFormat defines how dates are displayed
	// Options: "relative", "absolute", "iso"
	DateFormat string `mapstructure:"date_format"`

	// ShowEmojis determines whether to show emojis in the output
	ShowEmojis bool `mapstructure:"show_emojis"`

	// CompactMode shows notifications in a more compact format
	CompactMode bool `mapstructure:"compact_mode"`

	// OutputFormat defines the output format for commands that support it
	// Options: "table", "json", "yaml", "text"
	OutputFormat string `mapstructure:"output_format"`
}

// NotificationConfig holds notification-related configuration
type NotificationConfig struct {
	// DefaultFilter is the default filter to apply when listing notifications
	// Options: "all", "unread", "participating"
	DefaultFilter string `mapstructure:"default_filter"`

	// IncludeRepos is a list of repositories to include (whitelist)
	IncludeRepos []string `mapstructure:"include_repos"`

	// ExcludeRepos is a list of repositories to exclude (blacklist)
	ExcludeRepos []string `mapstructure:"exclude_repos"`

	// IncludeOrgs is a list of organizations to include (whitelist)
	IncludeOrgs []string `mapstructure:"include_orgs"`

	// ExcludeOrgs is a list of organizations to exclude (blacklist)
	ExcludeOrgs []string `mapstructure:"exclude_orgs"`

	// IncludeTypes is a list of notification types to include
	// Options: "issue", "pr", "release", "discussion", etc.
	IncludeTypes []string `mapstructure:"include_types"`

	// ExcludeTypes is a list of notification types to exclude
	ExcludeTypes []string `mapstructure:"exclude_types"`

	// AutoRefresh automatically refreshes notifications
	AutoRefresh bool `mapstructure:"auto_refresh"`

	// RefreshInterval is the interval in seconds to refresh notifications
	RefreshInterval int `mapstructure:"refresh_interval"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	// BaseURL is the base URL for the GitHub API
	// Default: https://api.github.com
	BaseURL string `mapstructure:"base_url"`

	// UploadURL is the upload URL for the GitHub API
	// Default: https://uploads.github.com
	UploadURL string `mapstructure:"upload_url"`

	// Timeout is the timeout in seconds for API requests
	Timeout int `mapstructure:"timeout"`

	// RetryCount is the number of times to retry failed API requests
	RetryCount int `mapstructure:"retry_count"`

	// RetryDelay is the delay in seconds between retries
	RetryDelay int `mapstructure:"retry_delay"`
}

// AdvancedConfig holds advanced configuration options
type AdvancedConfig struct {
	// Debug enables debug logging
	Debug bool `mapstructure:"debug"`

	// MaxConcurrent is the maximum number of concurrent API requests
	MaxConcurrent int `mapstructure:"max_concurrent"`

	// CacheTTL is the time-to-live in seconds for cached data
	CacheTTL int `mapstructure:"cache_ttl"`

	// CacheDir is the directory to store cached data
	CacheDir string `mapstructure:"cache_dir"`

	// Editor is the preferred editor for editing configuration
	Editor string `mapstructure:"editor"`
}

// ConfigManager manages the application configuration
type ConfigManager struct {
	// viper instance
	v *viper.Viper

	// config is the current configuration
	config *Config

	// configFile is the path to the configuration file
	configFile string

	// configDir is the directory containing the configuration file
	configDir string
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		v: viper.New(),
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Auth: AuthConfig{
			Scopes:       []string{"notifications", "repo"},
			TokenStorage: "auto",
		},
		Display: DisplayConfig{
			Theme:        "auto",
			DateFormat:   "relative",
			ShowEmojis:   true,
			CompactMode:  false,
			OutputFormat: "table",
		},
		Notifications: NotificationConfig{
			DefaultFilter:   "unread",
			AutoRefresh:     false,
			RefreshInterval: 60,
		},
		API: APIConfig{
			BaseURL:    "https://api.github.com",
			UploadURL:  "https://uploads.github.com",
			Timeout:    30,
			RetryCount: 3,
			RetryDelay: 1,
		},
		Advanced: AdvancedConfig{
			Debug:         false,
			MaxConcurrent: 5,
			CacheTTL:      3600,
			Editor:        getDefaultEditor(),
		},
	}
}

// getDefaultEditor returns the default editor based on the OS
func getDefaultEditor() string {
	// Check for environment variables
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Default based on OS
	switch runtime.GOOS {
	case "windows":
		return "notepad"
	case "darwin":
		return "nano"
	default:
		// Try to find a common editor
		for _, editor := range []string{"nano", "vim", "vi", "emacs"} {
			if _, err := exec.LookPath(editor); err == nil {
				return editor
			}
		}
		return "nano" // Fallback
	}
}

// Load loads the configuration from file and environment variables
func (cm *ConfigManager) Load() error {
	// Set up defaults
	defaultConfig := DefaultConfig()
	cm.setDefaults(defaultConfig)

	// Find config file locations
	if err := cm.setupConfigLocations(); err != nil {
		return err
	}

	// Set up environment variables
	cm.v.SetEnvPrefix("GH_NOTIF")
	cm.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cm.v.AutomaticEnv()

	// Read config file
	if err := cm.v.ReadInConfig(); err != nil {
		// Create default config if it doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := cm.createDefaultConfig(); err != nil {
				return fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal config
	config := DefaultConfig()
	if err := cm.v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := cm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.config = config
	return nil
}

// setDefaults sets the default values in viper
func (cm *ConfigManager) setDefaults(config *Config) {
	// Auth defaults
	cm.v.SetDefault("auth.scopes", config.Auth.Scopes)
	cm.v.SetDefault("auth.token_storage", config.Auth.TokenStorage)

	// Display defaults
	cm.v.SetDefault("display.theme", config.Display.Theme)
	cm.v.SetDefault("display.date_format", config.Display.DateFormat)
	cm.v.SetDefault("display.show_emojis", config.Display.ShowEmojis)
	cm.v.SetDefault("display.compact_mode", config.Display.CompactMode)
	cm.v.SetDefault("display.output_format", config.Display.OutputFormat)

	// Notification defaults
	cm.v.SetDefault("notifications.default_filter", config.Notifications.DefaultFilter)
	cm.v.SetDefault("notifications.auto_refresh", config.Notifications.AutoRefresh)
	cm.v.SetDefault("notifications.refresh_interval", config.Notifications.RefreshInterval)

	// API defaults
	cm.v.SetDefault("api.base_url", config.API.BaseURL)
	cm.v.SetDefault("api.upload_url", config.API.UploadURL)
	cm.v.SetDefault("api.timeout", config.API.Timeout)
	cm.v.SetDefault("api.retry_count", config.API.RetryCount)
	cm.v.SetDefault("api.retry_delay", config.API.RetryDelay)

	// Advanced defaults
	cm.v.SetDefault("advanced.debug", config.Advanced.Debug)
	cm.v.SetDefault("advanced.max_concurrent", config.Advanced.MaxConcurrent)
	cm.v.SetDefault("advanced.cache_ttl", config.Advanced.CacheTTL)
	cm.v.SetDefault("advanced.editor", config.Advanced.Editor)
}

// setupConfigLocations sets up the configuration file locations
func (cm *ConfigManager) setupConfigLocations() error {
	// Find home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to find home directory: %w", err)
	}

	// Set up config file name and type
	cm.v.SetConfigName(".gh-notif")
	cm.v.SetConfigType("yaml")

	// Add config paths in order of precedence
	// 1. Current directory
	cm.v.AddConfigPath(".")

	// 2. XDG config directory if available
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		cm.v.AddConfigPath(filepath.Join(xdgConfigHome, "gh-notif"))
	}

	// 3. Platform-specific config directory
	switch runtime.GOOS {
	case "windows":
		cm.v.AddConfigPath(filepath.Join(os.Getenv("APPDATA"), "gh-notif"))
	case "darwin":
		cm.v.AddConfigPath(filepath.Join(home, "Library", "Application Support", "gh-notif"))
	default: // Linux and others
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome == "" {
			cm.v.AddConfigPath(filepath.Join(home, ".config", "gh-notif"))
		}
	}

	// 4. Home directory
	cm.v.AddConfigPath(home)

	// Set config file and directory
	if cm.v.ConfigFileUsed() != "" {
		cm.configFile = cm.v.ConfigFileUsed()
		cm.configDir = filepath.Dir(cm.configFile)
	} else {
		// Default to home directory
		cm.configDir = home
		cm.configFile = filepath.Join(home, ".gh-notif.yaml")
	}

	return nil
}

// createDefaultConfig creates a default configuration file
func (cm *ConfigManager) createDefaultConfig() error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write default config
	if err := cm.v.SafeWriteConfigAs(cm.configFile); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// validateConfig validates the configuration
func (cm *ConfigManager) validateConfig(config *Config) error {
	// Validate display settings
	if !contains([]string{"dark", "light", "auto"}, config.Display.Theme) {
		return errors.New("invalid theme: must be 'dark', 'light', or 'auto'")
	}

	if !contains([]string{"relative", "absolute", "iso"}, config.Display.DateFormat) {
		return errors.New("invalid date format: must be 'relative', 'absolute', or 'iso'")
	}

	if !contains([]string{"table", "json", "yaml", "text"}, config.Display.OutputFormat) {
		return errors.New("invalid output format: must be 'table', 'json', 'yaml', or 'text'")
	}

	// Validate notification settings
	if !contains([]string{"all", "unread", "participating"}, config.Notifications.DefaultFilter) {
		return errors.New("invalid default filter: must be 'all', 'unread', or 'participating'")
	}

	if config.Notifications.RefreshInterval < 0 {
		return errors.New("invalid refresh interval: must be non-negative")
	}

	// Validate API settings
	if config.API.Timeout <= 0 {
		return errors.New("invalid timeout: must be positive")
	}

	if config.API.RetryCount < 0 {
		return errors.New("invalid retry count: must be non-negative")
	}

	if config.API.RetryDelay < 0 {
		return errors.New("invalid retry delay: must be non-negative")
	}

	// Validate advanced settings
	if config.Advanced.MaxConcurrent <= 0 {
		return errors.New("invalid max concurrent: must be positive")
	}

	if config.Advanced.CacheTTL < 0 {
		return errors.New("invalid cache TTL: must be non-negative")
	}

	return nil
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}


