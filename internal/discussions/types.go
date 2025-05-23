package discussions

import (
	"time"
)

// Discussion represents a GitHub Discussion
type Discussion struct {
	ID          string    `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	BodyHTML    string    `json:"body_html"`
	BodyText    string    `json:"body_text"`
	URL         string    `json:"url"`
	State       string    `json:"state"` // OPEN, CLOSED
	Locked      bool      `json:"locked"`
	Repository  Repository `json:"repository"`
	Category    Category  `json:"category"`
	Author      User      `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at,omitempty"`
	
	// Engagement metrics
	UpvoteCount   int `json:"upvote_count"`
	CommentCount  int `json:"comment_count"`
	ReactionCount int `json:"reaction_count"`
	
	// Answer information
	Answer        *Comment  `json:"answer,omitempty"`
	AnsweredAt    *time.Time `json:"answered_at,omitempty"`
	AnsweredBy    *User     `json:"answered_by,omitempty"`
	
	// Labels and assignments
	Labels     []Label `json:"labels"`
	Assignees  []User  `json:"assignees"`
	
	// Participation tracking
	ViewerDidAuthor    bool `json:"viewer_did_author"`
	ViewerSubscription string `json:"viewer_subscription"` // SUBSCRIBED, UNSUBSCRIBED, IGNORED
	ViewerCanReact     bool `json:"viewer_can_react"`
	ViewerCanUpdate    bool `json:"viewer_can_update"`
	ViewerCanDelete    bool `json:"viewer_can_delete"`
	
	// Comments (loaded separately for performance)
	Comments []Comment `json:"comments,omitempty"`
}

// Comment represents a discussion comment
type Comment struct {
	ID          string    `json:"id"`
	Body        string    `json:"body"`
	BodyHTML    string    `json:"body_html"`
	BodyText    string    `json:"body_text"`
	URL         string    `json:"url"`
	Author      User      `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Threading
	ParentID     *string    `json:"parent_id,omitempty"`
	ReplyTo      *Comment   `json:"reply_to,omitempty"`
	Replies      []Comment  `json:"replies,omitempty"`
	
	// Answer marking
	IsAnswer     bool       `json:"is_answer"`
	MarkedAsAnswerAt *time.Time `json:"marked_as_answer_at,omitempty"`
	MarkedAsAnswerBy *User      `json:"marked_as_answer_by,omitempty"`
	
	// Engagement
	UpvoteCount   int        `json:"upvote_count"`
	ReactionCount int        `json:"reaction_count"`
	Reactions     []Reaction `json:"reactions,omitempty"`
	
	// Viewer permissions
	ViewerDidAuthor bool `json:"viewer_did_author"`
	ViewerCanReact  bool `json:"viewer_can_react"`
	ViewerCanUpdate bool `json:"viewer_can_update"`
	ViewerCanDelete bool `json:"viewer_can_delete"`
	ViewerCanMarkAsAnswer bool `json:"viewer_can_mark_as_answer"`
}

// Category represents a discussion category
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Emoji       string `json:"emoji"`
	Slug        string `json:"slug"`
	IsAnswerable bool  `json:"is_answerable"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Repository represents repository information
type Repository struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    User   `json:"owner"`
	URL      string `json:"url"`
	Private  bool   `json:"private"`
}

// User represents a GitHub user
type User struct {
	ID        string `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	Type      string `json:"type"` // User, Bot, Organization
}

// Label represents a discussion label
type Label struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	URL         string `json:"url"`
}

// Reaction represents a reaction to a discussion or comment
type Reaction struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"` // +1, -1, laugh, hooray, confused, heart, rocket, eyes
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

// DiscussionFilter represents filtering options for discussions
type DiscussionFilter struct {
	// Basic filters
	Repository   string   `json:"repository,omitempty"`
	Category     string   `json:"category,omitempty"`
	State        string   `json:"state,omitempty"` // open, closed, all
	Author       string   `json:"author,omitempty"`
	Assignee     string   `json:"assignee,omitempty"`
	Labels       []string `json:"labels,omitempty"`
	
	// Content filters
	Query        string   `json:"query,omitempty"`
	Title        string   `json:"title,omitempty"`
	Body         string   `json:"body,omitempty"`
	
	// Time filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`
	
	// Engagement filters
	MinUpvotes    int `json:"min_upvotes,omitempty"`
	MinComments   int `json:"min_comments,omitempty"`
	MinReactions  int `json:"min_reactions,omitempty"`
	
	// Answer filters
	Answered      *bool `json:"answered,omitempty"`
	HasAnswer     *bool `json:"has_answer,omitempty"`
	
	// Participation filters
	Participating bool `json:"participating,omitempty"`
	Mentioned     bool `json:"mentioned,omitempty"`
	Subscribed    bool `json:"subscribed,omitempty"`
	
	// Sorting and pagination
	Sort      string `json:"sort,omitempty"`      // created, updated, comments, reactions
	Direction string `json:"direction,omitempty"` // asc, desc
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// DiscussionOptions represents options for discussion operations
type DiscussionOptions struct {
	// Fetching options
	IncludeComments  bool `json:"include_comments,omitempty"`
	IncludeReactions bool `json:"include_reactions,omitempty"`
	IncludeLabels    bool `json:"include_labels,omitempty"`
	MaxComments      int  `json:"max_comments,omitempty"`
	
	// Caching options
	UseCache  bool          `json:"use_cache,omitempty"`
	CacheTTL  time.Duration `json:"cache_ttl,omitempty"`
	
	// Performance options
	Concurrency int `json:"concurrency,omitempty"`
	Timeout     time.Duration `json:"timeout,omitempty"`
}

// DiscussionEvent represents a discussion-related event
type DiscussionEvent struct {
	Type        EventType   `json:"type"`
	Discussion  *Discussion `json:"discussion,omitempty"`
	Comment     *Comment    `json:"comment,omitempty"`
	User        User        `json:"user"`
	Timestamp   time.Time   `json:"timestamp"`
	Repository  Repository  `json:"repository"`
	
	// Event-specific data
	Changes     map[string]interface{} `json:"changes,omitempty"`
	PreviousState string               `json:"previous_state,omitempty"`
	NewState      string               `json:"new_state,omitempty"`
}

// EventType represents the type of discussion event
type EventType string

const (
	EventDiscussionCreated     EventType = "discussion_created"
	EventDiscussionUpdated     EventType = "discussion_updated"
	EventDiscussionClosed      EventType = "discussion_closed"
	EventDiscussionReopened    EventType = "discussion_reopened"
	EventDiscussionDeleted     EventType = "discussion_deleted"
	EventDiscussionLocked      EventType = "discussion_locked"
	EventDiscussionUnlocked    EventType = "discussion_unlocked"
	EventDiscussionLabeled     EventType = "discussion_labeled"
	EventDiscussionUnlabeled   EventType = "discussion_unlabeled"
	EventDiscussionCategorized EventType = "discussion_categorized"
	
	EventCommentCreated        EventType = "comment_created"
	EventCommentUpdated        EventType = "comment_updated"
	EventCommentDeleted        EventType = "comment_deleted"
	EventCommentMarkedAsAnswer EventType = "comment_marked_as_answer"
	EventCommentUnmarkedAsAnswer EventType = "comment_unmarked_as_answer"
	
	EventReactionAdded         EventType = "reaction_added"
	EventReactionRemoved       EventType = "reaction_removed"
	
	EventSubscribed            EventType = "subscribed"
	EventUnsubscribed          EventType = "unsubscribed"
)

// DiscussionAnalytics represents analytics data for discussions
type DiscussionAnalytics struct {
	Repository     Repository `json:"repository"`
	TimeRange      TimeRange  `json:"time_range"`
	
	// Overall metrics
	TotalDiscussions    int `json:"total_discussions"`
	OpenDiscussions     int `json:"open_discussions"`
	ClosedDiscussions   int `json:"closed_discussions"`
	AnsweredDiscussions int `json:"answered_discussions"`
	
	// Engagement metrics
	TotalComments       int     `json:"total_comments"`
	TotalReactions      int     `json:"total_reactions"`
	TotalUpvotes        int     `json:"total_upvotes"`
	AverageComments     float64 `json:"average_comments"`
	AverageReactions    float64 `json:"average_reactions"`
	AverageUpvotes      float64 `json:"average_upvotes"`
	
	// Time metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
	
	// Category breakdown
	CategoryStats map[string]CategoryStats `json:"category_stats"`
	
	// Top contributors
	TopAuthors      []UserStats `json:"top_authors"`
	TopCommenters   []UserStats `json:"top_commenters"`
	TopReactors     []UserStats `json:"top_reactors"`
	
	// Trending topics
	TrendingTopics  []TopicStats `json:"trending_topics"`
	PopularLabels   []LabelStats `json:"popular_labels"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CategoryStats represents statistics for a discussion category
type CategoryStats struct {
	Category        Category `json:"category"`
	DiscussionCount int      `json:"discussion_count"`
	CommentCount    int      `json:"comment_count"`
	ReactionCount   int      `json:"reaction_count"`
	AnswerRate      float64  `json:"answer_rate"`
}

// UserStats represents user engagement statistics
type UserStats struct {
	User            User `json:"user"`
	DiscussionCount int  `json:"discussion_count"`
	CommentCount    int  `json:"comment_count"`
	ReactionCount   int  `json:"reaction_count"`
	UpvoteCount     int  `json:"upvote_count"`
}

// TopicStats represents trending topic statistics
type TopicStats struct {
	Topic           string  `json:"topic"`
	DiscussionCount int     `json:"discussion_count"`
	CommentCount    int     `json:"comment_count"`
	TrendScore      float64 `json:"trend_score"`
}

// LabelStats represents label usage statistics
type LabelStats struct {
	Label      Label `json:"label"`
	UsageCount int   `json:"usage_count"`
}
