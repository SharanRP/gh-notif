package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v60/github"
	githubclient "github.com/SharanRP/gh-notif/internal/github"
	"github.com/SharanRP/gh-notif/internal/search"
)

// SearchModel represents the search UI model
type SearchModel struct {
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
	// SearchInput is the search input field
	SearchInput textinput.Model
	// Width is the terminal width
	Width int
	// Height is the terminal height
	Height int
	// Notifications are all notifications
	Notifications []*github.Notification
	// Results are the search results
	Results []*search.SearchResult
	// Searcher is the notification searcher
	Searcher *search.Searcher
	// Styles are the UI styles
	Styles Styles
	// Error is the current error, if any
	Error error
	// Loading indicates whether the UI is loading
	Loading bool
	// Quitting indicates whether the UI is quitting
	Quitting bool
}

// NewSearchModel creates a new search UI model
func NewSearchModel(ctx context.Context, client *githubclient.Client, notifications []*github.Notification, initialQuery string) SearchModel {
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
		{Title: "Score", Width: 10},
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

	// Create a search input
	ti := textinput.New()
	ti.Placeholder = "Type to search notifications..."
	ti.Focus()
	ti.Width = 80
	ti.Prompt = "Search: "
	ti.PromptStyle = styles.FilterPrompt
	ti.TextStyle = styles.FilterInput
	ti.SetValue(initialQuery)

	// Create a searcher
	options := search.DefaultSearchOptions()
	searcher := search.NewSearcher(options)

	// Create the model
	model := SearchModel{
		Client:        client,
		Context:       ctx,
		CancelFunc:    cancel,
		Table:         t,
		Spinner:       s,
		SearchInput:   ti,
		Notifications: notifications,
		Searcher:      searcher,
		Styles:        styles,
		Loading:       true,
	}

	return model
}

// Init initializes the model
func (m SearchModel) Init() tea.Cmd {
	// If there's an initial query, search for it
	if m.SearchInput.Value() != "" {
		return tea.Batch(
			spinner.Tick,
			textinput.Blink,
			func() tea.Msg {
				return searchMsg{query: m.SearchInput.Value()}
			},
		)
	}

	return tea.Batch(
		spinner.Tick,
		textinput.Blink,
	)
}

// Update updates the model
func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.Quitting = true
			m.CancelFunc()
			return m, tea.Quit
		case "enter":
			// Search for the query
			m.Loading = true
			return m, func() tea.Msg {
				return searchMsg{query: m.SearchInput.Value()}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Table.SetHeight(m.Height - 15)
		m.Table.SetWidth(m.Width - 4)
		m.SearchInput.Width = m.Width - 20
		return m, nil

	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		cmds = append(cmds, spinnerCmd)

	case searchMsg:
		// Search for the query
		m.Loading = true
		return m, func() tea.Msg {
			results, err := m.Searcher.Search(m.Context, m.Notifications, msg.query)
			if err != nil {
				return searchErrMsg{err}
			}
			return searchResultMsg{results: results}
		}

	case searchResultMsg:
		// Update the table with the search results
		m.Loading = false
		m.Results = msg.results
		rows := make([]table.Row, len(m.Results))
		for i, result := range m.Results {
			n := result.Notification
			rows[i] = table.Row{
				fmt.Sprintf("%.1f", result.Score),
				n.GetRepository().GetFullName(),
				n.GetSubject().GetType(),
				n.GetSubject().GetTitle(),
				n.GetUpdatedAt().Format(time.RFC3339),
			}
		}
		m.Table.SetRows(rows)
		return m, nil

	case searchErrMsg:
		m.Error = msg.err
		m.Loading = false
		return m, nil
	}

	// Update the search input
	var searchInputCmd tea.Cmd
	m.SearchInput, searchInputCmd = m.SearchInput.Update(msg)
	cmds = append(cmds, searchInputCmd)

	// Update the table
	m.Table, cmd = m.Table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the model
func (m SearchModel) View() string {
	if m.Quitting {
		return "Goodbye!\n"
	}

	// Build the view
	var s strings.Builder

	// Title
	s.WriteString(m.Styles.Header.Render("GitHub Notification Search"))
	s.WriteString("\n\n")

	// Search input
	s.WriteString(m.SearchInput.View())
	s.WriteString("\n\n")

	// Results
	s.WriteString(m.Styles.DetailHeader.Render("Search Results:"))
	s.WriteString("\n")
	if m.Loading {
		s.WriteString(m.Spinner.View() + " Searching...\n")
	} else if len(m.Results) == 0 {
		if m.SearchInput.Value() == "" {
			s.WriteString(m.Styles.NoNotifications.Render("Type a search query and press Enter."))
		} else {
			s.WriteString(m.Styles.NoNotifications.Render("No results found."))
		}
	} else {
		s.WriteString(fmt.Sprintf("Found %d results.\n", len(m.Results)))
		s.WriteString(m.Table.View())
	}
	s.WriteString("\n\n")

	// Error
	if m.Error != nil {
		s.WriteString(m.Styles.Error.Render(fmt.Sprintf("Error: %v", m.Error)))
		s.WriteString("\n\n")
	}

	// Help
	s.WriteString(m.Styles.HelpBar.Render("Press Enter to search, Esc/q to quit"))

	return m.Styles.App.Render(s.String())
}

// searchMsg is a message to search for a query
type searchMsg struct {
	query string
}

// searchResultMsg is a message containing search results
type searchResultMsg struct {
	results []*search.SearchResult
}

// searchErrMsg is a message containing an error
type searchErrMsg struct {
	err error
}

// RunSearchUI runs the search UI
func RunSearchUI(ctx context.Context, client *githubclient.Client, notifications []*github.Notification, initialQuery string) error {
	model := NewSearchModel(ctx, client, notifications, initialQuery)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
