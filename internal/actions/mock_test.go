package actions

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/SharanRP/gh-notif/internal/github"
)

// mockClient is a mock implementation of the GitHub client
type mockClient struct {
	// MarkThreadReadFunc is a function that mocks the MarkThreadRead method
	MarkThreadReadFunc func(threadID string) (*github.Response, error)
	// MarkAllNotificationsReadFunc is a function that mocks the MarkAllNotificationsRead method
	MarkAllNotificationsReadFunc func() error
	// MarkRepositoryNotificationsReadFunc is a function that mocks the MarkRepositoryNotificationsRead method
	MarkRepositoryNotificationsReadFunc func(owner, repo string) error
	// GetRawClientFunc is a function that mocks the GetRawClient method
	GetRawClientFunc func() *github.Client
}

// MarkThreadRead mocks the MarkThreadRead method
func (m *mockClient) MarkThreadRead(threadID string) (*github.Response, error) {
	if m.MarkThreadReadFunc != nil {
		return m.MarkThreadReadFunc(threadID)
	}
	return &github.Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
	}, nil
}

// MarkAllNotificationsRead mocks the MarkAllNotificationsRead method
func (m *mockClient) MarkAllNotificationsRead() error {
	if m.MarkAllNotificationsReadFunc != nil {
		return m.MarkAllNotificationsReadFunc()
	}
	return nil
}

// MarkRepositoryNotificationsRead mocks the MarkRepositoryNotificationsRead method
func (m *mockClient) MarkRepositoryNotificationsRead(owner, repo string) error {
	if m.MarkRepositoryNotificationsReadFunc != nil {
		return m.MarkRepositoryNotificationsReadFunc(owner, repo)
	}
	return nil
}

// GetRawClient mocks the GetRawClient method
func (m *mockClient) GetRawClient() *github.Client {
	if m.GetRawClientFunc != nil {
		return m.GetRawClientFunc()
	}
	return github.NewClient(nil)
}

// setupMockClient sets up a mock client for testing
func setupMockClient(t *testing.T) (*mockClient, func()) {
	// Create a mock client
	client := &mockClient{}

	// Save the original GetClient function
	originalGetClient := GetClient

	// Override the GetClient function
	GetClient = func(ctx context.Context) (*githubclient.Client, error) {
		return &githubclient.Client{}, nil
	}

	// Return the mock client and a cleanup function
	return client, func() {
		GetClient = originalGetClient
	}
}
