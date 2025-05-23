package components

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListComponent wraps the bubbles list component
type ListComponent struct {
	// Configuration
	width   int
	height  int

	// List model
	list    list.Model

	// State
	focused bool

	// Styling
	styles  ComponentStyles
}

// NewListComponent creates a new list component
func NewListComponent(items []list.Item) *ListComponent {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	return &ListComponent{
		list: l,
	}
}

// NewListComponentFactory creates a list component factory
func NewListComponentFactory(config ComponentConfig) Component {
	items, ok := config.Props["items"].([]list.Item)
	if !ok {
		items = []list.Item{}
	}

	listComp := NewListComponent(items)
	listComp.SetSize(config.Width, config.Height)
	listComp.SetStyles(config.Styles)

	return listComp
}

// Init initializes the list component
func (lc *ListComponent) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the list state
func (lc *ListComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				lc.SetSize(size.Width, size.Height)
			}
		case "items":
			if items, ok := msg.Data.([]list.Item); ok {
				lc.list.SetItems(items)
			}
		}
	default:
		if lc.focused {
			lc.list, cmd = lc.list.Update(msg)
		}
	}

	return lc, cmd
}

// View renders the list component
func (lc *ListComponent) View() string {
	return lc.list.View()
}

// SetSize sets the component dimensions
func (lc *ListComponent) SetSize(width, height int) {
	lc.width = width
	lc.height = height
	lc.list.SetSize(width, height)
}

// GetSize returns the component dimensions
func (lc *ListComponent) GetSize() (width, height int) {
	return lc.width, lc.height
}

// SetStyles sets the component styles
func (lc *ListComponent) SetStyles(styles ComponentStyles) {
	lc.styles = styles

	// Note: list.Model doesn't have SetStyles method in this version
	// Styles are applied through the delegate instead
}

// GetType returns the component type
func (lc *ListComponent) GetType() ComponentType {
	return ListComponentType
}

// SetFocused sets the focus state
func (lc *ListComponent) SetFocused(focused bool) {
	lc.focused = focused
}

// IsFocused returns the focus state
func (lc *ListComponent) IsFocused() bool {
	return lc.focused
}

// GetList returns the underlying list model
func (lc *ListComponent) GetList() *list.Model {
	return &lc.list
}

// SetItems sets the list items
func (lc *ListComponent) SetItems(items []list.Item) {
	lc.list.SetItems(items)
}

// GetItems returns the list items
func (lc *ListComponent) GetItems() []list.Item {
	return lc.list.Items()
}

// GetSelectedItem returns the currently selected item
func (lc *ListComponent) GetSelectedItem() list.Item {
	return lc.list.SelectedItem()
}

// GetSelectedIndex returns the currently selected index
func (lc *ListComponent) GetSelectedIndex() int {
	return lc.list.Index()
}

// SetSelectedIndex sets the selected index
func (lc *ListComponent) SetSelectedIndex(index int) {
	lc.list.Select(index)
}

// Layout represents a layout manager for organizing components
type Layout struct {
	// Configuration
	width      int
	height     int
	layoutType LayoutType

	// Components
	components map[string]Component
	order      []string

	// Layout-specific configuration
	splits     []int // For split layouts
	padding    int
	margin     int

	// State
	focused    bool
	focusIndex int

	// Styling
	styles     ComponentStyles
}

// LayoutType represents different layout types
type LayoutType int

const (
	// LayoutVertical arranges components vertically
	LayoutVertical LayoutType = iota
	// LayoutHorizontal arranges components horizontally
	LayoutHorizontal
	// LayoutGrid arranges components in a grid
	LayoutGrid
	// LayoutSplit creates resizable splits
	LayoutSplit
	// LayoutTabs creates tabbed interface
	LayoutTabs
)

// NewLayout creates a new layout manager
func NewLayout(layoutType LayoutType) *Layout {
	return &Layout{
		layoutType: layoutType,
		components: make(map[string]Component),
		order:      []string{},
		padding:    1,
		margin:     0,
	}
}

// AddComponent adds a component to the layout
func (l *Layout) AddComponent(id string, component Component) {
	l.components[id] = component
	l.order = append(l.order, id)
}

// RemoveComponent removes a component from the layout
func (l *Layout) RemoveComponent(id string) {
	delete(l.components, id)

	// Remove from order
	for i, compID := range l.order {
		if compID == id {
			l.order = append(l.order[:i], l.order[i+1:]...)
			break
		}
	}
}

// GetComponent returns a component by ID
func (l *Layout) GetComponent(id string) (Component, bool) {
	comp, exists := l.components[id]
	return comp, exists
}

// SetSplits sets the split ratios for split layouts
func (l *Layout) SetSplits(splits []int) {
	l.splits = splits
}

// SetPadding sets the padding between components
func (l *Layout) SetPadding(padding int) {
	l.padding = padding
}

// SetMargin sets the margin around the layout
func (l *Layout) SetMargin(margin int) {
	l.margin = margin
}

// Init initializes the layout and all components
func (l *Layout) Init() tea.Cmd {
	var cmds []tea.Cmd

	for _, component := range l.components {
		if cmd := component.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return tea.Batch(cmds...)
}

// Update handles messages and updates all components
func (l *Layout) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle focus navigation
		if l.focused {
			switch msg.String() {
			case "tab":
				l.nextFocus()
			case "shift+tab":
				l.prevFocus()
			}
		}

		// Pass message to focused component
		if l.focusIndex >= 0 && l.focusIndex < len(l.order) {
			id := l.order[l.focusIndex]
			if component, exists := l.components[id]; exists {
				updatedComp, cmd := component.Update(msg)
				l.components[id] = updatedComp
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case ComponentMessage:
		if msg.Type == ComponentResizeMsg {
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				l.SetSize(size.Width, size.Height)
			}
		}

		// Pass message to all components
		for id, component := range l.components {
			updatedComp, cmd := component.Update(msg)
			l.components[id] = updatedComp
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	default:
		// Pass message to all components
		for id, component := range l.components {
			updatedComp, cmd := component.Update(msg)
			l.components[id] = updatedComp
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return l, tea.Batch(cmds...)
}

// View renders the layout and all components
func (l *Layout) View() string {
	if len(l.components) == 0 {
		return ""
	}

	switch l.layoutType {
	case LayoutVertical:
		return l.renderVertical()
	case LayoutHorizontal:
		return l.renderHorizontal()
	case LayoutGrid:
		return l.renderGrid()
	case LayoutSplit:
		return l.renderSplit()
	case LayoutTabs:
		return l.renderTabs()
	default:
		return l.renderVertical()
	}
}

// renderVertical renders components vertically
func (l *Layout) renderVertical() string {
	var views []string

	availableHeight := l.height - (l.margin * 2) - (l.padding * (len(l.order) - 1))
	componentHeight := availableHeight / len(l.order)

	for _, id := range l.order {
		if comp, exists := l.components[id]; exists {
			comp.SetSize(l.width-(l.margin*2), componentHeight)
			views = append(views, comp.View())
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, views...)

	if l.margin > 0 {
		style := lipgloss.NewStyle().Padding(l.margin)
		content = style.Render(content)
	}

	return content
}

// renderHorizontal renders components horizontally
func (l *Layout) renderHorizontal() string {
	var views []string

	availableWidth := l.width - (l.margin * 2) - (l.padding * (len(l.order) - 1))
	componentWidth := availableWidth / len(l.order)

	for _, id := range l.order {
		if comp, exists := l.components[id]; exists {
			comp.SetSize(componentWidth, l.height-(l.margin*2))
			views = append(views, comp.View())
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	if l.margin > 0 {
		style := lipgloss.NewStyle().Padding(l.margin)
		content = style.Render(content)
	}

	return content
}

// renderGrid renders components in a grid (simplified)
func (l *Layout) renderGrid() string {
	// For simplicity, render as 2-column grid
	cols := 2
	rows := (len(l.order) + cols - 1) / cols

	componentWidth := (l.width - (l.margin * 2) - l.padding) / cols
	componentHeight := (l.height - (l.margin * 2) - (l.padding * (rows - 1))) / rows

	var rowViews []string

	for row := 0; row < rows; row++ {
		var colViews []string

		for col := 0; col < cols; col++ {
			index := row*cols + col
			if index < len(l.order) {
				id := l.order[index]
				if comp, exists := l.components[id]; exists {
					comp.SetSize(componentWidth, componentHeight)
					colViews = append(colViews, comp.View())
				}
			}
		}

		if len(colViews) > 0 {
			rowViews = append(rowViews, lipgloss.JoinHorizontal(lipgloss.Top, colViews...))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rowViews...)

	if l.margin > 0 {
		style := lipgloss.NewStyle().Padding(l.margin)
		content = style.Render(content)
	}

	return content
}

// renderSplit renders components with splits
func (l *Layout) renderSplit() string {
	if len(l.splits) == 0 {
		return l.renderVertical()
	}

	// Use splits to determine component sizes
	totalSplit := 0
	for _, split := range l.splits {
		totalSplit += split
	}

	var views []string
	availableWidth := l.width - (l.margin * 2)

	for i, id := range l.order {
		if i >= len(l.splits) {
			break
		}

		if comp, exists := l.components[id]; exists {
			splitRatio := float64(l.splits[i]) / float64(totalSplit)
			componentWidth := int(float64(availableWidth) * splitRatio)
			comp.SetSize(componentWidth, l.height-(l.margin*2))
			views = append(views, comp.View())
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, views...)

	if l.margin > 0 {
		style := lipgloss.NewStyle().Padding(l.margin)
		content = style.Render(content)
	}

	return content
}

// renderTabs renders components as tabs (simplified)
func (l *Layout) renderTabs() string {
	if len(l.order) == 0 {
		return ""
	}

	// Render tab headers
	var tabHeaders []string
	for i, id := range l.order {
		if _, exists := l.components[id]; exists {
			style := lipgloss.NewStyle().Padding(0, 1)
			if i == l.focusIndex {
				style = style.Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))
			} else {
				style = style.Background(lipgloss.Color("8")).Foreground(lipgloss.Color("15"))
			}
			tabHeaders = append(tabHeaders, style.Render(id))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabHeaders...)

	// Render active tab content
	var content string
	if l.focusIndex >= 0 && l.focusIndex < len(l.order) {
		id := l.order[l.focusIndex]
		if component, exists := l.components[id]; exists {
			component.SetSize(l.width-(l.margin*2), l.height-(l.margin*2)-2) // -2 for tab bar
			content = component.View()
		}
	}

	result := lipgloss.JoinVertical(lipgloss.Left, tabBar, content)

	if l.margin > 0 {
		style := lipgloss.NewStyle().Padding(l.margin)
		result = style.Render(result)
	}

	return result
}

// SetSize sets the layout dimensions
func (l *Layout) SetSize(width, height int) {
	l.width = width
	l.height = height
}

// GetSize returns the layout dimensions
func (l *Layout) GetSize() (width, height int) {
	return l.width, l.height
}

// SetFocused sets the focus state
func (l *Layout) SetFocused(focused bool) {
	l.focused = focused
	l.updateComponentFocus()
}

// IsFocused returns the focus state
func (l *Layout) IsFocused() bool {
	return l.focused
}

// GetType returns the component type (Layout implements Component interface)
func (l *Layout) GetType() ComponentType {
	return ComponentType("layout")
}

// SetStyles sets the component styles (Layout implements Component interface)
func (l *Layout) SetStyles(styles ComponentStyles) {
	// Layout doesn't use styles directly, but passes them to child components
}

// nextFocus moves focus to the next component
func (l *Layout) nextFocus() {
	if len(l.order) > 0 {
		l.focusIndex = (l.focusIndex + 1) % len(l.order)
		l.updateComponentFocus()
	}
}

// prevFocus moves focus to the previous component
func (l *Layout) prevFocus() {
	if len(l.order) > 0 {
		l.focusIndex--
		if l.focusIndex < 0 {
			l.focusIndex = len(l.order) - 1
		}
		l.updateComponentFocus()
	}
}

// updateComponentFocus updates focus state for all components
func (l *Layout) updateComponentFocus() {
	for i, id := range l.order {
		if component, exists := l.components[id]; exists {
			component.SetFocused(l.focused && i == l.focusIndex)
		}
	}
}
