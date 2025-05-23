package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressType represents different types of progress indicators
type ProgressType int

const (
	// ProgressBar shows a traditional progress bar
	ProgressBar ProgressType = iota
	// ProgressSpinner shows a spinner with text
	ProgressSpinner
	// ProgressCircular shows a circular progress indicator
	ProgressCircular
	// ProgressSteps shows step-by-step progress
	ProgressSteps
)

// ProgressStep represents a step in step-based progress
type ProgressStep struct {
	ID          string
	Title       string
	Description string
	Status      ProgressStepStatus
	Error       error
}

// ProgressStepStatus represents the status of a progress step
type ProgressStepStatus int

const (
	// StepPending indicates the step hasn't started
	StepPending ProgressStepStatus = iota
	// StepInProgress indicates the step is currently running
	StepInProgress
	// StepCompleted indicates the step completed successfully
	StepCompleted
	// StepFailed indicates the step failed
	StepFailed
	// StepSkipped indicates the step was skipped
	StepSkipped
)

// Progress represents a progress indicator component
type Progress struct {
	// Configuration
	width        int
	height       int
	progressType ProgressType

	// Progress state
	value   float64 // 0.0 to 1.0
	total   int64
	current int64

	// Text
	title       string
	description string

	// Steps (for step-based progress)
	steps       []ProgressStep
	currentStep int

	// Animation
	spinner    spinner.Model
	animFrame  int
	lastUpdate time.Time

	// State
	focused   bool
	completed bool
	failed    bool

	// Styling
	styles ComponentStyles

	// Colors for different states
	colors ProgressColors
}

// ProgressColors defines colors for different progress states
type ProgressColors struct {
	Pending    lipgloss.Color
	InProgress lipgloss.Color
	Completed  lipgloss.Color
	Failed     lipgloss.Color
	Skipped    lipgloss.Color
	Background lipgloss.Color
}

// DefaultProgressColors returns default progress colors
func DefaultProgressColors() ProgressColors {
	return ProgressColors{
		Pending:    lipgloss.Color("8"), // Gray
		InProgress: lipgloss.Color("4"), // Blue
		Completed:  lipgloss.Color("2"), // Green
		Failed:     lipgloss.Color("1"), // Red
		Skipped:    lipgloss.Color("3"), // Yellow
		Background: lipgloss.Color("0"), // Black
	}
}

// NewProgress creates a new progress component
func NewProgress(progressType ProgressType) *Progress {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	return &Progress{
		progressType: progressType,
		spinner:      s,
		colors:       DefaultProgressColors(),
		lastUpdate:   time.Now(),
	}
}

// NewProgressComponentFactory creates a progress component factory
func NewProgressComponentFactory(config ComponentConfig) Component {
	progressType, ok := config.Props["type"].(ProgressType)
	if !ok {
		progressType = ProgressBar
	}

	progress := NewProgress(progressType)
	progress.SetSize(config.Width, config.Height)
	progress.SetStyles(config.Styles)

	if title, ok := config.Props["title"].(string); ok {
		progress.SetTitle(title)
	}

	if description, ok := config.Props["description"].(string); ok {
		progress.SetDescription(description)
	}

	return progress
}

// SetValue sets the progress value (0.0 to 1.0)
func (p *Progress) SetValue(value float64) {
	p.value = math.Max(0.0, math.Min(1.0, value))
	p.completed = p.value >= 1.0
}

// SetProgress sets progress with current and total values
func (p *Progress) SetProgress(current, total int64) {
	p.current = current
	p.total = total
	if total > 0 {
		p.SetValue(float64(current) / float64(total))
	}
}

// SetTitle sets the progress title
func (p *Progress) SetTitle(title string) {
	p.title = title
}

// SetDescription sets the progress description
func (p *Progress) SetDescription(description string) {
	p.description = description
}

// AddStep adds a step to step-based progress
func (p *Progress) AddStep(step ProgressStep) {
	p.steps = append(p.steps, step)
}

// UpdateStep updates a step's status
func (p *Progress) UpdateStep(id string, status ProgressStepStatus, err error) {
	for i, step := range p.steps {
		if step.ID == id {
			p.steps[i].Status = status
			p.steps[i].Error = err

			if status == StepInProgress {
				p.currentStep = i
			}
			break
		}
	}
}

// SetFailed marks the progress as failed
func (p *Progress) SetFailed(failed bool) {
	p.failed = failed
}

// Init initializes the progress component
func (p *Progress) Init() tea.Cmd {
	return p.spinner.Tick
}

// Update handles messages and updates the progress state
func (p *Progress) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		p.spinner, cmd = p.spinner.Update(msg)
		cmds = append(cmds, cmd)

		// Update animation frame
		now := time.Now()
		if now.Sub(p.lastUpdate) > 100*time.Millisecond {
			p.animFrame++
			p.lastUpdate = now
		}

	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				p.SetSize(size.Width, size.Height)
			}
		case "progress":
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if value, ok := data["value"].(float64); ok {
					p.SetValue(value)
				}
				if current, ok := data["current"].(int64); ok {
					if total, ok := data["total"].(int64); ok {
						p.SetProgress(current, total)
					}
				}
			}
		case "step":
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if id, ok := data["id"].(string); ok {
					if status, ok := data["status"].(ProgressStepStatus); ok {
						var err error
						if errData, ok := data["error"].(error); ok {
							err = errData
						}
						p.UpdateStep(id, status, err)
					}
				}
			}
		}
	}

	return p, tea.Batch(cmds...)
}

// View renders the progress component
func (p *Progress) View() string {
	switch p.progressType {
	case ProgressBar:
		return p.renderProgressBar()
	case ProgressSpinner:
		return p.renderProgressSpinner()
	case ProgressCircular:
		return p.renderProgressCircular()
	case ProgressSteps:
		return p.renderProgressSteps()
	default:
		return p.renderProgressBar()
	}
}

// renderProgressBar renders a traditional progress bar
func (p *Progress) renderProgressBar() string {
	var parts []string

	// Title
	if p.title != "" {
		titleStyle := p.styles.Focused.Bold(true)
		parts = append(parts, titleStyle.Render(p.title))
	}

	// Progress bar
	barWidth := p.width - 4
	if barWidth < 10 {
		barWidth = 10
	}

	filled := int(p.value * float64(barWidth))

	var bar strings.Builder

	// Filled portion
	filledStyle := lipgloss.NewStyle().Foreground(p.colors.InProgress)
	if p.completed {
		filledStyle = lipgloss.NewStyle().Foreground(p.colors.Completed)
	} else if p.failed {
		filledStyle = lipgloss.NewStyle().Foreground(p.colors.Failed)
	}

	for i := 0; i < filled; i++ {
		bar.WriteString(filledStyle.Render("█"))
	}

	// Empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(p.colors.Background)
	for i := filled; i < barWidth; i++ {
		bar.WriteString(emptyStyle.Render("░"))
	}

	// Add percentage
	percentage := fmt.Sprintf(" %.1f%%", p.value*100)
	barLine := bar.String() + percentage

	// Add current/total if available
	if p.total > 0 {
		barLine += fmt.Sprintf(" (%d/%d)", p.current, p.total)
	}

	parts = append(parts, barLine)

	// Description
	if p.description != "" {
		descStyle := p.styles.Base.Foreground(lipgloss.Color("8"))
		parts = append(parts, descStyle.Render(p.description))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderProgressSpinner renders a spinner with text
func (p *Progress) renderProgressSpinner() string {
	var parts []string

	// Spinner with title
	spinnerLine := p.spinner.View()
	if p.title != "" {
		spinnerLine += " " + p.title
	}

	if p.completed {
		checkStyle := lipgloss.NewStyle().Foreground(p.colors.Completed)
		spinnerLine = checkStyle.Render("✓") + " " + p.title
	} else if p.failed {
		crossStyle := lipgloss.NewStyle().Foreground(p.colors.Failed)
		spinnerLine = crossStyle.Render("✗") + " " + p.title
	}

	parts = append(parts, spinnerLine)

	// Description
	if p.description != "" {
		descStyle := p.styles.Base.Foreground(lipgloss.Color("8"))
		parts = append(parts, "  "+descStyle.Render(p.description))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderProgressCircular renders a circular progress indicator
func (p *Progress) renderProgressCircular() string {
	// Simplified circular progress using Unicode characters
	segments := []string{"◜", "◝", "◞", "◟"}
	segment := segments[p.animFrame%len(segments)]

	if p.completed {
		segment = "●"
	} else if p.failed {
		segment = "✗"
	}

	var parts []string

	// Circular indicator with title
	indicatorStyle := lipgloss.NewStyle().Foreground(p.colors.InProgress)
	if p.completed {
		indicatorStyle = lipgloss.NewStyle().Foreground(p.colors.Completed)
	} else if p.failed {
		indicatorStyle = lipgloss.NewStyle().Foreground(p.colors.Failed)
	}

	line := indicatorStyle.Render(segment)
	if p.title != "" {
		line += " " + p.title
	}

	// Add percentage
	if p.value > 0 {
		line += fmt.Sprintf(" (%.1f%%)", p.value*100)
	}

	parts = append(parts, line)

	// Description
	if p.description != "" {
		descStyle := p.styles.Base.Foreground(lipgloss.Color("8"))
		parts = append(parts, "  "+descStyle.Render(p.description))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// renderProgressSteps renders step-by-step progress
func (p *Progress) renderProgressSteps() string {
	var parts []string

	// Title
	if p.title != "" {
		titleStyle := p.styles.Focused.Bold(true)
		parts = append(parts, titleStyle.Render(p.title))
		parts = append(parts, "")
	}

	// Steps
	for i, step := range p.steps {
		var icon string
		var style lipgloss.Style

		switch step.Status {
		case StepPending:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(p.colors.Pending)
		case StepInProgress:
			icon = p.spinner.View()
			style = lipgloss.NewStyle().Foreground(p.colors.InProgress)
		case StepCompleted:
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(p.colors.Completed)
		case StepFailed:
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(p.colors.Failed)
		case StepSkipped:
			icon = "⊘"
			style = lipgloss.NewStyle().Foreground(p.colors.Skipped)
		}

		line := style.Render(icon) + " " + step.Title
		parts = append(parts, line)

		// Add description if available
		if step.Description != "" {
			descStyle := p.styles.Base.Foreground(lipgloss.Color("8"))
			parts = append(parts, "  "+descStyle.Render(step.Description))
		}

		// Add error if failed
		if step.Status == StepFailed && step.Error != nil {
			errorStyle := lipgloss.NewStyle().Foreground(p.colors.Failed)
			parts = append(parts, "  "+errorStyle.Render("Error: "+step.Error.Error()))
		}

		if i < len(p.steps)-1 {
			parts = append(parts, "")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// SetSize sets the component dimensions
func (p *Progress) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// GetSize returns the component dimensions
func (p *Progress) GetSize() (width, height int) {
	return p.width, p.height
}

// SetStyles sets the component styles
func (p *Progress) SetStyles(styles ComponentStyles) {
	p.styles = styles
}

// GetType returns the component type
func (p *Progress) GetType() ComponentType {
	return ProgressComponentType
}

// SetFocused sets the focus state
func (p *Progress) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns the focus state
func (p *Progress) IsFocused() bool {
	return p.focused
}

// IsCompleted returns whether the progress is completed
func (p *Progress) IsCompleted() bool {
	return p.completed
}

// IsFailed returns whether the progress failed
func (p *Progress) IsFailed() bool {
	return p.failed
}

// GetValue returns the current progress value
func (p *Progress) GetValue() float64 {
	return p.value
}
