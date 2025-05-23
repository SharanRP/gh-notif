package discussions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SharanRP/gh-notif/internal/auth"
	"github.com/SharanRP/gh-notif/internal/config"
)

// GraphQLClient handles GitHub GraphQL API requests for discussions
type GraphQLClient struct {
	httpClient    *http.Client
	baseURL       string
	configManager *config.ConfigManager
	debug         bool
}

// NewGraphQLClient creates a new GraphQL client for discussions
func NewGraphQLClient(ctx context.Context) (*GraphQLClient, error) {
	// Get authenticated HTTP client
	httpClient, err := auth.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated client: %w", err)
	}

	// Get config manager
	configManager := config.NewConfigManager()
	if err := configManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	config := configManager.GetConfig()

	return &GraphQLClient{
		httpClient:    httpClient,
		baseURL:       config.API.BaseURL + "/graphql",
		configManager: configManager,
		debug:         config.Advanced.Debug,
	}, nil
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of a GraphQL error
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Execute executes a GraphQL query
func (c *GraphQLClient) Execute(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	// Create the request
	req := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	// Marshal the request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Debug logging
	if c.debug {
		fmt.Printf("GraphQL Request: %s\n", string(reqBody))
	}

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var graphqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphqlResp.Errors) > 0 {
		return fmt.Errorf("GraphQL errors: %v", graphqlResp.Errors)
	}

	// Debug logging
	if c.debug {
		fmt.Printf("GraphQL Response: %s\n", string(graphqlResp.Data))
	}

	// Unmarshal the data
	if err := json.Unmarshal(graphqlResp.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// GetDiscussions fetches discussions from a repository
func (c *GraphQLClient) GetDiscussions(ctx context.Context, owner, repo string, filter DiscussionFilter, options DiscussionOptions) ([]Discussion, error) {
	query := `
		query GetDiscussions($owner: String!, $name: String!, $first: Int!, $after: String, $orderBy: DiscussionOrder, $categoryId: ID, $answered: Boolean) {
			repository(owner: $owner, name: $name) {
				discussions(first: $first, after: $after, orderBy: $orderBy, categoryId: $categoryId, answered: $answered) {
					pageInfo {
						hasNextPage
						endCursor
					}
					nodes {
						id
						number
						title
						body
						bodyHTML
						bodyText
						url
						locked
						createdAt
						updatedAt
						isAnswered
						repository {
							id
							name
							nameWithOwner
							url
							isPrivate
							owner {
								login
								avatarUrl
								url
							}
						}
						category {
							id
							name
							description
							emoji
							isAnswerable
							createdAt
							updatedAt
						}
						author {
							login
							avatarUrl
							url
						}
						answer {
							id
							body
							bodyHTML
							bodyText
							url
							createdAt
							updatedAt
							author {
								login
								avatarUrl
								url
							}
							isAnswer
						}
						comments(first: 1) {
							totalCount
						}
						reactionGroups {
							content
							users {
								totalCount
							}
						}
						viewerDidAuthor
						viewerSubscription
						viewerCanReact
						viewerCanUpdate
						viewerCanDelete
					}
				}
			}
		}
	`

	// Build variables
	variables := map[string]interface{}{
		"owner": owner,
		"name":  repo,
		"first": 50, // Default limit
	}

	// Add filter parameters
	if filter.Category != "" {
		variables["categoryId"] = filter.Category
	}

	// Handle answered filter
	if filter.Answered != nil {
		variables["answered"] = *filter.Answered
	}

	// Add sorting
	orderBy := map[string]interface{}{
		"field":     "CREATED_AT",
		"direction": "DESC",
	}
	if filter.Sort != "" {
		switch filter.Sort {
		case "created":
			orderBy["field"] = "CREATED_AT"
		case "updated":
			orderBy["field"] = "UPDATED_AT"
		}
	}
	if filter.Direction != "" {
		switch strings.ToUpper(filter.Direction) {
		case "ASC":
			orderBy["direction"] = "ASC"
		case "DESC":
			orderBy["direction"] = "DESC"
		default:
			orderBy["direction"] = "DESC"
		}
	}
	variables["orderBy"] = orderBy

	// Execute the query
	var result struct {
		Repository struct {
			Discussions struct {
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []struct {
					ID         string    `json:"id"`
					Number     int       `json:"number"`
					Title      string    `json:"title"`
					Body       string    `json:"body"`
					BodyHTML   string    `json:"bodyHTML"`
					BodyText   string    `json:"bodyText"`
					URL        string    `json:"url"`
					Locked     bool      `json:"locked"`
					CreatedAt  time.Time `json:"createdAt"`
					UpdatedAt  time.Time `json:"updatedAt"`
					IsAnswered bool      `json:"isAnswered"`
					Repository struct {
						ID            string `json:"id"`
						Name          string `json:"name"`
						NameWithOwner string `json:"nameWithOwner"`
						URL           string `json:"url"`
						IsPrivate     bool   `json:"isPrivate"`
						Owner         struct {
							Login     string `json:"login"`
							AvatarURL string `json:"avatarUrl"`
							URL       string `json:"url"`
						} `json:"owner"`
					} `json:"repository"`
					Category struct {
						ID           string    `json:"id"`
						Name         string    `json:"name"`
						Description  string    `json:"description"`
						Emoji        string    `json:"emoji"`
						Slug         string    `json:"slug"`
						IsAnswerable bool      `json:"isAnswerable"`
						CreatedAt    time.Time `json:"createdAt"`
						UpdatedAt    time.Time `json:"updatedAt"`
					} `json:"category"`
					Author struct {
						Login     string `json:"login"`
						AvatarURL string `json:"avatarUrl"`
						URL       string `json:"url"`
					} `json:"author"`
					Answer *struct {
						ID        string    `json:"id"`
						Body      string    `json:"body"`
						BodyHTML  string    `json:"bodyHTML"`
						BodyText  string    `json:"bodyText"`
						URL       string    `json:"url"`
						CreatedAt time.Time `json:"createdAt"`
						UpdatedAt time.Time `json:"updatedAt"`
						Author    struct {
							Login     string `json:"login"`
							AvatarURL string `json:"avatarUrl"`
							URL       string `json:"url"`
						} `json:"author"`
						IsAnswer bool `json:"isAnswer"`
					} `json:"answer"`
					Comments struct {
						TotalCount int `json:"totalCount"`
					} `json:"comments"`
					ReactionGroups []struct {
						Content string `json:"content"`
						Users   struct {
							TotalCount int `json:"totalCount"`
						} `json:"users"`
					} `json:"reactionGroups"`
					ViewerDidAuthor    bool   `json:"viewerDidAuthor"`
					ViewerSubscription string `json:"viewerSubscription"`
					ViewerCanReact     bool   `json:"viewerCanReact"`
					ViewerCanUpdate    bool   `json:"viewerCanUpdate"`
					ViewerCanDelete    bool   `json:"viewerCanDelete"`
				} `json:"nodes"`
			} `json:"discussions"`
		} `json:"repository"`
	}

	if err := c.Execute(ctx, query, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	// Convert to our Discussion type
	discussions := make([]Discussion, len(result.Repository.Discussions.Nodes))
	for i, node := range result.Repository.Discussions.Nodes {
		// Calculate upvote count from reaction groups
		upvoteCount := 0
		reactionCount := 0
		for _, group := range node.ReactionGroups {
			if group.Content == "THUMBS_UP" {
				upvoteCount = group.Users.TotalCount
			}
			reactionCount += group.Users.TotalCount
		}

		discussion := Discussion{
			ID:            node.ID,
			Number:        node.Number,
			Title:         node.Title,
			Body:          node.Body,
			BodyHTML:      node.BodyHTML,
			BodyText:      node.BodyText,
			URL:           node.URL,
			State:         "OPEN", // Default state since discussions don't have explicit states
			Locked:        node.Locked,
			CreatedAt:     node.CreatedAt,
			UpdatedAt:     node.UpdatedAt,
			UpvoteCount:   upvoteCount,
			CommentCount:  node.Comments.TotalCount,
			ReactionCount: reactionCount,
			Repository: Repository{
				ID:       node.Repository.ID,
				Name:     node.Repository.Name,
				FullName: node.Repository.NameWithOwner,
				URL:      node.Repository.URL,
				Private:  node.Repository.IsPrivate,
				Owner: User{
					Login:     node.Repository.Owner.Login,
					AvatarURL: node.Repository.Owner.AvatarURL,
					URL:       node.Repository.Owner.URL,
				},
			},
			Category: Category{
				ID:           node.Category.ID,
				Name:         node.Category.Name,
				Description:  node.Category.Description,
				Emoji:        node.Category.Emoji,
				Slug:         node.Category.Slug,
				IsAnswerable: node.Category.IsAnswerable,
				CreatedAt:    node.Category.CreatedAt,
				UpdatedAt:    node.Category.UpdatedAt,
			},
			Author: User{
				Login:     node.Author.Login,
				AvatarURL: node.Author.AvatarURL,
				URL:       node.Author.URL,
			},
			ViewerDidAuthor:    node.ViewerDidAuthor,
			ViewerSubscription: node.ViewerSubscription,
			ViewerCanReact:     node.ViewerCanReact,
			ViewerCanUpdate:    node.ViewerCanUpdate,
			ViewerCanDelete:    node.ViewerCanDelete,
		}

		// Convert answer if present
		if node.Answer != nil {
			discussion.Answer = &Comment{
				ID:        node.Answer.ID,
				Body:      node.Answer.Body,
				BodyHTML:  node.Answer.BodyHTML,
				BodyText:  node.Answer.BodyText,
				URL:       node.Answer.URL,
				CreatedAt: node.Answer.CreatedAt,
				UpdatedAt: node.Answer.UpdatedAt,
				Author: User{
					Login:     node.Answer.Author.Login,
					AvatarURL: node.Answer.Author.AvatarURL,
					URL:       node.Answer.Author.URL,
				},
				IsAnswer: node.Answer.IsAnswer,
			}
		}

		// Initialize empty slices for labels and assignees (discussions don't have these)
		discussion.Labels = []Label{}
		discussion.Assignees = []User{}

		discussions[i] = discussion
	}

	return discussions, nil
}
