package actions

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v60/github"
	githubclient "github.com/user/gh-notif/internal/github"
)

func TestMuteRepository(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up the mock functions
	mockClient.MarkRepositoryNotificationsReadFunc = func(owner, repo string) error {
		if owner != "testowner" || repo != "testrepo" {
			t.Errorf("Expected owner/repo to be testowner/testrepo, got %s/%s", owner, repo)
		}
		return nil
	}

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call MuteRepository
	result, err := MuteRepository(ctx, "testowner/testrepo")
	if err != nil {
		t.Fatalf("MuteRepository failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionMute {
		t.Errorf("Expected action type %s, got %s", ActionMute, result.Action.Type)
	}
	if result.Action.RepositoryName != "testowner/testrepo" {
		t.Errorf("Expected repository name testowner/testrepo, got %s", result.Action.RepositoryName)
	}
}

func TestUnmuteRepository(t *testing.T) {
	// Set up a mock client
	mockClient, cleanup := setupMockClient(t)
	defer cleanup()

	// Set up a mock raw client that returns success for any request
	mockClient.GetRawClientFunc = func() *github.Client {
		return github.NewClient(nil)
	}

	// Create a context
	ctx := context.Background()

	// Call UnmuteRepository
	result, err := UnmuteRepository(ctx, "testowner/testrepo")
	if err != nil {
		t.Fatalf("UnmuteRepository failed: %v", err)
	}

	// Check the result
	if !result.Success {
		t.Errorf("Expected success, got failure")
	}
	if result.Action.Type != ActionMute {
		t.Errorf("Expected action type %s, got %s", ActionMute, result.Action.Type)
	}
	if result.Action.RepositoryName != "testowner/testrepo" {
		t.Errorf("Expected repository name testowner/testrepo, got %s", result.Action.RepositoryName)
	}
	if result.Action.Metadata == nil || result.Action.Metadata["unmute"] != true {
		t.Errorf("Expected metadata to have unmute=true")
	}
}


