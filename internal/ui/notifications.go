package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
)

var (
	baseStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))

	unreadStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	readStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

// NotificationModel represents the UI model for notifications
type NotificationModel struct {
	table  table.Model
	width  int
	height int
	notifications []*github.Notification
}

// DisplayNotifications shows the notifications in a terminal UI
func DisplayNotifications(notifications []*github.Notification) error {
	if len(notifications) == 0 {
		fmt.Println("No notifications found.")
		return nil
	}

	// Create table columns
	columns := []table.Column{
		{Title: "ID", Width: 5},
		{Title: "Type", Width: 10},
		{Title: "Repository", Width: 30},
		{Title: "Title", Width: 50},
		{Title: "Updated", Width: 20},
	}

	// Create table rows
	rows := []table.Row{}
	for i, n := range notifications {
		id := fmt.Sprintf("%d", i+1)
		notifType := n.GetSubject().GetType()
		repo := n.GetRepository().GetFullName()
		title := n.GetSubject().GetTitle()
		updated := n.GetUpdatedAt().Format(time.RFC822)

		rows = append(rows, table.Row{id, notifType, repo, title, updated})
	}

	// Create table
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)),
	)

	// Style the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	// Create model
	m := NotificationModel{
		table: t,
		notifications: notifications,
	}

	// Run the UI
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running UI: %w", err)
	}

	return nil
}

// Init initializes the model
func (m NotificationModel) Init() tea.Cmd {
	return nil
}

// Update handles UI events
func (m NotificationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			// Handle selection - open notification in browser or mark as read
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.table.SetHeight(m.height - 4)
		m.table.SetWidth(m.width)
		return m, nil
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// View renders the UI
func (m NotificationModel) View() string {
	if len(m.notifications) == 0 {
		return "No notifications found."
	}

	header := headerStyle.Render("GitHub Notifications")
	help := "↑/↓: navigate • enter: select • q: quit"

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		header,
		m.table.View(),
		help,
	)
}
