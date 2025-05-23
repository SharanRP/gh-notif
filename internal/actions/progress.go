package actions

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/SharanRP/gh-notif/internal/common"
)

// ProgressReporter provides progress reporting for batch operations
type ProgressReporter struct {
	// Writer is where progress is written
	Writer io.Writer
	// Total is the total number of operations
	Total int
	// Completed is the number of completed operations
	Completed int
	// Errors is the number of errors
	Errors int
	// StartTime is when the operation started
	StartTime time.Time
	// UseSpinner determines whether to use a spinner
	UseSpinner bool
	// UseProgressBar determines whether to use a progress bar
	UseProgressBar bool
	// ShowPercentage determines whether to show percentage
	ShowPercentage bool
	// ShowCount determines whether to show count
	ShowCount bool
	// ShowElapsed determines whether to show elapsed time
	ShowElapsed bool
	// ShowETA determines whether to show estimated time remaining
	ShowETA bool
	// mu protects the fields
	mu sync.RWMutex
}

// NewProgressReporter creates a new progress reporter
func NewProgressReporter(total int) *ProgressReporter {
	return &ProgressReporter{
		Writer:         os.Stdout,
		Total:          total,
		Completed:      0,
		Errors:         0,
		StartTime:      time.Now(),
		UseSpinner:     true,
		UseProgressBar: true,
		ShowPercentage: true,
		ShowCount:      true,
		ShowElapsed:    true,
		ShowETA:        true,
	}
}

// Update updates the progress
func (p *ProgressReporter) Update(completed, errors int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Completed = completed
	p.Errors = errors
}

// GetProgressCallback returns a callback for updating progress
func (p *ProgressReporter) GetProgressCallback() func(completed, total int) {
	return func(completed, total int) {
		p.Update(completed, 0)
		p.Report()
	}
}

// GetErrorCallback returns a callback for handling errors
func (p *ProgressReporter) GetErrorCallback() func(notificationID string, err error) {
	return func(notificationID string, err error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		p.Errors++
	}
}

// Report reports the current progress
func (p *ProgressReporter) Report() {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Calculate percentage
	percentage := float64(p.Completed) / float64(p.Total) * 100

	// Calculate elapsed time
	elapsed := time.Since(p.StartTime)

	// Calculate ETA
	var eta time.Duration
	if p.Completed > 0 {
		eta = time.Duration(float64(elapsed) / float64(p.Completed) * float64(p.Total-p.Completed))
	}

	// Build the progress string
	var parts []string

	// Add count
	if p.ShowCount {
		parts = append(parts, fmt.Sprintf("%d/%d", p.Completed, p.Total))
	}

	// Add percentage
	if p.ShowPercentage {
		parts = append(parts, fmt.Sprintf("%.1f%%", percentage))
	}

	// Add elapsed time
	if p.ShowElapsed {
		parts = append(parts, fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)))
	}

	// Add ETA
	if p.ShowETA && p.Completed > 0 {
		parts = append(parts, fmt.Sprintf("ETA: %s", formatDuration(eta)))
	}

	// Add errors
	if p.Errors > 0 {
		parts = append(parts, fmt.Sprintf("Errors: %d", p.Errors))
	}

	// Join the parts
	progressStr := strings.Join(parts, " | ")

	// Print the progress
	fmt.Fprintf(p.Writer, "\r%s", progressStr)

	// Print a newline if complete
	if p.Completed >= p.Total {
		fmt.Fprintln(p.Writer)
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	return common.FormatDuration(d)
}

// ProgressModel is a Bubble Tea model for displaying progress
type ProgressModel struct {
	// Progress is the progress bar
	Progress progress.Model
	// Total is the total number of operations
	Total int
	// Completed is the number of completed operations
	Completed int
	// Errors is the number of errors
	Errors int
	// StartTime is when the operation started
	StartTime time.Time
	// Width is the width of the progress bar
	Width int
	// Description is the description of the operation
	Description string
	// Finished indicates whether the operation is finished
	Finished bool
}

// NewProgressModel creates a new progress model
func NewProgressModel(total int, description string) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
		progress.WithoutPercentage(),
	)
	return ProgressModel{
		Progress:    p,
		Total:       total,
		Completed:   0,
		Errors:      0,
		StartTime:   time.Now(),
		Width:       80,
		Description: description,
		Finished:    false,
	}
}

// Init initializes the model
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update updates the model
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case ProgressUpdateMsg:
		m.Completed = msg.Completed
		m.Errors = msg.Errors
		percent := float64(m.Completed) / float64(m.Total)

		// Check if we're done
		if m.Completed >= m.Total {
			m.Finished = true
			return m, tea.Quit
		}

		return m, m.Progress.SetPercent(percent)

	default:
		return m, nil
	}
}

// View renders the model
func (m ProgressModel) View() string {
	// Calculate elapsed time
	elapsed := time.Since(m.StartTime)

	// Calculate ETA
	var eta time.Duration
	if m.Completed > 0 {
		eta = time.Duration(float64(elapsed) / float64(m.Completed) * float64(m.Total-m.Completed))
	}

	// Build the view
	var s strings.Builder

	// Add description
	s.WriteString(m.Description)
	s.WriteString("\n\n")

	// Add progress bar
	s.WriteString(m.Progress.View())
	s.WriteString("\n\n")

	// Add stats
	s.WriteString(fmt.Sprintf("%d/%d complete", m.Completed, m.Total))
	if m.Errors > 0 {
		s.WriteString(fmt.Sprintf(" (%d errors)", m.Errors))
	}
	s.WriteString("\n")

	// Add timing info
	s.WriteString(fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)))
	if !m.Finished && m.Completed > 0 {
		s.WriteString(fmt.Sprintf(" | ETA: %s", formatDuration(eta)))
	}

	return lipgloss.NewStyle().Width(m.Width).Render(s.String())
}

// ProgressUpdateMsg is a message for updating progress
type ProgressUpdateMsg struct {
	Completed int
	Errors    int
}

// RunProgressUI runs a progress UI
func RunProgressUI(total int, description string, updateCh <-chan ProgressUpdateMsg) error {
	model := NewProgressModel(total, description)
	p := tea.NewProgram(model)

	// Start a goroutine to send updates to the model
	go func() {
		for update := range updateCh {
			p.Send(update)
		}
	}()

	_, err := p.Run()
	return err
}
