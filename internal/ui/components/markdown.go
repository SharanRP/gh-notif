package components

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// MarkdownRenderer renders markdown content with styling
type MarkdownRenderer struct {
	// Configuration
	width   int
	height  int
	content string

	// Viewport for scrolling
	viewport viewport.Model

	// State
	focused bool

	// Styling
	styles   ComponentStyles
	mdStyles MarkdownStyles
}

// MarkdownStyles defines styles for different markdown elements
type MarkdownStyles struct {
	// Text styles
	Normal lipgloss.Style
	Bold   lipgloss.Style
	Italic lipgloss.Style
	Code   lipgloss.Style
	Link   lipgloss.Style

	// Block styles
	Heading1 lipgloss.Style
	Heading2 lipgloss.Style
	Heading3 lipgloss.Style
	Heading4 lipgloss.Style
	Heading5 lipgloss.Style
	Heading6 lipgloss.Style

	// List styles
	ListItem   lipgloss.Style
	ListBullet lipgloss.Style

	// Block elements
	Blockquote lipgloss.Style
	CodeBlock  lipgloss.Style

	// Table styles
	Table       lipgloss.Style
	TableHeader lipgloss.Style
	TableCell   lipgloss.Style

	// Horizontal rule
	HRule lipgloss.Style
}

// DefaultMarkdownStyles returns default markdown styles
func DefaultMarkdownStyles() MarkdownStyles {
	return MarkdownStyles{
		Normal: lipgloss.NewStyle(),
		Bold:   lipgloss.NewStyle().Bold(true),
		Italic: lipgloss.NewStyle().Italic(true),
		Code: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Background(lipgloss.Color("8")).
			Padding(0, 1),
		Link: lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")).
			Underline(true),

		Heading1: lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true).
			Padding(1, 0),
		Heading2: lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")).
			Bold(true).
			Padding(1, 0),
		Heading3: lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true),
		Heading4: lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true),
		Heading5: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true),
		Heading6: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Bold(true),

		ListItem: lipgloss.NewStyle().
			Padding(0, 0, 0, 2),
		ListBullet: lipgloss.NewStyle().
			Foreground(lipgloss.Color("4")),

		Blockquote: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 0, 0, 1),

		CodeBlock: lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Background(lipgloss.Color("0")).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")),

		Table: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")),
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("4")).
			Padding(0, 1),
		TableCell: lipgloss.NewStyle().
			Padding(0, 1),

		HRule: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Padding(1, 0),
	}
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(content string) *MarkdownRenderer {
	vp := viewport.New(0, 0)

	return &MarkdownRenderer{
		content:  content,
		viewport: vp,
		mdStyles: DefaultMarkdownStyles(),
	}
}

// NewMarkdownComponentFactory creates a markdown component factory
func NewMarkdownComponentFactory(config ComponentConfig) Component {
	content, ok := config.Props["content"].(string)
	if !ok {
		content = ""
	}

	md := NewMarkdownRenderer(content)
	md.SetSize(config.Width, config.Height)
	md.SetStyles(config.Styles)

	return md
}

// SetContent sets the markdown content
func (md *MarkdownRenderer) SetContent(content string) {
	md.content = content
	md.updateViewport()
}

// GetContent returns the current content
func (md *MarkdownRenderer) GetContent() string {
	return md.content
}

// Init initializes the markdown renderer
func (md *MarkdownRenderer) Init() tea.Cmd {
	md.updateViewport()
	return nil
}

// Update handles messages and updates the markdown renderer state
func (md *MarkdownRenderer) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if md.focused {
			md.viewport, cmd = md.viewport.Update(msg)
		}

	case ComponentMessage:
		switch msg.Type {
		case ComponentResizeMsg:
			if size, ok := msg.Data.(struct{ Width, Height int }); ok {
				md.SetSize(size.Width, size.Height)
			}
		case "content":
			if content, ok := msg.Data.(string); ok {
				md.SetContent(content)
			}
		}
	}

	return md, cmd
}

// View renders the markdown content
func (md *MarkdownRenderer) View() string {
	if md.content == "" {
		return md.styles.Base.Render("No content")
	}

	rendered := md.renderMarkdown(md.content)
	md.viewport.SetContent(rendered)

	return md.viewport.View()
}

// renderMarkdown renders markdown content to styled text
func (md *MarkdownRenderer) renderMarkdown(content string) string {
	lines := strings.Split(content, "\n")
	var rendered []string

	inCodeBlock := false
	var codeBlockLines []string

	for _, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End code block
				codeContent := strings.Join(codeBlockLines, "\n")
				rendered = append(rendered, md.mdStyles.CodeBlock.Render(codeContent))
				codeBlockLines = []string{}
				inCodeBlock = false
			} else {
				// Start code block
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line)
			continue
		}

		// Process regular lines
		rendered = append(rendered, md.processLine(line))
	}

	return strings.Join(rendered, "\n")
}

// processLine processes a single line of markdown
func (md *MarkdownRenderer) processLine(line string) string {
	// Handle headings
	if strings.HasPrefix(line, "# ") {
		return md.mdStyles.Heading1.Render(strings.TrimPrefix(line, "# "))
	}
	if strings.HasPrefix(line, "## ") {
		return md.mdStyles.Heading2.Render(strings.TrimPrefix(line, "## "))
	}
	if strings.HasPrefix(line, "### ") {
		return md.mdStyles.Heading3.Render(strings.TrimPrefix(line, "### "))
	}
	if strings.HasPrefix(line, "#### ") {
		return md.mdStyles.Heading4.Render(strings.TrimPrefix(line, "#### "))
	}
	if strings.HasPrefix(line, "##### ") {
		return md.mdStyles.Heading5.Render(strings.TrimPrefix(line, "##### "))
	}
	if strings.HasPrefix(line, "###### ") {
		return md.mdStyles.Heading6.Render(strings.TrimPrefix(line, "###### "))
	}

	// Handle blockquotes
	if strings.HasPrefix(line, "> ") {
		content := strings.TrimPrefix(line, "> ")
		return md.mdStyles.Blockquote.Render(content)
	}

	// Handle list items
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || strings.HasPrefix(line, "+ ") {
		bullet := md.mdStyles.ListBullet.Render("•")
		content := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "), "+ ")
		return md.mdStyles.ListItem.Render(bullet + " " + md.processInlineMarkdown(content))
	}

	// Handle numbered lists
	if matched, _ := regexp.MatchString(`^\d+\. `, line); matched {
		re := regexp.MustCompile(`^(\d+)\. (.*)`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			number := md.mdStyles.ListBullet.Render(matches[1] + ".")
			content := matches[2]
			return md.mdStyles.ListItem.Render(number + " " + md.processInlineMarkdown(content))
		}
	}

	// Handle horizontal rules
	if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" {
		rule := strings.Repeat("─", md.width-4)
		return md.mdStyles.HRule.Render(rule)
	}

	// Handle empty lines
	if strings.TrimSpace(line) == "" {
		return ""
	}

	// Process inline markdown for regular text
	return md.processInlineMarkdown(line)
}

// processInlineMarkdown processes inline markdown elements
func (md *MarkdownRenderer) processInlineMarkdown(text string) string {
	// Handle bold text (**text** or __text__)
	boldRegex := regexp.MustCompile(`\*\*(.*?)\*\*|__(.*?)__`)
	text = boldRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(strings.Trim(match, "*"), "_")
		return md.mdStyles.Bold.Render(content)
	})

	// Handle italic text (*text* or _text_)
	italicRegex := regexp.MustCompile(`\*(.*?)\*|_(.*?)_`)
	text = italicRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(strings.Trim(match, "*"), "_")
		return md.mdStyles.Italic.Render(content)
	})

	// Handle inline code (`code`)
	codeRegex := regexp.MustCompile("`([^`]+)`")
	text = codeRegex.ReplaceAllStringFunc(text, func(match string) string {
		content := strings.Trim(match, "`")
		return md.mdStyles.Code.Render(content)
	})

	// Handle links ([text](url))
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRegex.ReplaceAllStringFunc(text, func(match string) string {
		matches := linkRegex.FindStringSubmatch(match)
		if len(matches) == 3 {
			linkText := matches[1]
			return md.mdStyles.Link.Render(linkText)
		}
		return match
	})

	return text
}

// updateViewport updates the viewport content
func (md *MarkdownRenderer) updateViewport() {
	if md.content != "" {
		rendered := md.renderMarkdown(md.content)
		md.viewport.SetContent(rendered)
	}
}

// SetSize sets the component dimensions
func (md *MarkdownRenderer) SetSize(width, height int) {
	md.width = width
	md.height = height
	md.viewport.Width = width
	md.viewport.Height = height
	md.updateViewport()
}

// GetSize returns the component dimensions
func (md *MarkdownRenderer) GetSize() (width, height int) {
	return md.width, md.height
}

// SetStyles sets the component styles
func (md *MarkdownRenderer) SetStyles(styles ComponentStyles) {
	md.styles = styles

	// Update markdown styles based on component styles
	md.mdStyles.Normal = styles.Base
	md.mdStyles.Bold = styles.Base.Bold(true)
	md.mdStyles.Italic = styles.Base.Italic(true)
}

// GetType returns the component type
func (md *MarkdownRenderer) GetType() ComponentType {
	return MarkdownComponentType
}

// SetFocused sets the focus state
func (md *MarkdownRenderer) SetFocused(focused bool) {
	md.focused = focused
}

// IsFocused returns the focus state
func (md *MarkdownRenderer) IsFocused() bool {
	return md.focused
}

// ScrollUp scrolls the content up
func (md *MarkdownRenderer) ScrollUp() {
	md.viewport.LineUp(1)
}

// ScrollDown scrolls the content down
func (md *MarkdownRenderer) ScrollDown() {
	md.viewport.LineDown(1)
}

// ScrollToTop scrolls to the top of the content
func (md *MarkdownRenderer) ScrollToTop() {
	md.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom of the content
func (md *MarkdownRenderer) ScrollToBottom() {
	md.viewport.GotoBottom()
}

// GetScrollPercent returns the current scroll percentage
func (md *MarkdownRenderer) GetScrollPercent() float64 {
	return md.viewport.ScrollPercent()
}

// SetMarkdownStyles sets custom markdown styles
func (md *MarkdownRenderer) SetMarkdownStyles(styles MarkdownStyles) {
	md.mdStyles = styles
	md.updateViewport()
}
