package subscriptions

import (
	"time"
)

// Priority levels for repository subscriptions
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityCritical
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityCritical:
		return "critical"
	default:
		return "normal"
	}
}

// ParsePriority parses a priority string
func ParsePriority(s string) Priority {
	switch s {
	case "low":
		return PriorityLow
	case "normal":
		return PriorityNormal
	case "critical":
		return PriorityCritical
	default:
		return PriorityNormal
	}
}

// Frequency defines notification frequency
type Frequency int

const (
	FrequencyRealTime Frequency = iota
	FrequencyHourly
	FrequencyDaily
)

func (f Frequency) String() string {
	switch f {
	case FrequencyRealTime:
		return "real-time"
	case FrequencyHourly:
		return "hourly"
	case FrequencyDaily:
		return "daily"
	default:
		return "real-time"
	}
}

// ParseFrequency parses a frequency string
func ParseFrequency(s string) Frequency {
	switch s {
	case "real-time":
		return FrequencyRealTime
	case "hourly":
		return FrequencyHourly
	case "daily":
		return FrequencyDaily
	default:
		return FrequencyRealTime
	}
}

// ActivityType represents different types of repository activities
type ActivityType string

const (
	ActivityCommits     ActivityType = "commits"
	ActivityBranches    ActivityType = "branches"
	ActivityPRs         ActivityType = "prs"
	ActivityIssues      ActivityType = "issues"
	ActivityDiscussions ActivityType = "discussions"
	ActivityReleases    ActivityType = "releases"
	ActivityWiki        ActivityType = "wiki"
	ActivitySecurity    ActivityType = "security"
)

// AllActivityTypes returns all available activity types
func AllActivityTypes() []ActivityType {
	return []ActivityType{
		ActivityCommits,
		ActivityBranches,
		ActivityPRs,
		ActivityIssues,
		ActivityDiscussions,
		ActivityReleases,
		ActivityWiki,
		ActivitySecurity,
	}
}

// BranchFilter defines branch filtering options
type BranchFilter struct {
	// All branches if true, otherwise use patterns
	All bool `json:"all" yaml:"all"`
	
	// Include main/master branches
	MainOnly bool `json:"main_only" yaml:"main_only"`
	
	// Specific branch patterns (glob patterns)
	Patterns []string `json:"patterns" yaml:"patterns"`
	
	// Exclude patterns
	ExcludePatterns []string `json:"exclude_patterns" yaml:"exclude_patterns"`
}

// AuthorFilter defines author filtering options
type AuthorFilter struct {
	// All contributors if true
	All bool `json:"all" yaml:"all"`
	
	// Specific usernames to include
	Include []string `json:"include" yaml:"include"`
	
	// Specific usernames to exclude
	Exclude []string `json:"exclude" yaml:"exclude"`
	
	// Exclude bots
	ExcludeBots bool `json:"exclude_bots" yaml:"exclude_bots"`
}

// FileFilter defines file pattern filtering
type FileFilter struct {
	// Include patterns (glob patterns)
	Include []string `json:"include" yaml:"include"`
	
	// Exclude patterns (glob patterns)
	Exclude []string `json:"exclude" yaml:"exclude"`
	
	// File extensions to include
	Extensions []string `json:"extensions" yaml:"extensions"`
	
	// Paths to include
	Paths []string `json:"paths" yaml:"paths"`
}

// SubscriptionConfig holds the configuration for a repository subscription
type SubscriptionConfig struct {
	// Activity types to monitor
	ActivityTypes []ActivityType `json:"activity_types" yaml:"activity_types"`
	
	// Notification frequency
	Frequency Frequency `json:"frequency" yaml:"frequency"`
	
	// Branch filtering
	BranchFilter BranchFilter `json:"branch_filter" yaml:"branch_filter"`
	
	// Author filtering
	AuthorFilter AuthorFilter `json:"author_filter" yaml:"author_filter"`
	
	// File pattern filtering
	FileFilter FileFilter `json:"file_filter" yaml:"file_filter"`
	
	// Custom webhook URL (optional)
	WebhookURL string `json:"webhook_url,omitempty" yaml:"webhook_url,omitempty"`
	
	// Custom notification template (optional)
	Template string `json:"template,omitempty" yaml:"template,omitempty"`
}

// DefaultSubscriptionConfig returns a default subscription configuration
func DefaultSubscriptionConfig() SubscriptionConfig {
	return SubscriptionConfig{
		ActivityTypes: []ActivityType{
			ActivityPRs,
			ActivityIssues,
			ActivityReleases,
		},
		Frequency: FrequencyRealTime,
		BranchFilter: BranchFilter{
			All: true,
		},
		AuthorFilter: AuthorFilter{
			All:         true,
			ExcludeBots: true,
		},
		FileFilter: FileFilter{
			Include: []string{"*"},
		},
	}
}

// RepositorySubscription represents a subscription to a repository
type RepositorySubscription struct {
	// Repository full name (owner/repo) or pattern (owner/*)
	Repository string `json:"repository" yaml:"repository"`
	
	// Whether this is a wildcard pattern
	IsPattern bool `json:"is_pattern" yaml:"is_pattern"`
	
	// Priority level
	Priority Priority `json:"priority" yaml:"priority"`
	
	// Subscription configuration
	Config SubscriptionConfig `json:"config" yaml:"config"`
	
	// Whether the subscription is active
	Active bool `json:"active" yaml:"active"`
	
	// Creation timestamp
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	
	// Last updated timestamp
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	
	// Last access check (for permission validation)
	LastAccessCheck time.Time `json:"last_access_check" yaml:"last_access_check"`
	
	// Whether we have access to this repository
	HasAccess bool `json:"has_access" yaml:"has_access"`
	
	// Access error message if any
	AccessError string `json:"access_error,omitempty" yaml:"access_error,omitempty"`
	
	// Metadata for additional information
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SubscriptionList represents a collection of repository subscriptions
type SubscriptionList struct {
	// Version for compatibility
	Version string `json:"version" yaml:"version"`
	
	// Subscriptions
	Subscriptions []RepositorySubscription `json:"subscriptions" yaml:"subscriptions"`
	
	// Last updated timestamp
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
	
	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// SubscriptionStats provides statistics about subscriptions
type SubscriptionStats struct {
	Total          int `json:"total"`
	Active         int `json:"active"`
	Inactive       int `json:"inactive"`
	Patterns       int `json:"patterns"`
	Repositories   int `json:"repositories"`
	Critical       int `json:"critical"`
	Normal         int `json:"normal"`
	Low            int `json:"low"`
	AccessErrors   int `json:"access_errors"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ValidationError represents a subscription validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidationResult holds the result of subscription validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}
