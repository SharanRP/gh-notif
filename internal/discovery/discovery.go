package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FeatureState represents the state of a feature
type FeatureState struct {
	// Discovered indicates whether the feature has been discovered
	Discovered bool `json:"discovered"`

	// FirstSeen is the time when the feature was first seen
	FirstSeen time.Time `json:"first_seen"`

	// LastSeen is the time when the feature was last seen
	LastSeen time.Time `json:"last_seen"`

	// UsageCount is the number of times the feature has been used
	UsageCount int `json:"usage_count"`

	// Dismissed indicates whether the feature tip has been dismissed
	Dismissed bool `json:"dismissed"`
}

// DiscoveryState represents the state of feature discovery
type DiscoveryState struct {
	// Features is a map of feature IDs to their states
	Features map[string]FeatureState `json:"features"`

	// LastUpdated is the time when the state was last updated
	LastUpdated time.Time `json:"last_updated"`

	// UserLevel is the user's experience level
	UserLevel string `json:"user_level"`

	// CompletedTutorial indicates whether the user has completed the tutorial
	CompletedTutorial bool `json:"completed_tutorial"`

	// UsageCount is the number of times the application has been used
	UsageCount int `json:"usage_count"`

	// FirstRun is the time when the application was first run
	FirstRun time.Time `json:"first_run"`
}

// DiscoveryManager manages feature discovery
type DiscoveryManager struct {
	// State is the current discovery state
	State DiscoveryState

	// StatePath is the path to the state file
	StatePath string

	// EnableDiscovery indicates whether feature discovery is enabled
	EnableDiscovery bool

	// EnableTips indicates whether feature tips are enabled
	EnableTips bool

	// EnableProgressiveDisclosure indicates whether progressive disclosure is enabled
	EnableProgressiveDisclosure bool
}

// NewDiscoveryManager creates a new discovery manager
func NewDiscoveryManager() (*DiscoveryManager, error) {
	// Get the state file path
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home directory: %w", err)
	}

	statePath := filepath.Join(home, ".gh-notif", "discovery.json")

	// Create the manager
	manager := &DiscoveryManager{
		State: DiscoveryState{
			Features:          make(map[string]FeatureState),
			LastUpdated:       time.Now(),
			UserLevel:         "beginner",
			CompletedTutorial: false,
			UsageCount:        0,
			FirstRun:          time.Now(),
		},
		StatePath:                   statePath,
		EnableDiscovery:             true,
		EnableTips:                  true,
		EnableProgressiveDisclosure: true,
	}

	// Load the state if it exists
	if err := manager.Load(); err != nil {
		// If the file doesn't exist, that's fine
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error loading discovery state: %w", err)
		}
	}

	return manager, nil
}

// Load loads the discovery state from disk
func (m *DiscoveryManager) Load() error {
	// Read the state file
	data, err := os.ReadFile(m.StatePath)
	if err != nil {
		return err
	}

	// Parse the state
	if err := json.Unmarshal(data, &m.State); err != nil {
		return fmt.Errorf("error parsing discovery state: %w", err)
	}

	return nil
}

// Save saves the discovery state to disk
func (m *DiscoveryManager) Save() error {
	// Update the last updated time
	m.State.LastUpdated = time.Now()

	// Create the directory if it doesn't exist
	dir := filepath.Dir(m.StatePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Marshal the state
	data, err := json.MarshalIndent(m.State, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling discovery state: %w", err)
	}

	// Write the state file
	if err := os.WriteFile(m.StatePath, data, 0644); err != nil {
		return fmt.Errorf("error writing discovery state: %w", err)
	}

	return nil
}

// RecordUsage records usage of the application
func (m *DiscoveryManager) RecordUsage() error {
	// Increment the usage count
	m.State.UsageCount++

	// Save the state
	return m.Save()
}

// RecordFeatureUsage records usage of a feature
func (m *DiscoveryManager) RecordFeatureUsage(featureID string) error {
	// Get the feature state
	state, ok := m.State.Features[featureID]
	if !ok {
		// Feature not seen before
		state = FeatureState{
			Discovered: true,
			FirstSeen:  time.Now(),
			LastSeen:   time.Now(),
			UsageCount: 1,
			Dismissed:  false,
		}
	} else {
		// Feature seen before
		state.Discovered = true
		state.LastSeen = time.Now()
		state.UsageCount++
	}

	// Update the feature state
	m.State.Features[featureID] = state

	// Save the state
	return m.Save()
}

// DismissFeatureTip dismisses a feature tip
func (m *DiscoveryManager) DismissFeatureTip(featureID string) error {
	// Get the feature state
	state, ok := m.State.Features[featureID]
	if !ok {
		// Feature not seen before
		state = FeatureState{
			Discovered: false,
			FirstSeen:  time.Now(),
			LastSeen:   time.Now(),
			UsageCount: 0,
			Dismissed:  true,
		}
	} else {
		// Feature seen before
		state.Dismissed = true
	}

	// Update the feature state
	m.State.Features[featureID] = state

	// Save the state
	return m.Save()
}

// IsFeatureDiscovered returns whether a feature has been discovered
func (m *DiscoveryManager) IsFeatureDiscovered(featureID string) bool {
	// Get the feature state
	state, ok := m.State.Features[featureID]
	if !ok {
		return false
	}

	return state.Discovered
}

// IsFeatureTipDismissed returns whether a feature tip has been dismissed
func (m *DiscoveryManager) IsFeatureTipDismissed(featureID string) bool {
	// Get the feature state
	state, ok := m.State.Features[featureID]
	if !ok {
		return false
	}

	return state.Dismissed
}

// GetFeatureUsageCount returns the number of times a feature has been used
func (m *DiscoveryManager) GetFeatureUsageCount(featureID string) int {
	// Get the feature state
	state, ok := m.State.Features[featureID]
	if !ok {
		return 0
	}

	return state.UsageCount
}

// SetUserLevel sets the user's experience level
func (m *DiscoveryManager) SetUserLevel(level string) error {
	// Update the user level
	m.State.UserLevel = level

	// Save the state
	return m.Save()
}

// GetUserLevel returns the user's experience level
func (m *DiscoveryManager) GetUserLevel() string {
	return m.State.UserLevel
}

// SetCompletedTutorial sets whether the user has completed the tutorial
func (m *DiscoveryManager) SetCompletedTutorial(completed bool) error {
	// Update the completed tutorial flag
	m.State.CompletedTutorial = completed

	// Save the state
	return m.Save()
}

// HasCompletedTutorial returns whether the user has completed the tutorial
func (m *DiscoveryManager) HasCompletedTutorial() bool {
	return m.State.CompletedTutorial
}

// GetUsageCount returns the number of times the application has been used
func (m *DiscoveryManager) GetUsageCount() int {
	return m.State.UsageCount
}

// GetFirstRunTime returns the time when the application was first run
func (m *DiscoveryManager) GetFirstRunTime() time.Time {
	return m.State.FirstRun
}

// ShouldShowFeatureTip returns whether a feature tip should be shown
func (m *DiscoveryManager) ShouldShowFeatureTip(featureID string) bool {
	// Check if tips are enabled
	if !m.EnableTips {
		return false
	}

	// Check if the feature tip has been dismissed
	if m.IsFeatureTipDismissed(featureID) {
		return false
	}

	// Check if the feature has been discovered
	if m.IsFeatureDiscovered(featureID) {
		return false
	}

	// Check if progressive disclosure is enabled
	if !m.EnableProgressiveDisclosure {
		return true
	}

	// Check the user's experience level
	switch m.GetUserLevel() {
	case "beginner":
		// Show all tips to beginners
		return true
	case "intermediate":
		// Show intermediate and advanced tips to intermediate users
		return isIntermediateOrAdvancedFeature(featureID)
	case "advanced":
		// Show only advanced tips to advanced users
		return isAdvancedFeature(featureID)
	default:
		return true
	}
}

// isIntermediateOrAdvancedFeature returns whether a feature is intermediate or advanced
func isIntermediateOrAdvancedFeature(featureID string) bool {
	// Define intermediate and advanced features
	intermediateFeatures := map[string]bool{
		"filter.complex":      true,
		"filter.save":         true,
		"group.smart":         true,
		"search.regex":        true,
		"watch.desktop":       true,
		"ui.split":            true,
		"ui.keyboard":         true,
		"ui.batch":            true,
		"config.custom":       true,
		"profile.basic":       true,
		"scoring.basic":       true,
		"actions.batch":       true,
		"actions.undo":        true,
		"notifications.mute":  true,
		"notifications.watch": true,
	}

	advancedFeatures := map[string]bool{
		"filter.boolean":       true,
		"filter.composition":   true,
		"group.hierarchical":   true,
		"search.advanced":      true,
		"watch.backoff":        true,
		"ui.custom":            true,
		"ui.themes":            true,
		"config.advanced":      true,
		"profile.http":         true,
		"profile.benchmark":    true,
		"scoring.custom":       true,
		"actions.scripting":    true,
		"notifications.smart":  true,
		"notifications.score":  true,
		"notifications.custom": true,
	}

	return intermediateFeatures[featureID] || advancedFeatures[featureID]
}

// isAdvancedFeature returns whether a feature is advanced
func isAdvancedFeature(featureID string) bool {
	// Define advanced features
	advancedFeatures := map[string]bool{
		"filter.boolean":       true,
		"filter.composition":   true,
		"group.hierarchical":   true,
		"search.advanced":      true,
		"watch.backoff":        true,
		"ui.custom":            true,
		"ui.themes":            true,
		"config.advanced":      true,
		"profile.http":         true,
		"profile.benchmark":    true,
		"scoring.custom":       true,
		"actions.scripting":    true,
		"notifications.smart":  true,
		"notifications.score":  true,
		"notifications.custom": true,
	}

	return advancedFeatures[featureID]
}
