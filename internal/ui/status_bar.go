package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// NewStatusBar creates a new status bar
func NewStatusBar(text string) StatusBar {
	return StatusBar{
		text:  text,
		style: lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
	}
}

// SetText sets the text of the status bar
func (s *StatusBar) SetText(text string) {
	s.text = text
}

// SetStyle sets the style of the status bar
func (s *StatusBar) SetStyle(style lipgloss.Style) {
	s.style = style
}

// View renders the status bar
func (s StatusBar) View() string {
	return s.style.Render(s.text)
}
