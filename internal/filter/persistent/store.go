package persistent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SharanRP/gh-notif/internal/config"
)

// FilterPreset represents a saved filter preset
type FilterPreset struct {
	// Name is the name of the preset
	Name string `json:"name"`
	// Description is a human-readable description
	Description string `json:"description,omitempty"`
	// Expression is the filter expression
	Expression string `json:"expression"`
	// Created is when the preset was created
	Created time.Time `json:"created"`
	// LastUsed is when the preset was last used
	LastUsed time.Time `json:"last_used,omitempty"`
	// UseCount is how many times the preset has been used
	UseCount int `json:"use_count"`
	// Parent is the name of the parent preset (for inheritance)
	Parent string `json:"parent,omitempty"`
	// Tags are optional tags for categorizing presets
	Tags []string `json:"tags,omitempty"`
}

// FilterStore manages saved filter presets
type FilterStore struct {
	// presets is a map of preset name to preset
	presets map[string]*FilterPreset
	// configManager is used to save/load presets
	configManager *config.ConfigManager
	// filePath is the path to the presets file
	filePath string
	// mu protects the presets map
	mu sync.RWMutex
	// shortcuts maps shortcut names to preset names
	shortcuts map[string]string
}

// NewFilterStore creates a new filter store
func NewFilterStore(configManager *config.ConfigManager) (*FilterStore, error) {
	// Create the store
	store := &FilterStore{
		presets:       make(map[string]*FilterPreset),
		configManager: configManager,
		shortcuts:     make(map[string]string),
	}

	// Set up the presets file path
	configDir := filepath.Dir(configManager.GetConfigFile())
	store.filePath = filepath.Join(configDir, "filter_presets.json")

	// Load existing presets
	if err := store.Load(); err != nil {
		// If the file doesn't exist, that's fine
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load filter presets: %w", err)
		}
	}

	// Set up default shortcuts
	store.setupDefaultShortcuts()

	return store, nil
}

// setupDefaultShortcuts sets up default shortcuts for common filters
func (s *FilterStore) setupDefaultShortcuts() {
	s.shortcuts = map[string]string{
		"unread":     "is:unread",
		"read":       "is:read",
		"prs":        "type:PullRequest",
		"issues":     "type:Issue",
		"mentions":   "reason:mention",
		"assigned":   "reason:assign",
		"reviews":    "reason:review_requested",
		"today":      "updated:>24h",
		"week":       "updated:>7d",
		"high":       "score:>75",
		"medium":     "score:>50 score:<75",
		"low":        "score:<50",
	}
}

// Save saves a filter preset
func (s *FilterStore) Save(preset *FilterPreset) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set created time if not set
	if preset.Created.IsZero() {
		preset.Created = time.Now()
	}

	// Validate the preset
	if preset.Name == "" {
		return fmt.Errorf("preset name cannot be empty")
	}

	if preset.Expression == "" {
		return fmt.Errorf("preset expression cannot be empty")
	}

	// Check if the preset already exists
	if existing, ok := s.presets[preset.Name]; ok {
		// Update the existing preset
		existing.Description = preset.Description
		existing.Expression = preset.Expression
		existing.LastUsed = preset.LastUsed
		existing.UseCount = preset.UseCount
		existing.Parent = preset.Parent
		existing.Tags = preset.Tags
	} else {
		// Add the new preset
		s.presets[preset.Name] = preset
	}

	// Save to disk
	return s.saveToFile()
}

// Get gets a filter preset by name
func (s *FilterStore) Get(name string) (*FilterPreset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if it's a shortcut
	if shortcutExpr, ok := s.shortcuts[name]; ok {
		// Create a temporary preset for the shortcut
		return &FilterPreset{
			Name:       name,
			Expression: shortcutExpr,
			Created:    time.Now(),
		}, nil
	}

	// Look up the preset
	preset, ok := s.presets[name]
	if !ok {
		return nil, fmt.Errorf("filter preset not found: %s", name)
	}

	return preset, nil
}

// Delete deletes a filter preset
func (s *FilterStore) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the preset exists
	if _, ok := s.presets[name]; !ok {
		return fmt.Errorf("filter preset not found: %s", name)
	}

	// Delete the preset
	delete(s.presets, name)

	// Save to disk
	return s.saveToFile()
}

// List lists all filter presets
func (s *FilterStore) List() []*FilterPreset {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a list of presets
	presets := make([]*FilterPreset, 0, len(s.presets))
	for _, preset := range s.presets {
		presets = append(presets, preset)
	}

	return presets
}

// ListShortcuts lists all shortcuts
func (s *FilterStore) ListShortcuts() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy of the shortcuts map
	shortcuts := make(map[string]string, len(s.shortcuts))
	for name, expr := range s.shortcuts {
		shortcuts[name] = expr
	}

	return shortcuts
}

// Load loads filter presets from disk
func (s *FilterStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read the file
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	// Parse the JSON
	var presets map[string]*FilterPreset
	if err := json.Unmarshal(data, &presets); err != nil {
		return fmt.Errorf("failed to parse filter presets: %w", err)
	}

	// Set the presets
	s.presets = presets

	return nil
}

// saveToFile saves filter presets to disk
func (s *FilterStore) saveToFile() error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the presets to JSON
	data, err := json.MarshalIndent(s.presets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal filter presets: %w", err)
	}

	// Write the file
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write filter presets: %w", err)
	}

	return nil
}
