package components

import (
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ComponentType represents the type of a UI component
type ComponentType string

const (
	ListComponentType        ComponentType = "list"
	FormComponentType        ComponentType = "form"
	ProgressComponentType    ComponentType = "progress"
	MarkdownComponentType    ComponentType = "markdown"
	BadgeComponentType       ComponentType = "badge"
	PanelComponentType       ComponentType = "panel"
	VirtualListComponentType ComponentType = "virtuallist"
)

// Component represents a reusable UI component
type Component interface {
	// Init initializes the component
	Init() tea.Cmd

	// Update handles messages and updates the component state
	Update(tea.Msg) (Component, tea.Cmd)

	// View renders the component
	View() string

	// SetSize sets the component dimensions
	SetSize(width, height int)

	// GetSize returns the component dimensions
	GetSize() (width, height int)

	// SetStyles sets the component styles
	SetStyles(styles ComponentStyles)

	// GetType returns the component type
	GetType() ComponentType

	// SetFocused sets the focus state
	SetFocused(focused bool)

	// IsFocused returns the focus state
	IsFocused() bool
}

// ComponentStyles contains styling information for components
type ComponentStyles struct {
	Base     lipgloss.Style
	Focused  lipgloss.Style
	Disabled lipgloss.Style
	Error    lipgloss.Style
	Success  lipgloss.Style
	Warning  lipgloss.Style
	Info     lipgloss.Style
}

// ComponentRegistry manages UI components
type ComponentRegistry struct {
	mu         sync.RWMutex
	components map[string]Component
	factories  map[ComponentType]ComponentFactory
}

// ComponentFactory creates new component instances
type ComponentFactory func(config ComponentConfig) Component

// ComponentConfig contains configuration for creating components
type ComponentConfig struct {
	Type   ComponentType
	ID     string
	Width  int
	Height int
	Styles ComponentStyles
	Props  map[string]interface{}
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	registry := &ComponentRegistry{
		components: make(map[string]Component),
		factories:  make(map[ComponentType]ComponentFactory),
	}

	// Register default component factories
	registry.registerDefaultFactories()

	return registry
}

// RegisterFactory registers a component factory
func (r *ComponentRegistry) RegisterFactory(componentType ComponentType, factory ComponentFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[componentType] = factory
}

// CreateComponent creates a new component instance
func (r *ComponentRegistry) CreateComponent(config ComponentConfig) (Component, error) {
	r.mu.RLock()
	factory, exists := r.factories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, ErrComponentTypeNotRegistered
	}

	component := factory(config)

	r.mu.Lock()
	r.components[config.ID] = component
	r.mu.Unlock()

	return component, nil
}

// GetComponent retrieves a component by ID
func (r *ComponentRegistry) GetComponent(id string) (Component, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	component, exists := r.components[id]
	return component, exists
}

// RemoveComponent removes a component from the registry
func (r *ComponentRegistry) RemoveComponent(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.components, id)
}

// ListComponents returns all registered component IDs
func (r *ComponentRegistry) ListComponents() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.components))
	for id := range r.components {
		ids = append(ids, id)
	}
	return ids
}

// registerDefaultFactories registers the default component factories
func (r *ComponentRegistry) registerDefaultFactories() {
	r.RegisterFactory(ListComponentType, NewListComponentFactory)
	r.RegisterFactory(FormComponentType, NewFormComponentFactory)
	r.RegisterFactory(ProgressComponentType, NewProgressComponentFactory)
	r.RegisterFactory(MarkdownComponentType, NewMarkdownComponentFactory)
	r.RegisterFactory(BadgeComponentType, NewBadgeComponentFactory)
	r.RegisterFactory(PanelComponentType, NewPanelComponentFactory)
	r.RegisterFactory(VirtualListComponentType, NewVirtualListComponentFactory)
}

// Global component registry instance
var globalRegistry *ComponentRegistry
var registryOnce sync.Once

// GetGlobalRegistry returns the global component registry
func GetGlobalRegistry() *ComponentRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewComponentRegistry()
	})
	return globalRegistry
}

// ComponentMessage represents a message sent to/from components
type ComponentMessage struct {
	ComponentID string
	Type        string
	Data        interface{}
}

// ComponentEvent represents an event emitted by a component
type ComponentEvent struct {
	ComponentID string
	EventType   string
	Data        interface{}
}

// Common component messages
const (
	ComponentFocusMsg   = "focus"
	ComponentBlurMsg    = "blur"
	ComponentResizeMsg  = "resize"
	ComponentRefreshMsg = "refresh"
	ComponentSelectMsg  = "select"
	ComponentChangeMsg  = "change"
)

// Common errors
var (
	ErrComponentTypeNotRegistered = fmt.Errorf("component type not registered")
	ErrComponentNotFound          = fmt.Errorf("component not found")
	ErrInvalidComponentConfig     = fmt.Errorf("invalid component configuration")
)
