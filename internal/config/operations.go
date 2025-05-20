package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// GetConfigFile returns the path to the configuration file
func (cm *ConfigManager) GetConfigFile() string {
	return cm.configFile
}

// Save saves the current configuration to file
func (cm *ConfigManager) Save() error {
	// Set all values in viper
	cm.setConfigValues(cm.config)

	// Write config file
	return cm.v.WriteConfig()
}

// setConfigValues sets all configuration values in viper
func (cm *ConfigManager) setConfigValues(config *Config) {
	// Auth settings
	cm.v.Set("auth.client_id", config.Auth.ClientID)
	cm.v.Set("auth.client_secret", config.Auth.ClientSecret)
	cm.v.Set("auth.scopes", config.Auth.Scopes)
	cm.v.Set("auth.token_storage", config.Auth.TokenStorage)

	// Display settings
	cm.v.Set("display.theme", config.Display.Theme)
	cm.v.Set("display.date_format", config.Display.DateFormat)
	cm.v.Set("display.show_emojis", config.Display.ShowEmojis)
	cm.v.Set("display.compact_mode", config.Display.CompactMode)
	cm.v.Set("display.output_format", config.Display.OutputFormat)

	// Notification settings
	cm.v.Set("notifications.default_filter", config.Notifications.DefaultFilter)
	cm.v.Set("notifications.include_repos", config.Notifications.IncludeRepos)
	cm.v.Set("notifications.exclude_repos", config.Notifications.ExcludeRepos)
	cm.v.Set("notifications.include_orgs", config.Notifications.IncludeOrgs)
	cm.v.Set("notifications.exclude_orgs", config.Notifications.ExcludeOrgs)
	cm.v.Set("notifications.include_types", config.Notifications.IncludeTypes)
	cm.v.Set("notifications.exclude_types", config.Notifications.ExcludeTypes)
	cm.v.Set("notifications.auto_refresh", config.Notifications.AutoRefresh)
	cm.v.Set("notifications.refresh_interval", config.Notifications.RefreshInterval)

	// API settings
	cm.v.Set("api.base_url", config.API.BaseURL)
	cm.v.Set("api.upload_url", config.API.UploadURL)
	cm.v.Set("api.timeout", config.API.Timeout)
	cm.v.Set("api.retry_count", config.API.RetryCount)
	cm.v.Set("api.retry_delay", config.API.RetryDelay)

	// Advanced settings
	cm.v.Set("advanced.debug", config.Advanced.Debug)
	cm.v.Set("advanced.max_concurrent", config.Advanced.MaxConcurrent)
	cm.v.Set("advanced.cache_ttl", config.Advanced.CacheTTL)
	cm.v.Set("advanced.cache_dir", config.Advanced.CacheDir)
	cm.v.Set("advanced.editor", config.Advanced.Editor)
}

// GetValue gets a configuration value by key
func (cm *ConfigManager) GetValue(key string) (interface{}, error) {
	// Check if the key exists
	if !cm.v.IsSet(key) {
		return nil, fmt.Errorf("configuration key not found: %s", key)
	}

	// Get the value
	return cm.v.Get(key), nil
}

// SetValue sets a configuration value by key
func (cm *ConfigManager) SetValue(key string, value interface{}) error {
	// Validate the key
	if err := cm.validateKey(key); err != nil {
		return err
	}

	// Validate the value
	if err := cm.validateValue(key, value); err != nil {
		return err
	}

	// Set the value
	cm.v.Set(key, value)

	// Update the config struct
	if err := cm.v.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	// Save the configuration
	return cm.Save()
}

// validateKey validates a configuration key
func (cm *ConfigManager) validateKey(key string) error {
	// Split the key into parts
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid key format: %s (expected section.key)", key)
	}

	// Check if the section exists
	section := parts[0]
	validSections := []string{"auth", "display", "notifications", "api", "advanced"}
	if !contains(validSections, section) {
		return fmt.Errorf("invalid configuration section: %s (valid sections: %s)", section, strings.Join(validSections, ", "))
	}

	// Check if the key exists in the section
	configValue := reflect.ValueOf(cm.config).Elem()
	sectionValue := configValue.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, section)
	})

	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", section)
	}

	// Check if the key exists in the section
	keyName := parts[1]
	keyValue := sectionValue.FieldByNameFunc(func(name string) bool {
		return strings.EqualFold(name, keyName)
	})

	if !keyValue.IsValid() {
		return fmt.Errorf("invalid configuration key: %s", key)
	}

	return nil
}

// validateValue validates a configuration value
func (cm *ConfigManager) validateValue(key string, value interface{}) error {
	// Validate based on the key
	switch key {
	// Display settings
	case "display.theme":
		if str, ok := value.(string); ok {
			if !contains([]string{"dark", "light", "auto"}, str) {
				return errors.New("invalid theme: must be 'dark', 'light', or 'auto'")
			}
		} else {
			return errors.New("theme must be a string")
		}
	case "display.date_format":
		if str, ok := value.(string); ok {
			if !contains([]string{"relative", "absolute", "iso"}, str) {
				return errors.New("invalid date format: must be 'relative', 'absolute', or 'iso'")
			}
		} else {
			return errors.New("date format must be a string")
		}
	case "display.output_format":
		if str, ok := value.(string); ok {
			if !contains([]string{"table", "json", "yaml", "text"}, str) {
				return errors.New("invalid output format: must be 'table', 'json', 'yaml', or 'text'")
			}
		} else {
			return errors.New("output format must be a string")
		}

	// Notification settings
	case "notifications.default_filter":
		if str, ok := value.(string); ok {
			if !contains([]string{"all", "unread", "participating"}, str) {
				return errors.New("invalid default filter: must be 'all', 'unread', or 'participating'")
			}
		} else {
			return errors.New("default filter must be a string")
		}
	case "notifications.refresh_interval":
		if num, ok := value.(int); ok {
			if num < 0 {
				return errors.New("refresh interval must be non-negative")
			}
		} else {
			return errors.New("refresh interval must be an integer")
		}

	// API settings
	case "api.timeout":
		if num, ok := value.(int); ok {
			if num <= 0 {
				return errors.New("timeout must be positive")
			}
		} else {
			return errors.New("timeout must be an integer")
		}
	case "api.retry_count":
		if num, ok := value.(int); ok {
			if num < 0 {
				return errors.New("retry count must be non-negative")
			}
		} else {
			return errors.New("retry count must be an integer")
		}
	case "api.retry_delay":
		if num, ok := value.(int); ok {
			if num < 0 {
				return errors.New("retry delay must be non-negative")
			}
		} else {
			return errors.New("retry delay must be an integer")
		}

	// Advanced settings
	case "advanced.max_concurrent":
		if num, ok := value.(int); ok {
			if num <= 0 {
				return errors.New("max concurrent must be positive")
			}
		} else {
			return errors.New("max concurrent must be an integer")
		}
	case "advanced.cache_ttl":
		if num, ok := value.(int); ok {
			if num < 0 {
				return errors.New("cache TTL must be non-negative")
			}
		} else {
			return errors.New("cache TTL must be an integer")
		}
	}

	return nil
}

// ListConfig returns a formatted list of all configuration values
func (cm *ConfigManager) ListConfig() (string, error) {
	// Get all settings
	allSettings := cm.v.AllSettings()

	// Format as YAML
	data, err := yaml.Marshal(allSettings)
	if err != nil {
		return "", fmt.Errorf("failed to format configuration: %w", err)
	}

	return string(data), nil
}

// EditConfig opens the configuration file in an editor
func (cm *ConfigManager) EditConfig() error {
	// Get the editor from config or environment
	editor := cm.config.Advanced.Editor
	if editor == "" {
		editor = getDefaultEditor()
	}

	// Create the command
	cmd := exec.Command(editor, cm.configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the editor
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	// Reload the configuration
	return cm.Load()
}

// ExportConfig exports the configuration to a file
func (cm *ConfigManager) ExportConfig(format, filePath string) error {
	// Get all settings
	allSettings := cm.v.AllSettings()

	// Format the data
	var data []byte
	var err error

	switch format {
	case "json":
		data, err = json.MarshalIndent(allSettings, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(allSettings)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ImportConfig imports configuration from a file
func (cm *ConfigManager) ImportConfig(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Determine the format based on the file extension
	format := filepath.Ext(filePath)
	if format == "" {
		return fmt.Errorf("unknown file format: %s", filePath)
	}
	format = format[1:] // Remove the leading dot

	// Parse the data
	var settings map[string]interface{}

	switch format {
	case "json":
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	case "yaml", "yml":
		if err := yaml.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	// Update the configuration
	for k, v := range settings {
		cm.v.Set(k, v)
	}

	// Unmarshal the updated configuration
	if err := cm.v.Unmarshal(cm.config); err != nil {
		return fmt.Errorf("failed to update configuration: %w", err)
	}

	// Validate the configuration
	if err := cm.validateConfig(cm.config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Save the configuration
	return cm.Save()
}
