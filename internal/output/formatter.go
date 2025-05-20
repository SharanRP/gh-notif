package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

// Format represents the output format
type Format string

const (
	// FormatText outputs human-readable text
	FormatText Format = "text"
	// FormatJSON outputs JSON
	FormatJSON Format = "json"
	// FormatCSV outputs CSV
	FormatCSV Format = "csv"
	// FormatTemplate outputs using a custom template
	FormatTemplate Format = "template"
)

// Formatter formats notifications for output
type Formatter struct {
	// OutputFormat is the output format
	OutputFormat Format
	// Writer is the output writer
	Writer io.Writer
	// Template is the custom template for template format
	Template string
	// NoColor disables color output
	NoColor bool
	// Verbose enables verbose output
	Verbose bool
	// Fields specifies which fields to include in the output
	Fields []string
	// TemplateCache caches parsed templates
	TemplateCache map[string]*template.Template
}

// NewFormatter creates a new formatter
func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{
		OutputFormat:  FormatText,
		Writer:        w,
		NoColor:       false,
		Verbose:       false,
		Fields:        []string{"id", "repository", "type", "title", "updated", "status"},
		TemplateCache: make(map[string]*template.Template),
	}
}

// WithFormat sets the output format
func (f *Formatter) WithFormat(format Format) *Formatter {
	f.OutputFormat = format
	return f
}

// WithTemplate sets the custom template
func (f *Formatter) WithTemplate(tmpl string) *Formatter {
	f.Template = tmpl
	f.OutputFormat = FormatTemplate
	return f
}

// WithNoColor disables color output
func (f *Formatter) WithNoColor(noColor bool) *Formatter {
	f.NoColor = noColor
	return f
}

// WithVerbose enables verbose output
func (f *Formatter) WithVerbose(verbose bool) *Formatter {
	f.Verbose = verbose
	return f
}

// WithFields sets the fields to include in the output
func (f *Formatter) WithFields(fields []string) *Formatter {
	if len(fields) > 0 {
		f.Fields = fields
	}
	return f
}

// Format formats notifications for output
func (f *Formatter) Format(notifications []*github.Notification) error {
	switch f.OutputFormat {
	case FormatText:
		return f.formatText(notifications)
	case FormatJSON:
		return f.formatJSON(notifications)
	case FormatCSV:
		return f.formatCSV(notifications)
	case FormatTemplate:
		return f.formatTemplate(notifications)
	default:
		return fmt.Errorf("unsupported format: %s", f.OutputFormat)
	}
}

// formatText formats notifications as human-readable text
func (f *Formatter) formatText(notifications []*github.Notification) error {
	if len(notifications) == 0 {
		fmt.Fprintln(f.Writer, "No notifications found.")
		return nil
	}

	// Define styles
	var (
		headerStyle  lipgloss.Style
		unreadStyle  lipgloss.Style
		readStyle    lipgloss.Style
		repoStyle    lipgloss.Style
		typeStyle    lipgloss.Style
		timeStyle    lipgloss.Style
		reasonStyle  lipgloss.Style
		dividerStyle lipgloss.Style
	)

	if !f.NoColor {
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
		unreadStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
		readStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		repoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
		typeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
		timeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
		reasonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
		dividerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	}

	// Print header
	header := []string{}
	for _, field := range f.Fields {
		var headerText string
		switch strings.ToLower(field) {
		case "id":
			headerText = "ID"
		case "repository", "repo":
			headerText = "Repository"
		case "type":
			headerText = "Type"
		case "title":
			headerText = "Title"
		case "updated":
			headerText = "Updated"
		case "status":
			headerText = "Status"
		case "reason":
			headerText = "Reason"
		default:
			headerText = strings.Title(field)
		}
		if !f.NoColor {
			headerText = headerStyle.Render(headerText)
		}
		header = append(header, headerText)
	}
	fmt.Fprintln(f.Writer, strings.Join(header, " | "))

	// Print divider
	divider := strings.Repeat("-", 80)
	if !f.NoColor {
		divider = dividerStyle.Render(divider)
	}
	fmt.Fprintln(f.Writer, divider)

	// Print notifications
	for _, n := range notifications {
		row := []string{}
		for _, field := range f.Fields {
			var value string
			switch strings.ToLower(field) {
			case "id":
				value = n.GetID()
			case "repository", "repo":
				value = n.GetRepository().GetFullName()
				if !f.NoColor {
					value = repoStyle.Render(value)
				}
			case "type":
				value = n.GetSubject().GetType()
				if !f.NoColor {
					value = typeStyle.Render(value)
				}
			case "title":
				value = n.GetSubject().GetTitle()
				if !f.NoColor {
					if n.GetUnread() {
						value = unreadStyle.Render(value)
					} else {
						value = readStyle.Render(value)
					}
				}
			case "updated":
				value = formatTime(n.GetUpdatedAt().Time)
				if !f.NoColor {
					value = timeStyle.Render(value)
				}
			case "status":
				if n.GetUnread() {
					value = "Unread"
					if !f.NoColor {
						value = unreadStyle.Render(value)
					}
				} else {
					value = "Read"
					if !f.NoColor {
						value = readStyle.Render(value)
					}
				}
			case "reason":
				value = n.GetReason()
				if !f.NoColor {
					value = reasonStyle.Render(value)
				}
			default:
				value = "N/A"
			}
			row = append(row, value)
		}
		fmt.Fprintln(f.Writer, strings.Join(row, " | "))
	}

	return nil
}

// formatJSON formats notifications as JSON
func (f *Formatter) formatJSON(notifications []*github.Notification) error {
	// Create a custom representation for JSON output
	type jsonNotification struct {
		ID         string    `json:"id"`
		Repository string    `json:"repository"`
		Type       string    `json:"type"`
		Title      string    `json:"title"`
		URL        string    `json:"url"`
		UpdatedAt  time.Time `json:"updated_at"`
		Unread     bool      `json:"unread"`
		Reason     string    `json:"reason"`
	}

	jsonNotifications := make([]jsonNotification, 0, len(notifications))
	for _, n := range notifications {
		jsonNotifications = append(jsonNotifications, jsonNotification{
			ID:         n.GetID(),
			Repository: n.GetRepository().GetFullName(),
			Type:       n.GetSubject().GetType(),
			Title:      n.GetSubject().GetTitle(),
			URL:        n.GetSubject().GetURL(),
			UpdatedAt:  n.GetUpdatedAt().Time,
			Unread:     n.GetUnread(),
			Reason:     n.GetReason(),
		})
	}

	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonNotifications)
}

// formatCSV formats notifications as CSV
func (f *Formatter) formatCSV(notifications []*github.Notification) error {
	writer := csv.NewWriter(f.Writer)
	defer writer.Flush()

	// Write header
	header := make([]string, 0, len(f.Fields))
	for _, field := range f.Fields {
		switch strings.ToLower(field) {
		case "id":
			header = append(header, "ID")
		case "repository", "repo":
			header = append(header, "Repository")
		case "type":
			header = append(header, "Type")
		case "title":
			header = append(header, "Title")
		case "updated":
			header = append(header, "Updated")
		case "status":
			header = append(header, "Status")
		case "reason":
			header = append(header, "Reason")
		default:
			header = append(header, strings.Title(field))
		}
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write notifications
	for _, n := range notifications {
		row := make([]string, 0, len(f.Fields))
		for _, field := range f.Fields {
			var value string
			switch strings.ToLower(field) {
			case "id":
				value = n.GetID()
			case "repository", "repo":
				value = n.GetRepository().GetFullName()
			case "type":
				value = n.GetSubject().GetType()
			case "title":
				value = n.GetSubject().GetTitle()
			case "updated":
				value = n.GetUpdatedAt().Time.Format(time.RFC3339)
			case "status":
				if n.GetUnread() {
					value = "Unread"
				} else {
					value = "Read"
				}
			case "reason":
				value = n.GetReason()
			default:
				value = "N/A"
			}
			row = append(row, value)
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// formatTemplate formats notifications using a custom template
func (f *Formatter) formatTemplate(notifications []*github.Notification) error {
	if f.Template == "" {
		return fmt.Errorf("template not specified")
	}

	// Check if template is already parsed
	tmpl, ok := f.TemplateCache[f.Template]
	if !ok {
		// Parse template
		var err error
		tmpl, err = template.New("notifications").Funcs(template.FuncMap{
			"formatTime": formatTime,
		}).Parse(f.Template)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}
		f.TemplateCache[f.Template] = tmpl
	}

	// Execute template
	return tmpl.Execute(f.Writer, notifications)
}

// formatTime formats a time.Time into a human-readable string
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}

	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
