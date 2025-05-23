package components

import (
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// VirtualListItem represents an item in the virtual list
type VirtualListItem interface {
	// Render renders the item with the given width and style
	Render(width int, style lipgloss.Style) string
	
	// GetHeight returns the height of the item when rendered
	GetHeight() int
	
	// GetID returns a unique identifier for the item
	GetID() string
	
	// IsSelectable returns whether the item can be selected
	IsSelectable() bool
}

// VirtualList implements a high-performance virtualized list component
type VirtualList struct {
	// Configuration
	width         int
	height        int
	items         []VirtualListItem
	itemHeight    int // Fixed height per item for simplicity
	
	// State
	focused       bool
	selected      int
	offset        int
	visibleStart  int
	visibleEnd    int
	
	// Styling
	styles        ComponentStyles
	
	// Key bindings
	keyMap        VirtualListKeyMap
	
	// Performance tracking
	renderCache   map[string]string
	cacheEnabled  bool
}

// VirtualListKeyMap defines key bindings for the virtual list
type VirtualListKeyMap struct {
	Up         key.Binding
	Down       key.Binding
	PageUp     key.Binding
	PageDown   key.Binding
	Home       key.Binding
	End        key.Binding
	Select     key.Binding
}

// DefaultVirtualListKeyMap returns the default key bindings
func DefaultVirtualListKeyMap() VirtualListKeyMap {
	return VirtualListKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to bottom"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter", "select"),
		),
	}
}

// NewVirtualList creates a new virtual list component
func NewVirtualList(items []VirtualListItem, itemHeight int) *VirtualList {
	return &VirtualList{
		items:        items,
		itemHeight:   itemHeight,
		selected:     0,
		offset:       0,
		keyMap:       DefaultVirtualListKeyMap(),
		renderCache:  make(map[string]string),
		cacheEnabled: true,
	}
}

// NewVirtualListComponentFactory creates a virtual list component factory
func NewVirtualListComponentFactory(config ComponentConfig) Component {
	items, ok := config.Props["items"].([]VirtualListItem)
	if !ok {
		items = []VirtualListItem{}
	}
	
	itemHeight, ok := config.Props["itemHeight"].(int)
	if !ok {
		itemHeight = 1
	}
	
	vl := NewVirtualList(items, itemHeight)
	vl.SetSize(config.Width, config.Height)
	vl.SetStyles(config.Styles)
	
	return vl
}

// Init initializes the virtual list
func (vl *VirtualList) Init() tea.Cmd {
	vl.updateVisibleRange()
	return nil
}

// Update handles messages and updates the virtual list state
func (vl *VirtualList) Update(msg tea.Msg) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !vl.focused {
			return vl, nil
		}
		
		switch {
		case key.Matches(msg, vl.keyMap.Up):
			vl.moveUp()
		case key.Matches(msg, vl.keyMap.Down):
			vl.moveDown()
		case key.Matches(msg, vl.keyMap.PageUp):
			vl.pageUp()
		case key.Matches(msg, vl.keyMap.PageDown):
			vl.pageDown()
		case key.Matches(msg, vl.keyMap.Home):
			vl.goToTop()
		case key.Matches(msg, vl.keyMap.End):
			vl.goToBottom()
		case key.Matches(msg, vl.keyMap.Select):
			return vl, vl.selectItem()
		}
		
		vl.updateVisibleRange()
		
	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				vl.SetSize(size.Width, size.Height)
			}
		case ComponentRefreshMsg:
			vl.clearCache()
		}
	}
	
	return vl, nil
}

// View renders the virtual list
func (vl *VirtualList) View() string {
	if len(vl.items) == 0 {
		return vl.styles.Base.Render("No items")
	}
	
	var content []string
	
	// Render only visible items for performance
	for i := vl.visibleStart; i <= vl.visibleEnd && i < len(vl.items); i++ {
		item := vl.items[i]
		
		// Check cache first
		cacheKey := fmt.Sprintf("%s_%d_%t", item.GetID(), vl.width, i == vl.selected)
		if vl.cacheEnabled {
			if cached, exists := vl.renderCache[cacheKey]; exists {
				content = append(content, cached)
				continue
			}
		}
		
		// Determine style
		var style lipgloss.Style
		if i == vl.selected && vl.focused {
			style = vl.styles.Focused
		} else {
			style = vl.styles.Base
		}
		
		// Render item
		rendered := item.Render(vl.width, style)
		
		// Cache the result
		if vl.cacheEnabled {
			vl.renderCache[cacheKey] = rendered
		}
		
		content = append(content, rendered)
	}
	
	// Join all content
	result := lipgloss.JoinVertical(lipgloss.Left, content...)
	
	// Apply container style
	containerStyle := vl.styles.Base
	if vl.focused {
		containerStyle = vl.styles.Focused
	}
	
	return containerStyle.Width(vl.width).Height(vl.height).Render(result)
}

// SetSize sets the component dimensions
func (vl *VirtualList) SetSize(width, height int) {
	vl.width = width
	vl.height = height
	vl.updateVisibleRange()
	vl.clearCache() // Clear cache when size changes
}

// GetSize returns the component dimensions
func (vl *VirtualList) GetSize() (width, height int) {
	return vl.width, vl.height
}

// SetStyles sets the component styles
func (vl *VirtualList) SetStyles(styles ComponentStyles) {
	vl.styles = styles
	vl.clearCache() // Clear cache when styles change
}

// GetType returns the component type
func (vl *VirtualList) GetType() ComponentType {
	return VirtualListComponentType
}

// SetFocused sets the focus state
func (vl *VirtualList) SetFocused(focused bool) {
	vl.focused = focused
	vl.clearCache() // Clear cache when focus changes
}

// IsFocused returns the focus state
func (vl *VirtualList) IsFocused() bool {
	return vl.focused
}

// SetItems updates the list items
func (vl *VirtualList) SetItems(items []VirtualListItem) {
	vl.items = items
	vl.selected = 0
	vl.offset = 0
	vl.updateVisibleRange()
	vl.clearCache()
}

// GetItems returns the current items
func (vl *VirtualList) GetItems() []VirtualListItem {
	return vl.items
}

// GetSelected returns the currently selected item index
func (vl *VirtualList) GetSelected() int {
	return vl.selected
}

// SetSelected sets the selected item index
func (vl *VirtualList) SetSelected(index int) {
	if index >= 0 && index < len(vl.items) {
		vl.selected = index
		vl.ensureVisible()
		vl.updateVisibleRange()
	}
}

// GetSelectedItem returns the currently selected item
func (vl *VirtualList) GetSelectedItem() VirtualListItem {
	if vl.selected >= 0 && vl.selected < len(vl.items) {
		return vl.items[vl.selected]
	}
	return nil
}

// Movement methods
func (vl *VirtualList) moveUp() {
	if vl.selected > 0 {
		vl.selected--
		vl.ensureVisible()
	}
}

func (vl *VirtualList) moveDown() {
	if vl.selected < len(vl.items)-1 {
		vl.selected++
		vl.ensureVisible()
	}
}

func (vl *VirtualList) pageUp() {
	pageSize := vl.height / vl.itemHeight
	vl.selected = int(math.Max(0, float64(vl.selected-pageSize)))
	vl.ensureVisible()
}

func (vl *VirtualList) pageDown() {
	pageSize := vl.height / vl.itemHeight
	vl.selected = int(math.Min(float64(len(vl.items)-1), float64(vl.selected+pageSize)))
	vl.ensureVisible()
}

func (vl *VirtualList) goToTop() {
	vl.selected = 0
	vl.offset = 0
}

func (vl *VirtualList) goToBottom() {
	vl.selected = len(vl.items) - 1
	vl.ensureVisible()
}

// ensureVisible ensures the selected item is visible
func (vl *VirtualList) ensureVisible() {
	visibleItems := vl.height / vl.itemHeight
	
	if vl.selected < vl.offset {
		vl.offset = vl.selected
	} else if vl.selected >= vl.offset+visibleItems {
		vl.offset = vl.selected - visibleItems + 1
	}
	
	// Ensure offset is within bounds
	maxOffset := int(math.Max(0, float64(len(vl.items)-visibleItems)))
	vl.offset = int(math.Min(float64(maxOffset), math.Max(0, float64(vl.offset))))
}

// updateVisibleRange calculates which items are currently visible
func (vl *VirtualList) updateVisibleRange() {
	if vl.height == 0 || vl.itemHeight == 0 {
		return
	}
	
	visibleItems := vl.height / vl.itemHeight
	vl.visibleStart = vl.offset
	vl.visibleEnd = int(math.Min(float64(vl.offset+visibleItems-1), float64(len(vl.items)-1)))
}

// selectItem handles item selection
func (vl *VirtualList) selectItem() tea.Cmd {
	if vl.selected >= 0 && vl.selected < len(vl.items) {
		item := vl.items[vl.selected]
		if item.IsSelectable() {
			return func() tea.Msg {
				return ComponentEvent{
					ComponentID: "virtuallist",
					EventType:   "select",
					Data:        item,
				}
			}
		}
	}
	return nil
}

// clearCache clears the render cache
func (vl *VirtualList) clearCache() {
	if vl.cacheEnabled {
		vl.renderCache = make(map[string]string)
	}
}

// SetCacheEnabled enables or disables render caching
func (vl *VirtualList) SetCacheEnabled(enabled bool) {
	vl.cacheEnabled = enabled
	if !enabled {
		vl.clearCache()
	}
}
