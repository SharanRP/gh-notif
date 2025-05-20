package ui

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MarkdownRenderer renders markdown text to styled terminal output
type MarkdownRenderer struct {
	styles Styles
	width  int
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(styles Styles, width int) *MarkdownRenderer {
	return &MarkdownRenderer{
		styles: styles,
		width:  width,
	}
}

// Render converts markdown text to styled terminal output
func (r *MarkdownRenderer) Render(markdown string) string {
	if markdown == "" {
		return ""
	}

	// Process the markdown
	lines := strings.Split(markdown, "\n")
	var output bytes.Buffer

	inCodeBlock := false
	inQuote := false
	inList := false
	codeLanguage := ""

	for i, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				if len(line) > 3 {
					codeLanguage = strings.TrimSpace(line[3:])
				}
				continue
			} else {
				inCodeBlock = false
				codeLanguage = ""
				continue
			}
		}

		if inCodeBlock {
			// Render code with syntax highlighting
			output.WriteString(r.renderCodeLine(line, codeLanguage))
			output.WriteString("\n")
			continue
		}

		// Handle blockquotes
		if strings.HasPrefix(line, ">") {
			if !inQuote {
				inQuote = true
			}
			quotedText := strings.TrimSpace(line[1:])
			output.WriteString(r.renderQuote(quotedText))
			output.WriteString("\n")
			continue
		} else if inQuote {
			inQuote = false
		}

		// Handle headers
		if strings.HasPrefix(line, "#") {
			level := 0
			for i, char := range line {
				if char == '#' {
					level++
				} else {
					line = strings.TrimSpace(line[i:])
					break
				}
			}
			output.WriteString(r.renderHeader(line, level))
			output.WriteString("\n")
			continue
		}

		// Handle lists
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") || 
		   regexp.MustCompile(`^\d+\.\s`).MatchString(line) {
			if !inList {
				inList = true
			}
			output.WriteString(r.renderListItem(line))
			output.WriteString("\n")
			continue
		} else if inList && len(strings.TrimSpace(line)) > 0 {
			inList = false
		}

		// Handle horizontal rules
		if regexp.MustCompile(`^(\*\*\*|\-\-\-|___)\s*$`).MatchString(line) {
			output.WriteString(r.renderHorizontalRule())
			output.WriteString("\n")
			continue
		}

		// Handle regular text with inline formatting
		if len(strings.TrimSpace(line)) > 0 {
			output.WriteString(r.renderText(line))
			output.WriteString("\n")
		} else if i > 0 && len(strings.TrimSpace(lines[i-1])) > 0 {
			// Add empty line only if previous line wasn't empty
			output.WriteString("\n")
		}
	}

	return output.String()
}

// renderHeader renders a markdown header
func (r *MarkdownRenderer) renderHeader(text string, level int) string {
	style := r.styles.DetailHeader.Copy()
	
	switch level {
	case 1:
		style = style.Bold(true).Underline(true).
			MarginBottom(1).
			Foreground(lipgloss.Color("#89B4FA"))
	case 2:
		style = style.Bold(true).
			MarginBottom(1).
			Foreground(lipgloss.Color("#89B4FA"))
	case 3:
		style = style.Bold(true).
			Foreground(lipgloss.Color("#89B4FA"))
	default:
		style = style.Bold(true)
	}
	
	return style.Render(text)
}

// renderQuote renders a blockquote
func (r *MarkdownRenderer) renderQuote(text string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA0B0")).
		Italic(true).
		Padding(0, 0, 0, 2).
		Border(lipgloss.Border{Left: "│"}).
		BorderForeground(lipgloss.Color("#6C7086"))
	
	return style.Render(text)
}

// renderListItem renders a list item
func (r *MarkdownRenderer) renderListItem(text string) string {
	// Extract the bullet or number
	var bullet string
	var content string
	
	if strings.HasPrefix(text, "- ") {
		bullet = "•"
		content = strings.TrimSpace(text[2:])
	} else if strings.HasPrefix(text, "* ") {
		bullet = "•"
		content = strings.TrimSpace(text[2:])
	} else {
		// Numbered list
		parts := regexp.MustCompile(`^(\d+)\.`).FindStringSubmatch(text)
		if len(parts) > 1 {
			bullet = parts[1] + "."
			content = strings.TrimSpace(text[len(parts[0]):])
		} else {
			bullet = "•"
			content = text
		}
	}
	
	bulletStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#89B4FA")).
		Width(3).
		Align(lipgloss.Right)
	
	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4"))
	
	return lipgloss.JoinHorizontal(lipgloss.Top,
		bulletStyle.Render(bullet),
		" ",
		contentStyle.Render(r.renderText(content)),
	)
}

// renderHorizontalRule renders a horizontal rule
func (r *MarkdownRenderer) renderHorizontalRule() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6C7086"))
	
	return style.Render(strings.Repeat("─", r.width-4))
}

// renderCodeLine renders a line of code with syntax highlighting
func (r *MarkdownRenderer) renderCodeLine(code, language string) string {
	// Simple syntax highlighting for common languages
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F5E0DC")).
		Background(lipgloss.Color("#313244")).
		Padding(0, 1)
	
	// Apply basic syntax highlighting based on language
	switch language {
	case "go", "golang":
		code = r.highlightGo(code)
	case "js", "javascript":
		code = r.highlightJavaScript(code)
	case "py", "python":
		code = r.highlightPython(code)
	case "json":
		code = r.highlightJSON(code)
	}
	
	return style.Render(code)
}

// renderText renders regular text with inline formatting
func (r *MarkdownRenderer) renderText(text string) string {
	// Handle bold text
	boldPattern := regexp.MustCompile(`\*\*(.+?)\*\*`)
	text = boldPattern.ReplaceAllStringFunc(text, func(match string) string {
		inner := boldPattern.FindStringSubmatch(match)[1]
		return lipgloss.NewStyle().Bold(true).Render(inner)
	})
	
	// Handle italic text
	italicPattern := regexp.MustCompile(`\*(.+?)\*`)
	text = italicPattern.ReplaceAllStringFunc(text, func(match string) string {
		inner := italicPattern.FindStringSubmatch(match)[1]
		return lipgloss.NewStyle().Italic(true).Render(inner)
	})
	
	// Handle code spans
	codePattern := regexp.MustCompile("`(.+?)`")
	text = codePattern.ReplaceAllStringFunc(text, func(match string) string {
		inner := codePattern.FindStringSubmatch(match)[1]
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5E0DC")).
			Background(lipgloss.Color("#313244")).
			Padding(0, 1).
			Render(inner)
	})
	
	// Handle links
	linkPattern := regexp.MustCompile(`\[(.+?)\]\((.+?)\)`)
	text = linkPattern.ReplaceAllStringFunc(text, func(match string) string {
		parts := linkPattern.FindStringSubmatch(match)
		text := parts[1]
		url := parts[2]
		return fmt.Sprintf("%s (%s)",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Underline(true).Render(text),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(url),
		)
	})
	
	return text
}

// highlightGo applies basic syntax highlighting for Go code
func (r *MarkdownRenderer) highlightGo(code string) string {
	// Keywords
	keywords := []string{
		"func", "package", "import", "var", "const", "type", "struct", "interface",
		"map", "chan", "go", "defer", "if", "else", "switch", "case", "default",
		"for", "range", "return", "break", "continue",
	}
	
	for _, keyword := range keywords {
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		code = pattern.ReplaceAllString(code, 
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render(keyword))
	}
	
	// Comments
	commentPattern := regexp.MustCompile(`//.*$`)
	code = commentPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(match)
	})
	
	// Strings
	stringPattern := regexp.MustCompile(`"[^"]*"`)
	code = stringPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render(match)
	})
	
	return code
}

// highlightJavaScript applies basic syntax highlighting for JavaScript code
func (r *MarkdownRenderer) highlightJavaScript(code string) string {
	// Keywords
	keywords := []string{
		"function", "const", "let", "var", "if", "else", "switch", "case",
		"default", "for", "while", "do", "break", "continue", "return",
		"class", "new", "this", "super", "import", "export", "from", "as",
	}
	
	for _, keyword := range keywords {
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		code = pattern.ReplaceAllString(code, 
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render(keyword))
	}
	
	// Comments
	commentPattern := regexp.MustCompile(`//.*$`)
	code = commentPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(match)
	})
	
	// Strings
	stringPattern := regexp.MustCompile(`["'].*?["']`)
	code = stringPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render(match)
	})
	
	return code
}

// highlightPython applies basic syntax highlighting for Python code
func (r *MarkdownRenderer) highlightPython(code string) string {
	// Keywords
	keywords := []string{
		"def", "class", "import", "from", "as", "if", "elif", "else", "for",
		"while", "break", "continue", "return", "try", "except", "finally",
		"with", "lambda", "global", "nonlocal", "pass", "None", "True", "False",
	}
	
	for _, keyword := range keywords {
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		code = pattern.ReplaceAllString(code, 
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render(keyword))
	}
	
	// Comments
	commentPattern := regexp.MustCompile(`#.*$`)
	code = commentPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(match)
	})
	
	// Strings
	stringPattern := regexp.MustCompile(`["'].*?["']`)
	code = stringPattern.ReplaceAllStringFunc(code, func(match string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render(match)
	})
	
	return code
}

// highlightJSON applies basic syntax highlighting for JSON
func (r *MarkdownRenderer) highlightJSON(code string) string {
	// Keys
	keyPattern := regexp.MustCompile(`"([^"]+)"(\s*:)`)
	code = keyPattern.ReplaceAllString(code, 
		lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render("\"$1\"") + "$2")
	
	// Strings
	stringPattern := regexp.MustCompile(`:\s*"([^"]*)"`)
	code = stringPattern.ReplaceAllString(code, 
		": " + lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("\"$1\""))
	
	// Numbers
	numberPattern := regexp.MustCompile(`:\s*(\d+)`)
	code = numberPattern.ReplaceAllString(code, 
		": " + lipgloss.NewStyle().Foreground(lipgloss.Color("#F9E2AF")).Render("$1"))
	
	// Booleans and null
	boolPattern := regexp.MustCompile(`:\s*(true|false|null)`)
	code = boolPattern.ReplaceAllString(code, 
		": " + lipgloss.NewStyle().Foreground(lipgloss.Color("#CBA6F7")).Render("$1"))
	
	return code
}
