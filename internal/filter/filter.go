package filter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/google/go-github/v60/github"
)

// Operator represents a logical operator for combining filters
type Operator int

const (
	// And represents logical AND
	And Operator = iota
	// Or represents logical OR
	Or
	// Not represents logical NOT
	Not
)

// FilterFunc is a function that filters notifications
type FilterFunc func(*github.Notification) bool

// Filter represents a notification filter
type Filter interface {
	// Apply applies the filter to a notification
	Apply(*github.Notification) bool
	// Description returns a human-readable description of the filter
	Description() string
}

// CompositeFilter combines multiple filters with a logical operator
type CompositeFilter struct {
	Filters  []Filter
	Operator Operator
}

// Apply applies the composite filter to a notification
func (f *CompositeFilter) Apply(n *github.Notification) bool {
	if len(f.Filters) == 0 {
		return true
	}

	switch f.Operator {
	case And:
		for _, filter := range f.Filters {
			if !filter.Apply(n) {
				return false
			}
		}
		return true
	case Or:
		for _, filter := range f.Filters {
			if filter.Apply(n) {
				return true
			}
		}
		return false
	case Not:
		// NOT only makes sense with exactly one filter
		if len(f.Filters) != 1 {
			return false
		}
		return !f.Filters[0].Apply(n)
	default:
		return false
	}
}

// Description returns a human-readable description of the composite filter
func (f *CompositeFilter) Description() string {
	if len(f.Filters) == 0 {
		return "no filters"
	}

	var descriptions []string
	for _, filter := range f.Filters {
		descriptions = append(descriptions, filter.Description())
	}

	var op string
	switch f.Operator {
	case And:
		op = "AND"
	case Or:
		op = "OR"
	case Not:
		op = "NOT"
	}

	if f.Operator == Not {
		return fmt.Sprintf("NOT (%s)", descriptions[0])
	}

	return fmt.Sprintf("(%s)", strings.Join(descriptions, fmt.Sprintf(" %s ", op)))
}

// RepositoryFilter filters notifications by repository name
type RepositoryFilter struct {
	Pattern glob.Glob
	Raw     string
}

// NewRepositoryFilter creates a new repository filter
func NewRepositoryFilter(pattern string) (*RepositoryFilter, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid repository pattern: %w", err)
	}

	return &RepositoryFilter{
		Pattern: g,
		Raw:     pattern,
	}, nil
}

// Apply applies the repository filter to a notification
func (f *RepositoryFilter) Apply(n *github.Notification) bool {
	repo := n.GetRepository().GetFullName()
	return f.Pattern.Match(repo)
}

// Description returns a human-readable description of the repository filter
func (f *RepositoryFilter) Description() string {
	return fmt.Sprintf("repository matches %s", f.Raw)
}

// OrganizationFilter filters notifications by organization name
type OrganizationFilter struct {
	Pattern glob.Glob
	Raw     string
}

// NewOrganizationFilter creates a new organization filter
func NewOrganizationFilter(pattern string) (*OrganizationFilter, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid organization pattern: %w", err)
	}

	return &OrganizationFilter{
		Pattern: g,
		Raw:     pattern,
	}, nil
}

// Apply applies the organization filter to a notification
func (f *OrganizationFilter) Apply(n *github.Notification) bool {
	fullName := n.GetRepository().GetFullName()
	parts := strings.Split(fullName, "/")
	if len(parts) < 2 {
		return false
	}
	org := parts[0]
	return f.Pattern.Match(org)
}

// Description returns a human-readable description of the organization filter
func (f *OrganizationFilter) Description() string {
	return fmt.Sprintf("organization matches %s", f.Raw)
}

// TypeFilter filters notifications by type
type TypeFilter struct {
	Types []string
}

// NewTypeFilter creates a new type filter
func NewTypeFilter(types ...string) *TypeFilter {
	return &TypeFilter{
		Types: types,
	}
}

// Apply applies the type filter to a notification
func (f *TypeFilter) Apply(n *github.Notification) bool {
	notifType := n.GetSubject().GetType()
	for _, t := range f.Types {
		if strings.EqualFold(notifType, t) {
			return true
		}
	}
	return false
}

// Description returns a human-readable description of the type filter
func (f *TypeFilter) Description() string {
	return fmt.Sprintf("type is one of [%s]", strings.Join(f.Types, ", "))
}

// StatusFilter filters notifications by read status
type StatusFilter struct {
	Unread bool
}

// NewStatusFilter creates a new status filter
func NewStatusFilter(unread bool) *StatusFilter {
	return &StatusFilter{
		Unread: unread,
	}
}

// Apply applies the status filter to a notification
func (f *StatusFilter) Apply(n *github.Notification) bool {
	return n.GetUnread() == f.Unread
}

// Description returns a human-readable description of the status filter
func (f *StatusFilter) Description() string {
	if f.Unread {
		return "is unread"
	}
	return "is read"
}

// TimeFilter filters notifications by time
type TimeFilter struct {
	Since      time.Time
	Before     time.Time
	UsesSince  bool
	UsesBefore bool
}

// NewTimeFilter creates a new time filter
func NewTimeFilter() *TimeFilter {
	return &TimeFilter{}
}

// WithSince adds a since time to the filter
func (f *TimeFilter) WithSince(since time.Time) *TimeFilter {
	f.Since = since
	f.UsesSince = true
	return f
}

// WithBefore adds a before time to the filter
func (f *TimeFilter) WithBefore(before time.Time) *TimeFilter {
	f.Before = before
	f.UsesBefore = true
	return f
}

// Apply applies the time filter to a notification
func (f *TimeFilter) Apply(n *github.Notification) bool {
	updatedAt := n.GetUpdatedAt().Time

	if f.UsesSince && updatedAt.Before(f.Since) {
		return false
	}

	if f.UsesBefore && (updatedAt.After(f.Before) || updatedAt.Equal(f.Before)) {
		return false
	}

	return true
}

// Description returns a human-readable description of the time filter
func (f *TimeFilter) Description() string {
	var parts []string

	if f.UsesSince {
		parts = append(parts, fmt.Sprintf("updated after %s", f.Since.Format(time.RFC3339)))
	}

	if f.UsesBefore {
		parts = append(parts, fmt.Sprintf("updated before %s", f.Before.Format(time.RFC3339)))
	}

	return strings.Join(parts, " AND ")
}

// RegexFilter filters notifications by a regular expression
type RegexFilter struct {
	Pattern *regexp.Regexp
	Field   string
}

// NewRegexFilter creates a new regex filter
func NewRegexFilter(pattern string, field string) (*RegexFilter, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	return &RegexFilter{
		Pattern: re,
		Field:   field,
	}, nil
}

// Apply applies the regex filter to a notification
func (f *RegexFilter) Apply(n *github.Notification) bool {
	var value string

	switch strings.ToLower(f.Field) {
	case "title":
		value = n.GetSubject().GetTitle()
	case "repository", "repo":
		value = n.GetRepository().GetFullName()
	case "type":
		value = n.GetSubject().GetType()
	case "reason":
		value = n.GetReason()
	default:
		// Default to title
		value = n.GetSubject().GetTitle()
	}

	return f.Pattern.MatchString(value)
}

// Description returns a human-readable description of the regex filter
func (f *RegexFilter) Description() string {
	return fmt.Sprintf("%s matches regex %s", f.Field, f.Pattern.String())
}

// AllFilter matches all notifications
type AllFilter struct{}

// Apply returns true for all notifications
func (f *AllFilter) Apply(notification *github.Notification) bool {
	return true
}

// Description returns a human-readable description of the filter
func (f *AllFilter) Description() string {
	return "all notifications"
}

// ReadFilter filters notifications by read status
type ReadFilter struct {
	// Read is true to match read notifications, false to match unread
	Read bool
}

// Apply returns true if the notification matches the read status
func (f *ReadFilter) Apply(notification *github.Notification) bool {
	return notification.GetUnread() != f.Read
}

// Description returns a human-readable description of the filter
func (f *ReadFilter) Description() string {
	if f.Read {
		return "read notifications"
	}
	return "unread notifications"
}

// RepoFilter filters notifications by repository
type RepoFilter struct {
	// Repo is the repository name to match
	Repo string
}

// Apply returns true if the notification is from the specified repository
func (f *RepoFilter) Apply(notification *github.Notification) bool {
	return strings.EqualFold(notification.GetRepository().GetFullName(), f.Repo)
}

// Description returns a human-readable description of the filter
func (f *RepoFilter) Description() string {
	return fmt.Sprintf("repository is %s", f.Repo)
}

// OrgFilter filters notifications by organization
type OrgFilter struct {
	// Org is the organization name to match
	Org string
}

// Apply returns true if the notification is from the specified organization
func (f *OrgFilter) Apply(notification *github.Notification) bool {
	return strings.EqualFold(notification.GetRepository().GetOwner().GetLogin(), f.Org)
}

// Description returns a human-readable description of the filter
func (f *OrgFilter) Description() string {
	return fmt.Sprintf("organization is %s", f.Org)
}

// ReasonFilter filters notifications by reason
type ReasonFilter struct {
	// Reason is the notification reason to match
	Reason string
}

// Apply returns true if the notification has the specified reason
func (f *ReasonFilter) Apply(notification *github.Notification) bool {
	return strings.EqualFold(notification.GetReason(), f.Reason)
}

// Description returns a human-readable description of the filter
func (f *ReasonFilter) Description() string {
	return fmt.Sprintf("reason is %s", f.Reason)
}

// TextFilter filters notifications by text
type TextFilter struct {
	// Text is the text to search for
	Text string
}

// Apply returns true if the notification contains the specified text
func (f *TextFilter) Apply(notification *github.Notification) bool {
	// Check the title
	if strings.Contains(strings.ToLower(notification.GetSubject().GetTitle()), strings.ToLower(f.Text)) {
		return true
	}

	// Check the repository name
	if strings.Contains(strings.ToLower(notification.GetRepository().GetFullName()), strings.ToLower(f.Text)) {
		return true
	}

	// Check the type
	if strings.Contains(strings.ToLower(notification.GetSubject().GetType()), strings.ToLower(f.Text)) {
		return true
	}

	// Check the reason
	if strings.Contains(strings.ToLower(notification.GetReason()), strings.ToLower(f.Text)) {
		return true
	}

	return false
}

// Description returns a human-readable description of the filter
func (f *TextFilter) Description() string {
	return fmt.Sprintf("contains text '%s'", f.Text)
}

// AndFilter combines multiple filters with AND logic
type AndFilter struct {
	// Filters is the list of filters to combine
	Filters []Filter
}

// Apply returns true if all filters match
func (f *AndFilter) Apply(notification *github.Notification) bool {
	for _, filter := range f.Filters {
		if !filter.Apply(notification) {
			return false
		}
	}
	return true
}

// Description returns a human-readable description of the filter
func (f *AndFilter) Description() string {
	var descriptions []string
	for _, filter := range f.Filters {
		descriptions = append(descriptions, filter.Description())
	}
	return fmt.Sprintf("(%s)", strings.Join(descriptions, " AND "))
}

// OrFilter combines multiple filters with OR logic
type OrFilter struct {
	// Filters is the list of filters to combine
	Filters []Filter
}

// Apply returns true if any filter matches
func (f *OrFilter) Apply(notification *github.Notification) bool {
	for _, filter := range f.Filters {
		if filter.Apply(notification) {
			return true
		}
	}
	return false
}

// Description returns a human-readable description of the filter
func (f *OrFilter) Description() string {
	var descriptions []string
	for _, filter := range f.Filters {
		descriptions = append(descriptions, filter.Description())
	}
	return fmt.Sprintf("(%s)", strings.Join(descriptions, " OR "))
}

// NotFilter negates a filter
type NotFilter struct {
	// Filter is the filter to negate
	Filter Filter
}

// Apply returns true if the filter does not match
func (f *NotFilter) Apply(notification *github.Notification) bool {
	return !f.Filter.Apply(notification)
}

// Description returns a human-readable description of the filter
func (f *NotFilter) Description() string {
	return fmt.Sprintf("NOT (%s)", f.Filter.Description())
}
