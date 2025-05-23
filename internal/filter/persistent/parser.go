package persistent

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/SharanRP/gh-notif/internal/filter"
)

// Parser parses filter expressions into Filter objects
type Parser struct {
	// store is the filter store for resolving references
	store *FilterStore
}

// NewParser creates a new parser
func NewParser(store *FilterStore) *Parser {
	return &Parser{
		store: store,
	}
}

// Parse parses a filter expression into a Filter
func (p *Parser) Parse(expr string) (filter.Filter, error) {
	// Trim whitespace
	expr = strings.TrimSpace(expr)

	// Handle empty expression
	if expr == "" {
		return &filter.AllFilter{}, nil
	}

	// Check if it's a reference to a saved filter
	if strings.HasPrefix(expr, "@") {
		return p.parseReference(expr)
	}

	// Check if it's a complex expression with boolean operators
	if strings.Contains(expr, " AND ") || strings.Contains(expr, " OR ") || strings.Contains(expr, " NOT ") {
		return p.parseComplexExpression(expr)
	}

	// Parse as a simple expression
	return p.parseSimpleExpression(expr)
}

// parseReference parses a reference to a saved filter
func (p *Parser) parseReference(expr string) (filter.Filter, error) {
	// Extract the filter name
	name := strings.TrimPrefix(expr, "@")

	// Get the filter preset
	preset, err := p.store.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve filter reference: %w", err)
	}

	// Parse the referenced expression
	return p.Parse(preset.Expression)
}

// parseComplexExpression parses a complex expression with boolean operators
func (p *Parser) parseComplexExpression(expr string) (filter.Filter, error) {
	// Split the expression into tokens
	tokens, err := p.tokenize(expr)
	if err != nil {
		return nil, err
	}

	// Parse the tokens into a filter
	return p.parseTokens(tokens)
}

// Token represents a token in a filter expression
type Token struct {
	// Type is the token type
	Type string
	// Value is the token value
	Value string
}

// tokenize splits a complex expression into tokens
func (p *Parser) tokenize(expr string) ([]Token, error) {
	// Replace parentheses with spaces around them for easier parsing
	expr = strings.ReplaceAll(expr, "(", " ( ")
	expr = strings.ReplaceAll(expr, ")", " ) ")

	// Split the expression into parts
	parts := strings.Fields(expr)

	// Convert parts to tokens
	var tokens []Token
	for _, part := range parts {
		switch strings.ToUpper(part) {
		case "AND":
			tokens = append(tokens, Token{Type: "AND", Value: "AND"})
		case "OR":
			tokens = append(tokens, Token{Type: "OR", Value: "OR"})
		case "NOT":
			tokens = append(tokens, Token{Type: "NOT", Value: "NOT"})
		case "(":
			tokens = append(tokens, Token{Type: "LPAREN", Value: "("})
		case ")":
			tokens = append(tokens, Token{Type: "RPAREN", Value: ")"})
		default:
			// If it's not an operator or parenthesis, it's an expression
			tokens = append(tokens, Token{Type: "EXPR", Value: part})
		}
	}

	return tokens, nil
}

// parseTokens parses tokens into a filter using the shunting yard algorithm
func (p *Parser) parseTokens(tokens []Token) (filter.Filter, error) {
	// Implement a simplified version of the shunting yard algorithm
	// This is a basic implementation that handles AND, OR, NOT, and parentheses

	// Define operator precedence
	precedence := map[string]int{
		"NOT": 3,
		"AND": 2,
		"OR":  1,
	}

	// Output queue for expressions
	var output []filter.Filter

	// Operator stack
	var operators []Token

	// Process tokens
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		switch token.Type {
		case "EXPR":
			// Parse the expression and add it to the output queue
			f, err := p.parseSimpleExpression(token.Value)
			if err != nil {
				return nil, err
			}
			output = append(output, f)

		case "LPAREN":
			// Push left parenthesis onto the operator stack
			operators = append(operators, token)

		case "RPAREN":
			// Pop operators until we find a left parenthesis
			for len(operators) > 0 && operators[len(operators)-1].Type != "LPAREN" {
				// Pop an operator and apply it
				op := operators[len(operators)-1]
				operators = operators[:len(operators)-1]

				if err := p.applyOperator(op, &output); err != nil {
					return nil, err
				}
			}

			// Pop the left parenthesis
			if len(operators) > 0 && operators[len(operators)-1].Type == "LPAREN" {
				operators = operators[:len(operators)-1]
			} else {
				return nil, fmt.Errorf("mismatched parentheses")
			}

		case "AND", "OR", "NOT":
			// While there's an operator with higher precedence on the stack, pop it
			for len(operators) > 0 && operators[len(operators)-1].Type != "LPAREN" &&
				precedence[operators[len(operators)-1].Type] >= precedence[token.Type] {
				// Pop an operator and apply it
				op := operators[len(operators)-1]
				operators = operators[:len(operators)-1]

				if err := p.applyOperator(op, &output); err != nil {
					return nil, err
				}
			}

			// Push the current operator onto the stack
			operators = append(operators, token)
		}
	}

	// Pop remaining operators
	for len(operators) > 0 {
		op := operators[len(operators)-1]
		operators = operators[:len(operators)-1]

		if op.Type == "LPAREN" || op.Type == "RPAREN" {
			return nil, fmt.Errorf("mismatched parentheses")
		}

		if err := p.applyOperator(op, &output); err != nil {
			return nil, err
		}
	}

	// The output queue should now contain exactly one filter
	if len(output) != 1 {
		return nil, fmt.Errorf("invalid expression")
	}

	return output[0], nil
}

// applyOperator applies an operator to the output queue
func (p *Parser) applyOperator(op Token, output *[]filter.Filter) error {
	switch op.Type {
	case "AND":
		// AND requires two operands
		if len(*output) < 2 {
			return fmt.Errorf("not enough operands for AND")
		}

		// Pop the two operands
		right := (*output)[len(*output)-1]
		left := (*output)[len(*output)-2]
		*output = (*output)[:len(*output)-2]

		// Create a composite filter with AND
		*output = append(*output, &filter.CompositeFilter{
			Filters:  []filter.Filter{left, right},
			Operator: filter.And,
		})

	case "OR":
		// OR requires two operands
		if len(*output) < 2 {
			return fmt.Errorf("not enough operands for OR")
		}

		// Pop the two operands
		right := (*output)[len(*output)-1]
		left := (*output)[len(*output)-2]
		*output = (*output)[:len(*output)-2]

		// Create a composite filter with OR
		*output = append(*output, &filter.CompositeFilter{
			Filters:  []filter.Filter{left, right},
			Operator: filter.Or,
		})

	case "NOT":
		// NOT requires one operand
		if len(*output) < 1 {
			return fmt.Errorf("not enough operands for NOT")
		}

		// Pop the operand
		operand := (*output)[len(*output)-1]
		*output = (*output)[:len(*output)-1]

		// Create a composite filter with NOT
		*output = append(*output, &filter.CompositeFilter{
			Filters:  []filter.Filter{operand},
			Operator: filter.Not,
		})
	}

	return nil
}

// parseSimpleExpression parses a simple expression into a Filter
func (p *Parser) parseSimpleExpression(expr string) (filter.Filter, error) {
	// Check for key:value format
	if strings.Contains(expr, ":") {
		return p.parseKeyValueExpression(expr)
	}

	// Default to text search
	return &filter.TextFilter{Text: expr}, nil
}

// parseKeyValueExpression parses a key:value expression into a Filter
func (p *Parser) parseKeyValueExpression(expr string) (filter.Filter, error) {
	// Split into key and value
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid key:value expression: %s", expr)
	}

	key := strings.ToLower(parts[0])
	value := parts[1]

	// Handle different keys
	switch key {
	case "is":
		return p.parseIsExpression(value)
	case "repo", "repository":
		return p.parseRepoExpression(value)
	case "org", "organization":
		return p.parseOrgExpression(value)
	case "type":
		return p.parseTypeExpression(value)
	case "reason":
		return p.parseReasonExpression(value)
	case "updated", "since":
		return p.parseTimeExpression(value)
	case "score":
		return p.parseScoreExpression(value)
	default:
		// Default to regex search on the specified field
		return p.parseRegexExpression(key, value)
	}
}

// parseIsExpression parses an is:value expression
func (p *Parser) parseIsExpression(value string) (filter.Filter, error) {
	switch strings.ToLower(value) {
	case "read":
		return &filter.ReadFilter{Read: true}, nil
	case "unread":
		return &filter.ReadFilter{Read: false}, nil
	default:
		return nil, fmt.Errorf("invalid is: value: %s", value)
	}
}

// parseRepoExpression parses a repo:value expression
func (p *Parser) parseRepoExpression(value string) (filter.Filter, error) {
	return filter.NewRepositoryFilter(value)
}

// parseOrgExpression parses an org:value expression
func (p *Parser) parseOrgExpression(value string) (filter.Filter, error) {
	return filter.NewOrganizationFilter(value)
}

// parseTypeExpression parses a type:value expression
func (p *Parser) parseTypeExpression(value string) (filter.Filter, error) {
	return filter.NewTypeFilter(value), nil
}

// parseReasonExpression parses a reason:value expression
func (p *Parser) parseReasonExpression(value string) (filter.Filter, error) {
	return &filter.ReasonFilter{Reason: value}, nil
}

// parseTimeExpression parses a time-based expression
func (p *Parser) parseTimeExpression(value string) (filter.Filter, error) {
	// Check for comparison operators
	if strings.HasPrefix(value, ">") {
		// Greater than (newer than)
		duration, err := parseDuration(strings.TrimPrefix(value, ">"))
		if err != nil {
			return nil, err
		}
		return &filter.TimeFilter{
			UsesSince: true,
			Since:     time.Now().Add(-duration),
		}, nil
	} else if strings.HasPrefix(value, "<") {
		// Less than (older than)
		duration, err := parseDuration(strings.TrimPrefix(value, "<"))
		if err != nil {
			return nil, err
		}
		return &filter.TimeFilter{
			UsesBefore: true,
			Before:     time.Now().Add(-duration),
		}, nil
	}

	// Default to exact time
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("invalid time format: %s", value)
	}

	return &filter.TimeFilter{
		UsesSince:  true,
		UsesBefore: true,
		Since:      t,
		Before:     t.Add(24 * time.Hour),
	}, nil
}

// parseScoreExpression parses a score:value expression
func (p *Parser) parseScoreExpression(value string) (filter.Filter, error) {
	// Check for comparison operators
	if strings.HasPrefix(value, ">") {
		// Greater than
		score, err := strconv.Atoi(strings.TrimPrefix(value, ">"))
		if err != nil {
			return nil, fmt.Errorf("invalid score: %s", value)
		}
		return &ScoreFilter{
			MinScore: score,
		}, nil
	} else if strings.HasPrefix(value, "<") {
		// Less than
		score, err := strconv.Atoi(strings.TrimPrefix(value, "<"))
		if err != nil {
			return nil, fmt.Errorf("invalid score: %s", value)
		}
		return &ScoreFilter{
			MaxScore: score,
		}, nil
	}

	// Default to exact score
	score, err := strconv.Atoi(value)
	if err != nil {
		return nil, fmt.Errorf("invalid score: %s", value)
	}

	return &ScoreFilter{
		MinScore: score,
		MaxScore: score,
	}, nil
}

// parseRegexExpression parses a field:regex expression
func (p *Parser) parseRegexExpression(field, pattern string) (filter.Filter, error) {
	// Compile the regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %s", pattern)
	}

	return &filter.RegexFilter{
		Field:   field,
		Pattern: re,
	}, nil
}

// parseDuration parses a duration string
func parseDuration(s string) (time.Duration, error) {
	// Check for common time units
	if strings.HasSuffix(s, "h") {
		// Hours
		hours, err := strconv.Atoi(strings.TrimSuffix(s, "h"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(hours) * time.Hour, nil
	} else if strings.HasSuffix(s, "d") {
		// Days
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	} else if strings.HasSuffix(s, "w") {
		// Weeks
		weeks, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	}

	// Try to parse as a Go duration
	return time.ParseDuration(s)
}
