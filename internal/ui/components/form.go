package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FormField represents a field in a form
type FormField interface {
	// GetID returns the field ID
	GetID() string

	// GetLabel returns the field label
	GetLabel() string

	// GetValue returns the current value
	GetValue() interface{}

	// SetValue sets the field value
	SetValue(value interface{}) error

	// Validate validates the field value
	Validate() error

	// Render renders the field
	Render(width int, focused bool, styles ComponentStyles) string

	// Update handles input messages
	Update(tea.Msg) (FormField, tea.Cmd)

	// IsRequired returns whether the field is required
	IsRequired() bool

	// GetHelp returns help text for the field
	GetHelp() string
}

// TextInputField implements a text input form field
type TextInputField struct {
	id          string
	label       string
	input       textinput.Model
	required    bool
	validator   func(string) error
	help        string
	placeholder string
}

// NewTextInputField creates a new text input field
func NewTextInputField(id, label string) *TextInputField {
	ti := textinput.New()
	ti.CharLimit = 256

	return &TextInputField{
		id:    id,
		label: label,
		input: ti,
	}
}

// GetID returns the field ID
func (f *TextInputField) GetID() string {
	return f.id
}

// GetLabel returns the field label
func (f *TextInputField) GetLabel() string {
	return f.label
}

// GetValue returns the current value
func (f *TextInputField) GetValue() interface{} {
	return f.input.Value()
}

// SetValue sets the field value
func (f *TextInputField) SetValue(value interface{}) error {
	if str, ok := value.(string); ok {
		f.input.SetValue(str)
		return nil
	}
	return fmt.Errorf("invalid value type for text field")
}

// Validate validates the field value
func (f *TextInputField) Validate() error {
	value := f.input.Value()

	if f.required && strings.TrimSpace(value) == "" {
		return fmt.Errorf("field %s is required", f.label)
	}

	if f.validator != nil {
		return f.validator(value)
	}

	return nil
}

// Render renders the field
func (f *TextInputField) Render(width int, focused bool, styles ComponentStyles) string {
	var parts []string

	// Render label
	labelStyle := styles.Base
	if f.required {
		labelStyle = labelStyle.Foreground(lipgloss.Color("1")) // Red for required
	}
	parts = append(parts, labelStyle.Render(f.label+":"))

	// Validate and show error if any
	if err := f.Validate(); err != nil {
		parts = append(parts, f.input.View())
		parts = append(parts, styles.Error.Render("Error: "+err.Error()))
	} else {
		parts = append(parts, f.input.View())
	}

	// Add help text if available
	if f.help != "" {
		helpStyle := styles.Base.Foreground(lipgloss.Color("8")) // Dim
		parts = append(parts, helpStyle.Render(f.help))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// Update handles input messages
func (f *TextInputField) Update(msg tea.Msg) (FormField, tea.Cmd) {
	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	return f, cmd
}

// IsRequired returns whether the field is required
func (f *TextInputField) IsRequired() bool {
	return f.required
}

// GetHelp returns help text for the field
func (f *TextInputField) GetHelp() string {
	return f.help
}

// SetRequired sets whether the field is required
func (f *TextInputField) SetRequired(required bool) *TextInputField {
	f.required = required
	return f
}

// SetValidator sets a custom validator function
func (f *TextInputField) SetValidator(validator func(string) error) *TextInputField {
	f.validator = validator
	return f
}

// SetHelp sets help text for the field
func (f *TextInputField) SetHelp(help string) *TextInputField {
	f.help = help
	return f
}

// SetPlaceholder sets placeholder text
func (f *TextInputField) SetPlaceholder(placeholder string) *TextInputField {
	f.placeholder = placeholder
	f.input.Placeholder = placeholder
	return f
}

// Focus focuses the input field
func (f *TextInputField) Focus() tea.Cmd {
	return f.input.Focus()
}

// Blur removes focus from the input field
func (f *TextInputField) Blur() {
	f.input.Blur()
}

// Form represents an interactive form component
type Form struct {
	// Configuration
	width   int
	height  int
	title   string

	// Fields
	fields       []FormField
	currentField int

	// State
	focused      bool
	submitted    bool
	cancelled    bool

	// Styling
	styles       ComponentStyles

	// Key bindings
	keyMap       FormKeyMap
}

// FormKeyMap defines key bindings for forms
type FormKeyMap struct {
	NextField     key.Binding
	PrevField     key.Binding
	Submit        key.Binding
	Cancel        key.Binding
	Help          key.Binding
}

// DefaultFormKeyMap returns default form key bindings
func DefaultFormKeyMap() FormKeyMap {
	return FormKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("tab", "down"),
			key.WithHelp("tab/↓", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("shift+tab", "up"),
			key.WithHelp("shift+tab/↑", "previous field"),
		),
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s", "ctrl+enter"),
			key.WithHelp("ctrl+s", "submit"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// NewForm creates a new form component
func NewForm(title string) *Form {
	return &Form{
		title:        title,
		fields:       []FormField{},
		currentField: 0,
		keyMap:       DefaultFormKeyMap(),
	}
}

// NewFormComponentFactory creates a form component factory
func NewFormComponentFactory(config ComponentConfig) Component {
	title, ok := config.Props["title"].(string)
	if !ok {
		title = "Form"
	}

	form := NewForm(title)
	form.SetSize(config.Width, config.Height)
	form.SetStyles(config.Styles)

	// Add fields from config if provided
	if fields, ok := config.Props["fields"].([]FormField); ok {
		for _, field := range fields {
			form.AddField(field)
		}
	}

	return form
}

// AddField adds a field to the form
func (f *Form) AddField(field FormField) *Form {
	f.fields = append(f.fields, field)
	return f
}

// GetField returns a field by ID
func (f *Form) GetField(id string) FormField {
	for _, field := range f.fields {
		if field.GetID() == id {
			return field
		}
	}
	return nil
}

// GetValues returns all field values as a map
func (f *Form) GetValues() map[string]interface{} {
	values := make(map[string]interface{})
	for _, field := range f.fields {
		values[field.GetID()] = field.GetValue()
	}
	return values
}

// Validate validates all form fields
func (f *Form) Validate() []error {
	var errors []error
	for _, field := range f.fields {
		if err := field.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// IsValid returns whether the form is valid
func (f *Form) IsValid() bool {
	return len(f.Validate()) == 0
}

// Init initializes the form
func (f *Form) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Focus the first field if available
	if len(f.fields) > 0 {
		if textField, ok := f.fields[0].(*TextInputField); ok {
			cmds = append(cmds, textField.Focus())
		}
	}

	return tea.Batch(cmds...)
}

// Update handles messages and updates the form state
func (f *Form) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !f.focused {
			return f, nil
		}

		switch {
		case key.Matches(msg, f.keyMap.NextField):
			f.nextField()
			return f, nil

		case key.Matches(msg, f.keyMap.PrevField):
			f.prevField()
			return f, nil

		case key.Matches(msg, f.keyMap.Submit):
			if f.IsValid() {
				f.submitted = true
				return f, func() tea.Msg {
					return ComponentEvent{
						ComponentID: "form",
						EventType:   "submit",
						Data:        f.GetValues(),
					}
				}
			}
			return f, nil

		case key.Matches(msg, f.keyMap.Cancel):
			f.cancelled = true
			return f, func() tea.Msg {
				return ComponentEvent{
					ComponentID: "form",
					EventType:   "cancel",
					Data:        nil,
				}
			}
		}

		// Pass message to current field
		if f.currentField >= 0 && f.currentField < len(f.fields) {
			var cmd tea.Cmd
			f.fields[f.currentField], cmd = f.fields[f.currentField].Update(msg)
			cmds = append(cmds, cmd)
		}

	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				f.SetSize(size.Width, size.Height)
			}
		}
	}

	return f, tea.Batch(cmds...)
}

// View renders the form
func (f *Form) View() string {
	var parts []string

	// Render title
	if f.title != "" {
		titleStyle := f.styles.Focused.Bold(true)
		parts = append(parts, titleStyle.Render(f.title))
		parts = append(parts, "")
	}

	// Render fields
	for i, field := range f.fields {
		focused := f.focused && i == f.currentField
		fieldView := field.Render(f.width-4, focused, f.styles)
		parts = append(parts, fieldView)

		if i < len(f.fields)-1 {
			parts = append(parts, "")
		}
	}

	// Render validation errors
	if errors := f.Validate(); len(errors) > 0 && f.focused {
		parts = append(parts, "")
		errorStyle := f.styles.Error
		for _, err := range errors {
			parts = append(parts, errorStyle.Render("• "+err.Error()))
		}
	}

	// Render help
	if f.focused {
		parts = append(parts, "")
		helpStyle := f.styles.Base.Foreground(lipgloss.Color("8"))
		help := []string{
			"Tab/↓: Next field",
			"Shift+Tab/↑: Previous field",
			"Ctrl+S: Submit",
			"Esc: Cancel",
		}
		parts = append(parts, helpStyle.Render(strings.Join(help, " • ")))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, parts...)

	// Apply container style
	containerStyle := f.styles.Base.
		Width(f.width).
		Height(f.height).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder())

	if f.focused {
		containerStyle = containerStyle.BorderForeground(lipgloss.Color("4"))
	} else {
		containerStyle = containerStyle.BorderForeground(lipgloss.Color("8"))
	}

	return containerStyle.Render(content)
}

// SetSize sets the component dimensions
func (f *Form) SetSize(width, height int) {
	f.width = width
	f.height = height
}

// GetSize returns the component dimensions
func (f *Form) GetSize() (width, height int) {
	return f.width, f.height
}

// SetStyles sets the component styles
func (f *Form) SetStyles(styles ComponentStyles) {
	f.styles = styles
}

// GetType returns the component type
func (f *Form) GetType() ComponentType {
	return FormComponentType
}

// SetFocused sets the focus state
func (f *Form) SetFocused(focused bool) {
	f.focused = focused

	// Focus/blur current field
	if f.currentField >= 0 && f.currentField < len(f.fields) {
		if textField, ok := f.fields[f.currentField].(*TextInputField); ok {
			if focused {
				textField.Focus()
			} else {
				textField.Blur()
			}
		}
	}
}

// IsFocused returns the focus state
func (f *Form) IsFocused() bool {
	return f.focused
}

// nextField moves to the next field
func (f *Form) nextField() {
	if len(f.fields) == 0 {
		return
	}

	// Blur current field
	if f.currentField >= 0 && f.currentField < len(f.fields) {
		if textField, ok := f.fields[f.currentField].(*TextInputField); ok {
			textField.Blur()
		}
	}

	// Move to next field
	f.currentField = (f.currentField + 1) % len(f.fields)

	// Focus new field
	if textField, ok := f.fields[f.currentField].(*TextInputField); ok {
		textField.Focus()
	}
}

// prevField moves to the previous field
func (f *Form) prevField() {
	if len(f.fields) == 0 {
		return
	}

	// Blur current field
	if f.currentField >= 0 && f.currentField < len(f.fields) {
		if textField, ok := f.fields[f.currentField].(*TextInputField); ok {
			textField.Blur()
		}
	}

	// Move to previous field
	f.currentField--
	if f.currentField < 0 {
		f.currentField = len(f.fields) - 1
	}

	// Focus new field
	if textField, ok := f.fields[f.currentField].(*TextInputField); ok {
		textField.Focus()
	}
}
