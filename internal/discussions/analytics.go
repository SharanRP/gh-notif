package discussions

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// AnalyticsEngine provides discussion analytics and insights
type AnalyticsEngine struct {
	client *Client
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(client *Client) *AnalyticsEngine {
	return &AnalyticsEngine{
		client: client,
	}
}

// GenerateAnalytics generates comprehensive analytics for discussions
func (ae *AnalyticsEngine) GenerateAnalytics(ctx context.Context, repositories []string, timeRange TimeRange) (*DiscussionAnalytics, error) {
	// Fetch all discussions in the time range
	filter := DiscussionFilter{
		CreatedAfter: &timeRange.Start,
		CreatedBefore: &timeRange.End,
		State: "all",
	}

	options := DiscussionOptions{
		IncludeComments: true,
		IncludeReactions: true,
		UseCache: true,
	}

	discussions, err := ae.client.GetDiscussions(ctx, repositories, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discussions for analytics: %w", err)
	}

	// Generate analytics
	analytics := &DiscussionAnalytics{
		TimeRange: timeRange,
		CategoryStats: make(map[string]CategoryStats),
	}

	// Process each repository
	repoStats := make(map[string]*DiscussionAnalytics)
	for _, discussion := range discussions {
		repoKey := discussion.Repository.FullName
		if _, exists := repoStats[repoKey]; !exists {
			repoStats[repoKey] = &DiscussionAnalytics{
				Repository: discussion.Repository,
				TimeRange: timeRange,
				CategoryStats: make(map[string]CategoryStats),
			}
		}
		ae.processDiscussionForAnalytics(repoStats[repoKey], discussion)
	}

	// Aggregate repository stats
	for _, repoAnalytics := range repoStats {
		ae.aggregateAnalytics(analytics, repoAnalytics)
	}

	// Calculate derived metrics
	ae.calculateDerivedMetrics(analytics)

	// Generate trending topics
	analytics.TrendingTopics = ae.extractTrendingTopics(discussions)

	// Generate top contributors
	analytics.TopAuthors = ae.getTopAuthors(discussions)
	analytics.TopCommenters = ae.getTopCommenters(discussions)

	return analytics, nil
}

// GetRepositoryAnalytics generates analytics for a specific repository
func (ae *AnalyticsEngine) GetRepositoryAnalytics(ctx context.Context, repository string, timeRange TimeRange) (*DiscussionAnalytics, error) {
	return ae.GenerateAnalytics(ctx, []string{repository}, timeRange)
}

// GetTrendingDiscussions identifies trending discussions based on engagement
func (ae *AnalyticsEngine) GetTrendingDiscussions(ctx context.Context, repositories []string, timeRange TimeRange, limit int) ([]Discussion, error) {
	// Fetch recent discussions
	filter := DiscussionFilter{
		CreatedAfter: &timeRange.Start,
		State: "open",
		Sort: "updated",
		Direction: "desc",
	}

	options := DiscussionOptions{
		IncludeComments: true,
		IncludeReactions: true,
		UseCache: true,
	}

	discussions, err := ae.client.GetDiscussions(ctx, repositories, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discussions for trending analysis: %w", err)
	}

	// Calculate trend scores
	type trendingDiscussion struct {
		Discussion Discussion
		TrendScore float64
	}

	var trending []trendingDiscussion
	for _, discussion := range discussions {
		score := ae.calculateTrendScore(discussion, timeRange)
		trending = append(trending, trendingDiscussion{
			Discussion: discussion,
			TrendScore: score,
		})
	}

	// Sort by trend score
	sort.Slice(trending, func(i, j int) bool {
		return trending[i].TrendScore > trending[j].TrendScore
	})

	// Extract top discussions
	if limit > 0 && len(trending) > limit {
		trending = trending[:limit]
	}

	result := make([]Discussion, len(trending))
	for i, td := range trending {
		result[i] = td.Discussion
	}

	return result, nil
}

// GetUnansweredQuestions finds unanswered questions that need attention
func (ae *AnalyticsEngine) GetUnansweredQuestions(ctx context.Context, repositories []string, maxAge time.Duration) ([]Discussion, error) {
	// Calculate time threshold
	threshold := time.Now().Add(-maxAge)

	filter := DiscussionFilter{
		CreatedBefore: &threshold,
		State: "open",
		Answered: boolPtr(false),
	}

	options := DiscussionOptions{
		UseCache: true,
	}

	discussions, err := ae.client.GetDiscussions(ctx, repositories, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unanswered questions: %w", err)
	}

	// Filter for answerable categories only
	var questions []Discussion
	for _, discussion := range discussions {
		if discussion.Category.IsAnswerable && discussion.Answer == nil {
			questions = append(questions, discussion)
		}
	}

	// Sort by age (oldest first)
	sort.Slice(questions, func(i, j int) bool {
		return questions[i].CreatedAt.Before(questions[j].CreatedAt)
	})

	return questions, nil
}

// GetEngagementMetrics calculates engagement metrics for discussions
func (ae *AnalyticsEngine) GetEngagementMetrics(ctx context.Context, repositories []string, timeRange TimeRange) (map[string]float64, error) {
	analytics, err := ae.GenerateAnalytics(ctx, repositories, timeRange)
	if err != nil {
		return nil, err
	}

	metrics := map[string]float64{
		"total_discussions": float64(analytics.TotalDiscussions),
		"open_discussions": float64(analytics.OpenDiscussions),
		"closed_discussions": float64(analytics.ClosedDiscussions),
		"answered_discussions": float64(analytics.AnsweredDiscussions),
		"answer_rate": 0,
		"average_comments": analytics.AverageComments,
		"average_reactions": analytics.AverageReactions,
		"average_upvotes": analytics.AverageUpvotes,
		"engagement_score": 0,
	}

	// Calculate answer rate
	if analytics.TotalDiscussions > 0 {
		metrics["answer_rate"] = float64(analytics.AnsweredDiscussions) / float64(analytics.TotalDiscussions) * 100
	}

	// Calculate overall engagement score
	metrics["engagement_score"] = ae.calculateEngagementScore(analytics)

	return metrics, nil
}

// Helper methods

func (ae *AnalyticsEngine) processDiscussionForAnalytics(analytics *DiscussionAnalytics, discussion Discussion) {
	// Update basic counts
	analytics.TotalDiscussions++
	
	switch discussion.State {
	case "OPEN":
		analytics.OpenDiscussions++
	case "CLOSED":
		analytics.ClosedDiscussions++
	}

	if discussion.Answer != nil {
		analytics.AnsweredDiscussions++
	}

	// Update engagement metrics
	analytics.TotalComments += discussion.CommentCount
	analytics.TotalReactions += discussion.ReactionCount
	analytics.TotalUpvotes += discussion.UpvoteCount

	// Update category stats
	categoryKey := discussion.Category.Slug
	if stats, exists := analytics.CategoryStats[categoryKey]; exists {
		stats.DiscussionCount++
		stats.CommentCount += discussion.CommentCount
		stats.ReactionCount += discussion.ReactionCount
		if discussion.Answer != nil {
			// Update answer rate calculation
		}
		analytics.CategoryStats[categoryKey] = stats
	} else {
		analytics.CategoryStats[categoryKey] = CategoryStats{
			Category: discussion.Category,
			DiscussionCount: 1,
			CommentCount: discussion.CommentCount,
			ReactionCount: discussion.ReactionCount,
		}
	}
}

func (ae *AnalyticsEngine) aggregateAnalytics(target, source *DiscussionAnalytics) {
	target.TotalDiscussions += source.TotalDiscussions
	target.OpenDiscussions += source.OpenDiscussions
	target.ClosedDiscussions += source.ClosedDiscussions
	target.AnsweredDiscussions += source.AnsweredDiscussions
	target.TotalComments += source.TotalComments
	target.TotalReactions += source.TotalReactions
	target.TotalUpvotes += source.TotalUpvotes

	// Merge category stats
	for key, stats := range source.CategoryStats {
		if existing, exists := target.CategoryStats[key]; exists {
			existing.DiscussionCount += stats.DiscussionCount
			existing.CommentCount += stats.CommentCount
			existing.ReactionCount += stats.ReactionCount
			target.CategoryStats[key] = existing
		} else {
			target.CategoryStats[key] = stats
		}
	}
}

func (ae *AnalyticsEngine) calculateDerivedMetrics(analytics *DiscussionAnalytics) {
	if analytics.TotalDiscussions > 0 {
		analytics.AverageComments = float64(analytics.TotalComments) / float64(analytics.TotalDiscussions)
		analytics.AverageReactions = float64(analytics.TotalReactions) / float64(analytics.TotalDiscussions)
		analytics.AverageUpvotes = float64(analytics.TotalUpvotes) / float64(analytics.TotalDiscussions)
	}

	// Calculate answer rates for categories
	for key, stats := range analytics.CategoryStats {
		if stats.DiscussionCount > 0 && stats.Category.IsAnswerable {
			// This would need to be calculated during processing
			// stats.AnswerRate = float64(answeredCount) / float64(stats.DiscussionCount) * 100
		}
		analytics.CategoryStats[key] = stats
	}
}

func (ae *AnalyticsEngine) calculateTrendScore(discussion Discussion, timeRange TimeRange) float64 {
	// Calculate trend score based on various factors
	score := 0.0

	// Age factor (newer discussions get higher scores)
	age := time.Since(discussion.CreatedAt)
	maxAge := timeRange.End.Sub(timeRange.Start)
	ageFactor := 1.0 - (age.Seconds() / maxAge.Seconds())
	score += ageFactor * 0.3

	// Engagement factor
	engagementScore := float64(discussion.UpvoteCount)*0.5 + 
					  float64(discussion.CommentCount)*0.3 + 
					  float64(discussion.ReactionCount)*0.2
	score += engagementScore * 0.4

	// Activity factor (recent updates)
	timeSinceUpdate := time.Since(discussion.UpdatedAt)
	activityFactor := 1.0 / (1.0 + timeSinceUpdate.Hours()/24.0) // Decay over days
	score += activityFactor * 0.3

	return score
}

func (ae *AnalyticsEngine) calculateEngagementScore(analytics *DiscussionAnalytics) float64 {
	if analytics.TotalDiscussions == 0 {
		return 0
	}

	// Weighted engagement score
	score := analytics.AverageComments*0.4 + 
			 analytics.AverageReactions*0.3 + 
			 analytics.AverageUpvotes*0.3

	return score
}

func (ae *AnalyticsEngine) extractTrendingTopics(discussions []Discussion) []TopicStats {
	// Extract keywords from titles and bodies
	topicCounts := make(map[string]int)
	
	for _, discussion := range discussions {
		// Simple keyword extraction (could be enhanced with NLP)
		words := strings.Fields(strings.ToLower(discussion.Title))
		for _, word := range words {
			if len(word) > 3 { // Filter short words
				topicCounts[word]++
			}
		}
	}

	// Convert to TopicStats and sort
	var topics []TopicStats
	for topic, count := range topicCounts {
		topics = append(topics, TopicStats{
			Topic: topic,
			DiscussionCount: count,
			TrendScore: float64(count), // Simple trend score
		})
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].TrendScore > topics[j].TrendScore
	})

	// Return top 10
	if len(topics) > 10 {
		topics = topics[:10]
	}

	return topics
}

func (ae *AnalyticsEngine) getTopAuthors(discussions []Discussion) []UserStats {
	userStats := make(map[string]*UserStats)

	for _, discussion := range discussions {
		key := discussion.Author.Login
		if stats, exists := userStats[key]; exists {
			stats.DiscussionCount++
			stats.UpvoteCount += discussion.UpvoteCount
		} else {
			userStats[key] = &UserStats{
				User: discussion.Author,
				DiscussionCount: 1,
				UpvoteCount: discussion.UpvoteCount,
			}
		}
	}

	// Convert to slice and sort
	var authors []UserStats
	for _, stats := range userStats {
		authors = append(authors, *stats)
	}

	sort.Slice(authors, func(i, j int) bool {
		return authors[i].DiscussionCount > authors[j].DiscussionCount
	})

	// Return top 10
	if len(authors) > 10 {
		authors = authors[:10]
	}

	return authors
}

func (ae *AnalyticsEngine) getTopCommenters(discussions []Discussion) []UserStats {
	// This would require fetching comment data
	// For now, return empty slice
	return []UserStats{}
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
