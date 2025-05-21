package persistent

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/gh-notif/internal/config"
)

func TestFilterStore(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "filter-store-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a config manager
	configManager := config.NewConfigManager()
	// Set the config file path through environment variable
	os.Setenv("GH_NOTIF_CONFIG", configFile)

	// Create a filter store
	store, err := NewFilterStore(configManager)
	if err != nil {
		t.Fatalf("Failed to create filter store: %v", err)
	}

	// Test saving a filter
	preset := &FilterPreset{
		Name:        "test-filter",
		Description: "Test filter",
		Expression:  "repo:owner/repo is:unread",
		Created:     time.Now(),
	}

	if err := store.Save(preset); err != nil {
		t.Fatalf("Failed to save filter: %v", err)
	}

	// Test getting a filter
	retrieved, err := store.Get("test-filter")
	if err != nil {
		t.Fatalf("Failed to get filter: %v", err)
	}

	if retrieved.Name != preset.Name {
		t.Errorf("Expected name %s, got %s", preset.Name, retrieved.Name)
	}

	if retrieved.Expression != preset.Expression {
		t.Errorf("Expected expression %s, got %s", preset.Expression, retrieved.Expression)
	}

	// Test listing filters
	filters := store.List()
	if len(filters) != 1 {
		t.Errorf("Expected 1 filter, got %d", len(filters))
	}

	// Test deleting a filter
	if err := store.Delete("test-filter"); err != nil {
		t.Fatalf("Failed to delete filter: %v", err)
	}

	// Verify the filter was deleted
	_, err = store.Get("test-filter")
	if err == nil {
		t.Errorf("Expected error getting deleted filter, got nil")
	}

	// Test shortcuts
	shortcuts := store.ListShortcuts()
	if len(shortcuts) == 0 {
		t.Errorf("Expected shortcuts, got none")
	}

	// Test getting a shortcut
	shortcutPreset, err := store.Get("unread")
	if err != nil {
		t.Fatalf("Failed to get shortcut: %v", err)
	}

	if shortcutPreset.Expression != "is:unread" {
		t.Errorf("Expected expression 'is:unread', got '%s'", shortcutPreset.Expression)
	}
}

func TestParserSimple(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "parser-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a config manager
	configManager := config.NewConfigManager()
	// Set the config file path through environment variable
	os.Setenv("GH_NOTIF_CONFIG", configFile)

	// Create a filter store
	store, err := NewFilterStore(configManager)
	if err != nil {
		t.Fatalf("Failed to create filter store: %v", err)
	}

	// Create a parser
	parser := NewParser(store)

	// Test parsing a simple expression
	filter, err := parser.Parse("repo:owner/repo")
	if err != nil {
		t.Fatalf("Failed to parse expression: %v", err)
	}

	if filter == nil {
		t.Fatalf("Expected filter, got nil")
	}

	// Test parsing a complex expression
	filter, err = parser.Parse("repo:owner/repo AND is:unread")
	if err != nil {
		t.Fatalf("Failed to parse complex expression: %v", err)
	}

	if filter == nil {
		t.Fatalf("Expected filter, got nil")
	}

	// Test parsing a reference
	preset := &FilterPreset{
		Name:        "test-filter",
		Description: "Test filter",
		Expression:  "repo:owner/repo is:unread",
		Created:     time.Now(),
	}

	if err := store.Save(preset); err != nil {
		t.Fatalf("Failed to save filter: %v", err)
	}

	filter, err = parser.Parse("@test-filter")
	if err != nil {
		t.Fatalf("Failed to parse reference: %v", err)
	}

	if filter == nil {
		t.Fatalf("Expected filter, got nil")
	}
}

func TestParserComplex(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "parser-complex-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Create a config manager
	configManager := config.NewConfigManager()
	// Set the config file path through environment variable
	os.Setenv("GH_NOTIF_CONFIG", configFile)

	// Create a filter store
	store, err := NewFilterStore(configManager)
	if err != nil {
		t.Fatalf("Failed to create filter store: %v", err)
	}

	// Create a parser
	parser := NewParser(store)

	// Test parsing a complex expression with parentheses
	filter, err := parser.Parse("(repo:owner/repo OR repo:other/repo) AND is:unread")
	if err != nil {
		t.Fatalf("Failed to parse complex expression: %v", err)
	}

	if filter == nil {
		t.Fatalf("Expected filter, got nil")
	}

	// Test parsing a complex expression with NOT
	filter, err = parser.Parse("repo:owner/repo AND NOT is:read")
	if err != nil {
		t.Fatalf("Failed to parse complex expression with NOT: %v", err)
	}

	if filter == nil {
		t.Fatalf("Expected filter, got nil")
	}
}
