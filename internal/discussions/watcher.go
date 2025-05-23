package discussions

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/user/gh-notif/internal/config"
)

// DiscussionWatcher monitors discussions for changes and events
type DiscussionWatcher struct {
	client        *Client
	analytics     *AnalyticsEngine
	searchEngine  *SearchEngine
	configManager *config.ConfigManager
	
	// Watching state
	isWatching    bool
	stopChan      chan struct{}
	eventChan     chan DiscussionEvent
	errorChan     chan error
	
	// Configuration
	repositories  []string
	interval      time.Duration
	filters       DiscussionFilter
	options       DiscussionOptions
	
	// Callbacks
	eventCallback func(DiscussionEvent)
	errorCallback func(error)
	
	// Internal state
	mu            sync.RWMutex
	lastCheck     time.Time
	knownDiscussions map[string]*Discussion
	debug         bool
}

// WatcherOptions contains configuration for the discussion watcher
type WatcherOptions struct {
	// Repositories to watch
	Repositories []string `json:"repositories"`
	
	// Check interval
	Interval time.Duration `json:"interval"`
	
	// Filters to apply
	Filters DiscussionFilter `json:"filters"`
	
	// Discussion options
	Options DiscussionOptions `json:"options"`
	
	// Event callback
	EventCallback func(DiscussionEvent) `json:"-"`
	
	// Error callback
	ErrorCallback func(error) `json:"-"`
	
	// Enable debug logging
	Debug bool `json:"debug"`
}

// NewDiscussionWatcher creates a new discussion watcher
func NewDiscussionWatcher(client *Client, analytics *AnalyticsEngine, searchEngine *SearchEngine) *DiscussionWatcher {
	configManager := config.NewConfigManager()
	configManager.Load()

	return &DiscussionWatcher{
		client:           client,
		analytics:        analytics,
		searchEngine:     searchEngine,
		configManager:    configManager,
		stopChan:         make(chan struct{}),
		eventChan:        make(chan DiscussionEvent, 100),
		errorChan:        make(chan error, 10),
		knownDiscussions: make(map[string]*Discussion),
		interval:         5 * time.Minute, // Default interval
	}
}

// Start begins watching for discussion changes
func (dw *DiscussionWatcher) Start(ctx context.Context, options WatcherOptions) error {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	if dw.isWatching {
		return fmt.Errorf("watcher is already running")
	}

	// Configure the watcher
	dw.repositories = options.Repositories
	dw.filters = options.Filters
	dw.options = options.Options
	dw.eventCallback = options.EventCallback
	dw.errorCallback = options.ErrorCallback
	dw.debug = options.Debug

	if options.Interval > 0 {
		dw.interval = options.Interval
	}

	// Initialize known discussions
	if err := dw.initializeKnownDiscussions(ctx); err != nil {
		return fmt.Errorf("failed to initialize known discussions: %w", err)
	}

	dw.isWatching = true
	dw.lastCheck = time.Now()

	// Start the watching goroutine
	go dw.watchLoop(ctx)

	if dw.debug {
		fmt.Printf("Discussion watcher started for %d repositories\n", len(dw.repositories))
	}

	return nil
}

// Stop stops the discussion watcher
func (dw *DiscussionWatcher) Stop() {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	if !dw.isWatching {
		return
	}

	close(dw.stopChan)
	dw.isWatching = false

	if dw.debug {
		fmt.Println("Discussion watcher stopped")
	}
}

// IsWatching returns whether the watcher is currently running
func (dw *DiscussionWatcher) IsWatching() bool {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	return dw.isWatching
}

// GetEventChannel returns the channel for discussion events
func (dw *DiscussionWatcher) GetEventChannel() <-chan DiscussionEvent {
	return dw.eventChan
}

// GetErrorChannel returns the channel for errors
func (dw *DiscussionWatcher) GetErrorChannel() <-chan error {
	return dw.errorChan
}

// GetStats returns statistics about the watcher
func (dw *DiscussionWatcher) GetStats() map[string]interface{} {
	dw.mu.RLock()
	defer dw.mu.RUnlock()

	return map[string]interface{}{
		"is_watching":       dw.isWatching,
		"repositories":      len(dw.repositories),
		"known_discussions": len(dw.knownDiscussions),
		"last_check":        dw.lastCheck,
		"interval":          dw.interval,
	}
}

// watchLoop is the main watching loop
func (dw *DiscussionWatcher) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(dw.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dw.stopChan:
			return
		case <-ticker.C:
			if err := dw.checkForChanges(ctx); err != nil {
				dw.handleError(err)
			}
		}
	}
}

// checkForChanges checks for new or updated discussions
func (dw *DiscussionWatcher) checkForChanges(ctx context.Context) error {
	if dw.debug {
		fmt.Println("Checking for discussion changes...")
	}

	// Fetch current discussions
	currentDiscussions, err := dw.client.GetDiscussions(ctx, dw.repositories, dw.filters, dw.options)
	if err != nil {
		return fmt.Errorf("failed to fetch discussions: %w", err)
	}

	dw.mu.Lock()
	defer dw.mu.Unlock()

	// Process each discussion
	for _, discussion := range currentDiscussions {
		if known, exists := dw.knownDiscussions[discussion.ID]; exists {
			// Check for updates
			if dw.hasDiscussionChanged(known, &discussion) {
				dw.handleDiscussionUpdate(known, &discussion)
			}
		} else {
			// New discussion
			dw.handleNewDiscussion(&discussion)
		}

		// Update known discussions
		discussionCopy := discussion
		dw.knownDiscussions[discussion.ID] = &discussionCopy
	}

	// Check for deleted discussions (discussions that were known but not in current list)
	currentIDs := make(map[string]bool)
	for _, discussion := range currentDiscussions {
		currentIDs[discussion.ID] = true
	}

	for id, discussion := range dw.knownDiscussions {
		if !currentIDs[id] {
			dw.handleDiscussionDeleted(discussion)
			delete(dw.knownDiscussions, id)
		}
	}

	dw.lastCheck = time.Now()

	if dw.debug {
		fmt.Printf("Processed %d discussions, %d known\n", len(currentDiscussions), len(dw.knownDiscussions))
	}

	return nil
}

// initializeKnownDiscussions loads initial discussions
func (dw *DiscussionWatcher) initializeKnownDiscussions(ctx context.Context) error {
	discussions, err := dw.client.GetDiscussions(ctx, dw.repositories, dw.filters, dw.options)
	if err != nil {
		return err
	}

	dw.knownDiscussions = make(map[string]*Discussion)
	for _, discussion := range discussions {
		discussionCopy := discussion
		dw.knownDiscussions[discussion.ID] = &discussionCopy
	}

	if dw.debug {
		fmt.Printf("Initialized with %d known discussions\n", len(dw.knownDiscussions))
	}

	return nil
}

// hasDiscussionChanged checks if a discussion has been updated
func (dw *DiscussionWatcher) hasDiscussionChanged(old, new *Discussion) bool {
	// Check basic fields
	if old.Title != new.Title ||
	   old.Body != new.Body ||
	   old.State != new.State ||
	   old.Locked != new.Locked ||
	   old.UpvoteCount != new.UpvoteCount ||
	   old.CommentCount != new.CommentCount ||
	   old.ReactionCount != new.ReactionCount {
		return true
	}

	// Check timestamps
	if !old.UpdatedAt.Equal(new.UpdatedAt) {
		return true
	}

	// Check answer status
	if (old.Answer == nil) != (new.Answer == nil) {
		return true
	}

	if old.Answer != nil && new.Answer != nil {
		if old.Answer.ID != new.Answer.ID {
			return true
		}
	}

	// Check labels
	if len(old.Labels) != len(new.Labels) {
		return true
	}

	oldLabels := make(map[string]bool)
	for _, label := range old.Labels {
		oldLabels[label.ID] = true
	}

	for _, label := range new.Labels {
		if !oldLabels[label.ID] {
			return true
		}
	}

	return false
}

// handleNewDiscussion processes a new discussion
func (dw *DiscussionWatcher) handleNewDiscussion(discussion *Discussion) {
	event := DiscussionEvent{
		Type:       EventDiscussionCreated,
		Discussion: discussion,
		User:       discussion.Author,
		Timestamp:  time.Now(),
		Repository: discussion.Repository,
	}

	dw.sendEvent(event)

	if dw.debug {
		fmt.Printf("New discussion: %s #%d\n", discussion.Repository.FullName, discussion.Number)
	}
}

// handleDiscussionUpdate processes a discussion update
func (dw *DiscussionWatcher) handleDiscussionUpdate(old, new *Discussion) {
	// Determine what changed
	changes := make(map[string]interface{})

	if old.Title != new.Title {
		changes["title"] = map[string]string{"old": old.Title, "new": new.Title}
	}

	if old.Body != new.Body {
		changes["body"] = map[string]string{"old": old.Body, "new": new.Body}
	}

	if old.State != new.State {
		changes["state"] = map[string]string{"old": old.State, "new": new.State}
		
		// Send specific state change events
		if new.State == "CLOSED" {
			dw.sendEvent(DiscussionEvent{
				Type:          EventDiscussionClosed,
				Discussion:    new,
				User:          new.Author, // This should ideally be the user who closed it
				Timestamp:     time.Now(),
				Repository:    new.Repository,
				PreviousState: old.State,
				NewState:      new.State,
			})
		} else if new.State == "OPEN" && old.State == "CLOSED" {
			dw.sendEvent(DiscussionEvent{
				Type:          EventDiscussionReopened,
				Discussion:    new,
				User:          new.Author,
				Timestamp:     time.Now(),
				Repository:    new.Repository,
				PreviousState: old.State,
				NewState:      new.State,
			})
		}
	}

	if old.Locked != new.Locked {
		changes["locked"] = map[string]bool{"old": old.Locked, "new": new.Locked}
		
		if new.Locked {
			dw.sendEvent(DiscussionEvent{
				Type:       EventDiscussionLocked,
				Discussion: new,
				User:       new.Author,
				Timestamp:  time.Now(),
				Repository: new.Repository,
			})
		} else {
			dw.sendEvent(DiscussionEvent{
				Type:       EventDiscussionUnlocked,
				Discussion: new,
				User:       new.Author,
				Timestamp:  time.Now(),
				Repository: new.Repository,
			})
		}
	}

	// Check for answer changes
	if (old.Answer == nil) != (new.Answer == nil) {
		if new.Answer != nil {
			dw.sendEvent(DiscussionEvent{
				Type:       EventCommentMarkedAsAnswer,
				Discussion: new,
				Comment:    new.Answer,
				User:       new.Answer.Author,
				Timestamp:  time.Now(),
				Repository: new.Repository,
			})
		}
	}

	// Send general update event if there were changes
	if len(changes) > 0 {
		event := DiscussionEvent{
			Type:       EventDiscussionUpdated,
			Discussion: new,
			User:       new.Author,
			Timestamp:  time.Now(),
			Repository: new.Repository,
			Changes:    changes,
		}

		dw.sendEvent(event)

		if dw.debug {
			fmt.Printf("Updated discussion: %s #%d (%d changes)\n", 
				new.Repository.FullName, new.Number, len(changes))
		}
	}
}

// handleDiscussionDeleted processes a deleted discussion
func (dw *DiscussionWatcher) handleDiscussionDeleted(discussion *Discussion) {
	event := DiscussionEvent{
		Type:       EventDiscussionDeleted,
		Discussion: discussion,
		User:       discussion.Author,
		Timestamp:  time.Now(),
		Repository: discussion.Repository,
	}

	dw.sendEvent(event)

	if dw.debug {
		fmt.Printf("Deleted discussion: %s #%d\n", discussion.Repository.FullName, discussion.Number)
	}
}

// sendEvent sends an event through the appropriate channels
func (dw *DiscussionWatcher) sendEvent(event DiscussionEvent) {
	// Send to event channel (non-blocking)
	select {
	case dw.eventChan <- event:
	default:
		// Channel is full, log warning
		if dw.debug {
			fmt.Println("Warning: Event channel is full, dropping event")
		}
	}

	// Call callback if provided
	if dw.eventCallback != nil {
		go dw.eventCallback(event)
	}
}

// handleError handles errors from the watcher
func (dw *DiscussionWatcher) handleError(err error) {
	// Send to error channel (non-blocking)
	select {
	case dw.errorChan <- err:
	default:
		// Channel is full, log warning
		if dw.debug {
			fmt.Printf("Warning: Error channel is full, dropping error: %v\n", err)
		}
	}

	// Call callback if provided
	if dw.errorCallback != nil {
		go dw.errorCallback(err)
	}

	if dw.debug {
		fmt.Printf("Watcher error: %v\n", err)
	}
}
