package github

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
)

func TestClientCreation(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a mock client
	client, server, err := NewMockClient(ctx)
	defer server.Close()

	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Apply options
	WithMaxConcurrent(10)(client)
	WithRetryCount(3)(client)
	WithRetryDelay(1*time.Second)(client)
	WithTimeout(30*time.Second)(client)
	WithCacheTTL(5*time.Minute)(client)
	WithDebug(true)(client)

	// Check that the options were applied
	if client.maxConcurrent != 10 {
		t.Errorf("maxConcurrent = %d, want %d", client.maxConcurrent, 10)
	}
	if client.retryCount != 3 {
		t.Errorf("retryCount = %d, want %d", client.retryCount, 3)
	}
	if client.retryDelay != 1*time.Second {
		t.Errorf("retryDelay = %v, want %v", client.retryDelay, 1*time.Second)
	}
	if client.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", client.timeout, 30*time.Second)
	}
	if client.cacheTTL != 5*time.Minute {
		t.Errorf("cacheTTL = %v, want %v", client.cacheTTL, 5*time.Minute)
	}
	if client.debug != true {
		t.Errorf("debug = %v, want %v", client.debug, true)
	}
}

func TestCache(t *testing.T) {
	// Create a cache
	cache := NewCache()

	// Set a value
	cache.Set("key", "value", 1*time.Minute)

	// Get the value
	value, found := cache.Get("key")
	if !found {
		t.Errorf("Get() found = %v, want %v", found, true)
	}
	if value != "value" {
		t.Errorf("Get() value = %v, want %v", value, "value")
	}

	// Set a value with a short TTL
	cache.Set("expired", "value", 1*time.Nanosecond)
	time.Sleep(10 * time.Millisecond)

	// Get the expired value
	_, found = cache.Get("expired")
	if found {
		t.Errorf("Get() found = %v, want %v", found, false)
	}

	// Delete a value
	cache.Delete("key")
	_, found = cache.Get("key")
	if found {
		t.Errorf("Get() found = %v, want %v", found, false)
	}

	// Clear the cache
	cache.Set("key1", "value1", 1*time.Minute)
	cache.Set("key2", "value2", 1*time.Minute)
	cache.Clear()
	_, found = cache.Get("key1")
	if found {
		t.Errorf("Get() found = %v, want %v", found, false)
	}
	_, found = cache.Get("key2")
	if found {
		t.Errorf("Get() found = %v, want %v", found, false)
	}
}

func TestRateLimiting(t *testing.T) {
	// Create a mock server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a rate limit response
		w.Header().Set("X-RateLimit-Limit", "5000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", "1")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	})

	// Create a client with the mock server
	ctx := context.Background()
	client, server, err := NewTestClient(ctx, handler)
	defer server.Close()

	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Create a GitHub response with rate limit info
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: http.StatusForbidden,
			Status:     "403 Forbidden",
		},
		Rate: github.Rate{
			Limit:     5000,
			Remaining: 0,
			Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Second)},
		},
	}

	// Set a short delay for testing
	client.retryDelay = 10 * time.Millisecond

	// Test handleRateLimit
	start := time.Now()
	client.handleRateLimit(resp)
	elapsed := time.Since(start)

	// Should wait at least a little bit
	if elapsed < 5*time.Millisecond {
		t.Errorf("handleRateLimit() waited %v, want at least %v", elapsed, 5*time.Millisecond)
	}
}

func TestFilterNotifications(t *testing.T) {
	// Create a mock client
	ctx := context.Background()
	client, server, err := NewMockClient(ctx)
	defer server.Close()

	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	// Create test notifications
	notifications := []*github.Notification{
		{
			Repository: &github.Repository{
				FullName: github.String("org1/repo1"),
			},
		},
		{
			Repository: &github.Repository{
				FullName: github.String("org1/repo2"),
			},
		},
		{
			Repository: &github.Repository{
				FullName: github.String("org2/repo3"),
			},
		},
	}

	// Test filtering by repository
	filtered := client.filterNotifications(notifications, "org1/repo1", "")
	if len(filtered) != 1 {
		t.Errorf("filterNotifications() len = %d, want %d", len(filtered), 1)
	}
	if filtered[0].GetRepository().GetFullName() != "org1/repo1" {
		t.Errorf("filterNotifications() repo = %s, want %s", filtered[0].GetRepository().GetFullName(), "org1/repo1")
	}

	// Test filtering by organization
	filtered = client.filterNotifications(notifications, "", "org1")
	if len(filtered) != 2 {
		t.Errorf("filterNotifications() len = %d, want %d", len(filtered), 2)
	}
	for _, n := range filtered {
		if n.GetRepository().GetFullName() != "org1/repo1" && n.GetRepository().GetFullName() != "org1/repo2" {
			t.Errorf("filterNotifications() repo = %s, want org1/*", n.GetRepository().GetFullName())
		}
	}

	// Test no filtering
	filtered = client.filterNotifications(notifications, "", "")
	if len(filtered) != 3 {
		t.Errorf("filterNotifications() len = %d, want %d", len(filtered), 3)
	}
}
