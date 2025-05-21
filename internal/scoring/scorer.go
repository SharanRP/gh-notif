package scoring

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/google/go-github/v60/github"
)

// ScoreFactors contains the configurable factors for scoring notifications
type ScoreFactors struct {
	// Age factor weights
	AgeWeight float64 `json:"age_weight"`
	// Age decay (how quickly score decreases with age)
	AgeDecay float64 `json:"age_decay"`
	// Maximum age in hours to consider
	MaxAge float64 `json:"max_age"`

	// Activity factor weights
	ActivityWeight float64 `json:"activity_weight"`
	// Comment count weight
	CommentWeight float64 `json:"comment_weight"`
	// Reaction count weight
	ReactionWeight float64 `json:"reaction_weight"`

	// User involvement factor weights
	InvolvementWeight float64 `json:"involvement_weight"`
	// Author weight
	AuthorWeight float64 `json:"author_weight"`
	// Assignee weight
	AssigneeWeight float64 `json:"assignee_weight"`
	// Mention weight
	MentionWeight float64 `json:"mention_weight"`
	// Review request weight
	ReviewWeight float64 `json:"review_weight"`

	// Type factor weights
	TypeWeight float64 `json:"type_weight"`
	// Pull request weight
	PRWeight float64 `json:"pr_weight"`
	// Issue weight
	IssueWeight float64 `json:"issue_weight"`
	// Discussion weight
	DiscussionWeight float64 `json:"discussion_weight"`
	// Release weight
	ReleaseWeight float64 `json:"release_weight"`
	// Commit weight
	CommitWeight float64 `json:"commit_weight"`

	// Reason factor weights
	ReasonWeight float64 `json:"reason_weight"`
	// Assign reason weight
	AssignReasonWeight float64 `json:"assign_reason_weight"`
	// Author reason weight
	AuthorReasonWeight float64 `json:"author_reason_weight"`
	// Comment reason weight
	CommentReasonWeight float64 `json:"comment_reason_weight"`
	// Mention reason weight
	MentionReasonWeight float64 `json:"mention_reason_weight"`
	// Review request reason weight
	ReviewReasonWeight float64 `json:"review_reason_weight"`
	// State change reason weight
	StateChangeReasonWeight float64 `json:"state_change_reason_weight"`
	// Subscribed reason weight
	SubscribedReasonWeight float64 `json:"subscribed_reason_weight"`
	// Team mention reason weight
	TeamMentionReasonWeight float64 `json:"team_mention_reason_weight"`

	// Repository factor weights
	RepoWeight float64 `json:"repo_weight"`
	// Custom repository weights (repo name -> weight)
	CustomRepoWeights map[string]float64 `json:"custom_repo_weights"`
}

// DefaultScoreFactors returns the default score factors
func DefaultScoreFactors() *ScoreFactors {
	return &ScoreFactors{
		// Age factors
		AgeWeight: 0.3,
		AgeDecay:  0.1,
		MaxAge:    168, // 7 days

		// Activity factors
		ActivityWeight: 0.2,
		CommentWeight:  0.6,
		ReactionWeight: 0.4,

		// User involvement factors
		InvolvementWeight: 0.3,
		AuthorWeight:      0.8,
		AssigneeWeight:    0.9,
		MentionWeight:     0.7,
		ReviewWeight:      0.6,

		// Type factors
		TypeWeight:       0.1,
		PRWeight:         0.8,
		IssueWeight:      0.7,
		DiscussionWeight: 0.6,
		ReleaseWeight:    0.5,
		CommitWeight:     0.4,

		// Reason factors
		ReasonWeight:            0.1,
		AssignReasonWeight:      0.9,
		AuthorReasonWeight:      0.8,
		CommentReasonWeight:     0.7,
		MentionReasonWeight:     0.8,
		ReviewReasonWeight:      0.8,
		StateChangeReasonWeight: 0.6,
		SubscribedReasonWeight:  0.5,
		TeamMentionReasonWeight: 0.7,

		// Repository factors
		RepoWeight:        0.1,
		CustomRepoWeights: make(map[string]float64),
	}
}

// NotificationScore represents the score of a notification
type NotificationScore struct {
	// Total is the total score (0-100)
	Total int `json:"total"`
	// Components are the individual score components
	Components map[string]float64 `json:"components"`
	// Factors are the factors used to calculate the score
	Factors *ScoreFactors `json:"factors"`
}

// Scorer calculates scores for notifications
type Scorer struct {
	// Factors are the configurable factors for scoring
	Factors *ScoreFactors
	// Concurrency is the number of goroutines to use
	Concurrency int
	// BatchSize is the size of notification batches
	BatchSize int
	// Timeout is the maximum time to spend scoring
	Timeout time.Duration
	// Username is the current user's GitHub username
	Username string
}

// NewScorer creates a new scorer
func NewScorer(factors *ScoreFactors) *Scorer {
	if factors == nil {
		factors = DefaultScoreFactors()
	}

	return &Scorer{
		Factors:     factors,
		Concurrency: 5,
		BatchSize:   100,
		Timeout:     5 * time.Second,
	}
}

// WithUsername sets the username for the scorer
func (s *Scorer) WithUsername(username string) *Scorer {
	s.Username = username
	return s
}

// WithConcurrency sets the concurrency for the scorer
func (s *Scorer) WithConcurrency(concurrency int) *Scorer {
	if concurrency > 0 {
		s.Concurrency = concurrency
	}
	return s
}

// WithBatchSize sets the batch size for the scorer
func (s *Scorer) WithBatchSize(batchSize int) *Scorer {
	if batchSize > 0 {
		s.BatchSize = batchSize
	}
	return s
}

// WithTimeout sets the timeout for the scorer
func (s *Scorer) WithTimeout(timeout time.Duration) *Scorer {
	if timeout > 0 {
		s.Timeout = timeout
	}
	return s
}

// Score calculates scores for notifications
func (s *Scorer) Score(ctx context.Context, notifications []*github.Notification) (map[string]*NotificationScore, error) {
	if len(notifications) == 0 {
		return nil, nil
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	// For small sets, don't bother with concurrency
	if len(notifications) < s.BatchSize {
		return s.scoreSequential(notifications), nil
	}

	return s.scoreConcurrent(ctx, notifications)
}

// scoreSequential calculates scores sequentially
func (s *Scorer) scoreSequential(notifications []*github.Notification) map[string]*NotificationScore {
	scores := make(map[string]*NotificationScore, len(notifications))
	for _, n := range notifications {
		scores[n.GetID()] = s.scoreNotification(n)
	}
	return scores
}

// scoreConcurrent calculates scores concurrently
func (s *Scorer) scoreConcurrent(ctx context.Context, notifications []*github.Notification) (map[string]*NotificationScore, error) {
	// Create channels for input and output
	input := make(chan *github.Notification, s.BatchSize)
	output := make(chan struct {
		id    string
		score *NotificationScore
	}, s.BatchSize)
	done := make(chan struct{})

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < s.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := range input {
				score := s.scoreNotification(n)
				select {
				case output <- struct {
					id    string
					score *NotificationScore
				}{id: n.GetID(), score: score}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Start a goroutine to close the output channel when all workers are done
	go func() {
		wg.Wait()
		close(output)
		close(done)
	}()

	// Feed input channel with notifications
	go func() {
		defer close(input)
		for _, n := range notifications {
			select {
			case input <- n:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	scores := make(map[string]*NotificationScore, len(notifications))
	for {
		select {
		case result, ok := <-output:
			if !ok {
				return scores, nil
			}
			scores[result.id] = result.score
		case <-ctx.Done():
			return scores, ctx.Err()
		case <-done:
			return scores, nil
		}
	}
}

// scoreNotification calculates a score for a notification
func (s *Scorer) scoreNotification(n *github.Notification) *NotificationScore {
	// Create a score object
	score := &NotificationScore{
		Components: make(map[string]float64),
		Factors:    s.Factors,
	}

	// Calculate age score
	ageScore := s.calculateAgeScore(n)
	score.Components["age"] = ageScore

	// Calculate activity score
	activityScore := s.calculateActivityScore(n)
	score.Components["activity"] = activityScore

	// Calculate user involvement score
	involvementScore := s.calculateInvolvementScore(n)
	score.Components["involvement"] = involvementScore

	// Calculate type score
	typeScore := s.calculateTypeScore(n)
	score.Components["type"] = typeScore

	// Calculate reason score
	reasonScore := s.calculateReasonScore(n)
	score.Components["reason"] = reasonScore

	// Calculate repository score
	repoScore := s.calculateRepoScore(n)
	score.Components["repository"] = repoScore

	// Calculate total score (0-100)
	totalScore := ageScore + activityScore + involvementScore + typeScore + reasonScore + repoScore
	score.Total = int(math.Min(100, math.Max(0, totalScore*100)))

	return score
}

// calculateAgeScore calculates the age component of the score
func (s *Scorer) calculateAgeScore(n *github.Notification) float64 {
	// Get the age in hours
	age := time.Since(n.GetUpdatedAt().Time).Hours()

	// Cap the age at the maximum
	if age > s.Factors.MaxAge {
		age = s.Factors.MaxAge
	}

	// Calculate the age score (newer is better)
	// Use an exponential decay function
	ageScore := math.Exp(-s.Factors.AgeDecay * age / 24)

	// Apply the age weight
	return ageScore * s.Factors.AgeWeight
}

// calculateActivityScore calculates the activity component of the score
func (s *Scorer) calculateActivityScore(n *github.Notification) float64 {
	// This would ideally use the number of comments and reactions
	// For now, use a placeholder based on the updated time
	activityScore := 0.5

	// Apply the activity weight
	return activityScore * s.Factors.ActivityWeight
}

// calculateInvolvementScore calculates the user involvement component of the score
func (s *Scorer) calculateInvolvementScore(n *github.Notification) float64 {
	// This would ideally check if the user is the author, assignee, mentioned, etc.
	// For now, use a placeholder based on the reason
	involvementScore := 0.5

	// Apply the involvement weight
	return involvementScore * s.Factors.InvolvementWeight
}

// calculateTypeScore calculates the type component of the score
func (s *Scorer) calculateTypeScore(n *github.Notification) float64 {
	// Get the notification type
	typ := n.GetSubject().GetType()

	// Calculate the type score
	var typeScore float64
	switch typ {
	case "PullRequest":
		typeScore = s.Factors.PRWeight
	case "Issue":
		typeScore = s.Factors.IssueWeight
	case "Discussion":
		typeScore = s.Factors.DiscussionWeight
	case "Release":
		typeScore = s.Factors.ReleaseWeight
	case "Commit":
		typeScore = s.Factors.CommitWeight
	default:
		typeScore = 0.5
	}

	// Apply the type weight
	return typeScore * s.Factors.TypeWeight
}

// calculateReasonScore calculates the reason component of the score
func (s *Scorer) calculateReasonScore(n *github.Notification) float64 {
	// Get the notification reason
	reason := n.GetReason()

	// Calculate the reason score
	var reasonScore float64
	switch reason {
	case "assign":
		reasonScore = s.Factors.AssignReasonWeight
	case "author":
		reasonScore = s.Factors.AuthorReasonWeight
	case "comment":
		reasonScore = s.Factors.CommentReasonWeight
	case "mention":
		reasonScore = s.Factors.MentionReasonWeight
	case "review_requested":
		reasonScore = s.Factors.ReviewReasonWeight
	case "state_change":
		reasonScore = s.Factors.StateChangeReasonWeight
	case "subscribed":
		reasonScore = s.Factors.SubscribedReasonWeight
	case "team_mention":
		reasonScore = s.Factors.TeamMentionReasonWeight
	default:
		reasonScore = 0.5
	}

	// Apply the reason weight
	return reasonScore * s.Factors.ReasonWeight
}

// calculateRepoScore calculates the repository component of the score
func (s *Scorer) calculateRepoScore(n *github.Notification) float64 {
	// Get the repository name
	repo := n.GetRepository().GetFullName()

	// Check for a custom weight
	if weight, ok := s.Factors.CustomRepoWeights[repo]; ok {
		return weight * s.Factors.RepoWeight
	}

	// Default repository score
	return 0.5 * s.Factors.RepoWeight
}
