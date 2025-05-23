package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
	githubclient "github.com/SharanRP/gh-notif/internal/github"
	"github.com/SharanRP/gh-notif/internal/grouping"
)

// GroupModel represents the group UI model
type GroupModel struct {
	// Client is the GitHub client
	Client *githubclient.Client
	// Context is the context for cancellation
	Context context.Context
	// CancelFunc is the function to cancel the context
	CancelFunc context.CancelFunc
	// GroupTable is the group table
	GroupTable table.Model
	// NotificationTable is the notification table
	NotificationTable table.Model
	// Spinner is the loading spinner
	Spinner spinner.Model
	// Width is the terminal width
	Width int
	// Height is the terminal height
	Height int
	// Notifications are all notifications
	Notifications []*github.Notification
	// Groups are the notification groups
	Groups []*grouping.Group
	// SelectedGroup is the currently selected group
	SelectedGroup *grouping.Group
	// SelectedSubgroup is the currently selected subgroup
	SelectedSubgroup *grouping.Group
	// Grouper is the notification grouper
	Grouper *grouping.Grouper
	// Styles are the UI styles
	Styles Styles
	// Error is the current error, if any
	Error error
	// Loading indicates whether the UI is loading
	Loading bool
	// Quitting indicates whether the UI is quitting
	Quitting bool
	// GroupBy is the primary grouping type
	GroupBy string
	// SecondaryGroupBy is the secondary grouping type
	SecondaryGroupBy string
	// GroupTableFocused indicates whether the group table is focused
	GroupTableFocused bool
}

// NewGroupModel creates a new group UI model
func NewGroupModel(ctx context.Context, client *githubclient.Client, notifications []*github.Notification, groupBy, secondaryGroupBy string) GroupModel {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)

	// Create styles
	theme := DefaultDarkTheme()
	styles := NewStyles(theme)

	// Create a spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	// Create a group table
	groupColumns := []table.Column{
		{Title: "Name", Width: 40},
		{Title: "Count", Width: 10},
		{Title: "Unread", Width: 10},
		{Title: "Type", Width: 15},
	}
	gt := table.New(
		table.WithColumns(groupColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	gt.SetStyles(table.Styles{
		Header:   styles.TableHeader,
		Selected: styles.TableSelectedRow,
	})

	// Create a notification table
	notifColumns := []table.Column{
		{Title: "Repository", Width: 30},
		{Title: "Type", Width: 15},
		{Title: "Title", Width: 50},
		{Title: "Updated", Width: 20},
	}
	nt := table.New(
		table.WithColumns(notifColumns),
		table.WithFocused(false),
		table.WithHeight(10),
	)
	nt.SetStyles(table.Styles{
		Header:   styles.TableHeader,
		Selected: styles.TableSelectedRow,
	})

	// Create a grouper
	options := grouping.DefaultGroupOptions()
	if groupBy != "" {
		options.PrimaryGrouping = parseGroupType(groupBy)
	}
	if secondaryGroupBy != "" {
		options.SecondaryGrouping = parseGroupType(secondaryGroupBy)
	}
	grouper := grouping.NewGrouper(options)

	// Create the model
	model := GroupModel{
		Client:            client,
		Context:           ctx,
		CancelFunc:        cancel,
		GroupTable:        gt,
		NotificationTable: nt,
		Spinner:           s,
		Notifications:     notifications,
		Grouper:           grouper,
		Styles:            styles,
		Loading:           true,
		GroupBy:           groupBy,
		SecondaryGroupBy:  secondaryGroupBy,
		GroupTableFocused: true,
	}

	return model
}

// parseGroupType converts a string to a GroupType
func parseGroupType(groupType string) grouping.GroupType {
	switch strings.ToLower(groupType) {
	case "repository", "repo":
		return grouping.GroupByRepository
	case "owner", "org", "organization":
		return grouping.GroupByOwner
	case "type":
		return grouping.GroupByType
	case "reason":
		return grouping.GroupByReason
	case "thread":
		return grouping.GroupByThread
	case "time":
		return grouping.GroupByTime
	case "score":
		return grouping.GroupByScore
	case "smart":
		return grouping.GroupBySmart
	default:
		return grouping.GroupByRepository
	}
}

// Init initializes the model
func (m GroupModel) Init() tea.Cmd {
	return tea.Batch(
		spinner.Tick,
		func() tea.Msg {
			return groupMsg{
				groupBy:          m.GroupBy,
				secondaryGroupBy: m.SecondaryGroupBy,
			}
		},
	)
}

// Update updates the model
func (m GroupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.Quitting = true
			m.CancelFunc()
			return m, tea.Quit
		case "tab":
			// Toggle focus between tables
			// Note: We need to implement our own focus tracking since table.Model doesn't have SetFocused
			m.GroupTableFocused = !m.GroupTableFocused
			return m, nil
		case "enter":
			// If a group is selected, show its notifications
			if m.GroupTableFocused && len(m.Groups) > 0 {
				selectedRow := m.GroupTable.SelectedRow()
				if selectedRow[0] != "" {
					// Find the selected group
					for _, group := range m.Groups {
						if group.Name == selectedRow[0] {
							m.SelectedGroup = group
							m.SelectedSubgroup = nil

							// Update the notification table
							return m, func() tea.Msg {
								return selectGroupMsg{group: group}
							}
						}
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Adjust table heights
		tableHeight := (m.Height - 20) / 2
		m.GroupTable.SetHeight(tableHeight)
		m.NotificationTable.SetHeight(tableHeight)

		// Adjust table widths
		m.GroupTable.SetWidth(m.Width - 4)
		m.NotificationTable.SetWidth(m.Width - 4)

		return m, nil

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)

	case groupMsg:
		// Group the notifications
		m.Loading = true
		return m, func() tea.Msg {
			groups, err := m.Grouper.Group(m.Context, m.Notifications)
			if err != nil {
				return groupErrMsg{err}
			}
			return groupResultMsg{groups: groups}
		}

	case groupResultMsg:
		// Update the group table
		m.Loading = false
		m.Groups = msg.groups
		rows := make([]table.Row, len(m.Groups))
		for i, group := range m.Groups {
			rows[i] = table.Row{
				group.Name,
				fmt.Sprintf("%d", group.Count),
				fmt.Sprintf("%d", group.UnreadCount),
				string(group.Type),
			}
		}
		m.GroupTable.SetRows(rows)

		// If there are groups, select the first one
		if len(m.Groups) > 0 {
			m.SelectedGroup = m.Groups[0]
			return m, func() tea.Msg {
				return selectGroupMsg{group: m.Groups[0]}
			}
		}

		return m, nil

	case selectGroupMsg:
		// Update the notification table with the selected group's notifications
		m.SelectedGroup = msg.group

		// Update the notification table
		rows := make([]table.Row, len(msg.group.Notifications))
		for i, n := range msg.group.Notifications {
			rows[i] = table.Row{
				n.GetRepository().GetFullName(),
				n.GetSubject().GetType(),
				n.GetSubject().GetTitle(),
				n.GetUpdatedAt().Format(time.RFC3339),
			}
		}
		m.NotificationTable.SetRows(rows)

		return m, nil

	case groupErrMsg:
		m.Error = msg.err
		m.Loading = false
		return m, nil
	}

	// Update the group table
	if m.GroupTableFocused {
		m.GroupTable, cmd = m.GroupTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update the notification table
	if !m.GroupTableFocused {
		m.NotificationTable, cmd = m.NotificationTable.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m GroupModel) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	// Build the view
	var s strings.Builder

	// Title
	s.WriteString(m.Styles.Header.Render("GitHub Notification Groups"))
	s.WriteString("\n\n")

	// Groups
	s.WriteString(m.Styles.DetailHeader.Render("Groups:"))
	s.WriteString("\n")
	if m.Loading {
		s.WriteString(m.Spinner.View() + " Grouping notifications...\n")
	} else if len(m.Groups) == 0 {
		s.WriteString(m.Styles.NoNotifications.Render("No groups found."))
	} else {
		s.WriteString(m.GroupTable.View())
	}
	s.WriteString("\n\n")

	// Selected group notifications
	if m.SelectedGroup != nil {
		s.WriteString(m.Styles.DetailHeader.Render(fmt.Sprintf("Notifications in %s:", m.SelectedGroup.Name)))
		s.WriteString("\n")
		s.WriteString(m.NotificationTable.View())
	}
	s.WriteString("\n\n")

	// Error
	if m.Error != nil {
		s.WriteString(m.Styles.Error.Render(fmt.Sprintf("Error: %v", m.Error)))
		s.WriteString("\n\n")
	}

	// Help
	s.WriteString(m.Styles.HelpBar.Render("Press Tab to switch focus, Enter to select a group, Esc/q to quit"))

	return m.Styles.App.Render(s.String())
}

// groupMsg is a message to group notifications
type groupMsg struct {
	groupBy          string
	secondaryGroupBy string
}

// groupResultMsg is a message containing grouped notifications
type groupResultMsg struct {
	groups []*grouping.Group
}

// selectGroupMsg is a message to select a group
type selectGroupMsg struct {
	group *grouping.Group
}

// groupErrMsg is a message containing an error
type groupErrMsg struct {
	err error
}

// RunGroupUI runs the group UI
func RunGroupUI(ctx context.Context, client *githubclient.Client, notifications []*github.Notification, groupBy, secondaryGroupBy string) error {
	model := NewGroupModel(ctx, client, notifications, groupBy, secondaryGroupBy)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
