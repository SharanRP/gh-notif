package persistent

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"github.com/SharanRP/gh-notif/internal/filter"
	"github.com/SharanRP/gh-notif/internal/scoring"
)

// ScoreFilter filters notifications by score
type ScoreFilter struct {
	// MinScore is the minimum score to match (inclusive)
	MinScore int
	// MaxScore is the maximum score to match (inclusive)
	MaxScore int
	// Scorer is the scorer to use
	Scorer *scoring.Scorer
	// Scores is a cache of notification scores
	Scores map[string]*scoring.NotificationScore
}

// NewScoreFilter creates a new score filter
func NewScoreFilter(minScore, maxScore int, scorer *scoring.Scorer) *ScoreFilter {
	if scorer == nil {
		scorer = scoring.NewScorer(nil)
	}

	return &ScoreFilter{
		MinScore: minScore,
		MaxScore: maxScore,
		Scorer:   scorer,
		Scores:   make(map[string]*scoring.NotificationScore),
	}
}

// Apply applies the score filter to a notification
func (f *ScoreFilter) Apply(n *github.Notification) bool {
	// Get the notification score
	score, ok := f.Scores[n.GetID()]
	if !ok {
		// Score not cached, calculate it
		ctx := context.Background()
		scores, err := f.Scorer.Score(ctx, []*github.Notification{n})
		if err != nil {
			// If there's an error, assume the notification doesn't match
			return false
		}
		score = scores[n.GetID()]
		f.Scores[n.GetID()] = score
	}

	// Check if the score is within the range
	if f.MinScore > 0 && score.Total < f.MinScore {
		return false
	}

	if f.MaxScore > 0 && score.Total > f.MaxScore {
		return false
	}

	return true
}

// Description returns a human-readable description of the filter
func (f *ScoreFilter) Description() string {
	if f.MinScore > 0 && f.MaxScore > 0 {
		return fmt.Sprintf("score between %d and %d", f.MinScore, f.MaxScore)
	} else if f.MinScore > 0 {
		return fmt.Sprintf("score >= %d", f.MinScore)
	} else if f.MaxScore > 0 {
		return fmt.Sprintf("score <= %d", f.MaxScore)
	}
	return "any score"
}

// SetScores sets the scores for notifications
func (f *ScoreFilter) SetScores(scores map[string]*scoring.NotificationScore) {
	f.Scores = scores
}

// GetScores gets the scores for notifications
func (f *ScoreFilter) GetScores() map[string]*scoring.NotificationScore {
	return f.Scores
}

// ensure ScoreFilter implements filter.Filter
var _ filter.Filter = (*ScoreFilter)(nil)
