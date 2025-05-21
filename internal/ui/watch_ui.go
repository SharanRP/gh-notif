package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
	"github.com/user/gh-notif/internal/watch"
)

// WatchModel represents the watch UI model
type WatchModel struct {
	// Watcher is the notification watcher
	Watcher *watch.Watcher
	// Client is the GitHub client
	Client *githubclient.Client
	// Context is the context for cancellation
	Context context.Context
	// CancelFunc is the function to cancel the context
	CancelFunc context.CancelFunc
	// Table is the notification table
	Table table.Model
	// Spinner is the loading spinner
	Spinner spinner.Model
	// Width is the terminal width
	Width int
	// Height is the terminal height
	Height int
	// Stats are the watch statistics
	Stats watch.WatchStats
	// Notifications are the current notifications
	Notifications []*github.Notification
	// Events are the recent notification events
	Events []watch.NotificationEvent
	// MaxEvents is the maximum number of events to keep
	MaxEvents int
	// Styles are the UI styles
	Styles Styles
	// Error is the current error, if any
	Error error
	// Loading indicates whether the UI is loading
	Loading bool
	// Quitting indicates whether the UI is quitting
	Quitting bool
}

// NewWatchModel creates a new watch UI model
func NewWatchModel(ctx context.Context, client *githubclient.Client, watcher *watch.Watcher) WatchModel {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)

	// Create styles
	theme := DefaultDarkTheme()
	styles := NewStyles(theme)

	// Create a spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	// Create a table
	columns := []table.Column{
		{Title: "ID", Width: 10},
		{Title: "Repository", Width: 30},
		{Title: "Type", Width: 15},
		{Title: "Title", Width: 50},
		{Title: "Updated", Width: 20},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	t.SetStyles(table.Styles{
		Header:   styles.TableHeader,
		Selected: styles.TableSelectedRow,
	})

	// Create the model
	model := WatchModel{
		Watcher:    watcher,
		Client:     client,
		Context:    ctx,
		CancelFunc: cancel,
		Table:      t,
		Spinner:    s,
		MaxEvents:  10,
		Events:     make([]watch.NotificationEvent, 0, 10),
		Styles:     styles,
		Loading:    true,
	}

	// Set up event callback
	watcher.Options.EventCallback = func(event watch.NotificationEvent) {
		// Add the event to the list
		model.Events = append([]watch.NotificationEvent{event}, model.Events...)
		if len(model.Events) > model.MaxEvents {
			model.Events = model.Events[:model.MaxEvents]
		}
	}

	// Set up stats callback
	watcher.Options.StatsCallback = func(stats watch.WatchStats) {
		model.Stats = stats
	}

	return model
}

// Init initializes the model
func (m WatchModel) Init() tea.Cmd {
	// Start the watcher
	if err := m.Watcher.Start(); err != nil {
		return tea.Batch(
			spinner.Tick,
			func() tea.Msg {
				return errMsg{err}
			},
		)
	}

	return tea.Batch(
		spinner.Tick,
		func() tea.Msg {
			// Wait for the first refresh
			time.Sleep(1 * time.Second)
			return refreshMsg{}
		},
	)
}

// Update updates the model
func (m WatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			m.CancelFunc()
			return m, tea.Quit
		case "r":
			// Manual refresh
			m.Loading = true
			return m, func() tea.Msg {
				return refreshMsg{}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Table.SetHeight(m.Height - 15)
		m.Table.SetWidth(m.Width - 4)
		return m, nil

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)

	case refreshMsg:
		// Update the table with the latest notifications
		m.Loading = false
		m.Notifications = m.Watcher.Notifications
		rows := make([]table.Row, len(m.Notifications))
		for i, n := range m.Notifications {
			rows[i] = table.Row{
				n.GetID(),
				n.GetRepository().GetFullName(),
				n.GetSubject().GetType(),
				n.GetSubject().GetTitle(),
				n.GetUpdatedAt().Format(time.RFC3339),
			}
		}
		m.Table.SetRows(rows)

		// Schedule the next refresh
		cmds = append(cmds, func() tea.Msg {
			time.Sleep(1 * time.Second)
			return refreshMsg{}
		})

	case errMsg:
		m.Error = msg.err
		m.Loading = false
		return m, nil
	}

	// Update the table
	m.Table, cmd = m.Table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m WatchModel) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	// Build the view
	var s strings.Builder

	// Title
	s.WriteString(m.Styles.Header.Render("GitHub Notification Watch"))
	s.WriteString("\n\n")

	// Stats
	s.WriteString(m.Styles.DetailHeader.Render("Statistics:"))
	s.WriteString("\n")
	s.WriteString(fmt.Sprintf("Last refresh: %s | Next refresh: %s | New: %d | Updated: %d | Read: %d",
		m.Stats.LastRefreshTime.Format("15:04:05"),
		m.Stats.NextRefreshTime.Format("15:04:05"),
		m.Stats.NewNotificationCount,
		m.Stats.UpdatedNotificationCount,
		m.Stats.ReadNotificationCount))
	s.WriteString("\n\n")

	// Recent events
	s.WriteString(m.Styles.DetailHeader.Render("Recent Events:"))
	s.WriteString("\n")
	if len(m.Events) == 0 {
		s.WriteString(m.Styles.NoNotifications.Render("No events yet."))
	} else {
		for i, event := range m.Events {
			if i >= 5 {
				break
			}
			eventType := string(event.Type)
			switch event.Type {
			case watch.EventNew:
				eventType = lipgloss.NewStyle().Foreground(m.Styles.SuccessColor).Bold(true).Render("NEW")
			case watch.EventUpdated:
				eventType = lipgloss.NewStyle().Foreground(m.Styles.WarningColor).Bold(true).Render("UPDATED")
			case watch.EventRead:
				eventType = lipgloss.NewStyle().Foreground(m.Styles.InfoColor).Bold(true).Render("READ")
			}
			s.WriteString(fmt.Sprintf("%s %s - %s\n",
				eventType,
				event.Notification.GetRepository().GetFullName(),
				event.Notification.GetSubject().GetTitle()))
		}
	}
	s.WriteString("\n")

	// Notifications table
	s.WriteString(m.Styles.DetailHeader.Render("Notifications:"))
	s.WriteString("\n")
	if m.Loading {
		s.WriteString(m.Spinner.View() + " Loading notifications...\n")
	} else if len(m.Notifications) == 0 {
		s.WriteString(m.Styles.NoNotifications.Render("No notifications found."))
	} else {
		s.WriteString(m.Table.View())
	}
	s.WriteString("\n\n")

	// Help
	s.WriteString(m.Styles.HelpBar.Render("Press q to quit, r to refresh manually"))

	return m.Styles.App.Render(s.String())
}

// refreshMsg is a message to refresh the UI
type refreshMsg struct{}

// errMsg is a message containing an error
type errMsg struct {
	err error
}

// RunWatchUI runs the watch UI
func RunWatchUI(ctx context.Context, client *githubclient.Client, watcher *watch.Watcher) error {
	model := NewWatchModel(ctx, client, watcher)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
